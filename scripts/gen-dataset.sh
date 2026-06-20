#!/usr/bin/env bash
# Genera un dataset determinista de N archivos para benchmarks y el oráculo.
# Uso: gen-dataset.sh <dir> <n> [bytes-por-archivo]
# No pipefail: `yes | head` legitimately makes `yes` exit with SIGPIPE.
set -eu

dir="${1:?dir requerido}"
n="${2:?n requerido}"
bytes="${3:-256}"

mkdir -p "$dir"
for i in $(seq 1 "$n"); do
  # Deterministic content derived from the index (same i -> same bytes), via a
  # repeating line so it scales to large file sizes (printf alone is too short).
  yes "dvcgo-deterministic-$(printf '%010d' "$i")-padding-0123456789abcdef" \
    | head -c "$bytes" > "$dir/file_$(printf '%06d' "$i").bin"
done
echo "Generated $n files in $dir"
