# Tasks: PKI-CA-MERGE2

**Option**: Archive jose-ja + pki-ca; absorb both into sm-kms as "Crypto Operations Service"
**Status**: Research Only
**Created**: 2026-02-23
**Recommendation**: ⭐⭐ (Not recommended — production boundary violation)

---

## Phase Pre: Prerequisites (BLOCKING — same as MIGRATE/MERGE1)

### Task Pre.1: Fix jose-ja critical TODOs
- **Estimated**: 11h
- **Description**: Implement jwk_handler.go:358,368 and jwk_handler_material.go:234,244,254,264
- **Same as**: PKI-CA-MIGRATE Task B.1 + B.2

### Task Pre.2: Fix sm-kms migration debt
- **Estimated**: 19h
- **Description**: Remove application_core wrappers, 15 middleware files → template session, GORM storage, add tests
- **Same as**: PKI-CA-MIGRATE Phase A

### Task Pre.3: Extract template generic startup helper
- **Estimated**: 2h
- **Description**: server_start_helpers.go with StartServiceFromConfig()
- **Same as**: Task 6.0 in main fixes-v7

---

## Phase 1: API Merger Preparation

### Task 1.1: Design merged OpenAPI spec
- **Status**: ❌
- **Estimated**: 4h
- **Description**: Merge sm-kms + jose-ja + pki-ca OpenAPI specs into single unified spec. Three top-level path groups: /keys, /jwks, /certs. Ensure no naming conflicts.
- **Acceptance Criteria**:
  - [ ] Merged spec validates with oapi-codegen
  - [ ] No path conflicts between 3 API groups
  - [ ] Spec reviewed for RESTful consistency

### Task 1.2: Update oapi-codegen configs for merged spec
- **Status**: ❌
- **Estimated**: 2h
- **Description**: Update openapi-gen_config_server.yaml, model.yaml, client.yaml for sm-kms to reference merged spec.

---

## Phase 2: Port jose-ja into sm-kms

### Task 2.1: Port jose-ja api/handler/ → sm-kms/server/api/jwk/
- **Status**: ❌
- **Estimated**: 3h
- **Dependencies**: 1.1, Pre.1

### Task 2.2: Port jose-ja service/ → sm-kms/server/service/jwk/
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: 2.1

### Task 2.3: Port jose-ja openapi/ → merged spec (already done in 1.1)
- **Status**: ❌
- **Estimated**: 1h (verification only)

---

## Phase 3: Port pki-ca into sm-kms

### Task 3.1: Create GORM cert_repository in sm-kms
- **Status**: ❌
- **Estimated**: 4h
- **Dependencies**: Pre.2

### Task 3.2: Port pki-ca service/ → sm-kms/server/service/ca/
- **Status**: ❌
- **Estimated**: 4h
- **Dependencies**: 3.1

### Task 3.3: Port pki-ca api/handler/ → sm-kms/server/api/ca/
- **Status**: ❌
- **Estimated**: 4h
- **Dependencies**: 3.2

### Task 3.4: Port pki-ca compliance/, crypto/, profile/, security/
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: 3.3

---

## Phase 4: Wire and Archive

### Task 4.1: Wire jose-ja and pki-ca services into sm-kms server builder
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: 2.2, 3.3

### Task 4.2: Update CLI routing for merged service
- **Status**: ❌
- **Estimated**: 1h

### Task 4.3: Archive jose-ja and pki-ca
- **Status**: ❌
- **Estimated**: 1h

---

## Phase 5: Testing

### Task 5.1: Unit tests for merged handlers (JWK + CA)
- **Status**: ❌
- **Estimated**: 3h

### Task 5.2: Integration tests for merged service
- **Status**: ❌
- **Estimated**: 3h

### Task 5.3: E2E test for merged service (all 3 API groups)
- **Status**: ❌
- **Estimated**: 3h

---

## Summary Stats

| Phase | Tasks | Est Effort |
|-------|-------|-----------|
| Pre: prerequisites | 3 | 32h |
| 1: API merger prep | 2 | 6h |
| 2: Port jose-ja | 3 | 6h |
| 3: Port pki-ca | 4 | 14h |
| 4: Wire + archive | 3 | 4h |
| 5: Testing | 3 | 9h |
| **Total** | **18 tasks** | **~71h** |

Note: Highest effort of all 4 options due to OpenAPI merger complexity and prerequisite debt.
