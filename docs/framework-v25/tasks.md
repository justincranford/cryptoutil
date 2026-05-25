# Tasks â€” Framework v25: 8-to-7 PS-ID Consolidation

**Status**: 0 of 58 tasks complete (0%)
**Last Updated**: 2026-05-25
**Created**: 2026-05-25

---

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
  - [ ] `2009_pki_ca_items.up.sql` created in `internal/apps/sm-kms/server/repository/migrations/`
  - [ ] `2009_pki_ca_items.down.sql` created
  - [ ] `migrations.go` embed directive updated to include 2009
  - [ ] Migration runs cleanly on fresh SQLite in-memory DB
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
  handler_est.go, handler_ocsp.go) to `internal/apps/sm-kms/server/pki/handler/`.
  Update imports. The handlers will use the ported service packages.
- **Acceptance Criteria**:
  - [ ] All handler files compile
  - [ ] Handler unit tests pass (using Fiber app.Test())
- **Files**:
  - `internal/apps/sm-kms/server/pki/handler/handler.go` (NEW)
  - `internal/apps/sm-kms/server/pki/handler/handler_certs.go` (NEW)
  - `internal/apps/sm-kms/server/pki/handler/handler_est.go` (NEW)
  - `internal/apps/sm-kms/server/pki/handler/handler_ocsp.go` (NEW)
  - `internal/apps/sm-kms/server/pki/handler/handler_test.go` (NEW)
  - `internal/apps/sm-kms/server/pki/handler/handler_certs_test.go` (NEW)
  - `internal/apps/sm-kms/server/pki/handler/handler_est_test.go` (NEW)
  - `internal/apps/sm-kms/server/pki/handler/handler_ocsp_test.go` (NEW)
  - `internal/apps/sm-kms/server/pki/handler/testmain_test.go` (NEW)

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
- **Evidence**: `test-output/phase1/build.log`, `test-output/phase1/test.log`

---

## Phase 2: Extend sm-kms API with PKI Endpoints

**Phase Objective**: Add all 18 PKI endpoints to the sm-kms OpenAPI spec under `/pki/` path
prefix. Update oapi-codegen configs with PKI initialisms. Regenerate API code. Wire PKI
handlers into sm-kms server at both `/service/` and `/browser/` path prefixes.

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
  - [ ] `GET /service/api/v1/pki/profiles` returns 200 in integration test
  - [ ] `GET /browser/api/v1/pki/profiles` returns 200 in integration test
  - [ ] No 404s on any PKI endpoint
- **Files**:
  - `internal/apps/sm-kms/server/server.go` (UPDATE â€” +PKI route registration)

---

### Task 2.7: Add sm-kms PKI Integration Tests

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: Task 2.6
- **Description**: Add integration tests verifying all 18 PKI endpoints are reachable
  and return expected status codes. Use `fiber.app.Test()` pattern (no real listener).
- **Acceptance Criteria**:
  - [ ] All 18 PKI endpoints tested (GET, POST)
  - [ ] Binary content-type endpoints (OCSP, TSA, EST) tested with fixture data
  - [ ] Tests use `t.Parallel()` and table-driven pattern
  - [ ] Tests run under `_integration_test.go` suffix
- **Files**:
  - `internal/apps/sm-kms/server/pki_routes_integration_test.go` (NEW)

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
- **Evidence**: `test-output/phase2/build.log`, `test-output/phase2/test.log`, `test-output/phase2/coverage.out`

---

## Phase 3: Delete pki-ca Service Artifacts

**Phase Objective**: Remove the entire pki-ca standalone service. All its functionality is
now served by sm-kms. Delete API specs, all ~100 internal packages, cmd entry point, configs,
and deployment artifacts. Remove pki-ca TLS cert CN constants from magic_pki_tls.go.

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
- **Evidence**: `test-output/phase3/build.log`

---

## Phase 4: Delete pki Product and Update Registry

**Phase Objective**: Remove the pki product: CLI router, deployment artifacts, magic_pki.go,
and registry entries. Update magic_tier.go, api/cryptosuite-registry/registry.yaml,
lint_fitness entity registry, and lint_ports to reflect the new 3-product / 7-PS-ID topology.

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
  - [ ] No remaining `pki-ca` or `PKIProductName` references in lint_fitness code
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
  - [ ] No pki-ca port references in lint_ports
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
  - [ ] `grep -r "pki-ca\|PKIProductName\|PKICAServiceID" --include="*.go" internal/ api/ cmd/` returns no results
- **Evidence**: `test-output/phase4/lint-fitness.log`, `test-output/phase4/grep-pki-ca.log`

---

## Phase 5: Quality Gates

**Phase Objective**: Comprehensive end-to-end quality verification. Full build, lint, test,
coverage analysis, race detector, and cicd-lint suite. This phase confirms the consolidation
is complete and production-ready.

---

### Task 5.1: Full Build Verification

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 4.10
- **Description**: Build all targets including E2E and integration build tags.
- **Acceptance Criteria**:
  - [ ] `go build ./...` zero errors
  - [ ] `go build -tags e2e,integration ./...` zero errors
- **Evidence**: `test-output/phase5/build.log`

---

### Task 5.2: Full Linting Verification

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 5.1
- **Description**: Run golangci-lint with and without E2E/integration build tags.
- **Acceptance Criteria**:
  - [ ] `golangci-lint run ./...` zero warnings
  - [ ] `golangci-lint run --build-tags e2e,integration ./...` zero warnings
- **Evidence**: `test-output/phase5/lint.log`

---

### Task 5.3: Full Test Suite Verification

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 5.2
- **Description**: Run full test suite with shuffle and verify 100% passing.
- **Acceptance Criteria**:
  - [ ] `go test ./... -shuffle=on` 100% passing, zero skips
  - [ ] `go test -race -count=2 ./internal/apps/sm-kms/...` passes
- **Evidence**: `test-output/phase5/tests.log`

---

### Task 5.4: Coverage Analysis for New pki Packages

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 5.3
- **Description**: Run coverage and verify all new sm-kms/server/pki/ packages meet â‰¥95%.
- **Acceptance Criteria**:
  - [ ] All `internal/apps/sm-kms/server/pki/...` packages â‰¥95% coverage
  - [ ] No regressions in existing sm-kms package coverage
- **Evidence**: `test-output/phase5/coverage.out`, `test-output/phase5/coverage.html`

---

### Task 5.5: cicd-lint Full Suite

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 5.4
- **Description**: Run all relevant cicd-lint commands.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-text` passes
  - [ ] `go run ./cmd/cicd-lint lint-go` passes
  - [ ] `go run ./cmd/cicd-lint lint-go-test` passes
  - [ ] `go run ./cmd/cicd-lint lint-go-mod` passes
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes (3 products / 7 PS-IDs)
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes
  - [ ] `go run ./cmd/cicd-lint lint-openapi` passes
- **Evidence**: `test-output/phase5/cicd-lint.log`

---

## Phase 6: Knowledge Propagation

**Phase Objective**: Apply all lessons learned to permanent project artifacts. Update
ENG-HANDBOOK.md, agents, skills, instructions, and docs. Remove pki-ca from all documentation.
Verify propagation integrity.

---

### Task 6.1: Review All lessons.md Phase Entries

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 5.5
- **Description**: Review all 5 prior phase lessons entries. Identify patterns and insights
  to propagate to permanent artifacts.
- **Acceptance Criteria**:
  - [ ] All 5 phase lessons reviewed
  - [ ] List of propagation targets identified

---

### Task 6.2: Update ENG-HANDBOOK.md

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Task 6.1
- **Description**: Update ENG-HANDBOOK.md:
  - Remove pki-ca from service catalog
  - Update sm-kms API list to include PKI endpoints
  - Update product/PS-ID count (3 products / 7 PS-IDs)
  - Add any new patterns discovered during pki-ca migration
- **Acceptance Criteria**:
  - [ ] Service catalog reflects 7 PS-IDs
  - [ ] sm-kms description mentions PKI under `/pki/` path prefix
  - [ ] No stale pki-ca references
- **Files**: `docs/ENG-HANDBOOK.md` (UPDATE)

---

### Task 6.3: Update Agents, Skills, Instructions

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 6.2
- **Description**: Scan agents, skills, and instruction files for pki-ca references and
  update them to reflect the new topology. Run lint-docs to verify propagation.
- **Acceptance Criteria**:
  - [ ] No stale pki-ca references in agents or instructions
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**: Various `.github/agents/`, `.github/instructions/`, `.github/skills/` files as needed

---

### Task 6.4: Final Commit and Cleanup

- **Status**: âŒ Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 6.3
- **Description**: Final clean commit. Verify `git status --porcelain` returns empty.
  Update lessons.md Executive Summary and Actions sections.
- **Acceptance Criteria**:
  - [ ] `git status --porcelain` returns empty
  - [ ] lessons.md Executive Summary and Actions sections filled
  - [ ] All commits follow conventional commit format

---

## Cross-Cutting Tasks

### Testing

- [ ] Unit tests â‰¥95% coverage for all new pki packages in sm-kms
- [ ] Integration tests verify all 18 PKI endpoints reachable
- [ ] Race detector clean: `go test -race ./internal/apps/sm-kms/...`
- [ ] No skipped tests without documented justification

### Code Quality

- [ ] Linting passes: `golangci-lint run ./...` and with `--build-tags e2e,integration`
- [ ] No new TODOs without tracking in tasks.md
- [ ] Formatting clean after `golangci-lint run --fix ./...` + re-run pass
- [ ] No `pki-ca` or `PKIProductName` imports in any non-deleted file

### Registry Integrity

- [ ] `api/cryptosuite-registry/registry.yaml` shows 3 products / 7 PS-IDs
- [ ] lint_fitness entity registry consistent with registry.yaml
- [ ] magic_tier.go consistent with registry.yaml

### Documentation

- [ ] ENG-HANDBOOK.md service catalog updated
- [ ] README.md updated if it references pki-ca
- [ ] No stale references to pki port 8300 in documentation

---

## Notes / Deferred Work

- **Database-backed PKI storage**: pki-ca uses in-memory storage. Migrating to DB-backed
  storage (GORM + PostgreSQL) is a separate feature planned for a future phase.
- **pki port range 8300**: Freed by this plan but not reallocated. Available for future use.
- **pki migration range 5001â€“5999**: Freed but not reallocated.
- **pg_host_port 54323**: Freed but not reallocated.

---

## Evidence Archive

- `test-output/phase1/` â€” Phase 1 build and test logs
- `test-output/phase2/` â€” Phase 2 build, test, and coverage logs
- `test-output/phase3/` â€” Phase 3 post-deletion build logs
- `test-output/phase4/` â€” Phase 4 registry and lint-fitness logs
- `test-output/phase5/` â€” Full quality gate logs and coverage report
- `test-output/tokens/TOKENS-260525-004251.md` â€” Token tracking for planning session
