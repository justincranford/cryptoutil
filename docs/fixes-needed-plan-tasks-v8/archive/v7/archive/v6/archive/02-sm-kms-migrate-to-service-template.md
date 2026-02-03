# SM-KMS Migration to Service Template

## Document Purpose

This document provides a comprehensive migration plan to refactor `sm-kms` (Secrets Manager - Key Management Service) to leverage the service template pattern used by `cipher-im` and `jose-ja`.

## Executive Summary

The sm-kms service currently has a custom implementation that duplicates significant infrastructure code. Migration to the service template will:

- **Eliminate ~260+ lines** of duplicated infrastructure code per the ServerBuilder pattern
- **Add multi-tenancy support** (tenant isolation, realm management)
- **Add session management** (authentication, authorization)
- **Consolidate repositories** to GORM-only (removing 26 sqlrepository files)
- **Improve maintainability** through shared infrastructure updates

## Reference Implementations

### cipher-im Pattern (internal/apps/cipher/im/server/server.go)

```go
// Create server builder with template config.
builder := cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

// Register domain-specific migrations.
builder.WithDomainMigrations(cryptoutilAppsCipherImRepository.MigrationsFS, "migrations")

// Register domain-specific public routes.
builder.WithPublicRouteRegistration(func(
    base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
    res *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources,
) error {
    // Create repositories
    // Create public server with handlers
    // Register routes
    return nil
})

// Build complete service infrastructure.
resources, err := builder.Build()
```

### jose-ja Pattern (internal/apps/jose/ja/server/server.go)

Same pattern with domain-specific repositories (ElasticJWKRepository, MaterialJWKRepository, AuditConfigRepository, AuditLogRepository).

---

## Migration Phases Overview

| Phase | Name | Tasks | Est. LOE |
|-------|------|-------|----------|
| 1 | Repository Consolidation | Remove sqlrepository/, migrate to GORM-only | 3-5 days |
| 2 | Directory Structure Migration | Move from internal/kms/ to internal/apps/sm/kms/ | 1-2 days |
| 3 | ServerBuilder Integration | Replace custom server with template builder | 3-5 days |
| 4 | Multi-Tenancy & Session Support | Add tenant/realm/session management | 2-3 days |
| 5 | File Refactoring | Split large files to meet 500-line limit | 2-3 days |
| 6 | Test Coverage Improvements | Achieve ≥95% coverage target | 3-5 days |
| 7 | Deployment Updates | Update compose files, configs, Dockerfiles | 1-2 days |
| 8 | CI/CD Workflow Updates | Update workflows for new structure | 1 day |

**Total Estimated LOE**: 16-26 days

---

## Phase 1: Repository Consolidation

### Objective

Remove dual repository pattern (orm + sqlrepository) and consolidate to GORM-only per [03-04.database.instructions.md](.github/instructions/03-04.database.instructions.md).

### Current State

```
internal/kms/server/repository/
├── orm/           # 34 files - GORM models
└── sqlrepository/ # 26 files - Raw SQL (VIOLATION)
```

### Target State

```
internal/apps/sm/kms/repository/
├── models/        # GORM model definitions
├── migrations/    # SQL migrations (embedded fs.FS)
└── *.go          # Repository implementations (GORM-only)
```

### Tasks

#### Task 1.1: Analyze sqlrepository Dependencies

- [ ] List all sqlrepository functions and their callers
- [ ] Map each raw SQL operation to equivalent GORM operations
- [ ] Identify any raw SQL that cannot be converted (document blockers)

**Files to analyze**:
- `internal/kms/server/repository/sqlrepository/sql_barrier_keys.go`
- `internal/kms/server/repository/sqlrepository/sql_content_keys.go`
- `internal/kms/server/repository/sqlrepository/sql_elastic_keys.go`
- `internal/kms/server/repository/sqlrepository/sql_intermediate_keys.go`
- `internal/kms/server/repository/sqlrepository/sql_root_keys.go`
- `internal/kms/server/repository/sqlrepository/sql_unseal_keys.go`
- All 26 files total

#### Task 1.2: Create GORM Equivalents

- [ ] For each sqlrepository function, implement GORM equivalent in orm/
- [ ] Ensure cross-database compatibility (PostgreSQL + SQLite)
- [ ] Use `type:text` for UUIDs, `serializer:json` for arrays

**GORM Patterns Required**:
```go
// UUID handling (cross-DB compatible)
ID googleUuid.UUID `gorm:"type:text;primaryKey"`

// JSON arrays (cross-DB compatible)
AllowedScopes []string `gorm:"serializer:json"`

// Nullable UUIDs (use custom type, not pointer)
ClientProfileID NullableUUID `gorm:"type:text;index"`
```

#### Task 1.3: Update Business Logic Layer

- [ ] Change all businesslogic imports from sqlrepository to orm
- [ ] Update function signatures if needed
- [ ] Run tests to verify no regressions

#### Task 1.4: Remove sqlrepository Directory

- [ ] Delete all 26 files in `internal/kms/server/repository/sqlrepository/`
- [ ] Remove sqlrepository references from go.mod (if any)
- [ ] Update any remaining imports

#### Task 1.5: Verify GORM-Only Pattern

- [ ] Run `grep -r "database/sql" internal/kms/` should find NO direct usage
- [ ] Confirm all queries use GORM methods
- [ ] Run full test suite

---

## Phase 2: Directory Structure Migration

### Objective

Move sm-kms to `internal/apps/sm/kms/` to match project layout conventions.

### Current State

```
internal/kms/
├── client/
├── cmd/
└── server/
    ├── application/
    ├── businesslogic/
    ├── demo/
    ├── handler/
    ├── middleware/
    └── repository/
```

### Target State

```
internal/apps/sm/kms/
├── domain/        # Domain models (not GORM)
├── repository/    # GORM repositories
├── server/        # Server implementation
│   ├── apis/      # API handlers
│   ├── config/    # Configuration
│   └── server.go  # Main server using ServerBuilder
├── service/       # Business logic services
└── im.go          # Package entry point
```

### Tasks

#### Task 2.1: Create New Directory Structure

- [ ] Create `internal/apps/sm/kms/` directory
- [ ] Create subdirectories: domain/, repository/, server/, service/

#### Task 2.2: Move and Refactor Files

- [ ] Move repository models to `repository/models/`
- [ ] Move repository implementations to `repository/`
- [ ] Move handlers to `server/apis/`
- [ ] Move business logic to `service/`
- [ ] Move domain models to `domain/`

#### Task 2.3: Update Import Paths

- [ ] Update all import paths from `cryptoutil/internal/kms/` to `cryptoutil/internal/apps/sm/kms/`
- [ ] Update cmd/ references
- [ ] Update test imports

#### Task 2.4: Preserve Backward Compatibility

- [ ] Create re-export stubs in old location (temporary, for migration period)
- [ ] Add deprecation comments pointing to new location
- [ ] Plan removal of stubs in Phase 8

---

## Phase 3: ServerBuilder Integration

### Objective

Replace custom server implementation with service template ServerBuilder pattern.

### Current State

`internal/kms/server/application/application_listener.go` (1224 lines) duplicates:
- TLS configuration
- Fiber app setup
- Health check endpoints
- Middleware stack
- Admin server
- Public server

### Target State

`internal/apps/sm/kms/server/server.go` (~250 lines) using ServerBuilder:

```go
package server

import (
    "context"
    "fmt"

    "gorm.io/gorm"

    cryptoutilAppsSmKmsRepository "cryptoutil/internal/apps/sm/kms/repository"
    cryptoutilAppsSmKmsServerConfig "cryptoutil/internal/apps/sm/kms/server/config"
    cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
    cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
    cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"
    cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
    cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
    cryptoutilAppsTemplateServiceServerService "cryptoutil/internal/apps/template/service/server/service"
    cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
    cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// KMSServer represents the sm-kms service application.
type KMSServer struct {
    app *cryptoutilAppsTemplateServiceServer.Application
    db  *gorm.DB

    // Services from template.
    telemetryService      *cryptoutilSharedTelemetry.TelemetryService
    jwkGenService         *cryptoutilSharedCryptoJose.JWKGenService
    barrierService        *cryptoutilAppsTemplateServiceServerBarrier.Service
    sessionManagerService *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService
    realmService          cryptoutilAppsTemplateServiceServerService.RealmService

    // Domain-specific repositories.
    elasticKeyRepo      cryptoutilAppsSmKmsRepository.ElasticKeyRepository
    rootKeyRepo         cryptoutilAppsSmKmsRepository.RootKeyRepository
    intermediateKeyRepo cryptoutilAppsSmKmsRepository.IntermediateKeyRepository
    contentKeyRepo      cryptoutilAppsSmKmsRepository.ContentKeyRepository
    barrierKeyRepo      cryptoutilAppsSmKmsRepository.BarrierKeyRepository
    unsealKeyRepo       cryptoutilAppsSmKmsRepository.UnsealKeyRepository
    realmRepo           cryptoutilAppsTemplateServiceServerRepository.TenantRealmRepository
}

// NewFromConfig creates a new sm-kms server from KMSServerSettings.
func NewFromConfig(ctx context.Context, cfg *cryptoutilAppsSmKmsServerConfig.KMSServerSettings) (*KMSServer, error) {
    if ctx == nil {
        return nil, fmt.Errorf("context cannot be nil")
    } else if cfg == nil {
        return nil, fmt.Errorf("config cannot be nil")
    }

    // Create server builder with template config.
    builder := cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

    // Register sm-kms specific migrations.
    builder.WithDomainMigrations(cryptoutilAppsSmKmsRepository.MigrationsFS, "migrations")

    // Register sm-kms specific public routes.
    builder.WithPublicRouteRegistration(func(
        base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
        res *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources,
    ) error {
        // Create sm-kms specific repositories.
        elasticKeyRepo := cryptoutilAppsSmKmsRepository.NewElasticKeyRepository(res.DB)
        rootKeyRepo := cryptoutilAppsSmKmsRepository.NewRootKeyRepository(res.DB)
        intermediateKeyRepo := cryptoutilAppsSmKmsRepository.NewIntermediateKeyRepository(res.DB)
        contentKeyRepo := cryptoutilAppsSmKmsRepository.NewContentKeyRepository(res.DB)
        barrierKeyRepo := cryptoutilAppsSmKmsRepository.NewBarrierKeyRepository(res.DB)
        unsealKeyRepo := cryptoutilAppsSmKmsRepository.NewUnsealKeyRepository(res.DB)

        // Create public server with sm-kms handlers.
        publicServer, err := NewPublicServer(
            base,
            res.SessionManager,
            res.RealmService,
            elasticKeyRepo,
            rootKeyRepo,
            intermediateKeyRepo,
            contentKeyRepo,
            barrierKeyRepo,
            unsealKeyRepo,
            res.JWKGenService,
            res.BarrierService,
        )
        if err != nil {
            return fmt.Errorf("failed to create public server: %w", err)
        }

        // Register all routes.
        if err := publicServer.registerRoutes(); err != nil {
            return fmt.Errorf("failed to register public routes: %w", err)
        }

        return nil
    })

    // Build complete service infrastructure.
    resources, err := builder.Build()
    if err != nil {
        return nil, fmt.Errorf("failed to build sm-kms service: %w", err)
    }

    // ... create server struct with repositories
    return server, nil
}
```

### Tasks

#### Task 3.1: Create Config Structure

- [ ] Create `internal/apps/sm/kms/server/config/config.go`
- [ ] Embed `ServiceTemplateServerSettings` from template
- [ ] Add domain-specific config fields if needed

```go
package config

import (
    cryptoutilAppsTemplateServiceServerConfig "cryptoutil/internal/apps/template/service/server/config"
)

type KMSServerSettings struct {
    *cryptoutilAppsTemplateServiceServerConfig.ServiceTemplateServerSettings `yaml:",inline"`
    // Domain-specific settings here
}
```

#### Task 3.2: Create Server Struct

- [ ] Create `internal/apps/sm/kms/server/server.go`
- [ ] Follow jose-ja pattern exactly
- [ ] Include all accessor methods for tests

#### Task 3.3: Create Public Server

- [ ] Create `internal/apps/sm/kms/server/public_server.go`
- [ ] Move handler registration from application_listener.go
- [ ] Use PublicServerBase from template

#### Task 3.4: Create APIs Package

- [ ] Create `internal/apps/sm/kms/server/apis/` directory
- [ ] Move handlers from `internal/kms/server/handler/`
- [ ] Update handlers to use session context for tenant isolation

#### Task 3.5: Create Migration Files

- [ ] Create `internal/apps/sm/kms/repository/migrations/` directory
- [ ] Create SQL migrations starting at 2001 (domain-specific)
- [ ] Ensure template migrations (1001-1004) are preserved via merged FS

#### Task 3.6: Delete Old Infrastructure

- [ ] Delete `internal/kms/server/application/application_listener.go` (1224 lines)
- [ ] Delete `internal/kms/server/application/application_core.go` (replaced by builder)
- [ ] Delete custom TLS, middleware, health check code

---

## Phase 4: Multi-Tenancy & Session Support

### Objective

Add tenant isolation, realm management, and session management using template services.

### Current State

- No tenant isolation (single-tenant only)
- No realm support
- No session management
- Direct API access without authentication

### Target State

- Full multi-tenancy with tenant_id scoping all data
- Realm-based authentication context
- Session management with JWE/JWS/Opaque tokens
- Registration flow for tenant creation

### Tasks

#### Task 4.1: Add Tenant ID to All Domain Models

- [ ] Add `TenantID googleUuid.UUID` to ElasticKey model
- [ ] Add `TenantID` to RootKey, IntermediateKey, ContentKey models
- [ ] Add `TenantID` to BarrierKey, UnsealKey models
- [ ] Create database migrations for tenant_id columns

#### Task 4.2: Update Repository Queries

- [ ] Add tenant filtering to all repository methods
- [ ] Use session context to extract tenant_id
- [ ] Ensure no cross-tenant data access

```go
func (r *ElasticKeyRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*ElasticKey, error) {
    tenantID := cryptoutilAppsTemplateContext.GetTenantID(ctx)
    var key ElasticKey
    if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&key).Error; err != nil {
        return nil, err
    }
    return &key, nil
}
```

#### Task 4.3: Integrate Session Manager

- [ ] Use `SessionManagerService` from template
- [ ] Add session validation middleware to public routes
- [ ] Extract tenant context from session token

#### Task 4.4: Add Registration Endpoint

- [ ] Implement `/service/api/v1/auth/register` endpoint
- [ ] Support `create_tenant: true` for new tenant creation
- [ ] Return tenant_id, realm_id, session_token

#### Task 4.5: Update API Handlers

- [ ] Extract tenant_id from session context in all handlers
- [ ] Pass tenant_id to service layer
- [ ] Update error responses for tenant isolation violations

---

## Phase 5: File Refactoring

### Objective

Split large files to meet 500-line hard limit per [03-01.coding.instructions.md](.github/instructions/03-01.coding.instructions.md).

### Current Violations

| File | Lines | Status |
|------|-------|--------|
| application_listener.go | 1224 | 2.4x limit (DELETED in Phase 3) |
| businesslogic.go | 742 | 1.5x limit |

### Tasks

#### Task 5.1: Split businesslogic.go

Current `internal/kms/server/businesslogic/businesslogic.go` (742 lines) should become:

```
internal/apps/sm/kms/service/
├── elastic_key_service.go      # ElasticKey operations (~150 lines)
├── encryption_service.go       # Encrypt/Decrypt operations (~150 lines)
├── signing_service.go          # Sign/Verify operations (~150 lines)
├── key_rotation_service.go     # Key rotation logic (~100 lines)
├── barrier_service.go          # Barrier key operations (~100 lines)
└── service.go                  # Service interfaces and factory (~100 lines)
```

#### Task 5.2: Review All Files for Size

- [ ] Run `wc -l internal/apps/sm/kms/**/*.go | sort -n`
- [ ] Identify any files > 400 lines
- [ ] Plan refactoring for files approaching limits

---

## Phase 6: Test Coverage Improvements

### Objective

Achieve ≥95% coverage for production code, ≥98% for infrastructure/utility.

### Current Coverage

| Package | Coverage | Target | Gap |
|---------|----------|--------|-----|
| client | 74.9% | 95% | -20.1% |
| cmd | 0.0% | 95% | -95.0% |
| server/application | 77.1% | 95% | -17.9% |
| server/businesslogic | 39.0% | 95% | -56.0% |
| server/demo | 7.3% | 95% | -87.7% |
| server/handler | 79.9% | 95% | -15.1% |
| server/middleware | 53.1% | 95% | -41.9% |
| server/repository/orm | 88.9% | 95% | -6.1% |
| server/repository/sqlrepository | 78.0% | N/A | (DELETED) |

### Tasks

#### Task 6.1: Implement TestMain Pattern

- [ ] Create `internal/apps/sm/kms/server/testmain_test.go`
- [ ] Set up shared test database and server
- [ ] Register test tenant via registration flow

```go
var (
    testServer       *server.KMSServer
    testTenantID     googleUuid.UUID
    testRealmID      googleUuid.UUID
    testSessionToken string
)

func TestMain(m *testing.M) {
    ctx := context.Background()

    // Create server with test configuration
    cfg := config.NewTestSettings()
    testServer, _ = server.NewFromConfig(ctx, cfg)

    // Start server
    go testServer.Start(ctx)
    defer testServer.Shutdown(ctx)

    // Register test tenant through API
    resp := registerTestUser(testServer.PublicBaseURL())
    testTenantID = resp.TenantID
    testRealmID = resp.RealmID
    testSessionToken = resp.SessionToken

    os.Exit(m.Run())
}
```

#### Task 6.2: Add Table-Driven Tests

- [ ] Convert standalone test functions to table-driven
- [ ] Add `t.Parallel()` to all tests and subtests
- [ ] Use UUIDv7 for dynamic test data

#### Task 6.3: Add Handler Tests with app.Test()

- [ ] Use Fiber's `app.Test()` for in-memory handler testing
- [ ] Test happy path and error cases
- [ ] Never start real HTTPS listeners in tests

#### Task 6.4: Target RED Lines in Coverage

- [ ] Generate HTML coverage: `go test -coverprofile=coverage.out && go tool cover -html=coverage.out`
- [ ] Identify uncovered code paths
- [ ] Write targeted tests for RED lines

#### Task 6.5: Add internalMain Pattern for cmd/

- [ ] Create testable `internalMain()` function
- [ ] Thin `main()` delegates to `internalMain()`
- [ ] Test `internalMain()` with injected dependencies

---

## Phase 7: Deployment Updates

### Objective

Update Docker Compose, configs, and Dockerfiles for new structure.

### Tasks

#### Task 7.1: Update Dockerfile

- [ ] Update build paths to new location
- [ ] Ensure multi-stage build pattern
- [ ] Add secrets validation stage

#### Task 7.2: Update compose.yml

- [ ] Update service paths
- [ ] Fix healthcheck path: `/admin/api/v1/livez` (not `/admin/v1/livez`)
- [ ] Ensure Docker secrets pattern for all credentials

#### Task 7.3: Update Configuration Files

- [ ] Create new config structure matching `KMSServerSettings`
- [ ] Embed `ServiceTemplateServerSettings`
- [ ] Update YAML paths in compose

#### Task 7.4: Update Secrets Directory

- [ ] Ensure all secrets use Docker secrets pattern
- [ ] Verify 440 permissions on .secret files
- [ ] No inline credentials in compose.yml

---

## Phase 8: CI/CD Workflow Updates

### Objective

Update workflows for new directory structure.

### Tasks

#### Task 8.1: Update ci-e2e.yml

- [ ] Update paths to new location
- [ ] Update test commands
- [ ] Verify E2E tests pass

#### Task 8.2: Update ci-load.yml

- [ ] Update Gatling simulation paths if needed
- [ ] Verify load tests target correct endpoints

#### Task 8.3: Update ci-coverage.yml

- [ ] Update coverage paths
- [ ] Verify coverage thresholds enforced

#### Task 8.4: Remove Backward Compatibility Stubs

- [ ] Remove re-export stubs from old location
- [ ] Update any remaining references
- [ ] Verify all workflows pass

---

## Validation Checklist

### Phase Completion Gates

Each phase MUST pass ALL checks before proceeding:

- [ ] `go build ./...` - No errors
- [ ] `golangci-lint run` - No warnings
- [ ] `go test ./internal/apps/sm/kms/...` - All pass
- [ ] Coverage ≥95% for production code
- [ ] No files > 500 lines
- [ ] Git commit with conventional message

### Final Validation

- [ ] Full test suite passes: `go test ./...`
- [ ] E2E tests pass: `go test -tags=e2e ./...`
- [ ] Load tests pass: `mvn gatling:test`
- [ ] Docker Compose starts successfully
- [ ] All workflows pass in CI/CD

---

## Risk Assessment

### High Risk

| Risk | Mitigation |
|------|------------|
| Repository migration breaks existing functionality | Comprehensive test coverage before migration |
| Multi-tenancy breaks existing API contracts | Version API endpoints, maintain backward compatibility |
| Large file refactoring introduces bugs | Incremental refactoring with tests at each step |

### Medium Risk

| Risk | Mitigation |
|------|------------|
| Directory move breaks imports | Use IDE refactoring tools, verify all imports |
| Config structure changes break deployments | Test in staging environment first |

### Low Risk

| Risk | Mitigation |
|------|------------|
| CI/CD workflow path changes | Update workflows incrementally |

---

## Success Criteria

Migration is complete when:

1. ✅ All code in `internal/apps/sm/kms/` (not `internal/kms/`)
2. ✅ Using ServerBuilder pattern (not custom infrastructure)
3. ✅ GORM-only repositories (no sqlrepository/)
4. ✅ Multi-tenancy with tenant_id scoping
5. ✅ Session management with registration flow
6. ✅ All files ≤500 lines
7. ✅ Coverage ≥95% production, ≥98% infrastructure
8. ✅ All CI/CD workflows passing
9. ✅ E2E and load tests passing
10. ✅ Docker Compose deployment working

---

## References

- [02-02.service-template.instructions.md](.github/instructions/02-02.service-template.instructions.md)
- [03-08.server-builder.instructions.md](.github/instructions/03-08.server-builder.instructions.md)
- [03-04.database.instructions.md](.github/instructions/03-04.database.instructions.md)
- [07-01.testmain-integration-pattern.instructions.md](.github/instructions/07-01.testmain-integration-pattern.instructions.md)
- [internal/apps/cipher/im/server/server.go](internal/apps/cipher/im/server/server.go) - Reference implementation
- [internal/apps/jose/ja/server/server.go](internal/apps/jose/ja/server/server.go) - Reference implementation
