# RS Public Server Implementation Plan

## Overview

Complete the Resource Server (RS) implementation by adding the missing public HTTP server, following the established dual-server pattern used by Authz and IdP services.

**Status**: ❌ BLOCKING - RS service cannot start without public server implementation
**Priority**: P0 (blocks E2E/Load/DAST workflows)
**Estimated Effort**: 1-2 days
**Target Completion**: 2025-12-23

---

## Current State Analysis

### Existing Files (RS Service)

- ✅ `internal/identity/rs/server/admin.go` (102 lines) - Admin server implementation
- ✅ `internal/identity/rs/server/admin_test.go` (test file)
- ✅ `internal/identity/rs/server/application.go` (102 lines) - Application layer (admin-only)
- ❌ `internal/identity/rs/server/public_server.go` - **MISSING** (blocks service startup)

### Reference Implementations (Copy Pattern From)

- ✅ `internal/identity/authz/server/public_server.go` (165 lines) - OAuth 2.1 public endpoints
- ✅ `internal/identity/idp/server/public_server.go` (165 lines) - OIDC public endpoints

---

## Implementation Tasks

### Task 1: Create RS Public Server File

**File**: `internal/identity/rs/server/public_server.go`

**Pattern**: Copy from `authz/server/public_server.go`, modify for RS-specific endpoints

**Required Functions**:

1. `NewPublicServer(ctx, config) (*PublicServer, error)` - Server initialization
2. `PublicServer.Start(ctx) error` - Start HTTPS server on port 8182
3. `PublicServer.Shutdown(ctx) error` - Graceful shutdown
4. `PublicServer.ActualPort() (int, error)` - Return bound port
5. `setupRoutes(app *fiber.App)` - Configure Fiber routes

**Endpoints** (from RS OpenAPI spec):

- `GET /protected/resource` - Protected resource access (requires valid access token)
- `POST /protected/data` - Create protected data
- `GET /protected/data/:id` - Retrieve specific protected data
- `DELETE /protected/data/:id` - Delete protected data
- `GET /health` - Health check (public, no auth)
- Swagger UI endpoints: `/browser/swagger/*`, `/service/swagger/*`

**Dependencies**:

- Fiber HTTP framework
- TLS configuration (from config)
- Token validation middleware (OAuth 2.1 access token)
- CORS middleware (browser-facing endpoints)

**Security Middleware**:

- TLS client certificate validation (optional, configurable)
- OAuth 2.1 access token validation (required for /protected/*)
- Rate limiting per IP
- Request logging and telemetry

**Estimated Lines**: ~165 (match authz/idp pattern)

---

### Task 2: Update Application Layer

**File**: `internal/identity/rs/server/application.go`

**Changes Required**:

1. Add `publicServer *PublicServer` field to `Application` struct
2. Update `NewApplication()` to create both public and admin servers:

   ```go
   // Create public server
   publicServer, err := NewPublicServer(ctx, config)
   if err != nil {
       return nil, fmt.Errorf("failed to create public server: %w", err)
   }
   app.publicServer = publicServer
   ```

3. Update `Start()` method to launch both servers concurrently:

   ```go
   // Start both public and admin servers in goroutines
   go func() {
       if err := a.publicServer.Start(ctx); err != nil {
           errChan <- fmt.Errorf("public server failed: %w", err)
       }
   }()

   go func() {
       if err := a.adminServer.Start(ctx); err != nil {
           errChan <- fmt.Errorf("admin server failed: %w", err)
       }
   }()
   ```

4. Update `Shutdown()` method to stop both servers:

   ```go
   // Shutdown public server
   if a.publicServer != nil {
       if err := a.publicServer.Shutdown(ctx); err != nil {
           return fmt.Errorf("failed to shutdown public server: %w", err)
       }
   }
   ```

5. Add `PublicPort()` method:

   ```go
   func (a *Application) PublicPort() (int, error) {
       if a.publicServer == nil {
           return 0, fmt.Errorf("public server not initialized")
       }
       return a.publicServer.ActualPort()
   }
   ```

**Reference**: `internal/identity/authz/server/application.go` lines 1-143

---

### Task 3: Add Unit Tests

**File**: `internal/identity/rs/server/public_server_test.go`

**Test Cases**:

1. `TestNewPublicServer_Success` - Verify successful initialization
2. `TestNewPublicServer_NilContext` - Error on nil context
3. `TestNewPublicServer_NilConfig` - Error on nil config
4. `TestPublicServer_Start_Success` - Server starts on configured port
5. `TestPublicServer_Start_DynamicPort` - Port 0 allocation works
6. `TestPublicServer_Shutdown_Success` - Graceful shutdown
7. `TestPublicServer_ActualPort_Success` - Returns bound port
8. `TestPublicServer_ActualPort_NotStarted` - Error when not started
9. `TestPublicServer_ProtectedEndpoints` - Access token validation works
10. `TestPublicServer_HealthEndpoint` - Health check accessible without auth

**Coverage Target**: ≥95%

**Reference**: `internal/identity/authz/server/public_server_test.go` (if exists)

---

### Task 4: Update Application Tests

**File**: `internal/identity/rs/server/application_test.go` (create if missing)

**Test Cases**:

1. `TestNewApplication_BothServersCreated` - Verify public + admin created
2. `TestApplication_Start_BothServersRunning` - Both servers listen
3. `TestApplication_Shutdown_BothServersStopped` - Both servers stop
4. `TestApplication_PublicPort_ReturnsCorrectPort` - Port accessor works
5. `TestApplication_AdminPort_ReturnsCorrectPort` - Admin port accessor works

**Coverage Target**: ≥95%

**Reference**: `internal/identity/authz/server/application_test.go` (if exists)

---

### Task 5: Integration Tests

**File**: `internal/identity/rs/server/integration_test.go`

**Test Cases**:

1. `TestIntegration_ProtectedResourceAccess` - Full flow: get token → access resource
2. `TestIntegration_TokenValidation` - Invalid token rejected
3. `TestIntegration_ExpiredToken` - Expired token rejected
4. `TestIntegration_InsufficientScope` - Scope-based authorization works
5. `TestIntegration_HealthCheck` - Health endpoint accessible

**Coverage Target**: Integration tests verify end-to-end flows

**Dependencies**: Authz service running (to issue tokens for testing)

---

### Task 6: Docker Compose Verification

**Files**: `deployments/compose/compose.yml`, `deployments/identity/compose.integration.yml`

**Verification**:

1. RS service should use same pattern as authz/idp:
   - Public port: 8182 (configured via environment)
   - Admin port: 9091 (same as authz/idp)
   - Health check: `wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9091/admin/v1/livez`

2. Test Docker Compose startup:

   ```bash
   docker compose -f deployments/compose/compose.yml up -d
   docker compose ps  # Verify RS healthy
   docker compose logs compose-identity-rs-e2e-1  # Check for errors
   ```

3. Test health endpoints:

   ```bash
   curl -k https://localhost:8182/health  # Public health
   curl -k https://localhost:9091/admin/v1/livez  # Admin livez
   curl -k https://localhost:9091/admin/v1/readyz  # Admin readyz
   ```

---

### Task 7: E2E Workflow Verification

**Goal**: Verify E2E/Load/DAST workflows pass after RS implementation

**Workflows to Monitor**:

- `ci-e2e.yml` - E2E integration tests
- `ci-load.yml` - Load testing with Gatling
- `ci-dast.yml` - DAST security scans

**Expected Outcomes**:

- ✅ Identity authz service healthy
- ✅ Identity idp service healthy
- ✅ Identity rs service healthy (currently failing)
- ✅ E2E tests for protected resource access pass
- ✅ Load tests for RS endpoints pass
- ✅ DAST scans complete without startup failures

**Validation Commands**:

```bash
# Push changes to trigger workflows
git push origin main

# Monitor workflow status
gh run list --branch main --limit 5

# View specific workflow logs if failures occur
gh run view <run-id> --log-failed
```

---

## Implementation Checklist

- [ ] Create `internal/identity/rs/server/public_server.go` (165 lines, copy from authz)
- [ ] Update `internal/identity/rs/server/application.go` (add publicServer initialization)
- [ ] Create `internal/identity/rs/server/public_server_test.go` (95%+ coverage)
- [ ] Update `internal/identity/rs/server/application_test.go` (test dual-server pattern)
- [ ] Create `internal/identity/rs/server/integration_test.go` (E2E flows)
- [ ] Verify Docker Compose RS service starts successfully
- [ ] Test local RS health endpoints (public + admin)
- [ ] Push changes and monitor E2E/Load/DAST workflows
- [ ] Verify all workflows pass (11/11 green)
- [ ] Update constitution.md (mark RS COMPLETE)
- [ ] Update DETAILED.md Section 2 timeline (RS implementation entry)

---

## Success Criteria

1. **Code Complete**:
   - ✅ `public_server.go` exists with NewPublicServer(), Start(), Shutdown(), ActualPort()
   - ✅ `application.go` creates both public + admin servers
   - ✅ Unit tests ≥95% coverage
   - ✅ Integration tests cover protected resource access

2. **Local Testing**:
   - ✅ RS service starts via Docker Compose
   - ✅ Health endpoints return 200 OK
   - ✅ Protected endpoints require valid access tokens
   - ✅ No container crashes or health check failures

3. **CI/CD Validation**:
   - ✅ E2E workflow passes (identity services healthy)
   - ✅ Load workflow passes (RS endpoints respond)
   - ✅ DAST workflow passes (no startup failures)
   - ✅ All 11 workflows green

4. **Documentation**:
   - ✅ Constitution.md service status updated (RS marked COMPLETE)
   - ✅ DETAILED.md timeline entry added (RS implementation session)
   - ✅ EXECUTIVE.md risks updated (RS blocker removed)

---

## Next Steps After Completion

1. **Authz/IdP Verification**:
   - Check if authz/idp services also need debugging (workflows show failures)
   - Investigate actual error messages from container logs
   - May require configuration fixes beyond just adding public servers

2. **Phase 4 Quality Work**:
   - Resume quality improvements from QUALITY-TODOs.md
   - Focus on E2E test implementations (OAuth, JOSE, CA workflows)
   - Address MFA implementation gaps (TOTP, passkey, OTP)

3. **Clarification Session**:
   - Run `/speckit.clarify` to generate updated clarify.md
   - Process CLARIFY-QUIZME.md questions
   - Refine constitution/spec based on implementation learnings

---

## References

- **Constitution.md**: Service status table (lines 47-55), architecture blocker (lines 57-82)
- **DETAILED.md**: 2025-12-21 service status verification timeline (lines 1283-1365)
- **QUALITY-TODOs.md**: Priority 1 RS blocker (lines 23-53)
- **WORKFLOW-FIXES-CONSOLIDATED.md**: Round 7 investigation (missing public servers)
- **Authz Reference**: `internal/identity/authz/server/public_server.go` (165 lines)
- **IdP Reference**: `internal/identity/idp/server/public_server.go` (165 lines)

---

## Risk Mitigation

**Risk**: Authz/IdP services also failing despite public servers existing
**Mitigation**: Investigate actual error messages via `gh run download <run-id>` or workflow logs before assuming RS-only issue

**Risk**: Configuration issues beyond missing public server
**Mitigation**: Review Docker Compose configs, TLS settings, database DSN, OTEL endpoints

**Risk**: Implementation diverges from authz/idp pattern
**Mitigation**: Copy exact structure from authz/idp, only change endpoint paths

**Risk**: Integration tests fail without live authz service
**Mitigation**: Mock token validation for unit tests, use real authz for integration tests

---

**Last Updated**: 2025-12-21
**Status**: Ready for implementation
**Owner**: Implementation team
**Reviewer**: Architecture team
