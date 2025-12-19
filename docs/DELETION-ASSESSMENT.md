# Document Deletion Assessment

**Date**: December 19, 2025
**Purpose**: Assess whether archived clarification documents can be safely deleted

---

## Summary Recommendation

**KEEP ALL FILES** - Each serves distinct ongoing purpose:

| File | Size | Status | Recommendation |
|------|------|--------|----------------|
| SPECKIT-CONFLICTS-ANALYSIS.md | 588 lines | Round 1 quiz (answered) | KEEP - Reference |
| COPILOT-SUGGESTIONS.md | 451 lines | Unanswered optimization questions | KEEP - Future work |
| SPECKIT-COPILOT-SUGGESTIONS.md | 182 lines | Draft instruction file | KEEP - Phase 5 |
| PRODUCT-REFACTOR-TASKS.md | 745 lines | Architecture refactoring plan | KEEP - Active work |

---

## Detailed Assessment

### SPECKIT-CONFLICTS-ANALYSIS.md

**Purpose**: Round 1 clarification questions (26 total) answered 2025-12-19

**Status**:

- ✅ All 26 answers applied to constitution.md v3.0.0
- ✅ All 26 answers merged into clarify.md topical sections
- ✅ Commit f7f37262 documents complete application

**Deletion Candidate**: NO

**Rationale**:

- Historical reference for decision context (WHY questions were asked)
- Shows evolution from conflicts → resolution
- Small file (588 lines), minimal storage cost
- May be useful for post-mortem analysis or similar future conflict resolution

**Recommendation**: KEEP in docs/ as historical reference

---

### COPILOT-SUGGESTIONS.md

**Purpose**: 18 optimization questions for .github/instructions/*.instructions.md

**Status**:

- ❌ UNANSWERED - user has not provided choices yet
- ❌ NOT APPLIED - no changes made to instruction files based on this doc
- 📋 ACTIVE - represents actionable optimization work

**Deletion Candidate**: NO

**Rationale**:

- Contains 18 unanswered optimization questions:
  - S1-S2: Organization and structure (naming, file size)
  - C1-C3: Content clarity (terminology, CRITICAL prefix, code examples)
  - G1-G3: Missing topics (federation, cryptography, templates)
  - P1-P5: Performance considerations (token usage, file size, redundancy)
  - Q1-Q3: Quality assurance (consistency checks, anti-pattern tracking, regression prevention)
  - D1-D3: Developer experience (IDE hints, quick reference, workflow integration)
- User needs to answer these questions before applying to instruction files
- Represents future work, not completed work

**Next Steps**:

1. Create COPILOT-SUGGESTIONS-QUIZME.md with multiple choice format
2. User answers 18 questions
3. Apply answers to .github/instructions/ files
4. Then reassess COPILOT-SUGGESTIONS.md for deletion

**Recommendation**: KEEP until user answers and changes are applied

---

### SPECKIT-COPILOT-SUGGESTIONS.md

**Purpose**: Draft 06-01.speckit.instructions.md for LLM agents

**Status**:

- ❌ NOT ADDED - not yet created as .github/instructions/06-01.speckit.instructions.md
- 📋 PROPOSED - draft content ready for review
- 🎯 PHASE 5 TARGET - planned for Phase 5 maturity

**Deletion Candidate**: NO

**Rationale**:

- Contains valuable patterns for iterative Speckit development
- Proposed additions:
  - Mini-cycle feedback loop (every 3-5 tasks)
  - Implementation-driven constraints (living section pattern)
  - Constitution evolution triggers
  - Clarify.md update patterns
  - DETAILED.md timeline discipline
- Not yet in instruction files, would be lost if deleted
- Represents planned Phase 5 maturity enhancement

**Next Steps**:

1. Review draft content with user
2. Create .github/instructions/06-01.speckit.instructions.md
3. Integrate patterns into existing instruction files
4. Then reassess for deletion (likely keep as design document)

**Recommendation**: KEEP until instruction file created (Phase 5)

---

### PRODUCT-REFACTOR-TASKS.md

**Purpose**: Architecture refactoring plan for 4-product structure

**Status**:

- ❌ NOT STARTED - 0% implementation
- 📋 ACTIVE PLAN - represents major architectural work
- 🎯 FUTURE WORK - likely Phase 6 or later

**Deletion Candidate**: NO

**Rationale**:

- Documents comprehensive refactoring plan (745 lines, 20+ steps)
- Covers:
  - Product/service hierarchy (suite → product → service)
  - Private PKI infrastructure (suite CA → product CA → service CA)
  - Executable organization (cmd/dev/, cmd/product/, cmd/service/)
  - Configuration structure (per-product, per-service configs)
  - Deployment structure (Docker Compose product grouping)
- Not implemented, would lose roadmap if deleted
- May supersede some content in specs/002-cryptoutil/plan.md (investigate)

**Relationship to plan.md**:

- plan.md focuses on cryptographic features (phases 1-7)
- PRODUCT-REFACTOR-TASKS.md focuses on structural refactoring
- Both are complementary, not duplicative
- Refactoring likely occurs during/after Phase 5-6 implementation

**Recommendation**: KEEP as architecture roadmap, cross-reference with plan.md

---

## File Lifecycle Tracking

### Already Deleted (Commit f7f37262)

- ✅ specs/002-cryptoutil/CLARIFY-QA.md (obsolete Q&A format)
- ✅ specs/002-cryptoutil/CLARIFY-QUIZME2.md (merged into clarify.md)
- ✅ specs/002-cryptoutil/analyze-possibly-out-of-date.md (obsolete phase-based)
- ✅ specs/002-cryptoutil/clarify-decisions-2025-12-19.md (merged into clarify.md)
- ✅ specs/002-cryptoutil/clarify-possibly-out-of-date.md (replaced by clarify.md)
- ✅ specs/002-cryptoutil/other/IMPLEMENTATION-GUIDE.md (consolidated)
- ✅ specs/002-cryptoutil/other/MUTATION-TESTING-BASELINE.md (superseded)
- ✅ specs/002-cryptoutil/other/SESSION-SUMMARY.md (moved to DETAILED.md)
- ✅ specs/002-cryptoutil/other/SLOW-TEST-PACKAGES.md (addressed)
- ✅ specs/002-cryptoutil/other/SPEC-KIT-FILE-ANALYSIS.md (obsolete)
- ✅ specs/002-cryptoutil/plan-possibly-out-of-date.md (renamed to plan-probably-out-of-date.md)
- ✅ specs/002-cryptoutil/tasks-possibly-out-of-date.md (renamed to tasks-probably-out-of-date.md)

### Kept (Backup)

- 📁 specs/002-cryptoutil/clarify-old-backup.md (1973-line phase-based backup, delete after 1 week validation)

---

## Conclusion

**All assessed files serve distinct purposes and should be retained:**

1. **SPECKIT-CONFLICTS-ANALYSIS.md**: Historical reference for Round 1 clarifications
2. **COPILOT-SUGGESTIONS.md**: Unanswered optimization questions (18 remaining)
3. **SPECKIT-COPILOT-SUGGESTIONS.md**: Draft instruction file for Phase 5
4. **PRODUCT-REFACTOR-TASKS.md**: Architecture refactoring roadmap

**No deletions recommended at this time.**

**Future Deletion Candidates**:

- clarify-old-backup.md (after 1 week validation of new topical clarify.md)
- SPECKIT-CONFLICTS-ANALYSIS.md (after 6 months, if no longer referenced)
- COPILOT-SUGGESTIONS.md (after user answers and applies to instruction files)
