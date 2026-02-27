# Architecture Evolution Plan - fixes-v8

**Status**: 4/10 phases complete, Phase 5 in progress (48 tasks done, 3 in this session; Phases 5-10 active)
**Created**: 2026-02-26
**Updated**: 2026-02-27
**Purpose**: Architecture documentation quality + service-template readiness evaluation + skeleton-template (10th product-service stereotype) + PKI-CA clean-slate + CICD linter enhancements

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

fixes-v8 focuses on five priorities:

1. **Architecture Doc Quality** - Complete cleanup from fixes-v8 analysis (CONFIG-SCHEMA.md created, structural fixes committed)
2. **Service-Template Readiness Evaluation** - Evidence-based assessment of all 9 services against the service template pattern
3. **Skeleton-Template (10th Product-Service)** - Permanent stereotype service demonstrating best-practice use of service-template; empty of business logic except absolute minimum for stereotype purposes
4. **PKI-CA Clean-Slate** - Archive existing pki-ca, create clean-slate skeleton using skeleton-template as starting reference
5. **CICD Linter Enhancements** - New linter rules for PRODUCT and PRODUCT-SERVICE structural best practices

---

## Background

### Prior Work (fixes-v7)
- 220/220 tasks across 11 phases completed (docs/fixes-v7/ deleted — all complete)
- All 13 CI/CD workflows now green
- E2E tests passing for all 4 implemented services (sm-kms, sm-im, jose-ja, identity)
- Propagation marker system (`@source`/`@propagate`) implemented across ARCHITECTURE.md and 18 instruction files
- 13 root causes identified and fixed in E2E test suites

### Current State (commit 0adf04af1)
- ARCHITECTURE.md structural issues fixed
- docs/CONFIG-SCHEMA.md created
- All builds clean, all tests passing, clean working tree
- fixes-v7 deleted (100% complete)

---

## Executive Summary

The cryptoutil project has 5 products containing 10 services. The service-template (`internal/apps/template/service/`) is the reusable base for all services. This plan creates a 10th product-service called **skeleton-template** as a permanent, empty-of-business-logic reference implementation demonstrating best-practice template usage.

### Three-Tier Architecture Vision

| Tier | Directory | Purpose | Changes First? |
|------|-----------|---------|----------------|
| **Base** | `internal/apps/template/service/` | Reusable infrastructure (HTTPS, health, DB, telemetry, barrier, sessions) | Yes (infrastructure) |
| **Stereotype** | `internal/apps/skeleton/template/` | Best-practice demonstration of base usage; permanent 10th product-service | Yes (patterns) |
| **Services** | `internal/apps/{sm,pki,jose,identity}/...` | Business logic services building on base | Roll out after base+stereotype |

### Medium-Term Renames (NOT in fixes-v8 scope)

- `internal/apps/template/service/` → `internal/apps/template/product-service-base/` (or similar)
- `internal/apps/skeleton/template/` → `internal/apps/template/product-service-stereotype/` (or similar)

These renames are tracked for future work but **not executed in fixes-v8** to limit scope.

### Long-Term Workflow

1. Make structural/content change to base or stereotype first
2. Codify validation of the change via CICD linters
3. Create plan.md/tasks.md to roll out change to all 9 (or more) services
4. Execute rollout efficiently with automated validation

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
| **skeleton-template** | **0** | **0** | **Planned** | **Planned** | **Planned** | **New** |

### Port Assignment for skeleton-template

| Service | Public Port Range (Host) | Product Port Range | Suite Port Range | PostgreSQL Host Port |
|---------|--------------------------|-------------------|------------------|---------------------|
| skeleton-template | 8900-8999 | 18900-18999 | 28900-28999 | 54329 |

This is the only remaining available port range in the 8xxx block.

### Template Builder Adoption

ALL 9 existing services use `NewServerBuilder`. The SM/JOSE/PKI services are standalone (each has its own domain + migrations). The identity services share a common domain layer.

### Deployment Infrastructure

All 9 services have deployment directories in `deployments/` and config directories in `configs/`. skeleton-template will follow the same pattern.

### Migration Priority (from ARCHITECTURE.md)

> sm-im -> jose-ja -> sm-kms -> pki-ca -> identity services

The first 4 are already migrated. Identity services are the final frontier. skeleton-template is additive (no migration).

---

## Phase 1: Architecture Documentation Hardening ✅ COMPLETE

**Goal**: Close remaining low-priority items from ARCHITECTURE.md analysis.

### 1.1 Validate Propagation Markers
### 1.2 Long Line Audit
### 1.3 Empty Section Cleanup
### 1.4 Cross-Reference Integrity

**Quality Gate**: ✅ Zero broken links, zero stale propagation markers, all intentional gaps documented.

---

## Phase 2: Service-Template Readiness Evaluation ✅ COMPLETE

**Goal**: Generate evidence-based readiness scores for all 9 services.

### Consolidated Readiness Scorecard

| Dimension | sm-kms | sm-im | jose-ja | pki-ca | id-authz | id-idp | id-rs | id-rp | id-spa |
|-----------|--------|-------|---------|--------|----------|--------|-------|-------|--------|
| **Total** | **50** | **48** | **50** | **44** | **43** | **43** | **40** | **38** | **35** |
| **Grade** | **A** | **A** | **A** | **B+** | **B** | **B** | **C+** | **C** | **C-** |

**Quality Gate**: ✅ All 9 services scored, gaps documented.

---

## Phase 3: Identity Service Alignment Planning ✅ COMPLETE

**Goal**: Plan the migration path for identity services to full template compliance.

**Quality Gate**: ✅ Clear migration plan documented.

---

## Phase 4: Next Architecture Step Execution ✅ COMPLETE

**Goal**: Execute first concrete improvement from Phase 2/3 findings.

**Quality Gate**: ✅ Builds clean, lint 0 issues, 62/62 validators pass, all tests pass.

---

## Phase 5: skeleton-template Product-Service (10th Service)

**Goal**: Create `skeleton-template` as the 10th product-service. Permanent, empty of business logic, demonstrates best-practice use of service-template. On equal footing with all 9 other services throughout ARCHITECTURE.md, deployments, configs, magic constants, CICD, and the entire repository.

### 5.1 Magic Constants
Add to `internal/shared/magic/`: product name `skeleton`, service name `template`, service ID `skeleton-template`, port 8900, PostgreSQL port 54329.

### 5.2 Product-Level Wiring
Create `internal/apps/skeleton/skeleton.go` (product router) and `internal/apps/skeleton/skeleton_test.go`. Pattern: identical to `internal/apps/pki/pki.go`.

### 5.3 Service Entry Point
Create `internal/apps/skeleton/template/template.go` (service CLI handler). Pattern: identical to `internal/apps/jose/ja/ja.go`.

### 5.4 Server Implementation
Create `internal/apps/skeleton/template/server/server.go` using NewServerBuilder pattern. Minimal: dual HTTPS, health endpoints, empty handler registration with dual API paths.

### 5.5 Server Config
Create `internal/apps/skeleton/template/server/config/config.go`. Flat kebab-case YAML parsing with ServiceTemplateServerSettings embedding.

### 5.6 Repository & Migrations
Create `internal/apps/skeleton/template/repository/` with MigrationsFS and empty 2001 placeholder migration. Use `WithDomainMigrations()`.

### 5.7 Domain (Empty)
Create `internal/apps/skeleton/template/domain/` with minimal placeholder model (e.g., `TemplateItem` with ID + tenant_id + created_at).

### 5.8 CMD Entry Point
Create `cmd/skeleton-template/main.go`. Pattern: identical to `cmd/jose-ja/main.go`.

### 5.9 Suite Integration
Update `internal/apps/cryptoutil/cryptoutil.go` to add skeleton product routing.

### 5.10 Deployment Infrastructure
Create `deployments/skeleton-template/` with compose.yml, secrets, and include files. Create `deployments/skeleton/` for product-level deployment. Create `configs/skeleton/` for service config files.

### 5.11 Tests
Create comprehensive tests for all new code: server_test.go, config_test.go, template_test.go, skeleton_test.go. Coverage ≥95%.

### 5.12 E2E Test Skeleton
Create `internal/apps/skeleton/template/e2e/` with testmain_e2e_test.go and basic health check E2E test.

### 5.13 ARCHITECTURE.md Update
Add skeleton-template to: Service Catalog (3.2), Port Assignments (3.4), PostgreSQL Ports (3.4.2), Implementation Status table. Add section 3.2.X for Skeleton product.

### 5.14 Quality Gate Validation
Full validation: build, lint, test, deployment validators, health endpoints respond.

**Quality Gate**: skeleton-template is a fully functional 10th product-service demonstrating service-template best practices. Builds, runs, serves health endpoints, passes all quality gates.

---

## Phase 6: PKI-CA Archive & Clean-Slate Skeleton

**Goal**: Archive existing pki-ca (111 Go files, 27 directories), create new empty pki-ca using skeleton-template as the starting reference. Validates that the skeleton-template pattern is reproducible.

### 6.1 Archive Existing PKI-CA
Move `internal/apps/pki/ca/` to `internal/apps/pki/ca-archived/`. Temporarily stub references.

### 6.2 Create New PKI-CA from Skeleton Pattern
Create new `internal/apps/pki/ca/` by following the exact same patterns established in skeleton-template (Phase 5). The server, config, repository, domain, and tests should follow the same structure.

### 6.3 Wire Entry Points
Reconnect `cmd/pki-ca/main.go` and `internal/apps/pki/pki.go` to new skeleton.

### 6.4 Quality Gate Validation
Full validation: build, lint, test, deployment validators, health endpoints.

**Quality Gate**: New pki-ca skeleton builds, runs, passes all quality gates. Archived code preserved for future porting.

---

## Phase 7: Service-Template Reusability Analysis

**Goal**: Analyze both skeletons (skeleton-template and new pki-ca) to assess service-template reusability and identify improvements.

### 7.1 Minimal File Set Documentation
Document the minimal file set required for a conforming product-service. Compare against sm-kms (reference, 50/50).

### 7.2 Template Friction Points
Catalog friction, boilerplate, or missing abstractions encountered during skeleton creation.

### 7.3 Product-Service Pattern Improvements
Analyze product-level wiring for simplification opportunities.

### 7.4 Enhancement Proposals
Concrete, prioritized proposals for service-template improvements.

**Quality Gate**: Analysis documented in RESEARCH.md with actionable proposals.

---

## Phase 8: CICD Linter Enhancements

**Goal**: Implement new CICD linter rules enforcing structural best practices for all PRODUCT and PRODUCT-SERVICE directories.

### 8.1 Linter Gap Analysis
Compare existing validators against Phase 7 findings.

### 8.2 Validator Design
Design validators for: directory structure, required files, migration numbering, product wiring, test presence.

### 8.3 Validator Implementation
Implement in `cmd/cicd/`. Tests ≥98% coverage, mutation testing.

### 8.4 Apply to All 10 Services
Run against all 10 services (including skeleton-template). Fix non-conformance.

**Quality Gate**: New validators implemented, tested, passing for all 10 services. Zero regressions.

---

## Phase 9: Documentation & Research

**Goal**: Consolidate findings into docs/fixes-v8/RESEARCH.md.

### 9.1 Skeleton-Template Patterns
Document the creation process, minimal file set, patterns to follow.

### 9.2 Service-Template Learnings
Document strengths, weaknesses, proposed improvements.

### 9.3 Identity Future Roadmap
Document planned approach: archive existing identity services, create skeletons, achieve independent deployability. Identity E2E stays shared (ED-10).

### 9.4 Three-Tier Architecture Documentation
Document the base/stereotype/service architecture vision and long-term workflow.

### 9.5 RESEARCH.md Publication
Finalize and commit.

**Quality Gate**: RESEARCH.md complete, cross-referenced with ARCHITECTURE.md.

---

## Phase 10: ARCHITECTURE.md Propagation

**Goal**: Ensure all ARCHITECTURE.md changes from Phase 5 are propagated to instruction files via `@source`/`@propagate` markers.

### 10.1 Validate Propagation
Run `cicd validate-propagation` and `cicd validate-chunks`.

### 10.2 Update Instruction Files
Update any instruction files affected by service catalog changes (02-01.architecture.instructions.md).

### 10.3 Final Quality Gate
Full project validation: build, lint, test, deployment validators, propagation check.

**Quality Gate**: All propagation markers valid. Full project quality gates pass.

---

## Executive Decisions

| # | Decision | Rationale | Status |
|---|----------|-----------|--------|
| ED-1 | All 4 SM/JOSE/PKI services use builder pattern correctly | Confirmed via code review | ✅ Confirmed |
| ED-2 | All 5 identity services also use builder pattern | Confirmed | ✅ Confirmed |
| ED-3 | Identity uses shared domain layer | 44 domain + 47 repo files shared | ✅ Documented |
| ED-4 | Identity migration numbering non-standard (0002-0011) | Evaluated Phase 2 | ✅ Evaluated |
| ED-5 | identity-rp and identity-spa minimal (10 files each) | Scoped Phase 3 | ✅ Scoped |
| ED-6 | **Create skeleton-template as 10th product-service** (quizme Q1=E, expanded) | Permanent stereotype: demonstrates best-practice template usage. No business logic. | ⏳ Phase 5 |
| ED-7 | **Identity services must be independently deployable** (quizme Q2=E) | Each gets own DB, migration range, E2E. Deferred to post-fixes-v8. | 📋 Future |
| ED-8 | **Skeleton uses empty conforming 2001+ migrations** (quizme Q3=E) | Both skeleton-template and new pki-ca use 2001+ range. | ⏳ Phase 5-6 |
| ED-9 | **PKI-CA archive + clean-slate after skeleton-template** (quizme Q4=E, refined) | skeleton-template validates pattern; pki-ca follows same pattern. | ⏳ Phase 5-6 |
| ED-10 | **Identity E2E stays shared** (quizme Q5=A) | Single suite tests all 5 services together. | ✅ Decided |
| ED-11 | **Port 8900-8999 for skeleton-template** | Only remaining port range in 8xxx block; PostgreSQL port 54329. | ⏳ Phase 5 |
| ED-12 | **Product name "skeleton", service name "template"** | Follows PRODUCT-SERVICE pattern (skeleton-template). Short-term name; medium-term rename to product-service-stereotype. | ⏳ Phase 5 |
| ED-13 | **Medium-term renames deferred** | service-template → product-service-base, skeleton-template → product-service-stereotype. NOT in fixes-v8 scope. | 📋 Future |
| ED-14 | **Long-term: change base/stereotype first, validate, roll out** | Changes made in base+stereotype → codified in CICD linters → plan.md/tasks.md for 9-service rollout. | 📋 Future |

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| skeleton-template adds maintenance burden | Low | Low | Zero business logic; validates template patterns automatically via CICD |
| Port 8900 range conflicts with future services | Low | Low | 8900 is last available 8xxx range; next services use 9xxx |
| ARCHITECTURE.md changes break propagation markers | Medium | Medium | Phase 10 validates all propagation |
| Archived pki-ca business logic difficult to port | Medium | Medium | Archive preserves original; skeleton validates pattern first |
| New CICD linters false-positive on existing services | Medium | Medium | Test against all 10 services before merging |
| Identity archive+skeleton reveals deep coupling | High | Low | PKI-CA + skeleton-template validate approach first |
| skeleton-template naming awkward | Low | Low | Medium-term rename to product-service-stereotype planned (ED-13) |

---

## Quality Gates

| Gate | Criteria | Phase |
|------|----------|-------|
| QG-1 | Zero broken links in ARCHITECTURE.md | Phase 1 ✅ |
| QG-2 | Propagation markers valid | Phase 1 ✅ |
| QG-3 | All 9 services scored on 10 dimensions | Phase 2 ✅ |
| QG-4 | Identity migration strategy documented | Phase 3 ✅ |
| QG-5 | At least one service improvement committed | Phase 4 ✅ |
| QG-6 | skeleton-template is functional 10th product-service | Phase 5 |
| QG-7 | PKI-CA clean-slate skeleton builds and runs | Phase 6 |
| QG-8 | RESEARCH.md documents template reusability analysis | Phase 7 |
| QG-9 | New CICD linters implemented, tested (≥98% coverage) | Phase 8 |
| QG-10 | RESEARCH.md published with all findings | Phase 9 |
| QG-11 | All propagation markers valid after ARCHITECTURE.md changes | Phase 10 |

---

## Success Criteria

1. **Documentation**: ARCHITECTURE.md has zero known issues, CONFIG-SCHEMA.md complete
2. **Visibility**: All 9+1 services have quantified readiness scores
3. **Stereotype**: skeleton-template exists as 10th product-service demonstrating best-practice template usage
4. **PKI-CA Clean-Slate**: Archived existing, new skeleton builds and passes quality gates
5. **Template Reusability**: Friction points identified with concrete enhancement proposals
6. **CICD Enhancements**: New linter validators for project structure best practices
7. **Research**: docs/fixes-v8/RESEARCH.md published with all findings
8. **Propagation**: All ARCHITECTURE.md changes propagated to instruction files

---

## ARCHITECTURE.md Cross-References

- [Section 3 Product Suite Architecture](../../docs/ARCHITECTURE.md#3-product-suite-architecture) - Service catalog (skeleton-template added)
- [Section 3.4 Port Assignments](../../docs/ARCHITECTURE.md#34-port-assignments--networking) - Port 8900 for skeleton-template
- [Section 4.4.1 Go Project Structure](../../docs/ARCHITECTURE.md#441-go-project-structure) - Directory layout
- [Section 5.1 Service Template Pattern](../../docs/ARCHITECTURE.md#51-service-template-pattern) - Template components (base tier)
- [Section 5.2 Service Builder Pattern](../../docs/ARCHITECTURE.md#52-service-builder-pattern) - Builder usage, merged migrations
- [Section 7 Data Architecture](../../docs/ARCHITECTURE.md#7-data-architecture) - Migration versioning, multi-tenancy
- [Section 9.1 CLI Patterns & Strategy](../../docs/ARCHITECTURE.md#91-cli-patterns--strategy) - Product-service CLI wiring
- [Section 10 Testing Architecture](../../docs/ARCHITECTURE.md#10-testing-architecture) - Coverage targets, E2E strategy
- [Section 11.2 Quality Gates](../../docs/ARCHITECTURE.md#112-quality-gates) - Enforcement criteria
- [Section 12 Deployment Architecture](../../docs/ARCHITECTURE.md#12-deployment-architecture) - Compose, secrets, validation
- [Section 12.7 Documentation Propagation Strategy](../../docs/ARCHITECTURE.md#127-documentation-propagation-strategy) - Marker system
- [Section 13.1 Coding Standards](../../docs/ARCHITECTURE.md#131-coding-standards) - Code patterns
- [Section 13.6 Plan Lifecycle Management](../../docs/ARCHITECTURE.md#136-plan-lifecycle-management) - Plan management
- [Section 13.7 Infrastructure Blocker Escalation](../../docs/ARCHITECTURE.md#137-infrastructure-blocker-escalation) - Blocker handling
