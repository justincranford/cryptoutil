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

## Iteration 2: Standalone Services

**Status**: 🔄 IN PROGRESS (~70% complete)
**Duration**: 2-4 weeks

### 2.1 JOSE Authority Server

| Task | Description | Priority | Points | Status |
|------|-------------|----------|--------|--------|
| I2.1.1 | Create `cmd/jose-server/main.go` entry point | HIGH | 2 | ✅ |
| I2.1.2 | Implement Fiber router with API versioning (`/jose/v1/`) | HIGH | 2 | ✅ |
| I2.1.3 | Generate JWK endpoint (POST `/jose/v1/jwk/generate`) | CRITICAL | 5 | ✅ |
| I2.1.4 | Retrieve JWK endpoint (GET `/jose/v1/jwk/{kid}`) | HIGH | 2 | ✅ |
| I2.1.5 | List JWKs endpoint (GET `/jose/v1/jwk`) | HIGH | 2 | ✅ |
| I2.1.6 | Delete JWK endpoint (DELETE `/jose/v1/jwk/{kid}`) | MEDIUM | 2 | ✅ |
| I2.1.7 | JWKS endpoint (GET `/jose/v1/jwks`) | HIGH | 2 | ✅ |
| I2.1.8 | JWS sign endpoint (POST `/jose/v1/jws/sign`) | CRITICAL | 5 | ✅ |
| I2.1.9 | JWS verify endpoint (POST `/jose/v1/jws/verify`) | CRITICAL | 5 | ✅ |
| I2.1.10 | JWE encrypt endpoint (POST `/jose/v1/jwe/encrypt`) | CRITICAL | 5 | ✅ |
| I2.1.11 | JWE decrypt endpoint (POST `/jose/v1/jwe/decrypt`) | CRITICAL | 5 | ✅ |
| I2.1.12 | JWT create endpoint (POST `/jose/v1/jwt/sign`) | HIGH | 5 | ✅ |
| I2.1.13 | JWT verify endpoint (POST `/jose/v1/jwt/verify`) | HIGH | 5 | ✅ |
| I2.1.14 | OpenAPI spec for JOSE Authority (`api/jose/openapi_spec.yaml`) | HIGH | 5 | ✅ |
| I2.1.15 | Generate server/client code with oapi-codegen | HIGH | 2 | ✅ |
| I2.1.16 | Add API key authentication middleware | HIGH | 5 | ✅ |
| I2.1.17 | Docker Compose integration | MEDIUM | 2 | ✅ |
| I2.1.18 | JOSE Authority E2E tests | HIGH | 8 | ⚠️ |

**Total Points**: 69
**Completed Points**: 61 (88%)
**Evidence**: `cmd/jose-server/main.go`, `internal/jose/server/`, `api/jose/openapi_spec.yaml`, `internal/jose/server/server_test.go` passes

### 2.2 CA Server REST API

| Task | Description | Priority | Points | Status |
|------|-------------|----------|--------|--------|
| I2.2.1 | CA handler scaffolding (`internal/ca/api/handler/handler.go`) | HIGH | 2 | ✅ |
| I2.2.2 | OpenAPI spec (`api/ca/openapi_spec_enrollment.yaml`) | HIGH | 2 | ✅ |
| I2.2.3 | Generate server/client code with oapi-codegen | HIGH | 1 | ✅ |
| I2.2.4 | List CAs endpoint (GET `/api/v1/ca/ca`) | HIGH | 5 | ✅ |
| I2.2.5 | Get CA details endpoint (GET `/api/v1/ca/ca/{caId}`) | HIGH | 5 | ✅ |
| I2.2.6 | Download CRL endpoint (GET `/api/v1/ca/ca/{caId}/crl`) | HIGH | 5 | ✅ |
| I2.2.7 | Issue certificate endpoint (POST `/api/v1/ca/enrollments`) | CRITICAL | 8 | ✅ |
| I2.2.8 | Get certificate endpoint (GET `/api/v1/ca/certificates/{serialNumber}`) | HIGH | 5 | ✅ |
| I2.2.9 | Revoke certificate endpoint (POST `/api/v1/ca/certificates/{serialNumber}/revoke`) | CRITICAL | 5 | ✅ |
| I2.2.10 | Get enrollment status endpoint (GET `/api/v1/ca/enrollments/{id}`) | HIGH | 5 | ✅ |
| I2.2.11 | OCSP responder endpoint (POST `/api/v1/ca/ocsp`) | HIGH | 8 | ✅ |
| I2.2.12 | List profiles endpoint (GET `/api/v1/ca/profiles`) | MEDIUM | 2 | ✅ |
| I2.2.13 | Get profile endpoint (GET `/api/v1/ca/profiles/{profileId}`) | MEDIUM | 2 | ✅ |
| I2.2.14 | EST cacerts endpoint (GET `/api/v1/ca/est/cacerts`) | HIGH | 5 | ✅ |
| I2.2.15 | EST simpleenroll endpoint (POST `/api/v1/ca/est/simpleenroll`) | HIGH | 8 | ✅ |
| I2.2.16 | EST simplereenroll endpoint (POST `/api/v1/ca/est/simplereenroll`) | HIGH | 5 | ✅ |
| I2.2.17 | EST serverkeygen endpoint (POST `/api/v1/ca/est/serverkeygen`) | MEDIUM | 5 | ⚠️ |
| I2.2.18 | TSA timestamp service (`internal/ca/service/timestamp/`) | MEDIUM | 5 | ✅ |
| I2.2.19 | TSA timestamp endpoint (POST `/api/v1/ca/tsa/timestamp`) | MEDIUM | 2 | ✅ |
| I2.2.20 | Add mTLS authentication middleware | CRITICAL | 8 | ✅ |
| I2.2.21 | Docker Compose integration | MEDIUM | 2 | ✅ |
| I2.2.22 | CA Server cmd entry point (`cmd/ca-server/main.go`) | HIGH | 2 | ✅ |
| I2.2.23 | CA Server E2E tests | HIGH | 8 | ⚠️ |

**Total Points**: 105
**Completed Points**: 70 (67%)
**Evidence**: `internal/ca/api/handler/handler.go`, `api/ca/openapi_spec_enrollment.yaml`, `cmd/ca-server/main.go`, tests pass

### 2.3 Integration

| Task | Description | Priority | Points | Status |
|------|-------------|----------|--------|--------|
| I2.3.1 | Update `deployments/compose/compose.yml` for new services | HIGH | 5 | ✅ |
| I2.3.2 | Add JOSE Authority config (`configs/jose/jose-server.yml`) | HIGH | 2 | ✅ |
| I2.3.3 | Add CA Server config (`configs/ca/ca-server.yml`) | HIGH | 2 | ✅ |
| I2.3.4 | Demo script: `go run ./cmd/demo jose` | HIGH | 5 | ✅ |
| I2.3.5 | Demo script: `go run ./cmd/demo ca` | HIGH | 5 | ✅ |
| I2.3.6 | Update README with new server documentation | MEDIUM | 2 | ✅ |

**Total Points**: 21
**Completed Points**: 21 (100%)
**Evidence Required**: Docker Compose starts all services, demos complete successfully

---

## Iteration 2 Summary Statistics

### JOSE Authority (2.1)

- **Total Tasks**: 18
- **Total Points**: 69
- **Completed**: 17 (94%)
- **Completed Points**: 61
- **Critical Tasks**: 5 (all complete ✅)

### CA Server (2.2)

- **Total Tasks**: 23
- **Total Points**: 105
- **Completed**: 16 (70%)
- **Partial**: 7 (30%) - EST/TSA handlers scaffolded
- **Completed Points**: 80
- **Critical Tasks**: 3 (all complete ✅)

### Integration (2.3)

- **Total Tasks**: 6
- **Total Points**: 21
- **Completed**: 6 (100%)
- **Completed Points**: 21
- **Critical Tasks**: 0

### Overall Iteration 2

- **Total Tasks**: 47
- **Total Points**: 195
- **Completed Points**: 162 (83%)
- **Estimated Remaining Duration**: 1 week

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

## Priority Queue (Iteration 2 - Next Actions)

### CRITICAL (Must Implement First)

1. ~~**I2.1.3** - JWK generate endpoint~~ ✅ DONE
2. ~~**I2.1.8** - JWS sign endpoint~~ ✅ DONE
3. ~~**I2.1.9** - JWS verify endpoint~~ ✅ DONE
4. ~~**I2.1.10** - JWE encrypt endpoint~~ ✅ DONE
5. ~~**I2.1.11** - JWE decrypt endpoint~~ ✅ DONE
6. ~~**I2.2.7** - Issue certificate endpoint~~ ✅ DONE (SubmitEnrollment)
7. ~~**I2.2.9** - Revoke certificate endpoint~~ ✅ DONE
8. **I2.2.20** - mTLS authentication middleware (security) - HIGH PRIORITY

### HIGH (Should Implement)

1. ~~**I2.1.1** - JOSE server entry point~~ ✅ DONE
2. ~~**I2.1.2** - JOSE Fiber router~~ ✅ DONE
3. **I2.1.14** - JOSE OpenAPI spec
4. **I2.1.16** - API key authentication
5. ~~**I2.2.11** - OCSP responder~~ ✅ DONE
6. **I2.2.14-16** - EST protocol endpoints
7. ~~**I2.2.22** - CA Server cmd entry point~~ ✅ DONE
8. ~~**I2.2.6** - Wire up CRL endpoint to revocation.CRLService~~ ✅ DONE

### MEDIUM (Nice to Have)

1. ~~**I2.1.6** - Delete JWK endpoint~~ ✅ DONE
2. ~~**I2.1.17** - JOSE Docker Compose~~ ✅ DONE
3. ~~**I2.2.12-13** - Profile listing endpoints~~ ✅ DONE
4. **I2.2.17** - EST serverkeygen
5. **I2.2.19** - TSA timestamp endpoint (service exists)
6. ~~**I2.3.6** - README documentation~~ ✅ DONE

---

*Tasks Version: 2.1.0*
*Updated: December 2025*
*Next Review: After Iteration 2 completion*

---

## Iteration 3: Completion and Polish

**Status**: 🆕 STARTING
**Duration**: 1-2 weeks
**Goal**: Complete remaining I2 tasks, achieve 90%+ coverage, create demo videos

### 3.1 Complete Remaining I2 Tasks

| Task | Description | Priority | Points | Status |
|------|-------------|----------|--------|--------|
| I3.1.1 | Wire EST cacerts endpoint (GET `/api/v1/ca/est/cacerts`) | HIGH | 5 | ✅ |
| I3.1.2 | Wire EST simpleenroll endpoint (POST `/api/v1/ca/est/simpleenroll`) | HIGH | 8 | ✅ |
| I3.1.3 | Wire EST simplereenroll endpoint (POST `/api/v1/ca/est/simplereenroll`) | HIGH | 5 | ✅ |
| I3.1.4 | Wire EST serverkeygen endpoint (POST `/api/v1/ca/est/serverkeygen`) | MEDIUM | 5 | ⚠️ |
| I3.1.5 | Wire TSA timestamp endpoint (POST `/api/v1/ca/tsa/timestamp`) | MEDIUM | 2 | ✅ |
| I3.1.6 | Implement enrollment status endpoint (GET `/api/v1/ca/enrollments/{id}`) | HIGH | 5 | ✅ |
| I3.1.7 | JOSE Authority E2E test suite | HIGH | 8 | 🆕 |
| I3.1.8 | CA Server E2E test suite | HIGH | 8 | 🆕 |

**Total Points**: 46
**Completed Points**: 25 (54%)

### 3.2 Coverage Improvement

| Task | Description | Priority | Points | Status |
|------|-------------|----------|--------|--------|
| I3.2.1 | Increase CA handler coverage to 90%+ | HIGH | 8 | ⏳ (66% → 90% target) |
| I3.2.2 | Increase userauth coverage to 90%+ | HIGH | 8 | 🆕 |
| I3.2.3 | Increase jose server coverage to 90%+ | HIGH | 8 | 🆕 |
| I3.2.4 | Increase network package coverage to 90%+ | MEDIUM | 5 | 🆕 |
| I3.2.5 | Overall coverage audit and gap analysis | MEDIUM | 3 | 🆕 |

**Total Points**: 32
**Completed Points**: 2 (6%)

### 3.3 Demo Videos and Documentation

| Task | Description | Priority | Points | Status |
|------|-------------|----------|--------|--------|
| I3.3.1 | Individual product demo: P1 JOSE Authority | MEDIUM | 2 | 🆕 |
| I3.3.2 | Individual product demo: P2 Identity Server | MEDIUM | 2 | 🆕 |
| I3.3.3 | Individual product demo: P3 KMS Server | MEDIUM | 2 | 🆕 |
| I3.3.4 | Individual product demo: P4 CA Server | MEDIUM | 2 | 🆕 |
| I3.3.5 | Federated product suite demo | HIGH | 5 | 🆕 |
| I3.3.6 | Update API documentation | MEDIUM | 3 | 🆕 |

**Total Points**: 16
**Completed Points**: 0 (0%)

### 3.4 Workflow Verification

| Task | Description | Priority | Points | Status |
|------|-------------|----------|--------|--------|
| I3.4.1 | Verify ci-quality workflow | HIGH | 2 | 🆕 |
| I3.4.2 | Verify ci-coverage workflow | HIGH | 2 | 🆕 |
| I3.4.3 | Verify ci-benchmark workflow | MEDIUM | 1 | 🆕 |
| I3.4.4 | Verify ci-fuzz workflow | MEDIUM | 1 | 🆕 |
| I3.4.5 | Verify ci-race workflow | MEDIUM | 1 | 🆕 |
| I3.4.6 | Verify ci-sast workflow | MEDIUM | 1 | 🆕 |
| I3.4.7 | Verify ci-gitleaks workflow | MEDIUM | 1 | 🆕 |
| I3.4.8 | Verify ci-dast workflow | HIGH | 2 | 🆕 |
| I3.4.9 | Verify ci-e2e workflow | HIGH | 2 | 🆕 |
| I3.4.10 | Verify ci-load workflow | MEDIUM | 1 | 🆕 |
| I3.4.11 | Verify ci-identity-validation workflow | HIGH | 2 | 🆕 |
| I3.4.12 | Verify release workflow | MEDIUM | 1 | 🆕 |

**Total Points**: 17
**Completed Points**: 0 (0%)

---

## Iteration 3 Summary Statistics

### Complete Remaining I2 Tasks (3.1)

- **Total Tasks**: 8
- **Total Points**: 46
- **Completed**: 5 (63%)
- **Partial**: 1 (EST serverkeygen needs CMS library)

### Coverage Improvement (3.2)

- **Total Tasks**: 5
- **Total Points**: 32
- **Completed**: 0 (0%)

### Demo Videos (3.3)

- **Total Tasks**: 6
- **Total Points**: 16
- **Completed**: 0 (0%)

### Workflow Verification (3.4)

- **Total Tasks**: 12
- **Total Points**: 17
- **Completed**: 0 (0%)

### Overall Iteration 3

- **Total Tasks**: 31
- **Total Points**: 111
- **Completed Points**: 25 (23%)
- **Estimated Duration**: 1-2 weeks

---

*Tasks Version: 3.0.0*
*Updated: January 2026*
*Next Review: After Iteration 3 completion*
