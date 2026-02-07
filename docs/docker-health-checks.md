# Docker Health Check Analysis

## Current State Audit (13 Compose Files)

### KMS Services

1. `deployments/compose/compose.yml` - E2E testing compose
   - **cryptoutil-sqlite**: ✅ Uses curl to `/admin/api/v1/livez`
   - **cryptoutil-postgres-1**: ✅ Uses curl to `/admin/api/v1/livez`
   - **cryptoutil-postgres-2**: ✅ Uses curl to `/admin/api/v1/livez`
   - **postgres**: ✅ Uses pg_isready

2. `deployments/kms/compose.yml` - KMS production compose
   - Needs verification

### Cipher-IM Services

1. `deployments/cipher/compose.yml`
   - **cipher-im-sqlite**: ✅ Uses wget to `/admin/api/v1/livez`
   - **cipher-im-pg-1**: ✅ Uses wget to `/admin/api/v1/livez`
   - **cipher-im-pg-2**: ✅ Uses wget to `/admin/api/v1/livez`
   - **postgres**: ✅ Uses pg_isready
   - **grafana-otel-lgtm**: ✅ Uses curl to `/api/health`
   - **opentelemetry-collector-contrib**: ❌ NO HEALTHCHECK (minimal image)

### JOSE Services

1. `deployments/jose/compose.yml`
   - **jose-ja**: ⚠️ Uses wget to `/health` (WRONG - should be `/admin/api/v1/livez`)
   - **Dependencies**: opentelemetry-collector-contrib (NO HEALTHCHECK)

### Identity Services

1. `deployments/identity/*.yml` - Needs verification (covers 5-12 compose files)

## Issues Found

### Critical Issues

1. **Inconsistent Health Check Endpoints**:
   - KMS: `/admin/api/v1/livez` ✅ CORRECT
   - Cipher-IM: `/admin/api/v1/livez` ✅ CORRECT
   - JOSE: `/health` ❌ WRONG - should be `/admin/api/v1/livez`

2. **Inconsistent Tools**:
   - KMS: Uses `curl`
   - Cipher-IM: Uses `wget`
   - JOSE: Uses `wget`
   - **Recommendation**: Standardize on `wget` (available in Alpine by default)

3. **Missing Health Checks**:
   - opentelemetry-collector-contrib: No healthcheck (minimal image, no shell)

4. **Inconsistent Timing**:
   - start_period: varies (10s, 30s, 60s)
   - interval: varies (5s, 10s)
   - timeout: consistent (3-5s)
   - retries: consistent (5)

### Non-Critical Issues

1. **Port Inconsistencies**:
   - Most use admin port 9090
   - JOSE uses 9092 and checks public port 8060 (should check admin port)

## Best Practices (from Research)

### Docker Documentation

- Use lightweight commands (wget > curl for size)
- Set appropriate start_period (allow service initialization)
- Use short intervals after start_period (5-10s)
- Use 3-5s timeout
- Use 3-5 retries

### Kubernetes Health Probe Patterns

- **Liveness**: Is process alive? → `/admin/api/v1/livez`
- **Readiness**: Is service ready? → `/admin/api/v1/readyz`
- Docker healthcheck = Kubernetes liveness probe

### Alpine Linux Considerations

- `wget` is pre-installed
- `curl` requires installation (adds ~3MB to image)
- Prefer `wget` for minimal images

## Standardized Pattern

```yaml
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "--quiet",
         "--tries=1", "--spider",
         "https://127.0.0.1:9090/admin/api/v1/livez"]
  start_period: 30s    # Allow service initialization
  interval: 10s        # Check every 10s after start_period
  timeout: 5s          # Timeout after 5s
  retries: 5           # Retry 5 times before marking unhealthy
```

### PostgreSQL Pattern

```yaml
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U $(cat /run/secrets/postgres_username.secret) -d $(cat /run/secrets/postgres_database.secret)"]
  start_period: 10s
  interval: 10s
  timeout: 3s
  retries: 5
```

### Grafana OTEL LGTM Pattern

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://127.0.0.1:3000/api/health"]
  start_period: 15s
  interval: 30s
  timeout: 10s
  retries: 3
```

### OpenTelemetry Collector (No Health Check)

OpenTelemetry Collector Contrib uses a minimal image with no shell/curl/wget.
Use `service_started` condition instead of `service_healthy`.

```yaml
depends_on:
  opentelemetry-collector-contrib:
    condition: service_started  # NOT service_healthy
```

## Recommendations

### Phase 8.5.2-8.5.8 Implementation Priority

1. **Task 8.5.2**: Update all KMS services to use wget + standard timings
2. **Task 8.5.3**: Fix JOSE health check endpoint (use `/admin/api/v1/livez`)
3. **Task 8.5.4**: Audit and fix Identity services (13 files)
4. **Task 8.5.5**: Standardize PostgreSQL health checks
5. **Task 8.5.6**: Fix cipher-im E2E health checks (already correct?)
6. **Task 8.5.7**: Update documentation
7. **Task 8.5.8**: Verify all services with E2E tests

### Standardization Checklist

For each application service:
- ✅ Use `wget` (not `curl`)
- ✅ Check `/admin/api/v1/livez` (not `/health` or public endpoints)
- ✅ Use `--no-check-certificate` (self-signed TLS certs)
- ✅ Use `--quiet --tries=1 --spider` (minimal output)
- ✅ Use `127.0.0.1:9090` (admin port, NOT public port)
- ✅ Use `start_period: 30s` (allow initialization)
- ✅ Use `interval: 10s` (check frequency)
- ✅ Use `timeout: 5s` (command timeout)
- ✅ Use `retries: 5` (failure threshold)

For PostgreSQL:
- ✅ Use `pg_isready` with dynamic credentials from secrets
- ✅ Use `start_period: 10s`
- ✅ Use `interval: 10s`

For Grafana OTEL LGTM:
- ✅ Use `curl` (HTTP endpoint, not HTTPS)
- ✅ Check `http://127.0.0.1:3000/api/health`
- ✅ Use `start_period: 15s`

For OpenTelemetry Collector:
- ✅ NO healthcheck (minimal image)
- ✅ Use `depends_on: condition: service_started`
