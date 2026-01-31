# Runbook: High Latency

## Alert: HighLatency

**Severity**: Warning
**Alert Expression**: `histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le)) > 1`
**Duration**: 5 minutes

## Description

The 95th percentile response latency exceeds 1 second.

## Impact

- Poor user experience
- Client timeouts
- Potential cascading failures

## Investigation Steps

### 1. Identify Slow Endpoints

```bash
# Check Grafana for slow endpoints
# Query: topk(10, histogram_quantile(0.95, rate(http_server_request_duration_seconds_bucket[5m])) by (path))
```

### 2. Check Database Performance

```bash
# PostgreSQL slow queries
SELECT query, calls, mean_time, total_time 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;
```

### 3. Check Resource Utilization

```bash
# CPU and Memory
docker stats

# Goroutines
curl -k https://127.0.0.1:9090/debug/pprof/goroutine?debug=2
```

### 4. Review Recent Changes

```bash
git log --oneline -10
```

## Resolution Steps

### Database Slow

1. Analyze slow queries
2. Add missing indexes
3. Optimize query patterns
4. Consider connection pool tuning

### Resource Exhaustion

1. Scale horizontally (add instances)
2. Scale vertically (increase CPU/memory)
3. Implement caching

### Code Regression

1. Identify problematic commit
2. Rollback if necessary
3. Fix and redeploy

## Escalation

- **After 15 minutes**: Notify on-call engineer
- **After 30 minutes**: Page engineering lead

## Post-Incident

1. Profile application with pprof
2. Add latency SLO dashboards
3. Implement performance tests
