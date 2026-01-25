---
description: Create and maintain simple plan.md and tasks.md documentation for custom plans
name: plan-tasks-quizme
argument-hint: <directory-path> <action: create|update|review>
tools:
  - search/codebase
  - search
  - edit/editFiles
  - execute/runInTerminal
---

# Plan-Tasks Documentation Manager (Custom Plans)

## Purpose

This prompt helps you create, update, and maintain **simple custom plans** (NOT SpecKit - for SpecKit use `/speckit.*` agents).

**Custom plans** use **2 files**:

- **plan.md** - High-level implementation plan with phases and decisions
- **tasks.md** - Detailed task breakdown with acceptance criteria

**User must specify directory path** where plan.md and tasks.md will be created/updated.

## Directory Path Guidelines

**Existing Examples**:

- `docs\fixes-needed-plan-tasks\` (plan.md + tasks.md)
- `docs\fixes-needed-plan-tasks-v2\` (plan.md + tasks.md)

**Future Examples** (user specifies):

- `docs\small-feature\` (plan.md + tasks.md)
- `docs\simple-plan\` (plan.md + tasks.md)
- `docs\short-term-work\` (plan.md + tasks.md)
- `docs\feature-name\` (plan.md + tasks.md)

**Pattern**: Short directory name under `docs\`, containing ONLY `plan.md` and `tasks.md`

## SpecKit vs Custom Plans - CRITICAL DISTINCTION

**This prompt is for CUSTOM PLANS, NOT SpecKit**:

- ✅ Custom Plans: `docs\<directory>\` with plan.md + tasks.md (2 files)
- ❌ SpecKit: `specs\<nnn>-<project>\` with constitution.md + spec.md + clarify.md + plan.md + tasks.md (5+ files)

**For SpecKit**, use `/speckit.*` agents instead (see `.github/prompts/doc-sync.prompt.md` for differences).

## Usage Patterns

### 1. Create New Custom Plan

```
/plan-tasks-quizme docs\my-work\ create
```

This will:

- Create `docs\my-work\plan.md` from template
- Create `docs\my-work\tasks.md` from template
- Initialize directory if needed

### 2. Update Existing Plan

```
/plan-tasks-quizme docs\my-work\ update
```

This will:

- Analyze implementation status
- Update plan with actual LOE vs estimated
- Mark completed tasks in tasks.md
- Update decisions based on learnings

### 3. Review Documentation

```
/plan-tasks-quizme docs\my-work\ review
```

This will:

- Check consistency between plan.md and tasks.md
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
   docs/features/<feature-name>/
   ├── feature-plan.md
   ├── feature-tasks.md
   ├── feature-QUIZME-v1.md
   ├── clarify.md
   └── research.md
   ```

2. Create `plan.md` from template

3. Create `tasks.md` from template

4. Initialize with placeholders

#### UPDATE Action

1. Read current plan.md and tasks.md

2. Check git log for work done:

   ```bash
   git log --oneline --since="<creation-date>"
   ```

3. Update task statuses based on commits

4. Update LOE actuals from commit timestamps

5. Update technical decisions based on learnings

#### REVIEW Action

1. Load plan.md and tasks.md

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

## Related Files

**Examples**:

- `docs/fixes-needed-plan-tasks/plan.md`
- `docs/fixes-needed-plan-tasks/tasks.md`
- `docs/fixes-needed-plan-tasks-v2/plan.md`
- `docs/fixes-needed-plan-tasks-v2/tasks.md`

**For SpecKit** (different workflow):

- See `.github/prompts/doc-sync.prompt.md` for SpecKit vs Custom distinction
- Use `/speckit.*` agents for full SpecKit workflow

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
