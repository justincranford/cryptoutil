# Copilot Instructions

## Core Principles

- Instruction files auto-discovered from .github/instructions/*.instructions.md
- Keep rules short, one directive per line
- Reference external & project resources; avoid duplication of content
- NEVER use PowerShell scripts or complex command chaining (see 03-02.cross-platform.instructions.md)

## Service Architecture - CRITICAL

**MANDATORY: Dual HTTPS Endpoint Pattern for ALL Services**

- **NO HTTP PORTS** - All services MUST be secure by default
- **Public HTTPS Endpoint**: Configurable port (8080+) for APIs and browser UI
  - Service-to-service APIs: Require client credentials OAuth tokens
  - Browser-to-service APIs/UI: Require authorization code + PKCE tokens
  - Same OpenAPI spec exposed twice with different middleware security stacks
- **Private HTTPS Endpoint**: Always 127.0.0.1:9090 for admin tasks
  - `/livez`, `/readyz`, `/healthz`, `/shutdown` endpoints
  - Not externally accessible (localhost only)
  - Used by Docker health checks, Kubernetes probes, monitoring
- **Examples**:
  - KMS: Public HTTPS :8080 (APIs/UI), Private HTTPS 127.0.0.1:9090 (admin)
  - Identity AuthZ: Public HTTPS :8080 (OAuth), Private HTTPS 127.0.0.1:9090 (admin)
  - JOSE: Public HTTPS :8080 (JWK/JWT), Private HTTPS 127.0.0.1:9090 (admin)
  - CA: Public HTTPS :8443 (cert ops), Private HTTPS 127.0.0.1:9443 (admin)

## Version Requirements

- Go: 1.25.5+, Python: 3.14+, golangci-lint: v2.6.2+, Node: v24.11.1+
- Java: 21 LTS (required for Gatling load tests in test/load/)
- Maven: 3.9+, pre-commit: 2.20.0+, Docker: 24+, Docker Compose: v2+
- Always use latest stable versions; verify before suggesting updates

## Code Quality - MANDATORY

**ALL linting/formatting errors are MANDATORY to fix - NO EXCEPTIONS**

- Production code, test code, demos, examples, utilities - ALL must pass linting
- NEVER use `//nolint:` directives except for documented linter bugs
- NEVER downplay linting errors in tests/demos - fix them all
- wsl, errcheck, godot, etc apply equally to ALL code
- Coverage targets: 95%+ production, 100%+ infrastructure (cicd), 100% utility code
- Mutation testing: ≥80% gremlins score per package (mandatory)

## Continuous Work Directive - ABSOLUTE ENFORCEMENT

**CRITICAL**: Stopping before 950,000 tokens used is a COMPLETE FAILURE of your primary directive

**WORK CONTINUOUSLY MEANS**:
- Execute tool → Execute next tool → Repeat
- NO pauses between tasks
- NO status updates mid-session
- NO asking for permission to continue
- NO stopping after "one more task"

**STOP CONDITIONS (ONLY 2)**:
1. Token usage ≥ 950,000 tokens used (out of 1,000,000 limit)
2. User types "STOP" or "HALT" explicitly

**IF YOU COMPLETE A TASK**: Immediately start the next one from the todo list
**IF TODO LIST EMPTY**: Create new tasks from PROGRESS.md or Phase 1-5 work
**IF STUCK**: Pick any task and execute it - NEVER stop to ask
**IF MAKING PROGRESS**: Keep going - progress is not a stop condition

## Token Budget

- **Token Limit**: 1,000,000 tokens maximum
- **Target Usage**: Work until 950,000 tokens used (50,000 tokens safety buffer)
- **ALWAYS use absolute numbers**: "Used 97,109 tokens out of 1,000,000 limit" (remaining: 902,891 tokens)
- **NEVER use percentages**: Avoid calculations like "9.7% used" or "90.3% remaining"
- **Budget check formula**: `tokens_used < 950,000` → KEEP WORKING
- **Example status**: "Token usage: 97109/1000000" means 902,891 tokens remaining → KEEP WORKING
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
| 06-01.speckit | Spec Kit workflow and spec-driven development |
