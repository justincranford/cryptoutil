# Runbook: Health Check Failing

## Alert: HealthCheckFailing

**Severity**: Warning
**Alert Expression**: `http_server_request_duration_seconds_count{path="/admin/api/v1/readyz", status!="200"} > 0`
**Duration**: 3 minutes

## Description

A service's readiness endpoint is returning non-200 status codes.

## Impact

- Service may be removed from load balancer
- Requests may fail
- Degraded availability

## Investigation Steps

### 1. Check Readiness Endpoint

```bash
curl -k https://127.0.0.1:9090/admin/api/v1/readyz
```

### 2. Check Dependencies

```bash
# Database connectivity
psql -h localhost -U cryptoutil -c "SELECT 1"

# Check dependent services
curl -k https://127.0.0.1:8080/admin/api/v1/livez
```

### 3. Review Logs

```bash
docker compose logs <service-name> --tail=100 | grep -i "error\|warn"
```

## Resolution Steps

### Database Unavailable

1. Check PostgreSQL is running
2. Verify connection pool not exhausted
3. Check for deadlocks

### Dependency Service Down

1. Identify failing dependency from logs
2. Restart dependency service
3. Monitor recovery

### Configuration Issue

1. Review recent config changes
2. Rollback if necessary
3. Validate configuration

## Escalation

- **After 10 minutes**: Notify on-call engineer
- **After 30 minutes**: Page engineering lead

## Post-Incident

1. Identify root cause
2. Add monitoring for specific dependency
3. Update runbook with findings
