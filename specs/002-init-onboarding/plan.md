# Implementation Plan: Bootstrap standalone y onboarding (init + doctor)

**Branch**: `002-init-onboarding` | **Date**: 2026-06-20 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/002-init-onboarding/spec.md`

## Summary

Agregar `lode init` (byte-compatible con `dvc init` en modos scm y no-scm), `lode doctor` (diagnóstico de repo/cache/format/remotes/coexistencia) y errores que guían, para que un usuario nuevo vaya de cero a trackear+push usando solo el binario, sin Python ni DVC. Reutiliza toda la infraestructura del feature 001; no agrega dependencias. El riesgo es la byte-compat de la estructura de `init`, mitigado con un test-oráculo contra `dvc init` real (ambos modos), cuyos bytes ya fueron capturados (ver [research.md](research.md)).

## Technical Context

**Language/Version**: Go 1.23+ (`CGO_ENABLED=0`)

**Primary Dependencies**: `spf13/cobra` (comandos); reutiliza `internal/{repo,cache,remote,lock}` del feature 001; `minio-go` para reachability de remote; shell-out a `git` (solo modo scm). Sin dependencias nuevas.

**Storage**: Filesystem (estructura `.dvc/` byte-compatible con DVC); remotes S3-compatible (solo para el chequeo de alcanzabilidad de `doctor`).

**Testing**: `testing` + oráculo de bytes contra `dvc init` real (ambos modos); integración para `doctor` (siembra de problemas, MinIO apagado para "inalcanzable"); interop bidireccional repo lode↔dvc.

**Target Platform**: Linux, macOS, Windows (amd64 + arm64). Nota: el `git add` del modo scm hace shell-out a `git`; el camino standalone (`--no-scm`) no toca git.

**Project Type**: CLI + librería (single project, continúa la estructura del 001).

**Performance Goals**: N/A material (init/doctor son operaciones puntuales). `doctor` usa timeouts cortos para la reachability del remote.

**Constraints**: Byte-identidad con `dvc init` en ambos modos (invariante, Constitución I). `init` no debe dañar repos existentes. Ningún camino de onboarding (init/add/push/doctor) requiere Python/DVC (Constitución III: cero cgo se mantiene).

**Scale/Scope**: 2 comandos nuevos (`init`, `doctor`) + helper de errores guiados + ajuste de `repo.Init`. Feature chico.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Evaluado contra `.specify/memory/constitution.md` (v1.0.0):

- **I. DVC Byte-Compatibility**: ✅ La estructura de `init` es byte-idéntica a `dvc init` (ambos modos); cubierto por test-oráculo. El repo resultante es operable por DVC.
- **II. Oracle-Gated Format Changes**: ✅ `init` genera artefactos de formato (config, .gitignore, .dvcignore, btime) → su oráculo de bytes es gate antes del comando.
- **III. Zero-CGO Single Binary**: ✅ Sin dependencias nuevas con cgo; el shell-out a `git` no afecta el build y solo ocurre en modo scm (git ya presente).
- **IV. Performance Is the Product**: ✅ N/A (no es hot path); `doctor` acota la reachability con timeouts.
- **V. Coexistence Over Reinvention**: ✅ `init` replica a DVC exactamente; `doctor` chequea explícitamente la coexistencia (lock) y reporta si es seguro operar.

**Sin violaciones.** Complexity Tracking vacío.

## Project Structure

### Documentation (this feature)

```text
specs/002-init-onboarding/
├── plan.md          # Este archivo
├── research.md      # Bytes exactos de dvc init (ambos modos) + decisiones
├── data-model.md    # InitMode, InitOutcome, DoctorCheck/Report
├── quickstart.md    # Escenarios de validación (oráculo init, onboarding, doctor)
├── contracts/
│   └── cli.md       # Contrato de init y doctor
└── tasks.md         # (/speckit-tasks)
```

### Source Code (repository root) — cambios sobre la base del 001

```text
internal/
├── cli/
│   ├── init.go          # NUEVO: comando `init` (--no-scm, detección git, git add)
│   ├── doctor.go        # NUEVO: comando `doctor` (checks + reporte + exit code)
│   ├── errors.go        # NUEVO: helper de errores guiados (requireRepo, hints)
│   └── root.go          # +registrar init, doctor
├── repo/
│   ├── repo.go          # AJUSTE: Init byte-compat por modo; no crear cache; .dvcignore/btime
│   ├── git.go           # NUEVO: detección de work tree + `git add` (shell-out)
│   └── doctor.go         # NUEVO: lógica de chequeos (repo/cache/format/remote/coexistence)
└── remote/
    └── s3.go            # +Reachable(ctx) (HeadBucket/list mínimo con timeout) si no existe

tests/
├── oracle/
│   └── init_oracle_test.go   # NUEVO: bytes de `lode init` vs `dvc init` (scm + no-scm)
└── integration/
    └── doctor_test.go        # NUEVO: siembra de problemas + exit codes (MinIO)
```

**Structure Decision**: Continúa el single-project del 001. `init`/`doctor` viven en `internal/cli` cableando lógica de `internal/repo` (Init byte-compat, git, doctor). La reachability del remote se agrega a `internal/remote`. El gate de compat (`init_oracle_test`) aísla el riesgo de byte-identidad, igual que el oráculo del feature 001.

## Complexity Tracking

> No aplica — Constitution Check sin violaciones.
