# Service Template Application Layer - Implementation Summary

**Date**: 2026-01-05
**Status**: Implementation Complete - Ready for Testing

## What Was Implemented

### New Files Created

1. **`internal/template/server/application/application_basic.go`**
   - Foundation layer: telemetry, unseal keys, JWK generation
   - Extracted from sm-kms `application_basic.go`
   - Provides `StartApplicationBasic()` and `Shutdown()`

2. **`internal/template/server/application/application_core.go`**
   - Database provisioning layer
   - Handles three scenarios automatically:
     - SQLite in-memory (`file::memory:?cache=shared`)
     - PostgreSQL testcontainer (with fallback support)
     - External database connection
   - Provides `StartApplicationCore()` and `Shutdown()`
   - Includes `provisionDatabase()` helper

3. **`internal/template/server/application/application_listener.go`**
   - Top-level orchestrator
   - Manages public + admin servers
   - Provides `StartApplicationListener()`, `Start()`, `Shutdown()`
   - Interfaces: `IPublicServer`, `IAdminServer`

4. **`internal/template/server/application/README.md`**
   - Comprehensive documentation
   - Architecture explanation (Basic → Core → Listener layers)
   - Migration benefits (before/after comparison)
   - Testing strategy (unit, integration, E2E)
   - Phase migration plan

5. **`docs/cipher-im-migration/SERVICE-TEMPLATE-USAGE-GUIDE.md`**
   - Practical usage examples
   - TestMain patterns for integration/E2E tests
   - Database configuration matrix
   - Migration checklist

### Modified Files

1. **`internal/template/server/service_template.go`**
   - Added import for `application` package
   - Added `StartApplicationCore()` convenience wrapper

2. **`internal/cipher/server/server.go`**
   - Added `NewFromConfig(ctx, cfg)` constructor
   - Marked `New(ctx, cfg, db, dbType)` as DEPRECATED
   - Uses `StartApplicationCore()` for automatic DB provisioning

## Key Features

### Automatic Database Provisioning

**Configuration-Driven**:

- `DatabaseURL=""` + `DatabaseContainer="required"` → PostgreSQL testcontainer
- `DatabaseURL="file::memory:?cache=shared"` → SQLite in-memory
- `DatabaseURL="postgres://..."` + `DatabaseContainer="disabled"` → External DB

**Graceful Fallback**:

- `DatabaseContainer="preferred"` → Try testcontainer, fall back to external

### Simplified Test Setup

**Before** (manual management):

- 50+ lines per TestMain
- Manual container lifecycle
- Manual service initialization
- Complex defer cleanup chains

**After** (automatic):

- 15 lines per TestMain
- Single `NewFromConfig()` call
- Automatic cleanup via `Shutdown()`
- Zero manual container management

**Example**:

```go
func TestMain(m *testing.M) {
    cfg := &config.AppConfig{
        ServerSettings: cryptoutilConfig.ServerSettings{
            DatabaseURL:       "",
            DatabaseContainer: "required",
        },
        JWTSecret: uuid.Must(uuid.NewUUID()).String(),
    }

    server, _ := server.NewFromConfig(context.Background(), cfg)
    exitCode := m.Run()
    server.Shutdown(context.Background())
    os.Exit(exitCode)
}
```

## Architecture Pattern (Matches sm-kms)

```
ApplicationListener (Top Level)
├── ApplicationCore (Database + Infrastructure)
│   ├── ApplicationBasic (Foundation Services)
│   │   ├── TelemetryService
│   │   ├── UnsealKeysService
│   │   └── JWKGenService
│   ├── DB (GORM instance - auto-provisioned)
│   └── ShutdownDBContainer (cleanup function)
├── PublicServer (Business APIs)
└── AdminServer (Health Checks)
```

## Testing Strategy

### Unit Tests

- Test each layer independently
- Mock dependencies (telemetry, unseal, JWK)
- Validate database provisioning logic

### Integration Tests

Use `NewFromConfig` with PostgreSQL testcontainer:

```go
cfg.DatabaseURL = ""
cfg.DatabaseContainer = "required"
```

### E2E Tests

Use `NewFromConfig` with SQLite in-memory (faster):

```go
cfg.DatabaseURL = "file::memory:?cache=shared"
cfg.DatabaseContainer = "disabled"
```

## Next Steps

### Immediate (cipher-im validation)

1. **Update `internal/cipher/integration/testmain_integration_test.go`**:
   - Replace manual container setup with `NewFromConfig`
   - Validate PostgreSQL testcontainer provisioning
   - Run tests: `go test ./internal/cipher/integration -v`

2. **Update `internal/cipher/e2e/testmain_e2e_test.go`**:
   - Replace manual DB setup with `NewFromConfig`
   - Validate SQLite in-memory provisioning
   - Run tests: `go test ./internal/cipher/e2e -v`

3. **Verify Cleanup**:
   - Check no hanging containers: `docker ps`
   - Check no leaked connections: Database connection pool stats

### Future (other services)

After cipher-im validates the pattern:

**Phase 2**: One service at a time

- jose-ja server
- pki-ca server
- identity services (authz, idp, rs, rp, spa)

**Phase 3**: sm-kms last

- Only after ALL other services validated
- Validates template handles all edge cases

## Benefits Summary

| Aspect | Before | After | Improvement |
|--------|--------|-------|-------------|
| **TestMain Lines** | 50+ | 15 | 70% reduction |
| **Container Management** | Manual | Automatic | Zero manual code |
| **Service Init** | Manual | Automatic | Zero manual code |
| **Cleanup** | defer chains | Single Shutdown() | Simplified |
| **Database Support** | Hardcoded | Configurable | Flexible |
| **Fallback** | None | Built-in | Robust |

## Cross-References

**Documentation**:

- `internal/template/server/application/README.md` - Architecture details
- `docs/cipher-im-migration/SERVICE-TEMPLATE-USAGE-GUIDE.md` - Usage examples
- `02-02.service-template.instructions.md` - Service template requirements

**Reference Implementation**:

- `internal/kms/server/application/application_basic.go`
- `internal/kms/server/application/application_core.go`
- `internal/kms/server/application/application_listener.go`
- `internal/kms/server/repository/sqlrepository/sql_provider.go`

## Questions/Clarifications

**Q: Why not use `ApplicationListener` in tests?**
A: Tests may need `ApplicationCore` directly to access `Core.DB` for assertions. `NewFromConfig` creates both `ApplicationCore` and servers together.

**Q: Can I still use manual database setup?**
A: Yes, the `New(ctx, cfg, db, dbType)` constructor remains for backward compatibility (marked DEPRECATED).

**Q: What about main.go?**
A: `NewFromConfig()` works in main.go too - same pattern as tests, just different config sources (flags/env/file vs hardcoded).

**Q: How do I test database provisioning logic?**
A: Unit test `provisionDatabase()` function directly with mock telemetry and different config combinations.
