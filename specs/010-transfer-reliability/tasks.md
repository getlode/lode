---
description: "Task list for 010-transfer-reliability"
---

# Tasks: Confiabilidad de transferencia

- [X] T001 retry.go: RetryPolicy + retry() con backoff exponencial + jitter, cancelación de contexto per FR-001
- [X] T002 isTransient(): clasificar transitorio (timeout/5xx/throttling/reset) vs permanente (auth/not-found/refused/DNS) per FR-002
- [X] T003 push: envolver store.Put en retry (uploadSet) per FR-001/003
- [X] T004 fetch: envolver store.Get en retry; acumular fallos en Result.Failed en vez de abortar (resumible) per FR-004/005
- [X] T005 CLI: surfacea Failed y devuelve errPartialTransfer (exit code) en push y fetch per FR-004
- [X] T006 Tests: transitorio→éxito, permanente→falla rápido, agota, cancelación, clasificación (retry_test.go)
- [X] T007 Suite completa (MinIO) sin regresión; reanudación idempotente vía MissingOnRemote/c.Has per FR-005/006
