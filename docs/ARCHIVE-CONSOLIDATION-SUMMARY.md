# Archive Consolidation Summary

**Date**: December 21, 2025
**Purpose**: Track extraction and consolidation of archive files into copilot instructions and speckit documents

---

## Content Mapping Analysis

### Archive Files Already Covered in Copilot Instructions

**File**: GAP-ANALYSIS-2025-01-10.md
- **Content Type**: Coverage targets, constitutional compliance, gap assessment
- **Covered In**:
  - 01-04.testing.instructions.md: Coverage targets (95%/98%), mutation testing (≥80/98%)
  - 05-01.evidence-based-completion.instructions.md: Evidence requirements, gap tracking
- **Action**: DELETE after verifying constitution.md has coverage requirements

**File**: TEST-PERFORMANCE-ANALYSIS.md
- **Content Type**: Test timing analysis, GitHub vs local execution discrepancy
- **Covered In**:
  - 01-04.testing.instructions.md: Test timing targets (<15s unit, <45s e2e per package)
  - 02-01.github.instructions.md: GitHub Actions performance considerations
- **Action**: DELETE after extracting timing patterns to testing.instructions.md

**File**: TIMEOUT-FIXES-ANALYSIS.md
- **Content Type**: PostgreSQL timeout, health check timeout, network operation patterns
- **Covered In**:
  - 01-04.testing.instructions.md: Timeout configuration section (5s+ for network ops, 10s+ for TLS)
  - 02-01.github.instructions.md: PostgreSQL service requirements
- **Action**: DELETE after verifying coverage in instructions

**File**: SESSION-2025-01-08-LESSONS-LEARNED.md
- **Content Type**: Race condition patterns, timeout fixes, flaky test prevention
- **Covered In**:
  - 01-04.testing.instructions.md: Race condition prevention section
  - 07-01.anti-patterns.instructions.md: Race testing patterns
- **Action**: DELETE after verifying content coverage

**File**: SESSION-2025-01-08-RACE-FIXES.md
- **Content Type**: Specific race condition fixes, concurrency bugs
- **Covered In**:
  - 07-01.anti-patterns.instructions.md: Race condition anti-patterns
- **Action**: DELETE after extracting unique patterns

**File**: CGO-BAN-ENFORCEMENT.md
- **Content Type**: CGO ban policy, enforcement patterns
- **Covered In**:
  - 01-05.golang.instructions.md: CGO Ban section (CRITICAL, MANDATORY CGO_ENABLED=0)
- **Action**: DELETE (fully covered)

**File**: MUTATION-TESTING-FIXES.md
- **Content Type**: Mutation testing patterns, gremlins usage
- **Covered In**:
  - 01-04.testing.instructions.md: Mutation testing section (mandatory, ≥80/98%)
- **Action**: DELETE after verifying coverage

---

## Session Files - Extract to DETAILED.md Timeline

All SESSION-*.md files should have their key findings extracted to `specs/002-cryptoutil/implement/DETAILED.md` Section 2 timeline:

- SESSION-2025-12-08-PHASE4.md → DETAILED.md timeline entry
- SESSION-2025-12-08-RESTART3.md → DETAILED.md timeline entry
- SESSION-2025-12-09-CI-FIXES.md → DETAILED.md timeline entry
- SESSION-2025-12-09-TASK-3-FINAL-SUMMARY.md → DETAILED.md timeline entry
- SESSION-2025-12-09-TASK-3-IDENTITY-COVERAGE.md → DETAILED.md timeline entry
- SESSION-2025-12-09-WORKFLOW-FIXES.md → DETAILED.md timeline entry
- SESSION-2025-12-10-TASK-7-KMS-HANDLER-ANALYSIS.md → DETAILED.md timeline entry
- SESSION-COVERAGE-IMPROVEMENTS.md → DETAILED.md timeline entry
- SESSION-MFA-COVERAGE-PROGRESS.md → DETAILED.md timeline entry

**Pattern**: For each session file, create timeline entry in format:

```markdown
### YYYY-MM-DD: Session Title
- Work completed: Summary (commit hashes)
- Key findings: Discoveries or blockers
- Coverage/quality metrics: Before/after
- Violations found: Issues discovered
- Next steps: Follow-up needed
- Related commits: [hash] description
```

---

## Workflow Analysis Files - Extract Timing Patterns

**Files**:
- workflow-analysis.md
- WORKFLOW-clientauth-TEST-TIMES.md
- WORKFLOW-jose-server-TEST-TIMES.md
- WORKFLOW-jose-TEST-TIMES.md
- WORKFLOW-OVERHEAD-ANALYSIS.md
- WORKFLOW-sqlrepository-TEST-TIMES.md

**Action**: Extract timing patterns and optimization insights to 01-04.testing.instructions.md

**Key Patterns to Extract**:
- GitHub Actions 2.5-3.3× slower than local
- Parallel test execution with t.Parallel()
- Target <15s unit tests per package
- Total unit test suite <180s
- Health check timeout strategies

---

## Speckit Files - Extract to 06-01.speckit.instructions.md

**Files**:
- SPECKIT-ITERATION-1-REVIEW.md
- SPECKIT-PROGRESS.md

**Action**: Extract speckit workflow lessons to 06-01.speckit.instructions.md

**Key Patterns to Extract**:
- Iterative spec refinement
- Evidence-based completion
- Implementation-driven constraints
- Feedback loop patterns

---

## Speckit Passthrough Files

**Files**:
- docs/speckit/passthru02/grooming/GROOMING-SESSION-02.md
- docs/speckit/passthru03/grooming/GROOMING-SESSION-03.md

**Action**: Review for grooming insights, extract to clarify.md if Q&A content

---

## Files Requiring Special Handling

**File**: README.md (docs/archive/)
- **Action**: Review for unique content not in main docs/README.md
- **Decision**: DELETE if redundant, otherwise merge to docs/README.md

---

## Consolidation Strategy

### Phase 1: Quick Wins (Already Covered)

Delete files where content is 100% covered in copilot instructions:

1. CGO-BAN-ENFORCEMENT.md ✅ Fully covered in 01-05.golang.instructions.md
2. MUTATION-TESTING-FIXES.md (verify first, then delete)

### Phase 2: Extract Missing Content

For files with partial coverage, extract unique content:

1. GAP-ANALYSIS-2025-01-10.md → Extract coverage justification patterns
2. TEST-PERFORMANCE-ANALYSIS.md → Extract GitHub timing multipliers
3. TIMEOUT-FIXES-ANALYSIS.md → Extract health check timeout strategies
4. SESSION-2025-01-08-LESSONS-LEARNED.md → Verify race patterns covered
5. SESSION-2025-01-08-RACE-FIXES.md → Extract unique concurrency bugs

### Phase 3: Timeline Extraction

Extract all SESSION-*.md files to DETAILED.md Section 2 timeline

### Phase 4: Workflow Analysis Consolidation

Extract workflow timing analysis to testing.instructions.md

### Phase 5: Speckit Consolidation

Extract speckit workflow lessons to speckit.instructions.md

### Phase 6: Commit and Verify

Commit all changes with pre-commit hooks, verify all archive files deleted

---

## Token Efficiency Note

Given 917,797 remaining tokens and NO TIME PRESSURE (Speckit directive), I'll process files systematically without rushing. Quality and completeness are prioritized over speed.

---

## Next Steps

1. Review each archive file systematically
2. Extract unique content to appropriate instruction files
3. Update constitution.md/spec.md/clarify.md as needed
4. Create DETAILED.md timeline entries for session files
5. Delete archive files after verification
6. Commit with conventional format: `chore(docs): consolidate archived documentation`
