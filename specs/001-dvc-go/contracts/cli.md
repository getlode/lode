# CLI Contract: comandos del MVP

**Feature**: 001-dvc-go | **Phase**: 1

Contrato de la interfaz de línea de comandos. Nombres y semántica compatibles con DVC. La herramienta opera sobre un repo DVC existente (no crea formato propio). Flags globales heredados por todos los subcomandos.

## Flags globales (persistentes)

| Flag | Default | Descripción |
|---|---|---|
| `-v, --verbose` | off | Logs detallados |
| `-q, --quiet` | off | Solo errores |
| `-j, --jobs N` | ~4×CPU (remoto) / NumCPU (hash) | Concurrencia |
| `--cd DIR` | cwd | Ejecutar como si el cwd fuera DIR |

Salida: humana por defecto en stdout; errores accionables en stderr; código de salida 0 OK, ≠0 error. Comandos de consulta aceptan `--json`.

---

## `add <target>...`

Trackea archivo(s) o directorio(s).

- **Pre**: target existe; repo tiene `.dvc/`.
- **Efecto**: calcula md5 (paralelo), mueve contenido al cache (`files/md5/...`, modo 0o444), escribe `<target>.dvc`, agrega entrada a `<dir>/.gitignore`.
- **Salida**: lista de `.dvc` creados/actualizados.
- **Errores**: target inexistente; archivo modificado durante el add (falla segura, no escribe metadata).
- **Compat**: el `.dvc` y el/los objeto(s) de cache producidos son byte-compatibles con DVC (incluido el objeto `.dir`).

## `status [target]...`

Reporta cambios sin modificar el repo.

- **Efecto**: para cada out, compara estado del workspace vs `.dvc` usando el state DB (no rehashea si `(ino,mtime,size)` no cambió).
- **Salida**: por target, estado `not in cache` / `modified` / `new` / `deleted` / `up to date`. Con `--json`, objeto estructurado.
- **Pos**: no escribe nada en cache ni workspace.

## `checkout [target]...`

Materializa el workspace según los `.dvc`.

- **Pre**: objetos referenciados presentes en cache (si faltan, informa y sugiere `pull`).
- **Efecto**: materializa vía `cache.type` (reflink→hardlink/symlink→copy); no toca archivos que ya coinciden; aplica protección 0o444 a links.
- **Salida**: archivos añadidos/modificados/eliminados.
- **Errores**: objeto faltante en cache (lista lo que falta).

## `push [target]...`

Sube objetos al remote.

- **Pre**: remote configurado (default o `-r <name>`).
- **Efecto**: calcula status del remote (HEAD-por-objeto o LIST masivo según heurística), sube solo faltantes; dirs: contenidos primero, `.dir` solo si todos OK; transferencia concurrente; reanudable.
- **Flags**: `-r, --remote NAME`.
- **Salida**: N objetos subidos / ya presentes.
- **Errores**: credenciales faltantes/ inválidas (mensaje claro); transferencia interrumpida → reintento idempotente.

## `pull [target]...`

`fetch` + `checkout`.

- **Efecto**: descarga objetos faltantes del remote al cache (escritura atómica tmp+rename, verificación de integridad), luego materializa el workspace.
- **Flags**: `-r, --remote NAME`.
- **Errores**: hash mismatch → descarta objeto y falla; objeto inexistente en remote.

## `fetch [target]...`

Solo remote → cache (no toca workspace). Misma maquinaria/flags que `pull` sin la fase de checkout.

## `gc`

Elimina objetos no referenciados.

- **Pre**: requiere confirmación explícita o `-f, --force`.
- **Flags**: `-w/--workspace` (alcance referencias), `-c/--cloud` (también remote), `-r/--remote NAME`, `-f/--force`.
- **Efecto**: calcula objetos alcanzables desde refs vigentes; muestra qué eliminaría; con confirmación, borra del cache (y remote si `-c`).
- **Salida**: espacio recuperado / N objetos eliminados.
- **Pos**: versiones vigentes siguen restaurables.

---

## Invariantes transversales (todos los comandos)

1. Toman el lock global `.dvc/tmp/lock` (flock) antes de mutar; respetan rwlock para paridad.
2. Escrituras a cache/workspace son atómicas (tmp + rename); objetos de cache quedan 0o444.
3. Mensajes de error sin trazas internas en uso normal.
4. Archivos grandes vía streaming (sin cargar en memoria).
5. Salida estructurada (`--json`) disponible en comandos de consulta.
