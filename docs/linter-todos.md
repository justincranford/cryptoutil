# Linter TODOs

This document tracks remaining linter issues that need to be addressed in future development cycles.

**Last Updated:** September 21, 2025

## Summary

- **Total remaining issues:** 27 violations across 7 categories
- **Priority:** Medium to Low (no critical security issues)
- **Categories:** gosec (26 issues), stylecheck (1 issue)

## Remaining Issues by Category

### 1. G304: Potential file inclusion via variable (7 issues)

**Risk Level:** Medium - Potential security vulnerability if user input is not validated

- `cmd\pgtest\main.go:48:15`
- `cmd\pgtest\main.go:54:19` 
- `cmd\pgtest\main.go:60:17`
- `internal\common\crypto\asn1\der_pem.go:146:19`
- `internal\common\crypto\asn1\der_pem.go:155:19`
- `internal\common\util\files.go:35:19`
- `internal\common\util\files.go:76:15`

**Action Required:** Review file path handling to ensure proper validation and sanitization.

### 2. G404: Use of weak random number generator (7 issues)

**Risk Level:** Medium - Should use crypto/rand for cryptographic operations

- `internal\common\telemetry\telemetry_service_test.go:49:42`
- `internal\common\telemetry\telemetry_service_test.go:50:42`
- `internal\common\telemetry\telemetry_service_test.go:51:42`
- `internal\common\util\thread\chan_test.go:39:19`
- `internal\server\repository\sqlrepository\sql_provider.go:49:62`
- `internal\server\repository\sqlrepository\sql_provider.go:50:68`
- `internal\server\repository\sqlrepository\sql_provider.go:51:68`

**Action Required:** Replace math/rand with crypto/rand for security-sensitive operations.

### 3. G402: TLS Configuration Issues (5 issues)

**Risk Level:** Medium - TLS security concerns in test files

- `internal\client\client_test_util.go:77:24` - TLS InsecureSkipVerify set true
- `internal\common\crypto\certificate\certificates_test.go:40:22` - TLS MinVersion too low
- `internal\common\crypto\certificate\certificates_test.go:41:22` - TLS MinVersion too low  
- `internal\server\application\application_test.go:79:33` - TLS MinVersion too low
- `internal\server\application\application_test.go:85:25` - TLS InsecureSkipVerify set true

**Action Required:** Update TLS configurations to use stronger security settings, even in tests.

### 4. G115: Integer overflow conversion (5 issues)

**Risk Level:** Low to Medium - Potential integer overflow vulnerabilities

- `internal\common\config\config_test_util.go:37:38` - uint32 -> uint16
- `internal\common\config\config_test_util.go:44:38` - uint32 -> uint16
- `internal\common\util\random.go:93:29` - int64 -> uint64
- `internal\common\util\random.go:98:29` - int32 -> uint32
- `internal\common\util\random.go:103:29` - int16 -> uint16

**Action Required:** Add bounds checking before type conversions to prevent overflow.

### 5. ST1005: Error strings should not be capitalized (4 issues)

**Risk Level:** Low - Code style issue

- `internal\server\businesslogic\oam_orm_mapper.go:270:24`
- `internal\server\businesslogic\oam_orm_mapper.go:277:25`
- `internal\server\businesslogic\oam_orm_mapper.go:308:12`
- `internal\server\businesslogic\oam_orm_mapper.go:317:12`

**Action Required:** Lowercase the first letter of error strings.

### 6. G301: Expect directory permissions to be 0750 or less (2 issues)

**Risk Level:** Low - File system security

- `internal\common\crypto\asn1\der_pem.go:174:8`
- `internal\common\crypto\asn1\der_pem.go:193:8`

**Action Required:** Use more restrictive directory permissions (0750 or 0700).

### 7. ST1023: Should omit type func() from declaration (1 issue)

**Risk Level:** Low - Code style issue

- `internal\server\repository\sqlrepository\sql_provider.go:82:26`

**Action Required:** Remove explicit type declaration as it can be inferred.

## Recently Fixed Issues ✅

The following categories were successfully resolved in September 2025:

1. **errorlint** - Error format verbs and type assertions
2. **goimports** - Import organization
3. **goconst** - String constant extraction  
4. **misspell** - Spelling errors
5. **prealloc** - Slice pre-allocation
6. **unconvert** - Unnecessary type conversions
7. **ST1019** - Duplicate imports
8. **gosec G402** - Specific TLS MinVersion issues (client_test_util.go, application_listener.go)
9. **gosec G115** - Specific integer overflow issues (config_test_util.go, pool.go, combinations.go, random.go)

## Next Steps

1. **Priority 1 (Security):** Address G404 weak random number generators in production code
2. **Priority 2 (Security):** Fix G304 file inclusion vulnerabilities  
3. **Priority 3 (Configuration):** Update remaining G402 TLS settings and G115 integer conversions
4. **Priority 4 (Style):** Clean up ST1005 error strings and other style issues
5. **Priority 5 (Permissions):** Review G301 directory permissions

## How to Check Progress

Run individual category checks:

```bash
# Check specific categories
golangci-lint run --config .golangci.yml --enable-only=gosec --out-format=tab | findstr "G404"
golangci-lint run --config .golangci.yml --enable-only=gosec --out-format=tab | findstr "G304"
golangci-lint run --config .golangci.yml --enable-only=stylecheck --out-format=tab | findstr "ST1005"

# Full remaining issues check
golangci-lint run --config .golangci.yml --enable-only=gosec,stylecheck --out-format=tab
```