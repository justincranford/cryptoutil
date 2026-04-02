---
name: agent-scaffold
description: "Create a conformant .claude/agents/NAME.md with all mandatory sections. Use when adding a new agent to ensure correct YAML frontmatter, autonomous execution mode, quality gates, and ARCHITECTURE.md self-containment references. Both VS Code Copilot and Claude Code read .claude/agents/ natively."
argument-hint: "[agent-name]"
disable-model-invocation: true
---

Create a conformant `.claude/agents/NAME.md` with all mandatory sections.

## Purpose

Use when creating a new agent for VS Code Copilot or Claude Code. Ensures correct YAML frontmatter, mandatory
sections, and ARCHITECTURE.md references for agent self-containment.

## Template

```markdown
---
name: agent-name
description: One-line description of what this agent does
argument-hint: "<required-argument>"
---

# Agent Title

## Purpose

What this agent does and when to invoke it.

## AUTONOMOUS EXECUTION MODE

This agent executes autonomously. Do NOT ask clarifying questions, pause for confirmation, or request user input.

## Maximum Quality Strategy - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

## Prohibited Stop Behaviors - ALL FORBIDDEN

- Status summaries, "session complete" messages, "next steps" proposals
- Asking permission ("Should I continue?", "Shall I proceed?")
- Pauses between tasks, celebrations, premature completion claims
- Leaving uncommitted changes, stopping after analysis

## Continuous Execution Rule - MANDATORY

Task complete → Commit → IMMEDIATELY start next task (zero pause, zero text to user).

## Quality Gates (Per Task)

Before marking complete: Build clean → Lint clean → Tests pass → Coverage maintained.

Read [ARCHITECTURE.md Section 11.2 Quality Gates](../../../docs/ARCHITECTURE.md#112-quality-gates) for mandatory quality gate requirements — apply all pre-commit quality gate commands from this section before marking any task complete.

## Mandatory Review Passes

**MANDATORY: Minimum 3, maximum 5 review passes before marking any task complete.**

Read [ARCHITECTURE.md Section 2.5 Quality Strategy](../../../docs/ARCHITECTURE.md#25-quality-strategy) for mandatory review pass requirements — perform minimum 3, maximum 5 passes checking all 8 quality attributes before marking complete.
```

## Mandatory Checklist

- [ ] YAML frontmatter with `name`, `description`
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

## After Creating

1. Add entry to ARCHITECTURE.md Section 2.1.2 Agent Catalog table
2. Run `go run ./cmd/cicd-lint lint-docs` to validate cross-references

## References

**When generating a continuous-execution agent**: the generated file MUST contain `## Maximum Quality Strategy - MANDATORY`, `## Prohibited Stop Behaviors - ALL FORBIDDEN`, and `## Continuous Execution Rule - MANDATORY` sections with their full content — NOT just links. Agents do NOT load instruction files, so all required context must be present verbatim.

Read [ARCHITECTURE.md Section 2.1.1 Agent Architecture](../../../docs/ARCHITECTURE.md#211-agent-architecture) for the agent self-containment checklist — check that all required ARCHITECTURE.md sections are referenced in the generated file.

Read [.github/instructions/06-02.agent-format.instructions.md](../../instructions/06-02.agent-format.instructions.md) for format requirements and the complete list of mandatory sections.
