# Data Model: Bootstrap standalone y onboarding

**Feature**: 002-init-onboarding | **Date**: 2026-06-20 | **Phase**: 1

Estructuras lógicas. Reutiliza entidades del feature 001 (Repo, RepoConfig).

## InitMode

| Valor | Config generada | `.dvc/.gitignore` | git add |
|---|---|---|---|
| `scm` (con git) | `.dvc/config` vacío | `/config.local\n/tmp\n/cache\n` | sí |
| `no-scm` | `[core]\n    no_scm = True\n` | (no se crea) | no |

Selección: `scm` si el cwd está en un work tree de git; si no, requiere flag `--no-scm`.

## Artefactos de init (byte-exactos)

| Archivo | Contenido |
|---|---|
| `.dvc/config` | vacío (scm) / `[core]\n    no_scm = True\n` (no-scm) |
| `.dvc/.gitignore` | `/config.local\n/tmp\n/cache\n` (solo scm) |
| `.dvcignore` (raíz) | template de 139 bytes (3 líneas comentario, `\n` final) |
| `.dvc/tmp/btime` | archivo vacío (marcador) |

`.dvc/cache` **no** se crea en init (lazy en el primer `add`).

## InitOutcome (estado al inicializar)

| Estado | Cuándo | Acción |
|---|---|---|
| `created` | no había `.dvc/` ni repo ancestro | crea estructura |
| `already-initialized` | el cwd ya tiene `.dvc/` | no toca; informa |
| `inside-existing-repo` | hay `.dvc/` en un ancestro | no crea anidado; informa la raíz |
| `needs-no-scm` | sin git y sin `--no-scm` | error guiado (sugiere `--no-scm` o `git init`) |

## DoctorCheck

| Campo | Tipo | Notas |
|---|---|---|
| `name` | string | `repo`, `cache`, `format`, `remote:<name>`, `coexistence` |
| `status` | enum | `ok` / `warn` / `problem` |
| `detail` | string | qué se encontró |
| `suggestion` | string | acción concreta (vacío si OK) |

## DoctorReport

| Campo | Tipo | Notas |
|---|---|---|
| `checks` | `[]DoctorCheck` | resultados en orden |
| `exitCode` | int | 0 si ningún `problem`; ≠0 si hay al menos un `problem` |

Chequeos y sus estados:
- **repo**: `ok` si hay `.dvc/` válido; `problem` si no (sugerir `lode init`).
- **cache**: `ok` si escribible; `problem` si no (permisos / FS).
- **format**: `ok` (3.x) / `warn` (legacy 2.x detectado, lectura soportada).
- **remote:<name>**: `ok` alcanzable; `warn` no configurado; `problem` configurado pero inalcanzable/credenciales inválidas.
- **coexistence**: `ok` lock libre; `warn` lock tomado (otro proceso DVC/lode activo).

## Relaciones

```
InitMode ──determina──> Artefactos de init
repo.Find (001) ──> InitOutcome (detección de repo existente/ancestro)
DoctorReport 1───* DoctorCheck
DoctorCheck (remote) ──usa──> remote.NewS3 + reachability (001)
DoctorCheck (cache/coexistence) ──usa──> cache / lock (001)
```
