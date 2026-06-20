# Contributing to lode

Thanks for your interest in improving lode. This guide covers how to contribute
and the licensing terms for contributions.

## Contributor License Agreement (CLA)

lode is dual-licensed (AGPL-3.0 + a commercial license — see [COMMERCIAL.md](COMMERCIAL.md)).
To keep that model viable, **every contribution requires agreeing to the CLA**.

By submitting a pull request you agree that:

1. You have the right to submit the contribution (it is your original work, or
   you are authorized to contribute it).
2. You grant the project maintainer a perpetual, worldwide, irrevocable license
   to use, modify, and **relicense** your contribution, including under the
   commercial license. This is what allows the project to offer a non-AGPL
   commercial option while staying open source.
3. Your contribution is and remains available under the AGPL-3.0.

The first time you open a PR, add a comment stating: *"I have read the CLA and I agree
to it."* A maintainer will acknowledge it. You only need to do this once. (This is a
manual step today — no CLA bot is wired yet; honesty over theater.)

> Rationale: you cannot sell or relicense code you do not have the rights to.
> The CLA is what makes the open-core model legally possible.

## Development setup

```bash
make build        # CGO_ENABLED=0 single binary
make test-short   # unit + oracle (no external services)
make test         # full suite (needs MinIO + the real dvc binary, see README)
make lint
```

## Ground rules (from the project constitution)

The project has non-negotiable invariants — see
[`.specify/memory/constitution.md`](.specify/memory/constitution.md). The ones
that most affect contributions:

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
