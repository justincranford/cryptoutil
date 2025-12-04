# Archive List

This file lists the contents of `docs/archive/` with recommendations for deletion or retention.

## Summary
- `docs/archive/` contains the following top-level archive directories:
  - `cicd-refactoring-nov2025`
  - `codecov-nov2025`
  - `golangci-v2-migration-nov2025`
  - `identity-sprints`
  - `identityV1-legacy`

- Total files counted in `identity-sprints`: 165

---

## Per-archive recommendations

### `cicd-refactoring-nov2025`
- Files:
  - `COMPLETION-SUMMARY.md` (keeps record of tasks completed)
  - `planning/` (alignment analysis, plan)
  - `README.md`
- Recommendation: RETAIN for historical context and CI changes; move to `docs/knowledge` if needed.

### `codecov-nov2025`
- Files: Completion summary, README, tracking/*
- Recommendation: RETAIN `COMPLETION-SUMMARY.md` and `README.md`. Archive or DELETE tracking artifacts older than 6 months if no longer referenced.

### `golangci-v2-migration-nov2025`
- Files: migration docs, auto-fix plan, MIGRATION-COMPLETE.md, remaining-issues-tracker.md
- Recommendation: RETAIN `MIGRATION-COMPLETE.md` and `auto-fix-integration-plan.md`. Move `remaining-issues-tracker.md` into active `docs/todos-` only if issues remain; else DELETE.

### `identity-sprints`
- Files: many passthru*/ reports, postmortems, master plans.
- Recommendation:
  - RETAIN final `MASTER-PLAN.md` and `SESSION-SUMMARY-*.md` for each passthru that produced artifacts still referenced.
  - MOVE high-level summaries (e.g., `PROGRESS-REVIEW.md`, `README.md`) to `docs/identity/archived-sprints/`.
  - DELETE detailed interim artifacts older than 12 months if their contents were consolidated into `MASTER-PLAN` or `PROJECT-STATUS.md`.

### `identityV1-legacy`
- Files: original identity v1 docs
- Recommendation: ARCHIVE AS LEGACY (retain but mark `LEGACY`), do not delete. Keep as historical reference.

---

## Suggested actions (execute manually)
1. Review `docs/archive/identity-sprints/passthru1/` `TASK-12`, `TASK-14` COMPLETE files for duplication in `docs/identity/`.
2. Move high-level artifacts to `docs/knowledge/` and delete low-value interim files.
3. After pruning, run `git status` and commit the `ARCHIVE-LIST.md` and any moves.

If you want, I can open a branch and implement the suggested moves (create `docs/knowledge/` and migrate recommended files).
