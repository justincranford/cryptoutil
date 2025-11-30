# DEMO-INTEGRATION: KMS + Identity Integrated Demo (Passthru2)

**Purpose**: Ensure KMS and Identity integrate cleanly and have parity in demo UX
**Priority**: HIGH
**Timeline**: Day 5-7
**Updated**: 2025-11-30 (aligned with Grooming Session 1 decisions)

---

## Differences vs Passthru1

- Extract telemetry to shared compose to avoid coupling (Q6)
- Implement per-product `demo` mode with curated seed data
- **NO embedded Identity** - external only to avoid circular deps (Q15)
- Go CLI demo orchestration (NO bash/PowerShell scripts - banned per Q12)

---

## Token Validation Strategy (from Q17)

KMS will use a **mixed approach** (configurable):

```yaml
# KMS token validation config
token_validation:
  # Local JWT validation for speed (default)
  local:
    enabled: true
    cache_ttl: 5m
    jwks_refresh_interval: 1h
  
  # Introspection for revocation checks
  introspection:
    enabled: true
    check_on_sensitive_ops: true  # encrypt, sign, etc.
    endpoint: https://identity:8080/oauth2/introspect
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
```

---

## Key Tasks

### Token Validation Middleware

- [ ] Implement token validation middleware in KMS
- [ ] Implement local JWT validation with JWKS caching
- [ ] Implement introspection for revocation checks
- [ ] Make validation strategy configurable

### Scope Enforcement

- [ ] Implement hybrid scope model
- [ ] Add scope checks to all KMS endpoints
- [ ] Add comprehensive scope tests

### Integration Demo

- [ ] Create `cmd/demo-all/main.go` Go CLI (Q12)
- [ ] Create integration compose file
- [ ] Implement step-by-step demo script

---

## Quick Start Commands

```bash
# Docker Compose (primary - Q12 priority 1)
docker compose -f deployments/telemetry/compose.yml \
               -f deployments/identity/compose.demo.yml \
               -f deployments/kms/compose.demo.yml up -d

# Go CLI (Q12 priority 3)
go run ./cmd/demo-all

# Demo flow
# 1. Get token from Identity
curl -X POST https://localhost:8080/oauth2/token \
  -d "grant_type=client_credentials" \
  -d "client_id=demo-confidential-client" \
  -d "client_secret=demo-client-secret" \
  -d "scope=kms:encrypt kms:decrypt"

# 2. Use token with KMS
curl -X POST https://localhost:8081/api/v1/encrypt \
  -H "Authorization: Bearer <token>" \
  -d '{"plaintext": "hello", "key_id": "demo-key"}'
```

---

## Success Criteria

- [ ] `docker compose up` starts both services and telemetry
- [ ] KMS accepts tokens from Identity and enforces scopes correctly
- [ ] Integration demo script validates all flows
- [ ] Documentation updated and E2E tests implemented per product

---

**Status**: IN PROGRESS

