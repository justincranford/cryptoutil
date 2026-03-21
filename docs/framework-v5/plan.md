# Implementation Plan - Framework v5: Rigid Standardization & Cleanup

**Status**: Planning
**Created**: 2026-03-21
**Last Updated**: 2026-03-21
**Purpose**: Apply rigid standardization and enforcement across configs/, deployments/, cmd/, internal/apps/, and docs/ for 1 suite, 5 products, and 10 services. Clean dead code, rationalize non-standard entries, and ensure ARCHITECTURE.md is the complete roadmap SSOT.

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- Correctness: ALL documentation must be accurate and complete
- Completeness: NO steps skipped, NO steps de-prioritized, NO shortcuts
- Thoroughness: Evidence-based validation at every step
- Reliability: Quality gates enforced (>=95%/98% coverage/mutation)
- Efficiency: Optimized for maintainability and performance, NOT implementation speed
- Accuracy: Changes must address root cause, not just symptoms
- Time Pressure: NEVER rush, NEVER skip validation, NEVER defer quality checks
- Premature Completion: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions.**

## Overview

Framework v5 addresses six areas of structural inconsistency discovered through deep repository analysis on 2026-03-21:

1. **Dead code accumulation**: 161+ files across 9 archived directories and empty stubs
2. **Non-standard cmd/internal entries**: 5 cmd/ and 4 internal/apps/ entries that violate the PRODUCT/SERVICE hierarchy
3. **configs/ inconsistency**: The largest standardization gap — naming varies per product, duality with deployments/*/config/, empty directories, mismatched paths
4. **deployments/ refinement**: Product-level secret naming drift, archived legacy compose files, template ambiguity
5. **ARCHITECTURE.md gaps**: 872-line ARCHITECTURE-COMPOSE-MULTIDEPLOY.md exists outside SSOT, configs strategy under-documented, demo/workflow strategy absent
6. **Missing fitness linters**: No automated enforcement for configs/ naming, archive detection, or demo artifact standards

## Background

Framework v4 (COMPLETE) delivered 44 fitness linters and established the entity registry as canonical SSOT. It addressed naming drift, structural validation, and compose service name enforcement. Framework v5 carries forward the areas v4 did not cover: standalone configs/ standardization, dead code cleanup, non-standard entry rationalization, and ARCHITECTURE.md completeness.

**Key v4 outcomes**: All 44 fitness linters pass. 68 deployment validators pass. Entity registry covers all 10 product-services. Build is clean.

## Technical Context

- **Language**: Go 1.26.1
- **Framework**: Service template (`internal/apps/framework/service/`)
- **Database**: PostgreSQL OR SQLite with GORM (CGO-free modernc.org/sqlite)
- **Dependencies**: Entity registry, magic constants, lint_deployments, lint_fitness
- **Related Docs**: `docs/ARCHITECTURE.md` (5080 lines), `docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md` (872 lines)
- **Hierarchy**: 1 suite (cryptoutil), 5 products (identity, jose, pki, skeleton, sm), 10 services

## Inventory - Current State (Evidence from 2026-03-21 Analysis)

### Archived Code (161+ files, 9 directories)

| Directory | Files | Content |
|-----------|-------|---------|
| `internal/apps/identity/_archived/` | 92 | Bootstrap, cmd, compose, demo, e2e, healthcheck, integration, jwks, server, storage, test |
| `internal/apps/identity/_authz-archived/` | 8 | Archived identity-authz stubs |
| `internal/apps/identity/_idp-archived/` | 8 | Archived identity-idp stubs |
| `internal/apps/identity/_rp-archived/` | 8 | Archived identity-rp stubs |
| `internal/apps/identity/_rs-archived/` | 8 | Archived identity-rs stubs |
| `internal/apps/identity/_spa-archived/` | 8 | Archived identity-spa stubs |
| `internal/apps/pki/_ca-archived/` | 0 | Empty directory (orphan) |
| `internal/test/archived/` | 15 | Legacy test infrastructure |
| `deployments/archived/` | 14 | Legacy compose patterns (compose-legacy, cryptoutil-legacy) |

### Non-Standard Entries

**cmd/ (5 non-standard)**:

| Entry | Purpose | Anti-Pattern? |
|-------|---------|--------------|
| `cmd/cicd/` | CICD tooling CLI | Infrastructure tool, not a product/service |
| `cmd/demo/` | Unified demo CLI | Demo orchestration, not a product/service |
| `cmd/identity-compose/` | Identity compose orchestration | Violates "NO executables for subcommands" rule |
| `cmd/identity-demo/` | Identity demo | Violates "NO executables for subcommands" rule |
| `cmd/workflow/` | GitHub Actions workflow testing | Infrastructure tool, not a product/service |

**internal/apps/ (4 non-standard)**:

| Entry | Files | Purpose |
|-------|-------|---------|
| `internal/apps/cicd/` | Large | CICD lint/format/script infrastructure |
| `internal/apps/demo/` | 17 | Unified demo (CA, JOSE, KMS, identity stubs) |
| `internal/apps/pkiinit/` | 3 | PKI initialization tool |
| `internal/apps/workflow/` | 7 | Workflow testing infrastructure |

### configs/ Inconsistency Map

**Two config systems coexist**:
1. `configs/` - Standalone configs, inconsistent naming per product
2. `deployments/*/config/` - Docker-specific configs, standardized `{PS-ID}-app-{variant}.yml`

**Naming inconsistencies**:

| Path | Pattern | Expected Pattern |
|------|---------|-----------------|
| `configs/sm/kms/config-pg-1.yml` | `config-{variant}` (no PS-ID) | `sm-kms-{variant}.yml` |
| `configs/sm/im/config-pg-1.yml` | `config-{variant}` (no PS-ID) | `sm-im-{variant}.yml` |
| `configs/sm/im/im.yml` | `{service}.yml` (domain-specific) | Keep (domain config) |
| `configs/identity/authz/authz.yml` | `{service}.yml` | Keep (domain config) |
| `configs/identity/authz/authz-docker.yml` | `{service}-docker.yml` | Evaluate overlap with deployments/config/ |
| `configs/ca/ca-server.yml` | Under `ca/` not `pki/ca/` | Move to `configs/pki/ca/` |
| `configs/jose/jose-server.yml` | Product-level, not service-level | Move to `configs/jose/ja/jose-ja-server.yml` or keep at product |
| `configs/skeleton/skeleton-server.yml` | Product-level | Evaluate |
| `configs/pki/ca/` | EMPTY directory | Populate or remove |
| `configs/jose/ja/` | EMPTY directory | Populate or remove |

**deployments/*/config/ (standardized pattern)**:
All 10 services + template have 4 config files each (40 total):
- `{PS-ID}-app-common.yml` (shared settings)
- `{PS-ID}-app-sqlite-1.yml` (SQLite variant)
- `{PS-ID}-app-postgresql-1.yml` (PostgreSQL instance 1)
- `{PS-ID}-app-postgresql-2.yml` (PostgreSQL instance 2)

### Deployment Compose Sizes

| Tier | Service | Lines | Notes |
|------|---------|-------|-------|
| SERVICE | sm-kms, sm-im | 287 | Fully mature, 3 instances + infra |
| SERVICE | identity-authz, jose-ja, skeleton-template, pki-ca | 266-269 | Mature |
| SERVICE | identity-idp/rp/rs/spa | 260 | Slightly smaller (fewer features) |
| PRODUCT | identity | 819 | 5 services, largest product |
| PRODUCT | jose, skeleton, sm, pki | 262-283 | 1-2 services each |
| SUITE | cryptoutil-suite | 1507 | All 10 services |
| INFRA | shared-telemetry | 112 | OTel + Grafana |
| INFRA | shared-postgres | 205 | PostgreSQL shared |
| INFRA | shared-citus | 188 | Citus shared |
| TEMPLATE | template | 236 | Template for new services |

### Demo Artifacts

| Location | Files | Purpose |
|----------|-------|---------|
| `cmd/demo/main.go` -> `internal/apps/demo/` | 17 | Unified demo CLI (CA, JOSE, KMS stubs) |
| `cmd/identity-demo/main.go` -> `internal/apps/identity/demo/` | 1+? | Identity demo |
| `cmd/identity-compose/main.go` -> `internal/apps/identity/compose/` | 1+? | Identity compose orchestration |
| `internal/apps/pkiinit/` | 3 | PKI init tool |
| `docs/demo-brainstorm/` | 3 | Demo planning docs |

## Phases

### Phase 1: Archive and Dead Code Cleanup (4h) [Status: TODO]

**Objective**: Remove all archived/dead code directories to reduce repository noise and eliminate 161+ unused files.

**Strategy**: Git-delete all `_archived` and `archived/` directories. These contain superseded code from pre-v4 migrations. The entity registry and fitness linters ensure all active services are tracked — anything in `_archived` is definitively dead.

**Tasks**:
- Delete `internal/apps/identity/_archived/` (92 files)
- Delete `internal/apps/identity/_authz-archived/` through `_spa-archived/` (40 files)
- Delete `internal/apps/pki/_ca-archived/` (empty directory)
- Delete `internal/test/archived/` (15 files)
- Delete `deployments/archived/` (14 files)
- Delete `configs/orphaned/` (legacy configs, 6+ files including observability, template, test secrets)
- Verify `go build ./...` still clean after deletion
- Verify all 44 fitness linters still pass
- Verify all 68 deployment validators still pass

**Success**: Zero `_archived`, `archived/`, or `orphaned/` directories remain. Build and all linters pass.
**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 2: Non-Standard Entry Rationalization (6h) [Status: TODO]

**Objective**: Classify each non-standard cmd/ and internal/apps/ entry as INTENTIONAL INFRASTRUCTURE or VIOLATION, and fix violations.

**Analysis from repository survey**:
- `cmd/cicd/` + `internal/apps/cicd/`: INTENTIONAL INFRASTRUCTURE — provides `lint-*`, `format-*`, `github-cleanup` tooling. These are development tools, not product/service binaries. Documented in ARCHITECTURE.md Section 9.10.
- `cmd/workflow/` + `internal/apps/workflow/`: INTENTIONAL INFRASTRUCTURE — workflow testing tool. Referenced in CI/CD pipelines.
- `cmd/demo/` + `internal/apps/demo/`: EVALUATE — unified demo CLI. Consider whether this should be a subcommand of `cmd/cryptoutil` (suite-level CLI) rather than a standalone binary.
- `cmd/identity-compose/` + `internal/apps/identity/compose/`: VIOLATION — violates anti-pattern "NO executables for subcommands." Should be `cmd/identity compose` subcommand.
- `cmd/identity-demo/` + `internal/apps/identity/demo/`: VIOLATION — same anti-pattern. Should be `cmd/identity demo` subcommand.
- `internal/apps/pkiinit/`: EVALUATE — PKI initialization tool. May belong under `cmd/pki init` subcommand.

**Tasks**:
- Document `cmd/cicd/` and `cmd/workflow/` as intentional infrastructure tools in ARCHITECTURE.md Section 4.4.7
- Merge `cmd/identity-compose/` into `cmd/identity compose` subcommand (or archive if not used)
- Merge `cmd/identity-demo/` into `cmd/identity demo` subcommand (or archive if not used)
- Evaluate `cmd/demo/` — if useful, document as suite-level demo; if redundant, archive
- Evaluate `internal/apps/pkiinit/` — if useful, integrate into `cmd/pki init`; if redundant, archive
- Evaluate `docs/demo-brainstorm/` — archive if outdated
- Update entity registry and magic constants if entries change
- Verify build and linters pass

**Success**: All cmd/ entries are either documented infrastructure tools or follow the PRODUCT(-SERVICE) pattern. Zero anti-pattern violations.
**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 3: Configs Standardization (8h) [Status: TODO]

**Objective**: Apply rigid naming and structure standards to `configs/` directory, resolving the two-config-system duality and inconsistent naming.

**Design Decision Required**: What is the relationship between `configs/` (standalone) and `deployments/*/config/` (Docker)?

**Proposed model**:
- `deployments/*/config/` = Docker-focused: bind addresses, TLS cert paths as Docker secrets, Docker network hostnames. Standardized `{PS-ID}-app-{variant}.yml` naming. Already enforced by lint-deployments.
- `configs/` = Standalone/development-focused: domain-specific configs, environment configs, certificate profiles, auth policies. NOT duplicating what deployments/config/ provides.

**Key standardization rules for configs/**:
1. **Directory structure**: `configs/{PRODUCT}/{SERVICE}/` for all service-level configs (matching entity registry)
2. **Service template configs**: Use `{PS-ID}-{variant}.yml` naming (e.g., `sm-kms-pg-1.yml` not `config-pg-1.yml`)
3. **Domain configs**: Use `{PS-ID}-{purpose}.yml` naming (e.g., `pki-ca-server.yml` not `ca-server.yml`)
4. **Product-level configs**: `configs/{PRODUCT}/` for shared product configs (environment, policies, profiles)
5. **Suite-level configs**: `configs/cryptoutil/` for suite config
6. **Fix path mismatches**: Move `configs/ca/` content to `configs/pki/ca/` (matching entity registry product name)
7. **Empty directories**: Populate with `.gitkeep` or add proper configs
8. **No orphaned directory**: Already deleted in Phase 1

**Tasks**:
- Design and document configs/ canonical structure in ARCHITECTURE.md Section 12.5
- Rename `configs/ca/` to `configs/pki/ca/` (entity registry product is `pki`, not `ca`)
- Rename `configs/sm/kms/config-pg-1.yml` to `configs/sm/kms/sm-kms-pg-1.yml` (add PS-ID prefix)
- Rename `configs/sm/kms/config-pg-2.yml` to `configs/sm/kms/sm-kms-pg-2.yml`
- Rename `configs/sm/kms/config-sqlite.yml` to `configs/sm/kms/sm-kms-sqlite.yml`
- Rename `configs/sm/im/config-pg-1.yml` to `configs/sm/im/sm-im-pg-1.yml`
- Rename `configs/sm/im/config-pg-2.yml` to `configs/sm/im/sm-im-pg-2.yml`
- Rename `configs/sm/im/config-sqlite.yml` to `configs/sm/im/sm-im-sqlite.yml`
- Evaluate identity `-docker.yml` files for overlap with deployments/config/
- Document `configs/identity/` product-level configs (policies/, profiles/, environment YMLs)
- Populate empty `configs/jose/ja/` and `configs/pki/ca/` directories
- Update any code references to renamed config files
- Update lint_deployments mirror mapping for `configs/ca/` -> `configs/pki/ca/` rename
- Verify build and all linters pass

**Success**: All configs/ files follow `{PS-ID}-{purpose}.yml` naming. No empty directories without `.gitkeep`. `configs/` mirrors `deployments/` product structure. ARCHITECTURE.md Section 12.5 fully documents the standard.
**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 4: Deployments Refinement (4h) [Status: TODO]

**Objective**: Clean up product-level secret naming, remove template ambiguity, and verify compose delegation chain integrity.

**Issues identified**:
1. Product-level `deployments/{PRODUCT}/secrets/` may still have old-style naming alongside new-style
2. `deployments/template/` exists as the template for new services AND `deployments/skeleton-template/` is the skeleton service — naming overlap
3. Suite compose (1507 lines) may benefit from audit for unnecessary duplication

**Tasks**:
- Audit all `deployments/{PRODUCT}/secrets/` for old-style naming (e.g., `{product}-hash-pepper.secret` vs new `hash-pepper-v3.secret`)
- Standardize product-level secret naming to match documented pattern
- Document `deployments/template/` purpose vs `deployments/skeleton-template/` clearly
- Audit suite compose (1507 lines) for duplication opportunities
- Verify all 68 deployment validators still pass after changes
- Run `go run ./cmd/cicd lint-fitness` to confirm all 44 fitness linters pass

**Success**: All product-level secrets follow new naming convention. Template vs skeleton-template purpose documented. Delegation chain (SUITE->PRODUCT->SERVICE) verified.
**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 5: ARCHITECTURE.md Roadmap Consolidation (6h) [Status: TODO]

**Objective**: Ensure ARCHITECTURE.md is the complete SSOT for the project roadmap, absorbing content from satellite docs and adding missing strategies.

**Gaps identified**:
1. `docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md` (872 lines) — detailed compose tier documentation NOT in ARCHITECTURE.md. Should be merged into Section 12.3.
2. `configs/` standardization strategy — under-documented in Section 12.5 (currently describes current state, not target state)
3. Demo/workflow strategy — not documented anywhere as intentional infrastructure
4. Archive cleanup criteria — no defined policy for when code gets archived vs deleted
5. Non-standard cmd/ entries — Section 4.4.7 CLI Patterns doesn't document cicd/demo/workflow as intentional exceptions
6. Roadmap completion for LLM agents — user wants ARCHITECTURE.md to be sufficient for LLM agents to converge on end goals

**Tasks**:
- Merge ARCHITECTURE-COMPOSE-MULTIDEPLOY.md content into ARCHITECTURE.md Section 12.3 (compose tier deployment patterns)
- Delete ARCHITECTURE-COMPOSE-MULTIDEPLOY.md after merge (ARCHITECTURE.md is SSOT)
- Expand Section 12.5 with configs/ canonical naming standard from Phase 3
- Add Section 4.4.8 "Infrastructure CLI Tools" documenting cicd, workflow, demo as intentional non-product entries
- Add Section 13.9 "Archive and Dead Code Policy" defining when code is archived vs deleted
- Add or expand roadmap content section summarizing the vision: 1 suite / 5 products / 10 services with federation and 3-tier deployment
- Review all instruction files for alignment with updated ARCHITECTURE.md
- Run `go run ./cmd/cicd lint-docs` to verify propagation integrity

**Success**: ARCHITECTURE-COMPOSE-MULTIDEPLOY.md merged and deleted. ARCHITECTURE.md fully documents configs/ standardization, infrastructure CLI tools, archive policy, and roadmap vision. LLM agents reading ARCHITECTURE.md can understand the complete project structure and goals.
**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 6: Fitness Linter Expansion (6h) [Status: TODO]

**Objective**: Add new fitness linters to enforce standards established in Phases 1-5, preventing regression.

**New linters needed**:
1. **configs-naming**: Validate `configs/{PRODUCT}/{SERVICE}/` structure, `{PS-ID}-{variant}.yml` naming
2. **archive-detector**: Detect `_archived/`, `archived/`, `orphaned/` directories (ensure Phase 1 cleanup doesn't regress)
3. **cmd-anti-pattern**: Detect `cmd/{PRODUCT}-{subcommand}/` anti-pattern (e.g., `cmd/identity-compose/`)
4. **configs-empty-dir**: Detect empty directories in configs/ without `.gitkeep`
5. **configs-deployments-consistency**: Validate configs/ mirrors deployments/ service structure

**Tasks**:
- Implement `configs_naming` linter in `internal/apps/cicd/lint_fitness/`
- Implement `archive_detector` linter
- Implement `cmd_anti_pattern` linter
- Implement `configs_empty_dir` linter
- Implement `configs_deployments_consistency` linter
- Add tests for all new linters (>=98% coverage per infrastructure standard)
- Register new linters in the fitness catalog
- Verify all linters (44 existing + new) pass on current codebase
- Run mutation testing on new linter code

**Success**: 49+ fitness linters all pass. New linters prevent regression of Phase 1-5 work. >=98% coverage and mutation testing on all new linter code.
**Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 7: Knowledge Propagation (4h) [Status: TODO]

**Objective**: Apply lessons learned to permanent artifacts

- Review lessons.md from all prior phases
- Update ARCHITECTURE.md with new patterns and decisions
- Update agents (`.github/agents/*.agent.md`) with improved guidance
- Update skills (`.github/skills/*/SKILL.md`) with new patterns
- Update instructions (`.github/instructions/*.instructions.md`) with updated configs/deployment standards
- Verify propagation integrity (`go run ./cmd/cicd lint-docs validate-propagation`)

**Success**: All artifact updates committed; propagation check passes.

## Executive Decisions

### Decision 1: configs/ vs deployments/config/ Relationship

**Options**:
- A: Merge all configs/ into deployments/*/config/ (single config location)
- B: Keep both with clear separation — configs/ = domain/standalone, deployments/config/ = Docker-specific
- C: Move deployments/*/config/ into configs/ (centralize in configs/)
- D: Deprecate configs/ entirely, use only deployments/config/
- E:

**Decision**: Option B selected (tentative, pending quizme answer)

**Rationale**: Both serve distinct purposes. `deployments/config/` has Docker-specific settings (bind to 0.0.0.0, Docker network hostnames, TLS cert paths as Docker secrets). `configs/` has domain configs (CA profiles, auth policies), development settings, and standalone operation configs. Merging would create confusion between deployment-specific and domain-specific configuration.

**Impact**: Both config locations maintained with clear, documented separation. Naming standardized in both.

### Decision 2: Non-Standard cmd/ Entry Disposition

**Options**:
- A: Keep all as-is, just document them as intentional infrastructure
- B: Merge identity-compose and identity-demo into cmd/identity subcommands, keep cicd/demo/workflow
- C: Archive all demo entries (cmd/demo, cmd/identity-demo, internal/apps/demo), keep only cicd/workflow
- D: Move cicd/workflow under a new cmd/tools/ pattern
- E:

**Decision**: Option B selected (tentative, pending quizme answer)

**Rationale**: `cmd/cicd` and `cmd/workflow` are legitimate infrastructure tools documented in ARCHITECTURE.md. `cmd/identity-compose` and `cmd/identity-demo` violate the anti-pattern rule in Section 4.4.7 (NO executables for subcommands). Merging them into `cmd/identity compose` and `cmd/identity demo` subcommands follows the existing CLI pattern.

### Decision 3: Archive Deletion vs Preservation

**Options**:
- A: Delete all archived directories permanently (git history preserves content)
- B: Move archived to a separate branch before deleting from main
- C: Keep archived directories but add fitness linter to prevent growth
- D: Compress archived into .tar.gz files in docs/ for reference
- E:

**Decision**: Option A selected (tentative, pending quizme answer)

**Rationale**: Git history preserves all content permanently. The entity registry and fitness linters ensure active services are tracked. Archived code serves no purpose on main branch — it adds noise to searches, increases build times, and confuses LLM agents trying to understand the codebase.

### Decision 4: ARCHITECTURE-COMPOSE-MULTIDEPLOY.md Fate

**Options**:
- A: Merge into ARCHITECTURE.md Section 12.3 and delete
- B: Keep as a supplementary document referenced from ARCHITECTURE.md
- C: Convert to an instruction file (.github/instructions/)
- D: Delete without merging (ARCHITECTURE.md already covers the essentials)
- E:

**Decision**: Option A selected (tentative, pending quizme answer)

**Rationale**: ARCHITECTURE.md is the SSOT per project policy. The 872-line COMPOSE-MULTIDEPLOY doc contains detailed compose tier patterns that belong in Section 12.3. Keeping it separate creates information silos that LLM agents may miss.

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Config rename breaks runtime references | Medium | High | Search all Go code for config file path references before renaming |
| Archive deletion removes needed code | Low | Medium | Git history preserves all content; entity registry confirms active services |
| Fitness linter false positives | Medium | Medium | Table-driven tests with edge cases; run on real codebase before merging |
| ARCHITECTURE.md merge creates duplication | Low | Low | Careful deduplication during merge; lint-docs validates propagation |
| cmd/ consolidation breaks CI/CD | Medium | High | Check all workflow files for references to cmd/identity-compose and cmd/identity-demo |

## Quality Gates - MANDATORY

**Per-Action Quality Gates**:
- All tests pass (`go test ./...`) - 100% passing, zero skips
- Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`) - zero errors
- Linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`) - zero warnings
- No new TODOs without tracking in tasks.md

**Coverage Targets**:
- Production code: >=95% line coverage
- Infrastructure/utility code: >=98% line coverage
- main() functions: 0% acceptable if internalMain() >=95%
- Generated code: Excluded

**Mutation Testing Targets**:
- Production code: >=95% minimum
- Infrastructure/utility code: >=98% (NO EXCEPTIONS)

**Per-Phase Quality Gates**:
- Unit + integration tests complete before moving to next phase
- Deployment validators pass (`go run ./cmd/cicd lint-deployments`)
- Fitness linters pass (`go run ./cmd/cicd lint-fitness`)
- Race detector clean (`go test -race -count=2 ./...`) for modified packages

**Overall Project Quality Gates**:
- All phases complete with evidence
- All 44+ fitness linters passing
- All 68+ deployment validators passing
- CI/CD workflows green
- Documentation updated

## Success Criteria

- [ ] All phases complete
- [ ] Zero archived/orphaned directories remain
- [ ] All cmd/ entries documented or fixed
- [ ] configs/ uses rigid {PS-ID}-based naming
- [ ] ARCHITECTURE-COMPOSE-MULTIDEPLOY.md merged into ARCHITECTURE.md
- [ ] ARCHITECTURE.md sufficient for LLM agent convergence on roadmap goals
- [ ] 49+ fitness linters all passing
- [ ] 68+ deployment validators all passing
- [ ] All quality gates passing
- [ ] Evidence archived (test output, logs, analysis)

## ARCHITECTURE.md Cross-References - MANDATORY

| Topic | Section | When to Reference |
|-------|---------|-------------------|
| Testing Strategy | [Section 10](../../docs/ARCHITECTURE.md#10-testing-architecture) | ALL phases |
| Quality Gates | [Section 11.2](../../docs/ARCHITECTURE.md#112-quality-gates) | ALL phases |
| Coding Standards | [Section 13.1](../../docs/ARCHITECTURE.md#131-coding-standards) | Phases 2, 3, 6 |
| Version Control | [Section 13.2](../../docs/ARCHITECTURE.md#132-version-control) | ALL phases |
| Deployment Architecture | [Section 12](../../docs/ARCHITECTURE.md#12-deployment-architecture) | Phases 3, 4, 5 |
| Config File Architecture | [Section 12.5](../../docs/ARCHITECTURE.md#125-config-file-architecture) | Phase 3 |
| CLI Patterns | [Section 4.4.7](../../docs/ARCHITECTURE.md#447-cli-patterns) | Phase 2 |
| Entity Registry | [Section 9.11.2](../../docs/ARCHITECTURE.md#9112-entity-registry) | Phases 2, 3, 6 |
| Fitness Functions | [Section 9.11](../../docs/ARCHITECTURE.md#911-architecture-fitness-functions) | Phase 6 |
| Plan Lifecycle | [Section 13.6](../../docs/ARCHITECTURE.md#136-plan-lifecycle-management) | ALL phases |
| Post-Mortem | [Section 13.8](../../docs/ARCHITECTURE.md#138-phase-post-mortem--knowledge-propagation) | ALL phases |
| Infrastructure Blockers | [Section 13.7](../../docs/ARCHITECTURE.md#137-infrastructure-blocker-escalation) | ALL phases |
