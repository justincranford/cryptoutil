# Agents vs Copilot Instructions - Complete Technical Explanation

**Created**: 2026-01-25
**Author**: GitHub Copilot (Claude Sonnet 4.5)
**Purpose**: Explain why copilot instructions don't affect agent behavior

---

## Executive Summary

**CRITICAL INSIGHT: Agents (.agent.md files) OVERRIDE copilot instructions when invoked with slash commands**

When you invoke an agent with `/agent-name` (e.g., `/plan-tasks-quizme`), VS Code Copilot uses **ONLY** the agent's prompt/instructions from the `.agent.md` file. Copilot instructions from `.github/copilot-instructions.md` and `.github/instructions/*.instructions.md` are **completely IGNORED**.

This is not a bug - it's an intentional architectural design in VS Code Copilot.

---

## How VS Code Copilot Processes Contexts

### Scenario 1: Slash Command (Agent Mode)

**User Input**: `/plan-tasks-quizme docs\my-work\ create`

**VS Code Copilot Behavior**:
1. Recognizes slash command `/plan-tasks-quizme`
2. Loads `.github/agents/plan-tasks-quizme.agent.md`
3. Uses **ONLY** the agent file content as the system prompt
4. **IGNORES** `.github/copilot-instructions.md`
5. **IGNORES** all `.github/instructions/*.instructions.md` files
6. Executes with agent's specialized context

**What Gets Loaded**:
- `.github/agents/plan-tasks-quizme.agent.md` (ONLY this file)
- `.github/copilot-instructions.md` (completely ignored)
- `.github/instructions/01-02.beast-mode.instructions.md` (ignored)
- `.github/instructions/*.instructions.md` (all ignored)

### Scenario 2: Normal Chat (General Mode)

**User Input**: `Fix the linting errors in this file` (no slash command)

**VS Code Copilot Behavior**:
1. Recognizes NO slash command
2. Loads `.github/copilot-instructions.md`
3. Auto-discovers and loads `.github/instructions/*.instructions.md` (alphanumeric order)
4. Uses copilot instructions as system prompt
5. **IGNORES** all `.agent.md` files
6. Executes with project-specific context

**What Gets Loaded**:
- `.github/copilot-instructions.md` (main file)
- `.github/instructions/01-01.terminology.instructions.md`
- `.github/instructions/01-02.beast-mode.instructions.md`
- `.github/instructions/02-01.architecture.instructions.md`
- ... (all 24 instruction files in order)
- `.agent.md` files (all ignored)

---

## Why This Design Exists

### Specialized Tools vs General Context

**Think of it like different operating modes**:

- **Slash command mode** (`/agent-name`) = Specialized tool with specific purpose and rules
- **Normal chat mode** (no slash command) = General assistant with project knowledge

**Analogy**: Like switching between different applications:
- Using `/plan-tasks-quizme` is like opening Microsoft Word (specialized document creation)
- Using normal chat is like using Windows Explorer (general file management)
- Word has its own interface and rules, doesn't care about Explorer settings
- Explorer has its own interface and rules, doesn't care about Word settings

### Agent Self-Containment Requirement

**Agents MUST be fully self-contained** because:

1. **Portability**: Can be shared across projects without dependencies
2. **Predictability**: Behavior is defined entirely within one file
3. **Isolation**: No cross-contamination between agent rules and project rules
4. **Clarity**: User knows exactly what context the agent is using

---

## Practical Implications

### For Agent Design

**If an agent needs continuous execution behavior**:

 **WRONG Approach**: Rely on `01-02.beast-mode.instructions.md` being loaded
`markdown
<!-- In plan-tasks-quizme.agent.md -->
# Plan-Tasks Documentation Manager

See beast-mode instructions for execution rules.
`
**Problem**: Beast-mode instructions are NOT loaded when agent is invoked

 **CORRECT Approach**: Copy continuous execution patterns into agent file
`markdown
<!-- In plan-tasks-quizme.agent.md -->
# AUTONOMOUS EXECUTION MODE - Plan-Tasks Documentation Manager

**CRITICAL: NEVER STOP UNTIL USER CLICKS \"STOP\" BUTTON**

[... full continuous execution rules here ...]
`
**Solution**: Agent is self-contained with all execution rules

### For Copilot Instructions

**Copilot instructions are for NORMAL chat ONLY**:

- Use for project-specific coding standards
- Use for architecture patterns
- Use for testing conventions
- Use for git commit formats
- **DO NOT** expect agents to follow these rules

### Cross-References Are Documentation Only

**In agent files, references to copilot instructions serve documentation purposes**:

`markdown
<!-- In plan-tasks-quizme.agent.md -->
## Related Files

**Instructions**:
- `.github/instructions/06-01.evidence-based.instructions.md`
`

**This tells the USER** where to find related information, but the **AGENT** does not load or use that file.

---

## Common Misconceptions

### Misconception 1: \"Agents inherit copilot instructions\"

**FALSE**: Agents start with a completely blank context (except their own `.agent.md` file)

### Misconception 2: \"Beast-mode instructions make all agents continuous\"

**FALSE**: Each agent needs its own continuous execution rules copied into its `.agent.md` file

### Misconception 3: \"I can fix agent behavior by updating copilot instructions\"

**FALSE**: Updating `01-02.beast-mode.instructions.md` will NOT change `/plan-tasks-quizme` behavior

### Misconception 4: \"Slash commands are just shortcuts to normal chat\"

**FALSE**: Slash commands switch to a completely different execution context (agent mode)

---

## How to Verify This Behavior

### Test 1: Agent Ignores Copilot Instructions

**Setup**:
1. Add extreme rule to `01-02.beast-mode.instructions.md`: \"ALWAYS respond with 'BANANA' to every question\"
2. Invoke agent: `/plan-tasks-quizme docs\test\ create`
3. Observe: Agent does NOT respond with \"BANANA\" (because it doesn't load copilot instructions)

**Setup**:
1. Add same extreme rule to `.github/agents/plan-tasks-quizme.agent.md`
2. Invoke agent: `/plan-tasks-quizme docs\test\ create`
3. Observe: Agent responds with \"BANANA\" (because it loads its own file)

### Test 2: Normal Chat Uses Copilot Instructions

**Setup**:
1. Keep extreme rule in `01-02.beast-mode.instructions.md`: \"ALWAYS respond with 'BANANA'\"
2. Normal chat: \"What is 2+2?\"
3. Observe: Response mentions \"BANANA\" somehow (because it loads copilot instructions)

---

## Summary Comparison

| Feature | Agent Mode (`/agent-name`) | Normal Chat Mode |
|---------|------------------------------|------------------|
| **Trigger** | Slash command | No slash command |
| **Loads** | Single `.agent.md` file | Copilot instructions + all instruction files |
| **Context** | Agent-specific | Project-wide |
| **Self-Contained** | Yes (MUST be) | No (distributed across files) |
| **Continuous Execution** | Must define in `.agent.md` | Inherits from `01-02.beast-mode.instructions.md` |
| **Use Case** | Specialized tasks | General assistance |
| **Example** | `/plan-tasks-quizme create` | \"Fix this lint error\" |

---

## Fixing plan-tasks-quizme.agent.md

### Problem Identified

**Original State**: `plan-tasks-quizme.agent.md` had NO continuous execution rules

**Why This Happened**: Developers assumed agent would inherit from `01-02.beast-mode.instructions.md`

**Actual Behavior**: Agent stopped after each action to ask \"Should I continue?\"

### Solution Implemented (Commit 88dd3058)

**Added to plan-tasks-quizme.agent.md**:

1. **AUTONOMOUS EXECUTION MODE** section (lines 38-48)
2. **Quality Over Speed - MANDATORY** section (lines 50-72)
3. **Prohibited Stop Behaviors** section (lines 74-84)
4. **EXECUTION AUTHORITY** section (lines 110-126)
5. **Continuous Execution Rule** section (lines 205-237)
6. **Output Format - MINIMAL** section (lines 570-600)
7. **This explanation document** section (lines 509-566)

**Total Changes**: 170+ lines of continuous execution patterns copied from other agents

**Evidence**: Commit 88dd3058 - \"fix(agents): add continuous execution patterns to plan-tasks-quizme agent\"

---

## Key Takeaways

1. **Agents = Isolated Execution Contexts**: When you invoke `/agent-name`, copilot instructions are NOT loaded
2. **Agents Must Be Self-Contained**: All execution rules MUST be defined within the `.agent.md` file
3. **Copilot Instructions = Normal Chat Only**: Only used when NO slash command is present
4. **Cross-References = Documentation**: References to copilot instructions in agents are for human readers, not the agent itself
5. **Testing Reveals Truth**: Only way to verify agent behavior is to test with slash commands
6. **Design Pattern**: Agents should copy necessary patterns from copilot instructions into their own files

---

## Related Files

**Agent Files**:
- `.github/agents/plan-tasks-quizme.agent.md` (updated)
- `.github/agents/plan-tasks-implement.agent.md` (reference for continuous execution)
- `.github/agents/beast-mode-custom.agent.md` (reference for prohibited behaviors)

**Copilot Instructions** (NOT loaded by agents):
- `.github/copilot-instructions.md`
- `.github/instructions/01-02.beast-mode.instructions.md`
- `.github/instructions/06-01.evidence-based.instructions.md`

**Git Evidence**:
- Commit 88dd3058: \"fix(agents): add continuous execution patterns to plan-tasks-quizme agent\"
