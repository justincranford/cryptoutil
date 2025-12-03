---
description: "Define what you want to build - requirements and user stories"
---

# /speckit.specify

Define what you want to build by creating a feature specification with user stories and requirements.

## User Input

```
$ARGUMENTS
```

You MUST consider the user input before proceeding (if not empty).

## Outline

1. **Setup**: Identify or create the feature directory in `specs/XXX-feature-name/`.

2. **Feature analysis**:
   - Parse user input for feature requirements
   - Identify user personas and their needs
   - Extract functional requirements
   - Identify edge cases and error scenarios

3. **Generate specification** using `templates/spec-template.md`:
   - Feature name and branch
   - User scenarios with priorities (P1, P2, P3...)
   - Acceptance criteria (Given/When/Then format)
   - Functional requirements (FR-001, FR-002...)
   - Key entities (if data is involved)
   - Non-functional requirements

4. **Mark ambiguities** with `[NEEDS CLARIFICATION: specific question]`:
   - Never guess at unclear requirements
   - Document all assumptions explicitly

5. **Output**: Write `spec.md` to the feature directory.

## Specification Quality Checklist

Before completing:

- [ ] No `[NEEDS CLARIFICATION]` markers remain (or explained)
- [ ] Requirements are testable and unambiguous
- [ ] Success criteria are measurable
- [ ] User stories are prioritized (P1 = MVP)
- [ ] Each story is independently testable
- [ ] Edge cases are identified

## User Story Format

```markdown
### User Story N - [Brief Title] (Priority: PN)

[Describe this user journey in plain language]

**Why this priority**: [Explain the value and why it has this priority level]

**Independent Test**: [Describe how this can be tested independently]

**Acceptance Scenarios**:

1. **Given** [initial state], **When** [action], **Then** [expected outcome]
```

## cryptoutil-Specific Guidance

For cryptoutil features:

- Reference constitution principles in requirements
- Map to existing product areas (JOSE, Identity, KMS, Certificates)
- Consider FIPS 140-3 compliance requirements
- Include security considerations for all features
- Reference existing API patterns in `api/` directory
