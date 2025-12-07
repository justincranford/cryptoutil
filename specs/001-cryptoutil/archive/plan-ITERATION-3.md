# cryptoutil Implementation Plan - Iteration 3

## Overview

This plan outlines the technical implementation approach for Iteration 3 deliverables:

1. **D1: CI/CD Reliability Fixes** - Fix 8 failing workflows (27% → 100% pass rate)
2. **D2: Deferred Work Completion** - JOSE Docker, CA OCSP/EST, unified E2E tests
3. **D3: Test Methodology Enhancements** - Benchmarks, fuzz tests, property-based tests
4. **D4: Documentation Cleanup** - Process DELETE-ME files, update runbooks

**Estimated Total Effort**: ~40 hours (1 week sprint)

---

## CRITICAL: Test Concurrency Requirements

**!!! NEVER use `-p=1` or `-parallel=1` in test commands !!!**
**!!! ALWAYS use concurrent test execution with `-shuffle=on` !!!**

**Mandatory Test Execution**:

```bash
# CORRECT
go test ./... -cover -shuffle=on

# WRONG - NEVER DO THIS
go test ./... -p=1  # ❌ Hides concurrency bugs
```

**Test Data Isolation**:

- ✅ ALWAYS use UUIDv7 for test data uniqueness
- ✅ ALWAYS use dynamic ports (port 0 pattern)
- ✅ ALWAYS use TestMain for shared dependencies
- ✅ Real dependencies preferred (test containers, in-memory services)
- ✅ Mocks ONLY for hard-to-reach corner cases

**Rationale**: Concurrent tests provide fastest execution and reveal production concurrency bugs.

---

## Phase 1: Critical CI/CD Fixes (Days 1-2, ~16 hours)

### 1.1 Critical Workflow Failures

**Objective**: Fix 5 critical workflow failures blocking CI/CD pipeline

**Affected Workflows**:

- ci-race.yml (100% failure rate)
- ci-e2e.yml (100% failure rate)
- ci-load.yml (100% failure rate)
- ci-dast.yml (100% failure rate)
- ci-coverage.yml (80% failure rate)

### 1.2 Implementation Order

| Order | Task ID | Task | Dependencies | LOE |
|-------|---------|------|--------------|-----|
| 1 | ITER3-001 | Fix DATA RACE in CA handler (handler_comprehensive_test.go:1502) | None | 4h |
| 2 | ITER3-002 | Increase Identity ORM coverage 67.5% → 95% | None | 4h |
| 3 | ITER3-003 | Fix consent_decision_repository_test.go:160 test failure | ITER3-002 | 2h |
| 4 | ITER3-004 | Debug E2E/DAST/Load Docker Compose startup failures | None | 4h |
| 5 | ITER3-005 | Verify all 11 workflows pass after fixes | ITER3-001 to ITER3-004 | 2h |

**Phase 1 Subtotal**: ~16 hours

### 1.3 Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Race Detection | Fix with mutex/sync primitives | Constitution requires race-free code |
| Coverage Strategy | Add missing test cases | Target ≥95% per constitution v2.0.0 |
| E2E Debugging | Add diagnostic logging + health checks | Service startup timing issues |
| Verification | Run full workflow matrix locally | Use cmd/workflow for fast iteration |

### 1.4 Risk Mitigation

| Risk | Impact | Mitigation Strategy |
|------|--------|---------------------|
| Race fix breaks functionality | HIGH | Use table-driven tests + t.Parallel() to verify thread safety |
| Coverage tests slow CI | MEDIUM | Use selective test execution for local dev |
| E2E startup timing fragile | MEDIUM | Add retries + exponential backoff to health checks |
| Workflow fixes cause regressions | HIGH | Run full matrix locally before pushing |

---

## Phase 2: Deferred Work Completion (Days 3-4, ~15 hours)

### 2.1 Iteration 2 Deferred Features

**Objective**: Complete 4 deferred features from iteration 2 (83% → 100% completion)

**Features**:

- JOSE Docker Integration (2h)
- CA OCSP Handler (6h, RFC 6960)
- CA EST Handler (4h, RFC 7030)
- Unified E2E Test Suite (3h)

### 2.2 Implementation Order

| Order | Task ID | Task | Dependencies | LOE |
|-------|---------|------|--------------|-----|
| 1 | ITER3-006 | Create JOSE Dockerfile + compose.yml | Phase 1 complete | 2h |
| 2 | ITER3-007 | Implement CA OCSP handler (RFC 6960) | Phase 1 complete | 6h |
| 3 | ITER3-008 | Implement CA EST handler (RFC 7030) | Phase 1 complete | 4h |
| 4 | ITER3-009 | Create unified E2E test suite (JOSE + CA + Identity) | ITER3-006 to ITER3-008 | 3h |

**Phase 2 Subtotal**: ~15 hours

| Order | Task ID | Task | Dependencies | LOE |
|-------|---------|------|--------------|-----|
| 1 | TASK-6 | [Task description] | Phase 1 complete | 4h |
| 2 | TASK-7 | [Task description] | TASK-6 | 6h |
| 3 | TASK-8 | [Task description] | TASK-7 | 3h |

**Phase 2 Subtotal**: ~[H] hours

### 2.3 Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| JOSE Docker | Alpine multi-stage build | Consistency with CA/Identity |
| OCSP Handler | go.mozilla.org/pkcs7 | CMS library per CLARIFICATIONS.md #2 |
| EST Handler | RFC 7030 simple profile | Minimal viable implementation |
| E2E Framework | internal/test/e2e/ pattern | Reuse existing infrastructure |

### 2.4 Risk Mitigation

| Risk | Impact | Mitigation Strategy |
|------|--------|---------------------|
| OCSP/EST scope creep | HIGH | Stick to basic profiles, defer advanced features |
| E2E test flakiness | MEDIUM | Use health checks + retries, dynamic port allocation |
| Docker build breakage | MEDIUM | Test build locally before pushing |

---

## Phase 3: Test Methodology Enhancements (Day 5, ~9 hours)

### 3.1 Comprehensive Testing Coverage

**Objective**: Add benchmarks, fuzz tests, property-based tests per constitution v2.0.0

**Testing Types**:

- Benchmarks: All cryptographic operations (mandatory)
- Fuzz tests: All parsers/validators (≥15s, mandatory)
- Property-based: Crypto invariants (recommended)
- Manual mutation: Critical paths (gremlins blocked by v0.6.0 bug)

### 3.2 Implementation Order

| Order | Task ID | Task | Dependencies | LOE |
|-------|---------|------|--------------|-----|
| 1 | ITER3-010 | Add benchmarks for internal/common/crypto/keygen | Phase 2 complete | 2h |
| 2 | ITER3-011 | Add benchmarks for internal/crypto/* | Phase 2 complete | 2h |
| 3 | ITER3-012 | Add fuzz tests for internal/identity/authz parsers | Phase 2 complete | 2h |
| 4 | ITER3-013 | Add fuzz tests for internal/jose parsers | Phase 2 complete | 1h |
| 5 | ITER3-014 | Add property-based tests (gopter) for crypto invariants | Phase 2 complete | 2h |

**Phase 3 Subtotal**: ~9 hours

### 3.3 Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Benchmark scope | All crypto operations | Mandatory per constitution v2.0.0 |
| Fuzz duration | ≥15s per test | Minimum per testing instructions |
| Property framework | gopter | Recommended for Go, mature library |
| Mutation testing | Manual + document gremlins blocker | Tool crashes (v0.6.0 bug) |

### 3.4 Risk Mitigation

| Risk | Impact | Mitigation Strategy |
|------|--------|---------------------|
| Gremlins tool broken | HIGH | Manual mutation testing for critical paths, document blocker |
| Fuzz tests too slow | MEDIUM | Run in CI only, use -fuzztime flag for local dev |
| Property tests complex | LOW | Start with simple invariants (round-trip, idempotence) |

---

## Phase 4: Documentation & Optimization (Day 5, ~4 hours)

### 4.1 Documentation Cleanup

**Objective**: Process DELETE-ME files, update runbooks, consolidate lessons

**Files to Process**:

- DELETE-ME-LATER-SLOW-TEST-PACKAGES.md → Extract slow package data
- DELETE-ME-LATER-CROSS-REF-SPECKIT-COPILOT-TEMPLATE.md → Already applied
- NOT-FINISHED.md → Consolidate into specs/003 tracking

### 4.2 Implementation Order

| Order | Task ID | Task | Dependencies | LOE |
|-------|---------|------|--------------|-----|
| 1 | ITER3-015 | Extract slow package data from DELETE-ME file | None | 1h |
| 2 | ITER3-016 | Update runbooks with workflow analysis findings | Phase 1 complete | 1h |
| 3 | ITER3-017 | Delete processed DELETE-ME files | ITER3-015 | 0.5h |
| 4 | ITER3-018 | Consolidate NOT-FINISHED.md into PROGRESS.md | Phase 2 complete | 1h |
| 5 | ITER3-019 | Apply workflow optimizations (path filters, caching) | Phase 1 complete | 0.5h |

**Phase 4 Subtotal**: ~4 hours

### 4.3 Workflow Optimizations

**Objective**: Reduce CI feedback loop from 28min → <10min

**Optimization Strategies**:

| Optimization | Target Workflow | Expected Savings |
|--------------|----------------|------------------|
| Path filters (avoid unnecessary runs) | All workflows | ~5min avg |
| Parallel lint + test execution | ci-quality | ~3min |
| Go module caching (actions/setup-go@v6) | All Go workflows | ~2min |
| Selective test execution pattern | ci-coverage | ~10min |

### 4.4 Risk Mitigation

| Risk | Impact | Mitigation Strategy |
|------|--------|---------------------|
| Path filters too aggressive | MEDIUM | Conservative filters, test thoroughly |
| Caching stale dependencies | LOW | Cache invalidation on go.mod changes |
| Documentation drift | MEDIUM | Link docs to code, automated validation |

---

## Testing Strategy

### Unit Tests (≥95% coverage production, ≥100% infrastructure/utility)

**File Naming**: `*_test.go`

**Requirements**:

- Table-driven tests with `t.Parallel()`
- Test helpers marked with `t.Helper()`
- No magic values (use runtime UUIDv7 or magic constants)
- Dynamic port allocation (port 0 pattern)

**Coverage Targets by Package Type**:
- Production code: ≥95%
- Infrastructure (cicd): ≥100%
- Utility code: 100%

### Integration Tests

**File Naming**: `*_integration_test.go`

**Build Tag**: `//go:build integration`

**Requirements**:
- Docker Compose environment
- Real database (PostgreSQL + SQLite tests)
- Full API workflows
- Cleanup after tests

### Benchmark Tests (All hot paths)

**File Naming**: `*_bench_test.go`

**Requirements**:
- All cryptographic operations
- All API endpoints
- Database operations
- Baseline metrics documented

**Execution**:
```bash
go test -bench=. -benchmem ./internal/product/...
```

### Fuzz Tests (All parsers/validators)

**File Naming**: `*_fuzz_test.go`

**Requirements**:
- Unique fuzz function names (not substrings)
- All input parsers
- All validators
- Minimum 15s fuzz time

**Execution**:
```bash
go test -fuzz=FuzzFunctionName -fuzztime=15s ./internal/product/...
```

### Property-Based Tests (Invariants)

**Library**: gopter

**Requirements**:
- Round-trip encoding/decoding

**Requirements**:

- Round-trip encoding/decoding
- Invariant validation (e.g., encrypt(decrypt(x)) == x)
- Cryptographic properties

### Mutation Tests (≥80% mutation score - BLOCKED)

**Tool**: gremlins v0.6.0 (BLOCKED - crashes with panic)

**Status**: Tool installed but unusable, documented in CLARIFICATIONS.md #11

**Workaround**: Manual mutation testing for critical cryptographic operations

**Requirements**:

- Baseline per package (when tool fixed)
- Target ≥80% mutation score
- Manual testing interim: modify critical crypto code, verify tests fail

**Execution**:

```bash
# When tool fixed
gremlins unleash --tags=!integration
```

### E2E Tests

**File Naming**: `*_e2e_test.go` or in `test/e2e/`

**Requirements**:

- Full service stack (JOSE + CA + Identity + PostgreSQL + OTEL)
- Real telemetry infrastructure
- Docker Compose with health checks
- Unified test suite covering cross-service workflows

---

## Quality Gates

### Pre-Commit Gates

- [ ] `go build ./...` passes clean
- [ ] `golangci-lint run --fix` resolves all auto-fixable issues
- [ ] `golangci-lint run` passes with 0 errors
- [ ] File sizes ≤500 lines (refactor if exceeded)
- [ ] UTF-8 without BOM encoding
- [ ] No new TODOs without tracking

### Pre-Push Gates

- [ ] `go test ./...` passes all tests
- [ ] Coverage ≥95% production, ≥100% infrastructure/utility
- [ ] All benchmarks run successfully (`go test -bench=. -benchmem`)
- [ ] Dependency checks pass (`go list -u -m all`)
- [ ] Pre-commit hooks pass

### Pre-Merge Gates (CI/CD)

- [ ] All 11 CI workflows passing (100% pass rate)
- [ ] Coverage ≥95% production, ≥100% infrastructure/utility
- [ ] Code review approved (if applicable)
- [ ] Integration tests passing
- [ ] E2E tests passing (all 3 services)
- [ ] Docker Compose deployment healthy
- [ ] Documentation updated (runbooks, README)

---

## Success Milestones

| Day | Milestone | Deliverable | Verification |
|-----|-----------|-------------|--------------|
| 1-2 | Phase 1 Complete | All 5 critical workflows fixed | 11/11 workflows passing |
| 3-4 | Phase 2 Complete | 4 deferred features implemented | JOSE Docker, OCSP, EST, E2E suite |
| 5 | Phase 3 Complete | Test enhancements + docs cleanup | Benchmarks, fuzz, property tests added |
| 5 | Iteration 3 Done | All deliverables verified | CHECKLIST-ITERATION-003.md ✅ |

**Overall Success Criteria**:

- ✅ Workflow pass rate: 11/11 (100%)
- ✅ CI feedback loop: <10 minutes
- ✅ Coverage: ≥95% production, ≥100% infrastructure
- ✅ Deferred tasks: 4/4 completed (JOSE Docker, OCSP, EST, E2E)
- ✅ Test coverage: Benchmarks + fuzz + property tests operational
- ✅ Documentation: DELETE-ME files processed, runbooks updated

---

## Evidence-Based Completion Checklist

### Code Quality

- [ ] `go build ./...` clean
- [ ] `golangci-lint run` clean (0 errors)
- [ ] No new TODOs without tracking in tasks.md
- [ ] File sizes within limits (≤500 lines)
- [ ] UTF-8 without BOM encoding

### Test Coverage

- [ ] `go test ./... -shuffle=on` passes (concurrent execution)
- [ ] Coverage ≥95% production: `go test -cover ./internal/product/...`
- [ ] Coverage ≥100% infrastructure: `go test -cover ./internal/cmd/cicd/...`
- [ ] Coverage 100% utility: `go test -cover ./internal/common/util/...`
- [ ] No skipped tests without tracking

### Benchmarks

- [ ] All cryptographic operations benchmarked
- [ ] All hot path handlers benchmarked
- [ ] Baseline metrics documented
- [ ] No performance regressions

### Fuzz Tests

- [ ] All parsers fuzzed for ≥15s
- [ ] All validators fuzzed for ≥15s
- [ ] Crash-free fuzz execution

### Mutation Tests

- [ ] Gremlins baseline report created
- [ ] Mutation score ≥80% per package
- [ ] Weak tests identified and improved

### Integration

- [ ] Docker Compose deploys successfully
- [ ] All services report healthy
- [ ] E2E demo script passes
- [ ] Inter-service communication working

### Documentation

- [ ] README.md updated with new features
- [ ] API documentation generated (OpenAPI)
- [ ] Runbooks created for operations
- [ ] PROGRESS.md up-to-date
- [ ] EXECUTIVE-SUMMARY.md created
- [ ] CHECKLIST-ITERATION-NNN.md complete

---

## Dependency Management

### Version Requirements

- Go: 1.25.4+
- PostgreSQL: 14+
- Docker: 24+
- Docker Compose: v2+
- golangci-lint: v2.6.2+
- Node: v24.11.1+ (for Gatling load tests)

### Updating Dependencies

```bash
# Check for updates
go list -u -m all | grep '\[.*\]$'

# Update incrementally
go get -u [package]
go mod tidy

# Test after each update
go test ./...
golangci-lint run
```

---

## Workflow Integration

### CI/CD Pipelines

| Workflow | Trigger | Purpose | Duration |
|----------|---------|---------|----------|
| ci-quality | PR, push | Linting, formatting, builds | ~5 min |
| ci-coverage | PR, push | Test coverage | ~10 min |
| ci-benchmark | PR, push | Performance benchmarks | ~5 min |
| ci-fuzz | Scheduled | Fuzz testing | ~15 min |
| ci-race | PR, push | Race detection | ~15 min |
| ci-sast | PR, push | Static security | ~5 min |
| ci-dast | Scheduled | Dynamic security | ~10 min |
| ci-e2e | PR, push | Integration tests | ~20 min |

### Artifact Management

- Upload test reports: `actions/upload-artifact@v5.0.0`
- Upload SARIF: `github/codeql-action/upload-sarif@v3`
- Retention: 1 day (temporary), 30 days (reports)

---

## Post-Mortem Template

### Lessons Learned

**What Went Well**:

- Systematic workflow analysis with `gh` CLI provided actionable insights
- Constitution v2.0.0 quality standards clarified expectations
- Template-based approach accelerated iteration 3 planning
- Comprehensive testing strategy (benchmark + fuzz + property) matured

**What Needs Improvement**:

- Gremlins mutation testing tool stability (v0.6.0 crashes)
- CI/CD feedback loop optimization (28min → target <10min)
- Service startup reliability in Docker Compose (E2E/DAST/Load failures)
- Test execution performance (some packages >30s per test suite)

**Action Items for Next Iteration**:

- Monitor gremlins releases for v0.6.1+ stability fixes
- Apply path filters to all workflows to avoid unnecessary runs
- Investigate alternative mutation testing tools (go-mutesting, Stryker4s)
- Add retry logic + exponential backoff to health checks
- Document slow test packages and optimization strategies

---

## Template Usage Notes

**For LLM Agents**: This plan for iteration 3 includes:

- ✅ 4 phases with realistic LOE estimates (~40h total)
- ✅ Critical CI/CD fixes prioritized (Phase 1: 16h)
- ✅ Deferred work completion (Phase 2: 15h)
- ✅ Test methodology enhancements (Phase 3: 9h)
- ✅ Documentation cleanup + optimizations (Phase 4: 4h)
- ✅ Technical decisions with rationales
- ✅ Risk mitigation strategies
- ✅ Comprehensive testing strategy (unit, integration, benchmark, fuzz, property, mutation workaround, E2E)
- ✅ Quality gates (pre-commit, pre-push, pre-merge)
- ✅ Success milestones with verification criteria
- ✅ Evidence-based completion approach

**Iteration 3 Specifics**:

- Focuses on CI/CD reliability (8 failing workflows → 100% pass rate)
- Addresses gremlins mutation testing blocker with manual workaround
- Leverages workflow-analysis.md findings for optimization
- Completes iteration 2 deferred work (83% → 100%)
- Implements constitution v2.0.0 testing requirements

**References**:

- spec.md: Iteration 3 requirements and deliverables
- tasks.md: Granular task breakdown (ITER3-001 to ITER3-019)
- docs/workflow-analysis.md: CI/CD findings and action plan
- specs/002-cryptoutil/CLARIFICATIONS.md: Deferred work tracking
- .specify/memory/constitution.md v2.0.0: Quality requirements
