---
name: instruction-scaffold
description: "Create a conformant .github/instructions/NN-NN.name.instructions.md file. Use when adding project-specific instruction files to ensure correct YAML frontmatter, numbering scheme, @source propagation blocks, and ARCHITECTURE.md cross-references."
argument-hint: "[NN-NN.name description]"
---

Create a new `.github/instructions/NN-NN.name.instructions.md` file.

**Full Copilot original**: [.github/skills/instruction-scaffold/SKILL.md](.github/skills/instruction-scaffold/SKILL.md)

Provide: number (e.g., `03-06`), name (e.g., `error-handling`), topic description.

## Key Rules

- Filename pattern: `NN-NN.name.instructions.md` (two-part numeric prefix, dot-separated name)
- YAML frontmatter REQUIRED: `description:` and `applyTo:` fields
- Use `<!-- @source from="docs/ARCHITECTURE.md" as="chunk-id" -->` for propagated content
- Content in `@source` blocks MUST be byte-for-byte identical to ARCHITECTURE.md `@propagate` blocks
- Every section using ARCHITECTURE.md content MUST have a `See [ARCHITECTURE.md §X.Y](...)` reference
- Run `go run ./cmd/cicd-lint lint-docs` to validate propagation integrity after creating

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
