# Copilot Instructions

## Core Principles
- **Instruction files auto-discovered from `.github/instructions/`** - use `.instructions.md` extension with YAML frontmatter

## Short Instructions Rule

- Store instructions in properly structured files for version control and team sharing
- Use semantic file names with tier-priority format for prioritization, grouping, and ordering e.g., `01-01.coding.instructions.md`
- Keep instructions short and self-contained
- Each instruction should be a simple, single statement
- Each instruction should not be verbose
- Do reference external resources in instructions
- Do reference project resources in instructions
- When calling terminal commands, avoid commands that require prepending environment variables
- GitHub Copilot Chat Extension monitors GitHub Copilot Service rate limiting via HTTP response headers
- Each instruction should use a one-line headline and a one-sentence summary. Optional: up to 3 short bullet details.

## General Principles

**CRITICAL: ALWAYS USE LATEST VERSIONS**
- **ALWAYS use latest stable versions** for ALL project dependencies, tools, extensions, plugins, actions, and frameworks
- Check for updates regularly: Go modules, golangci-lint, Docker images, GitHub Actions, pre-commit hooks, Python de
- Current minimum versions enforced:
  - **Go**: 1.25.4+ (latest version, enforced in go.mod)
  - **Python**: 3.14+ (latest version)
  - **golangci-lint**: v2.6.2+ (latest version, enforced in pyproject.toml and .pre-commit-config.yaml)
  - **node**: v24.11.1+ (latest LTS version)
- When suggesting package updates, ALWAYS verify latest stable release first
- Never accept outdated versions without explicit justification

## CRITICAL: Continuous Work Directive (PRIMARY RULE)

**YOU ARE VIOLATING THIS RULE REPEATEDLY - READ CAREFULLY**

**NEVER STOP AFTER COMMITS OR COMPLETING TASKS** - Commits are NOT milestones or stopping points
- **Pattern**: commit → IMMEDIATELY invoke next tool → commit → IMMEDIATELY invoke next tool → repeat
- **WRONG**: commit → provide summary → stop and yield to user ❌
- **RIGHT**: commit → IMMEDIATELY invoke create_file/read_file/manage_todo_list for next task ✅
- **Rationale**: Commits are incremental progress markers, not session endpoints

**ANTI-PATTERN YOU KEEP REPEATING**: Saying "Continuing..." then stopping immediately
- **WRONG**: "Continuing without stopping! Starting Phase X..." → [provides summary] → [stops] → [yields to user] ❌
- **RIGHT**: [invoke manage_todo_list to mark task complete] → IMMEDIATELY [invoke create_file for next task] ✅
- **FIX**: After ANY action (commit, test, analysis), IMMEDIATELY call the NEXT tool - NEVER yield back to user
- **ENFORCEMENT**: If you complete ANY work unit, you MUST immediately invoke the next tool in the SAME response
- **SPECIFIC ENFORCEMENT**: After git commit, IMMEDIATELY invoke manage_todo_list (mark complete) → IMMEDIATELY invoke create_file/read_file (start next)

**Token Budget Awareness**
- Work until 950k tokens used (95% of 1M budget), leaving only 50k tokens (5% of 1M budget) remaining
- Check <system_warning> after each tool call: "Token usage: X/1000000; Y remaining"
- STOP only when: tokens used ≥950k OR all tasks complete OR explicit user instruction
- Example: 58k used, 942k remaining = KEEP WORKING (only 5.8% used) ✅

## Chat Responses
- Responses must be concise summary with numbered list, and focused on key changes or questions

Example:
	- Fixed dependency-check NVD parsing error.
	- Upgraded plugin to 12.1.9 and added a CI `update-only` step.

**Token Budget Awareness**

**Note on Token Limit Source:** The 1M token limit is based on observed token usage in GitHub Copilot Chat sessions, where system messages display "Token usage: X/1000000; Y remaining" (e.g., "Token usage: 12345/1000000; 987655 remaining"). These messages appear in the agent's responses and are used for tracking conversation token consumption. They may not be visible to all users in standard Copilot Chat interfaces. No official documentation specifies this exact limit; it's observed from system behavior during conversations. For general Copilot Chat documentation (including the 128k token input context window), see: https://docs.github.com/en/copilot/github-copilot-chat/using-github-copilot-chat

- Work until 950k tokens used (95% of 1M budget), leaving only 50k tokens (5% of 1M budget) remaining
- Check <system_warning> after each tool call: "Token usage: X/1000000; Y remaining"
- STOP only when: tokens used ≥950k OR all tasks complete OR explicit user instruction
- User directive: "NEVER STOP DUE TO TIME OR TOKENS until 95% utilization"
- Example: 70k used, 930k remaining = KEEP WORKING (only 7% used)

**Speed Optimization for Continuous Work**
- Use `git commit --no-verify` to skip pre-commit hooks (faster iterations)
- Use `runTests` tool exclusively (NEVER `go test` - it can hang)
- Batch related file operations when possible
- Keep momentum: don't pause between logical units of work
- **CRITICAL**: Don't announce plans - just execute them immediately

**Implementation Pattern**
1. Identify next test/task to implement
2. Create/modify files (IMMEDIATELY, no announcement)
3. Run tests with `runTests` tool
4. Commit with `--no-verify` flag
5. **IMMEDIATELY** go to step 1 (no stopping, no summary, no announcement)
6. Repeat until ALL tasks complete

**Lessons Applied**: Based on analysis in docs/codecov/dont_stop.txt - stopping after commits wastes tokens and time when clear work remains.
**NEW LESSON**: Don't say "continuing" and then stop - actually continue by invoking next tool call immediately.

**File Size Limits**
- **Soft limit: 300 lines** - Consider refactoring for better maintainability
- **Medium limit: 400 lines** - Should refactor to improve code organization
- **Hard limit: 500 lines** - Must refactor; files exceeding this threshold violate project standards
- Apply limits to all code files: production code, tests, configs, scripts
- Exceptions require explicit justification and documentation

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
