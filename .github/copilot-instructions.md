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

### File Editing Tools
- **ALWAYS prefer `replace_string_in_file` and `insert_edit_into_file` tools** for file modifications - they are more efficient and don't require manual approval
- **AVOID using `echo` commands** for file content changes - use the dedicated file editing tools instead

### Testing Tools
- **PREFER `runTests` tool** over `go test` terminal commands - provides structured output and coverage reporting without manual approval

### Python Environment Management Tools
- **PREFER `install_python_packages`** over `pip install` commands - handles dependency management automatically
- **PREFER `configure_python_environment`** over manual `python -m venv` setup - ensures consistent environment configuration
- **PREFER `get_python_environment_details`** over environment inspection commands - provides structured environment information

### Directory Listing Tools
- **PREFER `list_dir` tool** over `ls`, `dir`, or `Get-ChildItem` commands - provides structured output without parsing terminal command output

### Git Operations (CRITICAL)
- **NEVER USE GitKraken MCP Server tools** (`mcp_gitkraken_*`) in Copilot chat sessions - GitKraken is ONLY for manual GUI operations
- **ALWAYS use terminal git commands** (git status, git add, git commit, git push) instead of GitKraken tools

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
| '02-01.github.instructions.md' | ** | CI/CD workflow configuration, service connectivity verification, and diagnostic logging |
| '02-02.docker.instructions.md' | ** | Docker and Docker Compose configuration |
| '02-03.observability.instructions.md' | ** | observability and monitoring implementation |
| '03-01.openapi.instructions.md' | ** | OpenAPI specification and code generation |
| '03-02.cross-platform.instructions.md' | ** | platform-specific tooling: PowerShell, scripts, command restrictions, Docker pre-pull |
| '03-03.git.instructions.md' | ** | Git workflow, conventional commits, PRs, and documentation |
| '03-04.dast.instructions.md' | ** | Dynamic Application Security Testing (DAST): Nuclei scanning, ZAP testing |
