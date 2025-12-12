# Copilot Instructions

## Core Principles

- Instruction files auto-discovered from .github/instructions/*.instructions.md
- Keep rules short, one directive per line
- Reference external & project resources; avoid duplication of content

## Service Architecture - CRITICAL

**MANDATORY: Dual HTTPS Endpoint Pattern for ALL Services**

- **HTTPS ONLY** - All first-party products and services MUST use TLS 1.3 by default; never use HTTP
- **HTTPS Certificates** - All first-party services MUST support configuration to pick TLS cert chain and private key source:
  - Auto-generated TLS server cert chain all the way through to a private root CA,
  - Docker Secrets for Private Keys, and File Paths or PEM data for all of TLS Server cert and Trusted CA certs
- **Private HTTPS Endpoint**: Configurable HTTPS port for private APIs
  - Default ports: 0 for dynamic, 9090+ for static
  - /admin/v1/** prefix for all endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz`, `/admin/v1/shutdown`
  - Not externally accessible; IPv4 127.0.0.1 only, never IPv6, never localhost
  - Used by Docker health checks, Kubernetes probes, monitoring, testing
- **Public HTTPS Endpoint**: Configurable HTTPS port for public APIs/UI
  - Default ports: 0 for dynamic, 8080+ for static
  - /service/** prefix for non-browser clients: Enforce with Authentication token (access token in HTTP Authorization header, TLS client cert, or both), and Authorization (/service/ path prefix), and non-browser client middleware (e.g. IP Allowlist, Rate Limiting)
  - /browser/** prefix for browser-based clients: Enforce with Authentication token (session token in HTTP Cookie header, TLS client cert, or both), and Authorization (/browser/ path prefix), and browser-based client middleware (e.g. IP Allowlist, Rate Limiting, CSRF, CORS, CSP, etc)
  - /service access tokens must be obtained via OAuth 2.0 Client Authorization Flow; these are bearer tokens in HTTP Authorization header, and are only authorized to access /service/ request path prefix
  - /browser access tokens must be obtained via OAuth 2.0 Authorization Code + PKCE flow; these are session tokens in HTTP Cookie header, and are only authorized to access /browser/ request path prefix
  - /browser and /service support using the same reusable OpenAPI spec, but under different prefixes, for API consistency
  - /browser also supports UI-only request paths for browser-based clients only
- **Examples**:
  - KMS:            Public HTTPS APIs/UI :8080 (0.0.0.0:8080 in container), Private HTTPS APIs (127.0.0.1:9090 in container)
  - Identity AuthZ: Public HTTPS APIs/UI :8180 (0.0.0.0:8080 in container), Private HTTPS APIs (127.0.0.1:9090 in container)
  - Identity IdP:   Public HTTPS APIs/UI :8181 (0.0.0.0:8080 in container), Private HTTPS APIs (127.0.0.1:9090 in container)
  - Identity RS:    Public HTTPS APIs/UI :8182 (0.0.0.0:8080 in container), Private HTTPS APIs (127.0.0.1:9090 in container)
  - Identity RP:    Public HTTPS APIs/UI :8183 (0.0.0.0:8080 in container), Private HTTPS APIs (127.0.0.1:9090 in container)
  - JOSE:           Public HTTPS APIs/UI :8280 (0.0.0.0:8080 in container), Private HTTPS APIs (127.0.0.1:9090 in container)
  - CA:             Public HTTPS APIs/UI :8380 (0.0.0.0:8080 in container), Private HTTPS APIs (127.0.0.1:9090 in container)

## Version Requirements

- Go: 1.25.5+, Python: 3.14+, golangci-lint: v2.6.2+, Node: v24.11.1+
- Java: 21 LTS (required for Gatling load tests in test/load/)
- Maven: 3.9+, pre-commit: 2.20.0+, Docker: 24+, Docker Compose: v2+
- **CGO_ENABLED=0 MANDATORY** - CGO is BANNED except for race detector (Go toolchain limitation)
- **EXCEPTION**: Race detector (`-race`) requires CGO_ENABLED=1 (Go toolchain limitation)
- Always use latest stable versions; verify before suggesting updates

## Code Quality - MANDATORY

**ALL linting/formatting errors are MANDATORY to fix - NO EXCEPTIONS**

- Production code, test code, demos, examples, utilities - ALL must pass linting
- NEVER use `//nolint:` directives except for documented linter bugs
- NEVER downplay linting errors in tests/demos - fix them all
- wsl, errcheck, godot, etc apply equally to ALL code
- Coverage targets: 95%+ production, 100%+ infrastructure (cicd), 100% utility code
- Mutation testing: ≥80% gremlins score per package (mandatory)

## Race Condition Prevention - CRITICAL

- NEVER write to parent scope variables in parallel sub-tests
- NEVER use t.Parallel() with global state manipulation (os.Stdout, env vars)
- ALWAYS use inline assertions: `require.NoError(t, resp.Body.Close())`
- ALWAYS create fresh test data per test case (new sessions, UUIDs)
- ALWAYS protect shared maps/slices with sync.Mutex or sync.Map
- Detection: `go test -race -count=2` (requires CGO_ENABLED=1)
- Details: .github/instructions/01-02.testing.instructions.md

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
**IF TODO LIST EMPTY**: Create new tasks from implement/DETAILED.md timeline or TASKS.md
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
