# Tasks - Configs/Deployments/CICD Rigor & Consistency v3

**Status**: 0 of 57 tasks complete (0%)
**Last Updated**: 2026-02-17
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

**Phase Objective**: Restructure configs/ to mirror deployments/ SERVICE-level hierarchy (9 subdirs total)

#### Task 1.1: Rename cipher/ →  cipher-im/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Rename configs/cipher/ to configs/cipher-im/ for SERVICE-level naming consistency
- **Acceptance Criteria**:
  - [ ] Directory renamed: `git mv configs/cipher configs/cipher-im`
  - [ ] Files inside unchanged (no content modifications)
  - [ ] No broken paths (search codebase for "configs/cipher" references)
  - [ ] Tests pass: `go test ./...`
  - [ ] Build clean: `go build ./...`
- **Files**:
  - `configs/cipher-im/` (renamed from cipher/)
- **Evidence**:
  - `test-output/phase1/task-1.1-rename-cipher.log` - Rename verification

#### Task 1.2: Rename pki/ → pki-ca/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1 (sequential to avoid path conflicts)
- **Description**: Rename configs/pki/ to configs/pki-ca/ for SERVICE-level naming consistency
- **Acceptance Criteria**:
  - [ ] Directory renamed: `git mv configs/pki configs/pki-ca`
  - [ ] Files inside unchanged
  - [ ] No broken paths (search codebase for "configs/pki" references, exclude "configs/pki-ca")
  - [ ] Tests pass: `go test ./...`
  - [ ] Build clean: `go build ./...`
- **Files**:
  - `configs/pki-ca/` (renamed from pki/)
- **Evidence**:
  - `test-output/phase1/task-1.2-rename-pki.log`

#### Task 1.3: Restructure identity/ → 5 SERVICE subdirs
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.2
- **Description**: Create 5 SERVICE subdirs (authz, idp, rp, rs, spa) under configs/identity/ per Decision 1:C
- **Acceptance Criteria**:
  - [ ] 5 subdirs created: `identity/{authz,idp,rp,rs,spa}/`
  - [ ] Files moved to correct subdirs (e.g., authz-app.yml → identity/authz/app.yml)
  - [ ] Shared files preserved at parent: `identity/{policies/,profiles/,development.yml,production.yml,test.yml}` per Decision 2:B
  - [ ] Relative references updated in service configs: `../development.yml`
  - [ ] No broken paths (search for "configs/identity" references)
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/identity/authz/` (new subdir)
  - `configs/identity/idp/` (new subdir)
  - `configs/identity/rp/` (new subdir)
  - `configs/identity/rs/` (new subdir)
  - `configs/identity/spa/` (new subdir)
  - `configs/identity/policies/` (preserved)
  - `configs/identity/profiles/` (preserved)
  - `configs/identity/{development,production,test}.yml` (preserved)
- **Evidence**:
  - `test-output/phase1/task-1.3-identity-restructure.log`
  - `test-output/phase1/task-1.3-file-moves.txt` (list of moved files)

#### Task 1.4: Create sm-kms/ SERVICE directory
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.3
- **Description**: Create new configs/sm-kms/ SERVICE directory (previously missing from configs/)
- **Acceptance Criteria**:
  - [ ] Directory created: `mkdir -p configs/sm-kms/`
  - [ ] Placeholder config file: `configs/sm-kms/app.yml` (copy from deployments/sm-kms/ or create minimal)
  - [ ] Template validation passes (Task 3.4 will validate once implemented)
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - `configs/sm-kms/` (new directory)
  - `configs/sm-kms/app.yml` (new file)
- **Evidence**:
  - `test-output/phase1/task-1.4-sm-kms-creation.log`

#### Task 1.5: Rename jose/ → jose-ja/
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.4
- **Description**: Rename configs/jose/ to configs/jose-ja/ for SERVICE-level naming consistency
- **Acceptance Criteria**:
  - [ ] Directory renamed: `git mv configs/jose configs/jose-ja`
  - [ ] Files inside unchanged
  - [ ] No broken paths (search codebase for "configs/jose" references, exclude "configs/jose-ja")
  - [ ] Tests pass: `go test ./...`
  - [ ] Build clean: `go build ./...`
- **Files**:
  - `configs/jose-ja/` (renamed from jose/)
- **Evidence**:
  - `test-output/phase1/task-1.5-rename-jose.log`

#### Task 1.6: Update code references
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.5 (must complete all renames first)
- **Description**: Update ALL code references to old config paths
- **Acceptance Criteria**:
  - [ ] Search completed: `grep -r "configs/cipher[^-]" --include="*.go"` returns 0 results
  - [ ] Search completed: `grep -r "configs/pki[^-]" --include="*.go"` returns 0 results
  - [ ] Search completed: `grep -r "configs/jose[^-]" --include="*.go"` returns 0 results
  - [ ] All imports updated (if any config path constants)
  - [ ] All file.Open/os.ReadFile calls updated
  - [ ] Tests pass: `go test ./...`
  - [ ] Build clean: `go build ./...`
  - [ ] No linting errors: `golangci-lint run ./...`
- **Files**:
  - Various Go files (search results will identify)
- **Evidence**:
  - `test-output/phase1/task-1.6-code-references.log` (grep results before/after)

#### Task 1.7: Create configs/README.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.6
- **Description**: Document configs/ hierarchy with minimal README per Decision 8:A
- **Acceptance Criteria**:
  - [ ] README.md created at `configs/README.md`
  - [ ] Content: Purpose paragraph (what configs/ contains)
  - [ ] Content: SERVICE/PRODUCT/SUITE hierarchy explanation (brief)
  - [ ] Content: Link to ARCHITECTURE.md Section 12.5
  - [ ] Minimal length: 5-10 lines (no comprehensive docs per Decision 8:A)
  - [ ] Markdown formatting valid
- **Files**:
  - `configs/README.md` (new file)
- **Evidence**:
  - `test-output/phase1/task-1.7-readme-content.txt`

#### Task 1.8: Phase 1 E2E Verification
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.7
- **Description**: Comprehensive verification that Phase 1 restructuring is complete and correct
- **Acceptance Criteria**:
  - [ ] Directory count: 9 SERVICE subdirs (cipher-im, pki-ca, identity/{authz,idp,rp,rs,spa}, sm-kms, jose-ja)
  - [ ] Shared files preserved: identity/{policies/,profiles/,*.yml} at parent
  - [ ] All tests pass: `go test ./...`
  - [ ] Build clean: `go build ./...`
  - [ ] Linting clean: `golangci-lint run ./...`
  - [ ] No TODOs added without tracking
  - [ ] Evidence collected: All task logs archived
- **Files**:
  - N/A (verification only)
- **Evidence**:
  - `test-output/phase1/phase1-completion-verification.log`
  - `test-output/phase1/directory-structure.txt` (tree output)

---

### Phase 2: PRODUCT/SUITE Config Creation (6h)

**Phase Objective**: Create 5 PRODUCT-level + 1 SUITE-level configs with template pattern compliance

#### Task 2.1: Create cipher/ PRODUCT config
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 complete (Task 1.8)
- **Description**: Create PRODUCT-level config delegating to cipher-im/ SERVICE
- **Acceptance Criteria**:
  - [ ] Directory created: `configs/cipher/`
  - [ ] Config file: `configs/cipher/config.yml` with required keys (product-name, delegation: [cipher-im])
  - [ ] Port offset correct: PRODUCT ports = SERVICE + 10000 (per Decision 4A)
  - [ ] README.md: Purpose + delegation + ARCHITECTURE.md link (minimal per Decision 8:A)
  - [ ] Template validation passes (once Task 3.4 implemented)
- **Files**:
  - `configs/cipher/config.yml` (new)
  - `configs/cipher/README.md` (new)
- **Evidence**:
  - `test-output/phase2/task-2.1-cipher-product.yml` (generated config)

#### Task 2.2: Create pki/ PRODUCT config
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.1
- **Description**: Create PRODUCT-level config delegating to pki-ca/ SERVICE
- **Acceptance Criteria**:
  - [ ] Directory created: `configs/pki/`
  - [ ] Config file: `configs/pki/config.yml` with required keys (product-name, delegation: [pki-ca])
  - [ ] Port offset correct: PRODUCT ports = SERVICE + 10000
  - [ ] README.md: Minimal content per Decision 8:A
  - [ ] Template validation passes
- **Files**:
  - `configs/pki/config.yml` (new)
  - `configs/pki/README.md` (new)
- **Evidence**:
  - `test-output/phase2/task-2.2-pki-product.yml`

#### Task 2.3: Create identity/ PRODUCT config
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.2
- **Description**: Create PRODUCT-level config delegating to 5 identity services (authz, idp, rp, rs, spa)
- **Acceptance Criteria**:
  - [ ] Config file: `configs/identity/config.yml` (NOTE: identity/ already exists from Phase 1, add config.yml)
  - [ ] Delegation array: [identity-authz, identity-idp, identity-rp, identity-rs, identity-spa]
  - [ ] Port offset correct: PRODUCT ports = SERVICE + 10000
  - [ ] README.md: Update existing or create if missing (minimal per Decision 8:A)
  - [ ] Template validation passes
  - [ ] Shared files NOT affected (policies/, profiles/, *.yml preserved)
- **Files**:
  - `configs/identity/config.yml` (new, identity/ dir already exists)
  - `configs/identity/README.md` (new or updated)
- **Evidence**:
  - `test-output/phase2/task-2.3-identity-product.yml`

#### Task 2.4: Create sm/ PRODUCT config
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.3
- **Description**: Create PRODUCT-level config delegating to sm-kms/ SERVICE
- **Acceptance Criteria**:
  - [ ] Directory created: `configs/sm/`
  - [ ] Config file: `configs/sm/config.yml` with required keys (product-name, delegation: [sm-kms])
  - [ ] Port offset correct: PRODUCT ports = SERVICE + 10000
  - [ ] README.md: Minimal content per Decision 8:A
  - [ ] Template validation passes
- **Files**:
  - `configs/sm/config.yml` (new)
  - `configs/sm/README.md` (new)
- **Evidence**:
  - `test-output/phase2/task-2.4-sm-product.yml`

#### Task 2.5: Create jose/ PRODUCT config
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.4
- **Description**: Create PRODUCT-level config delegating to jose-ja/ SERVICE
- **Acceptance Criteria**:
  - [ ] Directory created: `configs/jose/`
  - [ ] Config file: `configs/jose/config.yml` with required keys (product-name, delegation: [jose-ja])
  - [ ] Port offset correct: PRODUCT ports = SERVICE + 10000
  - [ ] README.md: Minimal content per Decision 8:A
  - [ ] Template validation passes
- **Files**:
  - `configs/jose/config.yml` (new)
  - `configs/jose/README.md` (new)
- **Evidence**:
  - `test-output/phase2/task-2.5-jose-product.yml`

#### Task 2.6: Create cryptoutil/ SUITE config
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.5
- **Description**: Create SUITE-level config delegating to all 5 products (cipher, pki, identity, sm, jose)
- **Acceptance Criteria**:
  - [ ] Directory created: `configs/cryptoutil/`
  - [ ] Config file: `configs/cryptoutil/config.yml` with required keys (suite-name, delegation: [cipher, pki, identity, sm, jose])
  - [ ] Port offset correct: SUITE ports = SERVICE + 20000 (per Decision 4A)
  - [ ] README.md: Minimal content per Decision 8:A
  - [ ] Template validation passes
- **Files**:
  - `configs/cryptoutil/config.yml` (new)
  - `configs/cryptoutil/README.md` (new)
- **Evidence**:
  - `test-output/phase2/task-2.6-cryptoutil-suite.yml`

#### Task 2.7: Phase 2 E2E Verification
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.6
- **Description**: Verify all PRODUCT/SUITE configs created correctly
- **Acceptance Criteria**:
  - [ ] 5 PRODUCT configs exist: cipher/, pki/, identity/, sm/, jose/
  - [ ] 1 SUITE config exists: cryptoutil/
  - [ ] All configs have README.md (minimal content)
  - [ ] Template pattern compliance (manual check until Task 3.4 automated)
  - [ ] Tests pass: `go test ./...`
  - [ ] Build clean: `go build ./...`
- **Files**:
  - N/A (verification only)
- **Evidence**:
  - `test-output/phase2/phase2-completion-verification.log`
  - `test-output/phase2/generated-configs-list.txt`

---

### Phase 3: CICD Validation Implementation (23h)

**Phase Objective**: Implement 8 validators with ≥98% coverage/mutation, parallel execution, aggressive secrets detection

#### Task 3.1: Implement ValidateNaming
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 2 complete (Task 2.7)
- **Description**: Validate config filename patterns per Decision 4A naming rules
- **Acceptance Criteria**:
  - [ ] Validator implemented: `cmd/cicd/lint-deployments/validate-naming.go`
  - [ ] Unit tests: SERVICE (`{service-id}/{app-type}.yml`), PRODUCT (`{product-id}/config.yml`), SUITE (`cryptoutil/config.yml`)
  - [ ] Error messages: Moderate verbosity per Decision 14:B (file, line, issue, suggested fix)
  - [ ] Coverage ≥98%: `go test -cover ./cmd/cicd/lint-deployments/validate-naming`
  - [ ] Tests pass: `go test ./cmd/cicd/lint-deployments/validate-naming/...`
- **Files**:
  - `cmd/cicd/lint-deployments/validate-naming.go` (new)
  - `cmd/cicd/lint-deployments/validate-naming_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.1-naming-coverage.txt`

#### Task 3.2: Implement ValidateKebabCase
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.1
- **Description**: Validate kebab-case for keys and values
- **Acceptance Criteria**:
  - [ ] Validator implemented: `cmd/cicd/lint-deployments/validate-kebab-case.go`
  - [ ] Unit tests: Valid kebab-case, edge cases (numbers, leading/trailing hyphens, underscores)
  - [ ] Error messages: Moderate verbosity per Decision 14:B
  - [ ] Coverage ≥98%
  - [ ] Tests pass
- **Files**:
  - `cmd/cicd/lint-deployments/validate-kebab-case.go` (new)
  - `cmd/cicd/lint-deployments/validate-kebab-case_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.2-kebab-coverage.txt`

#### Task 3.3: Implement ValidateSchema
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.2
- **Description**: Validate configs against CONFIG-SCHEMA.md schema per Decision 10:D (embed + parse at init)
- **Acceptance Criteria**:
  - [ ] Embed CONFIG-SCHEMA.md: `//go:embed docs/CONFIG-SCHEMA.md`
  - [ ] Parse at init: `var schema = parseSchema(embeddedSchema)` (cached)
  - [ ] Validator implemented: `cmd/cicd/lint-deployments/validate-schema.go`
  - [ ] Unit tests: Valid schema, missing required keys, invalid types
  - [ ] ⚠️ **Q2 Note**: If user prefers Option E (delete CONFIG-SCHEMA.md), revise to hardcode schema in Go
  - [ ] Error messages: Moderate verbosity per Decision 14:B
  - [ ] Coverage ≥98%
  - [ ] Tests pass
- **Files**:
  - `cmd/cicd/lint-deployments/validate-schema.go` (new)
  -  `cmd/cicd/lint-deployments/validate-schema_test.go` (new)
  - `cmd/cicd/lint-deployments/embed.go` (new, for //go:embed directive)
- **Evidence**:
  - `test-output/phase3/task-3.3-schema-coverage.txt`
  - `test-output/phase3/task-3.3-q2-assumption.txt` (document Decision 10:D assumption)

#### Task 3.4: Implement ValidateTemplatePattern
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.3
- **Description**: Validate naming + structure + values per Decision 12:C
- **Acceptance Criteria**:
  - [ ] Naming validation: SERVICE/PRODUCT/SUITE filename patterns per Decision 4A
  - [ ] Structure validation: Required keys (service-name, delegation), optional keys
  - [ ] Value validation: Port offset calculations (PRODUCT=+10000, SUITE=+20000), delegation arrays, secrets paths
  - [ ] Validator implemented: `cmd/cicd/lint-deployments/validate-template-pattern.go`
  - [ ] Unit tests: All 3 validation types (naming, structure, values) with edge cases
  - [ ] Error messages: Moderate verbosity per Decision 14:B
  - [ ] Coverage ≥98%
  - [ ] Tests pass
- **Files**:
  - `cmd/cicd/lint-deployments/validate-template-pattern.go` (new)
  - `cmd/cicd/lint-deployments/validate-template-pattern_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.4-template-coverage.txt`

#### Task 3.5: Implement ValidatePorts
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.4
- **Description**: Validate port ranges, offsets, conflicts
- **Acceptance Criteria**:
  - [ ] Port range validation: SERVICE (8000-8999), PRODUCT (+10000), SUITE (+20000)
  - [ ] Port conflict detection: No two services on same port
  - [ ] Validator implemented:  `cmd/cicd/lint-deployments/validate-ports.go`
  - [ ] Unit tests: Valid ranges, offset calculations, conflicts
  - [ ] Error messages: Moderate verbosity per Decision 14:B
  - [ ] Coverage ≥98%
  - [ ] Tests pass
- **Files**:
  - `cmd/cicd/lint-deployments/validate-ports.go` (new)
  - `cmd/cicd/lint-deployments/validate-ports_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.5-ports-coverage.txt`

#### Task 3.6: Implement ValidateTelemetry
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.5
- **Description**: Validate OTLP endpoints and sidecar configuration
- **Acceptance Criteria**:
  - [ ] OTLP endpoint validation: Check protocol (grpc/http), host, port
  - [ ] Sidecar validation: otel-collector-contrib forwarding pattern
  - [ ] Validator implemented: `cmd/cicd/lint-deployments/validate-telemetry.go`
  - [ ] Unit tests: Valid OTLP, missing endpoints, wrong protocols
  - [ ] Error messages: Moderate verbosity per Decision 14:B
  - [ ] Coverage ≥98%
  - [ ] Tests pass
- **Files**:
  - `cmd/cicd/lint-deployments/validate-telemetry.go` (new)
  - `cmd/cicd/lint-deployments/validate-telemetry_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.6-telemetry-coverage.txt`

#### Task 3.7: Implement ValidateAdmin
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.6
- **Description**: Validate admin API patterns (127.0.0.1:9090 binding)
- **Acceptance Criteria**:
  - [ ] Admin binding validation: Must be 127.0.0.1 (NOT 0.0.0.0)
  - [ ] Admin port validation: Default 9090, no conflicts
  - [ ] Validator implemented: `cmd/cicd/lint-deployments/validate-admin.go`
  - [ ] Unit tests: Valid binding (127.0.0.1:9090), invalid binding (0.0.0.0), port conflicts
  - [ ] Error messages: Moderate verbosity per Decision 14:B
  - [ ] Coverage ≥98%
  - [ ] Tests pass
- **Files**:
  - `cmd/cicd/lint-deployments/validate-admin.go` (new)
  - `cmd/cicd/lint-deployments/validate-admin_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.7-admin-coverage.txt`

#### Task 3.8: Implement ValidateSecrets
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.7
- **Description**: Aggressive secrets detection with entropy analysis per Decision 15:C
- **Acceptance Criteria**:
  - [ ] Layer 1: Secrets files (.secret extensions, environment: *_FILE patterns)
  - [ ] Layer 2: Pattern matching (AWS keys, GitHub tokens, known patterns)
  - [ ] Layer 3: Entropy analysis (Shannon entropy >4.5 bits/char for inline strings)
  - [ ] Validator implemented: `cmd/cicd/lint-deployments/validate-secrets.go`
  - [ ] Unit tests: Secrets files, pattern matches, high-entropy strings, false positives (UUIDs, base64 data)
  - [ ] Error messages: Moderate verbosity per Decision 14:B
  - [ ] Coverage ≥98%
  - [ ] Tests pass
- **Files**:
  - `cmd/cicd/lint-deployments/validate-secrets.go` (new)
  - `cmd/cicd/lint-deployments/validate-secrets_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.8-secrets-coverage.txt`
  - `test-output/phase3/task-3.8-entropy-tests.log` (false positive analysis)

#### Task 3.9: Pre-Commit Integration (Parallel Validators)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.8
- **Description**: Integrate validators into pre-commit with parallel execution per Decision 11:E
- **Acceptance Criteria**:
  - [ ] Parallel execution: Run 8 validators concurrently (use goroutines + errgroup)
  - [ ] Target performance: <5s for incremental validation (50+ files)
  - [ ] Pre-commit hook: `.pre-commit-config.yaml` entry for cicd lint-deployments
  - [ ] Integration tests: Verify parallel execution (no race conditions)
  - [ ] Race detector clean: `go test -race ./cmd/cicd/lint-deployments/...`
  - [ ] Functional test: Run pre-commit locally, verify <5s
- **Files**:
  - `.pre-commit-config.yaml` (updated)
  - `cmd/cicd/lint-deployments/parallel.go` (new, goroutine orchestration)
- **Evidence**:
  - `test-output/phase3/task-3.9-precommit-timing.log` (performance measurement)

#### Task 3.10: Mutation Testing (ALL Validators)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.9
- **Description**: Mutation testing for ALL 8 validators, ≥98% score per Decision 17:A (NO exemptions)
- **Acceptance Criteria**:
  - [ ] Mutation testing: Run `gremlins unleash --tags=!integration ./cmd/cicd/lint-deployments/`
  - [ ] ALL validators ≥98%: ValidateNaming, ValidateKebabCase, ValidateSchema, ValidateTemplatePattern, ValidatePorts, ValidateTelemetry, ValidateAdmin, ValidateSecrets
  - [ ] Edge case tests added: Regex escaping, boundary conditions, special characters
  - [ ] NO exemptions per Decision 17:A (even "trivial" validators must achieve ≥98%)
  - [ ] Mutation report: `test-output/phase3/mutation-report.txt`
- **Files**:
  - Various `*_test.go` files (edge case tests added)
- **Evidence**:
  - `test-output/phase3/task-3.10-mutation-report.txt` (gremlins output)
  - `test-output/phase3/task-3.10-all-validators-98percent.txt` (verification)

#### Task 3.11: Performance Benchmarks
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.10
- **Description**: Benchmark individual validators and total validation time per Decision 11:E
- **Acceptance Criteria**:
  - [ ] Benchmark tests: `*_bench_test.go` for each validator
  - [ ] Metrics: Time per file, total time for 50+ files, memory allocation
  - [ ] Target: <5s total for incremental validation (parallel)
  - [ ] Baseline established: Document current performance
  - [ ] Optimization candidates: Identify slow validators for future optimization
- **Files**:
  - `cmd/cicd/lint-deployments/*_bench_test.go` (new, 8 files)
- **Evidence**:
  - `test-output/phase3/task-3.11-benchmark-results.txt`
  - `test-output/phase3/task-3.11-performance-baseline.txt`

#### Task 3.12: Validation Caching (Priority 1)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.11
- **Description**: File hash-based caching to skip validating unchanged files
- **Acceptance Criteria**:
  - [ ] Cache implementation: Hash file content, cache validation result
  - [ ] Cache invalidation: Rehash on file change
  - [ ] Integration: Pre-commit uses cache (skip unchanged files)
  - [ ] Performance improvement: Measure cache hit rate, time savings (~500ms per cached file)
  - [ ] Tests: Verify cache correctness (no false cache hits)
- **Files**:
  - `cmd/cicd/lint-deployments/cache.go` (new)
  - `cmd/cicd/lint-deployments/cache_test.go` (new)
- **Evidence**:
  - `test-output/phase3/task-3.12-cache-performance.txt`

#### Task 3.13: Phase 3 E2E Verification
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.12
- **Description**: Verify all 8 validators implemented with quality gates met
- **Acceptance Criteria**:
  - [ ] 8 validators exist: validate-{naming,kebab-case,schema,template-pattern,ports,telemetry,admin,secrets}
  - [ ] Coverage ≥98% for ALL validators
  - [ ] Mutation ≥98% for ALL validators (NO exemptions per Decision 17:A)
  - [ ] Pre-commit functional (parallel, <5s incremental)
  - [ ] All tests pass: `go test ./cmd/cicd/lint-deployments/...`
  - [ ] Race detector clean: `go test -race ./cmd/cicd/lint-deployments/...`
- **Files**:
  - N/A (verification only)
- **Evidence**:
  - `test-output/phase3/phase3-completion-verification.log`
  - `test-output/phase3/coverage-summary.txt` (all ≥98%)
  - `test-output/phase3/mutation-summary.txt` (all ≥98%)

---

### Phase 4: ARCHITECTURE.md Updates (6h)

**Phase Objective**: Add minimal ARCHITECTURE.md sections 12.4-12.6 with ASCII diagrams (reduced from 8h due to Decision 9:A)

#### Task 4.1: Add Section 12.4 - Deployment Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 3 complete (Task 3.13)
- **Description**: Minimal overview of 8 validators with ASCII diagram per Decision 9:A, Decision 16:B
- **Acceptance Criteria**:
  - [ ] Section added: `docs/ARCHITECTURE.md` Section 12.4
  - [ ] Content: Brief overview (1-2 paragraphs) of 8 validator types per Decision 9:A
  - [ ] ASCII diagram: Validation flow (configs → validators → pass/fail) per Decision 16:B
  - [ ] Defer details: Reference ValidateXXX implementations (no comprehensive docs)
  - [ ] Cross-references: Link to Section 12.5 (Config File Architecture)
  - [ ] Minimal depth: ~20-30 lines total (not comprehensive per Decision 9:A)
- **Files**:
  - `docs/ARCHITECTURE.md` (updated, Section 12.4 added)
- **Evidence**:
  - `test-output/phase4/task-4.1-section-12.4-content.txt`

#### Task 4.2: Add Section 12.5 - Config File Architecture
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.1
- **Description**: Minimal overview of SERVICE/PRODUCT/SUITE hierarchy with ASCII diagram per Decision 9:A, Decision 16:B
- **Acceptance Criteria**:
  - [ ] Section added: `docs/ARCHITECTURE.md` Section 12.5
  - [ ] Content: Brief hierarchy explanation (1-2 paragraphs) per Decision 9:A
  - [ ] ASCII diagram: Config delegation tree (SUITE → PRODUCT → SERVICE) per Decision 16:B
  - [ ] Template pattern reference: Link to Decision 4A (concrete rules)
  - [ ] Cross-references: Link to Section 12.4 (Deployment Validation), Section 12.6 (Secrets Management)
  - [ ] Minimal depth: ~20-30 lines total
- **Files**:
  - `docs/ARCHITECTURE.md` (updated, Section 12.5 added)
- **Evidence**:
  - `test-output/phase4/task-4.2-section-12.5-content.txt`

#### Task 4.3: Add Section 12.6 - Secrets Management
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.2
- **Description**: Minimal overview of Docker secrets priority and validation per Decision 9:A
- **Acceptance Criteria**:
  - [ ] Section added: `docs/ARCHITECTURE.md` Section 12.6
  - [ ] Content: Brief secrets priority (Docker secrets > YAML > CLI) per Decision 9:A
  - [ ] Content: Validation approach mention (aggressive detection with entropy per Decision 15:C)
  - [ ] Defer details: Reference ValidateSecrets implementation
  - [ ] Cross-references: Link to Section 12.4 (Deployment Validation)
  - [ ] Minimal depth: ~15-20 lines total
- **Files**:
  - `docs/ARCHITECTURE.md` (updated, Section 12.6 added)
- **Evidence**:
  - `test-output/phase4/task-4.3-section-12.6-content.txt`

#### Task 4.4: Add Cross-References
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.3
- **Description**: Link sections 12.4-12.6 to existing ARCHITECTURE.md sections
- **Acceptance Criteria**:
  - [ ] Cross-reference: Section 12.4 ← link from existing sections mentioning validation
  - [ ] Cross-reference: Section 12.5 ← link from existing sections mentioning configs/deployments
  - [ ] Cross-reference: Section 12.6 ← link from existing sections mentioning secrets
  - [ ] Bidirectional links: Section 12.4/12.5/12.6 link to relevant existing sections
  - [ ] Link validation: All section numbers correct (Task 4.5 will automate)
- **Files**:
  - `docs/ARCHITECTURE.md` (updated, cross-references added)
- **Evidence**:
  - `test-output/phase4/task-4.4-cross-references-list.txt`

#### Task 4.5: Cross-Reference Validation Tool (Priority 1)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.4
- **Description**: Automated tool to verify ARCHITECTURE.md section number consistency
- **Acceptance Criteria**:
  - [ ] Tool implemented: `cmd/cicd/check-arch-cross-refs`
  - [ ] Validation: All section references exist (e.g., "Section 12.4" → actual Section 12.4 present)
  - [ ] Validation: No broken links (orphaned references)
  - [ ] Output: List of broken references (if any)
  - [ ] Integration: Can be added to pre-commit or CI/CD
  - [ ] Tests: Unit tests for tool logic
- **Files**:
  - `cmd/cicd/check-arch-cross-refs/main.go` (new)
  - `cmd/cicd/check-arch-cross-refs/main_test.go` (new)
- **Evidence**:
  - `test-output/phase4/task-4.5-cross-ref-validation.log`

#### Task 4.6: Phase 4 E2E Verification
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.5
- **Description**: Verify ARCHITECTURE.md sections 12.4-12.6 complete with minimal depth per Decision 9:A
- **Acceptance Criteria**:
  - [ ] Sections exist: 12.4, 12.5, 12.6 in `docs/ARCHITECTURE.md`
  - [ ] ASCII diagrams present: Validation flow, config hierarchy (NOT Mermaid per Decision 16:B)
  - [ ] Minimal depth: Each section ~15-30 lines (not comprehensive)
  - [ ] Cross-references valid: Task 4.5 tool passes (no broken links)
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - N/A (verification only)
- **Evidence**:
  - `test-output/phase4/phase4-completion-verification.log`
  - `test-output/phase4/section-lengths.txt` (verify minimal depth)

---

### Phase 5: Instruction File Propagation (7h)

**Phase Objective**: Propagate ARCHITECTURE.md patterns via chunk-based verbatim copying per Decision 13:E (updated from 6h)

#### Task 5.1: Update 04-01.deployment.instructions.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 4 complete (Task 4.6)
- **Description**: Copy ARCHITECTURE.md Section 12.4 chunks verbatim per Decision 13:E
- **Acceptance Criteria**:
  - [ ] Chunks copied verbatim: Extract Section 12.4 text/diagrams from ARCHITECTURE.md, paste into 04-01.deployment.instructions.md
  - [ ] NO paraphrasing: Text must match ARCHITECTURE.md exactly per Decision 13:E
  - [ ] Content: 8 validator overview, validation flow diagram
  - [ ] Cross-reference: Link to ARCHITECTURE.md Section 12.4 (for full context)
  - [ ] Chunk boundaries: Clearly mark ARCHITECTURE.md source sections
- **Files**:
  - `.github/instructions/04-01.deployment.instructions.md` (updated)
- **Evidence**:
  - `test-output/phase5/task-5.1-deployment-instructions-chunks.txt` (copied chunks)

#### Task 5.2: Update 02-01.architecture.instructions.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.1
- **Description**: Copy ARCHITECTURE.md Section 12.5 chunks verbatim per Decision 13:E
- **Acceptance Criteria**:
  - [ ] Chunks copied verbatim: Extract Section 12.5 text/diagrams from ARCHITECTURE.md, paste into 02-01.architecture.instructions.md
  - [ ] NO paraphrasing: Text must match ARCHITECTURE.md exactly
  - [ ] Content: SERVICE/PRODUCT/SUITE hierarchy, config delegation diagram
  - [ ] Cross-reference: Link to ARCHITECTURE.md Section 12.5
  - [ ] Chunk boundaries: Clearly mark ARCHITECTURE.md source sections
- **Files**:
  - `.github/instructions/02-01.architecture.instructions.md` (updated)
- **Evidence**:
  - `test-output/phase5/task-5.2-architecture-instructions-chunks.txt`

#### Task 5.3: Chunk-Based Verification Tool
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.2
- **Description**: Implement tool to verify ARCHITECTURE.md chunks present in instruction files per Decision 13:E
- **Acceptance Criteria**:
  - [ ] Tool implemented: `cmd/cicd/check-chunk-propagation`
  - [ ] Extraction: Parse ARCHITECTURE.md, identify chunk boundaries (subsections, code blocks, diagrams)
  - [ ] Verification: Search for exact chunk match in instruction files (verbatim copying required)
  - [ ] Output: List of missing/mismatched chunks
  - [ ] Integration: Can be added to pre-commit or CI/CD
  - [ ] Tests: Unit tests for chunk extraction + matching logic
- **Files**:
  - `cmd/cicd/check-chunk-propagation/main.go` (new)
  - `cmd/cicd/check-chunk-propagation/main_test.go` (new)
- **Evidence**:
  - `test-output/phase5/task-5.3-chunk-verification.log`

#### Task 5.4: Doc Consistency Check (Priority 1)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.3
- **Description**: Checklist-based tool for systematic propagation verification
- **Acceptance Criteria**:
  - [ ] Tool implemented: `cmd/cicd/check-doc-consistency`
  - [ ] Checklist: Pre-defined list of patterns to verify in each instruction file
  - [ ] Example checks: "04-01.deployment.instructions.md contains ValidateNaming docs", "02-01.architecture.instructions.md contains config hierarchy"
  - [ ] Output: Checklist results (pass/fail per item)
  - [ ] Integration: Can be run manually or in CI/CD
  - [ ] Tests: Unit tests for checklist logic
- **Files**:
  - `cmd/cicd/check-doc-consistency/main.go` (new)
  - `cmd/cicd/check-doc-consistency/main_test.go` (new)
- **Evidence**:
  - `test-output/phase5/task-5.4-consistency-check.log`

#### Task 5.5: Phase 5 E2E Verification
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.4
- **Description**: Verify chunk-based propagation complete per Decision 13:E
- **Acceptance Criteria**:
  - [ ] ALL chunks present: Task 5.3 tool passes (no missing chunks)
  - [ ] Consistency check: Task 5.4 tool passes (all checklist items verified)
  - [ ] Instruction files updated: 04-01.deployment.instructions.md, 02-01.architecture.instructions.md have verbatim chunks
  - [ ] Tests pass: `go test ./...`
- **Files**:
  - N/A (verification only)
- **Evidence**:
  - `test-output/phase5/phase5-completion-verification.log`
  - `test-output/phase5/chunk-propagation-results.txt`

---

### Phase 6: E2E Validation (3h)

**Phase Objective**: Validate 100% pass rate for ALL configs + deployments

#### Task 6.1: Validate ALL configs/ Files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 5 complete (Task 5.5)
- **Description**: Run cicd lint-deployments against ALL configs/ files, verify 100% pass
- **Acceptance Criteria**:
  - [ ] Command: `cicd lint-deployments validate-all configs/`
  - [ ] Result: 100% pass (zero failures) for ALL 15 config dirs (9 SERVICE + 5 PRODUCT + 1 SUITE)
  - [ ] All 8 validators pass: naming, kebab-case, schema, template-pattern, ports, telemetry, admin, secrets
  - [ ] No false positives: Review any warnings/errors, confirm legitimate issues
  - [ ] Evidence collected: Validation output, pass/fail counts
- **Files**:
  - N/A (validation only)
- **Evidence**:
  - `test-output/phase6/task-6.1-configs-validation.log`
  - `test-output/phase6/task-6.1-pass-rate.txt` (must be 100%)

#### Task 6.2: Validate ALL deployments/ Files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.1
- **Description**: Run cicd lint-deployments against ALL deployments/ files, verify 100% pass
- **Acceptance Criteria**:
  - [ ] Command: `cicd lint-deployments validate-all deployments/`
  - [ ] Result: 100% pass (zero failures) for ALL deployment files
  - [ ] All 8 validators pass
  - [ ] No false positives: Review any warnings/errors
  - [ ] Evidence collected: Validation output, pass/fail counts
- **Files**:
  - N/A (validation only)
- **Evidence**:
  - `test-output/phase6/task-6.2-deployments-validation.log`
  - `test-output/phase6/task-6.2-pass-rate.txt` (must be 100%)

#### Task 6.3: Test Sample Violations
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.2
- **Description**: Verify validators detect intentional violations (8 validator types)
- **Acceptance Criteria**:
  - [ ] Sample violations: Create temporary config files with intentional errors (wrong naming, wrong ports, inline secrets, etc.)
  - [ ] Detection: Run validators, verify ALL 8 types detect their respective violations
  - [ ] Error messages: Verify moderate verbosity per Decision 14:B (issue, fix, file/line)
  - [ ] Cleanup: Delete sample violation files after testing
  - [ ] Evidence: Log of detected violations (8/8 validator types)
- **Files**:
  - Temporary test files (deleted after testing)
- **Evidence**:
  - `test-output/phase6/task-6.3-sample-violations.log`
  - `test-output/phase6/task-6.3-detection-rate.txt` (8/8 detected)

#### Task 6.4: Verify Pre-Commit Integration
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.3
- **Description**: Functional test of pre-commit hooks with parallel validators per Decision 11:E
- **Acceptance Criteria**:
  - [ ] Pre-commit installed: `pre-commit install`
  - [ ] Functional test: Stage config files, run `pre-commit run`
  - [ ] Performance: <5s for incremental validation (50+ files, parallel execution)
  - [ ] Validators run: All 8 validators executed in parallel
  - [ ] Evidence: Timing log, validator output
- **Files**:
  - N/A (functional test only)
- **Evidence**:
  - `test-output/phase6/task-6.4-precommit-timing.log` (must be <5s)
  - `test-output/phase6/task-6.4-parallel-execution.log`

#### Task 6.5: Phase 6 E2E Verification
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.4
- **Description**: Comprehensive E2E verification of entire implementation
- **Acceptance Criteria**:
  - [ ] ALL configs pass: 100% validation (Task 6.1)
  - [ ] ALL deployments pass: 100% validation (Task 6.2)
  - [ ] Sample violations detected: 8/8 validator types (Task 6.3)
  - [ ] Pre-commit functional: <5s parallel execution (Task 6.4)
  - [ ] All tests pass: `go test ./...`
  - [ ] Coverage ≥98%: `cmd/cicd/` infrastructure code
  - [ ] Mutation ≥98%: ALL validators (NO exemptions per Decision 17:A)
  - [ ] Race detector clean: `go test -race ./...`
  - [ ] Linting clean: `golangci-lint run ./...`
- **Files**:
  - N/A (verification only)
- **Evidence**:
  - `test-output/phase6/phase6-completion-verification.log`
  - `test-output/phase6/final-quality-gates.txt` (all gates passed)

---

## Cross-Cutting Tasks

### Testing
- [ ] Unit tests ≥98% coverage (cmd/cicd/ infrastructure code per copilot instructions)
- [ ] Integration tests pass (cross-validator interactions)
- [ ] E2E tests pass: 100% validation pass rate (Phase 6)
- [ ] Mutation testing ≥98% for ALL validators (NO exemptions per Decision 17:A)
- [ ] No skipped tests (except documented exceptions)
- [ ] Race detector clean: `go test -race ./...`

### Code Quality
- [ ] Linting passes: `golangci-lint run ./...`
- [ ] No new TODOs without tracking in tasks.md
- [ ] No security vulnerabilities: `gosec ./...`
- [ ] Formatting clean: `gofumpt -s -w ./`
- [ ] Imports organized: `goimports -w ./`

### Documentation
- [ ] ARCHITECTURE.md updated: Sections 12.4-12.6 (minimal depth per Decision 9:A)
- [ ] Instruction files updated: Chunk-based verbatim copying per Decision 13:E (04-01.deployment.instructions.md, 02-01.architecture.instructions.md)
- [ ] CONFIG-SCHEMA.md: Embedded in ValidateSchema per Decision 10:D (or deleted if user prefers Option E)
- [ ] README.md files: Minimal content per Decision 8:A (configs/, PRODUCT/, SUITE/)
- [ ] Comments: Moderate detail in validator code (primary docs per Decision 9:A)

### Deployment
- [ ] Pre-commit functional: Parallel validators, <5s incremental per Decision 11:E
- [ ] CI/CD integration: Optional (can add GitHub Actions workflow)
- [ ] Config files validated: 100% pass rate
- [ ] Deployment files validated: 100% pass rate

---

## Notes / Deferred Work

### Quizme-v2 Q2 Clarification Needed
- **Issue**: Q2 (CONFIG-SCHEMA.md integration) was blank (unanswered)
- **Assumption**: Using Decision 10:D (Embed + parse at init) as default
- **Alternative**: User may prefer Decision 10:E (Delete CONFIG-SCHEMA.md, hardcode schema)
- **Action**: User should confirm preference. If Option E preferred, revise Task 3.3 to delete CONFIG-SCHEMA.md and hardcode schema in `validate-schema.go`

### Priority 2 Improvements (Deferred to Future Iteration)
From ANALYSIS.md, these were identified but NOT added:
- Task 1.0: Config backup before restructure (rollback safety)
- Task 1.7A: Migration script (automate 50+ file moves)
- Task 2.0: PRODUCT config generation tool (reduce manual errors)
- Task 3.13A: CI/CD GitHub Actions workflow (automated validation on PR)
- Task 3.8A: Enhanced secrets detection (more entropy heuristics)

These can be added in quizme-v3 if user wants even more rigor.

---

## Evidence Archive

**Current Documentation**:
- `docs/fixes-v3/plan.md` - THIS FILE (implementation plan)
- `docs/fixes-v3/tasks.md` - Implementation plan with quizme-v2 answers integrated (57 tasks, 57h LOE)
- `docs/fixes-v3/quizme-v1.md` - DELETED (merged into Decisions 1-8)
- `docs/fixes-v3/quizme-v2.md` - TO BE DELETED (will merge into Decisions 9-18 after user confirmation of Q2)
- `docs/fixes-v3/ANALYSIS.md` - Deep analysis (15 improvements, Priority 1 applied)
- `docs/fixes-v3/COMPLETION-STATUS.md` - V2 completion summary

**Planning Evidence**:
- `test-output/fixes-v3-quizme-analysis/` - Quizme-v1 answers + analysis
- `test-output/fixes-v3-quizme-v2-analysis/` - Quizme-v2 answers + analysis

**Implementation Evidence** (to be created during Phase 1-6):
- `test-output/phase1/` - Phase 1 restructuring logs (8 tasks)
- `test-output/phase2/` - Phase 2 template generation logs (7 tasks)
- `test-output/phase3/` - Phase 3 validator implementation + coverage + mutation (13 tasks)
- `test-output/phase4/` - Phase 4 ARCHITECTURE.md updates (6 tasks)
- `test-output/phase5/` - Phase 5 propagation verification logs (5 tasks)
- `test-output/phase6/` - Phase 6 E2E validation results (5 tasks)

