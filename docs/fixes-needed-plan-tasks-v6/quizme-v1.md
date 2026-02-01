# QuizMe v1 - Phase 11 Decisions

**Created**: 2026-02-01
**Purpose**: Clarify decisions for Phase 11 cleanup tasks

---

## Question 1: test-output/ Directory Handling

**Context**: The `test-output/` directory contains ~40 coverage files from LLM analysis work. This directory is already in .gitignore.

**Files in test-output/**:
- `barrier-coverage-gap-analysis/coverage.html`, `coverage.out`
- `kms-migration-analysis/architecture-comparison.md`
- Various `*_coverage.out`, `*_coverage.html` files

**Question**: What should be done with `test-output/` coverage files?

- A) Delete ALL files in test-output/ (clean slate)
- B) Keep test-output/ as-is (analysis artifact directory, already gitignored)
- C) Keep named analysis directories (e.g., `barrier-coverage-gap-analysis/`, `kms-migration-analysis/`), delete loose files
- D) Delete only *.out and *.html files, keep markdown analysis docs
- E) [Fill in your preferred approach]:

**Choice**: ___

---

## Question 2: CICD Linter Scope

**Context**: Task 11.4 proposes adding a CICD linter to detect leftover coverage files.

**Question**: Where should the linter detect leftover files?

- A) Root directory only (most common mistake)
- B) Root directory + internal/ directories
- C) Root directory + internal/ + any directory except test-output/
- D) All directories including test-output/ (strict enforcement)
- E) [Fill in your preferred scope]:

**Choice**: ___

---

## Question 3: Coverage File Patterns

**Context**: The linter needs to know which file patterns to detect.

**Question**: Which patterns should the linter flag?

- A) Only `*.out` files
- B) Only `*.out` and `*coverage*.html` files
- C) `*.out`, `*coverage*.html`, and `cover.out` files
- D) All of C plus `*.prof` profiling files
- E) [Fill in your preferred patterns]:

**Choice**: ___

---

## Question 4: Linter Behavior

**Context**: When the linter finds leftover files, it can either warn or fail.

**Question**: Should the linter fail the build or just warn?

- A) Warn only (allow commit, just notify)
- B) Fail in CI/CD, warn locally
- C) Always fail (strict enforcement)
- D) Configurable via flag (default: warn)
- E) [Fill in your preferred behavior]:

**Choice**: ___

---

## Notes

After you fill in your choices, I will:
1. Execute Tasks 11.1-11.3 based on your decisions
2. Implement Task 11.4 CICD linter with your specified scope and behavior
3. Delete this quizme file and update plan.md/tasks.md with your decisions
