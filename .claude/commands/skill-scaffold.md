Create a new `.github/skills/NAME/SKILL.md` file with proper YAML frontmatter.

**Full Copilot original**: [.github/skills/skill-scaffold/SKILL.md](.github/skills/skill-scaffold/SKILL.md)

Provide: skill name (e.g., `rate-limiter-gen`), description, what it generates.

## Directory Structure

```
.github/skills/{skill-name}/
└── SKILL.md
```

The directory name MUST match the `name` field in SKILL.md exactly.

## SKILL.md Frontmatter

```yaml
---
name: {skill-name}               # MUST match directory name; max 64 chars; lowercase-hyphens
description: >                   # max 1024 chars; specific about capabilities AND use cases
  {What this skill does. When to use it. What it produces.
  Specific enough for auto-loading when the request matches.}
argument-hint: "[optional]"      # hint shown in chat input
user-invocable: true             # set false to hide from / menu (default: true)
disable-model-invocation: false  # set true for scaffold/ops-only skills (default: false)
---
```

**NEVER use `metadata:` sub-key** — it is not a valid SKILL.md field.

## Content Structure

```markdown
# {Skill Name}

## Purpose

One paragraph describing what this skill does.

## Usage

When to invoke: `/skill-name [args]`

## Template / Instructions

[The actual template or step-by-step guide]

## ARCHITECTURE.md References

- §X.Y — relevant section
```

## Registration

After creating the skill:
1. Add row to `.github/skills/README.md` catalog table
2. Add row to `.github/copilot-instructions.md` Available Skills table
3. Add row to `CLAUDE.md` Skills (Slash Commands) table
4. Create corresponding `.claude/commands/{skill-name}.md` for Claude Code

## Claude Code Bridge

Skills are automatically available in Claude Code as slash commands via `.claude/commands/`. The command file should capture the same template/instructions from SKILL.md.
