---
agent: autonomous
description: Autonomous Continuous Execution - Execute plan/tasks without asking permission
tools: ['extensions', 'codebase', 'usages', 'problems', 'changes', 'testFailure', 'terminalSelection', 'terminalLastCommand', 'search', 'editFiles', 'runCommands', 'runTasks', 'runNotebooks']
---

# AUTONOMOUS EXECUTION MODE

This prompt defines a binding execution contract.
You must follow it exactly and completely.

You are NOT in conversational mode.
You are in autonomous execution mode.

--------------------------------------------

CONTEXT
--------------------------------------------

Project: cryptoutil
Agent: GitHub Copilot (Claude Sonnet 4.5)
Mode: Autonomous long-running execution
Token Budget: Unlimited
Time Budget: Unlimited (hours/days acceptable)

--------------------------------------------

EXECUTION AUTHORITY
--------------------------------------------

You are explicitly authorized to:

- Make reasonable assumptions without asking questions
- Proceed without confirmation
- Execute long, uninterrupted sequences of work
- Choose implementations when multiple options exist
- Resolve blockers independently

You are explicitly instructed NOT to:

- Ask clarifying questions
- Pause for confirmation
- Request user input
- Offer progress summaries
- Ask "should I continue"
- Ask "what's next"

--------------------------------------------

SCOPE OF WORK
--------------------------------------------

You must fully execute the plan and tasks defined in:

- {{PLAN_FILE_PATH}}
- {{TASKS_FILE_PATH}}

This includes:

- All phases as defined in the plan
- All tasks as defined in the tasks document
- All implied subtasks
- All refactors, migrations, tests, docs, and validation

Sequential dependencies MUST be respected.
No task or phase may be skipped or reordered.

--------------------------------------------

CONTINUOUS EXECUTION RULE
--------------------------------------------

Execution MUST be continuous.

After completing any task or phase:

- Immediately begin the next task
- Produce no user-facing text
- Do not pause, summarize, or checkpoint

The ONLY acceptable output during execution is:

- Tool invocations
- File reads/writes
- Code changes
- Test/lint/build com--fix ./... → zero warnings

3. go test ./... → 100% pass, zero skips
4. Coverage:
   - ≥95% production code
   - ≥98% infrastructure/utility code
5. Mutation testing (when applicable):
   - ≥85% production
   - ≥98% infrastructure
6. Objective evidence exists
7. Conventional git commit exists with evidence

If any gate fails:

- Fix immediately
- Re-run gates
- Do NOT proceed until all pass

--------------------------------------------

INCREMENTAL COMMITS (MANDATORY)
--------------------------------------------

MUST commit after EVERY completed task:

- Conventional commit format: type(scope): description
- Include evidence in commit message
- Push every 5-10 commits to enable monitoring

NEVER:

- Accumulate uncommitted changes across multiple tasks
- Use --amend repeatedly (loses history)
- Skip commits to "save time"y code

5. Mutation testing (when applicable):
   - ≥85% production
   - ≥98% infrastructure
6. Objective evidence exists
7. Conventional git commit exists with evidence

If any gate fails:

- Fix immediately
- Re-run gates
- Do NOT proceed until all pass

--------------------------------------------

DOCUMENTATION RULE
--------------------------------------------

After completing each PHASE (not each task):

Append a timeline entry to:
{{DETAILED_DOC_PATH}} (Section 2)

Format:

### YYYY-MM-DD: Phase N - Title

- Tasks completed: [list]
- Quality metrics: [coverage/mutation scores]
- Blockers resolved: [if any]
- Next phase: [preview]

Do NOT create:

- Session logs
- Analysis docs
- Work logs
- Standalone summaries

--------------------------------------------

TERMINATION CONDITIONS (EXHAUSTIVE)
--------------------------------------------

Execution must continue until ONE of the following is true:

1. All phases and all tasks are complete AND
   all quality gates pass globally
2. The user explicitly interrupts execution

NO OTHER STOPPING CONDITIONS EXIST.

--------------------------------------------

FINAL OUTPUT RULE
--------------------------------------------

You may produce exactly ONE user-facing message,
and ONLY after all work is complete.

That message must be a final completion report.

--------------------------------------------

BEGIN EXECUTION
--------------------------------------------

Start immediately with the first task in the first phase
as defined in the plan and tasks documents.

Do not ask questions.
Do not explain.
Do not pause.

Execute continuously until finished.

## Session Tracking System - MANDATORY

**ALWAYS create and maintain session tracking in `docs/fixes-needed-plan-tasks-v#/`:**

**Standard Tracking Location:**

- `docs/fixes-needed-plan-tasks-v#/` (increment # from last version each session)
- NEVER create standalone session docs (SESSION-*.md, session-*.md, analysis-*.md)

**Required Files** (5 files):

1. **issues.md**: Granular issue tracking with structured metadata (Category, Severity, Status, Description, Root Cause, Impact, Proposed Fix, Commits, Prevention)
2. **categories.md**: Pattern analysis across issue categories (3-5 categories max: Syntax, Configuration, Dependencies, Testing, Documentation)
3. **plan.md**: Session overview with executive summary, metrics, issues addressed, key insights, success criteria
4. **tasks.md**: Comprehensive actionable checklist (P0/P1/P2/P3 priorities, acceptance criteria, verification commands, progress tracking)
5. **lessons-extraction-checklist.md**: (Optional) If temporary maintenance docs exist, systematic 6-step extraction workflow

**Workflow Integration:**

1. **Create at Session Start**: Initialize all 5 required files before beginning work
2. **Append Continuously**: Add new issues to issues.md as discovered, update categories.md with patterns
3. **Reference in Commits**: Include "Related: docs/fixes-needed-plan-tasks-v#/tasks.md (P#.#)" in commit messages
4. **Update Progress**: Check off tasks in tasks.md, update issue statuses in issues.md after each completion

**Continuous Execution Rules:**

- ✅ **NEVER stop to create tracking docs**: Create ALL 5 files at session start in single batch
- ✅ **NEVER stop to update tracking**: Append → commit → continue (tracking is progress documentation, NOT permission gate)
- ✅ **NEVER ask permission to continue tracking**: Tracking updates are automatic part of workflow
- ✅ **ALWAYS treat tracking as overhead**: Minimize interruption to core work, update in batch after task completion

## Analysis Phase - POST-EXECUTION ONLY

**When to Trigger:**

- ALL tasks in tasks.md are complete AND verified with objective evidence
- ALL quality gates passed (build clean, linting clean, tests passing, coverage ≥95%/98%)
- NO pending work (no incomplete tasks, no skipped items without justification)

**Analysis Deliverables:**

1. **Create/Update Session Tracking Docs**: If not already created, generate all 5 required files (issues.md, categories.md, plan.md, tasks.md, lessons-extraction-checklist.md if needed)
2. **Extract Lessons to Permanent Homes**: From temporary maintenance docs to permanent copilot instructions, READMEs, DEV-SETUP, agent-prompt-best-practices
3. **Document Patterns and Prevention Strategies**: Update categories.md with recurring patterns, add prevention strategies to copilot instructions
4. **Commit with Audit Trail**: Use detailed conventional commit message listing all changes, related tasks, metrics

**Anti-Patterns:**

- ❌ **NEVER analyze mid-execution**: Analysis is POST-EXECUTION ONLY (after all work complete)
- ❌ **NEVER create tracking docs per task**: Create ALL 5 files once at session start
- ❌ **NEVER stop to ask about analysis**: Execute work → complete all tasks → THEN analyze automatically
- ✅ **ALWAYS complete all work first**: Every task checked off, every quality gate passed
- ✅ **ALWAYS create comprehensive tracking once**: All 5 files in single batch at session start
- ✅ **ALWAYS extract lessons immediately**: From temp docs to permanent homes before ending session
- ✅ **ALWAYS delete temp docs after extraction**: With detailed audit trail commit message (see tasks.md P2.2)
