<!-- SPECKIT START -->
## Proyecto: lode (github.com/getlode/lode)

Reescritura en Go del núcleo de versionado de DVC. Binario `lode`, drop-in
compatible con DVC 3.x. Go 1.23+ (`CGO_ENABLED=0`). Stack: cobra, **minio-go**
(remotes S3-compatible), bbolt (state), gofrs/flock (lock), x/sys/unix (reflink).

- **001-dvc-go** — ✅ implementado: add, status, checkout, push, pull, fetch, gc.
- **Invariante (Constitución v1.0.0, Principio I)**: byte-compatibilidad con DVC
  (`.dvc`, objeto `.dir`, layout `files/md5/...`, estructura de `init`) prevalece
  sobre mejoras de diseño. Todo cambio de formato pasa por test-oráculo vs `dvc` real.

## Active Plan: 002-init-onboarding

`lode init` (byte-compat con `dvc init`, modos scm/no-scm) + `lode doctor` +
errores que guían → uso standalone sin Python. Plan: `specs/002-init-onboarding/plan.md`.
Reutiliza la infra del 001; sin deps nuevas; shell-out a `git` solo en modo scm.
Detalles: `specs/002-init-onboarding/{research,data-model,quickstart}.md`, `contracts/cli.md`.
<!-- SPECKIT END -->
