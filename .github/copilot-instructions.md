# Copilot Instructions

## Core Principles
- Follow README and instructions files.
- Refer to architecture and usage examples in README.
- **Optimize for fastest and most efficient context injection**: Keep instructions clear, concise, and non-duplicate
- Each instruction file should focus on its specific domain to minimize context overlap
- Avoid duplication of guidance across instruction files - each file should cover its unique area
- When adding new instruction files:
  1. Create the instruction file in `.github/instructions/` with the appropriate frontmatter and content
  2. Use `.instructions.md` extension and proper YAML frontmatter with `applyTo` and `description` properties
  3. Files are automatically discovered by VS Code - no manual registration required
  4. Commit and push changes to ensure Copilot uses the new instructions

## Terminal and File Management

- Minimize temporary file creation/removal to reduce interaction prompts
- Use command chaining (`;` in PowerShell) for related operations
- Prefer existing files over temporary demos when possible

## Pre-commit Hook Compliance

- **ALWAYS ensure files end with a single newline** (end-of-file-fixer hook)
- **NEVER leave trailing whitespace** on any line (trailing-whitespace hook)
- **USE LF line endings** for all files, never CRLF (Git line ending warnings)
- Follow formatting.instructions.md for complete file formatting guidelines

## Current Instruction Files

| Pattern | File Path | Description |
| ------- | --------- | ----------- |
| **/*.yml | '.github/instructions/docker.instructions.md' | Instructions for Docker and Docker Compose configuration |
| ** | '.github/instructions/crypto.instructions.md' | Instructions for cryptographic operations |
| ** | '.github/instructions/errors.instructions.md' | Instructions for error reporting |
| ** | '.github/instructions/formatting.instructions.md' | Instructions for file formatting and encoding |
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
| **/*.go | '.github/instructions/imports.instructions.md' | Instructions for Go import alias naming conventions |
| .github/workflows/*.yml | '.github/instructions/cicd.instructions.md' | Instructions for CI/CD workflow configuration and Go version consistency |
