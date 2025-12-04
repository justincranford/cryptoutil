# cryptoutil Specification

## Overview

**cryptoutil** is a Go-based cryptographic services platform providing secure key management, identity services, and certificate authority capabilities with FIPS 140-3 compliance.

## Product Suite

### P1: JOSE (JSON Object Signing and Encryption)

Core cryptographic primitives for web security standards. Serves as the embedded foundation for all other products.

**Architecture Note**: JOSE is embedded in all products (P2-P4). Current implementation is in `internal/common/crypto/jose/` and needs refactoring to `internal/product/jose/` as a standalone JOSE Authority service.

#### Capabilities

| Feature | Description | Status |
|---------|-------------|--------|
| JWK | JSON Web Key generation and management | ✅ Implemented |
| JWKS | JSON Web Key Set endpoints | ✅ Implemented |
| JWE | JSON Web Encryption operations | ✅ Implemented |
| JWS | JSON Web Signature operations | ✅ Implemented |
| JWT | JSON Web Token creation and validation | ✅ Implemented |
| JOSE Authority | Standalone JOSE service with full API | ⚠️ Needs Refactoring |

#### Supported Algorithms

| Algorithm Type | Algorithms | FIPS Status |
|----------------|-----------|-------------|
| Signing | PS256, PS384, PS512, RS256, RS384, RS512, ES256, ES384, ES512, EdDSA | ✅ Approved |
| Key Wrapping | RSA-OAEP, RSA-OAEP-256, A128KW, A192KW, A256KW | ✅ Approved |
| Content Encryption | A128GCM, A192GCM, A256GCM, A128CBC-HS256, A192CBC-HS384, A256CBC-HS512 | ✅ Approved |
| Key Agreement | ECDH-ES, ECDH-ES+A128KW, ECDH-ES+A192KW, ECDH-ES+A256KW | ✅ Approved |

---

### P2: Identity (OAuth 2.1 Authorization Server + OIDC IdP)

Complete identity and access management solution.

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
| `/device_authorization` | POST | Device Authorization Grant (RFC 8628) | ❌ Not Required |
| `/par` | POST | Pushed Authorization Requests (RFC 9126) | ❌ Not Required |

#### Identity Provider (IdP)

| Endpoint | Method | Description | Status |
|----------|--------|-------------|--------|
| `/oidc/v1/login` | GET/POST | User authentication | ✅ Working (HTML form rendered, session created) |
| `/oidc/v1/consent` | GET/POST | User consent for scopes | ✅ Working (HTML form rendered, consent recorded) |
| `/oidc/v1/logout` | GET/POST | Session termination | ✅ Working (session/token cleared) |
| `/oidc/v1/endsession` | GET | OpenID Connect End Session (RP-Initiated Logout) | ✅ Working |
| `/oidc/v1/userinfo` | GET | User information endpoint | ✅ Working (claims returned per scopes, JWT-signed optional) |
| `/oidc/v1/mfa/enroll` | POST | Administrative Enroll MFA factor | ❌ Not Implemented |
| `/oidc/v1/mfa/factors` | GET | Administrative List user MFA factors | ❌ Not Implemented |
| `/oidc/v1/mfa/factors/{id}` | DELETE | Administrative Remove MFA factor | ❌ Not Required |

#### Authentication Methods

| Method | Description | Status |
|--------|-------------|--------|
| client_secret_basic | HTTP Basic Auth with client_id:client_secret | ✅ Working |
| client_secret_post | client_id and client_secret in request body | ✅ Working |
| client_secret_jwt | JWT signed with client secret | ⚠️ Partial (HIGH Priority) |
| private_key_jwt | JWT signed with private key | ⚠️ Partial (HIGH Priority) |
| tls_client_auth | Mutual TLS client certificate authentication | ❌ Not Implemented |
| self_signed_tls_client_auth | Self-signed TLS client certificate authentication | ❌ Not Implemented |
| session_cookie | Browser session cookie for SPA UI | ❌ Not Implemented (Required) |

#### MFA Factors

| Factor | Description | Status |
|--------|-------------|--------|
| TOTP | Time-based One-Time Password | ✅ Working (HIGH Priority) |
| Passkey | WebAuthn/FIDO2 authentication | ✅ Working (HIGH Priority) |
| Hardware Security Keys | Dedicated hardware tokens (U2F/FIDO) | ❌ Not Implemented (HIGH Priority) |
| Email OTP | One-time password via email | ⚠️ Partial |
| SMS OTP | One-time password via SMS | ⚠️ Partial |
| HOTP | HMAC-based One-Time Password (counter-based) | ❌ Not Implemented |
| Recovery Codes | Backup codes for account recovery | ❌ Not Implemented |
| Push Notifications | Push-based authentication via mobile app | ❌ Not Required |
| Phone Call OTP | One-time password via voice call | ❌ Not Required |

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

**Authentication Strategy**: Configurable - support multiple methods including OAuth 2.1 federation to Identity (P2), API key, mTLS. Dual API exposure:

- `/browser/api/v1/` - User-to-browser APIs for SPA invocation
- `/service/api/v1/` - Service-to-service APIs

**Realm Configuration**: MANDATORY configurable realms for users and clients (file-based and database-based), with OPTIONAL federation to external IdPs and AuthZs.

#### ElasticKey Operations

| Operation | Method | Endpoint | Status |
|-----------|--------|----------|--------|
| Create | POST | `/elastickey` | ✅ Implemented |
| Read | GET | `/elastickey/{elasticKeyID}` | ✅ Implemented |
| List | GET | `/elastickeys` | ✅ Implemented |
| Update | PUT | `/elastickey/{elasticKeyID}` | ❌ Not Implemented |
| Delete | DELETE | `/elastickey/{elasticKeyID}` | ❌ Not Implemented |

#### MaterialKey Operations

| Operation | Method | Endpoint | Status |
|-----------|--------|----------|--------|
| Create | POST | `/elastickey/{elasticKeyID}/materialkey` | ✅ Implemented |
| Read | GET | `/elastickey/{elasticKeyID}/materialkey/{materialKeyID}` | ✅ Implemented |
| List | GET | `/elastickey/{elasticKeyID}/materialkeys` | ✅ Implemented |
| Global List | GET | `/materialkeys` | ✅ Implemented |
| Import | POST | `/elastickey/{elasticKeyID}/import` | ❌ Not Implemented |
| Revoke | POST | `/elastickey/{elasticKeyID}/materialkey/{materialKeyID}/revoke` | ❌ Not Implemented |

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

X.509 certificate lifecycle management with CA/Browser Forum compliance.

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

**Implementation Progress**: 20/20 tasks complete (100%)

#### Compliance Requirements

| Standard | Requirement |
|----------|-------------|
| RFC 5280 | X.509 certificate format and validation |
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
- Coverage targets: 80% production, 85% infrastructure, 95% utility
- Fuzz testing, benchmark testing, integration testing

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
- Coverage targets: 80% production, 85% infrastructure, 95% utility
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

### I8: Containers

- Docker Compose deployments
- Service mesh: cryptoutil, postgres, otel-collector, grafana-otel-lgtm
- Health checks via wget (Alpine containers)

### I9: Deployment

- GitHub Actions CI/CD
- Act for local workflow testing
- Multi-stage Docker builds with static linking

---

## Quality Requirements

### Code Coverage Targets

| Category | Target | Current |
|----------|--------|---------|
| Production Code | ≥80% | Varies |
| Infrastructure (cicd) | ≥85% | ~85% |
| Utility Code | ≥95% | ~95% |

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

| Service | Public Port | Admin Port | Backend |
|---------|-------------|------------|---------|
| identity-authz | 8090 | - | SQLite/PostgreSQL |
| identity-idp | 8091 | - | SQLite/PostgreSQL |

#### P3: KMS Services

| Service | Public Port | Admin Port | Backend |
|---------|-------------|------------|---------|
| kms-sqlite | 8080 | 9090 | SQLite in-memory |
| kms-postgres-1 | 8081 | 9090 | PostgreSQL |
| kms-postgres-2 | 8082 | 9090 | PostgreSQL |

#### P4: CA Services (Planned)

| Service | Public Port | Admin Port | Backend |
|---------|-------------|------------|---------|
| ca-sqlite | 8080 | 9090 | SQLite in-memory |
| ca-postgres-1 | 8081 | 9090 | PostgreSQL |
| ca-postgres-2 | 8082 | 9090 | PostgreSQL |

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

| Product | Endpoint | Purpose |
|---------|----------|---------|
| JOSE | `/livez` | Liveness probe |
| JOSE | `/readyz` | Readiness probe |
| Identity | `/livez` | Liveness probe |
| Identity | `/readyz` | Readiness probe |
| KMS | `/livez` | Liveness probe |
| KMS | `/readyz` | Readiness probe |
| CA | `/livez` | Liveness probe (planned) |
| CA | `/readyz` | Readiness probe (planned) |

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

*Specification Version: 1.1.0*
*Last Updated: January 2026*
