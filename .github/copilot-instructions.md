# Copilot Instructions

- Follow README and instructions files.
- Refer to architecture and usage examples in README.
- Keep all instruction files clear, concise, and focused on their specific domain
- Avoid duplication of guidance across instruction files - each file should cover its unique area
- When adding new instruction files:
  1. Create the instruction file in `.github/instructions/` with the appropriate frontmatter and content
  2. Register the file path in `.vscode/settings.json` under `github.copilot.chat.codeGeneration.instructionsFiles` array
  3. Commit and push both changes to ensure Copilot uses the new instructions

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
