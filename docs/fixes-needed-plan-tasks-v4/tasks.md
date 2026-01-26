# Tasks - Remaining Work (V4)

**Status**: 0 of 68 tasks complete (0%) - Fresh start from v3 incomplete work
**Last Updated**: 2026-01-26
**Previous Version**: docs/fixes-needed-plan-tasks-v3/ (47/115 tasks complete, 40.9%)

## Phase 0: Research & Discovery

**Objective**: Clarify unknowns before implementation
**Estimated**: 4h total
**Status**: ⏳ NOT STARTED

### Task 0.1: Service Template Comparison Analysis

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 2h
**Actual**: 
**Dependencies**: None
**Priority**: HIGH

**Description**: Create comprehensive comparison table analyzing kms, service-template, cipher-im, and jose-ja implementations to identify code duplication, inconsistencies, and opportunities for service-template extraction.

**Acceptance Criteria**:
- [ ] 0.1.1: Read all four service implementations
  - internal/kms/server/ (reference KMS implementation)
  - internal/apps/template/service/ (extracted template)
  - internal/apps/cipher/im/service/ (cipher-im service)
  - internal/apps/jose/ja/service/ (jose-ja service)
- [ ] 0.1.2: Create comparison table with columns:
  - Component (Server struct, Config, Handlers, Middleware, TLS setup, etc.)
  - KMS implementation (file location, pattern used)
  - Service-template implementation (file location, pattern used)
  - Cipher-IM implementation (file location, pattern used)
  - JOSE-JA implementation (file location, pattern used)
  - Duplication analysis (identical, similar, different)
  - Reusability recommendation (extract to template, keep service-specific, etc.)
- [ ] 0.1.3: Document findings in research.md
- [ ] 0.1.4: Identify top 10 duplication candidates for extraction
- [ ] 0.1.5: Estimate effort to extract each candidate

**Files**:
- docs/fixes-needed-plan-tasks-v4/research.md (new)

---

### Task 0.2: Mutation Efficacy Standards Clarification

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 1h
**Actual**: 
**Dependencies**: None
**Priority**: MEDIUM

**Description**: Clarify and document the distinction between 98% IDEAL target and 85% MINIMUM acceptable mutation efficacy standards in plan.md quality gates section.

**Acceptance Criteria**:
- [ ] 0.2.1: Document 98% as IDEAL target (Template ✅ 98.91%, JOSE-JA ✅ 97.20%)
- [ ] 0.2.2: Document 85% as MINIMUM acceptable (with documented blockers only)
- [ ] 0.2.3: Update plan.md quality gates section with clear distinction
- [ ] 0.2.4: Add examples of acceptable blockers (test unreachable code, etc.)
- [ ] 0.2.5: Commit: "docs(plan): clarify mutation efficacy 98% ideal vs 85% minimum"

**Files**:
- docs/fixes-needed-plan-tasks-v4/plan.md (update)

---

### Task 0.3: CI/CD Mutation Workflow Research

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 1h
**Actual**: 
**Dependencies**: None
**Priority**: MEDIUM

**Description**: Research and document Linux-based CI/CD mutation testing execution requirements, timeout configurations, and artifact collection patterns.

**Acceptance Criteria**:
- [ ] 0.3.1: Review existing .github/workflows/ci-mutation.yml
- [ ] 0.3.2: Document Linux execution requirements
- [ ] 0.3.3: Document timeout configuration (15min per package recommended)
- [ ] 0.3.4: Document artifact collection patterns
- [ ] 0.3.5: Create CI/CD execution checklist in research.md
- [ ] 0.3.6: Commit: "docs(research): CI/CD mutation testing patterns"

**Files**:
- docs/fixes-needed-plan-tasks-v4/research.md (update)
- .github/workflows/ci-mutation.yml (reference)

---

## Phase 1: JOSE-JA Service Error Coverage

**Objective**: Achieve 95% coverage for jose/service (currently 87.3%, gap: 7.7%)
**Estimated**: 8h total
**Status**: ⏳ NOT STARTED

### Task 1.1: Add createMaterialJWK Error Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 1.5h
**Actual**: 
**Dependencies**: Task 0.1 complete (comparison analysis)
**Priority**: HIGH

**Description**: Add error path tests for createMaterialJWK function to improve coverage.

**Acceptance Criteria**:
- [ ] 1.1.1: Analyze createMaterialJWK error paths
- [ ] 1.1.2: Write tests for invalid parameters
- [ ] 1.1.3: Write tests for JWKGen errors
- [ ] 1.1.4: Write tests for database errors
- [ ] 1.1.5: Verify coverage improvement
- [ ] 1.1.6: All tests pass
- [ ] 1.1.7: Commit: "test(jose/service): add createMaterialJWK error tests"

**Files**:
- internal/apps/jose/ja/service/service_test.go (update)

---

### Task 1.2: Add Encrypt Error Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 1.5h
**Actual**: 
**Dependencies**: Task 1.1 complete
**Priority**: HIGH

**Description**: Add error path tests for Encrypt function.

**Acceptance Criteria**:
- [ ] 1.2.1: Analyze Encrypt error paths
- [ ] 1.2.2: Write tests for invalid plaintext
- [ ] 1.2.3: Write tests for encryption failures
- [ ] 1.2.4: Write tests for repository errors
- [ ] 1.2.5: Verify coverage improvement
- [ ] 1.2.6: All tests pass
- [ ] 1.2.7: Commit: "test(jose/service): add Encrypt error tests"

**Files**:
- internal/apps/jose/ja/service/service_test.go (update)

---

### Task 1.3: Add RotateMaterial Error Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 1.5h
**Actual**: 
**Dependencies**: Task 1.2 complete
**Priority**: HIGH

**Description**: Add error path tests for RotateMaterial function.

**Acceptance Criteria**:
- [ ] 1.3.1: Analyze RotateMaterial error paths
- [ ] 1.3.2: Write tests for invalid key IDs
- [ ] 1.3.3: Write tests for rotation failures
- [ ] 1.3.4: Write tests for database errors
- [ ] 1.3.5: Verify coverage improvement
- [ ] 1.3.6: All tests pass
- [ ] 1.3.7: Commit: "test(jose/service): add RotateMaterial error tests"

**Files**:
- internal/apps/jose/ja/service/service_test.go (update)

---

### Task 1.4: Add CreateEncryptedJWT Error Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 1.5h
**Actual**: 
**Dependencies**: Task 1.3 complete
**Priority**: HIGH

**Description**: Add error path tests for CreateEncryptedJWT function.

**Acceptance Criteria**:
- [ ] 1.4.1: Analyze CreateEncryptedJWT error paths
- [ ] 1.4.2: Write tests for invalid claims
- [ ] 1.4.3: Write tests for JWE creation failures
- [ ] 1.4.4: Write tests for signing errors
- [ ] 1.4.5: Verify coverage improvement
- [ ] 1.4.6: All tests pass
- [ ] 1.4.7: Commit: "test(jose/service): add CreateEncryptedJWT error tests"

**Files**:
- internal/apps/jose/ja/service/service_test.go (update)

---

### Task 1.5: Add EncryptWithKID Error Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 1.5h
**Actual**: 
**Dependencies**: Task 1.4 complete
**Priority**: HIGH

**Description**: Add error path tests for EncryptWithKID function.

**Acceptance Criteria**:
- [ ] 1.5.1: Analyze EncryptWithKID error paths
- [ ] 1.5.2: Write tests for invalid KID
- [ ] 1.5.3: Write tests for key not found
- [ ] 1.5.4: Write tests for encryption failures
- [ ] 1.5.5: Verify coverage improvement
- [ ] 1.5.6: All tests pass
- [ ] 1.5.7: Commit: "test(jose/service): add EncryptWithKID error tests"

**Files**:
- internal/apps/jose/ja/service/service_test.go (update)

---

### Task 1.6: Verify 95% JOSE/Service Coverage Achieved

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 30min
**Actual**: 
**Dependencies**: Task 1.5 complete
**Priority**: HIGH

**Description**: Run coverage analysis and verify jose/service package achieves ≥95% coverage.

**Acceptance Criteria**:
- [ ] 1.6.1: Run coverage analysis: `go test -cover ./internal/apps/jose/ja/service/`
- [ ] 1.6.2: Verify coverage ≥95%
- [ ] 1.6.3: Update tasks.md with actual coverage
- [ ] 1.6.4: Update plan.md with Phase 1 completion
- [ ] 1.6.5: Commit: "docs(v4): mark Phase 1 complete - 95% jose/service coverage"

**Files**:
- docs/fixes-needed-plan-tasks-v4/tasks.md (update)
- docs/fixes-needed-plan-tasks-v4/plan.md (update)

---

## Phase 2: Cipher-IM Infrastructure Fixes

**Objective**: Unblock cipher-im mutation testing (currently 0% - UNACCEPTABLE)
**Estimated**: 5h + 6-10h (mutation killing)
**Status**: ⏳ NOT STARTED

### Task 2.1: Fix Cipher-IM Docker Infrastructure

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 2h
**Actual**: 
**Dependencies**: Phase 1 complete
**Priority**: CRITICAL

**Description**: Fix Docker compose issues blocking cipher-im mutation testing (OTEL mismatch, E2E tag bypass, health checks).

**Acceptance Criteria**:
- [ ] 2.1.1: Resolve OTEL HTTP/gRPC mismatch
- [ ] 2.1.2: Fix E2E tag bypass issue
- [ ] 2.1.3: Verify health checks pass
- [ ] 2.1.4: Run `docker compose -f cmd/cipher-im/docker-compose.yml up -d`
- [ ] 2.1.5: All services healthy (0 unhealthy)
- [ ] 2.1.6: Commit: "fix(cipher-im): unblock Docker compose for mutation testing"

**Files**:
- cmd/cipher-im/docker-compose.yml (fix)
- configs/cipher/ (update)

---

### Task 2.2: Run Gremlins Baseline on Cipher-IM

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 1h
**Actual**: 
**Dependencies**: Task 2.1 complete
**Priority**: HIGH

**Description**: Run initial gremlins mutation testing campaign on cipher-im to establish baseline efficacy.

**Acceptance Criteria**:
- [ ] 2.2.1: Run: `gremlins unleash ./internal/apps/cipher/im/`
- [ ] 2.2.2: Collect output to /tmp/gremlins_cipher_baseline.log
- [ ] 2.2.3: Extract efficacy percentage
- [ ] 2.2.4: Document baseline in research.md
- [ ] 2.2.5: Commit: "docs(cipher-im): mutation baseline - XX.XX% efficacy"

**Files**:
- docs/fixes-needed-plan-tasks-v4/research.md (update)

---

### Task 2.3: Analyze Cipher-IM Lived Mutations

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 1h
**Actual**: 
**Dependencies**: Task 2.2 complete
**Priority**: HIGH

**Description**: Analyze survived mutations from gremlins run, categorize by type and priority.

**Acceptance Criteria**:
- [ ] 2.3.1: Parse gremlins output for lived mutations
- [ ] 2.3.2: Categorize by mutation type (arithmetic, conditionals, etc.)
- [ ] 2.3.3: Prioritize by ROI (test complexity vs efficacy gain)
- [ ] 2.3.4: Document in research.md
- [ ] 2.3.5: Create kill plan (target 98% efficacy)
- [ ] 2.3.6: Commit: "docs(cipher-im): mutation analysis with kill plan"

**Files**:
- docs/fixes-needed-plan-tasks-v4/research.md (update)

---

### Task 2.4: Kill Cipher-IM Mutations for 98% Efficacy

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 6-10h (HIGH priority, depends on mutation count)
**Actual**: 
**Dependencies**: Task 2.3 complete
**Priority**: CRITICAL

**Description**: Write targeted tests to kill survived mutations and achieve ≥98% efficacy ideal target.

**Acceptance Criteria**:
- [ ] 2.4.1: Implement tests for HIGH priority mutations
- [ ] 2.4.2: Implement tests for MEDIUM priority mutations
- [ ] 2.4.3: Re-run gremlins: `gremlins unleash ./internal/apps/cipher/im/`
- [ ] 2.4.4: Verify efficacy ≥98%
- [ ] 2.4.5: All tests pass
- [ ] 2.4.6: Coverage maintained or improved
- [ ] 2.4.7: Commit: "test(cipher-im): achieve 98% mutation efficacy - XX.XX%"

**Files**:
- internal/apps/cipher/im/repository/*_test.go (add tests)
- internal/apps/cipher/im/service/*_test.go (add tests)

---

### Task 2.5: Verify Cipher-IM Mutation Testing Complete

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 30min
**Actual**: 
**Dependencies**: Task 2.4 complete
**Priority**: HIGH

**Description**: Final verification that cipher-im achieves ≥98% mutation efficacy.

**Acceptance Criteria**:
- [ ] 2.5.1: Run final gremlins: `gremlins unleash ./internal/apps/cipher/im/`
- [ ] 2.5.2: Verify efficacy ≥98%
- [ ] 2.5.3: Update tasks.md with actual efficacy
- [ ] 2.5.4: Update plan.md with Phase 2 completion
- [ ] 2.5.5: Document in completed.md
- [ ] 2.5.6: Commit: "docs(v4): mark Phase 2 complete - cipher-im 98% efficacy"

**Files**:
- docs/fixes-needed-plan-tasks-v4/tasks.md (update)
- docs/fixes-needed-plan-tasks-v4/plan.md (update)
- docs/fixes-needed-plan-tasks-v4/completed.md (new)

---

## Phase 3: Template Mutation Cleanup (OPTIONAL - LOW PRIORITY)

**Objective**: Address remaining template mutation (currently 98.91% efficacy)
**Estimated**: 2h total
**Status**: ⏳ DEFERRED (template already exceeds 98% target)

### Task 3.1: Analyze Remaining tls_generator.go Mutation

**Status**: ⏳ DEFERRED
**Owner**: LLM Agent
**Estimated**: 30min
**Actual**: 
**Dependencies**: Phase 2 complete
**Priority**: LOW (optional cleanup)

**Description**: Analyze the 1 remaining lived mutation in tls_generator.go to determine if killable.

**Acceptance Criteria**:
- [ ] 3.1.1: Review gremlins output for tls_generator.go mutation
- [ ] 3.1.2: Analyze mutation type and location
- [ ] 3.1.3: Determine if killable with tests
- [ ] 3.1.4: Document findings in research.md

**Files**:
- docs/fixes-needed-plan-tasks-v4/research.md (update)

---

### Task 3.2: Determine Killability or Inherent Limitation

**Status**: ⏳ DEFERRED
**Owner**: LLM Agent
**Estimated**: 30min
**Actual**: 
**Dependencies**: Task 3.1 complete
**Priority**: LOW

**Description**: Make decision on whether mutation is killable or represents inherent testing limitation.

**Acceptance Criteria**:
- [ ] 3.2.1: Assess test implementation complexity
- [ ] 3.2.2: Assess efficacy gain (0.09% to reach 99%)
- [ ] 3.2.3: Document decision (killable vs inherent limitation)
- [ ] 3.2.4: Update mutation-analysis.md

**Files**:
- docs/gremlins/mutation-analysis.md (update)

---

### Task 3.3: Implement Test if Feasible

**Status**: ⏳ DEFERRED
**Owner**: LLM Agent
**Estimated**: 1h
**Actual**: 
**Dependencies**: Task 3.2 complete
**Priority**: LOW

**Description**: If mutation determined killable with reasonable effort, implement test.

**Acceptance Criteria**:
- [ ] 3.3.1: Implement test (if feasible)
- [ ] 3.3.2: Run gremlins verification
- [ ] 3.3.3: Verify efficacy improvement (98.91% → 99%+)
- [ ] 3.3.4: Update tasks.md and plan.md
- [ ] 3.3.5: Commit: "test(template): kill final mutation - 99%+ efficacy"

**Files**:
- internal/apps/template/service/config/*_test.go (add test)

---

## Phase 4: Continuous Mutation Testing

**Objective**: Enable automated mutation testing in CI/CD
**Estimated**: 2h total
**Status**: ⏳ NOT STARTED

### Task 4.1: Verify ci-mutation.yml Workflow

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Estimated**: 30min
**Actual**: 
**Dependencies**: Phase 2 complete (cipher-im unblocked)
**Priority**: HIGH

**Description**: Verify existing CI/CD mutation testing workflow is correctly configured.

**Acceptance Criteria**:
- [ ] 4.1.1: Review .github/workflows/ci-mutation.yml
- [ ] 4.1.2: Verify workflow triggers correctly
- [ ] 4.1.3: Verify artifact upload configured
- [ ] 4.1.4: Document any required changes
- [ ] 4.1.5: Commit if changes needed: "ci(mutation): verify workflow configuration"

**Files**:
- .github/workflows/ci-mutation.yml (verify)

---

[Additional tasks 4.2-7.35 follow similar pattern - truncated for brevity]

---

## Cross-Cutting Tasks

### Documentation
- [ ] Update README.md with mutation testing instructions
- [ ] Update DEV-SETUP.md with workflow setup
- [ ] Create research.md with comparison table
- [ ] Update completed.md as phases finish

### Testing
- [ ] All tests pass (`runTests`)
- [ ] Coverage ≥95% production, ≥98% infrastructure
- [ ] Mutation efficacy ≥98% ideal (ALL services)
- [ ] Race detector clean on Linux

### Quality
- [ ] Linting passes (`golangci-lint run`)
- [ ] No new TODOs without tracking
- [ ] Conventional commits enforced
