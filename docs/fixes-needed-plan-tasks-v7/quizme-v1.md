# Quiz - Unified Service-Template Migration (V7)

**Purpose**: Clarify unknowns, risks, and inefficiencies before starting V7
**Created**: 2026-02-02

---

## Question 1: KMS Data Migration Strategy

**Context**: KMS currently uses SQLRepository with raw database/sql. Migration to GORM requires converting all queries and potentially migrating existing data.

**Question**: What is your preferred migration strategy for KMS data?

- A. **Incremental (table-by-table)**: Migrate one table at a time, validate, then proceed to next. Lower risk but longer timeline.
- B. **Big-bang (all at once)**: Migrate all tables simultaneously. Faster but higher risk.
- C. **Shadow mode (dual-write)**: Write to both old and new systems during transition, then cut over. Safest but most complex.
- D. **Fresh start**: Don't migrate existing data - start fresh with new schema (acceptable for dev/test environments).
- E. Other (specify): _____

**Choice**: _____

---

## Question 2: shared/barrier vs Template Barrier

**Context**: KMS uses `internal/shared/barrier` while service-template has its own barrier implementation. They may have different capabilities.

**Question**: What is your preference for barrier unification?

- A. **Replace completely**: Template barrier becomes the ONLY barrier implementation. shared/barrier is deleted.
- B. **Template wraps shared**: Template barrier delegates to shared/barrier internally. shared/barrier remains as underlying implementation.
- C. **Merge into template**: Move shared/barrier code INTO template barrier. One unified implementation.
- D. **Keep both**: Both implementations coexist. Services choose which to use. (Note: This contradicts V7 goals)
- E. Other (specify): _____

**Choice**: _____

---

## Question 3: KMS API Versioning

**Context**: Migrating KMS to OpenAPI strict server and JWT auth may change API behavior or contracts.

**Question**: How should API versioning be handled during migration?

- A. **Breaking change (v2)**: Create new `/service/api/v2/**` endpoints. Old v1 deprecated/removed.
- B. **Backward compatible (v1 extended)**: Keep v1 endpoints, add new fields/capabilities without breaking existing clients.
- C. **Dual support**: Both v1 and v2 during transition period. v1 removed after clients migrate.
- D. **Internal only**: KMS has no external clients - just make changes directly.
- E. Other (specify): _____

**Choice**: _____

---

## Question 4: Timeline Priority

**Context**: V7 has 8 phases with ~36h estimated LOE. Some phases can be parallelized, others have hard dependencies.

**Question**: What is your priority for V7 execution?

- A. **Correctness first**: Take whatever time needed to get architecture right. No shortcuts.
- B. **Speed first**: Get KMS working on service-template ASAP, even if some technical debt remains.
- C. **Risk-first**: Address highest-risk items first (barrier migration, data migration), then lower-risk items.
- D. **Dependency-ordered**: Strictly follow phase order (0→1→2→3→4→5→6→7).
- E. Other (specify): _____

**Choice**: _____

---

## Question 5: cipher-im and jose-ja Validation

**Context**: cipher-im and jose-ja already use service-template correctly. Phase 1 removes V6 optional modes which should NOT affect them.

**Question**: How thorough should validation be for cipher-im and jose-ja?

- A. **Minimal**: Run existing tests only. If they pass, proceed.
- B. **Standard**: Run tests + manual smoke test of key functionality.
- C. **Thorough**: Full regression suite including E2E, coverage checks, mutation testing.
- D. **Skip validation**: They're already working - focus all effort on KMS.
- E. Other (specify): _____

**Choice**: _____

---

## Question 6: Documentation Timing

**Context**: Phase 7 is documentation. Some prefer docs-as-you-go, others prefer docs-at-end.

**Question**: When should documentation be updated?

- A. **End only**: Update all docs in Phase 7 after code is complete.
- B. **Per-phase**: Update relevant docs after each phase completes.
- C. **Continuously**: Update docs immediately as code changes (adds overhead but ensures accuracy).
- D. **Minimal**: Update only critical docs (server-builder.instructions.md). Skip others.
- E. Other (specify): _____

**Choice**: _____

---

## Instructions

1. Review each question and context
2. Select your preferred option (A, B, C, D, or E with specification)
3. Fill in the **Choice** field
4. Run `/plan-tasks-quizme docs\fixes-needed-plan-tasks-v7 update` to merge answers

---

## Notes for LLM Agent

After user answers:
1. Update plan.md Technical Decisions section with user's choices
2. Adjust task priorities/dependencies based on choices
3. Delete this quizme-v1.md file
4. Commit changes with message: "docs(v7): merge quizme answers into plan"
