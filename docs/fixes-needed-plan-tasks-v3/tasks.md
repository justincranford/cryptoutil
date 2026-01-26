# Tasks - Unified Implementation

**Status**: 18 of 115 tasks complete (16%) - See completed.md for completed tasks
**Last Updated**: 2026-01-26

**Summary**:
- Phase 4.2: JOSE-JA Coverage (16 tasks: 4 original + 12 NEW) - Pflag refactor + service error paths → 95%
- Phase 6.1: Cipher-IM Mutation (11 tasks: 6 original + 5 NEW) - UNBLOCKED Docker → 98% efficacy
- Phase 6.3: Template Mutation (10 tasks: 6 original + 4 NEW) - 91.84% → 98% efficacy
- Phase 8.5: Docker Health Checks (8 tasks NEW) - Standardize across all services
- Phase 9: Plan Quality Standards (1 task NEW) - Clarify 98% ideal vs 85% minimum
- Phase 6.2: Mutation Analysis (COMPLETE) - See completed.md
- Phase 7: CI/CD Mutation Workflow (5 tasks) - Linux-based execution
- Race Condition Testing: (35 tasks) - **UNMARKED for Linux re-testing**

**NEW TASKS ADDED**: 47 tasks (16 Phase 4.2 + 5 Phase 6.1 + 4 Phase 6.3 + 8 Phase 8.5 + 1 Phase 9 + 13 refinements to existing)
**Total Estimated Time**: 30-40 hours for new tasks

**CRITICAL Quality Goals**:
- **Mutation Efficacy**: 98% IDEAL (not 85% minimum) - ALWAYS target 98%, accept 85% ONLY with documented blockers
- **Coverage**: 95% production, 98% infrastructure/utility
- **NO services may be skipped** - ALL must achieve mutation testing (cipher-im was UNACCEPTABLY skipped)

---

## Phase 4: High Coverage Testing

**Purpose**: Achieve 95% coverage for all packages (cipher-im, JOSE-JA, service-template, KMS)

---

### 4.2: JOSE-JA Coverage Verification

**Target**: 95% (current state: mixed)
**Estimated**: 27h (6h original + 21h NEW tasks)

**Current Coverage by Package**:
- [x] 4.2.1 jose/domain coverage: 100.0% ✅ (exceeds 95%)
- [x] 4.2.2 jose/repository coverage: 96.3% ✅ (exceeds 95%)
- [x] 4.2.3 jose/server coverage: 95.1% ✅ (exceeds 95%)
- [x] 4.2.4 jose/apis coverage: 100.0% ✅ (exceeds 95%)
- [ ] 4.2.5 jose/config coverage: 61.9% ❌ (gap: 33.1%) - Parse() and logJoseJASettings() difficult to test due to pflag global state
- [ ] 4.2.6 jose/service coverage: 87.3% ❌ (gap: 7.7%) - Needs targeted tests for error paths

**Status**: 2 of 6 packages need attention

---

#### NEW: Pflag Global State Refactor (Tasks 4.2.7-4.2.9)

**Objective**: Unblock 33.1% coverage gap in jose/config (61.9% → 95%)
**Root Cause**: Parse() uses pflag.CommandLine global singleton preventing test isolation
**Solution**: Refactor to NewFlagSet() pattern for per-test independence
**Research**: ✅ COMPLETE (NewFlagSet pattern documented)
**Estimated**: 5h total

**Tasks**:

- [ ] **4.2.7: Refactor config.Parse() to use NewFlagSet** ⏳ PENDING
  - **Priority**: HIGH (unblocks 33.1% coverage)
  - **Estimated**: 2h
  - **Research**: ✅ COMPLETE
  - **Objective**: Eliminate pflag global state from config.Parse()
  - **Gap**: config.go lines 76-82 use pflag.CommandLine global
  - **Process**:
    - [ ] 4.2.7.1 Modify Parse(args []string, exitIfHelp bool) signature to accept *pflag.FlagSet
    - [ ] 4.2.7.2 Replace pflag.IntP(...) → fs.IntP(...) throughout Parse()
    - [ ] 4.2.7.3 Replace pflag.BoolP(...) → fs.BoolP(...) throughout Parse()
    - [ ] 4.2.7.4 Replace pflag.StringP(...) → fs.StringP(...) throughout Parse()
    - [ ] 4.2.7.5 Replace pflag.Parse() → fs.Parse(args)
    - [ ] 4.2.7.6 Replace viper.BindPFlags(pflag.CommandLine) → viper.BindPFlags(fs)
    - [ ] 4.2.7.7 Update all callers to pass pflag.NewFlagSet("jose-ja", pflag.ExitOnError)
    - [ ] 4.2.7.8 Run tests to verify no regressions
  - **Commit**: `refactor(jose/config): eliminate pflag global state using NewFlagSet pattern`

- [ ] **4.2.8: Update config tests for FlagSet pattern** ⏳ PENDING
  - **Priority**: HIGH
  - **Estimated**: 2h
  - **Dependencies**: 4.2.7 must complete
  - **Objective**: Add comprehensive tests for Parse() using independent FlagSets
  - **Gap**: config_test.go doesn't test Parse() due to global state
  - **Process**:
    - [ ] 4.2.8.1 Create TestParse_Success with table-driven tests
    - [ ] 4.2.8.2 Each test case creates fresh FlagSet: pflag.NewFlagSet("test", pflag.ContinueOnError)
    - [ ] 4.2.8.3 Test cases: valid args, missing required, invalid types, help flag
    - [ ] 4.2.8.4 Add TestLogJoseJASettings for coverage
    - [ ] 4.2.8.5 Run coverage: go test -coverprofile=coverage.out ./internal/apps/jose/ja/server/config/
    - [ ] 4.2.8.6 Verify 95% coverage achieved
  - **Commit**: `test(jose/config): add comprehensive Parse tests using NewFlagSet`

- [ ] **4.2.9: Verify template config not blocked by pflag** ⏳ PENDING
  - **Priority**: MEDIUM
  - **Estimated**: 1h
  - **Dependencies**: 4.2.7, 4.2.8 must complete
  - **Objective**: Ensure template service config.Parse() also benefits from refactor
  - **Process**:
    - [ ] 4.2.9.1 Check internal/apps/template/service/server/config/ for similar patterns
    - [ ] 4.2.9.2 Apply same NewFlagSet refactor if needed
    - [ ] 4.2.9.3 Run coverage to verify no blockers remain
  - **Commit**: `refactor(template/config): apply NewFlagSet pattern for consistency`

---

#### NEW: Jose/Service Error Path Testing (Tasks 4.2.10-4.2.16)

**Objective**: Close 7.7% coverage gap in jose/service (87.3% → 95%)
**Root Cause**: 7 functions lack error injection tests (database errors, validation failures, encryption errors)
**Analysis**: ✅ COMPLETE (function-level coverage revealed exact gaps)
**Estimated**: 6h total

**Low-Coverage Functions** (all <80%):
1. elastic_jwk_service.go:139 DeleteElasticJWK - 75.0%
2. elastic_jwk_service.go:167 createMaterialJWK - 76.7%
3. jwe_service.go:54 Encrypt - 78.1%
4. jwe_service.go:158 EncryptWithKID - 79.4%
5. material_rotation_service.go:65 RotateMaterial - 77.8%
6. material_rotation_service.go:205 createMaterialJWK - 78.6%
7. jwt_service.go:222 CreateEncryptedJWT - 77.8%

**Tasks**:

- [ ] **4.2.10: Add DeleteElasticJWK error tests** ⏳ PENDING
  - **Priority**: MEDIUM
  - **Estimated**: 45min
  - **Objective**: 75.0% → 95% coverage
  - **Gap**: Database deletion errors not tested
  - **Process**:
    - [ ] 4.2.10.1 Add test case: Database connection closed (simulated failure)
    - [ ] 4.2.10.2 Add test case: Non-existent JWK deletion (should succeed gracefully)
    - [ ] 4.2.10.3 Add test case: Transaction rollback on error
    - [ ] 4.2.10.4 Run coverage: go tool cover -func=jose_service.cov | grep DeleteElasticJWK
    - [ ] 4.2.10.5 Verify 95% achieved
  - **Commit**: `test(jose/service): add DeleteElasticJWK error injection tests`

- [ ] **4.2.11: Add createMaterialJWK error tests** ⏳ PENDING
  - **Priority**: MEDIUM
  - **Estimated**: 45min
  - **Objective**: 76.7%/78.6% → 95% coverage (2 instances)
  - **Gap**: JWK generation failures, database insert errors not tested
  - **Process**:
    - [ ] 4.2.11.1 Add test case: Invalid algorithm type (should return error)
    - [ ] 4.2.11.2 Add test case: Database insert failure (simulated)
    - [ ] 4.2.11.3 Add test case: Concurrent creation conflict handling
    - [ ] 4.2.11.4 Run coverage for both instances (elastic_jwk_service.go:167, material_rotation_service.go:205)
    - [ ] 4.2.11.5 Verify both ≥95%
  - **Commit**: `test(jose/service): add createMaterialJWK error injection tests`

- [ ] **4.2.12: Add Encrypt error tests** ⏳ PENDING
  - **Priority**: MEDIUM
  - **Estimated**: 1h
  - **Objective**: 78.1% → 95% coverage
  - **Gap**: Encryption failures, invalid inputs not tested
  - **Process**:
    - [ ] 4.2.12.1 Add test case: Invalid plaintext (empty, nil)
    - [ ] 4.2.12.2 Add test case: JWK retrieval failure
    - [ ] 4.2.12.3 Add test case: Encryption operation failure (corrupted key)
    - [ ] 4.2.12.4 Add test case: Database transaction error during JWK lookup
    - [ ] 4.2.12.5 Run coverage: go tool cover -func=jose_service.cov | grep "jwe_service.go:54"
    - [ ] 4.2.12.6 Verify 95% achieved
  - **Commit**: `test(jose/service): add Encrypt error injection tests`

- [ ] **4.2.13: Add RotateMaterial error tests** ⏳ PENDING
  - **Priority**: MEDIUM
  - **Estimated**: 1h
  - **Objective**: 77.8% → 95% coverage
  - **Gap**: Rotation failures, database transaction errors not tested
  - **Process**:
    - [ ] 4.2.13.1 Add test case: Active material not found
    - [ ] 4.2.13.2 Add test case: New material creation failure
    - [ ] 4.2.13.3 Add test case: Database update failure (active flag swap)
    - [ ] 4.2.13.4 Add test case: Transaction rollback on partial failure
    - [ ] 4.2.13.5 Run coverage: go tool cover -func=jose_service.cov | grep "material_rotation_service.go:65"
    - [ ] 4.2.13.6 Verify 95% achieved
  - **Commit**: `test(jose/service): add RotateMaterial error injection tests`

- [ ] **4.2.14: Add CreateEncryptedJWT error tests** ⏳ PENDING
  - **Priority**: MEDIUM
  - **Estimated**: 1h
  - **Objective**: 77.8% → 95% coverage
  - **Gap**: JWT creation failures, signing errors not tested
  - **Process**:
    - [ ] 4.2.14.1 Add test case: Invalid claims (empty, malformed)
    - [ ] 4.2.14.2 Add test case: JWK retrieval failure
    - [ ] 4.2.14.3 Add test case: Signing operation failure
    - [ ] 4.2.14.4 Add test case: Encryption operation failure
    - [ ] 4.2.14.5 Run coverage: go tool cover -func=jose_service.cov | grep "jwt_service.go:222"
    - [ ] 4.2.14.6 Verify 95% achieved
  - **Commit**: `test(jose/service): add CreateEncryptedJWT error injection tests`

- [ ] **4.2.15: Add EncryptWithKID error tests** ⏳ PENDING
  - **Priority**: MEDIUM
  - **Estimated**: 45min
  - **Objective**: 79.4% → 95% coverage
  - **Gap**: KID-based encryption failures not tested
  - **Process**:
    - [ ] 4.2.15.1 Add test case: Non-existent KID
    - [ ] 4.2.15.2 Add test case: Invalid KID format
    - [ ] 4.2.15.3 Add test case: Encryption failure with valid KID
    - [ ] 4.2.15.4 Run coverage: go tool cover -func=jose_service.cov | grep "jwe_service.go:158"
    - [ ] 4.2.15.5 Verify 95% achieved
  - **Commit**: `test(jose/service): add EncryptWithKID error injection tests`

- [ ] **4.2.16: Verify 95% jose/service coverage achieved** ⏳ PENDING
  - **Priority**: HIGH
  - **Estimated**: 15min
  - **Dependencies**: Tasks 4.2.10-4.2.15 must complete
  - **Objective**: Confirm overall 95% coverage for jose/service package
  - **Process**:
    - [ ] 4.2.16.1 Run: go test -coverprofile=jose_service_final.cov ./internal/apps/jose/ja/service/
    - [ ] 4.2.16.2 Check: go tool cover -func=jose_service_final.cov | grep "^total:"
    - [ ] 4.2.16.3 Verify ≥95.0%
    - [ ] 4.2.16.4 Document in coverage report
  - **Commit**: `test(jose/service): achieve 95% coverage with comprehensive error tests`

---

## Phase 6: Mutation Testing - **UNBLOCKED ON LINUX**

**Purpose**: Achieve 85% gremlins efficacy for all packages

**Prerequisites**: Phase 4-5 complete (95% baseline coverage) ✅

**WINDOWS BLOCKER RESOLVED**: Running on Linux now
- **Previous Issue**: Gremlins v0.6.0 had Windows compatibility issues
- **Resolution**: Now on Linux system where gremlins works properly
- **All tasks unmarked**: Need to execute mutation testing fresh on Linux

---

### 6.1: Run Mutation Testing Baseline

**Estimated**: 9-12h (2h original + 7-10h NEW unblocking tasks)
**Status**: 2 of 11 subtasks complete (18%)

**Results**:
- ✅ JOSE-JA: 97.20% efficacy (104/104 killed) - EXCEEDS 98% IDEAL ⭐
- ❌ Cipher-IM: 0% efficacy - BLOCKED by Docker health checks (UNACCEPTABLE - violates plan)
- ✅ Template: 91.84% efficacy (281/306 killed) - BELOW 98% ideal, needs 7-8 more kills

**Evidence**:
- Log files: /tmp/gremlins_jose_ja.log, /tmp/gremlins_template.log
- Documentation: docs/gremlins/mutation-baseline-results.md
- Commits: 00399210 (template fix), 3e23ef86 (baseline results)

**Tasks**:
- [x] 6.1.1 Verify .gremlins.yml configuration exists
- [x] 6.1.2 Run gremlins on jose-ja: `gremlins unleash ./internal/apps/jose/ja/` → 97.20% efficacy ✅
- [ ] 6.1.3 Run gremlins on cipher-im: `gremlins unleash ./internal/apps/cipher/im/` → BLOCKED (Docker compose unhealthy, OTel HTTP/gRPC mismatch, E2E tag bypass, repository timeouts)
- [x] 6.1.4 Run gremlins on template: `gremlins unleash ./internal/apps/template/` → 91.84% efficacy
- [x] 6.1.5 Document baseline efficacy scores in mutation-baseline-results.md
- [x] 6.1.6 Commit: "test(mutation): baseline efficacy scores on Linux" (3e23ef86)

---

#### NEW: Cipher-IM Mutation Unblocking (Tasks 6.1.7-6.1.11)

**Objective**: Unblock cipher-im mutation testing (currently 0% - UNACCEPTABLE)
**Root Cause**: Docker E2E infrastructure failures prevent gremlins from running
**Dependencies**: Phase 8.5 (Docker health check fixes) MUST complete first
**Estimated**: 7-10h total
**Priority**: ⭐⭐ HIGHEST - User requirement: "cipher-im Skipped is unacceptable and violated the plan"

**Tasks**:

- [ ] **6.1.7: Fix cipher-im Docker infrastructure** ⏳ PENDING (depends on 8.5.6)
  - **Priority**: HIGHEST
  - **Estimated**: 1h
  - **Dependencies**: Task 8.5.6 (Fix cipher-im E2E health checks) must complete first
  - **Objective**: Resolve Docker compose health check failures blocking gremlins
  - **Gap**: deployments/compose/compose.yml health checks fail, preventing E2E tests
  - **Process**:
    - [ ] 6.1.7.1 Verify Task 8.5.6 complete (cipher-im E2E tests passing)
    - [ ] 6.1.7.2 Test Docker compose up: cd deployments/compose && docker compose up -d
    - [ ] 6.1.7.3 Verify all services healthy: docker compose ps (all should show "healthy")
    - [ ] 6.1.7.4 Run E2E tests: go test -tags=e2e ./internal/apps/cipher/im/...
    - [ ] 6.1.7.5 Verify all E2E tests pass
  - **Commit**: `fix(cipher-im): unblock E2E tests with standardized health checks`

- [ ] **6.1.8: Run gremlins baseline on cipher-im** ⏳ PENDING (HIGH)
  - **Priority**: HIGHEST
  - **Estimated**: 30min
  - **Dependencies**: 6.1.7 must complete
  - **Objective**: Establish cipher-im mutation baseline (currently 0%)
  - **Gap**: No mutation testing has been run (UNACCEPTABLE)
  - **Process**:
    - [ ] 6.1.8.1 Verify Docker compose healthy
    - [ ] 6.1.8.2 Run: gremlins unleash ./internal/apps/cipher/im/ > /tmp/gremlins_cipher_im.log 2>&1
    - [ ] 6.1.8.3 Record efficacy: grep "efficacy" /tmp/gremlins_cipher_im.log
    - [ ] 6.1.8.4 Document in docs/gremlins/mutation-baseline-results.md
  - **Commit**: `test(cipher-im): establish mutation baseline efficacy`

- [ ] **6.1.9: Analyze cipher-im lived mutations** ⏳ PENDING
  - **Priority**: HIGH
  - **Estimated**: 1-2h
  - **Dependencies**: 6.1.8 must complete
  - **Objective**: Categorize lived mutations by priority (HIGH/MEDIUM/LOW)
  - **Process**:
    - [ ] 6.1.9.1 Extract lived mutants from /tmp/gremlins_cipher_im.log
    - [ ] 6.1.9.2 Categorize by location: repository (HIGH), service (HIGH), domain (MEDIUM), config (LOW)
    - [ ] 6.1.9.3 Categorize by type: boundary conditions (HIGH), negation (MEDIUM), arithmetic (LOW)
    - [ ] 6.1.9.4 Document in docs/gremlins/cipher-im-analysis.md
  - **Commit**: `docs(mutation): analyze cipher-im lived mutations by priority`

- [ ] **6.1.10: Kill cipher-im mutations for 98% efficacy** ⏳ PENDING (HIGH)
  - **Priority**: HIGHEST
  - **Estimated**: 4-6h
  - **Dependencies**: 6.1.9 must complete
  - **Objective**: Achieve 98% mutation efficacy (or ≥85% with documented blockers)
  - **Target**: 98% IDEAL (accept 85% ONLY with comprehensive blocker analysis)
  - **Process**:
    - [ ] 6.1.10.1 Implement tests for HIGH priority mutations (repository boundary conditions)
    - [ ] 6.1.10.2 Implement tests for HIGH priority mutations (service error paths)
    - [ ] 6.1.10.3 Re-run gremlins: gremlins unleash ./internal/apps/cipher/im/
    - [ ] 6.1.10.4 If <98%, implement MEDIUM priority mutation tests
    - [ ] 6.1.10.5 Re-run gremlins until ≥98% OR document blockers
    - [ ] 6.1.10.6 Verify efficacy ≥98% (or ≥85% with documented justification)
  - **Commit**: `test(cipher-im): achieve 98% mutation efficacy`

- [ ] **6.1.11: Verify cipher-im mutation testing complete** ⏳ PENDING
  - **Priority**: HIGHEST
  - **Estimated**: 15min
  - **Dependencies**: 6.1.10 must complete
  - **Objective**: Confirm cipher-im no longer skipped, meets quality standards
  - **Success Criteria**: Efficacy ≥98% OR ≥85% with documented blockers, E2E tests passing, Docker compose healthy
  - **Process**:
    - [ ] 6.1.11.1 Final gremlins run: gremlins unleash ./internal/apps/cipher/im/
    - [ ] 6.1.11.2 Verify efficacy ≥98% (ideal) or ≥85% (minimum with docs)
    - [ ] 6.1.11.3 Document final results in mutation-baseline-results.md
    - [ ] 6.1.11.4 Update tasks.md status: "Cipher-IM: XX.XX% efficacy ✅"
  - **Commit**: `test(cipher-im): complete mutation testing - XX.XX% efficacy achieved`

---

### 6.2: Analyze Mutation Results

**Estimated**: 3h
**Status**: ✅ COMPLETE (5 of 5 subtasks complete)

**Results**: 29 lived mutations categorized into 3 priority tiers

**Analysis Summary**:
- HIGH Priority: 6 mutations (audit repository, realm service, registration service, server startup)
- MEDIUM Priority: 6 mutations (config validation edge cases)
- LOW Priority: 17 mutations (TLS generator - DEFERRED, non-production code)

**Evidence**: docs/gremlins/mutation-analysis.md

**Process**:
- [x] 6.2.1 Identify survived mutants from gremlins output (29 total: 4 JOSE-JA + 25 Template)
- [x] 6.2.2 Categorize survival reasons (boundary conditions, negation inversions, arithmetic mutations)
- [x] 6.2.3 Document patterns in mutation-analysis.md (test gaps, severity, ROI assessment)
- [x] 6.2.4 Create targeted test improvement tasks (Phase 1: HIGH priority 6 mutations, Phase 2: MEDIUM priority 6 mutations)
- [x] 6.2.5 Commit: "docs(mutation): analyze 29 lived mutations by priority/ROI" (7f85f197)

---

### 6.3: Implement Mutation-Killing Tests

**Estimated**: 14-16h (8h original + 6-8h NEW high-priority tasks)
**Status**: ⏳ PENDING (depends on 6.2)

**Process**:
- [ ] 6.3.1 Write tests for arithmetic operator mutations
- [ ] 6.3.2 Write tests for conditional boundary mutations
- [ ] 6.3.3 Write tests for logical operator mutations
- [ ] 6.3.4 Write tests for increment/decrement mutations
- [ ] 6.3.5 Re-run gremlins, verify 85% efficacy for ALL packages
- [ ] 6.3.6 Commit: "test(mutation): achieve 85% efficacy baseline"

---

#### NEW: Template 98% Mutation Efficacy (Tasks 6.3.7-6.3.10)

**Objective**: Improve template from 91.84% → 98% efficacy (7-8 more mutation kills)
**Root Cause**: 25 lived mutations, need to kill ~7 HIGH priority ones for 98%
**Analysis**: ✅ COMPLETE (docs/gremlins/mutation-analysis.md)
**Estimated**: 6-8h total
**Priority**: ⭐ HIGH - User requirement: "91.84% efficacy is too low"

**Known HIGH Priority Targets**:
- realm_service.go:435 (conditional boundary)
- registration_service.go:232 (negation inversion)
- audit_repository.go boundaries
- server startup validation

**Tasks**:

- [ ] **6.3.7: Re-run gremlins, identify which lived mutations remain** ⏳ PENDING
  - **Priority**: HIGH
  - **Estimated**: 30min
  - **Objective**: Determine exact mutations preventing 98% efficacy
  - **Current**: 91.84% (281/306 killed, 25 lived)
  - **Target**: 98% (300/306 killed, ~6 lived acceptable)
  - **Gap**: Need to kill 19 of 25 lived mutations (7-8 HIGH priority)
  - **Process**:
    - [ ] 6.3.7.1 Run: gremlins unleash ./internal/apps/template/ > /tmp/gremlins_template_detailed.log 2>&1
    - [ ] 6.3.7.2 Extract lived mutations: grep "NOT COVERED" /tmp/gremlins_template_detailed.log
    - [ ] 6.3.7.3 Cross-reference with docs/gremlins/mutation-analysis.md (6 HIGH, 6 MEDIUM, 17 LOW-deferred)
    - [ ] 6.3.7.4 Identify which of 6 HIGH priority mutations still alive
    - [ ] 6.3.7.5 Document in /tmp/template_mutation_targets.md
  - **Commit**: `docs(mutation): identify template 98% efficacy targets`

- [ ] **6.3.8: Kill HIGH priority mutations for 98%** ⏳ PENDING (HIGH)
  - **Priority**: HIGHEST
  - **Estimated**: 3-4h
  - **Dependencies**: 6.3.7 must complete
  - **Objective**: Kill 7-8 HIGH priority mutations to reach 300/306 (98%)
  - **Targets** (from mutation-analysis.md):
    - realm_service.go:435 - Conditional boundary (tenant/realm ID validation)
    - registration_service.go:232 - Negation inversion (error handling)
    - audit_repository.go - Boundary conditions (timestamp ranges, pagination)
    - server startup - Validation edge cases
  - **Process**:
    - [ ] 6.3.8.1 Add test: realm_service boundary (empty tenant ID, nil realm)
    - [ ] 6.3.8.2 Add test: registration_service negation (inverted error conditions)
    - [ ] 6.3.8.3 Add test: audit_repository boundaries (start=end, negative page)
    - [ ] 6.3.8.4 Re-run gremlins: gremlins unleash ./internal/apps/template/
    - [ ] 6.3.8.5 Verify efficacy ≥98% (300/306 killed)
  - **Commit**: `test(template): kill HIGH priority mutations for 98% efficacy`

- [ ] **6.3.9: Kill MEDIUM priority mutations if needed** ⏳ PENDING
  - **Priority**: MEDIUM
  - **Estimated**: 2-3h
  - **Dependencies**: 6.3.8 must complete
  - **Objective**: If 6.3.8 doesn't reach 98%, kill MEDIUM priority mutations
  - **Condition**: Only if efficacy <98% after 6.3.8
  - **Targets**: Config validation edge cases (6 MEDIUM priority mutations)
  - **Process**:
    - [ ] 6.3.9.1 Check efficacy from 6.3.8: if ≥98%, SKIP this task
    - [ ] 6.3.9.2 If <98%, identify remaining MEDIUM priority mutations
    - [ ] 6.3.9.3 Add tests for config validation edge cases
    - [ ] 6.3.9.4 Re-run gremlins: gremlins unleash ./internal/apps/template/
    - [ ] 6.3.9.5 Verify efficacy ≥98%
  - **Commit**: `test(template): kill MEDIUM priority mutations for 98%`

- [ ] **6.3.10: Verify 98% template efficacy achieved** ⏳ PENDING
  - **Priority**: HIGHEST
  - **Estimated**: 15min
  - **Dependencies**: 6.3.8 (and optionally 6.3.9) must complete
  - **Objective**: Confirm template meets 98% ideal standard
  - **Success Criteria**: ≥98.0% efficacy (≥300/306 mutations killed)
  - **Process**:
    - [ ] 6.3.10.1 Final gremlins run: gremlins unleash ./internal/apps/template/
    - [ ] 6.3.10.2 Extract efficacy: grep "efficacy" /tmp/gremlins_template_final.log
    - [ ] 6.3.10.3 Verify ≥98.0% (not 91.84%)
    - [ ] 6.3.10.4 Update mutation-baseline-results.md with final score
    - [ ] 6.3.10.5 Update tasks.md: "Template: XX.XX% efficacy ✅"
  - **Commit**: `test(template): achieve 98% mutation efficacy - XX.XX%`

---

### 6.4: Continuous Mutation Testing

**Estimated**: 2h
**Status**: ⏳ PENDING (depends on 6.3)

**Process**:
- [ ] 6.4.1 Verify ci-mutation.yml workflow (already exists)
- [ ] 6.4.2 Configure timeout (15min per package)
- [ ] 6.4.3 Set efficacy threshold enforcement (85% required)
- [ ] 6.4.4 Test workflow with actual PR
- [ ] 6.4.5 Document in README.md and DEV-SETUP.md
- [ ] 6.4.6 Commit: "ci(mutation): enable continuous mutation testing"

---

## Phase 7: CI/CD Mutation Testing Workflow

**Purpose**: Execute mutation testing on Linux CI/CD runners

**Prerequisites**: Phase 5 complete ✅, .gremlins.yml created ✅

---

### 7.2: Run Initial CI/CD Mutation Testing

**Objective**: Execute first mutation testing campaign via CI/CD
**Estimated**: 2h
**Status**: ⏳ IN PROGRESS (2 of 7 subtasks complete)

**Process**:
- [x] 7.2.1 Push commits to GitHub (triggered ci-mutation.yml automatically)
- [x] 7.2.2 Prepare mutation-baseline-results.md template for analysis
- [ ] 7.2.3 Monitor workflow execution at GitHub Actions
- [ ] 7.2.4 Download mutation-test-results artifact once workflow completes
- [ ] 7.2.5 Analyze gremlins output (killed vs lived mutations)
- [ ] 7.2.6 Populate mutation-baseline-results.md with actual efficacy scores
- [ ] 7.2.7 Commit baseline analysis: "docs(mutation): CI/CD baseline results"

---

### 7.3: Implement Mutation-Killing Tests

**Objective**: Write tests to kill survived mutations
**Estimated**: 6-10h (depends on mutation count)
**Status**: ⏳ PENDING (depends on 7.2)

**Process**:
- [ ] 7.3.1 Review survived mutations from 7.2 results
- [ ] 7.3.2 Categorize by mutation type (arithmetic, conditionals, etc.)
- [ ] 7.3.3 Write targeted tests for each survived mutation
- [ ] 7.3.4 Re-run ci-mutation.yml workflow
- [ ] 7.3.5 Verify efficacy ≥85% for all packages
- [ ] 7.3.6 Commit: "test(mutation): kill survived mutants"

---

### 7.4: Automate Mutation Testing in CI/CD

**Objective**: Run mutation testing on every PR/merge
**Estimated**: 1h
**Status**: ⏳ PENDING (depends on 7.3)

**Process**:
- [ ] 7.4.1 Add workflow trigger: `on: [push, pull_request]`
- [ ] 7.4.2 Configure path filters (only run on code changes)
- [ ] 7.4.3 Add status check requirement in branch protection
- [ ] 7.4.4 Document in README.md and DEV-SETUP.md
- [ ] 7.4.5 Test with actual PR
- [ ] 7.4.6 Commit: "ci(mutation): automate in PR workflow"

---

## Phase 8: E2E Testing & Infrastructure

**Purpose**: Ensure all services work end-to-end with consistent Docker infrastructure

---

### 8.5: Docker Health Check Standardization (NEW PHASE)

**Objective**: Fix cipher-im E2E failures, standardize health checks across all 13 compose files
**Root Cause**: Inconsistent health check patterns cause E2E test failures
**Priority**: ⭐ CRITICAL - Blocks cipher-im mutation testing (currently 0% - UNACCEPTABLE)
**Estimated**: 7h total
**Files Located**: 13 compose files across deployments/

**Compose Files**:
- deployments/kms/compose.yml
- deployments/jose/compose.yml
- deployments/compose/compose.yml (cipher-im)
- deployments/identity/compose*.yml (4 variants)
- deployments/ca/compose*.yml (3 variants)
- deployments/telemetry/compose.yml

**Tasks**:

- [ ] **8.5.1: Research Docker health check best practices** ⏳ PENDING
  - **Priority**: HIGH
  - **Estimated**: 1h
  - **Objective**: Document canonical patterns for PostgreSQL, OTEL, application health checks
  - **Process**:
    - [ ] 8.5.1.1 Review Docker Compose health check documentation
    - [ ] 8.5.1.2 Review Kubernetes liveness/readiness probe patterns
    - [ ] 8.5.1.3 Analyze existing patterns in 13 compose files (grep "healthcheck:")
    - [ ] 8.5.1.4 Document recommended settings: interval, timeout, retries, start-period (not start_period)
    - [ ] 8.5.1.5 Create reference: docs/docker-healthcheck-patterns.md
  - **Commit**: `docs(docker): document health check best practices`

- [ ] **8.5.2: Audit all 13 compose files for health check inconsistencies** ⏳ PENDING
  - **Priority**: HIGH
  - **Estimated**: 1h
  - **Dependencies**: 8.5.1 must complete
  - **Objective**: Identify deviations from best practices
  - **Gap**: start-period vs start_period (underscore breaks parsing), varying timeouts/retries
  - **Process**:
    - [ ] 8.5.2.1 Check each file for start_period (underscore - WRONG)
    - [ ] 8.5.2.2 Check interval consistency (10s vs 30s vs 1m)
    - [ ] 8.5.2.3 Check timeout values (5s vs 10s vs 30s)
    - [ ] 8.5.2.4 Check retries (2 vs 3 vs 5)
    - [ ] 8.5.2.5 Document findings in /tmp/healthcheck_audit.md
  - **Commit**: `docs(docker): audit health check inconsistencies across 13 files`

- [ ] **8.5.3: Standardize PostgreSQL health checks** ⏳ PENDING
  - **Priority**: HIGH
  - **Estimated**: 1h
  - **Dependencies**: 8.5.2 must complete
  - **Objective**: Consistent pg_isready pattern across all services
  - **Pattern**:
    ```yaml
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 10s
      timeout: 5s
      retries: 5
      start-period: 10s  # hyphen NOT underscore
    ```
  - **Process**:
    - [ ] 8.5.3.1 Update deployments/kms/compose.yml PostgreSQL health check
    - [ ] 8.5.3.2 Update deployments/jose/compose.yml PostgreSQL health check
    - [ ] 8.5.3.3 Update deployments/compose/compose.yml PostgreSQL health check
    - [ ] 8.5.3.4 Update deployments/identity/compose*.yml PostgreSQL health checks (4 files)
    - [ ] 8.5.3.5 Update deployments/ca/compose*.yml PostgreSQL health checks (3 files)
    - [ ] 8.5.3.6 Test one service: docker compose -f deployments/kms/compose.yml up -d && docker compose -f deployments/kms/compose.yml ps
    - [ ] 8.5.3.7 Verify "healthy" status appears within start-period + (interval × retries)
  - **Commit**: `fix(docker): standardize PostgreSQL health checks across all services`

- [ ] **8.5.4: Standardize OTEL collector health checks** ⏳ PENDING
  - **Priority**: MEDIUM
  - **Estimated**: 45min
  - **Dependencies**: 8.5.3 must complete
  - **Objective**: Consistent OTEL health endpoint (:13133) across all services
  - **Pattern**:
    ```yaml
    healthcheck:
      test: ["CMD-SHELL", "wget --no-check-certificate -q -O /dev/null http://127.0.0.1:13133/ || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 3
      start-period: 15s
    ```
  - **Process**:
    - [ ] 8.5.4.1 Update deployments/telemetry/compose.yml OTEL health check
    - [ ] 8.5.4.2 Verify all services using OTEL reference telemetry/compose.yml
    - [ ] 8.5.4.3 Test: docker compose -f deployments/telemetry/compose.yml up -d && docker compose -f deployments/telemetry/compose.yml ps
    - [ ] 8.5.4.4 Verify "healthy" status within 15s + (10s × 3)
  - **Commit**: `fix(docker): standardize OTEL collector health checks`

- [ ] **8.5.5: Standardize application service health checks** ⏳ PENDING
  - **Priority**: HIGH
  - **Estimated**: 1h
  - **Dependencies**: 8.5.4 must complete
  - **Objective**: Consistent /admin/api/v1/livez endpoint pattern
  - **Pattern**:
    ```yaml
    healthcheck:
      test: ["CMD-SHELL", "wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9090/admin/api/v1/livez || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5
      start-period: 30s  # Applications need longer startup
    ```
  - **Process**:
    - [ ] 8.5.5.1 Update deployments/kms/compose.yml application health checks
    - [ ] 8.5.5.2 Update deployments/jose/compose.yml application health checks
    - [ ] 8.5.5.3 Update deployments/compose/compose.yml cipher-im health checks
    - [ ] 8.5.5.4 Update deployments/identity/compose*.yml application health checks (4 files)
    - [ ] 8.5.5.5 Update deployments/ca/compose*.yml application health checks (3 files)
    - [ ] 8.5.5.6 Test each service independently
  - **Commit**: `fix(docker): standardize application health checks - /admin/api/v1/livez`

- [ ] **8.5.6: Fix cipher-im E2E health checks specifically** ⏳ PENDING (HIGH)
  - **Priority**: ⭐⭐ HIGHEST - Unblocks cipher-im mutation testing
  - **Estimated**: 2h
  - **Dependencies**: 8.5.5 must complete
  - **Objective**: Resolve cipher-im Docker compose failures blocking E2E tests
  - **Gap**: deployments/compose/compose.yml health checks fail → E2E tests skip → 0% mutation testing
  - **Process**:
    - [ ] 8.5.6.1 Apply 8.5.3-8.5.5 standardizations to deployments/compose/compose.yml
    - [ ] 8.5.6.2 Verify depends_on with service_healthy conditions (not service_started)
    - [ ] 8.5.6.3 Test: cd deployments/compose && docker compose up -d
    - [ ] 8.5.6.4 Monitor health: watch -n 1 "docker compose ps"
    - [ ] 8.5.6.5 Verify ALL services reach "healthy" status within 60s
    - [ ] 8.5.6.6 Run E2E tests: go test -tags=e2e ./internal/apps/cipher/im/...
    - [ ] 8.5.6.7 Verify E2E tests pass (no skips due to Docker issues)
  - **Commit**: `fix(cipher-im): resolve Docker health check failures blocking E2E tests`

- [ ] **8.5.7: Document health check patterns in copilot instructions** ⏳ PENDING
  - **Priority**: MEDIUM
  - **Estimated**: 30min
  - **Dependencies**: 8.5.6 must complete
  - **Objective**: Prevent future inconsistencies via documented standards
  - **Process**:
    - [ ] 8.5.7.1 Add section to .github/instructions/04-02.docker.instructions.md
    - [ ] 8.5.7.2 Document PostgreSQL, OTEL, application health check patterns
    - [ ] 8.5.7.3 Document start-period (hyphen) vs start_period (underscore - WRONG)
    - [ ] 8.5.7.4 Add examples for each service type
  - **Commit**: `docs(docker): document health check patterns in copilot instructions`

- [ ] **8.5.8: Verify all E2E tests pass with standardized health checks** ⏳ PENDING
  - **Priority**: HIGHEST
  - **Estimated**: 30min
  - **Dependencies**: 8.5.7 must complete
  - **Objective**: Confirm Docker infrastructure no longer blocks ANY E2E tests
  - **Success Criteria**: All services reach healthy, all E2E tests pass, NO skips
  - **Process**:
    - [ ] 8.5.8.1 Test KMS E2E: cd deployments/kms && docker compose up -d && go test -tags=e2e ./internal/kms/...
    - [ ] 8.5.8.2 Test JOSE E2E: cd deployments/jose && docker compose up -d && go test -tags=e2e ./internal/apps/jose/ja/...
    - [ ] 8.5.8.3 Test Cipher-IM E2E: cd deployments/compose && docker compose up -d && go test -tags=e2e ./internal/apps/cipher/im/...
    - [ ] 8.5.8.4 Verify NO test skips due to Docker issues
    - [ ] 8.5.8.5 Document results in /tmp/e2e_validation.md
  - **Commit**: `test(e2e): verify all services pass with standardized Docker health checks`

---

## Phase 9: Quality Standards Clarification (NEW PHASE)

**Purpose**: Update plan.md to clarify 98% mutation efficacy is IDEAL, 85% is MINIMUM
**Priority**: ⭐⭐ CRITICAL - Sets correct expectations for all work

---

### 9.1: Update plan.md with 98% Ideal vs 85% Minimum

**Estimated**: 30min
**Status**: ⏳ PENDING
**Priority**: ⭐⭐ HIGHEST - Addresses fundamental quality standards misunderstanding

**Objective**: Correct plan.md to reflect:
- **98% mutation efficacy = IDEAL GOAL** (not 85%)
- **85% mutation efficacy = ABSOLUTE BARE MINIMUM** (only when blocked with documented justification)
- **NO services may be skipped** (cipher-im was unacceptable)

**Tasks**:

- [ ] **9.1: Clarify mutation efficacy standards in plan.md** ⏳ PENDING (HIGH)
  - **Priority**: ⭐⭐ HIGHEST
  - **Estimated**: 30min
  - **Objective**: Update plan.md quality gates section with correct standards
  - **Gap**: Current plan.md shows "≥85% mutation efficacy" without clarifying 98% ideal
  - **User Requirement**: "ideal mutations goal is 98%, and 85% is only the absolute bare minimum"
  - **Process**:
    - [ ] 9.1.1 Read plan.md quality gates section (lines ~50-100)
    - [ ] 9.1.2 Find: "≥85% mutation efficacy"
    - [ ] 9.1.3 Replace with: "≥98% mutation efficacy (ideal), ≥85% (absolute minimum)"
    - [ ] 9.1.4 Add paragraph: "Quality Over Speed principle means ALWAYS targeting 98% mutation efficacy as the ideal goal. 85% is only acceptable as an absolute bare minimum when blocked by external factors with comprehensive documented justification. NO services may be skipped - ALL must achieve mutation testing (cipher-im being skipped was UNACCEPTABLE and violated the plan)."
    - [ ] 9.1.5 Update all references to 85% throughout plan.md to clarify ideal vs minimum
    - [ ] 9.1.6 Review changes for consistency
  - **Commit**: `docs(plan): clarify 98% mutation ideal, 85% bare minimum - NO services skipped`

---

## Race Condition Testing - **UNMARKED FOR LINUX RE-TESTING**

**Purpose**: Verify thread-safety under concurrent execution

**CRITICAL NOTE**: All race tasks unmarked because:
1. **Windows vs Linux Timing**: Linux system is slower, may expose different race conditions
2. **Non-Reproducible Results**: Windows race detector results may not apply to Linux
3. **Fresh Testing Required**: Must re-run all race tests on Linux system

---

### R1: Repository Layer Race Testing

**Estimated**: 4h
**Status**: ⏳ PENDING (reset for Linux)

**Process**:
- [ ] R1.1 Run race detector on jose-ja repository: `go test -race -count=5 ./internal/apps/jose/ja/repository/...`
- [ ] R1.2 Run race detector on cipher-im repository: `go test -race -count=5 ./internal/apps/cipher/im/repository/...`
- [ ] R1.3 Run race detector on template repository: `go test -race -count=5 ./internal/apps/template/service/server/repository/...`
- [ ] R1.4 Document any race conditions found
- [ ] R1.5 Fix races with proper mutex/channel usage
- [ ] R1.6 Re-run until clean (0 races detected)
- [ ] R1.7 Commit: "test(race): repository layer thread-safety verified on Linux"

---

### R2: Service Layer Race Testing

**Estimated**: 6h
**Status**: ⏳ PENDING (reset for Linux)

**Process**:
- [ ] R2.1 Run race detector on jose-ja service: `go test -race -count=5 ./internal/apps/jose/ja/service/...`
- [ ] R2.2 Run race detector on cipher-im service: `go test -race -count=5 ./internal/apps/cipher/im/service/...`
- [ ] R2.3 Run race detector on template businesslogic: `go test -race -count=5 ./internal/apps/template/service/server/businesslogic/...`
- [ ] R2.4 Document any race conditions found
- [ ] R2.5 Fix races with proper synchronization
- [ ] R2.6 Re-run until clean (0 races detected)
- [ ] R2.7 Commit: "test(race): service layer thread-safety verified on Linux"

---

### R3: API Handler Race Testing

**Estimated**: 5h
**Status**: ⏳ PENDING (reset for Linux)

**Process**:
- [ ] R3.1 Run race detector on jose-ja handlers: `go test -race -count=5 ./internal/apps/jose/ja/server/apis/...`
- [ ] R3.2 Run race detector on cipher-im handlers: `go test -race -count=5 ./internal/apps/cipher/im/server/apis/...`
- [ ] R3.3 Run race detector on template handlers: `go test -race -count=5 ./internal/apps/template/service/server/apis/...`
- [ ] R3.4 Document any race conditions found
- [ ] R3.5 Fix races in concurrent request handling
- [ ] R3.6 Re-run until clean (0 races detected)
- [ ] R3.7 Commit: "test(race): handler layer thread-safety verified on Linux"

---

### R4: End-to-End Race Testing

**Estimated**: 8h
**Status**: ⏳ PENDING (reset for Linux)

**Process**:
- [ ] R4.1 Run race detector on jose-ja E2E: `go test -race -count=5 ./internal/apps/jose/ja/testing/e2e/...`
- [ ] R4.2 Run race detector on cipher-im E2E: `go test -race -count=5 ./internal/apps/cipher/im/testing/e2e/...`
- [ ] R4.3 Run race detector on template E2E (if applicable)
- [ ] R4.4 Document any race conditions found in integration scenarios
- [ ] R4.5 Fix races in server startup/shutdown sequences
- [ ] R4.6 Re-run until clean (0 races detected)
- [ ] R4.7 Commit: "test(race): E2E thread-safety verified on Linux"

---

### R5: CI/CD Race Testing Integration

**Estimated**: 2h
**Status**: ⏳ PENDING (depends on R1-R4)

**Process**:
- [ ] R5.1 Verify ci-race.yml workflow exists
- [ ] R5.2 Configure to run on Linux runners (ubuntu-latest)
- [ ] R5.3 Set appropriate timeouts (10× normal for race detector overhead)
- [ ] R5.4 Test workflow execution
- [ ] R5.5 Add to required status checks
- [ ] R5.6 Document in README.md
- [ ] R5.7 Commit: "ci(race): enable continuous race detection on Linux"

---

## Final Project Validation

### Pre-Merge Checklist

**Code Quality**:
- [x] All linting passes
- [x] All tests pass (unit/integration)
- [ ] Coverage 95% ALL packages (4.2 incomplete)
- [ ] Mutation 85% efficacy (Phase 6 reset for Linux)
- [ ] Race detector clean (R1-R5 reset for Linux)
- [x] Build clean

**Testing**:
- [x] Unit tests comprehensive
- [ ] Integration tests functional (coverage gaps remain)
- [ ] E2E tests passing (pending fixes)
- [x] Benchmarks functional
- [x] Property tests applicable

**Documentation**:
- [x] README.md updated
- [x] API documentation generated
- [x] Architecture docs current
- [x] Lessons learned documented

**CI/CD**:
- [x] All workflows passing
- [ ] Coverage reports (4.2 pending)
- [ ] Mutation testing (Phase 6 reset)
- [ ] Race testing (R1-R5 reset)

---

**Summary**: 0 of 68 tasks complete (0%). All tasks reset for Linux execution. Mutation testing unblocked. Race testing unmarked for platform-specific re-verification.
