# cryptoutil Implementation Tasks

**Generated**: 2025-12-24
**Source**: specs/002-cryptoutil/plan.md
**Status**: Phase 2-7 tasks (Phase 1 complete)

---

## Task Format

Each task includes:

- **ID**: Unique identifier (P#.#.#)
- **Title**: Brief description (3-7 words)
- **Phase**: Implementation phase (2-7)
- **Effort**: S (1-2 days), M (3-5 days), L (1-2 weeks)
- **Dependencies**: Blocking tasks (must complete first)
- **Completion Criteria**: Evidence-based validation
- **Files/Packages**: Where work will be done

---

## Phase 2: Core Services

### P2.1: Admin Server Migration (BLOCKING)

#### P2.1.1: JOSE Admin Server Implementation

- **Title**: Implement JOSE admin server
- **Effort**: M (3-5 days)
- **Dependencies**: None (Phase 1 complete)
- **Files**:
  - `internal/jose/server/admin.go` (create)
  - `internal/jose/server/application.go` (update for dual servers)
  - `cmd/jose-server/main.go` (update startup)
- **Completion Criteria**:
  - ✅ Admin server binds to 127.0.0.1:9093
  - ✅ Health endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/shutdown`
  - ✅ Dual server pattern (public + admin) with concurrent startup
  - ✅ Tests pass: `go test ./internal/jose/server/...`
  - ✅ Coverage ≥95%: `go test -cover ./internal/jose/server/...`
  - ✅ Docker Compose health check passes
  - ✅ Commit: `feat(jose): implement admin server for health checks and graceful shutdown`

#### P2.1.2: CA Admin Server Implementation

- **Title**: Implement CA admin server
- **Effort**: M (3-5 days)
- **Dependencies**: P2.1.1 (pattern established)
- **Files**:
  - `internal/ca/server/admin.go` (create)
  - `internal/ca/server/application.go` (update for dual servers)
  - `cmd/ca-server/main.go` (update startup)
- **Completion Criteria**:
  - ✅ Admin server binds to 127.0.0.1:9092
  - ✅ Health endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/shutdown`
  - ✅ Dual server pattern with concurrent startup
  - ✅ Tests pass, coverage ≥95%
  - ✅ Docker Compose health check passes
  - ✅ Commit: `feat(ca): implement admin server for health checks and graceful shutdown`

---

### P2.2: Unified CLI Enhancements

#### P2.2.1: Unified CLI Database Migrations

- **Title**: Add database migration commands to unified CLI
- **Effort**: S (1-2 days)
- **Dependencies**: P2.1.1, P2.1.2
- **Files**:
  - `cmd/cryptoutil/main.go` (add migration subcommands)
  - `internal/shared/database/migrations.go` (create migration utilities)
- **Completion Criteria**:
  - ✅ Commands: `cryptoutil migrate up`, `cryptoutil migrate down`, `cryptoutil migrate status`
  - ✅ Support PostgreSQL and SQLite
  - ✅ Dry-run mode for validation
  - ✅ Tests pass, coverage ≥98%
  - ✅ Commit: `feat(cli): add database migration commands to unified CLI`

#### P2.2.2: Unified CLI Health Check

- **Title**: Add health check command to unified CLI
- **Effort**: S (1-2 days)
- **Dependencies**: P2.1.1, P2.1.2
- **Files**:
  - `cmd/cryptoutil/main.go` (add health subcommand)
- **Completion Criteria**:
  - ✅ Command: `cryptoutil health check --service=<service>` or `cryptoutil health check --all`
  - ✅ Polls `/admin/v1/livez` and `/admin/v1/readyz`
  - ✅ Exit code 0 (healthy) or 1 (unhealthy)
  - ✅ Tests pass, coverage ≥98%
  - ✅ Commit: `feat(cli): add health check command for service monitoring`

---

### P2.3: E2E Service API Tests (Priority)

#### P2.3.1: E2E Tests - JOSE /service/* API

- **Title**: E2E tests for JOSE /service/* endpoints
- **Effort**: M (3-5 days)
- **Dependencies**: P2.1.1
- **Files**:
  - `test/e2e/jose_service_api_test.go` (create)
  - `.github/workflows/ci-e2e.yml` (update matrix)
- **Completion Criteria**:
  - ✅ Test coverage: JWK generation, JWKS retrieval, JWE encryption/decryption, JWS signing/verification, JWT operations
  - ✅ Uses /service/* paths (token-based authentication)
  - ✅ Docker Compose deployment with health checks
  - ✅ Tests pass: `go test ./test/e2e/... -run Jose`
  - ✅ Workflow passes: ci-e2e
  - ✅ Commit: `test(e2e): add JOSE /service/* API end-to-end tests`

#### P2.3.2: E2E Tests - CA /service/* API

- **Title**: E2E tests for CA /service/* endpoints
- **Effort**: M (3-5 days)
- **Dependencies**: P2.1.2
- **Files**:
  - `test/e2e/ca_service_api_test.go` (create)
  - `.github/workflows/ci-e2e.yml` (update matrix)
- **Completion Criteria**:
  - ✅ Test coverage: CSR submission, certificate issuance, OCSP status, CRL retrieval, EST enrollment
  - ✅ Uses /service/* paths
  - ✅ Docker Compose deployment
  - ✅ Tests pass, workflow passes
  - ✅ Commit: `test(e2e): add CA /service/* API end-to-end tests`

#### P2.3.3: E2E Tests - KMS /service/* API

- **Title**: E2E tests for KMS /service/* endpoints
- **Effort**: M (3-5 days)
- **Dependencies**: None (KMS complete)
- **Files**:
  - `test/e2e/kms_service_api_test.go` (create)
  - `.github/workflows/ci-e2e.yml` (update matrix)
- **Completion Criteria**:
  - ✅ Test coverage: Elastic Key create/rotate, Material Key operations, encrypt/decrypt, sign/verify
  - ✅ Uses /service/* paths
  - ✅ Docker Compose deployment
  - ✅ Tests pass, workflow passes
  - ✅ Commit: `test(e2e): add KMS /service/* API end-to-end tests`

#### P2.3.4: E2E Tests - Identity AuthZ /service/* API

- **Title**: E2E tests for Identity AuthZ /service/* endpoints
- **Effort**: M (3-5 days)
- **Dependencies**: None (AuthZ complete)
- **Files**:
  - `test/e2e/identity_authz_service_api_test.go` (create)
  - `.github/workflows/ci-e2e.yml` (update matrix)
- **Completion Criteria**:
  - ✅ Test coverage: OAuth 2.1 flows (authorization code, client credentials), token introspection, token revocation
  - ✅ Uses /service/* paths
  - ✅ Docker Compose deployment
  - ✅ Tests pass, workflow passes
  - ✅ Commit: `test(e2e): add Identity AuthZ /service/* API end-to-end tests`

#### P2.3.5: E2E Tests - Identity IdP /service/* API

- **Title**: E2E tests for Identity IdP /service/* endpoints
- **Effort**: M (3-5 days)
- **Dependencies**: None (IdP complete)
- **Files**:
  - `test/e2e/identity_idp_service_api_test.go` (create)
  - `.github/workflows/ci-e2e.yml` (update matrix)
- **Completion Criteria**:
  - ✅ Test coverage: OIDC authentication, login, consent, logout, MFA enrollment
  - ✅ Uses /service/* paths
  - ✅ Docker Compose deployment
  - ✅ Tests pass, workflow passes
  - ✅ Commit: `test(e2e): add Identity IdP /service/* API end-to-end tests`

---

### P2.4: Session State SQL Implementation

#### P2.4.1: JWS Session Token Format

- **Title**: Implement JWS session token format
- **Effort**: L (1-2 weeks)
- **Dependencies**: None
- **Files**:
  - `internal/shared/session/jws.go` (create)
  - `internal/shared/session/storage_sql.go` (create for SQL storage)
  - `internal/identity/authz/handlers/token.go` (update to support JWS)
- **Completion Criteria**:
  - ✅ JWS token generation with claims (sub, exp, iat, jti, scope)
  - ✅ JWS token verification with signature validation
  - ✅ SQL storage for revocation support (store JTI in database)
  - ✅ Configuration-driven selection (default for headless clients)
  - ✅ Tests pass, coverage ≥95%
  - ✅ Commit: `feat(session): implement JWS session token format with SQL storage`

#### P2.4.2: OPAQUE Session Token Format

- **Title**: Implement OPAQUE session token format
- **Effort**: M (3-5 days)
- **Dependencies**: P2.4.1
- **Files**:
  - `internal/shared/session/opaque.go` (create)
  - `internal/shared/session/storage_sql.go` (update for opaque tokens)
- **Completion Criteria**:
  - ✅ OPAQUE token generation (cryptographically random, ≥256 bits)
  - ✅ SQL storage for all session data (sessions table)
  - ✅ Configuration-driven selection (default for browser clients in some deployments)
  - ✅ Tests pass, coverage ≥95%
  - ✅ Commit: `feat(session): implement OPAQUE session token format with SQL storage`

#### P2.4.3: JWE Session Token Format

- **Title**: Implement JWE session token format
- **Effort**: M (3-5 days)
- **Dependencies**: P2.4.1
- **Files**:
  - `internal/shared/session/jwe.go` (create)
  - `internal/shared/session/storage_sql.go` (update for revocation tracking)
- **Completion Criteria**:
  - ✅ JWE token generation with encrypted claims
  - ✅ JWE token decryption and verification
  - ✅ SQL storage for revocation tracking (store JTI only)
  - ✅ Configuration-driven selection (default for browser clients in production deployments)
  - ✅ Tests pass, coverage ≥95%
  - ✅ Commit: `feat(session): implement JWE session token format with SQL storage`

---

## Phase 3: Advanced Features

### P3.1: Browser E2E Tests

#### P3.1.1: E2E Tests - JOSE /browser/* API

- **Title**: E2E tests for JOSE /browser/* endpoints
- **Effort**: M (3-5 days)
- **Dependencies**: P2.3.1, P2.4.3 (JWE browser sessions)
- **Files**:
  - `test/e2e/jose_browser_api_test.go` (create)
- **Completion Criteria**:
  - ✅ Uses /browser/* paths (session-based authentication)
  - ✅ CORS, CSRF, CSP middleware validation
  - ✅ Tests pass, workflow passes
  - ✅ Commit: `test(e2e): add JOSE /browser/* API end-to-end tests`

#### P3.1.2: E2E Tests - CA /browser/* API

- **Title**: E2E tests for CA /browser/* endpoints
- **Effort**: M (3-5 days)
- **Dependencies**: P2.3.2, P2.4.3
- **Files**:
  - `test/e2e/ca_browser_api_test.go` (create)
- **Completion Criteria**:
  - ✅ Uses /browser/* paths
  - ✅ Middleware validation
  - ✅ Tests pass, workflow passes
  - ✅ Commit: `test(e2e): add CA /browser/* API end-to-end tests`

#### P3.1.3: E2E Tests - Identity /browser/* API

- **Title**: E2E tests for Identity /browser/* endpoints
- **Effort**: M (3-5 days)
- **Dependencies**: P2.3.4, P2.3.5, P2.4.3
- **Files**:
  - `test/e2e/identity_browser_api_test.go` (create)
- **Completion Criteria**:
  - ✅ Uses /browser/* paths
  - ✅ Middleware validation
  - ✅ Tests pass, workflow passes
  - ✅ Commit: `test(e2e): add Identity /browser/* API end-to-end tests`

---

### P3.2: Identity RP and SPA

#### P3.2.1: Identity RP Implementation

- **Title**: Implement Identity Relying Party service
- **Effort**: L (1-2 weeks)
- **Dependencies**: P2.3.4, P2.4.3
- **Files**:
  - `internal/identity/rp/server/public_server.go` (create)
  - `internal/identity/rp/server/admin.go` (create)
  - `cmd/identity-rp/main.go` (create)
- **Completion Criteria**:
  - ✅ Backend-for-Frontend pattern implementation
  - ✅ Dual servers (public 8300 + admin 9091)
  - ✅ OAuth 2.1 client implementation
  - ✅ Tests pass, coverage ≥95%
  - ✅ Docker Compose integration
  - ✅ Commit: `feat(identity): implement Relying Party service with BFF pattern`

#### P3.2.2: Identity SPA Implementation

- **Title**: Implement Identity Single Page Application hosting
- **Effort**: M (3-5 days)
- **Dependencies**: P3.2.1
- **Files**:
  - `internal/identity/spa/server/public_server.go` (create)
  - `internal/identity/spa/server/admin.go` (create)
  - `cmd/identity-spa/main.go` (create)
  - `web/spa/` (static files directory)
- **Completion Criteria**:
  - ✅ Static file hosting for SPA client
  - ✅ Dual servers (public 8400 + admin 9091)
  - ✅ CSP, CORS configuration
  - ✅ Tests pass, coverage ≥95%
  - ✅ Docker Compose integration
  - ✅ Commit: `feat(identity): implement SPA hosting service with static file serving`

---

## Phase 4: Scale & Multi-Tenancy

### P4.1: Database Sharding

#### P4.1.1: Tenant ID Partitioning Strategy

- **Title**: Implement tenant ID database sharding
- **Effort**: L (1-2 weeks)
- **Dependencies**: P3.2.1, P3.2.2
- **Files**:
  - `internal/shared/database/sharding.go` (create)
  - Schema migration scripts
- **Completion Criteria**:
  - ✅ Partition by tenant ID
  - ✅ Configuration-driven shard allocation
  - ✅ Migration tooling
  - ✅ Tests pass, coverage ≥95%
  - ✅ Commit: `feat(database): implement tenant ID sharding strategy`

---

### P4.2: Schema-Level Multi-Tenancy

#### P4.2.1: Schema Isolation Implementation

- **Title**: Implement schema-level multi-tenancy isolation
- **Effort**: L (1-2 weeks)
- **Dependencies**: P4.1.1
- **Files**:
  - `internal/shared/database/multitenancy.go` (create)
  - Schema creation/management utilities
- **Completion Criteria**:
  - ✅ Schema-level isolation (tenant_a.users, tenant_b.users)
  - ✅ Tenant provisioning automation
  - ✅ Connection pool per-tenant management
  - ✅ Tests pass, coverage ≥95%
  - ✅ Commit: `feat(database): implement schema-level multi-tenancy isolation`

---

## Phase 5: Production Readiness

### P5.1: Hash Service Refactoring

#### P5.1.1: Hash Registry Version Management

- **Title**: Refactor hash registries for version management
- **Effort**: L (1-2 weeks)
- **Dependencies**: P4.2.1
- **Files**:
  - `internal/shared/crypto/hashes/` (refactor all registries)
- **Completion Criteria**:
  - ✅ Version-based registry selection
  - ✅ Pepper rotation lazy migration support
  - ✅ Tests pass, coverage ≥98%
  - ✅ Mutation testing ≥98%
  - ✅ Commit: `refactor(hashes): implement version-based registry management`

---

### P5.2: Security Hardening

#### P5.2.1: mTLS Revocation Checking

- **Title**: Implement CRLDP and OCSP revocation checking
- **Effort**: L (1-2 weeks)
- **Dependencies**: P5.1.1
- **Files**:
  - `internal/shared/crypto/revocation/crldp.go` (create)
  - `internal/shared/crypto/revocation/ocsp.go` (create)
- **Completion Criteria**:
  - ✅ BOTH CRLDP and OCSP implemented
  - ✅ CRLDP immediate (not batched)
  - ✅ OCSP stapling nice-to-have
  - ✅ Tests pass, coverage ≥98%
  - ✅ Mutation testing ≥98%
  - ✅ Commit: `feat(security): implement CRLDP and OCSP revocation checking`

---

## Phase 6: Service Template Extraction

### P6.1: Extract Template from KMS

#### P6.1.1: Template Package Structure

- **Title**: Extract service template package structure
- **Effort**: M (3-5 days)
- **Dependencies**: P5.2.1
- **Files**:
  - `pkg/servicetemplate/` (create)
  - Documentation and examples
- **Completion Criteria**:
  - ✅ Dual-server pattern template
  - ✅ Database integration template
  - ✅ OTLP telemetry template
  - ✅ Configuration management template
  - ✅ Tests pass, coverage ≥98%
  - ✅ Commit: `feat(template): extract reusable service template from KMS`

---

## Phase 7: Learn-PS Validation

### P7.1: Learn-PS Implementation

#### P7.1.1: Learn-PS Service with Template

- **Title**: Implement Learn-PS Pet Store using service template
- **Effort**: M (3-5 days)
- **Dependencies**: P6.1.1
- **Files**:
  - `cmd/learn-ps/main.go` (create)
  - `internal/learn/petstore/` (create)
- **Completion Criteria**:
  - ✅ Uses service template package
  - ✅ Dual servers (public 8888 + admin 9090)
  - ✅ PostgreSQL/SQLite support
  - ✅ OTLP telemetry integration
  - ✅ Tests pass, coverage ≥95%
  - ✅ Docker Compose integration
  - ✅ Commit: `feat(learn-ps): implement Pet Store using service template`

---

## Summary Statistics

**Total Tasks**: 32 tasks
**Phase 2**: 13 tasks (admin servers, CLI, E2E service, sessions)
**Phase 3**: 5 tasks (browser E2E, RP/SPA)
**Phase 4**: 2 tasks (sharding, multi-tenancy)
**Phase 5**: 2 tasks (hash service, security)
**Phase 6**: 1 task (template extraction)
**Phase 7**: 1 task (Learn-PS validation)

**Effort Breakdown**:

- Small (S): 4 tasks (4-8 days total)
- Medium (M): 20 tasks (60-100 days total)
- Large (L): 8 tasks (64-128 days total)

**Critical Path**: P2.1.1 → P2.1.2 → P2.3.x → P2.4.x → P3.x → P4.x → P5.x → P6.x → P7.1.1

**Evidence-Based Completion**: ALL tasks require tests passing, coverage targets met, and conventional commits

---

**Note**: Tasks for Phases 8-15 (production migrations) will be generated after Phase 7 Learn-PS validation proves service template works correctly.
