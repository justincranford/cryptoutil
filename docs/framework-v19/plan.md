# Implementation Plan - Framework V19: Reality Reconciliation and Backlog Closure

**Status**: Complete
**Created**: 2026-04-28
**Last Updated**: 2026-04-28
**Purpose**: Reconcile plan documents (v16-v18), actual codebase state, and recent session execution evidence to close work that is marked complete but is incomplete, incorrect, or inefficient.

## Overview

Framework V19 is a corrective implementation plan. It is not a rewrite of v16-v18. It focuses on:
1. Closing measurable gaps between documented completion and real repository state.
2. Applying lessons from v16-v18 lessons files that were not operationalized in code, tasks, or quality gates.
3. Hardening execution discipline based on last-7-days transcript evidence.

## Evidence Inputs

- `docs/framework-v16/plan.md`
- `docs/framework-v16/tasks.md`
- `docs/framework-v16/lessons.md`
- `docs/framework-v17/plan.md`
- `docs/framework-v17/tasks.md`
- `docs/framework-v17/lessons.md`
- `docs/framework-v18/plan.md`
- `docs/framework-v18/tasks.md`
- `docs/framework-v18/lessons.md`
- `test-output/gap-analysis/framework-doc-line-counts.tsv`
- `test-output/gap-analysis/chat-logs-last7d-raw.txt`
- `test-output/gap-analysis/chat-sessions-last7d-summary.tsv`
- `test-output/gap-analysis/chat-tool-failures-last7d.tsv`

## High-Risk Findings Carried Into V19

1. `docs/framework-v17/tasks.md` remains mostly `Status: ❌` despite `37 of 43 complete` summary and completed narrative in `docs/framework-v17/lessons.md`.
2. `docs/framework-v18/tasks.md` has severe mojibake/encoding corruption (status symbols and punctuation degraded), reducing review reliability.
3. V18 claims "temporary knownExclusions removed" but linter code still contains substantial active exclusions:
   - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_test_patterns/apps_ps_id_test_patterns.go`
   - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_swagger_presence/apps_ps_id_swagger_presence.go`
   - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_server_package/apps_ps_id_server_package.go`
4. Last-7-days transcript evidence shows repeated tool failures and retry loops, including failed `create_file` attempts and clustered `list_dir` failures (`test-output/gap-analysis/chat-tool-failures-last7d.tsv`).
5. Chat telemetry in `debug-logs` is metadata-only for content review; transcript JSONL is the authoritative source for seven-day session analysis.

## Prior Lessons To Apply in V19 (Not Stored in V19 lessons.md)

- From v18: verify infrastructure preconditions (race detector toolchain) before gating phase completion.
- From v18: update docs near-time with structural/code changes, not only at final propagation phase.
- From v17: separate per-check exclusion maps and remove entries immediately when migrated.
- From v16: verify actual filesystem state before writing acceptance criteria and status claims.

## Technical Context

- **Language**: Go 1.26.1
- **Quality tools**: `go test`, `go build`, `golangci-lint`, `go run ./cmd/cicd-lint lint-fitness`, `go run ./cmd/cicd-lint lint-docs`
- **Primary code areas**:
  - `internal/apps-tools/cicd_lint/lint_fitness/**`
  - `internal/apps/{identity-*,sm-im,sm-kms,pki-ca,jose-ja,skeleton-template}/**`
- **Primary docs areas**:
  - `docs/framework-v17/**`
  - `docs/framework-v18/**`
  - `docs/ENG-HANDBOOK.md`

## Phases

### Phase 1: Baseline Reconciliation

**Objective**: Produce a source-of-truth matrix for v17/v18 claims vs code reality.

- Reconcile every v17 task status with objective evidence.
- Normalize and repair v18 document encoding issues where needed.
- Produce a v19 gap matrix for remaining exclusion-backed deviations.

**Success**:
- Discrepancy matrix complete and linked in tasks evidence.
- No unresolved contradiction between task status and code state for v17/v18.

### Phase 2: Incomplete Work Closure

**Objective**: Execute unresolved work currently hidden behind exclusions or stale status claims.

- Reduce `knownExclusions` maps to only intentional, documented exceptions.
- Close missing lifecycle/port-conflict/swagger/testmain coverage gaps for PS-IDs still excluded.
- Align `framework-v17`/`framework-v18` task files with actual completion state.

**Success**:
- `lint-fitness` passes with a minimized, justified exception set.
- All updated docs reflect real state.

### Phase 3: Session-Driven Efficiency Hardening

**Objective**: Apply lessons from transcript analysis to reduce failed/redundant operations.

- Add procedural guardrails for repeated failing tool patterns.
- Enforce evidence-first status updates for planning docs.
- Add explicit verification sequence before marking tasks complete.

**Success**:
- Reduced retry/failure loops in workflow.
- V19 tasks include evidence requirements for each completion claim.

**Mandatory Guardrails (Added in execution):**
- Status transitions to complete require an evidence artifact path under `test-output/`.
- When verification is inconclusive, record "I don't know" and keep the item unresolved.
- Retry ceiling is three attempts per failing tool/operation; then change strategy.
- Phase completion requires contradiction-check across plan/tasks/lessons/code before marking done.

**Contradiction-Check Template:**
1. Compare phase claims in `plan.md` with task states in `tasks.md`.
2. Compare both against repository reality (code/tests/docs).
3. Verify referenced evidence artifacts exist and are readable.
4. If any mismatch exists, keep status unresolved and document explicitly.

### Phase 4: Knowledge Propagation

**Objective**: Promote validated V19 patterns into permanent artifacts.

- Update `docs/ENG-HANDBOOK.md` and relevant instruction/skill files where V19 reveals missing guidance.
- Validate propagation integrity.

**Success**:
- `go run ./cmd/cicd-lint lint-docs` passes.
- All propagation changes committed with evidence.

## Quality Gates

- `go build ./...` and `go build -tags e2e,integration ./...` pass.
- `go test ./...` passes (or blocked with explicit infra dependency and mitigation).
- `golangci-lint run ./...` passes.
- `go run ./cmd/cicd-lint lint-fitness` passes with explicit documented exceptions only.
- `go run ./cmd/cicd-lint lint-docs` passes.

## Evidence Strategy

All V19 execution evidence must be archived under:
- `test-output/gap-analysis/`
- `test-output/completion-verification/`
- `test-output/v19-phase*/`

## Quizme Decisions Applied

1. Task completion truth is evidence-driven, not narrative-driven.
1. Requirement: V19 execution MUST validate task completion against `tasks.md` plus repository evidence.
1. Requirement: If evidence is insufficient, status remains unresolved ("I don't know") until verified.
1. Exclusions are validity-checked, not assumed.
1. Requirement: Every active exclusion MUST be validated against current codebase state.
1. Requirement: Stale exclusions MUST be removed immediately; remaining exclusions MUST include explicit rationale.
1. Encoding corruption handling is repair-in-place with semantic preservation.
1. Requirement: Mojibake in planning docs MUST be repaired to UTF-8 clean text while preserving original task intent.
1. Infrastructure blockers are blocking gates.
1. Requirement: A phase cannot be marked complete when mandatory gates are unexecuted due to missing prerequisites.
1. Requirement: Blocked gates require explicit mitigation path in plan/tasks.
1. Session evidence hierarchy is transcript-first.
1. Requirement: `transcripts/*.jsonl` is the primary source for substantive review.
1. Requirement: `debug-logs/*` and `models.json` are secondary metadata/index sources.

## Quizme Round 1 (2026-04-28)

### Question 1: Source of Truth When Task Status and Narrative Conflict

**Question**: If a tasks file shows many `Status: ❌` entries but lessons/executive narrative claims completion, which source should govern V19 execution gating?

**A)** Trust lessons/executive narrative as final truth.
**B)** Average both sources and proceed if most sections look complete.
**C)** Treat task-level status plus code evidence as primary truth, and downgrade narrative claims when inconsistent. (Probably recommended - newer reliability rule)
**D)** Ignore both and rely only on latest chat memory.
**E)** YOU MUST VALIDATE IF THE TASK WAS COMPLETED!!! DON'T GUESS. DON'T ASK ME QUESTIONS YOU ARE ABLE TO ANSWER YOURSELF. CHECK THE TASKS FILE AND THE CODEBASE AND THEN DECIDE IF THE TASK WAS COMPLETED OR NOT. THIS IS A CRITICAL PART OF YOUR JOB AND YOU MUST NOT SKIP IT. IF YOU ARE UNSURE, SAY "I DON'T KNOW" INSTEAD OF MAKING AN UNINFORMED GUESS. REMEMBER, YOUR DECISIONS IMPACT THE QUALITY OF THE WORK AND THE TRUST OF THE USERS. BE DILIGENT AND THOROUGH IN YOUR REVIEW PROCESS.

**Answer**: E

### Question 2: Exclusions Policy for Lint Fitness

**Question**: When plan docs say temporary exclusions were removed, but linter code still contains active exclusion maps, what should V19 do?

**A)** Keep exclusions as-is if lint currently passes.
**B)** Hide exclusions from docs and keep code unchanged.
**C)** Re-audit each exclusion, remove stale entries immediately, and document only intentional permanent exceptions. (Probably recommended - newer code-first policy)
**D)** Remove all exclusions immediately, even if it breaks builds, and fix later.
**E)** YOU MUST VALIDATE IF THE EXCLUSION IS STILL REQUIRED!!! DON'T GUESS. DON'T ASK ME QUESTIONS YOU ARE ABLE TO ANSWER YOURSELF. CHECK THE CODEBASE AND THEN DECIDE IF THE EXCLUSION IS STILL NEEDED OR NOT. THIS IS A CRITICAL PART OF YOUR JOB AND YOU MUST NOT SKIP IT. IF YOU ARE UNSURE, SAY "I DON'T KNOW" INSTEAD OF MAKING AN UNINFORMED GUESS. REMEMBER, YOUR DECISIONS IMPACT THE QUALITY OF THE WORK AND THE TRUST OF THE USERS. BE DILIGENT AND THOROUGH IN YOUR REVIEW PROCESS.

**Answer**: E

### Question 3: Handling Encoding-Corrupted Planning Documents

**Question**: For mojibake-corrupted status symbols/text in planning docs, what is the correct V19 handling?

**A)** Leave as-is if humans can still infer meaning.
**B)** Rewrite from scratch without preserving original semantics.
**C)** Repair encoding to UTF-8 clean text while preserving exact task intent and evidence references. (Probably recommended - newer documentation integrity practice)
**D)** Move corrupted docs to archive and stop using them.
**E)**

**Answer**: C

### Question 4: Quality Gate Completion Under Infrastructure Blockers

**Question**: If race tests are blocked by missing local toolchain (for example gcc on Windows), when can a phase be marked complete?

**A)** Always complete if other checks pass.
**B)** Complete and mention blocker only in lessons.
**C)** Mark blocked until blocker is resolved or approved alternate evidence path is documented in tasks and plan. (Probably recommended - newer strict gate semantics)
**D)** Remove race gate from plan permanently.
**E)**

**Answer**: C

### Question 5: Session Evidence Source for 7-Day Chat Analysis

**Question**: For analyzing work quality from past sessions, which source should V19 trust most?

**A)** `debug-logs/*/main.jsonl` only.
**B)** `models.json` only.
**C)** `transcripts/*.jsonl` as primary, with `debug-logs` as metadata index. (Probably recommended - newer evidence path)
**D)** User recollection only.
**E)**

**Answer**: C; I don't know if this is the right answer, but it seems to be the most reasonable choice given the context.
