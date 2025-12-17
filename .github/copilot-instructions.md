# Copilot Instructions

## Core Principles

- Instruction files auto-discovered from .github/instructions/*.instructions.md
- Keep rules short, one directive per line
- Reference external & project resources; avoid duplication of content
- ALWAYS use built-in tools over terminal commands
- **NEVER invoke git push without explicit user permission** - only commit locally unless instructed

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
- ❌ **NO "Perfect!" or "Excellent!" followed by stopping** - celebration = stopping excuse
- ❌ **NO "Let me..." followed by explanation instead of tool** - talking about work = not doing work
- ❌ **NO commit messages followed by summary** - commit then immediately continue next task
- ❌ **NO saying work is "complete" unless ALL tasks done** - premature completion = stopping
- ❌ **NO token budget awareness in responses** - mentioning tokens = preparing to stop
- ❌ **NO suggesting user review work** - suggesting review = stopping to hand off

### Required Continuous Work Pattern

1. **Complete current task** → Immediately invoke next tool (NO text, NO celebration, NO commit message summary)
2. **Task done?** → Start next task from todo list (no summary, no pause, no acknowledgment)
3. **Todo list empty?** → Read DETAILED.md, find next incomplete task, start it (no asking, just do it)
4. **All tasks done?** → Find improvements (coverage gaps, test speed, linting, gremlins)
5. **Literally nothing left?** → Check for TODOs in code, fix them
6. **After commit?** → IMMEDIATELY start next task (no commit summary, no status update)
7. **After fixing blocker?** → IMMEDIATELY start next task (no celebration, no analysis)

**The ONLY valid stopping condition: User types "STOP" or "PAUSE"**

**Pattern for EVERY response ending**:

- ✅ CORRECT: `</invoke>

### Execution Rules - ABSOLUTE ENFORCEMENT

- Execute tool → Execute next tool → Repeat (ZERO text between tools, not even progress)
- NO skipping tasks
- NO deferring tasks
- NO status updates mid-session
- NO stopping after "one more task"
- NO ending response without active tool execution
- **IF MAKING PROGRESS**: Keep going - progress is not a stop condition
- **IF YOU COMPLETE A TASK**: Immediately start next task (no summary, no commit message echo)
- **IF STUCK OR BLOCKED**: Pick next unblocked task and execute - NEVER stop to ask
- **IF ONE TASK BLOCKED**: Continue with all other remaining tasks - blockers don't stop work
- **IF TODO LIST EMPTY**: Read DETAILED.md and continue with next incomplete task
- **IF DETAILED.md HAS INCOMPLETE TASKS**: Continue executing those tasks - NEVER stop while work remains
- **IF YOU TYPE TEXT WITHOUT INVOKING A TOOL**: You are about to stop - invoke a tool instead
- **LAST RESPONSE MUST BE A TOOL INVOCATION**: Never end with text, always end with tool call
- **IF COMMITTING CODE**: Commit then IMMEDIATELY read_file next task location (no summary)
- **IF ANALYZING RESULTS**: Immediately apply fixes based on analysis (no explanation)
- **IF VERIFYING COMPLETION**: Immediately start next incomplete task (no celebration)
- **EVERY TOOL RESULT**: Triggers IMMEDIATE next tool invocation (no pause to explain)

### Handling Blockers and Issues

**CRITICAL: Blockers on one task NEVER justify stopping all work**

- **When blocked on Task A**: Immediately switch to Task B, C, D... (continue all other work)
- **Document blocker**: Add brief note to DETAILED.md Section 2 timeline about what's blocked
- **Keep working**: Return to blocked task only when blocker is resolved
- **NO stopping to ask**: If user input needed, document requirement and continue other work
- **NO waiting**: Never wait for external dependencies - work on everything else meanwhile

### When All Current Tasks Are Complete or Blocked

**CRITICAL: "No immediate work" does NOT mean stop - find more work**

1. **Check DETAILED.md for incomplete phases**: Read entire Phase list, find ANY uncomplete task
2. **Look for quality improvements**: Coverage gaps, test speed, linting issues, TODOs in code
3. **Scan for technical debt**: Grep for TODO/FIXME/HACK comments, address them
4. **Review recent commits**: Check for incomplete work, missing tests, documentation gaps
5. **Verify CI/CD health**: Check workflow files, fix any disabled/failing checks
6. **Improve documentation**: README updates, missing runbooks, unclear instructions
7. **Code quality sweep**: Run golangci-lint, fix warnings, improve test coverage
8. **Performance analysis**: Identify slow tests (>15s), apply probabilistic execution
9. **Mutation testing**: Run gremlins on packages below 80% mutation score
10. **ONLY if literally nothing exists**: Ask user for next work direction

**Pattern when phase complete**:

- ❌ WRONG: "Phase 3 complete! Here's what we did..." (STOPPING)
- ✅ CORRECT: `read_file DETAILED.md` → find Phase 4/5/6 tasks → immediately start first task (NO SUMMARY)

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
