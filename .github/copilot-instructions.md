# Copilot Instructions

## Core Principles

- Instruction files auto-discovered from .github/instructions/*.instructions.md
- Keep rules short, one directive per line
- Reference external & project resources; avoid duplication of content
- ALWAYS use built-in tools over terminal commands

## LLM Agent Continuous Work Directive - ABSOLUTE ENFORCEMENT

**CRITICAL: NEVER STOP WORKING UNTIL USER EXPLICITLY SAYS "STOP"**

**CRITICAL: NEVER FEEL TIME OR TOKEN PRESSURE WHEN WORKING ON SPECKIT**

**CRITICAL: ALWAYS COMMIT CHANGES IMMEDIATELY WHEN WORK IS COMPLETE**

### Time and Token Management for Speckit Work

- **NO TIME PRESSURE**: Speckit work can span hours or days - this is expected and acceptable
- **NO TOKEN PRESSURE**: Large tasks requiring many tool calls are normal - use as many tokens as needed
- **DECOMPOSE COMPLEX TASKS**: Break hard/long/complete issues into smaller, manageable subtasks
- **UPDATE DETAILED.md/TASKS.md**: Document decomposed tasks and progress continuously
- **CORRECTNESS > SPEED**: Always prioritize doing things correctly and completely over rushing
- **NO RUSHED DECISIONS**: Take time to analyze, plan, and implement properly
- **NO SKIPPING GOALS**: All goals and tasks must be completed - no shortcuts

**When facing complex work**:

1. Decompose into smaller tasks (add to DETAILED.md Section 1)
2. Update DETAILED.md Section 2 timeline with decomposition rationale
3. Execute each subtask completely and correctly
4. Mark subtasks complete with evidence in timeline
5. Continue to next subtask without stopping

### Prohibited Stop Behaviors (ALL FORBIDDEN)

- ❌ **NO status summaries at end of response** - summaries = stopping
- ❌ **NO "session complete" messages** - completion statements = stopping
- ❌ **NO "remaining tasks" lists at end** - task lists at end = stopping
- ❌ **NO "next steps" sections at end** - next steps at end = stopping
- ❌ **NO "strategic pivot" or "strategic decision" statements** - these are stopping excuses
- ❌ **NO "given time constraints" or "token budget" justifications** - these are stopping excuses
- ❌ **NO saying "NO STOPPING OCCURRED"** - if you say this, YOU STOPPED
- ❌ **NO pauses between tasks** - pause = stop
- ❌ **NO asking for permission** - asking = stopping
- ❌ **NO waiting for confirmation** - waiting = stopping
- ❌ **NO leaving uncommitted changes** - uncommitted changes = incomplete work
- ❌ **NO ending response with analysis or documentation** - if not executing tools, YOU STOPPED

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
- NO ending response without active tool execution
- **IF MAKING PROGRESS**: Keep going - progress is not a stop condition
- **IF YOU COMPLETE A TASK**: Immediately start next task (no summary)
- **IF STUCK OR BLOCKED**: Pick next unblocked task and execute - NEVER stop to ask
- **IF ONE TASK BLOCKED**: Continue with all other remaining tasks - blockers don't stop work
- **IF TODO LIST EMPTY**: Read DETAILED.md and continue with next incomplete task
- **IF YOU TYPE TEXT WITHOUT INVOKING A TOOL**: You are about to stop - invoke a tool instead
- **LAST RESPONSE MUST BE A TOOL INVOCATION**: Never end with text, always end with tool call

### Handling Blockers and Issues

**CRITICAL: Blockers on one task NEVER justify stopping all work**

- **When blocked on Task A**: Immediately switch to Task B, C, D... (continue all other work)
- **Document blocker**: Add brief note to DETAILED.md Section 2 timeline about what's blocked
- **Keep working**: Return to blocked task only when blocker is resolved
- **NO stopping to ask**: If user input needed, document requirement and continue other work
- **NO waiting**: Never wait for external dependencies - work on everything else meanwhile

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
