# Lessons - Framework V23: Follow-Up Work from Framework-V22

**Created**: 2026-05-17
**Last Updated**: 2026-05-22

## Pre-Execution: V22 Lessons to Enforce

Before starting implementation, the executor MUST enforce these lessons from framework-V22:

1. **Per-task status updates are MANDATORY**: Update `tasks.md` immediately after each task
   completes. NEVER accumulate multiple task completions before updating documentation. A
   `tasks.md` that does not reflect actual state is a blocking artifact inconsistency (V22 ISSUE-4).

2. **Docker Compose verification is MANDATORY within the same phase**: Any phase that modifies
   Compose files, configs consumed by containers, or cert paths MUST include a Docker Compose
   verification step (`docker compose up --wait` + health endpoint check) within the SAME phase.
   Phases declared complete without Compose verification are incomplete (V22 ISSUE-5).

## Phase 1: Integration Test Failures

_Populate after Phase 1 execution complete._

## Phase 2: pki-init E2E Compose — Docker Named Volumes for Cert Storage

_Populate after Phase 2 execution complete._

## Phase 3: POSTGRES_SECRETS_DIR Sync

_Populate after Phase 3 execution complete._

## Phase 4: sm-im E2E SKIP Reduction

_Populate after Phase 4 execution complete._

## Phase 5: Verification and Closure

_Populate after Phase 5 execution complete._
