# Tasks - Framework V23: Follow-Up Work from Framework-V22

**Status**: 0 of 16 tasks complete (0%)
**Created**: 2026-05-17
**Last Updated**: 2026-05-22

## Task Status Legend

| Symbol | Meaning |
|--------|---------|
| ❌ | Not started |
| 🔄 | In progress |
| ✅ | Complete |
| ⏳ | Blocked |

## Decision Constraints

1. All fixes must not introduce new test skips.
2. E2E compose named-volume change must not regress Linux CI behavior.
3. POSTGRES_SECRETS_DIR validator must be additive to existing lint-deployments checks.
4. sm-im E2E SKIP reduction requires each skip to be either resolved or explicitly tracked.

## Pre-Execution Enforcement

Before starting Phase 1, re-read `plan.md` Pre-Execution section. Per V22 lessons:
- Update `tasks.md` after EVERY completed task (not in batch).
- Any phase touching Compose files MUST include Docker Compose verification within the same phase.

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

## Phase 2: pki-init E2E Compose — Switch to Docker Named Volumes for Cert Storage

### Task 2.1: Audit all 11 compose files for bind mount cert usage

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All occurrences of `./certs:/certs` in `deployments/*/compose.yml` (10 PS-IDs) and
        `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml` enumerated.
  - [ ] Current violation count confirmed (expected: pki-init rw + 4 app services ro = 5 per file × 10 + 5 template = 55 bind mounts total).
  - [ ] Evidence archived in `test-output/v23-phase2/audit.txt`.

### Task 2.2: Update canonical template to use named volume

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml`:
        - pki-init: `./certs:/certs:rw` → `__PS_ID__-certs:/certs`
        - App services: `./certs:/certs:ro` → `__PS_ID__-certs:/certs:ro`
        - Top-level `volumes:` section declares `__PS_ID__-certs:`.
  - [ ] Template still parses as valid YAML.

### Task 2.3: Update all 10 PS-ID compose files to named volumes

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] For each PS-ID in {identity-authz, identity-idp, identity-rp, identity-rs, identity-spa,
        jose-ja, pki-ca, skeleton-template, sm-im, sm-kms}:
        - pki-init volume: `./certs:/certs:rw` → `{ps-id}-certs:/certs`
        - All app-service cert volumes: `./certs:/certs:ro` → `{ps-id}-certs:/certs:ro`
        - Top-level `volumes:` section added with `{ps-id}-certs:`.
  - [ ] All 10 files parse as valid Docker Compose YAML (`docker compose config` passes).

### Task 2.4: Verify CO-21/CO-22 compliance via lint-deployments

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-deployments` exits 0.
  - [ ] `docker compose -f deployments/sm-kms/compose.yml config` exits 0 (smoke-test one PS-ID).
  - [ ] Docker Compose `up --wait` passes on at least one PS-ID (e.g., sm-kms) to confirm cert
        volume works end-to-end.
  - [ ] Evidence archived in `test-output/v23-phase2/`.

## Phase 3: POSTGRES_SECRETS_DIR Sync

### Task 3.1: Audit POSTGRES_SECRETS_DIR occurrences

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All occurrences of `POSTGRES_SECRETS_DIR` in `deployments/` catalogued.
  - [ ] Any current mismatches between `shared-postgres` template and per-PS-ID `.env.postgres` identified.
  - [ ] Evidence archived in `test-output/v23-phase3/audit.txt`.

### Task 3.2: Implement POSTGRES_SECRETS_DIR sync validator

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] New validator in `lint_deployments` detects value inconsistency.
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes on current (consistent) repo.
  - [ ] `go test ./internal/apps-tools/cicd_lint/lint_deployments/...` passes.

### Task 3.3: Fix any current mismatches

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Zero current violations after fixes.
  - [ ] `go run ./cmd/cicd-lint lint-deployments` exits 0.
  - [ ] Evidence archived in `test-output/v23-phase3/`.

## Phase 4: Reduce sm-im E2E SKIP Cases

### Task 4.1: Audit sm-im E2E t.Skip instances

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All `t.Skip` calls in `internal/apps/sm-im/e2e/` enumerated.
  - [ ] For each: root cause documented (blocker or deferred).
  - [ ] Evidence archived in `test-output/v23-phase4/audit.txt`.

### Task 4.2: Resolve or track all sm-im E2E skips

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All skips either resolved (test runs) or converted to tracked non-skip stubs.
  - [ ] `grep -n 't.Skip' internal/apps/sm-im/e2e/` returns 0 matches.
  - [ ] `go test -tags e2e ./internal/apps/sm-im/e2e/...` passes (with Docker infra running).
  - [ ] Evidence archived in `test-output/v23-phase4/`.

## Phase 5: Verification and Closure

### Task 5.1: Full quality suite

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
  - [ ] Evidence archived in `test-output/v23-phase5/`.
