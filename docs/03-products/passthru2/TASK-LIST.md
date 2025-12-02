# Passthru2: Implementation Task List

**Purpose**: Tasks for applying improvements, demo parity, and developer experience improvements after passthru1
**Created**: 2025-11-30
**Updated**: 2025-12-01 (aligned with Grooming Sessions 1-5 decisions)

---

## Phase 0: Developer Experience Foundation (Day 1-2)

**Priority**: HIGHEST - Based on Q1 (DX/Demo UX) and Q3 (Aggressive timeline)
**Status**: ✅ COMPLETE

### Infrastructure Tasks

- [x] **P0.1**: Extract telemetry to `deployments/telemetry/compose.yml` (Q6)
- [x] **P0.2**: Create `deployments/<product>/config/` standardized structure (Q7)
- [x] **P0.3**: Convert all secrets to Docker secrets (Q7, Q10)
- [x] **P0.4**: Remove empty directories (`identity/identity/`, `identity/postgres/`)
- [x] **P0.5**: Create compose profiles: `dev`, `demo`, `ci` per product (Q8)

### Demo Seeding Tasks

- [x] **P0.6**: Add demo seed data for KMS (admin, tenant-admin, user, service accounts)
- [x] **P0.7**: Add demo seed data for Identity (demo users and clients)
- [x] **P0.8**: Create `compose.demo.yml` for KMS with health checks (Q12)
- [x] **P0.9**: Create `compose.demo.yml` for Identity with health checks (Q12)

### TLS/HTTPS Fix (CRITICAL from Session 2 Q20 + Session 3 Q1-5)

- [x] **P0.10**: Create `internal/infra/tls/` package (Session 3 Q1)
- [x] **P0.11**: Implement CA chain with configurable length, default 3 (Session 3 Q2)
- [x] **P0.12**: Use FQDN style CNs, configurable (Session 3 Q3)
- [x] **P0.13**: Enable mTLS for all internal comms, configurable per pair (Session 3 Q4)
- [x] **P0.14**: Ensure Identity reuses new `internal/infra/tls/` package

### TLS Infrastructure (Session 4 Q1-5)

- [x] **P0.15**: Use std lib + golang.org/x/crypto only (no ACME in passthru2) (Session 4 Q1)
- [x] **P0.16**: Support PEM + PKCS#12 cert storage formats (Session 4 Q2)
- [x] **P0.17**: Custom CA only for demo (no system trust store) (Session 4 Q3)
- [x] **P0.18**: ALWAYS full TLS validation (hostname, expiry, chain) (Session 4 Q4)
- [x] **P0.19**: TLS 1.3 only (no TLS 1.2 fallback) (Session 4 Q5)

### HSM Placeholder (Session 5 Q2)

- [x] **P0.20**: Create `internal/infra/tls/hsm/` placeholder package for future PKCS#11/YubiKey

---

## Phase 1: KMS Demo Parity (Day 2-3)

**Priority**: HIGH - Based on Q2 (KMS & Identity equal parity) and Q13 (all KMS features)
**Status**: 🔄 IN PROGRESS

### Swagger UI (Highest priority per Q13 notes)

- [x] **P1.1**: Ensure KMS Swagger UI works with demo credentials
- [x] **P1.2**: Add interactive demo steps in Swagger UI
- [x] **P1.3**: Document Swagger UI demo flow in DEMO-KMS.md

### Auto-seed Demo Mode

- [x] **P1.4**: Implement `--demo` flag for KMS server
- [x] **P1.5**: Auto-seed key pools on demo startup
- [x] **P1.6**: Auto-seed encryption keys for demo
- [x] **P1.7**: Implement `--reset-demo` flag for data cleanup (Q15)

### CLI Demo Orchestration (Session 3 Q11-15 + Session 4 Q11-15 + Session 5)

- [x] **P1.8**: Create `cmd/demo/main.go` single binary with subcommands (Session 3 Q11)
- [x] **P1.9**: Implement `demo kms` subcommand
- [x] **P1.10**: Support all output formats: human/JSON/structured (Session 3 Q12)
- [x] **P1.11**: Continue on error, report summary (Session 3 Q13)
- [x] **P1.12**: Implement health check waiting, 30s default (Session 3 Q14)
- [x] **P1.13**: Verify all demo entities after startup (Session 3 Q15)
- [x] **P1.13a**: Implement structured error aggregation with step/phase info (Session 4 Q11)
- [x] **P1.13b**: Handle partial success: report + keep running + configurable (Session 4 Q12)
- [x] **P1.13c**: Implement configurable retry strategy (Session 4 Q13)
- [x] **P1.13d**: Show progress with step counter + spinner (Session 4 Q14)
- [x] **P1.13e**: Implement exit codes: 0=success, 1=partial, 2=failure (Session 5 Q1)
- [x] **P1.13f**: Add CLI color output with Windows ANSI + --no-color flag (Session 5 Q15)

### KMS Realm Configuration (from Session 2 Q1-5 + Session 3 Q6-10 + Session 5)

- [x] **P1.14**: Create `realms.yml` in same directory as config (Session 3 Q6)
- [x] **P1.15**: Implement configurable PBKDF2: SHA-256, 600K iterations, 32-byte salt (Session 5 Q12)
- [x] **P1.16**: Implement full user schema with JSON metadata + validation schema (Session 5 Q13)
- [x] **P1.17**: Implement configurable hierarchical roles (Session 3 Q9)
- [x] **P1.18**: Use UUIDv4 for tenant IDs (Session 3 Q10 - max randomness)

### UUIDv4 Tenant ID (Session 4 Q6-10)

- [x] **P1.19**: Implement UUIDv4 generation matching v7 pattern (Session 4 Q6)
- [x] **P1.20**: Strict UUID format validation for tenant IDs (Session 4 Q7)
- [x] **P1.21**: Full UUID display format with hyphens (Session 4 Q8)
- [x] **P1.22**: Regenerate demo tenant IDs on each startup (Session 4 Q9)
- [ ] **P1.23**: Tenant ID always via Authorization header, never path/query (Session 4 Q10)

### Tenant Isolation (Session 5 Q14)

- [ ] **P1.24**: Implement schema-per-tenant isolation (SQLite + PostgreSQL compatible)

### Coverage Improvements

- [ ] **P1.25**: Add KMS handler unit tests (target: 85%)
- [ ] **P1.26**: Add KMS businesslogic unit tests (target: 85%)

---

## Phase 2: Identity Demo Parity (Day 3-5)

**Priority**: HIGH - Based on Q2 (KMS & Identity equal parity)
**Status**: ⏳ PENDING

### Missing Endpoints

- [ ] **P2.1**: Implement `/authorize` endpoint
- [ ] **P2.2**: Implement full PKCE validation
- [ ] **P2.3**: Implement redirect handling

### Token Management

- [ ] **P2.4**: Fix refresh token rotation
- [ ] **P2.5**: Complete introspection tests
- [ ] **P2.6**: Complete revocation tests

### Demo Mode

- [ ] **P2.7**: Implement `--demo` flag for Identity server
- [ ] **P2.8**: Create `cmd/demo-identity/main.go` Go CLI (Q12)
- [ ] **P2.9**: Seed demo users (admin, user, service)
- [ ] **P2.10**: Seed demo clients (public, confidential)
- [ ] **P2.11**: Implement `--reset-demo` flag for data cleanup (Q15)
- [ ] **P2.12**: Profile-based persistence: dev=persist, ci=ephemeral (Q12)

### Identity Coverage Improvements

- [ ] **P2.13**: Add Identity idp/userauth tests (target: 80%)
- [ ] **P2.14**: Add Identity handler tests (target: 80%)

---

## Phase 3: Integration Demo (Day 5-7)

**Priority**: HIGH - Based on Q2 (Integration demo parity)
**Status**: ⏳ PENDING

### Token Validation in KMS (from Q6-10)

- [ ] **P3.1**: Implement token validation middleware (Q17 - mixed approach)
- [ ] **P3.2**: Implement local JWT validation with in-memory JWKS caching (Q6)
- [ ] **P3.3**: Implement configurable JWKS TTL (Q6)
- [ ] **P3.4**: Implement introspection for revocation checks
- [ ] **P3.5**: Make revocation check frequency configurable (Q7): every-request / sensitive-only / interval
- [ ] **P3.6**: Implement 401/403 error split + configurable detail level (Q8)

### Service-to-Service Auth (from Q9)

- [ ] **P3.7**: Implement client credentials auth option
- [ ] **P3.8**: Implement mTLS auth option
- [ ] **P3.9**: Implement API key auth option
- [ ] **P3.10**: Make auth method configurable

### Claims & Scopes (from Q10, Q18)

- [ ] **P3.11**: Extract all OIDC + custom claims from tokens (Q10)
- [ ] **P3.12**: Implement hybrid scope model (Q18)
- [ ] **P3.13**: Add coarse scopes: `kms:admin`, `kms:read`, `kms:write`
- [ ] **P3.14**: Add fine scopes: `kms:encrypt`, `kms:decrypt`, `kms:sign`
- [ ] **P3.15**: Add scope enforcement tests

### Integration Demo (Session 3 Q11-15)

- [ ] **P3.16**: Add `demo identity` subcommand to single binary
- [ ] **P3.17**: Add `demo all` subcommand for full integration
- [ ] **P3.18**: Create integration compose file
- [ ] **P3.19**: Implement demo script (get token → KMS operation)

### Token Validation Implementation (Session 3 Q16-20)

- [ ] **P3.20**: Implement JWKS cache (library TBD per Q16)
- [ ] **P3.21**: Support single + batch introspection + dedup (Session 3 Q17)
- [ ] **P3.22**: Implement hybrid error responses (OAuth + Problem Details) (Session 3 Q18)
- [ ] **P3.23**: Implement structured scope parser with validation (Session 3 Q19)
- [ ] **P3.24**: Implement typed claims struct with OIDC fields (Session 3 Q20)

---

## Phase 4: KMS Realm Authentication (Day 7-9)

**Priority**: MEDIUM - Based on Q11 (realm-based auth for KMS)
**Status**: ⏳ PENDING

### File Realm Implementation

- [ ] **P4.1**: Design realm configuration schema
- [ ] **P4.2**: Implement file realm loader
- [ ] **P4.3**: Implement basic auth for file realm
- [ ] **P4.4**: Add file realm tests

### DB Realm Implementation (PostgreSQL only - Q3)

- [ ] **P4.5**: Design `kms_realm_users` table schema (separate from Identity)
- [ ] **P4.6**: Implement native realm repository
- [ ] **P4.7**: Add DB realm tests

### Tenant Isolation (from Q5)

- [ ] **P4.8**: Implement database-level tenant isolation
- [ ] **P4.9**: Support separate schemas per tenant
- [ ] **P4.10**: Add tenant isolation tests

### Federation Support

- [ ] **P4.11**: Implement identity provider federation config
- [ ] **P4.12**: Implement multi-tenant authority mapping
- [ ] **P4.13**: Add federation tests

---

## Phase 5: CI & Quality Gates (Day 9-10)

**Priority**: MEDIUM - Based on Q21 (80% coverage) and Q24 (all CI changes)
**Status**: ⏳ PENDING

### Coverage Gates (Session 3 Q21-25 + Session 4 Q16-20)

- [ ] **P5.1**: Add coverage threshold enforcement (80% minimum per Q21)
- [ ] **P5.2**: Add per-package coverage reporting
- [ ] **P5.3**: Add coverage trend tracking
- [ ] **P5.3a**: Store benchmark baseline in untracked local directory (Session 4 Q16)
- [ ] **P5.3b**: Compare benchmarks against previous run (Session 4 Q16)
- [ ] **P5.3c**: CI-based regression detection for benchmarks (Session 4 Q16)

### Demo CI Jobs

- [ ] **P5.4**: Add demo profile CI job for KMS
- [ ] **P5.5**: Add demo profile CI job for Identity
- [ ] **P5.6**: Add integration demo CI job

### Database Matrix

- [ ] **P5.7**: Add SQLite test runs in CI
- [ ] **P5.8**: Add PostgreSQL test runs in CI

### Testing Improvements (Session 3 Q21-25 + Session 5)

- [ ] **P5.9**: Implement testutil package with test factories for common entities (Session 3 Q21)
- [ ] **P5.10**: Add per-package factories as needed (Session 3 Q21)
- [ ] **P5.11**: Ensure all tests use UUIDv7 unique prefixes (CRITICAL)
- [ ] **P5.12**: Add basic benchmarks for critical paths (Session 3 Q24)
- [ ] **P5.13**: Set 60s configurable integration timeout (Session 3 Q25)
- [ ] **P5.14**: Add test case descriptions in code (Session 3 Q25)
- [ ] **P5.14a**: Store benchmark baselines in JSON + Go bench format + SQLite (Session 5 Q19)

### Config & Deployment (Session 4 Q21-25 + Session 5)

- [ ] **P5.15**: YAML primary config format with converter option (Session 4 Q21)
- [ ] **P5.16**: Config validation at load + startup time (Session 4 Q22)
- [ ] **P5.17**: Standard config paths (/etc, ~/.config search order) (Session 4 Q23)
- [ ] **P5.18**: Defer hot reload to passthru3 (document option C if easy) (Session 4 Q24)
- [ ] **P5.19**: Add prod profile to compose (dev, demo, ci + prod) (Session 4 Q25)

### Error Handling (Session 5 Q20)

- [ ] **P5.20**: Implement RFC 7807 Problem Details error format

---

## Phase 6: Migration & Cleanup (Day 10-14)

**Priority**: LOW - Based on Q22 (hybrid migration strategy)
**Status**: ⏳ PENDING

### Infra Package Migration

- [ ] **P6.1**: Move `internal/common/apperr` to `internal/infra/apperr`
- [ ] **P6.2**: Move `internal/common/config` to `internal/infra/config`
- [ ] **P6.3**: Move `internal/common/magic` to `internal/infra/magic`
- [ ] **P6.4**: Move `internal/common/telemetry` to `internal/infra/telemetry`
- [ ] **P6.5**: Create `internal/infra/tls/` package (Session 3 Q1)
- [ ] **P6.6**: Support chain length 3 (root→intermediate→leaf) (Session 3 Q2)
- [ ] **P6.7**: Support both FQDN and descriptive CNs via config (Session 3 Q3)
- [ ] **P6.8**: Implement mTLS required + configurable fallback (Session 3 Q4)
- [ ] **P6.9**: Handle clock skew in development mode only (Session 3 Q5)

### Product Package Migration

- [ ] **P6.10**: Consolidate Identity duplicate code
- [ ] **P6.11**: Update all import paths
- [ ] **P6.12**: Verify build + test + lint after each move

---

## Acceptance Criteria (from Q25)

All must be true before closing passthru2:

- [ ] **A**: KMS and Identity demos both start with `docker compose` and include seeded data
- [ ] **B**: KMS and Identity both have interactive demo scripts and Swagger UI usable with demo creds
- [ ] **C**: Integration demo runs and validates token-based auth and scopes
- [ ] **D**: All product tests pass with coverage targets achieved (80%+)
- [ ] **E**: Telemetry extracted to shared compose and secrets standardized
- [ ] **F**: TLS/HTTPS pattern fixed - Identity uses KMS cert utilities (CRITICAL)
- [ ] **G**: Single demo binary with subcommands (kms, identity, all) implemented (Session 3)
- [ ] **H**: UUIDv4 for realm/tenant IDs standardized (Session 3)
- [ ] **I**: 60s configurable integration timeout implemented (Session 3)
- [ ] **J**: TLS 1.3 only, full validation, PEM/PKCS#12 storage (Session 4)
- [ ] **K**: Demo CLI with structured errors, progress, exit codes (Session 4)
- [ ] **L**: Benchmark baseline tracking with CI regression detection (Session 4)

---

**Status**: 🔄 IN PROGRESS
**Updated**: 2025-12-01 (aligned with Grooming Sessions 1-5 decisions)
