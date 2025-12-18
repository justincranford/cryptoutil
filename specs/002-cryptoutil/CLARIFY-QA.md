# Requirement Validation Questions - 002-cryptoutil

**Date**: December 18, 2025
**Context**: Validation questions for UNANSWERED design decisions, NOT questions with known answers
**Purpose**: Ensure shared understanding before implementation decisions are made
**Format**: Open-ended questions requiring analysis and decision-making

---

## IMPORTANT: What Belongs in This File

**INCLUDE** (UNANSWERED questions requiring decisions):

- "How should we handle X edge case?" (when not specified in spec)
- "What trade-offs should we consider between A and B?"
- "Which approach is better for Y requirement?"
- "How do we prioritize conflicting requirements?"

**EXCLUDE** (questions with known answers):

- ❌ "What is the target test execution time?" → ANSWER IN clarify.md Q1.1 (≤12s)
- ❌ "Which probability constant for base algorithms?" → ANSWER IN clarify.md Q1.2 (TestProbAlways)
- ❌ "What coverage target for production code?" → ANSWER IN clarify.md Q2.1 (95%+)
- ❌ "How do we identify RED lines?" → ANSWER IN clarify.md Q2.4 (go tool cover -html)

**If a question can be answered by reading clarify.md, analyze.md, or PLAN.md, DO NOT include it here.**

---

## Phase 1: Test Performance Optimization - UNANSWERED Questions

### Q1.1: Probabilistic Execution Tuning Strategy

**Context**: kms/client uses probabilistic execution (TestProbQuarter/Tenth) but baseline timing unknown for 002-cryptoutil.

**Question**: If a package with probabilistic execution STILL exceeds 12s, how should we tune probabilities further?

**Options**:

A) Reduce probabilities (e.g., Quarter → Tenth, Tenth → Never)
B) Increase base algorithm coverage (Always → Quarter) to reduce total test count
C) Consolidate test cases (merge similar variant tests)
D) Use dynamic probability (run Always locally, Tenth in CI/CD)

**Decision Needed**: Which tuning strategy is most effective without compromising coverage?

### Q1.2: Server Startup Overhead Reduction

**Context**: jose/server, kms/server/application have HTTP handler and TLS handshake overhead.

**Question**: How should we optimize server startup overhead in tests?

**Options**:

A) Share single server instance across all tests (TestMain pattern)
B) Use sync.Once per test package (setup once, reuse)
C) Mock HTTP handlers (no real server, httptest.ResponseRecorder)
D) Reduce server test count (fewer server lifecycle tests)

**Decision Needed**: Which approach balances realistic testing with ≤12s target?

### Q1.3: PostgreSQL Container Test Overhead

**Context**: kms/server/application tests use PostgreSQL containers (slow startup, migration overhead).

**Question**: How should we handle PostgreSQL container overhead in timing-sensitive tests?

**Options**:

A) Use SQLite in-memory for timing-sensitive tests, PostgreSQL for integration tests
B) Share single PostgreSQL container across all test packages (TestMain pattern)
C) Pre-warm container with migrations before test execution
D) Use testcontainers library for faster container reuse

**Decision Needed**: Which strategy achieves ≤12s without losing database-specific test coverage?

---

## Phase 2: Coverage Targets - UNANSWERED Questions

### Q2.1: HTTP Handler Testing Strategy

**Context**: jose/server (62.1%), identity/authz (66.8%) have low coverage due to HTTP handlers.

**Question**: What testing strategy achieves 95%+ coverage for HTTP handlers?

**Options**:

A) Use httptest.ResponseRecorder for all handler tests (no real server)
B) Use testify/mock for request/response objects (full mocking)
C) Use integration tests with real server + HTTP client (realistic but slow)
D) Hybrid: Unit test handler logic, integration test middleware/routing

**Decision Needed**: Which approach achieves 95%+ without exceeding ≤12s timing target?

### Q2.2: Business Logic Coverage for Zero-Coverage Functions

**Context**: kms/businesslogic has 18 core operations at 0% coverage (AddElasticKey, Get*, Post*, Update, Delete, Import, Revoke).

**Question**: How should we test business logic functions with complex dependencies (database, crypto, barrier services)?

**Options**:

A) Mock all dependencies (database, crypto, barrier) using testify/mock
B) Use real dependencies with test fixtures (in-memory database, test keys)
C) Use integration tests with Docker Compose (realistic but slow)
D) Extract business logic to pure functions, test separately from infrastructure

**Decision Needed**: Which strategy achieves 95%+ for complex business logic efficiently?

### Q2.3: Coverage Gaps for Error Paths

**Context**: Many packages have covered happy paths but uncovered error paths (e.g., "connection refused", "invalid input").

**Question**: How should we systematically test error paths to reach 95%+?

**Options**:

A) Table-driven sad path tests for each function (explicit error scenarios)
B) Fault injection (simulate failures: database down, network errors, invalid crypto keys)
C) Property-based testing with gopter (generate invalid inputs)
D) Mutation testing to identify missing error assertions

**Decision Needed**: Which error path testing strategy is most comprehensive and maintainable?

---

## Phase 3: CI/CD Workflow Fixes - UNANSWERED Questions

### Q3.1: Mutation Testing Timeout Strategy

**Context**: ci-mutation workflow times out after 45 minutes (gremlins too slow for full codebase).

**Question**: How should we restructure mutation testing to finish in <20 minutes?

**Options**:

A) Run gremlins only on business logic packages (exclude tests, generated code, mocks)
B) Use GitHub Actions matrix strategy (parallelize packages into 4-6 jobs)
C) Set per-package gremlins timeout (fail fast for slow packages)
D) Run mutation testing nightly (not on every PR/push)

**Decision Needed**: Which strategy achieves 98%+ efficacy without CI/CD bottleneck?

### Q3.2: DAST Readyz Timeout Root Cause

**Context**: ci-dast workflow fails at /admin/v1/readyz timeout (service not ready in time).

**Question**: What is root cause of readyz timeout, and how should we fix it?

**Options**:

A) Increase timeout from 30s to 60s (GitHub Actions latency)
B) Optimize service startup (parallelize unseal, cache migrations)
C) Add retry logic with exponential backoff (handle transient failures)
D) Add diagnostic logging to identify startup bottleneck

**Decision Needed**: Which fix addresses root cause vs symptom?

### Q3.3: OpenTelemetry Collector Healthcheck Failure

**Context**: ci-fuzz and ci-load workflows fail at opentelemetry-collector-contrib healthcheck.

**Question**: Why does otel-collector healthcheck fail, and how should we fix it?

**Options**:

A) Increase healthcheck start_period (collector needs more startup time)
B) Fix healthcheck command (current command incorrect for otel-collector)
C) Add diagnostic logging to collector startup
D) Use sidecar health check (separate Alpine container with wget)

**Decision Needed**: Which approach reliably detects collector readiness?

---

## Phase 4: Mutation Testing Quality Assurance - UNANSWERED Questions

### Q4.1: Mutation Testing Prioritization Strategy

**Context**: 107 packages, ~200-300 functions total, mutation testing all would take days.

**Question**: How should we prioritize packages for mutation testing to maximize ROI?

**Options**:

A) Start with highest-risk packages (crypto, auth, business logic)
B) Start with lowest-coverage packages (identity/authz 66.8%, kms/businesslogic 39.0%)
C) Start with most-changed packages (git log --stat, identify hotspots)
D) Run all packages but set 80% efficacy target (not 98%)

**Decision Needed**: Which prioritization maximizes bug detection per hour invested?

### Q4.2: Lived Mutant Analysis Strategy

**Context**: Gremlins reports "lived mutants" (mutants that didn't fail tests).

**Question**: How should we systematically kill lived mutants to reach 98%+ efficacy?

**Options**:

A) Analyze lived mutant diff, write test targeting specific mutation
B) Use mutation testing report to identify weak test assertions
C) Property-based testing (gopter) to generate edge cases
D) Focus on boundary conditions and off-by-one errors

**Decision Needed**: Which lived mutant killing strategy is most efficient?

---

## Phase 5: Refactor Hashes - UNANSWERED Questions

### Q5.1: Hash Version Selection Algorithm

**Context**: Hash registries use version selection based on input size (0-31 bytes → v1 SHA256, 32-47 → v2 SHA384, 48+ → v3 SHA512).

**Question**: Should version selection be input-size-based or explicit version parameter?

**Options**:

A) Input-size-based (automatic version selection, simpler API)
B) Explicit version parameter (caller specifies, more control)
C) Hybrid (default to input-size-based, allow override)
D) Configuration-driven (version selection policy in config file)

**Decision Needed**: Which approach balances API simplicity with flexibility?

### Q5.2: Hash Output Format

**Context**: Hash output format must include version metadata for Verify() to work.

**Question**: What hash output format supports version-aware verification?

**Options**:

A) Prefix format: `{v}:<base64_hash>` (e.g., `{1}:abcd1234...`)
B) JSON format: `{"v":1,"hash":"abcd1234..."}`
C) Binary format: `[1 byte version][32 bytes hash]`
D) PHC string format: `$pbkdf2-sha256$v=1$rounds=...`

**Decision Needed**: Which format is most compact, human-readable, and future-proof?

### Q5.3: Hash Registry Migration Strategy

**Context**: Existing code uses PBKDF2 directly, must migrate to versioned registry.

**Question**: How should we migrate existing hash usage to new registry?

**Options**:

A) Big bang migration (replace all at once, test extensively)
B) Phased migration (add registry, deprecate direct PBKDF2, then remove)
C) Dual-support (registry for new code, legacy for existing)
D) Automated refactoring (tool to rewrite call sites)

**Decision Needed**: Which migration strategy minimizes risk and downtime?

---

## Phase 6: Server Architecture Unification - UNANSWERED Questions

### Q6.1: Template Parameterization Strategy

**Context**: 8 PRODUCT-SERVICE instances share infrastructure but differ in business logic.

**Question**: How should ServerTemplate be parameterized for service-specific customization?

**Options**:

A) Constructor injection (pass handlers, middleware, config at init time)
B) Interface-based customization (services implement ServerInterface)
C) Configuration-driven (YAML config specifies handlers, middleware)
D) Plugin architecture (services register plugins with template)

**Decision Needed**: Which parameterization approach balances flexibility with simplicity?

### Q6.2: Barrier Services Inclusion Strategy

**Context**: KMS uses barrier services (unseal, root, intermediate, content), but other services don't need them.

**Question**: How should template handle optional barrier services?

**Options**:

A) Make barrier services optional with nil checks (all services use same template)
B) Provide two templates (barrier-enabled, barrier-free)
C) Use feature flags (enable/disable barrier services via config)
D) Extract barrier services to separate package (not part of template)

**Decision Needed**: Which approach minimizes template complexity?

### Q6.3: Client SDK Generation Strategy

**Context**: Client SDKs must be generated from OpenAPI specs for all services.

**Question**: How should we automate client SDK generation?

**Options**:

A) Manual oapi-codegen runs (developer responsibility)
B) go:generate directives (auto-generate on `go generate`)
C) pre-commit hook (generate SDKs before commit)
D) CI/CD workflow (generate and commit SDKs automatically)

**Decision Needed**: Which generation strategy ensures SDK consistency without manual overhead?

---

## Phase 7: Learn-PS Demonstration Service - UNANSWERED Questions

### Q7.1: Pet Store API Scope

**Context**: Learn-PS is Pet Store example using service template.

**Question**: What API scope makes Learn-PS useful as copy-paste-modify starting point?

**Options**:

A) Minimal (CRUD only: Create/Read/Update/Delete pets)
B) Moderate (CRUD + pagination, filtering, sorting)
C) Comprehensive (CRUD + search, inventory management, order processing)
D) Realistic (full e-commerce: customers, orders, payments, inventory)

**Decision Needed**: Which scope demonstrates template capabilities without overwhelming users?

### Q7.2: Learn-PS Authentication Strategy

**Context**: Learn-PS demonstrates OAuth 2.1 integration, but Pet Store is tutorial example.

**Question**: Should Learn-PS require authentication for all operations?

**Options**:

A) Public API (no authentication, simplest demo)
B) Optional authentication (some endpoints public, some protected)
C) Required authentication (OAuth 2.1 flows mandatory, realistic)
D) Hybrid (public read, protected write, demonstrates both patterns)

**Decision Needed**: Which authentication pattern best demonstrates template without complicating tutorial?

### Q7.3: Learn-PS Database Strategy

**Context**: Learn-PS demonstrates dual-database support (PostgreSQL + SQLite).

**Question**: Should Learn-PS tutorials use PostgreSQL, SQLite, or both?

**Options**:

A) SQLite only (simplest for tutorials, no container setup)
B) PostgreSQL only (realistic for production, requires Docker)
C) Both (SQLite for local dev, PostgreSQL for Docker/production)
D) Configurable (users choose based on environment)

**Decision Needed**: Which database strategy reduces tutorial friction while demonstrating template flexibility?

---

## Cross-Cutting Concerns - UNANSWERED Questions

### Q8.1: Windows Firewall Exception Prevention

**Context**: Binding to 0.0.0.0 triggers Windows Firewall prompts, binding to 127.0.0.1 does not.

**Question**: How should we balance Windows dev experience with Docker deployment requirements?

**Options**:

A) Always use 127.0.0.1 for tests (never 0.0.0.0, no firewall prompts)
B) Use 127.0.0.1 for unit tests, 0.0.0.0 for integration tests (Docker only)
C) Document firewall exception process (users must allow manually)
D) Use environment variable (BIND_ADDRESS defaults to 127.0.0.1, override in Docker)

**Decision Needed**: Which approach minimizes dev friction while supporting Docker deployment?

### Q8.2: Coverage Baseline Tracking Strategy

**Context**: Coverage must be tracked per package to prevent regressions.

**Question**: How should we track coverage baselines to detect regressions?

**Options**:

A) Git-tracked baseline files (test-output/coverage_*.out committed)
B) CI/CD artifacts (upload/download baseline between runs)
C) Coverage service (Codecov, Coveralls, centralized tracking)
D) Pre-commit hook (compare against previous coverage, fail if drops)

**Decision Needed**: Which tracking strategy catches regressions earliest without CI/CD overhead?

### Q8.3: Gremlins Efficacy Baseline Tracking

**Context**: Mutation testing efficacy must be tracked per package (98%+ target).

**Question**: How should we track gremlins efficacy baselines?

**Options**:

A) Git-tracked baseline files (docs/GREMLINS-TRACKING.md committed)
B) CI/CD artifacts (upload/download gremlins reports)
C) Dedicated tracking doc (per-package efficacy table updated after each run)
D) Pre-commit hook (compare against previous efficacy, fail if drops)

**Decision Needed**: Which tracking strategy ensures efficacy improvements are maintained?

---

## Decision Log Format

When answering these questions, document decisions in PLAN.md or clarify.md using this format:

```markdown
### Decision: <Question Number> <Question Title>

**Date**: YYYY-MM-DD
**Decision**: <Selected Option> (<A/B/C/D>)
**Rationale**: <Why this option was chosen>
**Trade-offs**: <What we gave up by not choosing other options>
**Implementation**: <How to implement this decision>
**Validation**: <How to verify decision was correct>
```

---

**REMINDER**: This file contains ONLY UNANSWERED questions. If a question can be answered by reading existing docs (clarify.md, analyze.md, PLAN.md), it does NOT belong here.
