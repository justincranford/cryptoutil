# Lessons — Framework v8: Deployment Parameterization

**Status**: Filtered for user review. Items verified against current ENG-HANDBOOK.md and instructions.

---

## ADD to ENG-HANDBOOK.md (7 items — not yet captured)

1. **ADD** — Docker Compose secret file paths resolve relative to the INCLUDED file's
   directory, not the including file's directory. PRODUCT/SUITE can safely redefine secrets
   with own paths.
   *Target: Section 12.3.3 (Secrets Coordination Strategy)*

2. **ADD** — Docker Compose deduplicates shared infrastructure automatically when the same
   file is included via multiple paths. No special handling needed.
   *Target: Section 12 (Deployment Architecture)*

3. **ADD** — Regression guard pattern: When a structural element is permanently removed, flip
   fitness linter from "must exist" to "must NOT exist" to catch accidental re-introduction.
   *Target: Section 9.11 (Architecture Fitness Functions)*

4. **ADD** — Fitness linters must be updated BEFORE or DURING structural changes, not after.
   Otherwise the new structure cannot pass validation.
   *Target: Section 9.11 (Architecture Fitness Functions)*

5. **ADD** — Always check for existing fitness linters before creating new ones. Search
   `lint_fitness/` directory. Superset linters eliminate need for subset linters.
   *Target: Section 9.11 (Architecture Fitness Functions)*

6. **ADD** — Runtime E2E is mandatory for deployment refactors. `docker compose config` and
   lint validation alone cannot catch runtime startup failures (entrypoint binaries, init job
   collisions, runtime script assumptions).
   *Target: Section 10.4 (E2E Testing Strategy)*

7. **ADD** — PRODUCT/SUITE override layers should use per-PS-ID image tags (e.g.,
   `cryptoutil-sm-kms:dev`) when includes introduce multiple builders. Shared image tags
   across heterogeneous PS-ID binaries are unsafe in recursive include topologies.
   *Target: Section 12 (Deployment Architecture)*

---

## UPDATE in ENG-HANDBOOK.md (1 item)

1. **UPDATE** — Docker Compose minimum version: currently documented as v2+ in
   `02-02.versions.instructions.md` and `docs/DEV-SETUP.md`. Should be v2.24+ (required
   for `!override` YAML tag used by recursive include architecture).
   *Target: Section 2.2 (versions), DEV-SETUP.md*

---

## Already captured — no action needed (9 items removed from original list)

The following lessons from the original list are already in ENG-HANDBOOK.md or instructions:

- `!override` YAML tag → ENG-HANDBOOK Section 12.3.4
- Product composes only PS-ID includes (transitive inheritance) → ENG-HANDBOOK Section 12.3.4
- Container port always 8080 → 02-01.architecture.instructions.md Port Design Principles
- Helper services PS-ID-prefixed → 02-01.architecture.instructions.md Recursive Include Guardrails
- shared-postgres init SQL user-agnostic → 02-01.architecture.instructions.md Recursive Include Guardrails
- Dockerfile /sbin/tini → 02-01.architecture.instructions.md Recursive Include Guardrails
- Signature changes update both prod+test → generic coding discipline, not handbook-worthy
- Panic vs error convention → standard Go convention, well-documented elsewhere
- Carryover verification during planning → agent process lesson, not engineering handbook
