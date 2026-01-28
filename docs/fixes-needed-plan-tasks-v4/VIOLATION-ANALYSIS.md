# Violation Analysis: V4 Completed Tasks

**Date**: 2026-01-28
**Scope**: Phases 1, 1.5, 2, 3, 4, 5, 7 (91 completed tasks)
**References**: 
- review-tasks-v5.md (user manual review)
- .github/instructions/03-02.testing.instructions.md (testing copilot instructions)
- docs/arch/ARCHITECTURE.md (architecture patterns)

---

## Executive Summary

Deep analysis of completed V4 tasks identified **5 major violation categories** affecting approximately **15-20 files** across template, cipher-im, and shared packages. All violations stem from ignoring established copilot instructions and architectural patterns documented in `.github/instructions/` and `docs/arch/`.

**Impact**: Reduced maintainability, increased technical debt, coverage gaps attributed to "practical limits" when actually fixable with proper patterns.

---

## Violation Categories

### 1. Standalone Test Functions Instead of Table-Driven Tests

**Severity**: HIGH  
**Affected Files**: 3+  
**Impact**: Maintenance burden, code duplication, harder to add test cases

**Description**: Test functions created as individual standalone tests (e.g., `TestValidateConfiguration_InvalidProtocol`, `TestValidateConfiguration_InvalidLogLevel`, `TestSessionManager_ValidateBrowserSession_JWS_InvalidToken`, etc.) instead of table-driven tests with variants as rows.

**Copilot Instruction Violated**: `.github/instructions/03-02.testing.instructions.md#L287-340` - "Table-Driven Tests Pattern - MANDATORY"

**Evidence**:
- `internal/apps/template/service/config/config_validation_test.go` - Multiple `TestValidateConfiguration_*` functions
- `internal/apps/template/service/server/businesslogic/session_manager_jws_test.go` - Multiple `TestSessionManager_ValidateBrowserSession_JWS_*` functions
- `internal/apps/template/service/server/businesslogic/session_manager_jwe_test.go` - Multiple `TestSessionManager_ValidateBrowserSession_JWE_*` functions

**Root Cause**: Agent generated tests without reviewing table-driven pattern requirement in copilot instructions.

**Remediation**: Refactor all standalone test variants into table-driven tests with `tests := []struct{name, input, want}` pattern.

---

### 2. Sad Path Service Startup Tests Not Table-Driven

**Severity**: MEDIUM  
**Affected Files**: 2+  
**Impact**: Poor test organization, harder to understand failure scenarios

**Description**: Error path tests for service initialization (invalid config, context cancellation, dependency errors) written as separate test functions instead of table-driven sad path tests.

**Copilot Instruction Violated**: `.github/instructions/03-02.testing.instructions.md#L287-340` - "Table-Driven Tests Pattern - MANDATORY"

**Evidence**:
- `internal/apps/template/service/server/application/application_test.go` - Multiple `Test..._Error` functions
- `internal/apps/template/service/server/builder/server_builder_test.go` - Separate error path tests for `Build` method

**Root Cause**: Agent treated error paths as exceptional cases rather than standard variants in table-driven pattern.

**Remediation**: Consolidate into table-driven tests with `{name: "invalid config", setup: ..., wantErr: ...}` pattern.

---

### 3. Real HTTPS Listeners in Tests (FORBIDDEN Anti-Pattern)

**Severity**: CRITICAL  
**Affected Files**: 1 (servers_test.go)  
**Impact**: Windows Firewall prompts, slow execution, resource leaks, violates TestMain pattern

**Description**: Tests start real HTTPS listeners on network ports (e.g., `TestDualServers_StartBothServers`, `TestDualServers_HealthEndpoints`) instead of using Fiber's `app.Test()` for in-memory handler testing.

**Copilot Instruction Violated**: `.github/instructions/03-02.testing.instructions.md#L148-200` - "Handler Testing with app.Test() - MANDATORY"

**Evidence**:
- `internal/apps/template/service/server/listener/servers_test.go` - Multiple `TestDualServers_*` functions start real HTTPS servers

**Architecture Violation**: Violates "NO HTTPS listeners in tests" pattern established in copilot instructions and architecture docs.

**Root Cause**: Agent misunderstood integration testing requirements and implemented real server startup instead of app.Test() pattern.

**Remediation**: 
1. Delete all `TestDualServers_*` functions that start real servers
2. Rewrite using `app.Test()` pattern with in-memory HTTP requests
3. Verify <1ms execution time (vs >1s for real server tests)

---

### 4. Dead Code Not Removed

**Severity**: MEDIUM  
**Affected Files**: 3+ functions/methods  
**Impact**: Inflated LOC, reduced coverage metrics, confusion about production usage

**Description**: Functions/methods with 0% coverage that are never called in production or tests were left in codebase instead of being removed.

**Evidence**:
- `PublicServer.PublicBaseURL` - 0% coverage, never called
- `UnsealKeysServiceFromSettings` wrapper methods (`EncryptKey`, `DecryptKey`, `Shutdown`) - 0% coverage, struct never instantiated
- `EnsureSignatureAlgorithmType` - 23.1% coverage, not used in production (design flaw)

**Root Cause**: Agent accepted "practical coverage limits" without analyzing whether code is actually dead code.

**Remediation**: 
1. Remove `PublicServer.PublicBaseURL` method entirely
2. Remove `UnsealKeysServiceFromSettings` struct and all wrapper methods
3. Either remove `EnsureSignatureAlgorithmType` or document why it exists (future extensibility?)

---

### 5. Global State Pollution (t.Parallel() Removed Instead of Refactored)

**Severity**: MEDIUM  
**Affected Files**: 1+ (config tests)  
**Impact**: Tests cannot run in parallel, slower execution, masks concurrency bugs

**Description**: Tests removed `t.Parallel()` due to viper global state pollution instead of refactoring tests to avoid global state.

**Copilot Instruction Violated**: `.github/instructions/03-02.testing.instructions.md` - "Test Concurrency Requirements"

**Evidence**:
- `internal/apps/template/service/config/config_validation_test.go` - Comments indicate viper global state prevents t.Parallel()

**Root Cause**: Agent chose quick fix (remove t.Parallel()) over proper refactoring (isolate viper state per test).

**Remediation**: Refactor config tests to create isolated viper instances per test, restore t.Parallel().

---

## Violation Summary Table

| Category | Severity | Affected Files | LOC Impact | Remediation LOE |
|----------|----------|----------------|------------|-----------------|
| Standalone tests → table-driven | HIGH | 3+ | ~500-800 lines | 4-6 hours |
| Sad path not table-driven | MEDIUM | 2+ | ~200-300 lines | 2-3 hours |
| Real HTTPS listeners | CRITICAL | 1 | ~150-200 lines | 1-2 hours |
| Dead code not removed | MEDIUM | 3+ | ~200-300 lines | 1-2 hours |
| t.Parallel() removed | MEDIUM | 1+ | ~50-100 lines | 1-2 hours |
| **TOTAL** | - | **10-15** | **1100-1700** | **9-15 hours** |

---

## Root Cause Analysis

**Pattern**: All violations share a common root cause - **agent did not review copilot instructions before implementing tests**.

**Evidence**:
1. Table-driven pattern is MANDATORY in 03-02.testing.instructions.md, but ignored
2. app.Test() pattern is MANDATORY with explicit "NEVER start HTTPS listeners" rule, but ignored
3. Dead code removal is standard practice, but agent accepted "practical limits" without analysis

**Contributing Factors**:
1. **No pre-implementation instruction review**: Agent should read relevant copilot instructions before starting each task
2. **"Practical limits" justification misuse**: Coverage gaps attributed to practical limits when actually violations
3. **Quick fix over proper refactoring**: Removing t.Parallel() instead of fixing root cause

---

## Recommendations for Preventing Future Violations

### 1. Documentation Updates

**Update `.github/instructions/03-02.testing.instructions.md`**:
- Add CRITICAL tag: "BEFORE writing ANY tests, review this entire file"
- Add explicit anti-pattern section: "NEVER create standalone test variants - ALWAYS use table-driven"
- Add enforcement checklist: "All tests MUST: (1) use table-driven pattern for variants, (2) use app.Test() for handlers, (3) run t.Parallel() with isolated state"

**Update `docs/arch/ARCHITECTURE.md`**:
- Add "Testing Patterns" section referencing 03-02.testing.instructions.md
- Document app.Test() pattern as architectural standard
- Document table-driven pattern as architectural standard

### 2. Agent Workflow Changes

**MANDATORY pre-task instruction review**:
1. Before Task N.M: Read .github/instructions/03-02.testing.instructions.md sections relevant to task
2. Before implementing: Grep for examples in codebase matching required pattern
3. After implementing: Verify adherence to pattern via self-review

**Quality Gate Enhancement**:
- Add to completion checklist: "All tests use table-driven pattern where applicable"
- Add to completion checklist: "No real HTTPS listeners (grep for 'listener.Addr' in test files)"
- Add to completion checklist: "All dead code identified and removed (0% coverage = remove unless justified)"

### 3. Spec Kit Process Enhancement

**Add to constitution.md**:
- Testing patterns are MANDATORY and non-negotiable
- "Practical limits" justification requires evidence that pattern cannot be applied
- Dead code MUST be removed (0% coverage with no production references = delete)

---

## Next Steps

1. Create Phase 0.1: Violation Remediation (prepend to tasks.md)
2. Create Phase 0.2: Documentation Updates to prevent recurrence
3. Renumber existing incomplete phases (6 → 0.3, 8 → 0.4)
4. Execute Phase 0.1 tasks to fix all violations
5. Execute Phase 0.2 tasks to update copilot instructions and architecture docs
6. Re-run coverage/mutation analysis to verify improvements

---

## Evidence Archive

**Source Files**:
- docs/fixes-needed-plan-tasks-v5/review-tasks-v4.md (user manual review)
- .github/instructions/03-02.testing.instructions.md (violated instructions)
- internal/apps/template/service/config/config_validation_test.go (standalone tests)
- internal/apps/template/service/server/businesslogic/session_manager_jws_test.go (standalone tests)
- internal/apps/template/service/server/businesslogic/session_manager_jwe_test.go (standalone tests)
- internal/apps/template/service/server/application/application_test.go (sad path not table-driven)
- internal/apps/template/service/server/builder/server_builder_test.go (sad path not table-driven)
- internal/apps/template/service/server/listener/servers_test.go (real HTTPS listeners)
- docs/fixes-needed-plan-tasks-v4/completed.md (dead code evidence)

