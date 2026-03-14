# Framework v3 - Iteration Plan

**Status**: IN PROGRESS — Phase 1: 11/12 tasks done. **quizme-v4 pending** (D12 tentative; Q2=C applied to D14, Q3=E→D18, Q4=A confirmed, Q5=E→quizme-v4, Q6=A→D17)
**Created**: 2026-03-08
**Last Updated**: 2026-03-14
**Depends On**: `docs/framework-brainstorm/08-recommendations.md`, framework-v1 (archived), framework-v2 (archived — see Completed in Prior Iterations)
**Purpose**: Aggressive standardization of all 10 product-services as thin domain-only wrappers around service-template. Service-template owns 100% of reusable infrastructure (servers, clients, authn, authz, middleware, health, TLS, barrier, telemetry, tests). Product-services inject ONLY domain-specific: OpenAPI add-ons, DB migrations, business logic, config overrides.

**Guiding Principle**: This repo is alpha development. NO backward compatibility. NO legacy code. All 10 product-services MUST use latest-and-greatest framework patterns.

---

## Companion Documents

1. **plan.md** (this file) - phases, objectives, decisions
2. **tasks.md** - task checklist per phase
3. **lessons.md** - persistent memory: what worked, what did not, root causes, patterns
4. **quizme-v4.md** - 3 pending questions (D12 skeleton role, Docker testing, terminology)

---

## Context: Where We Are After Framework v1

### What Framework v1 Delivered

1. **ServiceServer interface** - compile-time contract for all 10 services (11 methods)
2. **Builder auto-defaults** - `Build()` auto-configures JWTAuth + StrictServer, services declare only add-ons
3. **23 fitness sub-linters** - automated ARCHITECTURE.md enforcement via `cicd lint-fitness`
4. **Cross-service contract tests** - `RunContractTests(t, server)` for behavioral consistency (4 of 10 services)
5. **Shared test infrastructure** - testdb, testserver, fixtures, assertions, healthclient
6. **air live reload** - `SERVICE=sm-im air` for 2-3x faster dev loop

### What Framework v1 Did NOT Do

1. **No GitHub Workflows updated** - `lint-fitness` only runs via pre-commit, not CI
2. **Identity services** - only got compile-time assertions, no contract tests, minimal conformance work
3. **PKI-CA** - minimal treatment (assertion + contract tests), domain still partial
4. **No auth contract tests** - 401 rejection tests wrongly deferred to service-specific tests (auth is 100% service-template owned)
5. **Contract tests incomplete** - only 4 of 10 services have `RunContractTests` (missing: identity-authz/idp/rp/rs/spa, pki-ca)
6. **Builder still has redundant config logic** - services pass paths that service-template already knows
7. **Sequential test exemptions** - 173 total, many avoidable (58 viper/pflag, 37 os.Chdir)
8. **Lessons not fully propagated** - timeout double-multiplication not in skills/instructions
9. **Agent semantic commit grouping not working** - last v1 commit was a bulk commit mixing unrelated changes
10. **No skeleton CRUD reference** - skeleton-template still minimal (but may not be needed given lint-fitness)

### Service Maturity After v1

| Service | Interface | Contract Tests | Builder Simplified | Domain Logic | Migration Status |
|---------|-----------|---------------|-------------------|-------------|-----------------|
| sm-im | Yes | Yes | Yes (already was) | Working CRUD | Complete |
| jose-ja | Yes | Yes | Yes (already was) | Working CRUD | Complete |
| sm-kms | Yes | Yes | Yes (v1 simplified) | Working CRUD | Complete |
| pki-ca | Yes | Yes | Yes (already was) | Partial | In Progress |
| skeleton-template | Yes | Yes | Yes (already was) | Minimal | Reference only |
| identity-authz | Yes | No | Yes | Stub | Not Started |
| identity-idp | Yes | No | Yes | Stub | Not Started |
| identity-rp | Yes | No | Yes | Stub | Not Started |
| identity-rs | Yes | No | Yes | Stub | Not Started |
| identity-spa | Yes | No | Yes | Stub | Not Started |

---

## Guiding Decisions (From v1 Review)

### D1: Auth is 100% Service-Template Owned

AuthN/AuthZ are NOT domain-specific. They are 100% owned by service-template. Auth contract tests (401/403 rejection) belong in `RunContractTests`, NOT in service-specific tests. The v1 lesson "auth contracts belong in service-specific tests" was WRONG and is corrected here.

### D2: NO Builder Backward Compatibility

This repo is alpha. NO backward compatibility for the builder API. If the builder interface needs to change, it changes. All 10 services update immediately. No deprecation period, no migration path.

### D3: Product-Services Are Thin Wrappers

Product-services MUST be thin domain-only wrappers (like Spring Boot `@SpringBootApplication`). They inject ONLY:
- Domain OpenAPI spec addons
- Domain DB migrations
- Domain business logic handlers
- Domain config overrides (via config object, not redundant `With*()` calls)

ALL reusable infrastructure (servers, clients, authn, authz, middleware, health, TLS, barrier, telemetry, sessions, tests) lives in service-template. If code is duplicated across >1 service, it belongs in service-template.

### D4: ARCHITECTURE.md is THE Single Source of Truth

ALL knowledge propagates FROM `docs/ARCHITECTURE.md`. Agents, skills, instructions, and plan docs are downstream consumers. Propagation is enforced by `cicd lint-docs validate-propagation`.

### D5: No Legacy Code

All 10 product-services MUST use the latest framework patterns. No service gets to stay on an old pattern because "it works."

### D6: Skeleton-Template Purpose

skeleton-template's purpose needs analysis: given lint-fitness (23 sub-linters enforcing conformance), is skeleton still needed as a reference? It may serve as a minimal working example for `cicd new-service` scaffolding, or it may be redundant.

### D7: Test Infrastructure Rules

- **Unit tests**: NEVER start servers, NEVER start DBs
- **Integration tests**: ONE server per service (via TestMain)
- **E2E tests**: ONE docker-compose file
- Violations of these rules MUST be bubbled up to lint-fitness

### D8: lint-fitness is Full Investment (quizme-v1 Q2=A)

lint-fitness (23 sub-linters, 10,500 lines) gets FULL investment: >=98% coverage, >=95% mutation, all synthetic test fixtures evaluated. The value of automated ARCHITECTURE.md enforcement is proven and worth the maintenance cost.

### D9: Single Domain Config Struct for Builder (quizme-v1 Q3=A)

Builder refactoring uses a SINGLE domain config struct (not per-With()-call options). Product-services pass one config object; service-template picks what it needs. This is the cleanest API and aligns with D3 (thin wrappers).

### D10: Sequential Exemption Reduction Starts Smallest-First (quizme-v1 Q4=D)

Phase 4 starts with the smallest exemption categories (os.Stderr=5, pgDriver=11) to build momentum before tackling larger categories (os.Chdir=37, viper/pflag=58). Quick wins first establish patterns before the complex refactoring.

### D11: Agent Semantic Commits via Instructions Only (quizme-v1 Q7=C)

No automated commit linting enforcement (no commitlint, no CI validation). Improve agent instructions to better enforce the Multi-Category Fix Commit Rule. Trust the AI instructions, avoid tooling overhead for an alpha repo.

### D12: Skeleton-Template Stays Minimal (quizme-v2 Q1 → A recommended)

skeleton-template (19 Go files, 61KB) keeps its triple role: human reference, `/new-service` scaffolding source, lint-fitness validation target. NO CRUD endpoint added (sm-im at 61 files/357KB already IS the CRUD reference). NO code generation engine (over-engineered for ~10 services). Cost: ~2h/quarter to keep current.

### D13: Identity Full Extraction + Staged Reintegration (quizme-v2 Q2=A)

All 5 identity services + pki-ca get clean-slate treatment: archive domain logic to `_archived/`, replace with fresh skeletons, stage reintegration (rp/rs/spa first → authz → idp → pki-ca). Precondition: service-template MUST be fully proven first (able to stand up 5 thin services that work with minimal effort).

### D14: InsecureSkipVerify Phase 2A Now + Phase 2B E2E After Phase 7 (quizme-v3 Q2=C)

Phase 2A scope = integration + contract tests (current position in plan). Add `TLSBundle()` accessor to `ServiceServer`, migrate test HTTP clients to trust server's auto-generated CA. Eliminates ~90% of InsecureSkipVerify (38 of 47 files). Phase 2B = E2E Docker TLS after Phase 7, using a PKI init approach: prepend a Docker Compose job to generate all TLS files in Docker volume(s), creating ephemeral complete PKI domains for E2E/Demo/UAT/OnPrem. mTLS (2C) and PostgreSQL TLS (2D) deferred. The 1 production InsecureSkipVerify (identity-rp) gets fixed by D13 extraction.

### D15: Fix ALL 6 ARCHITECTURE.md TLS Gaps in Phase 2 (quizme-v2 Q4=A)

All 6 TLS gaps fixed as part of Phase 2 when InsecureSkipVerify removal provides implementation context: (1) TLS Certificate Configuration table, (2) TLS secrets in 12.3.3, (3) TLS test bundle pattern in 10.3, (4) ServiceServer.TLSBundle() in 10.3.5, (5) mTLS deployment architecture, (6) TLS mode taxonomy (Static/Mixed/Auto).

### D16: Architecture Status Table Set to 0% for Identity (quizme-v2 Q5=E)

All 5 identity-* services marked "⚠️ Extraction Pending 0%" in ARCHITECTURE.md Section 3.2 status table. Current "Complete 100%" entries are false. Domain code must be extracted to archive, services replaced with skeleton code. Update table again after extraction completes.

### D17: sm-kms Full Application Layer Extraction (quizme-v3 Q6=A)

sm-kms gets the same treatment as jose-ja (framework-v2 Phase 2) and sm-im (framework-v2 Phase 3): full application layer extraction, proper test coverage (property/fuzz/bench), >=95% coverage. sm-kms has 10 custom middleware files (claims, errors, introspection, jwt, jwt_revocation, realm_context, scopes, service_auth, session, tenant) — 5 have partial template counterparts, 5 need new template capabilities. Quality is paramount (Q4=A).

### D18: Go-Based Copilot Tools Collection (quizme-v3 Q3=E)

A Go-based collection of reusable Copilot tools to replace inefficient built-in tools and ad-hoc generated scripts. NOT part of the cicd command — a separate strategic investment. Think of it as a toolbox of Go programs that Copilot can invoke via MCP or custom tool definitions to replace fragile PowerShell heredocs and one-off scripts. Exact scope to be refined during implementation.

Phases 2, 6, 7, 8 restructured per these decisions.

---

## Goals for Framework v3

### Goal 1: Service-Template Standardization

Make service-template the single source of ALL reusable infrastructure:

- [ ] **Auth contract tests in RunContractTests** - 401/403 rejection tests as cross-service contracts
- [ ] **Contract tests for ALL 10 services** - identity-authz/idp/rp/rs/spa + pki-ca
- [ ] **Contract tests for ALL 5 products** - parameterized product-level contract tests
- [ ] **Contract test for the suite** - suite-level contract test
- [ ] **Builder refactoring** - services pass config objects, service-template picks what it needs (eliminate redundant `WithBrowserBasePath`/`WithServiceBasePath` logic in each service)
- [ ] **ServiceServer interface expansion analysis** - determine what integration tests need beyond current 11 methods (telemetry? JWK? barrier? TLS bundle? config?)

### Goal 2: CI/CD and Quality Infrastructure

- [ ] **ci-fitness.yml** - GitHub Actions workflow for `cicd lint-fitness`
- [ ] **lint-fitness coverage/mutation** - verify 10,500 lines meet >=98% quality gates
- [ ] **lint-fitness value assessment** - confirm 10,500 lines truly standardize services (not waste)
- [ ] **Test infrastructure rule enforcement** - add fitness linter for unit-test-starts-server violations

### Goal 3: Sequential Exemption Reduction

Deep analysis and reduction of 173 `// Sequential:` exemptions:

- [ ] **viper/pflag global state (58)** - SEAM PATTERN to inject config instead of global viper
- [ ] **os.Chdir (37)** - many in lint_fitness use `CheckInDir` pattern; verify which are truly needed
- [ ] **os.Stderr capture (5)** - inject `io.Writer` instead of capturing stderr
- [ ] **Other categories** - seam variable (11), pgDriver mock (11), SQLite in-memory (10), shared state (13), injectable function variables (16), signals (6), port reuse (5)

### Goal 4: Knowledge Propagation Completion

- [ ] **Timeout double-multiplication lesson** - propagate to skills and instructions (currently only in ARCHITECTURE.md and lessons.md)
- [ ] **DisableKeepAlives** - verify propagation complete (already in 03-02.testing and contract-test-gen)
- [ ] **Review doc simplification** - framework-v1/review.md is overwhelming; future reviews should be concise
- [ ] **Agent semantic commit enforcement** - fix agent guidance so bulk commits don't happen

### Goal 5: Security Infrastructure

- [x] **Semgrep in pre-commit** - `.semgrep/rules/` directory, initial rules
- [ ] **Remove InsecureSkipVerify (G402)** - generate TLS cert chains in testserver, remove G402 from gosec.excludes

### Goal 6: Product-Service Domain Logic

Following migration priority (sm-im > jose-ja > sm-kms > pki-ca > identity):

- [ ] **pki-ca domain completion** - certificate issuance, revocation, CRL, OCSP
- [ ] **identity-authz** - OAuth 2.1 authorization server
- [ ] **identity-idp** - identity provider (OIDC)
- [ ] **identity-rp, identity-rs, identity-spa** - remaining identity services

---

## Phases

### Phase 1: Close v1 Gaps and Knowledge Propagation [Status: IN PROGRESS — 11/12 tasks done]

**Objective**: Fix immediate gaps from v1 review. Small items implemented immediately.

- Fix lessons.md auth contracts item (auth is service-template owned, not service-specific)
- Propagate timeout double-multiplication lesson to skills and instructions
- Clean up temp files from research
- Add ci-fitness.yml GitHub Actions workflow
- Integrate contract tests into remaining 6 services (identity-authz/idp/rp/rs/spa, pki-ca)
- Add auth contract tests (401/403 rejection) to `RunContractTests`
- Verify lint-fitness coverage/mutation meets >=98%
- **Success**: All 10 services have `RunContractTests`, auth contracts in cross-service suite, ci-fitness.yml in CI
- **Post-Mortem**: lessons.md updated

### Phase 2: Remove InsecureSkipVerify — Integration Tests Only (D14, D15) [Status: TODO]

**Objective**: Eliminate InsecureSkipVerify from integration + contract tests (~90% of 47 files). Fix all 6 ARCHITECTURE.md TLS gaps (D15). Phase 2B (E2E Docker TLS with PKI init) deferred to after Phase 7 per D14 (quizme-v3 Q2=C).

- Add `TLSBundle()` accessor to `ServiceServer` interface (exposes server's auto-generated CA cert)
- Add `TLSClientConfig(t)` helper that trusts the test server's CA
- Migrate all 10 services' test HTTP clients from `InsecureSkipVerify: true` to `TLSClientConfig(t)`
- Document TLS mode taxonomy (Static/Mixed/Auto) in ARCHITECTURE.md Section 6
- Document TLS test bundle pattern in ARCHITECTURE.md Section 10.3
- Document TLS secrets in ARCHITECTURE.md Section 12.3.3
- E2E/Docker TLS (Phase 2B), mTLS (2C), PostgreSQL TLS (2D) explicitly deferred
- **Success**: Zero `InsecureSkipVerify` in integration/contract tests, 6 ARCHITECTURE.md TLS gaps fixed
- **Post-Mortem**: lessons.md updated

### Phase 3: Builder Refactoring [Status: TODO]

**Objective**: Product-services pass config objects; service-template picks what it needs.

- Analyze current builder `With*()` call patterns across all 10 services
- Refactor builder to accept domain config struct (OpenAPI spec, migrations FS, route registration)
- Eliminate redundant `WithBrowserBasePath`/`WithServiceBasePath` per-service logic
- All 10 services updated to new builder API (NO backward compatibility)
- **Success**: Product-service `NewFromConfig` is <=10 lines, zero duplicated path setup
- **Post-Mortem**: lessons.md updated

### Phase 4: Sequential Exemption Reduction [Status: TODO]

**Objective**: Reduce 173 `// Sequential:` exemptions by applying SEAM PATTERN and dependency injection. **Smallest-first** ordering per D10.

- Start with smallest categories to build momentum and establish patterns:
  1. os.Stderr capture (5) - inject `io.Writer` seam
  2. pgDriver registration (11) - evaluate test isolation approach
  3. seam variables (11) - already correct pattern, align documentation
  4. os.Chdir (37) - evaluate t.TempDir() + relative paths, lint_fitness CheckInDir
  5. viper/pflag (58) - inject config reader, largest category last
- Target: reduce from 173 to <100 exemptions
- **Success**: Each remaining exemption has justified `// Sequential:` comment
- **Post-Mortem**: lessons.md updated

### Phase 5: ServiceServer Interface Expansion [Status: TODO]

**Objective**: Analyze and expand the interface to cover integration test needs.

- Audit what integration tests need beyond current 11 methods
- Candidates: TelemetryService, JWKGenService, BarrierService, TLS bundle, Config accessor
- Expand interface (NO backward compatibility - all services update immediately)
- Update contract tests to exercise new interface methods
- **Success**: Integration tests can access all framework services through the interface
- **Post-Mortem**: lessons.md updated

### Phase 5B: sm-kms Full Application Layer Extraction (D17) [Status: TODO]

**Objective**: Extract sm-kms application layer (same treatment as jose-ja/sm-im in framework-v2). Quality is paramount (Q4=A).

- Extract application layer from sm-kms (separate from server startup)
- Resolve 10 custom middleware files: migrate 5 with partial template counterparts to service-template, evaluate 5 needing new template capabilities
- Add property tests, fuzz tests, benchmark tests
- Achieve >=95% coverage
- **Precondition**: Phase 3 (builder refactoring) complete
- **Success**: sm-kms application layer cleanly separated, >=95% coverage, middleware consolidated
- **Post-Mortem**: lessons.md updated

### Phase 6: lint-fitness Value Assessment [Status: TODO]

**Objective**: Confirm 10,500 lines of lint-fitness truly standardize services.

- Coverage and mutation testing of all 23 sub-linters
- Identify any sub-linters testing synthetic content vs real project files
- Evaluate whether 10,500 lines are justified or can be reduced
- skeleton-template stays minimal per D12 (no CRUD, no code generation). Verify current as scaffolding source.
- Add test infrastructure rule enforcement (unit-test-starts-server detection)
- **Success**: >=98% coverage, >=95% mutation, documented value assessment
- **Post-Mortem**: lessons.md updated

### Phase 7: Domain Extraction and Fresh Skeletons (D13, D16) [Status: TODO]

**Objective**: Extract domain logic from identity-* and pki-ca, replace with fresh skeleton-template copies.

- Archive all identity shared packages to `internal/apps/identity/_archived/`
- Archive each per-service domain code (authz, idp, rp, rs, spa) to `_archived/`
- Archive pki-ca domain code to `internal/apps/pki/_ca-archived/` (already exists, verify complete)
- Replace all 6 services with fresh skeleton-template copies (builder + contract tests + health)
- Update ARCHITECTURE.md status table: all 6 services → "⚠️ Extraction Pending 0%" (D16)
- **Precondition**: Phases 1-5 complete (service-template patterns proven on sm-im, jose-ja, sm-kms)
- **Success**: 6 clean skeleton services pass `RunContractTests`, all domain logic safely archived
- **Post-Mortem**: lessons.md updated

### Phase 8: Staged Domain Reintegration (D13) [Status: TODO]

**Objective**: Reintroduce archived domain logic into fresh skeletons, smallest-first.

- Stage 1: rp, rs, spa (10-18 files each, trivial domain, quick proof-of-pattern)
- Stage 2: authz (133 files/916KB, OAuth 2.1 core — largest complexity)
- Stage 3: idp (129 files/862KB, OIDC provider — second largest)
- Stage 4: pki-ca (48KB active + 880KB archived, certificate lifecycle)
- Each stage: extract relevant domain from archive → adapt to latest builder → test → commit
- **Success**: All 6 services have working domain + latest framework patterns
- **Post-Mortem**: lessons.md updated

### Phase 8B: E2E TLS with PKI Init (D14 Phase 2B) [Status: TODO]

**Objective**: Eliminate InsecureSkipVerify from E2E and Docker Compose tests using a PKI init approach.

- Create Docker Compose init job that generates all TLS certificates into Docker volume(s)
- Ephemeral complete PKI domains for E2E / Demo / UAT / OnPrem deployments
- All E2E tests use real TLS (no InsecureSkipVerify)
- Docker Compose services mount TLS volume and use generated certs
- **Precondition**: Phase 7 complete (clean skeleton services), Phase 2 complete (TLS bundle infrastructure)
- **Success**: Zero InsecureSkipVerify in E2E tests, all Docker Compose deployments use real TLS
- **Post-Mortem**: lessons.md updated

### Phase 9: Quality and Knowledge Propagation [Status: TODO]

**Objective**: Final quality sweep and knowledge propagation.

- Full coverage and mutation testing enforcement across all services
- Performance benchmarking for crypto operations
- Improve agent instructions for semantic commit grouping (D11: instructions only, no automated tooling)
- Propagate ALL lessons to ARCHITECTURE.md, agents, skills, instructions
- Simplify review document format for future iterations
- **Success**: All quality gates pass, all knowledge propagated, clean lessons.md
- **Post-Mortem**: lessons.md updated

---

## Known ARCHITECTURE.md Gaps (from quizme-v1 analysis)

**Evidence**: `test-output/framework-v2-quizme-analysis/analysis.md`

All gaps resolved by quizme-v2 decisions. Assigned to implementation phases:

### TLS Gaps → Phase 2 (D14, D15)

1. **TLS Certificate Configuration table** — exists in instructions (02-01) but NOT in ARCHITECTURE.md Section 6
2. **Secrets Coordination Strategy (12.3.3)** — no TLS CA/cert/key secrets documented (only unseal/DB/API secrets)
3. **No TLS test bundle pattern** — Section 10.3 lacks integration test TLS bundle documentation
4. **No ServiceServer.TLSBundle() accessor** — Section 10.3.5 contract test pattern missing TLS accessor
5. **No mTLS deployment architecture** — Section 6.3 mentions mTLS but no implementation strategy
6. **TLS mode taxonomy missing** — Code has Static/Mixed/Auto modes (`tls_generator.go`) but ARCHITECTURE.md does not document them

### Identity/Skeleton Gaps → Phases 6-8 (D12, D13, D16)

1. **Section 3.1.5** — doesn't define skeleton-template vs lint-fitness vs `/new-service` skill relationship
2. **Section 9.11** — doesn't mention skeleton as lint-fitness validation target
3. **Section 3.2 status table STALE** — shows identity-authz/idp/rs as "Complete 100%" but they lack contract tests, auth tests, and latest builder patterns
4. **No service domain extraction strategy** — no documented approach for archiving and reintroducing domain logic
5. **Archival naming inconsistency** — pki-ca uses `_ca-archived/` but no standard naming convention documented

### Resolution

- **TLS gaps 1-6**: Fixed in Phase 2 (D15)
- **Identity/Skeleton gaps 1-2**: Addressed in Phase 6 (D12 — skeleton stays minimal, document relationship)
- **Gap 3 (status table)**: Fixed in Phase 7 (D16 — identity-* → 0%)
- **Gaps 4-5 (extraction/archival)**: Addressed in Phases 7-8 (D13 — extraction + staged reintegration)

---

## Completed in Prior Iterations

### Framework v2 (Complete — 2026-03-13)

5 phases, 23 tasks — all complete. Key accomplishments:
- **Phase 1**: Created shared `testdb` package (`NewInMemorySQLiteDB`, `NewClosedSQLiteDB`), 6 fitness rules
- **Phase 2**: jose-ja cleanup — extracted application layer, added property/fuzz/bench tests, ≥95% coverage
- **Phase 3**: sm-im cleanup — extracted application layer, added error-path tests, ≥95% coverage
- **Phase 4**: sm-kms assessment — coverage ceiling analysis, seam injection pattern documented
- **Phase 5**: Knowledge propagation — ARCHITECTURE.md Section 13.1.1 (Opportunistic Quality Fixes), UUID literal construction, import safety

**Carried to framework-v3**: `no_local_closed_db_helper` fitness rule registration (Task 6.4), lint-fitness/lint-docs exit code 1 fix (Task 9.5), sm-im OpenAPI models enforcement (Task 8.5)

---

## Cross-References

- **Framework v1**: Archived (was `docs/framework-v1/`, 48/48 tasks complete)
- **Framework v2**: Archived (was `docs/framework-v2/`, 23/23 tasks complete — see Completed in Prior Iterations)
- **Framework Brainstorm**: `docs/framework-brainstorm/` (00-overview through 08-recommendations)
- **Architecture**: `docs/ARCHITECTURE.md` (single source of truth)
- **Migration Priority**: ARCHITECTURE.md Section 2.2 (sm-im > jose-ja > sm-kms > pki-ca > identity)
- **Service Template**: ARCHITECTURE.md Section 5.1 (template pattern), Section 5.2 (builder pattern)
- **Testing Strategy**: ARCHITECTURE.md Section 10 (testing architecture)
- **Quality Gates**: ARCHITECTURE.md Section 11.2 (quality gates)
- **Fitness Functions**: ARCHITECTURE.md Section 9.11 (fitness function catalog)
- **Sequential Exemptions**: ARCHITECTURE.md Section 10.2.5 (sequential test exemption)
