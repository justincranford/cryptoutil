# skill-scaffold

Create a conformant `.github/skills/NAME.md` Copilot skill file.

## Purpose

Use when adding a new Copilot skill to the project. Ensures correct flat naming,
useful content structure, and skill catalogue registration.

## Key Rules

- Flat naming: `kebab-case-name.md` in `.github/skills/`
- Referenced as `#kebab-case-name` in Copilot Chat
- No YAML frontmatter needed (unlike agents and instructions)
- Content: purpose, key rules, templates/examples, references

## Template

```markdown
# skill-name

One-line description of what this skill helps with.

## Purpose

When and why to use this skill (2-3 sentences).

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

See [ARCHITECTURE.md Section X.Y](../../docs/ARCHITECTURE.md#xy-anchor) for related standards.
```

## After Creating

1. Add entry to `.github/skills/README.md` skill table
2. Add entry to ARCHITECTURE.md Section 2.1.5 skill catalogue table
3. Update relevant agents' frontmatter `skills:` list (if agents reference skills)

## Reference Quality

Each skill MUST reference at least one ARCHITECTURE.md section relevant to its domain.

## References

See [ARCHITECTURE.md Section 2.1.5 Copilot Skills](../../docs/ARCHITECTURE.md#215-copilot-skills) for the skill catalogue and naming conventions.
