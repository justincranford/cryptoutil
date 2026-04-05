---
name: sync-copilot-claude
description: "Keep Copilot and Claude AI configuration files synchronized. Use after adding/updating Copilot skills (.github/skills/NAME/SKILL.md) to create or update the matching Claude skill (.claude/skills/NAME/SKILL.md), or to audit all pairs for drift. Also checks agent pairs (Copilot .agent.md vs Claude .md)."
argument-hint: "[skill-name | 'all' | 'agents' | 'status']"
---

Synchronize Copilot skills and agents with their Claude counterparts in one pass.

**Full Copilot original**: [.github/skills/sync-copilot-claude/SKILL.md](.github/skills/sync-copilot-claude/SKILL.md)

**Preferred Claude format**: [.claude/skills/sync-copilot-claude/SKILL.md](.claude/skills/sync-copilot-claude/SKILL.md)

> **Note**: This command file is a legacy placeholder. The preferred Claude format is
> `.claude/skills/sync-copilot-claude/SKILL.md`. Use that skill instead. The `.claude/commands/`
> format will be migrated to `.claude/skills/` per `docs/framework-v8/carryover.md` item 2.1.

## Key Rules

- Copilot skills live at `.github/skills/<NAME>/SKILL.md`; Claude skills at `.claude/skills/<NAME>/SKILL.md`
- Body content MUST be identical between Copilot and Claude skill files
- Only allowed frontmatter differences: `tools:` / `allowed-tools:` field naming (Copilot vs Claude)
- Claude agents at `.claude/agents/<NAME>.md` must match Copilot agents at `.github/agents/<NAME>.agent.md`
- NEVER update only one file — always sync both in the same commit
- The `lint-agent-drift` linter (in `lint-docs`) enforces agent pair identity automatically
- After migration from `.claude/commands/` to `.claude/skills/`, delete the legacy command file
