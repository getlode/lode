# Quickstart & Validation Guide

**Feature**: 001-dvc-go | **Phase**: 1

Escenarios ejecutables que prueban el feature end-to-end. El binario se asume llamado `dvcgo`. Detalles de comandos en [contracts/cli.md](contracts/cli.md); estructuras en [data-model.md](data-model.md).

## Prerrequisitos

- Go 1.23+ (build), `CGO_ENABLED=0`.
- Docker (para tests de integración con MinIO).
- DVC de Python instalado (`pip install dvc[s3]`) **solo para los tests de compatibilidad/oráculo**.

## Build

```bash
go build -o dvcgo ./cmd/dvcgo
```

---

## Escenario 1 — Oráculo de compatibilidad de bytes (P1, gate del proyecto)

Prueba el riesgo #1: que el `.dvc` y el objeto `.dir` sean byte-idénticos a los de DVC.

```bash
# Preparar un dir con datos deterministas
mkdir -p oracle/data && printf 'a' > oracle/data/cat.jpeg && printf 'b' > oracle/data/index.jpeg
cd oracle && dvc init --no-scm -q

# Referencia: DVC de Python
dvc add data
cp data.dvc /tmp/ref.dvc
find .dvc/cache/files/md5 -name '*.dir' -exec cp {} /tmp/ref.dir \;

# Limpiar y correr nuestra herramienta sobre el mismo estado
rm -rf .dvc/cache data.dvc data/.gitignore
../dvcgo add data
```

**Esperado**:
- `data.dvc` byte-idéntico a `/tmp/ref.dvc` (`diff data.dvc /tmp/ref.dvc` sin salida).
- El objeto `.dir` generado byte-idéntico a `/tmp/ref.dir`.
- El path del `.dir` en cache coincide (`files/md5/<2>/<resto>.dir`).

## Escenario 2 — Velocidad de add/status (P1, SC-001/SC-005)

```bash
# Generar dataset grande (ej. 100k archivos pequeños)
./scripts/gen-dataset.sh big 100000

# add con nuestra herramienta vs DVC
time ../dvcgo add big       # esperado: >=10x más rápido que `time dvc add big`
../dvcgo status             # esperado: "up to date", sin rehashear (rápido, prop. a nº entradas)
touch big/file_00001.bin    # cambia mtime de un archivo (sin cambiar contenido)
../dvcgo status             # esperado: detecta correctamente vía state DB
```

**Esperado**: `add` ≥10× más rápido (SC-001); `status` sin cambios no recalcula hashes (SC-005); aprovecha múltiples núcleos.

## Escenario 3 — Round-trip completo con remote S3-compatible (P1/P2, SC-003)

Usar MinIO local (vía testcontainers en tests, o manual):

```bash
# MinIO manual
docker run -d -p 9000:9000 minio/minio server /data
dvcgo remote add -d local s3://bucket/store    # o editar .dvc/config
dvcgo remote modify local endpointurl http://localhost:9000

dvcgo add data
dvcgo push                      # sube solo faltantes; .dir tras sus contenidos
rm -rf .dvc/cache data          # simular clon limpio (queda data.dvc)
dvcgo pull                      # fetch + checkout, con verificación de integridad
```

**Esperado**: `data/` restaurado íntegro; integridad verificada al 100%; interoperable con `dvc pull` de Python sobre el mismo remote (probar ambos sentidos).

## Escenario 4 — Reanudación ante interrupción (SC-007)

```bash
dvcgo push & sleep 1 && kill -9 %1     # matar a mitad de transferencia
dvcgo push                             # reintento
```

**Esperado**: sin objetos corruptos a medio escribir; el reintento converge a remote íntegro; objetos ya presentes se omiten.

## Escenario 5 — checkout con estrategias de link (P2, FR-017)

```bash
dvcgo checkout                 # default reflink,copy
stat -c '%a' data/cat.jpeg     # objeto de cache 0o444 si linkeado
```

**Esperado**: materializa con la estrategia más eficiente disponible; degrada a copy en FS sin reflink; no toca archivos ya coincidentes.

## Escenario 6 — gc seguro (P3, FR-019/FR-020)

```bash
# crear y desreferenciar una versión
dvcgo gc                       # esperado: pide confirmación, muestra alcance
dvcgo gc -f                    # elimina no referenciados; reporta espacio
dvcgo checkout                 # versiones vigentes siguen restaurables
```

**Esperado**: solo se eliminan objetos no referenciados; confirmación requerida sin `-f`.

---

## Suite de validación automatizada

- **Unit/oráculo**: tests table-driven que comparan bytes de `.dvc` y `.dir` contra fixtures generados por DVC real (Escenario 1 como test). Gate de CI.
- **Integración remota**: `testcontainers-go/modules/minio` levanta S3-compatible; cubre Escenarios 3/4. Gateados con `testing.Short()`.
- **Interoperabilidad bidireccional** (SC-002): matriz de repos operados alternando `dvcgo` y `dvc` (Python) sin errores.
