# Copilot Instructions

## Core Principles

- Instruction files auto-discovered from .github/instructions/*.instructions.md
- Keep rules short, one directive per line
- Reference external & project resources; avoid duplication of content
- ALWAYS use built-in tools over terminal commands

## LLM Agent Continuous Work Directive - ABSOLUTE ENFORCEMENT

- Execute tool → Execute next tool → Repeat
- NO skipping tasks
- NO deferring tasks
- NO pauses between tasks
- NO status updates mid-session
- NO asking for permission to continue
- NO stopping after "one more task"

**IF MAKING PROGRESS**: Keep going - progress is not a stop condition
**IF YOU COMPLETE A TASK**: Immediately start the next one from the todo list
**IF STUCK**: Pick next task and execute it - NEVER stop to ask
**IF TODO LIST EMPTY**: Analyze project and find developer experience improvements: build speed, test speed, test coverage, mutation test quality, etc.

## Instruction Files Reference

| File | Description |
|------|-------------|
| 01-01.architecture | Products & Services Architecture |
| 01-02.versions | Minimum Versions & Consistency Requirements |
| 01-03.coding | Coding patterns & standards |
| 01-04.testing | Testing patterns & best practices |
| 01-05.golang | Go project structure & conventions |
| 01-06.database | Database & ORM patterns |
| 01-07.security | Security patterns |
| 01-08.linting | Code quality & linting standards |
| 02-01.github | CI/CD workflow |
| 02-02.docker | Docker & Compose |
| 02-03.observability | Observability & monitoring |
| 03-01.openapi | OpenAPI rules |
| 03-02.cross-platform | Cross-platform tooling |
| 03-03.git | Git workflow rules |
| 03-04.dast | DAST scanning |
| 04-01.sqlite-gorm | SQLite GORM config |
| 05-01.evidence-based | Evidence-based task completion |
