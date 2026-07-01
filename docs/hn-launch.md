# Hacker News launch plan

Use this when the repo is ready for a first HN post. Do not buy a domain before the
project has at least 10 stars and at least one external feedback signal.

## Suggested title

Show HN: lode - faster DVC add/status, same DVC repo

## Short post

I made lode after running into the part of DVC that hurts most for me: large dataset
repos with lots of small files, where `dvc add` and `dvc status` spend a long time
walking and hashing.

The idea is deliberately narrow. lode is not a new data-versioning format and it is
not trying to replace all of DVC. It reimplements the data-versioning hot path in Go
while writing the same `.dvc` files, `.dir` objects, cache layout, and S3-compatible
remote layout that DVC expects.

So the intended usage is: point it at an existing DVC repo, run `lode add` or
`lode status` for the fast path, and keep using DVC for everything else. If you stop
using lode, there is no export step. Just keep using DVC on the same repo.

On Tiny-ImageNet (100,200 files), `lode add` is 12.4x faster than DVC 3.67.1, and an
incremental add after changing one file is 13.2x faster. The benchmark doc also shows
where the advantage shrinks: on large files both tools become mostly hash-bound.

The project is still early. The supported scope today is the core data commands
(`init`, `add`, `status`, `checkout`, `push`, `fetch`, `pull`, `gc`, `verify`,
`doctor`) with S3-compatible remotes. Pipelines / `dvc repro` are not implemented;
for those, keep using DVC. Native GCS, Azure, and SSH remotes are still planned.

I would really value feedback from people with real DVC repos, especially if you have
many-small-file datasets or S3-compatible remotes outside plain AWS S3. I am most
interested in compatibility reports right now, including cases where it does not work.

## Shorter variant

I built lode because DVC is useful, but `dvc add` / `dvc status` can be painfully slow
on repos with lots of small files.

lode is a small Go implementation of the DVC data-versioning hot path. It writes the
same `.dvc` files, `.dir` objects, cache layout, and S3-compatible remote layout as
DVC, so it is not a migration to a new format. Use lode for the fast path, keep using
DVC for everything else, and stop using lode anytime without exporting anything.

On Tiny-ImageNet (100,200 files), `lode add` is 12.4x faster than DVC 3.67.1. The
project is early and scoped: core data commands and S3-compatible remotes work today;
DVC pipelines, native GCS, Azure, and SSH do not yet.

I would love feedback from people with real DVC repos, especially compatibility reports
and benchmark results on large datasets.

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
