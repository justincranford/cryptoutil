# Service Template Application Layer

This package provides a layered application architecture for cryptoutil services, extracted from the sm-kms reference implementation.

## Architecture Layers

### 1. ApplicationBasic (`application_basic.go`)
**Foundation layer** providing basic service infrastructure:
- **Telemetry Service**: Logging, metrics, tracing (OpenTelemetry)
- **Unseal Keys Service**: Key encryption at rest (HKDF-based key derivation)
- **JWK Generation Service**: Cryptographic key generation (RSA, ECDSA, EdDSA)

**Usage**: Internal dependency for ApplicationCore, not used directly by services.

### 2. ApplicationCore (`application_core.go`)
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

### 3. ApplicationListener (`application_listener.go`)
**Top-level orchestrator** managing HTTP servers and core infrastructure:
- **Public Server**: Business logic (APIs, UI, external clients)
- **Admin Server**: Health checks, graceful shutdown (Kubernetes probes)
- **Lifecycle Management**: Start/shutdown both servers concurrently

**Usage**: Public API for services (replace manual server management).

## Migration Benefits

### Before (Manual Management)
Services manually managed:
- Database containers in TestMain
- Multiple service dependencies injected into server constructors
- Complex setup/cleanup in test files

Example from `internal/cipher/integration/testmain_integration_test.go`:
```go
func TestMain(m *testing.M) {
    // Manual PostgreSQL container setup.
    sharedPGContainer, sharedConnStr, err = container.SetupSharedPostgresContainer(ctx)
    // ...
    sharedDB, err = gorm.Open(postgresDriver.Open(sharedConnStr), &gorm.Config{})
    // ...
    sharedServer, err = server.New(ctx, cfg, db, repository.DatabaseTypePostgreSQL)
    // Manual cleanup with defer.
}
```

### After (Service Template)
Service-template handles all infrastructure:
```go
func TestMain(m *testing.M) {
    cfg := &config.AppConfig{
        ServerSettings: cryptoutilConfig.ServerSettings{
            DatabaseURL:       "", // Empty = use testcontainer.
            DatabaseContainer: "required",
        },
        JWTSecret: uuid.Must(uuid.NewUUID()).String(),
    }
    
    // Single function call - handles everything.
    app, err := application.StartApplicationListener(ctx, &application.ApplicationListenerConfig{
        Settings:     &cfg.ServerSettings,
        PublicServer: publicServer,
        AdminServer:  adminServer,
    })
    
    exitCode := m.Run()
    app.Shutdown(context.Background()) // Automatic cleanup.
    os.Exit(exitCode)
}
```

**Reduction**:
- 50+ lines → 15 lines per TestMain
- Zero manual container management
- Zero manual service initialization
- Automatic cleanup (no defer chains)

## Implementation Checklist

### For New Services (e.g., cipher-im)
1. ✅ Create `internal/<service>/server/config/config.go` with `AppConfig` struct embedding `ServerSettings`
2. ✅ Create public server implementing `IPublicServer` interface
3. ✅ Create admin server implementing `IAdminServer` interface
4. ✅ Use `StartApplicationListener` in main.go and TestMain

### For Existing Services (e.g., sm-kms)
1. Migrate `application_basic.go` to use template version
2. Migrate `application_core.go` to use template version
3. Migrate `application_listener.go` to use template version
4. Update TestMain files to use new API

## Testing Strategy

### Unit Tests
Test individual layers independently:
- `ApplicationBasic`: Mock telemetry, unseal, JWK services
- `ApplicationCore`: Test database provisioning logic (SQLite, PostgreSQL, fallback)
- `ApplicationListener`: Test server lifecycle (start, shutdown, error handling)

### Integration Tests
Use `ApplicationCore` to provision test databases:
```go
func TestMain(m *testing.M) {
    core, _ := application.StartApplicationCore(ctx, settings)
    defer core.Shutdown()
    
    // Use core.DB for tests.
    sharedDB = core.DB
    
    os.Exit(m.Run())
}
```

### E2E Tests
Use `ApplicationListener` for full service stack:
```go
func TestMain(m *testing.M) {
    app, _ := application.StartApplicationListener(ctx, config)
    defer app.Shutdown(context.Background())
    
    // HTTP requests to app.PublicPort() and app.AdminPort().
    os.Exit(m.Run())
}
```

## Cross-References

**Related Files**:
- `internal/kms/server/application/application_basic.go` - Reference implementation
- `internal/kms/server/application/application_core.go` - Reference implementation
- `internal/kms/server/application/application_listener.go` - Reference implementation
- `internal/kms/server/repository/sqlrepository/sql_provider.go` - Database provisioning logic
- `internal/template/server/service_template.go` - Service template infrastructure

**Documentation**:
- `02-02.service-template.instructions.md` - Service template requirements
- `03-04.database.instructions.md` - Database patterns
- `03-05.sqlite-gorm.instructions.md` - SQLite configuration

## Phase Migration Plan

**Phase 1**: cipher-im service (validation) - CURRENT
**Phase 2**: One service at a time (jose-ja → pki-ca → identity services)
**Phase 3**: sm-kms last (after all other services validated)

See `02-02.service-template.instructions.md` for complete migration strategy.
