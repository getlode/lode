# Feature Specification: Bootstrap standalone y onboarding (`init` + `doctor`)

**Feature Branch**: `002-init-onboarding`

**Created**: 2026-06-20

**Status**: Draft

**Input**: User description: "Bootstrap standalone y onboarding de lode: un usuario nuevo con datos pero sin repo DVC no puede empezar porque no existe `lode init` y `lode add` falla con 'not a DVC repository', obligándolo a instalar DVC-Python. Agregar `lode init` (estructura byte-compatible con `dvc init`, con modo sin git), errores accionables que sugieran `lode init`, y `lode doctor` para diagnosticar el repo. Objetivo: primer minuto sin fricción y 100% sin Python, manteniendo compatibilidad byte-a-byte con DVC."

## Context & Strategic Rationale *(non-normative)*

El dogfooding expuso un bloqueante de adopción: hoy lode solo opera sobre repos que ya tienen `.dvc/` (creado por DVC-Python). Un usuario nuevo con datos hace `lode add` y recibe `Error: not a DVC repository`, sin pista de qué hacer. Para empezar necesita instalar DVC-Python y correr `dvc init` — lo que **rompe la propuesta de valor central** ("binario único, sin Python") justo en el primer minuto, que es donde se gana o se pierde la adopción.

Este feature cierra ese hueco: que un usuario vaya de un directorio vacío a datos trackeados y pusheados usando **solo el binario lode**, sin Python ni DVC, y que el repo resultante siga siendo operable por DVC de forma intercambiable.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Empezar de cero sin Python (Priority: P1)

Una persona con un dataset y sin DVC ni Python instalados quiere versionar sus datos. Ejecuta `lode init` en su carpeta, luego `lode add` y `lode push`, y todo funciona usando solo el binario. Más tarde, un compañero que sí usa DVC clona el repo y lo opera sin problemas.

**Why this priority**: Es el bloqueante de adopción. Sin esto, la promesa "sin Python" es falsa para usuarios nuevos y el primer minuto termina en un error sin salida. Es lo único que convierte a lode en una herramienta autónoma.

**Independent Test**: En un entorno sin DVC/Python, en un directorio vacío con datos: `lode init` → `lode add` → `lode push`. Verificar que el flujo completa sin Python y que un DVC real puede operar el mismo repo después.

**Acceptance Scenarios**:

1. **Given** un directorio sin `.dvc/` y sin DVC/Python instalados, **When** el usuario corre `lode init`, **Then** se crea la estructura de repositorio que DVC reconoce (config del proyecto, el archivo de ignore de datos en la raíz, y los directorios de trabajo), idéntica a la que produciría `dvc init` en el modo equivalente.
2. **Given** un repo recién inicializado con `lode init`, **When** el usuario corre `add`/`push`/`pull`/`checkout`/`gc`, **Then** funcionan sin requerir Python ni DVC.
3. **Given** un repo creado por `lode init`, **When** un usuario con DVC-Python opera ese repo, **Then** DVC lo reconoce y opera sin errores (y viceversa).
4. **Given** un directorio que ya tiene un repo DVC (creado por DVC), **When** el usuario corre `lode init`, **Then** lode lo detecta y no lo daña ni sobrescribe la configuración existente.

---

### User Story 2 - Errores que guían (Priority: P2)

Un usuario corre un comando de lode en una carpeta que no es un repo (o donde se equivocó de directorio). En vez de un error opaco, recibe un mensaje que le dice exactamente qué hacer a continuación.

**Why this priority**: Reduce el abandono en los primeros minutos. Un error accionable convierte un callejón sin salida en el siguiente paso obvio.

**Independent Test**: Correr cada comando que requiere repo fuera de un repo y verificar que el mensaje nombra el comando concreto a ejecutar (`lode init`) y, cuando aplica, la opción para apuntar a otro directorio.

**Acceptance Scenarios**:

1. **Given** un directorio sin `.dvc/`, **When** el usuario corre un comando que requiere repo (p. ej. `add`), **Then** el error explica que no hay repo y sugiere correr `lode init` (y menciona cómo apuntar a otro directorio si se equivocó de ubicación).
2. **Given** un repo sin remote configurado, **When** el usuario corre `push`, **Then** el error indica que falta configurar un remote y muestra el comando para hacerlo.

---

### User Story 3 - Diagnóstico con `lode doctor` (Priority: P2)

Cuando algo no funciona (no sube al remote, el cache no escribe, el repo parece de otra versión), el usuario corre `lode doctor` y obtiene un reporte claro de qué está bien, qué está mal, y cómo arreglarlo.

**Why this priority**: Acorta el tiempo de resolución de problemas y reduce la carga de soporte, clave para que la gente persista con la herramienta en lugar de abandonarla ante el primer obstáculo.

**Independent Test**: Sembrar cada clase de problema (sin repo, sin remote, remote inalcanzable, cache no escribible, formato legacy) y verificar que `lode doctor` lo identifica y propone una corrección, con código de salida apropiado.

**Acceptance Scenarios**:

1. **Given** un repo sano con remote alcanzable, **When** el usuario corre `lode doctor`, **Then** reporta todo OK y termina con código de éxito.
2. **Given** un repo con un problema (sin remote / remote inalcanzable / cache no escribible / formato legacy 2.x / sin repo), **When** el usuario corre `lode doctor`, **Then** identifica cada problema, sugiere una corrección concreta, y termina con código distinto de cero si hay un problema bloqueante.

---

### Edge Cases

- **Ya inicializado por lode o por DVC**: `lode init` detecta un `.dvc/` existente y no lo sobrescribe; informa que ya está inicializado.
- **Con git vs sin git**: `lode init` soporta el modo sin control de versiones de código (equivalente a `dvc init --no-scm`) y el modo con git; el contenido de la configuración refleja el modo elegido tal como lo haría DVC.
- **Init dentro de un subdirectorio de un repo existente**: se detecta el repo padre y se informa, en vez de crear un repo anidado por error.
- **`.dvcignore` o config preexistentes**: no se pisan; si ya existen, se respetan.
- **Cache en otro sistema de archivos o sin permisos de escritura**: `lode doctor` lo detecta y lo reporta como problema con sugerencia.
- **Credenciales de remote ausentes/inválidas**: `lode doctor` distingue "remote no configurado" de "remote configurado pero inalcanzable/credenciales inválidas".
- **Repo en formato legacy 2.x**: `lode doctor` lo identifica y explica las implicancias (lectura soportada) con la sugerencia correspondiente.

## Requirements *(mandatory)*

### Functional Requirements

#### Bootstrap (`init`)
- **FR-001**: `lode init` MUST crear la estructura de repositorio que DVC reconoce y opera, byte-compatible con la que produce `dvc init` en el modo equivalente (configuración del proyecto, archivo de ignore de datos en la raíz, directorios de trabajo necesarios).
- **FR-002**: `lode init` MUST soportar un modo sin control de versiones de código (equivalente a `dvc init --no-scm`) y un modo para repos con git, generando la configuración correspondiente a cada modo de forma compatible con DVC.
- **FR-003**: Tras `lode init`, el usuario MUST poder ejecutar el ciclo completo de versionado (add, status, checkout, push, fetch, pull, gc) sin requerir Python ni DVC instalados.
- **FR-004**: El repositorio resultante de `lode init` MUST permanecer operable por DVC-Python de forma intercambiable (invariante de compatibilidad byte-a-byte, constitución v1.0.0 Principio I).
- **FR-005**: `lode init` MUST ser seguro ante repos ya inicializados: detecta un `.dvc/` existente (creado por lode o por DVC) y no lo daña ni sobrescribe la configuración; informa el estado de forma clara.
- **FR-006**: `lode init` MUST detectar cuando se ejecuta dentro de un subdirectorio de un repo existente e informarlo en vez de crear un repo anidado.

#### Errores que guían
- **FR-007**: Todo comando que requiere un repo MUST, cuando no se encuentra ninguno, emitir un error accionable que nombre el comando exacto a ejecutar (`lode init`) e indique cómo apuntar a otro directorio.
- **FR-008**: Los errores de precondición frecuentes (sin remote configurado, objeto faltante en cache) MUST incluir la acción concreta sugerida (p. ej. el comando para configurar un remote, o `lode pull`).

#### Diagnóstico (`doctor`)
- **FR-009**: `lode doctor` MUST reportar el estado de: presencia y validez del `.dvc/`, ubicación del cache y si es escribible, remote(s) configurados y su alcanzabilidad, versión de formato detectada (legacy 2.x vs 3.x), y si la coexistencia con DVC es segura.
- **FR-010**: Por cada problema detectado, `lode doctor` MUST incluir una sugerencia de corrección concreta.
- **FR-011**: `lode doctor` MUST distinguir "remote no configurado" de "remote configurado pero inalcanzable/credenciales inválidas".
- **FR-012**: `lode doctor` MUST terminar con código de salida cero si el repo está sano y distinto de cero si hay un problema bloqueante, para uso en scripts/CI.

#### Onboarding
- **FR-013**: La salida de ayuda y el primer-run MUST hacer descubrible el camino de cero-a-primer-track (init → add).

### Key Entities

- **Estructura de repositorio**: el conjunto de archivos/directorios que definen un repo de datos (configuración del proyecto, archivo de ignore de datos, directorios de cache y de trabajo) que tanto lode como DVC reconocen.
- **Modo de inicialización**: con git vs sin git, que determina detalles de la configuración generada.
- **Reporte de diagnóstico**: el resultado estructurado de `lode doctor` — una lista de chequeos con estado (OK/problema), detalle y sugerencia.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En un entorno sin Python ni DVC, un usuario nuevo va de un directorio vacío con datos a "trackeado y pusheado" usando solo el binario lode, en no más de 3 comandos (`init`, `add`, `push`).
- **SC-002**: El 100% de los repos creados por `lode init` (en ambos modos) son operados por DVC-Python sin errores, y los creados por `dvc init` son operados por lode — verificado en una batería de casos.
- **SC-003**: Al ejecutar un comando que requiere repo fuera de un repo, el mensaje de error nombra el comando exacto a ejecutar a continuación en el 100% de los casos.
- **SC-004**: `lode doctor` identifica correctamente y propone una corrección para cada clase de problema definida (sin repo, sin remote, remote inalcanzable, cache no escribible, formato legacy) en el 100% de los casos sembrados, con el código de salida correcto.
- **SC-005**: Ningún camino de onboarding (init, primer add, primer push, diagnóstico) requiere Python ni DVC instalados.
- **SC-006**: La estructura generada por `lode init` es byte-idéntica a la de `dvc init` en el modo equivalente (verificado por comparación de bytes).

## Assumptions

- **Referencia de compatibilidad**: se toma como referencia el formato de `dvc init` de la versión estable vigente de DVC 3.x. El modo por defecto y los detalles de configuración (p. ej. el flag de "sin git") replican el comportamiento de DVC.
- **Alcance de `doctor`**: diagnóstico de estado del repo y del remote; no incluye reparación automática (más allá de sugerir comandos) en esta iteración.
- **Alcanzabilidad del remote**: "alcanzable" significa que se puede establecer conexión y autenticar contra el remote configurado; no implica verificar la integridad de todos los objetos.
- **Compatibilidad como invariante**: ante conflicto entre "mejorar el diseño de init" y "ser byte-compatible con dvc init", prevalece la compatibilidad.
- **Idioma de la interfaz**: la salida de usuario sigue la decisión global del proyecto (migración a inglés tratada en un feature aparte); este spec no fija el idioma, solo el contenido accionable de los mensajes.
