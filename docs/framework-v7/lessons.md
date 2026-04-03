# Lessons — Parameterization Opportunities

**Created**: 2026-03-29
**Last Updated**: 2026-04-02
**Status**: Phase 0 continuation in progress.

---

## Phase 0: Pre-Work Defect Fixes

### Task 0.10 — Hard Error on Absent Dirs (All Fitness Linters)

**Lesson**: `os.IsNotExist → return nil` (silent skip) in fitness linters is categorically wrong. When a required directory is absent it means the workspace is non-compliant — the linter MUST return a hard error so CI/CD fails visibly. All 71 fitness linter CheckInDir functions must return `fmt.Errorf(...)` when a required directory is not found. This is the project-wide standard documented in ARCHITECTURE.md §9.11.2.

**Unit test pattern — stub ALL required dirs**: Tests for dir-iterating fitness linters must create stubs for every directory the linter checks: all PS-ID config dirs, all PS-ID domain dirs, and any shared infrastructure dirs (e.g. the framework template migrations dir). Use registry-iterating helpers (`createAllConfigDirStubs`, `createAllPSIDDirStubs`) so new PS-IDs added to the registry are automatically covered.

**Structural ceiling pattern**: If a stub helper necessarily creates a parent directory (e.g., `createTemplateMigrationsDirStub` creates `internal/apps/framework/...` which also creates `internal/apps/`), then a test for "absent parent dir causes error" is structurally impossible when both stubs are required. Resolve by covering the absent-parent code path via a direct function test (not via CheckInDir) and documenting the ceiling with an explanatory comment where the deleted test was.

**TestCheck delegation tests**: Tests that call `Check(logger)` (which uses `"."` as rootDir) fail when run from a package directory that lacks `configs/`, `cmd/`, etc. Always fix by calling `CheckInDir(logger, findProjectRoot(t))` explicitly in tests that delegate to the real workspace.

---

## Phase 1 (Continuation): Parameterization Items #21–#27

*(To be filled during Phase 1 execution)*

---

## Phase 2 (Continuation): TLS Init Refactoring

*(To be filled during Phase 2 execution)*

---

## Phase 3 (Continuation): Framework CLI & Magic Cleanup

*(To be filled during Phase 3 execution)*

---

## Phase 4 (Continuation): Config Test File Reorganization

*(To be filled during Phase 4 execution)*

---

## Phase 5 (Continuation): Identity Product Refactoring

*(To be filled during Phase 5 execution)*

---

## Phase 6 (Continuation): Knowledge Propagation

*(To be filled during Phase 6 execution)*
