# Task R02: Implement User Login Page (Task 09 Remediation)

**Priority**: 🔴 **CRITICAL** - Users cannot authenticate
**Effort**: 14 hours (1.75 days)
**Owner**: Frontend/IdP engineer
**Dependencies**: R01 (authorization request persistence)
**Source**: GAP-ANALYSIS-DETAILED.md Priority 2 issues

---

## Problem Statement

The user login page returns JSON instead of HTML, preventing users from authenticating:

1. **Login Page Rendering**: `/oidc/v1/login` returns JSON object instead of HTML form (line 25 TODO)
2. **Consent Redirect**: After login, handler doesn't redirect to consent or callback (line 110 TODO)
3. **No Login UI**: Users see raw JSON data instead of username/password form
4. **Consent Skip**: No logic to skip consent if user previously consented to same client/scope

**Impact**: Users cannot authenticate through browser, blocking all OAuth flows requiring user interaction.

---

## Acceptance Criteria

- [ ] Login page renders HTML form (username, password, CSRF token)
- [ ] Login form styled with CSS framework (Bootstrap or custom)
- [ ] CSRF token generated and validated on form submission
- [ ] Login errors displayed to user (invalid credentials, account locked, etc.)
- [ ] After successful login, check for existing consent
- [ ] If consent exists and not expired, skip consent page and generate authorization code
- [ ] If no consent or expired, redirect to consent page
- [ ] All HTML templates have proper HTML5 structure and accessibility attributes
- [ ] Pre-commit hooks pass (golangci-lint, cspell, formatting)

---

## Implementation Steps

### Step 1: Create HTML Template Infrastructure (3 hours)

**Directory**: `internal/identity/idp/templates/`

**File**: `login.html`
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sign In - Identity Provider</title>
    <link rel="stylesheet" href="/static/css/auth.css">
</head>
<body>
    <div class="auth-container">
        <div class="auth-card">
            <h1>Sign In</h1>

            {{if .error}}
            <div class="alert alert-error">
                <strong>Error:</strong> {{.error_desc}}
            </div>
            {{end}}

            <form method="POST" action="/oidc/v1/login">
                <input type="hidden" name="request_id" value="{{.request_id}}">
                <input type="hidden" name="csrf_token" value="{{.csrf_token}}">

                <div class="form-group">
                    <label for="username">Username</label>
                    <input type="text" id="username" name="username" required autofocus>
                </div>

                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required>
                </div>

                <button type="submit" class="btn btn-primary">Sign In</button>
            </form>

            <div class="auth-footer">
                <a href="/oidc/v1/forgot-password">Forgot password?</a>
            </div>
        </div>
    </div>
</body>
</html>
```

**File**: `static/css/auth.css`
```css
/* Minimal auth form styling */
.auth-container {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 100vh;
    background: #f5f5f5;
}

.auth-card {
    background: white;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
    width: 100%;
    max-width: 400px;
}

.auth-card h1 {
    margin-bottom: 1.5rem;
    font-size: 1.5rem;
    text-align: center;
}

.form-group {
    margin-bottom: 1rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
}

.form-group input {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid #ddd;
    border-radius: 4px;
}

.btn {
    width: 100%;
    padding: 0.75rem;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
}

.btn-primary {
    background: #007bff;
    color: white;
}

.alert {
    padding: 1rem;
    margin-bottom: 1rem;
    border-radius: 4px;
}

.alert-error {
    background: #f8d7da;
    color: #721c24;
    border: 1px solid #f5c6cb;
}
```

**Tests**: Validate template rendering with mock data.

---

### Step 2: Implement Login Page Rendering (4 hours)

**File**: `internal/identity/idp/handlers_login.go`

**Replace line 25**:
```go
// Render login page HTML template
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

**CSRF Token Generation**:
```go
func generateCSRFToken(c *fiber.Ctx) string {
    // Generate 32-byte random token
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        panic(fmt.Sprintf("Failed to generate CSRF token: %v", err))
    }
    token := base64.RawURLEncoding.EncodeToString(b)

    // Store in session or cookie with 15-minute expiration
    c.Cookie(&fiber.Cookie{
        Name:     "csrf_token",
        Value:    token,
        MaxAge:   900, // 15 minutes
        HTTPOnly: true,
        Secure:   true,
        SameSite: "Lax",
    })

    return token
}

func validateCSRFToken(c *fiber.Ctx, providedToken string) bool {
    sessionToken := c.Cookies("csrf_token")
    return sessionToken != "" && sessionToken == providedToken
}
```

**Template Engine Setup** (in service initialization):
```go
import "html/template"

// Configure Fiber template engine
engine := html.New("./internal/identity/idp/templates", ".html")
app := fiber.New(fiber.Config{
    Views: engine,
})
```

**Tests**: Verify template rendering, CSRF token generation/validation, error message display.

---

### Step 3: Implement Consent Skip Logic (6 hours)

**File**: `internal/identity/idp/handlers_login.go`

**Replace line 110**:
```go
// Retrieve authorization request to determine next step
authRequest, err := s.authzRepo.GetAuthorizationRequestByID(ctx, requestID)
if err != nil {
    s.logger.Error("Invalid request ID", "request_id", requestID, "error", err)
    return c.Redirect("/oidc/v1/login?error=invalid_request&error_description=Invalid+request+ID", fiber.StatusSeeOther)
}

// Check if user has previously consented to this client/scope combination
existingConsent, err := s.consentRepo.GetConsent(ctx, userID, authRequest.ClientID, authRequest.Scope)
if err == nil && existingConsent != nil && !existingConsent.IsExpired() {
    // Skip consent, generate authorization code directly
    authCode := generateSecureCode()
    authRequest.Code = authCode
    authRequest.UserID = userID
    authRequest.ConsentID = existingConsent.ID

    if err := s.authzRepo.UpdateAuthorizationRequest(ctx, authRequest); err != nil {
        s.logger.Error("Failed to generate authorization code", "request_id", requestID, "error", err)
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to complete login")
    }

    // Redirect to callback with code and state
    callbackURL := fmt.Sprintf("%s?code=%s&state=%s", authRequest.RedirectURI, authCode, authRequest.State)
    s.logger.Info("Login successful, consent skipped", "user_id", userID, "client_id", authRequest.ClientID)
    return c.Redirect(callbackURL, fiber.StatusFound)
}

// No existing consent or consent expired, redirect to consent page
consentURL := fmt.Sprintf("/oidc/v1/consent?request_id=%s", requestID)
s.logger.Info("Login successful, redirecting to consent", "user_id", userID, "client_id", authRequest.ClientID)
return c.Redirect(consentURL, fiber.StatusFound)
```

**Consent Repository Method**:
```go
type ConsentRepository interface {
    GetConsent(ctx context.Context, userID googleUuid.UUID, clientID string, scope string) (*domain.ConsentDecision, error)
    CreateConsent(ctx context.Context, consent *domain.ConsentDecision) error
}
```

**Domain Method**:
```go
func (c *ConsentDecision) IsExpired() bool {
    return time.Now().After(c.ExpiresAt)
}
```

**Tests**: Verify consent lookup, skip logic, redirect to consent when needed, redirect to callback when consent exists.

---

### Step 4: Add Static File Serving (1 hour)

**File**: `internal/identity/idp/service.go` (initialization)

```go
// Serve static files (CSS, JS, images)
app.Static("/static", "./internal/identity/idp/static")
```

**Tests**: Verify CSS file accessible at `/static/css/auth.css`.

---

## Testing Requirements

### Unit Tests

**File**: `internal/identity/idp/handlers_login_test.go`
- Login page rendering (table-driven test with various error states)
- CSRF token generation and validation
- Consent skip logic (existing consent, expired consent, no consent)
- Redirect URL construction (consent page, callback URL)

### Integration Tests

**File**: `internal/identity/test/e2e/oauth_login_flow_test.go`
- End-to-end login flow:
  1. Client initiates authorization request
  2. User redirected to login page (verify HTML response, not JSON)
  3. User submits credentials with CSRF token
  4. Server validates CSRF token
  5. If consent exists, redirect to callback with code
  6. If no consent, redirect to consent page

### Manual Testing

- [ ] Login page renders correctly in Chrome, Firefox, Safari
- [ ] CSRF token validated on form submission
- [ ] Invalid credentials show error message
- [ ] Consent skip works for returning users
- [ ] CSS styling displays correctly
- [ ] Accessibility: keyboard navigation, screen reader support

---

## Pre-commit Enforcement

- Run `golangci-lint run --fix` on all modified files
- Fix all linting errors (wsl, godot, mnd, errcheck)
- Ensure test coverage ≥80% for handler code
- Run `go test ./internal/identity/idp/... -cover` before committing
- Validate HTML templates with W3C validator
- Commit with conventional commit message: `feat(idp): implement user login page with HTML rendering and CSRF protection`

---

## Validation

**Success Criteria**:
- [ ] Login page renders HTML form (not JSON)
- [ ] CSRF token generation and validation works
- [ ] Consent skip logic correctly handles existing/expired consent
- [ ] Redirects to consent page when no consent exists
- [ ] Redirects to callback when consent exists
- [ ] All unit tests pass
- [ ] Integration test completes full login flow
- [ ] Pre-commit hooks pass
- [ ] Code coverage ≥80% for handlers

---

## References

- GAP-ANALYSIS-DETAILED.md: Priority 2 issues (lines 178-244)
- REMEDIATION-MASTER-PLAN-2025.md: R02 section (lines 222-334)
- OIDC Core 1.0: https://openid.net/specs/openid-connect-core-1_0.html
- Fiber template engine: https://docs.gofiber.io/guide/templates
