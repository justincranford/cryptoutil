# Testing Validation Plan

## Executive Summary

Validate refactor planning completeness by running full test suite, checking coverage, and verifying CI/CD workflow compatibility.

**Status**: Planning
**Dependencies**: Tasks 1-18 (all planning documents complete)
**Risk Level**: Low (validation only, no code changes)

## Pre-Refactor Baseline

### Test Suite Execution

**Command**:
```powershell
go test ./... -cover -timeout=10m
```

**Expected Outcome**:
- All tests pass (KMS server, identity services, common utilities)
- Coverage meets targets: ≥80% production, ≥85% cicd, ≥95% util
- No race conditions detected
- No test flakiness observed

### Coverage Targets by Package

| Package Type | Target | Current Status |
|--------------|--------|----------------|
| Production code | ≥80% | To be validated |
| Infrastructure (cicd) | ≥85% | To be validated |
| Utility code (common/util) | ≥95% | To be validated |

### CI/CD Workflow Validation

**Workflows to verify**:
1. **ci-quality.yml**: Build and lint checks
2. **ci-coverage.yml**: Test coverage collection
3. **ci-benchmark.yml**: Performance benchmarks
4. **ci-race.yml**: Race condition detection
5. **ci-fuzz.yml**: Fuzz testing
6. **ci-e2e.yml**: End-to-end integration tests
7. **ci-dast.yml**: Dynamic application security testing
8. **ci-load.yml**: Load and performance testing

**Expected Results**:
- All workflows pass on current main branch
- No workflow failures from recent planning commits
- Path filters correctly match existing file structure

## Test Execution Plan

### Phase 1: Unit Test Suite Validation

**Run full test suite**:

```powershell
# Run all tests with coverage
go test ./... -cover -timeout=10m -coverprofile=test-output/coverage_all.out

# Check for test failures
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Test failures detected - investigate before proceeding"
    exit 1
}
```

**Expected Coverage Output**:
```
ok      cryptoutil/internal/server              5.123s  coverage: 85.2% of statements
ok      cryptoutil/internal/identity/authz      3.456s  coverage: 82.1% of statements
ok      cryptoutil/internal/identity/idp        4.789s  coverage: 81.9% of statements
ok      cryptoutil/internal/common/util         2.345s  coverage: 96.3% of statements
ok      cryptoutil/internal/cmd/cicd            6.012s  coverage: 87.5% of statements
```

### Phase 2: Race Condition Detection

**Run tests with race detector**:

```powershell
# Run all tests with race detector
go test ./... -race -timeout=15m
```

**Expected Output**:
```
ok      cryptoutil/internal/server              8.234s
ok      cryptoutil/internal/identity/authz      5.678s
ok      cryptoutil/internal/identity/idp        7.123s
ok      cryptoutil/internal/common/util         3.456s
```

**Failure Criteria**: Any race condition warnings = FAIL

### Phase 3: Fuzz Testing Validation

**Run fuzz tests for critical packages**:

```powershell
# Fuzz test key generation (15 seconds per test)
go test -fuzz=FuzzGenerateRSAKeyPair -fuzztime=15s ./internal/common/crypto/keygen
go test -fuzz=FuzzGenerateECDSAKeyPair -fuzztime=15s ./internal/common/crypto/keygen
go test -fuzz=FuzzGenerateEd25519KeyPair -fuzztime=15s ./internal/common/crypto/keygen

# Fuzz test digests
go test -fuzz=FuzzDigest -fuzztime=15s ./internal/common/crypto/digests
```

**Expected Output**:
```
fuzz: elapsed: 15s, execs: 12345 (823/sec), new interesting: 0
```

**Failure Criteria**: Any crashes or panics during fuzzing = FAIL

### Phase 4: CI/CD Workflow Verification

**Run workflows locally with act**:

```powershell
# Quick validation workflow (most critical)
go run ./cmd/workflow -workflows=quality

# Coverage workflow
go run ./cmd/workflow -workflows=coverage

# E2E workflow (Docker Compose orchestration)
go run ./cmd/workflow -workflows=e2e
```

**Expected Output**:
```
✅ Workflow: quality - PASSED
✅ Workflow: coverage - PASSED
✅ Workflow: e2e - PASSED
```

**Failure Investigation**:
- Check workflow analysis markdown files in root directory
- Review step timing diagnostics for bottlenecks
- Verify path filters match expected file structure

### Phase 5: Coverage Regression Analysis

**Compare current coverage with baseline**:

```powershell
# Generate coverage report
go test ./... -coverprofile=test-output/coverage_all.out

# Check coverage by package
go tool cover -func=test-output/coverage_all.out | Select-String "total:"
```

**Expected Total Coverage**: ≥85% overall

**Per-Package Coverage Check**:

```powershell
# Extract package-level coverage
go tool cover -func=test-output/coverage_all.out > test-output/coverage_detail.txt

# Check critical packages
Select-String -Path test-output/coverage_detail.txt -Pattern "internal/server|internal/identity|internal/common/util|internal/cmd/cicd"
```

**Coverage Thresholds**:
- **KMS Server** (`internal/server`): ≥80%
- **Identity Services** (`internal/identity/*`): ≥80%
- **Common Utilities** (`internal/common/util`): ≥95%
- **CICD Utilities** (`internal/cmd/cicd`): ≥85%

## Test Results Documentation

### Test Report Template

Create `test-output/test-validation-report.md`:

```markdown
# Pre-Refactor Test Validation Report

**Date**: [YYYY-MM-DD]
**Git Commit**: [commit hash]
**Go Version**: 1.25.4

## Test Suite Results

### Unit Tests
- **Total Tests**: [number]
- **Passed**: [number]
- **Failed**: [number]
- **Skipped**: [number]
- **Duration**: [seconds]

### Coverage Summary
- **Overall Coverage**: [percentage]%
- **KMS Server**: [percentage]%
- **Identity Services**: [percentage]%
- **Common Utilities**: [percentage]%
- **CICD Utilities**: [percentage]%

### Race Detection
- **Tests Run**: [number]
- **Race Conditions Detected**: [number]
- **Status**: PASS/FAIL

### Fuzz Testing
- **Total Fuzz Tests**: [number]
- **Crashes**: [number]
- **Panics**: [number]
- **Status**: PASS/FAIL

## CI/CD Workflow Results

### ci-quality.yml
- **Status**: PASS/FAIL
- **Duration**: [minutes]
- **Notes**: [any issues]

### ci-coverage.yml
- **Status**: PASS/FAIL
- **Coverage**: [percentage]%
- **Notes**: [any issues]

### ci-e2e.yml
- **Status**: PASS/FAIL
- **Duration**: [minutes]
- **Notes**: [any issues]

## Coverage Regression Analysis

### Packages Below Threshold

| Package | Current | Target | Delta |
|---------|---------|--------|-------|
| [package] | [%] | [%] | [+/-] |

### Packages Above Threshold

| Package | Current | Target | Delta |
|---------|---------|--------|-------|
| [package] | [%] | [%] | [+/-] |

## Validation Checklist

- [ ] All unit tests pass
- [ ] No race conditions detected
- [ ] Fuzz tests complete without crashes
- [ ] Coverage meets thresholds (≥80% production, ≥85% cicd, ≥95% util)
- [ ] CI/CD workflows pass (quality, coverage, e2e)
- [ ] No test flakiness observed
- [ ] No new lint errors introduced

## Issues Discovered

### Critical Issues
- [List any critical issues found]

### Non-Critical Issues
- [List any non-critical issues found]

## Recommendations

- [List any recommendations for improving test coverage or fixing issues]

## Sign-Off

**Validated By**: [Name]
**Date**: [YYYY-MM-DD]
**Status**: APPROVED/REJECTED
```

## Success Criteria

### Must-Have (Blocking)
- [ ] All unit tests pass (`go test ./...`)
- [ ] No race conditions detected (`go test ./... -race`)
- [ ] Fuzz tests complete without crashes
- [ ] Overall coverage ≥85%
- [ ] CI/CD quality workflow passes

### Should-Have (Warning)
- [ ] KMS server coverage ≥80%
- [ ] Identity services coverage ≥80%
- [ ] CICD utilities coverage ≥85%
- [ ] Common utilities coverage ≥95%
- [ ] CI/CD coverage workflow passes
- [ ] CI/CD e2e workflow passes

### Nice-to-Have (Informational)
- [ ] No test flakiness observed over 3 runs
- [ ] Benchmark results stable (no performance regression)
- [ ] Load tests pass (Gatling scenarios)
- [ ] DAST scans pass (Nuclei + ZAP)

## Known Test Issues (Pre-Refactor)

### SQLite Transaction Tests

**Issue**: Some SQLite transaction tests may be slow (10+ seconds) due to:
- PRAGMA settings (WAL mode, busy timeout)
- Connection pool configuration (MaxOpenConns=5 for GORM)
- Parallel test execution with `t.Parallel()`

**Mitigation**: Use `-timeout=10m` flag to allow sufficient time

### PostgreSQL Integration Tests

**Issue**: PostgreSQL tests require Docker Compose services running

**Mitigation**:
```powershell
# Start PostgreSQL before running integration tests
docker compose -f deployments/compose/compose.yml up -d postgres

# Wait for PostgreSQL readiness
Start-Sleep -Seconds 5

# Run tests
go test ./internal/identity/repository/orm -v
```

### E2E Test Timing

**Issue**: E2E tests with Docker Compose orchestration can take 5-10 minutes

**Mitigation**: Run E2E tests separately from unit tests
```powershell
# Unit tests only (fast)
go test ./internal/... -short

# E2E tests only (slow)
go test ./internal/test/e2e/... -v
```

## Post-Validation Actions

### If Tests Pass
1. Document baseline coverage in test report
2. Commit test report to `test-output/` directory
3. Proceed with refactor implementation (follow plans in `docs/01-refactor/`)

### If Tests Fail
1. Investigate test failures and fix issues
2. Re-run test suite to confirm fixes
3. Document any deviations from baseline
4. DO NOT proceed with refactor until tests pass

## Timeline

- **Phase 1**: Unit test suite validation (30 minutes)
- **Phase 2**: Race condition detection (45 minutes)
- **Phase 3**: Fuzz testing validation (30 minutes)
- **Phase 4**: CI/CD workflow verification (2 hours)
- **Phase 5**: Coverage regression analysis (30 minutes)
- **Documentation**: Test report creation (30 minutes)

**Total**: 5 hours (1 day)

## Cross-References

- [Testing Instructions](.github/instructions/01-02.testing.instructions.md) - Testing best practices
- [CI/CD Instructions](.github/instructions/02-01.github.instructions.md) - Workflow execution
- [Pre-commit Hooks](docs/pre-commit-hooks.md) - Local testing automation

## Next Steps

After testing validation:
1. **Document baseline**: Create test validation report
2. **Task 20**: Documentation finalization (handoff package)
3. **Begin implementation**: Follow refactor plans in sequence
