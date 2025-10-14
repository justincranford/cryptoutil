# Project TODO List - Active Tasks

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
    - Line 565: "TODO Enhance health checks with detailed status (database, dependencies, memory usage)" - **ALREADY IMPLEMENTED**
    - Line 566: "TODO Implement separate LivenessProbe vs ReadinessProbe functions for Kubernetes deployments" - **MIDDLEWARE IS SUFFICIENT**
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
