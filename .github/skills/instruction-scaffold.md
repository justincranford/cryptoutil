# instruction-scaffold

Create a conformant `.github/instructions/NN-NN.name.instructions.md` file.

## Purpose

Use when adding a new instruction file to the project. Ensures correct YAML
frontmatter, numbering scheme, and @source blocks for propagated content.

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

<!-- @source from="docs/ARCHITECTURE.md" as="chunk-id" -->
Content verbatim from ARCHITECTURE.md
<!-- @/source -->

See [ARCHITECTURE.md Section X.Y](../../docs/ARCHITECTURE.md#xy-anchor) for complete documentation.
```

## Mandatory Checklist

- [ ] YAML frontmatter with `description` and `applyTo`
- [ ] Numbered correctly (no gaps, no conflicts)
- [ ] Referenced in `.github/copilot-instructions.md` instruction file table
- [ ] `@source` blocks for any ARCHITECTURE.md propagated content
- [ ] `See [ARCHITECTURE.md ...]` cross-references for related sections

## After Creating

1. Add entry to `.github/copilot-instructions.md` instruction file table
2. Add file to ARCHITECTURE.md Appendix B.4 summary count (if adding to a category)
3. Run `go run ./cmd/cicd lint-docs` to validate cross-references

## References

See [ARCHITECTURE.md Section 2.1.4 Instruction File Organization](../../docs/ARCHITECTURE.md#214-instruction-file-organization) for numbering scheme.
See [ARCHITECTURE.md Section 12.7 Documentation Propagation Strategy](../../docs/ARCHITECTURE.md#127-documentation-propagation-strategy) for @source/@propagate system.
