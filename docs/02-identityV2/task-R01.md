# Task R01: Complete OAuth 2.1 Authorization Code Flow

**Priority**: 🔴 **CRITICAL** - Blocks all OAuth functionality
**Effort**: 16 hours (2 days)
**Owner**: OAuth/OIDC engineer
**Dependencies**: None
**Source**: GAP-ANALYSIS-DETAILED.md Priority 1 issues

---

## Problem Statement

The OAuth 2.1 authorization code flow is broken due to missing implementation in three critical areas:

1. **Authorization Request Persistence**: `/oauth2/v1/authorize` handler doesn't persist requests with PKCE challenge (lines 112-114 TODO)
2. **PKCE Verifier Validation**: `/oauth2/v1/token` handler doesn't validate code_verifier against stored code_challenge (lines 78-81 TODO)
3. **Consent Decision Storage**: `/oidc/v1/consent` handler doesn't store consent or generate authorization codes (lines 46-48 TODO)
4. **Placeholder User IDs**: Token endpoint uses randomly generated UUIDs instead of real user IDs from login/consent (lines 148-149 TODO)

**Impact**: No OAuth flows functional, blocks all authentication. Users cannot obtain access tokens.

---

## Acceptance Criteria

- [ ] Authorization requests persisted in database with PKCE challenge and 5-minute expiration
- [ ] Authorization endpoint redirects to IdP login with request_id parameter
- [ ] PKCE code_verifier validated against stored code_challenge using SHA-256
- [ ] Authorization codes enforced as single-use (marked as used after first redemption)
- [ ] Consent decisions stored with 30-day expiration
- [ ] Authorization codes generated after consent approval
- [ ] Tokens contain real user ID from authorization request (populated during consent)
- [ ] All repository methods have unit tests with table-driven test pattern
- [ ] Integration tests validate end-to-end authorization code flow
- [ ] Pre-commit hooks pass (golangci-lint, cspell, formatting)

---

## Implementation Steps

### Step 1: Create Repository Methods (4 hours)

**File**: `internal/identity/authz/repository/authorization_request_repository.go`

```go
type AuthorizationRequestRepository interface {
    CreateAuthorizationRequest(ctx context.Context, req *domain.AuthorizationRequest) error
    GetAuthorizationRequestByID(ctx context.Context, id googleUuid.UUID) (*domain.AuthorizationRequest, error)
    GetAuthorizationRequestByCode(ctx context.Context, code string) (*domain.AuthorizationRequest, error)
    UpdateAuthorizationRequest(ctx context.Context, req *domain.AuthorizationRequest) error
}
```

**Tests**: Create `authorization_request_repository_test.go` with table-driven tests for CRUD operations.

---

### Step 2: Implement Authorization Request Persistence (4 hours)

**File**: `internal/identity/authz/handlers_authorize.go`

**Replace lines 112-114**:
```go
// Store authorization request in database
authRequest := &domain.AuthorizationRequest{
    ID:                  googleUuid.Must(googleUuid.NewV7()),
    ClientID:            params.ClientID,
    RedirectURI:         params.RedirectURI,
    Scope:               params.Scope,
    State:               params.State,
    CodeChallenge:       params.CodeChallenge,
    CodeChallengeMethod: params.CodeChallengeMethod,
    CreatedAt:           time.Now(),
    ExpiresAt:           time.Now().Add(5 * time.Minute),
}

if err := s.authzRepo.CreateAuthorizationRequest(ctx, authRequest); err != nil {
    s.logger.Error("Failed to store authorization request", "error", err)
    return fiber.NewError(fiber.StatusInternalServerError, "Failed to store authorization request")
}

// Redirect to IdP login with request ID
loginURL := fmt.Sprintf("%s/oidc/v1/login?request_id=%s", s.cfg.IDP.BaseURL, authRequest.ID)
return c.Redirect(loginURL, fiber.StatusFound)
```

**Tests**: Verify persistence, expiration, redirect URL construction.

---

### Step 3: Implement PKCE Verifier Validation (6 hours)

**File**: `internal/identity/authz/handlers_token.go`

**Replace lines 78-81**:
```go
// Retrieve authorization request by code
authRequest, err := s.authzRepo.GetAuthorizationRequestByCode(ctx, authCode)
if err != nil || authRequest == nil {
    s.logger.Warn("Invalid authorization code", "code", authCode, "error", err)
    return fiber.NewError(fiber.StatusBadRequest, "Invalid authorization code")
}

// Validate code not expired
if time.Now().After(authRequest.ExpiresAt) {
    s.logger.Warn("Authorization code expired", "code", authCode, "expires_at", authRequest.ExpiresAt)
    return fiber.NewError(fiber.StatusBadRequest, "Authorization code expired")
}

// Validate code not already used (single-use enforcement)
if authRequest.Used {
    s.logger.Warn("Authorization code already used", "code", authCode, "used_at", authRequest.UsedAt)
    return fiber.NewError(fiber.StatusBadRequest, "Authorization code already used")
}

// Validate PKCE code_verifier
if !pkce.ValidateVerifier(codeVerifier, authRequest.CodeChallenge, authRequest.CodeChallengeMethod) {
    s.logger.Warn("Invalid PKCE code_verifier", "code", authCode, "client_id", authRequest.ClientID)
    return fiber.NewError(fiber.StatusBadRequest, "Invalid PKCE code_verifier")
}

// Mark authorization code as used
authRequest.Used = true
authRequest.UsedAt = time.Now()
if err := s.authzRepo.UpdateAuthorizationRequest(ctx, authRequest); err != nil {
    s.logger.Error("Failed to mark code as used", "code", authCode, "error", err)
    return fiber.NewError(fiber.StatusInternalServerError, "Failed to process authorization code")
}
```

**Tests**: Verify expiration enforcement, single-use validation, PKCE validation (SHA-256).

---

### Step 4: Implement Consent Decision Storage (5 hours)

**File**: `internal/identity/idp/handlers_consent.go`

**Replace lines 46-48**:
```go
// Store consent decision
consent := &domain.ConsentDecision{
    ID:        googleUuid.Must(googleUuid.NewV7()),
    UserID:    userID,
    ClientID:  authRequest.ClientID,
    Scope:     approvedScopes,
    GrantedAt: time.Now(),
    ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30-day consent
}

if err := s.consentRepo.CreateConsent(ctx, consent); err != nil {
    s.logger.Error("Failed to store consent", "user_id", userID, "client_id", authRequest.ClientID, "error", err)
    return fiber.NewError(fiber.StatusInternalServerError, "Failed to store consent decision")
}

// Generate authorization code (32 bytes crypto/rand)
authCode := generateSecureCode()
authRequest.Code = authCode
authRequest.UserID = userID
authRequest.ConsentID = consent.ID

if err := s.authzRepo.UpdateAuthorizationRequest(ctx, authRequest); err != nil {
    s.logger.Error("Failed to generate authorization code", "request_id", authRequest.ID, "error", err)
    return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate authorization code")
}

// Redirect to callback with code and state
callbackURL := fmt.Sprintf("%s?code=%s&state=%s", authRequest.RedirectURI, authCode, authRequest.State)
return c.Redirect(callbackURL, fiber.StatusFound)
```

**Helper Function**:
```go
func generateSecureCode() string {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        panic(fmt.Sprintf("Failed to generate secure code: %v", err))
    }
    return base64.RawURLEncoding.EncodeToString(b)
}
```

**Tests**: Verify consent storage, code generation uniqueness, redirect URL construction.

---

### Step 5: Replace Placeholder User IDs (1 hour)

**File**: `internal/identity/authz/handlers_token.go`

**Replace lines 148-149**:
```go
// Use real user ID from authorization request (populated during consent)
userID := authRequest.UserID
if userID == googleUuid.Nil {
    s.logger.Error("Authorization request missing user ID", "request_id", authRequest.ID)
    return fiber.NewError(fiber.StatusInternalServerError, "Authorization request missing user ID")
}
```

**Tests**: Verify tokens contain correct user ID, validate error handling for missing user ID.

---

## Testing Requirements

### Unit Tests

**File**: `internal/identity/authz/handlers_authorize_test.go`
- Authorization request persistence (table-driven test with valid/invalid clients, PKCE methods)
- Redirect URL construction validation
- Error handling (database failures, invalid parameters)

**File**: `internal/identity/authz/handlers_token_test.go`
- PKCE validation (SHA-256 challenge/verifier pairs)
- Authorization code expiration (5 minutes)
- Single-use enforcement (attempt code reuse)
- Placeholder user ID replacement

**File**: `internal/identity/idp/handlers_consent_test.go`
- Consent decision storage (30-day expiration)
- Authorization code generation (uniqueness, format)
- Callback redirect URL construction

### Integration Tests

**File**: `internal/identity/test/e2e/oauth_authorization_code_flow_test.go`
- End-to-end authorization code flow:
  1. Client initiates `/oauth2/v1/authorize` with PKCE challenge
  2. User redirected to `/oidc/v1/login`
  3. User authenticates and grants consent
  4. Client receives authorization code
  5. Client exchanges code for tokens with code_verifier
  6. Verify tokens contain real user ID
  7. Attempt code reuse fails (single-use validation)

---

## Pre-commit Enforcement

- Run `golangci-lint run --fix` on all modified files
- Fix all linting errors (wsl, godot, mnd, errcheck)
- Ensure test coverage ≥85% for infrastructure code (repository methods)
- Ensure test coverage ≥80% for application code (handlers)
- Run `go test ./... -cover` before committing
- Commit with conventional commit message: `feat(authz): complete OAuth 2.1 authorization code flow with PKCE`

---

## Validation

**Success Criteria**:
- [ ] Authorization request persisted with PKCE challenge
- [ ] PKCE code_verifier validated correctly
- [ ] Authorization codes single-use enforced
- [ ] Consent decisions stored
- [ ] Tokens contain real user IDs
- [ ] All unit tests pass
- [ ] Integration test completes full authorization code flow
- [ ] Pre-commit hooks pass
- [ ] Code coverage ≥85% (infrastructure) and ≥80% (application)

---

## References

- GAP-ANALYSIS-DETAILED.md: Priority 1 issues (lines 39-176)
- REMEDIATION-MASTER-PLAN-2025.md: R01 section (lines 36-220)
- OAuth 2.1 Draft: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1-11
- PKCE RFC 7636: https://datatracker.ietf.org/doc/html/rfc7636
