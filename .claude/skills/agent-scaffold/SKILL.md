---
name: agent-scaffold
description: "Create both .github/agents/NAME.agent.md (Copilot canonical, with tools whitelist) and .claude/agents/NAME.md (Claude Code canonical, without tools). Use when adding a new agent to ensure both files have correct YAML frontmatter, autonomous execution mode, quality gates, and ENG-HANDBOOK.md self-containment references."
argument-hint: "[agent-name]"
---

Create both `.github/agents/NAME.agent.md` (Copilot) and `.claude/agents/NAME.md` (Claude Code) with all mandatory sections.

## Purpose

Use when creating a new agent. Creates BOTH canonical files so the agent works
correctly in both VS Code Copilot (with tool whitelist) and Claude Code (inherits all tools).

## Key Rules

- ALWAYS create both files: `.github/agents/NAME.agent.md` (Copilot) AND `.claude/agents/NAME.md` (Claude Code)
- `tools:` field REQUIRED in Copilot file (whitelist); OMIT in Claude file (inherits all)
- Body content MUST be identical between both files; only frontmatter differs
- `name:` prefix: `copilot-NAME` in Copilot file, `claude-NAME` in Claude file
- MUST include ENG-HANDBOOK.md self-containment references (≥1 section reference)
- MUST include Autonomous Execution Mode and Prohibited Stop Behaviors sections

## Copilot Template (`.github/agents/NAME.agent.md`)

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

Read [ENG-HANDBOOK.md Section 11.2 Quality Gates](../../../docs/ENG-HANDBOOK.md#112-quality-gates) for mandatory quality gate requirements — apply all pre-commit quality gate commands from this section before marking any task complete.

## Mandatory Review Passes

**MANDATORY: Minimum 3, maximum 5 review passes before marking any task complete.**

Read [ENG-HANDBOOK.md Section 2.5 Quality Strategy](../../../docs/ENG-HANDBOOK.md#25-quality-strategy) for mandatory review pass requirements — perform minimum 3, maximum 5 passes checking all 8 quality attributes before marking complete.
```

## Claude Code Template (`.claude/agents/NAME.md`)

Identical body to the Copilot file. Only the frontmatter differs — omit all Copilot-only fields:

```markdown
---
name: agent-name
description: One-line description of what this agent does
argument-hint: "<required-argument>"
---

[same body as .github/agents/NAME.agent.md]
```

**Never add `tools:` to the Claude file** — Claude inherits all tools when the field is absent. Adding it with an empty list would restrict access.

## Mandatory Checklist

- [ ] Copilot file created: `.github/agents/NAME.agent.md` with `name`, `description`, `tools` (whitelist)
- [ ] Claude file created: `.claude/agents/NAME.md` with `name`, `description` only (no `tools:`)
- [ ] Both files have identical body content
- [ ] References to ENG-HANDBOOK.md (self-contained, agents don't load instructions)
- [ ] Section for Quality Gates with ENG-HANDBOOK.md cross-reference
- [ ] Section for Mandatory Review Passes (min 3, max 5)
- [ ] `argument-hint` if agent takes an argument

## Agent Self-Containment Rules

Agents do NOT inherit `.github/copilot-instructions.md` or `*.instructions.md`.
ALL relevant context MUST be in the agent file itself.

**Required ENG-HANDBOOK.md references** for code-modifying agents:
- Section 10 (Testing Architecture)
- Section 11 (Code Quality Standards)
- Section 13 (Development Practices)
- Section 2.5 (Quality Strategy — for coverage/mutation targets)

## After Creating

1. Add entry to ENG-HANDBOOK.md Section 2.1.2 Agent Catalog table
2. Add entries to CLAUDE.md Agents table (link to `.claude/agents/NAME.md`)
3. Run `go run ./cmd/cicd-lint lint-docs` to validate cross-references

## References

**When generating a continuous-execution agent**: the generated file MUST contain `## Maximum Quality Strategy - MANDATORY`, `## Prohibited Stop Behaviors - ALL FORBIDDEN`, and `## Continuous Execution Rule - MANDATORY` sections with their full content — NOT just links. Agents do NOT load instruction files, so all required context must be present verbatim.

Read [ENG-HANDBOOK.md Section 2.1.1 Agent Architecture](../../../docs/ENG-HANDBOOK.md#211-agent-architecture) for the agent self-containment checklist — check that all required ENG-HANDBOOK.md sections are referenced in the generated file.

Read [.github/instructions/06-02.agent-format.instructions.md](../../instructions/06-02.agent-format.instructions.md) for format requirements and the complete list of mandatory sections.
