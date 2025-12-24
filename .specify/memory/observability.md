# Observability and Monitoring

**Referenced by**: `.github/instructions/02-05.observability.instructions.md`

## Core Observability Stack

**MANDATORY: Use OpenTelemetry for all observability**:

- Tracing: Distributed request tracking across services
- Metrics: Performance counters, gauges, histograms
- Logging: Structured logs with trace correlation

**Standards**:

- OpenTelemetry SDK: <https://opentelemetry.io/docs/>
- OTLP Protocol: gRPC (4317) or HTTP (4318)
- Collector: OpenTelemetry Collector Contrib
- Backend: Grafana OTEL-LGTM (Loki, Tempo, Prometheus, Mimir)

---

## Telemetry Forwarding Architecture - CRITICAL

### MANDATORY Pattern: Sidecar Architecture

**ALL telemetry from cryptoutil services MUST be forwarded through otel-contrib sidecar**:

```
cryptoutil services → opentelemetry-collector (sidecar) → upstream telemetry platforms
```

**NEVER configure cryptoutil services to send telemetry directly to upstream platforms**:

- ❌ WRONG: `cryptoutil → grafana-otel-lgtm` (direct coupling)
- ✅ CORRECT: `cryptoutil → otel-collector → grafana-otel-lgtm` (sidecar pattern)

### Push-Based Telemetry Flow

**Application Telemetry** (Push-based from services):

```
cryptoutil services (OTLP gRPC:4317 or HTTP:4318) → OpenTelemetry Collector Contrib → Grafana-OTEL-LGTM (OTLP gRPC:14317 or HTTP:14318)
```

**Details**:

- **Purpose**: Business application traces, logs, and metrics
- **Protocol**: OTLP (OpenTelemetry Protocol) - push-based
- **Supported Protocols**:
  - **gRPC Protocol**: `grpc://host:port` - Efficient binary protocol for high-performance telemetry
  - **HTTP Protocol**: `http://host:port` or `https://host:port` - Universal compatibility, firewall-friendly
- **Configuration Guidelines**:
  - Use gRPC for internal service-to-service communication (default, more efficient)
  - Use HTTP for environments with restrictive firewalls or universal compatibility needs
  - Both protocols support traces, metrics, and logs
  - Endpoint format: `protocol://hostname:port` (e.g., `grpc://otel-collector:4317`, `http://otel-collector:4318`)
- **Data**: Crypto operations, API calls, business logic telemetry

**Collector Self-Monitoring** (Push-based from collector):

```
OpenTelemetry Collector Contrib (internal) → Grafana-OTEL-LGTM (OTLP HTTP:14318)
```

**Details**:

- **Purpose**: Monitor collector health and performance
- **Protocol**: OTLP - push-based (collector exports its own telemetry)
- **Data**: Collector throughput, error rates, queue depths, resource usage

### Configuration Requirements - MANDATORY

**cryptoutil Service Configuration**:

- `cryptoutil-otel.yml` MUST point to `opentelemetry-collector:4317` (or `:4318` for HTTP)
- NEVER configure cryptoutil to bypass otel-collector-contrib sidecar
- NEVER configure cryptoutil to send telemetry directly to grafana-otel-lgtm

**Collector Configuration**:

- The otel-contrib sidecar handles processing, filtering, and routing before forwarding to upstream platforms
- Collector receives telemetry on ports 4317 (gRPC) and 4318 (HTTP)
- Collector forwards to grafana-otel-lgtm on ports 14317 (gRPC) or 14318 (HTTP)

**Grafana Configuration**:

- Grafana receives telemetry ONLY from otel-collector (no direct service connections)
- NO Prometheus scraping of collector metrics (collector pushes its own metrics via OTLP)

### Rationale for Sidecar Pattern

**Benefits**:

- **Centralized Processing**: Single point for sampling, filtering, aggregation
- **Consistent Architecture**: Same pattern across all environments (dev, staging, prod)
- **Future Extensibility**: Easy to add sampling, enrichment, routing rules in sidecar
- **Decoupling**: Services don't need to know about downstream telemetry platforms
- **Failure Isolation**: Collector failures don't crash services (buffering, retry)

**Why NOT Direct Service → Platform**:

- Tight coupling between services and telemetry backend
- Harder to change backends (requires service reconfig/redeploy)
- No centralized processing/filtering layer
- Services need to handle telemetry backend failures

---

## Structured Logging - MANDATORY

**ALWAYS use structured logging with key-value pairs**:

```go
import "go.uber.org/zap"

logger.Info("Key created",
    zap.String("key_id", keyID),
    zap.String("algorithm", "RSA-2048"),
    zap.String("user_id", userID),
    zap.Duration("duration", elapsed),
)
```

**Benefits**:

- Machine-readable (JSON format)
- Easy querying and filtering
- Trace/span correlation via trace_id/span_id
- Consistent field names across services

**Standard Fields**:

- `timestamp` - ISO 8601 UTC
- `level` - DEBUG, INFO, WARN, ERROR, FATAL
- `message` - Human-readable description
- `trace_id` - OpenTelemetry trace ID (for correlation)
- `span_id` - OpenTelemetry span ID (for correlation)
- `service.name` - Service identifier
- `service.version` - Service version

---

## OTLP Export Configuration

**gRPC Endpoint** (Preferred):

```yaml
observability:
  otlp:
    protocol: grpc
    endpoint: opentelemetry-collector:4317
    service_name: cryptoutil-kms
    service_version: 1.0.0
```

**HTTP Endpoint** (Firewall-Friendly):

```yaml
observability:
  otlp:
    protocol: http
    endpoint: http://opentelemetry-collector:4318
    service_name: cryptoutil-kms
    service_version: 1.0.0
```

**Key Configuration Options**:

- `protocol`: `grpc` or `http`
- `endpoint`: Hostname and port (NO scheme for gRPC, include http:// or https:// for HTTP)
- `service_name`: Unique identifier for service (used in traces/metrics/logs)
- `service_version`: Version string (for rollback correlation)
- `insecure`: `true` for dev (no TLS), `false` for prod (TLS required)

---

## Health Endpoints - MANDATORY

**ALL services MUST expose health check endpoints**:

**Liveness Probe** (`/admin/v1/livez`):

- **Purpose**: Is the process alive?
- **Check**: Lightweight (process running, goroutines not deadlocked)
- **Failure Action**: Restart container/process
- **Response**: 200 OK (alive), 503 Service Unavailable (dead)

**Readiness Probe** (`/admin/v1/readyz`):

- **Purpose**: Is the service ready to accept traffic?
- **Check**: Heavyweight (database connected, dependencies healthy)
- **Failure Action**: Remove from load balancer (don't restart)
- **Response**: 200 OK (ready), 503 Service Unavailable (not ready)

**Why Separate Probes**:

- **Liveness**: Process stuck? Restart it
- **Readiness**: Dependencies down? Remove from LB, don't restart
- Combined health endpoint can't distinguish these two failure modes

**Pattern**:

```go
// Liveness: Lightweight check
func (s *Server) Livez(ctx context.Context) error {
    // Check if goroutines are deadlocked (optional)
    // Check if critical channels are blocked (optional)
    return nil  // Process is alive
}

// Readiness: Heavyweight check
func (s *Server) Readyz(ctx context.Context) error {
    // Check database connection
    if err := s.db.Ping(ctx); err != nil {
        return fmt.Errorf("database not ready: %w", err)
    }

    // Check dependent services (if federated)
    if err := s.checkDependencies(ctx); err != nil {
        return fmt.Errorf("dependencies not ready: %w", err)
    }

    return nil  // Service is ready
}
```

---

## Prometheus Metrics - MANDATORY

**Standard Metrics for ALL Services**:

**HTTP Server Metrics**:

- `http_requests_total{method, path, status}` - Counter of HTTP requests
- `http_request_duration_seconds{method, path}` - Histogram of request durations
- `http_requests_in_flight{method, path}` - Gauge of concurrent requests

**Database Metrics**:

- `db_connections_open` - Gauge of open database connections
- `db_connections_idle` - Gauge of idle database connections
- `db_query_duration_seconds{operation}` - Histogram of query durations
- `db_errors_total{operation}` - Counter of database errors

**Crypto Operation Metrics**:

- `crypto_operations_total{algorithm, operation}` - Counter of crypto operations
- `crypto_operation_duration_seconds{algorithm, operation}` - Histogram of operation durations
- `crypto_errors_total{algorithm, operation}` - Counter of crypto errors

**Key Management Metrics**:

- `keys_total{algorithm, status}` - Gauge of total keys
- `key_rotations_total{algorithm}` - Counter of key rotations
- `key_usage_total{key_id, operation}` - Counter of key usage

---

## Log Levels - MANDATORY

**Use appropriate log levels**:

**DEBUG**:

- Detailed diagnostic information
- Function entry/exit with parameters
- Intermediate computation results
- Only enabled in development

**INFO**:

- Significant application events
- Service startup/shutdown
- Configuration loaded
- Key created/rotated/deleted
- Request/response summaries

**WARN**:

- Degraded mode (using fallback, cache miss)
- Recoverable errors (retry succeeded)
- Deprecated API usage
- Approaching resource limits

**ERROR**:

- Unrecoverable errors (database unavailable)
- Request failures (400, 500 errors)
- External service failures
- Data integrity issues

**FATAL**:

- Unrecoverable startup errors
- Critical configuration missing
- Required dependencies unavailable
- Process must terminate

---

## Sensitive Data Protection - MANDATORY

**NEVER log sensitive data**:

- ❌ Passwords, API keys, tokens
- ❌ Private keys, certificates
- ❌ Personally Identifiable Information (PII)
- ❌ Credit card numbers, SSNs
- ❌ Session IDs, cookies

**Safe to Log**:

- ✅ Public key IDs (not keys themselves)
- ✅ User IDs (NOT usernames/emails if PII)
- ✅ Resource IDs (key IDs, certificate IDs)
- ✅ Operation types (create, delete, rotate)
- ✅ Durations, counts, errors

**Pattern for Sensitive Fields**:

```go
// ❌ WRONG: Logs full key
logger.Info("Key created", zap.Any("key", key))

// ✅ CORRECT: Logs only key ID
logger.Info("Key created",
    zap.String("key_id", key.ID),
    zap.String("algorithm", key.Algorithm),
)
```

---

## Cross-References

**Related Documentation**:

- Service template: `.specify/memory/service-template.md`
- Docker Compose: `.specify/memory/docker.md`
- Security: `.specify/memory/security.md`

**Tools**:

- OpenTelemetry: <https://opentelemetry.io/>
- Grafana OTEL-LGTM: <https://github.com/grafana/docker-otel-lgtm>
- Prometheus: <https://prometheus.io/>
