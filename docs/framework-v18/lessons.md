# Lessons — ENG-HANDBOOK.md Propagation + Prescriptive MANIFEST + Identity Conformance Migration

**Created**: 2026-04-27
**Purpose**: Phase post-mortem lessons — populated by the execution agent after each phase's quality gates pass.

> **MANDATORY per-phase structure** (4 sections per phase):
>
> **What Worked**: Patterns, tools, or decisions that accelerated the work or prevented issues.
>
> **What Didn't Work**: Friction points, incorrect assumptions, or approaches that required rework.
>
> **Root Causes**: Underlying causes of the "What Didn't Work" items.
>
> **Patterns for Future Phases**: Actionable takeaways for subsequent phases or future plans.

---

## Executive Summary

*(To be filled at plan completion — numbered links to each phase section with one-sentence outcome)*

## Actions

*(To be filled at plan completion — numbered list of concrete follow-up items for reviewer)*

---

## Phase 0: Pre-flight Build Health

*(To be filled during Phase 0 execution using the 4-section structure above)*

---

## Phase 1: ENG-HANDBOOK.md Documentation Propagation

**What Worked**:
- The conversation summary was dense enough to resume without re-reading source docs — `docs/claude-structure.md` and `docs/deployment-templates.md` section references were accurate.
- `grep -n "### 14.11"` pattern to find ENG-HANDBOOK.md anchor worked reliably; section anchors are stable identifiers across edits.
- `lint-docs` provides fast pass/fail signal (< 0.02s). All checks including `check-chunk-verification`, `validate-chunks`, `validate-propagation`, `propagation-coverage`, `lint-agent-drift`, `lint-skill-command-drift`, `lint-agent-self-containment`, and `lint-architecture-links` passed on the first run after §14.11 and §B.7 additions.
- `ORPHANED SECTIONS (135 of 496)` in `validate-propagation` output are informational WARNs, not failures — important to distinguish from blocking ERRORs. The EXECUTION SUMMARY line is the canonical pass/fail signal.
- Writing §14.11 as multiple numbered subsections (§14.11.1–§14.11.7) was more maintainable than a flat wall-of-prose approach. Makes the section scannable and future-editable.
- PowerShell `2>&1 | tee` captured lint evidence to `test-output/v18v19-phase1/` correctly; `grep_search` with `includeIgnoredFiles: true` needed to verify evidence files in `test-output/` (excluded from normal search scope).

**What Didn't Work**:
- Session context was compacted mid-phase (after §13.6 edits, before §14.11 work). The compaction summary was very thorough but required careful verification of "items pending" vs "items done" before resuming. Items 1.28-1.38 were still pending at compaction point, causing risk of double-applying already-done items.
- The tasks.md `**Status**: 0 of 49 tasks complete` was not updated during execution; all Phase 1 task status markings required a single bulk update at the end of the phase. This is a side-effect of the cross-session compaction but makes progress harder to track during execution.
- The `lint-architecture-links` check `Extracted 439 anchors` is a useful diagnostic but can give false confidence — it only validates anchors in ENG-HANDBOOK.md link targets, not that the prose above each table is well-structured.

**Root Causes**:
- Long session + large file (7400+ lines of ENG-HANDBOOK.md) triggers compaction. Compaction is unavoidable but the recovery overhead (re-reading summary, verifying what was done) is real.
- tasks.md updates at phase-end (not per-task) is a consequence of the compaction boundary occurring mid-phase. Future phases should update tasks.md immediately after each item to reduce recovery overhead.

**Patterns for Future Phases**:
1. **Always verify anchor existence before replacement**: Use `grep_search` to confirm the exact text of the oldString before calling `replace_string_in_file`. Long documents have very subtle differences between visual similarity and exact match.
2. **Save evidence immediately after each subtask**: `tee test-output/v18v19-phase1/lint-docs-TIMESTAMP.txt` after every lint-docs run. Avoids re-running after compaction.
3. **Cross-session task recovery**: When resuming after compaction, read tasks.md first (not summary), then verify each "✅ Complete" claim by searching ENG-HANDBOOK.md for the actual content before marking additional tasks. Prevents double-application.
4. **§B.7 table format discovery**: The original table had `Inputs` and `Outputs` columns that were sparsely populated. Replaced with a cleaner 2-column `Action | Description` format consistent with the rest of ENG-HANDBOOK.md appendix tables. The old format was vestigial from an early catalog draft.
5. **lint-docs exit code in PowerShell**: PowerShell 5.1 emits a `NativeCommandError` to the error stream when `go run` writes to stderr (even INFO-level output). This causes `$LASTEXITCODE = 1` in the pipeline even when lint-docs succeeds. Always verify via the `✅ SUCCESS` / `❌ FAILED` EXECUTION SUMMARY line, not the exit code alone.

---

## Phase 2: Prescriptive MANIFEST.yaml + Linter Extension

*(To be filled during Phase 2 execution using the 4-section structure above)*

---

## Phase 3: Identity Services Server Code Migration

*(To be filled during Phase 3 execution using the 4-section structure above)*

---

## Phase 4: sm-im Root Cleanup

*(To be filled during Phase 4 execution using the 4-section structure above)*

---

## Phase 5: Create Missing server/ Subdirectory Packages

*(To be filled during Phase 5 execution using the 4-section structure above)*

---

## Phase 6: Create Missing client/ Packages

*(To be filled during Phase 6 execution using the 4-section structure above)*

---

## Phase 7: Create Missing e2e/ Packages

*(To be filled during Phase 7 execution using the 4-section structure above)*

---

## Phase 8: Remove knownExclusions + Final Validation

*(To be filled during Phase 8 execution using the 4-section structure above)*

---

## Phase 9: Knowledge Propagation

*(To be filled during Phase 9 execution using the 4-section structure above)*
