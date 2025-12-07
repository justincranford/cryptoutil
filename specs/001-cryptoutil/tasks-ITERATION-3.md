# cryptoutil Tasks - Iteration 3

## Task Breakdown

This document provides granular task tracking for Iteration 3 implementation (CI/CD reliability + deferred work).

**Total Phases**: 4
**Total Tasks**: 19
**Estimated Effort**: ~44 hours (1 week sprint)

---

## CRITICAL: Test Concurrency Requirements

**!!! NEVER use `-p=1` or `-parallel=1` in test commands !!!**
**!!! ALWAYS use concurrent test execution with `-shuffle=on` !!!**

**Test Execution Commands**:

```bash
# CORRECT - Concurrent with shuffle
go test ./... -cover -shuffle=on

# WRONG - Sequential execution (hides bugs!)
go test ./... -p=1  # ❌ NEVER DO THIS
go test ./... -parallel=1  # ❌ NEVER DO THIS
```

**Test Data Isolation Requirements**:

- ✅ ALWAYS use UUIDv7 for all test data (thread-safe, process-safe)
- ✅ ALWAYS use dynamic ports (port 0 pattern for test servers)
- ✅ ALWAYS use TestMain for dependencies (start once per package)
- ✅ Real dependencies preferred (PostgreSQL containers, in-memory services)
- ✅ Mocks only for hard-to-reach corner cases or truly external dependencies

**Why Concurrent Testing is Mandatory**:

1. Fastest test execution (parallel tests = faster feedback)
2. Reveals production bugs (race conditions, deadlocks, data conflicts)
3. Production validation (if tests can't run concurrently, production code can't either)
4. Quality assurance (concurrent tests = higher confidence)

---

## Phase 1: Critical CI/CD Fixes (Days 1-2, ~16h)

### Task List

| ID | Title | Priority | Status | LOE | Dependencies |
|----|-------|----------|--------|-----|--------------|
| ITER3-001 | Fix DATA RACE in CA handler | CRITICAL | ❌ Not Started | 4h | None |
| ITER3-002 | Increase Identity ORM coverage 67.5% → 95% | CRITICAL | ❌ Not Started | 4h | None |
| ITER3-003 | Fix consent_decision_repository_test.go:160 | HIGH | ❌ Not Started | 2h | ITER3-002 |
| ITER3-004 | Debug E2E/DAST/Load Docker Compose startup | CRITICAL | ❌ Not Started | 4h | None |
| ITER3-005 | Verify all workflows pass (11/11) | HIGH | ❌ Not Started | 2h | ITER3-001 to 004 |

### ITER3-001: Fix DATA RACE in CA Handler

**Description**: Race condition at `handler_comprehensive_test.go:1502` (goroutines 339/341) causing ci-race.yml 100% failure.

**Acceptance Criteria**:

- [ ] Race eliminated (verified with `go test -race`)
- [ ] All CA handler tests pass with `-race` flag
- [ ] ci-race.yml passes in CI

**Files to Modify**:

- `internal/ca/server/handler/handler_comprehensive_test.go` (line 1502 area)

**Evidence**: `go test -race ./internal/ca/server/handler/...` clean + ci-race.yml passing

---

### ITER3-002: Increase Identity ORM Coverage

**Description**: Coverage at 67.5%, need ≥95%. Add missing test cases for uncovered paths.

**Acceptance Criteria**:

- [ ] Coverage ≥95% all Identity ORM packages
- [ ] ci-coverage.yml passes
- [ ] Table-driven tests with `t.Parallel()`

**Files to Modify**:

- `internal/identity/server/repository/sqlrepository/*_test.go`
- `internal/identity/domain/*_test.go`

**Evidence**: `go test -cover ./internal/identity/...` shows ≥95% + ci-coverage.yml passing

---

### ITER3-003: Fix Consent Decision Test

**Description**: Test failure at `consent_decision_repository_test.go:160`.

**Acceptance Criteria**:

- [ ] Test passes consistently (100% success rate)
- [ ] No flakiness (10 consecutive runs)

**Files to Modify**:

- `internal/identity/server/repository/sqlrepository/consent_decision_repository_test.go`

**Evidence**: Test passes 10 consecutive runs + ci-coverage.yml passing

---

### ITER3-004: Debug E2E/DAST/Load Startup

**Description**: Service startup failures in Docker Compose (E2E/DAST/Load 100% failure).

**Acceptance Criteria**:

- [ ] All services start successfully
- [ ] Health checks pass within timeout
- [ ] ci-e2e.yml, ci-dast.yml, ci-load.yml pass

**Files to Modify**:

- `deployments/compose/compose.yml` (health checks)
- `internal/*/server/application/application.go` (startup logging)

**Evidence**: `docker compose up` all healthy + workflows passing

---

### ITER3-005: Verify All Workflows Pass

**Description**: Confirm all 11 workflows pass after Phase 1 fixes.

**Acceptance Criteria**:

- [ ] 11/11 workflows passing (100% success rate)
- [ ] CI feedback loop ≤10 minutes

**Evidence**: `gh run list` shows 11/11 passing + screenshot

---

## Phase 2: Deferred Work Completion (Days 3-4, ~15h)

### Task List

| ID | Title | Priority | Status | LOE | Dependencies |
|----|-------|----------|--------|-----|--------------|
| ITER3-006 | JOSE Docker Integration | HIGH | ❌ Not Started | 2h | Phase 1 |
| ITER3-007 | CA OCSP Handler (RFC 6960) | HIGH | ❌ Not Started | 6h | Phase 1 |
| ITER3-008 | CA EST Handler (RFC 7030) | MEDIUM | ❌ Not Started | 4h | Phase 1 |
| ITER3-009 | Unified E2E Test Suite | HIGH | ❌ Not Started | 3h | ITER3-006 to 008 |

### ITER3-006: JOSE Docker Integration

**Description**: Create Dockerfile, compose.yml, demo for JOSE server.

**Acceptance Criteria**:

- [ ] `deployments/jose/Dockerfile.jose` created
- [ ] JOSE service in `compose.yml`
- [ ] `cmd/demo jose` command works

**Files to Create**:

- `deployments/jose/Dockerfile.jose`
- `deployments/jose/compose.jose.yml`
- `cmd/demo/jose.go`

**Evidence**: `docker compose up jose-server` + `cmd/demo jose` executes

---

### ITER3-007: CA OCSP Handler

**Description**: Implement RFC 6960 OCSP responder. Use `go.mozilla.org/pkcs7` per CLARIFICATIONS.md #2.

**Acceptance Criteria**:

- [ ] OCSP POST endpoint `/ocsp` functional
- [ ] CMS library integrated (go.mozilla.org/pkcs7)
- [ ] Tests ≥95% coverage

**Files to Create**:

- `internal/ca/server/handler/ocsp.go`
- `internal/ca/server/handler/ocsp_test.go`
- `api/ca/openapi_spec_paths.yaml` (add `/ocsp`)

**Evidence**: `go test ./internal/ca/server/handler/...` + API endpoint functional

---

### ITER3-008: CA EST Handler

**Description**: Implement RFC 7030 EST server (simple profile only).

**Acceptance Criteria**:

- [ ] EST endpoints: `/est/cacerts`, `/est/simpleenroll`
- [ ] Tests ≥95% coverage

**Files to Create**:

- `internal/ca/server/handler/est.go`
- `internal/ca/server/handler/est_test.go`
- `api/ca/openapi_spec_paths.yaml` (add `/est/*`)

**Evidence**: `go test ./internal/ca/server/handler/...` + API endpoints functional

---

### ITER3-009: Unified E2E Test Suite

**Description**: Cross-service E2E tests (JOSE + CA + Identity + PostgreSQL + OTEL).

**Acceptance Criteria**:

- [ ] E2E tests cover all 3 services
- [ ] Uses `internal/test/e2e/` infrastructure
- [ ] ci-e2e.yml passes

**Files to Create**:

- `internal/test/e2e/unified_test.go`

**Evidence**: `go test ./internal/test/e2e/...` + ci-e2e.yml passing

---

## Phase 3: Test Methodology Enhancements (Day 5, ~9h)

### Task List

| ID | Title | Priority | Status | LOE | Dependencies |
|----|-------|----------|--------|-----|--------------|
| ITER3-010 | Benchmarks: internal/common/crypto/keygen | MEDIUM | ❌ Not Started | 2h | Phase 2 |
| ITER3-011 | Benchmarks: internal/crypto/* | MEDIUM | ❌ Not Started | 2h | Phase 2 |
| ITER3-012 | Fuzz: internal/identity/authz parsers | MEDIUM | ❌ Not Started | 2h | Phase 2 |
| ITER3-013 | Fuzz: internal/jose parsers | MEDIUM | ❌ Not Started | 1h | Phase 2 |
| ITER3-014 | Property-based: gopter crypto invariants | LOW | ❌ Not Started | 2h | Phase 2 |

### ITER3-010: Benchmarks for keygen

**Description**: Add benchmarks for all `internal/common/crypto/keygen` operations (mandatory per constitution v2.0.0).

**Acceptance Criteria**:

- [ ] Benchmarks for RSA, ECDSA, ECDH, EdDSA, AES, HMAC key generation
- [ ] Baseline metrics documented

**Files to Create**:

- `internal/common/crypto/keygen/*_bench_test.go`

**Evidence**: `go test -bench=. -benchmem ./internal/common/crypto/keygen`

---

### ITER3-011: Benchmarks for internal/crypto

**Description**: Add benchmarks for all cryptographic operations in `internal/crypto`.

**Acceptance Criteria**:

- [ ] Benchmarks for encryption, decryption, signing, verification
- [ ] Happy and sad path benchmarks

**Files to Create**:

- `internal/crypto/*_bench_test.go`

**Evidence**: `go test -bench=. -benchmem ./internal/crypto`

---

### ITER3-012: Fuzz Tests for Identity Authz Parsers

**Description**: Add fuzz tests for all parsers/validators in `internal/identity/authz` (≥15s).

**Acceptance Criteria**:

- [ ] Fuzz tests for token parsers, validators
- [ ] Unique function names (not substrings)
- [ ] ≥15s fuzz time per test

**Files to Create**:

- `internal/identity/authz/*_fuzz_test.go`

**Evidence**: `go test -fuzz=FuzzName -fuzztime=15s ./internal/identity/authz`

---

### ITER3-013: Fuzz Tests for JOSE Parsers

**Description**: Add fuzz tests for JOSE parsers (JWS, JWE, JWK).

**Acceptance Criteria**:

- [ ] Fuzz tests for JWS/JWE/JWK parsers
- [ ] ≥15s fuzz time per test

**Files to Create**:

- `internal/jose/*_fuzz_test.go`

**Evidence**: `go test -fuzz=FuzzName -fuzztime=15s ./internal/jose`

---

### ITER3-014: Property-Based Tests (gopter)

**Description**: Add property-based tests for crypto invariants using gopter.

**Acceptance Criteria**:

- [ ] Tests for round-trip properties (encrypt/decrypt, sign/verify)
- [ ] Tests for idempotence, commutativity where applicable

**Files to Create**:

- `internal/crypto/*_property_test.go`
- `internal/common/crypto/keygen/*_property_test.go`

**Evidence**: `go test ./internal/crypto/... ./internal/common/crypto/keygen/...`

---

## Phase 4: Documentation & Optimization (Day 5, ~4h)

### Task List

| ID | Title | Priority | Status | LOE | Dependencies |
|----|-------|----------|--------|-----|--------------|
| ITER3-015 | Extract slow package data | LOW | ❌ Not Started | 1h | None |
| ITER3-016 | Update runbooks with workflow findings | MEDIUM | ❌ Not Started | 1h | Phase 1 |
| ITER3-017 | Delete processed DELETE-ME files | LOW | ❌ Not Started | 0.5h | ITER3-015 |
| ITER3-018 | Consolidate NOT-FINISHED.md | LOW | ❌ Not Started | 1h | Phase 2 |
| ITER3-019 | Apply workflow optimizations | MEDIUM | ❌ Not Started | 0.5h | Phase 1 |

### ITER3-015: Extract Slow Package Data

**Description**: Extract slow test package data from `DELETE-ME-LATER-SLOW-TEST-PACKAGES.md` for optimization tracking.

**Acceptance Criteria**:

- [ ] Slow packages documented in iteration 3 tracking
- [ ] Data extracted to PROGRESS.md or separate doc

**Evidence**: Data extracted + DELETE-ME file processed

---

### ITER3-016: Update Runbooks

**Description**: Update runbooks with findings from `docs/workflow-analysis.md`.

**Acceptance Criteria**:

- [ ] Runbooks updated with CI/CD fixes
- [ ] Troubleshooting sections enhanced

**Files to Modify**:

- `docs/runbooks/*.md`

**Evidence**: Runbooks reflect Phase 1 fixes

---

### ITER3-017: Delete DELETE-ME Files

**Description**: Delete processed DELETE-ME files.

**Acceptance Criteria**:

- [ ] DELETE-ME-LATER-SLOW-TEST-PACKAGES.md deleted
- [ ] DELETE-ME-LATER-CROSS-REF-SPECKIT-COPILOT-TEMPLATE.md deleted

**Evidence**: `git rm docs/DELETE-ME-LATER-*.md`

---

### ITER3-018: Consolidate NOT-FINISHED.md

**Description**: Consolidate `docs/NOT-FINISHED.md` into iteration 3 tracking.

**Acceptance Criteria**:

- [ ] Incomplete items moved to PROGRESS.md
- [ ] NOT-FINISHED.md deleted or updated

**Evidence**: NOT-FINISHED.md processed

---

### ITER3-019: Apply Workflow Optimizations

**Description**: Apply path filters, caching optimizations to workflows.

**Acceptance Criteria**:

- [ ] Path filters added to workflows (avoid unnecessary runs)
- [ ] Go module caching verified (actions/setup-go@v6)
- [ ] CI feedback loop <10 minutes

**Files to Modify**:

- `.github/workflows/*.yml`

**Evidence**: Workflow execution times documented <10min

---

## Overall Completion Checklist

### Phase 1: Critical CI/CD Fixes

- [ ] All 5 tasks complete (ITER3-001 to 005)
- [ ] Workflow pass rate: 11/11 (100%)
- [ ] ci-race.yml passing
- [ ] ci-coverage.yml passing (coverage ≥95%)
- [ ] ci-e2e.yml, ci-dast.yml, ci-load.yml passing

### Phase 2: Deferred Work

- [ ] All 4 tasks complete (ITER3-006 to 009)
- [ ] JOSE Docker operational
- [ ] CA OCSP handler functional
- [ ] CA EST handler functional
- [ ] Unified E2E test suite passing

### Phase 3: Test Enhancements

- [ ] All 5 tasks complete (ITER3-010 to 014)
- [ ] Benchmarks for all crypto operations
- [ ] Fuzz tests ≥15s for parsers
- [ ] Property-based tests operational

### Phase 4: Documentation & Optimization

- [ ] All 5 tasks complete (ITER3-015 to 019)
- [ ] DELETE-ME files processed
- [ ] Runbooks updated
- [ ] NOT-FINISHED.md consolidated
- [ ] Workflow optimizations applied
- [ ] CI feedback <10 minutes

### Overall Success Criteria

- [ ] 19/19 tasks complete
- [ ] Workflow pass rate: 11/11 (100%)
- [ ] Coverage: ≥95% production, ≥100% infrastructure
- [ ] CI feedback loop: <10 minutes
- [ ] Deferred features: 4/4 completed
- [ ] Test methodology: Benchmarks + fuzz + property operational
- [ ] Documentation: All cleanup items processed

**Iteration 3 Complete When**: All 4 phases verified + CHECKLIST-ITERATION-003.md ✅

---

## Template Usage Notes

**For LLM Agents**: This task breakdown for iteration 3 includes:

- ✅ 19 granular tasks across 4 phases
- ✅ LOE estimates per task (~44h total)
- ✅ Priority assignments (CRITICAL → LOW)
- ✅ Clear dependencies between tasks
- ✅ Acceptance criteria for each task
- ✅ Files to create/modify for each task
- ✅ Evidence requirements for completion verification
- ✅ Phase summaries with completion checklists
- ✅ Overall completion criteria

**Task Tracking Pattern**:

- Status: ❌ Not Started → ⚠️ In Progress → ✅ Complete
- Update status as work progresses
- Mark complete ONLY with evidence (no assumptions)
- Use PROGRESS.md for authoritative iteration tracking

**References**:

- spec.md: Iteration 3 functional requirements
- plan.md: 4-phase implementation approach
- docs/workflow-analysis.md: CI/CD findings
- specs/002-cryptoutil/CLARIFICATIONS.md: Deferred work
- .specify/memory/constitution.md v2.0.0: Quality requirements
