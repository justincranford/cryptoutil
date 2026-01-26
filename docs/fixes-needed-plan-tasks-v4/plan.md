# Implementation Plan - Remaining Work (V4)

**Status**: Planning (111 tasks total - 68 from v3 + 43 new coverage tasks)
**Created**: 2026-01-26
**Last Updated**: 2026-01-27
**Previous Version**: docs/fixes-needed-plan-tasks-v3/ (47/115 tasks complete, 40.9%)

## Overview

This plan contains the **remaining incomplete work** from v3 PLUS **new coverage improvement work** based on comprehensive analysis. V3 achieved significant milestones:

- ✅ Template: 98.91% mutation efficacy (exceeds 98% ideal)
- ✅ JOSE-JA: 97.20% mutation efficacy (exceeds 98% ideal)
- ✅ Phase 4.2: Pflag refactor complete (92.5% coverage)
- ✅ Phase 8.5: Docker health checks 100% standardized

**V4 Additions**:
- ✅ Coverage analysis complete: 52.2% total (test-output/coverage-analysis/)
- ✅ Gap identification: 15+ packages below ≥98% minimum
- 🆕 Phases 8-12: Coverage improvement to reach ≥98% minimum (≥99% ideal) for ALL packages

**Remaining Work**: 111 tasks across 12 phases (68 legacy + 43 coverage improvement)

## Technical Context

- **Language**: Go 1.25.5
- **Framework**: Service template pattern with dual HTTPS servers
- **Database**: PostgreSQL OR SQLite with GORM
- **Testing**: ≥98% coverage ideal (≥95% minimum), ≥98% mutation efficacy ideal (≥95% mandatory minimum)
- **Services**:
  - Template: ✅ 98.91% efficacy
  - JOSE-JA: ✅ 97.20% efficacy
  - Cipher-IM: ⏳ BLOCKED (needs Docker fixes)
  - KMS: ⏳ Not started

## Phases

### Phase 0: Research & Discovery

**Objective**: Clarify ambiguities before implementation

**Tasks**:
- [ ] 0.1: Service template comparison analysis
  - Compare KMS vs service-template vs cipher-im vs JOSE-JA
  - Identify duplication opportunities
  - Create comparison table in research.md
- [ ] 0.2: Mutation efficacy standards clarification
  - Document 98% ideal vs 95% minimum distinction
  - Update plan.md quality gates
- [ ] 0.3: CI/CD mutation workflow research
  - Linux execution requirements
  - Timeout configuration per package
  - Artifact collection patterns

**Deliverables**:
- research.md with comparison table
- Updated plan.md quality gates section
- CI/CD execution checklist

### Phase 1: JOSE-JA Service Error Coverage

**Objective**: Achieve 95% coverage for jose/service (currently 87.3%, gap: 7.7%)

**Current Status**:
- Coverage: 87.3% (target: 95%)
- Blocker: Closed DB testing limitation (documented in 4.2.10)
- Discovery: Multi-step error paths need decomposition

**Tasks** (6 tasks):
- [ ] 1.1: Add createMaterialJWK error tests
- [ ] 1.2: Add Encrypt error tests
- [ ] 1.3: Add RotateMaterial error tests
- [ ] 1.4: Add CreateEncryptedJWT error tests
- [ ] 1.5: Add EncryptWithKID error tests
- [ ] 1.6: Verify 95% coverage achieved

**Success Criteria**:
- Coverage ≥95% for jose/service
- All error paths tested independently
- No skipped tests without documentation

### Phase 2: Cipher-IM Infrastructure Fixes

**Objective**: Unblock cipher-im mutation testing (currently 0% - UNACCEPTABLE)

**Root Cause**: Docker compose unhealthy, E2E tag bypass, repository timeouts

**Tasks** (5 tasks):
- [ ] 2.1: Fix cipher-im Docker infrastructure
  - OTEL HTTP/gRPC mismatch resolution
  - E2E tag bypass fix
  - Health check verification
- [ ] 2.2: Run gremlins baseline on cipher-im
- [ ] 2.3: Analyze cipher-im lived mutations
- [ ] 2.4: Kill cipher-im mutations for 98% efficacy (HIGH)
- [ ] 2.5: Verify cipher-im mutation testing complete

**Success Criteria**:
- Docker compose healthy (all services pass health checks)
- Gremlins runs successfully without timeouts
- Mutation efficacy ≥98% (ideal target)

### Phase 3: Template Mutation Cleanup

**Objective**: Address remaining template mutations (currently 98.91% efficacy)

**Status**: Template already exceeds 98% target, but 1 lived mutation remains

**Tasks** (Optional - deferred LOW priority):
- [ ] 3.1: Analyze remaining tls_generator.go mutation
- [ ] 3.2: Determine if killable or inherent limitation
- [ ] 3.3: Implement test if feasible

**Success Criteria**:
- Document mutation as killable or inherent limitation
- Update mutation-analysis.md with findings

### Phase 4: Continuous Mutation Testing

**Objective**: Enable automated mutation testing in CI/CD

**Dependencies**: Phase 2 complete (cipher-im unblocked)

**Tasks** (6 tasks):
- [ ] 4.1: Verify ci-mutation.yml workflow
- [ ] 4.2: Configure timeout (per package)
- [ ] 4.3: Set efficacy threshold enforcement (95% required)
- [ ] 4.4: Test workflow with actual PR
- [ ] 4.5: Document in README.md and DEV-SETUP.md
- [ ] 4.6: Commit continuous mutation testing

**Success Criteria**:
- ci-mutation.yml runs on every PR
- Enforces 95% minimum efficacy
- Documents workflow in README

### Phase 5: CI/CD Mutation Campaign

**Objective**: Execute first Linux-based mutation testing campaign

**Dependencies**: Phase 4 complete

**Tasks** (11 tasks):
- [ ] 5.1: Monitor workflow execution at GitHub Actions
- [ ] 5.2: Download mutation-test-results artifact
- [ ] 5.3: Analyze gremlins output
- [ ] 5.4: Populate mutation-baseline-results.md
- [ ] 5.5: Commit baseline analysis
- [ ] 5.6: Review survived mutations
- [ ] 5.7: Categorize by mutation type
- [ ] 5.8: Write targeted tests for survived mutations
- [ ] 5.9: Re-run ci-mutation.yml workflow
- [ ] 5.10: Verify efficacy ≥95% for all packages
- [ ] 5.11: Commit mutation-killing tests

**Success Criteria**:
- All packages ≥95% efficacy (minimum)
- Baseline results documented
- CI/CD workflow passing

### Phase 6: Automation & Branch Protection

**Objective**: Enforce mutation testing on every PR

**Dependencies**: Phase 5 complete

**Tasks** (6 tasks):
- [ ] 6.1: Add workflow trigger: on: [push, pull_request]
- [ ] 6.2: Configure path filters (code changes only)
- [ ] 6.3: Add status check requirement in branch protection
- [ ] 6.4: Document in README.md and DEV-SETUP.md
- [ ] 6.5: Test with actual PR
- [ ] 6.6: Commit automation

**Success Criteria**:
- Mutation testing runs on every code change
- Branch protection enforces passing mutation tests
- Documented in project README

### Phase 7: Race Condition Testing

**Objective**: Verify thread-safety on Linux with race detector

**Current Status**: 35 tasks UNMARKED for Linux re-testing

**Tasks** (35 tasks organized by category):

**Repository Layer** (7 tasks):
- [ ] 7.1: Run race detector on jose-ja repository
- [ ] 7.2: Run race detector on cipher-im repository
- [ ] 7.3: Run race detector on template repository
- [ ] 7.4: Document any race conditions found
- [ ] 7.5: Fix races with proper mutex/channel usage
- [ ] 7.6: Re-run until clean (0 races detected)
- [ ] 7.7: Commit repository thread-safety verified on Linux

**Service Layer** (7 tasks):
- [ ] 7.8: Run race detector on jose-ja service
- [ ] 7.9: Run race detector on cipher-im service
- [ ] 7.10: Run race detector on template service
- [ ] 7.11: Document races
- [ ] 7.12: Fix races
- [ ] 7.13: Re-run until clean
- [ ] 7.14: Commit service thread-safety

**APIs Layer** (7 tasks):
- [ ] 7.15-7.21: Similar pattern for APIs layer

**Config Layer** (7 tasks):
- [ ] 7.22-7.28: Similar pattern for config layer

**Integration Tests** (7 tasks):
- [ ] 7.29-7.35: Similar pattern for integration tests

**Success Criteria**:
- All packages pass race detector (0 races)
- Fixes documented and committed
- Linux CI/CD race testing enabled

### Phase 8: Zero Coverage Packages

**Objective**: Establish test infrastructure for 0% coverage packages (infrastructure code)

**Current Status**: Application lifecycle, server builders, and infrastructure packages untested (0% coverage)

**Evidence**: test-output/coverage-analysis/gaps-analysis.md

**Tasks**:
- [ ] 8.1: Add tests for shared/container utilities
- [ ] 8.2: Add tests for shared/magic crypto functions
- [ ] 8.3: Add tests for apps/*/server/application lifecycle (StartBasic, Shutdown, InitializeServicesOnCore)
- [ ] 8.4: Add tests for apps/*/server/builder (server builder pattern)
- [ ] 8.5: Add tests for apps/*/server/listener (application listeners)
- [ ] 8.6: Add tests for apps/template/service/client (authentication client)
- [ ] 8.7: Document test patterns for infrastructure testing
- [ ] 8.8: Verify all previously 0% packages now ≥95%

**Success Criteria**:
- All previously 0% packages ≥95% coverage (≥98% ideal)
- Infrastructure testing patterns documented
- Test patterns reusable across services

**Note**: E2E infrastructure (apps/template/testing/e2e) excluded - test utilities expected to have lower coverage

### Phase 9: Severe Coverage Gaps

**Objective**: Improve pool and telemetry packages to ≥98% coverage

**Current Status**:
- shared/pool: 61.5% (need +36.5%)
- shared/telemetry: 67.5% (need +30.5%)

**Critical Functions Identified** (from function-level analysis):
- pool: closeChannelsThread 42.9%
- telemetry: initMetrics 48.9%, initTraces 48.6%, checkSidecarHealth 40.0%

**Tasks**:
- [ ] 9.1: Add unit tests for pool worker thread management
- [ ] 9.2: Add tests for pool cleanup (closeChannelsThread edge cases)
- [ ] 9.3: Add tests for pool error paths
- [ ] 9.4: Add tests for telemetry initMetrics with all backends
- [ ] 9.5: Add tests for telemetry initTraces with all configurations
- [ ] 9.6: Add tests for telemetry checkSidecarHealth (failure scenarios)
- [ ] 9.7: Add integration tests for telemetry with otel-collector
- [ ] 9.8: Verify pool ≥98% coverage
- [ ] 9.9: Verify telemetry ≥98% coverage

**Success Criteria**:
- shared/pool ≥98% coverage (from 61.5%)
- shared/telemetry ≥98% coverage (from 67.5%)
- All critical functions (currently <50%) ≥95%

### Phase 10: Barrier Services Coverage

**Objective**: Improve barrier service packages to ≥98% coverage

**Current Status**:
- barrier/intermediatekeysservice: 76.8% (need +21.2%)
- barrier/rootkeysservice: 79.0% (need +19.0%)
- barrier/unsealkeysservice: 89.8% (need +8.2%)

**Critical Functions Identified**:
- intermediate_keys_service: EncryptKey 72.7%, DecryptKey 70.0%
- root_keys_service: EncryptKey 72.7%
- unseal_keys_service: encryptKey 75.0%

**Tasks**:
- [ ] 10.1: Add unit tests for intermediate key encryption/decryption edge cases
- [ ] 10.2: Add unit tests for root key encryption/decryption edge cases
- [ ] 10.3: Add unit tests for unseal key encryption/decryption edge cases
- [ ] 10.4: Add integration tests for key hierarchy (unseal → root → intermediate)
- [ ] 10.5: Add error path tests (invalid keys, corrupted ciphertext)
- [ ] 10.6: Add concurrent operation tests (thread-safety verification)
- [ ] 10.7: Verify intermediatekeysservice ≥98%
- [ ] 10.8: Verify rootkeysservice ≥98%
- [ ] 10.9: Verify unsealkeysservice ≥98%

**Success Criteria**:
- All barrier services ≥98% coverage
- Key hierarchy integration tests passing
- Encryption/decryption edge cases covered

### Phase 11: Crypto Core Coverage

**Objective**: Improve crypto packages to ≥98% coverage

**Current Status**:
- crypto/certificate: 78.2% (need +19.8%)
- crypto/jose: 82.6% (need +15.4%)
- crypto/password: 81.8% (need +16.2%)
- crypto/pbkdf2: 85.4% (need +12.6%)
- crypto/tls: 85.8% (need +12.2%)
- crypto/keygen: 85.2% (need +12.8%)

**Critical Functions Identified**:
- jose: CreateJWEJWKFromKey 60.4%, CreateJWKFromKey 59.1%, EnsureSignatureAlgorithmType 23.1%
- certificate: startTLSEchoServer 56.5%

**Tasks**:
- [ ] 11.1: Add tests for crypto/jose key creation functions (CreateJWKFromKey, CreateJWEJWKFromKey)
- [ ] 11.2: Add tests for crypto/jose algorithm validation (EnsureSignatureAlgorithmType)
- [ ] 11.3: Add tests for crypto/certificate TLS server utilities
- [ ] 11.4: Add tests for crypto/password edge cases
- [ ] 11.5: Add tests for crypto/pbkdf2 parameter variations
- [ ] 11.6: Add tests for crypto/tls configuration edge cases
- [ ] 11.7: Add tests for crypto/keygen error paths
- [ ] 11.8: Verify all crypto packages ≥98%

**Success Criteria**:
- All crypto packages ≥98% coverage
- Key creation functions (currently <60%) ≥95%
- Algorithm validation comprehensive

### Phase 12: Near-Ideal Package Polish

**Objective**: Bring near-ideal packages (96-97%) to ≥99% coverage

**Current Status**:
- crypto/digests: 96.9% (need +2.1% for 99%)
- shared/util/network: 96.8% (need +2.2% for 99%)

**Tasks**:
- [ ] 12.1: Identify remaining uncovered lines in crypto/digests
- [ ] 12.2: Add tests for identified gaps in crypto/digests
- [ ] 12.3: Identify remaining uncovered lines in shared/util/network
- [ ] 12.4: Add tests for identified gaps in shared/util/network
- [ ] 12.5: Verify crypto/digests ≥99%
- [ ] 12.6: Verify shared/util/network ≥99%

**Success Criteria**:
- crypto/digests ≥99% coverage
- shared/util/network ≥99% coverage
- Demonstrates ≥99% ideal achievable

## Technical Decisions

### Decision 1: Mutation Efficacy Standards

**Chosen**: 98% IDEAL target (all packages), 95% MANDATORY MINIMUM (documented blockers only)
**Rationale**: V3 achieved 98.91% (Template) and 97.20% (JOSE-JA), proving 98% is achievable. 95% is floor, not target.
**Alternatives**: 85% minimum (REJECTED - too low), 95% universal target (REJECTED - sets bar at floor)
**Impact**: Higher quality standard, but v3 proves feasibility

**Documentation Needed**: Update plan.md quality gates to clarify distinction

### Decision 2: Service Template Reusability

**Chosen**: Maximize reuse in service-template, minimize duplication in services
**Rationale**: KMS, cipher-im, JOSE-JA all reimplementing similar patterns
**Alternatives**: Continue with duplication (rejected - maintenance burden)
**Impact**: Requires comparative analysis (Phase 0.1)

**Research Needed**: Compare KMS vs service-template vs cipher-im vs JOSE-JA implementations

### Decision 3: Cipher-IM Priority

**Chosen**: Unblock cipher-im BEFORE continuing other work
**Rationale**: Zero mutation testing coverage is UNACCEPTABLE
**Alternatives**: Skip cipher-im (rejected - violates "no services skipped" principle)
**Impact**: Phase 2 becomes critical blocker

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Cipher-IM Docker issues persist | Medium | High | Dedicated Phase 2, fallback: simplify compose |
| Race detector finds major issues | Medium | Medium | Phase 7 dedicated, prioritize fixes by severity |
| CI/CD mutation timeouts | Low | Medium | Timeout per package, parallelize execution |
| Service template refactor too large | Low | High | Phase 0.1 research quantifies scope, break into sub-phases if needed |
| 95% service coverage unreachable | Low | Medium | Document testing limitations (as done in 4.2.10), accept <95% with justification |

## Quality Gates

**Per-Task**:
- ✅ All tests pass (runTests)
- ✅ Coverage maintained or improved
- ✅ No new TODOs without tracking
- ✅ Linting clean (golangci-lint run)
- ✅ Commit follows conventional format

**Per-Phase**:
- ✅ Phase objectives achieved
- ✅ Success criteria verified with evidence
- ✅ Documentation updated (research.md, tasks.md, completed.md)
- ✅ Quality gates enforced

**Overall Project**:
- ✅ Mutation efficacy ≥98% ideal for ALL services (Template ✅, JOSE-JA ✅, Cipher-IM ⏳, KMS ⏳)
- ✅ Coverage ≥98% minimum (≥99% ideal) for ALL packages (currently 52.2%)
- ✅ Zero packages below ≥95% minimum (currently 15+ packages below)
- ✅ Infrastructure packages (pool, telemetry, barrier) ≥98%
- ✅ Crypto core packages ≥98%
- ✅ Application lifecycle and server builders ≥95%
- ✅ Race detector clean on Linux (0 races)
- ✅ CI/CD mutation testing automated
- ✅ Service template maximally reusable

## Success Criteria

- [ ] All 111 tasks complete (68 legacy + 43 coverage)
- [ ] Quality gates pass
- [ ] All services ≥98% mutation efficacy (ideal target)
- [ ] **ALL packages ≥98% coverage minimum (≥99% ideal)**
- [ ] **Project total coverage ≥95% (from current 52.2%)**
- [ ] Race detector clean on Linux
- [ ] CI/CD enforces mutation testing
- [ ] Service template comparison documented
- [ ] Documentation updated (README, DEV-SETUP, research.md)

## RECENT ACHIEVEMENTS (from V3)

**Phase 6.3: Template Mutation Testing**
- Efficacy: 98.91% (exceeds 98% ideal) ✅
- Coverage: 81.3% → 82.5%
- Tests: 4 functions, 14 subtests
- Commits: 5d68b8dc, eea5e19f

**Phase 6.2: JOSE-JA Mutation Testing**
- Efficacy: 97.20% (exceeds 98% ideal) ✅
- Verified: All gremlins pass

**Phase 4.2: JOSE-JA Pflag Refactor**
- Coverage: 61.9% → 92.5% (+30.6%)
- Tests: Comprehensive ParseWithFlagSet tests
- Commits: f8f8436c, 06ba5a94

**Phase 8.5: Docker Health Checks**
- Standardization: 100% (all 13 compose files)
- Cipher-IM: UNBLOCKED for mutation testing
- Commits: 32740220, 4a28a12b
