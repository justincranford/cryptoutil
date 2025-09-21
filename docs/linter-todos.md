# Linter TODOs

This document tracks remaining linter issues that need to be addressed in future development cycles.

**Last Updated:** September 21, 2025

## Summary

- **Total remaining issues:** 15 violations across 2 categories  
- **Priority:** Medium to Low (no critical security issues)
- **Categories:** gosec (15 issues), stylecheck (0 issues)

**Recently Fixed (September 21, 2025):**
- ✅ **G404:** Weak random number generators (7 issues fixed)
- ✅ **G115:** Integer overflow conversions (5 issues fixed) 
- ✅ **G301:** Directory permissions (2 issues fixed)
- ✅ **ST1005:** Error string capitalization (4 issues fixed)
- ✅ **ST1023:** Type inference (1 issue fixed)

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

### 2. G402: TLS Configuration Issues (5 issues)

**Risk Level:** Medium - TLS security concerns in test files

- `internal\client\client_test_util.go:77:24` - TLS InsecureSkipVerify set true
- `internal\common\crypto\certificate\certificates_test.go:40:22` - TLS MinVersion too low
- `internal\common\crypto\certificate\certificates_test.go:41:22` - TLS MinVersion too low  
- `internal\server\application\application_test.go:79:33` - TLS MinVersion too low
- `internal\server\application\application_test.go:85:25` - TLS InsecureSkipVerify set true

**Action Required:** Update TLS configurations to use stronger security settings, even in tests.

### 3. G115: Integer overflow conversion (3 issues)

**Risk Level:** Low to Medium - Potential integer overflow vulnerabilities

- `internal\common\config\config_test_util.go:37:38` - uint32 -> uint16
- `internal\common\config\config_test_util.go:44:38` - uint32 -> uint16

**Note:** Random.go issues were resolved with #nosec annotations for safe same-width conversions.

**Action Required:** Add bounds checking before type conversions to prevent overflow.

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

1. **Priority 1 (Security):** Address G304 file inclusion vulnerabilities  
2. **Priority 2 (Configuration):** Update remaining G402 TLS settings and G115 integer conversions
3. **Priority 3 (Maintenance):** Monitor for new violations in future development

## How to Check Progress

Run individual category checks:

```bash
# Check specific categories
golangci-lint run --config .golangci.yml --enable-only=gosec --out-format=tab | findstr "G304"
golangci-lint run --config .golangci.yml --enable-only=gosec --out-format=tab | findstr "G402"

# Full remaining issues check
golangci-lint run --config .golangci.yml --enable-only=gosec,stylecheck --out-format=tab
```