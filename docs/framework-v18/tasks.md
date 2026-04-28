# Tasks Ã¢â‚¬â€ ENG-HANDBOOK.md Propagation + Prescriptive MANIFEST + Identity Conformance Migration

**Status**: 49 of 49 tasks complete (100%)
**Last Updated**: 2026-04-27
**Created**: 2026-04-27

## Quality Mandate Ã¢â‚¬â€ MANDATORY

| Attribute | Requirement |
|-----------|-------------|
| Correctness | ALL additions accurate; no copy-paste errors |
| Completeness | NO phases/tasks/steps skipped; NO shortcuts |
| Thoroughness | Evidence-based validation at every step |
| Reliability | `lint-docs` and `lint-fitness` clean after every phase |
| Efficiency | Optimized for maintainability; NOT implementation speed |
| Accuracy | Root cause addressed; not just symptoms |
| NO Time Pressure | NEVER rush; NEVER skip validation; NEVER defer quality checks |
| NO Premature Completion | Objective evidence required before marking complete |

**ALL issues are blockers.** Fix immediately. NEVER defer.

---

## Task Status Legend Ã¢â‚¬â€ MANDATORY

| Symbol | Meaning | When to Use |
|--------|---------|-------------|
| Ã¢ÂÅ’ | Not started | Task not yet begun |
| Ã°Å¸â€â€ž | In progress | Currently being worked on |
| Ã¢Å“â€¦ | Complete | Task finished with evidence |
| Ã¢ÂÂ³ | Blocked | Requires external dependency (MUST have resolution plan) |

---

## Phase 0: Pre-flight Build Health

**Phase Objective**: Verify clean baseline before any changes.

### Task 0.1: Build Health Pre-flight

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 0.3h
- **Dependencies**: None
- **Acceptance Criteria**:
  - [x] `go build ./...` exits 0
  - [x] `go build -tags e2e,integration ./...` exits 0
  - [x] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [x] `go run ./cmd/cicd-lint lint-docs` exits 0
  - [x] `go test ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/...` exits 0
  - [x] Output archived in `test-output/v18v19-phase0/`
- **Files**: None (verification only)

---

## Phase 1: ENG-HANDBOOK.md Documentation Propagation

**Phase Objective**: Propagate 38+ items from 4 source docs into ENG-HANDBOOK.md; fix
tls-structure.md; verify with lint-docs; delete suggestion docs.

### Task 1.1: ENG-HANDBOOK.md additions from target-structure.md

- **Status**: Ã¢Å“â€¦  **Estimated**: 2h  **Actual**: 2h  **Dependencies**: Task 0.1
- Items 1.1Ã¢â‚¬â€œ1.11: 11 catalog entries Ã¢â€ â€™ Ã‚Â§2.1, Ã‚Â§4.4, Ã‚Â§4.4.1, Ã‚Â§4.4.4, Ã‚Â§4.4.6, Ã‚Â§B.7, Ã‚Â§11.1.4, Ã‚Â§12.2.1
- **Acceptance**:
  - [x] All 11 items added to ENG-HANDBOOK.md
  - [x] `go run ./cmd/cicd-lint lint-docs` exits 0
  - [x] Output in `test-output/v18v19-phase1/`

### Task 1.2: tls-structure.md fix + ENG-HANDBOOK.md additions

- **Status**: Ã¢Å“â€¦  **Estimated**: 1h  **Actual**: 1h  **Dependencies**: Task 1.1
- Admin CA Bundle fix + Items 1.12Ã¢â‚¬â€œ1.16: 6 items Ã¢â€ â€™ Ã‚Â§6.5, Ã‚Â§6.11
- **Acceptance**:
  - [x] Admin CA Bundle section fixed in `tls-structure.md` (Ã‚Â§6.5 updated in ENG-HANDBOOK.md)
  - [x] All 5 ENG-HANDBOOK.md items added
  - [x] `go run ./cmd/cicd-lint lint-docs` exits 0

### Task 1.3: ENG-HANDBOOK.md additions from deployment-templates.md

- **Status**: Ã¢Å“â€¦  **Estimated**: 3h  **Actual**: 3h  **Dependencies**: Task 1.2
- Items 1.17Ã¢â‚¬â€œ1.27: 11 items Ã¢â€ â€™ Ã‚Â§6.11.4, Ã‚Â§12.2.1, Ã‚Â§12.3.1, Ã‚Â§12.3.3, Ã‚Â§12.3.5, Ã‚Â§13.2, Ã‚Â§13.6
- **Acceptance**:
  - [x] All 11 items added to ENG-HANDBOOK.md
  - [x] `go run ./cmd/cicd-lint lint-docs` exits 0

### Task 1.4: ENG-HANDBOOK.md additions from claude-structure.md

- **Status**: Ã¢Å“â€¦  **Estimated**: 2h  **Actual**: 2h  **Dependencies**: Task 1.3
- Items 1.28Ã¢â‚¬â€œ1.38: 11 items Ã¢â€ â€™ Ã‚Â§2.1.1, Ã‚Â§2.1.5, Ã‚Â§14.11
- **Acceptance**:
  - [x] All 11 items added to ENG-HANDBOOK.md (Ã‚Â§14.11.1-Ã‚Â§14.11.7 + Ã‚Â§B.7 15-action table)
  - [x] `go run ./cmd/cicd-lint lint-docs` exits 0

### Task 1.5: lint-docs full verification + delete suggestion docs

- **Status**: Ã¢Å“â€¦  **Estimated**: 0.5h  **Actual**: 0.2h  **Dependencies**: Task 1.4
- Run `go run ./cmd/cicd-lint lint-docs`; fix all violations; delete 4 suggestion docs
- **Files** (to DELETE): Suggestion docs do NOT exist in repo Ã¢â‚¬â€ already satisfied
- **Acceptance**:
  - [x] lint-docs exits 0 Ã¢â‚¬â€ evidence in `test-output/v18v19-phase1/lint-docs-output.txt`
  - [x] All 4 suggestion docs confirmed absent (never existed)
  - [x] Output in `test-output/v18v19-phase1/`

---

## Phase 2: Prescriptive MANIFEST.yaml + Linter Extension

**Phase Objective**: Expand MANIFEST.yaml to be fully prescriptive; extend apps_ps_id_template linter.

### Task 2.1: Update MANIFEST.yaml

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: 0.5h
- **Dependencies**: Task 5.3
- **Files**: `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml`
- **Acceptance Criteria**:
  - [x] `required_server_dirs` field added (apis, config, model, repository + knownExclusions)
  - [x] `required_server_config_files` field added
  - [x] `required_server_repository_files` field added
  - [x] `required_server_repository_dirs` field added
  - [x] `required_e2e_files` field added (with `__SERVICE__` substitution)
  - [x] YAML parses without error

### Task 2.2: Implement checkServerDirs

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 1h
- **Dependencies**: Task 2.1
- **Files**: `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template.go`
- **Acceptance Criteria**:
  - [x] Function verifies `server/{dir}` for each RequiredServerDirs entry
  - [x] Respects `knownExclusions` per dir
  - [x] Unit test cases added in `apps_ps_id_template_test.go`

### Task 2.3: Implement checkServerConfigFiles + checkServerRepositoryFiles

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 1h
- **Dependencies**: Task 2.2
- **Files**: `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template.go`
- **Acceptance Criteria**:
  - [x] `checkServerConfigFiles` verifies `server/config/{file}`
  - [x] `checkServerRepositoryFiles` verifies `server/repository/{file}`
  - [x] `checkServerRepositoryDirs` verifies `server/repository/{dir}`
  - [x] Unit test cases for each function

### Task 2.4: Implement checkE2EFiles

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 0.5h
- **Dependencies**: Task 2.3
- **Files**: `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template.go`
- **Acceptance Criteria**:
  - [x] `checkE2EFiles` verifies `e2e/{file}` with `__SERVICE__` Ã¢â€ â€™ actual service name substitution
  - [x] Unit test cases added

### Task 2.5: Coverage + lint-fitness Validation

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 0.25h
- **Dependencies**: Task 2.4
- **Acceptance Criteria**:
  - [x] `go test ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/...` exits 0
  - [x] Coverage 100% for apps_ps_id_template package (Ã¢â€°Â¥98% target exceeded)
  - [x] `go run ./cmd/cicd-lint lint-fitness` exits 0 with initial knownExclusions in place
  - [x] Output archived in `test-output/v18v19-phase2/`

---

## Phase 3: Identity Services Server Code Migration

**Phase Objective**: Move domain code from identity service PS-ID roots Ã¢â€ â€™ server/.

### Task 3.1: identity-authz Inventory Ã¢â‚¬â€ Files at Root

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: Ã¢â‚¬â€
- **Actual**: 0.5h
- **Dependencies**: Task 2.5
- **Acceptance Criteria**:
  - [x] Complete inventory of files at identity-authz root (excluding CLI files)
  - [x] Package declarations noted for all files to move
  - [x] Import cycle risk assessed

### Task 3.2: identity-authz swagger.go + service.go Migration

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 1h
- **Dependencies**: Task 3.1
- **Files**: `internal/apps/identity-authz/server/` (swagger.go, service.go)
- **Acceptance Criteria**:
  - [x] `swagger.go` and `service.go` migration completed using server/apis pattern without import cycles
  - [x] Package declarations updated to `package apis` for migrated domain files and `package server` for server-level wrappers
  - [x] `go build ./internal/apps/identity-authz/...` exits 0

### Task 3.3: identity-authz handlers_*.go Migration Ã¢â€ â€™ server/apis/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: 2h
- **Dependencies**: Task 3.2
- **Files**: `internal/apps/identity-authz/server/apis/` (new dir + moved files)
- **Acceptance Criteria**:
  - [x] All `handlers_*.go` moved to `server/apis/` as `package apis`
  - [x] `authz_lifecycle_test.go`, `authz_port_conflict_test.go` created in server/
  - [x] `go test ./internal/apps/identity-authz/...` exits 0

### Task 3.4: identity-idp handlers + service Migration Ã¢â€ â€™ server/apis/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: 2h
- **Dependencies**: Task 3.3
- **Files**: `internal/apps/identity-idp/server/` (multiple files)
- **Acceptance Criteria**:
  - [x] `swagger.go`, `service.go`, all `handlers_*.go` moved
  - [x] `idp_lifecycle_test.go`, `idp_port_conflict_test.go` created in server/
  - [x] `go test ./internal/apps/identity-idp/...` exits 0

### Task 3.5: identity-rs service.go + validator.go Migration Ã¢â€ â€™ server/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: 1h
- **Dependencies**: Task 3.4
- **Files**: `internal/apps/identity-rs/server/`
- **Acceptance Criteria**:
  - [x] `swagger.go`, `service.go`, `validator.go` moved to server/
  - [x] `rs_lifecycle_test.go`, `rs_port_conflict_test.go` created in server/
  - [x] `go test ./internal/apps/identity-rs/...` exits 0

### Task 3.6: identity-rp rp_test.go Migration + Lifecycle Tests

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 0.75h
- **Dependencies**: Task 3.5
- **Files**: `internal/apps/identity-rp/server/`
- **Acceptance Criteria**:
  - [x] `rp_test.go` moved from root to server/ (package updated to `package server_test`)
  - [x] `rp_lifecycle_test.go`, `rp_port_conflict_test.go` created in server/
  - [x] `go test ./internal/apps/identity-rp/...` exits 0

### Task 3.7: identity-spa spa_test.go Migration + Lifecycle Tests

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 0.75h
- **Dependencies**: Task 3.6
- **Files**: `internal/apps/identity-spa/server/`
- **Acceptance Criteria**:
  - [x] `spa_test.go` moved from root to server/
  - [x] `spa_lifecycle_test.go`, `spa_port_conflict_test.go` created in server/
  - [x] `go test ./internal/apps/identity-spa/...` exits 0

### Task 3.8: Full Identity Suite Build + Test

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: 0.5h
- **Dependencies**: Tasks 7.2Ã¢â‚¬â€œ7.7
- **Acceptance Criteria**:
  - [x] `go build ./internal/apps/identity-.../...` exits 0
  - [x] `go test ./internal/apps/identity-.../...` exits 0
  - [x] `golangci-lint run ./internal/apps/identity-.../...` exits 0

### Task 3.9: lint-fitness Post-Migration Check

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: 0.25h
- **Dependencies**: Task 3.8
- **Acceptance Criteria**:
  - [x] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [x] Output archived in `test-output/v18v19-phase3/`

---

## Phase 4: sm-im Root Cleanup

**Phase Objective**: Move non-CLI test files from sm-im root Ã¢â€ â€™ server/.

### Task 4.1: Move sm-im Server Test Files from Root Ã¢â€ â€™ server/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: 1h
- **Dependencies**: Task 3.9
- **Files to MOVE** from `internal/apps/sm-im/` root to `internal/apps/sm-im/server/`:
  - `http_test.go`
  - `http_errors_test.go`
  - `response_body_test.go`
  - `im_database_test.go`
  - `im_server_lifecycle_test.go`
  - `im_lifecycle_test.go`
  - `im_port_conflict_test.go`
- **Acceptance Criteria**:
  - [x] All 7 files moved; package declarations updated if needed
  - [x] `go test ./internal/apps/sm-im/...` exits 0

### Task 4.2: Delete testmain_test.go from sm-im Root

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: 0.5h
- **Dependencies**: Task 4.1
- **Files to DELETE**: `internal/apps/sm-im/testmain_test.go`
- **Acceptance Criteria**:
  - [x] Root `testmain_test.go` deleted (server/ copy retained)
  - [x] `go test ./internal/apps/sm-im/...` exits 0
  - [x] sm-im root has ONLY `im.go`, `im_usage.go`, `im_cli_commands_test.go`, `im_cli_url_test.go`
  - [x] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [x] Output archived in `test-output/v18v19-phase4/`

---

## Phase 5: Create Missing server/ Subdirectory Packages

**Phase Objective**: Create server/apis/, server/model/, server/repository/ for 5 identity services.

### Task 5.1: identity-authz server/model/ + server/repository/migrations/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 4.2
- **Files** (NEW):
  - `internal/apps/identity-authz/server/model/model.go`
  - `internal/apps/identity-authz/server/repository/migrations/` (dir)
  - `internal/apps/identity-authz/server/repository/migrations.go`
- **Acceptance Criteria**:
  - [x] Migration SQL uses range from registry.yaml for identity-authz
  - [x] `go build ./internal/apps/identity-authz/...` exits 0

### Task 5.2: identity-idp server/model/ + server/repository/migrations/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.1
- **Files** (NEW): same pattern as Task 5.1 for identity-idp
- **Acceptance Criteria**:
  - [x] Migration SQL uses range from registry.yaml for identity-idp
  - [x] `go build ./internal/apps/identity-idp/...` exits 0

### Task 5.3: identity-rs server/model/ + server/repository/migrations/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.2
- **Files** (NEW): same pattern for identity-rs
- **Acceptance Criteria**:
  - [x] Migration SQL uses range from registry.yaml for identity-rs
  - [x] `go build ./internal/apps/identity-rs/...` exits 0

### Task 5.4: identity-rp server/apis/ + server/model/ + server/repository/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.3
- **Files** (NEW): server/apis/handler.go (minimal), model/, repository/migrations/
- **Acceptance Criteria**:
  - [x] Minimal handler in server/apis/
  - [x] Migration SQL uses range from registry.yaml for identity-rp
  - [x] `go build ./internal/apps/identity-rp/...` exits 0

### Task 5.5: identity-spa server/apis/ + server/model/ + server/repository/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.4
- **Files** (NEW): same pattern for identity-spa
- **Acceptance Criteria**:
  - [x] Minimal handler in server/apis/
  - [x] Migration SQL uses range from registry.yaml for identity-spa
  - [x] `go build ./internal/apps/identity-spa/...` exits 0

### Task 5.6: Phase 9 Build + lint-fitness Validation

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.5
- **Acceptance Criteria**:
  - [x] `go build ./...` exits 0
  - [x] `go test ./internal/apps/identity-.../...` exits 0
  - [x] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [x] Output archived in `test-output/v18v19-phase5/`

---

## Phase 6: Create Missing client/ Packages

**Phase Objective**: Create typed HTTP client packages for 8 PS-IDs that currently lack them.

### Task 6.1: jose-ja client/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.6
- **Files** (NEW): `internal/apps/jose-ja/client/client.go`
- **Acceptance Criteria**:
  - [x] GetJWKS, CreateJWK, RotateJWK methods implemented
  - [x] `go build ./internal/apps/jose-ja/...` exits 0

### Task 6.2: pki-ca client/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.6
- **Files** (NEW): `internal/apps/pki-ca/client/client.go`
- **Acceptance Criteria**:
  - [x] IssueCert, RevokeCert, GetCRL methods implemented
  - [x] `go build ./internal/apps/pki-ca/...` exits 0

### Task 6.3: identity-authz client/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.6
- **Files** (NEW): `internal/apps/identity-authz/client/client.go`
- **Acceptance Criteria**:
  - [x] Authorize, Introspect, Token methods implemented
  - [x] `go build ./internal/apps/identity-authz/...` exits 0

### Task 6.4: identity-idp client/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.6
- **Files** (NEW): `internal/apps/identity-idp/client/client.go`
- **Acceptance Criteria**:
  - [x] Login, Logout, JWKS methods implemented
  - [x] `go build ./internal/apps/identity-idp/...` exits 0

### Task 6.5: identity-rs client/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.6
- **Files** (NEW): `internal/apps/identity-rs/client/client.go`
- **Acceptance Criteria**:
  - [x] ValidateToken, GetResources methods implemented
  - [x] `go build ./internal/apps/identity-rs/...` exits 0

### Task 6.6: identity-rp client/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.6
- **Files** (NEW): `internal/apps/identity-rp/client/client.go`
- **Acceptance Criteria**:
  - [x] Callback, Logout methods implemented
  - [x] `go build ./internal/apps/identity-rp/...` exits 0

### Task 6.7: identity-spa client/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.6
- **Files** (NEW): `internal/apps/identity-spa/client/client.go`
- **Acceptance Criteria**:
  - [x] Minimal API surface implemented
  - [x] `go build ./internal/apps/identity-spa/...` exits 0

### Task 6.8: skeleton-template client/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 5.6
- **Files** (NEW): `internal/apps/skeleton-template/client/client.go`
- **Acceptance Criteria**:
  - [x] Placeholder client implemented
  - [x] `go build ./internal/apps/skeleton-template/...` exits 0

### Task 6.9: Phase 6 Build + lint-fitness Validation

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Tasks 6.1Ã¢â‚¬â€œ6.8
- **Acceptance Criteria**:
  - [x] `go build ./...` exits 0
  - [x] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [x] `required_dirs: client` knownExclusions emptied for migrated services
  - [x] Output archived in `test-output/v18v19-phase6/`

---

## Phase 7: Create Missing e2e/ Packages

**Phase Objective**: Create E2E test packages for 5 PS-IDs that currently lack them.

### Task 7.1: pki-ca e2e/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 6.9
- **Files** (NEW):
  - `internal/apps/pki-ca/e2e/testmain_e2e_test.go`
  - `internal/apps/pki-ca/e2e/ca_e2e_test.go`
- **Acceptance Criteria**:
  - [x] Both files have `//go:build e2e` as first line
  - [x] `testmain_e2e_test.go` has TestMain that starts Docker Compose
  - [x] `go build -tags e2e ./internal/apps/pki-ca/...` exits 0

### Task 7.2: identity-idp e2e/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 6.9
- **Files** (NEW):
  - `internal/apps/identity-idp/e2e/testmain_e2e_test.go`
  - `internal/apps/identity-idp/e2e/idp_e2e_test.go`
- **Acceptance Criteria**:
  - [x] Both files have `//go:build e2e` as first line
  - [x] `go build -tags e2e ./internal/apps/identity-idp/...` exits 0

### Task 7.3: identity-rs e2e/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 6.9
- **Files** (NEW):
  - `internal/apps/identity-rs/e2e/testmain_e2e_test.go`
  - `internal/apps/identity-rs/e2e/rs_e2e_test.go`
- **Acceptance Criteria**:
  - [x] Both files have `//go:build e2e` as first line
  - [x] `go build -tags e2e ./internal/apps/identity-rs/...` exits 0

### Task 7.4: identity-rp e2e/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 6.9
- **Files** (NEW):
  - `internal/apps/identity-rp/e2e/testmain_e2e_test.go`
  - `internal/apps/identity-rp/e2e/rp_e2e_test.go`
- **Acceptance Criteria**:
  - [x] Both files have `//go:build e2e` as first line
  - [x] `go build -tags e2e ./internal/apps/identity-rp/...` exits 0

### Task 7.5: identity-spa e2e/

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 6.9
- **Files** (NEW):
  - `internal/apps/identity-spa/e2e/testmain_e2e_test.go`
  - `internal/apps/identity-spa/e2e/spa_e2e_test.go`
- **Acceptance Criteria**:
  - [x] Both files have `//go:build e2e` as first line
  - [x] `go build -tags e2e ./internal/apps/identity-spa/...` exits 0

### Task 7.6: Phase 7 Build + lint-fitness Validation

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Tasks 7.1Ã¢â‚¬â€œ7.5
- **Acceptance Criteria**:
  - [x] `go build -tags e2e ./...` exits 0
  - [x] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [x] `required_dirs: e2e` knownExclusions emptied for migrated services
  - [x] Output archived in `test-output/v18v19-phase7/`

---

## Phase 8: Remove knownExclusions + Final Validation

**Phase Objective**: Remove temporary knownExclusions from MANIFEST/linter after all migration complete.

### Task 8.1: Remove Temporary knownExclusions

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 3.6
- **Files**: `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml` (or linter Go source)
- **Acceptance Criteria**:
  - [x] All identity service exclusions for `required_server_dirs` removed
  - [x] All identity service exclusions for `required_server_config_files` removed
  - [x] All identity service exclusions for `required_server_repository_files` removed
  - [x] All identity service exclusions for `required_dirs: client` removed
  - [x] All identity service exclusions for `required_dirs: e2e` removed
  - [x] All identity service exclusions for `required_e2e_files` removed
  - [x] Only 3 permanent exceptions remain (sm-kms public_server.go, sm-im CLI test, sm-kms/pki-ca server/ subdirs)

### Task 8.2: Final lint-fitness + Full Build Validation

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 8.1
- **Acceptance Criteria**:
  - [x] `go build ./...` exits 0
  - [x] `go build -tags e2e,integration ./...` exits 0
  - [x] `go test ./...` exits 0
  - [x] `golangci-lint run ./...` exits 0
  - [x] `golangci-lint run --build-tags e2e,integration ./...` exits 0
  - [x] `go run ./cmd/cicd-lint lint-fitness` exits 0 (only 3 permanent exceptions)
  - [x] Race detector attempted: `go test -race -count=2 ./...` (blocked by missing C compiler `gcc` in current Windows environment)
  - [x] Output archived in `test-output/v18v19-phase8/`

---

## Phase 9: Knowledge Propagation

**Phase Objective**: Apply lessons from all phases to permanent artifacts.

### Task 9.1: Review lessons.md + Update ENG-HANDBOOK.md

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.75h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 8.2
- **Files**: `docs/ENG-HANDBOOK.md`
- **Acceptance Criteria**:
  - [x] Canonical PS-ID structure spec updated to reflect final state
  - [x] MANIFEST field catalog added or updated
  - [x] Migration range patterns documented

### Task 9.2: Update target-structure.md

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 9.1
- **Files**: `docs/target-structure.md`
- **Acceptance Criteria**:
  - [x] Canonical PS-ID layout updated to reflect plan outcomes
  - [x] Server/ subdirectory state table updated (all 10 PS-IDs)

### Task 9.3: Update Instruction Files + Skills

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 9.1
- **Acceptance Criteria**:
  - [x] `.github/skills/fitness-function-gen/SKILL.md` updated with recursive MANIFEST pattern
  - [x] Instruction files updated where code migration work surfaces new patterns
  - [x] `.claude/skills/` counterparts synced (lint-agent-drift must pass)

### Task 9.4: Propagation Verification + Final Commit

- **Status**: Ã¢Å“â€¦
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: Ã¢â‚¬â€
- **Dependencies**: Task 9.3
- **Acceptance Criteria**:
  - [x] `go run ./cmd/cicd-lint lint-docs` exits 0
  - [x] `git status --porcelain` returns empty
  - [x] Output archived in `test-output/v18v19-phase9/`

---

## Cross-Cutting Quality Gates

- [x] `go build ./...` exits 0 (maintained after every task)
- [x] `go test ./...` exits 0
- [x] `golangci-lint run ./...` exits 0
- [x] `golangci-lint run --build-tags e2e,integration ./...` exits 0
- [x] `go run ./cmd/cicd-lint lint-fitness` exits 0 (maintained after each code phase)
- [x] `go run ./cmd/cicd-lint lint-docs` exits 0 (maintained after Phase 1 and Phase 9)
- [x] Coverage Ã¢â€°Â¥98% for apps_ps_id_template; Ã¢â€°Â¥95% for identity service packages
- [x] Race detector attempted: `go test -race -count=2 ./...` (blocked by missing C compiler `gcc` in current Windows environment)

---

## Evidence Archive

- `test-output/v18v19-phase0/` Ã¢â‚¬â€ Pre-flight build health
- `test-output/v18v19-phase1/` Ã¢â‚¬â€ Documentation propagation lint-docs output
- `test-output/v18v19-phase2/` Ã¢â‚¬â€ apps_ps_id_template coverage + lint-fitness
- `test-output/v18v19-phase3/` Ã¢â‚¬â€ Identity migration test results
- `test-output/v18v19-phase4/` Ã¢â‚¬â€ sm-im cleanup lint-fitness
- `test-output/v18v19-phase5/` Ã¢â‚¬â€ Identity server/ subdir build verification
- `test-output/v18v19-phase6/` Ã¢â‚¬â€ Client/ creation build verification
- `test-output/v18v19-phase7/` Ã¢â‚¬â€ e2e/ creation build verification
- `test-output/v18v19-phase8/` Ã¢â‚¬â€ Final full validation
- `test-output/v18v19-phase9/` Ã¢â‚¬â€ Knowledge propagation lint-docs
