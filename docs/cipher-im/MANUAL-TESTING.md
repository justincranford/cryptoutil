# Cipher-IM Manual Testing Guide

## Quick Start

### Prerequisites

- Docker Desktop installed and running
- Go 1.25.5+ installed
- Git bash or PowerShell

### Start Services

```bash
# From project root
cd deployments/compose/cipher-im
docker compose up -d

# Wait for services to be healthy (30-60 seconds)
docker compose ps  # All should show "healthy"
```

## Service Instances

Three cipher-im instances with different backends:

| Instance | Public API | Admin API | Database |
|----------|-----------|-----------|----------|
| cipher-im-sqlite | <https://127.0.0.1:8080> | <https://127.0.0.1:9090> | SQLite in-memory |
| cipher-im-postgres-1 | <https://127.0.0.1:8081> | <https://127.0.0.1:9091> | PostgreSQL (tenant_1 schema) |
| cipher-im-postgres-2 | <https://127.0.0.1:8082> | <https://127.0.0.1:9092> | PostgreSQL (tenant_2 schema) |

## Health Checks

### Livez (Lightweight - Process Alive)

```bash
# SQLite instance
autoapprove curl -k https://127.0.0.1:9090/admin/v1/livez

# PostgreSQL-1 instance
autoapprove curl -k https://127.0.0.1:9091/admin/v1/livez

# PostgreSQL-2 instance
autoapprove curl -k https://127.0.0.1:9092/admin/v1/livez
```

**Expected**: HTTP 200 OK

### Readyz (Heavyweight - Dependencies Healthy)

```bash
# SQLite instance
autoapprove curl -k https://127.0.0.1:9090/admin/v1/readyz

# PostgreSQL-1 instance
autoapprove curl -k https://127.0.0.1:9091/admin/v1/readyz

# PostgreSQL-2 instance
autoapprove curl -k https://127.0.0.1:9092/admin/v1/readyz
```

**Expected**: HTTP 200 OK

## Telemetry Services

### OpenTelemetry Collector

```bash
# Health check (gRPC port 4317 active)
docker compose ps opentelemetry-collector

# Logs
docker compose logs -f opentelemetry-collector
```

**Expected**: Container healthy, no errors in logs

### Grafana LGTM Stack

```bash
# Health check
curl http://127.0.0.1:3000/api/health

# Open in browser
start http://127.0.0.1:3000
```

**Expected**: HTTP 200 OK, Grafana UI accessible

**Default credentials**: admin/admin

## Client API Testing

### Create Tenant (SQLite Instance)

```bash
autoapprove curl -k -X POST https://127.0.0.1:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "test-tenant", "domain": "test.example.com"}'
```

**Expected**: HTTP 201 Created with tenant ID

### List Tenants (PostgreSQL-1 Instance)

```bash
autoapprove curl -k https://127.0.0.1:8081/api/v1/tenants
```

**Expected**: HTTP 200 OK with tenant array (may be empty initially)

### Cross-Instance Isolation Verification

```bash
# Create tenant in SQLite instance
autoapprove curl -k -X POST https://127.0.0.1:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "sqlite-tenant", "domain": "sqlite.example.com"}'

# Verify NOT visible in PostgreSQL-1 instance
autoapprove curl -k https://127.0.0.1:8081/api/v1/tenants
```

**Expected**: Tenant created in SQLite NOT returned by PostgreSQL-1 query

## Browser UI Testing

### SQLite Instance UI

```powershell
start https://127.0.0.1:8080/ui/
```

### PostgreSQL-1 Instance UI

```powershell
start https://127.0.0.1:8081/ui/
```

### PostgreSQL-2 Instance UI

```powershell
start https://127.0.0.1:8082/ui/
```

**Note**: Accept self-signed TLS certificate warnings in browser

## Telemetry Validation

### Verify Traces in Grafana

1. Open Grafana: <http://127.0.0.1:3000>
2. Navigate to **Explore** → **Tempo** (traces)
3. Filter by service: `cipher-im-sqlite` or `cipher-im-postgres-1`
4. Execute API requests via curl (creates traces)
5. Verify traces appear in Grafana

### Verify Metrics in Grafana

1. Navigate to **Explore** → **Prometheus** (metrics)
2. Query: `http_requests_total{service="cipher-im-sqlite"}`
3. Execute API requests via curl
4. Verify metric counters increment

### Verify Logs in Grafana

1. Navigate to **Explore** → **Loki** (logs)
2. Filter by service: `{job="cipher-im-sqlite"}`
3. Execute API requests via curl
4. Verify structured logs appear with request IDs, trace IDs

## Cleanup

```bash
# Stop all services
docker compose down -v

# Verify stopped
docker compose ps
```

**Expected**: All containers stopped and removed

## Troubleshooting

### Container Unhealthy

```bash
# Check container status
docker compose ps

# View logs for specific container
docker compose logs -f cipher-im-sqlite

# Restart specific container
docker compose restart cipher-im-sqlite
```

### Database Connection Errors

```bash
# Check PostgreSQL logs
docker compose logs -f postgres

# Verify database exists
docker compose exec postgres psql -U cryptoutil -l
```

### Telemetry Missing

```bash
# Check otel-collector config
docker compose exec opentelemetry-collector cat /etc/otelcol/config.yaml

# Verify receivers listening
docker compose exec opentelemetry-collector netstat -tuln | grep 4317
```

## Performance Testing

```bash
# Install Apache Bench (if not available)
# Windows: Download from https://www.apachelounge.com/download/

# Load test SQLite instance (100 requests, 10 concurrent)
ab -n 100 -c 10 -k https://127.0.0.1:8080/api/v1/tenants

# Load test PostgreSQL-1 instance
ab -n 100 -c 10 -k https://127.0.0.1:8081/api/v1/tenants
```

**Expected**: Average response time <100ms for health checks

## Integration with KMS

This docker compose pattern is identical to KMS service E2E testing:

- Multiple instances (SQLite + 2× PostgreSQL)
- Full telemetry stack (otel-collector + Grafana LGTM)
- Health check orchestration
- Cross-instance isolation verification

**Reference**: deployments/compose/kms/compose.yml
