# cryptoutil Architecture - Single Source of Truth

**Version**: 1.0.0
**Last Updated**: 2026-01-19
**Status**: DRAFT - Requires review and refinement

**Purpose**: This document is the SINGLE SOURCE OF TRUTH for all cryptoutil architecture and design decisions. All implementation must conform to patterns defined here. NO duplication or conflicting information elsewhere.

---

## Core Design Principles

**Quality Over Speed (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO tasks skipped, NO features deprioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step (build, lint, test, coverage, mutation)
- ✅ **Reliability**: ≥95% coverage production, ≥98% infrastructure/utility, ≥95% mutation minimum production (98% ideal), ≥98% mutation infrastructure
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark tasks complete without objective evidence

**Continuous Execution (NO STOPPING)**:
- Work continues until ALL tasks complete OR user clicks STOP button
- NEVER stop to ask permission between tasks ("Should I continue?")
- NEVER pause for status updates or celebrations ("Here's what we did...")
- NEVER give up when encountering complexity (find solutions, refactor, investigate)
- NEVER skip tasks to "save time" or because they seem "less important"
- Task complete → Commit → IMMEDIATELY start next task (zero pause, zero text to user)

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
| **Cipher** | cipher-im | 8888-8889 | E2E encrypted messaging |

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
  - `/admin/api/v1/livez` - Liveness (process alive? → restart if fail)
  - `/admin/api/v1/readyz` - Readiness (dependencies healthy? → remove from LB if fail)
  - `/admin/api/v1/shutdown` - Graceful shutdown (drain requests, close resources)

**2. Database Abstraction**:
- PostgreSQL (production) OR SQLite (dev/test) dual support
- GORM ORM for type-safe queries
- Embedded migrations (golang-migrate)
- Template migrations 1001-1999 (sessions, barrier, realms, multi-tenancy, pending_users)
- Domain migrations 2001+ (service-specific tables)

**3. Multi-Tenancy Infrastructure**:
- `tenants` table (tenant metadata)
- `tenant_realms` table (authn realms per tenant - authentication ONLY, NOT data filtering)
- `pending_users` table (registration requests awaiting approval, unique username per tenant)
- `sessions` table (user sessions with tenant_id)
- `users` table (with tenant_id FK)
- **Migration 1005**: pending_users table with unique(username, tenant_id) constraint, expiration cleanup
- **NO WithDefaultTenant()**: Services start "cold", all tenants created via registration flow

**4. Barrier Service** (MANDATORY - Multi-Layer Key Hierarchy):
- **Unseal Key**: NEVER stored in app, Docker secrets at runtime, HKDF-derived from shared secrets
- **Root Key**: Encrypted at rest with unseal key, rotated annually
- **Intermediate Keys**: Encrypted with root key, rotated quarterly
- **Content Keys**: Encrypted with intermediate keys, rotated per-operation (messages) or hourly (sessions/data)
- **HKDF Derivation**: Deterministic key derivation ensures all instances derive same keys from same secrets
- **Encryption-at-rest**: ALL sensitive data (JWKs, passwords, tokens) encrypted before storage
- **Pattern**: Per-message JWK encrypted with Barrier before storage in domain table (e.g., `message_jwks`)

**5. Telemetry**:
- OpenTelemetry (traces, metrics, logs via OTLP)
- Sidecar pattern: service → otel-collector → grafana-lgtm
- Structured logging with trace correlation

**6. Configuration**:
- Docker secrets for sensitive data (`file:///run/secrets/`)
- YAML files
- Environment variables (rarest)
- CLI flags (rare, e.g. is log-level)

**7. Session Management**:
- stateless JWE, stateless JWS, or stateful Opaque session tokens (configurable)
- PostgreSQL/SQLite storage for stateful Opaque session tokens (NO Redis/Memcached)
- No storage required for stateless JWE or stateless JWS
- 30-minute re-authentication for high-sensitivity operations

**8. Hash Service** (Version-Based Policy Framework):
- **4 Hash Registries**:
  - **LowEntropyDeterministic**: `PBKDF2(input||pepper, fixedSalt, HIGH_iter, 256)` - PII lookup (usernames, emails, IPs)
  - **LowEntropyRandom**: `PBKDF2(password||pepper, randomSalt, OWASP_iter, 256)` - Password hashing
  - **HighEntropyDeterministic**: `HKDF-Extract+Expand(input||pepper, fixedSalt, info, 256)` - Config blob integrity
  - **HighEntropyRandom**: `HKDF-Extract+Expand(key||pepper, randomSalt, info, 256)` - API key storage
- **Version Tuple**: Each version = (4 Registries + Unique Pepper) based on NIST/OWASP policy
- **Pepper Management**:
  - MANDATORY for ALL inputs (Docker/K8s secret preferred)
  - NEVER in DB/source (mutually exclusive from hashes)
  - Version-specific (different pepper per version)
  - Rotation requires version bump + lazy migration
- **Hash Output Format**: `{version}:{algorithm}:{iterations}:base64(salt):base64(hash)`
- **Lazy Migration**: Old hashes stay on original version, new hashes use current version, rehash on next authentication
- **Supported Versions**:
  - V5 (OWASP 2023, 600k iterations, 16-byte salt, HMAC-SHA256) - Current (Default)
  - V4 (NIST 2021, 310k iterations, 16-byte salt, HMAC-SHA256) - Legacy, testing only
  - V3 (OWASP 2019, 100k iterations, 16-byte salt, HMAC-SHA256) - Legacy, testing only
  - V2 (NIST 2010, 10k iterations, 16-byte salt, HMAC-SHA1) - Legacy, testing only
  - V1 (PKCS#5 v2.0 2000, 1k iterations, 8-byte salt, HMAC-SHA1) - Legacy, testing only

**9. FIPS 140-3 Compliance** (MANDATORY - ALWAYS Enabled):
- **Approved Algorithms ONLY**:
  - Asymmetric: RSA ≥2048, ECDSA (P-256/384/521), ECDH (P-256/384/521), EdDSA (25519/448)
  - Symmetric: AES ≥128 (GCM, CBC+HMAC)
  - Digest: SHA-256/384/512, HMAC-SHA256/384/512
  - KDF: PBKDF2-HMAC-SHA256/384/512, HKDF-SHA256/384/512
- **BANNED Algorithms**: bcrypt, scrypt, Argon2, MD5, SHA-1, RSA <2048, DES, 3DES
- **Algorithm Agility**: Configurable algorithms with FIPS-approved defaults
- **Secure Random**: ALWAYS `crypto/rand`, NEVER `math/rand`

**10. Elastic Key Rotation** (Active + Historical Keys):
- **Key Ring Pattern**: Active key (encrypt/sign) + Historical keys (decrypt/verify)
- **Key ID Embedding**: Embed key ID with ciphertext/signature to identify correct historical key
- **Rotation**: Generate new key → set as active → keep ALL old keys for decryption/verification
- **NO deletion**: Historical keys NEVER deleted (required for decrypting old data)

### Public Server APIs (Template Infrastructure)

**Health Endpoints** (PublicServerBase):
- `GET /service/api/v1/health` - Service-to-server health check (returns JSON status)
- `GET /browser/api/v1/health` - Browser-to-server health check (returns JSON status)

**Registration Endpoints** (RegisterRegistrationRoutes):
- `POST /browser/api/v1/register` - Browser user registration (NO authentication, rate-limited)
- `POST /service/api/v1/register` - Service registration (NO authentication, rate-limited)

**Session Endpoints** (Template infrastructure):
- `POST /service/api/v1/sessions/issue` - Issue session token (NO middleware, creates session)
- `POST /service/api/v1/sessions/validate` - Validate session token (NO middleware, checks validity)
- `POST /browser/api/v1/sessions/issue` - Browser session issue
- `POST /browser/api/v1/sessions/validate` - Browser session validate

**Authentication Endpoints** (Template infrastructure):
- `POST /browser/api/v1/authn` - Browser user login (returns session token)
- `POST /service/api/v1/authn` - Service login (returns session token)

**Session vs Authentication Endpoints**:
- **Authentication Endpoints**: Verify credentials BEFORE session creation (username/password, client credentials, OAuth callbacks)
- **Session Endpoints**: Manage session lifecycle AFTER authentication succeeds (issue/validate/refresh/revoke)
- **Flow**: Authentication verifies identity → Session endpoints create/manage session token → Middleware validates session on subsequent requests
- **Why Separate**: Authentication may involve multiple steps (MFA, OAuth redirects), sessions are simple token operations

**Admin Endpoints** (AdminServer, 127.0.0.1:9090 ONLY):
- `GET /admin/api/v1/livez` - Liveness probe (lightweight, restart on failure)
- `GET /admin/api/v1/readyz` - Readiness probe (heavyweight with dependency checks, remove from LB on failure)
- `POST /admin/api/v1/shutdown` - Graceful shutdown trigger

**Join Request Admin Endpoints** (Template infrastructure):
- `GET /browser/api/v1/tenant/join-requests` - List pending join requests
- `POST /browser/api/v1/tenant/join-requests/:id/approve` - Approve join request
- `POST /browser/api/v1/tenant/join-requests/:id/reject` - Reject join request
- `GET /service/api/v1/tenant/join-requests` - Service variant of join requests

**Domain-Specific Routes**: Services register additional routes via `WithPublicRouteRegistration()` callback

### Template Usage Pattern

```go
// ServerBuilder creates all infrastructure
builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg)

// Register domain-specific migrations (2001+)
builder.WithDomainMigrations(domainMigrationsFS, "migrations")

// Register domain-specific routes
builder.WithPublicRouteRegistration(func(base, res) error {
    // Create domain repos
    domainRepo := cryptoutilDomainRepository.NewDomainRepository(res.DB)

    // Create domain server
    publicServer := cryptoutilDomainServer.NewPublicServer(base, domainRepo, ...)

    // Register routes
    publicServer.RegisterRoutes()
    return nil
})

// Build returns all infrastructure
resources, err := builder.Build()
```

### Per-Message Key Rotation Pattern

**Message Key Rotation** (MANDATORY for all services):
- **Pattern**: Rotate per message (new JWK for each message)
- **Rationale**: Most secure pattern - limits key exposure to single message
- **Overhead**: Higher computational cost acceptable for cryptographic security

**Message JWK Storage** (MANDATORY for all services):
- **Pattern**: Separate domain table for JWK storage (e.g., `message_jwks`)
- **Encryption**: JWK encrypted with Barrier service BEFORE storing, decrypted AFTER retrieving
- **Why NOT Barrier-only**: Domain table provides control over JWK lifecycle, metadata, rotation tracking
- **Template Integration**: Barrier service encrypts/decrypts, domain table manages persistence

### ServerBuilder Pattern

**Eliminates 260+ Lines Per Service**:
- **Usage**: `NewServerBuilder(ctx, cfg).WithDomainMigrations(...).WithPublicRouteRegistration(...).Build()`
- **Returns**: `ServiceResources` struct with initialized:
  - DB (*gorm.DB)
  - TelemetryService (*telemetry.TelemetryService)
  - JWKGenService (*jose.JWKGenService)
  - BarrierService (*barrier.BarrierService)
  - UnsealKeysService (*barrier.UnsealKeysService) - **ADDED** to expose unseal keys management
  - SessionManager (*business_logic.SessionManagerService)
  - RealmService (service.RealmService)
  - RealmRepository (repository.TenantRealmRepository)
  - Application (*server.Application)
  - ShutdownCore (func())
  - ShutdownContainer (func())

**Planned Refactoring** (Phase W):
- Move service/repository bootstrap from `server_builder.go` to `ApplicationCore.StartApplicationCore()`
- ServerBuilder focuses ONLY on HTTPS listeners and route registration
- ApplicationCore handles all business logic initialization (repos, services, dependencies)
- Improves separation of concerns and testability

**Merged Migrations Pattern**:
- Template migrations (1001-1999) + Domain migrations (2001+) combined via `mergedMigrations` fs.FS implementation
- golang-migrate validates ALL versions against single unified filesystem
- Template migrations:
  - 1001: Sessions tables
  - 1002: Barrier encryption keys
  - 1003: Realm tables
  - 1004: Multi-tenancy (tenants, tenant_realms)
  - 1005: Pending users (registration approval workflow)
- Domain migrations: 2001+ (service-specific, e.g., cipher-im message tables, jose JWK tables)

**NO Default Tenant**:
- Services start "cold" without any pre-created tenant/realm
- ALL tenants created via `/browser/api/v1/register` or `/service/api/v1/register` endpoints
- Tests use TestMain pattern: start server once per package, register test tenant via HTTP

### Migration Priority

**CRITICAL ORDER**:
1. **cipher-im FIRST** (validate template works, identify gaps)
2. jose-ja, pki-ca, identity-* (sequential, fix template issues as discovered)
3. **sm-kms LAST** (most complex, template must be mature)

**WHY**: cipher-im is first real service to use template, validates patterns work. Other services migrate after cipher-im proves template stability.

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

### Service Federation Architecture

**Service Discovery**:
- **Pattern**: Static YAML configuration (NO dynamic discovery)
- **Rationale**: Explicit configuration over implicit discovery, deterministic behavior
- **Config**:
  ```yaml
  federation:
    identity_url: "https://identity-authz:8180"  # Static service name (Docker) or FQDN (K8s)
    identity_enabled: true
    jose_url: "https://jose-ja:8280"
  ```
- **NO dynamic discovery**: Config file specifies all federated service URLs, restart required on changes
- **Docker Compose**: Use service names (e.g., `identity-authz:8180`)
- **Kubernetes**: Use FQDN (e.g., `identity-authz.cryptoutil-ns.svc.cluster.local:8180`)

**Federation Fallback Pattern**:
- **NO circuit breakers**: Static config for federated services (NOT dynamic health checks)
- **Fallback Mechanism**: Per-service database realms + per-service config realms
- **If federated service down**: Per-service realms ALWAYS available for operator fallback
- **Example**: If identity-authz down, jose-ja uses local database realm for emergency access
- **Configuration**:
  ```yaml
  realms:
    - type: federated
      provider: identity-authz
      url: "https://identity-authz:8180"
    - type: database
      name: local-fallback
      enabled: true  # Always enabled for operator access
    - type: config
      name: emergency-operators
      users: ["admin@localhost"]
  ```

---

## Multi-Tenancy Architecture

### Core Registration Pattern

**NO Default Tenant**: Services start "cold" without any pre-created tenant/realm. ALL tenants created via registration endpoints.

**Registration Flow**:

**NEW USER - Create New Tenant** (omit tenant_id):

```http
POST /browser/api/v1/register
{
    "username": "admin@example.com",
    "password": "securepass"
}
```

**Response**: HTTP 403 Forbidden (pending approval)
- User saved in `pending_users` table (NOT `users`)
- Username MUST be unique per tenant (constraint: UNIQUE(tenant_id, username) across `users` AND `pending_users`)
- Admin must approve via admin panel
- After approval: user moved to `users`, new tenant created automatically, realm created
- After rejection: user deleted from `pending_users`
- All API calls return HTTP 403 until approved, HTTP 401 if rejected
- Pending user expiration configurable in hours (NOT days), default 72 hours (3 days)
- NO session_token issued until user approved (removed from registration response)

**EXISTING USER - Join Existing Tenant** (specify tenant_id):

```http
POST /browser/api/v1/register
{
    "username": "user@example.com",
    "password": "securepass",
    "tenant_id": "uuid-of-existing-tenant"
}
```

**Response**: HTTP 403 Forbidden (pending approval)
- User saved in `pending_users` table with specified tenant_id
- Tenant admin must approve join request via admin panel
- After approval: user moved to `users`, granted access to tenant
- User receives NO session token until approved

**Admin Approval Endpoints**:
- `GET /browser/api/v1/tenant/join-requests` - List pending join requests
- `POST /browser/api/v1/tenant/join-requests/:id/approve` - Approve join request
- `POST /browser/api/v1/tenant/join-requests/:id/reject` - Reject join request
- `GET /service/api/v1/tenant/join-requests` - Service variant of join requests

**Rate Limiting**:
- Per IP address only (10 registrations per hour, configurable)
- In-memory (sync.Map) - simple, single-node, lost on restart
- Configurable thresholds with low defaults

**Email Validation**:
- NO email validation on username field (username can be non-email)
- Email/password authentication is a DIFFERENT realm (not implemented yet)
- Username field accepts any string (simplified registration flow)

### Realms (Authentication Only)

**Purpose**: Determines HOW user authenticates (password, OAuth, LDAP, WebAuthn)

**NOT for data isolation**: All realms in same tenant see SAME data

**Realm Types**:
- DB-based username/password (default)
- File-based username/password
- OAuth 2.0/OIDC federated
- LDAP/AD
- WebAuthn/FIDO2

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
- Opaque (UUIDv7) - stateful, server-side lookup

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
3. OAuth 2.0 (Google, Microsoft, GitHub, Facebook, Apple, LinkedIn, Twitter/X, Amazon, Okta)
4. TOTP (Authenticator App - Google Authenticator, Authy)
5. HOTP (Hardware Token - YubiKey, RSA SecurID)
6. WebAuthn with Passkeys (Face ID, Touch ID, Windows Hello)
7. WebAuthn without Passkeys (YubiKey, Titan Key)
8. Push Notification (Mobile app push-based)
9. Magic Link (Email/SMS)
10. Random OTP (Email/SMS/Voice)
11. Recovery Codes (Backup single-use codes)

**Total**: 13 headless methods + 28 browser methods (MFA combinations supported)

### OAuth 2.1 Identity Product

**Flow Priority** (identity-authz):
- **Authorization Code + PKCE**: Browser and native apps (modern, secure)
- **Client Credentials**: Service-to-service (simplest for cryptoutil internal)
- **BOTH required**: Cover all use cases (NOT client credentials only)

**Token Storage Configuration** (identity-authz):
- **Configurable**: Same as session tokens (stateful opaque, stateless JWE, stateless JWS)
- **Config**: `token_storage_mode: opaque|jwe|jws` in YAML
- **Stateful Opaque**: PostgreSQL/SQLite storage, instant revocation, slower validation
- **Stateless JWE**: Encrypted JWT, no storage, fast validation, delayed revocation (requires expiry)
- **Stateless JWS**: Signed JWT, no storage, fast validation, delayed revocation (requires expiry)

### Authorization Pattern

**Zero Trust**: ALWAYS re-evaluate permissions (NO caching of authz decisions)

**Scope-based**: OAuth 2.0 scopes (read:messages, write:messages, delete:messages)

**RBAC**: Role-based access control (admin, user, viewer)

**Resource-level ACLs**: Fine-grained permissions per resource

---

## Testing Strategy

### Testing Patterns - ARCHITECTURAL STANDARDS

**MANDATORY: ALL tests MUST follow these architectural patterns**

#### Table-Driven Test Pattern

**ALWAYS use table-driven tests for multiple test cases**

**Rationale**: Single test function per error category (not per error variant), easy to add new error cases (just add table row), reduced code duplication (~200 lines saved per consolidation), faster execution (shared setup runs once)

**Pattern**:
```go
func TestIssueSession_ValidationErrors(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name    string
        setup   func() context.Context
        wantErr string
    }{
        {name: "missing realm", setup: ctxWithoutRealm, wantErr: "realm"},
        {name: "missing tenant", setup: ctxWithoutTenant, wantErr: "tenant"},
        {name: "invalid request", setup: ctxWithInvalid, wantErr: "invalid"},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            // Test logic using tc
        })
    }
}
```

**FORBIDDEN**: ❌ Creating multiple standalone test functions for similar test cases (e.g., `TestFunc_Variant1`, `TestFunc_Variant2`, `TestFunc_Variant3`)

**Reference**: See `.github/instructions/03-02.testing.instructions.md` Section "FORBIDDEN #1: Standalone Test Functions for Variants"

#### app.Test() Pattern for HTTP Handlers

**ALL HTTP handler tests (unit and integration) MUST use Fiber's app.Test() for in-memory testing**

**Rationale**: In-memory testing is fast (<1ms), reliable, no network binding, prevents Windows Firewall popups (blocks CI/CD automation), TestMain ONLY for instance setup (NOT for handler tests)

**Pattern**:
```go
func TestHealthcheck_Handler(t *testing.T) {
    t.Parallel()

    // Create standalone Fiber app - NO listener started
    app := fiber.New(fiber.Config{DisableStartupMessage: true})
    app.Get("/admin/api/v1/livez", healthcheckHandler)

    req := httptest.NewRequest("GET", "/admin/api/v1/livez", nil)
    resp, err := app.Test(req, -1)  // ← In-memory, <1ms, no network binding
    require.NoError(t, err)
    defer resp.Body.Close()

    require.Equal(t, 200, resp.StatusCode)
}
```

**FORBIDDEN**: ❌ Starting real HTTPS servers or binding to network ports in unit/integration tests

**Reference**: See `.github/instructions/03-02.testing.instructions.md` Section "FORBIDDEN #2: Real HTTPS Listeners in Tests"

#### TestMain Pattern for Heavyweight Dependencies

**ALWAYS use TestMain to start heavyweight services once per package**

**Rationale**: PostgreSQL containers take 10-30s startup (do ONCE, not per test), shared resources (testDB, testServer) across all tests, fast test execution (no per-test overhead)

**Pattern**:
```go
var testDB *gorm.DB

func TestMain(m *testing.M) {
    container, _ := postgres.RunContainer(ctx, ...)
    defer container.Terminate(ctx)

    testDB, _ = gorm.Open(...)  // ← Created ONCE
    os.Exit(m.Run())
}

func TestSomething(t *testing.T) {
    // Use shared testDB - instant startup
}
```

**FORBIDDEN**: ❌ Creating database per test (repeated 10-30s overhead)

**Reference**: See `.github/instructions/03-02.testing.instructions.md` Section "FORBIDDEN #3: Per-Test Database Creation"

#### Test Isolation with t.Parallel()

**ALL test functions and subtests MUST use t.Parallel()**

**Rationale**: Reveals race conditions, deadlocks, data conflicts, if tests can't run concurrently production code can't either, faster test execution (utilizes all CPU cores)

**Pattern**:
```go
func TestSomething(t *testing.T) {
    t.Parallel()  // ← Parent test parallel
    tests := []struct{ ... }{ ... }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()  // ← Subtest parallel
        })
    }
}
```

**FORBIDDEN**: ❌ Omitting t.Parallel() from test functions or subtests

**Reference**: See `.github/instructions/03-02.testing.instructions.md` Section "FORBIDDEN #5: Missing t.Parallel() in Subtests"

#### Dynamic Test Data with UUIDv7

**ALWAYS use UUIDv7 for test data (NEVER hardcoded UUIDs/strings)**

**Rationale**: Thread-safe, process-safe, time-ordered, prevents UNIQUE constraint violations in parallel tests

**Pattern**:
```go
func TestCreate(t *testing.T) {
    t.Parallel()
    id := googleUuid.NewV7()  // ← Unique per test, generate ONCE and reuse
    user := &User{ID: id, Name: fmt.Sprintf("user_%s", id)}
    repo.Create(ctx, user)
}
```

**FORBIDDEN**: ❌ Hardcoded UUIDs/strings (causes UNIQUE constraint failures in parallel tests)

**Reference**: See `.github/instructions/03-02.testing.instructions.md` Section "FORBIDDEN #4: Hardcoded Test Data"

---

### Unit/Integration Tests

**Coverage**: ≥95% production, ≥98% infrastructure/utility

**Pattern**: Table-driven with `t.Parallel()` for orthogonal data

**Test Data**: UUIDv7 for ALL values (usernames, passwords, tenant IDs, realm IDs, messages) - differentiates all test data

**Dynamic Ports**: ALWAYS port 0 (prevents TIME_WAIT on Windows)

**SQLite Config**:
```go
// MaxOpenConns=5 (GORM transactions need separate connection)
sqlDB.SetMaxOpenConns(5)
sqlDB.Exec("PRAGMA journal_mode=WAL;")
sqlDB.Exec("PRAGMA busy_timeout = 30000;")
```

**TestMain Pattern** (ALL product-services MUST use):

```go
var testServer *Server

func TestMain(m *testing.M) {
    ctx := context.Background()

    // Create config (same method as CLIs)
    cfg := config.NewTestSettings()

    // Config triggers PostgreSQL test-container OR in-memory SQLite
    // and does migrations automatically
    testServer, err := NewFromConfig(ctx, cfg)
    if err != nil {
        panic(err)
    }

    // Start server
    go testServer.Start()
    defer testServer.Shutdown(ctx)

    // Wait for ready
    testServer.WaitForReady(ctx, 10*time.Second)

    // Register test tenant through API (realistic user flow)
    registerTestUser(testServer.PublicBaseURL())

    // Run all tests (share same server instance)
    os.Exit(m.Run())
}
```

**CRITICAL**: Same "start server from config" method as CLIs. Config determines PostgreSQL test-container OR in-memory SQLite.

**Isolation**: Each test creates unique data using UUIDv7 (usernames like `user_{uuid}@example.com`)

### E2E Tests

**Location**: Per product-service e2e/ subdirectory (NOT `test/e2e/` or inline with unit tests)
- Example: `internal/apps/cipher/im/e2e/` for cipher-im E2E tests

**MUST test BOTH paths**:
- `/service/**` (headless clients)
- `/browser/**` (browser clients)

**Pattern**: Docker Compose with PostgreSQL 18+ (production-like deployment)

**Database**: Uses PostgreSQL instance created OUTSIDE "create server from config" (external docker compose PostgreSQL 18+, NOT test-container)

**Registration**: Via HTTP endpoints (realistic user flow)

**NO demo tenants**: E2E tests create tenants via `/auth/register` endpoint

**Database Cleanup** (MANDATORY):
- **MANDATORY**: `docker compose down -v` at end of TestMain
- **Rationale**: Removes PostgreSQL volumes, prevents disk space exhaustion, ensures clean state for next package
- **Pattern**:
  ```go
  func TestMain(m *testing.M) {
      // Start docker compose
      cmd := exec.Command("docker", "compose", "-f", "e2e/compose.yml", "up", "-d")
      cmd.Run()

      // Run tests
      exitCode := m.Run()

      // CRITICAL: Cleanup volumes
      cleanup := exec.Command("docker", "compose", "-f", "e2e/compose.yml", "down", "-v")
      cleanup.Run()

      os.Exit(exitCode)
  }
  ```
- **NO tenant cleanup**: Rely on database cleanup (volumes removed by `-v` flag)

### Mutation Testing

**Mutation Testing Classification** (gremlins):

| Category | Packages | Minimum Score | Rationale |
|----------|----------|---------------|-----------|
| **Production Code** | `internal/jose/ja/service/`, `internal/jose/ja/server/apis/`, `internal/ca/service/`, `internal/identity/*/service/` | ≥85% | Business logic with acceptable gaps in error paths |
| **Infrastructure** | `internal/shared/*`, `pkg/*`, `internal/apps/template/`, `internal/cmd/cicd/*` | ≥98% | Reusable utilities must be bulletproof, used across all services |
| **Generated Code** | `api/client/*`, `api/server/*`, `api/model/*` (oapi-codegen), GORM models | EXEMPT | Auto-generated code is stable, mutation testing adds no value |

**Execution**:
- Per-package during implementation: `gremlins unleash ./internal/jose/ja/service`
- Parallel in CI: GitHub Actions matrix (4-6 packages per job, 15-minute timeout)

**Coverage Requirements** (MANDATORY):
- ≥95% production code mutation efficacy minimum (98% ideal)
- ≥98% infrastructure/utility code mutation efficacy
- EXEMPT generated code (OpenAPI stubs, protobuf, GORM models)

**Timing Targets**:
- <20 minutes total with parallel execution (4-6 jobs)
- Sequential would be 45+ minutes (avoid)

---

## Deployment Patterns

### Docker Compose

**CRITICAL: Docker Secrets Pattern** (see `deployments/kms/compose.yml` for reference implementation)

**NO HARDCODED SECRETS**: NEVER hardcode sensitive data in compose.yml or environment variables

**NO ENVIRONMENT VARIABLES for sensitive data**: NEVER use `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` environment variables

**ALWAYS Docker Secrets** with 440 permissions (r--r-----):

```yaml
secrets:
  postgres_username.secret:
    file: ./secrets/postgres_username.secret
  postgres_password.secret:
    file: ./secrets/postgres_password.secret
  postgres_database.secret:
    file: ./secrets/postgres_database.secret
  unseal_1of5.secret:
    file: ./secrets/unseal_1of5.secret

services:
  postgres:
    image: postgres:18
    environment:
      # Use _FILE suffix to read from Docker secrets
      POSTGRES_USER_FILE: /run/secrets/postgres_username.secret
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password.secret
      POSTGRES_DB_FILE: /run/secrets/postgres_database.secret
    secrets:
      - postgres_username.secret
      - postgres_password.secret
      - postgres_database.secret
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $(cat /run/secrets/postgres_username.secret) -d $(cat /run/secrets/postgres_database.secret)"]

  cipher-im:
    image: cryptoutil:local
    command: ["server", "start", "--config=/app/config/cipher-im.yml"]
    secrets:
      - unseal_1of5.secret  # CRITICAL: ALL services use SAME unseal secrets
      - postgres_url.secret
    ports: ["8888:8888"]
    depends_on:
      postgres: {condition: service_healthy}
      otel-collector: {condition: service_started}
    healthcheck:
      # Use wget (available in Alpine), NOT curl
      # Use 127.0.0.1 (NOT localhost - may resolve to ::1 IPv6)
      test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/api/v1/livez"]
      start_period: 60s
      interval: 5s

  otel-collector:
    image: otel/opentelemetry-collector-contrib:latest
    # NO host port mappings (use container-to-container networking)
    networks:
      - telemetry-network

  grafana-lgtm:
    image: grafana/otel-lgtm:latest
    ports: ["3000:3000"]  # Only expose Grafana UI to host
```

**Secret Files** (440 permissions):
```bash
chmod 440 secrets/*.secret
```

**Application Config** (`file://` pattern):
```yaml
database-url: "file:///run/secrets/postgres_url.secret"
unseal-keys:
  - "file:///run/secrets/unseal_1of5.secret"
  - "file:///run/secrets/unseal_2of5.secret"
hash-pepper: "file:///run/secrets/hash_pepper.secret"
```

**Health Checks**: Via admin endpoints (livez/readyz)

**Networking**: Container-to-container (use service names like `opentelemetry-collector-contrib:4317`, NOT host ports)

### Kubernetes

**Patterns**: StatefulSets for stateful services, Deployments for stateless

**Probes**:
```yaml
livenessProbe:
  httpGet: {path: /admin/api/v1/livez, port: 9090, scheme: HTTPS}
  initialDelaySeconds: 10
  periodSeconds: 10

readinessProbe:
  httpGet: {path: /admin/api/v1/readyz, port: 9090, scheme: HTTPS}
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
- ≥95%/≥98% efficacy (minimum/ideal production, infrastructure/utility)

**E2E Workflow**:
- Docker Compose deployment
- BOTH `/service/**` and `/browser/**` paths
- Health check validation

### Pre-Merge Requirements

**ALL MUST PASS**:
- [ ] Quality gates (linting, formatting)
- [ ] Unit tests (≥95%/≥98% coverage)
- [ ] Mutation tests (≥95%/≥98% efficacy minimum/ideal production, infrastructure/utility)
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
