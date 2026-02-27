# Architecture Evolution Plan - fixes-v8

**Status**: 4/8 phases complete (45 tasks done, 0 in progress, 45 total; Phases 5-8 planned)
**Created**: 2026-02-26
**Updated**: 2026-02-26
**Purpose**: Architecture documentation quality + service-template readiness evaluation + PKI-CA clean-slate skeleton + CICD linter enhancements

---

## Quality Mandate

ALL deliverables MUST satisfy:
- Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`)
- Lint clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`)
- Tests pass (`go test ./... -shuffle=on`, zero skips)
- Coverage >=95% production, >=98% infrastructure/utility
- Mutation testing >=95% where applicable
- Conventional commits, incremental (never amend)
- ARCHITECTURE.md cross-references for all design decisions

---

## Overview

fixes-v8 focuses on four priorities:

1. **Architecture Doc Quality** - Complete cleanup from fixes-v8 analysis (CONFIG-SCHEMA.md created, structural fixes committed, remaining low-priority items)
2. **Service-Template Readiness Evaluation** - Evidence-based assessment of all 9 services against the service template pattern
3. **PKI-CA Clean-Slate Skeleton** - Archive existing pki-ca, create empty skeleton from service-template, validate reusability, identify improvements
4. **CICD Linter Enhancements** - New linter rules for all PRODUCT and PRODUCT-SERVICE to ensure best-practice project structure from the start

---

## Background

### Prior Work (fixes-v7)
- 220/220 tasks across 11 phases completed
- All 13 CI/CD workflows now green
- E2E tests passing for all 4 implemented services (sm-kms, sm-im, jose-ja, identity)
- Propagation marker system (`@source`/`@propagate`) implemented across ARCHITECTURE.md and 18 instruction files
- 13 root causes identified and fixed in E2E test suites

### Current State (commit 00dfeb5bc)
- ARCHITECTURE.md structural issues fixed: missing section 3.2.3, garbled sm-im text, broken file refs, section numbering conflicts
- docs/CONFIG-SCHEMA.md created to resolve broken reference
- All builds clean, all tests passing, clean working tree

---

## Executive Summary

The cryptoutil project has 4 products containing 9 services. Currently 4 services (sm-kms, sm-im, jose-ja, pki-ca) are implemented with the service template builder pattern. The 5 identity services (authz, idp, rs, rp, spa) also use the builder pattern but have a different internal architecture (shared domain layer, shared repository, ServerManager). This plan evaluates readiness and charts the migration path forward.

---

## Technical Context

### Service Implementation Status

| Service | Go Files | Test Files | Builder Pattern | Migrations | E2E Tests | Maturity |
|---------|----------|------------|-----------------|------------|-----------|----------|
| sm-kms | 119 | 78 | ✅ NewServerBuilder | ✅ 2001+ | ✅ | Reference impl |
| sm-im | 60 | 43 | ✅ NewServerBuilder | ✅ 2001+ | ✅ | Production-ready |
| jose-ja | 75 | 54 | ✅ NewServerBuilder | ✅ 2001+ | ✅ | Production-ready |
| pki-ca | 111 | 76 | ✅ NewServerBuilder | ✅ 2001+ | ✅ | Production-ready |
| identity-authz | 133 | 84 | ✅ NewServerBuilder | Shared (0002-0011) | ✅ Shared | Advanced |
| identity-idp | 129 | 74 | ✅ NewServerBuilder | Shared (0002-0011) | ✅ Shared | Advanced |
| identity-rs | 18 | 8 | ✅ NewServerBuilder | Shared | ✅ Shared | Early |
| identity-rp | 10 | 4 | ✅ NewServerBuilder | Shared | ✅ Shared | Minimal |
| identity-spa | 10 | 4 | ✅ NewServerBuilder | Shared | ✅ Shared | Minimal |

### Template Builder Adoption

ALL 9 services use `NewServerBuilder` with the service template pattern. The SM/JOSE/PKI services are standalone (each has its own domain + migrations). The identity services share:
- Common domain model (`internal/apps/identity/domain/` - 44 files, 23 domain test files)
- Common repository (`internal/apps/identity/repository/` - 47 files with 11 migration pairs)
- Common config (`internal/apps/identity/config/`)
- ServerManager for multi-service orchestration
- Per-service server packages (`authz/server/`, `idp/server/`, `rs/server/`, `rp/server/`, `spa/server/`)

### Deployment Infrastructure

All 9 services have deployment directories in `deployments/` and config directories in `configs/`. E2E infra exists for sm-kms, sm-im, jose-ja, and identity (shared across all 5 identity services).

### Migration Priority (from ARCHITECTURE.md)

> sm-im -> jose-ja -> sm-kms -> pki-ca -> identity services

The first 4 are already migrated. Identity services are the final frontier.

---

## Phase 1: Architecture Documentation Hardening

**Goal**: Close remaining low-priority items from ARCHITECTURE.md analysis.

### 1.1 Validate Propagation Markers
Verify all `@source` / `@propagate` markers in instruction files match ARCHITECTURE.md content byte-for-byte after the structural fixes in 00dfeb5bc.

### 1.2 Long Line Audit
Review the 68 lines >200 chars outside code blocks. Table rows are acceptable; fix any non-table long lines.

### 1.3 Empty Section Cleanup
Review 58 empty sections identified. Most are structural placeholders - confirm intentional vs. incomplete. Document any gaps.

### 1.4 Cross-Reference Integrity
Run full anchor validation against the updated ARCHITECTURE.md to confirm zero broken internal links post-fix.

**Quality Gate**: Zero broken links, zero stale propagation markers, all intentional gaps documented.

---

## Phase 2: Service-Template Readiness Evaluation

**Goal**: Generate evidence-based readiness scores for all 9 services.

### 2.1 Evaluation Criteria
Score each service on 10 dimensions (1-5 scale):
1. Builder pattern adoption (NewServerBuilder + Build + ServiceResources)
2. Domain migrations (2001+ range, merged FS)
3. OpenAPI spec (strict server, generated code)
4. Dual HTTPS endpoints (public + admin)
5. Health checks (livez, readyz, shutdown)
6. Dual API paths (/service/** + /browser/**)
7. Test coverage (unit + integration + E2E)
8. Deployment infrastructure (compose + config + secrets)
9. Telemetry integration (OTLP, structured logging)
10. Multi-tenancy (tenant_id scoping)

### 2.2 SM Services Assessment
Deep audit of sm-kms, sm-im alignment and consistency.

### 2.3 JOSE Service Assessment
Deep audit of jose-ja for pattern compliance.

### 2.4 PKI Service Assessment
Deep audit of pki-ca for pattern compliance.

### 2.5 Identity Services Assessment
Deep audit of all 5 identity services. Key questions:
- Is the shared domain/repository pattern compatible with per-service template builder?
- Are the identity migrations in the correct range (should be 2001+ not 0002-0011)?
- Does ServerManager need refactoring to align with template lifecycle?

**Quality Gate**: All 9 services scored, gaps documented, alignment issues identified with remediation paths.

---

## Phase 3: Identity Service Alignment Planning ✅ COMPLETE

**Goal**: Plan the migration path for identity services to full template compliance.

### 3.1 Migration Numbering
Identity migrations use 0002-0011 instead of the mandated 2001+ range. Plan renumbering strategy.

### 3.2 Service Separation Analysis
Determine if identity services should remain monolithic (shared domain) or be split into true standalone services per the template pattern.

### 3.3 rp/spa Buildout Scoping
identity-rp (10 files) and identity-spa (10 files) are minimal. Plan their buildout to match authz/idp maturity levels.

### 3.4 E2E Test Decomposition
Currently identity E2E tests are shared across all 5 services. Plan per-service E2E decomposition if needed.

**Quality Gate**: Clear migration plan with task breakdown, risk assessment, and dependency map.

---

## Phase 4: Next Architecture Step Execution ✅ COMPLETE

**Goal**: Execute the first concrete improvement from Phase 2/3 findings.

### 4.1 Quick Wins
Apply any alignment fixes that don't require structural changes (config normalization, missing health endpoints, telemetry gaps).

### 4.2 First Migration Task
Execute the highest-priority migration task from Phase 3 (likely migration renumbering or E2E decomposition).

### 4.3 Validation
Full quality gate validation: builds clean, tests pass, E2E green, deployment validators pass.

**Quality Gate**: At least one concrete service improvement committed with evidence.

---

## Phase 5: PKI-CA Archive & Clean-Slate Skeleton

**Goal**: Archive existing pki-ca (111 Go files, 27 directories), create a new empty pki-ca skeleton using sm-kms/sm-im/jose-ja as reference. Validate it builds, runs, and passes all quality gates. This serves as a clean slate for future porting of archived pki-ca business logic.

### 5.1 Archive Existing PKI-CA
Move `internal/apps/pki/ca/` to `internal/apps/pki/ca-archived/`. Update imports in `cmd/pki-ca/main.go` and `internal/apps/pki/pki.go` to temporarily disable. Ensure project still builds (with pki-ca excluded from active compilation).

### 5.2 Create Empty PKI-CA Skeleton
Create new `internal/apps/pki/ca/` following the sm-kms/sm-im/jose-ja directory structure:
- `server/` - Server with NewServerBuilder, dual HTTPS, health checks
- `server/config/` - Config parsing (flat kebab-case YAML)
- `server/apis/` or `server/handler/` - Empty handler registration
- `repository/` - Empty repository layer
- `repository/migrations/` - Empty 2001+ migration placeholder
- `domain/` - Empty domain models
- `e2e/` - Minimal E2E test

### 5.3 Wire Up Entry Points
Reconnect `cmd/pki-ca/main.go` and `internal/apps/pki/pki.go` to the new skeleton. Ensure `go build ./cmd/pki-ca/...` succeeds.

### 5.4 Quality Gate Validation
Verify: builds clean, lint clean, tests pass, deployment validators pass, health endpoints respond, dual HTTPS works.

**Quality Gate**: Empty pki-ca skeleton builds, runs, serves health endpoints, passes all quality gates. Existing pki-ca safely archived.

---

## Phase 6: Service-Template Reusability Analysis

**Goal**: Analyze the empty pki-ca skeleton to assess service-template reusability. Identify improvements to product-service patterns and to the template itself.

### 6.1 Skeleton Structure Analysis
Document the minimal file set required for a conforming service. Compare against sm-kms (reference implementation, 50/50 score).

### 6.2 Template Friction Points
Identify friction, boilerplate, or missing abstractions encountered while creating the skeleton. Catalog what the template provides vs what each new service must implement.

### 6.3 Product-Service Pattern Improvements
Analyze product-level wiring (`internal/apps/pki/pki.go`, `cmd/pki-ca/main.go`) for patterns that could be simplified or templated.

### 6.4 Service-Template Enhancement Proposals
Propose concrete enhancements to the service template that would reduce the effort for future new services.

**Quality Gate**: Analysis documented in RESEARCH.md with actionable enhancement proposals.

---

## Phase 7: CICD Linter Enhancements

**Goal**: Identify and implement new CICD linter rules that enforce best practices for all PRODUCT and PRODUCT-SERVICE directories, ensuring new projects are created correctly from the start.

### 7.1 Linter Gap Analysis
Compare existing `cicd lint-deployments` validators against the patterns discovered in Phase 6. Identify missing structural validators.

### 7.2 New Validator Design
Design new validators for: service directory structure, required files (server.go, config, migrations dir), migration numbering conformance, product-level wiring, test file presence.

### 7.3 Validator Implementation
Implement the new validators in `cmd/cicd/` following existing lint-deployments patterns. Include comprehensive tests (≥98% coverage, mutation testing).

### 7.4 Apply and Verify
Run new validators against all 9 services. Fix any non-conformance discovered. Ensure all existing tests still pass.

**Quality Gate**: New validators implemented, tested, passing for all services. Zero regressions.

---

## Phase 8: Documentation & Research

**Goal**: Document all findings, learnings, and patterns from the clean-slate exercise in docs/fixes-v8/RESEARCH.md.

### 8.1 PKI-CA Skeleton Patterns
Document the minimal service structure, what was trivial vs what required effort, and patterns to follow for future services.

### 8.2 Service-Template Learnings
Document what the template provides well, what's missing, and proposed improvements.

### 8.3 Identity Future Roadmap
Document the planned approach for identity services: archive existing, create similar skeletons, achieve independent deployability with own DB/migrations/E2E. Note: identity E2E stays shared for now (per quizme-v1 Q5 decision).

### 8.4 RESEARCH.md Publication
Consolidate all findings into `docs/fixes-v8/RESEARCH.md`.

**Quality Gate**: RESEARCH.md complete, actionable, and cross-referenced with ARCHITECTURE.md.

---

## Executive Decisions

| # | Decision | Rationale | Status |
|---|----------|-----------|--------|
| ED-1 | All 4 SM/JOSE/PKI services use builder pattern correctly | Confirmed via code review of server.go files | ✅ Confirmed |
| ED-2 | All 5 identity services also use builder pattern | Confirmed: authz/idp/rs/rp/spa all import NewServerBuilder | ✅ Confirmed |
| ED-3 | Identity uses shared domain layer (not standalone per-service) | 44 domain files + 47 repo files shared across 5 services | ✅ Documented |
| ED-4 | Identity migration numbering non-standard (0002-0011 vs 2001+) | Needs evaluation: may conflict with template migration range | ✅ Evaluated Phase 2 |
| ED-5 | identity-rp and identity-spa are minimal (10 files each) | Need buildout plan before they can be considered production-ready | ✅ Scoped Phase 3 |
| ED-6 | **Archive pki-ca, create clean-slate skeleton** (quizme Q1=E) | Validate service-template reusability before porting business logic. Use sm-kms/sm-im/jose-ja as reference. | ⏳ Phase 5 |
| ED-7 | **Identity services must be independently deployable** (quizme Q2=E) | Each service gets own logical DB, migration range, E2E tests. Deferred to post-pki-ca work. | 📋 Future |
| ED-8 | **Skeleton pki-ca uses empty conforming migrations** (quizme Q3=E) | Empty 2001+ migrations conforming to existing product-service patterns. No renumbering of archived code. | ⏳ Phase 5 |
| ED-9 | **PKI-CA skeleton first, identity later** (quizme Q4=E) | Validate template reusability on pki-ca, then apply same archive+skeleton approach to identity services. | ⏳ Phase 5-8 |
| ED-10 | **Identity E2E stays shared** (quizme Q5=A) | Single E2E suite tests all 5 identity services together. Simpler, tests interactions. | ✅ Decided |

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Identity migration renumbering breaks existing deployments | High | Medium | Careful down-migration testing, staged rollout |
| Identity shared domain is architecturally incompatible with template | High | Low | All 5 services already use NewServerBuilder; shared domain is additive |
| rp/spa buildout scope underestimated | Medium | Medium | Phase 2 assessment will quantify gaps accurately |
| Propagation marker staleness after structural fixes | Low | Medium | Phase 1.1 validates all markers |
| Archived pki-ca business logic difficult to port to skeleton | Medium | Medium | Archive preserves original structure; skeleton validates template pattern first |
| Service-template missing abstractions for pki-ca domain | Medium | Low | Phase 6 analysis identifies gaps before porting |
| New CICD linters produce false positives on existing services | Medium | Medium | Test against all 9 services before merging |
| Identity archive+skeleton approach reveals deep coupling | High | Low | PKI-CA validates approach first; identity deferred per ED-9 |

---

## Quality Gates

| Gate | Criteria | Phase |
|------|----------|-------|
| QG-1 | Zero broken links in ARCHITECTURE.md | Phase 1 ✅ |
| QG-2 | Propagation markers valid | Phase 1 ✅ |
| QG-3 | All 9 services scored on 10 dimensions | Phase 2 ✅ |
| QG-4 | Identity migration strategy documented | Phase 3 ✅ |
| QG-5 | At least one service improvement committed | Phase 4 ✅ |
| QG-6 | Empty pki-ca skeleton builds, runs, passes quality gates | Phase 5 |
| QG-7 | RESEARCH.md documents template reusability analysis | Phase 6 |
| QG-8 | New CICD linters implemented, tested (≥98% coverage) | Phase 7 |
| QG-9 | RESEARCH.md published with all findings | Phase 8 |

---

## Success Criteria

1. **Documentation**: ARCHITECTURE.md has zero known issues, CONFIG-SCHEMA.md complete
2. **Visibility**: All 9 services have quantified readiness scores
3. **Alignment**: SM/JOSE/PKI services confirmed consistent and efficient
4. **Roadmap**: Clear, prioritized migration path for identity services
5. **Clean-Slate PKI-CA**: Empty skeleton builds, runs, passes all quality gates
6. **Template Reusability**: Service-template friction points identified with concrete enhancement proposals
7. **CICD Enhancements**: New linter validators for project structure best practices
8. **Research**: docs/fixes-v8/RESEARCH.md published with all findings and learnings

---

## ARCHITECTURE.md Cross-References

- [Section 3 Product Suite Architecture](../../docs/ARCHITECTURE.md#3-product-suite-architecture) - Service catalog
- [Section 4.4.1 Go Project Structure](../../docs/ARCHITECTURE.md#441-go-project-structure) - Directory layout for new skeleton
- [Section 5.1 Service Template Pattern](../../docs/ARCHITECTURE.md#51-service-template-pattern) - Template components
- [Section 5.2 Service Builder Pattern](../../docs/ARCHITECTURE.md#52-service-builder-pattern) - Builder usage, merged migrations
- [Section 7 Data Architecture](../../docs/ARCHITECTURE.md#7-data-architecture) - Migration versioning, multi-tenancy
- [Section 9.1 CLI Patterns & Strategy](../../docs/ARCHITECTURE.md#91-cli-patterns--strategy) - Product-service CLI wiring
- [Section 10 Testing Architecture](../../docs/ARCHITECTURE.md#10-testing-architecture) - Coverage targets, E2E strategy
- [Section 11.2 Quality Gates](../../docs/ARCHITECTURE.md#112-quality-gates) - Enforcement criteria
- [Section 12 Deployment Architecture](../../docs/ARCHITECTURE.md#12-deployment-architecture) - Compose, secrets, validation
- [Section 12.7 Documentation Propagation Strategy](../../docs/ARCHITECTURE.md#127-documentation-propagation-strategy) - Marker system
- [Section 13.1 Coding Standards](../../docs/ARCHITECTURE.md#131-coding-standards) - Code patterns for skeleton
- [Section 13.7 Infrastructure Blocker Escalation](../../docs/ARCHITECTURE.md#137-infrastructure-blocker-escalation) - Blocker handling
