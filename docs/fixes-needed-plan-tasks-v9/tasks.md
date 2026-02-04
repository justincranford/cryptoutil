# Tasks - Lint Enhancement & Technical Debt (V9)

**Status**: 7 of 17 tasks complete (41%)
**Last Updated**: 2026-02-05
**Purpose**: Enhance lint tools and address technical debt from V8

---

## Phase 1: lint-ports Enhancement (Deferred from V8 Phase 19)

### Task 1.1: Add Container Port Validation
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: None
- **Description**: Enhance lint-ports to validate container ports match expected values
- **Acceptance Criteria**:
  - [ ] lint-ports checks Go config files for correct container ports
  - [ ] sm-kms uses 8080, cipher-im uses 8070, jose-ja uses 8060, pki-ca uses 8050
  - [ ] identity services use 8100/8110/8120/8130 appropriately
  - [ ] Test coverage for new validation
- **Files**:
  - `internal/cmd/cicd/lint_ports/lint_ports.go`
  - `internal/cmd/cicd/lint_ports/constants.go`

### Task 1.2: Add Host Port Range Validation
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 1.1
- **Description**: Validate host port mappings are within allocated ranges
- **Acceptance Criteria**:
  - [ ] lint-ports checks compose.yml host port mappings
  - [ ] Validates host ports within service range (e.g., 8070-8079 for cipher-im)
  - [ ] Reports violations with line numbers
- **Files**:
  - `internal/cmd/cicd/lint_ports/lint_ports.go`

### Task 1.3: Add Health Path Validation
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 1.2
- **Description**: Validate health paths use standard `/admin/api/v1/livez` on 9090
- **Acceptance Criteria**:
  - [ ] lint-ports checks Dockerfile HEALTHCHECK commands
  - [ ] lint-ports checks compose.yml healthcheck sections
  - [ ] Validates path is `/admin/api/v1/livez` and port is 9090
- **Files**:
  - `internal/cmd/cicd/lint_ports/lint_ports.go`

### Task 1.4: Add Compose File Port Validation
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 1.3
- **Description**: Comprehensive compose file port validation
- **Acceptance Criteria**:
  - [ ] lint-ports validates all port mappings in compose files
  - [ ] Checks public ports match service standards
  - [ ] Checks admin port NOT exposed (defers to lint-compose)
- **Files**:
  - `internal/cmd/cicd/lint_ports/lint_ports.go`

### Task 1.5: Add Documentation Port Validation
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 1.4
- **Description**: Validate documentation has correct port references
- **Acceptance Criteria**:
  - [ ] lint-ports checks architecture.instructions.md
  - [ ] Validates port references match standards
  - [ ] Reports documentation drift
- **Files**:
  - `internal/cmd/cicd/lint_ports/lint_ports.go`

### Task 1.6: Update lint_ports Tests
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Tasks 1.1-1.5
- **Description**: Add comprehensive tests for new validation features
- **Acceptance Criteria**:
  - [ ] Tests for container port validation
  - [ ] Tests for host port range validation
  - [ ] Tests for health path validation
  - [ ] Tests for compose file validation
  - [ ] Tests for documentation validation
  - [ ] Coverage ≥95%
- **Files**:
  - `internal/cmd/cicd/lint_ports/lint_ports_test.go`

### Task 1.7: Integration Verification
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Task 1.6
- **Description**: Verify all existing tests still pass
- **Acceptance Criteria**:
  - [ ] `go test ./internal/cmd/cicd/lint_ports/... -count=1` passes
  - [ ] `go run ./cmd/cicd lint-ports` passes
  - [ ] No regressions
- **Verification**:
  ```bash
  go test ./internal/cmd/cicd/lint_ports/... -count=1
  go run ./cmd/cicd lint-ports
  ```

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

- [ ] lint-ports validates container ports, host ranges, health paths
- [x] `golangci-lint run ./internal/cmd/cicd/lint_go/...` shows 0 issues
- [x] identity E2E tests pass (Task 3.3 complete - all 5 services healthy)
- [x] All existing tests continue to pass
- [ ] No regressions from V8 work
