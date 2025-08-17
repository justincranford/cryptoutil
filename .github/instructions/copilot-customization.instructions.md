---
description: "Instructions for VS Code Copilot customization"
applyTo: "**"
---
# VS Code Copilot Customization Instructions

## General Principles

- Keep instructions short and self-contained
- Each instruction should be a single, simple statement
- Don't reference external resources in instructions
- Store instructions in properly structured files for version control and team sharing

## File Structure

- Use `.github/copilot-instructions.md` as the primary instruction file
- Create task-specific instruction files in `.github/instructions/` with proper front matter
- Use the `applyTo` property to target specific file types
- Create reusable prompt files in `.github/prompts/` for common tasks

## Best Practices

- Split instructions into multiple files organized by topic or task type
- Maintain consistency between related instruction files
- Use the appropriate instruction type for each task:
  - Code generation: `.github/copilot-instructions.md` or task-specific `.instructions.md` files
  - Code review: Configure with `github.copilot.chat.reviewSelection.instructions` setting
  - Commit messages: Configure with `github.copilot.chat.commitMessageGeneration.instructions` setting
  - PR descriptions: Configure with `github.copilot.chat.pullRequestDescriptionGeneration.instructions` setting
