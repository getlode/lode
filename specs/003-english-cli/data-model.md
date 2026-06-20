# Data Model: CLI en inglés

**Feature**: 003-english-cli | **Date**: 2026-06-20 | **Phase**: 1

Feature de texto: la "entidad" principal es el glosario canónico. No hay estructuras de datos nuevas.

## Glossary (canonical terms)

| Concept | Canonical term | Notes |
|---|---|---|
| Remote storage | `remote` | shared with DVC |
| Local cache | `cache` | shared with DVC |
| Working tree | `workspace` | |
| Content object | `object` | |
| Tracked output | `output` / `tracked output` | |
| Repository | `repository` | abbreviate to "repo" only in flag help |
| Directory manifest | `directory manifest` (`.dir`) | |

Cada concepto se nombra con su término canónico en TODA la interfaz; los sinónimos listados como prohibidos no aparecen.

## Interface string (concepto)

Cualquier literal emitido a stdout/stderr por la herramienta: short/long de comandos, descripciones de flags, mensajes de progreso/resultado, warnings, errores. Todos en inglés.

## Plural helper (concepto)

Para mensajes con conteos: una función que devuelve la forma singular o plural según `n` (p. ej. `1 file` vs `3 files`), usada en los mensajes de resultado de push/fetch/pull/checkout/gc.
