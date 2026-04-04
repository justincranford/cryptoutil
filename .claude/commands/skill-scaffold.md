---
name: skill-scaffold
description: "Create a conformant .github/skills/NAME/SKILL.md Agent Skill with proper YAML frontmatter. Use when adding a new Copilot skill to ensure correct subdirectory structure, required name/description fields, and skill catalogue registration in ARCHITECTURE.md and README.md."
argument-hint: "[skill-name description]"
---

Create a new `.github/skills/NAME/SKILL.md` file with proper YAML frontmatter.

**Full Copilot original**: [.github/skills/skill-scaffold/SKILL.md](.github/skills/skill-scaffold/SKILL.md)

Provide: skill name (e.g., `rate-limiter-gen`), description, what it generates.

## Key Rules

- **Subdirectory structure**: `.github/skills/kebab-case-name/SKILL.md`
- **`name` must match directory**: `name: kebab-case-name` = directory `kebab-case-name/`
- **`name` constraints**: lowercase, hyphens for spaces, max 64 chars
- **`description` required**: specific about both capabilities AND use cases, max 1024 chars
- **`argument-hint` optional**: shown in chat when invoked as `/skill-name`
- Content: purpose, key rules, templates/examples, ARCHITECTURE.md references
- No YAML frontmatter duplicates from instructions (skills are on-demand, instructions always-on)

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
