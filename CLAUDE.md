<!-- SPECKIT START -->
## Active Plan: 001-dvc-go

Reimplementación en Go del núcleo de versionado de datos de DVC (binario único,
drop-in compatible con DVC 3.x). Plan: `specs/001-dvc-go/plan.md`.

- **Stack**: Go 1.23+ (`CGO_ENABLED=0`), cobra (CLI), aws-sdk-go-v2 +
  feature/s3/transfermanager (remotes S3-compatible), errgroup+crypto/md5+sync.Pool
  (hashing), bbolt (state DB), gofrs/flock (lock), x/sys/unix (reflink FICLONE),
  GoReleaser (distribución).
- **MVP**: add, status, checkout, push, pull, fetch, gc. Cache local + remotes
  S3-compatible (S3/MinIO/R2/B2). Pipelines/repro y backends no-S3 fuera del MVP.
- **Invariante**: compatibilidad byte-a-byte con DVC (`.dvc`, objeto `.dir`, layout
  `files/md5/...`) prevalece sobre mejoras de diseño. Riesgo #1: serialización exacta
  del `.dir` (separadores `", "`/`": "`, escape ASCII, sort por relpath) — cubierto
  por test-oráculo contra DVC real.
- Detalles: `specs/001-dvc-go/{research,data-model,quickstart}.md`, `contracts/cli.md`.
<!-- SPECKIT END -->
