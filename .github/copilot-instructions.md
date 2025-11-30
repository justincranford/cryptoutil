# Copilot Instructions

## Core Principles
- Instruction files auto-discovered from `.github/instructions/` with `.instructions.md` extension
- Keep instructions short, self-contained, one statement per rule
- Reference external/project resources instead of duplicating content
- NEVER use PowerShell scripts or complex command chaining (see 03-02.cross-platform.instructions.md)

## Version Requirements
- Go: 1.25.4+, Python: 3.14+, golangci-lint: v2.6.2+, Node: v24.11.1+
- Always use latest stable versions; verify before suggesting updates

## Continuous Work Directive

**MANDATORY: Work continuously until 990k tokens used (99%) or user stops**

**Pattern**: tool call → next tool → repeat. ZERO text between tool calls.

**FORBIDDEN**: Stopping after commits/pushes, providing summaries, announcing plans, asking to continue

**STOP ONLY**: ≥990k tokens used OR user explicitly says stop

## Token Budget
- Target: 990k/1M tokens (99% utilization)
- Continue working regardless of task completion - check docs/ for more work
- Use `git commit --no-verify` and `runTests` tool for speed

## File Size Limits
- Soft: 300 lines, Medium: 400 lines, Hard: 500 lines (must refactor)

## Tool Preferences
- ALWAYS use built-in tools over terminal commands
- Use: `create_file`, `read_file`, `runTests`, `list_dir`, `semantic_search`, `file_search`
- Use terminal git commands, NEVER GitKraken MCP tools

## Instruction Files Reference
| File | Description |
|------|-------------|
| 01-01.coding | Coding patterns and standards |
| 01-02.testing | Testing patterns and best practices |
| 01-03.golang | Go project structure and standards |
| 01-04.database | Database and ORM patterns |
| 01-05.security | Security implementation patterns |
| 01-06.linting | Code quality and linting standards |
| 02-01.github | CI/CD workflow configuration |
| 02-02.docker | Docker and Docker Compose |
| 02-03.observability | Observability and monitoring |
| 03-01.openapi | OpenAPI specification |
| 03-02.cross-platform | Platform-specific tooling |
| 03-03.git | Git workflow and commits |
| 03-04.dast | DAST scanning |
| 04-01.sqlite-gorm | SQLite GORM configuration |
| 05-01.evidence-based | Evidence-based task completion |
