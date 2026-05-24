## Doing Now

1. Create one small pilot migration first, likely the terminology instruction or the `propagation-check` skill pair, before converting the entire handbook to the new two-layer enforcement model.
2. Introduce a distinct section-to-appendix marker family so the linter can validate semantic composition separately from downstream propagation.
3. Split downstream files into `handbook-derived body` and `local glue` so only the handbook-derived body participates in strict appendix propagation.
4. Treat Copilot and Claude agent pairs as `shared body plus per-target frontmatter metadata`, not as whole-file mirrors, because the bodies are identical but frontmatter is intentionally different.
5. Apply the same `shared body plus per-target frontmatter metadata` pattern to skill pairs, which should be even easier because their bodies already match exactly.

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
