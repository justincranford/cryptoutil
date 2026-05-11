# Lessons - Framework V22: V21 Audit Fix Campaign

**Created**: 2026-05-11
**Last Updated**: 2026-05-11

> **Mandatory per-phase structure** (fill during each phase post-mortem after quality gates pass):
>
> **What Worked**: Techniques, patterns, or tools that produced good results
>
> **What Didn't Work**: Approaches that failed or caused rework
>
> **Root Causes**: Why failures occurred (not just symptoms)
>
> **Patterns for Future Phases**: Reusable guidance extracted from this phase's experience

---

## Executive Summary

*(To be filled at plan completion — numbered links to each phase section with one-sentence outcome)*

---

## Actions

*(To be filled at plan completion — numbered list of concrete follow-up items for reviewer, specific enough to copy-paste directly into Copilot Chat or Claude Code as a follow-up prompt)*

---

## Phase 1: Implement Empty Stub Packages

**What Worked**:
- Reusing existing framework patterns from server/testutil and service repository TestMain setup reduced design drift and avoided introducing new TLS/barrier abstractions.
- Running targeted build/lint gates first, then the phase-wide quality gate, caught local formatting issues before full-repo validation.
- Enforcing the file-length acceptance criterion explicitly prevented another false-positive "implemented" claim for near-empty stubs.

**What Didn't Work**:
- Initial implementation pass failed lint due to gofumpt and wsl_v5 spacing issues.
- The first bootstrap helper implementation met behavior criteria but failed the explicit >50-line requirement.

**Root Causes**:
- Lint-first discipline was applied after code changes but before formatting auto-fixes, so style violations surfaced late in the task cycle.
- Acceptance criteria included a structural threshold (>50 lines) that was not checked immediately after implementation.

**Patterns for Future Phases**:
- After each file implementation task, run an immediate three-check mini-gate: package build, package lint, and line-count/grep acceptance proof where required.
- Keep helper implementations concrete and deterministic (no package-level mutable state) to preserve t.Parallel safety for upcoming high-coverage test phases.
- Update tasks.md and lessons.md in the same execution window as quality-gate completion to prevent documentation lag.

---

## Phase 2: Self-Tests for All 7 Helper Packages

**What Worked**:
- Building seam-injection points directly into helper packages (`test_help_tls`, `test_help_barrier`, `test_help_db`) made previously hard-to-hit error paths deterministic and testable.
- Table-driven subtests with explicit sequential exemptions for package-level seam mutation prevented race-induced flakiness while preserving parallelism elsewhere.
- Generating per-package coverage profiles under `test-output/v22-phase2/` created objective evidence and made branch-gap diagnosis fast.

**What Didn't Work**:
- Initial test pass had widespread lint failures (`gofumpt`, `wsl_v5`, `importas`, `bodyclose`, `wrapcheck`) because high-volume file creation happened before a lint-fix pass.
- `test_help_db` coverage appeared inconsistent (`go test -cover` summary vs function totals) until package-scope seam-literal statements were explicitly exercised.
- Parallel subtests that mutated package-level seam vars caused interference in early TLS/barrier iterations.

**Root Causes**:
- Fast bulk test scaffolding without immediate lint verification introduced style/safety debt that blocked quality gates.
- Coverage accounting includes package-scope function-literal statements; function-level coverage alone masked remaining uncovered init-scope paths.
- Package-level mutable seam variables require explicit sequential handling to avoid cross-subtest contamination.

**Patterns for Future Phases**:
- For any helper package targeting ≥98% coverage, introduce seam variables up front and add one dedicated test that executes default seam literals.
- After creating >2 test files in a burst, run `golangci-lint run --fix` immediately, then mandatory second-pass `golangci-lint run`.
- Mark tests as sequential whenever package-level seam variables are mutated, and keep all other tests `t.Parallel()` to preserve concurrency coverage.

---

## Phase 3: Linter Coverage to ≥98%

*(To be filled during Phase 3 execution using the 4-section structure above)*

---

## Phase 4: Mutation Testing

*(To be filled during Phase 4 execution using the 4-section structure above)*

---

## Phase 5: test_orch_e2e Facade + 10 PS-ID E2E TestMain Migration + Linter

*(To be filled during Phase 5 execution using the 4-section structure above)*

---

## Phase 6: Framework-Internal TestMain Migration

*(To be filled during Phase 6 execution using the 4-section structure above)*

---

## Phase 7: sm-kms businesslogic + orm Migration

*(To be filled during Phase 7 execution using the 4-section structure above)*

---

## Phase 8: Consumer Migration + Old testing/ Deprecation

*(To be filled during Phase 8 execution using the 4-section structure above)*

---

## Phase 9: E2E Validation

*(To be filled during Phase 9 execution using the 4-section structure above)*

---

## Phase 10: TestMain Inventory Table

*(To be filled during Phase 10 execution using the 4-section structure above)*

---

## Phase 11: Knowledge Propagation

*(To be filled during Phase 11 execution using the 4-section structure above)*
