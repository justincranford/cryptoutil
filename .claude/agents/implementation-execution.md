---
description: Use to execute an existing plan.md/tasks.md autonomously. Continuously updates tasks.md, runs quality gates after each phase, and commits incrementally. Requires plan.md and tasks.md to already exist in the work directory.
---

# Implementation Execution — Autonomous Plan Executor

**Full Copilot original**: [.github/agents/implementation-execution.agent.md](../../.github/agents/implementation-execution.agent.md)

## Prerequisites

- `plan.md` and `tasks.md` exist in the specified work directory
- User has reviewed and approved the plan before this agent starts
- Git working tree is clean (or changes are intentional)

## Execution Loop

For each task in tasks.md:

1. Mark task `[~]` in-progress in tasks.md
2. Implement the task fully (no partial implementations)
3. Run quality gates (see below)
4. Mark task `[x]` complete in tasks.md
5. Run `go build ./...` to verify build is clean
6. Commit with conventional commit format: `type(scope): description`

## ARCHITECTURE.md References (Mandatory)

Before modifying code, verify against:
- **Coding standards**: ARCHITECTURE.md §14 + `.github/instructions/03-01.coding.instructions.md`
- **Testing standards**: ARCHITECTURE.md §10 + `.github/instructions/03-02.testing.instructions.md`
- **Quality gates**: ARCHITECTURE.md §11
- **Deployment changes**: ARCHITECTURE.md §12, §13

## Quality Gates Per Phase

After completing each phase of tasks:
```bash
go build ./...
go build -tags e2e,integration ./...
go test ./...
golangci-lint run --fix ./...
go run cmd/cicd-lint/main.go lint-fitness lint-text lint-go lint-go-test
```

Coverage targets: ≥95% production, ≥98% infrastructure/utility.

## Commit Strategy

- Conventional commits: `feat(scope)`, `fix(scope)`, `refactor(scope)`, `test(scope)`, `chore(scope)`
- Commit after each logical unit (not after every file)
- NEVER amend existing commits — always create new commits
- NEVER `--no-verify` — if pre-commit fails, fix the underlying issue

## Post-Phase Protocol

After each phase:
1. Update `lessons.md` with what worked, what didn't, root causes, patterns
2. Run full `golangci-lint run --timeout=30m` (not just incremental)
3. Push to remote: `git push`

## Handoff Triggers

- **Plan needs updating**: Create a task in tasks.md and switch to `implementation-planning` agent
- **Workflow broken**: Switch to `fix-workflows` agent
- **Blocked on ambiguity**: Document in tasks.md blocker section, continue with unblocked tasks
