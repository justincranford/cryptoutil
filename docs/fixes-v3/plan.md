# Implementation Plan - Configs/Deployments/CICD Rigor & Consistency v3

**Status**: Planning
**Created**: 2026-02-17
**Last Updated**: 2026-02-17
**Purpose**: Achieve absolute rigor and consistency for configs/, deployments/, and CICD linting to fully comply with ARCHITECTURE.md standards and eliminate all inconsistencies.

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks complete without objective evidence

**ALL issues are blockers - NO exceptions:**

- ✅ **Fix issues immediately** - When unknowns discovered, blockers identified, unit/integration/E2E/mutations/fuzz/bench/race/SAST/DAST/load/any tests fail, or quality gates are not met, STOP and address
- ✅ **Treat as BLOCKING**: ALL issues block progress to next task
- ✅ **Document root causes** - Root cause analysis is part of planning AND implementation, not optional; planning blockers must be resolved during planning, implementation blockers MUST be resolved during implementation
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task complete with known issues
- ✅ **NEVER de-prioritize quality** - Evidence-based verification is ALWAYS highest priority

**Rationale**: Maintaining maximum quality prevents cascading failures and rework.

## Overview

This plan addresses comprehensive rigor and consistency improvements across three critical areas:
1. **configs/** directory restructuring for naming pattern consistency
2. **deployments/** validation and gap closure
3. **CICD linting** enhancement for comprehensive validation

**Current State** (from Phase 0 research):
- deployments/: ✅ 20/20 deployments PASS structural validation
- configs/: ❌ Inconsistent naming, missing subdirectories, wrong patterns
- CICD linting: ⚠️ Partial implementation, missing validations

**Target State**:
- configs/: 100% consistent naming (PRODUCT-SERVICE-app-{common,sqlite-1,postgresql-1,postgresql-2}.yml)
- deployments/: Enhanced validation (PRODUCT/SUITE configs, README.md files)
- CICD linting: Comprehensive validation (8 new validation types)

## Background

**Prior Work** (fixes-v2):
- Phase 4: Implemented ValidateComposeFile (7 types) and ValidateConfigFile (5 types)
- Phase 5: Restructured configs/ directory, created orphaned/ archive
- Phase 6: Documented CONFIG-SCHEMA.md, updated ARCHITECTURE.md

**Lessons Learned from fixes-v2**:
1. deployments/ validation is strong (rigid delegation pattern enforced)
2. configs/ still has old naming patterns (config.yml, config-pg-N.yml)
3. CICD linting validates structure but NOT content rigorously
4. Missing PRODUCT/SUITE-level configs

**Gaps Carried Forward to fixes-v3**:
1. configs/ uses OLD naming (config.yml) instead of NEW naming (PRODUCT-SERVICE-app-common.yml)
2. configs/ca/ should be configs/pki-ca/ to match deployments/pki-ca/
3. configs/identity/ has mixed authz/idp/rs files (should be separate subdirs)
4. configs/sm/ missing sm-kms/ subdirectory
5. No PRODUCT-level configs (cipher/, jose/, identity/, sm/, pki/)
6. No SUITE-level configs (cryptoutil/)
7. CICD linting missing 8 validation types (naming, kebab-case, schema, ports, telemetry, admin, consistency)

## Executive Summary

**Critical Context**:
- configs/ directory has 55 files with OLD naming patterns that don't match deployments/*/config/
- deployments/ directory has 40 config files (in */config/) with CORRECT naming
- CICD linting validates deployment structure (✅) but NOT config content rigorously (❌)
- ARCHITECTURE.md defines complete requirements, but implementation is incomplete

**Assumptions & Risks**:
- Assumption: Renaming configs/ files won't break existing services (files are read-only reference, not actively used)
- Assumption: PRODUCT/SUITE configs are optional for now (services run independently)
- Risk: Breaking existing configs during rename (Mitigation: Test immediately after rename)
- Risk: configs/ and deployments/*/config/ diverge again (Mitigation: Add consistency validation to pre-commit)

## Technical Context

- **Language**: Go 1.25.5
- **Framework**: internal/cmd/cicd/lint_deployments/ (CICD tooling)
- **Database**: Not applicable (file system operations)
- **Dependencies**: docker, docker compose, golangci-lint, pre-commit, gopkg.in/yaml.v3
- **Related Files**:
  - `internal/cmd/cicd/lint_deployments/*.go` (enhance validation)
  - `configs/**/*.yml` (restructure and rename)
  - `deployments/**/*.yml` (add PRODUCT/SUITE configs)
  - `docs/ARCHITECTURE.md`, `docs/CONFIG-SCHEMA.md` (documentation)
  - `.pre-commit-config.yaml` (add new validations)

## Phases

### Phase 1: configs/ Directory Restructuring (12h) [Status: ☐ TODO]
**Objective**: Achieve 100% naming consistency in configs/ to match deployments/*/config/ patterns

**Scope**:
1. Rename ca/ → pki-ca/ (directory itself)
2. Restructure identity/ → identity-{authz,idp,rp,rs,spa}/ (5 subdirectories)
3. Create configs/sm-kms/ (currently missing)
4. Rename ALL config files to PRODUCT-SERVICE-app-{common,sqlite-1,postgresql-1,postgresql-2}.yml
5. Handle environment files (development.yml, production.yml, test.yml)

**Success Criteria**:
- 0 files with old naming (config.yml, config-pg-N.yml, config-sqlite.yml)
- 100% files match PRODUCT-SERVICE-app-VARIANT.yml pattern
- Directory structure matches deployments/ (pki-ca/, identity-*, sm-kms/)
- All tests pass, no broken imports

### Phase 2: PRODUCT/SUITE Config Creation (6h) [Status: ☐ TODO]
**Objective**: Add missing PRODUCT and SUITE-level configurations

**Scope**:
1. Create PRODUCT-level configs (cipher/, jose/, identity/, sm/, pki/)
2. Create SUITE-level configs (cryptoutil/)
3. Add README.md to each PRODUCT/SUITE explaining delegation pattern
4. Document config sharing/override patterns

**Success Criteria**:
- PRODUCT-level configs exist for all 5 products
- SUITE-level config exists for cryptoutil
- README.md files explain delegation clearly
- Configs validated by cicd lint-deployments

### Phase 3: CICD Linting Enhancement - Config Validation (18h) [Status: ☐ TODO]
**Objective**: Implement 8 missing validation types for config files

**Scope**:
1. Config file naming pattern enforcement
2. Kebab-case key validation (flat YAML, no camelCase/snake_case)
3. Schema completeness validation (required fields per CONFIG-SCHEMA.md)
4. Port offset validation (SERVICE +0, PRODUCT +10000, SUITE +20000)
5. Telemetry configuration validation (OTLP protocols: grpc/http, endpoints)
6. Admin policy enforcement (private bind MUST be 127.0.0.1)
7. deployments/*/config/ vs configs/ consistency check
8. Secret reference validation (file:// paths exist, correct suffixes)

**Success Criteria**:
- 8 new validation functions implemented with tests (≥95% coverage)
- All validations integrated into `ValidateConfigFile`
- Pre-commit hook runs new validations
- All existing configs PASS validation (or issues documented/fixed)

### Phase 4: CICD Linting Enhancement - Deployment Validation (12h) [Status: ☐ TODO]
**Objective**: Enhance deployment structure validation for PRODUCT/SUITE levels

**Scope**:
1. Validate PRODUCT-level compose delegation (includes SERVICE composes)
2. Validate SUITE-level compose delegation (includes PRODUCT composes)
3. Validate README.md existence in PRODUCT/SUITE deployments
4. Validate port offset consistency in compose files
5. Validate secret suffix consistency (validate .never files)

**Success Criteria**:
- PRODUCT/SUITE validation functions implemented
- All 20 deployments continue to PASS
- New validations catch misconfigurations
- Tests cover edge cases (≥95% coverage)

### Phase 5: Pre-Commit Integration (4h) [Status: ☐ TODO]
**Objective**: Add all new validations to pre-commit hooks for enforcement

**Scope**:
1. Add lint-deployments validate-mirror to pre-commit
2. Add lint-deployments validate-config to pre-commit (scan all configs/)
3. Add lint-deployments validate-compose to pre-commit (scan all compose.yml)
4. Configure hooks to run on relevant file changes only

**Success Criteria**:
- Pre-commit hooks prevent commits with validation errors
- Hooks run performantly (<30s for typical changes)
- Documentation updated (DEV-SETUP.md, CONTRIBUTING.md)

### Phase 6: Documentation & Testing (6h) [Status: ☐ TODO]
**Objective**: Complete documentation and comprehensive testing

**Scope**:
1. Update ARCHITECTURE.md with all new validation types
2. Update CONFIG-SCHEMA.md with examples for PRODUCT/SUITE configs
3. Create migration guide for old → new naming
4. Add E2E test: validate entire project structure
5. Create README.md for configs/ explaining structure

**Success Criteria**:
- ARCHITECTURE.md Section 12.4 fully documents all validations
- CONFIG-SCHEMA.md has PRODUCT/SUITE examples
- Migration guide tested with 1 service
- E2E test runs in CI/CD

## Executive Decisions

### Decision 1: configs/ Naming Migration Strategy

**Options**:
- A: In-place rename (git mv), test immediately
- B: Create new structure alongside old, migrate gradually
- C: Deprecate configs/, use only deployments/*/config/
- D: Script-based batch rename with rollback ✓ **SELECTED**

**Decision**: Option D selected - Script-based batch rename with automated rollback capability

**Rationale**:
- Automated script ensures consistency (no manual typos)
- Rollback script provides safety net
- Single PR minimizes coordination overhead
- configs/ is reference directory (low risk of breaking services)

**Alternatives Rejected**:
- Option A: Manual git mv is error-prone at scale (55 files)
- Option B: Gradual migration adds complexity, prolongs inconsistency
- Option C: configs/ serves as reference/examples, should not be deprecated

**Impact**:
- Technical: Requires rename script + rollback script
- Schedule: Adds 2h to Phase 1 for script development
- Risk: Low (configs/ is read-only reference)

**Evidence**: Prior rename in fixes-v2 (configs/template/, configs/orphaned/) was successful

### Decision 2: PRODUCT/SUITE Config Location

**Options**:
- A: configs/PRODUCT/ (flat, matches directory structure)
- B: deployments/PRODUCT/config/ (collocated with compose)
- C: configs/PRODUCT/ AND deployments/PRODUCT/config/ (duplicate)
- D: configs/PRODUCT/ as templates, deployments/PRODUCT/config/ for deployment ✓ **SELECTED**

**Decision**: Option D selected - configs/ for templates, deployments/ for deployment instances

**Rationale**:
- configs/ serves as reference/examples (single source of truth for schemas)
- deployments/ contains runtime configs (environment-specific overrides)
- Aligns with existing pattern (configs/PRODUCT-SERVICE/, deployments/PRODUCT-SERVICE/config/)
- Enables CICD validation: deployments/*/config/ must reference configs/ templates

**Alternatives Rejected**:
- Option A: Missing runtime configs in deployments/ (incomplete structure)
- Option B: Missing reference templates in configs/ (no schema examples)
- Option C: Duplication maintenance burden, conflicting sources of truth

**Impact**:
- Technical: Both directories exist, different purposes
- Schedule: Adds README.md requirement to explain relationship
- Risk: Low (clear separation of concerns)

### Decision 3: CICD Validation Enforcement Mode

**Options**:
- A: Warnings only (non-blocking)
- B: Errors for all violations (strict blocking)
- C: Hybrid: Errors for critical, warnings for minor ✓ **SELECTED**
- D: Configurable per-validation (max flexibility)

**Decision**: Option C selected - Errors for critical violations, warnings for minor issues

**Rationale**:
- Critical violations (wrong naming, missing required fields) MUST block CI/CD
- Minor issues (missing README.md, suggestion improvements) can be warnings
- Balances rigor (prevents bad commits) with pragmatism (doesn't block valid work)
- Aligns with existing linting patterns (golangci-lint uses errors/warnings)

**Alternatives Rejected**:
- Option A: Too lenient, allows inconsistencies to accumulate
- Option B: Too strict, may block legitimate work-in-progress
- Option D: Over-engineered, adds configuration complexity

**Impact**:
- Technical: Validation functions return {Errors: []string, Warnings: []string}
- Schedule: No change (already implemented in fixes-v2)
- Risk: Low (follows established patterns)

**Critical Validations (ERRORS)**:
1. Config file naming pattern
2. Kebab-case keys
3. Required fields missing
4. Invalid bind addresses/ports
5. Hardcoded credentials
6. Invalid secret references
7. Port offset violations
8. Admin policy violations

**Minor Validations (WARNINGS)**:
1. Missing README.md
2. Suboptimal OTLP protocol
3. Missing optional fields
4. Deprecated patterns (not yet removed)

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Breaking existing configs during rename | Medium | High | Script-based rename + rollback capability, immediate testing |
| configs/ and deployments/*/config/ diverge again | Medium | Medium | Add consistency validation to pre-commit hooks |
| CICD linting too strict blocks valid work | Low | Medium | Hybrid enforcement (errors for critical, warnings for minor) |
| Missing validation types cause future issues | Low | Low | Comprehensive Phase 3 coverage, extensible validation framework |
| Pre-commit hooks too slow (>30s) | Low | Low | Optimize file scanning, only validate changed files |
| PRODUCT/SUITE configs cause confusion | Medium | Low | Clear README.md files, separate configs/ (templates) from deployments/ (runtime) |

## Quality Gates - MANDATORY

**Per-Task Quality Gates**:
- ✅ All tests pass (`go test ./...`) - 100% passing, zero skips
- ✅ Build clean (`go build ./...`) - zero errors
- ✅ Linting clean (`golangci-lint run`) - zero warnings
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets** (from copilot instructions):
- ✅ Production code: ≥95% line coverage
- ✅ Infrastructure/utility code (CICD): ≥98% line coverage
- ✅ main() functions: 0% acceptable if internalMain() ≥95%
- ✅ Generated code: Excluded (OpenAPI stubs, GORM models)

**Mutation Testing Targets** (from copilot instructions):
- ✅ Production code: ≥95% minimum, ≥98% ideal
- ✅ Infrastructure/utility code (CICD): ≥98% (NO EXCEPTIONS)

**Per-Phase Quality Gates**:
- ✅ Unit tests complete before moving to next phase
- ✅ Renamed files validated (no broken imports)
- ✅ CICD validations pass on all configs (before/after)
- ✅ Pre-commit hooks tested locally

**Overall Project Quality Gates**:
- ✅ All phases complete with evidence
- ✅ All test categories passing (unit, integration)
- ✅ Coverage and mutation targets met
- ✅ CI/CD workflows green
- ✅ Documentation updated (ARCHITECTURE.md, CONFIG-SCHEMA.md)

## Success Criteria

- [ ] Phase 1: 0 files with old naming in configs/, 100% match new pattern
- [ ] Phase 2: PRODUCT/SUITE configs exist with README.md files
- [ ] Phase 3: 8 new config validation types implemented (≥95% coverage, ≥95% mutation)
- [ ] Phase 4: PRODUCT/SUITE deployment validations implemented (≥95% coverage, ≥95% mutation)
- [ ] Phase 5: Pre-commit hooks enforce all validations
- [ ] Phase 6: Documentation complete, E2E test passing
- [ ] All quality gates passing
- [ ] CI/CD workflows green
- [ ] Evidence archived (test-output/phase0-research/, test-output/phase1/, etc.)
