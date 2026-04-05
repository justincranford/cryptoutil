---
name: instruction-scaffold
description: "Create a conformant .github/instructions/NN-NN.name.instructions.md file. Use when adding project-specific instruction files to ensure correct YAML frontmatter, numbering scheme, @source propagation blocks, and ENG-HANDBOOK.md cross-references."
argument-hint: "[NN-NN.name description]"
disable-model-invocation: true
---

Create a conformant `.github/instructions/NN-NN.name.instructions.md` file.

## Purpose

Use when adding a new instruction file to the project. Ensures correct YAML
frontmatter, numbering scheme, and @source blocks for propagated content.

## Key Rules

- Filename pattern: `NN-NN.name.instructions.md` (two-part numeric prefix, dot-separated name)
- YAML frontmatter REQUIRED: `description:` and `applyTo:` fields
- Use `<!-- @source from="docs/ENG-HANDBOOK.md" as="chunk-id" -->` for propagated content
- Content in `@source` blocks MUST be byte-for-byte identical to ENG-HANDBOOK.md `@propagate` blocks
- Every section using ENG-HANDBOOK.md content MUST have a `See [ENG-HANDBOOK.md §X.Y](...)` reference
- Run `go run ./cmd/cicd-lint lint-docs` to validate propagation integrity after creating

## Numbering Scheme

| Range | Category |
|-------|---------|
| 01-01 to 01-99 | Core concepts (terminology, execution mode) |
| 02-01 to 02-99 | Architecture (services, versions, observability, API, security, authn) |
| 03-01 to 03-99 | Development (coding, testing, Go, data infra, linting) |
| 04-01 to 04-99 | Deployment (CI/CD, Docker) |
| 05-01 to 05-99 | Platform (cross-platform, git) |
| 06-01 to 06-99 | Quality (evidence-based, agent format) |

## Template

```markdown
---
description: "Short description of what this instruction covers"
applyTo: "**"
---
# Title

## Section One

Content here.

## Section Two

<!-- @source from="docs/ENG-HANDBOOK.md" as="chunk-id" -->
Content verbatim from ENG-HANDBOOK.md
<!-- @/source -->

See [ENG-HANDBOOK.md Section X.Y](../../../docs/ENG-HANDBOOK.md#xy-anchor) for complete documentation.
```

## Mandatory Checklist

- [ ] YAML frontmatter with `description` and `applyTo`
- [ ] Numbered correctly (no gaps, no conflicts)
- [ ] Referenced in `.github/copilot-instructions.md` instruction file table
- [ ] `@source` blocks for any ENG-HANDBOOK.md propagated content
- [ ] `See [ENG-HANDBOOK.md ...]` cross-references for related sections

## After Creating

1. Add entry to `.github/copilot-instructions.md` instruction file table
2. Add file to ENG-HANDBOOK.md Appendix B.4 summary count (if adding to a category)
3. Run `go run ./cmd/cicd-lint lint-docs` to validate cross-references

## References

Read [ENG-HANDBOOK.md Section 2.1.4 Instruction File Organization](../../../docs/ENG-HANDBOOK.md#214-instruction-file-organization) for numbering scheme — use the correct category prefix and next available two-digit number when naming the new file.
Read [ENG-HANDBOOK.md Section 13.4 Documentation Propagation Strategy](../../../docs/ENG-HANDBOOK.md#134-documentation-propagation-strategy) for @source/@propagate system — include correct `@source` markers for any ENG-HANDBOOK.md content copied into the instruction file.
