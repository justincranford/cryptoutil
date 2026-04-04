---
name: agent-scaffold
description: "Create both .github/agents/NAME.agent.md (Copilot canonical, with tools whitelist) and .claude/agents/NAME.md (Claude Code canonical, without tools). Use when adding a new agent to ensure both files have correct YAML frontmatter, autonomous execution mode, quality gates, and ARCHITECTURE.md self-containment references."
argument-hint: "[agent-name]"
---

Create BOTH canonical agent files: `.github/agents/NAME.agent.md` (Copilot) and `.claude/agents/NAME.md` (Claude Code).

**Full skill**: [.github/skills/agent-scaffold/SKILL.md](.github/skills/agent-scaffold/SKILL.md)

Provide: agent name (e.g., `security-audit`), description, purpose.

## Key Rules

- ALWAYS create both files: `.github/agents/NAME.agent.md` (Copilot) AND `.claude/agents/NAME.md` (Claude Code)
- `tools:` field REQUIRED in Copilot file (whitelist); OMIT in Claude file (inherits all)
- Body content MUST be identical between both files; only frontmatter differs
- `name:` prefix: `copilot-NAME` in Copilot file, `claude-NAME` in Claude file
- MUST include ARCHITECTURE.md self-containment references (≥1 section reference)
- MUST include Autonomous Execution Mode and Prohibited Stop Behaviors sections

## VS Code + Claude Code Compatibility

Two files are required. Copilot treats `tools:` as a **whitelist** — omitting it restricts tool access. Claude Code treats absent `tools:` as **inherit all**. A single file cannot satisfy both.

## Copilot Frontmatter (`.github/agents/NAME.agent.md`)

```yaml
---
name: {agent-name}
description: >
  {One paragraph: what this agent does, when to use it, what it produces.
  Be specific — this is used for auto-selection.}
tools:
  - category/toolName  # See ARCHITECTURE.md §2.1.6 for tool discovery
handoffs:
  - agent: implementation-execution
    trigger: "When plan is approved and ready for execution"
argument-hint: "[optional hint shown in chat input]"
---
```

## Claude Code Frontmatter (`.claude/agents/NAME.md`)

```yaml
---
name: {agent-name}
description: >
  {Same description as Copilot file.}
argument-hint: "[optional hint shown in chat input]"
---
```

**Never add `tools:` to the Claude file.** Body content must be identical to the Copilot file.

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

## Claude Code Bridge

After creating `.github/agents/NAME.agent.md`, also create `.claude/agents/NAME.md` with the same content adapted for Claude Code's agent format (YAML frontmatter with `description:` field).
