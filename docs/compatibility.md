# Compatibility matrix

`lode` is a DVC-compatible accelerator for the data-versioning hot path. It is not a
full replacement for every DVC command.

## Tested DVC compatibility

| Surface | Status | Notes |
|---|---|---|
| `.dvc` pointer files | Supported for tested `dvc add` outputs | Byte-compatible with DVC 3.x output in oracle tests for file and directory adds. |
| `.dir` objects | Supported for tested directory adds | Serialization is compared against real DVC output. |
| Local cache layout | Supported | Uses `files/md5/<2>/<rest>` and reads legacy DVC 2.x cache layout. |
| DVC 3.x interop | Supported for tested outputs | CI runs the byte-oracle against real DVC 3.67.1. |
| DVC 2.x cache reads | Supported | Legacy cache layout is read for compatibility. |
| DVC pipelines / `dvc repro` | Not supported | Keep using DVC for pipeline orchestration. |

## Commands

| Command | Status | Compatibility expectation |
|---|---|---|
| `lode init` | Supported | Creates a DVC-compatible repo; `--no-scm` supported. |
| `lode add` | Supported | Writes DVC-compatible pointer files and cache objects. |
| `lode status` | Supported | Uses the state DB to avoid rehashing unchanged files; use `--rehash` for safety-critical checks or when metadata cannot be trusted. |
| `lode checkout` | Supported | Materializes workspace files from cache. |
| `lode push` / `fetch` / `pull` | Implemented for S3-compatible object storage | Object layout matches DVC remote layout; CI runs the MinIO-backed integration tests. |
| `lode gc` | Supported | Reclaims unreferenced cache objects; remote GC requires explicit flags. |
| `lode verify` | Supported | Rehashes cached objects and checks recorded hashes. |
| `lode doctor` | Supported | Diagnoses repo, cache, remotes, and DVC coexistence. |

## Remotes

| Remote | Status | Notes |
|---|---|---|
| MinIO | Tested in CI | Used by the live S3-compatible integration job. |
| AWS S3-style API | Implemented | Uses standard AWS credential resolution and S3 object APIs. |
| Cloudflare R2 | Expected via S3-compatible API | Provider-specific reports welcome. |
| Backblaze B2 | Expected via S3-compatible API | Provider-specific reports welcome. |
| DigitalOcean Spaces | Expected via S3-compatible API | Provider-specific reports welcome. |
| GCS via S3-compatible endpoint | Untested | May work with HMAC keys and `endpointurl`; reports welcome. |
| Native `gs://` | Planned | Not implemented yet. |
| Azure Blob | Planned | Not implemented yet. |
| SSH | Planned | Not implemented yet. |

## Safety model

`lode` does not introduce a new repository format. The safety goal is ordinary DVC
interoperability for the supported command surface: DVC should be able to read the
files and objects produced by `lode`. The CI oracle compares format-sensitive add
outputs against the real DVC binary, and CI runs MinIO integration tests for
S3-compatible push/pull interop.

The state DB is an optimization, not a source of truth. It uses file metadata to skip
rehashing unchanged files. Use `--rehash` for NFS, restored backups, safety-critical
checks, or any environment where file metadata may not reflect content changes.

For a cautious first run, follow [Try without risk](try-without-risk.md).
