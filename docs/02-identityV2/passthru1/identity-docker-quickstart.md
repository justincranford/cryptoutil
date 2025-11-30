# Identity Services Docker Compose - Quick Start Guide

**Audience**: Developers, QA Engineers
**Purpose**: Practical guide for running, debugging, and scaling identity services locally
**Prerequisites**: Docker Desktop installed, identity-demo.yml configured with secrets

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Profiles](#profiles)
3. [Scaling Scenarios](#scaling-scenarios)
4. [Common Workflows](#common-workflows)
5. [Troubleshooting](#troubleshooting)
6. [Observability](#observability)

---

## Quick Start

### Using identity-orchestrator CLI (Recommended)

```bash
# Start demo profile (1 instance of each service)
go run ./cmd/identity-orchestrator -operation start -profile demo

# Check health status
go run ./cmd/identity-orchestrator -operation health -profile demo

# View logs (all services)
go run ./cmd/identity-orchestrator -operation logs -profile demo -tail 50

# View logs (specific service)
go run ./cmd/identity-orchestrator -operation logs -profile demo -service identity-authz -follow

# Stop services (keep volumes)
go run ./cmd/identity-orchestrator -operation stop -profile demo

# Stop services (remove volumes)
go run ./cmd/identity-orchestrator -operation stop -profile demo -remove-volumes
```

### Using docker compose Directly

```bash
# Start demo profile
docker compose -f deployments/compose/identity-demo.yml --profile demo up -d

# Check service status
docker compose -f deployments/compose/identity-demo.yml --profile demo ps

# View logs
docker compose -f deployments/compose/identity-demo.yml --profile demo logs -f --tail 50

# Stop services
docker compose -f deployments/compose/identity-demo.yml --profile demo down -v
```

---

## Profiles

### Demo Profile (1x1x1x1)

**Purpose**: Quick demonstration, functional testing, API exploration
**Resources**: Minimal (1 instance each: AuthZ, IdP, RS, SPA)

```bash
go run ./cmd/identity-orchestrator -operation start -profile demo
```

**Exposed Ports**:

- AuthZ: `http://localhost:8080` (OAuth 2.1 endpoints)
- IdP: `http://localhost:8100` (OIDC endpoints)
- RS: `http://localhost:8200` (Resource server API)
- SPA: `http://localhost:8300` (SPA relying party)
- PostgreSQL: `localhost:5433` (database)

---

### Development Profile (2x2x2x2)

**Purpose**: High availability testing, load balancer integration, failover scenarios
**Resources**: Medium (2 instances each)

```bash
go run ./cmd/identity-orchestrator -operation start -profile development
```

**Exposed Ports**:

- AuthZ: `http://localhost:8080-8081` (2 instances)
- IdP: `http://localhost:8100-8101` (2 instances)
- RS: `http://localhost:8200-8201` (2 instances)
- SPA: `http://localhost:8300-8301` (2 instances)

**Use Case**: Test session failover when one AuthZ instance goes down

---

### CI Profile (1x1x1x1)

**Purpose**: CI/CD pipelines, automated testing, minimal resource usage
**Resources**: Minimal (same as demo, optimized for CI)

```bash
go run ./cmd/identity-orchestrator -operation start -profile ci
```

**Differences from Demo**:

- Same topology as demo
- Optimized for fast startup in CI environments
- Health check retries configured for CI timeouts

---

### Production Profile (3x3x3x3)

**Purpose**: Production-like testing, stress testing, scalability validation
**Resources**: High (3 instances each)

```bash
go run ./cmd/identity-orchestrator -operation start -profile production
```

**Exposed Ports**:

- AuthZ: `http://localhost:8080-8082` (3 instances)
- IdP: `http://localhost:8100-8102` (3 instances)
- RS: `http://localhost:8200-8202` (3 instances)
- SPA: `http://localhost:8300-8302` (3 instances)

**Use Case**: Validate production deployment with load balancer and service mesh

---

## Scaling Scenarios

### Custom Scaling

**Override default replica counts for specific services:**

```bash
# 2x AuthZ, 1x IdP, 1x RS, 1x SPA
go run ./cmd/identity-orchestrator -operation start -profile demo -scaling "identity-authz=2"

# 3x AuthZ, 2x IdP, 1x RS, 1x SPA
go run ./cmd/identity-orchestrator -operation start -profile demo -scaling "identity-authz=3,identity-idp=2"

# High availability: 3x all services
go run ./cmd/identity-orchestrator -operation start -profile demo -scaling "identity-authz=3,identity-idp=3,identity-rs=3,identity-spa-rp=3"
```

### Using docker compose --scale

```bash
# Scale AuthZ to 2 instances
docker compose -f deployments/compose/identity-demo.yml --profile demo up -d --scale identity-authz=2

# Scale multiple services
docker compose -f deployments/compose/identity-demo.yml --profile demo up -d --scale identity-authz=3 --scale identity-idp=2
```

---

## Common Workflows

### 1. Single-Service Testing

**Test only AuthZ service without dependencies:**

```bash
# Start only PostgreSQL and AuthZ
docker compose -f deployments/compose/identity-demo.yml --profile demo up -d identity-postgres identity-authz

# Verify AuthZ health
curl -k https://localhost:9080/livez

# Test OAuth endpoints
curl -X POST https://localhost:8080/oauth/token \
  -d "grant_type=client_credentials" \
  -d "client_id=test-client" \
  -d "client_secret=test-secret"

# Stop when done
docker compose -f deployments/compose/identity-demo.yml --profile demo down -v
```

---

### 2. Debugging Service Failures

**Service fails to start? Check logs and health status:**

```bash
# View logs for specific service
go run ./cmd/identity-orchestrator -operation logs -profile demo -service identity-authz -tail 100

# Check health status
docker compose -f deployments/compose/identity-demo.yml --profile demo ps

# Inspect service details
docker inspect <container_id>

# Check database connectivity
docker compose -f deployments/compose/identity-demo.yml --profile demo exec identity-authz \
  wget --spider postgres://identity_user:identity_pass@identity-postgres:5432/identity_db
```

---

### 3. Testing Scaling and Failover

**Simulate instance failure:**

```bash
# Start development profile (2x2x2x2)
go run ./cmd/identity-orchestrator -operation start -profile development

# Kill one AuthZ instance
docker kill <authz_container_id>

# Verify other AuthZ instance handles traffic
curl -k https://localhost:8080/oauth/token (should still work)

# Restart failed instance
docker compose -f deployments/compose/identity-demo.yml --profile development up -d identity-authz
```

---

### 4. Load Testing Preparation

**Start production profile for load testing:**

```bash
# Start production profile (3x3x3x3)
go run ./cmd/identity-orchestrator -operation start -profile production

# Wait for all services to be healthy
go run ./cmd/identity-orchestrator -operation health -profile production

# Run load tests (example with Gatling)
cd test/load
./mvnw gatling:test -Dgatling.simulationClass=IdentityLoadTest

# Monitor resource usage
docker stats

# Check service logs for errors
go run ./cmd/identity-orchestrator -operation logs -profile production -tail 200
```

---

## Troubleshooting

### Health Checks Failing

**Symptom**: Services show as "unhealthy" in `docker compose ps`

**Common Causes**:

1. **Database not ready**: PostgreSQL initialization takes 5-10 seconds
2. **Incorrect health check URL**: Verify IPv4 loopback (127.0.0.1)
3. **Port conflicts**: Check if ports 8080-8309 are already in use

**Solutions**:

```bash
# Check PostgreSQL logs
docker compose -f deployments/compose/identity-demo.yml --profile demo logs identity-postgres

# Verify health check command
docker compose -f deployments/compose/identity-demo.yml --profile demo exec identity-authz \
  wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9090/livez

# Check port availability
netstat -an | findstr "8080"  # Windows
lsof -i :8080                 # macOS/Linux

# Increase health check retries
# Edit identity-demo.yml: healthcheck.retries = 10
```

---

### Service Fails to Connect to Database

**Symptom**: Logs show "connection refused" or "database does not exist"

**Common Causes**:

1. **Docker secrets not mounted**: Verify `/run/secrets/postgres_*` files exist
2. **Database not initialized**: PostgreSQL container not fully started
3. **Network configuration**: Services not on same Docker network

**Solutions**:

```bash
# Verify secrets are mounted
docker compose -f deployments/compose/identity-demo.yml --profile demo exec identity-authz ls -la /run/secrets/

# Check database initialization
docker compose -f deployments/compose/identity-demo.yml --profile demo logs identity-postgres | grep "database system is ready"

# Test database connectivity
docker compose -f deployments/compose/identity-demo.yml --profile demo exec identity-authz \
  wget --spider postgres://$(cat /run/secrets/postgres_user):$(cat /run/secrets/postgres_password)@identity-postgres:5432/$(cat /run/secrets/postgres_db)

# Restart services in correct order
docker compose -f deployments/compose/identity-demo.yml --profile demo down -v
docker compose -f deployments/compose/identity-demo.yml --profile demo up -d
```

---

### Port Conflicts

**Symptom**: `docker compose up` fails with "port already allocated"

**Common Causes**:

1. **Multiple Compose stacks running**: cryptoutil + identity services
2. **Previous containers not stopped**: zombie containers holding ports
3. **Host services using same ports**: local development server

**Solutions**:

```bash
# Find which process is using the port
netstat -ano | findstr "8080"  # Windows
lsof -i :8080                  # macOS/Linux

# Kill zombie containers
docker ps -a
docker rm -f <container_id>

# Stop cryptoutil services first
docker compose -f deployments/compose/compose.yml down -v

# Then start identity services
go run ./cmd/identity-orchestrator -operation start -profile demo
```

---

### Network Connectivity Issues

**Symptom**: Services cannot communicate with each other

**Common Causes**:

1. **Wrong network configuration**: Services not on identity-network
2. **Firewall blocking**: Docker Desktop firewall rules
3. **DNS resolution failing**: Service names not resolving

**Solutions**:

```bash
# Check network connectivity
docker compose -f deployments/compose/identity-demo.yml --profile demo exec identity-idp ping identity-authz

# Verify services on same network
docker network inspect identity-network

# Check DNS resolution
docker compose -f deployments/compose/identity-demo.yml --profile demo exec identity-idp nslookup identity-authz

# Recreate network
docker compose -f deployments/compose/identity-demo.yml --profile demo down -v
docker compose -f deployments/compose/identity-demo.yml --profile demo up -d
```

---

## Observability

### OTEL Collector Integration

**Identity services send telemetry to OTEL collector:**

```yaml
# identity-demo.yml configuration
environment:
  - OTLP_ENDPOINT=http://opentelemetry-collector-contrib:4317
```

**Verify telemetry is being sent:**

```bash
# Check OTEL collector logs
docker logs opentelemetry-collector-contrib

# Verify metrics endpoint
curl http://localhost:8888/metrics
```

---

### Grafana Dashboard

**Access Grafana to visualize identity service metrics:**

```bash
# Start Grafana stack (if not already running)
docker compose -f deployments/compose/compose.yml --profile observability up -d grafana-otel-lgtm

# Access Grafana UI
open http://localhost:3000  # Username: admin, Password: admin

# Add identity services dashboard
# Dashboards → New → Import → Upload identity-services.json
```

**Key Metrics to Monitor**:

- **Request Rate**: OAuth token requests/sec, OIDC authentication requests/sec
- **Error Rate**: Failed logins, token validation errors
- **Latency**: p50/p95/p99 response times
- **Resource Usage**: CPU, memory, database connections

---

### Prometheus Scraping

**OTEL collector exposes Prometheus metrics:**

```bash
# Collector self-metrics
curl http://localhost:8888/metrics

# Received metrics (re-exported from identity services)
curl http://localhost:8889/metrics
```

**Configure Prometheus to scrape OTEL collector:**

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'otel-collector'
    static_configs:
      - targets: ['localhost:8888']
  - job_name: 'identity-services'
    static_configs:
      - targets: ['localhost:8889']
```

---

## Advanced Topics

### Load Balancer Configuration

**Use nginx or HAProxy for load balancing multiple AuthZ instances:**

```nginx
# nginx.conf
upstream identity_authz {
    server localhost:8080;
    server localhost:8081;
    server localhost:8082;
}

server {
    listen 80;
    location / {
        proxy_pass http://identity_authz;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

### Service Mesh Integration

**For production deployments, integrate with service mesh (Istio, Linkerd):**

```yaml
# identity-demo.yml with Istio sidecar injection
services:
  identity-authz:
    labels:
      sidecar.istio.io/inject: "true"
```

---

## Quick Reference

| Operation | Command |
|-----------|---------|
| Start demo | `go run ./cmd/identity-orchestrator -operation start -profile demo` |
| Stop all | `go run ./cmd/identity-orchestrator -operation stop -profile demo -remove-volumes` |
| Health check | `go run ./cmd/identity-orchestrator -operation health -profile demo` |
| View logs | `go run ./cmd/identity-orchestrator -operation logs -profile demo -tail 50` |
| Scale AuthZ to 3 | `go run ./cmd/identity-orchestrator -operation start -profile demo -scaling "identity-authz=3"` |
| Follow logs | `go run ./cmd/identity-orchestrator -operation logs -profile demo -service identity-authz -follow` |

---

**Questions?** See `docs/02-identityV2/task-18-orchestration-suite.md` for detailed architecture documentation.
