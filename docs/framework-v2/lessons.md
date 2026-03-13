# Lessons Learned - Framework v2

This file captures lessons from each phase, used as:
1. Memory for the entire plan.md / tasks.md execution
2. Input for knowledge propagation to ARCHITECTURE.md, agents, skills, instructions

---

## Inherited from Framework v1

Key lessons carried forward (see `docs/framework-v1/lessons.md` for full details):
1. **HTTP keep-alive hang**: ALL test HTTP clients that call real servers MUST use `DisableKeepAlives: true`
2. **Duration constant usage**: Magic constants of type `time.Duration` MUST NOT be multiplied by `time.Second`
3. **SetReady(true) requirement**: Must always be called after `MustStartAndWaitForDualPorts`
4. **Auth contracts**: AuthN/AuthZ are 100% owned by service-template. Auth contract tests (401/403 rejection) belong in `RunContractTests` in service-template, NOT in service-specific tests.
5. **Contract test integration**: Minimal friction - one line per service
6. **GitHub Workflows gap**: Framework changes should include corresponding CI workflow updates

---

## Phase 1: testdb.NewClosedSQLiteDB Helper

*(To be filled during Phase 1 execution)*

---

## Phase 2: jose-ja Cleanup

*(To be filled during Phase 2 execution)*

---

## Phase 3: sm-im Cleanup

*(To be filled during Phase 3 execution)*

---

## Phase 4: sm-kms Assessment and Safe Cleanup

*(To be filled during Phase 4 execution)*

---

## Phase 5: Knowledge Propagation

*(To be filled during Phase 5 execution)*
