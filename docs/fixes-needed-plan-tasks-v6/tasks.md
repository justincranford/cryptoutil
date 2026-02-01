# Tasks - Service Template & CICD Fixes

**Status**: 42/44 tasks complete (95%) | 1 BLOCKED | 1 DEFERRED | Phase 11 NEW (cleanup)
**Last Updated**: 2026-02-01

## Summary

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1-4 | ✅ Complete | Instructions, CICD, Deployment, Critical Fixes |
| Phase 5 | ⚠️ Partial | Test Architecture (5.1 BLOCKED) |
| Phase 6-8 | ✅ Complete | Coverage, Cleanup, Race Detection |
| Phase 9 | ✅ Complete | KMS Analysis (9.3 deferred to Phase 10) |
| Phase 10 | DEFERRED | Optional ServerBuilder Extension |
| Phase 11 | ❌ NEW | Cleanup Leftover Coverage Files |

**Completed tasks archived**: See [completed.md](./completed.md)

---

## Incomplete Tasks

### Phase 5: Test Architecture Refactoring

#### Task 5.1: Refactor Listener Tests to app.Test()
- **Status**: ❌ BLOCKED
- **Blocker**: `StartApplicationListener` not yet implemented (returns "implementation in progress" error)
- **Estimated**: 3h
- **Files**:
  - `internal/apps/template/service/server/listener/servers_test.go`
  - `internal/apps/template/service/server/listener/application_listener_test.go`
- **Description**: Replace real HTTPS listeners with Fiber app.Test() for in-memory testing
- **Current State**: Tests only validate constructor/factory functions, no HTTP listeners started yet
- **Next Steps**:
  1. Complete `StartApplicationListener` implementation first
  2. THEN refactor tests to use app.Test() pattern
  3. Note: admin_test.go and public_test.go (1597 lines) are the actual files needing app.Test() refactoring
- **Acceptance Criteria**:
  - [ ] Blocked until StartApplicationListener implemented
  - [ ] No Windows Firewall triggers
  - [ ] No port binding in unit tests
  - [ ] Tests run faster (<1ms vs 10-50ms)
  - [ ] All tests still pass

---

### Phase 10: KMS ServerBuilder Extension (DEFERRED)

**Status**: DEFERRED - Optional future work

**Rationale**: Current KMS architecture with `application_listener.go` is correct, complete, and tested. ServerBuilder migration would provide consistency but requires significant extension work and is not blocking any production functionality.

#### Task 10.1: Extend ServerBuilder with SwaggerUI Support
- **Status**: ❌ Not Started (DEFERRED)
- **Estimated**: 4h
- **Description**: Add `WithSwaggerUI(username, password string)` method to ServerBuilder
- **Acceptance Criteria**:
  - [ ] ServerBuilder supports Swagger UI
  - [ ] Basic auth middleware included
  - [ ] CSRF script injection supported

#### Task 10.2: Extend ServerBuilder with OpenAPI Handler Registration
- **Status**: ❌ Not Started (DEFERRED)
- **Estimated**: 4h
- **Description**: Add `WithOpenAPIHandlers(strictServer interface{})` method to ServerBuilder
- **Acceptance Criteria**:
  - [ ] ServerBuilder supports oapi-codegen generated handlers
  - [ ] Request validation middleware included

#### Task 10.3: Extend ServerBuilder with Security Headers
- **Status**: ❌ Not Started (DEFERRED)
- **Estimated**: 2h
- **Description**: Add comprehensive security headers to ServerBuilder
- **Acceptance Criteria**:
  - [ ] CSP headers configurable
  - [ ] XSS protection included
  - [ ] HSTS configured

#### Task 10.4: Migrate KMS to Extended ServerBuilder
- **Status**: ❌ Not Started (DEFERRED)
- **Estimated**: 4h
- **Description**: After 10.1-10.3 complete, migrate KMS to use extended ServerBuilder
- **Dependencies**: 10.1, 10.2, 10.3
- **Acceptance Criteria**:
  - [ ] KMS uses extended ServerBuilder
  - [ ] application_listener.go deleted
  - [ ] All tests pass

---

### Phase 11: Cleanup Leftover Coverage Files (NEW)

**Status**: ❌ Not Started
**Discovery Date**: 2026-02-01
**Issue**: LLM autonomous work left 57 coverage files in project root and internal directories

#### Task 11.1: Delete Root-Level Coverage Files
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Description**: Delete all .out and coverage files in project root directory
- **Files to Delete** (17 files):
  ```
  ./jose_apis.out
  ./cipher_coverage.out
  ./barrier_coverage.html
  ./jose_service.out
  ./cipher_repo.out
  ./coverage.out
  ./jose_server.out
  ./template_coverage.out
  ./server_cov.out
  ./template_coverage_full.out
  ./jose_coverage.out
  ./jose_service_cov.out
  ./jose_repository.out
  ./lint_workflow_coverage.html
  ./apis_cov.out
  ./jose_domain.out
  ```
- **Acceptance Criteria**:
  - [ ] All root-level .out files deleted
  - [ ] All root-level coverage.html files deleted
  - [ ] git status shows deletions

#### Task 11.2: Delete Internal Directory Coverage Files
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Description**: Delete all .out and coverage files inside internal/ directories
- **Files to Delete** (4 files):
  ```
  ./internal/apps/template/service/server/barrier/cover.out
  ./internal/apps/template/service/server/application/coverage.out
  ./internal/apps/template/service/server/builder/coverage.html
  ./internal/apps/template/service/server/builder/coverage.out
  ```
- **Acceptance Criteria**:
  - [ ] All internal/ .out files deleted
  - [ ] All internal/ coverage.html files deleted
  - [ ] git status shows deletions

#### Task 11.3: Review test-output/ Coverage Files
- **Status**: ⚠️ USER DECISION REQUIRED
- **Estimated**: 0.5h
- **Description**: Review files in test-output/ - these may be intentional analysis artifacts
- **File Count**: ~40 files in test-output/
- **Options**:
  - A) Delete all test-output/*.out and *coverage*.html files
  - B) Keep test-output/ as analysis artifact directory (already gitignored)
  - C) Keep specific analysis directories, delete others
- **Acceptance Criteria**:
  - [ ] User decision captured
  - [ ] Files deleted or retained per decision

#### Task 11.4: Add CICD Linter for Leftover Coverage Files
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/cmd/cicd/lint_go/leftover_coverage.go`
- **Description**: Add cicd check to detect leftover .out and coverage.html files outside test-output/
- **Acceptance Criteria**:
  - [ ] Detects *.out files in root and internal/
  - [ ] Detects *coverage*.html files in root and internal/
  - [ ] Allows test-output/ directory (analysis artifacts)
  - [ ] Integrated into cicd lint-go command
  - [ ] Tests pass

---

## Cross-Cutting Tasks (Remaining)

### Quality
- [ ] No TODOs in production code (some remain, tracked)

### KMS Modernization (Phase 10 - DEFERRED)
- [ ] ServerBuilder extended for KMS-style services

---

## References

- Completed tasks: [completed.md](./completed.md)
- Analysis docs archived in [archive/](./archive/)
- Copilot instructions in `.github/instructions/`
- Comparison table: [comparison-table.md](./comparison-table.md)
