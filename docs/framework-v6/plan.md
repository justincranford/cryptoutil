# Implementation Plan - Framework v6: Corrective Standardization

**Status**: In Progress
**Created**: 2026-03-25
**Last Updated**: 2026-03-27
**Purpose**: Fix all deployment, config, and secret standardization failures left by framework-v5, which marked all tasks Complete despite massive gaps. This plan is a corrective pass based on deep analysis of actual repository state vs. target-structure.md specifications.

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- Correctness: ALL documentation must be accurate and complete
- Completeness: NO phases or tasks or steps skipped, NO shortcuts
- Thoroughness: Evidence-based validation at every step
- Reliability: Quality gates enforced (>=95%/98% coverage/mutation)
- Efficiency: Optimized for maintainability and performance, NOT implementation speed
- Accuracy: Changes must address root cause, not just symptoms
- Time Pressure: NEVER rush, NEVER skip validation, NEVER defer quality checks
- Premature Completion: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions.**

## Overview

Framework-v5 marked all 7 phases and all tasks as Complete. Deep analysis reveals 17+ categories of discrepancies between the repository and target-structure.md. This plan fixes every verified gap, resolves target-structure.md self-contradictions, and brings the repository into full compliance.

## Background

Framework-v5 completed 7 phases: (1) Archive/Dead Code Cleanup, (2) Non-Standard Entry Rationalization, (3) Configs Standardization, (4) Deployments Refinement, (5) ARCHITECTURE.md Roadmap Consolidation, (6) Fitness Linter Expansion, (7) Knowledge Propagation. All tasks marked Complete. However, spot-checking by the user revealed the following systematic failures.

## Root Cause Analysis

### RC-1: target-structure.md Self-Contradictions

The specification document itself contains contradictions that made compliant implementation impossible:

**E.3 vs E.4 Conflict**: E.3 mandates FLAT `configs/{PS-ID}/` directories (e.g., `configs/sm-im/sm-im.yml`). E.4 mandates NESTED `configs/{PRODUCT}/{SERVICE}/` directories (e.g., `configs/sm/im/im.yml`). These are mutually exclusive. The `configs_naming` fitness linter enforces the NESTED pattern (E.4), meaning E.3 is de facto wrong.

**F.1 Unseal Pattern vs Example Conflict**: F.1 pattern says `{PS-ID}-unseal-key-N-of-5-{hex-random-32-bytes}` but the example shows `im-0d6dfc52f2517a2820e11859fe9e4f3c` (SERVICE-only prefix `im-`, not PS-ID `sm-im-`, and no `unseal-key-N-of-5` infix). The actual repo matches the example, not the pattern.

**F.2 Duplicate Listing**: F.2 lists `unseal-5of5.secret` twice with different semantics (one says "MUST ALWAYS be overridden at PRODUCT LEVEL" with `{PRODUCT}-unseal-key-5-of-5-{hex}`, the other says `dev-unseal-key-5-of-5`).

**Resolution Strategy**: Fix target-structure.md FIRST to resolve contradictions. Use the fitness linter enforcement as the source of truth for structural patterns. Use the actual working services (sm-im, sm-kms) as reference for value patterns where the spec self-contradicts.

### RC-2: Secret Value Pattern Violations

Multiple categories of incorrect secret values across all tiers:

**Service-level unseal prefixes** (10 services): All use SERVICE-only prefix instead of PS-ID prefix:
- `identity-authz`: prefix `authz-` (should be `identity-authz-`)
- `identity-idp`: prefix `idp-` (should be `identity-idp-`)
- `identity-rp`: prefix `rp-` (should be `identity-rp-`)
- `identity-rs`: prefix `rs-` (should be `identity-rs-`)
- `identity-spa`: prefix `spa-` (should be `identity-spa-`)
- `jose-ja`: prefix `ja-` (should be `jose-ja-`)
- `pki-ca`: prefix `kms-` (COPY-PASTE from sm-kms! hardcoded hex)
- `skeleton-template`: prefix `template-` (should be `skeleton-template-`)
- `sm-im`: prefix `im-` (should be `sm-im-`)
- `sm-kms`: prefix `kms-` (should be `sm-kms-`)

**pki-ca is a complete copy of sm-kms**: All 5 pki-ca unseal secrets contain `kms-11111111...` through `kms-55555555...` values - identical to sm-kms. This means pki-ca was copy-pasted from sm-kms without updating ANY secret values.

**Product-level unseal prefixes** (5 products): 4 of 5 use generic `dev-unseal-key-N-of-5` instead of `{PRODUCT}-unseal-key-N-of-5-{hex}`:
- `identity`: uses `identity-00000000...N-unseal-Nof5` (non-standard format but at least product-prefixed)
- `jose`, `pki`, `skeleton`, `sm`: all use generic `dev-unseal-key-N-of-5`

**Suite-level unseal prefixes**: `cryptoutil-suite` uses `suite-00000000...N` pattern. Spec says `{SUITE}-unseal-key-N-of-5-{hex-random-32-bytes}` = `cryptoutil-unseal-key-N-of-5-{hex}`. Actual uses `suite-` prefix not `cryptoutil-`.

**Postgres secret values**: Missing `_database` suffix in postgres-database.secret and missing `_database_` infix in postgres-username.secret:
- `sm-im`: db=`sm_im` (spec: `sm_im_database`), user=`sm_im_user` (spec: `sm_im_database_user`)
- `pki-ca`: db=`ca_db` (spec: `pki_ca_database`), user=`ca_user` (spec: `pki_ca_database_user`)
- `jose-ja`: db=`jose_ja` (spec: `jose_ja_database`), user=`ja_user` (spec: `jose_ja_database_user`)
- `skeleton-template`: db=`skeleton_template` (spec: `skeleton_template_database`), user=`template_user` (spec: `skeleton_template_database_user`)

### RC-3: Missing Files

**Zero .never files across entire repo**: target-structure.md F.2 requires 4 `.secret.never` files per product (20 total) and F.3 requires 4 per suite (4 total). Actual count: 0.

**Missing sqlite-2 config overlays**: All 10 services missing `{PS-ID}-app-sqlite-2.yml` in `deployments/{PS-ID}/config/`. F.1 spec requires it.

**Missing product-level configs**: No `{PRODUCT}.yml` files exist per E.2 spec (0 of 5: `sm.yml`, `jose.yml`, `pki.yml`, `identity.yml`, `skeleton.yml` all missing).

**Missing domain/ subdirectories**: E.4 specifies `configs/{PRODUCT}/{SERVICE}/domain/` directories for domain-specific config files. None exist.

### RC-4: Config File Naming Violations

- `configs/jose/ja/jose-ja-server.yml`: should be `ja.yml` per E.4
- `configs/pki/ca/pki-ca-server.yml`: should be `ca.yml` per E.4
- `configs/skeleton/template/skeleton-template-server.yml`: should be `template.yml` per E.4

### RC-5: Orphaned/Legacy Files Still Present

- `configs/sm/im/sm-im-pg-1.yml`, `sm-im-pg-2.yml`, `sm-im-sqlite.yml`: deployment variants in configs/ (belong in deployments/ or should be deleted per E.5)
- `configs/sm/kms/sm-kms-pg-1.yml`, `sm-kms-pg-2.yml`, `sm-kms-sqlite.yml`: same issue
- `configs/pki/ca/pki-ca-config-schema.yaml`: unclear purpose, not in spec
- `configs/pki/ca/profiles/` (25 yaml files): not in target-structure.md (may be domain config, needs decision)
- `configs/identity/policies/` (3 yaml files): not in target-structure.md
- `configs/skeleton/skeleton-server.yml`: orphaned product-level file with wrong name
- `deployments/template/`: template directory still exists (reconcile with skeleton-template)
- `deployments/deployments-all-files.json`: build artifact, should be deleted or gitignored
- `deployments/pki-ca/README.md`: not in spec

### RC-6: Deployment Secret Naming Inconsistencies

Template dir uses underscores in filenames (`unseal_1of5.secret`) while all services use hyphens (`unseal-1of5.secret`). The template should match the service convention.

### RC-7: Framework-v5 Quality Gate Failure

All 7 phases were marked Complete without catching any of the above issues. The quality gates either were not run or were insufficient to detect these systematic problems. The root cause is likely that framework-v5 focused on structural directory creation and file renaming but did not verify file CONTENTS or test the full target-structure.md compliance.

## Technical Context

- **Language**: Go 1.26.1
- **Entity Hierarchy**: 1 Suite (cryptoutil), 5 Products (identity, jose, pki, skeleton, sm), 10 PS-IDs
- **PS-IDs**: identity-authz, identity-idp, identity-rp, identity-rs, identity-spa, jose-ja, pki-ca, skeleton-template, sm-im, sm-kms
- **Key Spec**: `docs/framework-v6/target-structure.md`
- **Fitness Linters**: configs_naming, configs_deployments_consistency, configs_empty_dir, deployment_dir_completeness, standalone_config_presence, unseal_secret_content (NEW), dockerfile_labels (NEW), secret_naming (NEW)

## Phases

### Phase 1: Fix target-structure.md Contradictions (1h) [Status: IN PROGRESS]

**Objective**: Resolve all internal contradictions in the spec so subsequent phases have a single unambiguous source of truth.
- Fix E.3 vs E.4 contradiction: E.4 should document that configs/ uses FLAT `configs/{PS-ID}/` pattern (matching Decision 2=B), NOT nested `{PRODUCT}/{SERVICE}/` dirs. Rewrite E.4 to align with E.3.
- Fix F.1 unseal pattern vs example: Apply Decision 1=A. Pattern is `{PS-ID}-unseal-key-N-of-5-{hex-random-32-bytes}`. Update examples to match.
- Fix F.2 duplicate unseal-5of5 listing.
- Document .never file spec clearly (already clear in F.2/F.3, just ensure consistency).
- Document `configs/pki-ca/profiles/` exception per Decision 3=B.
- Document `configs/identity-authz/domain/policies/` per Decision 4=A.

**Completed Work** (commit `286ea4588`): Nine target-structure.md corrections committed:
- **C. cmd/**: Replaced nested tree with flat listing (18 entries: 1 SUITE + 5 PRODUCT + 10 PS-ID + 2 INFRA-TOOL)
- **E.1**: Added `{SUITE}` parameterized pattern before concrete `cryptoutil` expansion
- **F.1**: Replaced incomplete pki-ca examples with full 14-secret listings for both sm-im and pki-ca (all 5 unseal shards with unique hex)
- **F.3**: Added `{SUITE}` parameterized pattern followed by concrete `cryptoutil-suite` expansion
- **F.6 (NEW)**: Added Dockerfile Parameterization section documenting multi-stage structure, parameterized ARGs/LABELs, concrete PS-ID values table. Notes suite Dockerfile has incorrect OCI labels.
- **G.1**: Restructured into G.1.1 (Suite & Product), G.1.2 (Service with nested `{PRODUCT}/{SERVICE}/` and actual subdirectories table), G.1.3 (Framework & Tools with new linter entries)
- **L.**: Fixed secret.never statement — secret VALUES contain tier-specific prefixes while FILENAMES are identical across tiers. Replaced hardcoded `cryptoutil` with `{SUITE}` in suite column.
- **M.**: Expanded `configs-no-deployment` patterns to full variant list. Added `unseal-secret-content`, `dockerfile-labels`, and `secret-naming` linters.
- **N.**: Fixed mcp.json from "Missing → CREATE" to "Present (GitHub + Playwright MCP) → KEEP"

**Remaining Work**: Original tasks 1.1-1.4 (E.3/E.4 rewrite, F.1 pattern alignment, F.2 dedup, full consistency verification).
- **Success**: target-structure.md has zero internal contradictions. Every pattern has matching examples.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 2: Create Missing .never Files (0.5h) [Status: TODO]

**Objective**: Create all 24 missing .secret.never marker files.
- Create 4 .secret.never files in each of 5 product secret dirs (20 files)
- Create 4 .secret.never files in cryptoutil-suite secret dir (4 files)
- Verify: `Get-ChildItem deployments -Recurse -Filter "*.never" | Measure-Object` = 24
- **Success**: 24 .never files exist in correct locations.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 3: Fix Service-Level Secret Values (2h) [Status: TODO]

**Objective**: Fix all service-level secret values to match target-structure.md patterns.
- Fix all 10 services' unseal secret prefixes (requires Decision 1 resolution from Phase 1)
- Fix pki-ca: regenerate ALL secrets (currently copy-pasted from sm-kms)
- Fix postgres-database.secret: add `_database` suffix for all 10 services
- Fix postgres-username.secret: add `_database_` infix for all 10 services
- Fix postgres-url.secret: update to reflect corrected db/user values
- Fix postgres-password.secret: ensure PS_ID prefix pattern
- Verify all browser-username/service-username patterns
- **Success**: Every service secret matches the spec pattern with unique random values.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 4: Fix Product-Level and Suite-Level Secret Values (1h) [Status: TODO]

**Objective**: Fix all product-level and suite-level secret values.
- Fix 4 products' unseal secrets (jose, pki, skeleton, sm) from generic `dev-unseal-key-N-of-5` to `{PRODUCT}-unseal-key-N-of-5-{hex}`
- Normalize identity product unseal format to match other products
- Fix suite unseal prefix from `suite-` to `cryptoutil-` (per spec `{SUITE}-unseal-key-N-of-5-{hex}`)
- Fix product-level postgres secrets to use `{PRODUCT}_database` etc.
- **Success**: All product and suite secrets match spec patterns.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 5: Restructure Config Directories — Flat Pattern (3h) [Status: TODO]

**Objective**: Restructure ALL config directories from nested `configs/{PRODUCT}/{SERVICE}/` to flat `configs/{PS-ID}/` per Decision 2=B.
- Move `configs/sm/im/*` to `configs/sm-im/` and rename `im.yml` to `sm-im.yml`
- Move `configs/sm/kms/*` to `configs/sm-kms/` and rename to `sm-kms.yml`
- Move `configs/jose/ja/*` to `configs/jose-ja/` and rename `jose-ja-server.yml` to `jose-ja.yml` (also fixes RC-4)
- Move `configs/pki/ca/*` to `configs/pki-ca/` and rename `pki-ca-server.yml` to `pki-ca.yml` (also fixes RC-4). Keep `profiles/` subdir per Decision 3=B.
- Move `configs/skeleton/template/*` to `configs/skeleton-template/` and rename `skeleton-template-server.yml` to `skeleton-template.yml` (also fixes RC-4)
- Move all 5 identity service configs: `configs/identity/{authz,idp,rp,rs,spa}/*` to `configs/identity-{authz,idp,rp,rs,spa}/`
- Move `configs/identity/policies/` to `configs/identity-authz/domain/policies/` per Decision 4=A. Rename `adaptive-auth.yml` to `adaptive-authorization.yml` (terminology fix).
- Delete empty `configs/{PRODUCT}/{SERVICE}/` and `configs/{PRODUCT}/` directories after moves
- Delete `configs/skeleton/skeleton-server.yml` (orphaned, RC-5)
- Rewrite `configs_naming` fitness linter from nested to flat pattern validation
- Update `configs_deployments_consistency` fitness linter if needed
- Update ALL compose file config volume mount paths
- Update ALL Go code references to config paths
- **Success**: All configs under `configs/{PS-ID}/`. No nested `configs/{PRODUCT}/{SERVICE}/` remains. Fitness linters pass.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 6: Create Missing Config Files (1h) [Status: TODO]

**Objective**: Create all missing config files per E.2 and F.1.
- Create 10 missing `{PS-ID}-app-sqlite-2.yml` deployment config overlays per F.1
- Create product-level config files per E.2 if still applicable under flat structure (evaluate whether `{PRODUCT}.yml` at `configs/` root is needed)
- Create `domain/` subdirectories per spec for services that need them
- **Success**: All config files from spec exist.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 7: Clean Up Orphaned/Legacy Files (1h) [Status: TODO]

**Objective**: Remove or relocate files not in target-structure.md.
- Delete deployment variant configs: `configs/sm-im/sm-im-pg-*.yml`, `configs/sm-im/sm-im-sqlite.yml`, `configs/sm-kms/sm-kms-pg-*.yml`, `configs/sm-kms/sm-kms-sqlite.yml` (6 files, now under flat dirs)
- Delete `configs/pki-ca/pki-ca-config-schema.yaml` (schema is hardcoded in Go per ARCHITECTURE.md)
- Delete or gitignore `deployments/deployments-all-files.json`
- Delete `deployments/pki-ca/README.md` (not in spec)
- Reconcile `deployments/template/` per Decision 5=C: merge useful content into `deployments/skeleton-template/`, then delete `deployments/template/`
- Fix template dir secret filenames (underscores to hyphens) before merge
- Fix suite Dockerfile labels (currently says "CA Server" / "Certificate Authority REST API Server" — should match suite identity)
- **Success**: No orphaned files remain. All files in configs/ and deployments/ match spec. Suite Dockerfile labels correct.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 8: Fitness Linter Verification (1.5h) [Status: TODO]

**Objective**: Implement new fitness linters discovered during Phase 1 and run all linters to verify full compliance.
- Implement `unseal-secret-content` fitness linter: validates unseal key value patterns (correct PS-ID/PRODUCT/SUITE prefix, unique hex per shard, correct `N-of-5` infix, correct tier prefix matching deployment tier)
- Implement `dockerfile-labels` fitness linter: validates OCI labels in all Dockerfiles (correct PS-ID in `org.opencontainers.image.title`, correct description, version label present)
- Implement `secret-naming` fitness linter: validates secret filenames follow hyphen convention (no underscores), correct count per tier
- Run `go run ./cmd/cicd-lint lint-fitness`
- Run `go run ./cmd/cicd-lint lint-deployments`
- Run `go build ./...` and `go build -tags e2e,integration ./...`
- Run `golangci-lint run` and `golangci-lint run --build-tags e2e,integration`
- Fix any violations found
- **Success**: All linters pass (including 3 new linters). Zero violations.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 9: Terminology Enforcement (0.5h) [Status: TODO]

**Objective**: Fix all banned `auth` terminology violations and prevent recurrence.
- Rename `configs/identity/policies/adaptive-auth.yml` to `adaptive-authorization.yml` during Phase 5 move
- Audit all plan/task docs for banned `auth` term usage
- Verify no new `auth` violations in any generated or moved files
- **Root Cause**: AI agent generating content did not pre-check against `.github/instructions/01-01.terminology.instructions.md` banned terms
- **Prevention**: All generated content (docs, configs, filenames) must be scanned for banned terms before commit
- **Success**: Zero instances of standalone `auth` in any config filename, plan document, or generated content.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 10: Knowledge Propagation (1.5h) [Status: TODO]

**Objective**: Apply lessons learned to permanent artifacts.
- Review lessons.md from all prior phases
- Update ARCHITECTURE.md with any new patterns or corrections, specifically:
  - Deployment secret naming patterns with same specificity as target-structure.md (PS-ID prefix, tier differentiation, value format)
  - Dockerfile parameterization patterns (ARGs, LABELs, multi-stage structure per tier)
  - Flat config directory pattern `configs/{PS-ID}/` and exceptions (profiles/, domain/policies/)
  - New fitness linter catalog entries (unseal-secret-content, dockerfile-labels, secret-naming)
- Update instruction files (especially 02-01.architecture, 04-01.deployment) to match ARCHITECTURE.md specificity
- Update agents and skills if deployment/config patterns changed
- Update target-structure.md if any additional issues found during implementation
- Verify propagation integrity (`go run ./cmd/cicd-lint lint-docs validate-propagation`)
- **Success**: All artifact updates committed; propagation check passes. ARCHITECTURE.md has same specificity as target-structure.md for deployment patterns.

## Executive Decisions

### Decision 1: Unseal Secret Naming Pattern

**Options**:
- A: Use `{PS-ID}-unseal-key-N-of-5-{hex-random-32-bytes}` (full PS-ID prefix with descriptive infix) **SELECTED** ✓
- B: Use `{PS-ID}-{hex-random-32-bytes}` (compact PS-ID prefix)
- C: Keep current SERVICE-only prefix (e.g., `im-{hex}`, `kms-{hex}`) and update spec to match
- D: Use current values as-is and only fix pki-ca (which is a broken copy of sm-kms)
- E:

**Answer**: A

**Decision**: Full PS-ID prefix with descriptive infix: `{PS-ID}-unseal-key-N-of-5-{hex-random-32-bytes}` (e.g., `sm-im-unseal-key-1-of-5-a1b2c3d4e5f6...`). Most descriptive, prevents cross-service collision. Requires updating all 50 service + 25 product + 5 suite unseal files.

**Rationale**: Option A provides the most descriptive, unambiguous naming. Prevents all cross-service collisions. Pattern matches F.1 spec. All unseal files across all tiers must be regenerated.

### Decision 2: Config Directory Structure (E.3 vs E.4)

**Options**:
- A: Keep nested `configs/{PRODUCT}/{SERVICE}/` (matches current repo, matches fitness linter, matches E.4) and update E.3 to match
- B: Flatten to `configs/{PS-ID}/` per E.3 (requires restructuring ALL configs, updating fitness linter, massive change) **SELECTED** ✓
- C:
- D:
- E:

**Answer**: B

**Decision**: Flatten config directories to `configs/{PS-ID}/` per E.3. This is a MAJOR scope change: requires restructuring ALL config directories, rewriting the `configs_naming` fitness linter, updating all compose file config volume mounts, and updating E.4 to match E.3.

**Impact**:
- `configs/{PRODUCT}/{SERVICE}/` → `configs/{PS-ID}/` for all 10 services
- `configs_naming` fitness linter: complete rewrite from nested to flat pattern
- `configs_deployments_consistency` fitness linter: may need updates
- Config file naming: `{SERVICE}.yml` → `{PS-ID}.yml` (e.g., `im.yml` → `sm-im.yml`)
- All compose files referencing `configs/` paths must be updated
- Phase 6 (Config File Naming) becomes part of the restructuring
- E.4 must be rewritten to match E.3 flat pattern

**Rationale**: User selected flat structure despite recommendation against it. The flat `configs/{PS-ID}/` pattern provides simpler, more direct mapping between PS-ID and config directory.

### Decision 3: PKI CA Profiles Directory

**Options**:
- A: Move `configs/pki/ca/profiles/` to `configs/pki/ca/domain/profiles/` per E.4 domain pattern
- B: Keep profiles at current relative location, update spec to document exception **SELECTED** ✓
- C: Delete profiles (they are generated/unused)
- D: Move to `deployments/pki-ca/config/profiles/`
- E:

**Answer**: B

**Decision**: Keep profiles at their current relative location. With Decision 2=B (flat configs), profiles move from `configs/pki/ca/profiles/` to `configs/pki-ca/profiles/` as part of the directory flattening. Update target-structure.md to document `profiles/` as a valid subdirectory of `configs/pki-ca/`.

**Rationale**: These are real certificate profile definitions (root-ca.yaml, tls-server.yaml, etc.) used by pki-ca. No need for an additional `domain/` wrapper.

### Decision 4: Identity Policies Directory

**Options**:
- A: Move to `configs/identity-authz/domain/policies/` (authz-specific) **SELECTED** ✓
- B: Keep at `configs/identity/policies/` (product-level shared config)
- C: Delete (unused/orphaned)
- D: Move to `configs/identity/domain/policies/` (product-level domain config)
- E:

**Answer**: A

**Decision**: Move authorization policy files to `configs/identity-authz/domain/policies/`. With Decision 2=B (flat configs), the target path becomes `configs/identity-authz/domain/policies/`. These are authorization policies (adaptive-authorization.yml, risk-scoring.yml, step-up.yml) scoped to the authz service.

**Note**: The file `adaptive-auth.yml` uses banned `auth` abbreviation. Must be renamed to `adaptive-authorization.yml` per terminology instructions.

**Rationale**: Authorization policies are authz-specific. Scoping them to the identity-authz config directory provides clear ownership.

### Decision 5: Deployments Template Directory

**Options**:
- A: Delete `deployments/template/` entirely (skeleton-template is the canonical template)
- B: Keep as reference-only template (no deployment, documentation purpose)
- C: Merge useful content into `deployments/skeleton-template/` then delete **SELECTED** ✓
- D:
- E:

**Answer**: C

**Decision**: Merge any useful content from `deployments/template/` into `deployments/skeleton-template/`, then delete `deployments/template/`. The template/ dir has underscore-named secrets (`unseal_1of5.secret`) violating the hyphen convention — these will not be carried forward.

**Rationale**: skeleton-template is the canonical deployable template service. Any useful parameterized compose patterns from template/ should be preserved in skeleton-template before deletion.

### Decision 6: Postgres Secret Value Format

**Options**:
- A: Fix ALL services to match spec exactly: db=`{PS_ID}_database`, user=`{PS_ID}_database_user`, url uses corrected values **SELECTED** ✓
- B: Keep current shorter format (e.g., `sm_im` instead of `sm_im_database`) and update spec
- C:
- D:
- E:

**Answer**: A

**Decision**: Fix ALL services to match spec exactly. Database name: `{PS_ID}_database`, username: `{PS_ID}_database_user`, postgres-url updated to reflect corrected values. Compose files must be updated to match.

**Rationale**: Full spec compliance prevents confusion. Compose files referencing these values must also be updated.

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Compose files reference old secret values | High | High | grep all compose files for old values before committing |
| Fitness linters reject changes | Medium | Medium | Run linters after each phase |
| Go code references old config filenames | Medium | High | Search all .go files for old names |
| E2E tests break from secret changes | High | High | Run E2E suite after secret changes |
| target-structure.md changes cascade | Low | Low | target-structure.md is documentation, not enforcement |\n| Dockerfile labels inconsistent across tiers | High | Medium | New dockerfile-labels linter catches all mismatches; known: suite Dockerfile has pki-ca labels |

## Quality Gates - MANDATORY

**Per-Action Quality Gates**:
- All tests pass (`go test ./...`) - 100% passing, zero skips
- Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`) - zero errors
- Linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`) - zero warnings
- No new TODOs without tracking in tasks.md

**Per-Phase Quality Gates**:
- Fitness linters pass: `go run ./cmd/cicd-lint lint-fitness`
- Deployment validators pass: `go run ./cmd/cicd-lint lint-deployments`
- Docker Compose health checks pass (after secret changes)

## Success Criteria

- [ ] target-structure.md has zero internal contradictions
- [ ] 24 .never files exist in correct locations
- [ ] All 10 service secret values match spec patterns (unique, correctly prefixed)
- [ ] All 5 product secret values match spec patterns
- [ ] Suite secret values match spec patterns
- [ ] All missing config files created (product-level, sqlite-2, domain dirs)
- [ ] Config file naming matches E.4 pattern
- [ ] Zero orphaned files remain
- [ ] All fitness linters pass (including 3 new: unseal-secret-content, dockerfile-labels, secret-naming)
- [ ] All deployment validators pass
- [ ] All Dockerfiles have correct OCI labels matching their PS-ID
- [ ] Build and lint clean
- [ ] ARCHITECTURE.md updated with deployment pattern specificity matching target-structure.md
- [ ] Evidence archived

## ARCHITECTURE.md Cross-References - MANDATORY

| Topic | Section | Relevance |
|-------|---------|-----------|
| Testing Strategy | [Section 10](../../docs/ARCHITECTURE.md#10-testing-architecture) | E2E validation after secret changes |
| Quality Gates | [Section 11.2](../../docs/ARCHITECTURE.md#112-quality-gates) | Per-phase quality enforcement |
| Coding Standards | [Section 13.1](../../docs/ARCHITECTURE.md#131-coding-standards) | Config file naming patterns |
| Version Control | [Section 13.2](../../docs/ARCHITECTURE.md#132-version-control) | Incremental commit strategy |
| Deployment Architecture | [Section 12](../../docs/ARCHITECTURE.md#12-deployment-architecture) | Docker secrets, compose patterns |
| Config Architecture | [Section 12.5](../../docs/ARCHITECTURE.md#125-config-file-architecture) | Config file naming and structure |
| Secrets Management | [Section 12.6](../../docs/ARCHITECTURE.md#126-secrets-management-in-deployments) | Secret file patterns |
| Secret Detection | [Section 6.10](../../docs/ARCHITECTURE.md#610-secrets-detection-strategy) | Inline secret detection rules |
| Plan Lifecycle | [Section 13.6](../../docs/ARCHITECTURE.md#136-plan-lifecycle-management) | Plan execution patterns |
| Post-Mortem | [Section 13.8](../../docs/ARCHITECTURE.md#138-phase-post-mortem--knowledge-propagation) | Phase post-mortem requirements |
