# DEMO-KMS: KMS-Only Working Demo (Passthru2)

**Purpose**: KMS demo parity and improved DX
**Priority**: HIGHEST
**Timeline**: Day 1-3
**Updated**: 2025-11-30 (aligned with Grooming Session 1 decisions)

---

## Differences vs Passthru1

- Add demo accounts, pre-seeding and a `demo` compose profile
- Ensure Swagger UI uses seeded identities for "Try it out" flows
- Add Go CLI demo orchestration (NO bash/PowerShell scripts - banned per Q12)
- Support realm-based authentication (file/DB) for standalone KMS (Q11)

---

## Key Tasks (Priority Order per Q13)

### Priority 1: Swagger UI "Try it out"

- [ ] Configure Swagger UI to accept demo credentials
- [ ] Pre-seed demo accounts that work with Swagger UI
- [ ] Add step-by-step demo documentation

### Priority 2: Demo Mode Auto-seed

- [ ] Create `--demo` flag for KMS server
- [ ] Auto-seed demo accounts on startup
- [ ] Auto-seed demo key pools and keys

### Priority 3: CLI Demo Orchestration

- [ ] Create `cmd/demo-kms/main.go` Go CLI
- [ ] Implement Docker Compose startup with health checks
- [ ] Execute demo flow: create pool → create key → encrypt → decrypt

---

## Demo Accounts (from Q11)

```yaml
# File realm configuration for demo mode
realms:
  file:
    enabled: true
    users:
      demo-admin:
        password: demo-admin-password
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
```

---

## Quick Start Commands

```bash
# Option 1: Docker Compose (primary - Q12 priority 1)
docker compose -f deployments/telemetry/compose.yml \
               -f deployments/kms/compose.demo.yml up -d

# Option 2: Go CLI (Q12 priority 2)
go run ./cmd/demo-kms

# Verify health
curl -k https://localhost:8080/livez
curl -k https://localhost:8080/readyz

# Open Swagger UI
# https://localhost:8080/ui/swagger/
```

---

## Success Criteria

- [ ] Docker compose `demo` starts KMS with pre-seeded accounts
- [ ] Demo script executes full flow automatically
- [ ] Swagger UI interactive calls work with demo accounts
- [ ] Documentation updated with quickstart and troubleshooting

---

**Status**: IN PROGRESS

