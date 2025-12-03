# Cryptoutil Security & Authentication TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: December 3, 2025
**Status**: Most OAuth tasks completed with Identity V2 implementation. Remaining items are hardening and compliance.

---

## ✅ COMPLETED - OAuth 2.1 Implementation (Identity V2)

The following tasks are now COMPLETED as part of Identity V2:

- **Task O1**: OAuth 2.0 Authorization Code Flow ✅ (implemented in `internal/identity/authz/`)
- **Task O2**: OAuth 2.0 Client Credentials Flow ✅ (implemented in `internal/identity/authz/`)
- **Task O3**: Token Scope Validation Middleware ✅ (implemented in `internal/kms/server/middleware/`)

**Reference**: See `specs/001-cryptoutil/EXECUTIVE-SUMMARY.md` for Identity V2 completion status.

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

### Task S5: Add Java Static Analysis to CI/CD Workflow

- **Description**: Re-add Java static analysis to a dedicated workflow for load testing code
- **Current State**: Java was removed from ci-sast.yml matrix to focus on Go and JavaScript analysis
- **Action Items**:
  - Create dedicated workflow for Java static analysis (SpotBugs, PMD, or similar)
  - Include Java load testing code in analysis
  - Configure appropriate security rules for Java code
  - Add SARIF upload to GitHub Security tab
- **Files**: New workflow file (ci-java-sast.yml or similar), Java analysis configuration
- **Expected Outcome**: Java code has automated static security analysis
- **Priority**: Low - Future enhancement
- **Timeline**: Future implementation

---

## Archived Documentation

The OAuth 2.0 implementation options appendix has been archived as Identity V2 is now complete with a custom OAuth 2.1 implementation. See:

- `internal/identity/authz/` - Authorization server implementation
- `internal/identity/idp/` - Identity provider endpoints
- `docs/sprints/2025-12-03-phase2-3-completion.md` - Implementation summary
- `archive/todos-security-ARCHIVED.md` - Original file with OAuth 2.0 planning details
