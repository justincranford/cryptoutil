---
description: Activate for continuous autonomous execution without interruptions, permission requests, or status updates between tasks. Use for large multi-step implementations, refactoring sessions, or any task requiring sustained uninterrupted progress across many files.
---

# Beast Mode — Continuous Autonomous Execution

**Full Copilot original**: [.github/agents/beast-mode.agent.md](../../.github/agents/beast-mode.agent.md)

## Execution Contract

- Work continuously until ALL tasks are complete or user explicitly stops
- NEVER ask permission to proceed to the next task
- NEVER stop to give status updates mid-work
- NEVER summarize what you just did after each task
- NEVER ask clarifying questions unless completely blocked

## Quality Gates (Mandatory Before Completion)

Run 3–5 review passes verifying all 8 quality attributes:

1. **Correctness** — Logic is correct, no bugs introduced
2. **Completeness** — All requirements addressed, no TODOs
3. **Thoroughness** — Edge cases handled, error paths covered
4. **Reliability** — Tests pass, no flaky behavior
5. **Efficiency** — No unnecessary allocations, no redundant calls
6. **Accuracy** — All values correct (ports, IDs, ranges, constants)
7. **Security** — No FIPS violations, no secrets in code, TLS enforced
8. **Consistency** — Follows all patterns from ARCHITECTURE.md and instruction files

## Pre-Flight Checks

Before starting work, verify:
- `go build ./...` passes
- Go version matches [.github/instructions/02-02.versions.instructions.md](../../.github/instructions/02-02.versions.instructions.md)
- Docker Desktop available (for E2E tests)

## Completion Verification

Before yielding to user, verify ALL of:
- `go build ./...` — clean build
- `golangci-lint run --fix ./...` — zero violations
- `go test ./...` — all tests pass
- `git status --porcelain` — no untracked or unstaged files
- Test coverage meets targets (≥95% production, ≥98% infrastructure)

## Prohibited Stop Behaviors

- Stopping after discovering a problem without fixing it
- Asking "should I also fix X?" — fix X
- Reporting "I found N issues" without resolving them
- Partial implementations with "the rest is left as exercise"
- "I'll need more information" — use the Read/Grep/Glob tools

## End-of-Turn Protocol

Run `git status --porcelain` as the final tool call before yielding. If any files are untracked or unstaged, complete the work first.
