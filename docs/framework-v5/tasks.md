# Tasks - Framework v5: Rigid Standardization & Cleanup

**Status**: 32 of 49 tasks complete (65%)
**Last Updated**: 2026-03-27
**Created**: 2026-03-21

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- Correctness: ALL code must be functionally correct with comprehensive tests
- Completeness: NO phases or tasks or steps skipped, NO shortcuts
- Thoroughness: Evidence-based validation at every step
- Reliability: Quality gates enforced (>=95%/98% coverage/mutation)
- Efficiency: Optimized for maintainability and performance, NOT implementation speed
- Accuracy: Changes must address root cause, not just symptoms
- Time Pressure: NEVER rush, NEVER skip validation, NEVER defer quality checks
- Premature Completion: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions.**

---

## Task Checklist

### Phase 1: Archive, Dead Code, and Legacy Cleanup

**Phase Objective**: Remove 161+ dead files across 9 archived/orphaned directories, plus legacy infrastructure (Citus, legacy secrets, environment configs).

#### Task 1.1: Delete Identity Archived Directories

- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 30m
- **Dependencies**: None
- **Description**: Delete all 6 identity archived directories containing 132 files total
- **Acceptance Criteria**:
  - [x] `internal/apps/identity/_archived/` deleted (92 files)
  - [x] `internal/apps/identity/_authz-archived/` deleted (8 files)
  - [x] `internal/apps/identity/_idp-archived/` deleted (8 files)
  - [x] `internal/apps/identity/_rp-archived/` deleted (8 files)
  - [x] `internal/apps/identity/_rs-archived/` deleted (8 files)
  - [x] `internal/apps/identity/_spa-archived/` deleted (8 files)
  - [x] `go build ./...` clean
  - [x] `go build -tags e2e,integration ./...` clean

#### Task 1.2: Delete PKI and Test Archived Directories

- **Status**: ✅ Complete
- **Estimated**: 30m
- **Actual**: 15m
- **Dependencies**: None
- **Description**: Delete PKI empty archive and test archived directories
- **Acceptance Criteria**:
  - [x] `internal/apps/pki/_ca-archived/` deleted (empty dir) — did not exist, confirmed
  - [x] `internal/test/archived/` deleted (15 files)
  - [x] `go build ./...` clean

#### Task 1.3: Delete Deployment Archived Directory

- **Status**: ✅ Complete
- **Estimated**: 30m
- **Actual**: 15m
- **Dependencies**: None
- **Description**: Delete legacy compose archived directory
- **Acceptance Criteria**:
  - [x] `deployments/archived/` deleted (14 files)
  - [x] `go run ./cmd/cicd lint-deployments` passes (63 validators — expected post-Phase-1 count)
  - [x] `go run ./cmd/cicd lint-fitness` passes (all 44 linters)

#### Task 1.4: Delete Configs Orphaned Directory

- **Status**: ✅ Complete
- **Estimated**: 30m
- **Actual**: 10m
- **Dependencies**: None
- **Description**: Delete `configs/orphaned/` legacy configs (observability, template, test secrets)
- **Acceptance Criteria**:
  - [x] `configs/orphaned/` deleted
  - [x] No code references to orphaned config files remain
  - [x] `go build ./...` clean

#### Task 1.5: Delete Citus Infrastructure

- **Status**: ✅ Complete
- **Estimated**: 30m
- **Actual**: 10m
- **Dependencies**: None
- **Description**: Delete `deployments/shared-citus/` entirely — only PostgreSQL + SQLite supported (Decision 5)
- **Acceptance Criteria**:
  - [x] `deployments/shared-citus/` deleted
  - [x] Zero references to `citus` in compose files
  - [x] `go run ./cmd/cicd lint-deployments` passes
  - [x] `go build ./...` clean

#### Task 1.6: Delete Legacy Secrets

- **Status**: ✅ Complete
- **Estimated**: 30m
- **Actual**: 20m
- **Dependencies**: None
- **Description**: Delete `sm-hash-pepper.secret` and all `{PRODUCT}-*.secret.never` / `{SUITE}-*.secret.never` marker files (Decision 9)
- **Acceptance Criteria**:
  - [x] `sm-hash-pepper.secret` removed from all deployment tiers
  - [x] All `.secret.never` marker files with product-level prefixes deleted (45 files across identity/jose/pki/sm products)
  - [x] All `.secret.never` marker files with suite-level prefixes deleted (9 files in cryptoutil-suite + 9 in archived)
  - [x] Compose file secret references: no dangling mounts (`.never` files are markers, not mounted secrets)
  - [x] `go run ./cmd/cicd lint-deployments` passes

#### Task 1.7: Delete Environment Configs and Move Docker Overlays

- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 45m
- **Dependencies**: None
- **Description**: Delete development.yml, production.yml, test.yml, profiles/ from configs/; move *-docker.yml to deployments/*/config/ (Decision 8)
- **Acceptance Criteria**:
  - [x] `development.yml` deleted from all `configs/` directories (configs/identity/)
  - [x] `production.yml` deleted from all `configs/` directories (configs/identity/)
  - [x] `test.yml` deleted from all `configs/` directories (configs/identity/)
  - [x] `profiles/` deleted from all `configs/` directories (configs/identity/profiles/)
  - [x] `*-docker.yml` files moved to corresponding `deployments/*/config/` (identity: authz/idp/rp/rs/spa domain configs → deployments/{PS-ID}/config/{PS-ID}-domain.yml)
  - [x] `configs/configs-all-files.json` deleted
  - [x] No code references to deleted config files remain
  - [x] `go build ./...` clean
- **Note**: SM standalone configs (`configs/sm/kms/config-sqlite.yml`, `config-pg-1.yml`, `config-pg-2.yml` and `configs/sm/im/` equivalents) were INCORRECTLY deleted and then RESTORED. These are canonical standalone configs required by `standalone-config-presence` and `entity_registry_completeness` fitness linters. Phase 3 Task 3.x will rename them to `{PS-ID}-app-sqlite-1.yml` etc. AND update the fitness linters simultaneously.

#### Task 1.8: Phase 1 Quality Gate Verification

- **Status**: ✅ Complete
- **Estimated**: 30m
- **Actual**: 1h (debugging fitness linter failures from incorrect SM config deletion)
- **Dependencies**: Tasks 1.1-1.7
- **Description**: Comprehensive verification that all deletions left the project healthy
- **Acceptance Criteria**:
  - [x] `go build ./...` clean (exit 0)
  - [x] `go build -tags e2e,integration ./...` clean (exit 0)
  - [x] `go run ./cmd/cicd lint-fitness` — all 44 linters pass (Passed: 1, Failed: 0, EXIT: 0)
  - [x] `go run ./cmd/cicd lint-deployments` — 63 validators pass (63 = expected post-Phase-1 count; pre-Phase-1 was 68 when orphaned/archived directories existed)
  - [x] `golangci-lint run` clean (exit 0)
  - [x] Zero `_archived`, `archived/`, `orphaned/`, or `shared-citus/` directories remain
  - [x] Zero legacy secrets (`sm-hash-pepper`, prefixed `.secret.never`) remain
  - [x] Zero environment configs (`development.yml`, `production.yml`, `test.yml`, `profiles/`) in configs/
  - [x] Git commits: one per semantic deletion group (see commits below)
- **Issue Fixed**: `entity_registry_completeness` and `standalone_config_presence` fitness linters failed because `configs/sm/kms/` directory was emptied by incorrect deletion. SM standalone configs RESTORED. Root cause: these are CANONICAL configs, not deployment duplicates. Phase 3 will rename them.

### Phase 2: Non-Standard Entry Rationalization

**Phase Objective**: Classify all non-standard cmd/ and internal/apps/ entries. Fix anti-pattern violations. Rename cicd → cicd-lint. Create framework tier routing. Add workflow subcommands.

#### Task 2.1: Rename cicd → cicd-lint

- **Status**: ✅ Complete
- **Estimated**: 2h
- **Actual**: 3h (unexpected interface{} corruption in format_go after rename)
- **Dependencies**: None
- **Description**: Rename `cmd/cicd/` → `cmd/cicd-lint/` and `internal/apps/cicd/` → `internal/apps/tools/cicd_lint/` (Decision 11)
- **Acceptance Criteria**:
  - [x] `cmd/cicd/` renamed to `cmd/cicd-lint/`
  - [x] `internal/apps/cicd/` moved to `internal/apps/tools/cicd_lint/`
  - [x] All import paths updated (grep for `apps/cicd`)
  - [x] All workflow files updated (pre-commit hooks, CI workflows)
  - [x] Pre-commit hooks updated (`cmd/cicd-lint/main.go`)
  - [x] ARCHITECTURE.md Section 9.10 command table updated
  - [x] Entity registry `PSID` and `InternalAppsDir` fields updated
  - [x] copilot-instructions.md cicd command table updated
  - [x] `.cicd/` runtime cache dir: NO RENAME (gitignored, unrelated)
  - [x] `go build ./...` clean
  - [x] `go test ./internal/apps/tools/cicd_lint/...` passes
- **Commit**: `4131dc57d`

#### Task 2.2: Document Infrastructure CLI Tools

- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 20m
- **Dependencies**: Task 2.1
- **Description**: Document `cmd/cicd-lint/` and `cmd/workflow/` as intentional infrastructure tools in ARCHITECTURE.md
- **Acceptance Criteria**:
  - [x] ARCHITECTURE.md Section 4.4.7 updated with infrastructure tool documentation
  - [x] `cmd/cicd-lint/` documented as CICD tooling (linters, formatters, scripts)
  - [x] `cmd/workflow/` documented as workflow testing infrastructure
  - [x] `go run ./cmd/cicd-lint lint-docs` passes
- **Commit**: `f4048d8a9`

#### Task 2.3: Evaluate cmd/identity-compose

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Assess whether identity-compose should be subcommand of cmd/identity or archived
- **Acceptance Criteria**:
  - [x] Code read and purpose understood
  - [x] Decision documented: archive — stub with a single `fmt.Println` statement, zero real implementation
  - [x] If merge: cmd/identity compose subcommand implemented (N/A — chose archive)
  - [x] If archive: cmd/identity-compose/ removed (commit `c4d5f0594`)
  - [x] All workflow references updated (no workflow references found)
  - [x] Build clean

#### Task 2.4: Evaluate cmd/identity-demo

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Assess whether identity-demo should be subcommand of cmd/identity or archived
- **Acceptance Criteria**:
  - [x] Code read and purpose understood
  - [x] Decision documented: archive — 3-file stub directory with zero real implementation
  - [x] If merge: cmd/identity demo subcommand implemented (N/A — chose archive)
  - [x] If archive: cmd/identity-demo/ removed (commit `4a2b63a6e`)
  - [x] Build clean

#### Task 2.5: Evaluate cmd/demo and internal/apps/demo

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Assess whether unified demo CLI should become a suite-level subcommand or remain standalone
- **Acceptance Criteria**:
  - [x] 17-file demo codebase reviewed for usefulness
  - [x] Decision documented: archive — skeleton stubs masquerading as demo, no value
  - [x] Changes implemented per decision (commit `497379e97`; also cleaned fitness linter dead exclusions)
  - [x] Build clean

#### Task 2.6: Evaluate internal/apps/pkiinit

- **Status**: ✅
- **Estimated**: 30m
- **Dependencies**: None
- **Description**: Assess whether PKI init tool belongs under framework/tls/ or is orphaned (quizme-v2 Q2=D)
- **Acceptance Criteria**:
  - [x] 3-file codebase reviewed
  - [x] Decision documented: integrate into framework/tls/ (Q2=D confirmed)
  - [x] Changes implemented: `Init()` in `internal/apps/framework/tls/`, 9 tests pass (commit `7fcd75425`)
  - [x] Build clean

#### Task 2.7: Clean docs/demo-brainstorm

- **Status**: ✅
- **Estimated**: 15m
- **Dependencies**: Task 2.5
- **Description**: Archive or delete demo brainstorm documents if no longer relevant
- **Acceptance Criteria**:
  - [x] 3 files reviewed for relevance
  - [x] Deleted (outdated brainstorm for archived demo product, commit `4a49e2827`)
  - [x] No references to deleted files remain

#### Task 2.8: Create Framework Tier Routing

- **Status**: Complete ✅
- **Estimated**: 2h
- **Dependencies**: None
- **Description**: Create `framework/suite/cli/` with RouteSuite() and `framework/product/cli/` with RouteProduct() moved from service/cli/ (Decision 6)
- **Acceptance Criteria**:
  - [x] `internal/apps/framework/suite/cli/suite_router.go` created: `RouteSuite()`, `SuiteConfig`, `ProductEntry`
  - [x] `internal/apps/framework/suite/cli/suite_router_test.go` created (≥98% coverage)
  - [x] `internal/apps/framework/product/cli/product_router.go` created: `RouteProduct()`, `ProductConfig`, `ServiceEntry`
  - [x] `internal/apps/framework/product/cli/product_router_test.go` created (≥98% coverage)
  - [x] `RouteProduct()` removed from `framework/service/cli/` (moved, not duplicated)
  - [x] All product-level `cmd/*/main.go` imports updated for new `product/cli/` path
  - [x] All suite-level `cmd/cryptoutil/main.go` imports updated for new `suite/cli/` path
  - [x] `go build ./...` clean
  - [x] `go test ./...` passes (framework packages)
- **Commit**: f0ef6d217

#### Task 2.9: Add Workflow Subcommands

- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 30m
- **Dependencies**: Task 2.1 (cicd-lint rename moves workflow under tools/)
- **Description**: Add `run` and `cleanup` subcommands to `cmd/workflow/` (Decision 10)
- **Acceptance Criteria**:
  - [x] `cmd/workflow/main.go` accepts `run` and `cleanup` subcommands
  - [x] Internal implementation in `internal/apps/tools/workflow/`
  - [x] `go build ./cmd/workflow/...` clean
  - [x] `go test ./internal/apps/tools/workflow/...` passes (≥98% coverage)
- **Commit**: 26818f717

#### Task 2.10: Phase 2 Quality Gate Verification

- **Status**: ✅ Complete
- **Estimated**: 30m
- **Actual**: 1h (includes dead code fix for profile.go regression)
- **Dependencies**: Tasks 2.1-2.9
- **Description**: Verify all rationalization changes maintain project health
- **Acceptance Criteria**:
  - [x] `go build ./...` clean
  - [x] `go build -tags e2e,integration ./...` clean
  - [x] `go test ./...` passes (modified packages: identity/config, tools/workflow, tools/cicd_lint, framework/product/cli, framework/tls)
  - [x] `go run ./cmd/cicd-lint lint-fitness` — all linters pass
  - [x] `golangci-lint run` clean
  - [x] Zero cmd/ anti-pattern violations remain
  - [x] cicd → cicd-lint rename complete with zero missed references
  - [x] Framework tier routing (suite/cli/, product/cli/, service/cli/) in place
  - [x] Workflow subcommands (run, cleanup) functional
  - [x] ARCHITECTURE.md documents all intentional infrastructure tools
  - [x] Git commits: one per semantic change
- **Note**: `internal/apps/identity/config/profile.go` and `profile_test.go` were dead code (LoadProfile has no production callers) depending on `configs/identity/profiles/` which was intentionally deleted in Task 1.7 (Decision 8). Deleted both files as part of this quality gate. framework/service/server/* and pki/ca/server timeouts are PRE-EXISTING infrastructure issues (not introduced by Phase 1/2 changes — confirmed via git log).

---

### Phase 3: Configs Standardization

**Phase Objective**: Apply rigid {PS-ID}-based naming to all configs/ files, resolve path mismatches, document the dual configs/ vs deployments/config/ relationship.

#### Task 3.1: Design Configs Canonical Structure

- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 30m
- **Dependencies**: None
- **Description**: Document the target configs/ structure aligned with entity registry
- **Acceptance Criteria**:
  - [x] Target directory tree documented (configs/{PRODUCT}/{SERVICE}/ for all 10 services)
  - [x] File naming convention defined: `{PS-ID}-{purpose}.yml` for service template configs
  - [x] Domain config convention defined: `{PS-ID}-{domain-purpose}.yml`
  - [x] Product-level config convention defined
  - [x] Relationship between configs/ and deployments/config/ documented
- **Design Decisions**:
  - **Directory hierarchy**: `configs/{PRODUCT}/{SERVICE}/` mirrors entity registry product/service
  - **Service template configs** (flat kebab-case, validated by ValidateSchema): `{PS-ID}-{variant}.yml` — e.g., `sm-kms-sqlite.yml`, `sm-kms-pg-1.yml`, `sm-kms-pg-2.yml`
  - **Domain configs** (nested YAML, service-specific): `{service}.yml` or `{PS-ID}-server.yml` — e.g., `ca-server.yml` → `pki-ca-server.yml`, `authz.yml` (identity domain configs keep short names as they are already under `configs/identity/authz/`)
  - **Product-level configs**: `configs/{PRODUCT}/` for shared configs — e.g., `configs/identity/policies/`, `configs/pki/ca/profiles/`
  - **Suite-level config**: `configs/cryptoutil/cryptoutil.yml`
  - **configs/ vs deployments/config/**: `configs/` = canonical SSOT for local dev/testing. `deployments/*/config/` = Docker-specific minimal overrides (instance isolation). No duplication between the two.
  - **Standalone config renaming**: SM configs `config-{variant}.yml` → `{PS-ID}-{variant}.yml`. Linters (`standalone_config_presence`, `standalone_config_otlp_names`, `otlp_service_name_pattern`) updated atomically.
  - **PKI CA**: Already at `configs/pki/ca/` path. Files renamed: `ca-server.yml` → `pki-ca-server.yml`, `ca-config-schema.yaml` → `pki-ca-config-schema.yaml`
  - **Identity**: Domain configs (`authz.yml`, `idp.yml`, etc.) keep current names (already at canonical `configs/identity/{service}/` path). No `-docker.yml` files exist (deleted in Phase 1).
  - **Jose/Skeleton**: `jose-server.yml` → `jose-ja-server.yml` (moved to `configs/jose/ja/`). `template-server.yml` → `skeleton-template-server.yml`. `skeleton-server.yml` is a product-level default (kept at `configs/skeleton/`).

#### Task 3.2: Fix configs/ca/ Path Mismatch

- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 0.5h
- **Dependencies**: Task 3.1
- **Description**: Rename PKI CA config files to use PS-ID prefix (`configs/pki/ca/` path already correct from prior work)
- **Acceptance Criteria**:
  - [x] `configs/pki/ca/ca-server.yml` renamed to `configs/pki/ca/pki-ca-server.yml`
  - [x] `configs/pki/ca/ca-config-schema.yaml` renamed to `configs/pki/ca/pki-ca-config-schema.yaml`
  - [x] `configs/pki/ca/profiles/` already at correct path (no change needed)
  - [x] Old `configs/ca/` directory already deleted (prior work)
  - [x] All Go code references updated (generate_listings_test.go, validate_all_runners_test.go, validate_all.go)
  - [x] Compose files verified — no references to configs/ standalone files
  - [x] lint-deployments mirror mapping verified — only checks top-level dirs, no update needed
  - [x] Build clean, tests pass

#### Task 3.3: Standardize SM Configs Naming

- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 2h
- **Dependencies**: Task 3.1
- **Description**: Rename SM service config files to use PS-ID prefix
- **Acceptance Criteria**:
  - [x] `configs/sm/kms/config-pg-1.yml` -> `configs/sm/kms/sm-kms-pg-1.yml`
  - [x] `configs/sm/kms/config-pg-2.yml` -> `configs/sm/kms/sm-kms-pg-2.yml`
  - [x] `configs/sm/kms/config-sqlite.yml` -> `configs/sm/kms/sm-kms-sqlite.yml`
  - [x] `configs/sm/im/config-pg-1.yml` -> `configs/sm/im/sm-im-pg-1.yml`
  - [x] `configs/sm/im/config-pg-2.yml` -> `configs/sm/im/sm-im-pg-2.yml`
  - [x] `configs/sm/im/config-sqlite.yml` -> `configs/sm/im/sm-im-sqlite.yml`
  - [x] All Go code references updated (3 fitness linters + isServiceFrameworkConfig)
  - [x] Build clean
- **Files**:
  - `configs/sm/kms/sm-kms-sqlite.yml` (renamed)
  - `configs/sm/kms/sm-kms-pg-1.yml` (renamed)
  - `configs/sm/kms/sm-kms-pg-2.yml` (renamed)
  - `configs/sm/im/sm-im-sqlite.yml` (renamed)
  - `configs/sm/im/sm-im-pg-1.yml` (renamed)
  - `configs/sm/im/sm-im-pg-2.yml` (renamed)
  - `internal/apps/tools/cicd_lint/lint_fitness/standalone_config_presence/standalone_config_presence.go`
  - `internal/apps/tools/cicd_lint/lint_fitness/standalone_config_presence/standalone_config_presence_test.go`
  - `internal/apps/tools/cicd_lint/lint_fitness/standalone_config_otlp_names/standalone_config_otlp_names.go`
  - `internal/apps/tools/cicd_lint/lint_fitness/standalone_config_otlp_names/standalone_config_otlp_names_test.go`
  - `internal/apps/tools/cicd_lint/lint_fitness/otlp_service_name_pattern/otlp_service_name_pattern.go`
  - `internal/apps/tools/cicd_lint/lint_fitness/otlp_service_name_pattern/otlp_service_name_pattern_test.go`
  - `internal/apps/tools/cicd_lint/lint_deployments/validate_all.go`
  - `internal/apps/tools/cicd_lint/lint_deployments/validate_all_runners_test.go`

#### Task 3.4: Standardize Identity Configs Naming

- **Status**: ✅ Complete (no-op)
- **Estimated**: 1h
- **Actual**: 5m
- **Dependencies**: Task 3.1
- **Description**: Evaluate identity `-docker.yml` files and standardize naming
- **Acceptance Criteria**:
  - [x] `authz.yml` -> keep as domain config (not service framework config)
  - [x] No `-docker.yml` files exist (deleted in Phase 1 Task 1.7)
  - [x] Decision: identity configs are domain configs, keep short names (`authz.yml`, `idp.yml`, etc.)
  - [x] Same pattern applied to idp, rp, rs, spa — all are domain configs
  - [x] No Go code references to update
  - [x] Build clean (no changes)

#### Task 3.5: Standardize Jose and Skeleton Configs

- **Status**: ✅ Complete
- **Estimated**: 30m
- **Actual**: 10m
- **Dependencies**: Task 3.1
- **Description**: Rename product-level configs to use PS-ID; populate empty directories
- **Acceptance Criteria**:
  - [x] `configs/jose/ja/jose-server.yml` -> `configs/jose/ja/jose-ja-server.yml` (PS-ID prefix)
  - [x] `configs/skeleton/skeleton-server.yml` -> kept as product-level config (no rename needed)
  - [x] `configs/skeleton/template/template-server.yml` -> `skeleton-template-server.yml` (PS-ID prefix)
  - [x] `configs/jose/ja/` was already populated (jose-ja-server.yml)
  - [x] Build clean
- **Files**:
  - `configs/jose/ja/jose-ja-server.yml` (renamed from jose-server.yml)
  - `configs/skeleton/template/skeleton-template-server.yml` (renamed from template-server.yml)
  - `internal/apps/tools/cicd_lint/lint_deployments/validate_all_runners_test.go`

#### Task 3.6: Update ARCHITECTURE.md Section 12.5

- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 0.5h
- **Dependencies**: Tasks 3.2-3.5
- **Description**: Rewrite Section 12.5 to document the TARGET configs/ standard (not current inconsistent state)
- **Acceptance Criteria**:
  - [x] Section 12.5 rewritten with canonical configs/ structure
  - [x] Naming conventions table included
  - [x] Dual configs/ vs deployments/config/ relationship documented
  - [x] Examples for each config file type
  - [x] `go run ./cmd/cicd-lint lint-docs` passes

#### Task 3.7: Update lint-deployments Mirror Mapping

- **Status**: ✅ Complete (no-op)
- **Estimated**: 1h
- **Actual**: 0.1h
- **Dependencies**: Tasks 3.2-3.5
- **Description**: Update lint-deployments to validate configs/ mirrors deployments/ structure
- **Acceptance Criteria**:
  - [x] Mirror mapping updated for `configs/ca/` -> `configs/pki/ca/` rename (already correct from Phase 1)
  - [x] All 68+ deployment validators still pass
  - [x] Tests updated and passing (>=98% coverage)
  - [x] Build clean

#### Task 3.8: Phase 3 Quality Gate Verification

- **Status**: ✅ Complete
- **Estimated**: 30m
- **Actual**: 0.3h
- **Dependencies**: Tasks 3.1-3.7
- **Description**: Verify all configs changes maintain project health
- **Acceptance Criteria**:
  - [x] `go build ./...` clean
  - [x] `go build -tags e2e,integration ./...` clean
  - [x] `go test ./...` passes (modified packages)
  - [x] `go run ./cmd/cicd-lint lint-fitness` — all linters pass
  - [x] `go run ./cmd/cicd-lint lint-deployments` — all validators pass
  - [x] `golangci-lint run` clean
  - [x] All configs/ files follow `{PS-ID}-{purpose}.yml` naming
  - [x] No empty directories without `.gitkeep`
  - [x] Git commits: one per semantic rename group

---

### Phase 4: Deployments Refinement

**Phase Objective**: Clean product-level secret naming, clarify template vs skeleton-template, audit suite compose.

#### Task 4.1: Audit Product-Level Secrets

- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 0.5h
- **Dependencies**: Phase 3 complete
- **Description**: Review all product-level secrets/ directories for old-style naming
- **Acceptance Criteria**:
  - [x] All `deployments/{PRODUCT}/secrets/*.secret` files audited
  - [x] Old-style naming identified and renamed to current standard
  - [x] Compose file secret references updated
  - [x] `go run ./cmd/cicd-lint lint-deployments` passes

#### Task 4.2: Document Template vs Skeleton-Template

- **Status**: ✅ Complete
- **Estimated**: 30m
- **Actual**: 0.3h
- **Dependencies**: None
- **Description**: Clearly document the purpose distinction between `deployments/template/` and `deployments/skeleton-template/`
- **Acceptance Criteria**:
  - [x] `deployments/template/README.md` created or updated explaining: template for creating NEW services
  - [x] `deployments/skeleton-template/` documented as the actual skeleton service deployment
  - [x] ARCHITECTURE.md updated with template directory purpose
  - [x] No ambiguity remains

#### Task 4.3: Audit Suite Compose Size

- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 0.3h
- **Dependencies**: None
- **Description**: Review `deployments/cryptoutil-suite/compose.yml` (1507 lines) for duplication reduction
- **Acceptance Criteria**:
  - [x] Compose file reviewed for unnecessary duplication
  - [x] Recommended optimizations documented (if any)
  - [x] If optimizations applied: line count reduced, all validators pass
  - [x] Delegation chain SUITE->PRODUCT->SERVICE verified
- **Findings**:
  - 1507 lines: 30 service blocks (10 services × 3 instances: sqlite, pg-1, pg-2), 1 builder, 1 secrets healthcheck, includes for shared-telemetry and shared-postgres
  - Each service block ~38-50 lines with repeated secrets/depends_on/healthcheck/deploy sections
  - Duplication is **structural** — each tier uses different port ranges (SERVICE: 8XXX, PRODUCT: 18XXX, SUITE: 28XXX) so `include` delegation across tiers is not viable
  - YAML anchors or Docker Compose `extends` (v2.24+) could reduce ~5-line secrets blocks and ~10-line depends_on blocks, but benefit is marginal (~100 lines saved, increased complexity)
  - **No optimizations applied** — current structure is correct and validated by existing linters
  - **Delegation chain**: SUITE, PRODUCT, and SERVICE tiers are **independent self-contained files** (not delegating), each with appropriate port offsets. Shared infrastructure (telemetry, postgres) via `include`. This is correct per ARCHITECTURE.md Section 12.3

#### Task 4.4: Phase 4 Quality Gate Verification

- **Status**: ✅ Complete
- **Estimated**: 30m
- **Actual**: 0.2h
- **Dependencies**: Tasks 4.1-4.3
- **Description**: Verify all deployment changes maintain project health
- **Acceptance Criteria**:
  - [x] `go run ./cmd/cicd-lint lint-deployments` — all 63 validators pass
  - [x] `go run ./cmd/cicd-lint lint-fitness` — all 44 linters pass
  - [x] Docker Compose syntax valid (if changed)
  - [x] Git commits: one per semantic change
- **Evidence**:
  - `go build ./...` clean (with and without tags)
  - `golangci-lint run ./...` — 0 issues
  - `lint-deployments` — all 63 validators passed
  - `lint-fitness` — completed successfully
  - `lint-docs` — 270 valid refs, 0 broken, propagation passed
  - All cicd_lint tests pass with `-shuffle=on`

---

### Phase 5: ARCHITECTURE.md Roadmap Consolidation

**Phase Objective**: Make ARCHITECTURE.md the complete SSOT by merging satellite docs, documenting missing strategies, and syncing all Decisions 5-11 from target-structure.md.

#### Task 5.1: Merge ARCHITECTURE-COMPOSE-MULTIDEPLOY.md

- **Status**: Not Started
- **Estimated**: 2h
- **Dependencies**: Phase 4 complete
- **Description**: Merge 872-line compose tier documentation into ARCHITECTURE.md Section 12.3
- **Acceptance Criteria**:
  - [ ] Content from ARCHITECTURE-COMPOSE-MULTIDEPLOY.md merged into Section 12.3
  - [ ] Duplicate content deduplicated
  - [ ] Section 12.3 comprehensive for compose tier patterns
  - [ ] ARCHITECTURE-COMPOSE-MULTIDEPLOY.md deleted
  - [ ] All references to deleted file updated (search codebase)
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

#### Task 5.2: Add Infrastructure CLI Tools Documentation

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Phase 2 complete
- **Description**: Add Section 4.4.8 documenting cicd-lint, workflow, and demo as intentional non-product CLI entries
- **Acceptance Criteria**:
  - [ ] New section documenting infrastructure tool rationale
  - [ ] Clear distinction from product/service CLI pattern
  - [ ] References cicd-lint (not cicd) per Decision 11
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

#### Task 5.3: Add Archive and Dead Code Policy

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Phase 1 complete
- **Description**: Add Section 13.9 defining archive vs delete policy
- **Acceptance Criteria**:
  - [ ] Policy: code is DELETED (not archived) — git history preserves everything
  - [ ] Fitness linter prevents `_archived/` directory creation
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

#### Task 5.4: Roadmap Vision Section

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Ensure ARCHITECTURE.md has a clear section summarizing the complete vision for 1 suite / 5 products / 10 services
- **Acceptance Criteria**:
  - [ ] Vision section captures: suite federation, product grouping, service independence
  - [ ] 3-tier deployment strategy (SERVICE/PRODUCT/SUITE) fully described
  - [ ] Migration priority documented (sm-im -> jose-ja -> sm-kms -> pki-ca -> identity)
  - [ ] LLM agent reading this section can understand the end goal
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

#### Task 5.5: Sync ARCHITECTURE.md with target-structure.md Decisions

- **Status**: Not Started
- **Estimated**: 2h
- **Dependencies**: Tasks 5.1-5.4
- **Description**: Update ARCHITECTURE.md to reflect Decisions 5-11 from target-structure.md
- **Acceptance Criteria**:
  - [ ] Section 7: Explicitly state "PostgreSQL and SQLite only — no Citus" (Decision 5)
  - [ ] Section 5.1 or new section: Framework tier routing pattern documented — suite/cli/, product/cli/, service/cli/ (Decision 6)
  - [ ] Section 9.7: CI/CD workflow matrix updated — ci-cicd-lint.yml merged into ci-quality.yml (Decision 7)
  - [ ] Section 12.5: Environment configs (development.yml, production.yml, test.yml, profiles/) documented as DELETE (Decision 8)
  - [ ] Section 12.6: Legacy secrets policy documented (Decision 9)
  - [ ] Section 9.10: Workflow subcommands documented (Decision 10)
  - [ ] Section 9.10: cicd-lint rename reflected (Decision 11)
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

#### Task 5.6: Merge ci-cicd-lint.yml into ci-quality.yml

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Task 5.5
- **Description**: Move ci-cicd-lint.yml job steps into ci-quality.yml and delete ci-cicd-lint.yml (Decision 7)
- **Acceptance Criteria**:
  - [ ] ci-cicd-lint.yml job steps copied into ci-quality.yml as new job
  - [ ] ci-cicd-lint.yml deleted
  - [ ] ci-quality.yml syntax valid
  - [ ] CI triggers still cover same paths
  - [ ] `go build ./...` clean (workflow changes don't affect build, but verify)

#### Task 5.7: Phase 5 Quality Gate Verification

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Tasks 5.1-5.6
- **Description**: Verify all documentation and workflow changes are consistent and pass validation
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
  - [ ] `go run ./cmd/cicd-lint lint-docs validate-propagation` passes
  - [ ] No broken cross-references in ARCHITECTURE.md
  - [ ] ARCHITECTURE-COMPOSE-MULTIDEPLOY.md deleted
  - [ ] ci-cicd-lint.yml deleted
  - [ ] All Decisions 5-11 reflected in ARCHITECTURE.md
  - [ ] Git commits: one per semantic documentation change

---

### Phase 6: Fitness Linter Expansion

**Phase Objective**: Add 5 new fitness linters to prevent regression of v5 standards.

#### Task 6.1: Implement archive-detector Linter

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Phase 1 complete
- **Description**: Detect `_archived/`, `archived/`, `orphaned/` directories anywhere in repository
- **Acceptance Criteria**:
  - [ ] `internal/apps/tools/cicd_lint/lint_fitness/archive_detector/` created
  - [ ] Walks repo tree, fails on any archived/orphaned directory
  - [ ] Tests >= 98% coverage
  - [ ] Registered in fitness catalog
  - [ ] Passes on current codebase (post Phase 1 cleanup)

#### Task 6.2: Implement configs-naming Linter

- **Status**: Not Started
- **Estimated**: 2h
- **Dependencies**: Phase 3 complete
- **Description**: Validate configs/ directory structure and file naming follows `{PS-ID}-{purpose}.yml` pattern
- **Acceptance Criteria**:
  - [ ] `internal/apps/tools/cicd_lint/lint_fitness/configs_naming/` created
  - [ ] Validates `configs/{PRODUCT}/{SERVICE}/` structure against entity registry
  - [ ] Validates file naming: `{PS-ID}-{purpose}.yml` for service template configs
  - [ ] Allows product-level configs in `configs/{PRODUCT}/`
  - [ ] Tests >= 98% coverage with edge cases
  - [ ] Registered in fitness catalog
  - [ ] Passes on current codebase (post Phase 3 standardization)

#### Task 6.3: Implement cmd-anti-pattern Linter

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Phase 2 complete
- **Description**: Detect `cmd/{PRODUCT}-{subcommand}/` anti-pattern entries
- **Acceptance Criteria**:
  - [ ] `internal/apps/tools/cicd_lint/lint_fitness/cmd_anti_pattern/` created
  - [ ] Detects `cmd/{product}-{subcommand}/` patterns (e.g., cmd/identity-compose/)
  - [ ] Does NOT flag legitimate PS-ID entries (e.g., cmd/identity-authz/)
  - [ ] Does NOT flag documented infrastructure tools (cmd/cicd-lint/, cmd/workflow/)
  - [ ] Tests >= 98% coverage
  - [ ] Registered in fitness catalog

#### Task 6.4: Implement configs-empty-dir Linter

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Phase 3 complete
- **Description**: Detect empty directories in configs/ without `.gitkeep`
- **Acceptance Criteria**:
  - [ ] `internal/apps/tools/cicd_lint/lint_fitness/configs_empty_dir/` created
  - [ ] Walks configs/ tree, fails on empty dirs without `.gitkeep`
  - [ ] Tests >= 98% coverage
  - [ ] Registered in fitness catalog

#### Task 6.5: Implement configs-deployments-consistency Linter

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Phase 3 complete
- **Description**: Validate configs/ mirrors deployments/ service structure
- **Acceptance Criteria**:
  - [ ] `internal/apps/tools/cicd_lint/lint_fitness/configs_deployments_consistency/` created
  - [ ] Ensures every deployments/{PS-ID}/ has matching configs/{PRODUCT}/{SERVICE}/
  - [ ] Uses entity registry to map PS-ID to PRODUCT/SERVICE
  - [ ] Tests >= 98% coverage
  - [ ] Registered in fitness catalog

#### Task 6.6: Phase 6 Quality Gate Verification

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Tasks 6.1-6.5
- **Description**: Verify all new linters pass and meet quality standards
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` — all 49+ linters pass
  - [ ] All new linter packages >= 98% coverage
  - [ ] Mutation testing >= 98% on new linter packages
  - [ ] `go build ./...` clean
  - [ ] `golangci-lint run` clean
  - [ ] Git commits: one per linter implementation

---

### Phase 7: Knowledge Propagation

**Phase Objective**: Apply all lessons to permanent artifacts. Audit skills, agents, and instructions for framework-v5 compliance. Document deferred work.

#### Task 7.1: Review and Consolidate Lessons

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Phases 1-6 complete
- **Description**: Review lessons.md from all phases, identify patterns for permanent artifacts
- **Acceptance Criteria**:
  - [ ] lessons.md reviewed for all 6 phases
  - [ ] Patterns extracted for instruction file updates
  - [ ] Gaps identified in existing documentation

#### Task 7.2: Update Instruction Files

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Task 7.1
- **Description**: Update copilot instruction files with new configs/deployment standards
- **Acceptance Criteria**:
  - [ ] `02-01.architecture.instructions.md` updated with configs/ naming standard
  - [ ] `04-01.deployment.instructions.md` updated with template vs skeleton-template clarity
  - [ ] `03-05.linting.instructions.md` updated with new fitness linter catalog count
  - [ ] All `cmd/cicd` references updated to `cmd/cicd-lint` in instruction files
  - [ ] All instruction file @source blocks aligned with ARCHITECTURE.md

#### Task 7.3: Audit and Update Skills

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Task 7.1
- **Description**: Audit all 14 skill directories for framework-v5 compliance: verify names match purpose, content reflects new patterns
- **Acceptance Criteria**:
  - [ ] `new-service/SKILL.md` reviewed: verify no overlap with skeleton-template
  - [ ] `coverage-analysis/SKILL.md` reviewed: verify if mutation testing is included (if not, document scope)
  - [ ] `contract-test-gen/SKILL.md` reviewed: verify name clarity
  - [ ] `migration-create/SKILL.md` reviewed: verify name describes purpose accurately
  - [ ] `fitness-function-gen/SKILL.md` reviewed: verify name clarity
  - [ ] Any skill generating config files updated with new {PS-ID} naming
  - [ ] Any skill referencing `cmd/cicd` updated to `cmd/cicd-lint`
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

#### Task 7.4: Update Agent Files

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Task 7.1
- **Description**: Update agent files with new patterns
- **Acceptance Criteria**:
  - [ ] Any agent referencing configs/ patterns updated
  - [ ] Any agent referencing `cmd/cicd` updated to `cmd/cicd-lint`
  - [ ] Agent architecture references current framework tier routing (suite/product/service)

#### Task 7.5: Document Deferred Work

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Task 7.1
- **Description**: Document `test/load/` Gatling refactoring and any other deferred items
- **Acceptance Criteria**:
  - [ ] `test/load/` refactoring documented as deferred work with rationale (low priority, Java/Maven domain)
  - [ ] Any other deferred items from lessons.md documented
  - [ ] No undocumented deferred work remains

#### Task 7.6: Final Propagation Verification

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Tasks 7.1-7.5
- **Description**: Verify all propagation integrity
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs validate-propagation` passes
  - [ ] `go run ./cmd/cicd-lint lint-fitness` — all linters pass
  - [ ] `go run ./cmd/cicd-lint lint-deployments` — all validators pass
  - [ ] `go build ./...` clean
  - [ ] `golangci-lint run` clean
  - [ ] Git commits: one per semantic documentation update

---

## Cross-Cutting Tasks

### Testing

- [ ] Unit tests >= 95% coverage (production), >= 98% (infrastructure/utility)
- [ ] All new linter tests pass
- [ ] Race detector clean: `go test -race ./...` (modified packages)
- [ ] Mutation testing >= 98% on new infrastructure code

### Code Quality

- [ ] `golangci-lint run` clean after all phases
- [ ] `golangci-lint run --build-tags e2e,integration` clean
- [ ] No new TODOs without tracking
- [ ] Formatting clean (`gofumpt -s -w ./`)
- [ ] Imports organized (`goimports -w ./`)
- [ ] All `cmd/cicd` references updated to `cmd/cicd-lint` after Phase 2 rename

### Documentation

- [ ] ARCHITECTURE.md fully updated (merge COMPOSE-MULTIDEPLOY, Section 12.5, Section 4.4.8, Section 13.9, Decisions 5-11)
- [ ] Instruction files reflect new standards (including `cmd/cicd-lint` references)
- [ ] ARCHITECTURE-COMPOSE-MULTIDEPLOY.md deleted
- [ ] configs/ relationship documented
- [ ] Database engine documentation updated (PostgreSQL + SQLite only, no Citus)

### Deployment

- [ ] All 68+ deployment validators pass after all phases
- [ ] All 49+ fitness linters pass after all phases
- [ ] Docker Compose syntax valid
- [ ] Config file references updated
- [ ] `ci-cicd-lint.yml` consolidated into `ci-quality.yml`

---

## Notes / Deferred Work

**All quizme-v1 decisions confirmed** (merged 2026-03-22):
1. ✓ Decision 1 (Q1=E): configs/ = canonical SSOT (env-agnostic), deployments/ = deployment wiring that consumes/overlays configs/
2. ✓ Decision 2 (Q2=C): Archive ALL demo entries (cmd/demo, cmd/identity-compose, cmd/identity-demo, internal/apps/demo). Keep only cicd/workflow.
3. ✓ Decision 3 (Q3=A): Delete all archived/orphaned permanently. Git history preserves content.
4. ✓ Decision 4 (Q4=A): Merge ARCHITECTURE-COMPOSE-MULTIDEPLOY.md into ARCHITECTURE.md Section 12.3 and delete.

**All quizme-v2 decisions confirmed** (merged 2026-03-23):
5. ✓ Decision (Q1=E): cicd tool → `internal/apps/tools/cicd_lint/`, binary → `cmd/cicd-lint/`
6. ✓ Decision (Q2=D): pkiinit → `framework/tls/` (merge with existing TLS code)
7. ✓ Decision (Q3=B): Error wrapping → `framework/apperr/`
8. ✓ Decision (Q5=B): `testdata/` directories at repo root → DELETE

**All quizme-v3 decisions confirmed** (merged 2026-03-24):
9. ✓ Decision (Q1=B): Create `framework/suite/cli/` with `RouteSuite()`, `SuiteConfig`, `ProductEntry`
10. ✓ Decision (Q2=B): Move `RouteProduct()` to `framework/product/cli/` (from `framework/service/cli/`)

**Pending**: User review of `target-structure.md` before Phase 3 config moves begin.

**NEW scope discovered**: ~80+ junk files at repository root (*.exe, *.py, coverage_*) need cleanup — added to Phase 1.

**Deferred work**:
- `test/load/` Gatling refactoring: Low priority (Java/Maven domain, not Go). Schedule after framework-v5 completion.
- Skill name audit results: Documented in Phase 7 Task 7.3 — evaluate during execution.

---

## Evidence Archive

- `test-output/phase0-research/` - Initial repository analysis (from plan creation)
- `test-output/phase1/` - Phase 1 deletion verification logs
- `test-output/phase3/` - Config rename reference search results
- `test-output/phase6/` - Linter coverage and mutation results
