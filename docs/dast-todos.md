# DAST TODO List - Active Tasks

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

---

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

---

### DAST Workflow Performance Optimization (🔵 LOW - Optional)Only

**Document Status**: Active Remediation Phase
**Created**: 2025-09-30
**Updated**: 2025-10-05
**Purpose**: Actionable task list for remaining DAST workflow improvements

> Maintenance Guideline: If a file/config/feature is removed or a decision makes a task permanently obsolete, DELETE its tasks and references here immediately. Keep only (1) active remediation work, (2) still-relevant observations, (3) forward-looking backlog items. Historical context belongs in commit messages or durable docs, not this actionable list.

---

## Executive Summary

**CURRENT STATUS** (2025-10-09): 🔄 **OAuth 2.0 Implementation Planning Underway**

- 🔄 **OAuth 2.0 implementation planning** - Separate flows for users (browser APIs) vs machines (service APIs)
- 🟡 **Security hardening in progress** - Cookie HttpOnly flags and JSON parsing issues remain
- 🔵 **Performance optimization pending** - Workflow parallelization opportunity identified

**Next Priority**: Implement OAuth 2.0 Authorization Code flows for secure API access separation

---

## Active Tasks

### DAST Workflow Performance Optimization (� LOW - Optional)

#### Task O2: Implement Parallel Step Execution (� LOW)
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

## Recent Completions (2025-10-05)

### ZAP Connectivity Analysis ✅
- **Issue**: ZAP scan failing in act workflow
- **Root Cause**: NOT networking - ZAP successfully connected and scanned 14 URLs
- **Actual Problem**: File permission error on Windows/WSL2 when writing reports
- **Solution**: Added pre-scan chmod 777 step for act on Windows
- **Analysis**: See `docs/zap-analysis-2025-10-05.md` for detailed investigation

### Key Findings
- ✅ ZAP networking works correctly with `--network=host`
- ✅ ZAP successfully targets `https://127.0.0.1:8080`
- ✅ All 110+ security checks executed and passed
- ✅ Fixed Windows/WSL2 volume mount permission issues

---

## Priority Execution Order

### NEXT PRIORITY - OAuth 2.0 Implementation (Q4 2025)
1. **Task O1**: Design OAuth 2.0 Authorization Code Flow for User vs Machine Access
2. **Task O2**: Update API Documentation for OAuth 2.0
3. **Task O3**: Implement Token Scope Validation Middleware

### MEDIUM PRIORITY - Remaining Security Tasks
1. **Task S1**: Fix Cookie HttpOnly Flag Security Issue
2. **Task S5**: Investigate JSON Parsing Issues in API Endpoints

### LOW PRIORITY - Performance Optimization
1. **Task O2**: Implement Parallel Step Execution (workflow optimization)

---

## Quick Reference

### Successful Configuration
- **Nuclei flags**: `-c 24 -rl 200 -timeout 600 -stats -ept tcp,javascript`
- **ZAP network**: `--network=host` targeting `https://127.0.0.1:8080`
- **Act compatibility**: `github.actor == 'nektos/act'` detection working
- **Artifact collection**: Local artifacts saved to `./dast-reports/`
- **Permission fix**: `chmod 777 ./dast-reports` before ZAP runs (act only)

### Testing Commands
```powershell
# Test ZAP fix with quick scan
.\scripts\run-act-dast.ps1 -ScanProfile quick -Timeout 600

# Verify report generation
ls .\dast-reports\*.html, .\dast-reports\*.json, .\dast-reports\*.md
```

---

**Last Updated**: 2025-10-09
**Recent completions**: Completed tasks removed from active list (2025-10-09)
**Status**: OAuth 2.0 implementation planning underway. Security hardening tasks remain active.
