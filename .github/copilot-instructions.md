# Copilot Instructions

## Core Principles
- **Instruction files auto-discovered from `.github/instructions/`** - use `.instructions.md` extension with YAML frontmatter

## General Principles

- Keep instructions short and self-contained
- Each instruction should be a single, simple statement
- Each instruction should not be verbose
- Don't reference external resources in instructions
- Store instructions in properly structured files for version control and team sharing
- Use built-in tools one at a time, to minimize prompts for approval
- When approaching rate limiting, wait between requests as needed
- GitHub Copilot Chat Extension monitors GitHub Copilot Service rate limiting via HTTP response headers
- **Rate Limit Monitoring**: Monitor HTTP response headers (`X-RateLimit-Remaining`, `X-RateLimit-Reset`) to detect approaching rate limit thresholds
- **Rate Limit Checking**: You can also call the GET /rate_limit endpoint to check your rate limit. Calling this endpoint does not count against your primary rate limit, but it can count against your secondary rate limit. See [REST API endpoints for rate limits](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api?apiVersion=2022-11-28#checking-the-status-of-your-rate-limit). When possible, you should use the rate limit response headers instead of calling the API to check your rate limit.
- **Rate Limit Error Handling**: Follow GitHub's best practices for handling rate limit errors. Use `retry-after` header when present, check `x-ratelimit-remaining` and `x-ratelimit-reset` headers, implement exponential backoff for secondary rate limits, and avoid making requests while rate limited to prevent integration bans. See [Best practices for using the REST API](https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api#handle-rate-limit-errors-appropriately).

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

## CRITICAL: Tool and Command Restrictions

### File Editing Tools
- **ALWAYS prefer `create_file`, and `create_directory`** for file modifications - they are purpose-built and avoid terminal auto-approval prompts
- **Document when file-level edits require multiple passes** so human reviewers can trace changes
- **AVOID using shell redirection commands** for file content changes; the editing tools provide cleaner diffs and auditability

### Testing Tools
- **PREFER `runTests` tool** over `go test` terminal commands - provides structured output and coverage reporting without manual approval

### Python Environment Management Tools
- **PREFER `install_python_packages`** over `pip install` commands - handles dependency management automatically
- **PREFER `configure_python_environment`** over manual `python -m venv` setup - ensures consistent environment configuration
- **PREFER `get_python_environment_details`** over environment inspection commands - provides structured environment information

### Directory Listing Tools
- **PREFER `list_dir` tool** over `ls`, `dir`, or `Get-ChildItem` commands - provides structured output without parsing terminal command output

### Workspace Tools
- **PREFER `read_file`** over `type`, `cat`, or `Get-Content` commands for file inspection
- **PREFER `file_search`** over `find`, `dir /s`, or `Get-ChildItem -Recurse`
- **PREFER `semantic_search`** over multi-step `grep` or `rg` shell pipelines when looking for concepts
- **PREFER `get_changed_files`** over `git status --short` when summarizing staged/unstaged work
- **PREFER `get_errors`** over running `go build`, `go vet`, or `golangci-lint run` purely for diagnostics
- **PREFER `list_code_usages`** over manual grep when tracing symbol usage
- **PREFER Pylance tools (`mcp_pylance_*`)** over ad-hoc Python shell commands for environment inspection
- Consult `docs/TOOLS.md` for the complete tool catalog before falling back to shell commands

### Git Operations (CRITICAL)
- **NEVER USE GitKraken MCP Server tools** (`mcp_gitkraken_*`) in Copilot chat sessions - GitKraken is ONLY for manual GUI operations
- **ALWAYS use terminal git commands** (git status, git add, git commit, git push) instead of GitKraken tools
