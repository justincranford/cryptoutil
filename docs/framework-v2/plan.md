# Framework v2 - Iteration Plan

**Status**: PLANNING
**Created**: 2026-03-08
**Depends On**: `docs/framework-v1/` (complete), `docs/framework-brainstorm/08-recommendations.md`
**Purpose**: Aggressive standardization of all 10 product-services as thin domain-only wrappers around service-template. Service-template owns 100% of reusable infrastructure (servers, clients, authn, authz, middleware, health, TLS, barrier, telemetry, tests). Product-services inject ONLY domain-specific: OpenAPI add-ons, DB migrations, business logic, config overrides.

**Guiding Principle**: This repo is alpha development. NO backward compatibility. NO legacy code. All 10 product-services MUST use latest-and-greatest framework patterns.

---

## Companion Documents

1. **plan.md** (this file) - phases, objectives, decisions
2. **tasks.md** - task checklist per phase
3. **lessons.md** - persistent memory: what worked, what did not, root causes, patterns

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

---

## Goals for Framework v2

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

### Phase 1: Close v1 Gaps and Knowledge Propagation [Status: TODO]

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

### Phase 2: Remove InsecureSkipVerify (G402) [Status: TODO]

**Objective**: Generate real TLS cert chains for all test servers, eliminate TLS bypass.

- Add `NewTestTLSBundle()` to testserver (self-signed CA + server cert)
- Add `TLSClientConfig(t)` helper trusting test CA cert
- Update `testserver.StartAndWait()` to accept/expose TLS bundle
- Migrate all 10 services from `InsecureSkipVerify: true` to `TLSClientConfig(t)`
- Remove `G402` from `gosec.excludes` in `.golangci.yml`
- Uncomment `no-tls-insecure-skip-verify` semgrep rule
- **Success**: Zero `InsecureSkipVerify` in codebase, G402 enabled, all tests pass
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

**Objective**: Reduce 173 `// Sequential:` exemptions by applying SEAM PATTERN and dependency injection.

- Deep-dive each category: viper (58), os.Chdir (37), seam variables (11), pgDriver (11), etc.
- Implement SEAM PATTERN for viper/pflag tests (inject config reader)
- Evaluate os.Chdir exemptions in lint_fitness (CheckInDir already parameterized)
- Inject `io.Writer` for os.Stderr capture tests
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

### Phase 6: lint-fitness Value Assessment [Status: TODO]

**Objective**: Confirm 10,500 lines of lint-fitness truly standardize services.

- Coverage and mutation testing of all 23 sub-linters
- Identify any sub-linters testing synthetic content vs real project files
- Evaluate whether 10,500 lines are justified or can be reduced
- Assess skeleton-template's continued purpose given lint-fitness
- Add test infrastructure rule enforcement (unit-test-starts-server detection)
- **Success**: >=98% coverage, >=95% mutation, documented value assessment
- **Post-Mortem**: lessons.md updated

### Phase 7: PKI-CA Domain Completion [Status: TODO]

**Objective**: Full certificate lifecycle for pki-ca.

- Certificate issuance, renewal, revocation
- CRL distribution, OCSP responder
- CA hierarchy (root > intermediate > issuing)
- **Success**: Full PKI lifecycle tests pass, domain logic complete
- **Post-Mortem**: lessons.md updated

### Phase 8: Identity Services - Aggressive Migration [Status: TODO]

**Objective**: All 5 identity services on latest framework with domain stubs.

- identity-authz: OAuth 2.1 authorization server core
- identity-idp: OIDC provider, user authentication flows
- identity-rp: relying party
- identity-rs: resource server
- identity-spa: single page application
- All using latest builder, all with contract tests, all with auth contract tests
- **Success**: All 5 identity services pass `RunContractTests` including auth contracts
- **Post-Mortem**: lessons.md updated

### Phase 9: Quality and Knowledge Propagation [Status: TODO]

**Objective**: Final quality sweep and knowledge propagation.

- Full coverage and mutation testing enforcement across all services
- Performance benchmarking for crypto operations
- Fix agent semantic commit enforcement
- Propagate ALL lessons to ARCHITECTURE.md, agents, skills, instructions
- Simplify review document format for future iterations
- **Success**: All quality gates pass, all knowledge propagated, clean lessons.md
- **Post-Mortem**: lessons.md updated

---

## Cross-References

- **Framework v1**: `docs/framework-v1/` (plan.md, tasks.md, lessons.md, review.md)
- **Framework Brainstorm**: `docs/framework-brainstorm/` (00-overview through 08-recommendations)
- **Architecture**: `docs/ARCHITECTURE.md` (single source of truth)
- **Migration Priority**: ARCHITECTURE.md Section 2.2 (sm-im > jose-ja > sm-kms > pki-ca > identity)
- **Service Template**: ARCHITECTURE.md Section 5.1 (template pattern), Section 5.2 (builder pattern)
- **Testing Strategy**: ARCHITECTURE.md Section 10 (testing architecture)
- **Quality Gates**: ARCHITECTURE.md Section 11.2 (quality gates)
- **Fitness Functions**: ARCHITECTURE.md Section 9.11 (fitness function catalog)
- **Sequential Exemptions**: ARCHITECTURE.md Section 10.2.5 (sequential test exemption)
