# Slow Test Packages - Iteration 3 Tracking

**Purpose**: Document slow-running test packages for optimization (extracted from DELETE-ME-LATER-SLOW-TEST-PACKAGES.md)

**Source**: `go test ./... -cover` output from early iteration

---

## Packages Requiring Optimization (≥20s execution)

**Sorted by execution time (descending)**:

| Package | Execution Time | Coverage | Priority | Optimization Strategy |
|---------|----------------|----------|----------|----------------------|
| `internal/identity/authz/clientauth` | **168.383s** | 78.4% | CRITICAL | Table-driven parallelism, selective test execution pattern |
| `internal/jose/server` | **94.342s** | 56.1% | CRITICAL | Parallel subtests, reduce setup/teardown overhead |
| `internal/kms/client` | **73.859s** | 76.2% | CRITICAL | Mock heavy dependencies, parallel execution |
| `internal/jose` | **67.003s** | 48.8% | HIGH | Improve coverage first (48.8% → 95%), then optimize |
| `internal/kms/server/application` | **27.596s** | 64.7% | HIGH | Parallel server tests, dynamic port allocation |

**Total packages ≥20s**: 5 packages
**Combined execution time**: 430.9 seconds (~7.2 minutes)

## Packages With Moderate Performance Impact (10-20s execution)

| Package | Execution Time | Coverage | Priority | Optimization Strategy |
|---------|----------------|----------|----------|----------------------|
| `internal/identity/authz` | **19.248s** | 77.2% | MEDIUM | Already parallelized, needs test data isolation review |
| `internal/identity/test/unit` | **17.896s** | [no statements] | LOW | Infrastructure tests, acceptable duration |
| `internal/identity/test/integration` | **16.370s** | [no statements] | LOW | Integration tests, acceptable duration |
| `internal/identity/idp` | **15.381s** | 54.9% | MEDIUM | Improve coverage (54.9% → 95%), reduce DB setup time |
| `internal/infra/realm` | **13.787s** | 85.6% | LOW | Good coverage, acceptable duration |
| `internal/kms/server/barrier` | **12.559s** | 75.5% | LOW | Parallel crypto operations tests |

## Acceptable Duration Packages (5-10s)

| Package | Execution Time | Coverage | Notes |
|---------|----------------|----------|-------|
| `internal/identity/rotation` | 7.674s | 83.7% | Crypto operations, acceptable |
| `internal/identity/jobs` | 7.448s | 89.0% | Background job tests, acceptable |
| `internal/common/crypto/keygen` | 6.394s | 85.2% | Crypto key generation, acceptable |

## Optimization Targets (Packages ≥20s)

### Critical Priority (>60s execution)

1. **clientauth (168s → target <30s)**
   - Strategy: Use `t.Parallel()` aggressively
   - Split into multiple test files by auth method
   - Use selective execution pattern for local dev

2. **jose/server (94s → target <20s)**
   - Strategy: Parallel subtests
   - Reduce Fiber app setup/teardown overhead
   - Use shared test server instance

3. **kms/client (74s → target <20s)**
   - Strategy: Mock KMS server dependency
   - Parallel test execution
   - Reduce network roundtrip simulation

### High Priority (30-70s execution)

1. **jose (67s → target <15s)**
   - Strategy: Increase coverage 48.8% → 95% FIRST
   - Then apply parallel execution
   - Reduce cryptographic operation redundancy

### Medium Priority (20-30s execution)

1. **kms/server/application (28s → target <10s)**
   - Strategy: Parallel server tests
   - Dynamic port allocation pattern
   - Reduce test server setup/teardown overhead

## Constitution v2.0.0 Impact

**Evidence-Based Completion**: Cannot mark Phase 4 complete without:

- [ ] Documentation of slow packages in iteration 3 tracking
- [ ] Optimization strategies identified for critical packages (>60s)
- [ ] Target execution times defined
- [ ] At least 1 critical package optimized as proof of concept

**Next Steps**:

1. Add benchmark tests to slow packages (mandatory per constitution)
2. Identify bottlenecks using `go test -bench=.` and `-cpuprofile`
3. Apply parallelization where thread-safe
4. Document optimizations in PROGRESS.md

---

**Extracted**: 2025-12-05 (from DELETE-ME-LATER-SLOW-TEST-PACKAGES.md)
**Status**: Active iteration 3 tracking document
**Ref**: specs/003-cryptoutil/tasks.md ITER3-015
