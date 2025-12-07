# Phase 3: Achieve Coverage Targets Implementation Guide

**Duration**: Days 7-9 (2-3 hours)  
**Prerequisites**: Phase 2 complete (all deferred features implemented)  
**Status**: ❌ Not Started

## Overview

Phase 3 closes coverage gaps in 5 packages to achieve the mandatory 95%+ coverage target for production code. Per 01-02.testing.instructions.md:

- **Target coverage**: 95%+ production, 100%+ infrastructure (cicd), 100% utility code
- **ALWAYS use table-driven tests** with `t.Parallel()`
- **NEVER hardcode test values** - use magic package constants OR runtime-generated UUIDv7

**Task Breakdown**:

- P3.1: ca/handler Coverage (47.2% → 95%) - 1h
- P3.2: auth/userauth Coverage (42.6% → 95%) - 1h
- P3.3: unsealkeysservice Coverage (78.2% → 95%) - 30min
- P3.4: network Coverage (88.7% → 95%) - 30min
- P3.5: apperr Coverage (96.6%) - ✅ Already complete

**Note**: P3.5 already meets target, leaving only 4 tasks requiring implementation.

## Task Details

---

### P3.1: ca/handler Coverage (47.2% → 95%) ⭐ CRITICAL

**Priority**: CRITICAL  
**Effort**: 1 hour  
**Status**: ❌ Not Started

**Objective**: Increase CA handler test coverage from 47.2% to ≥95% by adding comprehensive tests for all endpoints.

**Current Coverage Analysis**:

```bash
# Check current coverage
go test -coverprofile=test-output/coverage_ca_handler.out ./internal/ca/handler
go tool cover -func=test-output/coverage_ca_handler.out | grep total
```

**Current State**:

- Basic happy path tests exist
- Missing error path tests
- Incomplete coverage of EST endpoints
- No comprehensive validation tests

**Implementation Strategy**:

```bash
# Step 1: Identify untested code paths
go test -coverprofile=test-output/coverage_ca_handler.out ./internal/ca/handler
go tool cover -html=test-output/coverage_ca_handler.out

# Step 2: Create/enhance handler tests
# Files to modify/create:
# - internal/ca/handler/est_cacerts_test.go
# - internal/ca/handler/est_simpleenroll_test.go
# - internal/ca/handler/est_simplereenroll_test.go
# - internal/ca/handler/tsa_timestamp_test.go
# - internal/ca/handler/ocsp_test.go (if P2.2 implemented)
# - internal/ca/handler/crl_test.go
```

**Test Pattern for CA Handlers**:

```go
// File: internal/ca/handler/est_simpleenroll_test.go
package handler_test

import (
    "testing"
    "github.com/stretchr/testify/require"
    googleUuid "github.com/google/uuid"
)

func TestESTSimpleEnroll(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name      string
        csrPEM    string
        wantErr   bool
        errStatus int
    }{
        {
            name:    "valid RSA CSR",
            csrPEM:  generateValidRSACSR(t),
            wantErr: false,
        },
        {
            name:    "valid EC CSR",
            csrPEM:  generateValidECCSR(t),
            wantErr: false,
        },
        {
            name:      "invalid CSR format",
            csrPEM:    "invalid-pem",
            wantErr:   true,
            errStatus: http.StatusBadRequest,
        },
        {
            name:      "empty CSR",
            csrPEM:    "",
            wantErr:   true,
            errStatus: http.StatusBadRequest,
        },
        {
            name:      "CSR with weak key",
            csrPEM:    generateWeakKeyCSR(t),
            wantErr:   true,
            errStatus: http.StatusBadRequest,
        },
    }

    for _, tc := range tests {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            // Setup test CA handler
            handler := setupTestCAHandler(t)

            // Create HTTP request
            req := httptest.NewRequest(http.MethodPost, "/est/v1/simpleenroll", strings.NewReader(tc.csrPEM))
            req.Header.Set("Content-Type", "application/pkcs10")

            // Execute request
            rr := httptest.NewRecorder()
            handler.HandleSimpleEnroll(rr, req)

            // Validate response
            if tc.wantErr {
                require.Equal(t, tc.errStatus, rr.Code)
            } else {
                require.Equal(t, http.StatusOK, rr.Code)
                require.Equal(t, "application/pkcs7-mime", rr.Header().Get("Content-Type"))

                // Verify returned certificate
                certPEM := rr.Body.Bytes()
                require.NotEmpty(t, certPEM)

                // Parse and validate certificate
                cert, err := x509.ParseCertificate(certPEM)
                require.NoError(t, err)
                require.NotNil(t, cert)
            }
        })
    }
}

func setupTestCAHandler(t *testing.T) *handler.CAHandler {
    // Create test CA with in-memory database
    db := setupTestDB(t)
    ca := caservice.New(db)
    return handler.New(ca)
}

func generateValidRSACSR(t *testing.T) string {
    // Generate RSA private key
    key, err := rsa.GenerateKey(rand.Reader, 2048)
    require.NoError(t, err)

    // Create CSR
    template := &x509.CertificateRequest{
        Subject: pkix.Name{
            CommonName: "test-" + googleUuid.NewV7().String(),
        },
    }

    csrDER, err := x509.CreateCertificateRequest(rand.Reader, template, key)
    require.NoError(t, err)

    csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrDER})
    return string(csrPEM)
}
```

**Coverage Targets by File**:

| File | Current | Target | Focus Areas |
|------|---------|--------|-------------|
| est_cacerts.go | ~60% | 95% | Error paths, content negotiation |
| est_simpleenroll.go | ~40% | 95% | CSR validation, cert issuance errors |
| est_simplereenroll.go | ~40% | 95% | Renewal validation, expiration checks |
| tsa_timestamp.go | ~50% | 95% | TSA request parsing, signing |
| ocsp.go | 0% | 95% | All paths (new file from P2.2) |
| crl.go | ~70% | 95% | CRL generation edge cases |

**Acceptance Criteria**:

- ✅ All CA handler endpoints have table-driven tests
- ✅ Happy path and error path tests for each endpoint
- ✅ All tests use `t.Parallel()`
- ✅ Test data uses UUIDv7 for uniqueness
- ✅ Coverage ≥95% for internal/ca/handler package
- ✅ No hardcoded test values

**Validation Commands**:

```bash
# Run handler tests with coverage
go test -coverprofile=test-output/coverage_ca_handler.out ./internal/ca/handler -v

# View coverage report
go tool cover -func=test-output/coverage_ca_handler.out | grep total

# View HTML coverage report (identify missing lines)
go tool cover -html=test-output/coverage_ca_handler.out
```

---

### P3.2: auth/userauth Coverage (42.6% → 95%) ⭐ CRITICAL

**Priority**: CRITICAL  
**Effort**: 1 hour  
**Status**: ❌ Not Started

**Objective**: Increase user authentication test coverage from 42.6% to ≥95% by adding comprehensive authentication flow tests.

**Current Coverage Analysis**:

```bash
# Check current coverage
go test -coverprofile=test-output/coverage_userauth.out ./internal/identity/auth/userauth
go tool cover -func=test-output/coverage_userauth.out | grep total
```

**Current State**:

- Basic authentication tests exist
- Missing MFA flow tests
- Incomplete password validation tests
- No session management tests

**Implementation Strategy**:

```bash
# Step 1: Identify untested authentication flows
go test -coverprofile=test-output/coverage_userauth.out ./internal/identity/auth/userauth
go tool cover -html=test-output/coverage_userauth.out

# Step 2: Create comprehensive auth tests
# Files to modify/create:
# - internal/identity/auth/userauth/authenticate_test.go
# - internal/identity/auth/userauth/password_test.go
# - internal/identity/auth/userauth/session_test.go
# - internal/identity/auth/userauth/mfa_test.go
```

**Test Pattern for User Authentication**:

```go
// File: internal/identity/auth/userauth/authenticate_test.go
package userauth_test

import (
    "testing"
    "github.com/stretchr/testify/require"
    googleUuid "github.com/google/uuid"
)

func TestAuthenticate(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name         string
        username     string
        password     string
        mfaEnabled   bool
        mfaCode      string
        wantErr      bool
        errCode      string
    }{
        {
            name:       "valid credentials no MFA",
            username:   "user-" + googleUuid.NewV7().String(),
            password:   "ValidPass123!",
            mfaEnabled: false,
            wantErr:    false,
        },
        {
            name:       "valid credentials with MFA",
            username:   "user-" + googleUuid.NewV7().String(),
            password:   "ValidPass123!",
            mfaEnabled: true,
            mfaCode:    "123456",
            wantErr:    false,
        },
        {
            name:     "invalid password",
            username: "user-" + googleUuid.NewV7().String(),
            password: "wrong",
            wantErr:  true,
            errCode:  "invalid_credentials",
        },
        {
            name:       "missing MFA code",
            username:   "user-" + googleUuid.NewV7().String(),
            password:   "ValidPass123!",
            mfaEnabled: true,
            mfaCode:    "",
            wantErr:    true,
            errCode:    "mfa_required",
        },
        {
            name:       "invalid MFA code",
            username:   "user-" + googleUuid.NewV7().String(),
            password:   "ValidPass123!",
            mfaEnabled: true,
            mfaCode:    "000000",
            wantErr:    true,
            errCode:    "invalid_mfa",
        },
        {
            name:     "account locked",
            username: "locked-" + googleUuid.NewV7().String(),
            password: "ValidPass123!",
            wantErr:  true,
            errCode:  "account_locked",
        },
        {
            name:     "account disabled",
            username: "disabled-" + googleUuid.NewV7().String(),
            password: "ValidPass123!",
            wantErr:  true,
            errCode:  "account_disabled",
        },
    }

    for _, tc := range tests {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            // Setup test user with unique data
            ctx := context.Background()
            db := setupTestDB(t)
            authService := userauth.New(db)

            // Create test user
            user := createTestUser(t, db, tc.username, tc.password, tc.mfaEnabled)

            // Attempt authentication
            session, err := authService.Authenticate(ctx, tc.username, tc.password, tc.mfaCode)

            // Validate result
            if tc.wantErr {
                require.Error(t, err)
                require.Contains(t, err.Error(), tc.errCode)
                require.Nil(t, session)
            } else {
                require.NoError(t, err)
                require.NotNil(t, session)
                require.Equal(t, user.ID, session.UserID)
            }
        })
    }
}

func TestPasswordValidation(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        password string
        wantErr  bool
        errMsg   string
    }{
        {
            name:     "valid strong password",
            password: "ValidPass123!",
            wantErr:  false,
        },
        {
            name:     "too short",
            password: "short",
            wantErr:  true,
            errMsg:   "minimum 8 characters",
        },
        {
            name:     "no uppercase",
            password: "nouppercase123!",
            wantErr:  true,
            errMsg:   "uppercase letter",
        },
        {
            name:     "no lowercase",
            password: "NOLOWERCASE123!",
            wantErr:  true,
            errMsg:   "lowercase letter",
        },
        {
            name:     "no digit",
            password: "NoDigits!",
            wantErr:  true,
            errMsg:   "digit",
        },
        {
            name:     "no special char",
            password: "NoSpecial123",
            wantErr:  true,
            errMsg:   "special character",
        },
    }

    for _, tc := range tests {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            err := userauth.ValidatePassword(tc.password)

            if tc.wantErr {
                require.Error(t, err)
                require.Contains(t, err.Error(), tc.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

**Coverage Targets by File**:

| File | Current | Target | Focus Areas |
|------|---------|--------|-------------|
| authenticate.go | ~40% | 95% | All auth flows, error paths |
| password.go | ~30% | 95% | Validation rules, hashing |
| session.go | ~50% | 95% | Session creation, expiration |
| mfa.go | ~20% | 95% | TOTP, backup codes, enrollment |

**Acceptance Criteria**:

- ✅ All authentication flows tested (password, MFA, session)
- ✅ Password validation tests cover all requirements
- ✅ MFA tests cover TOTP and backup codes
- ✅ Session management tests cover creation and expiration
- ✅ Error paths tested (invalid credentials, locked accounts, etc.)
- ✅ Coverage ≥95% for internal/identity/auth/userauth package
- ✅ All tests use `t.Parallel()` and UUIDv7

**Validation Commands**:

```bash
# Run auth tests with coverage
go test -coverprofile=test-output/coverage_userauth.out ./internal/identity/auth/userauth -v

# View coverage report
go tool cover -func=test-output/coverage_userauth.out | grep total

# View HTML coverage report
go tool cover -html=test-output/coverage_userauth.out
```

---

### P3.3: unsealkeysservice Coverage (78.2% → 95%)

**Priority**: MEDIUM  
**Effort**: 30 minutes  
**Status**: ❌ Not Started

**Objective**: Increase unseal key service coverage from 78.2% to ≥95% by adding edge case and error handling tests.

**Current Coverage Analysis**:

```bash
# Check current coverage
go test -coverprofile=test-output/coverage_unseal.out ./internal/kms/server/unsealkeysservice
go tool cover -func=test-output/coverage_unseal.out | grep total
```

**Implementation Strategy**:

```bash
# Step 1: Identify untested edge cases
go tool cover -html=test-output/coverage_unseal.out

# Step 2: Add edge case tests
# File to modify: internal/kms/server/unsealkeysservice/unseal_keys_service_test.go
```

**Edge Cases to Test**:

- Invalid unseal secret format
- Mismatched unseal secret count
- Concurrent unseal operations
- Unseal with corrupted keys
- Reseal operations
- Key rotation scenarios

**Acceptance Criteria**:

- ✅ Edge case tests added
- ✅ Error handling tests comprehensive
- ✅ Coverage ≥95%
- ✅ All tests use `t.Parallel()`

**Validation Commands**:

```bash
# Run tests with coverage
go test -coverprofile=test-output/coverage_unseal.out ./internal/kms/server/unsealkeysservice -v

# View coverage
go tool cover -func=test-output/coverage_unseal.out | grep total
```

---

### P3.4: network Coverage (88.7% → 95%)

**Priority**: MEDIUM  
**Effort**: 30 minutes  
**Status**: ❌ Not Started

**Objective**: Increase network package coverage from 88.7% to ≥95% by adding error path and failure scenario tests.

**Current Coverage Analysis**:

```bash
# Check current coverage
go test -coverprofile=test-output/coverage_network.out ./internal/common/network
go tool cover -func=test-output/coverage_network.out | grep total
```

**Implementation Strategy**:

```bash
# Step 1: Identify untested error paths
go tool cover -html=test-output/coverage_network.out

# Step 2: Add network failure tests
# File to modify: internal/common/network/*_test.go
```

**Error Scenarios to Test**:

- Connection timeout
- Network unreachable
- TLS handshake failure
- Certificate validation errors
- DNS resolution failures
- Port binding conflicts

**Acceptance Criteria**:

- ✅ Error path tests added
- ✅ Network failure scenarios tested
- ✅ Coverage ≥95%
- ✅ All tests use `t.Parallel()`

**Validation Commands**:

```bash
# Run tests with coverage
go test -coverprofile=test-output/coverage_network.out ./internal/common/network -v

# View coverage
go tool cover -func=test-output/coverage_network.out | grep total
```

---

### P3.5: apperr Coverage (96.6%) ✅

**Priority**: LOW  
**Effort**: 5 minutes  
**Status**: ✅ Already Complete

**Current State**: Coverage already at 96.6%, exceeding 95% target.

**Validation**:

```bash
# Verify current coverage
go test -coverprofile=test-output/coverage_apperr.out ./internal/common/apperr
go tool cover -func=test-output/coverage_apperr.out | grep total
```

**Acceptance Criteria**:

- ✅ Verify coverage ≥95% (already met)
- ✅ No action required

---

## Common Testing Patterns

### Table-Driven Tests (MANDATORY)

Per 01-02.testing.instructions.md:

```go
func TestMyFunction(t *testing.T) {
    t.Parallel()

    tests := []struct{
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "good", false},
        {"invalid", "", true},
    }

    for _, tc := range tests {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            // test using tc fields
        })
    }
}
```

### Test Data Isolation

Use UUIDv7 for unique test data:

```go
// CORRECT - generate once, reuse
userID := googleUuid.NewV7()
user := &User{ID: userID, Name: "test-" + userID.String()}

// WRONG - hardcoded values
user := &User{ID: "12345", Name: "test-user"}
```

### TestMain Pattern for Shared Dependencies

```go
var testDB *sql.DB

func TestMain(m *testing.M) {
    // Start PostgreSQL container ONCE per package
    testDB = startPostgreSQLContainer()
    exitCode := m.Run()
    testDB.Close()
    os.Exit(exitCode)
}
```

## Progress Tracking

After completing each task, update `PROGRESS.md`:

```bash
# Edit PROGRESS.md to mark task complete
# Update executive summary percentages
# Commit and push
git add specs/001-cryptoutil/PROGRESS.md
git commit -m "docs(speckit): mark P3.X complete"
git push
```

## Validation Checklist

Before marking Phase 3 complete, verify:

- [ ] P3.1: ca/handler coverage ≥95%
- [ ] P3.2: auth/userauth coverage ≥95%
- [ ] P3.3: unsealkeysservice coverage ≥95%
- [ ] P3.4: network coverage ≥95%
- [ ] P3.5: apperr coverage verified ≥95%
- [ ] PROGRESS.md updated with all P3.1-P3.5 marked complete
- [ ] All tests passing with `go test ./...`
- [ ] No coverage regressions in other packages

## Next Phase

After Phase 3 complete:

- Proceed to Phase 4: Advanced Testing
- Use PHASE4-IMPLEMENTATION.md guide
- Update PROGRESS.md executive summary
