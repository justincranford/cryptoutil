---
name: propagation-check
description: "Detect @propagate/@source drift between ARCHITECTURE.md and instruction files, and generate corrected @source block content. Use before committing instruction file changes to ensure lint-docs passes and verbatim doc chunks stay synchronized."
argument-hint: "[instruction file or omit for full project check]"
---

Detect @propagate/@source drift and generate corrected @source block content.

## Purpose

Use when ARCHITECTURE.md sections have changed and you need to update downstream
`@source` blocks in instruction files or agents. Prevents copy-paste errors.

## Marker System

**Source (ARCHITECTURE.md)**:
```html
<!-- @propagate to=".github/instructions/FILE.md" as="chunk-id" -->
content here
<!-- @/propagate -->
```

**Target (instruction file OR agent file)**:
```html
<!-- @source from="docs/ARCHITECTURE.md" as="chunk-id" -->
content here (MUST be byte-for-byte identical)
<!-- @/source -->
```

> Note: Both `.github/instructions/*.instructions.md` files AND `.github/agents/*.agent.md` files can contain `@source` blocks. Agents do not inherit instruction files, so propagated content must be embedded directly in the agent file.

## Checking for Drift

```bash
# Run the automated validator
go run ./cmd/cicd-lint lint-docs

# Manual: extract @propagate block content
python3 - <<'EOF'
import re
with open('docs/ARCHITECTURE.md') as f: content = f.read()
# Find all propagate blocks
for m in re.finditer(r'<!-- @propagate to="([^"]+)" as="([^"]+)" -->(.*?)<!-- @/propagate -->', content, re.DOTALL):
    print(f"Target: {m.group(1)}, ID: {m.group(2)}")
    print(f"Content: {m.group(3)[:100]}...")
    print()
EOF
```

## Fix Workflow

1. Find the @propagate block in ARCHITECTURE.md
2. Copy its content verbatim
3. Paste between @source markers in the target file
4. Run `go run ./cmd/cicd-lint lint-docs` to verify match

## Rules

- Content between markers MUST be identical (byte-for-byte after whitespace normalization)
- Headings NEVER inside markers (put outside as section headings)
- No `See [ARCHITECTURE.md ...]` links inside markers (put outside as glue)
- Changes to ARCHITECTURE.md MUST propagate in the SAME commit

## References

Read [ARCHITECTURE.md Section 12.7 Documentation Propagation Strategy](../../../docs/ARCHITECTURE.md#127-documentation-propagation-strategy) for full marker system documentation — apply all marker system rules (byte-for-byte match, no headings inside markers, same-commit propagation) when checking and fixing drift.

See the `doc-sync` agent (`.github/agents/doc-sync.agent.md`) for orchestrating full documentation synchronization across ARCHITECTURE.md, instruction files, and agent files.
