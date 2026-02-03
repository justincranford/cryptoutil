# QuizMe - V8 Executive Decisions

**Purpose**: Clarify approach for completing KMS barrier migration

---

## Q1: Barrier Migration Approach

**Context**: V7 created orm_barrier_adapter.go but did NOT integrate it. The adapter allows KMS's OrmRepository to implement template barrier interfaces.

**Question**: How should we approach the barrier migration?

A) Use the existing orm_barrier_adapter.go to bridge KMS → template barrier
B) Refactor KMS to directly use template barrier Repository interface (remove OrmRepository layer)
C) Keep KMS using shared/barrier but ensure feature parity (abandon template unification)
D) Create new unified barrier in internal/shared that ALL services use (not template-specific)
E) There must be only one barrier implementation, and it must be in service-template. All services must use it directly consistently.

**Choice**: E

**Implications**:
- A: Fastest, uses existing work, but adds adapter layer complexity
- B: Cleaner architecture, but requires more extensive refactoring of KMS
- C: Abandons V7 goal of unified architecture
- D: Would require changes to cipher-im and jose-ja (regression risk)

---

## Q2: shared/barrier Deprecation Timeline

**Context**: If KMS migrates to template barrier, what happens to shared/barrier?

**Question**: What should happen to internal/shared/barrier/ after KMS migration?

A) Delete immediately after KMS migration complete
B) Keep as deprecated with warnings, delete in V9
C) Keep indefinitely (other future services might need it)
D) Merge any unique features into template barrier, then delete
E) There must be only one barrier implementation, and it must be in service-template. All services must use it directly consistently. The one in service-template must have all of the functionality, and the old one in shared/barrier must be deleted as soon as KMS migrates.

**Choice**: E

**Implications**:
- A: Clean codebase, but risky if migration has issues
- B: Safer, allows rollback if problems discovered
- C: Code duplication continues
- D: Most thorough, but V7 Task 5.2 claims features already merged

---

## Q3: Testing Priority for V8

**Context**: V7 skipped all Phase 6 testing tasks. KMS currently passes unit tests.

**Question**: What testing scope is required before considering V8 complete?

A) Unit tests only (fastest, least coverage)
B) Unit + integration tests (moderate)
C) Unit + integration + E2E (comprehensive)
D) Unit + integration + E2E + mutation testing (V7 full scope)
E) D; do Unit + integration + E2E tasks as part of every phase; do mutations as a separate phase at the end BUT DO NOT DEFER OR DE-PRIORITIZE IT BECAUSE THIS THE ORDERING IS STRATEGICALLY IMPORTANT.

**Choice**: E

**Implications**:
- A: Risk of integration/E2E issues discovered later
- B: Good balance of speed and coverage
- C: High confidence, but E2E can be flaky
- D: Highest confidence but longest timeline

---

## Q4: Documentation Update Timing

**Context**: V7 deferred all documentation updates to Phase 7.

**Question**: When should documentation be updated?

A) After all code changes complete (Phase 3)
B) Incrementally as each task completes
C) Only update instructions files that are actually wrong
D) Skip documentation updates entirely for V8 (do in V9)
E) Incrementally update instructions files that are actually wrong

**Choice**: E

**Implications**:
- A: Risk of forgetting context
- B: More overhead but always accurate
- C: Minimal effort, focused on actual gaps
- D: Documentation drift continues

---

## Instructions

1. Review each question
2. Select A-D or fill in option E
3. Fill in Choice field
4. After answering, agent will merge decisions into plan.md/tasks.md
5. quizme-v1.md will be deleted after merge
