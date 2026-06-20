# Contract: user-facing strings (ES → EN)

**Feature**: 003-english-cli | **Phase**: 1

El "contrato" de este feature es el texto en inglés. Tabla de referencia (la implementación puede pulir la redacción manteniendo el sentido y el glosario).

## Command short descriptions

| Command | English short |
|---|---|
| root | "Fast, drop-in compatible data versioning (DVC 3.x)" |
| `add` | "Track files or directories (drop-in compatible with DVC)" |
| `status` | "Show what changed, without modifying the repository" |
| `push` | "Upload objects to the remote" |
| `fetch` | "Download objects from the remote into the cache (no workspace changes)" |
| `pull` | "Download from the remote and materialize the workspace (fetch + checkout)" |
| `checkout` | "Materialize the workspace from the cache" |
| `gc` | "Remove unreferenced objects from the cache (and the remote with -c)" |
| `remote` | "Manage remotes in .dvc/config" |

## Flag descriptions

| Flag | English |
|---|---|
| `status --json` | "structured JSON output" |
| (other flags) | review and translate any remaining Spanish flag help in the same pass |

## Progress / result messages

| Context | English (with natural plurals) |
|---|---|
| add tracked | "tracked %s -> %s" |
| add changed-during | "the file changed during add; aborted" |
| push result | "uploaded %d object(s), %d already present, %d failed" |
| fetch result | "downloaded %d object(s), %d already in cache" |
| pull result | "updated workspace (%d output(s))" |
| checkout result | "materialized %d output(s)" |
| gc nothing | "No unreferenced objects to remove." |
| gc preview | "Will remove %d object(s) from the cache (%s)." |
| gc cancelled | "Cancelled." |
| gc freed | "Freed %s from the local cache." |
| gc remote | "Removed %d unreferenced object(s) from the remote." |
| gc prompt | "Continue? (yes/no): " (still accepts yes/y) |
| remote added | "remote %q added" |

## Errors

| Context | English |
|---|---|
| unknown remote option | "unknown remote option: %s" |
| remote not configured | "remote %q is not configured" |
| manifest not in cache | "manifest %s is not in the cache (add the data first): %w" |
| remote without url | "remote has no url" |
| bad s3 url scheme | "remote url must start with s3://" |
| s3 url without bucket | "remote url has no bucket" |

> Plurals: prefer a helper for natural singular/plural; the `object(s)`/`output(s)` forms above are placeholders for "1 object" vs "3 objects".
