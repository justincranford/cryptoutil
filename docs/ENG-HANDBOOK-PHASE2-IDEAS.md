# ENG-HANDBOOK Phase 2 — Implementation Status

Last updated: 2026-05-26

---

## Goal

Two-tier propagation from `docs/ENG-HANDBOOK.md` to all Copilot and Claude artifacts:

**Tier 1** — Semantic grouping (by topic):
```
Narrative section → @section-to-appendix → Appendix D semantic group
```

**Tier 2** — Per-artifact assembly (by downstream file):
```
Appendix D semantic group → assembled into Appendix E per-artifact section → @propagate → actual file
```

The defining property of Tier 2: **a human reviewer jumps to the end of the handbook and sees
exactly what content each downstream artifact receives, assembled in one section per file.**

**Current status: Tier 1 is done. Tier 2 is not started.**

---

## 1. Implemented (Tier 1 — Complete)

### 1.1 Section-to-Appendix Marker System

All narrative sections contributing handbook-derived content are tagged with
`@section-to-appendix`. Appendix D has matching `@appendix-propagate` blocks for all 69 chunks
across 8 semantic groups.

| Group ID | Section | Chunks |
|----------|---------|--------|
| `execution-agent-behavior` | D.1 | 6 |
| `architecture-service-template` | D.2 | 4 |
| `security-authn-authz` | D.3 | 8 |
| `api-openapi-contracts` | D.4 | 2 |
| `observability-deployment-tooling` | D.5 | 10 |
| `testing-quality-golang` | D.6 | 16 |
| `terminology-instruction-body` | D.7 | 3 |
| `skills-handbook-coupled-body` | D.8 | 5 |

### 1.2 Linter Rules for Appendix D Coverage

Implemented in `validate_chunks.go` and `validate_coverage.go`:

1. ✅ Every `@section-to-appendix` chunk must have a matching `@appendix-propagate` block.
2. ✅ Every `@appendix-propagate` block must have a matching `@section-to-appendix` source.
3. ✅ Every `@appendix-propagate` must have an adjacent `why-this-exists` attribute.
4. ✅ A chunk with `@section-to-appendix` cannot also have a direct `@propagate`.
5. ✅ Orphan appendix blocks that do not propagate to any downstream target are rejected.
6. ✅ Orphan section chunks that feed no appendix block are rejected.
7. ✅ Unstable chunk IDs (section-number-based) are rejected.
8. ✅ Linter output is sorted by appendix reading order.

### 1.3 Instruction File Structural Markers

All 19 instruction files under `.github/instructions/` have:
- `<!-- @local-glue:start/end -->` around the title heading.
- `<!-- @handbook-derived-body:start/end -->` around all body content.

These markers are structural annotations only; the linter does not yet validate them.

### 1.4 `@source` Blocks in Downstream Files

All instruction files, agent files, and handbook-coupled skill files use
`<!-- @source from="docs/ENG-HANDBOOK.md" as="CHUNK_ID" -->` blocks. The `validate-chunks`
linter verifies byte-for-byte identity between the handbook and downstream content.

**Current state**: `@appendix-propagate` blocks in Appendix D currently point **directly** to
these downstream files in their `to=` attribute. This is a transitional state — in the completed
two-tier model, Appendix D would point to Appendix E sections instead.

### 1.5 Agent Pair Body Identity

`lint-agent-drift` (in `cicd-lint lint-docs`) validates that Copilot and Claude agent files have
identical body content. Four pairs enforced: `beast-mode`, `fix-workflows`,
`implementation-execution`, `implementation-planning`.

### 1.6 Skill Pair Body Identity

`lint-skill-command-drift` validates body identity + `## Key Rules` presence for all 13 skill
pairs.

---

## 2. Not Implemented (Tier 2 — Not Started)

### 2.1 Appendix E: Per-Artifact Sections

**Appendix E does not exist.**

This is the primary unachieved goal. Appendix E would contain one section per downstream
artifact (or one section per dual-canonical pair). Each section would assemble all chunks for
that artifact in one place and propagate them via `@propagate to="TARGET_FILE"`.

The planned structure (28 sections):
- E.1–E.19: One section per instruction file.
- E.20–E.23: One section per agent pair (4 pairs).
- E.24–E.28: One section per handbook-coupled skill pair (5 pairs).

See `ENG-HANDBOOK-NEW.md` §6 for the complete per-section breakdown.

### 2.2 Redirecting Appendix D `to=` Attributes

Currently, `@appendix-propagate` blocks in Appendix D have `to=` values pointing to actual
downstream files. In the two-tier model, these must change to reference Appendix E section
anchors, making Appendix D a semantic index only.

### 2.3 Tier 2 Linter Rules

No linter rules exist yet for the two-hop chain (Appendix D → Appendix E → actual file).
Required rules:
1. Every Appendix D chunk must appear in the correct Appendix E section.
2. Every Appendix E section must propagate to exactly the file(s) listed in its header.
3. Chunk content in Appendix E must match chunk content in Appendix D exactly.

---

## 3. Future Candidates (Post Tier 2)

These are valid improvements but must wait until Tier 2 is complete.

1. Move duplicated handbook prose currently hand-copied into agent bodies into appendix-backed
   fragments, so the same rule text is reused consistently. (Current agents embed large blocks
   verbatim without `@source` markers.)
2. Add a generated `docs/required-propagations-v2.yaml` that records both graph edges:
   `section → appendix` and `appendix → downstream`.
3. Add `@local-glue` / `@handbook-derived-body` linter enforcement (currently structural only).
4. Consider whether `CLAUDE.md` should migrate to `@source` blocks for its handbook-derived
   content (agent/skill/instruction inventory tables).
5. Validate that every `@appendix-propagate` chunk in Appendix E appears in `@source` order
   matching the instruction file's actual section order.

---

## 4. Rejected / Out of Scope

1. Appendix ownership index as a separate generated document (superseded by Appendix E itself).
2. Domain-family grouping of instruction subsections inside Appendix E (E.1–E.19 ordered by
   filename is sufficient for navigation).
3. Making agent and skill frontmatter generated metadata (manual frontmatter is fine).
