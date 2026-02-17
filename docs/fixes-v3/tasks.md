# Tasks - Configs/Deployments/CICD Rigor & Consistency v3

**Status**: 0 of 56 tasks complete (0%)
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
- ✅ **Document root causes** - Root cause analysis is part of planning AND implementation, not optional; planning blockers must be resolved during planning, implementation blockers MUST be resolved during implementation
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task complete with known issues
- ✅ **NEVER de-prioritize quality** - Evidence-based verification is ALWAYS highest priority

**Rationale**: Maintaining maximum quality prevents cascading failures and rework.

---

## Task Checklist

### Phase 1: configs/ Directory Restructuring (12h)

**Phase Objective**: Achieve 100% naming consistency in configs/ to match deployments/*/config/ patterns

#### Task 1.1: Create Rename Scripts
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Automated scripts for rename + rollback
- **Acceptance Criteria**:
  - [ ] Script created: `scripts/rename-configs-v3.sh` (batch rename)
  - [ ] Script created: `scripts/rollback-configs-v3.sh` (undo rename)
  - [ ] Dry-run mode: `--dry-run` shows changes without executing
  - [ ] Validation: Scripts check for conflicts before rename
  - [ ] Tests: Scripts validated on test directory structure
- **Files**:
  - `scripts/rename-configs-v3.sh` (new)
  - `scripts/rollback-configs-v3.sh` (new)
  - `scripts/README.md` (update with script docs)
- **Evidence**:
  - `test-output/phase1/rename-dry-run.txt` - Dry-run output
  - `test-output/phase1/rename-validation.log` - Script validation

#### Task 1.2: Rename ca/ → pki-ca/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Rename directory to match deployments/pki-ca/
- **Acceptance Criteria**:
  - [ ] Directory renamed: `configs/ca/` → `configs/pki-ca/`
  - [ ] All files updated: ca-server.yml → pki-ca-app-common.yml
  - [ ] Profiles directory preserved: `configs/pki-ca/profiles/` (26 files)
  - [ ] No broken imports: `grep -r "configs/ca/" returns 0 matches`
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/pki-ca/` (renamed from ca/)
  - `configs/pki-ca/pki-ca-app-common.yml` (renamed from ca-server.yml)
- **Evidence**:
  - `test-output/phase1/task-1.2-rename-log.txt` - Rename execution log

#### Task 1.3: Restructure identity/ → identity-{authz,idp,rp,rs,spa}/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Split mixed identity/ into 5 service subdirs
- **Acceptance Criteria**:
  - [ ] Created: `configs/identity-authz/` with identity-authz-app-common.yml
  - [ ] Created: `configs/identity-idp/` with identity-idp-app-common.yml
  - [ ] Created: `configs/identity-rp/` with identity-rp-app-*.yml files
  - [ ] Created: `configs/identity-rs/` with identity-rs-app-common.yml
  - [ ] Created: `configs/identity-spa/` with identity-spa-app-*.yml files
  - [ ] Preserved: `configs/identity/policies/` (3 files - shared)
  - [ ] Preserved: `configs/identity/profiles/` (5 files - shared)
  - [ ] Handled: development.yml, production.yml, test.yml (document decision in task evidence)
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/identity-authz/identity-authz-app-common.yml` (renamed from authz.yml)
  - `configs/identity-authz/identity-authz-app-docker.yml` (renamed from authz-docker.yml)
  - `configs/identity-idp/identity-idp-app-common.yml` (renamed from idp.yml)
  - `configs/identity-idp/identity-idp-app-docker.yml` (renamed from idp-docker.yml)
  - `configs/identity-rs/identity-rs-app-common.yml` (renamed from rs.yml)
  - `configs/identity-rs/identity-rs-app-docker.yml` (renamed from rs-docker.yml)
  - `configs/identity-rp/identity-rp-app-common.yml` (new)
  - `configs/identity-spa/identity-spa-app-common.yml` (new)
  - `configs/identity/policies/` (preserved)
  - `configs/identity/profiles/` (preserved)
- **Evidence**:
  - `test-output/phase1/task-1.3-restructure-plan.md` - Restructure decisions
  - `test-output/phase1/task-1.3-env-files-decision.md` - What to do with development.yml, production.yml, test.yml

#### Task 1.4: Create configs/sm-kms/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Create missing sm-kms subdirectory with config files
- **Acceptance Criteria**:
  - [ ] Created: `configs/sm-kms/` directory
  - [ ] Created: `configs/sm-kms/sm-kms-app-common.yml`
  - [ ] Created: `configs/sm-kms/sm-kms-app-sqlite-1.yml`
  - [ ] Created: `configs/sm-kms/sm-kms-app-postgresql-1.yml`
  - [ ] Created: `configs/sm-kms/sm-kms-app-postgresql-2.yml`
  - [ ] Configs mirror deployments/sm-kms/config/ structure
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/sm-kms/sm-kms-app-common.yml` (new)
  - `configs/sm-kms/sm-kms-app-sqlite-1.yml` (new)
  - `configs/sm-kms/sm-kms-app-postgresql-1.yml` (new)
  - `configs/sm-kms/sm-kms-app-postgresql-2.yml` (new)
- **Evidence**:
  - `test-output/phase1/task-1.4-sm-kms-configs.txt` - Created files

#### Task 1.5: Rename cipher/im/ configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Rename cipher-im configs to new pattern
- **Acceptance Criteria**:
  - [ ] Renamed: config.yml → cipher-im-app-common.yml
  - [ ] Renamed: config-pg-1.yml → cipher-im-app-postgresql-1.yml
  - [ ] Renamed: config-pg-2.yml → cipher-im-app-postgresql-2.yml
  - [ ] Renamed: config-sqlite.yml → cipher-im-app-sqlite-1.yml
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/cipher/im/cipher-im-app-common.yml` (renamed)
  - `configs/cipher/im/cipher-im-app-postgresql-1.yml` (renamed)
  - `configs/cipher/im/cipher-im-app-postgresql-2.yml` (renamed)
  - `configs/cipher/im/cipher-im-app-sqlite-1.yml` (renamed)

#### Task 1.6: Rename jose/ config
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Rename jose-server.yml to new pattern
- **Acceptance Criteria**:
  - [ ] Renamed: jose-server.yml → jose-ja-app-common.yml
  - [ ] Created: jose-ja-app-sqlite-1.yml, jose-ja-app-postgresql-1.yml, jose-ja-app-postgresql-2.yml
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/jose/jose-ja-app-common.yml` (renamed from jose-server.yml)
  - `configs/jose/jose-ja-app-sqlite-1.yml` (new)
  - `configs/jose/jose-ja-app-postgresql-1.yml` (new)
  - `configs/jose/jose-ja-app-postgresql-2.yml` (new)

#### Task 1.7: Rename cryptoutil/ config
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Rename cryptoutil config to new pattern
- **Acceptance Criteria**:
  - [ ] Renamed: config.yml → cryptoutil-app-common.yml
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/cryptoutil/cryptoutil-app-common.yml` (renamed from config.yml)

#### Task 1.8: Rename cipher/ PRODUCT-level config
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Rename cipher PRODUCT-level config
- **Acceptance Criteria**:
  - [ ] Renamed: config.yml → cipher-app-common.yml
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/cipher/cipher-app-common.yml` (renamed from config.yml)

#### Task 1.9: Validate All Renames
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 1.2-1.8
- **Description**: Comprehensive validation after all renames
- **Acceptance Criteria**:
  - [ ] 0 files with old naming: `find configs/ -name "config.yml" -o -name "config-pg-*.yml" -o -name "config-sqlite.yml" | wc -l` = 0
  - [ ] All files match pattern: `find configs/ -name "*-app-*.yml" | wc -l` matches expected count
  - [ ] No broken imports: `go build ./...` clean
  - [ ] All tests pass: `go test ./...`
  - [ ] Linting clean: `golangci-lint run ./...`
  - [ ] CICD validates: `go run ./cmd/cicd lint-deployments validate-mirror`
- **Files**: None (validation task)
- **Evidence**:
  - `test-output/phase1/validation-results.txt` - Comprehensive validation
  - `test-output/phase1/file-count-comparison.txt` - Before/after file counts

#### Task 1.10: Create configs/ README.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.9
- **Description**: Document configs/ structure and purpose
- **Acceptance Criteria**:
  - [ ] Created: `configs/README.md`
  - [ ] Explains: Purpose (reference/templates vs deployments/ runtime)
  - [ ] Documents: Naming pattern (PRODUCT-SERVICE-app-{common,sqlite-1,postgresql-1,postgresql-2}.yml)
  - [ ] Lists: All subdirectories and their purpose
  - [ ] References: ARCHITECTURE.md, CONFIG-SCHEMA.md
- **Files**:
  - `configs/README.md` (new)

---

### Phase 2: PRODUCT/SUITE Config Creation (6h)

**Phase Objective**: Add missing PRODUCT and SUITE-level configurations

#### Task 2.1: Create cipher/ PRODUCT configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 complete
- **Description**: Add PRODUCT-level configs for cipher
- **Acceptance Criteria**:
  - [ ] Created: `configs/cipher/cipher-app-common.yml`
  - [ ] Created: `configs/cipher/cipher-app-sqlite-1.yml`
  - [ ] Created: `configs/cipher/cipher-app-postgresql-1.yml`
  - [ ] Created: `configs/cipher/cipher-app-postgresql-2.yml`
  - [ ] Created: `configs/cipher/README.md` (explains PRODUCT delegation)
- **Files**:
  - `configs/cipher/cipher-app-common.yml` (already renamed in Task 1.8)
  - `configs/cipher/cipher-app-sqlite-1.yml` (new)
  - `configs/cipher/cipher-app-postgresql-1.yml` (new)
  - `configs/cipher/cipher-app-postgresql-2.yml` (new)
  - `configs/cipher/README.md` (new)

#### Task 2.2: Create jose/ PRODUCT configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h  
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 complete
- **Description**: Add PRODUCT-level configs for jose
- **Acceptance Criteria**:
  - [ ] Created: `configs/jose/jose-app-common.yml` (PRODUCT-level, delegates to jose-ja)
  - [ ] Created: `configs/jose/jose-app-sqlite-1.yml`
  - [ ] Created: `configs/jose/jose-app-postgresql-1.yml`
  - [ ] Created: `configs/jose/jose-app-postgresql-2.yml`
  - [ ] Created: `configs/jose/README.md`
  - [ ] Preserved: `configs/jose/jose-ja-app-*.yml` (SERVICE-level from Task 1.6)
- **Files**:
  - `configs/jose/jose-app-common.yml` (new, PRODUCT-level)
  - `configs/jose/jose-app-sqlite-1.yml` (new)
  - `configs/jose/jose-app-postgresql-1.yml` (new)
  - `configs/jose/jose-app-postgresql-2.yml` (new)
  - `configs/jose/README.md` (new)

#### Task 2.3: Create identity/ PRODUCT configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 complete
- **Description**: Add PRODUCT-level configs for identity
- **Acceptance Criteria**:
  - [ ] Created: `configs/identity/identity-app-common.yml` (PRODUCT-level, delegates to 5 services)
  - [ ] Created: `configs/identity/identity-app-sqlite-1.yml`
  - [ ] Created: `configs/identity/identity-app-postgresql-1.yml`
  - [ ] Created: `configs/identity/identity-app-postgresql-2.yml`
  - [ ] Created: `configs/identity/README.md`
  - [ ] Preserved: `configs/identity/policies/`, `configs/identity/profiles/`
  - [ ] Preserved: `configs/identity-{authz,idp,rp,rs,spa}/` (from Task 1.3)
- **Files**:
  - `configs/identity/identity-app-common.yml` (new, PRODUCT-level)
  - `configs/identity/identity-app-sqlite-1.yml` (new)
  - `configs/identity/identity-app-postgresql-1.yml` (new)
  - `configs/identity/identity-app-postgresql-2.yml` (new)
  - `configs/identity/README.md` (new)

#### Task 2.4: Create sm/ PRODUCT configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 complete
- **Description**: Add PRODUCT-level configs for sm
- **Acceptance Criteria**:
  - [ ] Created: `configs/sm/sm-app-common.yml`
  - [ ] Created: `configs/sm/sm-app-sqlite-1.yml`
  - [ ] Created: `configs/sm/sm-app-postgresql-1.yml`
  - [ ] Created: `configs/sm/sm-app-postgresql-2.yml`
  - [ ] Created: `configs/sm/README.md`
  - [ ] Preserved: `configs/sm-kms/` (from Task 1.4)
- **Files**:
  - `configs/sm/sm-app-common.yml` (new)
  - `configs/sm/sm-app-sqlite-1.yml` (new)
  - `configs/sm/sm-app-postgresql-1.yml` (new)
  - `configs/sm/sm-app-postgresql-2.yml` (new)
  - `configs/sm/README.md` (new)

#### Task 2.5: Create pki/ PRODUCT configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 complete
- **Description**: Add PRODUCT-level configs for pki
- **Acceptance Criteria**:
  - [ ] Created: `configs/pki/pki-app-common.yml`
  - [ ] Created: `configs/pki/pki-app-sqlite-1.yml`
  - [ ] Created: `configs/pki/pki-app-postgresql-1.yml`
  - [ ] Created: `configs/pki/pki-app-postgresql-2.yml`
  - [ ] Created: `configs/pki/README.md`
  - [ ] Preserved: `configs/pki-ca/` (from Task 1.2)
- **Files**:
  - `configs/pki/pki-app-common.yml` (new)
  - `configs/pki/pki-app-sqlite-1.yml` (new)
  - `configs/pki/pki-app-postgresql-1.yml` (new)
  - `configs/pki/pki-app-postgresql-2.yml` (new)
  - `configs/pki/README.md` (new)

#### Task 2.6: Update cryptoutil/ SUITE config
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 2.1-2.5
- **Description**: Enhance SUITE-level config (already renamed in Task 1.7)
- **Acceptance Criteria**:
  - [ ] Updated: `configs/cryptoutil/cryptoutil-app-common.yml` (delegates to all 5 PRODUCTs)
  - [ ] Created: `configs/cryptoutil/cryptoutil-app-sqlite-1.yml`
  - [ ] Created: `configs/cryptoutil/cryptoutil-app-postgresql-1.yml`
  - [ ] Created: `configs/cryptoutil/cryptoutil-app-postgresql-2.yml`
  - [ ] Created: `configs/cryptoutil/README.md`
- **Files**:
  - `configs/cryptoutil/cryptoutil-app-common.yml` (updated)
  - `configs/cryptoutil/cryptoutil-app-sqlite-1.yml` (new)
  - `configs/cryptoutil/cryptoutil-app-postgresql-1.yml` (new)
  - `configs/cryptoutil/cryptoutil-app-postgresql-2.yml` (new)
  - `configs/cryptoutil/README.md` (new)

---

### Phase 3: CICD Linting Enhancement - Config Validation (18h)

**Phase Objective**: Implement 8 missing validation types for config files

#### Task 3.1: Implement Config File Naming Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 2 complete
- **Description**: Validate config file naming pattern
- **Acceptance Criteria**:
  - [ ] Function: `validateConfigFileNaming(configPath string, result *ConfigValidationResult)`
  - [ ] Validates: PRODUCT-SERVICE-app-{common,sqlite-1,postgresql-1,postgresql-2}.yml pattern
  - [ ] Validates: PRODUCT-app-VARIANT.yml pattern (for PRODUCT-level)
  - [ ] Validates: SUITE-app-VARIANT.yml pattern (for SUITE-level)
  - [ ] Errors: Files not matching pattern
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_config_naming.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_config_naming_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.1-naming-validation.log`

#### Task 3.2: Implement Kebab-Case Key Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 2 complete
- **Description**: Validate flat YAML with kebab-case keys only
- **Acceptance Criteria**:
  - [ ] Function: `validateKebabCaseKeys(configPath string, result *ConfigValidationResult)`
  - [ ] Validates: All keys are kebab-case (no camelCase, snake_case, PascalCase)
  - [ ] Validates: Flat YAML structure (no nesting beyond 1 level)
  - [ ] Errors: camelCase keys (bindPublicAddress)
  - [ ] Errors: snake_case keys (bind_public_address)
  - [ ] Errors: Nested YAML (server: {bind: {address: ...}})
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_config_kebab.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_config_kebab_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.2-kebab-validation.log`

#### Task 3.3: Implement Schema Completeness Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 2 complete
- **Description**: Validate required fields per CONFIG-SCHEMA.md
- **Acceptance Criteria**:
  - [ ] Function: `validateSchemaCompleteness(configPath string, configType string, result *ConfigValidationResult)`
  - [ ] Validates: Required fields present (bind-public-address, bind-public-port, bind-private-address, bind-private-port)
  - [ ] Validates: Database fields (database-url OR database-driver + database-dsn)
  - [ ] Validates: Telemetry fields (otlp-endpoint, otlp-protocol)
  - [ ] Errors: Missing required fields
  - [ ] Warnings: Missing optional fields (cors-allowed-origins)
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_config_schema.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_config_schema_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.3-schema-validation.log`

#### Task 3.4: Implement Port Offset Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 2 complete
- **Description**: Validate port offset consistency (SERVICE +0, PRODUCT +10000, SUITE +20000)
- **Acceptance Criteria**:
  - [ ] Function: `validatePortOffsets(configPath string, deploymentType string, result *ConfigValidationResult)`
  - [ ] Validates: SERVICE configs use base ports (8080, 9090)
  - [ ] Validates: PRODUCT configs use +10000 offset (18080, 19090)
  - [ ] Validates: SUITE config uses +20000 offset (28080, 29090)
  - [ ] Errors: Incorrect port offsets
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_config_ports.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_config_ports_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.4-ports-validation.log`

#### Task 3.5: Implement Telemetry Config Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 2 complete
- **Description**: Validate OTLP telemetry configuration
- **Acceptance Criteria**:
  - [ ] Function: `validateTelemetryConfig(configPath string, result *ConfigValidationResult)`
  - [ ] Validates: otlp-protocol is "grpc" or "http"
  - [ ] Validates: otlp-endpoint format (host:port)
  - [ ] Validates: service-name present
  - [ ] Errors: Invalid protocol
  - [ ] Errors: Invalid endpoint format
  - [ ] Warnings: Missing service-name (falls back to binary name)
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_config_telemetry.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_config_telemetry_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.5-telemetry-validation.log`

#### Task 3.6: Implement Admin Policy Enforcement
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 2 complete
- **Description**: Validate admin bind address is 127.0.0.1
- **Acceptance Criteria**:
  - [ ] Function: `validateAdminPolicy(configPath string, result *ConfigValidationResult)`
  - [ ] Validates: bind-private-address MUST be 127.0.0.1 (not 0.0.0.0)
  - [ ] Validates: bind-private-port in range 1-65535
  - [ ] Errors: bind-private-address = 0.0.0.0 (admin MUST NOT be exposed)
  - [ ] Errors: bind-private-port out of range
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_config_admin.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_config_admin_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.6-admin-validation.log`

#### Task 3.7: Implement deployments/*/config/ vs configs/ Consistency Check
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 2 complete
- **Description**: Validate deployments/*/config/ mirrors configs/ structure
- **Acceptance Criteria**:
  - [ ] Function: `validateConfigConsistency(deploymentsRoot string, configsRoot string) (*ConsistencyResult, error)`
  - [ ] Validates: Each deployments/PRODUCT-SERVICE/config/*.yml has corresponding configs/PRODUCT-SERVICE/*.yml
  - [ ] Validates: File count matches
  - [ ] Validates: Naming patterns match
  - [ ] Errors: Missing files in configs/
  - [ ] Warnings: Extra files in configs/ (templates/examples)
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_consistency.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_consistency_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.7-consistency-validation.log`

#### Task 3.8: Implement Secret Reference Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 2 complete
- **Description**: Validate secret file:// references exist and use correct suffixes
- **Acceptance Criteria**:
  - [ ] Function: `validateSecretReferences(configPath string, deploymentType string, result *ConfigValidationResult)`
  - [ ] Validates: file:// paths exist
  - [ ] Validates: Secret suffixes match deployment level (-SERVICEONLY, -PRODUCTONLY, -SUITEONLY, -SHARED)
  - [ ] Errors: file:// path does not exist
  - [ ] Errors: Wrong suffix for deployment type
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_config_secrets.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_config_secrets_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.8-secrets-validation.log`

#### Task 3.9: Integrate All Validations into ValidateConfigFile
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 3.1-3.8
- **Description**: Wire all 8 new validation types into existing ValidateConfigFile
- **Acceptance Criteria**:
  - [ ] Updated: `ValidateConfigFile` calls all 8 new validation functions
  - [ ] Updated: `ConfigValidationResult` includes all new errors/warnings
  - [ ] Updated: `FormatConfigValidationResult` displays all new errors/warnings
  - [ ] Tests: Integration tests cover all 8 validation types
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_config.go` (updated)
  - `internal/cmd/cicd/lint_deployments/validate_config_test.go` (updated)
- **Evidence**:
  - `test-output/phase3/task-3.9-integration-tests.log`

---

### Phase 4: CICD Linting Enhancement - Deployment Validation (12h)

**Phase Objective**: Enhance deployment structure validation for PRODUCT/SUITE levels

#### Task 4.1: Implement PRODUCT Compose Delegation Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 3 complete
- **Description**: Validate PRODUCT compose includes SERVICE composes
- **Acceptance Criteria**:
  - [ ] Function: `validateProductDelegation(productPath string, result *ValidationResult)`
  - [ ] Validates: PRODUCT compose includes all SERVICE composes (e.g., cipher/compose.yml includes ../cipher-im/compose.yml)
  - [ ] Validates: PRODUCT secrets include hash_pepper (shared by services)
  - [ ] Errors: Missing SERVICE includes
  - [ ] Errors: Missing hash_pepper secret
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_product.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_product_test.go` (new)
- **Evidence**:
  - `test-output/phase4/task-4.1-product-validation.log`

#### Task 4.2: Implement SUITE Compose Delegation Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.1
- **Description**: Validate SUITE compose includes PRODUCT composes
- **Acceptance Criteria**:
  - [ ] Function: `validateSuiteDelegation(suitePath string, result *ValidationResult)`
  - [ ] Validates: SUITE compose includes all 5 PRODUCT composes (sm, pki, cipher, jose, identity)
  - [ ] Validates: SUITE secrets include cryptoutil-hash_pepper
  - [ ] Validates: NO direct SERVICE includes (MUST delegate through PRODUCTs)
  - [ ] Errors: Missing PRODUCT includes
  - [ ] Errors: Direct SERVICE includes (violates rigid delegation)
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_suite.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_suite_test.go` (new)
- **Evidence**:
  - `test-output/phase4/task-4.2-suite-validation.log`

#### Task 4.3: Implement README.md Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.2
- **Description**: Validate README.md existence in PRODUCT/SUITE deployments
- **Acceptance Criteria**:
  - [ ] Function: `validateReadmeExists(deploymentPath string, deploymentType string, result *ValidationResult)`
  - [ ] Validates: README.md exists in PRODUCT/SUITE deployments
  - [ ] Warnings: Missing README.md (not blocking, but recommended)
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_readme.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_readme_test.go` (new)
- **Evidence**:
  - `test-output/phase4/task-4.3-readme-validation.log`

#### Task 4.4: Implement Port Offset Consistency in Compose
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.3
- **Description**: Validate port offsets in compose services
- **Acceptance Criteria**:
  - [ ] Function: `validateComposePortOffsets(composePath string, deploymentType string, result *ComposeValidationResult)`
  - [ ] Validates: SERVICE composes use base ports (8080:8080, 9090:9090)
  - [ ] Validates: PRODUCT composes use +10000 offset (18080:8080, 19090:9090)
  - [ ] Validates: SUITE compose uses +20000 offset (28080:8080, 29090:9090)
  - [ ] Errors: Incorrect port offsets
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_compose_ports.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_compose_ports_test.go` (new)
- **Evidence**:
  - `test-output/phase4/task-4.4-compose-ports-validation.log`

#### Task 4.5: Implement Secret Suffix Consistency Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.4
- **Description**: Validate .never files and secret suffixes
- **Acceptance Criteria**:
  - [ ] Function: `validateSecretSuffixes(deploymentPath string, deploymentType string, result *ValidationResult)`
  - [ ] Validates: SERVICE secrets use -SERVICEONLY suffix
  - [ ] Validates: PRODUCT secrets use -PRODUCTONLY or -SHARED suffix
  - [ ] Validates: SUITE secrets use -SUITEONLY or -SHARED suffix
  - [ ] Validates: .never files exist for forbidden secrets (e.g., unseal_*-SUITEONLY.never in SUITE dir)
  - [ ] Errors: Wrong suffix for deployment type
  - [ ] Warnings: Missing .never files for forbidden secrets
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_secret_suffixes.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_secret_suffixes_test.go` (new)
- **Evidence**:
  - `test-output/phase4/task-4.5-secret-suffix-validation.log`

---

### Phase 5: Pre-Commit Integration (4h)

**Phase Objective**: Add all new validations to pre-commit hooks for enforcement

#### Task 5.1: Add validate-mirror to Pre-Commit
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 4 complete
- **Description**: Add structural mirror validation to pre-commit
- **Acceptance Criteria**:
  - [ ] Updated: `.pre-commit-config.yaml` with validate-mirror hook
  - [ ] Hook runs: On changes to configs/ or deployments/ directories
  - [ ] Hook passes: `pre-commit run --all-files`
  - [ ] Documentation: Updated DEV-SETUP.md with new hook
- **Files**:
  - `.pre-commit-config.yaml` (updated)
  - `docs/DEV-SETUP.md` (updated)
- **Evidence**:
  - `test-output/phase5/task-5.1-precommit-test.log`

#### Task 5.2: Add validate-config to Pre-Commit
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.1
- **Description**: Add config validation to pre-commit
- **Acceptance Criteria**:
  - [ ] Updated: `.pre-commit-config.yaml` with validate-config hook
  - [ ] Hook runs: On changes to configs/**/*.yml or deployments/**/config/*.yml
  - [ ] Hook validates: All 8 new validation types
  - [ ] Hook passes: `pre-commit run --all-files`
- **Files**:
  - `.pre-commit-config.yaml` (updated)
- **Evidence**:
  - `test-output/phase5/task-5.2-precommit-test.log`

#### Task 5.3: Add validate-compose to Pre-Commit
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.2
- **Description**: Add compose validation to pre-commit
- **Acceptance Criteria**:
  - [ ] Updated: `.pre-commit-config.yaml` with validate-compose hook
  - [ ] Hook runs: On changes to deployments/**/compose.yml
  - [ ] Hook validates: All 7 compose validation types + new port offset validation
  - [ ] Hook passes: `pre-commit run --all-files`
- **Files**:
  - `.pre-commit-config.yaml` (updated)
- **Evidence**:
  - `test-output/phase5/task-5.3-precommit-test.log`

#### Task 5.4: Performance Optimization
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 5.1-5.3
- **Description**: Optimize pre-commit hooks for performance (<30s target)
- **Acceptance Criteria**:
  - [ ] Hooks only validate changed files (not all files)
  - [ ] Validation runs in parallel where possible
  - [ ] Total pre-commit time <30s for typical changes
  - [ ] Documentation: Performance tuning notes in .pre-commit-config.yaml
- **Files**:
  - `.pre-commit-config.yaml` (updated)
- **Evidence**:
  - `test-output/phase5/task-5.4-performance-benchmark.txt`

---

### Phase 6: Documentation & Testing (6h)

**Phase Objective**: Complete documentation and comprehensive testing

#### Task 6.1: Update ARCHITECTURE.md Section 12.4
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 5 complete
- **Description**: Document all new validation types
- **Acceptance Criteria**:
  - [ ] Updated: Section 12.4 with all 8 new config validations
  - [ ] Updated: Section 12.4 with all 5 new deployment validations
  - [ ] Cross-references: Link to validate_config_*.go, validate_product.go, validate_suite.go
  - [ ] Examples: CLI usage for each validation command
- **Files**:
  - `docs/ARCHITECTURE.md` (updated)
- **Evidence**:
  - `test-output/phase6/task-6.1-diff.txt` - Changes to ARCHITECTURE.md

#### Task 6.2: Update CONFIG-SCHEMA.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 5 complete
- **Description**: Add PRODUCT/SUITE config examples
- **Acceptance Criteria**:
  - [ ] Added: PRODUCT-level config example (jose-app-common.yml)
  - [ ] Added: SUITE-level config example (cryptoutil-app-common.yml)
  - [ ] Documented: Port offset strategy (+0, +10000, +20000)
  - [ ] Documented: Delegation patterns (SUITE → PRODUCT → SERVICE)
- **Files**:
  - `docs/CONFIG-SCHEMA.md` (updated)
- **Evidence**:
  - `test-output/phase6/task-6.2-diff.txt`

#### Task 6.3: Create Migration Guide
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 5 complete
- **Description**: Document old → new naming migration
- **Acceptance Criteria**:
  - [ ] Created: `docs/CONFIG-MIGRATION.md`
  - [ ] Documents: Old naming (config.yml, config-pg-N.yml) → New naming (PRODUCT-SERVICE-app-VARIANT.yml)
  - [ ] Provides: Script usage examples (rename-configs-v3.sh)
  - [ ] Lists: All renamed files (before/after)
  - [ ] References: ARCHITECTURE.md, CONFIG-SCHEMA.md
- **Files**:
  - `docs/CONFIG-MIGRATION.md` (new)
- **Evidence**:
  - `docs/CONFIG-MIGRATION.md`

#### Task 6.4: Create E2E Validation Test
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 6.1-6.3
- **Description**: End-to-end test validating entire project structure
- **Acceptance Criteria**:
  - [ ] Created: `internal/cmd/cicd/lint_deployments/e2e_validation_test.go`
  - [ ] Tests: All 20 deployments validate successfully
  - [ ] Tests: All configs/ files validate successfully
  - [ ] Tests: Structural mirror validation passes
  - [ ] Tests: Consistency validation passes
  - [ ] Tests: Run in CI/CD workflow
  - [ ] Coverage: ≥95%
- **Files**:
  - `internal/cmd/cicd/lint_deployments/e2e_validation_test.go` (new)
  - `.github/workflows/ci-test.yml` (updated - add E2E validation job)
- **Evidence**:
  - `test-output/phase6/task-6.4-e2e-test.log`

---

## Cross-Cutting Tasks

### Testing
- [ ] Unit tests ≥95% coverage (production), ≥98% (infrastructure/utility)
- [ ] Integration tests pass
- [ ] E2E test passes (Task 6.4)
- [ ] Mutation testing ≥95% minimum (≥98% infrastructure)
- [ ] No skipped tests (except documented exceptions)
- [ ] Race detector clean: `go test -race ./...`

### Code Quality
- [ ] Linting passes: `golangci-lint run ./...`
- [ ] No new TODOs without tracking
- [ ] No security vulnerabilities
- [ ] Formatting clean: `gofumpt -s -w ./`
- [ ] Imports organized: `goimports -w ./`

### Documentation
- [ ] ARCHITECTURE.md updated (Task 6.1)
- [ ] CONFIG-SCHEMA.md updated (Task 6.2)
- [ ] CONFIG-MIGRATION.md created (Task 6.3)
- [ ] README.md files created (Tasks 1.10, 2.1-2.6)
- [ ] DEV-SETUP.md updated (pre-commit hooks)

### Pre-Commit
- [ ] validate-mirror hook added (Task 5.1)
- [ ] validate-config hook added (Task 5.2)
- [ ] validate-compose hook added (Task 5.3)
- [ ] Hooks performant (<30s, Task 5.4)
- [ ] All hooks pass: `pre-commit run --all-files`

---

## Notes / Deferred Work

**Deferred to Future Iterations**:
- PRODUCT/SUITE-level deployments in deployments/ (currently only configs/ has PRODUCT/SUITE structures; deployments/ has compose.yml but no config/ subdirectories for PRODUCT/SUITE)
- Automated config generation from schemas (nice-to-have, not blocking)
- Config diff tool (compare configs/ vs deployments/*/config/, not blocking)

**Decision Log**:
- development.yml, production.yml, test.yml handling: Document in Task 1.3 evidence (archive vs keep vs relocate)
- PRODUCT/SUITE configs location: configs/ for templates, deployments/ for runtime (Decision 2 in plan.md)

---

## Evidence Archive

[Will be populated during implementation]
- `test-output/phase0-research/` - Phase 0 research findings (analysis.md, lint outputs)
- `test-output/phase1/` - Phase 1 rename logs and validation
- `test-output/phase2/` - Phase 2 PRODUCT/SUITE config creation
- `test-output/phase3/` - Phase 3 config validation implementation
- `test-output/phase4/` - Phase 4 deployment validation implementation
- `test-output/phase5/` - Phase 5 pre-commit integration
- `test-output/phase6/` - Phase 6 documentation and E2E tests
