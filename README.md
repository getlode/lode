# lode

**DVC `add` and `status`, much faster. Same repo. No migration.**

[![CI](https://github.com/getlode/lode/actions/workflows/ci.yml/badge.svg)](https://github.com/getlode/lode/actions/workflows/ci.yml)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](LICENSE)
[![Go 1.25+](https://img.shields.io/badge/Go-1.25%2B-00ADD8.svg)](go.mod)

`lode` is a DVC-compatible data-versioning core in Go. Point it at an existing DVC
repo and it writes the same `.dvc` files, `.dir` objects, local cache layout, and
S3-compatible remote layout as DVC. If you stop using it, nothing is trapped in a new
format: uninstall `lode` and keep using `dvc` on the same files.

```console
$ time dvc add big/      # 20,000 files
real    0m5.79s

$ time lode add big/     # same repo, byte-identical DVC metadata
real    0m0.44s
```

On a Tiny-ImageNet benchmark with 100,200 files, `lode add` is **12.4x faster** and
incremental `add` after changing one file is **13.2x faster** than DVC 3.67.1. The
benchmark harness verifies that `dvc status` accepts the repo produced by `lode`.

## Try it safely on a DVC repo

```bash
brew install getlode/tap/lode
cd your-existing-dvc-repo
lode doctor
lode status
lode add path/to/data
lode verify
# Optional proof: DVC should still read the repo normally.
dvc status
```

Nothing migrates. `lode` only writes standard DVC files and cache objects. The safe
rollback is deleting the `lode` binary and continuing with DVC. For a more careful
copy-based trial, see [Try without risk](docs/try-without-risk.md).

## Current scope

| Works today | Not yet |
|---|---|
| `init`, `add`, `status`, `checkout`, `push`, `fetch`, `pull`, `gc`, `verify`, `doctor` | `dvc repro` / pipelines |
| Existing DVC 3.x repos and legacy DVC 2.x cache reads | Native `gs://`, Azure, SSH remotes |
| S3-compatible remotes: AWS S3, MinIO, Cloudflare R2, Backblaze B2 | Replacing every DVC command |

See the [compatibility matrix](docs/compatibility.md) for the tested surface and known limits.

## Help wanted

The most useful feedback right now is a real DVC repo trial: run `lode doctor`,
`lode status`, `lode add`, then `dvc status`, and report the result in
[#23](https://github.com/getlode/lode/issues/23). Good early contribution areas are
native [GCS](https://github.com/getlode/lode/issues/24),
[Azure](https://github.com/getlode/lode/issues/25), and
[SSH](https://github.com/getlode/lode/issues/26) remotes, plus
[benchmark reports](https://github.com/getlode/lode/issues/27) on real datasets.

## Why

DVC is the standard for versioning datasets and ML models, but its CLI struggles on
large repos: hashing is CPU-bound and throttled by the Python runtime. `lode`
reimplements the data-versioning core in Go — a dependency-free binary, concurrent
hashing, and a local state DB that skips re-hashing unchanged files. No migration:
your repo stays a DVC repo.

## Install

```bash
brew install getlode/tap/lode
# or download a binary from the Releases page (no toolchain needed)
# or, if you have Go:  go install github.com/getlode/lode/cmd/lode@latest
```

Single static binary, no runtime, no dependencies. Linux / macOS / Windows, amd64 / arm64.

> **New to DVC?** You don't need DVC or Python installed — lode is standalone. Run
> `lode init` then `lode add <folder>` to start versioning a dataset, and `lode push`
> to back it up to S3. Because lode uses DVC's on-disk format, DVC's
> [docs and concepts](https://dvc.org/doc) apply directly if you want to go deeper.
> For ML *pipelines* (`dvc repro`), keep using DVC — lode accelerates the data layer
> and coexists with it.

## Quickstart: zero to versioned

```console
$ mkdir cats-dataset && cd cats-dataset && lode init --no-scm
Initialized lode repository in .../cats-dataset/.dvc

$ lode add images/                 # hash + cache the folder, write a tiny pointer
images               tracked -> images.dvc

$ cat images.dvc                   # this text file is what you commit to git — not the data
outs:
- md5: da80a810597fa6de9381d9d1b76b3517.dir
  size: 600000
  nfiles: 3
  hash: md5
  path: images

$ lode status                      # instant — unchanged files are not re-hashed
Data and pipelines are up to date.

$ lode remote add -d r s3://my-bucket/store && lode push   # back the data up
```

`images.dvc` is a few lines of text you version in git; the actual files live in the
cache and your remote. Change one image and `lode status` flags it; `lode push` ships
only what changed. It is a standard DVC repo — `dvc` reads it too.

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
lode verify                    # check integrity + prove DVC compatibility on your repo
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
| `verify` | Re-hash cached objects and check they match their recorded hash — proves integrity and, on a DVC repo, that lode computes the same hashes |
| `doctor` | Diagnose repo, cache, remotes, format and DVC coexistence; `--json`, CI-friendly exit codes |

### S3 remote credentials

For S3-compatible remotes, credentials resolve in this order:

1. Explicit remote config in `.dvc/config`: `access_key_id`, `secret_access_key`, `session_token`
2. AWS environment variables
3. Shared AWS credentials file, honoring the remote `profile` option or `AWS_PROFILE`
4. IAM role credentials from the runtime environment

Prefer environment variables, the shared AWS credentials file, or IAM roles over
storing static keys in `.dvc/config`, because `.dvc/config` is plain text.

## DVC compatibility

- **Byte-identical** `.dvc` files and `.dir` objects to DVC 3.x — verified by a
  byte-oracle test that compares `lode`'s output against the real `dvc` binary.
- Same content-addressed layout in cache and remote (`files/md5/<2>/<rest>`).
- **Bidirectional interop**: objects `lode` pushes are pulled by DVC and vice versa
  (validated end-to-end against MinIO).
- Reads the legacy DVC 2.x cache layout.

This is validated, not aspirational: the test suite runs against a real DVC 3.67.1
install and a real S3-compatible server. The tested surface is tracked in
[docs/compatibility.md](docs/compatibility.md).

## Benchmarks

On a real public dataset — **Tiny-ImageNet, 100,200 files**, 16-core, DVC 3.67.1,
median of 6 runs (execution order alternated to remove page-cache bias):

| operation | DVC | lode | speedup |
|-----------|----:|-----:|--------:|
| `add` (cold) | 25.40s | 2.05s | **12.4×** |
| `status` (no change) | 3.46s | 1.16s | **3.0×** |
| `add` (1 file changed, of 100k) | 6.10s | 0.46s | **13.2×** |

…and `dvc status` then reports *"up to date"* on the repo `lode` produced — drop-in, no
migration (the harness asserts this every run). The last row is the structural win:
change one file in a 100k dataset and DVC re-processes the directory; lode's state DB
skips the rest. The gap is ~12× on many small files and **narrows to ~3.7× on large
files** (both become hash-bound — shown honestly). Full methodology (median±σ, memory,
file-size regimes) and the documented DVC slowness this addresses:
**[BENCHMARKS.md](BENCHMARKS.md)**. Reproduce with [`scripts/benchmark.sh`](scripts/benchmark.sh). See also [Try without risk](docs/try-without-risk.md).

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

## Project status

lode is young and currently maintained by one person. The honest reason that's
low-risk to depend on: **lode does not own your data or its format.** Your repo is a
standard DVC repo — if lode stalls or you walk away, uninstall it and keep using `dvc`,
no migration, no export. The byte-compatibility that makes that true is enforced by a
test that runs against the real `dvc` on every CI build. Roadmap and how to help:
[ROADMAP.md](ROADMAP.md), [CONTRIBUTING.md](CONTRIBUTING.md), [ARCHITECTURE.md](docs/ARCHITECTURE.md).

## License

**[MPL-2.0](LICENSE)** — free to use, modify, and ship, **including commercially and
inside closed-source products**. MPL is file-level copyleft: changes to lode's own files
stay open, but you can combine it with proprietary code, and there's no license to buy.
Optional commercial **support and services** are offered around the core — see
[COMMERCIAL.md](COMMERCIAL.md).

Contributions are accepted under a [DCO](https://developercertificate.org/) (sign off
with `git commit -s`) — no CLA, no copyright assignment. See [CONTRIBUTING.md](CONTRIBUTING.md).

> `lode` is an independent project and is not affiliated with or endorsed by
> Iterative, Inc. or the DVC project. "DVC" is used only to describe compatibility.
