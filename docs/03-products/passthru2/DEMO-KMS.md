# DEMO-KMS: KMS-Only Working Demo (Passthru2)

**Purpose**: KMS demo parity and improved DX
**Priority**: HIGHEST
**Timeline**: Day 1-3
**Updated**: 2025-11-30 (aligned with Grooming Sessions 1 & 2 decisions)

---

## Differences vs Passthru1

- Add demo accounts, pre-seeding and a `demo` compose profile
- Ensure Swagger UI uses seeded identities for "Try it out" flows
- Add Go CLI demo orchestration (NO bash/PowerShell scripts - banned per Q12)
- Support realm-based authentication (file/DB) for standalone KMS (Q11)
- **CRITICAL**: Maintain CA-chained TLS certs pattern (never self-signed leaf certs)

---

## Key Tasks (Priority Order per Q13)

### Priority 1: Swagger UI "Try it out"

- [ ] Configure Swagger UI to accept demo credentials
- [ ] Pre-seed demo accounts that work with Swagger UI
- [ ] Add step-by-step demo documentation

### Priority 2: Demo Mode Auto-seed

- [ ] Create `--demo` flag for KMS server
- [ ] Auto-seed demo accounts on startup
- [ ] Auto-seed demo key pools and keys (real keys per Q11)
- [ ] Implement `--reset-demo` flag for cleanup (Q15)

### Priority 3: CLI Demo Orchestration

- [ ] Create `cmd/demo-kms/main.go` Go CLI
- [ ] Implement Docker Compose startup with health checks
- [ ] Execute demo flow: create pool → create key → encrypt → decrypt

---

## Realm Configuration (from Q1-5)

```yaml
# realms.yml - separate file per Q1
realms:
  priority: [file, db, federation]  # Configurable order per Q4

  file:
    enabled: true
    users:
      demo-admin:
        password: demo-admin-password  # Plaintext OK for demo (Q2)
        # password_hash: pbkdf2$sha256$...  # PBKDF2 for prod (Q2)
        roles: [admin]
      demo-tenant-admin:
        password: demo-tenant-admin-password
        roles: [tenant-admin]
        tenant: demo-tenant
      demo-user:
        password: demo-user-password
        roles: [user]
        tenant: demo-tenant
      demo-service:
        password: demo-service-password
        roles: [service]
        tenant: demo-tenant

  db:
    enabled: false  # PostgreSQL only
    table: kms_realm_users  # Separate from Identity per Q3
```

---

## Demo Data Persistence (from Q12)

```yaml
# Profile-based persistence
profiles:
  dev:
    persistence: true  # SQLite file / PostgreSQL
  demo:
    persistence: true  # For interactive demos
  ci:
    persistence: false  # Ephemeral, re-seeded each run
```

---

## TLS Configuration (CRITICAL from Q20)

**KMS establishes the TLS best practice pattern:**

1. **CA-chained certificates** - never self-signed leaf certs
2. **Cert utility functions** in `internal/crypto/` for reuse
3. **Configurable chain lengths** via config options
4. **TLS server/client cert parameters** pass-through

```yaml
# TLS config pattern
tls:
  enabled: true
  ca_chain_length: 2  # Root CA → Intermediate CA → Leaf
  server_cert:
    common_name: "kms.demo.local"
    validity_days: 365
  client_cert:
    enabled: true  # mTLS support
```

---

## Quick Start Commands

```bash
# Option 1: Docker Compose (primary - Q12 priority 1)
docker compose -f deployments/telemetry/compose.yml \
               -f deployments/kms/compose.demo.yml up -d

# Option 2: Go CLI (Q12 priority 2)
go run ./cmd/demo-kms

# Option 3: Reset demo data (Q15)
go run ./cmd/demo-kms --reset-demo

# Verify health (full dependency chain per Q16)
curl -k https://localhost:8081/livez   # KMS on port 8081 per Q19
curl -k https://localhost:8081/readyz

# Open Swagger UI
# https://localhost:8081/ui/swagger/
```

---

## Success Criteria

- [ ] Docker compose `demo` starts KMS with pre-seeded accounts
- [ ] Demo script executes full flow automatically
- [ ] Swagger UI interactive calls work with demo accounts
- [ ] Documentation updated with quickstart and troubleshooting
- [ ] TLS uses CA-chained certs (never self-signed leaf)
- [ ] Real functional keys created in demo (not placeholders)

---

**Status**: IN PROGRESS
