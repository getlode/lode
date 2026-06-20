# Implementation Plan: CLI en inglés y pulido de mensajes

**Branch**: `003-english-cli` | **Date**: 2026-06-20 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/003-english-cli/spec.md`

## Summary

Migrar toda la salida visible de la CLI (descripciones de comandos/flags, mensajes de progreso/resultado, errores) del español al inglés, con terminología canónica consistente y plurales naturales. Es un cambio de texto, no de comportamiento: no toca la semántica de comandos ni el formato de archivos. ~28 cadenas en `internal/cli` + `internal/remote/s3.go` (inventario en [research.md](research.md)). Se verifica con un test de barrido de la salida y re-habilitando el linter `misspell` (desactivado mientras el texto era español).

## Technical Context

**Language/Version**: Go 1.23+ (`CGO_ENABLED=0`)

**Primary Dependencies**: ninguna nueva. Reusa cobra. Helper de pluralización propio.

**Storage**: N/A (no toca archivos ni formato).

**Testing**: `tests/integration/english_cli_test.go` (barrido de `--help` + errores, sin caracteres/palabras en español); `misspell` re-habilitado en `.golangci.yml` (gate en CI).

**Target Platform**: igual que el resto (Linux/macOS/Windows × amd64/arm64).

**Project Type**: CLI (single project, continúa la base 001/002).

**Performance Goals**: N/A.

**Constraints**: cero cambios de comportamiento; mantener la semántica de prompts/flags (p. ej. `gc` sigue aceptando "yes/y"); preservar los términos compartidos con DVC (remote, cache, push, pull, checkout). Cero deps nuevas, cero cgo.

**Scale/Scope**: ~28 cadenas en ~10 archivos de `internal/cli` + 3 errores en `internal/remote/s3.go`. Sin cambios de API.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Evaluado contra `.specify/memory/constitution.md` (v1.0.0):

- **I. DVC Byte-Compatibility**: ✅ N/A — no cambia ningún artefacto de formato (solo texto a stdout/stderr).
- **II. Oracle-Gated Format Changes**: ✅ N/A — no hay cambios de formato; los oráculos existentes siguen verdes.
- **III. Zero-CGO Single Binary**: ✅ sin deps nuevas.
- **IV. Performance Is the Product**: ✅ N/A.
- **V. Coexistence Over Reinvention**: ✅ se preservan los términos que usa DVC; no cambia estado compartido.

**Sin violaciones.** Complexity Tracking vacío.

## Project Structure

### Documentation (this feature)

```text
specs/003-english-cli/
├── plan.md
├── research.md         # inventario de cadenas + glosario + verificación
├── data-model.md       # glosario canónico
├── contracts/
│   └── strings.md      # tabla ES → EN de referencia
└── quickstart.md       # validación (barrido de inglés)
```

### Source Code (cambios)

```text
internal/cli/
├── root.go, add.go, status.go, push.go, checkout.go, gc.go, remote.go, transfer.go
│                       # traducir short/long, flags, progreso, errores
└── strings.go          # NUEVO (opcional): helper de pluralización + textos compartidos
internal/remote/s3.go   # traducir 3 errores de parseo de url
.golangci.yml           # re-habilitar `misspell`
tests/integration/
└── english_cli_test.go # NUEVO: barrido de salida sin español
```

**Structure Decision**: Cambios localizados en `internal/cli` (donde viven los textos) más 3 errores en `internal/remote/s3.go`. Un `strings.go` opcional centraliza el helper de plural. El gate de regresión es el test de barrido + `misspell` en CI. No se introducen capas ni dependencias.

## Complexity Tracking

> No aplica — Constitution Check sin violaciones.
