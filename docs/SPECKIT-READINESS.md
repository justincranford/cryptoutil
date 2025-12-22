# Speckit Readiness - Blocking Questions for /speckit.plan

**Date**: 2025-12-21
**Purpose**: Identify blockers preventing restart with `/speckit.plan` command
**Context**: Recent instruction file reorganization and condensing completed

---

## Critical Blockers

### Blocker 1: Missing constitution.md

**Question**: Where is `.specify/memory/constitution.md`?
**Impact**: `/speckit.constitution` is step 1 of Spec Kit methodology - MUST exist before `/speckit.plan`
**Current State**:

- `.specify/memory/` directory does NOT exist
- `specs/002-cryptoutil/constitution.md` does NOT exist
- `specs/002-cryptoutil/spec.md` exists but references constitution.md
**Action Required**:
- Create `.specify/memory/constitution.md` OR
- Move existing constitutional requirements into new constitution.md OR
- Confirm if spec.md is being used AS constitution.md (non-standard)

---

### Blocker 2: Clarify vs CLARIFY-QUIZME status

**Question**: Is `specs/002-cryptoutil/clarify.md` complete or are there pending CLARIFY-QUIZME questions?
**Impact**: Unresolved questions block accurate planning
**Current State**:

- `specs/002-cryptoutil/clarify.md.old` exists
- `specs/002-cryptoutil/SPECKIT-CLARIFY-QUIZME-TEMPLATE.md` exists (template, not actual questions)
- Unknown if active CLARIFY-QUIZME file exists with pending questions
**Action Required**:
- Confirm all CLARIFY-QUIZME questions resolved OR
- Identify pending questions requiring user answers

---

### Blocker 3: Out-of-date plan/tasks/analyze files

**Question**: Why are plan, tasks, and analyze marked "probably-out-of-date"?
**Impact**: Cannot use `/speckit.plan` if existing plan.md is stale
**Current State**:

- `specs/002-cryptoutil/plan-probably-out-of-date.md` exists
- `specs/002-cryptoutil/tasks-probably-out-of-date.md` exists
- `specs/002-cryptoutil/analyze-probably-out-of-date.md` exists
**Action Required**:
- Delete out-of-date files OR
- Rename to active files if still valid OR
- Confirm these are archival and create fresh plan.md

---

### Blocker 4: Instruction files vs spec.md alignment

**Question**: Are recent instruction file changes (HTTPS ports, PKI, federation) reflected in spec.md?
**Impact**: Spec-code mismatch causes implementation confusion
**Current State**:

- spec.md last modified unknown
- Instruction files refactored on 2025-12-21 (commits c1f7a587, 50cd7e62, 0d3530af, d805b998)
- spec.md references "01-01.architecture.instructions.md" (should be 02-01.architecture.instructions.md)
- spec.md may have outdated HTTPS endpoint patterns
**Action Required**:
- Update spec.md to reference correct instruction file numbers
- Update spec.md HTTPS endpoint descriptions to match 02-03.https-ports.instructions.md
- Update spec.md to reflect condensed speckit.instructions.md patterns

---

## Medium Priority Blockers

### Blocker 5: DETAILED.md vs EXECUTIVE.md completeness

**Question**: Do implement/DETAILED.md and implement/EXECUTIVE.md reflect current implementation state?
**Impact**: Stale implementation tracking blocks accurate planning
**Current State**:

- `specs/002-cryptoutil/implement/DETAILED.md` exists
- `specs/002-cryptoutil/implement/EXECUTIVE.md` exists
- Unknown if they reflect current codebase state (last update date?)
**Action Required**:
- Review DETAILED.md Section 2 timeline for recent work
- Review EXECUTIVE.md for current phase, progress %, coverage, risks
- Update both files OR confirm they're current

---

### Blocker 6: Product Suite vs Service Template extraction status

**Question**: Has service template been extracted per Phase 6-10 requirements?
**Impact**: Phase dependencies require service template before new service implementation
**Current State**:

- spec.md mentions service template extraction
- Unknown if template exists in `internal/template/` or similar
- Unknown if learn-ps uses template (Phase 7 requirement)
**Action Required**:
- Confirm service template exists and location
- Confirm learn-ps implemented with template
- Update DETAILED.md/EXECUTIVE.md with template status

---

### Blocker 7: Quality gates current status

**Question**: What are current coverage, mutation, and timing metrics?
**Impact**: Quality gates determine phase completion and next steps
**Current State from spec.md**:

- Coverage target: ≥95% production, ≥98% infrastructure
- Mutation target: ≥85% Phase 4, ≥98% Phase 5+
- Timing target: <15s per unit test package, <120s total unit, <45s per E2E package, <240s total E2E
- Unknown: Current actual metrics
**Action Required**:
- Run `go test ./... -coverprofile=coverage.out`
- Run `gremlins unleash --tags=!integration`
- Document current metrics in EXECUTIVE.md

---

## Low Priority Questions (Non-Blocking)

### Question 8: Federation implementation status

**Question**: Which services have federation configuration implemented?
**Impact**: Determines readiness for cross-service E2E testing
**Current State**: Unknown
**Action**: Review implement/DETAILED.md for federation tasks completion

### Question 9: Docker Compose E2E test coverage

**Question**: Do all 9 services have working Docker Compose E2E tests?
**Impact**: Deployment validation completeness
**Current State**: Unknown
**Action**: Check `deployments/compose/` for service-specific compose files

### Question 10: PKI CA architecture implementation

**Question**: Which CA architecture pattern is implemented (Offline Root → Online Root → Issuing CA)?
**Impact**: Certificate chain configuration for TLS endpoints
**Current State**: Unknown
**Action**: Review pki-ca implementation or DETAILED.md

---

## Recommended Action Sequence

1. **CRITICAL**: Locate or create `.specify/memory/constitution.md`
2. **CRITICAL**: Resolve pending CLARIFY-QUIZME questions or confirm none exist
3. **CRITICAL**: Delete or rename "probably-out-of-date" files
4. **HIGH**: Update spec.md to reference correct instruction files (02-01, 02-03, 02-10, etc.)
5. **HIGH**: Review and update implement/DETAILED.md and implement/EXECUTIVE.md
6. **MEDIUM**: Run quality metrics (coverage, mutation, timing)
7. **MEDIUM**: Verify service template extraction status
8. **LOW**: Document federation, Docker Compose, PKI implementation status

---

## Validation Checklist Before /speckit.plan

- [ ] `.specify/memory/constitution.md` exists with complete sections
- [ ] No TBD/TODO/FIXME in constitution.md (`grep -i "TBD\|TODO\|FIXME" .specify/memory/constitution.md`)
- [ ] All CLARIFY-QUIZME questions resolved (no pending `clarify-QUIZME*.md` files)
- [ ] spec.md references correct instruction file numbers
- [ ] implement/DETAILED.md Section 2 timeline reflects recent work
- [ ] implement/EXECUTIVE.md has current phase, progress, coverage, risks
- [ ] Quality metrics documented (coverage, mutation, timing)
- [ ] Phase dependencies satisfied (template extracted if Phase 7+)

---

## Notes

- Instruction files recently reorganized (2025-12-21):
  - speckit.instructions.md condensed from 517 to 105 lines
  - bind-address.instructions.md renamed to https-ports.instructions.md
  - HTTPS endpoints and request paths moved from architecture to https-ports
  - Verbose examples simplified across multiple files
- Pre-push hooks validated and working correctly
