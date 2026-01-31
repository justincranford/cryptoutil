# Runbook: High Memory Usage

## Alert: HighMemoryUsage

**Severity**: Warning
**Alert Expression**: `process_resident_memory_bytes{job=~"cryptoutil.*"} / 1024 / 1024 > 512`
**Duration**: 5 minutes

## Description

Service memory usage exceeds 512MB.

## Impact

- Potential OOM kills
- Performance degradation
- Service instability

## Investigation Steps

### 1. Check Current Memory Usage

```bash
# Docker
docker stats --no-stream

# Process memory
ps aux --sort=-%mem | head -10
```

### 2. Profile Memory with pprof

```bash
# Get heap profile
curl -k https://127.0.0.1:9090/debug/pprof/heap > heap.prof
go tool pprof -http=:8080 heap.prof
```

### 3. Check for Memory Leaks

```bash
# Compare heap profiles over time
curl -k https://127.0.0.1:9090/debug/pprof/heap > heap1.prof
sleep 60
curl -k https://127.0.0.1:9090/debug/pprof/heap > heap2.prof
go tool pprof -base heap1.prof heap2.prof
```

### 4. Check Goroutine Count

```bash
curl -k https://127.0.0.1:9090/debug/pprof/goroutine?debug=1 | head -20
```

## Resolution Steps

### Memory Leak

1. Identify leaking allocations with pprof
2. Fix code and redeploy
3. Monitor recovery

### Large Cache

1. Review cache size limits
2. Implement eviction policies
3. Tune cache TTLs

### Too Many Goroutines

1. Identify goroutine leaks
2. Implement proper context cancellation
3. Add goroutine pool limits

### Temporary Fix

1. Restart service to reclaim memory
2. Increase memory limits temporarily
3. Scale horizontally

## Escalation

- **After 15 minutes**: Notify on-call engineer
- **If memory > 80%**: Page immediately

## Post-Incident

1. Add memory profiling to CI
2. Implement memory budget alerts
3. Add load testing for memory patterns
