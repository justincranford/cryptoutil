# Windows Firewall Root Cause Analysis

**Date**: 2026-01-02  
**Status**: ✅ ROOT CAUSE CONFIRMED  
**Impact**: HIGH - 20+ test files affected  
**Fix Priority**: CRITICAL  

---

## Executive Summary

**Root Cause**: Test files creating `&cryptoutilConfig.ServerSettings{}` with blank `BindPublicAddress=""` and `BindPublicPort=0` causes `fmt.Sprintf("%s:%d", "", 0)` → `":0"` bind address → `net.Listen("tcp", ":0")` binds to **all network interfaces (0.0.0.0)** → Windows Firewall exception prompt.

**Solution**: All test files MUST use `NewTestConfig(bindAddr, bindPort, devMode)` from `internal/shared/config/config_test_helper.go` to ensure safe defaults (`BindPublicAddress="127.0.0.1"`, `BindPublicPort=0` → `"127.0.0.1:0"` → localhost-only binding, no firewall prompt).

---

## Technical Analysis

### Critical Code Path

```go
// internal/template/server/listener/public.go line 168:
listener, err := listenConfig.Listen(ctx, "tcp", fmt.Sprintf("%s:%d", s.settings.BindPublicAddress, s.settings.BindPublicPort))
```

### String Formatting Behavior

```go
// Unsafe Pattern (triggers Windows Firewall):
BindPublicAddress = ""  // blank/empty string
BindPublicPort = 0
fmt.Sprintf("%s:%d", "", 0) == ":0"
net.Listen("tcp", ":0") == binds to 0.0.0.0:0 (ALL INTERFACES)
→ Windows Firewall exception prompt!

// Safe Pattern (localhost only):
BindPublicAddress = "127.0.0.1"
BindPublicPort = 0
fmt.Sprintf("%s:%d", "127.0.0.1", 0) == "127.0.0.1:0"
net.Listen("tcp", "127.0.0.1:0") == binds to loopback only
→ No firewall prompt ✅
```

### Go net.Listen Behavior

| Bind Address | Actual Binding | Windows Firewall? |
|--------------|----------------|-------------------|
| `":0"` | 0.0.0.0:dynamic_port (all interfaces) | ✅ Prompts |
| `"0.0.0.0:0"` | 0.0.0.0:dynamic_port (all interfaces) | ✅ Prompts |
| `"127.0.0.1:0"` | 127.0.0.1:dynamic_port (localhost only) | ❌ No prompt |
| `"[::1]:0"` | ::1:dynamic_port (IPv6 localhost) | ❌ No prompt |

**Key Insight**: Empty string in bind address defaults to all interfaces!

---

## Affected Files (20+)

### High Priority (Trigger Server Startup)

1. **internal/template/server/listener/servers_test.go** (line 15)
2. **internal/cipher/server/testmain_test.go** (line 90)
3. **internal/kms/server/application/application_init_test.go** (lines 43, 118, 136)
4. **internal/template/server/service_template_test.go** (~line 30-50)
5. **internal/template/server/barrier/barrier_service_test.go** (2 instances)
6. **internal/template/server/barrier/rotation_handlers_test.go** (~line 40)
7. **internal/cipher/server/server_lifecycle_test.go** (multiple instances)
8. **internal/cipher/server/realms/middleware_test.go** (~line 70)
9. **internal/kms/server/businesslogic/businesslogic_test.go** (2 instances)

### Low Priority (Configuration Testing Only)

10. **internal/shared/config/url_test.go** (6 instances) - Tests URL generation, doesn't start servers
11. **internal/shared/config/config_test.go** (5 instances) - Tests config validation, doesn't start servers

### Good Example (Already Correct)

✅ **internal/jose/server/server_test.go** - Uses `NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)`

---

## Solution: NewTestConfig Helper

### Helper Location

`internal/shared/config/config_test_helper.go` (lines 15-100)

### Usage Pattern

```go
import cryptoutilConfig "cryptoutil/internal/shared/config"
import cryptoutilMagic "cryptoutil/internal/shared/magic"

// ✅ CORRECT: Use NewTestConfig with explicit bind address
settings := cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)
// Returns fully-populated ServerSettings with:
// - BindPublicAddress: "127.0.0.1"
// - BindPublicPort: 0 (dynamic allocation)
// - BindPrivateAddress: "127.0.0.1"
// - BindPrivatePort: 0
// - Plus ~50 other fields with safe defaults
// - Includes call to validateConfiguration()

// ❌ WRONG: Partial ServerSettings creation
settings := &cryptoutilConfig.ServerSettings{
    BindPublicAddress: "127.0.0.1",  // Only 8-12 fields set
    BindPublicPort: 0,                // Leaves ~40 fields as zero values
    // Missing TLS, CORS, CSRF, database, OTLP, etc.
}

// ❌ WORST: Blank bind address (triggers firewall!)
settings := &cryptoutilConfig.ServerSettings{
    BindPublicPort: 0,  // BindPublicAddress defaults to "" → ":0" → 0.0.0.0
}
```

### NewTestConfig Benefits

1. **Safe Defaults**: All ~50 configuration fields properly initialized
2. **No Firewall Prompts**: `BindPublicAddress="127.0.0.1"` prevents 0.0.0.0 binding
3. **Validation**: Calls `validateConfiguration()` before returning
4. **Consistency**: Single source of truth for test configuration
5. **Maintainability**: Changes to defaults apply to all tests automatically

---

## Validation Enhancement (Phase 6)

### Current Gap

`validateConfiguration()` at `internal/shared/config/config.go:1379` does NOT validate bind addresses are non-blank.

### Recommended Additions

```go
func (s *ServerSettings) validateConfiguration() error {
    var errs []string

    // Network validation (CRITICAL - prevents firewall prompts)
    if s.BindPublicAddress == "" {
        errs = append(errs, "BindPublicAddress MUST NOT be blank (use '127.0.0.1' or '0.0.0.0')")
    }
    if s.BindPrivateAddress == "" {
        errs = append(errs, "BindPrivateAddress MUST NOT be blank (use '127.0.0.1')")
    }
    if s.BindPublicPort < 0 || s.BindPublicPort > 65535 {
        errs = append(errs, fmt.Sprintf("BindPublicPort must be 0-65535, got: %d", s.BindPublicPort))
    }
    if s.BindPrivatePort < 0 || s.BindPrivatePort > 65535 {
        errs = append(errs, fmt.Sprintf("BindPrivatePort must be 0-65535, got: %d", s.BindPrivatePort))
    }

    // Additional validations for TLS, CORS, database, OTLP, etc.
    // See SERVICE-TEMPLATE-v2.md Phase 6 for complete checklist

    if len(errs) > 0 {
        return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errs, "\n  - "))
    }
    return nil
}
```

**Benefit**: Early detection of blank bind addresses before server startup, failing fast with clear error message.

---

## Migration Checklist

See `docs/cipher-im-migration/SERVICE-TEMPLATE-v2.md` for detailed tracking:

- **Phase 5**: NewTestConfig migration (20+ files)
  - Priority: CRITICAL - Windows Firewall root cause
  - Files grouped by package and priority
  - Good example reference: `internal/jose/server/server_test.go`

- **Phase 6**: validateConfiguration() enhancement
  - Priority: Medium - Configuration robustness
  - Comprehensive field validation
  - Early error detection

---

## Detection Commands

```powershell
# Find test files with unsafe ServerSettings creation
grep -r "&cryptoutilConfig.ServerSettings{" **/*_test.go
grep -r "ServerSettings{" **/*_test.go

# Find files already using NewTestConfig (good examples)
grep -r "NewTestConfig" **/*_test.go

# Verify no blank bind addresses after migration
grep -r "BindPublicAddress.*\"\"" **/*_test.go
grep -r "BindPrivateAddress.*\"\"" **/*_test.go
```

---

## Prevention Guidelines

### Code Review Checklist

- [ ] All test files use `NewTestConfig()` for `ServerSettings` creation
- [ ] No direct instantiation of `&ServerSettings{}`
- [ ] No blank `BindPublicAddress` or `BindPrivateAddress`
- [ ] Bind addresses are `"127.0.0.1"` or `cryptoutilMagic.IPv4Loopback` in tests
- [ ] `0.0.0.0` used ONLY in Docker containers (never tests)

### Documentation Updates

- [x] `docs/cipher-im-migration/SERVICE-TEMPLATE-v2.md` - Added Phase 5 and Phase 6
- [x] `.github/instructions/06-02.anti-patterns.instructions.md` - Updated Windows Firewall section
- [x] `docs/cipher-im-migration/WINDOWS-FIREWALL-ROOT-CAUSE.md` - Created this document

### Related Instructions

- `.github/instructions/03-06.security.instructions.md` - Security patterns (Windows Firewall prevention)
- `.github/instructions/03-02.testing.instructions.md` - Testing patterns (dynamic port allocation)
- `.github/instructions/02-03.https-ports.instructions.md` - HTTPS ports and bind addresses

---

## Success Criteria

- [ ] All 20+ identified test files migrated to `NewTestConfig()`
- [ ] Zero Windows Firewall prompts during test execution
- [ ] `validateConfiguration()` enhanced with bind address validation
- [ ] All tests pass with localhost-only binding
- [ ] Code review checklist enforced for new test files
- [ ] Anti-patterns documentation updated
- [ ] Lessons learned documented for future reference

---

## Timeline

- **2026-01-02 14:00**: Root cause identified via code archaeology
- **2026-01-02 14:30**: Validated theory with grep searches and code reading
- **2026-01-02 15:00**: Documentation updated, migration tracking created
- **Next Steps**: Execute Phase 5 migration (20+ files), then Phase 6 validation enhancement

---

## Lessons Learned

1. **Empty strings in format strings can have dangerous defaults** (`":0"` → `0.0.0.0`)
2. **Partial struct initialization is unsafe** - Use helper functions for complex config
3. **Test helpers exist for good reasons** - `NewTestConfig()` was created to prevent exactly this issue
4. **Configuration validation is critical** - Catch errors early before server startup
5. **Code archaeology pays off** - Understanding WHY code exists prevents regressions

---

## References

- Internal Code: `internal/shared/config/config_test_helper.go` (NewTestConfig helper)
- Internal Code: `internal/template/server/listener/public.go` line 168 (critical bind path)
- Internal Code: `internal/shared/config/config.go` line 1379 (validateConfiguration)
- Tracking Doc: `docs/cipher-im-migration/SERVICE-TEMPLATE-v2.md` (Phase 5 and 6)
- Anti-Patterns: `.github/instructions/06-02.anti-patterns.instructions.md`
- Go Docs: `net.Listen()` behavior with empty bind addresses
