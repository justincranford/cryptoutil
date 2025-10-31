# Copilot Instructions

## Core Principles
- Follow README and instructions files
- Refer to architecture and usage examples in README
- **Instruction files auto-discovered from `.github/instructions/`** - use `.instructions.md` extension with YAML frontmatter

## Instruction File Structure

**Naming Convention**: Semantic names with ##-##. prefix (Tier-Priority format)
- **First ##** = Tier number (01-04, priority level)
- **Second ##** = Priority within tier (01-05, load sequence)
- **Format**: `##-##.semantic-name.instructions.md`
- **Example**: `02-03.architecture.instructions.md` (Tier 2, Priority 3)

### Tier 1: Foundation (Always Loads First)
| File | Pattern | Description |
|------|---------|-------------|
| 01-01.copilot-customization | ** | Git ops, terminal patterns, curl/wget rules, fuzz testing, conventional commits, TODO maintenance |

### Tier 2: Core Development (Slots 2-6)
| File | Pattern | Description |
|------|---------|-------------|
| 02-01.code-quality | ** | Linter compliance, wsl/godot rules, resource cleanup, pre-commit docs |
| 02-02.testing | ** | Test patterns, dependency mgmt, file organization, UUIDv7 concurrency |
| 02-03.architecture | ** | Layered arch, config patterns, lifecycle, factory patterns, atomic ops |
| 02-04.security | ** | Key hierarchy, IP allowlisting, rate limiting, TLS, secrets management |
| 02-05.docker | **/*.yml | Compose config, healthchecks, Docker secrets, OTEL forwarding |

### Tier 3: High Priority (Slots 7-11)
| File | Pattern | Description |
|------|---------|-------------|
| 03-01.crypto | ** | NIST FIPS 140-3 algorithms, keygen patterns, cryptographic operations |
| 03-02.cicd | .github/workflows/*.yml | Workflow configuration, Go version consistency, service connectivity |
| 03-03.observability | ** | OpenTelemetry integration, OTLP protocols, telemetry forwarding |
| 03-04.database | ** | GORM ORM patterns, migrations, PostgreSQL/SQLite support |
| 03-05.go-standards | **/*.go | Import aliases, dependencies, formatting (gofumpt), conditionals |

### Tier 4: Medium Priority (Slots 12-15)
| File | Pattern | Description |
|------|---------|-------------|
| 04-01.specialized-testing | ** | Act workflow testing, localhost vs 127.0.0.1 patterns |
| 04-02.project-config | ** | OpenAPI specs, magic values, linting exclusions |
| 04-03.platform-specific | scripts/** | PowerShell/Bash scripts, Docker image pre-pull |
| 04-04.specialized-domains | ** | CA/Browser Forum compliance, project layout, PR descriptions |

## Cross-Reference Guide

**Docker Compose**: 02-05.docker (primary) → 02-03.architecture, 03-03.observability, 02-04.security, 02-02.testing  
**CI/CD Workflows**: 03-02.cicd (primary) → 02-05.docker, 04-01.specialized-testing, 02-02.testing  
**Security**: 02-04.security (primary) → 02-05.docker, 03-01.crypto, 04-04.specialized-domains  
**Observability**: 03-03.observability (primary) → 02-05.docker, 02-03.architecture  
**Testing**: 02-02.testing (primary) → 02-05.docker, 04-01.specialized-testing, 02-01.code-quality  
**Go Code**: 03-05.go-standards (primary) → 04-02.project-config, 02-01.code-quality, 04-04.specialized-domains
