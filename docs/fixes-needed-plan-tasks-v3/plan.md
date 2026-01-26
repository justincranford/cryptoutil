# Unified Implementation Plan - Cryptoutil Service Template & Coverage

**Last Updated**: 2026-01-25
**Purpose**: Merged plan combining service-template migration (V1) and test coverage implementation (V2)
**Status**: Service template migration 74% complete, test coverage 79% complete

---

## Overview

This unified plan combines two complementary implementation tracks:

1. **Service-Template Migration (V1)**: JOSE-JA, cipher-im, and shared infrastructure
2. **Test Coverage Implementation (V2)**: Container mode, mTLS, configuration validation

Both tracks share common infrastructure (service-template pattern) and quality gates (95% coverage, **98% mutation efficacy IDEAL** / 85% minimum with documented blockers).

**CRITICAL Quality Standards Clarification**:

- **Mutation Efficacy Ideal Goal**: **≥98%** - This is the target we ALWAYS strive for
- **Mutation Efficacy Absolute Minimum**: ≥85% - Only acceptable when blocked by external factors with comprehensive documented justification
- **Quality Over Speed Principle**: NEVER settle for 85% when 98% is achievable
- **NO Services May Be Skipped**: ALL services must achieve mutation testing (cipher-im being skipped was UNACCEPTABLE and violated this plan)

---

## V1: Service-Template Migration

# JOSE-JA Refactoring Plan V4

**Last Updated**: 2026-01-18

**Architecture Reference**: See [docs/arch/ARCHITECTURE.md](../arch/ARCHITECTURE.md) for comprehensive design patterns, principles, and implementation guidelines.

**Tasks Reference**: See [tasks.md](tasks.md) for detailed task breakdown with checkboxes.

**Progress Tracking - MANDATORY**: As tasks are completed, check them off in tasks.md to track progress. Each checkbox represents objective evidence of completion (build succeeds, tests pass, coverage met, mutation score met, commit created).

**Critical Fixes from V3**:

- âœ… Port 9090 for admin endpoints (was incorrectly 8080)
- âœ… PostgreSQL 18+ requirement (was incorrectly 16+)
- âœ… Directory structure: deployments/jose-ja/, configs/jose-ja/ (was jose/)
- âœ… Removed CGO_ENABLED mentions (implied by project)
- âœ… Docker secrets > YAML config > ENV (was promoting ENVs)
- âœ… Separate browser vs service session configs
- âœ… OTLP only (removed Prometheus scraping endpoint)
- âœ… Consistent API paths: /admin/api/v1, /service/api/v1, /browser/api/v1
- âœ… No service name in paths (was /service/api/v1/jose/*)
- âœ… Realms are authn only (removed from repository WHERE clauses)
- âœ… No hardcoded passwords in tests (use magic constants or UUIDv7)
- âœ… key_type implied by algorithm (simplified API)
- âœ… All requirements mandatory (no "Future" deferrals)

---

## Core Principles - MANDATORY

**Quality Over Speed (NO EXCEPTIONS)**:

- âœ… **Correctness**: ALL code must be functionally correct with comprehensive tests
- âœ… **Completeness**: NO tasks skipped, NO features deprioritized, NO shortcuts
- âœ… **Thoroughness**: Evidence-based validation at every step (build, lint, test, coverage, mutation)
- âœ… **Reliability**: â‰¥95% coverage production, â‰¥98% infrastructure/utility, â‰¥85% mutation production, â‰¥98% mutation infrastructure
- âœ… **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- âŒ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- âŒ **Premature Completion**: NEVER mark tasks complete without objective evidence

**Continuous Execution (NO STOPPING)**:

- Work continues until ALL tasks complete OR user clicks STOP button
- NEVER stop to ask permission between tasks ("Should I continue?")
- NEVER pause for status updates or celebrations ("Here's what we did...")
- NEVER give up when encountering complexity (find solutions, refactor, investigate)
- NEVER skip tasks to "save time" or because they seem "less important"
- Task complete â†’ Commit â†’ IMMEDIATELY start next task (zero pause, zero text to user)

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

- âŒ REMOVE: `WithDefaultTenant(tenantID, realmID)` method
- âŒ REMOVE: Call to `ensureDefaultTenant()` in `Build()` method
- âŒ REMOVE: `ensureDefaultTenant()` helper method
- âŒ REMOVE: `defaultTenantID` and `defaultRealmID` fields from ServerBuilder struct
- âŒ REMOVE: Passing defaultTenantID/defaultRealmID to SessionManagerService constructor

**Quality Gates**:

1. Build: `go build ./internal/apps/template/...`
2. Linting: `golangci-lint run --fix ./internal/apps/template/...`
3. Tests: ALL template tests updated to use registration pattern
4. Evidence: Commit hash with conventional message

---

### 0.2 Remove EnsureDefaultTenant Helper

**File**: `internal/apps/template/service/server/repository/seeding.go`

**Changes**:

- âŒ DELETE entire file (contains only EnsureDefaultTenant function)

**Quality Gates**:

1. Build: `go build ./internal/apps/template/...`
2. Linting: `golangci-lint run --fix`
3. Evidence: File deleted, no references remain

---

### 0.3 Update SessionManagerService (Remove Single-Tenant Methods)

**File**: `internal/apps/template/service/server/businesslogic/session_manager_service.go`

**Changes**:

- âŒ REMOVE: `defaultTenantID` field
- âŒ REMOVE: `defaultRealmID` field
- âŒ REMOVE: `IssueBrowserSession(ctx, userID)` method (single-tenant version)
- âŒ REMOVE: `ValidateBrowserSession(ctx, token)` method (single-tenant version)
- âŒ REMOVE: `IssueServiceSession(ctx, clientID)` method (single-tenant version)
- âŒ REMOVE: `ValidateServiceSession(ctx, token)` method (single-tenant version)
- âœ… KEEP: `IssueBrowserSessionWithTenant(ctx, userID, tenantID, realmID)` (multi-tenant version)
- âœ… KEEP: `ValidateBrowserSessionWithTenant(ctx, token)` (multi-tenant version)
- âœ… KEEP: `IssueServiceSessionWithTenant(ctx, clientID, tenantID, realmID)` (multi-tenant version)
- âœ… KEEP: `ValidateServiceSessionWithTenant(ctx, token)` (multi-tenant version)
- âœ… UPDATE: Constructor `NewSessionManagerService()` to remove defaultTenantID/defaultRealmID params

**Quality Gates**:

1. Build: `go build ./internal/apps/template/...`
2. Linting: `golangci-lint run --fix`
3. Tests: All session manager tests updated
4. Evidence: Commit hash

---

### 0.4 Remove Template Magic Constants

**Files**: `internal/shared/magic/magic_template.go`, `internal/apps/template/service/server/repository/seeding_test.go` (if exists)

**Changes**:

- âŒ REMOVE: `TemplateDefaultTenantID` constant (if exists)
- âŒ REMOVE: `TemplateDefaultRealmID` constant (if exists)
- âœ… VERIFY: `grep -r "TemplateDefaultTenantID" internal/` returns 0 results
- âœ… VERIFY: `grep -r "TemplateDefaultRealmID" internal/` returns 0 results

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

- âŒ NO email validation on username field (username can be non-email)
- âœ… Email/password authentication is a DIFFERENT realm (not implemented yet)
- âœ… Username field accepts any string (simplified registration flow)

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

- âœ… **Q2.1**: Template infrastructure provides admin dashboard APIs (NOT domain-specific)
- âŒ **Q2.2**: NO email notifications (users poll status via login attempts)
- âŒ **Q2.3**: NO webhook callbacks (keep template simple)
- âŒ **Q2.4**: NO unauthenticated status API (users poll via login: HTTP 403=pending, HTTP 401=rejected, HTTP 200=approved)

**Rate Limiting**:

- âœ… **Q3.1**: Per IP address only (10 registrations per IP per hour)
- âœ… **Q3.2**: In-memory (sync.Map) - simple, single-node, lost on restart
- âœ… **Q3.3**: Configurable thresholds with low defaults

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

- âœ… Admin endpoints: `/admin/api/v1/*`
- âœ… Service endpoints: `/service/api/v1/*`
- âœ… Browser endpoints: `/browser/api/v1/*`
- âŒ NEVER use `/admin/v1/*` (inconsistent with /service/api/v1 and /browser/api/v1)
- âŒ NEVER include service name in paths (e.g., `/service/api/v1/jose/*` is WRONG)

**Quality Gates**:

1. Build: `go build ./internal/apps/template/...`
2. Tests: Integration tests for ALL endpoints
3. Coverage: â‰¥95% (handler code)
4. Security: CSRF protection on browser endpoints
5. Evidence: Test output + commit hash

---

### 0.9 Update ServerBuilder Registration

**File**: `internal/apps/template/service/server/builder/server_builder.go`

**Changes**:

- âœ… ADD: Registration handler routes in `Build()` method
- âœ… ADD: Join request handler routes in `Build()` method
- âœ… VERIFY: No `WithDefaultTenant()` calls remain

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
- [ ] Coverage: â‰¥95% production, â‰¥98% infrastructure
- [ ] Mutation: â‰¥85% production, â‰¥98% infrastructure
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

- âŒ REMOVE: Any `WithDefaultTenant()` calls (if exist)
- âœ… VERIFY: `grep -r "WithDefaultTenant" internal/apps/cipher/` returns 0 results
- âœ… VERIFY: All tests use `registerUser()` or `registerClient()` for tenant creation

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

- âŒ NEVER: `registerUser(server, "user1", "pass1", ...)`
- âœ… ALWAYS: `registerUser(server, "user1", cryptoutilMagic.TestPassword, ...)`
- âœ… ALTERNATIVE: `registerUser(server, "user1", googleUuid.NewV7().String(), ...)`

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
- [ ] Coverage: â‰¥95% production, â‰¥98% infrastructure
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

- âœ… ALL models MUST include `TenantID googleUuid.UUID` field
- âœ… Repository queries MUST filter by `tenant_id`
- âŒ Repository queries MUST NOT filter by `realm_id` (realms are authn only, NOT data scope)

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

- âœ… Use `TEXT` for UUIDs (NOT `uuid` type - SQLite doesn't support)
- âœ… Use `TIMESTAMP` for dates (NOT `timestamptz` - SQLite doesn't support)
- âœ… Use `CHECK` constraints for enums (portable)
- âœ… Use `DEFAULT CURRENT_TIMESTAMP` (portable)

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
// âœ… CORRECT: Filter by tenant_id only
db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&jwk)

// âŒ WRONG: NEVER filter by realm_id (realms are authn only)
db.Where("id = ? AND tenant_id = ? AND realm_id = ?", id, tenantID, realmID).First(&jwk)
```

**Quality Gates**:

1. Build: `go build ./internal/apps/jose/...`
2. Tests: `go test ./internal/apps/jose/ja/repository/... -cover`
3. Coverage: â‰¥98% (infrastructure code)
4. Mutation: â‰¥98% (gremlins score)
5. Security: SQL injection prevention (parameterized queries)
6. Evidence: Coverage + mutation reports + commit hash

---

### 2.4 Phase 2 Validation

**Validation Checklist**:

- [ ] Build: `go build ./internal/apps/jose/...`
- [ ] Linting: `golangci-lint run ./internal/apps/jose/...`
- [ ] Tests: `go test ./internal/apps/jose/ja/repository/... -cover` (100% pass)
- [ ] Coverage: â‰¥98% (infrastructure)
- [ ] Mutation: â‰¥98% (gremlins)
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

- âœ… CORRECT: `/service/api/v1/jwk/generate` (no service name)
- âŒ WRONG: `/service/api/v1/jose/jwk/generate` (includes service name)
- âœ… CORRECT: `/admin/api/v1/audit/config` (consistent /admin/api/v1 prefix)
- âŒ WRONG: `/admin/v1/audit/config` (inconsistent with /service/api/v1)

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

- âœ… `algorithm` is REQUIRED (determines key_type and key_size)
- âŒ `key_type` is REMOVED (implied by algorithm: RS256 â†’ RSA, ES256 â†’ EC, EdDSA â†’ OKP)
- âŒ `key_size` is REMOVED (implied by algorithm: RS256 â†’ 2048, ES256 â†’ P-256, EdDSA â†’ Ed25519)

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
2. Tests: Service unit tests (â‰¥95% coverage)
3. Mutation: â‰¥85% (gremlins score)
4. Evidence: Coverage + mutation reports + commit hash

---

### 3.5 Phase 3 Validation

**Validation Checklist**:

- [ ] Build: `go build ./internal/apps/jose/...`
- [ ] Linting: `golangci-lint run ./internal/apps/jose/...`
- [ ] Tests: `go test ./internal/apps/jose/... -cover` (100% pass)
- [ ] Coverage: â‰¥95% production, â‰¥98% infrastructure
- [ ] Mutation: â‰¥85% production, â‰¥98% infrastructure
- [ ] Paths: No service name in request paths
- [ ] Config: Docker secrets > YAML > ENV priority
- [ ] Git: Conventional commits

**Final Commit**: `git commit -m "feat(jose-ja): integrate ServerBuilder pattern"`

---

## Phase 4: JOSE-JA - Elastic JWK Implementation (4-5 days)

**See V3 for detailed tasks** - NO substantive changes beyond:

- âœ… Fix repository WHERE clauses (remove realm_id filtering)
- âœ… Fix test passwords (use cryptoutilMagic.TestPassword)
- âœ… Simplify Generate API (remove key_type, key_size)

---

## Phase 5: JOSE-JA - JWKS Endpoint (2-3 days)

**See V3 for detailed tasks** - NO substantive changes beyond:

- âœ… Cross-tenant JWKS access via tenant management API (not DB config)
- âœ… Fix API paths (no service name)

---

## Phase 6: JOSE-JA - Audit Logging (2-3 days)

**See V3 for detailed tasks** - NO substantive changes

---

## Phase 7: JOSE-JA - Path Migration (2-3 days)

**CRITICAL Changes**:

- âœ… Migrate from `/api/jose/*` to `/service/api/v1/*` and `/browser/api/v1/*`
- âœ… Remove service name from paths (was `/api/jose/jwk/*`, now `/api/v1/jwk/*`)
- âœ… Consistent admin paths: `/admin/api/v1/*` (NOT `/admin/v1/*`)

---

## Phase 8: JOSE-JA - E2E Testing (3-4 days)

**E2E Test Execution Pattern**:

- **Q9.1**: Docker Compose for E2E tests (realistic customer experience, NOT direct Go)
- **Q9.2**: Docker Compose starts PostgreSQL container (NOT test-containers, NOT SQLite)
- **Q9.3**: Per product-service e2e/ subdirectory (`internal/apps/jose/ja/e2e/` pattern)

**Directory Structure**:

```
internal/apps/jose/ja/
â”œâ”€â”€ domain/
â”œâ”€â”€ repository/
â”œâ”€â”€ service/
â”œâ”€â”€ server/
â””â”€â”€ e2e/              # E2E tests in product-service subdirectory
    â”œâ”€â”€ registration_test.go
    â”œâ”€â”€ jwk_generation_test.go
    â”œâ”€â”€ elastic_key_rotation_test.go
    â””â”€â”€ audit_logging_test.go
```

**See V3 for detailed tasks** - CRITICAL changes:

- âœ… TestMain pattern with registration flow
- âœ… NO hardcoded passwords
- âœ… Test both `/service/api/v1/*` and `/browser/api/v1/*` paths

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

- âŒ NO MIGRATION-GUIDE.md (pre-alpha project)
- âœ… API-REFERENCE.md (fix paths, simplify params)
- âœ… DEPLOYMENT.md (no ENVs, no K8s, OTLP only, port 9090)

---

## Phase W: Service-Template - Refactor ServerBuilder Bootstrap Logic

**Estimated Duration**: 2-3 days
**Dependencies**: Phase 0 complete
**Prerequisites**: Service-template registration flow implemented

### Problem Statement

**Current State**: `server_builder.go` contains mixed concerns:

- HTTPS listener configuration âœ… (correct responsibility)
- Route registration âœ… (correct responsibility)
- Internal service initialization âŒ (wrong layer)
- Repository bootstrap âŒ (wrong layer)

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
4. Coverage: â‰¥85% maintained
5. Mutation: Deferred to Phase Y
6. Commit: `refactor(service-template): move bootstrap logic to ApplicationCore`

---

## Directory Structure

```
cryptoutil/
â”œâ”€â”€ deployments/
â”‚   â””â”€â”€ jose-ja/          # Product-service naming (NOT jose/)
â”‚       â”œâ”€â”€ compose.yml
â”‚       â””â”€â”€ Dockerfile
â”œâ”€â”€ configs/
â”‚   â””â”€â”€ jose-ja/          # Product-service naming (NOT jose/)
â”‚       â””â”€â”€ jose-ja-server.yaml
â”œâ”€â”€ internal/
â”‚   â””â”€â”€ apps/
â”‚       â””â”€â”€ jose/
â”‚           â””â”€â”€ ja/
â”‚               â”œâ”€â”€ domain/
â”‚               â”œâ”€â”€ repository/
â”‚               â”œâ”€â”€ service/
â”‚               â””â”€â”€ server/
â””â”€â”€ docs/
    â””â”€â”€ jose-ja/
        â”œâ”€â”€ JOSE-JA-REFACTORING-PLAN-V4.md
        â”œâ”€â”€ JOSE-JA-REFACTORING-TASKS-V4.md
        â”œâ”€â”€ API-REFERENCE.md
        â””â”€â”€ DEPLOYMENT.md
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

---

## V2: Test Coverage Implementation

# Merged Plan: Workflow Fixes + Documentation Improvements

**Last Updated**: 2026-01-24
**Purpose**: Combined analysis of workflow fixes (V2) and documentation improvements (V3)
**Status**: V3 100% complete, V2 issues fixed but tests incomplete

## Session Overview

This document merges two distinct but complementary sessions:

- **V2**: Workflow test fixes for container mode, mTLS, DAST diagnostics
- **V3**: Documentation clarification, lessons extraction, prompt enhancements

---

## V2: Workflow Fixes Analysis

### Executive Summary

**Primary Fixes Made**:

1. **SQLite Container Mode** (commit 9e9da31c): Added `sqlite://` URL support for containerized environments
2. **mTLS Container Mode** (commit f58c6ff6): Disabled mTLS for container mode (0.0.0.0 binding) to fix healthchecks
3. **DAST Diagnostics** (commit 80a69d18): Improved artifact upload and diagnostic output for workflow failures
4. **DAST Configuration** (ongoing investigation): Potential YAML field mapping issue with `dev-mode` config

**Shift-Left Strategy**:

- Maximize reusability through service-template tests (shared across 9 product-services)
- Prioritize security-critical code coverage (mTLS, TLS client auth)
- Add comprehensive validation testing (config combinations, deployment modes)
- Implement integration tests for cross-module interactions

---

## Issue #1: Container Mode - Explicit Database URL Support

### Problem Statement

Containers needed:

- `0.0.0.0` bind address (Docker networking requirement - container ports must be exposed)
- Explicit database URL configuration (SQLite OR PostgreSQL - independent choice)

However, the only path to SQLite required `dev: true`, which was REJECTED when combined with `0.0.0.0` binding (security restriction to prevent Windows Firewall prompts in local development).

**Key Insight**: Database choice (SQLite vs PostgreSQL) is INDEPENDENT of bind address (127.0.0.1 vs 0.0.0.0).

### Root Cause

Configuration logic coupled database selection to mode flags instead of providing explicit URL configuration:

- `dev: true` â†’ SQLite (implicit) + MUST use 127.0.0.1 (security restriction)
- Production â†’ PostgreSQL (explicit URL) + MAY use 0.0.0.0

No explicit path existed for: Container networking (0.0.0.0) + SQLite database.

**Validation Logic**:

- Dev mode + 0.0.0.0 â†’ FAIL (intentional security restriction for local development)
- Container mode detection: `isContainerMode := settings.BindPublicAddress == "0.0.0.0"`

### Fix Implemented

Added explicit `sqlite://` URL support in `sql_settings_mapper.go`:

```go
if strings.HasPrefix(databaseURL, "sqlite://") {
    sqliteURL := strings.TrimPrefix(databaseURL, "sqlite://")
    telemetryService.Slogger.Debug("using SQLite database from explicit URL", "url", sqliteURL)
    return DBTypeSQLite, sqliteURL, nil
}
```

Updated container configs to use `database-url: "sqlite://file::memory:?cache=shared"` instead of `dev: true`.

**Result**: Containers can now use SQLite with 0.0.0.0 binding WITHOUT enabling dev mode (which would fail validation).

**Orthogonal Concerns Clarified**:

- **Database type** (SQLite vs PostgreSQL): Deployment choice, now explicitly configurable via URL
- **Bind address** (127.0.0.1 vs 0.0.0.0): Networking requirement
  - Dev mode: MUST use 127.0.0.1 (security restriction)
  - Container mode: MUST use 0.0.0.0 (Docker port mapping)
  - Production: Configurable based on deployment
- **Dev mode**: Security restriction for local development (prevents Windows Firewall prompts)

### Test Coverage Gap Analysis

**Existing Tests** (`sql_settings_mapper_test.go`):
âœ… Good unit test coverage (6 test cases)

- dev mode â†’ SQLite in-memory
- sqlite:// in-memory URL
- sqlite:// file-based URL
- postgres:// URL parsing
- Unsupported URL schemes
- Empty URL error handling

**Missing Tests**:
âŒ Integration tests validating config validation interactions:

- Container mode (0.0.0.0) + SQLite URL should pass validation
- Container mode (0.0.0.0) + dev mode should fail validation
- Container mode (0.0.0.0) + PostgreSQL URL should pass validation

âŒ Integration tests for complete configuration flow:

- Load config from YAML â†’ validate â†’ map database type â†’ initialize DB

### Lessons Learned

- Unit tests for individual functions are necessary but NOT sufficient
- Configuration validation needs tests for valid AND invalid combinations:
  - Valid: Container mode (0.0.0.0) + SQLite URL
  - Valid: Container mode (0.0.0.0) + PostgreSQL URL
  - Valid: Dev mode + 127.0.0.1 + (any database)
  - Invalid: Dev mode + 0.0.0.0 (security restriction)
- Database choice and bind address validation are ORTHOGONAL concerns
- Container mode should be treated as first-class deployment environment
- Integration tests needed to verify cross-module validation interactions

---

## Issue #2: mTLS Container Mode

### Problem Statement

Private admin server (port 9090) was configured with `tls.RequireAndVerifyClientCert` by default. This broke ALL healthchecks in container deployments because:

- Docker healthchecks use `wget` without client certificates
- `RequireAndVerifyClientCert` rejects connections without valid client certs
- Result: Healthcheck failures â†’ containers marked unhealthy â†’ workflow failures

### Root Cause

**CRITICAL: Zero unit tests for security-critical mTLS configuration logic.**

The `application_listener.go` code (lines 145-165) contains conditional logic for mTLS:

```go
privateClientAuth := tls.RequireAndVerifyClientCert
isContainerMode := settings.BindPublicAddress == cryptoutilSharedMagic.IPv4AnyAddress
if settings.DevMode || isContainerMode {
    privateClientAuth = tls.NoClientCert
}
```

grep search confirms: **ZERO tests** exist for this logic (searched for "mTLS|ClientAuth|RequireAndVerifyClientCert" in `internal/kms/server/application/*_test.go` - no matches found).

### Fix Implemented

Added container mode detection:

```go
isContainerMode := settings.BindPublicAddress == "0.0.0.0"
if settings.DevMode || isContainerMode {
    privateClientAuth = tls.NoClientCert
}
```

Removed `dev: true` from `postgresql-1.yml` and `postgresql-2.yml` configs since container mode now handles mTLS disable automatically.

### Test Coverage Gap Analysis

**Existing Tests**:
âŒ **ZERO tests for mTLS configuration logic** (most critical gap)

- No tests verifying `DevMode â†’ NoClientCert`
- No tests verifying container mode detection (`0.0.0.0` â†’ `NoClientCert`)
- No tests verifying production mode â†’ `RequireAndVerifyClientCert`
- No tests for private vs public server TLS client auth differences

âœ… Some middleware tests exist (`service_auth_test.go`) but DON'T test application-level TLS configuration

**Missing Tests (HIGH PRIORITY)**:
âŒ Unit tests for container mode detection:

- `BindPublicAddress == "0.0.0.0"` â†’ `isContainerMode = true`
- `BindPublicAddress == "127.0.0.1"` â†’ `isContainerMode = false`
- `BindPrivateAddress == "0.0.0.0"` â†’ should NOT affect container mode detection

âŒ Unit tests for mTLS configuration:

- `DevMode=true` â†’ `privateClientAuth = NoClientCert`
- Container mode (0.0.0.0) â†’ `privateClientAuth = NoClientCert`
- Production (127.0.0.1, DevMode=false) â†’ `privateClientAuth = RequireAndVerifyClientCert`
- Public server should NEVER use RequireAndVerifyClientCert (browser compatibility)

âŒ Integration tests:

- Container mode with SQLite â†’ healthcheck succeeds (wget without client cert)
- Container mode with PostgreSQL â†’ healthcheck succeeds
- Production mode â†’ healthcheck requires client cert OR uses different healthcheck method

### Lessons Learned

- Security-critical code (TLS/mTLS) MUST have comprehensive test coverage
- Configuration logic affecting security posture needs explicit testing
- Container mode detection is deployment-critical and needs dedicated tests
- Healthcheck compatibility should be validated in integration tests

---

## Issue #3: DAST Workflow Diagnostics

### Problem Statement

DAST workflow failures didn't upload artifacts, making diagnosis impossible. When containers failed to start, we had no stderr/stdout logs to determine root cause.

### Root Cause

Artifact upload step lacked `if: always()` condition, so it only ran when previous steps succeeded.

### Fix Implemented

Added `always()` condition and inline diagnostics:

```yaml
- name: GitHub Workflow artifacts
  if: always()  # Upload even on failure
  uses: actions/upload-artifact@v4

- name: Wait for servers ready
  run: |
    # ... health check logic ...
    if [ $? -ne 0 ]; then
      echo "âŒ Health check failed - Diagnostic output:"
      cat /tmp/kms-stderr.txt
      cat /tmp/kms-stdout.txt
      exit 1
    fi
```

### Test Coverage Gap Analysis

**Not applicable** - This is a workflow/CI/CD configuration issue, not application code.

However, it highlights the need for:

- Better local testing tools (`act` workflow runner, Docker Compose health checks)
- Diagnostic logging in application startup (already exists, just needed to be surfaced)

---

## Issue #4: DAST Configuration - dev-mode Field Mapping (UNDER INVESTIGATION)

### Problem Statement

DAST workflow config file shows `dev-mode: true` (kebab-case) in YAML, but application startup logs show `Dev mode (-d): false` (DevMode field not set).

**Evidence**:

```yaml
# Config file generation (ci-dast.yml)
dev-mode: true
bind-public-address: 127.0.0.1
bind-public-port: 8080
```

**Application logs**:

```
Dev mode (-d): false
Bind public address: 127.0.0.1
Bind public port: 8080
```

### Suspected Root Cause

YAML field mapping may not handle kebab-case â†’ PascalCase conversion correctly:

- `dev-mode` (YAML kebab-case) should map to `DevMode` (Go struct PascalCase)
- Other fields work: `bind-public-address` â†’ `BindPublicAddress` âœ…
- But `dev-mode` â†’ `DevMode` appears to fail âŒ

Possible issues:

1. Viper library mapstructure tag mismatch
2. Field name collision or override
3. Config loading order issue (CLI flag overriding YAML value)
4. Boolean field handling issue

### Investigation Needed

- Verify YAML â†’ struct field mapping for all boolean fields
- Check if other kebab-case fields map correctly
- Test config loading from file vs CLI flags vs environment variables
- Add debug logging to show config loading steps

### Test Coverage Gap Analysis

**Existing Tests**:
âŒ **ZERO tests for config file loading** (all existing tests use in-memory configs)

- No tests loading config from actual YAML files
- No tests verifying kebab-case â†’ PascalCase field mapping
- No tests checking config precedence (file vs CLI vs env vars)

**Missing Tests**:
âŒ YAML config loading tests:

- Load config from YAML file with kebab-case field names
- Verify all fields map correctly to struct (especially boolean fields)
- Test all casing styles: kebab-case, camelCase, PascalCase, snake_case

âŒ Config precedence tests:

- YAML file with `dev-mode: true` + CLI flag `-d=false` â†’ should use CLI value
- Environment variable overrides
- Default value fallbacks

âŒ Field validation tests:

- Boolean field parsing: `true`, `false`, `1`, `0`, `yes`, `no`
- Integer field parsing with ranges
- String field validation (bind address, log level enums)

### Lessons Learned

- Config loading from files needs explicit testing (not just in-memory configs)
- Field mapping (kebab-case, camelCase, PascalCase) should be validated
- Config precedence and overrides need test coverage
- Boolean fields are particularly error-prone (type coercion, string parsing)

---

## Cross-Cutting Issues

### Configuration Validation

**Problem**: Validation didn't account for valid container mode combinations (explicit database URLs with 0.0.0.0 binding)

**Key Clarification**: Database choice (SQLite vs PostgreSQL) is INDEPENDENT of bind address (127.0.0.1 vs 0.0.0.0). The validation rule is about dev-mode security restriction, NOT database type.

**Missing Tests**:

- Valid combinations:
  - Container + explicit SQLite URL + 0.0.0.0 binding
  - Container + PostgreSQL URL + 0.0.0.0 binding
  - Dev mode + 127.0.0.1 + (any database)
  - Production + configurable binding + (any database)
- Invalid combinations:
  - Dev mode + 0.0.0.0 binding (security restriction)
- Edge cases: empty fields, invalid IP addresses, out-of-range ports

### Container Mode Detection

**Problem**: Container mode not recognized as distinct deployment environment

**Missing Tests**:

- Container mode detection based on bind address (0.0.0.0)
- Container mode effects: mTLS disable, healthcheck compatibility, logging behavior
- Container mode validation interactions

### Health Check Compatibility

**Problem**: mTLS configuration broke healthchecks without detecting incompatibility

**Missing Tests**:

- Healthcheck endpoints (livez, readyz) should be accessible without client certs
- Container healthcheck commands should succeed (wget without TLS client cert)
- TLS client auth should NOT apply to health check endpoints

---

## Test Coverage Categories

### HAPPY Path Tests (Valid Configurations)

**1. Container Mode + SQLite**

- Bind: 0.0.0.0 (public), 127.0.0.1 (private)
- Database: sqlite://file::memory:?cache=shared
- DevMode: false
- Expected: mTLS disabled, validation passes, healthchecks work

**2. Container Mode + PostgreSQL**

- Bind: 0.0.0.0 (public), 127.0.0.1 (private)
- Database: postgres://...
- DevMode: false
- Expected: mTLS disabled, validation passes, healthchecks work

**3. Development Mode**

- Bind: 127.0.0.1 (both)
- Database: implicit SQLite (devMode=true)
- DevMode: true
- Expected: mTLS disabled, validation passes

**4. Production Mode**

- Bind: 127.0.0.1 or specific IP (NOT 0.0.0.0)
- Database: postgres://...
- DevMode: false
- Expected: mTLS enabled on private server, validation passes

### SAD Path Tests (Invalid Configurations)

**1. Dev Mode + 0.0.0.0 Binding**

- Should: Reject (Windows Firewall prevention)
- Error message: "bind address cannot be 0.0.0.0 in dev mode"

**2. Production + SQLite (Policy)**

- If policy disallows SQLite in production
- Should: Reject with clear error
- Error message: "SQLite not allowed in production (use PostgreSQL)"

**3. Empty/Invalid Database URL**

- Empty string, unsupported scheme, malformed URL
- Should: Reject with validation error

**4. Invalid Bind Addresses**

- Invalid IP format, out-of-range ports
- Should: Reject with validation error

**5. Missing Required Config Fields**

- Empty log level, missing TLS DNS names
- Should: Reject with validation error

---

## Service-Template Integration Strategy

**Goal**: Maximize test reusability across 9 product-services by placing tests in service-template where possible.

### Tests for Service Template (Reusable)

**Unit Tests** (`internal/apps/template/service/server/application/application_listener_test.go`):

- Container mode detection (0.0.0.0 â†’ isContainerMode)
- mTLS configuration (dev/container/production modes)
- TLS client auth logic (public vs private servers)

**Integration Tests** (`internal/apps/template/service/server/application/application_integration_test.go`):

- Config validation combinations (valid and invalid)
- TLS material generation for all modes (static, mixed, auto)
- Healthcheck endpoint accessibility

**Config Tests** (`internal/apps/template/service/config/config_loading_test.go`):

- YAML file loading with kebab-case field names
- Config precedence (file vs CLI vs env)
- Boolean/integer/string field parsing

### Tests for KMS-Specific Logic

**Unit Tests** (`internal/kms/server/repository/sqlrepository/sql_settings_mapper_test.go`):

- Database URL mapping (postgres://, sqlite://, dev mode)
- Container mode mapping (disabled, preferred, required)
- **ADD**: SQLite URL with query parameters
- **ADD**: Absolute file paths for SQLite

**Integration Tests** (`internal/kms/server/application/application_integration_test.go`):

- Complete config flow: load â†’ validate â†’ map DB â†’ initialize
- Container mode + SQLite integration
- Container mode + PostgreSQL integration

### Benefits of Service Template Tests

- **Reusability**: 1 test suite validates 9 services
- **Consistency**: All services use same patterns and validation logic
- **Maintainability**: Fix once, benefit everywhere
- **Coverage**: Higher effective coverage with less duplication

---

## Recommendations

### Immediate Actions (P0)

1. **Add mTLS Unit Tests** (CRITICAL - zero coverage for security code)
   - Test dev mode disables mTLS
   - Test container mode disables mTLS
   - Test production mode enables mTLS
   - Test public server never uses RequireAndVerifyClientCert

2. **Add Container Mode Detection Tests**
   - Test 0.0.0.0 detection on public address
   - Test 0.0.0.0 on private address (should NOT trigger container mode)
   - Test 127.0.0.1 (should NOT trigger container mode)

3. **Add YAML Config Loading Tests**
   - Load config from actual YAML file
   - Verify kebab-case â†’ PascalCase field mapping
   - Test boolean field parsing (dev-mode: true)

### Short-Term Actions (P1)

1. **Add Config Validation Integration Tests**
   - Valid combinations: container+SQLite, container+PostgreSQL, dev+SQLite, production+PostgreSQL
   - Invalid combinations: dev+0.0.0.0, missing required fields

2. **Add Database URL Parsing Tests**
   - SQLite with query parameters (?cache=shared&mode=rwc)
   - Absolute file paths (/var/lib/cryptoutil/db.sqlite)

3. **Add Healthcheck Integration Tests**
   - livez endpoint accessible without client cert
   - readyz endpoint dependency validation
   - Docker healthcheck commands (wget) succeed

### Long-Term Actions (P2)

1. **Add TLS Client Auth Integration Tests**
   - Container mode: wget healthcheck succeeds (no client cert)
   - Production mode: admin endpoints require client cert
   - Public endpoints never require client cert

2. **Add Config Precedence Tests**
   - YAML < Environment Variables < CLI Flags
   - Default value fallbacks
   - Config hot-reload (if supported)

3. **Add E2E Docker Tests**
   - Docker Compose stack startup
   - Healthcheck passes in containers
   - Service-to-service communication with/without mTLS

---

## Success Criteria

**Test Coverage Targets**:

- mTLS configuration logic: â‰¥95% coverage (currently 0%)
- Container mode detection: â‰¥95% coverage
- Config validation: â‰¥95% coverage
- Database URL mapping: â‰¥98% coverage (currently ~85%)

**Integration Test Goals**:

- All valid config combinations tested
- All invalid config combinations tested
- Container mode + SQLite integration verified
- Container mode + PostgreSQL integration verified

**Quality Gates**:

- All P0 tests implemented and passing
- Mutation testing â‰¥85% on affected modules
- No new TODOs or FIXMEs in test files
- All tests run in <15 seconds per package

**Workflow Impact**:

- DAST workflow should pass consistently
- Load testing workflow should pass consistently
- No config-related failures in CI/CD
