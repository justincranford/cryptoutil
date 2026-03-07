# Quiz - Framework v1 Decisions

**Purpose**: Clarify remaining unknowns before implementation begins.
**Created**: 2026-03-06
**Instructions**: For each question, write your answer (A, B, C, D, or E with custom text) on the **Answer:** line. After all questions are answered, this file will be merged into plan.md/tasks.md and deleted.

---

## Question 1: KMS Interface Compliance Strategy

**Question**: KMS has unique method signatures (Start/Shutdown without context, IsReady instead of SetReady). How should we handle KMS in the ServiceContract interface?

**A)** Adapter pattern: Create a `KMSAdapter` wrapper that wraps KMSServer and forwards calls with the expected signatures (e.g., `Start(ctx)` calls `kms.Start()`). KMS satisfies the interface through its adapter.
**B)** Two interfaces: Define `ServiceServer` (core, all services) and `ServiceServerWithContext` (extended, 9 standard services). KMS implements the core interface only. Contract tests run against the core interface.
**C)** Unify KMS: Modify KMS to match the standard signatures (`Start(ctx)`, `Shutdown(ctx) error`). This brings KMS into full conformance but requires internal changes to KMS.
**D)** Exclude KMS: KMS is infrastructure (provides barrier for others). Exclude it from the ServiceServer contract entirely and test it separately.
**E)**

**Answer**:

**Rationale**: KMS is the only service where Start/Shutdown signatures diverge. The choice affects how many interfaces we maintain and whether contract tests cover KMS.

---

## Question 2: Builder Simplification Scope

**Question**: How aggressively should we simplify the builder pattern? Some With*() methods are already effectively no-ops (barrier is always on), while others provide genuine configuration (JWTAuth modes, StrictServer OpenAPI spec injection).

**A)** Aggressive: Remove all standard With*() calls from service code. Build() auto-configures everything. Only `WithDomainMigrations()` and `WithPublicRouteRegistration()` survive. JWTAuth defaults to session mode. StrictServer auto-discovers from service's OpenAPI spec.
**B)** Moderate: Make barrier, sessions, realm, registration automatic. Keep `WithJWTAuth()` and `WithStrictServer()` explicit because they take configuration parameters that differ between services (KMS uses JWT required mode; services using OpenAPI need to pass their spec).
**C)** Conservative: Keep all With*() methods but add `NewStandardServerBuilder()` that pre-calls the standard set. Services use the standard builder; KMS uses the custom builder. No existing API changes.
**D)** Status quo: Don't simplify the builder. Focus effort on contracts/fitness/testing instead. Builder simplification is cosmetic and risks regressions.
**E)**

**Answer**:

**Rationale**: The user expressed strong preference for simplification, but the JWTAuth and StrictServer calls carry real configuration parameters. This question calibrates how much auto-configuration vs explicit configuration to use.

---

## Question 3: Fitness Functions — Existing Check Migration

**Question**: Several fitness-function-type checks already exist in `lint-go` and `lint-gotest` (circular_deps, cgo_free_sqlite, crypto_rand, product_structure, bind_address_safety, parallel_tests, etc.). Should we migrate these into `lint-fitness` or leave them?

**A)** Full migration: Move ALL architecture-enforcement checks from lint-go/lint-gotest into lint-fitness. lint-go/lint-gotest keep only Go language quality checks (formatting, unused vars, etc.). Clean separation of concerns.
**B)** Dual home: Leave existing checks in lint-go/lint-gotest (backward compat). Create NEW fitness checks in lint-fitness only. Over time, migrate during future cleanup phases.
**C)** Reference only: lint-fitness runs existing checks by calling into lint-go/lint-gotest (shared code, single implementation). lint-fitness adds new checks that don't fit lint-go/lint-gotest taxonomy.
**D)** New only: lint-fitness contains ONLY checks that don't exist yet. Existing lint-go/lint-gotest checks stay forever. Avoid churn.
**E)**

**Answer**:

**Rationale**: Moving checks incurs migration risk and could break existing pre-commit hooks. But having architecture checks split across 3 commands (lint-go, lint-gotest, lint-fitness) is confusing long-term.

---

## Question 4: Shared Test Infrastructure — Migration Scope

**Question**: How many services should be migrated to the shared test infrastructure in this plan iteration?

**A)** All 10: Migrate every service's TestMain and test helpers to shared packages. Maximum consistency from day one. Higher risk of regressions.
**B)** Core 3: Migrate sm-im, jose-ja, skeleton-template (the most active/tested services). Document migration path for remaining 7. Lower risk.
**C)** Template + 1: Migrate skeleton-template first (reference implementation), then one real service (sm-im or jose-ja). Validate the approach before wider rollout.
**D)** Create only: Build the shared test packages and write their own tests, but don't migrate any existing service. Services adopt voluntarily in future work. Zero regression risk.
**E)**

**Answer**:

**Rationale**: Migration scope affects risk vs. immediate benefit. Migrating all 10 services is highest ROI but also highest regression risk. User preference for "quality over speed" suggests a more careful approach may be warranted.

---

## Question 5: Contract Test Depth

**Question**: How deep should the cross-service contract tests go in this first iteration?

**A)** Shallow (infrastructure only): Health endpoints, error format, trace_id, content-type headers. Tests verify framework behavior without touching domain logic. 6-8 contract tests.
**B)** Medium (infrastructure + auth): Shallow tests PLUS authentication rejection (401 on unauthenticated), CORS behavior, CSRF protection on /browser/**. 12-15 contract tests.
**C)** Deep (infrastructure + auth + domain patterns): Medium tests PLUS CRUD pattern testing (POST returns 201, GET returns 200, DELETE returns 204, pagination format). 20+ contract tests.
**D)** Progressive: Start with shallow (Phase 6), add medium/deep contracts as separate tasks in Phase 7 or a future plan. Build the framework first, depth second.
**E)**

**Answer**:

**Rationale**: Deeper contracts catch more divergence but require more setup. The contract test framework itself is the main deliverable; depth can be added incrementally.
