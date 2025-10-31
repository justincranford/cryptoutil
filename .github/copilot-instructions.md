## CI/CD Cost Efficiency
- Minimize workflow run frequency (trigger only on relevant changes)
- Avoid unnecessary matrix builds; keep job matrices minimal
- Use dependency caching to reduce billed minutes
- Skip jobs for docs-only or trivial changes
- Use job filters/conditionals to avoid wasteful runs
# Copilot Instructions


## Core Principles
- Follow README and instructions files
- **ALWAYS check go.mod for correct Go version** before using Go in Docker, CI/CD, or tools
- Refer to architecture and usage examples in README
- **Instruction files auto-discovered from `.github/instructions/`** - use `.instructions.md` extension with YAML frontmatter

## Nuclei Vulnerability Scanning

**ALWAYS start cryptoutil services before running nuclei scans:**
```sh
# Prerequisites - start services first
docker compose -f ./deployments/compose/compose.yml down -v
docker compose -f ./deployments/compose/compose.yml up -d

# Verify services are ready (CI/CD workflow context only - NOT for local chat commands)
# Local alternative: docker compose exec cryptoutil-sqlite wget --no-check-certificate -q -O - https://127.0.0.1:8080/ui/swagger/doc.json
curl -k https://localhost:8080/ui/swagger/doc.json  # SQLite instance (CI/CD only)
curl -k https://localhost:8081/ui/swagger/doc.json  # PostgreSQL instance 1 (CI/CD only)
curl -k https://localhost:8082/ui/swagger/doc.json  # PostgreSQL instance 2 (CI/CD only)
```

**Manual Nuclei Scan Commands:**
```sh
# Quick security scan (info/low severity, fast)
nuclei -target https://localhost:8080/ -severity info,low
nuclei -target https://localhost:8081/ -severity info,low
nuclei -target https://localhost:8082/ -severity info,low

# Comprehensive security scan (medium/high/critical severity)
nuclei -target https://localhost:8080/ -severity medium,high,critical
nuclei -target https://localhost:8081/ -severity medium,high,critical
nuclei -target https://localhost:8082/ -severity medium,high,critical

# Targeted scans by vulnerability type
nuclei -target https://localhost:8080/ -tags cves,vulnerabilities
nuclei -target https://localhost:8080/ -tags security-misconfiguration
nuclei -target https://localhost:8080/ -tags exposure,misc

# Performance-optimized scans
nuclei -target https://localhost:8080/ -c 25 -rl 100 -severity high,critical
```

**Service Endpoints:**
- **cryptoutil-sqlite**: `https://localhost:8080/` (SQLite backend)
- **cryptoutil-postgres-1**: `https://localhost:8081/` (PostgreSQL backend)
- **cryptoutil-postgres-2**: `https://localhost:8082/` (PostgreSQL backend)

**Template Management:**
```sh
# Update nuclei templates
nuclei -update-templates

# List available templates
nuclei -tl

# Check template version
nuclei -templates-version
```

## Docker Compose Cross-Platform Path Requirements

- **CRITICAL: NEVER use absolute paths in `deployments/compose/compose.yml`**
- **ALWAYS use relative paths** for all file references (volumes, secrets, configs)
- Absolute Windows paths (`C:\...`) break cross-platform compatibility with:
  - GitHub Actions Ubuntu runners
  - `act` local workflow testing on Windows/WSL
  - Docker Compose path resolution in Linux containers
- Relative paths resolve correctly from the compose file's directory on all platforms
- Example corrections:
  - ❌ BAD: `file: C:\Dev\Projects\cryptoutil\deployments\compose\postgres\postgres_username.secret`
  - ✅ GOOD: `file: ./postgres/postgres_username.secret`
- This applies to ALL path references in compose.yml: secrets, volumes, configs, dockerfiles

## Copilot testing guidance for cicd utility

- When adding or updating the cicd utility (`internal/cmd/cicd/cicd.go`), always implement programmatic tests in `internal/cmd/cicd/cicd_test.go` that:
  - write any generated code to a temporary file (use the test's TempDir),
  - run the lint/check function against that temporary file, and
  - read and assert the temporary file contents/results programmatically.
  This avoids interactive prompts or external side-effects during Copilot-assisted sessions.

- Avoid asking Copilot to output or run ephemeral shell commands that create/update repository files via chat (for example: piping PowerShell Get-Content replace commands into Set-Content). Those patterns often trigger unwanted prompts and premium LLM watch requests. Instead, prefer programmatic edits via tests or using the repository's scripted tools.

## Workflow Testing with cmd/workflow

- Use `go run ./cmd/workflow` to execute GitHub Actions workflows locally with act
- The workflow runner is implemented in `internal/cmd/workflow/workflow.go` with cmd entry point at `cmd/workflow/main.go`
- Common usage patterns:
  ```bash
  go run ./cmd/workflow -workflows=e2e,dast
  go run ./cmd/workflow -workflows=quality -dry-run
  go run ./cmd/workflow -workflows=load -inputs="load_profile=quick"
  go run ./cmd/workflow -list
  ```
- Available workflows: e2e, dast, sast, robust, quality, load

## Pre-commit Hook Guidelines

- **NEVER use shell commands** (`sh`, `bash`, `powershell`) in pre-commit configurations
- Use cross-platform tools directly (e.g., `go`, `python`) or pre-commit's built-in hooks
- Ensure all hooks work on Windows, Linux, and macOS without shell dependencies

## Instruction File Structure

**T#-P# Naming Convention**: Tier-Priority format for explicit load order
- **T#** = Tier number (priority level)
- **P#** = Priority within tier (alphabetical load order)

### Tier 1: Foundation (Always Loads First)
| File | Pattern | Description |
|------|---------|-------------|
| T1-P1-copilot-customization | ** | Git ops, terminal patterns, curl/wget rules, fuzz testing, conventional commits, TODO maintenance |

### Tier 2: Core Development (Slots 2-6)
| File | Pattern | Description |
|------|---------|-------------|
| T2-P1-code-quality | ** | Linter compliance, wsl/godot rules, resource cleanup, pre-commit docs |
| T2-P2-testing | ** | Test patterns, dependency mgmt, file organization, UUIDv7 concurrency |
| T2-P3-architecture | ** | Layered arch, config patterns, lifecycle, factory patterns, atomic ops |
| T2-P4-security | ** | Key hierarchy, IP allowlisting, rate limiting, TLS, secrets management |
| T2-P5-docker | **/*.yml | Compose config, healthchecks, Docker secrets, OTEL forwarding |

### Tier 3: High Priority (Slots 7-11)
| File | Pattern | Description |
|------|---------|-------------|
| T3-P1-crypto | ** | NIST FIPS 140-3 algorithms, keygen patterns, cryptographic operations |
| T3-P2-cicd | .github/workflows/*.yml | Workflow configuration, Go version consistency, service connectivity |
| T3-P3-observability | ** | OpenTelemetry integration, OTLP protocols, telemetry forwarding |
| T3-P4-database | ** | GORM ORM patterns, migrations, PostgreSQL/SQLite support |
| T3-P5-go-standards | **/*.go | Import aliases, dependencies, formatting (gofumpt), conditionals |

### Tier 4: Medium Priority (Slots 12-15)
| File | Pattern | Description |
|------|---------|-------------|
| T4-P1-specialized-testing | ** | Act workflow testing, localhost vs 127.0.0.1 patterns |
| T4-P2-project-config | ** | OpenAPI specs, magic values, linting exclusions |
| T4-P3-platform-specific | scripts/** | PowerShell/Bash scripts, Docker image pre-pull |
| T4-P4-specialized-domains | ** | CA/Browser Forum compliance, project layout, PR descriptions |

## Cross-Reference Guide

**Docker Compose**: T2-P5-docker (primary) → T2-P3-architecture, T3-P3-observability, T2-P4-security, T2-P2-testing  
**CI/CD Workflows**: T3-P2-cicd (primary) → T2-P5-docker, T4-P1-specialized-testing, T2-P2-testing  
**Security**: T2-P4-security (primary) → T2-P5-docker, T3-P1-crypto, T4-P4-specialized-domains  
**Observability**: T3-P3-observability (primary) → T2-P5-docker, T2-P3-architecture  
**Testing**: T2-P2-testing (primary) → T2-P5-docker, T4-P1-specialized-testing, T2-P1-code-quality  
**Go Code**: T3-P5-go-standards (primary) → T4-P2-project-config, T2-P1-code-quality, T4-P4-specialized-domains
