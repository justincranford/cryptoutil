# Tasks: PKI-CA-MIGRATE

**Option**: Migrate pki-ca to service-template (same pattern as cipher-im/jose-ja)
**Status**: Research Only
**Created**: 2026-02-23

---

## Phase A: sm-kms Migration Debt Cleanup

**Phase Objective**: Remove old-pattern wrappers and unify sm-kms with template

### Task A.1: Audit sm-kms middleware vs template middleware
- **Status**: ❌
- **Estimated**: 2h
- **Description**: Catalog all 15 sm-kms middleware files. For each, determine if equivalent functionality exists in template (session.go is the only template middleware). Identify: (a) exact duplicates, (b) extensions that should go to template, (c) sm-kms-specific logic that stays.
- **Acceptance Criteria**:
  - [ ] Audit table created showing overlap/gap for each middleware file
  - [ ] Decision documented for each: move to template / keep in sm-kms / delete as duplicate
- **Files**: `internal/apps/sm/kms/server/middleware/*.go`, `internal/apps/template/service/server/middleware/session.go`

### Task A.2: Remove sm-kms application_core and application_basic wrappers
- **Status**: ❌
- **Estimated**: 4h
- **Dependencies**: A.1
- **Description**: `application_core.go` and `application_basic.go` are old-pattern wrappers that pre-date the template builder. They should be replaced by direct template builder usage. Requires understanding what StartServerApplicationCore() does vs what builder.Build() provides (ServiceResources has DB, BarrierService, TelemetryService, JWKGenService, SessionManager).
- **Acceptance Criteria**:
  - [ ] `application_core.go` replaced by direct ServiceResources usage
  - [ ] `application_basic.go` removed or consolidated
  - [ ] sm-kms server.go TODO(Phase2-5) comment removed
  - [ ] All tests pass: `go test ./internal/apps/sm/kms/...`
- **Files**: `internal/apps/sm/kms/server/server.go`, `internal/apps/sm/kms/server/application/`

### Task A.3: Unify sm-kms ORM repository with template GORM pattern
- **Status**: ❌
- **Estimated**: 4h
- **Dependencies**: A.2
- **Description**: sm-kms has `server/repository/orm/` that uses GORM but server.go TODO says "Migrate SQLRepository to template's ORM pattern". Align the repository to follow cipher-im's pattern (domain-specific GORM models + repository using template DB from ServiceResources).
- **Acceptance Criteria**:
  - [ ] sm-kms repository uses template's GORM DB from ServiceResources
  - [ ] TODO comments removed from server.go
  - [ ] `go test -cover ./internal/apps/sm/kms/server/repository/...` ≥95%
- **Files**: `internal/apps/sm/kms/server/repository/`, `internal/apps/sm/kms/server/server.go`

### Task A.4: Add sm-kms integration TestMain using template helper
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: A.2, A.3
- **Description**: sm-kms has NO integration tests. Create a TestMain following the cipher-im integration pattern, using the template `StartServiceFromConfig()` helper (from fixes-v7 Task 6.0).
- **Acceptance Criteria**:
  - [ ] `internal/apps/sm/kms/server/testmain_integration_test.go` created
  - [ ] TestMain uses template generic startup helper
  - [ ] Basic CRUD integration tests added
  - [ ] `go test -tags=integration ./internal/apps/sm/kms/...` passes
- **Files**: `internal/apps/sm/kms/server/testmain_integration_test.go` (new)

### Task A.5: Add sm-kms E2E test suite
- **Status**: ❌
- **Estimated**: 3h
- **Dependencies**: A.4
- **Description**: Create `internal/apps/sm/kms/e2e/` matching cipher-im E2E structure. Uses template `e2e_infra.ComposeManager` pointing to `deployments/sm-kms/compose.yml`.
- **Acceptance Criteria**:
  - [ ] `internal/apps/sm/kms/e2e/` created with testmain_e2e_test.go, e2e_test.go
  - [ ] Tests start Docker Compose stack, wait for health, run basic operations
  - [ ] `go test -tags=e2e -timeout=30m ./internal/apps/sm/kms/e2e/...` passes
- **Files**: `internal/apps/sm/kms/e2e/` (new)

---

## Phase B: jose-ja Critical TODO Implementation

**Phase Objective**: Implement unimplemented JWK cryptographic operations

### Task B.1: Implement JWK generation in jwk_handler.go
- **Status**: ❌
- **Estimated**: 3h
- **Dependencies**: None (can start in parallel with Phase A)
- **Description**: `jwk_handler.go:358,368` have TODO stubs for actual JWK generation and barrier-version signing. These must be implemented for jose-ja to be functional.
- **Acceptance Criteria**:
  - [ ] JWK generation implemented using `JWKGenService` from ServiceResources
  - [ ] Barrier version integrated for key versioning
  - [ ] `go test -cover ./internal/apps/jose/ja/... ≥95%`
- **Files**: `internal/apps/jose/ja/server/apis/jwk_handler.go`

### Task B.2: Implement sign/verify/encrypt/decrypt in jwk_handler_material.go
- **Status**: ❌
- **Estimated**: 8h
- **Dependencies**: B.1
- **Description**: `jwk_handler_material.go:234,244,254,264` have TODO stubs for: signing (JWS), signature verification, encryption (JWE), decryption. These are the core operations of the JWK Authority service. Implement using `internal/shared/crypto/jose` primitives.
- **Acceptance Criteria**:
  - [ ] Sign operation: Creates JWS with active elastic key
  - [ ] Verify operation: Verifies JWS with historical keys (elastic key ring)
  - [ ] Encrypt operation: Creates JWE with active elastic key
  - [ ] Decrypt operation: Decrypts JWE with historical keys
  - [ ] All 4 operations tested with ≥95% coverage
- **Files**: `internal/apps/jose/ja/server/apis/jwk_handler_material.go`

### Task B.3: Add jose-ja integration and E2E test suite
- **Status**: ❌
- **Estimated**: 3h
- **Dependencies**: B.1, B.2
- **Description**: Create `internal/apps/jose/ja/e2e/` matching cipher-im E2E structure. Also add fuller integration tests covering all 4 crypto operations.
- **Acceptance Criteria**:
  - [ ] jose-ja E2E test suite created
  - [ ] Sign/verify/encrypt/decrypt tested end-to-end via HTTP
  - [ ] `go test -tags=e2e -timeout=30m ./internal/apps/jose/ja/e2e/...` passes
- **Files**: `internal/apps/jose/ja/e2e/` (new), `internal/apps/jose/ja/server/*_integration_test.go`

---

## Phase C: pki-ca Migration to Service-Template

**Phase Objective**: Migrate pki-ca to follow cipher-im/jose-ja builder pattern fully

### Task C.1: Add GORM certificate storage (replace MemoryStore)
- **Status**: ❌
- **Estimated**: 6h
- **Dependencies**: A.3 (template GORM pattern established)
- **Description**: Convert `storage/MemoryStore` to GORM-backed implementation. Keep the `Store` interface — just swap the implementation. Create domain migrations 2001+ for the certificate tables.
- **Acceptance Criteria**:
  - [ ] `StoredCertificate` has GORM tags (`type:text`, `gorm:"column:..."`)
  - [ ] Domain migration `2001_certificates.up.sql` created
  - [ ] `GORMStore` implements `Store` interface (SQLite + PostgreSQL compatible)
  - [ ] `MemoryStore` kept as fallback for unit tests
  - [ ] `go test -cover ./internal/apps/pki/ca/storage/...` ≥95%
- **Files**: `internal/apps/pki/ca/storage/gorm_store.go` (new), `internal/apps/pki/ca/server/repository/migrations/2001_certificates.up.sql` (new)

### Task C.2: Consolidate pki-ca magic package to shared/magic
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: pki-ca has local `magic/` package. Consolidate to `internal/shared/magic/magic_pki.go` per fixes-v7 Phase 3 Task 3.2.
- **Acceptance Criteria**:
  - [ ] `internal/apps/pki/ca/magic/` removed
  - [ ] Constants moved to `internal/shared/magic/magic_pki.go`
  - [ ] All imports updated
  - [ ] Build clean
- **Files**: `internal/apps/pki/ca/magic/` (delete), `internal/shared/magic/magic_pki.go` (new)

### Task C.3: Fix SetReady(true) startup sequence
- **Status**: ❌
- **Estimated**: 30min
- **Dependencies**: None
- **Description**: pki-ca `ca.go`'s `caServerStart()` calls `srv.SetReady(true)` before `srv.Start()`. This is incorrect — SetReady(true) should be called by TestMain in tests, and by the server's own startup sequence after all components are initialized. Align with cipher-im/jose-ja pattern.
- **Acceptance Criteria**:
  - [ ] SetReady(true) called at correct point in startup sequence
  - [ ] Tests verify readyz endpoint returns 200 only after full initialization
- **Files**: `internal/apps/pki/ca/ca.go`, `internal/apps/pki/ca/server/server.go`

### Task C.4: Add pki-ca testing helper package
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: C.1, C.3
- **Description**: Create `internal/apps/pki/ca/testing/testmain_helper.go` with `StartCAServer()` and `SetupTestServer()` — matching cipher-im pattern, using template `StartServiceFromConfig()` helper.
- **Acceptance Criteria**:
  - [ ] `StartCAServer()` and `SetupTestServer()` created
  - [ ] TestServerResources struct includes CA-specific resources (Issuer, Storage, CRLService, OCSPService)
  - [ ] Used by integration and E2E TestMains
- **Files**: `internal/apps/pki/ca/testing/testmain_helper.go` (new)

### Task C.5: Add pki-ca integration test suite
- **Status**: ❌
- **Estimated**: 3h
- **Dependencies**: C.4
- **Description**: Create `internal/apps/pki/ca/integration/` matching cipher-im integration pattern. Tests cover: certificate issuance (EST enroll), revocation, OCSP queries, CRL generation.
- **Acceptance Criteria**:
  - [ ] Integration tests cover main CA operations
  - [ ] Uses TestMain with `StartCAServer()`
  - [ ] `go test -tags=integration -cover ./internal/apps/pki/ca/...` ≥95%
- **Files**: `internal/apps/pki/ca/integration/` (new)

### Task C.6: Add pki-ca E2E test suite
- **Status**: ❌
- **Estimated**: 3h
- **Dependencies**: C.5
- **Description**: Create `internal/apps/pki/ca/e2e/` matching cipher-im E2E structure. Uses template `e2e_infra.ComposeManager` pointing to `deployments/pki-ca/compose.yml`.
- **Acceptance Criteria**:
  - [ ] pki-ca E2E test suite created
  - [ ] Docker Compose stack starts successfully
  - [ ] Certificate issuance and revocation tested end-to-end
  - [ ] `go test -tags=e2e -timeout=30m ./internal/apps/pki/ca/e2e/...` passes
- **Files**: `internal/apps/pki/ca/e2e/` (new)

### Task C.7: Update CI E2E workflow for pki-ca
- **Status**: ❌
- **Estimated**: 30min
- **Dependencies**: C.6
- **Description**: Add pki-ca E2E tests to `ci-e2e.yml`. Remove `SERVICE_TEMPLATE_TODO: pki-ca not yet migrated` comment. Update migration status comment.
- **Acceptance Criteria**:
  - [ ] `ci-e2e.yml` runs pki-ca E2E tests
  - [ ] Migration status updated: `pki-ca ✅`
  - [ ] TODO comments removed
- **Files**: `.github/workflows/ci-e2e.yml`

---

## Optional: Service-Template Consistency Validator

### Task OPT.1: Add validate-service-template cicd subcommand
- **Status**: ❌ (Optional)
- **Estimated**: 4h
- **Description**: Add `go run ./cmd/cicd validate-service-template <service-path>` that checks: RouteService usage, NewServerBuilder usage, migration numbering, TestMain template import.
- **Files**: `internal/cmd/cicd/lint_service_template/` (new)

---

## Summary Stats

| Phase | Tasks | Total Effort |
|-------|-------|-------------|
| A: sm-kms cleanup | 5 tasks | 15h |
| B: jose-ja TODOs | 3 tasks | 14h |
| C: pki-ca migration | 7 tasks | 16h |
| Optional | 1 task | 4h |
| **Total (core)** | **15 tasks** | **45h** |
