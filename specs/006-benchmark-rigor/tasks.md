---
description: "Task list for 006-benchmark-rigor"
---

# Tasks: Benchmark riguroso y sin sesgo

**Tests**: el harness se autovalida (assert de interop). Feature de tooling/docs.

## Phase 1: Harness

- [X] T001 Fix `scripts/gen-dataset.sh`: generar archivos grandes reales (`yes | head -c`, sin `pipefail`) per FR-005
- [X] T002 Reescribir `scripts/benchmark.sh`: N corridas con mediana ± desvío, **alternancia de orden** dvc/lode para quitar el sesgo de page-cache, etiqueta de régimen (warm), flag de noise-floor per FR-001/002/003/006
- [X] T003 Agregar al harness: regímenes de tamaño de archivo (muchos-chicos y pocos-grandes) y medición de **RSS pico** por operación (GNU time) per FR-005/007
- [X] T004 Codificar el **assert de interoperabilidad** (tras `lode add`, `dvc status` == "up to date") en el harness per FR-008; correr sin DVC reporta solo lode per FR-009
- [X] T005 Capturar specs del entorno (cores, SO, versiones) en el encabezado del reporte per FR-004

## Phase 2: Re-medición y publicación

- [X] T006 Correr el harness (N≥5) sobre los regímenes y capturar los resultados
- [X] T007 Reescribir `BENCHMARKS.md`: metodología revisada (N, desvío, control de cache por alternancia, regímenes, RSS, interop) + cifras nuevas; reemplazar las cifras sesgadas anteriores per FR-010/011
- [X] T008 Actualizar la sección Benchmarks de `README.md` con las cifras honestas

## Phase 3: Validación

- [X] T009 Verificar SC-001 (sin sesgo de orden: invertir orden no cambia el speedup más allá del desvío) y SC-002/003/005 (mediana+desvío, regímenes+RSS, assert interop)

## Notes

- Cold real (drop_caches) no disponible sin root → warm + alternancia de orden, declarado honestamente.
- No toca el binario Go; los oráculos del 001/002 quedan intactos.
