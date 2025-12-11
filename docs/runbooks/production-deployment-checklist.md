# Production Deployment Checklist

**Last Updated**: 2025-11-24
**Version**: 1.0.0
**Scope**: Identity V2 Services and KMS Server

---

## Overview

This checklist covers production deployment procedures for cryptoutil services, including OAuth 2.1 Authorization Server, OIDC Identity Provider, Resource Server, and KMS server.

**Deployment Strategy**: Blue-Green deployment with health check validation
**Rollback Time Objective (RTO)**: <5 minutes
**Recovery Point Objective (RPO)**: Zero data loss (database backed up before deployment)

---

## Pre-Deployment Phase

### Prerequisites

- [ ] **Environment Validation**
  - [ ] Docker/Kubernetes platform operational and accessible
  - [ ] Required ports available (8080-8082 for Identity, 9090 for admin endpoints)
  - [ ] Network connectivity to databases confirmed
  - [ ] TLS certificates valid and accessible
  - [ ] Secret management system configured (Docker Secrets or Kubernetes Secrets)

- [ ] **Configuration Review**
  - [ ] All configuration files reviewed and approved
  - [ ] Database connection strings verified (PostgreSQL DSN format)
  - [ ] OTLP telemetry endpoints configured
  - [ ] CORS origins properly set for each service
  - [ ] Rate limiting thresholds configured
  - [ ] Log levels appropriate for production (INFO or WARN)

- [ ] **Security Validation**
  - [ ] All secrets stored in secure secret management system (NOT environment variables)
  - [ ] Unseal secrets configured for KMS (3-of-5 threshold)
  - [ ] PostgreSQL credentials stored in Docker/K8s secrets
  - [ ] TLS certificates have valid signatures and expiration dates
  - [ ] Security scans completed (gosec, DAST if act available)
  - [ ] No CRITICAL/HIGH vulnerabilities unaddressed

- [ ] **Testing Validation**
  - [ ] All unit tests passing (`go test ./...`)
  - [ ] Integration tests passing (`go test ./internal/identity/integration/...`)
  - [ ] E2E tests validated in staging environment
  - [ ] Test coverage ≥95% for production code

- [ ] **Backup Strategy**
  - [ ] Database backup created and verified
  - [ ] Backup restoration procedure tested within last 30 days
  - [ ] Previous deployment configuration saved for rollback

- [ ] **Stakeholder Communication**
  - [ ] Deployment window announced (email, Slack, Jira)
  - [ ] On-call engineers notified
  - [ ] Customer communications prepared (if user-facing)

---

## Deployment Phase

### Docker Compose Deployment

```bash
# 1. Pull latest images
docker compose -f deployments/compose/compose.yml pull

# 2. Verify images available
docker images | grep cryptoutil

# 3. Start services (detached mode)
docker compose -f deployments/compose/compose.yml up -d

# 4. Monitor startup logs
docker compose -f deployments/compose/compose.yml logs -f --tail=50
```

### Kubernetes Deployment (Production)

```bash
# 1. Apply ConfigMaps and Secrets first
kubectl apply -f deployments/kubernetes/configmaps/ -n cryptoutil-prod
kubectl apply -f deployments/kubernetes/secrets/ -n cryptoutil-prod

# 2. Deploy services
kubectl apply -f deployments/kubernetes/deployments/ -n cryptoutil-prod

# 3. Verify rollout status
kubectl rollout status deployment/cryptoutil-kms -n cryptoutil-prod
kubectl rollout status deployment/identity-authz -n cryptoutil-prod
kubectl rollout status deployment/identity-idp -n cryptoutil-prod
kubectl rollout status deployment/identity-rs -n cryptoutil-prod

# 4. Check pod status
kubectl get pods -n cryptoutil-prod
```

### Health Check Validation

- [ ] **Wait for Services to Start** (90-120 seconds)

- [ ] **Verify Docker Compose Health**:

```bash
# All services should show "healthy" status
docker compose -f deployments/compose/compose.yml ps
```

- [ ] **Verify Kubernetes Health**:

```bash
# All pods should show "Running" and pass readiness probes
kubectl get pods -n cryptoutil-prod
```

- [ ] **Test Health Endpoints**:

```bash
# KMS Server (if deployed)
curl -k https://127.0.0.1:9090/admin/v1/livez    # Liveness
curl -k https://127.0.0.1:9090/admin/v1/readyz   # Readiness

# Identity AuthZ (if deployed)
curl -k https://127.0.0.1:8080/health

# Identity IdP (if deployed)
curl -k https://127.0.0.1:8081/health

# Identity RS (if deployed)
curl -k https://127.0.0.1:8082/health
```

- [ ] **Database Connectivity**:

```bash
# PostgreSQL readiness
docker compose exec postgres pg_isready -U <username> -d <database>
```

---

## Post-Deployment Phase

### Functional Validation

- [ ] **Smoke Tests** (Critical Paths)
  - [ ] OAuth 2.1 authorization code flow (if Identity deployed)
  - [ ] Token introspection endpoint (if Identity deployed)
  - [ ] Resource access with valid token (if Identity deployed)
  - [ ] KMS encryption/decryption operations (if KMS deployed)
  - [ ] Key generation endpoints (if KMS deployed)

- [ ] **API Endpoint Validation**:

```bash
# Test Swagger UI accessible
curl -k https://127.0.0.1:8080/ui/swagger/doc.json

# Verify OpenAPI spec returns valid JSON
```

- [ ] **Telemetry Validation**:
  - [ ] OTLP collector receiving telemetry (<http://127.0.0.1:13133/>)
  - [ ] Grafana dashboards showing metrics (<http://127.0.0.1:3000>)
  - [ ] Logs flowing to Loki
  - [ ] Traces available in Tempo

### Performance Monitoring

- [ ] **Response Time Validation**:
  - [ ] Health endpoints respond < 100ms
  - [ ] Authorization endpoints respond < 500ms
  - [ ] Token generation responds < 1000ms

- [ ] **Resource Utilization**:
  - [ ] CPU usage within acceptable limits (<70% average)
  - [ ] Memory usage stable (no leaks detected)
  - [ ] Database connection pool healthy

- [ ] **Error Rate Monitoring**:
  - [ ] HTTP 5xx errors < 0.1%
  - [ ] Application error logs reviewed
  - [ ] No panic/crash logs observed

### Security Verification

- [ ] **TLS Configuration**:
  - [ ] All endpoints serving HTTPS with valid certificates
  - [ ] No SSL/TLS errors in logs
  - [ ] Certificate expiration > 30 days

- [ ] **Authentication/Authorization**:
  - [ ] Unauthorized requests properly rejected (401/403)
  - [ ] Token validation working correctly
  - [ ] Scope enforcement validated

### Documentation Updates

- [ ] **Deployment Record**:
  - [ ] Deployment timestamp recorded
  - [ ] Version/commit hash documented
  - [ ] Configuration changes noted
  - [ ] Issues encountered and resolutions logged

- [ ] **Stakeholder Notification**:
  - [ ] Deployment completion announced
  - [ ] Known issues communicated
  - [ ] Next monitoring window specified

---

## Rollback Procedures

### Rollback Triggers

Execute rollback immediately if any of these conditions occur:

- **CRITICAL**: Service health checks failing for >5 minutes
- **CRITICAL**: Error rate >5% for >3 minutes
- **HIGH**: Performance degradation >50% slower than baseline
- **HIGH**: Security incident detected (unauthorized access, data breach)
- **MEDIUM**: Functional defects in critical user flows

### Docker Compose Rollback

```bash
# 1. Stop new deployment
docker compose -f deployments/compose/compose.yml down -v

# 2. Restore previous configuration
git checkout <previous-commit> -- deployments/compose/

# 3. Restart with previous version
docker compose -f deployments/compose/compose.yml up -d

# 4. Verify health
docker compose -f deployments/compose/compose.yml ps
```

### Kubernetes Rollback

```bash
# 1. Rollback to previous revision
kubectl rollout undo deployment/cryptoutil-kms -n cryptoutil-prod
kubectl rollout undo deployment/identity-authz -n cryptoutil-prod
kubectl rollout undo deployment/identity-idp -n cryptoutil-prod
kubectl rollout undo deployment/identity-rs -n cryptoutil-prod

# 2. Monitor rollback status
kubectl rollout status deployment/cryptoutil-kms -n cryptoutil-prod

# 3. Verify pods healthy
kubectl get pods -n cryptoutil-prod
```

### Rollback Validation

- [ ] **Health Checks Passing**:
  - [ ] All services showing healthy status
  - [ ] Database connectivity restored

- [ ] **Functional Testing**:
  - [ ] Critical user flows operational
  - [ ] API endpoints responding correctly

- [ ] **Communication**:
  - [ ] Stakeholders notified of rollback
  - [ ] Incident ticket created
  - [ ] Root cause analysis scheduled

### Post-Rollback Actions

- [ ] **Incident Documentation**:
  - [ ] Rollback reason documented
  - [ ] Timeline of events recorded
  - [ ] Root cause identified (if known)
  - [ ] Action items for prevention

- [ ] **Lessons Learned**:
  - [ ] Team debrief scheduled within 48 hours
  - [ ] Process improvements identified
  - [ ] Testing gaps addressed

---

## Monitoring Dashboard

### Key Metrics to Monitor (First 24 Hours)

**Service Health**:

- Container/pod status (running, healthy)
- Health check success rate (target: 100%)
- Service uptime (target: 99.9%)

**Performance Metrics**:

- Request latency (p50, p95, p99)
- Throughput (requests/second)
- Error rate (2xx, 4xx, 5xx responses)

**Resource Utilization**:

- CPU usage (% per container/pod)
- Memory usage (MB per container/pod)
- Database connections (active, idle)

**Business Metrics** (Identity Services):

- Authorization requests/minute
- Token issuance rate
- Failed authentication attempts

**Business Metrics** (KMS):

- Encryption operations/minute
- Decryption operations/minute
- Key generation rate

### Grafana Dashboard URLs

- **Service Overview**: <http://127.0.0.1:3000/d/cryptoutil-overview>
- **Performance Metrics**: <http://127.0.0.1:3000/d/cryptoutil-performance>
- **Error Tracking**: <http://127.0.0.1:3000/d/cryptoutil-errors>

---

## Emergency Contacts

**On-Call Engineer**: See PagerDuty rotation
**Database Admin**: [Contact Info]
**Platform Team**: [Contact Info]
**Security Team**: [Contact Info]

---

## References

- [Docker Compose Configuration](../../deployments/compose/compose.yml)
- [Docker Compose Identity Demo](../../deployments/identity/compose.advanced.yml)
- [Kubernetes Deployment Guide](../../deployments/kubernetes/README.md) (when created)
- [Operational Runbook](./operational-runbook.md) (when created)
- [Token Rotation Runbook](../02-identityV2/historical/token-rotation-runbook.md)
- [Adaptive Auth Operations](./adaptive-auth-operations.md)
- [Health Check Troubleshooting](../02-identityV2/historical/identity-docker-quickstart.md#troubleshooting)

---

## Appendix: Known Limitations

### Identity V2 Integration Status

**BLOCKED**: Identity servers (authz, idp, rs) are not yet integrated into main `cryptoutil` binary:

- Standalone binaries exist in `cmd/identity/{authz,idp,rs}/main.go`
- Docker Compose files reference `cryptoutil identity` commands that are not implemented
- Resolution required before production deployment of Identity V2 services

**Workaround**: Deploy KMS server only until Identity V2 integration complete

### Tooling Requirements

**act (GitHub Actions Local Runner)**:

- Required for DAST scanning workflows
- Installation: `choco install act-cli` (Windows)
- See [Development Setup Guide](../DEV-SETUP.md)

---

**Checklist Version**: 1.0.0
**Last Reviewed**: 2025-11-24
**Next Review**: 2025-12-24 (monthly cadence)
