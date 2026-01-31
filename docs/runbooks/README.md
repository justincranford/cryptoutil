# Runbooks

Operational runbooks for CryptoUtil production alerts.

## Index

### Service Health

- [service-down.md](service-down.md) - Service not responding
- [health-check.md](health-check.md) - Readiness check failing

### Performance

- [high-latency.md](high-latency.md) - Response time exceeds SLA
- [high-error-rate.md](high-error-rate.md) - Error rate exceeds threshold

### Resources

- [high-memory.md](high-memory.md) - Memory usage too high
- [high-cpu.md](high-cpu.md) - CPU usage too high
- [goroutine-leak.md](goroutine-leak.md) - Too many goroutines

### Database

- [db-connections.md](db-connections.md) - Connection pool exhaustion
- [db-latency.md](db-latency.md) - Query latency too high
- [db-errors.md](db-errors.md) - Database error rate

### Security

- [auth-failures.md](auth-failures.md) - Authentication failures
- [cert-expiry.md](cert-expiry.md) - Certificate expiring soon

## Usage

1. Alert fires in Grafana/Alertmanager
2. Click runbook URL in alert annotation
3. Follow investigation steps
4. Apply resolution steps
5. Escalate if needed
6. Document post-incident findings

## Contributing

When adding new alerts:

1. Create corresponding runbook in this directory
2. Add `runbook_url` annotation to alert rule
3. Update this README
