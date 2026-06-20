# Implementation Plan: Confiabilidad de transferencia

**Branch**: `010-transfer-reliability` | **Date**: 2026-06-20 | **Spec**: [spec.md](spec.md)

## Summary

Agregar retry con backoff exponencial + jitter y clasificación transitorio/permanente alrededor de las operaciones de red (`store.Put`/`store.Get`), y manejo de fallo parcial: push y fetch acumulan los objetos fallidos, los reportan y devuelven un error accionable (exit code) que indica reanudar; una reanudación es idempotente (salta lo ya presente vía `c.Has`/`MissingOnRemote`).

## Technical Context

**Language/Version**: Go 1.23. Reutiliza el motor de transferencia (errgroup) y la idempotencia existente.
**Testing**: tests unitarios de `retry`/`isTransient` (transitorio→éxito, permanente→falla rápido, agota, cancelación, clasificación); suite completa con MinIO sin regresión.
**Constraints**: jitter para no sincronizar reintentos; "connection refused"/"no such host" se tratan como permanentes (fail-fast, no demorar endpoints muertos); cancelación de contexto corta de inmediato. Sin cambios de formato.

## Constitution Check

Sin violaciones. Cambia robustez de transferencia, no formato. III (cero cgo) intacto.

## Project Structure

```text
internal/transfer/retry.go        # RetryPolicy, retry(), isTransient(), DefaultRetry
internal/transfer/push.go         # Put envuelto en retry (uploadSet)
internal/transfer/fetch.go        # Get envuelto en retry; fetch acumula fallos (resumible)
internal/cli/push.go              # surfacea Failed y exit code en fallo parcial (push y fetch)
internal/cli/errors.go            # errPartialTransfer (guía a reanudar)
internal/transfer/retry_test.go   # tests del retry/clasificación
```

**Structure Decision**: la capa de resiliencia se localiza en `transfer`, desacoplada de minio (clasificación por net.Error + marcadores de error), sin tocar el core ni el formato.

## Complexity Tracking

> No aplica.
