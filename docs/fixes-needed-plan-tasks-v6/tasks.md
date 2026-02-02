# Tasks - Service Template & CICD Fixes

**Status**: 46/50 tasks complete (92%) | 1 BLOCKED | Phase 11-12 remaining
**Last Updated**: 2026-02-01

## Summary

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1-4 | ✅ Complete | Instructions, CICD, Deployment, Critical Fixes |
| Phase 5 | ⚠️ Partial | Test Architecture (5.1 BLOCKED) |
| Phase 6-8 | ✅ Complete | Coverage, Cleanup, Race Detection |
| Phase 9 | ✅ Complete | KMS Analysis (ServerBuilder extension needed) |
| Phase 10 | ✅ Complete | Cleanup (10.1-10.4 ✅) |
| Phase 11 | ❌ Not Started | KMS ServerBuilder Extension (REQUIRED) |
| Phase 12 | ❌ Not Started | KMS Before/After Comparison (REQUIRED) |

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
- **Next Steps**:
  1. Complete `StartApplicationListener` implementation first
  2. THEN refactor tests to use app.Test() pattern
- **Acceptance Criteria**:
  - [ ] Blocked until StartApplicationListener implemented
  - [ ] No Windows Firewall triggers
  - [ ] No port binding in unit tests
  - [ ] Tests run faster (<1ms vs 10-50ms)
  - [ ] All tests still pass

---

### Phase 10: Cleanup Leftover Coverage Files (REQUIRED)

**Status**: ✅ Complete (All 10.1-10.4 Complete)
**Discovery Date**: 2026-02-01
**Issue**: LLM autonomous work left 57+ coverage files in project root and internal directories

**User Decisions (from quizme-v1.md)**:
- Delete ALL files in test-output/ (clean slate)
- Detect leftover files in ALL directories including test-output/
- Patterns: `*.out`, `*.cov`, `*.prof`, `*coverage*.html`, `*coverage*.txt`
- Auto-delete files if found, with warning message

#### Task 10.1: Delete Root-Level Coverage Files
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.1h
- **Description**: Delete all .out and coverage files in project root directory
- **Evidence**: Commit `a07e9175` - 35 root-level files deleted
- **Acceptance Criteria**:
  - [x] All root-level .out files deleted
  - [x] All root-level coverage.html files deleted
  - [x] git status shows deletions

#### Task 10.2: Delete Internal Directory Coverage Files
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.1h
- **Description**: Delete all .out and coverage files inside internal/ directories
- **Evidence**: Commit `a07e9175` - 6 internal/ files deleted
- **Acceptance Criteria**:
  - [x] All internal/ .out files deleted
  - [x] All internal/ coverage.html files deleted
  - [x] git status shows deletions

#### Task 10.3: Delete ALL Files in test-output/
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.1h
- **Description**: Per user decision, delete ALL files in test-output/ for clean slate
- **Evidence**: Commit `a07e9175` - 202+ files deleted, test-output/ removed
- **Acceptance Criteria**:
  - [x] test-output/ directory is empty (or deleted)
  - [x] git status shows deletions

#### Task 10.4: Add CICD Linter for Leftover Coverage Files
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 45m
- **File**: `internal/cmd/cicd/lint_go/leftover_coverage.go`
- **Description**: Add cicd check to auto-delete leftover coverage files with warning
- **User Decisions**:
  - Scope: ALL directories including test-output/
  - Patterns (configurable list): `*.out`, `*.cov`, `*.prof`, `*coverage*.html`, `*coverage*.txt`
  - Behavior: Auto-delete files if found, with warning message
- **Evidence**: Commit `6f4399ef` - leftover_coverage.go and tests added, 17+ more files deleted
- **Acceptance Criteria**:
  - [x] Detects patterns in ALL directories
  - [x] Auto-deletes files found
  - [x] Prints warning message for each deleted file
  - [x] Pattern list easily updatable in code
  - [x] Integrated into cicd lint-go command
  - [x] Tests pass

---

### Phase 11: KMS ServerBuilder Extension (REQUIRED)

**Status**: ❌ Not Started
**Rationale**: Service-template MUST support all KMS functionality for lateral migration

**CRITICAL**: KMS migrating to service-template MUST be a lateral move - no loss of functionality, architecture, design intent, or test intent.

#### Task 11.1: Extend ServerBuilder with SwaggerUI Support
- **Status**: ❌ Not Started
- **Estimated**: 4h
- **Description**: Add `WithSwaggerUI(username, password string)` method to ServerBuilder
- **Acceptance Criteria**:
  - [ ] ServerBuilder supports Swagger UI
  - [ ] Basic auth middleware included
  - [ ] CSRF script injection supported
  - [ ] Tests pass

#### Task 11.2: Extend ServerBuilder with OpenAPI Handler Registration
- **Status**: ❌ Not Started
- **Estimated**: 4h
- **Description**: Add `WithOpenAPIHandlers(strictServer interface{})` method to ServerBuilder
- **Acceptance Criteria**:
  - [ ] ServerBuilder supports oapi-codegen generated handlers
  - [ ] Request validation middleware included
  - [ ] Tests pass

#### Task 11.3: Extend ServerBuilder with Security Headers
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Description**: Add comprehensive security headers to ServerBuilder
- **Acceptance Criteria**:
  - [ ] CSP headers configurable
  - [ ] XSS protection included
  - [ ] HSTS configured
  - [ ] Tests pass

#### Task 11.4: Migrate KMS to Extended ServerBuilder
- **Status**: ❌ Not Started
- **Estimated**: 4h
- **Description**: After 11.1-11.3 complete, migrate KMS to use extended ServerBuilder
- **Dependencies**: 11.1, 11.2, 11.3
- **Acceptance Criteria**:
  - [ ] KMS uses extended ServerBuilder
  - [ ] application_listener.go deleted
  - [ ] All KMS tests pass
  - [ ] All template tests pass
  - [ ] All cipher-im tests pass

---

### Phase 12: KMS Before/After Comparison (REQUIRED)

**Status**: ❌ Not Started
**Rationale**: Verify service-template reproduces ALL KMS functionality

**CRITICAL**: Service-template is the foundation of ALL 9 services. KMS switching to service-template MUST be a lateral move:
- No loss of functionality
- No divergence from design intent
- No loss of test coverage or test intent
- Maximum reusability
- Architecture preserved

#### Task 12.1: Document KMS Current Architecture
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Description**: Comprehensive documentation of KMS before migration
- **Output**: `test-output/kms-comparison/kms-before-architecture.md`
- **Acceptance Criteria**:
  - [ ] All endpoints documented
  - [ ] All middleware documented
  - [ ] All security features documented
  - [ ] Configuration options documented
  - [ ] Database schema documented

#### Task 12.2: Document KMS Current Test Coverage and Intent
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Description**: Document all KMS tests, their purpose, and coverage
- **Output**: `test-output/kms-comparison/kms-before-tests.md`
- **Acceptance Criteria**:
  - [ ] All test files listed
  - [ ] Test intent documented
  - [ ] Coverage metrics captured
  - [ ] Critical test cases identified

#### Task 12.3: Compare KMS vs Service-Template Feature Parity
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Description**: Side-by-side comparison of features
- **Output**: `test-output/kms-comparison/feature-parity.md`
- **Acceptance Criteria**:
  - [ ] Feature matrix complete
  - [ ] Gaps identified
  - [ ] Parity verified for all critical features

#### Task 12.4: Verify All KMS Tests Pass with Service-Template Backend
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Description**: Run all KMS tests after migration, verify no regressions
- **Acceptance Criteria**:
  - [ ] All KMS unit tests pass
  - [ ] All KMS integration tests pass
  - [ ] All KMS E2E tests pass
  - [ ] No new test failures
  - [ ] Coverage maintained or improved

#### Task 12.5: Document Intentional Differences and Rationale
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Description**: If any differences exist, document why they're intentional
- **Output**: `test-output/kms-comparison/intentional-differences.md`
- **Acceptance Criteria**:
  - [ ] All differences documented
  - [ ] Rationale for each difference
  - [ ] User approval for any intentional changes

#### Task 12.6: Create Final Comparison Report
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Description**: Comprehensive before/after comparison report
- **Output**: `test-output/kms-comparison/final-report.md`
- **Acceptance Criteria**:
  - [ ] Executive summary
  - [ ] Functionality verification complete
  - [ ] Design intent preserved
  - [ ] Test intent preserved
  - [ ] No regressions
  - [ ] User sign-off ready

---

## Cross-Cutting Tasks (Remaining)

### Quality
- [ ] No TODOs in production code (some remain, tracked)

---

## References

- Completed tasks: [completed.md](./completed.md)
- Analysis docs archived in [archive/](./archive/)
- Copilot instructions in `.github/instructions/`
- Comparison table: [comparison-table.md](./comparison-table.md)
