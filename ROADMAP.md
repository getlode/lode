# Roadmap

lode's north star: a fast, DVC-compatible acceleration layer for the
data-versioning hot path — your existing DVC repo, faster, with zero migration.

## Shipped

- Core versioning: `add`, `status`, `checkout`, `push`, `fetch`, `pull`, `gc`
- `init` + `doctor` — standalone, no Python or DVC needed to start
- `verify` — integrity check that also proves byte-compatibility with DVC
- S3-compatible remotes, with MinIO covered by CI and provider-specific reports welcome
- DVC-compatible metadata for tested outputs, gated in CI against the real `dvc`
- Rigorous, reproducible benchmark harness (median±σ, order-alternated, `scripts/benchmark.sh`)
- Signed release binaries (cosign), SBOM (syft), and SLSA build provenance
- S3 credential resolution via explicit config, environment, shared credentials, and IAM role providers
- Transfer reliability — retry/backoff with jitter and resumable partial failures
- State-DB robustness, with `--rehash` for NFS, restored backups, and other untrusted-metadata cases

## Planned (roughly in order)

The current public contribution targets are:

- **More remotes** — native GCS, then Azure. (GCS *may* already work via its
  S3-compatible endpoint — `lode remote modify <name> endpointurl https://storage.googleapis.com`
  with HMAC keys — but this is **untested**; reports welcome. Native `gs://` auth is planned.)
- **Pipelines / `repro`** — the larger piece; until then lode coexists with DVC for
  pipelines (DVC stays for `dvc.yaml`/`dvc repro`).

## Out of scope (for now)

- Becoming a *new* data-versioning system with its own format. lode keeps using DVC-compatible storage for the supported command surface.

## Want to help?

See [CONTRIBUTING.md](CONTRIBUTING.md) and the `good first issue` label. Good entry
points live in the planned items above — pick one and open an issue to discuss.
