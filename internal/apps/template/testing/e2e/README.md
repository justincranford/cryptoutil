# Template E2E Testing Framework

## Overview

This package provides reusable end-to-end (E2E) testing helpers for all cryptoutil services. The framework is designed to work with Docker Compose deployments and supports sophisticated health checking for various container patterns.

## Components

### ComposeManager

The `ComposeManager` orchestrates Docker Compose lifecycle operations:

- `NewComposeManager(composeFile string)` - Creates a new compose manager with TLS-enabled HTTP client
- `Start(ctx context.Context)` - Brings up the docker compose stack
- `Stop(ctx context.Context)` - Tears down the docker compose stack
- `WaitForHealth(healthURL string, timeout time.Duration)` - Polls a health endpoint until healthy
- `WaitForMultipleServices(services map[string]string, timeout time.Duration)` - Waits for multiple services concurrently
- `WaitForServicesHealthy(ctx context.Context, services []ServiceAndJob)` - Batch health check with 3-use-case support

### Health Checking Strategy

Docker Compose health checking supports **three distinct use cases**:

#### 1. Job-only Healthchecks

Standalone jobs that must exit successfully (ExitCode=0).

**Examples**: `healthcheck-secrets`, `builder-cryptoutil`

**Usage**:
```go
services := []ServiceAndJob{
    {Service: "", Job: "healthcheck-secrets"},
}
err := composeManager.WaitForServicesHealthy(ctx, services)
```

**Docker Compose Pattern**:
```yaml
healthcheck-secrets:
  image: alpine:latest
  command: ["sh", "-c", "validate-secrets.sh"]
  # Job exits with code 0 on success, non-zero on failure
```

#### 2. Service-only Healthchecks

Services with native HEALTHCHECK instructions in their container image or Dockerfile.

**Examples**: `cryptoutil-sqlite`, `cryptoutil-postgres-1`, `postgres`, `grafana-otel-lgtm`

**Usage**:
```go
services := []ServiceAndJob{
    {Service: "cryptoutil-sqlite", Job: ""},
    {Service: "postgres", Job: ""},
}
err := composeManager.WaitForServicesHealthy(ctx, services)
```

**Docker Compose Pattern**:
```yaml
cryptoutil-sqlite:
  image: cryptoutil:latest
  healthcheck:
    test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "https://127.0.0.1:8080/admin/api/v1/livez"]
    interval: 10s
    timeout: 5s
    retries: 3
    start_period: 20s
```

#### 3. Service with Healthcheck Job

Services that don't have native healthchecks use an external sidecar job for health verification.

**Example**: `opentelemetry-collector-contrib` with `healthcheck-opentelemetry-collector-contrib`

**Usage**:
```go
services := []ServiceAndJob{
    {Service: "opentelemetry-collector-contrib", Job: "healthcheck-opentelemetry-collector-contrib"},
}
err := composeManager.WaitForServicesHealthy(ctx, services)
```

**Docker Compose Pattern**:
```yaml
opentelemetry-collector-contrib:
  image: otel/opentelemetry-collector-contrib:latest
  # No native HEALTHCHECK in the container image
  
healthcheck-opentelemetry-collector-contrib:
  image: alpine:latest
  command:
    - sh
    - -c
    - |
      apk add --no-cache wget
      for i in $(seq 1 30); do
        if wget --quiet --tries=1 --spider --timeout=2 http://opentelemetry-collector-contrib:13133/ 2>/dev/null; then
          echo "OpenTelemetry Collector is ready"
          exit 0
        fi
        sleep 2
      done
      exit 1
  depends_on:
    opentelemetry-collector-contrib:
      condition: service_started
```

### Why Not Use `docker compose up --wait`?

The `--wait` flag only works with containers that have native `HEALTHCHECK` instructions. Many third-party containers (like `otel/opentelemetry-collector-contrib`) don't include native healthchecks. This framework's 3-use-case approach handles:

1. Standalone validation jobs
2. Services with native healthchecks
3. Services requiring external healthcheck sidecars

## Usage Example

```go
package myservice_test

import (
    "context"
    "testing"
    "time"
    
    cryptoutilTemplateE2E "cryptoutil/internal/apps/template/testing/e2e"
)

func TestE2E_MyService(t *testing.T) {
    ctx := context.Background()
    
    // Create compose manager
    cm := cryptoutilTemplateE2E.NewComposeManager("../../deployments/myservice/compose.yml")
    
    // Start services
    if err := cm.Start(ctx); err != nil {
        t.Fatalf("Failed to start services: %v", err)
    }
    defer cm.Stop(ctx)
    
    // Define services to check (mixed use cases)
    services := []cryptoutilTemplateE2E.ServiceAndJob{
        {Service: "", Job: "healthcheck-secrets"},                                          // Use case 1
        {Service: "myservice-sqlite", Job: ""},                                            // Use case 2
        {Service: "opentelemetry-collector-contrib", Job: "healthcheck-otel-collector"},  // Use case 3
    }
    
    // Wait for all services to be healthy
    if err := cm.WaitForServicesHealthy(ctx, services); err != nil {
        t.Fatalf("Services failed to become healthy: %v", err)
    }
    
    // Run your E2E tests here
}
```

## Comparison with Legacy E2E Framework

The old `internal/test/e2e/` framework provided comprehensive test infrastructure including:

- `docker_health.go` - 3-use-case health checking (✅ **Now available in template**)
- `docker_utils.go` - Docker compose utilities
- `assertions.go` - Service verification logic
- `infrastructure.go` - Docker service management
- `fixtures.go` - Test infrastructure setup
- `http_utils.go` - HTTP utilities
- `log_utils.go` - Logging utilities
- `test_suite.go` - Core test orchestration

The template E2E framework currently provides:

- ✅ `compose.go` - Basic compose manager
- ✅ `docker_health.go` - 3-use-case health checking (newly added)
- ✅ `docker_health_test.go` - Comprehensive tests (100% coverage)

Additional utilities can be added to the template framework as needed.

## Design Principles

1. **Reusability**: All services should use the template E2E framework
2. **Type Safety**: Strong typing for service/job configurations
3. **Comprehensive Testing**: 100% coverage for critical health checking logic
4. **Clear Documentation**: Three use cases clearly documented
5. **Real-world Patterns**: Based on actual deployment patterns in `deployments/`

## References

- Original design: `internal/test/e2e/docker_health.go`
- Telemetry healthcheck pattern: `deployments/shared-telemetry/compose.yml`
- Implementation plan: `docs/implementation-plan-v1/plan.md`
