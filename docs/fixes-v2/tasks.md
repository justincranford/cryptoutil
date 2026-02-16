# Tasks - Deployment & Config Structure Refactoring V2

**Status**: 0 of 54 tasks complete (0%)
**Last Updated**: 2026-02-16
**Created**: 2026-02-16

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code functionally correct with comprehensive tests
- ✅ **Completeness**: NO steps skipped, ALL features implemented
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability, NOT speed
- ✅ **Accuracy**: Address root cause, not symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation
- ❌ **Premature Completion**: NEVER mark complete without evidence

**ALL issues are blockers - Fix immediately, document root causes, NEVER defer**

---

## Phase 0.5: Demo Files Archive

**Phase Objective**: Archive demo files for future research

#### Task 0.5.1: Create Demo Brainstorm Directory Structure
- **Status**: ❌ | **Est**: 15min | **Owner**: Agent
- **Description**: Create directories for demo file archiving
- **Acceptance Criteria**:
  - [ ] mkdir -p docs/demo-brainstorm/archive/
  - [ ] Git commit: "feat: create demo brainstorm directory"
- **Commands**:
  ```bash
  mkdir -p docs/demo-brainstorm/archive/
  git add docs/demo-brainstorm/
  git commit -m "feat: create demo brainstorm directory"
  ```

#### Task 0.5.2: Archive Demo Compose File
- **Status**: ❌ | **Est**: 15min | **Owner**: Agent
- **Description**: Move sm-kms demo file to archive
- **Acceptance Criteria**:
  - [ ] `git mv deployments/sm-kms/compose.demo.yml docs/demo-brainstorm/archive/sm-kms-compose.demo.yml`
  - [ ] Verify file moved (not copied)
  - [ ] Git commit: "refactor: archive demo compose for future research"

#### Task 0.5.3: Create Demo Brainstorm Stub
- **Status**: ❌ | **Est**: 30min | **Owner**: Agent
- **Description**: Create DEMO-BRAINSTORM.md placeholder
- **Acceptance Criteria**:
  - [ ] Create docs/demo-brainstorm/DEMO-BRAINSTORM.md
  - [ ] Document current state (sm-kms demo archived)
  - [ ] Outline research questions for future phase
  - [ ] Git commit: "docs: add demo brainstorm placeholder"
- **Files**: `docs/demo-brainstorm/DEMO-BRAINSTORM.md`

---

## Phase 1: Structural Cleanup

**Phase Objective**: Remove redundant files, populate template

#### Task 1.1: Delete Redundant .gitkeep Files
- **Status**: ❌ | **Est**: 10min | **Owner**: Agent
- **Description**: Remove .gitkeep from non-empty directories
- **Acceptance Criteria**:
  - [ ] Delete deployments/cipher-im/config/.gitkeep
  - [ ] Delete configs/.gitkeep
  - [ ] Verify directories still exist
  - [ ] Git commit: "chore: remove redundant .gitkeep files"
- **Evidence**: test-output/phase1/gitkeep-analysis.txt

#### Task 1.2: Delete Duplicate Otel-Collector Configs
- **Status**: ❌ | **Est**: 15min | **Owner**: Agent
- **Description**: Remove duplicate otel configs, keep only canonical
- **Acceptance Criteria**:
  - [ ] Delete deployments/template/otel-collector-config.yaml
  - [ ] Delete deployments/cipher-im/otel-collector-config.yaml
  - [ ] Verify shared-telemetry/otel/otel-collector-config.yaml exists
  - [ ] Git commit: "chore: remove duplicate otel-collector configs"
- **Evidence**: test-output/phase1/otel-config-analysis.txt

#### Task 1.3: Create Template Config Files
- **Status**: ❌ | **Est**: 1h | **Owner**: Agent
- **Description**: Populate template/config/ with placeholder configs
- **Acceptance Criteria**:
  - [ ] Create PRODUCT-SERVICE-app-common.yml (placeholder pattern)
  - [ ] Create PRODUCT-SERVICE-app-sqlite-1.yml
  - [ ] Create PRODUCT-SERVICE-app-postgresql-1.yml
  - [ ] Create PRODUCT-SERVICE-app-postgresql-2.yml
  - [ ] Use uppercase placeholders (PRODUCT, SERVICE) for clarity
  - [ ] Add comments explaining template usage
  - [ ] Git commit: "feat: add template config files"
- **Files**:
  - `deployments/template/config/PRODUCT-SERVICE-app-common.yml`
  - `deployments/template/config/PRODUCT-SERVICE-app-sqlite-1.yml`
  - `deployments/template/config/PRODUCT-SERVICE-app-postgresql-1.yml`
  - `deployments/template/config/PRODUCT-SERVICE-app-postgresql-2.yml`

#### Task 1.4: Create RATIONALE.md
- **Status**: ❌ | **Est**: 30min | **Owner**: Agent
- **Description**: Document all cleanup decisions
- **Acceptance Criteria**:
  - [ ] Document .gitkeep deletion rationale
  - [ ] Document otel-collector canonical source decision
  - [ ] Document template/config/ population
  - [ ] Document demo archiving decision
  - [ ] Git commit: "docs: add structural cleanup rationale"
- **Files**: `docs/fixes-v2/RATIONALE.md`

---

## Phase 2: Enhance Docker Compose Validation

**Phase Objective**: Prevent schema errors reaching commits

#### Task 2.1: Add Docker Compose Schema Validation
- **Status**: ❌ | **Est**: 1.5h | **Owner**: Agent
- **Description**: Enhance lint-compose with schema checks
- **Acceptance Criteria**:
  - [ ] Add function to run `docker compose -f FILE config --quiet`
  - [ ] Parse validation errors
  - [ ] Return violations list
  - [ ] Test against all 24 compose files
  - [ ] Tests with ≥98% coverage
- **Files**:
  - `internal/cmd/cicd/lint_compose/lint_compose_schema.go`
  - `internal/cmd/cicd/lint_compose/lint_compose_schema_test.go`

#### Task 2.2: Integrate Schema Validation into lint-compose
- **Status**: ❌ | **Est**: 1h | **Owner**: Agent
- **Description**: Call schema validation in main lint flow
- **Acceptance Criteria**:
  - [ ] Update lint_compose.go to call schema validation
  - [ ] Aggregate schema + existing violations
  - [ ] Test with intentionally broken compose file
  - [ ] Verify catches VS Code validation errors
- **Files**: `internal/cmd/cicd/lint_compose/lint_compose.go`

#### Task 2.3: Test Enhanced Compose Validation
- **Status**: ❌ | **Est**: 30min | **Owner**: Agent
- **Description**: Validate against all compose files
- **Acceptance Criteria**:
  - [ ] Run against all 24 compose files
  - [ ] No false positives
  - [ ] Catches intentional errors
  - [ ] Evidence in test-output/phase2/compose-validation.log
  - [ ] Git commit: "feat: add Docker Compose schema validation"

---

## Phase 3: CICD Refactoring - Deployments

**Phase Objective**: Comprehensive deployment validation

#### Task 3.1: Complete Deployment File Lists
- **Status**: ❌ | **Est**: 2h | **Owner**: Agent
- **Description**: Comprehensive lists of all expected files
- **Acceptance Criteria**:
  - [ ] Suite directories list: [cryptoutil]
  - [ ] Product directories list: [sm, pki, identity, cipher, jose]
  - [ ] Service directories list: [9 services]
  - [ ] Shared directories list: [shared-postgres, shared-citus, shared-telemetry, template]
  - [ ] Complete file list for each directory type
  - [ ] Filtering functions (getSuiteFiles, getProductFiles, getServiceFiles)
  - [ ] Tests ≥98% coverage
- **Files**: `internal/cmd/cicd/lint_deployments/lint_required_contents_deployments.go`

#### Task 3.2: Add Credential Validation
- **Status**: ❌ | **Est**: 2h | **Owner**: Agent
- **Description**: Detect hardcoded credentials in compose files
- **Acceptance Criteria**:
  - [ ] Scan all compose.yml files
  - [ ] Check for hardcoded passwords, usernames, database names
  - [ ] Check for hardcoded pepper values
  - [ ] Check for hardcoded unseal secrets
  - [ ] Return violations with file:line
  - [ ] Tests ≥98% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/lint_credentials.go`
  - `internal/cmd/cicd/lint_deployments/lint_credentials_test.go`

#### Task 3.3: Validate Template Directory
- **Status**: ❌ | **Est**: 1h | **Owner**: Agent
- **Description**: Ensure template has all required files
- **Acceptance Criteria**:
  - [ ] Validate 4 compose template files exist
  - [ ] Validate 4 config files in template/config/
  - [ ] Validate otel-collector-config.yaml NOT in template (deleted)
  - [ ] Tests ≥98% coverage
- **Files**: Update lint_required_contents_deployments.go

#### Task 3.4: Validate Shared Directories
- **Status**: ❌ | **Est**: 1h | **Owner**: Agent
- **Description**: Validate shared-* directory structures
- **Acceptance Criteria**:
  - [ ] Validate shared-postgres structure
  - [ ] Validate shared-citus structure
  - [ ] Validate shared-telemetry structure (canonical otel config)
  - [ ] Tests ≥98% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/lint_shared_directories.go`
  - `internal/cmd/cicd/lint_deployments/lint_shared_directories_test.go`

---

## Phase 4: CICD Refactoring - Configs

**Phase Objective**: Establish rigid ./configs/ validation

#### Task 4.1: Design Config Structure Lists
- **Status**: ❌ | **Est**: 1.5h | **Owner**: Agent
- **Description**: Define expected ./configs/ structure (exact mirror of deployments)
- **Acceptance Criteria**:
  - [ ] Suite directory: configs/cryptoutil/
  - [ ] Product directories: configs/{sm,pki,identity,cipher,jose}/
  - [ ] Service directories: configs/{PRODUCT-SERVICE}/
  - [ ] Expected files per directory type
  - [ ] Document in test-output/phase4/config-structure-design.md
- **Files**: `test-output/phase4/config-structure-design.md`

#### Task 4.2: Implement Config File Validation
- **Status**: ❌ | **Est**: 2h | **Owner**: Agent
- **Description**: Comprehensive ./configs/ validation
- **Acceptance Criteria**:
  - [ ] Complete file lists for expected configs
  - [ ] Suite/product/service directory lists
  - [ ] Filtering functions
  - [ ] Validate against current structure (will fail until Phase 5)
  - [ ] Tests ≥98% coverage
- **Files**: `internal/cmd/cicd/lint_deployments/lint_required_contents_configs.go` (replace 34 lines!)

#### Task 4.3: Add Config Credential Validation
- **Status**: ❌ | **Est**: 1h | **Owner**: Agent
- **Description**: Check config files for hardcoded credentials
- **Acceptance Criteria**:
  - [ ] Scan .yml/.yaml files in ./configs/
  - [ ] Same validation as deployments
  - [ ] Return violations with file:line
  - [ ] Tests ≥98% coverage
- **Files**: Extend lint_credentials.go or create separate

#### Task 4.4: Integrate Config Validation
- **Status**: ❌ | **Est**: 1.5h | **Owner**: Agent
- **Description**: Wire up config validation in CICD
- **Acceptance Criteria**:
  - [ ] Add `lint-configs` subcommand to cmd/cicd
  - [ ] Call config content validation
  - [ ] Call credential validation
  - [ ] Aggregate violations
  - [ ] Tests ≥95% coverage
- **Files**: `cmd/cicd/main.go`, `internal/cmd/cicd/lint_deployments/lint_deployments.go`

---

## Phase 5: Config Directory Restructuring

**Phase Objective**: Migrate all 55 files to rigid structure

#### Task 5.1: Create New Config Directory Structure
- **Status**: ❌ | **Est**: 30min | **Owner**: Agent
- **Description**: Create mirrored directory hierarchy
- **Acceptance Criteria**:
  - [ ] mkdir configs/cryptoutil
  - [ ] mkdir configs/{sm,pki,identity,cipher,jose}
  - [ ] mkdir configs/{PRODUCT-SERVICE} for 9 services
  - [ ] Git commit: "feat: create rigid config directory structure"

#### Task 5.2: Migrate Config Files (Suite Level)
- **Status**: ❌ | **Est**: 30min | **Owner**: Agent
- **Description**: Move suite-level configs
- **Acceptance Criteria**:
  - [ ] Identify suite-level configs (if any)
  - [ ] `git mv` to configs/cryptoutil/
  - [ ] Git commit: "refactor: migrate suite-level configs"

#### Task 5.3: Migrate Config Files (Product Level)
- **Status**: ❌ | **Est**: 1h | **Owner**: Agent
- **Description**: Move product-level configs
- **Acceptance Criteria**:
  - [ ] Move sm configs
  - [ ] Move pki configs
  - [ ] Move identity configs (preserve profiles/)
  - [ ] Move cipher configs
  - [ ] Move jose configs
  - [ ] `git mv` for all files
  - [ ] Git commit: "refactor: migrate product-level configs"

#### Task 5.4: Migrate Config Files (Service Level)
- **Status**: ❌ | **Est**: 2h | **Owner**: Agent
- **Description**: Move service-level configs
- **Acceptance Criteria**:
  - [ ] Move all 9 service configs
  - [ ] Preserve subdirectories (policies/, etc)
  - [ ] `git mv` for all files
  - [ ] Verify all 55 files migrated (none orphaned)
  - [ ] Git commit: "refactor: migrate service-level configs"

#### Task 5.5: Update Code References
- **Status**: ❌ | **Est**: 2h | **Owner**: Agent
- **Description**: Update all config path references
- **Acceptance Criteria**:
  - [ ] Grep for old paths: configs/ca/, configs/cipher/im/, etc
  - [ ] Update Go code references
  - [ ] Update CLI default paths
  - [ ] Update test references
  - [ ] Tests pass: `go test ./...`
  - [ ] Git commit: "refactor: update config path references"

#### Task 5.6: Update Documentation References
- **Status**: ❌ | **Est**: 1h | **Owner**: Agent
- **Description**: Update docs with new config paths
- **Acceptance Criteria**:
  - [ ] Update README.md
  - [ ] Update ARCHITECTURE.md
  - [ ] Update getting started guides
  - [ ] Git commit: "docs: update config path references"

#### Task 5.7: Test Config Runs (All Levels)
- **Status**: ❌ | **Est**: 1h | **Owner**: Agent
- **Description**: Verify suite/product/service config runs work
- **Acceptance Criteria**:
  - [ ] Test suite-level: cryptoutil --config
  - [ ] Test product-level: identity, sm, pki, cipher, jose
  - [ ] Test service-level: all 9 services
  - [ ] Health checks pass
  - [ ] Evidence in test-output/phase5/config-runs.log
- **Files**: `test-output/phase5/config-runs.log`

#### Task 5.8: Validate CICD Against New Structure
- **Status**: ❌ | **Est**: 30min | **Owner**: Agent
- **Description**: Ensure CICD validation passes
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd lint-configs` passes
  - [ ] No violations
  - [ ] All expected files present
  - [ ] Evidence in test-output/phase5/cicd-validation.log

---

## Phase 6: Documentation Updates

**Phase Objective**: Update ARCHITECTURE.md and propagate

#### Task 6.1: Update ARCHITECTURE.md - Configs
- **Status**: ❌ | **Est**: 1.5h | **Owner**: Agent
- **Description**: Document new ./configs/ structure
- **Acceptance Criteria**:
  - [ ] Add section on ./configs/ rigid structure
  - [ ] Document suite/product/service patterns
  - [ ] Comparison table with ./deployments/
  - [ ] Git commit: "docs: add rigorous configs structure"
- **Files**: `docs/ARCHITECTURE.md`

#### Task 6.2: Update ARCHITECTURE.md - CICD
- **Status**: ❌ | **Est**: 1h | **Owner**: Agent
- **Description**: Document enhanced CICD validation
- **Acceptance Criteria**:
  - [ ] Document lint_required_contents enhancements
  - [ ] Document credential validation
  - [ ] Document Docker Compose schema validation
  - [ ] Document shared directory validation
  - [ ] Git commit: "docs: document enhanced CICD validation"

#### Task 6.3: Update ARCHITECTURE.md - Otel Config
- **Status**: ₓ | **Est**: 30min | **Owner**: Agent
- **Description**: Document canonical otel-collector config
- **Acceptance Criteria**:
  - [ ] Document shared-telemetry as canonical source
  - [ ] Document deletion of duplicates
  - [ ] Update telemetry architecture section
  - [ ] Git commit: "docs: document canonical otel config"

#### Task 6.4: Update ARCHITECTURE-COMPOSE-MULTIDEPLOY.md
- **Status**: ❌ | **Est**: 45min | **Owner**: Agent
- **Description**: Update compose deployment docs
- **Acceptance Criteria**:
  - [ ] Update template section (config files added)
  - [ ] Update otel-collector references
  - [ ] Cross-reference ARCHITECTURE.md
  - [ ] Git commit: "docs: update compose deployment patterns"

#### Task 6.5: Propagate to Instruction Files
- **Status**: ❌ | **Est**: 30min | **Owner**: Agent
- **Description**: Update instruction files with new patterns
- **Acceptance Criteria**:
  - [ ] Check .github/instructions/*.instructions.md
  - [ ] Update deployment instructions
  - [ ] Update data-infrastructure instructions
  - [ ] Git commit: "docs: propagate config structure to instructions"

---

## Phase 7: Quality Gates

**Phase Objective**: ALL quality requirements met

#### Task 7.1: Build Verification
- **Status**: ❌ | **Est**: 15min | **Owner**: Agent
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean (zero errors)
  - [ ] Evidence: test-output/phase7/build.log

#### Task 7.2: Unit Tests
- **Status**: ❌ | **Est**: 30min | **Owner**: Agent
- **Acceptance Criteria**:
  - [ ] `go test ./... -shuffle=on` passes 100%
  - [ ] Zero skipped tests
  - [ ] Evidence: test-output/phase7/unit-tests.log

#### Task 7.3: Coverage Check
- **Status**: ❌ | **Est**: 30min | **Owner**: Agent
- **Acceptance Criteria**:
  - [ ] Overall coverage ≥95%
  - [ ] CICD linting code ≥98%
  - [ ] Evidence: test-output/phase7/coverage/

#### Task 7.4: Linting
- **Status**: ❌ | **Est**: 15min | **Owner**: Agent
- **Acceptance Criteria**:
  - [ ] `golangci-lint run --fix` then `golangci-lint run` clean (0 issues)
  - [ ] Evidence: test-output/phase7/linting.log

#### Task 7.5: Pre-Commit Checks
- **Status**: ❌ | **Est**: 15min | **Owner**: Agent
- **Acceptance Criteria**:
  - [ ] `pre-commit run --all-files` passes
  - [ ] Evidence: test-output/phase7/pre-commit.log

#### Task 7.6: Integration Tests
- **Status**: ❌ | **Est**: 45min | **Owner**: Agent
- **Acceptance Criteria**:
  - [ ] Integration tests pass
  - [ ] Evidence: test-output/phase7/integration-tests.log

#### Task 7.7: E2E Tests
- **Status**: ❌ | **Est**: 1h | **Owner**: Agent
- **Acceptance Criteria**:
  - [ ] E2E tests using deployments/compose/compose.yml pass
  - [ ] Evidence: test-output/phase7/e2e-tests.log

#### Task 7.8: Race Detector
- **Status**: ❌ | **Est**: 45min | **Owner**: Agent
- **Acceptance Criteria**:
  - [ ] `go test -race -count=2 ./...` passes
  - [ ] Evidence: test-output/phase7/race-detector.log

#### Task 7.9: Mutation Testing
- **Status**: ❌ | **Est**: 2h | **Owner**: Agent
- **Acceptance Criteria**:
  - [ ] gremlins unleash ≥95% (≥98% CICD)
  - [ ] Evidence: test-output/phase7/mutations/

---

## Evidence Archive

- `test-output/phase1/` - Initial analysis (COMPLETE)
- `test-output/phase2/` - Compose validation
- `test-output/phase3/` - Deployments CICD
- `test-output/phase4/` - Configs CICD
- `test-output/phase5/` - Config restructuring
- `test-output/phase6/` - Documentation
- `test-output/phase7/` - Quality gates
