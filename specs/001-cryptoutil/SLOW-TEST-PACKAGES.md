# Slow Test Packages - Current Status

**Purpose**: Track slow-running test packages for optimization
**Last Updated**: 2025-01-10
**Test Command**: `go test ./... -cover -json` (CGO_ENABLED=0)

---

## Critical Packages (>50s execution)

**Immediate optimization targets**:

| Package | Max Elapsed | Coverage | Priority | Strategy |
|---------|-------------|----------|----------|----------|
| `internal/common/crypto/keygen` | **202.53s** | 85.2% | CRITICAL | Crypto key generation - inherently slow, already parallelized |
| `internal/jose` | **84.17s** | 48.8% | CRITICAL | Increase coverage 48.8% → 95% FIRST, then optimize |
| `internal/jose/server` | **76.33s** | 56.1% | CRITICAL | Parallel subtests, reduce setup/teardown |
| `internal/kms/client` | **64.83s** | 76.2% | CRITICAL | Mock KMS server, parallel execution |

**Total Critical Time**: ~428s (~7.1 minutes)

**Analysis**:

- `keygen` slowness is expected (RSA key generation, PBKDF2, crypto operations)
- `jose` and `jose/server` need coverage improvements BEFORE optimization
- `kms/client` should mock server dependencies for faster tests

---

## High Priority (30-50s execution)

| Package | Max Elapsed | Coverage | Strategy |
|---------|-------------|----------|----------|
| `internal/kms/server/application` | **32.93s** | 64.7% | Parallel server tests, dynamic port allocation |
| `internal/kms/server/barrier` | **32.32s** | 75.5% | Parallel crypto operations |
| `internal/identity/test/load` | **30.62s** | N/A | Load tests - acceptable duration |

---

## Medium Priority (20-30s execution)

| Package | Max Elapsed | Coverage | Notes |
|---------|-------------|----------|-------|
| `internal/identity/authz/clientauth` | **23.61s** | 78.4% | Already parallelized, acceptable for OAuth flows |
| `internal/identity/idp` | **16.97s** | 54.9% | Coverage improvement needed |
| `internal/identity/authz` | **13.23s** | 77.2% | Acceptable, complex auth logic |

---

## Acceptable Duration (<20s)

Most packages under 20s are within acceptable performance bounds for their functionality:

- **Crypto operations**: `keygen`, `barrier`, crypto primitives (inherently slow)
- **Integration tests**: `test/integration`, `test/load` (expected overhead)
- **Repository tests**: `sqlrepository`, `orm` (database operations)
- **Server tests**: Application startup, health checks (acceptable)

---

## Constitution v2.0.0 Compliance

### Evidence-Based Completion Requirements

**Current Status**: ✅ DOCUMENTED

- [x] Slow packages identified with current timings
- [x] Optimization strategies defined
- [x] Coverage gaps highlighted (`jose` 48.8%, `jose/server` 56.1%)
- [x] Inherent slowness acknowledged (`keygen` crypto operations)

### Recommendations

1. **Coverage First**: Fix `jose` (48.8% → 95%) and `jose/server` (56.1% → 95%)
   - Better coverage = more confidence in optimization changes
   - Avoid premature optimization without tests

2. **Accept Crypto Slowness**: `keygen` at 202s is acceptable
   - RSA 2048/4096 key generation is slow by design
   - PBKDF2 iterations (600k) are intentionally expensive
   - Already using parallel subtests

3. **Mock External Dependencies**: `kms/client` should use mock server
   - Current: Real KMS server calls over network
   - Target: Mock responses, <20s execution

4. **Parallel Server Tests**: `kms/server/application`, `jose/server`
   - Use dynamic port allocation (port 0 pattern)
   - Parallel subtests for independent operations

---

## Historical Context

**Previous Iteration** (specs/001-cryptoutil/SLOW-TEST-PACKAGES.md):

- `clientauth`: 168s → 23.61s ✅ **Optimized 87% reduction**
- `jose/server`: 94s → 76.33s ⚠️ **Improved 19% but still high**
- `kms/client`: 74s → 64.83s ⚠️ **Improved 12% but still high**

**Key Wins**:

- `clientauth` optimization from 168s to 23s through aggressive `t.Parallel()`
- Overall test suite <200s for all packages

**Remaining Work**:

- `jose` and `jose/server` coverage improvements
- `kms/client` mocking strategy

---

**Status**: Active tracking document
**Next Review**: After coverage improvements in `jose` packages
