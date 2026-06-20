# Research: Bootstrap standalone y onboarding (init + doctor)

**Feature**: 002-init-onboarding | **Date**: 2026-06-20 | **Phase**: 0

Verificado en vivo contra DVC 3.67.1 (`dvc init` y `dvc init --no-scm`). Reutiliza la infraestructura del feature 001 (repo, cache, lock, remote, dvcfile).

---

## 1. Estructura exacta que crea `dvc init`

**Decision**: Replicar byte-a-byte ambos modos.

**Modo con git (`dvc init`)**:
- `.dvc/config` — **vacío** (0 bytes).
- `.dvc/.gitignore` — `"/config.local\n/tmp\n/cache\n"`.
- `.dvcignore` (raíz) — template de 139 bytes con 3 líneas de comentario, terminado en `\n`:
  ```
  # Add patterns of files dvc should ignore, which could improve
  # the performance. Learn more at
  # https://dvc.org/doc/user-guide/dvcignore
  ```
- `.dvc/tmp/btime` — archivo **vacío** (marcador).
- DVC además hace `git add` de `.dvc/.gitignore`, `.dvc/config` y `.dvcignore`.

**Modo sin git (`dvc init --no-scm`)**:
- `.dvc/config` — `"[core]\n    no_scm = True\n"` (indent de 4 espacios).
- `.dvcignore` (raíz) — mismo template de 139 bytes.
- `.dvc/tmp/btime` — vacío.
- **No** crea `.dvc/.gitignore` (no hay git que ignorar).

**En ambos**: NO se crea `.dvc/cache` en init (se crea perezosamente en el primer `add`).

**Rationale**: SC-006 exige byte-identidad con `dvc init`; FR-004 exige interoperabilidad. Un init que difiera rompería el invariante (Principio I).

**Alternatives considered**: crear `cache/` eagermente (rechazado: diverge de DVC). Formato propio (rechazado: rompe drop-in).

> Nota de implementación: el `repo.Init` actual (feature 001) crea `.dvc/cache` y un `config` vacío sin `.dvcignore`/`btime`. Hay que ajustarlo para igualar exactamente cada modo y NO crear cache en init.

## 2. Selección de modo y detección de git

**Decision**: `lode init` detecta si el cwd está dentro de un work tree de git; si lo está → modo con git; si no → requiere `--no-scm` (igual que DVC), con error accionable.

**Rationale**: DVC sin git y sin `--no-scm` **falla** ("failed to initiate dvc"). Igualamos ese contrato, pero el error guía al usuario hacia `lode init --no-scm` o `git init`.

Detección de git: chequear la presencia de un work tree (existe `.git` hacia arriba, o `git rev-parse --is-inside-work-tree`). Shell-out a `git` es aceptable aquí: el modo con git presupone git instalado; el camino standalone (`--no-scm`) **no** invoca git.

**Alternatives considered**: auto-`--no-scm` cuando no hay git (rechazado: diverge de DVC y puede sorprender; preferimos error guiado).

## 3. Integración con git (modo scm)

**Decision**: tras crear los archivos en modo git, hacer `git add` de `.dvc/.gitignore`, `.dvc/config` y `.dvcignore`, replicando a DVC.

**Rationale**: paridad de comportamiento — el usuario que viene de DVC espera que init deje los archivos staged. Shell-out a `git add` (git ya está presente).

**Alternatives considered**: no stagear (rechazado: diverge; el usuario tendría que recordarlo). Usar una librería git en Go (rechazado: innecesario y agrega peso; shell-out es suficiente y sin cgo).

## 4. Seguridad ante repos existentes (FR-005/006)

**Decision**: antes de inicializar, `lode init` descubre repos existentes:
- Si el cwd ya tiene `.dvc/` → informar "ya inicializado", no tocar nada (salvo `--force`, opcional, fuera del MVP).
- Si está dentro de un repo (hay `.dvc/` en un ancestro) → informar la ubicación del repo padre y no crear uno anidado.

Reusa `repo.Find` (feature 001) para detectar el repo más cercano hacia arriba.

## 5. `lode doctor` — chequeos

**Decision**: un comando que corre una batería de chequeos y reporta estado + sugerencia por cada uno; exit 0 sano, ≠0 si hay bloqueante.

Chequeos (cada uno OK / warning / problem + sugerencia):
- **Repo**: ¿hay `.dvc/`? (si no → sugerir `lode init`). Validez básica (config legible).
- **Cache**: ubicación y si es escribible (intento de escritura/borrado de un tmp en `.dvc/cache` o el dir configurado).
- **Formato**: ¿layout 3.x (`files/md5/`) o legacy 2.x (`<2>/<resto>`)? Reportar y, si legacy, explicar (lectura soportada).
- **Remotes**: por cada remote configurado, alcanzabilidad — establecer conexión + autenticar (p. ej. un `HeadBucket`/list mínimo con timeout). Distinguir "no configurado" de "configurado pero inalcanzable/credenciales".
- **Coexistencia DVC**: ¿el lock global está libre o tomado por otro proceso? ¿hay un DVC corriendo? Reportar si es seguro operar.

**Rationale**: cubre las causas reales de fricción detectadas (no arranca, no pushea, cache no escribe, repo de otra versión). Reusa `remote.NewS3` + un check de alcanzabilidad, `repo.LoadConfig`, `lock`.

**Reachability del remote**: usar una operación liviana del SDK (existencia de bucket / list con `MaxKeys=1`) bajo un `context` con timeout corto. No verifica integridad de objetos (eso es `lode verify`, feature 004).

**Alternatives considered**: doctor con auto-reparación (rechazado en MVP: solo sugiere; reparar es del usuario). Ping de remote bajando un objeto (rechazado: costoso; basta autenticar+listar).

## 6. Errores que guían (FR-007/008)

**Decision**: centralizar la resolución de repo en un helper que, ante "no repo", devuelve un mensaje accionable nombrando `lode init` (y `--cd <dir>`); y enriquecer los errores de precondición existentes (sin remote → comando para configurarlo; objeto faltante → `lode pull`).

**Rationale**: convierte callejones sin salida en el siguiente paso obvio. Bajo costo, alto impacto en onboarding.

## 7. Stack / reutilización

**Decision**: sin dependencias nuevas. Reusa cobra (comandos), `internal/repo` (Init/Find/Config), `internal/remote` (alcanzabilidad S3 vía minio-go), `internal/lock` (coexistencia), `internal/cache` (escribibilidad). Shell-out a `git` solo en modo scm. Todo `CGO_ENABLED=0` (Constitución III).

**Alternatives considered**: librería git en Go (innecesaria); cliente S3 extra (ya tenemos minio-go).
