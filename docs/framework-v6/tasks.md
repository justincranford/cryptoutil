# Tasks - Framework v6: Corrective Standardization

**Status**: 0 of 46 tasks complete (0%)
**Last Updated**: 2026-03-26
**Created**: 2026-03-25

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- Correctness: ALL code must be functionally correct with comprehensive tests
- Completeness: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- Thoroughness: Evidence-based validation at every step
- Reliability: Quality gates enforced (>=95%/98% coverage/mutation)
- Efficiency: Optimized for maintainability and performance, NOT implementation speed
- Accuracy: Changes must address root cause, not just symptoms
- Time Pressure: NEVER rush, NEVER skip validation, NEVER defer quality checks
- Premature Completion: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions.**

---

## Task Checklist

### Phase 1: Fix target-structure.md Contradictions

**Phase Objective**: Resolve all internal contradictions so subsequent phases have a single unambiguous source of truth.

#### Task 1.1: Fix E.3 vs E.4 Contradiction (Decision 2=B: Flatten)

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: -
- **Dependencies**: None
- **Description**: Rewrite E.4 to document the FLAT `configs/{PS-ID}/` pattern that matches Decision 2=B. Remove the nested `configs/{PRODUCT}/{SERVICE}/` language from E.4. Make E.4 consistent with E.3. Config files named `{PS-ID}.yml` (not `{SERVICE}.yml`).
- **Acceptance Criteria**:
  - [ ] E.4 describes flat `configs/{PS-ID}/` pattern matching E.3
  - [ ] No mention of nested `configs/{PRODUCT}/{SERVICE}/` directories in E.4
  - [ ] E.4 examples show `configs/sm-im/sm-im.yml`, `configs/pki-ca/pki-ca.yml` etc.
  - [ ] Service config files named `{PS-ID}.yml`
  - [ ] Document `configs/pki-ca/profiles/` exception per Decision 3=B
  - [ ] Document `configs/identity-authz/domain/policies/` per Decision 4=A
- **Files**:
  - `docs/framework-v5/target-structure.md`

#### Task 1.2: Fix F.1 Unseal Pattern vs Example (Decision 1=A)

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: None (Decision 1=A confirmed)
- **Description**: Update F.1 to use canonical pattern `{PS-ID}-unseal-key-N-of-5-{hex-random-32-bytes}`. Fix example to match (e.g., `sm-im-unseal-key-1-of-5-a1b2c3d4e5f6...`). Remove contradictory SERVICE-only prefix examples.
- **Acceptance Criteria**:
  - [ ] F.1 pattern is `{PS-ID}-unseal-key-N-of-5-{hex-random-32-bytes}`
  - [ ] F.1 examples match pattern (PS-ID prefix, descriptive infix)
  - [ ] No SERVICE-only prefix examples remain
  - [ ] Unseal naming convention is unambiguous
- **Files**:
  - `docs/framework-v5/target-structure.md`

#### Task 1.3: Fix F.2 Duplicate unseal-5of5 Listing

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: None
- **Description**: Remove duplicate `unseal-5of5.secret` entry in F.2 that contradicts the first listing.
- **Acceptance Criteria**:
  - [ ] Only one unseal-5of5.secret entry in F.2
  - [ ] Value pattern matches other unseal entries
- **Files**:
  - `docs/framework-v5/target-structure.md`

#### Task 1.4: Verify All Spec Sections Internally Consistent

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Tasks 1.1-1.3
- **Description**: Full read of target-structure.md to find any remaining contradictions. Verify Decision 3=B (profiles/ exception), Decision 4=A (identity-authz policies), and Decision 6=A (postgres values) are all reflected.
- **Acceptance Criteria**:
  - [ ] Zero internal contradictions remain
  - [ ] Every pattern has matching examples
  - [ ] Cross-references between sections are consistent
  - [ ] `configs/pki-ca/profiles/` documented as valid exception
  - [ ] `configs/identity-authz/domain/policies/` documented
  - [ ] Postgres value pattern `{PS_ID}_database` / `{PS_ID}_database_user` documented

### Phase 2: Create Missing .never Files

**Phase Objective**: Create all 24 missing .secret.never marker files per F.2 and F.3.

#### Task 2.1: Create Product-Level .never Files (20 files)

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: None
- **Description**: Create 4 .secret.never files in each of 5 product secret directories (identity, jose, pki, skeleton, sm).
- **Acceptance Criteria**:
  - [ ] `deployments/identity/secrets/browser-password.secret.never` exists
  - [ ] `deployments/identity/secrets/browser-username.secret.never` exists
  - [ ] `deployments/identity/secrets/service-password.secret.never` exists
  - [ ] `deployments/identity/secrets/service-username.secret.never` exists
  - [ ] Same 4 files for jose, pki, skeleton, sm (20 total)
  - [ ] File contents indicate "MUST NEVER be overridden at PRODUCT level"
- **Files**:
  - `deployments/{identity,jose,pki,skeleton,sm}/secrets/*.secret.never` (20 files)

#### Task 2.2: Create Suite-Level .never Files (4 files)

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 5m
- **Actual**: -
- **Dependencies**: None
- **Description**: Create 4 .secret.never files in cryptoutil-suite secret directory.
- **Acceptance Criteria**:
  - [ ] `deployments/cryptoutil-suite/secrets/browser-password.secret.never` exists
  - [ ] `deployments/cryptoutil-suite/secrets/browser-username.secret.never` exists
  - [ ] `deployments/cryptoutil-suite/secrets/service-password.secret.never` exists
  - [ ] `deployments/cryptoutil-suite/secrets/service-username.secret.never` exists
- **Files**:
  - `deployments/cryptoutil-suite/secrets/*.secret.never` (4 files)

#### Task 2.3: Verify .never File Count

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 5m
- **Actual**: -
- **Dependencies**: Tasks 2.1-2.2
- **Description**: Count all .never files to verify exactly 24 exist.
- **Acceptance Criteria**:
  - [ ] `Get-ChildItem deployments -Recurse -Filter "*.never" | Measure-Object` = 24

### Phase 3: Fix Service-Level Secret Values

**Phase Objective**: Fix all service-level secret values to match target-structure.md patterns.

#### Task 3.1: Regenerate pki-ca Secrets (CRITICAL)

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: -
- **Dependencies**: Phase 1 (Decision 1)
- **Description**: Regenerate ALL pki-ca secrets. Currently copy-pasted from sm-kms with `kms-` prefix and hardcoded hex. Generate unique random values with correct `pki-ca-` prefix pattern.
- **Acceptance Criteria**:
  - [ ] All 5 unseal secrets have `pki-ca-` prefix (not `kms-`)
  - [ ] All hex values are unique (not `11111111...` through `55555555...`)
  - [ ] postgres-database = `pki_ca_database`
  - [ ] postgres-username = `pki_ca_database_user`
  - [ ] postgres-url reflects corrected db/user values
  - [ ] browser-username = `pki-ca-browser-user`
  - [ ] service-username = `pki-ca-service-user`
  - [ ] No values match sm-kms secrets
- **Files**:
  - `deployments/pki-ca/secrets/*.secret` (14 files)

#### Task 3.2: Fix Unseal Prefixes for All 10 Services

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: -
- **Dependencies**: Phase 1 (Decision 1)
- **Description**: Update unseal secret prefixes from SERVICE-only to PS-ID for all services (or per Decision 1 outcome).
- **Acceptance Criteria**:
  - [ ] identity-authz: prefix `identity-authz-` (was `authz-`)
  - [ ] identity-idp: prefix `identity-idp-` (was `idp-`)
  - [ ] identity-rp: prefix `identity-rp-` (was `rp-`)
  - [ ] identity-rs: prefix `identity-rs-` (was `rs-`)
  - [ ] identity-spa: prefix `identity-spa-` (was `spa-`)
  - [ ] jose-ja: prefix `jose-ja-` (was `ja-`)
  - [ ] pki-ca: handled in Task 3.1
  - [ ] skeleton-template: prefix `skeleton-template-` (was `template-`)
  - [ ] sm-im: prefix `sm-im-` (was `im-`)
  - [ ] sm-kms: prefix `sm-kms-` (was `kms-`)
  - [ ] Each service has 5 unique hex values (not copied from another service)
- **Files**:
  - `deployments/*/secrets/unseal-*of5.secret` (50 files across 10 services)

#### Task 3.3: Fix Postgres Database Secrets for All 10 Services

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: None
- **Description**: Update postgres-database.secret values to include `_database` suffix per spec.
- **Acceptance Criteria**:
  - [ ] sm-im: `sm_im_database` (was `sm_im`)
  - [ ] sm-kms: `sm_kms_database` (was unknown, likely `sm_kms`)
  - [ ] pki-ca: `pki_ca_database` (was `ca_db`)
  - [ ] jose-ja: `jose_ja_database` (was `jose_ja`)
  - [ ] skeleton-template: `skeleton_template_database` (was `skeleton_template`)
  - [ ] All 5 identity services follow same pattern
- **Files**:
  - `deployments/*/secrets/postgres-database.secret` (10 files)

#### Task 3.4: Fix Postgres Username Secrets for All 10 Services

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: None
- **Description**: Update postgres-username.secret values to include `_database_user` suffix per spec.
- **Acceptance Criteria**:
  - [ ] sm-im: `sm_im_database_user` (was `sm_im_user`)
  - [ ] pki-ca: `pki_ca_database_user` (was `ca_user`)
  - [ ] jose-ja: `jose_ja_database_user` (was `ja_user`)
  - [ ] skeleton-template: `skeleton_template_database_user` (was `template_user`)
  - [ ] All 5 identity services follow same pattern
- **Files**:
  - `deployments/*/secrets/postgres-username.secret` (10 files)

#### Task 3.5: Fix Postgres URL Secrets for All 10 Services

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Tasks 3.3, 3.4
- **Description**: Update postgres-url.secret to reflect corrected database and username values.
- **Acceptance Criteria**:
  - [ ] Each URL uses `{PS_ID}_database_user:{PS_ID}_database_pass@{PS-ID}-postgres:5432/{PS_ID}_database`
  - [ ] No stale references to old db/user names
- **Files**:
  - `deployments/*/secrets/postgres-url.secret` (10 files)

#### Task 3.6: Fix Postgres Password Secrets for All 10 Services

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: None
- **Description**: Verify/fix postgres-password.secret values match `{PS_ID}_database_pass-{base64}` pattern.
- **Acceptance Criteria**:
  - [ ] All 10 services use `{PS_ID}_database_pass-{base64}` pattern
- **Files**:
  - `deployments/*/secrets/postgres-password.secret` (10 files)

#### Task 3.7: Fix Hash Pepper Secrets for All 10 Services

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: None
- **Description**: Verify/fix hash-pepper-v3.secret values match `{PS-ID}-hash-pepper-v3-{base64}` pattern.
- **Acceptance Criteria**:
  - [ ] All 10 services use `{PS-ID}-hash-pepper-v3-{base64}` pattern
- **Files**:
  - `deployments/*/secrets/hash-pepper-v3.secret` (10 files)

#### Task 3.8: Verify All Service Secrets Complete

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Tasks 3.1-3.7
- **Description**: Cross-check every service secret against spec. Verify no duplicate values across services.
- **Acceptance Criteria**:
  - [ ] Each of 10 services has exactly 14 secret files
  - [ ] No two services share any unseal hex values
  - [ ] No two services share postgres credentials
  - [ ] pki-ca values completely different from sm-kms

### Phase 4: Fix Product-Level and Suite-Level Secret Values

**Phase Objective**: Fix all product and suite secret values to match spec patterns.

#### Task 4.1: Fix Product Unseal Secrets (5 products)

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: -
- **Dependencies**: Phase 1
- **Description**: Replace generic `dev-unseal-key-N-of-5` with `{PRODUCT}-unseal-key-N-of-5-{hex}` for jose, pki, skeleton, sm. Normalize identity format.
- **Acceptance Criteria**:
  - [ ] jose: `jose-unseal-key-N-of-5-{hex}` (was `dev-unseal-key-N-of-5`)
  - [ ] pki: `pki-unseal-key-N-of-5-{hex}` (was `dev-unseal-key-N-of-5`)
  - [ ] skeleton: `skeleton-unseal-key-N-of-5-{hex}` (was `dev-unseal-key-N-of-5`)
  - [ ] sm: `sm-unseal-key-N-of-5-{hex}` (was `dev-unseal-key-N-of-5`)
  - [ ] identity: normalized to `identity-unseal-key-N-of-5-{hex}` format
  - [ ] All hex values unique per product
- **Files**:
  - `deployments/{identity,jose,pki,skeleton,sm}/secrets/unseal-*of5.secret` (25 files)

#### Task 4.2: Fix Product Postgres Secrets (5 products)

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: None
- **Description**: Verify/fix product-level postgres secrets match `{PRODUCT}_database`, `{PRODUCT}_database_user`, `{PRODUCT}_database_pass-{base64}` patterns.
- **Acceptance Criteria**:
  - [ ] Each product's postgres-database = `{PRODUCT}_database`
  - [ ] Each product's postgres-username = `{PRODUCT}_database_user`
  - [ ] Each product's postgres-url uses corrected values
  - [ ] Each product's postgres-password matches pattern
- **Files**:
  - `deployments/{identity,jose,pki,skeleton,sm}/secrets/postgres-*.secret` (20 files)

#### Task 4.3: Fix Product Hash Pepper Secrets (5 products)

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: None
- **Description**: Verify/fix product-level hash-pepper-v3.secret matches `{PRODUCT}-hash-pepper-v3-{base64}` pattern.
- **Acceptance Criteria**:
  - [ ] All 5 products match pattern
- **Files**:
  - `deployments/{identity,jose,pki,skeleton,sm}/secrets/hash-pepper-v3.secret`

#### Task 4.4: Fix Suite Secrets

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 1
- **Description**: Fix cryptoutil-suite secrets to use `cryptoutil-` prefix (not `suite-`). Fix unseal values to `cryptoutil-unseal-key-N-of-5-{hex}`. Fix postgres values to use `cryptoutil_database` pattern.
- **Acceptance Criteria**:
  - [ ] Unseal prefix: `cryptoutil-` (was `suite-`)
  - [ ] postgres-database: `cryptoutil_database`
  - [ ] postgres-username: `cryptoutil_database_user`
  - [ ] postgres-url: correct references
  - [ ] hash-pepper-v3: `cryptoutil-hash-pepper-v3-{base64}`
- **Files**:
  - `deployments/cryptoutil-suite/secrets/*.secret` (10 files)

#### Task 4.5: Verify All Product/Suite Secrets

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: Tasks 4.1-4.4
- **Description**: Cross-check all product and suite secrets against spec.
- **Acceptance Criteria**:
  - [ ] No generic `dev-unseal-key-N-of-5` values remain anywhere
  - [ ] No `suite-` prefix anywhere (should be `cryptoutil-`)
  - [ ] All products have unique unseal hex values

### Phase 5: Restructure Config Directories — Flat Pattern (Decision 2=B)

**Phase Objective**: Restructure ALL config directories from nested `configs/{PRODUCT}/{SERVICE}/` to flat `configs/{PS-ID}/` per Decision 2=B. This phase also absorbs the old Phase 6 (config file naming) since files are renamed during the move.

#### Task 5.1: Move sm-im Configs to Flat Structure

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 1
- **Description**: Move `configs/sm/im/*` to `configs/sm-im/`. Rename `im.yml` to `sm-im.yml`. Delete deployment variant configs (`sm-im-pg-*.yml`, `sm-im-sqlite.yml`) during move.
- **Acceptance Criteria**:
  - [ ] `configs/sm-im/sm-im.yml` exists
  - [ ] `configs/sm/im/` does not exist
  - [ ] No deployment variant configs carried forward
  - [ ] All compose/code references updated

#### Task 5.2: Move sm-kms Configs to Flat Structure

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 1
- **Description**: Move `configs/sm/kms/*` to `configs/sm-kms/`. Rename to `sm-kms.yml`. Delete deployment variant configs during move.
- **Acceptance Criteria**:
  - [ ] `configs/sm-kms/sm-kms.yml` exists
  - [ ] `configs/sm/kms/` does not exist
  - [ ] No deployment variant configs carried forward

#### Task 5.3: Move jose-ja Configs to Flat Structure

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 1
- **Description**: Move `configs/jose/ja/*` to `configs/jose-ja/`. Rename `jose-ja-server.yml` to `jose-ja.yml` (fixes RC-4 naming).
- **Acceptance Criteria**:
  - [ ] `configs/jose-ja/jose-ja.yml` exists (not `jose-ja-server.yml`)
  - [ ] `configs/jose/ja/` does not exist

#### Task 5.4: Move pki-ca Configs to Flat Structure

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 1, Decision 3=B
- **Description**: Move `configs/pki/ca/*` to `configs/pki-ca/`. Rename `pki-ca-server.yml` to `pki-ca.yml` (fixes RC-4). Keep `profiles/` subdir per Decision 3=B. Delete `pki-ca-config-schema.yaml` (schema hardcoded in Go).
- **Acceptance Criteria**:
  - [ ] `configs/pki-ca/pki-ca.yml` exists (not `pki-ca-server.yml`)
  - [ ] `configs/pki-ca/profiles/` exists with all 25 YAML files
  - [ ] `pki-ca-config-schema.yaml` deleted
  - [ ] `configs/pki/ca/` does not exist

#### Task 5.5: Move skeleton-template Configs to Flat Structure

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 1
- **Description**: Move `configs/skeleton/template/*` to `configs/skeleton-template/`. Rename `skeleton-template-server.yml` to `skeleton-template.yml` (fixes RC-4). Delete `configs/skeleton/skeleton-server.yml` (orphaned).
- **Acceptance Criteria**:
  - [ ] `configs/skeleton-template/skeleton-template.yml` exists (not `skeleton-template-server.yml`)
  - [ ] `configs/skeleton/template/` does not exist
  - [ ] `configs/skeleton/skeleton-server.yml` deleted

#### Task 5.6: Move All 5 Identity Service Configs to Flat Structure

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: -
- **Dependencies**: Phase 1, Decision 4=A
- **Description**: Move `configs/identity/{authz,idp,rp,rs,spa}/*` to `configs/identity-{authz,idp,rp,rs,spa}/`. Move `configs/identity/policies/` to `configs/identity-authz/domain/policies/` per Decision 4=A. Rename `adaptive-auth.yml` to `adaptive-authorization.yml` (terminology fix).
- **Acceptance Criteria**:
  - [ ] `configs/identity-authz/` exists with authz service config
  - [ ] `configs/identity-idp/` exists with idp service config
  - [ ] `configs/identity-rp/` exists with rp service config
  - [ ] `configs/identity-rs/` exists with rs service config
  - [ ] `configs/identity-spa/` exists with spa service config
  - [ ] `configs/identity-authz/domain/policies/` exists with 3 authorization policy files
  - [ ] `adaptive-authorization.yml` (not `adaptive-auth.yml`) in policies dir
  - [ ] `configs/identity/{authz,idp,rp,rs,spa}/` directories do not exist
  - [ ] `configs/identity/policies/` does not exist

#### Task 5.7: Delete Empty Parent Directories

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 5m
- **Actual**: -
- **Dependencies**: Tasks 5.1-5.6
- **Description**: Delete empty `configs/{PRODUCT}/{SERVICE}/` and `configs/{PRODUCT}/` directories after all moves complete.
- **Acceptance Criteria**:
  - [ ] No empty `configs/sm/`, `configs/jose/`, `configs/pki/`, `configs/skeleton/`, `configs/identity/` directories remain

#### Task 5.8: Rewrite configs_naming Fitness Linter

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Actual**: -
- **Dependencies**: Tasks 5.1-5.7
- **Description**: Rewrite `internal/apps/tools/cicd_lint/lint_fitness/configs_naming/configs_naming.go` to validate flat `configs/{PS-ID}/` pattern instead of nested `configs/{PRODUCT}/{SERVICE}/`. Update tests.
- **Acceptance Criteria**:
  - [ ] Linter validates `configs/{PS-ID}/` directories
  - [ ] Linter rejects nested `configs/{PRODUCT}/{SERVICE}/` pattern
  - [ ] All 10 PS-ID directories validated
  - [ ] `configs/pki-ca/profiles/` exception handled
  - [ ] `configs/identity-authz/domain/policies/` validated
  - [ ] Tests updated and passing with >=98% coverage
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes

#### Task 5.9: Update configs_deployments_consistency Linter

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: -
- **Dependencies**: Task 5.8
- **Description**: Update `configs_deployments_consistency` fitness linter to work with flat config directory structure.
- **Acceptance Criteria**:
  - [ ] Linter correctly maps `configs/{PS-ID}/` to `deployments/{PS-ID}/`
  - [ ] Tests updated and passing

#### Task 5.10: Update All Compose File Config Paths

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: -
- **Dependencies**: Tasks 5.1-5.7
- **Description**: Update all compose files across `deployments/` to reference new flat config paths instead of nested paths.
- **Acceptance Criteria**:
  - [ ] No compose file references `configs/{PRODUCT}/{SERVICE}/` paths
  - [ ] All config volume mounts use `configs/{PS-ID}/` paths
  - [ ] `docker compose config` validates for all services

#### Task 5.11: Update All Go Code Config Path References

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: -
- **Dependencies**: Tasks 5.1-5.7
- **Description**: Search all Go source files for references to old nested config paths and update.
- **Acceptance Criteria**:
  - [ ] Zero Go code references to `configs/{PRODUCT}/{SERVICE}/` paths
  - [ ] `go build ./...` passes
  - [ ] `golangci-lint run` passes

#### Task 5.12: Verify Flat Config Structure Complete

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: Tasks 5.1-5.11
- **Description**: Verify the entire configs/ directory matches the flat structure. List all directories under configs/ and confirm each is a valid PS-ID.
- **Acceptance Criteria**:
  - [ ] Only 10 PS-ID directories plus any documented exceptions exist under configs/
  - [ ] No nested `configs/{PRODUCT}/{SERVICE}/` remains
  - [ ] Fitness linters pass
  - [ ] Deployment validators pass

### Phase 6: Create Missing Config Files

**Phase Objective**: Create all missing config files per E.2 and F.1.

#### Task 6.1: Create Missing sqlite-2 Config Overlays (10 files)

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: -
- **Dependencies**: Phase 5 (flat structure in place)
- **Description**: Create `{PS-ID}-app-sqlite-2.yml` in every service's deployment config dir per F.1. Copy from sqlite-1 and adjust as needed.
- **Acceptance Criteria**:
  - [ ] `deployments/*/config/{PS-ID}-app-sqlite-2.yml` exists for all 10 services
  - [ ] Content matches sqlite-1 pattern (database-driver: sqlite, database-url: file::memory:?cache=shared)
- **Files**:
  - `deployments/*/config/{PS-ID}-app-sqlite-2.yml` (10 files)

#### Task 6.2: Evaluate Product-Level Config Files

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 5 (flat structure)
- **Description**: Under flat `configs/{PS-ID}/` structure, evaluate whether product-level configs (E.2's `{PRODUCT}.yml`) are still applicable. If so, determine placement. If not, update spec.
- **Acceptance Criteria**:
  - [ ] Decision documented on whether product-level configs are needed under flat structure
  - [ ] If needed: files created at appropriate location
  - [ ] If not needed: target-structure.md E.2 updated

### Phase 7: Clean Up Orphaned/Legacy Files

**Phase Objective**: Remove or relocate files not in target-structure.md.

#### Task 7.1: Delete Deployment Artifact Files

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: None
- **Description**: Delete `deployments/deployments-all-files.json` (build artifact) and `deployments/pki-ca/README.md` (not in spec).
- **Acceptance Criteria**:
  - [ ] `deployments/deployments-all-files.json` deleted or gitignored
  - [ ] `deployments/pki-ca/README.md` deleted

#### Task 7.2: Reconcile deployments/template/ (Decision 5=C)

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: -
- **Dependencies**: None
- **Description**: Merge useful content from `deployments/template/` into `deployments/skeleton-template/` per Decision 5=C. Fix template secret filenames (underscores to hyphens: `unseal_1of5.secret` -> `unseal-1of5.secret`). Then delete `deployments/template/` entirely.
- **Acceptance Criteria**:
  - [ ] Any useful parameterized compose patterns merged into skeleton-template
  - [ ] `deployments/template/` deleted entirely
  - [ ] No underscore-named secret files remain anywhere

#### Task 7.3: Verify No Orphans Remain

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: Tasks 7.1-7.2, Phase 5
- **Description**: Scan configs/ and deployments/ for files not in target-structure.md. Note: configs/ cleanup (variant configs, schema, orphaned files, identity policies) was handled in Phase 5 during restructuring.
- **Acceptance Criteria**:
  - [ ] Every file in configs/ matches flat `configs/{PS-ID}/` spec
  - [ ] Every file in deployments/ matches F.1-F.3 spec
  - [ ] No `deployments/template/` directory exists

### Phase 8: Fitness Linter Verification

**Phase Objective**: Run all linters to verify full compliance.

#### Task 8.1: Run Fitness Linters

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phases 1-7
- **Description**: Run `go run ./cmd/cicd-lint lint-fitness` and fix any violations.
- **Acceptance Criteria**:
  - [ ] Zero violations from lint-fitness

#### Task 8.2: Run Deployment Validators

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phases 1-7
- **Description**: Run `go run ./cmd/cicd-lint lint-deployments` and fix any violations.
- **Acceptance Criteria**:
  - [ ] Zero violations from lint-deployments

#### Task 8.3: Run Build and Lint

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: Phases 1-7
- **Description**: Verify `go build ./...`, `go build -tags e2e,integration ./...`, `golangci-lint run`, `golangci-lint run --build-tags e2e,integration` all pass.
- **Acceptance Criteria**:
  - [ ] Build clean (both standard and tagged)
  - [ ] Lint clean (both standard and tagged)

#### Task 8.4: Run Compose Validation

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phases 3-4 (secret changes)
- **Description**: Verify compose files still reference correct secret names and values work.
- **Acceptance Criteria**:
  - [ ] All compose files reference valid secret file paths
  - [ ] `docker compose config` passes for representative services

### Phase 9: Terminology Enforcement

**Phase Objective**: Fix all banned `auth` terminology violations and establish prevention controls.

#### Task 9.1: Audit All Generated Files for Banned Terms

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 5 (config restructuring done, files in final locations)
- **Description**: Scan all config filenames, plan docs, and generated content for standalone `auth` (banned per `.github/instructions/01-01.terminology.instructions.md`). Must use `authn`, `authz`, `authentication`, or `authorization` instead.
- **Acceptance Criteria**:
  - [ ] `adaptive-auth.yml` renamed to `adaptive-authorization.yml` (done in Phase 5 Task 5.6)
  - [ ] Zero instances of standalone `auth` in any config filename
  - [ ] Zero instances of `auth` used as abbreviation in plan.md or tasks.md (context-dependent: e.g., `authn` and `authz` are correct)
  - [ ] No `auth` in any newly generated content

#### Task 9.2: Document Root Cause and Prevention

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: Task 9.1
- **Description**: Document in lessons.md: Root cause = AI agent generating content did not pre-check against banned terms list. Prevention = All generated/moved filenames and content must be scanned for banned terms before commit.
- **Acceptance Criteria**:
  - [ ] Root cause documented in lessons.md Phase 9 section
  - [ ] Prevention strategy documented

### Phase 10: Knowledge Propagation

**Phase Objective**: Apply lessons learned to permanent artifacts.

#### Task 10.1: Review Lessons and Update ARCHITECTURE.md

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: -
- **Dependencies**: All prior phases
- **Description**: Review lessons.md, update ARCHITECTURE.md with new patterns including flat config structure and terminology enforcement.
- **Acceptance Criteria**:
  - [ ] lessons.md reviewed
  - [ ] ARCHITECTURE.md updated where needed (flat config pattern, terminology enforcement)
  - [ ] Propagation check passes

#### Task 10.2: Update Instruction Files

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Task 10.1
- **Description**: Update instruction files if config/deployment patterns changed (especially 02-01.architecture.instructions.md for flat config pattern).
- **Acceptance Criteria**:
  - [ ] Instructions reflect flat `configs/{PS-ID}/` pattern
  - [ ] Terminology enforcement guidance strengthened

#### Task 10.3: Update target-structure.md with Implementation Learnings

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: All prior phases
- **Description**: Document any additional findings from implementation in target-structure.md.
- **Acceptance Criteria**:
  - [ ] target-structure.md is fully accurate vs repo state

---

## Cross-Cutting Tasks

### Testing

- [ ] Build passes after every phase
- [ ] Fitness linters pass after Phase 5 (config restructuring) and Phase 7 (cleanup)
- [ ] Deployment validators pass after Phase 8
- [ ] E2E tests pass after secret changes (if Docker Desktop available)

### Code Quality

- [ ] No new TODOs without tracking
- [ ] All file naming consistent
- [ ] No orphaned references to old filenames

### Documentation

- [ ] target-structure.md self-consistent
- [ ] ARCHITECTURE.md updated if patterns changed
- [ ] Instruction files updated if needed

### Deployment

- [ ] All compose files reference valid secrets
- [ ] All config overlay files exist per spec
- [ ] All .never marker files in place

---

## Notes / Deferred Work

- All 6 decisions confirmed (merged from quizme-v1.md)
- Decision 2=B (flatten configs) added Phase 5 with 12 new tasks — major scope increase
- Phase 9 (Terminology Enforcement) added for `auth` → `authz`/`authorization` fixes
- Total phases: 10 (was 9). Total tasks: 46 (was 52, restructured: old Phases 5-6 merged into new Phase 5 with 12 tasks)
- E2E testing may be blocked if Docker Desktop is not running

---

## Evidence Archive

- `test-output/framework-v6/phase0-research/` - Pre-plan research and analysis
