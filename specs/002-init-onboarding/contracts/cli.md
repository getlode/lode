# CLI Contract: `init` y `doctor` (feature 002)

**Feature**: 002-init-onboarding | **Phase**: 1

Nuevos comandos. Hereda los flags globales del feature 001 (`-v/--verbose`, `-q/--quiet`, `--cd`).

## `init`

Inicializa un repo de datos compatible con DVC en el directorio actual.

- **Flags**: `--no-scm` (inicializar sin git, para uso standalone).
- **Pre**: directorio sin `.dvc/` propio.
- **Comportamiento**:
  - Si el cwd está en un work tree de git y no se pasó `--no-scm` → modo `scm`: crea `.dvc/config` (vacío), `.dvc/.gitignore`, `.dvcignore`, `.dvc/tmp/btime`; hace `git add` de los tres archivos versionables.
  - Con `--no-scm` (o sin git y con `--no-scm`) → modo `no-scm`: crea `.dvc/config` (`[core] no_scm = True`), `.dvcignore`, `.dvc/tmp/btime`.
  - No crea `.dvc/cache` (lazy).
- **Salida**: confirma la inicialización y el modo; sugiere el siguiente paso (`lode add <target>`).
- **Errores**:
  - Sin git y sin `--no-scm` → error accionable: sugiere `lode init --no-scm` (standalone) o `git init` primero.
  - `.dvc/` ya existe en el cwd → informa "already initialized" (no destruye).
  - cwd dentro de un repo existente (`.dvc/` en ancestro) → informa la raíz del repo; no crea anidado.
- **Compat**: la estructura generada es byte-idéntica a la de `dvc init` / `dvc init --no-scm`.

## `doctor`

Diagnostica el estado del repo y del entorno.

- **Flags**: `--json` (salida estructurada); `-r/--remote NAME` (acotar el chequeo de remote).
- **Comportamiento**: corre los chequeos (repo, cache, format, remotes, coexistence) y reporta por cada uno estado + detalle + sugerencia.
- **Salida**: reporte legible (lista de chequeos con OK/warn/problem); con `--json`, el `DoctorReport` estructurado.
- **Exit code**: `0` si no hay ningún `problem`; `≠0` si hay al menos uno (apto para CI/scripts).
- **Reachability de remote**: conexión + autenticación + list mínimo, bajo timeout corto; no descarga objetos.

## Errores que guían (transversal, todos los comandos que requieren repo)

- Cuando no se encuentra repo: el error nombra `lode init` (y `--no-scm` para standalone) y menciona `--cd <dir>` si el usuario apuntó al directorio equivocado.
- `push`/`fetch`/`pull` sin remote configurado: el error muestra el comando para configurar uno (`lode remote add ...`).
- `checkout`/`pull` con objeto faltante: sugiere `lode pull` / verificar el remote.
