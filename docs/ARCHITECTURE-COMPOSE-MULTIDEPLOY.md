# Docker Compose Multi-Deploy Architecture

**Date**: 2026-02-16
**Status**: Implemented and validated
**Implementation**: SUITE/PRODUCT/SERVICE compose.yml files and hash_pepper secrets created

## 1. Overview

This document defines the Docker Compose composition patterns for deploying cryptoutil at three levels:

| Level | Scope | Example | Services |
|-------|-------|---------|----------|
| **SERVICE** | Single service | `sm-kms` | 1 service (3 instances: 1 SQLite + 2 PostgreSQL) |
| **PRODUCT** | Product group | `identity` | 1-5 services per product |
| **SUITE** | All services | `cryptoutil` | All 9 services across 5 products |

Each level is independently deployable while sharing infrastructure and secrets through compose `include` directives.

## 2. Key Findings from Experiments

### 2.1 Compose `include` Directive

**Finding**: `include` merges services from different compose files into a single project with shared networking.

```yaml
# product-level/compose.yml
include:
  - path: svc-a/compose.yml
  - path: svc-b/compose.yml
```

- All services share the same Docker network automatically.
- `depends_on` references work across included files.
- Each level produces a distinct project name (network prefix).

### 2.2 Secret Name Conflict Rule

**CRITICAL**: Secrets with the **same name** from different include files **CONFLICT** if they point to **different files**.

| Scenario | Result |
|----------|--------|
| Same name + same file path | ✅ Merged (deduplicated) |
| Same name + different file paths | ❌ **CONFLICT ERROR** |
| Different names + different files | ✅ Works |

**Implication**: Service-specific secrets (unseal keys, postgres credentials) MUST have unique names prefixed with `{PRODUCT}-{SERVICE}-` when used in PRODUCT/SUITE compositions.

**Shared secrets** (hash pepper) can use the same name if ALL services reference the same file (e.g., `../shared/secrets/hash_pepper.secret`).

### 2.3 Infrastructure Deduplication

**Finding**: When multiple services include the same infrastructure compose file, Docker Compose correctly deduplicates. The infrastructure service appears once in the merged configuration.

```yaml
# svc-a/compose.yml
include:
  - path: ../shared/infra/compose.yml  # shared-db defined here

# svc-b/compose.yml
include:
  - path: ../shared/infra/compose.yml  # same file

# product/compose.yml (includes both svc-a and svc-b)
# Result: shared-db appears ONCE, both services can depend on it
```

### 2.4 Three-Level Composition Chain

**Finding**: Compose `include` chains work transitively:

```
SUITE compose.yml
  └── include: PRODUCT compose.yml
       └── include: SERVICE compose.yml
            └── include: INFRA compose.yml
```

Each level can be deployed independently:
- `cd deployments/sm-kms && docker compose up` — single service
- `cd deployments/identity && docker compose up` — all identity services
- `cd deployments/cryptoutil && docker compose up` — all 9 services

### 2.5 Compose `extends` for Templates

**Finding**: `extends` inherits service configuration from a template file:

```yaml
services:
  svc-a-app:
    extends:
      file: ../template/compose.yml
      service: service-template
    secrets:
      - svc-a-unseal.secret
```

**Caveat**: `${VAR}` in compose files is compose-time interpolation (from host env or `.env` file), NOT container runtime variable expansion.

### 2.6 Path Resolution & Secret Inheritance

**Finding**: Docker resolves secret file paths relative to the compose file that defines them, then converts to absolute paths. These absolute paths must be accessible to the Docker daemon.

**Complete Secret Hierarchy** (all 5 products, all 9 services):

**Key Changes**:
- PRODUCT/SUITE levels use `.never` files for postgres secrets (documents prohibition)
- PRODUCT/SUITE levels use `.never` files for unseal keys (documents prohibition)
- Only SERVICE level contains actual postgres and unseal secrets (security isolation)

```
deployments/
├── cryptoutil/                                    # SUITE-level deployment
│   ├── compose.yml                                # Includes all PRODUCT compose files
│   └── secrets/
│       ├── cryptoutil-hash_pepper.secret          # SUITE pepper: shared by ALL 9 services
│       ├── cryptoutil-unseal_1of5.secret.never    # Documents: unseal keys MUST NOT be shared
│       ├── cryptoutil-unseal_2of5.secret.never
│       ├── cryptoutil-unseal_3of5.secret.never
│       ├── cryptoutil-unseal_4of5.secret.never
│       ├── cryptoutil-unseal_5of5.secret.never
│       ├── cryptoutil-postgres_url.secret.never   # Documents: postgres secrets MUST NOT be shared
│       ├── cryptoutil-postgres_username.secret.never
│       ├── cryptoutil-postgres_password.secret.never
│       └── cryptoutil-postgres_database.secret.never
│
├── sm/                                            # PRODUCT-level (single-service product)
│   ├── compose.yml → ../sm-kms/compose.yml        # Alias to SERVICE
│   └── secrets/
│       ├── sm-hash_pepper.secret                  # PRODUCT pepper: only for sm-kms
│       ├── sm-unseal_1of5.secret.never            # Documents: unseal keys MUST NOT be shared
│       ├── sm-unseal_2of5.secret.never
│       ├── sm-unseal_3of5.secret.never
│       ├── sm-unseal_4of5.secret.never
│       ├── sm-unseal_5of5.secret.never
│       ├── sm-postgres_url.secret.never           # Documents: postgres secrets MUST NOT be shared
│       ├── sm-postgres_username.secret.never
│       ├── sm-postgres_password.secret.never
│       └── sm-postgres_database.secret.never
│
├── pki/                                           # PRODUCT-level (single-service product)
│   ├── compose.yml → ../pki-ca/compose.yml
│   └── secrets/
│       ├── pki-hash_pepper.secret                 # PRODUCT pepper: only for pki-ca
│       ├── pki-unseal_1of5.secret.never           # Documents: unseal keys MUST NOT be shared
│       ├── pki-unseal_2of5.secret.never
│       ├── pki-unseal_3of5.secret.never
│       ├── pki-unseal_4of5.secret.never
│       ├── pki-unseal_5of5.secret.never
│       ├── pki-postgres_url.secret.never          # Documents: postgres secrets MUST NOT be shared
│       ├── pki-postgres_username.secret.never
│       ├── pki-postgres_password.secret.never
│       └── pki-postgres_database.secret.never
│
├── identity/                                      # PRODUCT-level (multi-service product)
│   ├── compose.yml                                # Includes 5 identity services
│   └── secrets/
│       ├── identity-hash_pepper.secret            # PRODUCT pepper: shared by 5 identity services
│       ├── identity-unseal_1of5.secret.never      # Documents: unseal keys MUST NOT be shared
│       ├── identity-unseal_2of5.secret.never
│       ├── identity-unseal_3of5.secret.never
│       ├── identity-unseal_4of5.secret.never
│       ├── identity-unseal_5of5.secret.never
│       ├── identity-postgres_url.secret.never     # Documents: postgres secrets MUST NOT be shared
│       ├── identity-postgres_username.secret.never
│       ├── identity-postgres_password.secret.never
│       └── identity-postgres_database.secret.never
│
├── cipher/                                        # PRODUCT-level (single-service product)
│   ├── compose.yml → ../cipher-im/compose.yml
│   └── secrets/
│       ├── cipher-hash_pepper.secret              # PRODUCT pepper: only for cipher-im
│       ├── cipher-unseal_1of5.secret.never        # Documents: unseal keys MUST NOT be shared
│       ├── cipher-unseal_2of5.secret.never
│       ├── cipher-unseal_3of5.secret.never
│       ├── cipher-unseal_4of5.secret.never
│       ├── cipher-unseal_5of5.secret.never
│       ├── cipher-postgres_url.secret.never       # Documents: postgres secrets MUST NOT be shared
│       ├── cipher-postgres_username.secret.never
│       ├── cipher-postgres_password.secret.never
│       └── cipher-postgres_database.secret.never
│
├── jose/                                          # PRODUCT-level (single-service product)
│   ├── compose.yml → ../jose-ja/compose.yml
│   └── secrets/
│       ├── jose-hash_pepper.secret                # PRODUCT pepper: only for jose-ja
│       ├── jose-unseal_1of5.secret.never          # Documents: unseal keys MUST NOT be shared
│       ├── jose-unseal_2of5.secret.never
│       ├── jose-unseal_3of5.secret.never
│       ├── jose-unseal_4of5.secret.never
│       ├── jose-unseal_5of5.secret.never
│       ├── jose-postgres_url.secret.never         # Documents: postgres secrets MUST NOT be shared
│       ├── jose-postgres_username.secret.never
│       ├── jose-postgres_password.secret.never
│       └── jose-postgres_database.secret.never
│
├── sm-kms/                                        # SERVICE-level (sm product, kms service)
│   ├── compose.yml
│   └── secrets/
│       ├── sm-kms-hash_pepper.secret              # SERVICE pepper: unique to sm-kms
│       ├── sm-kms-unseal_1of5.secret              # Actual unseal keys (service-specific)
│       ├── sm-kms-unseal_2of5.secret
│       ├── sm-kms-unseal_3of5.secret
│       ├── sm-kms-unseal_4of5.secret
│       ├── sm-kms-unseal_5of5.secret
│       ├── sm-kms-postgres_url.secret             # Actual postgres secrets (service-specific)
│       ├── sm-kms-postgres_username.secret
│       ├── sm-kms-postgres_password.secret
│       └── sm-kms-postgres_database.secret
│
├── pki-ca/                                        # SERVICE-level (pki product, ca service)
│   ├── compose.yml
│   └── secrets/
│       ├── pki-ca-hash_pepper.secret              # SERVICE pepper: unique to pki-ca
│       ├── pki-ca-unseal_1of5.secret
│       ├── pki-ca-unseal_2of5.secret
│       ├── pki-ca-unseal_3of5.secret
│       ├── pki-ca-unseal_4of5.secret
│       ├── pki-ca-unseal_5of5.secret
│       ├── pki-ca-postgres_url.secret
│       ├── pki-ca-postgres_username.secret
│       ├── pki-ca-postgres_password.secret
│       └── pki-ca-postgres_database.secret
│
├── identity-authz/                                # SERVICE-level (identity product, authz service)
│   ├── compose.yml
│   └── secrets/
│       ├── identity-authz-hash_pepper.secret      # SERVICE pepper: unique to identity-authz
│       ├── identity-authz-unseal_1of5.secret
│       ├── identity-authz-unseal_2of5.secret
│       ├── identity-authz-unseal_3of5.secret
│       ├── identity-authz-unseal_4of5.secret
│       ├── identity-authz-unseal_5of5.secret
│       ├── identity-authz-postgres_url.secret
│       ├── identity-authz-postgres_username.secret
│       ├── identity-authz-postgres_password.secret
│       └── identity-authz-postgres_database.secret
│
├── identity-idp/                                  # SERVICE-level (identity product, idp service)
│   ├── compose.yml
│   └── secrets/
│       ├── identity-idp-hash_pepper.secret        # SERVICE pepper: unique to identity-idp
│       ├── identity-idp-unseal_1of5.secret
│       ├── identity-idp-unseal_2of5.secret
│       ├── identity-idp-unseal_3of5.secret
│       ├── identity-idp-unseal_4of5.secret
│       ├── identity-idp-unseal_5of5.secret
│       ├── identity-idp-postgres_url.secret
│       ├── identity-idp-postgres_username.secret
│       ├── identity-idp-postgres_password.secret
│       └── identity-idp-postgres_database.secret
│
├── identity-rp/                                   # SERVICE-level (identity product, rp service)
│   ├── compose.yml
│   └── secrets/
│       ├── identity-rp-hash_pepper.secret         # SERVICE pepper: unique to identity-rp
│       ├── identity-rp-unseal_1of5.secret
│       ├── identity-rp-unseal_2of5.secret
│       ├── identity-rp-unseal_3of5.secret
│       ├── identity-rp-unseal_4of5.secret
│       ├── identity-rp-unseal_5of5.secret
│       ├── identity-rp-postgres_url.secret
│       ├── identity-rp-postgres_username.secret
│       ├── identity-rp-postgres_password.secret
│       └── identity-rp-postgres_database.secret
│
├── identity-rs/                                   # SERVICE-level (identity product, rs service)
│   ├── compose.yml
│   └── secrets/
│       ├── identity-rs-hash_pepper.secret         # SERVICE pepper: unique to identity-rs
│       ├── identity-rs-unseal_1of5.secret
│       ├── identity-rs-unseal_2of5.secret
│       ├── identity-rs-unseal_3of5.secret
│       ├── identity-rs-unseal_4of5.secret
│       ├── identity-rs-unseal_5of5.secret
│       ├── identity-rs-postgres_url.secret
│       ├── identity-rs-postgres_username.secret
│       ├── identity-rs-postgres_password.secret
│       └── identity-rs-postgres_database.secret
│
├── identity-spa/                                  # SERVICE-level (identity product, spa service)
│   ├── compose.yml
│   └── secrets/
│       ├── identity-spa-hash_pepper.secret        # SERVICE pepper: unique to identity-spa
│       ├── identity-spa-unseal_1of5.secret
│       ├── identity-spa-unseal_2of5.secret
│       ├── identity-spa-unseal_3of5.secret
│       ├── identity-spa-unseal_4of5.secret
│       ├── identity-spa-unseal_5of5.secret
│       ├── identity-spa-postgres_url.secret
│       ├── identity-spa-postgres_username.secret
│       ├── identity-spa-postgres_password.secret
│       └── identity-spa-postgres_database.secret
│
├── cipher-im/                                     # SERVICE-level (cipher product, im service)
│   ├── compose.yml
│   └── secrets/
│       ├── cipher-im-hash_pepper.secret           # SERVICE pepper: unique to cipher-im
│       ├── cipher-im-unseal_1of5.secret
│       ├── cipher-im-unseal_2of5.secret
│       ├── cipher-im-unseal_3of5.secret
│       ├── cipher-im-unseal_4of5.secret
│       ├── cipher-im-unseal_5of5.secret
│       ├── cipher-im-postgres_url.secret
│       ├── cipher-im-postgres_username.secret
│       ├── cipher-im-postgres_password.secret
│       └── cipher-im-postgres_database.secret
│
└── jose-ja/                                       # SERVICE-level (jose product, ja service)
    ├── compose.yml
    └── secrets/
        ├── jose-ja-hash_pepper.secret             # SERVICE pepper: unique to jose-ja
        ├── jose-ja-unseal_1of5.secret
        ├── jose-ja-unseal_2of5.secret
        ├── jose-ja-unseal_3of5.secret
        ├── jose-ja-unseal_4of5.secret
        ├── jose-ja-unseal_5of5.secret
        ├── jose-ja-postgres_url.secret
        ├── jose-ja-postgres_username.secret
        ├── jose-ja-postgres_password.secret
        └── jose-ja-postgres_database.secret
```

**Pepper Inheritance by Deployment Scenario:**

| Deployment Command | Pepper Used | Services Affected | Scope |
|-------------------|-------------|-------------------|-------|
| `cd sm-kms && docker compose up` | `sm-kms-hash_pepper.secret` | sm-kms only | SERVICE-only: unique pepper |
| `cd pki-ca && docker compose up` | `pki-ca-hash_pepper.secret` | pki-ca only | SERVICE-only: unique pepper |
| `cd identity-authz && docker compose up` | `identity-authz-hash_pepper.secret` | identity-authz only | SERVICE-only: unique pepper |
| `cd identity && docker compose up` | `identity-hash_pepper.secret` | All 5 identity services (authz, idp, rp, rs, spa) | PRODUCT-level: shared within product |
| `cd cryptoutil && docker compose up` | `cryptoutil-hash_pepper.secret` | ALL 9 services across 5 products | SUITE-level: shared globally |

**Key Rules:**

1. **SERVICE-only deployment**: Each service uses its own unique `{PRODUCT}-{SERVICE}-hash_pepper.secret` from its own `secrets/` directory.

2. **PRODUCT-level deployment**: All services within a product share `{PRODUCT}-hash_pepper.secret` from the PRODUCT directory's `secrets/` folder. The PRODUCT compose.yml defines this secret and it overrides the SERVICE-level secrets via Docker Compose merging rules.

3. **SUITE-level deployment**: All 9 services across all 5 products share `cryptoutil-hash_pepper.secret` from `deployments/cryptoutil/secrets/`. The SUITE compose.yml defines this secret at the top level, overriding both PRODUCT and SERVICE secrets.

4. **Secret precedence**: SUITE > PRODUCT > SERVICE (compose merging gives precedence to the parent that includes children).

5. **Unseal keys, DB credentials**: ALWAYS unique per service, NEVER shared. Only hash pepper has layered sharing.

## 3. Recommended Directory Structure

```
deployments/
├── shared/                              # Shared infrastructure ONLY (NO secrets)
│   └── infra/
│       └── compose.yml                  # Shared DB, telemetry
│
├── template/                            # Compose and config templates
│   ├── compose.yml                      # Service compose template
│   ├── compose-cryptoutil-PRODUCT-SERVICE.yml
│   ├── compose-cryptoutil-PRODUCT.yml
│   └── compose-cryptoutil.yml
│
├── sm-kms/                              # SERVICE level (standalone)
│   ├── compose.yml                      # include: ../shared/infra/compose.yml
│   ├── Dockerfile
│   ├── secrets/
│   │   ├── sm-kms-unseal_1of5.secret    # Service-specific (unique name)
│   │   └── ...
│   └── config/
│       ├── sm-kms-app-common.yml
│       └── ...
│
├── identity-authz/                      # SERVICE level
│   ├── compose.yml
│   ├── Dockerfile
│   ├── secrets/
│   │   ├── identity-authz-unseal_1of5.secret
│   │   └── ...
│   └── config/
│
├── identity/                            # PRODUCT level (future)
│   ├── compose.yml                      # include: ../identity-authz/compose.yml
│   │                                    #          ../identity-idp/compose.yml
│   │                                    #          ../identity-rp/compose.yml
│   │                                    #          ../identity-rs/compose.yml
│   │                                    #          ../identity-spa/compose.yml
│   └── secrets/
│       └── identity-hash_pepper.secret  # PRODUCT pepper: shared by 5 identity services
│
└── cryptoutil/                          # SUITE level (future)
    ├── compose.yml                      # include: ../sm/compose.yml
    │                                    #          ../pki/compose.yml
    │                                    #          ../identity/compose.yml
    │                                    #          ../cipher/compose.yml
    │                                    #          ../jose/compose.yml
    └── secrets/
        └── cryptoutil-hash_pepper.secret # SUITE pepper: shared by ALL 9 services
```

## 4. Composition Patterns

### 4.1 SERVICE Level (Standalone)

Each service compose.yml includes shared infrastructure and defines its own secrets:

```yaml
# deployments/sm-kms/compose.yml
include:
  - path: ../telemetry/compose.yml

services:
  sm-kms-app-sqlite-1:
    build:
      context: ../../
      dockerfile: deployments/sm-kms/Dockerfile
    depends_on:
      healthcheck-secrets:
        condition: service_completed_successfully
    secrets:
      - sm-kms-unseal_1of5.secret
      - sm-kms-unseal_2of5.secret
      - sm-kms-unseal_3of5.secret
      - sm-kms-unseal_4of5.secret
      - sm-kms-unseal_5of5.secret
      - sm-kms-hash_pepper.secret
      - sm-kms-postgres_url.secret
      - sm-kms-postgres_username.secret
      - sm-kms-postgres_password.secret
      - sm-kms-postgres_database.secret

secrets:
  sm-kms-unseal_1of5.secret:
    file: ./secrets/sm-kms-unseal_1of5.secret
  sm-kms-hash_pepper.secret:
    file: ./secrets/sm-kms-hash_pepper.secret
  # ... etc (other unseal keys, postgres credentials)
```

### 4.2 PRODUCT Level (Aggregation)

Product compose.yml includes all service compose files:

```yaml
# deployments/identity/compose.yml
include:
  - path: ../identity-authz/compose.yml
  - path: ../identity-idp/compose.yml
  - path: ../identity-rp/compose.yml
  - path: ../identity-rs/compose.yml
  - path: ../identity-spa/compose.yml
```

### 4.3 SUITE Level (Full Deployment)

Suite compose.yml includes all product compose files:

```yaml
# deployments/cryptoutil/compose.yml
include:
  - path: ../sm/compose.yml      # or ../sm-kms/compose.yml for single-service products
  - path: ../pki/compose.yml     # or ../pki-ca/compose.yml
  - path: ../identity/compose.yml
  - path: ../cipher/compose.yml  # or ../cipher-im/compose.yml
  - path: ../jose/compose.yml    # or ../jose-ja/compose.yml
```

## 5. Secret Naming Strategy

### 5.1 Naming Convention

To avoid secret name conflicts in multi-level composition:

| Secret Type | Naming Pattern | Example |
|-------------|---------------|---------|
| Unseal keys | `{PRODUCT}-{SERVICE}-unseal_{N}of5.secret` | `sm-kms-unseal_1of5.secret` |
| Hash pepper (SERVICE) | `{PRODUCT}-{SERVICE}-hash_pepper.secret` | `sm-kms-hash_pepper.secret` |
| Hash pepper (PRODUCT) | `{PRODUCT}-hash_pepper.secret` | `identity-hash_pepper.secret` (shared by 5 services) |
| Hash pepper (SUITE) | `cryptoutil-hash_pepper.secret` | `cryptoutil-hash_pepper.secret` (shared by all 9) |
| PostgreSQL URL | `{PRODUCT}-{SERVICE}-postgres_url.secret` | `sm-kms-postgres_url.secret` |
| PostgreSQL user | `{PRODUCT}-{SERVICE}-postgres_username.secret` | `jose-ja-postgres_username.secret` |
| PostgreSQL pass | `{PRODUCT}-{SERVICE}-postgres_password.secret` | `identity-authz-postgres_password.secret` |
| PostgreSQL DB | `{PRODUCT}-{SERVICE}-postgres_database.secret` | `pki-ca-postgres_database.secret` |

### 5.2 Layered Pepper Strategy

**SERVICE-only deployment** (`cd {PRODUCT}-{SERVICE} && docker compose up`):
- Each service has unique pepper: `{PRODUCT}-{SERVICE}-hash_pepper.secret`
- Example: `sm-kms` uses `sm-kms-hash_pepper.secret`, `pki-ca` uses `pki-ca-hash_pepper.secret`
- **Use case**: Maximum isolation during development/testing

**PRODUCT-level deployment** (`cd {PRODUCT} && docker compose up`):
- All services within product share pepper: `{PRODUCT}-hash_pepper.secret`
- Example: All 5 identity services (authz, idp, rp, rs, spa) share `identity-hash_pepper.secret`
- Single-service products (sm, pki, cipher, jose): `{PRODUCT}-hash_pepper.secret` = alias to SERVICE pepper
- **Use case**: Shared SSO/federation within product boundary

**SUITE-level deployment** (`cd cryptoutil && docker compose up`):
- All 9 services across 5 products share pepper: `cryptoutil-hash_pepper.secret`
- **Use case**: Cross-product SSO, unified identity federation

**Other Secrets (NEVER shared)**:
- Unseal keys: ALWAYS `{PRODUCT}-{SERVICE}-unseal_{N}of5.secret` (unique per service)
- PostgreSQL credentials: ALWAYS `{PRODUCT}-{SERVICE}-postgres_*.secret` (unique per service)
- **Level suffixes** (see ARCHITECTURE.md Section 12.3.3): `-SERVICEONLY`, `-PRODUCTONLY`, `-SUITEONLY` used in filename hints only

## 6. Migration Path

### Phase 1: Current State (SERVICE-only)

All 9 services deploy independently. Each has its own compose.yml with included telemetry. Secret names are NOT prefixed (e.g., `unseal_1of5.secret`).

### Phase 2: Secret Prefixing

Rename secret files to include `{PRODUCT}-{SERVICE}-` prefix:
- `unseal_1of5.secret` → `sm-kms-unseal_1of5.secret`
- Update compose.yml secret references
- Update linter validation

### Phase 3: Layered Pepper Strategy

Create layered pepper secrets:
- SERVICE-level: Each service has `{PRODUCT}-{SERVICE}-hash_pepper.secret` in its own `secrets/` directory
- PRODUCT-level: Create `deployments/{PRODUCT}/secrets/{PRODUCT}-hash_pepper.secret` for multi-service products
- SUITE-level: Create `deployments/cryptoutil/secrets/cryptoutil-hash_pepper.secret` for full deployment
- Update compose.yml files to reference appropriate pepper based on deployment level

### Phase 4: Product-Level Composition

Create product directories with aggregation compose files:
- `deployments/identity/compose.yml` includes all 5 identity services
- `deployments/sm/compose.yml` → `deployments/sm-kms/compose.yml`
- Test: `cd deployments/identity && docker compose up`

### Phase 5: Suite-Level Composition

Create suite directory with full aggregation:
- `deployments/cryptoutil/compose.yml` includes all products
- Test: `cd deployments/cryptoutil && docker compose up`

## 7. Experimental Evidence

All findings in this document were validated with Docker Compose v2.40.3 using alpine:3.19 containers.

### Experiments Conducted

| # | Description | Result |
|---|-------------|--------|
| 1 | `include` merges services from subdirectories | ✅ Services share network |
| 2 | Same secret name + same file across includes | ✅ Deduplicated correctly |
| 3 | Unique secret names per service | ✅ No conflicts |
| 4 | Same secret name + different files | ❌ **CONFLICT** (expected) |
| 5 | 3-level hierarchy (SUITE → PRODUCT → SERVICE) | ✅ All levels work |
| 6 | `extends` for template inheritance | ✅ Config inherited correctly |
| 7 | Combined include + deps + shared infra | ✅ Infra deduplicated, secrets shared |

### Key Constraints

1. Secret names MUST be globally unique within a compose project (unless pointing to same file).
2. `${VAR}` interpolation happens at compose-time, not container runtime.
3. All secret file paths must be accessible to the Docker daemon.
4. Infrastructure services included by multiple service files are correctly deduplicated.

## 8. Cross-References

- [ARCHITECTURE.md Section 12.3.3](ARCHITECTURE.md#1233-secrets-coordination-strategy) — Secrets Coordination Strategy
- [ARCHITECTURE.md Section 12.4](ARCHITECTURE.md#124-deployment-structure-validation) — Deployment Structure Validation
- [ARCHITECTURE.md Section 3.4](ARCHITECTURE.md#34-port-assignments--networking) — Port Assignments & Networking
