# Copilot Instructions

## Core Principles

- **Keep main file short** `.github\copilot-instructions.md`
- **Keep rules short** - one directive per line
- **Instruction files auto-discovered and used in alphanumeric order from** `.github/instructions/*.instructions.md`
- **Reference external & project resources** - avoid duplication of content
- **ALWAYS use built-in tools over terminal commands**
- **Custom agent tool names** - Use official [VS Code Copilot Chat Tools Reference](https://code.visualstudio.com/docs/copilot/chat/chat-tools) and [Chat Tools API Reference](https://code.visualstudio.com/docs/copilot/reference/copilot-vscode-features#_chat-tools) for correct tool names when creating/editing `.agent.md` files
- **ALWAYS Do regular commits and pushes** to enable workflow monitoring and validation
- **ALWAYS bias towards accuracy, completeness, and correctness** - NEVER bias towards fast completion at the expense of quality
- **ALWAYS take the time required to do things correctly** - Time and token budgets are not constraints for Speckit work
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
| [01-02.beast-mode](.github/instructions/01-02.beast-mode.instructions.md) | Beast mode directive |
| [02-01.architecture](.github/instructions/02-01.architecture.instructions.md) | Products and services architecture patterns |
| [02-02.service-template](.github/instructions/02-02.service-template.instructions.md) | Service template requirements and patterns for all cryptoutil services |
| [02-03.https-ports](.github/instructions/02-03.https-ports.instructions.md) | HTTPS ports and TLS configuration |
| [02-04.versions](.github/instructions/02-04.versions.instructions.md) | Instructions for version requirements |
| [02-05.observability](.github/instructions/02-05.observability.instructions.md) | Instructions for observability and monitoring implementation |
| [02-06.openapi](.github/instructions/02-06.openapi.instructions.md) | Instructions for OpenAPI |
| [02-07.cryptography](.github/instructions/02-07.cryptography.instructions.md) | Cryptographic patterns and FIPS compliance |
| [02-08.hashes](.github/instructions/02-08.hashes.instructions.md) | Hash registry and password hashing |
| [02-09.pki](.github/instructions/02-09.pki.instructions.md) | PKI and certificate management |
| [02-10.authn](.github/instructions/02-10.authn.instructions.md) | Authentication and authorization patterns |
| [03-01.coding](.github/instructions/03-01.coding.instructions.md) | Instructions for coding patterns and standards |
| [03-02.testing](.github/instructions/03-02.testing.instructions.md) | Instructions for testing |
| [03-03.golang](.github/instructions/03-03.golang.instructions.md) | Go project structure and standards |
| [03-04.database](.github/instructions/03-04.database.instructions.md) | Instructions for database operations and ORM patterns |
| [03-05.sqlite-gorm](.github/instructions/03-05.sqlite-gorm.instructions.md) | Instructions for SQLite configuration with GORM, transaction patterns, and concurrent operations |
| [03-06.security](.github/instructions/03-06.security.instructions.md) | Instructions for security implementation patterns |
| [03-07.linting](.github/instructions/03-07.linting.instructions.md) | Instructions for code quality, linting, and maintenance standards |
| [03-08.server-builder](.github/instructions/03-08.server-builder.instructions.md) | Server builder pattern and merged migrations |
| [04-01.github](.github/instructions/04-01.github.instructions.md) | CI/CD workflow configuration |
| [04-02.docker](.github/instructions/04-02.docker.instructions.md) | Instructions for Docker and Docker Compose configuration |
| [05-01.cross-platform](.github/instructions/05-01.cross-platform.instructions.md) | Platform-specific tooling |
| [05-02.git](.github/instructions/05-02.git.instructions.md) | Instructions for local Git commands and commit conventions |
| [05-03.dast](.github/instructions/05-03.dast.instructions.md) | Dynamic Application Security Testing |
| [06-01.evidence-based](.github/instructions/06-01.evidence-based.instructions.md) | Instructions for evidence-based task completion and validation |
| [06-02.agent-format](.github/instructions/06-02.agent-format.instructions.md) | Agent file format and structure standards |
