# Tasks - Test Coverage Implementation Plan (V2)

**Status**: 23 of 70 tasks complete (33%)
**Last Updated**: 2026-01-25

---

## Priority 5 (P3.2 Resolution - ApplicationCore Refactoring for Healthcheck Testing)

**Purpose**: Resolve P3.2 skip by refactoring ApplicationCore to enable standalone admin server testing

**Context**: P3.2 SKIPPED because ApplicationCore builder pattern starts admin server internally.
Testing healthcheck timeout requires standalone admin server initialization.

**Solution**: Extract admin server creation from ApplicationCore, create NewAdminServer() constructor.

---

### P5.1: Extract Admin Server from ApplicationCore

**Owner**: LLM Agent
**Estimated**: 3h
**Dependencies**: None
**Priority**: P1 (Critical - unblocks P3.2)

**Description**:
Refactor ApplicationCore to extract admin server initialization into standalone NewAdminServer() constructor.
Maintain backward compatibility with existing ApplicationCore.Build() API.

**Acceptance Criteria**:
- [ ] 5.1.1 Create NewAdminServer(settings AdminServerSettings) (*AdminServer, error) function
- [ ] 5.1.2 Extract admin server initialization logic from ApplicationCore.Build()
- [ ] 5.1.3 Update ApplicationCore.Build() to call NewAdminServer() internally
- [ ] 5.1.4 Add AdminServerSettings struct with required configuration
- [ ] 5.1.5 Add unit tests for NewAdminServer() standalone initialization
- [ ] 5.1.6 Verify existing ApplicationCore consumers still work (no breaking changes)
- [ ] 5.1.7 Run existing application tests: `go test ./internal/apps/template/service/server/application/... -v`
- [ ] 5.1.8 All tests pass (0 failures)
- [ ] 5.1.9 Coverage maintained: 95% for application package
- [ ] 5.1.10 Build clean: `go build ./...`
- [ ] 5.1.11 Linting clean: `golangci-lint run ./internal/apps/template/service/server/application/`
- [ ] 5.1.12 Commit with evidence: "refactor(application): extract NewAdminServer for testability"

**Files**:
- Modified: `internal/apps/template/service/server/application/application_core.go`
- Created: `internal/apps/template/service/server/listener/admin_server.go`
- Created: `internal/apps/template/service/server/listener/admin_server_test.go`

---

### P5.2: Implement Healthcheck Timeout Tests

**Owner**: LLM Agent
**Estimated**: 1h
**Dependencies**: P5.1 (requires NewAdminServer)
**Priority**: P1 (Critical)

**Description**:
Remove t.Skip() from healthcheck timeout tests and implement actual timeout testing logic
using standalone NewAdminServer() constructor.

**Acceptance Criteria**:
- [ ] 5.2.1 Remove t.Skip() from TestHealthcheck_CompletesWithinTimeout
- [ ] 5.2.2 Remove t.Skip() from TestHealthcheck_TimeoutExceeded
- [ ] 5.2.3 Implement TestHealthcheck_CompletesWithinTimeout:
  - Create admin server with NewAdminServer()
  - Set client timeout = 5s
  - Call /admin/api/v1/livez
  - Verify response within timeout (< 1s typical)
- [ ] 5.2.4 Implement TestHealthcheck_TimeoutExceeded:
  - Create admin server with NewAdminServer()
  - Add artificial delay in healthcheck handler (6s)
  - Set client timeout = 5s
  - Verify timeout error returned
- [ ] 5.2.5 Run tests: `go test -v -run="TestHealthcheck" ./internal/apps/template/service/server/application/`
- [ ] 5.2.6 Both tests pass (0 failures, 0 skips)
- [ ] 5.2.7 Test execution <30 seconds
- [ ] 5.2.8 Commit with evidence: "test(application): implement healthcheck timeout tests"

**Files**:
- Modified: `internal/apps/template/service/server/application/application_listener_test.go`

---

### P5.3: Update P3.2 Status to Complete

**Owner**: LLM Agent
**Estimated**: 15m
**Dependencies**: P5.2 (timeout tests working)
**Priority**: P1 (Critical)

**Description**:
Mark P3.2 task as complete with test results, remove skip status.

**Acceptance Criteria**:
- [ ] 5.3.1 Update P3.2 section to mark  COMPLETE with test output
- [ ] 5.3.2 Document ApplicationCore refactoring resolution in P3.2 skip analysis
- [ ] 5.3.3 Include test execution output in P3.2 verification evidence
- [ ] 5.3.4 Git commit: "docs(tasks): mark P3.2 complete - timeout tests working"

**Files**:
- Modified: `docs/fixes-needed-plan-tasks-v2/tasks.md`

---

## Priority 6 (P3.3 Resolution - Template Service E2E Verification)

**Purpose**: Resolve P3.3 "satisfied by existing" by verifying template service actually uses E2E infrastructure

**Context**: P3.3 marked "SATISFIED BY EXISTING TESTS" but did NOT verify template service has E2E tests.
Existing infrastructure documented (docker_health.go, infrastructure.go) but template service usage unverified.

**Solution**: Check if template service has E2E tests, create if missing, verify uses existing helpers.

---

### P6.1: Verify Template Service E2E Test Existence

**Owner**: LLM Agent
**Estimated**: 30m
**Dependencies**: None
**Priority**: P1 (Critical - validates P3.3 claim)

**Description**:
Check if internal/apps/template/testing/e2e/ directory exists with functional E2E tests.
Document findings and create tests if missing.

**Acceptance Criteria**:
- [ ] 6.1.1 Check if `internal/apps/template/testing/e2e/` directory exists
- [ ] 6.1.2 If exists, verify contains testmain_e2e_test.go with TestMain pattern
- [ ] 6.1.3 If exists, verify uses docker_health.go (ServiceAndJob) or ComposeManager helpers
- [ ] 6.1.4 If exists, verify template service in dockerComposeServicesForHealthCheck list
- [ ] 6.1.5 If missing, document gap and proceed to P6.2
- [ ] 6.1.6 Document findings in P3.3 verification section
- [ ] 6.1.7 Commit with evidence: "docs(tasks): verify template service E2E test status"

**Files**:
- Modified: `docs/fixes-needed-plan-tasks-v2/tasks.md`

---

### P6.2: Create Template Service E2E Tests (if missing)

**Owner**: LLM Agent
**Estimated**: 2h
**Dependencies**: P6.1 (gap identified)
**Priority**: P1 (Critical)

**Description**:
If template service E2E tests missing, create testmain_e2e_test.go using existing infrastructure patterns.
Reuse ComposeManager, InfrastructureManager, and docker_health.go helpers.

**Acceptance Criteria**:
- [ ] 6.2.1 Create `internal/apps/template/testing/e2e/` directory (if missing)
- [ ] 6.2.2 Create testmain_e2e_test.go with TestMain healthcheck flow
- [ ] 6.2.3 Import and use ComposeManager from internal/apps/template/testing/e2e/compose.go
- [ ] 6.2.4 Configure template service healthcheck URLs (admin: :9090/admin/api/v1/livez, public: :8080/ui/swagger/doc.json)
- [ ] 6.2.5 Add template service to dockerComposeServicesForHealthCheck list in docker_health.go
- [ ] 6.2.6 Create docker-compose-template-e2e.yml if missing
- [ ] 6.2.7 Implement test cases:
  - TestTemplateService_Healthcheck (verify admin server healthy)
  - TestTemplateService_PublicEndpoint (verify public server reachable)
- [ ] 6.2.8 Run E2E tests: `go test -tags=e2e -v ./internal/apps/template/testing/e2e/...`
- [ ] 6.2.9 All tests pass (0 failures)
- [ ] 6.2.10 Test execution <2 minutes (90s healthcheck timeout)
- [ ] 6.2.11 Build clean: `go build ./...`
- [ ] 6.2.12 Commit with evidence: "test(template): add E2E tests using existing infrastructure"

**Files**:
- Created: `internal/apps/template/testing/e2e/testmain_e2e_test.go`
- Created: `internal/apps/template/testing/e2e/template_e2e_test.go`
- Created: `deployments/template/docker-compose-template-e2e.yml` (if needed)
- Modified: `internal/test/e2e/docker_health.go` (add template to service list)

---

### P6.3: Update P3.3 Status to Complete

**Owner**: LLM Agent
**Estimated**: 15m
**Dependencies**: P6.2 (E2E tests verified/created)
**Priority**: P1 (Critical)

**Description**:
Mark P3.3 task as complete with verification evidence or new test results.

**Acceptance Criteria**:
- [ ] 6.3.1 Update P3.3 section to mark  COMPLETE with verification evidence
- [ ] 6.3.2 Document E2E test verification in P3.3 analysis section
- [ ] 6.3.3 Include test execution output or "already exists" confirmation
- [ ] 6.3.4 Git commit: "docs(tasks): mark P3.3 complete - E2E tests verified"

**Files**:
- Modified: `docs/fixes-needed-plan-tasks-v2/tasks.md`

---

## Success Criteria (Overall)

### Coverage Targets

- mTLS configuration logic: 95% coverage ( ACHIEVED)
- Container mode detection: 95% coverage ( ACHIEVED)
- Config validation: 95% coverage ( ACHIEVED)
- YAML field mapping: 95% coverage ( ACHIEVED)
- Database URL mapping: 98% coverage ( ACHIEVED)

### Quality Gates

- All P1 tests implemented and passing
- All P2 tests implemented and passing
- P3.1 tests implemented and passing
- P3.2 tests pending (ApplicationCore refactoring required)
- P3.3 tests pending (E2E verification required)
- [ ] Mutation testing 85% efficacy on affected modules
- [ ] No new TODOs or FIXMEs in test files
- All unit tests run in <15 seconds per package
- All integration tests run in <30 seconds per package

### Workflow Impact

- DAST workflow passes consistently (no config failures)
- Load testing workflow passes consistently
- E2E workflows pass with container mode
- No regression in existing tests

### Service Template Reusability

- 8 of 11 test tasks are service-template tests (reusable across 9 services)
- KMS-specific tests: 3 tasks (database URL mapping, integration tests)
- Total test coverage increase: ~100 new test cases across all services

---

**Note**: P1 (5/5), P2 (3/3), P3.1 (1/1) complete with 23 tasks. P3.2 (3 tasks) and P3.3 (3 tasks) pending - 6 incomplete tasks remaining (9%).
