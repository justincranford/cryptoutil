# Post-Mortem: fixes-v7 Quality Improvement Plan

## Executive Summary

**Duration**: 8 phases, 53 tasks, 22 commits
**Scope**: 506 files changed, 32,973 insertions, 32,302 deletions
**Result**: 52/53 tasks complete, 1 blocked (pre-existing OTel Docker socket issue)
**Quality Gates**: All passing — build clean, zero lint issues, 194 test packages pass, 62/62 deployment validators

## Phase Summary

### Phase 1: Critical Bug Fixes (3 tasks)
- **Commit**: a7644d49
- **Work**: Fixed poll.go edge case, ValidateUUIDs error message, container config test copy-paste bug
- **Impact**: Prevented runtime panics and misleading error messages

### Phase 2: Code Quality & Standards (8 tasks)
- **Commit**: cbd5fec7
- **Work**: File rename for consistency, ValidateUUID signature fix, nolint removal, pool.go if→switch, test measurement fix
- **Impact**: Improved code consistency and removed legacy lint suppressions

### Phase 3: Magic Constant Consolidation (5 tasks)
- **Commit**: eeca6c27 (165 files)
- **Work**: Relocated identity, PKI CA, demo magic constants to shared/magic/
- **Impact**: Single location for all magic constants, consistent import patterns

### Phase 4: Test Infrastructure (6 tasks)
- **Commit**: 47d0c6d7
- **Work**: Added t.Parallel() to 35+ tests, replaced time.Sleep with polling, split 4 oversized test files, added zero-coverage package tests
- **Impact**: Faster test execution, no flaky timing-dependent tests

### Phase 5: Code Cleanup (5 tasks)
- **Commit**: 20b9fe1f
- **Work**: Telemetry relocation, poller refactoring, nolint cleanup, unused sentinel removal, SQL injection defense
- **Impact**: Better architecture boundaries, defense-in-depth for SQL operations

### Phase 6: E2E Infrastructure (6 tasks, 1 blocked)
- **Commits**: 7b810a5f, e807d470, 96786989, 227fe70f
- **Work**: Generic DualPortServer helpers, TestMain migration, E2E suites for jose-ja and sm-kms, KMS config fix, JOSE routing standardization
- **Blocked**: Task 6.5 — OTel collector Docker socket issue (pre-existing, not introduced by this plan)
- **Impact**: Standardized service startup patterns, reusable E2E infrastructure

### Phase 7: Coverage & Mutation (3 tasks)
- **Commit**: a3cb4ec6
- **Work**: JWX coverage ceiling analysis (~90%, documented in JWX-COV-CEILING.md), database coverage improvement (93.8→98.5%), mutation testing baseline
- **Impact**: Evidence-based understanding of coverage limits, improved database reliability

### Phase 8: sm-im → sm-im Rename (5 tasks)
- **Commits**: 36c17996 (130 files), 0614cab3 (63 files), 58444101 (4 files)
- **Work**: Complete product rename — Go code, deployments, configs, docs, CI/CD, Dockerfile
- **Impact**: Former Cipher product eliminated, sm-im properly nested under SM product (3 products: sm, pki, jose)

## Issues Discovered

### Pre-Existing Issues (Not Introduced by Plan)

1. **OTel Collector Docker Socket**: resourcedetection processor requires Docker socket access inside container. E2E test infrastructure works but OTel blocks the startup chain.
   - **Status**: BLOCKED — requires OTel config change or Docker socket mount
   - **Impact**: E2E tests cannot run end-to-end despite functional framework

### Issues Found and Fixed During Execution

1. **goconst lint violation**: Adding sm-im service check to sm product created duplicate "sm" string literal. Fixed by introducing `smProduct` local variable.
2. **Delegation test assertion**: Old test asserted `result.Errors[0]` but sm product now has two services, creating multiple error cases. Fixed with `strings.Contains` loop.
3. **Orphaned .db files**: 405 SQLite database files left in `internal/ap../sm/` after git mv. Removed with `rm -rf`.
4. **Dockerfile user references**: Former Cipher user/group names in Dockerfile survived the rename. Fixed with targeted sed.
5. **tasks.md checkbox gap**: Tasks 6.0-6.2 marked as ❌ despite work being committed. Root cause: tasks.md wasn't updated when work was done (commit 49d30db0 says "mark Tasks 6.1-6.2 complete" but didn't touch tasks.md).

## Lessons Learned

1. **Always update tasks.md in the same commit as the work**: The gap between commit 49d30db0 (did work) and tasks.md (still ❌) caused confusion. Ensure task status updates are atomic with the work.

2. **Large renames benefit from phased commits**: The cipher→sm-im rename across 200+ files was cleanly split into code (130 files), deployment (63 files), and docs (4 files) commits. Each could be independently reviewed and bisected.

3. **Deployment validator counts change with structural changes**: Removing the SM product reduced validators from 65 to 62. Acceptance criteria should use dynamic counts or describe the expected change.

4. **Template standardization enables service migration**: Once JOSE-JA, sm-kms, and sm-im all use RouteService/RouteProduct/ParseWithFlagSet, adding new services or renaming existing ones is mechanical.

5. **Pre-existing infrastructure issues should be tracked separately**: The OTel Docker socket issue blocked E2E validation across multiple plans (v1, v6, v7). It deserves its own tracking item outside of feature plans.

## Final Metrics

| Metric | Before | After |
|--------|--------|-------|
| Products | 4 (SM, PKI, JOSE) | 3 (SM, PKI, JOSE) |
| SM services | 1 (kms) | 2 (kms, im) |
| Deployment validators | 65 | 62 |
| Test packages passing | 194 | 194 |
| Lint issues | 0 | 0 |
| Legacy nolint:wsl | 2 | 0 |
| Files >500 lines (non-gen) | 4 | 0 |
| Magic constants in shared/ | ~60% | 100% |
| Database coverage | 93.8% | 98.5% |
