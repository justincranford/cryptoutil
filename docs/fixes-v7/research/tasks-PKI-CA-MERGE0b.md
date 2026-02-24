# Tasks: PKI-CA-MERGE0b

**Option**: Merge cipher-im into sm-kms (combined KMS + messaging service)
**Status**: Research Only
**Created**: 2026-02-23
**Recommendation**: ⭐⭐ (Not recommended — see plan-PKI-CA-MERGE0b.md)

---

## Phase Pre: sm-kms Migration Debt (BLOCKING)

### Task Pre.1: Remove application_core / application_basic wrappers
- **Estimated**: 2h
- **Description**: Replace pre-builder patterns with direct ServiceResources usage
- **Same as**: PKI-CA-MIGRATE Task A.1

### Task Pre.2: Consolidate 15 custom middleware files → template session
- **Estimated**: 4h
- **Description**: Remove duplicate JWT/claims/session middleware, use template session.go
- **Same as**: PKI-CA-MIGRATE Task A.2

### Task Pre.3: Migrate SQLRepository to GORM via template
- **Estimated**: 6h
- **Description**: server.go:35 TODO — replace custom SQL with GORM from ServiceResources
- **Same as**: PKI-CA-MIGRATE Task A.3

### Task Pre.4: Add sm-kms integration tests
- **Estimated**: 4h
- **Same as**: PKI-CA-MIGRATE Task A.4

### Task Pre.5: Add sm-kms E2E tests
- **Estimated**: 3h
- **Same as**: PKI-CA-MIGRATE Task A.5

---

## Phase 1: Merge Domain Models

### Task 1.1: Add message migrations to sm-kms
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Pre.3
- **Description**: Create `server/repository/orm/migrations/2002_messages.up.sql` and `.down.sql` with messages and messages_recipient_jwks tables.
- **Acceptance Criteria**:
  - [ ] Migration adds both tables with correct indexes and foreign keys
  - [ ] Migration rollback tested

### Task 1.2: Add message domain models to sm-kms ORM entities
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: 1.1
- **Description**: Add `Message` and `MessageRecipientJWK` structs to `server/repository/orm/business_entities.go` (or new file `message_entities.go`).
- **Acceptance Criteria**:
  - [ ] `Message` and `MessageRecipientJWK` with GORM tags compatible with SQLite + PostgreSQL
  - [ ] Unit tests for model validation

---

## Phase 2: Merge Repositories

### Task 2.1: Port cipher-im repositories to sm-kms
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: 1.2
- **Description**: Port `message_repository.go`, `message_recipient_jwk_repository.go`, `user_repository.go`, `user_repository_adapter.go` from cipher-im into sm-kms `server/repository/`. Adapt to sm-kms ORM patterns.
- **Acceptance Criteria**:
  - [ ] All repository files ported with correct imports
  - [ ] `go test ./internal/apps/sm/kms/server/repository/...` passes

---

## Phase 3: Merge API and Business Logic

### Task 3.1: Merge OpenAPI spec
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Pre.3
- **Description**: Add message routes from cipher-im openapi spec into sm-kms openapi spec. Regenerate oapi-codegen stubs.
- **Acceptance Criteria**:
  - [ ] Merged spec validates with oapi-codegen
  - [ ] `go generate ./api/kms/...` produces updated server stubs

### Task 3.2: Port API handlers
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: 3.1
- **Description**: Port `server/apis/messages.go` and `server/apis/sessions.go` from cipher-im. Adapt to sm-kms handler patterns.

### Task 3.3: Port business logic (message service)
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: 3.2
- **Description**: Move cipher-im message business logic into sm-kms `server/businesslogic/businesslogic_messages.go`.

### Task 3.4: Wire new routes into sm-kms server
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: 3.3

---

## Phase 4: Testing and Archive

### Task 4.1: Unit tests for merged message handlers
- **Status**: ❌
- **Estimated**: 2h

### Task 4.2: Integration tests for combined service
- **Status**: ❌
- **Estimated**: 2h

### Task 4.3: Archive cipher-im
- **Status**: ❌
- **Estimated**: 30min
- **Description**: Move `internal/apps/cipher/im/` to `internal/apps/archived/cipher-im/`. Update cmd/cipher-im/main.go to import from archived path or remove.

### Task 4.4: Update ARCHITECTURE.md and CI
- **Status**: ❌
- **Estimated**: 1h

---

## Summary Stats

| Phase | Tasks | Est Effort |
|-------|-------|-----------|
| Pre: sm-kms migration debt | 5 | 19h |
| 1: Domain models | 2 | 2h |
| 2: Repositories | 1 | 2h |
| 3: API + business logic | 4 | 7h |
| 4: Testing + archive | 4 | 5.5h |
| **Total** | **16 tasks** | **~35.5h** |

**Compare with MERGE0a**: 4.5h for same product grouping goal.
**MERGE0b is 8× the effort of MERGE0a with worse maintainability.**
**Recommendation**: Choose MERGE0a.
