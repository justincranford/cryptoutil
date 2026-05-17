# Enhancements For `claude-fix-workflows`

Created: 2026-05-17

## Summary

The workflow-fixer agent is strong on local testing and Docker awareness, but it can be made more efficient by reducing duplicated process text and by adding a clearer path for timeouts, flaky runs, and log analysis.

## What To Keep

- The `go run ./cmd/cicd-workflow -workflows=<name>` rule is the right local validation entry point.
- Docker Desktop checks are necessary for service-dependent workflows.
- The separation between workflow fixes and related artifact fixes is good.

## What To Modify

### 1. Replace Sleep-Based Guidance With Polling

The current startup guidance includes fixed waits. Those should be replaced with explicit polling or readiness checks wherever possible.

Why:

- fixed sleeps waste time when Docker is already ready
- fixed sleeps still fail when Docker is slower than expected
- polling shortens the fast path and keeps the slow path bounded

### 2. Add A Log-First Triage Loop

The agent should explicitly say how to triage a failing workflow:

1. capture the failed step
2. identify the first failing command
3. classify it as syntax, config, dependency, timeout, or test failure
4. reproduce locally with the smallest command that fails

That would improve speed and reduce random command reruns.

### 3. Make Artifact Capture A First-Class Step

Workflow fixes are easier to review when the local evidence is saved in a consistent place. The agent should point to one canonical artifact layout under `workflow-reports/` and encourage concise filenames that encode workflow name and date.

### 4. Separate Fast Validation From Slow Validation

The agent already lists workflow types, but it should make the sequencing clearer:

- syntax and build checks first
- Docker-dependent checks second
- only then run the slower end-to-end or load workflow

That order reduces churn and makes the fastest failures show up first.

### 5. Make Commit Boundaries More Explicit

If a workflow fix requires changes to compose files, Dockerfiles, and pre-commit config, the agent should emphasize separate commits for separate artifact families. That is already implied, but it should be sharper because mixed fixes are hard to review.

## What To Remove Or De-Emphasize

- Remove repeated restatements of the same local-test rule.
- De-emphasize the generic motivational language.
- Trim the long platform startup instructions so the Windows path and the Docker readiness check are the only truly essential parts.

## Suggested Additions

- A workflow-failure decision tree that maps common symptoms to the next command.
- A short section that distinguishes workflow syntax errors from code/test failures.
- A note that if a local run succeeds but CI still fails, the next step is to compare environment-dependent inputs, not simply rerun the same command.

## Net Effect

The agent would stay workflow-specific, but it would become faster to operate during real incident response because the next command would be clearer and the slow waits would be reduced.
