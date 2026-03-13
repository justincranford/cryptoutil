# Lessons Learned - Framework v2

This file captures lessons from each phase, used as:
1. Memory for the entire plan.md / tasks.md execution
2. Input for knowledge propagation to ARCHITECTURE.md, agents, skills, instructions

---

## Phase 1: testdb.NewClosedSQLiteDB Helper

### What Worked

- **Test seam injection**: `buildClosedSQLiteDB` with injectable `openFn` achieves 100% coverage on the internal function while `NewClosedSQLiteDB` hits 80% (t.Fatalf ceiling).
- **Fitness rule architecture**: Following existing lint_fitness pattern (Check/CheckInDir/CheckFiles/CheckFile) made the rule easy to implement and test.
- **Table-driven tests**: All multi-case tests used tables, enabling 11 test functions for 100% coverage on the fitness rule.

### What Didn't Work

- **PowerShell heredoc**: Loses tab characters in Go files. Required Python scripts as workaround for file generation.
- **Pre-commit hook chain**: Commits take 3-5 minutes due to running all CI/CD checks (lint-fitness, lint-go, lint-go-test, lint-golangci, etc.). Multiple round-trips when hooks find issues.
- **Coverage target mismatch**: Task 1.3 specifies ≥98% for testdb package, but testdb has Docker-dependent functions (NewPostgresTestContainer, etc.) creating a structural ceiling of ~64%. The NEW code (NewClosedSQLiteDB/buildClosedSQLiteDB) is at 80-100%.

### Root Causes

- `goconst`: Repeated string literals in tests must be constants.
- `noctx`: Pre-existing violations (`sqlDB.Ping()` → `PingContext`) discovered during commit hooks.
- `lint-go literal-use`: Magic constants (`.git`, `vendor`) must use `cryptoutilSharedMagic` values, not string literals.

### Patterns Discovered

- **Incremental lint discovery**: Each pre-commit pass may find new issues from different linters. Fix all before re-staging.
- **Coverage ceiling documentation**: Docker-dependent packages need per-package exception docs per ARCHITECTURE.md §10.2.3.
- **Fitness rule registration timing**: Rule is NOT registered in lint_fitness.go during Phase 1 because it would fail on pre-existing jose-ja/sm-im violations. Registration deferred to Phase 5 after cleanup.

### Decisions

- **D1**: Fitness rule NOT registered until Phase 5 (would break TestLint_Integration and pre-commit).
- **D2**: testdb coverage accepted at 64.1% overall (structural ceiling from Docker deps); new code at 80-100%.
- **D3**: Pre-existing noctx violations fixed in same commit as new code (shared testdb package).

---

## Phase 2: jose-ja Cleanup

*(To be filled during Phase 2 execution)*

---

## Phase 3: sm-im Cleanup

*(To be filled during Phase 3 execution)*

---

## Phase 4: sm-kms Assessment and Safe Cleanup

*(To be filled during Phase 4 execution)*

---

## Phase 5: Knowledge Propagation

*(To be filled during Phase 5 execution)*
