# Workflow Fixes Analysis and Test Coverage Gap Plan

## Executive Summary

This document analyzes the root causes of issues fixed during recent workflow troubleshooting and identifies comprehensive test coverage gaps that allowed these issues to slip through. The goal is to shift-left by adding robust unit and integration tests that would catch these categories of problems before they reach CI/CD workflows.

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
- `dev: true` → SQLite (implicit) + MUST use 127.0.0.1 (security restriction)
- Production → PostgreSQL (explicit URL) + MAY use 0.0.0.0

No explicit path existed for: Container networking (0.0.0.0) + SQLite database.

**Validation Logic**:
- Dev mode + 0.0.0.0 → FAIL (intentional security restriction for local development)
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
✅ Good unit test coverage (6 test cases)
- dev mode → SQLite in-memory
- sqlite:// in-memory URL
- sqlite:// file-based URL
- postgres:// URL parsing
- Unsupported URL schemes
- Empty URL error handling

**Missing Tests**:
❌ Integration tests validating config validation interactions:
- Container mode (0.0.0.0) + SQLite URL should pass validation
- Container mode (0.0.0.0) + dev mode should fail validation
- Container mode (0.0.0.0) + PostgreSQL URL should pass validation

❌ Integration tests for complete configuration flow:
- Load config from YAML → validate → map database type → initialize DB

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
- Result: Healthcheck failures → containers marked unhealthy → workflow failures

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
❌ **ZERO tests for mTLS configuration logic** (most critical gap)
- No tests verifying `DevMode → NoClientCert`
- No tests verifying container mode detection (`0.0.0.0` → `NoClientCert`)
- No tests verifying production mode → `RequireAndVerifyClientCert`
- No tests for private vs public server TLS client auth differences

✅ Some middleware tests exist (`service_auth_test.go`) but DON'T test application-level TLS configuration

**Missing Tests (HIGH PRIORITY)**:
❌ Unit tests for container mode detection:
- `BindPublicAddress == "0.0.0.0"` → `isContainerMode = true`
- `BindPublicAddress == "127.0.0.1"` → `isContainerMode = false`
- `BindPrivateAddress == "0.0.0.0"` → should NOT affect container mode detection

❌ Unit tests for mTLS configuration:
- `DevMode=true` → `privateClientAuth = NoClientCert`
- Container mode (0.0.0.0) → `privateClientAuth = NoClientCert`
- Production (127.0.0.1, DevMode=false) → `privateClientAuth = RequireAndVerifyClientCert`
- Public server should NEVER use RequireAndVerifyClientCert (browser compatibility)

❌ Integration tests:
- Container mode with SQLite → healthcheck succeeds (wget without client cert)
- Container mode with PostgreSQL → healthcheck succeeds
- Production mode → healthcheck requires client cert OR uses different healthcheck method

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
      echo "❌ Health check failed - Diagnostic output:"
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
YAML field mapping may not handle kebab-case → PascalCase conversion correctly:
- `dev-mode` (YAML kebab-case) should map to `DevMode` (Go struct PascalCase)
- Other fields work: `bind-public-address` → `BindPublicAddress` ✅
- But `dev-mode` → `DevMode` appears to fail ❌

Possible issues:
1. Viper library mapstructure tag mismatch
2. Field name collision or override
3. Config loading order issue (CLI flag overriding YAML value)
4. Boolean field handling issue

### Investigation Needed
- Verify YAML → struct field mapping for all boolean fields
- Check if other kebab-case fields map correctly
- Test config loading from file vs CLI flags vs environment variables
- Add debug logging to show config loading steps

### Test Coverage Gap Analysis

**Existing Tests**:
❌ **ZERO tests for config file loading** (all existing tests use in-memory configs)
- No tests loading config from actual YAML files
- No tests verifying kebab-case → PascalCase field mapping
- No tests checking config precedence (file vs CLI vs env vars)

**Missing Tests**:
❌ YAML config loading tests:
- Load config from YAML file with kebab-case field names
- Verify all fields map correctly to struct (especially boolean fields)
- Test all casing styles: kebab-case, camelCase, PascalCase, snake_case

❌ Config precedence tests:
- YAML file with `dev-mode: true` + CLI flag `-d=false` → should use CLI value
- Environment variable overrides
- Default value fallbacks

❌ Field validation tests:
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
- Container mode detection (0.0.0.0 → isContainerMode)
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
- Complete config flow: load → validate → map DB → initialize
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
   - Verify kebab-case → PascalCase field mapping
   - Test boolean field parsing (dev-mode: true)

### Short-Term Actions (P1)

4. **Add Config Validation Integration Tests**
   - Valid combinations: container+SQLite, container+PostgreSQL, dev+SQLite, production+PostgreSQL
   - Invalid combinations: dev+0.0.0.0, missing required fields

5. **Add Database URL Parsing Tests**
   - SQLite with query parameters (?cache=shared&mode=rwc)
   - Absolute file paths (/var/lib/cryptoutil/db.sqlite)

6. **Add Healthcheck Integration Tests**
   - livez endpoint accessible without client cert
   - readyz endpoint dependency validation
   - Docker healthcheck commands (wget) succeed

### Long-Term Actions (P2)

7. **Add TLS Client Auth Integration Tests**
   - Container mode: wget healthcheck succeeds (no client cert)
   - Production mode: admin endpoints require client cert
   - Public endpoints never require client cert

8. **Add Config Precedence Tests**
   - YAML < Environment Variables < CLI Flags
   - Default value fallbacks
   - Config hot-reload (if supported)

9. **Add E2E Docker Tests**
   - Docker Compose stack startup
   - Healthcheck passes in containers
   - Service-to-service communication with/without mTLS

---

## Success Criteria

**Test Coverage Targets**:
- mTLS configuration logic: ≥95% coverage (currently 0%)
- Container mode detection: ≥95% coverage
- Config validation: ≥95% coverage
- Database URL mapping: ≥98% coverage (currently ~85%)

**Integration Test Goals**:
- All valid config combinations tested
- All invalid config combinations tested
- Container mode + SQLite integration verified
- Container mode + PostgreSQL integration verified

**Quality Gates**:
- All P0 tests implemented and passing
- Mutation testing ≥85% on affected modules
- No new TODOs or FIXMEs in test files
- All tests run in <15 seconds per package

**Workflow Impact**:
- DAST workflow should pass consistently
- Load testing workflow should pass consistently
- No config-related failures in CI/CD
