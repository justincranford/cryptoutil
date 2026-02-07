# cryptoutil Architecture

**Purpose**: Single source of truth for cryptoutil product suite architecture, design, directory structure, and deployment patterns.

**Companion Document**: [SERVICE-TEMPLATE.md](SERVICE-TEMPLATE.md) - Complete blueprint for building services.

---

## Product and Services - Authoritative Reference

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

### Product-Service Protocol Principles

These principles apply to both the Private Admin APIs and Public Business Logic APIs, for ALL 5 products and 9 services:

- ALWAYS use HTTPS for ALL API listeners; never HTTP
- ALWAYS use IPv4 127.0.0.1 OUTSIDE of Docker Compose and Kubernetes
- ALWAYS use port 0 for test configuration OUTSIDE of Docker Compose and Kubernetes; for example, in ALL unit tests and integration tests
- NEVER use localhost inside Docker Compose and Kubernetes, due to IPv4 vs IPv6 dual stack issues
- NEVER use IPv6

#### Private Admin API Listener Principles

These principles apply only to Admin APIs, for ALL 5 products and 9 services:

- ALWAYS use IPv4 127.0.0.1:9090 INSIDE Docker Compose and Kubernetes
- ALWAYS support same private health check paths (`/admin/api/v1/livez`, `/admin/api/v1/readyz`)
- ALWAYS support same private graceful shutdown path (`/admin/api/v1/shutdown`)
- NEVER expose as external port, ALWAYS keep private inside Docker Compose and Kubernetes

#### Public Business Logic API Principles

These principles apply only to Business Logic APIs, for ALL 5 products and 9 services:

- ALWAYS use IPv4 0.0.0.0:9090 INSIDE Docker Compose and Kubernetes
- ALWAYS use IPv4 127.0.0.1 OUTSIDE of Docker Compose and Kubernetes, and mutually exclusive port ranges
- ALWAYS map to external port OUTSIDE of Docker Compose and Kubernetes
- ALWAYS support same public health check path (`/admin/api/v1/health`)
- ALWAYS support same public business logic API path structure; `/service/api/v1/*` for non-browser (headless, microservice) clients, `/browser/api/v1/*` for browser-based clients

### PostgreSQL Ports Principles

ALWAYS use same IPv4 0.0.0.0:5432 INSIDE Docker Compose and Kubernetes
ALWAYS use map to unique, fixed host port per-service OUTSIDE Docker Compose and Kubernetes
ALWAYS use IPv4 127.0.0.1 and port 0 for test-containers OUTSIDE Docker Compose and Kubernetes; for example, in ALL unit tests and integration tests

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

### Telemetry Ports Principles

ALL 5 products and 9 services MUST use a single shared telemetry stack (otel-collector + grafana-otel-lgtm)
The shared telemetry stack MUST run INSIDE Docker Compose and Kubernetes, with fixed ports:

| Service | Host Port | Container Port | Protocol |
|---------|-----------|----------------|----------|
| opentelemetry-collector-contrib | 127.0.0.1:4317 | 0.0.0.0:4317 | OTLP gRPC (no TLS) |
| opentelemetry-collector-contrib | 127.0.0.1:4318 | 0.0.0.0:4318 | OTLP HTTP (no TLS) |
| grafana-otel-lgtm | 127.0.0.1:3000 | 0.0.0.0:3000 | HTTP (no TLS) |
| grafana-otel-lgtm | 127.0.0.1:4317 | 0.0.0.0:4317 | OTLP gRPC (no TLS) |
| grafana-otel-lgtm | 127.0.0.1:4318 | 0.0.0.0:4318 | OTLP HTTP (no TLS) |

---

## Directory Structure

### cmd/ - CLI Entry Points

```
cmd/
├── cryptoutil/main.go         # Suite-level CLI (all products): Thin main() call to `internal/apps/cryptoutil.go`
├── cipher/main.go             # Product-level Cipher CLI: Thin main() call to `internal/apps/cipher/cipher.go`
├── jose/main.go               # Product-level JOSE CLI: Thin main() call to `internal/apps/jose/jose.go`
├── pki/main.go                # Product-level PKI CLI: Thin main() call to `internal/apps/pki/pki.go`
├── identity/main.go           # Product-level Identity CLI: Thin main() call to `internal/apps/identity/identity.go`
├── sm/main.go                 # Product-level SM CLI: Thin main() call to `internal/apps/sm/sm.go`
├── cipher-im/main.go          # Service-level Cipher-IM CLI: Thin main() call to `internal/apps/cipher/im/im.go`
├── jose-ja/main.go            # Service-level JOSE-JA CLI: Thin main() call to `internal/apps/jose/ja/ja.go`
├── pki-ca/main.go             # Service-level PKI-CA CLI: Thin main() call to `internal/apps/pki/ca/ca.go`
├── identity-authz/main.go     # Service-level Identity-Authz CLI: Thin main() call to `internal/apps/identity/authz/authz.go`
├── identity-idp/main.go       # Service-level Identity-IDP CLI: Thin main() call to `internal/apps/identity/idp/idp.go`
├── identity-rp/main.go        # Service-level Identity-RP CLI: Thin main() call to `internal/apps/identity/rp/rp.go`
├── identity-rs/main.go        # Service-level Identity-RS CLI: Thin main() call to `internal/apps/identity/rs/rs.go`
├── identity-spa/main.go       # Service-level Identity-SPA CLI: Thin main() call to `internal/apps/identity/spa/spa.go`
└── sm-kms/main.go             # Service-level SM-KMS CLI (legacy): Thin main() call to `internal/apps/sm/kms/kms.go`
```

**Pattern**: Thin `main()` pattern for all cmd/ CLIs, with all logic in `internal/apps/` for maximum code reuse and testability.

1. `cmd/cryptoutil/` for suite-level CLI
```go
func main() {
    os.Exit(cryptoutilAppsSuite.Suite(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
```
2. `cmd/<product>/` for product-level CLI
```go
func main() {
    os.Exit(cryptoutilApps<PRODUCT>.<PRODUCT>(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
```
3. `cmd/<product>/<service>/` for service-level CLI
```go
func main() {
    os.Exit(cryptoutilApps<PRODUCT><SERVICE>.<SERVICE>(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
```

### internal/apps/ - Service Implementations

```
internal/apps/
├── template/                  # REUSABLE product-service template (all 9 services for all 5 products MUST reuse this template for maximum consistency and minimum duplication)
│   ├── service/
│   │   ├── config/            # ServiceTemplateServerSettings
│   │   ├── server/            # Application, PublicServerBase, AdminServer
│   │   │   ├── application/   # ApplicationCore, ApplicationBasic
│   │   │   ├── builder/       # ServerBuilder fluent API
│   │   │   ├── listener/      # AdminHTTPServer
│   │   │   ├── barrier/       # Encryption-at-rest service
│   │   │   ├── businesslogic/ # SessionManager, TenantRegistration
│   │   │   ├── repository/    # TenantRepo, RealmRepo, SessionRepo
│   │   │   └── realms/        # Authentication realm implementations
│   │   └── testutil/          # Test helpers (NewTestSettings)
│   └── testing/
│       └── e2e/               # ComposeManager for E2E orchestration
├── cipher/
│   └── im/                    # Cipher-IM service
│       ├── domain/            # Domain models (Message, Recipient)
│       ├── repository/        # Domain repos + migrations (2001+)
│       ├── server/            # CipherIMServer, PublicServer
│       │   ├── config/        # CipherImServerSettings embeds template
│       │   └── apis/          # HTTP handlers
│       ├── client/            # API client
│       ├── e2e/               # E2E tests (Docker Compose)
│       └── integration/       # Integration tests
├── jose/
│   └── ja/                    # JOSE-JA service (same structure)
├── pki/
│   └── ca/                    # PKI-CA service (same structure)
├── sm/
│   └── jose/                  # SM-KMS service (same structure)
└── identity/
    ├── authz/                 # OAuth 2.1 Authorization Server (same structure)
    ├── idp/                   # OIDC 1.0 Identity Provider (same structure)
    ├── rs/                    # OAuth 2.1 Resource Server (same structure)
    ├── rp/                    # OAuth 2.1 Relying Party (same structure)
    └── spa/                   # OAuth 2.1 Single Page Application (same structure)
```

### internal/shared/ - Shared Utilities

```
internal/shared/
├── apperr/                  # Application errors
├── container/               # Dependency injection container
├── config/                  # Configuration helpers
├── crypto/                  # Cryptographic utilities
├── magic/                   # Named constants (ports, timeouts, paths)
├── pool/                    # Generator pool utilities
├── pwdgen/                  # Password generator utilities
├── telemetry/               # OpenTelemetry integration
└── testutil/                # Shared test utilities
```

### deployments/ - Docker Compose

```
deployments/
├── telemetry/
│   └── compose.yml
├── sm-kms/
│   ├── config/
|   │   ├── common.yml        # common configuration for all 3 sm-kms instances
|   │   ├── postgresql-1.yml  # instance 1 of sm-kms; uses shared sm-kms PostgreSQL
|   │   ├── postgresql-2.yml  # instance 2 of sm-kms; uses shared sm-kms PostgreSQL
|   │   └── sqlite.yml        # instance 3 of sm-kms; uses non-shared in-memory sm-kms SQLite
│   ├── secrets/
|   │   ├──postgres_url.secret      # Docker Compose secret shared by 2 instances of sm-kms; PostgreSQL instances only
|   │   ├──postgres_database.secret # Docker Compose secret shared by 2 instances of sm-kms; PostgreSQL instances only
|   │   ├──postgres_username.secret # Docker Compose secret shared by 2 instances of sm-kms; PostgreSQL instances only
|   │   ├──postgres_password.secret # Docker Compose secret shared by 2 instances of sm-kms; PostgreSQL instances only
|   │   ├──unseal_1of5.secret       # Docker Compose secret shared by 3 instances of sm-kms; unseal service
|   │   ├──unseal_2of5.secret       # Docker Compose secret shared by 3 instances of sm-kms; unseal service
|   │   ├──unseal_3of5.secret       # Docker Compose secret shared by 3 instances of sm-kms; unseal service
|   │   ├──unseal_4of5.secret       # Docker Compose secret shared by 3 instances of sm-kms; unseal service
|   │   ├──unseal_5of5.secret       # Docker Compose secret shared by 3 instances of sm-kms; unseal service
|   │   └──hash_pepper.secret       # Docker Compose secret shared by 3 instances of sm-kms; hash registries of hash algorithms
│   ├── compose.yml                 # Docker Compose config: `builder-cryptoutil` builds Dockerfile, 3 instances of sm-kms depend on it
│   └── Dockerfile                  # Dockerfile: compose.yml `builder-cryptoutil` builds this Dockerfile
├── <PRODUCT>/
│   └── ... (same structure)
├── jose/
│   └── ... (same structure)
├── ca/
│   └── ... (same structure)
├── identity/
│   └── ... (same structure)
└── cipher/
    └── ... (same structure)
```

## CLI Patterns

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

### CLI Subcommand

All CLIs for all 9 services MUST support these subcommands, with consistent behavior and config parsing and flag parsing.
Consistency MUST be guaranteed by inheriting from service-template, which will reuse `internal/apps/template/service/<SUBCOMMAND>/` packages:

| Subcommand | Description |
|------------|-------------|
| `server` | CLI server start with dual HTTPS listeners, for Private Admin APIs vs Public Business Logic APIs |
| `health` | CLI client for Public health endpoint API check |
| `livez` | CLI client for Private liveness endpoint API check |
| `readyz` | CLI client for Private readiness endpoint API check |
| `shutdown` | CLI client for Private graceful shutdown endpoint API trigger |
| `client` | CLI client for Business Logic API interaction (n.b. domain-specific for each of the 9 services) |
| `init` | CLI client for Initialize static config, like TLS certificates |
| `demo` | CLI client for start server, inject Demo data, and run clients |

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

**Realms do NOT scope data** - all realms in same tenant see same data. Only `tenant_id` scopes data access.

#### Realms

Every REALM is configurable per-service as file-based static config, dynamic database-backed config, or both.
Services may include or omit any REALM, including file-based or database-backed.
Requirement: A minimum of 1 realm per service MUST be configured for authentication to work.
Recommendation: A minimum of 2 realms per service SHOULD be configured, 1 file-based for admin authentication, and 1 file-based or database-based for sessions cookie validation.

All REALMS must be implemented in service-template, and inherited by all 9 services for all 9 products, for maximum consistency and minimum duplication.

All users and clients must authenticate successfully with one of the `authentication-*` realms, to receive a session cookie.
All subsequent API calls must use the session cookie; The session cookie may be stateless (e.g. JWE, JWS) or stateful (e.g. opaque cookie).

**Login Realms for Every Service**

| Type | Scheme | Credential | Login or Session | Credential Store |
|------|--------|------------|-------------------------|=-----------------|
| `https-client-cert-login-file` | HTTP/mTLS Handshake | HTTPS Client Certificate | Login | File |
| `https-client-cert-login-database` | HTTP/mTLS Handshake | HTTPS Client Certificate | Login | Database |
| `https-client-cert-login-federated` | HTTP/mTLS Handshake | HTTPS Client Certificate | Login | Federated |
| `bearer-token-opaque-login-file` | HTTP Header 'Authorize' Bearer | Opaque Token | Login | File |
| `bearer-token-opaque-login-database` | HTTP Header 'Authorize' Bearer | Opaque Token | Login | Database |
| `bearer-token-opaque-login-federated` | HTTP Header 'Authorize' Bearer | Opaque Token | Login | Federated |
| `bearer-token-jwe-login-file` | HTTP Header 'Authorize' Bearer | JWE Token | Login | File |
| `bearer-token-jwe-login-database` | HTTP Header 'Authorize' Bearer | JWE Token | Login | Database |
| `bearer-token-jwe-login-federated` | HTTP Header 'Authorize' Bearer | JWE Token | Login | Federated |
| `bearer-token-jws-login-file` | HTTP Header 'Authorize' Bearer | JWS Token | Login | File |
| `bearer-token-jws-login-database` | HTTP Header 'Authorize' Bearer | JWS Token | Login | Database |
| `bearer-token-jws-login-federated` | HTTP Header 'Authorize' Bearer | JWS Token | Login | Federated |
| `basic-username-password-login-file` | HTTP Header 'Authorize' Basic | Username/Password | Login | File |
| `basic-username-password-login-database` | HTTP Header 'Authorize' Basic | Username/Password | Login | Database |
| `basic-username-password-login-federated` | HTTP Header 'Authorize' Basic | Username/Password | Login | Federated |
| `cookie-token-opaque-session-file` | HTTP Header 'Cookie' Token | Opaque Token | Session | File |
| `cookie-token-opaque-session-database` | HTTP Header 'Cookie' Token | Opaque Token | Session | Database |
| `cookie-token-opaque-session-federated` | HTTP Header 'Cookie' Token | Opaque Token | Session | Federated |
| `cookie-token-jwe-session-file` | HTTP Header 'Cookie' Token | JWE Token | Session | File |
| `cookie-token-jwe-session-database` | HTTP Header 'Cookie' Token | JWE Token | Session | Database |
| `cookie-token-jwe-session-federated` | HTTP Header 'Cookie' Token | JWE Token | Session | Federated |
| `cookie-token-jws-session-file` | HTTP Header 'Cookie' Token | JWS Token | Session | File |
| `cookie-token-jws-session-database` | HTTP Header 'Cookie' Token | JWS Token | Session | Database |
| `cookie-token-jws-session-federated` | HTTP Header 'Cookie' Token | JWS Token | Session | Federated |

| `authorization-code-opaque-login-file` | OAuth 2.1 Authorization Code Flow + PKCE | Opaque | Login | File |
| `authorization-code-jwe-login-file` | OAuth 2.1 Authorization Code Flow + PKCE | JWE | Login | File |
| `authorization-code-jws-login-file` | OAuth 2.1 Authorization Code Flow + PKCE | JWS | Login | File |
| `authorization-code-opaque-login-database` | OAuth 2.1 Authorization Code Flow + PKCE | Opaque | Login | Database |
| `authorization-code-jwe-login-database` | OAuth 2.1 Authorization Code Flow + PKCE | JWE | Login | Database |
| `authorization-code-jws-login-database` | OAuth 2.1 Authorization Code Flow + PKCE | JWS | Login | Database |
| `authorization-code-login-federated` | OAuth 2.1 Authorization Code Flow + PKCE | Opaque | Login | Federated |

| `webauthn-login-file` | WebAuthn (2013) | PublicKeyCredential | Login | File |
| `webauthn-login-database` | WebAuthn (2013) | PublicKeyCredential | Login | Database |
| `passkey-login-file` | WebAuthn (2021) Passkey | PublicKeyCredential | Login | File |
| `passkey-login-database` | WebAuthn (2021) Passkey | PublicKeyCredential | Login | Database |

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
