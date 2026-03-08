# Framework v2 - Iteration Plan

**Status**: PLANNING
**Created**: 2026-03-08
**Depends On**: `docs/framework-v1/` (complete), `docs/framework-brainstorm/08-recommendations.md`
**Purpose**: Iterate on the framework established in v1, iterate on product-services, and scale up volume and speed of completing all products-services.

---

## Companion Documents

1. **plan.md** (this file) - phases, objectives, decisions
2. **tasks.md** - task checklist per phase (created when planning begins)
3. **lessons.md** - persistent memory: what worked, what did not, root causes, patterns

---

## Context: Where We Are After Framework v1

### What Framework v1 Delivered

1. **ServiceServer interface** - compile-time contract for all 10 services (11 methods)
2. **Builder auto-defaults** - `Build()` auto-configures JWTAuth + StrictServer, services declare only add-ons
3. **23 fitness sub-linters** - automated ARCHITECTURE.md enforcement via `cicd lint-fitness`
4. **Cross-service contract tests** - `RunContractTests(t, server)` for behavioral consistency
5. **Shared test infrastructure** - testdb, testserver, fixtures, assertions, healthclient
6. **air live reload** - `SERVICE=sm-im air` for 2-3x faster dev loop

### What Framework v1 Did NOT Do

1. **No GitHub Workflows updated** - `lint-fitness` only runs via pre-commit, not CI
2. **Identity services** - only got compile-time assertions, no contract tests, minimal conformance work
3. **No skeleton CRUD reference** - skeleton-template is still minimal
4. **No `cicd new-service` scaffolding** - manual service creation still required
5. **No auth contract tests** - 401 rejection tests deferred to service-specific tests

### Service Maturity After v1

| Service | Interface | Contract Tests | Builder Simplified | Domain Logic | Migration Status |
|---------|-----------|---------------|-------------------|-------------|-----------------|
| sm-im |  |  |  (already was) |  Working CRUD | Complete |
| jose-ja |  |  |  (already was) |  Working CRUD | Complete |
| sm-kms |  |  |  (v1 simplified) |  Working CRUD | Complete |
| pki-ca |  |  |  (already was) | WARN Partial | In Progress |
| skeleton-template |  |  |  (already was) | WARN Minimal | Reference only |
| identity-authz |  | No |  | No Stub | Not Started |
| identity-idp |  | No |  | No Stub | Not Started |
| identity-rp |  | No |  | No Stub | Not Started |
| identity-rs |  | No |  | No Stub | Not Started |
| identity-spa |  | No |  | No Stub | Not Started |

---

## Goals for Framework v2

### Goal 0: Tooling & Security Infrastructure

Foundation work that unblocks Goals 1-3:

- [x] **Semgrep in pre-commit** - `.semgrep/rules/` directory, initial rules for `DisableKeepAlives` and per-test DB violations
- [ ] **Remove InsecureSkipVerify (G402)** - Generate TLS cert chains in `service-template/testing/testserver`, expose CA cert for test clients, remove `InsecureSkipVerify: true` across all 10 services, remove G402 from `gosec.excludes` in `.golangci.yml`

### Goal 1: Framework Iteration - Close v1 Gaps

Iterate on the framework to close gaps identified in v1:

- [ ] **CI/CD workflow for lint-fitness** - `ci-fitness.yml` GitHub Actions workflow
- [ ] **Contract tests for remaining 6 services** - identity-authz/idp/rp/rs/spa + pki-ca (if not already)
- [ ] **Auth contract tests** - 401/403 rejection tests that work across services
- [ ] **Fitness function coverage/mutation verification** - ensure 10,500 lines of lint_fitness meet quality gates

### Goal 2: Product-Service Iteration - Build Domain Logic

Iterate on product-services to implement actual domain logic:

- [ ] **pki-ca domain completion** - certificate issuance, revocation, CRL, OCSP
- [ ] **identity-authz** - OAuth 2.1 authorization server
- [ ] **identity-idp** - identity provider (OIDC)
- [ ] **identity-rp** - relying party
- [ ] **identity-rs** - resource server
- [ ] **identity-spa** - single page application
- [ ] **skeleton-template CRUD reference** - full domain layer for reference (deferred from v1 P0-3)

### Goal 3: Scale - Increase Velocity

Scale up the volume and speed of completing all products-services:

- [ ] **Evaluate `cicd new-service`** - if identity services share enough structure, scaffolding tool may pay off
- [ ] **Parallel service development** - identify which identity services can be developed concurrently
- [ ] **Migration priority** - follow ARCHITECTURE.md migration priority: sm-im > jose-ja > sm-kms > pki-ca > identity services
- [ ] **Batch operations** - identify patterns that apply to all identity services simultaneously

---

## Open Questions (To Be Resolved During Planning)

1. **Migration priority within identity**: Which identity service should be implemented first? authz (foundation for others) or idp (most complex)?
2. **Skeleton CRUD reference**: Is it worth building now that contract tests exist, or skip again?
3. **CI/CD workflow scope**: Should lint-fitness have its own workflow or be added to existing ci-quality?
4. **Auth middleware standardization**: How to make the auth middleware configurable enough for contract tests?
5. **Identity services concurrency**: Can authz and idp be developed in parallel, or is authz a prerequisite for idp?

---

## Phases (Draft - To Be Refined)

### Phase 1: Close v1 Gaps

- Add CI workflow for lint-fitness
- Integrate contract tests into remaining services
- Verify lint_fitness coverage/mutation

### Phase 2: Remove InsecureSkipVerify (G402)

**Prerequisite for all service integration/E2E tests without TLS bypass.**

- Add `NewTestTLSBundle()` to `internal/apps/template/service/testing/testserver/` that generates a self-signed CA + server cert pair at test startup
- Add `TLSClientConfig(t)` helper that returns `*tls.Config` trusting the test CA cert (replaces `InsecureSkipVerify: true`)
- Update `testserver.StartAndWait()` to accept and expose the TLS bundle
- For each of the 10 services: replace `InsecureSkipVerify: true` test HTTP clients with `TLSClientConfig(t)`
- Remove `G402` from `gosec.excludes` in `.golangci.yml`
- Uncomment `no-tls-insecure-skip-verify` rule in `.semgrep/rules/go-testing.yml`
- **Success**: `golangci-lint run` passes with G402 enabled, zero InsecureSkipVerify in production code, all integration/E2E tests pass with real TLS
- **Post-Mortem**: lessons.md updated

### Phase 3: PKI-CA Domain Completion

- Certificate issuance, renewal, revocation
- CRL distribution, OCSP responder
- CA hierarchy (root > intermediate > issuing)

### Phase 4: Identity Foundation (authz)

- OAuth 2.1 authorization server core
- Token issuance, validation, introspection
- Client registration

### Phase 5: Identity Provider (idp)

- OIDC provider
- User authentication flows
- Session management

### Phase 6: Identity Services (rp, rs, spa)

- Relying party implementation
- Resource server implementation
- Single page application

### Phase 7: Quality & Polish

- Coverage and mutation testing enforcement
- Performance benchmarking
- Documentation updates

---

## Cross-References

- **Framework v1**: `docs/framework-v1/` (plan.md, tasks.md, lessons.md, review.md)
- **Framework Brainstorm**: `docs/framework-brainstorm/` (00-overview through 08-recommendations)
- **Architecture**: `docs/ARCHITECTURE.md` (single source of truth)
- **Migration Priority**: ARCHITECTURE.md Section 2.2 (sm-im > jose-ja > sm-kms > pki-ca > identity)
