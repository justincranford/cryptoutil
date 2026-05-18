---
name: customization-scaffold
description: "Create repo-local customization files for agents, instructions, or skills, including required Claude counterparts for dual-canonical artifacts. Use when adding a new .github/.claude customization so file format, catalog registration, and drift rules stay compliant."
argument-hint: "[agent NAME | instruction NN-NN.name | skill NAME]"
---

Create the correct repo-local customization artifact and its required mirrored files.

## Purpose

Use when adding a new repository customization artifact under `.github/` or `.claude/`.
This single skill replaces the separate scaffold-only helpers for agents,
instructions, and skills.

## Key Rules

- Pick one artifact type per invocation: `agent`, `instruction`, or `skill`
- Agents are dual-canonical: create BOTH `.github/agents/NAME.agent.md` and `.claude/agents/NAME.md`
- Skills are dual-canonical: create BOTH `.github/skills/NAME/SKILL.md` and `.claude/skills/NAME/SKILL.md`
- Instructions are Copilot-only: create `.github/instructions/NN-NN.name.instructions.md`
- Agent and skill body content MUST stay identical across Copilot and Claude pairs; only permitted frontmatter differences may differ
- Update the relevant catalog surfaces in the same change: `.github/skills/README.md`, `.github/copilot-instructions.md`, `CLAUDE.md`, and `docs/ENG-HANDBOOK.md` when the artifact should be discoverable there
- Run `go run ./cmd/cicd-lint lint-docs` after creating or updating any customization artifact
- Use `sync-copilot-claude` to audit or repair existing drift; use this skill to create new artifacts with the correct structure from the start
- Maintain Copilot agent `tools:` allowlists here when VS Code, Copilot, extensions, or MCP servers change tool availability

## Agent Scaffold Rules

- Copilot file: `.github/agents/NAME.agent.md`
- Claude file: `.claude/agents/NAME.md`
- Copilot `name:` MUST use `copilot-NAME`; Claude `name:` MUST use `claude-NAME`
- Copilot file MUST include a `tools:` whitelist; Claude file MUST omit `tools:`
- Agents are self-contained and MUST embed the required autonomous-execution or domain guidance they rely on
- Code-modifying agents MUST reference the relevant `docs/ENG-HANDBOOK.md` sections for testing, quality, and coding standards

## Instruction Scaffold Rules

- Filename pattern: `.github/instructions/NN-NN.name.instructions.md`
- YAML frontmatter MUST contain `description:` and `applyTo:`
- Use `@source` blocks for propagated handbook content
- `@source` content MUST match the corresponding handbook `@propagate` block byte-for-byte
- Add the new instruction to `.github/copilot-instructions.md` when it is part of the active instruction catalogue

## Skill Scaffold Rules

- Copilot file: `.github/skills/NAME/SKILL.md`
- Claude file: `.claude/skills/NAME/SKILL.md`
- Skill directory name MUST match the `name:` field exactly
- Both files MUST contain a `## Key Rules` section
- Claude skills MUST omit Copilot-only frontmatter such as `disable-model-invocation`
- Add the skill to `.github/skills/README.md`, `.github/copilot-instructions.md`, `CLAUDE.md`, and `docs/ENG-HANDBOOK.md`

## Agent Tool Maintenance Rules

- Copilot agent tool maintenance belongs in this skill; do not create a separate Claude-oriented tool-maintenance skill for it
- Treat `.github/agents/*.agent.md` `tools:` lists as a Copilot allowlist contract
- Keep tool IDs in provider-native format:
  - Copilot built-in categories: `category/toolReferenceName`
  - Non-Copilot extension tools: `toolReferenceName` (or `name` if no `toolReferenceName`)
  - Explicit extension-prefixed tools: `publisher.extension/toolReferenceName`
- Claude agent files omit `tools:` and therefore do not need a tool-list maintenance workflow beyond normal body sync
- Validate tool sources before adding or removing any tool
- Prefer source-of-truth evidence over memory:
  - bundled VS Code extension manifests
  - installed marketplace extension manifests under `~/.vscode/extensions/*/package.json`
  - MCP config files (`.vscode/mcp.json` and user-profile `mcp.json`)
  - runtime tool picker or deferred-tool visibility in active agent sessions

## Agent Tool Source Types

Tool availability in Copilot agent mode comes from four source families:

1. Built-in tools (core + Copilot-provided categories)
2. Bundled VS Code extensions shipped with the VS Code install
3. Marketplace extensions under `~/.vscode/extensions/`
4. MCP servers configured in workspace or user-profile `mcp.json`

## Agent Tool Maintenance Workflow

1. Inventory all unique tools used in `.github/agents/*.agent.md`.
2. Resolve each tool to a real provider source.
3. Detect missing, renamed, or newly available tools after environment churn.
4. Patch the affected Copilot agent `tools:` lists.
5. Leave `.claude/agents/*.md` unchanged for tools; Claude omits the field.
6. Re-run `go run ./cmd/cicd-lint lint-docs`.

## Cryptoutil Agent Tool Baseline

Use this repository baseline as a quick sanity check:

- `agent/*`, `edit/*`, `execute/*`, `read/*`, `search/*`, `vscode/*`, `web/*`:
  source is GitHub Copilot Chat built-in tool categories
- `vscode.mermaid-chat-features/renderMermaidDiagram`:
  source is bundled VS Code extension `mermaid-chat-features`
- `selection`, `todo`:
  source is Copilot runtime/built-in tool surface

If any of these disappear or rename after an update, refresh the mapping from manifests and runtime visibility before editing agent files.

## Minimal Templates

### Agent

```markdown
---
name: copilot-example-agent
description: One-line purpose
tools:
  - edit/editFiles
argument-hint: "[arg]"
---

# Example Agent

## Purpose

What the agent does.

## Key Rules

- Rule 1.
- Rule 2.
```

### Instruction

```markdown
---
description: "Short description"
applyTo: "**"
---
# Title

## Key Rules

- Rule 1.
- Rule 2.
```

### Skill

```markdown
---
name: example-skill
description: "What it does and when to use it."
argument-hint: "[context]"
---

## Purpose

When to use this skill.

## Key Rules

- Rule 1.
- Rule 2.
```

## Checklist

- [ ] Correct file path and naming convention for the selected artifact type
- [ ] Required Copilot and Claude pair created for agents or skills
- [ ] Frontmatter fields valid for the selected file type
- [ ] `## Key Rules` present where required
- [ ] Handbook references added where the artifact relies on repo-specific standards
- [ ] Discovery/catalog entries updated in the relevant index files
- [ ] `go run ./cmd/cicd-lint lint-docs` passes

## References

Read [ENG-HANDBOOK.md Section 2.1.5 Copilot Skills](../../../docs/ENG-HANDBOOK.md#215-copilot-skills) for the project's customization taxonomy and catalogue expectations.

Read [ENG-HANDBOOK.md Section 13.4 Documentation Propagation Strategy](../../../docs/ENG-HANDBOOK.md#134-documentation-propagation-strategy) for `@propagate` and `@source` rules when the new artifact embeds propagated handbook content.

Read [.github/instructions/06-02.agent-format.instructions.md](../../../.github/instructions/06-02.agent-format.instructions.md) for dual-canonical agent and skill file requirements.
