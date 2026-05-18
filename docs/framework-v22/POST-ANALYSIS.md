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
the race detector requires CGO (`CGO_ENABLED=1`) and gcc, which are not present on the local
Windows development machine.

**Disposition**: This is an explicit, documented deferral — not a gap. The `ci-race.yml` workflow
runs this check. It is not a follow-up task for framework-V23.

**Action required**: None. Documented here for completeness.

---

### ISSUE-3 — UNTRACKED FOLLOW-UP: Four Action Items From lessons.md Not In Any Plan

**Category**: UNTRACKED FOLLOW-UP  
**Severity**: Medium

**Description**: lessons.md records four action items that are not tracked in any plan or
tasks document. These were identified during implementation and recorded as lessons but were
never promoted to tasks:

1. **LF line endings in secrets files** — Add a deployment fitness check for LF line endings in
   `deployments/*/secrets/*.secret` files. CRLF line endings in secret files caused the Windows
   PostgreSQL credential failure that required one of the 5 post-implementation Docker fixes.

2. **E2E cert-dir writable cleanup** — Add E2E orchestration check for cert-dir writable status
   before Compose startup on Windows. The `pki-init` container cannot write certificates if the
   `certs/` directory is owned or locked from a previous run.

3. **POSTGRES_SECRETS_DIR parameterization sync** — Keep the `POSTGRES_SECRETS_DIR` variable
   in sync across the shared-postgres template (`deployments/shared-postgres/`) and per-PS-ID
   `.env.postgres` instantiations. A drift here caused the PostgreSQL container to look in the
   wrong secrets directory.

4. **sm-im E2E SKIP reduction** — Track and reduce E2E test SKIP usage in
   `internal/apps/sm-im/e2e/`. The 2026-05-17 reconciliation commit documented at least 4 SKIP
   cases. These were analyzed but not resolved.

**Action required**: Framework-V23 should include tasks for all four items. They are
straightforward but were left out of scope during V22.

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
| ISSUE-2: Race detector deferred | INTENTIONAL DEFERRAL | Low | No | No |
| ISSUE-3: 4 untracked follow-ups | UNTRACKED FOLLOW-UP | Medium | Yes — Tasks 2-5 | No |
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
