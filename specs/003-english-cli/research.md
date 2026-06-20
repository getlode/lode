# Research: CLI en inglés y pulido

**Feature**: 003-english-cli | **Date**: 2026-06-20 | **Phase**: 0

Feature mecánico (cambia textos, no semántica). Sin dependencias nuevas, sin incógnitas técnicas. El "research" es el inventario de cadenas y el glosario canónico.

## 1. Inventario de cadenas en español a traducir

Relevadas en `internal/`. Cadenas visibles al usuario (help, progreso, errores):

| Archivo | Tipo | Cadena (ES) |
|---|---|---|
| `cli/root.go` | Short root | "Versionado de datos rápido, drop-in compatible con DVC" |
| `cli/add.go` | Short / progreso / error | "Trackea archivos o directorios..."; "%-20s tracked -> %s"; "el archivo cambió durante el add; abortado" |
| `cli/status.go` | Short / flag | "Reporta cambios sin modificar el repo"; "salida estructurada en JSON" |
| `cli/push.go` | Short ×3 / progreso ×3 | "Sube objetos al remote"; "Descarga objetos del remote al cache..."; "Descarga del remote y materializa..."; "%d archivos subidos, %d ya presentes, %d fallidos"; "%d objetos descargados, %d ya en cache"; "workspace actualizado (%d salidas)" |
| `cli/checkout.go` | Short / progreso | "Materializa el workspace según los .dvc..."; "%d salidas materializadas" |
| `cli/gc.go` | Short / progreso / prompt | "Elimina objetos no referenciados..."; "No hay objetos no referenciados que eliminar."; "Se eliminarán %d objetos del cache (%s)."; "Cancelado."; "Liberados %s del cache local."; "Eliminados %d objetos..."; "¿Continuar? (yes/no): " |
| `cli/remote.go` | Short / progreso / error | "Gestiona remotes en .dvc/config"; "remote %q agregado"; "opción de remote desconocida: %s" |
| `cli/transfer.go` | errores | "el remote %q no está configurado"; "manifiesto %s no está en cache (agregá los datos primero)" |
| `remote/s3.go` | errores | "remote sin url"; "url de remote debe empezar con s3://"; "url de remote sin bucket" |

Ya en inglés (features 002): `init.go`, `doctor.go`, `errors.go` y sus mensajes. Las descripciones de flags de remote/gc (`--default`, `--force`, etc.) ya están parcialmente en español — se revisan en la misma pasada.

**Decision**: traducir todas las de arriba, además de cualquier flag-description en español detectada durante la implementación.

## 2. Glosario canónico (términos únicos)

**Decision**: un término por concepto; se preservan los nombres que usa DVC para conceptos compartidos.

| Concepto | Término canónico (EN) | Prohibido (sinónimos) |
|---|---|---|
| Almacenamiento remoto | **remote** | (igual que DVC) |
| Cache local | **cache** | "local store" |
| Árbol de trabajo | **workspace** | "working tree" |
| Objeto de contenido | **object** | "blob", "artifact" |
| Salida trackeada | **tracked output** / **output** | "entry" |
| Repositorio | **repository** | "repo" en prosa de usuario (en flags/short ok abreviar) |
| Manifiesto de directorio | **directory manifest** (`.dir`) | — |

**Rationale**: familiaridad para quien viene de DVC; consistencia para reducir ambigüedad (FR-003/SC-003).

## 3. Pluralización

**Decision**: para mensajes con conteos (push/gc/checkout), usar un helper mínimo de singular/plural que lea natural en inglés (p. ej. "1 file" / "3 files").

`func plural(n int, singular, pluralForm string) string` o `"%d file(s)"` solo donde no valga la pena. Preferir el helper para los mensajes principales de resultado.

**Rationale**: FR-005/SC — los conteos deben leerse naturalmente.

**Alternatives considered**: librería de i18n con reglas plurales (rechazado: fuera de alcance, inglés único).

## 4. Verificación (cómo se prueba SC-001/SC-002)

**Decision**: dos mecanismos:
1. **Test de barrido**: correr `--help` de root y de cada subcomando, más caminos de error frecuentes, y afirmar que la salida no contiene caracteres no-ASCII de español ni palabras de una lista negra (ej. "archivo", "remoto", "objetos", "salidas"). En `tests/integration/english_cli_test.go`.
2. **Re-habilitar el linter `misspell`** en `.golangci.yml` (se había desactivado por el español): ahora que el texto es inglés, `misspell` actúa como gate de typos y de español remanente.

**Rationale**: SC-001 exige 100% inglés verificable; el barrido + misspell lo cubren de forma automatizada.

## 5. Alcance e invariantes

**Decision**: solo cambia texto de interfaz. NO se construye framework de i18n. NO cambia ninguna semántica de comando, flag, ni el formato de archivos (Constitución I/II intactos). Cero deps nuevas, cero cgo (III).

**Riesgo**: bajísimo. El único cuidado es no alterar cadenas que sean parte de un contrato (p. ej. el prompt de confirmación de `gc` acepta "yes/y" — se mantiene esa semántica; solo se traduce el texto).
