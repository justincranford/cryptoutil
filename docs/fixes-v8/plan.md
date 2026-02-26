# Architecture Evolution Plan - fixes-v8

**Status**: 3/4 phases complete (38 tasks done, 0 in progress, ~45 total estimated)
**Created**: 2026-02-26
**Updated**: 2026-02-26
**Purpose**: Architecture documentation quality + service-template readiness evaluation + next-service planning

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

fixes-v8 focuses on three priorities:

1. **Architecture Doc Quality** - Complete cleanup from fixes-v8 analysis (CONFIG-SCHEMA.md created, structural fixes committed, remaining low-priority items)
2. **Service-Template Readiness Evaluation** - Evidence-based assessment of all 9 services against the service template pattern
3. **Next-Service Planning** - Determine which service(s) to build/migrate next based on readiness data

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

## Phase 3: Identity Service Alignment Planning

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

## Phase 4: Next Architecture Step Execution

**Goal**: Execute the first concrete improvement from Phase 2/3 findings.

### 4.1 Quick Wins
Apply any alignment fixes that don't require structural changes (config normalization, missing health endpoints, telemetry gaps).

### 4.2 First Migration Task
Execute the highest-priority migration task from Phase 3 (likely migration renumbering or E2E decomposition).

### 4.3 Validation
Full quality gate validation: builds clean, tests pass, E2E green, deployment validators pass.

**Quality Gate**: At least one concrete service improvement committed with evidence.

---

## Executive Decisions

| # | Decision | Rationale | Status |
|---|----------|-----------|--------|
| ED-1 | All 4 SM/JOSE/PKI services use builder pattern correctly | Confirmed via code review of server.go files | ✅ Confirmed |
| ED-2 | All 5 identity services also use builder pattern | Confirmed: authz/idp/rs/rp/spa all import NewServerBuilder | ✅ Confirmed |
| ED-3 | Identity uses shared domain layer (not standalone per-service) | 44 domain files + 47 repo files shared across 5 services | ✅ Documented |
| ED-4 | Identity migration numbering non-standard (0002-0011 vs 2001+) | Needs evaluation: may conflict with template migration range | ⚠️ Evaluate in Phase 2 |
| ED-5 | identity-rp and identity-spa are minimal (10 files each) | Need buildout plan before they can be considered production-ready | ⚠️ Phase 3 |

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Identity migration renumbering breaks existing deployments | High | Medium | Careful down-migration testing, staged rollout |
| Identity shared domain is architecturally incompatible with template | High | Low | All 5 services already use NewServerBuilder; shared domain is additive |
| rp/spa buildout scope underestimated | Medium | Medium | Phase 2 assessment will quantify gaps accurately |
| Propagation marker staleness after structural fixes | Low | Medium | Phase 1.1 validates all markers |

---

## Quality Gates

| Gate | Criteria | Phase |
|------|----------|-------|
| QG-1 | Zero broken links in ARCHITECTURE.md | Phase 1 |
| QG-2 | Propagation markers valid | Phase 1 |
| QG-3 | All 9 services scored on 10 dimensions | Phase 2 |
| QG-4 | Identity migration strategy documented | Phase 3 |
| QG-5 | At least one service improvement committed | Phase 4 |

---

## Success Criteria

1. **Documentation**: ARCHITECTURE.md has zero known issues, CONFIG-SCHEMA.md complete
2. **Visibility**: All 9 services have quantified readiness scores
3. **Alignment**: SM/JOSE/PKI services confirmed consistent and efficient
4. **Roadmap**: Clear, prioritized migration path for identity services
5. **Execution**: At least one Phase 4 improvement shipped

---

## ARCHITECTURE.md Cross-References

- [Section 3 Product Suite Architecture](../../docs/ARCHITECTURE.md#3-product-suite-architecture) - Service catalog
- [Section 5.1 Service Template Pattern](../../docs/ARCHITECTURE.md#51-service-template-pattern) - Template components
- [Section 5.2 Service Builder Pattern](../../docs/ARCHITECTURE.md#52-service-builder-pattern) - Builder usage, merged migrations
- [Section 7 Data Architecture](../../docs/ARCHITECTURE.md#7-data-architecture) - Migration versioning, multi-tenancy
- [Section 10 Testing Architecture](../../docs/ARCHITECTURE.md#10-testing-architecture) - Coverage targets, E2E strategy
- [Section 11.2 Quality Gates](../../docs/ARCHITECTURE.md#112-quality-gates) - Enforcement criteria
- [Section 12 Deployment Architecture](../../docs/ARCHITECTURE.md#12-deployment-architecture) - Compose, secrets, validation
- [Section 12.7 Documentation Propagation Strategy](../../docs/ARCHITECTURE.md#127-documentation-propagation-strategy) - Marker system
