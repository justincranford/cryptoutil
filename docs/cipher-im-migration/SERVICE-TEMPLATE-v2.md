# Service Template Migration - Resource Cleanup Tracking

**Created**: 2026-01-02
**Scope**: Configuration safety, resource cleanup, and validation enhancement
**Target Directories**: `internal/cipher/`, `internal/template/server/`, `internal/kms/`, `internal/shared/config/`

## Overview

This document tracks test files requiring fixes for:

1. **Windows Firewall exceptions** (blank bind addresses)
2. **Configuration validation** (missing field validations)
3. **Resource cleanup** (missing Close/Shutdown calls)

---

## Phase 1: NewTestConfig Migration (CRITICAL - Windows Firewall Root Cause)

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

### Files Requiring NewTestConfig Migration (20+ files)

#### internal/template/server/

- [x] **`internal/template/server/listener/servers_test.go`** (line 15)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

- [x] **`internal/template/server/service_template_test.go`** (line ~30-50)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

- [x] **`internal/template/server/barrier/barrier_service_test.go`** (2 instances: lines ~65, ~145)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

- [x] **`internal/template/server/barrier/rotation_handlers_test.go`** (line ~40)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

#### internal/cipher/

- [x] **`internal/cipher/server/testmain_test.go`** (line 90)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

- [x] **`internal/cipher/server/server_lifecycle_test.go`** (multiple instances)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

- [x] **`internal/cipher/server/realms/middleware_test.go`** (line ~70)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

#### internal/kms/server/

- [x] **`internal/kms/server/application/application_init_test.go`** (3 instances: lines 43, 118, 136)
  - Current: `&cryptoutilConfig.ServerSettings{...}` with partial fields
  - Fix: Replace with `cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)`

- [x] **`internal/kms/server/businesslogic/businesslogic_test.go`** (2 instances)
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

#### Good Example (Already Correct)

- ✅ **`internal/jose/server/server_test.go`**
  - Uses: `cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)`
  - Result: `BindPublicAddress="127.0.0.1"`, `BindPublicPort=0` → `"127.0.0.1:0"` → localhost only, no firewall

---

## Phase 2: validateConfiguration() Enhancement (Configuration Robustness)

**Current State**: `validateConfiguration()` at `internal/shared/config/config.go:1379` validates some fields but is incomplete.

**Goal**: Add comprehensive validation for ALL ServerSettings fields to catch configuration errors early.

### Missing Validations to Add

#### Network Configuration

- [x] **BindPublicAddress** - MUST NOT be blank (reject "", enforce "127.0.0.1" or "0.0.0.0")
- [x] **BindPublicPort** - MUST be 0-65535 (reject negative)
- [x] **BindPrivateAddress** - MUST NOT be blank (reject "", enforce "127.0.0.1")
- [x] **BindPrivatePort** - MUST be 0-65535 (reject negative)
- [x] **BindPublicProtocol** - MUST be "https" (reject "http" in production)
- [x] **BindPrivateProtocol** - MUST be "https" (reject "http" in production)

#### TLS Configuration

- [x] **TLSPublicDNSNames** - MUST contain at least one entry (reject empty array)
- [x] **TLSPublicIPAddresses** - MUST contain at least one entry (reject empty array)
- [x] **TLSPrivateDNSNames** - MUST contain at least one entry
- [x] **TLSPrivateIPAddresses** - MUST contain at least one entry

#### CORS Configuration

- [x] **CORSAllowedOrigins** - MUST contain at least one entry for browser endpoints
- [x] Each origin MUST start with "http://" or "https://"
- [x] Each origin MUST NOT have trailing slash

#### CSRF Configuration

- [x] **CSRFEnabled** - If true, validate CSRFCookieName is non-blank
- [x] **CSRFCookieName** - MUST NOT be blank if CSRF enabled

#### Rate Limiting

- [x] **RateLimitEnabled** - If true, validate rate limit values are positive
- [x] **RateLimitRequestsPerMinute** - MUST be positive if rate limiting enabled
- [x] **RateLimitBurst** - MUST be positive if rate limiting enabled

#### Database Configuration

- [x] **DatabaseURL** - MUST NOT be blank
- [x] **DatabaseURL** - MUST start with "postgres://" or "file:" for SQLite
- [x] **DatabaseMaxOpenConnections** - MUST be positive
- [x] **DatabaseMaxIdleConnections** - MUST be ≤ MaxOpenConnections

#### OTLP Configuration

- [x] **OTLPEnabled** - If true, validate OTLP endpoint is non-blank
- [x] **OTLPEndpoint** - MUST NOT be blank if OTLP enabled
- [x] **OTLPServiceName** - MUST NOT be blank if OTLP enabled
- [x] **OTLPServiceVersion** - MUST NOT be blank if OTLP enabled

#### Security Configuration

- [x] **SessionSecretKey** - MUST be ≥32 bytes for production (reject weak keys)
- [x] **CSPEnabled** - If true, validate CSP directives are non-blank
- [x] **DevMode** - WARN if true in production environment

### Validation Error Message Pattern

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

## Phase 3: Critical TestMain Files (Resource Cleanup)

TestMain functions initialize shared resources that MUST be cleaned up to prevent resource leaks across test runs.

**Common Missing Cleanup Calls**:

- `sqlDB.Close()`
- `telemetryService.Shutdown()`
- `jwkGenService.Shutdown()`
- `unsealKeysService.Shutdown()`
- `barrierRepo.Shutdown()`
- `barrierService.Shutdown()`
- PostgreSQL test-container cleanup

### internal/cipher/

- [x] **`internal/cipher/e2e/testmain_e2e_test.go`**
  - Missing: `sharedJWKGenService.Shutdown()`
  - Missing: `unsealService.Shutdown()` (local variable line 99)
  - Missing: `barrierRepo.Shutdown()` (local variable line 106)
  - Missing: `testBarrierService.Shutdown()`
  - Missing: `sqlDB.Close()` (db from initTestDB line 89)
  - Has: `sharedTelemetryService.Shutdown()` ✅
  - Has: `testPublicServer.Shutdown()` ✅

- [x] **`internal/cipher/server/testmain_test.go`**
  - Missing: `testJWKGenService.Shutdown()`
  - Missing: `unsealKeysService.Shutdown()` (local variable line 115)
  - Missing: `barrierRepo.Shutdown()` (local variable line 120)
  - Missing: `testBarrierService.Shutdown()`
  - Missing: `testSQLDB.Close()` (deferred comment warns against closing too early, but should defer at function start)
  - Has: `testTelemetryService.Shutdown()` ✅ (but too late - after m.Run())
  - Note: Comment at line 160+ warns about parallel tests, but cleanup should still happen via defer

- [x] **`internal/cipher/integration/testmain_integration_test.go`**
  - Missing: `sqlDB.Close()` (from gorm db.DB() at line 78)
  - Has: `sharedPGContainer.Terminate()` ✅
  - Note: Opens GORM connection but doesn't close underlying sql.DB

### internal/template/server/

- [x] **`internal/template/server/barrier/barrier_service_test.go`**
  - Missing: `testJWKGenService.Shutdown()` (line 89)
  - Missing: `testTelemetryService.Shutdown()` (line 83)
  - Missing: `unsealService.Shutdown()` (local variable line 100)
  - Missing: `barrierRepo.Shutdown()` (local variable line 106)
  - Has: `testBarrierService.Shutdown()` ✅ (line 128)
  - Has: `testSQLDB.Close()` ✅ (line 130)

---

## Phase 4: Per-Test Service Initialization Files (Resource Cleanup)

These files create services within individual test functions that need cleanup via t.Cleanup() or defer.

### internal/cipher/

- [x] **`internal/cipher/server/realms/middleware_test.go`**
  - Function: `createTestPublicServer()` (line 66)
  - Missing: `telemetryService.Shutdown()` (line 76)
  - Missing: `jwkGenService.Shutdown()` (line 79)
  - Missing: `unsealService.Shutdown()` (line 86)
  - Missing: `barrierRepo.Shutdown()` (line 89)
  - Missing: `barrierService.Shutdown()` (line 92)
  - Missing: `sqlDB.Close()` (from initTestDB)
  - Has: `srv.Shutdown()` via t.Cleanup() ✅ (line 150+)
  - Note: Services created per-test should be cleaned up in t.Cleanup()

- [x] **`internal/cipher/server/server_lifecycle_test.go`**
  - Function: Multiple test functions create services
  - Missing: `telemetryService.Shutdown()` (line 126)
  - Missing: `jwkGenService.Shutdown()` (line 129)
  - Note: Uses testBarrierService from TestMain (shared), but creates new telemetry/jwk services per-test

### internal/template/server/

- [x] **`internal/template/server/barrier/rotation_handlers_test.go`**
  - Function: `setupRotationTestEnvironment()` (line 30)
  - Missing: `testSQLDB.Close()` (line 38)
  - Missing: `telemetryService.Shutdown()` (line 69)
  - Missing: `jwkGenService.Shutdown()` (line 73)
  - Missing: `unsealService.Shutdown()` (line 80)
  - Missing: `barrierRepo.Shutdown()` (line 84)
  - Missing: `barrierService.Shutdown()` (line 88)
  - Missing: `rotationService` cleanup (if it has Shutdown method)
  - Note: Helper function creates many services but doesn't clean them up

- [x] **`internal/template/server/barrier/gorm_barrier_repository_test.go`**
  - Multiple test functions create barrierRepo
  - Missing: `barrierRepo.Shutdown()` calls for instances at:
    - Line 67 (TestGormBarrierRepository_StoreRootKey_Success)
    - Line 168 (TestGormBarrierRepository_LoadRootKey_Success)
    - Line 269 (TestGormBarrierRepository_StoreIntermediateKey_Success)
    - Line 366 (TestGormBarrierRepository_LoadIntermediateKey_Success)
    - Line 410 (TestGormBarrierRepository_StoreContentKey_Success)
  - Has: `sqlDB.Close()` via defer ✅ (line 51)

---

## Phase 5: Shared Testutil Pattern Validation (Verification)

These TestMain files delegate to testutil.Initialize(). Verify testutil cleanup.

### internal/template/server/

- [x] **`internal/template/server/test_main_test.go`**
  - Delegates to: `cryptoutilTemplateServerTestutil.Initialize()`
  - Action: Verify testutil has proper cleanup (TLS configs don't need cleanup)

- [x] **`internal/template/server/repository/test_main_test.go`**
  - Delegates to: `cryptoutilTemplateServerTestutil.Initialize()`
  - Action: Same as above

- [x] **`internal/template/server/listener/test_main_test.go`**
  - Delegates to: `cryptoutilTemplateServerTestutil.Initialize()`
  - Action: Same as above

---

## Phase 6: Review Remaining Test Files (Verification)

Scan remaining test files for any missed service initialization.

### internal/cipher/

- [x] `internal/cipher/server/middleware_test.go` - Review for service cleanup
- [x] `internal/cipher/server/helpers_test.go` - Review for service cleanup
- [x] `internal/cipher/server/realm_validation_test.go` - Review for service cleanup
- [x] `internal/cipher/server/realms/realm_validation_test.go` - Review for service cleanup
- [x] `internal/cipher/e2e/browser_e2e_test.go` - Review for service cleanup
- [x] `internal/cipher/e2e/service_e2e_test.go` - Review for service cleanup
- [x] `internal/cipher/e2e/rotation_e2e_test.go` - Review for service cleanup
- [x] `internal/cipher/e2e/helpers_e2e_test.go` - Review for service cleanup
- [x] `internal/cipher/integration/concurrent_test.go` - Review for service cleanup
- [x] `internal/cipher/repository/message_recipient_jwk_repository_test.go` - Review for service cleanup
- [ ] `internal/cipher/crypto/password_test.go` - Review for service cleanup (likely no services)

### internal/template/server/

- [x] `internal/template/server/service_template_test.go` - Review for service cleanup
- [x] `internal/template/server/application_test.go` - Review for service cleanup
- [x] `internal/template/server/listener/admin_test.go` - Review for service cleanup
- [x] `internal/template/server/listener/public_test.go` - Review for service cleanup
- [x] `internal/template/server/listener/servers_test.go` - Review for service cleanup
- [x] `internal/template/server/repository/public_table_test.go` - Review for service cleanup
- [ ] `internal/template/server/repository/application_table_test.go` - Review for service cleanup
- [ ] `internal/template/server/barrier/status_handlers_test.go` - Review for service cleanup

---

## Cleanup Patterns Reference

### Defer Execution Order (LIFO - Last In, First Out)

**CRITICAL**: Go executes defer statements in **reverse order** of declaration (stack-based LIFO).

Resources created FIRST should be deferred LAST (cleanup happens in reverse of creation order).

```go
// Example: Correct cleanup order (reverse of creation)
func TestMain(m *testing.M) {
    ctx := context.Background()

    // 1. Create database (FIRST created)
    sqlDB, _ := sql.Open("sqlite", dsn)
    defer sqlDB.Close()  // LAST cleaned up ✅

    // 2. Create telemetry service (depends on nothing)
    telemetrySvc, _ := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
    defer telemetrySvc.Shutdown()  // 5th cleaned up ✅

    // 3. Create JWK service (depends on telemetry)
    jwkGenSvc, _ := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
    defer jwkGenSvc.Shutdown()  // 4th cleaned up ✅

    // 4. Create unseal service (depends on nothing)
    unsealSvc, _ := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
    defer unsealSvc.Shutdown()  // 3rd cleaned up ✅

    // 5. Create barrier repository (depends on DB)
    barrierRepo, _ := cryptoutilBarrier.NewGormBarrierRepository(sqlDB)
    defer barrierRepo.Shutdown()  // 2nd cleaned up ✅

    // 6. Create barrier service (depends on telemetry, jwkGen, barrierRepo, unseal) (LAST created)
    barrierSvc, _ := cryptoutilBarrier.NewBarrierService(ctx, telemetrySvc, jwkGenSvc, barrierRepo, unsealSvc)
    defer barrierSvc.Shutdown()  // FIRST cleaned up ✅

    // Run tests
    exitCode := m.Run()

    // Defer stack executes here in REVERSE order:
    // 1. barrierSvc.Shutdown()      (last defer, first cleanup)
    // 2. barrierRepo.Shutdown()     (depends on barrierSvc being done)
    // 3. unsealSvc.Shutdown()       (independent)
    // 4. jwkGenSvc.Shutdown()       (no dependents remaining)
    // 5. telemetrySvc.Shutdown()    (no dependents remaining)
    // 6. sqlDB.Close()              (first defer, last cleanup - all DB users gone)

    os.Exit(exitCode)
}
```

### Pattern 1: TestMain Cleanup (Deferred Immediately After Creation)

```go
func TestMain(m *testing.M) {
    ctx := context.Background()

    // Setup resources with IMMEDIATE defer after each creation
    sqlDB, _ := sql.Open("sqlite", dsn)
    defer sqlDB.Close()  // Deferred FIRST (cleaned up LAST)

    telemetrySvc, _ := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
    defer telemetrySvc.Shutdown()

    jwkGenSvc, _ := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
    defer jwkGenSvc.Shutdown()

    unsealSvc, _ := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
    defer unsealSvc.Shutdown()

    barrierRepo, _ := cryptoutilBarrier.NewGormBarrierRepository(sqlDB)
    defer barrierRepo.Shutdown()

    barrierSvc, _ := cryptoutilBarrier.NewBarrierService(ctx, telemetrySvc, jwkGenSvc, barrierRepo, unsealSvc)
    defer barrierSvc.Shutdown()  // Deferred LAST (cleaned up FIRST)

    // Run tests
    exitCode := m.Run()

    os.Exit(exitCode)
    // Defers execute here: barrierSvc → barrierRepo → unsealSvc → jwkGenSvc → telemetrySvc → sqlDB
}
```

### Pattern 2: Per-Test Cleanup (t.Cleanup - Also LIFO)

**CRITICAL**: `t.Cleanup()` also executes in **reverse order** (LIFO), just like defer.

```go
func setupTestServices(t *testing.T) (*BarrierService, *sql.DB) {
    t.Helper()

    // Register cleanups in CREATION order
    // They will execute in REVERSE order (LIFO)

    sqlDB, _ := sql.Open("sqlite", dsn)
    t.Cleanup(func() { _ = sqlDB.Close() })  // Registered FIRST, runs LAST ✅

    telemetrySvc, _ := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
    t.Cleanup(func() { telemetrySvc.Shutdown() })

    jwkGenSvc, _ := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
    t.Cleanup(func() { jwkGenSvc.Shutdown() })

    barrierRepo, _ := cryptoutilBarrier.NewGormBarrierRepository(sqlDB)
    t.Cleanup(func() { barrierRepo.Shutdown() })

    barrierSvc, _ := cryptoutilBarrier.NewBarrierService(ctx, telemetrySvc, jwkGenSvc, barrierRepo, unsealSvc)
    t.Cleanup(func() { barrierSvc.Shutdown() })  // Registered LAST, runs FIRST ✅

    return barrierSvc, sqlDB
    // After test completes, cleanups run: barrierSvc → barrierRepo → jwkGenSvc → telemetrySvc → sqlDB
}
```

### Pattern 3: Inline Defer (Simple Cases)

```go
func TestSomething(t *testing.T) {
    // Defer immediately after creation for correct LIFO cleanup
    telemetrySvc, _ := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
    defer telemetrySvc.Shutdown()  // Cleaned up 2nd

    jwkGenSvc, _ := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
    defer jwkGenSvc.Shutdown()  // Cleaned up 1st (depends on telemetry)

    // ... test logic

    // Function exit: jwkGenSvc.Shutdown() → telemetrySvc.Shutdown()
}
```

### Dependency Graph Example

```
Creation Order (Top to Bottom):
1. sqlDB              (no dependencies)
2. telemetrySvc       (no dependencies)
3. jwkGenSvc          (depends on: telemetrySvc)
4. unsealSvc          (no dependencies)
5. barrierRepo        (depends on: sqlDB)
6. barrierSvc         (depends on: telemetrySvc, jwkGenSvc, barrierRepo, unsealSvc)

Cleanup Order (Bottom to Top - REVERSE):
6. barrierSvc.Shutdown()      (FIRST cleanup - has most dependencies)
5. barrierRepo.Shutdown()     (safe: barrierSvc already shut down)
4. unsealSvc.Shutdown()       (safe: barrierSvc already shut down)
3. jwkGenSvc.Shutdown()       (safe: barrierSvc already shut down)
2. telemetrySvc.Shutdown()    (safe: all dependents shut down)
1. sqlDB.Close()              (LAST cleanup - all DB users shut down)
```

---

## Priority Execution Order

1. **Phase 1 (CRITICAL - Windows Firewall)**: Migrate all test files to NewTestConfig - prevents firewall prompts
2. **Phase 2 (Medium - Configuration Robustness)**: Enhance validateConfiguration() - catches config errors early
3. **Phase 3 (High - Resource Leaks)**: Fix all TestMain files - affects ALL tests in the package
4. **Phase 4 (Medium - Resource Leaks)**: Fix per-test service initialization - prevents leaks in individual tests
5. **Phase 5 (Low - Verification)**: Verify testutil pattern - likely already correct
6. **Phase 6 (Low - Verification)**: Scan remaining files to catch any missed cleanup

---

## Success Criteria

- [ ] All test files use NewTestConfig() for ServerSettings (no Windows Firewall prompts)
- [ ] validateConfiguration() enhanced with comprehensive field validation
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
- [ ] Cleanup order follows LIFO pattern (reverse of creation order)

---

## Notes

- **Defer Order**: Go executes defer in LIFO (Last In, First Out) - reverse of declaration order
- **t.Cleanup Order**: Also LIFO - cleanup functions run in reverse of registration order
- **Dependency Safety**: LIFO ensures dependents are cleaned up before dependencies
- **TestMain Pattern**: Use defer immediately after successful creation
- **Per-Test Pattern**: Use t.Cleanup() for test-scoped resources (also LIFO)
- **Error Handling**: Cleanup functions should handle errors gracefully (log but don't fail)
- **Parallel Tests**: t.Cleanup() is safe for parallel tests; defer in TestMain runs after all tests complete
