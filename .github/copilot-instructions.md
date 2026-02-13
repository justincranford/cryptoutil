# Copilot Instructions

## Core Principles

- **Keep main file short** `.github\copilot-instructions.md`
- **Keep rules short** - one directive per line
- **Instruction files auto-discovered and used in alphanumeric order from** `.github/instructions/*.instructions.md`
- **Reference external & project resources** - avoid duplication of content
- **ALWAYS use built-in tools over terminal commands**
- **Custom agent tool names** - Use official [VS Code Copilot Chat Tools Reference](https://code.visualstudio.com/docs/copilot/chat/chat-tools) and [Chat Tools API Reference](https://code.visualstudio.com/docs/copilot/reference/copilot-vscode-features#_chat-tools) for correct tool names when creating/editing `.agent.md` files
- **ALWAYS Do regular commits and pushes** to enable workflow monitoring and validation
- **ALWAYS bias towards quality, correctness, completeness, thoroughness, reliability, efficiency, and accuracy** - NEVER bias towards fast completion at the expense of quality
- **ALWAYS take the time required to do things correctly** - Time and token budgets are not constraints
- **ALWAYS prioritize doing things right over doing things quickly** - Quality over speed is mandatory
- **Prefer full execution over summaries**
- **Do not ask follow-up questions unless explicitly requested**
- **When given a plan, execute all steps completely**
- **Avoid conversational check-ins**
- **ALWAYS prefer lean documentation** - Append to existing docs (DETAILED.md, plan.md, tasks.md) instead of creating new analysis files
- **NEVER create verbose analysis files** - No ANALYSIS.md, COMPLETION-ANALYSIS.md, SESSION-*.md files

## Instruction Files Reference

**Note**: Maintain as a single concise table. DO NOT split into category subsections.

| File | Description |
|------|-------------|
| [01-01.terminology](.github/instructions/01-01.terminology.instructions.md) | RFC 2119 keywords (MUST, SHOULD, MAY, CRITICAL) |
| [01-02.beast-mode](.github/instructions/01-02.beast-mode.instructions.md) | Continuous work directive |
| [02-01.architecture](.github/instructions/02-01.architecture.instructions.md) | Architecture, service template, and HTTPS patterns |
| [02-02.versions](.github/instructions/02-02.versions.instructions.md) | Version requirements |
| [02-03.observability](.github/instructions/02-03.observability.instructions.md) | Observability and monitoring |
| [02-04.openapi](.github/instructions/02-04.openapi.instructions.md) | OpenAPI spec and code generation |
| [02-05.security](.github/instructions/02-05.security.instructions.md) | Security, cryptography, hashing, and PKI |
| [02-06.authn](.github/instructions/02-06.authn.instructions.md) | Authentication and authorization patterns |
| [03-01.coding](.github/instructions/03-01.coding.instructions.md) | Coding patterns and standards |
| [03-02.testing](.github/instructions/03-02.testing.instructions.md) | Testing standards and quality gates |
| [03-03.golang](.github/instructions/03-03.golang.instructions.md) | Go project structure and standards |
| [03-04.data-infrastructure](.github/instructions/03-04.data-infrastructure.instructions.md) | Database, SQLite/GORM, and server builder |
| [03-05.linting](.github/instructions/03-05.linting.instructions.md) | Code quality and linting |
| [04-01.deployment](.github/instructions/04-01.deployment.instructions.md) | CI/CD, Docker, and deployment |
| [05-01.cross-platform](.github/instructions/05-01.cross-platform.instructions.md) | Platform-specific tooling |
| [05-02.git](.github/instructions/05-02.git.instructions.md) | Git commands and commit conventions |
| [06-01.evidence-based](.github/instructions/06-01.evidence-based.instructions.md) | Evidence-based task completion |
| [06-02.agent-format](.github/instructions/06-02.agent-format.instructions.md) | Agent file format and structure |
