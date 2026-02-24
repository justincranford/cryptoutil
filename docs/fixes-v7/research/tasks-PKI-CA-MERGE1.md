# Tasks: PKI-CA-MERGE1

**Option**: Archive pki-ca; build new pki-ca from jose-ja base + CA logic
**Status**: Research Only
**Created**: 2026-02-23

---

## Phase Pre: jose-ja Prerequisites (BLOCKING)

### Task Pre.1: Implement jose-ja JWK generation (critical TODO)
- **Status**: ❌
- **Estimated**: 3h
- **Description**: `jwk_handler.go:358,368` stubs must be implemented
- **Acceptance Criteria**: JWK generation complete, tests pass

### Task Pre.2: Implement jose-ja sign/verify/encrypt/decrypt (critical TODO)
- **Status**: ❌
- **Estimated**: 8h
- **Description**: `jwk_handler_material.go:234,244,254,264` stubs must be implemented
- **Acceptance Criteria**: All 4 crypto operations complete, tests pass

### Task Pre.3: Extract template generic startup helper
- **Status**: ❌
- **Estimated**: 2h
- **Description**: `internal/apps/template/service/testing/server_start_helpers.go` with `StartServiceFromConfig()`
- **Acceptance Criteria**: Template testing package has generic startup helper

---

## Phase 1: Archive and Skeleton

### Task 1.1: Archive current pki-ca
- **Status**: ❌
- **Estimated**: 30min
- **Dependencies**: None (can do any time)
- **Description**: Move `internal/apps/pki/ca/` to `internal/apps/archived/pki-ca/`. Update any imports in cmd/pki-ca/main.go to point to archived location. Ensure archived version still builds.
- **Acceptance Criteria**:
  - [ ] `internal/apps/archived/pki-ca/` exists and builds
  - [ ] `cmd/pki-ca/main.go` imports from archived location
  - [ ] `go build ./cmd/pki-ca/...` passes (against archived version)
- **Files**: git mv, `cmd/pki-ca/main.go`

### Task 1.2: Create new pki-ca skeleton from jose-ja
- **Status**: ❌
- **Estimated**: 3h
- **Dependencies**: 1.1, Pre.1, Pre.2
- **Description**: Copy jose-ja as template for new pki-ca. Rename all package references from `jose-ja`/`ja` to `pki-ca`/`ca`. Set up `ca.go`, `server/server.go`, `server/config/config.go`.
- **Acceptance Criteria**:
  - [ ] `internal/apps/pki/ca/ca.go` has `RouteService` entry
  - [ ] `internal/apps/pki/ca/server/server.go` has `NewFromConfig` using template builder
  - [ ] `internal/apps/pki/ca/server/config/config.go` has `CAServerSettings`
  - [ ] `go build ./internal/apps/pki/ca/...` passes (empty CA — no handlers yet)
- **Files**: New skeleton from jose-ja

---

## Phase 2: GORM Certificate Storage

### Task 2.1: Create GORM certificate repository
- **Status**: ❌
- **Estimated**: 4h
- **Dependencies**: 1.2
- **Description**: Create `server/repository/certificate_repository.go` that implements `storage.Store` interface using GORM. Supports SQLite and PostgreSQL. Create domain migration 2001_certificates.up.sql.
- **Acceptance Criteria**:
  - [ ] `CertificateRepository` implements `storage.Store` interface
  - [ ] SQLite and PostgreSQL compatible (use `type:text` for UUIDs, `serializer:json` for arrays)
  - [ ] Migration file created with correct schema
  - [ ] `go test -cover ./internal/apps/pki/ca/server/repository/...` ≥95%
- **Files**: `server/repository/certificate_repository.go` (new), `server/repository/migrations/2001_certificates.up.sql` (new)

### Task 2.2: Wire GORM store into server
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: 2.1
- **Description**: Update `server/server.go` to initialize `CertificateRepository` from ServiceResources.DB. Pass to all CA services that need storage.
- **Acceptance Criteria**:
  - [ ] Server creates `CertificateRepository` from template DB
  - [ ] CA services receive `Store` interface (not concrete type)
  - [ ] `go test ./internal/apps/pki/ca/server/...` passes

---

## Phase 3: Port CA Business Logic

### Task 3.1: Port compliance, crypto, profile, security packages
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: 1.2
- **Description**: These packages are standalone (no storage deps). Cherry-pick from archived pki-ca with updated import paths.
- **Acceptance Criteria**:
  - [ ] `compliance/`, `crypto/`, `profile/`, `security/` ported
  - [ ] All existing tests pass with new import paths
  - [ ] No `internal/apps/archived/pki-ca/` imports from new pki-ca

### Task 3.2: Port issuer, revocation, timestamp services with GORM storage adapter
- **Status**: ❌
- **Estimated**: 4h
- **Dependencies**: 2.2, 3.1
- **Description**: Port `service/issuer/`, `service/revocation/`, `service/timestamp/` using the new GORM-backed Store interface. Verify all existing tests pass.
- **Acceptance Criteria**:
  - [ ] All 3 service packages ported and using GORM store
  - [ ] Existing tests ported and passing
  - [ ] `go test -cover ./internal/apps/pki/ca/service/...` ≥95%

### Task 3.3: Port intermediate, observability, bootstrap packages
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: 3.2
- **Description**: Port remaining business packages. Resolve any storage dependencies.
- **Acceptance Criteria**:
  - [ ] All packages ported with imports updated
  - [ ] Tests passing

---

## Phase 4: Port API Handlers

### Task 4.1: Port est, certs, ocsp handlers
- **Status**: ❌
- **Estimated**: 6h
- **Dependencies**: 3.2, 3.3
- **Description**: Port 25-file api/handler/ from archived pki-ca. These are the most test-rich components; careful porting required. Update storage references to use GORM store.
- **Acceptance Criteria**:
  - [ ] `api/handler/handler.go`, `handler_est.go`, `handler_certs.go`, `handler_ocsp.go` ported
  - [ ] All handler tests pass (25 test files)
  - [ ] `go test -cover ./internal/apps/pki/ca/api/handler/...` ≥95%

---

## Phase 5: Testing Infrastructure

### Task 5.1: Add magic constants to shared/magic
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: 1.2
- **Description**: Create `internal/shared/magic/magic_pki.go` with CA constants. Do NOT create local magic/ package.

### Task 5.2: Add testing/ helper package
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: 4.1
- **Description**: Create `testing/testmain_helper.go` with `StartCAServer()` and `SetupTestServer()` using template generic helper.

### Task 5.3: Add integration test suite
- **Status**: ❌
- **Estimated**: 3h
- **Dependencies**: 5.2
- **Description**: Create `integration/` with TestMain pattern.

### Task 5.4: Add E2E test suite
- **Status**: ❌
- **Estimated**: 3h
- **Dependencies**: 5.3
- **Description**: Create `e2e/` using template e2e_infra.

### Task 5.5: Update CI E2E workflow
- **Status**: ❌
- **Estimated**: 30min
- **Dependencies**: 5.4
- **Description**: Add pki-ca E2E to ci-e2e.yml, remove SERVICE_TEMPLATE_TODO comments.

---

## Summary Stats

| Phase | Tasks | Est Effort |
|-------|-------|-----------|
| Pre: jose-ja prerequisites | 3 | 13h |
| 1: Archive + skeleton | 2 | 3.5h |
| 2: GORM storage | 2 | 5h |
| 3: Business logic | 3 | 8h |
| 4: API handlers | 1 | 6h |
| 5: Testing | 5 | 9.5h |
| **Total** | **16 tasks** | **45h** |
