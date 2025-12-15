# Copilot Instructions

## Core Principles

- Instruction files auto-discovered from .github/instructions/*.instructions.md
- Keep rules short, one directive per line
- Reference external & project resources; avoid duplication of content
- ALWAYS use built-in tools over terminal commands

## LLM Agent Continuous Work Directive - ABSOLUTE ENFORCEMENT

**CRITICAL: NEVER STOP WORKING UNTIL USER EXPLICITLY SAYS "STOP"**

### Prohibited Stop Behaviors (ALL FORBIDDEN)

- ❌ **NO status summaries at end of response** - summaries = stopping
- ❌ **NO "session complete" messages** - completion statements = stopping
- ❌ **NO "remaining tasks" lists at end** - task lists at end = stopping
- ❌ **NO "next steps" sections at end** - next steps at end = stopping
- ❌ **NO saying "NO STOPPING OCCURRED"** - if you say this, YOU STOPPED
- ❌ **NO pauses between tasks** - pause = stop
- ❌ **NO asking for permission** - asking = stopping
- ❌ **NO waiting for confirmation** - waiting = stopping

### Required Continuous Work Pattern

1. **Complete current task** → Immediately invoke next tool
2. **Task done?** → Start next task from todo list (no summary, no pause)
3. **Todo list empty?** → Read DETAILED.md, find next incomplete task, start it
4. **All tasks done?** → Find improvements (coverage gaps, test speed, linting)
5. **Literally nothing left?** → Check for TODOs in code, fix them

**The ONLY valid stopping condition: User types "STOP" or "PAUSE"**

### Execution Rules

- Execute tool → Execute next tool → Repeat (no text between tools except brief progress)
- NO skipping tasks
- NO deferring tasks
- NO status updates mid-session
- NO stopping after "one more task"
- **IF MAKING PROGRESS**: Keep going - progress is not a stop condition
- **IF YOU COMPLETE A TASK**: Immediately start next task (no summary)
- **IF STUCK**: Pick next task and execute - NEVER stop to ask
- **IF TODO LIST EMPTY**: Read DETAILED.md and continue with next incomplete task

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
