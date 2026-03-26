# Tasks - Framework v6: Corrective Standardization

**Status**: 41 of 63 tasks complete (65%)
**Last Updated**: 2026-03-28
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

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: -
- **Dependencies**: None
- **Description**: Rewrite E.4 to document the FLAT `configs/{PS-ID}/` pattern that matches Decision 2=B. Remove the nested `configs/{PRODUCT}/{SERVICE}/` language from E.4. Make E.4 consistent with E.3. Config files named `{PS-ID}.yml` (not `{SERVICE}.yml`).
- **Acceptance Criteria**:
  - [x] E.4 describes flat `configs/{PS-ID}/` pattern matching E.3 (E.4 is now "What Gets DELETED", E.3 documents flat target)
  - [x] No mention of nested `configs/{PRODUCT}/{SERVICE}/` directories in E.4
  - [x] E.3 examples show `configs/sm-im/sm-im.yml`, `configs/pki-ca/pki-ca.yml` etc.
  - [x] Service config files named `{PS-ID}.yml`
  - [x] Document `configs/pki-ca/profiles/` exception per Decision 3=B
  - [x] Document `configs/identity-authz/domain/policies/` per Decision 4=A
- **Files**:
  - `docs/framework-v6/target-structure.md`

#### Task 1.2: Fix F.1 Unseal Pattern vs Example (Decision 1=A)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: None (Decision 1=A confirmed)
- **Description**: Update F.1 to use canonical pattern `{PS-ID}-unseal-key-N-of-5-{hex-random-32-bytes}`. Fix example to match (e.g., `sm-im-unseal-key-1-of-5-a1b2c3d4e5f6...`). Remove contradictory SERVICE-only prefix examples.
- **Acceptance Criteria**:
  - [x] F.1 pattern is `{PS-ID}-unseal-key-N-of-5-{hex-random-32-bytes}`
  - [x] F.1 examples match pattern (PS-ID prefix, descriptive infix)
  - [x] No SERVICE-only prefix examples remain
  - [x] Unseal naming convention is unambiguous
- **Files**:
  - `docs/framework-v5/target-structure.md`

#### Task 1.3: Fix F.2 Duplicate unseal-5of5 Listing

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: None
- **Description**: Remove duplicate `unseal-5of5.secret` entry in F.2 that contradicts the first listing.
- **Acceptance Criteria**:
  - [x] Only one unseal-5of5.secret entry in F.2
  - [x] Value pattern matches other unseal entries
- **Files**:
  - `docs/framework-v5/target-structure.md`

#### Task 1.4: Verify All Spec Sections Internally Consistent

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Tasks 1.1-1.3
- **Description**: Full read of target-structure.md to find any remaining contradictions. Verify Decision 3=B (profiles/ exception), Decision 4=A (identity-authz policies), and Decision 6=A (postgres values) are all reflected.
- **Acceptance Criteria**:
  - [x] Zero internal contradictions remain (16 corrections documented in header table)
  - [x] Every pattern has matching examples (F.1, F.2, F.3 concrete examples)
  - [x] Cross-references between sections are consistent (L. naming table matches F.1-F.3)
  - [x] `configs/pki-ca/profiles/` documented as valid exception (E.3)
  - [x] `configs/identity-authz/domain/policies/` documented (E.3)
  - [x] Postgres value pattern `{PS_ID}_database` / `{PS_ID}_database_user` documented (F.1, L.)

#### Task 1.5: Fix C. cmd/ Flat Structure (commit 286ea4588)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: 10m
- **Dependencies**: None
- **Description**: Replace misleading nested cmd/ tree with flat listing showing all 18 entries as direct children (1 SUITE + 5 PRODUCT + 10 PS-ID + 2 INFRA-TOOL).
- **Acceptance Criteria**:
  - [x] cmd/ tree shows flat structure
  - [x] All 18 entries listed with type annotations
  - [x] Concrete values shown (not just parameters)

#### Task 1.6: Fix E.1 and F.3 Suite Parameterization (commit 286ea4588)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: 5m
- **Dependencies**: None
- **Description**: Add `{SUITE}` parameterized patterns before concrete `cryptoutil`/`cryptoutil-suite` expansions in E.1 and F.3.
- **Acceptance Criteria**:
  - [x] E.1 shows `{SUITE}` pattern + concrete expansion
  - [x] F.3 shows `{SUITE}` pattern + concrete expansion

#### Task 1.7: Fix F.1 pki-ca Examples (commit 286ea4588)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: 15m
- **Dependencies**: None
- **Description**: Replace incomplete pki-ca examples with full 14-secret listings for both sm-im and pki-ca, including all 5 unseal shards with unique hex values.
- **Acceptance Criteria**:
  - [x] sm-im example: 14 secrets, `sm-im-` prefix, unique hex
  - [x] pki-ca example: 14 secrets, `pki-ca-` prefix, unique hex (not copied from sm-kms)

#### Task 1.8: Add F.6 Dockerfile Parameterization (commit 286ea4588)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: 25m
- **Dependencies**: None
- **Description**: Add new section documenting Dockerfile parameterization: multi-stage structure (validation → builder → runtime), parameterized ARGs/LABELs per tier, concrete PS-ID values table. Notes suite Dockerfile has incorrect OCI labels.
- **Acceptance Criteria**:
  - [x] F.6 section added to target-structure.md
  - [x] All three tiers (SERVICE, PRODUCT, SUITE) documented
  - [x] Concrete values table for all 10 PS-IDs
  - [x] Suite Dockerfile label issue flagged

#### Task 1.9: Fix G.1 internal/apps/ Structure (commit 286ea4588, corrected in follow-up)

- **Status**: ✅ Complete (with correction applied)
- **Owner**: LLM Agent
- **Estimated**: 25m
- **Actual**: 20m
- **Dependencies**: None
- **Description**: Restructure G.1 into G.1.1 (Suite & Product), G.1.2 (Service with flat `{PS-ID}/` pattern and actual subdirectory table), G.1.3 (Framework & Tools with new linter entries). **NOTE**: Original commit 286ea4588 incorrectly documented the NESTED `{PRODUCT}/{SERVICE}/` pattern. A follow-up correction rewrote G.1 intro and G.1.2 to describe the FLAT `internal/apps/{PS-ID}/` target (matching `cmd/{PS-ID}/` and `deployments/{PS-ID}/`).
- **Acceptance Criteria**:
  - [x] G.1.1 documents suite and product app patterns
  - [x] G.1.2 documents **flat** `{PS-ID}/` service pattern (NOT nested)
  - [x] G.1.3 documents framework/tools with new linters
  - [x] Comprehensive subdirectory table per PS-ID

#### Task 1.10: Fix L. Secret.never Statement (commit 286ea4588)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: 5m
- **Dependencies**: None
- **Description**: Correct misleading statement. Secret VALUES contain tier-specific prefixes; FILENAMES are identical across tiers. Replace hardcoded `cryptoutil` with `{SUITE}` in suite column.
- **Acceptance Criteria**:
  - [x] Statement corrected: values have prefix, not filenames
  - [x] Suite column uses `{SUITE}` parameter

#### Task 1.11: Fix M. configs-no-deployment and Add Linters (commit 286ea4588)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: 10m
- **Dependencies**: None
- **Description**: Expand `configs-no-deployment` to full variant list (`*-pg-*.yml`, `*-postgresql-*.yml`, `*-sqlite.yml`, `*-sqlite-*.yml`, plus environment files). Add three new linters: `unseal-secret-content`, `dockerfile-labels`, `secret-naming`.
- **Acceptance Criteria**:
  - [x] Full variant pattern list documented
  - [x] unseal-secret-content linter spec added
  - [x] dockerfile-labels linter spec added
  - [x] secret-naming linter spec added

#### Task 1.12: Fix N. mcp.json Status (commit 286ea4588)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 5m
- **Actual**: 3m
- **Dependencies**: None
- **Description**: Change mcp.json table entry from "Missing → Present → CREATE" to "Present (GitHub + Playwright MCP servers) → Present → KEEP (no change)".
- **Acceptance Criteria**:
  - [x] N. table shows "Present → KEEP"

#### Task 1.13: Verify F.2 Duplicate unseal-5of5 Fixed

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: Tasks 1.5-1.12
- **Description**: Verify that the F.1 rewrite (Task 1.7) resolved the F.2 duplicate unseal-5of5 listing, or make a targeted fix if not.
- **Acceptance Criteria**:
  - [x] Only one unseal-5of5 entry per example in F.1/F.2
  - [x] Value patterns consistent

### Phase 2: Create Missing .never Files

**Phase Objective**: Create all 24 missing .secret.never marker files per F.2 and F.3.

#### Task 2.1: Create Product-Level .never Files (20 files)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: None
- **Description**: Create 4 .secret.never files in each of 5 product secret directories (identity, jose, pki, skeleton, sm).
- **Acceptance Criteria**:
  - [x] `deployments/identity/secrets/browser-password.secret.never` exists
  - [x] `deployments/identity/secrets/browser-username.secret.never` exists
  - [x] `deployments/identity/secrets/service-password.secret.never` exists
  - [x] `deployments/identity/secrets/service-username.secret.never` exists
  - [x] Same 4 files for jose, pki, skeleton, sm (20 total)
  - [x] File contents indicate "MUST NEVER be overridden at PRODUCT level"
- **Files**:
  - `deployments/{identity,jose,pki,skeleton,sm}/secrets/*.secret.never` (20 files)

#### Task 2.2: Create Suite-Level .never Files (4 files)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 5m
- **Actual**: -
- **Dependencies**: None
- **Description**: Create 4 .secret.never files in cryptoutil-suite secret directory.
- **Acceptance Criteria**:
  - [x] `deployments/cryptoutil-suite/secrets/browser-password.secret.never` exists
  - [x] `deployments/cryptoutil-suite/secrets/browser-username.secret.never` exists
  - [x] `deployments/cryptoutil-suite/secrets/service-password.secret.never` exists
  - [x] `deployments/cryptoutil-suite/secrets/service-username.secret.never` exists
- **Files**:
  - `deployments/cryptoutil-suite/secrets/*.secret.never` (4 files)

#### Task 2.3: Verify .never File Count

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 5m
- **Actual**: -
- **Dependencies**: Tasks 2.1-2.2
- **Description**: Count all .never files to verify exactly 24 exist.
- **Acceptance Criteria**:
  - [x] `Get-ChildItem deployments -Recurse -Filter "*.never" | Measure-Object` = 24

### Phase 3: Fix Service-Level Secret Values

**Phase Objective**: Fix all service-level secret values to match target-structure.md patterns.

#### Task 3.1: Regenerate pki-ca Secrets (CRITICAL)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: 5m
- **Dependencies**: Phase 1 (Decision 1)
- **Description**: Regenerate ALL pki-ca secrets. Currently copy-pasted from sm-kms with `kms-` prefix and hardcoded hex. Generate unique random values with correct `pki-ca-` prefix pattern.
- **Acceptance Criteria**:
  - [x] All 5 unseal secrets have `pki-ca-` prefix (not `kms-`)
  - [x] All hex values are unique (not `11111111...` through `55555555...`)
  - [x] postgres-database = `pki_ca_database`
  - [x] postgres-username = `pki_ca_database_user`
  - [x] postgres-url reflects corrected db/user values
  - [x] browser-username = `pki-ca-browser-user`
  - [x] service-username = `pki-ca-service-user`
  - [x] No values match sm-kms secrets
- **Files**:
  - `deployments/pki-ca/secrets/*.secret` (14 files)

#### Task 3.2: Fix Unseal Prefixes for All 10 Services

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: 5m
- **Dependencies**: Phase 1 (Decision 1)
- **Description**: Update unseal secret prefixes from SERVICE-only to PS-ID for all services (or per Decision 1 outcome).
- **Acceptance Criteria**:
  - [x] identity-authz: prefix `identity-authz-` (was `authz-`)
  - [x] identity-idp: prefix `identity-idp-` (was `idp-`)
  - [x] identity-rp: prefix `identity-rp-` (was `rp-`)
  - [x] identity-rs: prefix `identity-rs-` (was `rs-`)
  - [x] identity-spa: prefix `identity-spa-` (was `spa-`)
  - [x] jose-ja: prefix `jose-ja-` (was `ja-`)
  - [x] pki-ca: handled in Task 3.1
  - [x] skeleton-template: prefix `skeleton-template-` (was `template-`)
  - [x] sm-im: prefix `sm-im-` (was `im-`)
  - [x] sm-kms: prefix `sm-kms-` (was `kms-`)
  - [x] Each service has 5 unique hex values (not copied from another service)
- **Files**:
  - `deployments/*/secrets/unseal-*of5.secret` (50 files across 10 services)

#### Task 3.3: Fix Postgres Database Secrets for All 10 Services

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: 5m
- **Dependencies**: None
- **Description**: Update postgres-database.secret values to include `_database` suffix per spec.
- **Acceptance Criteria**:
  - [x] sm-im: `sm_im_database` (was `sm_im`)
  - [x] sm-kms: `sm_kms_database` (was `cryptoutil_test`)
  - [x] pki-ca: `pki_ca_database` (was `ca_db`)
  - [x] jose-ja: `jose_ja_database` (was `jose_ja`)
  - [x] skeleton-template: `skeleton_template_database` (was `skeleton_template`)
  - [x] All 5 identity services follow same pattern
- **Files**:
  - `deployments/*/secrets/postgres-database.secret` (10 files)

#### Task 3.4: Fix Postgres Username Secrets for All 10 Services

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: 5m
- **Dependencies**: None
- **Description**: Update postgres-username.secret values to include `_database_user` suffix per spec.
- **Acceptance Criteria**:
  - [x] sm-im: `sm_im_database_user` (was `sm_im_user`)
  - [x] pki-ca: `pki_ca_database_user` (was `ca_user`)
  - [x] jose-ja: `jose_ja_database_user` (was `ja_user`)
  - [x] skeleton-template: `skeleton_template_database_user` (was `template_user`)
  - [x] All 5 identity services follow same pattern
- **Files**:
  - `deployments/*/secrets/postgres-username.secret` (10 files)

#### Task 3.5: Fix Postgres URL Secrets for All 10 Services

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: 5m
- **Dependencies**: Tasks 3.3, 3.4
- **Description**: Update postgres-url.secret to reflect corrected database and username values.
- **Acceptance Criteria**:
  - [x] Each URL uses `{PS_ID}_database_user:{PS_ID}_database_pass@{PS-ID}-postgres:5432/{PS_ID}_database`
  - [x] No stale references to old db/user names
- **Files**:
  - `deployments/*/secrets/postgres-url.secret` (10 files)

#### Task 3.6: Fix Postgres Password Secrets for All 10 Services

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: 5m
- **Dependencies**: None
- **Description**: Verify/fix postgres-password.secret values match `{PS_ID}_database_pass-{base64}` pattern.
- **Acceptance Criteria**:
  - [x] All 10 services use `{PS_ID}_database_pass-{base64}` pattern
- **Files**:
  - `deployments/*/secrets/postgres-password.secret` (10 files)

#### Task 3.7: Fix Hash Pepper Secrets for All 10 Services

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: 5m
- **Dependencies**: None
- **Description**: Verify/fix hash-pepper-v3.secret values match `{PS-ID}-hash-pepper-v3-{base64}` pattern.
- **Acceptance Criteria**:
  - [x] All 10 services use `{PS-ID}-hash-pepper-v3-{base64}` pattern
- **Files**:
  - `deployments/*/secrets/hash-pepper-v3.secret` (10 files)

#### Task 3.8: Verify All Service Secrets Complete

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: 5m
- **Dependencies**: Tasks 3.1-3.7
- **Description**: Cross-check every service secret against spec. Verify no duplicate values across services.
- **Acceptance Criteria**:
  - [x] Each of 10 services has exactly 14 secret files
  - [x] No two services share any unseal hex values
  - [x] No two services share postgres credentials
  - [x] pki-ca values completely different from sm-kms

### Phase 4: Fix Product-Level and Suite-Level Secret Values

**Phase Objective**: Fix all product and suite secret values to match spec patterns.

#### Task 4.1: Fix Product Unseal Secrets (5 products)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: 10m
- **Dependencies**: Phase 1
- **Description**: Replace generic `dev-unseal-key-N-of-5` with `{PRODUCT}-unseal-key-N-of-5-{hex}` for jose, pki, skeleton, sm. Normalize identity format.
- **Acceptance Criteria**:
  - [x] jose: `jose-unseal-key-N-of-5-{hex}` (was `dev-unseal-key-N-of-5`)
  - [x] pki: `pki-unseal-key-N-of-5-{hex}` (was `dev-unseal-key-N-of-5`)
  - [x] skeleton: `skeleton-unseal-key-N-of-5-{hex}` (was `dev-unseal-key-N-of-5`)
  - [x] sm: `sm-unseal-key-N-of-5-{hex}` (was `dev-unseal-key-N-of-5`)
  - [x] identity: normalized to `identity-unseal-key-N-of-5-{hex}` format
  - [x] All hex values unique per product
- **Files**:
  - `deployments/{identity,jose,pki,skeleton,sm}/secrets/unseal-*of5.secret` (25 files)

#### Task 4.2: Fix Product Postgres Secrets (5 products)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: 10m
- **Dependencies**: None
- **Description**: Verify/fix product-level postgres secrets match `{PRODUCT}_database`, `{PRODUCT}_database_user`, `{PRODUCT}_database_pass-{base64}` patterns.
- **Acceptance Criteria**:
  - [x] Each product's postgres-database = `{PRODUCT}_database`
  - [x] Each product's postgres-username = `{PRODUCT}_database_user`
  - [x] Each product's postgres-url uses corrected values
  - [x] Each product's postgres-password matches pattern
- **Files**:
  - `deployments/{identity,jose,pki,skeleton,sm}/secrets/postgres-*.secret` (20 files)

#### Task 4.3: Fix Product Hash Pepper Secrets (5 products)

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: 5m
- **Dependencies**: None
- **Description**: Verify/fix product-level hash-pepper-v3.secret matches `{PRODUCT}-hash-pepper-v3-{base64}` pattern.
- **Acceptance Criteria**:
  - [x] All 5 products match pattern
- **Files**:
  - `deployments/{identity,jose,pki,skeleton,sm}/secrets/hash-pepper-v3.secret`

#### Task 4.4: Fix Suite Secrets

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: 10m
- **Dependencies**: Phase 1
- **Description**: Fix cryptoutil-suite secrets to use `cryptoutil-` prefix (not `suite-`). Fix unseal values to `cryptoutil-unseal-key-N-of-5-{hex}`. Fix postgres values to use `cryptoutil_database` pattern.
- **Acceptance Criteria**:
  - [x] Unseal prefix: `cryptoutil-` (was `suite-`)
  - [x] postgres-database: `cryptoutil_database`
  - [x] postgres-username: `cryptoutil_database_user`
  - [x] postgres-url: correct references
  - [x] hash-pepper-v3: `cryptoutil-hash-pepper-v3-{base64}`
- **Files**:
  - `deployments/cryptoutil-suite/secrets/*.secret` (10 files)

#### Task 4.5: Verify All Product/Suite Secrets

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: 5m
- **Dependencies**: Tasks 4.1-4.4
- **Description**: Cross-check all product and suite secrets against spec.
- **Acceptance Criteria**:
  - [x] No generic `dev-unseal-key-N-of-5` values remain anywhere
  - [x] No `suite-` prefix anywhere (should be `cryptoutil-`)
  - [x] All products have unique unseal hex values

### Phase 5: Restructure Config Directories — Flat Pattern (Decision 2=B)

**Phase Objective**: Restructure ALL config directories from nested `configs/{PRODUCT}/{SERVICE}/` to flat `configs/{PS-ID}/` per Decision 2=B. This phase also absorbs the old Phase 6 (config file naming) since files are renamed during the move.

#### Task 5.1: Move sm-im Configs to Flat Structure

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 1
- **Description**: Move `configs/sm/im/*` to `configs/sm-im/`. Rename `im.yml` to `sm-im.yml`. Delete deployment variant configs (`sm-im-pg-*.yml`, `sm-im-sqlite.yml`) during move.
- **Acceptance Criteria**:
  - [x] `configs/sm-im/sm-im.yml` exists
  - [x] `configs/sm/im/` does not exist
  - [x] No deployment variant configs carried forward
  - [x] All compose/code references updated

#### Task 5.2: Move sm-kms Configs to Flat Structure

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 1
- **Description**: Move `configs/sm/kms/*` to `configs/sm-kms/`. Rename to `sm-kms.yml`. Delete deployment variant configs during move.
- **Acceptance Criteria**:
  - [x] `configs/sm-kms/sm-kms.yml` exists
  - [x] `configs/sm/kms/` does not exist
  - [x] No deployment variant configs carried forward

#### Task 5.3: Move jose-ja Configs to Flat Structure

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 1
- **Description**: Move `configs/jose/ja/*` to `configs/jose-ja/`. Rename `jose-ja-server.yml` to `jose-ja.yml` (fixes RC-4 naming).
- **Acceptance Criteria**:
  - [x] `configs/jose-ja/jose-ja.yml` exists (not `jose-ja-server.yml`)
  - [x] `configs/jose/ja/` does not exist

#### Task 5.4: Move pki-ca Configs to Flat Structure

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 1, Decision 3=B
- **Description**: Move `configs/pki/ca/*` to `configs/pki-ca/`. Rename `pki-ca-server.yml` to `pki-ca.yml` (fixes RC-4). Keep `profiles/` subdir per Decision 3=B. Delete `pki-ca-config-schema.yaml` (schema hardcoded in Go).
- **Acceptance Criteria**:
  - [x] `configs/pki-ca/pki-ca.yml` exists (not `pki-ca-server.yml`)
  - [x] `configs/pki-ca/profiles/` exists with all 24 YAML files
  - [x] `pki-ca-config-schema.yaml` deleted
  - [x] `configs/pki/ca/` does not exist

#### Task 5.5: Move skeleton-template Configs to Flat Structure

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phase 1
- **Description**: Move `configs/skeleton/template/*` to `configs/skeleton-template/`. Rename `skeleton-template-server.yml` to `skeleton-template.yml` (fixes RC-4). Delete `configs/skeleton/skeleton-server.yml` (orphaned).
- **Acceptance Criteria**:
  - [x] `configs/skeleton-template/skeleton-template.yml` exists (not `skeleton-template-server.yml`)
  - [x] `configs/skeleton/template/` does not exist
  - [x] `configs/skeleton/skeleton-server.yml` deleted

#### Task 5.6: Move All 5 Identity Service Configs to Flat Structure

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: -
- **Dependencies**: Phase 1, Decision 4=A
- **Description**: Move `configs/identity/{authz,idp,rp,rs,spa}/*` to `configs/identity-{authz,idp,rp,rs,spa}/`. Move `configs/identity/policies/` to `configs/identity-authz/domain/policies/` per Decision 4=A. Rename `adaptive-auth.yml` to `adaptive-authorization.yml` (terminology fix).
- **Acceptance Criteria**:
  - [x] `configs/identity-authz/` exists with authz service config
  - [x] `configs/identity-idp/` exists with idp service config
  - [x] `configs/identity-rp/` exists with rp service config
  - [x] `configs/identity-rs/` exists with rs service config
  - [x] `configs/identity-spa/` exists with spa service config
  - [x] `configs/identity-authz/domain/policies/` exists with 3 authorization policy files
  - [x] `adaptive-authorization.yml` (not `adaptive-auth.yml`) in policies dir
  - [x] `configs/identity/{authz,idp,rp,rs,spa}/` directories do not exist
  - [x] `configs/identity/policies/` does not exist

#### Task 5.7: Delete Empty Parent Directories

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 5m
- **Actual**: -
- **Dependencies**: Tasks 5.1-5.6
- **Description**: Delete empty `configs/{PRODUCT}/{SERVICE}/` and `configs/{PRODUCT}/` directories after all moves complete.
- **Acceptance Criteria**:
  - [x] No empty `configs/sm/`, `configs/jose/`, `configs/pki/`, `configs/skeleton/`, `configs/identity/` directories remain

#### Task 5.8: Rewrite configs_naming Fitness Linter

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Actual**: -
- **Dependencies**: Tasks 5.1-5.7
- **Description**: Rewrite `internal/apps/tools/cicd_lint/lint_fitness/configs_naming/configs_naming.go` to validate flat `configs/{PS-ID}/` pattern instead of nested `configs/{PRODUCT}/{SERVICE}/`. Update tests.
- **Acceptance Criteria**:
  - [x] Linter validates `configs/{PS-ID}/` directories
  - [x] Linter rejects nested `configs/{PRODUCT}/{SERVICE}/` pattern
  - [x] All 10 PS-ID directories validated
  - [x] `configs/pki-ca/profiles/` exception handled
  - [x] `configs/identity-authz/domain/policies/` validated
  - [x] Tests updated and passing with >=98% coverage
  - [x] `go run ./cmd/cicd-lint lint-fitness` passes

#### Task 5.9: Update configs_deployments_consistency Linter

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: -
- **Dependencies**: Task 5.8
- **Description**: Update `configs_deployments_consistency` fitness linter to work with flat config directory structure.
- **Acceptance Criteria**:
  - [x] Linter correctly maps `configs/{PS-ID}/` to `deployments/{PS-ID}/`
  - [x] Tests updated and passing

#### Task 5.10: Update All Compose File Config Paths

- **Status**: ✅ Complete (pre-satisfied)
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: -
- **Dependencies**: Tasks 5.1-5.7
- **Description**: Update all compose files across `deployments/` to reference new flat config paths instead of nested paths.
- **Acceptance Criteria**:
  - [x] No compose file references `configs/{PRODUCT}/{SERVICE}/` paths
  - [x] All config volume mounts use `configs/{PS-ID}/` paths
  - [x] `docker compose config` validates for all services

#### Task 5.11: Update All Go Code Config Path References

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: -
- **Dependencies**: Tasks 5.1-5.7
- **Description**: Search all Go source files for references to old nested config paths and update.
- **Acceptance Criteria**:
  - [x] Zero Go code references to `configs/{PRODUCT}/{SERVICE}/` paths
  - [x] `go build ./...` passes
  - [x] `golangci-lint run` passes

#### Task 5.12: Verify Flat Config Structure Complete

- **Status**: ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: Tasks 5.1-5.11
- **Description**: Verify the entire configs/ directory matches the flat structure. List all directories under configs/ and confirm each is a valid PS-ID.
- **Acceptance Criteria**:
  - [x] Only 10 PS-ID directories plus any documented exceptions exist under configs/
  - [x] No nested `configs/{PRODUCT}/{SERVICE}/` remains
  - [x] Fitness linters pass
  - [x] Deployment validators pass

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

#### Task 7.3: Fix Suite Dockerfile OCI Labels

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: None
- **Description**: Suite Dockerfile currently has incorrect labels ("CA Server" / "Certificate Authority REST API Server"). Update to match suite identity (e.g., "cryptoutil Suite" / "CryptoUtil Suite Server").
- **Acceptance Criteria**:
  - [ ] Suite Dockerfile `org.opencontainers.image.title` matches suite identity
  - [ ] Suite Dockerfile `org.opencontainers.image.description` matches suite identity
  - [ ] Labels do not reference any single PS-ID

#### Task 7.4: Verify No Orphans Remain

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

**Phase Objective**: Implement 3 new fitness linters and verify full compliance.

#### Task 8.1: Implement unseal-secret-content Linter

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Actual**: -
- **Dependencies**: Phases 3-4 (secret values corrected)
- **Description**: Create `internal/apps/tools/cicd_lint/lint_fitness/unseal_secret_content/unseal_secret_content.go`. Validates: correct PS-ID/PRODUCT/SUITE prefix, unique hex per shard, correct `N-of-5` infix, tier prefix matches deployment tier. Tests with >=98% coverage.
- **Acceptance Criteria**:
  - [ ] Linter validates unseal values for all 3 tiers (service, product, suite)
  - [ ] Detects wrong prefix (e.g., `kms-` in pki-ca)
  - [ ] Detects duplicate hex across services
  - [ ] Detects generic `dev-unseal-key-N-of-5` values
  - [ ] Tests pass with >=98% coverage
  - [ ] Registered in fitness linter catalog

#### Task 8.2: Implement dockerfile-labels Linter

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: -
- **Dependencies**: Task 7.3 (suite Dockerfile fixed)
- **Description**: Create `internal/apps/tools/cicd_lint/lint_fitness/dockerfile_labels/dockerfile_labels.go`. Validates OCI labels in all Dockerfiles: correct PS-ID in title, correct description, version label present. Tests with >=98% coverage.
- **Acceptance Criteria**:
  - [ ] Linter validates SERVICE, PRODUCT, SUITE Dockerfiles
  - [ ] Detects mismatched PS-ID in labels
  - [ ] Detects missing required labels
  - [ ] Tests pass with >=98% coverage
  - [ ] Registered in fitness linter catalog

#### Task 8.3: Implement secret-naming Linter

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: -
- **Dependencies**: Phase 3 (secret filenames normalized)
- **Description**: Create `internal/apps/tools/cicd_lint/lint_fitness/secret_naming/secret_naming.go`. Validates: all secret filenames use hyphens (no underscores), correct count per tier (14 service, 10+ product, 10+ suite), correct `.secret` extension. Tests with >=98% coverage.
- **Acceptance Criteria**:
  - [ ] Linter detects underscore filenames (e.g., `unseal_1of5.secret`)
  - [ ] Linter validates file count per deployment tier
  - [ ] Tests pass with >=98% coverage
  - [ ] Registered in fitness linter catalog

#### Task 8.4: Run Fitness Linters

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Tasks 8.1-8.3, Phases 1-7
- **Description**: Run `go run ./cmd/cicd-lint lint-fitness` and fix any violations.
- **Acceptance Criteria**:
  - [ ] Zero violations from lint-fitness (including 3 new linters)

#### Task 8.5: Run Deployment Validators

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Phases 1-7
- **Description**: Run `go run ./cmd/cicd-lint lint-deployments` and fix any violations.
- **Acceptance Criteria**:
  - [ ] Zero violations from lint-deployments

#### Task 8.6: Run Build and Lint

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: -
- **Dependencies**: Phases 1-7
- **Description**: Verify `go build ./...`, `go build -tags e2e,integration ./...`, `golangci-lint run`, `golangci-lint run --build-tags e2e,integration` all pass.
- **Acceptance Criteria**:
  - [ ] Build clean (both standard and tagged)
  - [ ] Lint clean (both standard and tagged)

#### Task 8.7: Run Compose Validation

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

### Phase 10: Migrate internal/apps/ to Flat PS-ID Structure

**Phase Objective**: Migrate all 10 service directories from nested `internal/apps/{PRODUCT}/{SERVICE}/` to flat `internal/apps/{PS-ID}/`, matching the `cmd/{PS-ID}/` and `deployments/{PS-ID}/` conventions already in use. Update the `service_structure` fitness linter accordingly.

#### Task 10.1: Move Service Directories to Flat PS-ID Structure

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Actual**: -
- **Dependencies**: Phases 1-9 complete (all config/secret changes done first)
- **Description**: Move all 10 service directories from their nested locations to flat PS-ID directories under `internal/apps/`. Use `git mv` to preserve history.
- **Acceptance Criteria**:
  - [ ] `internal/apps/sm-im/` exists (was `internal/apps/sm/im/`)
  - [ ] `internal/apps/sm-kms/` exists (was `internal/apps/sm/kms/`)
  - [ ] `internal/apps/jose-ja/` exists (was `internal/apps/jose/ja/`)
  - [ ] `internal/apps/pki-ca/` exists (was `internal/apps/pki/ca/`)
  - [ ] `internal/apps/skeleton-template/` exists (was `internal/apps/skeleton/template/`)
  - [ ] `internal/apps/identity-authz/` exists (was `internal/apps/identity/authz/`)
  - [ ] `internal/apps/identity-idp/` exists (was `internal/apps/identity/idp/`)
  - [ ] `internal/apps/identity-rp/` exists (was `internal/apps/identity/rp/`)
  - [ ] `internal/apps/identity-rs/` exists (was `internal/apps/identity/rs/`)
  - [ ] `internal/apps/identity-spa/` exists (was `internal/apps/identity/spa/`)
  - [ ] No nested `internal/apps/{PRODUCT}/{SERVICE}/` directories remain (only `{PRODUCT}.go` and shared packages remain in product dirs)
- **Files**:
  - `internal/apps/sm/im/` → `internal/apps/sm-im/`
  - `internal/apps/sm/kms/` → `internal/apps/sm-kms/`
  - `internal/apps/jose/ja/` → `internal/apps/jose-ja/`
  - `internal/apps/pki/ca/` → `internal/apps/pki-ca/`
  - `internal/apps/skeleton/template/` → `internal/apps/skeleton-template/`
  - `internal/apps/identity/{authz,idp,rp,rs,spa}/` → `internal/apps/identity-{authz,idp,rp,rs,spa}/`

#### Task 10.2: Update All Go Import Paths

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: -
- **Dependencies**: Task 10.1
- **Description**: Update every Go import path referencing the old nested service dirs. This includes `cmd/{PS-ID}/main.go` imports, cross-service imports, and any test files.
- **Acceptance Criteria**:
  - [ ] Zero references to `internal/apps/sm/im`, `internal/apps/sm/kms`, `internal/apps/jose/ja`, `internal/apps/pki/ca`, `internal/apps/skeleton/template`, `internal/apps/identity/authz`, `internal/apps/identity/idp`, `internal/apps/identity/rp`, `internal/apps/identity/rs`, `internal/apps/identity/spa`
  - [ ] `go build ./...` passes
  - [ ] `go build -tags e2e,integration ./...` passes

#### Task 10.3: Update service_structure Fitness Linter

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Actual**: -
- **Dependencies**: Task 10.1
- **Description**: Rewrite `internal/apps/tools/cicd_lint/lint_fitness/service_structure/service_structure.go` to validate flat `internal/apps/{PS-ID}/` pattern instead of nested `filepath.Join(appsDir, svc.Product, svc.Service)`. Add `PSID` field to the service struct; remove `Product`/`Service` path construction. Update tests with >=98% coverage.
- **Acceptance Criteria**:
  - [ ] `knownServices` slice uses PS-ID strings instead of `Product`/`Service` field pairs
  - [ ] `serviceDir` computed as `filepath.Join(appsDir, svc.PSID)`
  - [ ] Linter accepts flat `internal/apps/{PS-ID}/` directories
  - [ ] Linter rejects nested `internal/apps/{PRODUCT}/{SERVICE}/` pattern
  - [ ] Tests updated and passing with >=98% coverage
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/service_structure/service_structure.go`
  - `internal/apps/tools/cicd_lint/lint_fitness/service_structure/service_structure_test.go`

#### Task 10.4: Build, Test, and Lint Verification

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: -
- **Dependencies**: Tasks 10.1-10.3
- **Description**: Full build, test, and lint verification after the structural migration.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `go test ./...` passes (100%, zero skips)
  - [ ] `golangci-lint run` clean
  - [ ] `golangci-lint run --build-tags e2e,integration` clean
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes

---

### Phase 11: Knowledge Propagation

**Phase Objective**: Apply lessons learned to permanent artifacts.

#### Task 11.1: Review Lessons and Update ARCHITECTURE.md

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: -
- **Dependencies**: All prior phases
- **Description**: Review lessons.md, update ARCHITECTURE.md with new patterns including: flat config structure, deployment secret naming with PS-ID prefix specificity, Dockerfile parameterization (ARGs/LABELs/multi-stage per tier), new fitness linter catalog entries (unseal-secret-content, dockerfile-labels, secret-naming), and terminology enforcement. ARCHITECTURE.md must have same specificity as target-structure.md for deployment patterns.
- **Acceptance Criteria**:
  - [ ] lessons.md reviewed
  - [ ] ARCHITECTURE.md Section 12 updated: flat config pattern, secret naming patterns with PS-ID prefix specificity
  - [ ] ARCHITECTURE.md Section 12 updated: Dockerfile parameterization patterns (ARGs, LABELs, multi-stage structure per tier)
  - [ ] ARCHITECTURE.md Section 9.11 updated: 3 new fitness linter catalog entries
  - [ ] Propagation check passes

#### Task 11.2: Update Instruction Files

- **Status**: Not Started
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: -
- **Dependencies**: Task 11.1
- **Description**: Update instruction files to match ARCHITECTURE.md specificity. Key files: 02-01.architecture.instructions.md (flat config pattern, Dockerfile patterns), 04-01.deployment.instructions.md (secret naming, Docker patterns), 02-05.security.instructions.md (unseal naming patterns).
- **Acceptance Criteria**:
  - [ ] Instructions reflect flat `configs/{PS-ID}/` pattern
  - [ ] Instructions document Dockerfile parameterization
  - [ ] Instructions document secret value tier-prefix patterns
  - [ ] Terminology enforcement guidance strengthened
  - [ ] Propagation check passes (`go run ./cmd/cicd-lint lint-docs validate-propagation`)

#### Task 11.3: Update target-structure.md with Implementation Learnings

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
- Phase 1 expanded: 9 additional target-structure.md fixes completed (commit 286ea4588), plus 1 new verification task (1.13)
- Phase 7 expanded: Task 7.3 added for suite Dockerfile OCI label fix
- Phase 8 expanded: Tasks 8.1-8.3 added for 3 new fitness linters (unseal-secret-content, dockerfile-labels, secret-naming)
- Phase 10 expanded: Tasks 10.1-10.2 scope increased for ARCHITECTURE.md and instruction file specificity matching target-structure.md
- Total phases: 11. Total tasks: 63 (was 59; added 4 new Phase 10 tasks for internal/apps migration; renamed old Phase 10 Knowledge Propagation → Phase 11; renumbered old Tasks 10.1-10.3 → 11.1-11.3)
- E2E testing may be blocked if Docker Desktop is not running

---

## Evidence Archive

- `test-output/framework-v6/phase0-research/` - Pre-plan research and analysis
