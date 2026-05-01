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

### `server/**` Recursive Superset Across All 10 PS-IDs

- `apis`
- `apis/templates`
- `businesslogic`
- `cmd`
- `config`
- `handler`
- `middleware`
- `model`
- `repository`
- `repository/migrations`
- `repository/orm`
- `service`

### Root-Level PS-ID Directory Superset Across All 10

`api`, `auth`, `bootstrap`, `cli`, `client`, `clientauth`, `compliance`, `config`, `crypto`, `domain`, `domain-v2`, `dpop`, `e2e`, `intermediate`, `observability`, `pkce`, `profile`, `repository`, `repository-v2`, `security`, `server`, `service`, `storage`, `testing`, `unified`, `userauth`

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
