---
name: sync-copilot-claude
description: "Keep Copilot and Claude AI configuration files synchronized. Use after adding/updating Copilot skills (.github/skills/NAME/SKILL.md) to create or update the matching Claude skill (.claude/skills/NAME/SKILL.md), or to audit all pairs for drift. Also checks agent pairs (Copilot .agent.md vs Claude .md)."
argument-hint: "[skill-name | 'all' | 'agents' | 'status']"
---

Synchronize Copilot skills and agents with their Claude counterparts in one pass.

## Purpose

Use when:
- Adding a new Copilot skill → need to create matching Claude skill
- Updating a Copilot skill body → propagate changes to Claude skill
- Auditing all pairs for drift before a commit

## Key Rules

- Copilot skills live at `.github/skills/<NAME>/SKILL.md`; Claude skills at `.claude/skills/<NAME>/SKILL.md`
- Body content MUST be identical between Copilot and Claude skill files
- Only allowed frontmatter differences: `tools:` / `allowed-tools:` field naming (Copilot vs Claude)
- Claude agents at `.claude/agents/<NAME>.md` must match Copilot agents at `.github/agents/<NAME>.agent.md`
- NEVER update only one file — always sync both in the same commit
- The `lint-agent-drift` linter (in `lint-docs`) enforces agent pair identity automatically
- Verify discoverability after sync: update `.github/skills/README.md`, `.github/copilot-instructions.md`, `CLAUDE.md`, and `docs/ENG-HANDBOOK.md` when a new skill or agent should appear there
- Flag overlap explicitly: if two skills now describe the same creation or audit workflow, merge or narrow them in the same change instead of preserving redundant catalog entries
- If a skill becomes redundant after a merge, remove the dead catalog entries and orphaned directories in the same commit
- When syncing planning agents, also verify planning-triad readiness safeguards are present in BOTH files: `plan.md` + `tasks.md` + `lessons.md` consistency gate and false-ready prohibition
- If planning agents changed but triad safeguards are missing in either side, treat as drift and fix in the same commit
- Use `agent-tools-maintenance` first when the change scope includes Copilot agent `tools:` allowlist updates

## Argument Meanings

| Argument | Action |
|----------|--------|
| `sync-copilot-claude` (no arg) | Audit all skills and agents for drift |
| `sync-copilot-claude all` | Sync all out-of-date pairs (audit + fix) |
| `sync-copilot-claude agents` | Sync agent pairs only |
| `sync-copilot-claude <name>` | Sync the named skill pair (e.g., `test-table-driven`) |

## Workflow: Audit All Pairs

```bash
# Run the canonical drift validator
go run ./cmd/cicd-lint lint-docs
```

## Workflow: Create Missing Claude Skill

```bash
# Create the missing .claude/skills/NAME/SKILL.md pair in the same change
# Keep description and argument-hint identical to the Copilot skill
# Keep the body byte-identical
# Omit Copilot-only frontmatter fields from the Claude file
# Re-run go run ./cmd/cicd-lint lint-docs until lint-skill-command-drift passes
```

## Catalog Review After Sync

After the pair is in sync, verify the surrounding catalog stays coherent:

- README entry exists and points at the correct skill path
- Copilot and Claude command tables describe the same artifact consistently
- Merged or retired skills no longer appear in handbook tables or target-structure docs
- The synced skill does not duplicate the purpose of an adjacent skill without a clear scope boundary

## References

Copilot ↔ Claude dual canonical pairs are enforced by:
- `lint-agent-drift` (via `go run ./cmd/cicd-lint lint-docs`) — enforces agent pairs
- `lint-skill-command-drift` — enforces skill pairs

See [ENG-HANDBOOK.md Section 2.1.5 Copilot Skills](../../../docs/ENG-HANDBOOK.md#215-copilot-skills) for the active skill catalogue and [ENG-HANDBOOK.md Section 13.4 Documentation Propagation Strategy](../../../docs/ENG-HANDBOOK.md#134-documentation-propagation-strategy) for same-commit documentation update expectations.
