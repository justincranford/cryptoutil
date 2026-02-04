# Tasks - Lint Enhancement & Technical Debt (V9)

**Status**: 0 of 17 tasks complete (0%)
**Last Updated**: 2026-02-04
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

## Phase 2: lint_go Technical Debt

### Task 2.1: Fix errcheck Issues
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**: 
- **Dependencies**: None
- **Description**: Fix file.Close() errcheck warnings
- **Acceptance Criteria**:
  - [ ] cryptopatterns.go line 112: Check file.Close() error
  - [ ] cryptopatterns.go line 290: Check file.Close() error
- **Files**:
  - `internal/cmd/cicd/lint_go/cryptopatterns.go`

### Task 2.2: Fix goconst Issue
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**: 
- **Dependencies**: None
- **Description**: Make repeated string `)` a constant
- **Acceptance Criteria**:
  - [ ] Define constant for `)` string
  - [ ] Use constant in lint_go.go
- **Files**:
  - `internal/cmd/cicd/lint_go/lint_go.go`

### Task 2.3: Fix gosec Issues
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**: 
- **Dependencies**: None
- **Description**: Fix WriteFile permissions warnings
- **Acceptance Criteria**:
  - [ ] cryptopatterns_test.go: Use 0600 instead of 0644 (4 occurrences)
- **Files**:
  - `internal/cmd/cicd/lint_go/cryptopatterns_test.go`

### Task 2.4: Verify lint_go Clean
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**: 
- **Dependencies**: Tasks 2.1-2.3
- **Description**: Verify no linting issues remain
- **Acceptance Criteria**:
  - [ ] `golangci-lint run ./internal/cmd/cicd/lint_go/...` shows 0 issues
- **Verification**:
  ```bash
  golangci-lint run --fix ./internal/cmd/cicd/lint_go/...
  ```

---

## Phase 3: Identity E2E Docker Investigation

### Task 3.1: Analyze Identity E2E Failures
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**: 
- **Dependencies**: None
- **Description**: Investigate identity-authz E2E test Docker failures
- **Acceptance Criteria**:
  - [ ] Root cause identified
  - [ ] Docker Compose configuration issues documented
- **Files**:
  - `internal/apps/identity/authz/e2e/`
  - `deployments/identity/compose.yml`

### Task 3.2: Fix Docker Compose Issues
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**: 
- **Dependencies**: Task 3.1
- **Description**: Fix identified Docker Compose issues
- **Acceptance Criteria**:
  - [ ] Docker Compose configuration corrected
  - [ ] Health checks working
  - [ ] Port mappings correct
- **Files**:
  - `deployments/identity/compose.yml`

### Task 3.3: Verify E2E Tests Pass
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**: 
- **Dependencies**: Task 3.2
- **Description**: Verify identity E2E tests pass
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/identity/.../e2e/... -count=1` passes
  - [ ] Docker containers start correctly
  - [ ] Health checks respond
- **Verification**:
  ```bash
  go test ./internal/apps/identity/.../e2e/... -count=1
  ```

---

## V9 Success Criteria

- [ ] lint-ports validates container ports, host ranges, health paths
- [ ] `golangci-lint run ./internal/cmd/cicd/lint_go/...` shows 0 issues
- [ ] identity E2E tests pass (if addressed)
- [ ] All existing tests continue to pass
- [ ] No regressions from V8 work