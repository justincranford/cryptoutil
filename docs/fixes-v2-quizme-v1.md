# Deployment & Configuration Reorganization - Questions for Executive Decision

**Date**: 2026-02-15
**Context**: Reorganizing deployment structure for SUITE/PRODUCT/SERVICE-level secrets and config file naming

---

## 1. Secrets Naming Suffix Strategy

**Context**: User specified three secret levels with different suffixes:
- SUITE-level: `-SUITEONLY.secret` (in `deployments/cryptoutil/secrets/`)
- PRODUCT-level: `-PRODUCTONLY.secret` (in `deployments/PRODUCT/secrets/`)
- SERVICE-level: `-SERVICEONLY.secret` (in `deployments/PRODUCT-SERVICE/secrets/`)

**Question**: Should there be a fourth level for secrets shared across multiple levels, or are all secrets exclusive to one level?

**Options**:

**A**. Three exclusive levels only (SUITEONLY, PRODUCTONLY, SERVICEONLY) - each secret exists at exactly one level

**B**. Add shared secret level: `-SHARED.secret` for secrets used across multiple deployments

**C**. Add fallback pattern: SERVICE checks SERVICEONLY first, then PRODUCTONLY, then SUITEONLY (cascade lookup)

**D**. Add environment-specific suffixes: `-SUITEONLY-prod.secret`, `-SUITEONLY-dev.secret` for multi-environment support

**E**. Other (please specify):

**Answer:** B

---

## 2. Config File Naming Pattern Enforcement

**Context**: User specified config files MUST start with `PRODUCT-SERVICE-` prefix and end with `-#` like:
- `sm-kms-app-postgresql-1.yml`
- `sm-kms-app-sqlite-1.yml`

**Current State**: Many files don't follow this pattern:
- `sm-kms/config/kms-common.yml` (missing `sm-` prefix and `-app-`)
- `pki-ca/config/ca-sqlite.yml` (missing `pki-` prefix and `-app-` and `-1`)
- `jose-ja/config/ja-sqlite.yml` (missing `jose-` prefix and `-app-` and `-1`)

**Question**: What is the complete required pattern for ALL config file types?

**Options**:

**A**. All configs: `{PRODUCT}-{SERVICE}-app-{db-type}-{instance}.yml`
Example: `sm-kms-app-sqlite-1.yml`, `sm-kms-app-postgresql-2.yml`

**B**. Instance configs + common: `{PRODUCT}-{SERVICE}-app-{db-type}-{instance}.yml` AND `{PRODUCT}-{SERVICE}-app-common.yml`
Example: `sm-kms-app-common.yml`, `sm-kms-app-sqlite-1.yml`

**C**. Instance configs + special files: Same as B but allow `{PRODUCT}-{SERVICE}-<type>.yml` for demo/e2e
Example: `sm-kms-app-common.yml`, `sm-kms-e2e.yml`, `sm-kms-demo.yml`

**D**. Strict prefix only: ALL files must start with `{PRODUCT}-{SERVICE}-` but suffix flexible
Example: `sm-kms-anything.yml` allowed, `kms-anything.yml` forbidden

**E**. Other (please specify):

**Answer:** C

---

## 3. demo-seed.yml and integration.yml Usage

**Context**: User stated:
- Integration tests use TestMain (no compose)
- Demo runs on top of E2E (docker compose with demo data)
- But `sm-kms/config/demo-seed.yml` and `integration.yml` exist

**Question**: What should happen to these files?

**Options**:

**A**. Delete both - integration tests use TestMain, demo data is injected into E2E services, no separate config needed

**B**. Rename to match pattern - `sm-kms-demo-seed.yml` and `sm-kms-integration.yml` for backwards compatibility

**C**. Move to different directory - `deployments/sm-kms/demo/demo-seed.yml` and `deployments/sm-kms/integration/integration.yml`

**D**. Keep demo-seed for E2E data injection, delete integration.yml

**E**. Other (please specify):

**Answer:** D, but use naming pattern D from question 2: `sm-kms-e2e.yml`, `sm-kms-demo.yml`

---

## 4. Missing Config Files for Services

**Context**: Some services missing expected config files:
- **jose-ja**: has `ja-common.yml`, `ja-sqlite.yml` but missing `jose-ja-app-sqlite-1.yml`, `jose-ja-app-postgres-1.yml`, `jose-ja-app-postgres-2.yml`
- **cipher-im**: config/ directory is empty
- **identity services**: have mixed patterns (`authz-demo.yml`, `idp-e2e.yml`) not following `identity-authz-app-*` pattern

**Question**: Should all services have the same complete set of config files?

**Options**:

**A**. Standard set for all services:
- `{P}-{S}-app-common.yml`
- `{P}-{S}-app-sqlite-1.yml`
- `{P}-{S}-app-postgresql-1.yml`
- `{P}-{S}-app-postgresql-2.yml`

**B**. Standard + optional demo/e2e:
- Same as A plus optional `{P}-{S}-e2e.yml`, `{P}-{S}-demo.yml`

**C**. Minimal required + service-specific:
- Only `{P}-{S}-app-common.yml` required, others optional based on service needs

**D**. Template-based generation:
- Generate from template with placeholders, services add service-specific overrides

**E**. Other (please specify):

**Answer:** B

---

## 5. CICD Linter Validation Scope

**Context**: Need to enhance `lint_deployments` to validate config filenames and detect unexpected files.

**Question**: What level of strictness should the linter enforce?

**Options**:

**A**. Strict allowlist - Only files matching exact patterns allowed, everything else flagged as error

**B**. Pattern enforcement - All files must match naming pattern, but allow any valid pattern

**C**. Required + optional - Require specific files (common, sqlite-1), allow additional files matching pattern

**D**. Warning mode - Flag non-conformant files as warnings (not errors) during transition period

**E**. Other (please specify):

**Answer:** D, i need to review all of the warnings and decide which ones should be errors after transition period; the transaction period must be short, i want to transition to strict enforcement as soon as possible to avoid long-term confusion and technical debt

---

## 6. PRODUCT-Level vs SUITE-Level Deployments

**Context**: User specified three deployment levels:
- SUITE: `deployments/cryptoutil/` (all 9 services)
- PRODUCT: `deployments/{PRODUCT}/` (1-5 services per product)
- SERVICE: `deployments/{PRODUCT}-{SERVICE}/` (single service)

**Question**: Do PRODUCT and SUITE-level directories need compose.yml files, or just SERVICE-level?

**Options**:

**A**. All three levels have compose.yml - SUITE aggregates all products, PRODUCT aggregates services

**B**. Only SERVICE-level has compose.yml - SUITE/PRODUCT are organizational only

**C**. SERVICE + PRODUCT have compose.yml - SUITE references PRODUCT compose files

**D**. Current state is correct - reevaluate after examining existing compose files

**E**. Other (please specify): I don't know. I hope one compose.yml per service is sufficient, but probably not. you need to do deep analysis to see how to achieve the goals. Potentially, you might need per-suite and per-product and per-service deployment compose files for the secrets and per-suite/product/service yaml config files, and compose templates per service, and then use composition to inject/import them into a file usable compose for each situation? You may need to run experiments to work out all of the patterns, but in a fast way with alpine images, instead of the actual 9 services / 5 products / 1 suite, and do a writeup of the patterns (with evidence) in docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md.

**Answer:** E

---

## 7. Instruction File → ARCHITECTURE.md Link Granularity

**Context**: User wants "for every section in every copilot instructions file, there MUST be a markdown link reference to the relevant section(s) in ARCHITECTURE.md"

**Question**: What granularity of linking is required?

**Options**:

**A**. Every heading (##, ###, ####) gets a link - Maximum coverage, may be repetitive

**B**. Every top-level section (## only) gets links - Main sections only

**C**. Every section with substantive content - Skip headers that are just organizational

**D**. Link to most specific matching section - One link per concept to best ARCHITECTURE.md match

**E**. Other (please specify): Hybrid. Ideally docs/ARCHITECTURE.md format can be structured in a way that large chunks of anything can be copied verbatim with bidirectional links. That will make it easier to scale up Copilot's requirement for content sprawl into Copilot Instructions, Custom Agents, Custom Prompts, Custom Skills, AGENTS.md, and any and all other current and future files that Copilot needs. It's annoying the amount of duplication required by different Copilot things, so I need ARCHITECTURE.md to be the absolute single source of truth, and have it strategically structured to copy chunks verbatim with bidirectional links so i can approach updates in ARCHITECTURE.md first, and have Copilot propagate those verbatim updates to the other files.

**Answer:** E

---

## 8. ARCHITECTURE.md Documentation Priority

**Context**: Multiple documentation gaps identified:
1. Secrets three-level structure (SUITE/PRODUCT/SERVICE)
2. Config file naming patterns
3. demo-seed.yml / integration.yml usage
4. TestMain vs compose for integration tests

**Question**: Should all of these be documented in ARCHITECTURE.md before implementation?

**Options**:

**A**. Document all before implementation - Update ARCHITECTURE.md Section 12.4, then implement

**B**. Document incrementally - Add sections as each component is implemented

**C**. Minimal documentation - Add high-level overview, detailed rules in linter code comments

**D**. Post-implementation documentation - Implement first, document patterns afterward

**E**. Other (please specify):

**Answer:** A; ARCHITECTURE.md is the single source of truth; i want to ALWAYS be able to strategically apply updates to ARCHITECTURE.md first, then get Copilot to use the bidirectional links from ARCHITECTURE.md to strategically and autonomously propagate updates to all of the other Copilot files (Copilot Instructions, Custom Agents, Custom Prompts, Custom Skills, AGENTS.md, etc)

---

## 9. Backward Compatibility During Transition

**Context**: Many existing config files don't match new pattern. Services may be in use.

**Question**: How should we handle existing non-conformant files during transition?

**Options**:

**A**. Break immediately - Rename all files now, update all references

**B**. Deprecation period - Keep old files, add new files, mark old as deprecated

**C**. Symlinks - Create symlinks from new names to old files temporarily

**D**. Configuration mapper - Code accepts both old and new patterns during transition

**E**. Other (please specify):

**Answer:** A; no files are in use, this is pre-production repository, nothing has ever been deployed; it is too hard for me as just 1 developer to manually scale up the number of services and products, i need automation like the cicd linter for ./configs/ and ./deloyments/ to enforce rigid and repeatable directory structures, filename patterns, directory contents, file contents, etc. i need that to happen as soon as possible.

---

## 10. Validation Test Coverage Requirements

**Context**: User wants "happy path and sad path tests" for deployment linter enhancements.

**Question**: What sad path scenarios should be explicitly tested?

**Options**:

**A**. Comprehensive coverage:
- Wrong prefix (kms-app.yml instead of sm-kms-app.yml)
- Wrong suffix (sm-kms-app-sqlite.yml instead of sm-kms-app-sqlite-1.yml)
- Unexpected files (demo-seed.yml, integration.yml)
- Missing required files

**B**. Pattern violations only:
- Files not matching `{PRODUCT}-{SERVICE}-*` pattern
- Required files missing

**C**. Critical violations:
- Wrong PRODUCT prefix (pki-kms instead of sm-kms)
- Completely wrong filename format

**D**. User-specified errors only - Add tests as real errors are discovered

**E**. Other (please specify):

**Answer:** A

---

## Summary

Total questions: 10
Critical decisions (block implementation): Questions 1, 2, 3, 4
Nice-to-have clarifications: Questions 5, 6, 7, 8, 9, 10

**Recommended Answer Path**:
1. Answer Q1-Q4 first (secrets, config naming, demo files, missing files)
2. Answer Q5-Q7 next (linter scope, deployment levels, link granularity)
3. Answer Q8-Q10 last (documentation timing, backward compat, test coverage)
