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
| **SUITE** | All services | `cryptoutil` | All 10 services across 5 products |

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
- `cd deployments/cryptoutil-suite && docker compose up` — all 10 services

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

**Complete Secret Hierarchy** (all 5 products, all 10 services):

**Key Changes**:
- PRODUCT/SUITE levels use `.never` files for postgres secrets (documents prohibition)
- PRODUCT/SUITE levels use `.never` files for unseal keys (documents prohibition)
- Only SERVICE level contains actual postgres and unseal secrets (security isolation)

```
deployments/
├── cryptoutil/                                    # SUITE-level deployment
│   ├── compose.yml                                # Includes all PRODUCT compose files
│   └── secrets/
│       ├── cryptoutil-hash_pepper.secret          # SUITE pepper: shared by ALL 10 services
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
├── sm/                                            # PRODUCT-level (currently one service: kms)
│   ├── compose.yml → ../sm-kms/compose.yml        # Includes sm-kms SERVICE
│   └── secrets/
│       ├── sm-hash_pepper.secret                  # PRODUCT pepper: shared by all SM services (currently just sm-kms)
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
├── pki/                                           # PRODUCT-level (currently one service: ca)
│   ├── compose.yml → ../pki-ca/compose.yml
│   └── secrets/
│       ├── pki-hash_pepper.secret                 # PRODUCT pepper: shared by all PKI services (currently just pki-ca)
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
├── sm/                                        # PRODUCT-level (currently one service: im)
│   ├── compose.yml → ../sm-im/compose.yml
│   └── secrets/
│       ├── sm-hash_pepper.secret              # PRODUCT pepper: shared by all SM services (currently just sm-im)
│       ├── sm-unseal_1of5.secret.never        # Documents: unseal keys MUST NOT be shared
│       ├── sm-unseal_2of5.secret.never
│       ├── sm-unseal_3of5.secret.never
│       ├── sm-unseal_4of5.secret.never
│       ├── sm-unseal_5of5.secret.never
│       ├── sm-postgres_url.secret.never       # Documents: postgres secrets MUST NOT be shared
│       ├── sm-postgres_username.secret.never
│       ├── sm-postgres_password.secret.never
│       └── sm-postgres_database.secret.never
│
├── jose/                                          # PRODUCT-level (currently one service: ja)
│   ├── compose.yml → ../jose-ja/compose.yml
│   └── secrets/
│       ├── jose-hash_pepper.secret                # PRODUCT pepper: shared by all JOSE services (currently just jose-ja)
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
├── skeleton/                                      # PRODUCT-level (currently one service: template)
│   ├── compose.yml → ../skeleton-template/compose.yml
│   └── secrets/
│       ├── skeleton-hash_pepper.secret            # PRODUCT pepper: shared by all Skeleton services (currently just skeleton-template)
│       ├── skeleton-unseal_1of5.secret.never      # Documents: unseal keys MUST NOT be shared
│       ├── skeleton-unseal_2of5.secret.never
│       ├── skeleton-unseal_3of5.secret.never
│       ├── skeleton-unseal_4of5.secret.never
│       ├── skeleton-unseal_5of5.secret.never
│       ├── skeleton-postgres_url.secret.never     # Documents: postgres secrets MUST NOT be shared
│       ├── skeleton-postgres_username.secret.never
│       ├── skeleton-postgres_password.secret.never
│       └── skeleton-postgres_database.secret.never
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
├── sm-im/                                     # SERVICE-level (SM product, im service)
│   ├── compose.yml
│   └── secrets/
│       ├── sm-im-hash_pepper.secret           # SERVICE pepper: unique to sm-im
│       ├── sm-im-unseal_1of5.secret
│       ├── sm-im-unseal_2of5.secret
│       ├── sm-im-unseal_3of5.secret
│       ├── sm-im-unseal_4of5.secret
│       ├── sm-im-unseal_5of5.secret
│       ├── sm-im-postgres_url.secret
│       ├── sm-im-postgres_username.secret
│       ├── sm-im-postgres_password.secret
│       └── sm-im-postgres_database.secret
│
├── skeleton-template/                             # SERVICE-level (skeleton product, template service)
│   ├── compose.yml
│   └── secrets/
│       ├── skeleton-template-hash_pepper.secret   # SERVICE pepper: unique to skeleton-template
│       ├── skeleton-template-unseal_1of5.secret
│       ├── skeleton-template-unseal_2of5.secret
│       ├── skeleton-template-unseal_3of5.secret
│       ├── skeleton-template-unseal_4of5.secret
│       ├── skeleton-template-unseal_5of5.secret
│       ├── skeleton-template-postgres_url.secret
│       ├── skeleton-template-postgres_username.secret
│       ├── skeleton-template-postgres_password.secret
│       └── skeleton-template-postgres_database.secret
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
| `cd skeleton-template && docker compose up` | `skeleton-template-hash_pepper.secret` | skeleton-template only | SERVICE-only: unique pepper |
| `cd identity && docker compose up` | `identity-hash_pepper.secret` | All 5 identity services (authz, idp, rp, rs, spa) | PRODUCT-level: shared within product |
| `cd skeleton && docker compose up` | `skeleton-hash_pepper.secret` | All Skeleton services (currently just skeleton-template) | PRODUCT-level: shared within product |
| `cd cryptoutil && docker compose up` | `cryptoutil-hash_pepper.secret` | ALL 10 services across 5 products | SUITE-level: shared globally |

**Key Rules:**

1. **SERVICE-only deployment**: Each service uses its own unique `{PRODUCT}-{SERVICE}-hash_pepper.secret` from its own `secrets/` directory.

2. **PRODUCT-level deployment**: All services within a product share `{PRODUCT}-hash_pepper.secret` from the PRODUCT directory's `secrets/` folder. The PRODUCT compose.yml defines this secret and it overrides the SERVICE-level secrets via Docker Compose merging rules.

3. **SUITE-level deployment**: All 10 services across all 5 products share `cryptoutil-hash_pepper.secret` from `deployments/cryptoutil-suite/secrets/`. The SUITE compose.yml defines this secret at the top level, overriding both PRODUCT and SERVICE secrets.

4. **Secret precedence**: SUITE > PRODUCT > SERVICE (compose merging gives precedence to the parent that includes children).

5. **Unseal keys, DB credentials**: ALWAYS unique per service, NEVER shared. Only hash pepper has layered sharing.

**Config Files in ./deployments/**: Each SERVICE-level deployment includes a `config/` directory with 4 required instance-specific configuration files:

```
deployments/{PRODUCT}-{SERVICE}/config/
├── {PRODUCT}-{SERVICE}-app-common.yml        # Common settings (all instances)
├── {PRODUCT}-{SERVICE}-app-sqlite-1.yml      # SQLite instance #1 settings
├── {PRODUCT}-{SERVICE}-app-postgresql-1.yml  # PostgreSQL instance #1 settings
└── {PRODUCT}-{SERVICE}-app-postgresql-2.yml  # PostgreSQL instance #2 settings
```

**Complete Config Hierarchy** (all 10 services):

```
deployments/
├── sm-kms/config/
│   ├── sm-kms-app-common.yml
│   ├── sm-kms-app-sqlite-1.yml
│   ├── sm-kms-app-postgresql-1.yml
│   └── sm-kms-app-postgresql-2.yml
│
├── pki-ca/config/
│   ├── pki-ca-app-common.yml
│   ├── pki-ca-app-sqlite-1.yml
│   ├── pki-ca-app-postgresql-1.yml
│   └── pki-ca-app-postgresql-2.yml
│
├── sm-im/config/
│   ├── sm-im-app-common.yml
│   ├── sm-im-app-sqlite-1.yml
│   ├── sm-im-app-postgresql-1.yml
│   └── sm-im-app-postgresql-2.yml
│
├── jose-ja/config/
│   ├── jose-ja-app-common.yml
│   ├── jose-ja-app-sqlite-1.yml
│   ├── jose-ja-app-postgresql-1.yml
│   └── jose-ja-app-postgresql-2.yml
│
├── identity-authz/config/
│   ├── identity-authz-app-common.yml
│   ├── identity-authz-app-sqlite-1.yml
│   ├── identity-authz-app-postgresql-1.yml
│   └── identity-authz-app-postgresql-2.yml
│
├── identity-idp/config/
│   ├── identity-idp-app-common.yml
│   ├── identity-idp-app-sqlite-1.yml
│   ├── identity-idp-app-postgresql-1.yml
│   └── identity-idp-app-postgresql-2.yml
│
├── identity-rp/config/
│   ├── identity-rp-app-common.yml
│   ├── identity-rp-app-sqlite-1.yml
│   ├── identity-rp-app-postgresql-1.yml
│   └── identity-rp-app-postgresql-2.yml
│
├── identity-rs/config/
│   ├── identity-rs-app-common.yml
│   ├── identity-rs-app-sqlite-1.yml
│   ├── identity-rs-app-postgresql-1.yml
│   └── identity-rs-app-postgresql-2.yml
│
└── identity-spa/config/
    ├── identity-spa-app-common.yml
    ├── identity-spa-app-sqlite-1.yml
    ├── identity-spa-app-postgresql-1.yml
    └── identity-spa-app-postgresql-2.yml
```

**Total**: 40 deployment config files (10 services × 4 files each)

**Config File Content**: Deployment configs are minimal Docker Compose-specific settings, typically just:
- Server port overrides for instance isolation
- OTLP service names for telemetry differentiation
- Database URL references (via Docker secrets)

**Example** (`sm-kms-app-sqlite-1.yml`):
```yaml
server:
  port: 8080
```

**Archived Infrastructure** (`deployments/archived/compose-legacy/`):

```
deployments/archived/compose-legacy/
└── compose.yml                                # Legacy E2E testing infrastructure (archived)
```

**Purpose**: Previously standalone E2E testing compose that:
- Included `../shared-postgres/compose.yml` and `../shared-telemetry/compose.yml`
- Overrode otel-collector to expose ports for host-based E2E tests
- Provided service names matching test expectations (`cryptoutil-sqlite`, `cryptoutil-postgres-1`, `cryptoutil-postgres-2`)
- **ARCHIVED**: Replaced by SERVICE/PRODUCT/SUITE-level deployments per three-tier hierarchy

### 2.7 Standalone Configuration Directory (./configs/)

**Purpose**: Rich CLI/development configuration files for local development WITHOUT Docker Compose.

**Organization**: Product/service hierarchy with profiles and policies.

```
configs/
├── ca/
│   ├── ca-server.yml                          # PKI CA service config
│   └── profiles/                              # CA profiles (future)
│
├── sm/
│   ├── config.yml                             # Product-level config (future)
│   └── im/
│       ├── config.yml                         # Base config
│       ├── config-sqlite.yml                  # SQLite instance
│       ├── config-pg-1.yml                    # PostgreSQL instance #1
│       └── config-pg-2.yml                    # PostgreSQL instance #2
│
├── cryptoutil/
│   └── config.yml                             # SUITE-level config (future)
│
├── identity/
│   ├── authz.yml                              # Authorization server config
│   ├── authz-docker.yml                       # Docker-specific overrides
│   ├── idp.yml                                # Identity provider config
│   ├── idp-docker.yml                         # Docker-specific overrides
│   ├── rs.yml                                 # Resource server config
│   ├── rs-docker.yml                          # Docker-specific overrides
│   ├── development.yml                        # Development environment
│   ├── production.yml                         # Production environment
│   ├── test.yml                               # Test environment
│   ├── policies/
│   │   ├── adaptive-auth.yml                  # Adaptive authentication policy
│   │   ├── risk-scoring.yml                   # Risk scoring policy
│   │   └── step-up.yml                        # Step-up authentication policy
│   └── profiles/
│       ├── authz-idp.yml                      # Combined authz+idp
│       ├── authz-only.yml                     # Authorization only
│       ├── ci.yml                             # CI pipeline
│       ├── demo.yml                           # Demo mode
│       └── full-stack.yml                     # Full identity stack
│
├── jose/
│   └── jose-server.yml                        # JOSE service config
│
├── observability/
│   ├── grafana/                               # Grafana dashboards (future)
│   └── prometheus/
│       └── adaptive-auth-alerts.yml           # Prometheus alert rules
│
├── template/
│   ├── config-sqlite.yml                      # Template SQLite config
│   ├── config-pg-1.yml                        # Template PostgreSQL #1
│   └── config-pg-2.yml                        # Template PostgreSQL #2
│
└── test/
    └── config.yml                             # Test configuration
```

**Config File Content**: Standalone configs are comprehensive application settings:
- Server bindings (public/private protocol/address/port)
- TLS configuration (mode, certificates, key paths)
- Database URLs and connection pooling
- OTLP telemetry endpoints and sampling
- CORS settings and allowed origins
- Session configuration (algorithms, expiration)
- Realm definitions and authentication methods
- Complete application configuration (NOT minimal Docker overrides)

**Example** (`configs/sm/im/config-sqlite.yml` - partial):
```yaml
bind-public-protocol: "https"
bind-public-address: "0.0.0.0"
bind-public-port: 8070
bind-private-protocol: "https"
bind-private-address: "127.0.0.1"
bind-private-port: 9090
tls-public-mode: "auto"
tls-private-mode: "auto"
otlp: true
otlp-service: "sm-im-sqlite"
otlp-environment: "development"
otlp-endpoint: "http://sm-im-otel-collector:4317"
cors-max-age: 3600
cors-allowed-origins:
  - "https://localhost:8070"
  - "https://127.0.0.1:8070"
# ... many more settings ...
```

### 2.8 Configuration Directory Strategy: ./configs/ vs ./deployments/

**When to use ./configs/**:
- **Local development** with `go run` or compiled binaries
- **Direct CLI execution** without Docker Compose
- **Comprehensive configuration** with all application settings
- **Environment-specific profiles** (development, production, test)
- **Policy and profile variations** (adaptive auth, step-up, risk scoring)
- **Standalone service testing** without containerization

**When to use ./deployments/{SERVICE}/config/**:
- **Docker Compose deployment** with containers
- **Minimal instance-specific overrides** (port numbers, OTLP service names)
- **Multi-instance deployments** (1 SQLite + 2 PostgreSQL per service)
- **Container-based configuration injection** via volume mounts
- **Production deployment patterns** following Docker best practices

**Current Heuristics** (observed pattern):
- **./deployments/{SERVICE}/config/**: 4 files per service (common, sqlite-1, postgresql-1, postgresql-2)
  - Minimal configs: mainly port overrides for instance isolation
  - Docker Compose-specific: mounted as volumes into containers
  - Total: 40 files (10 services × 4 files)

- **./configs/**: Rich hierarchy with profiles and policies
  - Comprehensive configs: complete application settings
  - CLI/development-specific: used with `--config-file` flag
  - Total: 30+ files organized by product/service/purpose

**Best Practice Heuristics** (recommended pattern):
1. **Start with ./configs/** for new service development
   - Build comprehensive standalone configuration first
   - Validate all settings with local CLI execution
   - Easier to debug without Docker complexity

2. **Create ./deployments/{SERVICE}/config/** after CLI validation
   - Extract minimal Docker-specific overrides (ports, OTLP names)
   - Keep deployment configs as thin as possible
   - Reference ./configs/ examples for setting names and values

3. **Use ./configs/ profiles** for environment variations
   - development.yml, production.yml, test.yml
   - Policy-specific: adaptive-auth.yml, step-up.yml
   - Profile-specific: demo.yml, ci.yml, full-stack.yml

4. **Use ./deployments/{SERVICE}/config/** for instance differentiation
   - common.yml: shared settings for all instances
   - sqlite-1.yml: SQLite instance port override
   - postgresql-1.yml: PostgreSQL instance #1 port override
   - postgresql-2.yml: PostgreSQL instance #2 port override

**Key Differences**:

| Aspect | ./configs/ | ./deployments/{SERVICE}/config/ |
|--------|-----------|--------------------------------|
| **Purpose** | Local CLI development | Docker Compose deployment |
| **Scope** | Comprehensive settings | Minimal overrides |
| **Usage** | `go run cmd/{SERVICE}/ --config-file=configs/{SERVICE}/config.yml` | `docker compose up` |
| **Size** | Large files (100-300 lines) | Small files (2-20 lines) |
| **Content** | All application settings | Port numbers, OTLP names |
| **Examples** | TLS, CORS, sessions, realms | `server.port: 8080` |
| **Profiles** | Yes (dev, prod, test, demo) | No (instance-specific only) |

**Migration Pattern** (if consolidation needed):
- NOT RECOMMENDED: Deployment configs serve different purpose
- Current dual structure is optimal: rich CLI configs + minimal deployment overrides
- Future: Could generate deployment configs from ./configs/ templates, but provides little value

**Rationale**:
- Separation of concerns: Development (./configs/) vs Deployment (./deployments/)
- Docker Compose best practice: Minimal config overrides, environment-specific volumes
- Easier debugging: Rich standalone configs for local development
- Production-ready: Thin deployment configs with Docker secrets and environment variables

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
├── skeleton-template/                   # SERVICE level (skeleton product, template service)
│   ├── compose.yml
│   ├── Dockerfile
│   ├── secrets/
│   │   ├── skeleton-template-unseal_1of5.secret
│   │   └── ...
│   └── config/
│
├── skeleton/                            # PRODUCT level (currently one service: template)
│   ├── compose.yml                      # include: ../skeleton-template/compose.yml
│   └── secrets/
│       └── skeleton-hash_pepper.secret  # PRODUCT pepper: shared by all Skeleton services
│
└── cryptoutil/                          # SUITE level (future)
    ├── compose.yml                      # include: ../sm/compose.yml
    │                                    #          ../pki/compose.yml
    │                                    #          ../identity/compose.yml
    │                                    #          ../skeleton/compose.yml
    │                                    #          ../jose/compose.yml
    └── secrets/
        └── cryptoutil-hash_pepper.secret # SUITE pepper: shared by ALL 10 services
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
# deployments/cryptoutil-suite/compose.yml
include:
  - path: ../sm/compose.yml      # PRODUCT-level (currently includes sm-kms)
  - path: ../pki/compose.yml     # or ../pki-ca/compose.yml
  - path: ../identity/compose.yml
  - path: ../sm/compose.yml  # or ../sm-im/compose.yml
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
| Hash pepper (SUITE) | `cryptoutil-hash_pepper.secret` | `cryptoutil-hash_pepper.secret` (shared by all 10) |
| PostgreSQL URL | `{PRODUCT}-{SERVICE}-postgres_url.secret` | `sm-kms-postgres_url.secret` |
| PostgreSQL user | `{PRODUCT}-{SERVICE}-postgres_username.secret` | `jose-ja-postgres_username.secret` |
| PostgreSQL pass | `{PRODUCT}-{SERVICE}-postgres_password.secret` | `identity-authz-postgres_password.secret` |
| PostgreSQL DB | `{PRODUCT}-{SERVICE}-postgres_database.secret` | `pki-ca-postgres_database.secret` |
| Browser username | `{PRODUCT}-{SERVICE}-browser_username.secret` | `sm-im-browser_username.secret` |
| Browser password | `{PRODUCT}-{SERVICE}-browser_password.secret` | `sm-im-browser_password.secret` |
| Service username | `{PRODUCT}-{SERVICE}-service_username.secret` | `jose-ja-service_username.secret` |
| Service password | `{PRODUCT}-{SERVICE}-service_password.secret` | `pki-ca-service_password.secret` |

### 5.2 Layered Pepper Strategy

**SERVICE-only deployment** (`cd {PRODUCT}-{SERVICE} && docker compose up`):
- Each service has unique pepper: `{PRODUCT}-{SERVICE}-hash_pepper.secret`
- Example: `sm-kms` uses `sm-kms-hash_pepper.secret`, `pki-ca` uses `pki-ca-hash_pepper.secret`
- **Use case**: Maximum isolation during development/testing

**PRODUCT-level deployment** (`cd {PRODUCT} && docker compose up`):
- All services within product share pepper: `{PRODUCT}-hash_pepper.secret`
- Example: All 5 identity services (authz, idp, rp, rs, spa) share `identity-hash_pepper.secret`
- Single-service products (sm, pki, skeleton, jose): `{PRODUCT}-hash_pepper.secret` = alias to SERVICE pepper
- **Use case**: Shared SSO/federation within product boundary

**SUITE-level deployment** (`cd cryptoutil && docker compose up`):
- All 10 services across 5 products share pepper: `cryptoutil-hash_pepper.secret`
- **Use case**: Cross-product SSO, unified identity federation

**Other Secrets (NEVER shared)**:
- Unseal keys: ALWAYS `{PRODUCT}-{SERVICE}-unseal_{N}of5.secret` (unique per service)
- PostgreSQL credentials: ALWAYS `{PRODUCT}-{SERVICE}-postgres_*.secret` (unique per service)
- Browser credentials: ALWAYS `{PRODUCT}-{SERVICE}-browser_username.secret` and `{PRODUCT}-{SERVICE}-browser_password.secret` (unique per service)
- Service credentials: ALWAYS `{PRODUCT}-{SERVICE}-service_username.secret` and `{PRODUCT}-{SERVICE}-service_password.secret` (unique per service)
- **Level suffixes** (see ARCHITECTURE.md Section 12.3.3): `-SERVICEONLY`, `-PRODUCTONLY`, `-SUITEONLY` used in filename hints only

## 6. Migration Path

### Phase 1: Current State (SERVICE-only)

All 10 services deploy independently. Each has its own compose.yml with included telemetry. Secret names are NOT prefixed (e.g., `unseal_1of5.secret`).

### Phase 2: Secret Prefixing

Rename secret files to include `{PRODUCT}-{SERVICE}-` prefix:
- `unseal_1of5.secret` → `sm-kms-unseal_1of5.secret`
- Update compose.yml secret references
- Update linter validation

### Phase 3: Layered Pepper Strategy

Create layered pepper secrets:
- SERVICE-level: Each service has `{PRODUCT}-{SERVICE}-hash_pepper.secret` in its own `secrets/` directory
- PRODUCT-level: Create `deployments/{PRODUCT}/secrets/{PRODUCT}-hash_pepper.secret` for multi-service products
- SUITE-level: Create `deployments/cryptoutil-suite/secrets/cryptoutil-hash_pepper.secret` for full deployment
- Update compose.yml files to reference appropriate pepper based on deployment level

### Phase 4: Product-Level Composition

Create product directories with aggregation compose files:
- `deployments/identity/compose.yml` includes all 5 identity services
- `deployments/sm/compose.yml` → `deployments/sm-kms/compose.yml`
- Test: `cd deployments/identity && docker compose up`

### Phase 5: Suite-Level Composition

Create suite directory with full aggregation:
- `deployments/cryptoutil-suite/compose.yml` includes all products
- Test: `cd deployments/cryptoutil-suite && docker compose up`

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
