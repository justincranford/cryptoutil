# cryptoutil Architecture

**Purpose**: Single source of truth for cryptoutil product suite design, directory structure, and deployment patterns.

**Companion Document**: [SERVICE-TEMPLATE.md](SERVICE-TEMPLATE.md) - Complete blueprint for building services.

---

## Suite Overview - Product and Services - Complete Reference

This section provides the authoritative address and port bindings for all 9 services in 5 products.

| Product | Service | Product-Service Identifier | Host Public Address | Host Port Range | Container Public Address | Container Public Port Range | Container Admin Private Address | Container Admin Port Range | Description |
|---------|----------------|-----------------|------------|----------|------------|----------------|-------------------|----------|
| **Private Key Infrastructure (PKI)** | **Certificate Authority (CA)** | **pki-ca** | 127.0.0.1 | 8050-8059 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | X.509 certificates, EST, SCEP, OCSP, CRL |
| **JSON Object Signing and Encryption (JOSE)** | **JWK Authority (JA)** | **jose-ja** | 127.0.0.1 | 8060-8069 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | JWK/JWS/JWE/JWT operations |
| **Cipher** | **Instant Messenger (IM)** | **cipher-im** | 127.0.0.1 | 8070-8079 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | E2E encrypted messaging, encryption-at-rest |
| **Secrets Manager (SM)** | **Key Management Service (KMS)** | **sm-kms** | 127.0.0.1 | 8080-8089 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | Elastic key management, encryption-at-rest |
| **Identity** | **Authorization Server (Authz)** | **identity-authz** | 127.0.0.1 | 8100-8109 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | OAuth 2.1 authorization server |
| **Identity** | **Identity Provider (IdP)** | **identity-idp** | 127.0.0.1 | 8110-8119 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | OIDC 1.0 Identity Provider |
| **Identity** | **Resource Server (RS)** | **identity-rs** | 127.0.0.1 | 8120-8129 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | OAuth 2.1 Resource Server |
| **Identity** | **Relying Party (RP)** | **identity-rp** | 127.0.0.1 | 8130-8139 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | OAuth 2.1 Relying Party |
| **Identity** | **Single Page Application (SPA)** | **identity-spa** | 127.0.0.1 | 8140-8149 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | OAuth 2.1 Single Page Application |

### Product-Service Port Design Principles

- HTTPS protocol for all public and admin port bindings.
- Same HTTPS 127.0.0.1:9090 for Private HTTPS Admin APIs inside Docker Compose and Kubernetes; never localhost due to IPv4 vs IPv6 dual stack issues; never exposed outside of containers.
- Same HTTPS 0.0.0.0:8080 for Public HTTPS APIs inside Docker Compose and Kubernetes.
- Different HTTPS 127.0.0.1 port range mappings for Public APIs on Docker host, to avoid conflicts.
- Same health check paths (`/browser/api/v1/health`, `/service/api/v1/health`) on Public HTTPS listeners.
- Same health check paths (`/admin/api/v1/livez`, `/admin/api/v1/readyz`) on Private HTTPS Admin listeners.
- Same graceful shutdown path (`/admin/api/v1/shutdown`) on Private HTTPS Admin listeners.

### PostgreSQL Ports

| Product-Service Identifier | Host Address | Host Port | Container Address | Container Port |
|---------|-----------|----------------|----------|----------------|
| **pki-ca** | 127.0.0.1 | 54320 | 0.0.0.0 | 5432 |
| **jose-ja** | 127.0.0.1 | 54321 | 0.0.0.0 | 5432 |
| **cipher-im** | 127.0.0.1 | 54322 | 0.0.0.0 | 5432 |
| **sm-kms** | 127.0.0.1 | 54323 | 0.0.0.0 | 5432 |
| **identity-authz** | 127.0.0.1 | 54324 | 0.0.0.0 | 5432 |
| **identity-idp** | 127.0.0.1 | 54325 | 0.0.0.0 | 5432 |
| **identity-rs** | 127.0.0.1 | 54326 | 0.0.0.0 | 5432 |
| **identity-rp** | 127.0.0.1 | 54327 | 0.0.0.0 | 5432 |
| **identity-spa** | 127.0.0.1 | 54328 | 0.0.0.0 | 5432 |

### PostgreSQL Port Design Principles for Product-Service Databases

- Same 0.0.0.0:5432 inside Docker Compose and Kubernetes.
- Same 127.0.0.1 host address on Docker host.
- Different host port mappings (54320-54329) for each product-service to avoid conflicts on Docker host.

### Telemetry Ports (Shared)

| Service | Host Port | Container Port | Protocol |
|---------|-----------|----------------|----------|
| opentelemetry-collector-contrib | 4317 | 4317 | OTLP gRPC |
| opentelemetry-collector-contrib | 4318 | 4318 | OTLP HTTP |
| grafana-otel-lgtm | 3000 | 3000 | HTTP (UI) |
| grafana-otel-lgtm | 4317 | 4317 | OTLP gRPC |
| grafana-otel-lgtm | 4318 | 4318 | OTLP HTTP |

---

## Directory Structure

### cmd/ - CLI Entry Points

```
cmd/
├── cryptoutil/main.go         # Suite-level CLI (all products): Delegates to `internal/apps/`
├── cipher/main.go             # Product-level Cipher CLI: Delegates to `internal/apps/cipher/`
├── jose/main.go               # Product-level JOSE CLI: Delegates to `internal/apps/jose/`
├── pki/main.go                # Product-level PKI CLI: Delegates to `internal/apps/pki/`
├── identity/main.go           # Product-level Identity CLI: Delegates to `internal/apps/identity/`
├── sm/main.go                 # Product-level SM CLI: Delegates to `internal/apps/sm/`
├── cipher-im/main.go          # Service-level Cipher-IM CLI: Delegates to `internal/apps/cipher/im/`
├── jose-ja/main.go        # Service-level JOSE-JA CLI: Delegates to `internal/apps/jose/ja/`
├── pki-ca/main.go             # Service-level PKI-CA CLI: Delegates to `internal/apps/pki/ca/`
├── identity-authz/main.go     # Service-level Identity-Authz CLI: Delegates to `internal/apps/identity/authz/`
├── identity-idp/main.go       # Service-level Identity-IDP CLI: Delegates to `internal/apps/identity/idp/`
├── identity-rp/main.go        # Service-level Identity-RP CLI: Delegates to `internal/apps/identity/rp/`
├── identity-rs/main.go        # Service-level Identity-RS CLI: Delegates to `internal/apps/identity/rs/`
├── identity-spa/main.go       # Service-level Identity-SPA CLI: Delegates to `internal/apps/identity/spa/`
└── sm-kms/main.go             # Service-level SM-KMS CLI (legacy): Delegates to `internal/apps/sm/kms/`
```

**Pattern**: Thin `main()` delegates to:
- `internal/apps/cryptoutil/` for suite-level CLI
- `internal/apps/<product>/` for product-level CLI
- `internal/apps/<product>/<service>/` for service-level CLI

```go
func main() {
    os.Exit(cryptoutilAppsCipherIm.IM(os.Args, os.Stdin, os.Stdout, os.Stderr))
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
