# Framework-V22 Post-Analysis

Created: 2026-05-17
Analyst: Post-completion review
Source documents: plan.md, tasks.md, lessons.md, EXEC-SUMMARY.md, DETAIL-SUMMARY.md

## Overview

Framework-V22 executed 11 phases against the goal "Fix 10 issues from V21 audit SUMMARY.md."
70 of 71 tasks are complete. This document flags all issues identified during post-completion
review. Each issue is categorized: **INCOMPLETE** (work not done), **INTENTIONAL DEFERRAL**
(explicitly scoped out), **PROCESS VIOLATION** (execution discipline failure), or
**UNTRACKED FOLLOW-UP** (discovered but not planned).

---

## Issues

### ISSUE-1 — INCOMPLETE: Integration Tests Failing in 3 Packages

**Category**: INCOMPLETE  
**Severity**: Blocking (this is the 71st unchecked task in tasks.md)

**Description**: The cross-cutting task `[ ] Integration tests pass` in tasks.md is explicitly
unchecked. Three packages have known integration test failures as of the final session:

| Package | Failure Mode |
|---------|-------------|
| `internal/apps/sm-im/client` | TLS unknown authority — client cert not trusted |
| `internal/apps/sm-kms/client` | Missing Authorization header or connection timeout |
| `internal/apps/sm-kms/server/repository/orm` | Cleanup references missing `barrier_content_keys` table |

**Impact**: `go test -tags integration ./...` does not pass. The plan's completion criteria is not
met. Framework-V23 must include fixes for these three packages as Task 1.

**Evidence**: tasks.md line ~770 (cross-cutting section), lessons.md Phase 9 post-mortem.

---

### ISSUE-2 — INTENTIONAL DEFERRAL: Race Detector Not Run Locally

**Category**: INTENTIONAL DEFERRAL  
**Severity**: Low (by design)

**Description**: The cross-cutting task `[ ] Race detector clean: go test -race ./...` is
explicitly unchecked. The lessons.md and tasks.md both note this as scoped to Linux CI because
the race detector requires CGO (`CGO_ENABLED=1`) and a C toolchain (gcc), which are not present
on the local Windows development machine. Per the CGO Ban policy (ENG-HANDBOOK §11.1.2), CGO
prerequisites are NEVER installed on Windows dev machines — race detection on Windows is
therefore deferred entirely to CI/CD.

**Where race checks are triggered in the repository**:

| Trigger Location | Details |
|------------------|---------|
| `ci-race.yml` (GitHub Actions) | `ubuntu-latest`, `CGO_ENABLED=1`, `go test -race -count=5 ./...` — authoritative gate |
| `beast-mode.agent.md` | Listed in "RECOMMENDED Pre-Push Quality Gates" — Linux/macOS only, no platform caveat |
| `implementation-planning.agent.md` | Listed in Per-Phase Quality Gates and Cross-Cutting Tasks template — Linux/macOS only |
| `03-02.testing.instructions.md` | Listed in Test Execution section as reference — Linux/macOS only |
| Pre-commit / pre-push hooks | NOT present (would fail on Windows; correctly excluded) |

**Docker container option (evaluated, deferred)**: Race detection could run on Windows inside a
Docker container (`docker run --rm -v ${PWD}:/workspace golang:1.26.1 sh -c "cd /workspace &&
go test -race -count=2 ./..."`). This is technically viable but adds Docker startup overhead
for a check already enforced authoritatively in `ci-race.yml`. Standardising all three platforms
(Windows, Linux, macOS) on a container-based approach would add complexity for no practical gain.
Deferred until the community identifies Windows-only race conditions that CI/CD cannot catch.

**ENG-HANDBOOK update**: §10.9 has been updated with a platform support matrix, Windows note,
Docker container option evaluation, and a trigger inventory. See that section for the canonical
documentation.

**Disposition**: This is an explicit, documented deferral — not a gap. The `ci-race.yml` workflow
runs this check. It is not a follow-up task for framework-V23. The agent/instruction files that
list race detection in quality gates without a platform caveat are a minor documentation gap;
they are low priority compared to Phase 1–5 V23 work and can be addressed as a V24 editorial
task if desired.

**Action required**: None for V23. ENG-HANDBOOK §10.9 has been updated. Documented here for
completeness.

---

### ISSUE-3 — UNTRACKED FOLLOW-UP: Three Action Items From lessons.md Not In Any Plan

**Category**: UNTRACKED FOLLOW-UP  
**Severity**: Medium

**Description**: lessons.md recorded four action items that were not tracked in any plan. Item 1
(LF line endings) has since been resolved: LF line endings are now enforced in ALL text files
including `*.secret` files via `.gitattributes * text=auto eol=lf` and the `mixed-line-ending`
pre-commit hook. The remaining three are tracked in Framework-V23.

1. **pki-init E2E: switch to Docker named volumes for cert storage** — All 10 PS-ID compose
   files and the canonical template in `api/cryptosuite-registry/templates/` use OS filesystem
   bind mounts (`./certs:/certs:rw`) for cert delivery. This is non-compliant with deployment
   rules CO-21/CO-22 (see `docs/deployment-templates.md`) and ENG-HANDBOOK §6.11.3 which
   mandates named Docker volumes (`{ps-id}-certs:/certs`). On Windows, bind-mounted host
   directories can be locked from a prior run, causing `pki-init` cert generation to fail.
   The correct fix is to use named Docker volumes, eliminating the host filesystem dependency
   entirely. This affects all 10 `deployments/{ps-id}/compose.yml` files and the canonical
   template.

   _(Previously recorded as "E2E cert-dir writable cleanup" with the wrong fix approach of
   adding a writable-check in `test_orch_e2e`. The root cause is bind mounts, not test
   orchestration. Named volumes are the correct solution.)_

2. **POSTGRES_SECRETS_DIR parameterization sync** — Keep the `POSTGRES_SECRETS_DIR` variable
   in sync across the shared-postgres template (`deployments/shared-postgres/`) and per-PS-ID
   `.env.postgres` instantiations. A drift here caused the PostgreSQL container to look in the
   wrong secrets directory.

3. **sm-im E2E SKIP reduction** — Track and reduce E2E test SKIP usage in
   `internal/apps/sm-im/e2e/`. The 2026-05-17 reconciliation commit documented at least 4 SKIP
   cases. These were analyzed but not resolved.

**Action required**: Framework-V23 includes tasks for all three remaining items.

---

### ISSUE-4 — PROCESS VIOLATION: tasks.md Documentation Lag

**Category**: PROCESS VIOLATION  
**Severity**: Low (post-hoc remediation completed)

**Description**: DETAIL-SUMMARY.md records that tasks.md was not kept current during execution.
Phase headers were marked "TODO" while the underlying work was already complete. A reconciliation
commit on 2026-05-15 was required to bring tasks.md back into alignment with actual state.

This created a period where the plan artifact triad (plan.md + tasks.md + lessons.md) was
internally inconsistent — a blocker condition per ENG-HANDBOOK Section 11.

**Impact**: Temporary. The reconciliation commit resolved the inconsistency. No tasks were
actually missed; the documentation did not reflect reality.

**Root cause**: The agent executing Phases 4-11 accumulated multiple task completions before
committing task status updates. The correct pattern is task-complete → update tasks.md → commit
→ next task.

**Action required**: Reminder in framework-V23 lessons.md to enforce immediate per-task status
updates. No code changes needed.

---

### ISSUE-5 — PROCESS VIOLATION: Five Post-Implementation Docker Fixes Required

**Category**: PROCESS VIOLATION  
**Severity**: Medium

**Description**: EXEC-SUMMARY.md documents that after the plan was declared complete, five Docker
infrastructure blockers were found and required fixes:

1. CRLF line endings in Windows-generated secret files (fixed by adding LF enforcement)
2. PostgreSQL credential lookup in wrong directory (fixed by correcting `POSTGRES_SECRETS_DIR`)
3. `pki-init` cert-dir write permission failure (fixed by pre-run cleanup step)
4. Init readiness race (fixed by adding explicit health-check dependency)
5. GORM DSN path mismatch (fixed by correcting path construction)

All five required Compose restarts and debugging cycles post-declaration-of-completion.

**Root cause**: Insufficient Docker Compose integration testing during the implementation phases.
The architecture instruction states: "Phases that modify Docker Compose files... MUST include a
Docker Compose verification step within the same phase." This was not enforced during V22
implementation.

**Action required**: Framework-V23 and V24 plans must explicitly require Docker Compose
verification within each phase that touches Compose files. This is already a rule in
`04-01.deployment.instructions.md`; the follow-up is enforcement at phase-gate time.

---

### ISSUE-6 — STRUCTURAL GAP: Framework-V22 Did Not Advance 10-PS-ID Convergence

**Category**: STRUCTURAL GAP  
**Severity**: High

**Description**: Framework-V22's scope was "Fix 10 issues from V21 audit SUMMARY.md." This is
infrastructure maintenance, not convergence work. After V22, the 10 PS-ID services remain
divergent:

- `sm-kms/server/` still has `businesslogic/`, `handler/`, `repository/` (pre-refactor structure)
- `pki-ca/server/` still has `cmd/`, `config/`, `middleware/` only
- Other PS-IDs have varying directory structures
- Code duplication between PS-IDs is unmeasured

The user's core requirement — all 10 PS-IDs as thin callers into framework with maximum reuse —
was not advanced by V22. V22 fixed prerequisites (test infrastructure, linter tooling) but left
the convergence problem open.

**Action required**: Framework-V24 is the primary vehicle for convergence. It must:
1. Baseline the current directory structure of all 10 PS-IDs
2. Define the canonical thin-caller structure
3. Migrate PS-IDs to canonical structure (pilot first, then all 10)
4. Enforce conformity via MANIFEST-driven fitness checks

---

## Summary Table

| Issue | Category | Severity | Framework-V23 Task? | Framework-V24 Task? |
|-------|----------|----------|--------------------|--------------------|
| ISSUE-1: Integration tests failing | INCOMPLETE | Blocking | Yes — Task 1 | No |
| ISSUE-2: Race detector deferred | INTENTIONAL DEFERRAL | Low | No (ENG-HANDBOOK §10.9 updated) | No |
| ISSUE-3: 3 untracked follow-ups (item 1 resolved) | UNTRACKED FOLLOW-UP | Medium | Yes — Tasks 2-4 | No |
| ISSUE-4: tasks.md documentation lag | PROCESS VIOLATION | Low | Lessons note only | No |
| ISSUE-5: 5 post-impl Docker fixes | PROCESS VIOLATION | Medium | Lessons note only | Yes — phase gate |
| ISSUE-6: No convergence progress | STRUCTURAL GAP | High | No | Yes — primary scope |

## What Framework-V22 Completed Well

For balance: 70 of 71 tasks were completed. Specific achievements:

- Test infrastructure helpers (`test_help_db`, `test_help_bootstrap`, `test_help_tls`,
  `test_orch_integration`) are fully implemented and shared across all 10 PS-IDs.
- golangci-lint passes with zero violations across the entire monorepo.
- lint-fitness passes all fitness checks including `apps-ps-id-template` exact-match enforcement.
- Mutation testing ≥ 98% on all infrastructure/utility packages.
- Unit test coverage ≥ 98% on all infrastructure/utility packages.
- E2E tests pass with Docker Desktop for all 10 PS-IDs.
- ENG-HANDBOOK.md and instruction files updated with all new patterns.

The infrastructure work in V22 is the prerequisite for the convergence work in V24. It is
complete. The issue is that convergence itself requires a dedicated plan.
