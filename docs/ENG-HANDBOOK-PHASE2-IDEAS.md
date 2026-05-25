## Yes (IMPLEMENT NOW) — All 5 Implemented

1. **✅ IMPLEMENTED (2026-05-23 pilot + 2026-05-25 documented)**: Pilot migration complete.
   `docs/ENG-HANDBOOK.md` lines 99–155 contain three `@section-to-appendix` blocks for the
   terminology chunks (`rfc-2119-keywords`, `emphasis-keywords`, `abbreviations`) and matching
   `@appendix-why` + `@appendix-propagate` blocks in the "Terminology Appendix Propagation (Pilot)"
   subsection. Appendix ID: `terminology-instruction-body`. Target: `01-01.terminology.instructions.md`.

2. **✅ IMPLEMENTED (already present in linter, documented 2026-05-25)**: The distinct
   section-to-appendix marker family is fully implemented in `validate_chunks.go` and
   `validate_coverage.go`. Four bidirectional validation rules are enforced: (a) every
   `@section-to-appendix` chunk must have a matching `@appendix-propagate` block, (b) reverse,
   (c) every `@appendix-propagate` must have an adjacent `@appendix-why`, (d) a chunk with
   `@section-to-appendix` cannot also have a direct `@propagate`. Unstable chunk IDs are rejected.

3. **✅ IMPLEMENTED (2026-05-25)**: All 19 instruction files under `.github/instructions/` now
   have `<!-- @local-glue:start/end -->` wrapping the title heading and
   `<!-- @handbook-derived-body:start/end -->` wrapping all body content. Pilot file
   (`01-01.terminology.instructions.md`) had the markers already; remaining 18 were applied in
   this session. Confirmed with `go run ./cmd/cicd-lint lint-docs` — all 11 sub-linters pass.

4. **✅ ALREADY ENFORCED (documented 2026-05-25)**: Agent pairs already use the shared-body plus
   per-target frontmatter model. `lint-agent-drift` validates body identity for all four pairs
   (`beast-mode`, `fix-workflows`, `implementation-execution`, `implementation-planning`).
   Copilot files require `tools:` (whitelist); Claude files omit it. Documented in NEW.md §7.3.

5. **✅ ALREADY ENFORCED (documented 2026-05-25)**: Skill pairs already use the same shared-body
   model. `lint-skill-command-drift` validates body identity + `## Key Rules` presence for all 13
   pairs. Copilot skills may have `disable-model-invocation: true`; Claude files must not.
   Documented in NEW.md §7.4.

## Yes

1. Add a linter rule that rejects orphan appendix blocks that do not propagate to any downstream target.
2. Add a linter rule that rejects semantic contribution blocks that do not feed any appendix block.
3. Add a linter rule that rejects downstream targets that are populated directly from semantic sections instead of from appendix blocks.
4. Prefer stable semantic chunk ids based on meaning, not section numbers, so handbook renumbering does not force downstream id churn.
5. Consider requiring every appendix block to declare a short `why-this-exists` note outside the propagated text so reviewers can tell whether the block is semantic, structural, or compatibility glue.
6. Introduce an appendix review order in the linter output so failures surface in the same order humans read the appendixes.
7. Move duplicated handbook prose out of skills and agents into appendix-backed body fragments so the same rule text is reused consistently across more downstream surfaces.
8. Treat extremely handbook-coupled skills as early migration candidates: `propagation-check`, `sync-copilot-claude`, `copilot-customization`, `test-table-driven`, and `openapi-codegen`.

## Unsure

1. Make appendix block ids globally unique and require each downstream file section to be owned by exactly one appendix block.
2. Keep `CLAUDE.md` as a partial consumer of appendix content instead of forcing the entire file into a single appendix block.
3. Consider a generated `docs/required-propagations-v2.yaml` that records both edges of the graph: `section -> appendix` and `appendix -> downstream`.
4. Add a `phase maturity` tag to appendix blocks so some can remain whole-file mirrors in transition while others evolve into semantically assembled bodies.

## No

1. Add an appendix ownership index that lists every downstream file, its appendix owner, and the semantic section contributors that feed it.
2. Create a compact `semantic source registry` table near the top of the handbook listing each semantic bucket, the appendix blocks it feeds, and the target files reached through those appendix blocks.
3. Group instruction appendix subsections by domain family such as terminology, architecture, security, testing, and operations to make maintenance audits faster.
