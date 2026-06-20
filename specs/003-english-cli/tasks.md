---
description: "Task list for 003-english-cli: migrate CLI output to English"
---

# Tasks: CLI en inglés y pulido de mensajes

**Input**: Design documents from `/specs/003-english-cli/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/strings.md, quickstart.md

**Tests**: SE INCLUYE un test de barrido (la spec/quickstart lo define) + re-habilitar `misspell` como gate. Feature de texto, sin TDD por función.

**Organization**: Tareas por user story. MVP = US1 (salida en inglés). Cambio de texto, sin cambios de comportamiento.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: paralelizable (archivos distintos)
- **[Story]**: US1/US2; Setup/Foundational/Polish sin label

## Path Conventions

Single project Go: `internal/cli/`, `internal/remote/`, `tests/`.

---

## Phase 1: Setup

- [X] T001 [P] Agregar helper de pluralización `plural(n, singular, pluralForm)` (y cualquier texto compartido) en `internal/cli/strings.go` per data-model.md

---

## Phase 2: Foundational (gate de verificación)

**Purpose**: el test que valida que la salida quede 100% en inglés. Se escribe ahora; pasa cuando las traducciones aterrizan.

- [X] T002 Test de barrido de inglés: corre `--help` de root y de cada subcomando + caminos de operación/error frecuentes, y afirma que la salida no contiene caracteres no-ASCII de español ni palabras de una lista negra (archivo, remoto, objetos, salidas, materializa, subidos, fallidos, agregado, Continuar, etc.), en `tests/integration/english_cli_test.go` per quickstart / SC-001

**Checkpoint**: el test existe (en rojo hasta traducir) → guía la traducción.

---

## Phase 3: User Story 1 - Interfaz en inglés coherente (Priority: P1) 🎯 MVP

**Goal**: toda la salida de comandos/flags/progreso/resultado en inglés, con plurales naturales.

**Independent Test**: correr `--help` de cada comando y las operaciones normales; verificar inglés (test T002).

### Implementation for User Story 1

- [X] T003 [P] [US1] Traducir root (short) y `add` (short/long, help de flags, mensaje "tracked ->", error "file changed during add") en `internal/cli/root.go` y `internal/cli/add.go` per contracts/strings.md
- [X] T004 [P] [US1] Traducir `status` (short, flag `--json`) en `internal/cli/status.go`
- [X] T005 [US1] Traducir `push`/`fetch`/`pull` (shorts y mensajes de resultado con plurales naturales vía `plural`) en `internal/cli/push.go`
- [X] T006 [US1] Traducir `checkout` (short, "materialized N output(s)") en `internal/cli/checkout.go`
- [X] T007 [US1] Traducir `gc` (short, preview/prompt/cancelled/freed/removed, plurales; mantener que el prompt acepta yes/y) en `internal/cli/gc.go`
- [X] T008 [P] [US1] Traducir el comando `remote` (short, "remote added", help de flags) en `internal/cli/remote.go`

**Checkpoint**: la salida de operación normal y help está 100% en inglés.

---

## Phase 4: User Story 2 - Errores accionables y consistentes (Priority: P2)

**Goal**: errores en inglés, accionables, con terminología canónica consistente.

**Independent Test**: recorrer los errores de precondición y verificar inglés + términos canónicos + acción sugerida.

### Implementation for User Story 2

- [X] T009 [US2] Traducir errores en `internal/cli/transfer.go` ("remote %q is not configured", "manifest %s is not in the cache (add the data first)") y los 3 errores de URL en `internal/remote/s3.go` ("remote has no url", "remote url must start with s3://", "remote url has no bucket") per contracts/strings.md
- [X] T010 [US2] Pasada de consistencia de terminología: aplicar el glosario canónico (remote/cache/workspace/object/output/repository) en TODOS los mensajes y alinear la redacción de los textos del 002 (init/doctor/errores) si difieren, en `internal/cli/*.go` per data-model.md / FR-003

**Checkpoint**: errores en inglés y un único término por concepto en toda la herramienta.

---

## Phase 5: Polish & Cross-Cutting

- [X] T011 [P] Re-habilitar el linter `misspell` en `.golangci.yml` (se había desactivado por el español) y corregir lo que marque; correr `golangci-lint run`
- [X] T012 Ejecutar la validación del quickstart (barrido de `--help` + operaciones + errores) y confirmar el test T002 en verde; build + `go vet` + suite completa

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (T001)**: sin dependencias.
- **Foundational (T002)**: el test de barrido; se escribe tras Setup, queda rojo hasta traducir.
- **US1 (P1)**: usa `plural` (T001). Es el grueso (MVP).
- **US2 (P2)**: independiente de US1 (otros archivos/errores); usa el glosario.
- **Polish**: tras las traducciones (misspell solo tiene sentido con el texto ya en inglés).

### Within Each Story

- US1: las traducciones por archivo (T003/T004/T008 [P]; T005/T006/T007 secuenciales por tocar push/checkout/gc respectivos) hacen pasar T002.
- US2: errores + pasada de glosario.

### Parallel Opportunities

- Setup: T001 [P].
- US1: T003, T004, T008 [P] (archivos distintos). T005/T006/T007 cada uno en su archivo (también [P] entre sí, pero listados sin marca para claridad de orden).
- Polish: T011 [P].

---

## Parallel Example: US1 translations

```bash
Task: "T003 traducir root + add en internal/cli/{root,add}.go"
Task: "T004 traducir status en internal/cli/status.go"
Task: "T008 traducir remote en internal/cli/remote.go"
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Setup (T001) + Foundational (T002, test en rojo).
2. US1 (T003–T008): traducir toda la salida normal/help → T002 pasa.
3. **VALIDATE**: `--help` de cada comando en inglés; operaciones normales en inglés.

### Incremental Delivery

1. Setup + test → gate listo.
2. US1 → salida en inglés (MVP; el grueso del valor de adopción).
3. US2 → errores en inglés + consistencia de glosario.
4. Polish → misspell como gate + validación.

### Riesgo

Bajísimo: cambia texto, no comportamiento. Único cuidado: no alterar la semántica de prompts/flags (p. ej. `gc` sigue aceptando "yes/y"). El test de barrido + `misspell` evitan español remanente.

---

## Notes

- Cero dependencias nuevas, cero cgo, sin cambios de formato (oráculos del 001/002 intactos).
- Preservar los términos compartidos con DVC (remote, cache, push, pull, checkout).
- Marcar cada tarea [X] al completarla.
