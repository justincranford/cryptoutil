# Archive Review and Consolidation Tasks

**Created**: December 21, 2025
**Purpose**: Review all archived documentation, consolidate into copilot instructions and speckit documents, then delete redundant files

---

## Task Groups

### Group 1: Archive File Review (docs/archive)

Review and process all files in docs/archive subdirectories:

#### 1.1 Session Analyses (docs/archive/session-analyses-2025-01/)

- [ ] **Task 1.1.1**: Review GAP-ANALYSIS-2025-01-10.md
  - Break down into logical parts
  - Map to copilot instructions sections
  - Map to constitution.md (if architectural/requirement content)
  - Map to spec.md (if feature specification content)
  - Delete file after extraction

- [ ] **Task 1.1.2**: Review TEST-PERFORMANCE-ANALYSIS.md
  - Extract test performance patterns to 01-04.testing.instructions.md
  - Extract timing targets to constitution.md
  - Delete file after extraction

- [ ] **Task 1.1.3**: Review TIMEOUT-FIXES-ANALYSIS.md
  - Extract timeout patterns to 01-04.testing.instructions.md
  - Extract network operation patterns to 01-07.security.instructions.md
  - Delete file after extraction

#### 1.2 Session Summaries (docs/archive/sessions/)

- [ ] **Task 1.2.1**: Review SESSION-2025-01-08-LESSONS-LEARNED.md
  - Extract lessons to 07-01.anti-patterns.instructions.md
  - Extract to DETAILED.md Section 2 timeline (if not already there)
  - Delete file

- [ ] **Task 1.2.2**: Review SESSION-2025-01-08-RACE-FIXES.md
  - Extract race condition patterns to 01-04.testing.instructions.md
  - Extract to 07-01.anti-patterns.instructions.md
  - Delete file

- [ ] **Task 1.2.3**: Review SESSION-2025-12-08-PHASE4.md
  - Extract to DETAILED.md Section 2 timeline
  - Delete file

- [ ] **Task 1.2.4**: Review SESSION-2025-12-08-RESTART3.md
  - Extract to DETAILED.md Section 2 timeline
  - Delete file

- [ ] **Task 1.2.5**: Review SESSION-2025-12-09-CI-FIXES.md
  - Extract CI/CD patterns to 02-01.github.instructions.md
  - Extract to DETAILED.md Section 2 timeline
  - Delete file

- [ ] **Task 1.2.6**: Review SESSION-2025-12-09-TASK-3-FINAL-SUMMARY.md
  - Extract to DETAILED.md Section 2 timeline
  - Delete file

- [ ] **Task 1.2.7**: Review SESSION-2025-12-09-TASK-3-IDENTITY-COVERAGE.md
  - Extract coverage patterns to 01-04.testing.instructions.md
  - Extract to DETAILED.md Section 2 timeline
  - Delete file

- [ ] **Task 1.2.8**: Review SESSION-2025-12-09-WORKFLOW-FIXES.md
  - Extract workflow patterns to 02-01.github.instructions.md
  - Extract to DETAILED.md Section 2 timeline
  - Delete file

- [ ] **Task 1.2.9**: Review SESSION-2025-12-10-TASK-7-KMS-HANDLER-ANALYSIS.md
  - Extract KMS architecture to spec.md (if applicable)
  - Extract to DETAILED.md Section 2 timeline
  - Delete file

- [ ] **Task 1.2.10**: Review SESSION-COVERAGE-IMPROVEMENTS.md
  - Extract coverage improvement patterns to 01-04.testing.instructions.md
  - Extract to DETAILED.md Section 2 timeline
  - Delete file

- [ ] **Task 1.2.11**: Review SESSION-MFA-COVERAGE-PROGRESS.md
  - Extract MFA patterns to spec.md (if applicable)
  - Extract to DETAILED.md Section 2 timeline
  - Delete file

#### 1.3 Speckit Archives (docs/archive/speckit/)

- [ ] **Task 1.3.1**: Review SPECKIT-ITERATION-1-REVIEW.md
  - Extract speckit lessons to 06-01.speckit.instructions.md
  - Delete file

- [ ] **Task 1.3.2**: Review SPECKIT-PROGRESS.md
  - Extract to DETAILED.md Section 2 timeline
  - Delete file

#### 1.4 Workflow Analysis (docs/archive/workflow-analysis/)

- [ ] **Task 1.4.1**: Review workflow-analysis.md
  - Extract workflow patterns to 02-01.github.instructions.md
  - Delete file

- [ ] **Task 1.4.2**: Review WORKFLOW-clientauth-TEST-TIMES.md
  - Extract timing patterns to 01-04.testing.instructions.md
  - Delete file

- [ ] **Task 1.4.3**: Review WORKFLOW-jose-server-TEST-TIMES.md
  - Extract timing patterns to 01-04.testing.instructions.md
  - Delete file

- [ ] **Task 1.4.4**: Review WORKFLOW-jose-TEST-TIMES.md
  - Extract timing patterns to 01-04.testing.instructions.md
  - Delete file

- [ ] **Task 1.4.5**: Review WORKFLOW-OVERHEAD-ANALYSIS.md
  - Extract overhead patterns to 02-01.github.instructions.md
  - Delete file

- [ ] **Task 1.4.6**: Review WORKFLOW-sqlrepository-TEST-TIMES.md
  - Extract timing patterns to 01-04.testing.instructions.md
  - Delete file

#### 1.5 Top-Level Archive Files (docs/archive/)

- [ ] **Task 1.5.1**: Review CGO-BAN-ENFORCEMENT.md
  - Ensure coverage in 01-05.golang.instructions.md (CGO Ban section)
  - Delete file if fully covered

- [ ] **Task 1.5.2**: Review MUTATION-TESTING-FIXES.md
  - Extract mutation patterns to 01-04.testing.instructions.md
  - Delete file

- [ ] **Task 1.5.3**: Review README.md
  - Evaluate if still needed or content extracted
  - Delete if redundant

### Group 2: Speckit Passthrough Review (docs/speckit)

Review grooming sessions and extract relevant content:

- [ ] **Task 2.1**: Review docs/speckit/passthru02/grooming/GROOMING-SESSION-02.md
  - Extract grooming insights to 06-01.speckit.instructions.md
  - Extract to clarify.md (if Q&A content)
  - Delete file

- [ ] **Task 2.2**: Review docs/speckit/passthru03/grooming/GROOMING-SESSION-03.md
  - Extract grooming insights to 06-01.speckit.instructions.md
  - Extract to clarify.md (if Q&A content)
  - Delete file

### Group 3: Commit Archive Cleanup

- [ ] **Task 3.1**: Commit all archive file deletions
  - Use conventional commit: `chore(docs): consolidate archived documentation into copilot instructions`
  - Run with pre-commit hooks

---

## Task Groups for Instruction Optimization

### Group 4: Optimize Copilot Instructions

- [ ] **Task 4.1**: Analyze .github/copilot-instructions.md for redundancy
  - Identify duplicate content across instruction files
  - Remove redundant sections (keep one authoritative source)
  - Optimize table formats and examples

- [ ] **Task 4.2**: Review .github/instructions/*.instructions.md files
  - Identify overlapping content between files
  - Consolidate related patterns into single files
  - Remove verbose examples where terse guidance suffices
  - Maintain quality and coverage

- [ ] **Task 4.3**: Commit instruction optimizations
  - Use conventional commit: `docs(instructions): optimize copilot instructions for token efficiency`
  - Run with pre-commit hooks

---

## Task Groups for Speckit Document Review

### Group 5: Speckit Document Alignment

- [ ] **Task 5.1**: Review constitution.md for completeness
  - Verify all architectural constraints documented
  - Verify quality gates (coverage, mutation, timing)
  - Verify alignment with copilot instructions
  - Add missing sections

- [ ] **Task 5.2**: Review spec.md for completeness
  - Verify all product/service features documented
  - Verify all API contracts defined
  - Verify security requirements explicit
  - Verify alignment with copilot instructions

- [ ] **Task 5.3**: Review clarify.md for completeness
  - Verify topical organization
  - Verify all ambiguities resolved
  - Verify cross-references to constitution/spec

- [ ] **Task 5.4**: Commit speckit document updates
  - Use conventional commit: `docs(speckit): align constitution, spec, clarify with copilot instructions`
  - Run with pre-commit hooks

---

## Task Groups for Clarify Regeneration

### Group 6: Clarify Document Regeneration

- [ ] **Task 6.1**: Backup existing clarify documents
  - Rename specs/002-cryptoutil/clarify.md to clarify.md.old
  - Rename specs/002-cryptoutil/CLARIFY-QUIZME.md to CLARIFY-QUIZME.md.old

- [ ] **Task 6.2**: Analyze constitution.md and spec.md
  - Document current state and gaps

- [ ] **Task 6.3**: Run /speckit.clarify command
  - Generate new clarify.md from constitution + spec

- [ ] **Task 6.4**: Analyze new clarify.md
  - Compare to old versions
  - Identify omissions

- [ ] **Task 6.5**: Create clarify-OMISSIONS.md
  - Document what was not reproduced by /speckit.clarify
  - Compare against clarify.md.old and CLARIFY-QUIZME.md.old

- [ ] **Task 6.6**: Integrate omissions into clarify.md
  - Manually add missing content from OMISSIONS document

- [ ] **Task 6.7**: Delete backup files
  - Remove clarify.md.old
  - Remove CLARIFY-QUIZME.md.old

- [ ] **Task 6.8**: Create new CLARIFY-QUIZME.md
  - Format as multiple choice questions (A-D + E write-in)
  - Include only genuine unknowns requiring user input
  - Focus on problems, ambiguities, conflicts, risks

- [ ] **Task 6.9**: Prompt user for review
  - Ask user to review clarify-OMISSIONS.md
  - Ask user to review CLARIFY-QUIZME.md

---

## Execution Order

1. Execute Group 1 (Archive Review) - Tasks 1.1.1 through 1.5.3
2. Execute Group 2 (Speckit Passthrough) - Tasks 2.1 through 2.2
3. Execute Group 3 (Commit Archive) - Task 3.1
4. Execute Group 4 (Optimize Instructions) - Tasks 4.1 through 4.3
5. Execute Group 5 (Speckit Alignment) - Tasks 5.1 through 5.4
6. Execute Group 6 (Clarify Regeneration) - Tasks 6.1 through 6.9

---

## Success Criteria

- [ ] All archive files reviewed and content extracted
- [ ] All archive files deleted after extraction
- [ ] Copilot instructions optimized for token efficiency
- [ ] Constitution, spec, clarify aligned and complete
- [ ] New clarify.md generated with /speckit.clarify
- [ ] Omissions documented and integrated
- [ ] New CLARIFY-QUIZME.md created with proper format
- [ ] All changes committed with pre-commit hooks
- [ ] User prompted for final review
