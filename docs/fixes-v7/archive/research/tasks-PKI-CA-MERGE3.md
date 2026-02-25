# Tasks: PKI-CA-MERGE3

**Option**: Archive sm-im + jose-ja + pki-ca; absorb all into sm-kms as crypto monolith
**Status**: Research Only
**Created**: 2026-02-23
**Recommendation**: ⭐ (Strongly not recommended)

---

## Phase Pre: Prerequisites (BLOCKING — all of MERGE2 plus more)

### Task Pre.0: All MERGE2 prerequisites
- **Estimated**: 32h
- **Description**: Everything in tasks-PKI-CA-MERGE2.md Phase Pre (jose-ja TODOs, sm-kms debt, template helper)
- **Reference**: tasks-PKI-CA-MERGE2.md

### Task Pre.1: Fix sm-im E2E reliability
- **Estimated**: 1.5h
- **Description**: Eliminate intermittent timeouts in sm-im E2E tests

### Task Pre.2: Migrate sm-im TestMain to template WaitForServerPort
- **Estimated**: 1h
- **Description**: Replace raw 50×100ms polling loop in sm-im testmain

---

## Phase 1-5: All MERGE2 phases

Same as tasks-PKI-CA-MERGE2.md Phases 1-5.
- **Estimated**: ~39h

---

## Phase 6: Port sm-im into sm-kms

### Task 6.1: Design merged OpenAPI spec (add cipher section)
- **Status**: ❌
- **Estimated**: 2h
- **Description**: Extend MERGE2 merged spec to include /service/api/v1/messages from sm-im

### Task 6.2: Port sm-im api/handler/ → sm-kms/server/a../sm/
- **Status**: ❌
- **Estimated**: 2h

### Task 6.3: Port sm-im server/service/ → sm-kms/server/servi../sm/
- **Status**: ❌
- **Estimated**: 2h

### Task 6.4: Create message_repository.go in sm-kms from sm-im repository
- **Status**: ❌
- **Estimated**: 2h

### Task 6.5: Wire sm-im services into sm-kms server builder
- **Status**: ❌
- **Estimated**: 1h

### Task 6.6: Unit/integration/E2E tests for merged cipher operations
- **Status**: ❌
- **Estimated**: 4h

### Task 6.7: Archive sm-im
- **Status**: ❌
- **Estimated**: 30min

---

## Summary Stats

| Phase | Tasks | Est Effort |
|-------|-------|-----------|
| Pre: all prerequisites | 3 | ~34.5h |
| MERGE2 phases 1-5 | (18 tasks) | ~39h |
| 6: Port sm-im | 7 | ~13.5h |
| **Total** | **~28 tasks** | **~87h** |

Highest effort, highest risk, worst architectural outcome of all 4 options.
