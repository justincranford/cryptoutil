# Lessons — Framework v8: Deployment Parameterization

**Status**: Review pending — filtered for user approval.

Each item below is categorized as ADD (new to ENG-HANDBOOK/instructions), UPDATE (modify
existing content), or DELETE (already captured, no further action needed).

---

## Lessons to ADD to ENG-HANDBOOK.md / Instructions

1. **ADD** — Docker Compose `!override` YAML tag is MANDATORY for port array replacement in
   PRODUCT/SUITE compose files. Default array merge behavior APPENDS (does not replace).
   `!reset` clears arrays completely. Requires Docker Compose v2.24+.
   *Target: Section 12 (Deployment Architecture)*

2. **ADD** — Docker Compose secret file paths resolve relative to the INCLUDED file's directory,
   not the including file's directory. PRODUCT/SUITE can safely redefine secrets with own paths.
   *Target: Section 12.3.3 (Secrets Coordination Strategy)*

3. **ADD** — Docker Compose deduplicates shared infrastructure automatically when the same file
   is included via multiple paths. No special handling needed.
   *Target: Section 12 (Deployment Architecture)*

4. **ADD** — Regression guard pattern: When a structural element is permanently removed, flip
   fitness linter from "must exist" to "must NOT exist" to catch accidental re-introduction.
   *Target: Section 9.11 (Architecture Fitness Functions)*

5. **ADD** — Fitness linters must be updated BEFORE or DURING structural changes, not after.
   Otherwise the new structure cannot pass validation.
   *Target: Section 9.11 (Architecture Fitness Functions)*

6. **ADD** — Product composes need ONLY PS-ID includes — shared infrastructure (postgres,
   telemetry) is inherited transitively. Never include shared infrastructure directly at
   PRODUCT/SUITE level.
   *Target: Section 12 (Deployment Architecture)*

7. **ADD** — Container port is always 8080 in PS-ID composes. PRODUCT overrides use
   `["XXXX:8080"]`, SUITE overrides use `["XXXX:8080"]`. Old SUITE pattern using `:8000`
   is incorrect.
   *Target: Section 3.4.1 (Port Design Principles)*

8. **ADD** — Signature changes must update both production and test code atomically. Never
   leave test files with stale API calls.
   *Target: Section 14.1 (Coding Standards)*

9. **ADD** — Panic is appropriate for programmer errors (unknown enum/const values) vs errors
   for runtime failures (file not found, network timeout).
   *Target: Section 14.1 (Coding Standards)*

10. **ADD** — Always check for existing fitness linters before creating new ones. Search
    `lint_fitness/` directory. Superset linters eliminate need for subset linters.
    *Target: Section 9.11 (Architecture Fitness Functions)*

11. **ADD** — Carryover items should be verified against current codebase state during planning,
    not just copied forward blindly. Prior work may have already addressed the concern.
    *Target: Section 14.6 (Plan Lifecycle Management)*

12. **ADD** — Runtime E2E is mandatory for deployment refactors. `docker compose config` and
    lint validation alone cannot catch runtime startup failures (entrypoint binaries, init job
    collisions, runtime script assumptions).
    *Target: Section 10.4 (E2E Testing Strategy)*

13. **ADD** — PRODUCT/SUITE override layers should use per-PS-ID image tags (e.g.,
    `cryptoutil-sm-kms:dev`) when includes introduce multiple builders. Shared image tags
    across heterogeneous PS-ID binaries are unsafe in recursive include topologies.
    *Target: Section 12 (Deployment Architecture)*

14. **ADD** — Helper services in include-target compose files should be PS-ID-prefixed to
    avoid name collisions at PRODUCT/SUITE tiers. Use `{PS-ID}-init` not `init`.
    *Target: Section 12 (Deployment Architecture)*

15. **ADD** — shared-postgres init SQL must avoid fixed role ownership assumptions and remain
    runtime-user agnostic. Scripts should work with whatever username is provided via secrets.
    *Target: Section 12.3.3 (Secrets Coordination Strategy)*

16. **ADD** — Any Dockerfile using `/sbin/tini` ENTRYPOINT must install/copy `tini` in
    the runtime stage.
    *Target: Section 12 (Deployment Architecture)*

---

## Lessons to UPDATE in ENG-HANDBOOK.md / Instructions

1. **UPDATE** — Docker Compose minimum version: currently documented as v2+ in
    `02-02.versions.instructions.md` and `docs/DEV-SETUP.md`. Should be v2.24+ (required
    for `!override` YAML tag used by recursive include architecture).
    *Target: Section 2.2 (versions), DEV-SETUP.md*

---

## Lessons to DELETE (already captured or no longer relevant)

1. **DELETE** — "Sed port shifts with sequential `-e` flags caused double-replacement" — this
    is a one-time operational lesson, not a recurring pattern. Already mitigated.

2. **DELETE** — "Python is effective for complex structured insertions" — tooling preference,
    contradicts "Go first" scripting policy. Not worth propagating.

3. **DELETE** — "SUITE compose uses `:8000` container port" — this was the OLD incorrect
    pattern. Already fixed. Lesson 7 above captures the correct container port rule.

4. **DELETE** — "Sed-based bulk rename" details — one-time operational, not recurring.

5. **DELETE** — "Phase 7 commit ran through full 40+ step pre-commit hook" — observation
    about CI speed, not actionable.

6. **DELETE** — "Always update task-level acceptance criteria checkboxes" — process lesson
    for agent execution, not for ENG-HANDBOOK.
