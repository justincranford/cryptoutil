---
description: "Best practices for GitHub Copilot agent prompt engineering and autonomous execution patterns"
applyTo: "**"
---

# Agent Prompt Best Practices

This document provides comprehensive best practices for creating and enhancing GitHub Copilot agent prompts, autonomous execution patterns, and session tracking workflows.

## Table of Contents

- [YAML Frontmatter - REQUIRED](#yaml-frontmatter---required)
- [Autonomous Execution Patterns](#autonomous-execution-patterns)
- [Session Tracking Integration](#session-tracking-integration)
- [Memory Management](#memory-management)
- [Prompt Enhancement Guidelines](#prompt-enhancement-guidelines)
- [Common Anti-Patterns](#common-anti-patterns)
- [Workflow Integration Examples](#workflow-integration-examples)
- [References](#references)

---

## YAML Frontmatter - REQUIRED

**MANDATORY: All prompt files MUST include YAML frontmatter**

**Format**:

```yaml
---
description: "Brief description of prompt purpose and scope"
author: "Agent" # Or specific author name
created: "YYYY-MM-DD" # Date prompt was created
updated: "YYYY-MM-DD" # Last update date (optional, track major changes)
version: "X.Y" # Semantic version (optional, for complex prompts)
tags: ["workflow", "autonomous", "quality"] # Categorization tags
applyTo: "**" # File patterns this prompt applies to (optional)
---
```

**Purpose**:
- Machine-readable metadata for prompt discovery and organization
- Version tracking for prompt evolution
- Categorization and filtering in `.github/prompts/` directory
- Integration with tooling (prompt search, auto-discovery, validation)

**Common Tags**:
- `workflow` - Multi-step execution workflows
- `autonomous` - Continuous execution patterns
- `quality` - Quality gates and validation
- `session-tracking` - Issue/category tracking
- `analysis` - Post-execution analysis
- `documentation` - Documentation generation

**Example** (from workflow-fixing.prompt.md):

```yaml
---
description: "Systematic workflow for fixing GitHub Actions workflow files with session tracking and quality gates"
author: "Agent"
created: "2025-01-23"
tags: ["workflow", "github-actions", "quality", "session-tracking"]
---
```

---

## Autonomous Execution Patterns

### Core Pattern - MANDATORY

**ALWAYS execute ALL tasks until completion OR user clicks STOP button**

**Continuous Work Directive**:

```
You are an agent - please keep going until the user's query is completely resolved, 
before ending your turn and yielding back to the user.

NEVER stop to ask permission or give status updates between tasks. 
Complete current task → Commit → IMMEDIATELY start next task (ZERO text between).
```

**Key Principles**:

1. **Zero Stopping Behaviors**: No "Here's what we did", "Should I continue?", "What's next?"
2. **Immediate Progression**: Task complete → Commit → Next tool invocation (no pause)
3. **Quality Over Speed**: Take hours/days if needed - correctness matters more than fast completion
4. **Token Budget Unlimited**: Use as many tokens as needed for quality work

### Autonomous Directives - Include in Prompts

**Quality Expectations**:

```
CRITICAL: ALWAYS bias towards accuracy, completeness, and correctness.
NEVER bias towards fast completion at the expense of quality.

Quality Over Speed:
- Take the time required to do things correctly
- Time and token budgets are NOT constraints
- Correctness > Speed (NO EXCEPTIONS)
```

**Continuous Execution**:

```
Continuous Work:
- Execute ALL steps in the plan
- NEVER ask "Should I continue?" between tasks
- NEVER pause for status updates or celebrations
- Task complete → Commit → IMMEDIATELY start next task (zero pause)
```

**Evidence Requirements**:

```
Evidence-Based Completion:
- NEVER mark tasks complete without objective evidence
- Build clean: go build ./...
- Linting clean: golangci-lint run
- Tests passing: go test ./... (no skips without tracking)
- Coverage ≥95% production, ≥98% infrastructure/utility
```

### Quality Gates - Validate After Each Task

**Build Validation**:

```bash
# MANDATORY after every code change
go build ./...
golangci-lint run --fix
golangci-lint run
```

**Test Validation**:

```bash
# MANDATORY after every code change
go test ./... -cover
go test -race -count=2 ./...
```

**Coverage Validation**:

```bash
# Check coverage thresholds
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
# Production ≥95%, Infrastructure/Utility ≥98%
```

---

## Session Tracking Integration

### Standard Location - MANDATORY

**Location**: `docs/fixes-needed-plan-tasks-v#/` (increment version for each session)

**Required Files**:

1. **issues.md** - Granular issue tracking with metadata (ID, title, category, severity, status, description, root cause, impact, proposed fix, related commits)
2. **categories.md** - Pattern analysis across issue categories (category name, issue count, pattern description, prevention strategy, cross-cutting themes)
3. **lessons-extraction-checklist.md** (if extracting from temp docs) - Systematic workflow for permanent doc integration
4. **plan.md** - Session overview (executive summary, issues addressed, key insights, success criteria, next actions, metrics)
5. **tasks.md** - Implementation checklist (P0-P3 priorities, exact content, acceptance criteria, verification commands)

### Tracking Workflow - During Implementation

**As Problems Found**:

```markdown
# Step 1: Document in issues.md
## Issue #{N}: Brief Title
- **Category**: Documentation / Process / Tooling / Code / Testing
- **Severity**: Critical / High / Medium / Low
- **Status**: Identified / In Progress / Completed / Blocked
- **Description**: What went wrong
- **Root Cause**: Why it happened
- **Impact**: What it affects
- **Proposed Fix**: How to resolve
- **Related Commits**: Git commit hashes

# Step 2: Append to categories.md
## Category: {Category Name}
**Issue Count**: {N}
**Issues**: #{ID1}, #{ID2}, ...
**Pattern**: Common characteristics
**Prevention**: How to avoid in future
```

**Continuous Append Pattern**:

- NEVER create new session docs (docs/SESSION-*.md)
- ALWAYS append to existing issues.md and categories.md
- Update plan.md and tasks.md as scope changes

### Issue Template - For issues.md

```markdown
## Issue #{N}: {Brief Title (60 chars max)}

- **Category**: {Documentation | Process | Tooling | Code | Testing}
- **Severity**: {Critical | High | Medium | Low}
- **Status**: {Identified | In Progress | Completed | Blocked}
- **Description**: {What went wrong - 2-3 sentences}
- **Root Cause**: {Why it happened - technical explanation}
- **Impact**: {What it affects - users, workflows, quality}
- **Proposed Fix**: {How to resolve - actionable steps}
- **Related Commits**: {Git commit hashes if applicable}
- **Prevention**: {How to avoid in future - process/tooling changes}

---
```

---

## Memory Management

### Todo List Pattern - For Complex Workflows

**Location**: `./docs/todos-*.md` (1-6 files max, delete completed immediately)

**Purpose**: Track multi-phase work WITHOUT cluttering permanent docs

**Format**:

```markdown
# TODO: {Phase Name}

## In Progress
- [ ] Task 1 description
  - Blockers: {any blockers}
  - Notes: {context}

## Completed
- [x] Task 0 description (completed YYYY-MM-DD)

## Deferred
- [ ] Task 99 description
  - Reason: {why deferred}
  - Revisit: {when to revisit}
```

**Usage**:

- Create at workflow start if >10 tasks
- Update after each task completion
- DELETE immediately when phase complete (move to DETAILED.md if needed)
- NEVER let todos accumulate (max 6 files, each <100 lines)

### Progress Updates - Append to DETAILED.md

**Pattern**: Append chronological entries to DETAILED.md Section 2 timeline

```markdown
### YYYY-MM-DD: {Phase Name}

**Work Completed**:
- Task 1 description
- Task 2 description

**Coverage/Quality Metrics**:
- Before: X% coverage
- After: Y% coverage
- Mutation score: Z%

**Lessons Learned**:
- Lesson 1
- Lesson 2

**Constraints Discovered**:
- Added to constitution.md: {reference}

**Related Commits**: {commit hashes}
```

**Rule**: NEVER create separate session documentation files

---

## Prompt Enhancement Guidelines

### When to Create QUIZME.md

**Create When**:
- Requirements unclear (multiple valid interpretations)
- User domain knowledge needed (business rules, priorities)
- Design trade-offs exist (performance vs simplicity, flexibility vs constraints)

**Skip When**:
- Tasks straightforward (implementation details clear)
- Technical decisions obvious (standard patterns apply)
- All content already specified (exact additions provided)

**Format** (if needed):

```markdown
# Clarification Questions - {Feature Name}

## Question 1: {Brief topic}

**Context**: {Why this matters}

**Options**:
- A) {Option with rationale}
- B) {Option with rationale}
- C) {Option with rationale}
- D) {Option with rationale}
- E) Write-in: ________________

**Recommendation**: {Agent's suggested option with justification}
```

### Quality Over Speed - MANDATORY

**Time Investment Priorities**:

1. **Correctness** (highest) - Code works, tests pass, no regressions
2. **Completeness** - All tasks in checklist done, no skipped items
3. **Quality** - Coverage targets met, mutation score acceptable, linting clean
4. **Efficiency** - Reasonable implementation (not over-engineered)
5. **Speed** (lowest) - Hours/days acceptable for quality work

**NEVER**:
- Rush to completion
- Skip validation steps
- Defer quality checks to "later"
- Mark tasks complete without evidence

**ALWAYS**:
- Take time needed for correctness
- Run full validation after each task
- Fix issues immediately when found
- Document constraints/lessons discovered

### Continuous Work Directive - Reference Implementation

**From** `.github/instructions/01-02.continuous-work.instructions.md`:

- NEVER ask "Should I continue?" - Just continue
- NEVER give status updates between tasks - Task → Commit → Next task
- NEVER celebrate/summarize mid-workflow - Complete work first
- ONLY stopping point: ALL tasks complete AND user clicks STOP

**Execution Pattern**:

```
Task 1 Complete → git commit → [ZERO TEXT] → read_file (Task 2 location)
Task 2 Complete → git commit → [ZERO TEXT] → read_file (Task 3 location)
...
Task N Complete → git commit → [ZERO TEXT] → Verify all complete
All Complete → Final summary (ONLY if truly nothing left)
```

---

## Common Anti-Patterns

### ❌ NEVER Do These

**Stopping Behaviors**:
- "Here's what we accomplished..." → STOP violation
- "Should I proceed with X?" → STOP violation
- "What would you like me to do next?" → STOP violation
- "All X complete. What's next?" → STOP violation
- Celebrating progress mid-workflow → STOP violation

**Documentation Issues**:
- Creating `docs/SESSION-*.md` files → Use DETAILED.md timeline instead
- Creating multiple TODO files (>6) → Consolidate or delete completed
- Leaving temp docs after extracting lessons → DELETE with audit trail
- Not updating tracking documents (issues.md, categories.md) → Append continuously

**Quality Shortcuts**:
- Marking tasks complete without running tests → Evidence required
- Skipping linting "to save time" → Quality gate mandatory
- Deferring coverage analysis → Run immediately after each task
- Assuming HEAD is clean → ALWAYS restore from baseline if uncertain

### ✅ ALWAYS Do These

**Continuous Execution**:
- Complete task → Commit → IMMEDIATELY start next task (zero pause)
- Find blocker → Document in tracking → Switch to unblocked task
- Discover new tasks → Append to tasks.md → Continue execution
- Hit stopping point → Verify LITERALLY NOTHING LEFT before asking user

**Quality Assurance**:
- Run quality gates after EVERY code change (build, lint, test, coverage)
- Provide objective evidence for task completion (test output, coverage %, commit hash)
- Update tracking documents immediately when finding issues
- Document lessons learned in DETAILED.md timeline

**Session Management**:
- Use `docs/fixes-needed-plan-tasks-v#/` for session tracking
- Append to issues.md and categories.md as problems found
- Create plan.md and tasks.md at session start
- Delete temp docs ONLY after extracting all lessons to permanent homes

---

## Workflow Integration Examples

### Beast Mode Pattern

**Structure**:

```yaml
---
description: "Autonomous continuous execution with session tracking and quality gates"
tags: ["autonomous", "quality", "session-tracking"]
---

# Instructions
1. Create session tracking in docs/fixes-needed-plan-tasks-v#/
2. Execute ALL tasks until LITERALLY NOTHING LEFT
3. Update issues.md and categories.md as problems found
4. Run quality gates after each task
5. Document lessons in DETAILED.md timeline
6. NEVER ask permission to continue
7. ONLY stopping point: User clicks STOP button
```

### Workflow Fixing Pattern

**Structure**:

```yaml
---
description: "Fix GitHub Actions workflows with systematic session tracking"
tags: ["workflow", "github-actions", "quality", "session-tracking"]
---

# Session Tracking - MANDATORY
Location: docs/fixes-needed-plan-tasks-v#/
Files: issues.md, categories.md, plan.md, tasks.md

# Quality Gates - MANDATORY
After each fix:
1. Verify workflow syntax (act --dryrun)
2. Run workflow locally (act -j job-name)
3. Check no new issues introduced
4. Update tracking documents
5. Commit with conventional message
6. IMMEDIATELY start next fix (zero pause)
```

---

## References

**Core Copilot Instructions**:
- `.github/instructions/01-02.continuous-work.instructions.md` - Continuous work directive (MANDATORY reading)
- `.github/instructions/01-03.speckit.instructions.md` - SpecKit methodology and session tracking
- `.github/instructions/06-01.evidence-based.instructions.md` - Evidence-based task completion

**Related Prompts**:
- `.github/prompts/autonomous-execution.prompt.md` - Autonomous execution patterns
- `.github/prompts/workflow-fixing.prompt.md` - Workflow fixing with session tracking
- `.github/prompts/beast-mode-3.1.prompt.md` - Comprehensive autonomous execution workflow

**Quality Standards**:
- `.github/instructions/03-02.testing.instructions.md` - Testing standards and coverage requirements
- `.github/instructions/03-07.linting.instructions.md` - Linting and code quality standards
- `.github/instructions/05-02.git.instructions.md` - Git commit conventions and workflow patterns
