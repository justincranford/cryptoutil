# Evidence-Based Task Completion - Complete Specifications

**Referenced by**: `.github/instructions/06-01.evidence-based.instructions.md`

## Core Principle - CRITICAL

**NEVER mark tasks complete without objective evidence**

**Why**: Prevents premature completion, ensures quality gates, enables validation, catches regressions

## Mandatory Evidence Checklist

**Build**: `go build ./...` (no errors)
**Linting**: `golangci-lint run` (no warnings)
**TODOs**: 0 new TODOs vs baseline
**Coverage**: ≥95% production, ≥98% infrastructure/utility
**Tests**: `go test ./...` (0 failures, no skips without tracking)
**Mutation**: ≥80% early phases, ≥98% infrastructure/utility

## Progressive Validation (After Every Task)

1. TODO scan (0 new)
2. Test run (0 failures)
3. Coverage check (≥95% prod, ≥98% infra)
4. Mutation testing (≥80% early, ≥98% infra)
5. Integration test (E2E functional)
6. Documentation update (DETAILED.md Section 2)

## Quality Gate - MANDATORY

**Task NOT complete until ALL pass**: Build, linting, TODOs, tests, coverage, mutation, integration, documentation

**NO EXCEPTIONS**: If any check fails, task is incomplete

## Single Source of Truth - CRITICAL

### implement/DETAILED.md

**Section 1: Task Checklist** - Status symbols (❌/⚠️/✅), links to evidence
**Section 2: Append-Only Timeline** - Date, task, metrics, findings, violations, next steps

### implement/EXECUTIVE.md

**Sections**: Stakeholder overview, customer demos, risk tracking, post mortem, last updated
**Update**: After phase completion, major milestone, or stakeholder request

---

## Phase Dependencies - Strict Sequence

**Phase 1** (Foundation): Domain, schema, CRUD - ≥95% coverage, ≥80% mutation, 0 TODOs
**Phase 2** (Core Logic): Business logic, APIs, auth - E2E works, 0 CRITICAL TODOs (depends on Phase 1)
**Phase 3** (Advanced): MFA, WebAuthn, federation - ≥98% coverage/mutation (depends on Phase 1+2)

**Rationale**: Foundation bugs cascade, refactoring breaks later phases, prevents tech debt

---

## Post-Mortem Enforcement - MANDATORY

**Option 1** (<30 min): Immediate fix, update DETAILED.md
**Option 2** (>30 min): Create `specs/*/##.##-GAP_NAME.md`, schedule future

**NEVER**: Leave gaps undocumented

**Post-Mortem Template**: `docs/P0.X-INCIDENT_NAME.md` - Summary, root cause, timeline, impact, lessons, action items

## Common Violations

**NEVER**: Mark complete without validation, skip post-mortem, implement Phase 3 before Phase 2
**ALWAYS**: Run all checks, create post-mortems, respect phase dependencies
