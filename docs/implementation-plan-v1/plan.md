# Implementation Plan - Deployment Architecture Refactoring

**Status**: Planning
**Created**: 2026-02-17
**Last Updated**: 2026-02-17
**Purpose**: Refactor deployment structure to align with SUITE/PRODUCT/SERVICE hierarchy and consolidate E2E testing patterns

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation
- ❌ **Premature Completion**: NEVER mark complete without verification

**ALL issues are blockers - NO exceptions:**

- ✅ **Fix issues immediately** - When unknowns discovered, blockers identified, unit/integration/E2E/mutations/fuzz/bench/race/SAST/DAST/load/any tests fail, or quality gates are not met, STOP and address
- ✅ **Treat as BLOCKING** - ALL issues block progress to next phase or task
- ✅ **Document root causes** - Root cause analysis is part of planning AND implementation, not optional
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER de-prioritize quality** - Evidence-based verification is ALWAYS highest priority

**Rationale**: Maintaining maximum quality prevents cascading failures and rework.

## Overview

**Problem**: Current deployment structure has inconsistencies:
1. `deployments/cryptoutil/` doesn't indicate it's SUITE-level
2. `deployments/compose/` breaks hierarchical SUITE→PRODUCT→SERVICE pattern
3. No standardized SERVICE/PRODUCT/SUITE E2E compose files
4. Legacy `internal/test/e2e/` doesn't maximize service-template reuse

**Solution**: Refactor to clear three-tier deployment structure:
- `deployments/cryptoutil-suite/` - SUITE-level (all 9 services, port 28XXX)
- `deployments/cryptoutil-product/` - PRODUCT-level E2E (per-product testing, port 18XXX)
- `deployments/cryptoutil-service/` - SERVICE-level E2E (per-service testing, port 8XXX)

**Scope**:
- 9 services across 5 products
- 3 deployment levels (SERVICE, PRODUCT, SUITE)
- All E2E tests migrated to service-template patterns
- ARCHITECTURE.md and related documentation updated
- All quality gates passing (linting, tests, coverage, mutation)

## Background

Current deployment architecture follows three-tier hierarchy (SERVICE → PRODUCT → SUITE) per [ARCHITECTURE.md Section 12.3.4](../../docs/ARCHITECTURE.md#1234-multi-level-deployment-hierarchy):

**Existing Structure**:
- **SERVICE-level**: `deployments/{PRODUCT}-{SERVICE}/` (9 directories) - ✅ Correct
- **PRODUCT-level**: `deployments/{PRODUCT}/` (5 directories) - ✅ Correct
- **SUITE-level**: `deployments/cryptoutil/` - ⚠️ Naming doesn't indicate level
- **E2E Testing**: `deployments/compose/` - ❌ Breaks hierarchy pattern

**E2E Test Patterns**:
- **cipher-im**: Uses SERVICE-level compose, maximizes service-template reuse ✅
- **identity**: Uses PRODUCT-level compose, maximizes service-template reuse ✅
- **Legacy**: Uses `deployments/compose/`, does NOT follow hierarchy ❌

## Technical Context

- **Language**: Go 1.25.5
- **Framework**: Service template pattern (`internal/apps/template/`)
- **Docker Compose**: Multi-level includes with rigid delegation
- **Port Ranges**: SERVICE (8XXX), PRODUCT (18XXX), SUITE (28XXX)
- **E2E Helper**: `internal/apps/template/testing/e2e/ComposeManager`
- **Dependencies**: All services use service-template infrastructure

## Executive Summary

**Critical Context**:
- Three-tier deployment hierarchy is architecturally mandated
- Port offset strategy prevents conflicts (SERVICE+0, PRODUCT+10000, SUITE+20000)
- Service-template provides reusable E2E infrastructure (`ComposeManager`)
- All PRODUCT-SERVICE implementations MUST maximize service-template reuse

**Assumptions & Risks**:
- Assumption: Existing E2E tests are comprehensive and passing
- Assumption: Port ranges 8XXX/18XXX/28XXX don't conflict with other services
- Risk: Breaking existing E2E tests during migration
- Risk: Docker Compose include behavior changes between levels
- Mitigation: Incremental refactoring with continuous testing
- Mitigation: Comprehensive E2E test suite validates all levels

## Phases

### Phase 1: Discovery & Analysis (4h) [Status: ☐ TODO]
**Objective**: Comprehensive analysis of current deployment structure and E2E test patterns

**Tasks**:
- Inventory all compose files in `deployments/` (SERVICE, PRODUCT, SUITE, compose, template)
- Analyze all E2E test locations and patterns (`internal/apps/*/e2e/`, `internal/test/e2e/`)
- Document current port assignments and validate against ranges
- Identify all references to `deployments/compose/` and `deployments/cryptoutil/`
- Analyze service-template E2E helper usage patterns
- Document CI/CD workflows using deployment compose files

**Success Criteria**:
- Complete inventory in `test-output/phase1/deployment-inventory.txt`
- Port validation report in `test-output/phase1/port-validation.txt`
- Reference analysis in `test-output/phase1/reference-analysis.txt`
- E2E pattern analysis in `test-output/phase1/e2e-patterns.txt`

### Phase 2: Create New Directory Structure (3h) [Status: ☐ TODO]
**Objective**: Create new deployment directories with correct naming

**Tasks**:
- Create `deployments/cryptoutil-suite/` directory
- Create `deployments/cryptoutil-product/` directory  
- Create `deployments/cryptoutil-service/` directory
- Copy existing files to new locations (NO modifications yet)
- Validate directory structure with `cicd lint-deployments generate-listings`

**Success Criteria**:
- New directories exist with placeholder compose files
- No linting errors on new structure
- Original directories still intact (safe rollback)

### Phase 3: SUITE-Level Refactoring (5h) [Status: ☐ TODO]
**Objective**: Migrate `deployments/cryptoutil/` → `deployments/cryptoutil-suite/`

**Tasks**:
- Copy compose.yml to `deployments/cryptoutil-suite/compose.yml`
- Update port range validation in `internal/cmd/cicd/lint_deployments/validate_ports.go`
- Update deployment directory lists in `lint_required_contents_deployments.go`
- Migrate secrets directory to `deployments/cryptoutil-suite/secrets/`
- Update ARCHITECTURE.md references (Section 12.3.4)
- Update copilot instructions (02-01.architecture.instructions.md)
- Run linter validation: `go run ./cmd/cicd lint-deployments validate-all`
- Test SUITE deployment: `cd deployments/cryptoutil-suite && docker compose up -d`

**Success Criteria**:
- SUITE deployment works (all 9 services start)
- Health checks pass for all services
- Linting passes: `go run ./cmd/cicd lint-deployments validate-all`
- Port validation confirms 28XXX range usage

(Phases 4-13 continue in same format - truncated for brevity)

## Success Criteria

- [ ] All phases complete (13 phases)
- [ ] New directory structure in place
- [ ] Legacy directories archived
- [ ] All quality gates passing
- [ ] Documentation complete
- [ ] CI/CD workflows green
- [ ] Evidence archived
