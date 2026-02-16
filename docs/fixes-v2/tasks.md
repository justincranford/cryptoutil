# Tasks - Deployment/Config Refactoring v2

**Status**: 0 of 65 tasks complete (0%)
**Last Updated**: 2026-02-16
**Created**: 2026-02-16

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code functionally correct with comprehensive tests
- ✅ **Completeness**: NO tasks skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥98% coverage/mutation for CICD)
- ✅ **Efficiency**: Optimized for maintainability, NOT implementation speed
- ✅ **Accuracy**: Changes address root cause, not symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation
- ❌ **Premature Completion**: NEVER mark complete without evidence

**ALL issues are blockers - NO exceptions:**
- ✅ **Fix immediately** - When tests fail or quality gates not met, STOP
- ✅ **Treat as BLOCKING** - ALL issues block progress
- ✅ **Document root causes** - Analysis MANDATORY
- ✅ **NEVER defer** - No "fix later"
- ✅ **NEVER de-prioritize quality** - Evidence ALWAYS required

---

## Task Checklist

### Phase 0.5: Demo Files Archiving (2h)

**Phase Objective**: Archive demo files under docs/demo-brainstorm/ for future reference

#### Task 0.5.1: Create Demo Brainstorm Directory Structure
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: 5min
- **Dependencies**: None
- **Description**: Create directory hierarchy for archived demo files
- **Acceptance Criteria**:
  - [x] Directory exists: `docs/demo-brainstorm/`
  - [x] Subdirectories: `deployments/sm-kms/`
  - [x] README.md stub created explaining archive purpose
  - [x] Command: `ls -la docs/demo-brainstorm/`
- **Files**:
  - `docs/demo-brainstorm/README.md`
- **Evidence**: N/A (simple directory creation)

#### Task 0.5.2: Archive sm-kms Demo Compose File
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: 3min
- **Dependencies**: Task 0.5.1
- **Description**: Move compose.demo.yml to archive directory
- **Acceptance Criteria**:
  - [x] File moved: `deployments/sm-kms/compose.demo.yml` → `docs/demo-brainstorm/deployments/sm-kms/compose.demo.yml`
  - [x] Original file deleted from deployments/
  - [x] Git tracks move: `git mv` used
  - [x] Command: `ls deployments/sm-kms/` (no compose.demo.yml)
- **Files**:
  - `docs/demo-brainstorm/deployments/sm-kms/compose.demo.yml` (moved)
- **Evidence**: `git log --follow docs/demo-brainstorm/deployments/sm-kms/compose.demo.yml`

#### Task 0.5.3: Create DEMO-BRAINSTORM.md Documentation
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: 10min
- **Dependencies**: Task 0.5.2
- **Description**: Document demo file archive purpose and usage
- **Acceptance Criteria**:
  - [x] File created: `docs/demo-brainstorm/DEMO-BRAINSTORM.md`
  - [x] Explains archive purpose
  - [x] Lists archived files with rationale
  - [x] Provides guidance for creating future demos
  - [x] Command: `cat docs/demo-brainstorm/DEMO-BRAINSTORM.md`
- **Files**:
  - `docs/demo-brainstorm/DEMO-BRAINSTORM.md`
- **Evidence**: N/A (documentation task)

---

### Phase 1: Structural Cleanup (3h)

**Phase Objective**: Delete redundant files, create missing template configs

#### Task 1.1: Delete Redundant .gitkeep Files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Remove .gitkeep files from directories with other content
- **Acceptance Criteria**:
  - [ ] Deleted: `deployments/cipher-im/config/.gitkeep`
  - [ ] Deleted: `configs/.gitkeep`
  - [ ] Verification: `find deployments/ configs/ -name .gitkeep` (empty result)
  - [ ] Command: `git rm deployments/cipher-im/config/.gitkeep configs/.gitkeep`
- **Files**: None (deletions)
- **Evidence**: `test-output/phase1/gitkeep-analysis.txt` (from planning)

#### Task 1.2: Delete Duplicate OpenTelemetry Configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Remove duplicate otel configs, keep only shared-telemetry version
- **Acceptance Criteria**:
  - [ ] Deleted: `deployments/template/otel-collector-config.yaml`
  - [ ] Deleted: `deployments/cipher-im/otel-collector-config.yaml`
  - [ ] Kept: `deployments/shared-telemetry/otel-collector-config.yaml`
  - [ ] Verification: `find deployments/ -name otel-collector-config.yaml` (1 result only)
  - [ ] Command: `git rm deployments/template/otel-collector-config.yaml deployments/cipher-im/otel-collector-config.yaml`
- **Files**: None (deletions)
- **Evidence**: `test-output/phase1/otel-config-analysis.txt` (from planning)

#### Task 1.3: Create Template Config Files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Create 4 template config files matching sm-kms pattern
- **Acceptance Criteria**:
  - [ ] Created: `deployments/template/config/template-app-common.yml`
  - [ ] Created: `deployments/template/config/template-app-sqlite-1.yml`
  - [ ] Created: `deployments/template/config/template-app-postgresql-1.yml`
  - [ ] Created: `deployments/template/config/template-app-postgresql-2.yml`
  - [ ] Files match sm-kms structure with PRODUCT-SERVICE placeholders
  - [ ] Verification: `ls -la deployments/template/config/` (4 files)
  - [ ] YAML valid: `yamllint deployments/template/config/*.yml`
- **Files**:
  - `deployments/template/config/template-app-common.yml`
  - `deployments/template/config/template-app-sqlite-1.yml`
  - `deployments/template/config/template-app-postgresql-1.yml`
  - `deployments/template/config/template-app-postgresql-2.yml`
- **Evidence**: `test-output/phase1/template-config-analysis.txt` (from planning)

#### Task 1.4: Verify Structural Cleanup
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 1.1, 1.2, 1.3
- **Description**: Verify all Phase 1 changes correct
- **Acceptance Criteria**:
  - [ ] No .gitkeep in dirs with content: `find deployments/ configs/ -name .gitkeep | wc -l` (0)
  - [ ] Single otel config: `find deployments/ -name otel-collector-config.yaml | wc -l` (1)
  - [ ] Template has 4 configs: `ls deployments/template/config/ | wc -l` (4)
  - [ ] All tests passing: `go test ./...`
  - [ ] Linting clean: `golangci-lint run ./...`
- **Files**: None (verification task)
- **Evidence**: `test-output/phase1/verification.log`

---

### Phase 2: Compose Validation Enhancement (3h)

**Phase Objective**: Add docker compose config validation to pre-commit hooks

#### Task 2.1: Add Compose Schema Validation to Pre-Commit
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Update .pre-commit-config.yaml to add `docker compose config --quiet` validation
- **Acceptance Criteria**:
  - [ ] Updated: `.pre-commit-config.yaml` with new validation step
  - [ ] Hook runs: `docker compose config --quiet` on all compose files
  - [ ] Targets: `deployments/**/compose*.yml`
  - [ ] Stage: pre-commit (same as existing lint-compose)
  - [ ] Verification: `cat .pre-commit-config.yaml | grep "docker compose config"`
- **Files**:
  - `.pre-commit-config.yaml` (modified)
- **Evidence**: `test-output/phase2/precommit-config-diff.txt`

#### Task 2.2: Test Compose Validation with All Files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.1
- **Description**: Run pre-commit on all compose files to verify validation works
- **Acceptance Criteria**:
  - [ ] Command passes: `pre-commit run --all-files lint-compose`
  - [ ] Command passes: `pre-commit run --all-files docker-compose-config` (new hook)
  - [ ] All 24 compose files validated
  - [ ] No errors reported (all files valid)
  - [ ] Log output: `test-output/phase2/precommit-validation.log`
- **Files**: None (testing task)
- **Evidence**: `test-output/phase2/precommit-validation.log`

#### Task 2.3: Verify VS Code Errors Caught
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.2
- **Description**: Verify new validation catches schema errors user manually fixed
- **Acceptance Criteria**:
  - [ ] Review: VS Code errors user fixed in shared-postgres/compose.yml
  - [ ] Test: Reintroduce error, verify pre-commit catches it
  - [ ] Test: Fix error, verify pre-commit passes
  - [ ] Document: Error types caught in `test-output/phase2/error-types-caught.md`
- **Files**: None (verification task)
- **Evidence**: `test-output/phase2/error-types-caught.md`

---

### Phase 3: CICD Foundation (10h)

**Phase Objective**: Generate JSON listing files and implement structural mirror validation

#### Task 3.1: Generate JSON Listing Files with Metadata
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 4h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Create tool to generate JSON listings of deployments/ and configs/ with type/status metadata
- **Acceptance Criteria**:
  - [ ] File created: `internal/cmd/cicd/lint_deployments/generate_listings.go`
  - [ ] Function: `GenerateDeploymentsListing() ([]byte, error)` returns JSON
  - [ ] Function: `GenerateConfigsListing() ([]byte, error)` returns JSON
  - [ ] JSON format: `{"path/to/file": {"type": "compose|config|secret|docker", "status": "required|optional"}}`
  - [ ] Generated: `deployments/deployments_all_files.json`
  - [ ] Generated: `configs/configs_all_files.json`
  - [ ] Tests: `generate_listings_test.go` with ≥98% coverage
  - [ ] Command: `go run ./internal/cmd/cicd/lint_deployments generate-listings`
  - [ ] Verification: `cat deployments/deployments_all_files.json | jq . | head`
- **Files**:
  - `internal/cmd/cicd/lint_deployments/generate_listings.go`
  - `internal/cmd/cicd/lint_deployments/generate_listings_test.go`
  - `deployments/deployments_all_files.json` (generated)
  - `configs/configs_all_files.json` (generated)
- **Evidence**: `test-output/phase3/listings-generation.log`

#### Task 3.2: Implement ValidateStructuralMirror
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 4h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.1
- **Description**: Implement one-way validation (deployments → configs mirror required)
- **Acceptance Criteria**:
  - [ ] Function: `ValidateStructuralMirror(deploymentsJSON, configsJSON []byte) (*ValidationResult, error)`
  - [ ] Validation: Every deployments/ dir MUST have configs/ counterpart (quizme-v2 Q2:C)
  - [ ] Allowed: configs/ CAN have extras (orphans) - report as warnings
  - [ ] Excluded: Infrastructure deployments may be excluded (shared-postgres, etc)
  - [ ] Excluded: Template deployment may be excluded
  - [ ] Output: ValidationResult with errors (missing mirrors) and warnings (orphans)
  - [ ] Tests: `validate_mirror_test.go` with ≥98% coverage
  - [ ] Test case: Missing configs/ dir → error
  - [ ] Test case: Orphaned config → warning (not error)
  - [ ] Command: `go run ./internal/cmd/cicd/lint_deployments validate-mirror`
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_mirror.go`
  - `internal/cmd/cicd/lint_deployments/validate_mirror_test.go`
- **Evidence**: `test-output/phase3/mirror-validation-tests.log`

#### Task 3.3: Write Comprehensive Tests for Phase 3
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 3.1, 3.2
- **Description**: Integration tests for listing generation + mirror validation
- **Acceptance Criteria**:
  - [ ] Integration test: Generate listings → Validate mirror → Verify correctness
  - [ ] Test case: Valid mirror (all deployments have configs) → no errors
  - [ ] Test case: Missing configs/ dir → error reported
  - [ ] Test case: Orphaned config → warning (not error)
  - [ ] Test case: JSON parsing errors handled gracefully
  - [ ] Coverage: ≥98% for both generate_listings.go and validate_mirror.go
  - [ ] Command: `go test -cover ./internal/cmd/cicd/lint_deployments/... -run TestGenerate -run TestValidate`
  - [ ] Mutation: ≥98% gremlins score (run: `gremlins unleash --tags=!integration`)
- **Files**:
  - `internal/cmd/cicd/lint_deployments/integration_test.go` (expanded)
- **Evidence**: `test-output/phase3/coverage.html`, `test-output/phase3/mutation-report.txt`

---

### Phase 4: CICD Comprehensive Refactoring (28h)

**Phase Objective**: Rigorous validation for compose and config files (quizme-v2 Q3:C, Q4:C)

#### Task 4.0: Define Config File Schema
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Document config file schema for validation (prerequisite for Q4:C)
- **Acceptance Criteria**:
  - [ ] File created: `docs/CONFIG-SCHEMA.md` OR update `docs/ARCHITECTURE.md` Section 12.5
  - [ ] Schema documents: server settings (bind addresses, ports)
  - [ ] Schema documents: database settings (URL, credentials via secrets)
  - [ ] Schema documents: telemetry settings (OTLP endpoints)
  - [ ] Schema documents: security settings (TLS, secrets references)
  - [ ] Examples provided: `PRODUCT-SERVICE-app-common.yml` annotated
  - [ ] Validation rules: bind address format, port ranges, secret references
  - [ ] Command: `cat docs/CONFIG-SCHEMA.md` OR `cat docs/ARCHITECTURE.md | grep "Section 12.5"`
- **Files**:
  - `docs/CONFIG-SCHEMA.md` (new) OR `docs/ARCHITECTURE.md` (modified)
- **Evidence**: `test-output/phase4/schema-definition.md`

#### Task 4.1: Implement Comprehensive ValidateComposeFiles (7 validation types)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 10h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Full comprehensive compose validation (quizme-v2 Q3:C "rigourous!!!!")
- **Acceptance Criteria**:
  - [ ] Function: `ValidateComposeFiles(composePath string) (*ValidationResult, error)`
  - [ ] Validation 1: Schema validation (`docker compose config --quiet`)
  - [ ] Validation 2: Port conflict detection (overlapping host ports)
  - [ ] Validation 3: Health check presence (ALL services MUST have health checks)
  - [ ] Validation 4: Service dependency chains (depends_on references valid)
  - [ ] Validation 5: Secret reference validation (all secrets defined in secrets section)
  - [ ] Validation 6: No hardcoded credentials (environment vars checked)
  - [ ] Validation 7: Bind mount security (NO /run/docker.sock mounts)
  - [ ] Output: ValidationResult with errors for each violation type
  - [ ] Tests: `validate_compose_test.go` with ≥98% coverage
  - [ ] Command: `go run ./internal/cmd/cicd/lint_deployments validate-compose deployments/sm-kms/compose.yml`
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_compose.go`
  - `internal/cmd/cicd/lint_deployments/validate_compose_test.go`
- **Evidence**: `test-output/phase4/compose-validation-implementation.log`

#### Task 4.2: Write Comprehensive Tests for Compose Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.1
- **Description**: Test all 7 compose validation types
- **Acceptance Criteria**:
  - [ ] Test case: Schema validation catches invalid YAML
  - [ ] Test case: Port conflicts detected (two services same host port)
  - [ ] Test case: Missing health checks flagged
  - [ ] Test case: Invalid depends_on references caught
  - [ ] Test case: Undefined secrets flagged
  - [ ] Test case: Hardcoded passwords in env detected
  - [ ] Test case: /run/docker.sock mount flagged
  - [ ] Coverage: ≥98% for validate_compose.go
  - [ ] Mutation: ≥98% gremlins score
  - [ ] Command: `go test -cover ./internal/cmd/cicd/lint_deployments/ -run TestValidateCompose`
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_compose_test.go` (expanded)
- **Evidence**: `test-output/phase4/compose-validation-tests.log`

#### Task 4.3: Implement Comprehensive ValidateConfigFiles (5 validation types)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 8h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.0
- **Description**: Full comprehensive config validation (quizme-v2 Q4:C "rigourous!!!!")
- **Acceptance Criteria**:
  - [ ] Function: `ValidateConfigFiles(configPath string, composeServicesJSON []byte) (*ValidationResult, error)`
  - [ ] Validation 1: YAML syntax (parse and validate well-formed)
  - [ ] Validation 2: Format validation (bind addresses IPv4/IPv6, port 1-65535, database URL structure)
  - [ ] Validation 3: Cross-reference (service names match compose.yml services)
  - [ ] Validation 4: Policy enforcement (admin bind 127.0.0.1, public 0.0.0.0 in containers)
  - [ ] Validation 5: Secret references (database passwords via secrets, not inline)
  - [ ] Uses schema from Task 4.0 for validation rules
  - [ ] Output: ValidationResult with errors for each violation type
  - [ ] Tests: `validate_config_test.go` with ≥98% coverage
  - [ ] Command: `go run ./internal/cmd/cicd/lint_deployments validate-config configs/sm-kms/sm-kms-app-common.yml`
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_config.go`
  - `internal/cmd/cicd/lint_deployments/validate_config_test.go`
- **Evidence**: `test-output/phase4/config-validation-implementation.log`

#### Task 4.4: Write Comprehensive Tests for Config Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 4h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.3
- **Description**: Test all 5 config validation types
- **Acceptance Criteria**:
  - [ ] Test case: YAML syntax errors caught
  - [ ] Test case: Invalid bind addresses flagged (127.0.0.l, nonsense)
  - [ ] Test case: Port out of range caught (0, 70000)
  - [ ] Test case: Database URL format errors detected
  - [ ] Test case: Service name mismatch with compose flagged
  - [ ] Test case: Admin bind policy (not 127.0.0.1) violation
  - [ ] Test case: Inline passwords detected (not secret references)
  - [ ] Coverage: ≥98% for validate_config.go
  - [ ] Mutation: ≥98% gremlins score
  - [ ] Command: `go test -cover ./internal/cmd/cicd/lint_deployments/ -run TestValidateConfig`
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_config_test.go` (expanded)
- **Evidence**: `test-output/phase4/config-validation-tests.log`

#### Task 4.5: Integration Tests for All CICD Validations
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 4.1, 4.2, 4.3, 4.4
- **Description**: End-to-end integration tests for entire CICD validation pipeline
- **Acceptance Criteria**:
  - [ ] Test: Generate listings → Validate mirror → Validate compose → Validate config (full pipeline)
  - [ ] Test: Real deployments/ and configs/ directories (not mocks)
  - [ ] Test: All 24 compose files validated (no errors on valid files)
  - [ ] Test: All 55+ config files validated (errors on orphans handled per Q5:C)
  - [ ] Coverage: ≥98% for all CICD code combined
  - [ ] Command: `go test -cover ./internal/cmd/cicd/lint_deployments/... -run TestIntegration`
  - [ ] Performance: Pipeline completes in <60s for all files
- **Files**:
  - `internal/cmd/cicd/lint_deployments/integration_test.go` (expanded)
- **Evidence**: `test-output/phase4/integration-pipeline-tests.log`

---

### Phase 5: Config Directory Restructuring (6h)

**Phase Objective**: Mirror configs/ to match deployments/ structure, handle orphans (quizme-v2 Q5:C)

#### Task 5.1: Audit Current configs/ Structure
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Document current state before restructuring
- **Acceptance Criteria**:
  - [ ] Generated: `test-output/phase5/current-structure.txt` (tree of configs/)
  - [ ] Generated: `test-output/phase5/file-inventory.txt` (all 55 files listed)
  - [ ] Generated: `test-output/phase5/structure-mapping.md` (current vs target structure)
  - [ ] Command: `tree configs/ > test-output/phase5/current-structure.txt`
  - [ ] Command: `find configs/ -type f > test-output/phase5/file-inventory.txt`
- **Files**: None (audit task)
- **Evidence**: `test-output/phase5/current-structure.txt`, `test-output/phase5/file-inventory.txt`

#### Task 5.2: Identify Orphaned Config Files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.1, Phase 3 (ValidateStructuralMirror)
- **Description**: Identify configs/ files without corresponding deployments/ directories
- **Acceptance Criteria**:
  - [ ] Run: ValidateStructuralMirror to get warnings (orphans)
  - [ ] Generated: `test-output/phase5/orphans-list.txt` (all orphaned files)
  - [ ] Count: Number of orphans documented
  - [ ] Analysis: Why orphans exist (old services, renamed dirs, etc)
  - [ ] Command: `go run ./internal/cmd/cicd/lint_deployments validate-mirror > test-output/phase5/orphans-list.txt 2>&1`
- **Files**: None (analysis task)
- **Evidence**: `test-output/phase5/orphans-list.txt`, `test-output/phase5/orphan-analysis.md`

#### Task 5.3: Restructure configs/ and Handle Orphans
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 5.1, 5.2
- **Description**: Create exact mirror structure, move orphans to configs/orphaned/ (quizme-v2 Q5:C)
- **Acceptance Criteria**:
  - [ ] Created: `configs/orphaned/` directory
  - [ ] Moved: All orphaned configs to configs/orphaned/
  - [ ] Created: `configs/orphaned/README.md` explaining archive
  - [ ] Restructured: All valid configs match deployments/ structure
  - [ ] Example: `configs/sm-kms/config/sm-kms-app-common.yml` mirrors `deployments/sm-kms/config/`
  - [ ] Log: `test-output/phase5/orphaned-configs.txt` lists all moved files
  - [ ] Verification: ValidateStructuralMirror passes (no errors, orphans archived)
  - [ ] Command: `go run ./internal/cmd/cicd/lint_deployments validate-mirror` (should pass)
- **Files**:
  - `configs/orphaned/README.md` (created)
  - `configs/**/` (restructured to mirror deployments)
  - All orphaned files moved to `configs/orphaned/`
- **Evidence**: `test-output/phase5/orphaned-configs.txt`, `test-output/phase5/restructure.log`

#### Task 5.4: Validate Mirror Correctness
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.3
- **Description**: Verify exact mirror achieved, no validation errors
- **Acceptance Criteria**:
  - [ ] Run: ValidateStructuralMirror (no errors)
  - [ ] Verification: All deployments/ dirs have configs/ counterparts
  - [ ] Verification: Orphans archived in configs/orphaned/ (warnings acceptable)
  - [ ] Manual spot-check: 3 random services have matching structures
  - [ ] Command: `go run ./internal/cmd/cicd/lint_deployments validate-mirror`
  - [ ] Command: `diff -r <(tree deployments/sm-kms/config/) <(tree configs/sm-kms/config/)` (similar structure)
- **Files**: None (verification task)
- **Evidence**: `test-output/phase5/mirror-validation.log`

---

### Phase 6: Documentation & Integration (3h)

**Phase Objective**: Update documentation and integrate validations into CI/CD

#### Task 6.1: Update ARCHITECTURE.md with Config Schema
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.0
- **Description**: Add Section 12.5 Config Schema OR reference CONFIG-SCHEMA.md
- **Acceptance Criteria**:
  - [ ] Updated: `docs/ARCHITECTURE.md` Section 12.5 (or reference to CONFIG-SCHEMA.md)
  - [ ] Schema documented: Server, database, telemetry, security settings
  - [ ] Examples provided: Annotated config files
  - [ ] Cross-references: Links to validation code in lint_deployments/
  - [ ] Command: `cat docs/ARCHITECTURE.md | grep -A20 "Section 12.5"`
- **Files**:
  - `docs/ARCHITECTURE.md` (modified)
- **Evidence**: `git diff docs/ARCHITECTURE.md`

#### Task 6.2: Update Copilot Instructions for New Validations
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Update 04-01.deployment.instructions.md with new validation patterns
- **Acceptance Criteria**:
  - [ ] Updated: `.github/instructions/04-01.deployment.instructions.md`
  - [ ] Documented: JSON listing file format and usage
  - [ ] Documented: ValidateStructuralMirror usage
  - [ ] Documented: Comprehensive compose/config validation commands
  - [ ] Documented: Orphaned config handling pattern
  - [ ] Command: `cat .github/instructions/04-01.deployment.instructions.md | grep ValidateStructuralMirror`
- **Files**:
  - `.github/instructions/04-01.deployment.instructions.md` (modified)
- **Evidence**: `git diff .github/instructions/04-01.deployment.instructions.md`

#### Task 6.3: Integrate Validations into CI/CD Workflows
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 4 complete
- **Description**: Add CICD validations to GitHub Actions workflows
- **Acceptance Criteria**:
  - [ ] Updated: `.github/workflows/ci-quality.yml` (or create new workflow)
  - [ ] Step: Generate listings (deployments_all_files.json, configs_all_files.json)
  - [ ] Step: Validate structural mirror
  - [ ] Step: Validate all compose files (24 files)
  - [ ] Step: Validate all config files (55+ files)
  - [ ] Fail workflow: If any validation errors found
  - [ ] Artifacts: Upload ValidationResult JSON on failure
  - [ ] Command: `cat .github/workflows/ci-quality.yml | grep validate`
- **Files**:
  - `.github/workflows/ci-quality.yml` (modified or new)
- **Evidence**: `git diff .github/workflows/ci-quality.yml`

#### Task 6.4: Update Pre-Commit Hook Configuration
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 2, Phase 4
- **Description**: Ensure pre-commit runs all validations
- **Acceptance Criteria**:
  - [ ] Hook: lint-compose (existing, enhanced in Phase 2)
  - [ ] Hook: validate-mirror (new, runs ValidateStructuralMirror)
  - [ ] Hook: validate-compose-files (new, comprehensive validation)
  - [ ] Hook: validate-config-files (new, comprehensive validation)
  - [ ] Performance: All hooks complete in <90s
  - [ ] Command: `pre-commit run --all-files`
- **Files**:
  - `.pre-commit-config.yaml` (modified)
- **Evidence**: `test-output/phase6/precommit-all-hooks.log`

---

### Phase 7: Comprehensive Testing & Quality Gates (5h)

**Phase Objective**: End-to-end verification, mutation testing, evidence collection

#### Task 7.1: E2E Tests for All CICD Commands
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: All previous phases
- **Description**: End-to-end black-box tests of CICD tooling
- **Acceptance Criteria**:
  - [ ] Test: `cicd generate-listings` → JSON files created
  - [ ] Test: `cicd validate-mirror` → No errors on valid structure
  - [ ] Test: `cicd validate-compose` → All 24 compose files pass
  - [ ] Test: `cicd validate-config` → All valid configs pass
  - [ ] Test: Error injection (invalid compose) → Errors caught
  - [ ] Test: Error injection (invalid config) → Errors caught
  - [ ] Coverage: ≥98% for all CICD main.go and command handlers
  - [ ] Command: `go test -tags=e2e ./internal/cmd/cicd/lint_deployments/... -run TestE2E`
- **Files**:
  - `internal/cmd/cicd/lint_deployments/e2e_test.go` (new or expanded)
- **Evidence**: `test-output/phase7/e2e-tests.log`

#### Task 7.2: Mutation Testing for CICD Code
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.1
- **Description**: Run gremlins mutation testing on all CICD code
- **Acceptance Criteria**:
  - [ ] Command: `gremlins unleash --tags=!integration ./internal/cmd/cicd/lint_deployments/...`
  - [ ] Target: ≥98% mutation score (infrastructure/utility category)
  - [ ] Report: `test-output/phase7/mutation-report.txt`
  - [ ] Action: Fix any surviving mutants (add tests or fix logic)
  - [ ] Verification: Re-run gremlins until ≥98% achieved
- **Files**: None (testing task)
- **Evidence**: `test-output/phase7/mutation-report.txt`

#### Task 7.3: Pre-Commit Hook Verification
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.4
- **Description**: Verify all pre-commit hooks work correctly
- **Acceptance Criteria**:
  - [ ] Test: Run all hooks on clean codebase → Pass
  - [ ] Test: Inject compose error → Hook catches it
  - [ ] Test: Inject config error → Hook catches it
  - [ ] Test: Remove configs/ dir → Mirror validation fails
  - [ ] Performance: All hooks complete in <90s
  - [ ] Log: `test-output/phase7/precommit-verification.log`
  - [ ] Command: `pre-commit run --all-files > test-output/phase7/precommit-verification.log 2>&1`
- **Files**: None (verification task)
- **Evidence**: `test-output/phase7/precommit-verification.log`

#### Task 7.4: Coverage Report Generation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: All tests complete
- **Description**: Generate comprehensive coverage report for CICD code
- **Acceptance Criteria**:
  - [ ] Command: `go test -coverprofile=test-output/phase7/coverage.out ./internal/cmd/cicd/lint_deployments/...`
  - [ ] Command: `go tool cover -html=test-output/phase7/coverage.out -o test-output/phase7/coverage.html`
  - [ ] Target: ≥98% line coverage (infrastructure/utility)
  - [ ] Review: Coverage HTML for any RED lines
  - [ ] Action: Add tests for any uncovered lines
- **Files**: None (report generation)
- **Evidence**: `test-output/phase7/coverage.html`, `test-output/phase7/coverage.out`

#### Task 7.5: Final Smoke Test: Full Pipeline
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: All previous tasks
- **Description**: Run entire validation pipeline end-to-end as final smoke test
- **Acceptance Criteria**:
  - [ ] Step 1: Generate listings → `cicd generate-listings`
  - [ ] Step 2: Validate mirror → `cicd validate-mirror` (pass)
  - [ ] Step 3: Validate all 24 compose files → `cicd validate-compose deployments/**/compose*.yml` (pass)
  - [ ] Step 4: Validate all valid config files → `cicd validate-config configs/**/*.yml` (pass)
  - [ ] Step 5: Run pre-commit hooks → `pre-commit run --all-files` (pass)
  - [ ] Step 6: Run full test suite → `go test ./...` (100% pass)
  - [ ] Step 7: Run linting → `golangci-lint run ./...` (zero errors)
  - [ ] Log: `test-output/phase7/final-smoke-test.log`
  - [ ] Result: ALL PASS, ready for production
- **Files**: None (smoke test)
- **Evidence**: `test-output/phase7/final-smoke-test.log`

#### Task 7.6: Evidence Archive and Documentation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: All previous tasks
- **Description**: Archive all evidence and create COMPLETION.md summary
- **Acceptance Criteria**:
  - [ ] Archive: All test-output/ subdirectories organized
  - [ ] Created: `test-output/COMPLETION.md` (summary of all evidence)
  - [ ] Summary: All 7 phases complete, all quality gates passed
  - [ ] Metrics: Final coverage (≥98%), mutation score (≥98%), test count
  - [ ] Artifacts: Links to all evidence files
  - [ ] Command: `tree test-output/ > test-output/evidence-tree.txt`
- **Files**:
  - `test-output/COMPLETION.md` (created)
  - `test-output/evidence-tree.txt` (created)
- **Evidence**: N/A (final evidence collection task)

---

## Cross-Cutting Tasks

### Testing
- [ ] Unit tests ≥98% coverage (all CICD code)
- [ ] Integration tests pass (full pipeline)
- [ ] E2E tests pass (all commands)
- [ ] Mutation testing ≥98% (infrastructure/utility category)
- [ ] No skipped tests (except documented exceptions)
- [ ] Race detector clean: `go test -race ./...`

### Code Quality
- [ ] Linting passes: `golangci-lint run ./...`
- [ ] No new TODOs without tracking
- [ ] No security vulnerabilities: `gosec ./...`
- [ ] Formatting clean: `gofumpt -s -w ./`
- [ ] Imports organized: `goimports -w ./`

### Documentation
- [ ] ARCHITECTURE.md updated with config schema
- [ ] CONFIG-SCHEMA.md created (if separate from ARCHITECTURE.md)
- [ ] Instruction files updated (04-01.deployment.instructions.md)
- [ ] README.md updated with CICD validation commands
- [ ] Comments added for complex validation logic

### Deployment
- [ ] Pre-commit hooks tested and working
- [ ] CI/CD workflow integration complete
- [ ] Docker compose schema validation integrated
- [ ] All validation commands documented
- [ ] Evidence archived in test-output/

---

## Evidence Archive

- `test-output/quizme-v2-analysis/` - Quizme-v2 answers and impact analysis
- `test-output/phase0.5/` - Demo archiving evidence
- `test-output/phase1/` - Structural cleanup verification
- `test-output/phase2/` - Pre-commit enhancement tests
- `test-output/phase3/` - Listings generation + mirror validation
- `test-output/phase4/` - Comprehensive validation implementation
- `test-output/phase5/` - Config restructuring + orphan handling
- `test-output/phase6/` - Documentation updates
- `test-output/phase7/` - Final testing + quality gates
- `test-output/COMPLETION.md` - Final evidence summary
