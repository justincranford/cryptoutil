# Grooming Session 7: Final Clarifications (Optional)

## Purpose

Session 7 is an optional final clarification session. Based on the comprehensive answers from Sessions 5-6, most major decisions have been captured. This session covers any remaining edge cases or implementation details.

**Date**: November 30, 2025
**Status**: OPTIONAL - Skip if no major clarifications needed

---

## Summary of Decisions (Sessions 1-6)

All major architectural and implementation decisions have been captured:

### Architecture

- **Directory Structure**: `internal/infra/` + `internal/product/`
- **Product Priority**: KMS → Identity → JOSE Authority → Certificate Authority
- **Cross-Product**: Products only import infra, never each other
- **E2E Location**: `internal/product/e2e/` for cross-product tests

### Migration Strategy

- **First Packages**: magic, apperr, crypto, telemetry (batch move to infra)
- **KMS Migration**: Bottom-up (barrier → server → client)
- **Identity Migration**: Working packages first, broken packages later
- **Verification**: build + test + lint after each step

### Implementation Details

- **Import Aliases**: Keep existing `cryptoutil*` convention
- **Circular Deps**: Address as they appear
- **Identity Duplicates**: Keep both versions, consolidate later
- **JOSE Authority Source**: Extract from `internal/identity/issuer/`

### Demo Requirements

- **Per-Product**: Compose + health checks + sample API calls
- **Cross-Product Config**: YAML-based product relationships
- **Success Criteria**: All endpoints work, auth flows complete, offline capable

---

## Optional Questions (Q51-55)

Only answer if there's ambiguity or you want to clarify something.

### Q51. Migration Batch Size

How many files/packages should be moved per commit?

- [ ] A. One package per commit (most granular)
- [ ] B. Related packages together (e.g., all of `internal/common/`)
- [ ] C. Logical groups (infra batch, then KMS, then Identity)
- [ ] D. Whatever keeps tests passing
- [ ] E. No preference - your judgment

**Notes**: (Leave blank if default of D/E is fine)

---

### Q52. Test Failures During Migration

If tests fail during migration due to import issues:

- [ ] A. Fix immediately before continuing
- [ ] B. Document and fix in subsequent commit
- [ ] C. Skip failing tests temporarily (add TODO)
- [ ] D. Never acceptable - rollback and try different approach
- [ ] E. Depends on severity

**Notes**: (Leave blank if default of A is fine)

---

### Q53. Coverage Reporting Location

Where should coverage reports be generated?

- [ ] A. `./test-output/` (current location)
- [ ] B. Per-product: `internal/product/kms/coverage/`, etc.
- [ ] C. `./coverage/` at root
- [ ] D. No preference - current location fine
- [ ] E. Different locations for different report types

**Notes**: (Leave blank if current location is fine)

---

### Q54. Demo Data Format

Demo seed data format:

- [ ] A. SQL files (direct database seeding)
- [ ] B. JSON fixtures loaded by application
- [ ] C. Go code that creates demo data
- [ ] D. YAML configuration files
- [ ] E. Mix - different formats for different data types

**Notes**: (Leave blank if Go code default is fine)

---

### Q55. Breaking Change Threshold

What change would make you reconsider the refactoring approach?

- [ ] A. Nothing - committed to the plan
- [ ] B. If it takes more than 2 weeks
- [ ] C. If test coverage drops significantly
- [ ] D. If demos can't run
- [ ] E. Major unforeseen technical blocker

**Notes**: (Leave blank if A is fine per Session 5 Q18)

---

## Proceed to Implementation

If no clarifications needed above, grooming is complete.

**Next Steps**:

1. Update `TASK-LIST.md` with refined tasks based on all grooming sessions
2. Update `README.md` with architectural decisions
3. Begin implementation starting with Phase 1 (KMS Demo Verification)

---

**Status**: OPTIONAL - SKIP TO IMPLEMENTATION IF NO CLARIFICATIONS NEEDED
