# Benchmarks

`lode` vs DVC on the **local hot path** (`add` / `status`) — where a data-versioning
CLI is used every day and where DVC's Python runtime hurts. Reproduce everything
yourself with [`scripts/benchmark.sh`](scripts/benchmark.sh).

> **Honest scope.** These numbers cover `add`/`status`. `push`/`pull` are
> network-bound and roughly **on par** with DVC. As files get larger the work
> becomes hash-bound and the gap narrows (shown below). Numbers are **warm-cache**
> (see Methodology).

## Methodology

Designed to survive scrutiny, not to flatter:

- **N runs per cell** (default 5); we report the **median ± standard deviation** and
  the speedup of the medians. Cells whose median sits in the timing noise floor are
  flagged `(noise-floor)` rather than presented as a speedup.
- **Execution order is alternated** every run (dvc-first / lode-first), so neither
  tool systematically benefits from a page cache the other warmed. True cold
  (`drop_caches`) needs root and is **not** assumed — results are **warm** and labeled
  as such; the alternation removes the order bias.
- **Two file-size regimes**: many-small files and few-large files (the CPU-bound
  hashing case), so you can see where the advantage holds and where it narrows.
- **Peak memory (RSS)** is reported per operation (GNU `/usr/bin/time`).
- The **interop claim is encoded**: after `lode add`, the harness runs `dvc status`
  and asserts the repo is *up to date* — i.e. DVC reads what lode produced. Every run
  below reports `interop(lode->dvc): PASS`.

Environment for the numbers below: 16-core x86_64, Linux, DVC 3.67.1, RUNS=5, warm
cache with order alternated.

## Synthetic (deterministic, runs anywhere)

`scripts/benchmark.sh` with `REGIMES="50000:1024 16:67108864"`:

| files × size | operation | DVC median±sd | lode median±sd | speedup | lode RSS |
|---|---|--:|--:|--:|--:|
| 50,000 × 1KB | `add` | 12.39s ±0.06 | 1.01s ±0.02 | **12.3×** | 104 MB |
| 50,000 × 1KB | `status` | 1.91s ±0.01 | 0.59s ±0.01 | **3.2×** | 112 MB |
| 50,000 × 1KB | `add` (1 changed) | 3.38s ±0.02 | 0.23s ±0.01 | **14.7×** | 57 MB |
| 16 × 64MB (1 GB) | `add` | 1.17s ±0.12 | 0.23s ±0.00 | **5.1×** | 11 MB |
| 16 × 64MB (1 GB) | `status` | 0.34s ±0.08 | 0.14s ±0.01 | **2.4×** | 11 MB |
| 16 × 64MB (1 GB) | `add` (1 changed) | 0.69s ±0.08 | 0.10s ±0.03 | **6.9×** | 9 MB |

The honest story across regimes: **~12× on many small files** (DVC's per-file Python
overhead dominates), narrowing to **~5× on large files** (both tools become hash-bound,
so the gap shrinks — exactly as expected). `status` is a steadier ~2.4–3.2× because both
tools cache state. The `add (1 changed)` row is the structural win: change one file and
DVC reprocesses the directory; lode's state DB skips the rest.

## Real dataset: Tiny-ImageNet (100,200 files)

A real public dataset (`wget http://cs231n.stanford.edu/tiny-imagenet-200.zip`),
`train/` split, same rigorous method:

| operation | DVC median±sd | lode median±sd | speedup | lode RSS |
|---|--:|--:|--:|--:|
| `add` (cold) | 26.08s ±0.32 | 2.24s ±0.09 | **11.6×** | 153 MB |
| `status` (no change) | 3.58s ±0.04 | 1.22s ±0.02 | **2.9×** | 152 MB |
| `add` (1 file changed) | 6.31s ±0.10 | 0.48s ±0.03 | **13.1×** | 121 MB |

The real-dataset numbers match the synthetic many-small regime and barely move under
the rigorous method (median of 5, order-alternated) — i.e. the speedup is real, not a
page-cache artifact. At 100k files lode's resident memory is ~150 MB (it builds the file
list in memory); that scales with file count and is the honest cost of the parallel walk.

## How lode is faster

- **Parallel hashing** bounded to `NumCPU` (DVC is single-process Python, GIL-bound).
- **State DB** keyed by `(inode, mtime, size)` → skip re-hashing unchanged files.
- **No per-file fsync, no per-file DB transaction** — batched writes.
- **Single static Go binary**, no Python runtime overhead on startup or hashing.

## Why this matters (the pain is documented)

- [dvc#7607](https://github.com/iterative/dvc/issues/7607) — *"600k files… takes many hours with `dvc add`"* (most-upvoted perf issue).
- [dvc#7681](https://github.com/iterative/dvc/issues/7681) — *"3M files… `dvc status` takes 20+ minutes to calculate hashes."*
- [dvc#6977](https://github.com/iterative/dvc/issues/6977) — *"`dvc add` takes ~1 hour to re-compute md5… the large files are untouched."* (the incremental case lode eliminates)
- On Hacker News, the Oxen.ai maintainer: *"we built Oxen because DVC was painfully slow"*; another dev: *"we've written our own (10×) faster version of dvc."*

## Reproduce

```bash
make build
# synthetic regimes (count:bytes), N runs each
LODE=./lode DVC_BIN=$(which dvc) RUNS=5 REGIMES="50000:1024 16:67108864" scripts/benchmark.sh
# a real dataset directory
LODE=./lode DVC_BIN=$(which dvc) RUNS=5 scripts/benchmark.sh /path/to/dataset-dir
```
