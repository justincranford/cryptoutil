# Coverage Analysis - Post-Consolidation

**Date**: December 7, 2025
**Context**: Coverage analysis after consolidating iteration files
**Status**: ✅ ANALYSIS COMPLETE - Gaps identified, strategies defined

---

## Overall Coverage Status

| Category | Target | Current Status | Gap | Priority |
|----------|--------|----------------|-----|----------|
| Production Code | ≥95% | Varies by package | See breakdown | HIGH |
| Infrastructure (cicd) | ≥98% | ~90% | -10% | MEDIUM |
| Utility Code | ≥98% | ~100% | 0% | ✅ Met |

---

## Critical Coverage Gaps (Below 95%)

### Package: `internal/ca/handler` (47.2%)

**Gap**: -47.8 percentage points
**Priority**: CRITICAL
**Root Cause**: Missing tests for CA REST API handlers

**Uncovered Areas**:

1. Certificate issuance handlers (`POST /ca/v1/certificate`)
2. Certificate revocation handlers (`POST /ca/v1/certificate/{serial}/revoke`)
3. CRL download handlers (`GET /ca/v1/ca/{ca_id}/crl`)
4. OCSP responder handlers (`POST /ca/v1/ocsp`)
5. EST protocol handlers (cacerts, simpleenroll, simplereenroll)
6. Profile management handlers (`GET /ca/v1/profiles`)

**Strategy**:

- Create `internal/ca/handler/*_test.go` files for each handler
- Use table-driven tests with `t.Parallel()`
- Test happy paths AND error paths
- Mock certificate repository and crypto operations
- Target: 95%+ coverage (gain ~48 percentage points)

**Effort**: 1-2 hours

---

### Package: `internal/identity/auth/userauth` (42.6%)

**Gap**: -52.4 percentage points
**Priority**: CRITICAL
**Root Cause**: Missing tests for authentication flows

**Uncovered Areas**:

1. Password validation logic
2. MFA enrollment flows
3. MFA verification flows (TOTP, Passkey, Email OTP)
4. Session creation and validation
5. Password reset flows
6. Account lockout mechanisms

**Strategy**:

- Create comprehensive `userauth/*_test.go` files
- Test all MFA factor types (TOTP, Passkey, Email)
- Test password policies and validation
- Test session lifecycle (create, validate, expire)
- Use UUIDv7 for test data isolation
- Target: 95%+ coverage (gain ~52 percentage points)

**Effort**: 1-2 hours

---

### Package: `internal/jose` (48.8%)

**Gap**: -46.2 percentage points
**Priority**: HIGH
**Root Cause**: JOSE primitives lack comprehensive tests

**Uncovered Areas**:

1. JWK generation for all algorithm types
2. JWS signing/verification edge cases
3. JWE encryption/decryption error paths
4. JWT claims validation logic
5. JWKS endpoint logic
6. Key rotation scenarios

**Strategy**:

- **FIRST**: Increase coverage to 95%+ (prerequisite for performance optimization)
- Add tests for all supported algorithms (RSA, ECDSA, EdDSA, AES)
- Test error handling (invalid keys, malformed tokens)
- Test key rotation and versioning
- **THEN**: Apply performance optimizations (currently 67s execution)
- Target: 95%+ coverage (gain ~46 percentage points)

**Effort**: 2-3 hours

---

### Package: `internal/identity/idp` (54.9%)

**Gap**: -40.1 percentage points
**Priority**: MEDIUM
**Root Cause**: Missing tests for IdP handlers

**Uncovered Areas**:

1. Login handler UI rendering
2. Consent handler UI rendering
3. Logout flow
4. End session (RP-initiated logout)
5. UserInfo endpoint claims logic
6. Error page rendering

**Strategy**:

- Add handler tests for login, consent, logout
- Test HTML form rendering (use Fiber test utilities)
- Test session management
- Test claims mapping for UserInfo endpoint
- Reduce database setup time (use in-memory SQLite)
- Target: 95%+ coverage (gain ~40 percentage points)

**Effort**: 1-2 hours

---

### Package: `internal/jose/server` (56.1%)

**Gap**: -38.9 percentage points
**Priority**: HIGH
**Root Cause**: Missing E2E tests for JOSE Authority API

**Uncovered Areas**:

1. JWK generation endpoint (`POST /jose/v1/jwk/generate`)
2. JWK retrieval endpoints (`GET /jose/v1/jwk/{kid}`, `GET /jose/v1/jwk`)
3. JWKS endpoints (`GET /jose/v1/jwks`, `GET /.well-known/jwks.json`)
4. JWS sign/verify endpoints
5. JWE encrypt/decrypt endpoints
6. JWT issue/validate endpoints

**Strategy**:

- Create comprehensive E2E test suite (Task P2.1)
- Test all 10 JOSE API endpoints
- Use Docker Compose for integration testing
- Test error responses (400, 401, 500)
- Target: 95%+ coverage (gain ~39 percentage points)

**Effort**: 3-4 hours (part of Phase 2 task P2.1)

---

## Packages Close to Target (≥80%, <95%)

### Package: `internal/kms/server/barrier` (75.5%)

**Gap**: -19.5 percentage points
**Priority**: MEDIUM

**Strategy**: Add tests for crypto barrier edge cases (key unsealing failures, rotation scenarios)
**Effort**: 30 minutes

---

### Package: `internal/kms/client` (76.2%)

**Gap**: -18.8 percentage points
**Priority**: MEDIUM

**Strategy**: Add tests for client error handling, connection failures, retry logic
**Effort**: 30 minutes (combined with performance optimization P0.3)

---

### Package: `internal/identity/authz` (77.2%)

**Gap**: -17.8 percentage points
**Priority**: MEDIUM

**Strategy**: Add tests for OAuth 2.1 error flows (invalid_grant, invalid_client, etc.)
**Effort**: 30 minutes

---

### Package: `internal/identity/authz/clientauth` (78.4%)

**Gap**: -16.6 percentage points
**Priority**: MEDIUM

**Strategy**: Add tests for advanced client authentication (private_key_jwt, mTLS)
**Effort**: 30 minutes (combined with performance optimization P0.1)

---

### Package: `internal/kms/server/unsealkeysservice` (78.2%)

**Gap**: -16.8 percentage points
**Priority**: MEDIUM

**Strategy**: Add tests for unseal key derivation edge cases, rotation scenarios
**Effort**: 30 minutes (Task P3.3)

---

### Package: `internal/infra/realm` (85.6%)

**Gap**: -9.4 percentage points
**Priority**: LOW

**Strategy**: Add tests for realm configuration validation, edge cases
**Effort**: 20 minutes

---

### Package: `internal/identity/rotation` (83.7%)

**Gap**: -11.3 percentage points
**Priority**: LOW

**Strategy**: Add tests for key rotation edge cases, grace period scenarios
**Effort**: 20 minutes

---

### Package: `internal/identity/jobs` (89.0%)

**Gap**: -6.0 percentage points
**Priority**: LOW

**Strategy**: Add tests for background job error handling
**Effort**: 15 minutes

---

### Package: `internal/network` (88.7%)

**Gap**: -6.3 percentage points
**Priority**: LOW

**Strategy**: Add tests for network failure scenarios (Task P3.4)
**Effort**: 30 minutes

---

## Packages Exceeding Target (≥95%)

### Package: `internal/apperr` (96.6%) ✅

**Status**: ✅ Exceeds target by 1.6 percentage points
**Action**: No action required

---

### Package: `internal/common/crypto/keygen` (85.2%)

**Gap**: -9.8 percentage points
**Priority**: LOW

**Strategy**: Add tests for keygen pool edge cases, concurrent generation scenarios
**Effort**: 30 minutes

---

## Infrastructure Coverage (cicd) (~90%)

**Target**: ≥98%
**Gap**: -10 percentage points
**Priority**: MEDIUM

**Uncovered Areas**:

1. CICD command error handling
2. Edge case scenarios for automation commands
3. Cross-platform compatibility tests

**Strategy**:

- Add tests for all cicd commands in `internal/cmd/cicd/`
- Test error paths and edge cases
- Test cross-platform behavior (Windows, Linux, macOS)
- Target: 98%+ coverage

**Effort**: 1-2 hours

---

## Coverage Improvement Roadmap

### Phase 1: Critical Gaps (4-6 hours)

| Package | Current | Target | Effort |
|---------|---------|--------|--------|
| ca/handler | 47.2% | 95% | 1-2h |
| auth/userauth | 42.6% | 95% | 1-2h |
| jose | 48.8% | 95% | 2-3h |

**Total Effort**: 4-7 hours
**Impact**: 3 critical packages reach target

---

### Phase 2: High Priority (3-4 hours)

| Package | Current | Target | Effort |
|---------|---------|--------|--------|
| identity/idp | 54.9% | 95% | 1-2h |
| jose/server | 56.1% | 95% | 3-4h (E2E tests) |

**Total Effort**: 4-6 hours
**Impact**: 2 high-priority packages reach target

---

### Phase 3: Medium Priority (2-3 hours)

| Package | Current | Target | Effort |
|---------|---------|--------|--------|
| kms/server/barrier | 75.5% | 95% | 30min |
| kms/client | 76.2% | 95% | 30min |
| identity/authz | 77.2% | 95% | 30min |
| clientauth | 78.4% | 95% | 30min |
| unsealkeysservice | 78.2% | 95% | 30min |
| network | 88.7% | 95% | 30min |

**Total Effort**: 3 hours
**Impact**: 6 packages reach target

---

### Phase 4: Infrastructure (1-2 hours)

| Package | Current | Target | Effort |
|---------|---------|--------|--------|
| cicd commands | ~90% | 100% | 1-2h |

**Total Effort**: 1-2 hours
**Impact**: Infrastructure coverage reaches 98%

---

## Total Coverage Improvement Estimate

**Current State**: 11 packages below 95% target
**Target State**: All packages ≥95% coverage
**Total Effort**: 12-18 hours (can be parallelized across phases)

**Recommended Approach**: Focus on Phase 1 (critical gaps) in Phase 3 of implementation plan (Day 5, 2-3 hours)

---

## Mutation Testing Analysis

**Current State**: Unknown baseline (gremlins not yet run project-wide)
**Target**: ≥80% mutation score per package
**Recommended**: Run `gremlins unleash --tags=!integration` to establish baseline

**Packages Expected to Need Work**:

- Packages with low coverage (ca/handler, auth/userauth, jose) likely have low mutation scores
- Packages with high coverage but simple tests may have deceptively low mutation scores

**Effort**: 2-4 hours to improve mutation scores after baseline established

---

## Conclusion

**Coverage Analysis Status**: ✅ **COMPLETE**

**Key Findings**:

1. **11 packages below 95% target** (critical: 3, high: 2, medium: 6)
2. **Total gap**: ~400 percentage points across all packages
3. **Estimated effort**: 12-18 hours to reach 95%+ on all packages
4. **Priority order**: ca/handler → auth/userauth → jose → jose/server → others

**Recommendation**: Execute Phase 3 (Coverage Targets) from implementation plan to address critical gaps.

**Next Step**: Begin implementation (Phase 0: Slow Test Optimization).

---

*Coverage Analysis Version: 1.0.0*
*Analyst: GitHub Copilot (Agent)*
*Approved: Pending user validation*
