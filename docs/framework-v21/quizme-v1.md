# Quizme - Framework v21 Phase 2 API Design Decisions (Round 1)

Purpose: Resolve user-choice design decisions required to complete Task 2.2 and unblock Phases 3-8.

## Question 1: Default Fixture Scope Model for test_orch_integration

**Question**: What should be the default fixture lifecycle for integration TestMain orchestration?

**A)** Per-suite shared fixture by default, with opt-in per-test isolation.
**B)** Per-test isolated fixture by default, with opt-in shared fixture.
**C)** Hybrid default: shared DB + per-test app instance.
**D)** Hybrid default: shared app instance + per-test DB namespace.
**E)** I HAVE NO IDEAL OR CONTEXT OF WHAT YOU ARE ASKING. PLEASE EXPLAIN IN MORE DETAIL OR PROVIDE EXAMPLES!

**Answer**: E

**Rationale**: This decides deterministic cleanup semantics, speed vs isolation, and API defaults for all integration migrations.

## Question 2: Error-Path Fixture Creation Contract

**Question**: Which mechanism should be standard for DB/API failure-path setup in test_orch_integration?

**A)** Explicit factory APIs returning pre-broken fixtures (e.g., closed DB, invalid TLS, bad DSN).
**B)** Generic hook injection callbacks that mutate a valid fixture before startup.
**C)** Table-driven failure profile enum + framework-provided builders.
**D)** Direct suite-managed setup (no framework-level error fixture contract).
**E)** I HAVE NO IDEAL OR CONTEXT OF WHAT YOU ARE ASKING. PLEASE EXPLAIN IN MORE DETAIL OR PROVIDE EXAMPLES!

**Answer**: E

**Rationale**: Task 2.2 requires a clear mechanism for repeatable error-path tests without ad hoc suite code.

## Question 3: Readiness Endpoint Contract for Integration Orchestration

**Question**: What should the default readiness probe contract be for direct-start integration servers?

**A)** Require admin readyz only (`/admin/api/v1/readyz`) with optional extra probes.
**B)** Require admin readyz + public browser/service health probes before returning ready.
**C)** Require configurable probe list, with no fixed defaults.
**D)** Skip probes and rely on startup return + bounded sleep policy.
**E)** I HAVE NO IDEAL OR CONTEXT OF WHAT YOU ARE ASKING. PLEASE EXPLAIN IN MORE DETAIL OR PROVIDE EXAMPLES!

**Answer**: E

**Rationale**: This drives startup correctness, flake resistance, and consistency across 37 remaining migrations.

## Question 4: Port Allocation and Concurrency Safety Contract

**Question**: Which port policy should be mandatory in test_orch_integration API?

**A)** Always bind both listeners to port 0; expose resolved URLs via returned runtime handle.
**B)** Allow explicit fixed ports only when caller requests; default port 0 otherwise.
**C)** Reserve deterministic per-package port ranges to simplify debug reproducibility.
**D)** Keep existing per-suite behavior and document conflict retries.
**E)** I HAVE NO IDEAL OR CONTEXT OF WHAT YOU ARE ASKING. PLEASE EXPLAIN IN MORE DETAIL OR PROVIDE EXAMPLES!

**Answer**: E

**Rationale**: This controls concurrency reliability and Windows TIME_WAIT behavior during parallel tests.

## Question 5: Migration Compatibility Strategy

**Question**: How should existing helper packages transition to the new orchestration APIs?

**A)** Keep compatibility wrappers for one release cycle; new code must use orchestration packages directly.
**B)** Keep wrappers indefinitely for stability and low migration churn.
**C)** Remove wrappers immediately; require direct migration in one pass.
**D)** Keep wrappers only for PS-ID templates, not framework/internal callers.
**E)**

**Answer**: C

**Rationale**: This determines migration pacing, deprecation policy, and implementation sequencing in Phases 3-7.
