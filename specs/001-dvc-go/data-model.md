# Data Model: Núcleo de versionado de datos (DVC en Go)

**Feature**: 001-dvc-go | **Date**: 2026-06-19 | **Phase**: 1

Entidades del dominio y sus invariantes de compatibilidad. No es esquema de base de datos; describe las estructuras lógicas que se serializan a archivos compatibles con DVC y el state interno.

---

## DvcFile (descriptor `.dvc`)

Representa un archivo `*.dvc` commiteable. Una salida por entrada de `outs` (en el MVP, normalmente una).

| Campo | Tipo | Notas |
|---|---|---|
| `outs` | `[]Out` | Lista de salidas trackeadas |

Serialización: YAML 3.x, claves en orden de inserción, 2 espacios de indent, **un newline final**. Ver research §1.

## Out (salida dentro de un `.dvc`)

| Campo | Tipo | Reglas |
|---|---|---|
| `md5` | string | Hash del contenido (32 hex) o del dir (`<32hex>.dir`) |
| `size` | int64 | Tamaño en bytes; para dir = suma de archivos |
| `nfiles` | int64 | **Solo** presente si es directorio |
| `hash` | string | Literal `md5` (3.x); ausente en outputs legacy |
| `path` | string | Ruta relativa al `.dvc`, separador del SO al leer / POSIX al normalizar |

Orden de emisión: `md5`, `size`, (`nfiles`), `hash`, `path`.

**Validación**: `md5` que termina en `.dir` ⇒ requiere `nfiles`. `path` no vacío. `hash == "md5"` para outputs nuevos.

## DirManifest (objeto `.dir`)

Contenido del objeto de directorio en el cache/remote. Es la entidad más sensible a bytes.

| Campo | Tipo | Reglas |
|---|---|---|
| `entries` | `[]DirEntry` | Ordenadas ascendente por `relpath` |

`DirEntry`:
| Campo | Tipo | Reglas |
|---|---|---|
| `md5` | string | Hash del archivo contenido (32 hex) |
| `relpath` | string | Ruta POSIX relativa a la raíz del dir trackeado |

**Serialización exacta** (research §2): JSON con separadores `", "` / `": "`, claves alfabéticas (`md5` antes de `relpath`), `ensure_ascii` (escape `\uXXXX` para >0x7F), sin newline final. El oid del manifest = `md5(bytes) + ".dir"`.

## CacheObject

Contenido direccionado por hash, almacenado una vez.

| Campo | Tipo | Reglas |
|---|---|---|
| `oid` | string | 32 hex (archivo) o `<32hex>.dir` (manifest) |
| `path` | string | `<store>/files/md5/<oid[:2]>/<oid[2:]>` |
| `mode` | filemode | `0o444` (read-only) en cache local |

**Estados**: ausente → escribiendo (tmp) → presente (rename atómico + chmod 0o444). Nunca visible a medio escribir bajo su path final.

## Remote

| Campo | Tipo | Reglas |
|---|---|---|
| `name` | string | Nombre lógico (sección INI `['remote "name"']`) |
| `url` | string | `s3://<bucket>/<key-prefix>` |
| `endpointurl` | string | Opcional; requerido para MinIO/R2/B2/Spaces |
| `region` | string | Opcional |
| `credenciales` | — | Vía env/perfil AWS o `access_key_id`/`secret_access_key` |
| `sse`, `sse_kms_key_id`, `acl` | string | Opcionales |

Mapeo de objeto: `<url>/files/md5/<oid[:2]>/<oid[2:]>`. Acceso path-style (`UsePathStyle=true`).

## RepoConfig (`.dvc/config`, INI)

| Sección/clave | Notas |
|---|---|
| `[core] remote` | Remote por defecto |
| `[core] hardlink_lock` | Selecciona backend de lock (flock vs hardlink) |
| `[cache] type` | `reflink,copy` default; lista ordenada de estrategias |
| `['remote "name"']` | Campos de Remote (arriba) |

## StateEntry (state interno propio — no compartido con DVC-Python)

Optimización para no rehashear. Almacenado en `.dvc/tmp/dvcgo/state.db` (bbolt).

| Campo | Tipo | Reglas |
|---|---|---|
| `key` | string | Ruta absoluta del archivo |
| `ino` | uint64 | `st_ino` |
| `mtime` | int64 | nanosegundos |
| `size` | int64 | bytes |
| `md5` | string | Hash cacheado |

**Regla de invalidación**: si `(ino, mtime, size)` actuales == almacenados ⇒ devolver `md5` sin leer el archivo; si difiere ⇒ rehashear y actualizar.

## Locks (compartidos con DVC-Python)

- **GlobalLock**: `.dvc/tmp/lock`, exclusivo, `flock(LOCK_EX)`. Timeout 3s, reintentos.
- **RWLock**: `.dvc/tmp/rwlock` (JSON) + `.dvc/tmp/rwlock.lock`. `{read:{path:[{pid,cmd}]}, write:{path:{pid,cmd}}}`. Write bloquea ante cualquier reader/writer solapado; read bloquea solo ante writer. PIDs muertos se purgan.

---

## Relaciones

```
DvcFile 1───* Out
Out (dir) 1───1 DirManifest (.dir CacheObject) 1───* DirEntry
DirEntry *───1 CacheObject (archivo)
CacheObject *───* Remote (sincronización push/pull)
Workspace file *───1 CacheObject (vía link/copy en checkout)
StateEntry 1───1 Workspace file (cache de hash)
```

## Transiciones de estado de un Out

```
untracked ──add──> tracked(cached, gitignored)
tracked ──push──> tracked(remote-synced)
remote-synced ──(borrar cache)──> tracked(metadata-only)
metadata-only ──pull/fetch──> tracked(cached) ──checkout──> materialized(workspace)
materialized ──(editar archivo)──> modified (detectado por status vía state/hash)
cached-unreferenced ──gc──> removed
```
