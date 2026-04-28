# Tasks - Framework V19: Reality Reconciliation and Backlog Closure

**Status**: 0 of 24 tasks complete (0%)
**Created**: 2026-04-28
**Last Updated**: 2026-04-28

## Task Status Legend

| Symbol | Meaning |
|--------|---------|
| ❌ | Not started |
| 🔄 | In progress |
| ✅ | Complete |
| ⏳ | Blocked |

## Quizme Decision Constraints (Applied)

1. Completion claims require verification against `tasks.md` and repository evidence.
2. Exclusion entries require active validation against current codebase state.
3. If completion or exclusion status is not verifiable, mark as unresolved instead of guessing.
4. Encoding corruption in planning docs must be repaired as UTF-8 while preserving semantics.
5. Phase completion is blocked when mandatory quality gates are unexecuted.
6. Seven-day session analysis uses `transcripts/*.jsonl` as primary evidence.

## Phase 1: Baseline Reconciliation

### Task 1.1: Build discrepancy matrix

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Compare v17 tasks statuses against real code/doc state.
  - [ ] Compare v18 tasks statuses against real code/doc state.
  - [ ] For each claimed completion, include direct evidence or mark "I don't know".
  - [ ] Archive matrix in `test-output/gap-analysis/v19-discrepancy-matrix.md`.

### Task 1.2: Validate exclusion reality

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Enumerate all active `knownExclusions` and exception maps in fitness linters.
  - [ ] Map each exclusion to a concrete unresolved repository condition.
  - [ ] For each exclusion, explicitly mark: required, stale-removed, or unresolved ("I don't know").
  - [ ] Archive in `test-output/gap-analysis/v19-exclusion-inventory.tsv`.

### Task 1.3: Repair v18 doc encoding

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Remove mojibake corruption from `docs/framework-v18/tasks.md` (and other impacted files if present).
  - [ ] Preserve semantic content.
  - [ ] Confirm UTF-8 clean output.

### Task 1.4: Reconcile v17 status semantics

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Resolve contradiction between header summary and per-task statuses in `docs/framework-v17/tasks.md`.
  - [ ] Every task has one unambiguous status.

### Task 1.5: Baseline gates

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go build ./...` passes.
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes.
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes.
  - [ ] Archive outputs in `test-output/v19-phase1/`.

## Phase 2: Incomplete Work Closure

### Task 2.1: Close swagger/test pattern gaps

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Resolve PS-ID gaps currently hidden by `apps_ps_id_swagger_presence` exclusions.
  - [ ] Resolve lifecycle/port-conflict gaps currently hidden by `apps_ps_id_test_patterns` exclusions.

### Task 2.2: Reduce server package exclusions

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Reassess and reduce `knownExclusionsPublicServer` where migration is complete.
  - [ ] Document any permanent exceptions with rationale.

### Task 2.3: Reconcile service_structure exclusions

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Reassess `service_structure` legacy exclusions.
  - [ ] Convert stale temporary exclusions to either completed migrations or explicitly permanent exceptions.

### Task 2.4: Align docs to completed migrations

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Ensure v17/v18 tasks and plan docs reflect real migrated/unmigrated state.
  - [ ] Remove claims that cannot be evidenced.

### Task 2.5: Phase 2 verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes with revised exclusions.
  - [ ] Targeted package tests pass for modified services.
  - [ ] Evidence archived in `test-output/v19-phase2/`.

## Phase 3: Session-Driven Efficiency Hardening

### Task 3.1: Analyze failed tool patterns

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Categorize last-7-days tool failures by root cause.
  - [ ] Archive in `test-output/gap-analysis/v19-session-failure-taxonomy.md`.

### Task 3.2: Add planning guardrails

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Add explicit "status claim requires evidence file" rule to relevant framework planning docs.
  - [ ] Add explicit "if unsure, use I don't know and keep status unresolved" rule.
  - [ ] Add "retry ceiling then strategy change" guidance.

### Task 3.3: Add contradiction-check task template

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Add reusable checklist that compares plan/tasks/lessons/code before phase completion.

### Task 3.4: Phase 3 verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes after documentation/process updates.
  - [ ] Evidence archived in `test-output/v19-phase3/`.

## Phase 4: Knowledge Propagation

### Task 4.1: ENG-HANDBOOK updates

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Add V19-proven guidance for status-evidence integrity and exclusion lifecycle management.

### Task 4.2: Instruction and skill sync

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Update relevant instruction/skill files if V19 introduces stable new workflow patterns.
  - [ ] Validate drift checks.

### Task 4.3: Propagation validation

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes.
  - [ ] Propagation checks pass with no drift.

### Task 4.4: Final quality suite

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go build ./...` passes.
  - [ ] `go test ./...` passes or blocked with documented infra dependency.
  - [ ] If any mandatory gate is blocked, phase remains blocked with mitigation path recorded.
  - [ ] `golangci-lint run ./...` passes.
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes.

### Task 4.5: Final documentation alignment

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `docs/framework-v19/plan.md`, `docs/framework-v19/tasks.md`, and executed evidence are consistent.
  - [ ] Contradiction decisions from quizme are reflected in plan/tasks.

## Evidence Archive

- `test-output/gap-analysis/`
- `test-output/completion-verification/`
- `test-output/v19-phase1/`
- `test-output/v19-phase2/`
- `test-output/v19-phase3/`
- `test-output/v19-phase4/`
