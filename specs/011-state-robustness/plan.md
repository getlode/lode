# Implementation Plan: Robustez del state DB

**Branch**: `011-state-robustness` | **Date**: 2026-06-20 | **Spec**: [spec.md](spec.md)

## Summary

Garantizar que la optimización del state DB nunca produzca un falso "up to date": agregar un modo de rehash forzado (`--rehash` / `State.ForceRehash`) que ignora la heurística de metadata pero refresca el cache (auto-reparable), y degradar a rehash —en vez de abortar— cuando el state DB falta o está corrupto. Documentar la garantía: la heurística `(inode, mtime, size)` es optimización, no fuente de verdad.

## Technical Context

**Language/Version**: Go 1.23. State DB en bbolt.
**Testing**: unit tests (ForceRehash fuerza miss; cambio de tamaño/contenido detectado); e2e `status --rehash`; suite completa sin regresión.
**Constraints**: el caso común (metadata confiable) no se regresiona — el modo estricto es opcional. El rehash es la garantía última de corrección. Sin cambios de formato.

## Constitution Check

Sin violaciones. Refuerza corrección de detección, no toca formato/performance del caso común.

## Project Structure

```text
internal/hashfile/state.go        # campo ForceRehash; Get() lo respeta (miss forzado)
internal/cli/output.go            # helper openState(): degrada a rehash si el DB falla/corrupto
internal/cli/root.go              # flag persistente --rehash
internal/cli/{add,status}.go      # usan openState() (tolerante a fallo)
internal/hashfile/state_test.go   # tests de ForceRehash y detección de cambios
docs/ARCHITECTURE.md              # garantía de correctitud del state cache
```

**Structure Decision**: el state cache se vuelve degradable y anulable sin tocar el formato; la corrección se garantiza por el camino de rehash.

## Complexity Tracking

> No aplica.
