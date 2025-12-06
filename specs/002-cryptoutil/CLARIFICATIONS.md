# Iteration 2 Clarifications Checklist

## Purpose

This document identifies ambiguities, incomplete items, and clarifications needed before proceeding to Iteration 3.

**Created**: December 6, 2025
**Iteration**: 2
**Status**: ⚠️ Requires User Input

---

## 1. Test Runtime Optimization Strategy

### Clarification Needed

**Question**: How should test runtime optimization be implemented for selective test execution during local development vs CI?

**Current State**:
- DELETE-ME-LATER-SLOW-TEST-PACKAGES.md identifies 10+ slow packages
- Current approach: Run all tests sequentially with `-p=1` for reliability
- CI execution time: Long (10+ minutes for some packages)
- Local feedback loop: Slow for iterative development

**User Provided Clarification** ✅:
> Do selective (mix of deterministic and/or randomized) test execution to speed up slowest 10 Go packages during local unit/integration tests, for faster feedback. CI probably needs to do more or all tests, versus local execution of tests.

**Implementation Plan**:
- Create test tags for fast/slow/integration/e2e categories
- Add `make test-fast` for local development (subset of tests)
- Keep `make test-all` for CI (comprehensive)
- Document test selection strategy in testing instructions
- Identify slowest 10 packages from coverage data
- Implement randomized subset execution for local dev

**Priority**: MEDIUM
**Assigned To**: Iteration 3 - Task PERF-1

---

## 2. CMS and PKCS#7 Library Selection

### Clarification Needed

**Question**: Which Go library should be used for CMS (Cryptographic Message Syntax) and PKCS#7 operations?

**Current State**:
- No CMS/PKCS#7 implementation in project
- Needed for S/MIME, code signing, secure messaging
- Multiple options: go.mozilla.org/pkcs7, github.com/fullsailor/pkcs7

**User Provided Clarification** ✅:
> For CMS and PKCS#7, use go.mozilla.org/pkcs7 (stable, widely used)

**Implementation Plan**:
- Add `go.mozilla.org/pkcs7` dependency to go.mod
- Create wrapper package in `internal/common/crypto/cms`
- Implement CMS sign/verify/encrypt/decrypt operations
- Add FIPS compliance validation for algorithms
- Document in architecture docs

**Priority**: LOW (Future iteration)
**Assigned To**: Iteration 4 or later

---

## 3. Iteration 2 Incomplete Items

### 3.1 JOSE Docker Integration (JOSE-13)

**Status**: ❌ Not Started
**Blocker**: None
**Effort**: 2 hours

**What's Needed**:
- Create `deployments/jose/Dockerfile.jose`
- Create `deployments/jose/compose.jose.yml`
- Add health check endpoints
- Add to main `deployments/compose/compose.yml`

**Clarification Question**: Should JOSE Authority:
- [ ] Share PostgreSQL with Identity/KMS?
- [ ] Have separate PostgreSQL instance?
- [ ] Support in-memory SQLite for standalone?

**Recommended**: Support all three modes via configuration

---

### 3.2 CA OCSP Handler (CA-11)

**Status**: ❌ Not Started
**Blocker**: None
**Effort**: 6 hours

**What's Needed**:
- Implement RFC 6960 OCSP responder
- OCSP request parsing and validation
- OCSP response signing
- OCSP response caching strategy
- Integration with certificate storage

**Clarification Question**: OCSP responder configuration:
- [ ] Dedicated OCSP signing certificate?
- [ ] Use CA certificate for OCSP signing?
- [ ] OCSP response validity period (minutes/hours)?
- [ ] OCSP response caching duration?

**Recommended**: Dedicated OCSP signing cert, 24h validity, 1h cache

---

### 3.3 EST Protocol Endpoints (CA-14 through CA-17)

**Status**: ❌ Not Started (4 endpoints)
**Blocker**: None
**Effort**: 12 hours total

**What's Needed**:
- CA-14: `/est/cacerts` - Get CA certificate chain
- CA-15: `/est/simpleenroll` - Simple enrollment
- CA-16: `/est/simplereenroll` - Certificate renewal
- CA-17: `/est/serverkeygen` - Server-side key generation

**Clarification Questions**:
1. EST authentication method:
   - [ ] HTTP Basic Auth only?
   - [ ] Client certificate (mTLS)?
   - [ ] Both?

2. Server-side key generation (CA-17):
   - [ ] Store private keys server-side?
   - [ ] Return private keys in response?
   - [ ] Use HSM/TPM for key storage?

**Recommended**: Support both auth methods, return private keys encrypted (don't store server-side)

---

### 3.4 CA Integration Tests (CA-19)

**Status**: ⚠️ Partial (1.2% coverage)
**Blocker**: Missing endpoints implementation
**Effort**: 8 hours

**What's Needed**:
- Expand `internal/ca/api/handler/handler_test.go`
- Add E2E tests for all endpoints
- Test certificate lifecycle (issue → retrieve → revoke → check)
- Test CRL generation and retrieval
- Test OCSP responder
- Test EST enrollment flows
- Error scenario testing

**Clarification Question**: Test data strategy:
- [ ] Generate certificates on-the-fly in tests?
- [ ] Use pre-generated test fixtures?
- [ ] Both (fixtures for speed, generation for coverage)?

**Recommended**: Both approaches - fixtures for common cases, generation for edge cases

---

## 4. File Organization Issues

### 4.1 Misplaced Files Between specs/001 and specs/002

**Analysis**:
- ✅ All files in specs/001-cryptoutil belong to Iteration 1
- ✅ All files in specs/002-cryptoutil belong to Iteration 2
- ❌ specs/002 missing: ANALYSIS.md, CLARIFICATIONS.md, CHECKLIST-ITERATION-2.md

**Action Required**:
- Create ANALYSIS.md for specs/002 (coverage analysis of iteration 2)
- This file serves as CLARIFICATIONS.md for specs/002
- Create CHECKLIST-ITERATION-2.md when iteration 2 fully complete

---

### 4.2 DELETE-ME Files Organization

**Current State**:
- `docs/DELETE-ME-LATER-CROSS-REF-SPECKIT-COPILOT-TEMPLATE.md` (224 lines)
- `docs/DELETE-ME-LATER-SLOW-TEST-PACKAGES.md` (coverage data)
- Missing: `test-output/PRE-COMMIT-SPEED-COVERAGE.md`
- Missing: `test-output/PRE-COMMIT-SPEED-COVERAGE-REMAINING.md`

**Clarification Question**: Should DELETE-ME content be:
- [ ] Integrated into specs/003-cryptoutil during planning?
- [ ] Integrated into constitution/instructions?
- [ ] Both?

**Recommended**: Extract actionable items into specs/003 tasks, apply lessons to constitution/instructions, then delete

---

## 5. Coverage Target Increments

### Clarification Needed

**Question**: After incrementing coverage targets from 90/95/100 to 95/100/100, how aggressive should coverage improvement be?

**Current Targets** (from constitution):
- 90%+ production coverage
- 95%+ infrastructure (cicd)
- 100% utility code

**Proposed New Targets**:
- 95%+ production coverage (+5%)
- 100% infrastructure (cicd) (+5%)
- 100% utility code (unchanged)

**Clarification Question**:
1. Should we increment again in future iterations?
   - [ ] Yes, eventually reach 100% everywhere
   - [ ] No, 95/100/100 is sustainable target

2. Acceptable runtime regression for coverage increase:
   - [ ] No regression allowed
   - [ ] Up to 10% slower acceptable
   - [ ] Up to 20% slower acceptable

**Recommended**: 95/100/100 is sustainable, up to 10% regression acceptable

---

## 6. Gremlins Mutation Testing Integration

### Clarification Needed

**Question**: How should gremlins mutation testing be integrated into development workflow?

**Options**:
1. **CI Integration**: Run gremlins in dedicated workflow
2. **Local Development**: Manual execution only
3. **Pre-commit Hook**: Block commits with poor mutation scores
4. **Periodic Reports**: Weekly/monthly baseline updates

**Clarification Question**: Preferred approach?
- [ ] Option 1: CI workflow (adds time)
- [ ] Option 2: Manual only (developers forget)
- [ ] Option 3: Pre-commit (strict, may slow commits)
- [ ] Option 4: Periodic reports (less enforcement)
- [ ] Combination of options

**Recommended**: Start with Option 2 (manual) + Option 4 (periodic), graduate to Option 1 (CI) once baseline established

---

## 7. Property-Based Testing Framework

### Clarification Needed

**Question**: Which property-based testing framework should be used for Go?

**Options**:
1. **gopter**: Pure Go, good documentation
2. **go-fuzz**: Focus on fuzzing, less property-based
3. **rapid**: Modern, good API
4. **testing/quick**: Standard library (limited)

**Clarification Question**: Framework selection criteria:
- [ ] Prefer standard library?
- [ ] Prefer most popular (gopter)?
- [ ] Prefer most modern (rapid)?

**Recommended**: Use `gopter` for property-based tests, keep `go-fuzz` for fuzzing (different use cases)

---

## 8. Workflow Improvements Strategy

### Clarification Needed

**Question**: How should workflow improvements be prioritized and implemented?

**Current Issues** (from baseline analysis needed):
- Unknown: Slow workflows
- Unknown: Failing workflows
- Unknown: Missing best practices

**Clarification Question**: Implementation strategy:
- [ ] Fix all workflows in one sweep (high risk)
- [ ] Fix one workflow at a time (slow)
- [ ] Fix by category (linting, testing, security)
- [ ] Fix highest impact first (requires baseline)

**Recommended**: Create baseline report first, then fix highest impact workflows iteratively

---

## 9. Iteration 3 Scope

### Clarification Needed

**Question**: Should Iteration 3 focus on:

**Option A: Complete Iteration 2 Items**
- Finish all 14 remaining iteration 2 tasks
- JOSE Docker, CA OCSP/EST, unified suite
- Estimated: ~40 hours

**Option B: Iteration 2 Completion + New Features**
- Complete iteration 2
- Add new capabilities (gremlins, property tests, etc.)
- Estimated: ~80+ hours

**Option C: Hybrid Approach**
- Complete HIGH priority iteration 2 items only
- Start iteration 3 new features in parallel
- Estimated: ~60 hours

**User Clarification Needed**: Preferred approach?

**Recommendation Pending User Input**

---

## 10. Iteration 2 Completion vs Iteration 3 Start

### Analysis

**From DELETE-ME-LATER-CROSS-REF-SPECKIT-COPILOT-TEMPLATE.md**:
> **Recommendation: Complete Iteration 1 FIRST**
> 
> **Rationale**:
> 1. CHECKLIST-ITERATION-1.md claims 44/44 tasks (100%) but has known gaps
> 2. Tests require `-p=1` for reliability (test parallelism issues unfixed)
> 3. client_secret_jwt (70%) and private_key_jwt (50%) incomplete
> 4. ANALYSIS.md identifies 7 gaps not addressed

**Current Iteration 2 State**:
- Tasks: 39/47 complete (83%)
- Known gaps: 14 tasks remaining
- Test coverage: Below targets (48-56% JOSE, 1% CA)
- Docker integration: Incomplete

**Clarification Question**: Should we:
- [ ] Complete ALL iteration 2 tasks before starting iteration 3?
- [ ] Start iteration 3 with iteration 2 at 83%?
- [ ] Define "good enough" threshold (e.g., 90%)?

**Recommended**: Complete HIGH priority iteration 2 tasks (OCSP, tests, coverage), defer LOW priority (EST reenroll, serverkeygen) to iteration 4

---

## Summary of Clarifications Needed

| ID | Topic | Priority | User Input Needed |
|----|-------|----------|-------------------|
| 1 | Test Runtime Optimization | MEDIUM | ✅ Provided |
| 2 | CMS/PKCS#7 Library | LOW | ✅ Provided |
| 3.1 | JOSE Docker Config | MEDIUM | Database strategy |
| 3.2 | CA OCSP Config | HIGH | Certificate and caching strategy |
| 3.3 | EST Authentication | MEDIUM | Auth method preference |
| 3.4 | CA Test Data Strategy | HIGH | Fixtures vs generation |
| 5 | Coverage Target Philosophy | MEDIUM | Sustainable targets |
| 6 | Gremlins Integration | LOW | Workflow integration approach |
| 7 | Property Test Framework | LOW | Framework selection |
| 8 | Workflow Improvement Strategy | HIGH | Prioritization approach |
| 9 | Iteration 3 Scope | HIGH | ⚠️ **CRITICAL** - Scope decision |
| 10 | Iteration 2 Completion Gate | HIGH | ⚠️ **CRITICAL** - Completion threshold |

---

## Recommended Next Steps

### Immediate (Can Proceed Without Clarification)

1. ✅ Create specs/002-cryptoutil/PROGRESS.md - DONE
2. ✅ Create specs/002-cryptoutil/EXECUTIVE-SUMMARY.md - DONE
3. ✅ Create specs/002-cryptoutil/CLARIFICATIONS.md - THIS FILE
4. Create specs/000-cryptoutil-template/
5. Update constitution with lessons learned
6. Update copilot instructions

### Awaiting User Input (CRITICAL)

**Question 9**: Iteration 3 scope decision
**Question 10**: Iteration 2 completion threshold

### After User Input

7. Create specs/003-cryptoutil/ based on scope decision
8. Integrate DELETE-ME file content
9. Increment coverage targets
10. Begin gremlins baseline (if in scope)
11. Begin workflow analysis (if in scope)

---

*Clarifications Version: 2.0.0*
*Created: December 6, 2025*
*Status: Awaiting User Input on Critical Items (9, 10)*
