# Passthru2: Implementation & Improvement Plan

**Purpose**: Apply lessons learned from `passthru1` to rework demos, solidify best practices, and implement demo parity and developer experience improvements.
**Created**: 2025-11-30
**Updated**: 2025-11-30 (incorporated Grooming Session 1 answers)
**Status**: IN PROGRESS

---

## Summary of Decisions from Grooming Session 1

### Vision & Strategy

| Decision | Answer | Notes |
|----------|--------|-------|
| **Primary Goals** | B+C: DX/Demo UX + Repo reorganization | Focus on one-command demos and infra/product split |
| **Demo Parity** | C+D: KMS & Identity equal + Integration | All demos must have consistent UX |
| **Timeline** | A: Aggressive (1-2 weeks) | High velocity approach |
| **Breaking Changes** | B: Minor breaks OK | Maintain compatibility where possible |
| **Audience** | A: Contributors/Developers + LLM Agents | Personal use and AI tooling |

### Infrastructure & Deployment

| Decision | Answer | Notes |
|----------|--------|-------|
| **Telemetry** | A: Centralized in `deployments/telemetry/compose.yml` | Single source for telemetry |
| **Config Model** | A+C: Product config + Docker Secrets | `deployments/<product>/config/` + secrets |
| **Compose Profiles** | D: All (dev, demo, ci) | Full profile support |
| **Telemetry Default** | B: Telemetry ON by default | Parity with CI/demo |
| **Secret Management** | A: Docker Secrets everywhere | No environment variables for secrets |

### Products & Parity

| Decision | Answer | Notes |
|----------|--------|-------|
| **Pre-seeded Accounts** | A: Yes for all products | KMS needs realm-based auth (file/DB) |
| **Demo Scripts** | B+D: Go CLI + Docker Compose | NO Makefiles, NO bash/PowerShell |
| **KMS Parity** | D: All features | Priority: Swagger UI > Auto-seed > CLI |
| **JOSE Authority** | B: Phase 3 - future work | Not in immediate scope |
| **Embedded Identity** | B: No - external only | Avoid circular dependencies |

### Design & Security

| Decision | Answer | Notes |
|----------|--------|-------|
| **FIPS Compliance** | D: Need security audit | Ban bcrypt - PBKDF2 only |
| **Token Management** | C: Introspection + Local JWT | Configurable, mixed approach |
| **Scope Granularity** | C: Hybrid (coarse + fine) | `kms:admin` + `kms:encrypt` |
| **Audit Logging** | D: Minimal now, compliance later | Iterative approach |
| **Database Patterns** | C: SQLite (unit) + Postgres (integration) | Mixed testing strategy |

### Tests, CI & Migration

| Decision | Answer | Notes |
|----------|--------|-------|
| **Coverage Targets** | C: 80% minimum | Iterative improvement |
| **Migration Strategy** | C: Hybrid (low-risk infra first) | Move infra, then products in batches |
| **E2E Test Location** | D: All locations | Product, cross-product, and root |
| **CI Changes** | D: All (demo runs, coverage, matrix) | Comprehensive CI improvements |
| **Acceptance Criteria** | A+B+C+D+E: ALL must be true | Full acceptance criteria |

---

## KMS Authentication Strategy (from Q11)

KMS will support two identity/authentication/authorization strategies:

### Strategy 1: Simple Realm (File/DB)

For standalone KMS deployments without external Identity:

```yaml
# File Realm (config-mounted, for sqlite in-memory mode)
realms:
  file:
    enabled: true
    users:
      admin: { role: admin }
      tenant1-admin: { role: tenant-admin, tenant: tenant1 }
      tenant1-user: { role: user, tenant: tenant1 }
      tenant1-service: { role: service, tenant: tenant1 }

# DB Realm (for postgres mode - shared across instances)
realms:
  native:
    enabled: true
    storage: postgres
```

### Strategy 2: Federation to Identity

For multi-tenant KMS with external Identity deployments:

```yaml
# Each Identity deployment is authority for one or more tenants
identity-providers:
  - issuer: https://identity-1.example.com
    tenants: [tenant1, tenant2]
  - issuer: https://identity-2.example.com
    tenants: [tenant3]
```

---

## Demo Script Priorities (from Q12)

**BANNED**: Makefiles, Bash scripts, PowerShell scripts

**Priority Order**:

1. **Docker Compose with health checks** - Primary demo standardization
2. **Go CLI (per-product)** - `cmd/demo-kms`, `cmd/demo-identity`
3. **Go CLI (federation)** - `cmd/demo-all` orchestrating all products

---

## Deliverables

| Deliverable | Status | Location |
|-------------|--------|----------|
| README.md (this file) | ✅ Updated | `docs/03-products/passthru2/` |
| TASK-LIST.md | 🔄 Needs update | `docs/03-products/passthru2/` |
| DEMO-KMS.md | 🔄 Needs update | `docs/03-products/passthru2/` |
| DEMO-IDENTITY.md | 🔄 Needs update | `docs/03-products/passthru2/` |
| DEMO-INTEGRATION.md | 🔄 Needs update | `docs/03-products/passthru2/` |
| GROOMING-SESSION-1.md | ✅ Answered | `docs/03-products/passthru2/grooming/` |
| GROOMING-SESSION-2.md | 📝 To prepare | `docs/03-products/passthru2/grooming/` |

---

## Quick Start (After Passthru2 Implementation)

```bash
# KMS Demo (single command)
docker compose -f deployments/telemetry/compose.yml \
               -f deployments/kms/compose.demo.yml up -d

# Identity Demo (single command)
docker compose -f deployments/telemetry/compose.yml \
               -f deployments/identity/compose.demo.yml up -d

# Integration Demo (both products)
docker compose -f deployments/telemetry/compose.yml \
               -f deployments/kms/compose.demo.yml \
               -f deployments/identity/compose.demo.yml up -d

# Or using Go CLI
go run ./cmd/demo-kms
go run ./cmd/demo-identity
go run ./cmd/demo-all
```

---

## Next Steps

1. ✅ Complete GROOMING-SESSION-1.md answers
2. 📝 Prepare GROOMING-SESSION-2.md for deeper technical questions
3. 🔨 Implement Phase 0: Developer Experience foundation
4. 🔨 Implement telemetry extraction
5. 🔨 Implement demo seeding and compose profiles

---

**Status**: IN PROGRESS

