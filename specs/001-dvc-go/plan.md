# Implementation Plan: Versionado de datos de alta velocidad, drop-in compatible con DVC

**Branch**: `001-dvc-go` | **Date**: 2026-06-19 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-dvc-go/spec.md`

## Summary

Reimplementar en Go el núcleo de versionado de datos de DVC (Data Version Control) como binario único, drop-in compatible con repos DVC 3.x existentes. El requisito primario es **interoperabilidad byte-a-byte** (mismos `.dvc`, mismo objeto `.dir`, mismo layout de cache/remote) entregando ≥10× de velocidad en el camino caliente (hashing paralelo + state DB para no rehashear). MVP: `add`, `status`, `checkout`, `push`, `pull`, `fetch`, `gc`, sobre cache local + remotes S3-compatible. El enfoque técnico está validado contra el código fuente real de DVC (ver [research.md](research.md)); el mayor riesgo es la serialización exacta del `.dir`, mitigado con un test-oráculo byte-a-byte como gate.

## Technical Context

**Language/Version**: Go 1.23+ (`CGO_ENABLED=0` en toda la cadena para cross-compile sin toolchain de C)

**Primary Dependencies**: `spf13/cobra` (CLI); `aws-sdk-go-v2/service/s3` + `feature/s3/transfermanager` (remotes S3-compatible vía `BaseEndpoint`+`UsePathStyle`); `golang.org/x/sync/errgroup` + `crypto/md5` + `sync.Pool` (hashing paralelo); `go.etcd.io/bbolt` (state DB puro Go); `gofrs/flock` (lock cross-platform); `golang.org/x/sys/unix` (reflink FICLONE)

**Storage**: Filesystem (cache content-addressed `.dvc/cache/files/md5/...`, modo 0o444), bbolt para state local (`.dvc/tmp/dvcgo/state.db`), remotes S3-compatible (AWS S3, MinIO, Cloudflare R2, Backblaze B2)

**Testing**: `testing` estándar (table-driven, incluido el oráculo de bytes contra DVC real); `testcontainers-go/modules/minio` para integración de remote (gateado con `testing.Short()`)

**Target Platform**: Linux, macOS, Windows (amd64 + arm64); binario único distribuido vía GoReleaser + Homebrew + GitHub Releases

**Project Type**: CLI + librería Go embebible (single project)

**Performance Goals**: `add`/`status` ≥10× más rápido que DVC-Python en datasets de ≥100k archivos (SC-001); `status` sin cambios sin rehashear, tiempo proporcional al nº de entradas (SC-005); transferencias remotas concurrentes con throughput agregado > secuencial (SC-006)

**Constraints**: Compatibilidad byte-a-byte con DVC 3.x es invariante prevalente sobre mejoras de diseño durante el MVP; integridad verificada al 100% en pull/checkout; sin estado corrupto ante interrupciones (escritura atómica tmp+rename); streaming para archivos arbitrariamente grandes; coexistencia con DVC-Python sobre el mismo repo (honrar `.dvc/tmp/lock`)

**Scale/Scope**: ~7 comandos del MVP; datasets de cientos de miles de archivos / archivos individuales de cientos de GB; pipelines/repro y backends no-S3 fuera del MVP

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

La constitución del proyecto (`.specify/memory/constitution.md`) está sin ratificar (plantilla con placeholders, sin principios definidos). **No hay gates de constitución que evaluar.** Se aplican como guía las buenas prácticas de ingeniería del proyecto (CLAUDE.md): código limpio, SOLID/DRY/KISS, sin abstracciones prematuras, validar solo en bordes del sistema. No se detectan violaciones que requieran justificación en Complexity Tracking. **Re-evaluación post-Phase 1**: el diseño respeta estos principios (paquetes de responsabilidad única, sin abstracción prematura); sigue sin haber violaciones.

## Project Structure

### Documentation (this feature)

```text
specs/001-dvc-go/
├── plan.md              # Este archivo
├── research.md          # Phase 0 — decisiones técnicas verificadas vs código DVC
├── data-model.md        # Phase 1 — entidades e invariantes de compatibilidad
├── quickstart.md        # Phase 1 — escenarios de validación end-to-end
├── contracts/
│   └── cli.md           # Phase 1 — contrato de comandos del MVP
└── tasks.md             # Phase 2 (/speckit-tasks — NO creado por /speckit-plan)
```

### Source Code (repository root)

```text
cmd/
└── dvcgo/
    └── main.go              # Entrypoint; wiring de cobra

internal/
├── cli/                     # Comandos cobra: add, status, checkout, push, pull, fetch, gc
├── repo/                    # Descubrimiento del repo, RepoConfig (.dvc/config INI)
├── dvcfile/                 # Lectura/escritura byte-a-byte de archivos .dvc (YAML 3.x)
├── hashfile/
│   ├── hash.go              # MD5 streaming, worker pool (errgroup+sync.Pool)
│   ├── tree.go              # DirManifest: serialización exacta del .dir (oráculo)
│   └── state.go             # State DB bbolt: (ino,mtime,size)→md5
├── cache/                   # Cache local content-addressed (files/md5), protect 0o444, atomic write
├── remote/
│   ├── s3.go                # aws-sdk-go-v2 + transfermanager; BaseEndpoint/UsePathStyle
│   └── status.go            # oids_exist: HEAD-por-objeto vs LIST masivo (heurística)
├── transfer/                # push/fetch: orden .dir-tras-contenidos, concurrencia, reanudación
├── checkout/                # Materialización: reflink/hardlink/symlink/copy, relink, .gitignore
└── lock/                    # GlobalLock (flock) + rwlock JSON compatibles con DVC

tests/
├── oracle/                  # Fixtures generados por DVC real + comparación de bytes
├── integration/             # MinIO vía testcontainers (push/pull/gc round-trip)
└── unit/                    # Co-ubicados *_test.go por paquete

.goreleaser.yaml             # Build matrix CGO_ENABLED=0 + Homebrew tap
```

**Structure Decision**: Single project (CLI + librería). La librería vive en `internal/` por paquete de responsabilidad única (hashfile, cache, remote, transfer, checkout, lock, dvcfile, repo); `cmd/dvcgo` solo cablea cobra. Esta separación permite testear cada unidad de forma aislada y reutilizar la lógica como librería embebible. Los paquetes con mayor riesgo de compatibilidad (`dvcfile`, `hashfile/tree`) quedan aislados y cubiertos por el oráculo de bytes. (Si se decide exponer la librería públicamente para embeber en apps de terceros, los paquetes estables se promueven de `internal/` a un `pkg/` público en una fase posterior.)

## Complexity Tracking

> No aplica — Constitution Check sin violaciones.
