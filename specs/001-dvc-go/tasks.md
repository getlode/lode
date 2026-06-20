---
description: "Task list for 001-dvc-go: núcleo de versionado de datos en Go, drop-in DVC"
---

# Tasks: Versionado de datos de alta velocidad, drop-in compatible con DVC

**Input**: Design documents from `/specs/001-dvc-go/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli.md, quickstart.md

**Tests**: SE INCLUYEN tareas de test focalizadas porque la spec define una suite de validación explícita (oráculo de bytes, integración MinIO, interoperabilidad bidireccional). No es TDD exhaustivo: se testean los gates de viabilidad y compatibilidad, no cada función.

**Organization**: Tareas agrupadas por user story. MVP = User Story 1.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Puede correr en paralelo (archivos distintos, sin dependencias pendientes)
- **[Story]**: US1/US2/US3/US4 (fases de user story); Setup/Foundational/Polish sin label

## Path Conventions

Single project Go: `cmd/dvcgo/`, `internal/`, `tests/` en la raíz del repo (ver plan.md).

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Inicialización del proyecto Go y tooling.

- [X] T001 Inicializar módulo Go y estructura de directorios (`go mod init`, crear `cmd/dvcgo/`, `internal/{cli,repo,dvcfile,hashfile,cache,remote,transfer,checkout,lock}/`, `tests/{oracle,integration}/`) per plan.md
- [X] T002 Agregar dependencias en `go.mod`: `spf13/cobra`, `aws-sdk-go-v2/{config,service/s3,feature/s3/transfermanager,credentials}`, `golang.org/x/sync`, `golang.org/x/sys`, `go.etcd.io/bbolt`, `gofrs/flock`, `gopkg.in/ini.v1`, `testcontainers-go/modules/minio`
- [X] T003 [P] Configurar linting/formato: `.golangci.yml` + `gofumpt`, y `Makefile` con targets `build/test/lint/oracle`
- [X] T004 [P] Crear `.goreleaser.yaml` (matriz linux/darwin/windows × amd64/arm64, `CGO_ENABLED=0`, `-s -w`, brews/homebrew-tap) y workflow `.github/workflows/release.yml` disparado en tag `v*`
- [X] T005 [P] Crear `scripts/gen-dataset.sh` (genera datasets deterministas de N archivos) para benchmarks y oráculo, referenciado en quickstart.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Primitivas compartidas por TODAS las stories. Incluye el **gate de viabilidad** (oráculo de bytes del `.dir` y del `.dvc`), que es el riesgo #1 del proyecto.

**⚠️ CRITICAL**: Ninguna user story puede empezar hasta completar esta fase. El gate T016 debe pasar antes de invertir en comandos.

- [X] T006 Esqueleto CLI con cobra: root command + flags persistentes globales (`-v/--verbose`, `-q/--quiet`, `-j/--jobs`, `--cd`) en `internal/cli/root.go` y `cmd/dvcgo/main.go` per contracts/cli.md
- [X] T007 Descubrimiento del repo y rutas (`.dvc/` hacia arriba en el árbol, paths de cache/tmp) en `internal/repo/repo.go` per data-model.md
- [X] T008 [P] Parser/writer de `.dvc/config` (INI: `[core]`, `[cache]`, secciones `['remote "name"']`) en `internal/repo/config.go` per data-model.md (RepoConfig)
- [X] T009 [P] Infra de errores y logging: errores accionables sin stack traces en uso normal, modo verbose, salida humana/`--json` en `internal/cli/output.go`
- [X] T010 Paquete de locks: lock global exclusivo `flock(LOCK_EX)` sobre `.dvc/tmp/lock` (timeout 3s, reintentos) + rwlock JSON `.dvc/tmp/rwlock` (read/write por path, purga de PIDs muertos) en `internal/lock/lock.go` per research §6
- [X] T011 [P] Hashing MD5 streaming (chunks 1 MiB, `crypto/md5`) + worker pool (`errgroup.SetLimit(NumCPU)` + `sync.Pool` de buffers + `io.CopyBuffer`) en `internal/hashfile/hash.go` per research §2/§8
- [X] T012 [P] Serialización exacta del objeto `.dir` (DirManifest): JSON con separadores `", "`/`": "`, claves alfabéticas (`md5` antes de `relpath`), escape ASCII `\uXXXX` para >0x7F, orden ascendente por `relpath`, sin newline final, oid con sufijo `.dir`, en `internal/hashfile/tree.go` per research §2 ⚠️ RIESGO #1
- [X] T013 [P] State DB con bbolt: clave path → `{ino, mtime, size, md5}`, get/set, invalidación si `(ino,mtime,size)` difiere, en `internal/hashfile/state.go` (`.dvc/tmp/dvcgo/state.db`) per research §5
- [X] T014 Lectura/escritura byte-compatible de archivos `.dvc` (YAML 3.x, orden de claves `md5,size,nfiles,hash,path`, 2 espacios indent, un newline final) en `internal/dvcfile/dvcfile.go` per research §1 / data-model.md
- [X] T015 Cache store content-addressed: paths `files/md5/<2>/<resto>`, escritura atómica tmp+`os.Rename`, `protect` 0o444, `Has`/`Get`/`Add`, en `internal/cache/cache.go` per research §3
- [X] T016 [P] **Gate oráculo de bytes**: test que genera `.dvc` + objeto `.dir` con DVC real (vía `scripts/gen-dataset.sh` + fixtures) y compara byte-a-byte contra los producidos por `dvcfile` y `hashfile/tree`, en `tests/oracle/oracle_test.go` (cubre archivo individual, dir simple, dir con subdirs y nombres Unicode) per quickstart Escenario 1

**Checkpoint**: Primitivas listas y oráculo en verde → las user stories pueden comenzar.

---

## Phase 3: User Story 1 - Versionar datasets a alta velocidad (Priority: P1) 🎯 MVP

**Goal**: Trackear archivos/directorios y consultar estado en un repo DVC existente, byte-compatible y ≥10× más rápido.

**Independent Test**: Sobre un repo DVC real con dataset grande, `dvcgo add` produce `.dvc`/`.dir` idénticos a DVC y completa ≥10× más rápido; `dvcgo status` no rehashea si nada cambió.

### Tests for User Story 1

- [X] T017 [P] [US1] Test de no-rehash: tras `add`, modificar solo mtime y verificar que `status` resuelve vía state DB sin recalcular hash, en `tests/oracle/status_norehash_test.go` per SC-005
- [X] T018 [P] [US1] Test de `.gitignore`: `add data/foo.csv` crea `data/.gitignore` con línea `/foo.csv` idempotente, en `internal/checkout/gitignore_test.go` per research §7

### Implementation for User Story 1

- [X] T019 [P] [US1] Gestión de `.gitignore` por directorio (`<dir>/.gitignore`, entrada `"/" + relpath` POSIX, append idempotente preservando newline) en `internal/checkout/gitignore.go` per research §7
- [X] T020 [US1] Comando `add`: hashear (paralelo, T011), construir `.dir` para directorios (T012), mover contenido al cache (T015, 0o444), escribir `.dvc` (T014), actualizar `.gitignore` (T019), tomar lock (T010); falla segura si el archivo cambia durante el add. En `internal/cli/add.go` per contracts/cli.md
- [X] T021 [US1] Comando `status`: comparar workspace vs `.dvc` usando state DB (T013); estados `not in cache`/`modified`/`new`/`deleted`/`up to date`; salida humana y `--json`; sin escribir nada. En `internal/cli/status.go` per contracts/cli.md
- [X] T022 [US1] Cablear `add` y `status` en el root command (T006) y validar flujo end-to-end del quickstart Escenario 2 en `internal/cli/root.go`

**Checkpoint**: `add` + `status` funcionales y testeables de forma independiente → MVP demostrable.

---

## Phase 4: User Story 2 - Compartir datos vía remote S3-compatible (Priority: P1)

**Goal**: `push`/`pull`/`fetch` contra AWS S3, MinIO, R2, B2, interoperable con el cache/remote de DVC-Python.

**Independent Test**: Configurar MinIO, `add`+`push`, borrar cache, `pull` en clon limpio → datos íntegros; verificar interop bidireccional con `dvc` de Python sobre el mismo remote.

### Tests for User Story 2

- [X] T023 [P] [US2] Test de integración round-trip con MinIO (testcontainers): `add`→`push`→borrar cache/data→`pull`, integridad 100%, en `tests/integration/roundtrip_test.go` (gateado con `testing.Short()`) per SC-003 / quickstart Escenario 3
- [X] T024 [P] [US2] Test de interoperabilidad bidireccional: `dvcgo push` + `dvc pull` (Python) y viceversa sobre el mismo bucket, en `tests/integration/interop_test.go` per SC-002
- [X] T025 [P] [US2] Test de reanudación: matar `push` a mitad y reintentar → sin objetos corruptos, converge a remote íntegro, en `tests/integration/resume_test.go` per SC-007 / quickstart Escenario 4

### Implementation for User Story 2

- [X] T026 [P] [US2] Backend S3 con `aws-sdk-go-v2` + `feature/s3/transfermanager` (`BaseEndpoint`, `UsePathStyle`, credenciales env/perfil/estáticas, `sse`/`acl`), mapeo objeto→key `<prefix>/files/md5/<2>/<resto>`, en `internal/remote/s3.go` per research §3/§8
- [X] T027 [US2] Status del remote (`oidsExist`): estrategia HEAD-por-objeto vs LIST masivo por prefijo `00..ff` con heurística (`TRAVERSE_THRESHOLD_SIZE=500000`, `TRAVERSE_PREFIX_LEN=2`, `TRAVERSE_WEIGHT_MULTIPLIER=5`, `LIST_OBJECT_PAGE_SIZE=1000`), en `internal/remote/status.go` per research §4
- [X] T028 [US2] Motor de transferencia `push`: dirs antes que archivos, contenidos antes del `.dir`, `.dir` solo si todos los contenidos OK; concurrencia (`-j`); idempotente/reanudable; en `internal/transfer/push.go` per research §4
- [X] T029 [US2] `fetch` (remote→cache): descargar solo faltantes, escritura atómica tmp+rename, verificación de integridad (re-hash, borrar+fallar si mismatch) en `internal/transfer/fetch.go` per research §4
- [X] T030 [US2] Materialización básica por copia (helper compartido para `pull`) en `internal/checkout/materialize.go` (será enriquecido con link strategies en US3)
- [X] T031 [US2] Comandos `push`, `fetch`, `pull` (= fetch + materialize), flag `-r/--remote`, lock; y comandos `remote add`/`remote modify` (editan `.dvc/config`, T008) en `internal/cli/{push,pull,fetch,remote}.go` per contracts/cli.md
- [X] T032 [US2] Cablear comandos de US2 en el root y validar quickstart Escenario 3 en `internal/cli/root.go`

**Checkpoint**: US1 + US2 funcionan de forma independiente; flujo completo local↔remote operativo.

---

## Phase 5: User Story 3 - Restaurar y cambiar de versión el workspace (Priority: P2)

**Goal**: Comando `checkout` con estrategias de link eficientes y detección de relink.

**Independent Test**: Con dos versiones registradas, `checkout` de cada una deja el workspace exacto, materializando desde cache con la estrategia más eficiente disponible y degradando a copy.

### Tests for User Story 3

- [X] T033 [P] [US3] Test de estrategias de link: reflink cuando el FS lo soporta, fallback a copy; objeto de cache queda 0o444 en links; archivos ya coincidentes no se tocan; en `internal/checkout/link_test.go` per FR-017
- [X] T034 [P] [US3] Test de objeto faltante: `checkout` con objeto ausente en cache informa qué falta y sugiere `pull`, en `internal/cli/checkout_test.go` per contracts/cli.md

### Implementation for User Story 3

- [X] T035 [P] [US3] Link strategies: reflink (`golang.org/x/sys/unix` `IoctlFileClone`/FICLONE), fallback `CopyFileRange` → `io.Copy`; hardlink/symlink honrando `cache.type`; detección de errores (`EOPNOTSUPP`/`EXDEV`/`EINVAL`) en `internal/checkout/link.go` per research §7
- [X] T036 [US3] Detección de relink (hardlink: inodo difiere; symlink: target difiere; copy/reflink: no relink) y re-protección 0o444 en `internal/checkout/relink.go` per research §7
- [X] T037 [US3] Comando `checkout`: materializar según `cache.type` (default `reflink,copy`), saltar archivos coincidentes, eliminar los que ya no están en el manifest, reportar faltantes; refactorizar `materialize.go` (T030) para usar las link strategies. En `internal/cli/checkout.go` per contracts/cli.md
- [X] T038 [US3] Cablear `checkout` en el root y validar quickstart Escenario 5 en `internal/cli/root.go`

**Checkpoint**: US1 + US2 + US3 funcionales; ciclo de versionado completo (registrar → compartir → cambiar versión).

---

## Phase 6: User Story 4 - Recuperar espacio con gc (Priority: P3)

**Goal**: Eliminar objetos no referenciados de forma segura.

**Independent Test**: Crear varias versiones, desreferenciar algunas, `gc` elimina solo las no referenciadas; versiones vigentes siguen restaurables.

### Tests for User Story 4

- [X] T039 [P] [US4] Test de seguridad de gc: solo se borran objetos no referenciados, vigentes restaurables; sin `-f` pide confirmación; en `tests/integration/gc_test.go` per FR-019/FR-020 / quickstart Escenario 6

### Implementation for User Story 4

- [X] T040 [US4] Cálculo de alcanzabilidad: oids referenciados desde los `.dvc` vigentes del workspace, expandiendo los `.dir` a sus contenidos, en `internal/transfer/reachable.go` (o `internal/cache/gc.go`)
- [X] T041 [US4] Comando `gc`: alcance `-w/--workspace`, `-c/--cloud` (también remote, T026), `-r/--remote`, `-f/--force`; mostrar alcance y pedir confirmación sin `-f`; borrar del cache (y remote si `-c`); reportar espacio recuperado. En `internal/cli/gc.go` per contracts/cli.md
- [X] T042 [US4] Cablear `gc` en el root y validar quickstart Escenario 6 en `internal/cli/root.go`

**Checkpoint**: Los 4 user stories independientemente funcionales.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Mejoras transversales y validación final.

- [X] T043 [P] Benchmark de performance que valida SC-001 (≥10× vs DVC en ≥100k archivos) y SC-005/SC-006, en `tests/integration/bench_test.go` + reporte en `Makefile`
- [X] T044 [P] Completions de shell (bash/zsh/fish) y `man` vía cobra en `internal/cli/completion.go`
- [X] T045 [P] README con quickstart copy-paste, tabla de comandos y matriz de compatibilidad DVC, en `README.md`
- [X] T046 [P] CI: workflow `.github/workflows/ci.yml` (matriz OS × arch, `go test` con y sin `-short`, lint, oráculo) 
- [X] T047 Manejo de repos legacy 2.x: lectura básica del cache plano (`<2>/<resto>`, `md5-dos2unix`) con mensaje claro si se detecta, en `internal/cache/legacy.go` per research §3 (solo lectura, fuera de hot path)
- [X] T048 Ejecutar la suite completa de `quickstart.md` (6 escenarios) end-to-end y registrar resultados

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sin dependencias.
- **Foundational (Phase 2)**: depende de Setup. **BLOQUEA todas las stories.** El gate T016 debe pasar antes de las fases de comando.
- **User Stories (Phase 3-6)**: dependen de Foundational.
  - US1 (P1) y US2 (P1) son el corazón del producto. US2 reutiliza primitivas de Foundational; su `pull` usa la materialización básica T030 (independiente de US3).
  - US3 (P2) enriquece la materialización (refactoriza T030) — depende conceptualmente de que exista T030, pero es testeable de forma aislada.
  - US4 (P3) usa el cache/remote de Foundational/US2.
- **Polish (Phase 7)**: depende de las stories deseadas completas.

### User Story Dependencies

- **US1**: solo Foundational. Sin dependencia de otras stories.
- **US2**: Foundational. Independientemente testeable (puede operar sobre un repo cuya metadata creó DVC-Python).
- **US3**: Foundational + helper T030 de US2 (refactor). Testeable de forma aislada con datos ya en cache.
- **US4**: Foundational (+ remote de US2 si se usa `-c`).

### Within Each Story

- Tests primero (deben fallar), luego implementación.
- Primitivas (Foundational) antes que comandos.
- En cada comando: lógica de dominio antes del wiring en el root.

### Parallel Opportunities

- Setup: T003, T004, T005 en paralelo.
- Foundational: T008, T009, T011, T012, T013 en paralelo (archivos distintos); T016 (oráculo) en paralelo una vez que T012/T014 existen. T014 y T015 dependen de las primitivas que consumen.
- US1: T017/T018 (tests) en paralelo; T019 en paralelo con los tests.
- US2: T023/T024/T025 (tests) y T026 en paralelo.
- US3: T033/T034 (tests) y T035 en paralelo.
- Polish: T043/T044/T045/T046 en paralelo.

---

## Parallel Example: Foundational primitives

```bash
# Tras T006/T007, lanzar primitivas independientes juntas:
Task: "T011 hashing MD5 + worker pool en internal/hashfile/hash.go"
Task: "T012 serialización .dir en internal/hashfile/tree.go"
Task: "T013 state DB bbolt en internal/hashfile/state.go"
Task: "T008 parser .dvc/config en internal/repo/config.go"
Task: "T009 infra errores/logging en internal/cli/output.go"
```

## Parallel Example: User Story 2 tests

```bash
Task: "T023 round-trip MinIO en tests/integration/roundtrip_test.go"
Task: "T024 interop bidireccional en tests/integration/interop_test.go"
Task: "T025 reanudación en tests/integration/resume_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Phase 1 Setup.
2. Phase 2 Foundational — **no avanzar sin el gate T016 (oráculo de bytes) en verde**: es el test de viabilidad del proyecto entero.
3. Phase 3 US1 (`add` + `status`).
4. **STOP & VALIDATE**: quickstart Escenarios 1 y 2; medir el ≥10×.
5. Demo: "DVC pero 10× más rápido, sobre tu repo actual, sin migrar".

### Incremental Delivery

1. Setup + Foundational → base lista.
2. US1 → versionado local rápido (MVP).
3. US2 → compartir vía S3-compatible (el segundo P1, completa el caso de equipos).
4. US3 → checkout con link strategies.
5. US4 → gc.
6. Polish → distribución (GoReleaser/Homebrew), docs, benchmarks.

### Riesgo crítico

El gate T016 valida la serialización exacta del `.dir` (research §2). Si no pasa, la propuesta drop-in no es viable: resolverlo ANTES de construir comandos. Es el primer entregable real tras las primitivas T012/T014.

---

## Notes

- [P] = archivos distintos, sin dependencias pendientes.
- Tests incluidos solo en los gates de compatibilidad/viabilidad (oráculo, integración, interop, seguridad de gc), no exhaustivos.
- Compatibilidad byte-a-byte con DVC prevalece sobre mejoras de diseño (invariante de spec).
- Toda la cadena es `CGO_ENABLED=0` para el binario único cross-compile.
- Commit por tarea o grupo lógico; parar en cada checkpoint para validar la story.
