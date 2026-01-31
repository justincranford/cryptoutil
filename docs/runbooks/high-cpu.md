# Runbook: High CPU Usage

## Alert: HighCPUUsage

**Severity**: Warning
**Alert Expression**: `rate(process_cpu_seconds_total{job=~"cryptoutil.*"}[5m]) > 0.8`
**Duration**: 5 minutes

## Description

Service CPU usage exceeds 80% for 5 minutes.

## Impact

- Increased latency
- Potential request timeouts
- Service degradation

## Investigation Steps

### 1. Check CPU Usage

```bash
# Docker
docker stats --no-stream

# System-wide
top -bn1 | head -20
```

### 2. Profile CPU with pprof

```bash
# 30-second CPU profile
curl -k "https://127.0.0.1:9090/debug/pprof/profile?seconds=30" > cpu.prof
go tool pprof -http=:8080 cpu.prof
```

### 3. Check Goroutines

```bash
# Goroutine profile
curl -k https://127.0.0.1:9090/debug/pprof/goroutine?debug=2 | head -100
```

### 4. Check Request Rate

```bash
# Verify if traffic spike
# Query: rate(http_server_request_duration_seconds_count[5m])
```

## Resolution Steps

### Traffic Spike

1. Scale horizontally (add instances)
2. Enable rate limiting
3. Activate caching

### Inefficient Code

1. Analyze CPU profile
2. Optimize hot paths
3. Deploy fix

### Crypto Operations

1. Check key generation load
2. Implement key caching
3. Use connection pooling for HSM

### Infinite Loop / Busy Wait

1. Identify problematic goroutine
2. Fix code and redeploy
3. Implement timeout guards

## Escalation

- **After 15 minutes**: Notify on-call engineer
- **If CPU > 95%**: Page immediately

## Post-Incident

1. Add CPU profiling to performance tests
2. Implement CPU budgets
3. Add auto-scaling triggers
