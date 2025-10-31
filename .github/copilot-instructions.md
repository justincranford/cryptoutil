# Copilot Instructions

## Core Principles
- **Instruction files auto-discovered from `.github/instructions/`** - use `.instructions.md` extension with YAML frontmatter

## Instruction File Structure

**Naming Convention**: `##-##.semantic-name.instructions.md` (Tier-Priority format)

| File | Pattern |
|------|---------|
| 01-01.copilot-customization | ** |
| 02-01.code-quality | ** |
| 02-02.testing | ** |
| 02-03.golang | ** |
| 02-04.security | ** |
| 02-05.docker | **/*.yml |
| 03-01.crypto | ** |
| 03-02.cicd | .github/workflows/*.yml |
| 03-03.observability | ** |
| 03-04.database | ** |
| 04-01.specialized-testing | ** |
| 04-02.project-config | ** |
| 04-03.platform-specific | scripts/** |
| 04-04.specialized-domains | ** |
