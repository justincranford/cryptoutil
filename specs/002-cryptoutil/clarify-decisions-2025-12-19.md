# Clarification Decisions - December 19, 2025

**Source**: SPECKIT-CONFLICTS-ANALYSIS.md answered questions
**Purpose**: Document all clarification decisions for persistence to constitution.md, spec.md, plan.md, and copilot instructions

---

## Critical Priority Decisions (Blocking Implementation)

### C2: Mutation Testing Threshold

**Decision**: E - "85% for Phase 4, then raise to 98% in Phase 5+"

**Rationale**: Incremental quality improvement - start with achievable target (85%) in Phase 4, then raise bar to aspirational target (98%) in Phase 5+.

**Impact**:

- Phase 4 (Mutation Testing): Block on 85% mutation score minimum
- Phase 5+ (Advanced Features): Raise requirement to 98% mutation score
- Gremlins configuration: Set `threshold.efficacy: 85` for Phase 4, update to 98 for Phase 5

**Persistence**:

- ✅ **clarify.md**: Add Q&A with rationale
- ✅ **constitution.md**: Update Section IV "Mutation testing ≥85% (Phase 4), ≥98% (Phase 5+)"
- ✅ **spec.md**: Update Phase 4 requirements with 85% threshold, Phase 5+ with 98%
- ✅ **plan.md**: Adjust Phase 4 acceptance criteria to 85%, Phase 5+ to 98%
- ✅ **copilot instructions (01-04.testing.instructions.md)**: Update mutation testing section with phased thresholds

---

### C3: Test Execution Time Targets

**Decision**: E - "MANDATORY <15s per unit test packages (i.e. !integration and !e2e tests); overall <180s for entire !integration and !e2e tests"

**Rationale**: Stricter targets for fast feedback loop. Unit tests must be fast (<15s per package). Integration/E2E tests excluded from strict timing due to Docker startup overhead.

**Impact**:

- Unit test packages: HARD LIMIT 15 seconds per package
- Full unit test suite: HARD LIMIT 180 seconds (3 minutes) total
- Integration/E2E tests: Excluded from timing targets (startup overhead acceptable)
- Probabilistic execution MANDATORY for packages approaching 15s

**Persistence**:

- ✅ **clarify.md**: Add Q&A with rationale and timing breakdown
- ✅ **constitution.md**: Update Section IV "Test execution <15s per unit test package, <180s total unit tests"
- ✅ **spec.md**: Update Phase 1 requirements with 15s/180s targets
- ✅ **plan.md**: Update Phase 1 approach with stricter timing enforcement
- ✅ **copilot instructions (01-04.testing.instructions.md)**: Update timing targets section

---

### O1: Probability-Based Testing Not in Constitution

**Decision**: E - "Add to constitution.md, clarify.md, and copilot instructions"

**Rationale**: Probability-based execution is a CRITICAL testing strategy for maintaining <15s unit test timing. Must be elevated to constitutional principle.

**Impact**:

- TestProbAlways (100%): Base algorithms (RSA2048, AES256, ES256)
- TestProbQuarter (25%): Key size variants (RSA3072, AES192)
- TestProbTenth (10%): Less common variants (RSA4096, AES128)
- TestProbNever (0%): Deprecated or edge case algorithms

**Persistence**:

- ✅ **clarify.md**: Add comprehensive Q&A with probability strategy
- ✅ **constitution.md**: Add Section IV subsection "Probability-Based Test Execution"
- ✅ **copilot instructions (01-04.testing.instructions.md)**: Already exists - verify completeness

---

### O2: main() Pattern Not in Constitution

**Decision**: E - "Add to constitution.md, clarify.md, and copilot instructions"

**Rationale**: main() → internalMain() pattern is CRITICAL for achieving 95%+ coverage. Must be constitutional requirement.

**Impact**:

- ALL main() functions must delegate to testable internalMain()
- internalMain() accepts injected dependencies (args, stdin, stdout, stderr)
- Tests achieve 95%+ coverage on internalMain()
- main() 0% coverage acceptable if internalMain() ≥95%

**Persistence**:

- ✅ **clarify.md**: Add comprehensive Q&A with pattern and examples
- ✅ **constitution.md**: Add Section IV subsection "main() Function Testability Pattern"
- ✅ **copilot instructions (01-04.testing.instructions.md)**: Already exists - verify completeness

---

### O3: Windows Firewall Prevention Not in Constitution

**Decision**: E - "Add to constitution.md, clarify.md, and copilot instructions"

**Rationale**: 127.0.0.1 binding (NOT 0.0.0.0) prevents Windows Firewall prompts during tests. Critical for CI/CD automation.

**Impact**:

- Unit/integration tests: MUST bind to 127.0.0.1 (NOT 0.0.0.0)
- Docker containers: MUST bind to 0.0.0.0 (container networking requirement)
- Prevents test automation breaking on Windows development machines

**Persistence**:

- ✅ **clarify.md**: Add Q&A with Windows Firewall rationale
- ✅ **constitution.md**: Add Section V "127.0.0.1 binding for tests (Windows Firewall prevention)"
- ✅ **copilot instructions (01-07.security.instructions.md)**: Already exists - verify completeness

---

## High Priority Decisions (Architecture Impact)

### C4: CA Admin Port Inconsistency

**Decision**: D - "9092 (unique port per service)"

**Rationale**: Each service gets unique admin port to prevent conflicts in unified deployments.

**Impact**:

- KMS admin port: 9090
- Identity admin port: 9091
- CA admin port: 9092
- JOSE admin port: 9093

**Persistence**:

- ✅ **clarify.md**: Add Q&A with port assignment table
- ✅ **spec.md**: Update all service port tables with unique admin ports
- ✅ **plan.md**: Update Phase 3.5 (service integration) with port changes

---

### C7: CA Instance Count

**Decision**: A - "Yes, CA needs 3 instances: ca-sqlite (8443), ca-postgres-1 (8444), ca-postgres-2 (8445)"

**Rationale**: Consistency across all services - demonstrate multi-instance patterns for production.

**Impact**:

- CA deployments match KMS/JOSE pattern (3 instances each)
- ca-sqlite: Public 8443, Admin 9092, Backend SQLite
- ca-postgres-1: Public 8444, Admin 9092, Backend PostgreSQL
- ca-postgres-2: Public 8445, Admin 9092, Backend PostgreSQL

**Persistence**:

- ✅ **clarify.md**: Add Q&A with CA deployment architecture
- ✅ **spec.md**: Update CA section with 3-instance deployment
- ✅ **plan.md**: Add Phase 4 task for CA multi-instance Docker Compose

---

### A2: Package Classification for Coverage Targets

**Decision**: E (user didn't specify details)

**Default Classification** (until clarified):

- Production (95%): internal/{jose,identity,kms,ca}
- Infrastructure (100%): internal/cmd/cicd/*
- Utility (100%): internal/shared/*, pkg/*

**Persistence**:

- ⏸️ **Deferred**: User needs to provide explicit classification rules
- ⏸️ Add to CLARIFY-QUIZME2.md for next round

---

### A3: Real vs Mock Testing Strategy

**Decision**: A - "ALWAYS real dependencies (PostgreSQL, crypto, HTTP servers), ONLY mock external services (email, SMS)"

**Rationale**: Real dependencies reveal production bugs. Mocks only for services that can't run locally (cloud-only).

**Impact**:

- PostgreSQL: ALWAYS use test containers (NOT mocks)
- Crypto operations: ALWAYS use real crypto (NOT mocks)
- HTTP servers: ALWAYS use real servers (NOT httptest mocks unless corner cases)
- Email/SMS: Mock (external services)

**Persistence**:

- ✅ **clarify.md**: Add comprehensive Q&A with real-vs-mock decision tree
- ✅ **constitution.md**: Add Section IV "Real Dependencies Preferred Over Mocks"
- ✅ **copilot instructions (01-04.testing.instructions.md)**: Already mentions real dependencies - strengthen wording

---

### A4: Service Federation Configuration

**Decision**: A - "Static configuration: federation.identity_authz_url, federation.jose_url in YAML"

**Rationale**: Simple, explicit, no DNS/service mesh dependency. Clear configuration for operators.

**Impact**:

- KMS federated with Identity: Add `federation.identity_authz_url: https://identity-authz:8180` to kms.yml
- KMS federated with JOSE: Add `federation.jose_url: https://jose:8280` to kms.yml
- YAML configuration preferred over environment variables

**Persistence**:

- ✅ **clarify.md**: Add Q&A with federation configuration pattern
- ✅ **spec.md**: Update federation section with YAML configuration examples
- ✅ **plan.md**: Add Phase 7 (Learn-PS) federation configuration tasks

---

### A5: Phase Execution Order

**Decision**: B - "Parallel where possible (P1+P3 parallel, then P2, then P4, etc.)"

**Rationale**: Maximize velocity. Independent phases can run in parallel.

**Impact**:

- Phase 1 (Test Optimization) + Phase 3 (CI/CD Fixes): Parallel (independent)
- Phase 2 (Coverage): Blocks on Phase 1 completion (needs fast tests)
- Phase 4 (Mutation): Blocks on Phase 2 completion (needs 95%+ coverage)
- Phase 5+ (Features): Sequential after Phase 4

**Persistence**:

- ✅ **clarify.md**: Add Q&A with phase dependency graph
- ✅ **plan.md**: Update execution strategy with parallel phase execution rules

---

### C9: CLARIFY-QUIZME Cleanup

**Decision**: E - "document them in clarify.md"

**Rationale**: Answered questions moved to clarify.md with [ANSWERED] markers in CLARIFY-QUIZME.md.

**Impact**:

- Move all answered questions from CLARIFY-QUIZME.md to clarify.md
- Mark moved questions with [ANSWERED YYYY-MM-DD] in CLARIFY-QUIZME.md
- Keep only unanswered questions in CLARIFY-QUIZME.md

**Persistence**:

- ✅ **clarify.md**: Add all answered questions from SPECKIT-CONFLICTS-ANALYSIS.md
- ✅ **CLARIFY-QUIZME.md**: Mark questions as [ANSWERED 2025-12-19]

---

### O8: Spec Kit Reference Missing from spec.md

**Decision**: C - "Yes, add reference and summarize key gates"

**Rationale**: spec.md should reference Spec Kit methodology for implementation guidance.

**Impact**:

- Add "Spec Kit Workflow" section to spec.md Overview
- Summarize key gates (clarify, analyze, checklist)
- Link to constitution.md Section VIII for full workflow

**Persistence**:

- ✅ **spec.md**: Add "Spec Kit Workflow" subsection to Overview

---

## Medium Priority Decisions (Needs Clarification)

### O4: Hash Version Management Not in Constitution

**Decision**: E - "Yes, add to both constitution.md and spec.md Phase 5 and copilot instructions"

**Persistence**:

- ✅ **clarify.md**: Add Q&A with hash version management
- ✅ **constitution.md**: Add Section II subsection on hash versioning
- ✅ **spec.md**: Update Phase 5 with hash version details
- ✅ **copilot instructions**: Consider new 01-09.cryptography.instructions.md

---

### O5: Service Template Not in Constitution

**Decision**: E - "Yes, add to both constitution.md and spec.md Phase 5 and copilot instructions"

**Persistence**:

- ✅ **clarify.md**: Add Q&A with service template pattern
- ✅ **constitution.md**: Add Section I subsection on service template
- ✅ **spec.md**: Update Phase 6 with template extraction details
- ✅ **copilot instructions (01-01.architecture.instructions.md)**: Add service template section

---

### A6: Gremlins Windows Panic Status

**Decision**: E - "Make gremlins work on Windows, fix or workaround the panic; analyze root causes and identify solutions"

**Rationale**: Windows is primary development platform - mutation testing must work locally.

**Impact**:

- Investigate gremlins v0.6.0 panic root cause
- Test gremlins v0.7.0+ for Windows compatibility
- Document workarounds if panic persists
- Fallback: CI/CD only until fixed

**Persistence**:

- ✅ **clarify.md**: Add Q&A with Windows gremlins investigation plan
- ✅ **plan.md**: Add Phase 4 subtask for Windows gremlins troubleshooting

---

### Q1.1: Probabilistic Execution Tuning Strategy

**Decision**: C - "Consolidate test cases (merge similar variant tests, and also refactor similar functions into table-driven test function)"

**Persistence**:

- ✅ **clarify.md**: Add Q&A with test consolidation pattern

---

### Q1.2: Server Startup Overhead Reduction

**Decision**: A preferred, D as fallback - "Share single server instance across all tests (TestMain pattern)"

**Persistence**:

- ✅ **clarify.md**: Add Q&A with TestMain pattern for server sharing

---

### Q2.1: HTTP Handler Testing Strategy

**Decision**: "ALWAYS C; if-and-only-if some corner cases are still unreachable, then augment C with C, but only for the absolute minimum extract coverage needed; C is ALWAYS BY PREFERRED CHOICE, UNLESS RARE CASE-BY-CASE EXCEPTION(S) NEEDED!!!"

**Interpretation**: ALWAYS use integration tests with real server (Option C). Only use httptest/mocks for unreachable corner cases.

**Persistence**:

- ✅ **clarify.md**: Add Q&A with real server testing emphasis

---

### Q2.2: Business Logic Coverage

**Decision**: "B for integration tests, C for e2e tests, D for unit tests"

**Interpretation**:

- Unit tests (D): Extract pure functions, test logic without infrastructure
- Integration tests (B): Use real dependencies with test fixtures (in-memory DB)
- E2E tests (C): Full Docker Compose stack

**Persistence**:

- ✅ **clarify.md**: Add Q&A with test strategy per test type

---

### Q3.1: Mutation Testing Timeout Strategy

**Decision**: E - "Run gremlins only on business logic (exclude tests, generated code), with matrix strategy for parallel speed up, and Per-package timeout"

**Persistence**:

- ✅ **clarify.md**: Add Q&A with gremlins optimization strategy

---

### Q3.2: DAST Readyz Timeout Root Cause

**Decision**: D - "Add diagnostic logging to identify bottleneck"

**Persistence**:

- ✅ **clarify.md**: Add Q&A with diagnostic logging approach

---

### Q3.3: Otel Collector Healthcheck Failure

**Decision**: "D IS ONLY SOLUTION I KNOW THAT WORKS BECAUSE OTEL CONTAINER DOESN'T HAVE INTERNAL HEALTH CHECK COMMAND THAT I KNOWN OF; If you want to try C, you can do C to augment D, but D must be kept because it is the only known working method"

**Interpretation**: Use sidecar health check (D) as primary solution. Optional: add collector diagnostic logging (C) as supplementary.

**Persistence**:

- ✅ **clarify.md**: Add Q&A with sidecar health check pattern

---

### Q5.1: Hash Version Selection Algorithm

**Decision**: E - "D; but clarification needed - versions are date-based set of config, and each version always includes selection of SHA256||SHA384||SHA512 based on configurable input size - versions map to NIST/OWASP revisions of other parameters, like recommended iterations & salt sizes per digest within a version"

**Interpretation**: Configuration-driven (D), but version = date-based policy revision (v1=2020 NIST, v2=2023 NIST, v3=2025 OWASP), each version selects SHA variant by input size.

**Persistence**:

- ✅ **clarify.md**: Add comprehensive Q&A with hash version architecture

---

### Q5.2: Hash Output Format

**Decision**: A - "Prefix format: {v}:base64_hash (e.g., {1}:abcd1234...)"

**Persistence**:

- ✅ **clarify.md**: Add Q&A with hash output format

---

### Q6.1: Template Parameterization Strategy

**Decision**: A - "Constructor injection (pass handlers, middleware, config at init)"

**Persistence**:

- ✅ **clarify.md**: Add Q&A with service template parameterization

---

### Q6.3: Client SDK Generation Strategy

**Decision**: "B, but a user or LLM agent developer can do A during development too"

**Interpretation**: go:generate directives (B) as primary automation. Manual oapi-codegen (A) acceptable during development.

**Persistence**:

- ✅ **clarify.md**: Add Q&A with SDK generation workflow

---

### Q8.2: Coverage Baseline Tracking Strategy

**Decision**: B - "CI/CD artifacts (upload/download between runs)"

**Persistence**:

- ✅ **clarify.md**: Add Q&A with coverage artifact strategy

---

## Low Priority Decisions (Documentation)

### O6: File Size Limits Not in Constitution

**Decision**: E - "Include in all 3; copilot instructions (already exists), constitution.md, spec.md"

**Persistence**:

- ✅ **constitution.md**: Add Section VII subsection "File Size Limits"
- ✅ **spec.md**: Add file size limits to coding standards section
- ✅ **copilot instructions (01-03.coding.instructions.md)**: Already exists - verify completeness

---

### O7: Learn-PS Not in Constitution

**Decision**: E - "Include in all 3; copilot instructions, constitution.md, spec.md"

**Persistence**:

- ✅ **constitution.md**: Add Section I "Learn-PS Demonstration Requirement"
- ✅ **spec.md**: Already exists in Phase 7 - verify completeness
- ✅ **copilot instructions**: Add to 01-01.architecture.instructions.md if needed

---

### A1: Terminology Consistency

**Decision**: E - "I thought they were all logically equivalent"

**User Clarification Needed**: User believes MUST/REQUIRED/CRITICAL are synonyms. Need to document this explicitly.

**Recommendation**: Adopt RFC 2119 keywords (MUST, MUST NOT, SHOULD, MAY) for precision, but acknowledge user's intent that MANDATORY/REQUIRED/CRITICAL are synonyms for MUST.

**Persistence**:

- ✅ **clarify.md**: Add Q&A explaining terminology equivalence
- ✅ **constitution.md**: Add Section IX "Terminology Standards" (RFC 2119 + user's synonym mapping)

---

## Summary

**Total Decisions**: 26
**Requiring clarify.md updates**: 26
**Requiring constitution.md updates**: 13
**Requiring spec.md updates**: 8
**Requiring plan.md updates**: 4
**Requiring copilot instructions updates**: 8

**Next Steps**:

1. Apply all updates to clarify.md (comprehensive Q&A for all 26 decisions)
2. Apply critical updates to constitution.md (13 sections)
3. Apply architecture updates to spec.md (8 sections)
4. Apply execution updates to plan.md (4 sections)
5. Apply pattern updates to copilot instructions (8 files)
6. Commit changes with conventional commit messages
7. Create CLARIFY-QUIZME2.md with new clarification questions
