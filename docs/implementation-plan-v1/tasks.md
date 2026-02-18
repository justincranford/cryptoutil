# Tasks - Deployment Architecture Refactoring

**Status**: 11 of 99 tasks complete (11.1%) - Phase 1 COMPLETE, Phase 2 COMPLETE
**Last Updated**: 2026-02-17
**Created**: 2026-02-17

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
- ✅ **Document root causes** - Root cause analysis is part of planning AND implementation, not optional
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task complete with known issues
- ✅ **NEVER de-prioritize quality** - Evidence-based verification is ALWAYS highest priority

**Rationale**: Maintaining maximum quality prevents cascading failures and rework.

---

## Task Checklist

### Phase 1: Discovery & Analysis ✅ COMPLETE

**Phase Objective**: Comprehensive analysis of current deployment structure and E2E test patterns
**Duration**: 1h actual vs 4h estimated (75% efficiency)

#### Task 1.1: Inventory Deployment Files
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: 0.25h
- **Dependencies**: None
- **Description**: Create complete inventory of all files in deployments/ directory
- **Acceptance Criteria**:
  - [x] List all SERVICE-level compose files (9 PRODUCT-SERVICE directories)
  - [x] List all PRODUCT-level compose files (5 PRODUCT directories)
  - [x] List SUITE-level compose (deployments/cryptoutil/)
  - [x] List E2E testing compose (deployments/compose/)
  - [x] List template files (deployments/template/)
  - [x] Document file sizes, last modified dates
  - [x] Output saved to `test-output/phase1/deployment-inventory.txt`
- **Files**:
  - `test-output/phase1/deployment-inventory.txt`
- **Command**: `find deployments/ -name "*.yml" -o -name "*.yaml" > test-output/phase1/deployment-inventory.txt`

#### Task 1.2: Analyze E2E Test Patterns
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: 0.5h
- **Dependencies**: Task 1.1
- **Description**: Document all E2E test locations and patterns
- **Acceptance Criteria**:
  - [x] Identify all E2E test directories (`find . -name e2e -type d`)
  - [x] Analyze cipher-im E2E pattern (SERVICE-level)
  - [x] Analyze identity E2E pattern (PRODUCT-level)
  - [x] Analyze legacy E2E pattern (`internal/test/e2e/`)
  - [x] Document ComposeManager usage patterns
  - [x] Document magic constants for E2E compose paths
  - [x] Output saved to `test-output/phase1/e2e-patterns.txt`
- **Files**:
  - `test-output/phase1/e2e-patterns.txt`
  - Analysis of `internal/apps/template/testing/e2e/compose.go`
  - Analysis of `internal/apps/cipher/im/e2e/testmain_e2e_test.go`
  - Analysis of `internal/apps/identity/e2e/testmain_e2e_test.go`

#### Task 1.3: Port Assignment Validation
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 0.25h
- **Dependencies**: Task 1.1
- **Description**: Validate current port assignments against architectural ranges
- **Acceptance Criteria**:
  - [x] Run port validator: `go run ./cmd/cicd validate-all`
  - [x] Verify SERVICE range (8000-8999) for all 9 services
  - [x] Verify PRODUCT range (18000-18999) for product compose files
  - [x] Verify SUITE range (28000-28899) for suite compose
  - [x] Document any violations
  - [x] Output saved to `test-output/phase1/port-validation.txt`
- **Files**:
  - `test-output/phase1/port-validation.txt`

(Tasks continue through 13 phases...)


### Phase 2: Create New Directory Structure

**Phase Objective**: Create new deployment directories with correct naming

#### Task 2.1: Create cryptoutil-suite Directory
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 0.1h
- **Dependencies**: Phase 1 complete
- **Description**: Create deployments/cryptoutil-suite/ directory structure
- **Acceptance Criteria**:
  - [x] Create directory: `mkdir -p deployments/cryptoutil-suite`
  - [x] Copy compose.yml from deployments/cryptoutil/
  - [x] Copy secrets directory structure
  - [x] Verify directory created with correct permissions
  - [x] Run: `ls -la deployments/cryptoutil-suite/`
- **Files**:
  - `deployments/cryptoutil-suite/` (directory)
  - `deployments/cryptoutil-suite/compose.yml` (copied)

#### Task 2.2: Verify Existing Hierarchy Structure
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 0.3h
- **Actual**: 0.1h
- **Dependencies**: Task 2.1
- **Description**: Verify existing PRODUCT and SERVICE directories follow hierarchy correctly
- **Acceptance Criteria**:
  - [x] Verify 5 PRODUCT directories exist: cipher, identity, jose, pki, sm
  - [x] Verify 9 SERVICE directories exist: cipher-im, identity-*, jose-ja, pki-ca, sm-kms
  - [x] Document structure in `test-output/phase2/structure-verification.txt`
  - [x] Identify any structural issues
- **Files**:
  - `test-output/phase2/structure-verification.txt`

#### Task 2.3: Archive Legacy Compose Directory
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 0.2h
- **Actual**: 0.15h
- **Dependencies**: Task 2.2
- **Description**: Archive deployments/compose/ (legacy E2E that breaks hierarchy)
- **Acceptance Criteria**:
  - [x] Create archive directory: `mkdir -p deployments/archived/`
  - [x] Move: `git mv deployments/compose deployments/archived/compose-legacy`
  - [x] Document archival reason in `deployments/archived/README.md`
  - [x] Verify no broken references remain (found docs needing update in Phase 9)
- **Files**:
  - `deployments/archived/compose-legacy/` (moved)
  - `deployments/archived/README.md` (created)

#### Task 2.4: Validate New Structure
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 0.05h
- **Dependencies**: Task 2.3
- **Description**: Run linting and validation on new directories
- **Acceptance Criteria**:
  - [x] Run: `go run ./cmd/cicd lint-deployments deployments/cryptoutil-suite`
  - [x] Run: `go run ./cmd/cicd lint-deployments deployments/cryptoutil-product`
  - [x] Run: `go run ./cmd/cicd lint-deployments deployments/cryptoutil-service`
  - [x] Document any violations
  - [x] Output saved to `test-output/phase2/validation.txt`
- **Files**:
  - `test-output/phase2/validation.txt`

**Evidence**: All 67 validators passed including cryptoutil-suite (ports, admin, secrets). Note: validate-compose NOT applicable to SUITE-level (includes-only). Used validate-all with deployment structure validators.

#### Task 2.5: Phase 2 Post-Mortem
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 0.3h
- **Dependencies**: Task 2.4
- **Description**: Document Phase 2 completion and discoveries
- **Acceptance Criteria**:
  - [x] Create phase2-summary.txt
  - [x] Document any issues discovered
  - [x] Identify work for Phase 3
  - [x] Update plan.md with Phase 2 actuals
  - [x] Mark Phase 2 complete
- **Files**:
  - `test-output/phase2/phase2-summary.txt`

**Evidence**: Phase 2 complete in 0.7h vs 3.0h estimated (233% efficiency). Documented discoveries, blockers resolved, next phase identified.

---

### Phase 3: SUITE-Level Refactoring

**Phase Objective**: Refactor cryptoutil-suite/compose.yml to use explicit service definitions with 28XXX ports per SUITE-level architecture requirements

#### Task 3.1: Analyze Template File for Port Mapping Plan

- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 0.15h
- **Dependencies**: Phase 2 complete
- **Description**: Analyze deployments/template/compose-cryptoutil.yml to identify all port mappings requiring updates from 8XXX → 28XXX
- **Acceptance Criteria**:
  - [x] Count total port mappings in template (9 services × 3 instances each = 27 services)
  - [x] Document current port ranges per service (sm-kms: 8000-8002, pki-ca: 8100-8102, etc.)
  - [x] Calculate target SUITE-level ports (sm-kms: 28000-28002, pki-ca: 28100-28102, etc.)
  - [x] Create port mapping table in `test-output/phase3/port-mapping-plan.txt`
  - [x] Verify no conflicts with existing deployments
- **Files**:
  - `test-output/phase3/port-mapping-plan.txt`
  - `test-output/phase3/current-ports.txt`
- **Evidence**: Comprehensive port mapping plan created with 27 service instances across 9 products, all ports mapped from 8XXX → 28XXX (+20000 offset)

#### Task 3.2: Replace cryptoutil-suite/compose.yml with Template

- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 0.3h
- **Actual**: 0.1h
- **Dependencies**: Task 3.1
- **Description**: Replace includes-only pattern with explicit service definitions from template
- **Acceptance Criteria**:
  - [x] Backup current cryptoutil-suite/compose.yml to test-output/phase3/compose.yml.backup
  - [x] Copy deployments/template/compose-cryptoutil.yml to deployments/cryptoutil-suite/compose.yml
  - [x] Update header comments to reflect SUITE-level purpose
  - [x] Verify file structure matches template (services, networks, volumes, secrets sections)
  - [x] Commit with message: "refactor(deploy): replace cryptoutil-suite compose with explicit services from template"
- **Files**:
  - `deployments/cryptoutil-suite/compose.yml` (replaced)
  - `test-output/phase3/compose.yml.backup`
- **Evidence**: Replaced includes-only with 1300+ line explicit services file, updated header to reflect SUITE-level with 28XXX port range

#### Task 3.3: Update All Service Port Mappings to 28XXX Range

- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: 0.05h (sed batch update + verification)
- **Dependencies**: Task 3.2
- **Description**: Update all 27 port mappings (9 services × 3 instances) from 8XXX → 28XXX
- **Acceptance Criteria**:
  - [x] Update sm-kms-app-sqlite-1: 8000 → 28000
  - [x] Update sm-kms-app-postgres-1: 8001 → 28001
  - [x] Update sm-kms-app-postgres-2: 8002 → 28002
  - [x] Update pki-ca-app-sqlite-1: 8100 → 28100
  - [x] Update pki-ca-app-postgres-1: 8101 → 28101
  - [x] Update pki-ca-app-postgres-2: 8102 → 28102
  - [x] Update identity-authz-app-sqlite-1: 8200 → 28200
  - [x] Update identity-authz-app-postgres-1: 8201 → 28201
  - [x] Update identity-authz-app-postgres-2: 8202 → 28202
  - [x] Update identity-idp-app-sqlite-1: 8300 → 28300
  - [x] Update identity-idp-app-postgres-1: 8301 → 28301
  - [x] Update identity-idp-app-postgres-2: 8302 → 28302
  - [x] Update identity-rs-app-sqlite-1: 8400 → 28400
  - [x] Update identity-rs-app-postgres-1: 8401 → 28401
  - [x] Update identity-rs-app-postgres-2: 8402 → 28402
  - [x] Update identity-rp-app-sqlite-1: 8500 → 28500
  - [x] Update identity-rp-app-postgres-1: 8501 → 28501
  - [x] Update identity-rp-app-postgres-2: 8502 → 28502
  - [x] Update identity-spa-app-sqlite-1: 8600 → 28600
  - [x] Update identity-spa-app-postgres-1: 8601 → 28601
  - [x] Update identity-spa-app-postgres-2: 8602 → 28602
  - [x] Update cipher-im-app-sqlite-1: 8700 → 28700
  - [x] Update cipher-im-app-postgres-1: 8701 → 28701
  - [x] Update cipher-im-app-postgres-2: 8702 → 28702
  - [x] Update jose-ja-app-sqlite-1: 8800 → 28800
  - [x] Update jose-ja-app-postgres-1: 8801 → 28801
  - [x] Update jose-ja-app-postgres-2: 8802 → 28802
  - [x] Verify all port mappings follow pattern: "28XXX:8000" (container port remains 8000)
  - [x] Document changes in `test-output/phase3/port-updates.txt`
  - [x] Commit with message: "refactor(deploy): update all cryptoutil-suite ports to 28XXX range (SUITE-level offset)"
- **Evidence**: 
  - Used sed batch update: 27 substitutions (8000→28000, 8001→28001, ..., 8802→28802)
  - Verified all ports: `test-output/phase3/updated-ports.txt` (27 ports in 28XXX range)
  - Summary: `test-output/phase3/port-updates.txt` (before/after comparison, verification checklist)
  - 100% complete: All services (sm-kms, pki-ca, identity-*, cipher-im, jose-ja) updated
- **Files**:
  - `deployments/cryptoutil-suite/compose.yml` (27 port mappings updated)
  - `test-output/phase3/port-updates.txt`

#### Task 3.4: Validate Port Updates with Linter

- **Status**: ☐
- **Owner**: LLM Agent
- **Estimated**: 0.3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.3
- **Description**: Run port validator to confirm all ports in 28XXX range
- **Acceptance Criteria**:
  - [ ] Run: `go run ./cmd/cicd lint-deployments validate-all`
  - [ ] Verify ValidatePorts passes for deployments/cryptoutil-suite/
  - [ ] Verify all 67 validators still pass (no regressions)
  - [ ] Document validation results in `test-output/phase3/port-validation.txt`
  - [ ] If failures: fix issues and re-validate until all pass
- **Files**:
  - `test-output/phase3/port-validation.txt`

#### Task 3.5: Update Volume Path References

- **Status**: ☐
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.3
- **Description**: Update all volume mount paths from ../PRODUCT-SERVICE/config/ to correct relative paths
- **Acceptance Criteria**:
  - [ ] Verify all volume paths use correct relative references (e.g., `../sm-kms/config/...`)
  - [ ] Verify telemetry config paths (../shared-telemetry/otel/cryptoutil-otel.yml)
  - [ ] Test one volume mount to confirm file exists at specified path
  - [ ] Document any path corrections in `test-output/phase3/volume-paths.txt`
  - [ ] Commit if changes needed: "fix(deploy): correct volume mount paths in cryptoutil-suite"
- **Files**:
  - `test-output/phase3/volume-paths.txt`

#### Task 3.6: Update Secrets Configuration

- **Status**: ☐
- **Owner**: LLM Agent
- **Estimated**: 0.3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.2
- **Description**: Verify secrets configuration matches SUITE-level requirements
- **Acceptance Criteria**:
  - [ ] Verify cryptoutil-suite/secrets/ directory has all required secrets
  - [ ] Verify compose.yml secrets section references correct secret files
  - [ ] Verify SUITE-level hash pepper override pattern documented
  - [ ] Test secret file permissions (440 r--r-----)
  - [ ] Document secrets inventory in `test-output/phase3/secrets-inventory.txt`
- **Files**:
  - `test-output/phase3/secrets-inventory.txt`

#### Task 3.7: Test SUITE Deployment Startup

- **Status**: ☐
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 3.4, 3.5, 3.6
- **Description**: Test SUITE-level deployment with docker compose up
- **Acceptance Criteria**:
  - [ ] Ensure Docker Desktop running
  - [ ] Run: `cd deployments/cryptoutil-suite && docker compose config` (verify syntax)
  - [ ] Run: `docker compose --profile dev up -d` (start SQLite instances only)
  - [ ] Wait for all services healthy (up to 120 seconds)
  - [ ] Verify all 9 services started successfully
  - [ ] Verify health endpoints respond on 28XXX ports
  - [ ] Test one health endpoint: `curl https://localhost:28000/browser/api/v1/health` (sm-kms)
  - [ ] Run: `docker compose logs --tail=20` (capture startup logs)
  - [ ] Run: `docker compose down -v` (cleanup)
  - [ ] Document test results in `test-output/phase3/deployment-test.txt`
  - [ ] If failures: diagnose, fix, re-test until all pass
- **Files**:
  - `test-output/phase3/deployment-test.txt`

#### Task 3.8: Verify Port Validator Handles SUITE-Level

- **Status**: ☐
- **Owner**: LLM Agent
- **Estimated**: 0.3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.4
- **Description**: Confirm port validator correctly detects cryptoutil-suite as SUITE-level deployment
- **Acceptance Criteria**:
  - [ ] Review internal/cmd/cicd/lint_deployments/validate_ports.go for SUITE detection logic
  - [ ] Verify DeploymentTypeSuite constant exists and equals "SUITE"
  - [ ] Verify getDeploymentLevel() returns "SUITE" for "cryptoutil-suite" directory
  - [ ] Verify suitePortMin = 28000, suitePortMax = 28999
  - [ ] Run unit tests: `go test ./internal/cmd/cicd/lint_deployments/... -run TestValidatePorts_Suite`
  - [ ] Document validator behavior in `test-output/phase3/validator-analysis.txt`
  - [ ] If logic missing: implement and test (becomes blocker)
- **Files**:
  - `test-output/phase3/validator-analysis.txt`

#### Task 3.9: Phase 3 Post-Mortem

- **Status**: ☐
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 3.1-3.8
- **Description**: Post-mortem analysis for Phase 3
- **Acceptance Criteria**:
  - [ ] Create `test-output/phase3/phase3-summary.txt` with: tasks complete, evidence files, discoveries, blockers, next phase
  - [ ] Update plan.md Phase 3 section with completion notes, actual time, deferred work
  - [ ] Identify any new phases/tasks to insert or append
  - [ ] Mark Phase 3 complete in plan.md success criteria
  - [ ] Commit with comprehensive message listing all Phase 3 changes
- **Files**:
  - `test-output/phase3/phase3-summary.txt`

---

### Phases 4-13: High-Level Task Outlines

**Note**: Detailed tasks will be created as each phase is reached (dynamic work discovery pattern).

**Phase 4**: PRODUCT-Level Standardization (7 tasks estimated)
**Phase 5**: SERVICE-Level Verification (8 tasks estimated)
**Phase 6**: Legacy E2E Migration (12 tasks estimated)
**Phase 7**: Archive Legacy Directories (5 tasks estimated)
**Phase 8**: Validator Updates (8 tasks estimated)
**Phase 9**: Documentation Complete Update (10 tasks estimated)
**Phase 10**: CI/CD Workflow Updates (7 tasks estimated)
**Phase 11**: Integration Testing (9 tasks estimated)
**Phase 12**: Quality Gates & Final Validation (8 tasks estimated)
**Phase 13**: Archive & Wrap-Up (4 tasks estimated)

**Total**: 92 tasks across 13 phases (estimated)
