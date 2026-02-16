# Implementation Plan - Deployment/Config Refactoring v2

**Status**: Planning Complete, Ready for Execution
**Created**: 2026-02-16
**Last Updated**: 2026-02-16
**Purpose**: Comprehensive cleanup and restructuring of ./deployments/ and ./configs/ directories with rigorous validation

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
- ✅ **Fix issues immediately** - When tests fail or quality gates not met, STOP and address
- ✅ **Treat as BLOCKING** - ALL issues block progress to next phase or task
- ✅ **Document root causes** - Root cause analysis is MANDATORY
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical"
- ✅ **NEVER de-prioritize quality** - Evidence-based verification is ALWAYS highest priority

## Overview

Comprehensive refactoring of deployment and configuration management:
1. Archive demo files (preserve for future reference)
2. Structural cleanup (delete redundant files)
3. Enhanced Docker Compose validation (catch schema errors)
4. CICD lint_deployments refactoring (rigorous validation)
5. Config directory restructuring (exact mirror of deployments)
6. Documentation updates
7. Comprehensive testing and quality gates

## Background

User discovered multiple issues:
- `.gitkeep` files in directories with content
- Duplicate `otel-collector-config.yaml` files
- VS Code compose validation errors (manually fixed, must prevent)
- `deployments/template/config/` incorrectly empty (needs 4 files)
- `internal/cmd/cicd/lint_deployments/` lacking comprehensive validation
- `./configs/` directory structure misaligned with `./deployments/`

**Previous Work**:
- Phase 1 investigation completed (test-output/phase1/ evidence)
- Quizme-v1 answers: Archive demos, exact mirror, delete duplicates, full restructuring, autonomous execution
- Quizme-v2 answers: JSON listings, comprehensive validations (compose+config), orphan handling

## Executive Summary

**Critical Context**:
- User demands **RIGOROUS** comprehensive validation (quizme-v2 Q3:C, Q4:C)
- JSON listing files with metadata enable rich validation (quizme-v2 Q1:C)
- Deployments-driven mirror: all deployments/ MUST have configs/, but configs/ can have extras (quizme-v2 Q2:C)
- Orphaned configs preserved in configs/orphaned/ for review (quizme-v2 Q5:C)
- Autonomous execution: NO checkpoints, proceed through all phases (quizme-v1 Q4:A)

**Assumptions & Risks**:
- Assumption: Current configs/ has orphans (no deployments/ counterpart)
- Assumption: Config file structure is consistent enough for schema definition
- Risk: Comprehensive validation may discover many issues requiring fixes
- Risk: Schema definition iteration may extend Phase 4 timeline
- Mitigation: Extensive evidence collection, incremental validation testing

## Technical Context

- **Language**: Go 1.25.5
- **Framework**: internal/cmd/cicd/ CICD tooling
- **Database**: Not applicable (file system operations)
- **Dependencies**: docker, docker compose, golangci-lint, pre-commit
- **Related Files**:
  - `internal/cmd/cicd/lint_deployments/*.go` (refactor target)
  - `.pre-commit-config.yaml` (enhance validation)
  - `deployments/**` (36 config files, 24 compose files, secrets)
  - `configs/**` (55 config files to restructure)
  - `deployments/template/config/` (needs 4 template files)

**Configuration Schema Requirements** (NEW - quizme-v2 Q4:C):
- Config files: `PRODUCT-SERVICE-app-{common,sqlite-1,postgresql-1,postgresql-2}.yml`
- Schema must define: server settings, database settings, telemetry, security
- Validation rules: bind addresses, ports, database URLs, secret references
- Documentation: Add to ARCHITECTURE.md Section 12.5 or create CONFIG-SCHEMA.md

## Phases

### Phase 0.5: Demo Files Archiving (2h) [Status: ☐ TODO]
**Objective**: Archive demo files under docs/demo-brainstorm/ (preserving for future)
- Create directory structure
- Move sm-kms/compose.demo.yml
- Create DEMO-BRAINSTORM.md stub
- **Success**: No demo files in deployments/, preserved in docs/

### Phase 1: Structural Cleanup (3h) [Status: ☐ TODO]
**Objective**: Delete redundant files, create missing template configs
- Task 1.1: Delete 2 .gitkeep files
- Task 1.2: Delete 2 duplicate otel configs (keep shared-telemetry)
- Task 1.3: Create 4 template config files
- **Success**: No .gitkeep with content, single otel config, template complete

### Phase 2: Compose Validation Enhancement (3h) [Status: ☐ TODO]
**Objective**: Add docker compose config validation to pre-commit
- Update .pre-commit-config.yaml with schema validation
- Test with all compose files
- Verify catches VS Code-reported errors
- **Success**: Pre-commit prevents invalid compose commits

### Phase 3: CICD Foundation (10h) [Status: ✅ COMPLETE - 3.5h actual]
**Objective**: Create listing files and structural mirror validation
- Task 3.1: Generate JSON listing files with metadata (4h)
- Task 3.2: Implement ValidateStructuralMirror (4h)
- Task 3.3: Write comprehensive tests (2h)
- **Success**: JSON listings exist, mirror validation correct, ≥95% coverage

### Phase 4: CICD Comprehensive Refactoring (28h) [Status: ☐ TODO]
**Objective**: Rigorous validation for compose and config files
- Task 4.0: Define config file schema (3h) **[NEW]**
- Task 4.1: Implement comprehensive ValidateComposeFiles (10h)
- Task 4.2: Write tests for compose validation (3h)
- Task 4.3: Implement comprehensive ValidateConfigFiles (8h)
- Task 4.4: Write tests for config validation (4h)
- **Success**: All validation types implemented, tested, ≥95% coverage

**Compose Validation Types** (7 total - quizme-v2 Q3:C):
1. Schema validation (docker compose config --quiet)
2. Port conflict detection
3. Health check presence
4. Service dependency chains
5. Secret reference validation
6. No hardcoded credentials
7. Bind mount security

**Config Validation Types** (5 total - quizme-v2 Q4:C):
1. YAML syntax
2. Format validation (bind addresses, ports, URLs)
3. Cross-reference with compose services
4. Bind address policy enforcement
5. Secret reference validation

### Phase 5: Config Directory Restructuring (6h) [Status: ☐ TODO]
**Objective**: Mirror configs/ to match deployments/ structure
- Task 5.1: Audit current configs/ structure (1h)
- Task 5.2: Identify orphans (configs without deployments) (1h)
- Task 5.3: Create configs/orphaned/, move orphans, restructure valid (3h)
- Task 5.4: Validate mirror correctness (1h)
- **Success**: Exact mirror for valid configs, orphans preserved, validation passes

### Phase 6: Documentation & Integration (3h) [Status: ☐ TODO]
**Objective**: Update docs and integrate into workflows
- Update ARCHITECTURE.md with schema
- Update instructions files
- Integrate into CI/CD workflows
- **Success**: Docs accurate, CI/CD enforces validations

### Phase 7: Comprehensive Testing & Quality Gates (5h) [Status: ☐ TODO]
**Objective**: End-to-end verification and mutation testing
- E2E tests for all CICD commands
- Mutation testing (≥95% minimum)
- Pre-commit hook verification
- **Success**: All quality gates passing, evidence documented

## Executive Decisions

### Decision 1: File Listing Format (quizme-v2 Q1)

**Options**:
- A: Simple newline-separated list
- B: Hierarchical with indentation
- C: JSON with metadata ✓ **SELECTED**
- D: Simple list with comment headers

**Decision**: Option C selected - JSON with type/status metadata

**Rationale**:
- Rich metadata enables type-specific validation (compose vs config vs secret)
- Status tracking (required vs optional) supports validation rules
- Future extensibility (can add: owner, validation-rules, last-modified)
- Parsing complexity acceptable for quality benefits

**Alternatives Rejected**:
- Option A: Too simple, loses type information
- Option B: Human-readable but hard to parse programmatically
- Option D: Comments fragile, not machine-parseable

**Impact**:
- Task 3.1: Generate JSON (not text) - adds 1h LOE
- ValidateStructuralMirror: Must parse JSON - adds complexity
- Future: Can add validation-specific metadata fields

**Evidence**: User selected "C" in quizme-v2.md

### Decision 2: Mirror Strictness (quizme-v2 Q2)

**Options**:
- A: Exact 1:1 PRODUCT-SERVICE only
- B: Exact 1:1 for all with placeholders
- C: Deployments-driven strict, configs can have extras ✓ **SELECTED**
- D: Bidirectional loose with exceptions

**Decision**: Option C selected - One-way strict validation

**Rationale**: User stated "the validations must be strict within each of their ./configs/ and ./deployments/ directories, but the presence of extra config files that do not have a corresponding deployment is allowed. This allows for flexibility in handling infrastructure and template cases without blocking the migration."

**Interpretation**:
- MANDATORY: Every deployments/ directory MUST have configs/ counterpart
- ALLOWED: configs/ CAN have extras (orphaned files) - handled in Decision 5
- Infrastructure: May have minimal configs/ (README or excluded from validation)
- Template: Likely excluded from validation (special case)

**Alternatives Rejected**:
- Option A: Too restrictive, blocks infrastructure handling
- Option B: Creates unnecessary placeholder directories
- Option D: Too loose, doesn't enforce deployments → configs mirror

**Impact**:
- ValidateStructuralMirror: One-way check (deployments → configs)
- Orphan handling: Required (see Decision 5)
- Phase 5: Migration doesn't fail on orphans

**Evidence**: User selected "C" with detailed explanation in quizme-v2.md

### Decision 3: Compose Validation Scope (quizme-v2 Q3)

**Options**:
- A: Minimal (schema only) - 2h
- B: Moderate (schema + runtime) - 4-5h
- C: Comprehensive (schema + runtime + security) ✓ **SELECTED** - 8-10h
- D: Staged (A now, B later, C future)

**Decision**: Option C selected - Full comprehensive validation

**Rationale**: User stated "C; rigourous!!!!" indicating maximum quality requirement

**Validation Types** (7 total):
1. Schema: `docker compose config --quiet`
2. Port conflicts: Detect overlapping host port mappings
3. Health checks: ALL services MUST have health checks
4. Dependencies: Validate depends_on chains correct
5. Secrets: All secrets defined in compose secrets section
6. Credentials: NO hardcoded passwords in environment
7. Security: NO /run/docker.sock bind mounts

**Alternatives Rejected**:
- Option A: Too minimal, misses real issues
- Option B: Moderate but doesn't catch security issues
- Option D: Staged approach rejected (user wants rigorous NOW)

**Impact**:
- Task 4.1 LOE: 3h → 10h (7 validation types)
- Task 4.2 LOE: 2h → 3h (test coverage for all types)
- Phase 4 total: +7h from original estimate
- Pre-commit performance: May be slower (~30-60s for all validations)

**Evidence**: User selected "C; rigourous!!!!" in quizme-v2.md

### Decision 4: Config Validation Scope (quizme-v2 Q4)

**Options**:
- A: Minimal (YAML syntax only) - 1-2h
- B: Moderate (syntax + format) - 3-4h
- C: Comprehensive (syntax + format + cross-reference) ✓ **SELECTED** - 6-8h
- D: Staged (A now, B later, C future)

**Decision**: Option C selected - Full comprehensive validation

**Rationale**: User stated "C; rigourous!!!!" matching compose validation requirement

**Validation Types** (5 total):
1. YAML syntax: Parse and validate well-formed
2. Format: Bind addresses (valid IPv4/IPv6), port ranges (1-65535), database URL structure
3. Cross-reference: Config service names match compose.yml services
4. Policy: Admin bind MUST be 127.0.0.1, public SHOULD be 0.0.0.0 (containers)
5. Secrets: Database passwords reference secrets (not inline)

**PREREQUISITE WORK**:
- NEW Task 4.0: Define config file schema (2-3h)
- Schema must document: server settings, database, telemetry, security
- Add to ARCHITECTURE.md Section 12.5 or create CONFIG-SCHEMA.md

**Alternatives Rejected**:
- Option A: Too minimal, allows nonsense configs
- Option B: Moderate but doesn't validate correctness
- Option D: Staged approach rejected (user wants rigorous NOW)

**Impact**:
- NEW Task 4.0: Schema definition (3h)
- Task 4.3 LOE: 3h → 8h (5 validation types + schema)
- Task 4.4 LOE: 2h → 4h (test all validation types)
- Phase 4 total: +11h from original estimate
- Risk: Schema definition may require iteration

**Evidence**: User selected "C; rigourous!!!!" in quizme-v2.md

### Decision 5: Orphaned Config Handling (quizme-v2 Q5)

**Options**:
- A: Validation error (refuse migration until manual cleanup)
- B: Create placeholder deployments automatically
- C: Orphaned directory (move to configs/orphaned/) ✓ **SELECTED**
- D: Best-effort with logging (skip orphans)

**Decision**: Option C selected - Safe preservation in orphaned directory

**Rationale**: Non-destructive, preserves data, allows manual review after migration

**Implementation**:
1. Pre-migration: Identify configs/ files without deployments/ counterpart
2. Create configs/orphaned/ directory
3. Move all orphans during Phase 5 Task 5.3
4. Create configs/orphaned/README.md explaining what to review
5. Continue migration for valid configs
6. Log to test-output/phase5/orphaned-configs.txt

**Alternatives Rejected**:
- Option A: Too restrictive, blocks autonomous execution
- Option B: Auto-creation risky (may create wrong structure)
- Option D: Leaving orphans in place causes confusion

**Impact**:
- Phase 5: Add orphan handling logic (+ 1-2h)
- Manual review: Required AFTER migration (not blocking)
- Rollback: Simple (move files back from orphaned/)

**Evidence**: User selected "C" in quizme-v2.md

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Comprehensive validation discovers many issues | High | High | Incremental fixes, extensive testing, evidence collection |
| Schema definition requires multiple iterations | Medium | Medium | Reference existing configs, user review after draft |
| Pre-commit hooks too slow (>60s) | Medium | Medium | Optimize validation, consider parallel execution |
| Orphaned config count higher than expected | Medium | Low | Orphaned directory handles any count, manual review after |
| JSON parsing complexity causes bugs | Low | Medium | Comprehensive tests, schema validation for JSON structure |

## Quality Gates - MANDATORY

**Per-Action Quality Gates**:
- ✅ All tests pass (100%, zero skips)
- ✅ Build clean (go build ./...)
- ✅ Linting clean (golangci-lint run)
- ✅ No new TODOs without tracking

**Coverage Targets**:
- ✅ CICD code: ≥98% line coverage (infrastructure/utility category)
- ✅ main() functions: 0% acceptable if internalMain() ≥98%
- ✅ Test files themselves: Exempt from coverage

**Mutation Testing Targets**:
- ✅ CICD code: ≥98% (NO EXCEPTIONS - infrastructure/utility)
- ✅ Validation logic: ≥95% minimum

**Per-Phase Quality Gates**:
- ✅ Unit + integration tests complete before next phase
- ✅ E2E tests pass for CICD commands
- ✅ Pre-commit hooks verification
- ✅ Evidence documented (test output, logs)

**Overall Project Quality Gates**:
- ✅ All phases complete with evidence
- ✅ All test categories passing
- ✅ Coverage and mutation targets met
- ✅ CI/CD workflows green
- ✅ Documentation updated

## Success Criteria

- [ ] All 7 phases complete
- [ ] All quality gates passing (≥98% coverage, ≥98% mutations for CICD)
- [ ] Comprehensive validation prevents compose/config errors
- [ ] Mirror validation ensures deployments ↔ configs correctness
- [ ] Orphaned configs preserved for review
- [ ] Documentation updated (ARCHITECTURE.md + CONFIG-SCHEMA.md)
- [ ] CI/CD workflows green
- [ ] Evidence archived (test-output/)

## Timeline Summary

**Updated Estimate**: 40-45 hours total (increased from 30h due to comprehensive validation scope)

| Phase | Original | Updated | Reason |
|-------|----------|---------|--------|
| 0.5 | 2h | 2h | No change |
| 1 | 3h | 3h | No change |
| 2 | 3h | 3h | No change |
| 3 | 8h | 10h | JSON complexity (+2h) |
| 4 | 12h | 28h | Schema + comprehensive validations (+16h) |
| 5 | 4h | 6h | Orphan handling (+2h) |
| 6 | 3h | 3h | No change |
| 7 | 5h | 5h | No change |
| **Total** | **30h** | **40-45h** | **Comprehensive scope (+10-15h)** |

**Justification**: User demands rigorous comprehensive validation (quizme-v2 "rigourous!!!!"). Quality over speed.
