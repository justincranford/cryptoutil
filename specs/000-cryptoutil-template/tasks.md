# cryptoutil Tasks Template - Iteration NNN

## Task Breakdown

This document provides granular task tracking for Iteration NNN implementation.

**Total Phases**: 3
**Total Tasks**: [X]
**Estimated Effort**: ~[Y] hours

---

## CRITICAL: Test Concurrency Requirements

**!!! NEVER use `-p=1` or `-parallel=1` in test commands !!!**
**!!! ALWAYS use concurrent test execution with `-shuffle=on` !!!**

**Test Execution Commands**:

```bash
# CORRECT - Concurrent with shuffle
go test ./... -cover -shuffle=on

# WRONG - Sequential execution (hides bugs!)
go test ./... -p=1  # ❌ NEVER DO THIS
go test ./... -parallel=1  # ❌ NEVER DO THIS
```

**Test Data Isolation Requirements**:

- ✅ ALWAYS use UUIDv7 for all test data (thread-safe, process-safe)
- ✅ ALWAYS use dynamic ports (port 0 pattern for test servers)
- ✅ ALWAYS use TestMain for dependencies (start once per package)
- ✅ Real dependencies preferred (PostgreSQL containers, in-memory services)
- ✅ Mocks only for hard-to-reach corner cases or truly external dependencies

**Why Concurrent Testing is Mandatory**:

1. Fastest test execution (parallel tests = faster feedback)
2. Reveals production bugs (race conditions, deadlocks, data conflicts)
3. Production validation (if tests can't run concurrently, production code can't either)
4. Quality assurance (concurrent tests = higher confidence)

---

## Phase 1: [Phase Name]

### Overview

**Goals**: [Phase goals]
**Duration**: Week X-Y
**Estimated Effort**: ~[H] hours

### Task List

| ID | Title | Description | Priority | Status | LOE | Dependencies |
|----|-------|-------------|----------|--------|-----|--------------|
| TASK-1 | [Task title] | [Detailed description] | HIGH | ❌ Not Started | 2h | None |
| TASK-2 | [Task title] | [Detailed description] | HIGH | ❌ Not Started | 4h | TASK-1 |
| TASK-3 | [Task title] | [Detailed description] | MEDIUM | ❌ Not Started | 3h | TASK-2 |
| TASK-4 | [Task title] | [Detailed description] | LOW | ❌ Not Started | 2h | TASK-2 |

### Detailed Task Specifications

#### TASK-1: [Task Title]

**Description**: [Detailed description of what needs to be done]

**Acceptance Criteria**:

- [ ] [Criterion 1]
- [ ] [Criterion 2]
- [ ] [Criterion 3]

**Implementation Steps**:

1. [Step 1]
2. [Step 2]
3. [Step 3]

**Files to Create/Modify**:

- `path/to/file1.go`
- `path/to/file1_test.go`
- `path/to/file1_bench_test.go`

**Testing Requirements**:

- Unit tests with `t.Parallel()`
- Table-driven tests covering happy/sad paths
- Benchmark tests for performance
- Coverage ≥95%

**Evidence of Completion**:

- [ ] `go build ./...` passes
- [ ] `go test ./path/to/package` passes
- [ ] `golangci-lint run ./path/to/package` passes
- [ ] Coverage report shows ≥95%
- [ ] Code review approved

**Estimated LOE**: 2 hours

---

#### TASK-2: [Task Title]

**Description**: [Detailed description]

**Acceptance Criteria**:

- [ ] [Criterion 1]
- [ ] [Criterion 2]

**Implementation Steps**:

1. [Step 1]
2. [Step 2]

**Files to Create/Modify**:

- `path/to/file2.go`
- `path/to/file2_test.go`

**Testing Requirements**:

- Unit tests
- Integration tests (if applicable)
- Fuzz tests (if applicable)

**Evidence of Completion**:

- [ ] Tests passing
- [ ] Linting clean
- [ ] Coverage target met

**Estimated LOE**: 4 hours

**Dependencies**: TASK-1 must be complete

---

### Phase 1 Summary

| Metric | Target | Status |
|--------|--------|--------|
| Tasks Complete | 0/[X] | ❌ |
| Code Coverage | ≥95% | ❌ |
| LOE Consumed | 0/[H]h | ❌ |

---

## Phase 2: [Phase Name]

### Overview

**Goals**: [Phase goals]
**Duration**: Week Y-Z
**Estimated Effort**: ~[H] hours

### Task List

| ID | Title | Description | Priority | Status | LOE | Dependencies |
|----|-------|-------------|----------|--------|-----|--------------|
| TASK-5 | [Task title] | [Detailed description] | HIGH | ❌ Not Started | 6h | Phase 1 complete |
| TASK-6 | [Task title] | [Detailed description] | MEDIUM | ❌ Not Started | 4h | TASK-5 |
| TASK-7 | [Task title] | [Detailed description] | LOW | ❌ Not Started | 3h | TASK-5 |

### Detailed Task Specifications

#### TASK-5: [Task Title]

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

**Estimated LOE**: 6 hours

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
  - [ ] Coverage ≥95% (production) / ≥98% (infrastructure/utility)
  - [ ] Benchmarks for hot paths
  - [ ] Fuzz tests for parsers/validators

- [ ] **Documentation**
  - [ ] GoDoc comments on public APIs
  - [ ] README updated if needed
  - [ ] implement/DETAILED.md Section 2 (timeline) updated

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
- Infrastructure (cicd): ≥98%
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

### Iteration NNN Complete When

- [ ] All tasks status = ✅ Complete
- [ ] `go build ./...` passes clean
- [ ] `golangci-lint run` passes with 0 errors
- [ ] `go test ./... -shuffle=on` passes (concurrent execution)
- [ ] Coverage ≥95% production, ≥98% infrastructure/utility
- [ ] All benchmarks run successfully
- [ ] All fuzz tests run for ≥15s
- [ ] Gremlins mutation score ≥80%
- [ ] Docker Compose deployment healthy
- [ ] Integration tests passing
- [ ] E2E demo script working
- [ ] implement/DETAILED.md Section 2 (timeline) up-to-date
- [ ] implement/EXECUTIVE.md created/updated
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
- ✅ Coverage targets: 95% production, 98% infrastructure/utility

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
