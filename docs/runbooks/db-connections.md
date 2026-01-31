# Runbook: Database Connection Exhaustion

## Alert: DatabaseConnectionExhaustion

**Severity**: Critical
**Alert Expression**: `db_open_connections{job=~"cryptoutil.*"} / db_max_open_connections > 0.9`
**Duration**: 3 minutes

## Description

Database connection pool is 90%+ exhausted.

## Impact

- New connections blocked
- Request timeouts
- Service degradation
- Potential cascading failures

## Investigation Steps

### 1. Check Connection Pool Status

```bash
# PostgreSQL active connections
psql -h localhost -U cryptoutil -c "
SELECT count(*) as total,
       state,
       wait_event_type,
       query
FROM pg_stat_activity
WHERE datname = 'cryptoutil'
GROUP BY state, wait_event_type, query
ORDER BY total DESC;"
```

### 2. Identify Long-Running Queries

```bash
psql -h localhost -U cryptoutil -c "
SELECT pid, now() - pg_stat_activity.query_start AS duration, query, state
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > interval '30 seconds'
AND state != 'idle'
ORDER BY duration DESC;"
```

### 3. Check for Connection Leaks

```bash
# Check Grafana for connection trends
# Query: db_open_connections{job="cryptoutil.*"}
```

### 4. Review Application Logs

```bash
docker compose logs <service-name> --tail=200 | grep -i "connection\|pool\|database"
```

## Resolution Steps

### Long-Running Queries

1. Identify and kill problematic queries:
   ```sql
   SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE pid = <pid>;
   ```
2. Investigate query performance
3. Add query timeout limits

### Connection Leak

1. Review recent deployments
2. Check for missing connection closes
3. Deploy fix with proper defer patterns

### Traffic Spike

1. Increase max_open_connections temporarily
2. Scale application horizontally
3. Implement connection pooler (pgbouncer)

### Temporary Fix

```bash
# Increase pool size in config
# Restart service
docker compose restart <service-name>
```

## Escalation

- **Immediately**: Page on-call engineer
- **After 10 minutes**: Page engineering lead

## Post-Incident

1. Implement connection pool monitoring
2. Add connection leak detection tests
3. Review connection timeout settings
4. Consider adding pgbouncer
