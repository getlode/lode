# Feature Specification: Seguridad de supply-chain y manejo de credenciales

**Feature Branch**: `007-security-supplychain`

**Created**: 2026-06-20

**Status**: Draft

**Input**: Hallazgos de la auditoría pre-launch (persona de seguridad/compliance). Tres frentes: (1) los binarios de release se distribuyen **sin firmar, sin SBOM, sin provenance** — bloqueante para un review de supply-chain de una herramienta que maneja credenciales S3; (2) `lode remote modify <name> secret_access_key <valor>` toma el secreto por **argumento de línea de comandos** (queda en `ps`/historial de shell) y lo **devuelve por stdout** (eco en pantalla/logs de CI); (3) faltan `govulncheck` en CI y las dependencias `x/crypto`/`x/net`/`x/text` están en versiones viejas.

## Context & Strategic Rationale *(non-normative)*

Una CLI que toca credenciales de almacenamiento y datos versionados tiene que pasar el escrutinio de seguridad de una org antes de ser adoptada. Hoy, un binario sin firmar de un repo de un solo mantenedor no pasa ese review (se descarga "compilá desde fuente vos mismo"), y la fuga de secreto en `remote modify` es un antipatrón clásico. Cerrar estos frentes desbloquea adopción empresarial y eleva la confianza general, con cambios acotados y de alto impacto reputacional.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Binarios verificables (firma + SBOM + provenance) (Priority: P1)

Un ingeniero de seguridad que evalúa lode puede **verificar criptográficamente** que el binario que descarga fue construido por el pipeline oficial y no fue manipulado, y obtener una lista de materiales (SBOM) de lo que contiene.

**Why this priority**: Sin firma ni provenance no hay forma de verificar integridad/autoría más allá de un checksum que vive en el mismo release manipulable. Es el bloqueante de supply-chain.

**Independent Test**: Descargar un binario de release y su firma, y verificar la firma con la herramienta estándar (keyless/OIDC) contra la identidad del pipeline; obtener el SBOM del release y la attestation de provenance.

**Acceptance Scenarios**:

1. **Given** un binario de release, **When** el usuario verifica su firma, **Then** la verificación confirma que fue producido por el pipeline oficial (identidad/repo correctos) y no fue alterado.
2. **Given** un release, **When** el usuario lo inspecciona, **Then** encuentra un SBOM y una attestation de provenance asociadas a los artefactos.

---

### User Story 2 - Credenciales sin fuga (Priority: P1)

Un usuario configura un remote sin que su `secret_access_key` quede expuesto en la tabla de procesos, el historial de shell, ni se imprima de vuelta en pantalla/logs.

**Why this priority**: La exposición de secretos por argv/stdout es un riesgo concreto y se da en un comando de uso habitual. Es una falla de seguridad real, no teórica.

**Independent Test**: Configurar el secreto de un remote y verificar que el valor no aparece en la línea de comando observable (`ps`), no se imprime en la salida, y que existe una vía para ingresarlo sin pasarlo como argumento.

**Acceptance Scenarios**:

1. **Given** la configuración del secreto de un remote, **When** el usuario lo ingresa, **Then** existe una vía que no lo expone como argumento de línea de comando (p. ej. variable de entorno, stdin, o prompt sin eco).
2. **Given** que se configuró un secreto, **When** la herramienta confirma la operación, **Then** NO imprime el valor del secreto en su salida.
3. **Given** que el almacenamiento de configuración guarda secretos en texto plano (heredado de DVC), **When** el usuario lo usa, **Then** la herramienta lo advierte y recomienda la vía de entorno/credenciales en lugar de claves en config.

---

### User Story 3 - Higiene de dependencias y transparencia de red (Priority: P2)

Un revisor confirma que las dependencias no arrastran vulnerabilidades conocidas y que la herramienta no contacta ningún endpoint fuera del remote que el usuario configura.

**Why this priority**: Reduce la superficie de riesgo y desarma el miedo común ("¿manda algo a la red?"), con bajo esfuerzo.

**Independent Test**: Correr el escáner de vulnerabilidades sobre el proyecto y verificar cero hallazgos en código alcanzable; revisar la documentación de seguridad y confirmar que lista los endpoints que la herramienta contacta (solo el remote configurado).

**Acceptance Scenarios**:

1. **Given** el proyecto, **When** se corre el escáner de vulnerabilidades en CI, **Then** no hay vulnerabilidades en el código que la herramienta efectivamente llama, y las dependencias relevantes están en versiones soportadas.
2. **Given** la documentación de seguridad, **When** el usuario la lee, **Then** encuentra la afirmación explícita de "sin telemetría" y la lista de endpoints de red que la herramienta contacta.

---

### Edge Cases

- **Verificación de firma sin conectividad**: documentar cómo verificar offline (checksums + firma) si el verificador requiere red.
- **Secreto provisto por entorno/credenciales ya existentes**: la vía sin-fuga no debe romper a quien ya usa la cadena de credenciales estándar.
- **CI de release sin los permisos para attestation**: el pipeline debe declarar los permisos necesarios y fallar claramente si faltan, no publicar artefactos sin firmar silenciosamente.
- **Falsos positivos del escáner**: contemplar una vía documentada para suprimir/justificar hallazgos no alcanzables sin desactivar el gate.

## Requirements *(mandatory)*

### Functional Requirements

#### Firma, SBOM y provenance
- **FR-001**: Los artefactos de release MUST estar firmados de forma verificable (keyless/OIDC contra la identidad del pipeline oficial).
- **FR-002**: Cada release MUST publicar un SBOM de los artefactos.
- **FR-003**: Cada release MUST publicar una attestation de provenance (build) verificable.
- **FR-004**: La documentación MUST explicar cómo verificar firma, SBOM y provenance.

#### Credenciales
- **FR-005**: La herramienta MUST ofrecer una vía para configurar secretos de remote que no los exponga como argumentos de línea de comando.
- **FR-006**: La herramienta MUST NOT imprimir valores de secretos en su salida.
- **FR-007**: La herramienta MUST advertir que la configuración guarda secretos en texto plano y recomendar la cadena de entorno/credenciales.

#### Dependencias y red
- **FR-008**: El CI MUST incluir un escaneo de vulnerabilidades de dependencias que falle ante vulnerabilidades en código alcanzable.
- **FR-009**: Las dependencias relevantes MUST mantenerse en versiones soportadas (actualizar las desactualizadas).
- **FR-010**: La documentación de seguridad MUST afirmar la ausencia de telemetría y listar los endpoints de red que la herramienta contacta (solo el remote configurado).

### Key Entities

- **Artefacto de release**: binario + firma + checksum + SBOM + provenance.
- **Configuración de credencial**: el secreto y su vía de ingreso (entorno/stdin/prompt) y de almacenamiento.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100% de los binarios de release pueden verificarse criptográficamente contra la identidad del pipeline oficial, y traen SBOM + provenance.
- **SC-002**: El secreto de un remote puede configurarse sin que aparezca en la tabla de procesos ni en el historial de shell, y nunca se imprime en la salida.
- **SC-003**: El escaneo de vulnerabilidades corre en CI y no reporta vulnerabilidades en código alcanzable; las dependencias desactualizadas señaladas en la auditoría quedan actualizadas.
- **SC-004**: La documentación de seguridad lista los endpoints de red y afirma la ausencia de telemetría, verificable contra el código.

## Assumptions

- **Herramientas estándar**: se asume el uso de utilidades reconocidas de firma keyless, generación de SBOM y attestations de provenance integradas al pipeline de release.
- **AGPL y uso interno**: el frente legal/licencia se aborda en el feature de confianza/posicionamiento (009); aquí solo se cubre la superficie técnica de seguridad.
- **Compatibilidad**: ninguna de estas medidas cambia el formato de datos ni el comportamiento de los comandos (Constitución I intacta).
- **Sin telemetría es un invariante**: la herramienta no introduce ninguna llamada de red fuera del remote configurado.
