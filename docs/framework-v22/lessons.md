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

**What Worked**:
- Reader-function injection (`lintWithReader`, `checkInDirWithReader`, `findViolationsWithReader`) allowed deterministic error-path coverage without introducing package-level mutable seams.
- Adding focused internal tests beside external behavior tests closed branch gaps quickly, especially around non-happy-path filesystem and reader failures.
- Running package coverage first, then full fitness/build/lint gates, prevented broad validation cycles while branch-level gaps still existed.

**What Didn't Work**:
- Initial refactor left stale imports and style violations (`nlreturn`, `wsl_v5`), which blocked the global lint gate despite functional correctness.
- A first-pass read-error test for orchestration policy only exercised the server path; client-path error propagation remained uncovered.

**Root Causes**:
- Structural changes in linter files were made before immediate compile checks, leaving dead imports until the first coverage run.
- Branch targeting was initially coarse-grained (function-level), not path-granular (server/client + stat/read variants), causing repeated test iterations.

**Patterns for Future Phases**:
- For linter packages, design seam points as function parameters (not package vars) from the first change, and add direct internal tests for each decision branch.
- After each new internal test file, run `gofmt` and `golangci-lint run` immediately to avoid end-of-phase style debt.
- Use coverage evidence directories per phase (`test-output/v22-phase3/`) and require function-level reports before declaring ≥98% complete.

---

## Phase 4: Mutation Testing

**What Worked**:
- Running gremlins with `--workers=1` and higher timeout coefficient stabilized linter-package mutation runs enough to get deterministic efficacy results.
- Adding mutation-targeted assertions (exact line number checks and branch-focused directory discovery tests) eliminated surviving linter mutants.
- Centralizing all evidence under `test-output/v22-mutation/` made it straightforward to separate package-level passes from environment-driven failures.

**What Didn't Work**:
- Default Windows gremlins runs across helper packages produced frequent temp-folder unlink failures and large timeout clusters.
- Initial wildcard package invocation (`...`) failed coverage discovery for linter paths on Windows PowerShell, requiring direct package path invocation.

**Root Causes**:
- Windows file locking against large copied worktrees (especially under `test-output/` and transient temp copies) caused cleanup failures and unreliable non-zero exits unrelated to mutation efficacy.
- Some helper-package mutation scenarios are sensitive to process/runtime variability and require Linux CI execution for stable timeout behavior.

**Patterns for Future Phases**:
- For mutation work on Windows, always start with tuned gremlins flags: `--workers=1`, increased timeout coefficient, and explicit output artifacts.
- Treat helper-package mutation instability as CI-deferred only when local evidence is captured and the exact workflow step is referenced (`.github/workflows/ci-mutation.yml`, `Run mutation tests (informational)`).
- When a lived mutant appears, add the smallest assertion that validates the mutated semantic (for example, exact computed field values) before rerunning mutation.

---

## Phase 5: test_orch_e2e Facade + 10 PS-ID E2E TestMain Migration + Linter

**What Worked**:
- Introducing a dedicated `test_orch_e2e` facade with pass-through mode allowed both legacy full-orchestration E2E TestMains and identity trivial TestMains to converge on one import path with minimal churn.
- Adding a purpose-built `testmain-e2e-policy` fitness linter prevented regressions by enforcing both sides of the rule: required `test_orch_e2e` import and forbidden `testing/e2e_infra` import.
- Central registration updates in `lint_fitness.go` plus `lint-fitness-registry.yaml` avoided registry drift and immediately satisfied `fitness-registry-completeness` checks.
- Branch-focused tests (including injected walk/read seams) produced 100% coverage for the new policy package and made lint/error paths deterministic.

**What Didn't Work**:
- Early validation runs failed due temporary root artifacts (`coverage`, `coverage.out`) created during local coverage probing, which triggered `root-junk-detection` failures.
- Initial policy implementation tripped `if-else-chain` and `gofumpt`/`wsl_v5` style gates, requiring follow-up cleanup before quality gates could pass.
- Attempting `gofmt` on YAML during a combined command chain produced a non-Go parsing error and unnecessary rerun.

**Root Causes**:
- Coverage exploration commands were executed in root without immediate cleanup, and this repo treats root artifact hygiene as a blocking architecture gate.
- New linter logic was functionally correct but initially not shaped to project-specific lint expectations (consecutive-if style and strict spacing/formatting).
- Tool-chain batching mixed file types with Go-only formatters, causing avoidable command failure noise.

**Patterns for Future Phases**:
- After any ad hoc coverage investigation, immediately delete temporary root artifacts before running `lint-fitness`.
- For new fitness linters, add seam-based internal tests at creation time and target explicit branch closure before phase-wide gate runs.
- Keep registration synchronized across both execution registry (`lint_fitness.go`) and metadata registry (`lint-fitness-registry.yaml`) in the same change-set.
- Restrict `gofmt` invocations to Go files only; use YAML-specific tooling for manifest files.

---

## Phase 6: Framework-Internal TestMain Migration

**What Worked**:
- Adding `NewTestServerSettingsForTestMain()` and `NewTestTLSSettingsForTestMain()` kept TestMain setup deterministic without forcing `*testing.T` into a lifecycle that does not have one.
- Introducing `ConfigureTestFixtures(...)` preserved the existing server test accessor surface, so the migration stayed localized to TestMain files instead of cascading into every test consumer.
- Reusing the existing server subtree tests as the validation target made it easy to prove the new setup still supports the same fixture-driven expectations.

**What Didn't Work**:
- The first migration approach tried to reuse the `*testing.T`-based helper constructors directly from TestMain, which is not valid in that context.
- The migration would have stalled if we had tried to rewrite all consumers away from the shared accessors in the same phase.

**Root Causes**:
- TestMain executes outside the normal `testing.T` helper lifecycle, so helper APIs that require `t` need explicit no-`testing.T` wrappers.
- The server test fixture design still has some shared accessors, which means a compatibility setter is the lowest-risk bridge during a migration phase.

**Patterns for Future Phases**:
- When migrating TestMain setup, add explicit `ForTestMain` wrapper helpers rather than trying to adapt `*testing.T`-only helpers.
- Preserve shared fixture accessors until the surrounding test subtree is ready for a broader refactor; use a setter to bridge the new initialization path.
- Validate the migration with the narrowest relevant package tree first, then widen only if the focused subtree passes cleanly.

---

## Phase 7: sm-kms businesslogic + orm Migration

**What Worked**:
- Replacing `StartCore` with helper-driven setup kept the sm-kms test bootstrap aligned with the rest of the framework while removing the heavy core dependency.
- Moving the businesslogic tests onto a shared fixture built from local telemetry, JWK, barrier, and SQLite helpers preserved the existing test semantics without needing to rewrite the test bodies.
- The ORM package simplified cleanly to a shared in-memory SQLite DB plus local telemetry/JWK services, which made the migration small and easy to validate.

**What Didn't Work**:
- Trying to import the KMS server package from the businesslogic TestMain created an import cycle, so that path had to be abandoned.
- The first businesslogic fixture pass missed the KMS schema tables and the barrier tables, which caused runtime failures even though the package built.
- The ORM TestMain rewrite initially misread the helper return shape and tried to capture an error that does not exist.

**Root Causes**:
- Test package placement matters: code in `package businesslogic` cannot import the server package that itself imports businesslogic, so the migration had to stay self-contained.
- Helper-based TestMain replacement still needs explicit schema initialization for every table family the tests touch; `StartCore` had been hiding that work.
- Function signatures for test helpers are easy to misremember; the smallest safe fix is to read the local helper and adapt the call site, not infer the return contract.

**Patterns for Future Phases**:
- Prefer self-contained test fixtures inside the package when importing the production server would create a cycle.
- After removing a framework bootstrap, immediately enumerate the schema tables the tests actually touch and migrate them explicitly in TestMain.
- Verify helper signatures before editing call sites; if a helper is `Must`/single-return, keep the call site equally simple.

---

## Phase 8: Consumer Migration + Old testing/ Deprecation

**What Worked**:
- Building the consumer census first made it obvious that the phase was smaller and more concrete than the initial task text implied, which let the migration proceed file-by-file instead of guessing.
- Swapping the direct helper consumers to `test_help_db`, `test_help_cli`, and `test_orch_integration` kept the changes localized and preserved the existing test behavior.
- Using the real failing `sm-im/server/apis` package as a canary caught a hidden `TestMain` misuse early and prevented a false-green migration.
- Adding package-level `Deprecated` docs made the remaining legacy helper packages intentionally transitional instead of silently kept forever.

**What Didn't Work**:
- The first broad app-tree validation exposed a hidden `NewInMemorySQLiteDB(&testing.T{})` misuse in `messages_test.go`, which looked unrelated until the TestMain helper semantics were rechecked.
- A rough consumer estimate in the task text was too high; the actual census was lower, so the phase needed evidence rather than assumption to close cleanly.

**Root Causes**:
- `NewInMemorySQLiteDB` is the wrong helper for `TestMain`; it needs the TestMain-safe constructor so the database lifecycle matches the package scope.
- The old `testing/` packages were not a single category of work: some had direct replacements, while others were still useful shared legacy utilities that needed explicit deprecation instead of migration.
- Import swaps alone are insufficient when helper constructors differ subtly; lifecycle semantics matter as much as package names.

**Patterns for Future Phases**:
- Start consumer migrations with a census artifact and treat the count as evidence, not a planning guess.
- When a package already has a TestMain, prefer `NewInMemorySQLiteDBForTestMain()` over the per-test helper and validate the lifecycle with the package's own tests before widening the scope.
- If a legacy helper has no clean replacement, keep it only with explicit `Deprecated` documentation so the transition state is visible to future maintainers.

---

## Phase 9: E2E Validation

*(To be filled during Phase 9 execution using the 4-section structure above)*

---

## Phase 10: TestMain Inventory Table

*(To be filled during Phase 10 execution using the 4-section structure above)*

---

## Phase 11: Knowledge Propagation

*(To be filled during Phase 11 execution using the 4-section structure above)*
