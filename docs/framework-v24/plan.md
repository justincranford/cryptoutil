# Framework V24: PS-ID Convergence — Canonical Server Structure and 10-PS-ID Alignment

**Status**: Not started
**Created**: 2026-05-17
**Last Updated**: 2026-05-17
**Purpose**: Converge all 10 PS-ID services on the canonical thin-caller server structure.
Eliminate code duplication across services. Enforce conformity through fitness checks.
All 10 PS-IDs must be thin callers into the shared framework — no duplicated features, no
divergent directory layouts, no PS-ID-local reimplementation of framework capabilities.

## Problem Statement

The 10 PS-ID services have diverged in structure, directory layout, and internal organization.
Two services are missing canonical server subdirectories (`server/apis/`, `server/model/`,
`server/config/`). Several services carry transitional directories that should have been
retired. Linter exclusion maps contain stale entries that allow non-canonical layouts to
persist without detection. The canonical structure is defined but not fully enforced.

This plan implements the convergence: two pilot migrations (sm-kms, pki-ca), followed by a
tooling quality pass, and then a structured audit and alignment of all 10 PS-IDs.

## Canonical Structure Authority

The canonical recursive directory structure for all 10 PS-IDs is defined in
`docs/framework-v24/quizme-v2.md`. Key decisions from that research:

1. **`server/businesslogic/`** is the required canonical location for pure business logic.
   `server/service/` is not the standard — it is a transitional allowlist.
2. **`server/apis/`** is required for all PS-IDs. Missing services are migration gaps.
3. **`server/model/`** is required for all PS-IDs.
4. **`server/config/`** is required for all PS-IDs.
5. **`server/repository/`** is required for all PS-IDs.
6. **Transitional directories** (`handler/`, `service/`, `repository/orm/`) are allowlisted
   but must be retired: code migrated to canonical locations, then directories removed.
7. **Root-level packages** (`domain/`, `repository/`) are non-canonical once
   `server/repository/` is enforced.

## Current State

| PS-ID | Missing Canonical Dirs | Transitional Dirs Present |
|-------|----------------------|--------------------------|
| sm-kms | apis/, model/, config/ | businesslogic/, handler/, repository/orm/ |
| pki-ca | apis/, model/, repository/ | cmd/, middleware/ (accepted as-is) |
| sm-im | ✓ Complete | — |
| jose-ja | ✓ Complete | service/ (transitional) |
| skeleton-template | ✓ Complete | handler/ (transitional) |
| identity-authz | ✓ Complete | — |
| identity-idp | ✓ Complete | — |
| identity-rs | ✓ Complete | — |
| identity-rp | ✓ Complete | — |
| identity-spa | ✓ Complete | — |

## Phases

### Phase 1: sm-kms Server Subdirectory Migration (Pilot)

**Objective**: Reorganize `internal/apps/sm-kms/server/` to match the canonical template.
This is a pilot migration — learnings feed Phase 2.

Move OAS handler files from `server/handler/` into `server/apis/`.
Move ORM mapper (entity-layer) files from `server/businesslogic/` into `server/model/`.
Create `server/config/` with framework config wiring.
Update all import paths throughout sm-kms.
Confirm `businesslogic/` files that contain pure logic remain in place.

**Acceptance**:
- `go build ./internal/apps/sm-kms/...` clean.
- `go test ./internal/apps/sm-kms/...` passes.
- `lint-fitness` sm-kms exclusions for `apis`, `model`, `server/config`, `testmain_test.go` removed.

### Phase 2: pki-ca Server Subdirectory Migration (Pilot)

**Objective**: Add the three missing canonical subdirs to `internal/apps/pki-ca/server/`.

Create `server/apis/` with thin handler wrappers (delegating to existing domain layers).
Create `server/model/model.go` as the GORM entity struct home.
Create `server/repository/` with `migrations.go` + `server/repository/migrations/` directory.
Consolidate existing SQL migrations from `repository-v2/migrations/` into canonical path.
Create `server/config/config_test_helper.go`.

**Acceptance**:
- `go build ./internal/apps/pki-ca/...` clean.
- `go test ./internal/apps/pki-ca/...` passes.
- `lint-fitness` pki-ca exclusions for `apis`, `model`, `repository`, `config_test_helper.go`,
  `migrations.go` removed.

### Phase 3: Linter Exclusion Cleanup

**Objective**: Remove stale exclusions from `apps_ps_id_template.go` after Phases 1-2.

Remove sm-kms and pki-ca entries from all exclusion maps after migrations complete.

Remove stale `swagger.go`/`swagger_test.go` exclusions (all 10 PS-IDs already have these in
`server/` — the exclusion is a no-op and should be removed).

Move `*_lifecycle_test.go` and `*_port_conflict_test.go` to `server/` for remaining PS-IDs
that still have them at root level (sm-kms, jose-ja, pki-ca, skeleton-template), then remove
stale exclusions.

Update MANIFEST.yaml comments to remove any stale migration-pending notes.

**Acceptance**:
- `lint-fitness` passes with smaller exclusion maps.
- `go test ./internal/apps-tools/cicd_lint/lint_fitness/...` passes.

### Phase 4: Tooling Quality Improvements

**Objective**: Implement mojibake detection sub-linter and fix MANIFEST-driven test fixture.

4a. Implement `lint_text/mojibake` sub-linter: scans `.md` files in `docs/` for byte sequences
that indicate windows-1252-into-UTF-8 mojibake. Returns an error listing offending files and
line numbers. Must not false-positive on legitimate emoji or non-ASCII content.

4b. Register `mojibake` in `lint_text.go` alongside `utf8`.

4c. Update `apps_ps_id_template_test.go` `createFullPSIDRoot`: parse `required_server_dirs`
from MANIFEST.yaml instead of using the hardcoded `[]string{"apis", "model", "repository"}`.

**Acceptance**:
- `go run ./cmd/cicd-lint lint-text` passes (mojibake sub-linter exists and runs).
- `go test ./internal/apps-tools/cicd_lint/lint_text/...` passes.
- `go test ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/...` passes.

### Phase 5: All-10 PS-ID Canonical Structure Audit

**Objective**: For all 10 PS-IDs, verify compliance with the canonical structure defined in
`quizme-v2.md`. Produce a per-PS-ID gap report. Execute migrations for any non-pilot PS-IDs
that have gaps.

Tasks:
- Baseline inventory of each PS-ID's current recursive directory structure.
- Compare against canonical allowed-directory policy from `quizme-v2.md`.
- For transitional directories (handler/, service/, repository/orm/): plan sunset.
- For root-level non-canonical packages (domain/, root repository/): plan migration.
- Execute migrations per PS-ID in dependency order.
- Update MANIFEST.yaml `required_server_dirs` to enforce completed structure.
- Remove transitional allowlist entries from exclusion maps.

**Acceptance**:
- All 10 PS-IDs have `server/apis/`, `server/model/`, `server/config/`, `server/repository/`.
- No transitional directories remain unless explicitly approved with a retirement date.
- `lint-fitness apps-ps-id-template` passes with no exclusions for the above directories.

### Phase 6: Framework Duplication Audit

**Objective**: Identify code patterns duplicated across PS-IDs that should be extracted to
shared framework. Maximum reuse is the goal — PS-IDs should be thin callers, not duplicators.

Tasks:
- Audit each PS-ID's `server/apis/`, `server/config/`, `server/businesslogic/` for patterns
  duplicated across 3+ PS-IDs.
- For each duplicated pattern: create or extend a framework package in
  `internal/apps-framework/service/`.
- Migrate all PS-IDs that duplicate the pattern to call the shared implementation.
- Verify no behavioral regression.

**Acceptance**:
- Zero code patterns duplicated across 3+ PS-IDs where a framework equivalent exists.
- Each PS-ID's server layer is a thin wiring caller, not a duplicator of framework logic.

### Phase 7: Verification and Closure

**Objective**: Confirm all quality gates pass after all V24 work.

- `go build ./...` clean
- `go build -tags e2e,integration ./...` clean
- `go test ./...` passes
- `golangci-lint run ./...` clean
- `go run ./cmd/cicd-lint lint-go` passes
- `go run ./cmd/cicd-lint lint-text` passes (includes mojibake)
- `go run ./cmd/cicd-lint lint-fitness` passes
- `go run ./cmd/cicd-lint lint-docs` passes
- Docker Compose verification for each PS-ID that had server/ changes

### Phase 8: Knowledge Propagation

**Objective**: Capture V24 execution lessons and update durable guidance.

- Populate `docs/framework-v24/lessons.md` with per-phase post-mortems.
- Update `docs/ENG-HANDBOOK.md` if new durable patterns emerge.
- Update canonical templates in `api/cryptosuite-registry/templates/` if needed.
- Re-run `lint-docs` after any handbook/instruction changes.

## Quality Gates (All Phases)

- `go build ./...`
- `go build -tags e2e,integration ./...`
- `go test ./...`
- `golangci-lint run ./...`
- `go run ./cmd/cicd-lint lint-go`
- `go run ./cmd/cicd-lint lint-text`
- `go run ./cmd/cicd-lint lint-fitness`
- `go run ./cmd/cicd-lint lint-docs`

## Evidence Strategy

Archive V24 execution evidence under:
- `test-output/v24-phase1/`
- `test-output/v24-phase2/`
- `test-output/v24-phase3/`
- `test-output/v24-phase4/`
- `test-output/v24-phase5/`
- `test-output/v24-phase6/`
- `test-output/v24-phase7/`
- `test-output/v24-phase8/`

## Non-Goals

- Change sm-kms or pki-ca ORM/business logic semantics (structure only).
- Address `const-redefine-numeric` informational violations (these are informational).
- Implement behavioral test coverage for client/ stub packages (separate scope).
- Race detector on Windows (requires CGO, runs in CI only via `ci-race.yml`).

## Risks and Watchpoints

1. sm-kms `businesslogic/` is large (426 lines in businesslogic.go, 429 lines in
   businesslogic_crypto.go). The split into model + apis layers requires import-cycle checking.
2. pki-ca has non-canonical subdirs (`server/cmd/`, `server/middleware/`) that are accepted
   by the linter exclusion strategy. New dirs must not conflict with existing flat layout.
3. Removing `swagger.go` exclusions: verify the linter check logic validates presence of
   swagger.go using the `__SERVICE__` template pattern before removing the exclusion.
4. mojibake sub-linter false positives: the sub-linter must distinguish replacement-character
   sequences from legitimate multi-byte UTF-8 content (emoji, CJK, etc.).
5. Phase 6 (duplication audit) scope risk: duplication may be extensive. The phase must define
   a minimum-viable extraction scope to avoid infinite scope expansion.
6. Docker Compose verification: any Phase 1/2 migration that touches server-layer routing or
   initialization must be verified with `docker compose up --wait` + health endpoint check.

## Canonical Structure Reference

For complete canonical directory definitions, see: `docs/framework-v24/quizme-v2.md`.

Key decisions:
- `businesslogic/`: pure business logic — required canonical location, not deprecated.
- `handler/`, `service/`, `repository/orm/`: transitional allowlist — must sunset.
- `apis/`, `model/`, `config/`, `repository/`: required for all 10 PS-IDs.
- Root `domain/` and root `repository/`: non-canonical — migrate to `server/repository/`.

1. **Tooling quality debt** — Two items from V19 Action 2 and V18 Action 4:
   a. mojibake detection sub-linter in `lint_text` (V19 Action 2 — not implemented)
   b. `apps_ps_id_template` test fixture improvement: the hardcoded `[]string{"apis", "model",
      "repository"}` in `createFullPSIDRoot` should be parsed from MANIFEST.yaml `required_server_dirs`
      instead (V18 Action 4 — test still uses hardcoded list as of V21 start)

2. **All-10 PS-ID directory-shape convergence planning** — Based on Quizme Round 1 Q2 answer,
   V21 must produce a researched superset of recursive PS-ID directories and define the canonical
   recursive structure to be applied to all 10 PS-IDs. This includes explicit consolidation tasks
   for pki-ca package/subdirectory sprawl.

## Source Evidence

Work carried forward from:
- V18 lessons.md Actions 2, 4 (scaffolding tests, fixture improvement)
- V18 plan.md Phase 4 (sm-kms server/ deferred), Phase 5 (pki-ca server/ deferred)
- V19 lessons.md Actions 1 (stale exclusion audit), 2 (mojibake CI check)
- V20 plan.md Non-Goals (explicitly excluded sm-kms/pki-ca server/ migration)
- Current `apps_ps_id_template.go` exclusion maps (verified 2026-04-30)

## Affected Files

**sm-kms migration (Phase 1)**:
- `internal/apps/sm-kms/server/businesslogic/` — move contents to `server/apis/` + `server/model/`
- `internal/apps/sm-kms/server/handler/` — merge into `server/apis/`
- New: `internal/apps/sm-kms/server/config/config.go`, `config_test.go`, `config_test_helper.go`
- New: `internal/apps/sm-kms/server/model/model.go` (GORM entity structs)
- New/updated: `internal/apps/sm-kms/server/apis/` (handler files)

**pki-ca migration + consolidation planning (Phase 2)**:
- New: `internal/apps/pki-ca/server/apis/` (HTTP handler wrappers)
- New: `internal/apps/pki-ca/server/model/model.go`
- New: `internal/apps/pki-ca/server/repository/` + `migrations/`
- New: `internal/apps/pki-ca/server/config/config_test_helper.go`
- Existing sprawl to consolidate under canonical target:
  - `internal/apps/pki-ca/{api,bootstrap,compliance,crypto,domain,domain-v2,intermediate,observability,profile,security,service,storage}/`

**all-10 PS-ID canonicalization planning (Phase 2/3)**:
- `internal/apps/{sm-kms,sm-im,jose-ja,pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa,skeleton-template}/`
- `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml`
- `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template.go`

**Exclusion cleanup (Phase 3)**:
- `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template.go`
- `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml`
- `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template_test.go`

**Tooling quality (Phase 4)**:
- New: `internal/apps-tools/cicd_lint/lint_text/mojibake/mojibake.go` + `_test.go`
- `internal/apps-tools/cicd_lint/lint_text/lint_text.go` (register mojibake sub-linter)
- `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template_test.go`
  (drive `createFullPSIDRoot` from parsed MANIFEST required_server_dirs)

## Non-Goals

- Change sm-kms ORM/business logic semantics
- Address const-redefine-numeric informational violations (215 in identity/mfa tests) — these
  are informational and do not block CI
- Implement behavioral test coverage for client/ stub packages (separate planning scope)
- Race detector (Windows gcc toolchain gap — infrastructure blocker, not V21 scope)

## Key Decisions

1. sm-kms migration uses the canonical `server/apis/` layout: move all OAS handlers from
   `server/handler/` into `server/apis/`; move all GORM-entity structs from
   `server/businesslogic/` (the `oam_orm_mapper*.go` files) into `server/model/`.
   Pure business logic remains in `server/businesslogic/` and this subdirectory is treated as a
   required canonical location for pure business logic across PS-IDs (Quizme Round 1 Q1 answer).
2. pki-ca already has `server/cmd/`, `server/config/`, `server/middleware/` — these are
   non-standard but have been accepted as-is by the linter exclusion strategy. V21 adds the
   three missing canonical subdirs: `server/apis/`, `server/model/`, `server/repository/`.
3. pki-ca SQL migrations currently live in `internal/apps/pki-ca/repository-v2/migrations/`
   (`5001_ca_items.up.sql` and `5001_ca_items.down.sql` verified). V21 plans migration-path
   consolidation into `server/repository/migrations/` with explicit verification gates.
4. Quizme Round 1 Q4 selected active cleanup: move lifecycle/port-conflict tests into `server/`
   for remaining PS-IDs and remove corresponding stale exclusions once migrated.
5. The mojibake sub-linter checks `.md` files in `docs/` and `docs/framework-v*/` directories
   for the Unicode replacement markers `\xc3`, `\xe2`, `\xc2` (UTF-8 encodings of Ã, â, Â).
   It returns a lint error (not just a warning) when found outside of legitimate content.
6. The MANIFEST-driven fixture change in `apps_ps_id_template_test.go` parses
   `required_server_dirs` from MANIFEST.yaml and creates those directories in the synthetic
   root rather than the hardcoded `[]string{"apis", "model", "repository"}` slice.
7. Q2 remains open pending Quizme Round 2: canonical recursive directory structure for all 10
   PS-IDs and allowed-only directory policy will be finalized before broad refactor execution.

## Phase Summary

### Phase 1: sm-kms Server Subdirectory Migration

**Objective**: Reorganize `internal/apps/sm-kms/server/` to match the canonical template.

Move OAS handler files from `server/handler/` into `server/apis/`.
Move ORM mapper (entity-layer) files from `server/businesslogic/` into `server/model/`.
Create `server/config/` with framework config wiring.
Update all import paths throughout sm-kms.
Confirm `businesslogic.go` and related pure-logic files remain compilable.

**Success**:
- `go build ./internal/apps/sm-kms/...` clean.
- `go test ./internal/apps/sm-kms/...` passes.
- `lint-fitness` sm-kms exclusions for `apis`, `model`, `server/config`, `testmain_test.go` removed.

### Phase 2: pki-ca Server Subdirectory Migration

**Objective**: Add the three missing canonical subdirs to `internal/apps/pki-ca/server/`.

Create `server/apis/` with thin handler wrappers (delegating to existing domain layers).
Create `server/model/model.go` as the GORM entity struct home.
Create `server/repository/` with `migrations.go` + `server/repository/migrations/` directory.
Map existing `repository-v2/migrations/*.sql` into canonical `server/repository/migrations/`
with compatibility validation for embedded migration FS paths.
Create `server/config/config_test_helper.go`.
Draft a pki-ca consolidation map from current package sprawl to canonical allowed directories.

**Success**:
- `go build ./internal/apps/pki-ca/...` clean.
- `go test ./internal/apps/pki-ca/...` passes.
- `lint-fitness` pki-ca exclusions for `apis`, `model`, `repository`, `config_test_helper.go`,
  `migrations.go` removed.
- pki-ca consolidation map is complete with explicit per-subdirectory move/deprecate targets.

### Phase 3: Linter Exclusion Cleanup

**Objective**: Remove all stale exclusions from `apps_ps_id_template.go` after Phases 1-2.

Remove sm-kms and pki-ca entries from:
- `knownServerFileExclusions` (testmain_test.go for sm-kms)
- `knownServerDirExclusions` (apis, model for sm-kms and pki-ca; repository for pki-ca)
- `knownServerConfigFileExclusions` (config.go, config_test.go for sm-kms; config_test_helper.go for pki-ca and sm-kms)
- `knownServerRepositoryFileExclusions` (migrations.go for pki-ca)
- `knownServerRepositoryDirExclusions` (migrations for pki-ca)

Remove stale `swagger.go`/`swagger_test.go` exclusions (all 10 PS-IDs already have these in
`server/`; the exclusion map has been a no-op since at least V19).

Move `__SERVICE___lifecycle_test.go` and `__SERVICE___port_conflict_test.go` to `server/` for
remaining PS-IDs (sm-kms, jose-ja, pki-ca, skeleton-template), then remove stale exclusions.

Update MANIFEST.yaml comment to remove "pending V20 migration" references.

Update `apps_ps_id_template_test.go` `createFullPSIDRoot` to not hard-create the
`swagger.go`/`swagger_test.go` exclusion that is now empty.

**Success**:
- `lint-fitness` passes with smaller exclusion maps.
- `go test ./internal/apps-tools/...` passes.

### Phase 4: Tooling Quality Improvements

**Objective**: Implement mojibake detection sub-linter and fix MANIFEST-driven test fixture.

4a. Implement `lint_text/mojibake` sub-linter: scans `.md` files in `docs/` for byte sequences
`\xc3\x83`, `\xe2\x80`, `\xc2\x82` (common UTF-8 mojibake markers). Returns an error listing
offending files and line numbers.

4b. Register `mojibake` in `lint_text.go` alongside `utf8`.

4c. Update `apps_ps_id_template_test.go` `createFullPSIDRoot`: parse `required_server_dirs` from
MANIFEST.yaml and iterate that list to create server subdirectories, replacing the hardcoded
`[]string{"apis", "model", "repository"}` slice.

**Success**:
- `go run ./cmd/cicd-lint lint-text` passes (mojibake sub-linter exists and runs).
- `go test ./internal/apps-tools/cicd_lint/lint_text/...` passes with mojibake tests.
- `go test ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/...` passes with
  MANIFEST-driven fixture.

### Phase 5: Verification and Closure

**Objective**: Confirm all quality gates pass after V21 work.

- `go build ./...` clean
- `go build -tags e2e,integration ./...` clean
- `go test ./...` passes
- `golangci-lint run ./...` clean
- `go run ./cmd/cicd-lint lint-fitness` passes
- `go run ./cmd/cicd-lint lint-go` passes (no new blocking violations)
- `go run ./cmd/cicd-lint lint-text` passes
- `go run ./cmd/cicd-lint lint-docs` passes

### Phase 6: Knowledge Propagation

**Objective**: Capture V21 execution lessons and update durable guidance.

- Populate `docs/framework-v23/lessons.md` with per-phase post-mortems.
- Update `docs/ENG-HANDBOOK.md` if new durable patterns emerge.
- Re-run `lint-docs` after any handbook/instruction changes.

## Quality Gates

- `go build ./...`
- `go build -tags e2e,integration ./...`
- `go test ./...`
- `golangci-lint run ./...`
- `go run ./cmd/cicd-lint lint-go`
- `go run ./cmd/cicd-lint lint-text`
- `go run ./cmd/cicd-lint lint-fitness`
- `go run ./cmd/cicd-lint lint-docs`

## Evidence Strategy

Archive V21 execution evidence under:
- `test-output/v23-phase1/`
- `test-output/v23-phase2/`
- `test-output/v23-phase3/`
- `test-output/v23-phase4/`
- `test-output/v23-phase5/`
- `test-output/v23-phase6/`

## Risks and Watchpoints

1. sm-kms `businesslogic/` is large (businesslogic.go 426 lines, businesslogic_crypto.go 429 lines,
   multiple test files). Splitting into model + apis layers requires careful import-cycle checking.
2. pki-ca server/ has non-canonical subdirectories (cmd/, middleware/) that are already accepted
   by the linter. Creating apis/model/repository must not conflict with the existing flat layout.
3. Removing swagger.go exclusions: if the linter `checkServerFiles` expects swagger.go to exist
   in server/ AND it is now present, removing the exclusion may expose a validation gap where the
   linter checks for the file name using the **SERVICE** template pattern — verify the check logic
   before removing.
4. mojibake sub-linter false positives: emoji, some language characters, and intentional UTF-8
   content may trigger byte-sequence matches. The sub-linter must only flag the specific
   replacement-character sequences that indicate windows-1252-into-UTF-8 mojibake.
5. Canonical all-10 PS-ID directory-shape decision is not finalized until Quizme Round 2 closes.

## Quizme Round 1 (2026-04-30)

### Q1: sm-kms businesslogic split strategy

- **Question**: Should pure business logic stay in `server/businesslogic/`, move to `server/service/`, or another approach?
- **Answer**: D
- **Applied decision**: `server/businesslogic/` is the required canonical location for pure business logic; `server/service/` is not adopted.

### Q2: pki-ca APIs layer design

- **Question**: Should V21 use a placeholder `server/apis/`, full wrappers, or another approach?
- **Answer**: D
- **Applied decision**: Requires follow-up via Quizme Round 2 after researched superset of recursive PS-ID directory structures and canonical-all-10 proposal.

### Q3: pki-ca repository and SQL migrations scope

- **Question**: Where are current SQL migrations and how should V21 proceed?
- **Answer**: D
- **Applied decision**: Research confirmed SQL migration files at `internal/apps/pki-ca/repository-v2/migrations/5001_ca_items.up.sql` and `internal/apps/pki-ca/repository-v2/migrations/5001_ca_items.down.sql`; plan includes consolidation mapping into canonical `server/repository/migrations/`.

### Q4: lifecycle_test and port_conflict_test cleanup

- **Question**: Should V21 migrate these remaining root-level tests to `server/` now?
- **Answer**: A
- **Applied decision**: Include migration for remaining PS-IDs (sm-kms, jose-ja, pki-ca, skeleton-template) and remove stale linter exclusions after verification.
