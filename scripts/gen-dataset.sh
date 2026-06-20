#!/usr/bin/env bash
# Genera un dataset determinista de N archivos para benchmarks y el oráculo.
# Uso: gen-dataset.sh <dir> <n> [bytes-por-archivo]
set -euo pipefail

dir="${1:?dir requerido}"
n="${2:?n requerido}"
bytes="${3:-256}"

mkdir -p "$dir"
for i in $(seq 1 "$n"); do
  # Contenido determinista derivado del índice (mismo i -> mismos bytes).
  printf 'dvcgo-deterministic-%010d-' "$i" | head -c "$bytes" > "$dir/file_$(printf '%06d' "$i").bin"
done
echo "Generados $n archivos en $dir"
