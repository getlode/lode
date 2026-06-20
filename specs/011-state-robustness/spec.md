# Feature Specification: Robustez y corrección del state DB

**Feature Branch**: `011-state-robustness`

**Created**: 2026-06-20

**Status**: Draft

**Input**: Hallazgo de la auditoría (greybeard + MLOps): el state DB que evita rehashear usa la tupla `(inode, mtime, size)` para decidir si un archivo cambió. En filesystems o situaciones donde esa señal miente —NFS, inodos reciclados, restauración de backups que resetea mtimes, relojes desincronizados— la herramienta podría dar un falso "up to date" y NO detectar un cambio real. Para datos, un falso "sin cambios" es peor que ser lento. Además, cuando lode y DVC operan el mismo repo, el state de lode podría quedar desactualizado respecto a lo que hizo DVC.

## Context & Strategic Rationale *(non-normative)*

El state DB es la mayor optimización del producto (no rehashear lo intacto), pero su corrección descansa en una heurística de metadata. La pérdida silenciosa de un cambio (falso "up to date") socava la confianza en una herramienta de versionado de datos — el peor fallo posible. Endurecer la heurística, ofrecer una vía a prueba de fallos (rehash completo), y manejar la coexistencia con DVC, protege la propiedad más importante: nunca reportar "sin cambios" cuando los hubo.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Nunca un falso "sin cambios" (Priority: P1)

Un usuario en un filesystem donde la metadata es poco confiable (NFS, inodos reciclados, mtimes restaurados) obtiene siempre una detección de cambios correcta: si el contenido cambió, la herramienta lo nota; nunca reporta "up to date" sobre datos que en realidad cambiaron.

**Why this priority**: Un falso negativo en detección de cambios = pérdida silenciosa de una versión de datos. Es el fallo más grave para una herramienta de versionado.

**Independent Test**: Simular las condiciones problemáticas (mismo `(inode, mtime, size)` con contenido distinto; mtime reseteado por "restore"; inodo reciclado) y verificar que la herramienta detecta el cambio de contenido (no se fía ciegamente de la metadata) o, como mínimo, ofrece una vía garantizada para hacerlo.

**Acceptance Scenarios**:

1. **Given** un archivo cuyo contenido cambió pero cuya tupla de metadata coincide con la cacheada, **When** el usuario consulta el estado, **Then** la herramienta NO reporta falsamente "sin cambios" para ese archivo (lo detecta, o el modo a prueba de fallos lo detecta).
2. **Given** un entorno conocido por metadata poco confiable, **When** el usuario lo configura/declara, **Then** la herramienta usa una estrategia de detección más estricta o el rehash completo.
3. **Given** cualquier duda sobre la validez del state, **When** el usuario lo solicita, **Then** existe una vía explícita para forzar el rehash completo e ignorar el cache de estado.

---

### User Story 2 - Coexistencia consistente con DVC (Priority: P2)

Cuando lode y DVC operan el mismo repo a lo largo del tiempo, lode no se fía de un state propio que podría estar desactualizado respecto a lo que hizo DVC; el resultado de `status` sigue siendo correcto.

**Why this priority**: El producto vende coexistencia ("corré cualquiera de los dos"); si el state de lode queda stale respecto a DVC y produce un `status` engañoso, la promesa se rompe.

**Independent Test**: Operar el repo alternando lode y DVC, y verificar que el `status` de lode refleja el estado real tras las operaciones de DVC, sin falsos "up to date".

**Acceptance Scenarios**:

1. **Given** que DVC modificó el repo/datos después de lode, **When** el usuario corre `status` con lode, **Then** el resultado refleja el estado real (no un "up to date" basado en un state propio obsoleto).

---

### Edge Cases

- **Inodo reciclado**: un archivo nuevo reusa el inodo de uno borrado con misma `mtime`/`size` — no debe confundirse con el anterior.
- **mtime con baja resolución o reseteado**: backups/restore que normalizan mtimes no deben ocultar cambios.
- **Reloj del sistema hacia atrás**: cambios con mtime menor al cacheado deben detectarse igual.
- **Costo del modo estricto**: la estrategia a prueba de fallos (rehash) es más lenta; debe ser opcional/declarable para no perder la ventaja de performance en el caso común.
- **Corrupción del propio state DB**: debe degradar a rehash, nunca dar un resultado incorrecto.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: La herramienta MUST NOT reportar "sin cambios" para un archivo cuyo contenido cambió; ante señales de metadata poco confiable, debe favorecer la detección correcta por sobre la velocidad.
- **FR-002**: La herramienta MUST ofrecer una vía explícita para forzar el rehash completo, ignorando el cache de estado.
- **FR-003**: La herramienta MUST permitir declarar/operar en un modo de detección más estricto para entornos de metadata poco confiable (NFS, restore, etc.).
- **FR-004**: Ante un state DB corrupto o ilegible, la herramienta MUST degradar al rehash, nunca producir un resultado de estado incorrecto.
- **FR-005**: La herramienta MUST mantener un `status` correcto cuando el repo fue modificado por DVC entre operaciones de lode (no fiarse de un state propio obsoleto que produzca falsos "up to date").
- **FR-006**: La documentación MUST explicar las garantías y límites de la detección por metadata y cuándo usar el modo estricto/rehash.

### Key Entities

- **Entrada de state**: la señal cacheada por archivo (incluida la tupla de metadata) y su validez.
- **Modo de detección**: rápido (heurística de metadata) vs estricto (rehash), y la política de cuándo aplicar cada uno.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el 100% de los casos sembrados de "metadata que miente" (mismo `(inode, mtime, size)` con contenido distinto; mtime reseteado; inodo reciclado), la herramienta NO produce un falso "sin cambios" (lo detecta o el modo estricto lo detecta).
- **SC-002**: Existe y funciona una vía para forzar el rehash completo, verificable.
- **SC-003**: Alternando lode y DVC sobre el mismo repo, el `status` de lode refleja el estado real tras las operaciones de DVC.
- **SC-004**: Un state DB corrupto degrada a rehash sin producir un resultado incorrecto.

## Assumptions

- **Caso común sin regresión**: el endurecimiento no debe degradar la performance del caso común (metadata confiable); el modo estricto es opcional/declarable.
- **Reutiliza** el state DB existente; agrega validación/robustez y modos, sin cambiar el formato de datos del repo.
- **Detección de contenido**: la garantía última de corrección es el rehash; la heurística de metadata es la optimización que debe poder anularse.
- **Coexistencia con DVC**: el alcance es que lode no dé resultados incorrectos; no se requiere compartir el state interno con DVC (cada herramienta mantiene el suyo).
