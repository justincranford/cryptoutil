---
description: "Create technical implementation plans with your chosen tech stack"
---

# /speckit.plan

Create a comprehensive technical implementation plan from a feature specification.

## User Input

```
$ARGUMENTS
```

You MUST consider the user input before proceeding (if not empty).

## Outline

1. **Read specification**: Load `spec.md` from the feature directory.

2. **Analyze requirements**:
   - Extract user stories and priorities
   - Identify technical implications
   - Map requirements to implementation tasks

3. **Generate implementation plan** using `templates/plan-template.md`:
   - Summary of primary requirement and technical approach
   - Technical context (language, dependencies, storage, testing)
   - Constitution check gates
   - Project structure (documentation and source code)
   - Implementation phases with tasks

4. **Constitution compliance**: Verify plan adheres to:
   - FIPS 140-3 cryptographic requirements
   - Code quality standards
   - Evidence-based completion criteria
   - Security architecture principles

5. **Generate supporting documents** (if needed):
   - `research.md` - Technical decisions and library comparisons
   - `data-model.md` - Entity definitions and relationships
   - `contracts/` - API specifications
   - `quickstart.md` - Key validation scenarios

6. **Output**: Write `plan.md` to the feature directory.

## Plan Quality Checklist

Before completing:

- [ ] All `[NEEDS CLARIFICATION]` from spec addressed
- [ ] Technology choices documented with rationale
- [ ] Clear phase boundaries defined
- [ ] Dependencies between phases identified
- [ ] Success criteria for each phase
- [ ] Constitution gates passed

## cryptoutil Technical Context

For cryptoutil plans, default context:

- **Language/Version**: Go 1.25.4+
- **Primary Dependencies**: Fiber v2, GORM, lestrrat-go/jwx, google/uuid
- **Storage**: PostgreSQL (production), SQLite (development/testing)
- **Testing**: go test with testify, table-driven tests, t.Parallel()
- **Target Platform**: Linux containers, cross-compiled for darwin/windows
- **Performance Goals**: Sub-100ms API response times
- **Constraints**: FIPS 140-3 compliance, no CGO

## Project Structure

cryptoutil follows Standard Go Project Layout:

```text
specs/001-cryptoutil/
├── plan.md              # Implementation plan
├── spec.md              # Feature specification
├── tasks.md             # Task breakdown
├── data-model.md        # Entity definitions (if needed)
├── research.md          # Technical decisions (if needed)
├── quickstart.md        # Validation scenarios (if needed)
└── contracts/           # API specifications (if needed)
```

Source code structure:

```text
internal/
├── identity/            # P2: OAuth 2.1 / OIDC
├── kms/                 # P3: Key Management Service
├── crypto/              # P1: JOSE operations (embedded)
└── infra/               # Infrastructure (config, telemetry, etc.)
```
