# Tasks - Complete KMS Migration (V8)

**Status**: 0 of 12 tasks complete (0%)
**Last Updated**: 2026-02-02
**Purpose**: Complete the ACTUAL remaining work from V7

---

## Phase 1: Barrier Integration

### Task 1.1: Update KMS businesslogic.go to Use Template Barrier
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: None
- **Description**: Replace shared/barrier import with template barrier in businesslogic.go
- **Acceptance Criteria**:
  - [ ] Import changed from `cryptoutil/internal/shared/barrier` to template barrier
  - [ ] Use orm_barrier_adapter.go to bridge interfaces
  - [ ] All businesslogic tests pass
- **Files**:
  - `internal/kms/server/businesslogic/businesslogic.go`
  - `internal/kms/server/businesslogic/businesslogic_test.go`

### Task 1.2: Update KMS application_core.go to Use Template Barrier
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Task 1.1
- **Description**: Initialize template barrier instead of shared/barrier in application_core.go
- **Acceptance Criteria**:
  - [ ] Import changed to template barrier
  - [ ] BarrierService initialized via template
  - [ ] KMS builds successfully
- **Files**:
  - `internal/kms/server/application/application_core.go`

### Task 1.3: Update KMS application_basic.go to Use Template UnsealKeysService
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 1.2
- **Description**: Replace shared/barrier/unsealkeysservice with template equivalent
- **Acceptance Criteria**:
  - [ ] Import changed to template unsealkeysservice
  - [ ] Unseal workflow works
- **Files**:
  - `internal/kms/server/application/application_basic.go`

### Task 1.4: Remove TODO Comments from server.go
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Tasks 1.1-1.3
- **Description**: Remove the 3 TODO(Phase2-5) comments after migration complete
- **Acceptance Criteria**:
  - [ ] All 3 TODOs removed
  - [ ] Migration actually complete (not just comments removed)
- **Files**:
  - `internal/kms/server/server.go`

### Task 1.5: Verify Zero shared/barrier Imports
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Tasks 1.1-1.4
- **Description**: Verify no shared/barrier imports remain in KMS
- **Acceptance Criteria**:
  - [ ] `grep -r "shared/barrier" internal/kms/` returns empty
  - [ ] Build passes
  - [ ] All tests pass
- **Evidence**: grep command output

---

## Phase 2: Testing & Verification

### Task 2.1: KMS Unit Tests Pass
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 1 complete
- **Description**: Run all KMS unit tests
- **Acceptance Criteria**:
  - [ ] `go test ./internal/kms/... -count=1` passes
  - [ ] No test regressions
- **Evidence**: test output

### Task 2.2: cipher-im Regression Tests Pass
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 1 complete
- **Description**: Verify cipher-im not broken by any shared changes
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/cipher/... -count=1` passes
- **Evidence**: test output

### Task 2.3: jose-ja Regression Tests Pass
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 1 complete
- **Description**: Verify jose-ja not broken by any shared changes
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/jose/... -count=1` passes
- **Evidence**: test output

### Task 2.4: Full Build Verification
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Tasks 2.1-2.3
- **Description**: Verify full project builds
- **Acceptance Criteria**:
  - [ ] `go build ./...` passes
  - [ ] No errors or warnings
- **Evidence**: build output

---

## Phase 3: Documentation & Cleanup

### Task 3.1: Update server-builder.instructions.md
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 2 complete
- **Description**: Update ServerBuilder documentation
- **Acceptance Criteria**:
  - [ ] Remove outdated V6 optional mode references
  - [ ] Document unified architecture
- **Files**:
  - `.github/instructions/03-08.server-builder.instructions.md`

### Task 3.2: Remove V7 builder_adapter.go if Still Exists
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Phase 2 complete
- **Description**: Clean up V7 Task 1.5 (remove builder_adapter.go)
- **Acceptance Criteria**:
  - [ ] builder_adapter.go deleted (if exists)
  - [ ] Build still passes
- **Files**:
  - `internal/kms/server/builder_adapter.go` (DELETE if exists)

### Task 3.3: Archive V7 Documentation
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 3.2
- **Description**: Move V7 docs to archive
- **Acceptance Criteria**:
  - [ ] V7 files moved to `docs/fixes-needed-plan-tasks-v8/archive/v7/`
  - [ ] Clean documentation structure

---

## Summary Statistics

| Phase | Tasks | Completed | Percentage |
|-------|-------|-----------|------------|
| Phase 1: Barrier Integration | 5 | 0 | 0% |
| Phase 2: Testing | 4 | 0 | 0% |
| Phase 3: Documentation | 3 | 0 | 0% |
| **Total** | **12** | **0** | **0%** |

---

## V7 Tasks Carried Over (Not Actually Done)

| V7 Task | V8 Task | Reason |
|---------|---------|--------|
| 5.3: Integrate Template Barrier with KMS | 1.1-1.4 | Marked "Not Started" in V7 |
| 5.4: Remove shared/barrier Usage from KMS | 1.5 | Marked "Not Started" in V7 |
| 1.5: Remove KMS builder_adapter.go | 3.2 | Marked "Not Started" in V7 |
| 1.6: Update ServerBuilder Documentation | 3.1 | Marked "Not Started" in V7 |
| 6.1-6.7: All Phase 6 Testing | 2.1-2.4 | All marked "Not Started" in V7 |
| 7.1-7.4: All Phase 7 Documentation | 3.1-3.3 | All marked "Not Started" in V7 |
