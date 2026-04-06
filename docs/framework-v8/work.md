# Framework-v8 Work — Deployment Parameterization

**Created**: 2026-04-05
**Status**: Analysis Complete, Implementation Pending
**Purpose**: Deep analysis of `deployments/` architecture and recursive compose import strategy.

---

## 1. Current Architecture Summary

### 1.1 Three Deployment Tiers

| Tier | Directory | Port Offset | Instances/PS-ID | compose.yml Size |
|------|-----------|-------------|-----------------|------------------|
| PS-ID (SERVICE) | `deployments/{PS-ID}/` | +0 (8000-8999) | 4 (sqlite-1, sqlite-2, postgresql-1, postgresql-2) | 303-333 lines |
| PRODUCT | `deployments/{PRODUCT}/` | +10000 (18000-18999) | 3 (sqlite-1, postgres-1, postgres-2) | 261-818 lines |
| SUITE | `deployments/cryptoutil/` | +20000 (28000-28999) | 3 (sqlite-1, postgres-1, postgres-2) | 1504 lines |

### 1.2 Current Line Counts

```
PS-ID level (10 files):   3,118 lines total
PRODUCT level (5 files):  2,014 lines total
SUITE level (1 file):     1,504 lines total
Shared infra (2 files):   varies
────────────────────────────────────
Total:                    6,636+ lines
```

### 1.3 Current `include:` Usage

ALL compose files include shared telemetry:

```yaml
include:
  - path: ../shared-telemetry/compose.yml
```

ONLY the SUITE-level compose additionally includes shared PostgreSQL:

```yaml
include:
  - path: ../shared-telemetry/compose.yml
  - path: ../shared-postgres/compose.yml
```

**NO vertical includes exist**: PS-ID is NOT included by PRODUCT, PRODUCT is NOT included by SUITE.
Every tier re-declares all service definitions from scratch.

### 1.4 Artifacts Per Tier

| Artifact | PS-ID | PRODUCT | SUITE |
|----------|-------|---------|-------|
| compose.yml | ✅ | ✅ | ✅ |
| Dockerfile | ✅ | ❌ (missing) | ✅ |
| config/ (5 variants) | ✅ | ❌ (uses PS-ID configs) | ❌ (uses PS-ID configs) |
| secrets/ | ✅ (PS-ID scope) | ✅ (product scope) | ✅ (suite scope) |

**Key insight**: PRODUCT and SUITE already reference PS-ID config files via volume mounts
(e.g., `../sm-kms/config/sm-kms-app-sqlite-1.yml:/app/config/...`). No config duplication.

---

## 2. Identified Issues

### 2.1 sqlite-2 Missing at PRODUCT and SUITE Tiers

PS-ID level has 4 variants per service (sqlite-1, sqlite-2, postgresql-1, postgresql-2).
PRODUCT and SUITE levels only have 3 variants (sqlite-1, postgres-1, postgres-2).

**Impact**: sqlite-2 provides dual-instance testing for cross-database compatibility without
requiring PostgreSQL. Its absence at higher tiers means product-level and suite-level deployments
cannot verify dual-SQLite behavior.

### 2.2 Service Naming Inconsistency: `postgres` vs `postgresql`

| Tier | Pattern | Example |
|------|---------|---------|
| PS-ID | `postgresql` (full word) | `sm-im-app-postgresql-1` |
| PRODUCT | Mixed (`postgres` / `postgresql`) | `sm-kms-app-postgres-1` but `jose-ja-app-postgresql-1` |
| SUITE | `postgres` (abbreviated) | `sm-im-app-postgres-1` |

Config files consistently use `postgresql` (the full word). Service names SHOULD match.

**Resolution**: Standardize to `postgresql` everywhere (matching config file names and PS-ID
conventions). The config files are named `{PS-ID}-app-postgresql-{N}.yml`, so service names
should be `{PS-ID}-app-postgresql-{N}`.

### 2.3 SM Product Missing sm-im Services

`deployments/sm/compose.yml` only defines sm-kms services (3 variants). It completely omits
sm-im services. The SM product has 2 PS-IDs (sm-kms, sm-im) but only deploys one of them.

### 2.4 Massive Duplication Across Tiers

Each service definition (~25-40 lines) is copy-pasted across 3 tiers with only port numbers
changed. For 10 PS-IDs × 3-4 variants × 3 tiers, this is ~1,200 lines of pure duplication.

### 2.5 No Product-Level Dockerfiles

All 5 product directories lack Dockerfiles. All tiers use `deployments/cryptoutil/Dockerfile`.
The `builder-{PRODUCT}` services reference it explicitly:

```yaml
builder-sm:
  build:
    context: ../..
    dockerfile: deployments/cryptoutil/Dockerfile
```

---

## 3. Proposed Architecture: Recursive Compose Imports

### 3.1 Design Principle

Docker Compose `include:` supports recursive file inclusion. Each tier should include the
tier below it, with overrides for port offsets and secrets. This eliminates copy-paste of
service definitions across tiers.

### 3.2 Proposed Include Hierarchy

```
SUITE (deployments/cryptoutil/compose.yml)
├── include: ../shared-telemetry/compose.yml
├── include: ../shared-postgres/compose.yml
├── include: ../sm/compose.yml          ← PRODUCT
├── include: ../jose/compose.yml        ← PRODUCT
├── include: ../pki/compose.yml         ← PRODUCT
├── include: ../identity/compose.yml    ← PRODUCT
└── include: ../skeleton/compose.yml    ← PRODUCT

PRODUCT (e.g., deployments/sm/compose.yml)
├── include: ../shared-telemetry/compose.yml
├── include: ../sm-kms/compose.yml      ← PS-ID
└── include: ../sm-im/compose.yml       ← PS-ID

PS-ID (e.g., deployments/sm-im/compose.yml)
├── include: ../shared-telemetry/compose.yml
└── (unchanged — remains the canonical service definition)
```

### 3.3 Port Override Strategy

Docker Compose `include:` does NOT support overriding service ports from included files.
The included services retain their original port mappings.

**Two approaches**:

**Approach A — Environment Variable Substitution**:
Define ports via `${VAR:-default}` in PS-ID compose files:

```yaml
# PS-ID compose.yml
services:
  sm-im-app-sqlite-1:
    ports:
      - "${SM_IM_SQLITE_1_PORT:-8100}:8080"
```

```yaml
# PRODUCT compose.yml
include:
  - path: ../sm-im/compose.yml
    env_file: ./product-ports.env   # SM_IM_SQLITE_1_PORT=18100
```

**Approach B — Override Files**:
Use Docker Compose multiple `-f` flags:

```bash
# PS-ID
docker compose -f deployments/sm-im/compose.yml up

# PRODUCT (base + override)
docker compose \
  -f deployments/sm-im/compose.yml \
  -f deployments/sm/compose-overrides-sm-im.yml \
  up
```

**Approach C — `include:` with service overrides (PREFERRED)**:

Docker Compose v2.24.0+ supports `include:` with inline service attribute overrides via
top-level service redefinition. When a service name appears both in an included file and
in the including file, the including file's definition merges on top:

```yaml
# PRODUCT compose.yml
include:
  - path: ../sm-im/compose.yml

services:
  # Override ports from included sm-im compose
  sm-im-app-sqlite-1:
    ports:
      - "18100:8080"    # Override PS-ID port 8100 → PRODUCT port 18100
```

This is the cleanest approach — no env files, no multi-file flags.

### 3.4 Secret Scope Management

Each tier has its own `secrets/` directory with tier-scoped values:
- PS-ID: `{PS-ID}-unseal-key-N-of-5-{random}`
- PRODUCT: `{PRODUCT}-unseal-key-N-of-5-{random}`
- SUITE: `cryptoutil-unseal-key-N-of-5-{random}`

When using `include:`, the included PS-ID compose declares secrets with `file: ./secrets/...`
(relative to the PS-ID directory). The including PRODUCT compose needs to override secret
file paths to point to product-level secrets.

**Solution**: Redefine secrets in the PRODUCT/SUITE compose with product-scoped paths:

```yaml
# PRODUCT compose.yml
include:
  - path: ../sm-im/compose.yml

secrets:
  # Override secret file paths from included PS-ID
  unseal-1of5.secret:
    file: ./secrets/unseal-1of5.secret  # product-level secret
```

### 3.5 Network Isolation

Each PS-ID defines its own network (e.g., `sm-im-network`). When included at PRODUCT level,
these networks remain isolated per PS-ID (correct behavior — PS-IDs should not cross-communicate
except through defined APIs).

At PRODUCT level, a shared product network is added for the shared PostgreSQL instance.
Services needing PostgreSQL join both their PS-ID network and the product network.

### 3.6 Builder Consolidation

Currently each tier has its own `builder-{name}` service. With recursive includes, PS-ID
builders would be included in PRODUCT, creating duplicate build steps.

**Resolution**: The PRODUCT compose should NOT include PS-ID builders. Instead, the PRODUCT
compose defines a single `builder-{PRODUCT}` and PS-ID app services override `depends_on`
to depend on `builder-{PRODUCT}` instead of `builder-{PS-ID}`.

**Implication**: No Dockerfiles required at PRODUCT or SUITE level. All tiers build from
`deployments/cryptoutil/Dockerfile`. The existing `builder-{name}` services already reference
this single Dockerfile.

### 3.7 Database Service Handling

| Tier | Database Pattern |
|------|-----------------|
| PS-ID | Own PostgreSQL instance (`{PS-ID}-db-postgres-1`, port 543XX) |
| PRODUCT | Shared product PostgreSQL (`{PRODUCT}-db-postgres-1`, port 543XX) |
| SUITE | Shared suite PostgreSQL (leader/follower via `shared-postgres/compose.yml`) |

When PRODUCT includes PS-ID compose files, the PS-ID database services are also included.
The PRODUCT needs to either:
1. Remove PS-ID databases via compose profiles (PS-ID DB services get `profiles: ["standalone"]`)
2. Replace them with the shared product database

**Preferred**: Add a `standalone` profile to PS-ID database services. At PRODUCT level, the
`standalone` profile is never activated, so PS-ID databases don't start. Instead, the PRODUCT
compose defines its own shared database.

---

## 4. Implementation Plan

### Phase 1: Naming Standardization

1. Standardize service names to `postgresql` (matching PS-ID and config file conventions)
2. Add sqlite-2 variant to PRODUCT and SUITE levels (currently missing)
3. Add sm-im services to SM product compose (currently missing)

### Phase 2: PS-ID Compose Refactoring

1. Add `standalone` profile to PS-ID database services
2. Parameterize host ports via environment variable substitution
3. Ensure PS-ID compose files work both standalone AND as include targets

### Phase 3: PRODUCT Recursive Includes

1. Replace copy-pasted service definitions with `include:` of PS-ID compose files
2. Add port override services for PRODUCT-level port remapping (+10000)
3. Override secrets to use product-scoped values
4. Define shared PRODUCT database (replace per-PS-ID databases)
5. Define single `builder-{PRODUCT}` (replace included PS-ID builders)

### Phase 4: SUITE Recursive Includes

1. Replace all 30 service definitions with `include:` of 5 PRODUCT compose files
2. Add port override services for SUITE-level port remapping (+20000)
3. Override secrets to use suite-scoped values
4. Integrate with `shared-postgres/compose.yml` for leader/follower

### Phase 5: Validation

1. Update `lint-deployments` validators for recursive include support
2. Update `lint-ports` to validate through included files
3. Verify all 3 tiers start correctly (`docker compose up --profile dev`)
4. Verify PostgreSQL profiles work at all tiers
5. Run E2E tests against each tier

---

## 5. Expected Outcomes

### 5.1 Line Count Reduction

```
BEFORE:
  PS-ID (10):     3,118 lines
  PRODUCT (5):    2,014 lines
  SUITE (1):      1,504 lines
  Total:          6,636 lines

AFTER (estimated):
  PS-ID (10):     3,118 lines (unchanged — canonical definitions)
  PRODUCT (5):    ~500 lines (port overrides + shared DB + secrets)
  SUITE (1):      ~200 lines (port overrides + shared DB + secrets)
  Total:          ~3,818 lines (42% reduction)
```

### 5.2 Maintenance Benefits

- **Single source of truth**: Service definitions live ONLY in PS-ID compose files
- **Cascading updates**: Changes to PS-ID configs automatically propagate to PRODUCT and SUITE
- **No copy-paste drift**: Ports, healthchecks, resource limits stay consistent
- **sqlite-2 at all tiers**: 4 variants everywhere, not 3 at higher tiers

### 5.3 Dockerfile Strategy

No additional Dockerfiles needed at PRODUCT or SUITE levels. All tiers build from
`deployments/cryptoutil/Dockerfile` via `builder-{name}` services. The carryover item 2
("Create Product-Level Dockerfiles") can be re-evaluated — if recursive includes work,
the `builder-{PRODUCT}` service just builds the same image with different labels.

---

## 6. Docker Compose `include:` Reference

### 6.1 Syntax (Docker Compose v2.20+)

```yaml
include:
  - path: ../other/compose.yml
  - path: ../another/compose.yml
    env_file: ./overrides.env       # optional env overrides
    project_directory: ../another/  # resolve relative paths from here
```

### 6.2 Behavior

- Included services are merged into the including project
- Service name conflicts: including file wins (merge override)
- Networks, volumes, secrets from included files are available
- Relative paths in included files resolve from the included file's directory
- `project_directory` overrides relative path resolution

### 6.3 Limitations

- Cannot selectively exclude services from included files (use profiles instead)
- Cannot rename included services
- Secret `file:` paths resolve from the included file's directory (good for PS-ID isolation,
  but PRODUCT must redefine secrets for product-scoped values)

---

## 7. Open Questions

1. **Profile compatibility**: Do profiles in included files interact correctly with the
   including file's `--profile` flag? (Needs testing)
2. **Secret path resolution**: When PS-ID compose declares `file: ./secrets/unseal-1of5.secret`,
   does `include:` resolve this relative to the PS-ID directory? (Expected: yes)
3. **Network merging**: If PS-ID and PRODUCT both define `telemetry-network`, does Docker
   Compose merge them or conflict? (Expected: merge, since telemetry is included in both)
4. **Builder deduplication**: If PRODUCT includes 2 PS-IDs, both with `builder-{PS-ID}`,
   how to ensure only one build occurs? (Use profiles or remove from PS-ID-level)
5. **Carryover item 2**: Should Product Dockerfiles still be created, or is the builder
   pattern with `deployments/cryptoutil/Dockerfile` sufficient for all tiers?
