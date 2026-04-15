# Deployment Templates — Canonical File Content

**Companion to**: [target-structure.md](target-structure.md) (directory layout), [tls-structure.md](tls-structure.md) (certificate layout)

**Purpose**: Defines the exact canonical content for every file inside `deployments/` and `configs/`.
While `target-structure.md` specifies **what files exist**, this document specifies **what goes inside each file**.
While `tls-structure.md` specifies **what certificates exist**, this document specifies **how services reference them**.

**Canonical Templates**: Parameterized template files live in `api/cryptosuite-registry/templates/`.
Linters instantiate templates in-memory (substituting registry values) and compare byte-for-byte
against the actual files on disk. This document describes the templates; the templates themselves
are the single source of truth for enforcement.

**Enforcement**: `cicd-lint lint-fitness` and `cicd-lint lint-deployments` validators MUST enforce
these templates. Any deviation from the canonical templates is a blocking error.

---

## Table of Contents

- [A. Parameterization Table](#a-parameterization-table)
- [B. PS-ID Dockerfile Template](#b-ps-id-dockerfile-template)
- [C. PS-ID compose.yml Template](#c-ps-id-composeyml-template)
- [D. PS-ID Deployment Config Templates](#d-ps-id-deployment-config-templates)
- [E. PS-ID Standalone Config Template](#e-ps-id-standalone-config-template)
- [F. PS-ID Secrets Template](#f-ps-id-secrets-template)
- [G. Product compose.yml Template](#g-product-composeyml-template)
- [H. Product Dockerfile Template (Pending)](#h-product-dockerfile-template-pending)
- [I. Suite compose.yml Template](#i-suite-composeyml-template)
- [J. Suite Dockerfile Template](#j-suite-dockerfile-template)
- [K. Shared Services Templates](#k-shared-services-templates)
- [L. OTel Config Template](#l-otel-config-template)
- [M. Current Inconsistencies Inventory](#m-current-inconsistencies-inventory)
- [N. Enforcement Requirements](#n-enforcement-requirements)

---

## A. Parameterization Table

All templates use parameterized placeholders. Values for each PS-ID are defined in
`api/cryptosuite-registry/registry.yaml` and `internal/apps/tools/cicd_lint/lint_fitness/registry/registry.go`.

### A.1 Entity Parameters

| Parameter | Description | Example (sm-kms) | Example (jose-ja) |
|-----------|-------------|-------------------|----------------------------|
| `{SUITE}` | Suite name (always `cryptoutil`) | `cryptoutil` | `cryptoutil` |
| `{PS-ID}` | Product-Service identifier (kebab-case) | `sm-kms` | `jose-ja` |
| `{PS_ID}` | Underscore variant (PostgreSQL naming) | `sm_kms` | `jose_ja` |
| `{PRODUCT}` | Product name (kebab-case) | `sm` | `jose` |
| `{SERVICE}` | Service name within product | `kms` | `ja` |
| `{PRODUCT_DISPLAY_NAME}` | Human-readable product name | `Secrets Manager` | `JOSE` |
| `{SERVICE_DISPLAY_NAME}` | Human-readable service name | `Key Management Service` | `JWK Authority` |
| `{GITHUB_REPOSITORY_URL}` | Source repository URL | `https://github.com/justincranford/cryptoutil` | (same) |
| `{AUTHORS}` | Image authors | `Justin Cranford` | (same) |

### A.2 Port Parameters

| Parameter | Description | Formula | Example (sm-kms) | Example (jose-ja) |
|-----------|-------------|---------|-------------------|--------------------|
| `{SERVICE_APP_PORT_BASE}` | SERVICE-tier port block start | Per registry | `8000` | `8200` |
| `{SERVICE_APP_PORT_SQLITE_1}` | SQLite instance 1 host port | `{SERVICE_APP_PORT_BASE} + 0` | `8000` | `8200` |
| `{SERVICE_APP_PORT_SQLITE_2}` | SQLite instance 2 host port | `{SERVICE_APP_PORT_BASE} + 1` | `8001` | `8201` |
| `{SERVICE_APP_PORT_PG_1}` | PostgreSQL instance 1 host port | `{SERVICE_APP_PORT_BASE} + 2` | `8002` | `8202` |
| `{SERVICE_APP_PORT_PG_2}` | PostgreSQL instance 2 host port | `{SERVICE_APP_PORT_BASE} + 3` | `8003` | `8203` |
| `{SERVICE_PG_HOST_PORT}` | PostgreSQL container host port | Per registry | `54320` | `54322` |
| `{PRODUCT_APP_PORT_OFFSET}` | PRODUCT tier formula | `{SERVICE_APP_PORT_BASE} + 10000` | `18000` | `18200` |
| `{SUITE_APP_PORT_OFFSET}` | SUITE tier formula | `{SERVICE_APP_PORT_BASE} + 20000` | `28000` | `28200` |

### A.3 Build & Container Parameters

| Parameter | Description | Default Value |
|-----------|-------------|---------------|
| `{GO_VERSION}` | Go compiler version | `1.26.1` |
| `{ALPINE_VERSION}` | Alpine base image tag | `latest` |
| `{CGO_ENABLED}` | CGO linkage (MUST be `0`) | `0` |
| `{CONTAINER_UID}` | Non-root user ID for final container stage | `65532` |
| `{CONTAINER_GID}` | Non-root group ID for final container stage | `65532` |
| `{IMAGE_TAG}` | Docker image tag | `local` |
| `{HEALTHCHECK_INTERVAL}` | HEALTHCHECK `--interval` | `30s` |
| `{HEALTHCHECK_TIMEOUT}` | HEALTHCHECK `--timeout` | `10s` |
| `{HEALTHCHECK_START_PERIOD}` | HEALTHCHECK `--start-period` | `30s` |
| `{HEALTHCHECK_RETRIES}` | HEALTHCHECK `--retries` | `3` |

**UID/GID Security Rationale**: Running containers as a non-root user (UID 65532, GID 65532)
is a defense-in-depth measure that limits the blast radius of container escapes. If a
vulnerability allows code execution inside the container, the attacker inherits a
non-privileged UID that cannot modify system files, install packages, bind to privileged
ports, or access host resources via shared namespaces. UID 65532 is a well-known convention
for non-root container users (commonly named `nonroot` or `nobody`). Declaring UID and GID
as build ARGs serves two purposes: (1) de-duplicates the literal values across the
Dockerfile (used in `addgroup`, `adduser`, `chown`, and `USER` directives), and (2) allows
override during local debugging builds (e.g., `--build-arg CONTAINER_UID=0 --build-arg
CONTAINER_GID=0` to temporarily run as root for strace/debugging — NEVER in CI/CD or
production).

### A.4 Complete PS-ID Parameter Matrix

| PS-ID | PRODUCT | SERVICE | SERVICE_APP_PORT_BASE | SERVICE_PG_HOST_PORT | PRODUCT_DISPLAY_NAME | SERVICE_DISPLAY_NAME | PRODUCT_APP_PORT_OFFSET | SUITE_APP_PORT_OFFSET |
|-------|---------|---------|---------------------|--------------------|--------------------|--------------------|-----------------------|---------------------|
| `sm-kms` | `sm` | `kms` | `8000` | `54320` | Secrets Manager | Key Management | `18000` | `28000` |
| `sm-im` | `sm` | `im` | `8100` | `54321` | Secrets Manager | Instant Messenger | `18100` | `28100` |
| `jose-ja` | `jose` | `ja` | `8200` | `54322` | JOSE | JWK Authority | `18200` | `28200` |
| `pki-ca` | `pki` | `ca` | `8300` | `54323` | PKI | Certificate Authority | `18300` | `28300` |
| `identity-authz` | `identity` | `authz` | `8400` | `54324` | Identity | Authorization Server | `18400` | `28400` |
| `identity-idp` | `identity` | `idp` | `8500` | `54325` | Identity | Provider | `18500` | `28500` |
| `identity-rs` | `identity` | `rs` | `8600` | `54326` | Identity | Resource Server | `18600` | `28600` |
| `identity-rp` | `identity` | `rp` | `8700` | `54327` | Identity | Relying Party | `18700` | `28700` |
| `identity-spa` | `identity` | `spa` | `8800` | `54328` | Identity | Single Page App | `18800` | `28800` |
| `skeleton-template` | `skeleton` | `template` | `8900` | `54329` | Skeleton | Template | `18900` | `28900` |

---

## B. PS-ID Dockerfile Template

**ONE canonical pattern for all 10 PS-IDs.** Four stages: `validation` → `builder` → `runtime-deps` → `final`.

### B.1 Canonical Template

**Canonical file**: [`api/cryptosuite-registry/templates/deployments/__PS_ID__/Dockerfile`](../api/cryptosuite-registry/templates/deployments/__PS_ID__/Dockerfile)

The template file is the machine-readable source of truth. Linters instantiate it for each of the
10 PS-IDs by substituting all `__PS_ID__`, `__PRODUCT__`, `__SUITE__`, and numeric parameters,
then compare byte-for-byte against `deployments/{PS-ID}/Dockerfile`.

See the canonical template file for the full content. The template implements the
4-stage pattern (`validation` → `builder` → `runtime-deps` → `final`) described in
Section B.2 below.

### B.2 Template Rules (Enforceable)

| Rule ID | Rule | Rationale |
|---------|------|-----------|
| DF-01 | MUST have exactly 4 named stages: `validation`, `builder`, `runtime-deps`, `final` | Structural consistency |
| DF-02 | MUST use `ARG GO_VERSION={GO_VERSION}` | Version consistency |
| DF-03 | MUST use `ARG ALPINE_VERSION={ALPINE_VERSION}` with `# hadolint ignore=DL3007` | Security patch policy |
| DF-04 | MUST use `ARG CGO_ENABLED={CGO_ENABLED}` | CGO ban |
| DF-05 | Builder MUST use BuildKit cache mounts for `/go/pkg/mod` and `/root/.cache/go-build` | Build performance |
| DF-06 | MUST build `./cmd/{PS-ID}` and output to `/app/{PS-ID}` | Binary naming |
| DF-07 | MUST validate static linking with `ldd` check | Portability |
| DF-08 | Runtime-deps MUST install `ca-certificates`, `tzdata`, `tini` only (NO curl, NO wget) | Minimal attack surface |
| DF-09 | Final stage MUST NOT install any packages via `apk` | Minimal attack surface |
| DF-10 | MUST use `ARG CONTAINER_UID={CONTAINER_UID}` and `ARG CONTAINER_GID={CONTAINER_GID}` for user creation | Security (nonroot, parameterized) |
| DF-11 | MUST use `WORKDIR /app` (NOT `/app/run`) | Uniformity |
| DF-12 | MUST use compact multi-line `LABEL` block (NOT individual LABEL lines) | Style consistency |
| DF-13 | `LABEL org.opencontainers.image.title` MUST equal `{SUITE}-{PS-ID}` | Naming convention |
| DF-14 | MUST have `EXPOSE 8080` only (NO 9090) | Admin is 127.0.0.1-only |
| DF-15 | HEALTHCHECK MUST use parameterized timing: `--interval={HEALTHCHECK_INTERVAL} --timeout={HEALTHCHECK_TIMEOUT} --start-period={HEALTHCHECK_START_PERIOD} --retries={HEALTHCHECK_RETRIES}` | Configurable health probes |
| DF-16 | HEALTHCHECK command MUST use `/app/{PS-ID} livez \|\| exit 1` | Built-in CLI |
| DF-17 | ENTRYPOINT MUST be `["/sbin/tini", "--", "/app/{PS-ID}"]` | Signal handling |
| DF-18 | MUST end with `USER ${CONTAINER_UID}:${CONTAINER_GID}` (NOT commented out) | Security |
| DF-19 | MUST NOT set `GOMODCACHE` or `GOCACHE` env vars | Unnecessary in runtime |
| DF-20 | MUST NOT have `CMD` instruction (compose overrides command) | Compose controls |
| DF-21 | Header comment MUST reference `{SUITE}-{PS-ID}:{IMAGE_TAG}` and `{PRODUCT_DISPLAY_NAME} {SERVICE_DISPLAY_NAME}` | No copy-paste errors |
| DF-22 | `LABEL org.opencontainers.image.source` MUST equal `{GITHUB_REPOSITORY_URL}` | Repository link |
| DF-23 | `LABEL org.opencontainers.image.authors` MUST equal `{AUTHORS}` | Author attribution |
| DF-24 | `LABEL org.opencontainers.image.description` MUST equal `{PRODUCT_DISPLAY_NAME} {SERVICE_DISPLAY_NAME}` | Human-readable description |

### B.3 Template Variables Per PS-ID

| PS-ID | Binary Path | HEALTHCHECK Path | ENTRYPOINT | LABEL title |
|-------|-------------|------------------|------------|-------------|
| `sm-kms` | `/app/sm-kms` | `/app/sm-kms livez` | `["/sbin/tini", "--", "/app/sm-kms"]` | `{SUITE}-sm-kms` |
| `sm-im` | `/app/sm-im` | `/app/sm-im livez` | `["/sbin/tini", "--", "/app/sm-im"]` | `{SUITE}-sm-im` |
| `jose-ja` | `/app/jose-ja` | `/app/jose-ja livez` | `["/sbin/tini", "--", "/app/jose-ja"]` | `{SUITE}-jose-ja` |
| `pki-ca` | `/app/pki-ca` | `/app/pki-ca livez` | `["/sbin/tini", "--", "/app/pki-ca"]` | `{SUITE}-pki-ca` |
| `identity-authz` | `/app/identity-authz` | `/app/identity-authz livez` | `["/sbin/tini", "--", "/app/identity-authz"]` | `{SUITE}-identity-authz` |
| `identity-idp` | `/app/identity-idp` | `/app/identity-idp livez` | `["/sbin/tini", "--", "/app/identity-idp"]` | `{SUITE}-identity-idp` |
| `identity-rs` | `/app/identity-rs` | `/app/identity-rs livez` | `["/sbin/tini", "--", "/app/identity-rs"]` | `{SUITE}-identity-rs` |
| `identity-rp` | `/app/identity-rp` | `/app/identity-rp livez` | `["/sbin/tini", "--", "/app/identity-rp"]` | `{SUITE}-identity-rp` |
| `identity-spa` | `/app/identity-spa` | `/app/identity-spa livez` | `["/sbin/tini", "--", "/app/identity-spa"]` | `{SUITE}-identity-spa` |
| `skeleton-template` | `/app/skeleton-template` | `/app/skeleton-template livez` | `["/sbin/tini", "--", "/app/skeleton-template"]` | `{SUITE}-skeleton-template` |

---

## C. PS-ID compose.yml Template

**ONE canonical pattern for all 10 PS-IDs.** Four app instances + support services.

### C.1 Canonical Template

**Canonical file**: [`api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml`](../api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml)

The template file is the machine-readable source of truth. Linters instantiate it for each of the
10 PS-IDs by substituting all `__PS_ID__`, `__PRODUCT__`, `__SUITE__`, and numeric parameters,
then compare byte-for-byte against `deployments/{PS-ID}/compose.yml`.

### C.2 Compose Rules (Enforceable)

| Rule ID | Rule | Rationale |
|---------|------|-----------|
| CO-01 | Header MUST include `$schema` reference and PS-ID description | Documentation |
| CO-02 | MUST include `../shared-telemetry/compose.yml` and `../shared-postgres/compose.yml` | Required infrastructure |
| CO-03 | MUST have `healthcheck-secrets` service listing all 14 secrets with validation (`test -s` per file, exit 1 on failure) | Secret validation |
| CO-04 | MUST have `builder-{PS-ID}` service with `image: {SUITE}-{PS-ID}:{IMAGE_TAG}` and build context `../..` | Image building |
| CO-05 | MUST have `pki-init` service with command `["{PS-ID}", "/certs"]` (positional args: tier-id then target-dir) | TLS bootstrap |
| CO-06 | MUST have exactly 4 app instances: sqlite-1, sqlite-2, postgresql-1, postgresql-2 | Cross-DB testing |
| CO-07 | App service names MUST follow `{PS-ID}-app-{variant}` pattern | Naming convention |
| CO-08 | Container port MUST always be `8080` | Standardized internal port |
| CO-09 | Host ports MUST follow port formula: sqlite-1=+0, sqlite-2=+1, pg-1=+2, pg-2=+3 | Port consistency |
| CO-10 | Command MUST include: `server`, `--bind-public-port=8080`, `--config=...` args | Startup parameters |
| CO-11 | Config volume mount order: instance-specific, common, otel | Priority ordering |
| CO-12 | Healthcheck MUST use `["CMD", "/app/{PS-ID}", "livez"]` (built-in CLI, NOT wget) | Built-in healthcheck |
| CO-13 | Resource limits: 256M limit, 128M reservation | Resource control |
| CO-14 | Networks: `{PS-ID}-network` + `telemetry-network`; PostgreSQL instances add `postgres-network` | Network isolation |
| CO-15 | `working_dir: /tmp` on all app services | Writable temp dir |
| CO-16 | All 14 secrets MUST be declared in `secrets:` section with `file: ./secrets/` relative paths | Docker secrets |
| CO-17 | SQLite instances MUST mount 10 secrets (5 unseal + hash-pepper + 2 browser + 2 service); PostgreSQL instances add 4 postgres secrets (14 total) | Minimal secrets per variant |
| CO-18 | PostgreSQL instance 2 MUST depend on instance 1 `service_healthy` | Schema init ordering |
| CO-19 | Healthcheck timing MUST use parameterized values: `start-period: {HEALTHCHECK_START_PERIOD}`, `interval: {HEALTHCHECK_INTERVAL}`, `timeout: {HEALTHCHECK_TIMEOUT}`, `retries: {HEALTHCHECK_RETRIES}` | Configurable health probes |
| CO-20 | All `image:` references MUST use `{SUITE}-{PS-ID}:{IMAGE_TAG}` (NOT hardcoded suite name or tag) | Parameterized naming |

---

## D. PS-ID Deployment Config Templates

Five deployment config files per PS-ID, ALL using **kebab-case** YAML keys.

### D.1 `{PS-ID}-app-common.yml` (Shared Settings)

**Canonical file**: [`api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-common.yml`](../api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-common.yml)

### D.2 `{PS-ID}-app-sqlite-1.yml`

**Canonical file**: [`api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-sqlite-1.yml`](../api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-sqlite-1.yml)

### D.3 `{PS-ID}-app-sqlite-2.yml`

**Canonical file**: [`api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-sqlite-2.yml`](../api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-sqlite-2.yml)

### D.4 `{PS-ID}-app-postgresql-1.yml`

**Canonical file**: [`api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-postgresql-1.yml`](../api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-postgresql-1.yml)

### D.5 `{PS-ID}-app-postgresql-2.yml`

**Canonical file**: [`api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-postgresql-2.yml`](../api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-postgresql-2.yml)

All five config templates are the machine-readable source of truth. Linters instantiate them for each
of the 10 PS-IDs by substituting port numbers and PS-ID placeholders, then compare byte-for-byte
against `deployments/{PS-ID}/config/{PS-ID}-app-framework-{variant}.yml`.

### D.6 Config File Rules (Enforceable)

| Rule ID | Rule | Rationale |
|---------|------|-----------|
| CF-01 | ALL keys MUST be kebab-case (NEVER snake_case) | ENG-HANDBOOK §13.2 |
| CF-02 | Common file MUST set `bind-public-address: "0.0.0.0"` | Container networking |
| CF-03 | Common file MUST reference TLS cert paths from pki-init | TLS bootstrap |
| CF-04 | Common file MUST reference all 5 unseal secret paths | Barrier service |
| CF-05 | Common file MUST reference browser/service credential secret paths | Authentication |
| CF-06 | Instance files MUST set `cors-origins` with correct port | CORS correctness |
| CF-07 | Instance files MUST set `otlp-service` = `{PS-ID}-{variant}` | Telemetry identity |
| CF-08 | Instance files MUST set `otlp-hostname` = `{PS-ID}-{variant}` | Telemetry identity |
| CF-09 | SQLite instance files MUST set `database-url: "sqlite://file::memory:?cache=shared"` | Database config |
| CF-10 | PostgreSQL instance files MUST NOT set `database-url` (passed via compose command) | Database config |
| CF-11 | Instance files MUST NOT duplicate keys from common file | DRY principle |
| CF-12 | Header comment MUST reference `{PS-ID}`, NOT another PS-ID | No copy-paste errors |

---

## E. PS-ID Standalone Config Template

Each PS-ID has a standalone config at `configs/{PS-ID}/{PS-ID}.yml` for local development.

**Canonical file**: [`api/cryptosuite-registry/templates/configs/__PS_ID__/__PS_ID__-framework.yml`](../api/cryptosuite-registry/templates/configs/__PS_ID__/__PS_ID__-framework.yml)

The template file is the machine-readable source of truth. Linters instantiate it for each of the
10 PS-IDs by substituting port numbers and PS-ID placeholders, then compare byte-for-byte
against `configs/{PS-ID}/{PS-ID}-framework.yml`.

### E.1 Standalone Config Rules

| Rule ID | Rule | Rationale |
|---------|------|-----------|
| SC-01 | ALL keys MUST be kebab-case | ENG-HANDBOOK §13.2 |
| SC-02 | `bind-public-address` MUST be `127.0.0.1` (NOT `0.0.0.0`) | Windows firewall prevention |
| SC-03 | `bind-public-port` MUST equal `{SERVICE_APP_PORT_BASE}` from registry | Port consistency |
| SC-04 | `bind-admin-port` MUST be `9090` | Admin port standardization |
| SC-05 | `otlp-service` MUST equal `{PS-ID}` | Telemetry naming |
| SC-06 | Header comment MUST reference `{PRODUCT_DISPLAY_NAME} {SERVICE_DISPLAY_NAME}` and `{PS-ID}` correctly | No copy-paste errors |

---

## F. PS-ID Secrets Template

14 secret files per PS-ID in `deployments/{PS-ID}/secrets/`.

| Filename | Value Pattern | Notes |
|----------|---------------|-------|
| `unseal-1of5.secret` | `{PS-ID}-unseal-key-1-of-5-{base64-random-32-bytes}` | Unique per shard |
| `unseal-2of5.secret` | `{PS-ID}-unseal-key-2-of-5-{base64-random-32-bytes}` | Unique per shard |
| `unseal-3of5.secret` | `{PS-ID}-unseal-key-3-of-5-{base64-random-32-bytes}` | Unique per shard |
| `unseal-4of5.secret` | `{PS-ID}-unseal-key-4-of-5-{base64-random-32-bytes}` | Unique per shard |
| `unseal-5of5.secret` | `{PS-ID}-unseal-key-5-of-5-{base64-random-32-bytes}` | Unique per shard |
| `hash-pepper-v3.secret` | `{PS-ID}-hash-pepper-v3-{base64-random-32-bytes}` | Hash pepper |
| `postgres-url.secret` | `postgres://{PS_ID}_database_user:{PS_ID}_database_pass@shared-postgres-leader:5432/{PS_ID}_database?sslmode=disable` | Full DSN (v12: `sslmode=verify-full` + `sslrootcert`/`sslcert`/`sslkey` params) |
| `postgres-username.secret` | `{PS_ID}_database_user` | PostgreSQL user |
| `postgres-password.secret` | `{PS_ID}_database_pass-{base64-random-32-bytes}` | PostgreSQL password |
| `postgres-database.secret` | `{PS_ID}_database` | PostgreSQL database |
| `browser-username.secret` | `{PS-ID}-browser-user` | Browser auth user |
| `browser-password.secret` | `{PS-ID}-browser-pass-{base64-random-32-bytes}` | Browser auth password |
| `service-username.secret` | `{PS-ID}-service-user` | Service auth user |
| `service-password.secret` | `{PS-ID}-service-pass-{base64-random-32-bytes}` | Service auth password |

---

## G. Product compose.yml Template

**ONE canonical pattern for all 5 products.** Includes child PS-IDs, overrides ports.

**Canonical file**: [`api/cryptosuite-registry/templates/deployments/__PRODUCT__/compose.yml`](../api/cryptosuite-registry/templates/deployments/__PRODUCT__/compose.yml)

The template file is the machine-readable source of truth. Linters instantiate it for each of the
5 products by substituting product names, PS-ID include lists, and port offsets, then compare
byte-for-byte against `deployments/{PRODUCT}/compose.yml`.

### G.1 Product Compose Rules

| Rule ID | Rule | Rationale |
|---------|------|-----------|
| PC-01 | MUST include all child PS-ID compose files | Recursive architecture |
| PC-02 | MUST override `pki-init` command to `["{PRODUCT}", "/certs"]` (positional args: product tier-id then target-dir) | Product-scoped certs |
| PC-03 | Port overrides MUST use `!override` tag | Complete port replacement |
| PC-04 | Port formula: `SERVICE + 10000` | Product tier offset |
| PC-05 | Unseal secrets MUST use `{PRODUCT}-unseal-key-N-of-5-...` values | Product-scoped encryption |
| PC-06 | MUST include `browser-*.secret.never` and `service-*.secret.never` marker files | Credential scope |

---

## H. Product Dockerfile Template (Pending)

**Status**: Missing for all 5 products. Target-structure.md Section N marks this as CREATE pending.

Product Dockerfiles are unnecessary when each PS-ID builds its own binary. If the
architecture migrates to a single suite binary (`./cmd/cryptoutil`), product Dockerfiles
become needed. This is deferred until the suite binary migration decision is finalized.

---

## I. Suite compose.yml Template

**Canonical file**: [`api/cryptosuite-registry/templates/deployments/__SUITE__/compose.yml`](../api/cryptosuite-registry/templates/deployments/__SUITE__/compose.yml)

The template file is the machine-readable source of truth. Linters instantiate it for the suite
(`cryptoutil`) by substituting product include lists and port offsets (+20000), then compare
byte-for-byte against `deployments/cryptoutil/compose.yml`.

### I.1 Suite Compose Rules

| Rule ID | Rule | Rationale |
|---------|------|-----------|
| SU-01 | MUST include all 5 product compose files | Complete suite |
| SU-02 | Port formula: `SERVICE + 20000` | Suite tier offset |
| SU-03 | Compact inline port override syntax for 40 services | Readability |
| SU-04 | Unseal secrets MUST use `cryptoutil-unseal-key-N-of-5-...` values | Suite-scoped encryption |

---

## J. Suite Dockerfile Template

The suite Dockerfile builds the `{SUITE}` binary that can run any service via subcommands.

**Canonical file**: [`api/cryptosuite-registry/templates/deployments/__SUITE__/Dockerfile`](../api/cryptosuite-registry/templates/deployments/__SUITE__/Dockerfile)

The template follows the same 4-stage pattern as the PS-ID Dockerfile (Section B), substituting
the suite name (`cryptoutil`) for `{PS-ID}` in all binary paths, LABEL fields, HEALTHCHECK, and
ENTRYPOINT directives.

---

## K. Shared Services Templates

### K.1 shared-telemetry/compose.yml

Provides `opentelemetry-collector-contrib` + `grafana-otel-lgtm` for all services.

**Key rules**:
- Network: `telemetry-network`
- OTel collector ports: 4317 (gRPC), 4318 (HTTP)
- Grafana UI: 3000
- Healthcheck for OTel: `healthcheck-opentelemetry-collector-contrib` service

### K.2 shared-postgres/compose.yml

Provides PostgreSQL leader+follower for all services.

**Key rules**:
- Network: `postgres-network`
- Leader service: `postgres-leader`
- Init scripts in `docker-entrypoint-initdb.d/` create per-PS-ID databases
- Each PS-ID gets its own database created by init script

---

## L. OTel Config Template

`deployments/shared-telemetry/otel/cryptoutil-otel.yml`:

```yaml
# OpenTelemetry configuration
otlp: true
otlp-endpoint: http://opentelemetry-collector-contrib:4318
otlp-version: "0.0.1"
otlp-environment: "docker compose"
```

ALL keys MUST be kebab-case.

**V12 planned change**: `otlp-endpoint` will change to `https://` with mTLS when framework-v12 wires OTel TLS. See [`docs/framework-v12/plan.md`](framework-v12/plan.md) Phase 4.

---

## M. Current Inconsistencies Inventory

### M.1 Dockerfile Inconsistencies (CRITICAL)

Three fundamentally different Dockerfile patterns exist where there MUST be exactly one:

| Category | Affected PS-IDs | Deviation from Template |
|----------|----------------|------------------------|
| **Pattern A** (sm-kms style) | sm-kms, identity-authz, identity-idp, identity-rp, identity-rs | 4-stage but: `WORKDIR /app/run`, `GOMODCACHE`/`GOCACHE` env vars, `curl` installed in final, `USER` commented out, individual `LABEL` lines |
| **Pattern B** (jose-ja style) | jose-ja, pki-ca, skeleton-template | 3-stage (no `runtime-deps`): `adduser`-based user creation, compact `LABEL`, `CMD` with config path |
| **Pattern C** (sm-im) | sm-im | 2-stage (no `validation`): user `1000:1000` (WRONG), no BuildKit caches, no static link check |

### M.2 Specific Bugs

| PS-ID | Bug | Impact |
|-------|-----|--------|
| `skeleton-template` | Header says "JOSE Authority Server" | Documentation error |
| `skeleton-template` | Username is `jose`, dirs are `/etc/jose` | Wrong identity |
| `skeleton-template` | CMD uses `--config=/etc/jose/jose.yml` | Wrong config path |
| `identity-spa` | Builder builds `/app/identity-spa` but runtime COPY copies `/app/cryptoutil` | **Runtime failure** |
| `sm-im` | User UID:GID is `1000:1000` (should be `65532:65532`) | Security deviation |
| `sm-im` | No validation stage | Missing build arg validation |
| `sm-im` | No BuildKit cache mounts | Slow builds |
| `sm-im` | No static linking verification | Portability risk |
| All Pattern A | `USER 65532:65532` commented out | Running as root |
| All Pattern A | `curl` installed in final stage | Unnecessary attack surface |
| `cryptoutil` (suite) | No tini installed/copied | Missing signal handling |

### M.3 Config Key Naming Inconsistencies

| Convention | PS-IDs | Status |
|------------|--------|--------|
| **kebab-case** (CORRECT) | jose-ja, pki-ca, skeleton-template | Correct |
| **snake_case** (WRONG) | sm-kms, sm-im, identity-authz, identity-idp, identity-rp, identity-rs, identity-spa | Needs migration |

### M.4 Config Content Inconsistencies

| Category | PS-IDs Affected | Issue |
|----------|----------------|-------|
| Deployment common config structure | sm-im | Missing TLS cert paths, missing unseal config, only has bind + credentials |
| Deployment instance config structure | sm-im | Missing `cors-origins`, missing `otlp-hostname`, only has `otlp-service` |
| Deployment instance config structure | jose-ja, skeleton-template | Duplicates common settings (security-headers, rate-limiting) in every instance file |
| Standalone config content | skeleton-template | Header says "JOSE Authority Server", OTLP service says "skeleton-template-ja" |
| Standalone config content | sm-kms, sm-im | Uses snake_case keys (bind_address, max_open_conns, etc.) |
| Standalone config admin port | jose-ja, skeleton-template | Uses `bind-admin-port: 9092` (should be `9090`) |

### M.5 Compose Inconsistencies

Most compose files are consistent. Known issues:

| Category | Issue |
|----------|-------|
| Healthcheck timing | Some services use `retries: 10`, others `retries: 5` for PostgreSQL |
| Builder wait message | Some use "Build completed successfully", others "Build complete" |

---

## N. Enforcement Requirements

### N.1 Template-Comparison Linters (PRIMARY Enforcement Strategy)

**Architecture Decision**: All canonical templates are stored as parameterized template files in
`api/cryptosuite-registry/templates/`. Linters instantiate templates in-memory by substituting
registry values from `api/cryptosuite-registry/registry.yaml`, then compare the result
byte-for-byte against the actual file on disk. Any deviation is a linting error.

| Linter Name | Template File | Target Files | Comparison |
|-------------|---------------|--------------|------------|
| `template-compliance` | `api/cryptosuite-registry/templates/deployments/__PS_ID__/Dockerfile` | `deployments/{PS-ID}/Dockerfile` (×10) | Full byte-for-byte |
| `template-compliance` | `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml` | `deployments/{PS-ID}/compose.yml` (×10) | Full byte-for-byte |
| `template-compliance` | `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-common.yml` | `deployments/{PS-ID}/config/{PS-ID}-app-framework-common.yml` (×10) | Full byte-for-byte |
| `template-compliance` | `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-sqlite-{1,2}.yml` | `deployments/{PS-ID}/config/{PS-ID}-app-framework-sqlite-{1,2}.yml` (×20) | Full byte-for-byte |
| `template-compliance` | `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-postgresql-{1,2}.yml` | `deployments/{PS-ID}/config/{PS-ID}-app-framework-postgresql-{1,2}.yml` (×20) | Full byte-for-byte |
| `template-compliance` | `api/cryptosuite-registry/templates/configs/__PS_ID__/__PS_ID__-framework.yml` | `configs/{PS-ID}/{PS-ID}-framework.yml` (×10) | Full byte-for-byte |
| `template-compliance` | `api/cryptosuite-registry/templates/deployments/__PRODUCT__/compose.yml` | `deployments/{PRODUCT}/compose.yml` (×5) | Full byte-for-byte |
| `template-compliance` | `api/cryptosuite-registry/templates/deployments/__SUITE__/Dockerfile` | `deployments/cryptoutil/Dockerfile` (×1) | Full byte-for-byte |
| `template-compliance` | `api/cryptosuite-registry/templates/deployments/__SUITE__/compose.yml` | `deployments/cryptoutil/compose.yml` (×1) | Full byte-for-byte |
| `secrets-compliance` | N/A (validation-only) | `deployments/{PS-ID}/secrets/*.secret` (×140) | File count + naming pattern |

**Instantiation Process**:
1. Load `registry.yaml` → iterate `product-services`
2. For each PS-ID, compute all parameters from Section A (A.1-A.4)
3. Load template file, substitute `{PARAMETER}` placeholders with computed values
4. Read corresponding actual file from disk
5. Compare byte-for-byte — any difference is a BLOCKING error with unified diff output

**Benefits over rule-based linters**:
- Single source of truth: template file IS the specification (no drift between docs and linter)
- Complete coverage: checks EVERY byte, not just specific patterns
- Easy maintenance: update template file → all 10 PS-IDs automatically validated
- Clear error messages: unified diff shows exactly what differs

### N.2 Rule-Based Fitness Linters (SUPPLEMENTARY)

Rule-based linters remain useful for cross-cutting concerns not captured in templates:

| Linter Name | Scope | Purpose |
|-------------|-------|---------|
| `config_key_naming` | `configs/**/*.yml`, `deployments/*/config/*.yml` | Enforce all YAML keys are kebab-case |
| `config_header_identity` | `configs/**/*.yml`, `deployments/*/config/*.yml` | Verify header comments reference correct PS-ID |
| `config_instance_minimal` | `deployments/*/config/*-{variant}.yml` | Verify instance configs only contain cors-origins, otlp-service, otlp-hostname, database-url |
| `config_common_complete` | `deployments/*/config/*-common.yml` | Verify common configs contain all required shared keys |

**Note**: Dockerfile-specific rule linters (`dockerfile_structure`, `dockerfile_binary_name`, etc.) are
superseded by `template_dockerfile` (N.1). They MAY be retained as fast-fail checks but the
template-comparison linter is the authoritative enforcement mechanism.

### N.3 Existing Linter Enhancements

| Linter | Enhancement |
|--------|-------------|
| `compose_service_names` | Add rule: exactly 4 app instances per PS-ID |
| `standalone_config_presence` | Add rule: validate key naming is kebab-case |

### N.4 Enforcement Priority

1. **P0 (BLOCKING)**: Fix identity-spa COPY bug (runtime failure)
2. **P0 (BLOCKING)**: Fix skeleton-template Dockerfile (copy-paste from jose-ja)
3. **P1 (HIGH)**: Create canonical template files in `api/cryptosuite-registry/templates/`
4. **P1 (HIGH)**: Standardize all 10 Dockerfiles to canonical template
5. **P1 (HIGH)**: Enable `USER ${CONTAINER_UID}:${CONTAINER_GID}` in all Dockerfiles (currently commented out in 5)
6. **P1 (HIGH)**: Implement template-comparison linters (N.1)
7. **P2 (MEDIUM)**: Migrate snake_case configs to kebab-case (7 services)
8. **P2 (MEDIUM)**: Standardize deployment config overlay structure
9. **P3 (LOW)**: Fix suite Dockerfile (add tini)

---

## O. Canonical Template Architecture

### O.1 Overview

**Architecture Decision**: All canonical templates live as parameterized files in
`api/cryptosuite-registry/templates/`. This document (deployment-templates.md)
describes the templates and their rules; the `templates/` directory IS the
machine-readable, linter-consumable source of truth.

### O.2 Template File Catalog

All template files live in `api/cryptosuite-registry/templates/`. Paths below are relative to that directory.

| Template File | Section | Purpose | Instantiation Count |
|---------------|---------|---------|---------------------|
| `deployments/__PS_ID__/Dockerfile` | B | PS-ID Dockerfile | ×10 (one per PS-ID) |
| `deployments/__PS_ID__/compose.yml` | C | PS-ID compose | ×10 |
| `deployments/__PS_ID__/config/__PS_ID__-app-framework-common.yml` | D.1 | Deployment common config | ×10 |
| `deployments/__PS_ID__/config/__PS_ID__-app-framework-sqlite-1.yml` | D.2 | Deployment SQLite instance 1 config | ×10 |
| `deployments/__PS_ID__/config/__PS_ID__-app-framework-sqlite-2.yml` | D.3 | Deployment SQLite instance 2 config | ×10 |
| `deployments/__PS_ID__/config/__PS_ID__-app-framework-postgresql-1.yml` | D.4 | Deployment PostgreSQL instance 1 config | ×10 |
| `deployments/__PS_ID__/config/__PS_ID__-app-framework-postgresql-2.yml` | D.5 | Deployment PostgreSQL instance 2 config | ×10 |
| `configs/__PS_ID__/__PS_ID__-framework.yml` | E | Standalone framework config | ×10 |
| `deployments/__PRODUCT__/compose.yml` | G | Product compose | ×5 (one per product) |
| `deployments/__SUITE__/Dockerfile` | J | Suite Dockerfile | ×1 |
| `deployments/__SUITE__/compose.yml` | I | Suite compose | ×1 |
| `deployments/__PS_ID__/secrets/` | F | PS-ID secrets (validation only) | ×10 directories |
| `deployments/__PRODUCT__/secrets/` | — | Product secrets (validation only) | ×5 directories |
| `deployments/__SUITE__/secrets/` | — | Suite secrets (validation only) | ×1 directory |

### O.3 Template Syntax

Templates use `__PARAMETER__` placeholders (double underscores, ALL CAPS with underscores).
Parameters are resolved from `registry.yaml` and `internal/apps/tools/cicd_lint/lint_fitness/registry/registry.go`.

```
# Example template line:
LABEL org.opencontainers.image.title="__SUITE__-__PS_ID__" \
      org.opencontainers.image.source="__GITHUB_REPOSITORY_URL__" \
      org.opencontainers.image.authors="__AUTHORS__" \
      org.opencontainers.image.description="__PRODUCT_DISPLAY_NAME__ __SERVICE_DISPLAY_NAME__"

# After instantiation for jose-ja:
LABEL org.opencontainers.image.title="cryptoutil-jose-ja" \
      org.opencontainers.image.source="https://github.com/justincranford/cryptoutil" \
      org.opencontainers.image.authors="Justin Cranford" \
      org.opencontainers.image.description="JOSE JWK Authority"
```

### O.4 Relationship Between Documents

```
registry.yaml                                 → PS-ID definitions, port assignments, product groupings
  ↓
api/cryptosuite-registry/templates/           → Parameterized canonical content (machine source of truth)
  ↓
deployment-templates.md                       → Human-readable documentation of templates and rules
  ↓
lint-fitness template-compliance              → Instantiate templates, compare to disk, report deviations
lint-fitness secrets-compliance               → Validate secrets directory structure
```

---

## P. Cross-References

| Topic | Reference |
|-------|-----------|
| Directory layout (what files exist) | [target-structure.md](target-structure.md) |
| TLS certificate layout | [tls-structure.md](tls-structure.md) |
| Port assignments | [ENG-HANDBOOK.md §3.4](ENG-HANDBOOK.md#34-port-assignments--networking) |
| Deployment architecture | [ENG-HANDBOOK.md §12](ENG-HANDBOOK.md#12-deployment-architecture) |
| Config file architecture | [ENG-HANDBOOK.md §13.2](ENG-HANDBOOK.md#132-config-file-architecture) |
| Secrets management | [ENG-HANDBOOK.md §13.3](ENG-HANDBOOK.md#133-secrets-management-in-deployments) |
| Docker Compose rules | [04-01.deployment.instructions.md](../.github/instructions/04-01.deployment.instructions.md) |
| Secret naming conventions | [target-structure.md §L](target-structure.md#l-secret-file-naming-convention) |
| Entity registry | [registry.yaml](../api/cryptosuite-registry/registry.yaml) |
| Canonical templates | [api/cryptosuite-registry/templates/](../api/cryptosuite-registry/templates/) |
