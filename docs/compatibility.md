# Compatibility matrix

`lode` is a DVC-compatible accelerator for the data-versioning hot path. It is not a
full replacement for every DVC command.

## Tested DVC compatibility

| Surface | Status | Notes |
|---|---|---|
| `.dvc` pointer files | Supported | Byte-compatible with DVC 3.x output in oracle tests. |
| `.dir` objects | Supported | Serialization is compared against real DVC output. |
| Local cache layout | Supported | Uses `files/md5/<2>/<rest>` and reads legacy DVC 2.x cache layout. |
| DVC 3.x interop | Supported | CI runs against real DVC 3.67.1. |
| DVC 2.x cache reads | Supported | Legacy cache layout is read for compatibility. |
| DVC pipelines / `dvc repro` | Not supported | Keep using DVC for pipeline orchestration. |

## Commands

| Command | Status | Compatibility expectation |
|---|---|---|
| `lode init` | Supported | Creates a DVC-compatible repo; `--no-scm` supported. |
| `lode add` | Supported | Writes DVC-compatible pointer files and cache objects. |
| `lode status` | Supported | Uses the state DB to avoid rehashing unchanged files; `--rehash` forces the safe path. |
| `lode checkout` | Supported | Materializes workspace files from cache. |
| `lode push` / `fetch` / `pull` | Supported for S3-compatible remotes | Object layout matches DVC remote layout. |
| `lode gc` | Supported | Reclaims unreferenced cache objects; remote GC requires explicit flags. |
| `lode verify` | Supported | Rehashes cached objects and checks recorded hashes. |
| `lode doctor` | Supported | Diagnoses repo, cache, remotes, and DVC coexistence. |

## Remotes

| Remote | Status | Notes |
|---|---|---|
| AWS S3 | Supported | Uses standard AWS credential resolution. |
| MinIO | Supported | Used in integration tests. |
| Cloudflare R2 | Supported | S3-compatible endpoint. |
| Backblaze B2 | Supported | S3-compatible endpoint. |
| DigitalOcean Spaces | Supported | S3-compatible endpoint. |
| GCS via S3-compatible endpoint | Untested | May work with HMAC keys and `endpointurl`; reports welcome. |
| Native `gs://` | Planned | Not implemented yet. |
| Azure Blob | Planned | Not implemented yet. |
| SSH | Planned | Not implemented yet. |

## Safety model

`lode` does not introduce a new repository format. The safety guarantee is ordinary
DVC interoperability: DVC should be able to read the files and objects produced by
`lode`. The CI oracle compares format-sensitive output against the real DVC binary,
and integration tests validate S3-compatible push/pull interop.

For a cautious first run, follow [Try without risk](try-without-risk.md).
