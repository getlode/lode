# Release-Readiness Checklist: lode v0.1.0

**Purpose**: Gate for the first public release of `lode` (github.com/getlode/lode).
**Created**: 2026-06-20
**Feature**: [spec.md](../spec.md) · **Constitution**: v1.0.0
**Legend**: `[x]` verified · `[ ]` pending action before tagging `v0.1.0`

## Licensing & Legal

- [x] CHK001 Is an OSI-recognized open-source license present at `LICENSE`? (AGPL-3.0, canonical text)
- [x] CHK002 Is the dual-licensing model documented for commercial use? [COMMERCIAL.md]
- [x] CHK003 Is a CLA documented so contributions can be relicensed (open-core viability)? [CONTRIBUTING.md]
- [x] CHK004 Is a non-affiliation disclaimer re: DVC / Iterative present to limit trademark risk? [README]
- [ ] CHK005 Does GitHub detect and display the license (`licenseInfo`)? (still lagging after push — non-blocking, self-resolves on reprocess)
- [x] CHK006 Are SPDX/license headers decided? (deliberate skip for pre-1.0; the `LICENSE` file governs — revisit at 1.0)

## Documentation

- [x] CHK007 Does the README lead with the value proposition (drop-in + 10×) for a first-time visitor? [README]
- [x] CHK008 Is the install path documented and correct for the published module (`go install github.com/getlode/lode/cmd/lode@latest`)? [README]
- [x] CHK009 Are all shipped commands documented with their purpose? [README, contracts/cli.md]
- [x] CHK010 Is the DVC compatibility guarantee stated and scoped (byte-identical, interop, legacy read)? [README]
- [x] CHK011 Is current scope vs. out-of-scope (pipelines/`repro`, non-S3 remotes) explicit to set expectations? [README, spec Assumptions]
- [x] CHK012 Is there a CHANGELOG / release-notes source for `v0.1.0`? (GoReleaser auto-generates release notes from commits)
- [x] CHK013 Do README badge URLs resolve? (CI ran green → CI badge resolves; license/Go badges are static)

## CI / Build / Quality Gates

- [x] CHK014 Does the project build CGO-free and cross-compile (Constitution III)? (verified)
- [x] CHK015 Is `go vet ./...` clean and `go test ./...` green locally? (verified)
- [x] CHK016 Does `go test -short` pass without external services (S3-less CI path)? (verified)
- [x] CHK017 Is a CI workflow defined across the target OS matrix (linux/macOS/windows)? [.github/workflows/ci.yml]
- [x] CHK018 Has CI run green on GitHub for the pushed `main`? (run 27875209502: test×3 + lint + oracle all ✓)
- [x] CHK019 Does the oracle CI job install the reference `dvc` so byte-compat is gated in CI (Constitution II)? (oracle job green)

## Distribution & Packaging

- [x] CHK020 Is the GoReleaser build matrix correct (OS × arch, `CGO_ENABLED=0`, version ldflags)? [.goreleaser.yaml]
- [x] CHK021 Has the GoReleaser config been validated locally (`goreleaser check` + `build --snapshot`)? (6/6 binaries built; removed the `go mod tidy` hook that broke the build, pinned action to v2.4.4)
- [ ] CHK022 Is the Homebrew tap publish workable, given the default `GITHUB_TOKEN` cannot push cross-repo to `getlode/homebrew-tap`? (DECISION NEEDED: wire a PAT secret, or drop `brews` for v0.1.0) [.goreleaser.yaml, release.yml]
- [x] CHK023 Does the release workflow trigger on `v*` tags with the right permissions? [.github/workflows/release.yml]
- [x] CHK024 Is the version string injected and surfaced (`lode --version` shows the tag, not `dev`) in release builds? (snapshot shows `0.0.0-SNAPSHOT-<sha>`; a tag yields the tag)

## Security & Secrets

- [x] CHK025 Are there no committed secrets/credentials in the repo? (verified via grep)
- [x] CHK026 Does `.gitignore` exclude the binary, `dist/`, and DVC cache/tmp so artifacts never ship? [.gitignore]
- [x] CHK027 Is a `SECURITY.md` (vulnerability disclosure contact/policy) present? [SECURITY.md]
- [x] CHK028 Are dependencies reviewed for advisories? (govulncheck: 0 in our code; only stdlib vulns fixed by the latest Go 1.23.x patch that CI's setup-go installs; Dependabot enabled for gomod + actions)
- [x] CHK029 Does the release workflow avoid using untrusted input in `run:` steps (injection-safe)? [release.yml]

## DVC Compatibility (core product promise)

- [x] CHK030 Is byte-identical `.dvc`/`.dir`/cache-path compatibility validated against real DVC 3.x? [tests/oracle]
- [x] CHK031 Is bidirectional interop (lode↔DVC over the same S3 remote) validated? [tests/integration]
- [x] CHK032 Is the ~10× performance claim in the README backed by a reproducible measurement? [bench_test.go, README]

## Repository Metadata & Discoverability

- [x] CHK033 Is the repo public with a clear description and relevant topics? (verified)
- [x] CHK034 Does the module path match the repository owner so `go install` resolves (`getlode`)? (verified)
- [x] CHK035 Is the `getlode/homebrew-tap` repo present for the formula? (created — usability depends on CHK022)

## Notes

- **Only one real blocker remains: CHK022** (Homebrew publish token). Everything else is green or self-resolving.
- CHK005 self-resolves (GitHub license reprocessing).
- Deferred to v0.1.1+: nothing blocking; CHK006 is a deliberate skip.
