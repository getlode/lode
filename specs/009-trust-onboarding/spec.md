# Feature Specification: Confianza, posicionamiento y onboarding de contribuidores

**Feature Branch**: `009-trust-onboarding`

**Created**: 2026-06-20

**Status**: Draft

**Input**: Hallazgos de la auditoría pre-launch (greybeard HN + maintainer OSS + seguridad). El producto técnico es fuerte, pero la **narrativa de confianza** está floja y hay fricciones de adopción/contribución: (1) no se comunica el contraargumento clave al "bus factor 1" —que el repo sigue siendo un repo DVC, reversible—; (2) el README sobrevende "reemplazo / sin Python" cuando, sin pipelines, convive con DVC; (3) la AGPL frena a OSPOs aunque el uso CLI interno esté exento; (4) el CLA con relicencia comercial + un "bot de CLA" prometido-pero-inexistente ahuyenta contribuidores; (5) faltan ROADMAP/governance/good-first-issues; (6) queda un error **visible al usuario en español** (`fetch.go`, ruta de corrupción) que el sweep de inglés no cubre, más mensajes de test en español.

## Context & Strategic Rationale *(non-normative)*

Para un proyecto open-source nuevo de un solo mantenedor, la barrera de adopción no es el código (es bueno) sino la **confianza** y la **fricción de contribución**. El mejor activo de confianza ya existe y no se está comunicando: **cero lock-in** — tu repo sigue siendo un repo DVC; si lode desaparece, desinstalás y seguís con `dvc`. Comunicar eso, posicionar honestamente (acelerador que convive con DVC), desactivar el miedo a la AGPL para uso interno, y bajar la fricción para contribuir (governance, roadmap, good-first-issues, CLA justo) es lo que convierte "lindo, star" en adopción y masa crítica — lo que el modelo open-core necesita para no morir como intentos previos.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Confianza por reversibilidad y posicionamiento honesto (Priority: P1)

Un evaluador escéptico llega al README y, en los primeros segundos, entiende dos cosas: (a) que adoptar lode es **reversible y sin lock-in** (su repo sigue siendo DVC; puede volver a `dvc` cuando quiera), y (b) qué hace y qué **no** hace lode hoy (acelera el camino caliente y convive con DVC; no reemplaza pipelines).

**Why this priority**: Es el principal disolvente del miedo "proyecto nuevo de un dev tocando mis datos". Sin esto, el producto se subvalora y el lanzamiento pierde la objeción #1.

**Independent Test**: Una persona que no conoce el proyecto lee el README y puede responder: "¿qué pasa con mis datos si abandono lode?" (respuesta: nada, sigue siendo un repo DVC) y "¿reemplaza a DVC o convive?" (convive; acelera el hot path).

**Acceptance Scenarios**:

1. **Given** el README, **When** un evaluador lo lee, **Then** encuentra de forma prominente la garantía de reversibilidad/cero lock-in (con referencia a la prueba de interoperabilidad) cerca del inicio.
2. **Given** el README, **When** lo lee, **Then** el posicionamiento es honesto sobre el alcance actual (acelerador que convive con DVC; pipelines fuera de alcance), sin sobrevender "reemplazo total".
3. **Given** el README, **When** lo lee un usuario empresarial, **Then** encuentra una nota clara de que el uso de la CLI interno bajo AGPL no requiere licencia comercial.

---

### User Story 2 - Onboarding de contribuidores sin fricción tóxica (Priority: P1)

Un potencial contribuidor evalúa el proyecto y encuentra un camino claro para empezar (issues etiquetados, mapa de arquitectura, roadmap), términos de contribución justos y sin promesas incumplidas (el mecanismo de aceptación de contribución existe de verdad).

**Why this priority**: Bajar el bus factor depende de atraer contribuidores; el CLA tóxico + cero governance + un bot prometido que no existe los ahuyenta justo cuando más se necesitan.

**Independent Test**: Un contribuidor nuevo llega al repo y puede identificar por dónde empezar (issues "good first issue"), entender la arquitectura (doc/mapa), y sabe exactamente qué términos acepta al contribuir (mecanismo real, no aspiracional).

**Acceptance Scenarios**:

1. **Given** el repo, **When** un contribuidor lo explora, **Then** encuentra un roadmap público, un mapa de arquitectura, y varias tareas de entrada claras ("good first issue").
2. **Given** la guía de contribución, **When** un contribuidor la lee, **Then** los términos (CLA o equivalente) son claros, justos y el mecanismo de aceptación referenciado **existe de verdad** (no se promete algo inexistente).
3. **Given** una contribución pequeña, **When** se la envía, **Then** el proceso de aceptación de términos funciona como está documentado.

---

### User Story 3 - Salida 100% en inglés, también en rutas de error (Priority: P2)

Toda la salida visible al usuario está en inglés, incluidas las rutas de error poco frecuentes (p. ej. corrupción detectada en `fetch`) que el sweep actual no cubre.

**Why this priority**: Una cadena en español en una ruta de error delata "proyecto de una persona que piensa en español" y rompe la coherencia profesional; es una barrera (menor pero real) para contribuidores internacionales.

**Independent Test**: Forzar la ruta de error de corrupción en `fetch` y verificar que el mensaje está en inglés; extender el sweep automatizado para cubrir rutas de error, no solo `--help`.

**Acceptance Scenarios**:

1. **Given** una corrupción detectada durante `fetch`, **When** se reporta el error, **Then** el mensaje está en inglés.
2. **Given** el sweep de inglés automatizado, **When** corre, **Then** cubre las rutas de error y de operación además de `--help`, y queda como gate.

---

### Edge Cases

- **Reversibilidad mal entendida**: la garantía de "cero lock-in" debe ser literal y verificable (el repo es operable por DVC), no una promesa vaga.
- **Roadmap que envejece**: el roadmap debe diferenciar lo hecho de lo planeado y no prometer fechas que no se cumplen.
- **CLA vs DCO**: la decisión de términos de contribución debe quedar explícita y consistente entre el documento y el mecanismo real.
- **Mensajes de test en español**: no son visibles al usuario final, pero afectan al contribuidor; el alcance debe declarar si entran o no.

## Requirements *(mandatory)*

### Functional Requirements

#### Confianza y posicionamiento
- **FR-001**: El README MUST comunicar de forma prominente y temprana la reversibilidad/cero lock-in (el repo sigue siendo operable por DVC), con referencia a la prueba de interoperabilidad.
- **FR-002**: El README MUST posicionar honestamente el alcance actual (acelerador que convive con DVC; pipelines y backends no-S3 fuera de alcance), sin sobrevender "reemplazo total / sin Python" donde no aplica.
- **FR-003**: El README MUST incluir una nota clara de que el uso interno de la CLI bajo AGPL no requiere licencia comercial.
- **FR-004**: El proyecto MUST incluir una nota de "estado del proyecto / quién está detrás" que aborde de frente la madurez y el bus factor, en vez de ocultarlos.

#### Onboarding y governance
- **FR-005**: El proyecto MUST publicar un roadmap que distinga lo implementado de lo planeado.
- **FR-006**: El proyecto MUST publicar un mapa de arquitectura/contribución que explique la estructura del código.
- **FR-007**: El proyecto MUST ofrecer tareas de entrada claras para nuevos contribuidores ("good first issue").
- **FR-008**: Los términos de contribución (CLA o equivalente) MUST ser claros y su mecanismo de aceptación MUST existir realmente (sin prometer herramientas inexistentes).

#### Coherencia de idioma
- **FR-009**: Toda la salida visible al usuario MUST estar en inglés, incluidas las rutas de error (el error de corrupción en `fetch` corregido).
- **FR-010**: El gate automatizado de inglés MUST cubrir rutas de error y de operación, no solo `--help`.

### Key Entities

- **Documento de confianza**: las secciones del README (reversibilidad, posicionamiento, nota AGPL, estado del proyecto).
- **Artefactos de contribución**: roadmap, mapa de arquitectura, issues de entrada, términos de contribución y su mecanismo.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Un lector nuevo del README puede, en menos de 1 minuto, articular correctamente qué pasa con sus datos si abandona lode (nada: sigue siendo un repo DVC) y si lode reemplaza o convive con DVC.
- **SC-002**: El README contiene la garantía de reversibilidad, el posicionamiento honesto, la nota de AGPL/uso interno y la nota de estado del proyecto.
- **SC-003**: El repo expone roadmap, mapa de arquitectura y ≥3 tareas de entrada para contribuidores; los términos de contribución y su mecanismo son consistentes y reales.
- **SC-004**: El 100% de las rutas de salida al usuario (incluidas las de error) están en inglés, verificado por el gate automatizado ampliado.

## Assumptions

- **Reversibilidad real**: se apoya en la compatibilidad byte-a-byte ya validada (el repo es operable por DVC); este feature la comunica, no la crea.
- **Decisión de términos**: la elección concreta entre CLA y DCO (o un CLA con reciprocidad) es una decisión de governance a tomar; el spec exige que sea clara, justa y con mecanismo real, sin imponer cuál.
- **Alcance de idioma**: los mensajes de error visibles al usuario entran en alcance; la traducción de mensajes de test internos puede declararse fuera de alcance si así se decide.
- **Sin cambios de comportamiento**: salvo el texto del error en `fetch`, este feature es de documentación y onboarding, no cambia la semántica de la herramienta.
