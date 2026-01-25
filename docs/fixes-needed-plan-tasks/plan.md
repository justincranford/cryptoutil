# JOSE-JA Refactoring Plan V4

**Last Updated**: 2026-01-18

**Architecture Reference**: See [docs/arch/ARCHITECTURE.md](../arch/ARCHITECTURE.md) for comprehensive design patterns, principles, and implementation guidelines.

**Tasks Reference**: See [tasks.md](tasks.md) for detailed task breakdown with checkboxes.

**Progress Tracking - MANDATORY**: As tasks are completed, check them off in tasks.md to track progress. Each checkbox represents objective evidence of completion (build succeeds, tests pass, coverage met, mutation score met, commit created).

**Critical Fixes from V3**:

- ✅ Port 9090 for admin endpoints (was incorrectly 8080)
- ✅ PostgreSQL 18+ requirement (was incorrectly 16+)
- ✅ Directory structure: deployments/jose-ja/, configs/jose-ja/ (was jose/)
- ✅ Removed CGO_ENABLED mentions (implied by project)
- ✅ Docker secrets > YAML config > ENV (was promoting ENVs)
- ✅ Separate browser vs service session configs
- ✅ OTLP only (removed Prometheus scraping endpoint)
- ✅ Consistent API paths: /admin/api/v1, /service/api/v1, /browser/api/v1
- ✅ No service name in paths (was /service/api/v1/jose/*)
- ✅ Realms are authn only (removed from repository WHERE clauses)
- ✅ No hardcoded passwords in tests (use magic constants or UUIDv7)
- ✅ key_type implied by algorithm (simplified API)
- ✅ All requirements mandatory (no "Future" deferrals)

---

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

---

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

- Users register via `/browser/api/v1/register` or `/service/api/v1/register` (NOT /auth/register)
- Registration endpoints are unauthenticated (rate-limited, template infrastructure)
- User saved in pending_users table (NOT users table) until approved
- Registration parameter: `tenant_id` (absence = create new, presence = request join)
- If new tenant: User becomes admin upon approval, tenant created during approval
- If join existing: Creates join request, requires admin authorization via `/admin/api/v1/join-requests/:id`
- HTTP 403 Forbidden returned for ALL authn endpoints until user approved
- HTTP 401 Unauthorized returned if user rejected
- NO session_token issued until user approved (removed from registration response)
- Tests use TestMain pattern to start service once per package, tests needing tenant MUST register user via HTTP endpoint

---

### 0.1 Remove WithDefaultTenant from ServerBuilder

**File**: `internal/apps/template/service/server/builder/server_builder.go`

**Changes**:

- ❌ REMOVE: `WithDefaultTenant(tenantID, realmID)` method
- ❌ REMOVE: Call to `ensureDefaultTenant()` in `Build()` method
- ❌ REMOVE: `ensureDefaultTenant()` helper method
- ❌ REMOVE: `defaultTenantID` and `defaultRealmID` fields from ServerBuilder struct
- ❌ REMOVE: Passing defaultTenantID/defaultRealmID to SessionManagerService constructor

**Quality Gates**:

1. Build: `go build ./internal/apps/template/...`
2. Linting: `golangci-lint run --fix ./internal/apps/template/...`
3. Tests: ALL template tests updated to use registration pattern
4. Evidence: Commit hash with conventional message

---

### 0.2 Remove EnsureDefaultTenant Helper

**File**: `internal/apps/template/service/server/repository/seeding.go`

**Changes**:

- ❌ DELETE entire file (contains only EnsureDefaultTenant function)

**Quality Gates**:

1. Build: `go build ./internal/apps/template/...`
2. Linting: `golangci-lint run --fix`
3. Evidence: File deleted, no references remain

---

### 0.3 Update SessionManagerService (Remove Single-Tenant Methods)

**File**: `internal/apps/template/service/server/businesslogic/session_manager_service.go`

**Changes**:

- ❌ REMOVE: `defaultTenantID` field
- ❌ REMOVE: `defaultRealmID` field
- ❌ REMOVE: `IssueBrowserSession(ctx, userID)` method (single-tenant version)
- ❌ REMOVE: `ValidateBrowserSession(ctx, token)` method (single-tenant version)
- ❌ REMOVE: `IssueServiceSession(ctx, clientID)` method (single-tenant version)
- ❌ REMOVE: `ValidateServiceSession(ctx, token)` method (single-tenant version)
- ✅ KEEP: `IssueBrowserSessionWithTenant(ctx, userID, tenantID, realmID)` (multi-tenant version)
- ✅ KEEP: `ValidateBrowserSessionWithTenant(ctx, token)` (multi-tenant version)
- ✅ KEEP: `IssueServiceSessionWithTenant(ctx, clientID, tenantID, realmID)` (multi-tenant version)
- ✅ KEEP: `ValidateServiceSessionWithTenant(ctx, token)` (multi-tenant version)
- ✅ UPDATE: Constructor `NewSessionManagerService()` to remove defaultTenantID/defaultRealmID params

**Quality Gates**:

1. Build: `go build ./internal/apps/template/...`
2. Linting: `golangci-lint run --fix`
3. Tests: All session manager tests updated
4. Evidence: Commit hash

---

### 0.4 Remove Template Magic Constants

**Files**: `internal/shared/magic/magic_template.go`, `internal/apps/template/service/server/repository/seeding_test.go` (if exists)

**Changes**:

- ❌ REMOVE: `TemplateDefaultTenantID` constant (if exists)
- ❌ REMOVE: `TemplateDefaultRealmID` constant (if exists)
- ✅ VERIFY: `grep -r "TemplateDefaultTenantID" internal/` returns 0 results
- ✅ VERIFY: `grep -r "TemplateDefaultRealmID" internal/` returns 0 results

**Quality Gates**:

1. Build: `go build ./...`
2. Grep verification: 0 usages of removed constants
3. Evidence: Grep output + commit hash

---

### 0.5 Create pending_users Table Migration (NEW)

**Files**:

- `internal/apps/template/service/server/repository/migrations/1005_pending_users.up.sql`
- `internal/apps/template/service/server/repository/migrations/1005_pending_users.down.sql`

**Rationale**:

- Users NOT saved in users table until approved
- Saved in pending_users table during registration
- Moved to users table only upon admin approval
- HTTP 403 for all authn endpoints until approved
- HTTP 401 if rejected

**Schema Requirements**:

- **Q1.1 Uniqueness**: `username` unique per tenant across BOTH `pending_users` AND `users` tables (prevents duplicate registrations while pending)
- **Q1.2 Indexes**: Composite index `(username, tenant_id)`, status + requested_at index for cleanup queries
- **Q1.4 Expiration**: Configurable expiration in HOURS (not days), default 72 hours, auto-delete expired entries

**Schema**:

```sql
-- 1005_pending_users.up.sql
CREATE TABLE IF NOT EXISTS pending_users (
    id TEXT PRIMARY KEY NOT NULL,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    tenant_id TEXT,         -- NULL if creating new tenant
    requested_tenant_name TEXT,  -- For new tenant creation
    status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected')),
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    processed_by TEXT,
    rejection_reason TEXT,
    expires_at TIMESTAMP,   -- Q1.4: Expiration in hours (calculated from requested_at + config)
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (processed_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Q1.2: Composite unique index (username, tenant_id) prevents duplicates per tenant
CREATE UNIQUE INDEX IF NOT EXISTS idx_pending_users_username_tenant ON pending_users(username, tenant_id);

-- Q1.2: Status + requested_at for cleanup queries (find expired entries)
CREATE INDEX IF NOT EXISTS idx_pending_users_status_requested ON pending_users(status, requested_at);

-- Q1.4: Expiration cleanup index
CREATE INDEX IF NOT EXISTS idx_pending_users_expires ON pending_users(expires_at) WHERE expires_at IS NOT NULL;

-- 1005_pending_users.down.sql
DROP TABLE IF EXISTS pending_users;
```

**Email Validation**:

- ❌ NO email validation on username field (username can be non-email)
- ✅ Email/password authentication is a DIFFERENT realm (not implemented yet)
- ✅ Username field accepts any string (simplified registration flow)

**Expiration Configuration**:

```yaml
# configs/template/template-server.yaml
pending_users_expiration_hours: 72  # Default 72 hours (3 days)
```

**Quality Gates**:

1. Migration applies to PostgreSQL test container
2. Migration applies to SQLite in-memory
3. Migration rollback works (down migration)
4. Tests verify schema constraints (CHECK, FOREIGN KEY, UNIQUE)
5. Tests verify expiration cleanup (auto-delete expired entries)
6. Evidence: Test output + commit hash

---

### 0.6 REMOVED - pending_users table is sufficient (per Q5.1)

**User Decision**: "WTF is tenant_join_requests (1006)? Only pending_users (1005) needed?"

---

### 0.8 Create Registration HTTP Handlers

**Files**:

- `internal/apps/template/service/server/apis/registration_handlers.go`
- `internal/apps/template/service/server/apis/join_request_handlers.go`

**Admin Dashboard**:

- ✅ **Q2.1**: Template infrastructure provides admin dashboard APIs (NOT domain-specific)
- ❌ **Q2.2**: NO email notifications (users poll status via login attempts)
- ❌ **Q2.3**: NO webhook callbacks (keep template simple)
- ❌ **Q2.4**: NO unauthenticated status API (users poll via login: HTTP 403=pending, HTTP 401=rejected, HTTP 200=approved)

**Rate Limiting**:

- ✅ **Q3.1**: Per IP address only (10 registrations per IP per hour)
- ✅ **Q3.2**: In-memory (sync.Map) - simple, single-node, lost on restart
- ✅ **Q3.3**: Configurable thresholds with low defaults

**Rate Limiting Configuration**:

```yaml
# configs/template/template-server.yaml
registration_rate_limit_per_hour: 10  # Low default per IP
```

**Endpoints**:

#### Browser Registration

```
POST /browser/api/v1/auth/register
Content-Type: application/json

{
    "username": "user@example.com",
    "password": "securepassword",
    "tenant_id": "optional-uuid"  // Absence = create, presence = join
}

Response 201 (Create):
{
    "user_id": "uuid",
    "tenant_id": "uuid",
    "realm_id": "uuid",
    "session_token": "encrypted-jwt"
}

Response 202 (Join):
{
    "join_request_id": "uuid",
    "tenant_id": "uuid",
    "status": "pending",
    "message": "Join request submitted. Waiting for admin approval."
}
```

#### Service Registration

```
POST /service/api/v1/auth/register
Content-Type: application/json
Authorization: Basic <client_id>:<client_secret>

{
    "tenant_id": "optional-uuid"  // Absence = create, presence = join
}

Response 201 (Create):
{
    "client_id": "uuid",
    "tenant_id": "uuid",
    "realm_id": "uuid",
    "access_token": "bearer-token"
}

Response 202 (Join):
{
    "join_request_id": "uuid",
    "tenant_id": "uuid",
    "status": "pending"
}
```

#### Admin: List Join Requests

```
GET /admin/api/v1/join-requests?tenant_id=uuid&status=pending&page=1&size=20
Authorization: Bearer <admin-token>

Response 200:
{
    "items": [
        {
            "id": "uuid",
            "user_id": "uuid",
            "tenant_id": "uuid",
            "status": "pending",
            "requested_at": "2026-01-18T10:00:00Z"
        }
    ],
    "pagination": {
        "page": 1,
        "size": 20,
        "total": 5
    }
}
```

#### Admin: Authorize Join Request

```
PUT /admin/api/v1/join-requests/:id
Authorization: Bearer <admin-token>
Content-Type: application/json

{
    "action": "approve"  // or "reject"
}

Response 200:
{
    "id": "uuid",
    "status": "approved",
    "processed_at": "2026-01-18T10:05:00Z",
    "processed_by": "admin-user-uuid"
}
```

**CRITICAL: API Path Consistency**:

- ✅ Admin endpoints: `/admin/api/v1/*`
- ✅ Service endpoints: `/service/api/v1/*`
- ✅ Browser endpoints: `/browser/api/v1/*`
- ❌ NEVER use `/admin/v1/*` (inconsistent with /service/api/v1 and /browser/api/v1)
- ❌ NEVER include service name in paths (e.g., `/service/api/v1/jose/*` is WRONG)

**Quality Gates**:

1. Build: `go build ./internal/apps/template/...`
2. Tests: Integration tests for ALL endpoints
3. Coverage: ≥95% (handler code)
4. Security: CSRF protection on browser endpoints
5. Evidence: Test output + commit hash

---

### 0.9 Update ServerBuilder Registration

**File**: `internal/apps/template/service/server/builder/server_builder.go`

**Changes**:

- ✅ ADD: Registration handler routes in `Build()` method
- ✅ ADD: Join request handler routes in `Build()` method
- ✅ VERIFY: No `WithDefaultTenant()` calls remain

**Routes**:

```go
// Browser registration
publicServer.POST("/browser/api/v1/auth/register", registrationHandlers.RegisterBrowser)

// Service registration
publicServer.POST("/service/api/v1/auth/register", registrationHandlers.RegisterService)

// Admin join requests
adminServer.GET("/admin/api/v1/join-requests", joinRequestHandlers.List)
adminServer.PUT("/admin/api/v1/join-requests/:id", joinRequestHandlers.Authorize)
```

**Quality Gates**:

1. Build: `go build ./internal/apps/template/...`
2. E2E Tests: Full registration flow works
3. Evidence: E2E test output + commit hash

---

### 0.10 Phase 0 Validation

**Validation Checklist**:

- [ ] Build: `go build ./...` (zero errors)
- [ ] Linting: `golangci-lint run ./...` (zero warnings)
- [ ] Tests: `go test ./internal/apps/template/... -cover` (100% pass)
- [ ] Coverage: ≥95% production, ≥98% infrastructure
- [ ] Mutation: ≥85% production, ≥98% infrastructure
- [ ] E2E: Registration flow works (browser + service)
- [ ] E2E: Join request flow works (create, list, approve)
- [ ] Security: NO hardcoded passwords in tests
- [ ] Git: Conventional commits with evidence

**Final Commit**: `git commit -m "feat(service-template): remove default tenant pattern, implement registration flow"`

---

## Phase 1: Cipher-IM - Adapt to New Service-Template Pattern (BLOCKER)

### Problem Statement

**Current State**: cipher-im service adapted to service-template v3, may still have references to default tenant pattern.

**Desired State**: cipher-im fully compatible with Phase 0 changes, all tests use registration flow.

---

### 1.1 Remove cipher-im Default Tenant References

**Files**:

- `internal/apps/cipher/im/server/server.go`
- `internal/apps/cipher/im/server/server_test.go`

**Changes**:

- ❌ REMOVE: Any `WithDefaultTenant()` calls (if exist)
- ✅ VERIFY: `grep -r "WithDefaultTenant" internal/apps/cipher/` returns 0 results
- ✅ VERIFY: All tests use `registerUser()` or `registerClient()` for tenant creation

**Quality Gates**:

1. Build: `go build ./internal/apps/cipher/...`
2. Tests: `go test ./internal/apps/cipher/... -cover`
3. Grep verification: 0 WithDefaultTenant usages
4. Evidence: Grep output + test output + commit hash

---

### 1.2 Update cipher-im Tests to Registration Pattern

**Files**:

- `internal/apps/cipher/im/server/apis/*_test.go`
- `internal/apps/cipher/im/repository/*_test.go`

**Hash Service Configuration**:

- **Q4.1**: PBKDF2 iterations = 610,000 (ALREADY IMPLEMENTED in hash service - verify in magic constants)
- **Q4.2**: Lazy migration for pepper rotation (hash service ALREADY IMPLEMENTS this pattern)
- **Q4.3**: Multiple hash algorithm versions supported (hash service ALREADY IMPLEMENTS version prefix pattern)
- **Q4.4**: Global security policy (NOT per-tenant - consistent across all tenants)

**Implementation Notes**:

- Hash service in `internal/shared/crypto/hash/` already implements 4 registries:
  - LowEntropyDeterministicHashRegistry (PII with PBKDF2)
  - LowEntropyRandomHashRegistry (Passwords with PBKDF2)
  - HighEntropyDeterministicHashRegistry (Config blobs with HKDF)
  - HighEntropyRandomHashRegistry (API keys with HKDF)
- Verify current iteration count in `internal/shared/magic/magic_cryptography.go`
- Verify pepper rotation pattern in hash service tests

**Pattern**:

```go
func TestMain(m *testing.M) {
    ctx := context.Background()

    // Start server once per package
    testServer, _ = server.NewFromConfig(ctx, config.NewTestSettings())
    go testServer.Start()
    defer testServer.Shutdown(ctx)

    // Wait for ready
    testServer.WaitForReady(ctx, 10*time.Second)

    // Register test tenant through API (CRITICAL: Use magic constant for password)
    resp := registerUser(testServer.PublicBaseURL(), "testuser", cryptoutilMagic.TestPassword, true, nil)
    testTenantID = resp.TenantID
    testRealmID = resp.RealmID
    testSessionToken = resp.SessionToken

    // Run tests
    os.Exit(m.Run())
}
```

**CRITICAL: Password Handling**:

- ❌ NEVER: `registerUser(server, "user1", "pass1", ...)`
- ✅ ALWAYS: `registerUser(server, "user1", cryptoutilMagic.TestPassword, ...)`
- ✅ ALTERNATIVE: `registerUser(server, "user1", googleUuid.NewV7().String(), ...)`

**Migration Strategy**:

- **Q5.1**: ONLY pending_users (1005) needed - NO tenant_join_requests (1006) table
- **Q5.2**: DOWN migrations implemented for dev/test rollback (production forward-only)

**Quality Gates**:

1. Tests: `go test ./internal/apps/cipher/... -cover` (100% pass)
2. Coverage: Maintained or improved
3. Security: NO hardcoded passwords
4. Hash Service: Verify 610,000 iterations in magic constants
5. Evidence: Test output + commit hash

---

### 1.3 Phase 1 Validation

**Validation Checklist**:

- [ ] Build: `go build ./internal/apps/cipher/...` (zero errors)
- [ ] Linting: `golangci-lint run ./internal/apps/cipher/...` (zero warnings)
- [ ] Tests: `go test ./internal/apps/cipher/... -cover` (100% pass)
- [ ] Coverage: ≥95% production, ≥98% infrastructure
- [ ] Grep: 0 WithDefaultTenant usages
- [ ] Security: NO hardcoded passwords in tests
- [ ] Git: Conventional commits with evidence

**Final Commit**: `git commit -m "test(cipher-im): adapt to registration flow pattern"`

---

## Phase 2: JOSE-JA - Database Schema Migration (4-5 days)

### 2.1 Create JOSE Domain Models

**File**: `internal/apps/jose/ja/domain/models.go`

**Models**:

```go
type ElasticJWK struct {
    ID              googleUuid.UUID
    KID             string
    TenantID        googleUuid.UUID  // Multi-tenancy
    Algorithm       string
    Use             string  // "sig" or "enc"
    MaxMaterials    int
    Status          string  // "active", "rotated", "revoked"
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type MaterialKey struct {
    ID              googleUuid.UUID
    ElasticJWKID    googleUuid.UUID
    Status          string  // "active", "historical", "revoked"
    KeyData         []byte  // Encrypted with barrier service
    CreatedAt       time.Time
    RevokedAt       *time.Time
}

type JWKSConfig struct {
    ID              googleUuid.UUID
    TenantID        googleUuid.UUID
    AllowCrossTenant bool  // Set via tenant management API
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type AuditConfig struct {
    ID              googleUuid.UUID
    TenantID        googleUuid.UUID
    Operation       string  // "sign", "encrypt", "decrypt", "verify"
    Enabled         bool
    SamplingRate    float64
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type AuditLog struct {
    ID              googleUuid.UUID
    TenantID        googleUuid.UUID
    SessionID       googleUuid.UUID  // Added to track authn/authz context
    Operation       string
    ResourceType    string
    ResourceID      string
    Success         bool
    Metadata        map[string]interface{}
    CreatedAt       time.Time
}
```

**CRITICAL: Multi-Tenancy**:

- ✅ ALL models MUST include `TenantID googleUuid.UUID` field
- ✅ Repository queries MUST filter by `tenant_id`
- ❌ Repository queries MUST NOT filter by `realm_id` (realms are authn only, NOT data scope)

**Quality Gates**:

1. Build: `go build ./internal/apps/jose/...`
2. Linting: `golangci-lint run --fix ./internal/apps/jose/...`
3. Evidence: Build output + commit hash

---

### 2.2 Create JOSE Database Migrations

**Directory**: `internal/apps/jose/ja/repository/migrations/`

**Files**:

- `2001_elastic_jwk.up.sql` / `2001_elastic_jwk.down.sql`
- `2002_material_keys.up.sql` / `2002_material_keys.down.sql`
- `2003_jwks_config.up.sql` / `2003_jwks_config.down.sql`
- `2004_audit_config.up.sql` / `2004_audit_config.down.sql`
- `2005_audit_log.up.sql` / `2005_audit_log.down.sql`

**Schema Example** (`2001_elastic_jwk.up.sql`):

```sql
CREATE TABLE IF NOT EXISTS elastic_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    kid TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    use TEXT NOT NULL CHECK (use IN ('sig', 'enc')),
    max_materials INTEGER NOT NULL DEFAULT 10,
    status TEXT NOT NULL CHECK (status IN ('active', 'rotated', 'revoked')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE (kid, tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_elastic_jwks_tenant ON elastic_jwks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_elastic_jwks_status ON elastic_jwks(status);
```

**CRITICAL: Cross-Database Compatibility**:

- ✅ Use `TEXT` for UUIDs (NOT `uuid` type - SQLite doesn't support)
- ✅ Use `TIMESTAMP` for dates (NOT `timestamptz` - SQLite doesn't support)
- ✅ Use `CHECK` constraints for enums (portable)
- ✅ Use `DEFAULT CURRENT_TIMESTAMP` (portable)

**Quality Gates**:

1. Migration applies to PostgreSQL 18+
2. Migration applies to SQLite 3.19+
3. Migration rollback works (down migrations)
4. Tests verify schema constraints
5. Evidence: Test output + commit hash

---

### 2.3 Implement JOSE Repositories

**Files**:

- `internal/apps/jose/ja/repository/elastic_jwk_repository.go`
- `internal/apps/jose/ja/repository/material_key_repository.go`
- `internal/apps/jose/ja/repository/jwks_config_repository.go`
- `internal/apps/jose/ja/repository/audit_config_repository.go`
- `internal/apps/jose/ja/repository/audit_log_repository.go`

**Interface Example**:

```go
type ElasticJWKRepository interface {
    Create(ctx context.Context, jwk *domain.ElasticJWK) error
    GetByID(ctx context.Context, id, tenantID googleUuid.UUID) (*domain.ElasticJWK, error)
    GetByKID(ctx context.Context, kid string, tenantID googleUuid.UUID) (*domain.ElasticJWK, error)
    List(ctx context.Context, tenantID googleUuid.UUID, page, size int) ([]*domain.ElasticJWK, int, error)
    Update(ctx context.Context, jwk *domain.ElasticJWK) error
}
```

**CRITICAL: Repository WHERE Clauses**:

```go
// ✅ CORRECT: Filter by tenant_id only
db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&jwk)

// ❌ WRONG: NEVER filter by realm_id (realms are authn only)
db.Where("id = ? AND tenant_id = ? AND realm_id = ?", id, tenantID, realmID).First(&jwk)
```

**Quality Gates**:

1. Build: `go build ./internal/apps/jose/...`
2. Tests: `go test ./internal/apps/jose/ja/repository/... -cover`
3. Coverage: ≥98% (infrastructure code)
4. Mutation: ≥98% (gremlins score)
5. Security: SQL injection prevention (parameterized queries)
6. Evidence: Coverage + mutation reports + commit hash

---

### 2.4 Phase 2 Validation

**Validation Checklist**:

- [ ] Build: `go build ./internal/apps/jose/...`
- [ ] Linting: `golangci-lint run ./internal/apps/jose/...`
- [ ] Tests: `go test ./internal/apps/jose/ja/repository/... -cover` (100% pass)
- [ ] Coverage: ≥98% (infrastructure)
- [ ] Mutation: ≥98% (gremlins)
- [ ] Migrations: Apply to PostgreSQL 18+ and SQLite
- [ ] Repository: NO realm_id filtering in WHERE clauses
- [ ] Git: Conventional commits

**Final Commit**: `git commit -m "feat(jose-ja): implement database schema and repositories"`

---

## Phase 3: JOSE-JA - ServerBuilder Integration (3-4 days)

### 3.1 Create JOSE Server Configuration

**File**: `internal/apps/jose/ja/server/config/config.go`

**Configuration**:

```go
type Settings struct {
    ServiceTemplateServerSettings *cryptoutilTemplateBuilder.ServiceTemplateServerSettings

    // JOSE-specific settings (NONE - all in ServiceTemplateServerSettings)
}
```

**CRITICAL: Session Configuration**:

```yaml
# configs/jose-ja/jose-ja-server.yaml

# Browser sessions (user-facing)
browser-session-expiration: "24h"
browser-session-idle-timeout: "30m"

# Service sessions (machine-to-machine)
service-session-expiration: "1h"
service-session-idle-timeout: "15m"

# WRONG (not granular enough):
# session-expiration: "24h"
# session-idle-timeout: "30m"
```

**CRITICAL: Configuration Priority**:

1. **Docker Secrets** (mounted files) - HIGHEST PRIORITY
2. **Mounted YAML Config** (configs/jose-ja/jose-ja-server.yaml)
3. **Environment Variables** - ONLY for extreme corner cases (NOT secure, NOT preferred)
4. **CLI Parameters** - ONLY for temporary overrides (e.g., --debug)

**Quality Gates**:

1. Build: `go build ./internal/apps/jose/...`
2. Tests: Config loads from YAML, Docker secrets, ENV (in priority order)
3. Evidence: Test output + commit hash

---

### 3.2 Create JOSE Public Server

**File**: `internal/apps/jose/ja/server/server.go`

**Structure**:

```go
type JoseServer struct {
    app                 *cryptoutilTemplateServer.Application
    elasticJWKService   *service.ElasticJWKService
    jweService          *service.JWEService
    jwsService          *service.JWSService
    jwtService          *service.JWTService
    auditLogService     *service.AuditLogService
}

func NewFromConfig(ctx context.Context, cfg *config.Settings) (*JoseServer, error) {
    builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

    // Register domain migrations (2001-2005)
    builder.WithDomainMigrations(repository.MigrationsFS, "migrations")

    // Register domain-specific routes
    builder.WithPublicRouteRegistration(func(
        base *cryptoutilTemplateServer.PublicServerBase,
        res *cryptoutilTemplateBuilder.ServiceResources,
    ) error {
        // Create domain repositories
        elasticJWKRepo := repository.NewElasticJWKRepository(res.DB)
        materialKeyRepo := repository.NewMaterialKeyRepository(res.DB)

        // Create domain services
        elasticJWKService := service.NewElasticJWKService(elasticJWKRepo, materialKeyRepo, res.BarrierService)

        // Create HTTP handlers
        jwkHandlers := apis.NewJWKHandlers(elasticJWKService)

        // Register routes (NO service name in path)
        base.PublicServer.POST("/service/api/v1/jwk/generate", jwkHandlers.Generate)
        base.PublicServer.GET("/service/api/v1/jwk/list", jwkHandlers.List)
        base.PublicServer.GET("/service/api/v1/jwk/:id", jwkHandlers.Get)

        return nil
    })

    // Build complete infrastructure
    resources, err := builder.Build()
    if err != nil {
        return nil, err
    }

    return &JoseServer{
        app: resources.Application,
        // ... services
    }, nil
}
```

**CRITICAL: API Paths**:

- ✅ CORRECT: `/service/api/v1/jwk/generate` (no service name)
- ❌ WRONG: `/service/api/v1/jose/jwk/generate` (includes service name)
- ✅ CORRECT: `/admin/api/v1/audit/config` (consistent /admin/api/v1 prefix)
- ❌ WRONG: `/admin/v1/audit/config` (inconsistent with /service/api/v1)

**Quality Gates**:

1. Build: `go build ./internal/apps/jose/...`
2. Tests: Server starts, health checks pass
3. E2E: Public endpoints accessible
4. Evidence: Test output + commit hash

---

### 3.3 Create JOSE HTTP Handlers

**Files**:

- `internal/apps/jose/ja/server/apis/jwk_handlers.go`
- `internal/apps/jose/ja/server/apis/jws_handlers.go`
- `internal/apps/jose/ja/server/apis/jwe_handlers.go`
- `internal/apps/jose/ja/server/apis/jwt_handlers.go`
- `internal/apps/jose/ja/server/apis/jwks_handlers.go`
- `internal/apps/jose/ja/server/apis/audit_handlers.go`

**Example**: JWK Generate Handler

```go
func (h *JWKHandlers) Generate(c *fiber.Ctx) error {
    var req GenerateJWKRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
    }

    // Get tenant from session context (set by auth middleware)
    tenantID := c.Locals("tenant_id").(googleUuid.UUID)

    // Generate JWK
    jwk, err := h.elasticJWKService.Generate(c.Context(), &service.GenerateJWKParams{
        TenantID:      tenantID,
        Algorithm:     req.Algorithm,
        Use:           req.Use,
        MaxMaterials:  req.MaxMaterials,
        KID:           req.KID,
    })

    if err != nil {
        return err
    }

    return c.Status(fiber.StatusCreated).JSON(jwk)
}

type GenerateJWKRequest struct {
    Algorithm    string `json:"algorithm" validate:"required"`
    Use          string `json:"use" validate:"required,oneof=sig enc"`
    MaxMaterials int    `json:"max_material_keys,omitempty"`
    KID          string `json:"kid,omitempty"`
    // key_type REMOVED - implied by algorithm
    // key_size REMOVED - implied by algorithm
}
```

**CRITICAL: Request Parameter Simplification**:

- ✅ `algorithm` is REQUIRED (determines key_type and key_size)
- ❌ `key_type` is REMOVED (implied by algorithm: RS256 → RSA, ES256 → EC, EdDSA → OKP)
- ❌ `key_size` is REMOVED (implied by algorithm: RS256 → 2048, ES256 → P-256, EdDSA → Ed25519)

**Quality Gates**:

1. Build: `go build ./internal/apps/jose/...`
2. Tests: Handler unit tests (100% coverage)
3. Integration: Handler integration tests
4. Evidence: Coverage report + commit hash

---

### 3.4 Implement JOSE Business Logic Services

**Files**:

- `internal/apps/jose/ja/service/elastic_jwk_service.go`
- `internal/apps/jose/ja/service/material_rotation_service.go`
- `internal/apps/jose/ja/service/jws_service.go`
- `internal/apps/jose/ja/service/jwe_service.go`
- `internal/apps/jose/ja/service/jwt_service.go`
- `internal/apps/jose/ja/service/jwks_service.go`
- `internal/apps/jose/ja/service/audit_log_service.go`

**Quality Gates**:

1. Build: `go build ./internal/apps/jose/...`
2. Tests: Service unit tests (≥95% coverage)
3. Mutation: ≥85% (gremlins score)
4. Evidence: Coverage + mutation reports + commit hash

---

### 3.5 Phase 3 Validation

**Validation Checklist**:

- [ ] Build: `go build ./internal/apps/jose/...`
- [ ] Linting: `golangci-lint run ./internal/apps/jose/...`
- [ ] Tests: `go test ./internal/apps/jose/... -cover` (100% pass)
- [ ] Coverage: ≥95% production, ≥98% infrastructure
- [ ] Mutation: ≥85% production, ≥98% infrastructure
- [ ] Paths: No service name in request paths
- [ ] Config: Docker secrets > YAML > ENV priority
- [ ] Git: Conventional commits

**Final Commit**: `git commit -m "feat(jose-ja): integrate ServerBuilder pattern"`

---

## Phase 4: JOSE-JA - Elastic JWK Implementation (4-5 days)

**See V3 for detailed tasks** - NO substantive changes beyond:

- ✅ Fix repository WHERE clauses (remove realm_id filtering)
- ✅ Fix test passwords (use cryptoutilMagic.TestPassword)
- ✅ Simplify Generate API (remove key_type, key_size)

---

## Phase 5: JOSE-JA - JWKS Endpoint (2-3 days)

**See V3 for detailed tasks** - NO substantive changes beyond:

- ✅ Cross-tenant JWKS access via tenant management API (not DB config)
- ✅ Fix API paths (no service name)

---

## Phase 6: JOSE-JA - Audit Logging (2-3 days)

**See V3 for detailed tasks** - NO substantive changes

---

## Phase 7: JOSE-JA - Path Migration (2-3 days)

**CRITICAL Changes**:

- ✅ Migrate from `/api/jose/*` to `/service/api/v1/*` and `/browser/api/v1/*`
- ✅ Remove service name from paths (was `/api/jose/jwk/*`, now `/api/v1/jwk/*`)
- ✅ Consistent admin paths: `/admin/api/v1/*` (NOT `/admin/v1/*`)

---

## Phase 8: JOSE-JA - E2E Testing (3-4 days)

**E2E Test Execution Pattern**:

- **Q9.1**: Docker Compose for E2E tests (realistic customer experience, NOT direct Go)
- **Q9.2**: Docker Compose starts PostgreSQL container (NOT test-containers, NOT SQLite)
- **Q9.3**: Per product-service e2e/ subdirectory (`internal/apps/jose/ja/e2e/` pattern)

**Directory Structure**:

```
internal/apps/jose/ja/
├── domain/
├── repository/
├── service/
├── server/
└── e2e/              # E2E tests in product-service subdirectory
    ├── registration_test.go
    ├── jwk_generation_test.go
    ├── elastic_key_rotation_test.go
    └── audit_logging_test.go
```

**See V3 for detailed tasks** - CRITICAL changes:

- ✅ TestMain pattern with registration flow
- ✅ NO hardcoded passwords
- ✅ Test both `/service/api/v1/*` and `/browser/api/v1/*` paths

---

## Phase 9: JOSE-JA - Documentation (2-3 days)

**Copilot Instructions applyTo Patterns**:

- **Q6.1**: NO conditional applyTo patterns (all instructions apply to `**`)
- **Q6.2**: NO glob patterns (keep `applyTo: "**"` for all instruction files)

**Rationale**: Glob patterns add complexity without benefit. Global application (`**`) is simpler and sufficient.

**Prompt Implementation Priority**:

- **Q7.1**: NO prompts desired (user does not want code-review, test-generate, fix-bug, refactor-extract, optimize-performance, generate-docs)
- **Q7.2**: N/A (no prompts to implement)

**Agent Handoff Patterns**:

- **Q8.1**: N/A (user unclear on context)
- **Q8.2**: N/A (user unclear on context)

**Documentation Standards**:

- **Q10.1**: ARCHITECTURE.md high-level only (NO code examples, < 1000 lines)
- **Q10.2**: Update ARCHITECTURE.md when user decides (discretionary)
- **Q10.3**: NO versioning for ARCHITECTURE.md (git log provides history)
- **Q10.4**: Minimal code examples in instruction files (1-2 snippets per file)

**CRITICAL Changes**:

- ❌ NO MIGRATION-GUIDE.md (pre-alpha project)
- ✅ API-REFERENCE.md (fix paths, simplify params)
- ✅ DEPLOYMENT.md (no ENVs, no K8s, OTLP only, port 9090)

---

## Phase W: Service-Template - Refactor ServerBuilder Bootstrap Logic

**Estimated Duration**: 2-3 days
**Dependencies**: Phase 0 complete
**Prerequisites**: Service-template registration flow implemented

### Problem Statement

**Current State**: `server_builder.go` contains mixed concerns:

- HTTPS listener configuration ✅ (correct responsibility)
- Route registration ✅ (correct responsibility)
- Internal service initialization ❌ (wrong layer)
- Repository bootstrap ❌ (wrong layer)

**Desired State**: Clear separation of concerns:

- **ServerBuilder**: HTTPS servers, routes, middleware only
- **ApplicationCore**: Business logic bootstrap (repos, services, dependencies)

**Architecture Violation**: Builder pattern currently initializes:

```go
sqlDB
barrierRepo
barrierService
realmRepo
realmService
sessionManager
tenantRepo
userRepo
joinRequestRepo
registrationService
rotationService
statusService
```

These belong in ApplicationCore startup, not builder configuration.

---

### W.1 Move Bootstrap Logic to ApplicationCore

**File**: `internal/apps/template/service/server/application/application_core.go`

**New Method**: `StartApplicationCore(ctx context.Context, config *Config, db *gorm.DB) (*CoreServices, error)`

**Returns**:

```go
type CoreServices struct {
    DB                  *gorm.DB
    BarrierRepo         repository.BarrierRepository
    BarrierService      service.BarrierService
    UnsealKeysService   *barrier.UnsealKeysService  // CRITICAL: Missing from current ServiceResources
    RealmRepo           repository.TenantRealmRepository
    RealmService        service.RealmService
    SessionManager      *businesslogic.SessionManagerService
    TenantRepo          repository.TenantRepository
    UserRepo            repository.UserRepository
    JoinRequestRepo     repository.JoinRequestRepository
    RegistrationService service.RegistrationService
    RotationService     service.RotationService
    StatusService       service.StatusService
}
```

**Changes**:

- Move all repo/service initialization from `server_builder.go` to `application_core.go`
- Encapsulate dependencies (e.g., sessionManager depends on realmService)
- Return initialized services for route registration
- Populate `core.Basic.UnsealKeysService` (currently missing)

---

### W.2 Update ServerBuilder

**File**: `internal/apps/template/service/server/builder/server_builder.go`

**Changes**:

- Remove direct initialization of repos/services
- Call `ApplicationCore.StartApplicationCore(ctx, config, db)`
- Use returned `CoreServices` for route registration
- Focus ONLY on HTTPS listeners and route setup

---

### W.3 Update ServiceResources

**File**: `internal/apps/template/service/server/builder/server_builder.go`

**Changes**:

- Add `UnsealKeysService` field to `ServiceResources` struct
- Ensure all core services exposed via ServiceResources
- Update all main.go files that access ServiceResources

---

### W.4 Quality Gates

**Validation**:

1. Build: `go build ./internal/apps/template/...`
2. Linting: `golangci-lint run --fix ./internal/apps/template/...`
3. Tests: `go test ./internal/apps/template/... -cover`
4. Coverage: ≥85% maintained
5. Mutation: Deferred to Phase Y
6. Commit: `refactor(service-template): move bootstrap logic to ApplicationCore`

---

## Directory Structure

```
cryptoutil/
├── deployments/
│   └── jose-ja/          # Product-service naming (NOT jose/)
│       ├── compose.yml
│       └── Dockerfile
├── configs/
│   └── jose-ja/          # Product-service naming (NOT jose/)
│       └── jose-ja-server.yaml
├── internal/
│   └── apps/
│       └── jose/
│           └── ja/
│               ├── domain/
│               ├── repository/
│               ├── service/
│               └── server/
└── docs/
    └── jose-ja/
        ├── JOSE-JA-REFACTORING-PLAN-V4.md
        ├── JOSE-JA-REFACTORING-TASKS-V4.md
        ├── API-REFERENCE.md
        └── DEPLOYMENT.md
```

---

## Key Differences from V3

1. **Port 9090**: Admin endpoints on 9090 (was incorrectly 8080)
2. **PostgreSQL 18+**: Requirement (was incorrectly 16+)
3. **Directory Structure**: deployments/jose-ja/, configs/jose-ja/ (was jose/)
4. **No ENVs**: Docker secrets > YAML config (was promoting ENVs)
5. **Session Configs**: Separate browser vs service (was combined)
6. **No K8s Docs**: Only Docker Compose (was including Kubernetes)
7. **OTLP Only**: No Prometheus scraping endpoint (was exposing /admin/v1/metrics)
8. **Consistent Paths**: /admin/api/v1, /service/api/v1, /browser/api/v1 (was mixing /admin/v1 and /admin/api/v1)
9. **No Service Name in Paths**: /service/api/v1/jwk/*(was /service/api/v1/jose/jwk/*)
10. **Realms for Authn Only**: Removed realm_id from repository WHERE clauses
11. **No Hardcoded Passwords**: cryptoutilMagic.TestPassword or UUIDv7 (was using "pass1", "pass2")
12. **No Migration Guide**: Deleted (pre-alpha project)
13. **Simplified API**: key_type, key_size removed (implied by algorithm)
14. **No Future Deferrals**: All requirements mandatory (was deferring cross-tenant access)

---

## Timeline Summary

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
