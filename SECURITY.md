# Security Policy

## Reporting a vulnerability

Please report security issues **privately**, not via public issues.

- Preferred: open a [GitHub private vulnerability report](https://github.com/getlode/lode/security/advisories/new).
- Or email **j.s.torchia@gmail.com** with details and steps to reproduce.

You can expect an initial acknowledgement within a few business days. Please
allow reasonable time for a fix before public disclosure.

## Supported versions

lode is pre-1.0. Security fixes are applied to the latest released version.
Until 1.0, only the most recent tag is supported.

## Network egress (no telemetry)

lode has **no telemetry and no analytics**. The only network connections it makes are
to the **S3-compatible remote you configure** (its `endpointurl`, default AWS S3). There
are no other endpoints — you can verify with `grep -rn "net/http" internal/ cmd/` (no
HTTP client beyond the S3 SDK).

## Credentials

- Credentials are resolved in a predictable order: explicit config → environment →
  shared AWS credentials file (honoring the configured profile / `AWS_PROFILE`) →
  IAM (EC2 instance role, ECS, and EKS/IRSA). The IAM metadata probe is
  timeout-bounded, so off-cloud it fails fast rather than hanging. This lets CI on
  EKS/GKE authenticate with short-lived role credentials instead of static keys.
  lode does not invent its own credential store.
- **Set secrets without exposing them**: omit the value and pipe it in, so it never
  lands in argv / `ps` / shell history:
  `printf '%s' "$KEY" | lode remote modify <name> secret_access_key`
  lode never echoes secret values back.
- Note: like DVC, `.dvc/config` stores credentials in **plain text**. Prefer the
  environment / AWS credentials file over writing keys into the repo config.

## Verifying release artifacts

Release binaries are built by the tagged GitHub Actions pipeline, with:

- **`checksums.txt`** for each release;
- a **cosign keyless signature** of the checksums (`checksums.txt.sig` + `.pem`),
  verifiable with (substitute the release tag):
  ```
  cosign verify-blob checksums.txt \
    --certificate checksums.txt.pem --signature checksums.txt.sig \
    --certificate-identity-regexp 'https://github.com/getlode/lode/.github/workflows/release.yml@.*' \
    --certificate-oidc-issuer https://token.actions.githubusercontent.com
  ```
- an **SBOM** per archive (syft);
- a **GitHub build-provenance attestation** (`gh attestation verify <file> --repo getlode/lode`).
