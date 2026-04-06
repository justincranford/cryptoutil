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

### What Worked

- Removing per-PS-ID postgres DB services was clean: delete service block, volumes, secrets
- Adding `include:` entries for shared-postgres and shared-telemetry was a straightforward addition
- lint-fitness compose-service-names and compose-db-naming linters correctly caught the transition
- `docker compose config` validated all 10 PS-ID composes successfully after migration
- Updating compose-db-naming from a forward check (must exist) to a regression guard (must NOT exist) = correct strategy

### What Didn't Work (Initially)

- Lint-fitness compose-service-names expected 5 services (4 app + 1 DB); needed to be updated to expect exactly 4
- Lint-fitness compose-db-naming needed a complete rewrite from "DB service name must match" to "DB service must NOT be present"
- Some PS-ID composes had `postgres-network:` stanza in the top-level `networks:` section that also needed removal

### Root Cause

- Per-PS-ID postgres services were fully embedded in each compose file; removing them required touching many sections (services, networks, volumes, secrets)
- Fitness linters were validating the OLD shape; they had to be updated before the new shape could pass

### Patterns to Propagate

1. **Fitness linters must be updated BEFORE or DURING structural changes** — not after
2. **Regression guard pattern**: When a structural element is permanently removed, the linter should flip from "must exist" to "must NOT exist" (catches accidental re-introduction)
3. **Config validation is sufficient pre-E2E smoke test** — `docker compose config --quiet` catches structural errors without needing running containers

---

## Phase 3: PRODUCT Recursive Includes — Approach C

### What Worked

- `include:` + `!override` pattern works cleanly for port substitution at PRODUCT level
- Docker Compose deduplicates shared infrastructure (shared-postgres, shared-telemetry) automatically when included multiple times via different PS-ID composes
- Product-level `secrets:` section correctly overrides all 7 shared secrets from both PS-ID includes (critical for SM which includes sm-kms AND sm-im, both defining `postgres-url.secret`)
- postgres-username/password/database secrets do NOT need product-level override; shared-postgres defines them once and they are inherited correctly
- Line count reduction: 2 product composes (SM 501 lines, identity 1033 lines) reduced to 80 and 155 lines respectively; overall reduction >80%
- `docker compose --profile dev config | Select-String "published"` is the correct way to verify port overrides when services have profiles

### What Didn't Work (Initially)

- Plans called for "single builder-sm service" at product level but include-based approach inherits per-PS-ID builders (builder-sm-kms, builder-sm-im); this is correct behavior — Docker caches the build, only one actual build occurs
- The Includes acceptance criterion originally said "shared-telemetry, shared-postgres, sm-kms, sm-im" but shared-postgres and shared-telemetry are inherited transitively via PS-ID includes

### Root Cause

- Acceptance criteria reflected a pre-research understanding; Phase 0 research confirmed the actual behavior and the acceptance criteria needed updating to match reality

### Patterns to Propagate

1. **Product composes need ONLY PS-ID includes** — shared infrastructure is inherited transitively
2. **Product-level secrets: section must override ALL conflicting secrets** from multiple PS-ID includes
3. **postgres-username/password/database use shared-postgres-scoped paths** — never need overriding at PRODUCT level
4. **Use `--profile dev` flag** when testing port overrides (`docker compose config` without profile hides profiled services)
5. **Acceptance criteria should be updated** when plans are proven incorrect by implementation

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
