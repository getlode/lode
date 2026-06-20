---
description: "Task list for 008-cloud-auth"
---

# Tasks: Auth cloud de producción (S3)

- [X] T001 Reincorporar el provider IAM con cliente HTTP acotado por timeout (EC2/ECS/IRSA) en la cadena, sin panic ni cuelgue off-cloud, en internal/remote/s3.go per FR-001/002
- [X] T002 Soportar perfil de credenciales (FileAWSCredentials{Profile}/AWS_PROFILE) per FR-003
- [X] T003 Precedencia predecible explícito→env→archivo(perfil)→IAM per FR-004
- [X] T004 Test de regresión: off-cloud sin credenciales no panic/no cuelga, acotado (tests/integration/auth_test.go) per FR-002
- [X] T005 SECURITY.md: documentar la precedencia (incl. IAM/IRSA/perfil) per FR-004
- [X] T006 Suite completa (MinIO) confirma cero regresiones del cambio de chain
