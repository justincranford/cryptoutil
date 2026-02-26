# Remaining Work Plan - fixes-v7 Followup

**Status**: Active
**Created**: 2026-02-25
**Source**: Distilled from archived fixes-v7 (archive2/)

## Background

fixes-v7 completed 220/220 tasks across 11 phases. All 8 ARCHITECTURE.md gaps
from lessons.md were applied (ARCH-SUGGESTIONS.md). The propagation marker system
(`@source`/`@propagate`) was implemented across ARCHITECTURE.md and 18 instruction
files.

This plan captures the remaining open items and followup work identified during
the fixes-v7 strategy review.

## Phase 1: E2E Verification ✅ COMPLETE

Deep research uncovered 13 root causes across all 4 E2E test suites. All fixed
and committed.

### Root Causes Found

| # | Root Cause | Impact | Fix |
|---|-----------|--------|-----|
| 1 | OTel Docker detector requires socket | All services | `detectors: [env, system]` (7b3b78c2) |
| 2 | ComposeManager missing profile support | All E2E | Added Profiles field + --profile flags (da860dd8) |
| 3 | CLI args bug — os.Args instead of os.Args[1:] | 12 binaries | Fixed all main.go files (da860dd8) |
| 4 | Missing SQLite database URL | jose, identity | Added `-u sqlite://` to compose commands |
| 5 | Config format + port override | All services | Added `--bind-public-port=8080` CLI flag |
| 6 | browser_session_jwks test assumed rows | sm-im | Table-driven with `requireRow` flag (da860dd8) |
| 7 | Docker image caching | All E2E | Added `--build` to ComposeManager.Start() |
| 8 | "start" vs "server" subcommand | 8 compose + 3 Dockerfiles | Changed to "server" |
| 9 | sm-kms wrong postgres hostname | sm-kms | Fixed postgres-url.secret |
| 10 | GLOB CHECK in SQL migrations | sm-kms | Removed (app-layer validation) |
| 11 | BLOB type not valid PostgreSQL | sm-kms | Changed to BYTEA |
| 12 | DROP TABLE FK dependency | sm-kms | Added CASCADE via pre-drop |
| 13 | Identity unseal secrets too short | identity | Regenerated to 53 bytes |

### E2E Results

- **sm-im**: PASS (committed da860dd8)
- **jose-ja**: PASS (18.274s, committed 6086fb29)
- **sm-kms**: PASS (41.609s, committed 6086fb29)
- **identity**: PASS (6.823s, 5 services, committed 6086fb29)

### Additional Fixes (Config Format)

- Rewrote 20 identity config files from nested YAML to flat kebab-case
- Added jose-ja sqlite/postgresql config files with flat kebab-case format

## Phase 2: Propagation Infrastructure ✅ COMPLETE

Build tooling to validate and enforce the ARCHITECTURE.md propagation chain.

### 2.1 Reference Validation Script (R3 from strategy review) ✅

Created `cicd validate-propagation` subcommand (committed 7eb73294):
- Extracts all `ARCHITECTURE.md#anchor` references from instruction/agent files
- Validates against actual ARCHITECTURE.md section headers via GitHub-flavored anchor generation
- Reports broken links and orphaned sections (## and ### level)
- Result: 241 valid refs, 0 broken refs, 68 orphaned sections
- Fixed 1 broken anchor: `formatgo` → `format_go` in 03-01.coding.instructions.md
- 95.2% package coverage

### 2.2 Section 14 Instruction Coverage ✅

ARCHITECTURE.md Section 14 (Operational Excellence) is 33 lines with 5 subsections.
Added cross-references to existing instruction files (committed 5d63f222):
- 14.1 Monitoring & Alerting → 02-03.observability.instructions.md
- 14.2 Incident Management → 06-01.evidence-based.instructions.md
- 14.3 Performance Management → 02-03.observability.instructions.md
- 14.4 Capacity Planning → 04-01.deployment.instructions.md
- 14.5 Disaster Recovery → 04-01.deployment.instructions.md

### 2.3 ARCHITECTURE-INDEX.md Sync ✅

ARCHITECTURE-INDEX.md was stale (based on 3356-line version, now 4219 lines).
Regenerated with correct line numbers and new subsections (committed b80c6d4d):
- Updated all 16 section line ranges
- Added missing subsections: 6.10, 10.12, 12.5-12.10, 13.6-13.7
- Updated Quick Reference by Theme

## Phase 3: Propagation Quality ✅ COMPLETE

### 3.1 Lint Propagation Coverage (R4/R5 from strategy review) ✅

Extended `cicd validate-propagation` with per-level coverage statistics
(committed fae1ea12):
- Classifies sections as High (##), Medium (###), Low (####) impact
- Reports coverage percentage per level and combined
- Result: High 42%, Medium 49%, Combined 48%, Low 19%
- 94.6% package coverage maintained

### 3.2 Content Staleness Detection (R6 from strategy review) ✅

Created `cicd validate-chunks` subcommand (committed 11a9d615):
- Extracts @propagate blocks from ARCHITECTURE.md (source of truth)
- Extracts @source blocks from instruction files (downstream copies)
- Compares content byte-for-byte to detect stale propagated chunks
- Handles code fences correctly (skips outside, preserves inside propagate blocks)
- Reports match/mismatch/missing/file-not-found per chunk with line numbers
- Result: 27 chunks validated, 27 matched, 0 mismatched, 0 missing
- 94.6% package coverage, core extraction/validation functions at 100%
- Direct content comparison (no hash storage needed — simpler and more accurate)

## Phase 4: Quality Gate Fixes ✅

### 4.1 cmd-main-pattern Linter (pre-existing TestLint_Integration failure) ✅

Fixed cmd-main-pattern regex (committed 251b6e5a):
- Root cause: E2E fix #3 changed product/service main.go from os.Args to os.Args[1:]
- The cmd-main-pattern linter regex only matched `os.Args,` not `os.Args[1:],`
- 12 of 18 main.go files failed the pattern check → "lint-go completed with 1 errors"
- Fix: Updated regex to accept both `os.Args` and `os.Args[1:]` patterns
- Added test cases for os.Args[1:] pattern variant
- Result: TestLint_Integration now passes, full test suite has zero failures
