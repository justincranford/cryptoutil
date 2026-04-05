---
name: skill-scaffold
description: "Create a conformant .github/skills/NAME/SKILL.md Agent Skill with proper YAML frontmatter. Use when adding a new Copilot skill to ensure correct subdirectory structure, required name/description fields, and skill catalogue registration in ARCHITECTURE.md and README.md."
argument-hint: "[skill-name description]"
---

Create a conformant `.github/skills/NAME/SKILL.md` Agent Skill with proper YAML frontmatter.

## Purpose

Use when adding a new Copilot skill to the project. Ensures correct VS Code Agent
Skills structure: each skill lives in its own subdirectory with a `SKILL.md` file
containing YAML frontmatter. The `name` field **must match** the subdirectory name.

## Key Rules

- **Subdirectory structure**: `.github/skills/kebab-case-name/SKILL.md`
- **`name` must match directory**: `name: kebab-case-name` = directory `kebab-case-name/`
- **`name` constraints**: lowercase, hyphens for spaces, max 64 chars
- **`description` required**: specific about both capabilities AND use cases, max 1024 chars
- **`argument-hint` optional**: shown in chat when invoked as `/skill-name`
- Content: purpose, key rules, templates/examples, ARCHITECTURE.md references
- No YAML frontmatter duplicates from instructions (skills are on-demand, instructions always-on)

## Template

```markdown
---
name: kebab-case-name
description: "What it does AND when to use it. Max 1024 chars. Be specific."
argument-hint: "[optional hint text]"
# Optional: disable-model-invocation: true  # for scaffold/ops-only skills
# Optional: user-invocable: false          # for background knowledge skills
---

## Purpose

When and why to use this skill (2-3 sentences with concrete context).

## Key Rules

- Rule 1: description
- Rule 2: description

## Template / Pattern

\`\`\`go
// actual code template for the task
\`\`\`

## Examples

Brief before/after or input/output example.

## References

Read [ARCHITECTURE.md Section X.Y](../../../docs/ARCHITECTURE.md#xy-anchor) for related standards — follow all requirements from this section when implementing the skill's functionality.
```

## Directory Structure

```
.github/skills/
└── my-new-skill/          # Directory name MUST match `name` in SKILL.md
    └── SKILL.md           # Required, contains frontmatter + instructions
    └── example.go         # Optional: scripts, templates, examples
    └── examples/          # Optional: additional resources
```

## Mandatory Checklist

- [ ] YAML frontmatter with `name` (matches directory name exactly), `description`, optional `argument-hint`
- [ ] `disable-model-invocation: true` for scaffold/ops-only skills; omit for auto-loadable general skills
- [ ] `user-invocable: false` only for background knowledge skills hidden from / menu
- [ ] At least one `Read [ARCHITECTURE.md ...]` reference relevant to the skill domain
- [ ] Entry added to `.github/skills/README.md` skill table
- [ ] Entry added to ARCHITECTURE.md Section 2.1.5 skill catalogue table

## After Creating

1. Add entry to `.github/skills/README.md` skill table
2. Add entry to ARCHITECTURE.md Section 2.1.5 skill catalogue table (update path `skills/NAME/SKILL.md`)
3. Update relevant agents' frontmatter `skills:` list (if agents reference skills)
4. Run `go run ./cmd/cicd-lint lint-docs` to validate cross-references

## References

Read [ARCHITECTURE.md Section 2.1.5 Copilot Skills](../../../docs/ARCHITECTURE.md#215-copilot-skills) for the skill catalogue and naming conventions — add the new skill to both the skill catalogue table and `.github/skills/README.md` using the format established in this section.
