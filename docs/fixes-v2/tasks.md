# Tasks - Deployment & Config Structure Refactoring V2

**Status**: 0 of 78 tasks complete (0%)
**Last Updated**: 2026-02-16
**Created**: 2026-02-16

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

- ✅ **Fix issues immediately** - When tests fail or quality gates not met, STOP and address
- ✅ **Treat as BLOCKING**: ALL issues block progress to next task
- ✅ **Document root causes** - Root cause analysis part of every fix
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark task complete with known issues
- ✅ **NEVER de-prioritize quality** - Evidence-based verification ALWAYS highest priority

---

## Task Checklist

### Phase 1: Investigation & Analysis

**Phase Objective**: Understand current state, identify redundancies, establish patterns

#### Task 1.1: Catalog .gitkeep Files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Find all .gitkeep files, determine which are in non-empty directories
- **Acceptance Criteria**:
  - [ ] Complete list of .gitkeep files with directory status
  - [ ] Identified files to delete (non-empty dirs only)
  - [ ] Document in test-output/phase1/gitkeep-analysis.txt
- **Commands**:
  ```bash
  find ./deployments ./configs -name ".gitkeep" -type f > test-output/phase1/gitkeep-files.txt
  # For each, check if directory has other files
  ```
- **Files**:
  - `test-output/phase1/gitkeep-analysis.txt`

#### Task 1.2: Analyze Compose Files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Analyze deployments/compose/compose.yml and sm-kms/compose.demo.yml purposes
- **Acceptance Criteria**:
  - [ ] Document deployments/compose/ usage (E2E tests)
  - [ ] Check for suite/product level demo patterns
  - [ ] Determine if compose.demo.yml is redundant
  - [ ] Document findings in test-output/phase1/compose-analysis.md
- **Commands**:
  ```bash
  grep -r "compose.demo" ./deployments --include="*.yml"
  find ./deployments -name "compose.demo.yml" -o -name "compose.e2e.yml"
  head -50 ./deployments/cryptoutil/compose.yml  # Check for demo support
  ```
- **Files**:
  - `test-output/phase1/compose-analysis.md`

#### Task 1.3: Analyze otel-collector-config.yaml
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Determine canonical otel-collector config, identify duplicates
- **Acceptance Criteria**:
  - [ ] Identify all otel-collector-config.yaml locations
  - [ ] Compare file contents (diff)
  - [ ] Determine canonical source (shared-telemetry/otel/)
  - [ ] List files to delete
  - [ ] Document in test-output/phase1/otel-config-analysis.txt
- **Commands**:
  ```bash
  find ./deployments -name "otel-collector-config.yaml"
  diff ./deployments/template/otel-collector-config.yaml ./deployments/shared-telemetry/otel/otel-collector-config.yaml
  ```
- **Files**:
  - `test-output/phase1/otel-config-analysis.txt`

#### Task 1.4: Understand template/config/ Empty Directory
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Document why deployments/template/config/ is empty (intentional pattern)
- **Acceptance Criteria**:
  - [ ] Review template directory purpose
  - [ ] Document rationale for empty config/ directory
  - [ ] Update ARCHITECTURE.md section on templates
- **Files**:
  - `test-output/phase1/template-analysis.txt`

#### Task 1.5: Catalog ./configs/ Structure
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Document complete ./configs/ directory structure
- **Acceptance Criteria**:
  - [ ] Tree view of all directories
  - [ ] List all .yml/.yaml files with sizes
  - [ ] Identify patterns (profiles/, policies/, etc)
  - [ ] Document in test-output/phase1/configs-structure.txt
- **Commands**:
  ```bash
  tree -L 4 ./configs/ > test-output/phase1/configs-tree.txt
  find ./configs -type f \( -name "*.yml" -o -name "*.yaml" \) -exec ls -lh {} \; > test-output/phase1/configs-files.txt
  ```
- **Files**:
  - `test-output/phase1/configs-structure.txt`
  - `test-output/phase1/configs-tree.txt`
  - `test-output/phase1/configs-files.txt`

#### Task 1.6: Compare Deployments vs Configs Patterns
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.5
- **Description**: Analyze differences between ./deployments/ and ./configs/organization
- **Acceptance Criteria**:
  - [ ] Comparison table of structures
  - [ ] Identify what patterns should be shared
  - [ ] Identify what patterns should differ
  - [ ] Propose ./configs/ rigid structure
  - [ ] Document in test-output/phase1/structure-comparison.md
- **Files**:
  - `test-output/phase1/structure-comparison.md`

#### Task 1.7: Review Existing CICD Lint Implementation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Understand current lint_deployments implementation
- **Acceptance Criteria**:
  - [ ] Review lint_required_contents_deployments.go (243 lines)
  - [ ] Review lint_required_contents_configs.go (34 lines - incomplete!)
  - [ ] Review lint_deployments.go (505 lines)
  - [ ] Identify gaps in validation
  - [ ] Document refactoring needs in test-output/phase1/cicd-gaps.md
- **Files**:
  - `test-output/phase1/cicd-gaps.md`

---

### Phase 2: Structural Cleanup

**Phase Objective**: Remove redundant files, document architectural decisions

#### Task 2.1: Delete Redundant .gitkeep Files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Remove .gitkeep files from non-empty directories
- **Acceptance Criteria**:
  - [ ] Delete identified .gitkeep files
  - [ ] Verify directories still exist (not empty)
  - [ ] Git commit: "chore: remove redundant .gitkeep files"
- **Commands**:
  ```bash
  git rm ./deployments/cipher-im/config/.gitkeep  # Example
  git commit -m "chore: remove redundant .gitkeep files"
  ```

#### Task 2.2: Handle Redundant Compose Files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.2
- **Description**: Delete or document compose files based on analysis
- **Acceptance Criteria**:
  - [ ] Document deployments/compose/ as E2E infrastructure (keep)
  - [ ] Handle sm-kms/compose.demo.yml (delete if redundant, document if kept)
  - [ ] Update ARCHITECTURE.md with decisions
  - [ ] Git commit: "chore: cleanup redundant compose files" OR "docs: document compose file purposes"

#### Task 2.3: Delete Redundant otel-collector Configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.3
- **Description**: Remove duplicate otel-collector-config.yaml files
- **Acceptance Criteria**:
  - [ ] Keep shared-telemetry/otel/otel-collector-config.yaml (canonical)
  - [ ] Evaluate template/ (keep if intentional example, delete if duplicate)
  - [ ] Delete cipher-im/otel-collector-config.yaml if duplicate
  - [ ] Update compose files to reference shared-telemetry config
  - [ ] Git commit: "chore: remove duplicate otel-collector configs"
- **Commands**:
  ```bash
  git rm ./deployments/cipher-im/otel-collector-config.yaml
  # Update compose references if needed
  git commit -m "chore: remove duplicate otel-collector configs"
  ```

#### Task 2.4: Create RATIONALE.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 2.1, 2.2, 2.3
- **Description**: Document all cleanup decisions and rationale
- **Acceptance Criteria**:
  - [ ] Document .gitkeep deletion rationale
  - [ ] Document compose file decisions
  - [ ] Document otel-collector config decisions
  - [ ] Document template/config/ empty directory decision
  - [ ] Place in docs/fixes-v2/RATIONALE.md
  - [ ] Git commit: "docs: add structural cleanup rationale"
- **Files**:
  - `docs/fixes-v2/RATIONALE.md`

---

### Phase 3: CICD Refactoring - Deployments

**Phase Objective**: Enhance lint_deployments with comprehensive validation

#### Task 3.1: Design File Lists Structure
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.7
- **Description**: Design structure for comprehensive file lists
- **Acceptance Criteria**:
  - [ ] Define suite directories list
  - [ ] Define product directories list
  - [ ] Define service directories list
  - [ ] Define expected files per directory type
  - [ ] Document in test-output/phase3/file-lists-design.md
- **Files**:
  - `test-output/phase3/file-lists-design.md`

#### Task 3.2: Refactor lint_required_contents_deployments.go
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.1
- **Description**: Enhance with complete file lists and directory filtering
- **Acceptance Criteria**:
  - [ ] Add suite directory list (cryptoutil)
  - [ ] Add product directory lists (sm, pki, identity, cipher, jose)
  - [ ] Add service directory lists (9 services)
  - [ ] Complete file list for ALL expected files
  - [ ] Add filtering functions (by suite/product/service)
  - [ ] Tests with ≥98% coverage
- **Commands**:
  ```bash
  go test ./internal/cmd/cicd/lint_deployments/ -run TestLintRequiredContents -v
  go test ./internal/cmd/cicd/lint_deployments/ -cover
  ```
- **Files**:
  - `internal/cmd/cicd/lint_deployments/lint_required_contents_deployments.go`
  - `internal/cmd/cicd/lint_deployments/lint_required_contents_deployments_test.go`

#### Task 3.3: Add Credential Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Validate no hardcoded credentials in compose files
- **Acceptance Criteria**:
  - [ ] Function to check compose files for hardcoded passwords
  - [ ] Check for hardcoded database credentials
  - [ ] Check for hardcoded pepper values
  - [ ] Check for hardcoded unseal secrets
  - [ ] Return violations list
  - [ ] Tests with ≥98% coverage
- **Commands**:
  ```bash
  go test ./internal/cmd/cicd/lint_deployments/ -run TestCheckHardcodedCredentials -v
  ```
- **Files**:
  - `internal/cmd/cicd/lint_deployments/lint_credentials.go`
  - `internal/cmd/cicd/lint_deployments/lint_credentials_test.go`

#### Task 3.4: Enhance ValidateDeploymentStructure
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 3.2, 3.3
- **Description**: Update control layer to call all validation functions
- **Acceptance Criteria**:
  - [ ] Call deployments content validation
  - [ ] Call credential validation
  - [ ] Validate shared directories (shared-*, template)
  - [ ] Aggregate all violations
  - [ ] Return comprehensive results
  - [ ] Tests with ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/lint_deployments.go`
  - `internal/cmd/cicd/lint_deployments/lint_deployments_test.go`

#### Task 3.5: Test Deployments Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.4
- **Description**: Run complete validation suite
- **Acceptance Criteria**:
  - [ ] Unit tests: `go test ./internal/cmd/cicd/lint_deployments/` all pass
  - [ ] Coverage: ≥98% for lint files
  - [ ] Integration test: `go run ./cmd/cicd lint-deployments` passes
  - [ ] No false positives
  - [ ] Evidence in test-output/phase3/deployments-validation.log
- **Commands**:
  ```bash
  go test ./internal/cmd/cicd/lint_deployments/ -v -cover
  go run ./cmd/cicd lint-deployments
  ```
- **Files**:
  - `test-output/phase3/deployments-validation.log`

---

### Phase 4: CICD Refactoring - Configs

**Phase Objective**: Establish rigid ./configs/ validation matching ./deployments/ rigor

#### Task 4.1: Design Configs Rigid Structure
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.6
- **Description**: Design rigid structure for ./configs/ based on patterns
- **Acceptance Criteria**:
  - [ ] Define suite/product/service hierarchy
  - [ ] Define expected directories
  - [ ] Define expected files per level
  - [ ] Account for profiles/ and policies/ patterns
  - [ ] Document in test-output/phase4/configs-structure-design.md
- **Files**:
  - `test-output/phase4/configs-structure-design.md`

#### Task 4.2: Implement lint_required_contents_configs.go
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.1
- **Description**: Comprehensive config file validation
- **Acceptance Criteria**:
  - [ ] Complete file list for ALL expected config files
  - [ ] Suite directory list
  - [ ] Product directory lists
  - [ ] Service directory lists
  - [ ] Filtering functions (by suite/product/service)
  - [ ] Tests with ≥98% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/lint_required_contents_configs.go` (replace 34 lines!)
  - `internal/cmd/cicd/lint_deployments/lint_required_contents_configs_test.go`

#### Task 4.3: Add Config Credential Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.3 (can reuse pattern)
- **Description**: Validate no hardcoded credentials in config files
- **Acceptance Criteria**:
  - [ ] Check .yml/.yaml files in ./configs/
  - [ ] Same validation as deployments (passwords, peppers, unseals)
  - [ ] Return violations list
  - [ ] Tests with ≥98% coverage
- **Files**:
  - Reuse `lint_credentials.go` or extend for configs

#### Task 4.4: Validate Shared Directories
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Validate shared-* and template directories
- **Acceptance Criteria**:
  - [ ] Validate shared-citus structure
  - [ ] Validate shared-postgres structure
  - [ ] Validate shared-telemetry structure
  - [ ] Validate template structure
  - [ ] Tests with ≥98% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/lint_shared_directories.go`
  - `internal/cmd/cicd/lint_deployments/lint_shared_directories_test.go`

#### Task 4.5: Enhance ValidateConfigStructure
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 4.2, 4.3, 4.4
- **Description**: Create control layer for config validation
- **Acceptance Criteria**:
  - [ ] Call configs content validation
  - [ ] Call credential validation for configs
  - [ ] Call shared directories validation
  - [ ] Aggregate all violations
  - [ ] Return comprehensive results
  - [ ] Tests with ≥95% coverage
- **Files**:
  - Update `lint_deployments.go` with ValidateConfigStructure function

#### Task 4.6: Test Configs Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.5
- **Description**: Run complete config validation suite
- **Acceptance Criteria**:
  - [ ] Unit tests: all pass
  - [ ] Coverage: ≥98% for lint files
  - [ ] Integration test: `go run ./cmd/cicd lint-configs` passes
  - [ ] No false positives
  - [ ] Evidence in test-output/phase4/configs-validation.log
- **Commands**:
  ```bash
  go test ./internal/cmd/cicd/lint_deployments/ -v -cover
  go run ./cmd/cicd lint-configs  # May need to add subcommand
  ```
- **Files**:
  - `test-output/phase4/configs-validation.log`

---

### Phase 5: Config Directory Restructuring

**Phase Objective**: Apply rigid structure to ./configs/ matching ./deployments/ principles

#### Task 5.1: Create New Config Directory Structure
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.1
- **Description**: Create suite/product/service hierarchy matching design
- **Acceptance Criteria**:
  - [ ] Create configs/cryptoutil/ (suite level)
  - [ ] Create configs/{product}/ (pki, sm, cipher, jose, identity)
  - [ ] Create configs/{product}-{service}/ (9 service dirs)
  - [ ] Maintain existing profiles/ and policies/ where applicable
  - [ ] Git commit: "feat: create rigid config directory structure"
- **Commands**:
  ```bash
  mkdir -p configs/cryptoutil
  mkdir -p configs/sm configs/pki configs/cipher configs/jose configs/identity
  mkdir -p configs/sm-kms configs/pki-ca configs/cipher-im configs/jose-ja
  mkdir -p configs/identity-{authz,idp,rp,rs,spa}
  ```

#### Task 5.2: Migrate Existing Config Files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.1
- **Description**: Move existing config files to new structure
- **Acceptance Criteria**:
  - [ ] Map each existing file to new location
  - [ ] Use `git mv` to preserve history
  - [ ] Maintain profiles/ and policies/ subdirectories
  - [ ] No orphaned files
  - [ ] Git commit: "refactor: migrate configs to rigid structure"
- **Commands**:
  ```bash
  git mv configs/ca/config.yml configs/pki-ca/config.yml  # Example
  # Repeat for all files
  ```

#### Task 5.3: Update Config File References
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.2
- **Description**: Update all code/docs referencing old config paths
- **Acceptance Criteria**:
  - [ ] Grep for old config paths
  - [ ] Update Go code references
  - [ ] Update documentation references
  - [ ] Update CLI default paths
  - [ ] Tests pass: `go test ./...`
  - [ ] Git commit: "refactor: update config path references"
- **Commands**:
  ```bash
  grep -r "configs/ca/" ./internal ./cmd ./docs
  # Update each reference
  go test ./... -shuffle=on
  ```

#### Task 5.4: Test Suite Level Config Runs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.3
- **Description**: Verify suite-level config runs work
- **Acceptance Criteria**:
  - [ ] `./cmd/cryptoutil/main.go --config ./configs/cryptoutil/config.yml` works
  - [ ] All 9 services start successfully
  - [ ] Health checks pass
  - [ ] Document in test-output/phase5/suite-level-test.log
- **Commands**:
  ```bash
  go run ./cmd/cryptoutil --config ./configs/cryptoutil/config.yml --help
  # Test suite startup
  ```
- **Files**:
  - `test-output/phase5/suite-level-test.log`

#### Task 5.5: Test Product Level Config Runs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.3
- **Description**: Verify product-level config runs work
- **Acceptance Criteria**:
  - [ ] Test each product: sm, pki, identity, cipher, jose
  - [ ] Identity product (5 services) starts successfully
  - [ ] Single-service products work
  - [ ] Health checks pass
  - [ ] Document in test-output/phase5/product-level-test.log
- **Files**:
  - `test-output/phase5/product-level-test.log`

#### Task 5.6: Test Service Level Config Runs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.3
- **Description**: Verify service-level config runs work
- **Acceptance Criteria**:
  - [ ] Test each service independently
  - [ ] All 9 services start with their configs
  - [ ] Health checks pass
  - [ ] Document in test-output/phase5/service-level-test.log
- **Files**:
  - `test-output/phase5/service-level-test.log`

#### Task 5.7: Validate CICD Against New Structure
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 5.2, 4.6
- **Description**: Ensure CICD validation passes with new structure
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd lint-configs` passes
  - [ ] No violations found
  - [ ] All expected files present
  - [ ] Document in test-output/phase5/cicd-validation.log
- **Commands**:
  ```bash
  go run ./cmd/cicd lint-configs
  ```
- **Files**:
  - `test-output/phase5/cicd-validation.log`

---

### Phase 6: Documentation Updates

**Phase Objective**: Update ARCHITECTURE.md and propagate to linked docs

#### Task 6.1: Update ARCHITECTURE.md - Configs Structure
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 5 complete
- **Description**: Document new ./configs/ structure in ARCHITECTURE.md
- **Acceptance Criteria**:
  - [ ] Add section on ./configs/ rigid structure
  - [ ] Document suite/product/service patterns
  - [ ] Document profiles/ and policies/ patterns
  - [ ] Add comparison table with ./deployments/
  - [ ] Git commit: "docs: add rigorous configs structure to ARCHITECTURE.md"
- **Files**:
  - `docs/ARCHITECTURE.md` (Section 12.4 or new section)

#### Task 6.2: Update ARCHITECTURE.md - CICD Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phases 3&4 complete
- **Description**: Document enhanced CICD validation
- **Acceptance Criteria**:
  - [ ] Document lint_required_contents_deployments enhancements
  - [ ] Document lint_required_contents_configs implementation
  - [ ] Document credential validation
  - [ ] Document shared directory validation
  - [ ] Git commit: "docs: document enhanced CICD validation"
- **Files**:
  - `docs/ARCHITECTURE.md` (Section 12.4 or new section)

#### Task 6.3: Update ARCHITECTURE-COMPOSE-MULTIDEPLOY.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.4
- **Description**: Update with cleanup decisions and patterns
- **Acceptance Criteria**:
  - [ ] Document deployments/compose/ E2E purpose
  - [ ] Document demo file patterns (if kept)
  - [ ] Document otel-collector config canonical source
  - [ ] Cross-reference ARCHITECTURE.md
  - [ ] Git commit: "docs: update compose deployment patterns"
- **Files**:
  - `docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md`

#### Task 6.4: Propagate to Instruction Files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 6.1, 6.2, 6.3
- **Description**: Update instruction files with new patterns
- **Acceptance Criteria**:
  - [ ] Review .github/instructions/*.instructions.md for config references
  - [ ] Update deployment instructions if needed
  - [ ] Update testing instructions if needed
  - [ ] Update data-infrastructure instructions if needed
  - [ ] Git commit: "docs: propagate config structure to instructions"
- **Files**:
  - `.github/instructions/04-01.deployment.instructions.md` (maybe)
  - `.github/instructions/03-04.data-infrastructure.instructions.md` (maybe)

#### Task 6.5: Update README.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Update README with new config patterns if referenced
- **Acceptance Criteria**:
  - [ ] Check if README references ./configs/
  - [ ] Update examples if present
  - [ ] Update getting started if needed
  - [ ] Git commit: "docs: update README with config patterns"
- **Files**:
  - `README.md`

---

### Phase 7: Quality Gates

**Phase Objective**: Verify all quality requirements met

#### Task 7.1: Build Main Code
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: [Fill when complete]
- **Dependencies**: All code changes complete
- **Description**: Verify clean build
- **Acceptance Criteria**:
  - [ ] `go build ./...` completes with zero errors
  - [ ] No build warnings
  - [ ] Output in test-output/phase7/build-main.log
- **Commands**:
  ```bash
  go build ./... 2>&1 | tee test-output/phase7/build-main.log
  ```
- **Files**:
  - `test-output/phase7/build-main.log`

#### Task 7.2: Build Test Code
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: [Fill when complete]
- **Dependencies**: All test changes complete
- **Description**: Verify test code builds
- **Acceptance Criteria**:
  - [ ] `go test ./... -run=^$` completes with zero errors
  - [ ] All test packages compile
  - [ ] Output in test-output/phase7/build-tests.log
- **Commands**:
  ```bash
  go test ./... -run=^$ 2>&1 | tee test-output/phase7/build-tests.log
  ```
- **Files**:
  - `test-output/phase7/build-tests.log`

#### Task 7.3: Run Unit Tests
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.2
- **Description**: Execute all unit tests
- **Acceptance Criteria**:
  - [ ] `go test ./... -shuffle=on` passes 100%
  - [ ] Zero skipped tests
  - [ ] Output in test-output/phase7/unit-tests.log
- **Commands**:
  ```bash
  go test ./... -shuffle=on -v 2>&1 | tee test-output/phase7/unit-tests.log
  ```
- **Files**:
  - `test-output/phase7/unit-tests.log`

#### Task 7.4: Check Test Coverage
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.3
- **Description**: Verify coverage targets met
- **Acceptance Criteria**:
  - [ ] Overall coverage ≥95%
  - [ ] CICD linting code ≥98%
  - [ ] Coverage report in test-output/phase7/coverage/
- **Commands**:
  ```bash
  go test ./... -coverprofile=test-output/phase7/coverage/coverage.out
  go tool cover -html=test-output/phase7/coverage/coverage.out -o test-output/phase7/coverage/coverage.html
  go tool cover -func=test-output/phase7/coverage/coverage.out | grep total
  ```
- **Files**:
  - `test-output/phase7/coverage/coverage.out`
  - `test-output/phase7/coverage/coverage.html`
  - `test-output/phase7/coverage/coverage-summary.txt`

#### Task 7.5: Run Linting
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: [Fill when complete]
- **Dependencies**: All code changes complete
- **Description**: Verify zero linting issues
- **Acceptance Criteria**:
  - [ ] `golangci-lint run --fix` completes
  - [ ] Follow-up `golangci-lint run` shows 0 issues
  - [ ] Output in test-output/phase7/linting.log
- **Commands**:
  ```bash
  golangci-lint run --fix
  golangci-lint run 2>&1 | tee test-output/phase7/linting.log
  ```
- **Files**:
  - `test-output/phase7/linting.log`

#### Task 7.6: Run Pre-Commit Checks
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.5
- **Description**: Verify all pre-commit hooks pass
- **Acceptance Criteria**:
  - [ ] `pre-commit run --all-files` passes
  - [ ] No formatting issues
  - [ ] No UTF-8 BOM issues
  - [ ] Output in test-output/phase7/pre-commit.log
- **Commands**:
  ```bash
  pre-commit run --all-files 2>&1 | tee test-output/phase7/pre-commit.log
  ```
- **Files**:
  - `test-output/phase7/pre-commit.log`

#### Task 7.7: Run Integration Tests
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.3
- **Description**: Run TestMain-based integration tests
- **Acceptance Criteria**:
  - [ ] Integration tests pass
  - [ ] No PostgreSQL/SQLite errors
  - [ ] Output in test-output/phase7/integration-tests.log
- **Commands**:
  ```bash
  go test ./... -tags=integration -v 2>&1 | tee test-output/phase7/integration-tests.log
  ```
- **Files**:
  - `test-output/phase7/integration-tests.log`

#### Task 7.8: Run E2E Tests
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: All changes complete
- **Description**: Run Docker Compose E2E tests
- **Acceptance Criteria**:
  - [ ] E2E tests pass using deployments/compose/compose.yml
  - [ ] Health checks pass
  - [ ] API tests pass
  - [ ] Output in test-output/phase7/e2e-tests.log
- **Commands**:
  ```bash
  go test ./internal/test/e2e/ -tags=e2e -v 2>&1 | tee test-output/phase7/e2e-tests.log
  ```
- **Files**:
  - `test-output/phase7/e2e-tests.log`

#### Task 7.9: Run Race Detector
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.3
- **Description**: Check for race conditions
- **Acceptance Criteria**:
  - [ ] `go test -race -count=2 ./...` passes
  - [ ] No race conditions detected
  - [ ] Output in test-output/phase7/race-detector.log
- **Commands**:
  ```bash
  go test -race -count=2 ./... 2>&1 | tee test-output/phase7/race-detector.log
  ```
- **Files**:
  - `test-output/phase7/race-detector.log`

#### Task 7.10: Run Mutation Testing
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.4
- **Description**: Validate test quality
- **Acceptance Criteria**:
  - [ ] Mutation testing ≥95% (production code)
  - [ ] Mutation testing ≥98% (CICD lint files)
  - [ ] gremlins unleash --tags=!integration passes
  - [ ] Output in test-output/phase7/mutations/
- **Commands**:
  ```bash
  gremlins unleash --tags=!integration 2>&1 | tee test-output/phase7/mutations/gremlins.log
  ```
- **Files**:
  - `test-output/phase7/mutations/gremlins.log`

---

## Cross-Cutting Tasks

### Testing
- [ ] Unit tests ≥95% coverage (production), ≥98% (infrastructure/CICD)
- [ ] Integration tests pass
- [ ] E2E tests pass (Docker Compose)
- [ ] Mutation testing ≥95% minimum (≥98% CICD)
- [ ] No skipped tests (except documented exceptions)
- [ ] Race detector clean: `go test -race ./...`

### Code Quality
- [ ] Linting passes: `golangci-lint run ./...` (0 issues)
- [ ] No new TODOs without tracking
- [ ] No security vulnerabilities
- [ ] Formatting clean: `gofumpt -s -w ./`
- [ ] Imports organized: `goimports -w ./`
- [ ] Pre-commit hooks passing

### Documentation
- [ ] ARCHITECTURE.md updated with new patterns
- [ ] ARCHITECTURE-COMPOSE-MULTIDEPLOY.md updated
- [ ] Instruction files updated if needed
- [ ] RATIONALE.md created with decisions
- [ ] README.md updated if needed
- [ ] Comments added for complex logic

### CICD Validation
- [ ] lint-deployments command works and passes
- [ ] lint-configs command works and passes (may need to add)
- [ ] All expected files validated
- [ ] No hardcoded credentials detected
- [ ] Shared directories validated

---

## Notes / Deferred Work

[Track decisions deferred to future iterations or blocked tasks]

---

## Evidence Archive

[List test output directories created during this iteration]
- `test-output/phase1/` - Investigation and analysis
- `test-output/phase2/` - Cleanup decisions
- `test-output/phase3/` - Deployments CICD refactoring
- `test-output/phase4/` - Configs CICD refactoring
- `test-output/phase5/` - Config restructuring and testing
- `test-output/phase6/` - Documentation artifacts
- `test-output/phase7/` - Quality gate evidence
