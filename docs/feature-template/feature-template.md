# Feature Implementation Master Template

**Purpose**: Reusable template for planning and executing complex feature implementations with LLM Agent autonomy
**Version**: 2.0
**Last Updated**: November 26, 2025
**Designed For**: Extended autonomous LLM Agent sessions (minimal user authorization, continuous execution until completion)
**Enhancements**: Evidence-based validation, single source of truth, progressive validation, foundation-before-features enforcement

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Goals and Objectives](#goals-and-objectives)
3. [Context and Baseline](#context-and-baseline)
4. [Architecture and Design](#architecture-and-design)
5. [Implementation Tasks](#implementation-tasks)
6. [Task Execution Instructions](#task-execution-instructions)
7. [Post-Mortem and Corrective Actions](#post-mortem-and-corrective-actions)
8. [Quality Gates and Acceptance Criteria](#quality-gates-and-acceptance-criteria)
9. [Risk Management](#risk-management)
10. [Success Metrics](#success-metrics)

---

## Executive Summary

### Feature Overview

**Feature Name**: `<FEATURE_NAME>`
**Feature ID**: `<PROJECT_PREFIX>-<FEATURE_ID>` (e.g., IDENTITY-V2, KMS-REFACTOR, CA-IMPL)
**Status**: `PLANNING | IN_PROGRESS | BLOCKED | COMPLETE`
**Priority**: `🔴 CRITICAL | ⚠️ HIGH | 🟡 MEDIUM | 🟢 LOW`

### Current Reality

**Problem Statement**: Clear articulation of the problem this feature solves

**Current State Analysis**:

- What exists today
- What's broken or missing
- Impact on users/system
- Technical debt incurred

**Production Blockers** (if applicable):

1. 🔴 Blocker 1: Description with impact
2. 🔴 Blocker 2: Description with impact
3. ⚠️ High-priority issue: Description

### Completion Metrics

| Metric | Count | Percentage | Status |
|--------|-------|------------|--------|
| **Fully Complete** | X/Y | Z% | ✅/⚠️/❌ |
| **Documented Complete but Has Gaps** | X/Y | Z% | ✅/⚠️/❌ |
| **Incomplete/Not Started** | X/Y | Z% | ✅/⚠️/❌ |
| **Total Tasks** | Y | 100% | - |

### Production Readiness Assessment

**Production Ready**: ✅ YES | ❌ NO | ⚠️ PARTIAL

**Rationale**: Brief explanation of production readiness status

### Remediation Approach

**Strategy**: High-level approach (e.g., "Foundation First", "Parallel Streams", "Incremental Migration")

**Timeline**: X days/weeks/months (assumes full-time focus or specific allocation)

**Effort Distribution**:

- Foundation: X% (Y days)
- Core Features: X% (Y days)
- Advanced Features: X% (Y days)
- Integration & Testing: X% (Y days)
- Documentation & Handoff: X% (Y days)

---

## Goals and Objectives

### Primary Goals

**Goal 1**: `<GOAL_NAME>`

- **Description**: What this goal achieves
- **Success Criteria**: Measurable outcomes
- **Priority**: CRITICAL/HIGH/MEDIUM/LOW
- **Dependencies**: What must be complete first
- **Risk Level**: LOW/MEDIUM/HIGH/CRITICAL

**Goal 2**: `<GOAL_NAME>`

- **Description**: What this goal achieves
- **Success Criteria**: Measurable outcomes
- **Priority**: CRITICAL/HIGH/MEDIUM/LOW
- **Dependencies**: What must be complete first
- **Risk Level**: LOW/MEDIUM/HIGH/CRITICAL

### Secondary Goals (Nice-to-Have)

**Goal N**: `<GOAL_NAME>`

- **Description**: What this goal achieves
- **Success Criteria**: Measurable outcomes
- **Priority**: LOW
- **Dependencies**: Primary goals 1-N complete

### Non-Goals (Out of Scope)

- **Non-Goal 1**: Explicitly what won't be addressed and why
- **Non-Goal 2**: Features deferred to future phases
- **Non-Goal 3**: Alternative approaches rejected

### Constraints and Boundaries

**Technical Constraints**:

- Dependency restrictions (e.g., "Use ONLY existing go.mod dependencies")
- Technology stack limitations
- Performance requirements
- Security requirements (e.g., "FIPS 140-3 compliance mandatory")

**Architectural Constraints**:

- Domain isolation rules (e.g., "Identity module CANNOT import KMS packages")
- Import restrictions enforced by tooling
- Code reuse boundaries
- Magic values management patterns

**Resource Constraints**:

- Timeline restrictions
- Budget limitations
- Team size/availability
- Infrastructure limitations

**Operational Constraints**:

- Backward compatibility requirements
- Migration strategy restrictions
- Deployment constraints
- Monitoring/observability requirements

---

## Context and Baseline

### Historical Context

**Previous Attempts**:

- Attempt 1: What was tried, why it failed, lessons learned
- Attempt 2: What was tried, why it failed, lessons learned

**Related Work**:

- Related Feature 1: How it connects to this feature
- Related Feature 2: Dependencies or interactions

**Evolution Timeline**:

- Phase 1 (Date Range): What was implemented, outcomes
- Phase 2 (Date Range): What was implemented, outcomes
- Current State (Date): Where we are today

### Baseline Assessment

**Current Implementation Status**:

- ✅ Complete Components: List with verification evidence
- ⚠️ Partial Components: List with gap analysis
- ❌ Missing Components: List with impact assessment

**Code Analysis**:

- Total files: X files
- Total lines: Y lines (X production, Y test, Z docs)
- Test coverage: Z% overall (breakdown by package)
- TODO count: X critical, Y high, Z medium (detailed breakdown)
- Known bugs: X critical, Y high, Z medium

**Dependency Analysis**:

- External dependencies: List with versions
- Internal dependencies: Package dependency graph
- Circular dependencies: Identified and resolution plan
- Coupling analysis: Tight/loose coupling assessment

**Technical Debt Assessment**:

- Architecture debt: Areas not following design patterns
- Code quality debt: Linting/formatting issues
- Test debt: Missing test coverage areas
- Documentation debt: Undocumented or outdated areas

### Stakeholder Analysis

**Primary Stakeholders**:

- Stakeholder 1: Role, expectations, success criteria
- Stakeholder 2: Role, expectations, success criteria

**Secondary Stakeholders**:

- Stakeholder N: Role, expectations, success criteria

**User Impact**:

- User Group 1: Impact description, mitigation plan
- User Group 2: Impact description, mitigation plan

---

## Architecture and Design

### System Architecture

**High-Level Architecture Diagram**:

```
┌─────────────────────────────────────────────────┐
│                                                 │
│  [Architecture diagram in ASCII or reference]  │
│                                                 │
└─────────────────────────────────────────────────┘
```

**Component Breakdown**:

- Component 1: Responsibility, interfaces, dependencies
- Component 2: Responsibility, interfaces, dependencies
- Component N: Responsibility, interfaces, dependencies

**Data Flow**:

```
Client → API Gateway → Service A → Database
                    ↓
                Service B → Cache
```

**Technology Stack**:

- Language/Runtime: Version requirements, constraints
- Frameworks: Version requirements, rationale
- Databases: Type, version, usage patterns
- External Services: APIs, integrations, fallback strategies

### Design Patterns

**Pattern 1**: `<PATTERN_NAME>`

- **Use Case**: When/where this pattern applies
- **Implementation**: How it's implemented in codebase
- **Benefits**: Why this pattern was chosen
- **Trade-offs**: What compromises were made

**Pattern 2**: `<PATTERN_NAME>`

- **Use Case**: When/where this pattern applies
- **Implementation**: How it's implemented in codebase
- **Benefits**: Why this pattern was chosen
- **Trade-offs**: What compromises were made

### Directory Structure

**Target Directory Layout**:

```
project-root/
├── cmd/                          # CLI entry points
│   ├── service-a/               # Service A CLI
│   └── service-b/               # Service B CLI
├── internal/                     # Private application code
│   ├── service-a/               # Service A domain
│   │   ├── server/             # HTTP handlers
│   │   ├── businesslogic/      # Business logic
│   │   ├── repository/         # Data access
│   │   └── config/             # Configuration
│   └── common/                  # Shared utilities
├── pkg/                          # Public library code
│   └── shared/                  # Reusable packages
├── api/                          # OpenAPI specs
│   └── service-a/               # Service A API
├── configs/                      # Configuration templates
└── docs/                         # Documentation
```

**Migration Path** (if refactoring):

| Source | Destination | Rationale | Risk Level |
|--------|-------------|-----------|------------|
| `old/path/` | `new/path/` | Why this move | LOW/MEDIUM/HIGH |

### API Design

**Endpoint Structure**:

- `GET /api/v1/resource` - List resources
- `POST /api/v1/resource` - Create resource
- `GET /api/v1/resource/{id}` - Get resource
- `PUT /api/v1/resource/{id}` - Update resource
- `DELETE /api/v1/resource/{id}` - Delete resource

**OpenAPI Specification**:

- Location: `api/service-name/openapi.yaml`
- Generation: `oapi-codegen` configuration
- Validation: Request/response validation patterns

**Authentication/Authorization**:

- Auth mechanism: OAuth 2.1, mTLS, API keys, etc.
- Authorization model: RBAC, ABAC, ACL
- Token format: JWT, opaque tokens, etc.

### Database Schema

**Entity Relationship Diagram**:

```
┌─────────────┐       ┌─────────────┐
│    User     │───────│   Session   │
├─────────────┤       ├─────────────┤
│ id (PK)     │       │ id (PK)     │
│ username    │       │ user_id(FK) │
│ email       │       │ expires_at  │
└─────────────┘       └─────────────┘
```

**Schema Evolution Strategy**:

- Migration tool: golang-migrate, GORM AutoMigrate, etc.
- Versioning: Sequential numbering (0001, 0002, etc.)
- Rollback: Up/down migration scripts
- Zero-downtime: Expand-contract pattern for schema changes

**Cross-Database Compatibility** (if applicable):

- PostgreSQL: Production database
- SQLite: Development/testing database
- Type mapping: UUID as TEXT, JSON as TEXT with serializer, etc.
- Connection pool: MaxOpenConns settings per database type

### Security Design

**Threat Model**:

- Threat 1: Description, impact, mitigation
- Threat 2: Description, impact, mitigation

**Security Controls**:

- Control 1: Implementation, verification
- Control 2: Implementation, verification

**Compliance Requirements**:

- FIPS 140-3: Approved algorithms only
- GDPR: Data protection measures
- SOC 2: Access controls, audit logging

---

## Implementation Tasks

### Task Organization

**Task Numbering Convention**:

- Primary tasks: `01-<TASK_NAME>.md` through `##-<TASK_NAME>.md`
- Sub-tasks: `##.##-<TASK_NAME>.md` (e.g., `10.5-login-ui.md`)
- Parallel tracks: Use prefixes (e.g., `R01-`, `R02-` for remediation tasks)

**Task Categories**:

- **Foundation**: Core infrastructure and models (Tasks 01-05)
- **Core Features**: Primary functionality (Tasks 06-10)
- **Advanced Features**: Enhanced capabilities (Tasks 11-15)
- **Integration**: Testing and validation (Tasks 16-18)
- **Documentation**: Handoff and finalization (Tasks 19-20)

### Implementation Tasks Table

| Task | File | Status | Priority | Effort | Dependencies | Risk | Description |
|------|------|--------|----------|--------|--------------|------|-------------|
| 01 | `01-foundation.md` | ✅/⚠️/❌ | 🔴 CRITICAL | 2 days | None | LOW | Foundation setup |
| 02 | `02-storage.md` | ✅/⚠️/❌ | 🔴 CRITICAL | 1 day | 01 | MEDIUM | Database abstractions |
| 03 | `03-core-logic.md` | ✅/⚠️/❌ | ⚠️ HIGH | 3 days | 01, 02 | MEDIUM | Core business logic |
| 04 | `04-api-layer.md` | ✅/⚠️/❌ | ⚠️ HIGH | 2 days | 03 | LOW | HTTP APIs |
| 05 | `05-integration.md` | ✅/⚠️/❌ | 🟡 MEDIUM | 1 day | 04 | LOW | Integration tests |
| 10.5 | `10.5-sub-feature.md` | ✅/⚠️/❌ | 🔴 CRITICAL | 4 hours | 10 | HIGH | Sub-task for task 10 |
| R01 | `R01-remediation.md` | ✅/⚠️/❌ | 🔴 CRITICAL | 1 day | None | HIGH | Remediation task |

**Status Legend**:

- ✅ **COMPLETE** - Fully implemented, tested, documented, verified
- ⚠️ **PARTIAL** - Documented as complete but has implementation gaps
- 🔄 **IN PROGRESS** - Currently being worked on
- 📋 **PLANNED** - Not yet started
- ❌ **BLOCKED** - Cannot proceed due to dependencies or issues
- 🗄️ **ARCHIVED** - Obsolete or superseded

**Priority Legend**:

- 🔴 **CRITICAL** - Production blocker, must complete for basic functionality
- ⚠️ **HIGH** - Important feature or significant impact
- 🟡 **MEDIUM** - Nice-to-have, moderate impact
- 🟢 **LOW** - Optional enhancement, minimal impact

**Risk Legend**:

- **CRITICAL** - High probability of failure with severe impact
- **HIGH** - Moderate probability of failure with significant impact
- **MEDIUM** - Low probability with moderate impact or moderate probability with low impact
- **LOW** - Low probability with minimal impact

### Task Dependencies Graph

**Critical Path**:

```
01 → 02 → 03 → 04 → 05 (Primary sequential path)
         ↓
      10.5 (Parallel sub-task)
```

**Parallel Execution Opportunities**:

- Tasks X, Y, Z can run in parallel (no shared dependencies)
- Week 1: Tasks 01-03 (sequential)
- Week 2: Tasks 04, 05, R01 (parallel streams)

### Implementation Phases

**Phase 1: Foundation** (Days 1-X)

- Focus: Core infrastructure, domain models, data access
- Tasks: 01, 02
- Deliverables: Working database, domain models, repositories
- Exit Criteria: All tests passing, 90%+ coverage

**Phase 2: Core Features** (Days X-Y)

- Focus: Primary business logic, API endpoints
- Tasks: 03, 04
- Deliverables: Functional APIs, core workflows
- Exit Criteria: E2E tests passing, API documented

**Phase 3: Advanced Features** (Days Y-Z)

- Focus: Enhanced capabilities, optimizations
- Tasks: 05, 06, etc.
- Deliverables: Advanced features, performance optimizations
- Exit Criteria: Performance benchmarks met, security audited

**Phase 4: Integration & Testing** (Days Z-W)

- Focus: System integration, comprehensive testing
- Tasks: Testing tasks
- Deliverables: E2E test suites, integration verified
- Exit Criteria: 95%+ coverage, load tests passing

**Phase 5: Documentation & Handoff** (Days W-V)

- Focus: Documentation, knowledge transfer
- Tasks: Documentation tasks
- Deliverables: Complete documentation, runbooks
- Exit Criteria: Documentation review approved

---

## Task Execution Instructions

### LLM Agent Continuous Work Directives

**CRITICAL: PRIMARY RULE - NEVER STOP UNTIL ALL TASKS COMPLETE**

This directive ensures autonomous LLM Agent sessions continue until all work is finished:

#### 1. Continuous Execution Pattern

**ALWAYS follow this loop:**

```
START → Task 01 → Commit → Task 02 → Commit → ... → Task ## → Commit → END
```

**NEVER introduce delays between tasks:**

- ❌ WRONG: commit → summary message → stop → wait for user
- ✅ RIGHT: commit → IMMEDIATELY invoke next tool (manage_todo_list, read_file, create_file)

**Token Budget Management:**

- Work until 950k tokens used (95% of 1M budget)
- Check `<system_warning>` after each tool call: "Token usage: X/1000000; Y remaining"
- STOP only when: tokens used ≥950k OR explicit user instruction to stop
- "All tasks complete" NEVER means stop - check docs for additional work

#### 2. Tool Call Chaining

**ZERO TEXT between tool calls:**

```
✅ CORRECT:
commit → manage_todo_list (mark task complete) → read_file (next task doc) → create_file (implement)

❌ WRONG:
commit → "Now working on Task 2..." → read_file (violates zero-text rule)
```

**No announcements, no status updates, no summaries** - just tool calls until work complete.

#### 3. Task Progression Strategy

**Sequential Task Execution:**

1. Read task doc: `read_file` on `##-<TASK>.md`
2. Implement: `create_file`, `replace_string_in_file`, `multi_replace_string_in_file`
3. Test: `runTests` tool (NEVER use `go test` in terminal)
4. Commit: `run_in_terminal` with conventional commit message
5. IMMEDIATELY mark task complete: `manage_todo_list` (update status)
6. IMMEDIATELY start next task: `read_file` on next task doc

**Parallel Task Execution** (when applicable):

- Identify independent tasks (no shared file dependencies)
- Execute file reads in parallel batches
- Implement changes in parallel batches
- Commit separately for each logical unit

#### 4. Commit Strategy

**Commit frequently, push when ready:**

- Commit after each task completion (atomic units of work)
- Use conventional commit format: `type(scope): description`
- Push only when passing all quality gates (tests, linting, pre-commit)

**Skip pre-commit for speed during iteration:**

```bash
git commit --no-verify -m "feat(task-01): implement foundation models"
```

**Always run quality checks before push:**

```bash
go test ./... -cover
golangci-lint run ./...
pre-commit run --all-files  # Before git push
```

#### 5. Quality Gates

**ALWAYS enforce before marking task complete:**

- ✅ All tests passing: `runTests` tool shows 0 failures
- ✅ Coverage threshold: ≥95% for infrastructure, ≥90% for features, ≥100% for utilities
- ✅ No linting errors: `golangci-lint run ./...` shows 0 issues
- ✅ No TODO comments: Grep search shows 0 TODOs in modified files
- ✅ Documentation updated: README, inline docs, API specs

**Progressive quality approach:**

- During iteration: Use `git commit --no-verify` for speed
- Before task complete: Run full quality checks
- Before push: Run pre-commit hooks

#### 6. Error Handling

**When tests fail:**

1. Analyze failure output
2. Fix implementation
3. Re-run tests
4. Repeat until passing
5. NEVER skip failing tests

**When linting fails:**

1. Run `golangci-lint run --fix` first (auto-fixes formatting, imports, etc.)
2. Manual fix remaining issues
3. Re-run linter
4. Repeat until clean

**When blocked:**

1. Document blocker in task doc post-mortem section
2. Create new task doc for blocker resolution
3. Continue with other independent tasks
4. Return to blocked task after resolution

#### 7. Task Documentation Maintenance

**ALWAYS update task status immediately:**

```
manage_todo_list → update task status to "completed"
```

**ALWAYS create post-mortem after task completion:**

```
create_file → ##-<TASK>-POSTMORTEM.md with corrective actions
```

**ALWAYS check for new tasks:**

```
After clearing todo list, check docs/##-*/todos-*.md for additional work
```

### Task Execution Checklist

Each task MUST complete this checklist before marking complete:

#### Pre-Implementation

- [ ] Read task documentation thoroughly
- [ ] Understand acceptance criteria
- [ ] Identify dependencies (prior tasks, external packages)
- [ ] Review related code for patterns and conventions
- [ ] Check for existing implementations to reuse

#### Implementation

- [ ] Create/modify files according to task spec
- [ ] Follow project coding standards and patterns
- [ ] Add comprehensive inline documentation
- [ ] Use descriptive variable/function names
- [ ] Handle error cases explicitly

#### Testing

- [ ] Write table-driven tests for all functionality
- [ ] Test happy paths (expected inputs/outputs)
- [ ] Test sad paths (error conditions, edge cases)
- [ ] Use `t.Parallel()` for all tests (validates concurrent safety)
- [ ] Achieve target coverage (≥95% infrastructure, ≥90% features, ≥100% utilities)
- [ ] Run tests: `runTests` tool (NEVER `go test` in terminal)

#### Quality Assurance

- [ ] Run auto-fix: `golangci-lint run --fix`
- [ ] Fix remaining lint issues manually
- [ ] Verify no TODO comments introduced
- [ ] Check import aliases follow `.golangci.yml` conventions
- [ ] Validate magic values in `magic*.go` files (not inline)

#### Documentation

- [ ] Update inline code comments (godoc format)
- [ ] Update README if user-facing changes
- [ ] Update OpenAPI specs if API changes
- [ ] Document breaking changes in CHANGELOG
- [ ] Update architecture diagrams if structural changes

#### Commit

- [ ] Stage files: `git add <files>`
- [ ] Commit with conventional message: `git commit -m "type(scope): description"`
- [ ] Use `--no-verify` during iteration for speed
- [ ] Verify commit successful

#### Post-Mortem

- [ ] Create `##-<TASK>-POSTMORTEM.md`
- [ ] Document bugs encountered and fixes
- [ ] Document omissions discovered
- [ ] Document suboptimal implementation patterns
- [ ] Document test failures and resolutions
- [ ] List corrective actions for future tasks
- [ ] Identify instruction violations and corrections

#### Progressive Validation (MANDATORY)

- [ ] Run 6-step validation: `go run ./cmd/cicd identity-progressive-validation`
- [ ] Step 1 PASS: TODO scan (0 CRITICAL/HIGH TODOs)
- [ ] Step 2 PASS: Tests (100% pass rate)
- [ ] Step 3 PASS: Coverage (≥95% infrastructure, ≥90% features, ≥100% utilities)
- [ ] Step 4 PASS: Requirements (≥90% overall coverage)
- [ ] Step 5 PASS: Integration (E2E smoke test)
- [ ] Step 6 PASS: Documentation (PROJECT-STATUS.md <7 days old)
- [ ] All 6 steps passed before marking task complete

**If validation fails:**

- Fix issues immediately
- Re-run progressive validation
- Do NOT mark task complete until 6/6 steps pass

#### Handoff

- [ ] Mark task complete in `manage_todo_list`
- [ ] IMMEDIATELY start next task (no stopping, no summary)

### Testing Guidelines

**Test File Organization:**

- Unit tests: `*_test.go` (same package)
- Benchmarks: `*_bench_test.go` (performance testing)
- Fuzz tests: `*_fuzz_test.go` (property-based testing)
- Integration: `*_integration_test.go` (component interaction)
- E2E: `e2e_test.go` with `//go:build e2e` (full system)

**Test Patterns:**

```go
func TestFunction(t *testing.T) {
    t.Parallel() // ALWAYS enable parallel testing

    tests := []struct {
        name           string
        input          InputType
        wantOutput     OutputType
        wantError      bool
        wantContains   string
    }{
        {
            name: "happy path - valid input",
            input: validInput,
            wantOutput: expectedOutput,
            wantError: false,
        },
        {
            name: "sad path - invalid input",
            input: invalidInput,
            wantError: true,
            wantContains: "expected error message",
        },
    }

    for _, tc := range tests {
        tc := tc // Capture range variable
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel() // ALWAYS enable parallel subtests

            got, err := FunctionUnderTest(tc.input)

            if tc.wantError {
                require.Error(t, err)
                if tc.wantContains != "" {
                    require.Contains(t, err.Error(), tc.wantContains)
                }
            } else {
                require.NoError(t, err)
                require.Equal(t, tc.wantOutput, got)
            }
        })
    }
}
```

**Test Data Isolation:**

- Use `t.TempDir()` for file-based tests
- Use UUIDv7 for unique test data (time-ordered, no collisions)
- NEVER use sequential counters or timestamps for uniqueness
- Design tests for complete isolation (parallel-safe)

---

## Post-Mortem and Corrective Actions

### Post-Mortem Structure

**CRITICAL: EVERY task MUST have a post-mortem document**

**File Naming**: `##-<TASK>-POSTMORTEM.md` (e.g., `01-foundation-POSTMORTEM.md`)

**Template**:

```markdown
# Task ##: <TASK_NAME> - Post-Mortem

**Task File**: `##-<TASK>.md`
**Completion Date**: YYYY-MM-DD
**Total Time**: X hours (estimated: Y hours, variance: Z%)
**Overall Assessment**: ✅ SUCCESS | ⚠️ PARTIAL | ❌ BLOCKED

---

## Implementation Summary

### What Was Implemented

- Component 1: Brief description, file locations
- Component 2: Brief description, file locations
- Component N: Brief description, file locations

### What Was Deferred

- Deferred Item 1: Reason, target task for completion
- Deferred Item 2: Reason, target task for completion

### Unexpected Scope Additions

- Addition 1: Description, justification, impact
- Addition 2: Description, justification, impact

---

## Issues Encountered

### Bugs Discovered

| Bug ID | Description | Root Cause | Fix | Files Affected |
|--------|-------------|------------|-----|----------------|
| B01 | Concurrency issue in repo | Missing mutex | Added sync.RWMutex | `repository.go` |
| B02 | Test flakiness | Shared state | Isolated test data | `*_test.go` |

### Omissions Identified

| Omission ID | Description | Impact | Resolution |
|-------------|-------------|--------|------------|
| O01 | Missing error wrapping | Poor debugging | Added context to errors |
| O02 | No input validation | Security risk | Added validation layer |

### Suboptimal Implementation Patterns

| Pattern ID | Description | Why Suboptimal | Better Approach |
|------------|-------------|----------------|-----------------|
| P01 | Manual string building | Performance, errors | Use strings.Builder |
| P02 | Nested if statements | Readability | Extract guard clauses |

### Test Failures

| Test ID | Description | Root Cause | Resolution | Prevention |
|---------|-------------|------------|------------|------------|
| T01 | Race condition | Shared global | Use t.Parallel() | Always parallel tests |
| T02 | Database lock | MaxOpenConns=1 | Set MaxOpenConns=5 | GORM config review |

### Instruction Violations

| Violation ID | Instruction Violated | Description | Correction | Prevention |
|--------------|---------------------|-------------|------------|------------|
| V01 | "ALWAYS use table-driven tests" | Used separate test functions | Refactored to table-driven | Code review checklist |
| V02 | "NEVER use interface{}" | Used interface{} instead of any | Replaced with any | golangci-lint importas |

---

## Corrective Actions

### Immediate Corrective Actions (Current Task)

1. **Action 1**: Description of fix applied within this task
   - **Root Cause**: Why issue occurred
   - **Fix Applied**: What was done to resolve
   - **Verification**: How fix was validated
   - **Files**: Affected files

2. **Action 2**: Description of fix applied within this task
   - **Root Cause**: Why issue occurred
   - **Fix Applied**: What was done to resolve
   - **Verification**: How fix was validated
   - **Files**: Affected files

### Deferred Corrective Actions (Future Tasks)

#### New Task Documents to Create

| Task ID | Task Name | Priority | Rationale | Dependencies |
|---------|-----------|----------|-----------|--------------|
| ##.1 | Sub-task for gap | HIGH | Fix critical omission | Current task |
| ##+1 | Follow-up task | MEDIUM | Optimization opportunity | Current task |

**Template for new task docs**:
```bash
# Create new task doc
create_file → ##.##-<SUBTASK>.md

# Append to existing task sequence
create_file → ##-<NEXT_TASK>.md
```

#### Existing Task Documents to Update

| Task ID | Update Type | Description | Priority |
|---------|-------------|-------------|----------|
| ## | Add subtask | New subtask for X | HIGH |
| ##+1 | Modify acceptance criteria | Add validation for Y | MEDIUM |

**Update patterns**:

```bash
# Add subtask to existing task doc
replace_string_in_file → ##-<TASK>.md (add subtask section)

# Modify upcoming task acceptance criteria
replace_string_in_file → ##-<TASK>.md (update criteria)
```

#### Pattern Improvements

| Pattern ID | Current Pattern | Improved Pattern | Impact | Affected Tasks |
|------------|----------------|------------------|--------|----------------|
| PI01 | Manual validation | Validation middleware | Reusability | Tasks X, Y, Z |
| PI02 | String concatenation | StringBuilder pattern | Performance | All tasks |

---

## Lessons Learned

### What Went Well

1. **Success 1**: Description of what worked effectively
   - **Why It Worked**: Root cause of success
   - **Replication**: How to repeat in future tasks
   - **Documentation**: Where pattern is documented

2. **Success 2**: Description of what worked effectively
   - **Why It Worked**: Root cause of success
   - **Replication**: How to repeat in future tasks
   - **Documentation**: Where pattern is documented

### What Needs Improvement

1. **Improvement 1**: Description of what could be better
   - **Current Approach**: What we did
   - **Better Approach**: What we should do next time
   - **Action Items**: Specific steps to improve

2. **Improvement 2**: Description of what could be better
   - **Current Approach**: What we did
   - **Better Approach**: What we should do next time
   - **Action Items**: Specific steps to improve

### Process Improvements

| Process | Current | Proposed | Benefit | Effort |
|---------|---------|----------|---------|--------|
| Testing | Manual test writing | Test generation from specs | Faster, complete | LOW |
| Documentation | Post-implementation | Inline during implementation | Better context | MEDIUM |

---

## Metrics

### Time Metrics

- **Estimated Effort**: X hours
- **Actual Effort**: Y hours
- **Variance**: Z% (over/under estimate)
- **Breakdown**:
  - Implementation: X hours (Y%)
  - Testing: X hours (Y%)
  - Debugging: X hours (Y%)
  - Documentation: X hours (Y%)

### Quality Metrics

- **Code Coverage**: X% (target: Y%)
- **Linting Issues**: X (target: 0)
- **TODO Comments**: X (target: 0)
- **Test Pass Rate**: X% (target: 100%)
- **Bug Density**: X bugs per 100 LOC

### Complexity Metrics

- **Cyclomatic Complexity**: Average X (max acceptable: Y)
- **File Size**: Average X lines (max acceptable: Y)
- **Function Size**: Average X lines (max acceptable: Y)

---

## Risk Updates

### Risks Realized

| Risk ID | Description | Impact | Mitigation Applied | Outcome |
|---------|-------------|--------|-------------------|---------|
| R01 | Database migration failure | HIGH | Rollback script | Resolved |
| R02 | Breaking API changes | MEDIUM | Versioning | Mitigated |

### New Risks Identified

| Risk ID | Description | Probability | Impact | Mitigation Plan | Owner |
|---------|-------------|-------------|--------|----------------|-------|
| NR01 | Concurrency bottleneck | MEDIUM | HIGH | Connection pooling | Task ##+1 |
| NR02 | Test flakiness | LOW | MEDIUM | Isolate test data | Task ##+2 |

---

## Corrective Action Summary

**Summary checklist for quick reference:**

- [ ] Created X new task documents for gaps
- [ ] Updated Y existing task documents
- [ ] Applied Z immediate fixes in current task
- [ ] Identified N pattern improvements for future tasks
- [ ] Updated risk register with M new risks
- [ ] Documented P lessons learned for knowledge base

**Next Task Modifications**:

```
Task ##+1:
- Add subtask: X
- Update acceptance criteria: Y
- Include pattern improvement: Z

Task ##+2:
- New task for: A
- Priority: HIGH (addresses critical gap from this task)
```

```

### Real-World SDLC Corrective Action Patterns

Based on industry best practices, here are additional corrective action strategies:

#### 1. Retrospective Analysis Patterns

**Sprint Retrospective**:
- What worked well (continue doing)
- What didn't work (stop doing)
- What to try (start doing)
- Action items with owners and deadlines

**Root Cause Analysis (5 Whys)**:
- Ask "why" 5 times to get to root cause
- Document each level of causality
- Identify systemic vs. one-off issues

**Incident Post-Mortem**:
- Timeline of events
- Impact assessment (users affected, downtime, data loss)
- Root cause analysis
- Preventative measures
- Action items with owners

#### 2. Technical Debt Management

**Debt Tracking**:
- Create technical debt backlog items
- Prioritize: HIGH (pay down next sprint), MEDIUM (next quarter), LOW (opportunistic)
- Track debt accumulation rate
- Schedule dedicated debt reduction sprints

**Refactoring Opportunities**:
- Extract duplicated code to shared utilities
- Improve test coverage in low-coverage areas
- Simplify complex functions (cyclomatic complexity > 10)
- Update outdated dependencies

#### 3. Knowledge Management

**Documentation Updates**:
- Update architecture decision records (ADRs)
- Create runbooks for operational procedures
- Document gotchas and edge cases
- Update onboarding guides with new patterns

**Knowledge Sharing**:
- Tech talks on new patterns discovered
- Code review learnings shared with team
- Update coding standards based on learnings
- Create reusable code templates

#### 4. Process Improvements

**Automation Opportunities**:
- Automate repetitive manual steps
- Add pre-commit hooks for common mistakes
- Create CI/CD pipeline improvements
- Add monitoring/alerting for issues discovered

**Quality Gate Refinements**:
- Add new linter rules for patterns to avoid
- Update test coverage thresholds
- Add performance benchmarking requirements
- Enhance security scanning rules

#### 5. Risk Mitigation

**Risk Register Updates**:
- Add newly discovered risks
- Update probability/impact assessments
- Define mitigation strategies
- Assign risk owners

**Dependency Management**:
- Identify new external dependencies
- Assess dependency health (maintenance, security)
- Plan for dependency upgrades
- Identify alternatives for risky dependencies

#### 6. Metrics and Monitoring

**Performance Baselines**:
- Establish performance benchmarks
- Set up performance regression testing
- Define SLOs/SLAs for new features
- Add monitoring/alerting for degradation

**Quality Metrics**:
- Track defect density over time
- Monitor test coverage trends
- Measure code complexity trends
- Track technical debt accumulation

---

## Quality Gates and Acceptance Criteria

### Universal Acceptance Criteria

**EVERY task MUST meet these criteria before marking complete:**

#### Code Quality
- [ ] All files compile without errors
- [ ] Zero linting errors: `golangci-lint run ./...`
- [ ] Code follows project conventions (naming, structure, patterns)
- [ ] No hardcoded values (use magic constants in `magic*.go`)
- [ ] Proper error handling with context wrapping
- [ ] No TODO comments introduced (or tracked in issues)

#### Testing
- [ ] Table-driven tests for all functionality
- [ ] Happy path coverage (expected inputs/outputs)
- [ ] Sad path coverage (error conditions, edge cases)
- [ ] All tests pass: `runTests` tool shows 0 failures
- [ ] Coverage meets threshold (≥95% infra, ≥90% features, ≥100% utils)
- [ ] Tests use `t.Parallel()` (validates concurrent safety)
- [ ] Integration tests verify component interactions

#### Documentation
- [ ] Inline godoc comments for all exported symbols
- [ ] README updated if user-facing changes
- [ ] API documentation updated (OpenAPI specs)
- [ ] Architecture diagrams updated if structural changes
- [ ] Migration guide created if breaking changes

#### Single Source of Truth (Single Source Of Truth)

**MANDATORY: Maintain PROJECT-STATUS.md as authoritative status document**

**Purpose**: Prevent contradictory documentation (README claiming 100%, STATUS-REPORT showing 45%)

**Single Source Of Truth Structure** (`docs/[feature-name]/PROJECT-STATUS.md`):
```markdown
# [Feature Name] - Project Status

**Last Updated**: YYYY-MM-DD HH:MM UTC
**Status**: 🔴 NOT READY | 🟡 IN PROGRESS | 🟢 PRODUCTION READY

## Current Completion Metrics
- **Original Plan**: X% complete (Y/Z tasks)
- **Remediation Plan**: X% complete (Y/Z tasks)
- **Requirements Coverage**: X% (Y/Z requirements validated)
- **Test Coverage**: X% (≥95% target)
- **TODO Count**: X total (Y CRITICAL, Z HIGH, A MEDIUM, B LOW)

## Production Blockers
1. [Blocker 1]: Description, impact, ETA
2. [Blocker 2]: Description, impact, ETA

## Known Limitations
- Limitation 1: Description, workaround, fix plan
- Limitation 2: Description, workaround, fix plan

## Next Steps
1. [Next immediate task]
2. [Following task]
3. [Final verification task]
```

**Update Triggers** (MANDATORY - CI/CD enforced):

- **After EVERY task completion**: Update completion metrics, blockers, next steps
- **After EVERY TODO resolution**: Update TODO count
- **After EVERY test run**: Update test coverage if changed
- **Minimum frequency**: Weekly (even if no changes)

**Enforcement**:

```bash
# CI/CD check: Fail if PROJECT-STATUS.md not updated in >7 days
LAST_UPDATE=$(grep "Last Updated" docs/[feature]/PROJECT-STATUS.md | cut -d: -f2-)
DAYS_OLD=$(( ($(date +%s) - $(date -d "$LAST_UPDATE" +%s)) / 86400 ))
if [ $DAYS_OLD -gt 7 ]; then
  echo "ERROR: PROJECT-STATUS.md not updated in $DAYS_OLD days (>7 day limit)"
  exit 1
fi
```

**Benefits**:

- ✅ Single authoritative source for status queries
- ✅ Prevents contradictory claims across multiple docs
- ✅ Forces regular status assessment
- ✅ Easy to verify actual vs claimed progress

#### Architecture Compliance

- [ ] Follows project directory structure
- [ ] Respects domain boundaries (no prohibited imports)
- [ ] Import aliases follow `.golangci.yml` conventions
- [ ] Uses approved design patterns
- [ ] Maintains loose coupling

### Security

- [ ] No secrets in code (use configuration/secrets management)
- [ ] Input validation for all external inputs
- [ ] Output encoding to prevent injection attacks
- [ ] Cryptographic operations use approved algorithms (FIPS 140-3)
- [ ] Authentication/authorization properly implemented

### Performance

- [ ] No obvious performance bottlenecks
- [ ] Database queries optimized (indexes, pagination)
- [ ] Connection pooling configured appropriately
- [ ] Resource cleanup (defer close, context cancellation)
- [ ] Benchmarks created for critical paths

### Task-Specific Acceptance Criteria

**CRITICAL ENFORCEMENT**: Every acceptance criterion MUST include "Evidence Required" subsection with specific validation commands and expected outputs.

**Pattern**: NO vague criteria like "feature functional" - ONLY objective, verifiable outcomes with command-line evidence

**Why Evidence Required**:

- Makes completion objective and verifiable (not subjective "looks done")
- Prevents agent claiming completion without running validation commands
- Enables automated quality gate enforcement
- Provides audit trail for compliance/review

**Template for task-specific criteria:**

```markdown
## ✅ ACCEPTANCE CRITERIA

### Functional Requirements
- [ ] Requirement 1: Specific, measurable, testable outcome
  - **Evidence Required**:
    - [ ] Test result: `runTests ./path/to/package` shows TestRequirement1 passes (exit code 0)
    - [ ] TODO scan: `grep -r "TODO\|FIXME" <modified_files>` shows empty output (zero TODOs)
    - [ ] Requirements validation: `identity-requirements-check --strict` shows R01-01 through R01-05 verified
    - [ ] Code coverage: `go test -cover ./path/to/package` shows ≥90% coverage
- [ ] Requirement 2: Specific, measurable, testable outcome
  - **Evidence Required**:
    - [ ] Integration test: `runTests ./path/to/integration` shows TestRequirement2Integration passes
    - [ ] Manual verification: Documented curl/API test succeeds (terminal output captured)
    - [ ] Regression check: Existing tests still pass (runTests shows no new failures)
- [ ] Requirement N: Specific, measurable, testable outcome
  - **Evidence Required**:
    - [ ] Build output: `go build ./...` succeeds with exit code 0
    - [ ] Lint output: `golangci-lint run ./...` shows zero errors (exit code 0)
    - [ ] Documentation: README updated, OpenAPI synced (if applicable), godoc reviewed
    - [ ] Import compliance: `golangci-lint run --enable-only=importas ./...` shows zero errors

### Non-Functional Requirements
- [ ] Performance: Response time < Xms for Y% of requests
- [ ] Scalability: Supports Z concurrent users
- [ ] Reliability: Uptime > X%
- [ ] Security: Passes vulnerability scan with 0 critical/high issues

### Integration Requirements
- [ ] Integrates with System A via API X
- [ ] Data flows correctly to System B
- [ ] Error handling propagates appropriately

### User Experience Requirements
- [ ] API response times meet targets
- [ ] Error messages are clear and actionable
- [ ] Documentation enables self-service usage

### Operational Requirements
- [ ] Monitoring/alerting configured
- [ ] Logging captures key events
- [ ] Runbook created for common scenarios
- [ ] Rollback procedure documented and tested
```

### Automated Quality Gates

#### CRITICAL Requirements

Run these commands before marking task complete:

#### Code Quality

```bash
# Build verification (must succeed)
go build ./...

# Linting (must show zero errors)
golangci-lint run ./...

# TODO scan (must show zero CRITICAL/HIGH TODOs)
go run ./cmd/cicd go-identity-todo-scan --fail-on=critical,high

# Circular dependency check (must pass)
go run ./cmd/cicd go-check-circular-package-dependencies
```

#### Testing

```bash
# Unit tests (must show 100% pass rate)
runTests

# Coverage verification (must meet thresholds: ≥95% infra, ≥90% features)
go test ./internal/identity/... -coverprofile=test-output/coverage_identity.out
go tool cover -func=test-output/coverage_identity.out | grep total

# Integration tests (if applicable)
runTests files=["path/to/integration_test.go"]
```

#### Requirements Validation

```bash
# Requirements coverage check (must meet ≥90% per-task, ≥90% overall)
go run ./cmd/cicd go-identity-requirements-check --strict
```

#### Documentation Validation

```bash
# README consistency check
# Verify all task changes documented in appropriate README files

# OpenAPI synchronization (if API changes)
go run ./api/generate.go
git diff api/ # Must show no unexpected changes
```

### Quality Gate Enforcement

#### Pre-Commit Gate

Local development automated by pre-commit hooks:

```bash
# Automated by pre-commit hooks
- UTF-8 encoding validation
- File size limits enforcement
- Test pattern enforcement
- golangci-lint run --fix
- go test ./... -cover
```

#### Pre-Push Gate

Before git push - automated by pre-push hooks:

```bash
# Automated by pre-push hooks
- All pre-commit checks
- Integration tests passing
- Coverage thresholds met
- No circular dependencies
- Import alias validation
```

#### PR Merge Gate

CI/CD automated by GitHub Actions:

```bash
# Automated by GitHub Actions
- All tests passing (unit, integration, e2e)
- Code coverage > threshold
- Security scans clean (SAST, DAST)
- Dependency vulnerability scans clean
- Documentation built successfully
- Performance benchmarks within limits
```

#### Production Deployment Gate

Automated by deployment pipeline:

```bash
# Automated by deployment pipeline
- All PR merge gates passed
- Load tests passing
- Smoke tests in staging passing
- Rollback tested and verified
- Monitoring configured and alerting
- Runbook reviewed and approved
```

### Requirements Coverage Threshold

#### Minimum Coverage Requirements

**Per-Task Threshold**:

- **Minimum**: ≥95% of requirements validated per task before marking complete
- **Enforcement**: `go run ./cmd/cicd go-identity-requirements-check --strict --task-threshold=90`
- **Rationale**: Prevents claiming completion while leaving gaps
- **Acceptance**: Task NOT complete until threshold met

**Overall Coverage Threshold**:

- **Minimum**: ≥90% of total requirements validated across all tasks
- **Enforcement**: `go run ./cmd/cicd go-identity-requirements-check --strict --overall-threshold=85`
- **Rationale**: Ensures comprehensive feature implementation
- **Acceptance**: Feature NOT production-ready until threshold met

**CI/CD Integration**:

```yaml
# .github/workflows/ci-quality.yml
- name: Validate Requirements Coverage
  run: |
    go run ./cmd/cicd go-identity-requirements-check --strict \
      --task-threshold=90 \
      --overall-threshold=85
```

**Acceptance Criteria Addition**:

- [ ] Per-task coverage ≥95% (run `go-identity-requirements-check --strict`)
- [ ] Overall coverage ≥90% (verified in PROJECT-STATUS.md)
- [ ] CI/CD enforces thresholds (workflow passes)

---

## Risk Management

### Risk Categories

**Technical Risks**:

- Architecture complexity
- Technology maturity
- Dependency risks
- Performance/scalability
- Security vulnerabilities

**Process Risks**:

- Schedule delays
- Scope creep
- Resource availability
- Knowledge gaps
- Communication breakdown

**Operational Risks**:

- Deployment failures
- Data migration issues
- Backward compatibility breaks
- Service degradation
- Incident response

### Risk Assessment Matrix

| Risk ID | Description | Probability | Impact | Severity | Mitigation | Owner | Status |
|---------|-------------|-------------|--------|----------|------------|-------|--------|
| R01 | Database migration failure | MEDIUM | HIGH | HIGH | Rollback script, staging test | Team A | ⚠️ ACTIVE |
| R02 | Breaking API changes | LOW | HIGH | MEDIUM | API versioning, deprecation period | Team B | 🟢 MITIGATED |
| R03 | Security vulnerability | LOW | CRITICAL | HIGH | Security review, penetration test | Team C | ⚠️ ACTIVE |

**Probability Legend**:

- **LOW**: < 25% chance of occurrence
- **MEDIUM**: 25-75% chance of occurrence
- **HIGH**: > 75% chance of occurrence

**Impact Legend**:

- **LOW**: Minor inconvenience, easy workaround
- **MEDIUM**: Significant disruption, impacts subset of users
- **HIGH**: Major disruption, impacts most users
- **CRITICAL**: System-wide failure, data loss, security breach

**Severity Calculation**:

- CRITICAL: High probability + Critical impact OR Medium probability + Critical impact
- HIGH: High probability + High impact OR Medium probability + High impact
- MEDIUM: Any combination not meeting HIGH or CRITICAL thresholds
- LOW: Low probability + Low/Medium impact

### Risk Mitigation Strategies

**Avoidance**: Eliminate the risk by changing approach

- Example: Use proven technology instead of experimental framework

**Reduction**: Minimize probability or impact

- Example: Add automated tests to reduce probability of bugs

**Transfer**: Shift risk to third party

- Example: Use managed service instead of self-hosting

**Acceptance**: Acknowledge risk and plan response

- Example: Document known limitation and support workaround

### Risk Monitoring

**Risk Reviews**:

- Daily: Critical risks during active sprints
- Weekly: High risks and new risk identification
- Monthly: Medium/low risks and trend analysis
- Quarterly: Risk register cleanup and archival

**Risk Escalation**:

- Critical severity: Immediate escalation to stakeholders
- High severity: Escalate within 24 hours
- Medium severity: Escalate within 1 week
- Low severity: Track in backlog, periodic review

---

## Success Metrics

### Completion Metrics

**Task Completion Rate**:

- Target: 100% of planned tasks complete
- Current: X% (Y/Z tasks)
- Trend: +/-X% from last period

**On-Time Delivery**:

- Target: 90% of tasks on schedule
- Current: X% on-time
- Variance: Average +/-X days

**Quality Metrics**:

- Zero CRITICAL bugs in production
- < 5% MEDIUM/LOW bugs per release
- 100% test pass rate
- ≥95% code coverage (infrastructure)

### Performance Metrics

**Response Time**:

- Target: P95 < 100ms, P99 < 500ms
- Current: P95 = Xms, P99 = Yms
- SLA: 99.9% uptime

**Throughput**:

- Target: X requests/second
- Current: Y requests/second
- Peak: Z requests/second sustained

**Resource Utilization**:

- CPU: Target < 70%, Current X%
- Memory: Target < 80%, Current Y%
- Disk: Target < 75%, Current Z%

### Business Metrics

**User Adoption**:

- Active users: Target X, Current Y
- New user growth: Target +X% MoM, Current +Y%
- User retention: Target >X%, Current Y%

**Feature Usage**:

- Feature A adoption: Target X%, Current Y%
- Feature B usage: Target X requests/day, Current Y
- Feature satisfaction: Target NPS >X, Current Y

**Operational Efficiency**:

- Deployment frequency: Target X/week, Current Y
- Mean time to recovery (MTTR): Target <X hours, Current Y
- Change failure rate: Target <X%, Current Y%

### Quality Metrics

**Defect Density**:

- Target: <0.5 bugs per 100 LOC
- Current: X bugs per 100 LOC
- Trend: Decreasing/stable/increasing

**Test Coverage**:

- Target: ≥95% for infrastructure, ≥90% for features
- Current: X% overall, Y% infrastructure, Z% features
- Trend: Increasing/stable/decreasing

**Technical Debt**:

- Target: <X hours per sprint on debt reduction
- Current: Y hours per sprint
- Trend: Decreasing/stable/increasing

**Code Quality**:

- Cyclomatic complexity: Target <10, Current X avg
- File size: Target <500 lines, Current Y avg
- Linting issues: Target 0, Current Z

---

## Appendix

### A. Terminology

**Acceptance Criteria**: Specific conditions that must be met for a task to be considered complete
**Blocker**: Issue preventing progress on a task or feature
**Corrective Action**: Steps taken to address issues discovered during implementation
**Post-Mortem**: Analysis document created after task completion documenting issues and learnings
**Quality Gate**: Checkpoint with specific criteria that must be met before proceeding
**Risk**: Potential issue that could impact successful completion
**Technical Debt**: Shortcuts or suboptimal implementations that will require future rework

### B. References

**Internal Documentation**:

- [Coding Instructions](.github/instructions/01-01.coding.instructions.md)
- [Testing Instructions](.github/instructions/01-02.testing.instructions.md)
- [Go Instructions](.github/instructions/01-03.golang.instructions.md)
- [Database Instructions](.github/instructions/01-04.database.instructions.md)
- [Security Instructions](.github/instructions/01-05.security.instructions.md)
- [Linting Instructions](.github/instructions/01-06.linting.instructions.md)

**External Standards**:

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [OpenAPI 3.0 Specification](https://spec.openapis.org/oas/v3.0.3)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)

### C. Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-11-23 | Initial | Template creation based on analysis of 50+ planning docs |

### D. Template Usage Guidelines

**When to Use This Template**:

- New feature development (multi-day/week effort)
- Major refactoring initiatives
- Cross-cutting infrastructure changes
- Service extraction or migration

**When NOT to Use This Template**:

- Bug fixes (use issue tracker)
- Minor enhancements (single file changes)
- Documentation updates (use README)
- Dependency updates (use automated tools)

**Customization Guidelines**:

- Remove sections not applicable to your feature
- Add domain-specific sections as needed
- Adjust task granularity based on complexity
- Scale acceptance criteria based on criticality

**Living Document**:

- Update template based on learnings from each feature
- Incorporate feedback from team retrospectives
- Version control template changes
- Document template evolution in version history
