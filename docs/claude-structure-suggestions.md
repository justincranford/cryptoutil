# claude-structure.md → ENG-HANDBOOK.md Suggestions

## Executive Summary

Analysis of [claude-structure.md](claude-structure.md) against [ENG-HANDBOOK.md §2.1](ENG-HANDBOOK.md#21-agent-orchestration-strategy) and [§14.11](ENG-HANDBOOK.md#1411-claude-code-autonomous-execution) reveals that the handbook covers the dual canonical agent/skill strategy, the three execution modes, and the skill catalogue, but is missing the `.claude/` directory structure reference, CLAUDE.md format details, complete skill frontmatter fields, dynamic context injection syntax, sub-agent format details, path-scoped rules, and the agentskills.io open standard context. The following additions are suggested.

1. [.claude/ Directory Structure Reference](#1-claude-directory-structure-reference) — the tree of all `.claude/` subdirectories (agents/, skills/, rules/, settings.json, agent-memory/, worktrees/) is not documented.
2. [User-Level ~/.claude/ Structure](#2-user-level-claude-structure) — the user-global Claude configuration directory layout is absent.
3. [CLAUDE.md Format and Loading Behavior](#3-claudemd-format-and-loading-behavior) — delivery as user message (not system prompt), /compact survival, multiple-file concatenation, and user-level-first loading are not documented.
4. [Required CLAUDE.md Sections for cryptoutil](#4-required-claudemd-sections-for-cryptoutil) — the canonical section structure for this project's CLAUDE.md is not specified.
5. [Complete Skill Frontmatter Fields](#5-complete-skill-frontmatter-fields) — many SKILL.md frontmatter fields (`allowed-tools`, `model`, `effort`, `context`, `agent`, `paths`, `shell`, `user-invocable`) are not documented in the handbook.
6. [Dynamic Context Injection Syntax](#6-dynamic-context-injection-syntax) — the backtick-bang inline command substitution and all string substitution variables (`$ARGUMENTS`, `$0/$N`, `${CLAUDE_SESSION_ID}`, `${CLAUDE_SKILL_DIR}`) are absent.
7. [Skill Body Structure Template](#7-skill-body-structure-template) — the recommended markdown structure for SKILL.md bodies (Key Rules section, Template/Workflow section) is not documented.
8. [Sub-Agent Frontmatter Fields](#8-sub-agent-frontmatter-fields) — `disallowedTools`, `permissionMode`, `maxTurns`, `skills`, `memory`, and `color` fields are not in the handbook.
9. [Path-Scoped Rules (.claude/rules/)](#9-path-scoped-rules-clauderules) — the `.claude/rules/` directory, its auto-load behavior, the `paths` frontmatter key, and recommended cryptoutil rule files are not documented.
10. [agentskills.io Open Standard Context](#10-agentskillsio-open-standard-context) — the cross-agent shared frontmatter requirements and multi-tool adoption context are absent.
11. [CLAUDE.md Length and Scoping Strategy](#11-claudemd-length-and-scoping-strategy) — recommendations for large monorepos (per-directory CLAUDE.md files, lazy loading) and specific rule file suggestions for cryptoutil are not documented.

---

## Details

### 1. .claude/ Directory Structure Reference

**Current state in ENG-HANDBOOK.md**: §14.11 mentions `settings.local.json` and references agents/skills directories but provides no canonical tree of the full `.claude/` structure.

**Suggested addition to §14.11 or §2.1**:

```
.claude/                        # Claude Code project configuration
├── CLAUDE.md                   # Project instructions (loaded every session)
│                               # Target: <200 lines. Longer = reduced adherence.
│                               # Block HTML comments <!-- ... --> are stripped.
│                               # @path imports inline-expand (max 5 hops).
├── agents/                     # Custom sub-agent definitions
│   └── <name>.md               # Agent file (YAML frontmatter + system prompt)
├── skills/                     # Custom slash commands (directory-based format)
│   └── <name>/                 # Each skill is a DIRECTORY
│       ├── SKILL.md            # Required entrypoint
│       ├── references/         # Optional: detailed docs loaded on demand
│       ├── scripts/            # Optional: executable code
│       └── assets/             # Optional: templates, resources
├── rules/                      # Path-scoped project rules (auto-loaded by file match)
│   └── *.md                    # Rule files with optional `paths:` frontmatter
├── settings.json               # Project settings (team-level, committed to git)
├── settings.local.json         # Local settings (personal, gitignored)
├── agent-memory/               # Persistent memory for project-scoped agents
└── worktrees/                  # Isolated git worktrees (--worktree flag)
```

Note: The legacy `.claude/commands/` directory has been removed — all commands migrated to `.claude/skills/<name>/SKILL.md`.

---

### 2. User-Level ~/.claude/ Structure

**Current state in ENG-HANDBOOK.md**: Not described.

**Suggested addition to §14.11**:

User-level Claude configuration at `~/.claude/` applies across all projects:

```
~/.claude/
├── CLAUDE.md               # User-global instructions (loaded before project CLAUDE.md)
├── agents/                 # User-global custom agents
├── skills/                 # User-global custom skills
├── rules/                  # User-global path-scoped rules
└── projects/<proj>/        # Per-project memory store
    └── memory/             # Persistent memory files for the project
```

Loading order: user-level CLAUDE.md loads before project-level CLAUDE.md. Multiple files are concatenated, not overriding.

---

### 3. CLAUDE.md Format and Loading Behavior

**Current state in ENG-HANDBOOK.md**: §14.11 mentions CLAUDE.md exists and is updated when adding agents/skills. The delivery mechanism and survival behaviors are not documented.

**Suggested addition to §14.11**:

> **CLAUDE.md format**: Plain Markdown. No YAML frontmatter.
>
> **Delivery**: Content is delivered as a **user message** after the system prompt — not part of the system prompt itself. This means it is subject to context window limits and compaction.
>
> **Import syntax**: Use `@path/to/file` to inline-expand other files (max 5 hops deep). Block HTML comments `<!-- text -->` are stripped and do not consume context window space (useful for maintainer notes).
>
> **Survival**: CLAUDE.md survives `/compact` — Claude re-reads it from disk after compaction. Always keep it under 200 lines per file; longer files reduce instruction adherence.
>
> **Multiple files**: User-level `~/.claude/CLAUDE.md` loads before project-level CLAUDE.md. Files are concatenated, not overriding — both apply simultaneously.

---

### 4. Required CLAUDE.md Sections for cryptoutil

**Current state in ENG-HANDBOOK.md**: No canonical CLAUDE.md structure is specified for this project.

**Suggested addition to §14.11**:

Required sections for cryptoutil CLAUDE.md:

```markdown
# {Project Name} — Claude Code Instructions

## Architecture Source of Truth
(Links to ENG-HANDBOOK.md and key section index)

## Instruction Files
@.github/instructions/*.instructions.md references

## Agents
(Table: agent name -> when to use)

## Skills (Slash Commands)
(Table: skill name -> purpose)
```

---

### 5. Complete Skill Frontmatter Fields

**Current state in ENG-HANDBOOK.md**: §2.1.5 documents `name`, `description`, `argument-hint`, `user-invocable`, and `disable-model-invocation`. Several additional Claude Code-specific fields are not documented.

**Suggested addition to §2.1.5**:

Complete SKILL.md frontmatter for Claude Code skills:

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Max 64 chars, lowercase letters/numbers/hyphens, should match directory name |
| `description` | Yes | Max 1024 chars; specific about capabilities and when to invoke |
| `argument-hint` | No | Hint shown in chat autocomplete (e.g., `"[package-name]"`) |
| `disable-model-invocation` | No | Copilot-only — NEVER use in Claude skills |
| `user-invocable` | No | If false: hidden from `/` menu but Claude can auto-invoke. Default: true |
| `allowed-tools` | No | Tools allowed without per-use approval when skill is active (e.g., `Read Grep Glob`) |
| `model` | No | Model override when skill is active (`sonnet`, `opus`, `haiku`) |
| `effort` | No | `low`, `medium`, `high`, `max` (Opus 4.6+ only) |
| `context` | No | `fork` runs skill in isolated subagent with no conversation history |
| `agent` | No | Subagent type to use with `context: fork` (e.g., `Explore`) |
| `paths` | No | Load skill automatically only when working with matching files (e.g., `"internal/**/*.go"`) |
| `shell` | No | `bash` (default) or `powershell` for inline command execution |

---

### 6. Dynamic Context Injection Syntax

**Current state in ENG-HANDBOOK.md**: Not documented anywhere.

**Suggested addition to §2.1.5**:

Inline command execution in SKILL.md bodies:

```
# Current git status:
```!
git log --oneline -5
git status --short
```
```

The backtick-bang block executes before Claude sees the prompt; output replaces the placeholder.

String substitutions available in skill bodies:

| Substitution | Meaning |
|---|---|
| `$ARGUMENTS` | All arguments passed after the skill name |
| `$0`, `$1`, `$N` | Positional arguments (0-based) |
| `${CLAUDE_SESSION_ID}` | Current Claude Code session ID |
| `${CLAUDE_SKILL_DIR}` | Absolute path to the skill's directory |

---

### 7. Skill Body Structure Template

**Current state in ENG-HANDBOOK.md**: §2.1.5 lists the required `## Key Rules` section but does not specify the recommended overall body structure.

**Suggested addition to §2.1.5**:

Recommended SKILL.md body structure for cryptoutil skills:

```markdown
---
(frontmatter)
---

(One-paragraph summary of the skill's purpose and when to use it)

## Key Rules

- Brief rule 1 (mandatory; italicize exceptions)
- Brief rule 2
(6-12 rules maximum; enforce the most common failure modes)

## Template / Workflow

(Step-by-step instructions or code template the skill produces)

**Full reference**: [.github/skills/<NAME>/SKILL.md](.github/skills/<NAME>/SKILL.md)
```

The `## Key Rules` section is MANDATORY and enforced by `lint-skill-command-drift` in both Copilot and Claude skill files.

---

### 8. Sub-Agent Frontmatter Fields

**Current state in ENG-HANDBOOK.md**: §2.1.1 lists the documented fields (`name`, `description`, `argument-hint`). Several Claude Code-specific agent fields are absent.

**Suggested addition to §2.1.1**:

Complete Claude Code agent frontmatter (`.claude/agents/<name>.md`):

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique ID — lowercase letters and hyphens. Prefix: `claude-NAME` |
| `description` | Yes | When to delegate to this agent (Claude reads to decide) |
| `tools` | No | OMIT in Claude agents — Claude inherits all tools. Specifying restricts access. |
| `disallowedTools` | No | Explicitly deny tools from the inherited set (e.g., `WebSearch`) |
| `model` | No | `sonnet`, `opus`, `haiku`, or `inherit` (default) |
| `permissionMode` | No | `default`, `acceptEdits`, `auto`, `dontAsk`, `bypassPermissions`, `plan` |
| `maxTurns` | No | Maximum agentic turns before stopping (default: unlimited) |
| `skills` | No | Skills preloaded into agent context at startup (e.g., `test-table-driven`) |
| `memory` | No | Persistent memory scope: `user`, `project`, or `local` |
| `color` | No | Agent color in UI |
| `argument-hint` | No | Expected arguments — include for documentation purposes |

Key behavior: Subagents receive ONLY their system prompt + basic environment details. They do NOT inherit the parent conversation history, Claude Code system prompt, or parent skills (unless listed in the `skills` field). Agents MUST be self-contained — see the Agent Self-Containment Checklist in §2.1.1.

---

### 9. Path-Scoped Rules (.claude/rules/)

**Current state in ENG-HANDBOOK.md**: Not documented.

**Suggested addition to §14.11 or §2.1**:

Path-scoped rules in `.claude/rules/` auto-load based on which files Claude is editing. Useful for per-directory coding standards without polluting the global CLAUDE.md.

Example rule file:

```markdown
---
paths:
  - "internal/apps-framework/**/*.go"
  - "api/**/*.yaml"
---

# Framework Rules

When working in the framework package, always:
- Use function-parameter injection for seams (not package-level vars)
- Use testdb.NewInMemorySQLiteDB(t) for unit tests
```

Load behavior:
- **With `paths` frontmatter**: loaded lazily only when Claude works with matching files.
- **Without `paths` frontmatter**: loaded at session launch (same priority as `.claude/CLAUDE.md`).

Recommended rule files for cryptoutil:

| File | Paths Filter | Purpose |
|------|-------------|---------|
| `.claude/rules/framework.md` | `internal/apps-framework/**/*.go` | Framework seam injection and test helper requirements |
| `.claude/rules/tests.md` | `**/*_test.go` | t.Parallel, table-driven, UUIDv7, DisableKeepAlives |

---

### 10. agentskills.io Open Standard Context

**Current state in ENG-HANDBOOK.md**: §2.1.5 notes that Claude Code skills use the directory-based SKILL.md format but does not mention the open standard provenance or cross-tool adoption.

**Suggested addition to §2.1.5**:

Agent Skills open standard (agentskills.io): The `.claude/skills/<name>/SKILL.md` directory format is based on this open standard, developed by Anthropic and adopted by Gemini CLI, GitHub Copilot, OpenAI Codex, Amp, Kiro, Qodo, and VS Code.

Shared frontmatter fields required for cross-agent compatibility:

| Field | Constraint |
|-------|-----------|
| `name` | Max 64 chars, lowercase letters/numbers/hyphens, MUST match directory name |
| `description` | Max 1024 chars |

Because both Copilot and Claude now use the same SKILL.md format, the only expected differences between `.github/skills/<NAME>/SKILL.md` and `.claude/skills/<name>/SKILL.md` are Copilot-only fields (`handoffs:`, `disable-model-invocation:`) and Claude-only fields (`allowed-tools:`, `context: fork`). Body content MUST be identical — enforced by `lint-skill-command-drift`.

---

### 11. CLAUDE.md Length and Scoping Strategy

**Current state in ENG-HANDBOOK.md**: §14.11 mentions the 200-line target for CLAUDE.md but provides no guidance on scaling for large monorepos or what per-directory files might help.

**Suggested addition to §14.11**:

Scaling strategy for large monorepos: Consider adding subdirectory-level CLAUDE.md files. Claude loads them lazily when reading files in those directories — they supplement (not replace) the root CLAUDE.md.

Best practice for cryptoutil:
- Root `CLAUDE.md`: Architecture summary, agent/skill tables, `@.github/instructions/*` imports — keep under 200 lines.
- `.claude/rules/framework.md` (with `paths: internal/apps-framework/**`): Framework-specific seam injection and test helper requirements.
- `.claude/rules/tests.md` (with `paths: **/*_test.go`): Test-specific rules (t.Parallel, table-driven, DisableKeepAlives).

Monitoring adherence: If Claude consistently ignores a rule in CLAUDE.md, the file may have grown too large. Extract the violated section into a path-scoped rule file instead.
