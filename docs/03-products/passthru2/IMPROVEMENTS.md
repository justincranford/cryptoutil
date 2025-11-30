# Passthru2: Improvement & PR Task Suggestions

**Updated**: 2025-11-30 (aligned with Grooming Session 1 decisions)

This file lists suggested quick PR tasks and improvements prioritized for immediate work.

---

## Quick Fixes (High Impact, Low Risk)

| PR | Description | Priority | Related Decision |
|----|-------------|----------|------------------|
| PR-0001 | Extract telemetry to `deployments/telemetry/compose.yml` | HIGHEST | Q6 |
| PR-0002 | Create `compose.demo.yml` for KMS and Identity | HIGHEST | Q8, Q12 |
| PR-0003 | Convert Identity secrets to Docker secrets | HIGH | Q7, Q10 |
| PR-0004 | Standardize config locations under `deployments/<product>/config/` | HIGH | Q7 |
| PR-0005 | Add `--demo` flag to KMS server with seeding logic | HIGH | Q11, Q13 |

---

## Medium PRs (Demo Parity)

| PR | Description | Priority | Related Decision |
|----|-------------|----------|------------------|
| PR-0101 | KMS Swagger UI "Try it out" with demo credentials | HIGHEST | Q13 (priority B) |
| PR-0102 | KMS demo mode auto-seed (accounts, pools, keys) | HIGH | Q13 (priority A) |
| PR-0103 | Create `cmd/demo-kms/main.go` Go CLI | MEDIUM | Q12 (priority 2) |
| PR-0104 | Identity `/authorize` endpoint, PKCE, redirect flow | HIGH | Q2 |
| PR-0105 | Identity demo seed data and `--demo` mode | HIGH | Q11 |
| PR-0106 | Create `cmd/demo-identity/main.go` Go CLI | MEDIUM | Q12 (priority 2) |

---

## Integration PRs

| PR | Description | Priority | Related Decision |
|----|-------------|----------|------------------|
| PR-0201 | Token validation middleware (local + introspection) | HIGH | Q17 |
| PR-0202 | Hybrid scope enforcement for KMS | HIGH | Q18 |
| PR-0203 | Create `cmd/demo-all/main.go` Go CLI | MEDIUM | Q12 (priority 3) |
| PR-0204 | E2E tests for integration demo | MEDIUM | Q23 |

---

## KMS Realm Authentication PRs (from Q11)

| PR | Description | Priority | Related Decision |
|----|-------------|----------|------------------|
| PR-0301 | File realm configuration schema and loader | MEDIUM | Q11 |
| PR-0302 | DB realm for PostgreSQL mode | LOW | Q11 |
| PR-0303 | Identity federation configuration | LOW | Q11 |

---

## Larger PRs (Migration - Q22)

| PR | Description | Priority | Related Decision |
|----|-------------|----------|------------------|
| PR-1001 | Move `internal/common/apperr` → `internal/infra/apperr` | LOW | Q22 (hybrid) |
| PR-1002 | Move `internal/common/config` → `internal/infra/config` | LOW | Q22 (hybrid) |
| PR-1003 | Move `internal/common/magic` → `internal/infra/magic` | LOW | Q22 (hybrid) |
| PR-1004 | Move `internal/common/telemetry` → `internal/infra/telemetry` | LOW | Q22 (hybrid) |
| PR-1005 | Consolidate Identity duplicate infra code | LOW | Q22 (hybrid) |

---

## CI & Quality PRs (Q21, Q24)

| PR | Description | Priority | Related Decision |
|----|-------------|----------|------------------|
| CI-001 | Add demo profile CI job for KMS | MEDIUM | Q24 |
| CI-002 | Add demo profile CI job for Identity | MEDIUM | Q24 |
| CI-003 | Add coverage threshold enforcement (80%) | MEDIUM | Q21 |
| CI-004 | Add SQLite/PostgreSQL matrix runs | LOW | Q20, Q24 |

---

## BANNED Approaches (from Q12)

**DO NOT implement**:

- ❌ `make demo` or any Makefiles
- ❌ Bash scripts (`*.sh`)
- ❌ PowerShell scripts (`*.ps1`)
- ❌ Embedded Identity in KMS (Q15)

---

## Notes

- PR-0001 through PR-0005 are Phase 0 prerequisites
- PR-0101 is highest priority for KMS parity (Q13 priority B)
- PR-1001-1005 should be done one package at a time per Q22 (hybrid migration)
- All coverage targets are 80% minimum per Q21

