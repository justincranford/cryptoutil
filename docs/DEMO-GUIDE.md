# Product Demo Guide

This guide documents how to demonstrate each cryptoutil product individually and as a federated suite.

## Prerequisites

- Docker Desktop installed and running
- Go 1.25.5+ installed
- Git repository cloned

## Quick Commands Reference

```bash
# Build all Docker images
docker compose -f deployments/jose/compose.yml build
docker compose -f deployments/identity/compose.yml build
docker compose -f deployments/kms/compose.yml build
docker compose -f deployments/ca/compose.simple.yml build

# Run Go demos
go run ./cmd/demo jose
go run ./cmd/demo identity
go run ./cmd/demo kms
go run ./cmd/demo ca
go run ./cmd/demo all
```

---

## P1: JOSE Authority Demo

### Overview

The JOSE Authority provides standalone JSON Object Signing and Encryption operations:

- JWK generation and management
- JWS signing and verification
- JWE encryption and decryption
- JWT creation and verification

### Start JOSE Server (Docker)

```bash
# Start JOSE server
docker compose -f deployments/jose/compose.yml up -d

# Verify health
curl -k https://localhost:8080/health

# View logs
docker compose -f deployments/jose/compose.yml logs -f jose-server

# Stop
docker compose -f deployments/jose/compose.yml down
```

### Run JOSE Demo (Go)

```bash
# Run the JOSE demo
go run ./cmd/demo jose
```

**Expected Output:**

```text
🚀 Starting JOSE Authority Demo...
✅ Step 1: Generate EC key pair (ES256)
✅ Step 2: Sign payload with JWS
✅ Step 3: Verify JWS signature
✅ Step 4: Encrypt payload with JWE
✅ Step 5: Decrypt JWE ciphertext
✅ Demo completed successfully!
Duration: ~2s
Steps: 5 total, 5 passed, 0 failed
```

### JOSE Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/jose/v1/jwk/generate` | POST | Generate new JWK |
| `/jose/v1/jwk/{kid}` | GET | Retrieve JWK by ID |
| `/jose/v1/jwk` | GET | List all JWKs |
| `/jose/v1/jwks` | GET | JWKS endpoint |
| `/jose/v1/jws/sign` | POST | Sign payload |
| `/jose/v1/jws/verify` | POST | Verify JWS |
| `/jose/v1/jwe/encrypt` | POST | Encrypt payload |
| `/jose/v1/jwe/decrypt` | POST | Decrypt JWE |

---

## P2: Identity Server Demo

### Overview

The Identity Server provides OAuth 2.1 and OpenID Connect 1.0 capabilities:

- Authorization endpoint with PKCE
- Token endpoint (authorization_code, client_credentials, refresh_token)
- Userinfo endpoint
- Discovery endpoints (OIDC, OAuth AS)
- Login and consent UI

### Start Identity Server (Docker)

```bash
# Start Identity server with PostgreSQL
docker compose -f deployments/identity/compose.yml up -d

# Verify health
curl -k https://localhost:8080/health

# Check OIDC discovery
curl -k https://localhost:8080/.well-known/openid-configuration

# View logs
docker compose -f deployments/identity/compose.yml logs -f identity-authz

# Stop
docker compose -f deployments/identity/compose.yml down -v
```

### Run Identity Demo (Go)

```bash
# Run the Identity demo
go run ./cmd/demo identity
```

**Expected Output:**

```text
🚀 Starting Identity Server Demo...
✅ Step 1: Bootstrap demo client
✅ Step 2: Obtain token (client_credentials)
✅ Step 3: Validate token structure
✅ Step 4: Introspect token
✅ Step 5: Refresh token
✅ Demo completed successfully!
Duration: ~3s
Steps: 5 total, 5 passed, 0 failed
```

### Identity Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/.well-known/openid-configuration` | GET | OIDC Discovery |
| `/.well-known/oauth-authorization-server` | GET | OAuth AS Metadata |
| `/oidc/v1/authorize` | GET | Authorization endpoint |
| `/oidc/v1/token` | POST | Token endpoint |
| `/oidc/v1/userinfo` | GET/POST | Userinfo endpoint |
| `/oidc/v1/introspect` | POST | Token introspection |
| `/oidc/v1/revoke` | POST | Token revocation |
| `/oidc/v1/jwks` | GET | JWKS endpoint |

---

## P3: KMS Server Demo

### Overview

The KMS Server provides key management and cryptographic operations:

- ElasticKey management (create, get, list, delete)
- Encryption/decryption with managed keys
- Digital signatures (sign/verify)
- Key rotation and versioning

### Start KMS Server (Docker)

```bash
# Start KMS server
docker compose -f deployments/kms/compose.yml up -d

# Verify health
curl -k https://localhost:8080/health

# View logs
docker compose -f deployments/kms/compose.yml logs -f kms-server

# Stop
docker compose -f deployments/kms/compose.yml down
```

### Run KMS Demo (Go)

```bash
# Run the KMS demo
go run ./cmd/demo kms
```

**Expected Output:**

```text
🚀 Starting KMS Server Demo...
✅ Step 1: Create ElasticKey (AES-256-GCM)
✅ Step 2: Encrypt data with ElasticKey
✅ Step 3: Decrypt data with ElasticKey
✅ Step 4: Create signing key (ES256)
✅ Demo completed successfully!
Duration: ~2s
Steps: 4 total, 4 passed, 0 failed
```

### KMS Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/kms/keys` | POST | Create ElasticKey |
| `/api/v1/kms/keys/{kid}` | GET | Get ElasticKey |
| `/api/v1/kms/keys` | GET | List ElasticKeys |
| `/api/v1/kms/encrypt` | POST | Encrypt data |
| `/api/v1/kms/decrypt` | POST | Decrypt data |
| `/api/v1/kms/sign` | POST | Sign data |
| `/api/v1/kms/verify` | POST | Verify signature |

---

## P4: CA Server Demo

### Overview

The CA Server provides certificate authority operations:

- Certificate issuance
- Certificate revocation
- CRL generation
- OCSP responder
- EST protocol endpoints

### Start CA Server (Docker)

```bash
# Start CA server
docker compose -f deployments/ca/compose.simple.yml up -d

# Verify health
curl -k https://localhost:8080/health

# View logs
docker compose -f deployments/ca/compose.simple.yml logs -f ca-server

# Stop
docker compose -f deployments/ca/compose.simple.yml down
```

### Run CA Demo (Go)

```bash
# Run the CA demo
go run ./cmd/demo ca
```

**Expected Output:**

```text
🚀 Starting CA Server Demo...
✅ Step 1: List available CAs
✅ Step 2: Generate CSR
✅ Step 3: Submit enrollment request
✅ Step 4: Get issued certificate
✅ Step 5: Check OCSP status
✅ Demo completed successfully!
Duration: ~3s
Steps: 5 total, 5 passed, 0 failed
```

### CA Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/ca/ca` | GET | List CAs |
| `/api/v1/ca/ca/{caId}` | GET | Get CA details |
| `/api/v1/ca/ca/{caId}/crl` | GET | Download CRL |
| `/api/v1/ca/enrollments` | POST | Submit enrollment |
| `/api/v1/ca/certificates/{sn}` | GET | Get certificate |
| `/api/v1/ca/certificates/{sn}/revoke` | POST | Revoke certificate |
| `/api/v1/ca/ocsp` | POST | OCSP responder |

---

## Federated Suite Demo

### Overview

The federated suite demonstrates all four products working together:

- P2 Identity provides authentication
- P1 JOSE provides signing/encryption
- P3 KMS manages encryption keys
- P4 CA issues TLS certificates

### Start Full Stack (Docker Compose)

```bash
# Start all services with shared telemetry
docker compose -f deployments/compose.integration.yml up -d

# Verify all services healthy
docker compose -f deployments/compose.integration.yml ps

# View aggregated logs
docker compose -f deployments/compose.integration.yml logs -f

# Stop all services
docker compose -f deployments/compose.integration.yml down -v
```

### Run Full Demo (Go)

```bash
# Run the full integration demo
go run ./cmd/demo all
```

**Expected Output:**

```text
🚀 Starting Full Integration Demo...
✅ Step 1: Bootstrap Identity client
✅ Step 2: Obtain OAuth token from Identity
✅ Step 3: Create KMS encryption key
✅ Step 4: Encrypt data with KMS
✅ Step 5: Sign token with JOSE
✅ Step 6: Request certificate from CA
✅ Step 7: Verify federated flow
✅ Demo completed successfully!
Duration: ~3s
Steps: 7 total, 7 passed, 0 failed
```

---

## Troubleshooting

### Docker Issues

```bash
# Check container status
docker compose ps

# View container logs
docker compose logs <service-name>

# Restart a service
docker compose restart <service-name>

# Clean up everything
docker compose down -v --remove-orphans
docker system prune -f
```

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| Port already in use | Another service on port 8080 | Stop conflicting service or change port |
| Certificate errors | Self-signed certs | Use `-k` flag with curl or configure trust |
| Connection refused | Service not started | Check container status and logs |
| Database errors | Migration issue | Recreate volumes with `down -v` |

### Health Check URLs

| Service | Health Endpoint | Expected Response |
|---------|-----------------|-------------------|
| JOSE | `https://localhost:8080/health` | `{"status":"healthy"}` |
| Identity | `https://localhost:8080/health` | `{"status":"healthy"}` |
| KMS | `https://localhost:8080/health` | `{"status":"healthy"}` |
| CA | `https://localhost:8080/health` | `{"status":"healthy"}` |

---

## Recording Video Demos

### Individual Product Demo Recording

For each product, record:

1. **Start**: `docker compose up -d` and show container starting
2. **Health Check**: curl to health endpoint
3. **Demo Run**: `go run ./cmd/demo <product>`
4. **API Exploration**: Swagger UI at `/ui/swagger/`
5. **Cleanup**: `docker compose down`

### Federated Suite Demo Recording

Record the following sequence:

1. **Start All**: `docker compose -f deployments/compose.integration.yml up -d`
2. **Show All Healthy**: `docker compose ps` showing all green
3. **Run Demo**: `go run ./cmd/demo all`
4. **Show Telemetry**: Grafana dashboard with traces/metrics
5. **API Walkthrough**: Each product's Swagger UI
6. **Cleanup**: `docker compose down -v`

---

*Demo Guide Version: 1.0.0*
*Last Updated: January 2026*
