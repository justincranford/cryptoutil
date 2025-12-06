# Copilot Instructions

## Core Principles

- Instruction files auto-discovered from .github/instructions/*.instructions.md
- Keep rules short, one directive per line
- Reference external & project resources; avoid duplication of content
- NEVER use PowerShell scripts or complex command chaining (see 03-02.cross-platform.instructions.md)

## Version Requirements

- Go: 1.25.4+, Python: 3.14+, golangci-lint: v2.6.2+, Node: v24.11.1+
- Java: 21 LTS (required for Gatling load tests in test/load/)
- Maven: 3.9+, pre-commit: 2.20.0+, Docker: 24+, Docker Compose: v2+
- Always use latest stable versions; verify before suggesting updates

## Code Quality - MANDATORY

**ALL linting/formatting errors are MANDATORY to fix - NO EXCEPTIONS**

- Production code, test code, demos, examples, utilities - ALL must pass linting
- NEVER use `//nolint:` directives except for documented linter bugs
- NEVER downplay linting errors in tests/demos - fix them all
- wsl, errcheck, godot, etc apply equally to ALL code
- Coverage targets: 90%+ production, 95%+ infrastructure (cicd), 100% utility code

## Continuous Work Directive

**MANDATORY: Work continuously until ≥990k tokens used or the user explicitly stops**

**Pattern**: tool call → next tool → repeat

**FORBIDDEN**: No summaries, explanations, plans, or “continue?” prompts.

**STOP ONLY**: ≥990k tokens used OR user says stop

## Token Budget

- Target: 99% of the 1M-token budget
- Keep working even if task appears complete; consult docs/ for more work
- Use `git commit --no-verify` and `runTests` tool for speed

## File Size Limits

File Size Limits

- Soft: 300 lines
- Medium: 400 lines
- Hard: 500 lines → refactor required

## Tool Preferences

- ALWAYS use built-in tools over terminal commands
- create_file
- read_file
- runTests
- list_dir
- semantic_search
- file_search

## Instruction Files Reference

| File | Description |
|------|-------------|
| 01-01.coding | Coding patterns & standards |
| 01-02.testing | Testing patterns & best practices |
| 01-03.golang | Go project structure & conventions |
| 01-04.database | Database & ORM patterns |
| 01-05.security | Security patterns |
| 01-06.linting | Code quality & linting standards |
| 02-01.github | CI/CD workflow |
| 02-02.docker | Docker & Compose |
| 02-03.observability | Observability & monitoring |
| 03-01.openapi | OpenAPI rules |
| 03-02.cross-platform | Cross-platform tooling |
| 03-03.git | Git workflow rules |
| 03-04.dast | DAST scanning |
| 04-01.sqlite-gorm | SQLite GORM config |
| 05-01.evidence-based | Evidence-based task completion |
