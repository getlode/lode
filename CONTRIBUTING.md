# Contributing to lode

Thanks for your interest in improving lode. This guide covers how to contribute
and the licensing terms for contributions.

## Developer Certificate of Origin (DCO)

lode is licensed under the [MPL-2.0](LICENSE). Contributions are accepted under a
**Developer Certificate of Origin** — no copyright assignment, no relicensing rights.
You keep ownership of your work; you only certify that you have the right to submit it
under the project's license.

Sign off each commit with `git commit -s`, which appends a `Signed-off-by` line. By
doing so you certify the [DCO 1.1](https://developercertificate.org/): that you wrote
the change (or have the right to submit it under the project's license) and that it may
be distributed publicly under that license.

That's it — no CLA, no agreement to let anyone relicense your work commercially.

## Development setup

```bash
make build        # CGO_ENABLED=0 single binary
make test-short   # unit + oracle (no external services)
make test         # full suite (needs MinIO + the real dvc binary, see README)
make lint
```

## Ground rules

The project has a few non-negotiable invariants that most affect contributions:

- **DVC byte-compatibility is non-negotiable.** Anything that changes a
  serialized artifact (`.dvc`, `.dir`, cache/remote layout) MUST keep the
  byte-oracle test (vs the real `dvc`) green.
- **Zero CGO.** All dependencies must build with `CGO_ENABLED=0`.
- **Don't regress the hot path.** `add`/`status` performance is the product;
  avoid per-file fsync / per-file DB transactions.

## Pull requests

- Keep changes focused; one logical change per PR.
- Include or update tests. Format-affecting changes MUST include an oracle test.
- Run `make test-short` and `go vet ./...` before submitting.
- Write commit messages in imperative English (e.g., "Add S3 retry backoff").
