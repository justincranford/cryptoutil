# cryptoutil Specification

## Overview

**cryptoutil** is a Go-based cryptographic services platform providing secure key management, identity services, and certificate authority capabilities with FIPS 140-3 compliance.

### Spec Kit Workflow Reference

**Source**: SPECKIT-CONFLICTS-ANALYSIS O8 answer C, 2025-12-19

This specification follows the Spec Kit methodology (see `docs/SPECKIT-QUICK-GUIDE.md` for complete workflow). Key principles:

- **Iterative Clarification**: Ambiguities resolved via `clarify.md` → `CLARIFY-QUIZME.md` cycle
- **Constitution Authority**: `.specify/memory/constitution.md` defines immutable principles and quality gates
- **Evidence-Based Completion**: Tasks require objective evidence (coverage ≥95%, mutation ≥85%, all tests passing)
- **Feedback Loops**: Implementation insights update earlier documents (spec → constitution → clarify)
- **Phase Dependencies**: Strict sequencing (Phase 1 foundation before Phase 2 core features)

**Key Document Relationships**:

- `constitution.md`: Core principles, absolute requirements, quality gates
- `spec.md` (this file): Product requirements, technical specification
- `plan.md`: 7-phase implementation approach, task breakdown
- `clarify.md`: Authoritative Q&A for all resolved ambiguities
- `implement/DETAILED.md`: Task status tracking, append-only timeline

## Technical Constraints

### CGO Ban - CRITICAL

**!!! CGO IS BANNED EXCEPT FOR RACE DETECTOR !!!**

- **CGO_ENABLED=0** MANDATORY for builds, tests, Docker, production
- **ONLY EXCEPTION**: Race detector workflow requires CGO_ENABLED=1 (Go toolchain limitation)
- **NEVER** use CGO-dependent packages (e.g., `github.com/mattn/go-sqlite3`)
- **ALWAYS** use CGO-free alternatives (e.g., `modernc.org/sqlite`)
- **Rationale**: Maximum portability, static linking, cross-compilation for production

---

## Service Architecture

### Overview

**cryptoutil** consists of 4 independent products that can be deployed standalone or as an integrated suite:

1. **P1: JOSE Authority** - Cryptographic primitives (JWK, JWS, JWE, JWT)
2. **P2: Identity Services** - 3 microservices (AuthZ, IdP, Resource Server)
3. **P3: KMS** - Hierarchical key management (library + optional server)
4. **P4: CA** - Certificate authority with EST/OCSP/CRL support

### Dual-Server Architecture Pattern

**CRITICAL**: All services implement a dual HTTPS endpoint pattern for security and operational separation.

#### Public HTTPS Server

**Purpose**: Browser-facing UI/APIs vs headless-client APIs, with different authentication options, authorization options, and middleware security
**Bind**: `<configurable_address>:<configurable_port>` (e.g., ports: 8080, 8081, 8082)
**Security**:

- MUST use HTTPS TLS 1.3+ with server certificate; TLS client-certificate authentication can also be enforced, but it is a configuration option
- Address binding constraints
  - Unit/integration tests: MUST use IPv4 127.0.0.1 by default; if not, it triggers Windows Firewall Exception prompts, which defeats the purpose of test automation
  - Docker Containers: MUST use IPv4 0.0.0.0 binding by default inside containers; no IPv6 due to dual stack issues inside Docker, and no 127.0.0.1 because Docker networking can't map external port to internal 127.0.0.1 network interface
- API Contexts are based on request paths
  - `/browser/swagger/*` - Browser-to-service Swagger UI; UI is secured using middleware security and injected JavaScript customization
  - `/browser/api/v1/*` - Browser-to-service APIs for Swagger UI or SPA UI invocation
  - `/service/api/v1/*` - Service-to-service APIs for headless clients
- Access to request paths MUST be mutually exclusive for different clients based on /browser vs /service prefix
  - Headless-based clients MUST use /service/* paths
  - Browser-based clients MUST use /browser/* paths
- Mutually exclusive configuration for /browser vs /service prefixes
  - Headless-based clients: Unique configuration of authentication, authorization, and middleware to be applied to all /service/* paths
  - Browser-based clients: Unique configuration of authentication, authorization, and middleware to be applied to all /browser/* paths
  - Shared middleware: CIDR and IP whitelisting, rate limiting, telemetry collection, request logging
- Mutually exclusive middleware, plus some shared middleware, will be used to enforce different security for /browser vs /service prefixes
  - Common middleware:
  - Headless-based only middleware: Authentication must identify client as non-browser client
  - Browser-based only middleware: Authentication must identify client as browser client, CORS/CSRF/CSP/XSS
- ALL Authentication methods must support two configurations
  - Two Realm types per Authentication method: File Realm Type (YAML), Database Realm Type (GORM/SQL)
  - Priority: File Realm Type (YAML) is higher priority than Database Realm Type (GORM/SQL); if DB access is down, File Realm Type supports Availability (from CIA Extended Triad) and Continuity of Business
  - Minimum one File Realm required per service, for Admin access in case of DB disaster; that minimum one File Realm can be any type that doesn't depend on DB availability (e.g. Basic Authorization header OK, WebAuthn/Passkeys/RandomOTP not OK because it requires persisting a challenge)
- Initial Request Without Session Token
  - Unauthenticated browser-based clients MUST be redirected to authentication, supporting a different option depending if a service (e.g. SM-KMS, PKI-CA, JOSE-JA) is deployed standalone vs federated with Identity product:
    - SFA in Standalone product mode: Basic (Username/password), Basic (Email/Password), Bearer (API Token)
    - MFA in Federated Identity mode: Basic (Username/password), Basic (Email/Password), Bearer (API Token), WebAuthn (without Passkeys), WebAuthn (with Passkeys), random OTP via email||SMS, HOTP/TOTP via registering an Authenticator app, magic link via email||SMS, HTTPS client certificate, opaque||JWE||JWS OAuth 2.1 Access Token, opaque OAuth 2.1 Refresh Token, and MORE TO BE CLARIFIED (CRITICAL: Must be clarified via /speckit.clarify and CLARIFY-QUIZME.md)
  - Unauthenticated browser-based clients MUST be redirected to authentication, supporting a different option depending if a service (e.g. SM-KMS, PKI-CA, JOSE-JA) is deployed standalone vs federated with Identity product:
    - SFA in Standalone product mode: Basic (Clientid,clientsecret), Bearer (API Token)
    - MFA in Federated Identity mode: Basic (Clientid,clientsecret), Bearer (API Token), HTTPS client certificate, opaque||JWE||JWS OAuth 2.1 Access Token, opaque OAuth 2.1 Refresh Token, and MORE TO BE CLARIFIED (CRITICAL: Must be clarified via /speckit.clarify and CLARIFY-QUIZME.md)
- Issuance of Session Token
  - Browser-based clients that successfully prove authentication will be given a session cookie (opaque||JWE|JWS non-OAuth 2.1)
  - Headless-based clients that successfully prove authentication will be given a session cookie (opaque||JWE|JWS non-OAuth 2.1)
  - A session cookie can always be used to identify the client type as either browser-based vs headless-based; client type is mutually exclusive, and must be one of the two values
- Subsequent Request With Session Token
  - middleware for /service/* paths MUST use the session cookie to validate the client is a non-browser client; browser type client will be rejected, no||expired session token triggers authentication redirection
  - middleware for /browser/* paths MUST use the session cookie to validate the client is a browser client; browser type client will be rejected, no||expired session token triggers authentication redirection

##### `/browser/api/v1/*` - Browser-Based Client APIs

**Authentication**:

- **Session Tokens**: HTTP Cookie-based session tokens (HttpOnly, Secure, SameSite=Strict)
- **OAuth 2.1 Flow**: Authorization Code + PKCE (Proof Key for Code Exchange)
- **Token Acquisition**: User redirected to IdP `/authorize` endpoint, exchanges code for session token
- **Token Storage**: Server-side session storage, client receives opaque cookie
- **Token Validation**: Server validates cookie against session store on each request

**Authorization**:

- **Scope Enforcement**: Session token contains user's granted scopes
- **Resource-Level Access Control**: Middleware checks scopes against endpoint requirements
- **User Context**: Full user profile available in request context (user ID, email, roles)
- **Consent Tracking**: Session tracks which scopes user explicitly consented to

**Middleware Pipeline** (Applied in order):

1. **CORS (Cross-Origin Resource Sharing)**: Validates Origin header against allowlist
2. **CSRF (Cross-Site Request Forgery) Protection**: Validates CSRF token in request header/body
3. **CSP (Content Security Policy)**: Sets strict Content-Security-Policy headers
4. **Session Cookie Validation**: Extracts and validates session token from Cookie header
5. **Session Store Lookup**: Retrieves session data from Redis/database
6. **Scope Authorization**: Checks user's scopes match endpoint requirements
7. **Rate Limiting**: Per-user rate limiting (100 req/min default)
8. **IP Allowlist**: Optional IP/CIDR allowlist enforcement
9. **Request Logging**: OTLP trace logging with user context

**Request Headers Required**:

- `Cookie: session_token=<opaque_session_id>`
- `X-CSRF-Token: <csrf_token>` (for non-GET requests)
- `Origin: https://allowed-origin.com` (for CORS preflight)

**Response Headers Set**:

- `Set-Cookie: session_token=...; HttpOnly; Secure; SameSite=Strict`
- `Access-Control-Allow-Origin: https://allowed-origin.com`
- `Access-Control-Allow-Credentials: true`
- `Content-Security-Policy: default-src 'self'; script-src 'self'`
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`

**Use Cases**:

- Single Page Applications (SPAs) - React, Vue, Angular
- Progressive Web Apps (PWAs)
- Mobile web browsers
- Any user-facing browser-based client

---

##### `/service/api/v1/*` - Service-to-Service APIs

**Authentication**:

- **Access Tokens**: HTTP Authorization Bearer tokens (JWT format)
- **OAuth 2.1 Flow**: Client Credentials flow (client_id + client_secret)
- **Token Acquisition**: Service POSTs to `/oauth/token` with credentials, receives JWT
- **Token Storage**: Service stores token in memory, refreshes on expiry
- **Token Validation**: Server validates JWT signature, expiry, issuer, audience

**Authorization**:

- **Scope Enforcement**: JWT contains client's granted scopes in `scope` claim
- **Client Context**: JWT contains client_id, no user context
- **Service-Level Access Control**: Middleware checks scopes against endpoint requirements
- **mTLS Optional**: Can require mutual TLS for additional authentication layer

**Middleware Pipeline** (Applied in order):

1. **Authorization Header Extraction**: Extracts Bearer token from Authorization header
2. **JWT Signature Validation**: Validates token signature against JWKs from `/oauth/jwks`
3. **JWT Claims Validation**: Checks exp, iss, aud, nbf claims
4. **Scope Authorization**: Checks token's scope claim matches endpoint requirements
5. **Rate Limiting**: Per-client rate limiting (1000 req/min default)
6. **IP Allowlist**: Optional IP/CIDR allowlist enforcement
7. **Request Logging**: OTLP trace logging with client_id context

**Request Headers Required**:

- `Authorization: Bearer <jwt_access_token>`
- `Content-Type: application/json` (for POST/PUT/PATCH)

**Response Headers Set**:

- `Cache-Control: no-store` (prevent token caching)
- `X-Content-Type-Options: nosniff`

**CORS/CSRF/CSP**: **NOT APPLIED** (service-to-service APIs don't need browser protections)

**Use Cases**:

- Backend microservices calling each other
- Serverless functions (AWS Lambda, Azure Functions)
- Scheduled jobs/cron tasks
- Internal automation scripts
- Third-party API integrations

---

##### Why Separate `/browser/*` vs `/service/*` Paths?

**Security Isolation**:

- Browser middleware (CORS/CSRF/CSP) would break service-to-service calls
- Service tokens (JWTs) are too long-lived for browser security model
- Session cookies require server-side state, impractical for high-volume service APIs
- Prevents accidental exposure of service tokens to browser clients

**Performance**:

- Service APIs skip unnecessary browser middleware (CORS preflight, CSRF validation)
- Browser APIs use lightweight session cookies instead of large JWTs
- Rate limits tuned differently (browsers: 100 req/min, services: 1000 req/min)

**Compliance**:

- Browser APIs track user consent for audit trails
- Service APIs track client_id for non-repudiation
- Separate logs for user actions vs automated service actions

**API Consistency**:

- Both paths serve **identical OpenAPI spec** (same endpoints, request/response schemas)
- Only authentication and middleware differ, not the API contract
- Clients choose path based on their runtime environment (browser vs backend)

#### Private HTTPS Server

**Purpose**: Internal admin tasks, health checks, metrics
**Bind**: `127.0.0.1:9090` (not externally accessible)
**Security**:

- IP restriction to localhost only
- Minimal middleware (no CORS/CSRF)
- Optional mTLS for production environments
- Not exposed in Docker port mappings

**Admin Port Assignments** (Source: SPECKIT-CONFLICTS-ANALYSIS C4 answer D, 2025-12-19):

- **KMS**: Admin port 9090 (all KMS instances share, bound to 127.0.0.1)
- **Identity**: Admin port 9091 (all 5 Identity services share)
- **CA**: Admin port 9092 (all CA instances share)
- **JOSE**: Admin port 9093 (all JOSE instances share)

**Admin API Context**:

- `/admin/v1/livez` - Liveness probe (process alive)
- `/admin/v1/readyz` - Readiness probe (dependencies healthy)
- `/admin/v1/healthz` - Combined health check
- `/admin/v1/metrics` - Prometheus metrics endpoint
- `/admin/v1/shutdown` - Graceful shutdown trigger

**Why Dual Servers?**:

1. **Security**: Admin endpoints not exposed to public network
2. **Performance**: Health probes don't compete with user traffic
3. **Reliability**: Kubernetes/Docker health checks work even if public API is overloaded
4. **Compliance**: Separation of concerns for audit requirements

### Service Mesh Topology

```
┌─────────────────────────────────────────────────────────────────┐
│                     External Clients                             │
│              (Browsers, Mobile Apps, Services)                   │
└───────────────────────────┬─────────────────────────────────────┘
                            │ HTTPS (TLS 1.3+)
                            │ OAuth 2.1 Tokens
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│                     Reverse Proxy / API Gateway                  │
│                  (Traefik, nginx, Kong - optional)               │
└───────────┬───────────────┬───────────────┬─────────────────────┘
            │               │               │
    ┌───────▼────┐  ┌───────▼────┐  ┌──────▼──────┐  ┌──────────┐
    │   JOSE     │  │  Identity  │  │    KMS      │  │    CA    │
    │ Authority  │  │  Services  │  │   Server    │  │  Server  │
    │            │  │            │  │             │  │          │
    │ Port: 8280 │  │AuthZ: 8180 │  │ Port: 8080  │  │Port: 8380│
    │            │  │ IdP: 8181  │  │             │  │          │
    │            │  │  RS: 8182  │  │             │  │          │
    │            │  │  RP: 8183  │  │             │  │          │
    │            │  │ SPA: 8184  │  │             │  │          │
    └─────┬──────┘  └─────┬──────┘  └──────┬──────┘  └────┬─────┘
          │               │                │              │
          │ Admin:9093    │ Admin:9091     │ Admin:9090   │Admin:9092
          │ (127.0.0.1)   │ (127.0.0.1)    │ (127.0.0.1)  │(127.0.0.1)
          │               │                │              │
    ┌─────▼───────────────▼────────────────▼──────────────▼─────┐
    │              Kubernetes / Docker Health Checks            │
    │         (Liveness, Readiness via Private Endpoints)       │
    └───────────────────────────────────────────────────────────┘
          │               │                │              │
    ┌─────▼───────────────▼────────────────▼──────────────▼─────┐
    │                    PostgreSQL Database                     │
    │         (Shared for dev, isolated per service in prod)    │
    └───────────────────────────────────────────────────────────┘
          │               │                │              │
    ┌─────▼───────────────▼────────────────▼──────────────▼─────┐
    │              OpenTelemetry Collector                       │
    │         (Traces, Metrics, Logs → Grafana LGTM)            │
    └───────────────────────────────────────────────────────────┘
```

### Network Segmentation

| Network Zone | Services | Access Control |
|--------------|----------|----------------|
| **Public** | All services (ports 8080-8443) | OAuth 2.1 tokens, rate limiting, TLS 1.3+ |
| **Admin** | All services (port 9090) | Localhost only (127.0.0.1), optional mTLS |
| **Database** | PostgreSQL (port 5432) | Password auth, network isolation |
| **Telemetry** | OTLP Collector (ports 4317/4318) | Service mesh only, no external |

### Unified Command Interface - CRITICAL

**MANDATORY**: All services MUST be accessible via unified `cryptoutil` command:

```bash
# KMS (✅ COMPLETE - reference implementation)
cryptoutil kms start --config=kms.yml
cryptoutil kms status

# Identity (⚠️ PARTIAL - admin servers exist, needs cmd integration)
cryptoutil identity start --config=identity.yml
cryptoutil identity status

# JOSE (❌ BLOCKED - no admin server, no cmd integration)
cryptoutil jose start --config=jose.yml
cryptoutil jose status

# CA (❌ BLOCKED - no admin server, no cmd integration)
cryptoutil ca start --config=ca.yml
cryptoutil ca status
```

**Current Implementation Status**:

| Service | Admin Server | Port 9090 | Cmd Integration | Status |
|---------|--------------|-----------|-----------------|--------|
| KMS | ✅ Complete | ✅ Yes | ✅ internal/cmd/cryptoutil/kms | ✅ REFERENCE |
| Identity AuthZ | ✅ Exists | ✅ Yes | ❌ NO | ⚠️ PARTIAL |
| Identity IdP | ✅ Exists | ✅ Yes | ❌ NO | ⚠️ PARTIAL |
| Identity RS | ✅ Exists | ✅ Yes | ❌ NO | ⚠️ PARTIAL |
| JOSE | ❌ Missing | ❌ NO | ❌ NO | ❌ BLOCKED |
| CA | ❌ Missing | ❌ NO | ❌ NO | ❌ BLOCKED |

**Phase 3.5 Deliverable**: All services follow KMS pattern with dual-server architecture and unified command interface.

---

## Product Suite

### P1: JOSE (JSON Object Signing and Encryption)

Core cryptographic primitives for web security standards. Serves as the embedded foundation for all other products AND as a standalone JOSE Authority service.

**Architecture**:

- **Embedded Library**: JOSE primitives in `internal/jose/` used by P2/P3/P4
- **Standalone Service**: JOSE Authority service exposing REST API for external applications

**Current State**: JOSE primitives exist in `internal/common/crypto/jose/`. Iteration 2 refactors to `internal/jose/` as standalone authority.

#### Capabilities

| Feature | Description | Status |
|---------|-------------|--------|
| JWK | JSON Web Key generation and management | ✅ Implemented |
| JWKS | JSON Web Key Set endpoints | ✅ Implemented |
| JWE | JSON Web Encryption operations | ✅ Implemented |
| JWS | JSON Web Signature operations | ✅ Implemented |
| JWT | JSON Web Token creation and validation | ✅ Implemented |
| JOSE Authority | Standalone JOSE service with full API | ✅ Implemented |

#### JOSE Authority API (Iteration 2 - COMPLETE)

| Endpoint | Method | Description | Status |
|----------|--------|-------------|--------|
| `/jose/v1/keys` | POST | Generate new JWK | ✅ Implemented |
| `/jose/v1/keys/{kid}` | GET | Retrieve specific JWK | ✅ Implemented |
| `/jose/v1/keys` | GET | List JWKs with filters | ✅ Implemented |
| `/jose/v1/jwks` | GET | Public JWKS endpoint | ✅ Implemented |
| `/jose/v1/sign` | POST | Create JWS signature | ✅ Implemented |
| `/jose/v1/verify` | POST | Verify JWS signature | ✅ Implemented |
| `/jose/v1/encrypt` | POST | Create JWE encryption | ✅ Implemented |
| `/jose/v1/decrypt` | POST | Decrypt JWE payload | ✅ Implemented |
| `/jose/v1/jwt/issue` | POST | Issue JWT with claims | ✅ Implemented |
| `/jose/v1/jwt/validate` | POST | Validate JWT signature and claims | ✅ Implemented |

#### Supported Algorithms

| Algorithm Type | Algorithms | FIPS Status |
|----------------|-----------|-------------|
| Signing | PS256, PS384, PS512, RS256, RS384, RS512, ES256, ES384, ES512, EdDSA | ✅ Approved |
| Key Wrapping | RSA-OAEP, RSA-OAEP-256, A128KW, A192KW, A256KW | ✅ Approved |
| Content Encryption | A128GCM, A192GCM, A256GCM, A128CBC-HS256, A192CBC-HS384, A256CBC-HS512 | ✅ Approved |
| Key Agreement | ECDH-ES, ECDH-ES+A128KW, ECDH-ES+A192KW, ECDH-ES+A256KW | ✅ Approved |

---

### P2: Identity (OAuth 2.1 Authorization Server + OIDC IdP)

**Architecture**: 3 independent microservices that can be deployed standalone or together:

1. **AuthZ Server**: OAuth 2.1 Authorization Server (port 8080, admin 9090)
2. **IdP Server**: OIDC Identity Provider (port 8081, admin 9090)
3. **Resource Server**: Protected API with token validation (port 8082, admin 9090)

Each service has its own Docker image (`Dockerfile.authz`, `Dockerfile.idp`, `Dockerfile.rs`) and can scale independently.

**Priority Focus**: Login/Consent UI (minimal HTML, server-rendered, no JavaScript).

#### Authorization Server (AuthZ)

| Endpoint | Method | Description | Status |
|----------|--------|-------------|--------|
| `/oauth2/v1/authorize` | GET/POST | Authorization code flow with mandatory PKCE | ✅ Working |
| `/oauth2/v1/token` | POST | Token exchange (code, refresh, client_credentials) | ✅ Working |
| `/oauth2/v1/introspect` | POST | Token introspection (RFC 7662) | ✅ Working |
| `/oauth2/v1/revoke` | POST | Token revocation (RFC 7009) | ✅ Working |
| `/oauth2/v1/clients/{id}/rotate-secret` | POST | Administrative Rotate client secret with grace period | ✅ Implemented |
| `/.well-known/openid-configuration` | GET | OpenID Connect Discovery | ✅ Working |
| `/.well-known/jwks.json` | GET | JSON Web Key Set | ✅ Working |
| `/.well-known/oauth-authorization-server` | GET | OAuth 2.1 Authorization Server Metadata (RFC 8414) | ✅ Working |
| `/device_authorization` | POST | Device Authorization Grant (RFC 8628) | ✅ Implemented (backend complete - 18 tests passing) |
| `/par` | POST | Pushed Authorization Requests (RFC 9126) | ✅ Implemented (backend complete - 16 tests passing) |

#### Identity Provider (IdP)

| Endpoint | Method | Description | Status |
|----------|--------|-------------|--------|
| `/oidc/v1/login` | GET/POST | User authentication | ✅ Working (HTML form rendered, session created) |
| `/oidc/v1/consent` | GET/POST | User consent for scopes | ✅ Working (HTML form rendered, consent recorded) |
| `/oidc/v1/logout` | GET/POST | Session termination | ✅ Working (session/token cleared) |
| `/oidc/v1/endsession` | GET | OpenID Connect End Session (RP-Initiated Logout) | ✅ Working |
| `/oidc/v1/userinfo` | GET | User information endpoint | ✅ Working (claims returned per scopes, JWT-signed optional) |
| `/oidc/v1/mfa/enroll` | POST | Administrative Enroll MFA factor | ✅ Implemented (10 tests passing - backend complete) |
| `/oidc/v1/mfa/factors` | GET | Administrative List user MFA factors | ✅ Implemented (10 tests passing - backend complete) |
| `/oidc/v1/mfa/factors/{id}` | DELETE | Administrative Remove MFA factor | ✅ Implemented (10 tests passing - backend complete) |

#### Authentication Methods

| Method | Description | Status |
|--------|-------------|--------|
| client_secret_basic | HTTP Basic Auth with client_id:client_secret | ✅ Working |
| client_secret_post | client_id and client_secret in request body | ✅ Working |
| client_secret_jwt | JWT signed with client secret (RFC 7523 Section 3) | ✅ 100% (jti replay protection via jti_replay_cache table, 10-minute assertion lifetime validation, 10 tests passing) |
| private_key_jwt | JWT signed with private key (RFC 7523 Section 3) | ✅ 100% (jti replay protection, 10-minute assertion lifetime validation, JWKS support, 7 tests passing) |
| tls_client_auth | Mutual TLS client certificate authentication | ✅ 100% (CA certificate validation, subject DN matching, SHA-256 fingerprint verification, revocation checking, 6 tests passing) |
| self_signed_tls_client_auth | Self-signed TLS client certificate authentication | ✅ 100% (self-signed cert validation, subject DN matching, SHA-256 fingerprint verification, 6 tests passing) |
| session_cookie | Browser session cookie for SPA UI | ✅ 100% (HybridAuthMiddleware with session validation, SessionRepository with 11 tests passing, session expiration/revocation support) |

#### MFA Factors

| Factor | Description | Status | Priority |
|--------|-------------|--------|----------|
| Passkey | WebAuthn/FIDO2 authentication | ✅ Working | HIGHEST |
| TOTP | Time-based One-Time Password | ✅ Working | HIGH |
| Hardware Security Keys | Dedicated hardware tokens (U2F/FIDO) | ✅ 100% (WebAuthn/FIDO2 cross-platform authenticators, AAGUID identification, sign counter for replay prevention, 15+ tests passing) | HIGH |
| Email OTP | One-time password via email | ✅ 100% (EmailOTPService with MockEmailService for testing, RateLimiter (5 OTPs per 10 min), bcrypt hashing, 10 tests passing: SendOTP, VerifyOTP_Success/InvalidCode/AlreadyUsed/Expired, RateLimit, domain model tests) | MEDIUM |
| SMS OTP | One-time password via SMS | ✅ 100% (SMSOTPAuthenticator with MockSMSProvider for testing, RateLimiter integration, phone number validation, 12 tests passing: NewAuthenticator, Method, InitiateAuth with user/phone validation, VerifyAuth, ChallengeNotFound, unit/E2E flows) | LOW (NIST deprecated but MANDATORY) |
| HOTP | HMAC-based One-Time Password (counter-based) | ✅ 100% (RFC 4226 compliant, counter synchronization, lookahead window, 12 tests passing) | LOW |
| Recovery Codes | Backup codes for account recovery | ✅ 100% (10-code generation, single-use validation, secure hashing, 13 tests passing) | MEDIUM |
| Push Notifications | Push-based authentication via mobile app | ✅ 100% (PushNotificationAuthenticator with device token management, approval token generation, push notification delivery, 6 tests passing) | LOW |
| Phone Call OTP | One-time password via voice call | ✅ 100% (PhoneCallOTPAuthenticator with voice call delivery, OTP speech formatting, retry limit enforcement, 6 tests passing) | LOW |

#### Secret Rotation System

| Feature | Description | Status |
|---------|-------------|--------|
| ClientSecretVersion | Multiple secret versions per client | ✅ Implemented |
| Grace Period | Configurable overlap for rotation | ✅ Implemented |
| KeyRotationEvent | Audit trail for rotation events | ✅ Implemented |
| Scheduled Rotation | Automated rotation workflows | ✅ Implemented |
| NIST SP 800-57 | Compliance with key lifecycle standards | ✅ Demonstrated |

---

### P3: KMS (Key Management Service)

Hierarchical key management with versioning and rotation.

**Deployment Modes**:

- **Embedded Library**: KMS operations via `internal/kms/` package (used by Identity, CA, JOSE)
- **Standalone Server** (Planned): REST API server via `cmd/kms-server/` (not yet implemented)
- **Current Access**: Demo integration code in `internal/cmd/demo/integration.go`

**Authentication Strategy**: Configurable - support multiple methods including OAuth 2.1 federation to Identity (P2), API key, mTLS. Dual API exposure:

- `/browser/api/v1/` - User-to-browser APIs for SPA invocation
- `/service/api/v1/` - Service-to-service APIs

**Realm Configuration**: MANDATORY configurable realms for users and clients (file-based and database-based), with OPTIONAL federation to external IdPs and AuthZs.

**Docker Compose Deployment**: 3 instances in production config:

- `kms-sqlite` (port 8080): In-memory SQLite backend for development
- `kms-postgres-1` (port 8081): PostgreSQL backend instance 1
- `kms-postgres-2` (port 8082): PostgreSQL backend instance 2

**Rationale**: Fixed instances vs replicas to demonstrate multi-backend support and database-specific configurations.

#### ElasticKey Operations

| Operation | Method | Endpoint | Status |
|-----------|--------|----------|--------|
| Create | POST | `/elastickey` | ✅ Implemented |
| Read | GET | `/elastickey/{elasticKeyID}` | ✅ Implemented |
| List | GET | `/elastickeys` | ✅ Implemented |
| Update | PUT | `/elastickey/{elasticKeyID}` | ✅ Implemented (11 tests passing - mapper unit tests) |
| Delete | DELETE | `/elastickey/{elasticKeyID}` | ✅ Implemented (11 tests passing - mapper unit tests, soft delete) |

#### MaterialKey Operations

| Operation | Method | Endpoint | Status |
|-----------|--------|----------|--------|
| Create | POST | `/elastickey/{elasticKeyID}/materialkey` | ✅ Implemented |
| Read | GET | `/elastickey/{elasticKeyID}/materialkey/{materialKeyID}` | ✅ Implemented |
| List | GET | `/elastickey/{elasticKeyID}/materialkeys` | ✅ Implemented |
| Global List | GET | `/materialkeys` | ✅ Implemented |
| Import | POST | `/elastickey/{elasticKeyID}/import` | ✅ Implemented (10 tests passing - mapper unit tests) |
| Revoke | POST | `/elastickey/{elasticKeyID}/materialkey/{materialKeyID}/revoke` | ✅ Implemented (10 tests passing - mapper unit tests) |

#### Cryptographic Operations

| Operation | Method | Endpoint | Status |
|-----------|--------|----------|--------|
| Generate | POST | `/elastickey/{elasticKeyID}/generate` | ✅ Implemented |
| Encrypt | POST | `/elastickey/{elasticKeyID}/encrypt` | ✅ Implemented |
| Decrypt | POST | `/elastickey/{elasticKeyID}/decrypt` | ✅ Implemented |
| Sign | POST | `/elastickey/{elasticKeyID}/sign` | ✅ Implemented |
| Verify | POST | `/elastickey/{elasticKeyID}/verify` | ✅ Implemented |

#### Key Hierarchy

```
Unseal Secrets (file:///run/secrets/* or Yubikey)
    ↓
Root Keys (derived from unseal secrets)
    ↓
Intermediate Keys (per-tenant isolation)
    ↓
ElasticKey (policy container)
    ↓
MaterialKey (versioned key material)
```

#### Filtering Parameters

| Parameter | Description |
|-----------|-------------|
| `elastic_key_ids` | Filter by elastic key UUIDs |
| `names` | Filter by key names |
| `providers` | Filter by key providers |
| `algorithms` | Filter by algorithms |
| `statuses` | Filter by statuses (active, suspended, deleted) |
| `versioning_allowed` | Filter by versioning policy |
| `import_allowed` | Filter by import policy |
| `sorts` | Sorting criteria (name, created_at, updated_at, status) |
| `page_number` | Page number for pagination |
| `page_size` | Page size for pagination |

#### MaterialKey Filtering Parameters

| Parameter | Description |
|-----------|-------------|
| `material_key_ids` | Filter by material key UUIDs |
| `elastic_key_ids` | Filter by parent elastic key UUIDs (global list) |
| `minimum_generate_date` | Filter by minimum generation date |
| `maximum_generate_date` | Filter by maximum generation date |

#### Sorting Parameters

| Parameter | Direction |
|-----------|-----------|
| `name` | asc/desc |
| `created_at` | asc/desc |
| `updated_at` | asc/desc |
| `status` | asc/desc |

---

### P4: Certificates (Certificate Authority)

**Source**: SPECKIT-CONFLICTS-ANALYSIS C7 answer A, 2025-12-19

X.509 certificate lifecycle management with CA/Browser Forum compliance.

**Deployment Architecture**:

- **3-instance deployment pattern** (matches KMS/JOSE/Identity pattern for consistency)
- **ca-sqlite**: Port 8380 (public API), Port 9092 (admin), SQLite backend
- **ca-postgres-1**: Port 8381 (public API), Port 9092 (admin), PostgreSQL backend
- **ca-postgres-2**: Port 8382 (public API), Port 9092 (admin), PostgreSQL backend
- **Admin ports bound to 127.0.0.1** (not externally accessible, health checks only)

#### Implementation Status

| Task | Description | Priority | Status |
|------|-------------|----------|--------|
| 1. Domain Charter | Scope definition, compliance mapping | HIGH | ✅ Complete |
| 2. Config Schema | YAML schema for crypto, subject, certificate profiles | HIGH | ✅ Complete |
| 3. Crypto Providers | RSA, ECDSA, EdDSA, ECDH, EdDH, HMAC, future PQC | HIGH | ✅ Complete |
| 4. Subject Profile Engine | Template resolution for subject details, SANs | HIGH | ✅ Complete |
| 5. Certificate Profile Engine | 25+ profile archetypes | HIGH | ✅ Complete |
| 6. Root CA Bootstrap | Offline root CA creation | HIGH | ✅ Complete |
| 7. Intermediate CA Provisioning | Subordinate CA hierarchy | HIGH | ✅ Complete |
| 8. Issuing CA Lifecycle | Rotation, monitoring, status reporting | MEDIUM | ✅ Complete |
| 9. Enrollment API | EST API for CSR or CRMF submission, issuance | HIGH | ✅ Complete |
| 10. Revocation Services | CRL generation, OCSP responders | HIGH | ✅ Complete |
| 11. Time-Stamping | RFC 3161 TSA functionality | MEDIUM | ✅ Complete |
| 12. RA Workflows | Registration authority for validation | MEDIUM | ✅ Complete |
| 13. Profile Library | 24 predefined certificate profiles | HIGH | ✅ Complete |
| 14. Storage Layer | PostgreSQL/SQLite with ACID guarantees | HIGH | ✅ Complete |
| 15. CLI Tooling | bootstrap, issuance, revocation commands | MEDIUM | ✅ Complete |
| 16. Observability | OTLP metrics, tracing, audit logging | MEDIUM | ✅ Complete |
| 17. Security Hardening | STRIDE threat modeling, security validation | HIGH | ✅ Complete |
| 18. Compliance | CA/Browser Forum audit readiness | HIGH | ✅ Complete |
| 19. Deployment | Docker Compose, Kubernetes manifests | MEDIUM | ✅ Complete |
| 20. Handover | Documentation, runbooks | LOW | ✅ Complete |

**Implementation Progress**: 20/20 internal tasks complete (100%)

#### CA Server REST API (Iteration 2 - COMPLETE)

The CA Server exposes certificate lifecycle operations via REST API with mTLS authentication.

| Endpoint | Method | Description | Status |
|----------|--------|-------------|--------|
| `/ca/v1/health` | GET | Health check endpoint | ✅ Implemented |
| `/ca/v1/ca` | GET | List available CAs | ✅ Implemented |
| `/ca/v1/ca/{ca_id}` | GET | Get CA details and certificate chain | ✅ Implemented |
| `/ca/v1/ca/{ca_id}/crl` | GET | Download current CRL | ✅ Implemented |
| `/ca/v1/certificate` | POST | Issue certificate from CSR | ✅ Implemented |
| `/ca/v1/certificate/{serial}` | GET | Retrieve certificate by serial | ✅ Implemented |
| `/ca/v1/certificate/{serial}/revoke` | POST | Revoke certificate | ✅ Implemented |
| `/ca/v1/certificate/{serial}/status` | GET | Get certificate status | ✅ Implemented |
| `/ca/v1/ocsp` | POST | OCSP responder endpoint | ✅ Implemented |
| `/ca/v1/profiles` | GET | List certificate profiles | ✅ Implemented |
| `/ca/v1/profiles/{profile_id}` | GET | Get profile details | ✅ Implemented |
| `/ca/v1/est/cacerts` | GET | EST: Get CA certificates | ✅ Implemented |
| `/ca/v1/est/simpleenroll` | POST | EST: Simple enrollment | ✅ Implemented |
| `/ca/v1/est/simplereenroll` | POST | EST: Re-enrollment | ✅ Implemented |
| `/ca/v1/est/serverkeygen` | POST | EST: Server-side key generation | ✅ Implemented |
| `/ca/v1/tsa/timestamp` | POST | RFC 3161 timestamp request | ✅ Implemented |

**API Authentication Methods:**

- **mTLS**: Client certificate authentication (primary)
- **JWT Bearer**: For delegated access from Identity Server
- **API Key**: For automated systems (with IP allowlist)

**API Progress**: 16/16 endpoints implemented (100% complete)

#### Compliance Requirements

| Standard | Requirement |
|----------|-------------|
| RFC 5280 | X.509 certificate format and validation |
| RFC 6960 | OCSP protocol for certificate status |
| RFC 7030 | EST (Enrollment over Secure Transport) |
| RFC 3161 | Time-Stamp Protocol (TSP) |
| CA/Browser Forum | Baseline Requirements for TLS Server Certificates |
| Serial Numbers | ≥64 bits CSPRNG, non-sequential, >0, <2^159 |
| Validity Period | Maximum 398 days for subscriber certificates |
| Signature Algorithms | RSA ≥ 2048 bits, ECDSA P-256/P-384/P-521 |

---

## Infrastructure Components

### I1: Configuration

- YAML files and CLI flags (no environment variables for secrets)
- Validation on startup
- Feature flags support

### I2: Networking

- HTTPS with TLS 1.3+ minimum
- HTTP/2 support via Fiber framework
- CORS, CSRF protection
- Rate limiting per IP

### I3: Testing

- Table-driven tests with `t.Parallel()`
- Coverage targets: 95% production, 100% infrastructure, 100% utility
- Mutation testing: ≥98% gremlins score per package
- Fuzz testing, benchmark testing, integration testing

#### Test Execution Performance

**Requirements**:

- Individual package test time: <30 seconds per package
- Total test suite execution time: <100 seconds
- Race detector run: <200 seconds (slower due to CGO_ENABLED=1 overhead)

**Current Status**: Performance varies by package - optimization needed for slower packages.

#### Load Testing Coverage

**Implemented**:

- ✅ Service API (`/service/api/v1/*`): Gatling simulation exists (`test/load/src/test/java/cryptoutil/ServiceApiSimulation.java`)

**Missing**:

- ❌ Browser API (`/browser/api/v1/*`): No Gatling simulation
- ❌ Admin API (`/admin/v1/*`): No Gatling simulation
- ❌ Multi-product integration: No cross-service workflow tests

**Required**: Create `BrowserApiSimulation.java` and `AdminApiSimulation.java` to complete load test coverage.

#### E2E Test Scope

**Current**: Basic Docker Compose lifecycle tests (`internal/test/e2e/e2e_test.go`)

- Service startup/shutdown
- Health check connectivity
- Container log collection

**Missing Critical Workflows**:

- OAuth 2.1 authorization code flow (browser → AuthZ → IdP → consent → token)
- Certificate issuance and revocation (CSR → CA → certificate → CRL/OCSP)
- KMS key generation, encryption/decryption, and rotation
- JOSE token signing and verification workflows

**Required**: Expand E2E tests to cover end-to-end product workflows, not just infrastructure.

### CA Networking

- HTTPS with TLS 1.3+ minimum
- HTTP/2 support via Fiber framework
- CORS, CSRF protection
- Rate limiting per IP
- ACME protocol support for automated certificate issuance
- OCSP responder endpoints
- CRL distribution points

### CA Testing

- Table-driven tests with `t.Parallel()`
- Coverage targets: 95% production, 100% infrastructure, 100% utility
- Mutation testing: ≥98% gremlins score per package
- Certificate chain validation testing
- OCSP responder testing
- CRL generation testing
- ACME protocol testing

### I4: Performance

- Gatling load tests in `test/load/`
- Connection pooling
- Concurrent key generation pools

### I5: Telemetry

- OpenTelemetry instrumentation
- OTLP export to collector
- Grafana dashboards (Loki, Tempo, Prometheus)

### I6: Crypto

- FIPS 140-3 compliant algorithms
- Key generation pools (keygen package)
- Deterministic key derivation for interoperability

### I7: Database

- PostgreSQL (production/development/testing)
- SQLite (development/testing)
- GORM ORM with migrations
- WAL mode, busy_timeout for SQLite concurrency
- **GitHub Actions Dependency**: ALL workflows running `go test` MUST include PostgreSQL service container

#### PostgreSQL Service Requirements for CI/CD

**MANDATORY**: Any GitHub Actions workflow executing `go test` on packages using database repositories MUST configure PostgreSQL service container:

```yaml
services:
  postgres:
    image: postgres:18
    env:
      POSTGRES_DB: cryptoutil_test
      POSTGRES_PASSWORD: cryptoutil_test_password
      POSTGRES_USER: cryptoutil
    options: >-
      --health-cmd pg_isready
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
    ports:
      - 5432:5432
```

**Why Required**:

- Tests in `internal/kms/server/repository/sqlrepository` require PostgreSQL
- Tests in `internal/identity/domain/repository` require PostgreSQL
- Without service: Tests fail with "connection refused" after 2.5s timeout
- With service: PostgreSQL ready before tests start (50s startup window)

**Affected Workflows**: ci-race, ci-mutation, ci-coverage, any workflow running database tests

### I8: Containers

- Docker Compose deployments
- Service mesh: cryptoutil, postgres, otel-collector, grafana-otel-lgtm
- Health checks via wget (Alpine containers)

### I9: Deployment

- GitHub Actions CI/CD
- Act for local workflow testing
- Multi-stage Docker builds with static linking

#### CI/CD Workflow Inventory

| Workflow | Trigger | Duration Target | PostgreSQL Required | Purpose |
|----------|---------|-----------------|---------------------|---------|
| `ci-quality` | Push, PR | <5 min | ❌ | Linting, formatting, build validation |
| `ci-coverage` | Push, PR | <10 min | ✅ | Test coverage analysis (≥95% target) |
| `ci-race` | Push, PR | <15 min | ✅ | Race condition detection (CGO_ENABLED=1) |
| `ci-mutation` | Push, PR | <45 min | ✅ | Mutation testing (≥98% efficacy) |
| `ci-benchmark` | Push, PR | <10 min | ❌ | Performance benchmarks |
| `ci-fuzz` | Push, PR | <10 min | ❌ | Fuzz testing (keygen, digests, parsers) |
| `ci-sast` | Push, PR | <5 min | ❌ | Static security analysis (gosec) |
| `ci-gitleaks` | Push, PR | <2 min | ❌ | Secrets scanning |
| `ci-dast` | Push, PR | <15 min | ❌ | Dynamic security testing (Nuclei, ZAP) |
| `ci-e2e` | Push, PR | <20 min | ❌ | End-to-end Docker Compose tests |
| `ci-load` | Push, PR | <30 min | ❌ | Load testing (Gatling - Service API only) |
| `ci-identity-validation` | Push, PR | <5 min | ✅ | Identity-specific validation tests |
| `release` | Tag | <15 min | ❌ | Build and publish release artifacts |

**Total CI Feedback Loop Target**: <10 minutes for critical path (quality + coverage + race)
**Full Suite Target**: <60 minutes for all workflows to complete

**Health Check Pattern Standardization**:

- **Alpine containers**: Use `wget --no-check-certificate -q -O /dev/null <url>`
- **Non-Alpine containers**: Use `curl -k -f -s <url>`
- **Retry logic**: `start_period: 10s`, `interval: 5s`, `retries: 5`, `timeout: 5s`
- **Admin endpoints**: All services use `https://127.0.0.1:9090/admin/v1/livez` for Docker health checks

---

## Quality Requirements

### Code Coverage Targets

| Category | Target | Current |
|----------|--------|---------|
| Production Code | ≥95% | Varies |
| Infrastructure (cicd) | ≥100% | ~90% |
| Utility Code | ≥100% | ~100% |

### Mutation Testing Requirements

- Minimum ≥80% gremlins score per package
- Focus on business logic, parsers, validators, crypto operations
- Track improvements in baseline reports

### Linting Requirements

- golangci-lint v2.6.2+
- gofumpt (not gofmt)
- All linters enabled, no `//nolint:` exceptions without justification
- UTF-8 without BOM for all files

### File Size Limits

| Threshold | Lines | Action |
|-----------|-------|--------|
| Soft | 300 | Warning |
| Medium | 400 | Review required |
| Hard | 500 | Refactor required |

---

## Service Endpoints Summary

### Docker Compose Services

#### P1: JOSE Services

| Service | Public Port | Admin Port | Backend |
|---------|-------------|------------|---------|
| jose-sqlite | 8080 | 9090 | SQLite in-memory |
| jose-postgres-1 | 8081 | 9090 | PostgreSQL |
| jose-postgres-2 | 8082 | 9090 | PostgreSQL |

#### P2: Identity Services

**Note**: Identity consists of 3 independent microservices, each with its own admin endpoint.

| Service | Public Port | Admin Port | Backend | Status |
|---------|-------------|------------|---------|--------|
| identity-authz | 8080 | 9090 (planned) | SQLite/PostgreSQL | ⚠️ Admin API not yet implemented |
| identity-idp | 8081 | 9090 (planned) | SQLite/PostgreSQL | ⚠️ Admin API not yet implemented |
| identity-rs | 8082 | 9090 (planned) | SQLite/PostgreSQL | ⚠️ Admin API not yet implemented |

**Current Status**: Identity services use `/health` on public port. Migration to dual-server pattern (like KMS) is planned.

#### P3: KMS Services

| Service | Public Port | Admin Port | Backend |
|---------|-------------|------------|---------|
| kms-sqlite | 8080 | 9090 | SQLite in-memory |
| kms-postgres-1 | 8081 | 9090 | PostgreSQL |
| kms-postgres-2 | 8082 | 9090 | PostgreSQL |

#### P4: CA Services

**Note**: CA deployment incomplete - only development config exists.

| Service | Public Port | Admin Port | Backend | Status |
|---------|-------------|------------|---------|--------|
| ca-simple | 8443 | 9443 | SQLite | ✅ Development only (`compose.simple.yml`) |
| ca-postgres-1 | 8443 (planned) | 9443 (planned) | PostgreSQL | ⚠️ Production config missing |
| ca-postgres-2 | 8443 (planned) | 9443 (planned) | PostgreSQL | ⚠️ Production config missing |

**Required**: Create `deployments/ca/compose.yml` with multi-instance PostgreSQL deployment matching JOSE/KMS patterns.

#### Common Infrastructure Services

| Service | Public Port | Admin Port | Backend |
|---------|-------------|------------|---------|
| postgres | 5432 | - | - |
| otel-collector | 4317/4318 | 13133 | - |
| otel-collector-health | - | - | Health monitoring |
| secrets-test | - | - | Secrets validation |
| grafana-otel-lgtm | 3000 | - | Loki/Tempo/Prometheus |

### Health Endpoints

#### Private Admin API (`https://127.0.0.1:9090`)

Used for internal monitoring and health checks.

**CRITICAL**: All services MUST use `https://127.0.0.1:9090` for private admin APIs (not exposed externally).

| Product | Endpoint | Purpose |
|---------|----------|---------|
| JOSE | `/admin/v1/livez` | Liveness probe |
| JOSE | `/admin/v1/readyz` | Readiness probe |
| JOSE | `/admin/v1/healthz` | Combined health check |
| Identity | `/admin/v1/livez` | Liveness probe |
| Identity | `/admin/v1/readyz` | Readiness probe |
| Identity | `/admin/v1/healthz` | Combined health check |
| KMS | `/admin/v1/livez` | Liveness probe |
| KMS | `/admin/v1/readyz` | Readiness probe |
| KMS | `/admin/v1/healthz` | Combined health check |
| CA | `/admin/v1/livez` | Liveness probe (planned) |
| CA | `/admin/v1/readyz` | Readiness probe (planned) |
| CA | `/admin/v1/healthz` | Combined health check (planned) |

#### Public Browser-to-Service API

Used by browsers and external clients.

| Product | Endpoint | Purpose |
|---------|----------|---------|
| JOSE | `/health` | Public health check |
| JOSE | `/ui/swagger/doc.json` | OpenAPI specification |
| Identity | `/health` | Public health check |
| Identity | `/ui/swagger/doc.json` | OpenAPI specification |
| KMS | `/health` | Public health check |
| KMS | `/ui/swagger/doc.json` | OpenAPI specification |
| CA | `/health` | Public health check (planned) |
| CA | `/ui/swagger/doc.json` | OpenAPI specification (planned) |

#### Public Service-to-Service API

Used by other services for health checks.

| Product | Endpoint | Purpose |
|---------|----------|---------|
| JOSE | `/health` | Service health check |
| Identity | `/health` | Service health check |
| KMS | `/health` | Service health check |
| CA | `/health` | Service health check (planned) |

---

## Future Architecture Enhancements

### Hash Service Refactoring (Phase 5)

**Source**: SPECKIT-CONFLICTS-ANALYSIS Q5.1 answer E, Q5.2 answer A, 2025-12-19

**Goal**: Create unified hash service architecture supporting 4 hash registry types with version management.

**Version Architecture**:

- **Version = Date-Based Policy Revision**: v1 (2020 NIST), v2 (2023 NIST), v3 (2025 OWASP)
- **Algorithm Selection Within Version**: Input size-based (0-31→SHA-256, 32-47→SHA-384, 48+→SHA-512)
- **4 Registries × 3 Versions = 12 Configurations**: Each registry supports v1/v2/v3
- **Output Format**: Prefix `{v}:base64_hash` (e.g., `{1}:abcd1234...`, `{2}:efgh5678...`)
- **Verification**: Automatically tries all versions until match found (backward compatibility)

**Architecture**:

```
HashService
├── LowEntropyRandomHashRegistry (PBKDF2-based)
│   ├── v1: 0-31 bytes → PBKDF2-HMAC-SHA256 (OWASP rounds)
│   ├── v2: 32-47 bytes → PBKDF2-HMAC-SHA384
│   └── v3: 48+ bytes → PBKDF2-HMAC-SHA512
├── LowEntropyDeterministicHashRegistry (PBKDF2-based, no salt)
│   ├── v1: 0-31 bytes → PBKDF2-HMAC-SHA256
│   ├── v2: 32-47 bytes → PBKDF2-HMAC-SHA384
│   └── v3: 48+ bytes → PBKDF2-HMAC-SHA512
├── HighEntropyRandomHashRegistry (HKDF-based)
│   ├── v1: 0-31 bytes → HKDF-HMAC-SHA256
│   ├── v2: 32-47 bytes → HKDF-HMAC-SHA384
│   └── v3: 48+ bytes → HKDF-HMAC-SHA512
└── HighEntropyDeterministicHashRegistry (HKDF-based, no salt)
    ├── v1: 0-31 bytes → HKDF-HMAC-SHA256
    ├── v2: 32-47 bytes → HKDF-HMAC-SHA384
    └── v3: 48+ bytes → HKDF-HMAC-SHA512
```

**Registry API** (consistent across all 4 types):

- `HashWithLatest(input []byte) (string, error)` - Uses current version
- `HashWithVersion(input []byte, version int) (string, error)` - Uses specific version
- `Verify(input []byte, hashed string) (bool, error)` - Verifies against any version

**Hash Output Format**: Includes version metadata for version-aware verification

**Version Selection**: Automatic based on input size ranges (0-31, 32-47, 48+ bytes)

**Use Cases**:

- **Low Entropy Random**: Password hashing (PBKDF2, salted)
- **Low Entropy Deterministic**: Replay-resistant tokens (PBKDF2, no salt)
- **High Entropy Random**: Key derivation from high-entropy inputs (HKDF, salted)
- **High Entropy Deterministic**: Deterministic key derivation (HKDF, no salt)

**Benefits**:

- Version management supports algorithm upgrades without breaking existing hashes
- Consistent API across all hash types reduces implementation complexity
- Input size-based version selection automates algorithm selection
- FIPS 140-3 compliant (PBKDF2, HKDF, HMAC-SHA256/384/512)

---

### Service Template Extraction (Phase 6)

**Goal**: Extract reusable service template from KMS server, augment for all 8 PRODUCT-SERVICE instances.

**8 PRODUCT-SERVICE Target Instances**:

1. **sm-kms** - Secrets Manager - Key Management System
2. **pki-ca** - Public Key Infrastructure - Certificate Authority
3. **jose-ja** - JOSE - JWK Authority
4. **identity-authz** - Identity - Authorization Server
5. **identity-idp** - Identity - Identity Provider
6. **identity-rs** - Identity - Resource Server
7. **identity-rp** - Identity - Relying Party (BFF pattern)
8. **identity-spa** - Identity - Single Page Application (static hosting)

**Common Patterns** (extracted from KMS):

- **Dual HTTPS Servers**: Public API (0.0.0.0:configurable) + Admin API (127.0.0.1:9090)
- **Dual API Paths**: `/browser/api/v1/*` (session-based) vs `/service/api/v1/*` (token-based)
- **Middleware Pipeline**: CORS/CSRF/CSP (browser-only), rate limiting, IP allowlist, authentication
- **Database Abstraction**: PostgreSQL + SQLite dual support with GORM
- **OpenTelemetry Integration**: OTLP traces, metrics, logs
- **Health Check Endpoints**: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz`
- **Graceful Shutdown**: `/admin/v1/shutdown` endpoint

**Service-Specific Customization Points**:

- **API Endpoints**: Custom OpenAPI specs per service
- **Business Logic Handlers**: Service-specific request processing
- **Database Schemas**: Custom GORM models per service
- **Client SDK Generation**: Service-specific client interfaces
- **Barrier Services**: Optional (KMS-specific, not needed for other services)

**Template Packages**:

```
internal/template/
├── server/          # ServerTemplate base class
│   ├── dual_https.go       # Public + Admin server management
│   ├── router.go           # Route registration framework
│   ├── middleware.go       # Pipeline builder (CORS/CSRF/CSP/rate limit)
│   └── lifecycle.go        # Start/stop/reload lifecycle
├── client/          # ClientSDK base class
│   ├── http_client.go      # HTTP client with mTLS/retry
│   ├── auth.go             # OAuth 2.1/mTLS/API key strategies
│   └── codegen.go          # OpenAPI-based client generation
└── repository/      # Database abstraction
    ├── dual_db.go          # PostgreSQL + SQLite support
    ├── gorm_patterns.go    # Model registration, migrations
    └── transaction.go      # Transaction handling patterns
```

**Parameterization Strategy**:

- **Constructor Injection**: Pass handlers, middleware, config at initialization
- **Interface-Based Customization**: Services implement `ServerInterface`
- **Configuration-Driven**: YAML config specifies behavior (CORS origins, rate limits, etc.)
- **Runtime Discovery**: Service registers capabilities dynamically

**Benefits**:

- **Faster Service Development**: Copy-paste-modify instead of build from scratch
- **Consistency**: All services use same infrastructure patterns
- **Maintainability**: Single source of truth for common patterns
- **Quality**: Reuse well-tested, production-hardened components

---

### Learn-PS Demonstration Service (Phase 7)

**Goal**: Create working Pet Store service using service template, validate reusability and completeness.

**Learn-PS Overview**:

- **Product**: Learn (educational/demonstration product)
- **Service**: PS (Pet Store service)
- **Purpose**: Copy-paste-modify starting point for customers creating new services
- **Scope**: Complete CRUD API for pet store (pets, orders, customers)

**API Endpoints** (via `/browser/api/v1/*` and `/service/api/v1/*`):

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/pets` | POST | Create new pet | OAuth 2.1 (write:pets scope) |
| `/pets` | GET | List pets (paginated) | OAuth 2.1 (read:pets scope) |
| `/pets/{id}` | GET | Get pet details | OAuth 2.1 (read:pets scope) |
| `/pets/{id}` | PUT | Update pet | OAuth 2.1 (write:pets scope) |
| `/pets/{id}` | DELETE | Delete pet | OAuth 2.1 (admin:pets scope) |
| `/orders` | POST | Create order | OAuth 2.1 (write:orders scope) |
| `/orders` | GET | List orders | OAuth 2.1 (read:orders scope) |
| `/orders/{id}` | GET | Get order details | OAuth 2.1 (read:orders scope) |
| `/customers` | POST | Create customer | OAuth 2.1 (write:customers scope) |
| `/customers` | GET | List customers | OAuth 2.1 (read:customers scope) |
| `/customers/{id}` | GET | Get customer details | OAuth 2.1 (read:customers scope) |

**Database Schema**:

```sql
-- Pets table
CREATE TABLE pets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    species TEXT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Customers table
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Orders table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    total DECIMAL(10,2) NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'cancelled')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Order items table
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    pet_id UUID NOT NULL REFERENCES pets(id),
    quantity INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Service Template Usage Example**:

```go
// main.go
func main() {
    // 1. Instantiate ServerTemplate
    template := server.NewServerTemplate(server.Config{
        PublicPort: 8080,
        AdminPort: 9090,
        EnableBarrier: false, // No barrier services needed
    })

    // 2. Register API routes
    template.RegisterPublicRoutes(func(r fiber.Router) {
        r.Post("/pets", handlers.CreatePet)
        r.Get("/pets", handlers.ListPets)
        r.Get("/pets/:id", handlers.GetPet)
        r.Put("/pets/:id", handlers.UpdatePet)
        r.Delete("/pets/:id", handlers.DeletePet)
        // ... orders, customers
    })

    // 3. Apply middleware
    template.ApplyMiddleware(middleware.Config{
        CORS: middleware.CORSConfig{
            Origins: []string{"https://learn-ps.example.com"},
        },
        RateLimit: middleware.RateLimitConfig{
            RequestsPerMinute: 100,
        },
    })

    // 4. Start servers
    template.Start(context.Background())
}
```

**Documentation Deliverables**:

1. **README.md**: Quick start, API docs, development guide
2. **Tutorial Series**: 4-part series (using, understanding, customizing, deploying)
3. **Video Demonstration**: Service startup, API usage, code walkthrough

**Quality Targets**:

- 95%+ test coverage (production code)
- 98%+ mutation efficacy
- ≤12s test execution time
- Passes all CI/CD workflows

**Customer Value**:

- **Working Example**: See service template in action
- **Starting Point**: Copy entire Learn-PS directory, modify for use case
- **Best Practices**: Learn production-ready patterns (error handling, testing, deployment)
- **API Design**: Reference implementation for REST API design

---

## Known Gaps and Future Work

### High Priority

1. **Identity Admin API Migration**: Implement dual-server pattern (Public HTTPS + Private HTTPS) matching KMS architecture
   - Add `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz` endpoints
   - Update Docker Compose health checks
   - Update all test files and workflows

2. **CA Production Deployment**: Create `deployments/ca/compose.yml` with multi-instance PostgreSQL deployment

3. **Load Test Coverage**: Implement missing Gatling simulations for Browser API and Admin API

4. **E2E Workflow Tests**: Expand beyond health checks to test complete product workflows (OAuth flows, certificate lifecycle, KMS operations)

### Medium Priority

1. **KMS Standalone Server**: Create `cmd/kms-server/main.go` for standalone deployment (currently library-only)

2. **JOSE Admin API**: Verify and document private server implementation for admin endpoints

3. **Runbook Library**: Create incident response, backup/restore, and key rotation runbooks

4. **Health Check Standardization**: Audit all Docker Compose files for consistent retry logic and patterns

### Low Priority

1. **Fuzz Testing Expansion**: Add fuzzing for JWT validation, certificate parsing, OAuth token introspection

2. **CA Operational Documentation**: Create enrollment workflow guides and profile selection matrix

3. **Workflow Execution Metrics**: Implement timing instrumentation and alerting for slow workflows

---

*Specification Version: 1.2.0*
*Last Updated: December 11, 2025*
