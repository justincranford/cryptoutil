# Framework v4 - Anti-Drift Fitness Linter Expansion Plan

**Status**: TODO — Planning complete, no phases started.
**Created**: 2026-07-20
**Depends On**: `docs/framework-v3/` (complete), `internal/apps/cicd/lint_fitness/` (30 existing checks)
**Purpose**: Rigidly parameterize the location, naming, content, and structure of every file across all products (5), product-services (10), suites (1), and all helpers (cicd, etc). Create comprehensive anti-drift fitness linters that detect and reject any structural or naming drift across the entire repository.

**Guiding Principle**: Drift today = production incident tomorrow. Every naming convention, every file location, every config key, every compose service name, every magic constant must be machine-verifiable. The human should never be the last line of defense against drift.

---

## Companion Documents

1. **plan.md** (this file) - phases, objectives, decisions
2. **tasks.md** - task checklist per phase
3. **lessons.md** - persistent memory: what worked, what did not, root causes, patterns

---

## Context: Current State of lint-fitness

### What Framework v3 Delivered (30 checks)

Architecture-focused fitness checks covering code structure, security, and test quality:

**Migrated from lint_go (10)**: `cgo-free-sqlite`, `circular-deps`, `cmd-main-pattern`, `crypto-rand`, `insecure-skip-verify`, `migration-numbering`, `non-fips-algorithms`, `product-structure`, `product-wiring`, `service-structure`

**Migrated from lint_gotest (5)**: `bind-address-safety`, `no-hardcoded-passwords`, `parallel-tests`, `test-patterns`, `check-skeleton-placeholders`

**New checks (Phase 4-11 in v3, 15 more)**: `cross-service-import-isolation`, `domain-layer-isolation`, `file-size-limits`, `health-endpoint-presence`, `tls-minimum-version`, `admin-bind-address`, `service-contract-compliance`, `migration-range-compliance`, `no-local-closed-db-helper`, `no-postgres-in-non-e2e`, `no-unit-test-real-db`, `no-unit-test-real-server`, `gen-config-initialisms`, `require-api-dir`, `require-framework-naming`

### What v3 Did NOT Cover (Naming & Structural Drift)

The 30 existing checks are **code-quality focused**. None of them validate:

1. **Product/service name drift** — Old product names like "Cipher" or "cipher-im" cannot be detected. Already happened: CI/CD passed with "Cipher IM" references for months.
2. **Config file presence** — Missing `config-pg-1.yml` or `config-pg-2.yml` for a service silently fails at runtime.
3. **Compose service name patterns** — `sm-im-app-sqlite-1` vs `sm-im-sqlite` (without `-1`) goes undetected.
4. **OTLP service name conventions** — `otlp-service: sm-im-pg-1` vs correct `sm-im-postgres-1` (or `sm-kms-sqlite-1` in standalone).
5. **Magic constant alignment** — `IME2ESQLiteContainer = "sm-im-app-sqlite-1"` must match the actual compose service name.
6. **Deployment directory completeness** — Every product-service must have Dockerfile, compose.yml, secrets/, config/ with all 4 config files.
7. **Legacy directory detection** — `internal/apps/cipher/` should not exist (old product name).
8. **Comment header correctness** — Migration files saying "Cipher IM database schema" pass all 30 checks undetected.

---

## Entity Registry (Canonical Single Source of Truth)

This plan establishes the following as the canonical registry. Any drift from these tables is a violation.

### Products (5)

| Product ID | Display Name | internal/apps/ dir | cmd/ dir |
|------------|-------------|-------------------|---------|
| `identity` | Identity | `identity/` | `identity/` |
| `jose` | JOSE | `jose/` | `jose/` |
| `pki` | PKI | `pki/` | `pki/` |
| `skeleton` | Skeleton | `skeleton/` | `skeleton/` |
| `sm` | Secrets Manager | `sm/` | `sm/` |

### Product-Services (10)

| PS ID | Product | Service | Display Name | internal/apps/ dir | Port Range | PostgreSQL Port |
|-------|---------|---------|-------------|-------------------|-----------|----------------|
| `identity-authz` | identity | authz | Identity Authorization Server | `identity/authz/` | 8200-8299 | 54324 |
| `identity-idp` | identity | idp | Identity Provider | `identity/idp/` | 8300-8399 | 54325 |
| `identity-rp` | identity | rp | Identity Relying Party | `identity/rp/` | 8500-8599 | 54327 |
| `identity-rs` | identity | rs | Identity Resource Server | `identity/rs/` | 8400-8499 | 54326 |
| `identity-spa` | identity | spa | Identity Single Page App | `identity/spa/` | 8600-8699 | 54328 |
| `jose-ja` | jose | ja | JOSE JWK Authority | `jose/ja/` | 8800-8899 | 54321 |
| `pki-ca` | pki | ca | PKI Certificate Authority | `pki/ca/` | 8100-8199 | 54320 |
| `skeleton-template` | skeleton | template | Skeleton Template | `skeleton/template/` | 8900-8999 | 54329 |
| `sm-im` | sm | im | Secrets Manager Instant Messenger | `sm/im/` | 8700-8799 | 54322 |
| `sm-kms` | sm | kms | Secrets Manager Key Management | `sm/kms/` | 8000-8099 | 54323 |

### Suite (1)

| Suite ID | Display Name | cmd/ dir |
|---------|-------------|---------|
| `cryptoutil` | cryptoutil | `cryptoutil/` |

---

## Naming Convention Catalog

This section is the **complete parameterization specification** for every file type across the repo. Each entry defines the exact formula for constructing names.

### Go Code Naming

| Item | Pattern | Example |
|------|---------|---------|
| Product dir | `internal/apps/{PRODUCT}/` | `internal/apps/sm/` |
| Service dir | `internal/apps/{PRODUCT}/{SERVICE}/` | `internal/apps/sm/im/` |
| Product entry file | `{PRODUCT}.go` in product dir | `sm.go` |
| Product test file | `{PRODUCT}_test.go` in product dir | `sm_test.go` |
| Service entry file | `{SERVICE}.go` in service dir | `im.go` |
| Service usage file | `{SERVICE}_usage.go` in service dir | `im_usage.go` |
| cmd product binary | `cmd/{PRODUCT}/main.go` | `cmd/sm/main.go` |
| cmd service binary | `cmd/{PRODUCT}-{SERVICE}/main.go` | `cmd/sm-im/main.go` |
| Magic constants file | `magic_{PRODUCT}.go` or `magic_{PRODUCT}_{SERVICE}.go` | `magic_sm_im.go` |

### Deployment Naming

| Item | Pattern | Example (sm-im) | Example (jose-ja) |
|------|---------|----------------|------------------|
| Deployment dir | `deployments/{PS-ID}/` | `deployments/sm-im/` | `deployments/jose-ja/` |
| Dockerfile | `deployments/{PS-ID}/Dockerfile` | `deployments/sm-im/Dockerfile` | `deployments/jose-ja/Dockerfile` |
| Compose file | `deployments/{PS-ID}/compose.yml` | `deployments/sm-im/compose.yml` | `deployments/jose-ja/compose.yml` |
| Config dir | `deployments/{PS-ID}/config/` | same | same |
| Config common | `{PS-ID}-app-common.yml` | `sm-im-app-common.yml` | `jose-ja-app-common.yml` |
| Config sqlite-1 | `{PS-ID}-app-sqlite-1.yml` | `sm-im-app-sqlite-1.yml` | `jose-ja-app-sqlite-1.yml` |
| Config pg-1 | `{PS-ID}-app-postgresql-1.yml` | `sm-im-app-postgresql-1.yml` | `jose-ja-app-postgresql-1.yml` |
| Config pg-2 | `{PS-ID}-app-postgresql-2.yml` | `sm-im-app-postgresql-2.yml` | `jose-ja-app-postgresql-2.yml` |
| Secrets dir | `deployments/{PS-ID}/secrets/` | same | same |
| Compose header line 3 | `# {PS-ID-UPPER} Docker Compose Configuration` | `# SM-IM Docker Compose Configuration` | `# JOSE-JA Docker Compose Configuration` |
| Compose header line 5 | `# SERVICE-level deployment for {Display Name}.` | `# SERVICE-level deployment for Secrets Manager Instant Messenger.` | `# SERVICE-level deployment for JOSE JWK Authority.` |

### Compose Service Names

| Item | Pattern | Example (sm-im) |
|------|---------|----------------|
| App SQLite instance | `{PS-ID}-app-sqlite-1` | `sm-im-app-sqlite-1` |
| App Postgres instance 1 | `{PS-ID}-app-postgres-1` | `sm-im-app-postgres-1` |
| App Postgres instance 2 | `{PS-ID}-app-postgres-2` | `sm-im-app-postgres-2` |
| DB Postgres service | `{PS-ID}-db-postgres-1` | `sm-im-db-postgres-1` |
| DB container_name | `{PS-ID}-postgres` | `sm-im-postgres` |
| DB hostname | `{PS-ID}-postgres` | `sm-im-postgres` |
| Network name | `{PS-ID}-network` | `sm-im-network` |

### Standalone Config Naming (configs/{PRODUCT}/{SERVICE}/)

| Item | Pattern | Example (sm-im) | Example (sm-kms) |
|------|---------|----------------|-----------------|
| Standalone dir | `configs/{PRODUCT}/{SERVICE}/` | `configs/sm/im/` | `configs/sm/kms/` |
| SQLite config | `config-sqlite.yml` | `config-sqlite.yml` | `config-sqlite.yml` |
| PG-1 config | `config-pg-1.yml` | `config-pg-1.yml` | `config-pg-1.yml` |
| PG-2 config | `config-pg-2.yml` | `config-pg-2.yml` | `config-pg-2.yml` |
| OTLP service (standalone sqlite) | `{PS-ID}-sqlite-1` | `sm-im-sqlite-1` | `sm-kms-sqlite-1` |
| OTLP service (standalone pg-1) | `{PS-ID}-postgres-1` | `sm-im-postgres-1` | `sm-kms-postgres-1` |
| OTLP service (standalone pg-2) | `{PS-ID}-postgres-2` | `sm-im-postgres-2` | `sm-kms-postgres-2` |

**NOTE:** sm-kms standalone configs currently use `sm-kms-pg-1` / `sm-kms-pg-2` (legacy). Phase 1 fixes this inconsistency.

### Magic Constants Naming

| Item | Pattern | Example (sm-im) |
|------|---------|----------------|
| OTLP service constant | `OTLPService{PS-PASCAL}` | `OTLPServiceSMIM = "sm-im"` |
| E2E SQLite container | `{Service-PASCAL}E2ESQLiteContainer` | `IME2ESQLiteContainer = "sm-im-app-sqlite-1"` |
| E2E Postgres-1 container | `{Service-PASCAL}E2EPostgreSQL1Container` | `IME2EPostgreSQL1Container = "sm-im-app-postgres-1"` |
| E2E Postgres-2 container | `{Service-PASCAL}E2EPostgreSQL2Container` | `IME2EPostgreSQL2Container = "sm-im-app-postgres-2"` |
| E2E compose file | `{Service-PASCAL}E2EComposeFile` | `IME2EComposeFile = "../../../../../deployments/sm-im/compose.yml"` |

### Migration Comment Headers

| Item | Pattern | Example (sm-im) |
|------|---------|----------------|
| Domain migration up comment | First line: `-- {Display Name} database schema ...` | `-- SM IM database schema (SQLite + PostgreSQL compatible)` |
| Domain migration down comment | First line: `-- {Display Name} database schema rollback` | `-- SM IM database schema rollback` |

### Banned Names (Drift Indicators)

Any occurrence of the following in code/configs/migrations (excluding legitimate crypto terminology in cryptographic context) is a VIOLATION:

- `Cipher IM`, `cipher-im`, `cipher_im`, `CipherIM` (product rename drift)
- `sm-im-sqlite` without trailing `-1` (instance numbering drift)
- `sm-kms-pg-` (legacy abbreviation, should be `sm-kms-postgres-` in standalone configs)
- Legacy dir: `internal/apps/cipher/` (old product name)

---

## Guiding Decisions

### D1: All 10 Product-Services Get Full Parameterization

Every product-service (not just sm-im and sm-kms) must be covered by the new fitness checks. Identity services (authz, idp, rp, rs, spa) and pki-ca are currently under-checked.

### D2: Registry-Driven Checks (No Scattered Constants)

All checks read from a single shared registry (Go struct or map in a shared package `lint_fitness_common`). Adding a new product-service = update the registry once, all checks inherit the new entity automatically.

### D3: Standalone Config Rules Only for Services That Have Them

Currently only `sm-im` and `sm-kms` have the standardized `configs/{PRODUCT}/{SERVICE}/config-*.yml` pattern. The fitness check for standalone config naming applies ONLY to services in that pattern. Jose-ja, pki-ca, identity services use different config patterns and are out of scope for standalone config checks (Phase 2).

### D4: Zero Tolerance for Banned Names

Any file (yml, yaml, go, sql, md, txt) containing a banned drift indicator is a HARD ERROR. No exceptions. No allowlists. This prevents the "Cipher IM" recurrence.

### D5: Magic Constants Must Match Compose Service Names

`{Service}E2E*Container` constants must exactly match compose service names. A check verifying this cross-reference is MANDATORY (catches the `sm-im-sqlite` vs `sm-im-app-sqlite-1` class of bug).

### D6: Fix `sm-kms-pg-` Legacy Before Adding Checks

Before adding the `otlp-service-name-pattern` check, the pre-existing `sm-kms-pg-1` / `sm-kms-pg-2` legacy must be fixed. The check would fail immediately otherwise. Fix first, then add check.

### D7: Compose Header Is Machine-Verifiable

The compose.yml comment header lines 3 and 5 follow exact patterns. These are not documentation — they are machine-verified contracts. The check reads lines 3 and 5 of every compose.yml under `deployments/{PS-ID}/` and rejects any deviation.

### D8: No New Standalone Config Checks for Legacy Services

Services with ad-hoc config layouts (jose-ja, pki-ca, identity services) are explicitly out of scope for standalone config presence/naming checks in this plan. Standardizing those is a separate future initiative.

---

## Goals for Framework v4

1. **~25 new fitness checks** covering naming, structure, and content drift across all 16 entities
2. **Registry-driven architecture** — single Go registry drives all checks
3. **Zero manual drift** — any new product-service automatically covered after registry update
4. **Cross-reference validation** — magic constants cross-checked against compose service names
5. **Banned name detection** — old product names rejected everywhere
6. **ARCHITECTURE.md updated** — fitness catalog entry updated from 23 to ~55 checks

---

## Phases

### Phase 1: Fix Legacy `sm-kms-pg-` Naming and Add OTLP Service Name Check [Status: TODO]

**Objective**: Before adding the `otlp-service-name-pattern` fitness check, fix the pre-existing `sm-kms-pg-1` / `sm-kms-pg-2` naming inconsistency in standalone config files so the check can immediately pass.

- Fix `configs/sm/kms/config-pg-1.yml`: comment + `otlp-service: "sm-kms-pg-1"` → `"sm-kms-postgres-1"`
- Fix `configs/sm/kms/config-pg-2.yml`: comment + `otlp-service: "sm-kms-pg-2"` → `"sm-kms-postgres-2"`
- Add fitness check `otlp-service-name-pattern`: validate `otlp-service` values in `configs/{PRODUCT}/{SERVICE}/config-*.yml` follow `{PS-ID}-sqlite-1`, `{PS-ID}-postgres-1`, `{PS-ID}-postgres-2` patterns
- **Success**: `lint-fitness otlp-service-name-pattern` passes, sm-kms configs use correct names
- **Post-Mortem**: lessons.md updated

### Phase 2: Registry-Driven Foundation and Entity Registry Check [Status: TODO]

**Objective**: Create the shared entity registry and the first registry-driven check.

- Create `internal/apps/cicd/lint_fitness/registry/registry.go` — canonical entity registry (products, product-services, suite) as Go structs
- Add fitness check `entity-registry-completeness`: each entity in registry has required deployment dir, config dir, and magic constants file
- Verify that all 10 product-services satisfy the registry's structural requirements
- Document registry update procedure in ARCHITECTURE.md
- **Success**: Registry exists, `entity-registry-completeness` check passes for all 16 entities
- **Post-Mortem**: lessons.md updated

### Phase 3: Banned Name Detection [Status: TODO]

**Objective**: Prevent recurrence of the "Cipher IM" naming regression.

- Add fitness check `banned-product-names`: scan all `.go`, `.yml`, `.yaml`, `.sql`, `.md` files for banned drift indicators: `Cipher IM`, `cipher-im`, `cipher_im`, `CipherIM`, `cryptoutilCmdCipher`
  - Award: intentional crypto terminology (e.g., `cipher.Block`, `var ciphertext`) excluded via exact-match exclusion list
  - Exclusion rationale: ONLY exact banned phrases are rejected, not the substring `cipher` in any context
- Add fitness check `legacy-dir-detection`: reject `internal/apps/cipher/` if it exists
- **Success**: Both checks pass, would have caught the Cipher→SM regression at commit time
- **Post-Mortem**: lessons.md updated

### Phase 4: Deployment Directory Completeness [Status: TODO]

**Objective**: Every product-service must have a complete deployment directory.

- Add fitness check `deployment-dir-completeness`: for each PS in registry, verify:
  - `deployments/{PS-ID}/` exists
  - `deployments/{PS-ID}/Dockerfile` exists
  - `deployments/{PS-ID}/compose.yml` exists
  - `deployments/{PS-ID}/secrets/` exists
  - `deployments/{PS-ID}/config/` exists
  - `deployments/{PS-ID}/config/{PS-ID}-app-common.yml` exists
  - `deployments/{PS-ID}/config/{PS-ID}-app-sqlite-1.yml` exists
  - `deployments/{PS-ID}/config/{PS-ID}-app-postgresql-1.yml` exists
  - `deployments/{PS-ID}/config/{PS-ID}-app-postgresql-2.yml` exists
- **Success**: Check passes for all 10 product-services, missing files reported clearly
- **Post-Mortem**: lessons.md updated

### Phase 5: Compose File Header and Service Name Validation [Status: TODO]

**Objective**: Compose file headers and service names follow exact patterns.

- Add fitness check `compose-header-format`: for each PS in registry, verify `deployments/{PS-ID}/compose.yml` line 3 equals `# {PS-ID-UPPER} Docker Compose Configuration` and line 5 equals `# SERVICE-level deployment for {Display Name}.`
- Add fitness check `compose-service-names`: for each PS in registry, verify compose.yml contains service definitions named exactly: `{PS-ID}-app-sqlite-1`, `{PS-ID}-app-postgres-1`, `{PS-ID}-app-postgres-2`, `{PS-ID}-db-postgres-1`
- Add fitness check `compose-db-naming`: verify PostgreSQL service has `container_name: {PS-ID}-postgres` and `hostname: {PS-ID}-postgres`
- **Success**: All three checks pass for all 10 product-services
- **Post-Mortem**: lessons.md updated

### Phase 6: Magic Constants Cross-Reference Validation [Status: COMPLETE]

**Objective**: Magic constants must match actual compose service names.

- Add fitness check `magic-e2e-container-names`: for each PS in registry, parse `magic_{PRODUCT}_{SERVICE}.go` (or `magic_{PRODUCT}.go`), extract `*E2ESQLiteContainer`, `*E2EPostgreSQL1Container`, `*E2EPostgreSQL2Container` constant values, verify they match compose service names from `deployments/{PS-ID}/compose.yml`
- Add fitness check `magic-e2e-compose-path`: verify `*E2EComposeFile` constant resolves to the correct path from the e2e test directory
- **Success**: All magic container name constants are in sync with compose service names
- **Post-Mortem**: lessons.md updated

### Phase 7: Standalone Config File Presence and Naming [Status: COMPLETE]

**Objective**: Services with standalone configs have all required files with correct names.

- Scope: Only `sm-im` and `sm-kms` (services with standardized standalone configs in `configs/{PRODUCT}/{SERVICE}/`)
- Add fitness check `standalone-config-presence`: verify `configs/sm/im/` and `configs/sm/kms/` each contain `config-sqlite.yml`, `config-pg-1.yml`, `config-pg-2.yml`
- Add fitness check `standalone-config-otlp-names`: parse otlp-service values in each file, verify they match required pattern (sqlite→`{PS-ID}-sqlite-1`, pg-1→`{PS-ID}-postgres-1`, pg-2→`{PS-ID}-postgres-2`)
- **Success**: Both checks pass, standalone config OTLP names validated
- **Post-Mortem**: lessons.md updated

### Phase 8: Migration Comment Header Validation [Status: TODO]

**Objective**: Migration files have correct comment headers without old product names.

- Add fitness check `migration-comment-headers`: for each PS in registry, scan `internal/apps/{PRODUCT}/{SERVICE}/repository/migrations/*.up.sql` — first comment line must start with `-- {Display Name} database schema`
- Also check `*.down.sql` — first comment line must start with `-- {Display Name} database schema rollback`
- **Success**: All migration comment headers validated, would have caught `SM IM` vs `Cipher IM` regression
- **Post-Mortem**: lessons.md updated

### Phase 9: ARCHITECTURE.md Updates and CICD Tool Catalog [Status: TODO]

**Objective**: Update ARCHITECTURE.md fitness function catalog to include all new checks.

- Update ARCHITECTURE.md Section 9.11: change "23 total" to reflect actual count (~55)
- Add all new checks to the Sub-Linter Catalog table with their rule descriptions
- Add "Entity Registry" section: document registry location and update procedure
- Update `cicd-lint-fitness` workflow description to reflect expanded scope
- **Success**: ARCHITECTURE.md accurately reflects all fitness checks; `cicd lint-docs` passes
- **Post-Mortem**: lessons.md updated

### Phase 10: Knowledge Propagation [Status: TODO]

**Objective**: Propagate lessons and patterns to agents, skills, instructions, ARCHITECTURE.md.

- Propagate entity registry pattern to instructions (add to `02-01.architecture.instructions.md`)
- Propagate banned name list to `02-01.architecture.instructions.md`
- Add `fitness-function-gen` skill update: new checks use registry-driven pattern
- Update framework-v3 status to COMPLETE (Phase 9 already marked TODO, complete it)
- **Success**: All knowledge propagated, cross-artifact consistency verified
- **Post-Mortem**: lessons.md updated

---

## Known Pre-Existing Issues to Fix (Before Each Phase's Check Can Pass)

| Issue | Affected File(s) | Phase |
|-------|-----------------|-------|
| `otlp-service: "sm-kms-pg-1"` | `configs/sm/kms/config-pg-1.yml` | Phase 1 |
| `otlp-service: "sm-kms-pg-2"` | `configs/sm/kms/config-pg-2.yml` | Phase 1 |
| Comment says "settings for sm-kms-pg-1" | Same files | Phase 1 |

---

## Cross-References

- **Framework v1**: Archived, 48/48 tasks complete
- **Framework v2**: Archived, 23/23 tasks complete
- **Framework v3**: IN PROGRESS (Phase 9 todo), 30 fitness checks established
- **Architecture**: `docs/ARCHITECTURE.md` (single source of truth, Section 9.11)
- **Existing checks**: `internal/apps/cicd/lint_fitness/` (30 sub-linters)
- **Entity registry proposal**: to be created at `internal/apps/cicd/lint_fitness/registry/registry.go`
- **Testing Strategy**: ARCHITECTURE.md Section 10 (testing architecture)
- **Quality Gates**: ARCHITECTURE.md Section 11.2 (quality gates)
- **Fitness Functions**: ARCHITECTURE.md Section 9.11 (fitness function catalog)
