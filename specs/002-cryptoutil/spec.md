# cryptoutil Specification - Iteration 2

## Overview

**Iteration 2** builds on the stable Iteration 1 foundation to deliver:

1. **JOSE Authority** - Standalone JOSE service with full REST API
2. **CA Server REST API** - Certificate lifecycle management via HTTP
3. **Unified Suite** - All 4 products deployable as integrated stack

## Iteration 1 Foundation

Iteration 1 delivered:

- ✅ Identity V2 (OAuth 2.1 AuthZ + OIDC IdP) with UI
- ✅ KMS with hierarchical key management
- ✅ Integration demo (full-stack 7/7 steps)
- ✅ Test infrastructure with 80%+ coverage

---

## P1: JOSE Authority (Standalone Service)

### Current State

JOSE primitives exist in `internal/jose/` (embedded library). Iteration 2 exposes these as a standalone REST API service.

### Target Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    JOSE Authority                        │
├─────────────────────────────────────────────────────────┤
│  /jose/v1/keys          - Key management                │
│  /jose/v1/jwks          - Public JWKS endpoint          │
│  /jose/v1/sign          - JWS signing                   │
│  /jose/v1/verify        - JWS verification              │
│  /jose/v1/encrypt       - JWE encryption                │
│  /jose/v1/decrypt       - JWE decryption                │
│  /jose/v1/jwt/issue     - JWT issuance                  │
│  /jose/v1/jwt/validate  - JWT validation                │
└─────────────────────────────────────────────────────────┘
```

### API Endpoints

| Endpoint | Method | Description | Priority |
|----------|--------|-------------|----------|
| `/jose/v1/keys` | POST | Generate new JWK | HIGH |
| `/jose/v1/keys/{kid}` | GET | Retrieve specific JWK | HIGH |
| `/jose/v1/keys` | GET | List JWKs with filters | HIGH |
| `/jose/v1/jwks` | GET | Public JWKS endpoint | HIGH |
| `/jose/v1/sign` | POST | Create JWS signature | HIGH |
| `/jose/v1/verify` | POST | Verify JWS signature | HIGH |
| `/jose/v1/encrypt` | POST | Create JWE encryption | MEDIUM |
| `/jose/v1/decrypt` | POST | Decrypt JWE payload | MEDIUM |
| `/jose/v1/jwt/issue` | POST | Issue JWT with claims | HIGH |
| `/jose/v1/jwt/validate` | POST | Validate JWT | HIGH |

### Supported Algorithms

| Type | Algorithms | FIPS Status |
|------|-----------|-------------|
| Signing | PS256, PS384, PS512, RS256, RS384, RS512, ES256, ES384, ES512, EdDSA | ✅ Approved |
| Key Wrapping | RSA-OAEP, RSA-OAEP-256, A128KW, A192KW, A256KW | ✅ Approved |
| Content Encryption | A128GCM, A192GCM, A256GCM, A128CBC-HS256, A192CBC-HS384, A256CBC-HS512 | ✅ Approved |
| Key Agreement | ECDH-ES, ECDH-ES+A128KW, ECDH-ES+A192KW, ECDH-ES+A256KW | ✅ Approved |

### Implementation Tasks

| ID | Task | Description | LOE |
|----|------|-------------|-----|
| JOSE-1 | OpenAPI Spec | Create `api/jose/openapi_spec_*.yaml` | 2h |
| JOSE-2 | Server Scaffolding | Fiber server with middleware | 4h |
| JOSE-3 | Key Handler | Generate/list/get JWK endpoints | 4h |
| JOSE-4 | JWKS Handler | Public JWKS endpoint | 2h |
| JOSE-5 | Sign Handler | JWS creation endpoint | 3h |
| JOSE-6 | Verify Handler | JWS verification endpoint | 3h |
| JOSE-7 | Encrypt Handler | JWE creation endpoint | 3h |
| JOSE-8 | Decrypt Handler | JWE decryption endpoint | 3h |
| JOSE-9 | JWT Issue Handler | JWT issuance with claims | 4h |
| JOSE-10 | JWT Validate Handler | JWT validation endpoint | 4h |
| JOSE-11 | Integration Tests | E2E tests for all endpoints | 6h |
| JOSE-12 | Docker Compose | Add jose-server to compose.yml | 2h |

**Total LOE**: ~40 hours

---

## P4: CA Server REST API

### Current State

CA internal components exist in `internal/ca/`. Iteration 2 exposes REST API for external certificate operations.

### Target Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    CA Server                             │
├─────────────────────────────────────────────────────────┤
│  /ca/v1/health          - Health check                  │
│  /ca/v1/ca              - List/get CAs                  │
│  /ca/v1/certificate     - Issue/revoke/status           │
│  /ca/v1/ocsp            - OCSP responder                │
│  /ca/v1/profiles        - Certificate profiles          │
│  /ca/v1/est/*           - EST protocol endpoints        │
│  /ca/v1/tsa/timestamp   - RFC 3161 timestamps           │
└─────────────────────────────────────────────────────────┘
```

### API Endpoints

| Endpoint | Method | Description | Priority |
|----------|--------|-------------|----------|
| `/ca/v1/health` | GET | Health check | HIGH |
| `/ca/v1/ca` | GET | List available CAs | HIGH |
| `/ca/v1/ca/{ca_id}` | GET | Get CA details + chain | HIGH |
| `/ca/v1/ca/{ca_id}/crl` | GET | Download current CRL | HIGH |
| `/ca/v1/certificate` | POST | Issue cert from CSR | HIGH |
| `/ca/v1/certificate/{serial}` | GET | Get certificate | MEDIUM |
| `/ca/v1/certificate/{serial}/revoke` | POST | Revoke certificate | HIGH |
| `/ca/v1/certificate/{serial}/status` | GET | Certificate status | MEDIUM |
| `/ca/v1/ocsp` | POST | OCSP responder | HIGH |
| `/ca/v1/profiles` | GET | List profiles | MEDIUM |
| `/ca/v1/profiles/{profile_id}` | GET | Get profile details | MEDIUM |
| `/ca/v1/est/cacerts` | GET | EST: Get CA certs | MEDIUM |
| `/ca/v1/est/simpleenroll` | POST | EST: Simple enrollment | MEDIUM |
| `/ca/v1/est/simplereenroll` | POST | EST: Re-enrollment | LOW |
| `/ca/v1/est/serverkeygen` | POST | EST: Server keygen | LOW |
| `/ca/v1/tsa/timestamp` | POST | RFC 3161 timestamp | MEDIUM |

### Authentication Methods

| Method | Use Case |
|--------|----------|
| mTLS | Client certificate (primary) |
| JWT Bearer | Delegated from Identity Server |
| API Key | Automated systems (with IP allowlist) |

### Implementation Tasks

| ID | Task | Description | LOE |
|----|------|-------------|-----|
| CA-1 | OpenAPI Spec | Create `api/ca/openapi_spec_paths.yaml` | 4h |
| CA-2 | Server Scaffolding | Fiber server with mTLS | 4h |
| CA-3 | Health Handler | Health check endpoint | 1h |
| CA-4 | CA Handler | List/get CA endpoints | 4h |
| CA-5 | CRL Handler | CRL download endpoint | 3h |
| CA-6 | Certificate Issue | CSR-based issuance | 6h |
| CA-7 | Certificate Get | Retrieve by serial | 2h |
| CA-8 | Certificate Revoke | Revocation endpoint | 4h |
| CA-9 | Certificate Status | Status query endpoint | 2h |
| CA-10 | OCSP Handler | RFC 6960 responder | 6h |
| CA-11 | Profile Handlers | Profile list/get | 3h |
| CA-12 | EST cacerts | EST CA certs endpoint | 2h |
| CA-13 | EST simpleenroll | EST enrollment | 4h |
| CA-14 | EST simplereenroll | EST re-enrollment | 3h |
| CA-15 | EST serverkeygen | Server-side keygen | 3h |
| CA-16 | TSA Handler | RFC 3161 timestamps | 4h |
| CA-17 | Integration Tests | E2E tests | 8h |
| CA-18 | Docker Compose | Add ca-server to compose | 2h |

**Total LOE**: ~65 hours

---

## Unified Suite Architecture

### Deployment Topology

```
┌─────────────────────────────────────────────────────────────────┐
│                         Load Balancer                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐│
│  │   JOSE      │  │  Identity   │  │    KMS      │  │   CA    ││
│  │  Authority  │  │   Server    │  │   Server    │  │  Server ││
│  │  :8083      │  │  :8080      │  │  :8081      │  │  :8084  ││
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └────┬────┘│
│         │                │                │               │      │
│         └────────────────┴────────────────┴───────────────┘      │
│                               │                                   │
│                        ┌──────┴──────┐                           │
│                        │  PostgreSQL │                           │
│                        │    :5432    │                           │
│                        └─────────────┘                           │
└─────────────────────────────────────────────────────────────────┘
```

### Service Ports

| Service | Public Port | Admin Port | Description |
|---------|-------------|------------|-------------|
| identity-server | 8080 | 9090 | OAuth 2.1 + OIDC |
| kms-server | 8081 | 9091 | Key Management |
| jose-authority | 8083 | 9093 | JOSE Operations |
| ca-server | 8084 | 9094 | Certificate Authority |
| postgres | 5432 | - | Database |
| otel-collector | 4317/4318 | 13133 | Telemetry |
| grafana-otel-lgtm | 3000 | - | Observability UI |

### Implementation Tasks

| ID | Task | Description | LOE |
|----|------|-------------|-----|
| UNIFIED-1 | compose.yml | Update with all 4 services | 4h |
| UNIFIED-2 | Shared Secrets | Unified unseal secret config | 2h |
| UNIFIED-3 | Service Discovery | Inter-service communication | 4h |
| UNIFIED-4 | Health Checks | All services report healthy | 2h |
| UNIFIED-5 | E2E Demo | Full 4-product demo script | 6h |
| UNIFIED-6 | Documentation | Architecture and runbooks | 4h |

**Total LOE**: ~22 hours

---

## Quality Requirements

### Coverage Targets

| Category | Target |
|----------|--------|
| JOSE Authority | ≥80% |
| CA Server | ≥80% |
| Unified Suite | E2E passing |

### Testing Strategy

- Unit tests with `t.Parallel()`
- Integration tests with Docker Compose
- Load tests for JOSE/CA hot paths

---

## Timeline

### Phase 1: JOSE Authority (Week 1-2)

- JOSE-1 through JOSE-12
- Target: Fully functional JOSE Authority

### Phase 2: CA Server (Week 3-5)

- CA-1 through CA-18
- Target: Certificate lifecycle via REST

### Phase 3: Unified Suite (Week 6)

- UNIFIED-1 through UNIFIED-6
- Target: All 4 products deployable

---

## Success Criteria

| Criterion | Metric |
|-----------|--------|
| JOSE Authority | All 10 endpoints functional |
| CA Server | All 16 endpoints functional |
| Unified Suite | `docker compose up` healthy |
| Coverage | ≥80% for new code |
| Tests | All pass concurrently (`-shuffle=on`) |

---

## Dependencies

| Dependency | Status |
|------------|--------|
| Iteration 1 Identity | ✅ Complete |
| Iteration 1 KMS | ✅ Complete |
| CA Internal Components | ✅ Complete |
| JOSE Primitives | ✅ Complete |

---

*Specification Version: 2.0.0*
*Created: January 15, 2025*
*Last Updated: January 15, 2025*
