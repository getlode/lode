# Implementation Plan: Benchmark riguroso y sin sesgo

**Branch**: `006-benchmark-rigor` | **Date**: 2026-06-20 | **Spec**: [spec.md](spec.md)

## Summary

Endurecer el harness de benchmark (feature 005) para eliminar el sesgo de page-cache y dar rigor estadístico, y republicar cifras honestas. Sin código de producto nuevo: cambia `scripts/benchmark.sh`, `scripts/gen-dataset.sh`, `BENCHMARKS.md` y la sección Benchmarks del `README.md`.

## Technical Context

**Language/Version**: Bash + awk (harness); no toca el binario Go.

**Primary Dependencies**: GNU `/usr/bin/time` (RSS). Sin hyperfine (mediana/desvío en awk → cero deps extra, máxima reproducibilidad).

**Testing**: el propio harness corre y autovalida (assert de interop lode→dvc).

**Constraints**: cold real (drop_caches) requiere root y NO se asume → se mide WARM con **alternancia de orden** entre corridas para quitar el sesgo, declarado. Mediana ± desvío de N corridas. Regímenes de tamaño de archivo (muchos-chicos y pocos-grandes). RSS por operación. Honestidad: marcar/excluir celdas en el ruido; mantener caveats (push/pull paridad).

**Project Type**: tooling/docs (no cambia comportamiento del binario).

## Constitution Check

Evaluado contra v1.0.0: **Principio IV (Performance Is the Product)** es el directamente involucrado — este feature hace la medición *defendible*. I/II (byte-compat/oráculo) intactos (no toca formato). III (cero cgo) intacto. V intacto. Sin violaciones.

## Project Structure

```text
scripts/benchmark.sh     # reescrito: N corridas, alternancia de orden, RSS, regímenes, assert interop
scripts/gen-dataset.sh   # fix: archivos grandes reales (yes|head, sin pipefail)
BENCHMARKS.md            # metodología revisada + cifras nuevas
README.md                # sección Benchmarks actualizada con las cifras honestas
specs/006-benchmark-rigor/research.md  # decisiones (rootless cold, alternancia, awk stats)
```

**Structure Decision**: Solo tooling + docs. El harness queda reproducible por cualquiera con bash+awk+GNU time, sin dependencias.

## Complexity Tracking

> No aplica — sin violaciones.
