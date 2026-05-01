# Implementation Plan - Framework V21: Server Subdirectory Migration and Tooling Debt Closure

**Status**: Not started
**Created**: 2026-04-30
**Last Updated**: 2026-04-30
**Purpose**: Complete deferred structural migration for sm-kms and pki-ca server/ subdirectories,
clean up stale linter exclusions, add mojibake detection, and improve test fixture quality.
All items are carried-forward incomplete work from V16-V20.

## Overview

Framework V21 closes four categories of deferred work identified across V16-V20 plans:

1. **sm-kms server/ structural migration** — `businesslogic/` and `handler/` reorganized into
   `server/apis/`, `server/model/`, `server/config/`. Explicitly deferred from V18 to V20, then
   skipped in V20 (TLS-only scope). The `apps-ps-id-template` linter still carries sm-kms
   exclusions for `apis`, `model`, `server/config`, and `testmain_test.go`.

2. **pki-ca server/ structural migration** — pki-ca lacks `server/apis/`, `server/model/`,
   `server/repository/`, `server/config/config_test_helper.go`. The linter carries matching
   exclusions. The pki-ca CA architecture is complex (bootstrap, compliance, intermediate,
   issuer, storage) and was deferred from V18 citing complexity.

3. **Linter exclusion cleanup** — After migrations complete, `knownServerFileExclusions`,
   `knownServerDirExclusions`, `knownServerConfigFileExclusions`,
   `knownServerRepositoryFileExclusions`, and `knownServerRepositoryDirExclusions` entries
   for sm-kms and pki-ca can be removed. Additionally, `swagger.go`/`swagger_test.go` are
   already in `server/` for all 10 PS-IDs (the exclusion is stale since V18/V19),
   and the MANIFEST.yaml comments still reference "pending V20 migration."

4. **Tooling quality debt** — Two items from V19 Action 2 and V18 Action 4:
   a. mojibake detection sub-linter in `lint_text` (V19 Action 2 — not implemented)
   b. `apps_ps_id_template` test fixture improvement: the hardcoded `[]string{"apis", "model",
      "repository"}` in `createFullPSIDRoot` should be parsed from MANIFEST.yaml `required_server_dirs`
      instead (V18 Action 4 — test still uses hardcoded list as of V21 start)

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

**pki-ca migration (Phase 2)**:
- New: `internal/apps/pki-ca/server/apis/` (HTTP handler wrappers)
- New: `internal/apps/pki-ca/server/model/model.go`
- New: `internal/apps/pki-ca/server/repository/` + `migrations/`
- New: `internal/apps/pki-ca/server/config/config_test_helper.go`

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

- Re-architect pki-ca CA domain logic (bootstrap, compliance, intermediate, issuer layers)
- Change sm-kms ORM/business logic semantics
- Address const-redefine-numeric informational violations (215 in identity/mfa tests) — these
  are informational and do not block CI
- Implement behavioral test coverage for client/ stub packages (separate planning scope)
- Race detector (Windows gcc toolchain gap — infrastructure blocker, not V21 scope)

## Key Decisions

1. sm-kms migration uses the canonical `server/apis/` layout: move all OAS handlers from
   `server/handler/` into `server/apis/`; move all GORM-entity structs from
   `server/businesslogic/` (the `oam_orm_mapper*.go` files) into `server/model/`.
   Pure business logic in `server/businesslogic/` is preserved in place or moved to a new
   `server/service/` subdirectory if needed — the linter only requires `apis/`, `model/`, and
   `repository/` to exist.
2. pki-ca already has `server/cmd/`, `server/config/`, `server/middleware/` — these are
   non-standard but have been accepted as-is by the linter exclusion strategy. V21 adds the
   three missing canonical subdirs: `server/apis/`, `server/model/`, `server/repository/`.
3. The mojibake sub-linter checks `.md` files in `docs/` and `docs/framework-v*/` directories
   for the Unicode replacement markers `\xc3`, `\xe2`, `\xc2` (UTF-8 encodings of Ã, â, Â).
   It returns a lint error (not just a warning) when found outside of legitimate content.
4. The MANIFEST-driven fixture change in `apps_ps_id_template_test.go` parses
   `required_server_dirs` from MANIFEST.yaml and creates those directories in the synthetic
   root rather than the hardcoded `[]string{"apis", "model", "repository"}` slice.

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
Create `server/config/config_test_helper.go`.

**Success**:
- `go build ./internal/apps/pki-ca/...` clean.
- `go test ./internal/apps/pki-ca/...` passes.
- `lint-fitness` pki-ca exclusions for `apis`, `model`, `repository`, `config_test_helper.go`,
  `migrations.go` removed.

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

Remove stale `__SERVICE___lifecycle_test.go` and `__SERVICE___port_conflict_test.go` exclusions
for PS-IDs that now have those tests in `server/` (identity-*, sm-im already migrated; sm-kms and
jose-ja and pki-ca and skeleton-template still have them at root — keep those entries).

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

- Populate `docs/framework-v21/lessons.md` with per-phase post-mortems.
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
- `test-output/v21-phase1/`
- `test-output/v21-phase2/`
- `test-output/v21-phase3/`
- `test-output/v21-phase4/`
- `test-output/v21-phase5/`
- `test-output/v21-phase6/`

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
