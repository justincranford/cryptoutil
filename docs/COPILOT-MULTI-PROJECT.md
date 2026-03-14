# Multi-Project Copilot Sharing Strategy

**Created**: 2026-03-14
**Status**: Research / Strategy Document
**Purpose**: Centralize Copilot artifacts (agents, skills, instructions) across multiple projects

## Problem Statement

The `cryptoutil` project has a mature Copilot framework:
- 5 agents (beast-mode, doc-sync, fix-workflows, implementation-execution, implementation-planning)
- 14 skills (test-table-driven, coverage-analysis, fips-audit, etc.)
- 16 instruction files (terminology, architecture, testing, security, etc.)
- 1 copilot-instructions.md (master index)
- Propagation system (ARCHITECTURE.md → instructions via `@propagate`/`@source` markers)

Other projects (springlock, springs, spring, ~30 total in `c:\Dev\Projects\`) have minimal or no Copilot configuration. The goal is to share reusable Copilot artifacts across projects without manual copy-paste maintenance.

## Current State

| Project | Agents | Skills | Instructions | copilot-instructions.md |
|---------|--------|--------|-------------|------------------------|
| cryptoutil | 5 | 14 | 16 | ✅ |
| springlock | 1 (beast-mode only) | 0 | 0 | ❌ |
| springs | 0 | 0 | 0 | ❌ |
| spring | 0 | 0 | 0 | ❌ |
| ~26 others | 0 | 0 | 0 | ❌ |

## Artifact Classification

Not all Copilot artifacts are shareable. They fall into three categories:

### Universal (Shareable Across All Projects)

These encode general development practices, not project-specific knowledge:

- **Agents**: beast-mode (autonomous execution), implementation-planning, implementation-execution
- **Skills**: test-table-driven, test-fuzz-gen, test-benchmark-gen, coverage-analysis, agent-scaffold, instruction-scaffold, skill-scaffold
- **Instructions**: 01-01 terminology (RFC 2119), 01-02 beast-mode (continuous work), 05-02 git (commit conventions), 06-01 evidence-based, 06-02 agent-format

### Language-Specific (Shareable Within Go Projects)

These encode Go-specific patterns:

- **Instructions**: 03-01 coding, 03-02 testing, 03-03 golang, 03-05 linting
- **Skills**: migration-create, fitness-function-gen

### Project-Specific (NOT Shareable)

These encode cryptoutil domain knowledge:

- **Instructions**: 02-01 architecture, 02-03 observability, 02-04 openapi, 02-05 security, 02-06 authn, 03-04 data-infrastructure, 04-01 deployment
- **Skills**: fips-audit, new-service, contract-test-gen, openapi-codegen, propagation-check
- **Agents**: doc-sync, fix-workflows (heavily coupled to cryptoutil's CI/CD)

## Sharing Options Analysis

### Option 1: GitHub Organization `.github` Repository

**How it works**: Create a `.github` repository in your GitHub organization (e.g., `justincranford/.github`). Files in this repo apply organization-wide.

**Supported files**:
- `.github/copilot-instructions.md` — Organization-level Copilot instructions
- Profile-level README, funding, templates

**Limitations**:
- ❌ Does NOT support agents (`.github/agents/*.agent.md`)
- ❌ Does NOT support skills (`.github/skills/*/SKILL.md`)
- ❌ Does NOT support instruction files (`.github/instructions/*.instructions.md`)
- ✅ Supports only `copilot-instructions.md` at org level
- ❌ Requires GitHub organization (not personal repos)

**Verdict**: **Insufficient** — only shares one file (copilot-instructions.md), not agents/skills/instructions.

### Option 2: Git Submodules

**How it works**: Create a central repo (`copilot-central`) and include it as a git submodule in each project.

```bash
# In each project
git submodule add https://github.com/justincranford/copilot-central.git .github/shared
```

**Structure**:
```
copilot-central/
├── agents/
│   ├── beast-mode.agent.md
│   ├── implementation-planning.agent.md
│   └── implementation-execution.agent.md
├── skills/
│   ├── test-table-driven/SKILL.md
│   ├── coverage-analysis/SKILL.md
│   └── ...
├── instructions/
│   ├── 01-01.terminology.instructions.md
│   ├── 01-02.beast-mode.instructions.md
│   └── ...
└── copilot-instructions-base.md
```

**Critical problem**: VS Code Copilot expects agents in `.github/agents/`, skills in `.github/skills/`, and instructions in `.github/instructions/`. A submodule at `.github/shared/agents/` would NOT be discovered. You would need symlinks from `.github/agents/beast-mode.agent.md` → `.github/shared/agents/beast-mode.agent.md`.

**Limitations**:
- ❌ Requires symlinks to correct paths (Windows symlinks need admin or developer mode)
- ❌ Submodule update complexity (`git submodule update --remote`)
- ❌ CI/CD must handle `--recurse-submodules`
- ✅ Version pinning (each project can pin to specific commit)
- ✅ Atomic updates (update submodule ref = pull latest)

**Verdict**: **Viable but complex** — requires symlinks and careful submodule management.

### Option 3: Symbolic Links (Direct)

**How it works**: One canonical copy in a shared directory. Each project symlinks `.github/agents/`, `.github/skills/`, `.github/instructions/`.

```powershell
# Windows (requires Developer Mode or admin)
New-Item -ItemType SymbolicLink -Path ".github\agents\beast-mode.agent.md" `
  -Target "C:\Dev\Projects\copilot-central\agents\beast-mode.agent.md"
```

**Limitations**:
- ❌ Windows symlinks require Developer Mode enabled or admin privileges
- ❌ Symlinks don't travel with `git clone` (they're stored as text files unless `core.symlinks=true`)
- ❌ CI/CD (GitHub Actions) runs on Linux runners — Windows symlinks break
- ❌ Fragile: moving the shared directory breaks all links
- ✅ Zero-copy: changes in central repo immediately visible
- ✅ Simple to understand

**Verdict**: **Not recommended** — too fragile for cross-platform, breaks in CI/CD.

### Option 4: VS Code Multi-Root Workspaces

**How it works**: Create a `.code-workspace` file that includes multiple project folders.

```json
{
  "folders": [
    { "path": "C:\\Dev\\Projects\\cryptoutil" },
    { "path": "C:\\Dev\\Projects\\copilot-central" }
  ]
}
```

**How Copilot uses it**: In multi-root workspaces, VS Code Copilot loads `.github/` from the **active folder** (the folder containing the file you're editing). It does NOT merge `.github/` from multiple roots.

**Limitations**:
- ❌ Does NOT merge agents/skills/instructions across workspace roots
- ❌ Only the active folder's `.github/` applies
- ✅ Code search works across all roots (useful for context)
- ✅ Good for monorepo-like development experience

**Verdict**: **Does not solve the sharing problem** — Copilot only sees one `.github/` at a time.

### Option 5: Template Repository + Sync Script

**How it works**: Create a `copilot-central` template repo. For new projects, use GitHub's "Use this template" feature. For existing projects, use a sync script.

```bash
# Sync script (run periodically or via pre-commit hook)
#!/bin/bash
CENTRAL="$HOME/Dev/Projects/copilot-central"
TARGET=".github"

# Sync universal agents
cp "$CENTRAL/agents/beast-mode.agent.md" "$TARGET/agents/"
cp "$CENTRAL/agents/implementation-planning.agent.md" "$TARGET/agents/"
cp "$CENTRAL/agents/implementation-execution.agent.md" "$TARGET/agents/"

# Sync universal instructions
cp "$CENTRAL/instructions/01-01.terminology.instructions.md" "$TARGET/instructions/"
cp "$CENTRAL/instructions/01-02.beast-mode.instructions.md" "$TARGET/instructions/"
# ...
```

**Advantages**:
- ✅ Files are real files in each repo (no symlinks, no submodules)
- ✅ Works in CI/CD, any platform, any git client
- ✅ Each project can customize copied files (diverge when needed)
- ✅ Sync script can be selective (universal only, universal + Go, etc.)
- ✅ Version control: each project commits its own copy

**Limitations**:
- ❌ Copies diverge over time unless sync script is run regularly
- ❌ Manual sync discipline required
- ❌ Merge conflicts when central and local both change

**Verdict**: **Best practical option** — simple, reliable, cross-platform.

### Option 6: Go Module / Package (For Go Projects Only)

**How it works**: Package shared Copilot artifacts as a Go module.

```
module github.com/justincranford/copilot-go-standards

// Embed as //go:embed .github/
```

**Verdict**: **Over-engineered** — Copilot artifacts are markdown files, not Go code. A Go module adds unnecessary complexity.

### Option 7: Monorepo

**How it works**: Put all projects in one repository. One `.github/` serves all.

**Verdict**: **Impractical** — ~30 independent projects with different lifecycles, languages, and deployment targets. Monorepo adds organizational overhead that outweighs sharing benefits.

## Recommended Strategy: Template Repo + Sync Script (Option 5)

### Architecture

```
github.com/justincranford/copilot-central/
├── README.md
├── sync.go                           # Cross-platform sync tool (Go, not bash)
├── universal/                        # Shareable across ALL projects
│   ├── agents/
│   │   ├── beast-mode.agent.md
│   │   ├── implementation-planning.agent.md
│   │   └── implementation-execution.agent.md
│   ├── skills/
│   │   ├── test-table-driven/SKILL.md
│   │   ├── test-fuzz-gen/SKILL.md
│   │   ├── test-benchmark-gen/SKILL.md
│   │   ├── coverage-analysis/SKILL.md
│   │   ├── agent-scaffold/SKILL.md
│   │   ├── instruction-scaffold/SKILL.md
│   │   └── skill-scaffold/SKILL.md
│   └── instructions/
│       ├── 01-01.terminology.instructions.md
│       ├── 01-02.beast-mode.instructions.md
│       ├── 05-02.git.instructions.md
│       ├── 06-01.evidence-based.instructions.md
│       └── 06-02.agent-format.instructions.md
├── go/                               # Go-specific artifacts
│   ├── instructions/
│   │   ├── 03-01.coding.instructions.md
│   │   ├── 03-02.testing.instructions.md
│   │   ├── 03-03.golang.instructions.md
│   │   └── 03-05.linting.instructions.md
│   └── skills/
│       ├── migration-create/SKILL.md
│       └── fitness-function-gen/SKILL.md
└── templates/
    ├── copilot-instructions-go.md    # Base copilot-instructions.md for Go projects
    └── copilot-instructions-base.md  # Base copilot-instructions.md for any project
```

### Sync Tool (Go, Cross-Platform)

Write `sync.go` as a Go CLI tool (not bash/PowerShell — per cross-platform instructions):

```
go run github.com/justincranford/copilot-central/sync@latest \
  --target=. \
  --profile=go \
  --dry-run
```

Profiles:
- `base` — universal agents + skills + instructions only
- `go` — base + Go-specific artifacts
- `full` — everything (for cryptoutil-like projects)

### Workflow

1. **New project**: `go run .../sync@latest --target=. --profile=go` → copies universal + Go artifacts
2. **Existing project**: Same command, uses `--merge` flag to preserve local customizations
3. **Central update**: Edit `copilot-central`, push. Run sync in each project when ready.
4. **Divergence**: Local customizations are fine — the copied files belong to each project. Re-sync overwrites with `--force`, or merges with `--merge`.

### Migration Path

1. Create `github.com/justincranford/copilot-central` repo
2. Extract universal artifacts from cryptoutil (copy, not move — cryptoutil keeps originals)
3. Extract Go-specific artifacts
4. Write `sync.go` tool
5. Test on `springlock` first (already has beast-mode.agent.md)
6. Roll out to other projects incrementally

## Alternatives Considered But Not Recommended

| Option | Why Not |
|--------|---------|
| Org-level `.github` repo | Only supports `copilot-instructions.md`, not agents/skills/instructions |
| Git submodules | Requires symlinks, complex CI/CD handling |
| Direct symlinks | Breaks in CI/CD, requires Windows Developer Mode |
| Multi-root workspaces | Copilot only reads active folder's `.github/` |
| Go module packaging | Over-engineered for markdown files |
| Monorepo | Impractical for ~30 independent projects |

## VS Code Copilot Feature Summary

### What VS Code Copilot Loads (Current Architecture)

| Context | Source | Loaded When |
|---------|--------|-------------|
| copilot-instructions.md | `.github/copilot-instructions.md` | Every normal chat (NOT agents) |
| Instruction files | `.github/instructions/*.instructions.md` | Every normal chat (NOT agents) |
| Agent files | `.github/agents/*.agent.md` | Only when `/agent-name` invoked |
| Skill files | `.github/skills/*/SKILL.md` | When referenced by agent `skills:` or auto-triggered |
| VS Code settings | `.vscode/settings.json` | Always |

### Key Architectural Constraints

1. **Agent isolation**: Agents do NOT inherit copilot-instructions.md or instruction files
2. **Single .github/ root**: Copilot reads from one `.github/` directory per workspace folder
3. **No cross-folder merging**: Multi-root workspaces don't merge `.github/` from multiple folders
4. **No org-level agents**: GitHub org-level `.github` repo only supports `copilot-instructions.md`
5. **Skills are prompts**: Skills are markdown prompt templates, not executable code
6. **Handoffs are drafts**: `send: false` means handoff creates a draft message, user must approve

### What Would Help (Feature Requests)

1. **Workspace-level `.copilot/` directory**: A directory at the workspace root (not per-folder) that Copilot always loads — would enable multi-root workspace sharing
2. **Agent inheritance**: Allow agents to `extends: beast-mode` to inherit behavioral patterns — would reduce repetition
3. **Instruction packages**: Allow `imports: ["@justincranford/copilot-standards"]` in copilot-instructions.md — would enable npm-like sharing
4. **Remote instruction sources**: Allow `instructions: { remote: "https://github.com/justincranford/copilot-central/instructions/" }` — would enable centralized management

## Cross-References

- [ARCHITECTURE.md Section 2.1](ARCHITECTURE.md#21-agent-orchestration-strategy) — Agent architecture and catalog
- [ARCHITECTURE.md Section 2.1.1](ARCHITECTURE.md#211-agent-architecture) — Agent isolation principle
- [ARCHITECTURE.md Section 12.7](ARCHITECTURE.md#127-documentation-propagation-strategy) — Propagation marker system
