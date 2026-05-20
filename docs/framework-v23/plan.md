# Framework V23

**Status**: Not started
**Created**: 2026-05-17
**Last Updated**: 2026-05-20
**Purpose**: Close open items identified in post-analysis from the prior framework execution.
These are items that were incomplete or untracked at the time of the prior framework declaration.

## Background

The prior framework executed successfully but left open issues captured in the post-analysis:

- ISSUE-1: Integration test failures in 3 packages — all 3 confirmed passing at V23 plan
  creation (2026-05-17); Phase 1 provides explicit pre-flight verification before Phase 2.
- ISSUE-3: Three lessons.md action items never promoted to tasks (item 1 — LF line endings —
  resolved separately via `.gitattributes * text=auto eol=lf`).

This plan promotes those items to tracked, executable tasks.

## Pre-Execution: Lessons to Enforce

Before starting implementation, the executor MUST enforce these lessons:

1. **Per-task status updates are MANDATORY**: Update `tasks.md` immediately after each task
   completes. NEVER accumulate multiple task completions before updating documentation. A
   `tasks.md` that does not reflect actual state is a blocking artifact inconsistency.

2. **Docker Compose verification is MANDATORY within the same phase**: Any phase that modifies
   Compose files, configs consumed by containers, or cert paths MUST include a Docker Compose
   verification step (`docker compose up --wait` + health endpoint check) within the SAME phase.
   Phases declared complete without Compose verification are incomplete.

## Scope

### Phase 1: Verify ISSUE-1 Resolved

Three test packages were reported failing when ISSUE-1 was filed. All three were confirmed
passing at V23 plan creation (2026-05-17). Phase 1 provides an explicit pre-flight
confirmation run before Phase 2 begins.

1. `internal/apps/sm-im/client/` — verified passing at plan creation.
2. `internal/apps/sm-kms/client/` — verified passing at plan creation.
3. `internal/apps/sm-kms/server/repository/orm/` — verified passing at plan creation.

**Acceptance**:
- `go test -tags integration ./internal/apps/sm-im/client/...` exits 0.
- `go test -tags integration ./internal/apps/sm-kms/client/...` exits 0.
- `go test -tags integration ./internal/apps/sm-kms/server/repository/orm/...` exits 0.
- Evidence archived in `test-output/v23-phase1/`.

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
- CO-21/CO-22 validators implemented in `lint_deployments`; `lint-deployments` fails on
  pre-change compose files (bind mounts) and passes on post-change files (named volumes).
- All 10 PS-ID compose files use `{ps-id}-certs:/certs` named volumes (zero bind mounts for certs).
- Canonical template uses `__PS_ID__-certs:/certs` named volume.
- `go run ./cmd/cicd-lint lint-deployments` exits 0 (zero CO-21/CO-22 violations after migration).
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

Exactly 2 E2E test cases in `internal/apps/sm-im/e2e/` are currently skipped with `t.Skip`:

1. `e2e_registration_test.go:93` — framework limitation: `join_tenant_id` not yet supported
   by the registration handler (create_tenant flow only).
2. `e2e_test.go:67` — explicitly intentional: OTEL Collector health port 13133 not exposed
   to the host (intentional — prevents port conflicts across deployments).

This phase:
- Audits both `t.Skip` calls and documents their root cause.
- Skip 1 (`join_tenant_id`): either fix the framework registration handler or convert to a
  named constant (`skipReasonJoinTenantIDNotSupported`) for explicit tracking.
- Skip 2 (OTEL port): convert to a named constant (`skipReasonOtelPortNotExposed`) to make
  the intentional design decision self-documenting. This skip MUST NOT be removed.

**Acceptance**:
- All `t.Skip` calls reference a named constant (not inline strings).
- Skip 2 (OTEL port) remains as a named-constant skip — intentional, must not be removed.
- `go test -tags e2e ./internal/apps/sm-im/e2e/...` passes (with Docker infra running).
- Evidence archived in `test-output/v23-phase4/`.

### Phase 5: Verification and Closure

- `go test -tags integration ./...` passes.
- `go run ./cmd/cicd-lint lint-deployments` passes (includes new validators).
- `go run ./cmd/cicd-lint lint-fitness` passes.
- `go run ./cmd/cicd-lint lint-docs` passes.
- `git status --porcelain` returns empty.

### Phase 6: Knowledge Propagation

Review `lessons.md` populated during Phases 1–5 and apply insights to permanent artifacts.

This phase:
- Reviews all phase post-mortems in `lessons.md`.
- Updates `docs/ENG-HANDBOOK.md` with new patterns or architectural decisions discovered.
- Updates `.github/instructions/*.instructions.md` where new coding/testing patterns apply.
- Updates `.github/agents/*.agent.md` and `.github/skills/*/SKILL.md` where guidance improved.
- Verifies propagation: `go run ./cmd/cicd-lint lint-docs` exits 0.

**Acceptance**:
- All artifact updates committed with separate semantic commits per artifact type.
- `go run ./cmd/cicd-lint lint-docs` exits 0.
- Evidence archived in `test-output/v23-phase6/`.

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
- `golangci-lint run --build-tags e2e,integration ./...`
- `go run ./cmd/cicd-lint lint-go`
- `go run ./cmd/cicd-lint lint-deployments`
- `go run ./cmd/cicd-lint lint-fitness`
- `go run ./cmd/cicd-lint lint-docs`

## ENG-HANDBOOK.md Cross-References

| Topic | ENG-HANDBOOK.md Section | When to Reference |
|-------|------------------------|-------------------|
| Integration Testing | [Section 10.3](../../docs/ENG-HANDBOOK.md#103-integration-testing-strategy) | Phase 1 verification |
| E2E Testing | [Section 10.4](../../docs/ENG-HANDBOOK.md#104-e2e-testing-strategy) | Phase 4 skip reduction |
| Quality Gates | [Section 11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates) | ALL phases |
| Code Quality | [Section 11.3](../../docs/ENG-HANDBOOK.md#113-code-quality-standards) | ALL phases |
| Deployment Architecture | [Section 12](../../docs/ENG-HANDBOOK.md#12-deployment-architecture) | Phase 2 named volumes |
| Docker Compose Rules | [Section 12.1](../../docs/ENG-HANDBOOK.md#121-docker-compose-rules) | Phase 2 CO-21/CO-22 |
| Coding Standards | [Section 14.1](../../docs/ENG-HANDBOOK.md#141-coding-standards) | Phases 1-4 code changes |
| Version Control | [Section 14.2](../../docs/ENG-HANDBOOK.md#142-version-control) | ALL phases |
| Knowledge Propagation | [Section 14.8](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation) | Phase 6 |

## Evidence Strategy

- `test-output/v23-phase1/` — integration test failure evidence and fix verification.
- `test-output/v23-phase2/` — Docker named volumes for cert storage: compose diff and lint evidence.
- `test-output/v23-phase3/` — POSTGRES_SECRETS_DIR audit evidence.
- `test-output/v23-phase4/` — sm-im E2E skip reduction evidence.
- `test-output/v23-phase5/` — final verification evidence.
- `test-output/v23-phase6/` — knowledge propagation evidence.
