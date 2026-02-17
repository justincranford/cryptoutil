# Strategic Decisions Quiz - Deployment Refactoring v1

**Purpose**: Clarify unknowns, risks, and strategic decisions before implementation
**Created**: 2026-02-17
**Status**: Awaiting User Responses

---

## Question 1: Archive vs Delete Legacy Directories

**Question**: What should we do with the old directories after migration is complete and validated?

**A)** Archive to `deployments/archived/` with timestamp (e.g., `deployments/archived/compose-legacy-2026-02-17/`)

**B)** Delete immediately after validation (git history provides backup)

**C)** Keep as deprecated with warning comment for 1 release cycle, then delete

**D)** Move to separate git branch and delete from main

**E)** 

**Answer**:

**Rationale**: Migration will move `deployments/compose/` and `deployments/cryptoutil/` to new names. After validation passes, we need to decide how to handle the old directories.

**Impact**:
- Option A: Safe rollback, increases repo size temporarily
- Option B: Clean break, relies on git history for rollback
- Option C: Gradual deprecation, gives users time to adapt
- Option D: Separate tracking, requires branch management

---

## Question 2: PRODUCT-Level E2E Compose Strategy

**Question**: How should we implement PRODUCT-level E2E compose for single-service products (sm, pki, cipher, jose)?

**A)** Create `deployments/cryptoutil-product/compose.yml` that includes all 5 PRODUCT-level composes (even single-service products)

**B)** Single-service products use their SERVICE-level compose for E2E; PRODUCT-level compose only for multi-service products (identity only)

**C)** Create separate PRODUCT-level E2E compose for each product in `deployments/{product}/compose.e2e.yml`

**D)** Single-service products skip PRODUCT-level testing (redundant with SERVICE-level)

**E)** 

**Answer**:

**Rationale**: Products with single services (sm-kms, pki-ca, cipher-im, jose-ja) may not need separate PRODUCT-level E2E testing since SERVICE and PRODUCT are identical.

**Impact**:
- Option A: Complete consistency, all products tested at all levels
- Option B: Pragmatic, avoids redundant testing
- Option C: Flexible per-product E2E patterns
- Option D: Simplest, but breaks consistency

---

## Question 3: Legacy E2E Test Migration Complexity

**Question**: What should we do with the legacy E2E tests in `internal/test/e2e/`?

**A)** Migrate all scenarios to service-template pattern + new compose files (preserve test coverage)

**B)** Rewrite from scratch using cipher-im E2E pattern as template (fresh start, may miss edge cases)

**C)** Delete legacy E2E entirely and rely on SERVICE/PRODUCT-level E2E tests (simplest, may lose scenarios)

**D)** Keep legacy E2E alongside new pattern for 1 release cycle, then deprecate after parallel validation

**E)** 

**Answer**:

**Rationale**: `internal/test/e2e/` contains legacy test code that uses `deployments/compose/`. These tests may have unique scenarios not covered by SERVICE/PRODUCT E2E tests.

**Impact**:
- Option A: Preserves coverage, most work
- Option B: Clean design, risk of losing test scenarios
- Option C: Fastest, highest risk of regression
- Option D: Safe transition, temporary duplication

---

## Question 4: Deployment Linter Backward Compatibility

**Question**: Should the deployment linter support the OLD directory names during transition?

**A)** Yes, support both old and new names with deprecation warnings for 1 release

**B)** No, hard cutover - linter only recognizes new names after migration

**C)** Yes, but old names are ERRORS not warnings (blocks CI/CD immediately)

**D)** Support old names indefinitely for backward compatibility

**E)** 

**Answer**:

**Rationale**: The linter validates deployment structure (`cicd lint-deployments`). During migration, we need to decide if it should recognize old names.

**Impact**:
- Option A: Gradual transition, allows external projects to adapt
- Option B: Clean break, enforces migration
- Option C: Strict enforcement, blocks usage of old names
- Option D: Permanent backward compatibility burden

---

## Question 5: E2E Compose File Naming Convention

**Question**: If we create separate E2E compose files (per Decision 2), what naming convention should we use?

**A)** `compose.yml` in PRODUCT/SERVICE directories (production compose IS E2E environment)

**B)** `compose.e2e.yml` alongside production compose (separate E2E variants)

**C)** `compose.test.yml` alongside production compose (emphasizes testing nature)

**D)** No separate E2E compose - use production compose files directly (current cipher-im pattern)

**E)** 

**Answer**:

**Rationale**: E2E tests need compose files. We must decide if prod compose suffices or if E2E needs separate files.

**Impact**:
- Option A/D: Simpler, one source of truth, what you test IS what you deploy
- Option B/C: Flexibility for E2E-specific config, risk of drift from production

---

## Question 6: CI/CD Workflow Update Strategy

**Question**: How should we update GitHub Actions workflows during migration?

**A)** Feature branch with updated workflows → merge after validation

**B)** Update workflows in-place on main, use conditional paths (if old exists, use old; else use new)

**C)** Duplicate workflows for old and new structure during transition period

**D)** Update workflows only after full migration complete (accept temporary CI/CD failures)

**E)** 

**Answer**:

**Rationale**: CI/CD workflows reference `deployments/compose/` and `deployments/cryptoutil/`. Migration will break workflows unless handled carefully.

**Impact**:
- Option A: Safe, tested before merge, blocks other PRs during migration
- Option B: Complex conditionals, works during transition
- Option C: Duplication, clear separation, easier rollback
- Option D: Risky, CI/CD broken during migration

---

## Question 7: Documentation Update Priority

**Question**: When should we update ARCHITECTURE.md and related docs?

**A)** Update docs BEFORE code migration (documentation-driven development)

**B)** Update docs AFTER code migration is complete (code-driven documentation)

**C)** Update docs incrementally per phase (docs and code evolve together)

**D)** Update docs only at end after validation (all changes known)

**E)** 

**Answer**:

**Rationale**: ARCHITECTURE.md is single source of truth, propagates to instruction files. Timing affects whether docs lead or follow implementation.

**Impact**:
- Option A: Docs describe target state, implementation follows spec
- Option B: Docs reflect actual implementation, no aspirational content
- Option C: Balanced, reduces drift, more coordination overhead
- Option D: Complete picture, may have stale docs during work

---

## Question 8: Port Range Validation Enforcement

**Question**: Should port validation be ERRORS (blocking) or WARNINGS (advisory) during migration?

**A)** ERRORS for new directories (cryptoutil-suite/product/service), WARNINGS for old (compose, cryptoutil)

**B)** ERRORS everywhere (strict enforcement)

**C)** WARNINGS everywhere during migration, ERRORS after migration complete

**D)** WARNINGS permanently (advisory only, not blocking)

**E)** 

**Answer**:

**Rationale**: Port validation ensures SERVICE (8XXX), PRODUCT (18XXX), SUITE (28XXX) ranges. We need to decide enforcement level during transition.

**Impact**:
- Option A: Strict for new, lenient for old (smooth transition)
- Option B: Strict enforcement, may block work if violations exist
- Option C: Gradual tightening, clear cutover point
- Option D: No enforcement, relies on convention

---

## Question 9: Service-Template E2E Helper Enhancement

**Question**: Should we extend `internal/apps/template/testing/e2e/ComposeManager` during migration?

**A)** Yes, add new features needed by SUITE-level E2E (wait for all 9 services healthy)

**B)** No, keep ComposeManager minimal - complex logic stays in test code

**C)** Yes, but only if 3+ services need same feature (avoid premature generalization)

**D)** Refactor ComposeManager completely to handle all 3 deployment levels uniformly

**E)** 

**Answer**:

**Rationale**: ComposeManager provides reusable E2E orchestration. Migration may reveal needs for enhancement.

**Impact**:
- Option A: Enhanced reusability, risk of over-engineering
- Option B: Simplicity, potential duplication in tests
- Option C: Balanced, evidence-based enhancement
- Option D: Major refactor, risky during migration

---

## Question 10: Magic Constants Organization

**Question**: How should we organize magic constants for new deployment paths?

**A)** New constants in existing files (e.g., add to `magic_cicd.go`)

**B)** New file `magic_deployment_refactoring.go` for migration-specific constants

**C)** Rename constants in place (e.g., `ComposeDeploymentDir` → `ComposeLegacyDir`)

**D)** Deprecate old constants, add new constants, support both during transition

**E)** 

**Answer**:

**Rationale**: Magic constants define paths like `deployments/compose/compose.yml`. Refactoring requires updating or adding constants.

**Impact**:
- Option A: Minimal disruption, grows existing files
- Option B: Clear separation, temporary file
- Option C: Clean break, breaks existing code
- Option D: Backward compatibility, temporary duplication

