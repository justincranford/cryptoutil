# Cryptoutil Observability & Monitoring TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: October 14, 2025
**Status**: Grafana dashboard expansion planned for Q4 2025

---

## 🟡 MEDIUM - Observability & Monitoring

### Task OB1: Expand Grafana Dashboards for Custom Metrics
- **Description**: Current Grafana dashboard only covers basic HTTP metrics but misses all custom application metrics
- **Current State**: Dashboard shows only `http_requests_total` and `http_request_duration_seconds_bucket` from otelfiber middleware
- **Missing Metrics Categories**:
  - **Pool Performance Metrics**: `cryptoutil.pool.get`, `cryptoutil.pool.permission`, `cryptoutil.pool.generate` histograms
  - **Security Header Metrics**: `security_headers_missing_total` counter
  - **Business Logic Metrics**: None currently implemented but infrastructure ready
- **Action Items**:
  - Create comprehensive dashboard panels for pool performance monitoring
  - Add security metrics dashboard with header compliance tracking
  - Implement business logic metrics for cryptographic operations
  - Update dashboard JSON with proper Prometheus queries for OpenTelemetry metrics
  - Add alerting rules for security header violations and pool performance issues
- **Files**: `deployments/compose/grafana-otel-lgtm/dashboards/cryptoutil.json`
- **Expected Outcome**: Full observability of all custom application metrics
- **Priority**: MEDIUM - Observability improvement
- **Timeline**: Q4 2025

### Task OB2: Implement Prometheus Metrics Exposition
- **Description**: Add comprehensive Prometheus metrics for monitoring and alerting
- **Current State**: Basic HTTP metrics only
- **Action Items**:
  - Implement custom Prometheus metrics for application performance
  - Add business logic metrics (crypto operations, key generation)
  - Configure metrics endpoints and scraping
  - Set up alerting rules and SLI/SLO definitions
- **Files**: Metrics implementation, Prometheus configuration
- **Expected Outcome**: Production-grade monitoring capabilities
- **Priority**: Medium - Production readiness

### Task OB4: Enhance Readiness Checks Performance and Coverage
- **Description**: Improve readiness checks with concurrent execution and additional health validations
- **Current State**: Readiness checks run sequentially and may be incomplete
- **Action Items**:
  - Add more readiness checks as needed (database connectivity, external service dependencies, resource availability)
  - Implement concurrent readiness checks for better performance
  - Add timeout handling for individual checks
  - Improve error reporting and health status granularity
  - Consider adding dependency health checks (telemetry, pools, etc.)
- **Files**: `internal/server/application/application_listener.go` (readiness check implementation)
- **Expected Outcome**: Faster, more comprehensive readiness validation with better observability
- **Priority**: MEDIUM - Application reliability and startup performance
- **Timeline**: Q1 2026

---

## Appendix: Grafana Dashboard Expansion

### Current Dashboard Limitations

The existing `cryptoutil.json` dashboard only includes basic HTTP metrics from the `otelfiber` middleware:

```json
{
  "panels": [
    {
      "title": "Request Rate",
      "targets": [{"expr": "rate(http_requests_total[5m])"}]
    },
    {
      "title": "Response Time",
      "targets": [{"expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"}]
    }
  ]
}
```

### Missing Custom Metrics Categories

#### 1. Pool Performance Metrics
**Source**: `internal/common/pool/pool.go`

**Available Metrics**:
- `cryptoutil.pool.get` - Histogram of get operation duration (milliseconds)
- `cryptoutil.pool.permission` - Histogram of permission wait duration (milliseconds)
- `cryptoutil.pool.generate` - Histogram of generate operation duration (milliseconds)

**Attributes**: `workers`, `size`, `values`, `duration`, `type` (per pool instance)

**Recommended Dashboard Panels**:
```json
{
  "title": "Pool Get Latency (95th percentile)",
  "targets": [{
    "expr": "histogram_quantile(0.95, rate(cryptoutil_pool_get_bucket[5m]))",
    "legendFormat": "{{pool}} - {{type}}"
  }]
},
{
  "title": "Pool Permission Wait Time",
  "targets": [{
    "expr": "histogram_quantile(0.95, rate(cryptoutil_pool_permission_bucket[5m]))",
    "legendFormat": "{{pool}} permission wait"
  }]
},
{
  "title": "Pool Generation Time",
  "targets": [{
    "expr": "histogram_quantile(0.95, rate(cryptoutil_pool_generate_bucket[5m]))",
    "legendFormat": "{{pool}} generation time"
  }]
}
```

#### 2. Security Header Validation Metrics
**Source**: `internal/server/application/application_listener.go`

**Available Metrics**:
- `security_headers_missing_total` - Counter of requests with missing security headers

**Recommended Dashboard Panels**:
```json
{
  "title": "Security Header Violations",
  "targets": [{
    "expr": "rate(security_headers_missing_total[5m])",
    "legendFormat": "Missing headers per second"
  }]
},
{
  "title": "Security Header Compliance Rate",
  "targets": [{
    "expr": "(1 - (rate(security_headers_missing_total[5m]) / rate(http_requests_total[5m]))) * 100",
    "legendFormat": "Compliance %"
  }]
}
```

### Implementation Architecture

**Metrics Flow**:
```
Application (OpenTelemetry) → OTEL Collector → Grafana-OTEL-LGTM (Prometheus + Grafana)
```

**Dashboard Updates Needed**:
1. **Add Pool Performance Dashboard**:
   - Pool utilization metrics
   - Latency percentiles per pool type
   - Worker efficiency monitoring

2. **Add Security Dashboard**:
   - Header compliance rates
   - Violation trends
   - Alert thresholds for security issues

3. **Add Business Logic Dashboard** (Future):
   - Cryptographic operation metrics
   - Key generation performance
   - Database operation latency

### OpenTelemetry to Prometheus Metric Name Mapping

OpenTelemetry metrics are automatically converted to Prometheus format:
- `cryptoutil.pool.get` → `cryptoutil_pool_get`
- `security_headers_missing_total` → `security_headers_missing_total`

### Alerting Recommendations

```yaml
# Example alert rules for security headers
groups:
  - name: security
    rules:
      - alert: HighSecurityHeaderViolations
        expr: rate(security_headers_missing_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High rate of missing security headers"
```

### Priority Implementation Order

1. **Phase 1**: Add pool performance metrics dashboard
2. **Phase 2**: Add security header compliance dashboard
3. **Phase 3**: Implement business logic metrics and dashboard
4. **Phase 4**: Add alerting rules and thresholds

**Timeline**: Q4 2025 implementation alongside OAuth 2.0 work.
