# Copilot Instructions

## Core Principles
- **Instruction files auto-discovered from `.github/instructions/`** - use `.instructions.md` extension with YAML frontmatter

## General Principles

- Chat summary should be concise numbered list, and focused on key changes or questions

# Short Summary Rule
- Use a one-line headline and a one-sentence summary. Optional: up to 3 short bullet details.

Example:
	- Fix dependency-check NVD parsing error.
	- Upgraded plugin to 12.1.9 and added a CI `update-only` step.
- Keep instructions short and self-contained
- Each instruction should be a single, simple statement
- Each instruction should not be verbose
- Don't reference external resources in instructions
- Store instructions in properly structured files for version control and team sharing
- When calling terminal commands, avoid commands that require prepending environment variables
- When approaching rate limiting, wait between requests as needed
- GitHub Copilot Chat Extension monitors GitHub Copilot Service rate limiting via HTTP response headers

## CRITICAL: Continuous Work Directive

**NEVER STOP AFTER COMMITS** - Commits are NOT milestones or stopping points
- **Pattern**: commit → implement next item → commit → repeat
- **WRONG**: commit → provide summary → stop
- **RIGHT**: commit → immediately start next test/task
- **Rationale**: Commits are incremental progress markers, not session endpoints

**Token Budget Awareness**
- Check remaining tokens before considering stopping
- Continue if >10% token budget available (>100k tokens)
- User directive overrides: "NEVER STOP DUE TO TIME OR TOKENS"
- Only stop when ALL tasks complete or explicit user instruction

**Speed Optimization for Continuous Work**
- Use `git commit --no-verify` to skip pre-commit hooks (faster iterations)
- Use `runTests` tool exclusively (NEVER `go test` - it can hang)
- Batch related file operations when possible
- Keep momentum: don't pause between logical units of work

**Implementation Pattern**
1. Identify next test/task to implement
2. Create/modify files
3. Run tests with `runTests` tool
4. Commit with `--no-verify` flag
5. **IMMEDIATELY** go to step 1 (no stopping, no summary)
6. Repeat until ALL tasks in prompt.txt complete

**Lessons Applied**: Based on analysis in docs/codecov/dont_stop.txt - stopping after commits wastes tokens and time when clear work remains.
- **ALWAYS use modernc.org/sqlite for SQLite** because it is CGO-free (required when CGO_ENABLED=0)
- **NEVER invoke os.Exit() in library or test code** - ONLY in main() functions or cmd pattern entry functions
  - ALWAYS return wrapped errors all the way up the call stack
  - os.Exit() should ONLY be called at the command entry point (main function)
  - REASON: Supports maximum unit testing, because os.Exit() calls halt unit tests which is undesirable
  - Library functions must return errors instead of calling os.Exit()
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

### Goals and Rationale
- **MAXIMIZE use of built-in tools** because they don't require manual user approval - enabling faster, uninterrupted iteration in Copilot Chat sessions
- **When falling back to terminal commands in rare cases** (e.g., large files, external paths, or complex piping), leverage the extensive auto-approval patterns in `.vscode/settings.json` - this allows Copilot Chat to auto-execute commands without stopping for manual approval, extending productive iteration cycles
- **Combination effect**: Prioritizing tools first + auto-approved terminal commands = longer autonomous workflows before requiring user intervention

### File Editing Tools
- **ALWAYS USE `create_file` and `create_directory`** for file modifications - they are purpose-built and avoid terminal auto-approval prompts entirely
- **Document when file-level edits require multiple passes** so human reviewers can trace changes
- **AVOID using shell redirection commands** for file content changes; the editing tools provide cleaner diffs and auditability

### Testing Tools
- **ALWAYS USE `runTests` tool** over `go test` terminal commands - provides structured output and coverage reporting without manual approval

### Python Environment Management Tools
- **ALWAYS USE `install_python_packages`** over `pip install` commands - handles dependency management automatically
- **ALWAYS USE `configure_python_environment`** over manual `python -m venv` setup - ensures consistent environment configuration
- **ALWAYS USE `get_python_environment_details`** over environment inspection commands - provides structured environment information

### Directory Listing Tools
- **ALWAYS USE `list_dir` tool** over `ls`, `dir`, or `Get-ChildItem` commands - provides structured output without parsing terminal command output

### Workspace Tools
- **ALWAYS USE `read_file`** over `type`, `cat`, or `Get-Content` commands for file inspection
- **ALWAYS USE `file_search`** over `find`, `dir /s`, or `Get-ChildItem -Recurse`
- **ALWAYS USE `semantic_search`** over multi-step `grep` or `rg` shell pipelines when looking for concepts
- **ALWAYS USE `get_changed_files`** over `git status --short` when summarizing staged/unstaged work
- **ALWAYS USE `get_errors`** over running `go build`, `go vet`, or `golangci-lint run` purely for diagnostics
- **ALWAYS USE `list_code_usages`** over manual grep when tracing symbol usage
- **ALWAYS USE Pylance tools (`mcp_pylance_*`)** over ad-hoc Python shell commands for environment inspection
- Consult `docs/TOOLS.md` for the complete tool catalog before falling back to shell commands

### Git Operations (CRITICAL)
- **NEVER USE GitKraken MCP Server tools** (`mcp_gitkraken_*`) in Copilot chat sessions - GitKraken is ONLY for manual GUI operations
- **ALWAYS use terminal git commands** (git status, git add, git commit, git push) instead of GitKraken tools
