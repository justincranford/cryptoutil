# Service Template Migration - Resource Cleanup Tracking

**Created**: 2026-01-02
**Scope**: Missing Close() and Shutdown() calls for services and databases
**Target Directories**: `internal/cipher/`, `internal/template/server/`

## Overview

This document tracks test files that need resource cleanup fixes. Services from `internal/template/server/` must be properly shut down, and databases like SQLite need Close() calls.

**Common Missing Cleanup Calls**:
- `sqlDB.Close()`
- `telemetryService.Shutdown()`
- `jwkGenService.Shutdown()`
- `unsealKeysService.Shutdown()`
- `barrierRepo.Shutdown()`
- `barrierService.Shutdown()`
- PostgreSQL test-container cleanup

---

## Phase 1: Critical TestMain Files (High Priority)

TestMain functions initialize shared resources that MUST be cleaned up to prevent resource leaks across test runs.

### internal/cipher/

- [ ] **`internal/cipher/e2e/testmain_e2e_test.go`**
  - Missing: `sharedJWKGenService.Shutdown()`
  - Missing: `unsealService.Shutdown()` (local variable line 99)
  - Missing: `barrierRepo.Shutdown()` (local variable line 106)
  - Missing: `testBarrierService.Shutdown()`
  - Missing: `sqlDB.Close()` (db from initTestDB line 89)
  - Has: `sharedTelemetryService.Shutdown()` ✅
  - Has: `testPublicServer.Shutdown()` ✅

- [ ] **`internal/cipher/server/testmain_test.go`**
  - Missing: `testJWKGenService.Shutdown()`
  - Missing: `unsealKeysService.Shutdown()` (local variable line 115)
  - Missing: `barrierRepo.Shutdown()` (local variable line 120)
  - Missing: `testBarrierService.Shutdown()`
  - Missing: `testSQLDB.Close()` (deferred comment warns against closing too early, but should defer at function start)
  - Has: `testTelemetryService.Shutdown()` ✅ (but too late - after m.Run())
  - Note: Comment at line 160+ warns about parallel tests, but cleanup should still happen via defer

- [ ] **`internal/cipher/integration/testmain_integration_test.go`**
  - Missing: `sqlDB.Close()` (from gorm db.DB() at line 78)
  - Has: `sharedPGContainer.Terminate()` ✅
  - Note: Opens GORM connection but doesn't close underlying sql.DB

### internal/template/server/

- [ ] **`internal/template/server/barrier/barrier_service_test.go`**
  - Missing: `testJWKGenService.Shutdown()` (line 89)
  - Missing: `testTelemetryService.Shutdown()` (line 83)
  - Missing: `unsealService.Shutdown()` (local variable line 100)
  - Missing: `barrierRepo.Shutdown()` (local variable line 106)
  - Has: `testBarrierService.Shutdown()` ✅ (line 128)
  - Has: `testSQLDB.Close()` ✅ (line 130)

---

## Phase 2: Per-Test Service Initialization Files (Medium Priority)

These files create services within individual test functions that need cleanup via t.Cleanup() or defer.

### internal/cipher/

- [ ] **`internal/cipher/server/realms/middleware_test.go`**
  - Function: `createTestPublicServer()` (line 66)
  - Missing: `telemetryService.Shutdown()` (line 76)
  - Missing: `jwkGenService.Shutdown()` (line 79)
  - Missing: `unsealService.Shutdown()` (line 86)
  - Missing: `barrierRepo.Shutdown()` (line 89)
  - Missing: `barrierService.Shutdown()` (line 92)
  - Missing: `sqlDB.Close()` (from initTestDB)
  - Has: `srv.Shutdown()` via t.Cleanup() ✅ (line 150+)
  - Note: Services created per-test should be cleaned up in t.Cleanup()

- [ ] **`internal/cipher/server/server_lifecycle_test.go`**
  - Function: Multiple test functions create services
  - Missing: `telemetryService.Shutdown()` (line 126)
  - Missing: `jwkGenService.Shutdown()` (line 129)
  - Note: Uses testBarrierService from TestMain (shared), but creates new telemetry/jwk services per-test

### internal/template/server/

- [ ] **`internal/template/server/barrier/rotation_handlers_test.go`**
  - Function: `setupRotationTestEnvironment()` (line 30)
  - Missing: `testSQLDB.Close()` (line 38)
  - Missing: `telemetryService.Shutdown()` (line 69)
  - Missing: `jwkGenService.Shutdown()` (line 73)
  - Missing: `unsealService.Shutdown()` (line 80)
  - Missing: `barrierRepo.Shutdown()` (line 84)
  - Missing: `barrierService.Shutdown()` (line 88)
  - Missing: `rotationService` cleanup (if it has Shutdown method)
  - Note: Helper function creates many services but doesn't clean them up

- [ ] **`internal/template/server/barrier/gorm_barrier_repository_test.go`**
  - Multiple test functions create barrierRepo
  - Missing: `barrierRepo.Shutdown()` calls for instances at:
    - Line 67 (TestGormBarrierRepository_StoreRootKey_Success)
    - Line 168 (TestGormBarrierRepository_LoadRootKey_Success)
    - Line 269 (TestGormBarrierRepository_StoreIntermediateKey_Success)
    - Line 366 (TestGormBarrierRepository_LoadIntermediateKey_Success)
    - Line 410 (TestGormBarrierRepository_StoreContentKey_Success)
  - Has: `sqlDB.Close()` via defer ✅ (line 51)

---

## Phase 3: Shared Testutil Pattern Validation (Low Priority)

These TestMain files delegate to testutil.Initialize(). Verify testutil cleanup.

### internal/template/server/

- [ ] **`internal/template/server/test_main_test.go`**
  - Delegates to: `cryptoutilTemplateServerTestutil.Initialize()`
  - Action: Verify testutil has proper cleanup (TLS configs don't need cleanup)

- [ ] **`internal/template/server/repository/test_main_test.go`**
  - Delegates to: `cryptoutilTemplateServerTestutil.Initialize()`
  - Action: Same as above

- [ ] **`internal/template/server/listener/test_main_test.go`**
  - Delegates to: `cryptoutilTemplateServerTestutil.Initialize()`
  - Action: Same as above

---

## Phase 4: Review Remaining Test Files (Verification)

Scan remaining test files for any missed service initialization.

### internal/cipher/

- [ ] `internal/cipher/server/middleware_test.go` - Review for service cleanup
- [ ] `internal/cipher/server/helpers_test.go` - Review for service cleanup
- [ ] `internal/cipher/server/realm_validation_test.go` - Review for service cleanup
- [ ] `internal/cipher/server/realms/realm_validation_test.go` - Review for service cleanup
- [ ] `internal/cipher/e2e/browser_e2e_test.go` - Review for service cleanup
- [ ] `internal/cipher/e2e/service_e2e_test.go` - Review for service cleanup
- [ ] `internal/cipher/e2e/rotation_e2e_test.go` - Review for service cleanup
- [ ] `internal/cipher/e2e/helpers_e2e_test.go` - Review for service cleanup
- [ ] `internal/cipher/integration/concurrent_test.go` - Review for service cleanup
- [ ] `internal/cipher/repository/message_recipient_jwk_repository_test.go` - Review for service cleanup
- [ ] `internal/cipher/crypto/password_test.go` - Review for service cleanup (likely no services)

### internal/template/server/

- [ ] `internal/template/server/service_template_test.go` - Review for service cleanup
- [ ] `internal/template/server/application_test.go` - Review for service cleanup
- [ ] `internal/template/server/listener/admin_test.go` - Review for service cleanup
- [ ] `internal/template/server/listener/public_test.go` - Review for service cleanup
- [ ] `internal/template/server/listener/servers_test.go` - Review for service cleanup
- [ ] `internal/template/server/repository/public_table_test.go` - Review for service cleanup
- [ ] `internal/template/server/repository/application_table_test.go` - Review for service cleanup
- [ ] `internal/template/server/barrier/status_handlers_test.go` - Review for service cleanup

---

## Cleanup Patterns Reference

### Pattern 1: TestMain Cleanup (Deferred at Top)

```go
func TestMain(m *testing.M) {
    ctx := context.Background()

    // Setup resources
    telemetrySvc, _ := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
    defer telemetrySvc.Shutdown()

    jwkGenSvc, _ := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
    defer jwkGenSvc.Shutdown()

    unsealSvc, _ := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
    defer unsealSvc.Shutdown()

    barrierRepo, _ := cryptoutilBarrier.NewGormBarrierRepository(db)
    defer barrierRepo.Shutdown()

    barrierSvc, _ := cryptoutilBarrier.NewBarrierService(ctx, telemetrySvc, jwkGenSvc, barrierRepo, unsealSvc)
    defer barrierSvc.Shutdown()

    sqlDB, _ := sql.Open("sqlite", dsn)
    defer sqlDB.Close()

    // Run tests
    exitCode := m.Run()

    os.Exit(exitCode)
}
```

### Pattern 2: Per-Test Cleanup (t.Cleanup)

```go
func setupTestServices(t *testing.T) (*BarrierService, *sql.DB) {
    t.Helper()

    telemetrySvc, _ := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
    t.Cleanup(func() { telemetrySvc.Shutdown() })

    jwkGenSvc, _ := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
    t.Cleanup(func() { jwkGenSvc.Shutdown() })

    // ... more services

    sqlDB, _ := sql.Open("sqlite", dsn)
    t.Cleanup(func() { _ = sqlDB.Close() })

    return barrierSvc, sqlDB
}
```

### Pattern 3: Inline Defer (Simple Cases)

```go
func TestSomething(t *testing.T) {
    telemetrySvc, _ := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
    defer telemetrySvc.Shutdown()

    jwkGenSvc, _ := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
    defer jwkGenSvc.Shutdown()

    // ... test logic
}
```

---

## Phase 5: NewTestConfig Migration (CRITICAL - Windows Firewall Root Cause)

**Root Cause Identified**: Blank `BindPublicAddress=""` + `BindPublicPort=0` → `fmt.Sprintf("%s:%d", "", 0)` → `":0"` → `net.Listen()` binds to **0.0.0.0** (all interfaces) → **Windows Firewall exception prompt**

**Solution**: All test files MUST use `NewTestConfig(bindAddr, bindPort, devMode)` from `internal/shared/config/config_test_helper.go` instead of creating `&cryptoutilConfig.ServerSettings{}` with partial/blank fields.

**Critical Bind Address Pattern** (validated in code):
```go
// internal/template/server/listener/public.go line 168:
listener, err := listenConfig.Listen(ctx, "tcp", fmt.Sprintf("%s:%d", s.settings.BindPublicAddress, s.settings.BindPublicPort))

// When BindPublicAddress="" and BindPublicPort=0:
// fmt.Sprintf("%s:%d", "", 0) == ":0"
// net.Listen("tcp", ":0") == binds to 0.0.0.0:0 (ALL INTERFACES)
// Windows Firewall exception triggered!

// Correct pattern with NewTestConfig:
// settings := cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)
// fmt.Sprintf("%s:%d", "127.0.0.1", 0) == "127.0.0.1:0"
// net.Listen("tcp", "127.0.0.1:0") == localhost only, no firewall prompt
```

### Files Requiring NewTestConfig Migration (20+ files):

#### internal/template/server/

- [ ] **`internal/template/server/listener/servers_test.go`** (line 15)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

- [ ] **`internal/template/server/service_template_test.go`** (line ~30-50)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

- [ ] **`internal/template/server/barrier/barrier_service_test.go`** (2 instances: lines ~65, ~145)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

- [ ] **`internal/template/server/barrier/rotation_handlers_test.go`** (line ~40)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

#### internal/cipher/

- [ ] **`internal/cipher/server/testmain_test.go`** (line 90)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

- [ ] **`internal/cipher/server/server_lifecycle_test.go`** (multiple instances)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

- [ ] **`internal/cipher/server/realms/middleware_test.go`** (line ~70)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

#### internal/kms/server/

- [ ] **`internal/kms/server/application/application_init_test.go`** (3 instances: lines 43, 118, 136)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

- [ ] **`internal/kms/server/businesslogic/businesslogic_test.go`** (2 instances)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

#### internal/shared/config/ (Test files - LOW PRIORITY)

- [ ] **`internal/shared/config/url_test.go`** (6 instances)
  - Note: These test URL generation, NOT server startup, so lower priority
  - Current: `&ServerSettings{...}` with minimal fields for URL testing
  - Fix: Consider using NewTestConfig for consistency, but won't trigger firewall

- [ ] **`internal/shared/config/config_test.go`** (5 instances)
  - Note: These test configuration validation, NOT server startup
  - Current: `&ServerSettings{...}` with minimal fields for validation testing
  - Fix: Consider using NewTestConfig for consistency, but won't trigger firewall

#### Good Example (Already Correct):

- ✅ **`internal/jose/server/server_test.go`**
  - Uses: `cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)`
  - Result: `BindPublicAddress="127.0.0.1"`, `BindPublicPort=0` → `"127.0.0.1:0"` → localhost only, no firewall

---

## Phase 6: validateConfiguration() Enhancement (Configuration Robustness)

**Current State**: `validateConfiguration()` at `internal/shared/config/config.go:1379` validates some fields but is incomplete.

**Goal**: Add comprehensive validation for ALL ServerSettings fields to catch configuration errors early.

### Missing Validations to Add:

#### Network Configuration:

- [ ] **BindPublicAddress** - MUST NOT be blank (reject "", enforce "127.0.0.1" or "0.0.0.0")
- [ ] **BindPublicPort** - MUST be 0-65535 (reject negative)
- [ ] **BindPrivateAddress** - MUST NOT be blank (reject "", enforce "127.0.0.1")
- [ ] **BindPrivatePort** - MUST be 0-65535 (reject negative)
- [ ] **BindPublicProtocol** - MUST be "https" (reject "http" in production)
- [ ] **BindPrivateProtocol** - MUST be "https" (reject "http" in production)

#### TLS Configuration:

- [ ] **TLSPublicDNSNames** - MUST contain at least one entry (reject empty array)
- [ ] **TLSPublicIPAddresses** - MUST contain at least one entry (reject empty array)
- [ ] **TLSPrivateDNSNames** - MUST contain at least one entry
- [ ] **TLSPrivateIPAddresses** - MUST contain at least one entry

#### CORS Configuration:

- [ ] **CORSAllowedOrigins** - MUST contain at least one entry for browser endpoints
- [ ] Each origin MUST start with "http://" or "https://"
- [ ] Each origin MUST NOT have trailing slash

#### CSRF Configuration:

- [ ] **CSRFEnabled** - If true, validate CSRFCookieName is non-blank
- [ ] **CSRFCookieName** - MUST NOT be blank if CSRF enabled

#### Rate Limiting:

- [ ] **RateLimitEnabled** - If true, validate rate limit values are positive
- [ ] **RateLimitRequestsPerMinute** - MUST be positive if rate limiting enabled
- [ ] **RateLimitBurst** - MUST be positive if rate limiting enabled

#### Database Configuration:

- [ ] **DatabaseURL** - MUST NOT be blank
- [ ] **DatabaseURL** - MUST start with "postgres://" or "file:" for SQLite
- [ ] **DatabaseMaxOpenConnections** - MUST be positive
- [ ] **DatabaseMaxIdleConnections** - MUST be ≤ MaxOpenConnections

#### OTLP Configuration:

- [ ] **OTLPEnabled** - If true, validate OTLP endpoint is non-blank
- [ ] **OTLPEndpoint** - MUST NOT be blank if OTLP enabled
- [ ] **OTLPServiceName** - MUST NOT be blank if OTLP enabled
- [ ] **OTLPServiceVersion** - MUST NOT be blank if OTLP enabled

#### Security Configuration:

- [ ] **SessionSecretKey** - MUST be ≥32 bytes for production (reject weak keys)
- [ ] **CSPEnabled** - If true, validate CSP directives are non-blank
- [ ] **DevMode** - WARN if true in production environment

### Validation Error Message Pattern:

```go
// Example comprehensive validation:
func (s *ServerSettings) validateConfiguration() error {
    var errs []string

    // Network validation
    if s.BindPublicAddress == "" {
        errs = append(errs, "BindPublicAddress MUST NOT be blank (use '127.0.0.1' or '0.0.0.0')")
    }
    if s.BindPublicPort < 0 || s.BindPublicPort > 65535 {
        errs = append(errs, fmt.Sprintf("BindPublicPort must be 0-65535, got: %d", s.BindPublicPort))
    }
    if s.BindPrivateAddress == "" {
        errs = append(errs, "BindPrivateAddress MUST NOT be blank (use '127.0.0.1')")
    }

    // TLS validation
    if len(s.TLSPublicDNSNames) == 0 {
        errs = append(errs, "TLSPublicDNSNames MUST contain at least one DNS name")
    }
    if len(s.TLSPublicIPAddresses) == 0 {
        errs = append(errs, "TLSPublicIPAddresses MUST contain at least one IP address")
    }

    // Database validation
    if s.DatabaseURL == "" {
        errs = append(errs, "DatabaseURL MUST NOT be blank")
    }
    if !strings.HasPrefix(s.DatabaseURL, "postgres://") && !strings.HasPrefix(s.DatabaseURL, "file:") {
        errs = append(errs, "DatabaseURL must start with 'postgres://' or 'file:' for SQLite")
    }

    // OTLP validation
    if s.OTLPEnabled {
        if s.OTLPEndpoint == "" {
            errs = append(errs, "OTLPEndpoint MUST NOT be blank when OTLP enabled")
        }
        if s.OTLPServiceName == "" {
            errs = append(errs, "OTLPServiceName MUST NOT be blank when OTLP enabled")
        }
    }

    // Return aggregated errors
    if len(errs) > 0 {
        return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errs, "\n  - "))
    }
    return nil
}
```

---

## Priority Execution Order

1. **Phase 5 (CRITICAL - Windows Firewall)**: Migrate all test files to NewTestConfig - prevents firewall prompts
2. **Phase 1 (Critical - Resource Leaks)**: Fix all TestMain files - affects ALL tests in the package
3. **Phase 2 (Medium - Resource Leaks)**: Fix per-test service initialization - prevents leaks in individual tests
4. **Phase 6 (Medium - Configuration Robustness)**: Enhance validateConfiguration() - catches config errors early
5. **Phase 3 (Low - Verification)**: Verify testutil pattern - likely already correct
6. **Phase 4 (Low - Verification)**: Scan remaining files to catch any missed cleanup

---

## Success Criteria

- [ ] All TestMain files have proper cleanup via defer for all services
- [ ] All per-test service creation uses t.Cleanup() or defer
- [ ] No resource leak warnings in test output
- [ ] Tests can run repeatedly without "address already in use" errors
- [ ] SQLite in-memory databases properly closed
- [ ] PostgreSQL test-containers properly terminated
- [ ] All telemetry services shut down cleanly
- [ ] All JWK generation services shut down cleanly
- [ ] All barrier repositories shut down cleanly
- [ ] All barrier services shut down cleanly

---

## Notes

- **Defer Order**: Resources should be cleaned up in reverse order of creation (LIFO)
- **TestMain Pattern**: Use defer immediately after successful creation
- **Per-Test Pattern**: Use t.Cleanup() for test-scoped resources
- **Error Handling**: Cleanup functions should handle errors gracefully (log but don't fail)
- **Parallel Tests**: t.Cleanup() is safe for parallel tests; defer in TestMain runs after all tests complete
