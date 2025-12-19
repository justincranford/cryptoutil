# Speckit Document Conflicts Analysis - Multiple Choice Review

**Date**: December 19, 2025
**Purpose**: Quick-review format for resolving conflicts, omissions, and ambiguities
**Instructions**: Choose A, B, C, or D for each question. Use E to provide custom answer.

---

## Critical Priority Issues (Blocking Implementation)

### C2: Mutation Testing Threshold 🔥 CRITICAL

**Conflict**: constitution.md says "≥80%", clarify.md says "Minimum ≥80%, Target ≥98%", analyze.md says "98%+ required"

**Question**: What is the mutation testing requirement?

**A)** 80% minimum (BLOCKING), 98% aspirational (nice to have)
**B)** 98% minimum (BLOCKING), 80% is deprecated
**C)** 80% for Phase 4, then raise to 98% in Phase 5+
**D)** Package-dependent (80% utility, 98% crypto/auth)
**E)** Other: 85% for Phase 4, then raise to 98% in Phase 5+

**Your Answer**: E

---

### C3: Test Execution Time Targets 🔥 CRITICAL

**Conflict**: spec.md says "<30s per package", clarify.md says "≤30s target, ≤60s hard limit"

**Question**: What happens if a package takes 45 seconds?

**A)** Acceptable (between target and hard limit)
**B)** Requires optimization before proceeding
**C)** Only acceptable if probabilistic execution applied
**D)** Acceptable only for integration tests
**E)** Other: MANDATORY <15s per unit test packages (i.e. !integration and !e2e tests); overall <180s for entire !integration and !e2e tests

**Your Answer**: E

---

### O1: Probability-Based Testing Not in Constitution 🔥 CRITICAL

**Omission**: Probability-based execution documented in clarify.md but not constitution.md

**Question**: Should probability-based testing be added to constitution.md?

**A)** Yes, add to constitution.md Section IV (Go Testing Requirements)
**B)** No, keep only in clarify.md (implementation detail)
**C)** Yes, but only reference it (link to clarify.md)
**D)** Add to copilot instructions instead
**E)** Other: Add to constitution.md, clarify.md, and copilot instructions

**Your Answer**: E

---

### O2: main() Pattern Not in Constitution 🔥 CRITICAL

**Omission**: main() → internalMain() pattern documented in clarify.md but not constitution.md

**Question**: Should main() testability pattern be added to constitution.md?

**A)** Yes, add to constitution.md Section IV (Go Testing Requirements)
**B)** No, keep only in clarify.md (implementation detail)
**C)** Yes, add to both constitution.md and copilot instructions
**D)** Add to copilot instructions only
**E)** Other: Add to constitution.md, clarify.md, and copilot instructions

**Your Answer**: E

---

### O3: Windows Firewall Prevention Not in Constitution 🔥 CRITICAL

**Omission**: 127.0.0.1 binding requirement documented in clarify.md but not constitution.md

**Question**: Should Windows firewall prevention be added to constitution.md?

**A)** Yes, add to constitution.md Section V (Service Architecture)
**B)** No, keep only in clarify.md (platform-specific detail)
**C)** Yes, add to both constitution.md and copilot instructions
**D)** Add to copilot instructions only (already exists there)
**E)** Other: Add to constitution.md, clarify.md, and copilot instructions

**Your Answer**: E

---

## High Priority Issues (Architecture Impact)

### C4: CA Admin Port Inconsistency 🔴 HIGH

**Conflict**: spec.md shows CA admin port as 9443, all other services use 9090

**Question**: What should CA admin port be?

**A)** 9090 (consistency with all other services)
**B)** 9443 (PKI convention: 443-based ports)
**C)** Configurable (default 9090, allow override to 9443)
**D)** 9092 (unique port per service)
**E)** Other: _______________

**Your Answer**: D

---

### C7: CA Instance Count 🔴 HIGH

**Conflict**: KMS/JOSE have 3 instances (SQLite + 2× PostgreSQL), CA only has 1 (simple config)

**Question**: Should CA deployment match KMS/JOSE pattern?

**A)** Yes, CA needs 3 instances: ca-sqlite (8443), ca-postgres-1 (8444), ca-postgres-2 (8445)
**B)** No, CA is simpler: keep single instance deployment
**C)** Yes, but different ports: ca-sqlite (8443), ca-postgres-1 (8443), ca-postgres-2 (8443) with unique names
**D)** Defer to Phase 4+ after KMS/JOSE/Identity stabilized
**E)** Other: _______________

**Your Answer**: A

---

### A2: Package Classification for Coverage Targets 🔴 HIGH

**Ambiguity**: Unclear which packages are "production" (95%) vs "infrastructure" (100%) vs "utility" (100%)

**Question**: How should we classify packages?

**A)** Production: internal/{jose,identity,kms,ca}; Infrastructure: internal/cmd/cicd/*; Utility: internal/shared/*, pkg/*
**B)** Production: internal/*; Infrastructure: cmd/*, internal/cmd/*; Utility: pkg/*, scripts/*
**C)** Production: internal/{jose,identity,kms,ca}; Infrastructure+Utility: everything else (all 100%)
**D)** Case-by-case evaluation per package (document in clarify.md)
**E)** Other: _______________

**Your Answer**: E

---

### A3: Real vs Mock Testing Strategy 🔴 HIGH

**Ambiguity**: When to use real dependencies vs mocks in tests?

**Question**: What is the testing strategy for dependencies?

**A)** ALWAYS real dependencies (PostgreSQL, crypto, HTTP servers), ONLY mock external services (email, SMS)
**B)** ALWAYS mock dependencies (fast tests, no container overhead)
**C)** Real for integration tests, mocks for unit tests (hybrid approach)
**D)** Real if setup time <1s, mock if >1s (performance-based decision)
**E)** Other: _______________

**Your Answer**: A

---

### A4: Service Federation Configuration 🔴 HIGH

**Ambiguity**: How do services discover each other in "United Mode"?

**Question**: How should service federation work?

**A)** Static configuration: federation.identity_authz_url, federation.jose_url in YAML
**B)** DNS-based discovery: services resolve identity-authz, jose via internal DNS
**C)** Service mesh: Istio/Linkerd handles routing automatically
**D)** Environment variables: IDENTITY_AUTHZ_URL, JOSE_URL at runtime
**E)** Other: _______________

**Your Answer**: A

---

### A5: Phase Execution Order 🔴 HIGH

**Ambiguity**: Can phases run in parallel or must they be strictly sequential?

**Question**: What is the phase execution strategy?

**A)** Strictly sequential (P1→P2→P3→P4→P5→P6→P7, each blocks next)
**B)** Parallel where possible (P1+P3 parallel, then P2, then P4, etc.)
**C)** Flexible (start next phase when 80% of current phase complete)
**D)** Sequential per-product (KMS P1-7, then Identity P1-7, etc.)
**E)** Other: _______________

**Your Answer**: B

---

### C9: CLARIFY-QUIZME Cleanup 🔴 HIGH

**Issue**: CLARIFY-QUIZME.md contains 12 questions already answered in clarify.md

**Question**: How should we handle answered questions in CLARIFY-QUIZME.md?

**A)** Archive to "Decisions Made" section, keep only unanswered questions
**B)** Keep all questions for historical reference
**C)** Move answered questions to clarify.md with cross-references
**D)** Delete answered questions (info preserved in clarify.md)
**E)** Other: I think you said at end of specs\002-cryptoutil\CLARIFY-QUIZME.md that best practice is to document in them in clarify.md

**Your Answer**: E

---

### O8: Spec Kit Reference Missing from spec.md 🔴 HIGH

**Omission**: constitution.md has detailed Spec Kit workflow, spec.md doesn't reference it

**Question**: Should spec.md reference Spec Kit methodology?

**A)** Yes, add reference in spec.md Overview section
**B)** No, spec.md is implementation-focused (workflow not relevant)
**C)** Yes, add reference and summarize key gates
**D)** Add to README.md instead (top-level documentation)
**E)** Other: _______________

**Your Answer**: C

---

## Medium Priority Issues (Needs Clarification)

### O4: Hash Version Management Not in Constitution 🟡 MEDIUM

**Omission**: Hash registry version management (v1/v2/v3) documented in clarify.md but not constitution.md

**Question**: Should hash version management be added to constitution.md?

**A)** Yes, add to constitution.md Section II (Cryptographic Compliance)
**B)** No, keep only in clarify.md (Phase 5 implementation detail)
**C)** Yes, add to both constitution.md and spec.md Phase 5
**D)** Add to spec.md only (technical specification)
**E)** Other: Yes, add to both constitution.md and spec.md Phase 5 and copilot instructions

**Your Answer**: E

---

### O5: Service Template Not in Constitution 🟡 MEDIUM

**Omission**: Service template requirement documented in spec.md Phase 6 but not constitution.md

**Question**: Should service template be added to constitution.md?

**A)** Yes, add to constitution.md Section I (Product Delivery Requirements)
**B)** No, keep only in spec.md (Phase 6 implementation detail)
**C)** Yes, add high-level requirement to constitution.md, details in spec.md
**D)** Add to README.md instead (architecture overview)
**E)** Other:  Yes, add to both constitution.md and spec.md Phase 5 and copilot instructions

**Your Answer**: E

---

### A6: Gremlins Windows Panic Status 🟡 MEDIUM

**Ambiguity**: analyze.md mentions "gremlins v0.6.0 panics on Windows", workaround is "use CI/CD"

**Question**: What is the long-term plan for Windows mutation testing?

**A)** Temporary workaround: use CI/CD until gremlins v0.7.0 released
**B)** Permanent workaround: Windows devs use CI/CD, Linux devs run locally
**C)** Switch tools: evaluate go-mutesting or mutagen as alternatives
**D)** Document as known limitation: mutation testing requires Linux
**E)** Other: Make gremlins work on Windows, fix or workaround the panic; analyze root causes and identify solutions that you can use, and I can review for correctness

**Your Answer**: E

---

### Q1.1: Probabilistic Execution Tuning Strategy 🟡 MEDIUM

**Question**: If package with probabilistic execution STILL exceeds 12s, how should we tune?

**A)** Reduce probabilities (Quarter→Tenth, Tenth→Never)
**B)** Increase base coverage (Always→Quarter to reduce total tests)
**C)** Consolidate test cases (merge similar variant tests)
**D)** Dynamic probability (Always locally, Tenth in CI/CD)
**E)** Other: _______________

**Your Answer**: C

---

### Q1.2: Server Startup Overhead Reduction 🟡 MEDIUM

**Question**: How should we optimize server startup overhead in tests?

**A)** Share single server instance across all tests (TestMain pattern)
**B)** Use sync.Once per test package (setup once, reuse)
**C)** Mock HTTP handlers (no real server, httptest.ResponseRecorder)
**D)** Reduce server test count (fewer lifecycle tests)
**E)** Other: _______________

**Your Answer**: A

---

### Q2.1: HTTP Handler Testing Strategy 🟡 MEDIUM

**Question**: What testing strategy achieves 95%+ coverage for HTTP handlers?

**A)** httptest.ResponseRecorder for all handler tests (no real server)
**B)** testify/mock for request/response objects (full mocking)
**C)** Integration tests with real server + HTTP client (realistic)
**D)** Hybrid: Unit test handler logic, integration test middleware/routing
**E)** Other: _______________

**Your Answer**: ALWAYS C; if-and-only-if some corner cases are still unreachable, then augment C with C, but only for the absolute minimum extract coverage needed; C is ALWAYS BY PREFERRED CHOICE, UNLESS RARE CASE-BY-CASE EXCEPTION(S) NEEDED!!!

---

### Q2.2: Business Logic Coverage 🟡 MEDIUM

**Question**: How should we test business logic with complex dependencies?

**A)** Mock all dependencies (database, crypto, barrier) using testify/mock
**B)** Use real dependencies with test fixtures (in-memory database)
**C)** Integration tests with Docker Compose (realistic but slow)
**D)** Extract to pure functions, test separately from infrastructure
**E)** Other: _______________

**Your Answer**: B for integration tests, C for e2e tests, D for unit tests

---

### Q3.1: Mutation Testing Timeout Strategy 🟡 MEDIUM

**Question**: How should we restructure mutation testing to finish in <20 minutes?

**A)** Run gremlins only on business logic (exclude tests, generated code)
**B)** GitHub Actions matrix strategy (parallelize packages into 4-6 jobs)
**C)** Per-package timeout (fail fast for slow packages)
**D)** Run nightly (not on every PR/push)
**E)** Other:  Run gremlins only on business logic (exclude tests, generated code), with matrix strategy for parallel speed up, and Per-package timeout

**Your Answer**: E

---

### Q3.2: DAST Readyz Timeout Root Cause 🟡 MEDIUM

**Question**: What is root cause of /admin/v1/readyz timeout?

**A)** Increase timeout from 30s to 60s (GitHub Actions latency)
**B)** Optimize service startup (parallelize unseal, cache migrations)
**C)** Add retry logic with exponential backoff
**D)** Add diagnostic logging to identify bottleneck
**E)** Other: _______________

**Your Answer**: D

---

### Q3.3: Otel Collector Healthcheck Failure 🟡 MEDIUM

**Question**: Why does otel-collector healthcheck fail?

**A)** Increase healthcheck start_period (collector needs more time)
**B)** Fix healthcheck command (current command incorrect)
**C)** Add diagnostic logging to collector startup
**D)** Use sidecar health check (separate Alpine container)
**E)** Other: _______________

**Your Answer**: D IS ONLY SOLUTION I KNOW THAT WORKS BECAUSE OTEL CONTAINER DOESN'T HAVE INTERNAL HEALTH CHECK COMMAND THAT I KNOWN OF; If you want to try C, you can do C to augment D, but D must be kept because it is the only known working method

---

### Q5.1: Hash Version Selection Algorithm 🟡 MEDIUM

**Question**: Should hash version selection be input-size-based or explicit parameter?

**A)** Input-size-based (automatic: 0-31→v1, 32-47→v2, 48+→v3)
**B)** Explicit version parameter (caller specifies v1/v2/v3)
**C)** Hybrid (default auto-select, allow override)
**D)** Configuration-driven (version policy in config file)
**E)** Other: D; but clarification needed

- versions are date-based set of config, and each version always includes selection of SHA256||SHA384||SHA512 based on configurable input size
- versions map to NIST/OWASP revisions of other parameters, like recommended iterations & salt sizes per digest within a version

**Your Answer**: E

---

### Q5.2: Hash Output Format 🟡 MEDIUM

**Question**: What hash output format supports version-aware verification?

**A)** Prefix format: {v}:base64_hash (e.g., {1}:abcd1234...)
**B)** JSON format: {"v":1,"hash":"abcd1234..."}
**C)** Binary format: (1 byte version)(N bytes hash data)
**D)** PHC string format: $pbkdf2-sha256$v=1$rounds=...
**E)** Other: _______________

**Your Answer**: A

---

### Q6.1: Template Parameterization Strategy 🟡 MEDIUM

**Question**: How should ServerTemplate be parameterized for service-specific customization?

**A)** Constructor injection (pass handlers, middleware, config at init)
**B)** Interface-based (services implement ServerInterface)
**C)** Configuration-driven (YAML specifies handlers, middleware)
**D)** Plugin architecture (services register plugins)
**E)** Other: _______________

**Your Answer**: A

---

### Q6.3: Client SDK Generation Strategy 🟡 MEDIUM

**Question**: How should we automate client SDK generation?

**A)** Manual oapi-codegen runs (developer responsibility)
**B)** go:generate directives (auto-generate on `go generate`)
**C)** pre-commit hook (generate SDKs before commit)
**D)** CI/CD workflow (generate and commit SDKs automatically)
**E)** Other: B, but a user or LLM agent developer can do A during development too

**Your Answer**: ___

---

### Q8.2: Coverage Baseline Tracking Strategy 🟡 MEDIUM

**Question**: How should we track coverage baselines to detect regressions?

**A)** Git-tracked baseline files (test-output/*.out committed)
**B)** CI/CD artifacts (upload/download between runs)
**C)** Coverage service (Codecov, Coveralls, centralized)
**D)** Pre-commit hook (compare vs previous, fail if drops)
**E)** Other: _______________

**Your Answer**: B

---

## Low Priority Issues (Documentation)

### O6: File Size Limits Not in Constitution 🟢 LOW

**Question**: Should file size limits be added to constitution.md?

**A)** Yes, add to constitution.md Section VII (Code Quality Excellence)
**B)** No, keep only in spec.md (implementation detail)
**C)** Add to copilot instructions only (already exists)
**D)** Not needed (enforced by linters and reviews)
**E)** Other: Include in all 3; copilot instructions (already exists), constitution.md, spec.md

**Your Answer**: E

---

### O7: Learn-PS Not in Constitution 🟢 LOW

**Question**: Should Learn-PS demonstration service be added to constitution.md?

**A)** Yes, add to constitution.md Section I (Product Delivery Requirements)
**B)** No, keep only in spec.md Phase 7 (implementation detail)
**C)** Yes, add high-level requirement to constitution.md
**D)** Add to README.md instead (customer-facing documentation)
**E)** Other: Include in all 3; copilot instructions, constitution.md, spec.md

**Your Answer**: E

---

### A1: Terminology Consistency 🟢 LOW

**Question**: How should we standardize requirement terminology?

**A)** MANDATORY/MUST = hard requirement, REQUIRED = soft, CRITICAL = attention
**B)** MUST only (eliminate MANDATORY, REQUIRED, CRITICAL)
**C)** RFC 2119 keywords only (MUST, MUST NOT, SHOULD, MAY)
**D)** Current usage is fine (context determines meaning)
**E)** Other: I thought they were all logically equivalent

**Your Answer**: E

---

## Summary

**Total Questions**: 26
**Critical Priority**: 5
**High Priority**: 8
**Medium Priority**: 10
**Low Priority**: 3

**Answering Instructions**:

1. Write your answer (A/B/C/D/E) next to each question
2. For "E" answers, provide details in write-in space
3. Save file when complete
4. Answers will be persisted to appropriate Speckit documents (see below)

---

## How Speckit Persists Clarifications

Based on Speckit methodology and document analysis, clarifications should be persisted in **multiple locations** depending on content type:

### Primary Location: clarify.md

**Purpose**: Authoritative answers to ALL clarification questions
**Format**: Q&A with rationale, trade-offs, implementation guidance
**Update**: Add new sections for each CLARIFY-QUIZME.md question answered

### Secondary Locations (Selective)

#### 1. constitution.md

**When**: Clarification establishes new **fundamental principle or constraint**
**Examples**:

- Testing patterns (probability-based, main() pattern)
- Architecture requirements (dual HTTPS, Windows firewall)
- Quality thresholds (coverage targets, mutation scores)

#### 2. spec.md

**When**: Clarification adds **product/service requirements or technical specifications**
**Examples**:

- Service endpoints and ports
- Phase requirements and deliverables
- API specifications

#### 3. plan.md

**When**: Clarification affects **implementation approach or task sequencing**
**Examples**:

- Phase execution order
- Risk mitigation strategies
- Timeline adjustments

#### 4. copilot instructions (.github/instructions/*.instructions.md)

**When**: Clarification affects **LLM agent behavior or coding patterns**
**Examples**:

- Testing strategies (already has testing.instructions.md)
- Architecture patterns (already has architecture.instructions.md)
- Platform-specific patterns (already has cross-platform.instructions.md)

### Persistence Workflow

```
Answer CLARIFY-QUIZME.md questions
    ↓
1. ADD all answers to clarify.md (ALWAYS)
    ↓
2. Identify which answers establish principles
    ↓ (if yes)
3. ADD to constitution.md with source reference
    ↓
4. Identify which answers change requirements
    ↓ (if yes)
5. UPDATE spec.md with clarifications
    ↓
6. Identify which answers change implementation
    ↓ (if yes)
7. UPDATE plan.md with adjusted approach
    ↓
8. Identify which answers change agent behavior
    ↓ (if yes)
9. UPDATE copilot instructions with new patterns
    ↓
10. Mark questions as [ANSWERED] in CLARIFY-QUIZME.md
```

### Example: Mutation Testing Threshold (C2)

**After answering "A) 80% minimum, 98% target"**:

1. **clarify.md**: Add comprehensive Q&A with rationale
2. **constitution.md**: Update Section IV with "Minimum 80% (BLOCKING), Target 98% (RECOMMENDED)"
3. **spec.md**: Update Phase 4 requirements with clarified thresholds
4. **plan.md**: Adjust Phase 4 tasks based on 80% vs 98% requirement
5. **copilot instructions**: Update 01-04.testing.instructions.md if needed

**Traceability**: Each update should cite source question (e.g., "Source: CLARIFY-QUIZME.md C2, answered 2025-12-19")
