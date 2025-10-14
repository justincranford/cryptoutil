## CI/CD Cost Efficiency
- Minimize workflow run frequency (trigger only on relevant changes)
- Avoid unnecessary matrix builds; keep job matrices minimal
- Use dependency caching to reduce billed minutes
- Skip jobs for docs-only or trivial changes
- Use job filters/conditionals to avoid wasteful runs
# Copilot Instructions


## Core Principles
- Follow README and instructions files.
- **ALWAYS check go.mod for the correct Go version** before using any Go version in Docker images, CI/CD configs, or tool installations
- Refer to architecture and usage examples in README.
- **Optimize for fastest and most efficient context injection**: Keep instructions clear, concise, and non-duplicate
 - Keep explanations short and actionable; avoid overwhelming users with unnecessary detail.
- Each instruction file should focus on its specific domain to minimize context overlap
- Avoid duplication of guidance across instruction files - each file should cover its unique area
- **Instruction files MUST be as short and efficient as possible.**
  - Remove unnecessary words, boilerplate, and repetition.
  - Use bullet points and lists instead of paragraphs where possible.
  - Avoid restating project-wide rules in every file—link or reference instead.
  - Regularly review and refactor instructions to minimize token usage and keep Copilot context efficient.
- **NEVER create documentation files unless explicitly requested**
  - Comments in code/config files are sufficient
  - README and existing docs should be updated, not supplemented with new docs
  - User asked for technical solution, not documentation library
  - Focus on solving the problem, not creating supporting materials
- When adding new instruction files:
  1. Create the instruction file in `.github/instructions/` with the appropriate frontmatter and content
  2. Use `.instructions.md` extension and proper YAML frontmatter with `applyTo` and `description` properties
  3. Files are automatically discovered by VS Code - no manual registration required
  4. Commit and push changes to ensure Copilot uses the new instructions

## Continuous Learning and Improvement

- **Learn from mistakes**: When errors occur during task execution, immediately analyze the root cause and add specific instructions to prevent recurrence
- **Validate assumptions**: Never assume code, scripts, or configurations work without testing - always verify functionality before declaring completion
- **Update instructions proactively**: After encountering new classes of errors or edge cases, enhance relevant instruction files with specific prevention guidelines
- **Think systematically about completeness**: Consider what validation steps, edge cases, or common mistakes might be missing from current workflow
- **Evolve instruction quality**: Regularly refine instructions to be more specific, actionable, and comprehensive based on real-world usage patterns
- **Document failure patterns**: When the same mistake occurs multiple times, create explicit instructions with examples of what NOT to do
- **Test-driven instruction creation**: For any new functionality (especially scripts), include testing requirements in the instructions before implementation
- **Pre-commit hook awareness**: Always review file content for trailing whitespace and line ending issues BEFORE using create_file tool

## Terminal and File Management

- Minimize temporary file creation/removal to reduce interaction prompts
- Use command chaining (`;` in PowerShell) for related operations
- Prefer existing files over temporary demos when possible

## Pre-commit Hook Guidelines

- **NEVER use shell commands** (`sh`, `bash`, `powershell`) in pre-commit configurations
- Use cross-platform tools directly (e.g., `go`, `python`) or pre-commit's built-in hooks
- Ensure all hooks work on Windows, Linux, and macOS without shell dependencies

## Current Instruction Files

| Pattern | File Path | Description |
| ------- | --------- | ----------- |
| **/*.yml | '.github/instructions/docker.instructions.md' | Instructions for Docker and Docker Compose configuration |
| ** | '.github/instructions/crypto.instructions.md' | Instructions for cryptographic operations |
| ** | '.github/instructions/errors.instructions.md' | Instructions for error reporting |
| ** | '.github/instructions/formatting.instructions.md' | Instructions for file formatting and encoding |
| ** | '.github/instructions/linting-exclusions.instructions.md' | Instructions for consistent linting exclusions across pre-commit, CI/CD, and scripts |
| ** | '.github/instructions/testing.instructions.md' | Instructions for testing |
| ** | '.github/instructions/database.instructions.md' | Instructions for database operations and ORM patterns |
| ** | '.github/instructions/openapi.instructions.md' | Instructions for OpenAPI and code generation patterns |
| ** | '.github/instructions/security.instructions.md' | Instructions for security implementation patterns |
| ** | '.github/instructions/observability.instructions.md' | Instructions for observability and monitoring implementation |
| ** | '.github/instructions/architecture.instructions.md' | Instructions for configuration and application architecture |
| ** | '.github/instructions/project-layout.instructions.md' | Instructions for Go project layout structure |
| ** | '.github/instructions/copilot-customization.instructions.md' | Instructions for VS Code Copilot customization best practices |
| ** | '.github/instructions/documentation.instructions.md' | Instructions for documentation organization and structure |
| ** | '.github/instructions/powershell.instructions.md' | Instructions for PowerShell usage on Windows |
| ** | '.github/instructions/commits.instructions.md' | Instructions for conventional commit message formatting |
| **/*.go | '.github/instructions/go-dependencies.instructions.md' | Instructions for Go dependency management |
| **/*.go | '.github/instructions/imports.instructions.md' | Instructions for Go import alias naming conventions |
| .github/workflows/*.yml | '.github/instructions/cicd.instructions.md' | Instructions for CI/CD workflow configuration and Go version consistency |
| **/dast-todos.md | '.github/instructions/todo-maintenance.instructions.md' | Instructions for maintaining actionable TODO/task lists (delete completed tasks immediately) |
| ** | '.github/instructions/act-testing.instructions.md' | Instructions for testing GitHub Actions workflows locally with act (CRITICAL: proper timeouts) |
