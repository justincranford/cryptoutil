# Tasks - Framework V23

**Status**: 12 of 13 tasks complete (92%)
**Created**: 2026-05-17
**Last Updated**: 2026-05-20

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

Before starting Phase 1, re-read `plan.md` Pre-Execution section. Per prior framework lessons:
- Update `tasks.md` after EVERY completed task (not in batch).
- Any phase touching Compose files MUST include Docker Compose verification within the same phase.

## Phase 1: Verify ISSUE-1 Resolved

### Task 1.1: Pre-flight integration test verification

- **Status**: ✅
- **Acceptance Criteria**:
  - [x] `go test -tags integration ./internal/apps/sm-im/client/...` exits 0.
  - [x] `go test -tags integration ./internal/apps/sm-kms/client/...` exits 0.
  - [x] `go test -tags integration ./internal/apps/sm-kms/server/repository/orm/...` exits 0.
  - [x] `go test -tags integration ./...` exits 0 (full suite, zero failures).
  - [x] Evidence archived in `test-output/v23-phase1/`.

## Phase 2: pki-init E2E Compose — Switch to Docker Named Volumes for Cert Storage

### Task 2.1: Audit all 11 compose files for bind mount cert usage

- **Status**: ✅
- **Acceptance Criteria**:
  - [x] All occurrences of `./certs:/certs` in `deployments/*/compose.yml` (10 PS-IDs) and
        `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml` enumerated.
  - [x] Current violation count confirmed (pki-init rw + 4 app services ro = 5 per file × 10 + 5 template = 55 bind mounts total; see pre-migration counts in `migration-report.txt`).
  - [x] Evidence archived in `test-output/v23-phase2/audit.txt`.

### Task 2.2: Update canonical template to use named volume

- **Status**: ✅
- **Acceptance Criteria**:
  - [x] `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml`:
        - pki-init: `./certs:/certs:rw` → `__PS_ID__-certs:/certs`
        - App services: `./certs:/certs:ro` → `__PS_ID__-certs:/certs:ro`
        - Top-level `volumes:` section declares `__PS_ID__-certs:`.
  - [x] Template still parses as valid YAML.

### Task 2.3: Update all 10 PS-ID compose files to named volumes

- **Status**: ✅
- **Acceptance Criteria**:
  - [x] For each PS-ID in {identity-authz, identity-idp, identity-rp, identity-rs, identity-spa,
        jose-ja, pki-ca, skeleton-template, sm-im, sm-kms}:
        - pki-init volume: `./certs:/certs:rw` → `{ps-id}-certs:/certs`
        - All app-service cert volumes: `./certs:/certs:ro` → `{ps-id}-certs:/certs:ro`
        - Top-level `volumes:` section added with `{ps-id}-certs:`.
  - [x] All 10 files parse as valid Docker Compose YAML (`docker compose config` passes).

### Task 2.4: Implement CO-21/CO-22 validators in lint_deployments

- **Status**: ✅
- **Acceptance Criteria**:
  - [x] New validator in `lint_deployments` detects `./certs:/certs` bind mounts in
        `deployments/*/compose.yml` (CO-21 violation).
  - [x] New validator detects missing top-level named volume declaration `{ps-id}-certs:` (CO-22).
  - [x] Pre-change failure mode covered by validator unit tests (forbidden bind-mount fixture and missing-volume fixture).
  - [x] `go test ./internal/apps-tools/cicd_lint/lint_deployments/...` passes.
  - [x] Evidence archived in `test-output/v23-phase2/validators.txt`.

### Task 2.5: Verify CO-21/CO-22 compliance via lint-deployments

- **Status**: ✅
- **Acceptance Criteria**:
  - [x] `go run ./cmd/cicd-lint lint-deployments` exits 0 (zero CO-21/CO-22 violations after migration).
  - [x] `docker compose -f deployments/sm-kms/compose.yml config` exits 0 (smoke-test one PS-ID).
  - [x] Docker Compose `up --wait` passes on at least one PS-ID scoped validation path (`opentelemetry-collector-contrib` on sm-kms) to confirm cert
        volume works end-to-end.
  - [x] Evidence archived in `test-output/v23-phase2/`.

## Phase 3: POSTGRES_SECRETS_DIR Sync

### Task 3.1: Audit POSTGRES_SECRETS_DIR occurrences

- **Status**: ✅
- **Acceptance Criteria**:
  - [x] All occurrences of `POSTGRES_SECRETS_DIR` in `deployments/` catalogued.
  - [x] Any current mismatches between `shared-postgres` template and per-PS-ID `.env.postgres` identified.
  - [x] Evidence archived in `test-output/v23-phase3/audit.txt`.

### Task 3.2: Implement POSTGRES_SECRETS_DIR sync validator

- **Status**: ✅
- **Acceptance Criteria**:
  - [x] New validator in `lint_deployments` detects value inconsistency.
  - [x] `go run ./cmd/cicd-lint lint-deployments` passes on current (consistent) repo.
  - [x] `go test ./internal/apps-tools/cicd_lint/lint_deployments/...` passes.

### Task 3.3: Fix any current mismatches

- **Status**: ✅
- **Acceptance Criteria**:
  - [x] Zero current violations after fixes.
  - [x] `go run ./cmd/cicd-lint lint-deployments` exits 0.
  - [x] Evidence archived in `test-output/v23-phase3/`.

## Phase 4: Reduce sm-im E2E SKIP Cases

### Task 4.1: Audit sm-im E2E t.Skip instances

- **Status**: ✅
- **Acceptance Criteria**:
  - [x] Both `t.Skip` calls in `internal/apps/sm-im/e2e/` enumerated and root causes documented:
        - `e2e_registration_test.go:93`: `join_tenant_id` not yet supported.
        - `e2e_test.go:67`: intentional OTEL port isolation.
  - [x] Evidence archived in `test-output/v23-phase4/audit.txt`.

### Task 4.2: Convert t.Skip calls to named constants

- **Status**: ⏳
- **Acceptance Criteria**:
  - [x] Skip 1 (`join_tenant_id`): either fix the framework registration handler, or replace
        inline `t.Skip` string with `skipReasonJoinTenantIDNotSupported` named constant.
  - [x] Skip 2 (OTEL port): replace inline `t.Skip` string with `skipReasonOtelPortNotExposed`
        named constant. This skip MUST remain — intentional design, not a defect.
  - [x] `golangci-lint run --build-tags e2e ./internal/apps/sm-im/e2e/...` exits 0.
  - [ ] `go test -tags e2e ./internal/apps/sm-im/e2e/...` passes (with Docker infra running).
  - [x] Evidence archived in `test-output/v23-phase4/`.

## Phase 5: Verification and Closure

### Task 5.1: Full quality suite

- **Status**: ✅
- **Acceptance Criteria**:
  - [x] `go build ./...` exits 0.
  - [x] `go build -tags e2e,integration ./...` exits 0.
  - [x] `go test ./...` exits 0.
  - [x] `go test -tags integration ./...` exits 0.
  - [x] `golangci-lint run ./...` exits 0.
  - [x] `golangci-lint run --build-tags e2e,integration ./...` exits 0.
  - [x] `go run ./cmd/cicd-lint lint-go` exits 0.
  - [x] `go run ./cmd/cicd-lint lint-deployments` exits 0.
  - [x] `go run ./cmd/cicd-lint lint-fitness` exits 0.
  - [x] `go run ./cmd/cicd-lint lint-docs` exits 0.
  - [x] Evidence archived in `test-output/v23-phase5/`.

## Phase 6: Knowledge Propagation

### Task 6.1: Review lessons and update permanent artifacts

- **Status**: ✅
- **Acceptance Criteria**:
  - [x] All phase post-mortems in `lessons.md` reviewed.
  - [x] `docs/ENG-HANDBOOK.md` and deployment guidance reviewed for needed updates.
  - [x] `.github/instructions/*.instructions.md` reviewed for needed updates.
  - [x] `.github/agents/*.agent.md` and `.claude/agents/*.md` updated with the new compose-runtime artifact cleanup rule.
  - [x] `go run ./cmd/cicd-lint lint-docs` exits 0.
  - [x] All artifact updates committed with separate semantic commits per artifact type.
  - [x] Evidence archived in `test-output/v23-phase6/`.
