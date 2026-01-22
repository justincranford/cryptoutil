# JOSE-JA Deployment Guide

## Overview

JOSE-JA (JOSE Authority) provides JSON Object Signing and Encryption services as a containerized microservice.

## Prerequisites

- Docker 24+ with Docker Compose v2
- PostgreSQL 18+ (production) or SQLite (development)
- OpenTelemetry Collector (optional, for telemetry)

## Directory Structure

```
deployments/jose/
├── compose.yml           # Docker Compose configuration
├── config/
│   └── jose.yml          # JOSE-JA configuration
├── Dockerfile.jose       # Docker build file

configs/jose/
└── jose-server.yml       # Default configuration template
```

## Quick Start

### Development (SQLite In-Memory)

```bash
# Start JOSE-JA with SQLite
cd deployments/jose
docker compose up -d

# Verify health
curl -k https://localhost:9092/admin/api/v1/livez

# View logs
docker compose logs -f jose-server
```

### Production (PostgreSQL)

```bash
# Start with PostgreSQL backend
cd deployments/jose
docker compose -f compose.yml -f compose.postgres.yml up -d
```

## Configuration

### Configuration Priority

Configuration sources are loaded in priority order (highest to lowest):

1. **Docker Secrets** (`file:///run/secrets/secret_name`)
2. **YAML Configuration File** (`--config=/etc/jose/jose.yml`)
3. **Command Line Flags** (`--bind-public-port=8092`)

**CRITICAL**: Environment variables are NOT used for configuration. Use Docker secrets for sensitive values.

### Server Configuration

```yaml
# Server binding
bind-public-address: "0.0.0.0"      # Container binding (use 127.0.0.1 for local dev)
bind-public-port: 8092               # Public API port
bind-admin-address: "127.0.0.1"      # Admin always localhost
bind-admin-port: 9092                # Admin API port (health checks)

# TLS configuration
tls-enabled: true
tls-cert-file: "/run/secrets/tls_cert"
tls-key-file: "/run/secrets/tls_key"
```

### Database Configuration

**SQLite (Development)**:

```yaml
database-type: "sqlite"
database-dsn: "file::memory:?cache=shared"
```

**PostgreSQL (Production)**:

```yaml
database-type: "postgresql"
database-dsn: "file:///run/secrets/database_url"  # Docker secret
```

### Telemetry Configuration

JOSE-JA uses OpenTelemetry Protocol (OTLP) for telemetry export.

```yaml
# OTLP configuration (traces, metrics, logs)
otlp-enabled: true
otlp-endpoint: "opentelemetry-collector:4317"
otlp-service: "jose-ja"
otlp-hostname: "jose-server-1"
```

**Note**: Prometheus scraping is NOT supported. Use OTLP for all telemetry.

### CORS Configuration

```yaml
# CORS - HTTPS origins only
cors-origins:
  - "https://localhost:8092"
  - "https://127.0.0.1:8092"
```

### Session Configuration

```yaml
# Browser session configuration
browser-realms:
  - "browser-realm-1"
browser-session-lifetime: "24h"
browser-session-key-file: "/run/secrets/browser_session_key"

# Service session configuration
service-realms:
  - "service-realm-1"
service-session-lifetime: "1h"
service-session-key-file: "/run/secrets/service_session_key"
```

### JOSE-JA Specific Settings

```yaml
# Elastic key defaults
max-materials: 10                    # Max material keys per elastic key
audit-enabled: true                  # Enable audit logging
audit-sampling-rate: 100             # Audit 100% of operations (0-100)
```

## Docker Secrets

All sensitive configuration MUST use Docker secrets:

### Creating Secrets

```bash
# Create secrets directory
mkdir -p secrets

# Generate TLS certificate (development)
openssl req -x509 -newkey rsa:4096 -keyout secrets/tls.key -out secrets/tls.crt \
  -days 365 -nodes -subj "/CN=localhost"

# Database URL
echo "postgres://jose:secret@postgres:5432/jose?sslmode=disable" > secrets/database_url

# Session keys (random 32 bytes, base64 encoded)
openssl rand -base64 32 > secrets/browser_session_key
openssl rand -base64 32 > secrets/service_session_key

# Unseal key (random 32 bytes, base64 encoded)
openssl rand -base64 32 > secrets/unseal_key
```

### Docker Compose Secrets

```yaml
services:
  jose-server:
    secrets:
      - tls_cert
      - tls_key
      - database_url
      - browser_session_key
      - service_session_key
      - unseal_key
    command:
      - start
      - --config=/etc/jose/jose.yml
      - --tls-cert-file=file:///run/secrets/tls_cert
      - --tls-key-file=file:///run/secrets/tls_key
      - --database-dsn=file:///run/secrets/database_url

secrets:
  tls_cert:
    file: ./secrets/tls.crt
  tls_key:
    file: ./secrets/tls.key
  database_url:
    file: ./secrets/database_url
  browser_session_key:
    file: ./secrets/browser_session_key
  service_session_key:
    file: ./secrets/service_session_key
  unseal_key:
    file: ./secrets/unseal_key
```

## Health Endpoints

### Public Server Health

The public server provides health endpoints for external monitoring:

| Endpoint | Purpose | Kubernetes Probe |
|----------|---------|------------------|
| `GET /admin/api/v1/livez` | Liveness check | livenessProbe |
| `GET /admin/api/v1/readyz` | Readiness check | readinessProbe |

### Admin Server Health

Admin health endpoints are on a separate port (default: 9092) bound to localhost only:

```bash
# From within container
wget -q -O - https://127.0.0.1:9092/admin/api/v1/livez

# Docker health check configuration
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9092/admin/api/v1/livez"]
  start_period: 30s
  interval: 10s
  timeout: 5s
  retries: 5
```

### Graceful Shutdown

```bash
# Trigger graceful shutdown
curl -k -X POST https://127.0.0.1:9092/admin/api/v1/shutdown
```

## Networking

### Port Allocation

| Port | Protocol | Purpose |
|------|----------|---------|
| 8092 | HTTPS | Public API (service-to-service, browser) |
| 9092 | HTTPS | Admin API (health checks, shutdown) |

### Network Isolation

```yaml
networks:
  jose-network:
    driver: bridge
    internal: false    # Public API accessible
  admin-network:
    driver: bridge
    internal: true     # Admin API isolated
```

## Resource Limits

```yaml
deploy:
  resources:
    limits:
      memory: 256M
      cpus: '1.0'
    reservations:
      memory: 128M
      cpus: '0.25'
```

## Logging

JOSE-JA outputs structured JSON logs to stdout:

```json
{
  "level": "INFO",
  "timestamp": "2025-01-15T12:00:00Z",
  "message": "Server started",
  "service": "jose-ja",
  "public_port": 8092,
  "admin_port": 9092
}
```

### Log Levels

| Level | Description |
|-------|-------------|
| DEBUG | Detailed debugging information |
| INFO | Normal operational messages |
| WARN | Degraded mode or recoverable errors |
| ERROR | Unrecoverable errors |
| FATAL | Critical errors causing shutdown |

Configure via:

```yaml
log-level: "INFO"
```

## Backup and Recovery

### Database Backup

**PostgreSQL**:

```bash
# Backup
pg_dump -h postgres -U jose -d jose > jose_backup.sql

# Restore
psql -h postgres -U jose -d jose < jose_backup.sql
```

**SQLite**:

```bash
# Backup
sqlite3 jose.db ".backup jose_backup.db"

# Restore
sqlite3 jose.db ".restore jose_backup.db"
```

### Key Material Recovery

Elastic JWKs are encrypted with the barrier service. Ensure unseal key is backed up securely.

## Troubleshooting

### Common Issues

**1. "database is locked" (SQLite)**

Enable WAL mode in SQLite configuration:

```yaml
database-dsn: "file:jose.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=30000"
```

**2. "connection refused" to admin endpoint**

Admin server binds to localhost only. Access from within container:

```bash
docker exec jose-server wget -q -O - https://127.0.0.1:9092/admin/api/v1/livez
```

**3. TLS certificate errors**

Ensure TLS certificates include proper SANs:

```bash
# Check certificate SANs
openssl x509 -in tls.crt -text -noout | grep -A1 "Subject Alternative Name"
```

**4. Health check failures**

Check container logs:

```bash
docker compose logs jose-server
```

## Cross-References

- [API-REFERENCE.md](API-REFERENCE.md) - API documentation
- [Service Template](../service-template/) - Shared infrastructure patterns
- [Database Patterns](../../.github/instructions/03-04.database.instructions.md) - Database configuration
