# Implementation Plan — Framework v24: 10-to-8 PS-ID Consolidation

**Status**: Completed implementation; post-implementation validation and hardening in progress
**Created**: 2026-05-25
**Last Updated**: 2026-05-30
**Purpose**: Consolidate from 10 PS-IDs (5 products) to 8 PS-IDs (4 products) by merging
jose-ja APIs into sm-kms and sm-im APIs into sm-kms, then deleting jose-ja, sm-im, and the jose product.

---

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions:**
- ✅ **Fix issues immediately** — blockers must be resolved before advancing to the next phase
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task or step complete with known issues

---

## Overview

Consolidate the Cryptoutil suite from:
- **Before**: 5 products (sm, jose, pki, identity, skeleton), 10 PS-IDs
- **After**: 4 products (sm, pki, identity, skeleton), 8 PS-IDs

**Removed PS-IDs**: `jose-ja` (merged into `sm-kms`), `sm-im` (merged into `sm-kms`)
**Removed Product**: `jose` (no remaining PS-IDs after removing jose-ja and sm-im)

The consolidation preserves 100% of existing API surface. Every endpoint that existed in
jose-ja or sm-im becomes a new endpoint group in sm-kms. No API functionality is deleted.

---

## Background

### Current Service Inventory

| PS-ID | Product | Port | Migration Range | Kept/Merged |
|-------|---------|------|-----------------|-------------|
| sm-kms | sm | 8000 | 2001–2999 | ✅ Kept (receives merges) |
| sm-im | sm | 8100 | 3001–3999 | ❌ Merged → sm-kms |
| jose-ja | jose | 8200 | 4001–4999 | ❌ Merged → sm-kms |
| pki-ca | pki | 8300 | 5001–5999 | ✅ Kept |
| identity-authz | identity | 8400 | 6001–6999 | ✅ Kept |
| identity-idp | identity | 8500 | 7001–7999 | ✅ Kept |
| identity-rs | identity | 8600 | 8001–8999 | ✅ Kept |
| identity-rp | identity | 8700 | 9001–9999 | ✅ Kept |
| identity-spa | identity | 8800 | 10001–10999 | ✅ Kept |
| skeleton-template | skeleton | 8900 | 11001–11999 | ✅ Kept |

### What jose-ja Provides

sm-kms operates as a JWK Authority: JOSE-format key management (JWK containers keyed by `kid`,
`kty`, `alg`, `use`), JWKS endpoint for public key distribution, JWS sign/verify, JWE
encrypt/decrypt, JWT create/verify, material key rotation, and audit logging.

**jose-ja API endpoints unique vs sm-kms** (must be added to sm-kms):
- `GET /elastic-keys/{elasticKeyID}/material-keys/active` — get currently active material key
- `POST /elastic-keys/{elasticKeyID}/rotate` — rotate (retire active, generate new active) material key
- `GET /jwks` — public JWKS endpoint for key distribution

**jose-ja domain objects** (must be ported as new sm-kms DB tables):
- `elastic_jwks` — JOSE-format elastic key containers (kid, kty, alg, use)
- `material_jwks` — versioned JOSE material keys (JWK-encrypted, active flag, barrier version)
- `tenant_audit_config` — per-tenant operation audit settings
- `audit_log` — operation audit log entries

**jose-ja services** (must be ported to sm-kms):
- `ElasticJWKService` (CRUD for JWK containers)
- `MaterialRotationService` (rotate/manage material key versions)
- `JWKSService` (generate JWKS from active public keys)
- `JWSService` (sign/verify with material keys)
- `JWEService` (encrypt/decrypt with material keys)
- `JWTService` (create/verify JWTs)
- `AuditLogService` (record operation audit events)

### What sm-im Provides

sm-kms is an Encrypted Instant Messenger: end-to-end encrypted messaging with JWE key wrapping
per recipient. Messages are stored as JWE multi-recipient ciphertext with per-recipient key material.

**sm-im API endpoints** (all new for sm-kms):
- `GET /messages` — list messages (paginated)
- `GET /messages/{messageID}` — get a message
- `DELETE /messages/{messageID}` — delete a message (sender only)
- `POST /messages/send` — send encrypted message to one or more recipients
- `GET /messages/receive` — receive (list unread) messages for the caller

**sm-im domain objects** (must be ported as new sm-kms DB tables):
- `messages` — encrypted message records (JWE JSON, sender, created_at, read_at)
- `messages_recipient_jwks` — per-recipient encrypted JWK for message decryption

**sm-im services** (must be ported to sm-kms):
- `MessageHandler` (send, receive, get, delete, list messages)

---

## Technical Context

- **Language**: Go 1.26.1
- **Framework**: `internal/apps-framework/service/` service builder
- **Database**: PostgreSQL (E2E) + SQLite in-memory (unit/integration)
- **Test references**: [ENG-HANDBOOK §10](../../docs/ENG-HANDBOOK.md#10-testing-architecture), [§10.2](../../docs/ENG-HANDBOOK.md#102-unit-testing-strategy), [§10.3](../../docs/ENG-HANDBOOK.md#103-integration-testing-strategy), [§10.4](../../docs/ENG-HANDBOOK.md#104-e2e-testing-strategy), [§10.5](../../docs/ENG-HANDBOOK.md#105-mutation-testing-strategy)
- **Quality reference**: [ENG-HANDBOOK §11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates), [§11.3](../../docs/ENG-HANDBOOK.md#113-code-quality-standards)
- **Coding standards**: [ENG-HANDBOOK §14.1](../../docs/ENG-HANDBOOK.md#141-coding-standards)

### Affected Files — Complete Enumeration

#### Phase 1: jose-ja → sm-kms API merge

**New files (sm-kms receives sm-kms domain)**:
```
internal/apps/sm-kms/server/repository/migrations/
  2003_jwk_elastic_jwks.up.sql
  2003_jwk_elastic_jwks.down.sql
  2004_jwk_material_jwks.up.sql
  2004_jwk_material_jwks.down.sql
  2005_jwk_audit_config.up.sql
  2005_jwk_audit_config.down.sql
  2006_jwk_audit_log.up.sql
  2006_jwk_audit_log.down.sql
internal/apps/sm-kms/server/model/
  jwk_models.go                    (ElasticJWK, MaterialJWK, AuditConfig, AuditLogEntry)
  jwk_models_test.go
internal/apps/sm-kms/server/repository/
  elastic_jwk_repository.go
  elastic_jwk_repository_test.go
  material_jwk_repository.go
  material_jwk_repository_test.go
  audit_repository.go
  audit_repository_test.go
internal/apps/sm-kms/server/jwkservice/
  elastic_jwk_service.go
  elastic_jwk_service_test.go
  material_rotation_service.go
  material_rotation_service_test.go
  jwks_service.go
  jwks_service_test.go
  jws_service.go
  jws_service_test.go
  jwe_service.go
  jwe_service_test.go
  jwt_service.go
  jwt_service_test.go
  audit_log_service.go
  audit_log_service_test.go
  testmain_test.go
internal/apps/sm-kms/server/handler/
  jwk_handler.go                   (elastic JWK CRUD + rotate + active material key)
  jwk_handler_test.go
  jwks_handler.go                  (GET /jwks endpoint)
  jwks_handler_test.go
```

**Updated files (sm-kms API spec extended)**:
```
api/sm-kms/openapi_spec_paths.yaml         (+3 new paths: active material key, rotate, JWKS)
api/sm-kms/openapi_spec_components.yaml    (+JWK schemas: ElasticJWK, MaterialJWK, JWKS, audit)
api/sm-kms/server/server.gen.go            (regenerated via oapi-codegen)
api/sm-kms/models/models.gen.go            (regenerated)
api/sm-kms/client/client.gen.go            (regenerated)
internal/apps/sm-kms/server/server.go      (register JWK routes)
internal/apps/sm-kms/server/repository/migrations.go  (embed new migration files)
internal/apps/sm-kms/kms.go               (update usage docs)
```

#### Phase 2: sm-im → sm-kms API merge

**New files**:
```
internal/apps/sm-kms/server/repository/migrations/
  2007_im_messages.up.sql
  2007_im_messages.down.sql
  2008_im_recipient_jwks.up.sql
  2008_im_recipient_jwks.down.sql
internal/apps/sm-kms/server/model/
  message_models.go                (Message, MessageRecipientJWK)
  message_models_test.go
internal/apps/sm-kms/server/repository/
  message_repository.go
  message_repository_test.go
  message_recipient_jwk_repository.go
  message_recipient_jwk_repository_test.go
internal/apps/sm-kms/server/handler/
  message_handler.go               (send, receive, get, delete, list messages)
  message_handler_test.go
```

**Updated files**:
```
api/sm-kms/openapi_spec_paths.yaml         (+5 message endpoints)
api/sm-kms/openapi_spec_components.yaml    (+Message schemas)
api/sm-kms/server/server.gen.go            (regenerated)
api/sm-kms/models/models.gen.go            (regenerated)
api/sm-kms/client/client.gen.go            (regenerated)
internal/apps/sm-kms/server/server.go      (register message routes)
internal/apps/sm-kms/server/repository/migrations.go  (embed new migrations)
```

#### Phase 3: Delete jose-ja and sm-im

**Deleted (entire directories)**:
```
api/sm-kms/                       (7 files)
api/sm-kms/                         (7 files)
internal/apps/sm-kms/             (75 files)
internal/apps/sm-kms/               (60 files)
cmd/sm-kms/                       (1 file: main.go)
cmd/jose/                          (1 file: main.go)
cmd/sm-kms/                         (1 file: main.go)
configs/sm-kms/                   (~5 files)
configs/sm-kms/                     (~5 files)
deployments/sm-kms/               (~10 files)
deployments/jose/                  (~3 files)
deployments/sm-kms/                 (~10 files)
internal/shared/magic/magic_jose.go
internal/shared/magic/magic_sm_im.go
```

**Updated (references removed)**:
```
deployments/sm/compose.yml                  (remove sm-kms service)
deployments/cryptoutil/compose.yml          (remove jose service block)
cmd/sm/main.go                              (remove sm-kms subcommand routing)
cmd/cryptoutil/main.go                      (remove jose routing)
internal/apps/sm/sm.go                      (remove im references)
internal/apps/jose/ja.go → DELETE entirely
```

**Count**: 170 files deleted + 8 files updated in this phase

#### Phase 4: Registry, magic constants, fitness linters

```
api/cryptosuite-registry/registry.yaml     (remove jose-ja, sm-im entries; remove jose product)
internal/shared/magic/magic_tier.go        (remove JoseProductName tier, remove SMIM from SM tier)
internal/shared/magic/magic_cicd.go        (update service counts: 10→8, 5→4 products)
internal/shared/magic/magic_pki_tls.go     (remove AppJoseJA* TLS cert constants)
internal/shared/magic/magic_sm.go          (remove IM*-prefixed constants if any)
internal/apps-tools/cicd_lint/lint_ports/host_port_ranges/*.go  (remove sm-kms port range 8200s)
internal/apps-tools/cicd_lint/lint_ports/legacy_ports/*.go      (remove jose-ja/sm-im legacy ports)
internal/apps-tools/cicd_lint/lint_fitness/registry/registry.go (remove jose-ja, sm-im, jose product)
```

**Count**: ~8 files updated

**Derivation**: 6 SQL migration files + 30 new Go files + 10 updated Go files + 7 API spec files +
170 deleted files = ~220 files total touched across all phases.

---

## Phases

**Phase Status Legend**: `☐ TODO` | `🔄 IN PROGRESS` | `✅ COMPLETE` | `⏳ BLOCKED`

---

### Phase 1: jose-ja Domain → sm-kms (4d) [Status: ☐ TODO]

**Objective**: Port all sm-kms domain models, repositories, services, and API handlers into
sm-kms without deleting sm-kms yet. jose-ja and sm-im both run; sm-kms is now redundant.

1. Add DB migrations 2003–2006 to sm-kms (elastic_jwks, material_jwks, audit_config, audit_log)
2. Port sm-kms domain models to `internal/apps/sm-kms/server/model/jwk_models.go`
3. Port sm-kms repositories to `internal/apps/sm-kms/server/repository/` (elastic_jwk_repo, material_jwk_repo, audit_repo)
4. Port sm-kms services to `internal/apps/sm-kms/server/jwkservice/` (elastic, rotation, JWKS, JWS, JWE, JWT, audit)
5. Add JWK handlers to `internal/apps/sm-kms/server/handler/` (JWK CRUD + rotate + active + JWKS endpoint)
6. Extend sm-kms OpenAPI spec (paths + components) with sm-kms API surface
7. Regenerate oapi-codegen outputs for sm-kms
8. Register JWK routes in sm-kms `server.go`
9. Update sm-kms repository `migrations.go` to embed new files
10. Port sm-kms tests (adapted to sm-kms package structure)

- **Success**: `go build ./...` clean; all sm-kms tests pass; sm-kms tests still pass (both
  services work simultaneously); sm-kms API now includes JWK endpoints; coverage ≥95%
- **Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 2: sm-im Domain → sm-kms (2d) [Status: ☐ TODO]

**Objective**: Port all sm-kms domain models, repository, and handler into sm-kms. sm-kms can
now send/receive encrypted messages.

1. Add DB migrations 2007–2008 to sm-kms (messages, messages_recipient_jwks)
2. Port sm-kms domain models to `internal/apps/sm-kms/server/model/message_models.go`
3. Port sm-kms repositories to `internal/apps/sm-kms/server/repository/` (message_repo, message_recipient_jwk_repo)
4. Port sm-kms message handler to `internal/apps/sm-kms/server/handler/message_handler.go`
5. Extend sm-kms OpenAPI spec with messaging endpoints (send, receive, get, list, delete)
6. Regenerate oapi-codegen outputs for sm-kms
7. Register message routes in sm-kms `server.go`
8. Port sm-kms tests (adapted to sm-kms package structure)

- **Success**: `go build ./...` clean; all sm-kms tests pass (including messaging); sm-kms tests
  still pass; sm-kms API includes messaging endpoints; coverage ≥95%
- **Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 3: Delete jose-ja, sm-im, jose Product (1d) [Status: ☐ TODO]

**Objective**: Remove all jose-ja and sm-im artifacts from the codebase. The jose product
disappears because it has no remaining PS-IDs.

1. Delete `api/sm-kms/` entirely
2. Delete `api/sm-kms/` entirely
3. Delete `internal/apps/sm-kms/` entirely
4. Delete `internal/apps/sm-kms/` entirely
5. Delete `cmd/sm-kms/`, `cmd/jose/`, `cmd/sm-kms/`
6. Delete `configs/sm-kms/`, `configs/sm-kms/`
7. Delete `deployments/sm-kms/`, `deployments/jose/`, `deployments/sm-kms/`
8. Delete `internal/shared/magic/magic_jose.go`, `magic_sm_im.go`
9. Update `deployments/sm/compose.yml` — remove sm-kms service
10. Update `deployments/cryptoutil/compose.yml` — remove jose service block
11. Update `cmd/sm/main.go` — remove sm-kms routing
12. Update `cmd/cryptoutil/main.go` — remove jose routing
13. Update `internal/apps/sm/sm.go` — remove im references
14. Delete `internal/apps/jose/` (product-level coordinator; jose product is gone)

- **Success**: `go build ./...` clean; no references to jose-ja or sm-im anywhere; all remaining
  8 PS-ID services build and test successfully
- **Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 4: Registry, Magic Constants, Fitness Linters (1d) [Status: ☐ TODO]

**Objective**: Update the canonical registry and all derived constants to reflect 4 products
and 8 PS-IDs. Fitness linters must pass with zero errors.

1. Update `api/cryptosuite-registry/registry.yaml`:
   - Remove `sm-kms` from `product_services`
   - Remove `sm-kms` from `product_services`
   - Remove `jose` from `products`
   - Renumber migration ranges if needed (sk-kms now absorbs 2003–2008 from the merged services)
2. Update `internal/shared/magic/magic_tier.go`:
   - Remove `JoseProductName` tier entry
   - Remove `OTLPServiceSMIM` from SM tier entry
   - Remove `OTLPServiceJoseJA` from all tier lists
3. Update `internal/shared/magic/magic_cicd.go`: update service count comments (10→8, 5→4)
4. Update `internal/shared/magic/magic_pki_tls.go`: remove `AppJoseJA*` constants
5. Update `internal/shared/magic/magic_sm.go`: remove any IM-specific SM constants
6. Update `internal/apps-tools/cicd_lint/lint_ports/host_port_ranges/`: remove sm-kms range (8200s)
7. Update `internal/apps-tools/cicd_lint/lint_ports/legacy_ports/`: remove jose-ja/sm-im entries
8. Verify `go run ./cmd/cicd-lint lint-fitness` passes (entity-registry-completeness checks)
9. Verify `go run ./cmd/cicd-lint lint-deployments` passes

- **Success**: `go run ./cmd/cicd-lint lint-fitness` passes; `lint-deployments` passes; no
  stale references in magic files; all tests pass
- **Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 5: Full Quality Gate Verification (0.5d) [Status: ☐ TODO]

**Objective**: Comprehensive end-to-end quality validation of the consolidated codebase.

1. `go build ./...` — zero errors
2. `go build -tags e2e,integration ./...` — zero errors
3. `golangci-lint run ./...` — zero warnings
4. `golangci-lint run --build-tags e2e,integration ./...` — zero warnings
5. `go test ./...` — 100% pass, zero skips
6. `go test -race -count=2 ./...` — race detector clean
7. Coverage check: production ≥95%, infrastructure/utility ≥98%
8. `go run ./cmd/cicd-lint lint-fitness` — zero violations
9. `go run ./cmd/cicd-lint lint-deployments` — zero violations
10. `go run ./cmd/cicd-lint lint-go lint-go-test lint-golangci` — zero violations

- **Success**: All 10 checks above pass with zero errors or warnings
- **Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 6: Knowledge Propagation (0.5d) [Status: ☐ TODO]

**Objective**: Apply lessons learned from consolidation to permanent artifacts.

1. Review `lessons.md` from all prior phases
2. Update `docs/ENG-HANDBOOK.md` — add consolidation decision record, update product/PS-ID
   counts in §3 (product suite architecture)
3. Update `api/cryptosuite-registry/registry.yaml` doc comments
4. Update `README.md` service table (8 services instead of 10)
5. Update `docs/DEV-SETUP.md` if any setup steps referenced jose-ja or sm-im
6. Verify propagation: `go run ./cmd/cicd-lint lint-docs` — zero violations
7. Commit each artifact type separately (ENG-HANDBOOK, README, DEV-SETUP as separate commits)

- **Success**: `go run ./cmd/cicd-lint lint-docs` passes; ENG-HANDBOOK updated with consolidation
  rationale; documentation no longer references removed services

---

## Executive Decisions

### Decision 1: Migration Number Strategy for Merged Domains

**Options**:
- A: Use sm-kms migration range (2003–2008) for all merged tables ✓ **SELECTED**
- B: Keep original migration numbers (3001, 4001) in sm-kms — conflicts with range policy
- C: Allocate a new migration sub-range (e.g., 2100+ for jose, 2200+ for im)
- D: Merge all into a single migration file
- E:

**Decision**: Option A — use sm-kms domain range (2001–2999) for all new tables added to sm-kms.
sm-kms already has 2001 and 2002; new migrations start at 2003.

**Rationale**: The migration range policy assigns ranges per PS-ID domain ownership. Once sm-kms
and sm-kms are merged into sm-kms, their tables become sm-kms domain tables. Using 2003–2008
maintains sequential numbering within the sm-kms range without gaps or conflicts.

**Alternatives Rejected**:
- Option B: Original numbers (3001, 4001) would leave gaps in sm-kms's migration sequence and
  break the range-per-PS-ID invariant once the source PS-IDs are deleted.
- Option C: Sub-ranges add conceptual complexity without benefit.
- Option D: Single migration file loses granularity and cannot be independently rolled back.

---

### Decision 2: sm-kms JWK Model vs sm-kms ElasticKey Model

**Options**:
- A: Add sm-kms's `elastic_jwks`/`material_jwks` tables alongside sm-kms's existing `elastic_keys`/`material_keys` ✓ **SELECTED**
- B: Unify into a single `elastic_keys` table by extending the schema with JWK-specific columns
- C: Replace sm-kms model with sm-kms's JWK model
- D: Wrap sm-kms JWK operations via sm-kms's existing elastic key APIs
- E:

**Decision**: Option A — keep both domain models as separate table families in sm-kms.

**Rationale**: The two models serve different purposes:
- `elastic_keys` / `material_keys` (sm-kms): generic KMS with status state machine, provider,
  versioning control, import control — used for raw symmetric/asymmetric key operations
- `elastic_jwks` / `material_jwks` (sm-kms): JOSE-format key management with kid, kty, alg, use
  — used for JWK distribution, signing, encryption, JWT operations

Merging into one schema requires nullable columns or discriminator fields, violating database
normalization and making the state machine more complex. Two clean, focused schemas are better.

**Impact**: sm-kms gains two new table families and two new API endpoint groups (JWK and messaging).

---

### Decision 3: Handler Package Organization in sm-kms

**Options**:
- A: Place new JWK and message handlers in existing `server/handler/` alongside current OAS handlers ✓ **SELECTED**
- B: Create `server/jwk_handler/` and `server/message_handler/` subdirectories
- C: Create a single `server/apis/` directory (mimicking jose-ja/sm-im structure)
- D:
- E:

**Decision**: Option A — extend the existing `server/handler/` package.

**Rationale**: sm-kms already has `server/handler/` as its handler package. Keeping new handlers
in the same package avoids fragmentation and aligns with the existing file organization pattern.
Each domain gets a separate file (`jwk_handler.go`, `jwks_handler.go`, `message_handler.go`).

---

### Decision 4: sm-kms Service Package Location in sm-kms

**Options**:
- A: Create `internal/apps/sm-kms/server/jwkservice/` package for JOSE business logic ✓ **SELECTED**
- B: Extend existing `internal/apps/sm-kms/server/businesslogic/` with JWK functions
- C: Create `internal/apps/sm-kms/server/service/` (matches sm-kms's original package name)
- D:
- E:

**Decision**: Option A — create a dedicated `jwkservice/` package.

**Rationale**: sm-kms's `businesslogic/` is tightly coupled to the KMS elastic key domain model
and has a specific state machine. Adding JOSE operations to it would mix concerns. A new
`jwkservice/` package cleanly encapsulates all JOSE service logic with no cross-contamination.
This also mirrors the `sm-kms/server/handler/` approach — handlers call into `businesslogic/`
for KMS operations and into `jwkservice/` for JWK operations.

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Migration number collision with existing sm-kms migrations | Low | High | Use 2003–2008; verify no existing files in that range before starting |
| sm-kms openapi spec grows very large (>500 lines) | Medium | Medium | Split into multiple path files; keep components.yaml as shared file |
| Deleted sm-kms import still referenced in shared packages | Medium | High | Run `grep -r sm-kms ./internal` after deletion before committing |
| Fitness linter tests reference jose-ja/sm-im by name | High | Medium | Phase 4 explicitly updates lint_ports and lint_fitness; run `lint-fitness` immediately after |
| sm-kms test coverage drops below 95% during merge | Medium | High | Port all existing tests; add coverage for new code paths before closing phase |
| Import aliases conflict between merged packages | Low | Medium | Use distinct aliases; follow existing `cryptoutilApps*` pattern |
| Race conditions in shared SQLite during large test suites | Low | Medium | Each test package uses isolated in-memory SQLite (standard pattern) |

---

## Quality Gates - MANDATORY

**Per-Phase Quality Gates** (from [ENG-HANDBOOK §11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates)):
- ✅ `go build ./...` clean
- ✅ `go build -tags e2e,integration ./...` clean
- ✅ `golangci-lint run ./...` + `golangci-lint run --build-tags e2e,integration ./...` — zero warnings
- ✅ `go test ./...` — 100% pass, zero skips
- ✅ Coverage: production ≥95%, infrastructure/utility ≥98%
- ✅ `go run ./cmd/cicd-lint lint-fitness` — zero violations
- ✅ `go run ./cmd/cicd-lint lint-deployments` — zero violations

**3-Tier Database Strategy** ([ENG-HANDBOOK §10](../../docs/ENG-HANDBOOK.md#10-testing-architecture)):
- Unit tests: SQLite in-memory only, NEVER PostgreSQL
- Integration tests: ONE shared SQLite in-memory instance per package via TestMain
- E2E tests: Docker Compose with PostgreSQL (only tier where PostgreSQL is used)

---

## Success Criteria

- [ ] All 6 phases complete with evidence
- [ ] Build clean: `go build ./...` zero errors
- [ ] All tests pass: `go test ./...` 100% passing, zero skips
- [ ] No jose-ja or sm-im references anywhere in non-documentation code
- [ ] sm-kms OpenAPI spec includes all jose-ja and sm-im endpoints
- [ ] Registry shows 4 products and 8 PS-IDs
- [ ] `go run ./cmd/cicd-lint lint-fitness` passes
- [ ] `go run ./cmd/cicd-lint lint-deployments` passes
- [ ] Documentation updated

---

## ENG-HANDBOOK.md Cross-References

| Topic | Section | When Applicable |
|-------|---------|-----------------|
| Testing Strategy | [§10](../../docs/ENG-HANDBOOK.md#10-testing-architecture) | All phases |
| 3-Tier DB Strategy | [§10.2](../../docs/ENG-HANDBOOK.md#102-unit-testing-strategy) | Phases 1, 2 |
| Integration Testing | [§10.3](../../docs/ENG-HANDBOOK.md#103-integration-testing-strategy) | Phases 1, 2 |
| Quality Gates | [§11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates) | All phases |
| Coding Standards | [§14.1](../../docs/ENG-HANDBOOK.md#141-coding-standards) | Phases 1, 2, 3 |
| Version Control / Commits | [§14.2](../../docs/ENG-HANDBOOK.md#142-version-control) | All phases |
| Deployment Architecture | [§12](../../docs/ENG-HANDBOOK.md#12-deployment-architecture) | Phase 3 |
| Service Template | [§5.1](../../docs/ENG-HANDBOOK.md#51-service-framework-pattern) | Phases 1, 2 |
| API Architecture | [§8](../../docs/ENG-HANDBOOK.md#8-api-architecture) | Phases 1, 2 |
| Migration Ranges | [§7 / registry.yaml](../../api/cryptosuite-registry/registry.yaml) | Phases 1, 2, 4 |
| Knowledge Propagation | [§14.8](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation) | Phase 6 |
