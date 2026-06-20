# Feature Specification: Benchmark reproducible lode vs DVC

**Feature Branch**: `005-benchmark`

**Created**: 2026-06-20

**Status**: Draft

**Input**: User description: "Un comando/target que mida lode vs DVC en un dataset estándar y publique los números (README + página de resultados). Hace creíble el 10× y da material reproducible y 'screenshot-able' para el lanzamiento (Show HN)."

## Context & Strategic Rationale *(non-normative)*

El "~10× más rápido" es la propuesta de valor central, pero hoy es una afirmación apoyada en una medición manual puntual. Para el lanzamiento (Show HN / r/mlops / comunidad DVC) y para la credibilidad general, el número tiene que ser **reproducible por cualquiera con un comando** y estar **publicado con su metodología**. Un benchmark que el lector puede correr en su máquina y ver el mismo resultado convierte una afirmación de marketing en evidencia, y produce material concreto (tabla/gráfico) para el post de lanzamiento.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reproducir el benchmark con un comando (Priority: P1)

Un evaluador escéptico quiere comprobar el "10×" por sí mismo. Corre un único comando que genera un dataset estándar, mide lode y (si está disponible) DVC sobre el camino caliente, y muestra los tiempos y el speedup.

**Why this priority**: Sin reproducibilidad, el número no es creíble. Es la base del lanzamiento y de la confianza en la afirmación de performance.

**Independent Test**: Correr el comando de benchmark en una máquina limpia y obtener una tabla con tiempos de lode y DVC y el factor de speedup, de forma determinista (mismo dataset → resultados comparables).

**Acceptance Scenarios**:

1. **Given** un entorno con lode (y opcionalmente DVC) instalado, **When** el usuario corre el benchmark, **Then** se genera un dataset de tamaño definido y se reportan los tiempos del camino caliente (registrar y consultar estado) para lode y para DVC, más el factor de speedup.
2. **Given** que DVC no está instalado, **When** el usuario corre el benchmark, **Then** se reportan igualmente los tiempos de lode (con una nota de que no hubo comparación) sin fallar.
3. **Given** dos corridas en el mismo entorno con el mismo tamaño, **When** se comparan, **Then** los resultados son consistentes dentro de un margen razonable (el dataset y el método son deterministas).

---

### User Story 2 - Resultados publicados con metodología (Priority: P2)

Un lector del README o de la página de resultados ve los números, entiende cómo se midieron (tamaño del dataset, hardware, versiones, qué operación), y confía en que puede reproducirlos.

**Why this priority**: La transparencia metodológica es lo que separa un benchmark creíble de un número de marketing. Da el material para el lanzamiento.

**Independent Test**: Revisar la documentación de resultados y confirmar que incluye método, escala, entorno y versiones, y un comando exacto para reproducir.

**Acceptance Scenarios**:

1. **Given** la documentación de resultados, **When** el lector la revisa, **Then** encuentra el comando exacto para reproducir, el tamaño del dataset, las operaciones medidas, el entorno (hardware/SO) y las versiones de lode y DVC.
2. **Given** los números publicados, **When** se los compara con una corrida fresca en un entorno equivalente, **Then** coinciden dentro del margen declarado.

---

### Edge Cases

- **Escala configurable**: el benchmark permite distintos tamaños (cantidad y tamaño de archivos) para mostrar cómo escala la ventaja.
- **Ruido de medición**: el método mitiga el ruido (p. ej. descartando la primera corrida en frío o reportando varias) y lo documenta.
- **Operaciones medidas**: cubre al menos el camino caliente (registrar y consultar estado); idealmente también sincronización con remote, marcando claramente cuál se midió.
- **Comparación justa**: ambas herramientas se miden sobre el mismo dataset y la misma operación equivalente, partiendo del mismo estado.
- **Entorno sin red**: las mediciones locales (registrar/estado) no dependen de red; las de remote se marcan como opcionales.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El benchmark MUST generar un dataset estándar y determinista, con tamaño configurable (cantidad y tamaño de archivos).
- **FR-002**: El benchmark MUST medir el camino caliente (registrar datos y consultar estado) para lode y, cuando esté disponible, para DVC, partiendo del mismo estado y dataset.
- **FR-003**: El benchmark MUST reportar los tiempos de cada herramienta y el factor de speedup en un formato legible (tabla) apto para documentación/captura.
- **FR-004**: El benchmark MUST funcionar sin DVC instalado, reportando solo lode con una nota explícita, sin fallar.
- **FR-005**: El benchmark MUST ser reproducible: misma escala y entorno → resultados consistentes dentro de un margen declarado; el método de mitigación de ruido MUST estar documentado.
- **FR-006**: La documentación de resultados MUST incluir el comando exacto de reproducción, las operaciones medidas, la escala, el entorno (hardware/SO) y las versiones de lode y DVC.
- **FR-007**: El benchmark MUST poder medir distintas escalas para mostrar cómo evoluciona la ventaja.
- **FR-008** *(opcional)*: El benchmark MAY medir la sincronización con un remote, marcando claramente que esa medición depende de red y es separada del camino caliente.

### Key Entities

- **Configuración del benchmark**: escala (cantidad/tamaño de archivos), operaciones a medir, número de repeticiones.
- **Resultado de benchmark**: por herramienta y operación, el tiempo medido; más el factor de speedup y los metadatos de entorno/versiones.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Cualquier persona puede reproducir el benchmark con un único comando documentado y obtener un resultado consistente con el publicado, dentro del margen declarado.
- **SC-002**: El reporte muestra el speedup de lode vs DVC en el camino caliente para al menos dos escalas distintas.
- **SC-003**: La documentación de resultados incluye método, escala, entorno y versiones suficientes para reproducir sin ambigüedad.
- **SC-004**: El benchmark corre y reporta correctamente tanto con DVC instalado (comparación) como sin DVC (solo lode), sin fallar.
- **SC-005**: Los resultados publicados respaldan numéricamente la afirmación de performance del README (no hay discrepancia entre lo que dice el README y lo que mide el benchmark).

## Assumptions

- **Camino caliente como foco**: la medición principal es registrar datos y consultar estado (donde está el dolor y la ventaja); la sincronización con remote es secundaria y opcional por su dependencia de red.
- **Determinismo del dataset**: el generador de datos es determinista para que las comparaciones sean justas y repetibles.
- **Comparación honesta**: se compara la operación equivalente partiendo del mismo estado; las diferencias de entorno se documentan en vez de ocultarse.
- **Reutiliza** el generador de datasets y el banco de medición ya existentes en el proyecto, formalizándolos como herramienta reproducible y publicada.
