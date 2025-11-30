# Passthru2: Implementation Task List

**Purpose**: Tasks for applying improvements, demo parity, and developer experience improvements after passthru1
**Created**: 2025-11-30
**Updated**: 2025-11-30 (aligned with Grooming Session 1 decisions)

---

## Phase 0: Developer Experience Foundation (Day 1-2)

**Priority: HIGHEST** - Based on Q1 (DX/Demo UX) and Q3 (Aggressive timeline)

### Infrastructure Tasks

- [ ] **P0.1**: Extract telemetry to `deployments/telemetry/compose.yml` (Q6)
- [ ] **P0.2**: Create `deployments/<product>/config/` standardized structure (Q7)
- [ ] **P0.3**: Convert all secrets to Docker secrets (Q7, Q10)
- [ ] **P0.4**: Remove empty directories (`identity/identity/`, `identity/postgres/`)
- [ ] **P0.5**: Create compose profiles: `dev`, `demo`, `ci` per product (Q8)

### Demo Seeding Tasks

- [ ] **P0.6**: Add demo seed data for KMS (admin, tenant-admin, user, service accounts)
- [ ] **P0.7**: Add demo seed data for Identity (demo users and clients)
- [ ] **P0.8**: Create `compose.demo.yml` for KMS with health checks (Q12)
- [ ] **P0.9**: Create `compose.demo.yml` for Identity with health checks (Q12)

---

## Phase 1: KMS Demo Parity (Day 2-3)

**Priority: HIGH** - Based on Q2 (KMS & Identity equal parity) and Q13 (all KMS features)

### Swagger UI "Try it out" (Highest priority per Q13 notes)

- [ ] **P1.1**: Ensure KMS Swagger UI works with demo credentials
- [ ] **P1.2**: Add interactive demo steps in Swagger UI
- [ ] **P1.3**: Document Swagger UI demo flow in DEMO-KMS.md

### Auto-seed Demo Mode (Second priority)

- [ ] **P1.4**: Implement `--demo` flag for KMS server
- [ ] **P1.5**: Auto-seed key pools on demo startup
- [ ] **P1.6**: Auto-seed encryption keys for demo

### CLI Demo Orchestration (Third priority)

- [ ] **P1.7**: Create `cmd/demo-kms/main.go` Go CLI
- [ ] **P1.8**: Implement health check waiting
- [ ] **P1.9**: Implement demo flow execution (create pool → create key → encrypt → decrypt)

### Coverage Improvements

- [ ] **P1.10**: Add KMS handler unit tests (target: 85%)
- [ ] **P1.11**: Add KMS businesslogic unit tests (target: 85%)

---

## Phase 2: Identity Demo Parity (Day 3-5)

**Priority: HIGH** - Based on Q2 (KMS & Identity equal parity)

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

### Identity Coverage Improvements

- [ ] **P2.11**: Add Identity idp/userauth tests (target: 80%)
- [ ] **P2.12**: Add Identity handler tests (target: 80%)

---

## Phase 3: Integration Demo (Day 5-7)

**Priority: HIGH** - Based on Q2 (Integration demo parity)

### Token Validation in KMS

- [ ] **P3.1**: Implement token validation middleware (Q17 - mixed approach)
- [ ] **P3.2**: Implement local JWT validation with caching
- [ ] **P3.3**: Implement introspection for revocation checks
- [ ] **P3.4**: Make token validation configurable

### Scope Enforcement

- [ ] **P3.5**: Implement hybrid scope model (Q18)
- [ ] **P3.6**: Add coarse scopes: `kms:admin`, `kms:read`, `kms:write`
- [ ] **P3.7**: Add fine scopes: `kms:encrypt`, `kms:decrypt`, `kms:sign`
- [ ] **P3.8**: Add scope enforcement tests

### Integration Demo

- [ ] **P3.9**: Create `cmd/demo-all/main.go` Go CLI (Q12)
- [ ] **P3.10**: Create integration compose file
- [ ] **P3.11**: Implement demo script (get token → KMS operation)

---

## Phase 4: KMS Realm Authentication (Day 7-9)

**Priority: MEDIUM** - Based on Q11 (realm-based auth for KMS)

### File Realm Implementation

- [ ] **P4.1**: Design realm configuration schema
- [ ] **P4.2**: Implement file realm loader
- [ ] **P4.3**: Implement basic auth for file realm
- [ ] **P4.4**: Add file realm tests

### DB Realm Implementation (PostgreSQL only)

- [ ] **P4.5**: Design realm database schema
- [ ] **P4.6**: Implement native realm repository
- [ ] **P4.7**: Add DB realm tests

### Federation Support

- [ ] **P4.8**: Implement identity provider federation config
- [ ] **P4.9**: Implement multi-tenant authority mapping
- [ ] **P4.10**: Add federation tests

---

## Phase 5: CI & Quality Gates (Day 9-10)

**Priority: MEDIUM** - Based on Q21 (80% coverage) and Q24 (all CI changes)

### Coverage Gates

- [ ] **P5.1**: Add coverage threshold enforcement (80% minimum)
- [ ] **P5.2**: Add per-package coverage reporting
- [ ] **P5.3**: Add coverage trend tracking

### Demo CI Jobs

- [ ] **P5.4**: Add demo profile CI job for KMS
- [ ] **P5.5**: Add demo profile CI job for Identity
- [ ] **P5.6**: Add integration demo CI job

### Database Matrix

- [ ] **P5.7**: Add SQLite test runs in CI (Q20)
- [ ] **P5.8**: Add PostgreSQL test runs in CI (Q20)

---

## Phase 6: Migration & Cleanup (Day 10-14)

**Priority: LOW** - Based on Q22 (hybrid migration strategy)

### Infra Package Migration

- [ ] **P6.1**: Move `internal/common/apperr` to `internal/infra/apperr`
- [ ] **P6.2**: Move `internal/common/config` to `internal/infra/config`
- [ ] **P6.3**: Move `internal/common/magic` to `internal/infra/magic`
- [ ] **P6.4**: Move `internal/common/telemetry` to `internal/infra/telemetry`

### Product Package Migration

- [ ] **P6.5**: Consolidate Identity duplicate code
- [ ] **P6.6**: Update all import paths
- [ ] **P6.7**: Verify build + test + lint after each move

---

## Acceptance Criteria (from Q25)

All must be true before closing passthru2:

- [ ] **A**: KMS and Identity demos both start with `docker compose` and include seeded data
- [ ] **B**: KMS and Identity both have interactive demo scripts and Swagger UI usable with demo creds
- [ ] **C**: Integration demo runs and validates token-based auth and scopes
- [ ] **D**: All product tests pass with coverage targets achieved (80%+)
- [ ] **E**: Telemetry extracted to shared compose and secrets standardized

---

**Status**: IN PROGRESS
