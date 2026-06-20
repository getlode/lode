# dvcgo

Versionado de datos rápido en Go, **drop-in compatible con [DVC](https://dvc.org) 3.x**.
Apuntá `dvcgo` a tu repositorio DVC actual y obtené el mismo formato (mismos
`.dvc`, mismo objeto `.dir`, mismo layout de cache y remote) con un binario único
y hashing paralelo. En datasets grandes es ~10× más rápido que DVC de Python.

## Por qué

DVC es el estándar para versionar datasets y modelos de ML, pero su CLI sufre en
repos grandes: el hashing es CPU-bound y está limitado por el runtime de Python.
`dvcgo` reimplementa el núcleo de versionado en Go: binario estático sin
dependencias, hashing concurrente y un state DB que evita rehashear lo que no
cambió. Y **coexiste con DVC**: ambos operan el mismo repo de forma intercambiable.

## Instalación

```bash
go install github.com/jtorchia/dvcgo/cmd/dvcgo@latest
# o descargá un binario de Releases / brew install jtorchia/tap/dvcgo
```

## Uso

```bash
dvcgo add data/            # trackea un directorio (o un archivo)
dvcgo status               # qué cambió, sin rehashear lo intacto
dvcgo remote add -d r s3://bucket/store
dvcgo remote modify r endpointurl https://nyc3.digitaloceanspaces.com
dvcgo push                 # sube al remote S3-compatible
dvcgo pull                 # fetch + checkout en un clon limpio
dvcgo checkout             # materializa el workspace desde el cache
dvcgo gc -f                # libera objetos no referenciados
```

| Comando | Qué hace |
|---|---|
| `add` | Hashea (paralelo), cachea y escribe el `.dvc`; actualiza `.gitignore` |
| `status` | Reporta cambios usando el state DB (sin rehashear lo intacto) |
| `push` / `fetch` / `pull` | Sincroniza con un remote S3-compatible (S3, MinIO, R2, B2) |
| `checkout` | Materializa el workspace (reflink → hardlink/symlink → copy) |
| `gc` | Elimina objetos no referenciados del cache (y del remote con `-c`) |

## Compatibilidad con DVC

- Formato `.dvc` y objeto `.dir` **byte-idénticos** a DVC 3.x (validado por un
  test-oráculo contra el `dvc` real).
- Mismo layout content-addressed en cache y remote (`files/md5/<2>/<resto>`).
- Interoperabilidad **bidireccional**: lo que sube `dvcgo` lo baja DVC y viceversa.
- Lectura del cache legacy 2.x (`<2>/<resto>`).

### Fuera del alcance del MVP

Pipelines/`repro`, y backends de remote no-S3 (GCS, Azure, SSH).

## Desarrollo

```bash
make build           # binario (CGO_ENABLED=0)
make test            # todos los tests
make test-short      # omite la integración con S3
make oracle          # gate de compatibilidad de bytes contra DVC real

# Integración con S3 (MinIO) e interop con DVC real:
MINIO_ENDPOINT=http://localhost:9000 MINIO_ACCESS_KEY=... MINIO_SECRET_KEY=... \
  DVC_BIN=$(which dvc) DVCGO_BIN=$(pwd)/dvcgo go test ./tests/...
```

## Licencia

Apache-2.0 (núcleo). Ver issues/roadmap para la capa cloud/enterprise.
