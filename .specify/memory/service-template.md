# Service Template Specifications

**Version**: 1.0.0
**Last Updated**: 2025-12-24
**Referenced By**: `.github/instructions/02-02.service-template.instructions.md`

## Overview

**CRITICAL: NEVER duplicate service infrastructure code - ALWAYS use extracted template**

All cryptoutil services MUST use a shared service template to ensure consistency, reduce duplication, and maintain quality across the product suite.

## Template Components

**Template extracted from KMS reference implementation**:

### 1. Two HTTPS Servers

- **Public HTTPS Binding**: External client access (business APIs, browser UIs)
- **Private API Binding**: Admin operations (health checks, shutdown, diagnostics)

**See**: `.github/instructions/02-03.https-ports.instructions.md` for complete HTTPS binding patterns

### 2. Two Public API Paths

- **`/browser/api/v1/*`**: Session-based authentication for browser clients
  - Middleware: CSRF protection, CORS policies, CSP headers
  - Authentication: Cookie-based sessions
  - Authorization: Resource-level checks

- **`/service/api/v1/*`**: Token-based authentication for headless clients
  - Middleware: IP allowlist, rate limiting
  - Authentication: Bearer tokens, mTLS
  - Authorization: Scope-based checks

**See**: `.github/instructions/02-03.https-ports.instructions.md` for middleware stack details

### 3. Three Private APIs

- **`/admin/v1/livez`**: Liveness health probe
  - Purpose: Process alive check (lightweight)
  - Use: Kubernetes liveness probe
  - Response: 200 OK or 503 Service Unavailable

- **`/admin/v1/readyz`**: Readiness health probe
  - Purpose: Dependencies healthy check (heavyweight)
  - Use: Kubernetes readiness probe, load balancer health checks
  - Validates: Database connectivity, federated service availability

- **`/admin/v1/shutdown`**: Graceful shutdown trigger
  - Purpose: Orchestration-initiated shutdown
  - Use: CI/CD pipelines, deployment automation
  - Behavior: Drain connections, close resources, exit gracefully

### 4. Database Abstraction

- **PostgreSQL || SQLite dual support**
- **GORM ORM** for database operations
- **Embedded SQL migrations** with golang-migrate
- **Connection pooling** with proper timeouts
- **Cross-DB compatibility** patterns (UUID as TEXT, JSON serialization)

**See**: `.github/instructions/03-04.database.instructions.md`, `03-05.sqlite-gorm.instructions.md`

### 5. OpenTelemetry Integration

- **OTLP export**: Traces, metrics, logs to otel-collector-contrib sidecar
- **Structured logging**: JSON format with correlation IDs
- **Prometheus metrics**: Exposed via `/admin/v1/metrics` endpoint
- **Trace propagation**: W3C Trace Context headers

**See**: `.github/instructions/02-05.observability.instructions.md`

### 6. Config Management

- **YAML files**: Primary configuration source
- **CLI flags**: Override config file values
- **Docker secrets support**: File dereference pattern (`file:///run/secrets/secret_name`)
- **Environment variables**: NOT used for secrets (Docker/K8s secrets only)

**See**: `.github/instructions/04-02.docker.instructions.md` for Docker secrets patterns

## Template Parameterization

### Constructor Injection Pattern

```go
type ServiceTemplate struct {
    Config         *ServiceConfig
    Handlers       HandlerRegistry
    Middleware     MiddlewareChain
    OpenAPISpec    *openapi3.T
}

func NewService(config *ServiceConfig, handlers HandlerRegistry) (*ServiceTemplate, error) {
    // Inject parameters for configuration, handlers, middleware
    // Business logic separated from infrastructure concerns
}
```

### Service-Specific Customization

- **OpenAPI specs**: Service-specific endpoints, models, paths
- **Business logic handlers**: Injected via handler registry
- **Middleware customization**: Additional service-specific middleware
- **Database schema**: Service-specific tables, migrations

## Mandatory Usage Rules

### MUST DO

- ✅ Extract reusable template before implementing new services
- ✅ ALL new services MUST use template (consistency, reduced duplication)
- ✅ ALL existing services MUST be refactored to use template (iterative migration)
- ✅ Template success criteria: learn-ps service validates template works

### NEVER DO

- ❌ Copy-paste service infrastructure code between services
- ❌ Duplicate dual-server pattern, health checks, shutdown logic
- ❌ Reimplement middleware pipeline, telemetry integration, crypto, sql/gorm setup

### Migration Validation

**Template success criteria (learn-ps service)**:

1. Implements ALL template requirements (dual HTTPS, health checks, config, telemetry)
2. Passes all unit/integration/e2e tests
3. Passes all CI/CD workflows
4. Deep analysis shows NO blockers to migrate existing services
5. Documentation demonstrates template reusability

## Service Template Migration Priority

**HIGH PRIORITY phased migration**:

### Phase 1: learn-ps FIRST - CRITICAL

- **Purpose**: Implement and validate ALL template requirements
- **Success Criteria**: Passes all tests, workflows, no migration blockers
- **Timeline**: Complete before production service migrations

### Phase 2: One Service at a Time (excludes sm-kms)

- **Migration Order**: jose-ja → pki-ca → identity services
- **Pattern**: Sequential refactoring, full test validation per service
- **Rollback**: Keep original code until new template implementation validated

### Phase 3: sm-kms LAST

- **Reason**: Most mature service, used as template extraction source
- **Timing**: Only after ALL other services running excellently on template
- **Benefit**: Validates template handles all edge cases before touching reference implementation

## Key Takeaways

1. **Single Template**: Extract from KMS, reuse for all services (9 services total)
2. **Dual HTTPS**: Public (business) + Admin (health checks) servers mandatory
3. **Dual Paths**: `/browser/**` (session-based) vs `/service/**` (token-based)
4. **Health Checks**: Liveness (process alive) vs Readiness (dependencies healthy)
5. **Migration Priority**: learn-ps first (validation), production services sequential, sm-kms last
6. **Zero Duplication**: Template parameterization prevents code duplication across services
