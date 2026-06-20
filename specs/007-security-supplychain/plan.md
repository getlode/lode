# Implementation Plan: Seguridad supply-chain y credenciales

**Branch**: `007-security-supplychain` | **Date**: 2026-06-20 | **Spec**: [spec.md](spec.md)

## Summary

Cerrar tres frentes de la auditoría de seguridad: firmar releases (cosign keyless) + SBOM (syft) + provenance (GitHub attestation); eliminar la fuga de secreto en `remote modify` (leer de stdin, no de argv; nunca eco-ar); y endurecer deps (bump x/crypto/x/net/x/text) + govulncheck en CI. Cambios chicos de código (un comando) + config de CI/release + docs.

## Technical Context

**Language/Version**: Go 1.23 (deps bumpeadas, compatibles). Config: GitHub Actions + GoReleaser v2.4.4 (cosign/syft/attestation).
**Testing**: test de integración del enmascarado de secreto; `goreleaser check` valida la config de firma/SBOM; build/vet/lint verdes.
**Constraints**: la firma/SBOM/provenance se ejercitan en el próximo tag (CI). govulncheck queda informativo (los hallazgos de stdlib siguen el patch del toolchain — blanco móvil; Dependabot mantiene deps frescas). Sin cambios de formato.

## Constitution Check

Sin violaciones. No toca formato/performance. III (cero cgo) intacto; deps siguen puro-Go.

## Project Structure

```text
internal/cli/remote.go    # secreto por stdin (no argv) + nunca eco-ar; isSecretOption/optionValue
.goreleaser.yaml          # sboms (syft) + signs (cosign keyless en checksums)
.github/workflows/release.yml  # cosign-installer + syft + permisos id-token/attestations + attest step
.github/workflows/ci.yml  # job govulncheck (informativo)
SECURITY.md               # endpoints (solo el remote), no-telemetría, secreto en config, cómo verificar firmas
go.mod / go.sum           # bump x/crypto v0.31, x/net v0.33, x/text v0.21
tests/integration/security_test.go  # gate: el secreto no se eco-a y se guarda
```

**Structure Decision**: cambio de superficie de seguridad sin tocar el core. La firma/SBOM/provenance son config de release (efectivas en el próximo tag); el resto es verificable hoy.

## Complexity Tracking

> No aplica.
