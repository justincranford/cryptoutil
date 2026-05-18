# Tasks - Framework V22: Server Subdirectory Migration and Tooling Debt Closure

**Status**: 0 of 28 tasks complete (0%)
**Created**: 2026-04-30
**Last Updated**: 2026-04-30

## Task Status Legend

| Symbol | Meaning |
|--------|---------|
| ❌ | Not started |
| 🔄 | In progress |
| ✅ | Complete |
| ⏳ | Blocked |

## Decision Constraints

1. `docs/ENG-HANDBOOK.md` is the source of truth for all architectural patterns.
2. sm-kms migration must preserve all existing tests and business logic semantics.
3. pki-ca migration must not touch the CA domain layers (bootstrap, compliance, issuer, etc.).
4. Every exclusion removal must be validated by `lint-fitness` before marking complete.
5. Mojibake sub-linter must not produce false positives on legitimate UTF-8 content.
6. All phases complete only when `go test ./...` and `golangci-lint run` pass.
7. Quizme Round 1 decisions are binding: `server/businesslogic/` is canonical for pure business
  logic, lifecycle/port-conflict tests must move into `server/`, and pki-ca migration planning
  must include repository-v2 SQL migration consolidation mapping.
8. Quizme Round 2 must finalize canonical recursive directory structure to be enforced for all
  10 PS-IDs before broad all-service refactor execution begins.

## Phase 1: sm-kms Server Subdirectory Migration

### Task 1.1: Audit sm-kms server/ layout and create migration plan

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] All files in `server/businesslogic/` and `server/handler/` categorized: OAS-handler, ORM-entity, pure-logic.
  - [x] Target directories for each file identified: `server/apis/`, `server/model/`, or stays in `server/businesslogic/`.
  - [x] Import graph checked for cycles that would arise from the split.
  - [x] Evidence archived in `test-output/v23-phase1/`.

### Task 1.2: Create server/model/ with GORM entity structs

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `server/model/model.go` (and any additional model files) created from the ORM-entity files in `businesslogic/`.
  - [x] All moved types compile with updated package name `model`.
  - [x] `go build ./internal/apps/sm-kms/server/model/...` exits 0.

### Task 1.3: Create server/apis/ with OAS handler files

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] All OAS handler files from `server/handler/` moved or copied into `server/apis/`.
  - [x] Package name updated to `apis`.
  - [x] `go build ./internal/apps/sm-kms/server/apis/...` exits 0.

### Task 1.4: Create server/config/ with config wiring

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `server/config/config.go`, `config_test.go`, `config_test_helper.go` created.
  - [x] Config struct matches the canonical pattern used by other PS-IDs (identity-authz is the reference).
  - [x] `go build ./internal/apps/sm-kms/server/config/...` exits 0.

### Task 1.5: Update import paths and verify full sm-kms build

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] All files importing the old `businesslogic` and `handler` packages updated to new paths.
  - [x] `go build ./internal/apps/sm-kms/...` exits 0.
  - [x] `go test ./internal/apps/sm-kms/...` passes (zero failures, zero skips).
  - [x] `golangci-lint run ./internal/apps/sm-kms/...` exits 0.

### Task 1.6: Phase 1 verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `go build ./...` exits 0.
  - [x] `go test ./internal/apps/sm-kms/...` passes.
  - [x] Evidence archived in `test-output/v23-phase1/`.

## Phase 2: pki-ca Server Subdirectory Migration

### Task 2.1: Audit pki-ca server/ layout and identify gaps

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] Existing pki-ca server/ layout documented: current cmd/, config/, middleware/ are non-canonical but accepted.
  - [x] Three missing canonical dirs identified: apis/, model/, repository/.
  - [x] Existing pki-ca SQL migration source paths identified (repository-v2/migrations).
  - [x] Plan for thin handler wrappers in server/apis/ that delegate to existing domain layers documented.
  - [x] Evidence archived in `test-output/v23-phase2/`.

### Task 2.7: Produce pki-ca consolidation map to canonical structure

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] Current pki-ca recursive subdirectory set inventoried with evidence.
  - [x] Each current pki-ca subdirectory is mapped to one of: keep, migrate, merge, or deprecate.
  - [x] Mapping includes target canonical location under the selected all-10 PS-ID structure.
  - [x] Migration order is documented to avoid import-cycle and runtime-risk regressions.

### Task 2.8: Define all-10 PS-ID canonical directory rollout plan

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] Recursive directory supersets computed for focus PS-IDs and all 10 PS-IDs.
  - [x] Canonical required/optional directory policy drafted for all 10 PS-IDs.
  - [x] Linter/template update tasks listed to enforce allowed-only subdirectories.
  - [x] Rollout sequencing documented with pki-ca as high-sprawl migration case.

### Task 2.2: Create server/apis/ with thin handler wrappers

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `server/apis/` directory created with at least a skeleton handler file.
  - [x] Handlers delegate to existing pki-ca domain layers (not duplicating logic).
  - [x] `go build ./internal/apps/pki-ca/server/apis/...` exits 0.

### Task 2.3: Create server/model/ with GORM entity structs

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `server/model/model.go` created with pki-ca GORM entity types.
  - [x] Package name is `model`.
  - [x] `go build ./internal/apps/pki-ca/server/model/...` exits 0.

### Task 2.4: Create server/repository/ with migrations.go and migrations/

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `server/repository/` directory created with `migrations.go` (//go:embed for migrations FS).
  - [x] `server/repository/migrations/` directory created (may be empty initially, or contain existing SQL files).
  - [x] `go build ./internal/apps/pki-ca/server/repository/...` exits 0.

### Task 2.5: Create server/config/config_test_helper.go

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `server/config/config_test_helper.go` created with canonical test helper pattern.
  - [x] Compiles clean under `go build` and `go test`.

### Task 2.6: Phase 2 verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `go build ./internal/apps/pki-ca/...` exits 0.
  - [x] `go test ./internal/apps/pki-ca/...` passes (zero failures, zero skips).
  - [x] `golangci-lint run ./internal/apps/pki-ca/...` exits 0.
  - [x] Evidence archived in `test-output/v23-phase2/`.

## Phase 3: Linter Exclusion Cleanup

### Task 3.1: Remove sm-kms exclusions from apps_ps_id_template.go

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `knownServerFileExclusions["testmain_test.go"]` sm-kms entry removed.
  - [x] `knownServerDirExclusions["apis"]` sm-kms entry removed.
  - [x] `knownServerDirExclusions["model"]` sm-kms entry removed.
  - [x] `knownServerConfigFileExclusions["config.go"]` sm-kms entry removed.
  - [x] `knownServerConfigFileExclusions["config_test.go"]` sm-kms entry removed.
  - [x] `knownServerConfigFileExclusions["config_test_helper.go"]` sm-kms entry removed.
  - [x] `lint-fitness` passes after removals.

### Task 3.2: Remove pki-ca exclusions from apps_ps_id_template.go

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `knownServerDirExclusions["apis"]` pki-ca entry removed.
  - [x] `knownServerDirExclusions["model"]` pki-ca entry removed.
  - [x] `knownServerDirExclusions["repository"]` pki-ca entry removed.
  - [x] `knownServerConfigFileExclusions["config_test_helper.go"]` pki-ca entry removed.
  - [x] `knownServerRepositoryFileExclusions["migrations.go"]` pki-ca entry removed.
  - [x] `knownServerRepositoryDirExclusions["migrations"]` pki-ca entry removed.
  - [x] `lint-fitness` passes after removals.

### Task 3.3: Remove stale swagger.go exclusion from apps_ps_id_template.go

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `knownServerFileExclusions["swagger.go"]` entry removed (all 10 PS-IDs already have swagger.go in server/).
  - [x] `knownServerFileExclusions["swagger_test.go"]` entry removed.
  - [x] Verify `checkServerFiles` validates presence of swagger.go in server/ correctly after removal.
  - [x] `lint-fitness` passes after removals.

### Task 3.3a: Migrate remaining root lifecycle and port-conflict tests into server/

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] Move `__SERVICE___lifecycle_test.go` and `__SERVICE___port_conflict_test.go` into `server/` for sm-kms, jose-ja, pki-ca, and skeleton-template.
  - [x] Remove now-stale exclusions from `knownServerFileExclusions` for those files.
  - [x] `go test` and `lint-fitness` both pass after moves.

### Task 3.4: Update MANIFEST.yaml to remove stale migration references

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `required_server_dirs` comments in MANIFEST.yaml no longer reference "pending V20 migration" for sm-kms or pki-ca.
  - [x] `optional_server_dirs` updated to reflect current reality.
  - [x] `lint-fitness` passes after update.

### Task 3.5: Phase 3 verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `lint-fitness` passes with smaller/cleaner exclusion maps.
  - [x] `go test ./internal/apps-tools/cicd_lint/lint_fitness/...` passes.
  - [x] Evidence archived in `test-output/v23-phase3/`.

## Phase 4: Tooling Quality Improvements

### Task 4.1: Implement mojibake sub-linter

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `internal/apps-tools/cicd_lint/lint_text/mojibake/mojibake.go` created.
  - [x] Linter scans `.md` files under `docs/` for windows-1252 mojibake byte sequences.
  - [x] Returns an error (not just warning) when mojibake markers are found.
  - [x] Does NOT false-positive on legitimate emoji or non-ASCII language characters.
  - [x] `mojibake_test.go` covers: clean file passes, mojibake file fails, empty file passes, non-md file skipped.

### Task 4.2: Register mojibake sub-linter in lint_text.go

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `mojibake.Check` added to the sub-linter table in `lint_text.go`.
  - [x] `go run ./cmd/cicd-lint lint-text` runs mojibake check and passes on current repo.
  - [x] `go test ./internal/apps-tools/cicd_lint/lint_text/...` passes.

### Task 4.3: Fix createFullPSIDRoot to parse required_server_dirs from MANIFEST

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `createFullPSIDRoot` in `apps_ps_id_template_test.go` reads `required_server_dirs` from the parsed MANIFEST.yaml struct.
  - [x] The hardcoded `[]string{"apis", "model", "repository"}` slice is removed.
  - [x] Test still creates exactly the dirs that MANIFEST specifies, respecting exclusion maps.
  - [x] `go test ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/...` passes.

### Task 4.4: Phase 4 verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `go run ./cmd/cicd-lint lint-text` passes (mojibake sub-linter included and passes).
  - [x] `go test ./internal/apps-tools/cicd_lint/...` passes.
  - [x] Evidence archived in `test-output/v23-phase4/`.

## Phase 5: Verification and Closure

### Task 5.1: Full quality suite

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `go build ./...` exits 0.
  - [x] `go build -tags e2e,integration ./...` exits 0.
  - [x] `go test ./...` exits 0.
  - [x] `golangci-lint run ./...` exits 0.
  - [x] `go run ./cmd/cicd-lint lint-go` exits 0.
  - [x] `go run ./cmd/cicd-lint lint-text` exits 0 (includes mojibake).
  - [x] `go run ./cmd/cicd-lint lint-fitness` exits 0.
  - [x] `go run ./cmd/cicd-lint lint-docs` exits 0.

### Task 5.2: Refresh inventory and contradiction review

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] Final inventory of changed surfaces archived under `test-output/v23-phase5/`.
  - [x] `docs/framework-v23/plan.md`, `docs/framework-v23/tasks.md`, and repository state are consistent.
  - [x] No linter exclusion remains that was originally tagged "pending V20 migration."
  - [x] If any item remains uncertain, mark it unresolved instead of guessing.

## Phase 6: Knowledge Propagation

### Task 6.1: Review execution lessons for durable guidance updates

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] `docs/framework-v23/lessons.md` reviewed after all execution phases complete.
  - [x] Any new durable guidance for handbook or instructions explicitly identified.

### Task 6.2: Apply durable documentation/process updates

- **Status**: ❌
- **Acceptance Criteria**:
  - [x] Durable doc/process updates from V21 applied to handbook or instructions.
  - [x] `go run ./cmd/cicd-lint lint-docs` passes after any updates.
  - [x] Evidence archived in `test-output/v23-phase6/`.
