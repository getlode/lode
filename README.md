# lode

**DVC `add` and `status`, much faster. Same repo. No migration.**

[![CI](https://github.com/getlode/lode/actions/workflows/ci.yml/badge.svg)](https://github.com/getlode/lode/actions/workflows/ci.yml)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](LICENSE)
[![Go 1.25+](https://img.shields.io/badge/Go-1.25%2B-00ADD8.svg)](go.mod)

`lode` is a DVC-compatible accelerator for the data-versioning hot path in Go.
Point it at an existing DVC repo and, for the supported commands, it writes the same
`.dvc` pointer files, `.dir` objects, local cache layout, and S3-compatible remote
object paths that DVC expects. If you stop using it, nothing is trapped in a new
format: keep using `dvc` on the same repository.

```console
$ time dvc add big/      # 20,000 files
real    0m5.79s

$ time lode add big/     # same repo, DVC-compatible add output
real    0m0.44s
```

On a Tiny-ImageNet benchmark with 100,200 files, `lode add` is **12.4x faster** and
incremental `add` after changing one file is **13.2x faster** than DVC 3.67.1. The
benchmark harness verifies that `dvc status` accepts the `dvc add`-style outputs
produced by `lode`.

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

Nothing migrates. `lode` writes standard DVC files and cache objects for the commands
it supports. There is no format lock-in; if you do not want the files produced during
a trial, review or revert the `.dvc`/`.gitignore` changes with Git and keep using DVC.
For a more careful copy-based trial, see [Try without risk](docs/try-without-risk.md).

## Current scope

| Works today | Not yet |
|---|---|
| `init`, `add`, `status`, `checkout`, `push`, `fetch`, `pull`, `gc`, `verify`, `doctor` | `dvc repro` / pipelines |
| Existing DVC 3.x repos and legacy DVC 2.x cache reads | Native `gs://`, Azure, SSH remotes |
| S3-compatible remote implementation; tested directly with MinIO | Replacing every DVC command |

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

DVC is the standard for versioning datasets and ML models, but `dvc add` and
`dvc status` can have high per-file overhead on large repos, especially many-small-file
datasets. `lode` reimplements that data-versioning hot path in Go, with concurrent
hashing, batched writes, and a local state DB that skips re-hashing unchanged files.
No migration: your repo stays a DVC repo.

## Install

```bash
brew install getlode/tap/lode
# or download a binary from the Releases page (no toolchain needed)
# or, if you have Go:  go install github.com/getlode/lode/cmd/lode@latest
```

Single self-contained binary, no Python runtime. Linux / macOS / Windows, amd64 / arm64.

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

Already have a DVC repo? Skip `init` and point `lode` at it for the supported commands — same DVC format, no migration.

| Command | What it does |
|---|---|
| `init` | Create a DVC-compatible repo — standalone, no Python required |
| `add` | Hash (in parallel), cache, write the `.dvc`, update `.gitignore` |
| `status` | Report changes using the state DB (no re-hash of unchanged data); `--json` |
| `push` / `fetch` / `pull` | Sync with S3-compatible object storage; MinIO is covered in integration tests |
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

- **Byte-compatible for tested `dvc add` file and directory outputs** — verified by a
  byte-oracle test that compares `lode`'s `.dvc` and `.dir` output against the real
  `dvc` binary.
- Same content-addressed layout in cache and remote (`files/md5/<2>/<rest>`).
- **Bidirectional interop for the integration suite**: objects `lode` pushes are pulled
  by DVC and vice versa in the MinIO-backed tests.
- Reads the legacy DVC 2.x cache layout.

This is validated where it matters most for format risk: CI runs the byte-oracle
against a real DVC 3.67.1 install and runs the MinIO interop suite against a live
S3-compatible server. The tested surface is tracked in
[docs/compatibility.md](docs/compatibility.md).

## Benchmarks

On a real public dataset — **Tiny-ImageNet, 100,200 files**, 16-core, DVC 3.67.1,
median of 6 runs (execution order alternated to remove page-cache bias):

| operation | DVC | lode | speedup |
|-----------|----:|-----:|--------:|
| `add` (fresh repo, warm page cache) | 25.40s | 2.05s | **12.4x** |
| `status` (no change) | 3.46s | 1.16s | **3.0x** |
| `add` (1 file changed, of 100k) | 6.10s | 0.46s | **13.2x** |

...and `dvc status` then reports *"up to date"* on the repo `lode` produced for the
tested `add` outputs. The harness asserts this every run. The last row is the
structural win:
change one file in a 100k dataset and DVC re-processes the directory; lode's state DB
skips the rest. The gap is ~12x on many small files and **narrows to ~3.7x on large
files** (both become hash-bound — shown honestly). Full methodology (median +/- stddev, memory,
file-size regimes) and the documented DVC slowness this addresses:
**[BENCHMARKS.md](BENCHMARKS.md)**. Reproduce with [`scripts/benchmark.sh`](scripts/benchmark.sh). See also [Try without risk](docs/try-without-risk.md).

## How it works

- **Parallel, bounded hashing** (errgroup capped at `NumCPU`) with a reused buffer pool.
- **State DB** (embedded, pure-Go) keyed by file metadata -> hash, so `status` and
  re-`add` skip files that did not change. Use `--rehash` for safety-critical checks,
  NFS/restored-backup edge cases, or whenever metadata cannot be trusted.
- **Atomic, batched writes**: no per-file fsync, no per-file DB transaction — the two
  changes that turned an early 78 s run into 0.44 s.
- **Zero cgo** end-to-end, so the binary cross-compiles to every target without a C toolchain.

## Scope

In scope today: the data-versioning core — `add`, `status`, `checkout`, `push`,
`fetch`, `pull`, `gc` — over a local cache and S3-compatible object storage. MinIO is
covered by CI integration tests; provider-specific S3-compatible services should work
through the same API but need more field reports.

Not yet (planned / feedback-driven): the pipelines / `repro` engine, and non-S3
remotes (GCS, Azure, SSH).

## Development

```bash
make build         # CGO_ENABLED=0 single binary
make test-short    # unit + oracle, no external services
make test          # full suite — needs MinIO and the real `dvc` binary
make lint
```

CI runs unit tests, the DVC byte-oracle, and the MinIO-backed integration suite. To
run the integration suite locally, provide a MinIO (`MINIO_ENDPOINT`,
`MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`), the reference `dvc` (`DVC_BIN`), and the
built binary (`LODE_BIN`).
See `tests/` for details. Contributions: see [CONTRIBUTING.md](CONTRIBUTING.md).

## Project status

lode is young and currently maintained by one person. The honest reason that's
low-risk to depend on: **lode does not own your data or its format.** Your repo is a
standard DVC repo — if lode stalls or you walk away, keep using `dvc`, no migration,
no export. The tested byte-compatibility that makes that true is enforced by a
test that runs against the real `dvc` on every CI build. Roadmap and how to help:
[ROADMAP.md](ROADMAP.md), [CONTRIBUTING.md](CONTRIBUTING.md), [ARCHITECTURE.md](docs/ARCHITECTURE.md).

## License

**[MPL-2.0](LICENSE)** — free to use, modify, and ship, **including commercially and
inside closed-source products**. MPL is file-level copyleft: changes to lode's own files
stay open, but you can combine it with proprietary code, and there's no license to buy.
Commercial use is allowed under MPL-2.0; see [COMMERCIAL.md](COMMERCIAL.md) for the plain-English licensing note.

Contributions are accepted under a [DCO](https://developercertificate.org/) (sign off
with `git commit -s`) — no CLA, no copyright assignment. See [CONTRIBUTING.md](CONTRIBUTING.md).

> `lode` is an independent project and is not affiliated with or endorsed by
> Iterative, Inc. or the DVC project. "DVC" is used only to describe compatibility.
