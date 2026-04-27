# Lessons - Framework V17: internal/apps/ Structure Fitness Linters

**Created**: 2026-04-26
**Purpose**: Phase post-mortem lessons for V17 — populated by the execution agent after each phase's quality gates pass.

> **MANDATORY per-phase structure** (4 sections per phase):
>
> **What Worked**: Patterns, tools, or decisions that accelerated the work or prevented issues.
>
> **What Didn't Work**: Friction points, incorrect assumptions, or approaches that required rework.
>
> **Root Causes**: Underlying causes of the "What Didn't Work" items.
>
> **Patterns for Future Phases**: Actionable takeaways for subsequent phases or future plans.

---

## Executive Summary

1. [Phase 1: Gap Analysis & Linter Design](#phase-1-gap-analysis--linter-design) — Confirmed all gap matrix cells; all 10 PS-IDs audited against 12 structural invariants; linter spec finalized.
2. [Phase 2: Implement 6 New Fitness Linters](#phase-2-implement-6-new-fitness-linters) — Implemented `apps-ps-id-required-files`, `apps-ps-id-server-package`, `apps-ps-id-swagger-presence`, `apps-ps-id-test-patterns`, `apps-product-no-service-dirs`, `apps-suite-required-files` with ≥98% coverage.
3. [Phase 3: Registration, Integration & Knowledge Propagation](#phase-3-registration-integration--knowledge-propagation) — All 6 Phase 2 linters registered in lint_fitness.go and YAML; fitness_registry_completeness updated; ENG-HANDBOOK and target-structure updated.
4. [Phase 4: Template-Compliance Linters for cmd/ and internal/apps/](#phase-4-template-compliance-linters-for-cmd-and-internalapps) — Implemented 6 MANIFEST.yaml-driven linters: `apps-ps-id-template`, `apps-product-template`, `apps-suite-template`, `cmd-ps-id-template`, `cmd-product-template`, `cmd-suite-template`; 87 total linters passing.
5. [Phase 5: Conformance Migration — Fill All Gaps](#phase-5-conformance-migration--fill-all-gaps) — swagger.go moved to server/ for 7 PS-IDs; sm-kms server testmain/lifecycle/port_conflict created; 5 product service subdirs deleted; knownExclusions updated; tasks 5.2-5.7 (identity large moves, sm-im moves) deferred as GAPs.
6. [Phase 6: Knowledge Propagation](#phase-6-knowledge-propagation) — ENG-HANDBOOK catalog updated (12→24 architecture linters); SKILL.md linter counts updated (55→87); enforce_any test bugs fixed (pre-existing); format_go meta-test updated to match corrected assertions.

## Actions

1. Resolve 5 product service subdirectory GAPs (sm/kms/, sm/im/, jose/ja/, pki/ca/, skeleton/template/) — confirmed as usage stubs only, safe to delete in framework-v18.
2. Execute Tasks 5.2-5.7 (identity-authz, identity-idp, identity-rs, identity-rp, identity-spa large server code moves + sm-im test moves) — significant refactor, needs dedicated plan.
3. Fix pre-existing `literal-use` violations in `identity/` packages (TestLint_Integration fails with 769 violations in `identity/apperr/`, `identity/config/`, `identity/issuer/`, `identity/mfa/`) — not caused by V17 work.
4. Once GAP-E resolved, empty `knownServiceDirExceptions` in `apps-product-template` and `apps-product-no-service-dirs`.
5. Once Tasks 5.2-5.7 complete, empty all `knownExclusions` in Phase 2-4 linters — currently non-empty for lifecycle/port_conflict/swagger checks.

---

## Phase 1: Gap Analysis & Linter Design

**What Worked**:
- Registry-driven `AllProductServices()`, `AllProducts()`, `AllSuites()` API provided a single authoritative source for all PS-IDs — no hardcoded lists in gap analysis.
- Systematic per-PS-ID audit against 12 invariants surfaced all gaps quickly.
- Decided to separate `knownExclusions` per sub-check (testmain vs lifecycle vs port_conflict) — prevented over-scoping exclusions.

**What Didn't Work**:
- Initial plan assumed all lifecycle/port_conflict tests were already in `server/` — they were at PS-ID root. Required adding Phase 5 file-move tasks.
- swagger.go was at PS-ID root for all 8 PS-IDs; Phase 5 moves were required.

**Root Causes**:
- The original codebase structure predated the `server/` architectural decision. No linter enforced server/ location.
- Gap analysis was the first systematic audit of structural compliance across all 10 PS-IDs.

**Patterns for Future Phases**:
- Always audit file locations explicitly before writing linter specs — assume files may be in wrong directories.
- Separate exclusion lists per sub-check prevents cascading failures when some PS-IDs comply partially.

---

## Phase 2: Implement 6 New Fitness Linters

**What Worked**:
- Table-driven test structure with synthetic rootDir made tests independent of real codebase state — fast and reliable.
- Using `os.ReadDir` for suffix matching (lifecycle, port_conflict) was simpler and more reliable than glob patterns.
- `knownExclusions` map with TODO comments clearly documents what needs follow-on work.

**What Didn't Work**:
- `apps-ps-id-swagger-presence` initially checked PS-ID root instead of `server/` — had to update after Phase 5 swagger moves.
- `apps-ps-id-test-patterns` initially excluded sm-kms from testmain — after creating testmain in Phase 5, the exclusion was removed.

**Root Causes**:
- Linter invariants for file locations were defined before Phase 5 moves confirmed the exact target state.

**Patterns for Future Phases**:
- Write linters targeting the final desired state (server/ location), not current state — use exclusions for current gaps.
- Run `go run ./cmd/cicd-lint lint-fitness` after every linter change to catch registration errors immediately.

---

## Phase 3: Registration, Integration & Knowledge Propagation

**What Worked**:
- `fitness_registry_completeness` linter caught YAML/filesystem drift immediately — zero manual counting needed.
- Alphabetical ordering in lint_fitness.go and YAML was enforced by existing patterns.
- ENG-HANDBOOK §9.11.1 catalog update was straightforward with the existing table format.

**What Didn't Work**:
- Count updates in `fitness_registry_completeness_test.go` required careful arithmetic (68 + 6 = 74, then +12 more = 87 after Phase 4).
- Propagation drift validation (`lint-docs`) surfaced a few minor inconsistencies that required multiple passes.

**Root Causes**:
- No automated way to derive expected counts — must manually track +N across phases.

**Patterns for Future Phases**:
- Record the exact linter count at the start of each phase so arithmetic is clear.
- Run `lint-docs` and `lint-fitness` together after every documentation update.

---

## Phase 4: Template-Compliance Linters for cmd/ and internal/apps/

**What Worked**:
- MANIFEST.yaml-driven pattern separated configuration from code — adding new required files means only updating the YAML template, not the linter Go code.
- `cmd-suite-template` correctly identified the critical difference: suite uses `os.Args` (full), PS-IDs use `os.Args[1:]` (sliced) — this asymmetry was caught and documented.
- All 6 linters passed `lint-fitness` on first integration attempt.

**What Didn't Work**:
- Initial `apps-suite-required-files` (Phase 2) was superseded by `apps-suite-template` (Phase 4) — mild duplication before retirement.

**Root Causes**:
- Phase 2 linters were simpler implementations; Phase 4 MANIFEST.yaml approach was more robust but required Phase 2 as a stepping stone.

**Patterns for Future Phases**:
- When a template-driven linter supersedes a simpler one, retire the old linter in the same phase (not deferred) to keep counts accurate.
- Document `os.Args` vs `os.Args[1:]` asymmetry explicitly in linter comments — easy to get wrong.

---

## Phase 5: Conformance Migration — Fill All Gaps

**What Worked**:
- swagger.go moves to server/ were clean for 7 PS-IDs (sm-kms, sm-im, jose-ja, pki-ca, skeleton-template, identity-rp, identity-spa) — no import changes needed since swagger files have no exported symbols referenced externally.
- Creating sm-kms `server/testmain_test.go`, `kms_lifecycle_test.go`, `kms_port_conflict_test.go` followed the sm-im pattern exactly.
- Deleting 5 product service subdirs (sm/kms/, sm/im/, jose/ja/, pki/ca/, skeleton/template/) was safe — all contained only usage stubs already present in the PS-ID directories.

**What Didn't Work**:
- `testmain_test.go` initially used `RequireNewForTest(cryptoutilSharedMagic.OTLPServiceSMKMS)` without setting `cfg.DatabaseURL = SQLiteInMemoryDSN` — caused PostgreSQL connection failures in unit tests.
- `require.Greater(t, server.PublicPort(), uint16(0))` failed because `PublicPort()` returns `int` not `uint16` — type mismatch in testify assertion.
- Tasks 5.2-5.7 (identity-authz, identity-idp, identity-rs large moves + sm-im test moves) were too large to complete in V17 session — deferred as documented GAPs.

**Root Causes**:
- Default `DatabaseURL` in `RequireNewForTest` is PostgreSQL; must explicitly override with `SQLiteInMemoryDSN` for unit tests.
- `PublicPort()` and `AdminPort()` return `int`, not `uint16` — testify `require.Greater` is strict about type matching.
- identity services have 80+ root files that need coordinated moves — underestimated in planning.

**Patterns for Future Phases**:
- ALWAYS set `cfg.DatabaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN` after `RequireNewForTest` in unit test contexts.
- Use `require.Greater(t, server.PublicPort(), 0)` (int literal) NOT `uint16(0)` — check return type before writing assertions.
- When creating server lifecycle tests, look at sm-im as the canonical reference implementation.
- Document large moves as explicit GAP tasks with acceptance criteria before starting — never begin a 80-file refactor without a dedicated plan.

---

## Phase 6: Knowledge Propagation

**What Worked**:
- ENG-HANDBOOK.md catalog update (12→24 architecture linters) was clean using the existing table format.
- SKILL.md linter count update (55→87) was a single-line change in both `.github/` and `.claude/` counterparts.
- `lint-docs` passed immediately after both updates.

**What Didn't Work**:
- Pre-existing bug in `enforce_any_test.go`: `require.NotContains(t, ..., "any", ...)` was contradictory (checking that file doesn't contain "any" after replacing interface{} with "any"). Required fixing both occurrences.
- `format_go_self_mod_test.go` meta-test assertions were outdated — they checked for the old incorrect message and old constant format.
- Pre-existing `TestLint_Integration` failure: 769 `literal-use` violations in `identity/` packages unrelated to V17 work.

**Root Causes**:
- `enforce_any_test.go` bugs existed before V17 — the `NotContains` message said "any" but should check "interface{}" (the thing being removed).
- Meta-test in `format_go_self_mod_test.go` was written against an older version of the test file that used raw "any" literals instead of split concatenation.
- `TestLint_Integration` failures are long-standing technical debt in identity/ packages.

**Patterns for Future Phases**:
- After fixing test assertions, always run the meta-test (`TestEnforceAnyDoesNotModifyItself`) as part of the verification cycle.
- When a test fixes a pre-existing bug, document it explicitly in the commit message — distinguishes from new regressions.
- Pre-existing failing tests should be tracked as known issues before starting a plan — prevents confusion about whether new changes caused them.

## Phase 1: Gap Analysis & Linter Design

*(To be filled during Phase 1 execution using the 4-section structure above)*

---

## Phase 2: Implement 6 New Fitness Linters

*(To be filled during Phase 2 execution using the 4-section structure above)*

---

## Phase 3: Registration, Integration & Knowledge Propagation

*(To be filled during Phase 3 execution using the 4-section structure above)*
