# Copilot Instructions

## Core Principles
- Follow README and instructions files
- Refer to architecture and usage examples in README
- **Instruction files auto-discovered from `.github/instructions/`** - use `.instructions.md` extension with YAML frontmatter

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
