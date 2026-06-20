---
description: "Task list for 002-init-onboarding: lode init + doctor + guided errors"
---

# Tasks: Bootstrap standalone y onboarding (init + doctor)

**Input**: Design documents from `/specs/002-init-onboarding/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli.md, quickstart.md

**Tests**: SE INCLUYEN tareas de test focalizadas (la spec define una suite: oráculo de bytes de `init`, integración de `doctor`, interop). Gates de compat/viabilidad, no TDD exhaustivo.

**Organization**: Tareas por user story. MVP = US1 (`lode init`).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: paralelizable (archivos distintos, sin dependencias pendientes)
- **[Story]**: US1/US2/US3; Setup/Foundational/Polish sin label

## Path Conventions

Single project Go (continúa la base del 001): `cmd/lode/`, `internal/`, `tests/`.

---

## Phase 1: Setup

**Purpose**: datos compartidos. Sin dependencias nuevas (reusa toda la infra del 001).

- [ ] T001 [P] Definir las constantes byte-exactas de los artefactos de `init` (template `.dvcignore` de 139 bytes; contenido de `.dvc/.gitignore` = `/config.local\n/tmp\n/cache\n`; config por modo: vacío en scm, `[core]\n    no_scm = True\n` en no-scm) en `internal/repo/initdata.go` per research.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: primitivas de `init` compartidas + gate de byte-compat. Bloquea las user stories.

**⚠️ CRITICAL**: el gate T004 (oráculo de `init`) debe pasar antes de cablear comandos.

- [ ] T002 Reescribir `repo.Init` para byte-compat por modo (scm/no-scm): escribe el config correcto, `.dvcignore`, `.dvc/tmp/btime` (vacío); NO crea `.dvc/cache` (lazy); detecta repo existente en el cwd o en un ancestro (reusando `repo.Find`) devolviendo el `InitOutcome` correspondiente, en `internal/repo/repo.go` per data-model.md
- [ ] T003 [P] Detección de git work tree y `git add` de los archivos versionables (shell-out a `git`, solo modo scm) en `internal/repo/git.go` per research.md §2/§3
- [ ] T004 **Gate oráculo de bytes de `init`**: test que compara la estructura que crea lode (vía `repo.Init`) contra `dvc init` y `dvc init --no-scm` reales — bytes idénticos de `config`, `.dvc/.gitignore`, `.dvcignore`, presencia de `.dvc/tmp/btime` vacío y ausencia de `.dvc/cache`, en `tests/oracle/init_oracle_test.go` per quickstart Escenario 1 ⚠️ GATE

**Checkpoint**: `repo.Init` byte-idéntico a DVC en ambos modos → las stories pueden empezar.

---

## Phase 3: User Story 1 - Empezar de cero sin Python (Priority: P1) 🎯 MVP

**Goal**: `lode init` deja un repo operable sin Python; el flujo init→add→push funciona solo con el binario.

**Independent Test**: en un entorno sin DVC/Python, `lode init --no-scm` → `lode add` → `lode push` completa; un DVC real opera el repo resultante.

### Tests for User Story 1

- [ ] T005 [P] [US1] Test de interop: un repo creado por `lode init` (ambos modos) es operado por `dvc` sin errores, y un repo de `dvc init` es operado por lode, en `tests/integration/init_interop_test.go` per SC-002

### Implementation for User Story 1

- [ ] T006 [US1] Comando `init`: flag `--no-scm`; selección de modo (git presente vs `--no-scm`); manejo de los `InitOutcome` (created / already-initialized / inside-existing-repo / needs-no-scm con error accionable); `git add` en modo scm; mensaje de siguiente paso (`lode add <target>`), en `internal/cli/init.go` per contracts/cli.md
- [ ] T007 [US1] Cablear `init` en el root command y validar el flujo end-to-end del quickstart Escenario 2 (cero a pusheado sin Python) en `internal/cli/root.go`

**Checkpoint**: `lode init` funcional → un usuario nuevo arranca sin Python. MVP del feature.

---

## Phase 4: User Story 2 - Errores que guían (Priority: P2)

**Goal**: los comandos que requieren repo, y los errores de precondición frecuentes, nombran el siguiente paso concreto.

**Independent Test**: correr comandos fuera de un repo / sin remote y verificar que el mensaje sugiere el comando exacto.

### Tests for User Story 2

- [ ] T008 [P] [US2] Test: comandos que requieren repo, ejecutados fuera de un repo, sugieren `lode init` (y `--no-scm`/`--cd`); `push` sin remote sugiere configurar uno, en `tests/integration/guided_errors_test.go` per SC-003

### Implementation for User Story 2

- [ ] T009 [US2] Helper de errores guiados: `requireRepo` (mensaje accionable que nombra `lode init`, `--no-scm` y `--cd <dir>` cuando no hay repo) + hints reutilizables (sin remote → comando para configurarlo; objeto faltante → `lode pull`), en `internal/cli/errors.go` per contracts/cli.md
- [ ] T010 [US2] Aplicar `requireRepo` y los hints en `add`/`status`/`checkout`/`push`/`fetch`/`pull`/`gc` (reemplazar el `findRepo` directo y enriquecer los errores de precondición) en `internal/cli/*.go` per FR-007/FR-008
- [ ] T011 [US2] Validar quickstart Escenario 3 (errores que guían)

**Checkpoint**: ningún comando deja al usuario en un callejón sin salida.

---

## Phase 5: User Story 3 - Diagnóstico con `doctor` (Priority: P2)

**Goal**: `lode doctor` reporta estado de repo/cache/format/remotes/coexistencia con sugerencias y exit code correcto.

**Independent Test**: sembrar cada clase de problema y verificar detección + sugerencia + exit code.

### Tests for User Story 3

- [ ] T012 [P] [US3] Test integración `doctor`: sembrar sin-repo, sin-remote, remote-inalcanzable (MinIO apagado), cache-no-escribible y formato-legacy 2.x; verificar que cada uno se detecta con su sugerencia y el exit code correcto, en `tests/integration/doctor_test.go` per SC-004

### Implementation for User Story 3

- [ ] T013 [P] [US3] `remote.S3.Reachable(ctx)`: chequeo de alcanzabilidad (existencia de bucket / list con `MaxKeys=1`) bajo timeout corto, distinguiendo inalcanzable/credenciales de OK, en `internal/remote/s3.go` per research.md §5
- [ ] T014 [US3] Lógica de chequeos (repo, cache escribible, formato 3.x vs legacy 2.x, remotes vía Reachable, coexistencia vía lock) que produce el `DoctorReport`, en `internal/repo/doctor.go` per data-model.md
- [ ] T015 [US3] Comando `doctor`: `--json`, `-r/--remote`, reporte legible (OK/warn/problem + sugerencia), exit 0 si sano / ≠0 si hay problema bloqueante, en `internal/cli/doctor.go` per contracts/cli.md
- [ ] T016 [US3] Cablear `doctor` en el root y validar quickstart Escenario 5

**Checkpoint**: las 3 user stories funcionan de forma independiente.

---

## Phase 6: Polish & Cross-Cutting

**Purpose**: documentación y validación final.

- [ ] T017 [P] Actualizar `README.md`: agregar `lode init` al install/uso, mostrar el flujo standalone (init → add → push) y quitar cualquier dependencia implícita de `dvc init`
- [ ] T018 [P] Revisar `specs/001-dvc-go/quickstart.md` y otros docs que asuman `dvc init`, alineándolos a `lode init`
- [ ] T019 Ejecutar la suite completa del quickstart (5 escenarios) end-to-end y registrar resultados

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sin dependencias.
- **Foundational (Phase 2)**: depende de Setup. **Bloquea las stories.** El gate T004 debe pasar antes de cablear comandos.
- **US1 (P1)**: Foundational. Es el MVP (destraba la adopción).
- **US2 (P2)**: Foundational. Independiente de US1 (testeable con `repo.Init`/`repo.Find`).
- **US3 (P2)**: Foundational + `remote.Reachable`. Independiente de US1/US2.
- **Polish**: tras las stories deseadas.

### User Story Dependencies

- US1, US2, US3 son independientes entre sí una vez completado Foundational. US1 es el MVP.

### Within Each Story

- Tests primero; primitivas (Foundational) antes que comandos; lógica antes del wiring en el root.

### Parallel Opportunities

- Setup: T001 [P].
- Foundational: T003 [P] junto con T002; T004 tras T002/T003.
- Tests de cada story ([P]): T005, T008, T012; y T013 [P].
- Polish: T017, T018 [P].

---

## Parallel Example: Foundational + tests

```bash
# Tras T002, en paralelo:
Task: "T003 detección git + git add en internal/repo/git.go"
# Tests de stories (una vez Foundational listo), en paralelo:
Task: "T005 interop init en tests/integration/init_interop_test.go"
Task: "T008 errores guiados en tests/integration/guided_errors_test.go"
Task: "T012 doctor en tests/integration/doctor_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Setup (T001) + Foundational (T002–T004) — **no avanzar sin el gate T004 (oráculo de init) en verde**.
2. US1 (T005–T007): `lode init`.
3. **STOP & VALIDATE**: quickstart Escenario 2 (cero a pusheado sin Python) + Escenario 1 (oráculo).
4. Demo: "instalá el binario, `lode init`, `lode add`, `lode push` — sin Python".

### Incremental Delivery

1. Setup + Foundational → base.
2. US1 → bootstrap standalone (MVP, destraba adopción).
3. US2 → errores que guían (mejor onboarding).
4. US3 → `doctor` (diagnóstico).
5. Polish → docs + validación.

### Riesgo crítico

El gate T004 valida la byte-identidad de la estructura de `init` con `dvc init` en ambos modos. Si no pasa, la compat drop-in del onboarding no es válida: resolverlo antes de cablear comandos.

---

## Notes

- [P] = archivos distintos, sin dependencias pendientes.
- Sin dependencias nuevas; el shell-out a `git` solo ocurre en modo scm.
- Compatibilidad byte-a-byte con `dvc init` prevalece (Constitución I).
- Cero cgo en toda la cadena.
