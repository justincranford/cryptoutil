# Framework V23: Follow-Up Work from Framework-V22

**Status**: Not started
**Created**: 2026-05-17
**Last Updated**: 2026-05-22
**Purpose**: Close the open items identified in `docs/framework-v22/POST-ANALYSIS.md`.
These are items that were incomplete or untracked at the time of Framework-V22 declaration.

## Background

Framework-V22 executed successfully but left open issues captured in the post-analysis:

- ISSUE-1: Integration test failures in 3 packages.
- ISSUE-3: Three lessons.md action items never promoted to tasks (item 1 — LF line endings —
  resolved separately via `.gitattributes * text=auto eol=lf`).

This plan promotes those items to tracked, executable tasks.

## Pre-Execution: V22 Lessons to Enforce

Before starting implementation, the executor MUST enforce these lessons from framework-V22:

1. **Per-task status updates are MANDATORY**: Update `tasks.md` immediately after each task
   completes. NEVER accumulate multiple task completions before updating documentation. A
   `tasks.md` that does not reflect actual state is a blocking artifact inconsistency.

2. **Docker Compose verification is MANDATORY within the same phase**: Any phase that modifies
   Compose files, configs consumed by containers, or cert paths MUST include a Docker Compose
   verification step (`docker compose up --wait` + health endpoint check) within the SAME phase.
   Phases declared complete without Compose verification are incomplete.

## Scope

### Phase 1: Fix Integration Test Failures

Three test packages are currently failing under `go test -tags integration ./...`:

1. `internal/apps/sm-im/client/` — TLS handshake failures (likely server cert trust issue).
2. `internal/apps/sm-kms/client/` — Authorization test failure (scope validation).
3. `internal/apps/sm-kms/server/repository/orm/` — Missing `barrier_content_keys` table
   (migration not applied before test execution).

**Acceptance**:
- `go test -tags integration ./internal/apps/sm-im/client/...` passes.
- `go test -tags integration ./internal/apps/sm-kms/client/...` passes.
- `go test -tags integration ./internal/apps/sm-kms/server/repository/orm/...` passes.
- Zero new test skips introduced.

### Phase 2: pki-init E2E Compose — Switch to Docker Named Volumes for Cert Storage

All 10 PS-ID `deployments/{ps-id}/compose.yml` files and the canonical template in
`api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml` currently use OS
filesystem bind mounts for cert delivery:

- `pki-init`: `./certs:/certs:rw` (writable bind mount)
- App services: `./certs:/certs:ro` (read-only bind mount)

This violates deployment template rules CO-21 and CO-22 (see `docs/deployment-templates.md`
and ENG-HANDBOOK §6.11.3), which mandate named Docker volumes:

- CO-21: All services MUST use named volume `{ps-id}-certs:/certs` (NEVER bind mount
  `./certs/:/certs`).
- CO-22: Top-level `volumes:` section MUST declare `{ps-id}-certs:` (no `driver:` override).

On Windows, bind-mounted host directories can be locked from a prior run, causing `pki-init`
cert generation to fail. Named Docker volumes eliminate this host-filesystem dependency.

This phase changes all 10 compose files and the canonical template:
- pki-init: `./certs:/certs:rw` → `{ps-id}-certs:/certs` (Docker named volume, writable)
- App services: `./certs:/certs:ro` → `{ps-id}-certs:/certs:ro` (Docker named volume, read-only)
- Add top-level `volumes:` declaration: `{ps-id}-certs:` to each compose file

Affected files (11 total):
- `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml` (template: `__PS_ID__-certs`)
- `deployments/identity-authz/compose.yml`
- `deployments/identity-idp/compose.yml`
- `deployments/identity-rp/compose.yml`
- `deployments/identity-rs/compose.yml`
- `deployments/identity-spa/compose.yml`
- `deployments/jose-ja/compose.yml`
- `deployments/pki-ca/compose.yml`
- `deployments/skeleton-template/compose.yml`
- `deployments/sm-im/compose.yml`
- `deployments/sm-kms/compose.yml`

**Acceptance**:
- All 10 PS-ID compose files use `{ps-id}-certs:/certs` named volumes (zero bind mounts for certs).
- Canonical template uses `__PS_ID__-certs:/certs` named volume.
- `go run ./cmd/cicd-lint lint-deployments` passes (CO-21/CO-22 compliance).
- E2E tests succeed on sequential local runs without manual cert directory cleanup.
- No regression on Linux CI.
- Docker Compose verification run within this phase (`docker compose up --wait` + health check
  on at least one PS-ID).

### Phase 3: POSTGRES_SECRETS_DIR Sync Across Templates

`POSTGRES_SECRETS_DIR` is referenced in `shared-postgres` Docker Compose templates and in
per-PS-ID `.env.postgres` files. If the two are out of sync, secrets resolution fails.

This phase:
- Audits all occurrences of `POSTGRES_SECRETS_DIR` in `deployments/`.
- Adds a `lint-deployments` validator to detect when the PS-ID `.env.postgres` value
  differs from the `shared-postgres` template expectation.
- Fixes any current mismatches found.

**Acceptance**:
- `lint-deployments` fails when `POSTGRES_SECRETS_DIR` values are inconsistent.
- All current values are consistent (zero current violations).
- `go test ./internal/apps-tools/cicd_lint/lint_deployments/...` passes.

### Phase 4: Reduce sm-im E2E SKIP Cases

At least 4 E2E test cases in `internal/apps/sm-im/e2e/` are currently skipped with `t.Skip`.
Each skip should be resolved to a real test or a tracked issue with a rationale.

This phase:
- Audits all `t.Skip` in `internal/apps/sm-im/e2e/`.
- For each skip: either fix the underlying blocker and re-enable the test, or file a
  tracking issue and convert the `t.Skip` to a clear `t.Log` + explicit skipped-reason
  constant.
- Reduces skip count from ≥4 to 0 active skips.

**Acceptance**:
- `grep -n 't.Skip' internal/apps/sm-im/e2e/` returns 0 matches.
- Any former-skip scenarios now run (even as empty stubs asserting "not yet implemented")
  with explicit tracking.

### Phase 5: Verification and Closure

- `go test -tags integration ./...` passes.
- `go run ./cmd/cicd-lint lint-deployments` passes (includes new validators).
- `go run ./cmd/cicd-lint lint-fitness` passes.
- `go run ./cmd/cicd-lint lint-docs` passes.
- `git status --porcelain` returns empty.

## Non-Goals

- PS-ID convergence or canonical directory migration (see Framework-V24).
- Race detector on Windows (requires CGO; runs in CI only via `ci-race.yml` — see ENG-HANDBOOK §10.9).
- LF line endings: already enforced via `.gitattributes * text=auto eol=lf`.
- Changes to the V22 EXEC-SUMMARY.md or DETAIL-SUMMARY.md.

## Quality Gates

- `go build ./...`
- `go build -tags e2e,integration ./...`
- `go test ./...`
- `go test -tags integration ./...`
- `golangci-lint run ./...`
- `go run ./cmd/cicd-lint lint-go`
- `go run ./cmd/cicd-lint lint-deployments`
- `go run ./cmd/cicd-lint lint-fitness`
- `go run ./cmd/cicd-lint lint-docs`

## Evidence Strategy

- `test-output/v23-phase1/` — integration test failure evidence and fix verification.
- `test-output/v23-phase2/` — Docker named volumes for cert storage: compose diff and lint evidence.
- `test-output/v23-phase3/` — POSTGRES_SECRETS_DIR audit evidence.
- `test-output/v23-phase4/` — sm-im E2E skip reduction evidence.
- `test-output/v23-phase5/` — final verification evidence.
