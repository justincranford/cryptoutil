# Lessons — Framework v8: Deployment Parameterization

*This file is maintained by the implementation-execution agent. Each section is filled in after
the corresponding phase completes its quality gates. Lessons record what worked, what didn't,
root causes, and patterns to propagate to permanent artifacts.*

---

## Phase 0: Technical Research

### What Worked

- Minimal compose file approach isolated each behavior cleanly
- Docker Compose v2.24+ include and service override work as expected
- Deduplication of shared includes works correctly (no duplicate service errors)
- Profile inheritance through includes works correctly
- Secret file paths resolve relative to the INCLUDED file's directory

### What Didn't Work (Initially)

- **Plain service redefinition does NOT replace `ports:` arrays** — Docker Compose MERGES (appends) arrays by default
- This was the critical discovery: the plan's Approach C requires `!override` YAML tag

### Root Cause

- Docker Compose follows YAML merge rules for arrays: concatenation, not replacement
- The `!override` YAML tag (Docker Compose v2.24+) explicitly REPLACES the inherited value

### Patterns to Propagate

1. **`!override` tag is MANDATORY** for all port overrides in PRODUCT and SUITE compose files
2. **`!reset` clears arrays completely** (useful for removing inherited ports entirely, e.g., postgres)
3. **Secret paths resolve from included file's directory** — PRODUCT/SUITE can safely redefine secrets with their own paths
4. **Include deduplication works** — no special handling needed for shared infrastructure included via multiple paths
5. **Docker Compose v2.24+ is minimum version** — must document this requirement

---

## Phase 1: Naming Standardization + Missing Services

### What Worked

- Sed-based bulk rename for `app-postgres-N` to `app-postgresql-N` was fast and reliable
- Jose and skeleton PRODUCT composes already used `postgresql` — only sm, pki, identity needed fixing
- lint-deployments (54/54 validators passed) confirmed no regressions

### What Didn't Work (Initially)

- Sed port shifts with sequential `-e` flags caused double-replacement (18401->18402->18403)
- Had to restore from git and redo with reverse-order sed (higher numbers first)

### Root Cause

- Sed processes `-e` expressions sequentially on the same line buffer, so 18401->18402 then 18402->18403
- Solution: process in descending order (18402->18403 THEN 18401->18402)

### Patterns to Propagate

1. **Always shift ports in descending order** when using sed to avoid double-replacement
2. **SUITE compose uses `:8000` container port**, not `:8080` — different from PS-ID and PRODUCT composes
3. **Python is effective** for complex structured insertions (sqlite-2 block generation)

---

## Phase 2: Standalone Profile + Shared Infrastructure at All Tiers

*(To be filled during Phase 2 execution)*

---

## Phase 3: PRODUCT Recursive Includes — Approach C

*(To be filled during Phase 3 execution)*

---

## Phase 4: SUITE Recursive Includes — Approach C

*(To be filled during Phase 4 execution)*

---

## Phase 5: Validator + Linter Updates

*(To be filled during Phase 5 execution)*

---

## Phase 6: Fitness Linter — `usage_health_path_completeness`

*(To be filled during Phase 6 execution)*

---

## Phase 7: Documentation + ENG-HANDBOOK.md Updates

*(To be filled during Phase 7 execution)*

---

## Phase 8: E2E Validation

*(To be filled during Phase 8 execution)*

---

## Phase 9: Knowledge Propagation

*(To be filled during Phase 9 execution)*
