# Linter TODOs

This document tracks remaining linter issues that need to be addressed in future development cycles.

**Last Updated:** September 21, 2025

## Summary

- **Total remaining issues:** 115 violations across 11 categories  
- **Priority:** Low to Medium (mostly code quality improvements)
- **Top categories:** dupl (43), errcheck (29), gocyclo (17), unparam (15)

**Recently Fixed (September 21, 2025):**
- ✅ **G404:** Weak random number generators (7 issues fixed)
- ✅ **G115:** Integer overflow conversions (5 issues fixed) 
- ✅ **G301:** Directory permissions (2 issues fixed)
- ✅ **ST1005:** Error string capitalization (4 issues fixed)
- ✅ **ST1023:** Type inference (1 issue fixed)

## Remaining Issues by Category

### 1. dupl: Duplicate code blocks (43 issues)

**Risk Level:** Low - Code quality and maintainability  
**Impact:** Increased maintenance burden, potential inconsistencies

**Files with duplicates:**
- `internal\client\client_oam_mapper.go` (5 duplicates)
- `internal\common\crypto\certificate\serial_number.go` (2 duplicates) 
- `internal\common\crypto\jose\*.go` (multiple files, ~20 duplicates)
- `internal\common\telemetry\telemetry_service.go` (6 duplicates)
- `internal\server\handler\oam_oas_mapper.go` (8 duplicates)
- `internal\server\repository\orm\orm_transaction.go` (2 duplicates)
- `internal\server\repository\sqlrepository\sql_transaction.go` (2 duplicates)

**Action Required:** Refactor duplicate code into shared functions or utilities.

### 2. errcheck: Unchecked error returns (29 issues)

**Risk Level:** Medium - Potential silent failures  
**Impact:** Errors may go unnoticed, affecting reliability

**Key areas:**
- `internal\common\config\config.go` (8 issues) - Configuration parsing
- `internal\common\crypto\jose\*.go` (8 issues) - JOSE header setting
- `internal\common\crypto\certificate\certificates_server_test_util.go` (7 issues) - Test utilities
- `internal\common\crypto\asn1\der_pem.go` (1 issue) - File operations  
- `internal\server\application\application_listener.go` (1 issue) - HTTP request creation
- `internal\server\repository\sqlrepository\sql_provider.go` (3 issues) - Random number generation

**Action Required:** Add proper error handling for all returned errors.

### 3. gocyclo: High cyclomatic complexity (17 issues)

**Risk Level:** Medium - Code maintainability and testability  
**Impact:** Functions are complex, harder to test and maintain

**Functions with high complexity:**
- `internal\common\telemetry\telemetry_service.go:130` - TelemetryService.Shutdown (complexity 29)
- `internal\common\crypto\jose\jwe_jwk_util.go:37` - CreateJweJwkFromKey (complexity 24)
- `internal\common\crypto\jose\jwkgen_service.go:45` - NewJwkGenService (complexity 23)
- `internal\common\crypto\jose\jws_jwk_util.go:36` - CreateJwsJwkFromKey (complexity 23)
- `internal\server\repository\sqlrepository\sql_provider.go:73` - NewSqlRepository (complexity 22)

**Action Required:** Break down complex functions into smaller, more focused functions.

### 4. unparam: Unused parameters (15 issues)

**Risk Level:** Low - Code cleanliness  
**Impact:** Dead code, interface clarity

**Key areas:**
- `internal\common\crypto\certificate\certificates_verify_test_util.go` (2 issues) - Test utilities
- `internal\common\crypto\keygenpooltest\keygenpools_test.go` (2 issues) - Test parameters
- `internal\common\telemetry\telemetry_service.go` (3 issues) - Service initialization
- `internal\server\businesslogic\*.go` (5 issues) - Business logic mappers
- `internal\server\handler\oam_oas_mapper.go` (2 issues) - Response mappers
- `internal\server\repository\sqlrepository\sql_transaction.go` (1 issue) - Transaction creation

**Action Required:** Remove unused parameters or mark with `_` if part of interface.

### 5. unused: Unused variables/functions (4 issues)

**Risk Level:** Low - Code cleanliness  
**Impact:** Dead code

**Issues:**
- `internal\common\crypto\certificate\certificates_verify_test_util.go:55` - `verifyCertChain` function
- `internal\server\barrier\barrier_service_test.go:26` - `testDbType` variable
- `internal\server\barrier\contentkeysservice\content_keys_service_test.go:30` - `testDbType` variable  
- `internal\server\repository\orm\orm_repository.go:18` - `ormEntities` variable

**Action Required:** Remove unused code or mark as intentionally unused.

### 6. ineffassign: Ineffectual assignments (2 issues)

**Risk Level:** Low - Code logic errors  
**Impact:** Variables assigned but never used

**Issues:**
- `internal\server\businesslogic\businesslogic.go:350` - err assignment
- `internal\server\businesslogic\businesslogic.go:387` - err assignment

**Action Required:** Remove unnecessary assignments or use the values.

### 7. gosec: Security issues (1 issue)

**Risk Level:** Medium - Security vulnerability  
**Impact:** TLS security in tests

**Issue:**
- `internal\client\client_test_util.go:77:24` - G402: TLS InsecureSkipVerify set true

**Action Required:** Review TLS configuration for tests.

### 8. gosimple: Code simplification (1 issue)

**Risk Level:** Low - Code style  
**Impact:** Code readability

**Issue:**
- `internal\common\pool\pool.go:308` - S1000: should use simple channel send/receive instead of select

**Action Required:** Simplify channel operations.

### 9. staticcheck: Static analysis (1 issue)

**Risk Level:** Low - Code logic  
**Impact:** Unreachable code

**Issue:**
- `internal\common\crypto\jose\jws_message_util.go:161` - SA4004: surrounding loop is unconditionally terminated

**Action Required:** Fix loop logic or remove unreachable code.

### 10. bodyclose: HTTP response body handling (1 issue)

**Risk Level:** Low - Resource leak  
**Impact:** Potential memory leaks in tests

**Issue:**
- `internal\server\application\application_test.go:91:24` - response body must be closed

**Action Required:** Add proper defer body.Close() statements.

### 11. whitespace: Formatting (1 issue)

**Risk Level:** Low - Code formatting  
**Impact:** Consistency

**Issue:**
- `internal\common\crypto\keygenpooltest\keygenpools_test.go:143` - unnecessary trailing newline

**Action Required:** Remove trailing whitespace.

## Recently Fixed Issues ✅

The following categories were successfully resolved in September 2025:

**First Wave (Original 9 categories):**
1. **errorlint** - Error format verbs and type assertions
2. **goimports** - Import organization
3. **goconst** - String constant extraction  
4. **misspell** - Spelling errors
5. **prealloc** - Slice pre-allocation
6. **unconvert** - Unnecessary type conversions
7. **ST1019** - Duplicate imports
8. **gosec G402** - Specific TLS MinVersion issues (client_test_util.go, application_listener.go)
9. **gosec G115** - Specific integer overflow issues (config_test_util.go, pool.go, combinations.go, random.go)

**Second Wave (Additional 5 categories):**
10. **G404** - Weak random number generators (7 issues) ✅
11. **G115** - Remaining integer overflow conversions (5 issues) ✅  
12. **G301** - Directory permissions (2 issues) ✅
13. **ST1005** - Error string capitalization (4 issues) ✅
14. **ST1023** - Type inference (1 issue) ✅

## Next Steps

### Priority Levels

**Priority 1 (Critical/Security):**
- `gosec` issues (1 remaining) - TLS configuration

**Priority 2 (Code Quality/Reliability):**  
- `errcheck` issues (29) - Missing error handling
- `gocyclo` issues (17) - High complexity functions
- `ineffassign` issues (2) - Logic errors

**Priority 3 (Maintenance/Style):**
- `dupl` issues (43) - Code duplication  
- `unparam` issues (15) - Unused parameters
- `unused` issues (4) - Dead code
- `gosimple` (1) - Code simplification
- `staticcheck` (1) - Static analysis
- `bodyclose` (1) - Resource management
- `whitespace` (1) - Formatting

### Recommended Approach

1. **Security First:** Address the remaining `gosec` TLS issue
2. **Reliability:** Fix `errcheck` issues systematically by file/package
3. **Maintainability:** Refactor high `gocyclo` functions  
4. **Cleanup:** Address `dupl`, `unparam`, and `unused` issues during regular development

## How to Check Progress

Run category-specific checks:

```bash
# Check specific categories
golangci-lint run --config .golangci.yml --enable-only=dupl --out-format=tab
golangci-lint run --config .golangci.yml --enable-only=errcheck --out-format=tab  
golangci-lint run --config .golangci.yml --enable-only=gocyclo --out-format=tab
golangci-lint run --config .golangci.yml --enable-only=gosec --out-format=tab

# Get breakdown by linter
golangci-lint run --config .golangci.yml --out-format=json | ConvertFrom-Json | ForEach-Object { $_.Issues } | Group-Object FromLinter | Sort-Object Count -Descending

# Full check
golangci-lint run --config .golangci.yml --out-format=tab
```