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

1. [Phase 0: Pre-flight Build Health](#phase-0-pre-flight-build-health) - Established a repeatable baseline for build and lint validation before migrations.
2. [Phase 1: ENG-HANDBOOK.md Documentation Propagation](#phase-1-eng-handbookmd-documentation-propagation) - Completed handbook propagation updates and validated with lint-docs.
3. [Phase 2: Prescriptive MANIFEST.yaml + Linter Extension](#phase-2-prescriptive-manifestyaml--linter-extension) - Extended the PS-ID template linter and MANIFEST enforcement model to support deeper structure checks.
4. [Phase 3: Identity Services Server Code Migration](#phase-3-identity-services-server-code-migration) - Migrated identity service server structure into MANIFEST-conformant locations.
5. [Phase 4: sm-im Root Cleanup](#phase-4-sm-im-root-cleanup) - Removed root-level server artifacts and repaired test placement under `server/`.
6. [Phase 5: Create Missing server/ Subdirectory Packages](#phase-5-create-missing-server-subdirectory-packages) - Added missing `server/model`, `server/repository/migrations`, and required config helper markers across identity services.
7. [Phase 6: Create Missing client/ Packages](#phase-6-create-missing-client-packages) - Added typed client packages for all required PS-IDs and aligned them to lint rules.
8. [Phase 7: Create Missing e2e/ Packages](#phase-7-create-missing-e2e-packages) - Added required `e2e/` scaffolding and tests for remaining PS-IDs.
9. [Phase 8: Remove knownExclusions + Final Validation](#phase-8-remove-knownexclusions--final-validation) - Removed temporary template/linter exclusions and reached clean build, lint, and non-race test gates.
10. [Phase 9: Knowledge Propagation](#phase-9-knowledge-propagation) - Updated handbook/target structure and synced Copilot+Claude skill guidance.

## Actions

1. Install a Windows C toolchain (`gcc`) in developer environments so `go test -race -count=2 ./...` can run locally without infrastructure blocking.
2. Add meaningful behavioral tests for newly added `client/`, `server/model/`, and `server/repository/` scaffolding packages to raise package-level coverage beyond placeholder thresholds.
3. Consolidate recurring magic-usage `const-redefine` informational findings in identity and PKI packages into a dedicated follow-up refactor plan.
4. Extend `apps_ps_id_template` test fixtures to generate required MANIFEST categories from parsed manifest metadata instead of hardcoded fixture assumptions.

---

## Phase 0: Pre-flight Build Health

**What Worked**:
- Running build and lint gates before migration changes quickly separated baseline issues from migration regressions.

**What Didn't Work**:
- Initial evidence capture was fragmented across multiple transient outputs.

**Root Causes**:
- Early execution focused on fixing blockers before normalizing evidence paths.

**Patterns for Future Phases**:
1. Create phase-specific `test-output/` directories before running any validation commands.

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

**What Worked**:
- The `psIDExclusions` struct design (7 maps bundled into one value) eliminated a potential 7-parameter function signature. Using a struct keeps the call site readable and makes it easy to add an 8th exclusion category in the future without changing all call sites.
- Conditional `checkE2EFiles` logic (skip entirely if `e2e/` dir absent) was the right design choice. Alternatives (registering all ~5 services with no e2e/ as exclusions) would have created a large, noisy exclusion list that grows inverse to migration progress. The conditional approach becomes stricter automatically as e2e/ directories are created.
- `__SERVICE__` placeholder substitution for e2e file names (e.g., `__SERVICE___e2e_test.go` → `sm-kms_e2e_test.go`) leverages the existing PSID → Service mapping in `AllProductServices()`. No extra registry data needed.
- All 14 tests passed on the first run after implementation (0 failures, 0 regressions). 100% statement coverage.
- `lint-fitness` confirmed real workspace passes with initial knownExclusions in place — the exclusion maps correctly model the current (transitional) state of each PS-ID directory structure.
- `golangci-lint run --fix` caught one gofumpt violation in the new exclusion map initializer block (blank lines between map entries). Auto-fixed cleanly with no secondary violations.

**What Didn't Work**:
- Task 2.1 was marked ❌ in tasks.md but was actually completed in the prior session. The compaction summary correctly noted it was done, but the tasks.md file still had the ❌ marker. Required a "verify before implementing" check at session start to avoid duplicate work.
- The initial `ExportedCheckInDirNoExclusions` function in `export_test.go` needed updating from the old 4-parameter `checkInDirWithExclusions(logger, rootDir, rootExcl, serverExcl)` signature to the new `psIDExclusions` struct. The export seam pattern isolates this change correctly (no production code paths changed).

**Root Causes**:
- tasks.md out of sync with actual work completed (Phase 1 → Phase 2 boundary occurred across sessions). Mitigation: verify each task's actual state by reading source files before marking in-progress.
- gofumpt requires blank lines between map literal groups to be consistent with its group-separation rules. This is a style enforcement that `golangci-lint run --fix` handles automatically, but it means a single `--fix` pass is always needed after adding new map initializers.

**Patterns for Future Phases**:
1. **Verify task pre-completion before coding**: When resuming from a compaction, read the relevant source files to confirm what was already done vs. what the tasks.md shows. Tasks can be completed in prior sessions without tasks.md being updated.
2. **psIDExclusions pattern**: When adding new MANIFEST validation categories that will temporarily fail for some PS-IDs, add the exclusion to `psIDExclusions` (not to MANIFEST.yaml). MANIFEST.yaml is the canonical target state; exclusions in the Go linter are the transition mechanism.
3. **Conditional check for optional directories**: When a required directory (e2e/) doesn't exist yet for many services, make the check conditional (`if !dirExists { return nil }`) rather than populating a large exclusion list. This auto-tightens as migration progresses.
4. **golangci-lint two-pass always**: After adding new struct literals with map initializers, run `golangci-lint run --fix` first (for gofumpt/wsl), then re-run `golangci-lint run` without `--fix` to catch any secondary violations introduced by auto-fixers.
5. **Evidence file naming**: Use `test-output/v18v19-phase2/` pattern consistently. The subdirectory name encodes the plan version (v18v19 = ENG-HANDBOOK v18 → v19 migration) and the phase number, making it easy to correlate with tasks.md evidence archive.

---

## Phase 3: Identity Services Server Code Migration

**What Worked**:
- Moving identity server files into MANIFEST-expected layouts resolved structural drift quickly.

**What Didn't Work**:
- Some migrated tests referenced private root helpers and required package boundary fixes.

**Root Causes**:
- Earlier root-level layouts leaked package-private assumptions into tests.

**Patterns for Future Phases**:
1. When relocating tests, immediately convert them to exported entry points and `*_test` package boundaries.

---

## Phase 4: sm-im Root Cleanup

**What Worked**:
- Rehousing sm-im lifecycle and HTTP tests under `server/` restored template conformance and reduced root clutter.

**What Didn't Work**:
- Formatting and package-name mismatches reappeared after file moves.

**Root Causes**:
- Moved files preserved old package declarations and style assumptions.

**Patterns for Future Phases**:
1. Immediately run gofumpt and package-targeted tests after any file relocation.

---

## Phase 5: Create Missing server/ Subdirectory Packages

**What Worked**:
- Adding `server/model`, `server/repository/migrations`, and SQL migration headers for identity services satisfied linter and template checks.

**What Didn't Work**:
- Initial `config_test_helper.go` implementations duplicated existing symbols.

**Root Causes**:
- Marker-file requirement was misread as requiring functional helpers instead of presence-only files.

**Patterns for Future Phases**:
1. For structural template requirements, use minimal marker files unless logic is explicitly required.

---

## Phase 6: Create Missing client/ Packages

**What Worked**:
- Uniform typed client scaffolding made cross-PS-ID implementation fast and consistent.

**What Didn't Work**:
- New client files triggered `errcheck`, `wsl_v5`, and `literal-use` blockers.

**Root Causes**:
- Small style/literal constraints were missed during initial scaffold creation.

**Patterns for Future Phases**:
1. Include close-error handling and shared magic constants in the initial scaffold template.

---

## Phase 7: Create Missing e2e/ Packages

**What Worked**:
- Adding required `testmain_e2e_test.go` and service e2e tests unblocked template compliance.

**What Didn't Work**:
- Template test fixtures lagged new required e2e expectations and failed unit tests.

**Root Causes**:
- Fixture generators encoded old manifest assumptions.

**Patterns for Future Phases**:
1. Update linter fixture builders in the same change whenever MANIFEST requirements are tightened.

---

## Phase 8: Remove knownExclusions + Final Validation

**What Worked**:
- Removing exclusion maps and updating MANIFEST required dirs exposed true remaining conformance gaps.

**What Didn't Work**:
- Full race validation was blocked by local toolchain (`gcc`) absence.

**Root Causes**:
- Windows environment lacked C compiler required for Go race detector.

**Patterns for Future Phases**:
1. Treat race checks as infrastructure prerequisites and verify toolchain availability during pre-flight.

---

## Phase 9: Knowledge Propagation

**What Worked**:
- Updating `ENG-HANDBOOK.md`, `target-structure.md`, and paired Copilot/Claude skill files in one pass kept propagation drift minimal.

**What Didn't Work**:
- Late-stage documentation updates required retroactive alignment with already-completed code phases.

**Root Causes**:
- Execution prioritized code/lint blockers first; doc updates were deferred to final phase.

**Patterns for Future Phases**:
1. Update canonical docs immediately when template or structural invariants change to reduce end-phase sync work.
