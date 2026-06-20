---
description: "Task list for 007-security-supplychain"
---

# Tasks: Seguridad supply-chain y credenciales

## Phase 1: Firma, SBOM, provenance (release config)

- [X] T001 `.goreleaser.yaml`: bloque `signs` (cosign keyless sobre checksums) per FR-001
- [X] T002 `.goreleaser.yaml`: bloque `sboms` (syft por archive) per FR-002
- [X] T003 `release.yml`: cosign-installer + download-syft + permisos `id-token`/`attestations` + paso `attest-build-provenance` per FR-003
- [X] T004 SECURITY.md: cómo verificar firma/SBOM/provenance per FR-004
- [X] T005 `goreleaser check` valida la config

## Phase 2: Credenciales

- [X] T006 `remote modify`: leer el secreto de stdin (no argv) y nunca eco-arlo; `isSecretOption`/`optionValue` en internal/cli/remote.go per FR-005/006
- [X] T007 SECURITY.md: advertir que `.dvc/config` guarda secretos en texto plano y recomendar env/credenciales per FR-007
- [X] T008 Test de integración: el secreto no aparece en la salida y se guarda (tests/integration/security_test.go)

## Phase 3: Deps y red

- [X] T009 govulncheck en CI (job informativo) per FR-008
- [X] T010 Bump x/crypto/x/net/x/text a versiones soportadas per FR-009
- [X] T011 SECURITY.md: afirmar no-telemetría y listar endpoints (solo el remote configurado) per FR-010

## Phase 4: Validación

- [X] T012 build + vet + lint + suite verdes; `goreleaser check` OK
