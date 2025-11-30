# DEMO-IDENTITY: Identity-Only Working Demo (Passthru2)

**Purpose**: Stabilize Identity demo and provide feature parity for KMS demo experiences
**Priority**: HIGH
**Timeline**: Day 3-5
**Updated**: 2025-11-30 (aligned with Grooming Session 1 decisions)

---

## Differences vs Passthru1

- Prioritize demo seeding and a `demo` mode for Identity
- Fix missing flows (authorize, PKCE, refresh rotation)
- Seed clients and users with the same standard used in KMS demo
- Go CLI demo orchestration (NO bash/PowerShell scripts - banned per Q12)

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

### Token Endpoints

- [ ] Complete token introspection tests
- [ ] Complete token revocation tests

---

## Demo Accounts

```yaml
# Demo users
users:
  demo-admin:
    email: admin@demo.local
    password: demo-admin-password
    roles: [admin]
  demo-user:
    email: user@demo.local
    password: demo-user-password
    roles: [user]
  demo-service:
    email: service@demo.local
    password: demo-service-password
    roles: [service]

# Demo clients
clients:
  demo-public-client:
    type: public
    redirect_uris: [http://localhost:3000/callback]
    pkce_required: true
  demo-confidential-client:
    type: confidential
    client_secret: demo-client-secret
    scopes: [openid, profile, email]
```

---

## Quick Start Commands

```bash
# Option 1: Docker Compose (primary - Q12 priority 1)
docker compose -f deployments/telemetry/compose.yml \
               -f deployments/identity/compose.demo.yml up -d

# Option 2: Go CLI (Q12 priority 2)
go run ./cmd/demo-identity

# Verify health
curl -k https://localhost:8080/livez
curl -k https://localhost:8080/readyz

# Discovery endpoint
curl -k https://localhost:8080/.well-known/openid-configuration

# JWKS endpoint
curl -k https://localhost:8080/.well-known/jwks.json
```

---

## Success Criteria

- [ ] Docker compose `demo` starts Identity with seeded users and clients
- [ ] Discovery endpoint returns valid config and JWKS
- [ ] Authorization code + PKCE + token exchange flow works end-to-end
- [ ] Token introspection and revocation validated with demo scripts

---

**Status**: IN PROGRESS
