# Quickstart & Validation Guide: CLI en inglés

**Feature**: 003-english-cli | **Phase**: 1

Validación de que toda la salida visible está en inglés. Binario: `lode`.

## Escenario 1 — Help en inglés (SC-001/SC-002)

```bash
lode --help
for c in init add status push fetch pull checkout gc doctor remote; do lode $c --help; done
```

**Esperado**: títulos, descripciones de comandos y de flags en inglés; ninguna palabra en español ni caracteres acentuados de español; cada comando y flag con help no vacío.

## Escenario 2 — Mensajes de operación en inglés

```bash
mkdir t && cd t && lode init --no-scm
mkdir data && echo x > data/a && lode add data        # "tracked ..." en inglés
lode status                                            # en inglés
lode push 2>&1 || true                                 # error de no-remote en inglés (feature 002)
```

**Esperado**: progreso y resultado en inglés (p. ej. "uploaded N objects, ...", "materialized N outputs"); plurales naturales (1 object vs 3 objects).

## Escenario 3 — Errores en inglés

```bash
lode remote add r s3://b/p >/dev/null
lode remote modify r badoption x   # "unknown remote option: badoption"
```

**Esperado**: cada error en inglés y accionable.

## Suite automatizada

- **Barrido de inglés** (`tests/integration/english_cli_test.go`): corre `--help` de root y de cada subcomando + caminos de error; afirma que la salida no contiene caracteres no-ASCII de español ni palabras de una lista negra (archivo, remoto, objetos, salidas, materializa, etc.).
- **Linter `misspell`** re-habilitado en `.golangci.yml`: gate de typos y español remanente (corre en CI).
