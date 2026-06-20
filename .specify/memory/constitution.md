<!--
Sync Impact Report
==================
Version change: (template, unversioned) → 1.0.0
Bump rationale: Initial ratification of the project constitution (first concrete
version replacing the placeholder template). MAJOR baseline.

Modified principles:
  - [PRINCIPLE_1_NAME] → I. DVC Byte-Compatibility (NON-NEGOTIABLE)
  - [PRINCIPLE_2_NAME] → II. Oracle-Gated Format Changes
  - [PRINCIPLE_3_NAME] → III. Zero-CGO Single Binary
  - [PRINCIPLE_4_NAME] → IV. Performance Is the Product
  - [PRINCIPLE_5_NAME] → V. Coexistence Over Reinvention

Added sections:
  - Technology & Scope Constraints (Section 2)
  - Development Workflow & Quality Gates (Section 3)

Removed sections: none

Templates requiring updates:
  - ✅ .specify/templates/plan-template.md (Constitution Check gate made concrete)
  - ✅ .specify/templates/spec-template.md (no change needed; scope/SC already align)
  - ✅ .specify/templates/tasks-template.md (no change needed; oracle/perf tasks already modeled)

Follow-up TODOs: none
-->

# lode Constitution

## Core Principles

### I. DVC Byte-Compatibility (NON-NEGOTIABLE)

Every artifact lode writes that DVC also writes MUST be byte-identical to DVC 3.x:
the `.dvc` files, the `.dir` manifest objects, and the content-addressed layout of
the local cache and remotes (`files/md5/<2>/<rest>`). lode and DVC MUST be able to
operate the same repository interchangeably, in both directions.

When "improve the design" conflicts with "stay byte-compatible with DVC", compatibility
wins, unless the divergence is an explicit, versioned, opt-in feature documented as such.

Rationale: drop-in compatibility with the installed base is the entire reason the project
can gain adoption. A single divergent byte changes a hash and breaks interoperability.

### II. Oracle-Gated Format Changes

Any change that can affect a serialized artifact (hashing, `.dir` serialization, `.dvc`
emission, cache/remote key layout) MUST be covered by a byte-oracle test that compares
lode's output against the real `dvc` binary, and that test MUST be green before the
change is considered done. Command-level work MUST NOT proceed past a red oracle gate.

Rationale: the format is the contract (Principle I). The oracle is the only way to prove
compatibility objectively rather than by inspection.

### III. Zero-CGO Single Binary

The entire dependency chain MUST build with `CGO_ENABLED=0`. lode ships as a single
static binary and MUST cross-compile to linux/darwin/windows × amd64/arm64 without a C
toolchain. Dependencies that require cgo are not admissible; a pure-Go equivalent MUST be
chosen instead.

Rationale: frictionless single-binary distribution (Homebrew, GitHub Releases, `go install`)
is the distribution advantage over the Python original. cgo forfeits it.

### IV. Performance Is the Product

lode exists to be dramatically faster than DVC-Python on the hot path. `add`/`status` on
large datasets MUST stay at least ~10× faster than DVC-Python on comparable hardware.
Hot paths MUST parallelize CPU-bound work (bounded to NumCPU) and MUST NOT perform
per-file fsync or per-file database transactions; batch and stream instead. A measurable
regression on the hot path is a defect, not a tradeoff.

Rationale: "10× faster, drop-in" is the value proposition. Without the speed,
byte-compatibility alone gives users no reason to switch.

### V. Coexistence Over Reinvention

lode MUST honor the locking and on-disk conventions that let it run safely alongside
DVC-Python on the same repository (the global `.dvc/tmp/lock` via flock, the rwlock JSON).
New behavior MUST default to what DVC does; novel behavior MUST be additive and explicit,
never a silent change to shared state.

Rationale: users adopt incrementally and will run both tools during migration. Corrupting a
shared repo, or surprising DVC, destroys trust faster than any feature builds it.

## Technology & Scope Constraints

- Language: Go (currently 1.23+), `CGO_ENABLED=0` everywhere.
- Architecture: a CLI (`cmd/lode`) over a library of single-purpose packages under
  `internal/` (hashfile, cache, dvcfile, remote, transfer, checkout, lock, repo). Format-risk
  logic (`dvcfile`, `hashfile/tree`) MUST stay isolated and oracle-covered.
- Remotes: S3-compatible backends (AWS S3, MinIO, Cloudflare R2, Backblaze B2) via a pure-Go
  S3 client. Non-S3 backends (GCS, Azure, SSH) are out of current scope.
- Out of scope (until explicitly specced): the pipelines/`repro` engine and any feature that
  would require breaking Principle I.

## Development Workflow & Quality Gates

- Features follow the spec-kit flow: specify → plan → tasks → implement, with artifacts under
  `specs/<feature>/`.
- Before merging any change: `go vet ./...` clean and `go test ./...` green. `go test -short`
  MUST pass without external services (S3-less CI). Integration tests run against a real
  S3-compatible server (MinIO) and, for the oracle, the real `dvc` binary.
- Format-affecting work MUST land its oracle test first and keep it green (Principle II).
- Hot-path-affecting work SHOULD include or update a benchmark and MUST NOT regress
  performance (Principle IV).
- No silent truncation or scope reduction: if a change bounds coverage, it MUST say so.

## Governance

This constitution supersedes ad-hoc practice for the lode project. All plans and reviews
MUST verify compliance with the principles above; deviations MUST be justified in the plan's
Complexity Tracking (or the change rejected).

Amendments require: a written rationale, an update to this file with a Sync Impact Report,
a semantic-version bump, and propagation to any affected templates/docs. Versioning policy:
MAJOR for backward-incompatible governance/principle removals or redefinitions; MINOR for a
new principle/section or materially expanded guidance; PATCH for clarifications and wording.

Compliance is reviewed at each plan's Constitution Check gate and at code review. Runtime
development guidance for agents lives in `CLAUDE.md`.

**Version**: 1.0.0 | **Ratified**: 2026-06-20 | **Last Amended**: 2026-06-20
