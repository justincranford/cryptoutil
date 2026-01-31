# Runbook: High Error Rate

## Alert: HighErrorRate

**Severity**: Critical
**Alert Expression**: `sum(rate(http_server_request_duration_seconds_count{status=~"5.."}[5m])) / sum(rate(http_server_request_duration_seconds_count[5m])) > 0.05`
**Duration**: 5 minutes

## Description

More than 5% of requests are returning 5xx errors.

## Impact

- Service degradation
- User-facing errors
- Potential data inconsistency

## Investigation Steps

### 1. Identify Error Types

```bash
# Check Grafana for error breakdown
# Query: sum by (status) (rate(http_server_request_duration_seconds_count{status=~"5.."}[5m]))
```

### 2. Review Error Logs

```bash
docker compose logs <service-name> --tail=500 | grep -E "error|panic|fatal"
```

### 3. Check Specific Endpoints

```bash
# Identify failing endpoints
# Query: topk(5, sum by (path) (rate(http_server_request_duration_seconds_count{status=~"5.."}[5m])))
```

### 4. Check External Dependencies

```bash
# Database
psql -h localhost -U cryptoutil -c "SELECT 1"

# External services
curl -k https://dependency-service/health
```

## Resolution Steps

### Internal Server Errors (500)

1. Check application logs for stack traces
2. Identify root cause
3. Deploy fix or rollback

### Bad Gateway (502)

1. Check upstream service health
2. Verify network connectivity
3. Restart upstream service

### Service Unavailable (503)

1. Check if service is overloaded
2. Scale up instances
3. Implement circuit breaker

### Gateway Timeout (504)

1. Check slow dependencies
2. Increase timeout thresholds
3. Optimize slow operations

## Escalation

- **Immediately**: Page on-call engineer
- **After 10 minutes**: Page engineering lead
- **After 30 minutes**: Incident commander

## Post-Incident

1. Add specific error monitoring
2. Implement retry logic if missing
3. Add circuit breakers
4. Update SLO dashboards
