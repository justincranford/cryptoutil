---
name: agent-tools-maintenance
description: "Audit, validate, and refresh Copilot agent tool allowlists across all sources (built-in tools, bundled VS Code extensions, marketplace extensions, and MCP servers). Use when tool availability changes after VS Code or extension updates."
argument-hint: "[audit|map|refresh|verify]"
---

Audit and maintain Copilot agent tools with evidence from real provider sources.

## Purpose

Use when agent tools drift after VS Code updates, Copilot extension updates,
new extension installs, or MCP server changes.

This skill helps you:
- identify all tool sources quickly
- map tool IDs to their true providers
- update `.github/agents/*.agent.md` safely
- verify that updates remain valid in your current environment

## Key Rules

- Treat `.github/agents/*.agent.md` `tools:` lists as an allowlist contract
- Keep tool IDs in provider-native format:
  - Copilot built-in categories: `category/toolReferenceName`
  - Non-Copilot extension tools: `toolReferenceName` (or `name` if no `toolReferenceName`)
  - Explicit extension-prefixed tools: `publisher.extension/toolReferenceName`
- Validate tool sources before adding or removing any tool
- Prefer source-of-truth evidence over memory:
  - extension manifests (`contributes.languageModelTools`)
  - MCP config files (`mcp.json`)
  - runtime tool picker and deferred-tool list in agent sessions
- Re-run `go run ./cmd/cicd-lint lint-docs` after modifying agent or skill metadata files

## Source Types

Tool availability in Copilot agent mode comes from four source families:

1. Built-in tools (core + Copilot-provided categories)
2. Bundled VS Code extensions (shipped with VS Code install)
3. Marketplace extensions (`~/.vscode/extensions/*/package.json`)
4. MCP servers (`.vscode/mcp.json` and user-profile `mcp.json`)

## Environment Discovery Workflow

1. Inventory all tools used in Copilot agents.
2. Resolve each tool to a provider source.
3. Detect missing, renamed, or newly available tools.
4. Patch affected agent files.
5. Validate drift/consistency linters.

## Fast Checks

### 1) Inventory unique tools from Copilot agents

Use a script to extract `tools:` entries from `.github/agents/*.agent.md` frontmatter,
then deduplicate.

### 2) Discover extension-contributed tools

Scan extension manifests for `contributes.languageModelTools` and capture:
- extension ID/path
- `name`
- `toolReferenceName`
- optional `when` gating

### 3) Check MCP source files

Inspect both:
- user profile `mcp.json`
- workspace `.vscode/mcp.json`

Record enabled servers and exposed tool families.

### 4) Confirm bundled VS Code extension tools

Scan VS Code installation extension manifests, not only marketplace extensions.

## Current Cryptoutil Mapping Baseline

Use this baseline as a quick sanity check for this repository:

- `agent/*`, `edit/*`, `execute/*`, `read/*`, `search/*`, `vscode/*`, `web/*`:
  source is GitHub Copilot Chat tool catalog (built-in categories)
- `vscode.mermaid-chat-features/renderMermaidDiagram`:
  source is bundled VS Code extension `mermaid-chat-features`
- `selection`, `todo`:
  source is Copilot runtime/built-in tooling surface (not always obvious in public docs)

If any of these disappear or rename after an update, refresh the mapping from
actual manifests and runtime tool visibility.

## Update Strategy

When a tool changes:

1. Find provider evidence first (manifest or MCP config).
2. Update the relevant `tools:` lists in `.github/agents/*.agent.md`.
3. Keep `.claude/agents/*.md` unchanged for tools (Claude files omit `tools:`).
4. Re-run `go run ./cmd/cicd-lint lint-docs`.
5. If drift still exists, run the matching sync skill:
   - `/sync-copilot-claude` for dual-canonical consistency

## Troubleshooting

- Tool visible in UI but not callable in agent:
  - check tool approval and enablement state
  - verify `when` clauses or settings gates in provider manifest
- Tool listed in agent but invocation fails:
  - confirm provider still contributes the same `toolReferenceName`
  - confirm category prefix is still correct
- Tool appears in docs but not in environment:
  - validate installed extension version and VS Code version
  - confirm MCP server is trusted, enabled, and started

## References

Read [VS Code agent tools](https://code.visualstudio.com/docs/copilot/agents/agent-tools)
for built-in tool behavior and approval model.

Read [VS Code contribution points](https://code.visualstudio.com/api/references/contribution-points#contributes.languageModelTools)
for extension `languageModelTools` schema.

Read [VS Code MCP servers](https://code.visualstudio.com/docs/copilot/customization/mcp-servers)
for `mcp.json` locations, trust model, and server lifecycle.

For this repository's dual-canonical requirements, read
[.github/instructions/06-02.agent-format.instructions.md](../../../.github/instructions/06-02.agent-format.instructions.md).
