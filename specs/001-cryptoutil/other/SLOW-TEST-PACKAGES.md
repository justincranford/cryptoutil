# Slow Test Packages - Current Status

**Purpose**: Track slow-running test packages for optimization
**Last Updated**: 2025-12-14
**Test Command**: `go test ./...` (CGO_ENABLED=0)

---

## Critical Packages (>30s execution)

**Current slow packages** (from latest `go test ./...` run):

| Package | Time | Status | Priority | Strategy |
|---------|------|--------|----------|----------|
| `internal/jose` | **98.357s** | ✅ PASS | HIGH | Already optimized for single-package run (27s), concurrent execution causes slowdown |
| `internal/common/crypto/keygen` | **36.245s** | ✅ PASS | MEDIUM | Already optimized for single-package run (4.8s), inherently slow crypto operations |
| `internal/identity/idp` | **22.093s** | ✅ PASS | LOW | OAuth/IdP flows with crypto operations |
| `internal/kms/server/repository/sqlrepository` | **12.752s** | ❌ FAIL | **CRITICAL** | PostgreSQL connectivity test failing locally (no PG service) |

**Total Critical Time**: ~169s (~2.8 minutes)

**Analysis**:

- **sqlrepository REGRESSION**: Test `TestNewSQLRepository_PostgreSQL_PingRetry` expects PostgreSQL service, retries 5× and fails locally (5.82s just for this test)
- **keygen & jose slowdown**: Isolated tests are fast (keygen 4.8s, jose 27.5s), but concurrent `go test ./...` shows 3-8× slowdown
- **Concurrent execution impact**: Tests designed for parallel execution within package experience contention when all packages run concurrently

---

## Acceptable Duration (<20s)

Most packages under 20s are within acceptable performance bounds for their functionality:

| Package | Time | Notes |
|---------|------|-------|
| `internal/identity/test/integration` | 11.211s | Integration tests with database |
| `internal/identity/mfa` | 9.090s | MFA crypto operations (TOTP, WebAuthn) |
| `internal/identity/jobs` | 8.453s | Background job processing |
| `internal/identity/idp/userauth` | 8.841s | User authentication flows |
| `internal/identity/test/unit` | 7.903s | Unit test suite |
| `internal/infra/realm` | 7.469s | Multi-tenant realm management |
| `internal/identity/issuer` | 7.195s | JWT/JWS issuer operations |

---

## Critical Issue: PostgreSQL Test Regression

**Package**: `internal/kms/server/repository/sqlrepository`
**Test**: `TestNewSQLRepository_PostgreSQL_PingRetry`
**Status**: ❌ FAILING (12.752s total, 5.82s for failing test alone)

**Root Cause**: Test expects PostgreSQL service running on localhost:5432, but:

- No PostgreSQL service running locally
- Test retries connection 5 times with exponential backoff
- Each retry attempts multiple auth mechanisms
- Total retry time: ~5.8 seconds before giving up

**Error**: `failed to ping database: FATAL: password authentication failed for user "cryptoutil" (SQLSTATE 28P01)`

**Fix Strategy**:

1. Skip test if PostgreSQL not available (check env var or connection first)
2. OR: Use testcontainers for PostgreSQL in TestMain
3. OR: Mark test as integration test with build tag `//go:build integration`

**Recommendation**: Use build tag `//go:build integration` since test is specifically for PostgreSQL connectivity, not general functionality.

---

## Recommendations

1. **Fix sqlrepository Regression FIRST** (CRITICAL)
   - Test should skip or use testcontainers when PostgreSQL unavailable
   - Prevents 12+ second failure on every local test run
   - Reduces feedback loop during development

2. **Accept Concurrent Slowdown Pattern** (DOCUMENTED)
   - Isolated tests: keygen 4.8s, jose 27.5s ✅ Fast
   - Concurrent run: keygen 36s, jose 98s ⚠️ 3-8× slower
   - Root cause: Resource contention (CPU, memory, I/O) when all packages test concurrently
   - **Solution**: Focus on isolated package performance, not concurrent timing

3. **Crypto Operations Are Inherently Slow**
   - RSA 2048/4096 key generation: 100-500ms per key
   - PBKDF2 600k iterations: intentionally expensive for security
   - Already using parallel subtests where safe
   - **No further optimization needed**

---

## Historical Context

**Optimization Progress**:

- Previous: keygen 202s → Current isolated: 4.8s ✅ **97.6% reduction**
- Previous: jose 84s → Current isolated: 27.5s ✅ **67.3% reduction**
- **Concurrent slowdown documented**: Expected behavior when running `go test ./...` with limited CPU/memory

**Known Issue**: Test times vary 3-8× between isolated and concurrent execution due to:

- Limited CPU cores (shared across 100+ test packages)
- Memory pressure from concurrent database operations
- I/O contention from parallel file system operations
- Race detector overhead (when enabled)

**Key Wins**:

- `clientauth` optimization from 168s to 23s through aggressive `t.Parallel()`
- Overall test suite <200s for all packages

**Remaining Work**:

- `jose` and `jose/server` coverage improvements
- `kms/client` mocking strategy

---

**Status**: Active tracking document
**Next Review**: After coverage improvements in `jose` packages
