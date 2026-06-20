# Feature Specification: CLI en inglés y pulido de mensajes

**Feature Branch**: `003-english-cli`

**Created**: 2026-06-20

**Status**: Draft

**Input**: User description: "Convertir toda la salida de usuario y el help de lode a inglés, con mensajes de error accionables y un primer-run prolijo. Hoy el README está en inglés pero la CLI habla español, lo que resta credibilidad para un público global."

## Context & Strategic Rationale *(non-normative)*

lode apunta a un público global (la audiencia de DVC/MLOps es internacional y el README ya está en inglés), pero la salida de la CLI está en español ("Trackea archivos", "Materializa el workspace", "no hay remote configurado"). Esa inconsistencia entre un landing en inglés y una herramienta en español resta seriedad y frena la adopción justo cuando el usuario la prueba. Unificar todo a inglés es bajo esfuerzo y alto impacto en percepción, y de paso elimina una fricción del tooling (el corrector ortográfico en CI marcaba el español como errores).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Interfaz en inglés coherente con el producto (Priority: P1)

Un usuario de habla inglesa instala lode tras leer el README en inglés. Al correr cualquier comando —ayuda, descripciones, estado, progreso, errores— todo está en inglés, consistente con lo que esperaba.

**Why this priority**: La incoherencia idioma-landing vs idioma-CLI es un golpe de credibilidad inmediato. Es lo primero que ve quien prueba la herramienta y define si la toma en serio.

**Independent Test**: Ejecutar cada comando (incluido `--help` de cada subcomando) y los caminos de error frecuentes; verificar que ninguna cadena visible al usuario está en español.

**Acceptance Scenarios**:

1. **Given** cualquier comando o subcomando de lode, **When** el usuario pide ayuda (`--help`), **Then** el título, las descripciones de comandos y las descripciones de flags están en inglés.
2. **Given** una operación normal (add/status/push/pull/checkout/gc), **When** se ejecuta, **Then** los mensajes de progreso y resultado están en inglés.
3. **Given** un camino de error (sin repo, sin remote, objeto faltante, credenciales inválidas), **When** ocurre, **Then** el mensaje está en inglés y es accionable.

---

### User Story 2 - Mensajes de error accionables y consistentes (Priority: P2)

Un usuario que comete un error (directorio equivocado, falta configurar algo) recibe mensajes en inglés que explican el problema y nombran el siguiente paso, con una terminología consistente en toda la herramienta.

**Why this priority**: Refuerza el onboarding y reduce abandono; complementa el feature de errores que guían dándole una voz consistente.

**Independent Test**: Recorrer los errores de precondición y verificar que usan términos canónicos consistentes (p. ej. siempre "remote", "cache", "workspace") y proponen una acción.

**Acceptance Scenarios**:

1. **Given** dos errores distintos que refieren al mismo concepto, **When** se muestran, **Then** usan el mismo término canónico (sin sinónimos divergentes).
2. **Given** un error con una acción de remediación conocida, **When** se muestra, **Then** incluye el comando o paso concreto.

---

### Edge Cases

- **Cadenas en logs/verbose**: el modo detallado también debe estar en inglés.
- **Mensajes provenientes de dependencias** (errores de red/almacenamiento subyacentes): se presentan envueltos en un mensaje propio en inglés y accionable, sin volcar jerga interna en otro idioma.
- **Pluralización y formato** (1 file vs N files): los mensajes con conteos deben leerse naturalmente en inglés.
- **Términos compartidos con DVC** (remote, cache, push/pull): se usan los mismos términos que DVC para no confundir a usuarios que vienen de ahí.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Toda cadena visible al usuario MUST estar en inglés: títulos y descripciones de comandos, descripciones de flags, textos de ayuda, mensajes de progreso, mensajes de resultado, advertencias y errores.
- **FR-002**: Los mensajes de error MUST ser accionables y estar en inglés, nombrando el siguiente paso concreto cuando exista.
- **FR-003**: La herramienta MUST usar una terminología canónica y consistente (un único término por concepto: remote, cache, workspace, object, etc.), alineada con los términos que usa DVC para conceptos compartidos.
- **FR-004**: El modo detallado/verbose MUST estar también en inglés.
- **FR-005**: Los mensajes con conteos MUST leerse naturalmente (manejo de singular/plural) en inglés.
- **FR-006**: No MUST quedar ninguna cadena en español visible al usuario en ningún camino de ejecución (operación normal, vacío, error).

### Key Entities

- **Cadena de interfaz**: cualquier texto emitido al usuario (stdout/stderr) por la herramienta.
- **Glosario canónico**: el conjunto de términos preferidos y sus sinónimos prohibidos, para garantizar consistencia.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100% de las cadenas visibles al usuario (help de todos los comandos/flags, salidas normales, errores) están en inglés — verificable por inspección automatizada de la salida.
- **SC-002**: Cada comando y cada flag tiene un texto de ayuda en inglés no vacío.
- **SC-003**: Cada concepto compartido se nombra con un único término canónico en toda la herramienta (cero sinónimos divergentes para remote/cache/workspace/object).
- **SC-004**: El 100% de los mensajes de error de precondición conocidos incluyen una acción concreta sugerida.

## Assumptions

- **Alcance**: solo se establece inglés como idioma único de la interfaz. NO se construye un framework de localización/i18n multi-idioma en esta iteración (queda fuera de alcance; podría considerarse a futuro).
- **Términos compartidos con DVC**: se preservan los nombres que usa DVC (remote, cache, push, pull, checkout) por familiaridad y para no romper la mentalidad de quien migra.
- **Sin cambios de comportamiento**: este feature cambia textos, no la semántica de los comandos.
