# JOSE-JA Refactoring Plan v3 (Includes Service-Template Prerequisites)

**Last Updated**: 2026-01-16
**Based On**:
- JOSE-JA-QUIZME-ROUND2.md (24 questions answered)
- SERVER-BUILDER-QUIZME.md (20 questions answered)
- Service-template investigation findings

**Progress Tracking**:
- Use JOSE-JA-REFACTORING-TASKS-V3.md for task checklists ([ ] / [x])
- Phase 0 Status: ❌ NOT STARTED
- Phase 1 Status: ❌ BLOCKED BY PHASE 0
- Phase 2 Status: ⚠️ Infrastructure complete (commit 9f8fa445), tests pending, blocked by Phase 0/1

## Core Principles - MANDATORY

**Quality Over Speed (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO tasks skipped, NO features deprioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step (build, lint, test, coverage, mutation)
- ✅ **Reliability**: ≥95% coverage production, ≥98% infrastructure/utility, ≥85% mutation production, ≥98% mutation infrastructure
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark tasks complete without objective evidence

**Continuous Execution (NO STOPPING)**:
- Work continues until ALL tasks complete OR user clicks STOP button
- NEVER stop to ask permission between tasks ("Should I continue?")
- NEVER pause for status updates or celebrations ("Here's what we did...")
- NEVER give up when encountering complexity (find solutions, refactor, investigate)
- NEVER skip tasks to "save time" or because they seem "less important"
- Task complete → Commit → IMMEDIATELY start next task (zero pause, zero text to user)

## Executive Summary

This plan covers **THREE sequential phases** required for jose-ja implementation:

1. **Phase 0: Service-Template Refactoring** (BLOCKER) - Remove default tenant pattern
2. **Phase 1: Cipher-IM Migration** (BLOCKER) - Adapt to new service-template pattern
3. **Phase 2-9: JOSE-JA Implementation** - Original jose-ja work

**CRITICAL**: Phases 0-1 MUST complete before Phase 2 begins. Service-template and cipher-im are blocking issues for ALL future services.

---

## Phase 0: Service-Template - Remove Default Tenant Pattern (BLOCKER)

### Problem Statement

**Current State**: Service-template provides `WithDefaultTenant()` method that creates default tenant+realm via `EnsureDefaultTenant()` helper.

**Desired State**: NO default tenant pattern. ALL tenants created via user/client registration with explicit tenant creation OR join existing tenant flow.

**Architectural Requirement**:
- Users register via `/browser/api/v1/auth/register` or `/service/api/v1/auth/register`
- Choose: Create new tenant OR join existing tenant
- If new tenant: User becomes admin, others request to join (requires admin approval)
- If join existing: Requires admin authorization
- Tests use TestMain pattern to start service once per package, tests needing tenant MUST register user/client with tenant creation option

### Scope of Changes

#### Files to Modify:

1. **`internal/apps/template/service/server/builder/server_builder.go`**:
   - ❌ REMOVE: `WithDefaultTenant(tenantID, realmID)` method (lines 103-114)
   - ❌ REMOVE: Call to `ensureDefaultTenant()` (lines 183-189)
   - ❌ REMOVE: `ensureDefaultTenant()` helper method (lines 506+)
   - ❌ REMOVE: `defaultTenantID` and `defaultRealmID` fields from ServerBuilder struct
   - ❌ REMOVE: Passing defaultTenantID/defaultRealmID to SessionManagerService (lines 217-224)

2. **`internal/apps/template/service/server/repository/seeding.go`**:
   - ❌ REMOVE ENTIRE FILE: `EnsureDefaultTenant()` function (90 lines)

3. **`internal/apps/template/service/server/businesslogic/session_manager_service.go`**:
   - ❌ REMOVE: `defaultTenantID` and `defaultRealmID` fields
   - ❌ REMOVE: Single-tenant convenience methods (IssueBrowserSession, ValidateBrowserSession, IssueServiceSession, ValidateServiceSession without tenant/realm params)
   - ✅ KEEP ONLY: Multi-tenant methods (IssueBrowserSessionWithTenant, ValidateBrowserSessionWithTenant, IssueServiceSessionWithTenant, ValidateServiceSessionWithTenant)

4. **`internal/shared/magic/magic_template.go`** (if exists):
   - ❌ REMOVE: `TemplateDefaultTenantID` constant
   - ❌ REMOVE: `TemplateDefaultRealmID` constant

#### New Files to Create:

1. **`internal/apps/template/service/server/businesslogic/tenant_registration.go`**:
   - ✅ CREATE: `RegisterUserWithTenant(ctx, username, password, createTenant bool, existingTenantID *UUID)` - User registration flow
   - ✅ CREATE: `RegisterClientWithTenant(ctx, clientID, clientSecret, createTenant bool, existingTenantID *UUID)` - Service client registration flow
   - ✅ CREATE: `AuthorizeJoinRequest(ctx, adminUserID, joinRequestID, approve bool)` - Admin approval for tenant join requests
   - ✅ CREATE: `ListJoinRequests(ctx, tenantID)` - List pending join requests for tenant admin

2. **`internal/apps/template/service/server/repository/tenant_join_requests.go`**:
   - ✅ CREATE: Database model for join requests (id, user_id/client_id, tenant_id, status, requested_at, processed_at)
   - ✅ CREATE: Repository methods (Create, Update, ListByTenant, GetByID)

3. **Migration**: `internal/apps/template/service/server/repository/migrations/1005_tenant_join_requests.up.sql`:
   ```sql
   CREATE TABLE IF NOT EXISTS tenant_join_requests (
       id TEXT PRIMARY KEY NOT NULL,
       user_id TEXT,  -- NULL for service client requests
       client_id TEXT,  -- NULL for user requests
       tenant_id TEXT NOT NULL,
       status TEXT NOT NULL,  -- pending, approved, rejected
       requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
       processed_at TIMESTAMP,
       processed_by TEXT,  -- admin user_id who approved/rejected
       FOREIGN KEY (tenant_id) REFERENCES tenants(id),
       FOREIGN KEY (user_id) REFERENCES users(id),
       FOREIGN KEY (processed_by) REFERENCES users(id),
       CHECK ((user_id IS NOT NULL AND client_id IS NULL) OR (user_id IS NULL AND client_id IS NOT NULL))
   );

   CREATE INDEX IF NOT EXISTS idx_tenant_join_requests_tenant ON tenant_join_requests(tenant_id);
   CREATE INDEX IF NOT EXISTS idx_tenant_join_requests_status ON tenant_join_requests(status);
   ```

#### API Changes:

**NEW Registration Endpoints** (`/browser/api/v1/auth/` and `/service/api/v1/auth/`):

```
POST /browser/api/v1/auth/register
{
  "username": "alice",
  "password": "...",
  "create_tenant": true,  // OR join_tenant_id: "uuid"
  "tenant_name": "Alice's Org"  // if create_tenant=true
}

Response:
{
  "user_id": "uuid",
  "tenant_id": "uuid",
  "realm_id": "uuid",
  "session_token": "..."
}

POST /service/api/v1/auth/register
{
  "client_id": "service-x",
  "client_secret": "...",
  "create_tenant": true,  // OR join_tenant_id: "uuid"
  "tenant_name": "Service X Tenant"  // if create_tenant=true
}
```

**NEW Admin Endpoints** (`/browser/api/v1/admin/`):

```
GET /browser/api/v1/admin/join-requests?tenant_id=uuid
Response: [
  {
    "id": "uuid",
    "user_id": "uuid",
    "username": "bob",
    "tenant_id": "uuid",
    "status": "pending",
    "requested_at": "2026-01-16T12:00:00Z"
  }
]

POST /browser/api/v1/admin/join-requests/:id/approve
POST /browser/api/v1/admin/join-requests/:id/reject
```

### Testing Strategy

#### Unit Tests:

1. **`tenant_registration_test.go`**:
   - ✅ Test: Register user with `create_tenant=true` → User becomes admin, tenant created
   - ✅ Test: Register user with `join_tenant_id` → Join request created (pending)
   - ✅ Test: Admin approves join request → User added to tenant
   - ✅ Test: Admin rejects join request → User NOT added to tenant
   - ✅ Test: Service client registration flows (same as users)

#### Integration Tests:

1. **`template_service_integration_test.go`**:
   - ✅ Test: Full registration flow (create tenant, second user joins, admin approves)
   - ✅ Test: Session isolation after tenant creation (user1 cannot access user2's tenant)
   - ✅ Test: Admin-only endpoints reject non-admin users

#### Migration Tests:

1. **Existing Tests MUST be Updated**:
   - ❌ All tests using `WithDefaultTenant()` MUST be refactored to use registration flow
   - ✅ Pattern: TestMain registers user with `create_tenant=true`, stores tenant_id/realm_id, subsequent tests use those values

### Validation Criteria

- ✅ `WithDefaultTenant()` method removed from ServerBuilder
- ✅ `EnsureDefaultTenant()` removed from seeding.go
- ✅ SessionManagerService ONLY has multi-tenant methods
- ✅ Registration flow creates tenants correctly
- ✅ Join request flow works (request → admin approval → user added)
- ✅ All template tests pass with new registration pattern
- ✅ Zero linting/build errors
- ✅ Coverage ≥98% (infrastructure code)

### Estimated Duration: 5-7 days

---

## Phase 1: Cipher-IM - Adapt to New Service-Template Pattern (BLOCKER)

### Problem Statement

**Current State**: Cipher-IM uses `builder.WithDefaultTenant(cryptoutilMagic.CipherIMDefaultTenantID, cryptoutilMagic.CipherIMDefaultRealmID)` pattern.

**Desired State**: Cipher-IM uses new registration flow, tests register users with tenant creation.

### Scope of Changes

#### Files to Modify:

1. **`internal/apps/cipher/im/server/server.go`**:
   - ❌ REMOVE: `builder.WithDefaultTenant(...)` call
   - ✅ VERIFY: SessionManager integration uses multi-tenant methods only

2. **`internal/shared/magic/magic_cipher.go`**:
   - ❌ REMOVE: `CipherIMDefaultTenantID` constant
   - ❌ REMOVE: `CipherIMDefaultRealmID` constant

3. **ALL cipher-im tests** (`internal/apps/cipher/im/**/*_test.go`):
   - ❌ REMOVE: Hardcoded default tenant usage
   - ✅ ADD: TestMain pattern with user registration (create_tenant=true)
   - ✅ PATTERN: Store tenant_id/realm_id from registration, use in all test cases

#### Example Test Refactoring:

**BEFORE**:
```go
func TestSendMessage(t *testing.T) {
    // Assumes default tenant exists
    msg := &Message{
        SenderID: "user1",
        Content: "Hello",
    }
    err := repo.Create(ctx, msg)
    require.NoError(t, err)
}
```

**AFTER**:
```go
var (
    testTenantID googleUuid.UUID
    testRealmID googleUuid.UUID
    testUserID googleUuid.UUID
    testSessionToken string
)

func TestMain(m *testing.M) {
    // Start service once per package
    server := startTestServer(t)
    defer server.Shutdown()

    // Register user with create_tenant=true
    resp := registerUser(t, server, "testuser", "password", true, nil)
    testTenantID = resp.TenantID
    testRealmID = resp.RealmID
    testUserID = resp.UserID
    testSessionToken = resp.SessionToken

    // Run tests
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
    err := repo.Create(ctx, msg)
    require.NoError(t, err)
}
```

### Testing Strategy

1. **Unit Tests**: Verify all cipher-im tests pass with registration pattern
2. **Integration Tests**: Full cipher-im E2E with multi-tenant isolation
3. **Migration Tests**: Ensure existing cipher-im functionality unchanged (just different tenant creation mechanism)

### Validation Criteria

- ✅ Cipher-IM builds without errors
- ✅ All cipher-im tests pass with new pattern
- ✅ No default tenant constants remain
- ✅ TestMain pattern used consistently
- ✅ Multi-tenant isolation verified
- ✅ Coverage ≥95% maintained

### Estimated Duration: 3-4 days

---

## Phase 2-9: JOSE-JA Implementation (Original Work)

### Overview

With service-template and cipher-im blockers resolved, proceed with original jose-ja refactoring plan from JOSE-JA-REFACTORING-PLAN-V2.md.

**Key Decisions from Round 2 QUIZME**:

| Decision Area | Choice | Source |
|---------------|--------|--------|
| Material JWK Lifecycle | Max 1000 materials, never deleted, rotation FAILS at 1001st | QUIZME Q1/Q3 |
| Material Retirement | Old materials NEVER retired, always usable for decrypt/verify | QUIZME Q1 write-in |
| Material Rotation | Hybrid time-based + manual override | QUIZME Q5 |
| Audit Retention | Tiered by operation type | QUIZME Q8 |
| Audit Sampling | 1% default per elastic JWK, overridable | QUIZME Q9 |
| Audit Verbosity | Minimal metadata | QUIZME Q10 |
| JWKS Caching | 5 min TTL | QUIZME Q11 |
| Symmetric JWKS | 404 Not Found with security note | QUIZME Q12 |
| Cross-Tenant JWKS | Private by default, overridable per elastic JWK | QUIZME Q13 |
| Tenant UI | MANDATORY browser UI + service API | QUIZME Q14 |
| Tenant Admin | RBAC (service-template MUST support) | QUIZME Q15 |
| Realm Scoping | Tenant-scoped JWKs | QUIZME Q16 |
| Session Timeouts | Browser shorter, service longer | QUIZME Q17 |
| Session Invalidation | Independent of elastic JWK rotation | QUIZME Q18 |
| Rate Limiting | Per-browser/service at Fiber level, HTTP 429 | QUIZME Q19 |
| Barrier Rotation | No re-encryption needed | QUIZME Q20 |
| Barrier Version | Tracked in JWE kid header | QUIZME Q21 |
| OpenAPI | sm-kms pattern (components + paths separate) | QUIZME Q22 |
| API Versioning | URL versioning /api/v2/ | QUIZME Q23 |
| Test Isolation | UUIDv7 per test tenant | QUIZME Q24 |

### Phase 2: Database Schema & Repository (4-5 days)

**Tasks**:
1. Create migrations with multi-tenancy (tenant_id, realm_id)
2. Create elastic_jwk + jwk_materials tables
3. Create tenant_audit_config table
4. Create domain models (ElasticJWK, MaterialJWK, AuditLogEntry)
5. Create repositories (ElasticJWKRepository, MaterialJWKRepository, AuditRepository, AuditConfigRepository)
6. Unit tests (≥98% coverage)

**Database Schema**:

```sql
-- Elastic JWKs (proxy JWKs containing material JWKs)
CREATE TABLE IF NOT EXISTS elastic_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    realm_id TEXT NOT NULL,
    kid TEXT NOT NULL,  -- Elastic JWK ID (UUIDv7)
    kty TEXT NOT NULL,
    alg TEXT NOT NULL,
    use TEXT NOT NULL,
    max_materials INTEGER NOT NULL DEFAULT 1000,  -- Hard limit per QUIZME Q1
    current_material_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id, realm_id) REFERENCES tenant_realms(tenant_id, realm_id),
    UNIQUE(tenant_id, realm_id, kid)
);

CREATE INDEX IF NOT EXISTS idx_elastic_jwks_tenant_realm ON elastic_jwks(tenant_id, realm_id);
CREATE INDEX IF NOT EXISTS idx_elastic_jwks_kid ON elastic_jwks(kid);

-- Material JWKs (actual cryptographic keys)
CREATE TABLE IF NOT EXISTS material_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    elastic_jwk_id TEXT NOT NULL,
    material_kid TEXT NOT NULL,  -- Material JWK KID (UUIDv7)
    private_jwk_jwe TEXT NOT NULL,  -- Barrier-encrypted JWE
    public_jwk_jwe TEXT NOT NULL,   -- Barrier-encrypted JWE (even for public)
    active BOOLEAN NOT NULL DEFAULT FALSE,  -- Only 1 active per elastic JWK
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    retired_at TIMESTAMP,  -- NULL = still active/usable, NOT NULL = historical only
    barrier_version INTEGER NOT NULL,  -- Tracked per QUIZME Q21
    FOREIGN KEY (elastic_jwk_id) REFERENCES elastic_jwks(id),
    UNIQUE(elastic_jwk_id, material_kid)
);

CREATE INDEX IF NOT EXISTS idx_material_jwks_elastic ON material_jwks(elastic_jwk_id);
CREATE INDEX IF NOT EXISTS idx_material_jwks_active ON material_jwks(elastic_jwk_id, active);

-- Audit configuration (per-tenant operation toggles)
CREATE TABLE IF NOT EXISTS tenant_audit_config (
    tenant_id TEXT NOT NULL,
    operation TEXT NOT NULL,  -- generate, get, list, delete, sign, verify, encrypt, decrypt
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sampling_rate REAL NOT NULL DEFAULT 0.01,  -- 1% default per QUIZME Q9
    PRIMARY KEY (tenant_id, operation),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

-- Audit logs (minimal metadata per QUIZME Q10)
CREATE TABLE IF NOT EXISTS jwk_audit_log (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    realm_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    elastic_kid TEXT,  -- Elastic JWK kid
    material_kid TEXT,  -- Material JWK kid
    user_id TEXT,
    session_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT,  -- JSON, minimal per QUIZME Q10
    FOREIGN KEY (tenant_id, realm_id) REFERENCES tenant_realms(tenant_id, realm_id),
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE INDEX IF NOT EXISTS idx_jwk_audit_log_tenant_realm ON jwk_audit_log(tenant_id, realm_id);
CREATE INDEX IF NOT EXISTS idx_jwk_audit_log_timestamp ON jwk_audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_jwk_audit_log_elastic_kid ON jwk_audit_log(elastic_kid);
```

**Validation**:
- ✅ Migrations apply (PostgreSQL + SQLite)
- ✅ Multi-tenant isolation enforced
- ✅ Material count limit enforced (1000 per elastic JWK)
- ✅ Repository tests pass (≥98% coverage)

### Phase 3: ServerBuilder Integration (3-4 days)

**Tasks**:
1. Create JoseSettings (extends ServiceTemplateServerSettings)
2. Refactor server.go with builder
3. Register domain migrations via `builder.WithDomainMigrations()`
4. Register public routes via `builder.WithPublicRouteRegistration()`
5. NO DEFAULT TENANT - use registration flow like cipher-im
6. Update tests to use TestMain + registration pattern

**Builder Usage**:

```go
func NewServer(ctx context.Context, cfg *config.JoseSettings) (*Server, error) {
    builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

    // Register jose-ja domain migrations (elastic_jwks, material_jwks, audit tables)
    builder.WithDomainMigrations(repository.MigrationsFS, "migrations")

    // Register public routes
    builder.WithPublicRouteRegistration(func(
        base *cryptoutilTemplateServer.PublicServerBase,
        res *cryptoutilTemplateBuilder.ServiceResources,
    ) error {
        return registerJoseRoutes(base, res)
    })

    // Build complete infrastructure (NO WithDefaultTenant call)
    resources, err := builder.Build()
    if err != nil {
        return nil, err
    }

    return &Server{resources: resources}, nil
}
```

**Validation**:
- ✅ Builder creates complete infrastructure
- ✅ SessionManager integrated
- ✅ Both `/browser/**` and `/service/**` paths functional
- ✅ All tests pass with registration pattern

### Phase 4: Elastic JWK Implementation (4-5 days)

**Tasks**:
1. Implement elastic JWK service
2. Implement material JWK generation
3. **CRITICAL**: Enforce 1000 material limit (rotation FAILS at 1001st per QUIZME Q1/Q3)
4. Implement material rotation (hybrid time-based + manual per QUIZME Q5)
5. Sign/encrypt uses active material
6. Verify/decrypt looks up by embedded material_kid
7. Historical materials ALWAYS usable (NEVER deleted per QUIZME Q1)
8. Tests for key rotation and limit enforcement

**Material Rotation Behavior**:

```go
func (s *ElasticJWKService) RotateMaterial(ctx context.Context, elasticKID googleUuid.UUID) error {
    // Check current material count
    count, err := s.repo.CountMaterials(ctx, elasticKID)
    if err != nil {
        return err
    }

    if count >= 1000 {
        // CRITICAL: FAIL rotation at limit per QUIZME Q1
        return fmt.Errorf("elastic JWK %s at max 1000 materials, rotation blocked", elasticKID)
    }

    // Generate new material JWK
    newMaterial := generateMaterialJWK(...)

    // Set as active, retire old active (but keep it for decrypt/verify)
    err = s.repo.RotateMaterial(ctx, elasticKID, newMaterial)
    return err
}
```

**Validation**:
- ✅ Elastic JWK pattern works
- ✅ Material count limit enforced (fails at 1001)
- ✅ Material rotation works (hybrid time-based + manual)
- ✅ Sign/verify with historical materials works
- ✅ Old materials NEVER deleted

### Phase 5: JWKS Endpoint Implementation (2-3 days)

**Tasks**:
1. Implement per-elastic-JWK JWKS endpoint: `/service/api/v1/jose/{kid}/.well-known/jwks.json`
2. Return public material JWKs only (filter by `use=sig` or `use=enc`)
3. Symmetric JWKs: Return 404 with security note (per QUIZME Q12)
4. Cross-tenant access: Private by default, overridable (per QUIZME Q13)
5. Caching: 5 min TTL (per QUIZME Q11)
6. Tests for JWKS filtering and caching

**JWKS Endpoint Behavior**:

```go
// GET /service/api/v1/jose/{kid}/.well-known/jwks.json
func (h *Handler) GetJWKS(c *fiber.Ctx) error {
    elasticKID := c.Params("kid")
    tenantID := c.Locals("tenant_id").(googleUuid.UUID)

    // Get elastic JWK with tenant isolation
    elasticJWK, err := h.service.GetElasticJWK(ctx, tenantID, elasticKID)
    if err != nil {
        return err
    }

    // Symmetric keys: 404 per QUIZME Q12
    if elasticJWK.Use == "enc" && isSymmetric(elasticJWK.Kty) {
        return fiber.NewError(404, "Symmetric JWKs not exposed via JWKS endpoint")
    }

    // Get all material JWKs for this elastic JWK
    materials, err := h.service.ListMaterials(ctx, elasticKID)
    if err != nil {
        return err
    }

    // Build JWKS response (public keys only)
    jwks := buildJWKSResponse(materials)

    // Cache for 5 min per QUIZME Q11
    c.Set("Cache-Control", "public, max-age=300")

    return c.JSON(jwks)
}
```

**Validation**:
- ✅ Per-elastic-JWK JWKS works
- ✅ Symmetric keys return 404
- ✅ Caching headers correct (5 min)
- ✅ Cross-tenant access controlled

### Phase 6: Audit Logging (2-3 days)

**Tasks**:
1. Implement audit config service (per-tenant operation toggle)
2. Add audit logging to ALL JWK operations (generate, sign, verify, encrypt, decrypt)
3. Link audit logs to user_id + session_id
4. Implement sampling (1% default per QUIZME Q9)
5. Minimal metadata (per QUIZME Q10)
6. Admin API for audit config management
7. Tests for audit logging

**Audit Logging Pattern**:

```go
func (s *ElasticJWKService) Sign(ctx context.Context, tenantID, elasticKID googleUuid.UUID, payload []byte) (string, error) {
    // Get audit config for this tenant
    auditCfg, _ := s.auditConfig.Get(ctx, tenantID, "sign")

    // Check if auditing enabled
    if auditCfg.Enabled {
        // Check sampling rate (1% default)
        if rand.Float64() < auditCfg.SamplingRate {
            // Log audit entry
            sessionID := ctx.Value("session_id").(googleUuid.UUID)
            userID := ctx.Value("user_id").(googleUuid.UUID)

            s.audit.Log(ctx, AuditLogEntry{
                TenantID: tenantID,
                Operation: "sign",
                ElasticKID: elasticKID,
                UserID: userID,
                SessionID: sessionID,
                Metadata: "{}", // Minimal per QUIZME Q10
            })
        }
    }

    // Perform sign operation...
}
```

**Validation**:
- ✅ All operations logged (when enabled)
- ✅ Per-tenant config works
- ✅ Sampling rate enforced
- ✅ Minimal metadata stored

### Phase 7: Path Migration & Middleware (2-3 days)

**Tasks**:
1. Migrate all endpoints to `/browser/**` and `/service/**`
2. Add CSRF middleware to `/browser/**` paths
3. Add CORS middleware to `/browser/**` paths
4. Rate limiting: Per-browser/service at Fiber level, HTTP 429 (per QUIZME Q19)
5. Update OpenAPI specs for new paths (sm-kms pattern per QUIZME Q22)
6. Tests for middleware behavior

**Route Registration**:

```go
func registerJoseRoutes(base *cryptoutilTemplateServer.PublicServerBase, res *cryptoutilTemplateBuilder.ServiceResources) error {
    app := base.App()

    // Session middleware (from service-template)
    browserSession := middleware.BrowserSessionMiddleware(res.SessionManager)
    serviceSession := middleware.ServiceSessionMiddleware(res.SessionManager)

    // Browser paths (with CSRF, CORS)
    browserGroup := app.Group("/browser/api/v1/jose")
    browserGroup.Post("/elastic-jwks", browserSession, handler.CreateElasticJWK)
    browserGroup.Get("/elastic-jwks/:kid", browserSession, handler.GetElasticJWK)
    browserGroup.Post("/elastic-jwks/:kid/sign", browserSession, handler.Sign)
    browserGroup.Post("/elastic-jwks/:kid/verify", browserSession, handler.Verify)

    // Service paths (no CSRF)
    serviceGroup := app.Group("/service/api/v1/jose")
    serviceGroup.Post("/elastic-jwks", serviceSession, handler.CreateElasticJWK)
    serviceGroup.Get("/elastic-jwks/:kid", serviceSession, handler.GetElasticJWK)
    serviceGroup.Post("/elastic-jwks/:kid/sign", serviceSession, handler.Sign)
    serviceGroup.Post("/elastic-jwks/:kid/verify", serviceSession, handler.Verify)

    // JWKS endpoint (service only)
    serviceGroup.Get("/elastic-jwks/:kid/.well-known/jwks.json", handler.GetJWKS)

    return nil
}
```

**Validation**:
- ✅ All endpoints migrated
- ✅ CSRF protection works on `/browser/**`
- ✅ `/service/**` has no CSRF (correct)
- ✅ Rate limiting works (HTTP 429)
- ✅ OpenAPI specs updated

### Phase 8: Integration & E2E Testing (3-4 days)

**Tasks**:
1. E2E: Full JWK lifecycle (multi-tenant)
2. E2E: Elastic JWK rotation (verify 1000 limit)
3. E2E: Multi-instance deployment
4. E2E: Audit log verification
5. E2E: SessionManager integration
6. E2E: Historical material usage (sign with old material fails, verify with old material succeeds)
7. Load testing (Gatling)

**E2E Test Pattern**:

```go
func TestMain(m *testing.M) {
    // Start jose-ja service with Docker Compose
    composeCmd := exec.Command("docker", "compose", "up", "-d")
    composeCmd.Run()

    // Wait for health check
    waitForHealth("https://localhost:9443/admin/v1/readyz")

    // Register user with create_tenant=true
    resp := registerUser("alice", "password", true, nil)
    testTenantID = resp.TenantID
    testRealmID = resp.RealmID
    testSessionToken = resp.SessionToken

    // Run tests
    exitCode := m.Run()

    // Cleanup
    exec.Command("docker", "compose", "down").Run()
    os.Exit(exitCode)
}

func TestE2E_MaterialRotationLimit(t *testing.T) {
    // Create elastic JWK
    elasticKID := createElasticJWK(t, testTenantID, testRealmID)

    // Rotate 1000 times (should succeed)
    for i := 0; i < 1000; i++ {
        err := rotateMaterial(t, elasticKID)
        require.NoError(t, err)
    }

    // 1001st rotation should FAIL per QUIZME Q1
    err := rotateMaterial(t, elasticKID)
    require.Error(t, err)
    require.Contains(t, err.Error(), "max 1000 materials")
}

func TestE2E_HistoricalMaterialUsable(t *testing.T) {
    // Create elastic JWK
    elasticKID := createElasticJWK(t, testTenantID, testRealmID)

    // Sign with active material
    sig1, material1KID := signPayload(t, elasticKID, "hello")

    // Rotate material (material1 becomes historical)
    rotateMaterial(t, elasticKID)

    // Sign with NEW active material should succeed
    sig2, material2KID := signPayload(t, elasticKID, "world")
    require.NotEqual(t, material1KID, material2KID)

    // Verify with OLD material1 should SUCCEED (historical materials NEVER retire per QUIZME Q1)
    err := verifySignature(t, elasticKID, sig1, material1KID)
    require.NoError(t, err)

    // Verify with NEW material2 should SUCCEED
    err = verifySignature(t, elasticKID, sig2, material2KID)
    require.NoError(t, err)
}
```

**Validation**:
- ✅ All E2E tests pass
- ✅ Multi-instance works
- ✅ 1000 material limit enforced
- ✅ Historical materials usable forever
- ✅ Performance acceptable

### Phase 9: Documentation & Cleanup (2-3 days)

**Tasks**:
1. Create migration guide
2. Update API documentation (OpenAPI specs)
3. Update deployment guides
4. Final cleanup (linting, TODOs)
5. Update 03-08.server-builder.instructions.md with removal of default tenant pattern

**Validation**:
- ✅ Documentation complete
- ✅ No deprecated code
- ✅ All quality gates pass

---

## Timeline Estimate (COMPLETE)

| Phase | Duration | Risk | Dependencies |
|-------|----------|------|--------------|
| **Phase 0: Service-Template** | 5-7 days | High | NONE |
| **Phase 1: Cipher-IM** | 3-4 days | Medium | Phase 0 |
| **Phase 2: JOSE DB Schema** | 4-5 days | Low | Phase 1 |
| **Phase 3: JOSE ServerBuilder** | 3-4 days | Low | Phase 2 |
| **Phase 4: JOSE Elastic JWK** | 4-5 days | Medium | Phase 3 |
| **Phase 5: JOSE JWKS Endpoint** | 2-3 days | Low | Phase 4 |
| **Phase 6: JOSE Audit Logging** | 2-3 days | Low | Phase 4 |
| **Phase 7: JOSE Path Migration** | 2-3 days | Low | Phases 5+6 |
| **Phase 8: JOSE E2E Testing** | 3-4 days | High | Phase 7 |
| **Phase 9: JOSE Documentation** | 2-3 days | Low | Phase 8 |
| **TOTAL** | **30-41 days** | Medium | Sequential |

**CRITICAL**: Phases 0-1 are BLOCKING for jose-ja. Estimated 8-11 days before jose-ja work begins.

---

## Risk Assessment

### CRITICAL Risks

1. **Service-Template Default Tenant Removal**: Breaks cipher-im and potentially other services
   - **Mitigation**: Comprehensive testing, TestMain pattern validation, gradual rollout

2. **Multi-Tenant Data Isolation**: Cross-tenant data leakage would be security incident
   - **Mitigation**: Row-level security tests, tenant_id enforcement in all queries

3. **Material Rotation Limit Enforcement**: Must fail gracefully at 1000th material
   - **Mitigation**: Comprehensive rotation tests, graceful error handling

### High Risks

1. **Barrier Encryption**: Private key encryption/decryption correctness
2. **Multi-Instance Coordination**: Database locking, race conditions
3. **Cipher-IM Migration**: Test refactoring may introduce regressions

### Medium Risks

1. **Path Migration**: Existing API contracts change (but alpha project = acceptable)
2. **Performance**: Database queries + barrier encryption overhead
3. **Audit Log Volume**: High-frequency operations (sign/verify) generate large logs

---

## Success Criteria

### Service-Template (Phase 0)

- ✅ `WithDefaultTenant()` removed from ServerBuilder
- ✅ `EnsureDefaultTenant()` removed from seeding.go
- ✅ Registration flow creates tenants correctly
- ✅ Join request flow works (request → admin approval → user added)
- ✅ All template tests pass with new pattern
- ✅ Coverage ≥98% (infrastructure)

### Cipher-IM (Phase 1)

- ✅ Cipher-IM builds without errors
- ✅ All cipher-im tests pass with new pattern
- ✅ No default tenant constants remain
- ✅ TestMain pattern used consistently
- ✅ Coverage ≥95% maintained

### JOSE-JA (Phases 2-9)

- ✅ Multi-tenancy enforced (tenant + realm isolation)
- ✅ Elastic JWK pattern works (rotation + historical material)
- ✅ Material limit enforced (fails at 1001)
- ✅ Historical materials ALWAYS usable
- ✅ SessionManager integrated (browser + service paths)
- ✅ Audit logging complete (all operations, per-tenant config)
- ✅ Path split implemented (`/browser/**` + `/service/**`)
- ✅ Per-JWK JWKS endpoints functional
- ✅ Zero linting/build errors
- ✅ ≥95% coverage (production), ≥98% (infrastructure)
- ✅ ≥85% mutation score (production), ≥98% (infrastructure)

---

## Cross-References

- **Quiz Answers**: [JOSE-JA-QUIZME-ROUND2.md](JOSE-JA-QUIZME-ROUND2.md), [SERVER-BUILDER-QUIZME.md](SERVER-BUILDER-QUIZME.md)
- **Service-Template**: [03-08.server-builder.instructions.md](../../.github/instructions/03-08.server-builder.instructions.md)
- **cipher-im Reference**: [internal/apps/cipher/im/](../../internal/apps/cipher/im/)
- **Multi-Tenancy**: [02-10.authn.instructions.md](../../.github/instructions/02-10.authn.instructions.md)
