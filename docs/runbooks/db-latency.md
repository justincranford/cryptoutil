# Runbook: High Database Latency

## Alert: HighDatabaseLatency

**Severity**: Warning
**Alert Expression**: `histogram_quantile(0.95, sum(rate(db_query_duration_seconds_bucket[5m])) by (le)) > 0.5`
**Duration**: 5 minutes

## Description

Database query p95 latency exceeds 500ms.

## Impact

- Slow API responses
- Client timeouts
- Degraded user experience

## Investigation Steps

### 1. Identify Slow Queries

```bash
psql -h localhost -U cryptoutil -c "
SELECT query, 
       calls, 
       mean_time,
       total_time,
       rows
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 20;"
```

### 2. Check Query Plans

```bash
psql -h localhost -U cryptoutil -c "EXPLAIN ANALYZE <slow_query>;"
```

### 3. Check Table Statistics

```bash
psql -h localhost -U cryptoutil -c "
SELECT schemaname, relname, 
       n_live_tup, n_dead_tup,
       last_vacuum, last_autovacuum,
       last_analyze, last_autoanalyze
FROM pg_stat_user_tables
ORDER BY n_live_tup DESC;"
```

### 4. Check Index Usage

```bash
psql -h localhost -U cryptoutil -c "
SELECT schemaname, relname, indexrelname, 
       idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;"
```

## Resolution Steps

### Missing Indexes

1. Identify unindexed columns in WHERE clauses
2. Create appropriate indexes
3. Monitor query performance

### Table Bloat

1. Run VACUUM ANALYZE:
   ```sql
   VACUUM ANALYZE table_name;
   ```
2. Consider VACUUM FULL for extreme bloat

### Lock Contention

1. Identify blocking queries:
   ```sql
   SELECT * FROM pg_blocking_pids(pid);
   ```
2. Optimize transaction scope
3. Review isolation levels

### Hardware Limits

1. Check disk I/O metrics
2. Consider read replicas
3. Upgrade to faster storage

## Escalation

- **After 15 minutes**: Notify on-call engineer
- **If latency > 2s**: Page immediately

## Post-Incident

1. Add slow query logging
2. Implement query performance tests
3. Create index maintenance schedule
4. Consider query caching
