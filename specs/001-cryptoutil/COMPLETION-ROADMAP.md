# Completion Roadmap - specs/001-cryptoutil

**Date**: December 7, 2025
**Status**: 🎯 **READY TO EXECUTE**
**Total Effort**: 58-82 hours (ALL MANDATORY)
**Timeline**: 8-12 calendar days

---

## Executive Summary

After consolidating 22 overlapping iteration files into 4 essential documents and completing Speckit workflow steps 1-6, the cryptoutil project is **85% complete** with **42 tasks remaining (ALL MANDATORY)**.

**CRITICAL**: ALL phases and tasks are MANDATORY for Speckit completion. There are NO optional phases or tasks.

**✅ Speckit Compliance Restored**:

- Step 1: Constitution Review → No violations detected
- Step 2: Specification Verification → spec.md accurate
- Step 3: Clarifications → 6 ambiguities resolved
- Step 4: Technical Plan → 5 phases defined
- Step 5: Task Breakdown → 42 tasks with acceptance criteria
- Step 6: Coverage Analysis → 11 packages below 95%, strategies defined

**Next Step**: Begin implementation starting with **Phase 0: Slow Test Optimization** (foundation work).

---

## Implementation Sequence (5 Phases - ALL MANDATORY)

### Phase 0: Slow Test Optimization (Foundation) 🎯 **START HERE**

**Timeline**: Days 1-2 (8-10 hours)
**Priority**: CRITICAL - Enables fast feedback loop for all subsequent work

**Critical Packages (≥20s execution)**:

| Task | Package | Current | Target | Effort | Status |
|------|---------|---------|--------|--------|--------|
| P0.1 | clientauth | 168.0s | <30s | 2h | ⏳ TODO |
| P0.2 | jose/server | 93.9s | <20s | 1h | ⏳ TODO |
| P0.3 | kms/client | 74.3s | <20s | 2h | ⏳ TODO |
| P0.4 | jose | 66.7s | <15s | 1h | ⏳ TODO |
| P0.5 | kms/server/app | 28.0s | <10s | 1h | ⏳ TODO |

**Secondary Packages (10-20s execution)**:

| Task | Package | Current | Target | Effort | Status |
|------|---------|---------|--------|--------|--------|
| P0.6 | identity/authz | 19.2s | <10s | 1h | ⏳ TODO |
| P0.7 | identity/idp | 15.4s | <10s | 1h | ⏳ TODO |
| P0.8 | identity/test/unit | 17.9s | <10s | 30min | ⏳ TODO |
| P0.9 | identity/test/integration | 16.4s | <10s | 30min | ⏳ TODO |
| P0.10 | infra/realm | 13.8s | <10s | 30min | ⏳ TODO |
| P0.11 | kms/server/barrier | 12.6s | <10s | 30min | ⏳ TODO |

**Total Impact**: ~600s → <200s = **400s+ saved** (67% reduction)

**Why First**: Fast tests enable rapid iteration on all subsequent tasks. Without this, every code change takes 10+ minutes to validate.

**Files to Modify**: See TASKS.md P0.1-P0.11 for detailed file lists

---

### Phase 1: CI/CD Workflow Fixes 🎯 **NEXT PRIORITY**

**Timeline**: Days 3-4 (6-8 hours)
**Priority**: CRITICAL - Unblocks automated quality gates

**Current State**: 3/11 workflows passing (27%)
**Target State**: 11/11 workflows passing (100%)

**Priority Order (Highest to Lowest)**:

| Task | Workflow | Status | Effort | Priority |
|------|----------|--------|--------|----------|
| P1.1 | ci-coverage | ❌ Failing | 1h | 1-CRITICAL |
| P1.2 | ci-benchmark | ❌ Failing | 1h | 2-HIGH |
| P1.3 | ci-fuzz | ❌ Failing | 1h | 3-HIGH |
| P1.4 | ci-e2e | ❌ Failing | 1h | 4-HIGH |
| P1.5 | ci-dast | ❌ Failing | 1h | 5-MEDIUM |
| P1.6 | ci-race | ❌ Failing | 1h | 6-MEDIUM |
| P1.7 | ci-load | ❌ Failing | 30min | 7-MEDIUM |
| P1.8 | ci-sast | ❌ Failing | 30min | 8-LOW |

**Why Second**: Automated quality gates catch regressions early. Fast tests (from Phase 0) make workflow debugging tolerable.

**Dependencies**: Phase 0 complete (fast tests enable rapid workflow debugging)

---

### Phase 2: Deferred Features ⚠️ **PARTIAL PROGRESS**

**Timeline**: Days 5-6 (8-10 hours)
**Priority**: HIGH - Completes specification commitments

**Status**: 4/8 tasks complete (50%)

| Task | Feature | Status | Effort |
|------|---------|--------|--------|
| P2.1 | JOSE Authority E2E tests | ⏳ TODO | 4h |
| P2.2 | JOSE server OCSP support | ⏳ TODO | 3h |
| P2.3 | JOSE server Docker image | ⏳ TODO | 2h |
| P2.4 | EST serverkeygen | ⏳ TODO | 2h |
| P2.5 | CA E2E tests | ✅ Complete | 0h |
| P2.6 | CA server OCSP support | ✅ Complete | 0h |
| P2.7 | CA server Docker image | ✅ Complete | 0h |
| P2.8 | CA compose stack | ✅ Complete | 0h |

**CRITICAL**: EST serverkeygen (P2.4) is MANDATORY - PKCS#7 blocker must be resolved.

**Why Third**: Features are mostly complete (CA side done). JOSE Authority needs E2E tests, OCSP, Docker image.

**Dependencies**: Phase 0 complete (fast tests), Phase 1 complete (CI/CD validates E2E)

---

### Phase 3: Coverage Targets 📊 **QUALITY GATE**

**Timeline**: Days 7-9 (12-18 hours)
**Priority**: HIGH - Meets constitutional quality standards

**Target**: All packages ≥95% coverage

**Critical Gaps** (below 50%):

| Package | Current | Target | Gap | Effort |
|---------|---------|--------|-----|--------|
| ca/handler | 47.2% | 95% | -47.8% | 1-2h |
| auth/userauth | 42.6% | 95% | -52.4% | 1-2h |
| jose | 48.8% | 95% | -46.2% | 2-3h (must reach 95% before performance optimization) |

**Total Effort**: 4-7 hours (covers top 3 critical gaps)

**See ANALYSIS.md** for complete 11-package breakdown and strategies.

**Why Fourth**: Quality gates ensure production readiness. Must be done before declaring work complete.

**Dependencies**: Phase 0 complete (fast tests make coverage iteration tolerable)

---

### Phase 4: Advanced Testing 🔬 **MANDATORY**

**Timeline**: Days 10-11 (8-12 hours)
**Priority**: HIGH - Speckit requires advanced quality validation

| Task | Testing Type | Status | Effort |
|------|--------------|--------|--------|
| P4.1 | Mutation testing baseline | ⏳ TODO | 2h |
| P4.2 | Fuzz testing expansion | ⏳ TODO | 2-3h |
| P4.3 | Property-based testing | ⏳ TODO | 2-3h |
| P4.4 | Chaos engineering | ⏳ TODO | 2-4h |

**CRITICAL**: Advanced testing is MANDATORY for Speckit completion - not optional.

**Why Fourth**: Provides comprehensive quality confidence beyond basic coverage metrics.

---

### Phase 5: Demo Videos 🎥 **MANDATORY**

**Timeline**: Days 12-14 (16-24 hours)
**Priority**: MEDIUM - Speckit requires comprehensive documentation

| Task | Demo Topic | Status | Effort |
|------|------------|--------|--------|
| P5.1 | KMS quick start | ⏳ TODO | 2-3h |
| P5.2 | JOSE Authority usage | ⏳ TODO | 2-3h |
| P5.3 | Identity Server setup | ⏳ TODO | 3-4h |
| P5.4 | CA Server operations | ⏳ TODO | 3-4h |
| P5.5 | Multi-service integration | ⏳ TODO | 3-5h |
| P5.6 | Observability walkthrough | ⏳ TODO | 3-5h |

**CRITICAL**: Demo videos are MANDATORY for Speckit completion - not optional.

**Why Fifth**: Improves user experience and provides comprehensive onboarding documentation.

---

## Recommended Execution Plan (14 Days - ALL MANDATORY)

### **Days 1-2**: Slow Test Optimization (Phase 0) - 8-10 hours

```bash
# Optimize all 11 slow packages
go test ./... -v | grep -E "ok\s+internal" # Baseline current times
# Execute P0.1 → P0.2 → P0.3 → ... → P0.11
go test ./... -shuffle=on  # Validate <200s total
```

**Success Criteria**: Test suite runs in <200s (currently ~600s)

---

### **Days 3-4**: CI/CD Workflow Fixes (Phase 1) - 6-8 hours

```bash
# Fix workflows in priority order
# P1.1: ci-coverage (CRITICAL)
# P1.2: ci-benchmark (HIGH)
# P1.3: ci-fuzz (HIGH)
# P1.4: ci-e2e (HIGH)
# P1.5: ci-dast (MEDIUM)
# P1.6: ci-race (MEDIUM)
# P1.7: ci-load (MEDIUM)
# P1.8: ci-sast (LOW)

gh run list --limit 10  # Check current failures
go run ./cmd/workflow -workflows=coverage  # Test locally
```

**Success Criteria**: 11/11 workflows passing

---

### **Days 5-6**: Deferred Features (Phase 2) - 8-10 hours

```bash
# P2.1: JOSE E2E tests (4h)
go test ./internal/jose/server -run=TestE2E

# P2.2: JOSE OCSP support (3h)
# P2.3: JOSE Docker image (2h)
docker compose -f deployments/jose/compose.yml up -d
curl -k https://localhost:8080/ui/swagger/doc.json  # Validate

# P2.4: EST serverkeygen (2h - MANDATORY, resolve PKCS#7 blocker)
```

**Success Criteria**: All 8 Phase 2 tasks complete, JOSE Authority feature-complete

---

### **Days 7-9**: Coverage Targets (Phase 3) - 12-18 hours

```bash
# Focus on critical gaps first (ca/handler, auth/userauth, jose)
go test ./internal/ca/handler -cover  # Target 95%+
go test ./internal/identity/auth/userauth -cover  # Target 95%+
go test ./internal/jose -cover  # Target 95%+

# Then address remaining packages
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

**Success Criteria**: All packages ≥95% coverage

---

### **Days 10-11**: Advanced Testing (Phase 4) - 8-12 hours

```bash
# P4.1: Mutation testing (2h)
gremlins unleash --tags=!integration

# P4.2: Fuzz testing expansion (2-3h)
go test -fuzz=FuzzHKDFAllVariants -fuzztime=15s ./pkg/crypto/kdf

# P4.3: Property-based testing (2-3h)
# Implement gopter tests for crypto operations

# P4.4: Chaos engineering (2-4h)
# Implement failure injection tests
```

**Success Criteria**: Mutation score ≥80%, fuzz tests pass, property tests created

---

### **Days 12-14**: Demo Videos (Phase 5) - 16-24 hours

```bash
# P5.1: KMS quick start (2-3h)
# P5.2: JOSE Authority usage (2-3h)
# P5.3: Identity Server setup (3-4h)
# P5.4: CA Server operations (3-4h)
# P5.5: Multi-service integration (3-5h)
# P5.6: Observability walkthrough (3-5h)

# Record, edit, and publish all 6 demo videos
```

**Success Criteria**: All 6 demo videos complete and published

---

## How to Track Progress

### Primary Documents

1. **PROJECT-STATUS.md** - Single source of truth for task status
   - Update task status: ⏳ TODO → 🔄 IN PROGRESS → ✅ COMPLETE
   - Update completion percentages
   - Track blockers and decisions

2. **IMPLEMENTATION-GUIDE.md** - Day-by-day execution plan
   - Check off daily tasks as completed
   - Adjust timeline if needed

3. **TASKS.md** - Detailed task acceptance criteria
   - Reference for "what does done look like"
   - Lists files to modify per task

4. **ANALYSIS.md** - Coverage gap strategies
   - Reference for Phase 3 coverage work
   - Package-by-package improvement plans

### Git Workflow

```bash
# Daily commits after each task
git add <files>
git commit -m "feat(phase0): optimize clientauth package tests (P0.1)

- Reduced execution time from 168s to 25s
- Used in-memory SQLite instead of PostgreSQL testcontainers
- Maintained 95%+ coverage and mutation score

Closes: P0.1"

# Push at end of each day
git push origin main
```

### Validation Commands

```bash
# After each task
golangci-lint run --fix  # Mandatory - fix ALL linting errors
go test ./... -shuffle=on  # Verify tests pass
go test ./... -cover  # Check coverage impact

# After each phase
gh run list --limit 5  # Check CI/CD status
go test ./... -v | grep -E "(PASS|FAIL|ok|FAIL)\s+internal" # Review test results
git log --oneline -10  # Review commit history
```

---

## Success Metrics (ALL PHASES MANDATORY)

### Definition of Done for specs/001-cryptoutil

**CRITICAL**: ALL metrics must reach target. No phases are optional.

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| Test Suite Speed | ~600s (11 pkgs) | <200s | ⏳ Phase 0 |
| CI/CD Pass Rate | 27% (3/11) | 100% (11/11) | ⏳ Phase 1 |
| JOSE E2E Tests | ❌ None | ✅ All 10 endpoints | ⏳ Phase 2 |
| JOSE OCSP | ❌ Missing | ✅ Implemented | ⏳ Phase 2 |
| JOSE Docker | ❌ Missing | ✅ Published | ⏳ Phase 2 |
| EST serverkeygen | ❌ Missing | ✅ Implemented | ⏳ Phase 2 |
| Package Coverage | 11 below 95% | All ≥95% | ⏳ Phase 3 |
| Mutation Score | Unknown | ≥80% all packages | ⏳ Phase 4 |
| Fuzz Testing | Limited | Comprehensive | ⏳ Phase 4 |
| Property Testing | None | Core packages | ⏳ Phase 4 |
| Demo Videos | None | All 6 videos | ⏳ Phase 5 |

**Project Completion**: When ALL metrics reach target (Phases 0-5 complete).

---

## What Success Looks Like

### After Phase 0 (Days 1-2)

- ✅ Test suite runs in <200s (was ~600s)
- ✅ Can iterate on code changes rapidly (10min → 3min feedback)
- ✅ Developer experience dramatically improved
- ✅ All 11 slow packages optimized

### After Phase 1 (Days 3-4)

- ✅ All 11 CI/CD workflows passing (green checkmarks)
- ✅ Automated quality gates catch regressions
- ✅ Can merge PRs with confidence

### After Phase 2 (Days 5-6)

- ✅ JOSE Authority feature-complete (E2E tests, OCSP, Docker)
- ✅ All 4 products (JOSE, Identity, KMS, CA) have Docker images
- ✅ All specification commitments fulfilled
- ✅ EST serverkeygen implemented (PKCS#7 blocker resolved)

### After Phase 3 (Days 7-9)

- ✅ All packages ≥95% coverage (constitutional requirement met)
- ✅ Production-ready code quality

### After Phase 4 (Days 10-11)

- ✅ Mutation score ≥80% on all packages
- ✅ Comprehensive fuzz testing coverage
- ✅ Property-based testing on core packages
- ✅ Advanced quality validation complete

### After Phase 5 (Days 12-14)

- ✅ All 6 demo videos complete
- ✅ Comprehensive onboarding documentation
- ✅ User experience optimized

### After ALL Mandatory Work (Days 1-14)

- ✅ **specs/001-cryptoutil is 100% COMPLETE** 🎉
- ✅ cryptoutil project is production-ready
- ✅ All constitutional requirements met
- ✅ All specification commitments delivered
- ✅ Speckit workflow fully complete

---

## FAQs

### Q: Why is slow test optimization Phase 0 (first)?

**A**: Fast tests enable rapid iteration on all subsequent work. Without this, every code change takes 10+ minutes to validate, making development painful. Phase 0 reduces this to <3 minutes, a 67% speedup.

### Q: Can I skip Phase 0 and go straight to features?

**A**: **NO**. Slow tests will make all subsequent work take 3-4x longer. Phase 0 is foundation work that pays dividends throughout Phases 1-5.

### Q: Are Phases 4 and 5 really mandatory?

**A**: **YES**. Speckit requires:

- Phase 4 (Advanced Testing): Mutation testing, fuzz testing, property-based testing, chaos engineering
- Phase 5 (Demo Videos): Comprehensive user onboarding and documentation
Without these, Speckit workflow is incomplete.

### Q: When can I declare specs/001-cryptoutil complete?

**A**: **ONLY after completing ALL 5 phases (58-82 hours)**. ALL phases are MANDATORY for Speckit completion. No shortcuts.

### Q: What if I find issues during implementation?

**A**: Update PROJECT-STATUS.md immediately. Add new tasks if needed. Adjust timeline. Document blockers in PROJECT-STATUS.md "Known Limitations" section.

### Q: How do I know I haven't missed anything?

**A**: Cross-reference against TASKS.md (42 tasks), spec.md (requirements), and ANALYSIS.md (coverage gaps). If all tasks are ✅ COMPLETE and all spec requirements are ✅ DONE, you're finished.

---

## Next Immediate Step

**🎯 START HERE**: Execute **Phase 0, Task P0.1** (optimize clientauth package)

```bash
# 1. Baseline current state
cd c:\Dev\Projects\cryptoutil
go test ./internal/identity/authz/clientauth -v  # Currently 168s

# 2. Open task details
code specs/001-cryptoutil/TASKS.md  # Read P0.1 acceptance criteria

# 3. Implement optimization
# See TASKS.md P0.1 for file modifications and strategy

# 4. Validate
go test ./internal/identity/authz/clientauth -v  # Target <30s
go test ./internal/identity/authz/clientauth -cover  # Maintain 95%+

# 5. Update status
# Edit PROJECT-STATUS.md: P0.1 status ⏳ TODO → ✅ COMPLETE

# 6. Commit
git add <modified files>
git commit -m "perf(clientauth): optimize test execution time (P0.1)"
git push
```

**Estimated Time**: 1-1.5 hours
**Impact**: 168s → <30s = **138s saved** (82% reduction on this package)

---

## Document References

| Document | Purpose | When to Use |
|----------|---------|-------------|
| **COMPLETION-ROADMAP.md** (this file) | High-level execution plan | Starting work, daily planning |
| **PROJECT-STATUS.md** | Task status tracking | Updating progress, checking completion |
| **IMPLEMENTATION-GUIDE.md** | Day-by-day breakdown | Daily execution, timeline management |
| **TASKS.md** | Detailed acceptance criteria | Implementing specific tasks |
| **ANALYSIS.md** | Coverage gap strategies | Phase 3 coverage work |
| **PLAN.md** | Technical implementation plan | Understanding architecture decisions |
| **spec.md** | Requirements specification | Validating feature completeness |

---

## Conclusion

**Current State**: ✅ Speckit workflow steps 1-6 complete, project 85% complete
**Remaining Work**: 36 tasks (27 required, 9 optional) across 5 phases
**Timeline**: 3-5 calendar days (16-24 hours required work)
**Next Step**: Begin Phase 0, Task P0.1 (optimize clientauth package)

**You are ready to execute.** All planning is complete. All ambiguities are resolved. All tasks are defined with clear acceptance criteria. Start with Phase 0 and work sequentially through Phases 1-3 to reach 100% completion.

---

*Completion Roadmap Version: 1.0.0*
*Created: December 7, 2025*
*Author: GitHub Copilot (Agent)*
*Status: 🎯 READY TO EXECUTE*
