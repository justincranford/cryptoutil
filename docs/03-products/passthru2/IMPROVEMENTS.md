# Passthru2: Improvement & PR Task Suggestions

**Updated**: 2025-11-30 (aligned with Grooming Sessions 1 & 2 decisions)

This file lists suggested quick PR tasks and improvements prioritized for immediate work.

---

## CRITICAL: TLS/HTTPS Fix (from Q20)

| PR | Description | Priority | Related Decision |
|----|-------------|----------|------------------|
| PR-0000 | **FIX: Identity TLS to reuse KMS cert utility pattern** | CRITICAL | Q20 |

**Details**: passthru2 mixed HTTPS with HTTP. Identity MUST use KMS CA-chained cert pattern.

---

## Quick Fixes (High Impact, Low Risk)

| PR | Description | Priority | Related Decision |
|----|-------------|----------|------------------|
| PR-0001 | Extract telemetry to `deployments/telemetry/compose.yml` | HIGHEST | Q6 |
| PR-0002 | Create `compose.demo.yml` for KMS and Identity | HIGHEST | Q8, Q12 |
| PR-0003 | Convert Identity secrets to Docker secrets | HIGH | Q7, Q10 |
| PR-0004 | Standardize config locations under `deployments/<product>/config/` | HIGH | Q7 |
| PR-0005 | Add `--demo` flag to KMS server with seeding logic | HIGH | Q11, Q13 |
| PR-0006 | Add `--reset-demo` flag for demo data cleanup | HIGH | Q15 |

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
| PR-0107 | Profile-based persistence (dev=persist, ci=ephemeral) | MEDIUM | Q12 |

---

## Token Validation PRs (from Q6-10)

| PR | Description | Priority | Related Decision |
|----|-------------|----------|------------------|
| PR-0201 | In-memory JWKS caching with configurable TTL | HIGH | Q6 |
| PR-0202 | Configurable revocation check frequency | HIGH | Q7 |
| PR-0203 | 401/403 error split + configurable detail level | HIGH | Q8 |
| PR-0204 | Multi-method service auth (client-creds/mTLS/API-key) | MEDIUM | Q9 |
| PR-0205 | Full OIDC + custom claims extraction | MEDIUM | Q10 |
| PR-0206 | Token validation middleware integration | HIGH | Q17 |
| PR-0207 | Hybrid scope enforcement for KMS | HIGH | Q18 |

---

## Integration PRs

| PR | Description | Priority | Related Decision |
|----|-------------|----------|------------------|
| PR-0301 | Create `cmd/demo-all/main.go` Go CLI | MEDIUM | Q12 (priority 3) |
| PR-0302 | E2E tests for integration demo | MEDIUM | Q23 |
| PR-0303 | Full dependency chain health checks | HIGH | Q16 |
| PR-0304 | Per-product + shared telemetry networks | MEDIUM | Q17 |

---

## KMS Realm Authentication PRs (from Q1-5)

| PR | Description | Priority | Related Decision |
|----|-------------|----------|------------------|
| PR-0401 | Separate `realms.yml` config file format | HIGH | Q1 |
| PR-0402 | PBKDF2 + plaintext password support | HIGH | Q2 |
| PR-0403 | `kms_realm_users` table for DB realm | MEDIUM | Q3 |
| PR-0404 | Configurable realm priority order | MEDIUM | Q4 |
| PR-0405 | Database-level tenant isolation | MEDIUM | Q5 |
| PR-0406 | Identity federation configuration | LOW | Q11 |

---

## Testing PRs (from Q21-25)

| PR | Description | Priority | Related Decision |
|----|-------------|----------|------------------|
| PR-0501 | UUIDv7 unique prefixes for all tests (CRITICAL) | HIGHEST | Q23 |
| PR-0502 | Basic benchmarks for critical paths | MEDIUM | Q24 |
| PR-0503 | Test case descriptions in code | LOW | Q25 |
| PR-0504 | Integration test scope (startup + CRUD + flow) | MEDIUM | Q21 |
| PR-0505 | Testcontainers for unit/integration tests | MEDIUM | Q22 |

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
- ❌ Self-signed TLS leaf certs (Q20)
- ❌ bcrypt for password hashing (FIPS - use PBKDF2)

---

## Notes

- PR-0000 is CRITICAL and must be fixed first (TLS pattern)
- PR-0001 through PR-0006 are Phase 0 prerequisites
- PR-0101 is highest priority for KMS parity (Q13 priority B)
- PR-0501 is CRITICAL for parallel test execution
- PR-1001-1005 should be done one package at a time per Q22 (hybrid migration)
- All coverage targets are 80% minimum per Q21
