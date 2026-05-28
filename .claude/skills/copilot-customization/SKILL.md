---
name: copilot-customization
description: "Create, update, or delete repo-local customization files for agents, instructions, or skills, including required Claude counterparts and catalog updates. Use when changing .github/.claude customization artifacts so file format, discoverability, and drift rules stay compliant."
argument-hint: "[agent NAME | instruction NN-NN.name | skill NAME]"
---

Create, update, or remove the correct repo-local customization artifacts and their required mirrored files.

## Purpose

Use when creating, updating, or deleting repository customization artifacts under `.github/` or `.claude/`.
This single skill replaces the separate scaffold-only helpers for agents,
instructions, and skills.

## Key Rules

<!-- @from-eng-handbook as="skill-copilot-customization-core-rules" -->
- Pick one artifact type per invocation: `agent`, `instruction`, or `skill`
- Decide the operation up front: create, update, or delete
- Agents are dual-canonical: create BOTH `.github/agents/NAME.agent.md` and `.claude/agents/NAME.md`
- Skills are dual-canonical: create BOTH `.github/skills/NAME/SKILL.md` and `.claude/skills/NAME/SKILL.md`
- Agent and skill body content MUST stay identical across Copilot and Claude pairs; only permitted frontmatter differences may differ
- Run `go run ./cmd/cicd-lint lint-docs` after creating, updating, or deleting any customization artifact
<!-- @/from-eng-handbook -->
- Instruction files live in `.github/instructions/`, and Claude consumes them through the `## Instruction Files` list in `CLAUDE.md`
- Keep `CLAUDE.md` synchronized: update the `Instruction Files`, `Agents`, and `Skills` sections when their inventories change
- Update the relevant catalog surfaces in the same change: `.github/skills/README.md`, `.github/copilot-instructions.md`, `CLAUDE.md`, and `docs/ENG-HANDBOOK.md` when the artifact should be discoverable there
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
- Use `@from-eng-handbook` blocks for propagated handbook content
- `@from-eng-handbook` content MUST match the corresponding handbook `@to-appendix` block byte-for-byte
- Keep the `## Instruction Files` section in `CLAUDE.md` aligned with `.github/copilot-instructions.md`
- Add or remove the instruction in `.github/copilot-instructions.md` when it is part of the active instruction catalogue

## Skill Scaffold Rules

- Copilot file: `.github/skills/NAME/SKILL.md`
- Claude file: `.claude/skills/NAME/SKILL.md`
- Skill directory name MUST match the `name:` field exactly
- Both files MUST contain a `## Key Rules` section
- Claude skills MUST omit Copilot-only frontmatter such as `disable-model-invocation`
- Add or remove the skill in `.github/skills/README.md`, `.github/copilot-instructions.md`, `CLAUDE.md`, and `docs/ENG-HANDBOOK.md`

## Agent Tool Maintenance Rules

- Keep Copilot agent tool maintenance in this skill; do not split it into a separate tool-maintenance skill
- Treat `.github/agents/*.agent.md` `tools:` lists as a Copilot allowlist contract; Claude agent files omit `tools:`
- Validate tool IDs against real sources before changing them: built-in Copilot categories, bundled VS Code extensions, installed marketplace extensions, or MCP servers
- Use provider-native IDs: `category/toolReferenceName` for Copilot built-ins, `toolReferenceName` or `name` for extension tools, and `publisher.extension/toolReferenceName` when explicitly namespaced
- After any tool-list change, rerun `go run ./cmd/cicd-lint lint-docs`

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

- [ ] Correct file path and naming convention for the selected artifact type and operation
- [ ] Required Copilot and Claude pair created for agents or skills
- [ ] Frontmatter fields valid for the selected file type
- [ ] `## Key Rules` present where required
- [ ] Handbook references added where the artifact relies on repo-specific standards
- [ ] Discovery/catalog entries updated or removed in the relevant index files
- [ ] `go run ./cmd/cicd-lint lint-docs` passes

## References

Read [ENG-HANDBOOK.md Section 2.1.5 Copilot Skills](../../../docs/ENG-HANDBOOK.md#215-copilot-skills) for the project's customization taxonomy and catalogue expectations.

Read [ENG-HANDBOOK.md Section 13.4 Documentation Propagation Strategy](../../../docs/ENG-HANDBOOK.md#134-documentation-propagation-strategy) for `@to-appendix` and `@from-eng-handbook` rules when the new artifact embeds propagated handbook content.

Read [.github/instructions/06-02.agent-format.instructions.md](../../../.github/instructions/06-02.agent-format.instructions.md) for dual-canonical agent and skill file requirements.
