# Tasks - Remaining Work (V4)

**Status**: 31 of 115 tasks remaining (27.0% incomplete) - Phases 0.1, 0.2, 0.4, 0.5
**Last Updated**: 2026-01-28
**Priority Order**: Violation Remediation (Phase 0.1) → Documentation Updates (Phase 0.2) → KMS Modernization (Phase 0.4) → Template Mutation (Phase 0.5, DEFERRED)

**Completed Work**: See completed.md (92 of 115 tasks, 80.0%) - Phases 1, 1.5, 2, 3, 4, 5, 7, 0.3 COMPLETE

**User Feedback Resolution**: Task 0.3.1 COMPLETE ✅ (commit 234372f5) - Global mutation target fix: 20 replacements across 7 files correcting 85% → >=95% minimum (98% ideal)

**Note**: Phases 0.1, 0.2, 0.3 added to remediate violations identified in V4 completed work. Phase numbers adjusted: Former Phase 6 → Phase 0.4, Former Phase 8 → Phase 0.5. Phase 0.3 now COMPLETE.

## Phase 0.1: Violation Remediation (HIGHEST PRIORITY)

**Objective**: Fix 5 violation categories identified in V4 completed work
**Status**: ⏳ NOT STARTED
**Dependencies**: None - highest priority remediation
**Est. LOE**: 9-15 hours (from violation analysis)

**Reference**: docs/fixes-needed-plan-tasks-v4/VIOLATION-ANALYSIS.md

### Task 0.1.1: Refactor Standalone Tests to Table-Driven Pattern

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: None
**Priority**: HIGH
**Est. LOE**: 4-6 hours

**Description**: Refactor all standalone test variants into table-driven tests following copilot instructions pattern.

**Affected Files**:
- internal/apps/template/service/config/config_validation_test.go (TestValidateConfiguration_* functions)
- internal/apps/template/service/server/businesslogic/session_manager_jws_test.go (TestSessionManager_ValidateBrowserSession_JWS_* functions)
- internal/apps/template/service/server/businesslogic/session_manager_jwe_test.go (TestSessionManager_ValidateBrowserSession_JWE_* functions)

**Acceptance Criteria**:
- [ ] 0.1.1.1: Refactor config_validation_test.go to use `tests := []struct{name, input, wantErr}` pattern
- [ ] 0.1.1.2: Refactor session_manager_jws_test.go to use table-driven pattern
- [ ] 0.1.1.3: Refactor session_manager_jwe_test.go to use table-driven pattern
- [ ] 0.1.1.4: Verify all tests pass after refactoring
- [ ] 0.1.1.5: Commit: "refactor(template): convert standalone test variants to table-driven pattern"

**Files**:
- internal/apps/template/service/config/config_validation_test.go (refactored)
- internal/apps/template/service/server/businesslogic/session_manager_jws_test.go (refactored)
- internal/apps/template/service/server/businesslogic/session_manager_jwe_test.go (refactored)

---

### Task 0.1.2: Refactor Sad Path Tests to Table-Driven Pattern

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: None
**Priority**: MEDIUM
**Est. LOE**: 2-3 hours

**Description**: Consolidate error path tests for service initialization into table-driven sad path tests.

**Affected Files**:
- internal/apps/template/service/server/application/application_test.go (Test..._Error functions)
- internal/apps/template/service/server/builder/server_builder_test.go (error path tests for Build)

**Acceptance Criteria**:
- [ ] 0.1.2.1: Refactor application_test.go to use `tests := []struct{name, setup, wantErr}` pattern
- [ ] 0.1.2.2: Refactor server_builder_test.go to use table-driven error path pattern
- [ ] 0.1.2.3: Verify all tests pass after refactoring
- [ ] 0.1.2.4: Commit: "refactor(template): consolidate sad path tests into table-driven pattern"

**Files**:
- internal/apps/template/service/server/application/application_test.go (refactored)
- internal/apps/template/service/server/builder/server_builder_test.go (refactored)

---

### Task 0.1.3: Replace Real HTTPS Listeners with app.Test() Pattern

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: None
**Priority**: CRITICAL
**Est. LOE**: 1-2 hours

**Description**: Delete TestDualServers_* functions that start real HTTPS servers, rewrite using Fiber's app.Test() for in-memory handler testing.

**Affected Files**:
- internal/apps/template/service/server/listener/servers_test.go

**Acceptance Criteria**:
- [ ] 0.1.3.1: Delete all TestDualServers_* functions that start real servers
- [ ] 0.1.3.2: Rewrite using app.Test() pattern with in-memory HTTP requests
- [ ] 0.1.3.3: Verify execution time <1ms per test (vs >1s for real server tests)
- [ ] 0.1.3.4: Verify no Windows Firewall prompts triggered
- [ ] 0.1.3.5: Commit: "fix(template): replace real HTTPS listeners with app.Test() pattern (CRITICAL anti-pattern violation)"

**Files**:
- internal/apps/template/service/server/listener/servers_test.go (rewritten)

---

### Task 0.1.4: Remove Dead Code

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: None
**Priority**: MEDIUM
**Est. LOE**: 1-2 hours

**Description**: Remove functions/methods with 0% coverage that are never called in production or tests.

**Affected Code**:
- PublicServer.PublicBaseURL method (0% coverage, never called)
- UnsealKeysServiceFromSettings struct and wrapper methods (EncryptKey, DecryptKey, Shutdown) (0% coverage, never instantiated)
- EnsureSignatureAlgorithmType function (23.1% coverage, not used in production)

**Acceptance Criteria**:
- [ ] 0.1.4.1: Remove PublicServer.PublicBaseURL method
- [ ] 0.1.4.2: Remove UnsealKeysServiceFromSettings struct and all wrapper methods
- [ ] 0.1.4.3: Either remove EnsureSignatureAlgorithmType or document future extensibility justification
- [ ] 0.1.4.4: Verify all tests pass after removals
- [ ] 0.1.4.5: Verify go build ./... succeeds
- [ ] 0.1.4.6: Commit: "refactor(template): remove dead code (0% coverage, never called)"

**Files**:
- internal/apps/cipher/im/server/public_server.go (PublicBaseURL removed)
- internal/shared/barrier/unseal_keys_service.go (UnsealKeysServiceFromSettings removed)
- internal/jose/service/signature.go (EnsureSignatureAlgorithmType - remove or document)

---

### Task 0.1.5: Restore t.Parallel() in Config Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: None
**Priority**: MEDIUM
**Est. LOE**: 1-2 hours

**Description**: Refactor config tests to create isolated viper instances per test, then restore t.Parallel().

**Affected Files**:
- internal/apps/template/service/config/config_validation_test.go

**Acceptance Criteria**:
- [ ] 0.1.5.1: Refactor tests to use viper.New() for isolated instances instead of global viper
- [ ] 0.1.5.2: Restore t.Parallel() calls in all config tests
- [ ] 0.1.5.3: Verify tests pass with -race flag
- [ ] 0.1.5.4: Commit: "fix(template): restore t.Parallel() in config tests by isolating viper state"

**Files**:
- internal/apps/template/service/config/config_validation_test.go (refactored with isolated viper)

---

## Phase 0.2: Documentation Updates (Prevent Recurrence)

**Objective**: Update copilot instructions and architecture docs to prevent future violations
**Status**: ⏳ NOT STARTED
**Dependencies**: None - can run in parallel with Phase 0.1
**Est. LOE**: 2-3 hours

### Task 0.2.1: Update Testing Copilot Instructions

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: None
**Priority**: HIGH
**Est. LOE**: 1-2 hours

**Description**: Add CRITICAL tags and enforcement checklists to 03-02.testing.instructions.md to prevent future violations.

**Acceptance Criteria**:
- [ ] 0.2.1.1: Add CRITICAL tag at top: "BEFORE writing ANY tests, review this entire file"
- [ ] 0.2.1.2: Add explicit anti-pattern section: "NEVER create standalone test variants - ALWAYS use table-driven"
- [ ] 0.2.1.3: Add enforcement checklist: "All tests MUST: (1) use table-driven pattern for variants, (2) use app.Test() for handlers, (3) run t.Parallel() with isolated state"
- [ ] 0.2.1.4: Commit: "docs: enhance testing copilot instructions with CRITICAL tags and enforcement checklists"

**Files**:
- .github/instructions/03-02.testing.instructions.md (updated with prevention measures)

---

### Task 0.2.2: Update Architecture Documentation

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: None
**Priority**: MEDIUM
**Est. LOE**: 1 hour

**Description**: Add "Testing Patterns" section to ARCHITECTURE.md referencing copilot instructions.

**Acceptance Criteria**:
- [ ] 0.2.2.1: Add "Testing Patterns" section to ARCHITECTURE.md
- [ ] 0.2.2.2: Document app.Test() pattern as architectural standard
- [ ] 0.2.2.3: Document table-driven pattern as architectural standard
- [ ] 0.2.2.4: Cross-reference 03-02.testing.instructions.md
- [ ] 0.2.2.5: Commit: "docs: add testing patterns to architecture documentation"

**Files**:
- docs/arch/ARCHITECTURE.md (updated with testing patterns section)

---

## Phase 0.3: Global Mutation Target Fix (CRITICAL CORRECTION)

**Objective**: Fix mutation efficacy targets globally (85% → >=95% minimum, 98% ideal)
**Status**: ⏳ NOT STARTED
**Dependencies**: None - highest priority correction
**Est. LOE**: 1-2 hours

**User Feedback**: "one mistake i see if you reverted to minimum migrations >=85%. i changed the mutations floor to >=95% and ideal 98%; look in the entire project and fix those mutation targets globally"

### Task 0.3.1: Global Search and Replace Mutation Targets

**Status**: ✅ COMPLETE
**Owner**: LLM Agent
**Dependencies**: None
**Priority**: CRITICAL
**Est. LOE**: 1-2 hours
**Actual**: 2 hours
**Commit**: 234372f5

**Description**: Search entire project for "85%" mutation/coverage references and replace with ">=95% minimum, 98% ideal".

**Acceptance Criteria**:
- [x] 0.3.1.1: Search project for "85%" in mutation context: `grep -r "85" . --include="*.md" | grep -i "mutation\|efficacy\|coverage"`
- [x] 0.3.1.2: Replace with ">=95% minimum, 98% ideal" in: plan.md, completed.md, ARCHITECTURE.md, coverage docs, agent files (20 replacements across 7 files)
- [x] 0.3.1.3: Verify no instances of "≥85%" remain in mutation/coverage context (grep verification passed)
- [x] 0.3.1.4: Update Phase objectives to reflect >=95% minimum, 98% ideal
- [x] 0.3.1.5: Commit: "fix(docs): update mutation efficacy targets to >=95% minimum (98% ideal) globally - USER CORRECTION"

**Results**:
- 20 replacements across 7 files
- Pattern: Replace TARGET language ("≥85%", "85% target", "85% minimum") → "≥95% minimum" or "≥95% minimum, 98% ideal"
- Preserved: Historical ACHIEVEMENT numbers (85.3%, 87.9%, etc.) as factual data
- Copilot instructions (.github/instructions/03-02.testing.instructions.md): Already correct at >=95% (no changes needed)

**Files Modified**:
- docs/fixes-needed-plan-tasks-v4/plan.md (5 replacements)
- docs/fixes-needed-plan-tasks-v4/completed.md (7 replacements)
- docs/arch/ARCHITECTURE.md (3 replacements)
- docs/coverage-analysis-2026-01-27.md (1 replacement)
- docs/gremlins/MUTATIONS-TASKS.md (1 replacement)
- .github/agents/speckit.agent.md (1 replacement)
- .github/agents/plan-tasks-quizme.agent.md (2 replacements)

---

## Phase 0.4: KMS Modernization (LAST - Largest Migration)

**Objective**: Migrate KMS to service-template pattern, ≥95% coverage, ≥95% mutation
**Status**: ⏳ NOT STARTED - Tasks TBD after Phases 0.1-0.3
**Dependencies**: Phases 0.1-0.3 complete (all violations remediated, lessons learned)

**Note**: KMS is intentionally LAST - it's the largest service, most complex, and should benefit from all learnings from Phases 0.1-0.3 and historical Phases 1-5. Detailed tasks will be defined after completing violation remediation.

**Placeholder Tasks**:
- Task 0.4.1: TBD - Plan KMS migration strategy (formerly Task 6.1)
- Tasks 0.4.2-0.4.N: TBD - Implementation tasks (formerly Tasks 6.2-6.N)


## Phase 0.5: Template Mutation Improvement (DEFERRED)

**Objective**: Address remaining template mutation (currently 98.91% efficacy)
**Status**: ⏳ DEFERRED
**Priority**: LOW (template already exceeds 98% ideal)

### Task 0.5.1: Analyze Remaining TLS Generator Mutation

**Status**: ⏳ DEFERRED
**Owner**: LLM Agent
**Dependencies**: Phase 0.1 complete
**Priority**: LOW

**Description**: Analyze remaining tls_generator.go mutation (formerly Task 8.1).

**Acceptance Criteria**:
- [ ] 0.5.1.1: Review gremlins output
- [ ] 0.5.1.2: Identify survived mutation type
- [ ] 0.5.1.3: Analyze killability
- [ ] 0.5.1.4: Document findings

**Files**:
- test-output/template-mutation-analysis/ (create)
