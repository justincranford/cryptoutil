# Test Performance Analysis

## Executive Summary

**Local vs GitHub Execution Discrepancy**: Tests run 150x slower in GitHub workflows compared to local execution, with the disparity NOT caused by test code itself.

| Environment | sqlrepository | clientauth | jose/server | kms/client | jose |
|-------------|---------------|------------|-------------|------------|------|
| **Local** | <2s | Unknown | Unknown | Unknown | Unknown |
| **GitHub** | 303s | 168s | 100s | 74s | 67s |
| **Ratio** | 150x | TBD | TBD | TBD | TBD |

**Key Finding**: Individual tests execute in 0.00s (instant) - slowness is workflow/infrastructure overhead, NOT test code.

---

## sqlrepository Package Analysis

**Package**: `internal/kms/server/repository/sqlrepository`
**GitHub Execution**: 303s (after 50% optimization from 601s)
**Local Execution**: <2s
**Discrepancy**: 150x slower on GitHub

### Test Execution Pattern (from WORKFLOW-sqlrepository-TEST-TIMES.md)

```
=== RUN TestMain
--- PASS: TestMain (0.00s)
=== RUN TestNewSQLRepository_NilContext
=== PAUSE TestNewSQLRepository_NilContext        <- Parallel tests PAUSE
=== RUN TestNewSQLRepository_NilTelemetry
=== PAUSE TestNewSQLRepository_NilTelemetry      <- Parallel tests PAUSE
...
=== RUN TestSQLTransaction_PanicRecovery
=== PAUSE TestSQLTransaction_PanicRecovery       <- 28 parallel tests PAUSEd
--- PASS: TestSQLRepository_Shutdown (0.00s)     <- Sequential test runs instantly
=== CONT TestNewSQLRepository_NilContext         <- Parallel tests CONTINUE
=== CONT TestNewSQLRepository_NilTelemetry
...
--- PASS: TestNewSQLRepository_NilContext (0.00s) <- Tests execute instantly!
--- PASS: TestNewSQLRepository_NilTelemetry (0.00s)
```

**Critical Insights**:

1. **Test code is FAST**: All tests execute in 0.00s
2. **Parallel pattern correct**: Tests PAUSE → CONT → PASS pattern is expected
3. **Slowness is external**: 303s total execution NOT from test logic

**Optimization Status**:

- ✅ Added `t.Parallel()` to 28 tests (commits 340f7298, 1820dca0)
- ✅ Reduced execution time 601s → 303s (50% improvement)
- ❌ Still 150x slower than local (2s vs 303s)
- ⚠️ Further optimization requires identifying workflow overhead

---

## Pending Analysis (High Priority)

### clientauth Package (168s)

**Package**: `internal/identity/clientauth`
**Priority**: CRITICAL (slowest remaining package)
**Action Required**: Create `WORKFLOW-clientauth-TEST-TIMES.md`

**Analysis Plan**:

```powershell
go test ./internal/identity/clientauth -v -count=1 -timeout=10m | Out-File docs/WORKFLOW-clientauth-TEST-TIMES.md
```

**Expected Findings**:

- Identify if tests use `t.Parallel()` (likely missing)
- Measure individual test execution times
- Compare local vs GitHub execution ratio

### jose/server Package (100s)

**Package**: `internal/jose/server`
**Priority**: HIGH
**Current Coverage**: Unknown (needs measurement)
**Action Required**: Create `WORKFLOW-jose-server-TEST-TIMES.md`

**Analysis Plan**:

```powershell
go test ./internal/jose/server -v -count=1 -timeout=10m | Out-File docs/WORKFLOW-jose-server-TEST-TIMES.md
```

### kms/client Package (74s)

**Package**: `internal/kms/client`
**Priority**: HIGH
**Action Required**: Create `WORKFLOW-kms-client-TEST-TIMES.md`

**Analysis Plan**:

```powershell
go test ./internal/kms/client -v -count=1 -timeout=10m | Out-File docs/WORKFLOW-kms-client-TEST-TIMES.md
```

### jose Package (67s)

**Package**: `internal/jose`
**Priority**: MEDIUM
**Current Coverage**: 50.5% (needs improvement to 85%+ BEFORE optimization)
**Action Required**: Improve coverage first, then create timing report

**Analysis Plan**:

1. Increase coverage 50.5% → 85%+ (primary goal)
2. Then create `WORKFLOW-jose-TEST-TIMES.md`
3. Apply `t.Parallel()` if missing

---

## Root Cause Hypotheses

### Hypothesis 1: GitHub Runner Performance

**Theory**: GitHub-hosted runners are significantly slower than local development machines.

**Evidence**:

- Tests execute in 0.00s individually
- 150x slowdown is extreme even for shared infrastructure
- No test code changes explain the discrepancy

**Validation**: Compare GitHub runner specs vs local machine specs.

### Hypothesis 2: Container Startup Overhead

**Theory**: Docker container startup and PostgreSQL initialization add significant overhead.

**Evidence**:

- Tests use TestMain to start PostgreSQL containers once per package
- Container startup should be amortized across all tests
- sqlrepository uses in-memory SQLite (NO containers) yet still 150x slower

**Status**: ❌ DISPROVEN - sqlrepository uses SQLite in-memory, no containers

### Hypothesis 3: Workflow Job Overhead

**Theory**: GitHub Actions job setup, checkout, Go setup, and teardown consume most execution time.

**Evidence**:

- Workflow includes: checkout, setup-go, docker-compose, health checks, test execution, artifact upload
- Each step adds latency
- Multiple services (postgres, otel-collector, grafana) start before tests

**Validation Plan**:

1. Add timing instrumentation to ci-coverage workflow
2. Measure: job setup, docker-compose up, health checks, actual test execution
3. Identify largest time consumer

### Hypothesis 4: Network/I-O Bottlenecks

**Theory**: GitHub runner disk I/O or network latency slows test execution.

**Evidence**:

- Tests write to disk (test-output/ directory)
- Coverage file generation adds I/O overhead
- Database operations (even SQLite in-memory) may be slower on shared infrastructure

**Validation**: Profile I/O operations during test execution.

---

## Optimization Strategy (Evidence-Based)

### Phase 1: Measure Everything

**Objective**: Collect comprehensive timing data before making assumptions.

**Tasks**:

1. ✅ Create WORKFLOW-sqlrepository-TEST-TIMES.md (complete)
2. ⏳ Create WORKFLOW-clientauth-TEST-TIMES.md (168s package)
3. ⏳ Create WORKFLOW-jose-server-TEST-TIMES.md (100s package)
4. ⏳ Create WORKFLOW-kms-client-TEST-TIMES.md (74s package)
5. ⏳ Create WORKFLOW-jose-TEST-TIMES.md (67s package, after coverage improvement)
6. ⏳ Instrument ci-coverage workflow with timing checkpoints
7. ⏳ Compare local vs GitHub execution for ALL packages

### Phase 2: Apply t.Parallel() Where Missing

**Objective**: Ensure all test packages use concurrent execution.

**Priority Order** (by execution time):

1. clientauth (168s) - likely missing `t.Parallel()`
2. jose/server (100s) - verify parallel status
3. kms/client (74s) - verify parallel status
4. jose (67s) - add after coverage improvement

**Expected Impact**: 30-50% reduction per package (based on sqlrepository results)

### Phase 3: Optimize Workflow Overhead

**Objective**: Reduce non-test execution time in ci-coverage workflow.

**Potential Optimizations**:

- Reduce health check intervals (if safe)
- Parallelize service startup (docker-compose --detach)
- Cache Docker images (pre-pull optimization)
- Reduce artifact upload size (compress coverage files)

### Phase 4: Consider Self-Hosted Runners

**Objective**: Evaluate if self-hosted runners eliminate 150x slowdown.

**Decision Criteria**:

- If workflow overhead < 50s but test execution still 300s → Consider self-hosted
- If workflow overhead > 250s → Optimize workflow first
- Cost-benefit analysis required

---

## Test Execution Time Targets

**MANDATORY**: All test packages MUST execute in <30s each on GitHub workflows.

| Package | Current (GitHub) | Target | Status |
|---------|------------------|--------|--------|
| sqlrepository | 303s | <30s | ❌ 10x over target |
| clientauth | 168s | <30s | ❌ 5.6x over target |
| jose/server | 100s | <30s | ❌ 3.3x over target |
| kms/client | 74s | <30s | ❌ 2.5x over target |
| jose | 67s | <30s | ❌ 2.2x over target |

**Total Current**: 712s (11.9 minutes)
**Total Target**: <150s (2.5 minutes)
**Required Speedup**: 4.7x overall

---

## Action Items (Immediate)

**Priority 1 (CRITICAL - Do First)**:

1. Create WORKFLOW-clientauth-TEST-TIMES.md
2. Analyze clientauth timing report
3. Add `t.Parallel()` to clientauth tests if missing
4. Rerun and measure improvement

**Priority 2 (HIGH)**:

1. Create WORKFLOW-jose-server-TEST-TIMES.md
2. Create WORKFLOW-kms-client-TEST-TIMES.md
3. Add `t.Parallel()` where missing
4. Measure improvements

**Priority 3 (MEDIUM)**:

1. Improve jose package coverage 50.5% → 85%
2. Create WORKFLOW-jose-TEST-TIMES.md
3. Add `t.Parallel()` if missing

**Priority 4 (Investigation)**:

1. Instrument ci-coverage workflow with timing checkpoints
2. Identify largest time consumer (setup, docker, tests, teardown)
3. Create optimization plan based on data

---

## References

- Constitution: `.specify/memory/constitution.md` - Testing mandates (NEVER -p=1, ALWAYS concurrent)
- Test Instructions: `.github/instructions/01-02.testing.instructions.md`
- Spec Tasks: `specs/001-cryptoutil/tasks.md` - Phase 0 test optimization tasks
- Timing Report: `docs/WORKFLOW-sqlrepository-TEST-TIMES.md`

---

*Document Version: 1.0.0*
*Last Updated: 2025-01-05*
*Next Review: After completing Priority 1 tasks*
