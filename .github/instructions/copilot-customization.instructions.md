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
- Create task-specific instruction files in `.github/instructions/` with proper YAML frontmatter
- Use the `applyTo` property to target specific file types
- Create reusable prompt files in `.github/prompts/` for common tasks
- Files are automatically discovered by VS Code - no manual registration required

## Best Practices

- Split instructions into multiple files organized by topic or task type
- Maintain consistency between related instruction files  
- Use proper `.instructions.md` extension and YAML frontmatter for autodiscovery
- Enable instruction files with `github.copilot.chat.codeGeneration.useInstructionFiles: true` setting
- Use the appropriate instruction type for each task:
  - Code generation: `.github/copilot-instructions.md` or task-specific `.instructions.md` files
  - Code review: Configure with `github.copilot.chat.reviewSelection.instructions` setting
  - Commit messages: Configure with `github.copilot.chat.commitMessageGeneration.instructions` setting
  - PR descriptions: Configure with `github.copilot.chat.pullRequestDescriptionGeneration.instructions` setting

## Terminal Command Guidelines

- **Avoid `cd` commands** in terminal operations - they are not authorized and break agentic iteration
- **Use full paths** when referencing files/directories outside current context
- **Work within current directory** when possible to maintain context
- **Prefer authorized commands** like `docker`, `pwd`, `git` for navigation and operations
- **Use command flags** like `-f` or `--file` to specify paths instead of changing directories
