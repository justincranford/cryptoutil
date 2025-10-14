---
description: "Instructions for observability and monitoring implementation"
applyTo: "**"
---
# Observability and Monitoring Instructions

- Use [OpenTelemetry](https://opentelemetry.io/) for tracing, metrics, logging
- Use structured logging, OTLP export, health endpoints, Prometheus metrics, proper log levels, no sensitive data

## Telemetry Forwarding Architecture

**MANDATORY**: All telemetry from cryptoutil services MUST be forwarded through the otel-contrib sidecar (opentelemetry-collector) to upstream telemetry platforms (e.g., grafana/otel-lgtm).

### Dual Telemetry Flows for Complete Observability

**Application Telemetry (Push-based):**
```
cryptoutil services (OTLP gRPC:4317) → OpenTelemetry Collector Contrib → Grafana-OTEL-LGTM (OTLP HTTP:4318)
```
- **Purpose**: Business application traces, logs, and metrics
- **Protocol**: OTLP (OpenTelemetry Protocol) - push-based
- **Data**: Crypto operations, API calls, business logic telemetry

**Infrastructure Telemetry (Pull-based):**
```
Grafana-OTEL-LGTM (Prometheus) → OpenTelemetry Collector Contrib (HTTP:8889/metrics)
```
- **Purpose**: Monitor collector health and performance
- **Protocol**: Prometheus scraping - pull-based
- **Data**: Collector throughput, error rates, queue depths, resource usage

**Why Both Flows?** The collector both **receives application telemetry** (from cryptoutil) and **exposes its own metrics** (for monitoring). This provides complete observability of both your application and the telemetry pipeline itself.

### Architecture Flow:
```
cryptoutil services → opentelemetry-collector (sidecar) → upstream telemetry platforms
```

### Configuration Requirements:
- `cryptoutil-otel.yml` MUST point to `opentelemetry-collector:4317`
- **NEVER** configure `cryptoutil-otel.yml` to forward directly to upstream platforms (e.g., `grafana:4317`)
- The otel-contrib sidecar handles processing, filtering, and routing before forwarding to upstream platforms

### Rationale:
- Ensures centralized telemetry processing and filtering
- Maintains consistent architecture across environments
- Enables future enhancements (sampling, aggregation, etc.) in the sidecar layer
- Prevents direct coupling between application services and telemetry platforms
