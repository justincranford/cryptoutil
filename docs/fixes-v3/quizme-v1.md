# Quizme v1 - Configs/Deployments/CICD Clarifications

## Question 1: Environment-Specific Config Files Handling

**Question**: What should we do with existing environment-specific config files (development.yml, production.yml, test.yml) in configs/identity/?

**Context**: configs/identity/ contains:
- development.yml (unclear purpose)
- production.yml (unclear purpose)  
- test.yml (unclear purpose)
- policies/ directory (OK - shared policy definitions)
- profiles/ directory (OK - shared profile definitions)

**Impact**: Affects Task 1.3 (Restructure identity/)

**A)** Archive to configs/orphaned/identity-envs/ (treat as obsolete)

**B)** Keep in configs/identity/ as environment templates (rename to identity-ENV.yml for clarity)

**C)** Move to new locations: configs/identity-authz/identity-authz-app-ENV.yml (per-service overrides)

**D)** Analyze content first, then decide (may be different purposes: ENV configs vs deployment profiles)

**E)**

**Answer**:

**Rationale**: This affects how we structure configs/identity/ during Phase 1 Task 1.3

---

## Question 2: PRODUCT/SUITE Deployment Structure

**Question**: Should we create deployments/PRODUCT/config/ and deployments/cryptoutil/config/ directories (not just configs/PRODUCT/ and configs/cryptoutil/)?

**Context**: Currently:
- deployments/PRODUCT/compose.yml exists (delegation only)
- deployments/PRODUCT/ has NO config/ subdirectory
- configs/PRODUCT/ will have configs (templates/examples)

Decision 2 in plan.md: configs/ for templates, deployments/ for runtime

**Impact**: Affects Phase 2 scope and implementation

**A)** Yes, create both configs/PRODUCT/ AND deployments/PRODUCT/config/ (full parity)

**B)** No, only create configs/PRODUCT/ (PRODUCT-level deployments use delegation only, no runtime configs)

**C)** Create deployments/PRODUCT/config/ later if needed (defer until actual use case)

**D)** Create symlinks from deployments/PRODUCT/config/ → ../../configs/PRODUCT/ (avoid duplication)

**E)**

**Answer**:

**Rationale**: Affects whether Phase 2 creates configs in deployments/ or only in configs/

---

## Question 3: Validation Enforcement Strategy

**Question**: How strict should CICD config validation be for existing configs that don't comply with new patterns?

**Context**: After implementing Phase 3/4 validations, existing configs may have violations (old naming, missing fields, etc.)

**Impact**: Affects validation enforcement mode and migration timeline

**A)** Strict: All violations are ERRORS (blocks CI/CD until fixed)

**B)** Hybrid: Critical violations ERRORS, minor violations WARNINGS (allows gradual migration)

**C)** Lenient: All violations are WARNINGS initially (gives time to migrate)

**D)** Configurable: Add flag to toggle strict/lenient mode (max flexibility)

**E)**

**Answer**:

**Rationale**: Current Decision 3 in plan.md selects Option C (Hybrid), confirm or override

---

## Question 4: Migration Timing & Rollout

**Question**: Should we execute the full migration (Phase 1-6) in a single PR, or split into multiple PRs?

**Context**: 56 tasks across 6 phases, ~58 hours estimated

**Impact**: Affects implementation strategy and review process

**A)** Single PR: All phases together (atomic, prevents inconsistency)

**B)** Per-Phase PRs: 6 PRs, one per phase (easier review, incremental progress)

**C)** Per-Critical-Path PRs: Phase 1+2 together, Phase 3+4 together, Phase 5+6 together (logical groupings)

**D)** Feature Branch: All work in feature branch, squash to main when complete (clean history)

**E)**

**Answer**:

**Rationale**: Affects how we organize commits and PRs during implementation

---

## Question 5: Pre-Commit Hook Scope

**Question**: Should pre-commit hooks validate ALL files or only CHANGED files?

**Context**: Validating all configs/ and deployments/ files on every commit may be slow (>30s)

**Impact**: Affects Task 5.4 (Performance Optimization)

**A)** All files: Ensure consistency across entire project (slower, comprehensive)

**B)** Changed files only: Fast validation (faster, may miss cross-file issues)

**C)** Hybrid: Changed files for incremental commits, all files for pre-push/CI (balanced)

**D)** Configurable: Add VALIDATE_ALL env var to toggle (developer choice)

**E)**

**Answer**:

**Rationale**: Current plan assumes Option B (changed files only), confirm or override

---

## Question 6: Naming Pattern for PRODUCT-level Configs

**Question**: Should PRODUCT-level config files use PRODUCT-app-VARIANT.yml or PRODUCT-config-VARIANT.yml?

**Context**: SERVICE-level uses PRODUCT-SERVICE-app-VARIANT.yml (e.g., cipher-im-app-common.yml)

**Impact**: Affects Phase 2 naming consistency

**A)** PRODUCT-app-VARIANT.yml (parallel to SERVICE pattern: cipher-app-common.yml)

**B)** PRODUCT-config-VARIANT.yml (explicit "config" infix: cipher-config-common.yml)

**C)** PRODUCT-VARIANT.yml (shorter, no "app" infix: cipher-common.yml)

**D)** Keep existing pattern from Decision 1 in plan.md (PRODUCT-app-VARIANT.yml)

**E)**

**Answer**:

**Rationale**: Current plan uses Option A/D (PRODUCT-app-VARIANT.yml), confirm or override

---

## Question 7: Port Offset Validation Strictness

**Question**: Should port offset validation ERROR or WARN for deviations from standard offsets (+0, +10000, +20000)?

**Context**: Some services may legitimately need different ports (e.g., multiple instances for testing)

**Impact**: Affects Task 3.4 and Task 4.4 implementation

**A)** ERROR: Enforce strict port offsets (no exceptions)

**B)** WARN: Allow deviations with warning (document rationale)

**C)** CONFIGURABLE: Add allowlist for exceptions (e.g., test-specific ports)

**D)** SKIP: Don't validate port offsets (too restrictive)

**E)**

**Answer**:

**Rationale**: Current plan assumes ERROR enforcement, confirm or override

---

## Question 8: README.md Content Requirements

**Question**: What should PRODUCT/SUITE README.md files contain?

**Context**: Phase 2 Tasks 2.1-2.6 create README.md files in each PRODUCT/SUITE directory

**Impact**: Affects Task 2.* acceptance criteria

**A)** Minimal: Purpose, delegation pattern, link to ARCHITECTURE.md (quick overview)

**B)** Comprehensive: Purpose, delegation, config examples, secret sharing, port offsets (full reference)

**C)** Template-based: Use template from deployments/template/README.md (consistency)

**D)** Generated: Auto-generate from compose.yml and config files (always in sync)

**E)**

**Answer**:

**Rationale**: Current plan assumes Option A (minimal), confirm or override
