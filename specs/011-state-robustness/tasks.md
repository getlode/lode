---
description: "Task list for 011-state-robustness"
---

# Tasks: Robustez del state DB

- [X] T001 state.go: campo ForceRehash; Get() reporta miss cuando está activo (Put sigue refrescando) per FR-001/002/003
- [X] T002 CLI: flag persistente --rehash que activa ForceRehash per FR-002/003
- [X] T003 openState() degrada a rehash (nil + aviso) si el state DB falta o está corrupto, sin abortar per FR-004
- [X] T004 add/status usan openState() (tolerante) per FR-004
- [X] T005 Tests: ForceRehash fuerza miss; cambio de tamaño/contenido detectado (state_test.go) per FR-001
- [X] T006 docs/ARCHITECTURE.md: documentar la garantía (heurística = optimización, rehash = correcto) per FR-006
- [X] T007 e2e `status --rehash` + suite completa sin regresión; coexistencia DVC (mtime cambia → rehash) per FR-005
