# lode

**A fast, drop-in compatible reimplementation of [DVC](https://dvc.org)'s data-versioning core, in Go.**

[![CI](https://github.com/getlode/lode/actions/workflows/ci.yml/badge.svg)](https://github.com/getlode/lode/actions/workflows/ci.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPLv3-blue.svg)](LICENSE)
[![Go 1.23+](https://img.shields.io/badge/Go-1.23%2B-00ADD8.svg)](go.mod)

Point `lode` at your existing DVC repository and get the same format — identical
`.dvc` files, the same `.dir` objects, the same cache and remote layout — but as a
single static binary with parallel hashing. On large datasets it is **~10× faster**
than DVC-Python, and it coexists with DVC on the same repo: run either tool, in
either order.

```console
$ time dvc add big/      # 20,000 files
real    0m5.79s

$ time lode add big/     # same repo, same result, byte-identical metadata
real    0m0.44s
```

## Why

DVC is the standard for versioning datasets and ML models, but its CLI struggles on
large repos: hashing is CPU-bound and throttled by the Python runtime. `lode`
reimplements the data-versioning core in Go — a dependency-free binary, concurrent
hashing, and a local state DB that skips re-hashing unchanged files. No migration:
your repo stays a DVC repo.

## Install

```bash
go install github.com/getlode/lode/cmd/lode@latest
# or grab a binary from Releases, or: brew install getlode/tap/lode
```

Single static binary, no runtime, no dependencies. Linux / macOS / Windows, amd64 / arm64.

## Usage

```bash
lode init --no-scm             # start a repo with no Python and no DVC (use plain `lode init` in a git repo)
lode add data/                 # track a directory (or a file)
lode status                    # what changed — without re-hashing unchanged data
lode remote add -d r s3://bucket/store
lode remote modify r endpointurl https://nyc3.digitaloceanspaces.com
lode push                      # upload to an S3-compatible remote
lode pull                      # fetch + checkout on a clean clone
lode checkout                  # materialize the workspace from cache
lode gc -f                     # reclaim unreferenced objects
lode doctor                    # diagnose repo, cache, remotes and DVC coexistence
```

Already have a DVC repo? Skip `init` and point `lode` at it — same format, both tools interoperate.

| Command | What it does |
|---|---|
| `init` | Create a repo byte-compatible with `dvc init` — standalone, no Python required |
| `add` | Hash (in parallel), cache, write the `.dvc`, update `.gitignore` |
| `status` | Report changes using the state DB (no re-hash of unchanged data); `--json` |
| `push` / `fetch` / `pull` | Sync with an S3-compatible remote (AWS S3, MinIO, Cloudflare R2, Backblaze B2) |
| `checkout` | Materialize the workspace (reflink → hardlink/symlink → copy) |
| `gc` | Remove unreferenced objects from the cache (and remote with `-c`) |
| `doctor` | Diagnose repo, cache, remotes, format and DVC coexistence; `--json`, CI-friendly exit codes |

## DVC compatibility

- **Byte-identical** `.dvc` files and `.dir` objects to DVC 3.x — verified by a
  byte-oracle test that compares `lode`'s output against the real `dvc` binary.
- Same content-addressed layout in cache and remote (`files/md5/<2>/<rest>`).
- **Bidirectional interop**: objects `lode` pushes are pulled by DVC and vice versa
  (validated end-to-end against MinIO).
- Reads the legacy DVC 2.x cache layout.

This is validated, not aspirational: the test suite runs against a real DVC 3.67.1
install and a real S3-compatible server.

## How it works

- **Parallel, bounded hashing** (errgroup capped at `NumCPU`) with a reused buffer pool.
- **State DB** (embedded, pure-Go) keyed by `(inode, mtime, size)` → hash, so `status`
  and re-`add` skip files that didn't change.
- **Atomic, batched writes**: no per-file fsync, no per-file DB transaction — the two
  changes that turned an early 78 s run into 0.44 s.
- **Zero cgo** end-to-end, so the binary cross-compiles to every target without a C toolchain.

## Scope

In scope today: the data-versioning core — `add`, `status`, `checkout`, `push`,
`fetch`, `pull`, `gc` — over a local cache and S3-compatible remotes.

Not yet (planned / feedback-driven): the pipelines / `repro` engine, and non-S3
remotes (GCS, Azure, SSH).

## Development

```bash
make build         # CGO_ENABLED=0 single binary
make test-short    # unit + oracle, no external services
make test          # full suite — needs MinIO and the real `dvc` binary
make lint
```

The full integration/oracle suite expects a MinIO (`MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`,
`MINIO_SECRET_KEY`), the reference `dvc` (`DVC_BIN`), and the built binary (`LODE_BIN`).
See `tests/` for details. Contributions: see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Dual-licensed: **[AGPL-3.0](LICENSE)** for open-source use, plus a **commercial
license** for use without the AGPL's obligations — see [COMMERCIAL.md](COMMERCIAL.md).
Contributions require agreeing to the CLA ([CONTRIBUTING.md](CONTRIBUTING.md)).

> `lode` is an independent project and is not affiliated with or endorsed by
> Iterative, Inc. or the DVC project. "DVC" is used only to describe compatibility.
