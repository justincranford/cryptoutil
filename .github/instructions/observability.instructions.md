---
description: "Instructions for observability and monitoring implementation"
applyTo: "**"
---
# Observability and Monitoring Instructions

- Integrate OpenTelemetry for distributed tracing, metrics, and logging
- Use structured logging with slog for all application logs
- Export telemetry data via OTLP protocol for production environments
- Provide console output option for development environments
- Implement proper span creation and correlation for distributed tracing
- Use appropriate log levels (DEBUG, INFO, WARN, ERROR) consistently
- Include contextual information (transaction IDs, request IDs) in all logs
- Implement health check endpoints for Kubernetes readiness/liveness probes
- Provide Prometheus-compatible metrics for monitoring
- Log security events separately with appropriate detail levels
- Implement graceful shutdown with proper telemetry cleanup
- Use telemetry providers for traces, metrics, and logs consistently
- Include performance metrics for cryptographic operations
- Log database connection pool status and query performance
- Ensure sensitive data is never logged (keys, passwords, etc.)
- Use proper error correlation and stack trace capture
