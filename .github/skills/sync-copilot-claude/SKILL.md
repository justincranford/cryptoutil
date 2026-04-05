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
- Checking migration status of legacy `.claude/commands/` to `.claude/skills/`

## Key Rules

- Copilot skills live at `.github/skills/<NAME>/SKILL.md`; Claude skills at `.claude/skills/<NAME>/SKILL.md`
- Body content MUST be identical between Copilot and Claude skill files
- Only allowed frontmatter differences: `tools:` / `allowed-tools:` field naming (Copilot vs Claude)
- Claude agents at `.claude/agents/<NAME>.md` must match Copilot agents at `.github/agents/<NAME>.agent.md`
- NEVER update only one file — always sync both in the same commit
- The `lint-agent-drift` linter (in `lint-docs`) enforces agent pair identity automatically
- After migration from `.claude/commands/` to `.claude/skills/`, delete the legacy command file

## Argument Meanings

| Argument | Action |
|----------|--------|
| `sync-copilot-claude` (no arg) | Audit all skills and agents for drift |
| `sync-copilot-claude all` | Sync all out-of-date pairs (audit + fix) |
| `sync-copilot-claude agents` | Sync agent pairs only |
| `sync-copilot-claude <name>` | Sync the named skill pair (e.g., `test-table-driven`) |
| `sync-copilot-claude status` | Show migration status of `.claude/commands/` legacy files |

## Workflow: Audit All Pairs

```bash
# Run lint-agent-drift to check agent drift
go run ./cmd/cicd-lint lint-docs

# Manual skill audit
python3 - << 'EOF'
import os, pathlib

gh_skills = sorted(pathlib.Path(".github/skills").glob("*/SKILL.md"))
claude_skills = sorted(pathlib.Path(".claude/skills").glob("*/SKILL.md"))

gh_names = {p.parent.name for p in gh_skills}
claude_names = {p.parent.name for p in claude_skills}

print("=== Missing Claude skills (exist in Copilot, not Claude) ===")
for name in sorted(gh_names - claude_names):
    print(f"  MISSING: .claude/skills/{name}/SKILL.md")

print("\n=== Orphaned Claude skills (exist in Claude, not Copilot) ===")
for name in sorted(claude_names - gh_names):
    print(f"  ORPHAN: .claude/skills/{name}/SKILL.md")

print("\n=== Body drift check ===")
for name in sorted(gh_names & claude_names):
    gh = pathlib.Path(f".github/skills/{name}/SKILL.md").read_text(encoding="utf-8")
    cl = pathlib.Path(f".claude/skills/{name}/SKILL.md").read_text(encoding="utf-8")
    # Strip frontmatter (everything up to and including second ---)
    def strip_fm(text):
        parts = text.split("---", 2)
        return parts[2].strip() if len(parts) == 3 else text.strip()
    if strip_fm(gh) != strip_fm(cl):
        print(f"  DRIFT: {name}")
    else:
        print(f"  OK: {name}")
EOF
```

## Workflow: Create Missing Claude Skill

```bash
# For skill NAME that exists in .github/skills/NAME/ but not .claude/skills/NAME/
mkdir -p .claude/skills/NAME

# Copy Copilot skill as base
cp .github/skills/NAME/SKILL.md .claude/skills/NAME/SKILL.md

# Adapt frontmatter if needed (Claude uses allowed-tools:, Copilot uses tools:)
# Body stays IDENTICAL

# Verify
diff <(tail -n +4 .github/skills/NAME/SKILL.md | sed '1,/^---$/d') \
     <(tail -n +4 .claude/skills/NAME/SKILL.md | sed '1,/^---$/d')
```

## Workflow: Migrate Legacy Command → Skill

For each file in `.claude/commands/<name>.md` that lacks a `.claude/skills/<name>/` counterpart:

1. Create the skill directory: `mkdir -p .claude/skills/<name>`
2. Copy and adapt: `cp .claude/commands/<name>.md .claude/skills/<name>/SKILL.md`
3. Ensure the Claude skill frontmatter is compatible with Claude Code
4. Test: invoke `/<name>` in a Claude Code session
5. Delete: `rm .claude/commands/<name>.md`
6. Verify `lint-docs` passes: `go run ./cmd/cicd-lint lint-docs`
7. Commit both the skill creation and the command deletion together

## Legacy Command Migration Status

Check which `.claude/commands/` files still need migrating:

```bash
# Show all legacy commands without a corresponding skill
for f in .claude/commands/*.md; do
  name=$(basename "$f" .md)
  if [ ! -d ".claude/skills/$name" ]; then
    echo "NEEDS MIGRATION: $name"
  else
    echo "ALREADY MIGRATED: $name"
  fi
done
```

## References

Copilot ↔ Claude dual canonical pairs are enforced by:
- `lint-agent-drift` (via `go run ./cmd/cicd-lint lint-docs`) — enforces agent pairs
- `lint-skill-command-drift` — enforces skill pairs (check `.claude/skills/` post-migration)

See `docs/framework-v8/claude.md` for the full Claude Code file structure reference
and frontmatter options for both skills and agents.
