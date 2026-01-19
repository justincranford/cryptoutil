# JOSE-JA Refactoring Tasks V4

**Last Updated**: 2026-01-18
**Based On**: JOSE-JA-REFACTORING-PLAN-V4.md
**Supersedes**: JOSE-JA-REFACTORING-TASKS-V3.md (archived to docs/archive/jose-ja/)

**Architecture Reference**: See [docs/arch/ARCHITECTURE.md](../arch/ARCHITECTURE.md) for comprehensive design patterns, principles, and implementation guidelines.

**Critical Fixes from V3**:
- ✅ Port 9090 for admin endpoints
- ✅ PostgreSQL 18+ requirement
- ✅ Directory: deployments/jose-ja/, configs/jose-ja/
- ✅ Docker secrets > YAML > ENV priority
- ✅ Separate browser/service session configs
- ✅ OTLP only (no Prometheus scraping)
- ✅ Consistent paths: /admin/api/v1
- ✅ No service name in paths
- ✅ Realms authn only (no realm_id filtering)
- ✅ No hardcoded passwords
- ✅ key_type/key_size implied by algorithm
- ✅ All requirements mandatory (no deferrals)

---

## MANDATORY Execution Rules

**Quality Gates (EVERY task MUST pass ALL before marking complete)**:
1. ✅ **Build**: `go build ./...` (zero errors)
2. ✅ **Linting**: `golangci-lint run --fix ./...` (zero warnings)
3. ✅ **Tests**: `go test ./...` (100% pass, no skips without tracking)
4. ✅ **Coverage**: ≥95% production code, ≥98% infrastructure/utility code
5. ✅ **Mutation**: `gremlins unleash ./internal/[package]` ≥85% production, ≥98% infrastructure
6. ✅ **Evidence**: Objective proof of completion (build output, test output, coverage report, mutation score, commit hash)
7. ✅ **Git**: Conventional commit after EACH logical unit with evidence in commit message

**Continuous Execution (NO EXCEPTIONS)**:
- ❌ NEVER stop to ask "Should I continue with Task X?"
- ❌ NEVER pause between tasks for status updates
- ❌ NEVER skip validation steps to save time
- ❌ NEVER mark tasks complete without running ALL 7 quality gates
- ❌ NEVER defer mutation testing (run per package during implementation)
- ✅ ALWAYS commit after each task completion with evidence
- ✅ ALWAYS start next task immediately after commit (zero pause)
- ✅ ALWAYS update specs/002-cryptoutil/implement/DETAILED.md Section 2 timeline

**Evidence Requirements (NO task complete without ALL)**:
- **Build output**: `go build ./...` zero errors
- **Test output**: `go test ./...` 100% pass
- **Coverage report**: ≥95%/≥98% targets met
- **Mutation score**: ≥85%/≥98% targets met (when applicable)
- **Git commit hash**: Conventional message with evidence

---

## Phase 0: Service-Template - Remove Default Tenant Pattern

### 0.1 Remove WithDefaultTenant from ServerBuilder
**File**: `internal/apps/template/service/server/builder/server_builder.go`

- [ ] 0.1.1 Remove `defaultTenantID` field from ServerBuilder struct
- [ ] 0.1.2 Remove `defaultRealmID` field from ServerBuilder struct
- [ ] 0.1.3 Remove `WithDefaultTenant(tenantID, realmID)` method
- [ ] 0.1.4 Remove call to `ensureDefaultTenant()` in `Build()` method
- [ ] 0.1.5 Remove `ensureDefaultTenant()` helper method
- [ ] 0.1.6 Remove passing defaultTenantID/defaultRealmID to SessionManagerService
- [ ] 0.1.7 Run `golangci-lint run --fix`
- [ ] 0.1.8 Verify `go build ./internal/apps/template/...` succeeds

**Evidence**: Build succeeds, WithDefaultTenant method removed

---

### 0.2 Remove EnsureDefaultTenant Helper
**File**: `internal/apps/template/service/server/repository/seeding.go`

- [ ] 0.2.1 Delete entire file `seeding.go`
- [ ] 0.2.2 Run `golangci-lint run --fix`
- [ ] 0.2.3 Verify `go build ./internal/apps/template/...` succeeds

**Evidence**: File deleted, build succeeds

---

### 0.3 Update SessionManagerService
**File**: `internal/apps/template/service/server/businesslogic/session_manager_service.go`

- [ ] 0.3.1 Remove `defaultTenantID` field
- [ ] 0.3.2 Remove `defaultRealmID` field
- [ ] 0.3.3 Remove `IssueBrowserSession(ctx, userID)` method
- [ ] 0.3.4 Remove `ValidateBrowserSession(ctx, token)` method
- [ ] 0.3.5 Remove `IssueServiceSession(ctx, clientID)` method
- [ ] 0.3.6 Remove `ValidateServiceSession(ctx, token)` method
- [ ] 0.3.7 Update constructor to remove defaultTenantID/defaultRealmID params
- [ ] 0.3.8 Run `golangci-lint run --fix`
- [ ] 0.3.9 Verify `go build ./internal/apps/template/...` succeeds

**Evidence**: Single-tenant methods removed

---

### 0.4 Remove Template Magic Constants
**Files**: `internal/shared/magic/magic_template.go`

- [ ] 0.4.1 Remove `TemplateDefaultTenantID` constant (if exists)
- [ ] 0.4.2 Remove `TemplateDefaultRealmID` constant (if exists)
- [ ] 0.4.3 Verify `grep -r "TemplateDefaultTenantID" internal/` returns 0
- [ ] 0.4.4 Verify `grep -r "TemplateDefaultRealmID" internal/` returns 0
- [ ] 0.4.5 Run `golangci-lint run --fix`

**Evidence**: Magic constants removed, grep shows no usage

---

### 0.5 REMOVED - pending_users table is sufficient (per Q5.1)

**User Decision**: "WTF is tenant_join_requests (1006)? Only pending_users (1005) needed?"

**Implementation Details**:
- Username unique per tenant across pending_users AND users
- Composite index (username, tenant_id), status+requested_at index
- Expiration in HOURS (configurable, default 72h), auto-delete expired
- NO email validation on username (accepts any string)
- DOWN migrations for dev/test rollback

---

### 0.6 REMOVED - TenantJoinRequestRepository not needed

**User Decision**: Removed with migration 1005

---

### 0.7 REMOVED - Tenant Registration Service simplified

**User Decision**: Simplified to use pending_users table only

---

### 0.8 Create Registration HTTP Handlers
**Files**: `internal/apps/template/service/server/apis/{registration,join_request}_handlers.go`

**Implementation Details**:
- Admin dashboard in template infrastructure (NOT domain-specific)
- NO email notifications (users poll via login)
- NO webhook callbacks (keep simple)
- NO unauthenticated status API (poll via login: 403=pending, 401=rejected)
- Rate limiting per IP only (10 registrations/hour)
- In-memory rate limiting (sync.Map, single-node)
- Configurable thresholds with low defaults

- [ ] 0.8.1 Implement POST /browser/api/v1/auth/register (user registration)
- [ ] 0.8.2 Implement POST /service/api/v1/auth/register (client registration)
- [ ] 0.8.3 Implement GET /admin/api/v1/join-requests (list join requests)
- [ ] 0.8.4 Implement PUT /admin/api/v1/join-requests/:id (approve/reject)
- [ ] 0.8.5 **CRITICAL: Consistent paths (/admin/api/v1, NOT /admin/v1)**
- [ ] 0.8.6 **CRITICAL: tenant_id param (absence=create, presence=join)**
- [ ] 0.8.7 **NEW: In-memory rate limiting per IP (configurable threshold)**
- [ ] 0.8.8 Write integration tests (≥95% coverage)

**Evidence**: Coverage ≥95%, all endpoints tested

---

### 0.9 Update ServerBuilder Registration
**File**: `internal/apps/template/service/server/builder/server_builder.go`

- [ ] 0.9.1 Register POST /browser/api/v1/auth/register route
- [ ] 0.9.2 Register POST /service/api/v1/auth/register route
- [ ] 0.9.3 Register GET /admin/api/v1/join-requests route
- [ ] 0.9.4 Register PUT /admin/api/v1/join-requests/:id route
- [ ] 0.9.5 Verify no WithDefaultTenant() calls remain

**Evidence**: Routes registered, E2E tests pass

---

### 0.10 Phase 0 Validation

- [ ] 0.10.1 Build: `go build ./...` (zero errors)
- [ ] 0.10.2 Linting: `golangci-lint run ./...` (zero warnings)
- [ ] 0.10.3 Tests: `go test ./internal/apps/template/... -cover` (100% pass)
- [ ] 0.10.4 Coverage: ≥95% production, ≥98% infrastructure
- [ ] 0.10.5 Mutation: ≥85% production, ≥98% infrastructure
- [ ] 0.10.6 E2E: Registration flow works (browser + service)
- [ ] 0.10.7 E2E: Join request flow works (create, list, approve)
- [ ] 0.10.8 Security: NO hardcoded passwords in tests
- [ ] 0.10.9 Paths: Consistent /admin/api/v1, /service/api/v1, /browser/api/v1
- [ ] 0.10.10 Git: Conventional commit with evidence

**Final Commit**: `feat(service-template): remove default tenant pattern, implement registration flow`

---

## Phase 1: Cipher-IM - Adapt to Registration Flow

### 1.1 Remove cipher-im Default Tenant References
**Files**: `internal/apps/cipher/im/server/*`

- [ ] 1.1.1 Remove any WithDefaultTenant() calls (if exist)
- [ ] 1.1.2 Verify `grep -r "WithDefaultTenant" internal/apps/cipher/` returns 0
- [ ] 1.1.3 Verify all tests use registerUser() for tenant creation

**Evidence**: Grep shows 0 WithDefaultTenant usages

---

### 1.2 Update cipher-im Tests to Registration Pattern
**Files**: `internal/apps/cipher/im/server/apis/*_test.go`

**Hash Service Configuration**:
- Q4.1: Verify PBKDF2 iterations = 610,000 in `internal/shared/magic/magic_cryptography.go`
- Q4.2: Lazy migration for pepper rotation (already implemented in hash service)
- Q4.3: Multiple hash versions supported (already implemented in hash service)
- Q4.4: Global security policy (NOT per-tenant configuration)

- [ ] 1.2.1 Add TestMain pattern for per-package tenant setup
- [ ] 1.2.2 Use registerUser() with cryptoutilMagic.TestPassword
- [ ] 1.2.3 **CRITICAL: NO hardcoded passwords ("pass1", "pass2")**
- [ ] 1.2.4 **NEW: Verify hash service uses 610,000 PBKDF2 iterations**
- [ ] 1.2.5 Verify all tests pass

**Evidence**: All tests pass, NO hardcoded passwords

---

### 1.3 Phase 1 Validation

- [ ] 1.3.1 Build: `go build ./internal/apps/cipher/...` (zero errors)
- [ ] 1.3.2 Linting: `golangci-lint run ./internal/apps/cipher/...` (zero warnings)
- [ ] 1.3.3 Tests: `go test ./internal/apps/cipher/... -cover` (100% pass)
- [ ] 1.3.4 Coverage: ≥95% production, ≥98% infrastructure
- [ ] 1.3.5 Grep: 0 WithDefaultTenant usages
- [ ] 1.3.6 Security: NO hardcoded passwords
- [ ] 1.3.7 Git: Conventional commit

**Final Commit**: `test(cipher-im): adapt to registration flow pattern`

---

## Phase 2: JOSE-JA - Database Schema Migration

### 2.0 Prerequisites

- [ ] 2.0.1 Verify migration numbering ranges (template 1001-1999, JOSE 2001+)
- [ ] 2.0.2 Verify no conflicts with existing migrations
- [ ] 2.0.3 Document migration range allocation in commit message

**Evidence**: Migration ranges verified, no conflicts

---

### 2.1 Create JOSE Domain Models
**File**: `internal/apps/jose/ja/domain/models.go`

- [ ] 2.1.1 Create ElasticJWK model (with TenantID, NO realm_id)
- [ ] 2.1.2 Create MaterialKey model
- [ ] 2.1.3 Create JWKSConfig model (with AllowCrossTenant field)
- [ ] 2.1.4 Create AuditConfig model
- [ ] 2.1.5 Create AuditLog model (with SessionID field)
- [ ] 2.1.6 **CRITICAL: ALL models include TenantID**

**Evidence**: Build succeeds, all models have TenantID

---

### 2.2 Create JOSE Database Migrations
**Directory**: `internal/apps/jose/ja/repository/migrations/`

- [ ] 2.2.1 Create 2001_elastic_jwk.{up,down}.sql
- [ ] 2.2.2 Create 2002_material_keys.{up,down}.sql
- [ ] 2.2.3 Create 2003_jwks_config.{up,down}.sql
- [ ] 2.2.4 Create 2004_audit_config.{up,down}.sql
- [ ] 2.2.5 Create 2005_audit_log.{up,down}.sql
- [ ] 2.2.6 **CRITICAL: Use TEXT for UUIDs, TIMESTAMP for dates**
- [ ] 2.2.7 Test migrations on PostgreSQL 18+
- [ ] 2.2.8 Test migrations on SQLite 3.19+

**Evidence**: Migrations apply to both databases

---

### 2.3 Implement JOSE Repositories
**Files**: `internal/apps/jose/ja/repository/*_repository.go`

- [ ] 2.3.1 Implement ElasticJWKRepository (Create, GetByID, GetByKID, List, Update)
- [ ] 2.3.2 Implement MaterialKeyRepository
- [ ] 2.3.3 Implement JWKSConfigRepository
- [ ] 2.3.4 Implement AuditConfigRepository
- [ ] 2.3.5 Implement AuditLogRepository
- [ ] 2.3.6 **CRITICAL: Filter by tenant_id ONLY (NOT realm_id)**
- [ ] 2.3.7 Write unit tests (≥98% coverage)
- [ ] 2.3.8 Run mutation testing (≥98% score)

**Evidence**: Coverage ≥98%, mutation ≥98%, NO realm_id filtering

---

### 2.4 Phase 2 Validation

- [ ] 2.4.1 Build: `go build ./internal/apps/jose/...`
- [ ] 2.4.2 Linting: `golangci-lint run ./internal/apps/jose/...`
- [ ] 2.4.3 Tests: `go test ./internal/apps/jose/ja/repository/... -cover` (100% pass)
- [ ] 2.4.4 Coverage: ≥98% (infrastructure)
- [ ] 2.4.5 Mutation: ≥98% (gremlins)
- [ ] 2.4.6 Migrations: Apply to PostgreSQL 18+ and SQLite
- [ ] 2.4.7 Repository: NO realm_id filtering in WHERE clauses
- [ ] 2.4.8 Git: Conventional commit

**Final Commit**: `feat(jose-ja): implement database schema and repositories`

---

## Phase 3: JOSE-JA - ServerBuilder Integration

### 3.1 Create JOSE Server Configuration
**File**: `internal/apps/jose/ja/server/config/config.go`

- [ ] 3.1.1 Create Settings struct (wraps ServiceTemplateServerSettings)
- [ ] 3.1.2 **CRITICAL: Separate browser-session-* and service-session-* configs**
- [ ] 3.1.3 **CRITICAL: Docker secrets > YAML > ENV priority**
- [ ] 3.1.4 Write config loading tests

**Evidence**: Config loads correctly, priority order verified

---

### 3.2 Create JOSE Public Server
**File**: `internal/apps/jose/ja/server/server.go`

- [ ] 3.2.1 Create JoseServer struct
- [ ] 3.2.2 Implement NewFromConfig() using ServerBuilder
- [ ] 3.2.3 Register domain migrations (2001-2005)
- [ ] 3.2.4 Register domain routes
- [ ] 3.2.5 **CRITICAL: Paths /service/api/v1/* (NO /service/api/v1/jose/*)**
- [ ] 3.2.6 **CRITICAL: Paths /admin/api/v1/* (NOT /admin/v1/*)**

**Evidence**: Server starts, routes registered correctly

---

### 3.3 Create JOSE HTTP Handlers
**Files**: `internal/apps/jose/ja/server/apis/*_handlers.go`

- [ ] 3.3.1 Implement JWK handlers (Generate, List, Get, Rotate, Revoke)
- [ ] 3.3.2 Implement JWS handlers (Sign, Verify)
- [ ] 3.3.3 Implement JWE handlers (Encrypt, Decrypt)
- [ ] 3.3.4 Implement JWT handlers (Issue, Validate)
- [ ] 3.3.5 Implement JWKS handlers (GetJWKS)
- [ ] 3.3.6 Implement Audit handlers (GetConfig, SetConfig, ListLogs)
- [ ] 3.3.7 **CRITICAL: Simplify Generate request (remove key_type, key_size)**
- [ ] 3.3.8 Write handler tests (≥95% coverage)

**Evidence**: Coverage ≥95%, all endpoints tested

---

### 3.4 Implement JOSE Business Logic Services
**Files**: `internal/apps/jose/ja/service/*_service.go`

- [ ] 3.4.1 Implement ElasticJWKService
- [ ] 3.4.2 Implement MaterialRotationService
- [ ] 3.4.3 Implement JWSService
- [ ] 3.4.4 Implement JWEService
- [ ] 3.4.5 Implement JWTService
- [ ] 3.4.6 Implement JWKSService
- [ ] 3.4.7 Implement AuditLogService
- [ ] 3.4.8 Write service tests (≥95% coverage)
- [ ] 3.4.9 Run mutation testing (≥85% score)

**Evidence**: Coverage ≥95%, mutation ≥85%

---

### 3.5 Phase 3 Validation

- [ ] 3.5.1 Build: `go build ./internal/apps/jose/...`
- [ ] 3.5.2 Linting: `golangci-lint run ./internal/apps/jose/...`
- [ ] 3.5.3 Tests: `go test ./internal/apps/jose/... -cover` (100% pass)
- [ ] 3.5.4 Coverage: ≥95% production, ≥98% infrastructure
- [ ] 3.5.5 Mutation: ≥85% production, ≥98% infrastructure
- [ ] 3.5.6 Paths: No service name in request paths
- [ ] 3.5.7 Config: Docker secrets > YAML > ENV priority
- [ ] 3.5.8 Git: Conventional commit

**Final Commit**: `feat(jose-ja): integrate ServerBuilder pattern`

---

## Phase 4-8: JOSE-JA Implementation (Continued)

**See JOSE-JA-REFACTORING-PLAN-V4.md for detailed tasks**

**Key Changes from V3**:
- ✅ Phase 4: Fix repository realm_id filtering, test passwords
- ✅ Phase 5: Cross-tenant JWKS via tenant management API
- ✅ Phase 6: No changes
- ✅ Phase 7: Path migration (remove /jose/ from paths)
- ✅ Phase 8: TestMain pattern, NO hardcoded passwords

---

## Phase 9: JOSE-JA - Documentation

### 9.1 Update API Documentation
**File**: `docs/jose-ja/API-REFERENCE.md`

- [ ] 9.1.1 Fix base URLs (port 9090 for admin)
- [ ] 9.1.2 Remove /jose/ from all request paths
- [ ] 9.1.3 Simplify Generate request (remove key_type, key_size)
- [ ] 9.1.4 Update all endpoint examples
- [ ] 9.1.5 Document tenant_id parameter (absence=create, presence=join)
- [ ] 9.1.6 Document join request endpoints

**Evidence**: API docs updated, all examples correct

---

### 9.2 Update Deployment Guide
**File**: `docs/jose-ja/DEPLOYMENT.md`

- [ ] 9.2.1 Fix port 9090 for admin endpoints
- [ ] 9.2.2 Update PostgreSQL requirement to 18+
- [ ] 9.2.3 Fix directory structure (deployments/jose-ja/, configs/jose-ja/)
- [ ] 9.2.4 **CRITICAL: Remove ENV variable examples**
- [ ] 9.2.5 **CRITICAL: Document Docker secrets > YAML priority**
- [ ] 9.2.6 **CRITICAL: Remove Kubernetes documentation**
- [ ] 9.2.7 **CRITICAL: Remove Prometheus scraping endpoint**
- [ ] 9.2.8 **CRITICAL: OTLP telemetry only**
- [ ] 9.2.9 Separate browser-session-* and service-session-* configs
- [ ] 9.2.10 Document health endpoints on BOTH public and admin servers

**Evidence**: Deployment docs updated, NO ENVs, NO K8s, OTLP only

---

### 9.3 Update Copilot Instructions
**File**: `.github/instructions/02-02.service-template.instructions.md` (or relevant file)

- [ ] 9.3.1 Document Docker secrets > YAML > ENV > CLI priority
- [ ] 9.3.2 Document consistent API paths (/admin/api/v1, /service/api/v1, /browser/api/v1)
- [ ] 9.3.3 Document NO service name in paths
- [ ] 9.3.4 Document realms are authn only (NO data scope filtering)
- [ ] 9.3.5 Document NO hardcoded passwords in tests
- [ ] 9.3.6 Document tenant_id parameter pattern

**Evidence**: Copilot instructions updated

---

### 9.4 Final Cleanup

- [ ] 9.4.1 Remove all TODOs: `grep -r "TODO" internal/jose/` (0 legitimate TODOs remain)
- [ ] 9.4.2 Run `golangci-lint run ./...` (zero warnings)
- [ ] 9.4.3 Run all tests: `go test ./...` (all pass)
- [ ] 9.4.4 Verify coverage targets met (≥95% production, ≥98% infrastructure)
- [ ] 9.4.5 Verify mutation scores met (≥85% production, ≥98% infrastructure)

**Evidence**: TODOs removed, linting clean, all tests pass, quality gates met

---

### 9.5 Phase 9 Validation

- [ ] 9.5.1 Verify all documentation complete
- [ ] 9.5.2 Verify no deprecated code remains
- [ ] 9.5.3 Verify all quality gates pass
- [ ] 9.5.4 Git commit: `git commit -m "docs(jose-ja): update documentation (V4)"`

**Evidence**: All docs updated, quality gates pass

---

## Final Project Validation

- [ ] All 3 phases complete (service-template, cipher-im, jose-ja)
- [ ] Zero build errors across entire project
- [ ] Zero linting warnings across entire project
- [ ] All tests pass (unit, integration, E2E)
- [ ] Coverage targets met (≥95% production, ≥98% infrastructure)
- [ ] Mutation scores met (≥85% production, ≥98% infrastructure)
- [ ] Documentation complete (API-REFERENCE.md, DEPLOYMENT.md)
- [ ] Copilot instructions updated
- [ ] Git history clean (conventional commits)

**Final Commit**: `git commit -m "docs(jose-ja): complete V4 refactoring"`

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
