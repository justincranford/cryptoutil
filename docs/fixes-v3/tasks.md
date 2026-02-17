# Tasks - Configs/Deployments/CICD Rigor & Consistency v3

**Status**: 0 of 51 tasks complete (0%)
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

**Phase Objective**: Restructure configs/ to mirror deployments/ hierarchy at SERVICE level

#### Task 1.1: Rename cipher/ → cipher-im/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Rename configs/cipher/ to configs/cipher-im/ (SERVICE-level naming)
- **Acceptance Criteria**:
  - [ ] Directory renamed: `configs/cipher/` → `configs/cipher-im/`
  - [ ] Files updated: cipher.yml → cipher-im-app-common.yml (if exists)
  - [ ] Code references updated (search codebase for "configs/cipher/")
  - [ ] Tests pass: `go test ./...`
  - [ ] Build clean: `go build ./...`
- **Files**:
  - `configs/cipher-im/` (renamed from cipher/)
- **Evidence**:
  - `test-output/phase1/task-1.1-rename-log.txt` - Rename verification

#### Task 1.2: Rename pki/ → pki-ca/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Rename configs/pki/ to configs/pki-ca/ (SERVICE-level naming)
- **Acceptance Criteria**:
  - [ ] Directory renamed: `configs/pki/` → `configs/pki-ca/`
  - [ ] Files updated: pki-ca.yml → pki-ca-app-common.yml (follow naming pattern)
  - [ ] Code references updated
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/pki-ca/` (renamed from pki/)

#### Task 1.3: Restructure identity/ → identity-{authz,idp,rp,rs,spa}/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Split identity/ into 5 SERVICE subdirs per Decision 1 (quizme Q1 Answer C)
- **Acceptance Criteria**:
  - [ ] Created: `configs/identity-authz/` with identity-authz-app-*.yml files
  - [ ] Created: `configs/identity-idp/` with identity-idp-app-*.yml files
  - [ ] Created: `configs/identity-rp/` with identity-rp-app-*.yml files
  - [ ] Created: `configs/identity-rs/` with identity-rs-app-*.yml files
  - [ ] Created: `configs/identity-spa/` with identity-spa-app-*.yml files
  - [ ] Preserved: `configs/identity/policies/` (3 files - shared)
  - [ ] Preserved: `configs/identity/profiles/` (5 files - shared)
  - [ ] Preserved: `configs/identity/{development,production,test}.yml` at parent (Decision 2, quizme Q2 Answer B)
  - [ ] Service configs reference: `../development.yml` pattern
  - [ ] Handled: All existing identity/*.yml files moved/renamed correctly
  - [ ] Code references updated
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/identity-authz/identity-authz-app-common.yml` (new)
  - `configs/identity-authz/identity-authz-app-sqlite-1.yml` (new)
  - `configs/identity-authz/identity-authz-app-postgresql-1.yml` (new)
  - `configs/identity-authz/identity-authz-app-postgresql-2.yml` (new)
  - `configs/identity-idp/identity-idp-app-*.yml` (4 files, new)
  - `configs/identity-rp/identity-rp-app-*.yml` (4 files, new)
  - `configs/identity-rs/identity-rs-app-*.yml` (4 files, new)
  - `configs/identity-spa/identity-spa-app-*.yml` (4 files, new)
  - `configs/identity/policies/` (preserved)
  - `configs/identity/profiles/` (preserved)
  - `configs/identity/{development,production,test}.yml` (preserved)
- **Evidence**:
  - `test-output/phase1/task-1.3-restructure-plan.md` - Restructure decisions
  - `test-output/phase1/task-1.3-files-moved.txt` - File move verification

#### Task 1.4: Create sm-kms/ under configs/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Create configs/sm-kms/ (SERVICE-level, note: sm/ will be PRODUCT-level in Phase 2)
- **Acceptance Criteria**:
  - [ ] Created: `configs/sm-kms/`
  - [ ] Created: `configs/sm-kms/sm-kms-app-common.yml`
  - [ ] Created: `configs/sm-kms/sm-kms-app-sqlite-1.yml`
  - [ ] Created: `configs/sm-kms/sm-kms-app-postgresql-1.yml`
  - [ ] Created: `configs/sm-kms/sm-kms-app-postgresql-2.yml`
  - [ ] Content follows template pattern (Decision 4)
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/sm-kms/sm-kms-app-common.yml` (new)
  - `configs/sm-kms/sm-kms-app-sqlite-1.yml` (new)
  - `configs/sm-kms/sm-kms-app-postgresql-1.yml` (new)
  - `configs/sm-kms/sm-kms-app-postgresql-2.yml` (new)

#### Task 1.5: Rename jose/ → jose-ja/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Rename configs/jose/ to configs/jose-ja/ (SERVICE-level naming)
- **Acceptance Criteria**:
  - [ ] Directory renamed: `configs/jose/` → `configs/jose-ja/`
  - [ ] Files updated: jose.yml → jose-ja-app-common.yml (follow naming pattern)
  - [ ] Code references updated
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/jose-ja/` (renamed from jose/)

#### Task 1.6: Verify configs/ Structure Matches Deployments
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 1.1-1.5
- **Description**: Verify configs/ SERVICE-level subdirs mirror deployments/ SERVICE subdirs
- **Acceptance Criteria**:
  - [ ] All SERVICE subdirs exist: cipher-im/, pki-ca/, identity-{authz,idp,rp,rs,spa}/, sm-kms/, jose-ja/
  - [ ] Compare with deployments/ structure: `ls -la deployments/` vs `ls -la configs/`
  - [ ] No orphaned configs (all configs match deployed services)
  - [ ] Document structure mapping
  - [ ] Tests pass: `go test ./...`
- **Evidence**:
  - `test-output/phase1/task-1.6-structure-comparison.txt` - Structure verification

#### Task 1.7: Update Code References
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 1.1-1.5
- **Description**: Update all code references to renamed config directories
- **Acceptance Criteria**:
  - [ ] Search codebase: `grep -r "configs/cipher/" .` → 0 results
  - [ ] Search codebase: `grep -r "configs/pki/" . | grep -v pki-ca` → 0 results
  - [ ] Search codebase: `grep -r "configs/jose/" . | grep -v jose-ja` → 0 results
  - [ ] Search codebase: `grep -r "configs/identity/" . | grep -Ev "identity-|policies|profiles"` → only valid parent refs
  - [ ] All imports updated
  - [ ] Tests pass: `go test ./...`
  - [ ] Build clean: `go build ./...`
- **Evidence**:
  - `test-output/phase1/task-1.7-grep-results.txt` - Verification of no old references

#### Task 1.8: Create configs/README.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.6
- **Description**: Document configs/ structure and purpose
- **Acceptance Criteria**:
  - [ ] Created: `configs/README.md`
  - [ ] Explains: Purpose (reference/templates vs deployments/ runtime)
  - [ ] Documents: SERVICE/PRODUCT/SUITE hierarchy
  - [ ] Documents: Naming pattern (PRODUCT-SERVICE-app-{common,sqlite-1,postgresql-1,postgresql-2}.yml)
  - [ ] Lists: All subdirectories and their purpose
  - [ ] References: ARCHITECTURE.md Section 12.5, CONFIG-SCHEMA.md
  - [ ] Notes: Shared files (identity/policies/, identity/profiles/, environment yamls)
  - [ ] Minimal content (Decision 8, quizme Q8 Answer A)
- **Files**:
  - `configs/README.md` (new)

---

### Phase 2: PRODUCT/SUITE Config Creation (6h)

**Phase Objective**: Create PRODUCT-level (cipher/, jose/, identity/, sm/, pki/) and SUITE-level (cryptoutil/) configs following template patterns

#### Task 2.1: Create cipher/ PRODUCT configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 complete
- **Description**: Add PRODUCT-level configs for cipher (delegates to cipher-im/)
- **Acceptance Criteria**:
  - [ ] Created: `configs/cipher/cipher-app-common.yml` (PRODUCT-level, delegates to cipher-im/)
  - [ ] Created: `configs/cipher/cipher-app-sqlite-1.yml`
  - [ ] Created: `configs/cipher/cipher-app-postgresql-1.yml`
  - [ ] Created: `configs/cipher/cipher-app-postgresql-2.yml`
  - [ ] Created: `configs/cipher/README.md` (minimal: purpose, delegation, ARCHITECTURE.md link per Decision 8)
  - [ ] Preserved: `configs/cipher-im/` (from Phase 1)
  - [ ] Content follows template pattern (Decision 4, Decision 7)
  - [ ] Naming follows: {PRODUCT}-app-{variant}.yml pattern
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/cipher/cipher-app-common.yml` (new, PRODUCT-level)
  - `configs/cipher/cipher-app-sqlite-1.yml` (new)
  - `configs/cipher/cipher-app-postgresql-1.yml` (new)
  - `configs/cipher/cipher-app-postgresql-2.yml` (new)
  - `configs/cipher/README.md` (new)

#### Task 2.2: Create pki/ PRODUCT configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 complete
- **Description**: Add PRODUCT-level configs for pki (delegates to pki-ca/)
- **Acceptance Criteria**:
  - [ ] Created: `configs/pki/pki-app-common.yml`
  - [ ] Created: `configs/pki/pki-app-sqlite-1.yml`
  - [ ] Created: `configs/pki/pki-app-postgresql-1.yml`
  - [ ] Created: `configs/pki/pki-app-postgresql-2.yml`
  - [ ] Created: `configs/pki/README.md` (minimal)
  - [ ] Preserved: `configs/pki-ca/` (from Phase 1)
  - [ ] Content follows template pattern
- **Files**:
  - `configs/pki/pki-app-common.yml` (new)
  - `configs/pki/pki-app-sqlite-1.yml` (new)
  - `configs/pki/pki-app-postgresql-1.yml` (new)
  - `configs/pki/pki-app-postgresql-2.yml` (new)
  - `configs/pki/README.md` (new)

#### Task 2.3: Create identity/ PRODUCT configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 complete
- **Description**: Add PRODUCT-level configs for identity (delegates to 5 services)
- **Acceptance Criteria**:
  - [ ] Created: `configs/identity/identity-app-common.yml` (PRODUCT-level, delegates to 5 services)
  - [ ] Created: `configs/identity/identity-app-sqlite-1.yml`
  - [ ] Created: `configs/identity/identity-app-postgresql-1.yml`
  - [ ] Created: `configs/identity/identity-app-postgresql-2.yml`
  - [ ] Created: `configs/identity/README.md` (minimal)
  - [ ] Preserved: `configs/identity/policies/`, `configs/identity/profiles/`
  - [ ] Preserved: `configs/identity/{development,production,test}.yml`
  - [ ] Preserved: `configs/identity-{authz,idp,rp,rs,spa}/` (from Phase 1)
  - [ ] Content follows template pattern
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
- **Description**: Add PRODUCT-level configs for sm (delegates to sm-kms/)
- **Acceptance Criteria**:
  - [ ] Created: `configs/sm/sm-app-common.yml`
  - [ ] Created: `configs/sm/sm-app-sqlite-1.yml`
  - [ ] Created: `configs/sm/sm-app-postgresql-1.yml`
  - [ ] Created: `configs/sm/sm-app-postgresql-2.yml`
  - [ ] Created: `configs/sm/README.md` (minimal)
  - [ ] Preserved: `configs/sm-kms/` (from Phase 1)
  - [ ] Content follows template pattern
- **Files**:
  - `configs/sm/sm-app-common.yml` (new)
  - `configs/sm/sm-app-sqlite-1.yml` (new)
  - `configs/sm/sm-app-postgresql-1.yml` (new)
  - `configs/sm/sm-app-postgresql-2.yml` (new)
  - `configs/sm/README.md` (new)

#### Task 2.5: Create jose/ PRODUCT configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 complete
- **Description**: Add PRODUCT-level configs for jose (delegates to jose-ja/)
- **Acceptance Criteria**:
  - [ ] Created: `configs/jose/jose-app-common.yml`
  - [ ] Created: `configs/jose/jose-app-sqlite-1.yml`
  - [ ] Created: `configs/jose/jose-app-postgresql-1.yml`
  - [ ] Created: `configs/jose/jose-app-postgresql-2.yml`
  - [ ] Created: `configs/jose/README.md` (minimal)
  - [ ] Preserved: `configs/jose-ja/` (from Phase 1)
  - [ ] Content follows template pattern
- **Files**:
  - `configs/jose/jose-app-common.yml` (new)
  - `configs/jose/jose-app-sqlite-1.yml` (new)
  - `configs/jose/jose-app-postgresql-1.yml` (new)
  - `configs/jose/jose-app-postgresql-2.yml` (new)
  - `configs/jose/README.md` (new)

#### Task 2.6: Create cryptoutil/ SUITE configs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 2.1-2.5
- **Description**: Add SUITE-level configs (delegates to all 5 products: cipher, pki, identity, sm, jose)
- **Acceptance Criteria**:
  - [ ] Created: `configs/cryptoutil/cryptoutil-app-common.yml` (SUITE-level, delegates to 5 products)
  - [ ] Created: `configs/cryptoutil/cryptoutil-app-sqlite-1.yml`
  - [ ] Created: `configs/cryptoutil/cryptoutil-app-postgresql-1.yml`
  - [ ] Created: `configs/cryptoutil/cryptoutil-app-postgresql-2.yml`
  - [ ] Created: `configs/cryptoutil/README.md` (minimal)
  - [ ] Content follows template pattern
  - [ ] Delegation to all 5 products validated
- **Files**:
  - `configs/cryptoutil/cryptoutil-app-common.yml` (new, SUITE-level)
  - `configs/cryptoutil/cryptoutil-app-sqlite-1.yml` (new)
  - `configs/cryptoutil/cryptoutil-app-postgresql-1.yml` (new)
  - `configs/cryptoutil/cryptoutil-app-postgresql-2.yml` (new)
  - `configs/cryptoutil/README.md` (new)

---

### Phase 3: CICD Validation Implementation (18h)

**Phase Objective**: Implement comprehensive cicd lint-deployments validation (8 types from Decision 5, Decision 6)

#### Task 3.1: Implement ValidateNaming
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Implement naming validation (Decision 6: cicd lint-deployments validate-naming)
- **Acceptance Criteria**:
  - [ ] Function: `ValidateNaming(filePath string) error`
  - [ ] Validates: File names follow {PRODUCT-SERVICE}-{app|compose}-{variant}.{yml|yaml} pattern
  - [ ] Validates: Directory structure (SERVICE/PRODUCT/SUITE levels)
  - [ ] Examples: cipher-im-app-common.yml ✓, cipher_im_app.yml ✗, CipherIM-app.yml ✗
  - [ ] Unit tests: ≥98% coverage
  - [ ] Tests: Valid names pass, invalid names fail with clear error messages
  - [ ] Tests pass: `go test ./internal/cmd/cicd/lint_deployments/validate_naming_test.go`
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_naming.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_naming_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.1-coverage.txt` - Coverage report

#### Task 3.2: Implement ValidateKebabCase
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Implement kebab-case validation (Decision 6: cicd lint-deployments validate-kebab-case)
- **Acceptance Criteria**:
  - [ ] Function: `ValidateKebabCase(filePath string) error`
  - [ ] Validates: All YAML keys in kebab-case (no snake_case, no camelCase)
  - [ ] Examples: bind-public-address ✓, bind_public_address ✗, bindPublicAddress ✗
  - [ ] Recursively checks nested keys
  - [ ] Unit tests: ≥98% coverage
  - [ ] Tests: Valid kebab-case passes, violations fail with key path
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_kebab_case.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_kebab_case_test.go` (new)

#### Task 3.3: Implement ValidateSchema
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Implement schema validation (Decision 6: cicd lint-deployments validate-schema)
- **Acceptance Criteria**:
  - [ ] Function: `ValidateSchema(filePath string, schemaType SchemaType) error`
  - [ ] SchemaType: Config (CONFIG-SCHEMA.md keys) or Compose (Docker Compose spec)
  - [ ] Validates: All config keys exist in CONFIG-SCHEMA.md
  - [ ] Validates: All compose keys valid per Docker Compose spec
  - [ ] Reports: Unknown keys, typos, deprecated keys
  - [ ] Unit tests: ≥98% coverage
  - [ ] Integration tests: Validate against real config/compose files
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_schema.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_schema_test.go` (new)
  - `internal/cmd/cicd/lint_deployments/schema_config.go` (config schema definitions)
  - `internal/cmd/cicd/lint_deployments/schema_compose.go` (compose schema definitions)

#### Task 3.4: Implement ValidatePorts
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Implement port validation (Decision 6: cicd lint-deployments validate-ports)
- **Acceptance Criteria**:
  - [ ] Function: `ValidatePorts(filePath string) error`
  - [ ] Validates: Public ports 8XXX (service-specific ranges)
  - [ ] Validates: Admin ports 9090 (ALWAYS 127.0.0.1)
  - [ ] Validates: PostgreSQL ports 543XX
  - [ ] Validates: Telemetry ports (4317 gRPC, 4318 HTTP, 3000 Grafana)
  - [ ] Checks: No port conflicts within service
  - [ ] Unit tests: ≥98% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_ports.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_ports_test.go` (new)

#### Task 3.5: Implement ValidateTelemetry
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Implement telemetry validation (Decision 6: cicd lint-deployments validate-telemetry)
- **Acceptance Criteria**:
  - [ ] Function: `ValidateTelemetry(filePath string) error`
  - [ ] Validates: observability.otlp.protocol (grpc or http)
  - [ ] Validates: observability.otlp.endpoint exists
  - [ ] Validates: observability.otlp.service-name matches service
  - [ ] Validates: observability.otlp.insecure flag (true dev, false prod)
  - [ ] Unit tests: ≥98% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_telemetry.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_telemetry_test.go` (new)

#### Task 3.6: Implement ValidateAdmin
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Implement admin endpoint validation (Decision 6: cicd lint-deployments validate-admin)
- **Acceptance Criteria**:
  - [ ] Function: `ValidateAdmin(filePath string) error`
  - [ ] Validates: bind-private-address ALWAYS "127.0.0.1" (NEVER 0.0.0.0)
  - [ ] Validates: bind-private-port 9090
  - [ ] Validates: bind-private-protocol "https"
  - [ ] Security: CRITICAL - admin endpoints MUST NOT be publicly accessible
  - [ ] Unit tests: ≥98% coverage
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_admin.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_admin_test.go` (new)

#### Task 3.7: Implement ValidateConsistency
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 3.1-3.6
- **Description**: Implement config-compose consistency validation
- **Acceptance Criteria**:
  - [ ] Function: `ValidateConsistency(configPath, composePath string) error`
  - [ ] Validates: Services in config match services in compose
  - [ ] Validates: Ports in config match exposed ports in compose
  - [ ] Validates: Service names consistent (kebab-case)
  - [ ] Validates: Dependency ordering (depends_on matches config references)
  - [ ] Unit tests: ≥98% coverage
  - [ ] Integration tests: Validate real config+compose pairs
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_consistency.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_consistency_test.go` (new)

#### Task 3.8: Implement ValidateSecrets
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Implement secrets validation (Decision 6: cicd lint-deployments validate-secrets)
- **Acceptance Criteria**:
  - [ ] Function: `ValidateSecrets(filePath string) error`
  - [ ] Validates: NO inline credentials (password: "secret123" ✗)
  - [ ] Validates: ALL secrets use Docker secrets pattern (file:///run/secrets/*)
  - [ ] Validates: Secret file permissions 440 (r--r-----) if file exists
  - [ ] Detects: Passwords, API keys, tokens in config values
  - [ ] Unit tests: ≥98% coverage
  - [ ] Security: CRITICAL - prevent credential leaks
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_secrets.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_secrets_test.go` (new)

#### Task 3.9: Integrate All Validators into cicd lint-deployments
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 3.1-3.8
- **Description**: Wire all 8 validators into cicd lint-deployments command
- **Acceptance Criteria**:
  - [ ] Command: `cicd lint-deployments validate-naming <file>`
  - [ ] Command: `cicd lint-deployments validate-kebab-case <file>`
  - [ ] Command: `cicd lint-deployments validate-schema <file> --type={config|compose}`
  - [ ] Command: `cicd lint-deployments validate-ports <file>`
  - [ ] Command: `cicd lint-deployments validate-telemetry <file>`
  - [ ] Command: `cicd lint-deployments validate-admin <file>`
  - [ ] Command: `cicd lint-deployments validate-consistency --config=<file> --compose=<file>`
  - [ ] Command: `cicd lint-deployments validate-secrets <file>`
  - [ ] Command: `cicd lint-deployments validate-all <file>` (runs all 8 validators)
  - [ ] Help text for each command
  - [ ] Exit codes: 0 success, 1 validation failed, 2 usage error
  - [ ] Tests pass: `go test ./cmd/cicd/...`
- **Files**:
  - `cmd/cicd/lint_deployments.go` (updated)
  - `internal/cmd/cicd/lint_deployments/command.go` (new)

#### Task 3.10: Add Mutation Testing for Validators
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.9
- **Description**: Add mutation tests for all validators (≥98% mutation score)
- **Acceptance Criteria**:
  - [ ] Mutation tests: `gremlins unleash --tags=!integration ./internal/cmd/cicd/lint_deployments/`
  - [ ] Mutation score: ≥98% (infrastructure code requirement)
  - [ ] All validation logic mutation-tested
  - [ ] Edge cases covered
- **Evidence**:
  - `test-output/phase3/task-3.10-mutation-results.txt` - Mutation testing results

#### Task 3.11: Validator Performance Benchmarks (PRIORITY 1 - From Analysis)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.9
- **Description**: Add performance benchmarks for all validators to ensure <5s pre-commit
- **Acceptance Criteria**:
  - [ ] Benchmark tests: `go test -bench=. ./internal/cmd/cicd/lint_deployments/`
  - [ ] Target: <5s for incremental validation (staged files only)
  - [ ] Target: <30s for full validation (all 50+ configs)
  - [ ] Per-validator benchmarks: BenchmarkValidateNaming, BenchmarkValidateSchema, etc.
  - [ ] Results documented: `test-output/phase3/task-3.11-benchmark-results.txt`
  - [ ] Identify bottlenecks for optimization
- **Files**:
  - `internal/cmd/cicd/lint_deployments/validate_naming_bench_test.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_schema_bench_test.go` (new)
  - `internal/cmd/cicd/lint_deployments/validate_ports_bench_test.go` (new)
  - (similar for all 8 validators)
- **Evidence**:
  - `test-output/phase3/task-3.11-benchmark-results.txt` - Benchmark report

#### Task 3.12: Implement Validation Caching (PRIORITY 1 - From Analysis)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.11
- **Description**: Add file hash-based validation caching to achieve <5s pre-commit
- **Acceptance Criteria**:
  - [ ] Cache structure:
    ```go
    type ValidationCache struct {
        FileHash  string    // SHA256 of file content
        Result    error     // cached validation result
        Timestamp time.Time // cache creation time
    }
    ```
  - [ ] Cache storage: `.git/hooks/validation-cache.json`
  - [ ] Cache invalidation: On file content change (hash mismatch)
  - [ ] Cache hit: Skip validation, use cached result (saves ~500ms per file)
  - [ ] Cache miss: Run validation, update cache
  - [ ] Parallel validation: Both cached and uncached files
  - [ ] Tests: Unit tests for cache operations (≥98% coverage)
  - [ ] Performance: Measure cache hit improvement
- **Files**:
  - `internal/cmd/cicd/lint_deployments/cache.go` (new)
  - `internal/cmd/cicd/lint_deployments/cache_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.12-cache-performance.txt` - Before/after benchmarks

---

### Phase 4: ARCHITECTURE.md Updates (6h)

**Phase Objective**: Document deployment/config rigor patterns in ARCHITECTURE.md

#### Task 4.1: Add Section 12.4 - Deployment Validation Architecture
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 3 complete
- **Description**: Document 8 validation types in ARCHITECTURE.md
- **Acceptance Criteria**:
  - [ ] Created: Section 12.4 in docs/ARCHITECTURE.md
  - [ ] Documents: All 8 validation types (naming, kebab-case, schema, ports, telemetry, admin, consistency, secrets)
  - [ ] Documents: cicd lint-deployments command usage
  - [ ] Documents: Pre-commit hook integration
  - [ ] Documents: Validation workflow (pre-commit → CI/CD → deployment)
  - [ ] Examples: Valid and invalid patterns for each validator
  - [ ] Cross-references: CONFIG-SCHEMA.md, .pre-commit-config.yaml
  - [ ] Table of contents updated
- **Files**:
  - `docs/ARCHITECTURE.md` (updated)
- **Evidence**:
  - `test-output/phase4/task-4.1-section-preview.txt` - Section 12.4 preview

#### Task 4.2: Add Section 12.5 - Config File Architecture
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 2 complete
- **Description**: Document SERVICE/PRODUCT/SUITE config hierarchy
- **Acceptance Criteria**:
  - [ ] Created: Section 12.5 in docs/ARCHITECTURE.md
  - [ ] Documents: SERVICE/PRODUCT/SUITE hierarchy
  - [ ] Documents: Template pattern requirements (Decision 4, Decision 7)
  - [ ] Documents: Naming conventions ({PRODUCT-SERVICE}-app-{variant}.yml)
  - [ ] Documents: Config schema compliance (CONFIG-SCHEMA.md)
  - [ ] Documents: Shared files pattern (identity/policies/, environment yamls)
  - [ ] Examples: Config structures for each level
  - [ ] Diagrams: configs/ directory structure
  - [ ] Cross-references: Section 12.4 validation, CONFIG-SCHEMA.md
- **Files**:
  - `docs/ARCHITECTURE.md` (updated)

#### Task 4.3: Add Section 12.6 - Secrets Management
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Document secrets management patterns
- **Acceptance Criteria**:
  - [ ] Created: Section 12.6 in docs/ARCHITECTURE.md
  - [ ] Documents: Docker secrets pattern (MANDATORY)
  - [ ] Documents: File permissions (440 r--r-----)
  - [ ] Documents: NO inline credentials rule
  - [ ] Documents: Secret file locations (/run/secrets/)
  - [ ] Documents: Validation (ValidateSecrets hook)
  - [ ] Examples: Correct and incorrect secret handling
  - [ ] Security: Emphasizes CRITICAL nature of secrets validation
- **Files**:
  - `docs/ARCHITECTURE.md` (updated)

#### Task 4.4: Update ARCHITECTURE-INDEX.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 4.1-4.3
- **Description**: Update ARCHITECTURE-INDEX.md with new sections
- **Acceptance Criteria**:
  - [ ] Added: Section 12.4 Deployment Validation Architecture (line ranges)
  - [ ] Added: Section 12.5 Config File Architecture (line ranges)
  - [ ] Added: Section 12.6 Secrets Management (line ranges)
  - [ ] Verified: Line number ranges accurate
  - [ ] Verified: Semantic topics match new content
- **Files**:
  - `docs/ARCHITECTURE-INDEX.md` (updated)

#### Task 4.5: Validate ARCHITECTURE.md Cross-References (PRIORITY 1 - From Analysis)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 4.1-4.3
- **Description**: Create tool to validate ARCHITECTURE.md cross-references are consistent
- **Acceptance Criteria**:
  - [ ] Created: `cmd/cicd/validate-arch-consistency/`
  - [ ] Validates: All section references in ARCHITECTURE.md exist
  - [ ] Validates: All references to CONFIG-SCHEMA.md point to existing sections
  - [ ] Validates: All instruction file references to ARCHITECTURE.md sections exist
  - [ ] Extracts: Patterns from ARCHITECTURE.md Section 12.4-12.6
  - [ ] Checks: Each pattern mentioned in instruction files
  - [ ] Reports: Missing references, broken links, orphaned sections
  - [ ] Tests: Unit tests ≥98% coverage
  - [ ] Run: `cicd validate-arch-consistency`
- **Files**:
  - `cmd/cicd/validate_arch_consistency.go` (new)
  - `internal/cmd/cicd/validate_arch_consistency/validator.go` (new)
  - `internal/cmd/cicd/validate_arch_consistency/validator_test.go` (new)
- **Evidence**:
  - `test-output/phase4/task-4.5-cross-ref-report.txt` - Validation report

---

### Phase 5: Instruction File Propagation (4h)

**Phase Objective**: Propagate ARCHITECTURE.md deployment patterns to all instruction files

#### Task 5.1: Update 04-01.deployment.instructions.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 4 complete
- **Description**: Propagate deployment patterns to deployment instructions
- **Acceptance Criteria**:
  - [ ] Added: 8 CICD validation types section
  - [ ] Added: SERVICE/PRODUCT/SUITE hierarchy section
  - [ ] Added: Template pattern requirements
  - [ ] Added: Secrets management rules
  - [ ] Added: Cross-references to ARCHITECTURE.md sections 12.4-12.6
  - [ ] Updated: Existing deployment patterns for consistency
  - [ ] Verified: No conflicting patterns with ARCHITECTURE.md
  - [ ] Tests: Instruction file parses correctly (YAML frontmatter valid)
- **Files**:
  - `.github/instructions/04-01.deployment.instructions.md` (updated)
- **Evidence**:
  - `test-output/phase5/task-5.1-diff.txt` - Changes made

#### Task 5.2: Update 02-01.architecture.instructions.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 4 complete
- **Description**: Add deployment architecture quick reference
- **Acceptance Criteria**:
  - [ ] Added: Deployment validation quick reference (8 types)
  - [ ] Added: Config hierarchy quick reference (SERVICE/PRODUCT/SUITE)
  - [ ] Updated: Service template pattern (if deployment-related)
  - [ ] Cross-references: ARCHITECTURE.md sections 12.4-12.6
  - [ ] Verified: Consistency with ARCHITECTURE.md
- **Files**:
  - `.github/instructions/02-01.architecture.instructions.md` (updated)

#### Task 5.3: Verify Agent Files (if needed)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 5.1-5.2
- **Description**: Check if agent files reference deployment patterns, update if needed
- **Acceptance Criteria**:
  - [ ] Searched: `.github/agents/` for deployment/config/validation references
  - [ ] Updated: implementation-planning.agent.md if deployment patterns referenced
  - [ ] Updated: implementation-execution.agent.md if deployment patterns referenced
  - [ ] Verified: No outdated patterns in agent files
  - [ ] Documented: Changes made or "no changes needed"
- **Evidence**:
  - `test-output/phase5/task-5.3-agent-review.txt` - Agent file review results

#### Task 5.4: Automated Doc Consistency Check (PRIORITY 1 - From Analysis)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.3
- **Description**: Create tool to verify ARCHITECTURE.md patterns propagated to instruction files
- **Acceptance Criteria**:
  - [ ] Created: `cmd/cicd/check-doc-consistency/`
  - [ ] Extracts: Headings from ARCHITECTURE.md sections 12.4-12.6
  - [ ] Searches: Each heading in instruction files (.github/instructions/)
  - [ ] Reports: Patterns in ARCHITECTURE.md but missing in instructions (not propagated)
  - [ ] Reports: Patterns in instructions but not in ARCHITECTURE.md (orphaned)
  - [ ] Reports: Conflicting descriptions between ARCHITECTURE.md and instructions
  - [ ] Uses: Checklist-based approach (per deep-analysis.md recommendation)
  - [ ] Checklist example:
    ```markdown
    ### ARCHITECTURE.md Section 12.4 - Deployment Validation
    - [ ] 04-01.deployment.instructions.md: ValidateNaming documented
    - [ ] 04-01.deployment.instructions.md: ValidateKebabCase documented
    - [ ] 04-01.deployment.instructions.md: ValidateSchema documented
    - [ ] 04-01.deployment.instructions.md: ValidatePorts documented
    - [ ] 04-01.deployment.instructions.md: ValidateTelemetry documented
    - [ ] 04-01.deployment.instructions.md: ValidateAdmin documented
    - [ ] 04-01.deployment.instructions.md: ValidateConsistency documented
    - [ ] 04-01.deployment.instructions.md: ValidateSecrets documented
    - [ ] 02-01.architecture.instructions.md: Cross-reference to Section 12.4
    ```
  - [ ] Tests: Unit tests ≥98% coverage
  - [ ] Run: `cicd check-doc-consistency`
- **Files**:
  - `cmd/cicd/check_doc_consistency.go` (new)
  - `internal/cmd/cicd/check_doc_consistency/checker.go` (new)
  - `internal/cmd/cicd/check_doc_consistency/checker_test.go` (new)
  - `internal/cmd/cicd/check_doc_consistency/checklist.go` (propagation checklist)
- **Evidence**:
  - `test-output/phase5/task-5.4-consistency-report.txt` - Consistency check report

---

### Phase 6: E2E Validation (3h)

**Phase Objective**: Validate all configs and deployments pass comprehensive validation

#### Task 6.1: Run Validation Against All configs/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 3 complete
- **Description**: Run cicd lint-deployments validate-all against ALL configs/ files
- **Acceptance Criteria**:
  - [ ] Run: `find configs/ -name "*.yml" -o -name "*.yaml" | xargs -I {} cicd lint-deployments validate-all {}`
  - [ ] Result: 100% pass (0 failures)
  - [ ] Report: Summary of files validated
  - [ ] Report: Validation time per file
  - [ ] If failures: Fix config files OR adjust validation rules (document decision)
- **Evidence**:
  - `test-output/phase6/task-6.1-configs-validation.txt` - Validation results

#### Task 6.2: Run Validation Against All deployments/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 3 complete
- **Description**: Run cicd lint-deployments validate-all against ALL deployments/ files
- **Acceptance Criteria**:
  - [ ] Run: `find deployments/ -name "*.yml" -o -name "*.yaml" | xargs -I {} cicd lint-deployments validate-all {}`
  - [ ] Result: 100% pass (0 failures)
  - [ ] Report: Summary of files validated
  - [ ] If failures: Fix deployment files OR adjust validation rules
- **Evidence**:
  - `test-output/phase6/task-6.2-deployments-validation.txt` - Validation results

#### Task 6.3: Test Pre-Commit Integration
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 6.1-6.2
- **Description**: Verify pre-commit hooks call cicd lint-deployments correctly
- **Acceptance Criteria**:
  - [ ] Modified: Sample config file (introduce validation error)
  - [ ] Run: `git add <file> && pre-commit run`
  - [ ] Result: Pre-commit hook catches validation error
  - [ ] Result: Clear error message shown
  - [ ] Revert: Sample file to clean state
  - [ ] Verified: Pre-commit functional
- **Evidence**:
  - `test-output/phase6/task-6.3-precommit-test.txt` - Pre-commit test results

#### Task 6.4: Test Sample Violations
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 3 complete
- **Description**: Verify each validator detects violations correctly
- **Acceptance Criteria**:
  - [ ] Created: `test-output/phase6/sample-violations/` directory
  - [ ] Created: Sample files with known violations (one per validator)
  - [ ] Tested: Each validator against its sample violation
  - [ ] Result: All violations detected correctly
  - [ ] Result: Clear error messages for each
- **Evidence**:
  - `test-output/phase6/sample-violations/` - Sample violation files
  - `test-output/phase6/task-6.4-violations-test.txt` - Violation detection results

---

## Cross-Cutting Tasks

### Testing
- [ ] Unit tests ≥98% coverage (infrastructure code: CICD validators)
- [ ] Integration tests pass (validate real config/compose files)
- [ ] Mutation testing ≥98% (infrastructure code: CICD validators)
- [ ] No skipped tests (except documented exceptions)
- [ ] Race detector clean: `go test -race ./...`

### Code Quality
- [ ] Linting passes: `golangci-lint run ./...`
- [ ] No new TODOs without tracking
- [ ] No security vulnerabilities: `gosec ./...`
- [ ] Formatting clean: `gofumpt -s -w ./`
- [ ] Imports organized: `goimports -w ./`

### Documentation
- [ ] ARCHITECTURE.md updated (sections 12.4-12.6)
- [ ] ARCHITECTURE-INDEX.md updated
- [ ] Instruction files updated (04-01, 02-01)
- [ ] Agent files reviewed (implementation-planning, implementation-execution)
- [ ] README.md files created (configs/, each PRODUCT/SUITE dir)
- [ ] CONFIG-SCHEMA.md referenced correctly
- [ ] Comments added for complex validation logic

### Deployment
- [ ] All configs/ pass validation (100%)
- [ ] All deployments/ pass validation (100%)
- [ ] Pre-commit hooks functional
- [ ] Sample violations detected correctly
- [ ] cicd lint-deployments command fully functional

---

## Notes / Deferred Work

*None at this time. All planned work in scope for v3.*

---

## Evidence Archive

- `test-output/fixes-v3-quizme-analysis/` - Quizme-v1 answers analysis
- `test-output/phase1/` - configs/ restructuring verification
- `test-output/phase2/` - PRODUCT/SUITE config creation
- `test-output/phase3/` - CICD validation implementation + testing
- `test-output/phase4/` - ARCHITECTURE.md updates verification
- `test-output/phase5/` - Instruction file propagation checks
- `test-output/phase6/` - E2E validation results
