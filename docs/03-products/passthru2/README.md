# Passthru2: Implementation & Improvement Plan

**Purpose**: Apply lessons learned from `passthru1` to rework demos, solidify best practices, and implement demo parity and developer experience improvements.
**Created**: 2025-11-30
**Updated**: 2025-11-30 (incorporated Grooming Sessions 1, 2 & 3 answers)
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

## Summary of Decisions from Grooming Session 3

### TLS/HTTPS Implementation (Q1-5)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Cert Utility Location** | C: `internal/infra/tls/` | Create new shared package |
| **CA Chain Length** | D: Configurable, default 3 | Root→Policy→Issuing→Leaf |
| **Certificate CNs** | C+D: FQDN + configurable | `kms.cryptoutil.demo.local` style |
| **mTLS Mode** | A+D: Required + configurable | mTLS for all internal, configurable per pair |
| **Cert Rotation** | D: Defer to passthru3 | Auto-rotation planned later |

### Realm Configuration Details (Q6-10)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Realm File Location** | A: Same directory as config | Simple co-location |
| **Password Hash Format** | Configurable | Digest, iterations, salt length configurable |
| **User Schema** | A+B+C+D: All fields | Full flexibility with JSON metadata |
| **Role Definition** | B+C+D: Configurable + hierarchical | Mapped from Identity when federated |
| **Tenant ID Format** | **UUIDv4** | Special case: max randomness, unpredictable |

### Demo CLI Implementation (Q11-15)

| Decision | Answer | Notes |
|----------|--------|-------|
| **CLI Architecture** | A+C: Single binary + library | `demo kms`, `demo identity`, `demo all` |
| **Output Format** | A+B+C+D: All formats | Human, JSON, structured logging |
| **Failure Handling** | B+D: Continue + configurable | Report summary at end |
| **Health Timeout** | B: 30s default configurable | Reasonable default |
| **Data Verification** | A: Verify all entities | Query and validate after startup |

### Token Validation Implementation (Q16-20)

| Decision | Answer | Notes |
|----------|--------|-------|
| **JWKS Cache** | No preference | Choose appropriate library |
| **Introspection Batching** | A+B+C: Single + batch + dedup | Support both patterns |
| **Error Response** | D: Hybrid | OAuth for auth, Problem Details otherwise |
| **Scope Parsing** | B: Structured parser | With validation |
| **Claims Propagation** | A+B: Typed struct + OIDC | Custom context key |

### Testing & Quality (Q21-25)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Test Factories** | A+B: Both patterns | testutil + per-package helpers |
| **Benchmark Targets** | A+B+C+D: All | Crypto, tokens, DB, APIs |
| **E2E Parallelization** | A+C+D: Sequential default | Parallel for independent, configurable |
| **Coverage Reporting** | D: All formats | Native, HTML, Codecov |
| **Integration Timeout** | B+D: 60s configurable | Per-test timeout |

---

## Summary of Decisions from Grooming Session 4

### TLS Infrastructure (Q1-5)

| Decision | Answer | Notes |
|----------|--------|-------|
| **TLS Dependencies** | A+B+D: Std lib + x/crypto, minimal | No ACME for passthru2 |
| **Cert Storage** | A+B+D: PEM + PKCS#12 configurable | PEM default, PKCS#11/YubiKey future |
| **Root CA Trust** | A: Custom CA only | 99% always custom CAs |
| **Validation Strictness** | A: Full validation ALWAYS | CRITICAL: No relaxed modes |
| **TLS Version** | B: TLS 1.3 only | Best security |

### UUIDv4 Tenant ID (Q6-10)

| Decision | Answer | Notes |
|----------|--------|-------|
| **UUIDv4 Generation** | C: Match existing v7 pattern | Use keygen consistency |
| **Tenant Validation** | A: Strict UUID format only | Standard format |
| **Display Format** | A: Full UUID with hyphens | Consistent display |
| **Demo Tenants** | B: Random per startup | Regenerated each run |
| **Tenant in Header** | ALWAYS Authorization header | Never path/query params |

### Demo CLI Error Handling (Q11-15)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Error Aggregation** | B+C: Structured + tree | apperr-style nested |
| **Partial Success** | A+B+D: Report + keep running | Graceful degradation |
| **Retry Strategy** | D: Configurable | Exponential backoff option |
| **Progress Display** | A+C: Counter + spinner | Dual progress indicators |
| **Exit Codes** | C+D: sysexits.h or 0/1/2 | Need clarification |

### Benchmark & Coverage (Q16-20)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Benchmark Baseline** | C+D: Previous run + CI | Store baseline locally |
| **Test Fixtures** | B: testdata/ directories | Standard Go convention |
| **Test Isolation** | C+D: Tx rollback + UUIDv7 | Dual isolation strategy |

### Config & Deployment (Q21-25)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Config Format** | A+D: YAML primary | Converter for others |
| **Config Validation** | A+B+D: Load + startup | Fail fast |
| **Config Location** | C: Standard paths | /etc, ~/.config |
| **Hot Reload** | D: Defer to passthru3 | Maybe C if easy |
| **Compose Profiles** | B: dev, demo, ci + prod | Production template |

---

## Summary of Decisions from Grooming Session 5

### Clarifications (Q1-5)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Exit Code Strategy** | A: Simple (0/1/2) | 0=success, 1=partial, 2=failure |
| **PKCS#11/YubiKey** | B+D: Extensible API + placeholder | `internal/infra/tls/hsm/` placeholder |
| **System Trust Store** | A: Feature flag (disabled) | Design for future enablement |
| **Auth Priority** | C: Both in parallel | Bearer + Basic in different paths |
| **TLS Client SAN** | A: URI SAN with tenant | Consider alternatives for tenant representation |

### Implementation Priorities (Q6-10)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Phase 0 First Task** | Priority: P0.6 → P0.10 → P0.1 → P0.5 | Demo seed → TLS pkg → Telemetry → Profiles |
| **TLS Package Scope** | D: Full scope | CA chain + certs + config + mTLS helpers + validation |
| **Config Validation** | A: Same strictness | Demo mode = production strictness |
| **Demo Compose Order** | Use KMS compose.yml order | Match existing patterns |
| **Coverage Baseline** | D: After Phase 0 | Start tracking after Phase 0 complete |

### Technical Deep Dive (Q11-15)

| Decision | Answer | Notes |
|----------|--------|-------|
| **mTLS Cert Rotation** | Priority: FS watcher → Admin API → SIGHUP | Multiple notification options |
| **PBKDF2 Defaults** | A: SHA-256, 600K iterations, 32-byte salt | OWASP 2024 recommendation |
| **User Metadata Schema** | B+C: Optional schema + required top-level keys | Working schema file for validation |
| **Tenant Isolation** | A: Schema-per-tenant | Works for both SQLite and PostgreSQL |
| **Demo CLI Color** | D: All options | Windows ANSI + auto-disable CI + --no-color flag |

### Documentation & Testing (Q16-20)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Demo Docs Format** | D: All formats | Markdown + Mermaid + screenshots |
| **API Documentation** | D: Update if API changes | Observe emergent design |
| **Coverage Exclusions** | A: Only generated code | api/client, api/server excluded |
| **Benchmark Storage** | A+B+C: All formats | JSON + Go bench format + SQLite |
| **Error Message Pattern** | B: RFC 7807 Problem Details | Standardized error format |

### Final Confirmations (Q21-25)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Scope Lock** | C: Open for demo improvements | Additions OK if improve demo UX |
| **Implementation Order** | D: Sequential only | Clean commits, milestone checkpoints |
| **Breaking Changes** | N/A - Unreleased project | No migration needed |
| **Demo Recordings** | D: If time permits | Documentation sufficient |
| **Success Metrics** | A+B: Criteria + coverage | Acceptance criteria + 80%+ coverage |

---

### Auth Strategy from Session 4 (Q10 Notes)

**Authorization Header Mapping Strategy:**

| API Type | Auth Method | Tenant Resolution |
|----------|-------------|-------------------|
| Service APIs | Configurable | Header-linked |
| User UI/APIs | Session cookie (UUIDv7) | Redis cache server-side |

**Authz Provider Options:**

| Method | Token Type | State |
|--------|------------|-------|
| Bearer | UUID access token (UUIDv7/v4) | Stateful (issuer maps to tenant) |
| Bearer | JWT access token | Stateless (tenant in claims) |

**KMS Realm Options:**

| Method | Storage | Notes |
|--------|---------|-------|
| Basic | File realm | Config-mounted, Base64URL encoded |
| Basic | DB realm | Shared across instances |
| Bearer | Federated to Identity | JWT (stateless) or UUID (stateful) |
| TLS Client | Custom SAN extension | URI or other (future) |

---

## CRITICAL FIX: TLS/HTTPS Pattern (from Q20)

| Decision | Answer | Notes |
|----------|--------|-------|
| **CLI Architecture** | A+C: Single binary + library | `demo kms`, `demo identity`, `demo all` |
| **Output Format** | A+B+C+D: All formats | Human, JSON, structured logging |
| **Failure Handling** | B+D: Continue + configurable | Report summary at end |
| **Health Timeout** | B: 30s default configurable | Reasonable default |
| **Data Verification** | A: Verify all entities | Query and validate after startup |

### Token Validation Implementation (Q16-20)

| Decision | Answer | Notes |
|----------|--------|-------|
| **JWKS Cache** | No preference | Choose appropriate library |
| **Introspection Batching** | A+B+C: Single + batch + dedup | Support both patterns |
| **Error Response** | D: Hybrid | OAuth for auth, Problem Details otherwise |
| **Scope Parsing** | B: Structured parser | With validation |
| **Claims Propagation** | A+B: Typed struct + OIDC | Custom context key |

### Testing & Quality (Q21-25)

| Decision | Answer | Notes |
|----------|--------|-------|
| **Test Factories** | A+B: Both patterns | testutil + per-package helpers |
| **Benchmark Targets** | A+B+C+D: All | Crypto, tokens, DB, APIs |
| **E2E Parallelization** | A+C+D: Sequential default | Parallel for independent, configurable |
| **Coverage Reporting** | D: All formats | Native, HTML, Codecov |
| **Integration Timeout** | B+D: 60s configurable | Per-test timeout |

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
| GROOMING-SESSION-3.md | ✅ Answered | `docs/03-products/passthru2/grooming/` |
| GROOMING-SESSION-4.md | ✅ Answered | `docs/03-products/passthru2/grooming/` |
| GROOMING-SESSION-5.md | ✅ Answered | `docs/03-products/passthru2/grooming/` |
| GROOMING-SESSION-6.md | 📝 Awaiting answers | `docs/03-products/passthru2/grooming/` |

---

## Quick Start (After Passthru2 Implementation)

```bash
# KMS Demo (single command)
docker compose -f deployments/telemetry/compose.yml \
               -f deployments/kms/compose.demo.yml up -d

# Identity Demo (single command)
docker compose -f deployments/telemetry/compose.yml \
               -f deployments/identity/compose.simple.yml up -d

# Integration Demo (both products)
docker compose -f deployments/telemetry/compose.yml \
               -f deployments/kms/compose.demo.yml \
               -f deployments/identity/compose.simple.yml up -d

# Or using Go CLI (single binary with subcommands - from Session 3)
go run ./cmd/demo kms
go run ./cmd/demo identity
go run ./cmd/demo all
```

---

## Next Steps

1. ✅ Complete GROOMING-SESSION-1.md answers
2. ✅ Complete GROOMING-SESSION-2.md answers
3. ✅ Complete GROOMING-SESSION-3.md answers
4. ✅ Complete GROOMING-SESSION-4.md answers
5. ✅ Complete GROOMING-SESSION-5.md answers
6. 📝 Answer GROOMING-SESSION-6.md for final implementation decisions
7. 🔨 Continue Phase 1: KMS Demo Parity
8. 🔨 Continue Phase 2: Identity Demo Parity

---

**Status**: IN PROGRESS
