# Passthru2: Implementation & Improvement Plan

**Purpose**: Apply lessons learned from `passthru1` to rework demos, solidify best practices, and implement demo parity and developer experience improvements.
**Created**: 2025-11-30
**Updated**: 2025-11-30 (incorporated Grooming Sessions 1 & 2 answers)
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

## Summary of Decisions from Grooming Session 2

### KMS Realm Implementation (Q1-5)

| Decision | Answer | Notes |
|----------|--------|-------|
| **File Realm Config** | B: Separate `realms.yml` file | External config for flexibility |
| **Password Storage** | B+D: PBKDF2 + plaintext support | PBKDF2 for prod, plaintext OK for demo |
| **DB Realm Schema** | A: New `kms_realm_users` table | Separate from Identity users |
| **Realm Priority** | A+C: File→DB→Federation, configurable | Flexible authentication order |
| **Tenant Isolation** | A: Database-level isolation | Separate schemas/databases per tenant |

### Token Validation Details (Q6-10)

| Decision | Answer | Notes |
|----------|--------|-------|
| **JWKS Caching** | A: In-memory with configurable TTL | Simple, effective caching |
| **Revocation Checks** | A+B+C: Every request OR sensitive-only OR interval | Fully configurable |
| **Error Handling** | C+D: 401/403 split + configurable detail | Proper HTTP semantics |
| **Service Auth** | A+B+C: Client creds + mTLS + API key | Multiple auth options |
| **Claims Extraction** | D: All OIDC + custom claims | Full claim support |

### Demo Data Details (Q11-15)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Demo Key Material** | A: Real functional keys | Fully functional demos |
| **Data Persistence** | D: Profile-based (dev=persist, ci=ephemeral) | Smart persistence |
| **Demo Passwords** | A+D: Predictable + documented | Clear demo credentials |
| **Client Secrets** | A+C: Predictable + Docker secrets | Security even in demo |
| **Data Cleanup** | A: `--reset-demo` flag | Easy reset capability |

### Compose & Deployment (Q16-20)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Health Checks** | D: Full dependency chain | DB + Identity + Telemetry |
| **Network Architecture** | B: Per-product + shared telemetry | Logical network separation |
| **Volume Strategy** | A: Named volumes for all data | Consistent persistence |
| **Port Allocation** | B: Product-specific ports | KMS=8081, Identity=8082 |
| **TLS in Demo** | A: CA-chained certs (CRITICAL FIX) | **Fix HTTPS/HTTP mix!** |

### Testing Strategy (Q21-25)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Integration Test Scope** | D: All (startup + CRUD + full flow) | Comprehensive testing |
| **E2E Environment** | D: Compose for local/CI, Testcontainers for unit | Mixed approach |
| **Test Data Isolation** | A: UUIDv7 unique prefixes | **CRITICAL for parallel tests** |
| **Performance Testing** | A: Basic benchmarks for critical paths | Quality imperative |
| **Test Documentation** | A+B: Coverage reports + test descriptions | Inline documentation |

---

## CRITICAL FIX: TLS/HTTPS Pattern (from Q20)

**Problem Identified**: passthru2 mixed HTTPS with HTTP incorrectly.

**Solution**: Identity MUST reuse KMS's established TLS pattern:

1. Use KMS cert utility functions for CA-chained certificates
2. Never use self-signed TLS leaf node certs
3. Pass config options for cert chain lengths and TLS server/client parameters
4. Consistent HTTPS across all services

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
| GROOMING-SESSION-2.md | ✅ Answered | `docs/03-products/passthru2/grooming/` |
| GROOMING-SESSION-3.md | 📝 Awaiting answers | `docs/03-products/passthru2/grooming/` |

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
2. ✅ Complete GROOMING-SESSION-2.md answers
3. 📝 Answer GROOMING-SESSION-3.md for implementation details
4. 🔨 Implement Phase 0: Developer Experience foundation
5. 🔨 Fix TLS/HTTPS pattern (Identity reuse KMS cert utils)
6. 🔨 Implement demo seeding and compose profiles

---

**Status**: IN PROGRESS
