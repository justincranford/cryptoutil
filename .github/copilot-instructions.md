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
| [01-01.terminology](.github/instructions/01-01.terminology.instructions.md) | RFC 2119 keywords (MUST, SHOULD, MAY, CRITICAL) |
| [01-02.continuous-work](.github/instructions/01-02.continuous-work.instructions.md) | LLM Agent continuous work directive - NEVER STOP until user clicks stop |
| [01-03.speckit](.github/instructions/01-03.speckit.instructions.md) | SpecKit methodology |
| [02-01.architecture](.github/instructions/02-01.architecture.instructions.md) | Products and services architecture patterns |
| [02-02.service-template](.github/instructions/02-02.service-template.instructions.md) | Service template requirements and patterns for all cryptoutil services |
| [02-03.https-ports](.github/instructions/02-03.https-ports.instructions.md) | HTTPS ports, bind addresses, TLS addresses, CORS Origins, TLS configuration, and request paths for public and admin endpoints |
| [02-04.versions](.github/instructions/02-04.versions.instructions.md) | Instructions for version requirements |
| [02-05.observability](.github/instructions/02-05.observability.instructions.md) | Instructions for observability and monitoring implementation |
| [02-06.openapi](.github/instructions/02-06.openapi.instructions.md) | Instructions for OpenAPI |
| [02-07.cryptography](.github/instructions/02-07.cryptography.instructions.md) | Instructions for cryptographic patterns, FIPS compliance, hash versioning, and algorithm agility |
| [02-08.hashes](.github/instructions/02-08.hashes.instructions.md) | Hash registry pepper/salt requirements, password hashing patterns, and hash service architecture |
| [02-09.pki](.github/instructions/02-09.pki.instructions.md) | Instructions for PKI, CA, certificate management, and CA/Browser Forum compliance |
| [02-10.authn](.github/instructions/02-10.authn.instructions.md) | Authentication and authorization tactical implementation patterns |
| [03-01.coding](.github/instructions/03-01.coding.instructions.md) | Instructions for coding patterns and standards |
| [03-02.testing](.github/instructions/03-02.testing.instructions.md) | Instructions for testing |
| [03-03.golang](.github/instructions/03-03.golang.instructions.md) | Instructions for Go project structure, architecture, and coding standards |
| [03-04.database](.github/instructions/03-04.database.instructions.md) | Instructions for database operations and ORM patterns |
| [03-05.sqlite-gorm](.github/instructions/03-05.sqlite-gorm.instructions.md) | Instructions for SQLite configuration with GORM, transaction patterns, and concurrent operations |
| [03-06.security](.github/instructions/03-06.security.instructions.md) | Instructions for security implementation patterns |
| [03-07.linting](.github/instructions/03-07.linting.instructions.md) | Instructions for code quality, linting, and maintenance standards |
| [04-01.github](.github/instructions/04-01.github.instructions.md) | Instructions for CI/CD workflow configuration, service connectivity verification, and diagnostic logging |
| [04-02.docker](.github/instructions/04-02.docker.instructions.md) | Instructions for Docker and Docker Compose configuration |
| [05-01.cross-platform](.github/instructions/05-01.cross-platform.instructions.md) | Instructions for platform-specific tooling: PowerShell, scripts, Docker pre-pull |
| [05-02.git](.github/instructions/05-02.git.instructions.md) | Instructions for local Git commands and commit conventions |
| [05-03.dast](.github/instructions/05-03.dast.instructions.md) | Instructions for Dynamic Application Security Testing (DAST): Nuclei scanning, ZAP testing, and workflow execution |
| [06-01.evidence-based](.github/instructions/06-01.evidence-based.instructions.md) | Instructions for evidence-based task completion and validation |
| [06-02.anti-patterns](.github/instructions/06-02.anti-patterns.instructions.md) | Common anti-patterns and mistakes to avoid based on post-mortems and session learnings |
