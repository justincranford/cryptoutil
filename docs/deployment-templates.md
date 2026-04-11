# Deployment Templates — Canonical File Content

**Companion to**: [target-structure.md](target-structure.md) (directory layout), [tls-structure.md](tls-structure.md) (certificate layout)

**Purpose**: Defines the exact canonical content for every file inside `deployments/` and `configs/`.
While `target-structure.md` specifies **what files exist**, this document specifies **what goes inside each file**.
While `tls-structure.md` specifies **what certificates exist**, this document specifies **how services reference them**.

**Enforcement**: `cicd-lint lint-fitness` and `cicd-lint lint-deployments` validators MUST enforce these templates.
Any deviation from these templates is a blocking error.

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
|-----------|-------------|-------------------|--------------------|
| `{PS-ID}` | Product-Service identifier (kebab-case) | `sm-kms` | `jose-ja` |
| `{PS_ID}` | Underscore variant (PostgreSQL naming) | `sm_kms` | `jose_ja` |
| `{PRODUCT}` | Product name (kebab-case) | `sm` | `jose` |
| `{SERVICE}` | Service name within product | `kms` | `ja` |
| `{DISPLAY_NAME}` | Human-readable name | `Secrets Manager Key Management Service` | `JOSE JWK Authority` |
| `{SUITE}` | Suite name (always `cryptoutil`) | `cryptoutil` | `cryptoutil` |

### A.2 Port Parameters

| Parameter | Description | Formula | Example (sm-kms) | Example (jose-ja) |
|-----------|-------------|---------|-------------------|--------------------|
| `{PORT_BASE}` | SERVICE-tier port block start | Per registry | `8000` | `8200` |
| `{PORT_SQLITE_1}` | SQLite instance 1 host port | `{PORT_BASE} + 0` | `8000` | `8200` |
| `{PORT_SQLITE_2}` | SQLite instance 2 host port | `{PORT_BASE} + 1` | `8001` | `8201` |
| `{PORT_PG_1}` | PostgreSQL instance 1 host port | `{PORT_BASE} + 2` | `8002` | `8202` |
| `{PORT_PG_2}` | PostgreSQL instance 2 host port | `{PORT_BASE} + 3` | `8003` | `8203` |
| `{PG_HOST_PORT}` | PostgreSQL container host port | Per registry | `54320` | `54322` |
| `{PRODUCT_OFFSET}` | PRODUCT tier formula | `{PORT_BASE} + 10000` | `18000` | `18200` |
| `{SUITE_OFFSET}` | SUITE tier formula | `{PORT_BASE} + 20000` | `28000` | `28200` |

### A.3 Complete PS-ID Parameter Matrix

| PS-ID | Product | PORT_BASE | PG_HOST_PORT | DISPLAY_NAME |
|-------|---------|-----------|--------------|--------------|
| `sm-kms` | `sm` | `8000` | `54320` | Secrets Manager Key Management Service |
| `sm-im` | `sm` | `8100` | `54321` | Secrets Manager Instant Messenger |
| `jose-ja` | `jose` | `8200` | `54322` | JOSE JWK Authority |
| `pki-ca` | `pki` | `8300` | `54323` | PKI Certificate Authority |
| `identity-authz` | `identity` | `8400` | `54324` | Identity Authorization Server |
| `identity-idp` | `identity` | `8500` | `54325` | Identity Provider |
| `identity-rs` | `identity` | `8600` | `54326` | Identity Resource Server |
| `identity-rp` | `identity` | `8700` | `54327` | Identity Relying Party |
| `identity-spa` | `identity` | `8800` | `54328` | Identity Single Page App |
| `skeleton-template` | `skeleton` | `8900` | `54329` | Skeleton Template |

---

## B. PS-ID Dockerfile Template

**ONE canonical pattern for all 10 PS-IDs.** Four stages: `validation` → `builder` → `runtime-deps` → `final`.

### B.1 Canonical Template

```dockerfile
#############################################################################################
# {PS-ID} Dockerfile — {DISPLAY_NAME}
#
# Build: DOCKER_BUILDKIT=1 docker build \
#          --build-arg APP_VERSION=<ver> \
#          --build-arg VCS_REF=$(git rev-parse HEAD) \
#          --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
#          -t cryptoutil-{PS-ID}:dev \
#          -f deployments/{PS-ID}/Dockerfile .
#############################################################################################
ARG APP_VERSION=UNSET
ARG VCS_REF=UNSET
ARG BUILD_DATE=UNSET
#############################################################################################
ARG GO_VERSION=1.26.1
# hadolint ignore=DL3007
ARG ALPINE_VERSION=latest
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
ARG LDFLAGS="-s -extldflags '-static'"
#############################################################################################

#############################################################################################
# Stage 1: Validate build arguments
#############################################################################################
FROM alpine:${ALPINE_VERSION} AS validation
ARG APP_VERSION=UNSET
ARG VCS_REF=UNSET
ARG BUILD_DATE=UNSET

RUN set -e; \
    errors=""; \
    if [ "$APP_VERSION" = "UNSET" ]; then \
        errors="${errors}ERROR: APP_VERSION build argument is required\n"; \
    fi; \
    if [ "$VCS_REF" = "UNSET" ]; then \
        errors="${errors}ERROR: VCS_REF build argument is required\n"; \
    fi; \
    if [ "$BUILD_DATE" = "UNSET" ]; then \
        errors="${errors}ERROR: BUILD_DATE build argument is required\n"; \
    fi; \
    if [ -n "$errors" ]; then \
        printf "%b" "$errors" >&2; \
        exit 1; \
    fi

RUN mkdir -p /app && \
    echo "APP_VERSION=${APP_VERSION}" > /app/.build-params && \
    echo "VCS_REF=${VCS_REF}" >> /app/.build-params && \
    echo "BUILD_DATE=${BUILD_DATE}" >> /app/.build-params

#############################################################################################
# Stage 2: Build Go binary
#############################################################################################
FROM golang:${GO_VERSION} AS builder
WORKDIR /src

ARG APP_VERSION
ARG VCS_REF
ARG BUILD_DATE
ARG GO_VERSION
ARG CGO_ENABLED
ARG GOOS
ARG GOARCH
ARG LDFLAGS

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -a -tags netgo -trimpath -ldflags="${LDFLAGS}" \
    -o /app/{PS-ID} ./cmd/{PS-ID}

# Validate static linking
SHELL ["/bin/bash", "-c"]
RUN if ldd /app/{PS-ID} 2>/dev/null; then \
        echo "Binary is dynamically linked - failing build"; \
        exit 1; \
    fi

RUN mkdir -p /app && chmod 555 /app && chown -R 65532:65532 /app

#############################################################################################
# Stage 3: Runtime dependencies
#############################################################################################
# hadolint ignore=DL3006
FROM alpine:${ALPINE_VERSION} AS runtime-deps
# hadolint ignore=DL3018
RUN apk --no-cache add ca-certificates tzdata tini && \
    update-ca-certificates

#############################################################################################
# Stage 4: Final image
#############################################################################################
# hadolint ignore=DL3006
FROM alpine:${ALPINE_VERSION} AS final

# Copy runtime dependencies
COPY --from=runtime-deps /sbin/tini /sbin/tini
COPY --from=runtime-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy application binary and build metadata
COPY --from=builder /app /app
COPY --from=validation /app/.build-params /app/.build-params

# Create non-root user
RUN addgroup -g 65532 -S appgroup && \
    adduser -u 65532 -S appuser -G appgroup -h /app -s /sbin/nologin

WORKDIR /app

# OCI image labels
ARG APP_VERSION
ARG VCS_REF
ARG BUILD_DATE
LABEL org.opencontainers.image.title="cryptoutil-{PS-ID}" \
      org.opencontainers.image.description="{DISPLAY_NAME}" \
      org.opencontainers.image.source="https://github.com/justincranford/cryptoutil" \
      org.opencontainers.image.authors="Justin Cranford" \
      org.opencontainers.image.version="${APP_VERSION}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.created="${BUILD_DATE}"

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD /app/{PS-ID} livez || exit 1

ENTRYPOINT ["/sbin/tini", "--", "/app/{PS-ID}"]

USER 65532:65532
```

### B.2 Template Rules (Enforceable)

| Rule ID | Rule | Rationale |
|---------|------|-----------|
| DF-01 | MUST have exactly 4 named stages: `validation`, `builder`, `runtime-deps`, `final` | Structural consistency |
| DF-02 | MUST use `ARG GO_VERSION=1.26.1` | Version consistency |
| DF-03 | MUST use `ARG ALPINE_VERSION=latest` with `# hadolint ignore=DL3007` | Security patch policy |
| DF-04 | MUST use `ARG CGO_ENABLED=0` | CGO ban |
| DF-05 | Builder MUST use BuildKit cache mounts for `/go/pkg/mod` and `/root/.cache/go-build` | Build performance |
| DF-06 | MUST build `./cmd/{PS-ID}` and output to `/app/{PS-ID}` | Binary naming |
| DF-07 | MUST validate static linking with `ldd` check | Portability |
| DF-08 | Runtime-deps MUST install `ca-certificates`, `tzdata`, `tini` only (NO curl, NO wget) | Minimal attack surface |
| DF-09 | Final stage MUST NOT install any packages via `apk` | Minimal attack surface |
| DF-10 | MUST create user with UID:GID `65532:65532` | Security (nonroot) |
| DF-11 | MUST use `WORKDIR /app` (NOT `/app/run`) | Uniformity |
| DF-12 | MUST use compact multi-line `LABEL` block (NOT individual LABEL lines) | Style consistency |
| DF-13 | `LABEL org.opencontainers.image.title` MUST equal `cryptoutil-{PS-ID}` | Naming convention |
| DF-14 | MUST have `EXPOSE 8080` only (NO 9090) | Admin is 127.0.0.1-only |
| DF-15 | HEALTHCHECK MUST use `/app/{PS-ID} livez \|\| exit 1` | Built-in CLI |
| DF-16 | ENTRYPOINT MUST be `["/sbin/tini", "--", "/app/{PS-ID}"]` | Signal handling |
| DF-17 | MUST end with `USER 65532:65532` (NOT commented out) | Security |
| DF-18 | MUST NOT set `GOMODCACHE` or `GOCACHE` env vars | Unnecessary in runtime |
| DF-19 | MUST NOT have `CMD` instruction (compose overrides command) | Compose controls |
| DF-20 | Header comment MUST reference `{PS-ID}` and `{DISPLAY_NAME}` correctly | No copy-paste errors |

### B.3 Template Variables Per PS-ID

| PS-ID | Binary Path | HEALTHCHECK Path | ENTRYPOINT | LABEL title |
|-------|-------------|------------------|------------|-------------|
| `sm-kms` | `/app/sm-kms` | `/app/sm-kms livez` | `["/sbin/tini", "--", "/app/sm-kms"]` | `cryptoutil-sm-kms` |
| `sm-im` | `/app/sm-im` | `/app/sm-im livez` | `["/sbin/tini", "--", "/app/sm-im"]` | `cryptoutil-sm-im` |
| `jose-ja` | `/app/jose-ja` | `/app/jose-ja livez` | `["/sbin/tini", "--", "/app/jose-ja"]` | `cryptoutil-jose-ja` |
| `pki-ca` | `/app/pki-ca` | `/app/pki-ca livez` | `["/sbin/tini", "--", "/app/pki-ca"]` | `cryptoutil-pki-ca` |
| `identity-authz` | `/app/identity-authz` | `/app/identity-authz livez` | `["/sbin/tini", "--", "/app/identity-authz"]` | `cryptoutil-identity-authz` |
| `identity-idp` | `/app/identity-idp` | `/app/identity-idp livez` | `["/sbin/tini", "--", "/app/identity-idp"]` | `cryptoutil-identity-idp` |
| `identity-rs` | `/app/identity-rs` | `/app/identity-rs livez` | `["/sbin/tini", "--", "/app/identity-rs"]` | `cryptoutil-identity-rs` |
| `identity-rp` | `/app/identity-rp` | `/app/identity-rp livez` | `["/sbin/tini", "--", "/app/identity-rp"]` | `cryptoutil-identity-rp` |
| `identity-spa` | `/app/identity-spa` | `/app/identity-spa livez` | `["/sbin/tini", "--", "/app/identity-spa"]` | `cryptoutil-identity-spa` |
| `skeleton-template` | `/app/skeleton-template` | `/app/skeleton-template livez` | `["/sbin/tini", "--", "/app/skeleton-template"]` | `cryptoutil-skeleton-template` |

---

## C. PS-ID compose.yml Template

**ONE canonical pattern for all 10 PS-IDs.** Four app instances + support services.

### C.1 Canonical Template

```yaml
# $schema: https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json
#
# {PS-ID} Docker Compose Configuration
#
# SERVICE-level deployment for {DISPLAY_NAME}.
# Supports 4 instances: 2 SQLite + 2 PostgreSQL.
#
# Port allocation (SERVICE level: {PORT_BASE}-{PORT_BASE+3}):
#   - {PS-ID}-app-sqlite-1:      {PORT_SQLITE_1}
#   - {PS-ID}-app-sqlite-2:      {PORT_SQLITE_2}
#   - {PS-ID}-app-postgresql-1:  {PORT_PG_1}
#   - {PS-ID}-app-postgresql-2:  {PORT_PG_2}
#
# Dual Role: Standalone deployable AND include target for PRODUCT/SUITE compose files.
# Usage:
#   docker compose -f compose.yml up -d
#   docker compose -f compose.yml up {PS-ID}-app-sqlite-1 -d
#   docker compose -f compose.yml logs -f
#   docker compose -f compose.yml down -v
#
# Health Checks:
#   - Public: https://localhost:{PORT_SQLITE_1}/browser/api/v1/health
#   - Admin:  https://127.0.0.1:9090/admin/api/v1/livez (container-internal only)
#
include:
  - path: ../shared-telemetry/compose.yml
  - path: ../shared-postgres/compose.yml

services:
  # Docker secrets availability check.
  healthcheck-secrets:
    image: alpine:latest
    secrets:
      - unseal-1of5.secret
      - unseal-2of5.secret
      - unseal-3of5.secret
      - unseal-4of5.secret
      - unseal-5of5.secret
      - hash-pepper-v3.secret
      - postgres-url.secret
      - postgres-username.secret
      - postgres-password.secret
      - postgres-database.secret
      - browser-username.secret
      - browser-password.secret
      - service-username.secret
      - service-password.secret
    command: ["sh", "-c", "ls -l /run/secrets/*.secret"]
    networks:
      - {PS-ID}-network

  # Build image from source.
  builder-{PS-ID}:
    image: cryptoutil-{PS-ID}:dev
    build:
      context: ../..
      dockerfile: deployments/{PS-ID}/Dockerfile
      args:
        APP_VERSION: "dev"
        VCS_REF: "local"
        BUILD_DATE: "2026-01-01T00:00:00Z"
    entrypoint: ["sh", "-c"]
    command: ["echo 'Build completed successfully'"]
    depends_on:
      healthcheck-secrets:
        condition: service_completed_successfully

  # PKI init: bootstrap TLS certificates.
  pki-init:
    image: cryptoutil-{PS-ID}:dev
    command: ["init", "--output-dir=/certs"]
    volumes:
      - ./certs/:/certs/:rw
    depends_on:
      builder-{PS-ID}:
        condition: service_completed_successfully
    networks:
      - {PS-ID}-network

  # App: SQLite instance 1.
  {PS-ID}-app-sqlite-1:
    image: cryptoutil-{PS-ID}:dev
    command: >-
      server
      --bind-public-port=8080
      --config=/certs/tls-config.yml
      --config=/app/config/{PS-ID}-app-sqlite-1.yml
      --config=/app/config/{PS-ID}-app-common.yml
      --config=/app/otel/otel.yml
      -u sqlite://file::memory:?cache=shared
    working_dir: /tmp
    ports:
      - "{PORT_SQLITE_1}:8080"
    volumes:
      - ./config/{PS-ID}-app-sqlite-1.yml:/app/config/{PS-ID}-app-sqlite-1.yml:ro
      - ./config/{PS-ID}-app-common.yml:/app/config/{PS-ID}-app-common.yml:ro
      - ../shared-telemetry/otel/cryptoutil-otel.yml:/app/otel/otel.yml:ro
      - ./certs/:/certs/:ro
    secrets:
      - unseal-1of5.secret
      - unseal-2of5.secret
      - unseal-3of5.secret
      - unseal-4of5.secret
      - unseal-5of5.secret
    depends_on:
      healthcheck-secrets:
        condition: service_completed_successfully
      builder-{PS-ID}:
        condition: service_completed_successfully
      pki-init:
        condition: service_completed_successfully
      opentelemetry-collector-contrib:
        condition: service_started
      healthcheck-opentelemetry-collector-contrib:
        condition: service_completed_successfully
    healthcheck:
      test: ["CMD", "wget", "--ca-certificate=/certs/root-ca.pem", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/api/v1/livez"]
      start_period: 60s
      interval: 5s
      timeout: 5s
      retries: 10
    networks:
      - {PS-ID}-network
      - telemetry-network
    deploy:
      resources:
        limits:
          memory: 256M
        reservations:
          memory: 128M

  # App: SQLite instance 2.
  {PS-ID}-app-sqlite-2:
    # ... identical to sqlite-1 except:
    #   port: {PORT_SQLITE_2}:8080
    #   config: {PS-ID}-app-sqlite-2.yml

  # App: PostgreSQL instance 1.
  {PS-ID}-app-postgresql-1:
    # ... identical to sqlite-1 except:
    #   port: {PORT_PG_1}:8080
    #   config: {PS-ID}-app-postgresql-1.yml
    #   command: -u file:///run/secrets/postgres-url.secret (instead of sqlite URL)
    #   additional secret: postgres-url.secret
    #   additional depends_on: postgres-leader (condition: service_healthy)
    #   additional network: postgres-network

  # App: PostgreSQL instance 2.
  {PS-ID}-app-postgresql-2:
    # ... identical to postgresql-1 except:
    #   port: {PORT_PG_2}:8080
    #   config: {PS-ID}-app-postgresql-2.yml
    #   additional depends_on: {PS-ID}-app-postgresql-1 (condition: service_healthy)

networks:
  {PS-ID}-network:
    driver: bridge

secrets:
  unseal-1of5.secret:
    file: ./secrets/unseal-1of5.secret # pragma: allowlist secret
  unseal-2of5.secret:
    file: ./secrets/unseal-2of5.secret
  unseal-3of5.secret:
    file: ./secrets/unseal-3of5.secret
  unseal-4of5.secret:
    file: ./secrets/unseal-4of5.secret
  unseal-5of5.secret:
    file: ./secrets/unseal-5of5.secret
  hash-pepper-v3.secret:
    file: ./secrets/hash-pepper-v3.secret
  postgres-url.secret:
    file: ./secrets/postgres-url.secret
  postgres-username.secret:
    file: ./secrets/postgres-username.secret
  postgres-password.secret:
    file: ./secrets/postgres-password.secret
  postgres-database.secret:
    file: ./secrets/postgres-database.secret
  browser-username.secret:
    file: ./secrets/browser-username.secret
  browser-password.secret:
    file: ./secrets/browser-password.secret
  service-username.secret:
    file: ./secrets/service-username.secret
  service-password.secret:
    file: ./secrets/service-password.secret
```

### C.2 Compose Rules (Enforceable)

| Rule ID | Rule | Rationale |
|---------|------|-----------|
| CO-01 | Header MUST include `$schema` reference and PS-ID description | Documentation |
| CO-02 | MUST include `../shared-telemetry/compose.yml` and `../shared-postgres/compose.yml` | Required infrastructure |
| CO-03 | MUST have `healthcheck-secrets` service listing all 14 secrets | Secret validation |
| CO-04 | MUST have `builder-{PS-ID}` service with build context `../..` | Image building |
| CO-05 | MUST have `pki-init` service with `["init", "--output-dir=/certs"]` | TLS bootstrap |
| CO-06 | MUST have exactly 4 app instances: sqlite-1, sqlite-2, postgresql-1, postgresql-2 | Cross-DB testing |
| CO-07 | App service names MUST follow `{PS-ID}-app-{variant}` pattern | Naming convention |
| CO-08 | Container port MUST always be `8080` | Standardized internal port |
| CO-09 | Host ports MUST follow port formula: sqlite-1=+0, sqlite-2=+1, pg-1=+2, pg-2=+3 | Port consistency |
| CO-10 | Command MUST include: `server`, `--bind-public-port=8080`, `--config=...` args | Startup parameters |
| CO-11 | Config volume mount order: instance-specific, common, otel | Priority ordering |
| CO-12 | Healthcheck MUST use wget with `--ca-certificate=/certs/root-ca.pem` to admin livez | TLS-aware healthcheck |
| CO-13 | Resource limits: 256M limit, 128M reservation | Resource control |
| CO-14 | Networks: `{PS-ID}-network` + `telemetry-network`; PostgreSQL instances add `postgres-network` | Network isolation |
| CO-15 | `working_dir: /tmp` on all app services | Writable temp dir |
| CO-16 | All 14 secrets MUST be declared in `secrets:` section with `file: ./secrets/` relative paths | Docker secrets |
| CO-17 | SQLite instances do NOT mount `postgres-url.secret` | Minimal secrets |
| CO-18 | PostgreSQL instance 2 MUST depend on instance 1 `service_healthy` | Schema init ordering |

---

## D. PS-ID Deployment Config Templates

Five deployment config files per PS-ID, ALL using **kebab-case** YAML keys.

### D.1 `{PS-ID}-app-common.yml` (Shared Settings)

```yaml
# {PS-ID} Common Configuration
# Shared settings across all {PS-ID} deployment instances.

# Binding address — 0.0.0.0 allows connections from Docker network.
bind-public-address: "0.0.0.0"

# TLS certificate (generated by pki-init).
tls-cert-file: /app/tls_public_server_certificate_0.pem
tls-key-file: /app/tls_public_server_private_key.pem

# Unseal configuration (3-of-5 Shamir shards).
unseal-mode: "3-of-5"
unseal-files:
  - /run/secrets/unseal-1of5.secret
  - /run/secrets/unseal-2of5.secret
  - /run/secrets/unseal-3of5.secret
  - /run/secrets/unseal-4of5.secret
  - /run/secrets/unseal-5of5.secret

# Authentication credentials (Docker secrets).
browser-username-file: /run/secrets/browser-username.secret
browser-password-file: /run/secrets/browser-password.secret
service-username-file: /run/secrets/service-username.secret
service-password-file: /run/secrets/service-password.secret

# IP allowlist — allow all for development.
allowed-ips:
  - "127.0.0.1"
  - "::1"
  - "::ffff:127.0.0.1"
allowed-cidrs:
  - "0.0.0.0/0"
  - "::/0"

# CSRF — disabled for API testing.
csrf-token-single-use-token: false
```

### D.2 `{PS-ID}-app-sqlite-1.yml`

```yaml
# {PS-ID} SQLite Instance 1 Configuration
# Settings UNIQUE to the '{PS-ID}-app-sqlite-1' compose service.

cors-origins:
  - "https://localhost:{PORT_SQLITE_1}"
  - "https://127.0.0.1:{PORT_SQLITE_1}"
  - "https://[::1]:{PORT_SQLITE_1}"

otlp-service: {PS-ID}-sqlite-1
otlp-hostname: {PS-ID}-sqlite-1

database-url: "sqlite://file::memory:?cache=shared"
```

### D.3 `{PS-ID}-app-sqlite-2.yml`

```yaml
# {PS-ID} SQLite Instance 2 Configuration
# Settings UNIQUE to the '{PS-ID}-app-sqlite-2' compose service.

cors-origins:
  - "https://localhost:{PORT_SQLITE_2}"
  - "https://127.0.0.1:{PORT_SQLITE_2}"
  - "https://[::1]:{PORT_SQLITE_2}"

otlp-service: {PS-ID}-sqlite-2
otlp-hostname: {PS-ID}-sqlite-2

database-url: "sqlite://file::memory:?cache=shared"
```

### D.4 `{PS-ID}-app-postgresql-1.yml`

```yaml
# {PS-ID} PostgreSQL Instance 1 Configuration
# Settings UNIQUE to the '{PS-ID}-app-postgresql-1' compose service.

cors-origins:
  - "https://localhost:{PORT_PG_1}"
  - "https://127.0.0.1:{PORT_PG_1}"
  - "https://[::1]:{PORT_PG_1}"

otlp-service: {PS-ID}-postgresql-1
otlp-hostname: {PS-ID}-postgresql-1
```

### D.5 `{PS-ID}-app-postgresql-2.yml`

```yaml
# {PS-ID} PostgreSQL Instance 2 Configuration
# Settings UNIQUE to the '{PS-ID}-app-postgresql-2' compose service.

cors-origins:
  - "https://localhost:{PORT_PG_2}"
  - "https://127.0.0.1:{PORT_PG_2}"
  - "https://[::1]:{PORT_PG_2}"

otlp-service: {PS-ID}-postgresql-2
otlp-hostname: {PS-ID}-postgresql-2
```

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

```yaml
# {DISPLAY_NAME} Configuration
# Local development config for {PS-ID}.
# Override with deployment configs in deployments/{PS-ID}/config/

# Server binding (local development).
bind-public-address: "127.0.0.1"
bind-public-port: {PORT_BASE}
bind-admin-address: "127.0.0.1"
bind-admin-port: 9090

# TLS (auto-generated for local dev).
tls-enabled: true
tls-cert-file: ""
tls-key-file: ""

# CORS (local development ports).
cors-origins:
  - "https://localhost:{PORT_BASE}"
  - "https://127.0.0.1:{PORT_BASE}"

# Telemetry (disabled for local dev).
otlp-enabled: false
otlp-endpoint: ""
otlp-service: "{PS-ID}"
otlp-hostname: "localhost"

# Logging.
log-level: "INFO"
```

### E.1 Standalone Config Rules

| Rule ID | Rule | Rationale |
|---------|------|-----------|
| SC-01 | ALL keys MUST be kebab-case | ENG-HANDBOOK §13.2 |
| SC-02 | `bind-public-address` MUST be `127.0.0.1` (NOT `0.0.0.0`) | Windows firewall prevention |
| SC-03 | `bind-public-port` MUST equal `{PORT_BASE}` from registry | Port consistency |
| SC-04 | `bind-admin-port` MUST be `9090` | Admin port standardization |
| SC-05 | `otlp-service` MUST equal `{PS-ID}` | Telemetry naming |
| SC-06 | Header comment MUST reference `{PS-ID}` correctly | No copy-paste errors |

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
| `postgres-url.secret` | `postgres://{PS_ID}_database_user:{PS_ID}_database_pass@shared-postgres-leader:5432/{PS_ID}_database?sslmode=disable` | Full DSN |
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

```yaml
# $schema: https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json
#
# {PRODUCT} Product Docker Compose Configuration
#
# PRODUCT-level deployment for {PRODUCT} product ({N} services: {list}).
# Recursive includes: {PS-ID-1} and {PS-ID-2} (which include shared-postgres and shared-telemetry).
#
# Port allocation (PRODUCT level: SERVICE + 10000):
#   {PS-ID-1} (SERVICE {PORT_BASE_1}-{PORT_BASE_1+3}):
#   - {PS-ID-1}-app-sqlite-1:      {PORT_BASE_1 + 10000}
#   ...
#
include:
  - path: ../{PS-ID-1}/compose.yml
  - path: ../{PS-ID-2}/compose.yml

services:
  # PRODUCT-level PKI init (overrides PS-ID cert material).
  pki-init:
    image: cryptoutil-{PS-ID-1}:dev
    command: ["init", "--output-dir=/certs", "--domain={PRODUCT}"]
    volumes:
      - ./certs/:/certs/:rw
    depends_on:
      builder-{PS-ID-1}:
        condition: service_completed_successfully

  # Port overrides: SERVICE ports + 10000
  {PS-ID-1}-app-sqlite-1:
    image: cryptoutil-{PS-ID-1}:dev
    ports: !override
      - "{PORT_BASE_1 + 10000}:8080"
  # ... repeat for all 4 instances of each PS-ID ...

secrets:
  # PRODUCT-level secrets (override PS-ID secrets).
  unseal-1of5.secret:
    file: ./secrets/unseal-1of5.secret # pragma: allowlist secret
  # ... all 14 secrets with PRODUCT-level values ...
  # Plus 4 .secret.never marker files for browser/service credentials
```

### G.1 Product Compose Rules

| Rule ID | Rule | Rationale |
|---------|------|-----------|
| PC-01 | MUST include all child PS-ID compose files | Recursive architecture |
| PC-02 | MUST override `pki-init` with `--domain={PRODUCT}` | Product-scoped certs |
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

```yaml
# $schema: https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json
#
# CRYPTOUTIL SUITE Deployment (SUITE-Level)
#
# Purpose: Complete suite deployment with ALL 10 services across 5 products.
# Port formula: SUITE_PORT = SERVICE_BASE + 20000
#
include:
  - path: ../sm/compose.yml
  - path: ../jose/compose.yml
  - path: ../pki/compose.yml
  - path: ../identity/compose.yml
  - path: ../skeleton/compose.yml

services:
  # SUITE-level PKI init.
  pki-init:
    image: cryptoutil-sm-kms:dev
    command: ["init", "--output-dir=/certs", "--domain=cryptoutil"]
    volumes:
      - ./certs/:/certs/:rw
    depends_on:
      builder-sm-kms:
        condition: service_completed_successfully

  # Port overrides: SERVICE + 20000 (compact inline syntax).
  sm-kms-app-sqlite-1: {ports: !override ["28000:8080"]}
  sm-kms-app-sqlite-2: {ports: !override ["28001:8080"]}
  # ... repeat for all 40 app instances (10 PS-IDs × 4 instances) ...

secrets:
  # SUITE-level secrets (override PRODUCT/PS-ID secrets).
  unseal-1of5.secret:
    file: ./secrets/unseal-1of5.secret # pragma: allowlist secret
  # ... all secrets with SUITE-level values ...
```

### I.1 Suite Compose Rules

| Rule ID | Rule | Rationale |
|---------|------|-----------|
| SU-01 | MUST include all 5 product compose files | Complete suite |
| SU-02 | Port formula: `SERVICE + 20000` | Suite tier offset |
| SU-03 | Compact inline port override syntax for 40 services | Readability |
| SU-04 | Unseal secrets MUST use `cryptoutil-unseal-key-N-of-5-...` values | Suite-scoped encryption |

---

## J. Suite Dockerfile Template

The suite Dockerfile builds the `cryptoutil` binary that can run any service via subcommands.

```dockerfile
#############################################################################################
# cryptoutil Suite Dockerfile — Full Suite Binary
#
# Follows the same 4-stage pattern as PS-ID Dockerfiles.
# Binary: ./cmd/cryptoutil → /app/cryptoutil
#############################################################################################
# [Same 4-stage pattern as Section B, substituting:]
#   - {PS-ID} → cryptoutil
#   - ./cmd/{PS-ID} → ./cmd/cryptoutil
#   - LABEL title → cryptoutil
#   - HEALTHCHECK → /app/cryptoutil livez || exit 1
#   - ENTRYPOINT → ["/sbin/tini", "--", "/app/cryptoutil"]
```

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

### N.1 New cicd-lint Fitness Linters Needed

| Linter Name | Scope | Purpose |
|-------------|-------|---------|
| `dockerfile_structure` | `deployments/*/Dockerfile` | Enforce 4-stage pattern (validation, builder, runtime-deps, final) |
| `dockerfile_binary_name` | `deployments/*/Dockerfile` | Verify `go build -o /app/{PS-ID} ./cmd/{PS-ID}` |
| `dockerfile_user` | `deployments/*/Dockerfile` | Verify `USER 65532:65532` not commented out |
| `dockerfile_entrypoint` | `deployments/*/Dockerfile` | Verify `ENTRYPOINT ["/sbin/tini", "--", "/app/{PS-ID}"]` |
| `dockerfile_no_cmd` | `deployments/*/Dockerfile` | Verify no `CMD` instruction (compose controls) |
| `dockerfile_no_curl` | `deployments/*/Dockerfile` | Verify final stage has no `apk add curl` |
| `dockerfile_workdir` | `deployments/*/Dockerfile` | Verify `WORKDIR /app` (not `/app/run`) |
| `dockerfile_no_goenv` | `deployments/*/Dockerfile` | Verify no `GOMODCACHE`/`GOCACHE` env vars |
| `config_key_naming` | `configs/**/*.yml`, `deployments/*/config/*.yml` | Enforce all YAML keys are kebab-case |
| `config_header_identity` | `configs/**/*.yml`, `deployments/*/config/*.yml` | Verify header comments reference correct PS-ID |
| `config_instance_minimal` | `deployments/*/config/*-{variant}.yml` | Verify instance configs only contain cors-origins, otlp-service, otlp-hostname, database-url |
| `config_common_complete` | `deployments/*/config/*-common.yml` | Verify common configs contain all required shared keys |

### N.2 Existing Linter Enhancements

| Linter | Enhancement |
|--------|-------------|
| `dockerfile_labels` | Add rule: compact multi-line LABEL (not individual lines) |
| `dockerfile_healthcheck` | Already enforced — no changes needed |
| `compose_service_names` | Add rule: exactly 4 app instances per PS-ID |
| `standalone_config_presence` | Add rule: validate key naming is kebab-case |

### N.3 Enforcement Priority

1. **P0 (BLOCKING)**: Fix identity-spa COPY bug (runtime failure)
2. **P0 (BLOCKING)**: Fix skeleton-template Dockerfile (copy-paste from jose-ja)
3. **P1 (HIGH)**: Standardize all 10 Dockerfiles to canonical template
4. **P1 (HIGH)**: Enable `USER 65532:65532` in all Dockerfiles (currently commented out in 5)
5. **P2 (MEDIUM)**: Migrate snake_case configs to kebab-case (7 services)
6. **P2 (MEDIUM)**: Standardize deployment config overlay structure
7. **P3 (LOW)**: Implement enforcement linters
8. **P3 (LOW)**: Fix suite Dockerfile (add tini)

---

## O. Cross-References

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
