# SM-KMS Deep Analysis

**Date:** 2026-01-31
**Scope:** `internal/kms/` - main code, unit tests, integration tests, E2E tests, deployments, workflows, load testing

---

## Executive Summary

### Statistics
- **Total Go Files**: 100
- **Test Files**: 63 (63% of codebase)
- **All Unit Tests**: ✅ PASSING
- **Load Testing**: ✅ Gatling framework available

### Coverage Summary (CRITICAL - Below 95% Target)
| Package | Coverage | Status |
|---------|----------|--------|
| client | 74.9% | ⚠️ Below 95% target |
| cmd | 0.0% | ❌ No coverage |
| server/application | 77.1% | ⚠️ Below 95% target |
| server/businesslogic | 39.0% | ❌ Critical gap |
| server/demo | 7.3% | ❌ Critical gap |
| server/handler | 79.9% | ⚠️ Below 95% target |
| server/middleware | 53.1% | ❌ Critical gap |
| server/repository/orm | 88.9% | ⚠️ Below 95% target |
| server/repository/sqlrepository | 78.0% | ⚠️ Below 95% target |

**ALL packages are below the 95% coverage target - CRITICAL ISSUE.**

---

## Architecture Analysis

### Current State: NOT Using Service Template

**CRITICAL FINDING**: SM-KMS does NOT leverage the service template pattern used by cipher-im and jose-ja.

**Evidence**:
- `internal/kms/server/application/application_listener.go` has 1224 lines of custom HTTP server setup
- Duplicates middleware configuration, TLS setup, health checks, Fiber app initialization
- Does NOT use `internal/apps/template/service/server/builder/server_builder.go`
- Does NOT use merged migrations pattern from service template

### Current Architecture

```
internal/kms/
├── client/                          # KMS client library (3 files)
│   ├── client_oam_mapper.go
│   ├── client_test.go
│   └── client_test_util.go
├── cmd/                             # Server entry point (1 file)
│   └── server.go
└── server/                          # Server implementation
    ├── application/                 # Application layer (9 files)
    │   ├── application_basic.go     # Basic services (telemetry, unseal, JWK)
    │   ├── application_core.go      # Core services (SQL, ORM, barrier, business)
    │   ├── application_init.go      # Server initialization
    │   └── application_listener.go  # HTTP listener (1224 lines!)
    ├── businesslogic/               # Business logic (7 files)
    │   ├── businesslogic.go         # 742 lines of business logic
    │   └── elastic_key_status_state_machine.go
    ├── demo/                        # Demo seed data (2 files)
    ├── handler/                     # OpenAPI handlers (3 files)
    ├── middleware/                  # Auth middleware (14 files)
    └── repository/                  # Data layer
        ├── orm/                     # GORM repository (34 files!)
        └── sqlrepository/           # Raw SQL repository (26 files!)
```

---

## Issues Identified

### Category 1: Architecture Non-Conformance (CRITICAL)

#### Issue 1.1: Not Using Service Template Pattern
**Severity:** CRITICAL
**Impact:** Code duplication, maintenance burden, inconsistency across services

**Evidence**:
- `application_listener.go` duplicates 90% of functionality from service template
- Manual middleware stack configuration (lines 180-210)
- Custom TLS setup instead of using `server_builder.go`
- No merged migrations pattern

**Required Change**: Migrate to service template pattern like cipher-im and jose-ja.

---

#### Issue 1.2: Massive File Sizes
**Severity:** High
**Files Exceeding 500-Line Hard Limit**:

| File | Lines | Status |
|------|-------|--------|
| application_listener.go | 1224 | ❌ 2.4x hard limit |
| businesslogic.go | 742 | ❌ 1.5x hard limit |

**Per 03-01.coding.instructions.md**:
- Soft limit: 300 lines
- Medium limit: 400 lines
- Hard limit: 500 lines → refactor required

---

#### Issue 1.3: Dual Repository Pattern Anti-Pattern
**Severity:** Medium
**Location:** `internal/kms/server/repository/`

**Issue**: KMS has TWO separate repository implementations:
- `orm/` - GORM-based ORM repository (34 files)
- `sqlrepository/` - Raw database/sql repository (26 files)

**Per 03-04.database.instructions.md**:
> **ORM**: GORM (MANDATORY, never raw database/sql)

**Impact**: 60 repository files instead of ~20, maintenance duplication, inconsistent patterns.

---

### Category 2: Test Architecture Issues

#### Issue 2.1: No TestMain Pattern (V4 Carryover)
**Severity:** High
**Location:** All test files in `internal/kms/`

**Per 03-02.testing.instructions.md**:
> TestMain Pattern (REQUIRED for heavyweight dependencies)

**Current State**: Many test files create per-test database connections instead of shared TestMain setup.

---

#### Issue 2.2: Coverage Below 95% in ALL Packages
**Severity:** CRITICAL

| Package | Coverage | Gap |
|---------|----------|-----|
| businesslogic | 39.0% | -56% |
| middleware | 53.1% | -42% |
| demo | 7.3% | -88% |
| application | 77.1% | -18% |
| sqlrepository | 78.0% | -17% |

**Per 03-02.testing.instructions.md**:
> Coverage targets: ≥95% production, ≥98% infrastructure/utility

---

#### Issue 2.3: Missing t.Parallel() in Tests
**Severity:** Medium
**Impact:** Slower test execution, hidden race conditions

---

### Category 3: Missing Service Template Features

#### Issue 3.1: No Registration/Auth Flow
**Severity:** High

**Per 03-08.server-builder.instructions.md**:
> All services using the builder pattern MUST support the registration flow

**Current State**: KMS has no `/service/api/v1/auth/register` endpoint. Uses direct API key/token authentication without tenant registration.

---

#### Issue 3.2: No Multi-Tenancy Support
**Severity:** High

**Per 02-02.service-template.instructions.md**:
> Tenant-Based Data Isolation: `tenant_id` scopes ALL data access

**Current State**: KMS does not have tenant_id isolation in database queries.

---

#### Issue 3.3: No Session Management
**Severity:** High

**Per 02-02.service-template.instructions.md**:
> Session management via SessionManagerService

**Current State**: KMS uses custom JWT middleware without session management.

---

#### Issue 3.4: No Realm Support
**Severity:** Medium

**Per 02-02.service-template.instructions.md**:
> Realm-Based Authentication Only: `realm_id` authentication context ONLY

**Current State**: KMS has no realm concept.

---

### Category 4: Deployment Issues

#### Issue 4.1: Healthcheck Path Mismatch
**Severity:** Medium
**Location:** `deployments/kms/compose.yml:123`

**Current**:
```yaml
test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/v1/livez"]
```

**Expected** (per 02-02.service-template.instructions.md):
```yaml
test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/api/v1/livez"]
```

**Note**: Missing `/api/` in path.

---

#### Issue 4.2: Incomplete Docker Secrets Pattern
**Severity:** Low
**Location:** `deployments/kms/secrets/`

**Current Secrets**:
- hash_pepper_v3.secret (✅)
- postgres_*.secret (✅)
- unseal_*of5.secret (✅)

**Missing** (if migrating to service template):
- tls_cert.secret
- tls_key.secret

---

### Category 5: Load Testing Analysis

#### Issue 5.1: Load Tests Target KMS Directly
**Severity:** Low
**Location:** `test/load/`

**Current State**: Gatling tests target KMS API endpoints directly.
- ServiceApiSimulation.java - Tests /service/api/v1/* endpoints
- BrowserApiSimulation.java - Tests /browser/api/v1/* endpoints

**Post-Migration**: Tests should continue to work but may need path updates if API changes.

---

### Category 6: Workflow Analysis

#### Issue 6.1: ci-load.yml Targets KMS
**Severity:** Low
**Location:** `.github/workflows/ci-load.yml`

**Current State**: Load testing workflow is KMS-specific. Will continue to work post-migration.

---

## Migration Impact Analysis

### Files Requiring Major Refactoring

| File | Lines | Effort | Notes |
|------|-------|--------|-------|
| application_listener.go | 1224 | HIGH | Replace with server_builder.go |
| application_core.go | ~150 | MEDIUM | Adapt to builder pattern |
| application_basic.go | ~80 | MEDIUM | Merge with builder resources |
| businesslogic.go | 742 | MEDIUM | Keep, add tenant_id filtering |
| All middleware files | 14 files | HIGH | Replace with template middleware |
| orm/ directory | 34 files | MEDIUM | Add tenant_id, realm support |
| sqlrepository/ directory | 26 files | HIGH | REMOVE (violates GORM-only rule) |

### Files to Remove (After Migration)

| Directory | Files | Reason |
|-----------|-------|--------|
| sqlrepository/ | 26 | Violates GORM-only requirement |
| Duplicated middleware | ~10 | Use template middleware |

### Files to Add (Migration)

| File | Purpose |
|------|---------|
| migrations/2001_kms_*.sql | Domain-specific migrations |
| server.go using builder | New server entry point |
| registration handlers | User/tenant registration |
| session management | Template session integration |

---

## Recommendations Summary

### CRITICAL (Must Fix Before Release)
1. **Migrate to service template** - Eliminate 1000+ lines of duplicated infrastructure
2. **Fix coverage gaps** - 39% businesslogic coverage is unacceptable
3. **Remove sqlrepository/** - Violates GORM-only requirement
4. **Add multi-tenancy** - tenant_id isolation required

### HIGH Priority
5. **Refactor large files** - application_listener.go, businesslogic.go
6. **Implement TestMain pattern** - Required for integration tests
7. **Add registration flow** - Required by service template

### MEDIUM Priority
8. **Add realm support** - Authentication context
9. **Fix healthcheck paths** - /admin/api/v1/ not /admin/v1/
10. **Add t.Parallel()** - Test parallelism

### LOW Priority
11. **Update load tests** - May need path updates post-migration
12. **Add TLS secrets** - If template requires them

---

## Cross-References

- [02-service-template-analysis.md](02-service-template-analysis.md) - Service template analysis
- [03-02.testing.instructions.md](../../.github/instructions/03-02.testing.instructions.md) - Testing standards
- [03-04.database.instructions.md](../../.github/instructions/03-04.database.instructions.md) - Database patterns
- [03-08.server-builder.instructions.md](../../.github/instructions/03-08.server-builder.instructions.md) - Builder pattern

---

## Appendix: Line Counts

```
internal/kms/server/application/application_listener.go: 1224 lines
internal/kms/server/businesslogic/businesslogic.go: 742 lines
internal/kms/server/repository/sqlrepository/sql_provider.go: 321 lines
internal/kms/server/repository/orm/business_entities.go: ~400 lines
internal/kms/server/repository/orm/barrier_entities.go: ~300 lines
```

**Total KMS main code**: ~5000 lines
**Total KMS test code**: ~8000 lines
**Estimated post-migration main code**: ~2500 lines (50% reduction)
