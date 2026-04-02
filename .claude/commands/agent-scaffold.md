Create a new `.claude/agents/NAME.md` file with all mandatory sections.

**Full Copilot original**: [.github/skills/agent-scaffold/SKILL.md](.github/skills/agent-scaffold/SKILL.md)

**Note**: `.claude/agents/` is the single canonical source — VS Code Copilot and Claude Code both natively read this directory.

Provide: agent name (e.g., `security-audit`), description, purpose.

## YAML Frontmatter (Required)

```yaml
---
name: {agent-name}
description: >
  {One paragraph: what this agent does, when to use it, what it produces.
  Be specific — this is used for auto-selection.}
argument-hint: "[optional hint shown in chat input]"
---
```

## Mandatory Sections

Every agent MUST include:

1. **Autonomous Execution Mode** — continuous work, no interruptions
2. **Maximum Quality Strategy** (ARCHITECTURE.md §11) — 8 quality attributes
3. **Prohibited Stop Behaviors** — explicit list of what NOT to do
4. **Pre-Flight Checks** — verification before starting
5. **Completion Criteria** — what "done" means for this agent
6. **ARCHITECTURE.md References** — which sections apply (MUST have ≥1)

## Agent Self-Containment Checklist (MANDATORY)

Per ARCHITECTURE.md §2.1.1, agents do NOT inherit instruction files. Each agent MUST explicitly reference:

- Coding standards: ARCHITECTURE.md §14 (if modifying code)
- Testing standards: ARCHITECTURE.md §10 (if writing tests)
- Quality gates: ARCHITECTURE.md §11 (always)
- Deployment: ARCHITECTURE.md §12, §13 (if touching deployments)
- Documentation propagation: ARCHITECTURE.md §13.4 (if touching docs/skills/instructions)

## Tool Discovery

```python
# Scan for extension tools
import json, pathlib
ext_dir = pathlib.Path.home() / ".vscode" / "extensions"
for d in sorted(ext_dir.iterdir()):
    pkg = d / "package.json"
    if pkg.is_file():
        data = json.loads(pkg.read_text(encoding="utf-8"))
        tools = data.get("contributes", {}).get("languageModelTools")
        if tools:
            for t in tools:
                print(f"  {t.get('toolReferenceName', t.get('name', ''))}")
```

## Claude Code + VS Code Compatibility

Both VS Code Copilot and Claude Code natively read `.claude/agents/*.md`. No separate bridge file is needed.
