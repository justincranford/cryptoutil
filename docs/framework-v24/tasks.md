 Tasks - Framework V24: PS-ID Convergence â€” Canonical Server Structure and 10-PS-ID Alignment

**Status**: 0 of 34 tasks complete (0%)
**Created**: 2026-05-17
**Last Updated**: 2026-05-17

## Task Status Legend

| Symbol | Meaning |
|--------|---------|
| âŒ | Not started |
| ðŸ”„ | In progress |
| âœ… | Complete |
| â³ | Blocked |

## Decision Constraints

1. `docs/ENG-HANDBOOK.md` is the source of truth for all architectural patterns.
2. `docs/framework-v24/quizme-v2.md` is the canonical structure authority for all PS-ID
   directory decisions. Decisions recorded there are binding.
3. sm-kms migration must preserve all existing tests and business logic semantics.
4. pki-ca migration must not touch the CA domain layers (bootstrap, compliance, issuer, etc.).
5. Every exclusion removal must be validated by `lint-fitness` before marking complete.
6. Mojibake sub-linter must not produce false positives on legitimate UTF-8 content.
7. All phases complete only when `go test ./...` and `golangci-lint run` pass.
8. Docker Compose verification MUST be run within the same phase as any server-layer change.

## Phase 1: sm-kms Server Subdirectory Migration

### Task 1.1: Audit sm-kms server/ layout and create migration plan

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] All files in `server/businesslogic/` and `server/handler/` categorized: OAS-handler, ORM-entity, pure-logic.
  - [ ] Target directories for each file identified: `server/apis/`, `server/model/`, or stays in `server/businesslogic/`.
  - [ ] Import graph checked for cycles that would arise from the split.
  - [ ] Evidence archived in `test-output/v24-phase1/`.

### Task 1.2: Create server/model/ with GORM entity structs

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `server/model/model.go` (and any additional model files) created from the ORM-entity files in `businesslogic/`.
  - [ ] All moved types compile with updated package name `model`.
  - [ ] `go build ./internal/apps/sm-kms/server/model/...` exits 0.

### Task 1.3: Create server/apis/ with OAS handler files

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] All OAS handler files from `server/handler/` moved or copied into `server/apis/`.
  - [ ] Package name updated to `apis`.
  - [ ] `go build ./internal/apps/sm-kms/server/apis/...` exits 0.

### Task 1.4: Create server/config/ with config wiring

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `server/config/config.go`, `config_test.go`, `config_test_helper.go` created.
  - [ ] Config struct matches the canonical pattern used by other PS-IDs (identity-authz is the reference).
  - [ ] `go build ./internal/apps/sm-kms/server/config/...` exits 0.

### Task 1.5: Update import paths and verify full sm-kms build

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] All files importing the old `businesslogic` and `handler` packages updated to new paths.
  - [ ] `go build ./internal/apps/sm-kms/...` exits 0.
  - [ ] `go test ./internal/apps/sm-kms/...` passes (zero failures, zero skips).
  - [ ] `golangci-lint run ./internal/apps/sm-kms/...` exits 0.

### Task 1.6: Phase 1 verification

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0.
  - [ ] `go test ./internal/apps/sm-kms/...` passes.
  - [ ] Evidence archived in `test-output/v24-phase1/`.

## Phase 2: pki-ca Server Subdirectory Migration

### Task 2.1: Audit pki-ca server/ layout and identify gaps

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] Existing pki-ca server/ layout documented: current cmd/, config/, middleware/ are non-canonical but accepted.
  - [ ] Three missing canonical dirs identified: apis/, model/, repository/.
  - [ ] Existing pki-ca SQL migration source paths identified (repository-v2/migrations).
  - [ ] Plan for thin handler wrappers in server/apis/ that delegate to existing domain layers documented.
  - [ ] Evidence archived in `test-output/v24-phase2/`.

### Task 2.2: Create server/apis/ with thin handler wrappers

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `server/apis/` directory created with at least a skeleton handler file.
  - [ ] Handlers delegate to existing pki-ca domain layers (not duplicating logic).
  - [ ] `go build ./internal/apps/pki-ca/server/apis/...` exits 0.

### Task 2.3: Create server/model/ with GORM entity structs

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `server/model/model.go` created with pki-ca GORM entity types.
  - [ ] Package name is `model`.
  - [ ] `go build ./internal/apps/pki-ca/server/model/...` exits 0.

### Task 2.4: Create server/repository/ with migrations.go and migrations/

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `server/repository/` directory created with `migrations.go` (//go:embed for migrations FS).
  - [ ] `server/repository/migrations/` directory created with existing SQL consolidated from `repository-v2/migrations/`.
  - [ ] `go build ./internal/apps/pki-ca/server/repository/...` exits 0.

### Task 2.5: Create server/config/config_test_helper.go

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `server/config/config_test_helper.go` created with canonical test helper pattern.
  - [ ] Compiles clean under `go build` and `go test`.

### Task 2.6: Phase 2 verification

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `go build ./internal/apps/pki-ca/...` exits 0.
  - [ ] `go test ./internal/apps/pki-ca/...` passes (zero failures, zero skips).
  - [ ] `golangci-lint run ./internal/apps/pki-ca/...` exits 0.
  - [ ] Evidence archived in `test-output/v24-phase2/`.

## Phase 3: Linter Exclusion Cleanup

### Task 3.1: Remove sm-kms exclusions from apps_ps_id_template.go

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `knownServerFileExclusions["testmain_test.go"]` sm-kms entry removed.
  - [ ] `knownServerDirExclusions["apis"]` sm-kms entry removed.
  - [ ] `knownServerDirExclusions["model"]` sm-kms entry removed.
  - [ ] `knownServerConfigFileExclusions["config.go"]` sm-kms entry removed.
  - [ ] `knownServerConfigFileExclusions["config_test.go"]` sm-kms entry removed.
  - [ ] `knownServerConfigFileExclusions["config_test_helper.go"]` sm-kms entry removed.
  - [ ] `lint-fitness` passes after removals.

### Task 3.2: Remove pki-ca exclusions from apps_ps_id_template.go

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `knownServerDirExclusions["apis"]` pki-ca entry removed.
  - [ ] `knownServerDirExclusions["model"]` pki-ca entry removed.
  - [ ] `knownServerDirExclusions["repository"]` pki-ca entry removed.
  - [ ] `knownServerConfigFileExclusions["config_test_helper.go"]` pki-ca entry removed.
  - [ ] `knownServerRepositoryFileExclusions["migrations.go"]` pki-ca entry removed.
  - [ ] `knownServerRepositoryDirExclusions["migrations"]` pki-ca entry removed.
  - [ ] `lint-fitness` passes after removals.

### Task 3.3: Remove stale swagger.go exclusion

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `knownServerFileExclusions["swagger.go"]` entry removed (all 10 PS-IDs already have swagger.go in server/).
  - [ ] `knownServerFileExclusions["swagger_test.go"]` entry removed.
  - [ ] Verify `checkServerFiles` validates presence of swagger.go in server/ correctly after removal.
  - [ ] `lint-fitness` passes after removals.

### Task 3.4: Migrate remaining root lifecycle and port-conflict tests into server/

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] Move `*_lifecycle_test.go` and `*_port_conflict_test.go` into `server/` for sm-kms, jose-ja, pki-ca, skeleton-template.
  - [ ] Remove now-stale exclusions from `knownServerFileExclusions` for those files.
  - [ ] `go test` and `lint-fitness` both pass after moves.

### Task 3.5: Update MANIFEST.yaml stale comments

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `required_server_dirs` comments in MANIFEST.yaml no longer contain stale migration-pending notes.
  - [ ] `optional_server_dirs` updated to reflect current reality.
  - [ ] `lint-fitness` passes after update.

### Task 3.6: Phase 3 verification

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `lint-fitness` passes with smaller/cleaner exclusion maps.
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_fitness/...` passes.
  - [ ] Evidence archived in `test-output/v24-phase3/`.

## Phase 4: Tooling Quality Improvements

### Task 4.1: Implement mojibake sub-linter

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `internal/apps-tools/cicd_lint/lint_text/mojibake/mojibake.go` created.
  - [ ] Linter scans `.md` files under `docs/` for windows-1252 mojibake byte sequences.
  - [ ] Returns an error (not just warning) when mojibake markers are found.
  - [ ] Does NOT false-positive on legitimate emoji or non-ASCII language characters.
  - [ ] `mojibake_test.go` covers: clean file passes, mojibake file fails, empty file passes, non-md file skipped.

### Task 4.2: Register mojibake sub-linter in lint_text.go

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `mojibake.Check` added to the sub-linter table in `lint_text.go`.
  - [ ] `go run ./cmd/cicd-lint lint-text` runs mojibake check and passes on current repo.
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_text/...` passes.

### Task 4.3: Fix createFullPSIDRoot to parse required_server_dirs from MANIFEST

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `createFullPSIDRoot` in `apps_ps_id_template_test.go` reads `required_server_dirs` from the parsed MANIFEST.yaml struct.
  - [ ] The hardcoded `[]string{"apis", "model", "repository"}` slice is removed.
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/...` passes.

### Task 4.4: Phase 4 verification

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-text` passes (mojibake sub-linter included and passes).
  - [ ] `go test ./internal/apps-tools/cicd_lint/...` passes.
  - [ ] Evidence archived in `test-output/v24-phase4/`.

## Phase 5: All-10 PS-ID Canonical Structure Audit

### Task 5.1: Baseline all-10 PS-ID recursive directory inventory

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] Complete recursive directory listing for all 10 PS-IDs archived in `test-output/v24-phase5/`.
  - [ ] Each directory compared against canonical allowed policy in `quizme-v2.md`.
  - [ ] Per-PS-ID gap table produced: missing required dirs, present transitional dirs.

### Task 5.2: Produce per-PS-ID migration plans

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] For each PS-ID with gaps: a task list documenting required changes.
  - [ ] For each transitional dir: a retirement plan (migrate to canonical location + removal date).
  - [ ] Migration order documented to avoid import-cycle regressions.

### Task 5.3: Execute per-PS-ID migrations

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] All identified gaps resolved for all 10 PS-IDs.
  - [ ] All transitional directories retired or have explicit retirement dates committed.
  - [ ] `go build ./internal/apps/...` exits 0 after all migrations.
  - [ ] `go test ./internal/apps/...` passes after all migrations.

### Task 5.4: Update MANIFEST.yaml and exclusion maps

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `required_server_dirs` in MANIFEST.yaml reflects final canonical required set.
  - [ ] Transitional allowlist entries removed from exclusion maps for retired directories.
  - [ ] `lint-fitness apps-ps-id-template` passes with no false exclusions remaining.

### Task 5.5: Phase 5 verification

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] All 10 PS-IDs have `server/apis/`, `server/model/`, `server/config/`, `server/repository/`.
  - [ ] `lint-fitness` passes with no exclusions for the required directories.
  - [ ] Evidence archived in `test-output/v24-phase5/`.

## Phase 6: Framework Duplication Audit

### Task 6.1: Audit PS-IDs for duplicated patterns

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] Each PS-ID's `server/apis/`, `server/config/`, `server/businesslogic/` audited for patterns duplicated across 3+ services.
  - [ ] Duplicated patterns catalogued with per-service file paths.
  - [ ] Priority order for extraction established (highest duplication first).

### Task 6.2: Extract top-N duplicated patterns to framework

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] Top-N duplicated patterns (N defined during Task 6.1 audit) extracted to `internal/apps-framework/service/`.
  - [ ] All PS-IDs migrated to call the shared implementation.
  - [ ] `go test ./...` passes after extraction (no behavioral regression).

### Task 6.3: Phase 6 verification

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] Zero code patterns duplicated across 3+ PS-IDs where a framework equivalent now exists.
  - [ ] Evidence archived in `test-output/v24-phase6/`.

## Phase 7: Verification and Closure

### Task 7.1: Full quality suite

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0.
  - [ ] `go build -tags e2e,integration ./...` exits 0.
  - [ ] `go test ./...` exits 0.
  - [ ] `golangci-lint run ./...` exits 0.
  - [ ] `go run ./cmd/cicd-lint lint-go` exits 0.
  - [ ] `go run ./cmd/cicd-lint lint-text` exits 0 (includes mojibake).
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0.
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0.

### Task 7.2: Docker Compose verification for changed PS-IDs

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `docker compose up --wait` succeeds for each PS-ID that had server-layer changes.
  - [ ] Health endpoint check passes for each.
  - [ ] Evidence archived in `test-output/v24-phase7/`.

## Phase 8: Knowledge Propagation

### Task 8.1: Review execution lessons for durable guidance updates

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] `docs/framework-v24/lessons.md` reviewed after all execution phases complete.
  - [ ] Any new durable guidance for handbook or instructions explicitly identified.

### Task 8.2: Apply durable documentation/process updates

- **Status**: âŒ
- **Acceptance Criteria**:
  - [ ] Durable doc/process updates from V24 applied to handbook or instructions.
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes after any updates.
  - [ ] Evidence archived in `test-output/v24-phase8/`.
