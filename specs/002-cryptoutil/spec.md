# cryptoutil Specification

## Overview

**cryptoutil** is a Go-based cryptographic services platform providing secure key management, identity services, and certificate authority capabilities with FIPS 140-3 compliance.

## Spec Kit Methodology

<https://github.com/github/spec-kit>

The Spec Kit Methodology, driven by GitHub's open-source toolkit, is a Spec-Driven Development (SDD) process that uses AI to transform detailed specifications into code, emphasizing structure, clarity, and controlled scope for better quality AI-assisted development. It moves beyond "vibe coding" by breaking features into manageable tasks.

### Spec Kit Steps

| Step | Output | Notes |
|------|--------|-------|
| 1. /speckit.constitution | .specify\memory\constitution.md | |
| 2. /speckit.specify | specs\002-cryptoutil\spec.md | |
| 3. /speckit.clarify | specs\002-cryptoutil\clarify.md and specs\002-cryptoutil\CLARIFY-QUIZME.md | (optional: after specify, before plan) |
| 4. /speckit.plan | specs\002-cryptoutil\plan.md | |
| 5. /speckit.tasks | specs\002-cryptoutil\tasks.md | |
| 6. /speckit.analyze | specs\002-cryptoutil\analyze.md | (optional: after tasks, before implement) |
| 7. /speckit.implement | (e.g., implement/DETAILED.md and implement/EXECUTIVE.md) | |

### Spec Kit Customizations

- **Phase Dependencies**: Strict sequencing (Phase 1 → Phase 2 → Phase 3, etc.)
- **Progress Tracking**: implement/DETAILED.md and implement/EXECUTIVE.md
- **Feedback Loops**: insights from implement/DETAILED.md and implement/EXECUTIVE.md are applied to earlier documents (implement → constitution+spec → clarify)
- **Evidence-Based Completion**: Implementation requires objective evidence (coverage ≥95%, mutation ≥85%, all tests passing)

## Technical Constraints

### CGO Ban - CRITICAL

**!!! CGO IS BANNED EXCEPT FOR RACE DETECTOR !!!**

- **CGO_ENABLED=0** MANDATORY for builds, tests, Docker, production
- **ONLY EXCEPTION**: Race detector workflow requires CGO_ENABLED=1 (Go toolchain limitation)
- **NEVER** use CGO-dependent packages (e.g., `github.com/mattn/go-sqlite3`)
- **ALWAYS** use CGO-free alternatives (e.g., `modernc.org/sqlite`)
- **Rationale**: Maximum portability, static linking, cross-compilation for production

**Race Detector Limitations** (Source: SPECKIT-CLARIFY-QUIZME-05 Q11, 2025-12-24):

**CRITICAL**: Go race detector is PROBABILISTIC - not all race conditions are guaranteed to be detected.

- **Execution-dependent**: Race detection depends on timing and scheduling during test execution
- **False negatives possible**: Passing race detector does NOT guarantee absence of race conditions
- **Best effort**: Run race detector on EVERY test execution to maximize detection probability
- **Complement with**: Code review, static analysis (e.g., `go vet`), and stress testing

**CI Race Detector Workflow**:

```yaml
# .github/workflows/ci-race.yml
- name: Run race detector
  run: |
    # Run MULTIPLE times to increase detection probability
    for i in {1..3}; do
      echo "Race detector run $i/3"
      go test -race -count=1 ./...
    done
  env:
    CGO_ENABLED: 1  # Required for race detector
```

**Rationale**: Accept probabilistic nature, run frequently to maximize coverage over time.

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q11

---

## Service Architecture

### Overview

**cryptoutil** consists of 4 products with 8 services. Products can be deployed standalone or as an integrated suite:

### Products and Services

| Service Alias | Product | Service | Public Ports | Admin Port | Description |
|---------------|-----------|-------------|------------|-------------|
| **sm-kms** | Secrets Manager | Key Management Service (KMS) | 8080-8089 | 127.0.0.1:9090 | REST APIs for per-tenant Elastic Keys |
| **pki-ca** | Public Key Infrastructure | Certificate Authority (CA) | 8050-8059 | 127.0.0.1:9090 | X.509 certificate lifecycle, EST, SCEP, OCSP, CRLDP, CMPv2, CMC, time-stamping |
| **jose-ja** | JOSE | JWK Authority (JA) | 8060-8069 | 127.0.0.1:9090 | JWK, JWKS, JWE, JWS, JWT operations |
| **identity-authz** | Identity | Authorization Server (authz) | 18000-18009 | 127.0.0.1:9090 | OAuth 2.1 authorization server, OIDC Discovery |
| **identity-idp** | Identity | Identity Provider (IdP) | 18100-18109 | 127.0.0.1:9090 | OIDC 1.0 authentication, login/consent UI, MFA enrollment |
| **identity-rs** | Identity | Resource Server (RS) | 18200-18209 | 127.0.0.1:9090 | Protected API with token validation (reference implementation) |
| **identity-rp** | Identity | Relying Party (RP) | 18300-18309 | 127.0.0.1:9090 | Backend-for-Frontend pattern (reference implementation) |
| **identity-spa** | Identity  Single Page Application (SPA) | 18400-18409 | 127.0.0.1:9090 | Static hosting for SPA clients (reference implementation) |

| Service Alias | Product | Service | Public Port | Admin Port | Description |
|---------------|-----------|-------------|------------|-------------|
| **cipher-im** | Cipher | InstantMessenger | 8070-8079 | 127.0.0.1:9090 | Encrypted messaging demonstration service validating service template |

**Source**: Architecture instructions (01-01.architecture.instructions.md), constitution.md Section I

## Product Suite Architecture - CRITICAL

### Dual-Endpoint Architecture Pattern

**MANDATORY: All Services in All Products MUST support run as containers; this is preferred for production and end-to-end testing**

**MANDATORY: All Services in All Products MUST support Two HTTPS Endpoints**

Separate HTTPS endpoints for public operations vs private administration MUST be supported. TLS server certificate authentication MUST be enforced for both endpoints; TLS client certificate authentication may be enabled per endpoint, and set to either preferred or required, via configuration; HTTP is NEVER allowed.

#### Deployment Environments

**Production Deployments**:

- Public endpoints MUST use 0.0.0.0 IPv4 bind address inside containers (enables external access)
- Public endpoints MAY use configurable IPv4 or IPv6 bind address outside containers (defaults to 127.0.0.1)
- Private endpoints MUST use 127.0.0.1:9090 inside containers (not mapped outside)
- No IPv6 inside containers: All endpoints must use IPv4 inside containers, due to dual-stack limitations in container runtimes (e.g. Docker Desktop for Windows)

**Development/Test Environments**:

For address binding:

- Public and private endpoints MUST use 127.0.0.1 IPv4 bind address (prevents Windows Firewall prompts)
- Rationale: 0.0.0.0 binding triggers Windows Firewall exception prompts, blocking automated execution of tests

For port binding:

- Public and private endpoints MUST use port 0 (dynamic allocation) to avoid port collisions
- Rationale: static ports cause port collisions during parallel test automation

**Admin Port Isolation** (Unified Deployments):

- Admin ports (127.0.0.1:9090) REQUIRE containerization for multi-service deployments
- Each container has isolated localhost namespace, preventing port collisions
- Non-containerized unified deployments NOT SUPPORTED
- Rationale: Multiple services using 127.0.0.1:9090 would collide on shared localhost without container isolation

#### CA Architecture Pattern

##### TLS Issuing CA Configurations

Examples of Issuing CA Cert Chains, in order of highest to lowest preference:

- Offline Root CA -> Online Root CA -> Online Issuing CA
- Online Root CA -> Online Issuing CA
- Online Root CA
- Online Root CA -> Policy Root CA -> Online Issuing CA
- Offline Root CA -> Online Root CA -> Policy CA -> Online Issuing CA

#### Two Endpoint HTTPS Architecture Pattern

##### TLS Certificate Configuration

All services in all products MUST support separate configuration for two HTTPS endpoints, including separate configuration for TLS Server and TLS client of each endpoint.

These main options MUST be supported via configuration from the point of view (POV) of the HTTPS Issuing CA certificate chain:

1. All Externally; useful for production, where HTTPS Issuing CA certificate chain is provided without private key, and the HTTPS Server certificate chain is provided with private key
2. Mixed External and Auto-Generated; where Issuing CA chain is provided with private key, and HTTPS Server certificate and private key are generated and signed by the Issuing CA; useful for per-product development and testing
3. All Auto-Generated; where HTTPS Server certificate chain and private key are all generated by the service instance; useful for standalone service development and testing, not production

HTTPS Issuing CA certificate chains for TLS Server MAY BE per-suite, per-product, or per-service type.

HTTPS Issuing CA certificate chains for TLS Client MUST BE per-service type (preferred).

**Static TLS Certificates** (Externally Provided):

These main options MUST be supported via configuration from the point of view (POV) of the HTTPS Server certificate chain and private key:

- Private keys stored in Docker Secrets (production and development)
- Certificate chains provided via file paths or PEM-encoded data in configuration files
- Trusted CA certificates configurable for client verification
- Required for production deployments with organizational PKI

HTTPS Issuing CA for TLS Server Certs SHOULD BE shared across all products (preferred), all services of a product, per-service instance type, or per-service instance.

HTTPS Issuing CA for TLS Client Certs MUST BE shared per per-service instance type.

#### Public HTTPS Endpoint

**Purpose**: Offers two public facing sets of APIs: browser-facing UI/APIs for browser-clients, and headless APIs for non-browser clients. Each set of APIs has different authentication options, authorization options, and middleware security

**Terminology Clarification** (Two HTTPS Endpoint Strategy):

- **Admin Port**: Non-exposed, ALWAYS 127.0.0.1:9090 (never configurable)
- **Exported Port**:
  - **Inside Container**: 0.0.0.0:8080 (container default, enables external access)
  - **Outside Container**:
    - **Tests**: Mapped to 127.0.0.1 with unique static port range per service (prevents Windows Firewall prompts, avoids port collisions)
    - **Production**: Mapped to `<configurable_address>:<configurable_port>` with unique static port range per service

**Bind**: `<configurable_address>:<configurable_port>` (e.g., ports: 8080, 8081, 8082 for sm-kms service instances)
**Security**:

- Address binding constraints
  - Unit/integration tests: MUST use IPv4 127.0.0.1; if not, Windows Firewall Exception prompts are triggered, which defeats the purpose of test automation
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
    - MFA in Federated Identity mode: 27 total authentication methods (see Authentication and Authorization Requirements section for complete list including WebAuthn, TOTP, HOTP, Magic Link, Random OTP, Social Login)
  - Unauthenticated headless-based clients MUST be redirected to authentication, supporting a different option depending if a service (e.g. SM-KMS, PKI-CA, JOSE-JA) is deployed standalone vs federated with Identity product:
    - SFA in Standalone product mode: Basic (Clientid,clientsecret), Bearer (API Token)
    - MFA in Federated Identity mode: 10 total authentication methods (see Authentication and Authorization Requirements section for complete list including mTLS, OAuth 2.1 Access/Refresh Tokens)
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
- **Introspection Caching**: Cache positive results with configurable TTL (default: token TTL), cache negative results for 1 minute
  - Positive results (active=true): TTL configured per deployment (default: match token expiry)
  - Negative results (active=false): Fixed 1-minute cache to prevent abuse
  - Provides operational flexibility while maintaining security

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

---

### Service Federation and Discovery

**CRITICAL**: Services MUST support configurable federation for cross-service communication.

#### Federation Architecture

Services discover and communicate with other cryptoutil services via **configuration** (NEVER hardcoded URLs). Example KMS federation configuration:

```yaml
# KMS federation configuration example
federation:
  # Identity service for OAuth 2.1 authentication
  identity_url: "https://identity-authz:8180"
  identity_enabled: true
  identity_timeout: 10s  # Per-service timeout (MANDATORY)

  # JOSE service for external JWE/JWS operations
  jose_url: "https://jose-server:8280"
  jose_enabled: true
  jose_timeout: 15s      # Different timeout per service (MANDATORY)

  # CA service for TLS certificate operations
  ca_url: "https://ca-server:8380"
  ca_enabled: false  # Optional - KMS can use internal TLS certs
  ca_timeout: 30s        # Longer timeout for CA operations (MANDATORY)

# Graceful degradation settings
federation_fallback:
  # When identity service unavailable
  identity_fallback_mode: "local_validation"  # Options: "reject_all", "allow_all" (dev only)

  # When JOSE service unavailable
  jose_fallback_mode: "internal_crypto"  # Use internal JWE/JWS implementation

  # When CA service unavailable
  ca_fallback_mode: "self_signed"  # Generate self-signed TLS certs
```

#### Session Migration During Federation Transitions

When a service transitions from non-federated (standalone) to federated mode, existing sessions MUST be handled via **grace period dual-format support**:

- **Grace Period**: Accept BOTH old and new token formats during transition (e.g., 24 hours)
- **Natural Expiration**: Old-format tokens expire according to their TTL (no forced invalidation)
- **New Issuance**: New logins immediately receive new-format tokens
- **Configuration Example**:

```yaml
federation:
  migration:
    grace_period_hours: 24  # Accept both formats during transition
    old_format_enabled: true  # Temporary backward compatibility
    new_format_enabled: true  # Forward compatibility
```

**Rationale**: Prevents service disruption by allowing gradual token migration without forcing user re-authentication.

#### Service Discovery Mechanisms

**1. Configuration File** (Preferred for static deployments):

```yaml
# Explicit URLs in config.yaml
federation:
  identity_url: "https://identity.example.com:8180"
  jose_url: "https://jose.example.com:8280"
```

**2. Docker Compose Service Names**:

```yaml
# Docker networks provide DNS resolution
federation:
  identity_url: "https://identity-authz:8180"  # Service name from compose.yml
  jose_url: "https://jose-server:8280"
```

**3. Kubernetes Service Discovery**:

```yaml
# Kubernetes DNS provides service resolution
federation:
  identity_url: "https://identity-authz.cryptoutil-ns.svc.cluster.local:8180"
  jose_url: "https://jose-server.cryptoutil-ns.svc.cluster.local:8280"
```

**DNS Caching** (Source: SPECKIT-CLARIFY-QUIZME-05 Q18, 2025-12-24):

**MANDATORY**: DNS lookups for federated services MUST NOT be cached - perform lookup on EVERY request.

**Rationale**: Kubernetes service endpoints change dynamically (pod restarts, scaling), stale DNS cache causes request failures.

**Implementation**:

```go
// Disable DNS caching for HTTP client
dialer := &net.Dialer{
    Timeout:   30 * time.Second,
    KeepAlive: 30 * time.Second,
}

transport := &http.Transport{
    DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
        // Perform DNS lookup on EVERY dial (no caching)
        return dialer.DialContext(ctx, network, addr)
    },
    DisableKeepAlives:   false,  // Keep connections alive
    MaxIdleConns:        100,    // Pool idle connections
    MaxIdleConnsPerHost: 10,
}
```

**Trade-off**: Slight latency increase (DNS lookup per request) for guaranteed fresh endpoints.

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q18

**4. Environment Variables** (Overrides config file):

```bash
# Environment variables override config file
CRYPTOUTIL_FEDERATION_IDENTITY_URL="https://identity:8180"
CRYPTOUTIL_FEDERATION_JOSE_URL="https://jose:8280"
```

#### Graceful Degradation

**Circuit Breaker**: Automatically disable federated service after N consecutive failures

- **Failure Threshold**: Open circuit after N failures (default: 5)
- **Timeout**: Reset circuit after timeout (default: 60s)
- **Half-Open Requests**: Test N requests before closing circuit (default: 3)

**Circuit Breaker State Transitions**:

- **Closed State**: Normal operation, all retry strategies active (exponential backoff, timeout escalation)
- **Open State**: Circuit opened after N failures, **FAIL-FAST** mode activated:
  - All requests immediately fail without retry attempts
  - No retry mechanisms execute (no exponential backoff, no health checks)
  - Remains open for timeout duration (default: 60s)
- **Half-Open State**: After timeout expires, test service availability:
  - Allow N test requests with retry logic enabled (default: 3 requests)
  - Success → transition to Closed state
  - Failure → return to Open state, reset timeout

**Critical**: Retry mechanisms (exponential backoff, timeout escalation, health checks) ONLY execute in Closed and Half-Open states. In Open state, requests fail immediately without any retry attempts.

**Fallback Modes**:

- **Identity Unavailable**:
  - `local_validation`: Cached public keys for token validation
  - `reject_all`: Strict mode, deny all requests (production recommended)
  - `allow_all`: Development only, bypass authentication
- **JOSE Unavailable**:
  - `internal_crypto`: Use service's own JWE/JWS implementation
- **CA Unavailable**:
  - `self_signed`: Generate self-signed TLS certificates (development)
  - `cached_certs`: Use cached certificates (production)

**Retry Strategies**:

- **Exponential Backoff**: 1s, 2s, 4s, 8s, 16s (max 5 retries)
- **Timeout Escalation**: Increase timeout 1.5x per retry (10s → 15s → 22.5s)
- **Health Check Before Retry**: Poll `/admin/api/v1/livez` endpoint (fast liveness check) before resuming traffic

**Federation Timeout Configuration** (Source: SPECKIT-CLARIFY-QUIZME-05 Q16, 2025-12-24):

**MANDATORY**: Federation timeouts MUST be configurable per service (NOT global timeout).

**Rationale**: Different services have different latency characteristics:

- **Identity**: Fast token validation (5-10s timeout)
- **JOSE**: Fast signing/encryption (10-15s timeout)
- **CA**: Slow certificate issuance (30-60s timeout)

**Global Timeout Forbidden**:

```yaml
# ❌ FORBIDDEN - DO NOT USE
federation:
  global_timeout: 10s  # Breaks CA operations
```

**Per-Service Timeout Required**:

```yaml
# ✅ REQUIRED
federation:
  identity_timeout: 10s
  jose_timeout: 15s
  ca_timeout: 30s
```

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q16

**API Versioning** (Source: SPECKIT-CLARIFY-QUIZME-05 Q17, 2025-12-24):

**MANDATORY**: Federation APIs MUST maintain N-1 backward compatibility.

**Version Support Policy**:

- **Current version**: v2 (latest)
- **Supported versions**: v2, v1 (N-1)
- **Deprecated versions**: v0 (removed)

**Example**:

```yaml
# Service supports both v2 and v1 simultaneously
federation:
  identity_url_v2: "https://identity-authz:8180/api/v2"  # Preferred
  identity_url_v1: "https://identity-authz:8180/api/v1"  # Fallback for old clients
```

**Upgrade Strategy**:

1. **Deploy v2**: Add new endpoints alongside v1
2. **Migrate clients**: Gradually update clients to v2
3. **Deprecate v1**: After 6 months, announce v1 deprecation
4. **Remove v1**: After 12 months, remove v1 endpoints

**Rationale**: N-1 compatibility enables zero-downtime rolling upgrades across federated services.

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q17

#### Federation Health Monitoring

**Regular Health Checks**:

- Check federated service health every 30 seconds
- Log warnings when services become unhealthy
- Activate fallback mode when health checks fail

**Metrics and Alerts**:

- `federation_request_duration_seconds{service="identity"}` - Latency tracking
- `federation_request_failures_total{service="identity"}` - Error rate
- `federation_circuit_breaker_state{service="identity"}` - Circuit state (closed/open/half-open)

#### Cross-Service Authentication

**Service-to-Service mTLS** (Preferred):

```yaml
federation:
  identity_url: "https://identity-authz:8180"
  identity_client_cert: "file:///run/secrets/kms_client_cert"
  identity_client_key: "file:///run/secrets/kms_client_key"
  identity_ca_cert: "file:///run/secrets/identity_ca_cert"
```

**OAuth 2.1 Client Credentials** (Alternative):

```yaml
federation:
  identity_url: "https://identity-authz:8180"
  identity_client_id: "kms-service"
  identity_client_secret: "file:///run/secrets/kms_client_secret"
  identity_token_endpoint: "https://identity-authz:8180/service/token"
```

#### Federation Testing Requirements

**Integration Tests MUST**:

- Test each federated service independently (mock others)
- Test graceful degradation when federated service unavailable
- Test circuit breaker behavior (failure thresholds, timeouts, recovery)
- Test retry logic (exponential backoff, max retries)
- Verify timeout configurations prevent cascade failures

**E2E Tests MUST**:

- Deploy full stack (all federated services)
- Test cross-service communication paths
- Test federation with Docker Compose service discovery
- Verify health checks detect service failures
- Test failover and recovery scenarios

---

#### Private HTTPS Server

**Purpose**: Internal admin tasks, health checks, metrics
**Bind**: `127.0.0.1:9090` (not externally accessible)
**Security**:

- IP restriction to localhost only
- Minimal middleware (no CORS/CSRF)
- Optional mTLS for production environments
- Not exposed in Docker port mappings

**Admin Port Assignments** (Source: constitution.md, 2025-12-24):

- **ALL SERVICES**: Admin port 9090 (bound to 127.0.0.1, NEVER exposed to host)
- **TESTS**: Admin port 0 (dynamic allocation)
- **Rationale**: Admin endpoints localhost-only, container network namespace isolation allows same port across all services

**Admin API Context**:

- `/admin/api/v1/livez` - Liveness probe (lightweight check: service running, process alive)
- `/admin/api/v1/readyz` - Readiness probe (heavyweight check: dependencies healthy, ready for traffic)
- `/admin/api/v1/metrics` - Prometheus metrics endpoint
- `/admin/api/v1/shutdown` - Graceful shutdown trigger

**Health Check Semantics**:

- **livez**: Fast, lightweight check (~1ms) - verifies process is alive, TLS server responding
- **readyz**: Slow, comprehensive check (~100ms+) - verifies database connectivity, downstream services, resource availability
- **Use livez for**: Docker healthchecks (fast, frequent), liveness probes (restart on failure)
- **Use readyz for**: Kubernetes readiness probes (remove from load balancer), deployment validation

**Health Check Failure Behavior** (Source: SPECKIT-CLARIFY-QUIZME-05 Q15, 2025-12-24):

**Kubernetes** (Production):

```yaml
livenessProbe:
  httpGet:
    path: /admin/api/v1/livez
    port: 9090
    scheme: HTTPS
  failureThreshold: 3  # After 3 failures, pod restarted
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /admin/api/v1/readyz
    port: 9090
    scheme: HTTPS
  failureThreshold: 1  # After 1 failure, pod removed from LB
  periodSeconds: 5
```

**Behavior**:

- **Liveness failure**: Kubernetes KILLS pod and starts replacement
- **Readiness failure**: Kubernetes REMOVES pod from service load balancer endpoints (pod continues running)
- **Recovery**: Pod re-added to LB when readyz passes again

**Docker Compose** (Development/Testing):

```yaml
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/api/v1/livez"]
  interval: 5s
  timeout: 5s
  retries: 3
  start_period: 10s
```

**Behavior**:

- **Liveness failure**: Docker marks container as `unhealthy` but DOES NOT restart
- **Manual intervention**: Operator must manually restart unhealthy containers
- **Monitoring**: `docker ps` shows health status, alerts on unhealthy state

**Rationale**: Kubernetes prioritizes availability (auto-restart), Docker Compose prioritizes debugging (preserve unhealthy state for inspection).

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q15

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
          │ Admin:9090    │ Admin:9090     │ Admin:9090   │Admin:9090
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
| **Public** | All services (ports 8050-8089) | OAuth 2.1 tokens, rate limiting, TLS 1.3+ |
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

## Shared Code Organization - CRITICAL

**MANDATORY**: Reusable code needed by service template and all 8+1 services MUST be located in `internal/shared/` packages. This ensures code reuse, prevents duplication, and enables consistent behavior across all services.

### Required Shared Packages

| Package Location | Purpose | Usage |
|------------------|---------|-------|
| `internal/shared/crypto/certificate/` | TLS cert chain generation (Root CAs, Intermediate CAs, Policy CAs, Issuing CAs, TLS Servers, TLS Clients) | ✅ **CRITICAL** - All services need TLS cert generation for dual HTTPS servers |
| `internal/shared/crypto/certificate/tlsconfig/` | TLS configurations (TLS Server configs, TLS Client configs) | ✅ **CRITICAL** - All services need TLS configuration for secure communications |
| `internal/shared/crypto/jose/` | JOSE crypto generation service (JWK Gen Service, JWE/JWS utilities) | ✅ **CRITICAL** - Required by cipher-im (JWE for encrypted messaging), jose-ja, sm-kms |
| `internal/shared/telemetry/` | OpenTelemetry service (traces, metrics, logs) | ✅ **IMPORTANT** - All services need consistent telemetry |
| `internal/shared/magic/` | Magic constants and variables | ✅ **IMPORTANT** - Shared constants across all services |
| `internal/shared/crypto/digests/` | Digest algorithms (SHA-256, SHA-384, SHA-512, HMAC) | ✅ **IMPORTANT** - Cryptographic digests for all services |
| `internal/shared/crypto/hash/` | Hash service (PBKDF2, HKDF, password hashing) | ✅ **IMPORTANT** - Password and key derivation across services |
| `internal/shared/util/` | Utility code (validators, converters, helpers) | ✅ **IMPORTANT** - Common utilities across services |

### Course Correction: Move internal/jose/crypto

**CRITICAL**: The package `internal/jose/crypto` is currently in the wrong location. It contains reusable JOSE crypto code needed by multiple services:

- **cipher-im**: Requires JWE for encrypt+MAC secure Instant Messaging
- **jose-ja**: JOSE Authority service itself
- **sm-kms**: Key management service for key wrapping

**Action Required**: Phase 1.1 MUST move `internal/jose/crypto` to `internal/shared/crypto/jose/` BEFORE work starts on cipher-im service. This ensures:

1. JWK Gen Service is reusable by all services needing JOSE operations
2. JWE/JWS utilities are available for cipher-im encrypted messaging
3. No circular dependencies between services
4. Consistent JOSE implementation across all products

### Course Correction: Better TLS Code Reuse

**CRITICAL**: Current service template work shows duplication of TLS certificate generation code. The reusable TLS code in `internal/shared/crypto/certificate/` and `internal/shared/crypto/keygen/` MUST be used for auto-generating TLS server cert chains.

**Issues Found**:

- Service template creating new TLS generation code that duplicates existing patterns
- Hard-coded values in service template methods instead of parameter injection
- Creates technical debt that will need fixing when migrating existing services (Phases 3-9)

**Action Required**: Phase 1.2 MUST refactor service template to:

1. Use reusable TLS code from `internal/shared/crypto/certificate/` for cert generation
2. Use parameter injection patterns instead of hard-coded values
3. Eliminate duplication between service template and existing TLS infrastructure
4. Document parameter injection patterns for service customization

### Rationale for Shared Packages

**Code Reuse**: Prevents duplication across 9 services (sm-kms, pki-ca, jose-ja, identity-authz, identity-idp, identity-rs, identity-rp, identity-spa, cipher-im)

**Consistency**: Ensures all services use same TLS patterns, JOSE operations, telemetry, hashing algorithms

**Maintainability**: Single location for cryptographic primitives, easier to update and audit

**Testing**: Shared packages can be tested once and reused with confidence across all services

**Security**: Centralized crypto code reduces risk of inconsistent security implementations

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

**Architecture**: 5 independent microservices that can be deployed standalone or together:

1. **AuthZ Server**: OAuth 2.1 Authorization Server (identity-authz, port 8180, admin 9090)
2. **IdP Server**: OIDC Identity Provider (identity-idp, port 8181, admin 9090)
3. **Resource Server**: Protected API with token validation (identity-rs, port 8182, admin 9090) - reference implementation
4. **Relying Party**: Backend-for-Frontend pattern (identity-rp, port 8183, admin 9090) - reference implementation
5. **Single Page Application**: Static hosting for SPA clients (identity-spa, port 8184, admin 9090) - reference implementation

Each service has its own Docker image and can scale independently.

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

#### Authentication and Authorization Requirements

**Source**: QUIZME-02 answers (Q1-Q15)
**Reference**: See `.specify/memory/authn-authz-factors.md` for authoritative authentication/authorization factor list

**Single Factor Authentication Methods (SFA)**:

- **Headless-Based Clients** (`/service/*` paths): 10 methods (3 non-federated + 7 federated)
- **Browser-Based Clients** (`/browser/*` paths): 28 methods (6 non-federated + 22 federated)
- **Complete list with per-factor storage realms**: `.specify/memory/authn-authz-factors.md`

**Storage Realm Pattern**:

- **YAML + SQL (Config > DB priority)**: Static credentials, provider configs (disaster recovery - service starts without database)
- **SQL ONLY**: User-specific enrollment data, one-time tokens/codes (dynamic per-user)
- **Details**: See `.specify/memory/authn-authz-factors.md` Section "Storage Realm Specifications"

**Multi-Factor Authentication (MFA)**:

- MFA = Combination of 2+ single factor authentication methods
- Factor priority order: Passkey > TOTP > Hardware Keys > Email OTP > SMS OTP > HOTP > Recovery Codes > Push Notifications > Phone Call OTP
- **Common combinations and patterns**: See `.specify/memory/authn-authz-factors.md` Section "Multi-Factor Authentication"

**Authorization Methods**:

**Headless-Based Clients** (2 methods):

- Scope-Based Authorization
- Role-Based Access Control (RBAC)

**Browser-Based Clients** (4 methods):

- Scope-Based Authorization
- Role-Based Access Control (RBAC)
- Resource-Level Access Control
- Consent Tracking (scope+resource tuples)

**Session Token Format** (Q3 - Configuration-Driven):

Token format selection is configuration-driven per service deployment. Administrators configure format via YAML configuration files, with defaults of opaque tokens for browser-based clients and JWS tokens for headless clients. All three formats (opaque, JWE, JWS) MUST be supported by all services to enable flexible deployment patterns.

```yaml
# Non-Federated Mode - Product decides format
session:
  token_format: opaque  # or jwe, jws (default: opaque for browser, jws for headless)

# Federated Mode - Identity Provider decides format
federation:
  identity:
    session_token_format: jwe  # or jws, opaque (default: opaque for browser, jws for headless)
```

**Configuration Pattern Details** (see [.github/instructions/02-10.authentication.instructions.md](../../.github/instructions/02-10.authentication.instructions.md)):

- Format selection: Administrator-configured via YAML deployment configuration
- Default behavior: Opaque tokens for browser clients, JWS tokens for headless clients
- Mandatory support: All services MUST implement all three formats (opaque, JWE, JWS)
- Rationale: Enables deployment flexibility, security/performance tradeoffs per environment

**Session Storage Backend** (Q4 - PostgreSQL/SQLite Only):

```yaml
# Single-node deployments
database:
  driver: sqlite
  dsn: "file:sessions.db?cache=shared"

# Distributed/HA deployments
database:
  driver: postgres
  dsn: "postgres://user:pass@host:5432/sessions?sslmode=require"
```

**MFA Step-Up Authentication** (Q6 - Time-Based):

- Re-authentication MANDATORY every 30 minutes for sensitive resources
- Applies to operations: key rotation, client secret rotation, admin actions
- Session remains valid for low-sensitivity operations

**MFA Enrollment Workflow** (Q7 - Optional with Limited Access):

- Enrollment OPTIONAL during initial setup
- Access LIMITED until additional factors enrolled (read-only access)
- User MUST enroll at least one factor for write operations
- Only one identifying factor required for initial login

**Realm Failover Behavior** (Q10 - Priority List):

```yaml
realms:
  priority_list:
    - type: file
      realm: config.yaml
    - type: database
      realm: postgresql_production
    - type: database
      realm: sqlite_fallback
```

System tries each Realm+Type in priority order until one succeeds or all fail.

**Zero Trust Authorization** (Q11 - No Caching):

- Authorization decisions MUST be evaluated on EVERY request
- NO caching of authorization decisions (prevents stale permissions)
- Performance via efficient policy evaluation, not caching

**Cross-Service Authorization** (Q12 - Direct Token Validation):

- Session token passed between federated services via HTTP headers
- Each service independently validates token and enforces authorization
- NO token transformation or delegation

**Consent Tracking Granularity** (Q15 - Scope+Resource Tuples):

- Tracked as `(scope, resource)` tuples
- Example: `("read:keys", "key-123")` separate from `("read:keys", "key-456")`
- Enables fine-grained consent revocation per resource

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
- **ca-sqlite**: Port 8380 (public API), Port 9090 (admin), SQLite backend
- **ca-postgres-1**: Port 8381 (public API), Port 9090 (admin), PostgreSQL backend
- **ca-postgres-2**: Port 8382 (public API), Port 9090 (admin), PostgreSQL backend
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
- Coverage targets: 95% production, 98% infrastructure, 98% utility
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
- ❌ Admin API (`/admin/api/v1/*`): No Gatling simulation
- ❌ Multi-product integration: No cross-service workflow tests

**Required**: Create `BrowserApiSimulation.java` and `AdminApiSimulation.java` to complete load test coverage.

**Performance Metrics Approach** (Source: CLARIFY-QUIZME-01 Q9, 2025-12-22):

- **No Hard Targets**: Load tests validate scalability trends and identify bottlenecks only
- **Baseline Establishment**: Initial load testing establishes baseline performance metrics
- **Iterative Improvement**: Track trends (requests/second, latency percentiles, error rates) and improve over time
- **Rationale**: Performance requirements vary by deployment scale and hardware; focus on bottleneck identification

#### E2E Test Scope

**Current**: Basic Docker Compose lifecycle tests (`internal/test/e2e/e2e_test.go`)

- Service startup/shutdown
- Health check connectivity
- Container log collection

**Phase 2 Priority** (Source: CLARIFY-QUIZME-01 Q10, 2025-12-22):

**JOSE + CA + KMS** (implement first):

- ❌ JOSE signing and verification workflows (JWS, JWT, JWE)
- ❌ CA certificate issuance and revocation (CSR → certificate → CRL/OCSP)
- ❌ KMS key generation, encryption/decryption, rotation

**Identity Product** (Phase 3+):

- ❌ OAuth 2.1 authorization code flow (browser → AuthZ → IdP → consent → token)
- ❌ OIDC authentication flow
- ❌ Token validation workflows

**Rationale**: JOSE, CA, and KMS are standalone products with clear E2E scenarios. Identity product has complex multi-service interactions that benefit from later implementation after other products stabilize.

**Required**: Expand E2E tests to cover end-to-end product workflows, not just infrastructure.

**E2E Test Coverage Requirements** (Source: SPECKIT-CLARIFY-QUIZME-05 Q13, 2025-12-24):

**MANDATORY**: E2E tests MUST cover BOTH `/service/*` and `/browser/*` API paths.

**Priority**: Implement `/service/*` path tests FIRST, then `/browser/*` path tests.

**`/service/*` Path Tests** (Service-to-Service APIs):

```go
// Test OAuth 2.1 Client Credentials flow
func TestServiceAPIAuthentication(t *testing.T) {
    // 1. Obtain access token via client_credentials
    token := oauth.GetClientToken(clientID, clientSecret)

    // 2. Call /service/api/v1/keys with Bearer token
    resp := http.Get("/service/api/v1/keys",
        http.Header{"Authorization": "Bearer " + token})

    // 3. Verify response
    assert.Equal(t, 200, resp.StatusCode)
}
```

**`/browser/*` Path Tests** (Browser-Based APIs):

```go
// Test Authorization Code + PKCE flow
func TestBrowserAPIAuthentication(t *testing.T) {
    // 1. Initiate authorization code flow
    authURL := oauth.StartAuthCodeFlow(redirectURI, pkce)

    // 2. Simulate user login and consent
    code := browser.Login(authURL, username, password)

    // 3. Exchange code for session cookie
    cookie := oauth.ExchangeCodeForSession(code, pkce)

    // 4. Call /browser/api/v1/keys with session cookie
    resp := http.Get("/browser/api/v1/keys",
        http.Header{"Cookie": cookie})

    // 5. Verify response
    assert.Equal(t, 200, resp.StatusCode)
}
```

**Test Environment**: Docker Compose with full service stack (authz, idp, kms, jose, ca)

**Rationale**: `/service/*` path simpler to implement (no browser simulation), validates core OAuth 2.1 flows. `/browser/*` path requires browser automation (Playwright/Selenium) for login flows.

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q13

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
- Coverage targets: 95% production, 98% infrastructure, 98% utility
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
- SQLite (development/testing/small-scale production)
- GORM ORM with migrations
- WAL mode, busy_timeout for SQLite concurrency
- **GitHub Actions Dependency**: ALL workflows running `go test` MUST include PostgreSQL service container

**SQLite Production Support** (Source: CLARIFY-QUIZME-01 Q3, 2025-12-22):

- **Acceptable**: SQLite for production single-instance deployments with <1000 requests/day
- **Recommended**: PostgreSQL for all other production deployments
- **Rationale**: Small-scale deployments benefit from SQLite's simplicity (no separate database server, zero-configuration)
- **Limitation**: Traffic threshold ensures SQLite's single-writer limitation isn't violated

#### PostgreSQL Service Requirements for CI/CD

**CRITICAL**: PostgreSQL integration testing MUST use `test-containers` library, NOT GitHub Actions service containers.

**MANDATORY**: Any GitHub Actions workflow executing `go test` on packages using database repositories MUST use test-containers:

```yaml
steps:
  - name: Run tests with PostgreSQL
    run: |
      # test-containers library automatically:
      # 1. Pulls postgres:18 image
      # 2. Starts container with random port
      # 3. Waits for pg_isready health check
      # 4. Injects connection string via environment
      # 5. Cleans up container after tests
      go test -v ./internal/kms/server/repository/...
```

**Why test-containers Over Service Containers**:

- **Parallelism**: Each test package gets isolated PostgreSQL instance (no port conflicts)
- **Cleanup**: Containers automatically removed after test completion
- **Portability**: Works locally, in CI, and in Docker-in-Docker environments
- **Flexibility**: Tests control PostgreSQL version, extensions, and configuration per package

**Service Containers Forbidden**:

```yaml
# ❌ FORBIDDEN - DO NOT USE
services:
  postgres:
    image: postgres:18
    # Problems: shared instance, port conflicts, manual cleanup
```

**Test-containers Implementation**:

```go
// internal/test/testdb/postgres.go
func StartPostgres(ctx context.Context) (*PostgresContainer, error) {
    req := testcontainers.ContainerRequest{
        Image:        "postgres:18",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_DB":       "cryptoutil_test",
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_USER":     "cryptoutil",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections"),
    }
    return testcontainers.GenericContainer(ctx, req)
}
```

**Affected Workflows**: ci-race, ci-mutation, ci-coverage, any workflow running database tests

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q10

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
| `ci-e2e` | Push, PR | <20 min | ❌ | End-to-end Docker Compose tests (BOTH `/service/*` and `/browser/*` paths) |
| `ci-load` | Push, PR | <30 min | ❌ | Load testing (Gatling - Service API only) |
| `ci-identity-validation` | Push, PR | <5 min | ✅ | Identity-specific validation tests |
| `release` | Tag | <15 min | ❌ | Build and publish release artifacts |

**Total CI Feedback Loop Target**: <10 minutes for critical path (quality + coverage + race)
**Full Suite Target**: <60 minutes for all workflows to complete

**Docker Pre-pull Optimization** (Source: SPECKIT-CLARIFY-QUIZME-05 Q9, 2025-12-24):

**CRITICAL**: Docker image pre-pull MUST ONLY be used in workflows that actually use Docker.

**Pre-pull Required**:

- `ci-e2e`: Uses Docker Compose for end-to-end testing
- `ci-dast`: Uses Nuclei/ZAP Docker containers for security scanning

**Pre-pull NOT Required**:

- `ci-quality`: Linting/formatting (no Docker)
- `ci-coverage`: Go test coverage (uses test-containers, auto-pulls)
- `ci-race`: Race detector (uses test-containers, auto-pulls)
- `ci-mutation`: Mutation testing (no Docker)
- `ci-benchmark`: Benchmarks (no Docker)
- `ci-fuzz`: Fuzz testing (no Docker)
- `ci-sast`: Static analysis (no Docker)
- `ci-gitleaks`: Secrets scanning (no Docker)

**Example Pre-pull Configuration**:

```yaml
# .github/workflows/ci-e2e.yml
steps:
  - name: Pre-pull Docker images
    run: |
      docker pull postgres:18
      docker pull grafana/otel-lgtm:latest
      docker pull otel/opentelemetry-collector-contrib:latest

  - name: Run E2E tests
    run: docker compose -f deployments/compose.integration.yml up --abort-on-container-exit
```

**Rationale**: Pre-pull reduces E2E test flakiness from registry rate limits, but adds unnecessary overhead to workflows that don't use Docker.

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q9

**Health Check Pattern Standardization**:

- **Alpine containers**: Use `wget --no-check-certificate -q -O /dev/null <url>`
- **Non-Alpine containers**: Use `curl -k -f -s <url>`
- **Retry logic**: `start_period: 10s`, `interval: 5s`, `retries: 5`, `timeout: 5s`
- **Admin endpoints**: All services use `https://127.0.0.1:9090/admin/api/v1/livez` for Docker health checks

---

## Quality Requirements

### Code Coverage Targets

| Category | Target | Current |
|----------|--------|---------|
| Production Code | ≥95% | Varies |
| Infrastructure (cicd) | ≥98% | ~90% |
| Utility Code | ≥98% | ~98% |

### Mutation Testing Requirements

- Minimum ≥80% gremlins score per package
- Focus on business logic, parsers, validators, crypto operations
- Track improvements in baseline reports

**Mutation Exemptions** (Source: SPECKIT-CLARIFY-QUIZME-05 Q12, 2025-12-24):

**MANDATORY**: The following code categories are EXEMPT from mutation testing:

1. **OpenAPI-generated code**: `api/client/`, `api/server/`, `api/model/`
   - **Rationale**: Generated code quality depends on OpenAPI spec, not mutation testing
   - **Alternative**: Validate OpenAPI spec correctness, test API contracts

2. **GORM migration files**: `internal/*/repository/migrations/*.sql`
   - **Rationale**: SQL migration correctness validated by integration tests, not mutation testing
   - **Alternative**: Test migration up/down sequences, data integrity checks

3. **Protobuf-generated code**: `*.pb.go`, `*_grpc.pb.go`
   - **Rationale**: Generated from .proto files, mutation testing adds no value
   - **Alternative**: Validate protobuf schema correctness, test serialization/deserialization

**Gremlins Configuration**:

```yaml
# .gremlins.yaml
exemptions:
  - "api/client/**"
  - "api/server/**"
  - "api/model/**"
  - "**/migrations/*.sql"
  - "**/*.pb.go"
  - "**/*_grpc.pb.go"
```

**Rationale**: Focus mutation testing effort on hand-written business logic where it provides maximum value.

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q12

### Linting Requirements

- golangci-lint v2.7.2+
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
| ca-simple | 8050 | 9090 | SQLite | ✅ Development only (`compose.simple.yml`) |
| ca-postgres-1 | 8051 (planned) | 9090 (planned) | PostgreSQL | ⚠️ Production config missing |
| ca-postgres-2 | 8052 (planned) | 9090 (planned) | PostgreSQL | ⚠️ Production config missing |

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
| JOSE | `/admin/api/v1/livez` | Liveness probe (lightweight) |
| JOSE | `/admin/api/v1/readyz` | Readiness probe (heavyweight) |
| Identity | `/admin/api/v1/livez` | Liveness probe (lightweight) |
| Identity | `/admin/api/v1/readyz` | Readiness probe (heavyweight) |
| KMS | `/admin/api/v1/livez` | Liveness probe (lightweight) |
| KMS | `/admin/api/v1/readyz` | Readiness probe (heavyweight) |
| CA | `/admin/api/v1/livez` | Liveness probe (lightweight, planned) |
| CA | `/admin/api/v1/readyz` | Readiness probe (heavyweight, planned) |

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
├── LowEntropyRandomHashRegistry (PBKDF2-based, salted)
│   ├── v1: 0-31 bytes → PBKDF2-HMAC-SHA256 (OWASP rounds)
│   ├── v2: 32-47 bytes → PBKDF2-HMAC-SHA384
│   └── v3: 48+ bytes → PBKDF2-HMAC-SHA512
├── LowEntropyDeterministicHashRegistry (PBKDF2-based, fixed + derived salt)
│   ├── v1: 0-31 bytes → PBKDF2-HMAC-SHA256 (fixed salt per version, derive actual salt from fixed + cleartext)
│   ├── v2: 32-47 bytes → PBKDF2-HMAC-SHA384 (different fixed salt than v1)
│   └── v3: 48+ bytes → PBKDF2-HMAC-SHA512 (different fixed salt than v1/v2)
├── HighEntropyRandomHashRegistry (HKDF-based, salted)
│   ├── v1: 0-31 bytes → HKDF-HMAC-SHA256
│   ├── v2: 32-47 bytes → HKDF-HMAC-SHA384
│   └── v3: 48+ bytes → HKDF-HMAC-SHA512
└── HighEntropyDeterministicHashRegistry (HKDF-based, fixed + derived salt)
    ├── v1: 0-31 bytes → HKDF-HMAC-SHA256 (fixed salt per version, derive actual salt from fixed + cleartext)
    ├── v2: 32-47 bytes → HKDF-HMAC-SHA384 (different fixed salt than v1)
    └── v3: 48+ bytes → HKDF-HMAC-SHA512 (different fixed salt than v1/v2)
```

**Salt Encoding Requirements** (CRITICAL for Security):

- **LowEntropyRandomHashRegistry / HighEntropyRandomHashRegistry**:
  - MUST encode version AND all parameters (iterations, salt, algorithm) WITH the hash
  - Format: `{version}:{algorithm}:{params}:base64(salt):base64(hash)`
  - Example: `{1}:PBKDF2-HMAC-SHA256:rounds=600000:abcd1234...:efgh5678...`
  - Rationale: Random salt must be stored to verify later
- **LowEntropyDeterministicHashRegistry / HighEntropyDeterministicHashRegistry**:
  - MUST encode version ONLY (NEVER encode salt or parameters in output)
  - Format: `{version}:base64(hash)`
  - Example: `{1}:abcd1234...`
  - Rationale: Revealing salt in DB would be crypto bug
  - MUST use different fixed configurable salt per version (v1/v2/v3 each have unique salt)
  - MUST derive ACTUAL SALT from combination of:
    - Configured fixed salt (acts as pepper, secret key)
    - Input cleartext (adds input-specific entropy)
  - Derivation similar to AES-GCM-SIV (derive IV from nonce + cleartext) but for different purpose
  - Purpose: Obfuscate actual salt used (pepper-like concept per OWASP Password Storage Cheat Sheet)
  - Security: Fixed salt never revealed, derived salt unique per input
- **LowEntropyDeterministicHashRegistry / HighEntropyDeterministicHashRegistry**:
  - MUST encode version ONLY (NEVER encode salt or parameters in output)
  - Format: `{version}:base64(hash)`
  - Example: `{1}:abcd1234...`
  - Rationale: Revealing salt in DB would be crypto bug
  - MUST use different fixed configurable salt per version (v1/v2/v3 each have unique salt)
  - MUST derive ACTUAL SALT from combination of:
    - Configured fixed salt (acts as pepper, secret key)
    - Input cleartext (adds input-specific entropy)
  - Derivation similar to AES-GCM-SIV (derive IV from nonce + cleartext) but for different purpose
  - Purpose: Obfuscate actual salt used (pepper-like concept per OWASP Password Storage Cheat Sheet)
  - Security: Fixed salt never revealed, derived salt unique per input

**Pepper Rotation Strategy** (Source: SPECKIT-CLARIFY-QUIZME-05 Q3, 2025-12-24):

**MANDATORY**: Pepper rotation MUST use lazy migration (re-hash on re-authentication).

**Rotation Process**:

1. **Add New Pepper Version**: Deploy new pepper configuration with incremented version

   ```yaml
   hash_registry:
     versions:
       - version: 1
         pepper: "old-pepper-value"
       - version: 2
         pepper: "new-pepper-value"  # New version added
   ```

2. **Verify Existing Hashes**: Continue accepting hashes with old pepper (version 1)

3. **Re-hash on Re-authentication**: When user logs in, re-hash password with new pepper (version 2)

   ```go
   if oldHash.Version == 1 {
       newHash := registry.HashWithVersion(password, 2)
       db.UpdatePasswordHash(userID, newHash)
   }
   ```

4. **Gradual Migration**: Over time, all active users migrated to new pepper

5. **Retire Old Pepper**: After grace period (e.g., 90 days), remove version 1 pepper from config

**NO Forced Re-authentication**: Users NOT required to re-authenticate immediately.

**Rationale**: Lazy migration balances security (pepper rotation) with user experience (no forced password resets).

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q3

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

- **Dual HTTPS Servers**: Public API (<configurable_address>:<configurable_port>) + Admin API (127.0.0.1:9090)
- **Admin Endpoints**: `/livez`, `/readyz`, `/shutdown` on 127.0.0.1:9090
  - Admin prefix configurable (default: `/admin/api/v1`)
  - Implementation: gofiber middleware (reference: sm-kms `internal/kms/server/application/application_listener.go`)
- **Dual API Paths**: `/browser/api/v1/*` (session-based) vs `/service/api/v1/*` (token-based)
- **Middleware Pipeline**: CORS/CSRF/CSP (browser-only), rate limiting, IP allowlist, authentication
- **Database Abstraction**: PostgreSQL + SQLite dual support with GORM
- **OpenTelemetry Integration**: OTLP traces, metrics, logs
- **Health Check Endpoints**: `/admin/api/v1/livez` (liveness), `/admin/api/v1/readyz` (readiness)
- **Graceful Shutdown**: `/admin/api/v1/shutdown` endpoint
- **Docker Compose Requirements**:
  - OpenTelemetry Collector Contrib MUST use separate health check job (does NOT expose external health endpoint)
  - Reference: KMS Docker Compose `deployments/compose/compose.yml` (working pattern)
  - MUST include Docker Secrets validation job (fast-fail check before starting services)

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

### Cipher-IM Demonstration Service (Phase 3)

**Goal**: Create working InstantMessenger service using service template, validate reusability and completeness.

**Implementation Priority**: HIGH - CRITICAL to implement cipher-im FIRST before migrating production services

**Service Template Migration Priority Order**:

1. **cipher-im FIRST** (Phase 3):
   - CRITICAL: Implement cipher-im using extracted service template
   - Through iterative implementation, testing, validation, analysis
   - GUARANTEE ALL requirements of service template are met
   - Proves template is production-ready before migrating product services
2. **One service at a time** (Phase 4+, excludes sm-kms):
   - MUST refactor each production service to use service template sequentially
   - Identify and fix issues in service template to unblock current service
   - Avoid creating technical debt affecting remaining migrations
   - Order: jose-ja, pki-ca, identity-authz, identity-idp, identity-rs, identity-rp, identity-spa
3. **sm-kms LAST** (Phase 10):
   - ALL other services MUST be refactored and running excellently on service template
   - Only migrate KMS reference implementation after template proven stable across 8 services
   - Reference implementation stays stable until template is battle-tested

**Cipher-IM Overview**:

- **Product**: Cipher (educational/demonstration product)
- **Service**: IM (InstantMessenger service)
- **Purpose**: Copy-paste-modify starting point for customers creating new services
- **Scope**: Encrypted messaging API (PUT/GET/DELETE for /tx and /rx endpoints)

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
            Origins: []string{"https://cipher-im.example.com"},
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
- **Starting Point**: Copy entire Cipher-IM directory, modify for use case
- **Best Practices**: Learn production-ready patterns (error handling, testing, deployment)
- **API Design**: Reference implementation for REST API design

---

---

## Non-Functional Requirements

### Performance and Scaling

#### Vertical Scaling

**Resource Limits** (Per-service configuration):

- CPU limits: 500m-2000m (0.5-2 CPU cores)
- Memory limits: 256Mi-1Gi (configurable per service)
- Connection pool sizing: Based on workload (PostgreSQL 10-50, SQLite 5)
- Concurrent request handling: Configurable (default: 100 concurrent requests)

**Resource Monitoring**:

- OTLP metrics: CPU usage, memory usage, goroutine count
- Health checks: Resource exhaustion detection
- Graceful degradation: Circuit breaker when resources depleted

#### Horizontal Scaling

**Load Balancing Patterns**:

- **Layer 7 (HTTP/HTTPS)**: Use reverse proxy (nginx, Traefik, Envoy) for path-based routing
- **Layer 4 (TCP)**: Use TCP load balancer for raw connection distribution
- **DNS-based**: Round-robin DNS for simple load distribution
- **Service mesh**: Istio/Linkerd for advanced traffic management

**Session State Management for Horizontal Scaling**:

**CRITICAL**: Sessions MUST use SQL-only storage (PostgreSQL or SQLite). Redis is NOT supported.

**Session Formats**:

1. **JWS (JSON Web Signature)**: Stateless signed tokens
   - **Pros**: No server-side storage, horizontal scaling trivial
   - **Cons**: Cannot revoke before expiry, larger cookie size
   - **Use case**: Low-security browser sessions with short TTL

2. **OPAQUE**: Server-side database storage with opaque session ID
   - **Pros**: Immediate revocation, smaller cookie size
   - **Cons**: Database lookup on every request, horizontal scaling requires shared database
   - **Use case**: Standard browser sessions requiring revocation

3. **JWE (JSON Web Encryption)**: Encrypted stateless tokens
   - **Pros**: Privacy protection, no server-side storage, horizontal scaling trivial
   - **Cons**: Cannot revoke before expiry, larger cookie size, encryption overhead
   - **Use case**: High-security browser sessions with encrypted claims

**Implementation Priority**: JWS → OPAQUE → JWE

**Deployment Priority**: JWE → OPAQUE → JWS

**Rationale**: Implement simplest first (JWS) to validate architecture, deploy most secure first (JWE) for production.

**Legacy Session Migration**: NOT SUPPORTED - session format changes require re-authentication

**Database Scaling Patterns**:

- **Read replicas**: NOT SUPPORTED - all reads directed to primary database
  - **Rationale**: Replication lag introduces consistency issues for security-critical operations
  - **Alternative**: Vertical scaling of primary database + connection pooling

- **Connection pooling**: REQUIRED - configurable and hot-reloadable without service restart
  - **Configuration**: Max connections, idle timeout, connection lifetime
  - **Hot-reload**: SIGHUP signal triggers config reload without dropping existing connections
  - **Implementation**: GORM connection pool settings + custom reload handler

- **Database sharding**: Phase 4 implementation (deferred)
  - **Partitioning strategy**: Tenant ID-based partitioning
  - **Shard routing**: Application-level routing based on tenant context
  - **Cross-shard queries**: NOT SUPPORTED - tenant data isolated per shard

- **Caching**: Redis/Memcached NOT USED - SQL-only architecture
  - **Alternative**: PostgreSQL query result caching + prepared statements

**Distributed Caching Strategy**:

- **Cache invalidation**: TTL-based expiration, event-driven invalidation
- **Cache consistency**: Write-through, write-behind, or cache-aside patterns
- **Cache tiers**: L1 (in-memory), L2 (Redis), L3 (database)

**Deployment Patterns**:

- **Blue-Green**: Zero-downtime deployments with instant rollback
- **Canary**: Gradual rollout to subset of users
- **Rolling updates**: Kubernetes-style progressive replacement

**Source**: constitution.md Section VB, clarify.md Q17

### Backup and Recovery

**Database Backup**:

- **PostgreSQL**: `pg_dump` for logical backups, `pg_basebackup` for physical backups
- **SQLite**: File-based backups (copy .db file)
- **Backup frequency**: Daily automated backups, retain 30 days
- **Backup validation**: Test restore procedure monthly

**Disaster Recovery**:

- **Database migrations**: Embedded SQL with golang-migrate provides schema versioning
- **Key rotation**: Version-based key management (KeyRing pattern) enables key recovery
- **Configuration backups**: YAML configs stored in version control
- **Recovery procedure**: Restore database from backup + apply migrations + restore keys from backup

**Documented in**:

- .github/instructions/01-06.database.instructions.md (database migrations)
- .github/instructions/01-09.cryptography.instructions.md (key versioning and rotation)

**Source**: constitution.md Section VB, clarify.md Q18

### Observability

**Telemetry Forwarding**:

- All telemetry forwarded through otel-contrib sidecar (MANDATORY)
- Application: OTLP gRPC:4317 or HTTP:4318 → otel-collector → Grafana OTLP:14317/14318
- Collector self-monitoring: Internal → Grafana OTLP HTTP:14318

**Sampling Strategy** (Source: SPECKIT-CLARIFY-QUIZME-05 Q14, 2025-12-24):

**MANDATORY**: OTLP Collector configuration MUST include ALL sampling strategies as commented options.

**Tail-based Sampling** (REQUIRED - uncommented by default):

```yaml
# otel-collector-config.yaml
processors:
  # Tail-based sampling (ACTIVE)
  tail_sampling:
    policies:
      - name: sample-errors
        type: status_code
        status_code: {status_codes: [ERROR]}
      - name: sample-slow
        type: latency
        latency: {threshold_ms: 1000}
      - name: sample-random
        type: probabilistic
        probabilistic: {sampling_percentage: 10}

  # Head-based sampling (COMMENTED - alternative strategy)
  # probabilistic_sampler:
  #   sampling_percentage: 10

  # Attribute-based sampling (COMMENTED - alternative strategy)
  # attribute_sampler:
  #   sample_by_attribute:
  #     - key: http.status_code
  #       values: ["500", "502", "503"]
```

**Configuration Strategy**:

1. **Default**: Tail-based sampling active (sample errors, slow requests, 10% random)
2. **Commented Alternatives**: Head-based, attribute-based sampling available for operator customization
3. **Documentation**: Explain trade-offs in config comments

**Rationale**: Tail-based sampling provides intelligent sampling (errors + slow requests always captured), while preserving operator flexibility to switch strategies.

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q14

**Telemetry Data Retention and Privacy** (Source: CLARIFY-QUIZME-01 Q6, 2025-12-22):

- **Default Retention**: 90 days
- **Default Redaction**: None (full observability preferred)
- **Configurable Redaction**: Operators MAY configure custom redaction patterns per deployment for compliance (GDPR, CCPA)
- **Rationale**: Full observability aids troubleshooting and forensics; compliance requirements vary by deployment

**Resource Limits for OTLP Collector**:

- Memory limit: 512Mi
- CPU limit: 500m (0.5 CPU cores)
- Sampling strategy: Adaptive based on throughput (100% at low load, 10% at high load)

**Source**: clarify.md Q15, Q8.1

### Security

**mTLS Certificate Revocation** (Source: SPECKIT-CLARIFY-QUIZME-05 Q1, 2025-12-24):

**MANDATORY**: mTLS deployments MUST support BOTH CRL Distribution Points (CRLDP) and OCSP for certificate revocation.

**CRLDP Requirements**:

- **Distribution**: One serial number per HTTPS URL with base64-url-encoded serial (e.g., `https://ca.example.com/crl/EjOrvA.crl`)
- **Encoding**: Serial numbers MUST be base64-url-encoded (RFC 4648) - uses `-_` instead of `+/`, no padding `=`
- **Signing**: CRLs MUST be signed by issuing CA before publication
- **Availability**: CRLs MUST be available immediately after revocation (NOT batched/delayed)
- **Format**: DER-encoded CRL per RFC 5280
- **Example**: Certificate serial `0x123ABC` → base64-url encode → `EjOrvA` → `https://ca.example.com/crl/EjOrvA.crl`

**OCSP Requirements**:

- **Responder**: MUST implement RFC 6960 OCSP responder
- **Response Time**: <1 second for cached responses, <5 seconds for database lookups
- **Stapling**: Nice-to-have (NOT blocking) - server-side OCSP stapling reduces client latency

**Revocation Check Priority**:

1. Check OCSP if available (faster, real-time status)
2. Fall back to CRLDP if OCSP unavailable
3. Fail-closed: Reject certificate if BOTH unavailable (strict mode) or fail-open (permissive mode, configurable)

**Configuration**:

```yaml
mtls:
  revocation_check: both  # Options: ocsp, crl, both
  fail_mode: closed       # Options: closed (reject on check failure), open (allow on check failure)
  ocsp_timeout: 5s
  crl_cache_ttl: 3600s    # Cache CRLs for 1 hour
```

**Rationale**: CRLDP provides guaranteed availability (HTTP GET), OCSP provides real-time status. BOTH required for production resilience.

**Docker Secrets**:

- **File permissions**: 400 (r--------) or 440 (r--r-----) (read-only for owner or owner+group)
- **Dockerfile validation**: ALL Dockerfiles MUST include validation stage to verify secrets exist with correct permissions
- **Pattern**: See KMS Dockerfile validator stage (alpine:3.19 AS validator)

**Source**: .github/instructions/02-02.docker.instructions.md, clarify.md Q9.1

**Unseal Secrets** (Source: SPECKIT-CLARIFY-QUIZME-05 Q2, 2025-12-24):

**MANDATORY**: Services MUST support BOTH key derivation (HKDF) and pre-generated JWKs for unseal secrets.

**Key Derivation Mode (HKDF)**:

```yaml
unseal:
  mode: derive
  master_key: file:///run/secrets/master_key  # 32-byte master key
  derivation_info: "cryptoutil-kms-v1"        # Application-specific context
  algorithm: HKDF-SHA256
```

**Pre-generated JWK Mode**:

```yaml
unseal:
  mode: jwk
  jwk_path: file:///run/secrets/unseal_jwk    # Pre-generated JWK (RSA-4096, EC P-384, etc.)
```

**Configuration-Driven Selection**:

- **Development**: Use HKDF derivation for simplicity (single master key)
- **Production**: Use pre-generated JWKs for HSM integration and key rotation
- **Hybrid**: Support both modes simultaneously (multi-key unlock)

**Implementation Requirements**:

```go
type UnsealConfig struct {
    Mode           string `yaml:"mode"`            // "derive" or "jwk"
    MasterKey      string `yaml:"master_key"`     // For HKDF mode
    DerivationInfo string `yaml:"derivation_info"` // For HKDF mode
    JWKPath        string `yaml:"jwk_path"`       // For JWK mode
}
```

**Rationale**: HKDF simplifies key management for single-key scenarios, pre-generated JWKs enable HSM integration and advanced key rotation strategies.

### Multi-Tenancy

**Tenant Isolation** (Source: constitution.md, plan.md, 2025-12-24):

**REQUIRED**: Dual-layer tenant isolation for defense-in-depth:

**Layer 1: Per-Row Tenant ID** (PostgreSQL + SQLite):

- ALL tables MUST have `tenant_id UUID NOT NULL` column
- `tenant_id` is foreign key to `tenants.id` (UUIDv4)
- ALL queries MUST filter by `WHERE tenant_id = $1`
- Enforced at application layer (SQL query construction)
- Works on BOTH PostgreSQL and SQLite

**Layer 2: Schema-Level Isolation** (PostgreSQL ONLY):

- Each tenant gets separate schema: `CREATE SCHEMA tenant_<UUID>`
- Connection sets search_path: `SET search_path TO tenant_<UUID>`
- Provides database-level isolation for PostgreSQL deployments
- NOT applicable to SQLite (no schema support)

**Architecture**:

```sql
-- Layer 1: Per-row tenant_id (PostgreSQL + SQLite)
CREATE TABLE users (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id),
  email TEXT NOT NULL,
  UNIQUE(tenant_id, email)
);

-- Layer 2: Schema-level (PostgreSQL only)
CREATE SCHEMA tenant_a;
CREATE SCHEMA tenant_b;
SET search_path TO tenant_a;  -- All queries scoped to tenant_a
```

**Configuration**:

```yaml
multi_tenancy:
  isolation: dual-layer  # per-row tenant_id + schema-level (PostgreSQL only)
  tenant_id_header: X-Tenant-ID
  auto_create_schema: true  # PostgreSQL only, auto-provision schemas
```

**Rationale**:

- Layer 1 (per-row tenant_id): Works on PostgreSQL + SQLite, mandatory defense
- Layer 2 (schema-level): PostgreSQL-only enhancement, additional isolation
- Dual-layer provides defense-in-depth (application-level + database-level)
- NEVER use row-level security (RLS) - layers 1+2 sufficient
- NOT SUPPORTED: Separate databases per tenant (connection pool exhaustion)

**Runtime Tenant Switching**:

```go
// Set PostgreSQL search_path per request
db.Exec("SET search_path TO ?, public", tenantID)
```

**Rationale**: Schema isolation provides database-level isolation without separate database connections, balancing security and resource efficiency.

**Source**: SPECKIT-CLARIFY-QUIZME-05 Q6

### Certificate Profiles

**Custom Certificate Profiles** (Source: CLARIFY-QUIZME-01 Q5, 2025-12-22):

- **24 Predefined Profiles**: Cover most use cases (DV, OV, EV, code signing, etc.)
- **YAML-Based Extensibility**: Organizations can define custom profiles via YAML configuration files
- **Runtime Loading**: Profiles loaded at startup from configuration directory
- **No Database/Plugin Support**: File-based configuration strikes balance between flexibility and simplicity

**Standard Profiles**:

- **DV (Domain Validation)**: domain_only validation, 90 days validity
- **OV (Organization Validation)**: organization validation, 397 days validity
- **EV (Extended Validation)**: extended validation, 397 days validity

**Per-client configuration**: Client can request specific profile based on trust requirements

**Policy enforcement**: CA policy engine enforces profile constraints

**Source**: clarify.md Q10.3

---

## Known Gaps and Future Work

### High Priority

1. **Identity Admin API Migration**: Implement dual-server pattern (Public HTTPS + Private HTTPS) matching KMS architecture
   - Add `/admin/api/v1/livez`, `/admin/api/v1/readyz` endpoints
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

## Clarifications

### Session 2025-12-23

**Q1: When circuit breaker opens after 5 failures, does retry mechanism continue running?**

A: Stop retrying immediately - fail-fast until half-open state after 60s timeout

- Circuit breaker states: Closed (normal), Open (fail-fast), Half-Open (testing)
- Retry mechanisms ONLY active in Closed and Half-Open states
- Open state: All requests fail immediately without retry attempts
- After timeout (60s), transition to Half-Open for testing
- Prevents resource exhaustion and cascading failures

**Q2: How do multiple services avoid admin port collisions in unified deployments?**

A: Containerization requirement - each container has isolated localhost namespace, non-containerized unified deployments not supported

- Admin ports fixed at 127.0.0.1:9090 for all services
- Containerization REQUIRED for multi-service deployments (each container isolates localhost)
- Non-containerized unified deployments NOT SUPPORTED (would cause port collisions)
- Single-service standalone deployments can run non-containerized
- Rationale: Container isolation enables consistent admin port across all services

**Q3: What determines session token format selection between opaque, JWE, and JWS?**

A: Configuration-driven per service deployment - admin configures via YAML, default opaque for browser/JWS for headless, all three formats must be supported

- Format selection: Administrator-configured via YAML deployment configuration
- Default behavior: Opaque tokens for browser-based clients, JWS tokens for headless clients
- Mandatory support: All services MUST implement all three formats (opaque, JWE, JWS)
- Rationale: Enables deployment flexibility, security/performance tradeoffs per environment
- See: Session Token Format section and [.github/instructions/02-10.authentication.instructions.md](../../.github/instructions/02-10.authentication.instructions.md) for configuration examples

**Q4: When service transitions from non-federated to federated mode, how are existing sessions handled?**

A: Grace period dual-format support - accept BOTH formats during transition (e.g., 24h), old tokens expire naturally, new tokens issued for new logins

- Grace period: Accept both old-format (non-federated) and new-format (federated) tokens during transition
- Default grace period: 24 hours (configurable)
- Old token handling: Expire naturally according to their TTL (no forced invalidation)
- New token issuance: New logins immediately receive federated-format tokens
- Prevents: Service disruption and forced user re-authentication
- See: Federation Configuration section for migration configuration examples

**Q5: How should introspection results be cached?**

A: Cache positive results with configurable TTL, cache negative results for 1 minute - provides operational flexibility while maintaining security

- Positive results (active=true): Configurable TTL per deployment (default: match token expiry)
- Negative results (active=false): Fixed 1-minute cache to prevent abuse and reduce load
- Rationale: Positive caching reduces load on authorization server, negative caching prevents attackers from overwhelming introspection endpoint
- Configuration example:

  ```yaml
  introspection:
    cache_positive_ttl: 3600  # 1 hour (or match token TTL)
    cache_negative_ttl: 60    # 1 minute (fixed)
  ```

- See: Token Validation section for implementation details

---

*Specification Version: 1.2.0*
*Last Updated: December 23, 2025*
