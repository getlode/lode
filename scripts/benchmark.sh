#!/usr/bin/env bash
# Reproducible benchmark: lode vs DVC on the local hot path (add / status).
#
# Usage:
#   LODE=./lode DVC_BIN=$(which dvc) scripts/benchmark.sh
#   SCALES="1000 10000 50000" scripts/benchmark.sh           # synthetic scales
#   scripts/benchmark.sh /path/to/real/dataset-dir            # a real dataset
#
# Methodology: each operation is a single cold run on the same data, same
# machine, DVC defaults vs lode defaults (md5, reflink/copy). Indicative, not a
# statistical study — run it yourself; numbers track the structural difference
# (parallel hashing + a state DB that skips unchanged files), not luck.
set -uo pipefail
export LC_ALL=C
export DVC_NO_ANALYTICS=1

LODE="${LODE:-$(pwd)/lode}"
DVC="${DVC_BIN:-dvc}"
SCALES="${SCALES:-1000 10000 50000}"
GEN="$(cd "$(dirname "$0")" && pwd)/gen-dataset.sh"

have_dvc=1; command -v "$DVC" >/dev/null 2>&1 || have_dvc=0

elapsed() { awk -v a="$1" -v b="$2" 'BEGIN{printf "%.2f", b-a}'; }
timeit() { local t0 t1; t0=$(date +%s.%N); "$@" >/dev/null 2>&1 || true; t1=$(date +%s.%N); elapsed "$t0" "$t1"; }
speedup() { awk -v d="$1" -v l="$2" 'BEGIN{ if (l>0) printf "%.1fx", d/l; else printf "-" }'; }

# bench_dir <data-parent> <subdir>: prints a markdown row per op.
bench_one() {
  local parent="$1" sub="$2" nfiles
  nfiles=$(find "$parent/$sub" -type f | wc -l)
  cd "$parent"

  # DVC
  local da="-" ds="-" di="-"
  if [ "$have_dvc" = 1 ]; then
    rm -rf .dvc "${sub}.dvc" .dvcignore .gitignore
    "$DVC" init --no-scm -q
    da=$(timeit "$DVC" add "$sub")
    ds=$(timeit "$DVC" status)
    printf 'x' >> "$(find "$sub" -type f | head -1)"
    di=$(timeit "$DVC" add "$sub")
    rm -rf .dvc "${sub}.dvc" .dvcignore .gitignore
  fi

  # lode
  "$LODE" init --no-scm >/dev/null 2>&1
  local la ls li
  la=$(timeit "$LODE" add "$sub")
  ls=$(timeit "$LODE" status)
  printf 'x' >> "$(find "$sub" -type f | head -1)"
  li=$(timeit "$LODE" add "$sub")
  rm -rf .dvc "${sub}.dvc" .dvcignore .gitignore

  printf '| %s | add (cold) | %ss | %ss | **%s** |\n' "$nfiles" "$da" "$la" "$(speedup "$da" "$la")"
  printf '| %s | status (no change) | %ss | %ss | **%s** |\n' "$nfiles" "$ds" "$ls" "$(speedup "$ds" "$ls")"
  printf '| %s | add (1 file changed) | %ss | %ss | **%s** |\n' "$nfiles" "$di" "$li" "$(speedup "$di" "$li")"
}

echo "Environment: $(nproc) cores | $(uname -srm)"
[ "$have_dvc" = 1 ] && echo "DVC: $("$DVC" --version 2>/dev/null | head -1)" || echo "DVC: not found (lode-only)"
echo "lode: $("$LODE" --version 2>/dev/null)"
echo
echo "| files | operation | DVC | lode | speedup |"
echo "|------:|-----------|----:|-----:|--------:|"

if [ $# -ge 1 ]; then
  # Real dataset directory passed.
  parent="$(cd "$(dirname "$1")" && pwd)"; sub="$(basename "$1")"
  ( bench_one "$parent" "$sub" )
else
  for n in $SCALES; do
    work=$(mktemp -d)
    bash "$GEN" "$work/data" "$n" 256 >/dev/null
    ( bench_one "$work" "data" )   # subshell: contain the cd
    rm -rf "$work"
  done
fi
