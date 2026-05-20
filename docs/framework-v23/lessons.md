# Lessons - Framework V23

**Created**: 2026-05-17
**Last Updated**: 2026-05-20

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

_Populate after Phase 2 execution complete._

## Phase 3: POSTGRES_SECRETS_DIR Sync

_Populate after Phase 3 execution complete._

## Phase 4: Reduce sm-im E2E SKIP Cases

_Populate after Phase 4 execution complete._

## Phase 5: Verification and Closure

_Populate after Phase 5 execution complete._

## Phase 6: Knowledge Propagation

_Populate after Phase 6 execution complete._
