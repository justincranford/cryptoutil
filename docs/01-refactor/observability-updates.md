# Observability Updates Plan

## Executive Summary

Update OpenTelemetry service names, Grafana dashboards, and telemetry configuration to reflect new KMS, identity, and CA service group structure.

**Status**: Planning
**Dependencies**: Tasks 10-17 (service extractions, CLI restructure, workflow updates, importas migration complete)
**Risk Level**: Low (configuration changes, no functional impact)

## Current OTLP Service Names

From `deployments/compose/cryptoutil/configs/`:

### KMS Services (3 instances)

**cryptoutil-sqlite.yml**:
```yaml
otlp-service: cryptoutil-sqlite
otlp-hostname: cryptoutil-sqlite
```

**cryptoutil-postgresql-1.yml**:
```yaml
otlp-service: cryptoutil-postgresql-1
otlp-hostname: cryptoutil-postgresql-1
```

**cryptoutil-postgresql-2.yml**:
```yaml
otlp-service: cryptoutil-postgresql-2
otlp-hostname: cryptoutil-postgresql-2
```

**Issue**: Service names use generic "cryptoutil" prefix instead of service-group-specific "kms" prefix

## Target OTLP Service Names

### KMS Services (New Naming)

**cryptoutil-sqlite.yml** → **kms-sqlite.yml**:
```yaml
otlp-service: kms-server-sqlite
otlp-hostname: kms-server-sqlite
```

**cryptoutil-postgresql-1.yml** → **kms-postgresql-1.yml**:
```yaml
otlp-service: kms-server-postgresql-1
otlp-hostname: kms-server-postgresql-1
```

**cryptoutil-postgresql-2.yml** → **kms-postgresql-2.yml**:
```yaml
otlp-service: kms-server-postgresql-2
otlp-hostname: kms-server-postgresql-2
```

### Identity Services (Future)

**identity-authz.yml**:
```yaml
otlp-service: identity-authz-server
otlp-hostname: identity-authz-server
```

**identity-idp.yml**:
```yaml
otlp-service: identity-idp-server
otlp-hostname: identity-idp-server
```

**identity-rs.yml**:
```yaml
otlp-service: identity-rs-server
otlp-hostname: identity-rs-server
```

**identity-spa-rp.yml**:
```yaml
otlp-service: identity-spa-rp-server
otlp-hostname: identity-spa-rp-server
```

### CA Services (Future - Skeleton)

**ca-server.yml**:
```yaml
otlp-service: ca-server
otlp-hostname: ca-server
```

## Implementation Phases

### Phase 1: Update KMS Configuration Files

**Rename config files**:

```bash
# Rename config files to match service group
git mv deployments/compose/cryptoutil/configs/cryptoutil-common.yml deployments/compose/cryptoutil/configs/kms-common.yml
git mv deployments/compose/cryptoutil/configs/cryptoutil-sqlite.yml deployments/compose/cryptoutil/configs/kms-sqlite.yml
git mv deployments/compose/cryptoutil/configs/cryptoutil-postgresql-1.yml deployments/compose/cryptoutil/configs/kms-postgresql-1.yml
git mv deployments/compose/cryptoutil/configs/cryptoutil-postgresql-2.yml deployments/compose/cryptoutil/configs/kms-postgresql-2.yml
```

**Update config file content**:

```yaml
# deployments/compose/cryptoutil/configs/kms-sqlite.yml

# CRITICAL: This file contains UNIQUE settings for the 'kms-sqlite' service in compose.yml
# ALL settings here MUST BE UNIQUE and CORRESPOND TO the service name in compose.yml

# CORS configuration - HTTPS origins only
cors-origins:
  - "https://localhost:8080"
  - "https://127.0.0.1:8080"
  - "https://[::1]:8080"
  - "https://[::ffff:127.0.0.1]:8080"

otlp-service: kms-server-sqlite      # CHANGED from cryptoutil-sqlite
otlp-hostname: kms-server-sqlite     # CHANGED from cryptoutil-sqlite

# Development mode - enables in-memory SQLite
dev: true
```

**Update kms-postgresql-1.yml**:
```yaml
otlp-service: kms-server-postgresql-1
otlp-hostname: kms-server-postgresql-1
```

**Update kms-postgresql-2.yml**:
```yaml
otlp-service: kms-server-postgresql-2
otlp-hostname: kms-server-postgresql-2
```

### Phase 2: Update Docker Compose Service Configuration

**Update `deployments/compose/compose.yml` volume mounts**:

```yaml
services:
  cryptoutil-sqlite:
    image: cryptoutil:latest
    container_name: cryptoutil-sqlite
    command: ["kms", "server", "start", "--config", "/app/configs/kms-sqlite.yml"]
    volumes:
      # OLD
      # - ./cryptoutil/configs/cryptoutil-common.yml:/app/configs/cryptoutil-common.yml:ro
      # - ./cryptoutil/configs/cryptoutil-sqlite.yml:/app/configs/cryptoutil-sqlite.yml:ro

      # NEW
      - ./cryptoutil/configs/kms-common.yml:/app/configs/kms-common.yml:ro
      - ./cryptoutil/configs/kms-sqlite.yml:/app/configs/kms-sqlite.yml:ro

  cryptoutil-postgres-1:
    volumes:
      - ./cryptoutil/configs/kms-common.yml:/app/configs/kms-common.yml:ro
      - ./cryptoutil/configs/kms-postgresql-1.yml:/app/configs/kms-postgresql-1.yml:ro

  cryptoutil-postgres-2:
    volumes:
      - ./cryptoutil/configs/kms-common.yml:/app/configs/kms-common.yml:ro
      - ./cryptoutil/configs/kms-postgresql-2.yml:/app/configs/kms-postgresql-2.yml:ro
```

**Update command argument**:

```yaml
services:
  cryptoutil-sqlite:
    command: ["kms", "server", "start", "--config", "/app/configs/kms-sqlite.yml"]  # CHANGED path

  cryptoutil-postgres-1:
    command: ["kms", "server", "start", "--config", "/app/configs/kms-postgresql-1.yml"]

  cryptoutil-postgres-2:
    command: ["kms", "server", "start", "--config", "/app/configs/kms-postgresql-2.yml"]
```

### Phase 3: Update OpenTelemetry Collector Configuration

**Current `deployments/compose/otel/otel-collector-config.yaml`**:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024

exporters:
  otlp/grafana:
    endpoint: http://grafana-otel-lgtm:4317
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/grafana]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/grafana]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/grafana]
```

**No changes needed** - OTEL Collector configuration is service-agnostic (accepts telemetry from any service)

### Phase 4: Update Grafana Dashboards (If Exists)

**Dashboard service filter updates**:

If Grafana dashboards exist with hardcoded service names:

```json
// OLD dashboard filter
{
  "name": "service_name",
  "query": "cryptoutil-sqlite|cryptoutil-postgresql-1|cryptoutil-postgresql-2"
}

// NEW dashboard filter
{
  "name": "service_name",
  "query": "kms-server-sqlite|kms-server-postgresql-1|kms-server-postgresql-2"
}
```

**Dashboard panels**:

```json
// OLD panel query
{
  "expr": "rate(http_requests_total{service=\"cryptoutil-sqlite\"}[5m])"
}

// NEW panel query
{
  "expr": "rate(http_requests_total{service=\"kms-server-sqlite\"}[5m])"
}
```

**Note**: If no custom dashboards exist, Grafana will auto-discover new service names via OTLP metrics

### Phase 5: Update Identity Configuration Files (Future)

**Create identity service config files**:

```yaml
# deployments/compose/identity/configs/identity-authz.yml

server:
  name: identity-authz-server
  bind_address: 0.0.0.0
  port: 9090

otlp-service: identity-authz-server
otlp-hostname: identity-authz-server

database:
  type: postgresql
  dsn: postgres://user:pass@postgres:5432/identity?sslmode=disable
```

**Repeat for identity-idp, identity-rs, identity-spa-rp**

### Phase 6: Testing & Validation

**Test OTLP service name propagation**:

```bash
# Start services
docker compose -f deployments/compose/compose.yml up -d

# Check OTLP service names in Grafana
# Navigate to http://127.0.0.1:3000
# Explore → Metrics → Filter by service.name

# Expected service names:
# - kms-server-sqlite
# - kms-server-postgresql-1
# - kms-server-postgresql-2
```

**Validation checklist**:
- [ ] Config files renamed correctly
- [ ] Docker Compose services start successfully
- [ ] OTLP telemetry received by collector
- [ ] Grafana shows new service names in service.name label
- [ ] Traces, metrics, logs tagged with correct service names
- [ ] No telemetry data loss during migration

### Phase 7: Documentation Updates

**Update README.md**:

```markdown
## Observability

cryptoutil services export telemetry to OpenTelemetry Collector:

### Service Names

- `kms-server-sqlite` - KMS server (SQLite backend)
- `kms-server-postgresql-1` - KMS server instance 1 (PostgreSQL backend)
- `kms-server-postgresql-2` - KMS server instance 2 (PostgreSQL backend)

Future services:
- `identity-authz-server` - OAuth 2.1 Authorization Server
- `identity-idp-server` - OIDC Identity Provider
- `identity-rs-server` - Resource Server
- `identity-spa-rp-server` - SPA Relying Party
- `ca-server` - Certificate Authority

### Grafana Access

Access Grafana UI at http://127.0.0.1:3000 (admin/admin)

Filter by service: `service.name="kms-server-sqlite"`
```

**Update observability documentation**:

```markdown
# docs/observability.md

## OTLP Service Naming Convention

All cryptoutil services follow the pattern: `<service-group>-<service>-<instance>`

Examples:
- `kms-server-sqlite` (service-group=kms, service=server, instance=sqlite)
- `kms-server-postgresql-1` (service-group=kms, service=server, instance=postgresql-1)
- `identity-authz-server` (service-group=identity, service=authz-server)
- `ca-server` (service-group=ca, service=server)

### Telemetry Flow

```
KMS Services → OTEL Collector (4317/4318) → Grafana LGTM (14317/14318)
Identity Services → OTEL Collector → Grafana LGTM
CA Services → OTEL Collector → Grafana LGTM
```

### Grafana Dashboard Filters

Filter by service group:
```
service.name =~ "kms-.*"        # All KMS services
service.name =~ "identity-.*"   # All identity services
service.name =~ "ca-.*"         # All CA services
```
```

## Configuration File Naming Conventions

### KMS Service Group

**Config directory**: `deployments/compose/cryptoutil/configs/`

| File | Purpose | Service |
|------|---------|---------|
| `kms-common.yml` | Shared KMS settings (all instances) | N/A |
| `kms-sqlite.yml` | SQLite-specific settings | cryptoutil-sqlite |
| `kms-postgresql-1.yml` | PostgreSQL instance 1 settings | cryptoutil-postgres-1 |
| `kms-postgresql-2.yml` | PostgreSQL instance 2 settings | cryptoutil-postgres-2 |

### Identity Service Group (Future)

**Config directory**: `deployments/compose/identity/configs/`

| File | Purpose | Service |
|------|---------|---------|
| `identity-common.yml` | Shared identity settings | N/A |
| `identity-authz.yml` | OAuth 2.1 Authorization Server | identity-authz |
| `identity-idp.yml` | OIDC Identity Provider | identity-idp |
| `identity-rs.yml` | Resource Server | identity-rs |
| `identity-spa-rp.yml` | SPA Relying Party | identity-spa-rp |

### CA Service Group (Future - Skeleton)

**Config directory**: `deployments/compose/ca/configs/`

| File | Purpose | Service |
|------|---------|---------|
| `ca-common.yml` | Shared CA settings | N/A |
| `ca-server.yml` | CA server settings | ca-server |

## OTLP Service Name Mapping

### Current vs New Service Names

| Docker Service | Old OTLP Service Name | New OTLP Service Name | Config File |
|----------------|----------------------|----------------------|-------------|
| cryptoutil-sqlite | cryptoutil-sqlite | kms-server-sqlite | kms-sqlite.yml |
| cryptoutil-postgres-1 | cryptoutil-postgresql-1 | kms-server-postgresql-1 | kms-postgresql-1.yml |
| cryptoutil-postgres-2 | cryptoutil-postgresql-2 | kms-server-postgresql-2 | kms-postgresql-2.yml |

**Rationale**: New names align with service group taxonomy (kms, identity, ca) for consistent telemetry filtering

## Risk Assessment

### Low Risks

1. **OTLP Service Name Change**
   - Mitigation: No breaking changes in OTLP protocol (just metadata)
   - Impact: Historical telemetry uses old names, new telemetry uses new names
   - Coexistence: Both old and new service names visible in Grafana during transition

2. **Config File Renaming**
   - Mitigation: Docker Compose volume mounts updated atomically
   - Rollback: Revert config file renames and volume mount changes

3. **Dashboard Filters**
   - Mitigation: Use regex filters to support both old and new names during transition
   - Example: `service.name =~ "(cryptoutil|kms-server)-.*"`

## Success Metrics

- [ ] Config files renamed to kms-* pattern
- [ ] Docker Compose services start successfully
- [ ] OTLP telemetry tagged with new service names
- [ ] Grafana shows new service names (kms-server-sqlite, etc.)
- [ ] No telemetry data loss during migration
- [ ] Documentation updated with new naming conventions

## Timeline

- **Phase 1**: Update KMS configuration files (1 hour)
- **Phase 2**: Update Docker Compose service configuration (30 minutes)
- **Phase 3**: Update OpenTelemetry Collector configuration (30 minutes)
- **Phase 4**: Update Grafana dashboards (1 hour)
- **Phase 5**: Update identity configuration files (1 hour)
- **Phase 6**: Testing & validation (1 hour)
- **Phase 7**: Documentation updates (1 hour)

**Total**: 6 hours (1 day)

## Grafana Query Examples

### Filter by Service Group

```promql
# All KMS services
{service_name=~"kms-.*"}

# All identity services (future)
{service_name=~"identity-.*"}

# All CA services (future)
{service_name=~"ca-.*"}
```

### Service-Specific Queries

```promql
# KMS SQLite HTTP request rate
rate(http_requests_total{service_name="kms-server-sqlite"}[5m])

# KMS PostgreSQL 1 error rate
rate(http_errors_total{service_name="kms-server-postgresql-1"}[5m])

# Identity AuthZ token issuance rate (future)
rate(oauth2_tokens_issued_total{service_name="identity-authz-server"}[5m])
```

### Cross-Service Aggregation

```promql
# Total KMS request rate (all instances)
sum(rate(http_requests_total{service_name=~"kms-server-.*"}[5m]))

# Average KMS response time (all instances)
avg(http_request_duration_seconds{service_name=~"kms-server-.*"})
```

## OpenTelemetry Resource Attributes

### KMS Services

```yaml
service.name: kms-server-sqlite
service.namespace: cryptoutil
service.version: 1.0.0
deployment.environment: production
```

### Identity Services (Future)

```yaml
service.name: identity-authz-server
service.namespace: cryptoutil
service.version: 1.0.0
deployment.environment: production
```

### CA Services (Future)

```yaml
service.name: ca-server
service.namespace: cryptoutil
service.version: 1.0.0
deployment.environment: production
```

## Cross-References

- [Service Groups Taxonomy](service-groups.md) - Service group definitions
- [CLI Restructure](cli-restructure.md) - CLI command changes
- [Workflow Updates](workflow-updates.md) - CI/CD configuration changes
- [Observability Instructions](.github/instructions/02-03.observability.instructions.md) - Telemetry architecture

## Next Steps

After observability updates:
1. **Task 19**: Integration testing (full test suite validation)
2. **Task 20**: Documentation finalization (handoff package)
