# Quiz Me - Framework V21: Canonical PS-ID Recursive Structure (Round 2)

**Created**: 2026-04-30
**Purpose**: Close the remaining Q2 decision by selecting the canonical recursive directory structure that will be enforced for all 10 PS-IDs.

---

## Research Snapshot (Evidence-Based)

### Requested Focus Services (interpreting `jose-ca` as `jose-ja`)

- `sm-kms` currently has: `server/businesslogic`, `server/handler`, `server/repository`, `server/repository/migrations`, `server/repository/orm`
- `sm-im` currently has: `server/apis`, `server/config`, `server/model`, `server/repository`, `server/repository/migrations`
- `jose-ja` currently has: `server/apis`, `server/config`, `server/model`, `server/repository`, `server/repository/migrations`, `server/service`
- `skeleton-template` currently has: `server/apis`, `server/config`, `server/handler`, `server/model`, `server/repository`, `server/repository/migrations`

## Recursive Directory Inventory with Per-Directory CSV File Lists

### Group 1: sm-kms, sm-im, jose-ja, skeleton-template

```text
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template} | __PS_ID__.go, __PS_ID___usage.go, __PS_ID___cli_test.go, __PS_ID___lifecycle_test.go, __PS_ID___port_conflict_test.go, testmain_test.go, README.md
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/client | client.go, client_*.go, package_test.go, *_test.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/e2e | e2e_*.go, testmain_e2e_test.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/testing | testmain_helper.go, testmain_helper_test.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/domain | *.go, *_test.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/repository | *.go, *_test.go, migrations.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/repository/migrations | *.up.sql, *.down.sql
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/server | server.go, public_server.go, admin.go, service.go, validator.go, swagger.go, swagger_test.go, *_lifecycle_test.go, *_port_conflict_test.go, *_integration_test.go, testmain_test.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/server/apis | *.go, *_test.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/server/config | config.go, config_test.go, config_test_helper.go, config_*_test.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/server/model | model.go, models.go, *_test.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/server/repository | *.go, *_test.go, migrations.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/server/repository/migrations | *.up.sql, *.down.sql
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/server/businesslogic | *.go, *_bench_test.go, *_fuzz_test.go, *_property_test.go, *_test.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/server/handler | *.go, *_test.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/server/service | *.go, *_test.go
internal/apps/{sm-kms,sm-im,jose-ja,skeleton-template}/server/repository/orm | *.go, *_test.go
```

### Group 2: pki-ca, identity-*

```text
internal/apps/{pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa} | __PS_ID__.go, __PS_ID___usage.go, __PS_ID___cli_test.go, __PS_ID___contract_test.go, __PS_ID___lifecycle_test.go, __PS_ID___port_conflict_test.go, testmain_test.go, README.md, *.TODO
internal/apps/{pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa}/client | client.go, package_test.go, client_*.go, *_test.go
internal/apps/{pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa}/e2e | *_e2e_test.go, testmain_e2e_test.go
internal/apps/{identity-authz,identity-idp,identity-rp,identity-rs,identity-spa}/unified | __PS_ID__.go
internal/apps/{identity-authz,identity-idp}/auth | *.go, *_test.go
internal/apps/{identity-authz,identity-idp}/clientauth | *.go, *_test.go
internal/apps/{identity-authz}/dpop | *.go, *_test.go
internal/apps/{identity-authz}/pkce | *.go, *_test.go
internal/apps/{identity-idp}/userauth | *.go, *_test.go
internal/apps/{identity-idp}/userauth/mocks | *.go, *_test.go
internal/apps/{pki-ca}/api | *.go
internal/apps/{pki-ca}/api/handler | *.go, *_test.go
internal/apps/{pki-ca}/{bootstrap,cli,compliance,config,crypto,domain,domain-v2,intermediate,observability,security,storage} | *.go, *_test.go
internal/apps/{pki-ca}/profile/certificate | *.go, *_test.go
internal/apps/{pki-ca}/profile/subject | *.go, *_test.go
internal/apps/{pki-ca}/service/{issuer,ra,revocation,timestamp} | *.go, *_bench_test.go, *_test.go
internal/apps/{pki-ca}/repository-v2 | migrations.go, migrations_test.go
internal/apps/{pki-ca}/repository-v2/migrations | *.up.sql, *.down.sql
internal/apps/{pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa}/server | server.go, public_server.go, admin.go, service.go, validator.go, swagger.go, swagger_test.go, *_lifecycle_test.go, *_port_conflict_test.go, *_integration_test.go, *_test.go, testmain_test.go
internal/apps/{pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa}/server/apis | *.go, *_test.go
internal/apps/{identity-idp}/server/apis/templates | *.html
internal/apps/{pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa}/server/config | config.go, config_test.go, config_test_helper.go, config_*_test.go
internal/apps/{pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa}/server/model | model.go, *_test.go
internal/apps/{pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa}/server/repository | migrations.go, *.go, *_test.go
internal/apps/{identity-authz,identity-idp,identity-rp,identity-rs,identity-spa}/server/repository/migrations | *.up.sql, *.down.sql
internal/apps/{pki-ca}/server/{cmd,middleware} | *.go, *_test.go
```

### pki-ca SQL Migration Evidence

- Current migration SQL files are in:
  - `internal/apps/pki-ca/repository-v2/migrations/5001_ca_items.up.sql`
  - `internal/apps/pki-ca/repository-v2/migrations/5001_ca_items.down.sql`

---

## Question 1: Canonical `server/**` recursive structure to enforce for all 10 PS-IDs

**Question**: Which policy should V21 adopt as the target canonical recursive `server/**` structure across all 10 PS-IDs (with linter/template enforcement)?

**A)** Strict immediate canonical set:
- Required everywhere: `server/apis`, `server/businesslogic`, `server/config`, `server/model`, `server/repository`, `server/repository/migrations`
- Forbidden everywhere: `server/handler`, `server/service`, `server/cmd`, `server/middleware`, `server/repository/orm`, `server/apis/templates`
- One-shot migration for all 10 in V21

**B)** Transitional canonical set with sunset (recommended):
- Required everywhere: `server/apis`, `server/businesslogic`, `server/config`, `server/model`, `server/repository`, `server/repository/migrations`
- Temporary allowlist (must be retired by scheduled phases): `server/handler`, `server/service`, `server/cmd`, `server/middleware`, `server/repository/orm`, `server/apis/templates`
- Linter enforces required-now plus time-boxed deprecation plan

**C)** Minimal convergence:
- Require only: `server/apis`, `server/model`, `server/repository`
- Keep service-specific subdirectories indefinitely (no sunset)

**D)** Keep current mixed structure and only ensure required dirs exist (no consolidation mandate)

**E)**

**Answer**:

**Rationale**: This decision controls the all-10 migration scope, linter invariants, and how aggressively sprawl (especially pki-ca) is reduced.

---

## Question 2: pki-ca consolidation strategy under the selected canonical policy

**Question**: For pki-ca package/subdirectory sprawl, which execution strategy should tasks implement?

**A)** Full consolidation in V21:
- Move/merge pki-ca subdirectories to canonical targets immediately
- Migrate domain packages that sit outside canonical paths
- Remove legacy directories in same phase

**B)** Two-stage consolidation (recommended):
- Stage 1 (V21): establish canonical `server/**` directories, introduce wrappers/adapters, migrate SQL paths from `repository-v2/migrations` to `server/repository/migrations`
- Stage 2 (next phase): move domain-heavy packages (`bootstrap`, `compliance`, `intermediate`, `profile`, `service`, `storage`, etc.) behind canonical boundaries and remove legacy paths after compatibility gates pass

**C)** Structural-only for V21:
- Create canonical dirs and linter checks
- Keep pki-ca legacy package sprawl untouched

**D)** pki-ca-specific exception:
- Exempt pki-ca from canonical structure and keep bespoke layout

**E)**

**Answer**:

**Rationale**: Determines whether V21 includes concrete pki-ca sprawl reduction tasks versus deferring most consolidation work.

---

## Question 3: Root-level PS-ID directory policy for all 10 services

**Question**: Should V21 enforce a canonical root-level PS-ID directory policy in addition to `server/**` policy?

**A)** Yes, strict required-only root set for all 10 (recommended):
- Required: `client`, `e2e`, `server`
- Optional (explicitly approved only): `testing`, `unified`, authn/authz-specific modules
- All other root-level directories must be migrated or explicitly sunset

**B)** Yes, but service-class based policy:
- Identity services may keep additional authn/authz roots
- pki-ca may keep additional PKI roots
- SM/JOSE services follow strict root set

**C)** No root-level policy in V21; enforce only `server/**`

**D)** Keep current root-level sprawl and rely on naming conventions only

**E)**

**Answer**:

**Rationale**: This controls whether V21 includes all-10 root-level cleanup tasks or limits scope to `server/**` only.
