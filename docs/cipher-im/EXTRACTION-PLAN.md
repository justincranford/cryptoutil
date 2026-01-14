# Cipher-IM → Service-Template Extraction Plan

**Created**: 2026-01-14  
**Status**: CRITICAL - BLOCKING ALL SERVICES  
**Priority**: P0 - Must complete before any other service implementations

---

## Executive Summary

Cipher-IM currently contains extensive reusable infrastructure code that MUST be extracted to service-template to enable the other 8 product-services (jose-ja, pki-ca, identity-authz, identity-idp, identity-rp, identity-rs, identity-spa, sm-kms).

**Critical Issues Identified**:
1. **Migration numbering conflict**: Template 1001-1004, Cipher-IM 1005 - Cannot add template migrations after 1004
2. **Session middleware duplication**: Cipher-IM has SessionMiddleware (100+ lines) not in template
3. **JWT utilities duplication**: GenerateJWT function only in cipher-im
4. **Server infrastructure duplication**: Public server health endpoints, lifecycle methods
5. **Test infrastructure duplication**: HTTP error tests, server initialization patterns
6. **Multi-tenancy enforcement**: Cipher-IM shows "single-tenant" pattern, MUST be multi-tenant

---

## Phase 1: Migration Numbering Fix (CRITICAL BLOCKER)

### Problem
Current numbering:
- Template: 1001-1004
- Cipher-IM: 1005

**Cannot add template migrations after 1004 without conflicting with cipher-im 1005.**

### Solution: Reserved Number Ranges

**New numbering scheme**:
- **Template migrations**: 1001-1999 (999 slots reserved)
- **Service migrations**: 2001+ (999+ slots per service)

**Cipher-IM migrations**:
- Renumber 1005 → 2001 (messages, messages_recipient_jwks tables)

**Benefits**:
- Template can add migrations 1005-1999 without conflicts
- Each service can add migrations 2001+ independently
- Clear separation: 1xxx = shared infrastructure, 2xxx+ = service-specific

**Files to modify**:
1. Rename `internal/apps/cipher/im/repository/migrations/1005_init.up.sql` → `2001_init.up.sql`
2. Rename `internal/apps/cipher/im/repository/migrations/1005_init.down.sql` → `2001_init.down.sql`
3. Update `internal/apps/cipher/im/repository/migrations.go` documentation
4. Update `internal/apps/cipher/im/im_database_test.go` comments
5. Verify mergedFS combines template 1001-1999 + cipher-im 2001+

**Testing**:
- Run `go test ./internal/apps/cipher/im/repository/...` - Verify migration numbering
- Run `go test ./internal/apps/cipher/im/...` - All tests must pass
- Run E2E tests - Verify database initialized correctly with new numbering

---

## Phase 2: Session Infrastructure Extraction

### Problem
Cipher-IM has `SessionMiddleware` (108 lines) that:
- Validates Bearer tokens
- Calls SessionManagerService (browser vs service sessions)
- Stores session, user_id, tenant_id, realm_id in context
- **NOT IN SERVICE-TEMPLATE** (template only has basic JWTMiddleware)

### Analysis: SessionMiddleware vs JWTMiddleware

**Template JWTMiddleware** (internal/apps/template/service/server/realms/middleware.go):
- Simple JWT validation with HMAC-SHA256
- Extracts user_id from token
- No session management, no multi-tenancy

**Cipher-IM SessionMiddleware** (internal/apps/cipher/im/server/middleware/session_middleware.go):
- Advanced session validation using SessionManagerService
- Supports browser + service sessions
- Multi-tenant session tracking (tenant_id, realm_id)
- Production-grade pattern

**DECISION**: SessionMiddleware is the CORRECT pattern for all 9 services.

### Extraction Plan

**Move to service-template**:
1. `internal/apps/cipher/im/server/middleware/session_middleware.go` → `internal/apps/template/service/server/middleware/session_middleware.go`
2. Create `internal/apps/template/service/server/middleware/` package
3. SessionMiddleware, BrowserSessionMiddleware, ServiceSessionMiddleware functions
4. Update cipher-im to import from template

**Dependencies to verify**:
- SessionManagerService MUST be in template (check if already exists)
- fiber.Handler compatibility
- Context locals pattern (session, user_id, client_id, tenant_id, realm_id)

**Files to modify**:
1. Create `internal/apps/template/service/server/middleware/session_middleware.go`
2. Update `internal/apps/cipher/im/server/public_server.go` import
3. Update `internal/apps/cipher/im/server/middleware/session_middleware.go` → DELETE (moved to template)
4. Add tests: `internal/apps/template/service/server/middleware/session_middleware_test.go`

---

## Phase 3: JWT Utilities Extraction

### Problem
Cipher-IM has `GenerateJWT` function (internal/apps/cipher/im/server/util/jwt.go):
- Creates JWT with UserID, Username, expiration
- Uses HMAC-SHA256 signing
- **NOT IN SERVICE-TEMPLATE** (template realms/jwt.go only has Claims struct)

### Extraction Plan

**Move to service-template**:
1. `GenerateJWT` function → `internal/apps/template/service/server/realms/jwt.go`
2. Keep `Claims` struct in template (already exists)
3. Update cipher-im to import GenerateJWT from template

**Files to modify**:
1. Update `internal/apps/template/service/server/realms/jwt.go` - Add GenerateJWT function
2. Update `internal/apps/cipher/im/server/util/jwt.go` → DELETE (moved to template)
3. Update cipher-im imports to use template GenerateJWT
4. Add tests: Extend `internal/apps/template/service/server/realms/jwt_test.go`

---

## Phase 4: Public Server Infrastructure Extraction

### Problem
Cipher-IM public_server.go (298 lines) contains MASSIVE reusable infrastructure:

**Health endpoints** (lines 164-201):
- `handleServiceHealth()` - Service-to-service health check
- `handleBrowserHealth()` - Browser health check
- **IDENTICAL PATTERN** for all 9 services

**Server lifecycle** (lines 203-298):
- `Start()` - TCP listener, TLS listener, goroutine management, context cancellation
- `Shutdown()` - Graceful shutdown with timeout
- `ActualPort()` - Thread-safe port accessor
- `PublicBaseURL()` - Base URL generator
- **IDENTICAL PATTERN** for all 9 services

**Route registration** (lines 133-162):
- Session endpoints: /sessions/issue, /sessions/validate
- User endpoints: /users/register, /users/login
- Health endpoints: /health
- Middleware: browserSessionMiddleware, serviceSessionMiddleware
- **80% REUSABLE** across all services

### Extraction Plan

**Create template PublicServerBase**:
1. Extract common server infrastructure to `internal/apps/template/service/server/public_server_base.go`
2. Provide extensibility points for service-specific routes
3. Composition pattern: Services embed PublicServerBase

**PublicServerBase structure**:
```go
type PublicServerBase struct {
    bindAddress string
    port        int
    app         *fiber.App
    mu          sync.RWMutex
    shutdown    bool
    actualPort  int
    tlsMaterial *TLSMaterial
    ctx         context.Context
    cancel      context.CancelFunc
    
    // Extensibility hooks
    registerCustomRoutes func(*fiber.App)
}
```

**Common methods in PublicServerBase**:
- Start() - Full TCP/TLS listener logic
- Shutdown() - Graceful shutdown
- ActualPort() - Port accessor
- PublicBaseURL() - URL generator
- RegisterCommonRoutes() - Health, session, user endpoints
- handleServiceHealth() - Service health
- handleBrowserHealth() - Browser health

**Service-specific pattern**:
```go
type CipherIMPublicServer struct {
    *PublicServerBase  // Composition
    messageHandler *apis.MessageHandler
}

func (s *CipherIMPublicServer) registerCustomRoutes(app *fiber.App) {
    // Service-specific routes (messages)
    app.Put("/service/api/v1/messages/tx", s.messageHandler.HandleSendMessage())
}
```

**Files to create**:
1. `internal/apps/template/service/server/public_server_base.go` - Base infrastructure
2. `internal/apps/template/service/server/public_server_base_test.go` - Base tests

**Files to modify**:
1. `internal/apps/cipher/im/server/public_server.go` - Use composition with PublicServerBase
2. Extract 80% of code to template, keep 20% cipher-im specific (message routes)

---

## Phase 5: Test Infrastructure Extraction

### Problem 1: http_errors_test.go (118 lines)
Contains reusable test patterns:
- Mock HTTP servers (testMockServerOK, testMockServerError)
- Health/livez/readyz endpoint testing
- Slow response testing
- **80% REUSABLE** for all 9 services

### Problem 2: http_test.go (194 lines)
Contains anti-pattern:
- Creates new server per test function (NOT shared TestMain)
- Reusable test config initialization
- Reusable TLS client creation
- **MUST REFACTOR** to use TestMain shared server

### Extraction Plan

**Extract to template test utilities**:
1. Create `internal/apps/template/service/testutil/http_test_helpers.go`
2. Move mock server creation (testMockServerOK, testMockServerError)
3. Move TLS client creation helper
4. Move health endpoint test patterns

**Fix http_test.go anti-pattern**:
1. Create TestMain in `internal/apps/cipher/im/http_test.go`
2. Start server ONCE before all tests
3. Share server across all test functions
4. Use `t.Parallel()` for concurrent tests

**Files to create**:
1. `internal/apps/template/service/testutil/http_test_helpers.go` - Shared test utilities
2. `internal/apps/template/service/testutil/http_test_helpers_test.go` - Utility tests

**Files to modify**:
1. `internal/apps/cipher/im/http_errors_test.go` - Import from template testutil
2. `internal/apps/cipher/im/http_test.go` - Add TestMain, use shared server, import testutil

---

## Phase 6: Realm Validation Test Extraction

### Problem
`realm_validation_test.go` (223 lines) contains:
- Password validation tests for default + enterprise realms
- Username validation tests
- **100% REUSABLE** for all 9 services (tests template realms package)

### Extraction Plan

**Move to template tests**:
1. `internal/apps/cipher/im/server/realm_validation_test.go` → `internal/apps/template/service/server/realms/password_validation_test.go`
2. Tests validate template realms.ValidatePasswordForRealm function
3. Should be in template package, NOT service package

**Files to modify**:
1. Create `internal/apps/template/service/server/realms/password_validation_test.go` - Move all tests
2. Delete `internal/apps/cipher/im/server/realm_validation_test.go` - No longer needed

---

## Phase 7: Multi-Tenancy Enforcement

### Problem
Cipher-IM shows single-tenant pattern:
- `CipherIMDefaultTenantID` used as hardcoded default
- UserFactory sets TenantID to default
- **VIOLATES REQUIREMENT**: All services MUST be multi-tenant

### Solution

**Multi-tenant session pattern**:
1. Extract tenant_id from session token (already in SessionMiddleware)
2. Pass tenant_id to all repository operations
3. Enforce tenant isolation at database level
4. NO hardcoded tenant IDs

**Files to modify**:
1. `internal/apps/cipher/im/server/public_server.go` - Remove CipherIMDefaultTenantID usage
2. `internal/apps/cipher/im/server/apis/message_handler.go` - Extract tenant_id from context
3. `internal/apps/cipher/im/repository/message_repository.go` - Add tenant_id to queries
4. All handlers MUST use c.Locals("tenant_id") for multi-tenancy

**Testing**:
- Add E2E test: Multiple tenants creating messages
- Verify tenant isolation (tenant A cannot read tenant B messages)

---

## Phase 8: E2E Test Verification

### Test Coverage Requirements

**E2E tests MUST verify**:
1. Migration numbering (1001-1999 template, 2001+ cipher-im)
2. Session middleware (browser + service paths)
3. Multi-tenancy (tenant isolation)
4. Health endpoints (service + browser)
5. Docker Compose (all services start, APIs reachable)

**Files to verify**:
1. `internal/apps/cipher/im/e2e/e2e_test.go` - Update for new numbering, multi-tenancy
2. Add test: TestE2E_MultiTenantIsolation
3. Add test: TestE2E_MigrationNumbering

---

## Phase 9: Docker Compose Verification

### Verification Steps

1. **KMS Compose**: `docker compose -f deployments/kms/compose.yml up -d`
   - Verify all services start
   - Verify APIs reachable on ports 8080-8082
   - Verify health checks pass

2. **Cipher-IM Compose**: `docker compose -f deployments/compose/compose.yml up -d`
   - Verify all services start
   - Verify APIs reachable on ports 8888-8890
   - Verify PostgreSQL migrations applied correctly
   - Verify SQLite migrations applied correctly

3. **API Reachability**:
   - Test `/service/api/v1/health` endpoints
   - Test `/browser/api/v1/health` endpoints
   - Test `/admin/v1/livez` endpoints
   - Test `/admin/v1/readyz` endpoints

---

## Execution Priority (Per User Requirements)

### CRITICAL FIRST: E2E Tests
- Fix E2E tests to pass with current code
- Add E2E tests for multi-tenancy
- Add E2E tests for migration numbering

### CRITICAL SECOND: Migration Numbering
- Renumber cipher-im migrations 1005 → 2001
- Reserve template range 1001-1999
- Test migration sequence (template first, then service)

### CRITICAL THIRD: Reusable Code Extraction
- Phase 2: Session middleware
- Phase 3: JWT utilities
- Phase 4: Public server base
- Phase 5: Test infrastructure
- Phase 6: Realm validation tests
- Phase 7: Multi-tenancy enforcement

### CRITICAL FOURTH: Verification
- All unit tests pass
- All integration tests pass
- All E2E tests pass
- All docker compose commands work
- APIs reachable and functional

---

## Success Criteria

**Migration numbering**:
- ✅ Template migrations: 1001-1999
- ✅ Cipher-IM migrations: 2001+
- ✅ Can add template migrations without conflicts

**Code extraction**:
- ✅ SessionMiddleware in service-template
- ✅ GenerateJWT in service-template
- ✅ PublicServerBase in service-template
- ✅ Test utilities in service-template
- ✅ Realm validation tests in service-template

**Multi-tenancy**:
- ✅ No hardcoded tenant IDs
- ✅ Tenant extracted from session
- ✅ Tenant isolation verified in E2E tests

**Testing**:
- ✅ All cipher-im tests pass
- ✅ All service-template tests pass
- ✅ E2E tests verify migration numbering
- ✅ E2E tests verify multi-tenancy
- ✅ Docker Compose verified working

**Reusability**:
- ✅ Other 8 services can import from service-template
- ✅ No code duplication across services
- ✅ Clear separation: shared vs service-specific

---

## Estimated Effort

**Phase 1 (Migration numbering)**: 30 minutes
**Phase 2 (Session middleware)**: 45 minutes
**Phase 3 (JWT utilities)**: 20 minutes
**Phase 4 (Public server base)**: 90 minutes
**Phase 5 (Test infrastructure)**: 60 minutes
**Phase 6 (Realm validation)**: 20 minutes
**Phase 7 (Multi-tenancy)**: 45 minutes
**Phase 8 (E2E tests)**: 60 minutes
**Phase 9 (Docker Compose)**: 30 minutes

**Total**: ~6 hours of implementation + testing

---

## Risk Assessment

**HIGH RISK**:
- Migration renumbering could break existing deployments (MITIGATION: Test thoroughly before commit)
- Public server base extraction complex (MITIGATION: Incremental composition pattern)

**MEDIUM RISK**:
- Session middleware dependencies (MITIGATION: Verify SessionManagerService in template)
- Multi-tenancy enforcement (MITIGATION: Add comprehensive E2E tests)

**LOW RISK**:
- JWT utilities extraction (simple function move)
- Test infrastructure extraction (no runtime impact)
- Realm validation tests (pure test code move)

---

## Next Steps

1. Execute Phase 1 (migration numbering) - CRITICAL BLOCKER
2. Execute Phase 8 (E2E tests) - CRITICAL FIRST per user
3. Execute Phases 2-7 (extraction) - CRITICAL THIRD per user
4. Execute Phase 9 (Docker Compose) - Final verification
5. Commit all changes with comprehensive testing
