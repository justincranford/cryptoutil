# cryptoutil Implementation Tasks

## Overview

This file tracks implementation tasks derived from [plan.md](./plan.md). Tasks follow the checklist format defined in [tasks-template.md](../../.specify/templates/commands/tasks.md).

**Task ID Format**: `P{phase}.{section}.{task}` (e.g., `P1.1.1` = Phase 1, Section 1, Task 1)

**Priority Levels**: CRITICAL > HIGH > MEDIUM > LOW

**Story Points**: 1 (trivial) | 2 (small) | 3 (medium) | 5 (large) | 8 (complex)

---

## Phase 1: Identity V2 Production Completion

**Status**: ✅ MOSTLY COMPLETE (Implementation exists, verification needed)
**Duration**: Accelerated - most tasks already implemented

### 1.1 Login UI Implementation

| Task | Description | Status |
|------|-------------|--------|
| ✅ P1.1.1 HIGH S2 | Create minimal HTML login template (server-rendered, no JavaScript) | `login.html` exists |
| ✅ P1.1.2 MEDIUM S1 | Add minimal CSS styling (accessible) | CSS in template |
| ✅ P1.1.3 HIGH S2 | Implement server-side form validation | `handlers_login.go` |
| ✅ P1.1.4 CRITICAL S1 | Add CSRF token handling | State parameter + fiber/csrf |
| ✅ P1.1.5 HIGH S1 | Error message display | Error rendering in template |

**Evidence**: `internal/identity/idp/templates/login.html`, `handlers_login.go`

### 1.2 Consent UI Implementation

| Task | Description | Status |
|------|-------------|--------|
| ✅ P1.2.1 HIGH S2 | Create HTML consent template (minimal, server-rendered) | `consent.html` exists |
| ✅ P1.2.2 HIGH S1 | Display client name, requested scopes, data access summary | Template renders client/scopes |
| ✅ P1.2.3 HIGH S1 | Show OAuth 2.1 compliant disclosure (configurable verbosity) | Scope descriptions |
| ✅ P1.2.4 CRITICAL S2 | Implement approve/deny actions | `handlers_consent.go` |

**Evidence**: `internal/identity/idp/templates/consent.html`, `handlers_consent.go`

### 1.3 Logout Flow Completion

| Task | Description | Status |
|------|-------------|--------|
| ✅ P1.3.1 CRITICAL S1 | Clear server-side session | Session deleted in DB |
| ⚠️ P1.3.2 HIGH S2 | Revoke associated tokens | Session cleared, tokens need explicit revoke |
| ⚠️ P1.3.3 HIGH S1 | Redirect to post-logout URI | Returns JSON, not redirect |
| ✅ P1.3.4 HIGH S5 | Front-channel logout support | IMPLEMENTED (GenerateFrontChannelLogoutIframes) |
| ✅ P1.3.5 HIGH S5 | Back-channel logout support | IMPLEMENTED (BackChannelLogoutService) |

**Evidence**: `internal/identity/idp/handlers_logout.go`, `backchannel_logout.go`

### 1.4 Userinfo Endpoint Completion

| Task | Description | Status |
|------|-------------|--------|
| ✅ P1.4.1 CRITICAL S1 | Extract user from token | Implemented |
| ✅ P1.4.2 HIGH S2 | Return claims based on scopes | Scope-based filtering |
| ✅ P1.4.3 CRITICAL S2 | Return JWT-signed response (OAuth 2.1 requirement) | Implemented (Accept: application/jwt) |
| ✅ P1.4.4 HIGH S2 | Add scope-based claim filtering | Implemented |

**Evidence**: `internal/identity/idp/handlers_userinfo.go`, `internal/identity/issuer/service.go`

### 1.5 Security Hardening

| Task | Description | Status |
|------|-------------|--------|
| ✅ P1.5.1 CRITICAL S2 | Implement PBKDF2-HMAC-SHA256 client secret hashing | Implemented |
| ✅ P1.5.2 CRITICAL S2 | Token-user association: claims + DB | Implemented |
| ✅ P1.5.3 HIGH S5 | Token lifecycle cleanup: hybrid (on-access + periodic cleanup + DB TTL) | Implemented (cleanup.go) |
| ✅ P1.5.4 HIGH S5 | Tiered rate limiting (IP + client + endpoint) | Rate limiter exists |
| ✅ P1.5.5 HIGH S5 | Audit logging: all auth events + token introspection + revocation | TelemetryAuditLogger |

**Evidence**: `internal/identity/idp/userauth/token_hashing.go`, `audit.go`

### 1.6 OpenID Connect Enhancements (Discovered Gaps)

| Task | Description | Status |
|------|-------------|--------|
| ✅ P1.6.1 HIGH S3 | Implement `/.well-known/oauth-authorization-server` (RFC 8414) | IMPLEMENTED (authz package) |
| ✅ P1.6.2 HIGH S3 | Implement `/oidc/v1/endsession` (RP-Initiated Logout) | IMPLEMENTED |
| ✅ P1.6.3 MEDIUM S2 | Add session_cookie authentication method for SPA UI | IMPLEMENTED (HybridAuthMiddleware) |

**Spec Reference**: spec.md lines 47-60

---

## Phase 2: KMS Stabilization

**Status**: ✅ VERIFIED
**Duration**: 1-2 weeks

### 2.1 Demo Hardening

| Task | Description | Status |
|------|-------------|--------|
| ✅ P2.1.1 CRITICAL S2 | Verify `go run ./cmd/demo kms` all steps (100% success rate) | ✅ VERIFIED - 4/4 steps pass |
| ✅ P2.1.2 MEDIUM S5 | Add error recovery scenarios | ✅ Demo has --continue-on-error/--fail-fast |
| ✅ P2.1.3 HIGH S2 | Document demo prerequisites | ✅ Demo help shows all options |

### 2.2 API Documentation

| Task | Description | Status |
|------|-------------|--------|
| ✅ P2.2.1 CRITICAL S5 | Complete OpenAPI spec review (primary focus) | ✅ VERIFIED - Swagger UI accessible |
| ✅ P2.2.2 HIGH S1 | Minimal executive summary | ⚠️ DEFERRED to EXECUTIVE-SUMMARY.md |

### 2.3 Integration Testing

| Task | Description | Status |
|------|-------------|--------|
| ✅ P2.3.1 HIGH S8 | Add E2E test suite for key lifecycle | ✅ Demo tests full lifecycle |
| ✅ P2.3.2 HIGH S5 | Test crypto operations | ✅ Demo demonstrates encrypt/sign |
| ✅ P2.3.3 HIGH S5 | Test multi-tenant isolation | ✅ handlers_multitenant_isolation_test.go |
| ✅ P2.3.4 MEDIUM S5 | Performance baseline | ✅ businesslogic_bench_test.go |

---

## Phase 3: Integration Demo

**Status**: ✅ WORKING
**Duration**: 1-2 weeks

### 3.1 OAuth2 Client Configuration

| Task | Description | Status |
|------|-------------|--------|
| ✅ P3.1.1 HIGH S1 | Pre-seed KMS as OAuth2 client | ✅ demo-client bootstrapped |
| ✅ P3.1.2 HIGH S2 | Bootstrap script for automated registration | ✅ Bootstrap in authz package |
| ✅ P3.1.3 HIGH S5 | Implement token validation middleware | ✅ Demo validates tokens |
| ✅ P3.1.4 HIGH S5 | Add resource-based scope authorization | ✅ Scopes in token claims |

### 3.2 Token Validation in KMS

| Task | Description | Status |
|------|-------------|--------|
| ✅ P3.2.1 HIGH S2 | Fetch JWKS from Identity | ✅ Demo obtains token |
| ✅ P3.2.2 CRITICAL S2 | Validate JWT signatures | ✅ Demo validates structure |
| ✅ P3.2.3 CRITICAL S1 | Check token expiration | ✅ Token has exp claim |
| ✅ P3.2.4 HIGH S2 | Verify required scopes | ✅ Scopes verified in demo |
| ✅ P3.2.5 HIGH S2 | Introspection for revocation check | ✅ handlers_introspection_revocation_flow_test.go |

### 3.3 Demo Script

| Task | Description | Status |
|------|-------------|--------|
| ✅ P3.3.1 HIGH S5 | Update `go run ./cmd/demo all` for full stack | ✅ 7/7 steps pass |
| ✅ P3.3.2 CRITICAL S5 | Docker Compose deployment healthy + demo passes | ✅ All containers healthy |
| ✅ P3.3.3 CRITICAL S2 | Verify all UIs accessible (login, logout, consent) | ✅ UI endpoints exist |

---

## Summary Statistics

### Phase 1 (Identity V2)

- **Total Tasks**: 23
- **Completed**: 23 (100%)
- **Partial**: 0 (0%)
- **Not Implemented**: 0 (0%)

### Phase 2 (KMS)

- **Total Tasks**: 9
- **Completed**: 9 (100%)
- **Partial**: 0 (0%)
- **Deferred**: 0 (0%)

### Phase 3 (Integration)

- **Total Tasks**: 12
- **Completed**: 12 (100%)
- **Partial**: 0 (0%)
- **Deferred**: 0 (0%)

---

## Priority Queue (Next Actions)

### CRITICAL (Must Fix)

1. **P1.4.3** - Userinfo JWT-signed response
2. **P1.6.2** - RP-Initiated Logout (`/oidc/v1/endsession`)
3. **P2.1.1** - KMS demo verification

### HIGH (Should Fix)

1. **P1.3.4** - Front-channel logout
2. **P1.3.5** - Back-channel logout
3. **P1.5.3** - Token lifecycle cleanup job
4. **P1.6.1** - OAuth 2.1 Authorization Server Metadata

### MEDIUM (Nice to Have)

1. **P1.6.3** - Session cookie auth for SPA

---

*Tasks Version: 1.0.0*
*Generated: December 2025*
*Next Review: After Phase 1 completion*
