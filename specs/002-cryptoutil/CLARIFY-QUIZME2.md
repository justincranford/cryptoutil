# Requirement Validation Questions Round 2 - 002-cryptoutil

**Date**: December 19, 2025
**Context**: Second round of clarification questions after SPECKIT-CONFLICTS-ANALYSIS.md answers
**Purpose**: New ambiguities, gaps, and conflicts discovered after first clarification round
**Format**: Multiple choice A-D with E write-in for quick resolution

---

## IMPORTANT: Scope of This Document

**INCLUDE** (NEW questions from Round 2 analysis):

- New ambiguities discovered after applying SPECKIT-CONFLICTS-ANALYSIS answers
- Implementation details requiring decisions (not covered in Round 1)
- Trade-offs between conflicting requirements
- Edge cases not addressed in constitution.md or spec.md

**EXCLUDE** (questions already answered):

- ❌ Questions from SPECKIT-CONFLICTS-ANALYSIS.md (all 26 answered 2025-12-19)
- ❌ Questions with answers in clarify.md
- ❌ Questions with answers in constitution.md Section VIII (Spec Kit workflow)

---

## Architecture and Service Integration

### A2.1: Package Classification Details (Follow-up to A2)

**Context**: User answered "E" but didn't provide classification details

**Question**: Which specific packages belong to each coverage target category?

**A)** Production (95%): internal/{jose,identity,kms,ca}; Infrastructure (100%): internal/cmd/cicd/*; Utility (100%): internal/shared/*, pkg/*
**B)** Production (95%): internal/{jose,identity,kms,ca}, internal/infra/*; Infrastructure (100%): cmd/*, internal/cmd/*; Utility (100%): pkg/*
**C)** All internal/* packages are production (95%); cmd/*and pkg/* are utility (100%)
**D)** Case-by-case per package (document each in clarify.md)
**E)** Other: _______________

**Your Answer**: ___

---

### A4.1: Federation Configuration File Location

**Context**: User chose static YAML configuration (A4 answer: A)

**Question**: Where should federation configuration be stored?

**A)** Each service has own federation section (kms.yml has federation.identity_url, federation.jose_url)
**B)** Shared federation.yml file (all services read same file for service discovery)
**C)** Environment-specific configs (federation-dev.yml, federation-prod.yml)
**D)** Docker Compose environment variables (FEDERATION_IDENTITY_URL, etc.)
**E)** Other: _______________

**Your Answer**: ___

---

### A4.2: Federation Failure Handling

**Context**: Static federation configuration (A4 answer: A)

**Question**: How should services handle federated service unavailability?

**A)** Fail fast at startup (can't start if Identity/JOSE unreachable)
**B)** Graceful degradation (start but disable federated features)
**C)** Retry with exponential backoff (keep trying to reach federated service)
**D)** Circuit breaker pattern (fail after N attempts, retry periodically)
**E)** Other: _______________

**Your Answer**: ___

---

### C4.1: Admin Port Collision in Unified Deployment

**Context**: User chose unique admin ports per service (C4 answer: D - 9090/9091/9092/9093)

**Question**: How do we prevent admin port collisions when multiple instances of same service run?

**A)** Admin ports bound to 127.0.0.1 only (not externally accessible, safe for collisions)
**B)** Dynamic admin ports (port 0, log actual port at startup)
**C)** Instance-specific admin ports (kms-1: 9090, kms-2: 9094, kms-3: 9095)
**D)** Single admin port per Docker Compose (only first instance has admin endpoint)
**E)** Other: _______________

**Your Answer**: ___

---

### C7.1: CA Database Schema Differences

**Context**: User chose 3-instance CA deployment matching KMS/JOSE (C7 answer: A)

**Question**: Does CA have different database schema requirements than KMS/JOSE?

**A)** Yes, CA schema is significantly different (certificates, CRLs, OCSP) - needs custom migrations
**B)** Partially, CA shares some tables (audit, config) but has unique certificate tables
**C)** No, CA follows same repository patterns as KMS/JOSE (minimal schema differences)
**D)** Unknown, requires Phase 4 CA implementation analysis
**E)** Other: _______________

**Your Answer**: ___

---

## Testing and Quality Assurance

### C2.1: Mutation Score Transition Between Phases

**Context**: User chose phased targets (C2 answer: E - "85% Phase 4, 98% Phase 5+")

**Question**: When exactly does mutation score requirement change from 85% to 98%?

**A)** At start of Phase 5 (all packages must reach 98% before Phase 5 work begins)
**B)** Gradually during Phase 5 (new code 98%, existing code can stay 85% until refactored)
**C)** Per-package decision (packages touched in Phase 5+ must reach 98%)
**D)** End of Phase 7 (98% is final goal, not intermediate requirement)
**E)** Other: _______________

**Your Answer**: ___

---

### C3.1: Test Timing Enforcement Mechanism

**Context**: User set strict timing targets (C3 answer: E - "<15s per package, <180s total")

**Question**: How should we enforce test timing targets in CI/CD?

**A)** Fail CI/CD build if any package exceeds 15s (hard enforcement)
**B)** Warning only (log slow packages but don't fail build)
**C)** Fail only if total suite exceeds 180s (per-package timing informational)
**D)** Adaptive thresholds (15s target, 30s hard limit per package)
**E)** Other: _______________

**Your Answer**: ___

---

### C3.2: Integration Test Timing Targets

**Context**: User excluded integration/e2e from strict timing (C3 answer: E)

**Question**: Should we set any timing targets for integration/e2e tests?

**A)** No timing targets (Docker startup overhead unpredictable)
**B)** Soft targets only (<5 minutes per integration package, <15 minutes total)
**C)** Hard targets with generous margins (<10 minutes per package, <30 minutes total)
**D)** Per-service targets (KMS: 3min, Identity: 5min, CA: 4min, JOSE: 2min)
**E)** Other: _______________

**Your Answer**: ___

---

### Q1.1.1: Test Consolidation Impact on Coverage

**Context**: User chose consolidate test cases (Q1.1 answer: C)

**Question**: How do we ensure consolidation doesn't reduce coverage?

**A)** Require coverage unchanged before/after consolidation (strict validation)
**B)** Allow minor coverage drops (<1%) if timing improves significantly (>5s faster)
**C)** Merge only truly redundant tests (same code path, different data)
**D)** Use mutation testing to validate consolidation (mutation score unchanged)
**E)** Other: _______________

**Your Answer**: ___

---

### Q1.2.1: TestMain Pattern for Multiple Services

**Context**: User chose TestMain pattern for server sharing (Q1.2 answer: A)

**Question**: How should TestMain handle packages testing multiple services?

**A)** Start all required services in TestMain (e.g., KMS + Identity + JOSE)
**B)** Use Docker Compose in TestMain (full multi-service stack)
**C)** Split tests into separate packages (one package per service)
**D)** Use test fixtures (mock other services)
**E)** Other: _______________

**Your Answer**: ___

---

### Q2.1.1: Integration Test vs httptest Trade-off

**Context**: User strongly prefers real servers (Q2.1 answer: "ALWAYS C")

**Question**: At what coverage threshold should we add httptest mocks for corner cases?

**A)** Never use httptest (always real servers, no exceptions)
**B)** Use httptest only if real server can't reach branch (<95% coverage impossible)
**C)** Use httptest for error injection (network failures, timeout simulation)
**D)** Use httptest for security testing (malformed requests, boundary conditions)
**E)** Other: _______________

**Your Answer**: ___

---

## Cryptography and Hash Service

### Q5.1.1: Hash Version Selection Triggers

**Context**: User chose config-driven with date-based versions (Q5.1 answer: E)

**Question**: What triggers hash version updates in production?

**A)** Manual operator decision (update config, restart service, new hashes use v2)
**B)** Automatic on deployment (config includes version, deployment updates version)
**C)** Gradual rollout (new tenants use v2, existing tenants stay on v1)
**D)** Per-API basis (password hashing v1, PII hashing v2)
**E)** Other: _______________

**Your Answer**: ___

---

### Q5.1.2: Hash Version Verification During Migration

**Context**: Date-based policy revisions (v1=2020, v2=2023, v3=2025)

**Question**: How do we verify old hashes during version migration?

**A)** Verification automatically tries all versions (v1, v2, v3) until match found
**B)** Hash output includes version prefix ({1}:hash, {2}:hash) - verify using correct version
**C)** Database stores version alongside hash - lookup version before verification
**D)** Version inference from hash length/format (SHA-256 output = v1, SHA-512 = v3)
**E)** Other: _______________

**Your Answer**: ___

---

### Q5.2.1: Hash Output Format Backward Compatibility

**Context**: User chose prefix format (Q5.2 answer: A - "{v}:base64_hash")

**Question**: How do we handle existing hashes without version prefix?

**A)** Assume all unprefixed hashes are v1 (default version)
**B)** Require migration script to add version prefix to all existing hashes
**C)** Support both formats (try unprefixed as v1, prefixed with explicit version)
**D)** Reject unprefixed hashes (force re-hash on next authentication)
**E)** Other: _______________

**Your Answer**: ___

---

## Service Template and Reusability

### Q6.1.1: Template Initialization Pattern

**Context**: User chose constructor injection (Q6.1 answer: A)

**Question**: Should service template use builder pattern or direct constructor?

**A)** Builder pattern (ServiceTemplate.New().WithHandlers(...).WithMiddleware(...).Build())
**B)** Direct constructor (NewServiceTemplate(handlers, middleware, config))
**C)** Functional options (NewServiceTemplate(handlers, WithMiddleware(...), WithConfig(...)))
**D)** Configuration struct (NewServiceTemplate(&ServiceConfig{Handlers: ..., Middleware: ...}))
**E)** Other: _______________

**Your Answer**: ___

---

### Q6.1.2: Template Customization Points

**Context**: Service template must support 8 different services (sm-kms, pki-ca, jose-ja, identity-*)

**Question**: Which customization points should template expose?

**A)** Handlers only (template handles middleware, telemetry, config automatically)
**B)** Handlers + middleware (template handles telemetry, config)
**C)** Handlers + middleware + telemetry (template handles config only)
**D)** All aspects customizable (fully parameterized template)
**E)** Other: _______________

**Your Answer**: ___

---

### Q6.3.1: SDK Generation Automation Timing

**Context**: User chose go:generate directives (Q6.3 answer: "B, but a user or LLM agent developer can do A during development too")

**Question**: When should go:generate run for SDK generation?

**A)** Manually during development (developer runs `go generate ./...` when OpenAPI changes)
**B)** Pre-commit hook (auto-generate SDKs before every commit)
**C)** CI/CD workflow (generate and commit SDKs automatically on OpenAPI changes)
**D)** On OpenAPI file modification (file watcher triggers go generate)
**E)** Other: _______________

**Your Answer**: ___

---

## Observability and Telemetry

### Q3.2.1: Diagnostic Logging Implementation

**Context**: User chose diagnostic logging for DAST readyz timeout (Q3.2 answer: D)

**Question**: What level of diagnostic detail should we add?

**A)** Timestamps only (startup phase timing: TLS 2.1s, DB 5.3s, unseal 1.8s)
**B)** Timestamps + component names (2025-12-19 10:15:30 [TLS] Server started)
**C)** Structured logging with phases (JSON: {"phase":"tls","duration_ms":2100,"status":"complete"})
**D)** OpenTelemetry traces (span per startup phase with timing and metadata)
**E)** Other: _______________

**Your Answer**: ___

---

### Q3.3.1: Otel Collector Sidecar Healthcheck Details

**Context**: User confirmed sidecar health check is only known working solution (Q3.3 answer: "D IS ONLY SOLUTION")

**Question**: Should we investigate otel-collector internal health endpoints?

**A)** Yes, investigate /healthz or /metrics endpoints (may exist but undocumented)
**B)** Yes, but keep sidecar as fallback (belt-and-suspenders approach)
**C)** No, sidecar works reliably (don't fix what isn't broken)
**D)** Yes, but Phase 7 only (not critical for MVP)
**E)** Other: _______________

**Your Answer**: ___

---

## Deployment and Docker

### Docker Compose Instance Naming

**Question**: How should we name multiple instances in Docker Compose?

**A)** Product-backend-number (kms-sqlite, kms-postgres-1, kms-postgres-2)
**B)** Product-instance-number (kms-1, kms-2, kms-3) with backend in config
**C)** Product-role (kms-primary, kms-replica-1, kms-replica-2)
**D)** Current naming is fine (cryptoutil-sqlite, cryptoutil-postgres-1, ...)
**E)** Other: _______________

**Your Answer**: ___

---

### Docker Compose Health Check Start Period

**Context**: Current start_period=30s may be insufficient for full stack

**Question**: Should we increase health check start_period based on diagnostic logging results?

**A)** Keep 30s (services should optimize startup to meet target)
**B)** Increase to 45s (realistic for TLS + migrations + unseal)
**C)** Increase to 60s (conservative, prevents flaky health checks)
**D)** Dynamic per service (KMS: 30s, Identity: 45s, CA: 60s)
**E)** Other: _______________

**Your Answer**: ___

---

## CI/CD and Automation

### Gremlins Windows Compatibility Investigation

**Context**: User wants gremlins working on Windows (A6 answer: E)

**Question**: What level of effort should we invest in Windows compatibility?

**A)** High priority - Block Phase 4 until Windows gremlins works (developer experience critical)
**B)** Medium priority - Investigate in Phase 4, document findings, keep CI/CD workaround
**C)** Low priority - Use CI/CD only, Windows investigation in Phase 7 if time permits
**D)** Community contribution - Document issue, submit gremlins bug report, wait for upstream fix
**E)** Other: _______________

**Your Answer**: ___

---

### Coverage Baseline Artifact Retention

**Context**: User chose CI/CD artifacts for baseline tracking (Q8.2 answer: B)

**Question**: How long should we retain coverage baseline artifacts?

**A)** 30 days (GitHub default, sufficient for PR review)
**B)** 90 days (longer trend analysis)
**C)** Indefinitely (download and commit selected baselines to git)
**D)** Per-release (tag baseline artifacts with release versions)
**E)** Other: _______________

**Your Answer**: ___

---

## Documentation and Communication

### Clarify.md Organization Strategy

**Question**: How should we organize clarify.md as it grows with Round 2 answers?

**A)** Chronological (add Round 2 answers at end, maintain timeline)
**B)** Topical reorganization (merge Round 1 + Round 2 into unified topic sections)
**C)** Separate files (clarify-round1.md, clarify-round2.md)
**D)** Index by question number (Q1.1, Q1.2, Q2.1, Q2.2, ...) for easy cross-reference
**E)** Other: _______________

**Your Answer**: ___

---

### Spec Kit Feedback Loop Timing

**Question**: How frequently should we update constitution/spec during implementation?

**A)** After every Phase completion (7 update cycles)
**B)** After every 3-5 tasks (mini-cycle pattern from SPECKIT-REFINEMENT-GUIDE)
**C)** When implementation insights contradict spec (as-needed basis)
**D)** End of iteration only (constitution v3.0.0, v4.0.0, ...)
**E)** Other: _______________

**Your Answer**: ___

---

## Summary

**Total Questions**: 28 new clarifications
**Architecture**: 6 questions
**Testing**: 6 questions
**Cryptography**: 4 questions
**Service Template**: 3 questions
**Observability**: 2 questions
**Deployment**: 2 questions
**CI/CD**: 2 questions
**Documentation**: 2 questions
**Process**: 1 question

**Next Steps**:

1. Answer all questions with A/B/C/D or E write-in
2. Process answers and update clarify.md
3. Update constitution.md/spec.md/plan.md as needed
4. Update copilot instructions with new patterns
5. Mark questions as [ANSWERED] with date

---

## Best Practices Reminder

**From SPECKIT-CONFLICTS-ANALYSIS.md end section**:

Answered clarifications should be persisted in:

- **clarify.md**: ALWAYS (authoritative Q&A record)
- **constitution.md**: If establishes fundamental principle/constraint
- **spec.md**: If adds product/service requirement
- **plan.md**: If affects implementation approach
- **copilot instructions**: If affects LLM agent behavior

Each update should cite source: "Source: CLARIFY-QUIZME2.md answered YYYY-MM-DD"
