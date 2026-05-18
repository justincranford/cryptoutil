# Tasks - Framework V23: Follow-Up Work from Framework-V22

**Status**: 0 of 18 tasks complete (0%)
**Created**: 2026-05-17
**Last Updated**: 2026-05-17

## Task Status Legend

| Symbol | Meaning |
|--------|---------|
| ❌ | Not started |
| 🔄 | In progress |
| ✅ | Complete |
| ⏳ | Blocked |

## Decision Constraints

1. All fixes must not introduce new test skips.
2. The LF fitness check must not produce false positives on non-secret text files.
3. E2E cert-dir cleanup must not regress Linux CI behavior.
4. POSTGRES_SECRETS_DIR validator must be additive to existing lint-deployments checks.
5. sm-im E2E SKIP reduction requires each skip to be either resolved or explicitly tracked.

## Phase 1: Fix Integration Test Failures

### Task 1.1: Diagnose sm-im/client TLS integration failure

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Root cause of TLS handshake failure identified and documented.
  - [ ] Evidence archived in `test-output/v23-phase1/sm-im-client-tls.txt`.

### Task 1.2: Fix sm-im/client TLS failure

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go test -tags integration ./internal/apps/sm-im/client/...` exits 0.
  - [ ] No new skips introduced.
  - [ ] `golangci-lint run ./internal/apps/sm-im/client/...` exits 0.

### Task 1.3: Diagnose sm-kms/client authorization failure

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Root cause of authorization test failure identified (scope validation logic).
  - [ ] Evidence archived in `test-output/v23-phase1/sm-kms-client-authz.txt`.

### Task 1.4: Fix sm-kms/client authorization failure

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go test -tags integration ./internal/apps/sm-kms/client/...` exits 0.
  - [ ] No new skips introduced.

### Task 1.5: Diagnose sm-kms/server/repository/orm barrier_content_keys failure

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Root cause identified: migration not applied before test execution, or missing table.
  - [ ] Evidence archived in `test-output/v23-phase1/sm-kms-barrier.txt`.

### Task 1.6: Fix sm-kms/server/repository/orm barrier_content_keys failure

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go test -tags integration ./internal/apps/sm-kms/server/repository/orm/...` exits 0.
  - [ ] No new skips introduced.

### Task 1.7: Phase 1 integration test verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go test -tags integration ./...` exits 0.
  - [ ] Evidence archived in `test-output/v23-phase1/final-integration.txt`.

## Phase 2: LF Line-Ending Fitness Check for Secrets Files

### Task 2.1: Implement LF fitness check in lint_deployments

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] New sub-validator in `lint_deployments` scans `deployments/*/secrets/*.secret`.
  - [ ] Returns an error with file path and line number for any CRLF content found.
  - [ ] `_test.go` covers: clean LF file passes, CRLF file fails, empty file passes.

### Task 2.2: Wire and verify LF fitness check

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Validator registered in `lint_deployments` main validator table.
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes on current repository.
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_deployments/...` passes.
  - [ ] Evidence archived in `test-output/v23-phase2/`.

## Phase 3: E2E Cert-Dir Writable Cleanup on Windows

### Task 3.1: Implement cert-dir writable check in test_orch_e2e

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `SetupE2ETestMain` (or equivalent) checks if cert-dir is writable before Compose starts.
  - [ ] If not writable, attempts `os.RemoveAll` + re-create with warning log.
  - [ ] `go test ./internal/apps-framework/service/test_orch_e2e/...` passes.

### Task 3.2: Verify E2E behavior on second local run

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Two sequential local E2E runs succeed without manual cert-dir cleanup.
  - [ ] Linux CI behavior unaffected (cert dirs already writable, no-op code path).
  - [ ] Evidence archived in `test-output/v23-phase3/`.

## Phase 4: POSTGRES_SECRETS_DIR Sync

### Task 4.1: Audit POSTGRES_SECRETS_DIR occurrences

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All occurrences of `POSTGRES_SECRETS_DIR` in `deployments/` catalogued.
  - [ ] Any current mismatches between `shared-postgres` template and per-PS-ID `.env.postgres` identified.
  - [ ] Evidence archived in `test-output/v23-phase4/audit.txt`.

### Task 4.2: Implement POSTGRES_SECRETS_DIR sync validator

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] New validator in `lint_deployments` detects value inconsistency.
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes on current (consistent) repo.
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_deployments/...` passes.

### Task 4.3: Fix any current mismatches

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Zero current violations after fixes.
  - [ ] `go run ./cmd/cicd-lint lint-deployments` exits 0.
  - [ ] Evidence archived in `test-output/v23-phase4/`.

## Phase 5: Reduce sm-im E2E SKIP Cases

### Task 5.1: Audit sm-im E2E t.Skip instances

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All `t.Skip` calls in `internal/apps/sm-im/e2e/` enumerated.
  - [ ] For each: root cause documented (blocker or deferred).
  - [ ] Evidence archived in `test-output/v23-phase5/audit.txt`.

### Task 5.2: Resolve or track all sm-im E2E skips

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All skips either resolved (test runs) or converted to tracked non-skip stubs.
  - [ ] `grep -n 't.Skip' internal/apps/sm-im/e2e/` returns 0 matches.
  - [ ] `go test -tags e2e ./internal/apps/sm-im/e2e/...` passes (with Docker infra running).
  - [ ] Evidence archived in `test-output/v23-phase5/`.

## Phase 6: Verification and Closure

### Task 6.1: Full quality suite

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0.
  - [ ] `go build -tags e2e,integration ./...` exits 0.
  - [ ] `go test ./...` exits 0.
  - [ ] `go test -tags integration ./...` exits 0.
  - [ ] `golangci-lint run ./...` exits 0.
  - [ ] `go run ./cmd/cicd-lint lint-go` exits 0.
  - [ ] `go run ./cmd/cicd-lint lint-deployments` exits 0.
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0.
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0.
  - [ ] Evidence archived in `test-output/v23-phase6/`.
