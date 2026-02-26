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

## Phase 2: Propagation Infrastructure

Build tooling to validate and enforce the ARCHITECTURE.md propagation chain.

### 2.1 Reference Validation Script (R3 from strategy review)

Create a `cicd lint-docs validate-propagation` subcommand that:
1. Extracts all `See [ARCHITECTURE.md Section X.Y]` references from instruction/agent files
2. Resolves them against actual ARCHITECTURE.md section headers
3. Reports broken links (references to non-existent sections)
4. Reports stale references (section renumbered or removed)

### 2.2 Section 14 Instruction Coverage

ARCHITECTURE.md Section 14 (Operational Excellence) has zero instruction file
coverage. Either:
- Add Ops content to an existing instruction file (e.g., 04-01.deployment), or
- Create a new instruction file if Section 14 is substantial enough

### 2.3 ARCHITECTURE-INDEX.md Sync

Verify ARCHITECTURE-INDEX.md is current with ARCHITECTURE.md. If stale, update
or add a validator.

## Phase 3: Propagation Quality (Medium-Term)

### 3.1 Lint Propagation Coverage (R4/R5 from strategy review)

Extend `cicd lint-docs` to report:
- ARCHITECTURE.md sections with zero downstream references (currently ~63%)
- Instruction file coverage percentage (currently ~37% of sections referenced)
- Target: 60% coverage of high-impact sections

### 3.2 Content Hash Staleness Detection (R6 from strategy review)

For each `@source` marker, store SHA-256 hash of source content at sync time.
CI/CD check flags when source has changed but downstream hasn't been updated.
