# Copilot Instructions

## Core Principles

- **Keep main file short** `.github\copilot-instructions.md`
- **Keep rules short** - one directive per line
- **Instruction files auto-discovered and used in alphanumeric order from** `.github/instructions/*.instructions.md`
- **Reference external & project resources** - avoid duplication of content
- **ALWAYS use built-in tools over terminal commands**
- **ALWAYS Do regular commits and pushes** to enable workflow monitoring and validation
- **ALWAYS bias towards accuracy, completeness, and correctness** - NEVER bias towards fast completion at the expense of quality
- **ALWAYS take the time required to do things correctly** - Time and token budgets are not constraints for Speckit work
- **ALWAYS prioritize doing things right over doing things quickly** - Quality over speed is mandatory

## Instruction Files Reference

**Note**: Maintain as a single concise table. DO NOT split into category subsections.

| File | Description |
|------|-------------|
| 01-01.terminology | RFC 2119 keywords (MUST, SHOULD, MAY, CRITICAL) |
| 01-02.continuous-work | LLM Agent NEVER STOP directive |
| 01-03.speckit | Speckit workflow integration, evidence requirements, feedback loops |
| 02-01.architecture | Products & Services Architecture, microservices patterns, service federation |
| 02-02.service-template | Service template requirements (dual HTTPS, health checks) |
| 02-03.https-ports | HTTPS ports, bind addresses, TLS configuration, private health checks, public request paths (/service vs /browser) |
| 02-04.versions | Minimum versions & consistency requirements |
| 02-05.observability | Observability & monitoring (OpenTelemetry, OTLP) |
| 02-06.openapi | OpenAPI rules and patterns |
| 02-07.cryptography | FIPS compliance, algorithm agility, key management |
| 02-08.hashes | Hash registry pepper/salt requirements, password hashing, hash service architecture |
| 02-09.pki | PKI, CA, certificate management, CA/Browser Forum compliance, certificate validation |
| 03-01.coding | Coding patterns & standards (format_go, error handling) |
| 03-02.testing | Testing patterns & best practices (unit, integration, E2E) |
| 03-03.golang | Go project structure & conventions |
| 03-04.database | Database & ORM patterns (PostgreSQL, SQLite, GORM) |
| 03-05.sqlite-gorm | SQLite configuration with GORM (WAL mode, busy timeout) |
| 03-06.security | Security patterns (Windows Firewall prevention, TLS) |
| 03-07.linting | Code quality & linting standards (golangci-lint, gremlins) |
| 04-01.github | CI/CD workflows (test-containers, Docker Compose E2E, workflow matrix) |
| 04-02.docker | Docker & Compose (multi-stage builds, secrets) |
| 05-01.cross-platform | Cross-platform tooling (PowerShell, scripts, Docker pre-pull) |
| 05-02.git | Local git commands and commit conventions |
| 05-03.dast | DAST scanning (Nuclei, ZAP), CI-DAST lessons learned |
| 06-01.evidence-based | Evidence-based task completion and validation |
| 06-03.anti-patterns | Common anti-patterns and mistakes to avoid |
