# cryptoutil Architecture

**Purpose**: Single source of truth for cryptoutil product suite architecture, design, directory structure, and deployment patterns.

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

- ALWAYS use HTTPS for ALL API listeners; never HTTP; TLS supports autoconfig or static config
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

## Database Architecture

### Dual Database Support

All 9 services MUST support using one of PostgreSQL or SQLite, specified via configuration at startup.

Typical usages for each database for different purposes:
- Unit tests, Fuzz tests, Benchmark tests, Mutations tests => Ephemeral SQLite instance (e.g. in-memory)
- Integration tests, Load tests => Ephemeral PostgreSQL instance (i.e. test-container)
- End-to-End tests => Static PostgreSQL instance (e.g. Docker Compose)
- Production => Static PostgreSQL instance (e.g. Cloud hosted)
- Local Development => Static SQLite instance (e.g. file); used for local development

Caveat: End-to-End Docker Compose tests use both PostgreSQL and SQLite, for isolation testing; 3 service instances, 2 using a shared PostgreSQL container, and 1 using in-memory SQLite

### Cross-DB Compatibility Rules

```go
// UUID fields: ALWAYS type:text (SQLite has no native UUID)
ID googleUuid.UUID `gorm:"type:text;primaryKey"`

// Nullable UUIDs: Use NullableUUID (NOT *googleUuid.UUID)
ClientProfileID NullableUUID `gorm:"type:text;index"`

// JSON arrays: ALWAYS serializer:json (NOT type:json)
AllowedScopes []string `gorm:"serializer:json"`
```

### SQLite Configuration

```go
sqlDB.Exec("PRAGMA journal_mode=WAL;")       // Concurrent reads + 1 writer
sqlDB.Exec("PRAGMA busy_timeout = 30000;")   // 30s retry on lock
sqlDB.SetMaxOpenConns(5)                     // GORM transactions need multiple
```

### SQLite DateTime (CRITICAL)

**ALWAYS use `.UTC()` when comparing with SQLite timestamps**:

```go
// ❌ WRONG: time.Now() without .UTC()
if session.CreatedAt.After(time.Now()) { ... }

// ✅ CORRECT: Always use .UTC()
if session.CreatedAt.After(time.Now().UTC()) { ... }
```

**Pre-commit hook auto-converts** `time.Now()` → `time.Now().UTC()`.

### File Numbering for SQL Go-Migrations DDL/DML Files

| Range | Owner | Examples |
|-------|-------|----------|
| 1001-1999 | Service Template | Sessions (1001), Barrier (1002), Realms (1003), Tenants (1004), PendingUsers (1005) |
| 2001+ | Domain | cipher-im messages (2001), jose JWKs (2001) |

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

### Authentication Realms

**CRITICAL**: Realms define authentication METHOD and POLICY, NOT data scoping.

**Realms do NOT scope data** - all realms in same tenant see same data. Only `tenant_id` scopes data access.

#### Authentication Realm Types

| Realm Type | Purpose | Scheme | Credential | Credential Validators |
|------|--------|------------|-------------------------|-----------------|
| `https-client-cert-factor` | Create or Upgrade Session | HTTP/mTLS Handshake | HTTPS Client Certificate | File, Database, Federated |
| `webauthn-resident-synced-factor` | Create or Upgrade Session | WebAuthn L2 Resident Synced (aka Passkeys) | Local PublicKeyCredential | File, Database, Federated |
| `webauthn-resident-unsynced-factor` | Create or Upgrade Session | WebAuthn L2 Resident Unsynced (e.g. Windows Hello) | Local PublicKeyCredential | File, Database, Federated |
| `webauthn-nonresident-synced-factor` | Create or Upgrade Session | WebAuthn L2 Non-Resident Synced (e.g. Azure AD) | Cloud PublicKeyCredential | File, Database, Federated |
| `webauthn-nonresident-unsynced-factor` | Create or Upgrade Session | WebAuthn L2 Non-Resident Unsynced (e.g. YubiKey) | Cloud PublicKeyCredential | File, Database, Federated |
| `authorization-code-opaque-factor` | Create or Upgrade Session | OAuth 2.1 Authorization Code Flow + PKCE | Opaque | File, Database, Federated |
| `authorization-code-jwe-factor` | Create or Upgrade Session | OAuth 2.1 Authorization Code Flow + PKCE | JWE | File, Database, Federated |
| `authorization-code-jws-factor` | Create or Upgrade Session | OAuth 2.1 Authorization Code Flow + PKCE | JWS | File, Database, Federated |
| `bearer-token-opaque-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Bearer | Opaque Token | File, Database, Federated |
| `bearer-token-jwe-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Bearer | JWE Token | File, Database, Federated |
| `bearer-token-jws-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Bearer | JWS Token | File, Database, Federated |
| `basic-username-password-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | Username/Password | File, Database, Federated |
| `basic-email-password-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | Email/Password | File, Database, Federated |
| `basic-email-otp-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | Email/RandomOTP | File, Database, Federated |
| `basic-email-magiclink-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic + Query String | Email/Nothing & QueryParameter | File, Database, Federated |
| `basic-sms-password-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | Phone/Password | File, Database, Federated |
| `basic-sms-otp-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | Phone/RandomOTP | File, Database, Federated |
| `basic-sms-magiclink-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic + Query String | Phone/Nothing & QueryParameter | File, Database, Federated |
| `basic-voice-otp-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | Phone/RandomOTP | File, Database, Federated |
| `basic-id-otp-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | ID/Nothing & HOTP/TOTP | File, Database, Federated |
| `cookie-token-opaque-session` | Use Session | HTTP Header 'Cookie' Token | Opaque Token | File, Database, Federated |
| `cookie-token-jwe-session` | Use Session | HTTP Header 'Cookie' Token | JWE Token | File, Database, Federated |
| `cookie-token-jws-session` | Use Session | HTTP Header 'Cookie' Token | JWS Token | File, Database, Federated |

### Authentication Realm Principals

1. Every service MUST configure a prioritized list of realm instances; multiple realm instances of same realm type are allowed.
2. Every service MUST configure one or more factor realms, for creating or upgrading sessions; zero factor realms is NOT allowed.
3. Every service MUST configure one or more session realms, for using sessions; zero session realms is NOT allowed.
4. Every realm instance MUST specify one-and-only-one credential validator; the only valid credential validator options are file-backed, database-backed, or federated.
5. Every factor realm instance MUST return a created or rotated session cookie on successful authentication.
6. Every session realm instance MAY return a rotated session cookie on successful authentication; mitigates session fixation.
7. Every service is RECOMMENDED to include at least one file-based factor realm for fallback session creation, plus at least one file-based session realm for session use.
8. Realm design and implementation MUST be encapsulated in service-template, and inherited by all 9 services for all 5 products, for maximum consistency and minimum duplication.
9. All browser users and headless clients MUST first authenticate successfully to a factor realm for session creation, and use the session token for all subsequent API calls.

### Tenant and Member Registration Flow

Tenants are never created on their own; they are automatically created when a new user registers and omits `tenant_id`; a new tenant is automatically created, and the new user is automatically assigned as the new tenant's admin.

**New Tenant** (omit tenant_id):

```http
POST /browser/api/v1/register
{ "realm-type": "basic-username-password-factor", "realm-name": "admins", "username": "admin@example.com", "password": "securepass" }
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

## Security Architecture

### FIPS 140-3 Compliance (ALWAYS Enabled)

**Approved Algorithms**:

| Category | Algorithms |
|----------|------------|
| Asymmetric | RSA ≥2048, DH ≥2048, ECDSA P256/P384/P521, ECDH P256/P384/P521, EdDSA 25519/448, EdDH X25519/X448 |
| Symmetric | AES-128/192/256 (GCM, CBC+HMAC, CMAC) |
| Digest | SHA-256/384/512, HMAC-SHA-256/384/512 |
| KDF | PBKDF2-HMAC-SHA256/384/512, HKDF-SHA256/384/512 |

**Banned**: bcrypt, scrypt, Argon2, MD5, SHA-1, RSA <2048, DH <2048, EC < P256, DES, 3DES.

### Key Hierarchy (Barrier Service)

```
Unseal Key (Docker secrets, NEVER stored)
    └── Root Key (encrypted-at-rest with unseal key(s), rotated manually or automatically annually)
        └── Intermediate Key (encrypted-at-rest with root key, rotated manually or automatically quarterly)
            └── Content Key (encrypted-at-rest with intermediate key, rotated manually or automatically monthly)
                └── Domain Data (encrypted-at-rest with content key) - Examples: Cipher-IM messages, SM-KMS JWKs, JOSE-JA JWKs, PKI-CA private keys, Identity user credentials
```

Design Intent: Unseal secret(s) or unseal key(s) are loaded by service instances at startup. To decrypt and reuse existing, sealed root keys in a database, each service instance MUST use unseal credentials to unseal the root keys. This is design intent for barrier service.

### Hash Service (Version-Based)

#### Hash Types

Hash service supports 4 hash types.
1. Low-entropy, random-salt => Used for short values that DON'T need to be indexed or searched in a database (e.g. Passwords)
2. Low-entropy, fixed-salt => Used for short values that DO need to be indexed and searched in a database (e.g. PII, Usernames, Emails, Addresses, Phone Numbers, SIN/SSN/NIN, IPs, MACs)
3. High-entropy, random-salt => Used for long values that DON'T need to be indexed or searched in a database (e.g. Private Keys)
4. High-entropy, fixed-salt => Used for long values that DO need to be indexed and searched in a database, inputs MUST have a minimum of 256-bits (32-bytes) of entropy

##### Low-entropy vs High-entropy

Low entropy: Values with >= 256-bits (32-bytes) or higher of brute-force search space; values are hashed with high-iterations PBKDF2 to mitigate brute-force attacks, because small search spaces are not big enough to mitigate brute-force attacks on their own; do not use HKDF, it does not add sufficient security for low-entropy values

High entropy: Values with < 256-bits (32-bytes) of brute-force search space; values are hashed with one-iteration HKDF, because large search space is big enough to mitigate brute-force attacks on its own; do not use PBKDF2, extra iterations do not add meaningful security

##### Random-salt vs Fixed-salt

Random salt: Used for values that DON'T require indexing or searching in a database; non-deterministic hash outputs for the same input is best practice for security

Fixed-salt: Used for values that DO require indexing or searching in a database; deterministic hash outputs for the same input are required for indexing and searching, which overrides best practice for security; to mitigate reduced security of using fixed-salt, pepper MUST be applied to all values before passing them into hash functions

#### Pepper

##### Pepper Usage

Pepper MUST be used on all values passed into hash functions that use fixed-salt.
Pepper SHOULD be used on all values passed into hash functions that use random-salt
For consistency, pepper usage WILL be used on all values passed to all hash functions, regardless of salt type.

##### Pepper Algorithms

Pepper before deterministic hashing MUST use AES-GCM-SIV.
Pepper before non-deterministic hashing MUST use AES-GCM-SIV or AES-GCM. The AES-256 key MUST be generated and used for the lifetime of the hash.

##### Pepper Generation and Usage

Use AES-GCM-SIV for pepper before doing deterministic hashing.

Generation:
1. Generate a random 32-byte AES key.
2. Select a unique 12-byte nonce; random bytes, derived bytes, or monotonic increasing counter.
3. Select an optional associated data (AAD).
4. Encode the 3 values together as the pepper.
5. Use Barrier Service to encrypt-at-rest the pepper.
6. Persist the encrypted pepper in the database.

Usage (Index and Search):
1. Load the encrypted pepper from the database.
2. Use Barrier Service to decrypt the pepper at runtime.
3. Use the 3 clear values of the pepper together on the input secret.
4. Encode the pepper inputs with the ciphertext to create in-memory encoded pepper output.
5. Deterministic hash the encoded pepper output.
6. Encode the pepper for persistence; type and version only.
7. Concatenate the encoded pepper and encoded hash.
8. Use the concatenated value for indexing or searching.

Use AES-GCM for pepper before doing non-deterministic hashing:

Generation:
1. Generate a random 32-byte AES key.
2. Encode the 1 value as the pepper.
3. Use Barrier Service to encrypt-at-rest the pepper.
4. Persist the encrypted pepper in the database.

Usage (Store):
1. Load the encrypted pepper from the database.
2. Use Barrier Service to decrypt the pepper at runtime.
3. Select a unique 12-byte nonce; random bytes, derived bytes, or monotonic increasing counter.
4. Select an optional associated data (AAD).
5. Use the 1 clear value of the pepper, and the 2 clear selected values, together on the input secret.
6. Non-deterministic hash the encoded pepper output.
7. Encode the pepper for persistence; type, version, nonce, and optional AAD.
8. Concatenate the encoded pepper and encoded hash.
9. Use the concatenated value for storage.

Usage (Validate):
1. Load the encrypted pepper from the database.
2. Use Barrier Service to decrypt the pepper at runtime.
3. Parse the unique 12-byte nonce from the stored encoded pepper.
4. Parse the optional associated data (AAD) from the stored encoded pepper.
5. Use the 1 clear value of the pepper, and the 2 clear parsed values, together on the input secret.
6. Non-deterministic hash the ciphertext output from the pepper step.
7. Compare the store hash to the computed hash.

##### Low-Entropy Hash Format

```
Format: {pepperTypeAndVersion}base64(optionalPepperNonce):base64(optionalPepperAAD)#{hashTypeAndVersion}:{algorithm}:{iterations}:base64(salt):base64(hash)
Deterministic Example:     {d2}#{f5}:PBKDF2-HMAC-SHA256:600000:abc123...:def456...
Non-Deterministic Example: {n2}nonce#{f5}PBKDF2-HMAC-SHA256:600000:abc123...:def456...
Non-Deterministic Example: {n2}nonce:aad#{f5}PBKDF2-HMAC-SHA256:600000:abc123...:def456...
```

##### High-Entropy Hash Format

```
Format: {pepperTypeAndVersion}base64(optionalPepperNonce):base64(optionalPepperAAD)#{hashTypeAndVersion}:{algorithm}:base64(salt):base64(info):base64(hash)
Deterministic Example:     {d2}#{F5}:HKDF-HMAC-SHA256:abc123...:def456...:ghi789...
Non-Deterministic Example: {n2}nonce#{R5}HKDF-HMAC-SHA256:abc123...:def456...:ghi789...
Non-Deterministic Example: {n2}nonce:aad#{R5}HKDF-HMAC-SHA256:abc123...:def456...:ghi789...
```

---

## Docker Compose Patterns

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

**Use wget (Alpine), 127.0.0.1 (not localhost), port 9090.**

---

## Quality Gates

### Pre-Commit

1. `golangci-lint run --fix` → zero warnings
2. `go build ./...` → clean build
3. `go test -coverprofile=./test-reports/cov.out ./... && go tool cover -html=./test-reports/cov.out -o ./test-reports/cov.html` → 100% tests pass with 98% code coverage

### CI/CD

| Workflow | Requirement |
|----------|-------------|
| Coverage | ≥95% production, ≥98% infrastructure/utility |
| Mutation (gremlins) | ≥85% production, ≥98% infrastructure |
| E2E | BOTH `/service/**` AND `/browser/**` paths |

---

## Configuration Priority

**Load Order** (highest to lowest):

1. **Docker Secrets** (`file:///run/secrets/secret_name`) - Sensitive values
2. **YAML Configuration** (`--config=/path/to/config.yml`) - Primary configuration
3. **CLI Parameters** (`--bind-public-port=8080`) - Overrides

**CRITICAL: Environment variables NOT desirable for configuration** (security risk, not scalable, auditability).

---

## *FromSettings Factory Pattern (PREFERRED)

Services should use settings-based factories for testability and consistency:

```go
// ✅ PREFERRED: Settings-based factory
type UnsealKeysSettings struct {
    KeyPaths []string `yaml:"key_paths"`
}

func NewUnsealKeysServiceFromSettings(settings *UnsealKeysSettings) (*UnsealKeysService, error) {
    if settings == nil {
        return nil, errors.New("settings required")
    }
    return &UnsealKeysService{
        keyPaths: settings.KeyPaths,
    }, nil
}

// Usage in ServerBuilder
builder.WithUnsealKeysService(func(settings *UnsealKeysSettings) (*UnsealKeysService, error) {
    return NewUnsealKeysServiceFromSettings(settings)
})
```

**Benefits**:
- All configuration in one struct
- Easy to test (pass test settings)
- Consistent initialization across codebase
- Self-documenting dependencies

---

## Test Settings Factory

Every service config should have a test settings factory:

```go
// NewTestSettings returns configuration suitable for testing
func NewTestSettings() *CipherImServerSettings {
    return &CipherImServerSettings{
        ServiceTemplateServerSettings: cryptoutilTemplateTestutil.NewTestSettings(),
        MaxMessageSize:                65536,
    }
}
```

**NewTestSettings() configures**:
- SQLite in-memory (`:memory:`)
- Port 0 (dynamic allocation, no conflicts)
- Auto-generated TLS certificates
- Disabled telemetry export
- Short timeouts for fast tests

---

## Testing Patterns (MANDATORY)

### TestMain Pattern

**ALL integration tests MUST use TestMain**:

```go
var (
    testDB     *gorm.DB
    testServer *Server
)

func TestMain(m *testing.M) {
    ctx := context.Background()

    // Create server with test configuration
    cfg := config.NewTestSettings()
    var err error
    testServer, err = NewFromConfig(ctx, cfg)
    if err != nil {
        log.Fatalf("Failed to create test server: %v", err)
    }

    // Start server
    go func() {
        if err := testServer.Start(); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()

    // Wait for ready
    if err := testServer.WaitForReady(ctx, 10*time.Second); err != nil {
        log.Fatalf("Server not ready: %v", err)
    }

    // Run tests
    exitCode := m.Run()

    // Cleanup
    testServer.Shutdown(ctx)
    os.Exit(exitCode)
}
```

### Table-Driven Tests (MANDATORY)

```go
func TestSendMessage_Validation(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        request SendMessageRequest
        wantErr string
    }{
        {
            name:    "empty content",
            request: SendMessageRequest{Content: ""},
            wantErr: "content required",
        },
        {
            name:    "no recipients",
            request: SendMessageRequest{Content: "hello", Recipients: nil},
            wantErr: "at least one recipient",
        },
        {
            name: "valid request",
            request: SendMessageRequest{
                Content:    "hello",
                Recipients: []string{googleUuid.NewV7().String()},
            },
            wantErr: "",
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            // Use unique test data
            tenantID := googleUuid.NewV7()

            err := testServer.SendMessage(ctx, tenantID, tc.request)

            if tc.wantErr != "" {
                require.Error(t, err)
                require.Contains(t, err.Error(), tc.wantErr)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Handler Tests with app.Test()

**ALWAYS use Fiber's in-memory testing**:

```go
func TestListMessages_Handler(t *testing.T) {
    t.Parallel()

    // Create standalone Fiber app
    app := fiber.New(fiber.Config{DisableStartupMessage: true})

    // Register handler under test
    msgRepo := repository.NewMessageRepository(testDB)
    handler := NewPublicServer(nil, msgRepo, nil, nil, nil)
    app.Get("/browser/api/v1/messages", handler.ListMessages)

    // Create HTTP request (no network call)
    req := httptest.NewRequest("GET", "/browser/api/v1/messages", nil)
    req.Header.Set("X-Tenant-ID", testTenantID.String())

    // Test handler in-memory
    resp, err := app.Test(req, -1)
    require.NoError(t, err)
    defer resp.Body.Close()

    require.Equal(t, 200, resp.StatusCode)
}
```

### Dynamic Test Data (UUIDv7)

```go
func TestCreate_UniqueConstraint(t *testing.T) {
    t.Parallel()

    // Generate unique IDs per test
    tenantID := googleUuid.NewV7()
    userID := googleUuid.NewV7()
    msgID := googleUuid.NewV7()

    msg := &Message{
        ID:       msgID,
        TenantID: tenantID,
        SenderID: userID,
        Content:  fmt.Sprintf("test_%s", msgID),
    }

    err := testRepo.Create(ctx, msg)
    require.NoError(t, err)
}
```

---

## E2E Testing

### ComposeManager

Use `internal/apps/template/testing/e2e/compose.go`:

```go
func TestE2E_SendMessage(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    ctx := context.Background()

    // Start Docker Compose stack
    manager := e2e.NewComposeManager(t, "../../../deployments/cipher-im")
    manager.Up(ctx)
    defer manager.Down(ctx)

    // Wait for service healthy
    manager.WaitForHealthy(ctx, "cipher-im", 60*time.Second)

    // Get TLS-enabled HTTP client
    client := manager.HTTPClient()

    // Test API
    resp, err := client.Post(
        manager.ServiceURL("cipher-im") + "/browser/api/v1/messages",
        "application/json",
        strings.NewReader(`{"content":"hello","recipients":["user-id"]}`),
    )
    require.NoError(t, err)
    defer resp.Body.Close()

    require.Equal(t, 201, resp.StatusCode)
}
```
