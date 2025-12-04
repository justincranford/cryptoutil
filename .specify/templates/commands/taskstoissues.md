---
description: "Convert tasks to GitHub Issues for project tracking"
---

# /speckit.taskstoissues

Convert tasks from tasks.md into GitHub Issues for project tracking.

## User Input

```
$ARGUMENTS
```

You MUST consider the user input before proceeding (if not empty).

## Outline

1. **Load tasks**: Read `tasks.md` from the feature directory.

2. **Parse task entries**:
   - Extract task ID, description, priority, status
   - Identify dependencies between tasks
   - Group by phase/milestone

3. **Generate GitHub Issues**:
   - Create issue for each task
   - Add appropriate labels (priority, phase, type)
   - Include acceptance criteria from plan.md
   - Link to specification requirements

4. **Create milestone structure**:
   - Phase 1, Phase 2, etc. as milestones
   - Assign issues to milestones
   - Set target dates if available

5. **Output issue definitions** in format suitable for `gh issue create`.

## Issue Template

```markdown
## Task: [Task ID] - [Task Description]

**Priority**: [CRITICAL/HIGH/MEDIUM/LOW]
**Phase**: [Phase number]
**Story Points**: [Estimated points]

### Description

[Detailed task description from tasks.md]

### Acceptance Criteria

- [ ] [Criterion 1 from plan.md]
- [ ] [Criterion 2 from plan.md]

### Dependencies

- Blocked by: [Task IDs if any]
- Blocks: [Task IDs if any]

### Related

- Spec: `specs/001-cryptoutil/spec.md#section`
- Plan: `specs/001-cryptoutil/plan.md#section`
```

## Label Mapping

| Priority | Label |
|----------|-------|
| CRITICAL | `priority:critical` |
| HIGH | `priority:high` |
| MEDIUM | `priority:medium` |
| LOW | `priority:low` |

| Phase | Label |
|-------|-------|
| 1 | `phase:1-identity` |
| 2 | `phase:2-kms` |
| 3 | `phase:3-integration` |
| 4 | `phase:4-ca` |

## Command Output

```bash
# Phase 1 Issues
gh issue create --title "P1.1.1 - Create minimal HTML login template" \
  --body "..." \
  --label "priority:high,phase:1-identity,type:feature" \
  --milestone "Phase 1: Identity V2"

gh issue create --title "P1.1.2 - Add minimal CSS styling" \
  --body "..." \
  --label "priority:medium,phase:1-identity,type:feature" \
  --milestone "Phase 1: Identity V2"
```
