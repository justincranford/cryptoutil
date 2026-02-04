# cryptoutil Architecture

**Purpose**: Single source of truth for cryptoutil product suite design, directory structure, and deployment patterns.

**Companion Document**: [SERVICE-TEMPLATE.md](SERVICE-TEMPLATE.md) - Complete blueprint for building services.

---

## Product Suite Overview

**4 Products, 9 Services**:

| Product | Service Alias | Ports | Description |
|---------|---------------|-------|-------------|
| **Secrets Manager** | sm-kms | 8080-8089 | Elastic key management, encryption-at-rest |
| **PKI** | pki-ca | 8050-8449 | X.509 certificates, EST, SCEP, OCSP, CRL |
| **JOSE** | jose-ja | 8060 | JWK/JWS/JWE/JWT operations |
| **Identity** | identity-authz | 8080-8089 | OAuth 2.1 authorization server |
| **Identity** | identity-idp | 8100-8109 | OIDC Identity Provider |
| **Identity** | identity-rs | 8200-8209 | Resource Server (reference) |
| **Identity** | identity-rp | 8300-8309 | Relying Party (reference) |
| **Identity** | identity-spa | 8400-8409 | Single Page Application (reference) |
| **Cipher** | cipher-im | 8070-8071 | E2E encrypted messaging |

**Admin Port**: ALL services use 127.0.0.1:9090 for health checks and graceful shutdown.

---

## Service Ports - Complete Reference

This section provides the authoritative port assignments for all 9 product-services.

### Port Assignment Table

| Service | Container Port | Host Port Range | Admin Port | Protocol | Status |
|---------|----------------|-----------------|------------|----------|--------|
| **sm-kms** | 8080 | 8080-8082 (SQLite:8080, PG1:8081, PG2:8082) | 9090 | HTTPS | Implemented |
| **pki-ca** | 8050 | 8050-8445 (SQLite:8050, PG1:8444, PG2:8445) | 9090* | HTTPS | Implemented |
| **jose-ja** | 8060 | 8060 | 9092 | HTTPS | Implemented |
| **identity-authz** | 8080 | 8080-8089 (scaling) | 9090 | HTTPS | Planned |
| **identity-idp** | 8081 | 8100-8109 (scaling) | 9090 | HTTPS | Planned |
| **identity-rs** | 8082 | 8200-8209 (scaling) | 9090 | HTTPS | Planned |
| **identity-rp** | 8083 | 8300-8309 (scaling) | 9090 | HTTPS | Planned |
| **identity-spa** | 8084 | 8400-8409 (scaling) | 9090 | HTTPS | Planned |
| **cipher-im** | 8070 | 8880-8882 (SQLite:8880, PG1:8881, PG2:8882) | 9090 | HTTPS | Implemented |

*Note: pki-ca uses non-standard health check paths (`/livez`, `/readyz`) without `/admin/api/v1/` prefix.

### Port Design Principles

1. **Container Port Consistency**: Each service uses a fixed container port
2. **Host Port Scaling**: Host ports allow multiple instances (port ranges)
3. **Admin Port Isolation**: Admin APIs bind to 127.0.0.1:9090 (localhost only)
4. **No Host Exposure for Admin**: Admin ports NEVER exposed to Docker host

### Current Implementation vs Instructions Discrepancy

The `.github/instructions/02-01.architecture.instructions.md` file documents:
- jose-ja: 8060-8069 (documented) matches actual implementation
- identity-*: 8100-8139 (documented) matches actual implementation

**Status**: Port standardization complete. All identity services now use 8100 series ports:
- identity-authz/idp: 8100-8109
- identity-rs: 8110-8119
- identity-rp: 8120-8129
- identity-spa: 8130-8139

### PostgreSQL Ports

| Service | Host Port | Container Port | Notes |
|---------|-----------|----------------|-------|
| kms-postgres | 5432 | 5432 | Default PostgreSQL |
| ca-postgres | 5432 | 5432 | Default PostgreSQL |
| identity-postgres | 5433 | 5432 | Offset to avoid conflict |
| template-postgres | 5433 | 5432 | Offset to avoid conflict |

### Telemetry Ports (Shared)

| Service | Host Port | Container Port | Protocol |
|---------|-----------|----------------|----------|
| opentelemetry-collector-contrib | (internal) | 4317 | OTLP gRPC |
| opentelemetry-collector-contrib | (internal) | 4318 | OTLP HTTP |
| grafana-otel-lgtm | 3000 | 3000 | HTTP (UI) |

---

## Directory Structure

### cmd/ - CLI Entry Points

```
cmd/
├── cryptoutil/main.go         # Suite-level CLI (all products)
├── cipher/main.go             # Cipher product CLI (all cipher services)
├── cipher-im/main.go          # Cipher-IM service CLI
├── jose/main.go               # JOSE product CLI
├── jose-server/main.go        # JOSE-JA service CLI
├── pki/main.go                # PKI product CLI
├── ca-server/main.go          # PKI-CA service CLI
├── identity/main.go           # Identity product CLI
├── identity-unified/main.go   # Identity unified service CLI
└── sm-kms/main.go             # SM-KMS service CLI (legacy)
```

**Pattern**: Thin `main()` delegates to `internal/apps/<product>/<service>/`.

```go
func main() {
    os.Exit(cryptoutilAppsCipherIm.IM(os.Args[1:], os.Stdout, os.Stderr))
}
```

### internal/apps/ - Service Implementations

```
internal/apps/
├── template/                # REUSABLE template (all services import this)
│   ├── service/
│   │   ├── config/          # ServiceTemplateServerSettings
│   │   ├── server/          # Application, PublicServerBase, AdminServer
│   │   │   ├── application/ # ApplicationCore, ApplicationBasic
│   │   │   ├── builder/     # ServerBuilder fluent API
│   │   │   ├── listener/    # AdminHTTPServer
│   │   │   ├── barrier/     # Encryption-at-rest service
│   │   │   ├── businesslogic/ # SessionManager, TenantRegistration
│   │   │   ├── repository/  # TenantRepo, RealmRepo, SessionRepo
│   │   │   └── realms/      # Authentication realm implementations
│   │   └── testutil/        # Test helpers (NewTestSettings)
│   └── testing/
│       └── e2e/             # ComposeManager for E2E orchestration
├── cipher/
│   └── im/                  # Cipher-IM service
│       ├── domain/          # Domain models (Message, Recipient)
│       ├── repository/      # Domain repos + migrations (2001+)
│       ├── server/          # CipherIMServer, PublicServer
│       │   ├── config/      # CipherImServerSettings embeds template
│       │   └── apis/        # HTTP handlers
│       ├── client/          # API client
│       ├── e2e/             # E2E tests (Docker Compose)
│       └── integration/     # Integration tests
├── jose/
│   └── ja/                  # JOSE-JA service (same structure)
├── pki/
│   └── ca/                  # PKI-CA service (same structure)
└── identity/
    ├── authz/               # OAuth 2.1 Authorization Server
    ├── idp/                 # OIDC Identity Provider
    ├── rs/                  # Resource Server
    ├── rp/                  # Relying Party
    └── spa/                 # Single Page Application
```

### internal/shared/ - Shared Utilities

```
internal/shared/
├── barrier/                 # Unseal keys service (HKDF derivation)
├── config/                  # Configuration helpers
├── crypto/
│   ├── hash/                # Version-based hash registries (PBKDF2, HKDF)
│   ├── jose/                # JWK generation, JWE/JWS operations
│   ├── tls/                 # TLS configuration helpers
│   └── certificate/         # X.509 certificate generation
├── magic/                   # Named constants (ports, timeouts, paths)
├── telemetry/               # OpenTelemetry integration
└── testutil/                # Shared test utilities
```

### internal/kms/ - Legacy KMS (Template Source)

```
internal/kms/                # Original KMS implementation
├── server/
│   ├── application/         # ApplicationListener (template source)
│   ├── businesslogic/       # Business logic services
│   ├── handler/             # HTTP handlers
│   ├── middleware/          # Request middleware
│   └── repository/          # Data access
├── client/                  # KMS API client
└── cmd/                     # KMS CLI
```

**Note**: sm-kms is the template extraction source. Migrate LAST to ensure template maturity.

### deployments/ - Docker Compose

```
deployments/
├── telemetry/               # SHARED: otel-collector + grafana-lgtm
│   └── compose.yml
├── cipher/
│   ├── compose.yml          # Include telemetry, service definition
│   ├── Dockerfile.cipher
│   ├── config/              # YAML configs
│   └── secrets/             # Docker secrets (*.secret, 440 perms)
├── <PRODUCT>/
│   └── ... (same structure)
├── jose/
│   └── ... (same structure)
├── ca/
│   └── ... (same structure)
├── identity/
│   └── ... (same structure)
└── sm/
    └── ... (same structure)
```

### configs/ - Application Configuration

```
configs/
├── test/                    # Test configurations
└── <product>/               # Production configurations
    └── <service>.yml
```

---

## CLI Patterns

### Subcommands (All Services Support)

| Subcommand | Description |
|------------|-------------|
| `server` | Start dual HTTPS servers (default) |
| `client` | CLI client for API interaction |
| `health` | Check all health endpoints |
| `livez` | Check liveness only |
| `readyz` | Check readiness only |
| `shutdown` | Trigger graceful shutdown |
| `init` | Initialize database/config |
| `demo` | Run E2E demonstration |

### CLI Hierarchy

```
# Product-Service pattern (preferred)
cipher-im server --config=/etc/cipher/im.yml

# Service pattern
im server --config=/etc/cipher/im.yml

# Product pattern (routes to service)
cipher im server --config=/etc/cipher/im.yml

# Suite pattern (routes to product, then service)
cryptoutil cipher im server --config=/etc/cipher/im.yml
```

```
# Product-Service pattern (preferred)
jose-ja server --config=/etc/jose/ja.yml

# Service pattern
ja server --config=/etc/jose/ja.yml

# Product pattern (routes to service)
jose ja server --config=/etc/jose/ja.yml

# Suite pattern (routes to product, then service)
cryptoutil jose ja server --config=/etc/jose/ja.yml
```

```
# Product-Service pattern (preferred)
sm-kms server --config=/etc/sm/kms.yml

# Service pattern
kms server --config=/etc/sm/kms.yml

# Product pattern (routes to service)
sm kms server --config=/etc/sm/kms.yml

# Suite pattern (routes to product, then service)
cryptoutil sm kms server --config=/etc/sm/kms.yml
```

---

## Multi-Tenancy Architecture

### Data Isolation

**tenant_id** scopes ALL data access:

```go
// ✅ CORRECT: Filter by tenant_id
db.Where("tenant_id = ?", tenantID).Find(&messages)

// ❌ WRONG: Filter by realm_id (authentication only)
db.Where("tenant_id = ? AND realm_id = ?", tenantID, realmID).Find(&messages)
```

### Realms (Authentication Configuration Only)

**CRITICAL**: Realms define authentication METHOD and POLICY, NOT data scoping.

**realm_id** determines:
1. **HOW** users authenticate (authentication method)
2. **Password policies** (min length, complexity requirements)
3. **Session policies** (timeout, refresh, absolute max)
4. **MFA requirements** (required, allowed methods)
5. **Rate limiting** (per-realm overrides)

**Realms do NOT scope data** - all realms in same tenant see same data. Only `tenant_id` scopes data access.

#### Realm Types (16 Supported)

**Federated Realm Types** (external identity providers):

| Type | Description | Use Case |
|------|-------------|----------|
| `username_password` | Database-stored credentials | Default, internal users |
| `ldap` | LDAP/Active Directory | Enterprise directory |
| `oauth2` | OAuth 2.0/OIDC provider | Social login, SSO |
| `saml` | SAML 2.0 federation | Enterprise SSO |

**Non-Federated Browser Realm Types** (`/browser/**` paths, session-based):

| Type | Description | Use Case |
|------|-------------|----------|
| `jwe-session-cookie` | Encrypted JWT in cookie | Stateless secure sessions |
| `jws-session-cookie` | Signed JWT in cookie | Stateless sessions with visibility |
| `opaque-session-cookie` | Server-side session storage | Traditional sessions |
| `basic-username-password` | HTTP Basic + username/password | Simple browser auth |
| `bearer-api-token` | Bearer token from browser | API access from SPA |
| `https-client-cert` | mTLS client certificate | High-security browser access |

**Non-Federated Service Realm Types** (`/service/**` paths, token-based):

| Type | Description | Use Case |
|------|-------------|----------|
| `jwe-session-token` | Encrypted JWT token | Encrypted service tokens |
| `jws-session-token` | Signed JWT token | Signed service tokens |
| `opaque-session-token` | Server-side token lookup | Traditional service tokens |
| `basic-client-id-secret` | HTTP Basic + client credentials | Service-to-service |
| `bearer-api-token` | Bearer token (shared) | API access |
| `https-client-cert` | mTLS (shared) | High-security service auth |

#### Realm Configuration

Each realm has configurable policies:

```go
type RealmConfig struct {
    // Password validation
    PasswordMinLength        int   // Default: 12
    PasswordRequireUppercase bool  // Default: true
    PasswordRequireLowercase bool  // Default: true
    PasswordRequireDigits    bool  // Default: true
    PasswordRequireSpecial   bool  // Default: true
    PasswordMinUniqueChars   int   // Default: 8
    PasswordMaxRepeatedChars int   // Default: 3

    // Session configuration
    SessionTimeout        int   // Seconds, default: 3600 (1 hour)
    SessionAbsoluteMax    int   // Seconds, default: 86400 (24 hours)
    SessionRefreshEnabled bool  // Default: true

    // Multi-factor authentication
    MFARequired bool     // Default: false
    MFAMethods  []string // e.g., ["totp", "webauthn", "sms"]

    // Rate limiting overrides
    LoginRateLimit   int // Attempts per minute, default: 5
    MessageRateLimit int // Messages per minute, default: 10
}
```

#### Realm vs Tenant Relationship

```
Tenant (data isolation boundary)
├── Realm A (username_password, default policy)
│   └── Users authenticate via database credentials
├── Realm B (ldap, enterprise policy)
│   └── Users authenticate via Active Directory
└── Realm C (oauth2, federated policy)
    └── Users authenticate via external OIDC provider

All realms access SAME data (scoped by tenant_id only)
```

**Key Insight**: Users from different realms in the same tenant see the same data. The realm only controls HOW they authenticate, not WHAT they can access.

### Registration Flow

**New Tenant** (omit tenant_id):

```http
POST /browser/api/v1/register
{ "username": "admin@example.com", "password": "securepass" }
```

- User saved to `pending_users` (NOT `users`)
- Returns HTTP 403 until admin approves
- After approval: tenant created, user moved to `users`

**Join Existing Tenant** (specify tenant_id):

```http
POST /browser/api/v1/register
{ "username": "user@example.com", "password": "securepass", "tenant_id": "uuid" }
```

- User saved to `pending_users` with tenant_id
- Tenant admin must approve via admin panel

---

## HTTPS Endpoint Architecture

### Dual Server Pattern (ALL Services)

| Server | Bind Address | Purpose |
|--------|--------------|---------|
| **Public** | Configurable (0.0.0.0 in containers) | Business APIs, browser UIs |
| **Admin** | 127.0.0.1:9090 ALWAYS | Health checks, graceful shutdown |

### Public API Paths

| Path Prefix | Client Type | Middleware |
|-------------|-------------|------------|
| `/service/api/v1/*` | Headless (service-to-service) | Bearer tokens, mTLS, IP allowlist |
| `/browser/api/v1/*` | Browser (user-facing) | Session cookies, CSRF, CORS, CSP |

### Admin API Paths

| Endpoint | Purpose | Failure Action |
|----------|---------|----------------|
| `GET /admin/api/v1/livez` | Liveness (process alive?) | Restart container |
| `GET /admin/api/v1/readyz` | Readiness (dependencies healthy?) | Remove from LB |
| `POST /admin/api/v1/shutdown` | Graceful shutdown | N/A |

---

## Service Federation

### Configuration Pattern (Static)

```yaml
federation:
  identity_url: "https://identity-authz:8180"  # Docker Compose service name
  identity_enabled: true
  jose_url: "https://jose-ja:8280"
```

**No dynamic discovery**: Service URLs in config file, restart required on changes.

### Fallback Pattern

If federated service unavailable:

```yaml
realms:
  - type: federated
    provider: identity-authz
    url: "https://identity-authz:8180"
  - type: database
    name: local-fallback
    enabled: true  # Emergency operator access
```

---

## Database Architecture

### Dual Database Support

| Database | Use Case | Connection |
|----------|----------|------------|
| PostgreSQL 18+ | Production, E2E tests | External or test-container |
| SQLite (in-memory) | Unit/integration tests, dev | `:memory:` or file |

### Cross-DB Compatibility Rules

```go
// UUID fields: ALWAYS type:text (SQLite has no native UUID)
ID googleUuid.UUID `gorm:"type:text;primaryKey"`

// Nullable UUIDs: Use NullableUUID (NOT *googleUuid.UUID)
ClientProfileID NullableUUID `gorm:"type:text;index"`

// JSON arrays: ALWAYS serializer:json (NOT type:json)
AllowedScopes []string `gorm:"serializer:json"`
```

### SQLite Configuration (Tests)

```go
sqlDB.Exec("PRAGMA journal_mode=WAL;")       // Concurrent reads + 1 writer
sqlDB.Exec("PRAGMA busy_timeout = 30000;")   // 30s retry on lock
sqlDB.SetMaxOpenConns(5)                     // GORM transactions need multiple
```

### Migration Versioning

| Range | Owner | Examples |
|-------|-------|----------|
| 1001-1999 | Template | Sessions (1001), Barrier (1002), Realms (1003), Tenants (1004), PendingUsers (1005) |
| 2001+ | Domain | cipher-im messages (2001), jose JWKs (2001) |

---

## Security Architecture

### FIPS 140-3 Compliance (ALWAYS Enabled)

**Approved Algorithms**:

| Category | Algorithms |
|----------|------------|
| Asymmetric | RSA ≥2048, ECDSA P-256/384/521, EdDSA 25519/448 |
| Symmetric | AES-128/192/256 (GCM, CBC+HMAC) |
| Digest | SHA-256/384/512, HMAC-SHA256/384/512 |
| KDF | PBKDF2-HMAC-SHA256/384/512, HKDF-SHA256/384/512 |

**Banned**: bcrypt, scrypt, Argon2, MD5, SHA-1, RSA <2048, DES, 3DES.

### Key Hierarchy (Barrier Service)

```
Unseal Key (Docker secrets, NEVER stored)
    └── Root Key (encrypted at rest, rotated annually)
        └── Intermediate Key (rotated quarterly)
            └── Content Key (rotated per-operation or hourly)
```

**HKDF Derivation**: All instances with same unseal secrets derive identical keys.

### Hash Service (Version-Based)

```
Hash Format: {version}:{algorithm}:{iterations}:base64(salt):base64(hash)
Example:     {5}:PBKDF2-HMAC-SHA256:rounds=600000:abc123...:def456...
```

**Current Version**: V5 (OWASP 2023, 600k iterations, HMAC-SHA256).

**Pepper**: MANDATORY from Docker secrets, NEVER in DB/source.

---

## Observability

### Telemetry Stack

```
Service → otel-collector:4317 → grafana-otel-lgtm:14317
```

**NEVER bypass otel-collector** (sidecar pattern mandatory).

### Structured Logging

```go
logger.Info("User registered",
    zap.String("user_id", userID.String()),
    zap.String("tenant_id", tenantID.String()),
    zap.Duration("duration", elapsed),
)
```

### Metrics

| Category | Metrics |
|----------|---------|
| HTTP | `http_requests_total`, `http_request_duration_seconds`, `http_requests_in_flight` |
| Database | `db_connections_open`, `db_query_duration_seconds`, `db_errors_total` |
| Crypto | `crypto_operations_total`, `crypto_operation_duration_seconds` |

---

## Docker Compose Patterns

### Shared Telemetry (Include Pattern)

```yaml
# deployments/cipher/compose.yml
include:
  - path: ../telemetry/compose.yml

services:
  cipher-im:
    # ...
```

### Docker Secrets (MANDATORY)

```yaml
secrets:
  postgres_password.secret:
    file: ./secrets/postgres_password.secret  # chmod 440

services:
  postgres:
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password.secret
    secrets:
      - postgres_password.secret
```

**NEVER use inline environment variables for credentials.**

### Health Checks

```yaml
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null",
         "https://127.0.0.1:9090/admin/api/v1/livez"]
  start_period: 60s
  interval: 5s
```

**Use wget (Alpine), 127.0.0.1 (not localhost).**

---

## Quality Gates

### Pre-Commit

1. `golangci-lint run --fix` → zero warnings
2. `go test ./...` → 100% pass
3. `go build ./...` → clean build

### CI/CD

| Workflow | Requirement |
|----------|-------------|
| Coverage | ≥95% production, ≥98% infrastructure/utility |
| Mutation (gremlins) | ≥85% production, ≥98% infrastructure |
| E2E | BOTH `/service/**` AND `/browser/**` paths |

### Test Patterns

See [SERVICE-TEMPLATE.md](SERVICE-TEMPLATE.md) for mandatory test patterns.

---

## Migration Priority

| Order | Service | Status | Rationale |
|-------|---------|--------|-----------|
| 1 | cipher-im | ✅ Complete | Template validation |
| 2 | jose-ja | ✅ Complete | Second stereotype |
| 3 | pki-ca | Pending | Certificate authority |
| 4 | identity-* | Pending | OAuth/OIDC stack |
| 5 | sm-kms | Last | Template source, migrate after template mature |

---

## Factory Patterns

### *FromSettings Pattern (PREFERRED)

Services should use `*FromSettings` factory functions for consistent, testable initialization:

```go
// ✅ PREFERRED: Settings-based factory
func NewUnsealKeysServiceFromSettings(settings *UnsealKeysSettings) (*UnsealKeysService, error) {
    // Configuration-driven initialization
}

// Also acceptable: Direct constructor when settings not applicable
func NewMessageRepository(db *gorm.DB) *MessageRepository {
    return &MessageRepository{db: db}
}
```

**Benefits**:
- Configuration-driven behavior
- Testable (inject test settings)
- Consistent across all services
- Self-documenting dependencies

---

## Configuration Priority

**Load Order** (highest to lowest):

1. **Docker Secrets** (`file:///run/secrets/secret_name`) - Sensitive values
2. **YAML Configuration** (`--config=/path/to/config.yml`) - Primary configuration
3. **CLI Parameters** (`--bind-public-port=8080`) - Overrides

**CRITICAL: Environment variables NOT supported for configuration** (security, auditability).

---

## API Path Patterns

**NO service name in request paths**:

```
✅ /service/api/v1/elastic-jwks     (correct)
✅ /browser/api/v1/sessions         (correct)
❌ /service/api/v1/jose/elastic-jwks (WRONG - no service name in path)
```

**Standard Prefixes**:

| Prefix | Purpose | Authentication |
|--------|---------|----------------|
| `/service/api/v1/*` | Headless APIs | Bearer tokens, mTLS |
| `/browser/api/v1/*` | Browser APIs | Session cookies |
| `/admin/api/v1/*` | Admin APIs | localhost only |
| `/.well-known/*` | Discovery | Public |
