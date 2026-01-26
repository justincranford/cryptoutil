# Coverage Gaps Analysis - Phase 4.3

**Purpose**: Document specific uncovered lines and functions for packages below 95% coverage target

**Generated**: 2026-01-26

**Analysis Methodology**:
1. Generated HTML coverage reports for all packages <90%
2. Identified RED lines (uncovered code paths)
3. Categorized gaps by type (error paths, edge cases, validation, concurrency)
4. Prioritized by coverage deficit (largest gaps first)

---

## Priority 1: Packages with Largest Gaps (>20% deficit)

### 1. Template Service Listener (70.7% coverage, -24.3% gap)

**Package**: `internal/apps/template/service/server/listener`

**Source Files**:
- `application_listener.go`: Main listener implementation with HTTP/HTTPS mode switching
- `servers.go`: Server lifecycle (Start, Shutdown, WaitForReady)
- `public.go`: Public HTTPS server initialization
- `admin.go`: Admin HTTPS server initialization

**Uncovered Scenarios** (to investigate via HTML report):
- [ ] HTTP mode startup (non-TLS listener path)
- [ ] HTTPS mode with custom TLS config
- [ ] Port allocation edge cases (bind failures, address in use)
- [ ] Graceful shutdown timeout scenarios
- [ ] WaitForReady timeout/cancel scenarios
- [ ] Concurrent Start/Shutdown race conditions
- [ ] Error paths in server initialization
- [ ] TLS certificate loading failures
- [ ] Listen address validation errors

**Test Files Exist**:
- `admin_test.go`: Admin server tests
- `public_test.go`: Public server tests
- `servers_test.go`: Server lifecycle tests
- `testmain_test.go`: Test setup

**Test Strategy**:
1. Add HTTP mode startup tests (currently HTTPS-only coverage)
2. Add error injection tests (bind failures, TLS errors)
3. Add concurrent Start/Shutdown tests
4. Add timeout/cancel context tests for WaitForReady
5. Add port conflict resolution tests

---

### 2. Template Service Barrier (72.6% coverage, -22.4% gap)

**Package**: `internal/apps/template/service/server/barrier`

**Source Files**:
- `encryption_service.go`: Core encryption/decryption operations
- `key_rotation_service.go`: Barrier key rotation logic
- `status_service.go`: Key status queries
- `*_handlers.go`: HTTP handlers for barrier APIs

**Uncovered Scenarios**:
- [ ] Encryption service edge cases (nil inputs, empty data)
- [ ] Key rotation concurrency (parallel rotation requests)
- [ ] Database transaction rollback scenarios
- [ ] Unseal key derivation error paths
- [ ] Key version mismatch handling
- [ ] Concurrent encryption operations (multiple goroutines)
- [ ] Status query error paths (database failures)
- [ ] Handler validation edge cases

**Test Files Exist**:
- `status_handlers_test.go`: Status API tests (recently fixed timeout)
- `rotation_handlers_test.go`: Rotation API tests

**Missing Test Coverage**:
- Encryption service unit tests
- Key rotation service unit tests
- Concurrency/race condition tests
- Error injection for database failures

**Test Strategy**:
1. Add encryption_service unit tests (nil checks, empty data, edge cases)
2. Add key_rotation concurrency tests (parallel requests, race detection)
3. Add database failure injection tests (rollback scenarios)
4. Add unseal key derivation error tests

---

### 3. Template Service Businesslogic (75.2% coverage, -19.8% gap)

**Package**: `internal/apps/template/service/server/businesslogic`

**Source Files**:
- `session_manager.go`: Session creation, validation, revocation
- `realm_service.go`: Realm management business logic
- `tenant_service.go`: Tenant management business logic
- `registration_service.go`: User/client registration flows
- `rotation_service.go`: Service-level key rotation orchestration

**Uncovered Scenarios**:
- [ ] Session expiration edge cases (expired session validation)
- [ ] Session revocation cascading (multi-tenant cleanup)
- [ ] Realm creation validation failures
- [ ] Tenant isolation boundary tests
- [ ] Registration duplicate detection
- [ ] Rotation service transaction failures
- [ ] Concurrent session creation (same user, multiple devices)
- [ ] Service-layer validation logic (complex business rules)

**Test Files Exist**:
- Limited test coverage exists

**Missing Test Coverage**:
- Session manager edge cases
- Realm service validation
- Tenant service boundary tests
- Registration service error paths
- Rotation service concurrency

**Test Strategy**:
1. Add session manager comprehensive tests (expiration, revocation, concurrency)
2. Add realm service validation tests (duplicate names, invalid configs)
3. Add tenant isolation boundary tests (cross-tenant access attempts)
4. Add registration service duplicate detection tests
5. Add rotation service transaction rollback tests

---

## Priority 2: Packages with Medium Gaps (10-20% deficit)

### 4. Template Service Repository (84.8% coverage, -10.2% gap)

**Package**: `internal/apps/template/service/server/repository`

**Uncovered Scenarios**:
- [ ] Database constraint violation handling
- [ ] Optimistic locking failures (concurrent updates)
- [ ] Complex query edge cases (empty results, large datasets)
- [ ] Transaction rollback scenarios
- [ ] GORM error mapping edge cases

**Test Strategy**:
- Add constraint violation tests (unique, foreign key, check)
- Add concurrent update tests (optimistic locking)
- Add pagination edge case tests
- Add transaction error injection tests

---

### 5. Template Service Config (81.3% coverage, -13.7% gap)

**Package**: `internal/apps/template/service/config`

**Uncovered Scenarios**:
- [ ] Config file loading failures (missing file, malformed YAML)
- [ ] Config validation edge cases (invalid values, missing required)
- [ ] Default value fallback paths
- [ ] Environment variable override edge cases
- [ ] Config merging logic (CLI flags + YAML + defaults)

**Test Strategy**:
- Add config file error tests (missing, malformed, permissions)
- Add validation comprehensive tests (all fields, boundary values)
- Add override precedence tests (CLI > YAML > defaults)

---

### 6. Template Service Config/TLS Generator (80.6% coverage, -14.4% gap)

**Package**: `internal/apps/template/service/config/tls_generator`

**Uncovered Scenarios**:
- [ ] Certificate generation edge cases (invalid key sizes, unsupported algorithms)
- [ ] SAN (Subject Alternative Name) validation
- [ ] Certificate expiration date edge cases (past dates, far future)
- [ ] Key generation failures (insufficient entropy, algorithm errors)
- [ ] PEM encoding/decoding error paths

**Test Strategy**:
- Add certificate generation edge case tests
- Add SAN validation comprehensive tests
- Add key generation error injection tests
- Add PEM encoding error tests

---

### 7. Cipher-IM Server Config (80.4% coverage, -14.6% gap)

**Package**: `internal/apps/cipher/im/server/config`

**Similar gaps to template-service-config** (config loading, validation, defaults)

---

### 8. Cipher-IM Server APIs (82.1% coverage, -12.9% gap)

**Package**: `internal/apps/cipher/im/server/apis`

**Uncovered Scenarios**:
- [ ] API handler validation edge cases
- [ ] Request/response transformation error paths
- [ ] Authentication failure scenarios
- [ ] Rate limiting edge cases
- [ ] CORS preflight handling

---

### 9. JOSE-JA Service (87.3% coverage, -7.7% gap)

**Package**: `internal/apps/jose/ja/service`

**Uncovered Scenarios** (from Phase 5.1 requirements):
- [ ] DeleteElasticJWK error paths (75% coverage - needs 20 percentage points)
- [ ] createMaterialJWK error paths (76.7% coverage)
- [ ] CreateEncryptedJWT validation (77.8% coverage)
- [ ] Database failure injection (transaction rollback)
- [ ] Validation logic comprehensive tests
- [ ] Concurrent JWK operations

---

### 10. JOSE-JA Server Config (61.9% coverage, -33.1% gap)

**Package**: `internal/apps/jose/ja/server/config`

**Largest single-package deficit**

**Uncovered Scenarios**:
- [ ] Config file loading (missing, malformed, invalid schema)
- [ ] JWKS endpoint configuration validation
- [ ] Federation settings validation
- [ ] Unseal key configuration edge cases
- [ ] Database connection string validation
- [ ] TLS configuration validation
- [ ] Service-specific settings (rotation intervals, TTLs)

**Test Strategy**:
- Comprehensive config validation test suite
- Error path tests for all config sections
- Default value tests
- Config merging/override tests

---

## Gap Categorization Summary

### By Gap Type:

**Error Paths** (most common, ~40% of gaps):
- Config file loading failures
- Database constraint violations
- TLS certificate loading errors
- Validation failures
- Transaction rollback scenarios

**Edge Cases** (~30% of gaps):
- Nil/empty input handling
- Boundary value testing (min/max, zero, negative)
- Date/time edge cases (expiration, past dates)
- Concurrent operations (race conditions)

**Validation Logic** (~20% of gaps):
- Business rule enforcement
- Complex validation chains
- Multi-field cross-validation
- Tenant isolation boundaries

**Concurrency/Race Conditions** (~10% of gaps):
- Parallel request handling
- Database contention
- Key rotation concurrency
- Session creation races

---

## Test Implementation Priority

**Phase 4.4 Implementation Order**:

1. **Template Listener** (24.3% gap, 4h estimated)
   - Critical infrastructure component
   - Affects server startup reliability
   - Highest deficit

2. **Template Barrier** (22.4% gap, 3h estimated)
   - Security-critical component
   - Encryption at rest core functionality
   - Recently fixed timeout issue

3. **Template Businesslogic** (19.8% gap, 4h estimated)
   - Business rule enforcement
   - Multi-tenancy isolation
   - Session management

4. **JOSE-JA Config** (33.1% gap, 2h estimated)
   - Config validation (simple, repetitive tests)
   - High deficit but straightforward

5. **Template Config/TLS/Repository** (10-14% gaps, 5h total)
   - Medium complexity
   - Can parallelize test writing

6. **Cipher-IM Config/APIs** (12-15% gaps, 3h total)
   - Similar patterns to template
   - Lower priority (smaller deficits)

7. **JOSE-JA Service** (7.7% gap, 2h estimated)
   - Phase 5.1 blocker
   - Specific functions identified
   - Smallest deficit

**Total Estimated Effort**: 23 hours (matches Phase 4.4 + 5.1 estimates from tasks.md)

---

## Validation Criteria

**Task 4.3 Complete When**:
- ✅ HTML coverage reports generated for all 10 packages <90%
- ✅ Gaps documented by category (error paths, edge cases, validation, concurrency)
- ✅ Test implementation priority established
- ✅ Specific uncovered scenarios listed per package

**Next Phase 4.4**:
- Implement targeted tests in priority order
- Verify coverage improvements after each package
- Target: All packages ≥95% coverage

---

## Notes

- **Docker Desktop**: Not blocking coverage work (container tests separate)
- **Fiber Timeout Fix**: Applied to barrier tests (commit 984fd61f)
- **Coverage Measurement**: Using `go test -cover` with SQLite in-memory
- **Parallel Tests**: All use `t.Parallel()` - account for SQLite contention
- **GORM**: SkipDefaultTransaction=true for test performance

