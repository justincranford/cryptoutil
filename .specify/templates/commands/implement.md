---
description: "Execute the implementation plan by processing and executing all tasks defined in tasks.md"
---

# /speckit.implement

Execute the implementation plan by processing and executing all tasks defined in tasks.md.

## User Input

```
$ARGUMENTS
```

You MUST consider the user input before proceeding (if not empty).

## Outline

1. **Setup**: Navigate to the feature directory (specs/001-cryptoutil/) and parse available documents.

2. **Check checklists status** (if FEATURE_DIR/checklists/ exists):
   - Scan all checklist files in the checklists/ directory
   - For each checklist, count:
     - Total items: All lines matching `- [ ]` or `- [X]` or `- [x]`
     - Completed items: Lines matching `- [X]` or `- [x]`
     - Incomplete items: Lines matching `- [ ]`
   - Create a status table showing completion percentages
   - If any checklist is incomplete:
     - Display the table with incomplete item counts
     - Ask: "Some checklists are incomplete. Do you want to proceed anyway? (yes/no)"
     - Wait for user response before continuing
   - If all checklists are complete: Automatically proceed

3. **Load and analyze the implementation context**:
   - REQUIRED: Read tasks.md for the complete task list and execution plan
   - REQUIRED: Read plan.md for tech stack, architecture, and file structure
   - IF EXISTS: Read data-model.md for entities and relationships
   - IF EXISTS: Read contracts/ for API specifications and test requirements
   - IF EXISTS: Read research.md for technical decisions and constraints
   - IF EXISTS: Read quickstart.md for integration scenarios

4. **Project Setup Verification**:
   - Verify .dockerignore exists and has proper patterns
   - Verify go.mod and go.sum are valid
   - Check that golangci-lint configuration is present
   - Verify constitution compliance requirements

5. **Parse tasks.md structure** and extract:
   - Task phases: Setup, Tests, Core, Integration, Polish
   - Task dependencies: Sequential vs parallel execution rules
   - Task details: ID, description, file paths, parallel markers `[P]`
   - Execution flow: Order and dependency requirements

6. **Execute implementation** following the task plan:
   - Phase-by-phase execution: Complete each phase before moving to the next
   - Respect dependencies: Run sequential tasks in order, parallel tasks `[P]` can run together
   - Follow TDD approach: Execute test tasks before their corresponding implementation tasks
   - File-based coordination: Tasks affecting the same files must run sequentially
   - Validation checkpoints: Verify each phase completion before proceeding

7. **Implementation execution rules**:
   - Setup first: Initialize project structure, dependencies, configuration
   - Tests before code: If you need to write tests for contracts, entities, and integration scenarios
   - Core development: Implement models, services, CLI commands, endpoints
   - Integration work: Database connections, middleware, logging, external services
   - Polish and validation: Unit tests, performance optimization, documentation

8. **Progress tracking and error handling**:
   - Report progress after each completed task
   - Halt execution if any non-parallel task fails
   - For parallel tasks `[P]`, continue with successful tasks, report failed ones
   - Provide clear error messages with context for debugging
   - Suggest next steps if implementation cannot proceed
   - **IMPORTANT**: For completed tasks, mark the task off as `[X]` in the tasks file

9. **Completion validation**:
   - Verify all required tasks are completed
   - Check that implemented features match the original specification
   - Validate that tests pass and coverage meets requirements (85%+ production code)
   - Confirm the implementation follows the technical plan
   - Run `golangci-lint run --fix` to ensure code quality
   - Report final status with summary of completed work

## cryptoutil-Specific Execution Rules

For cryptoutil Go project implementation:

### Build Verification

After each implementation task:

```bash
go build ./...
golangci-lint run --fix
```

### Test Execution

After test-related tasks:

```bash
go test ./... -cover
```

### Code Quality Gates

Before marking any phase complete:

1. All code builds without errors
2. All existing tests pass
3. golangci-lint passes with no errors
4. Coverage targets maintained (85%+ production, 90%+ infrastructure)

### Commit Strategy

- Commit after each logical task or group of related tasks
- Use conventional commit format: `type(scope): description`
- Example: `feat(identity): implement login UI template`

### Progress Tracking

Update PROGRESS.md in the spec directory after each task with:

- Task ID and description
- Completion timestamp
- Any issues encountered
- Post-mortem notes (gaps, bugs, improvements needed)

## Error Recovery

If implementation fails:

1. Document the error in PROGRESS.md
2. Identify root cause
3. Update tasks.md if task needs modification
4. Resume from the failed task

## Note

This command assumes a complete task breakdown exists in tasks.md. If tasks are incomplete or missing, suggest running `/speckit.tasks` first to regenerate the task list.
