# Cryptoutil Project TODOs - Merged and Reorganized

**Last Updated**: October 14, 2025
**Status**: All active tasks from todos.md and missing.md merged and reorganized

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

## 🟡 MEDIUM - Testing Infrastructure & Quality Assurance

### Task T1: Add Fuzz Testing for Security-Critical Code
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
- **Dependencies**: Separate invocations required

### Task T2: Add Benchmark Testing Infrastructure
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
- **Dependencies**: Separate invocations required

### Task T3: Optimize Test Parallelization Strategy
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
- **Dependencies**: Task T1/T2 completion

### Task T4: Performance/Load Testing Framework
- **Description**: Implement comprehensive performance and load testing capabilities
- **Current State**: Basic unit tests only, no performance testing
- **Action Items**:
  - Set up performance testing framework (k6, Artillery, or custom)
  - Create load test scenarios for key API endpoints
  - Implement performance regression detection
  - Add performance testing to CI pipeline
- **Files**: Performance test scripts, CI workflow updates
- **Expected Outcome**: Automated performance validation and regression detection
- **Priority**: Medium - Production readiness

### Task T5: Integration Test Automation
- **Description**: Implement automated integration testing in CI pipeline
- **Current State**: Unit tests only, no integration testing
- **Action Items**:
  - Create integration test suite with database and external dependencies
  - Set up test database instances for CI
  - Implement API contract testing
  - Add integration tests to CI pipeline
- **Files**: Integration test files, CI workflow updates
- **Expected Outcome**: End-to-end testing validation
- **Priority**: Medium - Production readiness

---

## 🟡 MEDIUM - Code Quality & Linting Enhancements

### Task CQ1: Address TODO Comments in Codebase
- **Description**: Multiple TODO/FIXME comments found throughout codebase requiring attention
- **godox Issues**:
  - `internal/client/client_oam_mapper.go:89` - "TODO nil allowed if import not nil"
  - `internal/client/client_test.go:334` - "TODO validate public key does not contain any private key or secret key material"
  - `internal/client/client_test_util.go:22` - "TODO Add error checking for https with rootCAsPool=nil"
  - `internal/common/crypto/jose/jwkgen_service.go:46` - "TODO read from settings"
  - `internal/common/crypto/jose/jws_message_util.go:148` - "TODO support multiple signatures"
  - `internal/pool/pool.go:43` - "TODO change generateCounter and getCounter from uint64 to telemetryService.MetricsProvider.Counter()"
  - `internal/server/application/application_listener.go` - Multiple TODOs:
    - Line 54: "TODO Add separate timeouts for different shutdown phases (drain, force close, etc.)"
    - Line 93: "TODO Only use InsecureSkipVerify for DevMode"
    - Line 195: "TODO Replace this with improved otelFiberTelemetryMiddleware; unstructured logs and no OpenTelemetry are undesirable"
    - Line 200: "TODO Limit this to Swagger GET APIs, not Swagger UI static content"
    - Line 239: "TODO Disable Swagger UI in production environments (check settings.DevMode or add settings.Environment)"
    - Line 240: "TODO Add authentication middleware for Swagger UI access"
    - Line 241: "TODO Add specific rate limiting for Swagger UI endpoints"
- **Action Items**:
  - Review each TODO comment for relevance and priority
  - Implement high-priority TODOs or convert to proper issues
  - Remove obsolete TODOs
  - Add proper documentation for complex TODOs
- **Files**: Multiple files across codebase
- **Expected Outcome**: Clean codebase with actionable TODOs only
- **Priority**: LOW - Code maintainability improvement

### Task CQ2: Enhanced Magic Number Detection
- **Description**: Improve detection of repeated strings and numeric constants
- **Current State**: Basic goconst configuration
- **Action Items**:
  - Update goconst settings (min-len: 2, min-occurrences: 2, numbers: true)
  - Enable gocritic linter with magic number detection
  - Add gomnd (Go Magic Number Detector) linter
  - Configure appropriate thresholds for hugeParam and rangeValCopy
- **Files**: `.golangci.yml`
- **Expected Outcome**: Better detection of magic strings and numeric literals
- **Priority**: Medium - Code quality improvement

### Task CQ3: Enable Additional Quality Linters
- **Description**: Add more golangci-lint linters for enhanced code quality
- **Current State**: Core linters enabled
- **Action Items**:
  - Evaluate and enable: exportloopref, gocognit, goheader, gomoddirectives, gomodguard, importas, lll, nlreturn, testpackage, wsl
  - Configure appropriate settings for each linter
  - Test CI performance impact
- **Files**: `.golangci.yml`
- **Expected Outcome**: Enhanced code quality and consistency checks
- **Priority**: Medium - Code quality improvement

---

## 🟡 MEDIUM - Observability & Monitoring

### Task OB1: Expand Grafana Dashboards for Custom Metrics
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

### Task OB2: Implement Prometheus Metrics Exposition
- **Description**: Add comprehensive Prometheus metrics for monitoring and alerting
- **Current State**: Basic HTTP metrics only
- **Action Items**:
  - Implement custom Prometheus metrics for application performance
  - Add business logic metrics (crypto operations, key generation)
  - Configure metrics endpoints and scraping
  - Set up alerting rules and SLI/SLO definitions
- **Files**: Metrics implementation, Prometheus configuration
- **Expected Outcome**: Production-grade monitoring capabilities
- **Priority**: Medium - Production readiness

### Task OB3: Distributed Tracing Examples
- **Description**: Implement distributed tracing examples and documentation
- **Current State**: Basic tracing infrastructure exists
- **Action Items**:
  - Create tracing examples for key operations
  - Document tracing setup and best practices
  - Add trace correlation across services
- **Files**: Tracing examples, documentation
- **Expected Outcome**: Enhanced debugging and performance analysis
- **Priority**: Medium - Observability enhancement

---

## 🟡 MEDIUM - Infrastructure & Deployment

### Task INF1: Automated Release Pipeline
- **Description**: Implement automated release pipeline with semantic versioning
- **Current State**: Manual releases only
- **Action Items**:
  - Create `.github/workflows/release.yml` with automated changelog generation
  - Implement semantic versioning and automated releases
  - Set up container registry publishing
  - Configure multi-environment deployment strategy (dev → staging → production)
- **Files**: `.github/workflows/release.yml`, release scripts
- **Expected Outcome**: Automated, reliable release process
- **Priority**: High - Production deployment

### Task INF2: Kubernetes Deployment Manifests
- **Description**: Create production-ready Kubernetes deployment configurations
- **Current State**: Docker Compose only
- **Action Items**:
  - Create Kubernetes deployment, service, and ingress manifests
  - Implement ConfigMaps and Secrets management
  - Set up health checks and readiness probes
  - Configure resource limits and requests
- **Files**: `deployments/kubernetes/` directory with YAML manifests
- **Expected Outcome**: Production Kubernetes deployment capability
- **Priority**: Medium - Production infrastructure

### Task INF3: Helm Charts for Flexible Deployment
- **Description**: Create Helm charts for flexible, templated deployments
- **Current State**: No Helm support
- **Action Items**:
  - Create Helm chart with configurable values
  - Implement chart templating for different environments
  - Add chart testing and validation
  - Document Helm deployment procedures
- **Files**: `deployments/helm/cryptoutil/` directory
- **Expected Outcome**: Flexible deployment across environments
- **Priority**: Medium - Production infrastructure

### Task INF4: Pin Docker Image Versions
- **Description**: Pin all Docker image versions in compose configuration for reproducible builds
- **Current State**: Some images may use latest tags
- **Action Items**:
  - Audit all Docker images in compose files
  - Pin versions to specific tags (avoid :latest)
  - Set up automated dependency updates for security patches
- **Files**: `deployments/compose/*.yml`
- **Expected Outcome**: Reproducible and secure container deployments
- **Priority**: Medium - Infrastructure stability

---

## 🟢 LOW - Development Workflow & Configuration

### Task DW1: Implement 12-Factor App Standards Compliance
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

### Task DW2: Implement Hot Config File Reload
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

### Task DW3: Pre-commit Hook Enhancements
- **Description**: Enhance pre-commit hooks with additional validation checks
- **Current State**: Basic pre-commit setup
- **Action Items**:
  - Add script shebang validation (`check-executables-have-shebangs`)
  - Add script executable permissions (`check-shebang-scripts-are-executable`)
  - Enable shell script linting (`shellcheck` for `.sh` files)
  - Enable PowerShell script analysis (`PSScriptAnalyzer` for `.ps1` files)
  - Add private key detection (`detect-private-key` with `.pem` exclusions)
- **Files**: `.pre-commit-config.yaml`
- **Expected Outcome**: Enhanced development workflow security and quality
- **Priority**: Low - Development tooling improvement

### Task DW4: Implement Parallel Step Execution
- **Description**: Parallelize setup steps that don't depend on each other
- **Context**: Currently all setup steps run sequentially, but some can run in parallel
- **Action Items**:
  - Run directory creation in background (`mkdir -p configs/test & mkdir -p ./dast-reports &`)
  - Parallelize config file creation with other setup tasks
  - Optimize application startup sequence
- **Files**: `.github/workflows/dast.yml` (Start application step)
- **Expected Savings**: ~10-15 seconds per run (minor optimization)
- **Priority**: Low - workflow already runs efficiently with scan profiles

---

## 🟢 LOW - Documentation & API Management

### Task DOC1: API Versioning Strategy Documentation
- **Description**: Document comprehensive API versioning strategy and deprecation policy
- **Current State**: Basic API versioning exists but not formally documented
- **Action Items**:
  - Document API versioning conventions (URL-based, header-based, etc.)
  - Create API deprecation policy and timeline
  - Document backward compatibility guarantees
  - Create migration guides for API changes
- **Files**: `docs/api-versioning.md`, OpenAPI specifications
- **Expected Outcome**: Clear API evolution and compatibility guidelines
- **Priority**: Low - API management

### Task DOC2: Performance Benchmarks Documentation
- **Description**: Create comprehensive performance benchmarks and documentation
- **Current State**: Performance testing exists but not documented
- **Action Items**:
  - Document performance benchmarks for key operations
  - Create performance comparison charts and metrics
  - Document performance testing methodology
  - Add performance expectations to API documentation
- **Files**: `docs/performance-benchmarks.md`, benchmark results
- **Expected Outcome**: Performance transparency and expectations
- **Priority**: Low - Documentation enhancement

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
})// Hydra client setup
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

**Timeline**: Q4 2025 implementation as planned in Task O1.

---

## Priority Execution Order

### NEXT PRIORITY - OAuth 2.0 Implementation (Q4 2025)
1. **Task O1**: Design OAuth 2.0 Authorization Code Flow for User vs Machine Access
2. **Task O2**: Update API Documentation for OAuth 2.0
3. **Task O3**: Implement Token Scope Validation Middleware

### MEDIUM PRIORITY - Security & Testing (Q4 2025)
1. **Task S1**: Fix Cookie HttpOnly Flag Security Issue
2. **Task S2**: Investigate JSON Parsing Issues in API Endpoints
3. **Task S3**: Fix golangci-lint staticcheck Integration Issue
4. **Task T1**: Add Fuzz Testing for Security-Critical Code
5. **Task T2**: Add Benchmark Testing Infrastructure
6. **Task T3**: Optimize Test Parallelization Strategy
7. **Task T4**: Performance/Load Testing Framework
8. **Task T5**: Integration Test Automation

### MEDIUM PRIORITY - Infrastructure & Observability (Q4 2025 - Q1 2026)
1. **Task INF1**: Automated Release Pipeline
2. **Task INF2**: Kubernetes Deployment Manifests
3. **Task INF3**: Helm Charts for Flexible Deployment
4. **Task INF4**: Pin Docker Image Versions
5. **Task OB1**: Expand Grafana Dashboards for Custom Metrics
6. **Task OB2**: Implement Prometheus Metrics Exposition
7. **Task OB3**: Distributed Tracing Examples

### LOW PRIORITY - Quality & Documentation (Ongoing)
1. **Task CQ1**: Address TODO Comments in Codebase
2. **Task CQ2**: Enhanced Magic Number Detection
3. **Task CQ3**: Enable Additional Quality Linters
4. **Task DW1**: Implement 12-Factor App Standards Compliance
5. **Task DW2**: Implement Hot Config File Reload
6. **Task DW3**: Pre-commit Hook Enhancements
7. **Task DW4**: Implement Parallel Step Execution
8. **Task DOC1**: API Versioning Strategy Documentation
9. **Task DOC2**: Performance Benchmarks Documentation
10. **Task S4**: Advanced Threat Modeling Documentation

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

**Last Updated**: October 14, 2025
**Status**: All content from todos.md and missing.md merged and reorganized. Ready for splitting recommendations.</content>
<parameter name="filePath">c:\Dev\Projects\cryptoutil\docs\merged-todos.md
