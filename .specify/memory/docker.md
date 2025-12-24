# Docker and Docker Compose Configuration - Complete Specifications

**Referenced by**: `.github/instructions/04-02.docker.instructions.md`

## Core Docker Compose Rules

**Command**: Use `docker compose` (NOT `docker-compose` v1)
**Paths**: ALWAYS relative (NOT absolute), breaks cross-platform

---

## Multi-Stage Dockerfile Patterns

**Global ARGs**: Declare at top (GO_VERSION, VCS_REF, BUILD_DATE)
**Stage Redeclaration**: Redeclare ARGs in each stage using them
**WORKDIR**: Builder=`/src`, Runtime=`/app`
**LABELs**: Final published image only (NOT builder/validator)
**Copy**: Use validator stage, not builder (`COPY --from=validator`)

## Docker Secrets Management - CRITICAL

**Interoperability**: NEVER modify unseal secrets (breaks HKDF deterministic key derivation)
**Permissions**: chmod 440 (r--r-----) on all .secret files
**Usage**: ALWAYS `file:///run/secrets/secret_name`, NOT env vars
**Validation**: ALL Dockerfiles MUST include secrets validation stage

## Networking Configuration

**Localhost**: ALWAYS `127.0.0.1` in containers (NOT `localhost`, Alpine resolves to IPv6)
**Healthcheck**: Use `wget --no-check-certificate -q -O /dev/null https://127.0.0.1:PORT/path`
**Why wget**: Pre-installed in Alpine, lighter than curl

## Sidecar Health Checks

**For minimal containers** (no shell/wget): Separate Alpine sidecar performs checks

---

## Docker Compose Latency Hiding Strategies - CRITICAL

**Strategy 1**: Single build, shared image (prevents 3× build time)
**Strategy 2**: First instance initializes DB, others wait
**Strategy 3**: Health check dependencies (service_healthy, not service_started)
**Strategy 4**: Expected startup times (builder 30-60s, postgres 5-30s, cryptoutil 10-35s)
**Strategy 5**: Diagnostic logging with timestamps for bottleneck identification

## Service Ports Quick Reference

| Service | Public API | Admin API | Backend |
|---------|-----------|-----------|---------|
| kms-sm-sqlite | 8080 | 9090 | SQLite in-memory |
| kms-sm-postgres-1/2 | 8081/8082 | 9090 | PostgreSQL |
| otel-collector | 4317/4318 | 13133 | - |
| grafana-otel-lgtm | 3000 | - | Loki/Tempo/Prometheus |

## Configuration Files Organization

**Shared** (`cryptoutil-common.yml`): TLS certs, unseal secrets, security policies (affects ALL instances)
**Instance-Specific** (`cryptoutil-{backend}-{N}.yml`): CORS, OTLP service_name/hostname, bind addresses/ports (MUST match compose service name)
