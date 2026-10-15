# Tasks â€” Framework v25: 8-to-7 PS-ID Consolidation

**Status**: Complete for the original consolidation scope; review follow-up status is recorded below.
**Last Updated**: 2026-10-15
**Created**: 2026-05-25

---

**Archived task status note**: Phase 1–4 task labels and acceptance checkboxes were not maintained during execution and are archived planning metadata, not current implementation status. Original consolidation verification is recorded in the historical Phase 5 artifacts; those artifacts do not substantiate the October 2026 review findings. Current follow-up resolutions, evidence, and gaps are tracked in the final section.

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- âœ… **Correctness**: ALL code must be functionally correct with comprehensive tests
- âœ… **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- âœ… **Thoroughness**: Evidence-based validation at every step
- âœ… **Reliability**: Quality gates enforced (â‰¥95%/98% coverage/mutation)
- âœ… **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- âœ… **Accuracy**: Changes must address root cause, not just symptoms
- âŒ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- âŒ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

---

## Task Status Legend â€” MANDATORY

| Symbol | Meaning | When to Use |
|--------|---------|-------------|
| âŒ | Not started | Task not yet begun |
| ðŸ”„ | In progress | Currently being worked on |
| âœ… | Complete | Task finished with evidence |
| â³ | Blocked | Requires external dependency (MUST have resolution plan) |

---

## Phase 1: Port pki-ca Packages to sm-kms

**Phase Objective**: Copy all pki-ca internal packages to `internal/apps/sm-kms/server/pki/`
with updated package names and import paths. Add DB migration 2009 for ca_items. Move
certificate profile configs to sm-kms. Verify ported packages compile and tests pass.

**Archived status**: The ported packages, migration, and profiles are present. Historical build and test evidence is in `test-output/phase5/`; it does not validate the October 2026 follow-up changes.

**Prerequisite**: framework-v24 must be fully executed before starting this phase.

---

### Task 1.1: Add sm-kms DB Migration 2009 (ca_items)

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: framework-v24 complete (migrations 2001â€“2008 in place)
- **Description**: Port pki-ca's `5001_ca_items.up.sql` as sm-kms migration 2009.
  Create matching down migration.
- **Acceptance Criteria**:
  - [x] `2009_pki_ca_items.up.sql` created in `internal/apps/sm-kms/server/repository/migrations/`
  - [x] `2009_pki_ca_items.down.sql` created
  - [x] Existing `migrations/*.sql` embed pattern includes the new migration
  - [x] Migration runs cleanly on fresh SQLite in-memory DB
- **Files**:
  - `internal/apps/sm-kms/server/repository/migrations/2009_pki_ca_items.up.sql` (NEW)
  - `internal/apps/sm-kms/server/repository/migrations/2009_pki_ca_items.down.sql` (NEW)
  - `internal/apps/sm-kms/server/repository/migrations.go` (UPDATE embed directive)

---

### Task 1.2: Port pki-ca compliance Package

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Task 1.1
- **Description**: Copy `internal/apps/pki-ca/compliance/` to
  `internal/apps/sm-kms/server/pki/compliance/`. Update package declarations and all
  import paths referencing pki-ca sub-packages to use new sm-kms paths.
- **Acceptance Criteria**:
  - [ ] Package compiles: `go build ./internal/apps/sm-kms/server/pki/compliance/...`
  - [ ] All tests pass: `go test ./internal/apps/sm-kms/server/pki/compliance/...`
  - [ ] No references to old `pki-ca/compliance` import path in new files
- **Files**:
  - `internal/apps/sm-kms/server/pki/compliance/checker.go` (NEW â€” ported)
  - `internal/apps/sm-kms/server/pki/compliance/checker_test.go` (NEW â€” ported)

---

### Task 1.3: Port pki-ca crypto Package

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Task 1.2
- **Description**: Copy `internal/apps/pki-ca/crypto/` to
  `internal/apps/sm-kms/server/pki/crypto/`. Update all cross-package imports.
- **Acceptance Criteria**:
  - [ ] Package compiles
  - [ ] All tests pass
- **Files**:
  - `internal/apps/sm-kms/server/pki/crypto/provider.go` (NEW)
  - `internal/apps/sm-kms/server/pki/crypto/provider_test.go` (NEW)

---

### Task 1.4: Port pki-ca storage Package

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Task 1.3
- **Description**: Copy `internal/apps/pki-ca/storage/` to
  `internal/apps/sm-kms/server/pki/storage/`. This is the in-memory certificate
  storage using `sync.RWMutex` + Go maps.
- **Acceptance Criteria**:
  - [ ] Package compiles
  - [ ] All storage interface tests pass
- **Files**:
  - `internal/apps/sm-kms/server/pki/storage/storage.go` (NEW)
  - `internal/apps/sm-kms/server/pki/storage/storage_test.go` (NEW)

---

### Task 1.5: Port pki-ca domain Package

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 1.1
- **Description**: Copy `internal/apps/pki-ca/domain-v2/` to
  `internal/apps/sm-kms/server/pki/domain/`. Contains `CAItem` GORM model.
  Update GORM model imports and package declaration.
- **Acceptance Criteria**:
  - [ ] GORM model builds with sm-kms imports
  - [ ] `TableName()` returns `"ca_items"`
- **Files**:
  - `internal/apps/sm-kms/server/pki/domain/model.go` (NEW)
  - `internal/apps/sm-kms/server/pki/domain/model_test.go` (NEW, if any)

---

### Task 1.6: Port pki-ca profile Packages

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Task 1.5
- **Description**: Copy `internal/apps/pki-ca/profile/certificate/` and
  `internal/apps/pki-ca/profile/subject/` to their equivalents under
  `internal/apps/sm-kms/server/pki/profile/`.
- **Acceptance Criteria**:
  - [ ] Both sub-packages compile
  - [ ] All tests pass
- **Files**:
  - `internal/apps/sm-kms/server/pki/profile/certificate/profile.go` (NEW)
  - `internal/apps/sm-kms/server/pki/profile/certificate/profile_test.go` (NEW)
  - `internal/apps/sm-kms/server/pki/profile/subject/profile.go` (NEW)
  - `internal/apps/sm-kms/server/pki/profile/subject/profile_test.go` (NEW)

---

### Task 1.7: Port pki-ca security Package

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Task 1.6
- **Description**: Copy `internal/apps/pki-ca/security/` (CSR validation + threat model)
  to `internal/apps/sm-kms/server/pki/security/`. Update imports.
- **Acceptance Criteria**:
  - [ ] Package compiles
  - [ ] CSR validation tests pass
- **Files**:
  - `internal/apps/sm-kms/server/pki/security/validator.go` (NEW)
  - `internal/apps/sm-kms/server/pki/security/validator_test.go` (NEW)
  - `internal/apps/sm-kms/server/pki/security/threat_model.go` (NEW, if exists)

---

### Task 1.8: Port pki-ca service/issuer Package

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: Tasks 1.3, 1.4, 1.6, 1.7
- **Description**: Copy `internal/apps/pki-ca/service/issuer/` to
  `internal/apps/sm-kms/server/pki/service/issuer/`. This is the core cert issuance
  service. Has the most cross-package dependencies.
- **Acceptance Criteria**:
  - [ ] Package compiles
  - [ ] Issuer unit tests pass
- **Files**:
  - `internal/apps/sm-kms/server/pki/service/issuer/issuer.go` (NEW)
  - `internal/apps/sm-kms/server/pki/service/issuer/issuer_test.go` (NEW)

---

### Task 1.9: Port pki-ca service/ra Package

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Dependencies**: Task 1.8
- **Description**: Copy `internal/apps/pki-ca/service/ra/` (Registration Authority)
  to `internal/apps/sm-kms/server/pki/service/ra/`. Update imports.
- **Acceptance Criteria**:
  - [ ] Package compiles
  - [ ] RA unit tests pass
- **Files**:
  - `internal/apps/sm-kms/server/pki/service/ra/ra.go` (NEW)
  - `internal/apps/sm-kms/server/pki/service/ra/ra_test.go` (NEW)

---

### Task 1.10: Port pki-ca service/revocation Package

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Dependencies**: Task 1.8
- **Description**: Copy `internal/apps/pki-ca/service/revocation/` (CRL + OCSP)
  to `internal/apps/sm-kms/server/pki/service/revocation/`. Update imports.
- **Acceptance Criteria**:
  - [ ] Package compiles
  - [ ] CRL and OCSP unit tests pass
- **Files**:
  - `internal/apps/sm-kms/server/pki/service/revocation/revocation.go` (NEW)
  - `internal/apps/sm-kms/server/pki/service/revocation/revocation_test.go` (NEW)

---

### Task 1.11: Port pki-ca service/timestamp Package

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Task 1.8
- **Description**: Copy `internal/apps/pki-ca/service/timestamp/` (RFC 3161 TSA)
  to `internal/apps/sm-kms/server/pki/service/timestamp/`. Update imports.
- **Acceptance Criteria**:
  - [ ] Package compiles
  - [ ] TSA unit tests pass
- **Files**:
  - `internal/apps/sm-kms/server/pki/service/timestamp/timestamp.go` (NEW)
  - `internal/apps/sm-kms/server/pki/service/timestamp/timestamp_test.go` (NEW)

---

### Task 1.12: Port pki-ca observability, bootstrap, intermediate Packages

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Tasks 1.8, 1.9, 1.10
- **Description**: Copy remaining pki-ca packages: `observability/` (OTel metrics),
  `bootstrap/` (offline root CA creation), `intermediate/` (intermediate CA management).
- **Acceptance Criteria**:
  - [ ] All three packages compile
  - [ ] Tests pass in all three
- **Files**:
  - `internal/apps/sm-kms/server/pki/observability/metrics.go` (NEW)
  - `internal/apps/sm-kms/server/pki/observability/metrics_test.go` (NEW)
  - `internal/apps/sm-kms/server/pki/bootstrap/bootstrap.go` (NEW)
  - `internal/apps/sm-kms/server/pki/bootstrap/bootstrap_test.go` (NEW)
  - `internal/apps/sm-kms/server/pki/intermediate/intermediate.go` (NEW)
  - `internal/apps/sm-kms/server/pki/intermediate/intermediate_test.go` (NEW)

---

### Task 1.13: Port pki-ca HTTP Handlers

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: Tasks 1.9, 1.10, 1.11
- **Description**: Copy `internal/apps/pki-ca/api/handler/` (handler.go, handler_certs.go,
handler_est.go, handler_ocsp.go) to `internal/apps/sm-kms/server/pki/api/handler/`.
  Update imports. The handlers will use the ported service packages.
- **Acceptance Criteria**:
  - [ ] All handler files compile
  - [ ] Handler unit tests pass (using Fiber app.Test())
- **Files**:
  - `internal/apps/sm-kms/server/pki/api/handler/` (ported handler implementation and tests)

---

### Task 1.14: Move Certificate Profile Configs

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Move `configs/pki-ca/profiles/` to `configs/sm-kms/profiles/`.
  These are YAML certificate profile configuration files read at CA startup.
- **Acceptance Criteria**:
  - [ ] `configs/sm-kms/profiles/` exists with all profile files
  - [ ] `go run ./cmd/cicd-lint lint-deployments` still passes (no orphaned config references)
- **Files**:
  - `configs/sm-kms/profiles/` (NEW directory â€” copied from configs/pki-ca/profiles/)

---

### Task 1.15: Phase 1 Quality Gate Verification

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Tasks 1.1â€“1.14
- **Description**: Run full build, lint, and test suite on sm-kms to verify all ported
  packages are clean before proceeding to API wiring.
- **Acceptance Criteria**:
  - [ ] `go build ./internal/apps/sm-kms/...` clean
  - [ ] `golangci-lint run ./internal/apps/sm-kms/...` clean
  - [ ] `go test ./internal/apps/sm-kms/...` 100% passing
  - [ ] No TODO/FIXME markers in ported code without task tracking
- **Evidence**: `test-output/phase5/build.log`, `test-output/phase5/tests.log`, `test-output/phase5/lint.log`

---

## Phase 2: Extend sm-kms API with PKI Endpoints

**Phase Objective**: Add all 18 PKI endpoints to the sm-kms OpenAPI spec under `/pki/` path
prefix. Update oapi-codegen configs with PKI initialisms. Regenerate API code. Wire PKI
handlers into sm-kms server at both `/service/` and `/browser/` path prefixes.

**Status**: Complete - OpenAPI validation, generated-code builds, handler tests, and the dual-prefix route-registration test passed.

---

### Task 2.1: Add PKI Schemas to sm-kms OpenAPI Components

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Dependencies**: Task 1.15
- **Description**: Port PKI schemas from `api/pki-ca/openapi_spec_enrollment.yaml` into
  `api/sm-kms/openapi_spec_components.yaml`. This includes: CA, Certificate,
  CertificateChain, EnrollmentRequest, EnrollmentStatus, Profile, OCSPRequest,
  OCSPResponse, TimestampRequest, TimestampResponse, and related schema objects.
- **Acceptance Criteria**:
  - [ ] All PKI schema components added to sm-kms components spec
  - [ ] Schema references are valid (`$ref` paths resolve)
  - [ ] No duplicate schema names with existing sm-kms schemas
- **Files**:
  - `api/sm-kms/openapi_spec_components.yaml` (UPDATE â€” +PKI schemas)

---

### Task 2.2: Add PKI Paths to sm-kms OpenAPI Paths Spec

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: Task 2.1
- **Description**: Add all 18 PKI endpoint paths to `api/sm-kms/openapi_spec_paths.yaml`
  under the `/pki/` prefix. Follow the dual `/service/api/v1/pki/` and `/browser/api/v1/pki/`
  pattern. Handle binary content types for OCSP, TSA, and EST endpoints.

  **18 endpoints to add** (both /service/ and /browser/ variants):
  - `GET  /pki/cas`
  - `GET  /pki/cas/{caID}`
  - `GET  /pki/cas/{caID}/crl`
  - `POST /pki/enrollments`
  - `GET  /pki/enrollments/{requestID}`
  - `GET  /pki/certificates`
  - `GET  /pki/certificates/{serialNumber}`
  - `GET  /pki/certificates/{serialNumber}/chain`
  - `POST /pki/certificates/{serialNumber}/revoke`
  - `GET  /pki/profiles`
  - `GET  /pki/profiles/{profileID}`
  - `POST /pki/ocsp` (application/ocsp-request â†’ application/ocsp-response)
  - `GET  /pki/est/cacerts`
  - `POST /pki/est/simpleenroll`
  - `POST /pki/est/simplereenroll`
  - `POST /pki/est/serverkeygen`
  - `GET  /pki/est/csrattrs`
  - `POST /pki/tsa/timestamp`

- **Acceptance Criteria**:
  - [ ] All 18 PKI paths added (at both /service/ and /browser/ mount points)
  - [ ] Binary content types correctly declared (application/ocsp-request, etc.)
  - [ ] Pagination parameters on list endpoints (GET /pki/cas, /pki/certificates, /pki/profiles)
  - [ ] OpenAPI 3.0.3 spec validates without errors
- **Files**:
  - `api/sm-kms/openapi_spec_paths.yaml` (UPDATE â€” +18 PKI path definitions)

---

### Task 2.3: Update oapi-codegen Configs with PKI Initialisms

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 2.2
- **Description**: Add PKI-specific initialisms to all three sm-kms oapi-codegen config files.
  Per `02-04.openapi.instructions.md`, pki-ca domain additions are: `CSR`, `CA`, `CRL`,
  `OCSP`, `URI`, `SAN`, `DN`, `CN`, `OU`.
- **Acceptance Criteria**:
  - [ ] `openapi-gen_config_server.yaml` has all 9 PKI initialisms in additional-initialisms
  - [ ] `openapi-gen_config_model.yaml` has all 9 PKI initialisms
  - [ ] `openapi-gen_config_client.yaml` has all 9 PKI initialisms
- **Files**:
  - `api/sm-kms/openapi-gen_config_server.yaml` (UPDATE)
  - `api/sm-kms/openapi-gen_config_model.yaml` (UPDATE)
  - `api/sm-kms/openapi-gen_config_client.yaml` (UPDATE)

---

### Task 2.4: Regenerate sm-kms API Code

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 2.3
- **Description**: Run `oapi-codegen` for sm-kms to regenerate server.gen.go, models.gen.go,
  and client.gen.go with the new PKI endpoint types and handlers.
- **Acceptance Criteria**:
  - [ ] `server.gen.go` regenerated successfully
  - [ ] `models.gen.go` regenerated with all PKI types
  - [ ] `client.gen.go` regenerated with PKI client methods
  - [ ] Generated code compiles: `go build ./api/sm-kms/...`
- **Files**:
  - `api/sm-kms/server/server.gen.go` (REGENERATED)
  - `api/sm-kms/model/models.gen.go` (REGENERATED)
  - `api/sm-kms/client/client.gen.go` (REGENERATED)

---

### Task 2.5: Implement PKI StrictServer Interface in sm-kms

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: Tasks 1.13, 2.4
- **Description**: Create the sm-kms PKI StrictServer implementation that delegates to the
  ported pki handler package. Wire the StrictServer to the generated interface. Connect the
  in-memory storage and CA services at construction time.
- **Acceptance Criteria**:
  - [ ] PKI StrictServer implements all generated handler interfaces
  - [ ] All PKI handler methods delegate to ported pki/handler package
  - [ ] Unit tests cover all 18 handler methods
- **Files**:
  - `internal/apps/sm-kms/server/pki_server.go` (NEW â€” StrictServer impl)
  - `internal/apps/sm-kms/server/pki_server_test.go` (NEW)

---

### Task 2.6: Register PKI Routes in sm-kms Server

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Dependencies**: Task 2.5
- **Description**: Update `internal/apps/sm-kms/server/server.go` to register PKI routes
  at `/service/api/v1/pki/...` and `/browser/api/v1/pki/...`. Apply the same middleware
  chain as other sm-kms routes. Initialize PKI storage and CA services during startup.
- **Acceptance Criteria**:
  - [ ] PKI routes registered at both /service/ and /browser/ path prefixes
  - [x] `GET /service/api/v1/pki/profiles` returns 200 in an in-memory `app.Test()` request
  - [x] `GET /browser/api/v1/pki/profiles` returns 200 in an in-memory `app.Test()` request
  - [x] All 18 route definitions are present under both mounts; PKI handler behavior is exercised by the PKI API tests
- **Files**:
  - `internal/apps/sm-kms/server/server.go` (UPDATE â€” +PKI route registration)

---

### Task 2.7: Add sm-kms PKI Route and Handler Tests

- **Status**: Complete - route registration and handler behavior are covered by in-memory Fiber tests.
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: Task 2.6
- **Description**: Verify all PKI route registrations and exercise mounted profiles endpoints
  with `fiber.app.Test()`. Handler-level tests cover PKI request and response behavior.
- **Acceptance Criteria**:
  - [x] The route-registration test enumerates all 18 PKI endpoints under both public API prefixes
  - [x] `GET /pki/profiles` returns 200 under both prefixes using `app.Test()`
  - [x] OCSP, EST, and TSA handler paths have handler-level request/response tests
  - [x] Route and handler tests use in-memory Fiber requests; the route test is a unit test and correctly uses the `_test.go` suffix rather than an integration suffix
- **Files**:
  - `internal/apps/sm-kms/server/pki_routes_test.go`
  - `internal/apps/sm-kms/server/pki/api/handler/*_test.go`
- **Evidence**: `test-output/phase5/tests.log`

---

### Task 2.8: Phase 2 Quality Gate Verification

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Tasks 2.1â€“2.7
- **Description**: Full quality gate check for Phase 2 deliverables.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `golangci-lint run ./...` clean
  - [ ] `go test ./internal/apps/sm-kms/...` 100% passing
  - [ ] sm-kms coverage â‰¥95% for all new pki packages
- **Evidence**: `test-output/phase5/build.log`, `test-output/phase5/tests.log`, `test-output/phase5/pki-coverage.out`

---

## Phase 3: Delete pki-ca Service Artifacts

**Phase Objective**: Remove the entire pki-ca standalone service. All its functionality is
now served by sm-kms. Delete API specs, all ~100 internal packages, cmd entry point, configs,
and deployment artifacts. Remove pki-ca TLS cert CN constants from magic_pki_tls.go.

**Status**: Complete - standalone pki-ca directories are absent, and normal/tagged builds and deployment validation passed.

---

### Task 3.1: Delete api/pki-ca Directory

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 2.8
- **Description**: Delete entire `api/pki-ca/` directory tree (OpenAPI spec file,
  oapi-codegen config files, and generated server/model/client packages).
- **Acceptance Criteria**:
  - [ ] `api/pki-ca/` directory no longer exists
  - [ ] `go build ./...` still clean after deletion
- **Files**: `api/pki-ca/` (DELETE entire directory)

---

### Task 3.2: Delete internal/apps/pki-ca Directory

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 3.1
- **Description**: Delete entire `internal/apps/pki-ca/` directory tree (~100 files).
  All functionality has been ported to `internal/apps/sm-kms/server/pki/`.
- **Acceptance Criteria**:
  - [ ] `internal/apps/pki-ca/` directory no longer exists
  - [ ] `go build ./...` clean after deletion (no orphaned imports)
- **Files**: `internal/apps/pki-ca/` (DELETE entire directory â€” ~100 files)

---

### Task 3.3: Delete cmd/pki-ca Entry Point

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 3.2
- **Description**: Delete `cmd/pki-ca/main.go` â€” the standalone service binary entry point.
- **Acceptance Criteria**:
  - [ ] `cmd/pki-ca/` directory no longer exists
  - [ ] `go build ./cmd/...` still clean
- **Files**: `cmd/pki-ca/main.go` (DELETE)

---

### Task 3.4: Delete configs/pki-ca Directory

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 1.14 (profiles already moved to sm-kms)
- **Description**: Delete `configs/pki-ca/` directory (pki-ca-domain.yml,
  pki-ca-framework.yml, and any remaining files; profiles already moved in Task 1.14).
- **Acceptance Criteria**:
  - [ ] `configs/pki-ca/` directory no longer exists
  - [ ] `go run ./cmd/cicd-lint lint-deployments` still passes
- **Files**: `configs/pki-ca/` (DELETE entire directory)

---

### Task 3.5: Delete deployments/pki-ca Directory

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 3.4
- **Description**: Delete `deployments/pki-ca/` directory (compose.yml, Dockerfile,
  configs/, secrets/, ca-crls/, certs/).
- **Acceptance Criteria**:
  - [ ] `deployments/pki-ca/` directory no longer exists
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes
- **Files**: `deployments/pki-ca/` (DELETE entire directory)

---

### Task 3.6: Remove AppPKICA* Constants from magic_pki_tls.go

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 3.5
- **Description**: Remove the pki-ca server cert CN constants from
  `internal/shared/magic/magic_pki_tls.go`. These are: `AppPKICASQLite1ServerCertCN`,
  `AppPKICASQLite2ServerCertCN`, `AppPKICAPostgres1ServerCertCN`,
  `AppPKICAPostgres2ServerCertCN`. Keep all other service CNs (identity, etc.).
- **Acceptance Criteria**:
  - [ ] All 4 AppPKICA* constants removed from magic_pki_tls.go
  - [ ] `go build ./...` clean (no compilation errors from removed constants)
  - [ ] Other service CNs remain unchanged
- **Files**:
  - `internal/shared/magic/magic_pki_tls.go` (UPDATE â€” remove AppPKICA* lines)

---

### Task 3.7: Phase 3 Quality Gate Verification

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Tasks 3.1â€“3.6
- **Description**: Verify clean build and lint after all pki-ca artifacts are deleted.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `golangci-lint run ./...` clean
  - [ ] `golangci-lint run --build-tags e2e,integration ./...` clean
  - [ ] `go test ./...` 100% passing
- **Evidence**: `test-output/phase5/build.log`, `test-output/phase5/cicd-lint.log`

---

## Phase 4: Delete pki Product and Update Registry

**Phase Objective**: Remove the pki product: CLI router, deployment artifacts, magic_pki.go,
and registry entries. Update magic_tier.go, api/cryptosuite-registry/registry.yaml,
lint_fitness entity registry, and lint_ports to reflect the new 3-product / 7-PS-ID topology.

**Status**: Complete - the registry and fitness checks validate the 3-product / 7-PS-ID topology.

---

### Task 4.1: Delete internal/apps/pki Product Router

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 3.7
- **Description**: Delete `internal/apps/pki/pki.go` and `internal/apps/pki/pki_test.go`.
  This was the pki product CLI router that delegated to pki-ca and pki-init.
- **Acceptance Criteria**:
  - [ ] `internal/apps/pki/` directory no longer exists
  - [ ] `go build ./...` clean
- **Files**: `internal/apps/pki/` (DELETE entire directory)

---

### Task 4.2: Delete cmd/pki Entry Point

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 4.1
- **Description**: Delete `cmd/pki/main.go` â€” the pki product binary entry point.
- **Acceptance Criteria**:
  - [ ] `cmd/pki/` directory no longer exists
  - [ ] `go build ./cmd/...` clean
- **Files**: `cmd/pki/main.go` (DELETE)

---

### Task 4.3: Delete deployments/pki Directory

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 4.2
- **Description**: Delete `deployments/pki/` directory (compose.yml, secrets/).
- **Acceptance Criteria**:
  - [ ] `deployments/pki/` directory no longer exists
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes
- **Files**: `deployments/pki/` (DELETE entire directory)

---

### Task 4.4: Delete magic_pki.go

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Tasks 4.1, 4.2
- **Description**: Delete `internal/shared/magic/magic_pki.go`. This file contains:
  `OTLPServicePKICA`, `PKICAServiceID`, `PKIProductName`, `PKICAServiceName`,
  `PKICAServicePort`, `PKICADisplayName`, E2E test port constants (8300â€“8303).
  All are unused after pki-ca deletion.
- **Acceptance Criteria**:
  - [ ] `magic_pki.go` deleted
  - [ ] `go build ./...` clean (no usages remaining)
- **Files**: `internal/shared/magic/magic_pki.go` (DELETE)

---

### Task 4.5: Update magic_tier.go

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 4.4
- **Description**: Update `internal/shared/magic/magic_tier.go` to remove:
  - `PKIProductName` entry from `ProductToPSIDs` map
  - `OTLPServicePKICA` entry from `AllPSIDs` slice
  Resulting in 3 products and 7 PS-IDs.
- **Acceptance Criteria**:
  - [ ] `ProductToPSIDs` has 3 entries (sm, identity, skeleton)
  - [ ] `AllPSIDs` has 7 entries
  - [ ] `go build ./...` clean
- **Files**:
  - `internal/shared/magic/magic_tier.go` (UPDATE â€” remove PKIProductName and OTLPServicePKICA)

---

### Task 4.6: Update api/cryptosuite-registry/registry.yaml

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 4.5
- **Description**: Remove `pki` product entry and `pki-ca` PS-ID entry from
  `api/cryptosuite-registry/registry.yaml`. The port range 8300 and pg_host_port 54323
  and migration range 5001â€“5999 are freed but not reallocated in this plan.
- **Acceptance Criteria**:
  - [ ] `registry.yaml` has 3 products (sm, identity, skeleton)
  - [ ] `registry.yaml` has 7 PS-IDs
  - [ ] No remaining references to pki-ca in registry.yaml
- **Files**:
  - `api/cryptosuite-registry/registry.yaml` (UPDATE â€” remove pki product + pki-ca PS-ID)

---

### Task 4.7: Update lint_fitness Entity Registry

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 4.6
- **Description**: Update `internal/apps-tools/cicd_lint/lint_fitness/registry/registry.go`
  to remove the pki-ca entry from `allProductServices`. Update all fitness functions that
  reference pki product names or pki-ca paths (e.g., apps-ps-id-template fitness linter
  iterates AllProductServices).
- **Acceptance Criteria**:
  - [ ] Registry has 7 PS-ID entries (pki-ca removed)
  - [ ] `go run ./cmd/cicd-lint lint-fitness` runs without panics
  - [x] No live pki-ca product or PS-ID references remain in lint_fitness; retired-value references are limited to regression tests
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/registry/registry.go` (UPDATE)

---

### Task 4.8: Update cryptoutil Suite Router

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 4.7
- **Description**: Update `internal/apps/cryptoutil/cryptoutil.go` (and related cmd/cryptoutil)
  to remove pki product delegation. The suite router iterates products â€” pki must be removed.
- **Acceptance Criteria**:
  - [ ] Cryptoutil suite router has no pki product entry
  - [ ] `go build ./cmd/cryptoutil/...` clean
- **Files**:
  - `internal/apps/cryptoutil/cryptoutil.go` (UPDATE â€” remove pki delegation)
  - `cmd/cryptoutil/main.go` (UPDATE if needed)

---

### Task 4.9: Update lint_ports (Remove pki-ca Port References)

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 4.8
- **Description**: Update `internal/apps-tools/cicd_lint/lint_ports/` to remove
  `PKICAServicePort = 8300` and any pki-ca port assignment references. Update tests.
- **Acceptance Criteria**:
  - [x] No active pki-ca port mapping remains in lint_ports; regression tests retain the retired value to assert it is rejected
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_ports/...` passes
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_ports/` (UPDATE â€” remove pki-ca port refs)

---

### Task 4.10: Phase 4 Quality Gate Verification

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Tasks 4.1â€“4.9
- **Description**: Full quality gate verification after pki product deletion.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `golangci-lint run ./...` clean
  - [ ] `go test ./...` 100% passing
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes â€” reports 3 products / 7 PS-IDs
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes
  - [x] No live pki-ca product/service imports or registrations remain; the retired-ID sentinel and regression tests are intentionally retained
- **Evidence**: `test-output/phase5/build.log`, `test-output/phase5/tests.log`, `test-output/phase5/cicd-lint.log`

---

## Phase 5: Quality Gates

**Phase Objective**: Comprehensive end-to-end quality verification. Full build, lint, test,
coverage analysis, race detector, and cicd-lint suite for the original consolidation. This
historical evidence does not validate the October 2026 review follow-up findings.

**Archived status**: The originally recorded quality gates passed; see `test-output/phase5/`.

---

### Task 5.1: Full Build Verification

- **Status**: Complete — normal and tagged builds passed.
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 4.10
- **Description**: Build all targets including E2E and integration build tags.
- **Acceptance Criteria**:
  - [x] `go build ./...` zero errors
  - [x] `go build -tags e2e,integration ./...` zero errors
- **Evidence**: `test-output/phase5/build.log`

---

### Task 5.2: Full Linting Verification

- **Status**: Complete — normal and tagged Go lint passed.
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 5.1
- **Description**: Run golangci-lint with and without E2E/integration build tags.
- **Acceptance Criteria**:
  - [x] `golangci-lint run ./...` zero warnings
  - [x] `golangci-lint run --build-tags e2e,integration ./...` zero warnings
- **Evidence**: `test-output/phase5/lint.log`

---

### Task 5.3: Full Test Suite Verification

- **Status**: Complete — the full shuffled suite passed; sm-kms race tests passed in the Go 1.26.1 Linux container because this Windows host has no GCC compiler.
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 5.2
- **Description**: Run full test suite with shuffle and verify 100% passing.
- **Acceptance Criteria**:
  - [x] `go test ./... -shuffle=on` 100% passing, zero skips
  - [x] `go test -race -count=2 ./internal/apps/sm-kms/...` passes (run in the Go 1.26.1 Linux container; see `test-output/phase5/race-docker.log`)
- **Evidence**: `test-output/phase5/tests.log`, `test-output/phase5/race-docker.log`

---

### Task 5.4: Coverage Analysis for New pki Packages

- **Status**: Complete - all PKI packages with executable statements meet the 95% threshold; aggregate sm-kms coverage increased from the retained 78.3% baseline to 79.9%. Focused coverage review closed meaningful gaps in storage filtering, gauge initialization, and handler JSON error propagation.
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 5.3
- **Description**: Run coverage and verify all new sm-kms/server/pki/ packages meet â‰¥95%.
- **Acceptance Criteria**:
  - [x] All `internal/apps/sm-kms/server/pki/...` packages with statements meet >=95% coverage
  - [x] No regression in retained aggregate or package-specific coverage baselines (79.9% vs. 78.3% aggregate; businesslogic 95.3% vs. 95.2%; handler remains 100%)
  - [x] Issuer/date filter boundaries, decrementing an uninitialized gauge, and error-response JSON failure propagation have focused tests; their target functions reach 100%.
- **Evidence**: `test-output/phase5/coverage.log`, `test-output/phase5/pki-coverage.out`,
  `test-output/phase5/sm-kms-coverage.log`, `test-output/phase5/sm-kms-coverage.out`,
  `test-output/coverage-analysis/pki-consolidation-final.out`,
  `test-output/coverage-analysis/pki-consolidation-final.html`,
  `test-output/coverage-analysis/pki-storage-target.out`,
  `test-output/coverage-analysis/pki-observability-target.out`,
  `test-output/coverage-analysis/pki-handler-target.out`,
  `test-output/coverage-analysis/sm-kms-businesslogic.cover`,
  `test-output/coverage-analysis/sm-kms-handler.cover`

---

### Task 5.5: cicd-lint Full Suite

- **Status**: Complete - all applicable cicd-lint checks, including docs propagation validation, passed.
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 5.4
- **Description**: Run all relevant cicd-lint commands.
- **Acceptance Criteria**:
  - [x] `go run ./cmd/cicd-lint lint-text` passes
  - [x] `go run ./cmd/cicd-lint lint-go` passes
  - [x] `go run ./cmd/cicd-lint lint-go-test` passes
  - [x] `go run ./cmd/cicd-lint lint-go-mod` passes
  - [x] `go run ./cmd/cicd-lint lint-fitness` passes (3 products / 7 PS-IDs)
  - [x] `go run ./cmd/cicd-lint lint-deployments` passes
  - [x] `go run ./cmd/cicd-lint lint-openapi` passes
  - [x] `go run ./cmd/cicd-lint lint-docs` passes
- **Evidence**: `test-output/phase5/cicd-lint.log`, `test-output/phase5/docs-final.log`

---

## Phase 6: Knowledge Propagation

**Phase Objective**: Apply all lessons learned to permanent project artifacts. Update
ENG-HANDBOOK.md, agents, skills, instructions, and active docs. Remove stale pki-ca references
from current project documentation while retaining migration history where it is relevant.
Verify propagation integrity.

**Status**: Complete - lessons and documentation are reconciled, and docs propagation validation passes.

---

### Task 6.1: Review All lessons.md Phase Entries

- **Status**: Complete - all six phase entries were reviewed and propagation/follow-up actions recorded.
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 5.5
- **Description**: Review all 5 prior phase lessons entries. Identify patterns and insights
  to propagate to permanent artifacts.
- **Acceptance Criteria**:
  - [x] All prior phase lessons reviewed
  - [x] Propagation targets identified and applied to the handbook, README, and repository trackers

---

### Task 6.2: Update ENG-HANDBOOK.md

- **Status**: Complete - the product/service topology and PKI placement are updated.
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Task 6.1
- **Description**: Update ENG-HANDBOOK.md:
  - Remove pki-ca from service catalog
  - Update sm-kms API list to include PKI endpoints
  - Update product/PS-ID count (3 products / 7 PS-IDs)
  - Add any new patterns discovered during pki-ca migration
- **Acceptance Criteria**:
  - [x] Service catalog reflects 7 PS-IDs
  - [x] sm-kms description mentions PKI under `/pki/` path prefix
  - [x] No stale pki-ca references remain in the handbook
- **Files**: `docs/ENG-HANDBOOK.md` (UPDATE)

---

### Task 6.3: Update Agents, Skills, Instructions

- **Status**: Complete - no stale pki-ca references were found in agent, skill, or instruction artifacts.
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 6.2
- **Description**: Scan agents, skills, and instruction files for pki-ca references and
  update them to reflect the new topology. Run lint-docs to verify propagation.
- **Acceptance Criteria**:
  - [x] No stale pki-ca references in agents or instructions
  - [x] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**: Various `.github/agents/`, `.github/instructions/`, `.github/skills/` files as needed

---

### Task 6.4: Final Commit and Cleanup

- **Status**: Complete - final changes are committed and the worktree is clean.
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 6.3
- **Description**: Final clean commit. Verify `git status --porcelain` returns empty.
  Update lessons.md Executive Summary and Actions sections.
- **Acceptance Criteria**:
  - [x] `git status --porcelain` returns empty after the final commit
  - [x] lessons.md Executive Summary and Actions sections filled
  - [x] All commits follow conventional commit format

---

## Cross-Cutting Tasks

### Testing

- [x] All new PKI packages with statements meet >=95% coverage
- [x] All 18 routes are registered under both API prefixes; the profiles endpoint returns 200 at both mounts
- [x] Race detector clean: `go test -race -count=2 ./internal/apps/sm-kms/...` in the Go 1.26.1 Linux container
- [x] No skipped tests without documented justification

### Code Quality

- [x] Linting passes: `golangci-lint run ./...` and with `--build-tags e2e,integration`
- [x] No new TODO/FIXME/HACK markers in the ported PKI code
- [x] Formatting/lint checks pass in normal and tagged configurations
- [x] No live pki-ca or PKIProductName imports/registrations remain; retired sentinels and regression tests are intentional

### Registry Integrity

- [x] `api/cryptosuite-registry/registry.yaml` shows 3 products / 7 PS-IDs
- [x] lint_fitness entity registry consistent with registry.yaml
- [x] magic_tier.go consistent with registry.yaml

### Documentation

- [x] ENG-HANDBOOK.md service catalog updated
- [x] README.md now describes PKI under sm-kms rather than as a separate product
- [x] No stale pki port 8300 references in active README/ENG-HANDBOOK content; the framework-v25 plan retains the retired allocation as historical context

---

## Follow-up Resolution and Registry Decisions

**Status note**: The checked PKI persistence items below are prior implementation claims, not
independent validation. The October 2026 follow-up section at the end of this document is the
current source for their disposition and evidence.

- [x] PKI certificate and authority data use sm-kms GORM persistence on the shared SQLite/PostgreSQL database, via migrations 2010 and 2011.
- [x] The CA private key is barrier-encrypted before persistence.
- [x] CRL and OCSP responses use tenant-scoped persisted revocation status, including after process restart.
- [x] Retired PKI port 8300, PostgreSQL host port 54323, and migration range 5001-5999 remain unassigned; any future allocation requires an explicit canonical registry change.

---

## Evidence Archive

- Phase 1-4 evidence directories were not retained; final-state verification is recorded in the Phase 5 logs and registry/deployment validators.
- `test-output/phase5/` - Full quality-gate logs, coverage profiles, coverage reports, and `docs-final.log`.
- `test-output/coverage-analysis/` - Retained historical aggregate and package-specific coverage baselines.

## Review Follow-Up: Findings, Disposition, and Evidence

The separate recommendation plan and per-recommendation statuses are in
[`review-follow-up-plan.md`](./review-follow-up-plan.md). This table remains the source for
implementation dispositions, retained evidence, and outstanding verification.

The following table supersedes earlier follow-up completion claims. Build, focused test,
integration, lint, and workflow evidence is under `test-output/review-followup/`. Focused
validation is not a claim that the full repository suite or Docker/PostgreSQL deployment ran.

| # | Disposition | Evidence | Remaining verification |
|---|---|---|---|
| 1 | Coverage gate fails closed on inventory/parser/result errors and enforces infrastructure thresholds; committed as `414688e8b`. | Workflow lint and normal/tagged builds. | CI runner execution not repeated locally. |
| 2 | Product/suite databases are isolated by PS-ID; KMS capabilities remain together in `sm-kms`; committed as `dd5520c`. | Deployment lint. | No live product/suite stack run. |
| 3 | Durable SQLite variants use file-backed databases and persistent volumes; memory variants are disposable; committed as `e3a8a67`. | Deployment lint. | No live container restart test. |
| 4 | CRL/OCSP publication and lookup are issuer-wide and do not require a tenant session. | Focused PKI tests and persistence integration log. | PostgreSQL multi-instance runtime not exercised. |
| 5 | CRL numbers allocate atomically in storage, scoped by issuer key ID. | Persistence integration log verifies the sequence continues after reopen. | PostgreSQL multi-instance runtime not exercised. |
| 6 | Ordinary certificate updates cannot restore revoked records or erase revocation metadata. | Focused PKI tests. | Full repository suite not run. |
| 7 | Shutdown-request state is separate from shutdown execution; HTTP shutdown performs the drain and listener close. | Focused listener tests. | No separate live-listener end-to-end test. |
| 8 | `pki-init` stages and validates material before publication, rolls back partial publication, and refuses rotation of existing material. | Focused TLS tests; `docs/tls-structure.md` describes initialization-only behavior. | No live consumer reload/rotation test; rotation is unsupported. |
| 9 | Compose readiness requires completed initialization jobs and authenticated readiness instead of matching `"healthy"` or ignoring host-check failures; committed as `3d2cc7b04`. | Workflow lint. | No live Compose stack run. |
| 10 | Production migrations, barrier encryption, file-backed SQLite reopen, revocation, CRL/OCSP continuity, and historical issuer publication are exercised together. | `pki-persistence-integration.log`. | PostgreSQL cross-instance issuance/revocation remains unverified. |
| 11 | Existing PostgreSQL volumes have a data-preserving, idempotent database upgrade path; committed as `0132019`. | Deployment lint. | Upgrade procedure not executed against an existing PostgreSQL volume. |
| 12 | Enrollment status uses tenant-scoped persistence as authority; the process-local tracker no longer leaks status across tenants. | Focused PKI tests. | Full repository suite not run. |
| 13 | Tenant-scoped request IDs have a unique database constraint, with duplicate errors mapped consistently. | Persistence integration and focused storage tests. | Legacy duplicate-row remediation is not automated. |
| 14 | Listener shutdown completion waits are bounded and report timeout errors. | Focused listener tests include shutdown/startup paths. | Full race suite not run. |
| 15 | Authority configuration, active validity checks, historical issuer restoration, and a persisted stable authority ID on new certificate records are covered. | Focused PKI tests and `pki-persistence-integration.log`. | Renewal/rollover operations remain initialization-only; no live PostgreSQL test. |
| 16 | Historical claims are separated from current evidence; this table records both verified work and unrun runtime checks. | `lint-docs` and `lint-text` logs. | Full-suite and Docker/PostgreSQL evidence remains outstanding. |

## Review Follow-Up Recommendation Disposition and Evidence

Recommendation numbers below refer to `review-follow-up-plan.md`, not to the historical task rows above. Recommendations 1 and 28 are intentionally excluded from this implementation scope.

| Recommendation | Disposition | Evidence | Remaining work |
|---|---|---|---|
| 2 | Partially verified | SQLite migration tests and `postgres-migrations-2012-2015.log` cover migration behavior; the PostgreSQL migration-runner test is added but not executed. | Run `TestPKIMigrations_PostgreSQLUpgradeAndRollback` in a separate Docker-enabled environment; no further Docker lifecycle operations were run after the recorded volume incident. |
| 3 | Complete | `durable-sqlite-compose-health.log` records healthy Compose startup; `durable-sqlite-authenticated-readiness.log` records authenticated admin readiness HTTP 200; `final-repository-validators.log` passes deployment validation. | None. |
| 4 | Complete | The baseline-maintenance procedure is in `docs/ENG-HANDBOOK.md` §10.2.3 and `.github/instructions/03-02.testing.instructions.md`; `coverage-baseline-docs.log`, `coverage-baseline-workflow.log`, and `coverage-baseline-text.log` pass. Both inventories are uploaded by the workflow. | None. |
| 5 | Partially verified | `test-output/coverage-analysis/review-followup/coverage-gate-runner-summary.md` records 152/320 packages passing and 168/320 below target (71.5% aggregate); failures span 84 `internal/apps-tools`, 54 `internal/apps`, 18 `internal/apps-framework`, and 12 `internal/shared` packages. In-scope PKI package coverage is handler 90.8%, bootstrap 93.2%, server 52.4%, revocation 80.4%, and storage 69.4%. Six synthetic gate cases pass; the complete inventory parses without errors. | Add targeted tests across all below-target packages; for any proposed exception, complete and document the per-package coverage-ceiling analysis and mitigation. Confirm a passing CI runner execution; do not lower thresholds to clear the gate. |
| 6 | Partially verified | `final-focused-tests.log` passes listener tests; `race-sm-kms-listener.log` records that race validation cannot run because Windows lacks `gcc`. | Run required race validation on a supported CGO-enabled host. |
| 7 | Complete | `final-focused-tests.log` passes TLS publication and failure-path tests; `docs/tls-structure.md` documents initialization-only behavior and Compose activation gating. | None; rotation/reload remains unsupported. |
| 8 | Complete | `ci-load.yml` uses authenticated mTLS readiness; `postgres-authenticated-readiness.log` and `durable-sqlite-authenticated-readiness.log` record live readiness responses; `final-repository-validators.log` passes workflow lint. | Full load workload was not run. |
| 9 | Complete | PKI test conventions were audited; `final-focused-tests.log` passes the relevant shuffled PKI packages and `postgres-continuity-restart.log` passes the PostgreSQL continuity test. | None. |
| 10 | Complete | `test-output/review-followup/pki-persistence-integration.log`; see historical task row 10. | None. |
| 11 | Complete | `postgres-continuity-restart.log` passes the PostgreSQL restart test, reinitializing authority components and verifying restored certificate, encrypted-key envelope, and revocation state. | None. |
| 12 | Complete | `docs/framework-v25/plan.md` remains the completed historical plan; `test-output/review-followup/docs-validation.log`. | None. |
| 13 | Complete | `test-output/reviewfollowup/final-builds.log` records passing normal and `e2e,integration` builds; `final-golangci-normal-verify.log` and `final-golangci-tagged-verify.log` report no issues; `final-repository-validators.log` passes all 9 applicable validators. `final-go-test-shuffle.log` records the serial full shuffled suite passing. An initial full-suite run concurrent with validators timed out in RP/RS health checks; isolated tests and shuffled package reruns passed before the serial suite rerun. | None. |
| 14 | Complete | `test-output/reviewfollowup/commit-trailer-audit.log` verifies the required co-author trailer on every task commit from the recorded baseline through the implementation, disposition, and post-review fix commits. | None. |
| 15 | Complete | `test-output/review-followup/postgres-existing-volume-upgrade.log` verifies database access for the application role. | None. |
| 16 | Complete | `test-output/review-followup/postgres-existing-volume-upgrade.log` verifies existing ownership normalization and malformed-publication rejection. | None. |
| 17 | Complete | `test-output/review-followup/postgres-existing-volume-upgrade.log` exercises the upgrade on PostgreSQL 18; `test-output/review-followup/tracking-deploy-validation.log` records the post-documentation deployment lint. | None. |
| 18 | Complete | `durable-sqlite-pki-init.log`, `durable-sqlite-runtime.log`, `durable-sqlite-compose-health.log`, and `durable-sqlite-authenticated-readiness.log` exercise initialization and runtime behavior against the image. | None. |
| 19 | Partially verified | `durable-sqlite-compose-parity.log` verifies distinct PS-ID/instance data volumes and project scoping for all 14 SQLite instances. | Product/suite runtime stacks were not exercised; run them in a separate Docker-enabled environment. No further Docker lifecycle operations were run after the recorded volume incident. |
| 20 | Complete | `ci-load.yml` uses certificate-authenticated PostgreSQL readiness; `postgres-authenticated-readiness.log` records a live HTTP 200 response. | Full load workload was not run. |
| 21 | Complete | `ci-load.yml` passes inputs through environment variables and validates numeric ranges; `final-repository-validators.log` passes workflow lint. | None. |
| 22 | Complete | `final-focused-tests.log` passes staged publication and rollback tests; `docs/tls-structure.md` documents the Compose activation gate and unsupported rotation behavior. | None. |
| 23 | Complete | `final-focused-tests.log` passes deterministic filesystem failure-path and platform permission tests. | None. |
| 24 | Complete | `test-output/review-followup/compose-json-parser.log` records 46 shuffled `e2e_infra` package tests; `compose-json-lint.log` records successful `lint-go`, `lint-go-test`, and `lint-workflow`. `ci-e2e.yml` and `ci-load.yml` both provide the action's required `required-completed-services` input. | None. |
| 25 | Complete | `final-focused-tests.log` passes SQLite CRL allocation tests; the implementation rejects SQLite versions without `UPDATE ... RETURNING`. | None. |
| 26 | Complete | `final-focused-tests.log` passes bounded retry/error-classification tests; PostgreSQL classification uses SQLSTATE and SQLite allocation tests pass. | None. |
| 27 | Complete | `docs/ENG-HANDBOOK.md` specifies permitted CRL-number gaps and no reuse; focused allocation/failure tests pass. | None. |
| 29 | Complete | `final-focused-tests.log` passes persistence tests for the zero-UUID sentinel and duplicate request IDs. | None. |
| 30 | Complete | `final-focused-tests.log` and `post-review-focused-tests.log` pass authority-ID, issuer-key, and issuer-DN consistency tests. The latter covers valid legacy-certificate revocation and rejects invalid PEM and serial mismatches; `docs/ENG-HANDBOOK.md` §6.5 documents the verified compatibility rule. | None. |
| 31 | Complete | `docs/ENG-HANDBOOK.md` and focused PKI tests establish issuer-wide CRLs for the shared authority, not tenant-scoped CRLs. | None. |
| 32 | Complete | Authority configuration tests verify deployment-time CRL/OCSP URL and profile metadata precedence after restart. | None. |
| 33 | Complete | `final-repository-validators.log` passes deployment and template checks for endpoint hostnames, ports, and TLS paths across PS-ID configs. | None. |
| 34 | Complete | `postgres-migrations-2012-2015.log` verifies migration defaults; `postgres-continuity-restart.log` verifies persisted envelope-version restoration after PostgreSQL restart. | None. |
| 35 | Complete | `final-focused-tests.log` passes authority creation, encryption/persistence failure, and concurrency tests. | None. |
| 36 | Complete | `final-focused-tests.log` passes storage and handler tests proving stale updates cannot clear revocation metadata. | None. |
| 37 | Complete | `postgres-continuity-restart.log` passes concurrent CRL allocation and verifies issuance/revocation and authority-key continuity across PostgreSQL restart. | None. |

## Post-Implementation Deep-Review Findings

Finding identifiers F-1 and F-2 are distinct from review recommendation numbers.

| Finding | Disposition | Evidence | Remaining work |
|---|---|---|---|
| F-1 | Complete | `post-review-focused-tests.log` passes the Compose credential validation cases; `post-review-deployment-validation.log` passes all 51 deployment validators; `post-review-golangci.log` reports no issues. | None. |
| F-2 | Complete | `post-review-focused-tests.log` passes valid legacy-certificate revocation and rejects invalid PEM and serial mismatches; `docs/ENG-HANDBOOK.md` §6.5 documents the issuer-verification rule. | None. |

**Environment incident during live verification:** An accidental `docker compose down --volumes`
removed the globally named volumes `cryptoutil_postgres_leader_volume`,
`cryptoutil_postgres_follower_volume`, and `cryptoutil_grafana_volume`. Their former contents
cannot be determined or restored in this session. No further Docker lifecycle commands were run.
