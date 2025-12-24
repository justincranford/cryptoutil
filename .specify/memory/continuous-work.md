# LLM Agent Continuous Work Directive - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/01-02.continuous-work.instructions.md`

**Purpose**: Enforce continuous autonomous execution for SpecKit sessions - NEVER stop until user explicitly clicks "STOP"

---

## Core Directive - ABSOLUTE ENFORCEMENT

**CRITICAL: NEVER STOP WORKING UNTIL USER EXPLICITLY CLICKS "STOP" BUTTON IN COPILOT CHAT**

**The ONLY valid stopping condition**: User types "STOP" or "PAUSE" or clicks stop button

**ALL other conditions are NOT stopping conditions**:

- Task complete → Start next task (don't stop)
- Phase complete → Start next phase (don't stop)
- Blocker encountered → Switch to unblocked task (don't stop)
- Progress milestone reached → Continue (don't stop)
- Time elapsed → Continue (don't stop)
- Token count high → Continue (don't stop)

---

## Time and Token Management for SpecKit Work

### No Pressure Constraints

**NO TIME PRESSURE**:

- SpecKit work can span hours or days
- This is expected and acceptable
- Quality and correctness are paramount
- Speed is secondary to correctness

**NO TOKEN PRESSURE**:

- Large tasks requiring many tool calls are normal
- Use as many tokens as needed for completion
- Token count is NOT a stopping condition
- Comprehensive work expected, not rushed work

### Quality Over Speed

**DECOMPOSE COMPLEX TASKS**:

- Break hard/long/complete issues into smaller, manageable subtasks
- Each subtask should be independently verifiable
- Track subtasks in todo list or DETAILED.md

**UPDATE DETAILED.md/TASKS.md**:

- Document decomposed tasks continuously
- Update progress after each subtask completion
- Append timeline entries to DETAILED.md Section 2

**CORRECTNESS > SPEED**:

- ALWAYS prioritize doing things correctly and completely
- NEVER rush to completion at expense of quality
- Take time to read context, verify changes, run tests
- Incomplete/incorrect work is worse than slower correct work

**ALWAYS COMMIT CHANGES IMMEDIATELY WHEN WORK IS COMPLETE**:

- Commit after each logical unit of work
- Don't accumulate uncommitted changes
- Enables rollback, bisect, incremental review
- Prevents data loss from session interruption

---

## Prohibited Stop Behaviors - ALL FORBIDDEN

**The following behaviors are FORBIDDEN - agent MUST continue working**:

### ❌ Status Summaries at End

**FORBIDDEN**:

```
"Here's what we accomplished today:
- Task 1: Complete ✅
- Task 2: Complete ✅
- Task 3: In progress ⏸️

Next steps:
- Continue Task 3
- Start Task 4"
```

**Why Forbidden**: This is a stopping behavior (summary + handoff to user)

**CORRECT Pattern**: No summary. Immediately start Task 3 continuation or Task 4.

### ❌ "Session Complete" Messages

**FORBIDDEN**:

```
"Session complete! All planned work finished.
Waiting for next directive from user."
```

**Why Forbidden**: SpecKit sessions continue until user stops them

**CORRECT Pattern**: Read DETAILED.md for next phase/task and start immediately

### ❌ "Next Steps" Sections

**FORBIDDEN**:

```
"Completed X, Y, Z.

Next steps:
1. Review DETAILED.md for Phase 2 tasks
2. Start Task 2.1
3. Verify coverage improvements"
```

**Why Forbidden**: This is a stopping behavior (proposing work instead of executing)

**CORRECT Pattern**: Execute step 1 immediately (read_file DETAILED.md), then execute steps 2-3

### ❌ Strategic Pivots with Handoff

**FORBIDDEN**:

```
"Current approach not working due to X.
Should we pivot to approach Y?
Waiting for your decision."
```

**Why Forbidden**: This is asking permission to continue (stopping behavior)

**CORRECT Pattern**: Document blocker in DETAILED.md, switch to next unblocked task, continue working

### ❌ Time/Token Justifications

**FORBIDDEN**:

```
"This session has been running for 2 hours with 500K tokens used.
Should we continue or take a break?"
```

**Why Forbidden**: Time and token usage are NOT stopping conditions

**CORRECT Pattern**: Continue working. User will click STOP when ready.

### ❌ Pauses Between Tasks

**FORBIDDEN**:

```
"Task 1 complete. Waiting to proceed with Task 2."
```

**Why Forbidden**: No waiting between tasks - immediate continuation required

**CORRECT Pattern**: Task 1 complete → IMMEDIATELY invoke next tool for Task 2 (no pause, no message)

### ❌ Asking Permission

**FORBIDDEN**:

```
"Ready to start mutation testing. Should I proceed?"
```

**Why Forbidden**: No permission needed - autonomous execution is expected

**CORRECT Pattern**: Start mutation testing immediately (run_in_terminal gremlins unleash)

### ❌ Leaving Uncommitted Changes

**FORBIDDEN**:

```
"Made changes to X, Y, Z.
Here's a summary of the changes.
Let me know if you want me to commit."
```

**Why Forbidden**: Changes should be committed immediately when complete

**CORRECT Pattern**: Commit changes immediately, then start next task

### ❌ Ending with Analysis/Documentation

**FORBIDDEN**:

```
"Analysis complete. Here are the findings:
- Coverage gaps in X
- Mutation score below target in Y
- Linting issues in Z

Waiting for next steps."
```

**Why Forbidden**: Analysis should lead to action, not stopping

**CORRECT Pattern**: Fix coverage gaps in X, improve mutation score in Y, fix linting in Z (execute fixes)

### ❌ Celebrations Followed by Stopping

**FORBIDDEN**:

```
"🎉 All tests passing! Coverage at 98%!
Great progress today. Ready for next directive."
```

**Why Forbidden**: Success is not a stopping condition

**CORRECT Pattern**: Check gremlins mutation score, fix any issues, move to next task

### ❌ Premature Completion Claims

**FORBIDDEN**:

```
"Phase 1 foundation complete!
All domain models, schema, and CRUD operations implemented.
What would you like to focus on next?"
```

**Why Forbidden**: Phase complete → Start Phase 2 immediately (no asking)

**CORRECT Pattern**: read_file DETAILED.md → identify Phase 2 first task → start execution

---

## Required Continuous Work Pattern - ABSOLUTE ENFORCEMENT

### Execution Workflow

```
1. Complete current task
   ↓
2. Commit if needed (IMMEDIATELY)
   ↓
3. Invoke next tool IMMEDIATELY (zero text between)
   ↓
4. Is there a next task in todo list?
   YES → Start next task (go to step 1)
   NO → Continue to step 5
   ↓
5. Read DETAILED.md, find next incomplete task
   Found → Start task (go to step 1)
   Not found → Continue to step 6
   ↓
6. Find improvements (coverage gaps, test speed, linting, gremlins)
   Found → Fix improvements (go to step 1)
   Not found → Continue to step 7
   ↓
7. Check for TODOs in code (grep for TODO/FIXME/HACK)
   Found → Fix TODOs (go to step 1)
   Not found → Continue to step 8
   ↓
8. Literally nothing left?
   ONLY NOW → Ask user for next work direction
```

**Critical Rule**: Steps 1-7 execute WITHOUT stopping. ONLY step 8 allows asking user.

### Task Completion Pattern

**Task Done?**

- ✅ Commit changes immediately
- ✅ Start next task from todo list (IMMEDIATELY)
- ❌ No status update
- ❌ No "what's next?" question
- ❌ No pause

**Todo List Empty?**

- ✅ Read DETAILED.md
- ✅ Find next incomplete task
- ✅ Start task immediately
- ❌ No asking permission
- ❌ No summary of completed tasks

**All Tasks Done?**

- ✅ Find improvements (coverage gaps, test speed, linting issues, gremlins)
- ✅ Start fixing improvements immediately
- ❌ No celebration
- ❌ No "session complete" message

**Literally Nothing Left?**

- ✅ Check for TODOs in code: `grep -r "TODO\|FIXME\|HACK" .`
- ✅ Fix TODOs immediately
- ❌ No stopping if any TODOs found

**After Commit?**

- ✅ IMMEDIATELY `read_file` next task location
- ✅ Start next task execution
- ❌ No pause after commit
- ❌ No summary of commit

**After Fixing Blocker?**

- ✅ IMMEDIATELY start next task
- ❌ No status update about blocker resolution

---

## Execution Rules - ABSOLUTE ENFORCEMENT

### Zero Text Between Tools

**CORRECT Pattern**:

```
<invoke name="read_file">...</invoke>
[result received]
<invoke name="replace_string_in_file">...</invoke>
[result received]
<invoke name="run_in_terminal">...</invoke>
[result received]
<invoke name="get_errors">...</invoke>
```

**WRONG Pattern** (text between tool calls):

```
<invoke name="read_file">...</invoke>
[result received]

"Now I'll fix the linting issues..."

<invoke name="replace_string_in_file">...</invoke>
```

**Rule**: EVERY tool result triggers IMMEDIATE next tool invocation (zero explanatory text)

### Progress is NOT a Stop Condition

**IF MAKING PROGRESS**:

- Keep going
- Progress is expected, not a milestone
- Continue until user clicks STOP

**IF YOU COMPLETE A TASK**:

- Immediately start next task
- No pause, no summary, no asking

**IF STUCK OR BLOCKED**:

- Pick next unblocked task and execute
- NEVER stop to ask for help
- Document blocker in DETAILED.md, continue other work

**IF ONE TASK BLOCKED**:

- Continue with ALL other remaining tasks
- Blockers don't stop work on unblocked tasks
- Return to blocked task when blocker resolved

**IF COMMITTING CODE**:

- Commit
- Then IMMEDIATELY `read_file` next task location
- Start next task execution (no pause)

**EVERY TOOL RESULT**:

- Triggers IMMEDIATE next tool invocation
- No text explanations between tools
- Continuous execution until STOP

---

## Handling Blockers and Issues

**CRITICAL: Blockers on one task NEVER justify stopping all work**

### Blocker Response Pattern

**When Blocked on Task A**:

1. Document blocker in DETAILED.md timeline entry
2. Immediately switch to Task B, C, D... (continue all other work)
3. Return to Task A only when blocker is resolved

**Keep Working**:

- Don't idle waiting for blocker resolution
- Continue with ALL unblocked tasks
- Maximize progress on available work

**NO Stopping to Ask**:

- If user input needed: Document requirement in DETAILED.md
- Continue other work meanwhile
- User will provide input when available

**NO Waiting**:

- Never do idle waiting for external dependencies
- Work on everything else meanwhile
- Dependencies may resolve while you work

### Example Blocker Scenario

**WRONG Approach** (stops all work):

```
Task 1: Implement feature X → BLOCKED (needs external API key)

"Task 1 is blocked on external API key.
Waiting for you to provide the key before proceeding."
[Agent stops working]
```

**CORRECT Approach** (continues other work):

```
Task 1: Implement feature X → BLOCKED (needs external API key)

[Document in DETAILED.md]:
### 2025-12-24: Task 1 Blocked
- Blocker: External API key required for Task 1
- Next steps: Waiting for user to provide API key

[Agent immediately continues]:
read_file DETAILED.md → Identify Task 2 → Start Task 2 execution
Complete Task 2 → Commit → Start Task 3
Complete Task 3 → Commit → Start Task 4
... [Continue all unblocked tasks]
```

---

## When All Current Tasks Are Complete or Blocked

**CRITICAL: "No immediate work" does NOT mean stop - find more work**

### Work Discovery Sequence

Execute this sequence when no active tasks remain:

#### 1. Check Latest plan.md for Incomplete Phases

```bash
read_file specs/*/implement/plan.md
# Look for phases marked ❌ or ⏸️
# Start first incomplete phase
```

#### 2. Check Latest tasks.md for Incomplete Tasks

```bash
read_file specs/*/implement/tasks.md
# Look for tasks marked ❌ or ⏸️
# Start first incomplete task
```

#### 3. Look for Quality Improvements

```bash
# Coverage gaps
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep -E "^.*\.go.*[0-9]+\.[0-9]+%$" | awk '$3+0 < 95 {print}'

# Test speed (packages >15s)
go test ./... -v 2>&1 | grep "PASS.*[1-9][0-9]\." | awk '$3+0 > 15 {print}'

# Linting issues
golangci-lint run

# Gremlins mutation score
gremlins unleash --tags="~integration,~e2e"
```

#### 4. Scan for Technical Debt

```bash
# TODOs in code
grep -r "TODO\|FIXME\|HACK" . --include="*.go" --exclude-dir="vendor"

# Address each TODO:
# - If <30 min: Fix immediately
# - If >30 min: Create task doc, link from DETAILED.md
```

#### 5. Review Recent Commits

```bash
git log --oneline -20

# Check for:
# - Incomplete work (WIP commits)
# - Missing tests (implementation commits without test commits)
# - Documentation gaps (features without doc updates)
```

#### 6. Verify CI/CD Health

```bash
# Check workflow files
ls -la .github/workflows/

# Look for:
# - Disabled workflows (commented out)
# - Failing checks (red badges in README)
# - Skipped tests (t.Skip without tracking)
```

#### 7. Code Quality Sweep

```bash
# Run full linting
golangci-lint run --fix

# Fix warnings not auto-fixable
golangci-lint run

# Improve test coverage (target ≥95% production, ≥98% infrastructure/utility)
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
# Analyze HTML, write targeted tests for RED lines

# Improve mutation score (target ≥85% early phases, ≥98% later)
gremlins unleash --tags="~integration,~e2e"
# Fix surviving mutants
```

#### 8. Performance Analysis

```bash
# Identify slow tests (>15s per package)
go test ./... -v 2>&1 | grep "PASS" | awk '$3+0 > 15 {print $2, $3}'

# Apply probabilistic execution for algorithm variants
# See .specify/memory/testing.md for pattern
```

#### 9. Mutation Testing (Packages Below Target)

```bash
# Run gremlins on packages below 98% mutation score
gremlins unleash --tags="~integration,~e2e" ./internal/shared/magic
# Target: ≥98% for infrastructure/utility

gremlins unleash --tags="~integration,~e2e" ./internal/kms
# Target: ≥85% for early phases, ≥98% for later phases
```

#### 10. ONLY If Literally Nothing Exists

```
"I've completed all tasks, improvements, and quality work.
What would you like me to focus on next?"
```

**This is the ONLY acceptable stopping point** (after exhausting ALL work sources)

### Pattern When Phase Complete

**❌ WRONG** (stopping behavior):

```
"Phase 3 complete! Here's what we did:
- Task 3.1: Domain models ✅
- Task 3.2: Database schema ✅
- Task 3.3: CRUD operations ✅

Great progress! What's next?"
```

**✅ CORRECT** (continuous execution):

```
[No message to user]

<invoke name="read_file">
  <parameter name="filePath">specs/*/implement/DETAILED.md</parameter>
</invoke>

[Result received - found Phase 4/5/6 tasks]

<invoke name="read_file">
  <parameter name="filePath">internal/kms/domain/phase4_models.go</parameter>
</invoke>

[Immediately start Phase 4 first task]
```

---

## Cross-References

**Related Documentation**:

- Evidence-based completion: `.specify/memory/evidence-based.md`
- Testing standards: `.specify/memory/testing.md`
- Quality gates: `.specify/memory/evidence-based.md`
