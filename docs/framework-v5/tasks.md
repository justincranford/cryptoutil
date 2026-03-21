# Tasks - Framework v5: Rigid Standardization & Cleanup

**Status**: 0 of 42 tasks complete (0%)
**Last Updated**: 2026-03-22
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

### Phase 1: Archive and Dead Code Cleanup

**Phase Objective**: Remove 161+ dead files across 9 archived/orphaned directories.

#### Task 1.1: Delete Identity Archived Directories

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Delete all 6 identity archived directories containing 132 files total
- **Acceptance Criteria**:
  - [ ] `internal/apps/identity/_archived/` deleted (92 files)
  - [ ] `internal/apps/identity/_authz-archived/` deleted (8 files)
  - [ ] `internal/apps/identity/_idp-archived/` deleted (8 files)
  - [ ] `internal/apps/identity/_rp-archived/` deleted (8 files)
  - [ ] `internal/apps/identity/_rs-archived/` deleted (8 files)
  - [ ] `internal/apps/identity/_spa-archived/` deleted (8 files)
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean

#### Task 1.2: Delete PKI and Test Archived Directories

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: None
- **Description**: Delete PKI empty archive and test archived directories
- **Acceptance Criteria**:
  - [ ] `internal/apps/pki/_ca-archived/` deleted (empty dir)
  - [ ] `internal/test/archived/` deleted (15 files)
  - [ ] `go build ./...` clean

#### Task 1.3: Delete Deployment Archived Directory

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: None
- **Description**: Delete legacy compose archived directory
- **Acceptance Criteria**:
  - [ ] `deployments/archived/` deleted (14 files)
  - [ ] `go run ./cmd/cicd lint-deployments` passes (all 68+ validators)
  - [ ] `go run ./cmd/cicd lint-fitness` passes (all 44 linters)

#### Task 1.4: Delete Configs Orphaned Directory

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: None
- **Description**: Delete `configs/orphaned/` legacy configs (observability, template, test secrets)
- **Acceptance Criteria**:
  - [ ] `configs/orphaned/` deleted
  - [ ] No code references to orphaned config files remain
  - [ ] `go build ./...` clean

#### Task 1.5: Phase 1 Quality Gate Verification

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Tasks 1.1-1.4
- **Description**: Comprehensive verification that all deletions left the project healthy
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `go run ./cmd/cicd lint-fitness` — all 44 linters pass
  - [ ] `go run ./cmd/cicd lint-deployments` — all 68 validators pass
  - [ ] `golangci-lint run` clean
  - [ ] Zero `_archived`, `archived/`, or `orphaned/` directories remain
  - [ ] Git commit: `refactor: delete 161+ archived and orphaned files`

---

### Phase 2: Non-Standard Entry Rationalization

**Phase Objective**: Classify all non-standard cmd/ and internal/apps/ entries. Fix anti-pattern violations.

#### Task 2.1: Document Infrastructure CLI Tools

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Document `cmd/cicd/` and `cmd/workflow/` as intentional infrastructure tools in ARCHITECTURE.md
- **Acceptance Criteria**:
  - [ ] ARCHITECTURE.md Section 4.4.7 updated with infrastructure tool documentation
  - [ ] `cmd/cicd/` documented as CICD tooling (linters, formatters, scripts)
  - [ ] `cmd/workflow/` documented as workflow testing infrastructure
  - [ ] `go run ./cmd/cicd lint-docs` passes

#### Task 2.2: Evaluate cmd/identity-compose

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Assess whether identity-compose should be subcommand of cmd/identity or archived
- **Acceptance Criteria**:
  - [ ] Code read and purpose understood
  - [ ] Decision documented: merge into cmd/identity subcommand, OR archive (with rationale)
  - [ ] If merge: cmd/identity compose subcommand implemented
  - [ ] If archive: cmd/identity-compose/ removed
  - [ ] All workflow references updated
  - [ ] Build clean

#### Task 2.3: Evaluate cmd/identity-demo

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Assess whether identity-demo should be subcommand of cmd/identity or archived
- **Acceptance Criteria**:
  - [ ] Code read and purpose understood
  - [ ] Decision documented: merge into cmd/identity subcommand, OR archive (with rationale)
  - [ ] If merge: cmd/identity demo subcommand implemented
  - [ ] If archive: cmd/identity-demo/ removed
  - [ ] Build clean

#### Task 2.4: Evaluate cmd/demo and internal/apps/demo

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Assess whether unified demo CLI should become a suite-level subcommand or remain standalone
- **Acceptance Criteria**:
  - [ ] 17-file demo codebase reviewed for usefulness
  - [ ] Decision documented: keep as standalone documented tool, merge into cmd/cryptoutil demo, OR archive
  - [ ] Changes implemented per decision
  - [ ] Build clean

#### Task 2.5: Evaluate internal/apps/pkiinit

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: None
- **Description**: Assess whether PKI init tool belongs under cmd/pki init or is orphaned
- **Acceptance Criteria**:
  - [ ] 3-file codebase reviewed
  - [ ] Decision documented: integrate into cmd/pki init, OR archive
  - [ ] Changes implemented per decision
  - [ ] Build clean

#### Task 2.6: Clean docs/demo-brainstorm

- **Status**: Not Started
- **Estimated**: 15m
- **Dependencies**: Task 2.4
- **Description**: Archive or delete demo brainstorm documents if no longer relevant
- **Acceptance Criteria**:
  - [ ] 3 files reviewed for relevance
  - [ ] Deleted if outdated, kept if active planning reference
  - [ ] No references to deleted files remain

#### Task 2.7: Phase 2 Quality Gate Verification

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Tasks 2.1-2.6
- **Description**: Verify all rationalization changes maintain project health
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `go test ./...` passes (modified packages)
  - [ ] `go run ./cmd/cicd lint-fitness` — all linters pass
  - [ ] `golangci-lint run` clean
  - [ ] Zero cmd/ anti-pattern violations remain
  - [ ] ARCHITECTURE.md documents all intentional infrastructure tools
  - [ ] Git commits: one per semantic change

---

### Phase 3: Configs Standardization

**Phase Objective**: Apply rigid {PS-ID}-based naming to all configs/ files, resolve path mismatches, document the dual configs/ vs deployments/config/ relationship.

#### Task 3.1: Design Configs Canonical Structure

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Document the target configs/ structure aligned with entity registry
- **Acceptance Criteria**:
  - [ ] Target directory tree documented (configs/{PRODUCT}/{SERVICE}/ for all 10 services)
  - [ ] File naming convention defined: `{PS-ID}-{purpose}.yml` for service template configs
  - [ ] Domain config convention defined: `{PS-ID}-{domain-purpose}.yml`
  - [ ] Product-level config convention defined
  - [ ] Relationship between configs/ and deployments/config/ documented

#### Task 3.2: Fix configs/ca/ Path Mismatch

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Task 3.1
- **Description**: Move `configs/ca/` to `configs/pki/ca/` (entity registry product is `pki`, not `ca`)
- **Acceptance Criteria**:
  - [ ] `configs/ca/ca-server.yml` moved to `configs/pki/ca/pki-ca-server.yml`
  - [ ] `configs/ca/ca-config-schema.yaml` moved to `configs/pki/ca/pki-ca-config-schema.yaml`
  - [ ] `configs/ca/profiles/` moved to `configs/pki/ca/profiles/`
  - [ ] Old `configs/ca/` directory deleted
  - [ ] All Go code references updated (search: `configs/ca/`)
  - [ ] All compose file references updated
  - [ ] lint-deployments mirror mapping updated
  - [ ] Build clean

#### Task 3.3: Standardize SM Configs Naming

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Task 3.1
- **Description**: Rename SM service config files to use PS-ID prefix
- **Acceptance Criteria**:
  - [ ] `configs/sm/kms/config-pg-1.yml` -> `configs/sm/kms/sm-kms-pg-1.yml`
  - [ ] `configs/sm/kms/config-pg-2.yml` -> `configs/sm/kms/sm-kms-pg-2.yml`
  - [ ] `configs/sm/kms/config-sqlite.yml` -> `configs/sm/kms/sm-kms-sqlite.yml`
  - [ ] `configs/sm/im/config-pg-1.yml` -> `configs/sm/im/sm-im-pg-1.yml`
  - [ ] `configs/sm/im/config-pg-2.yml` -> `configs/sm/im/sm-im-pg-2.yml`
  - [ ] `configs/sm/im/config-sqlite.yml` -> `configs/sm/im/sm-im-sqlite.yml`
  - [ ] All Go code references updated
  - [ ] Build clean

#### Task 3.4: Standardize Identity Configs Naming

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Task 3.1
- **Description**: Evaluate identity `-docker.yml` files and standardize naming
- **Acceptance Criteria**:
  - [ ] `authz.yml` -> `identity-authz.yml` (or keep if already domain config)
  - [ ] `authz-docker.yml` evaluated for overlap with deployments/identity-authz/config/
  - [ ] Decision: remove `-docker.yml` duplicates OR rename to `{PS-ID}-docker.yml`
  - [ ] Same pattern applied to idp, rp, rs, spa
  - [ ] All Go code references updated
  - [ ] Build clean

#### Task 3.5: Standardize Jose and Skeleton Configs

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Task 3.1
- **Description**: Rename product-level configs to use PS-ID; populate empty directories
- **Acceptance Criteria**:
  - [ ] `configs/jose/jose-server.yml` -> evaluate: product-level or move to `configs/jose/ja/jose-ja-server.yml`
  - [ ] `configs/skeleton/skeleton-server.yml` -> evaluate: product-level or move to `configs/skeleton/template/`
  - [ ] `configs/skeleton/template/template-server.yml` renamed to `skeleton-template-server.yml`
  - [ ] `configs/jose/ja/` populated (was empty)
  - [ ] Build clean

#### Task 3.6: Update ARCHITECTURE.md Section 12.5

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Tasks 3.2-3.5
- **Description**: Rewrite Section 12.5 to document the TARGET configs/ standard (not current inconsistent state)
- **Acceptance Criteria**:
  - [ ] Section 12.5 rewritten with canonical configs/ structure
  - [ ] Naming conventions table included
  - [ ] Dual configs/ vs deployments/config/ relationship documented
  - [ ] Examples for each config file type
  - [ ] `go run ./cmd/cicd lint-docs` passes

#### Task 3.7: Update lint-deployments Mirror Mapping

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Tasks 3.2-3.5
- **Description**: Update lint-deployments to validate configs/ mirrors deployments/ structure
- **Acceptance Criteria**:
  - [ ] Mirror mapping updated for `configs/ca/` -> `configs/pki/ca/` rename
  - [ ] All 68+ deployment validators still pass
  - [ ] Tests updated and passing (>=98% coverage)
  - [ ] Build clean

#### Task 3.8: Phase 3 Quality Gate Verification

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Tasks 3.1-3.7
- **Description**: Verify all configs changes maintain project health
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `go test ./...` passes (modified packages)
  - [ ] `go run ./cmd/cicd lint-fitness` — all linters pass
  - [ ] `go run ./cmd/cicd lint-deployments` — all validators pass
  - [ ] `golangci-lint run` clean
  - [ ] All configs/ files follow `{PS-ID}-{purpose}.yml` naming
  - [ ] No empty directories without `.gitkeep`
  - [ ] Git commits: one per semantic rename group

---

### Phase 4: Deployments Refinement

**Phase Objective**: Clean product-level secret naming, clarify template vs skeleton-template, audit suite compose.

#### Task 4.1: Audit Product-Level Secrets

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Phase 3 complete
- **Description**: Review all product-level secrets/ directories for old-style naming
- **Acceptance Criteria**:
  - [ ] All `deployments/{PRODUCT}/secrets/*.secret` files audited
  - [ ] Old-style naming identified and renamed to current standard
  - [ ] Compose file secret references updated
  - [ ] `go run ./cmd/cicd lint-deployments` passes

#### Task 4.2: Document Template vs Skeleton-Template

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: None
- **Description**: Clearly document the purpose distinction between `deployments/template/` and `deployments/skeleton-template/`
- **Acceptance Criteria**:
  - [ ] `deployments/template/README.md` created or updated explaining: template for creating NEW services
  - [ ] `deployments/skeleton-template/` documented as the actual skeleton service deployment
  - [ ] ARCHITECTURE.md updated with template directory purpose
  - [ ] No ambiguity remains

#### Task 4.3: Audit Suite Compose Size

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Review `deployments/cryptoutil-suite/compose.yml` (1507 lines) for duplication reduction
- **Acceptance Criteria**:
  - [ ] Compose file reviewed for unnecessary duplication
  - [ ] Recommended optimizations documented (if any)
  - [ ] If optimizations applied: line count reduced, all validators pass
  - [ ] Delegation chain SUITE->PRODUCT->SERVICE verified

#### Task 4.4: Phase 4 Quality Gate Verification

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Tasks 4.1-4.3
- **Description**: Verify all deployment changes maintain project health
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd lint-deployments` — all 68+ validators pass
  - [ ] `go run ./cmd/cicd lint-fitness` — all 44 linters pass
  - [ ] Docker Compose syntax valid (if changed)
  - [ ] Git commits: one per semantic change

---

### Phase 5: ARCHITECTURE.md Roadmap Consolidation

**Phase Objective**: Make ARCHITECTURE.md the complete SSOT by merging satellite docs and documenting missing strategies.

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
  - [ ] `go run ./cmd/cicd lint-docs` passes

#### Task 5.2: Add Infrastructure CLI Tools Documentation

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Phase 2 complete
- **Description**: Add Section 4.4.8 documenting cicd, workflow, and demo as intentional non-product CLI entries
- **Acceptance Criteria**:
  - [ ] New section documenting infrastructure tool rationale
  - [ ] Clear distinction from product/service CLI pattern
  - [ ] `go run ./cmd/cicd lint-docs` passes

#### Task 5.3: Add Archive and Dead Code Policy

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Phase 1 complete
- **Description**: Add Section 13.9 defining archive vs delete policy
- **Acceptance Criteria**:
  - [ ] Policy: code is DELETED (not archived) — git history preserves everything
  - [ ] Fitness linter prevents `_archived/` directory creation
  - [ ] `go run ./cmd/cicd lint-docs` passes

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
  - [ ] `go run ./cmd/cicd lint-docs` passes

#### Task 5.5: Phase 5 Quality Gate Verification

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Tasks 5.1-5.4
- **Description**: Verify all documentation changes are consistent and pass validation
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd lint-docs` passes
  - [ ] `go run ./cmd/cicd lint-docs validate-propagation` passes
  - [ ] No broken cross-references in ARCHITECTURE.md
  - [ ] ARCHITECTURE-COMPOSE-MULTIDEPLOY.md deleted
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
  - [ ] `internal/apps/cicd/lint_fitness/archive_detector/` created
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
  - [ ] `internal/apps/cicd/lint_fitness/configs_naming/` created
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
  - [ ] `internal/apps/cicd/lint_fitness/cmd_anti_pattern/` created
  - [ ] Detects `cmd/{product}-{subcommand}/` patterns (e.g., cmd/identity-compose/)
  - [ ] Does NOT flag legitimate PS-ID entries (e.g., cmd/identity-authz/)
  - [ ] Does NOT flag documented infrastructure tools (cmd/cicd/, cmd/workflow/)
  - [ ] Tests >= 98% coverage
  - [ ] Registered in fitness catalog

#### Task 6.4: Implement configs-empty-dir Linter

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Phase 3 complete
- **Description**: Detect empty directories in configs/ without `.gitkeep`
- **Acceptance Criteria**:
  - [ ] `internal/apps/cicd/lint_fitness/configs_empty_dir/` created
  - [ ] Walks configs/ tree, fails on empty dirs without `.gitkeep`
  - [ ] Tests >= 98% coverage
  - [ ] Registered in fitness catalog

#### Task 6.5: Implement configs-deployments-consistency Linter

- **Status**: Not Started
- **Estimated**: 1h
- **Dependencies**: Phase 3 complete
- **Description**: Validate configs/ mirrors deployments/ service structure
- **Acceptance Criteria**:
  - [ ] `internal/apps/cicd/lint_fitness/configs_deployments_consistency/` created
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
  - [ ] `go run ./cmd/cicd lint-fitness` — all 49+ linters pass
  - [ ] All new linter packages >= 98% coverage
  - [ ] Mutation testing >= 98% on new linter packages
  - [ ] `go build ./...` clean
  - [ ] `golangci-lint run` clean
  - [ ] Git commits: one per linter implementation

---

### Phase 7: Knowledge Propagation

**Phase Objective**: Apply all lessons to permanent artifacts.

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
  - [ ] All instruction file @source blocks aligned with ARCHITECTURE.md

#### Task 7.3: Update Agent and Skill Files

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Task 7.1
- **Description**: Update agent/skill files with new patterns if applicable
- **Acceptance Criteria**:
  - [ ] Any agent referencing configs/ patterns updated
  - [ ] Any skill generating config files updated with new naming
  - [ ] `go run ./cmd/cicd lint-docs` passes

#### Task 7.4: Final Propagation Verification

- **Status**: Not Started
- **Estimated**: 30m
- **Dependencies**: Tasks 7.1-7.3
- **Description**: Verify all propagation integrity
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd lint-docs validate-propagation` passes
  - [ ] `go run ./cmd/cicd lint-fitness` — all linters pass
  - [ ] `go run ./cmd/cicd lint-deployments` — all validators pass
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

### Documentation

- [ ] ARCHITECTURE.md fully updated (merge COMPOSE-MULTIDEPLOY, Section 12.5, Section 4.4.8, Section 13.9)
- [ ] Instruction files reflect new standards
- [ ] ARCHITECTURE-COMPOSE-MULTIDEPLOY.md deleted
- [ ] configs/ relationship documented

### Deployment

- [ ] All 68+ deployment validators pass after all phases
- [ ] All 49+ fitness linters pass after all phases
- [ ] Docker Compose syntax valid
- [ ] Config file references updated

---

## Notes / Deferred Work

**All quizme-v1 decisions confirmed** (merged 2026-03-22):
1. ✓ Decision 1 (Q1=E): configs/ = canonical SSOT (env-agnostic), deployments/ = deployment wiring that consumes/overlays configs/
2. ✓ Decision 2 (Q2=C): Archive ALL demo entries (cmd/demo, cmd/identity-compose, cmd/identity-demo, internal/apps/demo). Keep only cicd/workflow.
3. ✓ Decision 3 (Q3=A): Delete all archived/orphaned permanently. Git history preserves content.
4. ✓ Decision 4 (Q4=A): Merge ARCHITECTURE-COMPOSE-MULTIDEPLOY.md into ARCHITECTURE.md Section 12.3 and delete.

**Pending**: User review of `target-structure.md` before Phase 3 config moves begin.

**NEW scope discovered**: ~80+ junk files at repository root (*.exe, *.py, coverage_*) need cleanup — added to Phase 1.

---

## Evidence Archive

- `test-output/phase0-research/` - Initial repository analysis (from plan creation)
- `test-output/phase1/` - Phase 1 deletion verification logs
- `test-output/phase3/` - Config rename reference search results
- `test-output/phase6/` - Linter coverage and mutation results
