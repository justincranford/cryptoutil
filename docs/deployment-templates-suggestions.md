# deployment-templates.md → ENG-HANDBOOK.md Suggestions

## Executive Summary

Analysis of [deployment-templates.md](deployment-templates.md) against [ENG-HANDBOOK.md §12–§13.6](ENG-HANDBOOK.md#12-deployment-architecture) reveals that the handbook covers high-level template enforcement strategy (§13.6), deployment patterns (§12.3), and config file architecture (§13.2), but is missing detailed parameterization tables, enforceable rule catalogs, known inconsistency inventory, and shared infrastructure wiring details. The following additions are suggested.

1. [Complete Parameterization Table](#1-complete-parameterization-table) — entity display names, port formulas, build/container parameters, and the full PS-ID parameter matrix are absent from the handbook.
2. [Container UID/GID Security Rationale](#2-container-uidgid-security-rationale) — the rationale for UID 65532, the ARG parameterization pattern, and the debug override procedure are not documented.
3. [Dockerfile Enforceable Rules (DF-01–DF-24)](#3-dockerfile-enforceable-rules-df-01df-24) — 24 machine-checkable Dockerfile rules with rationale are only in deployment-templates.md.
4. [Compose Enforceable Rules (CO-01–CO-22)](#4-compose-enforceable-rules-co-01co-22) — 22 compose rules including the named-volume mandate (CO-21/CO-22) are not in the handbook.
5. [Deployment Config File Rules (CF-01–CF-17)](#5-deployment-config-file-rules-cf-01cf-17) — 17 config overlay rules including the PostgreSQL mTLS cert path rules (CF-13–CF-17) are absent.
6. [Standalone Config Rules (SC-01–SC-06)](#6-standalone-config-rules-sc-01sc-06) — 6 standalone config rules including the 127.0.0.1 bind address requirement are not stated.
7. [Product and Suite Compose Rules](#7-product-and-suite-compose-rules) — PC-01–PC-06 (product) and SU-01–SU-04 (suite) compose rules are absent.
8. [Secret File Value Patterns](#8-secret-file-value-patterns) — the complete table of 14 secret filenames with their exact value format patterns is not in the handbook.
9. [Shared PostgreSQL mTLS Cert Reference Table](#9-shared-postgresql-mtls-cert-reference-table) — which PKI category each PostgreSQL node uses, and logical cert ownership per node, is not documented.
10. [Current Inconsistency Inventory](#10-current-inconsistency-inventory) — the three Dockerfile divergence patterns (A/B/C), specific per-PS-ID bugs, and config/compose inconsistencies are exclusively in deployment-templates.md.
11. [Template Syntax Specification](#11-template-syntax-specification) — the `__KEY__` placeholder format, the rationale for double-underscore delimiters, and the template file catalog are not fully captured.

---

## Details

### 1. Complete Parameterization Table

**Current state in ENG-HANDBOOK.md**: §13.6 mentions placeholder substitution and lists `__PS_ID__`, `__PUBLIC_PORT__` as examples. §3.4 has the port table. No unified parameterization reference exists.

**Suggested addition to §13.6 or new §13.7**:

#### Entity Parameters

| Parameter | Description | Example (sm-kms) | Example (jose-ja) |
|-----------|-------------|------------------|-------------------|
| `{SUITE}` | Suite name | `cryptoutil` | `cryptoutil` |
| `{PS-ID}` | Product-Service ID (kebab-case) | `sm-kms` | `jose-ja` |
| `{PS_ID}` | Underscore variant (SQL/secrets) | `sm_kms` | `jose_ja` |
| `{PRODUCT}` | Product name | `sm` | `jose` |
| `{SERVICE}` | Service name within product | `kms` | `ja` |
| `{PRODUCT_DISPLAY_NAME}` | Human-readable product name | `Secrets Manager` | `JOSE` |
| `{SERVICE_DISPLAY_NAME}` | Human-readable service name | `Key Management Service` | `JWK Authority` |
| `{GITHUB_REPOSITORY_URL}` | Source repo URL | `https://github.com/justincranford/cryptoutil` | (same) |
| `{AUTHORS}` | Image authors | `Justin Cranford` | (same) |

#### Port Parameters

| Parameter | Formula | Example (sm-kms) | Example (jose-ja) |
|-----------|---------|-----------------|------------------|
| `{SERVICE_APP_PORT_SQLITE_1}` | `base + 0` | `8000` | `8200` |
| `{SERVICE_APP_PORT_SQLITE_2}` | `base + 1` | `8001` | `8201` |
| `{SERVICE_APP_PORT_PG_1}` | `base + 2` | `8002` | `8202` |
| `{SERVICE_APP_PORT_PG_2}` | `base + 3` | `8003` | `8203` |
| `{SERVICE_PG_HOST_PORT}` | Per registry | `54320` | `54322` |
| `{PRODUCT_APP_PORT_OFFSET}` | `base + 10000` | `18000` | `18200` |
| `{SUITE_APP_PORT_OFFSET}` | `base + 20000` | `28000` | `28200` |

#### Build/Container Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `{GO_VERSION}` | Go compiler version | `1.26.1` |
| `{ALPINE_VERSION}` | Alpine base image tag | `latest` (+ hadolint DL3007 ignore) |
| `{CGO_ENABLED}` | CGO linkage (MUST be `0`) | `0` |
| `{CONTAINER_UID}` | Non-root user ID | `65532` |
| `{CONTAINER_GID}` | Non-root group ID | `65532` |
| `{IMAGE_TAG}` | Docker image tag | `local` |
| `{HEALTHCHECK_INTERVAL}` | HEALTHCHECK `--interval` | `30s` |
| `{HEALTHCHECK_TIMEOUT}` | HEALTHCHECK `--timeout` | `10s` |
| `{HEALTHCHECK_START_PERIOD}` | HEALTHCHECK `--start-period` | `30s` |
| `{HEALTHCHECK_RETRIES}` | HEALTHCHECK `--retries` | `3` |

#### Complete PS-ID Parameter Matrix

| PS-ID | PORT_BASE | PG_PORT | PRODUCT_APP_PORT | SUITE_APP_PORT |
|-------|-----------|---------|-----------------|----------------|
| `sm-kms` | `8000` | `54320` | `18000` | `28000` |
| `sm-im` | `8100` | `54321` | `18100` | `28100` |
| `jose-ja` | `8200` | `54322` | `18200` | `28200` |
| `pki-ca` | `8300` | `54323` | `18300` | `28300` |
| `identity-authz` | `8400` | `54324` | `18400` | `28400` |
| `identity-idp` | `8500` | `54325` | `18500` | `28500` |
| `identity-rs` | `8600` | `54326` | `18600` | `28600` |
| `identity-rp` | `8700` | `54327` | `18700` | `28700` |
| `identity-spa` | `8800` | `54328` | `18800` | `28800` |
| `skeleton-template` | `8900` | `54329` | `18900` | `28900` |

---

### 2. Container UID/GID Security Rationale

**Current state in ENG-HANDBOOK.md**: Non-root container user is mentioned in passing but the rationale and parameterization strategy are not documented.

**Suggested addition to §12.2.1**:

> **Container UID/GID 65532**: Running containers as non-root (UID 65532, GID 65532) limits blast radius of container escapes — an attacker inheriting this UID cannot modify system files, install packages, bind to privileged ports, or access host resources via shared namespaces. UID 65532 is a widely adopted convention (commonly named `nonroot` or `nobody`).
>
> **Parameterized as ARG** for two reasons: (1) de-duplicates the literal value across `addgroup`, `adduser`, `chown`, and `USER` directives in the Dockerfile; (2) allows override during local debugging (`--build-arg CONTAINER_UID=0 --build-arg CONTAINER_GID=0` to run as root for strace/debugging — NEVER in CI/CD or production).

---

### 3. Dockerfile Enforceable Rules (DF-01–DF-24)

**Current state in ENG-HANDBOOK.md**: §12.2.1 describes the 4-stage pattern conceptually. No machine-checkable rule catalog exists.

**Suggested addition to §12.2.1**:

| Rule | Requirement | Rationale |
|------|-------------|-----------|
| DF-01 | Exactly 4 named stages: `validation`, `builder`, `runtime-deps`, `final` | Structural consistency |
| DF-02 | `ARG GO_VERSION={GO_VERSION}` | Version consistency |
| DF-03 | `ARG ALPINE_VERSION={ALPINE_VERSION}` with `# hadolint ignore=DL3007` | Security patch policy |
| DF-04 | `ARG CGO_ENABLED=0` | CGO ban |
| DF-05 | BuildKit cache mounts for `/go/pkg/mod` and `/root/.cache/go-build` | Build performance |
| DF-06 | Build `./cmd/{PS-ID}`, output to `/app/{PS-ID}` | Binary naming |
| DF-07 | Validate static linking with `ldd` check | Portability |
| DF-08 | `runtime-deps` installs `ca-certificates`, `tzdata`, `tini` ONLY (NO curl/wget) | Minimal attack surface |
| DF-09 | `final` stage MUST NOT install packages via `apk` | Minimal attack surface |
| DF-10 | `ARG CONTAINER_UID` and `ARG CONTAINER_GID` for user creation | Security (parameterized nonroot) |
| DF-11 | `WORKDIR /app` (NOT `/app/run`) | Uniformity |
| DF-12 | Compact multi-line `LABEL` block (NOT individual LABEL lines) | Style consistency |
| DF-13 | `LABEL org.opencontainers.image.title` = `{SUITE}-{PS-ID}` | Naming convention |
| DF-14 | `EXPOSE 8080` only (NO 9090) | Admin is 127.0.0.1-only |
| DF-15 | HEALTHCHECK uses parameterized timing ARGs | Configurable health probes |
| DF-16 | HEALTHCHECK command: `/app/{PS-ID} livez \|\| exit 1` | Built-in CLI, not wget |
| DF-17 | `ENTRYPOINT ["/sbin/tini", "--", "/app/{PS-ID}"]` | Signal handling |
| DF-18 | `USER ${CONTAINER_UID}:${CONTAINER_GID}` NOT commented out | Security |
| DF-19 | MUST NOT set `GOMODCACHE` or `GOCACHE` env vars | Unnecessary in runtime |
| DF-20 | MUST NOT have `CMD` instruction | Compose controls the command |
| DF-21 | Header comment references `{SUITE}-{PS-ID}:{IMAGE_TAG}` and display names | No copy-paste errors |
| DF-22 | `LABEL org.opencontainers.image.source` = `{GITHUB_REPOSITORY_URL}` | Repository link |
| DF-23 | `LABEL org.opencontainers.image.authors` = `{AUTHORS}` | Author attribution |
| DF-24 | `LABEL org.opencontainers.image.description` = `{PRODUCT_DISPLAY_NAME} {SERVICE_DISPLAY_NAME}` | Human-readable description |

---

### 4. Compose Enforceable Rules (CO-01–CO-22)

**Current state in ENG-HANDBOOK.md**: §12.3.1 covers Docker Compose conventions and secrets. §12.3.5 covers recursive includes. The named-volume mandate (CO-21/CO-22) and the full rule set are absent.

**Suggested addition to §12.3.1**:

| Rule | Requirement | Rationale |
|------|-------------|-----------|
| CO-01 | Header includes `$schema` reference and PS-ID description | Documentation |
| CO-02 | Includes `../shared-telemetry/compose.yml` and `../shared-postgres/compose.yml` | Required infrastructure |
| CO-03 | `healthcheck-secrets` service validates all 14 secrets (`test -s` per file) | Secret validation before start |
| CO-04 | `builder-{PS-ID}` service with `image: {SUITE}-{PS-ID}:{IMAGE_TAG}` | Image building |
| CO-05 | `pki-init` service command `["{PS-ID}", "/certs"]` (positional: tier-id, target-dir) | TLS bootstrap |
| CO-06 | Exactly 4 app instances: sqlite-1, sqlite-2, postgresql-1, postgresql-2 | Cross-DB testing |
| CO-07 | App service names follow `{PS-ID}-app-{variant}` pattern | Naming convention |
| CO-08 | Container port always `8080` | Standardized internal port |
| CO-09 | Host ports follow formula: sqlite-1=+0, sqlite-2=+1, pg-1=+2, pg-2=+3 | Port consistency |
| CO-10 | Command includes `server`, `--bind-public-port=8080`, `--config=...` args | Startup parameters |
| CO-11 | Config volume mount order: instance-specific, common, otel | Priority ordering |
| CO-12 | Healthcheck: `["CMD", "/app/{PS-ID}", "livez"]` (NOT wget) | Built-in healthcheck |
| CO-13 | Resource limits: 256M limit, 128M reservation | Resource control |
| CO-14 | Networks: `{PS-ID}-network` + `telemetry-network`; PostgreSQL adds `postgres-network` | Network isolation |
| CO-15 | `working_dir: /tmp` on all app services | Writable temp dir |
| CO-16 | All 14 secrets declared in `secrets:` with `file: ./secrets/` relative paths | Docker secrets |
| CO-17 | SQLite instances mount 10 secrets; PostgreSQL instances mount 14 (adds 4 postgres secrets) | Minimal secrets per variant |
| CO-18 | PostgreSQL instance 2 depends on instance 1 `service_healthy` | Schema init ordering |
| CO-19 | Healthcheck timing uses parameterized values (start-period, interval, timeout, retries) | Configurable health probes |
| CO-20 | All `image:` references use `{SUITE}-{PS-ID}:{IMAGE_TAG}` (NOT hardcoded) | Parameterized naming |
| CO-21 | All services use named volume `{PS-ID}-certs:/certs` (NEVER bind mount `./certs/:/certs`) | Portability — named volumes work in Docker Desktop, Swarm, and Kubernetes without host-path dependencies |
| CO-22 | Top-level `volumes:` declares `{PS-ID}-certs:` (no `driver:` override) | Named volume declaration required |

---

### 5. Deployment Config File Rules (CF-01–CF-17)

**Current state in ENG-HANDBOOK.md**: §13.2 describes kebab-case keys and the flat directory pattern. The PostgreSQL mTLS cert path rules are in §6.11.4. No unified config rule catalog exists.

**Suggested addition to §13.2**:

| Rule | Requirement | Rationale |
|------|-------------|-----------|
| CF-01 | ALL keys MUST be kebab-case (NEVER snake_case) | §13.2 consistency |
| CF-02 | Common file sets `bind-public-address: "0.0.0.0"` | Container networking |
| CF-03 | Common file references TLS cert paths from pki-init | TLS bootstrap |
| CF-04 | Common file references all 5 unseal secret paths | Barrier service |
| CF-05 | Common file references browser/service credential secret paths | Authentication |
| CF-06 | Instance files set `cors-origins` with correct port | CORS correctness |
| CF-07 | Instance files set `otlp-service` = `{PS-ID}-{variant}` | Telemetry identity |
| CF-08 | Instance files set `otlp-hostname` = `{PS-ID}-{variant}` | Telemetry identity |
| CF-09 | SQLite instance files set `database-url: "sqlite://file::memory:?cache=shared"` | In-memory SQLite |
| CF-10 | PostgreSQL instance files MUST NOT set `database-url` (passed via compose command) | DSN from secret |
| CF-11 | Instance files MUST NOT duplicate keys from common file | DRY principle |
| CF-12 | Header comment references correct `{PS-ID}` (not another service) | No copy-paste errors |
| CF-13 | PostgreSQL instance files set `database-sslmode: verify-full` | PostgreSQL mTLS |
| CF-14 | PostgreSQL instance files set `database-sslcert` (Cat 14 per-instance cert path) | PostgreSQL mTLS |
| CF-15 | PostgreSQL instance files set `database-sslkey` (Cat 14 per-instance key path) | PostgreSQL mTLS |
| CF-16 | PostgreSQL instance files set `database-sslrootcert` (Cat 10 truststore path) | PostgreSQL mTLS |
| CF-17 | SQLite instance files MUST NOT set any `database-ssl*` fields | No PG certs for SQLite |

---

### 6. Standalone Config Rules (SC-01–SC-06)

**Current state in ENG-HANDBOOK.md**: §13.2 mentions standalone configs at `configs/{PS-ID}/{PS-ID}.yml`. No rule catalog exists.

**Suggested addition to §13.2**:

| Rule | Requirement | Rationale |
|------|-------------|-----------|
| SC-01 | ALL keys MUST be kebab-case | §13.2 consistency |
| SC-02 | `bind-public-address` MUST be `127.0.0.1` (NOT `0.0.0.0`) | Windows firewall prevention |
| SC-03 | `bind-public-port` MUST equal `{SERVICE_APP_PORT_BASE}` from registry | Port consistency |
| SC-04 | `bind-admin-port` MUST be `9090` | Admin port standardization |
| SC-05 | `otlp-service` MUST equal `{PS-ID}` | Telemetry naming |
| SC-06 | Header comment references correct `{PRODUCT_DISPLAY_NAME} {SERVICE_DISPLAY_NAME}` and `{PS-ID}` | No copy-paste errors |

---

### 7. Product and Suite Compose Rules

**Current state in ENG-HANDBOOK.md**: §12.3.5 covers recursive include semantics and port override formulas. The enforceable rule sets for product and suite compose files are absent.

**Suggested addition to §12.3.5**:

**Product compose rules (PC-01–PC-06)**:

| Rule | Requirement | Rationale |
|------|-------------|-----------|
| PC-01 | MUST include all child PS-ID compose files | Recursive architecture |
| PC-02 | MUST override `pki-init` command to `["{PRODUCT}", "/certs"]` | Product-scoped certs |
| PC-03 | Port overrides MUST use `!override` tag | Complete replacement |
| PC-04 | Port formula: SERVICE + 10000 | Product tier offset |
| PC-05 | Unseal secrets use `{PRODUCT}-unseal-key-N-of-5-...` values | Product-scoped encryption |
| PC-06 | MUST include `browser-*.secret.never` and `service-*.secret.never` marker files | Credential scope enforcement |

**Suite compose rules (SU-01–SU-04)**:

| Rule | Requirement | Rationale |
|------|-------------|-----------|
| SU-01 | MUST include all 5 product compose files | Complete suite |
| SU-02 | Port formula: SERVICE + 20000 | Suite tier offset |
| SU-03 | Compact inline port override syntax for 40 services | Readability |
| SU-04 | Unseal secrets use `cryptoutil-unseal-key-N-of-5-...` values | Suite-scoped encryption |

---

### 8. Secret File Value Patterns

**Current state in ENG-HANDBOOK.md**: §12.3.3 covers the secret naming convention and tier-specific value prefixes. The full table of all 14 secret files with exact value format patterns is not present.

**Suggested addition to §12.3.3**:

| Filename | Value Pattern | Notes |
|----------|---------------|-------|
| `unseal-1of5.secret` | `{PS-ID}-unseal-key-1-of-5-{base64-random-32-bytes}` | Unique per shard |
| `unseal-2of5.secret` | `{PS-ID}-unseal-key-2-of-5-{base64-random-32-bytes}` | Unique per shard |
| `unseal-3of5.secret` | `{PS-ID}-unseal-key-3-of-5-{base64-random-32-bytes}` | Unique per shard |
| `unseal-4of5.secret` | `{PS-ID}-unseal-key-4-of-5-{base64-random-32-bytes}` | Unique per shard |
| `unseal-5of5.secret` | `{PS-ID}-unseal-key-5-of-5-{base64-random-32-bytes}` | Unique per shard |
| `hash-pepper-v3.secret` | `{PS-ID}-hash-pepper-v3-{base64-random-32-bytes}` | Hash pepper |
| `postgres-url.secret` | `postgres://{PS_ID}_database_user:{PS_ID}_database_pass@shared-postgres-leader:5432/{PS_ID}_database` | Base DSN ONLY — no `sslmode=` param; SSL configured via YAML `database-ssl*` fields |
| `postgres-username.secret` | `{PS_ID}_database_user` | PostgreSQL username |
| `postgres-password.secret` | `{PS_ID}_database_pass-{base64-random-32-bytes}` | PostgreSQL password |
| `postgres-database.secret` | `{PS_ID}_database` | PostgreSQL database name |
| `browser-username.secret` | `{PS-ID}-browser-user` | Browser authn username |
| `browser-password.secret` | `{PS-ID}-browser-pass-{base64-random-32-bytes}` | Browser authn password |
| `service-username.secret` | `{PS-ID}-service-user` | Service authn username |
| `service-password.secret` | `{PS-ID}-service-pass-{base64-random-32-bytes}` | Service authn password |

---

### 9. Shared PostgreSQL mTLS Cert Reference Table

**Current state in ENG-HANDBOOK.md**: §6.11.4 covers the staged wiring sequence and D2/D4/D5 decisions. The per-node cert ownership table and PKI category reference for each shared-postgres node are absent.

**Suggested addition to §6.11.4 or §12.3.1**:

**PKI category reference for PostgreSQL nodes**:

| PKI Cat | Name | Role |
|---------|------|------|
| 10 | `postgres-tls-server-issuing-ca` | Server CA truststore — app clients verify leader cert |
| 11 | `postgres-tls-server-entity-leader` | Leader server cert+key |
| 11 | `postgres-tls-server-entity-follower` | Follower server cert+key |
| 12 | `postgres-tls-client-issuing-ca` | Client CA truststore — leader verifies app + replication client certs |
| 13 | `postgres-tls-client-entity-follower-replication` | Follower replication client cert+key |
| 14 | `postgres-tls-client-entity-leader-{PS-ID}-postgres-1` | App postgres-1 client cert+key |
| 14 | `postgres-tls-client-entity-leader-{PS-ID}-postgres-2` | App postgres-2 client cert+key |

**Logical cert ownership per node**:

| Node | Certs Used | Why |
|------|-----------|-----|
| postgres-leader | Cat 11 leader (server) + Cat 12 (verify clients) | Serves TLS; requires client certs |
| postgres-follower | Cat 11 follower (server) + Cat 12 (verify clients) + Cat 10 (verify leader) + Cat 13 (replication client) | Serves TLS; mTLS replication to leader |
| `{PS-ID}-postgres-1` | Cat 14 postgres-1 client + Cat 10 (verify leader) | mTLS client to leader |
| `{PS-ID}-postgres-2` | Cat 14 postgres-2 client + Cat 10 | mTLS client to leader |
| `{PS-ID}-sqlite-1/2` | **None** | SQLite instances do NOT connect to PostgreSQL |

---

### 10. Current Inconsistency Inventory

**Current state in ENG-HANDBOOK.md**: §13.6 describes what the template-compliance linter enforces but does not document the known pre-existing deviations.

**Suggested addition to §13.6 or §M (new appendix)**:

**Three divergent Dockerfile patterns** where one canonical pattern must exist:

| Pattern | Affected PS-IDs | Key Deviations |
|---------|----------------|----------------|
| Pattern A (sm-kms style) | sm-kms, identity-authz, identity-idp, identity-rp, identity-rs | `WORKDIR /app/run`, `GOMODCACHE`/`GOCACHE` env vars, `curl` in final, `USER` commented out, individual LABEL lines |
| Pattern B (jose-ja style) | jose-ja, pki-ca, skeleton-template | 3-stage (no `runtime-deps`), `adduser`-based user creation, compact LABEL, `CMD` instruction present |
| Pattern C (sm-im) | sm-im | 2-stage (no `validation`), UID `1000:1000` (wrong), no BuildKit caches, no static link check |

**Specific bugs requiring P0/P1 fixes**:

| PS-ID | Bug | Impact |
|-------|-----|--------|
| `identity-spa` | Builder builds `/app/identity-spa` but `COPY` copies `/app/cryptoutil` | Runtime failure — wrong binary |
| `skeleton-template` | Header, username, dirs, and CMD all reference `jose` | Documentation + functional error |
| `sm-im` | UID:GID is `1000:1000` (should be `65532:65532`) | Security deviation |
| All Pattern A | `USER 65532:65532` commented out | Containers running as root |
| `cryptoutil` (suite) | `tini` not installed/copied | Missing signal handling |

**Config key naming inconsistencies**:

| Convention | PS-IDs |
|------------|--------|
| kebab-case (CORRECT) | jose-ja, pki-ca, skeleton-template |
| snake_case (WRONG — needs migration) | sm-kms, sm-im, identity-authz, identity-idp, identity-rp, identity-rs, identity-spa |

---

### 11. Template Syntax Specification

**Current state in ENG-HANDBOOK.md**: §13.6 mentions `__KEY__` format but does not document the full syntax specification or explain the delimiter choice.

**Suggested addition to §13.6**:

> **Placeholder format**: `__PARAMETER__` (double underscore prefix and suffix, ALL_CAPS with underscores). Double underscores avoid conflicts with Dockerfile `${VAR}` shell syntax and YAML `${}` variable references.
>
> **Path-level vs content-level placeholders**:
> - `__PS_ID__` in a file path triggers per-PS-ID expansion (×10 instances)
> - `__PRODUCT__` in a file path triggers per-product expansion (×5 instances)
> - `__SUITE__` in a file path triggers suite-level expansion (×1 instance)
> - Content-only placeholders like `__PS_ID_UPPER__` and `__SUITE__` are substituted within file content without triggering path expansion.
>
> **Template file catalog** (relative to `api/cryptosuite-registry/templates/`):
>
> | Template | Instantiation Count |
> |----------|---------------------|
> | `deployments/__PS_ID__/Dockerfile` | ×10 |
> | `deployments/__PS_ID__/compose.yml` | ×10 |
> | `deployments/__PS_ID__/config/__PS_ID__-app-framework-common.yml` | ×10 |
> | `deployments/__PS_ID__/config/__PS_ID__-app-framework-sqlite-1.yml` | ×10 |
> | `deployments/__PS_ID__/config/__PS_ID__-app-framework-sqlite-2.yml` | ×10 |
> | `deployments/__PS_ID__/config/__PS_ID__-app-framework-postgresql-1.yml` | ×10 |
> | `deployments/__PS_ID__/config/__PS_ID__-app-framework-postgresql-2.yml` | ×10 |
> | `configs/__PS_ID__/__PS_ID__-framework.yml` | ×10 |
> | `deployments/__PRODUCT__/compose.yml` | ×5 |
> | `deployments/__SUITE__/Dockerfile` | ×1 |
> | `deployments/__SUITE__/compose.yml` | ×1 |
