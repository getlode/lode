# Architecture

lode is a CLI (`cmd/lode`) over a library of single-purpose packages under
`internal/`. Data flows: **command → repo/dvcfile/hashfile → cache → transfer ↔ remote**.

## Packages (`internal/`)

| Package | Responsibility |
|---|---|
| `cli` | Cobra commands; flag parsing; user-facing output and guided errors. Thin — delegates to the packages below. |
| `repo` | Repository discovery, `init` (byte-compatible with `dvc init`), `.dvc/config` (INI), well-known paths, git detection. |
| `dvcfile` | Read/write `.dvc` files **byte-exact** with DVC 3.x. |
| `hashfile` | MD5 hashing (parallel, bounded to NumCPU); the `.dir` manifest serialization (the trickiest compat detail — matches Python's `json.dumps`); the state DB (bbolt) that skips re-hashing unchanged files. |
| `cache` | Content-addressed object store (`files/md5/<2>/<rest>`), atomic writes, `0o444` protection, legacy 2.x read. |
| `remote` | S3-compatible backend (minio-go): object↔key mapping, presence, transfer, reachability. |
| `transfer` | push/fetch engines: remote status (HEAD vs LIST), `.dir`-after-contents ordering, integrity verification, idempotent resume. |
| `checkout` | Workspace materialization (reflink → hardlink/symlink → copy), relink detection, `.gitignore` management. |
| `lock` | DVC-compatible locking (global flock + rwlock JSON) so lode and DVC coexist. |

## The non-negotiable invariant

**Byte-compatibility with DVC.** Anything that changes a serialized artifact (`.dvc`,
`.dir`, cache/remote layout) must keep the oracle test (`tests/oracle/`, which runs the
real `dvc` and compares bytes) green. See [`.specify/memory/constitution.md`](../.specify/memory/constitution.md).

## Correctness of the state cache

The state DB `(inode, mtime, size) -> md5` is an **optimization, never a source of
truth**. It can produce a false "up to date" only if a file's content changes while
all three keys stay identical (e.g. NFS quirks, restored backups that reset mtimes,
recycled inodes). For those cases, use `--rehash` to force the always-correct full
rehash path. A corrupt or unreadable state DB also degrades to a full rehash.

## Where things live

- Commands: `internal/cli/<command>.go`
- Format-risk logic (most careful code): `internal/dvcfile`, `internal/hashfile/tree.go`
- Tests: `tests/oracle` (byte-compat vs real DVC), `tests/integration` (MinIO, interop, doctor, verify)
- Specs per feature: `specs/<NNN>-<name>/`

## Build & test

```bash
make build       # CGO_ENABLED=0 single binary
make test-short  # unit + oracle, no external services
make test        # full suite — needs MinIO and the real `dvc`
```
