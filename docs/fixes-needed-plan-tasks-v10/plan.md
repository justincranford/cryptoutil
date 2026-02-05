# Implementation Plan V10 - Critical Regressions and Completion Fixes

**Status**: Planning
**Created**: 2026-02-05
**Last Updated**: 2026-02-05

## Overview

V10 addresses critical gaps discovered in V8/V9 completion claims and recurring E2E health timeout regressions. This plan focuses on evidence-based verification, root cause analysis, and systematic resolution of incomplete work.

## Executive Summary

**Critical Findings**:
1. **E2E Health Timeouts** - Recurring regression across multiple services (cipher-im confirmed, others suspected)
2. **V8 Incomplete Work** - Claims 98% (58/59) but 1 task remains unfinished
3. **V9 Incomplete Work** - Claims 71% (12/17) but 5 tasks incomplete (Options B/C expansion deferred)
4. **Import Path Breakage** - `cryptoutilClient "cryptoutil/internal/kms/client"` broken after KMS migration
5. **unsealkeysservice Location** - Kept in `internal/shared/barrier/unsealkeysservice/` (shared by template + all services)
6. **sm-kms cmd Inconsistency** - Uses `cmd/sm-kms/main.go` while cipher-im/jose-ja have richer structures

## Technical Context

- **Language**: Go 1.25.5
- **Framework**: Fiber (HTTP), GORM (database), Docker Compose (orchestration)
- **Database**: PostgreSQL 18 + SQLite (dual support)
- **Services**: sm-kms, cipher-im, jose-ja, pki-ca, identity-*
- **Testing**: Unit + Integration + E2E (Docker Compose)
- **Critical Files**:
  - `internal/test/e2e/assertions.go` - Broken import after KMS migration
  - `cmd/cipher-im/docker-compose.yml` - E2E health check configuration
  - `internal/apps/cipher/im/e2e/e2e_test.go` - E2E health timeout pattern
  - `cmd/sm-kms/main.go` - Minimal structure vs cipher-im/jose-ja

## Phases

### Phase 0: Evidence Collection & Root Cause Analysis (3h)

**Objective**: Gather evidence for all claimed issues, verify actual vs claimed completion status

#### 0.1 E2E Health Timeout Analysis (1h)

- Reproduce cipher-im E2E timeout (90s failure)
- Check jose-ja, sm-kms, pki-ca, identity-* E2E tests for similar patterns
- Analyze health check configurations across all docker-compose.yml files
- Compare health endpoint paths: `/admin/api/v1/livez` vs `/service/api/v1/health`
- Document: Which services have timeouts, which pass, configuration differences

#### 0.2 V8 Completion Verification (0.5h)

- Review docs/fixes-needed-plan-tasks-v8/tasks.md
- Identify the 1 incomplete task (58/59 = 98%)
- Determine if it's truly incomplete or mislabeled
- Document: Actual status, blocker if any, estimated LOE

#### 0.3 V9 Completion Verification (0.5h)

- Review docs/fixes-needed-plan-tasks-v9/tasks.md
- List all 5 incomplete tasks (12/17 = 71%)
- Classify: Deferred to V10, blocked, or incorrectly marked incomplete
- Document: Which need immediate attention vs future work

#### 0.4 Import Path Breakage Analysis (0.5h)

- Verify `internal/test/e2e/assertions.go` compile error
- Check all references to `cryptoutil/internal/kms/client`
- Identify new location: `cryptoutil/internal/apps/sm/kms/client`
- List all broken files requiring refactoring
- Document: Scope of refactoring, estimated LOE

#### 0.5 unsealkeysservice Location Audit (0.5h)

- Verify `internal/shared/barrier/unsealkeysservice/` exists
- Check template usage: grep -r "unsealkeysservice" internal/apps/template
- Check service usage: grep -r "unsealkeysservice" internal/apps/sm/kms
- Identify duplicate code if any
- Document: Why kept, shared vs duplicated code

### Phase 1: E2E Health Timeout Root Cause & Fix (4h)

**Objective**: Eliminate recurring E2E health timeout regression

#### 1.1 Health Endpoint Standardization Analysis (1h)

- Audit ALL services for health endpoint registration
- Check: `/admin/api/v1/livez` vs `/service/api/v1/health` vs both
- Verify: Template base registers both paths correctly
- Identify: Services with non-standard patterns
- Document: Standard pattern, violations, fix plan

#### 1.2 Docker Compose Health Check Audit (1h)

- Review ALL docker-compose.yml files for health check configuration
- Check: Port (9090 vs 8xxx), path, interval, timeout, retries, start_period
- Compare: cipher-im (fails) vs jose-ja (passes) vs sm-kms (unknown)
- Identify: Configuration patterns causing timeouts
- Document: Standard configuration, violations, recommended values

#### 1.3 E2E Test Health Check Pattern Analysis (1h)

- Review cipher-im E2E test: WaitForHealth implementation
- Check: Timeout values, retry logic, endpoint paths used
- Compare with template E2E helpers if they exist
- Identify: Test-side vs service-side issues
- Document: Recommended E2E health check pattern

#### 1.4 Fix and Validate All Services (1h)

- Apply fixes to all identified violations
- Standardize health checks across all docker-compose.yml
- Run E2E tests for ALL services (not just cipher-im)
- Verify: All pass within reasonable timeouts (30-60s max)
- Document: Changes made, test results

### Phase 2: Import Path Breakage Fix (2h)

**Objective**: Fix all broken imports after KMS migration

#### 2.1 Refactor internal/test/e2e/assertions.go (0.5h)

- Change: `cryptoutil/internal/kms/client`  `cryptoutil/internal/apps/sm/kms/client`
- Verify: Compile succeeds
- Run: E2E tests to ensure no runtime issues
- Document: Files changed

#### 2.2 Audit and Fix All KMS Client Imports (1h)

- grep -r "internal/kms/client" across entire codebase
- Refactor: All occurrences to new path
- Check: Test files, main files, other services
- Verify: `go build ./...` succeeds
- Document: Complete list of files changed

#### 2.3 Verify No Legacy KMS Paths Remain (0.5h)

- grep -r "internal/kms" across codebase (should only be migration docs)
- Check: No compile errors, no test failures
- Run: Full test suite to verify
- Document: Verification results

### Phase 3: V8 Completion (1h)

**Objective**: Complete the 1 remaining V8 task

#### 3.1 Complete Task 58/59 (0.75h)

- Implement missing task (TBD based on Phase 0.2 findings)
- Verify: Acceptance criteria met
- Test: No regressions
- Document: Work completed

#### 3.2 Update V8 Status to 100% (0.25h)

- Mark task complete in docs/fixes-needed-plan-tasks-v8/tasks.md
- Update plan.md completion percentage
- Document: Final V8 status

### Phase 4: V9 Priority Tasks Completion (3h)

**Objective**: Complete V9 tasks that should be in V10 scope

#### 4.1 Review V9 Incomplete Tasks (0.5h)

- Based on Phase 0.3 findings
- Classify: V10 immediate vs future work
- Prioritize: Which must be done now
- Document: Task prioritization decision

#### 4.2 Complete Priority V9 Tasks (2h)

- Work on tasks classified as V10 immediate
- Verify: Each task's acceptance criteria
- Test: No regressions
- Document: Work completed per task

#### 4.3 Update V9 Status (0.5h)

- Mark completed tasks in docs/fixes-needed-plan-tasks-v9/tasks.md
- Update plan.md with new completion percentage
- Document: What remains deferred and why

### Phase 5: sm-kms cmd Structure Consistency (2h)

**Objective**: Align sm-kms cmd with cipher-im/jose-ja patterns

#### 5.1 Analyze cmd Structure Differences (0.5h)

- Compare: cmd/sm-kms/ vs cmd/cipher-im/ vs cmd/jose-ja/
- Identify: Missing files (README.md, docker-compose.yml, Dockerfile, etc.)
- Determine: Which files are necessary vs optional
- Document: Gap analysis, rationale for each file

#### 5.2 Implement cmd Structure Alignment (1h)

- Add missing files to cmd/sm-kms/
- Standardize structure across all services
- Verify: Docker build, compose up still works
- Document: Files added, changes made

#### 5.3 Validation (0.5h)

- Build: `go build ./cmd/sm-kms/`
- Test: Docker build, compose up, health checks
- Compare: Final structure matches pattern
- Document: Verification results

### Phase 6: unsealkeysservice Code Audit (1.5h)

**Objective**: Verify no duplicate code between template and shared

#### 6.1 Map unsealkeysservice Usage (0.5h)

- grep -r "unsealkeysservice" internal/apps/template
- grep -r "unsealkeysservice" internal/apps/sm/kms
- grep -r "unsealkeysservice" internal/apps/cipher/im
- grep -r "unsealkeysservice" internal/apps/jose/ja
- Document: Which services use it, how it's imported

#### 6.2 Analyze Template Barrier vs Shared unsealkeysservice (0.5h)

- Compare: internal/apps/template/service/server/barrier/ code
- Check: Does template duplicate unsealkeysservice logic?
- Verify: Template uses unsealkeysservice correctly (imports, no duplication)
- Document: Code reuse pattern, any duplications found

#### 6.3 Fix Duplications if Found (0.5h)

- Refactor: Remove any duplicate code
- Standardize: All services use shared unsealkeysservice
- Verify: Tests pass, no compile errors
- Document: Refactoring done

### Phase 7: Quality Gates & Documentation (2h)

**Objective**: Ensure all fixes meet quality standards

#### 7.1 Comprehensive Testing (1h)

- Unit: `go test ./...`
- Integration: All service integration tests
- E2E: All docker-compose E2E tests (cipher-im, jose-ja, sm-kms, etc.)
- Verify: Zero failures, zero timeouts
- Document: Test results

#### 7.2 Linting and Build Verification (0.5h)

- `golangci-lint run`
- `go run ./cmd/cicd lint-ports`
- `go run ./cmd/cicd lint-compose`
- `go run ./cmd/cicd lint-go`
- `go build ./...`
- Document: All clean

#### 7.3 Documentation Updates (0.5h)

- Update: V8, V9, V10 plan.md and tasks.md with final status
- Document: Lessons learned about health timeouts
- Add: Best practices for E2E health checks
- Update: Architecture docs if cmd structure changed
- Document: All updates made

## Technical Decisions

### Decision 1: E2E Health Timeout Root Cause

- **Chosen**: TBD (based on Phase 0.1 + 1.1-1.3 analysis)
- **Rationale**: Need evidence before deciding (config issue vs code issue vs test issue)
- **Alternatives**: (Will document after analysis)
- **Impact**: Determines fix approach

### Decision 2: V9 Task Prioritization

- **Chosen**: TBD (based on Phase 0.3 findings)
- **Rationale**: Some V9 tasks may be Option B/C scope (deferred), others may be V10 critical
- **Alternatives**: Complete all V9 tasks now vs defer some to future
- **Impact**: V10 scope and timeline

### Decision 3: sm-kms cmd Structure

- **Chosen**: TBD (based on Phase 5.1 analysis)
- **Rationale**: Need to determine which files are mandatory vs nice-to-have
- **Alternatives**: Full alignment vs minimal alignment
- **Impact**: Consistency across services

### Decision 4: unsealkeysservice Shared Location

- **Chosen**: Keep in internal/shared/barrier/unsealkeysservice/ (already decided in V8)
- **Rationale**: Used by template + all services, truly shared code
- **Alternatives**: Move to template (but then services can't import template)
- **Impact**: Confirms current location is correct

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| E2E health timeouts persist after fixes | Medium | High | Deep root cause analysis before applying fixes |
| V8/V9 incomplete tasks have hidden blockers | Low | Medium | Thorough evidence collection in Phase 0 |
| Import path refactoring breaks tests | Low | High | Incremental refactoring with testing after each change |
| sm-kms cmd changes break Docker builds | Low | Medium | Test Docker build after each file addition |
| unsealkeysservice has undiscovered duplications | Low | Low | Comprehensive code audit in Phase 6 |

## Quality Gates

- All tests pass (`go test ./...` with NO timeouts)
- E2E tests pass for ALL services (cipher-im, jose-ja, sm-kms, pki-ca)
- Coverage 95% production, 98% infrastructure (maintained)
- Mutation testing 95% minimum, 98% infrastructure (maintained)
- Linting clean (`golangci-lint run`)
- CICD linters pass (lint-ports, lint-compose, lint-go)
- No compile errors (`go build ./...`)
- No broken imports anywhere
- V8 100% complete (59/59 tasks)
- V9 status accurate (deferred tasks documented)
- Docker Compose works for all services

## Success Criteria

- [ ] E2E health timeouts eliminated (ALL services pass)
- [ ] V8 completion verified as 100%
- [ ] V9 priority tasks complete
- [ ] All import paths working
- [ ] sm-kms cmd structure consistent with other services
- [ ] unsealkeysservice code audit complete (no duplications)
- [ ] All quality gates pass
- [ ] Documentation updated with lessons learned
- [ ] CI/CD green across all workflows

## Evidence Location

All verification artifacts stored in:
- `test-output/v10-evidence/` - Root cause analysis, test results
- `test-output/v10-e2e-health/` - E2E health timeout investigation
- `test-output/v10-import-fix/` - Import path refactoring verification
- `test-output/v10-completion/` - V8/V9 completion verification
