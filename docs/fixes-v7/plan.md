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

## Phase 1: E2E Verification (Blocked — Docker Desktop)

Verify the OTel collector config fix (`detectors: [env, system]`) actually
unblocks E2E tests. Requires Docker Desktop running.

- Run `go test -tags=e2e -timeout=30m ./internal/apps/sm/im/e2e/...`
- Verify sm-im E2E passes end-to-end

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
