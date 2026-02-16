# Implementation Plan - Deployment & Config Structure Refactoring V2

**Status**: Ready for Execution
**Created**: 2026-02-16
**Last Updated**: 2026-02-16
**Purpose**: Refactor deployment/config structure, enhance CICD validation, establish rigid patterns

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO steps skipped, ALL features fully implemented
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability, NOT speed
- ✅ **Accuracy**: Address root cause, not symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation
- ❌ **Premature Completion**: NEVER mark complete without evidence

**ALL issues are blockers - NO exceptions - Fix immediately**

## Overview

This plan refactors deployment/config structure with these objectives:
1. Clean up structural inconsistencies (delete redundant files)
2. Establish rigid validated patterns for ./deployments/ and ./configs/
3. Mirror ./configs/ structure to match ./deployments/ (exact mirror per user decision)
4. Comprehensive CICD validation with schema checks
5. Archive demo files for future brainstorming phase

## Executive Decisions (From Quizme Answers)

### Decision 1: Demo Files → Archive for Future Research
**Selected**: Archive under docs/demo-brainstorm/archive/, create DEMO-BRAINSTORM.md
**Rationale**: Too much other work (deployment/config refactoring). Need deep research on demo best practices before establishing patterns. Defer to separate phase after research complete.
**Impact**: Phase 0.5 - Archive demo files, no implementation yet

### Decision 2: ./configs/ Structure → Exact Mirror of ./deployments/
**Selected**: Option A - configs/{cryptoutil,PRODUCT,PRODUCT-SERVICE}/ matching deployments exactly
**Rationale**: Maximum rigor and consistency. Supports suite/product/service level runs for CLI development.
**Impact**: Major restructuring in Phase 5 (all 55 files)

### Decision 3: Otel-Collector Configs → Single Canonical Source
**Selected**: Keep ONLY shared-telemetry/otel/otel-collector-config.yaml, delete template & cipher-im copies
**Rationale**: shared-telemetry handles ALL 27 possible service instances (9 SUITE + 9 PRODUCT + 9 SERVICE). Single source of truth prevents confusion.
**Impact**: Delete 2 duplicate files, update docs and CICD validation

### Decision 4: Implementation Priority → Execute All Phases Autonomously NOW
**Selected**: Start immediately, execute all phases autonomously
**Rationale**: User mandate - "STOP ASKING ME TO CONFIRM!!!"
**Impact**: No checkpoints, continuous execution until complete

### Decision 5: Config Restructuring Scope → Full Migration
**Selected**: Move ALL 55 files, update ALL references, comprehensive migration
**Rationale**: Maximum consistency, matches exact mirror decision
**Impact**: High-risk but highest quality outcome

## Additional Clarifications

1. **template/config/**: Must have 4 config files (PRODUCT-SERVICE-app-{common,sqlite-1,postgresql-1,postgresql-2}.yml) matching sm-kms/config/ pattern. CICD must validate.

2. **Pre-Commit Compose Validation**: lint-compose exists but missed VS Code validation errors. Enhance to use `docker compose config --quiet` for schema validation.

3. **Simple Analysis During Planning**: .gitkeep files, otel configs, template requirements analyzed in Phase 1 (already complete).

## Technical Context

- **Language**: Go 1.25.5
- **CICD Tool**: internal/cmd/cicd/lint_deployments/
- **Directories**: ./deployments/ (36 config files), ./configs/ (55 config files)
- **Architecture Docs**: ARCHITECTURE.md, ARCHITECTURE-COMPOSE-MULTIDEPLOY.md
- **Phase 1 Analysis**: Complete (evidence in test-output/phase1/)

## Phases

### Phase 0.5: Demo Files Archive (1h) [Status: ☐ TODO]
**Objective**: Archive demo files for future research
- Create docs/demo-brainstorm/ and docs/demo-brainstorm/archive/
- Move compose.demo.yml to archive
- Create DEMO-BRAINSTORM.md stub for future research
- **Success**: Demo files archived, path cleared for main work

### Phase 1: Structural Cleanup (2h) [Status: ☐ TODO]
**Objective**: Remove redundant files based on completed analysis
- Delete 2 .gitkeep files (cipher-im/config/, configs/)
- Delete 2 duplicate otel-collector-config.yaml files (template/, cipher-im/)
- Create template/config/ files (4 config placeholders)
- Document decisions in RATIONALE.md
- **Success**: Clean structure, template complete, decisions documented

### Phase 2: Enhance Docker Compose Validation (3h) [Status: ☐ TODO]
**Objective**: Prevent VS Code validation errors from reaching commits
- Enhance lint-compose to use `docker compose config --quiet`
- Add schema validation for all compose files
- Test against all 24 compose files
- Add tests with ≥98% coverage
- **Success**: Comprehensive compose validation catches schema errors

### Phase 3: CICD Refactoring - Deployments (6h) [Status: ☐ TODO]
**Objective**: Comprehensive deployment validation
- Complete file lists for ALL expected files
- Add suite/product/service directory filtering
- Validate template/ contents (compose files + config files)
- Validate shared-* directories
- Credential validation (no hardcoded passwords/peppers/unseals)
- Tests with ≥98% coverage
- **Success**: Rigorous deployment structure validation

### Phase 4: CICD Refactoring - Configs (6h) [Status: ☐ TODO]
**Objective**: Establish rigid ./configs/ validation matching ./deployments/
- Design exact mirror structure (cryptoutil/, PRODUCT/, PRODUCT-SERVICE/)
- Implement comprehensive file lists
- Add credential validation
- Validate shared directories
- Tests with ≥98% coverage
- **Success**: ./configs/ validation matches ./deployments/ rigor

### Phase 5: Config Directory Restructuring (8h) [Status: ☐ TODO]
**Objective**: Migrate all 55 config files to rigid structure
- Create suite/product/service hierarchy
- Migrate ALL 55 files using `git mv`
- Update ALL references in code/docs
- Test suite/product/service level CLI runs
- CICD validation passes
- **Success**: ./configs/ mirrors ./deployments/, all workflows passing

### Phase 6: Documentation Updates (4h) [Status: ☐ TODO]
**Objective**: Update ARCHITECTURE.md and propagate changes
- Document ./configs/ rigid structure
- Document CICD validation enhancements
- Document otel-collector canonical source
- Update ARCHITECTURE-COMPOSE-MULTIDEPLOY.md
- Propagate via bidirectional links
- **Success**: Complete accurate documentation

### Phase 7: Quality Gates (4h) [Status: ☐ TODO]
**Objective**: ALL quality requirements met
- Build: `go build ./...`
- Tests: `go test ./...` ≥95% coverage
- Linting: `golangci-lint run` clean
- Pre-commit: All hooks passing
- Integration/E2E: All passing
- Mutations: ≥95% (≥98% CICD)
- **Success**: All quality gates green

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Breaking CLI workflows | Medium | High | Test every suite/product/service run after migration |
| Breaking E2E tests | Low | High | E2E tests after every structural change |
| Incomplete CICD validation | Low | Medium | Comprehensive test coverage ≥98%, mutation testing |
| Config migration errors | Medium | High | Use `git mv`, validate all references, keep evidence |
| Documentation drift | Low | Medium | Bidirectional link validation, propagation checks |

## Success Criteria

- [ ] All redundant files removed (2 .gitkeep, 2 otel configs)
- [ ] Demo files archived for future research
- [ ] template/config/ has 4 required config files
- [ ] Docker Compose validation enhanced (schema checks)
- [ ] CICD validates ./deployments/ comprehensively
- [ ] CICD validates ./configs/ comprehensively
- [ ] ./configs/ exact mirror of ./deployments/ (55 files migrated)
- [ ] All quality gates passing (build, tests, coverage, mutations, linting)
- [ ] Documentation complete and propagated
- [ ] CI/CD workflows green
