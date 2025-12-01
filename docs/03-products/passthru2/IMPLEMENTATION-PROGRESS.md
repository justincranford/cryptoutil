# Passthru2 Implementation Progress

**Purpose**: Track task completion for session recovery - if session crashes, continue from here.
**Created**: 2025-11-30
**Last Updated**: 2025-11-30

---

## Quick Status

| Phase | Status | Progress | Next Task |
|-------|--------|----------|-----------|
| **Phase 0** | 🔄 IN PROGRESS | 11/19 | P0.5 |
| **Phase 1** | ⏳ PENDING | 0/25 | - |
| **Phase 2** | ⏳ PENDING | 0/14 | - |
| **Phase 3** | ⏳ PENDING | 0/24 | - |
| **Phase 4** | ⏳ PENDING | 0/13 | - |
| **Phase 5** | ⏳ PENDING | 0/19 | - |
| **Phase 6** | ⏳ PENDING | 0/12 | - |

---

## Phase 0: Developer Experience Foundation

**Priority**: HIGHEST
**Target**: Day 1-2

### Infrastructure Tasks

| Task | Status | Notes |
|------|--------|-------|
| P0.1 | ✅ | Extract telemetry to `deployments/telemetry/compose.yml` - DONE |
| P0.2 | ✅ | Create `deployments/<product>/config/` structure - DONE: KMS and Identity use config/ and secrets/ |
| P0.3 | ✅ | Convert all secrets to Docker secrets - DONE: Both products now use Docker secrets |
| P0.4 | ✅ | Remove empty directories - DONE: Removed identity/identity/ and identity/postgres/ |
| P0.5 | ⏳ | Create compose profiles: dev, demo, ci |

### Demo Seeding Tasks

| Task | Status | Notes |
|------|--------|-------|
| P0.6 | ⏳ | Add demo seed data for KMS |
| P0.7 | ⏳ | Add demo seed data for Identity |
| P0.8 | ⏳ | Create compose.demo.yml for KMS |
| P0.9 | ⏳ | Create compose.demo.yml for Identity |

### TLS/HTTPS (CRITICAL)

| Task | Status | Notes |
|------|--------|-------|
| P0.10 | ✅ | Create `internal/infra/tls/` package - DONE: config.go, storage.go, chain.go, tls_test.go |
| P0.11 | ✅ | Implement CA chain (configurable, default 3) - DONE: DefaultCAChainLength=3 in chain.go |
| P0.12 | ✅ | Use FQDN style CNs, configurable - DONE: ValidateFQDN(), CNStyle (FQDN/Descriptive) |
| P0.13 | ⏳ | Enable mTLS for internal comms |
| P0.14 | ⏳ | Identity reuses `internal/infra/tls/` |
| P0.15 | ✅ | Use std lib + x/crypto only - DONE: only uses crypto/* and golang.org/x/crypto |
| P0.16 | 🔄 | Support PEM + PKCS#12 storage - PEM done, PKCS#12 placeholder |
| P0.17 | ✅ | Custom CA only for demo - DONE: internal/infra/demo/ with DemoCA, GetDemoCA() |
| P0.18 | ✅ | ALWAYS full TLS validation - DONE: ValidateConfig enforces |
| P0.19 | ✅ | TLS 1.3 only - DONE: MinTLSVersion = tls.VersionTLS13 |

---

## Phase 1: KMS Demo Parity

**Priority**: HIGH
**Target**: Day 2-3

### Swagger UI (Highest Priority)

| Task | Status | Notes |
|------|--------|-------|
| P1.1 | ⏳ | Swagger UI works with demo credentials |
| P1.2 | ⏳ | Interactive demo steps in Swagger |
| P1.3 | ⏳ | Document Swagger UI demo flow |

### Auto-seed Demo Mode

| Task | Status | Notes |
|------|--------|-------|
| P1.4 | ⏳ | Implement --demo flag for KMS |
| P1.5 | ⏳ | Auto-seed key pools |
| P1.6 | ⏳ | Auto-seed encryption keys |
| P1.7 | ⏳ | Implement --reset-demo flag |

### CLI Demo Orchestration

| Task | Status | Notes |
|------|--------|-------|
| P1.8 | ⏳ | Create cmd/demo/main.go single binary |
| P1.9 | ⏳ | Implement demo kms subcommand |
| P1.10 | ⏳ | Support human/JSON/structured output |
| P1.11 | ⏳ | Continue on error, report summary |
| P1.12 | ⏳ | Health check waiting (30s default) |
| P1.13 | ⏳ | Verify all demo entities |
| P1.13a | ⏳ | Structured error aggregation |
| P1.13b | ⏳ | Handle partial success |
| P1.13c | ⏳ | Configurable retry strategy |
| P1.13d | ⏳ | Progress counter + spinner |
| P1.13e | ⏳ | Exit codes (sysexits/0/1/2) |

### KMS Realm Configuration

| Task | Status | Notes |
|------|--------|-------|
| P1.14 | ⏳ | Create realms.yml |
| P1.15 | ⏳ | Configurable PBKDF2 |
| P1.16 | ⏳ | Full user schema with JSON metadata |
| P1.17 | ⏳ | Configurable hierarchical roles |
| P1.18 | ⏳ | UUIDv4 for tenant IDs |
| P1.19 | ⏳ | UUIDv4 generation matching v7 pattern |
| P1.20 | ⏳ | Strict UUID format validation |
| P1.21 | ⏳ | Full UUID display with hyphens |
| P1.22 | ⏳ | Regenerate demo tenants on startup |
| P1.23 | ⏳ | Tenant ID via Authorization header |

### Coverage

| Task | Status | Notes |
|------|--------|-------|
| P1.24 | ⏳ | KMS handler tests (85%) |
| P1.25 | ⏳ | KMS businesslogic tests (85%) |

---

## Phase 2: Identity Demo Parity

**Priority**: HIGH
**Target**: Day 3-5

### Missing Endpoints

| Task | Status | Notes |
|------|--------|-------|
| P2.1 | ⏳ | Implement /authorize endpoint |
| P2.2 | ⏳ | Implement full PKCE validation |
| P2.3 | ⏳ | Implement redirect handling |

### Token Management

| Task | Status | Notes |
|------|--------|-------|
| P2.4 | ⏳ | Fix refresh token rotation |
| P2.5 | ⏳ | Complete introspection tests |
| P2.6 | ⏳ | Complete revocation tests |

### Demo Mode

| Task | Status | Notes |
|------|--------|-------|
| P2.7 | ⏳ | Implement --demo flag for Identity |
| P2.8 | ⏳ | Create cmd/demo-identity/main.go |
| P2.9 | ⏳ | Seed demo users |
| P2.10 | ⏳ | Seed demo clients |
| P2.11 | ⏳ | Implement --reset-demo flag |
| P2.12 | ⏳ | Profile-based persistence |

### Coverage

| Task | Status | Notes |
|------|--------|-------|
| P2.13 | ⏳ | Identity idp/userauth tests (80%) |
| P2.14 | ⏳ | Identity handler tests (80%) |

---

## Phase 3: Integration Demo

**Priority**: HIGH
**Target**: Day 5-7

### Token Validation in KMS

| Task | Status | Notes |
|------|--------|-------|
| P3.1 | ⏳ | Token validation middleware |
| P3.2 | ⏳ | Local JWT validation with JWKS caching |
| P3.3 | ⏳ | Configurable JWKS TTL |
| P3.4 | ⏳ | Introspection for revocation |
| P3.5 | ⏳ | Configurable revocation check frequency |
| P3.6 | ⏳ | 401/403 error split |

### Service-to-Service Auth

| Task | Status | Notes |
|------|--------|-------|
| P3.7 | ⏳ | Client credentials auth |
| P3.8 | ⏳ | mTLS auth |
| P3.9 | ⏳ | API key auth |
| P3.10 | ⏳ | Configurable auth method |

### Claims & Scopes

| Task | Status | Notes |
|------|--------|-------|
| P3.11 | ⏳ | Extract OIDC + custom claims |
| P3.12 | ⏳ | Hybrid scope model |
| P3.13 | ⏳ | Coarse scopes |
| P3.14 | ⏳ | Fine scopes |
| P3.15 | ⏳ | Scope enforcement tests |

### Integration Demo

| Task | Status | Notes |
|------|--------|-------|
| P3.16 | ⏳ | demo identity subcommand |
| P3.17 | ⏳ | demo all subcommand |
| P3.18 | ⏳ | Integration compose file |
| P3.19 | ⏳ | Demo script (token → KMS op) |

### Token Validation Implementation

| Task | Status | Notes |
|------|--------|-------|
| P3.20 | ⏳ | JWKS cache implementation |
| P3.21 | ⏳ | Single + batch introspection |
| P3.22 | ⏳ | Hybrid error responses |
| P3.23 | ⏳ | Structured scope parser |
| P3.24 | ⏳ | Typed claims struct |

---

## Phase 4: KMS Realm Authentication

**Priority**: MEDIUM
**Target**: Day 7-9

### File Realm

| Task | Status | Notes |
|------|--------|-------|
| P4.1 | ⏳ | Realm configuration schema |
| P4.2 | ⏳ | File realm loader |
| P4.3 | ⏳ | Basic auth for file realm |
| P4.4 | ⏳ | File realm tests |

### DB Realm

| Task | Status | Notes |
|------|--------|-------|
| P4.5 | ⏳ | kms_realm_users table schema |
| P4.6 | ⏳ | Native realm repository |
| P4.7 | ⏳ | DB realm tests |

### Tenant Isolation

| Task | Status | Notes |
|------|--------|-------|
| P4.8 | ⏳ | Database-level tenant isolation |
| P4.9 | ⏳ | Separate schemas per tenant |
| P4.10 | ⏳ | Tenant isolation tests |

### Federation

| Task | Status | Notes |
|------|--------|-------|
| P4.11 | ⏳ | Identity provider federation config |
| P4.12 | ⏳ | Multi-tenant authority mapping |
| P4.13 | ⏳ | Federation tests |

---

## Phase 5: CI & Quality Gates

**Priority**: MEDIUM
**Target**: Day 9-10

### Coverage Gates

| Task | Status | Notes |
|------|--------|-------|
| P5.1 | ⏳ | Coverage threshold enforcement (80%) |
| P5.2 | ⏳ | Per-package coverage reporting |
| P5.3 | ⏳ | Coverage trend tracking |
| P5.3a | ⏳ | Benchmark baseline storage |
| P5.3b | ⏳ | Compare previous run |
| P5.3c | ⏳ | CI regression detection |

### Demo CI Jobs

| Task | Status | Notes |
|------|--------|-------|
| P5.4 | ⏳ | Demo profile CI job for KMS |
| P5.5 | ⏳ | Demo profile CI job for Identity |
| P5.6 | ⏳ | Integration demo CI job |

### Database Matrix

| Task | Status | Notes |
|------|--------|-------|
| P5.7 | ⏳ | SQLite test runs in CI |
| P5.8 | ⏳ | PostgreSQL test runs in CI |

### Testing Improvements

| Task | Status | Notes |
|------|--------|-------|
| P5.9 | ⏳ | testutil package with factories |
| P5.10 | ⏳ | Per-package factories |
| P5.11 | ⏳ | UUIDv7 unique prefixes (CRITICAL) |
| P5.12 | ⏳ | Basic benchmarks |
| P5.13 | ⏳ | 60s configurable integration timeout |
| P5.14 | ⏳ | Test case descriptions |

### Config & Deployment

| Task | Status | Notes |
|------|--------|-------|
| P5.15 | ⏳ | YAML primary config format |
| P5.16 | ⏳ | Config validation at load + startup |
| P5.17 | ⏳ | Standard config paths |
| P5.18 | ⏳ | Defer hot reload |
| P5.19 | ⏳ | Add prod profile to compose |

---

## Phase 6: Migration & Cleanup

**Priority**: LOW
**Target**: Day 10-14

### Infra Package Migration

| Task | Status | Notes |
|------|--------|-------|
| P6.1 | ⏳ | Move apperr to internal/infra/ |
| P6.2 | ⏳ | Move config to internal/infra/ |
| P6.3 | ⏳ | Move magic to internal/infra/ |
| P6.4 | ⏳ | Move telemetry to internal/infra/ |
| P6.5 | ⏳ | Create internal/infra/tls/ |
| P6.6 | ⏳ | Support chain length 3 |
| P6.7 | ⏳ | Support FQDN and descriptive CNs |
| P6.8 | ⏳ | Implement mTLS required + fallback |
| P6.9 | ⏳ | Handle clock skew in dev mode |

### Product Package Migration

| Task | Status | Notes |
|------|--------|-------|
| P6.10 | ⏳ | Consolidate Identity duplicate code |
| P6.11 | ⏳ | Update all import paths |
| P6.12 | ⏳ | Verify build + test + lint |

---

## Acceptance Criteria

| Criteria | Status | Notes |
|----------|--------|-------|
| A | ⏳ | KMS and Identity demos start with docker compose |
| B | ⏳ | Interactive demo scripts and Swagger UI |
| C | ⏳ | Integration demo validates token-based auth |
| D | ⏳ | All tests pass with 80%+ coverage |
| E | ⏳ | Telemetry extracted, secrets standardized |
| F | ⏳ | TLS/HTTPS pattern fixed |
| G | ⏳ | Single demo binary with subcommands |
| H | ⏳ | UUIDv4 for tenant IDs |
| I | ⏳ | 60s configurable integration timeout |
| J | ⏳ | TLS 1.3 only, full validation, PEM/PKCS#12 |
| K | ⏳ | Demo CLI with structured errors, progress, exit codes |
| L | ⏳ | Benchmark baseline tracking with CI regression |

---

## Session Recovery Instructions

If session crashes or stops unexpectedly:

1. Read this file to find last completed task
2. Check git log for last commit
3. Continue from next pending (⏳) task in order
4. Update this file as tasks complete

**Legend**:

- ✅ = Completed
- 🔄 = In Progress
- ⏳ = Pending
- ❌ = Blocked

---

**Last Session**: Initial setup, grooming docs committed
**Next Task**: P0.10 - Create internal/infra/tls/ package
