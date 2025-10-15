# Project TODO List - Active Tasks

## 🟡 MEDIUM - Testing Infrastructure Improvements

#### Task T2: Add Fuzz Testing for Security-Critical Code (🟡 MEDIUM)
- **Description**: Implement fuzz testing for cryptographic operations and key generation functions
- **Current State**: No fuzz tests implemented (no `func Fuzz*` functions found)
- **Action Items**:
  - Add fuzz tests for key generation functions in `pkg/keygen`
  - Add fuzz tests for cryptographic operations in `internal/crypto`
  - Configure appropriate fuzz timeouts (5-30 minutes)
  - Add fuzz testing to nightly/release CI pipeline
- **Files**: `*_test.go` files in `pkg/keygen/`, `internal/crypto/`
- **Expected Outcome**: Property-based testing for security-critical code paths
- **Priority**: Medium - Security testing enhancement
- **Dependencies**: Task T1 (separate invocations)

#### Task T3: Add Benchmark Testing Infrastructure (🟡 MEDIUM)
- **Description**: Implement performance benchmarking for key cryptographic operations
- **Current State**: No benchmark tests implemented (no `func Benchmark*` functions found)
- **Action Items**:
  - Add benchmarks for key generation performance
  - Add benchmarks for cryptographic operations
  - Configure benchmark testing in CI pipeline
  - Add performance regression detection
- **Files**: `*_test.go` files in `pkg/keygen/`, `internal/crypto/`
- **Expected Outcome**: Performance monitoring and regression detection
- **Priority**: Medium - Performance validation
- **Dependencies**: Task T1 (separate invocations)

#### Task T5: Optimize Test Parallelization Strategy (🟡 MEDIUM)
- **Description**: Fine-tune test parallelization for optimal CI performance
- **Current State**: Basic `-p=2` parallelization, but may not be optimal for all packages
- **Action Items**:
  - Analyze test execution times by package
  - Optimize `-p` flag usage (parallel packages vs parallel tests within packages)
  - Balance test load across CI runners
  - Monitor and adjust based on CI performance metrics
- **Files**: `.github/workflows/ci.yml`, test execution analysis
- **Expected Outcome**: Faster CI execution with better resource utilization
- **Priority**: Medium - CI performance optimization
- **Dependencies**: Task T1 completion

---

##  CRITICAL - OAuth 2.0 Implementation Planning

#### Task O1: Design OAuth 2.0 Authorization Code Flow for User vs Machine Access (🔴 CRITICAL)
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

#### Task O2: Update API Documentation for OAuth 2.0 (🟡 MEDIUM)
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

#### Task O3: Implement Token Scope Validation Middleware (� MEDIUM)
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

## 🟡 MEDIUM - Remaining Security Hardening

#### Task O2: Implement Parallel Step Execution (🔵 LOW)
- **Description**: Parallelize setup steps that don't depend on each other
- **Context**: Currently all setup steps run sequentially, but some can run in parallel
- **Action Items**:
  - Run directory creation in background (`mkdir -p configs/test & mkdir -p ./dast-reports &`)
  - Parallelize config file creation with other setup tasks
  - Optimize application startup sequence
- **Files**: `.github/workflows/dast.yml` (Start application step)
- **Expected Savings**: ~10-15 seconds per run (minor optimization)
- **Priority**: Low - workflow already runs efficiently with scan profiles

## Security Findings Remediation (🟡 MEDIUM)

#### Task S1: Fix Cookie HttpOnly Flag Security Issue (🟡 MEDIUM)
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

#### Task S5: Investigate JSON Parsing Issues in API Endpoints (🟡 MEDIUM)
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

#### Task S6: Fix golangci-lint staticcheck Integration Issue (🟡 MEDIUM)
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

---

### Project Workflow Performance Optimization (🔵 LOW - Optional)

**Document Status**: Active Remediation Phase
**Created**: 2025-09-30
**Updated**: 2025-10-12
**Purpose**: Actionable task list for remaining project workflow improvements and code quality issues

> Maintenance Guideline: If a file/config/feature is removed or a decision makes a task permanently obsolete, DELETE its tasks and references here immediately. Keep only (1) active remediation work, (2) still-relevant observations, (3) forward-looking backlog items. Historical context belongs in commit messages or durable docs, not this actionable list.

## Priority Execution Order

### NEXT PRIORITY - OAuth 2.0 Implementation (Q4 2025)
1. **Task O1**: Design OAuth 2.0 Authorization Code Flow for User vs Machine Access
2. **Task O2**: Update API Documentation for OAuth 2.0
3. **Task O3**: Implement Token Scope Validation Middleware

### MEDIUM PRIORITY - Remaining Security Tasks
1. **Task S1**: Fix Cookie HttpOnly Flag Security Issue
2. **Task S5**: Investigate JSON Parsing Issues in API Endpoints
3. **Task S6**: Fix golangci-lint staticcheck Integration Issue
4. **Task C6**: Pin Docker Image Versions in Compose Config

### LOW PRIORITY - Performance Optimization
1. **Task O2**: Implement Parallel Step Execution (workflow optimization)

---

## Code Quality TODOs (godox lint warnings)

#### Task CQ1: Address TODO comments in codebase (🔵 LOW)
- **Description**: Multiple TODO/FIXME comments found throughout codebase requiring attention
- **godox Issues**:
  - `internal/client/client_oam_mapper.go:89` - "TODO nil allowed if import not nil"
  - `internal/client/client_test.go:334` - "TODO validate public key does not contain any private key or secret key material"
  - `internal/client/client_test_util.go:22` - "TODO Add error checking for https with rootCAsPool=nil"
  - `internal/common/crypto/jose/jwkgen_service.go:46` - "TODO read from settings"
  - `internal/common/crypto/jose/jws_message_util.go:148` - "TODO support multiple signatures"
  - `internal/common/pool/pool.go:43` - "TODO change generateCounter and getCounter from uint64 to telemetryService.MetricsProvider.Counter()"
  - `internal/server/application/application_listener.go` - Multiple TODOs:
    - Line 54: "TODO Add separate timeouts for different shutdown phases (drain, force close, etc.)"
    - Line 93: "TODO Only use InsecureSkipVerify for DevMode"
    - Line 195: "TODO Replace this with improved otelFiberTelemetryMiddleware; unstructured logs and no OpenTelemetry are undesirable"
    - Line 200: "TODO Limit this to Swagger GET APIs, not Swagger UI static content"
    - Line 239: "TODO Disable Swagger UI in production environments (check settings.DevMode or add settings.Environment)"
    - Line 240: "TODO Add authentication middleware for Swagger UI access"
    - Line 241: "TODO Add specific rate limiting for Swagger UI endpoints"
  - `internal/server/businesslogic/businesslogic.go:250` - "TODO cache GetElasticKey"
  - `internal/server/businesslogic/businesslogic.go:328` - "TODO Use encryptParams.Context for encryption"
- **Action Items**:
  - Review each TODO comment for relevance and priority
  - Implement high-priority TODOs or convert to proper issues
  - Remove obsolete TODOs
  - Add proper documentation for complex TODOs
- **Files**: Multiple files across codebase
- **Expected Outcome**: Clean codebase with actionable TODOs only
- **Priority**: LOW - Code maintainability improvement
- **Note**: godox linter disabled in favor of manual tracking in this file

---

## Quick Reference

### Testing Commands
```powershell
# Test ZAP fix with quick scan
.\scripts\run-act-dast.ps1 -ScanProfile quick -Timeout 600

# Verify report generation
ls .\dast-reports\*.html, .\dast-reports\*.json, .\dast-reports\*.md
```

---

**Last Updated**: 2025-10-14
**Status**: OAuth 2.0 implementation planning underway. Security hardening tasks remain active. Staticcheck integration issue and Docker image version pinning task added.

---

## Configuration & UX Improvements (🔵 LOW)

#### Task C1: Implement 12-Factor App Standards Compliance (🔵 LOW)
- **Description**: Ensure application follows 12-factor app methodology for cloud-native deployment
- **12-Factor Requirements**:
  - **I. Codebase**: One codebase tracked in revision control, many deploys
  - **II. Dependencies**: Explicitly declare and isolate dependencies
  - **III. Config**: Store config in the environment (✅ Environment variables implemented)
  - **IV. Backing services**: Treat backing services as attached resources
  - **V. Build, release, run**: Strictly separate build and run stages
  - **VI. Processes**: Execute the app as one or more stateless processes
  - **VII. Port binding**: Export services via port binding (✅ Implemented)
  - **VIII. Concurrency**: Scale out via the process model
  - **IX. Disposability**: Maximize robustness with fast startup and graceful shutdown
  - **X. Dev/prod parity**: Keep development, staging, and production as similar as possible
  - **XI. Logs**: Treat logs as event streams (✅ OTLP logging implemented)
  - **XII. Admin processes**: Run admin/management tasks as one-off processes
- **Current State**: Environment variables and port binding implemented, others need review
- **Action Items**:
  - Audit codebase for 12-factor compliance gaps
  - Implement missing factors (config separation, stateless processes, etc.)
  - Update deployment configurations for 12-factor compliance
  - Document 12-factor compliance status
- **Files**: Docker configs, deployment files, application architecture
- **Expected Outcome**: Cloud-native, scalable application following industry best practices
- **Priority**: LOW - Best practices alignment
- **Timeline**: Ongoing maintenance

#### Task C2: Implement Hot Config File Reload (🔵 LOW)
- **Description**: Add ability to reload configuration files without restarting the server
- **Current State**: Configuration loaded only at startup
- **Action Items**:
  - Add file watcher for config files (development mode only)
  - Implement graceful config reload with validation
  - Add reload endpoint for runtime config updates
  - Handle config reload failures gracefully
  - Add configuration versioning/checksum validation
- **Files**: `internal/common/config/config.go`, server startup code
- **Expected Outcome**: Development workflow improvement with live config reloading
- **Priority**: LOW - Developer experience enhancement
- **Timeline**: Q1 2026

#### Task D1: Expand Grafana Dashboards for Custom Metrics (🟡 MEDIUM)
- **Description**: Current Grafana dashboard only covers basic HTTP metrics but misses all custom application metrics
- **Current State**: Dashboard shows only `http_requests_total` and `http_request_duration_seconds_bucket` from otelfiber middleware
- **Missing Metrics Categories**:
  - **Pool Performance Metrics**: `cryptoutil.pool.get`, `cryptoutil.pool.permission`, `cryptoutil.pool.generate` histograms
  - **Security Header Metrics**: `security_headers_missing_total` counter
  - **Business Logic Metrics**: None currently implemented but infrastructure ready
- **Action Items**:
  - Create comprehensive dashboard panels for pool performance monitoring
  - Add security metrics dashboard with header compliance tracking
  - Implement business logic metrics for cryptographic operations
  - Update dashboard JSON with proper Prometheus queries for OpenTelemetry metrics
  - Add alerting rules for security header violations and pool performance issues
- **Files**: `deployments/compose/grafana-otel-lgtm/dashboards/cryptoutil.json`
- **Expected Outcome**: Full observability of all custom application metrics
- **Priority**: MEDIUM - Observability improvement
- **Timeline**: Q4 2025

---

## Appendix: Grafana Dashboard Expansion

### Current Dashboard Limitations

The existing `cryptoutil.json` dashboard only includes basic HTTP metrics from the `otelfiber` middleware:

```json
{
  "panels": [
    {
      "title": "Request Rate",
      "targets": [{"expr": "rate(http_requests_total[5m])"}]
    },
    {
      "title": "Response Time",
      "targets": [{"expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"}]
    }
  ]
}
```

### Missing Custom Metrics Categories

#### 1. Pool Performance Metrics
**Source**: `internal/common/pool/pool.go`

**Available Metrics**:
- `cryptoutil.pool.get` - Histogram of get operation duration (milliseconds)
- `cryptoutil.pool.permission` - Histogram of permission wait duration (milliseconds)
- `cryptoutil.pool.generate` - Histogram of generate operation duration (milliseconds)

**Attributes**: `workers`, `size`, `values`, `duration`, `type` (per pool instance)

**Recommended Dashboard Panels**:
```json
{
  "title": "Pool Get Latency (95th percentile)",
  "targets": [{
    "expr": "histogram_quantile(0.95, rate(cryptoutil_pool_get_bucket[5m]))",
    "legendFormat": "{{pool}} - {{type}}"
  }]
},
{
  "title": "Pool Permission Wait Time",
  "targets": [{
    "expr": "histogram_quantile(0.95, rate(cryptoutil_pool_permission_bucket[5m]))",
    "legendFormat": "{{pool}} permission wait"
  }]
},
{
  "title": "Pool Generation Time",
  "targets": [{
    "expr": "histogram_quantile(0.95, rate(cryptoutil_pool_generate_bucket[5m]))",
    "legendFormat": "{{pool}} generation time"
  }]
}
```

#### 2. Security Header Validation Metrics
**Source**: `internal/server/application/application_listener.go`

**Available Metrics**:
- `security_headers_missing_total` - Counter of requests with missing security headers

**Recommended Dashboard Panels**:
```json
{
  "title": "Security Header Violations",
  "targets": [{
    "expr": "rate(security_headers_missing_total[5m])",
    "legendFormat": "Missing headers per second"
  }]
},
{
  "title": "Security Header Compliance Rate",
  "targets": [{
    "expr": "(1 - (rate(security_headers_missing_total[5m]) / rate(http_requests_total[5m]))) * 100",
    "legendFormat": "Compliance %"
  }]
}
```

### Implementation Architecture

**Metrics Flow**:
```
Application (OpenTelemetry) → OTEL Collector → Grafana-OTEL-LGTM (Prometheus + Grafana)
```

**Dashboard Updates Needed**:
1. **Add Pool Performance Dashboard**:
   - Pool utilization metrics
   - Latency percentiles per pool type
   - Worker efficiency monitoring

2. **Add Security Dashboard**:
   - Header compliance rates
   - Violation trends
   - Alert thresholds for security issues

3. **Add Business Logic Dashboard** (Future):
   - Cryptographic operation metrics
   - Key generation performance
   - Database operation latency

### OpenTelemetry to Prometheus Metric Name Mapping

OpenTelemetry metrics are automatically converted to Prometheus format:
- `cryptoutil.pool.get` → `cryptoutil_pool_get`
- `security_headers_missing_total` → `security_headers_missing_total`

### Alerting Recommendations

```yaml
# Example alert rules for security headers
groups:
  - name: security
    rules:
      - alert: HighSecurityHeaderViolations
        expr: rate(security_headers_missing_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High rate of missing security headers"
```

### Priority Implementation Order

1. **Phase 1**: Add pool performance metrics dashboard
2. **Phase 2**: Add security header compliance dashboard
3. **Phase 3**: Implement business logic metrics and dashboard
4. **Phase 4**: Add alerting rules and thresholds

**Timeline**: Q4 2025 implementation alongside OAuth 2.0 work.

---

**Last Updated**: 2025-10-14
**Status**: OAuth 2.0 implementation planning underway. Security hardening tasks remain active. Staticcheck integration issue and Docker image version pinning task added.

---

## Appendix: OAuth 2.0 & OIDC Implementation Options

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
- Active community and commercial support available
- Fits well with your existing Fiber + OpenTelemetry stack

### Alternative OAuth2/OIDC Servers

**Production-Ready Options:**
- **Keycloak**: Java-based, feature-rich, enterprise-grade
- **Auth0**: SaaS solution, easy integration
- **Dex**: CNCF project, Go-based, Kubernetes-native
- **Fosite**: Ory's OAuth2 framework for custom implementations

**GitHub OAuth2 Integration:```go
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

**Timeline**: Q4 2025 implementation as planned in Task O1.

---

## 🟡 MEDIUM - Magic Number Detection Enhancement

#### Task MN1: Enhanced goconst Configuration (🟡 MEDIUM)
- **Description**: Update goconst linter settings for more aggressive detection of repeated strings and numeric constants
- **Current State**: Basic goconst configuration with min-len: 3, min-occurrences: 3
- **Action Items**:
  - Lower min-len to 2 for shorter repeated strings
  - Lower min-occurrences to 2 for more aggressive detection
  - Enable numbers: true for numeric constant detection
  - Enable match-constant: true to also match existing constants
  - **Note**: goconst is available both as standalone tool (`goconst`) and built into golangci-lint
- **Files**: `.golangci.yml`
- **Expected Outcome**: Better detection of magic strings and numeric literals
- **Priority**: Medium - Code quality improvement

#### Task MN2: Enable gocritic for Magic Number Detection (🟡 MEDIUM)
- **Description**: Add gocritic linter with magic number detection capabilities
- **Current State**: gocritic not enabled in golangci-lint configuration
- **Action Items**:
  - Enable gocritic linter in `.golangci.yml`
  - Configure magic number detection rules
  - Set appropriate thresholds for hugeParam and rangeValCopy
  - Test and tune disabled-checks for project compatibility
  - **Note**: gocritic is available both as standalone tool (`gocritic`) and built into golangci-lint
- **Files**: `.golangci.yml`
- **Expected Outcome**: Detection of numeric literals that should be named constants
- **Priority**: Medium - Code quality improvement

#### Task MN3: Enable gomnd for Additional Magic Number Detection (🟡 MEDIUM)
- **Description**: Add gomnd (Go Magic Number Detector) linter as complement to gocritic for magic number detection
- **Current State**: gomnd not enabled in golangci-lint configuration
- **Action Items**:
  - Enable gomnd linter in `.golangci.yml`
  - Configure appropriate settings for magic number detection
  - Test for overlap/redundancy with gocritic magic number rules
  - **Note**: gomnd is available both as standalone tool (`gomnd`) and built into golangci-lint
- **Files**: `.golangci.yml`
- **Expected Outcome**: Additional detection of numeric literals that should be named constants
- **Priority**: Medium - Code quality improvement
- **Dependencies**: Task MN2 (gocritic) completion

**Common Magic Values to Watch For:**
- HTTP status codes: `200`, `404`, `500`
- Timeouts: `30`, `60`, `300` (seconds)
- Buffer sizes: `1024`, `4096`
- Retry counts: `3`, `5`, `10`
- Port numbers: `8080`, `5432`
- String literals: `"localhost"`, `"admin"`, `"default"`

---

## 🟡 MEDIUM - Additional Linters Enhancement

#### Task AL1: Evaluate and Enable Additional Linters (🟡 MEDIUM)
- **Description**: Assess and enable additional golangci-lint linters for improved code quality
- **Current State**: Core linters enabled, additional quality linters available
- **Action Items**:
  - Evaluate these additional linters for project compatibility:
    - **Built into golangci-lint only**: None in this list
    - **Available as separate applications AND built into golangci-lint**:
      - `exportloopref` - Detects exported loop variable references (standalone: `exportloopref`)
      - `gocognit` - Cognitive complexity analysis (standalone: `gocognit`)
      - `goheader` - License header checking (standalone: `goheader`)
      - `gomoddirectives` - Go module directive checking (standalone: `gomoddirectives`)
      - `gomodguard` - Block specific modules (standalone: `gomodguard`)
      - `importas` - Enforce import aliasing (standalone: `importas`)
      - `lll` - Long line detection (standalone: `lll`)
      - `nlreturn` - Require newlines before return (standalone: `nlreturn`)
      - `testpackage` - Test package naming conventions (standalone: `testpackage`)
      - `wsl` - Whitespace linting (standalone: `wsl`)
    - **Only available as separate applications**: `ineffassign`, `misspell`, `errcheck` (already enabled)
  - Enable compatible linters in `.golangci.yml`
  - Configure appropriate settings for each enabled linter
  - Test CI performance impact and adjust as needed
- **Files**: `.golangci.yml`
- **Expected Outcome**: Enhanced code quality and consistency checks
- **Priority**: Medium - Code quality improvement

---
