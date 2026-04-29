# Lessons - Framework V19: Reality Reconciliation and Backlog Closure

**Created**: 2026-04-28
**Purpose**: Populated during V19 execution only. Do not store pre-V19 lessons in this file.

## Executive Summary

1. [Phase 1: Baseline Reconciliation](#phase-1-baseline-reconciliation) - Built an evidence-backed discrepancy matrix, repaired v18 encoding corruption, and reconciled v17/v18 status semantics.
2. [Phase 2: Incomplete Work Closure](#phase-2-incomplete-work-closure) - Removed stale exclusion entries, retained only evidence-backed exceptions, and validated fitness/tests after each reduction.
3. [Phase 3: Session-Driven Efficiency Hardening](#phase-3-session-driven-efficiency-hardening) - Converted transcript-derived failure patterns into explicit evidence, uncertainty, retry, and contradiction-check guardrails.
4. [Phase 4: Knowledge Propagation](#phase-4-knowledge-propagation) - Propagated V19 guardrails into ENG-HANDBOOK, instruction, and skill artifacts with lint/docs validation evidence.

## Actions

1. Audit `apps_ps_id_template` exclusion maps and remove stale entries once V20 migrations land (starting with `swagger.go` and `__SERVICE___lifecycle_test.go` exclusions).
2. Add a lightweight CI check that fails when planning docs contain mojibake markers (`Ã`, `â`, `Â`) to prevent encoding regressions.
3. Introduce a reusable script that generates exclusion inventory TSV from linter maps plus filesystem probes to reduce manual reconciliation drift.
4. Run a full v17 task-by-task replay evidence pass and convert historical unresolved statuses into explicit completed/blocked states with artifacts.

> Per-phase structure (mandatory):
> What Worked
> What Did Not Work
> Root Causes
> Patterns for Future Phases

## Phase 1: Baseline Reconciliation

### What Worked

- Running preflight gates before edits exposed a clean technical baseline and reduced uncertainty.
- Building a discrepancy matrix with explicit evidence pointers kept reconciliation objective.
- Repairing `docs/framework-v18/tasks.md` with deterministic cp1252->utf8 conversion preserved semantics.

### What Did Not Work

- Initial PowerShell attempts for encoding repair and regex scanning were brittle due shell parsing/Unicode handling.
- Early log redirection produced UTF-16 output artifacts that are harder to inspect quickly.

### Root Causes

- Historical docs mixed narrative summaries and task-level states without enforced evidence coupling.
- Encoding was degraded by prior write paths that did not enforce UTF-8 consistency.

### Patterns for Future Phases

- Always generate a discrepancy matrix before reconciling plan/task status claims.
- Treat status contradictions as blockers, not documentation polish items.
- Prefer deterministic conversion strategy previews before writing repaired files.

## Phase 2: Incomplete Work Closure

### What Worked

- Direct filesystem probes by PS-ID gave fast, objective evidence for exclusion decisions.
- Removing exclusions incrementally and rerunning targeted tests prevented overcorrection.
- Capturing required/stale-removed states in TSV made decisions reviewable.

### What Did Not Work

- Some legacy comments in linter files lagged current behavior and created analysis noise.
- Broad wording like "all removed" in prior docs masked nuanced permanent exceptions.

### Root Causes

- Exclusion lifecycle was not treated as a first-class maintenance workflow.
- Migration completion and documentation updates were previously decoupled.

### Patterns for Future Phases

- Maintain an exclusion inventory artifact every time maps are edited.
- Keep exceptions narrowly scoped and tied to concrete unresolved repository conditions.
- Remove stale entries immediately after evidence confirms conformance.

## Phase 3: Session-Driven Efficiency Hardening

### What Worked

- Transcript-first analysis produced concrete failure categories and directly actionable guardrails.
- Embedding retry ceilings and contradiction-check templates into planning docs improved execution discipline.

### What Did Not Work

- Prior sessions used repeated retries without method changes, increasing noise and latency.
- Metadata logs were occasionally treated as if equivalent to transcript content.

### Root Causes

- No explicit retry strategy boundaries were enforced.
- Evidence hierarchy was implied rather than codified.

### Patterns for Future Phases

- Use transcript JSONL as primary evidence and debug logs as index-only metadata.
- Enforce three-attempt retry ceiling followed by mandatory strategy change.
- Require explicit "I don't know" unresolved status when verification remains inconclusive.

## Phase 4: Knowledge Propagation

### What Worked

- A minimal, focused propagation set (ENG-HANDBOOK + evidence-based instruction + mirrored skill docs) reduced drift risk.
- Keeping `.github` and `.claude` skill updates in lockstep avoided downstream drift lint failures.

### What Did Not Work

- Large artifact surfaces increase chance of partial propagation if updates are not scoped.

### Root Causes

- Process guidance had existed in fragments but lacked consolidated V19-specific wording for status-evidence integrity and exclusion lifecycle.

### Patterns for Future Phases

- Add propagation updates immediately after pattern validation, not at end-of-session only.
- Keep propagation edits concise and behavior-focused to reduce maintenance overhead.
