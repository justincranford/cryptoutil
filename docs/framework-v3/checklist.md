# Framework v3 — Master Checklist

**Purpose**: Single reference for everything that must be done, verified, or decided during framework-v3.
Derived from plan.md, tasks.md, lessons.md, and all planning sessions.

**Status Key**: ☐ Not Started | ⚙ In Progress | ✅ Complete | ⚠ Blocked | 🔵 Deferred

---

## Outstanding Work

### Decisions to Apply

- ☐ D7 (3-tier test strategy) enforced by lint-fitness for all 10 services + skeleton-template
- ☐ D9 (single domain config struct) implemented in builder refactor (Phase 3)
- ☐ D10 (sequential exemption reduction, smallest-first) started (Phase 4)
- ☐ D11 (semantic commit instructions) updated in agents (Task 9.2)
- ☐ D12 (skeleton-template OpenAPI CRUD example) implemented (Task 10.4)
- ☐ D13 (identity full extraction + staged reintegration) done (Phases 7-8)
- ☐ D14 Phase 2A (InsecureSkipVerify removed from integration/contract tests) done (Phase 2)
- ☐ D14 Phase 2B (E2E TLS with PKI init) done (Phase 8B)
- ☐ D15 (fix 6 ARCHITECTURE.md TLS gaps) done (Task 2.9)
- ☐ D16 (identity status table set to 0%) done (Task 7.4)
- ☐ D17 (sm-kms full application layer extraction) done (Phase 5B)
- ☐ D18 (Go-based Copilot tools) scoped and started (Phase 9 D26 catalog) then built in background
- ☐ D19 (test strategy in ARCHITECTURE.md + Copilot artifacts + lint-fitness enforcement)
- ☐ D20 (rename to service-framework) done (Phase 11 — FINAL)
- ☐ D21 (OpenAPI directory naming standardized) done (Phase 10)
- ☐ D22 (initialisms consolidated to base list + domain additions) done (Phase 10)
- ☐ D23 (FiberHandlerOpenAPISpec deduplicated) done (Phase 10)
- ☐ D24 (all services have api/<service-name>/ dir, lint-fitness enforced) done (Phase 10)
- ☐ D25 (MCP adopted only after ≥3 Go tools exist; prompt files/chat modes not needed)
- ☐ D26 (project tool catalog added to 04-01.deployment.instructions.md) done (Task 9.8)

### Phase 1: Close v1 Gaps and Knowledge Propagation

- ☐ 1.12 Phase 1 validation and post-mortem: full quality gate run, lessons.md updated

### Phase 2: Remove InsecureSkipVerify — Integration Tests Only (D14, D15)

- ☐ 2.1 Add TLSBundle() accessor to ServiceServer interface in service-template testserver
- ☐ 2.2 Migrate sm-im test HTTP clients to trust server's auto-generated CA via TLSBundle()
- ☐ 2.3 Migrate jose-ja test HTTP clients to trust server's auto-generated CA
- ☐ 2.4 Migrate sm-kms test HTTP clients to trust server's auto-generated CA
- ☐ 2.5 Migrate pki-ca test HTTP clients to trust server's auto-generated CA
- ☐ 2.6 Migrate all 5 identity service test HTTP clients
- ☐ 2.7 Migrate skeleton-template test HTTP clients
- ☐ 2.8 Remove G402 from gosec.excludes; activate semgrep no-tls-insecure-skip-verify rule
- ☐ 2.9 Fix ALL 6 ARCHITECTURE.md TLS gaps (D15): cert config table, 12.3.3 secrets, 10.3 TLS bundle, 10.3.5 TLSBundle() accessor, 6.3 mTLS strategy, TLS mode taxonomy

### Phase 3: Builder Refactoring

- ☐ 3.1 Analyze current builder With*() call patterns across all 10 services
- ☐ 3.2 Design new builder domain config API (single struct per D9)
- ☐ 3.3 Implement builder refactoring: accept domain config struct replacing redundant With*() calls
- ☐ 3.4 Migrate ALL 10 services to new builder API (no backward compat per D2)
- ☐ 3.5 Phase 3 validation and post-mortem

### Phase 4: Sequential Exemption Reduction

- ☐ 4.1 Categorize and triage all 173 t.Parallel() sequential exemptions
- ☐ 4.2 Inject io.Writer for os.Stderr tests (5 exemptions)
- ☐ 4.3 Fix pgDriver registration exemptions (11 exemptions)
- ☐ 4.4 Seam variable audit — fix where possible (11 exemptions)
- ☐ 4.5 Evaluate os.Chdir exemptions (37 — may need PathFS seam or abstraction)
- ☐ 4.6 Seam pattern for viper/pflag tests (58 exemptions — ParseWithFlagSet pattern)
- ☐ 4.7 Remaining exemption categories (remaining 51)
- ☐ 4.8 Phase 4 validation and post-mortem

### Phase 5: ServiceServer Interface Expansion

- ☐ 5.1 Audit ALL integration tests to identify missing ServiceServer accessor methods
- ☐ 5.2 Expand ServiceServer interface with required accessors (e.g., TLSBundle(), DB(), etc.)
- ☐ 5.3 Phase 5 validation and post-mortem

### Phase 5B: sm-kms Full Application Layer Extraction (D17)

- ☐ 5B.1 Extract sm-kms application layer (same pattern as jose-ja framework-v2 Phase 2)
- ☐ 5B.2 Analyze and migrate 10 custom sm-kms middleware files to service-template or new packages
- ☐ 5B.3 Add property, fuzz, and benchmark tests for sm-kms domain logic; reach ≥95% coverage
- ☐ 5B.4 Phase 5B validation and post-mortem

### Phase 6: lint-fitness Value Assessment

- ☐ 6.1 Run coverage (≥98%) and mutation (≥95%) on all 23 lint-fitness sub-linters
- ☐ 6.2 Classify each sub-linter: real vs synthetic test content; plan conversion
- ☐ 6.3 Update skeleton-template to use latest builder API (prerequisite for Phase 10 Task 10.4)
- ☑ 6.4 Add test infrastructure linters: detect unit tests starting servers/DBs; register no_local_closed_db_helper; add no_local_create_closed_database rule
- ☐ 6.6 Add PostgreSQL isolation enforcement linters (D19): block postgres.RunContainer and testdb.NewPostgresTestContainer outside E2E build tag
- ☐ 6.7 Phase 6 validation and post-mortem

### Phase 7: Domain Extraction and Fresh Skeletons (D13, D16)

- ☐ 7.1 Archive all identity shared packages to internal/apps/identity/_archived/
- ☐ 7.2 Archive per-service domain code (authz, idp, rp, rs, spa) and pki-ca domain to _archived/
- ☐ 7.3 Replace all 6 services with fresh skeleton-template copies (builder + contract tests + health)
- ☐ 7.4 Update ARCHITECTURE.md status table: 6 services → "⚠ Extraction Pending 0%" (D16)
- ☐ 7.5 Phase 7 validation and post-mortem

### Phase 8: Staged Domain Reintegration (D13)

- ☐ 8.1 Reintegrate rp, rs, spa domain — Stage 1 (simple services, token relay, sessions)
- ☐ 8.2 Reintegrate authz domain — Stage 2 (OAuth 2.1 flows, scope management)
- ☐ 8.3 Reintegrate idp domain — Stage 3 (identity provider, MFA, social login stubs)
- ☐ 8.4 Reintegrate pki-ca domain — Stage 4 (certificate issuance, CA hierarchy, CRL/OCSP)
- ☐ 8.5 Enforce OpenAPI-generated models for ALL service HTTP handlers (sm-im enforced per v1 deferred task)
- ☐ 8.6 Phase 8 validation and post-mortem

### Phase 8B: E2E TLS with PKI Init (D14 Phase 2B)

- ☐ 8B.1 Design PKI init Docker Compose job (prepend to E2E compose, generate all TLS files in volume)
- ☐ 8B.2 Implement PKI init certificate generator (ephemeral PKI for E2E/Demo/UAT/OnPrem)
- ☐ 8B.3 Migrate E2E Docker Compose to use real TLS from PKI init volume
- ☐ 8B.4 Phase 8B validation and post-mortem

### Phase 9: Quality and Knowledge Propagation

- ☐ 9.1 Full coverage (≥95% prod, ≥98% infra) and mutation enforcement across all services
- ☐ 9.2 Improve agent semantic commit instructions (D11: instructions-only, no automated tooling)
- ☐ 9.3 Propagate ALL lessons from lessons.md to permanent artifacts (ARCHITECTURE.md, agents, skills, instructions)
- ☐ 9.4 Design and document simplified review document format for future iterations
- ☐ 9.5 Fix lint-fitness and lint-docs exit code 1 (pre-existing CI/CD stderr issue)
- ☐ 9.6 Verify Docker Desktop startup directive present in all agents/skills involving Docker or E2E
- ☐ 9.7 Propagate D7/D19 3-tier test strategy to ARCHITECTURE.md Section 10, 03-02.testing.instructions.md, all agents
- ☐ 9.8 Add project-specific cicd tool catalog to 04-01.deployment.instructions.md (D26)
- ☐ 9.9 Phase 9 validation and post-mortem

### Phase 10: OpenAPI Standardization (D21–D24, D12 Part)

- ☐ 10.1 Rename api/kms/ → api/sm-kms/, api/ca/ → api/pki-ca/, api/jose/ → api/jose-ja/ (D21)
- ☐ 10.2 Split api/identity/ → api/identity-authz/, api/identity-idp/, api/identity-rs/, api/identity-rp/, api/identity-spa/ (D21)
- ☐ 10.3 Create api/sm-im/ with canonical structure and OpenAPI spec (D21, D24)
- ☐ 10.4 Create api/skeleton-template/ + OpenAPI Item CRUD spec + ~100-line Item domain example (D12, D21)
- ☐ 10.5 Document canonical base initialisms list in ARCHITECTURE.md Section 8.1; update all gen configs (D22)
- ☑ 10.6 Refactor FiberHandlerOpenAPISpec into shared service-template factory taking rawSpec() (D23)
- ☑ 10.7 Add require_api_dir lint-fitness sub-linter enforcing api/<service-name>/ structure (D24)
- ☑ 10.8 Phase 10 validation and post-mortem

### Phase 11: service-framework Rename — FINAL (D20)

- ☑ 11.1 Enumerate all ~340 files referencing internal/apps/template as framework; plan rename
- ☑ 11.2 Rename internal/apps/framework/ → internal/apps/framework/; update all imports/identifiers
- ☑ 11.3 Update ALL docs + Copilot artifacts (ARCHITECTURE.md, agents, skills, instructions, copilot-instructions.md)
- ☐ 11.4 Add lint-fitness rule blocking new internal/apps/template framework imports (skeleton-template whitelisted)
- ☐ 11.5 Update all GitHub workflows, Dockerfiles, docker-compose files
- ☐ 11.6 Phase 11 validation and post-mortem (FINAL)

### Cross-Cutting Work

- ☐ After each phase: review .semgrep/rules/go-testing.yml for new relevant patterns
- ☐ After Phase 2 complete: uncomment no-tls-insecure-skip-verify in go-testing.yml
- ☐ Design parameterized product-level contract tests (5 products) after Phase 1 service-level complete
- ☐ Design suite-level contract test (1 suite)

### ARCHITECTURE.md Gaps to Resolve

- ☐ Section 3.1.5 — define skeleton-template vs lint-fitness vs /new-service skill relationship (Task 6.3)
- ☐ Section 6 — add TLS Certificate configuration table (Task 2.9)
- ☐ Section 6.3 — add mTLS implementation strategy (Task 2.9)
- ☐ Section 8.1 — add canonical OpenAPI initialisms list (Task 10.5)
- ☐ Section 9.11 — mention skeleton as lint-fitness validation target (Task 6.3)
- ☐ Section 10 — comprehensive 3-tier test strategy (unit/integration/E2E) with PostgreSQL isolation rule (Task 9.7)
- ☐ Section 10.3 — add TLS test bundle pattern documentation (Task 2.9)
- ☐ Section 10.3.5 — add TLSBundle() to ServiceServer contract test pattern (Task 2.9)
- ☐ Section 12.3.3 — add TLS CA/cert/key secrets to secrets coordination (Task 2.9)
- ☐ Section 3.2 — update service status table; identity services → 0% pending extraction (Task 7.4)
- ☐ TLS mode taxonomy (Static/Mixed/Auto from tls_generator.go) documented (Task 2.9)

### Copilot Artifact Propagation Needed

- ☐ 03-02.testing.instructions.md — propagate D7/D19 3-tier test strategy from ARCHITECTURE.md (Task 9.7)
- ☐ All agents (implementation-execution, implementation-planning, beast-mode) — reference D7 canonical strategy
- ☐ 04-01.deployment.instructions.md — add project tool catalog (D26, Task 9.8)
- ☐ beast-mode.agent.md — update semantic commit grouping examples (Task 9.2)
- ☐ implementation-execution.agent.md — update commit checkpoint pattern (Task 9.2)
- ☐ cicd lint-docs validate-propagation passes after all updates

### Fitness Linter Additions Needed

- ☑ no_local_closed_db_helper registered in lint_fitness.go (deferred from v2)
- ☑ no_local_create_closed_database (detects createClosedDatabase/createClosedDBHandler outside testdb)
- ☑ detect unit tests starting real servers
- ☑ detect unit tests starting real databases
- ☐ postgres_in_unit_integration_tests: block postgres.RunContainer outside E2E tag
- ☐ require_api_dir: enforce api/<service-name>/ structure for all 10 services + skeleton
- ☐ openapi_gen_config_initialisms: flag gen configs duplicating canonical base-list items
- ☐ no_service_template_framework_import: after rename, block internal/apps/template framework imports
- ☐ no_tls_insecure_skip_verify activated (after Phase 2 complete)

### Service-Level Work Summary

| Service    | Contract Tests | InsecureSkipVerify | App Layer | OpenAPI api/ dir | Status |
|------------|---------------|-------------------|-----------|-----------------|--------|
| sm-im      | ✅ Done        | ☐ Phase 2          | ✅ Done    | ☐ Phase 10       | Phase 2 next |
| jose-ja    | ✅ Done        | ☐ Phase 2          | ✅ Done    | ☐ Rename Phase 10| Phase 2 next |
| sm-kms     | ✅ Done        | ☐ Phase 2          | ☐ Phase 5B | ☐ Rename Phase 10| Phase 5B |
| pki-ca     | ✅ Done        | ☐ Phase 2          | Partial   | ☐ Rename Phase 10| Phase 2/8 |
| skeleton   | ✅ Done        | ☐ Phase 2          | Minimal   | ☐ Phase 10 new  | Phase 2/10 |
| id-authz   | ✅ Done (v3)   | ☐ Phase 2          | ☐ Phase 7  | ☐ Phase 10 new  | Phase 7 |
| id-idp     | ✅ Done (v3)   | ☐ Phase 2          | ☐ Phase 7  | ☐ Phase 10 new  | Phase 7 |
| id-rp      | ✅ Done (v3)   | ☐ Phase 2          | ☐ Phase 7  | ☐ Phase 10 new  | Phase 7 |
| id-rs      | ✅ Done (v3)   | ☐ Phase 2          | ☐ Phase 7  | ☐ Phase 10 new  | Phase 7 |
| id-spa     | ✅ Done (v3)   | ☐ Phase 2          | ☐ Phase 7  | ☐ Phase 10 new  | Phase 7 |

---

## Completed Work

*(This section grows as items above are finished. Each item moves here with evidence — commit hash or test output reference.)*

### Phase 1: Close v1 Gaps and Knowledge Propagation (Tasks 1.1–1.11 ✅)

- ✅ 1.1 Fix lessons.md auth contracts item — corrected v1 lesson "auth contracts belong in service-specific tests" was WRONG
- ✅ 1.2 Propagate timeout double-multiplication lesson to ARCHITECTURE.md, instructions, skills
- ✅ 1.3 Clean up temp research files (test-output/framework-v2-quizme-analysis/ et al.)
- ✅ 1.4 Add ci-fitness.yml GitHub Actions workflow — lint-fitness now runs in CI
- ✅ 1.5 Add auth contract tests to RunContractTests (401/403 rejection, not in service-specific tests)
- ✅ 1.6 Integrate RunContractTests into identity-authz
- ✅ 1.7 Integrate RunContractTests into identity-idp
- ✅ 1.8 Integrate RunContractTests into identity-rp
- ✅ 1.9 Integrate RunContractTests into identity-rs
- ✅ 1.10 Integrate RunContractTests into identity-spa
- ✅ 1.11 Verify lint-fitness coverage and mutation (documented; coverage improvement deferred)

### Planning and Architecture Decisions (All D1–D26 Confirmed)

- ✅ D1: Auth is 100% Service-Template Owned (confirmed, v1 lesson corrected)
- ✅ D2: NO Builder Backward Compatibility (confirmed)
- ✅ D3: Product-Services Are Thin Wrappers (confirmed)
- ✅ D4: ARCHITECTURE.md is THE Single Source of Truth (confirmed)
- ✅ D5: No Legacy Code (confirmed)
- ✅ D6: Skeleton-Template Purpose (resolved by D12)
- ✅ D7: Test Infrastructure Rules (confirmed + expanded with 3-tier strategy)
- ✅ D8: lint-fitness is Full Investment (confirmed)
- ✅ D9: Single Domain Config Struct for Builder (confirmed)
- ✅ D10: Sequential Exemption Reduction Starts Smallest-First (confirmed)
- ✅ D11: Agent Semantic Commits via Instructions Only (confirmed, no automated tooling)
- ✅ D12: Skeleton-Template Gets OpenAPI CRUD Example (confirmed quizme-v4 Q1=B)
- ✅ D13: Identity Full Extraction + Staged Reintegration (confirmed)
- ✅ D14: InsecureSkipVerify Phase 2A Now + Phase 2B E2E After Phase 7 (confirmed)
- ✅ D15: Fix ALL 6 ARCHITECTURE.md TLS Gaps in Phase 2 (confirmed)
- ✅ D16: Architecture Status Table Set to 0% for Identity (confirmed)
- ✅ D17: sm-kms Full Application Layer Extraction (confirmed quizme-v3 Q6=A)
- ✅ D18: Go-Based Copilot Tools Collection — strategy defined (D25 MCP strategy added)
- ✅ D19: Test Strategy Canonical Documentation + Enforcement (confirmed quizme-v4 Q2)
- ✅ D20: Rename to "service-framework" — FINAL Phase (confirmed quizme-v4 Q3)
- ✅ D21: OpenAPI Directory Naming Standardization (confirmed)
- ✅ D22: OpenAPI Initialisms Consolidation (confirmed)
- ✅ D23: FiberHandlerOpenAPISpec Deduplication (confirmed)
- ✅ D24: All Services Must Have api/ Directories (confirmed)
- ✅ D25: MCP Adoption Strategy — start with terminal, MCP later; no prompt files/chat modes needed
- ✅ D26: Project-Specific Tool Catalog in Instructions (confirmed)

### Quizme Series (All Answered and Deleted)

- ✅ quizme-v1: Q1–Q7 answered (D8–D11, D13–D16 set)
- ✅ quizme-v2: Q1–Q5 answered (D12 tentative A, D13 A, D14 tentative B, D15 A, D16 E)
- ✅ quizme-v3: Q1–Q6 answered (D12→E again, D14→C, D18→E, D1 confirmed, D4 confirmed, D17→A)
- ✅ quizme-v4: Q1–Q3 answered (D12→B final, D7 expanded, D20 added)
- ✅ All quizme files deleted after answers applied

### Framework v2 (Complete — committed 2026-03-13)

- ✅ Phase 1: Created shared testdb package (NewInMemorySQLiteDB, NewClosedSQLiteDB), 6 fitness rules
- ✅ Phase 2: jose-ja cleanup — extracted application layer, property/fuzz/bench tests, ≥95% coverage
- ✅ Phase 3: sm-im cleanup — extracted application layer, error-path tests, ≥95% coverage
- ✅ Phase 4: sm-kms assessment — coverage ceiling analysis, seam injection pattern documented
- ✅ Phase 5: Knowledge propagation — ARCHITECTURE.md Section 13.1.1, UUID literal construction, import safety

### Agent and Instruction Improvements Applied

- ✅ VS Code hot-exit mitigation added to implementation-execution.agent.md
- ✅ Quizme lifecycle rules added to implementation-planning.agent.md (one quizme at a time, delete after applying, carry forward unanswered with more depth)
- ✅ COPILOT-MULTI-PROJECT.md created (multi-project Copilot pattern documentation)
