# Quickstart & Validation Guide: init + doctor

**Feature**: 002-init-onboarding | **Phase**: 1

Escenarios ejecutables. Binario: `lode`. Contratos en [contracts/cli.md](contracts/cli.md); estructuras en [data-model.md](data-model.md).

## Prerrequisitos

- Go 1.23+ (`CGO_ENABLED=0`).
- DVC de Python instalado **solo** para los tests-oráculo de byte-compat.
- git disponible para los escenarios de modo `scm`.

## Escenario 1 — Oráculo de bytes de `init` (gate de compat, SC-006)

Prueba que `lode init` produce exactamente lo mismo que `dvc init` en ambos modos.

```bash
# no-scm
mkdir t1 && cd t1 && lode init --no-scm
mkdir ../ref1 && cd ../ref1 && dvc init --no-scm -q
diff -r <(cd ../t1 && find . -type f | sort) <(find . -type f | sort)   # mismos archivos
cmp ../t1/.dvc/config .dvc/config && cmp ../t1/.dvcignore .dvcignore     # mismos bytes

# scm
mkdir ../t2 && cd ../t2 && git init -q && lode init
mkdir ../ref2 && cd ../ref2 && git init -q && dvc init -q
cmp ../t2/.dvc/.gitignore .dvc/.gitignore && cmp ../t2/.dvc/config .dvc/config
```

**Esperado**: misma estructura y bytes idénticos en `config`, `.gitignore`, `.dvcignore`; `.dvc/tmp/btime` presente y vacío; sin `.dvc/cache`.

## Escenario 2 — De cero a pusheado sin Python (P1, SC-001/SC-005)

```bash
mkdir proj && cd proj
mkdir data && printf 'hello' > data/a.txt
lode init --no-scm          # 1
lode add data               # 2
# (configurar remote y) push
lode remote add -d r s3://bucket/store && lode remote modify r endpointurl http://localhost:9000
lode push                   # 3
```

**Esperado**: el flujo completa sin DVC/Python; 3 comandos del onboarding; el repo resultante es operable por DVC (correr `dvc status` y verificar "up to date").

## Escenario 3 — Errores que guían (P2, SC-003)

```bash
cd /tmp && mkdir empty && cd empty
lode add foo        # esperado: error que sugiere `lode init` (y --no-scm / --cd)
lode init --no-scm && lode push   # esperado: error que sugiere configurar un remote
```

## Escenario 4 — Init seguro ante repos existentes (FR-005/006)

```bash
lode init --no-scm && lode init --no-scm   # 2da vez: "already initialized", no destruye
mkdir sub && cd sub && lode init --no-scm  # detecta repo padre; no crea anidado
```

## Escenario 5 — `doctor` (P2, SC-004)

```bash
lode doctor                 # repo sano: todo OK, exit 0
# sembrar problemas:
rm -rf .dvc                 # sin repo
lode doctor || echo "exit=$?"   # detecta "no repo", sugiere init, exit≠0
# sin remote / remote inalcanzable / cache no escribible / formato legacy:
#   verificar que cada uno se reporta con sugerencia y exit code correcto
```

**Esperado**: cada clase de problema (sin repo, sin remote, remote inalcanzable, cache no escribible, formato legacy 2.x) se identifica con su sugerencia; exit 0 solo si todo OK.

## Suite automatizada

- **Oráculo init** (`tests/oracle`): compara bytes de la estructura `lode init` vs `dvc init` en ambos modos. Gate de compat.
- **Integración doctor** (`tests/integration`): siembra cada clase de problema (incl. remote inalcanzable vía MinIO apagado) y verifica detección + exit code.
- **Interop** (SC-002): repo creado por `lode init` operado por `dvc`, y viceversa.
