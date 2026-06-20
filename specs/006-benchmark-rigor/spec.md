# Feature Specification: Benchmark riguroso y sin sesgo

**Feature Branch**: `006-benchmark-rigor`

**Created**: 2026-06-20

**Status**: Draft

**Input**: Hallazgos de la auditoría pre-launch (panel de 5 personas techies). El benchmark actual (feature 005) tiene defectos metodológicos que un revisor de Hacker News destrozaría: corre DVC primero y lode después sobre los mismos archivos (page cache caliente → el "cold run" de lode es en realidad warm, sesgo sistemático a favor de lode), corrida única (N=1) sobre números en el orden del ruido (centésimas de segundo), archivos de juguete de 256 bytes (mide `open()/stat()`, no hashing), sin medir memoria, y la afirmación clave ("DVC lee lo que produce lode") está en prosa, no codificada.

## Context & Strategic Rationale *(non-normative)*

El "~10× más rápido" es la propuesta de valor central y el material del lanzamiento. Pero publicar tablas con "×" en negrita mientras el harness tiene sesgo de page-cache y N=1 es pedir la credibilidad del número sin pagar su costo — y es exactamente lo que el público técnico (HN/r-mlops) ataca primero. La ventaja de lode es **estructural** (hashing paralelo + state DB que evita rehashear), no un truco; un benchmark riguroso lo va a confirmar. Este feature convierte la demo medida con cronómetro en una medición defendible y reproducible por terceros, sin sesgo.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Números creíbles y sin sesgo (Priority: P1)

Un evaluador escéptico corre el benchmark y obtiene resultados que no dependen del orden de ejecución ni del estado del page cache, repetidos varias veces, con su dispersión a la vista. Puede confiar en que el número refleja el algoritmo, no el ruido ni un sesgo.

**Why this priority**: Sin esto, el número no es publicable y el lanzamiento se apoya en una medición que el público técnico invalida en el primer comentario. Es el bloqueante de credibilidad.

**Independent Test**: Correr el benchmark dos veces invirtiendo el orden de las herramientas y verificar que el speedup reportado no cambia significativamente (sin sesgo de orden); verificar que cada celda reporta una medida agregada de N corridas con su dispersión, no una corrida única.

**Acceptance Scenarios**:

1. **Given** una herramienta corre antes que la otra y deja archivos en page cache, **When** se mide la otra, **Then** el método neutraliza el sesgo de cache (cache frío real igualado para ambas, o alternancia de orden entre corridas) de modo que el resultado no favorece sistemáticamente a ninguna.
2. **Given** una operación a medir, **When** se la ejecuta, **Then** se corre N veces (N configurable, ≥ un mínimo razonable) y se reporta una medida robusta (mediana o mínimo) junto con la dispersión (desvío o rango).
3. **Given** un resultado cuyo tiempo está en el orden del ruido de medición, **When** se lo reporta, **Then** se lo marca como tal (o se excluye) en vez de presentarlo como un speedup contundente.

---

### User Story 2 - Cobertura de regímenes reales (Priority: P2)

El benchmark cubre los regímenes que importan: muchos-archivos-chicos Y pocos-archivos-grandes (para probar el hashing CPU-bound), cold/warm/incremental claramente etiquetados, y reporta el uso de memoria — porque el motivo declarado del proyecto son datasets de cientos de miles a millones de archivos.

**Why this priority**: Los 256 bytes nunca prueban el "parallel hashing"; el público pide ver dónde la ventaja se mantiene y dónde se cierra. La memoria importa a la escala (600k–3M archivos) que el propio proyecto cita como motivación.

**Independent Test**: Correr el benchmark sobre al menos dos tamaños de archivo muy distintos (p. ej. KB y cientos de MB) y dos escalas de cantidad, y verificar que el reporte incluye el pico de memoria por operación y la etiqueta de régimen (cold/warm/incremental).

**Acceptance Scenarios**:

1. **Given** un eje de tamaño de archivo, **When** se barre de archivos chicos a grandes, **Then** el reporte muestra cómo evoluciona el speedup (incluido el régimen hash-bound donde el gap se cierra, declarado honestamente).
2. **Given** cualquier operación medida, **When** se la ejecuta, **Then** se reporta el pico de memoria residente (RSS) de cada herramienta.
3. **Given** los tres regímenes (cold real, warm, incremental), **When** se reportan, **Then** cada fila está etiquetada sin ambigüedad con su régimen.

---

### User Story 3 - Corrección codificada, no prometida (Priority: P2)

La afirmación más importante del proyecto —"DVC lee lo que produce lode"— se verifica como parte del propio harness (se ejecuta y se asevera), no se deja como una frase en la documentación.

**Why this priority**: Es el argumento de confianza central (drop-in real). Codificarlo lo vuelve verificable por cualquiera que corra el benchmark.

**Independent Test**: Tras la fase de registro de lode, el harness ejecuta la verificación con DVC (cuando está disponible) y falla/avisa si DVC no reconoce el repo; el resultado de esa aserción aparece en el reporte.

**Acceptance Scenarios**:

1. **Given** un repo recién registrado por lode, **When** el harness verifica con DVC, **Then** asevera que DVC reconoce el repo ("up to date") y refleja ese resultado en la salida.
2. **Given** que DVC no está disponible, **When** corre el harness, **Then** omite esa aserción informándolo, sin fallar de forma confusa.

---

### Edge Cases

- **Sin permisos para limpiar el page cache** (drop_caches requiere privilegios): el método debe tener una estrategia alternativa válida (alternancia de orden, o copias de datos separadas por herramienta) y declarar cuál usó.
- **Entorno ruidoso** (máquina compartida): el reporte debe capturar y exponer la dispersión para que el lector juzgue la confiabilidad.
- **DVC no instalado**: el benchmark corre igual reportando solo lode, con nota explícita.
- **Datasets grandes que no entran en disco/tiempo**: la escala debe ser parametrizable para acotar el costo, documentando qué se corrió.
- **Variabilidad de hardware** (NVMe vs SATA vs red): el reporte documenta las specs relevantes del entorno (no solo "16 cores").

## Requirements *(mandatory)*

### Functional Requirements

#### Eliminación de sesgo y rigor estadístico
- **FR-001**: El benchmark MUST neutralizar el sesgo de page cache entre herramientas: igualar el estado de cache (cache frío real para ambas) o alternar el orden de ejecución entre corridas, de modo que el resultado no favorezca sistemáticamente a ninguna.
- **FR-002**: El benchmark MUST ejecutar cada operación N veces (N configurable, con un mínimo razonable) y reportar una medida robusta (mediana o mínimo) más una medida de dispersión (desvío estándar o rango).
- **FR-003**: El benchmark MUST evitar presentar como speedup contundente cualquier resultado cuyo tiempo esté en el orden del ruido de medición (marcarlo o excluirlo).
- **FR-004**: El benchmark MUST capturar y reportar las specs relevantes del entorno (CPU/cores, tipo de disco, SO, versiones de las herramientas).

#### Cobertura de regímenes
- **FR-005**: El benchmark MUST permitir barrer el tamaño de archivo (de chicos a grandes) además de la cantidad, cubriendo al menos un régimen de muchos-archivos-chicos y uno de pocos-archivos-grandes.
- **FR-006**: El benchmark MUST etiquetar sin ambigüedad cada medición con su régimen: cold (cache frío real), warm, o incremental.
- **FR-007**: El benchmark MUST reportar el pico de memoria residente (RSS) de cada herramienta por operación.

#### Corrección codificada
- **FR-008**: El benchmark MUST ejecutar y asentar la verificación de interoperabilidad (que DVC reconoce el repo producido por lode) como parte de la corrida, cuando DVC esté disponible.
- **FR-009**: El benchmark MUST funcionar sin DVC instalado, reportando solo lode con una nota explícita, sin fallar.

#### Publicación honesta
- **FR-010**: La documentación de resultados (BENCHMARKS.md / README) MUST reflejar la metodología revisada (N, dispersión, control de cache, regímenes, memoria) y reemplazar las cifras de corrida única sesgadas por las nuevas.
- **FR-011**: La documentación MUST seguir declarando honestamente dónde lode NO gana (push/pull network-bound; single-file grande donde el gap se cierra).

### Key Entities

- **Configuración de corrida**: N repeticiones, escalas (cantidad y tamaño de archivo), regímenes a medir, estrategia de control de cache.
- **Resultado por celda**: medida robusta + dispersión + memoria + etiqueta de régimen, por herramienta y operación.
- **Reporte**: tabla/curvas con las specs del entorno y la metodología, apto para publicar.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Invirtiendo el orden de ejecución de las herramientas, el speedup reportado para una operación no cambia más allá de la dispersión declarada (ausencia de sesgo de orden demostrable).
- **SC-002**: Cada celda del reporte presenta una medida robusta de N≥ el mínimo definido, con su dispersión; ninguna fila publicada como speedup contundente se apoya en un tiempo dentro del ruido de medición.
- **SC-003**: El reporte cubre al menos dos escalas de cantidad y dos regímenes de tamaño de archivo, e incluye el pico de memoria por operación.
- **SC-004**: Un tercero puede reproducir el benchmark en un entorno equivalente y obtener resultados consistentes con los publicados dentro del margen declarado.
- **SC-005**: La aserción de interoperabilidad con DVC se ejecuta en la corrida y su resultado aparece en el reporte (cuando DVC está disponible).
- **SC-006**: La documentación publicada no contiene cifras provenientes del harness sesgado anterior; toda cifra refleja la metodología revisada.

## Assumptions

- **Reutiliza** el harness y el generador de datasets existentes (feature 005), corrigiendo su metodología; no se construye desde cero.
- **Herramienta de medición**: se asume el uso de una utilidad de benchmarking robusta (mediana/min, warmup, dispersión) en lugar de cronometraje manual de una corrida.
- **Control de cache**: si el entorno permite limpiar el page cache se usa esa vía; si no, se usa alternancia de orden o copias separadas, documentando la elección.
- **Honestidad como invariante**: ante conflicto entre "número más alto" y "número defendible", prevalece el defendible; los caveats (push/pull, single-file grande) se mantienen.
- **Alcance**: este feature endurece la medición y su publicación; no cambia el comportamiento de lode.
