# Runbook: Database Errors

## Alert: DatabaseErrors

**Severity**: Critical
**Alert Expression**: `rate(db_errors_total{job=~"cryptoutil.*"}[5m]) > 0.1`
**Duration**: 3 minutes

## Description

Database error rate exceeds 0.1 per second (6 errors/minute).

## Impact

- Failed transactions
- Data inconsistency risk
- Service degradation

## Investigation Steps

### 1. Check PostgreSQL Logs

```bash
# Docker
docker compose logs cryptoutil-postgres --tail=200 | grep -i "error\|fatal\|panic"

# System PostgreSQL
tail -100 /var/log/postgresql/postgresql-*.log | grep -i error
```

### 2. Check Application Logs

```bash
docker compose logs <service-name> --tail=200 | grep -i "database\|sql\|postgres"
```

### 3. Verify Database Status

```bash
psql -h localhost -U cryptoutil -c "SELECT pg_is_in_recovery();"
psql -h localhost -U cryptoutil -c "SELECT count(*) FROM pg_stat_activity;"
```

### 4. Check for Deadlocks

```bash
psql -h localhost -U cryptoutil -c "
SELECT blocked_locks.pid AS blocked_pid,
       blocked_activity.usename AS blocked_user,
       blocking_locks.pid AS blocking_pid,
       blocking_activity.usename AS blocking_user,
       blocked_activity.query AS blocked_statement,
       blocking_activity.query AS blocking_statement
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;"
```

## Resolution Steps

### Connection Errors

1. Check database is running
2. Verify connection string
3. Check network connectivity

### Constraint Violations

1. Review application logic
2. Add proper validation
3. Fix data integrity issues

### Deadlocks

1. Optimize transaction ordering
2. Reduce transaction scope
3. Add retry logic

### Disk Full

1. Clear old data
2. Expand storage
3. Archive historical data

## Escalation

- **Immediately**: Page on-call engineer
- **After 10 minutes**: Page engineering lead

## Post-Incident

1. Add specific error monitoring
2. Implement retry patterns
3. Review transaction patterns
4. Add database health checks
