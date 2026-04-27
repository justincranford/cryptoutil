# Tasks - Framework V19: Prescriptive MANIFEST.yaml + Identity Conformance Migration

**Status**: 0 of 46 tasks complete (0%)
**Last Updated**: 2026-04-27
**Created**: 2026-04-27

## Quality Mandate — MANDATORY

| Attribute | Requirement |
|-----------|-------------|
| Correctness | ALL code functionally correct; comprehensive tests |
| Completeness | NO phases/tasks/steps skipped; NO shortcuts |
| Thoroughness | Evidence-based validation at every step |
| Reliability | ≥98% infrastructure/utility coverage; ≥98% mutation (cicd_lint packages) |
| Efficiency | Optimized for maintainability; NOT implementation speed |
| Accuracy | Root cause addressed; not just symptoms |
| NO Time Pressure | NEVER rush; NEVER skip validation; NEVER defer quality checks |
| NO Premature Completion | Objective evidence required before marking complete |

**ALL issues are blockers.** Fix immediately. NEVER defer.

---

## Task Status Legend — MANDATORY

| Symbol | Meaning | When to Use |
|--------|---------|-------------|
| ❌ | Not started | Task not yet begun |
| 🔄 | In progress | Currently being worked on |
| ✅ | Complete | Task finished with evidence |
| ⏳ | Blocked | Requires external dependency (MUST have resolution plan) |

---

## Phase 0: Pre-flight Build Health

**Phase Objective**: Verify clean baseline before any V19 changes.

### Task 0.1: Build Health Pre-flight

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0
  - [ ] `go build -tags e2e,integration ./...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0 (all 87 linters pass)
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/...` exits 0
  - [ ] Output archived in `test-output/v19-phase0/build-health.txt`
- **Files**: None (verification only)

---

## Phase 1: Prescriptive MANIFEST.yaml + Linter Extension

**Phase Objective**: Expand **PS_ID** MANIFEST.yaml from 3 shallow fields to a fully
recursive structure spec. Extend linter to validate all new fields.

### Task 1.1: Update **PS_ID** MANIFEST.yaml

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 0.1
- **Description**: Replace sparse MANIFEST.yaml with fully prescriptive version including:
  client/ and e2e/ in required_dirs; required_server_dirs; required_server_config_files;
  required_server_repository_files; required_server_repository_dirs; required_e2e_files.
  Each field annotated with purpose comments, package patterns, exception rationale.
- **Acceptance Criteria**:
  - [ ] required_dirs includes client and e2e (with knownExclusions documented in comments)
  - [ ] required_server_dirs: [apis, config, model, repository] with exception notes
  - [ ] required_server_config_files: [config.go, config_test.go, config_test_helper.go]
  - [ ] required_server_repository_files: [migrations.go]
  - [ ] required_server_repository_dirs: [migrations]
  - [ ] required_e2e_files: [testmain_e2e_test.go, __SERVICE___e2e_test.go]
  - [ ] All fields have YAML comments explaining purpose and exceptions
  - [ ] File encodes UTF-8 without BOM (verified with `lint-text`)
- **Files**:
  - `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml`

### Task 1.2: Extend psIDManifest Struct

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Add 5 new fields to the `psIDManifest` struct in apps_ps_id_template.go
  and 5 corresponding knownExclusions maps.
- **Acceptance Criteria**:
  - [ ] `RequiredServerDirs []string` field added (yaml:"required_server_dirs")
  - [ ] `RequiredServerConfigFiles []string` field added (yaml:"required_server_config_files")
  - [ ] `RequiredServerRepositoryFiles []string` field added
  - [ ] `RequiredServerRepositoryDirs []string` field added
  - [ ] `RequiredE2EFiles []string` field added (yaml:"required_e2e_files")
  - [ ] 5 new knownExclusions maps registered with correct PS-ID sets (see plan.md Phase 1a table)
  - [ ] `go build ./internal/apps-tools/...` exits 0
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template.go`

### Task 1.3: Implement 5 New Check Functions

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.2
- **Description**: Implement check functions for each new MANIFEST field:
  checkServerDirs, checkServerConfigFiles, checkServerRepositoryFiles,
  checkServerRepositoryDirs, checkE2EFiles. Each respects its knownExclusions map.
  checkE2EFiles applies **SERVICE** substitution before checking.
- **Acceptance Criteria**:
  - [ ] `checkServerDirs` verifies `server/{dir}` for each RequiredServerDirs entry
  - [ ] `checkServerConfigFiles` verifies `server/config/{file}` for each entry
  - [ ] `checkServerRepositoryFiles` verifies `server/repository/{file}` for each entry
  - [ ] `checkServerRepositoryDirs` verifies `server/repository/{dir}` for each entry
  - [ ] `checkE2EFiles` verifies `e2e/{file}` with **SERVICE** → actual service name substitution
  - [ ] All 5 functions integrated into Lint() main flow
  - [ ] `go build ./internal/apps-tools/...` exits 0
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template.go`

### Task 1.4: Write Tests for New Check Functions

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.3
- **Description**: Add table-driven tests covering: (a) missing required_server_dir error,
  (b) missing required_server_config_file error, (c) missing required_server_repository_file,
  (d) missing e2e file, (e) correct exclusion logic (excluded PS-ID does not trigger error),
  (f) **SERVICE** substitution in required_e2e_files.
- **Acceptance Criteria**:
  - [ ] ≥6 new test cases covering all 5 check functions
  - [ ] Tests for both happy path and error path per function
  - [ ] Exclusion map logic tested (excluded PS-ID = no error; non-excluded = error)
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/...` exits 0
  - [ ] Coverage ≥98% for apps_ps_id_template package
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template_test.go`

### Task 1.5: Validate MANIFEST + Linter Passes Current Codebase

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.4
- **Description**: Run lint-fitness against current codebase with new linter active. ALL
  violations must be in knownExclusions (no unexpected failures). Verify exclusion counts
  match survey data in plan.md.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] No unexpected linter violations (all failures are in knownExclusions)
  - [ ] `go run ./cmd/cicd-lint lint-fitness` output reviewed; knownExclusions count matches plan
  - [ ] `golangci-lint run ./internal/apps-tools/...` exits 0
- **Files**: None (verification only)

---

## Phase 2: Identity Services Server Code Migration

**Phase Objective**: Move domain code from PS-ID root → server/ for all 5 identity services.
Create lifecycle_test.go and port_conflict_test.go for each.

### Task 2.1: identity-authz — Move swagger.go + service files to server/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.5
- **Description**: First batch: move swagger.go, swagger_test.go, service.go, service_lifecycle_test.go,
  service_test.go from root to server/. Update package declarations from `identity_authz` →
  `server`. Update all intra-package imports in test files.
- **Acceptance Criteria**:
  - [ ] swagger.go, swagger_test.go now in server/
  - [ ] service.go, service_lifecycle_test.go, service_test.go now in server/
  - [ ] Package declarations updated (identity_authz → server;_test packages updated)
  - [ ] `go build ./internal/apps/identity-authz/...` exits 0
  - [ ] `go test ./internal/apps/identity-authz/...` exits 0
- **Files**: see plan.md Phase 2a

### Task 2.2: identity-authz — Move handlers + domain files to server/apis/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.1
- **Description**: Move all handlers_*.go files to server/apis/ as `package apis`. Move
  remaining domain files (authorization_request.go, cleanup.go, client_authentication.go,
  code_generator.go, device_code_generator.go, dpop/, middleware.go, pkce/,
  request_uri_generator.go, routes.go, authz_test.go, performance_bench_test.go,
  test_helpers_test.go) to server/.
- **Acceptance Criteria**:
  - [ ] All handlers_*.go files in server/apis/ as package apis
  - [ ] All domain .go files in server/ root (not root of identity-authz)
  - [ ] server/apis/ package compiles
  - [ ] `go build ./internal/apps/identity-authz/...` exits 0
  - [ ] `go test ./internal/apps/identity-authz/...` exits 0
- **Files**: see plan.md Phase 2a

### Task 2.3: identity-authz — Create lifecycle + port_conflict tests in server/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.2
- **Description**: Create `server/authz_lifecycle_test.go` and `server/authz_port_conflict_test.go`
  using sm-kms or jose-ja as reference patterns.
- **Acceptance Criteria**:
  - [ ] authz_lifecycle_test.go exists in server/ with TestAuthz_DualPortStartup and TestAuthz_GracefulShutdown
  - [ ] authz_port_conflict_test.go exists in server/ with TestAuthz_PortConflict
  - [ ] `go test ./internal/apps/identity-authz/server/...` passes
  - [ ] identity-authz root has ONLY: authz.go, authz_usage.go, authz_cli_test.go, authz_contract_test.go
- **Files**:
  - `internal/apps/identity-authz/server/authz_lifecycle_test.go`
  - `internal/apps/identity-authz/server/authz_port_conflict_test.go`

### Task 2.4: identity-idp — Move swagger.go + service files to server/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.3
- **Description**: Same pattern as Task 2.1 for identity-idp. Move swagger.go, swagger_test.go,
  service.go and service tests, random.go to server/.
- **Acceptance Criteria**:
  - [ ] swagger.go, swagger_test.go in server/
  - [ ] service.go, random.go + tests in server/
  - [ ] Package declarations updated
  - [ ] `go build ./internal/apps/identity-idp/...` exits 0
  - [ ] `go test ./internal/apps/identity-idp/...` exits 0

### Task 2.5: identity-idp — Move handlers + auth/ + templates/ to server/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.4
- **Description**: Move handlers_*.go to server/apis/; move auth/ subdir, templates/ subdir,
  backchannel_logout.go, client_secret.go, middleware.go, routes.go, magic_test_constants.go,
  and all remaining domain files to server/.
- **Acceptance Criteria**:
  - [ ] All handlers_*.go in server/apis/ as package apis
  - [ ] auth/, templates/ moved to server/
  - [ ] `go build ./internal/apps/identity-idp/...` exits 0
  - [ ] `go test ./internal/apps/identity-idp/...` exits 0

### Task 2.6: identity-idp — Create lifecycle + port_conflict tests

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.5
- **Acceptance Criteria**:
  - [ ] idp_lifecycle_test.go in server/
  - [ ] idp_port_conflict_test.go in server/
  - [ ] identity-idp root has ONLY: idp.go, idp_usage.go, idp_cli_test.go, idp_contract_test.go

### Task 2.7: identity-rs — Move swagger.go + service files to server/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.6
- **Description**: Move swagger.go, swagger_test.go, service.go, service_admin_test.go,
  service_test.go, validator.go from root → server/.
- **Acceptance Criteria**:
  - [ ] swagger.go, swagger_test.go in server/
  - [ ] service.go, validator.go + tests in server/
  - [ ] Package declarations updated
  - [ ] `go build ./internal/apps/identity-rs/...` exits 0
  - [ ] `go test ./internal/apps/identity-rs/...` exits 0

### Task 2.8: identity-rs — Create lifecycle + port_conflict tests

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.7
- **Acceptance Criteria**:
  - [ ] rs_lifecycle_test.go in server/
  - [ ] rs_port_conflict_test.go in server/
  - [ ] identity-rs root has ONLY: rs.go, rs_usage.go, rs_cli_test.go, rs_contract_test.go

### Task 2.9: identity-rp — Move rp_test.go + create lifecycle/port_conflict tests

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.8
- **Description**: Move rp_test.go from root → server/ (it's an HTTP handler test, not CLI).
  Create rp_lifecycle_test.go and rp_port_conflict_test.go in server/.
- **Acceptance Criteria**:
  - [ ] rp_test.go now in server/
  - [ ] rp_lifecycle_test.go in server/
  - [ ] rp_port_conflict_test.go in server/
  - [ ] identity-rp root has ONLY: rp.go, rp_usage.go, rp_cli_test.go
  - [ ] `go test ./internal/apps/identity-rp/...` exits 0

### Task 2.10: identity-spa — Move spa_test.go + create lifecycle/port_conflict tests

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.9
- **Acceptance Criteria**:
  - [ ] spa_test.go now in server/
  - [ ] spa_lifecycle_test.go in server/
  - [ ] spa_port_conflict_test.go in server/
  - [ ] identity-spa root has ONLY: spa.go, spa_usage.go, spa_cli_test.go

### Task 2.11: Phase 2 Final Build + Lint + Test

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 2.1–2.10
- **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0
  - [ ] `go test ./internal/apps/identity-.../...` exits 0 for all 5 PS-IDs
  - [ ] `golangci-lint run ./internal/apps/identity-.../...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0

---

## Phase 3: sm-im Root Cleanup

**Phase Objective**: Move all non-CLI test files from sm-im root to server/. Delete redundant testmain_test.go.

### Task 3.1: Move sm-im root test files to server/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.11
- **Description**: Move http_test.go, http_errors_test.go, response_body_test.go,
  im_database_test.go, im_server_lifecycle_test.go, im_lifecycle_test.go,
  im_port_conflict_test.go from root → server/. Delete root testmain_test.go.
- **Acceptance Criteria**:
  - [ ] All 7 files now in server/
  - [ ] testmain_test.go deleted from root (server/ copy retained)
  - [ ] sm-im root has ONLY: im.go, im_usage.go, im_cli_commands_test.go, im_cli_url_test.go
  - [ ] `go build ./internal/apps/sm-im/...` exits 0
  - [ ] `go test ./internal/apps/sm-im/...` exits 0
- **Files**: `internal/apps/sm-im/` (7 moves + 1 delete)

### Task 3.2: Phase 3 Final Validation

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.1
- **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0
  - [ ] `go test ./internal/apps/sm-im/...` exits 0
  - [ ] `golangci-lint run ./internal/apps/sm-im/...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0

---

## Phase 4: Create Missing server/ Subdirectory Packages

**Phase Objective**: Create server/apis/, server/model/, server/repository/ (+migrations/) for
5 identity services. sm-kms and pki-ca explicitly deferred to V20.

### Task 4.1: identity-authz — Create server/model/ and server/repository/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.2
- **Description**: Create server/model/ package with domain type structs. Create
  server/repository/ with migrations.go (//go:embed), migrations/ subdir, initial SQL files.
  Migration number range from registry.yaml for identity-authz.
- **Acceptance Criteria**:
  - [ ] server/model/model.go exists with at least one GORM-tagged struct
  - [ ] server/repository/migrations.go exists with //go:embed migrations/*.sql
  - [ ] server/repository/migrations/ has NNNN_init.up.sql and NNNN_init.down.sql
  - [ ] Migration number in correct range per registry.yaml
  - [ ] `go build ./internal/apps/identity-authz/...` exits 0
- **Files**: see plan.md Phase 4

### Task 4.2: identity-idp — Create server/model/ and server/repository/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.1
- **Acceptance Criteria** (same pattern as 4.1 for identity-idp):
  - [ ] server/model/ with domain types
  - [ ] server/repository/ with migrations.go + migrations/ subdir
  - [ ] Migration number in identity-idp range from registry.yaml
  - [ ] `go build ./internal/apps/identity-idp/...` exits 0

### Task 4.3: identity-rs — Create server/model/ and server/repository/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.2
- **Acceptance Criteria** (same pattern for identity-rs):
  - [ ] server/model/ + server/repository/ with migrations
  - [ ] `go build ./internal/apps/identity-rs/...` exits 0

### Task 4.4: identity-rp — Create server/apis/, server/model/, server/repository/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.3
- **Acceptance Criteria**:
  - [ ] server/apis/ with minimal handler stub
  - [ ] server/model/ with domain types
  - [ ] server/repository/ with migrations
  - [ ] `go build ./internal/apps/identity-rp/...` exits 0

### Task 4.5: identity-spa — Create server/apis/, server/model/, server/repository/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.4
- **Acceptance Criteria** (same pattern for identity-spa):
  - [ ] server/apis/ + server/model/ + server/repository/
  - [ ] `go build ./internal/apps/identity-spa/...` exits 0

### Task 4.6: Phase 4 Validation — lint-fitness knownExclusions reduced

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.5
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] Required_server_dirs knownExclusions for identity services removed (only sm-kms/pki-ca remain)
  - [ ] `go build ./...` exits 0

---

## Phase 5: Create Missing client/ Packages

**Phase Objective**: Create client/ package for 8 PS-IDs currently missing it.

### Task 5.1: Create client/ for jose-ja

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.6
- **Description**: Create `internal/apps/jose-ja/client/client.go` — typed HTTP client wrapper
  for JWK Authority API (GetJWKS, CreateJWK, RotateJWK). Uses oapi-generated client internally.
- **Acceptance Criteria**:
  - [ ] client/client.go exists with package client
  - [ ] At least 2 typed API methods
  - [ ] `go build ./internal/apps/jose-ja/...` exits 0

### Task 5.2: Create client/ for pki-ca

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.1
- **Description**: Create `internal/apps/pki-ca/client/client.go` — typed client for CA API.
- **Acceptance Criteria**:
  - [ ] client/client.go exists with package client
  - [ ] `go build ./internal/apps/pki-ca/...` exits 0

### Task 5.3: Create client/ for identity-authz

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.2

### Task 5.4: Create client/ for identity-idp

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.3

### Task 5.5: Create client/ for identity-rs

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.4

### Task 5.6: Create client/ for identity-rp

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.5

### Task 5.7: Create client/ for identity-spa

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.6

### Task 5.8: Create client/ for skeleton-template

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.7

### Task 5.9: Phase 5 Validation

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.8
- **Acceptance Criteria**:
  - [ ] All 10 PS-IDs have client/ directory
  - [ ] `go build ./...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] required_dirs: client knownExclusions for all 8 PS-IDs removed

---

## Phase 6: Create Missing e2e/ Packages

**Phase Objective**: Create e2e/ packages for 5 PS-IDs that currently lack them.

### Task 6.1: Create e2e/ for pki-ca

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.9
- **Description**: Create testmain_e2e_test.go (TestMain with Docker Compose) and
  ca_e2e_test.go (basic health + cert issuance smoke test). ALL files have //go:build e2e.
  Use sm-im/e2e/ or jose-ja/e2e/ as reference.
- **Acceptance Criteria**:
  - [ ] e2e/testmain_e2e_test.go with //go:build e2e; TestMain starts Compose
  - [ ] e2e/ca_e2e_test.go with //go:build e2e; at least one smoke test
  - [ ] `go build -tags e2e ./internal/apps/pki-ca/...` exits 0
- **Files**:
  - `internal/apps/pki-ca/e2e/testmain_e2e_test.go`
  - `internal/apps/pki-ca/e2e/ca_e2e_test.go`

### Task 6.2: Create e2e/ for identity-idp

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.1
- **Acceptance Criteria**:
  - [ ] testmain_e2e_test.go + idp_e2e_test.go with //go:build e2e
  - [ ] `go build -tags e2e ./internal/apps/identity-idp/...` exits 0

### Task 6.3: Create e2e/ for identity-rs

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.2
- **Acceptance Criteria**:
  - [ ] testmain_e2e_test.go + rs_e2e_test.go with //go:build e2e
  - [ ] `go build -tags e2e ./internal/apps/identity-rs/...` exits 0

### Task 6.4: Create e2e/ for identity-rp

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.3
- **Acceptance Criteria**:
  - [ ] testmain_e2e_test.go + rp_e2e_test.go with //go:build e2e

### Task 6.5: Create e2e/ for identity-spa

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.4
- **Acceptance Criteria**:
  - [ ] testmain_e2e_test.go + spa_e2e_test.go with //go:build e2e

### Task 6.6: Phase 6 Validation

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.5
- **Acceptance Criteria**:
  - [ ] All 10 PS-IDs have e2e/ directory
  - [ ] `go build -tags e2e ./...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] required_dirs: e2e knownExclusions for 5 PS-IDs removed

---

## Phase 7: Remove knownExclusions + Final Validation

**Phase Objective**: Remove temporary knownExclusions after all migrations complete. Confirm
only 3 permanent exceptions remain.

### Task 7.1: Remove temporary knownExclusions from apps_ps_id_template

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.6
- **Description**: Remove all knownExclusions entries added in Phase 1 for identity services.
  Permanent exceptions that REMAIN: sm-kms public_server.go; sm-im __SERVICE___cli_test.go;
  sm-kms + pki-ca for server_dirs [apis, model, repository].
- **Acceptance Criteria**:
  - [ ] All identity service exclusions removed from all knownExclusions maps
  - [ ] 3 permanent exceptions documented with comments: "PERMANENT — legacy sm-kms", etc.
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0

### Task 7.2: Full Quality Gate Run

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.1
- **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0
  - [ ] `go build -tags e2e,integration ./...` exits 0
  - [ ] `go test ./...` exits 0 (no skips)
  - [ ] `golangci-lint run ./...` exits 0
  - [ ] `golangci-lint run --build-tags e2e,integration ./...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] `go test -race -count=2 ./...` exits 0
  - [ ] Output archived in `test-output/v19-phase7/`

---

## Phase 8: Knowledge Propagation

**Phase Objective**: Apply V19 lessons to permanent artifacts. NEVER skip.

### Task 8.1: Update ENG-HANDBOOK.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.2
- **Description**: Add V19 MANIFEST structure to §5.1 canonical PS-ID layout. Update linter
  count (87 → new count). Add migration patterns section if new patterns discovered.

### Task 8.2: Update docs/target-structure.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 8.1
- **Description**: Update G.1.2 gap matrix to reflect reduced gaps after V19. Update canonical
  PS-ID structure diagram.

### Task 8.3: Update Instruction Files and Skills

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 8.2
- **Description**: Update `.github/skills/fitness-function-gen/SKILL.md` with recursive MANIFEST
  pattern. Update instruction files where V19 work surfaces new patterns.

### Task 8.4: Verify Propagation + Final Commit

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 8.3
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
  - [ ] All changes committed with semantic commit messages
  - [ ] `git status --porcelain` returns empty

---

## Cross-Cutting Tasks

### Testing

- [ ] All new check functions tested with table-driven tests (≥6 cases per function)
- [ ] ≥98% coverage on apps_ps_id_template package
- [ ] Race detector clean: `go test -race -count=2 ./...`
- [ ] No skipped tests

### Code Quality

- [ ] `golangci-lint run ./...` exits 0 after each phase
- [ ] `golangci-lint run --build-tags e2e,integration ./...` exits 0 after each phase
- [ ] No new TODOs without tracking
- [ ] UTF-8 without BOM on all new files

---

## Notes / Deferred Work

- **sm-kms server/ subdirectory migration** (businesslogic/ → apis/model/) — DEFERRED to V20
- **pki-ca server/ subdirectory migration** (complex CA structure) — DEFERRED to V20
- **V18 ENG-HANDBOOK.md propagation plan** — independent plan in docs/framework-v18/; execute before V19 if needed for ENG-HANDBOOK.md quality

---

## Evidence Archive

- `test-output/v19-phase0/` — Phase 0 build health baseline
- `test-output/v19-phase1/` — MANIFEST + linter extension verification
- `test-output/v19-phase2/` — Identity service migration test output
- `test-output/v19-phase3/` — sm-im cleanup verification
- `test-output/v19-phase4/` — server/ subdir creation verification
- `test-output/v19-phase5/` — client/ creation verification
- `test-output/v19-phase6/` — e2e/ creation verification
- `test-output/v19-phase7/` — Full quality gate run
