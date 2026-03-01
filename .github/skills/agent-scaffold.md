# agent-scaffold

Create a conformant `.github/agents/NAME.agent.md` with all mandatory sections.

## Purpose

Use when creating a new Copilot agent. Ensures correct YAML frontmatter, mandatory
sections, and ARCHITECTURE.md references for agent self-containment.

## Template

```markdown
---
name: agent-name
description: One-line description of what this agent does
tools:
  - edit/editFiles
  - execute/runInTerminal
  - execute/getTerminalOutput
  - read/problems
  - search/codebase
  - search/usages
  - search/changes
handoffs:
  - next-agent-name
argument-hint: "<required-argument>"
---

# Agent Title

## Purpose

What this agent does and when to invoke it.

## AUTONOMOUS EXECUTION MODE

This agent executes autonomously. See [ARCHITECTURE.md Section 2.4](../../docs/ARCHITECTURE.md#24-implementation-strategy) for continuous execution patterns.

## Quality Gates (Per Task)

Before marking complete: Build clean → Lint clean → Tests pass → Coverage maintained.

See [ARCHITECTURE.md Section 11.2 Quality Gates](../../docs/ARCHITECTURE.md#112-quality-gates) for mandatory quality gate requirements.

## Mandatory Review Passes

**MANDATORY: Minimum 3, maximum 5 review passes before marking any task complete.**

See [ARCHITECTURE.md Section 2.5 Quality Strategy](../../docs/ARCHITECTURE.md#25-quality-strategy) for mandatory review pass requirements.
```

## Mandatory Checklist

- [ ] YAML frontmatter with `name`, `description`, `tools`
- [ ] References to ARCHITECTURE.md (self-contained, agents don't load instructions)
- [ ] Section for Quality Gates with ARCHITECTURE.md cross-reference
- [ ] Section for Mandatory Review Passes (min 3, max 5)
- [ ] `argument-hint` if agent takes an argument

## Agent Self-Containment Rules

Agents do NOT inherit `.github/copilot-instructions.md` or `*.instructions.md`.
ALL relevant context MUST be in the agent file itself.

**Required ARCHITECTURE.md references** for code-modifying agents:
- Section 10 (Testing Architecture)
- Section 11 (Code Quality Standards)
- Section 13 (Development Practices)
- Section 2.5 (Quality Strategy — for coverage/mutation targets)

## References

See [ARCHITECTURE.md Section 2.1.1 Agent Architecture](../../docs/ARCHITECTURE.md#211-agent-architecture) for agent self-containment checklist.
See [docs/ARCHITECTURE.md Section 06-02.agent-format](../../.github/instructions/06-02.agent-format.instructions.md) for format requirements.
