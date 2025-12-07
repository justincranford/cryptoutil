# Iteration 2 Clarifications Checklist

## Purpose

This document identifies ambiguities, incomplete items, and clarifications needed before proceeding to Iteration 3.

**Created**: December 6, 2025  
**Updated**: December 7, 2025 - CORRECTED per user feedback  
**Iteration**: 2  
**Status**: ✅ All Critical Items Resolved

---

## 1. Test Runtime Optimization Strategy ✅

### User Provided Clarification

**Question**: How should test runtime optimization be implemented for selective test execution during local development vs CI?

**User Decision** ✅:

> Do selective (mix of deterministic and/or randomized) test execution to speed up slowest 10 Go packages during local unit/integration tests, for faster feedback. CI probably needs to do more or all tests, versus local execution of tests.

**Implementation Strategy** (Based on Top 100 OSS Projects Research):

**Option A: Seeded Randomness (Elasticsearch Pattern)**:

- Use magic constants for percentage-based test execution (0.0-1.0)
- Example: `randomBoolean()` (50%), `randomLong(0, 100)`, `randomString("joe", "justin", "john")`
- Apply to individual tests OR code branches (if/else/switch)
- Centralize percentages as magic constants for easy tuning
- Allows fast local iterations while maintaining coverage sampling

**Option B: Canary/Smoke Tests**:

- Define subset of tests that always run (smoke tests)
- Example: KMS client doesn't need every key length (AES 128/192/256, RSA 2048/3072/4096, EC P256/P384/P521)
- Test one representative key length per algorithm during local dev
- Full matrix runs in CI only

**Option C: Subtest Sampling**:

- Keep single `TestX` function
- Randomly/conditionally run heavy sub-cases
- More granular than build tags

**CRITICAL REQUIREMENTS**:

- ✅ **ALWAYS use concurrent test execution** (`t.Parallel()` + `-shuffle`)
- ✅ **NEVER use `-p=1`** (sequential execution)
- ✅ **ALWAYS use UUIDv7 for test data uniqueness** (thread-safe, process-safe)
- ✅ **ALWAYS use dynamic ports** (port 0 pattern for servers)
- ✅ **ALWAYS use TestMain for test dependencies** (start once per package, reuse across tests)

**Test Dependency Pattern** (PostgreSQL Example):

```go
var testDB *sql.DB

func TestMain(m *testing.M) {
    // Start PostgreSQL container ONCE per package
    testDB = startPostgreSQLContainer()
    exitCode := m.Run()
    testDB.Close()
    os.Exit(exitCode)
}

func TestUserCreate(t *testing.T) {
    t.Parallel() // Safe - each test uses unique UUIDv7 data
    userID := googleUuid.NewV7()
    user := &User{ID: userID, Name: "test-" + userID.String()}
    // Test creates orthogonal data - no conflicts
}
```

**Implementation Plan**:

- Apply seeded randomness to slowest packages (clientauth: 168s, jose/server: 94s, kms/client: 74s)
- Define smoke test subset for KMS crypto operations (one key length per algorithm)
- Use subtest sampling for heavy integration tests
- Document patterns in `01-02.testing.instructions.md`

**Priority**: MEDIUM  
**Assigned To**: Iteration 3 - Task PERF-1

---

## 2. CMS and PKCS#7 Library Selection ✅

### User Provided Clarification

**Question**: Which Go library should be used for CMS (Cryptographic Message Syntax) and PKCS#7 operations?

**User Decision** ✅:

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

## 3. Product Database Architecture ✅

### User Provided Clarification

**Question**: Should JOSE Authority share PostgreSQL with Identity/KMS or have separate database?

**User Decision** ✅:

> All products MUST use separate DB. JOSE product must have its own SQLite/PostgreSQL DB support. KMS product must have its own SQLite/PostgreSQL DB support. Identity product must have its own SQLite/PostgreSQL DB support. Certificate Authority product must have its own SQLite/PostgreSQL DB support.

**Rationale**:

- Each product is independently start/stop via docker compose
- Isolated E2E testing from other products
- Separate DB enables true product independence

**Future Enhancement (TBD)**:

- Single PostgreSQL instance with multiple logical DBs (e.g., `kms1`, `kms2`, `kms3`, `jose1`, `jose2`, `ca1-5`, `identity1-4`)
- Parameterized instances per product in compose.yml
- Requires config/code support for dynamic logical DB naming
- Complexity vs benefit TBD

**Implementation Plan**:

- JOSE: Create SQLite + PostgreSQL support (same pattern as KMS/Identity/CA)
- Docker Compose: Separate PostgreSQL instance per product
- Configuration: Each product has independent DB connection config

**Priority**: HIGH  
**Assigned To**: Iteration 3 - Task JOSE-13

---

## 4. CA OCSP Handler Configuration ✅

### User Provided Clarification

**Question**: OCSP responder configuration details?

**User Decision** ✅:

> Dedicated OCSP signing cert, 24h validity, 1h cache

**Implementation Plan**:

- Create dedicated OCSP signing certificate (separate from CA cert)
- OCSP response validity: 24 hours
- OCSP response caching: 1 hour
- Implement RFC 6960 OCSP responder
- OCSP request parsing and validation
- OCSP response signing
- Integration with certificate storage

**Priority**: HIGH  
**Assigned To**: Iteration 3 - Task CA-11

---

## 5. EST Protocol Configuration ✅

### User Provided Clarification

**Question**: EST authentication method and server-side key generation preferences?

**User Decision** ✅:

**EST Authentication**:

> Configurable support for HTTP Basic Auth only, Client certificate, or both!

**Server-Side Key Generation**:

> Configurable per certificate profile; for example, storing encryption private keys should be configurable, and default to stored server-side, and storing signing private keys should be configurable, but default to no server-side storage. Also, HSM/TPM for key storage must be configurable per certificate profile too.

**Implementation Plan**:

**EST Endpoints**:

- CA-14: `/est/cacerts` - Get CA certificate chain
- CA-15: `/est/simpleenroll` - Simple enrollment
- CA-16: `/est/simplereenroll` - Certificate renewal
- CA-17: `/est/serverkeygen` - Server-side key generation

**Authentication**:

- Support HTTP Basic Auth
- Support Client certificate (mTLS)
- Support both simultaneously (configurable per certificate profile)

**Server-Side Key Generation**:

- Configurable per certificate profile
- Encryption keys: Default to server-side storage
- Signing keys: Default to NO server-side storage
- HSM/TPM support: Configurable per certificate profile

**Priority**: MEDIUM  
**Assigned To**: Iteration 3 - Tasks CA-14 through CA-17

---

## 6. CA Integration Tests Strategy ✅

### User Provided Clarification

**Question**: Test data strategy for CA integration tests?

**User Decision** ✅:

> CA Integration Tests must generate certificates on-the-fly in tests, MUST ALWAYS use FIPS 140-3 approved algorithms, but can use smallest key size allowed by FIPS.

**Implementation Requirements**:

- Generate certificates on-the-fly (no pre-generated fixtures for primary tests)
- ALWAYS use FIPS 140-3 approved algorithms
- Use smallest FIPS-approved key sizes for speed (RSA 2048, EC P-256, AES 128)
- Test certificate lifecycle: issue → retrieve → revoke → check status
- Test CRL generation and retrieval
- Test OCSP responder
- Test EST enrollment flows
- Error scenario testing
- Coverage target: ≥95% (currently 1.2%)

**Dependency Pattern**:

- Use TestMain to start test dependencies ONCE per package
- Real dependencies preferred over mocks (PostgreSQL via test container, in-memory telemetry)
- Mocks only for hard-to-reach corner cases or truly external dependencies

**Priority**: HIGH  
**Assigned To**: Iteration 3 - Task CA-19

---

## 7. Coverage Target Increments ✅

### User Provided Clarification

**Question**: Should we increment test coverage targets in future iterations?

**User Decision** ✅:

> Assume no. Acceptable runtime regression for coverage increase: Up to 10% slower

**Current Targets** (Constitution v2.0.0):

- 95%+ production coverage
- 100% infrastructure (cicd)
- 100% utility code

**Philosophy**:

- 95/100/100 is sustainable long-term target
- Focus on quality of coverage (meaningful tests) over quantity
- Up to 10% slower test execution acceptable for coverage improvements
- No further target increments planned

**Priority**: LOW  
**Assigned To**: N/A (targets frozen)

---

## 8. Gremlins Mutation Testing Integration ✅

### User Provided Clarification

**Question**: How should gremlins mutation testing be integrated into development workflow?

**User Decision** ✅:

> Start with Option 2 (manual) + Option 4 (periodic), graduate to Option 1 (CI) once baseline established. NEVER block pre-commit or pre-push on gremlins, it would be too slow.

**Implementation Plan**:

- **Phase 1** (Current): Manual execution for critical crypto operations
- **Phase 2** (After baseline): Periodic reports (weekly/monthly)
- **Phase 3** (After maturity): CI workflow integration (non-blocking)
- **NEVER**: Pre-commit or pre-push blocking (too slow)

**Gremlins v0.6.0 Blocker**:

- Tool crashes with panic "error, this is temporary"
- Manual mutation testing interim solution for critical paths
- Monitor upstream releases for v0.6.1+ fixes
- Document blocker in `docs/todos-gremlins.md` and CLARIFICATIONS.md #11

**Priority**: LOW (Blocked by tooling)  
**Assigned To**: Iteration 4 or later (after gremlins fix)

---

## 9. Workflow Improvements Strategy ✅

### User Provided Clarification

**Question**: How should workflow improvements be prioritized and implemented?

**User Decision** ✅:

**Fix Phase** (All Failed Workflows):

1. Identify workflows triggered by last push
2. Filter out workflows that passed
3. Fix each workflow, making one or more commits per workflow
4. Batch push the commits
5. Repeat until all workflows pass

**Optimize Phase** (Slowest Workflows):

1. Identify workflows slowest to fastest
2. Filter out any faster than 5 minutes
3. Optimize each workflow, making one or more commits per workflow
4. Batch push the commits
5. Repeat until all slow workflows are faster than 10 minutes

**Success Criteria**:

- Fastest workflows: <5 minutes
- Slowest workflows: <10 minutes
- All workflows: 100% pass rate

**Priority**: HIGH  
**Assigned To**: Iteration 3 - Phase 1 (Fix), Phase 4 (Optimize)

---

## 10. Iteration 3 Scope ✅

### User Provided Clarification

**Question**: Should Iteration 3 focus on completing Iteration 2 items or adding new features?

**User Decision** ✅:

> Option A Complete Iteration 2 Items. Defer new capabilities (gremlins, property tests, etc.) to next iteration.

**Iteration 3 Scope**:

- Complete ALL remaining iteration 2 tasks (14 tasks, ~40 hours)
- JOSE Docker Integration (JOSE-13)
- CA OCSP Handler (CA-11)
- CA EST Protocol Endpoints (CA-14 through CA-17)
- CA Integration Tests to ≥95% coverage (CA-19)
- Unified E2E Test Suite
- Workflow fixes and optimizations

**Deferred to Iteration 4**:

- Gremlins mutation testing integration (tooling blocked)
- Property-based testing (gopter)
- Advanced EST features (reenroll, serverkeygen)
- Additional test methodology enhancements

**Priority**: CRITICAL  
**Assigned To**: Iteration 3 - All Phases

---

## 11. Iteration 2 Completion vs Iteration 3 Start ✅

### User Provided Clarification

**Question**: Should we complete ALL iteration 2 tasks before starting iteration 3?

**User Decision** ✅:

> Complete ALL iteration 2 tasks before starting iteration 3.

**Completion Criteria**:

- All 47 iteration 2 tasks at 100% (currently 83%, 39/47)
- Coverage targets met (≥95% production, ≥100% infrastructure)
- All workflows passing (100% pass rate)
- Docker integration complete for all products
- E2E test suite operational

**Priority**: CRITICAL  
**Assigned To**: Iteration 3 is iteration 2 completion

---

## 12. Gremlins Mutation Testing Tool - BLOCKED ⚠️

### Tool Status

**Current State**:

- ✅ Gremlins v0.6.0 installed successfully
- ✅ Configuration file `.gremlins.yaml` exists
- ❌ Tool crashes on execution: `panic: error, this is temporary`

**Resolution**:

- Constitution v2.0.0 made mutation testing mandatory ≥80% score
- **Interim Solution**: Manual mutation testing for critical crypto operations
- **Long-term**: Monitor upstream for v0.6.1+ fixes, revisit in iteration 4
- Amendment to Constitution v2.0.1: Make mutation testing RECOMMENDED (not mandatory) until tooling stable

**Priority**: HIGH (Blocks quality gates)  
**Impact**: Cannot achieve ≥80% mutation score requirement  
**Assigned To**: Pending gremlins upstream fix

**Reference**: `docs/todos-gremlins.md`

---

## Summary of Clarifications

| ID | Topic | Status | Priority |
|----|-------|--------|----------|
| 1 | Test Runtime Optimization | ✅ RESOLVED | MEDIUM |
| 2 | CMS/PKCS#7 Library | ✅ RESOLVED | LOW |
| 3 | Product Database Architecture | ✅ RESOLVED | HIGH |
| 4 | CA OCSP Configuration | ✅ RESOLVED | HIGH |
| 5 | EST Protocol Configuration | ✅ RESOLVED | MEDIUM |
| 6 | CA Integration Tests Strategy | ✅ RESOLVED | HIGH |
| 7 | Coverage Target Philosophy | ✅ RESOLVED | LOW |
| 8 | Gremlins Integration | ✅ RESOLVED | LOW |
| 9 | Workflow Improvements Strategy | ✅ RESOLVED | HIGH |
| 10 | Iteration 3 Scope | ✅ RESOLVED | CRITICAL |
| 11 | Iteration 2 Completion Gate | ✅ RESOLVED | CRITICAL |
| 12 | Gremlins Tool Blocker | ⚠️ BLOCKED | HIGH |

---

## Speckit Iteration Structure Clarification

**Question**: Is `specs/001-cryptoutil/` an iteration or does it contain multiple iterations?

**Answer**: `specs/001-cryptoutil/` is a **PRODUCT specification** containing **MULTIPLE iterations**:

- `specs/001-cryptoutil/CHECKLIST-ITERATION-1.md` - Iteration 1 completion checklist
- `specs/001-cryptoutil/CHECKLIST-ITERATION-2.md` - Iteration 2 completion checklist
- `specs/001-cryptoutil/CHECKLIST-ITERATION-3.md` - Iteration 3 completion checklist

**Structure**:

```
specs/001-cryptoutil/          # Product: cryptoutil (all iterations)
├── spec.md                     # Product requirements (all iterations)
├── plan.md                     # Implementation plan (all iterations)
├── tasks.md                    # Task breakdown (all iterations)
├── PROGRESS.md                 # Iteration progress tracking
├── CHECKLIST-ITERATION-1.md    # Iteration 1 gates
├── CHECKLIST-ITERATION-2.md    # Iteration 2 gates
├── CHECKLIST-ITERATION-3.md    # Iteration 3 gates
├── ANALYSIS.md                 # Technical analysis
├── CLARIFICATIONS.md           # Ambiguity tracking
└── EXECUTIVE-SUMMARY.md        # Stakeholder overview
```

**Why Multiple Iterations Per Product**:

- Product evolves incrementally (iteration 1 → 2 → 3 → N)
- Each iteration adds features, improves quality, completes deferred work
- CHECKLIST-ITERATION-N.md tracks completion gates per iteration
- Single product spec accumulates all iteration requirements

---

*Clarifications Version: 3.0.0 (CORRECTED)*  
*Created: December 6, 2025*  
*Updated: December 7, 2025 - All clarifications resolved*  
*Status: ✅ Ready for Iteration 3 Execution*
