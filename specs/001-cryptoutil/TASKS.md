# Task Breakdown - Post-Consolidation

**Date**: December 7, 2025
**Context**: Detailed task breakdown after consolidating iteration files
**Status**: ✅ 36 tasks identified (27 required, 9 optional)

---

## Task Summary

| Phase | Required Tasks | Optional Tasks | Total | Effort |
|-------|----------------|----------------|-------|--------|
| Phase 0: Slow Test Optimization | 5 | 0 | 5 | 4-5h |
| Phase 1: CI/CD Workflows | 8 | 0 | 8 | 4-5h |
| Phase 2: Deferred I2 Features | 7 | 1 | 8 | 6-8h |
| Phase 3: Coverage Targets | 4 | 1 | 5 | 2-3h |
| Phase 4: Advanced Testing | 0 | 4 | 4 | 4-6h |
| Phase 5: Documentation | 0 | 6 | 6 | 8-12h |
| **Total** | **27** | **9** | **36** | **28-39h** |

---

## Phase 0: Optimize Slow Test Packages (5 tasks, 4-5h)

### P0.1: Optimize clientauth Package (168s → <30s)

**Priority**: CRITICAL
**Effort**: 2 hours
**Status**: ❌ Not Started

**Acceptance Criteria**:
- Apply aggressive `t.Parallel()` to all test cases
- Split into multiple test files by auth method (basic, post, jwt, private_key_jwt)
- Implement selective execution pattern for local dev
- Execution time <30s
- No test failures
- Coverage maintained at 78.4% or higher

**Files to Modify**:
- `internal/identity/authz/clientauth/*_test.go`

---

### P0.2: Optimize jose/server Package (94s → <20s)

**Priority**: CRITICAL
**Effort**: 1 hour
**Status**: ❌ Not Started

**Acceptance Criteria**:
- Implement parallel subtests
- Reduce Fiber app setup/teardown overhead (shared test server instance)
- Execution time <20s
- Coverage improved from 56.1% → 70%+

**Files to Modify**:
- `internal/jose/server/*_test.go`

---

### P0.3: Optimize kms/client Package (74s → <20s)

**Priority**: CRITICAL
**Effort**: 1 hour
**Status**: ❌ Not Started

**Acceptance Criteria**:
- Mock KMS server dependency
- Implement parallel test execution
- Reduce network roundtrip simulation
- Execution time <20s
- Coverage maintained at 76.2% or higher

**Files to Modify**:
- `internal/kms/client/*_test.go`

---

### P0.4: Optimize jose Package (67s → <15s)

**Priority**: HIGH
**Effort**: 30 minutes
**Status**: ❌ Not Started

**Acceptance Criteria**:
- Increase coverage 48.8% → 70%+ FIRST
- Then apply parallel execution
- Reduce cryptographic operation redundancy
- Execution time <15s

**Files to Modify**:
- `internal/jose/*_test.go`

---

### P0.5: Optimize kms/server/application Package (28s → <10s)

**Priority**: MEDIUM
**Effort**: 30 minutes
**Status**: ❌ Not Started

**Acceptance Criteria**:
- Implement parallel server tests
- Use dynamic port allocation pattern
- Reduce test server setup/teardown overhead
- Execution time <10s
- Coverage maintained at 64.7% or higher

**Files to Modify**:
- `internal/kms/server/application/*_test.go`

---

## Phase 1: Fix CI/CD Workflows (8 tasks, 4-5h)

### P1.1-P1.8: Fix Individual Workflows

**Common Pattern**:
1. Run workflow locally with Act: `go run ./cmd/workflow -workflows=<name>`
2. Identify failure root cause
3. Implement fix
4. Verify fix locally
5. Commit and verify in GitHub Actions

| Task | Workflow | Root Cause | Effort | Status |
|------|----------|------------|--------|--------|
| P1.1 | ci-dast | Service connectivity | 1h | ❌ |
| P1.2 | ci-e2e | Docker Compose setup | 1h | ❌ |
| P1.3 | ci-load | Gatling configuration | 30min | ❌ |
| P1.4 | ci-coverage | Coverage aggregation | 30min | ❌ |
| P1.5 | ci-race | Race conditions | 1h | ❌ |
| P1.6 | ci-benchmark | Benchmark baselines | 30min | ❌ |
| P1.7 | ci-fuzz | Fuzz test execution | 30min | ❌ |
| P1.8 | (TBD) | TBD | 30min | ❌ |

---

## Phase 2: Complete Deferred I2 Features (8 tasks, 6-8h)

### P2.1: JOSE E2E Test Suite

**Priority**: HIGH
**Effort**: 3-4 hours
**Status**: ❌ Not Started

**Acceptance Criteria**:
- Create `internal/jose/server/*_integration_test.go`
- Test all 10 JOSE API endpoints end-to-end
- Integration with Docker Compose
- Tests execute in <2 minutes
- Coverage >95% for JOSE server package

**Files to Create**:
- `internal/jose/server/jwk_integration_test.go`
- `internal/jose/server/jws_integration_test.go`
- `internal/jose/server/jwe_integration_test.go`
- `internal/jose/server/jwt_integration_test.go`

---

### P2.2: CA OCSP Responder

**Priority**: HIGH
**Effort**: 2 hours
**Status**: ❌ Not Started

**Acceptance Criteria**:
- Implement `/ca/v1/ocsp` endpoint
- RFC 6960 OCSP protocol support
- Return certificate status (good, revoked, unknown)
- Integration with CRL generation

**Files to Modify**:
- `internal/ca/handler/ocsp.go` (create)
- `internal/ca/server/routes.go`

---

### P2.3: JOSE Docker Integration

**Priority**: HIGH
**Effort**: 1-2 hours
**Status**: ❌ Not Started

**Acceptance Criteria**:
- Add jose-sqlite, jose-postgres-1, jose-postgres-2 services
- Configure ports 8080-8082 (public), 9090 (admin)
- Health checks via wget
- Services start and pass health checks

**Files to Modify**:
- `deployments/compose/compose.yml`
- `deployments/jose/docker-compose.yml` (create)
- `configs/jose/*.yml`

---

### P2.4-P2.7: Already Complete ✅

| Task | Feature | Status |
|------|---------|--------|
| P2.4 | EST cacerts | ✅ Complete |
| P2.5 | EST simpleenroll | ✅ Complete |
| P2.6 | EST simplereenroll | ✅ Complete |
| P2.7 | TSA timestamp | ✅ Complete |

---

### P2.8: EST serverkeygen (OPTIONAL)

**Priority**: LOW
**Effort**: 3-4 hours (if PKCS#7 library resolved)
**Status**: ⚠️ BLOCKED on PKCS#7 library

**Acceptance Criteria**:
- Integrate `go.mozilla.org/pkcs7` library
- Implement server-side key generation
- Return PKCS#7/CMS envelope
- Full RFC 7030 compliance

**Files to Modify**:
- `internal/ca/handler/est_serverkeygen.go`
- `go.mod` (add pkcs7 dependency)

---

## Phase 3: Achieve Coverage Targets (5 tasks, 2-3h)

### P3.1: ca/handler Coverage (47.2% → 95%)

**Priority**: CRITICAL
**Effort**: 1 hour
**Status**: ❌ Not Started

**Acceptance Criteria**:
- Add tests for all CA handler endpoints
- Test happy paths and error paths
- Use table-driven tests with `t.Parallel()`
- Coverage ≥95%

**Files to Create/Modify**:
- `internal/ca/handler/*_test.go`

---

### P3.2: auth/userauth Coverage (42.6% → 95%)

**Priority**: CRITICAL
**Effort**: 1 hour
**Status**: ❌ Not Started

**Acceptance Criteria**:
- Add authentication flow tests
- Test MFA flows, password validation, session management
- Coverage ≥95%

**Files to Create/Modify**:
- `internal/identity/auth/userauth/*_test.go`

---

### P3.3: unsealkeysservice Coverage (78.2% → 95%)

**Priority**: MEDIUM
**Effort**: 30 minutes
**Status**: ❌ Not Started

**Acceptance Criteria**:
- Add edge case tests
- Test error handling
- Coverage ≥95%

---

### P3.4: network Coverage (88.7% → 95%)

**Priority**: MEDIUM
**Effort**: 30 minutes
**Status**: ❌ Not Started

**Acceptance Criteria**:
- Add error path tests
- Test network failure scenarios
- Coverage ≥95%

---

### P3.5: Verify apperr Coverage (96.6%)

**Priority**: LOW
**Effort**: 5 minutes
**Status**: ✅ Already Complete

**Acceptance Criteria**:
- Verify current coverage ≥95% (already at 96.6%)
- No action required

---

## Phase 4: Advanced Testing (4 tasks, 4-6h, OPTIONAL)

### P4.1: Add Benchmark Tests

**Effort**: 2 hours
**Status**: ❌ Not Started

**Files to Create**:
- `internal/common/crypto/keygen/*_bench_test.go`
- `internal/jose/*_bench_test.go`
- `internal/ca/crypto/*_bench_test.go`

---

### P4.2: Add Fuzz Tests

**Effort**: 2 hours
**Status**: ❌ Not Started

**Files to Create**:
- `internal/jose/jwt_parser_fuzz_test.go`
- `internal/ca/parser/*_fuzz_test.go`

---

### P4.3: Add Property-Based Tests

**Effort**: 2 hours
**Status**: ❌ Not Started

**Files to Create**:
- `internal/common/crypto/*_property_test.go`

---

### P4.4: Mutation Testing Baseline

**Effort**: 1 hour
**Status**: ❌ Not Started

**Command**: `gremlins unleash --tags=!integration`

---

## Phase 5: Documentation & Demo (6 tasks, 8-12h, OPTIONAL)

### P5.1-P5.6: Demo Videos

| Task | Demo | Duration | Effort | Status |
|------|------|----------|--------|--------|
| P5.1 | JOSE Authority | 5-10min | 2h | ❌ |
| P5.2 | Identity Server | 10-15min | 2-3h | ❌ |
| P5.3 | KMS | 10-15min | 2-3h | ❌ |
| P5.4 | CA Server | 10-15min | 2-3h | ❌ |
| P5.5 | Integration | 15-20min | 3-4h | ❌ |
| P5.6 | Unified Suite | 20-30min | 3-4h | ❌ |

---

## Conclusion

**Task Breakdown Status**: ✅ **COMPLETE**

36 tasks identified with clear acceptance criteria, effort estimates, and priorities.

**Next Step**: Execute /speckit.analyze to perform coverage check.

---

*Task Breakdown Version: 1.0.0*
*Author: GitHub Copilot (Agent)*
*Approved: Pending user validation*
