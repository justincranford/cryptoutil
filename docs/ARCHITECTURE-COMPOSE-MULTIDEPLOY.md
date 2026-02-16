# Docker Compose Multi-Deploy Architecture

**Date**: 2025-02-15
**Status**: Validated with experiments
**Reference**: Q6 from [fixes-v2-quizme-v1.md](fixes-v2-quizme-v1.md)

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

### 2.6 Path Resolution

**Finding**: Docker resolves secret file paths relative to the compose file that defines them, then converts to absolute paths. These absolute paths must be accessible to the Docker daemon.

## 3. Recommended Directory Structure

```
deployments/
├── shared/                              # Shared infrastructure and secrets
│   ├── infra/
│   │   └── compose.yml                  # Shared DB, telemetry
│   └── secrets/
│       └── hash_pepper-SHARED.secret    # Suite-wide hash pepper
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
│   └── compose.yml                      # include: ../identity-authz/compose.yml
│                                        #          ../identity-idp/compose.yml
│                                        #          ../identity-rp/compose.yml
│                                        #          ../identity-rs/compose.yml
│                                        #          ../identity-spa/compose.yml
│
└── cryptoutil/                          # SUITE level (future)
    └── compose.yml                      # include: ../sm/compose.yml
                                         #          ../pki/compose.yml
                                         #          ../identity/compose.yml
                                         #          ../cipher/compose.yml
                                         #          ../jose/compose.yml
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
      - hash_pepper.secret
      - sm-kms-postgres_url.secret
      - sm-kms-postgres_username.secret
      - sm-kms-postgres_password.secret
      - sm-kms-postgres_database.secret

secrets:
  sm-kms-unseal_1of5.secret:
    file: ./secrets/sm-kms-unseal_1of5.secret
  hash_pepper.secret:
    file: ../shared/secrets/hash_pepper.secret
  # ... etc
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
| Hash pepper (shared) | `hash_pepper_v3.secret` | Same file referenced by all services |
| PostgreSQL URL | `{PRODUCT}-{SERVICE}-postgres_url.secret` | `sm-kms-postgres_url.secret` |
| PostgreSQL user | `{PRODUCT}-{SERVICE}-postgres_username.secret` | `jose-ja-postgres_username.secret` |
| PostgreSQL pass | `{PRODUCT}-{SERVICE}-postgres_password.secret` | `identity-authz-postgres_password.secret` |
| PostgreSQL DB | `{PRODUCT}-{SERVICE}-postgres_database.secret` | `pki-ca-postgres_database.secret` |

### 5.2 Sharing Rules

- **Unique per service**: Unseal keys, PostgreSQL credentials → prefix with `{PRODUCT}-{SERVICE}-`
- **Shared across services**: Hash pepper → single name, single file in `shared/secrets/`
- **Level suffixes** (see ARCHITECTURE.md Section 12.3.3): `-SERVICEONLY`, `-PRODUCTONLY`, `-SUITEONLY`, `-SHARED`

## 6. Migration Path

### Phase 1: Current State (SERVICE-only)

All 9 services deploy independently. Each has its own compose.yml with included telemetry. Secret names are NOT prefixed (e.g., `unseal_1of5.secret`).

### Phase 2: Secret Prefixing

Rename secret files to include `{PRODUCT}-{SERVICE}-` prefix:
- `unseal_1of5.secret` → `sm-kms-unseal_1of5.secret`
- Update compose.yml secret references
- Update linter validation

### Phase 3: Shared Infrastructure

Create `deployments/shared/` directory:
- Move hash pepper to `shared/secrets/hash_pepper_v3.secret`
- Service compose files reference shared pepper via relative path
- Each service remains independently deployable

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
