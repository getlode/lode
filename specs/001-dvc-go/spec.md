# Feature Specification: Versionado de datos de alta velocidad, drop-in compatible con DVC

**Feature Branch**: `001-dvc-go`

**Created**: 2026-06-19

**Status**: Draft

**Input**: User description: "Crear un proyecto open source que genere adopción y permita monetizar, reimplementando/rediseñando en Go una herramienta popular que hoy sufre por el runtime de Python. Tras investigar gaps de mercado se eligió reescribir el núcleo de versionado de datos de DVC (Data Version Control): binario único, hashing paralelo, drop-in compatible con repos DVC existentes."

## Context & Strategic Rationale *(non-normative)*

Esta sección documenta el porqué del proyecto; no es un requisito.

DVC (Data Version Control) es el estándar de facto para versionar datasets y modelos de ML ("Git para datos"). Su CLI sufre lentitud estructural en repos grandes: el recálculo de hashes sobre cientos de miles de archivos es CPU-bound y está limitado por el runtime de Python, y operaciones simples cargan todo el grafo de stages. No existe hoy ninguna alternativa viva con tracción que ataque este dolor (los intentos en otros lenguajes están abandonados o resuelven otro problema).

La oportunidad: una herramienta que opera sobre el **mismo formato de repo DVC existente** (mismos archivos de metadata, mismo layout de cache, mismos nombres de comando) entregando órdenes de magnitud más velocidad en el camino caliente, sin obligar a nadie a migrar. El usuario apunta la herramienta a su repo actual y obtiene la mejora inmediatamente, pudiendo seguir usando DVC de Python sobre el mismo repo. Esto convierte a los 15k+ usuarios existentes en base de adopción inmediata.

## Clarifications

### Session 2026-06-20

- Q: ¿`fetch` (remote→cache sin checkout) es comando del MVP en los requisitos? → A: Sí; se nombra explícitamente en FR-004.
- Q: ¿Alcance de la salida `--json` (FR-025, "comandos de consulta")? → A: Solo `status`; el resto emite progreso/contadores legibles, no datos estructurados.
- Q: ¿Comportamiento al leer metadata DVC con campos desconocidos/más nuevos? → A: Tolerar e ignorar los desconocidos y operar con los conocidos; dvcgo no reescribe `.dvc` ajenos en el MVP.
- Q: ¿Cómo se verifica SC-003 con los cuatro backends S3 (AWS S3, MinIO, R2, B2)? → A: Test automatizado contra MinIO (protocolo S3 idéntico; `endpointurl`+path-style cubre los cuatro) + smoke manual documentado para AWS/R2/B2.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Versionar datasets a alta velocidad en un repo existente (Priority: P1)

Un ingeniero de ML/datos tiene un repositorio que ya usa DVC, con datasets grandes (decenas de GB, cientos de miles de archivos). Hoy registrar un cambio de datos (`add`) o consultar el estado (`status`) tarda minutos por el hashing secuencial. Con esta herramienta, ejecuta el mismo comando apuntando a la misma carpeta y obtiene el resultado en una fracción del tiempo, generando archivos de metadata idénticos a los que produciría DVC.

**Why this priority**: Es el dolor central y la propuesta de valor que diferencia el proyecto ("10x más rápido, sin migrar"). Sin esto no hay razón para adoptar. Es independientemente demostrable y suficiente como MVP mínimo.

**Independent Test**: Tomar un repo DVC real con un dataset grande, correr el comando de tracking con esta herramienta y verificar que (a) produce archivos de metadata byte-compatibles con los de DVC, y (b) completa en una fracción del tiempo medible respecto a DVC de Python.

**Acceptance Scenarios**:

1. **Given** un repositorio con DVC ya inicializado y un directorio de datos nuevo, **When** el usuario trackea ese directorio con la herramienta, **Then** se genera el archivo de metadata correspondiente, los datos se mueven al cache local y la entrada se ignora en el control de versiones de código, de forma idéntica a como lo haría DVC.
2. **Given** un dataset ya trackeado cuyos archivos no cambiaron, **When** el usuario consulta el estado, **Then** la herramienta reporta "sin cambios" sin recalcular hashes innecesariamente y sin escribir nada.
3. **Given** un dataset ya trackeado donde cambió un subconjunto de archivos, **When** el usuario consulta el estado, **Then** la herramienta reporta exactamente qué archivos cambiaron.
4. **Given** un dataset grande de cientos de miles de archivos, **When** el usuario lo trackea, **Then** la operación aprovecha múltiples núcleos y completa significativamente más rápido que DVC de Python sobre el mismo hardware.

---

### User Story 2 - Compartir datos vía almacenamiento remoto S3-compatible (Priority: P1)

El mismo ingeniero necesita subir los datos versionados a un almacenamiento remoto compartido (AWS S3, MinIO, Cloudflare R2 o Backblaze B2) para que su equipo o su pipeline de CI los recupere. Ejecuta `push` para subir y `pull` para bajar, sobre el mismo remote ya configurado en el repo DVC.

**Why this priority**: El versionado local sin sincronización remota no resuelve el caso de uso real de equipos. Push/pull son el segundo eslabón imprescindible del flujo y deben interoperar con el cache que produce/consume DVC.

**Independent Test**: Configurar un remote S3-compatible (p. ej. MinIO local), trackear un dataset, hacer push, borrar el cache local, hacer pull en un clon limpio y verificar que los datos se restauran íntegros y que DVC de Python puede operar sobre el mismo remote.

**Acceptance Scenarios**:

1. **Given** un dataset trackeado y un remote S3-compatible configurado, **When** el usuario hace push, **Then** solo se suben los objetos faltantes (los ya presentes se omiten) y al finalizar el remote contiene todos los objetos referenciados.
2. **Given** un repo recién clonado sin cache local pero con la metadata de datos, **When** el usuario hace pull, **Then** los datos se descargan del remote, se reconstruyen en el workspace y su integridad se verifica contra los hashes esperados.
3. **Given** una transferencia interrumpida a mitad de camino, **When** el usuario reintenta push o pull, **Then** la operación se reanuda sin recorromper ni reduplicar lo ya transferido.
4. **Given** un remote poblado por DVC de Python, **When** el usuario hace pull con esta herramienta, **Then** los datos se recuperan correctamente (interoperabilidad bidireccional del layout de objetos).

---

### User Story 3 - Restaurar y cambiar de versión el workspace (Priority: P2)

Tras cambiar de rama o de commit en el control de versiones de código, los archivos de metadata de datos apuntan a otra versión. El usuario ejecuta `checkout` para que el workspace de datos refleje lo que indican esos archivos de metadata, materializando los archivos desde el cache.

**Why this priority**: Completa el ciclo de versionado (cambiar de versión de datos junto con el código). Importante pero secundario respecto a registrar y compartir; un usuario puede obtener valor con P1 sin esto.

**Independent Test**: Tener dos versiones de un dataset registradas, hacer checkout de cada una y verificar que el workspace queda exactamente con el contenido correspondiente, materializado desde el cache sin volver a descargar si ya está presente.

**Acceptance Scenarios**:

1. **Given** un archivo de metadata que referencia una versión de datos presente en el cache, **When** el usuario hace checkout, **Then** el workspace se materializa con esa versión usando la estrategia de enlace más eficiente disponible (evitando copias innecesarias cuando el sistema de archivos lo permite).
2. **Given** un workspace con datos que no coinciden con la metadata, **When** el usuario hace checkout, **Then** los archivos se actualizan para coincidir y los que ya coinciden no se tocan.
3. **Given** datos referenciados que no están en el cache local, **When** el usuario hace checkout, **Then** la herramienta informa claramente qué falta y sugiere recuperar desde el remote.

---

### User Story 4 - Recuperar espacio eliminando datos no referenciados (Priority: P3)

Con el tiempo el cache local (y el remote) acumula versiones de datos que ya nadie referencia. El usuario ejecuta `gc` para eliminar de forma segura los objetos no alcanzables desde las versiones vigentes.

**Why this priority**: Mantenimiento útil pero no bloqueante para el valor inicial; puede entregarse después del ciclo principal.

**Independent Test**: Crear varias versiones, dejar de referenciar algunas, correr gc y verificar que solo se eliminan los objetos no referenciados y que las versiones vigentes siguen restaurables.

**Acceptance Scenarios**:

1. **Given** un cache con objetos referenciados y no referenciados, **When** el usuario corre gc, **Then** se eliminan únicamente los no referenciados y se reporta el espacio recuperado.
2. **Given** una operación de gc, **When** el usuario la solicita sin confirmación explícita, **Then** la herramienta muestra qué eliminaría y requiere confirmación antes de borrar.

---

### Edge Cases

- **Archivos modificados durante el tracking**: si un archivo cambia mientras se está hasheando/moviendo, la herramienta debe detectar la inconsistencia y fallar de forma segura sin registrar metadata corrupta.
- **Corrupción / hash mismatch**: al recuperar datos (pull/checkout), un objeto cuyo contenido no coincide con su hash esperado debe rechazarse y reportarse, nunca materializarse silenciosamente.
- **Ejecuciones concurrentes**: dos procesos operando sobre el mismo repo/cache no deben corromper el estado; debe existir un mecanismo de bloqueo o detección de concurrencia coherente con el que usa DVC.
- **Transferencias interrumpidas**: push/pull cortados deben poder reanudarse sin dejar objetos a medio escribir visibles como válidos.
- **Credenciales de remote ausentes o inválidas**: error claro y accionable, sin volcar trazas internas.
- **Sistemas de archivos sin soporte de enlaces eficientes** (reflink/hardlink/symlink): checkout debe degradar a copia y seguir funcionando correctamente.
- **Archivos muy grandes** (un único archivo de cientos de GB): el hashing y la transferencia no deben requerir cargar el archivo entero en memoria.
- **Metadata producida por una versión más nueva de DVC con campos desconocidos**: la herramienta MUST tolerar los campos desconocidos —los ignora y opera con los campos conocidos— sin fallar. Dado que en el MVP nunca reescribe un `.dvc` ajeno, no hay pérdida en round-trip; la preservación de campos desconocidos al reescribir queda para una fase futura si se edita metadata existente.
- **Gestión del archivo de exclusión del control de versiones de código** (que los datos no se commiteen como código): debe mantenerse igual que DVC.

## Requirements *(mandatory)*

### Functional Requirements

#### Compatibilidad drop-in
- **FR-001**: La herramienta MUST operar sobre repositorios DVC existentes sin requerir conversión, migración ni reinicialización.
- **FR-002**: La herramienta MUST leer y escribir los archivos de metadata de datos (descriptores de datos versionados, archivo de lock de pipeline cuando exista, configuración de remotes) en un formato compatible byte-a-byte con el que produce y consume DVC, de modo que ambos puedan operar sobre el mismo repo de forma intercambiable.
- **FR-003**: La herramienta MUST usar la misma estructura y direccionamiento por contenido del cache local y del almacenamiento remoto que DVC, garantizando interoperabilidad bidireccional (datos producidos por una herramienta son consumibles por la otra).
- **FR-004**: La herramienta MUST exponer los comandos del MVP con los mismos nombres y semántica básica que DVC: registrar/trackear datos (`add`), consultar estado (`status`), materializar workspace (`checkout`), descargar del remote al cache sin tocar el workspace (`fetch`), subir al remote (`push`), descargar y materializar (`pull` = `fetch` + `checkout`) y recolección de basura (`gc`).
- **FR-005**: La herramienta MUST mantener el archivo de exclusión del control de versiones de código (para que los datos trackeados no se versionen como código) con el mismo comportamiento que DVC.

#### Versionado y estado (P1)
- **FR-006**: Los usuarios MUST poder trackear un archivo o directorio, lo que calcula su hash por contenido, lo traslada al cache y genera el descriptor de metadata correspondiente.
- **FR-007**: La herramienta MUST calcular hashes en paralelo aprovechando múltiples núcleos.
- **FR-008**: La herramienta MUST evitar recalcular el hash de archivos cuyo estado (tamaño/marca temporal u otra señal equivalente a la que usa DVC) indique que no cambiaron.
- **FR-009**: Los usuarios MUST poder consultar el estado y obtener qué datos trackeados cambiaron, se agregaron o faltan, sin modificar el repo.

#### Sincronización remota (P1)
- **FR-010**: La herramienta MUST soportar almacenamiento remoto S3-compatible (AWS S3, MinIO, Cloudflare R2, Backblaze B2 y equivalentes) mediante la configuración de remote ya presente en el repo.
- **FR-011**: Los usuarios MUST poder subir (push) los objetos referenciados al remote, transfiriendo únicamente los que aún no están presentes.
- **FR-012**: Los usuarios MUST poder bajar (pull) los objetos referenciados desde el remote y materializar el workspace.
- **FR-013**: La herramienta MUST verificar la integridad de cada objeto recuperado contra su hash esperado y rechazar los que no coincidan.
- **FR-014**: La herramienta MUST reanudar de forma segura transferencias interrumpidas sin duplicar ni corromper objetos.
- **FR-015**: Las transferencias remotas MUST ejecutarse concurrentemente para aprovechar el ancho de banda disponible.

#### Materialización del workspace (P2)
- **FR-016**: Los usuarios MUST poder materializar (checkout) el workspace para que coincida con los descriptores de metadata vigentes.
- **FR-017**: La herramienta MUST usar la estrategia de enlace más eficiente disponible (reflink/hardlink/symlink según configuración y capacidades del sistema de archivos) y degradar a copia cuando no esté disponible, de forma consistente con DVC.
- **FR-018**: La herramienta MUST dejar intactos los archivos del workspace que ya coinciden con la metadata.

#### Recolección de basura (P3)
- **FR-019**: Los usuarios MUST poder eliminar del cache (y opcionalmente del remote) los objetos no referenciados por las versiones vigentes.
- **FR-020**: La herramienta MUST requerir confirmación explícita (o un flag equivalente) antes de eliminar datos, mostrando previamente el alcance de la eliminación.

#### Seguridad operativa y robustez
- **FR-021**: La herramienta MUST prevenir la corrupción de estado ante ejecuciones concurrentes sobre el mismo repo/cache mediante bloqueo o detección coherente con DVC.
- **FR-022**: La herramienta MUST procesar archivos arbitrariamente grandes mediante streaming, sin cargarlos completos en memoria.
- **FR-023**: La herramienta MUST emitir mensajes de error claros y accionables (credenciales faltantes, objetos faltantes, hash mismatch, permisos) sin volcar trazas internas en uso normal.
- **FR-024**: La herramienta MUST distribuirse como binario único autocontenido, sin requerir un runtime ni dependencias externas instaladas por el usuario.
- **FR-025**: La herramienta MUST ofrecer salida legible por humanos en todos los comandos y, para el comando de consulta `status`, una salida estructurada (`--json`) apta para automatización/CI. Los comandos de transferencia y `gc` reportan progreso/contadores legibles, no `--json`.

### Key Entities

- **Descriptor de datos versionados**: representa un archivo o directorio bajo seguimiento; asocia rutas del workspace con sus hashes por contenido y metadatos (tamaño, número de archivos). Es la unidad de metadata que se versiona junto al código.
- **Objeto de cache**: contenido direccionado por su hash, almacenado una sola vez localmente y/o en el remote; múltiples versiones/archivos idénticos comparten el mismo objeto.
- **Remote**: destino de almacenamiento compartido (S3-compatible) configurado en el repo, donde se sincronizan los objetos de cache.
- **Configuración del repo**: ajustes del proyecto (remotes definidos, estrategia de cache/enlaces) compartidos con DVC.
- **Workspace**: el árbol de archivos materializado con el que trabaja el usuario, reconciliado contra los descriptores y el cache.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En un dataset de referencia de ≥100.000 archivos, registrar cambios (`add`) y consultar estado (`status`) completa al menos 10× más rápido que DVC de Python sobre el mismo hardware.
- **SC-002**: 100% de los repos DVC de prueba pueden ser operados de forma intercambiable: tras una operación de la herramienta, DVC de Python opera el mismo repo sin errores, y viceversa (compatibilidad de metadata y de layout de cache/remote verificada en una batería de casos).
- **SC-003**: Un ciclo completo de versionado (trackear → push → borrar cache → pull en clon limpio → checkout) restaura los datos íntegros, con verificación de integridad al 100% de los objetos. Dado que el protocolo S3 es idéntico entre backends (la diferencia es `endpointurl` + estilo path), la verificación se automatiza contra MinIO y se complementa con un smoke manual documentado contra AWS S3, Cloudflare R2 y Backblaze B2.
- **SC-004**: La instalación se completa con un único artefacto descargable sin dependencias adicionales, y la herramienta queda operativa en menos de 1 minuto desde la descarga en sistemas soportados.
- **SC-005**: En consultas de estado sobre datasets sin cambios, la herramienta no recalcula hashes y responde en tiempo proporcional al número de entradas, no al volumen de datos.
- **SC-006**: Las transferencias remotas saturan de forma demostrable más de un hilo/conexión, logrando throughput agregado significativamente mayor que una transferencia secuencial sobre el mismo enlace.
- **SC-007**: Ante interrupciones forzadas (corte de transferencia, kill del proceso) en cualquier punto, no se produce estado corrupto: una reanudación posterior converge a un resultado íntegro en el 100% de los ensayos.

## Assumptions

- **Lenguaje y forma de distribución**: el proyecto se implementa en un lenguaje compilado a binario único (Go) — decisión de producto central por su ventaja de distribución y concurrencia nativa, no un detalle de implementación negociable. La forma exacta de empaquetado por canal (gestores de paquetes, releases) se define en la fase de planificación.
- **Alcance del MVP**: se limita al núcleo de versionado de datos (trackear, estado, checkout, push, pull, gc). El motor de pipelines reproducibles (definición y ejecución de DAG/repro) queda explícitamente fuera del MVP y se aborda en una fase posterior.
- **Backends remotos del MVP**: cache local + remotes S3-compatible. Google Cloud Storage, Azure Blob, SSH, HDFS y otros quedan fuera del MVP.
- **Base de compatibilidad**: se toma como referencia el formato de repo y el layout de cache/remote de la versión estable vigente de DVC; la compatibilidad se valida contra esa versión.
- **Usuarios objetivo**: ingenieros de ML/datos y de plataforma que ya usan DVC o evalúan adoptar versionado de datos, cómodos con CLI y flujos tipo Git.
- **Monetización (fuera del alcance del spec, contexto de producto)**: el binario y el núcleo son open source; los ingresos previstos provienen de una capa cloud/enterprise posterior (almacenamiento gestionado, linaje, colaboración). La licencia y el modelo se definen antes de la primera publicación pública.
- **Interoperabilidad como invariante**: en caso de conflicto entre "mejorar el diseño" y "mantener compatibilidad drop-in con DVC", durante el MVP prevalece la compatibilidad.
