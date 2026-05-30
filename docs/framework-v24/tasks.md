# Tasks - Framework v24: 10-to-8 PS-ID Consolidation

**Status**: Execution in progress (major Phases 1-4 implementation completed; Phase 5 full-suite migration validation still in progress)
**Last Updated**: 2026-05-29
**Created**: 2026-05-25

## Execution Checkpoint (2026-05-29)

- Implemented merged sm-kms compatibility routes/handlers/repositories/migrations for former jose-ja and sm-im APIs.
- Removed jose-ja/sm-im/jose runtime directories from api, cmd, internal/apps, configs, and deployments.
- Updated topology artifacts to 4 products / 8 PS-IDs in registry/config/deployment wiring and lint tooling.
- Verified clean compile gates: `go build ./...` and `go build -tags e2e,integration ./...`.
- Verified lint gates: `golangci-lint run --fix`, `golangci-lint run`, `go run ./cmd/cicd-lint lint-fitness`, `go run ./cmd/cicd-lint lint-deployments lint-openapi lint-docs`.
- Remaining blocker for full Phase 5 completion: broad repository test suite still has legacy 10-PS-ID assumptions in many test packages that must be migrated to 8-PS-ID topology.

---

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- [x] **Correctness**: ALL code must be functionally correct with comprehensive tests
- [x] **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- [x] **Thoroughness**: Evidence-based validation at every step
- [x] **Reliability**: Quality gates enforced (>=95%/98% coverage/mutation)
- [x] **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- [x] **Accuracy**: Changes must address root cause, not just symptoms
- [ ] **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- [ ] **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions:**
- [x] **Fix issues immediately** - blockers must be resolved before advancing
- [x] **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- [x] **NEVER skip**: Cannot mark phase or task or step complete with known issues

---

## Task Status Legend

| Symbol | Meaning | When to Use |
|--------|---------|-------------|
| [ ] | Not started | Task not yet begun |
| [~] | In progress | Currently being worked on |
| [x] | Complete | Task finished with evidence |
| [!] | Blocked | Requires external dependency (MUST have resolution plan) |

---

## Phase 1: jose-ja Domain -> sm-kms

**Phase Objective**: Port all jose-ja domain models, repositories, services, and API handlers
into sm-kms. After this phase, sm-kms can serve all jose-ja API endpoints AND all existing
sm-kms endpoints simultaneously.

---

#### Task 1.1: sm-kms DB Migrations for JWK Domain (2003-2006)

- **Status**: [ ] Not Started
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Create SQL migration files for the four jose-ja tables in the sm-kms migration range (2003-2006). Copy and adapt from `internal/apps/jose-ja/server/repository/migrations/4001-4004`.
- **Acceptance Criteria**:
  - [ ] `2003_jwk_elastic_jwks.up.sql` and `.down.sql` created
  - [ ] `2004_jwk_material_jwks.up.sql` and `.down.sql` created
  - [ ] `2005_jwk_audit_config.up.sql` and `.down.sql` created
  - [ ] `2006_jwk_audit_log.up.sql` and `.down.sql` created
  - [ ] Table names match jose-ja originals (`elastic_jwks`, `material_jwks`, `tenant_audit_config`, `audit_log`)
  - [ ] Migrations embed correctly in `migrations.go`
- **Files**:
  - `internal/apps/sm-kms/server/repository/migrations/2003_jwk_elastic_jwks.up.sql`
  - `internal/apps/sm-kms/server/repository/migrations/2003_jwk_elastic_jwks.down.sql`
  - `internal/apps/sm-kms/server/repository/migrations/2004_jwk_material_jwks.up.sql`
  - `internal/apps/sm-kms/server/repository/migrations/2004_jwk_material_jwks.down.sql`
  - `internal/apps/sm-kms/server/repository/migrations/2005_jwk_audit_config.up.sql`
  - `internal/apps/sm-kms/server/repository/migrations/2005_jwk_audit_config.down.sql`
  - `internal/apps/sm-kms/server/repository/migrations/2006_jwk_audit_log.up.sql`
  - `internal/apps/sm-kms/server/repository/migrations/2006_jwk_audit_log.down.sql`
  - `internal/apps/sm-kms/server/repository/migrations.go` (embed updated)

---

#### Task 1.2: JWK Domain Models in sm-kms

- **Status**: [ ] Not Started
- **Estimated**: 1h
- **Dependencies**: Task 1.1
- **Description**: Create `jwk_models.go` in sm-kms model package with `ElasticJWK`, `MaterialJWK`, `AuditConfig`, `AuditLogEntry` structs (ported from jose-ja).
- **Acceptance Criteria**:
  - [ ] `ElasticJWK` struct with GORM tags (`elastic_jwks` table)
  - [ ] `MaterialJWK` struct with GORM tags (`material_jwks` table)
  - [ ] `AuditConfig` struct (`tenant_audit_config` table)
  - [ ] `AuditLogEntry` struct (`audit_log` table)
  - [ ] Cross-DB compatible field types (`type:text` for UUIDs, `serializer:json` for arrays)
  - [ ] `jwk_models_test.go` with table validation tests
  - [ ] Build and tests pass
- **Files**:
  - `internal/apps/sm-kms/server/model/jwk_models.go`
  - `internal/apps/sm-kms/server/model/jwk_models_test.go`

---

#### Task 1.3: JWK Repositories in sm-kms

- **Status**: [ ] Not Started
- **Estimated**: 2h
- **Dependencies**: Task 1.2
- **Description**: Port the three jose-ja repositories (elastic JWK, material JWK, audit) to sm-kms.
- **Acceptance Criteria**:
  - [ ] `ElasticJWKRepository` interface + impl (CRUD with tenant filtering)
  - [ ] `MaterialJWKRepository` interface + impl (CRUD + active key query)
  - [ ] `AuditRepository` interface + impl (create + list audit events)
  - [ ] All repositories use `getDB(ctx, r.db)` context transaction pattern
  - [ ] All fields use `type:text` for UUID columns
  - [ ] Unit tests with in-memory SQLite (`testdb.NewInMemorySQLiteDB(t)`)
  - [ ] Tests use `t.Parallel()` throughout
  - [ ] Coverage >=95%
- **Files**:
  - `internal/apps/sm-kms/server/repository/elastic_jwk_repository.go`
  - `internal/apps/sm-kms/server/repository/elastic_jwk_repository_test.go`
  - `internal/apps/sm-kms/server/repository/material_jwk_repository.go`
  - `internal/apps/sm-kms/server/repository/material_jwk_repository_test.go`
  - `internal/apps/sm-kms/server/repository/audit_repository.go`
  - `internal/apps/sm-kms/server/repository/audit_repository_test.go`

---

#### Task 1.4: JWK Business Logic Services in sm-kms

- **Status**: [ ] Not Started
- **Estimated**: 3h
- **Dependencies**: Task 1.3
- **Description**: Create `server/jwkservice/` package with all jose-ja services ported to sm-kms.
- **Acceptance Criteria**:
  - [ ] `ElasticJWKService` interface + impl (create, get, list, delete)
  - [ ] `MaterialRotationService` interface + impl (rotate, get active, list)
  - [ ] `JWKSService` interface + impl (generate public JWKS)
  - [ ] `JWSService` interface + impl (sign/verify)
  - [ ] `JWEService` interface + impl (encrypt/decrypt)
  - [ ] `JWTService` interface + impl (create/verify JWTs)
  - [ ] `AuditLogService` interface + impl (record events)
  - [ ] Services use constructor injection (no package-level vars)
  - [ ] `testmain_test.go` for shared test setup
  - [ ] All service unit tests with in-memory SQLite + table-driven cases
  - [ ] Coverage >=95%
- **Files**:
  - `internal/apps/sm-kms/server/jwkservice/elastic_jwk_service.go`
  - `internal/apps/sm-kms/server/jwkservice/elastic_jwk_service_test.go`
  - `internal/apps/sm-kms/server/jwkservice/material_rotation_service.go`
  - `internal/apps/sm-kms/server/jwkservice/material_rotation_service_test.go`
  - `internal/apps/sm-kms/server/jwkservice/jwks_service.go`
  - `internal/apps/sm-kms/server/jwkservice/jwks_service_test.go`
  - `internal/apps/sm-kms/server/jwkservice/jws_service.go`
  - `internal/apps/sm-kms/server/jwkservice/jws_service_test.go`
  - `internal/apps/sm-kms/server/jwkservice/jwe_service.go`
  - `internal/apps/sm-kms/server/jwkservice/jwe_service_test.go`
  - `internal/apps/sm-kms/server/jwkservice/jwt_service.go`
  - `internal/apps/sm-kms/server/jwkservice/jwt_service_test.go`
  - `internal/apps/sm-kms/server/jwkservice/audit_log_service.go`
  - `internal/apps/sm-kms/server/jwkservice/audit_log_service_test.go`
  - `internal/apps/sm-kms/server/jwkservice/testmain_test.go`

---

#### Task 1.5: JWK API Handlers in sm-kms

- **Status**: [ ] Not Started
- **Estimated**: 2h
- **Dependencies**: Task 1.4
- **Description**: Port jose-ja API handlers into sm-kms `server/handler/` package.
- **Acceptance Criteria**:
  - [ ] `jwk_handler.go` with elastic JWK CRUD + rotate + active material key handlers
  - [ ] `jwks_handler.go` with `GET /jwks` handler
  - [ ] All handlers use Fiber `app.Test()` in tests (no real listeners)
  - [ ] Handler tests are table-driven
  - [ ] Coverage >=95%
- **Files**:
  - `internal/apps/sm-kms/server/handler/jwk_handler.go`
  - `internal/apps/sm-kms/server/handler/jwk_handler_test.go`
  - `internal/apps/sm-kms/server/handler/jwks_handler.go`
  - `internal/apps/sm-kms/server/handler/jwks_handler_test.go`

---

#### Task 1.6: Extend sm-kms OpenAPI Spec (JWK Endpoints)

- **Status**: [ ] Not Started
- **Estimated**: 1.5h
- **Dependencies**: Task 1.5
- **Description**: Add jose-ja endpoint paths and component schemas to sm-kms OpenAPI spec.
- **Acceptance Criteria**:
  - [ ] `openapi_spec_paths.yaml` gains JWK paths (at minimum: active material key, rotate, JWKS)
  - [ ] All jose-ja `/elastic-keys/` paths present (review against `api/jose-ja/openapi_spec.yaml`)
  - [ ] `openapi_spec_components.yaml` gains ElasticJWK, MaterialJWK, JWKS, AuditConfig schemas
  - [ ] `oapi-codegen` gen configs updated with JWK+JWKS+OKP+URI initialisms if missing
  - [ ] Spec validates: `go run ./cmd/cicd-lint lint-openapi`
- **Files**:
  - `api/sm-kms/openapi_spec_paths.yaml`
  - `api/sm-kms/openapi_spec_components.yaml`
  - `api/sm-kms/openapi-gen_config_server.yaml` (if initialism update needed)
  - `api/sm-kms/openapi-gen_config_model.yaml` (if initialism update needed)
  - `api/sm-kms/openapi-gen_config_client.yaml` (if initialism update needed)

---

#### Task 1.7: Regenerate sm-kms oapi-codegen Outputs (Post JWK)

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 1.6
- **Description**: Run oapi-codegen to regenerate server, model, and client stubs for sm-kms.
- **Acceptance Criteria**:
  - [ ] `api/sm-kms/server/server.gen.go` regenerated (includes JWK strict-server interfaces)
  - [ ] `api/sm-kms/models/models.gen.go` regenerated (includes JWK model types)
  - [ ] `api/sm-kms/client/client.gen.go` regenerated
  - [ ] `go build ./...` clean after regeneration
- **Files**:
  - `api/sm-kms/server/server.gen.go`
  - `api/sm-kms/models/models.gen.go`
  - `api/sm-kms/client/client.gen.go`

---

#### Task 1.8: Register JWK Routes in sm-kms server.go

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 1.7
- **Description**: Wire JWK handlers and JWKS handler into sm-kms server routing.
- **Acceptance Criteria**:
  - [ ] JWK routes registered in `internal/apps/sm-kms/server/server.go`
  - [ ] JWKS route registered
  - [ ] Routes appear at both `/service/api/v1/` and `/browser/api/v1/` paths
  - [ ] `go build ./...` clean
- **Files**:
  - `internal/apps/sm-kms/server/server.go`

---

#### Task 1.9: Phase 1 Quality Gate

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Tasks 1.1-1.8
- **Description**: Verify Phase 1 quality gates all pass.
- **Acceptance Criteria**:
  - [ ] `go build ./...` zero errors
  - [ ] `golangci-lint run ./...` zero warnings
  - [ ] `go test ./internal/apps/sm-kms/...` 100% pass
  - [ ] `go test ./internal/apps/jose-ja/...` still passes (jose-ja not deleted yet)
  - [ ] sm-kms coverage >=95%
  - [ ] `go run ./cmd/cicd-lint lint-openapi` passes

---

## Phase 2: sm-im Domain -> sm-kms

**Phase Objective**: Port all sm-im domain models, repository, and handler into sm-kms.
After this phase, sm-kms can send and receive encrypted messages.

---

#### Task 2.1: sm-kms DB Migrations for Message Domain (2007-2008)

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 1.9 (Phase 1 must be complete)
- **Description**: Create SQL migration files for sm-im tables in sm-kms migration range.
- **Acceptance Criteria**:
  - [ ] `2007_im_messages.up.sql` and `.down.sql` created
  - [ ] `2008_im_recipient_jwks.up.sql` and `.down.sql` created
  - [ ] Table names match sm-im originals (`messages`, `messages_recipient_jwks`)
  - [ ] Migrations embed correctly in updated `migrations.go`
- **Files**:
  - `internal/apps/sm-kms/server/repository/migrations/2007_im_messages.up.sql`
  - `internal/apps/sm-kms/server/repository/migrations/2007_im_messages.down.sql`
  - `internal/apps/sm-kms/server/repository/migrations/2008_im_recipient_jwks.up.sql`
  - `internal/apps/sm-kms/server/repository/migrations/2008_im_recipient_jwks.down.sql`
  - `internal/apps/sm-kms/server/repository/migrations.go` (embed updated)

---

#### Task 2.2: Message Domain Models in sm-kms

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 2.1
- **Description**: Create `message_models.go` in sm-kms model package with `Message` and `MessageRecipientJWK` structs.
- **Acceptance Criteria**:
  - [ ] `Message` struct with GORM tags (`messages` table)
  - [ ] `MessageRecipientJWK` struct with GORM tags (`messages_recipient_jwks` table)
  - [ ] Cross-DB compatible field types
  - [ ] `message_models_test.go` with validation tests
- **Files**:
  - `internal/apps/sm-kms/server/model/message_models.go`
  - `internal/apps/sm-kms/server/model/message_models_test.go`

---

#### Task 2.3: Message Repositories in sm-kms

- **Status**: [ ] Not Started
- **Estimated**: 1h
- **Dependencies**: Task 2.2
- **Description**: Port the sm-im message repositories to sm-kms.
- **Acceptance Criteria**:
  - [ ] `MessageRepository` interface + impl (CRUD with sender/recipient filtering)
  - [ ] `MessageRecipientJWKRepository` interface + impl
  - [ ] Context transaction pattern used (`getDB(ctx, r.db)`)
  - [ ] Unit tests with in-memory SQLite
  - [ ] Coverage >=95%
- **Files**:
  - `internal/apps/sm-kms/server/repository/message_repository.go`
  - `internal/apps/sm-kms/server/repository/message_repository_test.go`
  - `internal/apps/sm-kms/server/repository/message_recipient_jwk_repository.go`
  - `internal/apps/sm-kms/server/repository/message_recipient_jwk_repository_test.go`

---

#### Task 2.4: Message Handler in sm-kms

- **Status**: [ ] Not Started
- **Estimated**: 2h
- **Dependencies**: Task 2.3
- **Description**: Port sm-im message handler into `internal/apps/sm-kms/server/handler/message_handler.go`.
- **Acceptance Criteria**:
  - [ ] `MessageHandler` struct with constructor injection
  - [ ] `HandleSendMessage`, `HandleReceiveMessages`, `HandleGetMessage`, `HandleDeleteMessage`, `HandleListMessages` handlers
  - [ ] All handlers tested using Fiber `app.Test()`
  - [ ] Table-driven tests
  - [ ] Coverage >=95%
- **Files**:
  - `internal/apps/sm-kms/server/handler/message_handler.go`
  - `internal/apps/sm-kms/server/handler/message_handler_test.go`

---

#### Task 2.5: Extend sm-kms OpenAPI Spec (Messaging Endpoints)

- **Status**: [ ] Not Started
- **Estimated**: 1h
- **Dependencies**: Task 2.4
- **Description**: Add sm-im messaging paths and component schemas to sm-kms OpenAPI spec.
- **Acceptance Criteria**:
  - [ ] `openapi_spec_paths.yaml` gains message paths (send, receive, get, list, delete)
  - [ ] `openapi_spec_components.yaml` gains Message, SendMessageRequest, MessageRecipient schemas
  - [ ] Spec validates: `go run ./cmd/cicd-lint lint-openapi`
- **Files**:
  - `api/sm-kms/openapi_spec_paths.yaml`
  - `api/sm-kms/openapi_spec_components.yaml`

---

#### Task 2.6: Regenerate sm-kms oapi-codegen Outputs (Post Messaging)

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 2.5
- **Description**: Regenerate sm-kms server/model/client stubs after messaging spec additions.
- **Acceptance Criteria**:
  - [ ] `server.gen.go`, `models.gen.go`, `client.gen.go` regenerated
  - [ ] `go build ./...` clean
- **Files**:
  - `api/sm-kms/server/server.gen.go`
  - `api/sm-kms/models/models.gen.go`
  - `api/sm-kms/client/client.gen.go`

---

#### Task 2.7: Register Message Routes in sm-kms server.go

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 2.6
- **Description**: Wire message handler into sm-kms server routing.
- **Acceptance Criteria**:
  - [ ] Message routes registered in `internal/apps/sm-kms/server/server.go`
  - [ ] Routes at both `/service/api/v1/` and `/browser/api/v1/` paths
  - [ ] `go build ./...` clean
- **Files**:
  - `internal/apps/sm-kms/server/server.go`

---

#### Task 2.8: Phase 2 Quality Gate

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Tasks 2.1-2.7
- **Description**: Verify Phase 2 quality gates all pass.
- **Acceptance Criteria**:
  - [ ] `go build ./...` zero errors
  - [ ] `golangci-lint run ./...` zero warnings
  - [ ] `go test ./internal/apps/sm-kms/...` 100% pass (incl. message tests)
  - [ ] `go test ./internal/apps/sm-im/...` still passes
  - [ ] sm-kms coverage >=95%

---

## Phase 3: Delete jose-ja, sm-im, jose Product

**Phase Objective**: Remove all jose-ja and sm-im artifacts from the codebase. The jose product
is deleted because it has no remaining PS-IDs after jose-ja is removed.

---

#### Task 3.1: Delete jose-ja API Directory

- **Status**: [ ] Not Started
- **Estimated**: 0.25h
- **Dependencies**: Task 2.8 (Phase 2 must be complete)
- **Description**: Delete `api/jose-ja/` entirely.
- **Acceptance Criteria**:
  - [ ] `api/jose-ja/` directory and all contents removed
  - [ ] No import of `cryptoutil/api/jose-ja/...` exists in any Go file
  - [ ] `go build ./...` clean

---

#### Task 3.2: Delete sm-im API Directory

- **Status**: [ ] Not Started
- **Estimated**: 0.25h
- **Dependencies**: Task 3.1
- **Description**: Delete `api/sm-im/` entirely.
- **Acceptance Criteria**:
  - [ ] `api/sm-im/` directory and all contents removed
  - [ ] No import of `cryptoutil/api/sm-im/...` exists in any Go file
  - [ ] `go build ./...` clean

---

#### Task 3.3: Delete jose-ja Internal App Directory

- **Status**: [ ] Not Started
- **Estimated**: 0.25h
- **Dependencies**: Task 3.2
- **Description**: Delete `internal/apps/jose-ja/` entirely (75 Go files).
- **Acceptance Criteria**:
  - [ ] `internal/apps/jose-ja/` and all contents removed
  - [ ] No import of `cryptoutil/internal/apps/jose-ja/...` in any Go file
  - [ ] `go build ./...` clean

---

#### Task 3.4: Delete sm-im Internal App Directory

- **Status**: [ ] Not Started
- **Estimated**: 0.25h
- **Dependencies**: Task 3.3
- **Description**: Delete `internal/apps/sm-im/` entirely (60 Go files).
- **Acceptance Criteria**:
  - [ ] `internal/apps/sm-im/` and all contents removed
  - [ ] No import of `cryptoutil/internal/apps/sm-im/...` in any Go file
  - [ ] `go build ./...` clean

---

#### Task 3.5: Delete cmd and internal/apps/jose (product coordinator)

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 3.4
- **Description**: Delete `cmd/jose-ja/`, `cmd/jose/`, and `internal/apps/jose/` (the product-level jose coordinator). Update `cmd/cryptoutil/main.go` to remove jose routing.
- **Acceptance Criteria**:
  - [ ] `cmd/jose-ja/` removed
  - [ ] `cmd/jose/` removed
  - [ ] `internal/apps/jose/` removed (if it exists as a product coordinator)
  - [ ] `cmd/cryptoutil/main.go` no longer routes to jose
  - [ ] `go build ./...` clean

---

#### Task 3.6: Delete cmd/sm-im and update cmd/sm

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 3.5
- **Description**: Delete `cmd/sm-im/` and remove sm-im routing from `cmd/sm/main.go` and `internal/apps/sm/sm.go`.
- **Acceptance Criteria**:
  - [ ] `cmd/sm-im/` removed
  - [ ] `cmd/sm/main.go` no longer routes to sm-im
  - [ ] `internal/apps/sm/sm.go` no longer references sm-im
  - [ ] `go build ./...` clean

---

#### Task 3.7: Delete configs and deployments for jose-ja, jose, sm-im

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 3.6
- **Description**: Delete deployment and config directories for removed services.
- **Acceptance Criteria**:
  - [ ] `configs/jose-ja/` removed
  - [ ] `configs/sm-im/` removed
  - [ ] `deployments/jose-ja/` removed
  - [ ] `deployments/jose/` removed
  - [ ] `deployments/sm-im/` removed
  - [ ] `deployments/sm/compose.yml` no longer includes or references sm-im service
  - [ ] `deployments/cryptoutil/compose.yml` no longer includes or references jose service block
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes

---

#### Task 3.8: Delete magic_jose.go and magic_sm_im.go

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 3.7
- **Description**: Delete the magic constant files for removed services. Clean up any residual references.
- **Acceptance Criteria**:
  - [ ] `internal/shared/magic/magic_jose.go` deleted
  - [ ] `internal/shared/magic/magic_sm_im.go` deleted
  - [ ] No remaining references to `OTLPServiceJoseJA`, `OTLPServiceSMIM`, `JoseProductName`, `JoseJAServiceID` in production code
  - [ ] `go build ./...` clean
  - [ ] `golangci-lint run ./...` clean

---

#### Task 3.9: Phase 3 Quality Gate

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Tasks 3.1-3.8
- **Description**: Verify Phase 3 quality gates all pass and no stale references remain.
- **Acceptance Criteria**:
  - [ ] `go build ./...` zero errors
  - [ ] `golangci-lint run ./...` zero warnings
  - [ ] `grep -r "jose-ja\|sm-im\|jose_ja\|OTLPServiceJoseJA\|OTLPServiceSMIM" internal/ api/ cmd/ configs/ deployments/` returns zero results (excluding docs/)
  - [ ] `go test ./...` 100% pass

---

## Phase 4: Registry, Magic Constants, Fitness Linters

**Phase Objective**: Update the canonical registry and all derived constants to reflect 4 products
and 8 PS-IDs. All fitness linters and deployment linters pass with zero errors.

---

#### Task 4.1: Update api/cryptosuite-registry/registry.yaml

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 3.9 (Phase 3 must be complete)
- **Description**: Remove jose-ja, sm-im, and jose entries from the registry.
- **Acceptance Criteria**:
  - [ ] `jose-ja` removed from `product_services`
  - [ ] `sm-im` removed from `product_services`
  - [ ] `jose` removed from `products`
  - [ ] sm-kms `migration_range_end` updated to reflect 2001-2999 range (unchanged since new migrations are within range)
  - [ ] PostgreSQL ports 54321 (sm-im) and 54322 (jose-ja) noted as freed
  - [ ] Registry count comments updated (10 PS-IDs -> 8, 5 products -> 4)

---

#### Task 4.2: Update magic_tier.go

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 4.1
- **Description**: Remove jose product tier and sm-im from SM tier.
- **Acceptance Criteria**:
  - [ ] `JoseProductName` entry and its `{OTLPServiceJoseJA}` slice removed from tier map
  - [ ] `OTLPServiceSMIM` removed from `SMProductName` entry in tier map
  - [ ] `OTLPServiceJoseJA` removed from all PS-ID lists (AllServices, etc.)
  - [ ] `OTLPServiceSMIM` removed from all PS-ID lists
  - [ ] `go build ./...` and `golangci-lint run ./...` clean

---

#### Task 4.3: Update magic_cicd.go, magic_pki_tls.go, magic_sm.go

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 4.2
- **Description**: Remove jose-ja and sm-im references from remaining magic files.
- **Acceptance Criteria**:
  - [ ] `magic_cicd.go`: service count comments updated (10->8, "sm, jose, pki, identity, skeleton"->"sm, pki, identity, skeleton")
  - [ ] `magic_pki_tls.go`: `AppJoseJASQLite1ServerCertCN`, `AppJoseJASQLite2ServerCertCN`, `AppJoseJAPostgres1ServerCertCN`, `AppJoseJAPostgres2ServerCertCN` constants removed
  - [ ] `magic_sm.go`: IM-specific constants removed if any
  - [ ] `go build ./...` clean

---

#### Task 4.4: Update cicd_lint port range and legacy port files

- **Status**: [ ] Not Started
- **Estimated**: 0.75h
- **Dependencies**: Task 4.3
- **Description**: Remove jose-ja (8200 range) and sm-im (8100 range) from port range definitions and legacy port lists.
- **Acceptance Criteria**:
  - [ ] `lint_ports/host_port_ranges/*.go`: jose-ja range (8200-8299) removed or marked inactive
  - [ ] `lint_ports/host_port_ranges/*.go`: sm-im range (8100-8199) removed or marked inactive
  - [ ] `lint_ports/legacy_ports/*.go`: jose-ja legacy port entries removed
  - [ ] `lint_ports/legacy_ports/*.go`: sm-im legacy port entries removed
  - [ ] Associated tests updated to remove jose-ja and sm-im test cases
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_ports/...` passes

---

#### Task 4.5: Run lint-fitness and lint-deployments

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 4.4
- **Description**: Run all cicd-lint commands to verify registry-driven fitness linters pass.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes (entity-registry-completeness, etc.)
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes
  - [ ] `go run ./cmd/cicd-lint lint-go lint-go-test` passes
  - [ ] `go run ./cmd/cicd-lint lint-ports` passes
  - [ ] Zero violations in all checks

---

#### Task 4.6: Phase 4 Quality Gate

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Tasks 4.1-4.5
- **Description**: Full quality gate verification after registry updates.
- **Acceptance Criteria**:
  - [ ] `go build ./...` zero errors
  - [ ] `golangci-lint run ./...` zero warnings
  - [ ] `go test ./...` 100% pass
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes

---

## Phase 5: Full Quality Gate Verification

**Phase Objective**: Comprehensive end-to-end quality validation of the fully consolidated codebase.

---

#### Task 5.1: Build Verification

- **Status**: [ ] Not Started
- **Estimated**: 0.25h
- **Dependencies**: Task 4.6
- **Description**: Verify both tagged and untagged builds are clean.
- **Acceptance Criteria**:
  - [ ] `go build ./...` zero errors
  - [ ] `go build -tags e2e,integration ./...` zero errors
  - [ ] `go vet ./...` zero issues

---

#### Task 5.2: Lint Verification

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 5.1
- **Description**: Full linting pass across all packages.
- **Acceptance Criteria**:
  - [ ] `golangci-lint run ./...` zero warnings
  - [ ] `golangci-lint run --build-tags e2e,integration ./...` zero warnings
  - [ ] `go run ./cmd/cicd-lint lint-go lint-go-test lint-golangci lint-text` all pass

---

#### Task 5.3: Test Suite Verification

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 5.2
- **Description**: Full test pass including shuffle and race detection.
- **Acceptance Criteria**:
  - [ ] `go test ./... -shuffle=on` 100% pass
  - [ ] `go test -race -count=2 ./...` passes (race detector clean)
  - [ ] Zero skipped tests (or all skips documented with tracking)

---

#### Task 5.4: Coverage Verification

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 5.3
- **Description**: Verify coverage targets met for all packages.
- **Acceptance Criteria**:
  - [ ] `go test -coverprofile=test-output/coverage-v24/coverage.out ./...`
  - [ ] sm-kms coverage >=95%
  - [ ] cicd_lint/* coverage >=98%
  - [ ] No production packages below 95%

---

#### Task 5.5: cicd-lint Verification

- **Status**: [ ] Not Started
- **Estimated**: 0.25h
- **Dependencies**: Task 5.4
- **Description**: All cicd-lint commands pass.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes
  - [ ] `go run ./cmd/cicd-lint lint-openapi` passes

---

## Phase 6: Knowledge Propagation

**Phase Objective**: Apply lessons learned from the consolidation to permanent project artifacts.

---

#### Task 6.1: Update docs/ENG-HANDBOOK.md

- **Status**: [ ] Not Started
- **Estimated**: 1h
- **Dependencies**: Task 5.5
- **Description**: Update ENG-HANDBOOK.md to reflect the 4-product, 8-PS-ID architecture.
- **Acceptance Criteria**:
  - [ ] Section 3 (product suite architecture) updated: 5 products -> 4 products, 10 PS-IDs -> 8 PS-IDs
  - [ ] Service table updated
  - [ ] Migration range documentation updated (sm-kms now 2001-2999 inclusive of merged tables)
  - [ ] Decision record added: why jose-ja and sm-im were merged into sm-kms
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes after update

---

#### Task 6.2: Update README.md and docs/DEV-SETUP.md

- **Status**: [ ] Not Started
- **Estimated**: 0.5h
- **Dependencies**: Task 6.1
- **Description**: Update user-facing documentation to remove jose-ja and sm-im references.
- **Acceptance Criteria**:
  - [ ] `README.md` service table shows 8 services not 10
  - [ ] `docs/DEV-SETUP.md` no longer references jose-ja or sm-im setup steps
  - [ ] No dead links to deleted services in docs

---

#### Task 6.3: Propagation Check

- **Status**: [ ] Not Started
- **Estimated**: 0.25h
- **Dependencies**: Task 6.2
- **Description**: Verify propagation integrity between ENG-HANDBOOK.md and instruction files.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes (validate-propagation)

---

#### Task 6.4: Final Commit and Tag

- **Status**: [ ] Not Started
- **Estimated**: 0.25h
- **Dependencies**: Task 6.3
- **Description**: All changes committed. git status clean. Final review.
- **Acceptance Criteria**:
  - [ ] `git status --porcelain` returns empty
  - [ ] Each phase's commits are semantically coherent
  - [ ] No uncommitted changes

---

## Cross-Cutting Tasks

### Testing

- [ ] Unit tests >=95% coverage (production), >=98% (infrastructure/utility)
- [ ] Integration tests pass
- [ ] Table-driven test pattern for all multi-case tests
- [ ] `t.Parallel()` on all tests and subtests
- [ ] No hardcoded UUIDs - use `googleUuid.NewV7()`
- [ ] Fiber `app.Test()` for all handler tests (no real network listeners)
- [ ] Race detector clean: `go test -race -count=2 ./...`
- [ ] No skipped tests without documented exceptions

### Code Quality

- [ ] Linting clean: `golangci-lint run ./...` zero warnings
- [ ] No new TODOs without tracking
- [ ] Import aliases follow `cryptoutilApps*` conventions
- [ ] Formatting clean: `gofumpt -w .`

### Documentation

- [ ] README.md service count updated (8 instead of 10)
- [ ] ENG-HANDBOOK.md architecture section updated
- [ ] Instruction files reviewed for jose-ja/sm-im references

### Deployment

- [ ] `go run ./cmd/cicd-lint lint-deployments` passes
- [ ] deployments/sm/ no longer includes sm-im
- [ ] deployments/cryptoutil/ no longer includes jose

---

## Evidence Archive

- `test-output/tokens/` - token usage logs for this agent invocation
- `test-output/coverage-v24/` - coverage profiles generated during Phase 5
