# Service Template Application Listener - Implementation Guide

## Overview

This guide explains the new `ApplicationListener` pattern for service-template and cipher-im, modeled after the clean encapsulation in `internal/kms/server/application/application_listener.go`.

**Goal**: Consistent, simple service startup across all cryptoutil products with minimal TestMain boilerplate.

## Core Concepts

### Problem Statement

**Before**: Each TestMain had messy, inconsistent startup code:

- Different patterns for creating telemetry, JWK gen, barrier service
- Duplicate TLS certificate generation
- Manual server creation with different parameter orders
- No standard health check or shutdown patterns

**After**: Unified ApplicationListener provides:

- Single entry point: `StartApplicationListener(ctx, cfg)`
- Consistent health checks: `SendLivenessCheck(settings)`, `SendReadinessCheck(settings)`
- Graceful shutdown: `SendShutdownRequest(settings)` or `listener.Shutdown()`
- Reusable across ALL services (cipher-im, jose-ja, identity-*, sm-kms, pki-ca)

### Architecture Layers

The service template provides a layered application architecture with three distinct levels:

#### 1. ApplicationBasic

**Foundation layer** providing basic service infrastructure:

- **Telemetry Service**: Logging, metrics, tracing (OpenTelemetry)
- **Unseal Keys Service**: Key encryption at rest (HKDF-based key derivation)
- **JWK Generation Service**: Cryptographic key generation (RSA, ECDSA, EdDSA)

**Usage**: Internal dependency for ApplicationCore, not used directly by services.

#### 2. ApplicationCore

**Database layer** extending ApplicationBasic with database provisioning:

- **Automatic Database Management**: SQLite in-memory, PostgreSQL testcontainer, or external DB
- **Configuration-Driven**: Uses `ServerSettings.DatabaseURL` and `ServerSettings.DatabaseContainer`
- **Graceful Shutdown**: Cleanup of database containers and connections

**Database Provisioning Modes**:

| DatabaseURL | DatabaseContainer | Result |
|-------------|------------------|--------|
| `file::memory:?cache=shared` | Any | SQLite in-memory (tests) |
| `postgres://...` | `disabled` | External PostgreSQL |
| `postgres://...` | `required` | PostgreSQL testcontainer (fails if unavailable) |
| `postgres://...` | `preferred` | PostgreSQL testcontainer with fallback to external |

**Usage**: Internal dependency for ApplicationListener, provides `Core.DB` (GORM instance).

#### 3. ApplicationListener

**Top-level orchestrator** managing HTTP servers and core infrastructure:

- **Public Server**: Business logic (APIs, UI, external clients)
- **Admin Server**: Health checks, graceful shutdown (Kubernetes probes)
- **Lifecycle Management**: Start/shutdown both servers concurrently

**Usage**: Public API for services (replace manual server management).

```
ApplicationListener (Top-Level Orchestrator)
    ├── ApplicationConfig (product-specific injection)
    │   ├── ServerSettings (common: bind, TLS, OTLP)
    │   ├── Database (test-container OR production pool)
    │   ├── PublicHandlers (product-specific routes)
    │   └── AdminHandlers (optional: barrier rotation, diagnostics)
    │
    ├── ApplicationCore (Database Layer)
    │   ├── Database Provisioning (SQLite/PostgreSQL/External)
    │   └── Connection Management
    │
    ├── ApplicationBasic (Foundation Layer)
    │   ├── Telemetry (OTLP, structured logging)
    │   ├── JWK Generation (crypto key pools)
    │   └── Barrier Service (key encryption at rest)
    │
    └── Application (Dual Servers)
        ├── Public Server (business APIs, browser UI)
        └── Admin Server (health checks, shutdown)
```

## Implementation Roadmap

### Phase 1: Core Infrastructure (Complete ✅)

**File**: `internal/template/server/listener/application_listener.go`

**Status**: Created with interfaces and TODO markers

**Components**:

- ✅ `ApplicationListener` struct
- ✅ `ApplicationConfig` injection point
- ✅ `HandlerRegistration` function type
- ✅ Health check functions (liveness, readiness)
- ✅ Shutdown patterns

**TODOs Remaining**:

1. Public server factory integration (product-specific)
2. Admin server creation from template
3. Application startup orchestration
4. Actual port extraction after dynamic allocation

### Phase 2: Product-Specific Server Factories

Each product service needs a factory function matching this signature:

```go
// Example: cipher-im public server factory
func NewPublicServerFromConfig(
    ctx context.Context,
    cfg *ApplicationConfig,
    template *cryptoutilTemplateServer.ServiceTemplate,
) (cryptoutilTemplateServer.IPublicServer, error) {
    // 1. Create repositories using template.DB()
    userRepo := repository.NewUserRepository(cfg.DB)
    messageRepo := repository.NewMessageRepository(cfg.DB)

    // 2. Generate TLS config for this server instance
    tlsCfg, err := tlsGenerator.GenerateAutoTLSGeneratedSettings(...)

    // 3. Create public server with product-specific parameters
    publicServer, err := NewPublicServer(
        ctx,
        int(cfg.ServerSettings.BindPublicPort),
        userRepo,
        messageRepo,
        template.JWKGen(),
        template.Barrier(), // May be nil for demo services
        cfg.JWTSecret, // Product-specific setting
        tlsCfg,
    )

    // 4. Register handlers via injection pattern
    if cfg.PublicHandlers != nil {
        if err := cfg.PublicHandlers(publicServer); err != nil {
            return nil, fmt.Errorf("failed to register handlers: %w", err)
        }
    }

    return publicServer, nil
}
```

**Services Needing Factories**:

- [ ] cipher-im: `NewPublicServerFromConfig` in `internal/cipher/server/`
- [ ] jose-ja: `NewPublicServerFromConfig` in `internal/jose/server/`
- [ ] identity-authz: `NewPublicServerFromConfig` in `internal/identity/authz/server/`
- [ ] identity-idp: `NewPublicServerFromConfig` in `internal/identity/idp/server/`
- [ ] pki-ca: `NewPublicServerFromConfig` in `internal/ca/server/`

**Note**: sm-kms already has this pattern via `application_listener.go` (reference implementation).

### Phase 3: ApplicationListener Implementation

**File**: `internal/template/server/listener/application_listener.go`

**Tasks**:

1. Remove TODO markers from `StartApplicationListener`
2. Integrate product-specific server factory (passed via cfg)
3. Create admin server using existing `NewAdminHTTPServer`
4. Extract actual ports from started servers
5. Populate TLS configs for client verification

**Implementation Pattern**:

```go
func StartApplicationListener(ctx context.Context, cfg *ApplicationConfig) (*ApplicationListener, error) {
    // ... validation ...

    // Create ServiceTemplate (telemetry, JWK gen, barrier)
    template, err := cryptoutilTemplateServer.NewServiceTemplate(ctx, cfg.ServerSettings, cfg.DB, cfg.DBType)
    if err != nil {
        return nil, fmt.Errorf("failed to create service template: %w", err)
    }

    // Create public server via product-specific factory
    publicServer, err := cfg.PublicServerFactory(ctx, cfg, template)
    if err != nil {
        return nil, fmt.Errorf("failed to create public server: %w", err)
    }

    // Create admin server (reusable across all services)
    adminTLSCfg, err := tlsGenerator.GenerateAutoTLSGeneratedSettings(...)
    adminServer, err := NewAdminHTTPServer(ctx, cfg.ServerSettings, adminTLSCfg)
    if err != nil {
        return nil, fmt.Errorf("failed to create admin server: %w", err)
    }

    // Register optional admin handlers (barrier rotation, diagnostics)
    if cfg.AdminHandlers != nil {
        if err := cfg.AdminHandlers(adminServer); err != nil {
            return nil, fmt.Errorf("failed to register admin handlers: %w", err)
        }
    }

    // Create Application (manages both servers)
    app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
    if err != nil {
        return nil, fmt.Errorf("failed to create application: %w", err)
    }

    // Start servers in background
    errChan := make(chan error, 1)
    go func() {
        errChan <- app.Start(ctx)
    }()

    // Wait for startup or early error
    select {
    case err := <-errChan:
        return nil, fmt.Errorf("failed to start application: %w", err)
    case <-time.After(2 * time.Second): // Servers started successfully
    }

    // Extract actual ports (important for port 0 dynamic allocation)
    actualPublicPort := publicServer.ActualPort()
    actualPrivatePort, _ := adminServer.ActualPort()

    // Create shutdown function
    shutdownFunc := func() {
        template.Shutdown() // Telemetry, JWK gen, barrier
        _ = app.Shutdown(ctx) // Public + admin servers
    }

    return &ApplicationListener{
        app:               app,
        config:            cfg.ServerSettings,
        shutdownFunc:      shutdownFunc,
        actualPublicPort:  uint16(actualPublicPort),
        actualPrivatePort: uint16(actualPrivatePort),
        // TODO: Populate TLS configs for client verification
    }, nil
}
```

### Phase 4: TestMain Migration

**Example**: `internal/cipher/server/testmain_test.go`

**Before** (150+ lines of boilerplate):

```go
func TestMain(m *testing.M) {
    ctx := context.Background()

    // Manual database setup
    dbID, _ := googleUuid.NewV7()
    dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"
    testSQLDB, err = sql.Open("sqlite", dsn)
    // ... 20 lines of SQLite configuration ...

    // Manual GORM setup
    testDB, err = gorm.Open(sqlite.Dialector{Conn: testSQLDB}, &gorm.Config{...})

    // Manual migrations
    err = repository.ApplyMigrations(testSQLDB, repository.DatabaseTypeSQLite)

    // Manual telemetry setup
    testTelemetryService, err = cryptoutilTelemetry.NewTelemetryService(ctx, ...)
    defer testTelemetryService.Shutdown()

    // Manual JWK Gen setup
    testJWKGenService, err = cryptoutilJose.NewJWKGenService(ctx, testTelemetryService, false)
    defer testJWKGenService.Shutdown()

    // Manual TLS config generation
    testTLSCfg, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(...)

    // Manual server creation
    testCipherIMServer, baseURL, adminURL, err = createTestCipherIMServer(testDB)
    defer testCipherIMServer.Shutdown(context.Background())

    os.Exit(m.Run())
}
```

**After** (30 lines, clean and consistent):

```go
var (
    testListener *listener.ApplicationListener
    baseURL      string
    adminURL     string
)

func TestMain(m *testing.M) {
    ctx := context.Background()

    // Create in-memory SQLite database
    db, sqlDB, err := createInMemoryDB(ctx) // Helper function (reusable)
    if err != nil {
        panic("failed to create database: " + err.Error())
    }
    defer sqlDB.Close()

    // Apply migrations
    if err := repository.ApplyMigrations(sqlDB, repository.DatabaseTypeSQLite); err != nil {
        panic("failed to apply migrations: " + err.Error())
    }

    // Configure application
    cfg := &listener.ApplicationConfig{
        ServerSettings:      cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true),
        DB:                  db,
        DBType:              cryptoutilTemplateServerRepository.DatabaseTypeSQLite,
        PublicServerFactory: NewPublicServerFromConfig, // Product-specific factory
        PublicHandlers:      nil, // Optional: register additional routes
        AdminHandlers:       registerBarrierRoutes, // Optional: barrier rotation endpoints
    }

    // Start application (one line!)
    testListener, err = listener.StartApplicationListener(ctx, cfg)
    if err != nil {
        panic("failed to start application: " + err.Error())
    }
    defer testListener.Shutdown()

    // Extract URLs for tests
    baseURL = fmt.Sprintf("https://127.0.0.1:%d", testListener.ActualPublicPort())
    adminURL = fmt.Sprintf("https://127.0.0.1:%d", testListener.ActualPrivatePort())

    os.Exit(m.Run())
}
```

**Helper Function** (reusable across all services):

```go
// createInMemoryDB creates an in-memory SQLite database configured for concurrent operations.
// Returns GORM DB, sql.DB (for migrations), and error.
func createInMemoryDB(ctx context.Context) (*gorm.DB, *sql.DB, error) {
    dbID, _ := googleUuid.NewV7()
    dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

    sqlDB, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to open SQLite: %w", err)
    }

    // Configure SQLite for concurrent operations
    if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
        sqlDB.Close()
        return nil, nil, fmt.Errorf("failed to enable WAL: %w", err)
    }

    if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
        sqlDB.Close()
        return nil, nil, fmt.Errorf("failed to set busy timeout: %w", err)
    }

    sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
    sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
    sqlDB.SetConnMaxLifetime(0)

    // Wrap with GORM
    db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
        SkipDefaultTransaction: true,
    })
    if err != nil {
        sqlDB.Close()
        return nil, nil, fmt.Errorf("failed to create GORM DB: %w", err)
    }

    return db, sqlDB, nil
}
```

## Benefits

### Consistency Across Services

**All services use same pattern**:

- cipher-im TestMain
- jose-ja TestMain
- identity-* TestMain
- sm-kms TestMain (already has this pattern)
- pki-ca TestMain

**Result**: 150+ lines → 30 lines per TestMain

### Encapsulation

**Hides complexity**:

- Telemetry initialization
- JWK generation service
- Barrier service (optional)
- TLS certificate generation
- Server lifecycle management

**Exposes simple interface**:

- StartApplicationListener(ctx, cfg)
- SendLivenessCheck(settings)
- SendReadinessCheck(settings)
- SendShutdownRequest(settings)
- listener.Shutdown()

### Testing Benefits

**Simplified TestMain**:

- Consistent database setup pattern
- Single application startup call
- Automatic port extraction (no hardcoded ports)
- Clean shutdown via defer

**Better isolation**:

- Each test package gets isolated server
- Database cleaned between tests (cleanTestDB helper)
- No shared state between packages

### Production Benefits

**Same code path**:

- TestMain uses ApplicationListener
- Production `cmd/cipher-im/main.go` uses ApplicationListener
- Same initialization, same health checks, same shutdown

**Operational consistency**:

- Health check endpoints work identically
- Graceful shutdown behavior matches across environments
- TLS configuration consistent (test vs production)

## Migration Checklist

### Per-Service Tasks

**For each service** (cipher-im, jose-ja, identity-*, pki-ca):

- [ ] Create `NewPublicServerFromConfig` factory function
  - [ ] Accept ApplicationConfig parameter
  - [ ] Create repositories from template.DB()
  - [ ] Generate TLS config
  - [ ] Call existing NewPublicServer constructor
  - [ ] Register handlers via cfg.PublicHandlers

- [ ] Update TestMain to use ApplicationListener
  - [ ] Replace manual setup with StartApplicationListener
  - [ ] Use createInMemoryDB helper (reusable)
  - [ ] Extract URLs from listener.ActualPublicPort/ActualPrivatePort
  - [ ] Simplify shutdown to single defer listener.Shutdown()

- [ ] Update production main.go
  - [ ] Replace manual setup with StartApplicationListener
  - [ ] Use same ApplicationConfig pattern
  - [ ] Graceful shutdown via SendShutdownRequest

### Cross-Service Tasks

- [ ] Extract createInMemoryDB helper to `internal/template/testing/database`
  - [ ] Reusable across all services
  - [ ] Handles SQLite WAL configuration
  - [ ] Returns both GORM and sql.DB

- [ ] Finish StartApplicationListener implementation
  - [ ] Integrate PublicServerFactory
  - [ ] Create admin server
  - [ ] Start Application
  - [ ] Extract actual ports

- [ ] Add health check tests
  - [ ] Verify liveness endpoint responds <100ms
  - [ ] Verify readiness endpoint checks dependencies
  - [ ] Test graceful shutdown sequence

## Testing Strategy

### Unit Tests (per-service)

**Test NewPublicServerFromConfig**:

- [ ] Validates nil parameters
- [ ] Creates server with correct bind port
- [ ] Registers handlers successfully
- [ ] Returns error on invalid configuration

### Integration Tests

**Test ApplicationListener lifecycle**:

- [ ] StartApplicationListener succeeds with valid config
- [ ] Liveness check responds after startup
- [ ] Readiness check validates dependencies
- [ ] Graceful shutdown completes within 30s
- [ ] Double shutdown is idempotent (no panic)

### E2E Tests

**Test full service stack**:

- [ ] Start listener with PostgreSQL test-container
- [ ] Send requests to public APIs
- [ ] Verify health checks pass
- [ ] Trigger shutdown via admin API
- [ ] Verify servers stop accepting requests

## Cross-References

- **sm-kms Reference**: `internal/kms/server/application/application_listener.go` (complete working example)
- **Service Template**: `internal/template/server/` (shared infrastructure)
- **Testing Patterns**: `03-02.testing.instructions.md` (TestMain patterns)
- **Architecture**: `02-02.service-template.instructions.md` (dual server pattern)
- **Database**: `03-05.sqlite-gorm.instructions.md` (SQLite configuration)

## Future Enhancements

### Phase 5: Configuration Validation

**Add pre-flight checks**:

- Validate bind addresses (127.0.0.1 for tests, 0.0.0.0 for containers)
- Verify TLS certificates exist and are valid
- Check database connectivity before server startup
- Validate OTLP endpoint reachability

### Phase 6: Metrics and Monitoring

**Add telemetry to ApplicationListener**:

- Track startup duration
- Monitor health check latency
- Alert on readiness failures
- Expose Prometheus metrics for server lifecycle

### Phase 7: Hot-Reload Configuration

**Support runtime configuration changes**:

- Watch config file for changes
- Reload TLS certificates without restart
- Update rate limits dynamically
- Reconfigure database pool settings

## Summary

The ApplicationListener pattern provides:

1. **Consistent**: Same startup code across all services
2. **Simple**: 150+ lines → 30 lines per TestMain
3. **Encapsulated**: Hides infrastructure complexity
4. **Testable**: Clean interfaces for mocking
5. **Production-Ready**: Same code path for test and production

**Next Steps**:

1. Implement Phase 2 factories (cipher-im first)
2. Complete Phase 3 ApplicationListener
3. Migrate Phase 4 TestMain (cipher-im validation)
4. Roll out to remaining services (jose-ja, identity-*, pki-ca)
