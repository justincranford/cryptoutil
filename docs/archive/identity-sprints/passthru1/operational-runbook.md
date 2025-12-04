# Identity Services - Operational Runbook

## Table of Contents

1. [Deployment Procedures](#deployment-procedures)
2. [Health Checks](#health-checks)
3. [Graceful Shutdown](#graceful-shutdown)
4. [Troubleshooting](#troubleshooting)
5. [Monitoring and Metrics](#monitoring-and-metrics)
6. [Database Operations](#database-operations)
7. [Backup and Recovery](#backup-and-recovery)
8. [Performance Tuning](#performance-tuning)

---

## Deployment Procedures

### Docker Compose Deployment

#### Standard Deployment

```bash
# Navigate to compose directory
cd deployments/compose

# Start all services
docker compose -f identity-compose.yml up -d

# Verify services started
docker compose -f identity-compose.yml ps

# Check logs
docker compose -f identity-compose.yml logs -f
```

#### Build and Deploy

```bash
# Build and deploy in one command
docker compose -f identity-compose.yml up -d --build

# Build specific service
docker compose -f identity-compose.yml build identity-authz

# Restart specific service
docker compose -f identity-compose.yml restart identity-authz
```

#### Service Startup Order

The compose file handles dependency ordering automatically:

1. **PostgreSQL** starts first (healthcheck: `pg_isready`)
2. **AuthZ Server** starts after PostgreSQL is healthy
3. **IdP Server** starts after PostgreSQL and AuthZ are healthy
4. **RS Server** starts after AuthZ is healthy
5. **SPA-RP** starts after AuthZ and IdP are healthy

**Verification**: All services should reach "healthy" status within 30 seconds.

### Standalone Binary Deployment

#### AuthZ Server

```bash
# Build
go build -o authz ./cmd/identity/authz

# Run with config
./authz --config=configs/identity/production.yml
```

#### IdP Server

```bash
# Build
go build -o idp ./cmd/identity/idp

# Run with config
./idp --config=configs/identity/production.yml
```

#### Resource Server

```bash
# Build
go build -o rs ./cmd/identity/rs

# Run with config
./rs --config=configs/identity/production.yml
```

---

## Health Checks

### Endpoints

#### Public Health Endpoints

| Service | Endpoint | Expected Response |
|---------|----------|-------------------|
| AuthZ | `GET http://127.0.0.1:8080/health` | `{"status":"ok","service":"authz"}` (planned) |
| IdP | `GET http://127.0.0.1:8081/health` | `{"status":"ok","service":"idp"}` (planned) |
| RS | `GET http://127.0.0.1:8082/api/v1/public/health` | `{"status":"ok"}` |

#### Admin Health Endpoints (Internal Only)

All servers expose admin endpoints on 127.0.0.1:9090 (not exposed outside containers):

```bash
# Inside container
wget --no-check-certificate -q -O - https://127.0.0.1:9090/livez
wget --no-check-certificate -q -O - https://127.0.0.1:9090/readyz
```

### Docker Compose Health Checks

```bash
# Check all services health status
docker compose -f deployments/compose/identity-compose.yml ps

# Expected output:
# identity-postgres    healthy
# identity-authz       healthy
# identity-idp         healthy
# identity-rs          healthy
# identity-spa-rp      healthy
```

### Manual Health Verification

```bash
# Test AuthZ server (planned)
curl http://127.0.0.1:8080/health

# Test IdP server (planned)
curl http://127.0.0.1:8081/health

# Test RS server
curl http://127.0.0.1:8082/api/v1/public/health

# Test PostgreSQL
docker compose -f deployments/compose/identity-compose.yml exec identity-postgres \
    pg_isready -U identity_user -d identity_db
```

---

## Graceful Shutdown

### Docker Compose Shutdown

```bash
# Graceful shutdown (sends SIGTERM, waits for clean exit)
docker compose -f deployments/compose/identity-compose.yml down

# Shutdown with volume cleanup (WARNING: deletes data)
docker compose -f deployments/compose/identity-compose.yml down -v

# Force stop (sends SIGKILL after timeout)
docker compose -f deployments/compose/identity-compose.yml down -t 3
```

### Programmatic Shutdown via API

```bash
# Trigger graceful shutdown via admin endpoint (inside container)
docker compose exec identity-authz \
    wget --no-check-certificate --post-data='' -q -O - https://127.0.0.1:9090/shutdown
```

### Process Lifecycle

**Graceful shutdown sequence:**

1. Server receives SIGTERM or shutdown API call
2. Server stops accepting new connections
3. Server waits for in-flight requests to complete (30s timeout)
4. Background jobs are stopped
5. Database connections are closed
6. Server exits with code 0

**Timeout handling:**

- Request timeout: 30 seconds (configurable)
- Shutdown timeout: 30 seconds (configurable)
- If timeout exceeded, remaining connections are forcibly terminated

---

## Troubleshooting

### Common Issues

#### Service Won't Start

**Symptom**: Container exits immediately or health check fails

**Diagnosis**:

```bash
# Check logs
docker compose -f deployments/compose/identity-compose.yml logs identity-authz

# Check container status
docker compose -f deployments/compose/identity-compose.yml ps identity-authz

# Inspect container
docker inspect identity-authz
```

**Common causes**:

- Database not ready (check PostgreSQL healthcheck)
- Port already in use (check `netstat -an | grep :8080`)
- Invalid configuration file
- Missing TLS certificates (if TLS enabled)

#### Database Connection Failures

**Symptom**: Errors like "connection refused" or "database not found"

**Diagnosis**:

```bash
# Verify PostgreSQL is running
docker compose ps identity-postgres

# Test connection from container
docker compose exec identity-authz \
    pg_isready -h identity-postgres -p 5432 -U identity_user -d identity_db

# Check PostgreSQL logs
docker compose logs identity-postgres
```

**Resolution**:

- Ensure PostgreSQL is healthy before starting application services
- Verify database DSN in configuration
- Check network connectivity between containers

#### Token Validation Failures

**Symptom**: 401 Unauthorized responses when accessing protected resources

**Diagnosis**:

1. Check token format (should be "Bearer <token>")
2. Verify token hasn't expired (check `exp` claim)
3. Confirm token was issued by correct issuer
4. Validate token signature with correct JWK

**Debugging**:

```bash
# Decode JWT token (without verification)
echo "<token>" | cut -d '.' -f 2 | base64 -d | jq

# Check issuer configuration
docker compose exec identity-authz cat /app/configs/identity/production.yml | grep issuer

# Introspect token via AuthZ API
curl -X POST http://127.0.0.1:8080/oauth2/v1/introspect \
    -d "token=<access_token>" \
    -d "client_id=<client_id>" \
    -d "client_secret=<secret>"
```

#### High Memory Usage

**Symptom**: Container OOM killed or high memory consumption

**Diagnosis**:

```bash
# Check resource usage
docker stats identity-authz

# Check memory limits
docker inspect identity-authz | jq '.[0].HostConfig.Memory'
```

**Resolution**:

- Increase memory limits in compose file (`deploy.resources.limits.memory`)
- Review token cache size configuration
- Check for memory leaks in logs
- Analyze with pprof endpoint (if enabled)

---

## Monitoring and Metrics

### Telemetry Stack

**OpenTelemetry Collector** aggregates telemetry from all services:

- Endpoint: `http://127.0.0.1:4317` (gRPC), `http://127.0.0.1:4318` (HTTP)
- Health: `http://127.0.0.1:13133/`

**Grafana OTEL LGTM** provides visualization:

- UI: `http://127.0.0.1:3000` (admin/admin)
- OTLP Receiver: `http://127.0.0.1:14317` (gRPC), `http://127.0.0.1:14318` (HTTP)

### Key Metrics to Monitor

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `http_requests_total` | Total HTTP requests | N/A |
| `http_request_duration_seconds` | Request latency | p95 > 500ms |
| `token_issuance_total` | Tokens issued | Sudden drop |
| `token_validation_failures_total` | Failed validations | > 10% error rate |
| `database_connections_active` | Active DB connections | > 80% of pool |
| `background_job_duration_seconds` | Job execution time | > 5 minutes |
| `memory_usage_bytes` | Memory consumption | > 90% of limit |

### Log Levels

Configure via environment variable or config file:

- **ERROR**: Production default (only critical issues)
- **WARN**: Important warnings
- **INFO**: General information (recommended for production)
- **DEBUG**: Detailed debugging (development only)
- **TRACE**: Very detailed debugging (performance impact)

Set log level:

```yaml
# Config file
log-level: "INFO"
```

Or environment variable:

```bash
LOG_LEVEL=INFO
```

### Accessing Logs

```bash
# Stream all logs
docker compose -f deployments/compose/identity-compose.yml logs -f

# Filter specific service
docker compose logs -f identity-authz

# Follow last 100 lines
docker compose logs --tail=100 -f identity-authz

# Export logs to file
docker compose logs identity-authz > authz-logs.txt
```

---

## Database Operations

### Database Migrations

**Automatic migrations** run on server startup (handled by GORM AutoMigrate).

**Manual migration verification:**

```bash
# Connect to database
docker compose exec identity-postgres psql -U identity_user -d identity_db

# List tables
\dt

# Describe table
\d tokens

# Check schema version (if migration tracking implemented)
SELECT * FROM schema_migrations;
```

### Database Backup

```bash
# Backup PostgreSQL database
docker compose exec identity-postgres pg_dump -U identity_user identity_db > identity_backup.sql

# Backup with timestamp
docker compose exec identity-postgres pg_dump -U identity_user identity_db > \
    "identity_backup_$(date +%Y%m%d_%H%M%S).sql"
```

### Database Restore

```bash
# Restore from backup
docker compose exec -T identity-postgres psql -U identity_user identity_db < identity_backup.sql

# Restore with docker compose down first
docker compose down
docker compose up -d identity-postgres
# Wait for PostgreSQL to be ready
docker compose exec -T identity-postgres psql -U identity_user identity_db < identity_backup.sql
docker compose up -d
```

### Database Cleanup

```bash
# Manual cleanup of expired tokens (when DeleteExpiredBefore implemented)
docker compose exec identity-postgres psql -U identity_user identity_db <<EOF
DELETE FROM tokens WHERE expires_at < NOW();
EOF

# Manual session cleanup
docker compose exec identity-postgres psql -U identity_user identity_db <<EOF
DELETE FROM sessions WHERE expires_at < NOW();
EOF
```

**Note**: Background cleanup jobs run automatically every hour.

---

## Backup and Recovery

### Backup Strategy

**What to backup:**

- PostgreSQL database (tokens, sessions, clients, users)
- Configuration files (`configs/identity/*.yml`)
- TLS certificates (if not using generated certs)
- Secrets (Docker secrets in `deployments/compose/`)

**Backup frequency:**

- **Database**: Daily automated backups
- **Configs**: Version controlled (git)
- **Secrets**: Secure backup on change

### Disaster Recovery

**Recovery Time Objective (RTO)**: < 5 minutes
**Recovery Point Objective (RPO)**: < 24 hours (daily backups)

**Recovery steps:**

1. **Restore configuration**:

   ```bash
   git pull origin main  # Or restore from backup
   cd deployments/compose
   ```

2. **Restore database**:

   ```bash
   docker compose up -d identity-postgres
   docker compose exec -T identity-postgres psql -U identity_user identity_db < backup.sql
   ```

3. **Start services**:

   ```bash
   docker compose up -d
   ```

4. **Verify health**:

   ```bash
   docker compose ps
   curl http://127.0.0.1:8080/health
   curl http://127.0.0.1:8081/health
   curl http://127.0.0.1:8082/api/v1/public/health
   ```

---

## Performance Tuning

### Resource Limits

**Current Docker Compose limits:**

| Service | Memory Limit | Memory Reservation | CPU Limit |
|---------|--------------|-------------------|-----------|
| PostgreSQL | 512M | 256M | N/A |
| AuthZ | 256M | 128M | N/A |
| IdP | 256M | 128M | N/A |
| RS | 256M | 128M | N/A |

**Tuning guidance:**

- Increase memory if OOM killed
- Increase CPU limit for high-throughput scenarios
- Monitor with `docker stats`

### Database Connection Pooling

**PostgreSQL connection pool** (GORM defaults):

- Max open connections: 100
- Max idle connections: 10
- Connection max lifetime: 1 hour

**Tuning** (in config.yml):

```yaml
database:
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s
```

### HTTP Server Tuning

**Fiber server defaults:**

- Read timeout: 30 seconds
- Write timeout: 120 seconds
- Idle timeout: 120 seconds

**Adjust in server config** (magic constants):

```go
DefaultReadTimeout  = 30
DefaultWriteTimeout = 120
DefaultIdleTimeout  = 120
```

---

## Appendix: Quick Reference

### Port Reference

| Service | Public Port | Admin Port (Internal) | Database Port |
|---------|-------------|------------------------|---------------|
| AuthZ | 8080 | 9090 | N/A |
| IdP | 8081 | 9090 | N/A |
| RS | 8082 | 9090 | N/A |
| SPA-RP | 8083 | 9090 | N/A |
| PostgreSQL | 5433 (host) | N/A | 5432 (internal) |
| OTEL Collector | 4317/4318 | 13133 | N/A |
| Grafana | 3000 | N/A | N/A |

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `LOG_LEVEL` | Logging verbosity | `INFO` |
| `OTLP_ENDPOINT` | Telemetry collector | `http://opentelemetry-collector-contrib:4317` |

### Configuration Files

- `deployments/compose/identity-compose.yml` - Docker orchestration
- `configs/identity/production.yml` - Production configuration
- `configs/identity/development.yml` - Development configuration
- `configs/identity/test.yml` - Test configuration

### Support Contacts

- **Documentation**: `docs/identityV2/`
- **Architecture**: `docs/identityV2/topology-diagram.md`
- **Task Tracking**: `docs/identityV2/identityV2_master.md`

---

**Last Updated**: November 10, 2025
**Version**: 1.0 (Task 10 Completion)
