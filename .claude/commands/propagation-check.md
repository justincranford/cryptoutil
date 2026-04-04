---
name: propagation-check
description: "Detect @propagate/@source drift between ARCHITECTURE.md and instruction files, and generate corrected @source block content. Use before committing instruction file changes to ensure lint-docs passes and verbatim doc chunks stay synchronized."
argument-hint: "[instruction file or omit for full project check]"
---

Detect @propagate/@source drift between docs/ARCHITECTURE.md and .github/instructions/ files.

**Full Copilot original**: [.github/skills/propagation-check/SKILL.md](.github/skills/propagation-check/SKILL.md)

## Automated Check

```bash
go run cmd/cicd-lint/main.go lint-docs
```

This validates all `@propagate`/`@source` pairs are byte-for-byte identical.

## Marker System

**In ARCHITECTURE.md** (source):
```markdown
<!-- @propagate to=".github/instructions/XX-XX.name.instructions.md" as="block-id" -->
Content that must be copied verbatim...
<!-- @/propagate -->
```

**In instruction files** (targets):
```markdown
<!-- @source from="docs/ARCHITECTURE.md" as="block-id" -->
Content that must be copied verbatim...
<!-- @/source -->
```

## Key Rules

1. Content inside `@source` blocks MUST be byte-for-byte identical to the corresponding `@propagate` block
2. ARCHITECTURE.md changes to a propagated block MUST be propagated to ALL target files in the same commit
3. Section headings, `See` cross-references, and transitions are NOT inside markers — they differ between files
4. No markdown links `[text](url)` inside `@source` blocks (links are in the non-propagated glue)
5. Never manually edit `@source` blocks — always update ARCHITECTURE.md first, then propagate

## Manual Fix Workflow

1. Run `go run cmd/cicd-lint/main.go lint-docs` to find drifted blocks
2. For each drifted block, copy the ARCHITECTURE.md content verbatim into the instruction file's `@source` block
3. Re-run lint-docs to confirm zero drift
4. Commit ARCHITECTURE.md and all affected instruction files together

## Coverage Accounting

The propagation system tracks what percentage of ARCHITECTURE.md content is covered by `@propagate` blocks. Run `go run cmd/cicd-lint/main.go lint-docs lint-coverage` to check.
