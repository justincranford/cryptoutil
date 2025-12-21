# Quality Improvement TODOs

## Overview

This document tracks quality improvement tasks discovered through code analysis, coverage reports, and workflow monitoring. Organized by priority and phase alignment.

**Last Updated**: 2025-12-21

---

## Phase 4: Quality Gates (Current Phase)

### Priority 1: Critical Blockers

#### Identity Services - Missing Public Servers

**Package**: `internal/identity/rs/server/`

**Issue**: Resource Server (RS) missing public HTTP server implementation

**Evidence**:

- File verification (2025-12-21): Only `admin.go` and `application.go` exist
- `application.go` only creates `adminServer`, missing `publicServer` initialization
- Authz and IdP public servers implemented (165 lines each), RS incomplete

**Impact**:

- E2E workflows expected to fail for RS-dependent flows
- Load/DAST workflows cannot test RS protected resource access
- Token validation E2E tests blocked

**Tasks**:

- [ ] Create `internal/identity/rs/server/public_server.go` (copy pattern from authz/idp)
- [ ] Update `application.go` to create both `publicServer` and `adminServer`
- [ ] Implement RS public endpoints (protected resource API, token validation)
- [ ] Add unit tests for RS public server (target 95%+ coverage)
- [ ] Add integration tests for RS public endpoints
- [ ] Verify E2E workflows pass after RS implementation

**Estimated Effort**: 1-2 days

**Related**:

- Constitution.md service status table (RS marked INCOMPLETE)
- DETAILED.md 2025-12-21 service status verification
- WORKFLOW-FIXES-CONSOLIDATED.md Round 7 investigation

---

### Priority 2: Test Coverage Gaps

#### E2E Tests - Skipped Workflows

**Packages**: `internal/test/e2e/`, `internal/identity/test/e2e/`

**Issue**: Multiple E2E tests skipped with TODO comments for Phase 4 implementation

**Skipped Tests**:

1. **OAuth 2.1 Workflows** (`oauth_workflow_test.go`):
   - `TestOAuthE2ESuite/TestAuthorizationCodeFlow` - Skip: "TODO P4.1: Full OAuth 2.1 E2E implementation - requires client registration API"
   - `TestOAuthE2ESuite/TestTokenRefreshFlow` - Skip: "TODO P4.1: Full OAuth 2.1 E2E implementation - requires client registration API"

2. **JOSE Workflows** (`jose_workflow_test.go`):
   - `TestJOSEE2ESuite/TestJWTSignAndVerify` - Skip: "TODO P4.4: Implement full JOSE JWT sign/verify workflow"
   - `TestJOSEE2ESuite/TestJWKSEndpoint` - Skip: "TODO P4.4: Implement JOSE JWKS endpoint E2E test"
   - `TestJOSEE2ESuite/TestJWKRotation` - Skip: "TODO P4.4: Implement JOSE JWK rotation E2E test"
   - `TestJOSEE2ESuite/TestJWEEncryptionDecryption` - Skip: "TODO P4.4: Implement JOSE JWE encryption E2E test"

3. **CA Workflows** (`ca_workflow_test.go`):
   - `TestCAE2ESuite/TestCertificateEnrollmentWorkflow` - Skip: "TODO P4.3: Implement full CA enrollment workflow"
   - `TestCAE2ESuite/TestOCSPWorkflow` - Skip: "TODO P4.3: Implement CA OCSP workflow E2E test"
   - `TestCAE2ESuite/TestCRLDistribution` - Skip: "TODO P4.3: Implement CA CRL distribution E2E test"
   - `TestCAE2ESuite/TestCertificateProfiles` - Skip: "TODO P4.3: Implement CA certificate profiles E2E test"

4. **MFA Flows** (`internal/identity/test/e2e/mfa_flows_test.go`):
   - MFA chain testing (line 64)
   - Step-up authentication testing (line 108)
   - Risk-based authentication testing (line 163)
   - Client MFA chain testing (line 192)

5. **Observability** (`internal/identity/test/e2e/observability_test.go`):
   - Grafana Tempo API trace queries (line 240)
   - Grafana Loki API log queries (line 252)

**Tasks**:

- [ ] Implement OAuth 2.1 client registration API (authz OpenAPI spec)
- [ ] Implement E2E test: OAuth authorization code flow
- [ ] Implement E2E test: OAuth token refresh flow
- [ ] Implement E2E test: JOSE JWT sign/verify workflow
- [ ] Implement E2E test: JOSE JWKS endpoint
- [ ] Implement E2E test: JOSE JWK rotation
- [ ] Implement E2E test: JOSE JWE encryption/decryption
- [ ] Implement E2E test: CA certificate enrollment (CSR generation, cert retrieval)
- [ ] Implement E2E test: CA OCSP workflow
- [ ] Implement E2E test: CA CRL distribution
- [ ] Implement E2E test: CA certificate profiles
- [ ] Implement E2E test: MFA chain testing
- [ ] Implement E2E test: Step-up authentication
- [ ] Implement E2E test: Risk-based authentication
- [ ] Implement E2E test: Client MFA chain
- [ ] Implement E2E test: Grafana Tempo trace queries
- [ ] Implement E2E test: Grafana Loki log queries

**Estimated Effort**: 5-10 days (17 E2E test implementations)

**Related**:

- Plan.md Phase 4 E2E implementation tasks
- Spec.md E2E testing requirements
- DETAILED.md Phase 4 progress tracking

---

#### MFA Implementation Gaps

**Packages**: `internal/identity/idp/auth/`

**Issue**: MFA authenticators have stub implementations with TODO comments

**Affected Files**:

1. **TOTP/HOTP** (`totp.go`):
   - Line 45: "TODO: Fetch MFA factors for user"
   - Line 46: "TODO: Validate TOTP/HOTP code using library (e.g., pquerna/otp)"
   - Line 47: "TODO: Return user object if validation succeeds"

2. **Passkey/WebAuthn** (`passkey.go`):
   - Line 47: "TODO: Validate WebAuthn assertion using library (e.g., go-webauthn/webauthn)"
   - Line 48: "TODO: Fetch user by credential ID"
   - Line 49: "TODO: Verify signature and challenge"
   - Line 50: "TODO: Return user object if validation succeeds"

3. **OTP via Email/SMS** (`otp.go`):
   - Line 31: "TODO: Add dependencies for email/SMS delivery"
   - Line 41: "TODO: Generate OTP code (6-digit numeric)"
   - Line 42: "TODO: Store OTP with expiration (5 minutes)"
   - Line 43: "TODO: Send OTP via email/SMS based on method"
   - Line 52: "TODO: Fetch stored OTP for user"
   - Line 53: "TODO: Validate OTP code matches"
   - Line 54: "TODO: Check OTP not expired"
   - Line 55: "TODO: Invalidate OTP after successful validation"

4. **MFA OTP** (`mfa_otp.go`):
   - Line 139: "TODO: Retrieve user ID from authentication context"

**Tasks**:

- [ ] Integrate TOTP/HOTP library (pquerna/otp or similar)
- [ ] Implement MFA factor storage and retrieval
- [ ] Implement TOTP/HOTP code validation
- [ ] Integrate WebAuthn library (go-webauthn/webauthn or similar)
- [ ] Implement WebAuthn credential storage
- [ ] Implement WebAuthn assertion validation
- [ ] Implement OTP code generation (6-digit numeric)
- [ ] Implement OTP storage with expiration (Redis or database)
- [ ] Integrate email/SMS delivery service
- [ ] Implement OTP delivery via email
- [ ] Implement OTP delivery via SMS
- [ ] Implement OTP validation and expiration checking
- [ ] Implement OTP invalidation after use
- [ ] Add unit tests for all MFA authenticators (95%+ coverage)
- [ ] Add integration tests for MFA workflows

**Estimated Effort**: 3-5 days

**Related**:

- Spec.md MFA requirements
- Plan.md Phase 3 MFA implementation tasks
- Constitution.md FIPS 140-3 compliance (password hashing, key storage)

---

#### Notification System Gaps

**Package**: `internal/identity/notifications/`

**Issue**: Notification delivery methods have stub implementations

**Affected Files**:

1. **Webhook Notifier** (`notifiers.go` line 46):
   - "TODO: Implement HTTP POST to webhook URL with notification payload"

2. **Email Notifier** (`notifiers.go` line 72):
   - "TODO: Implement SMTP email sending with notification details"

**Tasks**:

- [ ] Implement HTTP webhook POST delivery
- [ ] Add webhook retry logic (exponential backoff)
- [ ] Implement SMTP email sending
- [ ] Add email template system
- [ ] Add configuration for webhook URLs
- [ ] Add configuration for SMTP settings (host, port, credentials)
- [ ] Add unit tests for webhook delivery (95%+ coverage)
- [ ] Add unit tests for email delivery (95%+ coverage)
- [ ] Add integration tests with mock SMTP server

**Estimated Effort**: 2-3 days

**Related**:

- Spec.md notification requirements
- Plan.md Phase 3 notification implementation

---

### Priority 3: Test Infrastructure Improvements

#### Rate Limiting Implementation

**Package**: `internal/identity/idp/`

**Issue**: Rate limiting deferred (MEDIUM priority TODO)

**Evidence**:

- `handlers_security_validation_test.go` line 481: "Note: Rate limiting implementation is deferred (MEDIUM priority TODO)"
- `handlers_security_validation_test.go` line 522: "TODO: When rate limiting is implemented, update this test to expect..."

**Tasks**:

- [ ] Design rate limiting strategy (token bucket, sliding window, fixed window)
- [ ] Implement rate limiter middleware
- [ ] Add rate limiter configuration (requests per second, burst size)
- [ ] Add rate limiter storage (in-memory, Redis)
- [ ] Update security validation tests to verify rate limiting
- [ ] Add unit tests for rate limiter (95%+ coverage)
- [ ] Add integration tests for rate limiting behavior

**Estimated Effort**: 2-3 days

**Related**:

- Security instructions (01-07.security.instructions.md)
- Architecture instructions (per-IP rate limiting)

---

#### Key Rotation Testing Infrastructure

**Package**: `internal/identity/integration/`

**Issue**: Key rotation testing infrastructure not implemented

**Evidence**:

- `integration_test.go` line 195: "TODO: Implement proper key rotation testing infrastructure"

**Tasks**:

- [ ] Design key rotation test infrastructure
- [ ] Implement key version tracking in tests
- [ ] Implement rotation trigger mechanisms (time-based, manual)
- [ ] Add verification that old keys still decrypt data
- [ ] Add verification that new keys encrypt new data
- [ ] Add integration tests for key rotation workflows
- [ ] Document key rotation testing patterns

**Estimated Effort**: 2-3 days

**Related**:

- KMS hierarchical key security (constitution.md Section III)
- Cryptography instructions (key rotation pattern)

---

#### Authentication Strength Enum

**Package**: `internal/identity/domain/`

**Issue**: AuthenticationStrength enum not defined

**Evidence**:

- `internal/identity/test/e2e/client_mfa_test.go` line 252: "TODO: Define AuthenticationStrength enum in domain package"
- `internal/identity/test/e2e/client_mfa_test.go` line 286: "Compare string representations (TODO: use enum when defined)"

**Tasks**:

- [ ] Define AuthenticationStrength enum in domain package
- [ ] Values: NONE, BASIC, MFA, STRONG_MFA (or similar)
- [ ] Add JSON marshaling/unmarshaling
- [ ] Add database column type support
- [ ] Update client MFA tests to use enum instead of string comparison
- [ ] Add unit tests for enum (100% coverage)

**Estimated Effort**: 1 day

**Related**:

- Domain model patterns (Go instructions)
- MFA requirements (spec.md)

---

#### Key Repository Query Logic

**Package**: `internal/identity/repository/orm/`

**Issue**: FindByUsage query logic needs fixing

**Evidence**:

- `key_repository_test.go` line 124: "TODO: Add TestKeyRepository_FindByUsage tests - requires fixing active field query logic"

**Tasks**:

- [ ] Fix active field query logic in KeyRepository.FindByUsage
- [ ] Add unit tests for FindByUsage (95%+ coverage)
- [ ] Test query with active=true, active=false, no filter
- [ ] Verify GORM query translation to SQL
- [ ] Add integration tests with real database

**Estimated Effort**: 1 day

---

### Priority 4: Code Quality Improvements

#### Context.TODO() Usage

**Issue**: Multiple uses of `context.TODO()` instead of proper context propagation

**Occurrences**:

1. `internal/identity/test/contract/delivery_service_test.go` (lines 136, 141)
2. `internal/identity/repository/migrations.go` (line 46)
3. `internal/identity/rs/server/admin_test.go` (line 132)

**Tasks**:

- [ ] Review all `context.TODO()` usages
- [ ] Replace with proper context propagation where applicable
- [ ] Document cases where `context.TODO()` is acceptable (migrations, cleanup)
- [ ] Add linting rule to detect new `context.TODO()` usage

**Estimated Effort**: 1 day

---

## Summary

**Total Tasks**: 70+ discrete tasks across 8 categories

**Estimated Total Effort**: 15-30 days

**Priority Breakdown**:

- **P1 (Critical Blockers)**: 6 tasks, 1-2 days (RS public server implementation)
- **P2 (Test Coverage Gaps)**: 45+ tasks, 10-18 days (E2E tests, MFA, notifications)
- **P3 (Infrastructure)**: 15+ tasks, 6-9 days (rate limiting, key rotation, enums)
- **P4 (Code Quality)**: 4+ tasks, 1 day (context.TODO() cleanup)

**Recommended Approach**:

1. **Immediate**: Complete P1 (RS public server) to unblock E2E workflows
2. **Short Term**: Focus on P2 (E2E tests, MFA) for Phase 4 quality gates
3. **Medium Term**: P3 (infrastructure improvements) for robustness
4. **Ongoing**: P4 (code quality) as refactoring opportunities arise

**Success Criteria**:

- All E2E tests passing (no skips)
- Coverage ≥95% for all packages
- Mutation score ≥98% (Phase 5 target)
- All TODOs resolved or tracked in tasks.md
- All workflows passing (11/11 green)
