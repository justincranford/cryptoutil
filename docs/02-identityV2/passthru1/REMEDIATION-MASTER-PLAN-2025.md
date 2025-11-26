# Identity V2 Remediation Master Plan - 2025

**Plan Date**: 2025-01-XX
**Scope**: Complete incomplete OAuth 2.1 flows, harden security, validate production readiness
**Baseline**: Post-Task 20 state (Tasks 11-15, 17-20 complete; Tasks 02-10 incomplete/partial)
**Goal**: Production-ready identity platform in 11.5 days

---

## Executive Summary

This plan remediates the **critical disconnect** between documentation completion claims and actual implementation status. While the system has impressive advanced features (WebAuthn, adaptive auth, hardware credentials), **foundational OAuth 2.1 flows are broken** due to 16 TODO comments blocking authorization code flow, user login, consent, and token generation.

**Current State**:
- ✅ 9/20 tasks production-ready (Tasks 01, 04, 10.5, 10.6, 11-15, 17-20)
- ⚠️ 5/20 tasks documented complete but have blocking gaps (Tasks 06-08, partial 07, partial 09)
- ❌ 6/20 tasks incomplete or not started (Tasks 02, 03, 05, 09, 10.7, 16)

**Production Blockers**:
1. 🔴 **Authorization code flow** broken (request persistence, PKCE validation missing)
2. 🔴 **User login** returns JSON instead of HTML (no login UI)
3. 🔴 **Token generation** uses placeholder user IDs (not associated with real users)
4. 🔴 **Consent flow** missing decision storage and redirect
5. ⚠️ **Client authentication** missing secret hashing and CRL/OCSP validation
6. ⚠️ **Logout** not implemented (session/token leaks)

**Remediation Approach**:
- **Week 1**: Complete OAuth 2.1 authorization code flow (Tasks 06, 09)
- **Week 2**: Security hardening (Tasks 07, 10)
- **Week 3**: Testing quality and documentation sync (Tasks 03, 10.7, 19)

---

## Remediation Tasks (R01-R11)

### Week 1: Critical Path - OAuth 2.1 Flows

---

#### R01: Complete OAuth 2.1 Authorization Code Flow (Task 06 Remediation)

**Priority**: 🔴 **CRITICAL** - Blocks all OAuth functionality
**Effort**: 16 hours (2 days)
**Owner**: OAuth/OIDC engineer
**Dependencies**: None

**Objectives**:
1. Implement authorization request persistence with PKCE challenge storage
2. Add PKCE code_verifier validation in token endpoint
3. Implement consent decision storage and retrieval
4. Replace placeholder user IDs with real user from login flow
5. Enforce single-use authorization codes
6. Add comprehensive unit and integration tests

**Deliverables**:

**D1.1: Authorization Request Persistence** (4 hours)
- File: `internal/identity/authz/handlers_authorize.go` (lines 112-114)
- Implementation:
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
      return fiber.NewError(fiber.StatusInternalServerError, "Failed to store request")
  }

  // Redirect to IdP login
  loginURL := fmt.Sprintf("%s/oidc/v1/login?request_id=%s", s.cfg.IDP.BaseURL, authRequest.ID)
  return c.Redirect(loginURL, fiber.StatusFound)
  ```
- Repository methods:
  - `CreateAuthorizationRequest(ctx, *AuthorizationRequest) error`
  - `GetAuthorizationRequestByID(ctx, uuid.UUID) (*AuthorizationRequest, error)`
  - `GetAuthorizationRequestByCode(ctx, string) (*AuthorizationRequest, error)`
- Tests:
  - Request persistence validation
  - Expiration enforcement (5 minutes)
  - Redirect to IdP with request_id

**D1.2: PKCE Verifier Validation** (6 hours)
- File: `internal/identity/authz/handlers_token.go` (lines 78-81)
- Implementation:
  ```go
  // Retrieve authorization request by code
  authRequest, err := s.authzRepo.GetAuthorizationRequestByCode(ctx, authCode)
  if err != nil || authRequest == nil {
      return fiber.NewError(fiber.StatusBadRequest, "Invalid authorization code")
  }

  // Validate code not expired or already used
  if time.Now().After(authRequest.ExpiresAt) {
      return fiber.NewError(fiber.StatusBadRequest, "Authorization code expired")
  }
  if authRequest.Used {
      return fiber.NewError(fiber.StatusBadRequest, "Authorization code already used")
  }

  // Validate PKCE code_verifier against stored code_challenge
  if !pkce.ValidateVerifier(codeVerifier, authRequest.CodeChallenge, authRequest.CodeChallengeMethod) {
      return fiber.NewError(fiber.StatusBadRequest, "Invalid PKCE code_verifier")
  }

  // Mark code as used (single-use enforcement)
  authRequest.Used = true
  authRequest.UsedAt = time.Now()
  if err := s.authzRepo.UpdateAuthorizationRequest(ctx, authRequest); err != nil {
      return fiber.NewError(fiber.StatusInternalServerError, "Failed to mark code used")
  }
  ```
- Repository methods:
  - `UpdateAuthorizationRequest(ctx, *AuthorizationRequest) error`
- Tests:
  - PKCE S256 validation success/failure
  - Code expiration detection
  - Single-use enforcement (replay prevention)
  - Code reuse attempt logging

**D1.3: Consent Decision Storage** (5 hours)
- File: `internal/identity/idp/handlers_consent.go` (lines 46-48)
- Implementation:
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
      return fiber.NewError(fiber.StatusInternalServerError, "Failed to store consent")
  }

  // Generate authorization code
  authCode := generateSecureCode() // crypto/rand 32 bytes
  authRequest.Code = authCode
  authRequest.UserID = userID
  authRequest.ConsentID = consent.ID

  if err := s.authzRepo.UpdateAuthorizationRequest(ctx, authRequest); err != nil {
      return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate code")
  }

  // Redirect to callback with code
  callbackURL := fmt.Sprintf("%s?code=%s&state=%s", authRequest.RedirectURI, authCode, authRequest.State)
  return c.Redirect(callbackURL, fiber.StatusFound)
  ```
- Repository methods:
  - `CreateConsent(ctx, *ConsentDecision) error`
  - `GetConsent(ctx, userID, clientID, scope) (*ConsentDecision, error)`
- Helper functions:
  - `generateSecureCode() string` - 32-byte crypto/rand base64url
- Tests:
  - Consent persistence
  - Consent expiration (30 days)
  - Authorization code generation (cryptographically secure)
  - Callback redirect validation

**D1.4: Real User ID in Tokens** (1 hour)
- File: `internal/identity/authz/handlers_token.go` (lines 148-149)
- Implementation:
  ```go
  // Use real user ID from authorization request
  userID := authRequest.UserID
  if userID == googleUuid.Nil {
      return fiber.NewError(fiber.StatusInternalServerError, "Request missing user ID")
  }
  ```
- Tests:
  - Token contains correct user ID
  - Token validation with real user
  - Missing user ID error handling

**Success Criteria**:
- [ ] Authorization requests persist in database with PKCE challenge
- [ ] PKCE code_verifier validation against code_challenge works
- [ ] Consent decisions stored and retrieved correctly
- [ ] Tokens contain real user IDs (no placeholders)
- [ ] Authorization codes enforced single-use
- [ ] Unit test coverage ≥95% on new code
- [ ] Integration tests validate end-to-end authorization code flow
- [ ] Zero TODO comments in production code paths

**Validation**:
```bash
# Run authorization code flow E2E test
go test ./internal/identity/test/e2e -run TestAuthorizationCodeFlow -v

# Verify database persistence
psql -d identity -c "SELECT id, client_id, code_challenge, used FROM authorization_requests;"

# Check token user ID
psql -d identity -c "SELECT id, user_id, client_id FROM access_tokens WHERE user_id != '00000000-0000-0000-0000-000000000000';"
```

---

#### R02: Implement User Login Page (Task 09 Remediation)

**Priority**: 🔴 **CRITICAL** - Users cannot authenticate
**Effort**: 14 hours (1.75 days)
**Owner**: Frontend/UX engineer
**Dependencies**: R01 (authorization request persistence)

**Objectives**:
1. Replace JSON response with HTML login page rendering
2. Add CSRF protection for login form
3. Implement consent redirect logic with skip-consent optimization
4. Style login/consent pages with consistent UX

**Deliverables**:

**D2.1: Login Page HTML Template** (8 hours)
- File: `internal/identity/idp/templates/login.html` (new)
- File: `internal/identity/idp/handlers_login.go` (line 25 fix)
- Implementation:
  ```go
  // Render login page template
  return c.Render("login", fiber.Map{
      "request_id":   requestID,
      "client_id":    clientID,
      "redirect_uri": redirectURI,
      "scope":        scope,
      "csrf_token":   generateCSRFToken(c),
      "error":        c.Query("error"),
      "error_desc":   c.Query("error_description"),
  })
  ```
- Template structure:
  ```html
  <!DOCTYPE html>
  <html>
  <head>
      <title>Sign In</title>
      <link rel="stylesheet" href="/static/css/login.css">
  </head>
  <body>
      <div class="login-container">
          <h1>Sign In</h1>
          {{if .error}}
          <div class="error-message">{{.error_desc}}</div>
          {{end}}
          <form method="POST" action="/oidc/v1/login">
              <input type="hidden" name="request_id" value="{{.request_id}}">
              <input type="hidden" name="csrf_token" value="{{.csrf_token}}">
              <label>Username</label>
              <input type="text" name="username" required>
              <label>Password</label>
              <input type="password" name="password" required>
              <button type="submit">Sign In</button>
          </form>
      </div>
  </body>
  </html>
  ```
- CSS styling:
  - Responsive design (mobile-first)
  - Accessibility (WCAG 2.1 AA)
  - Consistent branding
- CSRF protection:
  - Token generation: `generateCSRFToken(c)` using crypto/rand
  - Token validation on POST
  - Token expiration (15 minutes)
- Tests:
  - Page renders with correct parameters
  - CSRF token validation
  - Error message display
  - Form submission redirects to consent

**D2.2: Consent Redirect with Skip-Consent** (6 hours)
- File: `internal/identity/idp/handlers_login.go` (line 110)
- Implementation:
  ```go
  // Retrieve authorization request
  authRequest, err := s.authzRepo.GetAuthorizationRequestByID(ctx, requestID)
  if err != nil {
      return fiber.NewError(fiber.StatusBadRequest, "Invalid request ID")
  }

  // Check existing consent (skip-consent optimization)
  existingConsent, err := s.consentRepo.GetConsent(ctx, userID, authRequest.ClientID, authRequest.Scope)
  if err == nil && existingConsent != nil && !existingConsent.IsExpired() {
      // Skip consent, generate code directly
      authCode := generateSecureCode()
      authRequest.Code = authCode
      authRequest.UserID = userID
      authRequest.ConsentID = existingConsent.ID

      if err := s.authzRepo.UpdateAuthorizationRequest(ctx, authRequest); err != nil {
          return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate code")
      }

      callbackURL := fmt.Sprintf("%s?code=%s&state=%s", authRequest.RedirectURI, authCode, authRequest.State)
      return c.Redirect(callbackURL, fiber.StatusFound)
  }

  // Redirect to consent page
  consentURL := fmt.Sprintf("/oidc/v1/consent?request_id=%s", requestID)
  return c.Redirect(consentURL, fiber.StatusFound)
  ```
- Tests:
  - Skip-consent when valid consent exists
  - Redirect to consent when no consent exists
  - Consent expiration handling
  - Code generation after skip-consent

**Success Criteria**:
- [ ] Login page renders HTML (not JSON)
- [ ] CSRF protection active and validated
- [ ] Login form submits credentials securely
- [ ] Skip-consent optimization works for returning users
- [ ] New users redirected to consent page
- [ ] Responsive design on mobile/desktop
- [ ] Accessibility validation passes
- [ ] Unit test coverage ≥90% on new code

**Validation**:
```bash
# Manual browser test
open https://localhost:8080/oidc/v1/login?request_id=<uuid>

# Automated test
go test ./internal/identity/idp -run TestHandleLogin -v

# Verify CSRF protection
curl -X POST https://localhost:8080/oidc/v1/login -d "username=test&password=test" # Should fail (no CSRF token)
```

---

### Week 2: Security Hardening

---

#### R03: Client Authentication Security (Task 07 Hardening)

**Priority**: ⚠️ **HIGH** - Security vulnerabilities
**Effort**: 18 hours (2.25 days)
**Owner**: Security engineer
**Dependencies**: None

**Objectives**:
1. Implement bcrypt secret hashing for client credentials
2. Migrate existing plaintext secrets to hashed format
3. Add CRL/OCSP validation for mTLS certificates
4. Implement certificate revocation caching

**Deliverables**:

**D3.1: Client Secret bcrypt Hashing** (8 hours)
- Files:
  - `internal/identity/authz/clientauth/secret_hash.go` (new)
  - `internal/identity/authz/clientauth/client_secret_basic.go` (update)
  - `internal/identity/authz/clientauth/client_secret_post.go` (update)
- Implementation:
  ```go
  import "golang.org/x/crypto/bcrypt"

  // Hash client secret before storing
  func HashClientSecret(secret string) (string, error) {
      hashedBytes, err := bcrypt.GenerateFromPassword([]byte(secret), 12)
      if err != nil {
          return "", fmt.Errorf("failed to hash secret: %w", err)
      }
      return string(hashedBytes), nil
  }

  // Verify client secret against hash
  func VerifyClientSecret(hashedSecret, providedSecret string) bool {
      err := bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(providedSecret))
      return err == nil
  }
  ```
- Migration:
  ```sql
  -- Migrate existing plaintext secrets
  ALTER TABLE clients ADD COLUMN client_secret_hash TEXT;
  UPDATE clients SET client_secret_hash = bcrypt_hash(client_secret) WHERE client_secret IS NOT NULL;
  ALTER TABLE clients DROP COLUMN client_secret;
  ALTER TABLE clients RENAME COLUMN client_secret_hash TO client_secret;
  ```
- Update authentication methods:
  - `client_secret_basic`: verify hash instead of plaintext comparison
  - `client_secret_post`: verify hash instead of plaintext comparison
- Tests:
  - Hash generation and verification
  - Migration script validation
  - Authentication with hashed secrets
  - Failed authentication logging

**D3.2: CRL/OCSP Validation for mTLS** (10 hours)
- Files:
  - `internal/identity/authz/clientauth/cert_revocation.go` (new)
  - `internal/identity/authz/clientauth/tls_client_auth.go` (update)
  - `internal/identity/authz/clientauth/self_signed_auth.go` (update)
- Implementation:
  ```go
  import (
      "crypto/x509"
      "golang.org/x/crypto/ocsp"
  )

  // Validate certificate revocation status
  func ValidateCertificateRevocation(cert *x509.Certificate) error {
      // Check OCSP if available (preferred)
      if len(cert.OCSPServer) > 0 {
          ocspStatus, err := checkOCSP(cert)
          if err != nil {
              return fmt.Errorf("OCSP check failed: %w", err)
          }
          if ocspStatus == ocsp.Revoked {
              return fmt.Errorf("certificate revoked (OCSP)")
          }
          return nil // Good or Unknown
      }

      // Fallback to CRL
      if len(cert.CRLDistributionPoints) > 0 {
          revoked, err := checkCRL(cert)
          if err != nil {
              return fmt.Errorf("CRL check failed: %w", err)
          }
          if revoked {
              return fmt.Errorf("certificate revoked (CRL)")
          }
      }

      return nil
  }

  // OCSP client with caching
  func checkOCSP(cert *x509.Certificate) (int, error) {
      // Check cache first (1 hour TTL)
      if cachedStatus, found := ocspCache.Get(cert.SerialNumber); found {
          return cachedStatus.(int), nil
      }

      // Query OCSP responder
      ocspRequest, err := ocsp.CreateRequest(cert, issuerCert, nil)
      if err != nil {
          return 0, err
      }

      resp, err := http.Post(cert.OCSPServer[0], "application/ocsp-request", bytes.NewReader(ocspRequest))
      if err != nil {
          return 0, err
      }
      defer resp.Body.Close()

      ocspResp, err := ocsp.ParseResponse(resp.Body, issuerCert)
      if err != nil {
          return 0, err
      }

      // Cache result
      ocspCache.Set(cert.SerialNumber, ocspResp.Status, 1*time.Hour)

      return ocspResp.Status, nil
  }

  // CRL client with caching
  func checkCRL(cert *x509.Certificate) (bool, error) {
      // Download CRL from distribution point
      crlURL := cert.CRLDistributionPoints[0]

      // Check cache (24 hour TTL)
      cacheKey := fmt.Sprintf("crl:%s", crlURL)
      if cachedCRL, found := crlCache.Get(cacheKey); found {
          return isCertificateRevoked(cert, cachedCRL.(*pkix.CertificateList)), nil
      }

      // Download CRL
      resp, err := http.Get(crlURL)
      if err != nil {
          return false, err
      }
      defer resp.Body.Close()

      crlBytes, err := io.ReadAll(resp.Body)
      if err != nil {
          return false, err
      }

      crl, err := x509.ParseCRL(crlBytes)
      if err != nil {
          return false, err
      }

      // Cache CRL
      crlCache.Set(cacheKey, crl, 24*time.Hour)

      return isCertificateRevoked(cert, crl), nil
  }
  ```
- Caching:
  - OCSP responses: 1 hour TTL
  - CRL downloads: 24 hour TTL
  - In-memory cache with LRU eviction
- Tests:
  - OCSP validation (good, revoked, unknown)
  - CRL validation (revoked, not revoked)
  - Cache hit/miss scenarios
  - Network failure handling
  - Expired certificate detection

**Success Criteria**:
- [ ] Client secrets hashed with bcrypt cost 12
- [ ] Migration completes without data loss
- [ ] Authentication methods verify hashed secrets
- [ ] OCSP validation queries responders
- [ ] CRL validation downloads and parses lists
- [ ] Revoked certificates rejected
- [ ] OCSP/CRL responses cached
- [ ] Unit test coverage ≥90% on new code
- [ ] Security scan shows no plaintext secrets

**Validation**:
```bash
# Verify secrets hashed in database
psql -d identity -c "SELECT id, client_id, client_secret FROM clients LIMIT 5;" # Should see bcrypt hashes

# Test authentication with hashed secret
go test ./internal/identity/authz/clientauth -run TestClientSecretAuth -v

# Test CRL/OCSP validation
go test ./internal/identity/authz/clientauth -run TestCertificateRevocation -v

# Security scan
gosec ./internal/identity/authz/clientauth/...
```

---

#### R04: Logout Implementation (Task 10 Completion)

**Priority**: ⚠️ **HIGH** - Session/token leaks
**Effort**: 8 hours (1 day)
**Owner**: Backend engineer
**Dependencies**: R01 (token generation with real user IDs)

**Objectives**:
1. Implement logout handler with session cleanup
2. Revoke access and refresh tokens on logout
3. Validate post-logout redirect URIs
4. Add logout audit logging

**Deliverables**:

**D4.1: Logout Handler** (8 hours)
- File: `internal/identity/idp/handlers_logout.go`
- Implementation:
  ```go
  func (s *Service) HandleLogout(c *fiber.Ctx) error {
      ctx := c.Context()

      // Extract session ID
      sessionID := c.Cookies("session_id")
      if sessionID == "" {
          sessionID = c.Query("session_id")
      }

      if sessionID == "" {
          return fiber.NewError(fiber.StatusBadRequest, "No session ID")
      }

      // Retrieve session
      session, err := s.sessionRepo.GetSession(ctx, sessionID)
      if err != nil {
          c.ClearCookie("session_id")
          return c.Redirect("/login?logout=success", fiber.StatusSeeOther)
      }

      // Revoke all tokens
      if err := s.tokenRepo.RevokeTokensBySession(ctx, sessionID); err != nil {
          s.logger.Error("Failed to revoke tokens", "session_id", sessionID, "error", err)
      }

      if err := s.tokenRepo.RevokeRefreshTokensBySession(ctx, sessionID); err != nil {
          s.logger.Error("Failed to revoke refresh tokens", "session_id", sessionID, "error", err)
      }

      // Delete session
      if err := s.sessionRepo.DeleteSession(ctx, sessionID); err != nil {
          return fiber.NewError(fiber.StatusInternalServerError, "Failed to complete logout")
      }

      // Audit log
      s.logger.Info("User logged out",
          "session_id", sessionID,
          "user_id", session.UserID,
          "ip_address", c.IP(),
      )

      // Clear cookie
      c.ClearCookie("session_id")

      // Validate post-logout redirect
      postLogoutRedirect := c.Query("post_logout_redirect_uri")
      if postLogoutRedirect == "" {
          postLogoutRedirect = "/login?logout=success"
      } else {
          // Validate redirect URI against registered client URIs
          if !s.validatePostLogoutRedirect(ctx, postLogoutRedirect) {
              postLogoutRedirect = "/login?logout=success"
          }
      }

      return c.Redirect(postLogoutRedirect, fiber.StatusSeeOther)
  }
  ```
- Repository methods:
  - `GetSession(ctx, sessionID) (*Session, error)`
  - `DeleteSession(ctx, sessionID) error`
  - `RevokeTokensBySession(ctx, sessionID) error`
  - `RevokeRefreshTokensBySession(ctx, sessionID) error`
- Tests:
  - Logout with valid session
  - Logout with expired session
  - Token revocation validation
  - Session cleanup verification
  - Post-logout redirect validation
  - Audit logging

**Success Criteria**:
- [ ] Logout handler cleans up sessions
- [ ] All tokens revoked on logout
- [ ] Post-logout redirect validated
- [ ] Audit logs capture logout events
- [ ] Sessions removed from database
- [ ] Cookies cleared
- [ ] Unit test coverage ≥90%

**Validation**:
```bash
# Manual logout test
curl -X POST https://localhost:8080/oidc/v1/logout -b "session_id=<uuid>"

# Verify session deleted
psql -d identity -c "SELECT * FROM sessions WHERE id = '<uuid>';" # Should return 0 rows

# Verify tokens revoked
psql -d identity -c "SELECT * FROM access_tokens WHERE session_id = '<uuid>' AND revoked = true;"

# Run automated tests
go test ./internal/identity/idp -run TestHandleLogout -v
```

---

### Week 3: Quality & Documentation

---

#### R05: E2E Test Internal Validation (Task 19 Enhancement)

**Priority**: 🔵 **MEDIUM** - Testing quality
**Effort**: 12 hours (1.5 days)
**Owner**: QA engineer
**Dependencies**: R01-R04 (implementation complete)

**Objectives**:
1. Add internal state validation to E2E tests
2. Detect incomplete implementations (TODOs, placeholders)
3. Validate database persistence in OAuth flows
4. Prevent false confidence from mock services

**Deliverables**:

**D5.1: Internal State Validation Tests** (12 hours)
- File: `internal/identity/test/e2e/oauth_flows_internal_test.go` (new)
- Implementation:
  ```go
  // Validate authorization request persisted
  func TestAuthorizationRequestPersistence(t *testing.T) {
      t.Parallel()

      // Make authorization request
      params := url.Values{
          "client_id":             {clientID},
          "redirect_uri":          {redirectURI},
          "scope":                 {"openid profile"},
          "response_type":         {"code"},
          "state":                 {state},
          "code_challenge":        {codeChallenge},
          "code_challenge_method": {"S256"},
      }

      resp, err := client.Get("/oauth2/v1/authorize?" + params.Encode())
      require.NoError(t, err)
      require.Equal(t, 302, resp.StatusCode) // Redirect to login

      // Verify request persisted in database
      var authRequest domain.AuthorizationRequest
      err = db.Where("client_id = ? AND state = ?", clientID, state).First(&authRequest).Error
      require.NoError(t, err, "Authorization request should be persisted")
      require.Equal(t, codeChallenge, authRequest.CodeChallenge)
      require.Equal(t, "S256", authRequest.CodeChallengeMethod)
      require.False(t, authRequest.Used, "Code should not be used yet")
  }

  // Validate PKCE code_verifier validation
  func TestPKCEVerifierValidation(t *testing.T) {
      t.Parallel()

      // Complete login and consent to get authorization code
      authCode := completeLoginAndConsent(t, clientID, redirectURI, codeChallenge, state)

      // Exchange code for token with CORRECT verifier
      tokenResp, err := exchangeCodeForToken(t, authCode, codeVerifier)
      require.NoError(t, err)
      require.Equal(t, 200, tokenResp.StatusCode)

      // Verify code marked as used
      var authRequest domain.AuthorizationRequest
      err = db.Where("code = ?", authCode).First(&authRequest).Error
      require.NoError(t, err)
      require.True(t, authRequest.Used, "Code should be marked as used")
      require.NotNil(t, authRequest.UsedAt)

      // Attempt code reuse (should fail)
      replayResp, err := exchangeCodeForToken(t, authCode, codeVerifier)
      require.NoError(t, err)
      require.Equal(t, 400, replayResp.StatusCode) // Bad Request
  }

  // Validate tokens contain real user IDs (not placeholders)
  func TestTokensContainRealUserIDs(t *testing.T) {
      t.Parallel()

      // Complete OAuth flow
      accessToken := completeOAuthFlow(t, clientID, redirectURI, codeChallenge, codeVerifier)

      // Decode access token (JWT)
      claims, err := jwt.Parse(accessToken)
      require.NoError(t, err)

      // Verify user ID is real (not placeholder UUID)
      userID := claims["sub"].(string)
      require.NotEqual(t, "00000000-0000-0000-0000-000000000000", userID, "Should not be placeholder")

      // Verify user exists in database
      var user domain.User
      err = db.Where("id = ?", userID).First(&user).Error
      require.NoError(t, err, "User should exist")
  }

  // Detect TODO comments in production code paths
  func TestNoTODOsInProductionHandlers(t *testing.T) {
      productionHandlers := []string{
          "internal/identity/authz/handlers_authorize.go",
          "internal/identity/authz/handlers_token.go",
          "internal/identity/idp/handlers_login.go",
          "internal/identity/idp/handlers_consent.go",
          "internal/identity/idp/handlers_logout.go",
      }

      for _, handler := range productionHandlers {
          content, err := os.ReadFile(handler)
          require.NoError(t, err)

          require.NotContains(t, string(content), "// TODO:", "Handler should not have TODOs: %s", handler)
      }
  }
  ```
- Tests:
  - Authorization request persistence
  - PKCE verifier validation
  - Consent decision storage
  - Token real user IDs
  - Single-use code enforcement
  - TODO comment detection
- CI integration:
  - Add to `ci-e2e.yml` workflow
  - Fail builds if internal validation fails

**Success Criteria**:
- [ ] Internal state validated in E2E tests
- [ ] Database persistence checked
- [ ] Placeholder values detected and failed
- [ ] TODO comments detected and failed
- [ ] Tests run in CI/CD pipeline
- [ ] Coverage ≥95% on OAuth critical paths

**Validation**:
```bash
# Run internal validation tests
go test ./internal/identity/test/e2e -run TestInternal -v -tags=e2e

# Verify tests fail with placeholders
# (Temporarily revert R01.4 and confirm test failure)

# Check CI integration
act -j e2e-tests
```

---

#### R06: Configuration Validation (Task 03 Completion)

**Priority**: 🔵 **MEDIUM** - Configuration consistency
**Effort**: 6 hours
**Owner**: DevOps engineer
**Dependencies**: None

**Objectives**:
1. Validate configuration values on startup
2. Enforce cross-service consistency
3. Add configuration tests
4. Document configuration contracts

**Deliverables**:

**D6.1: Configuration Validation** (6 hours)
- File: `internal/identity/config/validation.go` (new)
- Implementation:
  ```go
  func ValidateConfig(cfg *Config) error {
      // Validate port ranges
      if cfg.AuthZ.Port < 1024 || cfg.AuthZ.Port > 65535 {
          return fmt.Errorf("invalid AuthZ port: %d", cfg.AuthZ.Port)
      }
      if cfg.IDP.Port < 1024 || cfg.IDP.Port > 65535 {
          return fmt.Errorf("invalid IdP port: %d", cfg.IDP.Port)
      }

      // Validate URLs parseable
      if _, err := url.Parse(cfg.AuthZ.BaseURL); err != nil {
          return fmt.Errorf("invalid AuthZ base URL: %w", err)
      }
      if _, err := url.Parse(cfg.IDP.BaseURL); err != nil {
          return fmt.Errorf("invalid IdP base URL: %w", err)
      }

      // Validate cross-service consistency
      if cfg.IDP.AuthZURL != cfg.AuthZ.BaseURL {
          return fmt.Errorf("IdP AuthZ URL mismatch: %s != %s", cfg.IDP.AuthZURL, cfg.AuthZ.BaseURL)
      }

      // Validate token lifetimes
      if cfg.Tokens.AccessTokenLifetime > cfg.Tokens.RefreshTokenLifetime {
          return fmt.Errorf("access token lifetime exceeds refresh token lifetime")
      }
      if cfg.Tokens.AccessTokenLifetime < 1*time.Minute {
          return fmt.Errorf("access token lifetime too short: %s", cfg.Tokens.AccessTokenLifetime)
      }

      // Validate database connection string
      if cfg.Database.DSN == "" {
          return fmt.Errorf("database DSN not configured")
      }

      return nil
  }
  ```
- Startup integration:
  - Call `ValidateConfig()` before starting services
  - Fail fast on validation errors
  - Log validation errors clearly
- Tests:
  - Valid configuration passes
  - Invalid port ranges rejected
  - Malformed URLs rejected
  - Cross-service mismatches rejected
  - Token lifetime validations

**Success Criteria**:
- [ ] Configuration validated on startup
- [ ] Invalid configs fail fast
- [ ] Cross-service consistency enforced
- [ ] Tests cover all validation rules
- [ ] Documentation updated

**Validation**:
```bash
# Test with invalid config
cat > configs/identity/authz/invalid.yml <<EOF
port: 99999  # Invalid port
base_url: "not-a-url"
EOF

./identity start --config configs/identity/authz/invalid.yml # Should fail with clear error

# Run validation tests
go test ./internal/identity/config -run TestValidateConfig -v
```

---

#### R07: OpenAPI Spec Synchronization (Task 10.7 Completion)

**Priority**: 🔵 **MEDIUM** - Documentation consistency
**Effort**: 8 hours (1 day)
**Owner**: API engineer
**Dependencies**: R01-R04 (implementation complete)

**Objectives**:
1. Update OpenAPI specs to match current implementation
2. Regenerate client libraries
3. Update Swagger UI
4. Add CI validation to prevent spec drift

**Deliverables**:

**D7.1: OpenAPI Spec Updates** (8 hours)
- Files:
  - `api/identity/authz/openapi.yml`
  - `api/identity/idp/openapi.yml`
  - `api/identity/rs/openapi.yml`
- Updates:
  - `/oauth2/v1/authorize`: Add `code_challenge`, `code_challenge_method` parameters (PKCE)
  - `/oauth2/v1/token`: Add `code_verifier` parameter, document error responses
  - `/oidc/v1/login`: Change response from `application/json` to `text/html`
  - `/oidc/v1/consent`: Document consent flow, scope approval
  - `/oidc/v1/logout`: Document logout endpoint, post-logout redirect
- Client library regeneration:
  ```bash
  # Regenerate Go client
  oapi-codegen -config api/identity/authz/oapi-config.yml api/identity/authz/openapi.yml > api/identity/authz/client.gen.go
  ```
- Swagger UI update:
  - Deploy updated specs to `/ui/swagger`
  - Test interactive API documentation
- CI validation:
  ```yaml
  # .github/workflows/ci-openapi.yml
  - name: Validate OpenAPI specs
    run: |
      # Lint OpenAPI specs
      spectral lint api/identity/authz/openapi.yml
      spectral lint api/identity/idp/openapi.yml

      # Validate specs match implementation
      go run ./scripts/validate-openapi.go
  ```
- Tests:
  - Spec validation (syntax, semantics)
  - Client library generation
  - Swagger UI rendering

**Success Criteria**:
- [ ] OpenAPI specs match implementation
- [ ] Client libraries regenerated
- [ ] Swagger UI updated
- [ ] CI validates spec consistency
- [ ] Documentation reflects reality

**Validation**:
```bash
# Lint OpenAPI specs
spectral lint api/identity/authz/openapi.yml

# Regenerate client libraries
make generate-openapi

# Test Swagger UI
open https://localhost:8080/ui/swagger

# Run CI validation
act -j openapi-validation
```

---

## Remediation Timeline

### Week 1: Critical Path (30 hours)

| Day | Task | Hours | Status |
|-----|------|-------|--------|
| Mon | R01: Authorization request persistence | 4 | ⏳ Pending |
| Mon-Tue | R01: PKCE verifier validation | 6 | ⏳ Pending |
| Tue | R01: Consent decision storage | 5 | ⏳ Pending |
| Tue | R01: Real user ID in tokens | 1 | ⏳ Pending |
| Wed | R02: Login page HTML template | 8 | ⏳ Pending |
| Thu | R02: Consent redirect logic | 6 | ⏳ Pending |

**Deliverable**: End-to-end OAuth authorization code flow functional

---

### Week 2: Security Hardening (26 hours)

| Day | Task | Hours | Status |
|-----|------|-------|--------|
| Fri | R03: Client secret bcrypt hashing | 8 | ⏳ Pending |
| Mon | R03: CRL/OCSP validation | 10 | ⏳ Pending |
| Tue | R04: Logout implementation | 8 | ⏳ Pending |

**Deliverable**: Production-grade security hardening complete

---

### Week 3: Quality & Documentation (26 hours)

| Day | Task | Hours | Status |
|-----|------|-------|--------|
| Wed | R05: E2E test internal validation | 12 | ⏳ Pending |
| Thu | R06: Configuration validation | 6 | ⏳ Pending |
| Fri | R07: OpenAPI spec synchronization | 8 | ⏳ Pending |

**Deliverable**: Testing quality and documentation consistency validated

---

## Additional Security Tasks (10 hours)

### R08: Rate Limiting (Priority: 🔧 LOW)

**Effort**: 4 hours
**Files**: `internal/identity/middleware/rate_limit.go`

Implementation:
```go
import "github.com/gofiber/fiber/v2/middleware/limiter"

func RateLimitMiddleware() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        10,
        Expiration: 1 * time.Second,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return fiber.NewError(fiber.StatusTooManyRequests, "Rate limit exceeded")
        },
    })
}
```

Apply to endpoints:
- `/oauth2/v1/token` (10 req/sec per IP)
- `/oidc/v1/login` (5 req/sec per IP)

---

### R09: Audit Logging OAuth Flows (Priority: 🔧 LOW)

**Effort**: 6 hours
**Files**: All OAuth handlers

Add audit logs:
```go
s.logger.Info("Authorization code granted",
    "client_id", clientID,
    "user_id", userID,
    "scope", scope,
    "code_challenge_method", challengeMethod,
    "ip_address", c.IP(),
)
```

Events to log:
- Authorization request
- Authorization code generation
- Token issuance (access, refresh, ID)
- Token revocation
- Login success/failure
- Logout

---

## FIPS-140-3 Compliance & Crypto Agility (MANDATORY)

**Requirement**: All cryptographic operations, defaults, and test harnesses MUST be FIPS-140-3 compliant by default. `bcrypt` is NOT FIPS-approved and therefore cannot be the default hashing algorithm for production secrets.

Key obligations:
- **Configurable algorithms**: Every crypto operation (password hashing, key derivation, signing, encryption, MAC, key wrapping) MUST support configurable algorithms and parameters. Defaults must be FIPS-140-3 approved (e.g., PBKDF2-HMAC-SHA256 for KDF/secret hashing with strong iteration counts, AES-GCM with 128/256-bit keys, RSA >= 2048, ECDSA on approved curves such as P-256/P-384, EdDSA where FIPS-approved alternatives exist).
- **Algorithm agility**: Implement an algorithm registry and configuration layer so operators can swap algorithms and parameters without code changes. Provide migration helpers for key and secret re-encryption/hashing.
- **FIPS mode**: Add a `--crypto-mode=fips` startup flag and configuration entry which enforces only FIPS-approved algorithms and rejects non-compliant configurations at startup.
- **Testing**: Unit and integration tests must include FIPS-mode test matrices verifying that code paths behave correctly under FIPS defaults and reject disallowed algorithms.

Implementation notes:
- Replace `bcrypt` defaults with PBKDF2-HMAC-SHA256 (configurable) or KDFs approved by FIPS 140-3. If Argon2 is preferred for memory-hard hashing, provide it as an opt-in non-FIPS mode and document tradeoffs.
- Provide wrappers in `internal/crypto` that expose `HashSecret`, `VerifySecret`, `DeriveKey`, `Sign`, `Verify` and route to configured algorithms.

Acceptance criteria:
- FIPS mode startup validates configuration and exits on non-FIPS algorithms.
- All identity tasks use `internal/crypto` wrappers.
- CI includes matrix testing for FIPS-mode and non-FIPS-mode.

---

## Cross-Database Testing & PostgreSQL 18.1 Requirement

**Requirement**: Cross-DB compatibility must be validated. PostgreSQL 18.1 is the canonical integration test DB and must be used in CI where Postgres-specific behavior is tested.

Policy:
- Unit tests may run against SQLite for speed, but any DB-compatible behavior that differs (SQL types, transactions, locking, JSONB semantics, concurrency) must have targeted integration tests that run against a **PostgreSQL 18.1** Docker container in CI. The repository already references `postgres:18` images in compose files; CI must pin to `postgres:18.1` for integration runs when supported.
- Classify tests:
  - Unit tests: fast, isolated, may use in-memory/SQLite.
  - Integration tests: database-backed tests that run against PostgreSQL 18.1 container (annotated `//go:build integration` or similar). These tests are required for schema, migrations, transaction semantics and concurrency behaviors.

Validation steps:
```bash
# Start Postgres 18.1 for integration tests
docker run --rm -d --name ci-postgres-18.1 -e POSTGRES_USER=usr -e POSTGRES_PASSWORD=pwd -e POSTGRES_DB=identity postgres:18.1

# Run integration tests
go test ./internal/identity/... -tags=integration -run TestIntegration -v
```

---

## Coverage, Benchmarking, and Fuzz Testing Requirements

**Requirement**: Code coverage, benchmarks, and fuzz tests must be part of the remediation effort and validated in CI.

Policy:
- **Coverage**: Identity packages must meet a minimum of 80% statement coverage; critical cryptographic and OAuth core packages should target ≥90%. Add coverage checks to CI and fail builds that regress coverage below thresholds.
- **Benchmarks**: Add `*_bench_test.go` for performance-critical paths (token issuance, PKCE verification, crypto key wraps). Run benchmarks in CI optionally or nightly, and record baselines.
- **Fuzz tests**: Add fuzz tests for parsers, token input validation, JWS/JWE handling, and any untrusted input surfaces (e.g., client metadata parsing). Use `go test -fuzz` in CI/nightly runs.

Acceptance criteria:
- Coverage gates in CI for identity packages.
- Benchmarks added for token and crypto hot paths with baselines stored in `test-output/benchmarks`.
- Fuzz tests added and run nightly (or per PR for changed packages) with failure triage documented.

---

## Hardware Crypto Support (PKCS#11 & Tokens)

**Requirement**: Support multiple hardware crypto backends via PKCS#11 and token integrations.

Policy & implementation:
- Provide a PKCS#11 abstraction in `internal/crypto/pkcs11backend` with connectors for SoftHSM2 (for CI and dev), and configuration hooks for enterprise HSMs (Utimaco, Luna, nShield). Support multiple slot discovery and PIN/passphrase mechanisms.
- Provide token support for YubiKey (CTAP/U2F/WebAuthn attestation flows) and other vendor SDKs where required.
- The system must be able to fallback to software keys when hardware is not configured, but in FIPS mode prefer HSM-backed keys where available.
- Provide example Docker Compose profiles for testing SoftHSM2 and for YubiKey passthrough (document host requirements).

Tests:
- SoftHSM2 integration tests in CI for key generation, signing, and key import/export where supported.
- Manual tests documented for YubiKey (developer has a YubiKey installed) and CI harnesses that can optionally access a YubiKey if provided by the runner.

---

## WebAuthn / Passkeys Testing Strategy

**Requirement**: WebAuthn testing must include virtual authenticators for headless CI and optional real-device testing (YubiKey).

Policy & Implementation:
- Use Chrome in Docker with Selenium/Playwright and the Chrome Virtual Authenticator API to simulate passkeys in CI for registration and authentication flows.
- Provide a `webauthn-e2e` compose profile that runs a Chrome container with Virtual Authenticator enabled and a test harness that exercises registration/auth flows.
- Provide optional real-device tests that can be run locally (or in a dedicated hardware lab) that use the developer's YubiKey via USB passthrough.
- Default server crypto parameters for WebAuthn: use ECP-384 (P-384) as default for attestation and assertion; support configuration for ECDSA P-256, RSA, Ed25519, and others via `internal/crypto` registry.

---

## OpenTelemetry Integration & Testing

**Requirement**: OTEL must forward to `opentelemetry-collector-contrib` and onward to `grafana/otel-lgtm` for dashboarding. Integration tests must validate traces/metrics ingestion.

Policy & Implementation:
- Compose file includes `opentelemetry-collector-contrib` and `grafana/otel-lgtm`. CI must bring up the collector and a lightweight Grafana (or LGTM) stack for integration tests.
- Provide a smoke test that sends a sample trace and validates via OTLP HTTP that the collector accepted the data and Grafana datasource is available.

---

## Pre-commit / Pre-push Enforcement

**Requirement**: During remediation, ALL code changes must pass linters, formatters, and pre-commit hooks. Commits MUST NOT use `--no-verify` in remediation branches and CI must fail on pre-commit violations.

Policy:
- Update CI to run the same pre-commit hooks used locally. Failing hooks block merges.
- Each remediation task includes a checklist item: "pre-commit hooks pass locally and in CI".

---

## Final Validation Checklist

### Functional Requirements

**OAuth 2.1 Authorization Code Flow**:
- [ ] Authorization request persisted with PKCE challenge
- [ ] Login page renders HTML (not JSON)
- [ ] Consent page stores approval decision
- [ ] PKCE code_verifier validated against code_challenge
- [ ] Authorization code single-use enforced
- [ ] Tokens contain real user ID (not placeholder)
- [ ] Authorization code flow works end-to-end without mocks

**Client Authentication**:
- [ ] Client secrets hashed with bcrypt
- [ ] mTLS certificates validated (CRL/OCSP)
- [ ] Authentication failures logged
- [ ] Migration completes without data loss

**Session Management**:
- [ ] Logout revokes all tokens
- [ ] Sessions cleaned up on logout
- [ ] Post-logout redirect validated
- [ ] Session lifecycle managed correctly

**Testing**:
- [ ] E2E tests validate internal state (not just HTTP responses)
- [ ] Unit tests cover all new code paths (≥90%)
- [ ] Integration tests validate database persistence
- [ ] Tests detect placeholders and fail

**Security**:
- [ ] Rate limiting active on authentication endpoints
- [ ] Audit logging covers all OAuth operations
- [ ] No TODO comments in production code paths
- [ ] Security scan passes (gosec, staticcheck)

**Documentation**:
- [ ] OpenAPI specs match implementation
- [ ] Configuration validation enforced
- [ ] Runbooks updated with new flows
- [ ] README updated with deployment instructions

---

## Monitoring Post-Remediation

### Key Metrics

**Authorization Code Flow**:
- `authz_requests_persisted_total`
- `pkce_validations_total{result="success|failure"}`
- `authz_code_reuse_attempts_total` (should be zero in production)

**Token Service**:
- `tokens_issued_total{type="access|refresh|id"}`
- `tokens_placeholder_user_total` (should be zero)
- `tokens_revoked_total`

**Session Management**:
- `sessions_active_gauge`
- `logout_events_total{result="success|failure"}`
- `sessions_cleaned_total`

**Security**:
- `rate_limit_exceeded_total{endpoint}`
- `client_auth_failures_total{method}`
- `cert_validation_failures_total{type="crl|ocsp"}`

---

## Success Criteria

### Pre-Deployment

- [ ] All remediation tasks (R01-R07) complete
- [ ] Functional validation checklist 100% complete
- [ ] Security scan passes with zero critical/high findings
- [ ] E2E tests pass (internal validation enabled)
- [ ] Documentation synchronized with implementation
- [ ] Performance baseline established
- [ ] DR procedures documented and tested

### Post-Deployment

- [ ] OAuth flows operational in production
- [ ] User authentication successful
- [ ] Token issuance with real user IDs
- [ ] Logout cleans up sessions/tokens
- [ ] Security monitoring alerts configured
- [ ] Metrics dashboards deployed
- [ ] Runbooks validated in production
- [ ] Incident response team trained

---

## Conclusion

This 11.5-day remediation plan transforms the Identity V2 system from **"impressive demo with advanced features"** to **"production-ready identity platform"**. The critical path (Week 1) completes foundational OAuth 2.1 flows, unblocking user authentication and authorization. Security hardening (Week 2) addresses vulnerabilities and missing logout. Quality improvements (Week 3) prevent future regressions and synchronize documentation.

**Key Insight**: Once critical path complete, advanced features (WebAuthn, adaptive auth, hardware credentials) become immediately usable, delivering exceptional security capabilities atop a solid OAuth 2.1 foundation.

**Next Action**: Begin R01 (Complete OAuth 2.1 Authorization Code Flow) to unblock all downstream work.
