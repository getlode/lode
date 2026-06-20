#!/usr/bin/env bash
# Rigorous benchmark: lode vs DVC on the local hot path (add / status).
#
# Methodology (addresses the usual "single cold run" criticism):
#   - N runs per cell (RUNS, default 5); reports MEDIAN ± stddev and the speedup
#     of the medians. Cells whose median is in the timing noise floor are flagged.
#   - Execution ORDER is alternated each run (dvc-first / lode-first) so neither
#     tool benefits systematically from the page cache the other warmed. True
#     cold (drop_caches) needs root and is not assumed; results are WARM and
#     labeled as such — the alternation removes the order bias.
#   - Two file-size regimes: many-small and few-large (the CPU-bound hashing
#     case), plus peak RSS per operation (GNU /usr/bin/time).
#   - Encodes the interop claim: after `lode add`, asserts `dvc status` says the
#     repo is up to date (DVC reads what lode produced).
#
# Usage: LODE=./lode DVC_BIN=$(which dvc) RUNS=5 scripts/benchmark.sh
set -uo pipefail
export LC_ALL=C DVC_NO_ANALYTICS=1

LODE="${LODE:-$(pwd)/lode}"
DVC="${DVC_BIN:-dvc}"
RUNS="${RUNS:-6}"
GEN="$(cd "$(dirname "$0")" && pwd)/gen-dataset.sh"
TIME=/usr/bin/time
NOISE=0.05 # seconds; medians at/below this are flagged as noise-floor

have_dvc=1; command -v "$DVC" >/dev/null 2>&1 || have_dvc=0

# "count:bytes" regimes. Override with REGIMES="...".
REGIMES="${REGIMES:-20000:1024 8:67108864}"

tmproot=$(mktemp -d); trap 'rm -rf "$tmproot"' EXIT

# measure CMD... -> "elapsed_seconds max_rss_kb"
measure() {
  local tf="$tmproot/t"
  "$TIME" -f '%e %M' -o "$tf" "$@" >/dev/null 2>&1 || true
  cat "$tf"
}

# stats <file-of-numbers> -> "median stddev min"
stats() {
  sort -n "$1" | awk '
    {a[NR]=$1; s+=$1; ss+=$1*$1}
    END{
      n=NR; if(n==0){print "0 0 0"; exit}
      m=(n%2)?a[(n+1)/2]:(a[n/2]+a[n/2+1])/2
      mean=s/n; sd=(n>1)?sqrt(ss/n-mean*mean):0
      printf "%.3f %.3f %.3f", m, sd, a[1]
    }'
}

reset_repo() { rm -rf .dvc "$1.dvc" .dvcignore .gitignore; }

# one_pass <tool-bin> <init-args...> writes add/status/incr samples for $sub.
# Globals: sub, and append files $samp_<tool>_<op>_{t,r}
record() { # tool op elapsed rss
  echo "$3" >> "$tmproot/$1.$2.t"; echo "$4" >> "$tmproot/$1.$2.r"
}

pass() { # tool bin parent sub
  local tool="$1" bin="$2" sub="$4"
  reset_repo "$sub"
  "$bin" init --no-scm -q >/dev/null 2>&1 || "$bin" init --no-scm >/dev/null 2>&1
  read e r < <(measure "$bin" add "$sub");      record "$tool" add "$e" "$r"
  read e r < <(measure "$bin" status);          record "$tool" status "$e" "$r"
  printf 'x' >> "$(find "$sub" -type f | head -1)"
  read e r < <(measure "$bin" add "$sub");      record "$tool" incr "$e" "$r"
  reset_repo "$sub"
}

interop_check() { # parent sub -> echoes PASS/FAIL/SKIP
  [ "$have_dvc" = 1 ] || { echo SKIP; return; }
  reset_repo "$2"
  "$LODE" init --no-scm >/dev/null 2>&1
  "$LODE" add "$2" >/dev/null 2>&1
  if "$DVC" status 2>&1 | grep -qi "up to date"; then echo PASS; else echo FAIL; fi
  reset_repo "$2"
}

echo "Environment: $(nproc) cores | $(uname -srm) | RUNS=$RUNS | cache=WARM (order alternated)"
[ "$have_dvc" = 1 ] && echo "DVC: $($DVC --version 2>/dev/null|head -1) | lode: $($LODE --version 2>/dev/null)"
echo

bench_dataset() { # work sub header
  local work="$1" sub="$2" header="$3" op lm lmed lrss dm dmed sp flag i iop
  ( cd "$work"
    rm -f "$tmproot"/lode.* "$tmproot"/dvc.* 2>/dev/null
    for i in $(seq 1 "$RUNS"); do
      if [ "$have_dvc" = 1 ] && [ $((i % 2)) -eq 1 ]; then
        pass dvc "$DVC" "$work" "$sub"; pass lode "$LODE" "$work" "$sub"
      else
        pass lode "$LODE" "$work" "$sub"
        [ "$have_dvc" = 1 ] && pass dvc "$DVC" "$work" "$sub"
      fi
    done
    iop=$(interop_check "$work" "$sub")

    echo "### $header | interop(lode->dvc): $iop"
    echo "| operation | DVC median±sd | lode median±sd | speedup | DVC RSS | lode RSS |"
    echo "|-----------|--------------:|---------------:|--------:|--------:|---------:|"
    for op in add status incr; do
      lm=$(stats "$tmproot/lode.$op.t"); lmed=${lm%% *}
      lrss=$(sort -n "$tmproot/lode.$op.r" | tail -1)
      if [ "$have_dvc" = 1 ]; then
        dm=$(stats "$tmproot/dvc.$op.t"); dmed=${dm%% *}
        sp=$(awk -v d="$dmed" -v l="$lmed" 'BEGIN{ if(l>0) printf "%.1fx", d/l; else printf "-" }')
        flag=$(awk -v d="$dmed" -v l="$lmed" -v n="$NOISE" 'BEGIN{ if(d<=n||l<=n) printf " (noise-floor)"; }')
        drss=$(sort -n "$tmproot/dvc.$op.r" | tail -1)
        printf '| %s | %ss ±%s | %ss ±%s | **%s**%s | %sMB | %sMB |\n' \
          "$op" "$dmed" "$(echo $dm|cut -d' ' -f2)" "$lmed" "$(echo $lm|cut -d' ' -f2)" "$sp" "$flag" \
          "$(awk -v k="$drss" 'BEGIN{printf "%.0f", k/1024}')" \
          "$(awk -v k="$lrss" 'BEGIN{printf "%.0f", k/1024}')"
      else
        printf '| %s | (no dvc) | %ss ±%s | - | %sMB |\n' "$op" "$lmed" "$(echo $lm|cut -d' ' -f2)" \
          "$(awk -v k="$lrss" 'BEGIN{printf "%.0f", k/1024}')"
      fi
    done
    echo
  )
}

if [ $# -ge 1 ] && [ -d "$1" ]; then
  # Real dataset directory.
  parent="$(cd "$(dirname "$1")" && pwd)"; sub="$(basename "$1")"
  nf=$(find "$parent/$sub" -type f | wc -l)
  bench_dataset "$parent" "$sub" "Real dataset: $sub ($nf files)"
else
  for regime in $REGIMES; do
    count="${regime%%:*}"; bytes="${regime##*:}"
    work="$tmproot/w"; rm -rf "$work"; mkdir -p "$work"
    bash "$GEN" "$work/data" "$count" "$bytes" >/dev/null
    totMB=$(( count * bytes / 1024 / 1024 ))
    bench_dataset "$work" "data" "Regime: $count files x ${bytes}B (~${totMB} MB total)"
  done
fi
