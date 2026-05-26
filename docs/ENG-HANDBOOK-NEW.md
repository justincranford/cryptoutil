---
title: cryptoutil Architecture - Proposed Structured Handbook
version: 0.3
date: 2026-05-26
status: Phase 2 In Progress - Tier 1 Done, Tier 2 Not Started
purpose: New handbook structure separating semantic sections from per-artifact appendix mirrors.
---

# cryptoutil Architecture - Proposed Structured Handbook

## 1. Scope

This document tracks the redesign of [docs/ENG-HANDBOOK.md](docs/ENG-HANDBOOK.md).
It does not replace the handbook — it defines what the final structure must look like
and records what has been built versus what remains.

### 1.1 Goal: Two-Tier Propagation

The target propagation model has two tiers:

**Tier 1** — Semantic grouping (by topic):
> Narrative section content → `@section-to-appendix` → Appendix D semantic group

**Tier 2** — Per-artifact assembly (by downstream target):
> Appendix D semantic group → assembled into **Appendix E per-artifact section** → `@propagate` → actual Copilot/Claude file

The intended chain is:

```
Narrative sections
    ↓ @section-to-appendix
Appendix D (organized by semantic topic)
    ↓ @appendix-propagate (feeds)
Appendix E (organized by downstream artifact — one section per file)
    ↓ @propagate
Actual .github/instructions/, .github/agents/, .claude/agents/, .github/skills/, .claude/skills/ files
```

The key property of Appendix E: **a human reviewer can jump to the end of the document and see exactly what content each Copilot/Claude artifact file receives, assembled in one place per file.**

### 1.2 Current State (as of 2026-05-26)

**Tier 1 is implemented.**

Appendix D exists in `ENG-HANDBOOK.md`. All 69 chunks across 7 semantic groups are tagged with
`@section-to-appendix` in their narrative sections and have matching `@appendix-propagate`
blocks in Appendix D. The linter enforces bi-directional coverage.

**Tier 2 is NOT implemented.**

Appendix E (per-artifact sections) does not exist.
The `@appendix-propagate` blocks in Appendix D currently point directly to downstream files,
bypassing the per-artifact assembly layer entirely.
A human reviewer cannot jump to the end of the handbook and see the assembled content
for a specific instruction file or agent file.

### 1.3 What Must Be Built

To complete Tier 2:

1. **Appendix E structure**: Add a new appendix at the end of `ENG-HANDBOOK.md` with one section
   per downstream artifact file. Sections are organized by artifact family:
   - Instruction files (19 files)
   - Agent pairs (4 Copilot + 4 Claude = 8 files, documented as 4 pairs)
   - Skill pairs (13 Copilot + 13 Claude = 26 files, documented as 13 pairs)
2. **Per-artifact content**: Each Appendix E section contains the full assembled verbatim text
   for that artifact, drawn from the semantic chunks already present in Appendix D.
3. **`@propagate` in Appendix E**: Each Appendix E section has a `@propagate to="TARGET_FILE"`
   block (or blocks for paired files) that pushes the assembled content to the actual file.
4. **Redirect Appendix D**: The `to=` attributes in Appendix D `@appendix-propagate` blocks
   change from `to="ACTUAL_FILE"` to `to="APPENDIX_E_SECTION_ANCHOR"`. Appendix D becomes
   a semantic index; Appendix E becomes the distribution surface.
   (Alternatively: Appendix D `@appendix-propagate` retains the actual file as a semantic
   annotation, and Appendix E `@propagate` blocks are the canonical live propagation path.
   This is an implementation decision to be resolved in Phase 2.)
5. **Linter support**: `validate_coverage.go` needs rules to verify the two-hop chain:
   every chunk tagged for a semantic group must appear in the relevant Appendix E sections,
   and every Appendix E section must propagate to its target file.
6. **Downstream files updated**: The `@source` blocks in actual instruction/agent/skill files
   must remain in sync with the Appendix E content (not Appendix D content).

## 2. Problems In The Current Handbook

The current handbook is strong as a single source of truth but has three maintenance costs.

1. Propagation blocks are distributed across Appendix D semantic groups, so downstream
   relationships are locally correct but globally hard to inspect per-file.
2. The `@appendix-propagate to=` attribute lists many files on a single block — when a
   chunk goes to 9 targets, that target list must be maintained in Appendix D but the
   per-file view is invisible.
3. The `@source` blocks in instruction and agent files reference semantic chunk IDs rather
   than a per-file appendix section, so the relationship between "what is in the file" and
   "where it came from" requires cross-referencing Appendix D.

## 3. Design Objectives

The new structure optimizes for four things.

1. **Human readability**: core architecture content reads top-down without downstream file
   details embedded in it.
2. **Reviewer navigability**: jump to Appendix E at the end to see any single artifact's
   complete assembled content.
3. **Referential integrity**: every downstream artifact has exactly one Appendix E section
   as its home. Linter proves completeness.
4. **Two-tier linter enforcement**: the linter validates both tiers independently — semantic
   completeness (every chunk has a group) and distribution completeness (every artifact has
   an Appendix E section that propagates to the actual file).

## 4. Marker Taxonomy

The marker system uses three families.

### 4.1 Tier 1 Markers (implemented)

In the handbook narrative section where content originates:

```html
<!-- @section-to-appendix to="GROUP_ID" as="CHUNK_ID" -->
... semantic content (readable in-place) ...
<!-- @/section-to-appendix -->
```

In Appendix D, one block per chunk:

```html
<!-- @appendix-propagate from="GROUP_ID" to="TARGET_FILE" as="CHUNK_ID" why-this-exists="..." -->
... verbatim chunk content ...
<!-- @/appendix-propagate -->
```

> **Note**: `to="TARGET_FILE"` currently points to actual downstream files. In the completed
> two-tier model, this will change to an Appendix E section anchor instead (or be retained
> as a semantic annotation while Appendix E carries the live `@propagate`).

### 4.2 Tier 2 Markers (not yet implemented)

In each Appendix E per-artifact section, one propagation block per chunk:

```html
<!-- @propagate to="TARGET_FILE" as="CHUNK_ID" -->
... verbatim chunk content (copied from Appendix D) ...
<!-- @/propagate -->
```

For paired files (agent pairs, skill pairs), the `to=` lists both targets:

```html
<!-- @propagate to=".github/agents/NAME.agent.md, .claude/agents/NAME.md" as="CHUNK_ID" -->
```

### 4.3 Downstream File Markers (implemented)

In actual Copilot/Claude files, each propagated region is bracketed:

```html
<!-- @source from="docs/ENG-HANDBOOK.md" as="CHUNK_ID" -->
... verbatim content (must match handbook exactly) ...
<!-- @/source -->
```

Instruction files additionally use structural region markers:

```html
<!-- @local-glue:start -->
# File Title
<!-- @local-glue:end -->
<!-- @handbook-derived-body:start -->
[all @source blocks and bridge text]
<!-- @handbook-derived-body:end -->
```

## 5. Appendix D: Semantic Groups (current state)

Appendix D is already built. It contains 8 semantic groups.

| Group ID | Section | Chunk count | Downstream consumers |
|----------|---------|-------------|----------------------|
| `execution-agent-behavior` | D.1 | 6 | beast-mode, evidence-based, agent-format instructions; all 4 agent pairs |
| `architecture-service-template` | D.2 | 4 | architecture, data-infrastructure, testing instructions |
| `security-authn-authz` | D.3 | 8 | security, authn instructions |
| `api-openapi-contracts` | D.4 | 2 | openapi instruction; openapi-codegen skill |
| `observability-deployment-tooling` | D.5 | 10 | observability, deployment, cross-platform, linting instructions; beast-mode, fix-workflows agents |
| `testing-quality-golang` | D.6 | 16 | testing, coding, golang, git, cross-platform instructions; skill pairs |
| `terminology-instruction-body` | D.7 | 3 | terminology instruction |
| `skills-handbook-coupled-body` | D.8 | 5 | propagation-check, sync-copilot-claude, copilot-customization, test-table-driven, openapi-codegen skill pairs |

Each `@appendix-propagate` block in Appendix D propagates a single chunk to one or more
downstream files. Currently those files are the actual Copilot/Claude files. In the
two-tier model they would reference Appendix E sections instead.

## 6. Appendix E: Per-Artifact Sections (planned, not built)

Appendix E is the missing tier. It must be added to the end of `ENG-HANDBOOK.md`
after Appendix D. Each subsection corresponds to exactly one downstream artifact
(or one dual-canonical pair).

### 6.1 Instruction File Sections (19 files → E.1–E.19)

| Section | Target File | Source groups (chunks) |
|---------|-------------|------------------------|
| E.1 | `01-01.terminology.instructions.md` | D.7 (`rfc-2119-keywords`, `emphasis-keywords`, `abbreviations`) |
| E.2 | `01-02.beast-mode.instructions.md` | D.1 (`quality-attributes`, `end-of-turn-commit-protocol`) |
| E.3 | `02-01.architecture.instructions.md` | D.2 (`service-framework-components`, `tls-provision-mode`) |
| E.4 | `02-02.versions.instructions.md` | D.6 (`minimum-versions`) |
| E.5 | `02-03.observability.instructions.md` | D.5 (`otel-collector-constraints`) |
| E.6 | `02-04.openapi.instructions.md` | D.4 (`base-initialisms`, `http-status-codes`) |
| E.7 | `02-05.security.instructions.md` | D.3 (`secrets-detection-strategy`, `tls-client-policy`) |
| E.8 | `02-06.authn.instructions.md` | D.3 (`key-principles`, `session-token-formats`, `headless-authn`, `browser-authn`, `mfa-combinations`, `authz-methods`) |
| E.9 | `03-01.coding.instructions.md` | D.6 (`validator-error-aggregation`, `format-go-protection`) |
| E.10 | `03-02.testing.instructions.md` | D.2 (`three-tier-database-strategy`), D.6 (`test-file-suffixes`, `production-closure-body-coverage`, `sequential-test-exemption`, `disable-keep-alives-test-transport`, `timeout-double-multiplication-antipattern`, `postgres-mtls-client-identity`, `mutation-common-survivors`) |
| E.11 | `03-03.golang.instructions.md` | D.6 (`crypto-acronyms-caps`) |
| E.12 | `03-04.data-infrastructure.instructions.md` | D.2 (`three-tier-database-strategy`, `sqlite-barrier-outside-tx`) |
| E.13 | `03-05.linting.instructions.md` | D.5 (`utf8-without-bom`, `cicd-bulk-hook-architecture`) |
| E.14 | `04-01.deployment.instructions.md` | D.5 (`docker-compose-rules`, `cicd-command-naming`, `cicd-lint-constraints`, `docker-compose-verification-in-scope`) |
| E.15 | `05-01.cross-platform.instructions.md` | D.5 (`docker-desktop-startup`, `docker-desktop-upgrade`), D.6 (`scripting-language-policy`) |
| E.16 | `05-02.git.instructions.md` | D.6 (`conventional-commits`, `incremental-commits`, `restore-from-baseline`, `platform-line-ending-operations`) |
| E.17 | `06-01.evidence-based.instructions.md` | D.1 (`mandatory-review-passes`, `per-task-status-updates`), D.5 (`infrastructure-blocker-escalation`) |
| E.18 | `06-02.agent-format.instructions.md` | D.1 (`agent-self-containment`) |
| E.19 | `06-03.tool-efficiency.instructions.md` | D.6 (`tool-preference-order`) |

### 6.2 Agent Pair Sections (4 pairs → E.20–E.23)

Agent sections propagate to both the Copilot and Claude canonical files simultaneously.

| Section | Pair | Copilot target | Claude target | Source chunks |
|---------|------|----------------|---------------|---------------|
| E.20 | beast-mode | `.github/agents/beast-mode.agent.md` | `.claude/agents/beast-mode.md` | D.1 (`mandatory-review-passes`), D.5 (`platform-line-ending-operations`, `cicd-bulk-hook-architecture`) |
| E.21 | fix-workflows | `.github/agents/fix-workflows.agent.md` | `.claude/agents/fix-workflows.md` | D.1 (`mandatory-review-passes`), D.5 (`platform-line-ending-operations`) |
| E.22 | implementation-execution | `.github/agents/implementation-execution.agent.md` | `.claude/agents/implementation-execution.md` | D.1 (`mandatory-review-passes`, `per-task-status-updates`, `lessons-md-structure`), D.5 (`platform-line-ending-operations`, `docker-compose-verification-in-scope`) |
| E.23 | implementation-planning | `.github/agents/implementation-planning.agent.md` | `.claude/agents/implementation-planning.md` | D.1 (`mandatory-review-passes`, `per-task-status-updates`, `lessons-md-structure`), D.5 (`platform-line-ending-operations`, `docker-compose-verification-in-scope`) |

### 6.3 Skill Pair Sections (5 handbook-coupled pairs → E.24–E.28)

Only the handbook-coupled skills have Appendix D chunks today. The other 8 skill pairs
have self-contained bodies not derived from handbook chunks.

| Section | Pair | Copilot target | Claude target | Source chunk |
|---------|------|----------------|---------------|--------------|
| E.24 | propagation-check | `.github/skills/propagation-check/SKILL.md` | `.claude/skills/propagation-check/SKILL.md` | D.8 (`skill-propagation-check-core-rules`) |
| E.25 | sync-copilot-claude | `.github/skills/sync-copilot-claude/SKILL.md` | `.claude/skills/sync-copilot-claude/SKILL.md` | D.8 (`skill-sync-copilot-claude-core-rules`) |
| E.26 | copilot-customization | `.github/skills/copilot-customization/SKILL.md` | `.claude/skills/copilot-customization/SKILL.md` | D.8 (`skill-copilot-customization-core-rules`) |
| E.27 | test-table-driven | `.github/skills/test-table-driven/SKILL.md` | `.claude/skills/test-table-driven/SKILL.md` | D.8 (`skill-test-table-driven-core-rules`) |
| E.28 | openapi-codegen | `.github/skills/openapi-codegen/SKILL.md` | `.claude/skills/openapi-codegen/SKILL.md` | D.8 (`skill-openapi-codegen-core-rules`) |

### 6.4 Example Appendix E Section Structure

```markdown
### E.2 01-02.beast-mode.instructions.md

Target: `.github/instructions/01-02.beast-mode.instructions.md`

Chunks assembled from: D.1 (`quality-attributes`, `end-of-turn-commit-protocol`)

<!-- @propagate to=".github/instructions/01-02.beast-mode.instructions.md" as="quality-attributes" -->
**Quality Attributes (NO EXCEPTIONS)**:
- ✅ Correctness: ALL code functionally correct with comprehensive tests
...
<!-- @/propagate -->

<!-- @propagate to=".github/instructions/01-02.beast-mode.instructions.md" as="end-of-turn-commit-protocol" -->
**MANDATORY: NEVER end a turn with uncommitted changes...**
...
<!-- @/propagate -->
```

And the instruction file contains matching `@source` blocks:

```markdown
<!-- @source from="docs/ENG-HANDBOOK.md" as="quality-attributes" -->
...identical content...
<!-- @/source -->
```

## 7. Implementation Path

To complete the two-tier model, work must proceed in this order.

### Step 1: Decide the `@propagate` grammar for Appendix E

Choose between:

- **Option A**: Legacy `@propagate to="FILE" as="CHUNK_ID"` (already parsed by the linter). Lower
  effort, avoids introducing a third marker family.
- **Option B**: New `@artifact-propagate from="E.N" to="FILE" as="CHUNK_ID"` (new marker family).
  Enables stronger two-hop linter validation but requires linter changes before Appendix E can exist.

Start with Option A unless the Tier 2 linter rules are being built in the same phase.

### Step 2: Build Appendix E in ENG-HANDBOOK.md

Add Appendix E at the end of `ENG-HANDBOOK.md`, after Appendix D. Populate all 28 sections
(E.1–E.28) with `@propagate` blocks containing the verbatim chunk content copied from the
matching `@appendix-propagate` blocks in Appendix D.

### Step 3: Update `@appendix-propagate` `to=` in Appendix D

Change the `to=` attributes in Appendix D `@appendix-propagate` blocks from actual downstream
file paths to Appendix E section anchors. This makes Appendix D a pure semantic index and
Appendix E the live distribution surface.

### Step 4: Verify downstream `@source` blocks are unchanged

The downstream files currently use `@source from="docs/ENG-HANDBOOK.md" as="CHUNK_ID"`.
The chunk IDs are stable. The `validate-chunks` linter verifies content identity regardless
of which appendix section is the canonical source. No change to downstream files is required
unless the chunk content itself changes.

### Step 5: Add Tier 2 linter rules

Add a new check in `validate_coverage.go` that:
1. Parses all Appendix E `@propagate` blocks → collects (target_file, chunk_id, content).
2. Verifies every Appendix D chunk is represented in the correct Appendix E section.
3. Verifies every Appendix E section propagates to exactly the file(s) listed in its header.
4. Verifies the chunk content in Appendix E matches the chunk content in Appendix D exactly.

## 8. Structural Markers Already In Place

These markers are already applied and do not need to be re-added.

### 8.1 Instruction File Structural Markers

All 19 instruction files under `.github/instructions/` use:

```html
<!-- @local-glue:start -->
# File Title
<!-- @local-glue:end -->
<!-- @handbook-derived-body:start -->
[all @source blocks and bridge text]
<!-- @handbook-derived-body:end -->
```

These markers are structural annotations only. They are not currently validated by
the `lint-docs` linter. Future enforcement will ensure the `@handbook-derived-body`
region contains only `@source` blocks plus minimal local bridge text.

### 8.2 Agent Pair Body Identity

`lint-agent-drift` validates that Copilot and Claude agent files have identical body
content after the frontmatter delimiter. Permitted frontmatter differences:
`name:`, `tools:`, `handoffs:`, `model:`, `argument-hint:`.

Four pairs enforced: `beast-mode`, `fix-workflows`, `implementation-execution`,
`implementation-planning`.

### 8.3 Skill Pair Body Identity

`lint-skill-command-drift` validates that Copilot and Claude skill files have identical
body content and that both files contain a `## Key Rules` section.

Thirteen pairs enforced (see PHASE2-IDEAS.md §3 for full list).

## 9. Why CLAUDE.md Lives With Instructions

`CLAUDE.md` is the loader that enumerates and imports the Copilot instruction set for
Claude-side use. For maintenance purposes it belongs beside the instruction appendix
(Appendix E instruction sections), not beside agents or skills.

`CLAUDE.md` does not currently use `@source` blocks for its handbook-derived content.
It is a candidate for future migration once Appendix E instruction sections are live.
