---
description: "Clarify underspecified areas through targeted questions"
---

# /speckit.clarify

Clarify underspecified areas in the specification through targeted questions.

## User Input

```
$ARGUMENTS
```

You MUST consider the user input before proceeding (if not empty).

## Outline

1. **Load specification**: Read `spec.md` from the feature directory.

2. **Identify ambiguities**:
   - Find all `[NEEDS CLARIFICATION]` markers
   - Detect implicit assumptions
   - Identify missing edge case definitions
   - Flag incomplete acceptance criteria

3. **Generate clarification questions**:
   - For each ambiguity, create specific question
   - Provide context for why clarification is needed
   - Suggest reasonable default options

4. **Present questions** in priority order:
   - Critical: Blocking implementation
   - High: Affects architecture decisions
   - Medium: Affects detailed implementation
   - Low: Nice-to-have clarity

5. **Update specification**:
   - Incorporate user answers
   - Remove `[NEEDS CLARIFICATION]` markers
   - Document decisions and rationale

## Question Format

```markdown
## Q1: [Topic] (Priority: CRITICAL)

**Current state**: The specification says "[quote from spec]"

**Ambiguity**: [Explain what's unclear]

**Options**:
A) [Option with implications]
B) [Option with implications]
C) [Option with implications]

**Recommendation**: [Your suggestion based on context]
```

## Common Clarification Areas

For cryptoutil:

- Authentication method (OAuth 2.1 flows, client types)
- Data retention periods
- Error response formats
- Rate limiting thresholds
- Security event logging granularity
- Token lifetime configurations
- Supported key algorithms
