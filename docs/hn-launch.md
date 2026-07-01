# Hacker News launch plan

Use this when the repo is ready for a first HN post. Do not buy a domain before the
project has at least 10 stars and at least one external feedback signal.

## Suggested title

Show HN: lode - a faster DVC-compatible data versioning core in Go

## Short post

I built lode because DVC is great for dataset versioning, but `dvc add` and
`dvc status` can get painfully slow on repos with many small files. lode reimplements
the data-versioning hot path in Go while keeping DVC's on-disk format: same `.dvc`
files, `.dir` objects, cache layout, and S3-compatible remote layout.

The important part is that it is not a migration. You can point lode at an existing
DVC repo, run `lode add` / `lode status`, and keep using DVC on the same files. The
CI suite compares format-sensitive output against the real DVC binary and validates
S3-compatible interop against MinIO.

On Tiny-ImageNet (100,200 files), `lode add` is 12.4x faster than DVC 3.67.1, and
incremental add after changing one file is 13.2x faster. The gap is smaller on large
files, where both tools are mostly hash-bound; the benchmark doc includes those
numbers too.

It is early and intentionally scoped: data-versioning commands are supported, but
DVC pipelines / `dvc repro` are not. S3-compatible remotes work today; native GCS,
Azure, and SSH are planned.

I would especially value feedback from people with large DVC repos, many-small-file
datasets, or unusual S3-compatible remotes.

## Pre-post checklist

- README first screen says what it does, why it is safe, and how to try it.
- Compatibility matrix is linked from README.
- Try-without-risk guide is linked from README.
- At least 4 public issues exist for feedback/contribution paths.
- Latest release assets are downloadable.
- CI on `main` is green after final docs changes.
- Homebrew install path works: `brew install getlode/tap/lode`.

## Expected objections and replies

**Why not contribute this to DVC?**

The format stays DVC-compatible, so this is not trying to split the ecosystem. lode is
focused on the hot path where a single static Go binary and parallel hashing help.

**Is it safe for my data?**

The safety property is DVC interoperability, not trust in a new format. lode writes
standard DVC files and cache objects. If you stop using it, continue with DVC.

**Does this replace DVC?**

No. It accelerates the data-versioning core. Keep DVC for pipelines and commands lode
does not implement.

**Are the benchmarks cherry-picked?**

The benchmark doc includes many-small-file and large-file regimes, warm-cache
methodology, alternating order, memory, and cases where the speedup narrows.
