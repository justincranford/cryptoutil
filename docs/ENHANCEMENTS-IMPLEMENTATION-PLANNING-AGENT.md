# Enhancements For `claude-implementation-planning`

Created: 2026-05-17

## Summary

The planning agent is useful, but it is overbuilt for many tasks. It spends a lot of space on internal process, token tracking, and repeated gates, while the real value is in producing a clean plan/tasks/lessons triad with clear scope and validation.

## What To Keep

- The plan/tasks/lessons triad is the right artifact set.
- The triad consistency gate is valuable.
- The baseline commit requirement is useful for long-running planning sessions.
- The scope-isolated blocker protocol is important.

## What To Modify

### 1. Tighten The Planning Output Contract

The agent should require a small fixed set of planning sections:

- scope
- assumptions
- affected files
- risks
- validation plan
- rollback or fallback

That would make plans easier to review and reduce the chance of oversized, low-signal prose.

### 2. Reduce Repetition Around Baseline Checks

The baseline commit rule appears more than once. Keep it once, keep it prominent, and remove the redundant restatements.

### 3. Make `quizme` Truly Conditional

The quiz file is useful only when there are unresolved decisions. It should be explicitly framed as an exception path, not a default artifact.

That would keep planning fast for tasks that are already well constrained.

### 4. Add A Clear Scope Boundary For Planning-Only Work

The agent already says planning-only requests should report blockers only. It should also say that implementation suggestions are not part of the planning deliverable unless the user asks for them.

That keeps the output focused and prevents accidental drift into execution commentary.

### 5. Simplify The Plan Artifact Lifecycle

The current doc is heavy on meta-rules. A smaller lifecycle would be easier to operate:

- create plan/tasks/lessons
- record the baseline commit
- update the triad after each phase
- reconcile at the end

The rest can live in the handbook.

## What To Remove Or De-Emphasize

- The token-usage tracking requirement should not be in the main planning contract unless a specific session needs it.
- The detailed planning examples are longer than they need to be.
- The repeated warnings about not stopping are redundant once the execution contract is stated clearly.
- The artifact triad gate should be concise and checklist-like rather than a second planning system.

## Suggested Additions

- A short "plan quality" checklist that requires a user-visible outcome for each phase.
- A short "validation intent" field in tasks.md so each task explicitly states how success will be checked.
- A small default section for "speed risks" so likely slow areas are called out before implementation starts.

## Net Effect

The planning agent would still be strict, but it would be less bulky and easier to apply to smaller tasks without turning planning into its own project.
