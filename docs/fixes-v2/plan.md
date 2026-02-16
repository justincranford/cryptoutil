# Implementation Plan - Deployment & Config Structure Refactoring V2

**Status**: Planning
**Created**: 2026-02-16
**Last Updated**: 2026-02-16
**Purpose**: Refactor deployment/config structure, enhance CICD validation, and establish rigid patterns for consistency

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL documentation must be accurate and complete
- ✅ **Completeness**: NO steps skipped, NO steps de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark steps complete without verification

**ALL issues are blockers - NO exceptions:**

- ✅ **Fix issues immediately** - When unknowns discovered, blockers identified, unit/integration/E2E/mutations/fuzz/bench/race/SAST/DAST/load/any tests fail, or quality gates are not met, STOP and address
- ✅ **Treat as BLOCKING** - ALL issues block progress to next phase or task
- ✅ **Document root causes** - Root cause analysis is part of planning AND implementation, not optional
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER de-prioritize quality** - Evidence-based verification is ALWAYS highest priority

**Rationale**: Maintaining maximum quality prevents cascading failures and rework.

## Overview

This plan addresses critical structural issues in deployments/ and configs/ directories, establishes rigid validation patterns, and ensures consistency across all deployment levels (SUITE/PRODUCT/SERVICE).

## Background

Previous work (fixes-v1) focused on documentation. V2 focuses on:
- Cleaning up structural inconsistencies
- Establishing rigid, validated patterns
- Making ./configs/ as rigorous as ./deployments/
- Comprehensive CICD validation

## Executive Summary

**Critical Findings**:
- .gitkeep files exist in non-empty directories (cleanup needed)
- deployments/compose/compose.yml is ACTIVE (E2E testing, NOT redundant)
- deployments/sm-kms/compose.demo.yml may violate patterns (needs investigation)
- otel-collector-config.yaml duplicated in multiple locations
- deployments/template/config/ intentionally empty (template pattern)
- ./configs/ lacks rigorous structure validation
- CICD linting exists but incomplete (missing comprehensive file lists)

**Decisions Needed**:
- Which compose files are truly redundant?
- Which otel-collector configs should be removed?
- What rigid structure for ./configs/?
- How to validate both ./deployments/ and ./configs/ comprehensively?

## Technical Context

- **Language**: Go 1.25.5
- **CICD Tool**: internal/cmd/cicd/lint_deployments/
- **Directories**: ./deployments/ (36 config files), ./configs/ (55 config files)
- **Architecture Docs**: ARCHITECTURE.md, ARCHITECTURE-COMPOSE-MULTIDEPLOY.md
- **Current Validation**: Partial (lint_required_contents_deployments.go exists)

## Phases

### Phase 1: Investigation & Analysis (4h) [Status: ☐ TODO]
**Objective**: Understand current state, identify redundancies, establish patterns
- Catalog all .gitkeep files and determine which to remove
- Analyze compose file purposes (deployments/compose/, sm-kms/compose.demo.yml)
- Analyze otel-collector-config.yaml locations and determine canonical source
- Document why deployments/template/config/ is empty
- Understand ./configs/ structure vs ./deployments/ patterns
- Review existing CICD lint implementation
- **Success**: Clear understanding of what to clean, what to keep, what patterns to enforce

### Phase 2: Structural Cleanup (3h) [Status: ☐ TODO]
**Objective**: Remove redundant files, explain architectural decisions
- Delete .gitkeep files in non-empty directories
- Document or delete redundant compose files
- Document or delete redundant otel-collector-config.yaml files
- Create RATIONALE.md explaining architectural decisions
- **Success**: Clean directory structure with documented decisions

### Phase 3: CICD Refactoring - Deployments (6h) [Status: ☐ TODO]
**Objective**: Enhance lint_deployments with comprehensive validation
- Refactor lint_required_contents_deployments.go with complete file lists
- Add suite/product/service directory lists for filtering
- Implement credential validation (no hardcoded passwords/peppers/unseals)
- Enhance ValidateDeploymentStructure control layer
- Add tests for all validation logic
- **Success**: ≥95% coverage, comprehensive deployment validation

### Phase 4: CICD Refactoring - Configs (6h) [Status: ☐ TODO]
**Objective**: Establish rigid ./configs/ validation matching ./deployments/ rigor
- Design ./configs/ rigid structure (suite/product/service patterns)
- Implement comprehensive lint_required_contents_configs.go
- Add credential validation for config files
- Validate shared directories (shared-*, template)
- Add tests for config validation
- **Success**: ≥95% coverage, rigorous config validation

### Phase 5: Config Directory Restructuring (8h) [Status: ☐ TODO]
**Objective**: Apply rigid structure to ./configs/ matching ./deployments/ patterns
- Create suite/product/service hierarchy in ./configs/
- Migrate existing config files to new structure
- Ensure CLI development workflows unaffected
- Update all references to config files
- Test suite/product/service level runs
- **Success**: ./configs/ matches ./deployments/ rigor, all workflows passing

### Phase 6: Documentation Updates (4h) [Status: ☐ TODO]
**Objective**: Update ARCHITECTURE.md and propagate to linked docs
- Document new ./configs/ structure in ARCHITECTURE.md
- Document CICD validation enhancements
- Update ARCHITECTURE-COMPOSE-MULTIDEPLOY.md if needed
- Propagate changes via bidirectional links
- Update instruction files if needed
- **Success**: Complete, accurate documentation

### Phase 7: Quality Gates (4h) [Status: ☐ TODO]
**Objective**: Verify all quality requirements met
- Build main code: `go build ./...`
- Build test code: `go test ./... -run=^$`
- Unit tests: `go test ./...` with ≥95% coverage
- Linting: `golangci-lint run` clean
- Pre-commit checks: All hooks passing
- Integration tests: TestMain patterns passing
- E2E tests: Docker Compose scenarios passing
- Mutation testing: ≥95% (production), ≥98% (infrastructure)
- **Success**: All quality gates green

## Executive Decisions

### Decision 1: deployments/compose/compose.yml

**Options**:
- A: Delete (violates e2e pattern, redundant with suite-level)
- B: Keep as alternative e2e approach
- C: Keep as official E2E compose file (NOT redundant) ✓ **SELECTED**
- D: Move to internal/test/e2e/compose/

**Decision**: Option C selected - Keep as official E2E infrastructure

**Rationale**:
- Actively referenced in internal/shared/magic/magic_docker.go
- Used by internal/test/e2e/ test suite
- Overrides otel-collector ports for host-based E2E testing
- NOT redundant - serves specific E2E testing purpose
- Documented in compose.yml header: "E2E Testing Compose Configuration"

**Impact**: No changes needed, document purpose in ARCHITECTURE.md

**Evidence**: grep shows active usage, magic_docker.go constants reference this file

### Decision 2: deployments/sm-kms/compose.demo.yml

**Options**:
- A: Delete (violates demo pattern, use suite-level demo)
- B: Keep as service-specific demo
- C: Move to deployments/template/ as pattern
- D: Investigate if suite/product level demo patterns exist, then decide

**Decision**: Option D selected - Investigate first

**Rationale**: Need to understand if suite/product level demo patterns exist before deciding if service-level demo files should exist

**Impact**: Phase 1 investigation required

### Decision 3: otel-collector-config.yaml Files

**Options**:
- A: Delete all except shared-telemetry/otel/otel-collector-config.yaml (canonical)
- B: Keep template/, delete cipher-im/
- C: Keep all for customization
- D: Investigate usage, then decide ✓ **SELECTED**

**Decision**: Option D selected - Investigate first  

**Rationale**:
- shared-telemetry/otel/ is canonical source
- template/ might be intentional pattern example
- cipher-im/ likely duplicate
- Need to verify no customizations before deletion

**Impact**: Phase 1 investigation, likely Phase 2 deletion

### Decision 4: ./configs/ Structure

**Options**:
- A: Minimal validation (current state)
- B: Mirror ./deployments/ exactly (suite/product/service)
- C: Hybrid approach (less strict than deployments)
- D: Design custom structure based on CLI usage patterns ✓ **SELECTED**

**Decision**: Option D selected - Design for CLI workflows

**Rationale**:
- ./configs/ is for non-compose local development
- Should support suite/product/service level runs like ./deployments/
- Needs rigorous validation like ./deployments/
- But may need different patterns (profiles/, policies/)

**Impact**: Major restructuring in Phase 5, CICD validation in Phase 4

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Breaking CLI workflows | Medium | High | Test all suite/product/service runs after restructuring |
| Breaking E2E tests | Medium | High | Run E2E tests after every change, maintain compose paths |
| Incomplete CICD validation | Medium | Medium | Comprehensive test coverage ≥95%, mutation testing |
| Config migration errors | Medium | High | Validate all references updated, keep backups |
| Documentation drift | Low | Medium | Bidirectional link validation, comprehensive reviews |

## Quality Gates - MANDATORY

**Per-Action Quality Gates**:
- ✅ All tests pass (`go test ./...`) - 100% passing, zero skips
- ✅ Build clean (`go build ./...`) - zero errors
- ✅ Linting clean (`golangci-lint run`) - zero warnings
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets**:
- ✅ Production code: ≥95% line coverage
- ✅ Infrastructure/utility code: ≥98% line coverage  
- ✅ CICD linting tools: ≥98% (critical infrastructure)

**Mutation Testing Targets**:
- ✅ CICD validation logic: ≥98% (NO EXCEPTIONS)
- ✅ Production code: ≥95%

**Per-Phase Quality Gates**:
- ✅ Unit tests complete before next phase
- ✅ Integration tests pass where applicable
- ✅ E2E tests pass for deployment changes
- ✅ CICD validation passes for structure changes

**Overall Project Quality Gates**:
- ✅ All phases complete with evidence
- ✅ All test categories passing (unit, integration, E2E)
- ✅ Coverage and mutation targets met
- ✅ CI/CD workflows green
- ✅ Documentation updated and validated

## Success Criteria

- [ ] All redundant files removed with documented rationale
- [ ] CICD lint_deployments comprehensively validates ./deployments/
- [ ] CICD lint_deployments comprehensively validates ./configs/
- [ ] ./configs/ has rigid structure matching ./deployments/ principles
- [ ] All quality gates passing
- [ ] Documentation complete and accurate (ARCHITECTURE.md updated)
- [ ] CI/CD workflows green
- [ ] Evidence archived (test output, validation logs)
