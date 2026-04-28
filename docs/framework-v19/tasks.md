# Tasks - Framework V19: Reality Reconciliation and Backlog Closure

**Status**: 0 of 24 tasks complete (0%)
**Created**: 2026-04-28
**Last Updated**: 2026-04-28

## Task Status Legend

| Symbol | Meaning |
|--------|---------|
| Ã¢ÂÅ’ | Not started |
| Ã°Å¸â€â€ž | In progress |
| Ã¢Å“â€¦ | Complete |
| Ã¢ÂÂ³ | Blocked |

## Phase 1: Baseline Reconciliation

### Task 1.1: Build discrepancy matrix

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Compare v17 tasks statuses against real code/doc state.
  - [ ] Compare v18 tasks statuses against real code/doc state.
  - [ ] Archive matrix in `test-output/gap-analysis/v19-discrepancy-matrix.md`.

### Task 1.2: Validate exclusion reality

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Enumerate all active `knownExclusions` and exception maps in fitness linters.
  - [ ] Map each exclusion to a concrete unresolved repository condition.
  - [ ] Archive in `test-output/gap-analysis/v19-exclusion-inventory.tsv`.

### Task 1.3: Repair v18 doc encoding

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Remove mojibake corruption from `docs/framework-v18/tasks.md` (and other impacted files if present).
  - [ ] Preserve semantic content.
  - [ ] Confirm UTF-8 clean output.

### Task 1.4: Reconcile v17 status semantics

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Resolve contradiction between header summary and per-task statuses in `docs/framework-v17/tasks.md`.
  - [ ] Every task has one unambiguous status.

### Task 1.5: Baseline gates

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] `go build ./...` passes.
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes.
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes.
  - [ ] Archive outputs in `test-output/v19-phase1/`.

## Phase 2: Incomplete Work Closure

### Task 2.1: Close swagger/test pattern gaps

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Resolve PS-ID gaps currently hidden by `apps_ps_id_swagger_presence` exclusions.
  - [ ] Resolve lifecycle/port-conflict gaps currently hidden by `apps_ps_id_test_patterns` exclusions.

### Task 2.2: Reduce server package exclusions

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Reassess and reduce `knownExclusionsPublicServer` where migration is complete.
  - [ ] Document any permanent exceptions with rationale.

### Task 2.3: Reconcile service_structure exclusions

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Reassess `service_structure` legacy exclusions.
  - [ ] Convert stale temporary exclusions to either completed migrations or explicitly permanent exceptions.

### Task 2.4: Align docs to completed migrations

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Ensure v17/v18 tasks and plan docs reflect real migrated/unmigrated state.
  - [ ] Remove claims that cannot be evidenced.

### Task 2.5: Phase 2 verification

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes with revised exclusions.
  - [ ] Targeted package tests pass for modified services.
  - [ ] Evidence archived in `test-output/v19-phase2/`.

## Phase 3: Session-Driven Efficiency Hardening

### Task 3.1: Analyze failed tool patterns

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Categorize last-7-days tool failures by root cause.
  - [ ] Archive in `test-output/gap-analysis/v19-session-failure-taxonomy.md`.

### Task 3.2: Add planning guardrails

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Add explicit "status claim requires evidence file" rule to relevant framework planning docs.
  - [ ] Add "retry ceiling then strategy change" guidance.

### Task 3.3: Add contradiction-check task template

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Add reusable checklist that compares plan/tasks/lessons/code before phase completion.

### Task 3.4: Phase 3 verification

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes after documentation/process updates.
  - [ ] Evidence archived in `test-output/v19-phase3/`.

## Phase 4: Knowledge Propagation

### Task 4.1: ENG-HANDBOOK updates

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Add V19-proven guidance for status-evidence integrity and exclusion lifecycle management.

### Task 4.2: Instruction and skill sync

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] Update relevant instruction/skill files if V19 introduces stable new workflow patterns.
  - [ ] Validate drift checks.

### Task 4.3: Propagation validation

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes.
  - [ ] Propagation checks pass with no drift.

### Task 4.4: Final quality suite

- **Status**: Ã¢ÂÅ’
- **Acceptance Criteria**:
  - [ ] `go build ./...` passes.
  - [ ] `go test ./...` passes or blocked with documented infra dependency.
  - [ ] `golangci-lint run ./...` passes.
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes.

### Task 4.5: Final documentation alignment

- **Status**: Ã¢ÂÅ’
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
