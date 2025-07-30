# Copilot Instructions

- Follow README and instructions files.
- Refer to architecture and usage examples in README.
- When adding new instruction files:
  1. Create the instruction file in `.github/instructions/` with the appropriate frontmatter and content
  2. Register the file path in `.vscode/settings.json` under `github.copilot.chat.codeGeneration.instructionsFiles` array
  3. Commit and push both changes to ensure Copilot uses the new instructions
