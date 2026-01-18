# JOSE-JA Deployment Guide

This guide covers deploying the JOSE-JA (JWK Authority) service in various environments.

## Prerequisites

- Docker 24+ and Docker Compose v2+
- Go 1.25.5+ (for building from source)
- PostgreSQL 16+ (for production) or SQLite (for development)

## Quick Start (Development)

### Docker Compose

```bash
# Start JOSE-JA with SQLite (development mode)
docker compose -f deployments/jose/compose.yml up -d

# Verify service is running
curl -k https://127.0.0.1:8080/admin/v1/livez

# View logs
docker compose -f deployments/jose/compose.yml logs -f jose-ja
```

### Binary

```bash
# Build
CGO_ENABLED=0 go build -o jose-server ./cmd/jose-server

# Run in development mode (SQLite in-memory)
./jose-server --dev

# Run with configuration file
./jose-server --config configs/jose/jose-server.yaml
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `JOSE_BIND_PUBLIC_ADDRESS` | Public server bind address | `127.0.0.1` |
| `JOSE_BIND_PUBLIC_PORT` | Public server port | `8080` |
| `JOSE_BIND_PRIVATE_ADDRESS` | Admin server bind address | `127.0.0.1` |
| `JOSE_BIND_PRIVATE_PORT` | Admin server port | `9090` |
| `JOSE_DATABASE_URL` | Database connection string | SQLite in-memory |
| `JOSE_LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |

### Configuration File

```yaml
# configs/jose/jose-server.yaml
bind-public-address: "0.0.0.0"
bind-public-port: 8080
bind-private-address: "127.0.0.1"
bind-private-port: 9090

database-url: "postgres://user:pass@postgres:5432/jose?sslmode=disable"

log-level: "info"
verbose-mode: false
dev-mode: false

# TLS Configuration
tls-public-mode: "auto"  # auto generates self-signed certs
tls-private-mode: "auto"

# OTLP Telemetry
otlp-enabled: true
otlp-endpoint: "opentelemetry-collector:4317"
otlp-service: "jose-ja"

# Rate Limiting
browser-ip-rate-limit: 100
service-ip-rate-limit: 500

# Session Configuration
browser-session-expiration: "24h"
service-session-expiration: "1h"
session-idle-timeout: "30m"
```

## Docker Compose Deployment

### Development (SQLite)

```yaml
# deployments/jose/compose.yml
services:
  jose-ja:
    build:
      context: ../..
      dockerfile: deployments/jose/Dockerfile
    ports:
      - "8080:8080"  # Public API
      # Admin port NOT exposed (127.0.0.1 only)
    environment:
      - JOSE_DEV_MODE=true
      - JOSE_BIND_PUBLIC_ADDRESS=0.0.0.0
    healthcheck:
      test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/v1/livez"]
      interval: 10s
      timeout: 5s
      retries: 5
```

### Production (PostgreSQL)

```yaml
# deployments/jose/compose.prod.yml
services:
  jose-ja:
    image: cryptoutil/jose-ja:latest
    ports:
      - "8080:8080"
    secrets:
      - database_url
      - unseal_key
      - tls_cert
      - tls_key
    environment:
      - JOSE_BIND_PUBLIC_ADDRESS=0.0.0.0
      - JOSE_DATABASE_URL=file:///run/secrets/database_url
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/v1/livez"]
      interval: 10s
      timeout: 5s
      retries: 5

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: jose
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
      POSTGRES_DB: jose
    secrets:
      - postgres_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "jose"]
      interval: 10s
      timeout: 5s
      retries: 5

secrets:
  database_url:
    file: ./secrets/database_url.secret
  unseal_key:
    file: ./secrets/unseal_key.secret
  tls_cert:
    file: ./secrets/tls_cert.pem
  tls_key:
    file: ./secrets/tls_key.pem
  postgres_password:
    file: ./secrets/postgres_password.secret

volumes:
  postgres_data:
```

## Kubernetes Deployment

### Deployment

```yaml
# deployments/jose/kubernetes/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jose-ja
  labels:
    app: jose-ja
spec:
  replicas: 3
  selector:
    matchLabels:
      app: jose-ja
  template:
    metadata:
      labels:
        app: jose-ja
    spec:
      containers:
        - name: jose-ja
          image: cryptoutil/jose-ja:latest
          ports:
            - name: public
              containerPort: 8080
            - name: admin
              containerPort: 9090
          env:
            - name: JOSE_BIND_PUBLIC_ADDRESS
              value: "0.0.0.0"
            - name: JOSE_DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: jose-secrets
                  key: database-url
          livenessProbe:
            httpGet:
              path: /admin/v1/livez
              port: admin
              scheme: HTTPS
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /admin/v1/readyz
              port: admin
              scheme: HTTPS
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            requests:
              memory: "256Mi"
              cpu: "200m"
            limits:
              memory: "512Mi"
              cpu: "500m"
```

### Service

```yaml
# deployments/jose/kubernetes/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: jose-ja
spec:
  selector:
    app: jose-ja
  ports:
    - name: public
      port: 8080
      targetPort: 8080
    - name: admin
      port: 9090
      targetPort: 9090
  type: ClusterIP
```

### Ingress

```yaml
# deployments/jose/kubernetes/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: jose-ja
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
spec:
  tls:
    - hosts:
        - jose.example.com
      secretName: jose-tls
  rules:
    - host: jose.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: jose-ja
                port:
                  number: 8080
```

## Multi-Tenant Setup

### Initial Setup

1. **Deploy JOSE-JA** using one of the methods above.

2. **Register First User with Tenant**:
   ```bash
   curl -X POST https://jose.example.com/service/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{
       "username": "admin",
       "password": "securepassword",
       "create_tenant": true
     }'
   ```

3. **Store credentials securely** - the response contains `tenant_id`, `realm_id`, and `session_token`.

4. **Invite additional users**:
   ```bash
   # User requests to join tenant
   curl -X POST https://jose.example.com/service/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{
       "username": "user2",
       "password": "password2",
       "join_tenant_id": "<tenant_id>"
     }'
   
   # Admin approves join request
   curl -X POST https://jose.example.com/service/api/v1/admin/join-requests/<request_id>/approve \
     -H "Authorization: Bearer <admin_session_token>"
   ```

### Cross-Tenant Access (Future)

Cross-tenant JWKS access requires explicit configuration:

```sql
-- Enable cross-tenant access for specific JWK
UPDATE elastic_jwks 
SET allow_cross_tenant = true 
WHERE kid = '019bd10b-5d65-7bdd-a717-d1f057c85b8a';
```

## Security Best Practices

### TLS Configuration

1. **Use proper certificates** in production (not self-signed).
2. **Configure TLS 1.3 minimum**:
   ```yaml
   tls-min-version: "1.3"
   ```
3. **Enable mTLS** for service-to-service:
   ```yaml
   tls-client-auth: "require"
   tls-client-ca-file: "/etc/jose/ca.crt"
   ```

### Network Security

1. **Admin API** - Never expose externally (127.0.0.1 only).
2. **Rate limiting** - Configure appropriate limits per environment.
3. **IP allowlisting** - Restrict access to known IPs:
   ```yaml
   allowed-ips:
     - "10.0.0.0/8"
     - "172.16.0.0/12"
   ```

### Secret Management

1. **Use Docker/Kubernetes secrets** - Never environment variables for sensitive data.
2. **Rotate unseal keys** periodically.
3. **Audit key access** - Enable audit logging for all operations.

## Monitoring

### Prometheus Metrics

JOSE-JA exposes metrics at `/admin/v1/metrics`:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'jose-ja'
    scheme: https
    tls_config:
      insecure_skip_verify: true  # Use proper certs in production
    static_configs:
      - targets: ['jose-ja:9090']
```

### Key Metrics

| Metric | Description |
|--------|-------------|
| `jose_jwk_operations_total` | Total JWK operations by type |
| `jose_sign_duration_seconds` | Sign operation latency |
| `jose_verify_duration_seconds` | Verify operation latency |
| `jose_encrypt_duration_seconds` | Encrypt operation latency |
| `jose_decrypt_duration_seconds` | Decrypt operation latency |
| `jose_rate_limit_exceeded_total` | Rate limit violations |

### OpenTelemetry

Configure OTLP export for traces and logs:

```yaml
otlp-enabled: true
otlp-endpoint: "opentelemetry-collector:4317"
otlp-service: "jose-ja"
otlp-environment: "production"
```

## Troubleshooting

### Common Issues

**Service won't start**:
- Check database connectivity: `pg_isready -h postgres -U jose`
- Verify secrets are mounted: `ls -la /run/secrets/`
- Check logs: `docker compose logs jose-ja`

**Authentication failures**:
- Verify session token is valid (not expired)
- Check tenant/realm IDs match
- Ensure user has appropriate permissions

**Rate limit exceeded**:
- Check current limits in configuration
- Review `jose_rate_limit_exceeded_total` metric
- Increase limits if legitimate traffic

**High latency**:
- Check database connection pool settings
- Review `jose_*_duration_seconds` metrics
- Consider scaling replicas

### Health Check Endpoints

```bash
# Liveness (is the process alive?)
curl -k https://127.0.0.1:9090/admin/v1/livez

# Readiness (is the service ready for traffic?)
curl -k https://127.0.0.1:9090/admin/v1/readyz
```

### Debug Logging

Enable debug logging temporarily:

```yaml
log-level: "debug"
verbose-mode: true
```

Or via environment:
```bash
JOSE_LOG_LEVEL=debug JOSE_VERBOSE_MODE=true ./jose-server
```
