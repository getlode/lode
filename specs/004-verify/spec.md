# Feature Specification: `lode verify` — verificación de integridad y compatibilidad

**Feature Branch**: `004-verify`

**Created**: 2026-06-20

**Status**: Draft

**Input**: User description: "Un comando `lode verify` que pruebe la compatibilidad y la integridad: corre lode (y dvc si está) sobre el repo y demuestra que produce hashes/objetos idénticos, y chequea que el cache/remote no estén corruptos. Convierte al escéptico que teme que una herramienta nueva le rompa los datos."

## Context & Strategic Rationale *(non-normative)*

La barrera #1 para adoptar un reemplazo drop-in de DVC es la confianza: "¿una herramienta nueva no me va a corromper o divergir de mis datos?". lode ya es byte-compatible y lo valida en su propia suite, pero el usuario escéptico no corre nuestra suite — necesita poder **probarlo él mismo, sobre su repo, en un comando**. `lode verify` convierte esa desconfianza en evidencia: verifica que los datos versionados están íntegros y, cuando hay un DVC instalado, demuestra que lode y DVC producen exactamente los mismos hashes y objetos para ese repo. Es una herramienta de adopción tanto como de operación.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Verificar integridad de los datos versionados (Priority: P1)

Un usuario quiere confirmar que su cache y sus metadatos están sanos (que nada se corrompió y que todo lo referenciado existe). Corre `lode verify` y obtiene un reporte: qué objetos están presentes e íntegros, qué falta, qué está corrupto.

**Why this priority**: Es la base de la confianza y una operación útil por sí misma (detectar corrupción antes de que duela). No depende de tener DVC instalado.

**Independent Test**: Sobre un repo con datos trackeados, correr `lode verify` y obtener "todo íntegro". Luego corromper/borrar un objeto del cache y verificar que `verify` lo detecta y lo reporta con precisión.

**Acceptance Scenarios**:

1. **Given** un repo con datos trackeados y su cache completo e íntegro, **When** el usuario corre `lode verify`, **Then** reporta que todos los objetos referenciados están presentes y su contenido coincide con su hash, y termina con éxito.
2. **Given** un objeto del cache cuyo contenido no coincide con su hash, **When** el usuario corre `lode verify`, **Then** lo reporta como corrupto e identifica a qué archivo/versión pertenece.
3. **Given** metadatos que referencian objetos ausentes del cache, **When** el usuario corre `lode verify`, **Then** lista exactamente qué falta y sugiere recuperarlo (`lode pull`).

---

### User Story 2 - Demostrar igualdad byte-a-byte con DVC (Priority: P1)

Un usuario escéptico que ya usa DVC quiere comprobar, antes de confiarle datos a lode, que lode produce exactamente lo mismo que DVC. Con DVC instalado, corre `lode verify` en modo cruzado y obtiene la prueba: para los datos de su repo, lode y DVC generan hashes y objetos idénticos.

**Why this priority**: Ataca directamente la barrera de adopción. Es la diferencia entre "confiá en mí" y "compruébalo vos mismo en tu repo".

**Independent Test**: En un repo con DVC instalado, correr la verificación cruzada y confirmar que reporta igualdad total; introducir artificialmente una divergencia simulada y confirmar que la detecta.

**Acceptance Scenarios**:

1. **Given** un repo y un DVC instalado, **When** el usuario corre la verificación cruzada contra DVC, **Then** reporta si los hashes/objetos que produce lode son idénticos a los de DVC para los mismos datos, con un veredicto claro (idéntico / divergente y en qué).
2. **Given** que no hay DVC instalado, **When** el usuario corre la verificación cruzada, **Then** la herramienta lo informa y continúa con la verificación de integridad propia, sin fallar de forma confusa.

---

### User Story 3 - Verificar el remote (Priority: P2)

Antes de borrar el cache local o de confiar en una copia remota, el usuario quiere confirmar que el remote tiene todos los objetos referenciados y que están íntegros. Corre `lode verify` apuntando al remote.

**Why this priority**: Da seguridad operativa (la copia remota es confiable), pero es secundario respecto a la integridad local y la prueba de compatibilidad.

**Independent Test**: Con un remote configurado, verificar que reporta completitud; quitar/corromper un objeto en el remote y confirmar que lo detecta.

**Acceptance Scenarios**:

1. **Given** un remote configurado con todos los objetos referenciados, **When** el usuario corre la verificación de remote, **Then** reporta completitud e integridad.
2. **Given** un remote al que le falta un objeto referenciado, **When** el usuario corre la verificación de remote, **Then** lista lo que falta.

---

### Edge Cases

- **Repo grande**: la verificación debe ser eficiente (aprovechar paralelismo) y poder acotarse a un target específico.
- **Objeto de directorio (`.dir`)**: verificar tanto el manifiesto como sus contenidos referenciados.
- **Formato legacy 2.x**: la verificación debe contemplar el layout/algoritmo legacy al chequear integridad de esos objetos.
- **Remote inalcanzable o credenciales inválidas**: distinguir "no se pudo verificar" de "verificado y falta algo".
- **Verificación cruzada cuando la versión de DVC instalada difiere** de la de referencia: reportar la versión usada y advertir si no es comparable.

## Requirements *(mandatory)*

### Functional Requirements

#### Integridad local
- **FR-001**: `lode verify` MUST comprobar que todos los objetos referenciados por los metadatos vigentes están presentes en el cache.
- **FR-002**: `lode verify` MUST recalcular el hash de cada objeto verificado y reportar como corrupto cualquiera cuyo contenido no coincida, identificando a qué archivo/versión pertenece.
- **FR-003**: `lode verify` MUST verificar los objetos de directorio (`.dir`): el manifiesto y la presencia/integridad de sus contenidos.
- **FR-004**: `lode verify` MUST listar exactamente los objetos faltantes y sugerir la acción de recuperación.

#### Compatibilidad cruzada con DVC
- **FR-005**: Cuando hay un DVC disponible, `lode verify` MUST poder demostrar si los hashes/objetos que produce lode para los datos del repo son idénticos a los de DVC, con un veredicto claro (idéntico / divergente y en qué).
- **FR-006**: Cuando no hay DVC disponible, la verificación cruzada MUST informarlo y continuar con el resto sin fallar de forma confusa.
- **FR-007**: La verificación cruzada MUST reportar la versión de DVC utilizada y advertir si no es comparable con la de referencia.

#### Remote
- **FR-008**: `lode verify` MUST poder verificar que el remote configurado contiene todos los objetos referenciados e (opcionalmente) su integridad.
- **FR-009**: La verificación de remote MUST distinguir "no se pudo verificar" (inalcanzable/credenciales) de "verificado y falta/algo corrupto".

#### Reporte y uso
- **FR-010**: `lode verify` MUST poder acotarse a un target específico y aprovechar paralelismo para repos grandes.
- **FR-011**: `lode verify` MUST emitir un reporte legible y una salida estructurada (`--json`), y terminar con código de salida cero si todo está OK y distinto de cero si hay problemas.

### Key Entities

- **Resultado de verificación**: por objeto/target, un estado (presente-íntegro / faltante / corrupto / no-verificable) con su detalle.
- **Veredicto de compatibilidad**: idéntico vs divergente respecto a DVC, con el detalle de dónde difiere si aplica.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `lode verify` detecta el 100% de las corrupciones y ausencias sembradas (objeto modificado, objeto borrado, contenido de `.dir` faltante) y reporta su ubicación con precisión.
- **SC-002**: En un repo donde lode y DVC son compatibles, la verificación cruzada reporta "idéntico" con cero falsos negativos; ante una divergencia sembrada, la detecta con cero falsos positivos.
- **SC-003**: La verificación de remote identifica el 100% de los objetos faltantes/corruptos sembrados en el remote.
- **SC-004**: Los códigos de salida son correctos (cero = sano, distinto de cero = problema) para uso en scripts/CI, en el 100% de los casos.
- **SC-005**: Un usuario escéptico puede, en su propio repo y con un solo comando, obtener evidencia de igualdad byte-a-byte con DVC (cuando DVC está instalado).

## Assumptions

- **Reutiliza la verificación de integridad** ya existente en las transferencias (re-hash y descarte de objetos corruptos); este feature la expone como comando de auditoría de primera clase.
- **La verificación cruzada** requiere un DVC instalado; sin él, se omite esa parte (no es un requisito duro del comando).
- **Alcance**: `verify` reporta y diagnostica; no repara automáticamente (sugiere `pull`/`gc`/re-add según corresponda).
- **Integridad de remote**: por defecto verifica presencia; la verificación de contenido remoto puede implicar descarga y por eso ser opcional/explícita.
