# Feature Specification: Autenticación cloud de producción para remotes S3

**Feature Branch**: `008-cloud-auth`

**Created**: 2026-06-20

**Status**: Draft

**Input**: Hallazgo de la auditoría (persona de MLOps/Plataforma, confirmado por seguridad): el cliente S3 omite a propósito el provider de IAM (instance role), así que en CI sobre EKS/GKE —donde se autentica por **IAM role / IRSA**— lode no resuelve credenciales sin inyectar **claves estáticas de larga vida**, justo el antipatrón que seguridad prohíbe. Es el bloqueante #1 para meter lode en CI. La omisión original se hizo porque el provider de IAM crasheaba con un cliente HTTP nulo y colgaba el endpoint de metadata fuera de la nube.

## Context & Strategic Rationale *(non-normative)*

El 80% del valor para un equipo de MLOps está en CI (runners efímeros). Hoy lode solo autentica por claves explícitas o archivo de credenciales — inutilizable en EKS/GKE con roles de instancia/IRSA y en orgs con SSO. Habilitar la cadena de credenciales de producción (roles temporales rotados) desbloquea la adopción en CI sin obligar a claves estáticas, alineándose con los baselines de seguridad cloud.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Autenticar con credenciales temporales (rol de instancia / IRSA) (Priority: P1)

Un pipeline de CI corre en un runner con un rol de IAM asociado (instance role / IRSA) y lode autentica contra S3 usando esas credenciales temporales, sin claves estáticas en ningún lado.

**Why this priority**: Es el bloqueante concreto para CI en EKS/GKE, donde vive la mayor parte del uso real. Sin esto, lode no entra a producción.

**Independent Test**: En un entorno con rol de instancia/IRSA configurado (o un mock del endpoint de credenciales), correr `push`/`pull` sin claves estáticas y verificar que autentica y opera; sin el rol, falla con un error claro.

**Acceptance Scenarios**:

1. **Given** un entorno con credenciales por rol de instancia/IRSA y sin claves estáticas, **When** el usuario corre una operación remota, **Then** lode obtiene credenciales temporales y autentica correctamente.
2. **Given** que ninguna fuente de credenciales está disponible, **When** el usuario corre una operación remota, **Then** lode falla con un error accionable que explica qué fuentes intentó y cómo configurar una.
3. **Given** un entorno fuera de la nube (sin endpoint de metadata), **When** lode intenta resolver credenciales, **Then** la resolución termina rápido (con timeout acotado) sin colgarse ni crashear.

---

### User Story 2 - Perfiles y SSO (Priority: P2)

Un usuario que trabaja con perfiles de AWS (incluido SSO) puede seleccionar el perfil y lode autentica con esas credenciales.

**Why this priority**: Cubre el flujo local/dev de muchas orgs con SSO, complementario al de CI.

**Independent Test**: Con un perfil configurado, indicar el perfil y verificar que lode autentica usándolo.

**Acceptance Scenarios**:

1. **Given** un perfil de credenciales configurado, **When** el usuario lo selecciona, **Then** lode autentica usando ese perfil.

---

### Edge Cases

- **Cadena de credenciales con múltiples fuentes**: el orden de precedencia debe ser predecible y documentado (explícitas → entorno → perfil/archivo → rol de instancia/IRSA).
- **Endpoint de metadata inalcanzable o lento**: timeout corto y degradación limpia (sin cuelgues ni panics), tal como la causa que originó la omisión.
- **Credenciales temporales que expiran a mitad de una operación larga**: deben renovarse o fallar con un error claro, no corromper la operación.
- **Compatibilidad con DVC**: la configuración de credenciales debe seguir siendo legible/coherente con lo que DVC espera en el repo compartido.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: La herramienta MUST soportar credenciales por rol de instancia/IRSA (credenciales temporales del entorno de ejecución) para remotes S3, sin requerir claves estáticas.
- **FR-002**: La resolución de credenciales por endpoint de metadata MUST estar acotada por un timeout y degradar de forma limpia fuera de la nube (sin cuelgues ni fallos no controlados).
- **FR-003**: La herramienta MUST soportar la selección de perfil de credenciales (incluido el flujo SSO).
- **FR-004**: La herramienta MUST aplicar un orden de precedencia de fuentes de credenciales predecible y documentado.
- **FR-005**: Ante ausencia total de credenciales, la herramienta MUST emitir un error accionable que enumere las fuentes intentadas y cómo configurar una.
- **FR-006**: La configuración de credenciales MUST permanecer coherente con la que DVC usa sobre el mismo repo (interoperabilidad).

### Key Entities

- **Fuente de credenciales**: explícita / entorno / perfil-archivo / rol de instancia-IRSA, con su precedencia.
- **Credencial temporal**: con vencimiento y, si aplica, renovación.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En un entorno con rol de instancia/IRSA (o su mock), lode autentica y completa push/pull sin ninguna clave estática presente.
- **SC-002**: Fuera de la nube, la resolución de credenciales termina dentro de un timeout corto y nunca cuelga ni crashea el proceso.
- **SC-003**: Un usuario con perfil/SSO puede autenticar seleccionando el perfil.
- **SC-004**: Ante ausencia de credenciales, el error nombra las fuentes intentadas y la acción de configuración, en el 100% de los casos.

## Assumptions

- **Backends S3-compatible**: el alcance es la cadena de credenciales de AWS/S3-compatible; GCS/Azure son features de remote aparte.
- **Reutiliza** el cliente S3 existente, reincorporando el provider de credenciales por rol con un cliente HTTP acotado por timeout (la causa de la omisión original).
- **Sin cambios de formato**: este feature cambia la autenticación, no el layout de datos ni el comportamiento de los comandos.
