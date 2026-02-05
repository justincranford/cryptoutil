# Tasks V10 - Critical Regressions and Completion Fixes

**Status**: 0 of 56 tasks complete (0%)
**Last Updated**: 2026-02-05

## Quality Mandate - MANDATORY

**ALL issues are blockers - NO exceptions:**

- ✅ **Fix issues immediately** - E2E timeouts, test failures, build errors = STOP and FIX
- ✅ **Treat as BLOCKING** - ALL issues block next task
- ✅ **Do NOT defer** - No "later", no "non-critical", no "nice-to-have"
- ❌ **NEVER skip** - Cannot mark complete with known issues
- ❌ **NEVER de-prioritize** - Quality ALWAYS highest priority

**Example of WRONG approach**: Treating cipher-im E2E timeouts as "non-blocking" was WRONG.

## Task Checklist

### Phase 0: Evidence Collection & Root Cause Analysis

#### Task 0.1: E2E Health Timeout Reproduction

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: None
- **Description**: Reproduce cipher-im E2E timeout to confirm issue exists
- **Acceptance Criteria**:
  - [ ] Run `go test ./internal/apps/cipher/im/e2e -v`
  - [ ] Confirm 90s timeout failure
  - [ ] Capture error messages and logs
  - [ ] Document: Exact failure pattern

#### Task 0.2: Multi-Service E2E Health Check Survey

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: None
- **Description**: Test E2E health checks for all services
- **Acceptance Criteria**:
  - [ ] Run E2E tests for jose-ja, sm-kms, pki-ca
  - [ ] Document: Which pass, which fail, which timeout
  - [ ] Note: Timeout durations and failure patterns
  - [ ] Create: Comparison table

#### Task 0.3: Docker Compose Health Configuration Audit

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: None
- **Description**: Audit all docker-compose.yml health check configurations
- **Acceptance Criteria**:
  - [ ] Check: cmd/cipher-im/docker-compose.yml
  - [ ] Check: deployments/*/compose*.yml
  - [ ] Document: Port, path, interval, timeout, retries, start_period
  - [ ] Identify: Inconsistencies across services

#### Task 0.4: Health Endpoint Path Comparison

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: None
- **Description**: Compare health endpoint paths across services
- **Acceptance Criteria**:
  - [ ] Check: What template base registers (both paths expected)
  - [ ] Check: What docker-compose health checks use
  - [ ] Check: What E2E tests query
  - [ ] Document: Any path mismatches

#### Task 0.5: V8 Incomplete Task Identification

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: None
- **Description**: Find the 1 incomplete V8 task (58/59 = 98%)
- **Acceptance Criteria**:
  - [ ] Read: docs/fixes-needed-plan-tasks-v8/tasks.md
  - [ ] Identify: Which task is incomplete
  - [ ] Determine: Is it truly incomplete or mislabeled?
  - [ ] Document: Task details, blocker if any

#### Task 0.6: V9 Incomplete Tasks Classification

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: None
- **Description**: List and classify all 5 incomplete V9 tasks
- **Acceptance Criteria**:
  - [ ] Read: docs/fixes-needed-plan-tasks-v9/tasks.md
  - [ ] List: All incomplete tasks
  - [ ] Classify: V10 immediate vs deferred
  - [ ] Document: Rationale for each classification

#### Task 0.7: Import Path Breakage Verification

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: None
- **Description**: Confirm internal/test/e2e/assertions.go import error
- **Acceptance Criteria**:
  - [ ] Try: `go build ./internal/test/e2e/`
  - [ ] Confirm: Compile error on missing `internal/kms/client`
  - [ ] Verify: New location `internal/apps/sm/kms/client` exists
  - [ ] Document: Exact error message

#### Task 0.8: KMS Client Import Audit

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 0.7
- **Description**: Find all files importing old KMS client path
- **Acceptance Criteria**:
  - [ ] Run: `grep -r "internal/kms/client" . --include="*.go"`
  - [ ] List: All affected files
  - [ ] Estimate: Refactoring LOE
  - [ ] Document: Scope of changes needed

#### Task 0.9: unsealkeysservice Location Verification

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: None
- **Description**: Verify unsealkeysservice location and usage
- **Acceptance Criteria**:
  - [ ] Confirm: `internal/shared/barrier/unsealkeysservice/` exists
  - [ ] Check: Template uses it correctly
  - [ ] Check: Services use it correctly
  - [ ] Document: Import pattern, why shared

#### Task 0.10: unsealkeysservice Duplication Audit

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 0.9
- **Description**: Check for duplicate unseal logic in template
- **Acceptance Criteria**:
  - [ ] Compare: Template barrier code vs shared unsealkeysservice
  - [ ] Identify: Any duplicated logic
  - [ ] Document: What's shared vs what's unique
  - [ ] Note: If duplications found, they need fixing

### Phase 1: E2E Health Timeout Root Cause & Fix

#### Task 1.1: Service Health Endpoint Audit (Enhanced with E2E Comparative Analysis)

- **Status**: \u274c Not Started
- **Owner**: LLM Agent
- **Estimated**: 1h (increased from 0.5h for deeper analysis)
- **Actual**:
- **Dependencies**: Task 0.4
- **Description**: Audit ALL services for health endpoint registration + Docker health check consistency + cmd/ structure patterns
- **Acceptance Criteria**:
  - [ ] **Dockerfile Location Audit**: cipher-im has Dockerfile in cmd/ AND deployments/ (dual location drift risk), jose-ja/sm-kms only in deployments/ (correct centralized pattern)
  - [ ] **Health Endpoint Audit**: jose-ja uses WRONG /health endpoint (should be /admin/api/v1/livez), sm-kms verify endpoint correctness, cipher-im correct but dual Dockerfile creates drift risk
  - [ ] **Health Check Tool Audit**: sm-kms may use curl (should standardize on wget), cipher-im uses wget (correct), jose-ja TBD
  - [ ] **cmd/ Structure Audit**: cipher-im 9 files (rich), jose-ja/sm-kms 1 file (minimal) - both valid but inconsistent
  - [ ] **E2E Pattern Audit**: Verify jose-ja/sm-kms E2E tests exist, use WaitForMultipleServices pattern, compare timeout configs vs cipher-im
  - [ ] **Template Verification**: Confirm template registers BOTH /admin/api/v1/livez AND /service/api/v1/health
  - [ ] **Document Findings**: Specific violations per service, fix recommendations
- **Comparative Findings** (from analysis):
  - cipher-im: Dockerfile drift risk (cmd/ vs deployments/), correct health endpoint, 180s E2E timeout (cascade dependencies), FAILS E2E
  - jose-ja: Centralized Dockerfile (correct), WRONG health endpoint /health, E2E status TBD
  - sm-kms: Centralized Dockerfile (correct), health endpoint TBD (verify), health check tool may be curl not wget, E2E status TBD
- **Fix Plan**:
  - Remove cmd/cipher-im/Dockerfile OR sync with deployments/cipher/Dockerfile.cipher
  - Fix deployments/jose/Dockerfile.jose: /health \u2192 /admin/api/v1/livez:9090
  - Standardize ALL services: wget (not curl), /admin/api/v1/livez:9090
  - Verify jose-ja/sm-kms E2E tests exist and use template patterns

#### Task 1.2: Docker Compose Health Check Standardization

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 0.3, Task 1.1
- **Description**: Define standard health check configuration
- **Acceptance Criteria**:
  - [ ] Analyze: cipher-im (fails) vs jose-ja (passes) configurations
  - [ ] Define: Standard port (9090), path (`/admin/api/v1/livez`)
  - [ ] Define: Recommended interval, timeout, retries, start_period
  - [ ] Document: Standard configuration pattern

#### Task 1.3: E2E Test Health Check Pattern Analysis

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 0.1, Task 0.4
- **Description**: Analyze E2E test health check patterns
- **Acceptance Criteria**:
  - [ ] Review: cipher-im E2E test WaitForHealth logic
  - [ ] Check: Timeout values (current vs recommended)
  - [ ] Check: Endpoint paths used (should match docker-compose)
  - [ ] Document: Test-side issues if any

#### Task 1.4: Root Cause Determination

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Tasks 1.1, 1.2, 1.3
- **Description**: Determine root cause of E2E health timeouts
- **Acceptance Criteria**:
  - [ ] Analyze: All evidence from Phase 0 and Tasks 1.1-1.3
  - [ ] Identify: Primary root cause (config vs code vs test)
  - [ ] Document: Root cause analysis with evidence
  - [ ] Update: plan.md Decision 1 with findings

#### Task 1.5: Fix cipher-im Health Checks

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 1.4
- **Description**: Apply fixes to cipher-im based on root cause
- **Acceptance Criteria**:
  - [ ] Fix: docker-compose.yml health checks (if needed)
  - [ ] Fix: E2E test patterns (if needed)
  - [ ] Fix: Service health registration (if needed)
  - [ ] Document: Changes made

#### Task 1.6: Standardize All Service Health Checks

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 1.5
- **Description**: Apply standard health check pattern to all services
- **Acceptance Criteria**:
  - [ ] Update: jose-ja, sm-kms, pki-ca, identity-* docker-compose files
  - [ ] Standardize: Port 9090, path `/admin/api/v1/livez`
  - [ ] Standardize: Timing values
  - [ ] Document: All files changed

#### Task 1.7: E2E Test Validation

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 1.6
- **Description**: Run all E2E tests to verify fixes
- **Acceptance Criteria**:
  - [ ] Test: cipher-im E2E (should pass within 60s)
  - [ ] Test: jose-ja E2E (should still pass)
  - [ ] Test: sm-kms E2E (should pass)
  - [ ] Test: pki-ca E2E (should pass)
  - [ ] Document: All test results, zero timeouts

#### Task 1.8: Document Health Timeout Lessons

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 1.7
- **Description**: Document lessons learned about E2E health timeouts
- **Acceptance Criteria**:
  - [ ] Document: Root cause analysis
  - [ ] Document: Fix approach
  - [ ] Document: Best practices for future E2E tests
  - [ ] Add: Section to plan.md

### Phase 2: Import Path Breakage Fix

#### Task 2.1: Fix internal/test/e2e/assertions.go Import

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Actual**:
- **Dependencies**: Task 0.7
- **Description**: Refactor assertions.go to use new KMS client path
- **Acceptance Criteria**:
  - [ ] Change: `cryptoutil/internal/kms/client`  `cryptoutil/internal/apps/sm/kms/client`
  - [ ] Verify: `go build ./internal/test/e2e/` succeeds
  - [ ] Document: File changed

#### Task 2.2: Audit All KMS Client Imports

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 0.8
- **Description**: Find and list all files importing old KMS client
- **Acceptance Criteria**:
  - [ ] Run: `grep -r "internal/kms/client" . --include="*.go" --exclude-dir=vendor`
  - [ ] List: All files needing refactoring
  - [ ] Verify: None missed
  - [ ] Document: Complete list

#### Task 2.3: Refactor All KMS Client Imports

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 2.2
- **Description**: Update all files to use new KMS client path
- **Acceptance Criteria**:
  - [ ] Update: All identified files
  - [ ] Verify: `go build ./...` succeeds
  - [ ] Test: Run affected tests
  - [ ] Document: All files changed

#### Task 2.4: Verify No Legacy KMS Paths

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 2.3
- **Description**: Confirm no legacy `internal/kms` imports remain
- **Acceptance Criteria**:
  - [ ] Run: `grep -r "internal/kms" . --include="*.go" --exclude-dir=vendor`
  - [ ] Verify: Only migration docs remain (allowed)
  - [ ] Verify: No compile errors
  - [ ] Document: Verification results

#### Task 2.5: E2E Tests with New Imports

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 2.4
- **Description**: Run E2E tests to verify import fixes work
- **Acceptance Criteria**:
  - [ ] Run: All E2E tests
  - [ ] Verify: No import-related runtime errors
  - [ ] Verify: Tests pass
  - [ ] Document: Test results

### Phase 3: V8 Completion

#### Task 3.1: Review V8 Incomplete Task

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Actual**:
- **Dependencies**: Task 0.5
- **Description**: Review the 1 incomplete V8 task in detail
- **Acceptance Criteria**:
  - [ ] Read: Task description and acceptance criteria
  - [ ] Determine: What work remains
  - [ ] Check: Any blockers
  - [ ] Document: Work plan

#### Task 3.2: Complete V8 Task 58/59

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 3.1
- **Description**: Complete the remaining V8 task
- **Acceptance Criteria**:
  - [ ] Implement: Required work (TBD from Task 3.1)
  - [ ] Verify: Acceptance criteria met
  - [ ] Test: No regressions
  - [ ] Document: Work completed

#### Task 3.3: Update V8 Status to 100%

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Actual**:
- **Dependencies**: Task 3.2
- **Description**: Mark V8 complete in documentation
- **Acceptance Criteria**:
  - [ ] Update: docs/fixes-needed-plan-tasks-v8/tasks.md (mark task complete)
  - [ ] Update: docs/fixes-needed-plan-tasks-v8/plan.md (59/59 tasks, 100%)
  - [ ] Commit: Changes with message
  - [ ] Document: V8 now 100% complete

### Phase 4: V9 Priority Tasks Completion

#### Task 4.1: Classify V9 Incomplete Tasks

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 0.6
- **Description**: Determine which V9 tasks are V10 immediate
- **Acceptance Criteria**:
  - [ ] Review: All 5 incomplete V9 tasks
  - [ ] Classify: V10 immediate (do now) vs deferred (future)
  - [ ] Prioritize: Order for immediate tasks
  - [ ] Document: Classification rationale

#### Task 4.2: Complete V9 Priority Task 1

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 4.1
- **Description**: Complete first V9 priority task (TBD from Task 4.1)
- **Acceptance Criteria**:
  - [ ] Implement: Required work
  - [ ] Verify: Acceptance criteria met
  - [ ] Test: No regressions
  - [ ] Document: Work completed

#### Task 4.3: Complete V9 Priority Task 2

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 4.2
- **Description**: Complete second V9 priority task (if classified immediate)
- **Acceptance Criteria**:
  - [ ] Implement: Required work
  - [ ] Verify: Acceptance criteria met
  - [ ] Test: No regressions
  - [ ] Document: Work completed

#### Task 4.4: Complete V9 Priority Task 3

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 4.3
- **Description**: Complete third V9 priority task (if classified immediate)
- **Acceptance Criteria**:
  - [ ] Implement: Required work
  - [ ] Verify: Acceptance criteria met
  - [ ] Test: No regressions
  - [ ] Document: Work completed

#### Task 4.5: Update V9 Status

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Tasks 4.2, 4.3, 4.4
- **Description**: Update V9 documentation with new completion status
- **Acceptance Criteria**:
  - [ ] Update: docs/fixes-needed-plan-tasks-v9/tasks.md (mark completed tasks)
  - [ ] Update: docs/fixes-needed-plan-tasks-v9/plan.md (new percentage)
  - [ ] Document: What remains deferred and why
  - [ ] Commit: Changes

### Phase 5: sm-kms cmd Structure Consistency

#### Task 5.1: cmd Structure Gap Analysis

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: None
- **Description**: Compare sm-kms cmd with cipher-im/jose-ja
- **Acceptance Criteria**:
  - [ ] List: Files in cmd/cipher-im/ (baseline)
  - [ ] List: Files in cmd/jose-ja/
  - [ ] List: Files in cmd/sm-kms/
  - [ ] Identify: Missing files in sm-kms
  - [ ] Document: Gap analysis

#### Task 5.2: Determine Necessary vs Optional Files

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 5.1
- **Description**: Classify which files are mandatory for sm-kms
- **Acceptance Criteria**:
  - [ ] Classify: main.go (mandatory)
  - [ ] Classify: Dockerfile, docker-compose.yml (necessary for E2E)
  - [ ] Classify: README.md, API.md, ENCRYPTION.md (documentation)
  - [ ] Classify: .dockerignore (build optimization)
  - [ ] Document: Rationale for each

#### Task 5.3: Add sm-kms Dockerfile

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 5.2
- **Description**: Create Dockerfile for sm-kms (if determined necessary)
- **Acceptance Criteria**:
  - [ ] Create: cmd/sm-kms/Dockerfile
  - [ ] Follow: cipher-im pattern
  - [ ] Test: `docker build -t sm-kms -f cmd/sm-kms/Dockerfile .`
  - [ ] Document: File created

#### Task 5.4: Add sm-kms docker-compose.yml

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 5.3
- **Description**: Create docker-compose.yml for sm-kms (if determined necessary)
- **Acceptance Criteria**:
  - [ ] Create: cmd/sm-kms/docker-compose.yml
  - [ ] Follow: cipher-im pattern (SQLite + PostgreSQL instances)
  - [ ] Include: Standard health checks
  - [ ] Document: File created

#### Task 5.5: Add sm-kms Documentation

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 5.4
- **Description**: Add README and other docs to cmd/sm-kms/ (if determined necessary)
- **Acceptance Criteria**:
  - [ ] Create: cmd/sm-kms/README.md
  - [ ] Consider: API.md, ENCRYPTION.md (if relevant)
  - [ ] Follow: cipher-im documentation pattern
  - [ ] Document: Files created

#### Task 5.6: Validation

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Tasks 5.3, 5.4, 5.5
- **Description**: Verify sm-kms cmd structure works
- **Acceptance Criteria**:
  - [ ] Build: `go build ./cmd/sm-kms/`
  - [ ] Docker: `docker build -f cmd/sm-kms/Dockerfile .`
  - [ ] Compose: `docker compose -f cmd/sm-kms/docker-compose.yml up`
  - [ ] Health: All containers healthy
  - [ ] Document: Validation results

### Phase 6: unsealkeysservice Code Audit

#### Task 6.1: Map unsealkeysservice Usage Across Services

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 0.9
- **Description**: Document how all services use unsealkeysservice
- **Acceptance Criteria**:
  - [ ] Check: Template usage pattern
  - [ ] Check: sm-kms usage pattern
  - [ ] Check: cipher-im usage pattern
  - [ ] Check: jose-ja usage pattern
  - [ ] Document: Import paths, usage patterns

#### Task 6.2: Template Barrier Code Analysis

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 6.1
- **Description**: Analyze template barrier for unseal logic duplication
- **Acceptance Criteria**:
  - [ ] Review: internal/apps/template/service/server/barrier/ code
  - [ ] Compare: Against internal/shared/barrier/unsealkeysservice/
  - [ ] Identify: Any duplicated unseal logic
  - [ ] Document: Findings

#### Task 6.3: Fix Duplications if Found

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 6.2
- **Description**: Refactor to eliminate duplicate code (if found)
- **Acceptance Criteria**:
  - [ ] Refactor: Remove duplications
  - [ ] Verify: Template imports shared unsealkeysservice
  - [ ] Test: All barrier tests pass
  - [ ] Document: Refactoring done (or N/A if no duplications)

### Phase 7: Quality Gates & Documentation

#### Task 7.1: Run Unit Tests

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: All previous tasks
- **Description**: Verify all unit tests pass
- **Acceptance Criteria**:
  - [ ] Run: `go test ./... -short`
  - [ ] Verify: Zero failures
  - [ ] Document: Test results

#### Task 7.2: Run Integration Tests

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 7.1
- **Description**: Verify all integration tests pass
- **Acceptance Criteria**:
  - [ ] Run: Integration tests for all services
  - [ ] Verify: Zero failures
  - [ ] Document: Test results

#### Task 7.3: Run E2E Tests

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 7.2
- **Description**: Verify all E2E tests pass with NO timeouts
- **Acceptance Criteria**:
  - [ ] Run: cipher-im E2E
  - [ ] Run: jose-ja E2E
  - [ ] Run: sm-kms E2E
  - [ ] Run: pki-ca E2E
  - [ ] Verify: All pass, zero timeouts
  - [ ] Document: Test results

#### Task 7.4: Run Linting

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: None (parallel with tests)
- **Description**: Verify all linters pass
- **Acceptance Criteria**:
  - [ ] Run: `golangci-lint run`
  - [ ] Run: `go run ./cmd/cicd lint-ports`
  - [ ] Run: `go run ./cmd/cicd lint-compose`
  - [ ] Run: `go run ./cmd/cicd lint-go`
  - [ ] Verify: All clean
  - [ ] Document: Results

#### Task 7.5: Verify Build Clean

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Actual**:
- **Dependencies**: None (parallel with tests)
- **Description**: Verify all packages build successfully
- **Acceptance Criteria**:
  - [ ] Run: `go build ./...`
  - [ ] Verify: Zero errors
  - [ ] Document: Build clean

#### Task 7.6: Update V8 Documentation

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Actual**:
- **Dependencies**: Task 3.3
- **Description**: Finalize V8 plan.md and tasks.md updates
- **Acceptance Criteria**:
  - [ ] Verify: docs/fixes-needed-plan-tasks-v8/tasks.md shows 59/59 (100%)
  - [ ] Verify: docs/fixes-needed-plan-tasks-v8/plan.md status Complete
  - [ ] Document: Final V8 status

#### Task 7.7: Update V9 Documentation

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Actual**:
- **Dependencies**: Task 4.5
- **Description**: Finalize V9 plan.md and tasks.md updates
- **Acceptance Criteria**:
  - [ ] Verify: Completed tasks marked
  - [ ] Document: Deferred tasks with rationale
  - [ ] Update: Completion percentage
  - [ ] Document: What remains for future work

#### Task 7.8: Add Health Timeout Lessons to Docs

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 1.8
- **Description**: Document E2E health timeout lessons learned
- **Acceptance Criteria**:
  - [ ] Add: Section to V10 plan.md
  - [ ] Document: Root cause, fix approach, best practices
  - [ ] Add: Recommendations for future E2E tests
  - [ ] Consider: Update architecture docs with health check patterns

#### Task 7.9: Update V10 Plan with Final Status

- **Status**:  Not Started
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Actual**:
- **Dependencies**: All tasks
- **Description**: Update V10 plan.md to status Complete
- **Acceptance Criteria**:
  - [ ] Update: plan.md status from Planning to Complete
  - [ ] Verify: All success criteria checked
  - [ ] Verify: All quality gates checked
  - [ ] Document: Final V10 completion

## Summary

**Total Tasks**: 47
**Completed**: 0
**In Progress**: 0
**Blocked**: 0
**Not Started**: 47

**Estimated Total LOE**: ~18.5h
**Actual Total LOE**: TBD
