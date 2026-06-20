# Benchmarks

`lode` vs DVC on the **local hot path** (`add` / `status`) — where a data-versioning
CLI is used every day and where DVC's Python runtime hurts. Reproduce everything
yourself with [`scripts/benchmark.sh`](scripts/benchmark.sh).

> **Honest scope.** These numbers cover `add`/`status` on **many small files** — the
> common ML-data case and the one DVC users complain about most (see *Why this matters*
> below). `push`/`pull` are network-bound and roughly **on par** with DVC. Large
> single-file throughput is hash-bound and the gap narrows. Numbers are single cold
> runs on one machine — indicative, not a statistical study. Run the script and see.

## Real dataset: Tiny-ImageNet (100,200 files)

A real public dataset (`wget http://cs231n.stanford.edu/tiny-imagenet-200.zip`),
`train/` split — 100,200 small JPEGs. 16-core machine, DVC 3.67.1, defaults on both.

| operation | DVC (Python) | lode (Go) | speedup |
|-----------|-------------:|----------:|--------:|
| `add` (cold, hashes everything) | 25.77s | 2.25s | **11.5×** |
| `status` (no change) | 3.48s | 1.31s | **2.7×** |
| `add` (1 file changed, of 100k) | 6.23s | 0.47s | **13.3×** |

And — the whole point — **DVC reads what lode produced**: after `lode init && lode add train`,
`dvc status` reports *"Data and pipelines are up to date."* Same `.dvc`, same `.dir`
objects, same cache layout. Drop-in, no migration.

The **`add` after changing one file** row is the structural win: DVC re-processes the
directory (6.23s) even though 99,999 files are untouched; lode's state DB skips them
(0.47s). That is exactly the pain reported in DVC issues (below) — and it is not a
constant-factor speedup, it is *not doing the work*.

## Synthetic scaling (deterministic, runs anywhere)

`scripts/benchmark.sh` with `SCALES="1000 10000 50000"` — deterministic generated
datasets, so anyone gets comparable results.

16-core machine, DVC 3.67.1, defaults on both. Single cold run per cell.

| files | operation | DVC | lode | speedup |
|------:|-----------|----:|-----:|--------:|
| 1,000 | `add` (cold) | 0.92s | 0.06s | **15.3×** |
| 1,000 | `status` (no change) | 0.43s | 0.03s | **14.3×** |
| 1,000 | `add` (1 file changed) | 0.63s | 0.02s | **31.5×** |
| 10,000 | `add` (cold) | 3.11s | 0.29s | **10.7×** |
| 10,000 | `status` (no change) | 0.70s | 0.17s | **4.1×** |
| 10,000 | `add` (1 file changed) | 1.16s | 0.06s | **19.3×** |
| 50,000 | `add` (cold) | 14.94s | 1.25s | **12.0×** |
| 50,000 | `status` (no change) | 2.20s | 0.66s | **3.3×** |
| 50,000 | `add` (1 file changed) | 3.96s | 0.28s | **14.1×** |

Consistent with the real dataset: `add` stays ~10–15× across scales, the incremental
case 14–31×, and `status` 3–14× (both tools cache state, so the gap there is smaller
but real).

## How lode is faster

- **Parallel hashing** bounded to `NumCPU` (DVC is single-process Python, GIL-bound).
- **State DB** keyed by `(inode, mtime, size)` → skip re-hashing unchanged files.
- **No per-file fsync, no per-file DB transaction** — batched writes (the two fixes
  that turned an early 78s run into sub-second).
- **Single static Go binary**, no Python runtime overhead on startup or hashing.

## Why this matters (the pain is documented)

DVC's slowness on large/many-file datasets is a long-standing, widely-reported issue:

- [dvc#7607](https://github.com/iterative/dvc/issues/7607) — *"600k files… takes many hours with `dvc add`"* (most-upvoted perf issue).
- [dvc#7681](https://github.com/iterative/dvc/issues/7681) — *"3M files… `dvc status` takes 20+ minutes to calculate hashes."*
- [dvc#6977](https://github.com/iterative/dvc/issues/6977) — *"`dvc add` takes ~1 hour to re-compute md5… the large files are untouched."* (the incremental case lode eliminates)
- On Hacker News, the Oxen.ai maintainer: *"we built Oxen because DVC was painfully slow"*; another dev: *"we've written our own (10×) faster version of dvc."*

`lode` keeps your existing DVC repo and removes that pain on the daily path.

## Reproduce

```bash
make build
# synthetic scaling
LODE=./lode DVC_BIN=$(which dvc) SCALES="1000 10000 50000" scripts/benchmark.sh
# a real dataset directory
LODE=./lode DVC_BIN=$(which dvc) scripts/benchmark.sh /path/to/dataset
```
