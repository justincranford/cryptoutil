# Tasks - Unified Implementation

**Status**: 219 of 295 tasks complete (74%)
**Last Updated**: 2026-01-25

**Summary**:
- Phase 1: CI/CD Enhancement (NEW) - 0 of 3 tasks complete (5.5h estimated)
- Phase 2: Healthcheck Testing with app.Test() (REVISED) - 0 of 2 tasks complete (2.5h estimated)
- Phase 3: V2 Priorities (E2E Verification) - 0 of 2 tasks complete (2.5h estimated)
- Phase 4-6: V1 Phases (Coverage, Blockers, Mutations) - 0 of 70 tasks complete (41h estimated)
- Completed: 219 tasks (see completed.md)

---

## Phase 1: CI/CD Enhancement - time.Now().UTC() Formatter

**Purpose**: Prevent recurring LLM agent mistakes by enforcing `time.Now().UTC()` through automated formatting

**Context**:
LLM agents repeatedly use `time.Now()` without `.UTC()` despite copilot instructions warning about SQLite+GORM timezone issues. Non-UTC timestamps cause test failures in non-UTC timezones (tests pass in CI/UTC but fail locally in PST/EST).

**Root Cause**: Manual instruction compliance is unreliable for repetitive patterns. Automated enforcement through pre-commit hooks is the only reliable solution.

**Solution**:
Add new formatter to format-go subcommand that finds all `time.Now()` calls (without `.UTC()`) and replaces them with `time.Now().UTC()`. This formatter runs during pre-commit hooks to block mistaken code from being committed.

**Evidence**: Similar pattern exists with `enforce_any.go` (replaces `interface{}` with `any`). This demonstrates viability of AST-based formatters for project-specific code standards.

---

### 1.1: Implement time.Now().UTC() Formatter

**Owner**: LLM Agent
**Estimated**: 4h
**Dependencies**: None
**Priority**: P0 (Critical - prevents test failures)

**Description**:
Extend format-go subcommand with new formatter that enforces UTC standardization for all time.Now() calls.

**Acceptance Criteria**:
- [ ] 1.1.1 Create formatter function `enforceTimeNowUTC()` in format-go package
- [ ] 1.1.2 Use AST traversal to find all `time.Now()` call expressions
- [ ] 1.1.3 Detect if `.UTC()` selector already present (skip if already correct)
- [ ] 1.1.4 Replace `time.Now()` with `time.Now().UTC()` using AST rewriting
- [ ] 1.1.5 Add self-exclusion filter (skip format-go package itself to avoid self-modification)
- [ ] 1.1.6 Add unit tests with test cases:
  - [ ] 1.1.6.1 `time.Now()`  `time.Now().UTC()` (basic replacement)
  - [ ] 1.1.6.2 `time.Now().UTC()`  `time.Now().UTC()` (already correct, no change)
  - [ ] 1.1.6.3 `time.Now().Add(1*time.Hour)`  `time.Now().UTC().Add(1*time.Hour)` (chained method calls)
  - [ ] 1.1.6.4 `t := time.Now(); t.UTC()`  `t := time.Now().UTC(); t.UTC()` (variable assignment)
  - [ ] 1.1.6.5 Format-go package files excluded (self-exclusion verification)
- [ ] 1.1.7 Add integration test: Run formatter on test files containing time.Now(), verify replacements
- [ ] 1.1.8 Update format-go CLI to include new formatter in default execution
- [ ] 1.1.9 Run tests: `go test ./internal/cmd/cicd/format_go/... -v`
- [ ] 1.1.10 All tests pass (0 failures)
- [ ] 1.1.11 Coverage 98% (infrastructure code requirement)
- [ ] 1.1.12 Build clean: `go build ./cmd/cicd/`
- [ ] 1.1.13 Linting clean: `golangci-lint run ./internal/cmd/cicd/format_go/`
- [ ] 1.1.14 Commit: "feat(cicd): add time.Now().UTC() enforcement formatter"

**Files**:
- Modified: `internal/cmd/cicd/format_go/format_go.go`
- Created: `internal/cmd/cicd/format_go/enforce_time_now_utc.go`
- Created: `internal/cmd/cicd/format_go/enforce_time_now_utc_test.go`

**Reference Implementation**: See `internal/cmd/cicd/format_go/enforce_any.go` for similar AST-based formatter pattern.

---

### 1.2: Add Pre-Commit Hook Integration

**Owner**: LLM Agent
**Estimated**: 1h
**Dependencies**: 1.1
**Priority**: P0 (Critical - enforcement mechanism)

**Description**:
Update pre-commit hooks configuration to run time.Now().UTC() formatter automatically before commits.

**Acceptance Criteria**:
- [ ] 1.2.1 Update `.pre-commit-config.yaml` to include format-go in `go-formatters` hook
- [ ] 1.2.2 Verify hook runs on `.go` files containing `time.Now()`
- [ ] 1.2.3 Test hook prevents commit of files with `time.Now()` (without UTC)
- [ ] 1.2.4 Test hook allows commit after auto-fix to `time.Now().UTC()`
- [ ] 1.2.5 Update `docs/pre-commit-hooks.md` with new formatter documentation
- [ ] 1.2.6 Run pre-commit test: Create file with `time.Now()`, attempt commit, verify auto-fix
- [ ] 1.2.7 Commit: "ci(hooks): add time.Now().UTC() formatter to pre-commit"

**Files**:
- Modified: `.pre-commit-config.yaml`
- Modified: `docs/pre-commit-hooks.md`

---

### 1.3: Update Copilot Instructions

**Owner**: LLM Agent
**Estimated**: 30m
**Dependencies**: 1.1, 1.2
**Priority**: P1 (Important - documentation)

**Description**:
Document the new formatter in copilot instructions as a defensive measure against LLM agent mistakes.

**Acceptance Criteria**:
- [ ] 1.3.1 Add section to `.github/instructions/03-02.testing.instructions.md`:
  - Anti-pattern: Using `time.Now()` without `.UTC()` in GORM/SQLite contexts
  - Solution: Formatter enforces `.UTC()` automatically
  - Rationale: Prevents timezone-related test failures
- [ ] 1.3.2 Add reference to formatter in SQLite DateTime UTC Comparison section
- [ ] 1.3.3 Commit: "docs(instructions): document time.Now().UTC() formatter"

**Files**:
- Modified: `.github/instructions/03-02.testing.instructions.md`

---

## Phase 2: Healthcheck Testing with app.Test() Pattern (P3.2 Resolution)

**Purpose**: Implement healthcheck timeout tests without starting actual HTTPS listeners

**Root Cause Analysis**:
- **Original Problem**: ApplicationCore bundles admin server internally, tests can't control healthcheck timing
- **Original Approach (WRONG)**: Extract admin server to start/stop in tests → violates TestMain pattern
- **Real Problem**: Starting/stopping HTTPS listeners repeatedly triggers Windows Firewall warnings
- **Best Practice Solution**: Use Fiber's `app.Test()` to test handlers WITHOUT starting listeners

**Evidence**:
- `internal/ca/api/handler/*_test.go` - Extensive use of `app.Test(req)` for handler testing
- `internal/apps/template/service/server/apis/sessions_test.go` - Handler tests without server start
- `internal/identity/authz/swagger_test.go` - Multiple invocations of handlers via `app.Test()`

**From V2 Priority 5 (P3.2 Resolution) - REVISED APPROACH**

---

### 2.1: Implement Healthcheck Handler Tests with app.Test()

**Owner**: LLM Agent
**Estimated**: 2h
**Dependencies**: None
**Priority**: P0 (Critical)

**Description**:
Replace skipped healthcheck timeout tests with `app.Test()` pattern that doesn't require starting HTTPS listeners.

**Acceptance Criteria**:
- [ ] 2.1.1 Remove t.Skip() from TestHealthcheck_CompletesWithinTimeout
- [ ] 2.1.2 Remove t.Skip() from TestHealthcheck_TimeoutExceeded
- [ ] 2.1.3 Create standalone Fiber app with healthcheck handler registered
- [ ] 2.1.4 Implement timeout completion test:
  - [ ] 2.1.4.1 Create HTTP client with 5s timeout
  - [ ] 2.1.4.2 Use `app.Test(req)` to invoke healthcheck handler
  - [ ] 2.1.4.3 Verify response received within timeout (<1s actual)
  - [ ] 2.1.4.4 Assert 200 OK status
- [ ] 2.1.5 Implement timeout exceeded test:
  - [ ] 2.1.5.1 Create HTTP client with 100ms timeout
  - [ ] 2.1.5.2 Add artificial delay to healthcheck handler (200ms)
  - [ ] 2.1.5.3 Use `app.Test(req, 100*time.Millisecond)` to enforce timeout
  - [ ] 2.1.5.4 Verify timeout error returned
- [ ] 2.1.6 Add table-driven test structure for multiple timeout scenarios
- [ ] 2.1.7 Run tests: `go test -v -run="TestHealthcheck" ./internal/apps/template/service/server/application/... -count=5`
- [ ] 2.1.8 All tests pass (0 skips, 0 flakiness)
- [ ] 2.1.9 Test execution <5s (no HTTPS listener overhead)
- [ ] 2.1.10 Coverage ≥95% on healthcheck handler code
- [ ] 2.1.11 Build clean: `go build ./internal/apps/template/...`
- [ ] 2.1.12 Linting clean: `golangci-lint run ./internal/apps/template/service/server/application/`
- [ ] 2.1.13 Commit: "test(application): implement healthcheck tests with app.Test() pattern"

**Files**:
- Modified: `internal/apps/template/service/server/application/application_listener_test.go`

**Reference Implementation**:
See `internal/ca/api/handler/handler_est_csrattrs_test.go` for `app.Test(req)` pattern example.

**Pattern Example**:
```go
func TestHealthcheck_CompletesWithinTimeout(t *testing.T) {
    t.Parallel()

    app := fiber.New()

    // Register healthcheck handler
    app.Get("/admin/api/v1/livez", func(c *fiber.Ctx) error {
        return c.Status(200).SendString("OK")
    })

    req := httptest.NewRequest("GET", "/admin/api/v1/livez", nil)

    // Test without starting listener - uses app.Test()
    resp, err := app.Test(req, 5*time.Second)
    require.NoError(t, err)
    require.Equal(t, 200, resp.StatusCode)

    defer resp.Body.Close()
}
```

---

### 2.2: Update Copilot Instructions

**Owner**: LLM Agent
**Estimated**: 30m
**Dependencies**: 2.1
**Priority**: P1 (Important - documentation)

**Description**:
Document app.Test() pattern as best practice for handler testing without HTTPS listeners.

**Acceptance Criteria**:
- [ ] 2.2.1 Add section to `.github/instructions/03-02.testing.instructions.md`:
  - Anti-pattern: Starting/stopping HTTPS listeners in tests (Windows Firewall warnings)
  - Solution: Use `app.Test(req)` for handler testing without listeners
  - Rationale: Avoids firewall prompts, faster execution, respects TestMain pattern
- [ ] 2.2.2 Add reference examples from CA handler tests
- [ ] 2.2.3 Update TestMain section to emphasize app.Test() for handler-level tests
- [ ] 2.2.4 Commit: "docs(instructions): document app.Test() pattern for handler testing"

**Files**:
- Modified: `.github/instructions/03-02.testing.instructions.md`

---

## Phase 3: Template Service E2E Verification (P3.3 Resolution)

**Purpose**: Verify template service E2E test infrastructure actually used

**From V2 Priority 6 (P3.3 Resolution)**

---

### 3.1: Verify E2E Test Existence

**Owner**: LLM Agent
**Estimated**: 30m
**Dependencies**: None
**Priority**: P0

**Description**: Check if internal/apps/template/testing/e2e/ exists with functional tests.

**Acceptance Criteria**:
- [ ] 3.1.1 Check `internal/apps/template/testing/e2e/` directory
- [ ] 3.1.2 Verify testmain_e2e_test.go exists
- [ ] 3.1.3 Verify uses docker_health.go helpers
- [ ] 3.1.4 Verify template service in healthcheck list
- [ ] 3.1.5 Document findings
- [ ] 3.1.6 Proceed to 3.2 if missing
- [ ] 3.1.7 Commit: "docs(tasks): verify template E2E status"

---

### 3.2: Create E2E Tests (if missing)

**Owner**: LLM Agent
**Estimated**: 2h
**Dependencies**: 3.1
**Priority**: P0

**Description**: Create template E2E tests using existing infrastructure patterns.

**Acceptance Criteria**:
- [ ] 3.2.1 Create `internal/apps/template/testing/e2e/` directory
- [ ] 3.2.2 Create testmain_e2e_test.go with TestMain healthcheck
- [ ] 3.2.3 Import ComposeManager from compose.go
- [ ] 3.2.4 Configure healthcheck URLs (admin :9090/livez, public :8080/swagger)
- [ ] 3.2.5 Add template to dockerComposeServicesForHealthCheck
- [ ] 3.2.6 Create docker-compose-template-e2e.yml if missing
- [ ] 3.2.7 Implement test cases (healthcheck, public endpoint)
- [ ] 3.2.8 Run: `go test -tags=e2e -v ./internal/apps/template/testing/e2e/...`
- [ ] 3.2.9 All tests pass
- [ ] 3.2.10 Test execution <2min
- [ ] 3.2.11 Build clean
- [ ] 3.2.12 Commit: "test(template): add E2E tests"

**Files**:
- Created: `internal/apps/template/testing/e2e/testmain_e2e_test.go`
- Created: `internal/apps/template/testing/e2e/template_e2e_test.go`
- Created: `deployments/template/docker-compose-template-e2e.yml`
- Modified: `internal/test/e2e/docker_health.go`

---

### 3.3: Update P3.3 Status

**Owner**: LLM Agent
**Estimated**: 15m
**Dependencies**: 3.2
**Priority**: P0

**Acceptance Criteria**:
- [ ] 3.3.1 Mark P3.3 complete with verification
- [ ] 3.3.2 Document E2E verification
- [ ] 3.3.3 Include test output or confirmation
- [ ] 3.3.4 Commit: "docs(tasks): mark P3.3 complete"

---

## Phase 4: High Coverage Testing

**Purpose**: Achieve 95% coverage for all packages (cipher-im, JOSE-JA, service-template, KMS)

**From V1 Phase X**

---

### 4.1: Increase cipher-im Repository Coverage

**Target**: 95% (current: 89.2%, gap: 5.8%)
**Estimated**: 2h

**Missing Test Cases**:
- [ ] 4.1.1 Database connection failures
- [ ] 4.1.2 GORM error path handling
- [ ] 4.1.3 Concurrent repository access patterns
- [ ] 4.1.4 Repository transaction rollback scenarios

---

### 4.2: JOSE-JA Coverage Verification

**Target**: 95% (current: 96.3% aggregate - verify per-package)
**Estimated**: 3h

**Package-Level Analysis**:
- [ ] 4.2.1 Review jose/domain coverage (verify 95%)
- [ ] 4.2.2 Review jose/repository coverage (verify 95%)
- [ ] 4.2.3 Review jose/service coverage (verify 95%)
- [ ] 4.2.4 Review jose/apis coverage (verify 95%)
- [ ] 4.2.5 Review jose/server coverage (verify 95%)

---

### 4.3: Identify Remaining Coverage Gaps

**Objective**: Analyze ALL packages for specific uncovered lines
**Estimated**: 4h

**Process**:
- [ ] 4.3.1 Generate HTML coverage reports for ALL packages
- [ ] 4.3.2 Identify RED lines (uncovered)
- [ ] 4.3.3 Categorize gaps (error paths, edge cases, concurrency, validation)
- [ ] 4.3.4 Document findings in coverage-gaps.md
- [ ] 4.3.5 Create targeted tasks

---

### 4.4: Implement Targeted Coverage Tests

**Objective**: Write tests for RED lines from 4.3
**Dependencies**: 4.3
**Estimated**: 8h

**Process**:
- [ ] 4.4.1 Implement tests for error paths
- [ ] 4.4.2 Implement tests for edge cases
- [ ] 4.4.3 Implement tests for concurrency gaps
- [ ] 4.4.4 Implement tests for validation gaps
- [ ] 4.4.5 Re-run coverage reports
- [ ] 4.4.6 Validate 95% target met

---

## Phase 5: Resolve Remaining Blockers

**Purpose**: Complete JOSE services coverage and Phase 4 validation

**From V1 Phase Y.4-Y.5**

---

### 5.1: JOSE Services Coverage

**Objective**: Increase JOSE service layer coverage to 95%
**Current**: JWK ~85%, Registration ~90%, Rotation ~88%
**Estimated**: 6h

**Missing Test Cases**:
- [ ] 5.1.1 JWK service error paths (database failures, validation)
- [ ] 5.1.2 Registration service edge cases (duplicate keys, invalid algorithms)
- [ ] 5.1.3 Rotation service concurrency (parallel requests)
- [ ] 5.1.4 Service transaction rollback scenarios
- [ ] 5.1.5 Service-layer validation logic

---

### 5.2: Phase 4 Validation and Completion

**Objective**: Ensure 95% coverage verified across ALL packages
**Dependencies**: 4.1-4.4, 5.1
**Estimated**: 2h

**Process**:
- [ ] 5.2.1 Run comprehensive coverage: `go test -coverprofile=coverage.out ./...`
- [ ] 5.2.2 Generate HTML: `go tool cover -html=coverage.out -o coverage.html`
- [ ] 5.2.3 Review ALL packages for 95% target
- [ ] 5.2.4 Identify remaining gaps
- [ ] 5.2.5 Write coverage-report.md
- [ ] 5.2.6 Mark Phase 4 complete

---

## Phase 6: Mutation Testing

**Purpose**: Achieve 85% gremlins efficacy for all packages

**From V1 Phase Z - BLOCKED ON PHASES 4-5 COMPLETION**

**Prerequisites**: Phase 4 complete (95% baseline coverage)

---

### 6.1: Run Mutation Testing Baseline

**Estimated**: 2h

**Process**:
- [ ] 6.1.1 Run gremlins on cipher-im: `gremlins unleash ./internal/apps/cipher/im/...`
- [ ] 6.1.2 Run gremlins on JOSE-JA: `gremlins unleash ./internal/apps/jose/ja/...`
- [ ] 6.1.3 Run gremlins on service-template: `gremlins unleash ./internal/apps/template/...`
- [ ] 6.1.4 Run gremlins on KMS: `gremlins unleash ./internal/kms/...`
- [ ] 6.1.5 Document baseline efficacy scores

---

### 6.2: Analyze Mutation Results

**Estimated**: 3h

**Process**:
- [ ] 6.2.1 Identify survived mutants
- [ ] 6.2.2 Categorize survival reasons (weak assertions, missing edge cases)
- [ ] 6.2.3 Document in mutation-gaps.md
- [ ] 6.2.4 Create targeted tasks

---

### 6.3: Implement Mutation-Killing Tests

**Estimated**: 8h

**Process**:
- [ ] 6.3.1 Write tests for arithmetic operator mutations
- [ ] 6.3.2 Write tests for conditional boundary mutations
- [ ] 6.3.3 Write tests for logical operator mutations
- [ ] 6.3.4 Write tests for increment/decrement mutations
- [ ] 6.3.5 Verify 85% efficacy for ALL packages

---

### 6.4: Continuous Mutation Testing

**Estimated**: 2h

**Process**:
- [ ] 6.4.1 Add gremlins to CI/CD (run on merge)
- [ ] 6.4.2 Configure timeout (15min per package)
- [ ] 6.4.3 Set efficacy threshold (85% required)
- [ ] 6.4.4 Document in README.md

---

## Final Project Validation

### Pre-Merge Checklist

**Code Quality**:
- [x] All linting passes
- [x] All tests pass
- [ ] Coverage 95% ALL packages (Phase 4 incomplete)
- [ ] Mutation 85% efficacy (Phase 6 not started)
- [ ] No new TODOs without tracking
- [x] Build clean

**Testing**:
- [x] Unit tests comprehensive
- [ ] Integration tests functional (Phase 4 gaps)
- [ ] E2E tests passing (4 failures - healthcheck timing)
- [ ] Benchmarks functional
- [x] Property tests applicable

**Documentation**:
- [x] README.md updated
- [x] API documentation generated
- [x] Architecture docs current
- [x] Lessons learned documented

**CI/CD**:
- [x] All workflows passing
- [ ] Coverage reports (Phase 4 pending)
- [ ] Mutation testing (Phase 6 not started)
- [ ] E2E workflows (4 failures)

### Known Issues

**4 E2E Test Failures** (healthcheck timing):
1. TestJoseServer_E2E_HealthCheck
2. TestCipherServer_E2E_HealthCheck
3. TestTemplateServer_E2E_HealthCheck
4. TestKMSServer_E2E_HealthCheck

**Root Cause**: Polling starts before Docker Compose fully initialized
**Solution**: Increase timeout 30s60s, exponential backoff (1s, 2s, 4s, 8s, 16s)
**Estimated**: 1h

### Timeline Summary

- Phase 1: 5.5h (CI/CD formatter)
- Phase 2-3: 7h (V2 ApplicationCore + E2E)
- Phase 4-5: 23h (V1 coverage completion)
- Phase 6: 15h (V1 mutation testing)
- E2E fixes: 1h
- **Total Remaining: 51.5 hours (~7-10 days)**

---

**Summary**: 219 of 295 tasks complete (74%). Phases 0-3, 9, W complete. Phases 1-6 represent unified remaining work from V1+V2.
