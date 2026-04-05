# Claude Code — Best Practices & File Structure

**Created**: 2026-04-05
**Source**: Anthropic Claude Code official documentation + agentskills.io open standard

---

## 1. Copilot Skills → Claude Skills Mapping

In the initial dual-format strategy (framework-v6/v7), Copilot Skills (`.github/skills/<NAME>/SKILL.md`)
were mapped to Claude Commands (`.claude/commands/<NAME>.md`) because that was the only Claude
Code mechanism for custom slash commands at the time.

Claude Code has since introduced `.claude/skills/<name>/SKILL.md` as a first-class concept under
the [Agent Skills open standard](https://agentskills.io/). All 15 legacy commands have been migrated
to skills and the `.claude/commands/` directory has been removed.

| Copilot | Claude Code | Status |
|---------|-------------|--------|
| `.github/skills/<NAME>/SKILL.md` | `.claude/skills/<name>/SKILL.md` | ✅ CORRECT |
| `.github/agents/<NAME>.agent.md` | `.claude/agents/<NAME>.md` | ✅ CORRECT |

---

## 2. `.claude/` Directory Structure Reference

```
.claude/                        # Claude Code project configuration
├── CLAUDE.md                   # Project instructions (loaded every session)
│                               # Target: <200 lines. Longer = reduced adherence.
│                               # Block HTML comments <!-- ... --> are stripped.
│                               # @path imports inline-expand (max 5 hops).
├── agents/                     # Custom sub-agent definitions
│   └── <name>.md               # Agent file (YAML frontmatter + system prompt)
├── skills/                     # Custom slash commands (preferred new format)
│   └── <name>/                 # Each skill is a DIRECTORY
│       ├── SKILL.md            # Required entrypoint
│       ├── references/         # Optional: detailed docs loaded on demand
│       ├── scripts/            # Optional: executable code
│       └── assets/             # Optional: templates, resources
├── commands/                   # Legacy command files (REMOVED — migrated to skills/)
│   └── <name>.md               # Single-file commands (flat, no directories)
├── rules/                      # Path-scoped project rules
│   └── *.md                    # Rule files (optional `paths` frontmatter)
├── settings.json               # Project settings (team-level, commit)
├── settings.local.json         # Local settings (personal, gitignore)
├── agent-memory/               # Persistent memory for project-scoped agents
└── worktrees/                  # Isolated git worktrees (--worktree flag)
```

**User-level** (`~/.claude/`): `CLAUDE.md`, `agents/`, `skills/`, `rules/`, `projects/<proj>/memory/`

---

## 3. `CLAUDE.md` Format

**Format**: Plain Markdown. **No YAML frontmatter.**

### Content Guidelines

- Target **under 200 lines** per file. Shorter = better adherence (no hard maximum enforced).
- Content is delivered as a **user message** after the system prompt (not part of system prompt).
- Use `@path/to/file` to import other files inline (max 5 hops deep).
- Block HTML commments `<!-- text -->` are stripped (useful for maintainer notes).
- Survives `/compact` — re-read from disk afterward.
- Multiple files concatenated (not overriding). User-level CLAUDE.md loads before project-level.

### Required Sections for cryptoutil

```markdown
# {Project Name} — Claude Code Instructions

## Architecture Source of Truth
(Links to ARCHITECTURE.md and key section index)

## Instruction Files
@.github/instructions/*.instructions.md references

## Agents
(Table of custom agents)

## Skills (Slash Commands)
(Table of available skills + when to use)
```

---

## 4. Skills Format (`.claude/skills/<name>/SKILL.md`) — PREFERRED

Skills are **directories** with `SKILL.md` as the required entrypoint. The directory structure
allows supporting files (scripts, references, assets) alongside the skill prompt.

### Complete YAML Frontmatter

```yaml
---
name: my-skill                  # Optional if matches directory name. Lowercase, hyphens, max 64 chars.
description: "..."              # When Claude should activate. Truncated at 250 chars in listing.
argument-hint: "[args]"         # Shown in autocomplete. E.g., "[package-name]"
disable-model-invocation: true  # If true: only user can invoke via /name. Default: false.
user-invocable: true            # If false: hidden from / menu but Claude can auto-invoke. Default: true.
allowed-tools: Read Grep Glob   # Tools allowed without per-use approval when skill active.
model: sonnet                   # Model override when skill is active.
effort: medium                  # low, medium, high, max (Opus 4.6 only).
context: fork                   # Run skill in isolated subagent (no conversation history).
agent: Explore                  # Subagent to use with context: fork.
paths:                          # Load skill automatically only when working with matching files.
  - "internal/**/*.go"
shell: bash                     # bash (default) or powershell
---
```

### Dynamic Context Injection

```markdown
# Current git status:
```!
git log --oneline -5
git status --short
```

# Working on: $ARGUMENTS

```

Inline: `` !`git branch --show-current` `` runs before Claude sees the prompt; output replaces the placeholder.

### String Substitutions

| Substitution | Meaning |
|---|---|
| `$ARGUMENTS` | All arguments passed to the skill |
| `$0`, `$1`, `$N` | Positional arguments (0-based) |
| `${CLAUDE_SESSION_ID}` | Current session ID |
| `${CLAUDE_SKILL_DIR}` | Absolute path to the skill's directory |

### Skill Body Structure (for cryptoutil)

```markdown
---
(frontmatter)
---

(One-paragraph summary of the skill's purpose)

## Key Rules

- Brief rule 1 (mandatory, italicize exceptions)
- Brief rule 2
(6-12 key rules maximum)

## Template / Workflow

(Step-by-step instructions or code template)

**Full reference**: [.github/skills/<NAME>/SKILL.md](.github/skills/<NAME>/SKILL.md)
```

---

## 5. Legacy Commands Format (`.claude/commands/<name>.md`) — REMOVED

The legacy single-file command format has been fully migrated to the skills directory format.
All 15 `.claude/commands/*.md` files have been moved to `.claude/skills/<name>/SKILL.md` and
the `commands/` directory has been deleted.

The `lint-skill-command-drift` linter now checks `.claude/skills/<name>/SKILL.md` exclusively.

---

## 6. Sub-agents Format (`.claude/agents/<name>.md`)

```markdown
---
name: agent-name               # Required. Unique ID, lowercase letters and hyphens.
description: "When to delegate" # Required. Claude reads this to decide when to delegate.
tools: Read Grep Glob Bash     # Allowlist. Inherits all if omitted.
disallowedTools: WebSearch     # Denylist (removed from inherited/specified).
model: sonnet                  # sonnet, opus, haiku, or inherit (default).
permissionMode: acceptEdits    # default, acceptEdits, auto, dontAsk, bypassPermissions, plan.
maxTurns: 20                   # Max agentic turns before stopping.
skills:                        # Skills preloaded into agent's context at startup.
  - test-table-driven
  - coverage-analysis
memory: project                # Persistent memory scope: user, project, or local.
color: blue                    # Agent color in UI.
---

(System prompt for the subagent. Does NOT inherit parent conversation history.)
```

**Key behavior**: Subagents receive ONLY their system prompt + basic env details. They do NOT
inherit: the full Claude Code system prompt, conversation history, or parent skills (unless
listed in the `skills` field). Agents must be self-contained.

---

## 7. Path-Scoped Rules (`.claude/rules/`)

Rules auto-load based on which files Claude is working with. Useful for per-directory coding
standards, API conventions, or test patterns.

```markdown
---
paths:
  - "internal/apps/framework/**/*.go"
  - "api/**/*.yaml"
---

# Framework Rules

When working in the framework package, always:
- Use function-parameter injection for seams (not package-level vars)
- Check testdb.NewInMemorySQLiteDB(t) for unit tests
```

**Without `paths`**: loaded at launch (same priority as `.claude/CLAUDE.md`).
**With `paths`**: loaded lazily only when Claude works with matching files.

---

## 8. Agent Skills Open Standard (agentskills.io)

The `.claude/skills/` format is based on the [Agent Skills open standard](https://agentskills.io/)
developed by Anthropic and adopted by multiple AI tools: Gemini CLI, GitHub Copilot, OpenAI Codex,
Amp, Kiro, Qodo, VS Code.

**Shared frontmatter fields** (cross-agent):

| Field | Required | Constraint |
|-------|----------|-----------|
| `name` | Yes | Max 64 chars, lowercase letters/numbers/hyphens, must match directory name |
| `description` | Yes | Max 1024 chars |

**cryptoutil strategy**: Since both Copilot and Claude now share the same `SKILL.md` format
(same directory name, same YAML fields), the dual canonical skill format is:
- `.github/skills/<NAME>/SKILL.md` — Copilot tool
- `.claude/skills/<name>/SKILL.md` — Claude Code tool

Body content SHOULD be identical. The only expected difference is in handling (Copilot may have
`handoffs:`, Claude may have `allowed-tools:` or `context: fork`).

---

## 9. Dual Canonical Strategy — Corrected

### Skills (Slash Commands)

| Concept | Copilot File | Claude File | Format |
|---------|-------------|-------------|--------|
| Skill | `.github/skills/<NAME>/SKILL.md` | `.claude/skills/<name>/SKILL.md` | SKILL.md in directory |

### Agents (Sub-agents)

| Concept | Copilot File | Claude File |
|---------|-------------|-------------|
| Agent | `.github/agents/<NAME>.agent.md` | `.claude/agents/<NAME>.md` |

### Linter Alignment

The `lint-skill-command-drift` linter in `internal/apps/tools/cicd_lint/lint_docs/` checks
`.claude/skills/<name>/SKILL.md` for each Copilot skill in `.github/skills/<NAME>/SKILL.md`.

---

## 10. Migration Checklist: Commands → Skills — COMPLETED

All 15 legacy `.claude/commands/*.md` files have been migrated to `.claude/skills/<name>/SKILL.md`.
The `lint-skill-command-drift` linter now checks `.claude/skills/` exclusively.
The `.claude/commands/` directory has been removed.

---

## 11. CLAUDE.md Length and Scoping Strategy

Current `CLAUDE.md` is at the repo root and references all instruction files via `@` imports.
This keeps the root CLAUDE.md concise (target <200 lines) while allowing deep content via imports.

**For large monorepos**: Consider adding `.claude/CLAUDE.md` alongside subdirectory-level
`CLAUDE.md` files. Claude loads them lazily when reading files in those directories.

**Best practice for cryptoutil**:
- Root `CLAUDE.md`: Architecture summary, agent/skill tables, `@.github/instructions/*` imports
- `.claude/rules/framework.md` with `paths: internal/apps/framework/**` for framework coding rules
- `.claude/rules/tests.md` with `paths: **/*_test.go` for test-specific rules
