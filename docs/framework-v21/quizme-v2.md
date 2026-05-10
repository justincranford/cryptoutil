# Quizme - Framework v21 Phase 2 API Design Decisions (Round 2)

Purpose: Resolve only the remaining unresolved design decisions (Q1-Q4) with concrete context and examples.

Already resolved from Round 1:
1. Migration compatibility strategy: C (one-pass direct migration, no compatibility wrappers).

## Context Snapshot (Why these choices matter)

1. Scope: 39 in-scope TestMain entries must converge on framework orchestration.
2. Remaining design blocker: Task 2.2 in [docs/framework-v21/tasks.md](docs/framework-v21/tasks.md).
3. Runtime target for integration orchestration: one direct-start PS-ID server, SQLite in-memory, dynamic dual ports, deterministic cleanup.
4. Current pain points to remove: panic-style startup helpers, ad hoc sleep/wait logic, and inconsistent DB/error fixture setup across packages.

## Question 1: Default Fixture Scope Model for test_orch_integration

Decision to make: default lifecycle for integration test fixtures.

Practical examples:
1. Per-suite shared fixture means one app+DB setup in TestMain; test cases reuse it. Fastest, but state leakage risk unless reset hooks are strict.
2. Per-test isolated fixture means each test gets fresh app+DB. Strong isolation, slower for large suites.
3. Hybrid shared DB + per-test app instance reduces app-state leakage while keeping DB setup cost lower.
4. Hybrid shared app + per-test DB namespace keeps app startup cheap but requires strong DB namespace isolation controls.

Trade-off summary:
1. Speed increases as sharing increases.
2. Isolation increases as per-test setup increases.
3. Flake risk increases when mutable shared state is not reset perfectly.

**Question**: Which default should the API enforce?

**A)** Per-suite shared fixture by default, with opt-in per-test isolation.
**B)** Per-test isolated fixture by default, with opt-in shared fixture.
**C)** Hybrid default: shared DB + per-test app instance.
**D)** Hybrid default: shared app instance + per-test DB namespace.
**E)**

**Answer**:

**Rationale**: This sets the default behavior for most of the 37 remaining migrations.

## Question 2: Error-Path Fixture Creation Contract

Decision to make: how tests create deterministic failure scenarios.

Practical examples:
1. Explicit factory APIs: `NewClosedDBFixture`, `NewBadTLSFixture`, `NewInvalidDSNFixture`.
2. Hook injection callbacks: start with valid fixture then mutate before startup.
3. Failure profile enum: `FailureClosedDB`, `FailureInvalidTLS`, etc., with framework builders.
4. Direct suite-managed setup leaves each package to craft failure fixtures independently.

Trade-off summary:
1. Explicit factory APIs are easiest to read and most deterministic.
2. Hook injection is flexible but can become hard to audit.
3. Enum profiles are consistent but require maintaining profile surface.
4. Direct suite-managed setup is fastest short-term but causes drift.

**Question**: Which mechanism should be standard?

**A)** Explicit factory APIs returning pre-broken fixtures (closed DB, invalid TLS, bad DSN).
**B)** Generic hook injection callbacks that mutate a valid fixture before startup.
**C)** Table-driven failure profile enum plus framework-provided builders.
**D)** Direct suite-managed setup (no framework-level error fixture contract).
**E)**

**Answer**:

**Rationale**: Task 2.2 requires a repeatable error-path contract across all migration targets.

## Question 3: Readiness Endpoint Contract for Integration Orchestration

Decision to make: what "ready" means before tests execute.

Practical examples:
1. Admin-ready only: wait for `/admin/api/v1/readyz`, then proceed.
2. Admin + public probes: require admin readyz plus browser/service health checks.
3. Fully caller-configurable probe list with no default.
4. No probes; rely on startup return and bounded sleep.

Trade-off summary:
1. Admin-only is simple and aligns with server readiness semantics.
2. Admin + public probes catches route/middleware miswiring earlier.
3. Fully configurable is flexible but can produce inconsistent defaults.
4. Sleep-based readiness is least reliable and increases flaky failures.

**Question**: What default readiness contract should be used?

**A)** Require admin readyz only (`/admin/api/v1/readyz`) with optional extra probes.
**B)** Require admin readyz plus public browser/service health probes before returning ready.
**C)** Require configurable probe list, with no fixed defaults.
**D)** Skip probes and rely on startup return plus bounded sleep policy.
**E)** Require admin readyz only (`/admin/api/v1/readyz`) with optional extra probes.; CHECK DOCS/ENG-HANDBOOK.MD, THIS IS ALREADY THE DEFAULT FOR INTEGRATION TESTS AND INCLUDES PROPAGATIONS TO ALL COPILOT+CLAUDE INSTRUCTIONS/AGNTS/SKILLS!!!!!!!!!!!!!!

**Answer**: E

**Rationale**: This determines startup determinism and flake resistance in integration suites.

## Question 4: Port Allocation and Concurrency Safety Contract

Decision to make: how the integration orchestrator binds ports under parallel test execution.

Practical examples:
1. Always port 0: OS assigns ephemeral ports for both listeners; API returns resolved URLs.
2. Default port 0 but allow explicit fixed ports for debugging.
3. Deterministic package ranges require central coordination and conflict handling.
4. Keep current mixed behavior and document retries.

Trade-off summary:
1. Port 0 default is strongest for parallel safety and Windows TIME_WAIT avoidance.
2. Optional explicit fixed ports can improve local debug reproducibility.
3. Fixed ranges are predictable but increase collision management overhead.
4. Retry-based behavior keeps existing risk profile.

**Question**: Which port policy should be mandatory?

**A)** Always bind both listeners to port 0; expose resolved URLs via returned runtime handle.
**B)** Allow explicit fixed ports only when caller requests; default port 0 otherwise.
**C)** Reserve deterministic per-package port ranges to simplify debug reproducibility.
**D)** Keep existing per-suite behavior and document conflict retries.
**E)** Always bind both listeners to port 0; expose resolved URLs via returned runtime handle; CHECK DOCS/ENG-HANDBOOK.MD, THIS IS ALREADY THE DEFAULT FOR INTEGRATION TESTS AND INCLUDES PROPAGATIONS TO ALL COPILOT+CLAUDE INSTRUCTIONS/AGNTS/SKILLS!!!!!!!!!!!!!!

**Answer**: E

**Rationale**: This controls parallel reliability and startup/shutdown stability across integration packages.
