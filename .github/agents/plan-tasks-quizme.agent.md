---
description: Create and maintain simple plan.md and tasks.md documentation for custom plans
name: plan-tasks-quizme
argument-hint: <directory-path> <action: create|update|review>
tools:
	- edit/editFiles
	- execute/createAndRunTask
	- execute/getTerminalOutput
	- execute/runInTerminal
	- execute/runNotebookCell
	- execute/runTask
	- execute/testFailure
	- read/getNotebookSummary
	- read/getTaskOutput
	- read/problems
	- read/readNotebookCellOutput
	- read/terminalLastCommand
	- read/terminalSelection
	- search
	- search/changes
	- search/codebase
	- search/searchResults
	- search/usages
	- vscode/extensions
	- vscode/getProjectSetupInfo
	- vscode/installExtension
	- vscode/newWorkspace
	- vscode/openSimpleBrowser
	- vscode/runCommand
	- vscode/vscodeAPI
	- web/fetch
	- web/githubRepo
---

# Plan-Tasks Documentation Manager (Custom Plans)

## Purpose

This prompt helps you create, update, and maintain **simple custom plans**.

**Custom plans** use **2 input files** (created/updated by this agent):

- **`<work-dir>/plan.md`** - High-level implementation plan with phases and decisions
- **`<work-dir>/tasks.md`** - Detailed task breakdown with acceptance criteria

**Plus optional quizme file** (ephemeral, deleted after answers merged):

- **`<work-dir>/quizme-v#.md`** - Questions to clarify unknowns, risks, inefficiencies ONLY
  - Format: A-D multiple choice + E (blank fill-in) with blank Choice field
  - Temporary - deleted after answers merged into plan.md/tasks.md

**User must specify directory path** where files will be created/updated.

## Directory Path Guidelines

**Existing Examples**:

- `docs\fixes-needed-plan-tasks\` (plan.md + tasks.md)
- `docs\fixes-needed-plan-tasks-v2\` (plan.md + tasks.md)

**Future Examples** (user specifies):

- `docs\small-feature\` (plan.md + tasks.md)
- `docs\simple-plan\` (plan.md + tasks.md)
- `docs\short-term-work\` (plan.md + tasks.md)
- `docs\feature-name\` (plan.md + tasks.md)

**Pattern**: Short directory name under `docs\`, containing files: plan.md, tasks.md, and optionally quizme-v#.md

## Usage Patterns

### 1. Create New Custom Plan

```
/plan-tasks-quizme <work-dir> create
```

This will:

- Create `<work-dir>/plan.md` from template
- Create `<work-dir>/tasks.md` from template
- Optionally create `<work-dir>/quizme-v1.md` for unknowns/risks/inefficiencies
  - A-D and E (blank fill-in) questions
  - Choice field blank for user to answer
- Initialize directory if needed

### 2. Update Existing Plan

```
/plan-tasks-quizme <work-dir> update
```

This will:

- Analyze implementation status
- Update `<work-dir>/plan.md` with actual LOE vs estimated
- Mark completed tasks in `<work-dir>/tasks.md`
- Update decisions based on learnings
- Merge quizme answers if `<work-dir>/quizme-v#.md` exists (then delete it)

### 3. Review Documentation

```
/plan-tasks-quizme <work-dir> review
```

This will:

- Check consistency between `<work-dir>/plan.md` and `<work-dir>/tasks.md`
- Verify task completion status
- Identify gaps or inconsistencies

## File Templates

### plan.md Structure

```markdown
# Implementation Plan - <Plan Name>

**Status**: [Planning|In Progress|Complete]
**Created**: YYYY-MM-DD
**Last Updated**: YYYY-MM-DD

## Overview

[Brief description of work]

## Technical Context

- **Language**: Go 1.25.5
- **Framework**: [Framework if applicable]
- **Database**: PostgreSQL OR SQLite with GORM
- **Dependencies**: [Key dependencies]

## Phases

### Phase 0: Research & Discovery (Xh)
- Research unknowns identified in Technical Context
- Document decisions in research.md
- Resolve all "NEEDS CLARIFICATION" items

### Phase 1: Foundation (Xh)
- Database schema design
- Domain model implementation
- Repository layer with tests

### Phase 2: Business Logic (Xh)
- Service layer implementation
- Validation rules
- Unit tests (≥95% coverage)

### Phase 3: API Layer (Xh)
- HTTP handlers
- OpenAPI spec
- Integration tests

### Phase 4: E2E Testing (Xh)
- Docker Compose setup
- E2E test scenarios
- Performance testing

## Technical Decisions

### Decision 1: [Topic]
- **Chosen**: [What was chosen]
- **Rationale**: [Why chosen]
- **Alternatives**: [What else considered]
- **Impact**: [Implications]

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| [Risk description] | Low/Med/High | Low/Med/High | [Mitigation strategy] |

## Quality Gates

- ✅ All tests pass (`runTests`)
- ✅ Coverage ≥95% production, ≥98% infrastructure
- ✅ Mutation testing ≥85% (early), ≥98% (infrastructure)
- ✅ Linting clean (`golangci-lint run`)
- ✅ No new TODOs without tracking
- ✅ Docker Compose E2E passes

## Success Criteria

- [ ] All phases complete
- [ ] Quality gates pass
- [ ] E2E demo functional
- [ ] Documentation updated
- [ ] CI/CD green
```

### tasks.md Structure

```markdown
# Tasks - <Plan Name>

**Status**: X of Y tasks complete (Z%)
**Last Updated**: YYYY-MM-DD

## Task Checklist

### Phase 0: Research

#### Task 0.1: Research [Topic]
- **Status**: ❌ Not Started | ⚠️ In Progress | ✅ Complete
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: [What needs research]
- **Acceptance Criteria**:
  - [ ] Decision documented
  - [ ] Alternatives evaluated
  - [ ] Rationale provided

### Phase 1: Foundation

#### Task 1.1: Database Schema
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**:
- **Dependencies**: Task 0.1
- **Description**: Design and implement database schema
- **Acceptance Criteria**:
  - [ ] Migrations created (up/down)
  - [ ] Schema documented
  - [ ] Constraints defined
  - [ ] Indexes planned
- **Files**:
  - `internal/domain/migrations/0001_init.up.sql`
  - `internal/domain/migrations/0001_init.down.sql`

#### Task 1.2: Domain Models
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Task 1.1
- **Description**: Implement domain entities and value objects
- **Acceptance Criteria**:
  - [ ] Models with GORM tags
  - [ ] Validation methods
  - [ ] Tests with ≥95% coverage
- **Files**:
  - `internal/domain/models.go`
  - `internal/domain/models_test.go`

[Continue for all tasks...]

## Cross-Cutting Tasks

### Documentation
- [ ] README.md updated
- [ ] API documentation generated
- [ ] Architecture diagrams created

### Testing
- [ ] Unit tests ≥95% coverage
- [ ] Integration tests pass
- [ ] E2E tests pass
- [ ] Mutation testing ≥85%

### Quality
- [ ] Linting passes
- [ ] No security vulnerabilities
```

## Workflow Steps

### Step 1: Analyze User Input

Extract:

- **Directory path** from first argument (e.g., `docs\my-work\`)
- **Action** (create|update|review) from second argument

### Step 2: Search for Existing Documentation

```bash
# Check for existing plan in specified directory
ls <directory-path>/plan.md

# Check for existing tasks in specified directory
ls <directory-path>/tasks.md
```

### Step 3: Execute Action

#### CREATE Action

1. Create directory if needed

   ```
   <work-dir>/
   ├── plan.md
   ├── tasks.md
   └── quizme-v1.md (optional, ephemeral)
   ```

2. Create `<work-dir>/plan.md` from template

3. Create `<work-dir>/tasks.md` from template

4. Optionally create `<work-dir>/quizme-v#.md` for unknowns/risks/inefficiencies ONLY
   - Contains A-D and E (blank fill-in) multiple choice questions
   - Contains choice field left blank for user to fill
   - ONLY for: unknowns, risks, inefficiencies that need clarification
   - Ephemeral - deleted after answers merged into plan.md/tasks.md

5. Initialize with placeholders

#### UPDATE Action

1. Read current `<work-dir>/plan.md` and `<work-dir>/tasks.md`

2. Check git log for work done:

   ```bash
   git log --oneline --since="<creation-date>"
   ```

3. Update task statuses based on commits

4. Update LOE actuals from commit timestamps

5. Update technical decisions based on learnings

#### REVIEW Action

1. Load `<work-dir>/plan.md` and `<work-dir>/tasks.md`

2. Check consistency:
   - Do tasks align with plan phases?
   - Are technical decisions documented?
   - Are acceptance criteria testable?

3. Identify gaps:
   - Tasks without tests
   - Phases without success criteria
   - Missing risk mitigations

4. Generate report with actionable items

## Best Practices

### Plan/Tasks Syncing

**Maintain bidirectional links**:

- Plan phases → Task groups
- Technical decisions → Affected tasks
- Risks → Mitigation tasks
- Quality gates → Verification tasks

### Evidence-Based Updates

**NEVER mark tasks complete without**:

- ✅ Git commits referencing task
- ✅ Tests passing with coverage
- ✅ Linting clean
- ✅ Acceptance criteria verified

### Quizme File Purpose

**Only create `<work-dir>/quizme-v#.md` for**:

- ✅ Unknowns that need clarification before planning
- ✅ Risks that need assessment
- ✅ Inefficiencies that need decision

**Quizme Format** (A-D and E blank fill-in):

- Multiple choice questions A-D with one correct answer
- Option E: blank fill-in for custom answer
- Choice field: LEFT BLANK for user to select/fill

**After user answers**: Merge into plan.md/tasks.md, DELETE quizme-v#.md

## Related Files

**Examples**:

- `<work-dir>/plan.md` - High-level implementation plan
- `<work-dir>/tasks.md` - Detailed task breakdown
- `<work-dir>/quizme-v#.md` - Optional questions for unknowns/risks/inefficiencies (ephemeral)

**Instructions**:

- `.github/instructions/06-01.evidence-based.instructions.md`

## Example Usage

**Create new custom plan**:

```
/plan-tasks-quizme docs\database-migration\ create
```

**Update existing plan**:

```
/plan-tasks-quizme docs\fixes-needed-plan-tasks\ update
```

**Review consistency**:

```
/plan-tasks-quizme docs\my-work\ review
```

## Output Format

Always provide:

1. **Summary** of what was created/updated
2. **File paths** to affected documents
3. **Next steps** for user
4. **Warnings** about inconsistencies
5. **Statistics** (tasks complete, etc.)
