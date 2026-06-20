# Roadmap

lode's north star: a fast, **drop-in byte-compatible** acceleration of DVC's
data-versioning core — your existing DVC repo, faster, with zero migration.

## Shipped

- Core versioning: `add`, `status`, `checkout`, `push`, `fetch`, `pull`, `gc`
- `init` + `doctor` — standalone, no Python or DVC needed to start
- `verify` — integrity check that also proves byte-compatibility with DVC
- S3-compatible remotes (AWS S3, MinIO, Cloudflare R2, Backblaze B2)
- Byte-for-byte compatibility, gated in CI against the real `dvc`
- Reproducible benchmark harness (`scripts/benchmark.sh`)

## Planned (roughly in order)

Each is specced under [`specs/`](specs/):

- **Security & supply chain** — signed release binaries (cosign), SBOM, provenance;
  harden credential handling.
- **Production cloud auth** — IAM instance role / IRSA / SSO for S3 (works in CI on EKS/GKE).
- **Transfer reliability** — retry/backoff and partial-failure handling for push/pull at scale.
- **State-DB robustness** — never a false "up to date" on tricky filesystems (NFS,
  recycled inodes, restored mtimes); strict/rehash mode.
- **More remotes** — GCS, then Azure.
- **Pipelines / `repro`** — the larger piece; until then lode coexists with DVC for
  pipelines (DVC stays for `dvc.yaml`/`dvc repro`).

## Out of scope (for now)

- Becoming a *new* data-versioning system with its own format. lode stays drop-in with DVC.

## Want to help?

See [CONTRIBUTING.md](CONTRIBUTING.md) and the `good first issue` label. Good entry
points live in the planned items above — pick one and open an issue to discuss.
