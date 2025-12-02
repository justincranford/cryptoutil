# cryptoutil Specification

## Overview

**cryptoutil** is a Go-based cryptographic services platform providing secure key management, identity services, and certificate authority capabilities with FIPS 140-3 compliance.

## Product Suite

### P1: JOSE (JSON Object Signing and Encryption)

Core cryptographic primitives for web security standards.

#### Capabilities

| Feature | Description | Status |
|---------|-------------|--------|
| JWK | JSON Web Key generation and management | ✅ Implemented |
| JWKS | JSON Web Key Set endpoints | ✅ Implemented |
| JWE | JSON Web Encryption operations | ✅ Implemented |
| JWS | JSON Web Signature operations | ✅ Implemented |
| JWT | JSON Web Token creation and validation | ✅ Implemented |

#### Supported Algorithms

| Algorithm Type | Algorithms | FIPS Status |
|----------------|-----------|-------------|
| Signing | RS256, RS384, RS512, ES256, ES384, ES512, PS256, PS384, PS512, EdDSA | ✅ Approved |
| Key Wrapping | RSA-OAEP, RSA-OAEP-256, A128KW, A192KW, A256KW | ✅ Approved |
| Content Encryption | A128GCM, A192GCM, A256GCM, A128CBC-HS256, A192CBC-HS384, A256CBC-HS512 | ✅ Approved |
| Key Agreement | ECDH-ES, ECDH-ES+A128KW, ECDH-ES+A192KW, ECDH-ES+A256KW | ✅ Approved |

---

### P2: Identity (OAuth 2.1 Authorization Server + OIDC IdP)

Complete identity and access management solution.

#### Authorization Server (AuthZ)

| Endpoint | Method | Description | Status |
|----------|--------|-------------|--------|
| `/oauth2/v1/authorize` | GET/POST | Authorization code flow with mandatory PKCE | ✅ Working |
| `/oauth2/v1/token` | POST | Token exchange (code, refresh, client_credentials) | ✅ Working |
| `/oauth2/v1/introspect` | POST | Token introspection (RFC 7662) | ✅ Working |
| `/oauth2/v1/revoke` | POST | Token revocation (RFC 7009) | ✅ Working |
| `/.well-known/openid-configuration` | GET | OpenID Connect Discovery | ✅ Working |
| `/.well-known/jwks.json` | GET | JSON Web Key Set | ✅ Working |

#### Identity Provider (IdP)

| Endpoint | Method | Description | Status |
|----------|--------|-------------|--------|
| `/oidc/v1/login` | GET/POST | User authentication | ⚠️ API Only (No UI) |
| `/oidc/v1/consent` | GET/POST | User consent for scopes | ⚠️ API Only (No UI) |
| `/oidc/v1/logout` | GET/POST | Session termination | ⚠️ Partial |
| `/oidc/v1/userinfo` | GET | User information endpoint | ⚠️ Partial |

#### Authentication Methods

| Method | Description | Status |
|--------|-------------|--------|
| client_secret_basic | HTTP Basic Auth with client_id:client_secret | ✅ Working |
| client_secret_post | client_id and client_secret in request body | ✅ Working |
| client_secret_jwt | JWT signed with client secret | ⚠️ Not Tested |
| private_key_jwt | JWT signed with private key | ❌ Not Implemented |

#### MFA Factors

| Factor | Description | Status |
|--------|-------------|--------|
| TOTP | Time-based One-Time Password | ✅ Working |
| Passkey | WebAuthn/FIDO2 authentication | ⚠️ Partial |
| Email OTP | One-time password via email | ❌ Not Implemented |
| SMS OTP | One-time password via SMS | ❌ Not Implemented |

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

#### ElasticKey Operations

| Operation | Method | Endpoint | Description |
|-----------|--------|----------|-------------|
| Create | POST | `/api/v1/elastic-keys` | Create new elastic key with policy |
| Read | GET | `/api/v1/elastic-keys/{id}` | Get elastic key by ID |
| List | GET | `/api/v1/elastic-keys` | List with filtering, sorting, pagination |
| Update | PUT | `/api/v1/elastic-keys/{id}` | Update elastic key metadata |
| Delete | DELETE | `/api/v1/elastic-keys/{id}` | Soft delete elastic key |

#### MaterialKey Operations

| Operation | Method | Endpoint | Description |
|-----------|--------|----------|-------------|
| Create | POST | `/api/v1/elastic-keys/{id}/material-keys` | Create new version |
| Read | GET | `/api/v1/material-keys/{id}` | Get material key by ID |
| List | GET | `/api/v1/elastic-keys/{id}/material-keys` | List versions |
| Revoke | POST | `/api/v1/material-keys/{id}/revoke` | Revoke key version |
| Import | POST | `/api/v1/elastic-keys/{id}/import` | Import external key |

#### Key Hierarchy

```
Unseal Secrets (file:///run/secrets/*)
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
| `elastic_key_id` | Filter by UUID |
| `name` | Filter by key name |
| `provider` | Filter by key provider |
| `algorithm` | Filter by algorithm |
| `status` | Filter by status (active, suspended, deleted) |
| `versioning_allowed` | Filter by versioning policy |
| `import_allowed` | Filter by import policy |

#### Sorting Parameters

| Parameter | Direction |
|-----------|-----------|
| `name` | asc/desc |
| `created_at` | asc/desc |
| `updated_at` | asc/desc |
| `status` | asc/desc |

---

### P4: Certificates (Certificate Authority) - PLANNED

X.509 certificate lifecycle management with CA/Browser Forum compliance.

#### Planned Capabilities

| Task | Description | Priority |
|------|-------------|----------|
| 1. Domain Charter | Scope definition, compliance mapping | HIGH |
| 2. Config Schema | YAML schema for crypto, subject, certificate profiles | HIGH |
| 3. Crypto Providers | RSA, ECDSA, EdDSA, HMAC, future PQC | HIGH |
| 4. Subject Profile Engine | Template resolution for subject details, SANs | HIGH |
| 5. Certificate Profile Engine | 20+ profile archetypes | HIGH |
| 6. Root CA Bootstrap | Offline root CA creation | HIGH |
| 7. Intermediate CA Provisioning | Subordinate CA hierarchy | HIGH |
| 8. Issuing CA Lifecycle | Rotation, monitoring, status reporting | MEDIUM |
| 9. Enrollment API | REST API for CSR submission, issuance | HIGH |
| 10. Revocation Services | CRL generation, OCSP responders | HIGH |
| 11. Time-Stamping | RFC 3161 TSA functionality | MEDIUM |
| 12. RA Workflows | Registration authority for validation | MEDIUM |
| 13. Profile Library | 20+ predefined certificate profiles | HIGH |
| 14. Storage Layer | PostgreSQL/SQLite with ACID guarantees | HIGH |
| 15. CLI Tooling | bootstrap, issuance, revocation commands | MEDIUM |
| 16. Observability | OTLP metrics, Grafana dashboards | MEDIUM |
| 17. Security Hardening | STRIDE threat modeling, HSM planning | HIGH |
| 18. Compliance | CA/Browser Forum audit readiness | HIGH |
| 19. Deployment | Docker Compose, Kubernetes manifests | MEDIUM |
| 20. Handover | Runbooks, training, DR drills | LOW |

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

- HTTPS with TLS 1.2+ minimum
- HTTP/2 support via Fiber framework
- CORS, CSRF protection
- Rate limiting per IP

### I3: Testing

- Table-driven tests with `t.Parallel()`
- Coverage targets: 80% production, 85% infrastructure, 95% utility
- Fuzz testing, benchmark testing, integration testing

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

- PostgreSQL (production)
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

| Service | Public Port | Admin Port | Backend |
|---------|-------------|------------|---------|
| cryptoutil-sqlite | 8080 | 9090 | SQLite in-memory |
| cryptoutil-postgres-1 | 8081 | 9090 | PostgreSQL |
| cryptoutil-postgres-2 | 8082 | 9090 | PostgreSQL |
| identity-authz | 8090 | - | SQLite/PostgreSQL |
| identity-idp | 8091 | - | SQLite/PostgreSQL |
| postgres | 5432 | - | - |
| otel-collector | 4317/4318 | 13133 | - |
| grafana-otel-lgtm | 3000 | - | Loki/Tempo/Prometheus |

### Health Endpoints

| Endpoint | Purpose |
|----------|---------|
| `/livez` | Liveness probe (admin API) |
| `/readyz` | Readiness probe (admin API) |
| `/health` | Public health check |
| `/ui/swagger/doc.json` | OpenAPI specification |

---

*Specification Version: 1.0.0*
*Last Updated: December 2025*
