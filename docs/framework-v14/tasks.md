# Tasks — Framework v14: v13 Completion

**Status**: 0 of 24 tasks complete (0%)
**Last Updated**: 2026-04-19
**Created**: 2026-04-19

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

**ALL issues are blockers — NO exceptions.**

---

## Task Status Legend

| Symbol | Meaning |
|--------|---------|
| ❌ | Not started |
| 🔄 | In progress |
| ✅ | Complete |
| ⏳ | Blocked (must have resolution plan) |

---

## Task Checklist

### Phase 1: Close v13 Cross-Cutting Quality Gates

**Phase Objective**: Explicitly verify and close every unchecked cross-cutting task inherited from
v13. These were effectively done through v13's phases but never individually confirmed.

#### Task 1.1: Build + Lint Verification
- **Status**: ❌
- **Estimated**: 10m
- **Dependencies**: None
- **Description**: Run all build and lint checks; confirm clean baseline.
- **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0
  - [ ] `go build -tags e2e,integration ./...` exits 0
  - [ ] `golangci-lint run` exits 0, zero violations
  - [ ] `golangci-lint run --build-tags e2e,integration` exits 0, zero violations
- **Evidence**: `test-output/v14-phase1/build-lint.log`

#### Task 1.2: Full Test Suite Verification
- **Status**: ❌
- **Estimated**: 15m
- **Dependencies**: Task 1.1
- **Description**: Run the full unit + integration test suite with shuffle.
- **Acceptance Criteria**:
  - [ ] `go test ./... -shuffle=on -count=1` — 100% pass, zero skips
  - [ ] Zero `InsecureSkipVerify: true` usages in e2e_tls_test.go files (CA-validated path)
  - [ ] All 4 PS-ID testmain files initialize `sharedHTTPClientWithCA` after `WaitForMultipleServices`
- **Evidence**: `test-output/v14-phase1/test-results.log`

#### Task 1.3: Fitness + Doc Linting
- **Status**: ❌
- **Estimated**: 5m
- **Dependencies**: Task 1.1
- **Description**: Run all cicd-lint checks.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
- **Evidence**: `test-output/v14-phase1/cicd-lint.log`

#### Task 1.4: Close v13 Cross-Cutting Checklist
- **Status**: ❌
- **Estimated**: 10m
- **Dependencies**: Tasks 1.1-1.3
- **Description**: Update `docs/framework-v13/tasks.md` to mark all cross-cutting items ✅.
- **Acceptance Criteria**:
  - [ ] All Testing cross-cutting items marked ✅ with evidence reference
  - [ ] All Code Quality cross-cutting items marked ✅ with evidence reference
  - [ ] All Documentation cross-cutting items marked ✅ with evidence reference
  - [ ] All Deployment cross-cutting items marked ✅ with evidence reference
  - [ ] Commit: `docs(framework-v13): close cross-cutting quality tasks`

#### Task 1.5: Phase 1 Post-Mortem
- **Status**: ❌
- **Estimated**: 10m
- **Dependencies**: Tasks 1.1-1.4
- **Description**: Update lessons.md with Phase 1 findings.
- **Acceptance Criteria**:
  - [ ] lessons.md Phase 1 section populated with What Worked / What Didn't / Root Causes / Patterns

---

### Phase 2: Admin mTLS Full Round-Trip Test

**Phase Objective**: Write a Go E2E test that verifies admin endpoint mTLS from INSIDE the Docker
container network, filling the gap left by v13 Phase 2 (which only tested port isolation).

#### Task 2.1: Audit Docker Exec Infrastructure
- **Status**: ❌
- **Estimated**: 30m
- **Dependencies**: Phase 1 complete
- **Description**: Review `composeManager.BuildDockerExecArgs` added in v13; understand what
  tooling (curl, wget, openssl) is available inside the sm-kms app container.
- **Acceptance Criteria**:
  - [ ] Confirmed: sm-kms app container image (Alpine or similar) has `wget` or `curl`
  - [ ] `BuildDockerExecArgs` API understood; sample invocation drafted
  - [ ] Admin cert paths inside the container confirmed (from compose volume mounts)

#### Task 2.2: Write `e2e_admin_mtls_test.go`
- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 2.1
- **Description**: Add `e2e_admin_mtls_test.go` to `internal/apps/sm-kms/e2e/` with table-driven
  tests for admin mTLS happy and sad paths via docker exec.
- **Acceptance Criteria**:
  - [ ] Happy path: `docker exec` with correct admin client cert → HTTP 200 from `/admin/api/v1/livez`
  - [ ] Sad path: `docker exec` without client cert → TLS handshake error (rejected)
  - [ ] Table-driven structure with `t.Parallel()` on outer test; subtests sequential (docker exec)
  - [ ] `//go:build e2e` build tag on file
  - [ ] `golangci-lint run --build-tags e2e` clean
- **Files**: `internal/apps/sm-kms/e2e/e2e_admin_mtls_test.go`

#### Task 2.3: Run New Test Against Docker Stack
- **Status**: ❌
- **Estimated**: 30m
- **Dependencies**: Task 2.2, Docker Desktop running
- **Description**: Start sm-kms Docker Compose stack and verify the new test passes.
- **Acceptance Criteria**:
  - [ ] `docker compose -f deployments/sm-kms/compose.yml up --wait` succeeds
  - [ ] `go test -tags e2e -run TestE2E_AdminMTLS ./internal/apps/sm-kms/e2e/...` passes
  - [ ] Both happy and sad path subtests pass
  - [ ] `docker compose down -v` cleans up after test

#### Task 2.4: Phase 2 Post-Mortem
- **Status**: ❌
- **Estimated**: 10m
- **Dependencies**: Tasks 2.1-2.3
- **Description**: Update lessons.md with Phase 2 findings.
- **Acceptance Criteria**:
  - [ ] lessons.md Phase 2 section populated

---

### Phase 3: pki-init Coverage Ceiling Mitigation

**Phase Objective**: Apply the `internalMain` pattern to pki-init CLI entry points to raise
coverage from 92.4% (accepted v11 ceiling) to ≥95% (mandatory production target).

#### Task 3.1: Audit pki-init CLI Entry Point Structure
- **Status**: ❌
- **Estimated**: 30m
- **Dependencies**: None
- **Description**: Read the pki-init CLI entry point (`cmd/` and `internal/apps/framework/tls/`)
  to understand the `productionNew*` functions and what blocks coverage.
- **Acceptance Criteria**:
  - [ ] Identified: which functions are currently uncoverable
  - [ ] Measured: current coverage with `go test -coverprofile=coverage.out -coverpkg=./... ./internal/apps/framework/tls/...`
  - [ ] Baseline documented in `test-output/v14-phase3/coverage-baseline.txt`
  - [ ] `internalMain` refactor scope defined: which functions to inject

#### Task 3.2: Refactor to `internalMain` Pattern
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Task 3.1
- **Description**: Refactor the CLI entry point to use function-parameter injection per
  ENG-HANDBOOK.md §10.2.4. Move production init code into injectable form.
- **Acceptance Criteria**:
  - [ ] `internalMain(args []string, stdin io.Reader, stdout, stderr io.Writer) int` function exists
  - [ ] `productionNew*` functions accepted as parameters (fn fields or function args)
  - [ ] `main()` in `cmd/` delegates to `internalMain` with production implementations
  - [ ] `golangci-lint run` clean on refactored code
  - [ ] Existing tests still pass

#### Task 3.3: Add `internalMain` Unit Tests
- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 3.2
- **Description**: Write unit tests for the `internalMain` function with injected stubs.
- **Acceptance Criteria**:
  - [ ] Tests cover: successful invocation, bad args, logger init failure, generator init failure
  - [ ] `go test -coverprofile=coverage.out ./internal/apps/framework/tls/...` → ≥95%
  - [ ] Coverage delta documented in `test-output/v14-phase3/coverage-after.txt`
  - [ ] `t.Parallel()` on all tests and subtests

#### Task 3.4: Re-run gremlins on pki-init
- **Status**: ❌
- **Estimated**: 15m
- **Dependencies**: Task 3.3
- **Description**: Re-run mutation testing to confirm efficacy is maintained or improved.
- **Acceptance Criteria**:
  - [ ] `gremlins unleash --tags=!integration ./internal/apps/framework/tls` passes
  - [ ] Efficacy: ≥95% (was 100% in v13 — new tests should maintain this)
  - [ ] Zero new LIVED (survived) mutations
  - [ ] Results in `test-output/v14-phase3/mutation-report.txt`

#### Task 3.5: Phase 3 Post-Mortem
- **Status**: ❌
- **Estimated**: 10m
- **Dependencies**: Tasks 3.1-3.4
- **Description**: Update lessons.md with Phase 3 findings.
- **Acceptance Criteria**:
  - [ ] lessons.md Phase 3 section populated
  - [ ] Retrospective Issue #10 explicitly closed (reference in commit message)

---

### Phase 4: E2E Framework Redesign — Shared TestMain Factory

**Phase Objective**: Eliminate ~200-line copy-paste TestMain boilerplate across 4 PS-ID e2e
directories. Implement a parameterized factory in `e2e_infra/` per v13 deferred Item 7.

#### Task 4.1: Audit Current TestMain Boilerplate
- **Status**: ❌
- **Estimated**: 30m
- **Dependencies**: None
- **Description**: Read all 4 PS-ID `testmain_e2e_test.go` files side-by-side; identify shared vs
  PS-ID-specific code; design the factory interface.
- **Acceptance Criteria**:
  - [ ] Shared code identified: `ComposeManager` setup, `WaitForMultipleServices`, HTTP client init
  - [ ] PS-ID-specific code identified: magic constants (compose path, health URLs, CA cert path)
  - [ ] Factory interface drafted: method signatures documented in task notes
  - [ ] Findings in `test-output/v14-phase4/testmain-audit.md`

#### Task 4.2: Implement `testmain_factory.go` in `e2e_infra`
- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Task 4.1
- **Description**: Write the shared TestMain factory in `internal/apps/framework/service/testing/e2e_infra/testmain_factory.go`.
- **Acceptance Criteria**:
  - [ ] `E2ETestEnv` struct with fields: `ComposeManager`, `InsecureClient *http.Client`, `SecureClient *http.Client`
  - [ ] `NewE2ETestEnv(cfg E2ETestConfig) (*E2ETestEnv, error)` constructor
  - [ ] `E2ETestConfig` struct accepts: compose file path, health URLs ([]string), CA cert path
  - [ ] `SetupE2ETestMain(m *testing.M, cfg E2ETestConfig) int` top-level helper (wraps Setup + os.Exit)
  - [ ] `//go:build e2e` tag on file
  - [ ] `golangci-lint run --build-tags e2e` clean
- **Files**: `internal/apps/framework/service/testing/e2e_infra/testmain_factory.go`

#### Task 4.3: Unit Tests for `testmain_factory.go`
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 4.2
- **Description**: Write unit tests for the factory with seam-injected compose manager and HTTP client stubs.
- **Acceptance Criteria**:
  - [ ] Tests cover: successful init, missing CA cert file, compose startup failure, health check timeout
  - [ ] `go test -cover -tags e2e ./internal/apps/framework/service/testing/e2e_infra/...` → ≥95%
  - [ ] `t.Parallel()` on all applicable tests

#### Task 4.4: Migrate All 4 PS-ID TestMain Files
- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 4.3
- **Description**: Update `testmain_e2e_test.go` in all 4 PS-ID e2e directories to use the factory.
- **Acceptance Criteria**:
  - [ ] `internal/apps/sm-kms/e2e/testmain_e2e_test.go` uses `SetupE2ETestMain`
  - [ ] `internal/apps/jose-ja/e2e/testmain_e2e_test.go` uses `SetupE2ETestMain`
  - [ ] `internal/apps/sm-im/e2e/testmain_e2e_test.go` uses `SetupE2ETestMain`
  - [ ] `internal/apps/skeleton-template/e2e/testmain_e2e_test.go` uses `SetupE2ETestMain`
  - [ ] Old boilerplate removed (net line reduction ≥100 lines across 4 files)
  - [ ] `golangci-lint run --build-tags e2e,integration` clean

#### Task 4.5: Smoke-Test Migrated E2E Suites
- **Status**: ❌
- **Estimated**: 30m
- **Dependencies**: Task 4.4, Docker Desktop running
- **Description**: Run one PS-ID's E2E suite to confirm the factory works end-to-end.
- **Acceptance Criteria**:
  - [ ] Start skeleton-template Docker Compose stack: `docker compose up --wait`
  - [ ] `go test -tags e2e -run TestE2E ./internal/apps/skeleton-template/e2e/...` passes
  - [ ] `docker compose down -v` after

#### Task 4.6: Phase 4 Post-Mortem
- **Status**: ❌
- **Estimated**: 10m
- **Dependencies**: Tasks 4.1-4.5
- **Description**: Update lessons.md with Phase 4 findings.
- **Acceptance Criteria**:
  - [ ] lessons.md Phase 4 section populated
  - [ ] v13 Item 7 explicitly closed (reference in commit message)

---

### Phase 5: Mutation Testing on e2e_infra Code

**Phase Objective**: Run gremlins mutation testing on the `e2e_infra` package, which gained
significant new code in v13 (`BuildDockerExecArgs`) and v14 (TestMain factory). These additions
were never mutation-tested.

#### Task 5.1: Run gremlins on `e2e_infra`
- **Status**: ❌
- **Estimated**: 20m
- **Dependencies**: Phase 4 complete
- **Description**: Run `gremlins unleash` on the e2e_infra package.
- **Acceptance Criteria**:
  - [ ] `gremlins unleash --tags=e2e ./internal/apps/framework/service/testing/e2e_infra` runs without error
  - [ ] Efficacy ≥95% (all mutations killed or timed out)
  - [ ] LIVED count: 0 (no surviving mutations)
  - [ ] NOT COVERED paths documented and justified
  - [ ] Results in `test-output/v14-phase5/mutation-report.txt`

#### Task 5.2: Fix Surviving Mutations (if any)
- **Status**: ❌
- **Estimated**: 30m
- **Dependencies**: Task 5.1
- **Description**: Add tests for any surviving mutations found in Task 5.1.
- **Acceptance Criteria**:
  - [ ] All surviving mutations addressed (new tests or documented impossibility)
  - [ ] Re-run confirms efficacy ≥95%

#### Task 5.3: Race Detection on `e2e_infra`
- **Status**: ❌
- **Estimated**: 10m
- **Dependencies**: Phase 4 complete
- **Description**: Run the race detector on e2e_infra unit tests.
- **Acceptance Criteria**:
  - [ ] `CGO_ENABLED=1 go test -race -count=2 -tags=e2e ./internal/apps/framework/service/testing/e2e_infra/...` → zero data races
  - [ ] Results in `test-output/v14-phase5/race-report.txt`

#### Task 5.4: Phase 5 Post-Mortem
- **Status**: ❌
- **Estimated**: 10m
- **Dependencies**: Tasks 5.1-5.3
- **Description**: Update lessons.md with Phase 5 findings.
- **Acceptance Criteria**:
  - [ ] lessons.md Phase 5 section populated

---

### Phase 6: Knowledge Propagation

**Phase Objective**: Apply v14 lessons to permanent artifacts.

#### Task 6.1: Review All Lessons
- **Status**: ❌
- **Estimated**: 10m
- **Dependencies**: Phases 1-5 complete
- **Description**: Review lessons.md Phases 1-5; categorize by artifact impact.
- **Acceptance Criteria**:
  - [ ] All lessons categorized (ENG-HANDBOOK, agents, skills, instructions, code, tests, workflows)
  - [ ] Priority assigned per lesson

#### Task 6.2: Update ENG-HANDBOOK.md
- **Status**: ❌
- **Estimated**: 30m
- **Dependencies**: Task 6.1
- **Description**: Update ENG-HANDBOOK.md with v14 patterns.
- **Acceptance Criteria**:
  - [ ] §10.4 E2E Testing: admin mTLS docker exec test pattern documented
  - [ ] §10.2.3 Coverage Targets: `internalMain` applicability to pki-init documented
  - [ ] §10.3.6 Shared Test Infrastructure: TestMain factory pattern added
  - [ ] lint-docs passes with zero errors

#### Task 6.3: Update Agents, Skills, Instructions
- **Status**: ❌
- **Estimated**: 15m
- **Dependencies**: Task 6.1
- **Description**: Update instruction/skill files where v14 work exposed gaps.
- **Acceptance Criteria**:
  - [ ] Relevant instruction files updated (if applicable)
  - [ ] No agent/skill updates needed (or updates made with lint-agent-drift passing)
  - [ ] lint-docs passes

#### Task 6.4: Verify Propagation Integrity
- **Status**: ❌
- **Estimated**: 5m
- **Dependencies**: Tasks 6.2, 6.3
- **Description**: Run `go run ./cmd/cicd-lint lint-docs` to confirm zero propagation drift.
- **Acceptance Criteria**:
  - [ ] lint-docs passes with zero errors

---

## Cross-Cutting Tasks

### Testing (verified in Phase 1; maintained throughout)
- [ ] All tests pass (`go test ./... -shuffle=on`)
- [ ] pki-init coverage ≥95% (raised from v11/v13 ceiling)
- [ ] e2e_infra coverage ≥95%
- [ ] Mutation efficacy ≥95% for pki-init and e2e_infra
- [ ] Race detector clean on all new code

### Code Quality
- [ ] `golangci-lint run` clean
- [ ] `golangci-lint run --build-tags e2e,integration` clean
- [ ] `go build ./...` clean
- [ ] `go build -tags e2e,integration ./...` clean

### Documentation
- [ ] ENG-HANDBOOK.md updated with v14 patterns
- [ ] lint-docs passes
- [ ] v13 cross-cutting tasks explicitly closed

### Deployment
- [ ] sm-kms Docker Compose: admin mTLS test passes
- [ ] skeleton-template Docker Compose: migrated E2E suite passes

---

## Notes / Deferred Work

- **Framework-v15 prerequisite**: This plan MUST be complete before beginning framework-v15
  (OTel/Grafana mTLS + Public PS-ID App TLS Trust — now in `docs/framework-v15/`).
- **16-deployment orchestrator**: v13 Item 7 mentioned this scope. Phase 4 implements the shared
  factory (core redesign). A full 16-deployment test runner (all 4 PS-IDs × 4 variants in one run)
  may be addressed in a future plan if needed.

---

## Evidence Archive

- `test-output/v14-phase1/` — Cross-cutting quality gate verification logs
- `test-output/v14-phase2/` — Admin mTLS test results
- `test-output/v14-phase3/` — pki-init coverage baseline + after; mutation report
- `test-output/v14-phase4/` — TestMain factory audit and smoke test logs
- `test-output/v14-phase5/` — e2e_infra mutation and race results
