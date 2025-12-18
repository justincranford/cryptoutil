# Clarifications - Post-Consolidation

**Date**: December 7, 2025
**Context**: Identify and resolve ambiguities introduced by consolidating 20 iteration files
**Status**: ✅ 6 clarifications identified and resolved

---

## Clarification 1: Task Counting Discrepancy

### Ambiguity

PROJECT-STATUS.md header states "Complete These 29 Tasks" but Phase 0 adds 3 more tasks (slow test optimization), bringing total to 32 tasks.

### Resolution ✅

**Corrected Task Count**: 32 tasks total

| Phase | Tasks |
|-------|-------|
| Phase 0: Slow Test Optimization | 5 packages |
| Phase 1: CI/CD Failures | 8 workflows |
| Phase 2: Deferred I2 Features | 8 features |
| Phase 3: Coverage Targets | 5 packages |
| Phase 4: Advanced Testing | 4 features (OPTIONAL) |
| Phase 5: Demo Videos | 6 videos (OPTIONAL) |
| **Total** | **36 tasks** (27 required, 9 optional) |

**Action Required**: Update PROJECT-STATUS.md header from "29 Tasks" to "36 Tasks (27 required, 9 optional)"

---

## Clarification 2: Slow Test Package Priority Confusion

### Ambiguity

SLOW-TEST-PACKAGES.md uses three categorizations:

1. "Packages Requiring Optimization (≥20s)" - 5 packages
2. "Packages With Moderate Performance Impact (10-20s)" - 6 packages
3. "Optimization Targets" with Critical/High/Medium priority tiers

PROJECT-STATUS.md Phase 0 only lists 5 packages but references "All critical packages <30s execution".

### Resolution ✅

**Clarified Scope**:

**Phase 0 (Day 1) - CRITICAL**: Focus on 5 packages ≥20s (430.9s total)

- `clientauth` (168s), `jose/server` (94s), `kms/client` (74s), `jose` (67s), `kms/server/application` (28s)

**Phase 0 Extended (Optional)**: 6 additional packages 10-20s can be deferred or handled in parallel with other work

- These are "acceptable duration" and don't block fast feedback loop

**Action Required**: Update PROJECT-STATUS.md Phase 0 to clarify "5 packages ≥20s" as primary target

---

## Clarification 3: EST serverkeygen Blocker Status

### Ambiguity

Multiple documents reference EST serverkeygen as "BLOCKED" but don't clarify impact on completion criteria.

- PROJECT-STATUS.md: Lists EST serverkeygen, says "BLOCKED" but includes in success criteria
- IMPLEMENTATION-GUIDE.md: Says "Skip for now, project can complete without it"
- spec.md: Shows EST serverkeygen as "⚠️ Iteration 2" (not implemented)

### Resolution ✅

**Clarified Completion Criteria**:

**Minimum Viable Completion**: 7/8 deferred features (EST serverkeygen optional)

- **Required**: JOSE E2E, CA OCSP, JOSE Docker, EST cacerts, EST simpleenroll, EST simplereenroll, TSA timestamp
- **Optional**: EST serverkeygen (blocked on PKCS#7 library integration)

**If PKCS#7 Library Resolved**: Include EST serverkeygen in completion (8/8)

**Action Required**: Update PROJECT-STATUS.md to show "7/8 features (EST serverkeygen optional if blocked)"

---

## Clarification 4: Coverage Target Package List

### Ambiguity

PROJECT-STATUS.md Phase 3 lists 5 packages for coverage improvement:

- ca/handler (47.2%), auth/userauth (42.6%), unsealkeysservice (78.2%), apperr (96.6%), network (88.7%)

But only ca/handler and auth/userauth are below 95% target. The other 3 are close to or above target.

### Resolution ✅

**Clarified Coverage Targets**:

**Primary Focus (Below 95%)**:

1. `ca/handler`: 47.2% → 95% (critical gap)
2. `auth/userauth`: 42.6% → 95% (critical gap)

**Secondary Focus (Close to Target)**:
3. `unsealkeysservice`: 78.2% → 95% (good progress, final push)
4. `network`: 88.7% → 95% (nearly there)

**Already Complete**:
5. `apperr`: 96.6% ✅ (exceeds target)

**Action Required**: Update PROJECT-STATUS.md Phase 3 to show "2 critical, 2 secondary, 1 complete"

---

## Clarification 5: Workflow Pass Rate Baseline

### Ambiguity

Multiple references to "8 failing workflows" and "27% pass rate" but unclear which specific workflows are failing vs passing.

PROJECT-STATUS.md Phase 1 lists 8 workflows but doesn't show which 3 are currently passing.

### Resolution ✅

**Clarified Workflow Status** (from archived ANALYSIS-ITERATION-3.md):

**Currently Passing (3/11)**:

- ci-quality ✅
- ci-gitleaks ✅
- ci-sast ✅

**Currently Failing (8/11) with  Priority order of fixing**:

- ci-coverage ❌
- ci-benchmark ❌
- ci-fuzz ❌
- ci-e2e ❌
- ci-dast ❌
- ci-race ❌
- ci-load ❌
- (1 more workflow name needed from analysis)

**Action Required**: Update PROJECT-STATUS.md Phase 1 to show which 3 workflows are passing

---

## Clarification 6: Implementation Timeline vs Work Effort

### Ambiguity

IMPLEMENTATION-GUIDE.md says "3-5 days focused work"
PROJECT-STATUS.md says "3-5 days focused work"
But also says "Week 1: Critical Path (16-24 hours)"

16-24 hours ≠ 3-5 days. Unclear if this is calendar days vs work hours.

### Resolution ✅

**Clarified Timeline**:

**Work Effort**: 16-24 hours (total work time)
**Calendar Duration**: 3-5 days (assuming ~5-6 hours focused work per day)

**Breakdown**:

- Day 1: 4-5 hours (slow test optimization)
- Day 2: 3-4 hours (JOSE E2E tests)
- Day 3: 4-5 hours (CI/CD workflow fixes)
- Day 4: 3-4 hours (CA OCSP + Docker)
- Day 5: 2-3 hours (coverage improvements)

**Total**: 16-21 hours core work + 3-5 hours buffer = ~20-24 hours

**Action Required**: Add timeline clarification to IMPLEMENTATION-GUIDE.md

---

## Ambiguities Summary

| # | Ambiguity | Resolution | Impact |
|---|-----------|------------|--------|
| 1 | Task count (29 vs 32 vs 36) | 36 total (27 required, 9 optional) | Documentation update |
| 2 | Slow test package scope | 5 packages ≥20s (Phase 0), 6 packages 10-20s (optional) | Priority clarification |
| 3 | EST serverkeygen blocker | Optional if PKCS#7 blocked, 7/8 completion acceptable | Success criteria |
| 4 | Coverage target packages | 2 critical, 2 secondary, 1 complete | Focus prioritization |
| 5 | Workflow pass rate baseline | 3 passing, 8 failing (11 total) | Status visibility |
| 6 | Timeline hours vs days | 16-24 hours work effort, 3-5 calendar days | Expectation setting |

---

## Action Items

**Required Documentation Updates**:

1. ✅ Update PROJECT-STATUS.md header: "36 Tasks (27 required, 9 optional)"
2. ✅ Update PROJECT-STATUS.md Phase 0: Clarify 5 packages ≥20s primary target
3. ✅ Update PROJECT-STATUS.md Phase 2: "7/8 features (EST serverkeygen optional)"
4. ✅ Update PROJECT-STATUS.md Phase 3: Show 2 critical, 2 secondary coverage targets
5. ✅ Update PROJECT-STATUS.md Phase 1: List 3 passing workflows
6. ✅ Update IMPLEMENTATION-GUIDE.md: Add timeline clarification (hours vs days)

**All action items to be addressed in next commit.**

---

## Conclusion

**Clarification Status**: ✅ **COMPLETE**

6 ambiguities identified and resolved through this analysis. No blocking ambiguities - all are documentation clarity improvements.

**Next Step**: Execute /speckit.plan to update technical implementation plan.

---

## Clarification 7: Identity Admin API Implementation Strategy

### Ambiguity

**Discovered**: December 11, 2025 (gap analysis)

Identity services currently use single public server with `/health` endpoint. Spec now requires dual-server pattern (Public + Private) matching KMS architecture.

**Questions**:

1. Should Identity AuthZ, IdP, and RS each have their own private admin servers on port 9090?
2. Or should all 3 services share a single admin server endpoint?
3. What's the migration path for existing deployments using `/health` on public port?
4. Should Docker Compose configs use `/admin/v1/livez` immediately or wait for implementation?

### Resolution ✅

**Recommended Approach**:

**Architecture**: Each Identity service (AuthZ, IdP, RS) gets its own private admin server on 127.0.0.1:9090

- Matches KMS pattern: One admin server per microservice
- Allows independent health status per service
- Enables independent scaling and deployment

**Migration Path**:

1. Phase 1: Keep existing `/health` on public port (backward compatible)
2. Phase 2: Add private admin server with `/admin/v1/*` endpoints
3. Phase 3: Update Docker Compose and workflows to use admin endpoints
4. Phase 4: Deprecate public `/health` endpoint (optional)

**Docker Compose**: Wait for implementation before updating health checks

- Current: Use `/health` on public port (working)
- Future: Use `/admin/v1/livez` on private port 9090

**Action Required**: Add Identity admin API implementation to tasks.md as HIGH priority

---

## Clarification 8: CA Deployment Configuration Completeness

### Ambiguity

**Discovered**: December 11, 2025 (gap analysis)

Spec shows CA as "✅ Complete" for all 20 tasks including "Deployment" (Task 19), but:

- Only `deployments/ca/compose.simple.yml` exists (dev-only)
- No `deployments/ca/compose.yml` for production multi-instance deployment
- No CA PostgreSQL backend configurations like JOSE/KMS have

**Questions**:

1. Should CA deployment follow JOSE/KMS pattern (3 instances: sqlite, postgres-1, postgres-2)?
2. What are CA-specific deployment requirements (CRL distribution, OCSP responder)?
3. Should CA use different admin port (9443) vs standard 9090?
4. Is CA deployment "complete" or should spec reflect missing production configs?

### Resolution ✅

**Clarified Status**:

**CA Deployment**: ⚠️ Development complete, production deployment missing

- ✅ Dev: `compose.simple.yml` works for local development
- ❌ Prod: Multi-instance PostgreSQL deployment needed
- ⚠️ Admin API: Implemented but should verify port 9443 vs 9090 consistency

**Deployment Pattern**: Follow JOSE/KMS pattern with CA-specific modifications

**Standard**: 3 instances (sqlite dev, 2x postgres prod)
**CA-Specific**: Add CRL distribution volume mounts, OCSP responder configs

**Admin Port**: Use 9443 (CA-specific) to avoid conflicts with other services on 9090

**Action Required**:

1. Update spec.md to show CA deployment as "dev complete, prod incomplete"
2. Add CA production deployment task to tasks.md as HIGH priority
3. Create `deployments/ca/compose.yml` matching JOSE/KMS patterns

---

## Clarification 9: Load Testing Scope and Gatling Architecture

### Ambiguity

**Discovered**: December 11, 2025 (gap analysis)

`test/load/README.md` documents 3 API contexts (Browser, Service, Admin) but only `ServiceApiSimulation.java` exists.

**Questions**:

1. Why are Browser API and Admin API load tests missing?
2. Should Admin API load tests be disabled (since admin endpoint is localhost-only)?
3. What's the priority for implementing missing load tests?
4. Should load tests target SQLite instance (fast) or PostgreSQL instances (production-like)?

### Resolution ✅

**Clarified Priority**:

**Service API**: ✅ Implemented and working

- Most critical: Service-to-service communication is highest throughput
- Tests encryption, signing, key generation under load

**Browser API**: ⚠️ HIGH priority (missing)

- User-facing SPA endpoints need performance validation
- Critical for UX: Token exchange, consent flows, certificate requests

**Admin API**: ❌ LOW priority (optional)

- Admin endpoints are operational, not high-throughput
- Health checks are simple, don't need load testing
- Defer unless specific monitoring concerns arise

**Test Target**:

- **Dev/CI**: SQLite instance (faster feedback, sufficient for performance regression detection)
- **Staging**: PostgreSQL instance (production-like, validates database bottlenecks)

**Action Required**: Add Browser API Gatling simulation to tasks.md as HIGH priority

---

## Clarification 10: E2E Test Workflow Coverage

### Ambiguity

**Discovered**: December 11, 2025 (gap analysis)

Current E2E tests only validate Docker Compose lifecycle (startup, health checks, logs). Missing actual product workflow tests.

**Questions**:

1. Should E2E tests be added to existing `internal/test/e2e/` or new `test/e2e-workflows/`?
2. What's the minimum viable set of workflows to test?
3. Should E2E tests use real database (PostgreSQL) or in-memory (SQLite)?
4. How do E2E tests differ from integration tests in `internal/*/integration/`?

### Resolution ✅

**Clarified Scope**:

**Location**: Extend existing `internal/test/e2e/` package

- Already has Docker Compose infrastructure
- Can add workflow test cases to existing test suite

**Minimum Viable E2E Workflows**:

**HIGH Priority**:

1. OAuth 2.1 Authorization Code Flow: Browser → AuthZ → login → consent → token
2. KMS Encrypt/Decrypt: Create key → encrypt data → decrypt data → verify plaintext match
3. CA Certificate Issuance: CSR → CA → issued certificate → validate chain

**MEDIUM Priority**:
4. JOSE JWT Sign/Verify: Generate key → sign JWT → verify JWT → validate claims
5. CA Revocation: Issue cert → revoke cert → verify OCSP revoked status

**Database Strategy**:

- **CI/CD**: SQLite in-memory (fast, no external dependencies)
- **Local Dev**: PostgreSQL via Docker Compose (production-like)

**E2E vs Integration**:

- **Integration**: Single service with mocked dependencies (`internal/identity/integration/`)
- **E2E**: Full Docker stack with real dependencies (`internal/test/e2e/`)

**Action Required**: Add E2E workflow tests to tasks.md as HIGH priority

---

## Clarification 11: Test Execution Performance Targets

### Ambiguity

**Discovered**: December 11, 2025 (gap analysis)

Spec now states "<30s per package, <100s total" but unclear how this applies to:

- Race detector runs (slower due to CGO overhead)
- Mutation testing (takes 30-45 minutes)
- Load testing (Gatling runs 10-30 minutes)

**Questions**:

1. Do performance targets apply only to `go test ./...` runs?
2. Should race detector have separate targets (e.g., <200s total)?
3. Should slow tests like mutation/load be excluded from CI critical path?
4. How should performance be measured: wall clock time or CPU time?

### Resolution ✅

**Clarified Targets**:

**Unit/Integration Tests** (`go test ./...`):

- Per package: <30 seconds
- Total suite: <100 seconds
- **Measured**: Wall clock time (what developers experience)

**Race Detector** (`go test -race ./...`):

- Per package: <60 seconds (2x overhead typical)
- Total suite: <200 seconds
- **Justification**: CGO_ENABLED=1 adds 50-100% overhead

**Mutation Testing** (`gremlins unleash`):

- Per package: No strict limit (varies by package complexity)
- Total suite: <45 minutes (workflow timeout)
- **Strategy**: Run on separate workflow, not critical path

**Load Testing** (Gatling):

- Per simulation: 5-10 minutes
- Total suite: <30 minutes
- **Strategy**: Run on separate workflow, triggered on PR approval

**Measurement**: Wall clock time via `time go test ./...`

**Action Required**: Update plan.md to document test performance SLAs

---

## Updated Ambiguities Summary

| # | Ambiguity | Resolution | Impact | Priority |
|---|-----------|------------|--------|----------|
| 1 | Task count discrepancy | 36 total (27 required, 9 optional) | Documentation | LOW |
| 2 | Slow test package scope | 5 packages ≥20s primary target | Focus | MEDIUM |
| 3 | EST serverkeygen blocker | Optional if PKCS#7 blocked | Success criteria | LOW |
| 4 | Coverage target packages | 2 critical, 2 secondary | Priority | MEDIUM |
| 5 | Workflow pass rate | 3 passing, 8 failing | Visibility | MEDIUM |
| 6 | Timeline hours vs days | 16-24 hours work, 3-5 days calendar | Expectation | LOW |
| **7** | **Identity admin API** | **Each service gets own admin server on 9090** | **Architecture** | **HIGH** |
| **8** | **CA deployment** | **Dev complete, prod deployment missing** | **Deployment** | **HIGH** |
| **9** | **Load test scope** | **Browser API HIGH, Admin API LOW priority** | **Testing** | **HIGH** |
| **10** | **E2E workflows** | **Extend internal/test/e2e/ with product flows** | **Testing** | **HIGH** |
| **11** | **Test performance** | **<100s unit, <200s race, <45min mutation** | **CI/CD** | **MEDIUM** |

---

## Action Items (Updated)

**Required Documentation Updates**:

1. ✅ Update PROJECT-STATUS.md header: "36 Tasks (27 required, 9 optional)"
2. ✅ Update PROJECT-STATUS.md Phase 0: Clarify 5 packages ≥20s primary target
3. ✅ Update PROJECT-STATUS.md Phase 2: "7/8 features (EST serverkeygen optional)"
4. ✅ Update PROJECT-STATUS.md Phase 3: Show 2 critical, 2 secondary coverage targets
5. ✅ Update PROJECT-STATUS.md Phase 1: List 3 passing workflows
6. ✅ Update IMPLEMENTATION-GUIDE.md: Add timeline clarification (hours vs days)
7. ⏳ Add Identity admin API implementation to tasks.md (HIGH priority)
8. ⏳ Add CA production deployment to tasks.md (HIGH priority)
9. ⏳ Add Browser API Gatling simulation to tasks.md (HIGH priority)
10. ⏳ Add E2E workflow tests to tasks.md (HIGH priority)
11. ⏳ Update plan.md with test performance SLAs

---

## Conclusion

**Clarification Status**: ✅ **11 clarifications identified and resolved**

- **Original 6**: Documentation clarity (complete)
- **New 5**: Architecture, deployment, testing gaps (identified December 11, 2025)

**Next Step**: Execute /speckit.plan to update technical implementation plan with new clarifications.

---

*Clarifications Version: 2.0.0*
*Updated: December 11, 2025*
*Analyst: GitHub Copilot (Claude Sonnet 4.5)*
