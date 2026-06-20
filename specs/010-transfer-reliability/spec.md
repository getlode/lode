# Feature Specification: Confiabilidad de transferencia (retry/backoff y fallos parciales)

**Feature Branch**: `010-transfer-reliability`

**Created**: 2026-06-20

**Status**: Draft

**Input**: Hallazgo de la auditoría (greybeard + MLOps): en el path de red de `push`/`pull` no hay retry/backoff a nivel aplicación. Para el caso de uso que el producto vende —muchos archivos— un push de 100k objetos contra S3/R2 va a comer throttling y errores 5xx transitorios, y hoy la herramienta solo cuenta los fallos sin reintentar ni manejar bien el fallo parcial.

## Context & Strategic Rationale *(non-normative)*

La propuesta de valor son los datasets grandes (cientos de miles a millones de objetos). A esa escala, los errores transitorios de red/almacenamiento (throttling, 503, cortes momentáneos) son la norma, no la excepción. Sin reintentos con backoff y un manejo claro del fallo parcial, una transferencia grande falla por un hipo transitorio y obliga a empezar de cero o deja al usuario sin saber qué quedó. Endurecer esto es necesario para que el producto cumpla su promesa en su caso de uso estrella.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Resistir errores transitorios (Priority: P1)

Un usuario sube/baja cientos de miles de objetos y la transferencia sobrevive a los errores transitorios habituales (throttling, 5xx, cortes momentáneos) reintentando con espera creciente, en vez de abortar al primer hipo.

**Why this priority**: A escala, los transitorios son inevitables; sin resiliencia, la operación estrella (muchos archivos) falla en la práctica.

**Independent Test**: Inyectar errores transitorios en una fracción de las operaciones de un push/pull grande y verificar que la transferencia completa igual, reintentando los afectados.

**Acceptance Scenarios**:

1. **Given** un objeto cuya transferencia falla con un error transitorio, **When** ocurre, **Then** la herramienta reintenta con espera creciente (backoff) hasta un límite, antes de darlo por fallido.
2. **Given** throttling del almacenamiento (límite de tasa), **When** se alcanza, **Then** la herramienta reduce el ritmo/espera y continúa, en vez de saturar y fallar.
3. **Given** un error permanente (credenciales inválidas, objeto inexistente), **When** ocurre, **Then** NO se reintenta inútilmente y se reporta de inmediato.

---

### User Story 2 - Fallo parcial claro y reanudable (Priority: P2)

Si tras los reintentos algunos objetos no se pudieron transferir, el usuario obtiene un reporte claro de qué quedó pendiente y puede reanudar para completar solo lo faltante.

**Why this priority**: A escala, un fallo parcial debe ser informativo y recuperable, no un "falló" opaco que obligue a rehacer todo.

**Independent Test**: Forzar que un subconjunto falle de forma permanente, verificar que el reporte lista exactamente lo no transferido, y que una reanudación posterior completa solo lo faltante.

**Acceptance Scenarios**:

1. **Given** que algunos objetos fallan tras agotar reintentos, **When** termina la operación, **Then** el reporte indica cuántos y cuáles quedaron pendientes y el código de salida refleja el fallo parcial.
2. **Given** una transferencia con fallo parcial, **When** el usuario la reanuda, **Then** solo se transfiere lo faltante (idempotencia), sin reduplicar lo ya hecho.

---

### Edge Cases

- **Tormenta de reintentos**: el backoff debe incluir jitter para no sincronizar reintentos y empeorar el throttling.
- **Límite de reintentos alcanzado**: debe haber un tope configurable y un comportamiento definido al alcanzarlo (reportar, no colgar).
- **Distinguir transitorio de permanente**: clasificar correctamente los errores para no reintentar lo irreintentable ni rendirse ante lo transitorio.
- **Cancelación del usuario**: un corte (Ctrl-C) durante la transferencia debe dejar un estado consistente y reanudable.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: La transferencia (`push`/`pull`/`fetch`) MUST reintentar los errores transitorios con backoff (espera creciente) y jitter, hasta un límite configurable.
- **FR-002**: La herramienta MUST clasificar errores en transitorios (reintentables) y permanentes (no reintentables) y actuar en consecuencia.
- **FR-003**: La herramienta MUST adaptarse al throttling del almacenamiento reduciendo el ritmo/esperando, en vez de saturar.
- **FR-004**: Ante fallo parcial tras agotar reintentos, la herramienta MUST reportar claramente cuántos y cuáles objetos quedaron pendientes, con un código de salida acorde.
- **FR-005**: Una transferencia con fallo parcial MUST ser reanudable de forma idempotente (solo lo faltante, sin reduplicar).
- **FR-006**: La cancelación por el usuario MUST dejar un estado consistente y reanudable (sin objetos a medio escribir visibles como válidos).

### Key Entities

- **Política de reintentos**: máximo de intentos, base/tope de backoff, jitter, clasificación de errores.
- **Resultado de transferencia**: transferidos / omitidos / fallidos (con detalle de los pendientes).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Con una fracción de operaciones fallando transitoriamente (p. ej. inyección de 5xx en X%), una transferencia grande completa al 100% gracias a los reintentos.
- **SC-002**: Los errores permanentes se reportan de inmediato sin reintentos inútiles.
- **SC-003**: Ante fallo parcial, el reporte lista exactamente lo pendiente y una reanudación posterior lo completa sin reduplicar.
- **SC-004**: Una cancelación a mitad de transferencia no deja estado corrupto; la reanudación converge a un resultado íntegro.

## Assumptions

- **Reutiliza** el motor de transferencia existente (orden .dir-tras-contenidos, idempotencia, verificación de integridad) agregándole la capa de resiliencia.
- **Clasificación de errores**: se asume poder distinguir transitorios de permanentes a partir de los códigos/errores del almacenamiento.
- **Sin cambios de formato**: cambia la robustez de la transferencia, no el layout de objetos ni la compatibilidad.
