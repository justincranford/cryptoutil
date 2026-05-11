# Implementation Plan - Framework V22: V21 Audit Fix Campaign

**Status**: Not started
**Created**: 2026-05-11
**Last Updated**: 2026-05-11
**Purpose**: Fix all 10 issues identified in the objective post-mortem audit of framework-v21
(`docs/framework-v21/SUMMARY.md`). The audit found: 4 substantive packages with zero test
coverage, 3 packages are empty stubs, 0 of 10 PS-ID E2E TestMains migrated, framework-internal
TestMains not migrated, sm-kms businesslogic/orm not migrated, old testing/ packages not cleaned
up, 2 linters below ≥98% coverage target, E2E validation never ran, mutation testing never ran,
and the claimed TestMain count of 39 was unsupported (only 20 found). All 10 issues are blockers.

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
- ✅ **Fix issues immediately** - When unknowns discovered, blockers identified, any tests fail, or
  quality gates are not met, STOP and address
- ✅ **Treat as BLOCKING** - ALL issues block progress to next phase or task
- ✅ **Document root causes** - Root cause analysis is part of planning AND implementation
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task or step complete with known issues

## Codex Execution Guidance

This plan is designed for execution by gpt-5.3-codex (or similar non-autonomous LLM agents). The following rules MUST be followed to avoid failure modes observed in framework-v21:

**Bulk task anti-pattern**: Phase 5 originally had one task “Migrate 10 PS-ID E2E TestMains.” This has been expanded to 10 individual tasks (5.2a–5.2j) in tasks.md. NEVER collapse them back into a single task — bulk operations cause incomplete migrations where some PS-IDs are missed.

**Verification discipline**: `go build ./...` passing does NOT mean migration succeeded — old imports still compile. Always run the `grep` verification command specified in each task’s acceptance criteria.

**Quality gate non-negotiability**: Each phase has a quality gate task (N.4 or N.5 or N.7). Execute ALL commands in the quality gate. Only mark the phase complete after the quality gate task is ✅.

**tasks.md as source of truth**: Update tasks.md status (❌ → ✅) immediately after each task completes. This is NOT optional documentation — it is the execution record.

**Phase ordering matters**: Phases 5–8 have dependencies on Phase 1. Do not start Phase 5 or Phase 6 until Phase 1 is ✅ complete. See the Phase Dependency Map in tasks.md.

## Overview

Framework V21 created the architectural foundation (8 new directories, 2 new linters, canonical
template update, substantive code in 4 packages) but failed to deliver on all stated goals. This
plan closes every gap the audit identified in a sequence that minimizes rework:

1. Implement the 3 empty stub packages so they exist before any consumer depends on them.
2. Write self-tests for all 7 new framework helper packages (4 substantive + 3 stubs).
3. Refactor the two new linters to reach ≥98% coverage via fs.FS injection.
4. Run mutation testing to confirm test quality across all 9 packages.
5. Implement `test_orch_e2e` facade and migrate all 10 PS-ID E2E TestMains to it.
6. Migrate the 3 framework-internal TestMains off `testutil.Initialize()`.
7. Migrate sm-kms businesslogic and orm to `test_orch_integration`/`test_help_db`.
8. Migrate all consumers off old `testing/` import paths; deprecate old packages.
9. Run E2E validation with Docker Desktop (no "Docker-deferred" deferrals).
10. Produce a definitive per-file TestMain inventory resolving the count discrepancy.
11. Propagate lessons learned to permanent artifacts.

## Technical Context

- **Language**: Go 1.26.1, CGO_ENABLED=0
- **Coverage targets**: ≥95% production, ≥98% infrastructure/utility
- **Mutation targets**: ≥98% infrastructure/utility packages
- **Key audit source**: `docs/framework-v21/SUMMARY.md` (issues 1–10)

### Affected Files

**Phase 1 — Implement stubs (3 files)**:
```
internal/apps-framework/service/test_help_tls/tls.go
internal/apps-framework/service/test_help_barrier/barrier.go
internal/apps-framework/service/test_help_bootstrap/bootstrap.go
```

**Phase 2 — Self-tests for 7 helper packages (7 new test files)**:
```
internal/apps-framework/service/test_help_tls/tls_test.go
internal/apps-framework/service/test_help_barrier/barrier_test.go
internal/apps-framework/service/test_help_bootstrap/bootstrap_test.go
internal/apps-framework/service/test_orch_integration/test_orch_integration_test.go
internal/apps-framework/service/test_help_db/database_test.go
internal/apps-framework/service/test_help_api/api_test.go
internal/apps-framework/service/test_help_cli/cli_test.go
```

**Phase 3 — Linter coverage fixes (4 files: 2 prod refactored + 2 tests updated)**:
```
internal/apps-tools/cicd_lint/lint_fitness/testmain_orchestration_policy/testmain_orchestration_policy.go
internal/apps-tools/cicd_lint/lint_fitness/testmain_orchestration_policy/testmain_orchestration_policy_test.go
internal/apps-tools/cicd_lint/lint_fitness/testmain_integration_tag_policy/testmain_integration_tag_policy.go
internal/apps-tools/cicd_lint/lint_fitness/testmain_integration_tag_policy/testmain_integration_tag_policy_test.go
```

**Phase 4 — Mutation testing (evidence only)**:
```
test-output/v22-mutation/
```

**Phase 5 — test_orch_e2e facade + 10 PS-ID E2E migrations + enforcement linter (14 files)**:
```
internal/apps-framework/service/test_orch_e2e/testmain_e2e.go                     (new facade)
internal/apps-tools/cicd_lint/lint_fitness/testmain_e2e_policy/testmain_e2e_policy.go
internal/apps-tools/cicd_lint/lint_fitness/testmain_e2e_policy/testmain_e2e_policy_test.go
internal/apps-tools/cicd_lint/lint_fitness/testmain_e2e_policy/testmain_e2e_policy_register.go
internal/apps-tools/cicd_lint/lint_fitness/registry/registry.go                   (register new linter)
internal/apps/{sm-kms,sm-im,jose-ja,pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa,skeleton-template}/e2e/testmain_e2e_test.go
```
(10 PS-ID files migrated + 4 new linter/facade files + 1 registry update = 15 file touches)

**Phase 6 — Framework-internal TestMain migration (3 files modified)**:
```
internal/apps-framework/service/server/test_main_test.go
internal/apps-framework/service/server/listener/testmain_test.go
internal/apps-framework/service/server/repository/test_main_test.go
```

**Phase 7 — sm-kms businesslogic + orm migration (2 files modified)**:
```
internal/apps/sm-kms/server/businesslogic/testmain_test.go
internal/apps/sm-kms/server/repository/orm/testmain_test.go
```
(May require additional test coverage work for businesslogic ≥95%, orm ≥95%)

**Phase 8 — Old testing/ consumer migration (17+ files modified + deprecation stubs)**:
```
# 17 consumer test files (exact list produced in task 8.0 via grep)
# Deprecation: //Deprecated: comments added to all files under:
internal/apps-framework/service/testing/{testdb,testcli,testserver,assertions,fixtures,healthclient,stubs,e2e_infra,e2e_helpers}/
```

**Phase 9 — E2E validation (evidence only)**:
```
test-output/v22-e2e/
```

**Phase 10 — TestMain inventory table (1 doc)**:
```
test-output/v22-inventory/testmain-inventory.md
```

**Phase 11 — Knowledge propagation (ENG-HANDBOOK, agents, instructions)**

**Grand total**: ~55 file touches
Formula: 3 stub impl + 7 test files + 4 linter files + 15 e2e phase touches + 3 framework migration
         + 2 sm-kms migration + 17+ consumer migrations + deprecation markers = ~55

---

## Phases

**Phase Status Legend**: `☐ TODO` | `🔄 IN PROGRESS` | `✅ COMPLETE` | `⏳ BLOCKED`

---

### Phase 1: Implement Empty Stub Packages (4h) [Status: ☐ TODO]

**Objective**: Replace the 3 empty stub packages with substantive implementations.
These are dependencies for Phases 2, 5, 6, and 7.

**Audit Issue**: Issue 2 — "Three packages are empty stubs"

**Packages**:
- `test_help_tls/tls.go`: TLS material helpers using `config/tls_generator.go`. Provide
  `NewTestTLSSettings(t) *TLSGeneratedSettings`, `NewInsecureHTTPSClient(t)`, and
  `NewMTLSClient(t, certPath, keyPath, caPool)`. Pattern: reuse logic from
  `server/testutil/helpers.go` without package-level mutable state (return values directly
  for safe `t.Parallel()` use).
- `test_help_barrier/barrier.go`: Barrier service fixture for encryption-at-rest tests.
  Provide `NewTestBarrierService(t, db *gorm.DB)` that creates an in-memory barrier with
  auto-generated unseal keys using `shared/barrier/` package.
- `test_help_bootstrap/bootstrap.go`: Config wiring helpers. Provide
  `NewTestServerSettings(t) *ServiceFrameworkServerSettings` with port=0 and auto-TLS,
  bridging framework config types to test-time construction.

**Success**:
- All 3 files exceed 50 lines (not stubs)
- `go build ./internal/apps-framework/service/test_help_{tls,barrier,bootstrap}/...` exits 0
- `golangci-lint run` passes for all 3 packages

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned — what worked,
what didn't, root causes, patterns. Evaluate artifacts for contradictions/omissions; create fix
tasks immediately.

---

### Phase 2: Self-Tests for All 7 Helper Packages (8h) [Status: ☐ TODO]

**Objective**: Write `_test.go` files for all 7 packages with 0.0% coverage; reach ≥98% each.

**Audit Issue**: Issue 1 — "Zero test coverage for test_orch_integration, test_help_api,
test_help_db, test_help_cli"

**Per-package approach**:
1. `test_help_tls` — table-driven tests for each helper function
2. `test_help_barrier` — happy + error paths for NewTestBarrierService
3. `test_help_bootstrap` — happy + error paths for NewTestServerSettings
4. `test_orch_integration` — test StartIntegrationServer, StartIntegrationServerForTestMain,
   dual-port URL accessors, health-poll timeout error path (stub server or seam injection)
5. `test_help_db` — test NewInMemorySQLiteDB, NewInMemorySQLiteDBForTestMain, NewClosedSQLiteDB;
   NewPostgresTestContainer under `//go:build integration` tag
6. `test_help_api` — HealthClient construction; Livez/Readyz/ServiceHealth/BrowserHealth using
   Fiber `app.Test()` (no real listener)
7. `test_help_cli` — RunCLITests with stub EntryFunc covering all 3 standard cases

**Mandatory test patterns**:
- `t.Parallel()` on all parent tests and subtests
- Table-driven tests for multi-case scenarios
- `app.Test()` for HTTP tests (NO real listeners in unit tests)
- UUIDv7 for dynamic test data

**Success**:
- `go test -cover ./internal/apps-framework/service/test_help_{tls,barrier,bootstrap}/...` ≥ 98% each
- `go test -cover ./internal/apps-framework/service/test_orch_integration/...` ≥ 98%
- `go test -cover ./internal/apps-framework/service/test_help_{db,api,cli}/...` ≥ 98% each
- `golangci-lint run` passes

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 3: Linter Coverage to ≥98% (3h) [Status: ☐ TODO]

**Objective**: Raise testmain_orchestration_policy (91.2%) and testmain_integration_tag_policy
(89.6%) to ≥98% by injecting OS error paths via a `readFileFn` parameter.

**Audit Issue**: Issue 7 — "Linter coverage below target"

**Pattern** (ENG-HANDBOOK §14.1.4 standalone function seam injection):
```go
// Keep public signature unchanged
func Lint(logger *zap.Logger) error {
    return lintWithReader(logger, os.ReadFile)
}

// New internal helper accepts injectable reader
func lintWithReader(logger *zap.Logger, readFileFn func(string)([]byte, error)) error {
    // ... existing logic, but call readFileFn instead of os.ReadFile
}
```
Apply the same pattern to both linters. Test with error-returning stubs to cover the OS paths.

**Success**:
- `go test -cover ./internal/apps-tools/cicd_lint/lint_fitness/testmain_orchestration_policy/...` ≥ 98%
- `go test -cover ./internal/apps-tools/cicd_lint/lint_fitness/testmain_integration_tag_policy/...` ≥ 98%
- `go run ./cmd/cicd-lint lint-fitness` exits 0
- `golangci-lint run` passes

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 4: Mutation Testing (2h) [Status: ☐ TODO]

**Objective**: Confirm test quality — not just coverage quantity — for all 9 newly-tested
packages (7 helper + 2 linter).

**Audit Issue**: Issue 9 — "Mutation testing never ran"

**Packages** (9 total):
```
internal/apps-framework/service/test_help_{tls,barrier,bootstrap,api,cli}/
internal/apps-framework/service/test_orch_integration/
internal/apps-framework/service/test_help_db/
internal/apps-tools/cicd_lint/lint_fitness/testmain_orchestration_policy/
internal/apps-tools/cicd_lint/lint_fitness/testmain_integration_tag_policy/
```

**Command** (run on Linux CI/CD — gremlins v0.6.0 panics on Windows):
```bash
gremlins unleash --tags=!integration ./path/to/package/...
```

**Target**: ≥98% mutation efficacy per package (TIMED OUT counts as detected; only LIVED fails)

**Note**: If Docker-less CI is the only gremlins option, document as CI-deferred with the
specific CI workflow step reference. NEVER mark complete without mutation evidence.

**Success**:
- All 9 packages ≥ 98% mutation efficacy OR CI-deferred with specific CI step reference
- Evidence archived in `test-output/v22-mutation/`

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 5: test_orch_e2e Facade + 10 PS-ID E2E TestMain Migration + Linter (8h) [Status: ☐ TODO]

**Objective**: Add a `SetupE2ETestMain` facade to `test_orch_e2e`, migrate all 10 PS-ID E2E
TestMains from the old `e2e_infra.SetupE2ETestMain`, and add a lint-fitness linter to prevent
regression.

**Audit Issue**: Issue 3 — "E2E TestMain migration: 0 of 10 PS-IDs"

**Task 5.1 — Add facade to test_orch_e2e**:
Add `SetupE2ETestMain(m, cfg, onReady)` to `test_orch_e2e` package. Options:
- Wrap `e2e_infra.SetupE2ETestMain` (less code, preserves existing logic)
- Re-implement using `test_orch_e2e` types (cleaner separation)
Choose wrapping first; if circular dependency, re-implement. Run `go build` before migrating.

**Task 5.2 — Migrate 10 PS-ID E2E TestMains**:
Target pattern after migration:
```go
import cryptoutilTestOrchE2e "cryptoutil/internal/apps-framework/service/test_orch_e2e"

func TestMain(m *testing.M) {
    os.Exit(cryptoutilTestOrchE2e.SetupE2ETestMain(m, cryptoutilTestOrchE2e.E2ETestConfig{...}, func(env *cryptoutilTestOrchE2e.E2ETestEnv) { ... }))
}
```
Files: `internal/apps/{sm-kms,sm-im,jose-ja,pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa,skeleton-template}/e2e/testmain_e2e_test.go`

**Task 5.3 — testmain_e2e_policy linter**:
Enforce that every file matching `*/e2e/testmain_e2e_test.go` imports `test_orch_e2e` and does
NOT import `testing/e2e_infra`.

**Task 5.4 — Register and test linter** (≥98% coverage target)

**Success**:
- `grep -r "testing/e2e_infra" --include="*testmain_e2e_test.go" internal/` returns 0 matches
- `go build -tags e2e ./...` exits 0
- `go run ./cmd/cicd-lint lint-fitness` exits 0
- `go test -cover ./internal/apps-tools/cicd_lint/lint_fitness/testmain_e2e_policy/...` ≥ 98%

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 6: Framework-Internal TestMain Migration (2h) [Status: ☐ TODO]

**Objective**: Migrate 3 framework-internal TestMains off `testutil.Initialize()`.

**Audit Issue**: Issue 4 — "Framework internal TestMains not migrated"

**Files**:
- `internal/apps-framework/service/server/test_main_test.go`
- `internal/apps-framework/service/server/listener/testmain_test.go`
- `internal/apps-framework/service/server/repository/test_main_test.go`

**Pattern**: Replace `cryptoutilAppsFrameworkServiceServerTestutil.Initialize()` with explicit
calls to `test_help_bootstrap.NewTestServerSettings(t)` and/or `test_help_tls` helpers from
Phase 1. The framework packages test TLS-dependent behavior; use `test_help_tls` for cert pools.

**Success**:
- `grep -r "testutil.Initialize" internal/apps-framework/service/server/` returns 0 matches
- `go test ./internal/apps-framework/service/server/...` passes
- `go test ./internal/apps-framework/service/server/listener/...` passes
- `go test ./internal/apps-framework/service/server/repository/...` passes

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 7: sm-kms businesslogic + orm Migration (3h) [Status: ☐ TODO]

**Objective**: Migrate sm-kms businesslogic and orm off `application.StartCore()` to use
`test_orch_integration` / `test_help_db`; raise both packages to ≥95% coverage.

**Audit Issue**: Issue 5 — "sm-kms businesslogic and orm not migrated"

**Current state**: businesslogic 93.2% (target ≥95%), orm 91.5% (target ≥95%).

**Work**:
- Investigate current TestMain in `sm-kms/server/businesslogic/` and `sm-kms/server/repository/orm/`
- Replace `application.StartCore()` with `test_orch_integration.StartIntegrationServer()`
- Use `test_help_db.NewInMemorySQLiteDB(t)` for DB fixture
- Add targeted tests to cover any newly-exposed uncovered lines

**If coverage ceiling found**: Perform coverage ceiling analysis per ENG-HANDBOOK §10.2.3;
document structural ceiling with mitigation plan. NEVER accept as permanent without plan.

**Success**:
- `go test -cover ./internal/apps/sm-kms/server/businesslogic/...` ≥ 95%
- `go test -cover ./internal/apps/sm-kms/server/repository/orm/...` ≥ 95%
- `go test ./internal/apps/sm-kms/...` fully passes

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 8: Consumer Migration + Old testing/ Deprecation (4h) [Status: ☐ TODO]

**Objective**: Migrate the 17 consumer test files off old `testing/` import paths; add
`//Deprecated:` markers to all old packages.

**Audit Issue**: Issue 6 — "Old testing/ packages not cleaned up"

**Task 8.0 — Enumerate consumers**:
```bash
grep -r "service/testing/" --include="*_test.go" internal/ | sort > test-output/v22-consumer-migration/consumers.txt
```
Confirm the count is 17 (or correct it if audit was wrong).

**Task 8.1 — Migrate each consumer** (import path mapping):

| Old import | New import |
|---|---|
| `service/testing/testdb` | `service/test_help_db` |
| `service/testing/testcli` | `service/test_help_cli` |
| `service/testing/testserver` | `service/test_orch_integration` |
| `service/testing/healthclient` | `service/test_help_api` |
| `service/testing/assertions` | assess — keep if no replacement yet |
| `service/testing/fixtures` | assess — keep if no replacement yet |
| `service/testing/e2e_infra` | `service/test_orch_e2e` (done in Phase 5) |
| `service/testing/e2e_helpers` | assess — partial overlap with test_orch_e2e |

**Task 8.2 — Add //Deprecated: comments** to package-level doc of all remaining old testing/ files.

**Task 8.3 — Verify**:
```bash
go build ./... && go test ./...
```

**Success**:
- `grep -r "service/testing/testdb\|service/testing/testcli\|service/testing/testserver" --include="*_test.go" internal/` returns 0 matches
- All old testing/ packages have `//Deprecated:` in their package doc
- `go build ./...` exits 0, `go test ./...` passes

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 9: E2E Validation (3h) [Status: ☐ TODO]

**Objective**: Run E2E tests with Docker Desktop. No "Docker-deferred" deferrals — this is
a BLOCKER per instructions.

**Audit Issue**: Issue 8 — "E2E validation never ran"

**Pre-condition**: `docker ps` must succeed before starting this phase.

**Work**:
- Build service images: `docker compose -f deployments/cryptoutil/compose.yml build`
- Run E2E tests for sm-kms and sm-im (representative sample):
  ```bash
  go test -tags e2e ./internal/apps/sm-kms/e2e/... -v
  go test -tags e2e ./internal/apps/sm-im/e2e/... -v
  ```
- Capture output to `test-output/v22-e2e/`
- If E2E fails: root cause, fix, re-run. DO NOT defer.

**Success**:
- E2E tests for sm-kms and sm-im pass
- Output archived in `test-output/v22-e2e/`
- No deferred items

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 10: TestMain Inventory Table (1h) [Status: ☐ TODO]

**Objective**: Definitive per-file TestMain inventory resolving the 39-vs-20 count discrepancy.

**Audit Issue**: Issue 10 — "TestMain count discrepancy"

**Work**:
```bash
grep -r "func TestMain" --include="*_test.go" internal/ | sort
```
Classify each occurrence: apps/e2e, apps/server, apps/client, framework/server,
framework/listener, framework/repository, linter, other. Record which import pattern each uses.
Write to `test-output/v22-inventory/testmain-inventory.md`.

**Success**:
- `test-output/v22-inventory/testmain-inventory.md` contains per-file table with grand total
- Total count explained with derivation formula

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 11: Knowledge Propagation (2h) [Status: ☐ TODO]

**Objective**: Apply all phase lessons to permanent artifacts. MANDATORY — never skipped.

**Work**:
- Review all phase post-mortems in `lessons.md`
- Update `docs/ENG-HANDBOOK.md` with new patterns: test helper usage, linter seam injection,
  migration sequencing, coverage ceiling analysis examples
- Update `.github/instructions/03-02.testing.instructions.md` if new test patterns emerged
- Update relevant `.github/agents/*.agent.md` files if agent guidance improved
- Run `go run ./cmd/cicd-lint lint-docs` to verify propagation integrity
- Commit each artifact type as a separate semantic commit

**Success**:
- `go run ./cmd/cicd-lint lint-docs` passes
- All lessons exposing ENG-HANDBOOK gaps are addressed
- All agents that should reference new patterns are updated

**Post-Mortem**: Final lessons.md entry with overall V22 outcome assessment.

---

## Executive Decisions

### Decision 1: Stub implementation approach (test_help_tls)

**Selected**: Thin wrappers over `server/testutil/helpers.go` in Phase 1, replaced with direct
implementations if coverage gaps remain in Phase 2. Fastest path to unblocking dependencies.
`server/testutil` already performs the same operations; re-implementing from scratch adds risk.

### Decision 2: Linter fs.FS refactor approach

**Selected**: Internal helper `lintWithReader(logger, readFileFn)`, public `Lint()` calls it
with `os.ReadFile`. Follows ENG-HANDBOOK §14.1.4 standalone-function seam pattern. Does not
break the public API.

### Decision 3: Old testing/ package removal timing

**Selected**: Add `//Deprecated:` comments only in this plan; full deletion in a future plan.
Reduces risk of breaking consumers missed during audit. Safe and immediate.

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| test_help_barrier depends on internal barrier API that changed | Medium | Medium | Read barrier package API before implementing; fail fast |
| sm-kms businesslogic coverage cannot reach ≥95% without seam refactor | Medium | High | Ceiling analysis per §10.2.3; document structural ceiling with mitigation plan |
| gremlins not available on Windows (v0.6.0 panics) | High | Medium | Run in CI/CD; document as CI-deferred with specific workflow step reference |
| E2E Docker tests fail due to Docker Desktop version mismatch | Low | High | Check `docker ps` first; verify testcontainers-go compatibility |
| 17-consumer count is wrong | Low | Medium | grep in task 8.0 for accurate count before migrating |
| test_orch_e2e facade creates circular import | Low | High | Run `go build` immediately after adding facade; re-implement if circular |

---

## Quality Gates - MANDATORY

**Per-Phase**:
- ✅ `go build ./...` and `go build -tags e2e,integration ./...` exit 0
- ✅ `golangci-lint run` and `golangci-lint run --build-tags e2e,integration` exit 0
- ✅ All relevant tests pass with zero skips
- ✅ Coverage targets met per package

**Overall**:
- ✅ All 11 phases complete with objective evidence
- ✅ All 7 helper packages ≥ 98% coverage
- ✅ Both linters ≥ 98% coverage
- ✅ All 10 PS-ID E2E TestMains use `test_orch_e2e`
- ✅ All 3 framework-internal TestMains use new helpers
- ✅ sm-kms businesslogic ≥ 95%, orm ≥ 95%
- ✅ Old testing/ packages deprecated
- ✅ E2E tests pass with Docker Desktop
- ✅ TestMain inventory table produced
- ✅ `go run ./cmd/cicd-lint lint-fitness` exits 0
- ✅ `go run ./cmd/cicd-lint lint-docs` exits 0
- ✅ Knowledge propagated to permanent artifacts

## ENG-HANDBOOK.md Cross-References - MANDATORY

| Topic | Section | When Referenced |
|-------|---------|----------------|
| Testing Strategy | §10 | All phases |
| Unit Testing | §10.2 | Phases 2, 6, 7, 8 |
| Coverage Ceiling | §10.2.3 | Phase 7 (sm-kms gap analysis) |
| Test Seam Injection | §10.2.4 | Phase 3 (linter fs.FS), Phase 7 |
| Integration Testing | §10.3 | Phases 2, 6, 7 |
| E2E Testing | §10.4 | Phases 5, 9 |
| Mutation Testing | §10.5 | Phase 4 |
| Quality Gates | §11.2 | All phases |
| Coding Standards | §14.1 | All phases |
| Version Control | §14.2 | All phases |
| Infrastructure Blockers | §14.7 | Phase 9 |
| Plan Lifecycle | §14.6 | This document |
| Post-Mortem & Knowledge Propagation | §14.8 | Phase 11 + all phase post-mortems |

## Success Criteria

- [ ] All 11 phases complete with objective evidence
- [ ] All 10 SUMMARY.md audit issues resolved
- [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
- [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
- [ ] Evidence archived under `test-output/v22-*/`
