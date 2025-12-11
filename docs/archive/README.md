# Documentation Archive

**Purpose**: Historical documentation moved here after content consolidation
**Archive Date**: December 10, 2025
**Reason**: Cleanup to reduce documentation sprawl and improve navigability

---

## Archive Structure

### `/sessions/` - Session-Specific Documentation

Historical session notes documenting work completed in specific chat/development sessions:

- `SESSION-2025-01-08-LESSONS-LEARNED.md` - Testing and race condition lessons
- `SESSION-2025-01-08-RACE-FIXES.md` - Race detector fixes
- `SESSION-2025-12-08-PHASE4.md` - Phase 4 implementation notes
- `SESSION-2025-12-08-RESTART3.md` - Session restart documentation
- `SESSION-2025-12-09-CI-FIXES.md` - CI workflow fixes
- `SESSION-2025-12-09-TASK-3-FINAL-SUMMARY.md` - Task 3 completion
- `SESSION-2025-12-09-TASK-3-IDENTITY-COVERAGE.md` - Identity coverage work
- `SESSION-2025-12-09-WORKFLOW-FIXES.md` - Workflow debugging
- `SESSION-2025-12-10-TASK-7-KMS-HANDLER-ANALYSIS.md` - KMS handler analysis
- `SESSION-COVERAGE-IMPROVEMENTS.md` - Coverage improvement strategies
- `SESSION-MFA-COVERAGE-PROGRESS.md` - MFA coverage progress

**Status**: Historical reference only. Lessons learned have been integrated into:

- `.github/instructions/*.md` (copilot instructions)
- `.specify/memory/constitution.md` (core principles)
- `specs/001-cryptoutil/implement/EXECUTIVE.md` (post mortem section)

### `/workflow-analysis/` - Workflow Performance Analysis

Performance analysis documents for specific GitHub Actions workflows:

- `WORKFLOW-clientauth-TEST-TIMES.md` - clientauth package test performance
- `WORKFLOW-jose-server-TEST-TIMES.md` - jose/server package test performance
- `WORKFLOW-jose-TEST-TIMES.md` - jose package test performance
- `WORKFLOW-OVERHEAD-ANALYSIS.md` - Overall workflow overhead analysis
- `WORKFLOW-sqlrepository-TEST-TIMES.md` - sqlrepository package test performance

**Status**: Historical baselines. Current performance tracking in:

- `specs/001-cryptoutil/SLOW-TEST-PACKAGES.md` (consolidated metrics)
- Phase 0 tasks in `specs/001-cryptoutil/implement/DETAILED.md`

### `/speckit/` - Spec Kit Iteration Reviews

Historical Spec Kit iteration reviews and progress tracking:

- `SPECKIT-ITERATION-1-REVIEW.md` - Iteration 1 completion review
- `SPECKIT-PROGRESS.md` - Historical progress tracker (superseded)

**Status**: Iteration 1 complete. Active tracking in:

- `specs/001-cryptoutil/implement/DETAILED.md` (timeline section)
- `specs/001-cryptoutil/implement/EXECUTIVE.md` (stakeholder view)

### Root Archive Files

- `CGO-BAN-ENFORCEMENT.md` - Historical CGO ban enforcement documentation
  - **Integrated into**: `.specify/memory/constitution.md` Section II.A
  - **Integrated into**: `.github/instructions/01-03.golang.instructions.md`

- `MUTATION-TESTING-FIXES.md` - Mutation testing optimization strategies
  - **Integrated into**: `.github/instructions/01-02.testing.instructions.md`

---

## Why Archive Instead of Delete?

These documents contain valuable historical context:

- **Decision rationale**: Why specific approaches were chosen
- **Problem-solving patterns**: Debugging workflows and approaches
- **Performance baselines**: Historical metrics for comparison
- **Learning artifacts**: Evolution of understanding over time

They are archived rather than deleted to:

1. Preserve institutional knowledge
2. Allow future developers to understand the project's evolution
3. Provide examples of problem-solving approaches
4. Maintain audit trail for major architectural decisions

---

## Active Documentation Locations

For current, actively maintained documentation, see:

| Category | Location |
|----------|----------|
| **Constitution** | `.specify/memory/constitution.md` |
| **Spec Kit Artifacts** | `specs/001-cryptoutil/*.md` (core files) |
| **Implementation Status** | `specs/001-cryptoutil/implement/` |
| **Developer Guides** | `docs/README.md`, `docs/DEV-SETUP.md` |
| **Copilot Instructions** | `.github/instructions/*.md` |
| **Runbooks** | `docs/runbooks/` |
| **Test Documentation** | `docs/TEST-PERFORMANCE-ANALYSIS.md` |

---

## Restoration Process

If you need to restore archived content:

```bash
# Check archive history
git log --follow docs/archive/sessions/SESSION-2025-01-08-LESSONS-LEARNED.md

# Restore specific file
git checkout <commit-hash> -- docs/archive/sessions/SESSION-2025-01-08-LESSONS-LEARNED.md
```
