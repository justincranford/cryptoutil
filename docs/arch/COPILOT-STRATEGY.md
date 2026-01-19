# Copilot Continuous Work Strategy

**Version**: 2.0.0
**Last Updated**: 2025-01-18
**Purpose**: Define continuous iteration loop for implementing features and fixing bugs until complete

---

## Core Principle: Continuous Iteration Loop

**NEVER stop until problem is completely solved and all tests pass**

```
┌─────────────────────────────────────────────┐
│ 1. Understand Problem                      │
│    - Read requirements/bug reports         │
│    - Read existing code and tests          │
│    - Identify affected files               │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 2. Make Changes                             │
│    - Implement feature OR fix bug           │
│    - Update tests                           │
│    - Update documentation                   │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 3. Run Unit Tests                           │
│    - go test ./... -cover                   │
│    - Coverage ≥95% production               │
│    - Coverage ≥98% infrastructure           │
└──────────────┬──────────────────────────────┘
               │
        ┌──────┴──────┐
        │ Tests Pass? │
        └──────┬──────┘
               │
        ┌──────┴──────┐
        │     NO      │
        └──────┬──────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 4. Debug Failures                           │
│    - Read test output                       │
│    - Identify root cause                    │
│    - Fix implementation OR tests            │
│    - Go back to step 2                      │
└──────────────┬──────────────────────────────┘
               │
        ┌──────┴──────┐
        │     YES     │
        └──────┬──────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 5. Run Linting                              │
│    - golangci-lint run --fix                │
│    - Fix all warnings                       │
└──────────────┬──────────────────────────────┘
               │
        ┌──────┴──────┐
        │  Clean?     │
        └──────┬──────┘
               │
        ┌──────┴──────┐
        │     NO      │
        └──────┬──────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 6. Fix Linting Issues                       │
│    - Apply fixes manually                   │
│    - Go back to step 2                      │
└──────────────┬──────────────────────────────┘
               │
        ┌──────┴──────┐
        │     YES     │
        └──────┬──────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 7. Run Integration Tests (if applicable)    │
│    - go test ./... -tags=integration        │
│    - Verify database operations             │
│    - Verify external service interactions   │
└──────────────┬──────────────────────────────┘
               │
        ┌──────┴──────┐
        │ Tests Pass? │
        └──────┬──────┘
               │
        ┌──────┴──────┐
        │     NO      │
        └──────┬──────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 8. Debug Integration Failures               │
│    - Check container logs                   │
│    - Verify test data isolation             │
│    - Fix integration code                   │
│    - Go back to step 2                      │
└──────────────┬──────────────────────────────┘
               │
        ┌──────┴──────┐
        │     YES     │
        └──────┬──────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 9. Run E2E Tests (if applicable)            │
│    - docker compose up -d                   │
│    - go test ./internal/.../e2e/... -v      │
│    - Test /service/** AND /browser/** paths │
└──────────────┬──────────────────────────────┘
               │
        ┌──────┴──────┐
        │ Tests Pass? │
        └──────┬──────┘
               │
        ┌──────┴──────┐
        │     NO      │
        └──────┬──────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 10. Debug E2E Failures                      │
│     - Check service logs                    │
│     - Verify configuration                  │
│     - Fix E2E code or services              │
│     - Go back to step 2                     │
└──────────────┬──────────────────────────────┘
               │
        ┌──────┴──────┐
        │     YES     │
        └──────┬──────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 11. Run Mutation Tests (if applicable)      │
│     - gremlins unleash ./internal/...       │
│     - Efficacy ≥85% production              │
│     - Efficacy ≥98% infrastructure          │
└──────────────┬──────────────────────────────┘
               │
        ┌──────┴──────┐
        │   Pass?     │
        └──────┬──────┘
               │
        ┌──────┴──────┐
        │     NO      │
        └──────┬──────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 12. Fix Surviving Mutants                   │
│     - Identify weak assertions              │
│     - Add targeted test cases               │
│     - Go back to step 2                     │
└──────────────┬──────────────────────────────┘
               │
        ┌──────┴──────┐
        │     YES     │
        └──────┬──────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 13. Commit Changes                          │
│     - git add -A                            │
│     - git commit -m "type(scope): message"  │
│     - Conventional commits format           │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 14. Check for More Work                     │
│     - Read todo list                        │
│     - Read plan/tasks documents             │
│     - Check for TODOs in code               │
└──────────────┬──────────────────────────────┘
               │
        ┌──────┴──────┐
        │ More Work?  │
        └──────┬──────┘
               │
        ┌──────┴──────┐
        │     YES     │
        └──────┬──────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 15. Start Next Task                         │
│     - Go back to step 1                     │
└─────────────────────────────────────────────┘
               │
        ┌──────┴──────┐
        │     NO      │
        └──────┬──────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ 16. DONE - All Work Complete                │
│     - All tests pass                        │
│     - All quality gates met                 │
│     - All tasks completed                   │
└─────────────────────────────────────────────┘
```

---

## Quality Gates (MANDATORY)

**Code Quality**:
- ✅ `go build ./...` - No compilation errors
- ✅ `golangci-lint run` - No warnings
- ✅ No new TODOs without tracking

**Test Quality**:
- ✅ `go test ./...` - All tests pass
- ✅ Coverage ≥95% (production code)
- ✅ Coverage ≥98% (infrastructure/utility code)
- ✅ No skipped tests without documentation

**Mutation Testing**:
- ✅ `gremlins unleash` - Efficacy ≥85% (production)
- ✅ `gremlins unleash` - Efficacy ≥98% (infrastructure)
- ✅ No surviving mutants without justification

**Integration/E2E** (when applicable):
- ✅ Database operations work (PostgreSQL + SQLite)
- ✅ External services integrate correctly
- ✅ Both `/service/**` and `/browser/**` paths tested

**Git**:
- ✅ Conventional commit format
- ✅ Clean working tree
- ✅ Changes match task scope

---

## When to Use Different Strategies

### Continuous Work Loop (DEFAULT)
**Use For**: Feature implementation, bug fixes, refactoring, documentation updates

**Pattern**: Implement → test → debug → iterate until complete

**Evidence Required**: All quality gates pass before moving to next task

### SpecKit Workflow
**Use For**: New features requiring design-before-implement, unclear requirements, architectural changes

**Pattern**: Constitution → Clarify → Spec → Plan → Tasks → Implement

**See**: [01-03.speckit.instructions.md](.github/instructions/01-03.speckit.instructions.md)

### Investigation
**Use For**: Understanding existing codebase, debugging complex issues, performance analysis

**Pattern**: Read code → run tests → trace execution → document findings

**Tools**: `semantic_search`, `grep_search`, `read_file`, `list_code_usages`

---

## Communication Patterns

**Progress Updates**: Update todo list after each task completion (NOT after each commit)

**Problem Identification**: Document blockers in DETAILED.md timeline, continue with unblocked tasks

**Completion**: Mark task complete ONLY when all quality gates pass with evidence

**Status**: Report status at natural breakpoints (task complete, phase complete), NOT mid-task

---

## Anti-Patterns - NEVER DO

❌ **Stop mid-loop to ask permission** - Continue until all quality gates pass
❌ **Skip quality gates** - Every gate must pass before proceeding
❌ **Batch commits** - Commit after each logical unit of work
❌ **Ask "should I continue?"** - Continue automatically until complete
❌ **Mark tasks complete without evidence** - Require objective proof (test output, coverage reports)
❌ **Stop when encountering blockers** - Document blocker, switch to unblocked task
❌ **Create session documentation files** - Append to DETAILED.md Section 2 timeline
❌ **Amend commits repeatedly** - Use incremental commits for history preservation

---

## Copilot Instructions vs Prompts vs Agents

### Instructions (.github/instructions/*.instructions.md)
**Purpose**: Persistent rules applied automatically based on file patterns

**Use For**:
- Coding standards (Go patterns, testing patterns)
- Architecture decisions (database, Docker, PKI)
- Project-specific conventions (import aliases, magic constants)

**Example**: `03-02.testing.instructions.md` applies automatically when working in `*_test.go` files

### Prompts (.github/prompts/*.prompt.md)
**Purpose**: Reusable workflows for common tasks

**Use For**:
- SpecKit workflows (constitution, clarify, plan, tasks, implement)
- Autonomous execution patterns
- Plan/task/QUIZME management

**Example**: `speckit.implement.prompt.md` for implementing features with SpecKit methodology

### Agents (NOT USED)
**Decision**: Project does NOT use agent files (.agent.md)

**Rationale**: Low maintenance burden, instructions + prompts provide sufficient guidance

---

## Key Takeaways

1. **Continuous Iteration**: Never stop until all quality gates pass
2. **Quality Gates Mandatory**: Coverage, mutation, linting, tests ALL must pass
3. **Evidence-Based Completion**: Objective proof required before marking complete
4. **Incremental Commits**: Commit after each logical unit, NOT batch
5. **Autonomous Execution**: Continue automatically without asking permission
6. **SpecKit for Design**: Use SpecKit when requirements unclear or architecture changes needed
7. **Low Maintenance**: Instructions + prompts only (no agents)

