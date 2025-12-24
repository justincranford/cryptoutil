# Docker and Docker Compose Configuration - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/04-02.docker.instructions.md`

## Core Docker Compose Rules

### Command Syntax

**ALWAYS use `docker compose` (NOT `docker-compose`)**:
- `docker-compose` is legacy v1 (deprecated)
- `docker compose` is v2 (current standard)

### Path Handling - CRITICAL

**NEVER use absolute paths in compose.yml - breaks cross-platform compatibility**

**WRONG**:
```yaml
secrets:
  postgres_user:
    file: /home/user/project/secrets/postgres_user.secret  # ❌ Breaks on Windows
```

**CORRECT**:
```yaml
secrets:
  postgres_user:
    file: ./postgres/postgres_user.secret  # ✅ Relative paths work everywhere
```

---

## Multi-Stage Dockerfile Patterns

### Global ARGs and Stage Redeclaration

**Pattern**:
```dockerfile
# Global ARGs at top for visibility
ARG GO_VERSION=1.25.5
ARG VCS_REF
ARG BUILD_DATE

# Builder stage
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
# Redeclare ARGs for use in LABELs
ARG VCS_REF
ARG BUILD_DATE
# Build logic...

# Validation stage
FROM alpine:3.19 AS validator
# Validation logic...

# Final runtime stage
FROM alpine:3.19 AS runtime
WORKDIR /app
# Redeclare ARGs for LABELs on published image
ARG VCS_REF
ARG BUILD_DATE
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
# Copy from validator, not builder
COPY --from=validator /app/cryptoutil /app/cryptoutil
```

**WORKDIR Standards**:
- Builder stage: `/src` (source code compilation)
- Runtime stage: `/app` (application execution)

**LABEL Standards**:
- ALL LABELs on final published image only (NOT builder/validator stages)
- Enforces required ARGs (VCS_REF, BUILD_DATE) via validation stage

---

## Docker Secrets Management - CRITICAL

### Interoperability Requirements

**NEVER modify Docker Compose secrets - breaks cryptographic key hierarchy**:
- ALL cryptoutil instances MUST use SAME unseal secrets for interoperability
- Changing unseal secrets breaks deterministic key derivation (HKDF)
- Each service instance generates same unseal JWKs from same secrets

### Secret File Permissions - MANDATORY

**All secrets files MUST have 440 permissions (r--r-----)**:

```bash
# Set correct permissions
chmod 440 deployments/compose/*/secrets/*.secret

# Verify permissions
ls -la deployments/compose/*/secrets/
# Output should show: -r--r----- for all .secret files
```

**Rationale**:
- Prevents unauthorized access to secrets
- Allows group read (Docker daemon group)
- Owner read-only prevents accidental modification

### Secret Usage Pattern

**ALWAYS use secrets with file:// URLs, NOT environment variables**:

```yaml
services:
  app:
    secrets:
      - database_url_secret
      - unseal_key_secret
    command:
      - "app"
      - "--database-url=file:///run/secrets/database_url_secret"
      - "--unseal-key=file:///run/secrets/unseal_key_secret"

secrets:
  database_url_secret:
    file: ./secrets/database_url.secret
  unseal_key_secret:
    file: ./secrets/unseal_key.secret
```

### Dockerfile Secrets Validation Job - MANDATORY

**ALL Dockerfiles MUST include validation stage**:

```dockerfile
# Validation stage - verify secrets exist with correct permissions
FROM alpine:3.19 AS validator
WORKDIR /validation

# Copy secrets from builder stage (if applicable) or expect at runtime
COPY --from=builder /run/secrets/ /run/secrets/ 2>/dev/null || true

# Validate secrets existence and permissions
RUN echo "🔍 Validating Docker secrets..." && \
    ls -la /run/secrets/ || echo "⚠️ No secrets found (OK for build-time, required at runtime)" && \
    if [ -d /run/secrets/ ]; then \
        echo "✅ Secrets directory exists"; \
        for secret in database_url_secret unseal_key_secret tls_cert_secret tls_key_secret; do \
            if [ -f "/run/secrets/$secret" ]; then \
                echo "✅ Secret $secret exists"; \
                chmod 440 "/run/secrets/$secret" 2>/dev/null || true; \
            fi; \
        done; \
    fi

# Final runtime stage
FROM alpine:3.19 AS runtime
# Copy from validator, not builder
COPY --from=validator /app/cryptoutil /app/cryptoutil
COPY --from=validator /run/secrets/ /run/secrets/
```

**Enforcement**: CI/CD workflows SHOULD validate Dockerfile includes secrets validation job.

---

## Networking Configuration

### Localhost vs 127.0.0.1 - CRITICAL

**ALWAYS use `127.0.0.1` in containers (NOT `localhost`)**:
- Alpine Linux resolves `localhost` to IPv6 `::1`
- IPv4 services bound to `0.0.0.0` or `127.0.0.1` won't accept IPv6 connections
- Use explicit IPv4 address `127.0.0.1` to avoid resolution issues

**Example** (healthcheck):
```yaml
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/v1/livez"]
  start_period: 10s
  interval: 5s
  retries: 5
```

### wget vs curl in Alpine

**ALWAYS use `wget` for healthchecks** (available in Alpine by default):
- `curl` requires installation in Alpine
- `wget` is lighter and pre-installed
- Use `--no-check-certificate` for self-signed TLS certs

---

## Sidecar Health Checks

### Pattern for Containers Without Shell

**For minimal containers (e.g., otel-collector-contrib)**:

```yaml
services:
  otel-collector:
    image: otel/opentelemetry-collector-contrib:latest
    # No healthcheck (no shell/wget available)

  otel-collector-health-check:
    image: alpine:latest
    command: ["wget", "--quiet", "--tries=1", "--spider", "http://otel-collector:13133/"]
    depends_on:
      otel-collector:
        condition: service_started
    restart: on-failure
```

**Why**: Minimal images lack healthcheck utilities; separate Alpine sidecar performs checks

---

## Docker Compose Latency Hiding Strategies - CRITICAL

### Strategy 1: Single Build, Shared Image

**Build once, reuse image for multiple instances**:

```yaml
services:
  builder:
    build:
      context: ./
      dockerfile: Dockerfile
    image: cryptoutil:local  # Tagged image for reuse

  cryptoutil-postgres-1:
    image: cryptoutil:local  # Reuses built image
    depends_on:
      builder:
        condition: service_completed_successfully

  cryptoutil-postgres-2:
    image: cryptoutil:local  # Reuses built image
    depends_on:
      builder:
        condition: service_completed_successfully
```

**Impact**: Prevents 3× build time (60s → 60s instead of 60s × 3 = 180s)

### Strategy 2: Schema Initialization by First Instance

**First instance initializes DB, others wait**:

```yaml
services:
  cryptoutil-postgres-1:
    depends_on:
      postgres:
        condition: service_healthy  # Wait for DB ready

  cryptoutil-postgres-2:
    depends_on:
      cryptoutil-postgres-1:
        condition: service_healthy  # Wait for schema init

  cryptoutil-postgres-3:
    depends_on:
      cryptoutil-postgres-1:
        condition: service_healthy  # Wait for schema init
```

**Impact**: Prevents schema initialization race conditions (3× parallel → sequential first, then parallel)

### Strategy 3: Health Check Dependencies

**Services start ONLY after dependencies are healthy**:

```yaml
services:
  cryptoutil-sqlite:
    healthcheck:
      test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/v1/livez"]
      start_period: 10s
      interval: 5s
      retries: 5

  otel-collector:
    depends_on:
      cryptoutil-sqlite:
        condition: service_healthy  # Wait for app ready, not just started
```

**Impact**: Prevents cascading failures from premature startup

### Strategy 4: Expected Startup Times

| Service | Expected Time | Strategy |
|---------|--------------|----------|
| builder | 30-60s | One-time build, cached for all instances |
| postgres | 5-30s | start_period=5s + (5s×5 retries) = max 30s |
| cryptoutil (first) | 10-35s | start_period=10s + (5s×5 retries) + unseal |
| cryptoutil (others) | 5-15s | Schema already initialized, just unseal |
| otel-collector | 10-40s | Waits for cryptoutil, 10s sleep + 15 retries |

**Total Expected**: 60-150s for full stack in optimal conditions
**GitHub Actions**: Add 50-100% margin for shared CPU, network latency, cold starts

### Strategy 5: Diagnostic Logging for Bottlenecks

**Dockerfile timing**:

```dockerfile
RUN echo "🏗️ Build started: $(date -u +'%Y-%m-%d %H:%M:%S UTC')" && \
    go build -o /app/cryptoutil ./cmd/cryptoutil && \
    echo "✅ Build completed: $(date -u +'%Y-%m-%d %H:%M:%S UTC')"
```

**Healthcheck timing**:

```yaml
healthcheck:
  test: |
    wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9090/admin/v1/livez && \
    echo "✅ Health check passed: $(date -u +'%Y-%m-%d %H:%M:%S UTC')"
```

**Rationale**: Identify slow stages (build, schema init, network, health checks) for optimization

---

## Service Ports Quick Reference

| Service | Public API | Admin API | Backend |
|---------|-----------|-----------|---------|
| kms-sm-sqlite | 8080 | 9090 | SQLite in-memory |
| kms-sm-postgres-1 | 8081 | 9090 | PostgreSQL (APP) |
| kms-sm-postgres-2 | 8082 | 9090 | PostgreSQL (APP) |
| kms-sm-postgres | 5432 | - | PostgreSQL (DB) |
| otel-collector | 4317 (gRPC), 4318 (HTTP) | 13133 | - |
| grafana-otel-lgtm | 3000 | - | Loki/Tempo/Prometheus |

---

## Configuration Files Organization

**Shared Configuration** (`cryptoutil-common.yml`):
- TLS certificates and keys
- Unseal secrets
- Security policies
- **Affects ALL instances** (common settings)

**Instance-Specific Configuration** (`cryptoutil-sqlite.yml`, `cryptoutil-postgresql-{1,2}.yml`):
- CORS allowed origins (unique per instance)
- OTLP service name (unique per instance, matches compose service name)
- OTLP hostname (unique per instance)
- Bind addresses/ports (unique per instance)

**CRITICAL Rule**: Instance config values MUST be unique and match service name in compose.yml

**Example**:
```yaml
# cryptoutil-postgresql-1.yml
observability:
  otlp:
    service_name: "cryptoutil-postgresql-1"  # MUST match compose service name
    service_hostname: "cryptoutil-postgresql-1"  # MUST match compose service name

# cryptoutil-postgresql-2.yml
observability:
  otlp:
    service_name: "cryptoutil-postgresql-2"  # Different from instance 1
    service_hostname: "cryptoutil-postgresql-2"  # Different from instance 1
```
