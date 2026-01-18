# cryptoutil Architecture - Single Source of Truth

**Version**: 1.0.0  
**Last Updated**: 2026-01-18  
**Status**: DRAFT - Requires review and refinement

**Purpose**: This document is the SINGLE SOURCE OF TRUTH for all cryptoutil architecture and design decisions. All implementation must conform to patterns defined here. NO duplication or conflicting information elsewhere.

---

## Table of Contents

1. [Product Suite Overview](#product-suite-overview)
2. [Service Template Pattern](#service-template-pattern)
3. [Product-Service Pattern](#product-service-pattern)
4. [Multi-Tenancy Architecture](#multi-tenancy-architecture)
5. [Authentication & Authorization](#authentication--authorization)
6. [Testing Strategy](#testing-strategy)
7. [Deployment Patterns](#deployment-patterns)
8. [Quality Gates](#quality-gates)

---

## Product Suite Overview

**4 Products, 9 Services total**:

| Product | Services | Ports | Purpose |
|---------|----------|-------|---------|
| **Secrets Manager** | sm-kms | 8080-8089 | Elastic key management, encryption-at-rest |
| **PKI** | pki-ca | 8443-8449 | X.509 certificates, EST, SCEP, OCSP, CRL |
| **JOSE** | jose-ja | 9443-9449 | JWK/JWS/JWE/JWT operations |
| **Identity** | identity-authz, identity-idp, identity-rs, identity-rp, identity-spa | 18000-18409 | OAuth 2.1, OIDC 1.0 |
| **Cipher** (Demo) | cipher-im | 8888-8889 | E2E encrypted messaging (template validation) |

**All services**: Public API (product-specific ports) + Admin API (127.0.0.1:9090)

**Deployment**: Container-first, Docker Compose for E2E, Kubernetes for production

---

## Service Template Pattern

**Location**: `internal/apps/template/service/`

### Core Principle

**ALL 9 services MUST use template for common infrastructure**. Zero duplication of server/database/telemetry/health check logic.

### Template Provides

**1. Dual HTTPS Servers**:
- **Public Server**: Business APIs (configurable bind address/port)
  - `/browser/api/v1/*` - Browser clients (session cookies, CSRF, CORS, CSP)
  - `/service/api/v1/*` - Headless clients (bearer tokens, mTLS, IP allowlist)
- **Admin Server**: Health checks (ALWAYS 127.0.0.1:9090)
  - `/admin/v1/livez` - Liveness (process alive? → restart if fail)
  - `/admin/v1/readyz` - Readiness (dependencies healthy? → remove from LB if fail)
  - `/admin/v1/shutdown` - Graceful shutdown (drain requests, close resources)

**2. Database Abstraction**:
- PostgreSQL (production) OR SQLite (dev/test) dual support
- GORM ORM for type-safe queries
- Embedded migrations (golang-migrate)
- Template migrations 1001-1004 (sessions, barrier, realms, multi-tenancy)
- Domain migrations 2001+ (service-specific tables)

**3. Multi-Tenancy Infrastructure**:
- `tenants` table (tenant metadata)
- `tenant_realms` table (authn realms per tenant)
- `sessions` table (user sessions with tenant_id)
- `tenant_join_requests` table (join requests for existing tenants)
- `users` table (with tenant_id FK)
- `pending_users` table (registration requests awaiting approval)

**4. Barrier Service** (Optional):
- Unseal key → Root key → Intermediate keys → Content keys
- HKDF-based deterministic key derivation
- Encryption-at-rest for sensitive data

**5. Telemetry**:
- OpenTelemetry (traces, metrics, logs via OTLP)
- Sidecar pattern: service → otel-collector → grafana-lgtm
- Structured logging with trace correlation

**6. Configuration**:
- YAML files (primary)
- Docker secrets for sensitive data (`file:///run/secrets/`)
- CLI flags (overrides)
- NEVER environment variables for secrets

**7. Session Management**:
- JWE, JWS, or Opaque session tokens (configurable)
- PostgreSQL/SQLite session storage (NO Redis/Memcached)
- 30-minute re-authentication for high-sensitivity operations

**8. Hash Service**:
- PBKDF2-HMAC-SHA256 for passwords (FIPS 140-3 compliant)
- Global pepper (via Docker secret)
- Version-based hash policies
- Lazy migration on pepper rotation

### Template Usage Pattern

```go
// ServerBuilder creates all infrastructure
builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg)

// Register domain-specific migrations (2001+)
builder.WithDomainMigrations(domainMigrationsFS, "migrations")

// Register domain-specific routes
builder.WithPublicRouteRegistration(func(base, res) error {
    // Create domain repos
    messageRepo := cryptoutilCipherRepository.NewMessageRepository(res.DB)
    
    // Create domain server
    publicServer := cryptoutilCipherServer.NewPublicServer(base, messageRepo, ...)
    
    // Register routes
    publicServer.RegisterRoutes()
    return nil
})

// Build returns all infrastructure
resources, err := builder.Build()
```

### Migration Priority

**CRITICAL ORDER**:
1. **cipher-im FIRST** (validate template works, identify gaps)
2. jose-ja, pki-ca, identity-* (sequential, fix template issues as discovered)
3. **sm-kms LAST** (most complex, template must be mature)

**WHY**: cipher-im is demo service, safe to iterate. Production services migrate after template validated.

---

## Product-Service Pattern

### CLI Structure

**Pattern**: `cmd/PRODUCT-SERVICE/main.go` → `internal/apps/PRODUCT/SERVICE/`

**Examples**:
- `cmd/cipher-im/main.go` → `internal/apps/cipher/im/im.go`
- `cmd/jose-ja/main.go` → `internal/apps/jose/ja/ja.go`
- `cmd/pki-ca/main.go` → `internal/apps/pki/ca/ca.go`

**Subcommands** (ALL services support):
- `server` - Start dual HTTPS servers
- `client` - CLI client for API interaction
- `health` - Check all health endpoints
- `livez` - Check liveness only
- `readyz` - Check readiness only
- `shutdown` - Trigger graceful shutdown
- `init` - Initialize database/config
- `compose` - Generate docker-compose.yml
- `demo` - Run E2E demonstration
- `e2e` - Run E2E tests

### Directory Organization

```
internal/apps/
├── template/           # Reusable template (all services use this)
│   └── service/
│       ├── builder/    # ServerBuilder
│       ├── config/     # Settings
│       ├── server/     # Application, PublicServerBase, AdminServer
│       └── repository/ # TenantRepo, RealmRepo, SessionRepo
├── cipher/
│   └── im/            # Cipher-IM service
│       ├── config/
│       ├── server/    # Domain-specific public server
│       └── repository/# Domain-specific repos (messages, recipients)
├── jose/
│   └── ja/            # JOSE-JA service
└── pki/
    └── ca/            # PKI-CA service
```

**Rule**: `internal/apps/template/` is CANONICAL. All services import and extend template.

---

## Multi-Tenancy Architecture

### Registration Flow

**NEW USER - Create New Tenant**:

```http
POST /browser/api/v1/register
{
    "username": "admin@example.com",
    "password": "securepass",
    "create_tenant": true,
    "tenant_name": "Acme Corp"
}
```

**Response**: HTTP 403 Forbidden (pending approval)
- User saved in `pending_users` table (NOT `users`)
- Admin must approve via admin panel
- After approval: user moved to `users`, tenant/realm created
- After rejection: user deleted from `pending_users`
- All API calls return HTTP 403 until approved, HTTP 401 if rejected

**EXISTING USER - Join Existing Tenant**:

```http
POST /browser/api/v1/register  
{
    "username": "user@example.com",
    "password": "securepass",
    "create_tenant": false,
    "tenant_id": "uuid-of-existing-tenant"
}
```

**Response**: HTTP 403 Forbidden (pending approval)
- User saved in `pending_users` table
- Join request saved in `tenant_join_requests` table with status "pending"
- Tenant admin must approve join request
- After approval: user moved to `users`, granted access to tenant
- User receives NO session token until approved

### Realms (Authentication Only)

**Purpose**: Determines HOW user authenticates (password, OAuth, LDAP, WebAuthn)

**NOT for data isolation**: All realms in same tenant see SAME data

**Realm Types**:
- DB-based username/password (default)
- File-based username/password
- OAuth 2.0/OIDC federated
- LDAP/AD
- WebAuthn/FIDO2
- SAML 2.0

**Repository Pattern**: Filter by `tenant_id` ONLY, NOT `realm_id`

```go
// ✅ CORRECT
db.Where("tenant_id = ?", tenantID).Find(&messages)

// ❌ WRONG
db.Where("tenant_id = ? AND realm_id = ?", tenantID, realmID).Find(&messages)
```

### Session Management

**Session contains**: user_id, tenant_id, realm_id (for audit), issued_at, expires_at

**Session formats** (configurable):
- JWE (encrypted JWT) - stateless, secure
- JWS (signed JWT) - stateless, tamper-proof
- Opaque (UUID) - stateful, server-side lookup

**Storage**: PostgreSQL or SQLite (NO Redis/Memcached)

**Expiry**: Configurable (default: 1 hour browser, 24 hours service)

**Step-up**: Re-authenticate every 30 minutes for high-sensitivity operations

---

## Authentication & Authorization

### Authentication Methods

**Headless Clients** (`/service/**` paths):
1. Bearer (API Token)
2. Basic (Client ID/Secret)
3. mTLS (HTTPS Client Certificate)
4. JWE Access Token (OAuth 2.1)
5. JWS Access Token (OAuth 2.1)
6. Opaque Access Token (OAuth 2.1)

**Browser Clients** (`/browser/**` paths):
1. Session Cookie (JWE/JWS/Opaque)
2. Basic (Username/Password)
3. OAuth 2.0 (Google, Microsoft, GitHub, etc.)
4. SAML 2.0 (Enterprise SSO)
5. TOTP (Authenticator App)
6. WebAuthn (Passkeys, Security Keys)
7. Magic Link (Email/SMS)
8. OTP (Email/SMS/Voice)

### Authorization Pattern

**Zero Trust**: ALWAYS re-evaluate permissions (NO caching of authz decisions)

**Scope-based**: OAuth 2.0 scopes (read:messages, write:messages, delete:messages)

**RBAC**: Role-based access control (admin, user, viewer)

**Resource-level ACLs**: Fine-grained permissions per resource

---

## Testing Strategy

### Unit Tests

**Coverage**: ≥95% production, ≥98% infrastructure/utility

**Pattern**: Table-driven with `t.Parallel()` for orthogonal data

**Test Data**: UUIDv7 for all values (thread-safe, process-safe)

**Dynamic Ports**: ALWAYS port 0 (prevents TIME_WAIT on Windows)

**SQLite Config**:
```go
// MaxOpenConns=5 (GORM transactions need separate connection)
sqlDB.SetMaxOpenConns(5)
sqlDB.Exec("PRAGMA journal_mode=WAL;")
sqlDB.Exec("PRAGMA busy_timeout = 30000;")
```

### Integration Tests

**TestMain Pattern** (per package):
```go
func TestMain(m *testing.M) {
    ctx := context.Background()
    
    // Start PostgreSQL test-container ONCE
    container, _ := postgres.RunContainer(ctx, ...)
    defer container.Terminate(ctx)
    
    // Connect and migrate
    testDB, _ := gorm.Open(postgres.Open(connStr))
    migrate(testDB)
    
    // Run all tests (reuse container)
    os.Exit(m.Run())
}
```

**Isolation**: Each test creates unique users (UUIDv7 in usernames)

### E2E Tests

**MUST test BOTH paths**:
- `/service/**` (headless clients)
- `/browser/**` (browser clients)

**Pattern**: Docker Compose with test-containers

**Registration**: Via HTTP endpoints (realistic user flow)

**NO demo tenants**: E2E tests create tenants via `/register` endpoint

### Mutation Testing

**gremlins**: ≥85% production, ≥98% infrastructure/utility (per package)

**Parallel execution**: GitHub Actions matrix (4-6 packages per job)

---

## Deployment Patterns

### Docker Compose

**Structure**:
```yaml
services:
  cipher-im:
    image: cryptoutil:local
    command: ["server", "--config=/app/configs/cipher/cipher-im.yml"]
    secrets: [database_url, unseal_key, tls_cert, tls_key, hash_pepper]
    ports: ["8888:8888"]
    depends_on:
      postgres: {condition: service_healthy}
      otel-collector: {condition: service_started}
      
  postgres:
    image: postgres:18-alpine
    environment: {POSTGRES_DB: cryptoutil, POSTGRES_USER: cryptoutil}
    healthcheck: {test: ["CMD", "pg_isready"], interval: 10s}
    
  otel-collector:
    image: otel/opentelemetry-collector-contrib:latest
    ports: ["4317:4317", "4318:4318"]
    
  grafana-lgtm:
    image: grafana/otel-lgtm:latest
    ports: ["3000:3000"]
```

**Secrets**: ALWAYS `file:///run/secrets/` pattern (NEVER env vars)

**Health Checks**: Via admin endpoints (livez/readyz)

**Networking**: Container-to-container (use service names, NOT host ports)

### Kubernetes

**Patterns**: StatefulSets for stateful services, Deployments for stateless

**Probes**:
```yaml
livenessProbe:
  httpGet: {path: /admin/v1/livez, port: 9090, scheme: HTTPS}
  initialDelaySeconds: 10
  periodSeconds: 10
  
readinessProbe:
  httpGet: {path: /admin/v1/readyz, port: 9090, scheme: HTTPS}
  initialDelaySeconds: 15
  periodSeconds: 5
```

**Secrets**: Kubernetes Secrets mounted as files

**Service Mesh**: Istio/Linkerd for mTLS (service-to-service)

---

## Quality Gates

### Pre-Commit (Automated)

1. `golangci-lint run --fix` → zero warnings
2. `go test ./...` → 100% pass
3. `go build ./...` → clean build
4. `grep -r "TODO\|FIXME"` → zero new TODOs

### CI/CD (GitHub Actions)

**Quality Workflow**:
- golangci-lint v2.7.2+
- gofmt, gofumpt, goimports
- Zero linting errors (NO exceptions)

**Test Workflow**:
- Unit tests (per package, parallel)
- Coverage reports (≥95%/≥98%)
- Race detector (`-race -count=2`)

**Mutation Workflow**:
- gremlins (parallel by package)
- ≥85%/≥98% efficacy

**E2E Workflow**:
- Docker Compose deployment
- BOTH `/service/**` and `/browser/**` paths
- Health check validation

### Pre-Merge Requirements

**ALL MUST PASS**:
- [ ] Quality gates (linting, formatting)
- [ ] Unit tests (≥95%/≥98% coverage)
- [ ] Mutation tests (≥85%/≥98% efficacy)
- [ ] E2E tests (BOTH paths)
- [ ] Conventional commit format
- [ ] Documentation updated
- [ ] Zero new TODOs without tracking

---

## Revision Instructions

**To update this document**:

1. Identify section needing change
2. Propose change with rationale
3. Update ONLY affected section (keep rest unchanged)
4. Commit with: `docs(arch): update [section] - [reason]`
5. NO duplication of this content elsewhere

**Questions for refinement**:

- Is template migration priority clear?
- Are registration flow patterns complete?
- Do testing strategies cover all scenarios?
- Are deployment patterns production-ready?
- What's missing that you repeat explaining?

**Next iteration**: Address identified gaps, add concrete code examples, refine based on feedback.
