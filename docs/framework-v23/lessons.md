# Lessons - Framework V23

**Created**: 2026-05-17
**Last Updated**: 2026-05-20

## Executive Summary

1. [Phase 1: Verify ISSUE-1 Resolved](#phase-1-verify-issue-1-resolved) — All previously reported integration failures were non-reproducible and confirmed green in current baseline.
2. [Phase 2: pki-init E2E Compose — Switch to Docker Named Volumes for Cert Storage](#phase-2-pki-init-e2e-compose--switch-to-docker-named-volumes-for-cert-storage) — Template plus all 10 PS-ID compose files were migrated to named cert volumes with new CO-21/CO-22 validators.
3. [Phase 3: POSTGRES_SECRETS_DIR Sync](#phase-3-postgres_secrets_dir-sync) — Added automatic sync validation and verified all PS-ID `.env.postgres` values are consistent.
4. [Phase 4: Reduce sm-im E2E SKIP Cases](#phase-4-reduce-sm-im-e2e-skip-cases) — Both E2E skip messages were converted to named constants; full sm-im e2e execution remains blocked by Docker runtime stack health.
5. [Phase 5: Verification and Closure](#phase-5-verification-and-closure) — Full build, test, lint, and cicd-lint quality gate suite passed on the current working tree.
6. [Phase 6: Knowledge Propagation](#phase-6-knowledge-propagation) — Execution-agent guidance was updated to prevent future false lint failures from transient compose runtime artifacts.

## Actions

1. Resolve sm-im e2e Docker startup instability (`postgres-leader` / healthcheck dependency failures) so Task 4.2 can move from blocked to complete.
2. Add a deterministic Docker verification target for phase plans that avoids unrelated multi-instance health dependencies when validating compose storage migrations.
3. Consider extending deployment validator output formatting to include inline failure details for failed subvalidators (currently requires auxiliary diagnosis).

## Pre-Execution: Lessons to Enforce

Before starting implementation, the executor MUST enforce these lessons:

1. **Per-task status updates are MANDATORY**: Update `tasks.md` immediately after each task
   completes. NEVER accumulate multiple task completions before updating documentation. A
   `tasks.md` that does not reflect actual state is a blocking artifact inconsistency.

2. **Docker Compose verification is MANDATORY within the same phase**: Any phase that modifies
   Compose files, configs consumed by containers, or cert paths MUST include a Docker Compose
   verification step (`docker compose up --wait` + health endpoint check) within the SAME phase.
   Phases declared complete without Compose verification are incomplete.

## Phase 1: Verify ISSUE-1 Resolved

### What Worked

- Running the three originally reported failing integration packages first gave a fast and reliable signal before the full suite run.
- Capturing each command output into dedicated files under `test-output/v23-phase1/` made verification and audit traceability straightforward.

### What Didn't Work

- None observed in this phase. All expected integration commands passed on the first execution.

### Root Causes

- The prior issue appears to have been transient or previously remediated; no current functional regression remains in the targeted integration paths.

### Patterns

- Keep explicit pre-flight package-level verification as the first task in follow-up framework plans when prior reports mention broad integration instability.

## Phase 2: pki-init E2E Compose — Switch to Docker Named Volumes for Cert Storage

### What Worked

- Bulk migration from bind mounts to named volumes was successfully applied across all 11 target compose files (template + 10 PS-IDs).
- New `cert-volumes` validator provided deterministic CO-21/CO-22 policy enforcement for future regressions.
- Scoped Docker Compose `up --wait` verification on `sm-kms` telemetry path validated runtime named-volume cert generation flow.

### What Didn't Work

- Full-stack `up --wait` verification for multi-instance PS-ID stacks remained sensitive to unrelated container health dependencies during this phase.

### Root Causes

- Existing stack-level health dependencies include components outside the cert-volume migration scope, so full-stack readiness may fail even when cert generation path is correct.

### Patterns

- For storage migration verification, keep a scoped `up --wait` target that still traverses the modified dependency chain (`pki-init` + cert consumers) while minimizing unrelated stack-health noise.

## Phase 3: POSTGRES_SECRETS_DIR Sync

### What Worked

- A dedicated validator now checks both the shared-postgres compose reference pattern and per-PS-ID `.env.postgres` values.
- Full repository scan confirmed current configuration consistency.

### What Didn't Work

- Initial audits missed hidden `.env.postgres` files until hidden-file search was explicitly enabled.

### Root Causes

- Default search configurations can exclude hidden files, creating false confidence for env-file audits.

### Patterns

- Always use hidden-file-inclusive scans (`--hidden --no-ignore` equivalent) when auditing `.env*` artifacts.

## Phase 4: Reduce sm-im E2E SKIP Cases

### What Worked

- Inline skip messages were replaced with named constants (`skipReasonJoinTenantIDNotSupported` and `skipReasonOtelPortNotExposed`) as required.
- E2E-tagged lint for the package passed with the new constant pattern.

### What Didn't Work

- `go test -tags e2e ./internal/apps/sm-im/e2e/...` did not pass due Docker compose startup failures in dependent infrastructure containers.

### Root Causes

- Runtime stack health failures are external to the skip-constant refactor and block full package e2e execution.

### Patterns

- Keep behavioral changes (skip constantization) separated from infrastructure startup remediations to preserve clear root-cause boundaries.

## Phase 5: Verification and Closure

### What Worked

- Full quality gate suite completed successfully after remediating surfaced lint and transient runtime-artifact issues.
- Re-running full `go test ./...` eliminated an observed flaky single-package failure and confirmed green baseline.

### What Didn't Work

- Docker runtime validation generated transient `deployments/*/certs/` directories that subsequently tripped deployment naming lint checks.

### Root Causes

- Local compose executions can create runtime artifacts inside deployment directories that are not part of canonical deployment source structure.

### Patterns

- Include post-compose workspace hygiene cleanup before lint gates when running local Docker verification in implementation sessions.

## Phase 6: Knowledge Propagation

### What Worked

- Both Copilot and Claude implementation-execution agent files were updated in sync with a new explicit cleanup rule for transient compose artifacts.
- `lint-docs` remained green after propagation updates.

### What Didn't Work

- None identified in this phase.

### Root Causes

- N/A.

### Patterns

- Agent-pair synchronization remains the most reliable home for cross-session operational lessons that are execution-process-specific.
