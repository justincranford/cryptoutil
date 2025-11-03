# Copilot Instructions

## Core Principles
- **Instruction files auto-discovered from `.github/instructions/`** - use `.instructions.md` extension with YAML frontmatter

## General Principles

- Keep instructions short and self-contained
- Each instruction should be a single, simple statement
- Each instruction should not be verbose
- Don't reference external resources in instructions
- Store instructions in properly structured files for version control and team sharing

## CRITICAL: Tool and Command Restrictions

### Language/Shell Restrictions in Chat Sessions
- **NEVER use python** - not installed in Windows PowerShell or Alpine container images
- **NEVER use bash** - not available in Windows PowerShell
- **NEVER use powershell.exe** - not needed when already in PowerShell
- **NEVER use -SkipCertificateCheck** in PowerShell commands - only exists in PowerShell 6+
  - Alternative for PS 5.1: `[System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true}`

## Instruction File Structure

**Naming Convention**: `##-##.semantic-name.instructions.md` (Tier-Priority format)

| File | Applies To | Description |
| ------- | --------- | ----------- |
| '01-01.coding.instructions.md' | ** | coding patterns and standards |
| '01-02.testing.instructions.md' | ** | testing patterns, methodologies, and best practices |
| '01-03.golang.instructions.md' | ** | Go project structure, architecture, and coding standards |
| '01-04.database.instructions.md' | ** | database operations and ORM patterns |
| '01-05.security.instructions.md' | ** | security implementation, cryptographic operations, and network patterns |
| '01-06.linting.instructions.md' | ** | code quality, linting, and maintenance standards |
| '02-01.github.instructions.md' | .github/**/*.yml | CI/CD workflow configuration, service connectivity verification, and diagnostic logging |
| '02-02.docker.instructions.md' | **/*.yml | Docker and Docker Compose configuration |
| '02-03.observability.instructions.md' | ** | observability and monitoring implementation |
| '03-01.openapi.instructions.md' | ** | OpenAPI specification and code generation |
| '03-02.cross-platform.instructions.md' | ** | platform-specific tooling: PowerShell, scripts, command restrictions, Docker pre-pull |
| '03-03.git.instructions.md' | ** | Git workflow, conventional commits, PRs, and documentation |
| '03-04.dast.instructions.md' | ** | Dynamic Application Security Testing (DAST): Nuclei scanning, ZAP testing |
