# cryptoutil Tasks - Iteration 3

## Task Breakdown

This document provides granular task tracking for Iteration 3 implementation (CI/CD reliability + deferred work).

**Total Phases**: 4
**Total Tasks**: 19
**Estimated Effort**: ~44 hours (1 week sprint)

---

## Phase 1: Critical CI/CD Fixes (Days 1-2)

### Overview

**Goals**: Fix 5 critical failing workflows, increase pass rate from 27% to 100%
**Duration**: Days 1-2
**Estimated Effort**: ~16 hours

### Task List

| ID | Title | Description | Priority | Status | LOE | Dependencies |
|----|-------|-------------|----------|--------|-----|--------------|
| ITER3-001 | Fix DATA RACE in CA handler | Fix goroutine race at handler_comprehensive_test.go:1502 | CRITICAL | ❌ Not Started | 4h | None |
| ITER3-002 | Increase Identity ORM coverage | Increase coverage from 67.5% to ≥95% | CRITICAL | ❌ Not Started | 4h | None |
| ITER3-003 | Fix consent decision test | Fix test failure at consent_decision_repository_test.go:160 | HIGH | ❌ Not Started | 2h | ITER3-002 |
| ITER3-004 | Debug E2E/DAST/Load startup | Fix Docker Compose service startup failures | CRITICAL | ❌ Not Started | 4h | None |
| ITER3-005 | Verify all workflows pass | Run full workflow matrix locally + CI | HIGH | ❌ Not Started | 2h | ITER3-001 to ITER3-004 |

### Detailed Task Specifications

#### ITER3-001: Fix DATA RACE in CA Handler

**Description**: Race condition detected between goroutines 339 and 341 in `internal/ca/server/handler/handler_comprehensive_test.go:1502` causing ci-race.yml workflow 100% failure rate.

**Acceptance Criteria**:

- [ ] Race condition eliminated (verified with `go test -race`)
- [ ] All CA handler tests pass with `-race` flag
- [ ] ci-race.yml workflow passes in CI

**Implementation Steps**:

1. Analyze race detector output from workflow logs (goroutines 339/341)
2. Identify shared state accessed without synchronization
3. Add mutex/sync primitives or refactor to eliminate shared mutable state
4. Run `go test -race ./internal/ca/server/handler/...` locally
5. Verify fix with table-driven parallel tests (`t.Parallel()`)
6. Push changes and monitor ci-race.yml workflow

**Files to Create/Modify**:

- `internal/ca/server/handler/handler_comprehensive_test.go` (line 1502 area)
- Potentially `internal/ca/server/handler/handler.go` (if shared state in handler)

**Testing Requirements**:

- All existing tests must pass with `-race` flag
- Table-driven tests with `t.Parallel()` to expose concurrency issues
- No new race conditions introduced

**Evidence of Completion**:

- [ ] `go test -race ./internal/ca/server/handler/...` passes locally
- [ ] ci-race.yml workflow passes in GitHub Actions
- [ ] No race detector warnings in logs
- [ ] Code review confirms proper synchronization

**Estimated LOE**: 4 hours

---

#### ITER3-002: Increase Identity ORM Coverage

**Description**: Identity ORM coverage at 67.5% (below 95% target), causing ci-coverage.yml failures. Add missing test cases for uncovered code paths.

**Acceptance Criteria**:

- [ ] Coverage ≥95% for all Identity ORM packages
- [ ] ci-coverage.yml workflow passes
- [ ] All new tests use table-driven pattern with `t.Parallel()`

**Implementation Steps**:

1. Run `go test -coverprofile=coverage.out ./internal/identity/...`
2. Analyze coverage report: `go tool cover -html=coverage.out`
3. Identify uncovered code paths (error handling, edge cases, validators)
4. Write table-driven tests for missing coverage
5. Verify coverage ≥95%: `go test -cover ./internal/identity/...`
6. Run ci-coverage.yml locally: `go run ./cmd/workflow -workflows=coverage`

**Files to Create/Modify**:

- `internal/identity/server/repository/sqlrepository/*_test.go` (likely)
- `internal/identity/domain/*_test.go` (validators, models)
- Potentially add integration tests if unit tests insufficient

**Testing Requirements**:

- Table-driven tests covering happy paths, error paths, edge cases
- Use `t.Parallel()` for all test cases
- Dynamic UUIDv7 for test data (no hardcoded values)
- Coverage ≥95% per package

**Evidence of Completion**:

- [ ] `go test -cover ./internal/identity/...` shows ≥95%
- [ ] ci-coverage.yml workflow passes
- [ ] All tests pass: `go test ./internal/identity/...`
- [ ] `golangci-lint run ./internal/identity/...` passes

**Estimated LOE**: 4 hours

---

#### ITER3-003: Fix Consent Decision Test Failure

**Description**: Test failure in `consent_decision_repository_test.go:160` blocking ci-coverage workflow. Related to ITER3-002 Identity ORM work.

**Acceptance Criteria**:

- [ ] Test passes consistently (100% success rate)
- [ ] Root cause identified and documented
- [ ] No regressions in related tests

**Implementation Steps**:

1. Review test failure logs from ci-coverage workflow
2. Run test locally: `go test -v -run=TestConsentDecision ./internal/identity/...`
3. Identify failure cause (likely data setup, assertions, or timing)
4. Fix test logic or underlying code
5. Verify fix: run test 10 times to ensure no flakiness
6. Commit fix with detailed explanation in commit message

**Files to Create/Modify**:

- `internal/identity/server/repository/sqlrepository/consent_decision_repository_test.go` (line 160 area)

**Testing Requirements**:

- Test must pass consistently (no flakiness)
- Use proper test fixtures and cleanup
- Table-driven test pattern

**Evidence of Completion**:

- [ ] Test passes 10 consecutive runs locally
- [ ] ci-coverage.yml passes in CI
- [ ] No related test failures

**Estimated LOE**: 2 hours

**Dependencies**: ITER3-002 (likely shares ORM coverage work)

---

#### ITER3-004: Debug E2E/DAST/Load Docker Compose Startup

**Description**: E2E, DAST, and Load testing workflows have 100% failure rate due to Docker Compose service startup issues. Services fail health checks or timeout during initialization.

**Acceptance Criteria**:

- [ ] All services start successfully in Docker Compose
- [ ] Health checks pass within timeout
- [ ] ci-e2e.yml, ci-dast.yml, ci-load.yml workflows pass

**Implementation Steps**:

1. Review workflow logs for service startup errors
2. Add diagnostic logging to service startup (application.go files)
3. Increase health check timeouts/retries if timing issues
4. Add exponential backoff to health check probes
5. Test locally: `docker compose -f deployments/compose/compose.yml up`
6. Verify health: `docker compose ps` (all services "healthy")
7. Run workflows locally: `go run ./cmd/workflow -workflows=e2e,dast,load`

**Files to Create/Modify**:

- `deployments/compose/compose.yml` (health check configurations)
- `internal/*/server/application/application.go` (startup logging)
- `.github/workflows/ci-e2e.yml` (health check verification)
- `.github/workflows/ci-dast.yml` (health check verification)
- `.github/workflows/ci-load.yml` (health check verification)

**Testing Requirements**:

- Manual verification: services start without errors
- Health checks complete within configured timeout
- Workflows pass end-to-end

**Evidence of Completion**:

- [ ] `docker compose up` succeeds, all services healthy
- [ ] ci-e2e.yml passes in CI
- [ ] ci-dast.yml passes in CI
- [ ] ci-load.yml passes in CI
- [ ] Logs show clean startup (no errors/warnings)

**Estimated LOE**: 4 hours

---

#### ITER3-005: Verify All Workflows Pass

**Description**: After fixes from ITER3-001 to ITER3-004, verify all 11 CI workflows pass (100% success rate).

**Acceptance Criteria**:

- [ ] All 11 workflows pass in single commit
- [ ] Workflow pass rate: 100% (11/11)
- [ ] CI feedback loop ≤10 minutes

**Implementation Steps**:

1. Commit all Phase 1 fixes
2. Push to GitHub and monitor workflow runs
3. Use `gh run list` to check status
4. Use `gh run view --log-failed` for any failures
5. Document workflow execution times
6. Create summary report of fixes applied

**Files to Create/Modify**:

- `specs/003-cryptoutil/PROGRESS.md` (track Phase 1 completion)

**Testing Requirements**:

- All workflows must pass
- No flaky tests (run matrix multiple times if needed)

**Evidence of Completion**:

- [ ] `gh run list` shows 11/11 workflows passing
- [ ] Screenshot of GitHub Actions dashboard (all green)
- [ ] CI feedback loop documented (target <10min)
- [ ] Phase 1 marked complete in PROGRESS.md

**Estimated LOE**: 2 hours

**Dependencies**: ITER3-001, ITER3-002, ITER3-003, ITER3-004

---

### Phase 1 Summary

| Metric | Target | Status |
|--------|--------|--------|
| Tasks Complete | 0/5 | ❌ |
| Workflow Pass Rate | 100% (11/11) | ❌ (currently 27%, 3/11) |
| LOE Consumed | 0/16h | ❌ |
| Critical Issues Fixed | 5 | ❌ (0/5) |

---

## Phase 2: Deferred Work Completion (Days 3-4)

### Overview

**Goals**: Complete 4 deferred features from iteration 2 (83% → 100% completion)
**Duration**: Days 3-4
**Estimated Effort**: ~15 hours

### Task List

| ID | Title | Description | Priority | Status | LOE | Dependencies |
|----|-------|-------------|----------|--------|-----|--------------|
| ITER3-006 | JOSE Docker Integration | Create Dockerfile, compose.yml, demo for JOSE | HIGH | ❌ Not Started | 2h | Phase 1 complete |
| ITER3-007 | CA OCSP Handler | Implement RFC 6960 OCSP responder | HIGH | ❌ Not Started | 6h | Phase 1 complete |
| ITER3-008 | CA EST Handler | Implement RFC 7030 EST server (simple profile) | MEDIUM | ❌ Not Started | 4h | Phase 1 complete |
| ITER3-009 | Unified E2E Test Suite | Cross-service E2E tests (JOSE + CA + Identity) | HIGH | ❌ Not Started | 3h | ITER3-006 to ITER3-008 |

### Detailed Task Specifications
**Dependencies**: Phase 1 complete

---

### Phase 2 Summary

| Metric | Target | Status |
|--------|--------|--------|
| Tasks Complete | 0/[X] | ❌ |
| Code Coverage | ≥95% | ❌ |
| LOE Consumed | 0/[H]h | ❌ |

---

## Phase 3: [Phase Name]

### Overview

**Goals**: [Phase goals]
**Duration**: Week Z-W
**Estimated Effort**: ~[H] hours

### Task List

| ID | Title | Description | Priority | Status | LOE | Dependencies |
|----|-------|-------------|----------|--------|-----|--------------|
| TASK-8 | [Task title] | [Detailed description] | HIGH | ❌ Not Started | 4h | Phase 2 complete |
| TASK-9 | [Task title] | [Detailed description] | HIGH | ❌ Not Started | 2h | TASK-8 |
| TASK-10 | [Task title] | [Detailed description] | MEDIUM | ❌ Not Started | 6h | TASK-9 |

### Detailed Task Specifications

#### TASK-8: [Task Title]

**Description**: [Detailed description]

**Acceptance Criteria**:
- [ ] [Criterion 1]
- [ ] [Criterion 2]

**Implementation Steps**:
1. [Step 1]
2. [Step 2]

**Files to Create/Modify**:
- [List files]

**Testing Requirements**:
- [Testing approach]

**Evidence of Completion**:
- [Verification steps]

**Estimated LOE**: 4 hours

**Dependencies**: Phase 2 complete

---

### Phase 3 Summary

| Metric | Target | Status |
|--------|--------|--------|
| Tasks Complete | 0/[X] | ❌ |
| Code Coverage | ≥95% | ❌ |
| LOE Consumed | 0/[H]h | ❌ |

---

## Overall Iteration Summary

### Progress Metrics

| Phase | Tasks | Complete | Partial | Remaining | Progress |
|-------|-------|----------|---------|-----------|----------|
| Phase 1: [Name] | [X] | 0 | 0 | [X] | 0% ❌ |
| Phase 2: [Name] | [Y] | 0 | 0 | [Y] | 0% ❌ |
| Phase 3: [Name] | [Z] | 0 | 0 | [Z] | 0% ❌ |
| **Total** | **[X+Y+Z]** | **0** | **0** | **[X+Y+Z]** | **0%** ❌ |

### LOE Tracking

| Phase | Estimated | Actual | Variance | Notes |
|-------|-----------|--------|----------|-------|
| Phase 1 | [H]h | 0h | 0h | Not started |
| Phase 2 | [H]h | 0h | 0h | Not started |
| Phase 3 | [H]h | 0h | 0h | Not started |
| **Total** | **[H]h** | **0h** | **0h** | - |

---

## Task Dependencies Graph

```
TASK-1 ─→ TASK-2 ─→ TASK-3
                 │
                 └─→ TASK-4

Phase 1 Complete ─→ TASK-5 ─→ TASK-6
                            │
                            └─→ TASK-7

Phase 2 Complete ─→ TASK-8 ─→ TASK-9 ─→ TASK-10
```

---

## Priority Matrix

### HIGH Priority (Critical Path)

| ID | Task | Blocker For | Impact |
|----|------|-------------|--------|
| TASK-1 | [Task] | TASK-2, TASK-3, TASK-4 | Blocks Phase 1 |
| TASK-5 | [Task] | Phase 2 | Core functionality |

### MEDIUM Priority (Important)

| ID | Task | Reason |
|----|------|--------|
| TASK-3 | [Task] | [Reason] |
| TASK-6 | [Task] | [Reason] |

### LOW Priority (Nice to Have)

| ID | Task | Reason |
|----|------|--------|
| TASK-4 | [Task] | [Reason] |
| TASK-7 | [Task] | [Reason] |

---

## Quality Checklist

### Per-Task Quality Gates

For EACH task, verify:

- [ ] **Code Quality**
  - [ ] `go build ./...` passes
  - [ ] `golangci-lint run` passes with 0 errors
  - [ ] No new TODOs without tracking
  - [ ] File sizes ≤500 lines
  - [ ] UTF-8 without BOM encoding

- [ ] **Testing**
  - [ ] Unit tests with `t.Parallel()`
  - [ ] Table-driven tests
  - [ ] Coverage ≥95% (production) / ≥100% (infrastructure/utility)
  - [ ] Benchmarks for hot paths
  - [ ] Fuzz tests for parsers/validators

- [ ] **Documentation**
  - [ ] GoDoc comments on public APIs
  - [ ] README updated if needed
  - [ ] PROGRESS.md updated

---

## Risk Tracking

### Task-Specific Risks

| Task ID | Risk | Impact | Mitigation |
|---------|------|--------|------------|
| TASK-1 | [Risk] | HIGH/MEDIUM/LOW | [Mitigation] |
| TASK-5 | [Risk] | HIGH/MEDIUM/LOW | [Mitigation] |

### Phase-Level Risks

| Phase | Risk | Impact | Mitigation |
|-------|------|--------|------------|
| Phase 1 | [Risk] | HIGH/MEDIUM/LOW | [Mitigation] |
| Phase 2 | [Risk] | HIGH/MEDIUM/LOW | [Mitigation] |

---

## Iteration-Wide Testing Requirements

### Unit Tests

**Target Coverage**:
- Production code: ≥95%
- Infrastructure (cicd): ≥100%
- Utility code: 100%

**Requirements**:
- Table-driven tests
- `t.Parallel()` for all tests
- No magic values (use UUIDv7 or magic constants)
- Dynamic port allocation (port 0 pattern)

### Integration Tests

**Tag**: `//go:build integration`

**Requirements**:
- Docker Compose environment
- PostgreSQL and SQLite tests
- Full API workflows
- Cleanup after tests

### Benchmark Tests

**Files**: `*_bench_test.go`

**Requirements**:
- All cryptographic operations
- All hot path handlers
- Database operations
- Baseline metrics documented

### Fuzz Tests

**Files**: `*_fuzz_test.go`

**Requirements**:
- All input parsers
- All validators
- Minimum 15s fuzz time
- Unique function names (not substrings)

### Mutation Tests

**Tool**: gremlins

**Requirements**:
- Baseline per package
- Target ≥80% mutation score
- Regular execution

### E2E Tests

**Requirements**:
- Full service stack
- Real telemetry infrastructure
- Demo script automation

---

## Completion Evidence

### Iteration NNN Complete When:

- [ ] All tasks status = ✅ Complete
- [ ] `go build ./...` passes clean
- [ ] `golangci-lint run` passes with 0 errors
- [ ] `go test ./...` passes (with and without `-p=1`)
- [ ] Coverage ≥95% production, ≥100% infrastructure/utility
- [ ] All benchmarks run successfully
- [ ] All fuzz tests run for ≥15s
- [ ] Gremlins mutation score ≥80%
- [ ] Docker Compose deployment healthy
- [ ] Integration tests passing
- [ ] E2E demo script working
- [ ] PROGRESS.md up-to-date
- [ ] EXECUTIVE-SUMMARY.md created
- [ ] CHECKLIST-ITERATION-NNN.md complete
- [ ] No new TODOs without tracking

---

## Template Usage Notes

**For LLM Agents**: This tasks template includes:
- ✅ Granular task breakdown with LOE estimates
- ✅ Detailed task specifications with acceptance criteria
- ✅ Implementation steps and file lists
- ✅ Testing requirements per task
- ✅ Evidence-based completion verification
- ✅ Dependency tracking and visualization
- ✅ Priority matrix for task ordering
- ✅ Quality checklist per task
- ✅ Risk tracking per task and phase
- ✅ Comprehensive testing requirements (unit, integration, benchmark, fuzz, mutation, E2E)
- ✅ Coverage targets: 95% production, 100% infrastructure/utility

**Customization**:
- Adjust task granularity based on complexity
- Update LOE estimates from actual experience
- Add task-specific notes for complex items
- Update status as tasks progress (❌ → ⚠️ → ✅)

**Status Icons**:
- ❌ Not Started
- ⚠️ In Progress / Partial
- ✅ Complete

**References**:
- spec.md: Functional requirements
- plan.md: Implementation approach
- Constitution: Quality requirements
- Copilot Instructions: Coding patterns
