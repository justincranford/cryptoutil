# DEMO-IDENTITY: Identity-Only Working Demo (Passthru2)

**Purpose**: Stabilize Identity demo and provide feature parity for KMS demo experiences
**Priority**: HIGH
**Timeline**: Day 3-5
**Updated**: 2025-11-30 (aligned with Grooming Sessions 1 & 2 decisions)

---

## Differences vs Passthru1

- Prioritize demo seeding and a `demo` mode for Identity
- Fix missing flows (authorize, PKCE, refresh rotation)
- Seed clients and users with the same standard used in KMS demo
- Go CLI demo orchestration (NO bash/PowerShell scripts - banned per Q12)
- **CRITICAL FIX (Q20)**: Reuse KMS TLS pattern with CA-chained certs

---

## CRITICAL: TLS Pattern Fix (from Q20)

**Problem**: passthru2 mixed HTTPS with HTTP incorrectly.

**Solution**: Identity MUST reuse KMS cert utility functions:

```go
// Identity should import and use KMS cert utilities
import cryptoutilCert "cryptoutil/internal/crypto/cert"

// Use same CA-chained cert generation
certChain, err := cryptoutilCert.GenerateCertChain(cryptoutilCert.ChainConfig{
    ChainLength:   2,  // Root CA → Intermediate CA → Leaf
    LeafCN:        "identity.demo.local",
    ValidityDays:  365,
})
```

**Rules**:

1. Never use self-signed TLS leaf node certs
2. Always use CA-chained certificates
3. Pass config options for cert chain lengths
4. Consistent HTTPS across all services

---

## Key Tasks

### Missing Endpoints

- [ ] Implement `/authorize` endpoint
- [ ] Implement full PKCE validation
- [ ] Implement redirect handling
- [ ] Fix refresh token rotation

### Demo Mode

- [ ] Add `--demo` flag to identity server for auto seeding demo data
- [ ] Seed demo users (admin, user, service)
- [ ] Seed demo clients (public with PKCE, confidential)
- [ ] Add `cmd/demo-identity/main.go` Go CLI (Q12)
- [ ] Implement `--reset-demo` flag for cleanup (Q15)
- [ ] Profile-based persistence: dev=persist, ci=ephemeral (Q12)

### Token Endpoints

- [ ] Complete token introspection tests
- [ ] Complete token revocation tests

---

## Demo Accounts (from Q13-14)

```yaml
# Demo users - predictable passwords documented (Q13)
users:
  demo-admin:
    email: admin@demo.local
    password: demo-admin-password  # WARNING: Demo only
    roles: [admin]
  demo-user:
    email: user@demo.local
    password: demo-user-password   # WARNING: Demo only
    roles: [user]
  demo-service:
    email: service@demo.local
    password: demo-service-password  # WARNING: Demo only
    roles: [service]

# Demo clients - predictable secrets with Docker secrets (Q14)
clients:
  demo-public-client:
    type: public
    redirect_uris: [http://localhost:3000/callback]
    pkce_required: true
  demo-confidential-client:
    type: confidential
    client_secret: demo-client-secret  # Also in Docker secret
    scopes: [openid, profile, email, kms:encrypt, kms:decrypt]
```

---

## Quick Start Commands

```bash
# Option 1: Docker Compose (primary - Q12 priority 1)
docker compose -f deployments/telemetry/compose.yml \
               -f deployments/identity/compose.simple.yml up -d

# Option 2: Go CLI (Q12 priority 2)
go run ./cmd/demo-identity

# Option 3: Reset demo data (Q15)
go run ./cmd/demo-identity --reset-demo

# Verify health (full dependency chain per Q16)
curl -k https://localhost:8082/livez   # Identity on port 8082 per Q19
curl -k https://localhost:8082/readyz

# Discovery endpoint
curl -k https://localhost:8082/.well-known/openid-configuration

# JWKS endpoint
curl -k https://localhost:8082/.well-known/jwks.json
```

---

## Success Criteria

- [ ] Docker compose `demo` starts Identity with seeded users and clients
- [ ] Discovery endpoint returns valid config and JWKS
- [ ] Authorization code + PKCE + token exchange flow works end-to-end
- [ ] Token introspection and revocation validated with demo scripts
- [ ] TLS uses CA-chained certs (reusing KMS cert utilities)
- [ ] Profile-based persistence working (dev=persist, ci=ephemeral)

---

**Status**: IN PROGRESS
