# Copilot Instructions

## Core Principles

- **Instruction files auto-discovered from** `.github/instructions/*.instructions.md`
- **Keep rules short** - one directive per line
- **Reference external & project resources** - avoid duplication of content
- **ALWAYS use built-in tools over terminal commands**
- **MUST: Do regular commits and pushes** to enable workflow monitoring and validation
- **MUST: ALWAYS bias towards accuracy, completeness, and correctness** - NEVER bias towards completing fast
- **MUST: Take the time required to do things correctly** - Time and token budgets are not constraints for Speckit work
- **MUST: Prioritize doing things right over doing things quickly** - Quality over speed is mandatory

## Quick Reference - CRITICAL Patterns

**Terminology**: See `01-01.terminology.instructions.md` for RFC 2119 keywords (MUST, SHOULD, MAY, CRITICAL, ALWAYS, NEVER)

**Continuous Work**: See `01-02.continuous-work.instructions.md` - NEVER STOP WORKING until user clicks "STOP" button

**Regression Prevention**:

- Format_go self-modification: See `03-01.coding.instructions.md`
- Windows Firewall exceptions: See `03-05.security.instructions.md`
- Git workflow patterns: See `05-02.git.instructions.md`
- CI-DAST lessons learned: See `05-03.dast.instructions.md`

## Instruction Files Reference

### 01-## Copilot Core

| File | Description |
|------|-------------|
| 01-01.terminology | RFC 2119 keywords (MUST, SHOULD, MAY, CRITICAL) |
| 01-02.continuous-work | LLM Agent continuous work directive (NEVER STOP) |
| 01-03.speckit | Speckit workflow quick reference |

### 02-## Architecture & Design

| File | Description |
|------|-------------|
| 02-01.architecture | Products & Services Architecture, microservices patterns, service federation |
| 02-02.service-template | Service template requirements (dual HTTPS, health checks) |
| 02-03.health-checks | Health check endpoint patterns (livez, readyz) |
| 02-04.bind-address | Public HTTPS bind address patterns (configurable, not hardcoded 0.0.0.0) |
| 02-05.hash-registry | Hash registry pepper and salt requirements (all 4 registries) |
| 02-06.versions | Minimum versions & consistency requirements |
| 02-07.cryptography | FIPS compliance, hash versioning, algorithm agility |
| 02-08.pki | PKI, CA, certificate management, CA/Browser Forum compliance |
| 02-09.observability | Observability & monitoring (OpenTelemetry, OTLP) |
| 02-10.database | Database & ORM patterns (PostgreSQL, SQLite, GORM) |
| 02-11.openapi | OpenAPI rules and patterns |

### 03-## Implementation

| File | Description |
|------|-------------|
| 03-01.coding | Coding patterns & standards (format_go, error handling) |
| 03-02.testing | Testing patterns & best practices (unit, integration, E2E) |
| 03-03.golang | Go project structure & conventions |
| 03-04.sqlite-gorm | SQLite configuration with GORM (WAL mode, busy timeout) |
| 03-05.security | Security patterns (Windows Firewall prevention, TLS) |
| 03-06.linting | Code quality & linting standards (golangci-lint, gremlins) |

### 04-## CI/CD

| File | Description |
|------|-------------|
| 04-01.github | CI/CD workflow (PostgreSQL service, cost efficiency) |
| 04-02.docker | Docker & Compose (multi-stage builds, secrets) |

### 05-## Tooling

| File | Description |
|------|-------------|
| 05-01.cross-platform | Cross-platform tooling (PowerShell, scripts, Docker pre-pull) |
| 05-02.git | Git workflow rules (commits, PRs, documentation) |
| 05-03.dast | DAST scanning (Nuclei, ZAP), CI-DAST lessons learned |

### 06-## Methodology

| File | Description |
|------|-------------|
| 06-01.evidence-based | Evidence-based task completion and validation |
| 06-02.speckit-detailed | Speckit methodology workflow integration & feedback loops |
| 06-03.anti-patterns | Common anti-patterns and mistakes to avoid |
