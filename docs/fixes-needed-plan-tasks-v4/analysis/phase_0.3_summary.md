# Phase 0.3: Global Mutation Target Fix - COMPLETE ✅

**Status**: ✅ COMPLETE
**Duration**: 2 hours
**Commits**: 234372f5 (global fix), 7e0e14c5 (tasks.md update)

## Objective

Fix all instances of ">=85%" minimum mutation/coverage targets globally, replacing with ">=95% minimum, 98% ideal" per user's critical correction.

## User Feedback

"one mistake i see if you reverted to minimum migrations >=85%. i changed the mutations floor to >=95% and ideal 98%; look in the entire project and fix those mutation targets globally"

## Execution Summary

### Task 0.3.1: Global Search and Replace Mutation Targets

**Acceptance Criteria**: 9/9 complete ✅

1. ✅ Search for "85" in docs/fixes-needed-plan-tasks-v4/*.md (20+ matches found)
2. ✅ Update plan.md (5 replacements: success criteria, phase objectives, decision rationale)
3. ✅ Update completed.md (7 replacements: header, tasks 1.5.1.4/1.5.2.4/1.5.3/2.5, Task 0.2)
4. ✅ Search .github/instructions/*.md (verified: already correct at >=95%)
5. ✅ Update ARCHITECTURE.md (3 replacements: principles, requirements, pre-merge)
6. ✅ Update remaining docs (2 replacements: coverage-analysis, gremlins/MUTATIONS-TASKS)
7. ✅ Update agent files (3 replacements: speckit.agent.md, plan-tasks-quizme.agent.md)
8. ✅ Verification: grep confirms 0 remaining "≥85%" targets in mutation/coverage context
9. ✅ Commit with detailed message documenting global correction

## Results

**Total Changes**: 20 replacements across 7 files

**Pattern Applied**:
- Replace: TARGET language ("≥85%", "85% target", "85% minimum")
- With: "≥95% minimum" or "≥95% minimum, 98% ideal"
- Preserve: Historical ACHIEVEMENT numbers (85.3%, 87.9%, etc.) as factual data
- Add: Clarifications for practical limits showing gap between reality and target

**Files Modified**:
1. docs/fixes-needed-plan-tasks-v4/plan.md (5 replacements)
   - Phase 1.5 success criteria (session manager, TLS generator)
   - Phase 2 coverage status
   - Task 2.5 verification criteria
   - Practical limit clarifications
   - Decision 1 alternatives rationale
   
2. docs/fixes-needed-plan-tasks-v4/completed.md (7 replacements)
   - Header note explaining practical limits vs new targets
   - Task 1.5.1.4, 1.5.2.4, 1.5.3 targets
   - Task 2.5 title, status, coverage table
   - Task 0.2 mutation efficacy standards
   
3. docs/arch/ARCHITECTURE.md (3 replacements)
   - Core design principles (Reliability section)
   - Mutation coverage requirements
   - Pre-merge requirements checklist
   
4. docs/coverage-analysis-2026-01-27.md (1 replacement)
   - Mutation efficacy standards
   
5. docs/gremlins/MUTATIONS-TASKS.md (1 replacement)
   - Business logic targets
   
6. .github/agents/speckit.agent.md (1 replacement)
   - Quality gate standards
   
7. .github/agents/plan-tasks-quizme.agent.md (2 replacements)
   - Mutation testing references in checklists

**Copilot Instructions Verification**:
- .github/instructions/03-02.testing.instructions.md: Already correct at >=95% (no changes needed)
- .github/instructions/06-01.evidence-based.instructions.md: Already correct at >=95% (no changes needed)

## Key Learnings

1. **Pattern Recognition Critical**: Must distinguish TARGET language ("≥85%", "85% target") from ACHIEVEMENT data ("85.3% achieved")
2. **Practical Limits vs Targets**: Achieved ranges (85-90%) ≠ target standards (>=95%, ideal 98%)
3. **Historical Preservation**: Achievement numbers are factual data points, not targets to update
4. **Comprehensive Search Required**: Mutation targets appear in docs, architecture, agents - not just task files
5. **Clarification Value**: Adding "(target >=95%, ideal 98%)" to practical limits shows gap between reality and expectations

## Impact on Quality Standards

**Before**: Ambiguous 85% references mixed with 95%/98% standards
**After**: Consistent >=95% minimum (98% ideal) across entire project

**Coverage Impact**:
- Production code: >=95% minimum coverage
- Infrastructure/utility: >=98% coverage
- Mutation production: >=95% minimum efficacy (98% ideal)
- Mutation infrastructure: >=98% efficacy

## Next Steps

1. ✅ Phase 0.3 COMPLETE - Move to completed.md archive
2. Proceed to analyze review-tasks-v5.md findings (step 8 of workflow)
3. Continue with remaining phases per PRIMARY DIRECTIVE
