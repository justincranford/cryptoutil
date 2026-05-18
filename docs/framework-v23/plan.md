# Framework V23: Follow-Up Work from Framework-V22

**Status**: Not started
**Created**: 2026-05-17
**Last Updated**: 2026-05-17
**Purpose**: Close the five open items identified in `docs/framework-v22/POST-ANALYSIS.md`.
These are items that were incomplete or untraceable at the time of Framework-V22 declaration.

## Background

Framework-V22 executed successfully but left five open issues captured in the post-analysis:

- ISSUE-1: Integration test failures in 3 packages.
- ISSUE-3: Four lessons.md action items never promoted to tasks.

This plan promotes those items to tracked, executable tasks.

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

### Phase 2: LF Line-Ending Fitness Check for Secrets Files

Deployment secrets files (`deployments/*/secrets/*.secret` and related) can acquire CRLF
line endings on Windows, which breaks runtime secret parsing in Alpine containers.

A new fitness check in `lint_deployments` must:
- Scan all `deployments/*/secrets/*.secret` files.
- Return an error for any file that contains CRLF (`\r\n`) line endings.
- Be wired into `go run ./cmd/cicd-lint lint-deployments`.

**Acceptance**:
- `lint-deployments` fails with a clear message when a `.secret` file contains CRLF.
- `lint-deployments` passes on the current repository.
- `go test ./internal/apps-tools/cicd_lint/lint_deployments/...` passes.

### Phase 3: E2E Cert-Dir Writable Cleanup on Windows

E2E test orchestration creates cert directories before Compose startup. On Windows, these
directories can be locked by prior test runs, causing cert generation failures.

The E2E orchestration helper must:
- Check if the cert-dir is writable before Docker Compose starts.
- If the dir exists and is not writable, attempt `os.RemoveAll` and re-create.
- Log a warning if cleanup fails (do not panic — let Docker Compose surface the error).

**Acceptance**:
- E2E tests succeed on a second local run without manual `rm -rf` of cert directories.
- `go test ./internal/apps-framework/service/test_orch_e2e/...` passes.
- No regression on Linux CI (where cert dirs are always writable).

### Phase 4: POSTGRES_SECRETS_DIR Sync Across Templates

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

### Phase 5: Reduce sm-im E2E SKIP Cases

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

### Phase 6: Verification and Closure

- `go test -tags integration ./...` passes.
- `go run ./cmd/cicd-lint lint-deployments` passes (includes new validators).
- `go run ./cmd/cicd-lint lint-fitness` passes.
- `go run ./cmd/cicd-lint lint-docs` passes.
- `git status --porcelain` returns empty.

## Non-Goals

- PS-ID convergence or canonical directory migration (see Framework-V24).
- Race detector on Windows (requires CGO; runs in CI only via `ci-race.yml`).
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
- `test-output/v23-phase2/` — LF fitness check evidence.
- `test-output/v23-phase3/` — E2E cert-dir cleanup evidence.
- `test-output/v23-phase4/` — POSTGRES_SECRETS_DIR audit evidence.
- `test-output/v23-phase5/` — sm-im E2E skip reduction evidence.
- `test-output/v23-phase6/` — final verification evidence.
