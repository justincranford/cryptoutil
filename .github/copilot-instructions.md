# Copilot Instructions

## Core Principles
- **Instruction files auto-discovered from `.github/instructions/`** - use `.instructions.md` extension with YAML frontmatter

## Instruction File Structure

**Naming Convention**: `##-##.semantic-name.instructions.md` (Tier-Priority format)

| File | Applies To | Description |
| ------- | --------- | ----------- |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\01-01.copilot-customization.instructions.md' | ** | Instructions for VS Code Copilot customization and critical restrictions |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\02-01.coding.instructions.md' | ** | Instructions for coding patterns and standards |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\02-02.testing.instructions.md' | ** | Instructions for testing patterns, methodologies, and best practices |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\02-03.golang.instructions.md' | ** | Instructions for Go project structure, architecture, and coding standards |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\02-04.linting.instructions.md' | ** | Instructions for code quality, linting, and maintenance standards |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\02-05.security.instructions.md' | ** | Instructions for security implementation, cryptographic operations, and network patterns |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\03-01.docker.instructions.md' | **/*.yml | Instructions for Docker and Docker Compose configuration |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\03-02.cicd.instructions.md' | .github/workflows/*.yml | Instructions for CI/CD workflow configuration, service connectivity verification, and act testing |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\03-03.database.instructions.md' | ** | Instructions for database operations and ORM patterns |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\03-04.observability.instructions.md' | ** | Instructions for observability and monitoring implementation |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\04-01.openapi.instructions.md' | ** | Instructions for OpenAPI specification and code generation |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\04-02.cross-platform.instructions.md' | ** | Instructions for platform-specific tooling: PowerShell, scripts, command restrictions, Docker pre-pull |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\04-03.git.instructions.md' | ** | Instructions for Git workflow, conventional commits, PRs, and documentation |
| 'c:\Dev\Projects\cryptoutil\.github\instructions\04-04.dast.instructions.md' | ** | Instructions for Dynamic Application Security Testing (DAST): Nuclei scanning, ZAP testing |
