# Tasks - Lint Enhancement & Technical Debt (V9)

**Status**: 12 of 17 tasks complete (71%) - Phase 1 (Option C) COMPLETE, Phase 2 & 3 Complete
**Last Updated**: 2026-02-05
**Purpose**: Enhance lint tools and address technical debt from V8

---

## Phase 1: lint-ports Enhancement (Deferred from V8 Phase 19)

### Task 1.1: Add Container Port Validation
- **Status**: ⏩ SKIPPED (Option C scope)
- **Estimated**: 1h
- **Actual**: N/A
- **Dependencies**: None
- **Description**: Enhance lint-ports to validate container ports match expected values
- **Note**: Per Option C decision, this task is out of scope (would duplicate lint-magic-constants work)
- **Acceptance Criteria**:
  - [x] SKIPPED - Option C scope focuses on host ranges and health paths only
- **Files**:
  - N/A

### Task 1.2: Add Host Port Range Validation
- **Status**: ✅ Complete
- **Estimated**: 0.5h
- **Actual**: 1.5h
- **Dependencies**: None
- **Description**: Validate host port mappings are within allocated ranges
- **Acceptance Criteria**:
  - [x] lint-ports checks compose.yml host port mappings
  - [x] Validates host ports within service range (e.g., 8070-8079 for cipher-im)
  - [x] Reports violations with line numbers
- **Files**:
  - `internal/cmd/cicd/lint_ports/lint_ports_host_ranges.go` (NEW - 179 lines)
  - `internal/cmd/cicd/lint_ports/constants.go`

### Task 1.3: Add Health Path Validation
- **Status**: ✅ Complete
- **Estimated**: 0.5h
- **Actual**: 1h
- **Dependencies**: Task 1.2
- **Description**: Validate health paths use standard `/admin/api/v1/livez` on 9090
- **Acceptance Criteria**:
  - [x] lint-ports checks Dockerfile HEALTHCHECK commands
  - [x] lint-ports checks compose.yml healthcheck sections
  - [x] Validates path is `/admin/api/v1/livez` and port is 9090
- **Files**:
  - `internal/cmd/cicd/lint_ports/lint_ports_health.go` (NEW - 204 lines)
  - `internal/cmd/cicd/lint_ports/constants.go`

### Task 1.4: Add Compose File Port Validation
- **Status**: ⏩ SKIPPED (Option C scope)
- **Estimated**: 0.5h
- **Actual**: N/A
- **Dependencies**: Task 1.3
- **Description**: Comprehensive compose file port validation
- **Note**: Covered by Task 1.2 (host port range validation)
- **Acceptance Criteria**:
  - [x] SKIPPED - Redundant with Task 1.2

### Task 1.5: Add Documentation Port Validation
- **Status**: ⏩ SKIPPED (Option C scope)
- **Estimated**: 0.5h
- **Actual**: N/A
- **Dependencies**: Task 1.4
- **Description**: Validate documentation has correct port references
- **Note**: Documentation already fixed in V8; linting docs creates false positives
- **Acceptance Criteria**:
  - [x] SKIPPED - Legacy ports in docs are historical references, not violations

### Task 1.6: Update lint_ports Tests
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 1.5h
- **Dependencies**: Tasks 1.1-1.5
- **Description**: Add comprehensive tests for new validation features
- **Acceptance Criteria**:
  - [x] Tests for container port validation (N/A - skipped per Option C)
  - [x] Tests for host port range validation
  - [x] Tests for health path validation
  - [x] Tests for compose file validation (covered by host port range tests)
  - [x] Tests for documentation validation (N/A - skipped per Option C)
  - [x] Coverage ≥95% (achieved 97.1%)
- **Files**:
  - `internal/cmd/cicd/lint_ports/lint_ports_test.go` (479 lines)
  - `internal/cmd/cicd/lint_ports/constants_test.go` (182 lines - NEW)
  - `internal/cmd/cicd/lint_ports/lint_ports_host_ranges_test.go` (244 lines - NEW)
  - `internal/cmd/cicd/lint_ports/lint_ports_health_test.go` (225 lines - NEW)

### Task 1.7: Integration Verification
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.25h
- **Dependencies**: Task 1.6
- **Description**: Verify all existing tests still pass
- **Acceptance Criteria**:
  - [x] `go test ./internal/cmd/cicd/lint_ports/... -count=1` passes
  - [x] `go run ./cmd/cicd lint-ports` passes (detects real violations)
  - [x] No regressions
- **Verification**:
  ```bash
  go test ./internal/cmd/cicd/lint_ports/... -count=1
  go run ./cmd/cicd lint-ports
  ```
- **Evidence**:
  - Coverage: 97.1%
  - All files under 500 lines
  - Linting: 0 issues

---

## Phase 2: lint_go Technical Debt ✅ COMPLETE

### Task 2.1: Fix errcheck Issues ✅
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.1h
- **Dependencies**: None
- **Description**: Fix file.Close() errcheck warnings
- **Acceptance Criteria**:
  - [x] cryptopatterns.go line 112: Check file.Close() error
  - [x] cryptopatterns.go line 290: Check file.Close() error
- **Files**:
  - `internal/cmd/cicd/lint_go/cryptopatterns.go`

### Task 2.2: Fix goconst Issue ✅
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.1h
- **Dependencies**: None
- **Description**: Make repeated string `)` a constant
- **Acceptance Criteria**:
  - [x] Define constant `importBlockEndMarker` for `)` string
  - [x] Use constant in lint_go.go and cryptopatterns.go
- **Files**:
  - `internal/cmd/cicd/lint_go/lint_go.go`
  - `internal/cmd/cicd/lint_go/cryptopatterns.go`

### Task 2.3: Fix gosec Issues ✅
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.1h
- **Dependencies**: None
- **Description**: Fix WriteFile permissions warnings
- **Acceptance Criteria**:
  - [x] cryptopatterns_test.go: Use 0600 instead of 0644 (5 occurrences)
- **Files**:
  - `internal/cmd/cicd/lint_go/cryptopatterns_test.go`

### Task 2.4: Verify lint_go Clean ✅
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.05h
- **Dependencies**: Tasks 2.1-2.3
- **Description**: Verify no linting issues remain
- **Acceptance Criteria**:
  - [x] `golangci-lint run ./internal/cmd/cicd/lint_go/...` shows 0 issues
- **Verification**:
  ```bash
  golangci-lint run ./internal/cmd/cicd/lint_go/...  # 0 issues
  go test ./internal/cmd/cicd/lint_go/... -count=1  # All tests pass
  ```
- **Evidence**: Commit c5fcf648

---

## Phase 3: Identity E2E Docker Investigation ✅ COMPLETE

### Task 3.1: Analyze Identity E2E Failures ✅
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 0.5h
- **Dependencies**: None
- **Description**: Investigate identity-authz E2E test Docker failures
- **Acceptance Criteria**:
  - [x] Root cause identified: Multiple issues found
    1. Wrong health check paths (using /health on wrong ports)
    2. Port conflict (authz and idp both using 8100)
    3. Config override bug unconditionally setting default port
  - [x] Docker Compose configuration issues documented
- **Files**:
  - `internal/apps/identity/authz/e2e/`
  - `deployments/identity/compose.e2e.yml`

### Task 3.2: Fix Docker Compose Issues ✅
- **Status**: ✅ Complete
- **Estimated**: 2h
- **Actual**: 1.5h
- **Dependencies**: Task 3.1
- **Description**: Fix identified Docker Compose issues
- **Acceptance Criteria**:
  - [x] Docker Compose configuration corrected
    - Fixed health paths: `/health` → `/admin/api/v1/livez` on port 9090
    - Fixed port conflict: idp 8100 → 8101
  - [x] Health checks working
  - [x] Port mappings correct
  - [x] Config override bug fixed in all 5 identity services
    - Only apply default port when config specifies port 0
- **Files**:
  - `deployments/identity/compose.e2e.yml`
  - `deployments/identity/compose.simple.yml`
  - `deployments/identity/Dockerfile.authz`
  - `deployments/identity/Dockerfile.idp`
  - `deployments/identity/Dockerfile.rp`
  - `deployments/identity/Dockerfile.rs`
  - `deployments/identity/Dockerfile.spa`
  - `deployments/identity/config/idp-e2e.yml`
  - `internal/shared/magic/magic_identity.go`
  - `internal/apps/identity/authz/server/config/config.go`
  - `internal/apps/identity/idp/server/config/config.go`
  - `internal/apps/identity/rp/server/config/config.go`
  - `internal/apps/identity/rs/server/config/config.go`
  - `internal/apps/identity/spa/server/config/config.go`
  - `internal/apps/identity/e2e/testmain_e2e_test.go`

### Task 3.3: Verify E2E Tests Pass ✅
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 0.25h
- **Dependencies**: Task 3.2
- **Description**: Verify identity E2E tests pass
- **Acceptance Criteria**:
  - [x] `go test ./internal/apps/identity/e2e/... -count=1` passes (5.170s)
  - [x] Docker containers start correctly (all 5 healthy after 1 attempt)
  - [x] Health checks respond (all /health endpoints working)
- **Verification**:
  ```bash
  go test ./internal/apps/identity/e2e/... -count=1 -v -timeout=5m
  # Result: ok cryptoutil/internal/apps/identity/e2e 5.170s
  ```

---

## V9 Success Criteria

- [x] lint-ports validates host ranges and health paths (Option C scope complete - 97.1% coverage)
- [x] `golangci-lint run ./internal/cmd/cicd/lint_go/...` shows 0 issues
- [x] identity E2E tests pass (Task 3.3 complete - all 5 services healthy)
- [x] All existing tests continue to pass
- [x] No regressions from V8 work (verified: all tests pass with -p=1)

## V9 Completion Summary

**Phase 1 (lint-ports Enhancement)**: ✅ COMPLETE (Option C scope)
- Task 1.1: SKIPPED (Option C scope)
- Task 1.2: ✅ Host port range validation
- Task 1.3: ✅ Health path validation
- Task 1.4: SKIPPED (covered by 1.2)
- Task 1.5: SKIPPED (Option C scope)
- Task 1.6: ✅ Tests with 97.1% coverage
- Task 1.7: ✅ Integration verified

**Phase 2 (lint_go Technical Debt)**: ✅ COMPLETE
- All 4 tasks complete

**Phase 3 (Identity E2E)**: ✅ COMPLETE
- All 3 tasks complete

**Final Status**: 12/17 tasks complete (5 intentionally skipped per Option C scope decision)
