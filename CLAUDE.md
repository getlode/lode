<!-- SPECKIT START -->
## Proyecto: lode (github.com/getlode/lode)

Acelerador en Go para el camino caliente de versionado de datos de DVC. Binario `lode`,
compatible con repos DVC 3.x para los comandos soportados. Go 1.25+ (`CGO_ENABLED=0`). Stack: cobra, **minio-go**
(remotes S3-compatible), bbolt (state), gofrs/flock (lock), x/sys/unix (reflink).

- **001-dvc-go** — ✅ implementado: add, status, checkout, push, pull, fetch, gc.
- **002-init-onboarding** — ✅ implementado: `lode init` (byte-compat con `dvc init`),
  `lode doctor`, errores que guían. Standalone sin Python.
- **Invariante (Constitución v1.0.0, Principio I)**: byte-compatibilidad con DVC
  (`.dvc`, objeto `.dir`, layout `files/md5/...`, estructura de `init`) prevalece
  sobre mejoras de diseño. Todo cambio de formato pasa por test-oráculo vs `dvc` real.

## Active Plan: 003-english-cli

Migrar toda la salida de la CLI al inglés (descripciones, progreso, errores) con
glosario canónico y plurales naturales. Solo texto, sin cambios de comportamiento.
Plan: `specs/003-english-cli/plan.md`. Verificación: test de barrido + re-habilitar
`misspell` en `.golangci.yml`. Los strings nuevos del 002 (init/doctor/errores) ya
están en inglés; faltan los de 001 (add/status/push/checkout/gc/remote/transfer/s3).
<!-- SPECKIT END -->
