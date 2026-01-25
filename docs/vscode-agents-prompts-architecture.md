# VS Code Agents vs Prompts - Architecture Analysis

## Summary

**Agents** and **Prompts** are NOT duplicates - they work together in a parent-child relationship:

- .github/agents/*.agent.md - **Agents** = Dropdown entries with tools, instructions, model settings
- .github/prompts/*.prompt.md - **Prompts** = Reusable workflows that reference agents via gent: field

## The Relationship

### Agents (Parent)

Agents define the **execution context**:

- What tools are available
- What instructions to follow
- Which model to use
- Tool permissions (read-only vs full edit)

### Prompts (Child)

Prompts define **specific workflows** that run within an agent context:

- Step-by-step task procedures
- Template generation
- Guided interactions
- Triggered with "/" in chat

## Example: autonomous-execution

**Agent** (.github/agents/autonomous-execution.agent.md):

- Defines tools: ['extensions', 'codebase', 'usages', 'problems', 'changes', ...]
- Appears in Agents dropdown
- Sets execution context (autonomous mode, no asking permission)

**Prompt** (.github/prompts/autonomous-execution.prompt.md):

- References agent: gent: autonomous
- Provides detailed workflow instructions
- Triggered with "/autonomous-execution" in chat
- Runs within autonomous agent context

## Why Beastmode Appears in Dropdown

**Root Cause**: beast-mode-3.1.prompt.md has  ools: field in frontmatter

**Explanation**: VS Code interprets any file with  ools: field as a potential agent for backward compatibility with old .chatmode.md files. The  ools: field is agent-specific metadata.

**Solution Options**:

1. Remove  ools: field from beast-mode-3.1.prompt.md (makes it pure prompt)
2. Create beast-mode.agent.md and have prompt reference it
3. Move beast-mode to .github/agents/ and rename to .agent.md

## Current Architecture

**Agents (12)**:

- autonomous-execution
- doc-sync  
- workflow-fixing
- speckit.analyze
- speckit.checklist
- speckit.clarify
- speckit.constitution
- speckit.implement
- speckit.plan
- speckit.specify
- speckit.tasks
- speckit.taskstoissues

**Prompts That Reference Agents (14)**:

- autonomous-execution.prompt.md  agent: autonomous
- doc-sync.prompt.md  agent: docsync
- workflow-fixing.prompt.md  agent: workflow
- plan-tasks-quizme.prompt.md  agent: plan
- speckit.*.prompt.md  agent: speckit.*
- beast-mode-3.1.prompt.md  NO agent reference (has tools: field instead)

**Standalone Prompts (0)**: None currently - all reference agents or have tools: field

## Best Practices

1. **Agents**: Define execution context (tools, permissions, model)
2. **Prompts**: Define workflows that run within agent context
3. **Avoid**: Putting  ools: in prompt files (causes agent dropdown appearance)
4. **Use**: gent: field in prompts to reference the agent context

## Documentation References

- Custom Agents: <https://code.visualstudio.com/docs/copilot/customization/custom-agents>
- Prompt Files: <https://code.visualstudio.com/docs/copilot/customization/prompt-files>
- Tool Priority: Prompt tools > Agent tools > Default tools
