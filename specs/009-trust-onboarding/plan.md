# Implementation Plan: Confianza, posicionamiento y onboarding

**Branch**: `009-trust-onboarding` | **Date**: 2026-06-20 | **Spec**: [spec.md](spec.md)

## Summary

Comunicar la reversibilidad/cero lock-in (el activo de confianza que ya existe técnicamente), posicionar honestamente (acelerador que convive con DVC), desarmar el miedo a la AGPL para uso interno, bajar la fricción de contribución (roadmap, arquitectura, good-first-issues, CLA honesto), y cerrar el último leak de español en la salida de usuario. Casi todo es docs; el único cambio de código es traducir un error en `fetch.go`.

## Technical Context

**Language/Version**: Go 1.23 (un fix de string) + Markdown (docs).
**Testing**: build/vet/lint/suite verdes; el gate de inglés (`english_cli_test`) cubre rutas de error.
**Constraints**: sin cambios de comportamiento (salvo el texto del error). Constitución I/II/III/V intactas.
**Project Type**: docs + un fix puntual.

## Constitution Check

Sin violaciones. No toca formato ni performance. Refuerza la coexistencia (V) comunicándola.

## Project Structure

```text
internal/transfer/fetch.go   # fix: error de corrupción a inglés
README.md                    # reversibilidad/cero lock-in, posicionamiento, nota AGPL, project status
ROADMAP.md                   # nuevo: hecho vs planeado
docs/ARCHITECTURE.md         # nuevo: mapa de paquetes
CONTRIBUTING.md              # CLA honesto (sin prometer un bot inexistente)
GitHub issues                # good-first-issues etiquetados
```

**Structure Decision**: docs + onboarding; un único cambio de código (string). La elección CLA vs DCO se mantiene como CLA con redacción honesta (decisión de governance), cumpliendo "claro, justo, mecanismo real".

## Complexity Tracking

> No aplica.
