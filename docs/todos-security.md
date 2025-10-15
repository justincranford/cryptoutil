# Cryptoutil Security & Authentication TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: October 14, 2025
**Status**: Critical OAuth 2.0 implementation planning underway. Security hardening tasks remain active.

---

## 🔴 CRITICAL - OAuth 2.0 Implementation Planning

### Task O1: Design OAuth 2.0 Authorization Code Flow for User vs Machine Access
- **Description**: Implement separate OAuth 2.0 flows for browser users vs service machines
- **Architecture Decision**: Users get bearer tokens for `/browser/api/v1/**`, machines get tokens for `/service/api/v1/**`
- **Current State**: Both API paths currently accessible without authentication differentiation
- **Action Items**:
  - Design OAuth 2.0 Authorization Code flow for browser users (redirect-based)
  - Design OAuth 2.0 Client Credentials flow for service machines (direct token exchange)
  - Implement token validation middleware that checks token scope/audience
  - Update OpenAPI specs to reflect authentication requirements
  - Add OAuth 2.0 provider integration (Auth0, Keycloak, or custom)
- **Files**: Authentication middleware, OAuth provider configuration, OpenAPI specs
- **Expected Outcome**: Secure separation between user and machine API access
- **Priority**: CRITICAL - Foundation for API security model
- **Timeline**: Q4 2025 implementation

### Task O2: Update API Documentation for OAuth 2.0
- **Description**: Update OpenAPI specs to reflect OAuth 2.0 authentication requirements
- **Current State**: APIs currently have no authentication documented
- **Action Items**:
  - Add OAuth 2.0 security schemes to OpenAPI specs
  - Document different flows for browser vs service APIs
  - Update error responses to include authentication errors
  - Add authentication examples in Swagger UI
- **Files**: `api/openapi_spec_*.yaml`, Swagger UI configuration
- **Expected Outcome**: Clear API authentication documentation
- **Priority**: MEDIUM - Documentation follows implementation
- **Dependencies**: Task O1 completion

### Task O3: Implement Token Scope Validation Middleware
- **Description**: Add middleware to validate OAuth tokens have appropriate scope for endpoint access
- **Current State**: No authentication middleware implemented
- **Action Items**:
  - Create JWT validation middleware
  - Implement scope checking (`browser:api` vs `service:api`)
  - Add token refresh handling
  - Implement proper 401/403 error responses
- **Files**: Authentication middleware, error handling
- **Expected Outcome**: Runtime enforcement of API access separation
- **Priority**: MEDIUM - Security enforcement
- **Dependencies**: Task O1 completion

---

## 🟡 MEDIUM - Security Hardening & Compliance

### Task S1: Fix Cookie HttpOnly Flag Security Issue
- **Description**: ZAP detected cookies without HttpOnly flag set (Rule 10010)
- **Root Cause**: CSRF and other security cookies may not have HttpOnly enabled
- **Current State**: CSRF middleware uses configurable `CookieHTTPOnly` setting
- **Action Items**:
  - Ensure all security-related cookies have `HttpOnly: true` in `application_listener.go`
  - Verify CSRF token cookies are properly configured
  - Test that cookies are HttpOnly in browser developer tools
- **Files**: `internal/server/application/application_listener.go` (CSRF middleware configuration)
- **Expected Outcome**: All cookies flagged by ZAP rule 10010 have HttpOnly enabled
- **Priority**: Medium - Cookie security hardening
- **ZAP Reference**: WARN-NEW: Cookie No HttpOnly Flag [10010] x 6

### Task S2: Investigate JSON Parsing Issues in API Endpoints
- **Description**: ZAP VariantJSONQuery failing to parse request bodies
- **Root Cause**: API endpoints expect JSON but receive string data
- **Current State**: Multiple WARN messages about invalid JSON parsing
- **Action Items**:
  - Identify endpoints with JSON parsing issues
  - Review API request validation and content-type handling
  - Fix API contracts to properly handle JSON vs string inputs
  - Test API endpoints with proper JSON payloads
- **Files**: API handler files, OpenAPI specifications
- **Expected Outcome**: All JSON API endpoints properly parse JSON request bodies
- **Priority**: Medium - API contract consistency
- **ZAP Reference**: Multiple WARN messages about VariantJSONQuery parsing failures

### Task S3: Fix golangci-lint staticcheck Integration Issue
- **Description**: golangci-lint staticcheck integration is broken and produces no issues despite being enabled
- **Root Cause**: staticcheck is enabled in .golangci.yml but integration fails silently
- **Current State**: golangci-lint produces 39KB SARIF vs standalone staticcheck 316KB with comprehensive analysis
- **Action Items**:
  - Investigate why golangci-lint staticcheck integration fails
  - Test standalone staticcheck vs golangci-lint staticcheck output
  - Fix integration or document limitation in CI workflow
  - Ensure staticcheck security findings are properly reported to GitHub Security tab
- **Files**: `.golangci.yml`, `.github/workflows/ci.yml` (staticcheck step)
- **Expected Outcome**: Either fix golangci-lint integration or clearly document why separate staticcheck run is required
- **Priority**: Medium - Code quality and security scanning reliability
- **Timeline**: Q4 2025 investigation and fix

### Task S4: Advanced Threat Modeling Documentation
- **Description**: Create comprehensive threat modeling documentation for advanced security analysis
- **Current State**: Basic security practices implemented but no formal threat modeling
- **Action Items**:
  - Document threat modeling methodology (STRIDE, PASTA, or OCTAVE)
  - Identify potential attack vectors and mitigation strategies
  - Document security boundaries and trust zones
  - Create threat modeling diagrams and risk assessments
- **Files**: `docs/security-threat-model.md` or similar
- **Expected Outcome**: Comprehensive security analysis framework
- **Priority**: Low - Advanced security documentation

---

## Appendix: OAuth 2.0 Implementation Options

### Recommended Architecture: Hybrid Approach

#### Option 1: Ory Hydra + Custom Provider (RECOMMENDED)
**Best for your requirements - supports both custom auth provider and GitHub**

**Dependencies to Add:**
```go
require (
    github.com/ory/hydra-client-go v2.2.0+incompatible
    github.com/coreos/go-oidc/v3 v3.14.1
    github.com/zitadel/oidc/v2 v2.12.3
    golang.org/x/oauth2 v0.28.1
)
```

**Pros:**
- ✅ Ory Hydra is Go-based, production-ready OAuth2/OIDC server
- ✅ Supports Authorization Code + PKCE natively
- ✅ Easy integration with GitHub OAuth2
- ✅ Your custom provider can delegate to Hydra
- ✅ Excellent security defaults
- ✅ Active CNCF project

**Implementation Example:**
```go
// Hydra client setup
hydraAdmin := hydra.NewAPIClient(&hydra.Configuration{
    Host:   "hydra.yourdomain.com",
    Scheme: "https",
})

// GitHub OAuth2 config
githubOAuth2 := &oauth2.Config{
    ClientID:     "your-github-client-id",
    ClientSecret: "your-github-client-secret",
    Scopes:       []string{"user:email"},
    Endpoint:     github.Endpoint,
}

// Custom provider using Hydra
func (h *Handler) handleLogin(c *fiber.Ctx) error {
    // Redirect to Hydra with PKCE
    loginURL := fmt.Sprintf("%s/oauth2/auth?%s",
        hydraURL,
        url.Values{
            "client_id":     {clientID},
            "response_type": {"code"},
            "scope":         {"openid profile email"},
            "redirect_uri":  {redirectURI},
            "code_challenge": {pkceChallenge},
            "code_challenge_method": {"S256"},
        }.Encode())

    return c.Redirect(loginURL)
}
```

#### Option 2: Zitadel OIDC Framework (Pure Go)
**Excellent for custom implementation with GitHub support**

**Dependencies:**
```go
require github.com/zitadel/oidc/v2 v2.12.3
```

**Pros:**
- ✅ Pure Go implementation
- ✅ Built-in Authorization Code + PKCE support
- ✅ Excellent for custom auth providers
- ✅ GitHub OAuth2 integration
- ✅ Modern security standards

**Implementation:**
```go
// Zitadel OIDC provider
provider, err := oidc.NewProvider(context.Background(),
    "https://your-custom-provider.com")

// GitHub provider
githubProvider := oidc.NewProvider(context.Background(),
    "https://github.com")

// PKCE-enabled flow
func (h *Handler) handleCallback(c *fiber.Ctx) error {
    code := c.Query("code")
    codeVerifier := getPKCEVerifier(c) // From session

    // Exchange code for tokens with PKCE verification
    tokens, err := provider.Exchange(c.Context(),
        code, codeVerifier, redirectURI)

    // Validate ID token
    idToken, err := oidc.VerifyIDToken(c.Context(),
        tokens.IDToken, provider)

    return h.createSession(c, idToken)
}
```

#### Option 3: CoreOS go-oidc + Custom OAuth2 Server
**If you want full control over the OAuth2 server**

**Dependencies:**
```go
require (
    github.com/coreos/go-oidc/v3 v3.14.1
    github.com/go-oauth2/oauth2/v4 v4.5.2
)
```

**Pros:**
- ✅ Maximum control over implementation
- ✅ Can build custom OAuth2 server
- ✅ GitHub integration straightforward
- ✅ Good for learning OAuth2 internals

### Machine-to-Machine (M2M) Implementation

For service clients, use **OAuth 2.0 Client Credentials Flow**:

```go
// Client Credentials flow for M2M
func (h *Handler) handleM2MToken(c *fiber.Ctx) error {
    clientID := c.Get("X-Client-ID")
    clientSecret := c.Get("X-Client-Secret")

    // Validate client credentials
    if !h.validateClient(clientID, clientSecret) {
        return c.Status(401).JSON(fiber.Map{"error": "invalid_client"})
    }

    // Issue access token
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
        "iss": "cryptoutil",
        "sub": clientID,
        "aud": "service-api",
        "scope": "service:api",
        "exp": time.Now().Add(time.Hour).Unix(),
    })

    signedToken, err := token.SignedString(h.privateKey)
    return c.JSON(fiber.Map{"access_token": signedToken})
}
```

### Fiber Integration Pattern

```go
// Middleware for route protection
func (h *Handler) authMiddleware(requiredScope string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        auth := c.Get("Authorization")
        if auth == "" {
            return c.Status(401).JSON(fiber.Map{"error": "missing_token"})
        }

        token := strings.TrimPrefix(auth, "Bearer ")

        // Verify JWT token
        claims, err := h.verifyToken(token)
        if err != nil {
            return c.Status(401).JSON(fiber.Map{"error": "invalid_token"})
        }

        // Check scope
        if !strings.Contains(claims["scope"].(string), requiredScope) {
            return c.Status(403).JSON(fiber.Map{"error": "insufficient_scope"})
        }

        c.Locals("claims", claims)
        return c.Next()
    }
}

// Route setup
app.Get("/browser/api/v1/*", h.authMiddleware("browser:api"))
app.Get("/service/api/v1/*", h.authMiddleware("service:api"))
```

### Security Best Practices

1. **PKCE Always**: Use PKCE for all public clients (browsers)
2. **State Parameter**: Protect against CSRF attacks
3. **Nonce**: Prevent replay attacks in OIDC
4. **Secure Cookies**: HttpOnly, Secure, SameSite for session cookies
5. **Token Storage**: Never store tokens in localStorage (use secure httpOnly cookies)
6. **Refresh Tokens**: Implement secure refresh token rotation
7. **Rate Limiting**: Protect auth endpoints from abuse

### Recommended Choice for Cryptoutil Project

**Go with Option 1 (Ory Hydra)** because:
- Production-ready OAuth2/OIDC server
- Excellent GitHub integration
- Your custom provider can leverage Hydra's battle-tested implementation
- Fits well with your existing Fiber + OpenTelemetry stack
- Active community and commercial support available

### Alternative OAuth2/OIDC Servers

**Production-Ready Options:**
- **Keycloak**: Java-based, feature-rich, enterprise-grade
- **Auth0**: SaaS solution, easy integration
- **Dex**: CNCF project, Go-based, Kubernetes-native
- **Fosite**: Ory's OAuth2 framework for custom implementations

**GitHub OAuth2 Integration:**
```go
// Direct GitHub OAuth2 (simplest)
githubConfig := &oauth2.Config{
    ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
    ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
    Scopes:       []string{"user:email", "read:user"},
    Endpoint:     github.Endpoint,
}
```

### Implementation Roadmap

1. **Phase 1**: Set up Ory Hydra in Docker Compose
2. **Phase 2**: Implement Authorization Code flow with PKCE
3. **Phase 3**: Add GitHub OAuth2 provider
4. **Phase 4**: Implement Client Credentials flow for M2M
5. **Phase 5**: Add token validation middleware
6. **Phase 6**: Update OpenAPI specs and documentation

**Timeline**: Q4 2025 implementation as planned in Task O1
