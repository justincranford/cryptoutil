---
description: "Cross-artifact consistency and coverage analysis"
---

# /speckit.analyze

Run cross-artifact consistency and coverage analysis on spec-kit documents.

## User Input

```
$ARGUMENTS
```

You MUST consider the user input before proceeding (if not empty).

## Outline

1. **Load all artifacts** from the feature directory:
   - `spec.md` - User stories and requirements
   - `plan.md` - Implementation plan
   - `tasks.md` - Task breakdown
   - `data-model.md` - Entity definitions (if exists)
   - `contracts/` - API specifications (if exists)

2. **Requirement coverage analysis**:
   - Map each requirement (FR-XXX) to implementing tasks
   - Identify requirements without corresponding tasks
   - Flag tasks that don't trace to requirements

3. **User story coverage analysis**:
   - Verify each user story has implementation tasks
   - Check acceptance criteria have test coverage
   - Validate stories are independently implementable

4. **Consistency checks**:
   - Entity names match between data-model.md and implementation
   - API endpoints in contracts/ match route definitions in tasks
   - Technical decisions in research.md reflected in plan.md

5. **Gap identification**:
   - Missing error handling scenarios
   - Undocumented edge cases
   - Security requirements without implementations

6. **Generate report**:
   - Coverage matrix (requirements → tasks)
   - Gap list with severity levels
   - Recommendations for addressing gaps

## Output Format

```markdown
# Consistency Analysis Report

## Coverage Matrix

| Requirement | Tasks | Status |
|-------------|-------|--------|
| FR-001 | T012, T015 | ✅ Covered |
| FR-002 | - | ❌ GAP |

## Identified Gaps

1. **HIGH**: [Gap description]
2. **MEDIUM**: [Gap description]

## Recommendations

1. [Specific action to address gap]
```

## When to Run

- After `/speckit.tasks` to validate task completeness
- Before `/speckit.implement` to ensure full coverage
- After major specification changes
