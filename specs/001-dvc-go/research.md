# Research: Núcleo de versionado de datos en Go, drop-in compatible con DVC

**Feature**: 001-dvc-go | **Date**: 2026-06-19 | **Phase**: 0

Todo verificado contra código fuente de `iterative/dvc`, `iterative/dvc-data`, `iterative/dvc-objects`, `iterative/scmrepo` (rama `main`, vigente a 2026) y doc oficial `dvc.org/doc`. Target de compatibilidad: **DVC 3.x** (formato `hash: md5`, cache en `files/md5/`).

---

## 1. Formato del archivo `.dvc`

**Decision**: Replicar el formato YAML 3.x con orden de claves determinístico por inserción (NO alfabético).

**Rationale**: El `.dvc` se commitea a Git; cualquier diferencia de bytes rompe la interoperabilidad y ensucia diffs. DVC no usa `sort_keys` en el `.dvc`; el orden lo fija `Output.dumpd()`.

Archivo individual generado por `add archivo.csv`:
```yaml
outs:
- md5: a304afb96060aad90176268345e10355
  size: 47482
  hash: md5
  path: data.csv
```
Directorio generado por `add dir/`:
```yaml
outs:
- md5: f437247ec66d73ba66b0ade0246fcd49.dir
  size: 32153
  nfiles: 2
  hash: md5
  path: dir
```
Invariantes: indentación 2 espacios, item de lista sin indentar respecto a `outs`, **un único newline final**. `nfiles` solo para directorios. `size` de dir = suma de tamaños. Orden efectivo de claves: `md5`, `size`, (`nfiles`), `hash`, `path`.

**Alternatives considered**: Formato propio más rico (rechazado en spec: rompe drop-in). Usar un YAML lib genérico de Go con sort de claves (rechazado: produce orden alfabético incompatible).

---

## 2. Hashing de archivos y del objeto `.dir` (CRÍTICO)

**Decision**: MD5 binario puro sobre contenido (chunks de 1 MiB). Para directorios, emitir el JSON del `.dir` con un **serializador propio** que reproduzca exactamente la salida de `json.dumps(sort_keys=True)` de Python.

**Rationale**: El hash del directorio se deriva del JSON del `.dir`; un solo byte distinto cambia el hash y rompe todo. `encoding/json` de Go difiere de Python en separadores y escape Unicode.

Reglas byte-a-byte del objeto `.dir`:
- Cada entry: `{"md5": "<hash32>", "relpath": "<ruta-posix>"}`.
- Lista ordenada ascendente por `relpath` (byte order).
- Claves alfabéticas dentro del objeto (`md5` antes de `relpath`).
- **Separadores con espacio**: `", "` entre items y `": "` entre clave/valor (default de Python sin `separators=`). `encoding/json` NO los pone → hay que generarlo manual.
- **`ensure_ascii=True`**: todo char > 0x7F escapado a `\uXXXX`.
- `relpath` siempre POSIX (`/`), incluso en Windows.
- **Sin newline final** en el contenido del `.dir`.
- El hash del dir = MD5 de esos bytes UTF-8, **con `.dir` apendado** al hex digest (resultado de 36 chars: 32 hex + `.dir`).

Ejemplo de contenido `.dir` (bytes exactos):
```json
[{"md5": "de7371b0119f4f75f9de703c7c3bac16", "relpath": "cat.jpeg"}, {"md5": "402e97968614f583ece3b35555971f64", "relpath": "index.jpeg"}]
```

El hash del contenido individual es solo del binario; el nombre no entra. Symlinks dentro de un dir trackeado: se sigue el contenido del destino.

**Alternatives considered**: `json.Marshal` directo (rechazado: sin espacios ni escape ASCII). Hash distinto a MD5 (rechazado: rompe compat y direccionamiento del cache).

**Riesgo #1 del proyecto**: la serialización del `.dir`. Mitigación: test de oráculo byte-a-byte contra `.dir` generados por DVC real, antes de avanzar (ver quickstart).

---

## 3. Layout de cache y remote

**Decision**: Content-addressed `<prefix>/files/md5/<oid[:2]>/<oid[2:]>`, mismo esquema en cache local y remote. Objetos del cache en modo `0o444`.

**Rationale**: DVC 3.x usa este layout idéntico en local y remoto; el remote es otro store con el mismo `oid_to_path`. Reusar el esquema garantiza interoperabilidad bidireccional.

- Cache local: `.dvc/cache/files/md5/ec/1d2935...`
- Remote S3: `s3://<bucket>/<key-prefix>/files/md5/ec/1d2935...`
- Objeto `.dir`: mismo store, sufijo `.dir` en el oid.
- Legacy 2.x (solo lectura, opcional): `.dvc/cache/<oid[:2]>/<oid[2:]>` con hash `md5-dos2unix`. **Fuera del MVP** salvo lectura básica si aparece.

**Alternatives considered**: Layout propio (rechazado: rompe interoperabilidad de remote y cache).

---

## 4. Push / pull / checkout / status

**Decision**: Replicar la maquinaria de `transfer`: status del remote con dos estrategias (HEAD-por-objeto vs LIST masivo por prefijo) + orden ".dir solo después de sus contenidos" + escritura atómica tmp+rename.

**Rationale**: Es el comportamiento correcto y robusto de DVC; respetar el orden evita remotes con `.dir` colgando (bug histórico #4343).

- **status remoto** (`oids_exist`): heurística con constantes `TRAVERSE_PREFIX_LEN=2`, `TRAVERSE_THRESHOLD_SIZE=500000`, `TRAVERSE_WEIGHT_MULTIPLIER=5`, `LIST_OBJECT_PAGE_SIZE=1000`. Muchos oids → un LIST masivo (prefijos `00`..`ff` en paralelo); pocos → `HEAD`/exists por objeto en lotes.
- **push**: dirs antes que archivos; por cada dir, subir contenidos y solo si todos OK, subir el `.dir`. Concurrencia `jobs` (default ~4×CPU para S3).
- **fetch**: remote→cache (solo faltantes). **checkout**: cache→workspace (link/copy). **pull** = fetch+checkout.
- **Verificación**: cache local `verify=False` (confía en el path). Remotes/`verify=True`: re-hashear al bajar; si no matchea, **borrar objeto y fallar** (`ObjectFormatError`).
- **Atomicidad**: descarga a tmp + `os.Rename` atómico en el mismo FS; destino existente → skip idempotente.

**Alternatives considered**: Siempre HEAD por objeto (rechazado: O(n) requests en remotes grandes). Subir `.dir` primero (rechazado: deja remote inconsistente ante fallo parcial).

---

## 5. State DB (evitar rehashear)

**Decision**: State propio en Go (no leer la diskcache de Python) con clave `(st_ino, st_mtime, st_size) → md5`. Motor: **bbolt** (KV puro Go) o `modernc.org/sqlite`.

**Rationale**: DVC usa `diskcache` (SQLite+pickle p4), costoso de interoperar desde Go y no requerido por la spec (solo se comparte el repo, no el state interno de `.dvc/tmp`). Mantener state propio sobre la misma tupla da la misma ganancia (mayor optimización real de DVC) sin acoplarse al formato pickle.

- Detección de cambio: si los 3 (ino, mtime, size) coinciden, devolver md5 cacheado sin leer el archivo.
- Ubicación propia: dentro de `.dvc/tmp/` con un nombre que no colisione con los de DVC-Python (p. ej. `.dvc/tmp/lode/state.db`).

**Alternatives considered**: Leer/escribir la diskcache de Python (rechazado: acoplamiento a pickle p4 y a internals inestables). modernc.org/sqlite si se quiere SQL/inspección; bbolt si se quiere el mínimo KV. **Elección por defecto: bbolt** por simplicidad y cero superficie SQL.

---

## 6. Locking (coexistencia con DVC-Python)

**Decision**: Tomar el lock global `.dvc/tmp/lock` vía `flock(LOCK_EX)` (compatible con `zc.lockfile`) y respetar el `rwlock` JSON de DVC para concurrencia fina.

**Rationale**: Para coexistir sobre el mismo repo sin corromper estado hay que honrar el mismo lock que DVC.

- Lock global exclusivo: `.dvc/tmp/lock` (backend `zc.lockfile`/flock por default; `flufl.lock`/hardlink si `core.hardlink_lock=true`).
- rwlock fino: `.dvc/tmp/rwlock` (+ `.lock` para editar el JSON), estructura `{read:{path:[{pid,cmd}]}, write:{path:{pid,cmd}}}`, con purga de PIDs muertos.
- MVP: honrar al menos el lock global exclusivo (flock); implementar rwlock para paridad completa.

**Alternatives considered**: Lock propio independiente (rechazado: no impide colisión con un `dvc` de Python corriendo en paralelo).

---

## 7. Materialización (link strategies) y `.gitignore`

**Decision**: `cache.type` default `reflink,copy` con fallback; objetos linkeados quedan `0o444`. `.gitignore` por directorio con entrada `"/" + relpath` POSIX, idempotente.

**Rationale**: Igualar el comportamiento de DVC para no romper el workspace ni la protección del cache.

- Orden: reflink → (hardlink/symlink si configurado) → copy. Reflink Linux vía `unix.IoctlFileClone` (FICLONE = `0x40049409`); macOS `clonefile`; fallback `copy_file_range` → `io.Copy`. Detectar `EOPNOTSUPP`/`EXDEV`/`EINVAL` para degradar.
- Relink: hardlink si inodo difiere; symlink si target difiere; copy/reflink no relinkea.
- `.gitignore`: `<dir>/.gitignore`, línea `/<relpath>`, append sin duplicar, preservando newline previo.

**Alternatives considered**: Solo copy en MVP (descartado parcialmente: reflink es barato con `x/sys/unix` y es el default de DVC; se implementa reflink+copy, hardlink/symlink honrando config).

---

## 8. Stack Go

**Decision**: cobra (CLI) · aws-sdk-go-v2 + `feature/s3/transfermanager` (GA 2026) con `BaseEndpoint`+`UsePathStyle` · errgroup `SetLimit(NumCPU)` + `sync.Pool` + `crypto/md5` · bbolt (state) · tmp+`os.Rename` + `gofrs/flock` · GoReleaser con `CGO_ENABLED=0` · `testing` + testcontainers-go/minio.

**Rationale**: Hilo conductor = **cero cgo** en toda la cadena, lo que habilita el binario único cross-compile (requisito de producto). `BaseEndpoint`+`UsePathStyle` cubre AWS S3 + MinIO + R2 + B2 con un solo SDK. errgroup acota el hashing CPU-bound a NumCPU sin GC churn.

| Área | Elección | Razón |
|---|---|---|
| CLI | `spf13/cobra` | Estándar, completions, paridad UX con DVC |
| S3 | `aws-sdk-go-v2/service/s3` + `feature/s3/transfermanager` | `BaseEndpoint`+`UsePathStyle`; multipart concurrente GA |
| Hashing | `errgroup.SetLimit(NumCPU)` + `sync.Pool` + `crypto/md5` | CPU-bound acotado, MD5 = compat |
| State DB | `go.etcd.io/bbolt` | KV puro Go, sin cgo, mínima superficie |
| Atomicidad/lock | tmp+`os.Rename`, `gofrs/flock` | Rename atómico; flock cross-platform compatible con DVC |
| Reflink | `golang.org/x/sys/unix` (`IoctlFileClone`) | CoW nativo, mismo FICLONE que DVC |
| Distribución | GoReleaser `CGO_ENABLED=0` | Matriz multi-OS trivial sin cgo |
| Testing | `testing` + `testcontainers-go/modules/minio` | Table tests + S3-compatible real |

**Alternatives considered**: kong (buena, menos ecosistema/completions); minio-go (más simple para S3-compat pero un SDK extra; se deja como posible adaptador futuro); mattn/go-sqlite3 (rechazado: cgo rompe cross-compile); modernc.org/sqlite (válido si se quiere SQL, pero bbolt es más simple para KV puro).
