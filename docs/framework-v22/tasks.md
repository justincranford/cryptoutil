# Tasks - Framework V22: V21 Audit Fix Campaign

**Status**: 0 of 62 tasks complete (0%)
**Last Updated**: 2026-05-11
**Created**: 2026-05-11

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

---

## Task Status Legend

| Symbol | Meaning | When to Use |
|--------|---------|-------------|
| ❌ | Not started | Task not yet begun |
| 🔄 | In progress | Currently being worked on |
| ✅ | Complete | Task finished with evidence |
| ⏳ | Blocked | Requires external dependency (MUST have resolution plan) |

---

## Phase 1: Implement Empty Stub Packages

**Phase Objective**: Replace 3 empty stub packages with functional implementations.
Fixes SUMMARY.md Issue 2.

### Task 1.1: Implement test_help_tls/tls.go

- **Status**: ❌
- **Estimated**: 1.5h
- **Actual**: —
- **Dependencies**: None
- **Acceptance Criteria**:
  - [ ] File > 50 lines (not a stub)
  - [ ] `NewTestTLSSettings(t)` returns a TLSGeneratedSettings with auto-generated self-signed cert
  - [ ] `NewInsecureHTTPSClient(t)` returns `*http.Client` with `InsecureSkipVerify: true`
  - [ ] `NewMTLSClient(t, certPath, keyPath, caPool)` returns `*http.Client` with mTLS configured
  - [ ] No package-level mutable state (all functions take `t *testing.T`, return values directly)
  - [ ] `go build ./internal/apps-framework/service/test_help_tls/...` exits 0
  - [ ] `golangci-lint run ./internal/apps-framework/service/test_help_tls/...` exits 0
- **Files**: `internal/apps-framework/service/test_help_tls/tls.go`

### Task 1.2: Implement test_help_barrier/barrier.go

- **Status**: ❌
- **Estimated**: 1.5h
- **Actual**: —
- **Dependencies**: None
- **Acceptance Criteria**:
  - [ ] File > 50 lines (not a stub)
  - [ ] `NewTestBarrierService(t, db *gorm.DB)` creates an in-memory barrier with auto-generated unseal keys
  - [ ] Uses `shared/barrier/` package; no custom crypto
  - [ ] `go build ./internal/apps-framework/service/test_help_barrier/...` exits 0
  - [ ] `golangci-lint run` passes
- **Files**: `internal/apps-framework/service/test_help_barrier/barrier.go`

### Task 1.3: Implement test_help_bootstrap/bootstrap.go

- **Status**: ❌
- **Estimated**: 1h
- **Actual**: —
- **Dependencies**: None
- **Acceptance Criteria**:
  - [ ] File > 50 lines (not a stub)
  - [ ] `NewTestServerSettings(t)` returns `*ServiceFrameworkServerSettings` with port=0, auto-TLS
  - [ ] Return values are safe for `t.Parallel()` (no shared state)
  - [ ] `go build ./internal/apps-framework/service/test_help_bootstrap/...` exits 0
  - [ ] `golangci-lint run` passes
- **Files**: `internal/apps-framework/service/test_help_bootstrap/bootstrap.go`

### Task 1.4: Phase 1 quality gate

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0
  - [ ] `golangci-lint run` exits 0
  - [ ] All 3 files > 50 lines

---

## Phase 2: Self-Tests for All 7 Helper Packages

**Phase Objective**: Write test files reaching ≥98% coverage for all 7 helper packages.
Fixes SUMMARY.md Issue 1.

### Task 2.1: Tests for test_help_tls

- **Status**: ❌
- **Estimated**: 1h
- **Acceptance Criteria**:
  - [ ] `t.Parallel()` on all tests and subtests
  - [ ] Table-driven tests for each function
  - [ ] `go test -cover ./internal/apps-framework/service/test_help_tls/...` ≥ 98%
- **Files**: `internal/apps-framework/service/test_help_tls/tls_test.go`

### Task 2.2: Tests for test_help_barrier

- **Status**: ❌
- **Estimated**: 1h
- **Acceptance Criteria**:
  - [ ] Happy path + error paths for NewTestBarrierService
  - [ ] `go test -cover ./internal/apps-framework/service/test_help_barrier/...` ≥ 98%
- **Files**: `internal/apps-framework/service/test_help_barrier/barrier_test.go`

### Task 2.3: Tests for test_help_bootstrap

- **Status**: ❌
- **Estimated**: 0.5h
- **Acceptance Criteria**:
  - [ ] Happy path + error paths for NewTestServerSettings
  - [ ] `go test -cover ./internal/apps-framework/service/test_help_bootstrap/...` ≥ 98%
- **Files**: `internal/apps-framework/service/test_help_bootstrap/bootstrap_test.go`

### Task 2.4: Tests for test_orch_integration

- **Status**: ❌
- **Estimated**: 2h
- **Acceptance Criteria**:
  - [ ] Tests for StartIntegrationServer, StartIntegrationServerForTestMain
  - [ ] Dual-port URL accessor tests
  - [ ] Health-poll timeout error path covered (seam injection or stub server)
  - [ ] `go test -cover ./internal/apps-framework/service/test_orch_integration/...` ≥ 98%
- **Files**: `internal/apps-framework/service/test_orch_integration/test_orch_integration_test.go`

### Task 2.5: Tests for test_help_db

- **Status**: ❌
- **Estimated**: 1.5h
- **Acceptance Criteria**:
  - [ ] Tests for NewInMemorySQLiteDB, NewInMemorySQLiteDBForTestMain, NewClosedSQLiteDB (no container)
  - [ ] NewPostgresTestContainer under `//go:build integration` tag
  - [ ] `go test -cover ./internal/apps-framework/service/test_help_db/...` ≥ 98%
- **Files**: `internal/apps-framework/service/test_help_db/database_test.go`

### Task 2.6: Tests for test_help_api

- **Status**: ❌
- **Estimated**: 1h
- **Acceptance Criteria**:
  - [ ] HealthClient construction test
  - [ ] Livez/Readyz/ServiceHealth/BrowserHealth tested with `app.Test()` (NO real listener)
  - [ ] `go test -cover ./internal/apps-framework/service/test_help_api/...` ≥ 98%
- **Files**: `internal/apps-framework/service/test_help_api/api_test.go`

### Task 2.7: Tests for test_help_cli

- **Status**: ❌
- **Estimated**: 0.5h
- **Acceptance Criteria**:
  - [ ] RunCLITests with stub EntryFunc covering all 3 standard cases (nil args, help, version)
  - [ ] `go test -cover ./internal/apps-framework/service/test_help_cli/...` ≥ 98%
- **Files**: `internal/apps-framework/service/test_help_cli/cli_test.go`

### Task 2.8: Phase 2 quality gate

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All 7 packages ≥ 98% coverage (verified with `-coverprofile`)
  - [ ] `go build ./...` exits 0
  - [ ] `golangci-lint run` exits 0
  - [ ] Evidence: `test-output/v22-phase2/coverage-*.txt`

---

## Phase 3: Linter Coverage to ≥98%

**Phase Objective**: Reach ≥98% coverage for both new linters via fs.FS/readFileFn injection.
Fixes SUMMARY.md Issue 7.

### Task 3.1: Refactor testmain_orchestration_policy for injectable reader

- **Status**: ❌
- **Estimated**: 1.5h
- **Acceptance Criteria**:
  - [ ] `lintWithReader(logger, readFileFn)` internal helper added
  - [ ] Public `Lint(logger)` delegates to it with `os.ReadFile`
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0 (no regression)
  - [ ] `golangci-lint run` passes
- **Files**: `internal/apps-tools/cicd_lint/lint_fitness/testmain_orchestration_policy/testmain_orchestration_policy.go`

### Task 3.2: Refactor testmain_integration_tag_policy for injectable reader

- **Status**: ❌
- **Estimated**: 1.5h
- **Acceptance Criteria**:
  - [ ] Same pattern as 3.1
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
- **Files**: `internal/apps-tools/cicd_lint/lint_fitness/testmain_integration_tag_policy/testmain_integration_tag_policy.go`

### Task 3.3: Update tests for testmain_orchestration_policy to reach ≥98%

- **Status**: ❌
- **Estimated**: 1h
- **Acceptance Criteria**:
  - [ ] Error-returning stub injected for readFileFn to cover OS error path
  - [ ] `go test -cover ./internal/apps-tools/cicd_lint/lint_fitness/testmain_orchestration_policy/...` ≥ 98%
- **Files**: `internal/apps-tools/cicd_lint/lint_fitness/testmain_orchestration_policy/testmain_orchestration_policy_test.go`

### Task 3.4: Update tests for testmain_integration_tag_policy to reach ≥98%

- **Status**: ❌
- **Estimated**: 1h
- **Acceptance Criteria**:
  - [ ] `go test -cover ./internal/apps-tools/cicd_lint/lint_fitness/testmain_integration_tag_policy/...` ≥ 98%
- **Files**: `internal/apps-tools/cicd_lint/lint_fitness/testmain_integration_tag_policy/testmain_integration_tag_policy_test.go`

### Task 3.5: Phase 3 quality gate

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Both linters ≥ 98% coverage
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] `go build ./...` exits 0
  - [ ] `golangci-lint run` exits 0

---

## Phase 4: Mutation Testing

**Phase Objective**: ≥98% mutation efficacy for all 9 newly-tested packages.
Fixes SUMMARY.md Issue 9.

### Task 4.1: Run gremlins for all 7 helper packages

- **Status**: ❌
- **Estimated**: 1h (CI run time)
- **Acceptance Criteria**:
  - [ ] Each of 7 packages ≥ 98% mutation efficacy OR documented CI-deferred with CI step reference
  - [ ] Evidence in `test-output/v22-mutation/helpers-*.txt`
- **Note**: gremlins v0.6.0 panics on Windows — run on Linux CI/CD

### Task 4.2: Run gremlins for both linter packages

- **Status**: ❌
- **Estimated**: 0.5h
- **Acceptance Criteria**:
  - [ ] testmain_orchestration_policy ≥ 98% mutation efficacy
  - [ ] testmain_integration_tag_policy ≥ 98% mutation efficacy
  - [ ] Evidence in `test-output/v22-mutation/linters-*.txt`

### Task 4.3: Phase 4 quality gate

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All 9 packages have mutation evidence (passed or documented CI-deferred)
  - [ ] No LIVED mutations without documented rationale

---

## Phase 5: test_orch_e2e Facade + 10 PS-ID E2E TestMain Migration + Linter

**Phase Objective**: Add SetupE2ETestMain to test_orch_e2e, migrate all 10 PS-ID E2E TestMains,
add enforcement linter. Fixes SUMMARY.md Issue 3.

### Task 5.1: Add SetupE2ETestMain facade to test_orch_e2e

- **Status**: ❌
- **Estimated**: 2h
- **Acceptance Criteria**:
  - [ ] `SetupE2ETestMain(m, cfg, onReady)` exists in `test_orch_e2e` package
  - [ ] `E2ETestConfig` and `E2ETestEnv` types accessible from `test_orch_e2e`
  - [ ] `go build -tags e2e ./internal/apps-framework/service/test_orch_e2e/...` exits 0
  - [ ] No circular imports
- **Files**: `internal/apps-framework/service/test_orch_e2e/testmain_e2e.go`

### Task 5.2: Migrate 10 PS-ID E2E TestMains

- **Status**: ❌
- **Estimated**: 2h
- **Acceptance Criteria**:
  - [ ] All 10 `internal/apps/*/e2e/testmain_e2e_test.go` import `test_orch_e2e`
  - [ ] `grep -r "testing/e2e_infra" --include="*testmain_e2e_test.go" internal/` returns 0 matches
  - [ ] `go build -tags e2e ./...` exits 0
- **Files**: 10 × `internal/apps/{PSID}/e2e/testmain_e2e_test.go`

### Task 5.3: Create testmain_e2e_policy lint-fitness linter

- **Status**: ❌
- **Estimated**: 2h
- **Acceptance Criteria**:
  - [ ] Linter detects `*/e2e/testmain_e2e_test.go` files that do NOT import `test_orch_e2e`
  - [ ] Linter detects `*/e2e/testmain_e2e_test.go` files that DO import `testing/e2e_infra`
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] `go test -cover ./internal/apps-tools/cicd_lint/lint_fitness/testmain_e2e_policy/...` ≥ 98%
- **Files**:
  - `internal/apps-tools/cicd_lint/lint_fitness/testmain_e2e_policy/testmain_e2e_policy.go`
  - `internal/apps-tools/cicd_lint/lint_fitness/testmain_e2e_policy/testmain_e2e_policy_test.go`
  - `internal/apps-tools/cicd_lint/lint_fitness/testmain_e2e_policy/testmain_e2e_policy_register.go`
  - `internal/apps-tools/cicd_lint/lint_fitness/registry/registry.go` (add registration)

### Task 5.4: Phase 5 quality gate

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All 10 E2E TestMains use test_orch_e2e (confirmed by grep)
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] `go build -tags e2e,integration ./...` exits 0
  - [ ] `golangci-lint run --build-tags e2e,integration` exits 0

---

## Phase 6: Framework-Internal TestMain Migration

**Phase Objective**: Migrate 3 framework-internal TestMains off testutil.Initialize().
Fixes SUMMARY.md Issue 4.

### Task 6.1: Migrate server/test_main_test.go

- **Status**: ❌
- **Estimated**: 0.5h
- **Acceptance Criteria**:
  - [ ] `testutil.Initialize()` removed; replaced with test_help_bootstrap/test_help_tls calls
  - [ ] `go test ./internal/apps-framework/service/server/...` passes
- **Files**: `internal/apps-framework/service/server/test_main_test.go`

### Task 6.2: Migrate server/listener/testmain_test.go

- **Status**: ❌
- **Estimated**: 0.5h
- **Acceptance Criteria**:
  - [ ] `testutil.Initialize()` removed
  - [ ] `go test ./internal/apps-framework/service/server/listener/...` passes
- **Files**: `internal/apps-framework/service/server/listener/testmain_test.go`

### Task 6.3: Migrate server/repository/test_main_test.go

- **Status**: ❌
- **Estimated**: 0.5h
- **Acceptance Criteria**:
  - [ ] `testutil.Initialize()` removed
  - [ ] `go test ./internal/apps-framework/service/server/repository/...` passes
- **Files**: `internal/apps-framework/service/server/repository/test_main_test.go`

### Task 6.4: Phase 6 quality gate

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `grep -r "testutil.Initialize" internal/apps-framework/service/server/` returns 0 matches
  - [ ] `go test ./internal/apps-framework/service/server/...` passes (all sub-packages)
  - [ ] `golangci-lint run` exits 0

---

## Phase 7: sm-kms businesslogic + orm Migration

**Phase Objective**: Migrate sm-kms businesslogic and orm to test_orch_integration/test_help_db;
raise both to ≥95% coverage. Fixes SUMMARY.md Issue 5.

### Task 7.1: Investigate and migrate businesslogic/testmain_test.go

- **Status**: ❌
- **Estimated**: 1.5h
- **Acceptance Criteria**:
  - [ ] `application.StartCore()` replaced with `test_orch_integration.StartIntegrationServer()`
  - [ ] `test_help_db.NewInMemorySQLiteDB(t)` used for DB fixture
  - [ ] `go test -cover ./internal/apps/sm-kms/server/businesslogic/...` ≥ 95%
- **Files**: `internal/apps/sm-kms/server/businesslogic/testmain_test.go` (modified)

### Task 7.2: Investigate and migrate repository/orm/testmain_test.go

- **Status**: ❌
- **Estimated**: 1.5h
- **Acceptance Criteria**:
  - [ ] `application.StartCore()` replaced
  - [ ] `go test -cover ./internal/apps/sm-kms/server/repository/orm/...` ≥ 95%
- **Files**: `internal/apps/sm-kms/server/repository/orm/testmain_test.go` (modified)

### Task 7.3: Phase 7 quality gate

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] businesslogic ≥ 95%, orm ≥ 95% (both confirmed by -coverprofile)
  - [ ] `go test ./internal/apps/sm-kms/...` fully passes
  - [ ] `golangci-lint run` exits 0

---

## Phase 8: Consumer Migration + Old testing/ Deprecation

**Phase Objective**: Migrate 17 consumers off old testing/ imports; deprecate old packages.
Fixes SUMMARY.md Issue 6.

### Task 8.0: Enumerate all consumers

- **Status**: ❌
- **Estimated**: 0.25h
- **Acceptance Criteria**:
  - [ ] `grep -r "service/testing/" --include="*_test.go" internal/ | sort` output saved
  - [ ] Count confirmed (expected ~17; correct if wrong)
  - [ ] List written to `test-output/v22-consumer-migration/consumers.txt`

### Task 8.1: Migrate consumers of service/testing/testdb

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All files importing `service/testing/testdb` now import `service/test_help_db`
  - [ ] `go build ./...` exits 0

### Task 8.2: Migrate consumers of service/testing/testcli

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All files importing `service/testing/testcli` now import `service/test_help_cli`
  - [ ] `go build ./...` exits 0

### Task 8.3: Migrate consumers of service/testing/testserver

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All files importing `service/testing/testserver` now import `service/test_orch_integration`
  - [ ] `go build ./...` exits 0

### Task 8.4: Migrate consumers of service/testing/healthclient

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All files importing `service/testing/healthclient` now import `service/test_help_api`
  - [ ] `go build ./...` exits 0

### Task 8.5: Assess assertions, fixtures, stubs, e2e_helpers consumers

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] For each: either migrate (if new equivalent exists) or keep + deprecate
  - [ ] Decision documented in task notes

### Task 8.6: Add //Deprecated: to all old testing/ package docs

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Every package under `internal/apps-framework/service/testing/` has `//Deprecated:` in package doc
  - [ ] `go build ./...` exits 0

### Task 8.7: Phase 8 quality gate

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `grep -r "service/testing/testdb\|service/testing/testcli\|service/testing/testserver\|service/testing/healthclient" --include="*_test.go" internal/` returns 0 matches
  - [ ] `go build ./...` exits 0
  - [ ] `go test ./...` passes
  - [ ] `golangci-lint run` exits 0

---

## Phase 9: E2E Validation

**Phase Objective**: Run E2E tests with Docker Desktop. No deferrals.
Fixes SUMMARY.md Issue 8.

### Task 9.1: Verify Docker Desktop is running

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `docker ps` exits 0

### Task 9.2: Build service images

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `docker compose -f deployments/cryptoutil/compose.yml build` exits 0

### Task 9.3: Run sm-kms E2E tests

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go test -tags e2e ./internal/apps/sm-kms/e2e/... -v` passes
  - [ ] Output archived in `test-output/v22-e2e/sm-kms.log`

### Task 9.4: Run sm-im E2E tests

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go test -tags e2e ./internal/apps/sm-im/e2e/... -v` passes
  - [ ] Output archived in `test-output/v22-e2e/sm-im.log`

### Task 9.5: Phase 9 quality gate

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Both sm-kms and sm-im E2E tests pass
  - [ ] Evidence in `test-output/v22-e2e/`
  - [ ] No deferred items

---

## Phase 10: TestMain Inventory Table

**Phase Objective**: Definitive per-file TestMain inventory resolving count discrepancy.
Fixes SUMMARY.md Issue 10.

### Task 10.1: Generate complete TestMain inventory

- **Status**: ❌
- **Estimated**: 1h
- **Acceptance Criteria**:
  - [ ] `grep -r "func TestMain" --include="*_test.go" internal/ | sort` output captured
  - [ ] Each file classified by category and import pattern
  - [ ] Written to `test-output/v22-inventory/testmain-inventory.md`
  - [ ] Grand total with derivation formula documented
  - [ ] V21's "39" claim assessed with evidence

---

## Phase 11: Knowledge Propagation

**Phase Objective**: Apply all lessons to permanent artifacts. MANDATORY — never skipped.

### Task 11.1: Review all phase post-mortems in lessons.md

- **Status**: ❌
- **Estimated**: 0.5h

### Task 11.2: Update ENG-HANDBOOK.md

- **Status**: ❌
- **Estimated**: 1h
- **Acceptance Criteria**:
  - [ ] New patterns documented: test helper usage, linter seam injection, migration sequencing
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

### Task 11.3: Update instruction files

- **Status**: ❌
- **Estimated**: 0.5h
- **Acceptance Criteria**:
  - [ ] `.github/instructions/03-02.testing.instructions.md` updated if new patterns emerged
  - [ ] Other instruction files updated as needed

### Task 11.4: Update agent files

- **Status**: ❌
- **Estimated**: 0.25h
- **Acceptance Criteria**:
  - [ ] Relevant `.github/agents/*.agent.md` files updated if guidance improved
  - [ ] Copilot and Claude counterpart files kept in sync

### Task 11.5: Final quality gate

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] `go build ./...` exits 0
  - [ ] `golangci-lint run` exits 0
  - [ ] All 11 phases marked ✅ with objective evidence

---

## Cross-Cutting Tasks

### Testing

- [ ] Unit tests ≥ 98% coverage (infrastructure packages: all 7 helpers + 3 linters)
- [ ] Unit tests ≥ 95% coverage (production packages: businesslogic, orm)
- [ ] Integration tests pass (`go test -tags integration ./...`)
- [ ] E2E tests pass with Docker Desktop (Phase 9)
- [ ] Mutation testing ≥ 98% for all infrastructure packages (Phase 4)
- [ ] Race detector clean: `go test -race ./...`

### Code Quality

- [ ] `golangci-lint run ./...` passes
- [ ] `golangci-lint run --build-tags e2e,integration ./...` passes
- [ ] No new TODOs without tracking in tasks.md
- [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0

### Documentation

- [ ] `go run ./cmd/cicd-lint lint-docs` passes
- [ ] ENG-HANDBOOK.md updated (Phase 11)
- [ ] Instruction files updated (Phase 11)

---

## Evidence Archive

- `test-output/v22-phase2/` — Phase 2 coverage reports
- `test-output/v22-mutation/` — Phase 4 mutation testing results
- `test-output/v22-consumer-migration/` — Phase 8 consumer enumeration
- `test-output/v22-e2e/` — Phase 9 E2E test logs
- `test-output/v22-inventory/` — Phase 10 TestMain inventory table
