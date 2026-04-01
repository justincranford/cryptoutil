Create a new `.github/instructions/NN-NN.name.instructions.md` file.

**Full Copilot original**: [.github/skills/instruction-scaffold/SKILL.md](.github/skills/instruction-scaffold/SKILL.md)

Provide: number (e.g., `03-06`), name (e.g., `error-handling`), topic description.

## Numbering Scheme

| Range | Category |
|-------|----------|
| 01-xx | Core concepts (terminology, execution modes) |
| 02-xx | Architecture (service patterns, security, API, authn) |
| 03-xx | Development (coding, testing, Go, data, linting) |
| 04-xx | Deployment (CI/CD, Docker) |
| 05-xx | Platform (cross-platform, git) |
| 06-xx | Quality (evidence-based, agent format) |

## File Template

```markdown
---
description: {One-line description of what this instruction file covers}
applyTo: "**"
---

# {Topic Name}

<!-- @source from="docs/ARCHITECTURE.md" as="{block-id}" -->
{Verbatim content propagated from ARCHITECTURE.md @propagate block}
<!-- @/source -->

## Additional Context

{Non-propagated content: headings, transitions, cross-references, examples}

See [ARCHITECTURE.md §X.Y](../docs/ARCHITECTURE.md#xy-section-anchor) for full context.
```

## Registration

After creating the file, register it in:
1. `.github/copilot-instructions.md` — add row to Instruction Files Reference table
2. `CLAUDE.md` — add `@.github/instructions/NN-NN.name.instructions.md` import line

## @source Block Rules

- Content inside `@source` blocks MUST be byte-for-byte identical to ARCHITECTURE.md `@propagate` blocks
- Non-propagated glue (headings, `See` cross-references) goes OUTSIDE `@source` blocks
- No markdown links inside `@source` blocks
- Run `go run cmd/cicd-lint/main.go lint-docs` to validate
