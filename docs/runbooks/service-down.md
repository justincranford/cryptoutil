# Runbook: Service Down

## Alert: ServiceDown

**Severity**: Critical
**Alert Expression**: `up{job=~"cryptoutil.*"} == 0`
**Duration**: 1 minute

## Description

A CryptoUtil service is not responding to health checks.

## Impact

- Service unavailable to users
- Dependent services may fail
- Potential data loss if transactions in flight

## Investigation Steps

### 1. Check Service Status

```bash
# Docker Compose
docker compose ps

# Kubernetes
kubectl get pods -l app=cryptoutil
```

### 2. Check Service Logs

```bash
# Docker Compose
docker compose logs <service-name> --tail=100

# Kubernetes
kubectl logs -l app=cryptoutil --tail=100
```

### 3. Check Resource Usage

```bash
# Docker
docker stats

# Kubernetes
kubectl top pods -l app=cryptoutil
```

### 4. Check Network Connectivity

```bash
# Test health endpoint
curl -k https://127.0.0.1:9090/admin/api/v1/livez
```

## Resolution Steps

### Service Crashed

1. Restart the service:
   ```bash
   docker compose restart <service-name>
   ```
2. Monitor logs for crash reason
3. If OOM, increase memory limits

### Network Issue

1. Check Docker/Kubernetes network
2. Verify firewall rules
3. Check DNS resolution

### Database Connection Failed

1. Verify database is running
2. Check connection string
3. Verify credentials

## Escalation

- **After 5 minutes**: Page on-call engineer
- **After 15 minutes**: Page engineering lead
- **After 30 minutes**: Incident commander

## Post-Incident

1. Update incident timeline
2. Conduct root cause analysis
3. Create follow-up issues
4. Update this runbook if needed
