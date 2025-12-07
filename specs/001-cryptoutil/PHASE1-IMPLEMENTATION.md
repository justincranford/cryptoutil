# Phase 1: Fix CI/CD Workflows Implementation Guide

**Duration**: Days 3-4 (6-8 hours)
**Prerequisites**: Phase 0 complete (all test packages <10s)
**Status**: ❌ Not Started

## Overview

Phase 1 fixes 8 failing GitHub Actions workflows in priority order. Each workflow follows the same pattern:

1. Run locally with Act: `go run ./cmd/workflow -workflows=<name>`
2. Identify failure root cause
3. Implement fix
4. Verify fix locally with Act
5. Commit and push
6. Verify fix in GitHub Actions

## Priority Order (MANDATORY)

Workflows MUST be fixed in this specific order per 02-01.github.instructions.md:

1. **ci-coverage** (P1.1) - CRITICAL - Foundation for code quality metrics
2. **ci-benchmark** (P1.4) - HIGH - Performance baseline required early
3. **ci-fuzz** (P1.6) - HIGH - Security testing foundation
4. **ci-e2e** (P1.3) - HIGH - Integration validation
5. **ci-dast** (P1.7) - MEDIUM - Security scanning
6. **ci-race** (P1.5) - MEDIUM - Concurrency validation
7. **ci-load** (P1.8) - MEDIUM - Performance validation
8. **ci-sast** (P1.2) - LOW - Static analysis (slowest, runs last)

## Task Details

---

### P1.1: Fix ci-coverage Workflow ⭐ CRITICAL

**Priority**: 1-CRITICAL (MUST BE FIXED FIRST)
**Effort**: 1 hour
**Root Cause**: Coverage aggregation and reporting failures

**Current Failure Symptoms**:

- Coverage reports not generated correctly
- Coverage merge across packages failing
- Artifacts upload issues

**Implementation Strategy**:

```bash
# Step 1: Run locally to reproduce failure
go run ./cmd/workflow -workflows=coverage

# Step 2: Identify coverage collection pattern
# Expected pattern per 01-02.testing.instructions.md:
# Place coverage files in ./test-output/:
# go test ./pkg -coverprofile=test-output/coverage_pkg.out
```

**Files to Investigate**:

- `.github/workflows/ci-coverage.yml`
- Coverage collection commands
- Artifact upload configuration

**Acceptance Criteria**:

- ✅ Local Act run succeeds with coverage reports
- ✅ GitHub Actions run succeeds
- ✅ Coverage artifacts uploaded to workflow artifacts
- ✅ Coverage reports show accurate percentages
- ✅ All packages show individual coverage metrics

**Validation Commands**:

```bash
# Local validation
go run ./cmd/workflow -workflows=coverage

# Post-fix verification (after push)
gh run list --workflow=ci-coverage.yml --limit 1
gh run view <run-id>
```

---

### P1.4: Fix ci-benchmark Workflow ⭐ HIGH

**Priority**: 2-HIGH
**Effort**: 1 hour
**Root Cause**: Benchmark baseline generation and comparison issues

**Current Failure Symptoms**:

- Benchmark baselines not created/stored
- Comparison against previous runs failing
- Performance regression detection not working

**Implementation Strategy**:

```bash
# Step 1: Run locally
go run ./cmd/workflow -workflows=benchmark

# Step 2: Check benchmark output format
# Per 01-02.testing.instructions.md:
# Run: go test -bench=. -benchmem ./pkg/crypto
```

**Files to Investigate**:

- `.github/workflows/ci-benchmark.yml`
- Benchmark result storage mechanism
- Baseline comparison logic

**Acceptance Criteria**:

- ✅ Local Act run succeeds with benchmark results
- ✅ GitHub Actions run succeeds
- ✅ Benchmark baselines stored as artifacts
- ✅ Performance regression detection works
- ✅ Benchmark comparison output readable

**Validation Commands**:

```bash
# Local validation
go run ./cmd/workflow -workflows=benchmark

# Verify benchmarks run successfully
go test -bench=. -benchmem ./pkg/crypto -run=^$
```

---

### P1.6: Fix ci-fuzz Workflow ⭐ HIGH

**Priority**: 3-HIGH
**Effort**: 1 hour
**Root Cause**: Fuzz test execution and timeout configuration

**Current Failure Symptoms**:

- Fuzz tests timing out
- Fuzz corpus not preserved
- Multiple fuzz tests interfering with each other

**Implementation Strategy**:

```bash
# Step 1: Run locally
go run ./cmd/workflow -workflows=fuzz

# Step 2: Check fuzz test execution
# Per 01-02.testing.instructions.md:
# ALWAYS run from project root: go test -fuzz=FuzzXXX -fuzztime=15s ./path
# Use unquoted names, PowerShell ; for chaining
# Minimum fuzz time: 15 seconds per test
```

**Files to Investigate**:

- `.github/workflows/ci-fuzz.yml`
- Fuzz test discovery mechanism
- Fuzz corpus storage/retrieval

**Critical Requirements**:

- Fuzz test names MUST be unique, NOT substrings of others
- Each fuzz test runs for minimum 15 seconds
- Tests run sequentially (no parallel fuzz execution)

**Acceptance Criteria**:

- ✅ Local Act run succeeds with fuzz results
- ✅ GitHub Actions run succeeds
- ✅ All fuzz tests execute for ≥15 seconds
- ✅ Fuzz corpus preserved as artifacts
- ✅ No test name substring conflicts

**Validation Commands**:

```bash
# Local validation
go run ./cmd/workflow -workflows=fuzz

# Test individual fuzz targets
go test -fuzz=FuzzHKDFAllVariants -fuzztime=15s ./internal/crypto/keygen
```

---

### P1.3: Fix ci-e2e Workflow ⭐ HIGH

**Priority**: 4-HIGH
**Effort**: 1 hour
**Root Cause**: Docker Compose service startup and health checks

**Current Failure Symptoms**:

- Services not starting correctly
- Health checks timing out
- Service dependencies not respected

**Implementation Strategy**:

```bash
# Step 1: Run locally
go run ./cmd/workflow -workflows=e2e

# Step 2: Check Docker Compose configuration
# Per 02-01.github.instructions.md:
# ci-e2e uses Full Docker stack with health checks
# Per 02-02.docker.instructions.md:
# ALWAYS use 127.0.0.1 in containers (not localhost)
# Use wget for health checks (available in Alpine)
```

**Files to Investigate**:

- `.github/workflows/ci-e2e.yml`
- `deployments/compose/compose.yml`
- Health check configurations
- Service dependencies

**Critical Requirements**:

- All services must have proper health checks
- Services must wait for dependencies (depends_on with condition)
- Use `127.0.0.1` for loopback addresses (not `localhost`)
- Health checks use `wget` (not `curl` - Alpine compatibility)

**Acceptance Criteria**:

- ✅ Local Act run succeeds with all services healthy
- ✅ GitHub Actions run succeeds
- ✅ All services start in correct order
- ✅ Health checks pass within timeout
- ✅ E2E tests execute successfully

**Validation Commands**:

```bash
# Local validation
go run ./cmd/workflow -workflows=e2e

# Manual Docker Compose verification
docker compose -f ./deployments/compose/compose.yml up -d
docker compose -f ./deployments/compose/compose.yml ps
docker compose -f ./deployments/compose/compose.yml down -v
```

---

### P1.7: Fix ci-dast Workflow

**Priority**: 5-MEDIUM
**Effort**: 1 hour
**Root Cause**: Service connectivity for security scanning

**Current Failure Symptoms**:

- Nuclei scanner cannot reach services
- Service endpoints incorrect
- TLS certificate verification issues

**Implementation Strategy**:

```bash
# Step 1: Run locally
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"

# Step 2: Verify service endpoints
# Per 03-04.dast.instructions.md:
# cryptoutil-sqlite: https://localhost:8080/
# cryptoutil-postgres-1: https://localhost:8081/
# cryptoutil-postgres-2: https://localhost:8082/
```

**Files to Investigate**:

- `.github/workflows/ci-dast.yml`
- Service connectivity verification commands
- Nuclei scan configuration

**Critical Requirements**:

- Services MUST be started before nuclei scans
- Use `curl -k` for TLS certificate verification bypass in CI
- Verify `/ui/swagger/doc.json` endpoint before scanning

**Acceptance Criteria**:

- ✅ Local Act run succeeds with quick scan
- ✅ GitHub Actions run succeeds
- ✅ All service endpoints reachable
- ✅ Nuclei scans complete without connectivity errors
- ✅ DAST reports uploaded as artifacts

**Validation Commands**:

```bash
# Local validation (quick scan, 3-5 min)
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"

# Manual service verification (Windows PowerShell)
Invoke-WebRequest -Uri https://localhost:8080/ui/swagger/doc.json -SkipCertificateCheck
```

---

### P1.5: Fix ci-race Workflow

**Priority**: 6-MEDIUM
**Effort**: 1 hour
**Root Cause**: Race detector configuration and timeout issues

**Current Failure Symptoms**:

- Race detector tests timing out
- Too many goroutines for race detector
- Race conditions not being reported correctly

**Implementation Strategy**:

```bash
# Step 1: Run locally
go run ./cmd/workflow -workflows=race

# Step 2: Check race detector configuration
# Race detector uses -race flag
# Slower than normal tests (2-20x overhead)
```

**Files to Investigate**:

- `.github/workflows/ci-race.yml`
- Race detector timeout configuration
- Test parallelism settings with race detector

**Acceptance Criteria**:

- ✅ Local Act run succeeds with race detection
- ✅ GitHub Actions run succeeds
- ✅ All tests pass race detection
- ✅ No false positives reported
- ✅ Reasonable execution time (<15 minutes)

**Validation Commands**:

```bash
# Local validation
go run ./cmd/workflow -workflows=race

# Manual race detection
go test -race ./internal/...
```

---

### P1.8: Fix ci-load Workflow

**Priority**: 7-MEDIUM
**Effort**: 30 minutes
**Root Cause**: Gatling load test configuration

**Current Failure Symptoms**:

- Gatling tests not finding target services
- Load test configuration incorrect
- Results reporting failures

**Implementation Strategy**:

```bash
# Step 1: Run locally
go run ./cmd/workflow -workflows=load

# Step 2: Check Gatling configuration
# Per 02-01.github.instructions.md:
# ci-load uses Full Docker stack
# Load tests in test/load/ directory
# Requires Java 21 LTS + Maven 3.9+
```

**Files to Investigate**:

- `.github/workflows/ci-load.yml`
- `test/load/` Gatling configurations
- Service endpoint URLs in load tests

**Acceptance Criteria**:

- ✅ Local Act run succeeds with load tests
- ✅ GitHub Actions run succeeds
- ✅ Gatling scenarios execute successfully
- ✅ Load test results uploaded as artifacts
- ✅ Performance metrics reported correctly

**Validation Commands**:

```bash
# Local validation
go run ./cmd/workflow -workflows=load

# Manual Gatling execution
cd test/load
mvn gatling:test
```

---

### P1.2: Fix ci-sast Workflow

**Priority**: 8-LOW (LAST - slowest workflow)
**Effort**: 30 minutes
**Root Cause**: Static analysis tool configuration

**Current Failure Symptoms**:

- Static analysis tools reporting false positives
- Tool version mismatches
- Analysis timeout issues

**Implementation Strategy**:

```bash
# Step 1: Run locally
go run ./cmd/workflow -workflows=sast

# Step 2: Check static analysis tools
# Common tools: gosec, staticcheck, etc.
# Per 02-01.github.instructions.md:
# ci-sast has no service dependencies
```

**Files to Investigate**:

- `.github/workflows/ci-sast.yml`
- Static analysis tool configurations
- SARIF report generation

**Acceptance Criteria**:

- ✅ Local Act run succeeds with SAST results
- ✅ GitHub Actions run succeeds
- ✅ SARIF reports uploaded to Security tab
- ✅ No critical false positives
- ✅ Analysis completes within timeout

**Validation Commands**:

```bash
# Local validation
go run ./cmd/workflow -workflows=sast

# Manual static analysis
gosec ./...
staticcheck ./...
```

---

## Common Patterns Across All Workflows

### Diagnostic Logging (MANDATORY)

Per 02-01.github.instructions.md, steps >10s MUST include timing:

```yaml
- name: Execute long-running step
  run: |
    echo "📋 Starting: Long operation"
    START_TIME=$(date +%s)

    # ... actual operation ...

    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    echo "⏱️ Duration: ${DURATION}s"
    echo "✅ Complete: Long operation"
```

**Emojis for consistency**:

- 📋 Start of operation
- 📅 Timestamps
- ⏱️ Duration
- ✅ Success
- ❌ Error

### Artifact Management

```yaml
- name: Upload artifacts
  if: always()  # Upload even on failure
  uses: actions/upload-artifact@v5.0.0
  with:
    name: workflow-results
    path: ./test-output/
    retention-days: 1  # Temporary artifacts
```

### GitHub CLI for Debugging

```bash
# List recent workflow runs
gh run list --workflow=ci-coverage.yml --limit 5

# View specific run details
gh run view <run-id>

# Download artifacts for analysis
gh run download <run-id>

# Re-run failed jobs
gh run rerun <run-id> --failed
```

## Progress Tracking

After completing each task, update `PROGRESS.md`:

```bash
# Edit PROGRESS.md to mark task complete
# Update executive summary percentages
# Commit and push
git add specs/001-cryptoutil/PROGRESS.md
git commit -m "docs(speckit): mark P1.X complete"
git push
```

## Validation Checklist

Before marking Phase 1 complete, verify:

- [ ] All 8 workflows passing in GitHub Actions
- [ ] Each workflow completed in priority order
- [ ] PROGRESS.md updated with all P1.1-P1.8 marked complete
- [ ] No new failures introduced
- [ ] Workflow execution times reasonable
- [ ] Artifacts uploaded correctly for each workflow
- [ ] SARIF reports visible in Security tab (ci-sast)

## Next Phase

After Phase 1 complete:

- Proceed to Phase 2: Complete Deferred I2 Features
- Use PHASE2-IMPLEMENTATION.md guide
- Update PROGRESS.md executive summary
