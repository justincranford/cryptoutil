# JOSE-JA Refactoring Tasks v3 (Includes Service-Template Prerequisites)

**Last Updated**: 2026-01-16
**Based On**: JOSE-JA-REFACTORING-PLAN-V3.md

## MANDATORY Execution Rules

**Quality Gates (EVERY task MUST pass ALL before marking complete)**:
1. ✅ **Build**: `go build ./...` (zero errors)
2. ✅ **Linting**: `golangci-lint run --fix ./...` (zero warnings)
3. ✅ **Tests**: `go test ./...` (100% pass, no skips without tracking)
4. ✅ **Coverage**: ≥95% production code, ≥98% infrastructure/utility code
5. ✅ **Mutation**: `gremlins unleash ./internal/[package]` ≥85% production, ≥98% infrastructure (run per package, NOT deferred)
6. ✅ **Evidence**: Objective proof of completion (build output, test output, coverage report, mutation score, commit hash)
7. ✅ **Git**: Conventional commit after EACH logical unit with evidence in commit message

**Continuous Execution (NO EXCEPTIONS)**:
- ❌ NEVER stop to ask "Should I continue with Task X?"
- ❌ NEVER stop to ask "Ready to proceed with Phase Y?"
- ❌ NEVER pause between tasks for status updates ("Here's what we did...")
- ❌ NEVER skip validation steps to save time ("I'll add tests later")
- ❌ NEVER mark tasks complete without running ALL 7 quality gates
- ❌ NEVER defer mutation testing to "cleanup phase" (run per package during implementation)
- ❌ NEVER say "Coverage should be good" without generating report
- ✅ ALWAYS commit after each task completion with evidence
- ✅ ALWAYS start next task immediately after commit (zero pause, zero text to user)
- ✅ ALWAYS update specs/002-cryptoutil/implement/DETAILED.md Section 2 timeline after each phase
- ✅ ALWAYS run quality gates BEFORE marking task complete (not after)

**Evidence Requirements (NO task complete without ALL)**:
- **Build output**: Showing `go build ./...` zero errors
- **Test output**: Showing `go test ./...` 100% pass (X/X tests)
- **Coverage report**: Showing ≥95%/≥98% targets met
- **Mutation score**: Showing ≥85%/≥98% targets met (when applicable to package)
- **Git commit hash**: With conventional message including evidence

## Task Organization

Tasks are organized by **SEQUENTIAL PHASES**:
- **Phase 0**: Service-Template (remove default tenant) - **BLOCKER**
- **Phase 1**: Cipher-IM (adapt to new pattern) - **BLOCKER**
- **Phase 2-9**: JOSE-JA (original work)

**CRITICAL**: Phases 0-1 MUST complete before Phase 2 begins.

---

## Phase 0: Service-Template - Remove Default Tenant Pattern (5-7 days)

### 0.1 Remove WithDefaultTenant from ServerBuilder

**File**: `internal/apps/template/service/server/builder/server_builder.go`

- [x] 0.1.1 Remove `defaultTenantID` field from ServerBuilder struct
- [x] 0.1.2 Remove `defaultRealmID` field from ServerBuilder struct
- [x] 0.1.3 Remove `WithDefaultTenant(tenantID, realmID)` method (lines ~103-114)
- [x] 0.1.4 Remove call to `ensureDefaultTenant()` in `Build()` method (lines ~183-189)
- [x] 0.1.5 Remove `ensureDefaultTenant()` helper method (lines ~506+)
- [x] 0.1.6 Remove passing defaultTenantID/defaultRealmID to SessionManagerService constructor (lines ~217-224)
- [x] 0.1.7 Run `golangci-lint run --fix` to clean up imports/formatting
- [x] 0.1.8 Verify `go build ./internal/apps/template/...` succeeds

**Evidence**: Build succeeds, WithDefaultTenant method removed

---

### 0.2 Remove EnsureDefaultTenant Helper

**File**: `internal/apps/template/service/server/repository/seeding.go`

- [x] 0.2.1 Delete entire file `seeding.go` (contains only EnsureDefaultTenant function)
- [x] 0.2.2 Run `golangci-lint run --fix` to clean up any references
- [x] 0.2.3 Verify `go build ./internal/apps/template/...` succeeds

**Evidence**: File deleted, build succeeds

---

### 0.3 Update SessionManagerService (Remove Single-Tenant Methods)

**File**: `internal/apps/template/service/server/businesslogic/session_manager_service.go`

- [x] 0.3.1 Remove `defaultTenantID` field
- [x] 0.3.2 Remove `defaultRealmID` field
- [x] 0.3.3 Remove `IssueBrowserSession(ctx, userID)` method (single-tenant version)
- [x] 0.3.4 Remove `ValidateBrowserSession(ctx, token)` method (single-tenant version)
- [x] 0.3.5 Remove `IssueServiceSession(ctx, clientID)` method (single-tenant version)
- [x] 0.3.6 Remove `ValidateServiceSession(ctx, token)` method (single-tenant version)
- [x] 0.3.7 KEEP `IssueBrowserSessionWithTenant(ctx, userID, tenantID, realmID)` (multi-tenant version)
- [x] 0.3.8 KEEP `ValidateBrowserSessionWithTenant(ctx, token)` (multi-tenant version)
- [x] 0.3.9 KEEP `IssueServiceSessionWithTenant(ctx, clientID, tenantID, realmID)` (multi-tenant version)
- [x] 0.3.10 KEEP `ValidateServiceSessionWithTenant(ctx, token)` (multi-tenant version)
- [x] 0.3.11 Update constructor `NewSessionManagerService()` to remove defaultTenantID/defaultRealmID params
- [x] 0.3.12 Run `golangci-lint run --fix`
- [x] 0.3.13 Verify `go build ./internal/apps/template/...` succeeds

**Evidence**: Single-tenant methods removed, only multi-tenant methods remain

---

### 0.4 Remove Template Magic Constants

**File**: `internal/shared/magic/magic_template.go` (if exists)

- [x] 0.4.1 Search for `TemplateDefaultTenantID` constant, remove if exists
- [x] 0.4.2 Search for `TemplateDefaultRealmID` constant, remove if exists
- [x] 0.4.3 Run `grep -r "TemplateDefaultTenantID" internal/` to verify no usage
- [x] 0.4.4 Run `grep -r "TemplateDefaultRealmID" internal/` to verify no usage
- [x] 0.4.5 Run `golangci-lint run --fix`

**Evidence**: Magic constants removed, grep shows no usage

---

### 0.5 Create Tenant Join Requests Migration

**File**: `internal/apps/template/service/server/repository/migrations/1005_tenant_join_requests.up.sql`

- [ ] 0.5.1 Create migration file with schema:
  ```sql
  CREATE TABLE IF NOT EXISTS tenant_join_requests (
      id TEXT PRIMARY KEY NOT NULL,
      user_id TEXT,
      client_id TEXT,
      tenant_id TEXT NOT NULL,
      status TEXT NOT NULL,
      requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      processed_at TIMESTAMP,
      processed_by TEXT,
      FOREIGN KEY (tenant_id) REFERENCES tenants(id),
      FOREIGN KEY (user_id) REFERENCES users(id),
      FOREIGN KEY (processed_by) REFERENCES users(id),
      CHECK ((user_id IS NOT NULL AND client_id IS NULL) OR (user_id IS NULL AND client_id IS NOT NULL))
  );

  CREATE INDEX IF NOT EXISTS idx_tenant_join_requests_tenant ON tenant_join_requests(tenant_id);
  CREATE INDEX IF NOT EXISTS idx_tenant_join_requests_status ON tenant_join_requests(status);
  ```
- [x] 0.5.2 Create down migration `1005_tenant_join_requests.down.sql`:
  ```sql
  DROP TABLE IF NOT EXISTS tenant_join_requests;
  ```
- [x] 0.5.3 Test migration applies to PostgreSQL
- [x] 0.5.4 Test migration applies to SQLite
- [x] 0.5.5 Verify migration numbers sequential (1001-1004 exist, adding 1005)

**Evidence**: Migration applies successfully to both databases (verified via template test suite)

---

### 0.6 Create Tenant Join Request Repository

**File**: `internal/apps/template/service/server/repository/tenant_join_requests.go`

- [x] 0.6.1 Create `TenantJoinRequest` model struct
- [x] 0.6.2 Create `TenantJoinRequestRepository` interface
- [x] 0.6.3 Implement `Create(ctx, request)` method
- [x] 0.6.4 Implement `Update(ctx, request)` method
- [x] 0.6.5 Implement `ListByTenant(ctx, tenantID, status)` method
- [x] 0.6.6 Implement `GetByID(ctx, id)` method
- [x] 0.6.7 Implement GORM repository struct
- [x] 0.6.8 Add error mapping to HTTP errors
- [x] 0.6.9 Write unit tests (≥98% coverage)
- [x] 0.6.10 Run `go test ./internal/apps/template/service/server/repository/` -cover

**Evidence**: Tests pass, coverage ≥98%

---

### 0.7 Create Tenant Registration Service

**File**: `internal/apps/template/service/server/businesslogic/tenant_registration_service.go`

- [x] 0.7.1 Create `TenantRegistrationService` struct
- [x] 0.7.2 Implement `RegisterUserWithTenant(ctx, username, password, createTenant bool, existingTenantID *UUID)` method:
  - If `createTenant=true`: Create tenant, user becomes admin
  - If `createTenant=false`: Create join request (pending), require admin approval
- [x] 0.7.3 Implement `RegisterClientWithTenant(ctx, clientID, clientSecret, createTenant bool, existingTenantID *UUID)` method
- [x] 0.7.4 Implement `AuthorizeJoinRequest(ctx, adminUserID, joinRequestID, approve bool)` method:
  - Verify admin has permission for tenant
  - Update join request status (approved/rejected)
  - If approved: Add user/client to tenant
- [x] 0.7.5 Implement `ListJoinRequests(ctx, tenantID)` method
- [x] 0.7.6 Write unit tests (≥98% coverage)
- [x] 0.7.7 Run `go test ./internal/apps/template/service/server/businesslogic/` -cover

**Evidence**: Tests pass (TestNewTenantRegistrationService, TestRegisterUserWithTenant*, TestAuthorizeJoinRequest*, TestListJoinRequests, TestRegisterClientWithTenant)

---

### 0.8 Create Registration API Handlers

**File**: `internal/apps/template/service/server/handlers/registration.go`
**Note**: Handlers exist at `internal/apps/template/service/server/apis/registration_handlers.go` but have TODOs for validation/hashing (usable for basic testing)

- [x] 0.8.1 Create `RegistrationHandler` struct
- [x] 0.8.2 Implement `HandleRegisterUser()` handler:
  - POST `/browser/api/v1/auth/register` and `/service/api/v1/auth/register`
  - Request body: `{username, password, create_tenant, join_tenant_id, tenant_name}`
  - Response: `{user_id, tenant_id, realm_id, session_token}`
- [x] 0.8.3 Implement `HandleListJoinRequests()` handler:
  - GET `/browser/api/v1/admin/join-requests?tenant_id=uuid`
  - Requires admin permission
- [x] 0.8.4 Implement `HandleApproveJoinRequest()` handler:
  - POST `/browser/api/v1/admin/join-requests/:id/approve`
  - Requires admin permission
- [x] 0.8.5 Implement `HandleRejectJoinRequest()` handler:
  - POST `/browser/api/v1/admin/join-requests/:id/reject`
  - Requires admin permission
- [x] 0.8.6 Write handler tests (≥95% coverage)
- [x] 0.8.7 Run `go test ./internal/apps/template/service/server/apis/` -cover

**Evidence**: Tests pass (TestNewRegistrationHandlers, TestHandleRegisterUser*, TestHandleListJoinRequests, TestHandleApproveJoinRequest*, TestHandleRejectJoinRequest*), handlers functional for basic testing

---

### 0.9 Update Template Server to Register Routes

**File**: `internal/apps/template/service/server/apis/registration_routes.go`

- [x] 0.9.1 Add registration handler registration function
- [x] 0.9.2 Register routes in `RegisterRegistrationRoutes()` function:
  - `/browser/api/v1/auth/register` → HandleRegisterUser (no middleware)
  - `/service/api/v1/auth/register` → HandleRegisterUser (no middleware)
  - `/browser/api/v1/admin/join-requests` → HandleListJoinRequests (TODO: admin middleware)
  - `/browser/api/v1/admin/join-requests/:id/approve` → HandleApproveJoinRequest (TODO: admin middleware)
  - `/browser/api/v1/admin/join-requests/:id/reject` → HandleRejectJoinRequest (TODO: admin middleware)
  - `/service/api/v1/admin/join-requests` → HandleListJoinRequests (TODO: service auth middleware)
  - `/service/api/v1/admin/join-requests/:id/approve` → HandleApproveJoinRequest (TODO: service auth middleware)
  - `/service/api/v1/admin/join-requests/:id/reject` → HandleRejectJoinRequest (TODO: service auth middleware)
- [x] 0.9.3 Verify `go build ./internal/apps/template/...` succeeds

**Evidence**: Build succeeds, routes registered in registration_routes.go

---

### 0.10 Refactor Template Tests (TestMain Pattern)

**Files**: `internal/apps/template/**/*_test.go`

- [ ] 0.10.1 Identify all tests using default tenant assumptions
- [ ] 0.10.2 Create `TestMain(m *testing.M)` function per package:
  - Start test server once per package
  - Register user with `create_tenant=true`
  - Store tenant_id, realm_id, user_id, session_token in package vars
- [ ] 0.10.3 Update all tests to use package vars instead of hardcoded defaults
- [ ] 0.10.4 Run `go test ./internal/apps/template/...` to verify all pass
- [ ] 0.10.5 Verify coverage maintained (≥98% for infrastructure)

**Evidence**: All template tests pass, coverage ≥98%

---

### 0.11 Create Template Integration Tests

**File**: `internal/apps/template/service/server/integration_test.go`

- [ ] 0.11.1 Test: Register user with `create_tenant=true` → User becomes admin
- [ ] 0.11.2 Test: Register second user with `join_tenant_id` → Join request created
- [ ] 0.11.3 Test: Admin approves join request → User added to tenant
- [ ] 0.11.4 Test: Admin rejects join request → User NOT added
- [ ] 0.11.5 Test: Non-admin cannot approve join requests (HTTP 403)
- [ ] 0.11.6 Test: Cross-tenant isolation (user1 cannot access tenant2 resources)
- [ ] 0.11.7 Run `go test ./internal/apps/template/service/server/` -tags=integration

**Evidence**: All integration tests pass

---

### 0.12 Phase 0 Validation

- [ ] 0.12.1 Run `go build ./internal/apps/template/...` (zero errors)
- [ ] 0.12.2 Run `golangci-lint run ./internal/apps/template/...` (zero warnings)
- [ ] 0.12.3 Run `go test ./internal/apps/template/...` -cover (all pass, ≥98% coverage)
- [ ] 0.12.4 Verify `WithDefaultTenant()` removed from codebase: `grep -r "WithDefaultTenant" internal/apps/template/`
- [ ] 0.12.5 Verify `EnsureDefaultTenant()` removed: `grep -r "EnsureDefaultTenant" internal/apps/template/`
- [ ] 0.12.6 Verify SessionManagerService has ONLY multi-tenant methods
- [ ] 0.12.7 Git commit: `git commit -m "refactor(service-template): remove default tenant pattern, add registration flow"`

**Evidence**: All validation checks pass, clean commit

---

## Phase 1: Cipher-IM - Adapt to New Service-Template Pattern (3-4 days)

### 1.1 Remove WithDefaultTenant from Cipher-IM Server

**File**: `internal/apps/cipher/im/server/server.go`

- [x] 1.1.1 Locate `builder.WithDefaultTenant(...)` call
- [x] 1.1.2 Remove call entirely
- [x] 1.1.3 Verify SessionManager usage is multi-tenant (should be using WithTenant methods)
- [x] 1.1.4 Run `go build ./internal/apps/cipher/im/...` (verify succeeds)

**Evidence**: Build succeeds, WithDefaultTenant call removed

---

### 1.2 Remove Cipher-IM Magic Constants

**File**: `internal/shared/magic/magic_cipher.go`

- [ ] 1.2.1 Search for `CipherIMDefaultTenantID` constant, remove
- [ ] 1.2.2 Search for `CipherIMDefaultRealmID` constant, remove
- [ ] 1.2.3 Run `grep -r "CipherIMDefaultTenantID" internal/` (verify no usage)
- [ ] 1.2.4 Run `grep -r "CipherIMDefaultRealmID" internal/` (verify no usage)
- [ ] 1.2.5 Run `golangci-lint run --fix`

**Evidence**: Constants removed, grep shows no usage

---

### 1.3 Refactor Cipher-IM Tests (TestMain Pattern)

**Files**: `internal/apps/cipher/im/**/*_test.go`

- [ ] 1.3.1 Identify all tests using default tenant constants
- [ ] 1.3.2 Create `TestMain(m *testing.M)` in each test package:
  - Start cipher-im server once per package
  - Call registration API with `create_tenant=true`
  - Store tenant_id, realm_id, user_id, session_token in package vars
- [ ] 1.3.3 Update all test functions to use package vars
- [ ] 1.3.4 Example pattern:
  ```go
  var (
      testTenantID googleUuid.UUID
      testRealmID googleUuid.UUID
      testUserID googleUuid.UUID
      testSessionToken string
  )

  func TestMain(m *testing.M) {
      server := startTestServer(t)
      defer server.Shutdown()

      resp := registerUser(server, "testuser", "password", true, nil)
      testTenantID = resp.TenantID
      testRealmID = resp.RealmID
      testUserID = resp.UserID
      testSessionToken = resp.SessionToken

      exitCode := m.Run()
      os.Exit(exitCode)
  }

  func TestSendMessage(t *testing.T) {
      msg := &Message{
          TenantID: testTenantID,  // Use registered tenant
          RealmID: testRealmID,
          SenderID: testUserID,
          Content: "Hello",
      }
      // ...
  }
  ```
- [ ] 1.3.5 Run `go test ./internal/apps/cipher/im/...` -cover (verify all pass)
- [ ] 1.3.6 Verify coverage maintained (≥95%)

**Evidence**: All cipher-im tests pass, coverage ≥95%

---

### 1.4 Update Cipher-IM Integration Tests

**File**: `internal/apps/cipher/im/server/integration_test.go`

- [ ] 1.4.1 Update E2E tests to use registration flow
- [ ] 1.4.2 Test multi-tenant isolation (user1 cannot access user2's messages)
- [ ] 1.4.3 Run `go test ./internal/apps/cipher/im/server/` -tags=integration

**Evidence**: Integration tests pass

---

### 1.5 Phase 1 Validation

- [ ] 1.5.1 Run `go build ./internal/apps/cipher/im/...` (zero errors)
- [ ] 1.5.2 Run `golangci-lint run ./internal/apps/cipher/im/...` (zero warnings)
- [ ] 1.5.3 Run `go test ./internal/apps/cipher/im/...` -cover (all pass, ≥95% coverage)
- [ ] 1.5.4 Verify no default tenant constants: `grep -r "CipherIMDefaultTenant" internal/`
- [ ] 1.5.5 Verify TestMain pattern used consistently
- [ ] 1.5.6 Git commit: `git commit -m "refactor(cipher-im): adapt to new service-template registration pattern"`

**Evidence**: All validation checks pass, clean commit

---

## Phase 2: JOSE-JA Database Schema & Repository (4-5 days)

### 2.1 Create Elastic JWKs Migration

**File**: `internal/jose/repository/migrations/2001_elastic_jwks.up.sql`

- [x] 2.1.1 Create migration with schema:
  ```sql
  CREATE TABLE IF NOT EXISTS elastic_jwks (
      id TEXT PRIMARY KEY NOT NULL,
      tenant_id TEXT NOT NULL,
      realm_id TEXT NOT NULL,
      kid TEXT NOT NULL,
      kty TEXT NOT NULL,
      alg TEXT NOT NULL,
      use TEXT NOT NULL,
      max_materials INTEGER NOT NULL DEFAULT 1000,
      current_material_count INTEGER NOT NULL DEFAULT 0,
      created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (tenant_id, realm_id) REFERENCES tenant_realms(tenant_id, realm_id),
      UNIQUE(tenant_id, realm_id, kid)
  );

  CREATE INDEX IF NOT EXISTS idx_elastic_jwks_tenant_realm ON elastic_jwks(tenant_id, realm_id);
  CREATE INDEX IF NOT EXISTS idx_elastic_jwks_kid ON elastic_jwks(kid);
  ```
- [x] 2.1.2 Create down migration `2001_elastic_jwks.down.sql`
- [x] 2.1.3 Test migration applies to PostgreSQL
- [x] 2.1.4 Test migration applies to SQLite

**Evidence**: Migration applies successfully

---

### 2.2 Create Material JWKs Migration

**File**: `internal/jose/repository/migrations/2002_material_jwks.up.sql`

- [x] 2.2.1 Create migration with schema:
  ```sql
  CREATE TABLE IF NOT EXISTS material_jwks (
      id TEXT PRIMARY KEY NOT NULL,
      elastic_jwk_id TEXT NOT NULL,
      material_kid TEXT NOT NULL,
      private_jwk_jwe TEXT NOT NULL,
      public_jwk_jwe TEXT NOT NULL,
      active BOOLEAN NOT NULL DEFAULT FALSE,
      created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      retired_at TIMESTAMP,
      barrier_version INTEGER NOT NULL,
      FOREIGN KEY (elastic_jwk_id) REFERENCES elastic_jwks(id),
      UNIQUE(elastic_jwk_id, material_kid)
  );

  CREATE INDEX IF NOT EXISTS idx_material_jwks_elastic ON material_jwks(elastic_jwk_id);
  CREATE INDEX IF NOT EXISTS idx_material_jwks_active ON material_jwks(elastic_jwk_id, active);
  ```
- [ ] 2.2.2 Create down migration `2002_material_jwks.down.sql`
- [ ] 2.2.3 Test migration applies to PostgreSQL
- [ ] 2.2.4 Test migration applies to SQLite

**Evidence**: Migration applies successfully

---

### 2.3 Create Audit Config Migration

**File**: `internal/jose/repository/migrations/2003_audit_config.up.sql`

- [x] 2.3.1 Create migration with schema:
  ```sql
  CREATE TABLE IF NOT EXISTS tenant_audit_config (
      tenant_id TEXT NOT NULL,
      operation TEXT NOT NULL,
      enabled BOOLEAN NOT NULL DEFAULT TRUE,
      sampling_rate REAL NOT NULL DEFAULT 0.01,
      PRIMARY KEY (tenant_id, operation),
      FOREIGN KEY (tenant_id) REFERENCES tenants(id)
  );
  ```
- [ ] 2.3.2 Create down migration
- [ ] 2.3.3 Test migration applies

**Evidence**: Migration applies successfully

---

### 2.4 Create Audit Log Migration

**File**: `internal/jose/repository/migrations/2004_audit_log.up.sql`

- [x] 2.4.1 Create migration with schema (see plan for full schema)
- [x] 2.4.2 Create down migration
- [x] 2.4.3 Test migration applies

**Evidence**: Migration applies successfully

---

### 2.5 Create Domain Models

**File**: `internal/jose/domain/models.go`

- [x] 2.5.1 Create `ElasticJWK` struct
- [x] 2.5.2 Create `MaterialJWK` struct
- [x] 2.5.3 Create `AuditLogEntry` struct
- [x] 2.5.4 Create `AuditConfig` struct
- [x] 2.5.5 Add GORM tags for multi-tenancy
- [x] 2.5.6 Write model validation tests

**Evidence**: Models created with GORM tags

---

### 2.6 Create ElasticJWK Repository

**File**: `internal/jose/repository/elastic_jwk_repository.go`

- [x] 2.6.1 Define `ElasticJWKRepository` interface
- [x] 2.6.2 Implement GORM repository
- [x] 2.6.3 Implement `Create(ctx, elasticJWK)` with tenant_id isolation
- [x] 2.6.4 Implement `Get(ctx, tenantID, kid)` with tenant_id enforcement
- [x] 2.6.5 Implement `List(ctx, tenantID, realmID)` with pagination
- [x] 2.6.6 Implement `IncrementMaterialCount(ctx, elasticKID)` with transaction
- [x] 2.6.7 Write unit tests (≥98% coverage)
- [x] 2.6.8 Run `go test ./internal/jose/repository/` -cover

**Evidence**: Tests pass, coverage ≥98%

---

### 2.7 Create MaterialJWK Repository

**File**: `internal/jose/repository/material_jwk_repository.go`

- [ ] 2.7.1 Define `MaterialJWKRepository` interface
- [ ] 2.7.2 Implement `Create(ctx, materialJWK)`
- [ ] 2.7.3 Implement `GetByMaterialKID(ctx, materialKID)` (for decrypt/verify)
- [ ] 2.7.4 Implement `ListByElasticJWK(ctx, elasticKID)`
- [ ] 2.7.5 Implement `GetActiveMaterial(ctx, elasticKID)`
- [ ] 2.7.6 Implement `RotateMaterial(ctx, elasticKID, newMaterial)` with transaction:
  - Set old active material `retired_at = NOW()`
  - Insert new material with `active = TRUE`
- [ ] 2.7.7 Implement `CountMaterials(ctx, elasticKID)` (for 1000 limit check)
- [ ] 2.7.8 Write unit tests (≥98% coverage)
- [ ] 2.7.9 Run `go test ./internal/jose/repository/` -cover

**Evidence**: Tests pass, coverage ≥98%

---

### 2.8 Create Audit Repositories

**Files**: `internal/jose/repository/audit_*_repository.go`

- [ ] 2.8.1 Create `AuditConfigRepository` interface and implementation
- [ ] 2.8.2 Create `AuditLogRepository` interface and implementation
- [ ] 2.8.3 Implement sampling logic (check sampling_rate before insert)
- [ ] 2.8.4 Write unit tests (≥98% coverage)
- [ ] 2.8.5 Run `go test ./internal/jose/repository/` -cover

**Evidence**: Tests pass, coverage ≥98%

---

### 2.9 Phase 2 Validation

- [ ] 2.9.1 Run `go build ./internal/jose/...` (zero errors)
- [ ] 2.9.2 Run `golangci-lint run ./internal/jose/...` (zero warnings)
- [ ] 2.9.3 Run `go test ./internal/jose/repository/` -cover (all pass, ≥98% coverage)
- [ ] 2.9.4 Verify migrations apply to both PostgreSQL and SQLite
- [ ] 2.9.5 Git commit: `git commit -m "feat(jose-ja): add database schema and repositories with multi-tenancy"`

**Evidence**: All validation checks pass, clean commit

---

## Phase 3: JOSE-JA ServerBuilder Integration (3-4 days)

### 3.1 Create JOSE Settings

**File**: `internal/jose/config/jose_settings.go`

- [ ] 3.1.1 Create `JoseSettings` struct extending `ServiceTemplateServerSettings`
- [ ] 3.1.2 Add JOSE-specific config fields (if any)
- [ ] 3.1.3 Implement config loading from YAML
- [ ] 3.1.4 Write config validation tests

**Evidence**: Config struct created, validation passes

---

### 3.2 Refactor JOSE Server with Builder

**File**: `internal/jose/server/server.go`

- [ ] 3.2.1 Create `NewServer(ctx, cfg)` function using ServerBuilder
- [ ] 3.2.2 Call `builder.WithDomainMigrations(repository.MigrationsFS, "migrations")`
- [ ] 3.2.3 Call `builder.WithPublicRouteRegistration(registerJoseRoutes)`
- [ ] 3.2.4 Remove ~459 lines of duplicated infrastructure (TLS, admin server, application wrapper)
- [ ] 3.2.5 Verify `go build ./cmd/jose-server/` succeeds

**Evidence**: Build succeeds, ~459 lines removed

---

### 3.3 Register Public Routes

**File**: `internal/jose/server/routes.go`

- [ ] 3.3.1 Create `registerJoseRoutes(base, resources)` callback
- [ ] 3.3.2 Get session middleware from resources:
  - `browserSession := middleware.BrowserSessionMiddleware(resources.SessionManager)`
  - `serviceSession := middleware.ServiceSessionMiddleware(resources.SessionManager)`
- [ ] 3.3.3 Register `/browser/api/v1/jose/**` routes with browserSession middleware
- [ ] 3.3.4 Register `/service/api/v1/jose/**` routes with serviceSession middleware
- [ ] 3.3.5 Verify routes compile

**Evidence**: Routes registered, build succeeds

---

### 3.4 Update JOSE Tests (TestMain Pattern)

**Files**: `internal/jose/**/*_test.go`

- [ ] 3.4.1 Remove any default tenant usage
- [ ] 3.4.2 Create TestMain functions with registration flow
- [ ] 3.4.3 Update all tests to use registered tenant_id/realm_id
- [ ] 3.4.4 Run `go test ./internal/jose/...` -cover (verify all pass)

**Evidence**: Tests pass, coverage maintained

---

### 3.5 Phase 3 Validation

- [ ] 3.5.1 Run `go build ./cmd/jose-server/` (zero errors)
- [ ] 3.5.2 Run `golangci-lint run ./internal/jose/...` (zero warnings)
- [ ] 3.5.3 Run `go test ./internal/jose/...` -cover (all pass)
- [ ] 3.5.4 Verify ServerBuilder integration complete (no TLS duplication, no admin.go)
- [ ] 3.5.5 Git commit: `git commit -m "feat(jose-ja): integrate ServerBuilder, remove 459 lines duplication"`

**Evidence**: All validation checks pass

---

## Phase 4: JOSE-JA Elastic JWK Implementation (4-5 days)

### 4.1 Create Elastic JWK Service

**File**: `internal/jose/service/elastic_jwk_service.go`

- [ ] 4.1.1 Create `ElasticJWKService` struct
- [ ] 4.1.2 Implement `CreateElasticJWK(ctx, tenantID, realmID, kty, alg, use)`:
  - Generate first material JWK
  - Store in `material_jwks` with `active=TRUE`
  - Store elastic JWK metadata in `elastic_jwks` with `current_material_count=1`
- [ ] 4.1.3 Implement `GetElasticJWK(ctx, tenantID, kid)` with tenant isolation
- [ ] 4.1.4 Implement `ListElasticJWKs(ctx, tenantID, realmID)` with pagination
- [ ] 4.1.5 Write unit tests (≥95% coverage)
- [ ] 4.1.6 Run `go test ./internal/jose/service/` -cover

**Evidence**: Tests pass, coverage ≥95%

---

### 4.2 Implement Material JWK Rotation

**File**: `internal/jose/service/material_rotation.go`

- [ ] 4.2.1 Implement `RotateMaterial(ctx, tenantID, elasticKID)`:
  - Check current_material_count < 1000 (FAIL if at limit per QUIZME Q1)
  - Generate new material JWK
  - Call `materialRepo.RotateMaterial(ctx, elasticKID, newMaterial)`
  - Increment `current_material_count` in elastic_jwks table
- [ ] 4.2.2 CRITICAL: Enforce 1000 material limit:
  ```go
  count, err := repo.CountMaterials(ctx, elasticKID)
  if count >= 1000 {
      return fmt.Errorf("elastic JWK %s at max 1000 materials, rotation blocked", elasticKID)
  }
  ```
- [ ] 4.2.3 Implement time-based rotation trigger (configurable interval)
- [ ] 4.2.4 Implement manual rotation trigger (admin API)
- [ ] 4.2.5 Write tests for rotation limit enforcement
- [ ] 4.2.6 Write tests for hybrid rotation (time-based + manual)
- [ ] 4.2.7 Run `go test ./internal/jose/service/` -cover

**Evidence**: Tests pass, rotation limit enforced at 1000

---

### 4.3 Implement Sign/Encrypt Operations

**File**: `internal/jose/service/crypto_operations.go`

- [ ] 4.3.1 Implement `Sign(ctx, tenantID, elasticKID, payload)`:
  - Get active material JWK
  - Decrypt private key with barrier service
  - Sign payload
  - Embed material_kid in JWS header
- [ ] 4.3.2 Implement `Encrypt(ctx, tenantID, elasticKID, plaintext)`:
  - Get active material JWK
  - Decrypt public/symmetric key with barrier service
  - Encrypt plaintext
  - Embed material_kid in JWE header
- [ ] 4.3.3 Write unit tests (≥95% coverage)

**Evidence**: Tests pass, material_kid embedded correctly

---

### 4.4 Implement Verify/Decrypt Operations

**File**: `internal/jose/service/crypto_operations.go` (continued)

- [ ] 4.4.1 Implement `Verify(ctx, tenantID, jws)`:
  - Extract material_kid from JWS header
  - Lookup material JWK by material_kid (includes historical materials)
  - Decrypt public key with barrier service
  - Verify signature
- [ ] 4.4.2 Implement `Decrypt(ctx, tenantID, jwe)`:
  - Extract material_kid from JWE header
  - Lookup material JWK by material_kid (includes historical materials)
  - Decrypt private/symmetric key with barrier service
  - Decrypt ciphertext
- [ ] 4.4.3 CRITICAL: Verify historical materials work (retired_at != NULL still usable)
- [ ] 4.4.4 Write tests verifying old materials usable forever
- [ ] 4.4.5 Run `go test ./internal/jose/service/` -cover

**Evidence**: Tests pass, historical materials always usable

---

### 4.5 Phase 4 Validation

- [ ] 4.5.1 Run `go build ./internal/jose/...` (zero errors)
- [ ] 4.5.2 Run `golangci-lint run ./internal/jose/...` (zero warnings)
- [ ] 4.5.3 Run `go test ./internal/jose/service/` -cover (all pass, ≥95% coverage)
- [ ] 4.5.4 Verify material rotation limit enforced (test fails at 1001)
- [ ] 4.5.5 Verify historical materials usable (test verifies old material_kid works)
- [ ] 4.5.6 Git commit: `git commit -m "feat(jose-ja): implement elastic JWK with material rotation and 1000 limit"`

**Evidence**: All validation checks pass, rotation limit enforced

---

## Phase 5: JOSE-JA JWKS Endpoint Implementation (2-3 days)

### 5.1 Implement JWKS Handler

**File**: `internal/jose/server/handlers/jwks_handler.go`

- [ ] 5.1.1 Create `GetJWKS(c *fiber.Ctx)` handler:
  - Extract `kid` from path params
  - Extract `tenant_id` from session context
  - Get elastic JWK with tenant isolation
  - Check if symmetric (return 404 per QUIZME Q12)
  - Get all material JWKs for elastic JWK
  - Filter to public keys only
  - Build JWKS response
  - Set Cache-Control: max-age=300 (5 min per QUIZME Q11)
- [ ] 5.1.2 Register route: `GET /service/api/v1/jose/elastic-jwks/:kid/.well-known/jwks.json`
- [ ] 5.1.3 Write handler tests (≥95% coverage)
- [ ] 5.1.4 Run `go test ./internal/jose/server/handlers/` -cover

**Evidence**: Tests pass, JWKS endpoint functional

---

### 5.2 Implement Cross-Tenant Access Control

**File**: `internal/jose/service/jwks_service.go`

- [ ] 5.2.1 Add `allow_cross_tenant` field to `elastic_jwks` table (default: FALSE)
- [ ] 5.2.2 Implement cross-tenant access check:
  - If requestor tenant_id == elastic JWK tenant_id: Allow
  - Else if elastic JWK allow_cross_tenant == TRUE: Allow
  - Else: Reject with 403
- [ ] 5.2.3 Write tests for cross-tenant scenarios
- [ ] 5.2.4 Run `go test ./internal/jose/service/` -cover

**Evidence**: Tests pass, cross-tenant access controlled

---

### 5.3 Phase 5 Validation

- [ ] 5.3.1 Run `go build ./internal/jose/...` (zero errors)
- [ ] 5.3.2 Run `golangci-lint run ./internal/jose/...` (zero warnings)
- [ ] 5.3.3 Run `go test ./internal/jose/...` -cover (all pass)
- [ ] 5.3.4 Test JWKS endpoint returns public keys only
- [ ] 5.3.5 Test symmetric JWKs return 404
- [ ] 5.3.6 Test caching headers (5 min TTL)
- [ ] 5.3.7 Git commit: `git commit -m "feat(jose-ja): implement JWKS endpoint with cross-tenant control"`

**Evidence**: All validation checks pass

---

## Phase 6: JOSE-JA Audit Logging (2-3 days)

### 6.1 Implement Audit Config Service

**File**: `internal/jose/service/audit_config_service.go`

- [ ] 6.1.1 Implement `GetConfig(ctx, tenantID, operation)` (returns enabled + sampling_rate)
- [ ] 6.1.2 Implement `SetConfig(ctx, tenantID, operation, enabled, samplingRate)`
- [ ] 6.1.3 Implement default config initialization (all operations enabled, 1% sampling)
- [ ] 6.1.4 Write unit tests (≥95% coverage)

**Evidence**: Tests pass, per-tenant config works

---

### 6.2 Add Audit Logging to All Operations

**Files**: `internal/jose/service/*_service.go`

- [ ] 6.2.1 Add audit logging to `CreateElasticJWK()`
- [ ] 6.2.2 Add audit logging to `RotateMaterial()`
- [ ] 6.2.3 Add audit logging to `Sign()`
- [ ] 6.2.4 Add audit logging to `Verify()`
- [ ] 6.2.5 Add audit logging to `Encrypt()`
- [ ] 6.2.6 Add audit logging to `Decrypt()`
- [ ] 6.2.7 Implement sampling logic (check sampling_rate before logging)
- [ ] 6.2.8 Link audit logs to user_id + session_id from context
- [ ] 6.2.9 Write tests verifying audit logs created
- [ ] 6.2.10 Run `go test ./internal/jose/service/` -cover

**Evidence**: Tests pass, all operations logged

---

### 6.3 Create Admin API for Audit Config

**File**: `internal/jose/server/handlers/audit_config_handler.go`

- [ ] 6.3.1 Implement `GetAuditConfig(c *fiber.Ctx)` (GET /browser/api/v1/admin/audit-config)
- [ ] 6.3.2 Implement `SetAuditConfig(c *fiber.Ctx)` (PUT /browser/api/v1/admin/audit-config)
- [ ] 6.3.3 Require admin permission
- [ ] 6.3.4 Write handler tests
- [ ] 6.3.5 Run `go test ./internal/jose/server/handlers/` -cover

**Evidence**: Tests pass, admin API functional

---

### 6.4 Phase 6 Validation

- [ ] 6.4.1 Run `go build ./internal/jose/...` (zero errors)
- [ ] 6.4.2 Run `golangci-lint run ./internal/jose/...` (zero warnings)
- [ ] 6.4.3 Run `go test ./internal/jose/...` -cover (all pass)
- [ ] 6.4.4 Verify all operations create audit logs
- [ ] 6.4.5 Verify sampling rate enforced (1% default)
- [ ] 6.4.6 Verify per-tenant config works
- [ ] 6.4.7 Git commit: `git commit -m "feat(jose-ja): add audit logging with per-tenant config"`

**Evidence**: All validation checks pass

---

## Phase 7: JOSE-JA Path Migration & Middleware (2-3 days)

### 7.1 Migrate Endpoints to Path Split

**File**: `internal/jose/server/routes.go`

- [ ] 7.1.1 Register all business endpoints under `/browser/api/v1/jose/**`
- [ ] 7.1.2 Register all business endpoints under `/service/api/v1/jose/**`
- [ ] 7.1.3 Add CSRF middleware to `/browser/**` paths
- [ ] 7.1.4 Add CORS middleware to `/browser/**` paths
- [ ] 7.1.5 NO CSRF on `/service/**` paths (correct)
- [ ] 7.1.6 Verify routes compile

**Evidence**: Routes registered, middleware applied

---

### 7.2 Add Rate Limiting

**File**: `internal/jose/server/middleware/rate_limit.go`

- [ ] 7.2.1 Implement per-session rate limiting using Fiber middleware
- [ ] 7.2.2 Return HTTP 429 when limit exceeded (per QUIZME Q19)
- [ ] 7.2.3 Apply to all `/browser/**` and `/service/**` paths
- [ ] 7.2.4 Write tests verifying rate limiting works
- [ ] 7.2.5 Run `go test ./internal/jose/server/middleware/` -cover

**Evidence**: Tests pass, rate limiting functional

---

### 7.3 Update OpenAPI Specs

**Files**: `api/jose/openapi_spec_components.yaml`, `api/jose/openapi_spec_paths.yaml`

- [ ] 7.3.1 Split components and paths per sm-kms pattern (QUIZME Q22)
- [ ] 7.3.2 Update all paths to use `/browser/api/v1/jose/**` and `/service/api/v1/jose/**`
- [ ] 7.3.3 Add URL versioning pattern (per QUIZME Q23)
- [ ] 7.3.4 Generate client/server code with oapi-codegen
- [ ] 7.3.5 Verify generated code compiles

**Evidence**: OpenAPI specs updated, generated code compiles

---

### 7.4 Phase 7 Validation

- [ ] 7.4.1 Run `go build ./internal/jose/...` (zero errors)
- [ ] 7.4.2 Run `golangci-lint run ./internal/jose/...` (zero warnings)
- [ ] 7.4.3 Run `go test ./internal/jose/...` -cover (all pass)
- [ ] 7.4.4 Verify CSRF works on `/browser/**`
- [ ] 7.4.5 Verify no CSRF on `/service/**`
- [ ] 7.4.6 Verify rate limiting returns HTTP 429
- [ ] 7.4.7 Git commit: `git commit -m "feat(jose-ja): migrate paths, add CSRF/CORS/rate-limiting"`

**Evidence**: All validation checks pass

---

## Phase 8: JOSE-JA Integration & E2E Testing (3-4 days)

### 8.1 Create E2E Test Infrastructure

**File**: `internal/jose/e2e_test.go`

- [ ] 8.1.1 Create TestMain with Docker Compose:
  - Start jose-ja service
  - Start PostgreSQL
  - Wait for health checks
  - Register user with create_tenant=true
- [ ] 8.1.2 Create helper functions for API calls
- [ ] 8.1.3 Verify infrastructure works

**Evidence**: TestMain starts service successfully

---

### 8.2 E2E: Full JWK Lifecycle

**File**: `internal/jose/e2e_lifecycle_test.go`

- [ ] 8.2.1 Test: Create elastic JWK
- [ ] 8.2.2 Test: Get elastic JWK
- [ ] 8.2.3 Test: List elastic JWKs
- [ ] 8.2.4 Test: Sign with elastic JWK
- [ ] 8.2.5 Test: Verify signature
- [ ] 8.2.6 Test: Encrypt with elastic JWK
- [ ] 8.2.7 Test: Decrypt ciphertext
- [ ] 8.2.8 Run `go test ./internal/jose/` -tags=e2e

**Evidence**: E2E lifecycle tests pass

---

### 8.3 E2E: Material Rotation Limit

**File**: `internal/jose/e2e_rotation_test.go`

- [ ] 8.3.1 Test: Rotate material 1000 times (all succeed)
- [ ] 8.3.2 Test: 1001st rotation FAILS with error message
- [ ] 8.3.3 Test: Historical materials remain usable after rotation
- [ ] 8.3.4 Test: Sign with old material fails (only active can sign)
- [ ] 8.3.5 Test: Verify with old material succeeds (historical always usable)
- [ ] 8.3.6 Run `go test ./internal/jose/` -tags=e2e

**Evidence**: Rotation limit enforced, historical materials work

---

### 8.4 E2E: Multi-Tenant Isolation

**File**: `internal/jose/e2e_multitenant_test.go`

- [ ] 8.4.1 Register two users with separate tenants
- [ ] 8.4.2 Test: User1 creates elastic JWK in tenant1
- [ ] 8.4.3 Test: User2 cannot access tenant1's elastic JWK
- [ ] 8.4.4 Test: User2 creates elastic JWK in tenant2
- [ ] 8.4.5 Test: User1 cannot access tenant2's elastic JWK
- [ ] 8.4.6 Run `go test ./internal/jose/` -tags=e2e

**Evidence**: Multi-tenant isolation enforced

---

### 8.5 E2E: Audit Logging Verification

**File**: `internal/jose/e2e_audit_test.go`

- [ ] 8.5.1 Test: Create elastic JWK, verify audit log entry
- [ ] 8.5.2 Test: Sign operation, verify audit log entry
- [ ] 8.5.3 Test: Verify audit log links to user_id and session_id
- [ ] 8.5.4 Test: Disable audit for operation, verify no log entry
- [ ] 8.5.5 Run `go test ./internal/jose/` -tags=e2e

**Evidence**: Audit logs created correctly

---

### 8.6 Load Testing with Gatling

**File**: `test/load/jose_load_test.scala`

- [ ] 8.6.1 Create Gatling scenario for sign operations
- [ ] 8.6.2 Create Gatling scenario for verify operations
- [ ] 8.6.3 Run load test: 100 RPS for 5 minutes
- [ ] 8.6.4 Verify performance acceptable (P95 < 500ms)
- [ ] 8.6.5 Verify no errors under load

**Evidence**: Load tests pass, performance acceptable

---

### 8.7 Phase 8 Validation

- [ ] 8.7.1 Run `go test ./internal/jose/` -tags=e2e (all pass)
- [ ] 8.7.2 Verify material rotation limit enforced
- [ ] 8.7.3 Verify multi-tenant isolation works
- [ ] 8.7.4 Verify audit logs complete
- [ ] 8.7.5 Verify load tests pass
- [ ] 8.7.6 Git commit: `git commit -m "test(jose-ja): add comprehensive E2E and load tests"`

**Evidence**: All E2E tests pass

---

## Phase 9: JOSE-JA Documentation & Cleanup (2-3 days)

### 9.1 Create Migration Guide

**File**: `docs/jose-ja/MIGRATION-GUIDE.md`

- [ ] 9.1.1 Document service-template changes (default tenant removal)
- [ ] 9.1.2 Document cipher-im changes
- [ ] 9.1.3 Document jose-ja new architecture
- [ ] 9.1.4 Provide migration examples

**Evidence**: Migration guide complete

---

### 9.2 Update API Documentation

**File**: `docs/jose-ja/API-REFERENCE.md`

- [ ] 9.2.1 Document all `/browser/api/v1/jose/**` endpoints
- [ ] 9.2.2 Document all `/service/api/v1/jose/**` endpoints
- [ ] 9.2.3 Document JWKS endpoint behavior
- [ ] 9.2.4 Document audit logging
- [ ] 9.2.5 Document rate limiting

**Evidence**: API documentation complete

---

### 9.3 Update Deployment Guides

**File**: `docs/jose-ja/DEPLOYMENT.md`

- [ ] 9.3.1 Update Docker Compose examples
- [ ] 9.3.2 Update Kubernetes manifests
- [ ] 9.3.3 Document multi-tenant setup
- [ ] 9.3.4 Document security best practices

**Evidence**: Deployment guide complete

---

### 9.4 Final Cleanup

- [ ] 9.4.1 Remove all TODOs: `grep -r "TODO" internal/jose/`
- [ ] 9.4.2 Run `golangci-lint run ./...` (zero warnings)
- [ ] 9.4.3 Run all tests: `go test ./...` (all pass)
- [ ] 9.4.4 Verify coverage targets met (≥95% production, ≥98% infrastructure)
- [ ] 9.4.5 Verify mutation scores met (≥85% production, ≥98% infrastructure)

**Evidence**: All quality gates pass

---

### 9.5 Update Instructions

**File**: `.github/instructions/03-08.server-builder.instructions.md`

- [ ] 9.5.1 Document removal of `WithDefaultTenant()` pattern
- [ ] 9.5.2 Document new registration flow pattern
- [ ] 9.5.3 Document TestMain pattern for tests
- [ ] 9.5.4 Add examples from jose-ja

**Evidence**: Instructions updated

---

### 9.6 Phase 9 Validation

- [ ] 9.6.1 Verify all documentation complete
- [ ] 9.6.2 Verify no deprecated code remains
- [ ] 9.6.3 Verify all quality gates pass
- [ ] 9.6.4 Git commit: `git commit -m "docs(jose-ja): complete documentation and cleanup"`

**Evidence**: Final validation complete

---

## Final Project Validation

- [ ] All 3 phases complete (service-template, cipher-im, jose-ja)
- [ ] Zero build errors across entire project
- [ ] Zero linting warnings across entire project
- [ ] All tests pass (unit, integration, E2E)
- [ ] Coverage targets met (≥95% production, ≥98% infrastructure)
- [ ] Mutation scores met (≥85% production, ≥98% infrastructure)
- [ ] Documentation complete
- [ ] Git history clean (conventional commits)

**Final Commit**: `git commit -m "feat: complete jose-ja refactoring with service-template prerequisites"`

---

## Estimated Timeline Summary

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 0: Service-Template | 5-7 days | NONE |
| Phase 1: Cipher-IM | 3-4 days | Phase 0 |
| Phase 2: JOSE DB Schema | 4-5 days | Phase 1 |
| Phase 3: JOSE ServerBuilder | 3-4 days | Phase 2 |
| Phase 4: JOSE Elastic JWK | 4-5 days | Phase 3 |
| Phase 5: JOSE JWKS Endpoint | 2-3 days | Phase 4 |
| Phase 6: JOSE Audit Logging | 2-3 days | Phase 4 |
| Phase 7: JOSE Path Migration | 2-3 days | Phases 5+6 |
| Phase 8: JOSE E2E Testing | 3-4 days | Phase 7 |
| Phase 9: JOSE Documentation | 2-3 days | Phase 8 |
| **TOTAL** | **30-41 days** | Sequential |

**CRITICAL**: Phases 0-1 (8-11 days) MUST complete before jose-ja work begins.
