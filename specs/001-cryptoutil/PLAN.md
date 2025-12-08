# Technical Implementation Plan - Post-Consolidation

**Date**: December 7, 2025
**Context**: Updated implementation plan after consolidating iteration files
**Status**: ✅ PLAN COMPLETE - 5-phase approach with clear priorities

---

## Plan Overview

**Goal**: Complete cryptoutil project to 100% functional state with all 4 products working

**Approach**: 5-phase sequential execution (Phase 0-4 required, Phase 5 optional)

**Timeline**: 16-24 hours work effort, 3-5 calendar days

### CRITICAL: Execution Mandate

**WORK CONTINUOUSLY until ≥990k tokens OR user says "STOP"**:

- Complete task → immediately start next task
- Push changes → immediately continue working
- Update docs → immediately start next task
- NO stopping to provide summaries
- NO asking for permission between tasks
- NO pausing after git operations
- Token budget: 1M tokens, stop at 990k minimum

---

## Service Architecture - Dual HTTPS Endpoint Pattern

**CRITICAL: ALL services MUST implement dual HTTPS endpoints - NO HTTP PORTS**

### Architecture Requirements

Every service implements two HTTPS endpoints:

1. **Public HTTPS Endpoint** (configurable port, default 8080+)
   - Serves business APIs and browser UI
   - TLS required (never HTTP)
   - TWO security middleware stacks on SAME OpenAPI spec:
     - **Service-to-service APIs**: Require OAuth 2.1 client credentials tokens
     - **Browser-to-service APIs/UI**: Require OAuth 2.1 authorization code + PKCE tokens
   - Middleware enforces authorization boundaries
   
2. **Private HTTPS Endpoint** (always 127.0.0.1:9090 or similar)
   - Admin/operations endpoints: `/livez`, `/readyz`, `/healthz`, `/shutdown`
   - Localhost-only binding (not externally accessible)
   - TLS required (never HTTP)
   - Used by Docker health checks, Kubernetes probes, monitoring

### Service Port Assignments

| Service | Public HTTPS | Private HTTPS | Notes |
|---------|--------------|---------------|-------|
| KMS | :8080 | 127.0.0.1:9090 | Key management APIs |
| Identity AuthZ | :8080 | 127.0.0.1:9090 | OAuth 2.1 endpoints |
| Identity IdP | :8081 | 127.0.0.1:9090 | OIDC endpoints |
| Identity RS | :8082 | 127.0.0.1:9090 | Resource server |
| JOSE | :8080 | 127.0.0.1:9092 | JWK/JWT operations |
| CA | :8443 | 127.0.0.1:9443 | Certificate operations |

### Implementation Checklist

- ✅ NO HTTP endpoints on ANY port
- ✅ Health checks use HTTPS with `--no-check-certificate`
- ✅ Admin endpoints bound to 127.0.0.1 only
- ✅ Public endpoints support both service and browser clients
- ✅ Different OAuth token flows per client type

---

## Phase 0: Optimize Slow Test Packages (Day 1, 4-5 hours) 🚀

**Objective**: Enable fast development feedback loop by optimizing 5 slowest test packages

**Rationale**: Currently 430.9s combined execution time blocks rapid iteration. Optimizing these packages first provides foundation for efficient development in subsequent phases.

### Implementation Strategy

| Package | Current | Target | Strategy | Effort |
|---------|---------|--------|----------|--------|
| `clientauth` | 168s | <30s | Aggressive `t.Parallel()`, split test files by auth method | 2h |
| `jose/server` | 94s | <20s | Parallel subtests, reduce Fiber setup/teardown | 1h |
| `kms/client` | 74s | <20s | Mock KMS dependencies, parallel execution | 1h |
| `jose` | 67s | <15s | Increase coverage 48.8%→95% first, then optimize | 0.5h |
| `kms/server/app` | 28s | <10s | Parallel server tests, dynamic port allocation | 0.5h |

**Success Criteria**:
- All 5 packages execute in <30s each
- Total execution time <100s (down from 430.9s)
- No test failures introduced
- Coverage maintained or improved

---

## Phase 1: Fix CI/CD Workflows (Day 3, 4-5 hours) ⚠️

**Objective**: Achieve 11/11 workflow pass rate (currently 3/11 passing, 27%)

**Rationale**: CI/CD reliability is critical for continuous integration and deployment confidence.

### Failing Workflows (8 total)

| Workflow | Current | Root Cause | Fix Strategy | Effort |
|----------|---------|------------|--------------|--------|
| ci-dast | Failing | Service connectivity issues | Fix health check timing, HTTPS endpoints | 1h |
| ci-e2e | Failing | Docker Compose setup | Fix service dependencies, wait patterns | 1h |
| ci-load | Failing | Gatling test configuration | Update load test scenarios | 0.5h |
| ci-coverage | Failing | Test execution issues | Fix coverage aggregation | 0.5h |
| ci-race | Failing | Race condition detection | Fix identified race conditions | 1h |
| ci-benchmark | Failing | Benchmark test failures | Update benchmark baselines | 0.5h |
| ci-fuzz | Failing | Fuzz test configuration | Fix fuzz test execution | 0.5h |
| (1 more) | Failing | TBD from analysis | TBD | 0.5h |

**Success Criteria**:
- All 11 workflows passing (100% pass rate)
- CI feedback loop <10 minutes
- No flaky tests

---

## Phase 2: Complete Deferred I2 Features (Day 2 + 4, 6-8 hours) 🔧

**Objective**: Finish 7/8 deferred Iteration 2 features (EST serverkeygen optional)

### Day 2: JOSE E2E Tests (3-4 hours)

**Scope**: Comprehensive E2E test suite for JOSE Authority service

**Implementation**:
- Create `internal/jose/server/*_integration_test.go` files
- Test all 10 JOSE API endpoints end-to-end
- Validate JWK generation, JWKS endpoints, JWS sign/verify, JWE encrypt/decrypt, JWT issue/validate
- Integration with Docker Compose deployment

**Success Criteria**:
- All 10 JOSE endpoints tested end-to-end
- Tests run in <2 minutes total
- Coverage >95% for JOSE server package

### Day 4: CA OCSP + Docker Integration (3-4 hours)

**CA OCSP Responder** (2 hours):
- Implement OCSP responder endpoint `/ca/v1/ocsp`
- Support RFC 6960 OCSP protocol
- Return certificate status (good, revoked, unknown)
- Integration with CRL generation

**JOSE Docker Integration** (1-2 hours):
- Add `jose-sqlite`, `jose-postgres-1`, `jose-postgres-2` services to Docker Compose
- Configure ports (8080-8082 public, 9090 admin)
- Health checks via wget
- Integration with common infrastructure (postgres, otel-collector)

**Success Criteria**:
- OCSP responder returns valid responses
- JOSE services start and pass health checks
- Docker Compose deployment working end-to-end

---

## Phase 3: Achieve Coverage Targets (Day 5, 2-3 hours) 📊

**Objective**: Get 5/5 packages to ≥95% coverage

### Critical Gaps (2 packages, 2 hours)

| Package | Current | Gap | Strategy | Effort |
|---------|---------|-----|----------|--------|
| `ca/handler` | 47.2% | -47.8% | Add handler tests for all endpoints | 1h |
| `auth/userauth` | 42.6% | -52.4% | Add authentication flow tests | 1h |

### Secondary Targets (2 packages, 1 hour)

| Package | Current | Gap | Strategy | Effort |
|---------|---------|-----|----------|--------|
| `unsealkeysservice` | 78.2% | -16.8% | Add edge case tests | 0.5h |
| `network` | 88.7% | -6.3% | Add error path tests | 0.5h |

**Success Criteria**:
- All 5 packages ≥95% coverage
- No coverage regressions in other packages
- Tests use `t.Parallel()` and UUIDv7 patterns

---

## Phase 4: Advanced Testing (Optional, 4-6 hours) 🧪

**Objective**: Add advanced testing methodologies for quality assurance

### Benchmark Tests (2 hours)

- Add `_bench_test.go` files to cryptographic packages
- Benchmark key generation, encryption, signing operations
- Establish performance baselines

### Fuzz Tests (2 hours)

- Add `_fuzz_test.go` files to parsers and validators
- Fuzz JWT parsing, certificate parsing, input validation
- Run for 15s minimum per test

### Property-Based Tests (2 hours)

- Add `_property_test.go` files using gopter
- Test round-trip properties (encrypt→decrypt, sign→verify)
- Validate cryptographic invariants

**Success Criteria**:
- Benchmarks established for all crypto operations
- Fuzz tests run without failures
- Property tests validate core invariants

---

## Phase 5: Documentation & Demo (Optional, 8-12 hours) 📹

**Objective**: Create demo videos showcasing each product

### Demo Video Plan

| Demo | Duration | Content | Effort |
|------|----------|---------|--------|
| JOSE Authority | 5-10 min | JWK generation, signing, encryption | 2h |
| Identity Server | 10-15 min | OAuth flow, OIDC, MFA | 2-3h |
| KMS | 10-15 min | Key hierarchy, encryption, rotation | 2-3h |
| CA Server | 10-15 min | Certificate issuance, revocation, OCSP | 2-3h |
| Integration | 15-20 min | All 4 products working together | 3-4h |
| Unified Suite | 20-30 min | Complete deployment walkthrough | 3-4h |

**Success Criteria**:
- All 6 demo videos recorded and published
- Demos show working functionality
- Narration explains architecture and features

---

## Success Metrics

### Required for Completion

- ✅ Phase 0: All 5 slow packages <30s execution
- ✅ Phase 1: 11/11 CI/CD workflows passing
- ✅ Phase 2: 7/8 deferred features complete (EST serverkeygen optional)
- ✅ Phase 3: 5/5 packages ≥95% coverage

### Optional Enhancements

- ⚠️ Phase 4: Advanced testing methodologies added
- ⚠️ Phase 5: Demo videos created

### Quality Gates

- All linting passes (`golangci-lint run`)
- All tests pass (`go test ./... -cover -shuffle=on`)
- No CRITICAL/HIGH security vulnerabilities
- Integration demos work (`go run ./cmd/demo all`)

---

## Risk Management

| Risk | Impact | Mitigation |
|------|--------|------------|
| Test optimization breaks tests | HIGH | Incremental changes, verify after each package |
| CI/CD fixes introduce flakiness | MEDIUM | Use health checks, exponential backoff patterns |
| Coverage improvements reduce quality | LOW | Require meaningful tests, not just coverage numbers |
| EST serverkeygen remains blocked | LOW | Already optional, 7/8 completion acceptable |

---

## Dependencies

**External**:
- PKCS#7 library (optional, for EST serverkeygen)
- None other - all work is internal implementation

**Internal**:
- Phase 0 must complete before Phases 1-3 (foundation)
- Phase 2 (JOSE E2E) before Phase 1 (some CI workflows depend on it)
- Phase 3 can run in parallel with Phases 1-2
- Phases 4-5 only after Phases 0-3 complete

---

## Conclusion

**Plan Status**: ✅ **COMPLETE AND EXECUTABLE**

This plan provides:
- Clear 5-phase sequential approach
- Effort estimates per task
- Success criteria per phase
- Risk mitigation strategies
- Dependency management

**Next Step**: Execute /speckit.tasks to generate detailed task breakdown.

---

*Technical Implementation Plan Version: 1.0.0*
*Author: GitHub Copilot (Agent)*
*Approved: Pending user validation*
