---
description: Use to create or update plan.md, tasks.md, and lessons.md scaffold for a non-trivial implementation task. Creates phased plans with scope, LOE, rationale, and detailed task breakdowns before any code is written.
---

# Implementation Planning — Plan + Tasks + Lessons

**Full Copilot original**: [.github/agents/implementation-planning.agent.md](../../.github/agents/implementation-planning.agent.md)

## Output Files

Create in `<work-dir>/` (e.g., `docs/framework-v7/` or `docs/<feature>/`):

| File | Purpose | Who Updates |
|------|---------|-------------|
| `plan.md` | Phase plan: scope, LOE, rationale, constraints, risks | Planning creates; execution updates |
| `tasks.md` | Task breakdown with `[ ]`/`[~]`/`[x]` checkbox tracking | Execution updates continuously |
| `lessons.md` | Post-mortem: what worked, root causes, patterns observed | Execution populates after each phase |

## Unknowns Clarification (quizme)

For ambiguous requirements, create `quizme-v1.md` with A–D multiple-choice + E blank for each unknown. Present to user, merge answers into plan.md, then delete quizme file.

## plan.md Structure

```markdown
# [Feature] Implementation Plan

## Phase 1: [Name]
**Scope**: What is included and explicitly excluded
**LOE**: S/M/L/XL estimate with rationale
**Rationale**: Why this phase ordering
**Dependencies**: What must be done first
**Risks**: What could go wrong

### Tasks
- [ ] Task 1: specific, actionable, single-responsibility
- [ ] Task 2: ...

## Phase 2: [Name]
...
```

## tasks.md Structure

```markdown
# Tasks

## Phase 1: [Name]
- [ ] 1.1 Description (file: path/to/file.go)
- [~] 1.2 In progress
- [x] 1.3 Completed

## Phase 2: [Name]
...

## Blockers
- None
```

## ARCHITECTURE.md References (Mandatory)

Planning MUST reference:
- §10 Testing architecture — test patterns for new code
- §11 Quality strategy — coverage and mutation targets
- §14 Development practices — coding standards
- §2.1 Agent catalog — which agents to use for execution
- §13.4 Documentation propagation — if docs are changing

## Constraints

- Plans MUST be reviewable by the user before execution starts
- Each task MUST reference the specific file(s) it modifies
- LOE estimates MUST include rationale (not just S/M/L)
- Risks MUST include mitigation strategy
