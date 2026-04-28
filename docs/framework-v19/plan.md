# Implementation Plan - Framework V19: Reality Reconciliation and Backlog Closure

**Status**: Planning
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

## Quizme Round 1 (2026-04-28)

See `docs/framework-v19/quizme-v1.md` for contradiction-resolution decisions.
