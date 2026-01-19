---
description: Create and maintain feature plan, tasks, and QUIZME documentation for SpecKit workflow
name: plan-tasks-quizme
argument-hint: <feature-name> <action: create|update|clarify|review>
agent: agent
tools:
  - read_file
  - create_file
  - replace_string_in_file
  - multi_replace_string_in_file
  - semantic_search
  - grep_search
  - list_dir
  - git
---

# Plan-Tasks-QUIZME Documentation Manager

## Purpose

This prompt helps you create, update, and maintain the complete documentation set for SpecKit feature development:

- **feature-plan.md** - High-level implementation plan with phases and technical decisions
- **feature-tasks.md** - Detailed task breakdown with dependencies and acceptance criteria
- **feature-QUIZME-v#.md** - Multiple-choice questions for UNKNOWN answers requiring user clarification

## Usage Patterns

### 1. Create New Feature Documentation

```
/plan-tasks-quizme my-feature create
```

This will:
- Create `docs/features/my-feature/feature-plan.md` based on plan template
- Create `docs/features/my-feature/feature-tasks.md` based on tasks template
- Create `docs/features/my-feature/feature-QUIZME-v1.md` for unknowns
- Initialize directory structure

### 2. Update Existing Documentation

```
/plan-tasks-quizme my-feature update
```

This will:
- Analyze current implementation status
- Update plan with actual LOE vs estimated
- Mark completed tasks in tasks.md
- Update technical decisions based on learnings

### 3. Generate Clarification Questions

```
/plan-tasks-quizme my-feature clarify
```

This will:
- Analyze plan.md and tasks.md for ambiguities
- Search codebase/instructions for KNOWN answers
- Generate QUIZME questions ONLY for UNKNOWN answers
- Format as multiple-choice with A-D options + E write-in

### 4. Review and Sync Documentation

```
/plan-tasks-quizme my-feature review
```

This will:
- Check consistency between plan, tasks, and QUIZME
- Verify all QUIZME questions are either answered or still unknown
- Move answered questions to clarify.md
- Update constitution.md and spec.md with finalized decisions

## File Templates

### feature-plan.md Structure

```markdown
# Feature Implementation Plan - <Feature Name>

**Status**: [Planning|In Progress|Complete]
**Created**: YYYY-MM-DD
**Last Updated**: YYYY-MM-DD

## Overview

[Brief description of feature]

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

### feature-tasks.md Structure

```markdown
# Feature Tasks - <Feature Name>

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
- **Description**: [What needs to be researched]
- **Acceptance Criteria**:
  - [ ] Decision documented in research.md
  - [ ] Alternatives evaluated
  - [ ] Rationale provided
- **Notes**: [Any relevant notes]

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
- [ ] Performance benchmarks meet targets
```

### feature-QUIZME-v#.md Structure

```markdown
# <Feature Name> CLARIFY-QUIZME v1

**Created**: YYYY-MM-DD
**Purpose**: Identify unknowns, risks, incompleteness before executing plan/tasks
**Scope**: [Feature scope description]

**Instructions**: Answer ALL questions. For multiple-choice, select ONE option (A-D) or provide write-in answer (E).

---

## SECTION 1: [Topic Area]

### Q1.1: [Question Title]

**Context**: [Background information explaining the decision point]

**Pattern A** (Current approach):
```
[Code or description of first approach]
```

**Pattern B** (Alternative approach):
```
[Code or description of second approach]
```

**Question**: [Clear question about which pattern/approach to use]

**A)** [Option A with rationale]
**B)** [Option B with rationale]
**C)** [Option C with rationale]
**D)** [Option D with rationale]
**E)** Write-in: ________________

**Answer**: ________________

**Follow-up**: [Clarifying sub-questions or implications]

---

### Q1.2: [Next Question Title]

[Same structure as Q1.1]

---

## SECTION 2: [Next Topic Area]

[Continue with more questions...]

---

## Status

**Open Questions**: X of Y
**Answered Questions**: Y of Y
**Ready to Proceed**: ❌ | ✅

**Next Steps**:
1. [What happens after all questions answered]
2. [How answers integrate back into plan/tasks]
```

## Workflow Steps

### Step 1: Analyze User Input

Extract:
- **Feature name** from first argument
- **Action** (create|update|clarify|review) from second argument
- **Target directory**: `docs/features/<feature-name>/` or `specs/<nnn>-<feature-name>/`

### Step 2: Search for Existing Documentation

```bash
# Check for existing plan
find docs/features specs -name "*.plan.md" -o -name "plan.md"

# Check for existing tasks
find docs/features specs -name "*.tasks.md" -o -name "tasks.md"

# Check for existing QUIZME
find docs/features specs -name "*QUIZME*.md"
```

### Step 3: Execute Action

#### CREATE Action

1. Create directory structure:
   ```
   docs/features/<feature-name>/
   ├── feature-plan.md
   ├── feature-tasks.md
   ├── feature-QUIZME-v1.md
   ├── clarify.md
   └── research.md
   ```

2. Populate from templates (see File Templates section above)

3. Initialize with placeholders for user to fill

4. Create initial QUIZME questions based on:
   - Missing technical context
   - Unclear requirements
   - Undefined dependencies
   - Ambiguous acceptance criteria

#### UPDATE Action

1. Read current plan.md and tasks.md

2. Compare with git log for actual work done:
   ```bash
   git log --oneline --grep="<feature-name>" --since="<creation-date>"
   ```

3. Update task statuses based on commits

4. Update LOE actuals from commit timestamps

5. Update technical decisions based on implementation learnings

6. Identify new unknowns that surfaced during implementation

#### CLARIFY Action

1. Read plan.md and tasks.md

2. Identify ambiguities:
   - "TBD" markers
   - "TODO" comments
   - "NEEDS CLARIFICATION" notes
   - Vague acceptance criteria
   - Missing technical decisions

3. Search codebase/instructions for answers:
   ```bash
   # Search copilot instructions
   grep -r "pattern-name" .github/instructions/

   # Search existing code
   semantic_search "similar feature implementation"

   # Check constitution/spec
   grep "architecture decision" specs/*/constitution.md
   ```

4. Generate QUIZME questions ONLY for truly UNKNOWN items

5. Format as multiple-choice with context and implications

#### REVIEW Action

1. Load all documentation files

2. Check consistency:
   - Do tasks align with plan phases?
   - Are all QUIZME questions answered?
   - Are technical decisions documented in plan?
   - Are acceptance criteria testable?

3. Identify gaps:
   - Tasks without tests
   - Phases without success criteria
   - Unanswered QUIZME questions
   - Missing risk mitigations

4. Generate report with actionable items

### Step 4: Update Related Documentation

When answers are provided:

1. **Move from QUIZME to clarify.md**:
   - Extract question and answer
   - Add to clarify.md with rationale
   - Remove from QUIZME-v#.md

2. **Update constitution.md**:
   - Add architectural decisions
   - Add constraints discovered
   - Add patterns to follow

3. **Update spec.md**:
   - Add finalized requirements
   - Add API contracts
   - Add data models

4. **Increment QUIZME version**:
   - If new unknowns surface: create QUIZME-v2.md
   - Keep old versions for history

## Best Practices

### QUIZME Question Guidelines

**CRITICAL: QUIZME is ONLY for UNKNOWN answers requiring user input**

Reference: `.github/instructions/01-03.speckit.instructions.md` lines 28-40

**MANDATORY: Search BEFORE Creating Questions**:
1. Search codebase: `semantic_search`, `grep_search`, `file_search`
2. Search copilot instructions: `.github/instructions/*.instructions.md`
3. Search existing documentation: `docs/`, `specs/*/constitution.md`, `specs/*/spec.md`
4. Search implementation: `internal/`, `cmd/`, `api/`
5. Only after exhaustive search: Add question to QUIZME

**NEVER Include Questions With Known Answers**:
- ❌ Answers found in codebase/documentation → Add to clarify.md instead
- ❌ Answers found in copilot instructions → Document in constitution.md instead
- ❌ Answers found in implementation → Document in plan.md/tasks.md instead
- ❌ Agent-provided answers in QUIZME → Violates QUIZME purpose

**Historical Lesson Learned (2025-01-16)**:
- QUIZME v4 created with 20 questions ALL having agent-provided answers
- Violated format: "DO NOT include questions for which you already know the answer"
- Corrected: Removed all 20 questions, created v5 documenting no unknowns
- Prevention: Search exhaustively BEFORE adding any question to QUIZME

**DO**:
- ✅ Search codebase/instructions EXHAUSTIVELY before adding any question
- ✅ Verify question has NO answer in existing documentation
- ✅ Provide concrete code examples in context
- ✅ Include implications of each choice
- ✅ Keep questions focused and specific
- ✅ Use multiple-choice format for guidance
- ✅ Leave Answer: field BLANK (user fills it)

**DON'T**:
- ❌ Ask questions with answers in existing documentation
- ❌ Ask questions with answers in codebase implementation
- ❌ Ask questions with answers in copilot instructions
- ❌ Pre-fill Answer: field with agent-provided answers
- ❌ Pre-fill write-in (E) answers with examples
- ❌ Ask vague or open-ended questions
- ❌ Include multiple decisions in one question
- ❌ Leave questions without context

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

## Integration with SpecKit

This prompt complements the existing SpecKit agents:

- **speckit.plan** - Delegates to this for initial plan creation
- **speckit.tasks** - Delegates to this for task breakdown
- **speckit.clarify** - Uses QUIZME format from this
- **speckit.implement** - Updates tasks during implementation

## Related Files

**Templates** (read these for structure):
- `.specify/templates/plan-template.md`
- `.specify/templates/tasks-template.md`
- `docs/speckit/SPECKIT-CLARIFY-QUIZME-TEMPLATE.md`

**Examples** (reference these):
- `docs/fixes-needed-plan-tasks/fixes-needed-PLAN.md`
- `docs/fixes-needed-plan-tasks/fixes-needed-TASKS.md`
- `docs/fixes-needed-plan-tasks/CLARIFY-QUIZME-v1.md`
- `specs/002-cryptoutil/plan.md`
- `specs/002-cryptoutil/tasks.md`

**Instructions** (follow these):
- `.github/instructions/01-03.speckit.instructions.md`
- `.github/instructions/06-01.evidence-based.instructions.md`

## Variables

You can reference these variables in your input:

- `${workspaceFolder}` - Root directory of workspace
- `${file}` - Currently open file
- `${input:featureName}` - Prompt for feature name
- `${input:action}` - Prompt for action (create|update|clarify|review)

## Example Usage

**Create new feature documentation**:
```
/plan-tasks-quizme multi-tenancy create
```

**Update existing feature progress**:
```
/plan-tasks-quizme service-template update
```

**Generate clarification questions**:
```
/plan-tasks-quizme oauth-integration clarify
```

**Review documentation consistency**:
```
/plan-tasks-quizme identity-migration review
```

## Output Format

Always provide:

1. **Summary** of what was created/updated
2. **File paths** to all affected documents
3. **Next steps** for the user
4. **Warnings** about any inconsistencies found
5. **Statistics** (tasks complete, questions answered, etc.)
