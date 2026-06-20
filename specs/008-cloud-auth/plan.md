# Implementation Plan: Auth cloud de producción (S3)

**Branch**: `008-cloud-auth` | **Date**: 2026-06-20 | **Spec**: [spec.md](spec.md)

## Summary

Reincorporar el provider IAM (EC2 instance role / ECS / EKS-IRSA) a la cadena de credenciales S3 con un cliente HTTP **acotado por timeout** —la causa por la que se había omitido (panic con client nil + cuelgue off-cloud)— y soportar perfiles de credenciales. Precedencia predecible: explícito → env → archivo (perfil) → IAM. Desbloquea CI en EKS/GKE sin claves estáticas.

## Technical Context

**Language/Version**: Go 1.23. minio-go credentials chain.
**Testing**: test de integración del fast-fail off-cloud sin panic/cuelgue; la suite completa (MinIO) confirma que el cambio del chain no rompe push/pull/interop/doctor.
**Constraints**: la resolución por endpoint de metadata se acota con `http.Client{Timeout}` para degradar limpio fuera de la nube. Perfiles vía `FileAWSCredentials{Profile}`/`AWS_PROFILE`. Sin cambios de formato.

## Constitution Check

Sin violaciones. Cambia autenticación, no formato/performance. III (cero cgo) intacto.

## Project Structure

```text
internal/remote/s3.go     # chain env -> file(profile) -> IAM(timeout-bounded client)
tests/integration/auth_test.go  # regresión: off-cloud sin creds no panic/no cuelga, acotado
SECURITY.md               # precedencia de credenciales (incl. IAM/IRSA/perfil)
```

**Structure Decision**: cambio localizado en `NewS3`. Lo testeable hoy (fast-fail/no-panic + no regresión de la suite) se cubre; la validación contra EC2/EKS reales queda para el entorno cloud.

## Complexity Tracking

> No aplica.
