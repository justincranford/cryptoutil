# Deployment Guide

## Overview

cryptoutil supports multiple deployment scenarios from local development to production Kubernetes clusters. This guide covers all deployment options with detailed configuration and operational procedures.

## Deployment Options

### 1. Docker Compose (Recommended for Production)

The easiest way to deploy cryptoutil with PostgreSQL in a production-ready configuration.

#### Quick Start

```bash
cd deployments/compose
docker compose up -d
```

#### Services Deployed

- **PostgreSQL Database**: Persistent data storage with secrets management
- **cryptoutil Server**: Production-configured cryptographic service
- **Health Monitoring**: Automatic health checks and restart policies

#### Architecture

```
┌─────────────────────┐    ┌─────────────────────┐
│   cryptoutil        │    │    PostgreSQL       │
│   Port 8080 (HTTPS) │◄──►│   Port 5432         │
│   Port 9090 (HTTP)  │    │   Persistent Volume │
└─────────────────────┘    └─────────────────────┘
         │
         ▼
┌─────────────────────┐
│   Docker Secrets    │
│   • Database URL    │
│   • Unseal Keys     │
│   • Configuration   │
└─────────────────────┘
```

### 2. Kubernetes Deployment

Production-ready Kubernetes deployment with auto-scaling and monitoring.

#### Prerequisites

- Kubernetes cluster (1.21+)
- kubectl configured
- Persistent volume support
- Ingress controller (optional)

#### Basic Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cryptoutil
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cryptoutil
  template:
    metadata:
      labels:
        app: cryptoutil
    spec:
      containers:
      - name: cryptoutil
        image: cryptoutil:latest
        ports:
        - containerPort: 8080
          name: public
        - containerPort: 9090
          name: private
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: cryptoutil-secrets
              key: database-url
        livenessProbe:
          httpGet:
            path: /livez
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 9090
          initialDelaySeconds: 5
          periodSeconds: 5
```

#### Service Configuration

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: cryptoutil-service
spec:
  selector:
    app: cryptoutil
  ports:
  - name: public
    port: 8080
    targetPort: 8080
  - name: private
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

### 3. Standalone Binary

Direct deployment for development or simple production scenarios.

```bash
# Build
go build -o cryptoutil cmd/cryptoutil/main.go

# Run with configuration
./cryptoutil --config=production.yaml
```

## Configuration Management

### Docker Compose Configuration

#### Primary Configuration Files

```
deployments/compose/
├── compose.yml                    # Main Docker Compose configuration
├── cryptoutil/
│   ├── postgresql.yml            # PostgreSQL database configuration
│   ├── sqlite.yml                # SQLite development configuration
│   ├── cryptoutil_database_url.secret
│   ├── cryptoutil_unseal_1of5.secret
│   ├── cryptoutil_unseal_2of5.secret
│   ├── cryptoutil_unseal_3of5.secret
│   ├── cryptoutil_unseal_4of5.secret
│   └── cryptoutil_unseal_5of5.secret
└── postgres/
    ├── postgres_database.secret
    ├── postgres_password.secret
    └── postgres_username.secret
```

#### Docker Compose File Structure

```yaml
# deployments/compose/compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    restart: unless-stopped
    environment:
      POSTGRES_DB_FILE: /run/secrets/postgres_database
      POSTGRES_USER_FILE: /run/secrets/postgres_username
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
    secrets:
      - postgres_database
      - postgres_username
      - postgres_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  cryptoutil:
    build:
      context: ../..
      dockerfile: deployments/Dockerfile
    restart: unless-stopped
    ports:
      - "8080:8080"
      - "9090:9090"
    secrets:
      - cryptoutil_database_url
      - cryptoutil_unseal_1of5
      - cryptoutil_unseal_2of5
      - cryptoutil_unseal_3of5
      - cryptoutil_unseal_4of5
      - cryptoutil_unseal_5of5
    volumes:
      - ./cryptoutil/postgresql.yml:/app/config.yaml:ro
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9090/livez"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  postgres_data:

secrets:
  postgres_database:
    file: ./postgres/postgres_database.secret
  postgres_username:
    file: ./postgres/postgres_username.secret
  postgres_password:
    file: ./postgres/postgres_password.secret
  cryptoutil_database_url:
    file: ./cryptoutil/cryptoutil_database_url.secret
  cryptoutil_unseal_1of5:
    file: ./cryptoutil/cryptoutil_unseal_1of5.secret
  cryptoutil_unseal_2of5:
    file: ./cryptoutil/cryptoutil_unseal_2of5.secret
  cryptoutil_unseal_3of5:
    file: ./cryptoutil/cryptoutil_unseal_3of5.secret
  cryptoutil_unseal_4of5:
    file: ./cryptoutil/cryptoutil_unseal_4of5.secret
  cryptoutil_unseal_5of5:
    file: ./cryptoutil/cryptoutil_unseal_5of5.secret
```

### Secret Management

#### Docker Secrets

Secrets are managed through Docker's built-in secrets system:

```bash
# Create secret files
echo "cryptoutil" > postgres/postgres_database.secret
echo "cryptoutil_user" > postgres/postgres_username.secret
echo "$(openssl rand -base64 32)" > postgres/postgres_password.secret

# Database URL secret
echo "postgres://cryptoutil_user:$(cat postgres/postgres_password.secret)@postgres:5432/cryptoutil?sslmode=disable" > cryptoutil/cryptoutil_database_url.secret

# Unseal key secrets (M-of-N sharing)
for i in {1..5}; do
  openssl rand -base64 64 > cryptoutil/cryptoutil_unseal_${i}of5.secret
done
```

#### Kubernetes Secrets

```bash
# Create Kubernetes secrets
kubectl create secret generic cryptoutil-secrets \
  --from-literal=database-url="postgres://user:pass@postgres:5432/cryptoutil" \
  --from-file=unseal-key-1=./unseal_1of5.key \
  --from-file=unseal-key-2=./unseal_2of5.key \
  --from-file=unseal-key-3=./unseal_3of5.key
```

## Container Configuration

### Dockerfile

```dockerfile
# deployments/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go generate ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cryptoutil cmd/cryptoutil/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

COPY --from=builder /app/cryptoutil .
COPY --from=builder /app/configs/ ./configs/

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:9090/livez || exit 1

CMD ["./cryptoutil", "--config=/app/config.yaml"]
```

### Multi-Stage Build Benefits

- **Smaller final image**: Only runtime dependencies included
- **Security**: No build tools in production image
- **Performance**: Faster deployment and startup
- **Compliance**: Clean production environment

## Network Configuration

### Port Mapping

| Port | Protocol | Interface | Purpose |
|------|----------|-----------|---------|
| 8080 | HTTPS | Public | API endpoints (browser/service contexts) |
| 9090 | HTTP | Private | Management and health checks |
| 5432 | TCP | Internal | PostgreSQL database |

### Firewall Rules

```bash
# Allow public API access
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT

# Restrict management interface to localhost
iptables -A INPUT -p tcp --dport 9090 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 9090 -j REJECT

# Database access (internal only)
iptables -A INPUT -p tcp --dport 5432 -s 172.17.0.0/16 -j ACCEPT
iptables -A INPUT -p tcp --dport 5432 -j REJECT
```

### Load Balancer Configuration

#### Nginx Reverse Proxy

```nginx
# /etc/nginx/sites-available/cryptoutil
upstream cryptoutil_backend {
    server 127.0.0.1:8080;
    # Add multiple instances for load balancing
    # server 127.0.0.1:8081;
    # server 127.0.0.1:8082;
}

server {
    listen 443 ssl http2;
    server_name cryptoutil.example.com;

    ssl_certificate /path/to/certificate.pem;
    ssl_certificate_key /path/to/private-key.pem;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;

    location / {
        proxy_pass https://cryptoutil_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Rate limiting
        limit_req zone=api burst=20 nodelay;
    }

    # Health check endpoint (restricted access)
    location /health {
        proxy_pass http://127.0.0.1:9090/livez;
        allow 10.0.0.0/8;
        deny all;
    }
}

# Rate limiting zone
http {
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
}
```

## Monitoring and Health Checks

### Health Check Endpoints

#### Liveness Probe
```bash
curl -f http://localhost:9090/livez
# Returns 200 if application is alive
```

#### Readiness Probe
```bash
curl -f http://localhost:9090/readyz
# Returns 200 if application is ready to serve requests
```

### Monitoring Integration

#### Prometheus Metrics

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'cryptoutil'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

#### Grafana Dashboard

Key metrics to monitor:

- **Request Rate**: Requests per second by endpoint
- **Error Rate**: HTTP error rate (4xx, 5xx)
- **Response Time**: Request latency percentiles
- **Key Operations**: Key generation, encryption, decryption rates
- **Database Performance**: Connection pool, query latency
- **System Resources**: CPU, memory, disk usage

### Log Management

#### Structured Logging

```json
{
  "timestamp": "2025-09-12T10:30:00Z",
  "level": "INFO",
  "message": "HTTP request completed",
  "request_id": "req_123456",
  "method": "POST",
  "path": "/service/api/v1/elastickey",
  "status": 200,
  "duration": "45ms",
  "client_ip": "192.168.1.100"
}
```

#### Log Aggregation

**ELK Stack Integration**:
```yaml
# logstash.conf
input {
  file {
    path => "/var/log/cryptoutil/*.log"
    codec => "json"
  }
}

filter {
  if [level] == "ERROR" {
    mutate {
      add_tag => ["alert"]
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "cryptoutil-%{+YYYY.MM.dd}"
  }
}
```

## Backup and Recovery

### Database Backup

#### Automated PostgreSQL Backup

```bash
#!/bin/bash
# backup-cryptoutil.sh

BACKUP_DIR="/backups/cryptoutil"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DB_NAME="cryptoutil"

# Create backup directory
mkdir -p $BACKUP_DIR

# Perform backup
pg_dump -h postgres -U cryptoutil_user -d $DB_NAME | \
  gzip > $BACKUP_DIR/cryptoutil_$TIMESTAMP.sql.gz

# Encrypt backup
gpg --cipher-algo AES256 --compress-algo 1 --s2k-cipher-algo AES256 \
    --s2k-digest-algo SHA512 --s2k-mode 3 --s2k-count 65011712 \
    --force-mdc --quiet --no-greeting --batch --yes \
    --passphrase-file /secrets/backup_passphrase \
    --output $BACKUP_DIR/cryptoutil_$TIMESTAMP.sql.gz.gpg \
    --symmetric $BACKUP_DIR/cryptoutil_$TIMESTAMP.sql.gz

# Clean up unencrypted backup
rm $BACKUP_DIR/cryptoutil_$TIMESTAMP.sql.gz

# Retain only last 30 days
find $BACKUP_DIR -name "*.gpg" -mtime +30 -delete
```

### Disaster Recovery

#### Recovery Procedure

1. **Restore Database**:
   ```bash
   # Decrypt backup
   gpg --quiet --batch --yes \
       --passphrase-file /secrets/backup_passphrase \
       --output cryptoutil_backup.sql.gz \
       --decrypt cryptoutil_20250912_103000.sql.gz.gpg

   # Restore database
   gunzip -c cryptoutil_backup.sql.gz | \
     psql -h postgres -U cryptoutil_user -d cryptoutil
   ```

2. **Restore Unseal Keys**:
   ```bash
   # Copy unseal key secrets to deployment directory
   cp /backup/unseal_keys/* deployments/compose/cryptoutil/
   ```

3. **Restart Services**:
   ```bash
   cd deployments/compose
   docker compose down
   docker compose up -d
   ```

## Performance Tuning

### Database Optimization

#### PostgreSQL Configuration

```postgresql
-- postgresql.conf optimizations
shared_buffers = '256MB'
effective_cache_size = '1GB'
maintenance_work_mem = '64MB'
checkpoint_completion_target = 0.9
wal_buffers = '16MB'
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
```

#### Connection Pooling

```yaml
# Configuration for database connections
database_url: "postgres://user:pass@host:5432/db?pool_max_conns=25&pool_min_conns=5"
```

### Application Performance

#### Key Generation Pool Tuning

```yaml
# Optimize key generation pools
key_generation_pools:
  rsa_2048: 100    # Pre-generated RSA 2048-bit keys
  rsa_4096: 50     # Pre-generated RSA 4096-bit keys
  ecdsa_p256: 200  # Pre-generated ECDSA P-256 keys
  aes_256: 500     # Pre-generated AES 256-bit keys
```

#### Memory Management

```bash
# Set Go runtime parameters
export GOGC=100           # Garbage collection target percentage
export GOMEMLIMIT=2GiB    # Memory limit for Go runtime
```

## Security Hardening

### Production Security Checklist

- [ ] **TLS Configuration**: Use production certificates, not self-signed
- [ ] **IP Allowlisting**: Configure strict IP/CIDR allowlists
- [ ] **Rate Limiting**: Set conservative rate limits
- [ ] **Database Security**: Use encrypted connections (sslmode=require)
- [ ] **Secret Management**: Rotate unseal keys and database passwords
- [ ] **Container Security**: Run as non-root user, read-only filesystem
- [ ] **Network Security**: Isolate management interface
- [ ] **Monitoring**: Set up comprehensive alerting
- [ ] **Backup**: Implement encrypted backup strategy
- [ ] **Updates**: Establish security update procedures

### Container Hardening

```dockerfile
# Production Dockerfile with security hardening
FROM alpine:latest

# Create non-root user
RUN addgroup -g 1001 -S cryptoutil && \
    adduser -u 1001 -D -S -G cryptoutil cryptoutil

# Install minimal dependencies
RUN apk --no-cache add ca-certificates curl

# Copy application
COPY --from=builder /app/cryptoutil /usr/local/bin/
RUN chmod +x /usr/local/bin/cryptoutil

# Set up read-only filesystem
USER cryptoutil
WORKDIR /tmp

# Run with read-only root filesystem
# docker run --read-only --tmpfs /tmp cryptoutil
```

## Troubleshooting

### Common Issues

#### 1. Service Won't Start

**Symptoms**: Container exits immediately
**Diagnosis**:
```bash
docker compose logs cryptoutil
```
**Common Causes**:
- Database connection failure
- Invalid configuration
- Missing unseal keys
- Port conflicts

#### 2. Health Checks Failing

**Symptoms**: Health check endpoints return errors
**Diagnosis**:
```bash
curl -v http://localhost:9090/livez
curl -v http://localhost:9090/readyz
```
**Common Causes**:
- Database connectivity issues
- Unseal key problems
- Resource exhaustion

#### 3. CSRF Token Issues

**Symptoms**: Browser API requests fail with 403
**Diagnosis**: Check browser developer tools for CSRF token
**Resolution**: Ensure cookies are enabled and same-origin policy

### Performance Issues

#### High Memory Usage

```bash
# Monitor memory usage
docker stats cryptoutil

# Check Go memory stats
curl http://localhost:9090/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

#### Database Performance

```sql
-- Check slow queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- Check connection usage
SELECT count(*) as total_connections,
       count(*) FILTER (WHERE state = 'active') as active_connections,
       count(*) FILTER (WHERE state = 'idle') as idle_connections
FROM pg_stat_activity;
```

This deployment guide provides comprehensive coverage for all deployment scenarios while maintaining security and operational best practices.
