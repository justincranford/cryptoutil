# Implementation Plan - Complete KMS Migration (V8)

**Status**: Planning
**Created**: 2026-02-02
**Last Updated**: 2026-02-02
**Purpose**: Complete the ACTUAL KMS barrier migration that V7 did not finish

## Executive Summary

**V7 Post-Mortem**: V7 tasks.md shows "23 of 40 tasks complete (57.5%)" but code archaeology reveals:
- Tasks 5.3, 5.4 (barrier integration) are NOT done despite Task 5.1, 5.2 being "Complete"
- KMS server.go has 3 TODO comments explicitly stating migration is incomplete
- KMS still imports `shared/barrier` in 5 files (businesslogic.go, application_core.go, etc.)
- The orm_barrier_adapter.go was created but is NOT integrated into KMS production code

**V8 Goal**: Actually complete the barrier migration and remove shared/barrier usage from KMS.

## Technical Context

### ACTUAL Current State (Verified by Code)

| Service | Database | Auth | OpenAPI | Barrier | Migrations |
|---------|----------|------|---------|---------|------------|
| **cipher-im** | GORM ✅ | JWT/Realms ✅ | Strict ✅ | Template ✅ | Template+Domain ✅ |
| **jose-ja** | GORM ✅ | JWT/Realms ✅ | Strict ✅ | Template ✅ | Template+Domain ✅ |
| **sm-kms** | GORM (OrmRepository) ✅ | Basic HTTP ❌ | Strict ✅ | **shared/barrier** ❌ | Custom ❌ |

### Evidence of Incomplete V7

```bash
# KMS still uses shared/barrier
$ grep -r "shared/barrier" internal/kms/ --include="*.go"
internal/kms/server/businesslogic/businesslogic.go:     cryptoutilBarrierService "cryptoutil/internal/shared/barrier"
internal/kms/server/businesslogic/businesslogic_test.go:        cryptoutilBarrierService "cryptoutil/internal/shared/barrier"
internal/kms/server/application/application_basic.go:   cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
internal/kms/server/application/application_core.go:    cryptoutilBarrierService "cryptoutil/internal/shared/barrier"
internal/kms/server/server.go:// Currently KMS has its own SQLRepository and shared/barrier - these need to be replaced

# server.go explicitly acknowledges incomplete migration
$ grep "TODO" internal/kms/server/server.go
# TODO(Phase2-5): KMS needs to be migrated to use template's GORM database and barrier.
# TODO(Phase2-5): Replace with template's GORM database and barrier.
# TODO(Phase2-5): Switch to TemplateWithDomain mode once KMS uses template DB.
```

## V8 Scope (Focused on Actual Remaining Work)

### What V7 Actually Completed
- ✅ Phase 0: Research & Discovery (all 4 tasks)
- ✅ Phase 1: Removed V6 optional modes (Tasks 1.1-1.4, 1.7)
- ✅ Phase 2: KMS GORM models/migrations/repositories exist
- ✅ Phase 3: JWT middleware files created
- ✅ Phase 4: OpenAPI strict server (pre-existing)
- ✅ Phase 5 Tasks 5.1, 5.2: Analysis complete (barrier comparison docs)
- ✅ orm_barrier_adapter.go created (adapter infrastructure)

### What V7 Did NOT Complete
- ❌ Task 1.5: Remove KMS builder_adapter.go
- ❌ Task 1.6: Update ServerBuilder documentation
- ❌ Task 5.3: Integrate Template Barrier with KMS
- ❌ Task 5.4: Remove shared/barrier Usage from KMS
- ❌ All Phase 6 tasks (Integration & Testing)
- ❌ All Phase 7 tasks (Documentation & Cleanup)

## V8 Phases

### Phase 1: Barrier Integration (V7 Tasks 5.3, 5.4)

**Objective**: Actually integrate template barrier and remove shared/barrier

- Update KMS businesslogic.go to use template barrier service
- Update KMS application_core.go to initialize template barrier
- Remove shared/barrier imports from all KMS files
- Verify all KMS tests still pass

### Phase 2: Testing & Verification (V7 Phase 6)

**Objective**: Comprehensive testing after barrier migration

- All KMS unit tests pass
- cipher-im regression tests pass
- jose-ja regression tests pass
- E2E tests work

### Phase 3: Documentation & Cleanup (V7 Phase 7)

**Objective**: Update documentation to reflect actual architecture

- Update server-builder.instructions.md
- Remove obsolete V6 documentation
- Create service migration guide

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Barrier interface mismatch | Medium | High | orm_barrier_adapter.go already handles type conversion |
| Unseal key derivation differences | Low | High | Template uses same HKDF pattern |
| Tests break after migration | Medium | Medium | Run tests incrementally |

## Quality Gates

- ✅ All tests pass (`go test ./internal/kms/... -count=1`)
- ✅ No shared/barrier imports in KMS (`grep -r "shared/barrier" internal/kms/` returns empty)
- ✅ No TODO(Phase2-5) comments remaining
- ✅ Build clean (`go build ./...`)
- ✅ Linting clean (`golangci-lint run`)

## Success Criteria

- [ ] KMS uses ONLY template barrier (zero shared/barrier imports)
- [ ] All 3 TODOs in server.go resolved
- [ ] All KMS tests pass
- [ ] cipher-im and jose-ja not regressed
- [ ] Documentation accurate
