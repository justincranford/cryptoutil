# Runbook: Goroutine Leak

## Alert: TooManyGoroutines

**Severity**: Warning
**Alert Expression**: `go_goroutines{job=~"cryptoutil.*"} > 1000`
**Duration**: 5 minutes

## Description

Service has more than 1000 goroutines running.

## Impact

- High memory usage
- Scheduling overhead
- Potential deadlocks
- Performance degradation

## Investigation Steps

### 1. Get Current Goroutine Count

```bash
curl -k https://127.0.0.1:9090/debug/pprof/goroutine?debug=1 | head -5
```

### 2. Analyze Goroutine Stacks

```bash
# Full stack traces
curl -k https://127.0.0.1:9090/debug/pprof/goroutine?debug=2 > goroutines.txt
cat goroutines.txt | grep -A5 "goroutine" | head -100
```

### 3. Identify Common Patterns

```bash
# Count goroutines by function
curl -k https://127.0.0.1:9090/debug/pprof/goroutine?debug=1 | \
  grep -E "^[0-9]+ @" | sort | uniq -c | sort -rn | head -20
```

### 4. Monitor Over Time

```bash
# Watch goroutine count
watch -n 5 'curl -sk https://127.0.0.1:9090/debug/pprof/goroutine?debug=1 | head -5'
```

## Resolution Steps

### Connection Leak

1. Identify unclosed connections
2. Implement proper defer close patterns
3. Add connection pool limits

### Channel Leak

1. Find blocking channel operations
2. Add context cancellation
3. Use buffered channels appropriately

### HTTP Client Leak

1. Ensure response bodies are read and closed
2. Set client timeouts
3. Implement connection pooling

### Worker Pool Issue

1. Review worker pool implementation
2. Add proper shutdown handling
3. Implement bounded concurrency

## Immediate Mitigation

```bash
# Restart service to clear goroutines
docker compose restart <service-name>
```

## Escalation

- **After 15 minutes**: Notify on-call engineer
- **If goroutines > 5000**: Page immediately

## Post-Incident

1. Add goroutine profiling to CI
2. Implement goroutine budgets
3. Add load testing for concurrency patterns
4. Review context propagation
