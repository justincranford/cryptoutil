# DEMO-INTEGRATION: KMS + Identity Integrated Demo (Passthru2)

**Purpose**: Ensure KMS and Identity integrate cleanly and have parity in demo UX
**Priority**: HIGH
**Timeline**: Day 5-7
**Updated**: 2025-11-30 (aligned with Grooming Sessions 1 & 2 decisions)

---

## Differences vs Passthru1

- Extract telemetry to shared compose to avoid coupling (Q6)
- Implement per-product `demo` mode with curated seed data
- **NO embedded Identity** - external only to avoid circular deps (Q15)
- Go CLI demo orchestration (NO bash/PowerShell scripts - banned per Q12)
- Full dependency chain health checks (Q16)

---

## Token Validation Strategy (from Q6-10)

KMS will use a **fully configurable approach**:

```yaml
# KMS token validation config
token_validation:
  # Local JWT validation for speed (Q6)
  local:
    enabled: true
    cache_type: in-memory  # Per Q6
    cache_ttl: 5m
    jwks_refresh_interval: 1h

  # Introspection for revocation checks (Q7)
  introspection:
    enabled: true
    # Configurable frequency per Q7
    mode: sensitive-only  # Options: every-request | sensitive-only | interval
    interval: 5m  # Used when mode=interval
    sensitive_operations: [encrypt, decrypt, sign, wrap, unwrap]
    endpoint: https://identity:8082/oauth2/introspect

  # Service-to-service authentication (Q9)
  service_auth:
    method: client-credentials  # Options: client-credentials | mtls | api-key
    client_id: kms-service
    client_secret_file: /run/secrets/kms_client_secret  # Docker secret

  # Error handling (Q8)
  errors:
    detail_level: standard  # Options: minimal | standard | detailed
    # 401 for auth issues (invalid token, expired, bad signature)
    # 403 for scope/permission issues
```

---

## Claims Extraction (from Q10)

Extract ALL OIDC + custom claims:

```yaml
claims_extraction:
  # Standard OIDC claims
  standard:
    - sub       # Subject/User ID
    - iss       # Issuer
    - aud       # Audience
    - exp       # Expiration
    - iat       # Issued At
    - scope     # Scopes

  # Custom claims for multi-tenancy
  custom:
    - tenant_id
    - roles
    - permissions
```

---

## Scope Model (from Q18)

Hybrid scope granularity:

```yaml
# Coarse scopes (for general access)
coarse_scopes:
  - kms:admin     # Full administrative access
  - kms:read      # Read-only access (list, get)
  - kms:write     # Write access (create, update, delete)

# Fine-grained scopes (for specific operations)
fine_scopes:
  - kms:encrypt   # Encryption operations
  - kms:decrypt   # Decryption operations
  - kms:sign      # Signing operations
  - kms:verify    # Verification operations
  - kms:wrap      # Key wrapping operations
  - kms:unwrap    # Key unwrapping operations

# Scope hierarchy (coarse includes fine)
hierarchy:
  kms:admin: [kms:read, kms:write, kms:encrypt, kms:decrypt, kms:sign, kms:verify, kms:wrap, kms:unwrap]
  kms:write: [kms:encrypt, kms:sign, kms:wrap]
  kms:read: [kms:decrypt, kms:verify, kms:unwrap]
```

---

## Health Checks (from Q16)

Full dependency chain verification:

```yaml
health_checks:
  # Liveness - service is running
  livez:
    checks: [process]

  # Readiness - service can handle requests
  readyz:
    checks:
      - database      # DB connection OK
      - identity      # Identity reachable (for token validation)
      - telemetry     # OTLP collector reachable (optional)
    timeout: 5s
    interval: 10s
```

---

## Network Architecture (from Q17)

Per-product networks + shared telemetry:

```yaml
networks:
  # Product-specific networks
  kms-network:
    driver: bridge
  identity-network:
    driver: bridge

  # Shared telemetry network
  telemetry-network:
    driver: bridge
    external: true  # Created by telemetry compose
```

---

## Key Tasks

### Token Validation Middleware (from Q6-10)

- [ ] Implement token validation middleware in KMS
- [ ] Implement local JWT validation with in-memory JWKS caching (Q6)
- [ ] Implement configurable JWKS TTL
- [ ] Implement introspection with configurable frequency (Q7)
- [ ] Support all three auth methods: client-creds, mTLS, API key (Q9)
- [ ] Implement 401/403 error split + configurable detail (Q8)
- [ ] Extract all OIDC + custom claims (Q10)

### Scope Enforcement

- [ ] Implement hybrid scope model
- [ ] Add scope checks to all KMS endpoints
- [ ] Implement scope hierarchy resolution
- [ ] Add comprehensive scope tests

### Integration Demo

- [ ] Create `cmd/demo-all/main.go` Go CLI (Q12)
- [ ] Create integration compose file
- [ ] Implement step-by-step demo script
- [ ] Add `--reset-demo` flag for cleanup (Q15)

---

## Quick Start Commands

```bash
# Docker Compose (primary - Q12 priority 1)
docker compose -f deployments/telemetry/compose.yml \
               -f deployments/identity/compose.simple.yml \
               -f deployments/kms/compose.demo.yml up -d

# Go CLI (Q12 priority 3)
go run ./cmd/demo-all

# Reset demo data (Q15)
go run ./cmd/demo-all --reset-demo

# Demo flow
# 1. Get token from Identity (port 8082 per Q19)
curl -k -X POST https://localhost:8082/oauth2/token \
  -d "grant_type=client_credentials" \
  -d "client_id=demo-confidential-client" \
  -d "client_secret=demo-client-secret" \
  -d "scope=kms:encrypt kms:decrypt"

# 2. Use token with KMS (port 8081 per Q19)
curl -k -X POST https://localhost:8081/api/v1/encrypt \
  -H "Authorization: Bearer <token>" \
  -d '{"plaintext": "hello", "key_id": "demo-key"}'
```

---

## Success Criteria

- [ ] `docker compose up` starts both services and telemetry
- [ ] Full dependency chain health checks pass (Q16)
- [ ] KMS accepts tokens from Identity and enforces scopes correctly
- [ ] Token validation configurable (local/introspection/both)
- [ ] 401/403 error responses correct
- [ ] Integration demo script validates all flows
- [ ] Documentation updated and E2E tests implemented per product
- [ ] All services use CA-chained TLS certs

---

**Status**: IN PROGRESS
