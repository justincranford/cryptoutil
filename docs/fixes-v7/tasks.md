# Remaining Tasks - fixes-v7 Followup

**Source**: Distilled from archived fixes-v7 (archive2/)
**Status**: Active

## Phase 1: E2E Verification ✅ COMPLETE

- [x] Run sm-im E2E — PASS (committed da860dd8)
- [x] Run jose-ja E2E — PASS (18.274s, committed 6086fb29)
- [x] Run sm-kms E2E — PASS (41.609s, committed 6086fb29)
- [x] Run identity E2E — PASS (6.823s, 5 services, committed 6086fb29)
- [x] Deep research: identified 13 root causes (see plan.md for details)
- [x] Fix root cause #1: OTel docker detector (committed 7b3b78c2)
- [x] Fix root cause #2: ComposeManager profiles (committed da860dd8)
- [x] Fix root cause #3: CLI args os.Args bug (committed da860dd8)
- [x] Fix root cause #4: Missing SQLite database URL (committed 6086fb29)
- [x] Fix root cause #5: Port override via CLI flag (committed 6086fb29)
- [x] Fix root cause #6: browser_session_jwks test (committed da860dd8)
- [x] Fix root cause #7: Docker image caching --build (committed 6086fb29)
- [x] Fix root cause #8: "start" vs "server" subcommand (committed 6086fb29)
- [x] Fix root cause #9: sm-kms postgres hostname (committed 6086fb29)
- [x] Fix root cause #10: GLOB CHECK SQLite-only (committed 6086fb29)
- [x] Fix root cause #11: BLOB type PostgreSQL (committed 6086fb29)
- [x] Fix root cause #12: DROP TABLE FK cascade (committed 6086fb29)
- [x] Fix root cause #13: Identity unseal secrets too short (committed 6086fb29)
- [x] Rewrite 20 identity config files to flat kebab-case (committed 6086fb29)
- [x] All 62 deployment validators pass

## Phase 2: Propagation Infrastructure ✅ COMPLETE

### 2.1 Reference Validation Script

- [x] Create `cicd validate-propagation` subcommand (committed 7eb73294)
- [x] Extract all `ARCHITECTURE.md#anchor` refs from .github/instructions/*.md
- [x] Extract all `ARCHITECTURE.md#anchor` refs from .github/agents/*.md
- [x] Resolve refs against actual ARCHITECTURE.md section headers
- [x] Report broken links (ref to non-existent section)
- [x] Report orphaned sections (## and ### level with zero refs)
- [x] Add tests for the validator (95.2% package coverage)
- [x] Fix broken anchor: formatgo → format_go in 03-01.coding.instructions.md
- [x] Result: 241 valid refs, 0 broken refs, 68 orphaned sections

### 2.2 Section 14 Instruction Coverage

- [x] Review ARCHITECTURE.md Section 14 content scope (33 lines, 5 subsections)
- [x] Add Operational Excellence cross-references to existing instruction files (committed 5d63f222)
- [x] 14.1 Monitoring & Alerting → 02-03.observability
- [x] 14.2 Incident Management → 06-01.evidence-based
- [x] 14.3 Performance Management → 02-03.observability
- [x] 14.4 Capacity Planning → 04-01.deployment
- [x] 14.5 Disaster Recovery → 04-01.deployment

### 2.3 ARCHITECTURE-INDEX.md Sync

- [x] Compare ARCHITECTURE-INDEX.md against current ARCHITECTURE.md section headers
- [x] Update all line number ranges (was based on 3356-line version, now 4219)
- [x] Add missing subsections: 6.10, 10.12, 12.5-12.10, 13.6-13.7 (committed b80c6d4d)
- [x] Update Quick Reference by Theme with new sections

## Phase 3: Propagation Quality ✅ COMPLETE

### 3.1 Lint Propagation Coverage

- [x] Extend `cicd validate-propagation` with per-level coverage statistics (committed fae1ea12)
- [x] Classify sections as High (##), Medium (###), Low (####) impact
- [x] Report coverage percentage per level and combined
- [x] Result: High 42%, Medium 49%, Combined 48%, Low 19%

### 3.2 Content Staleness Detection

- [x] Create `cicd validate-chunks` subcommand (committed 11a9d615)
- [x] Extract @propagate blocks from ARCHITECTURE.md (source of truth)
- [x] Extract @source blocks from instruction files (downstream copies)
- [x] Compare content byte-for-byte for staleness detection
- [x] Handle code fences correctly (skip outside, preserve inside propagate blocks)
- [x] Report match/mismatch/missing/file-not-found per chunk with line numbers
- [x] Result: 27 chunks validated, 27 matched, 0 stale
- [x] 94.6% package coverage, core functions at 100%

## Phase 4: Quality Gate Fixes ✅

### 4.1 cmd-main-pattern Linter

- [x] Identified root cause: regex expected os.Args, 12 main.go files use os.Args[1:]
- [x] Updated regex to accept both os.Args and os.Args[1:] patterns (committed 251b6e5a)
- [x] Added test cases for os.Args[1:] variant
- [x] TestLint_Integration passes, full test suite zero failures
