# Specification Coverage Analysis

## Purpose

This document validates requirement-to-task coverage as part of `/speckit.analyze`.

**Date**: December 4, 2025
**Iteration**: 1 (completing skipped analyze step)

---

## Coverage Matrix: P2 Identity (OAuth 2.1 + OIDC)

### Authorization Server (AuthZ) Endpoints

| Requirement | Task ID | Status | Evidence |
|-------------|---------|--------|----------|
| `/oauth2/v1/authorize` | P1.1.x, P1.2.x | ✅ | handlers_authorize.go |
| `/oauth2/v1/token` | P1.3.x | ✅ | handlers_token.go |
| `/oauth2/v1/introspect` | P1.4.x | ✅ | handlers_introspection_*.go |
| `/oauth2/v1/revoke` | P1.4.x | ✅ | handlers_revocation.go |
| `/oauth2/v1/clients/{id}/rotate-secret` | P1.5.x | ✅ | rotation/ package |
| `/.well-known/openid-configuration` | P1.6.1 | ✅ | handlers_discovery.go |
| `/.well-known/jwks.json` | P1.6.1 | ✅ | handlers_jwks.go |
| `/.well-known/oauth-authorization-server` | P1.6.1 | ✅ | handlers_oauth_metadata.go |

### Identity Provider (IdP) Endpoints

| Requirement | Task ID | Status | Evidence |
|-------------|---------|--------|----------|
| `/oidc/v1/login` | P1.1.x | ✅ | handlers_login.go |
| `/oidc/v1/consent` | P1.2.x | ✅ | handlers_consent.go |
| `/oidc/v1/logout` | P1.3.x | ✅ | handlers_logout.go |
| `/oidc/v1/endsession` | P1.6.2 | ✅ | handlers_endsession.go |
| `/oidc/v1/userinfo` | P1.4.x | ✅ | handlers_userinfo.go |

### Authentication Methods

| Requirement | Task ID | Status | Gap |
|-------------|---------|--------|-----|
| client_secret_basic | P1.5.x | ✅ | None |
| client_secret_post | P1.5.x | ✅ | None |
| client_secret_jwt | **MISSING** | ⚠️ 70% | Missing: jti replay, lifetime validation |
| private_key_jwt | **MISSING** | ⚠️ 50% | Missing: client JWKS registration |
| session_cookie | P1.6.3 | ✅ | HybridAuthMiddleware |

### MFA Factors

| Requirement | Task ID | Status | Gap |
|-------------|---------|--------|-----|
| Passkey/WebAuthn | P1.5.x | ✅ | None |
| TOTP | P1.5.x | ✅ | None |
| Hardware Security Keys | **MISSING** | ❌ | No task defined |
| Email OTP | **MISSING** | ⚠️ 30% | No delivery service |
| SMS OTP | **MISSING** | ⚠️ 20% | No delivery service |
| Recovery Codes | **MISSING** | ❌ | No task defined |

---

## Coverage Matrix: P3 KMS

### ElasticKey Operations

| Requirement | Task ID | Status | Evidence |
|-------------|---------|--------|----------|
| Create ElasticKey | P2.x | ✅ | handlers_elastickey.go |
| Read ElasticKey | P2.x | ✅ | handlers_elastickey.go |
| List ElasticKeys | P2.x | ✅ | handlers_elastickey.go |
| Update ElasticKey | **MISSING** | ❌ | Not in spec as required |
| Delete ElasticKey | **MISSING** | ❌ | Not in spec as required |

### MaterialKey Operations

| Requirement | Task ID | Status | Evidence |
|-------------|---------|--------|----------|
| Create MaterialKey | P2.x | ✅ | handlers_materialkey.go |
| Read MaterialKey | P2.x | ✅ | handlers_materialkey.go |
| List MaterialKeys | P2.x | ✅ | handlers_materialkey.go |
| Global List | P2.x | ✅ | handlers_materialkey.go |
| Import | **MISSING** | ❌ | Not in spec as required |
| Revoke | **MISSING** | ❌ | Not in spec as required |

### Cryptographic Operations

| Requirement | Task ID | Status | Evidence |
|-------------|---------|--------|----------|
| Generate | P2.x | ✅ | handlers_crypto.go |
| Encrypt | P2.x | ✅ | handlers_crypto.go |
| Decrypt | P2.x | ✅ | handlers_crypto.go |
| Sign | P2.x | ✅ | handlers_crypto.go |
| Verify | P2.x | ✅ | handlers_crypto.go |

---

## Coverage Matrix: P4 Certificates (CA)

| Requirement | Task ID | Status | Evidence |
|-------------|---------|--------|----------|
| Task 1-20 | P4.x | ✅ | internal/ca/ package |
| CA Server API | **MISSING** | ❌ | Code exists but no REST server |

---

## Identified Gaps

### GAP-1: client_secret_jwt Missing Tasks

**Severity**: HIGH
**Impact**: OAuth 2.1 compliance incomplete

**Required Tasks**:
- P1.7.1: Implement jti claim uniqueness verification (replay protection)
- P1.7.2: Implement assertion lifetime validation (exp, iat, nbf)
- P1.7.3: Add test coverage for client_secret_jwt

### GAP-2: private_key_jwt Missing Tasks

**Severity**: HIGH
**Impact**: OAuth 2.1 compliance incomplete

**Required Tasks**:
- P1.8.1: Implement client JWKS registration endpoint
- P1.8.2: Implement client public key storage
- P1.8.3: Implement jti claim uniqueness verification
- P1.8.4: Implement kid header matching to client keys
- P1.8.5: Add test coverage for private_key_jwt

### GAP-3: Email/SMS OTP Missing Tasks

**Severity**: MEDIUM
**Impact**: MFA options limited

**Required Tasks**:
- P1.9.1: Design email/SMS provider abstraction interface
- P1.9.2: Implement email delivery service (SMTP adapter)
- P1.9.3: Implement SMS delivery service (Twilio adapter stub)
- P1.9.4: Implement OTP storage with expiration
- P1.9.5: Implement rate limiting per contact method

### GAP-4: Hardware Security Keys Missing Tasks

**Severity**: HIGH
**Impact**: FIDO2 compliance incomplete

**Required Tasks**:
- P1.10.1: Research U2F/FIDO hardware key support
- P1.10.2: Implement hardware key registration
- P1.10.3: Implement hardware key authentication
- P1.10.4: Add test coverage with mock hardware

### GAP-5: Recovery Codes Missing Tasks

**Severity**: MEDIUM
**Impact**: Account recovery limited

**Required Tasks**:
- P1.11.1: Implement recovery code generation
- P1.11.2: Implement recovery code storage (hashed)
- P1.11.3: Implement recovery code validation (single-use)
- P1.11.4: Add test coverage

### GAP-6: CA Server Missing Tasks

**Severity**: HIGH
**Impact**: P4 not deployable standalone

**Required Tasks**:
- P4.21: Create CA REST API server package
- P4.22: Implement CA server main function
- P4.23: Create CA Docker Compose deployment
- P4.24: Add CA E2E tests

### GAP-7: JOSE Authority Refactor Missing Tasks

**Severity**: HIGH
**Impact**: P1 not deployable standalone

**Required Tasks**:
- P1.12.1: Refactor internal/common/crypto/jose → internal/jose
- P1.12.2: Create JOSE REST API server package
- P1.12.3: Implement JOSE server main function
- P1.12.4: Create JOSE Docker Compose deployment
- P1.12.5: Add JOSE E2E tests

---

## Test Parallelism Issues Identified

### Issue 1: Port Conflicts

**Location**: `internal/identity/integration/integration_test.go`
**Problem**: Hardcoded ports 18080, 18081, 18082
**Impact**: Tests fail when run in parallel with other test packages

**Fix Required**:
- Use dynamic port allocation (port 0)
- Extract actual assigned port from listener

### Issue 2: Database Connection Closure

**Location**: Multiple authz test files
**Problem**: Database connections being closed while other tests still using them
**Impact**: "sql: database is closed" errors in parallel execution

**Fix Required**:
- Ensure each test creates its own database connection
- Use unique database names per test (already using UUIDs, but shared cache may be issue)
- Check for premature Close() calls in defer blocks

### Issue 3: Test Mutex Serialization

**Location**: `internal/identity/integration/integration_test.go`
**Problem**: `testMutex` forces serial execution, hiding parallelism issues
**Impact**: Tests pass individually but fail in bulk CI runs

**Fix Required**:
- Remove testMutex after fixing port allocation
- Convert to proper parallel execution

---

## Summary Statistics

| Category | Total | Covered | Gaps | Coverage |
|----------|-------|---------|------|----------|
| P2 Identity AuthZ | 8 | 8 | 0 | 100% |
| P2 Identity IdP | 5 | 5 | 0 | 100% |
| P2 Auth Methods | 7 | 4 | 3 | 57% |
| P2 MFA Factors | 9 | 2 | 7 | 22% |
| P3 KMS Operations | 15 | 11 | 4 | 73% |
| P4 CA Server | 1 | 0 | 1 | 0% |
| **Overall** | 45 | 30 | 15 | **67%** |

---

## Recommendations

1. **Iteration 1 Completion**: Fix test parallelism issues before marking complete
2. **Iteration 2 Priority**: JOSE Authority refactor + CA Server (enable 4-product deployment)
3. **Iteration 3 Priority**: client_secret_jwt/private_key_jwt completion (OAuth 2.1 compliance)
4. **Iteration 4 Priority**: Hardware Security Keys + Recovery Codes (MFA completion)
5. **Future**: Email/SMS OTP (lower priority, NIST deprecates SMS)

---

*Analysis Version: 1.0.0*
*Generated By: /speckit.analyze*
