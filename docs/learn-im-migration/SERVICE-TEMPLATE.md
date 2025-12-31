# Learn-IM Service Template Migration - Active Tasks

**Last Updated**: 2025-12-30
**Status**: 10/17 phases remaining

Most of the Code Resides In

- internal/learn/
- internal/template/server/

Most of the Docs Resides In

- docs/learn-im-migration/SERVICE-TEMPLATE.md
- docs/learn-im-migration/CMD-PATTERN.md
- docs/learn-im-migration/KNOWN-ISSUES.md
- docs/learn-im-migration/evidence/

## CRITICAL INSTRUCTIONS

MUST: Do regular commit and push
MUST: Must track progress in docs\SERVICE-TEMPLATE.md with checkmarks for complete tasks and phases
MUST: Create evidence of completion per phase and task in docs/learn-im-migration/evidence/; prefixed with phase number and task number
MUST: Write post-mortem analysis for each completed phase in docs/learn-im-migration/post-mortems/
MUST: Use port-mortem analysis to improve future phases and tasks; add/append/insert extra tasks or phases as needed

---

## 📊 QUICK PROGRESS TRACKER

### ⏳ Phase 0: Test Baseline Establishment

- [ ] Re-run all unit, integration, and e2e tests with code coverage; save the baseline
- [ ] Identify broken tests
- [ ] Fix all broken tests
- [ ] Fix all lint and format errors

### ⏳ Phase 0.1: Test Performance Optimization

- [ ] Re-run all unit, integration, and e2e tests with code coverage; save the baseline
- [ ] Identify top 10 slowest tests
- [ ] Analyze root causes, and fix problems or make them more efficient
- [ ] Fix all lint and format errors

### ⏳ Phase 0.2: TestMain Pattern (PARTIAL)

- [ ] Re-run all unit, integration, and e2e tests with code coverage; save the baseline
- [x] server/testmain_test.go created
- [x] crypto/testmain_test.go created
- [x] template/server/test_main_test.go created
- [ ] e2e/testmain_e2e_test.go needs creation
- [ ] integration/testmain_integration_test.go needs creation
- [ ] Measure test speedup (before/after)
- [ ] Re-run all unit, integration, and e2e tests with code coverage; save the baseline

### ⏳ Phase 0.3: Refactor internal/learn/server/ Files

- [ ] Move repository code from internal/learn/server/ to internal/learn/server/repository/
- [ ] Move authentication/authorization code from internal/learn/server/ to internal/learn/server/realms/
- [ ] Move auth-related middleware from internal/learn/server/ to internal/learn/server/realms/
- [ ] Move realm validation code from internal/learn/server/ to internal/learn/server/realms/
- [ ] Move API handler code from internal/learn/server/ to internal/learn/server/apis/
- [ ] Move business logic code from internal/learn/server/ to internal/learn/server/businesslogic/
- [ ] Move utility functions from internal/learn/server/ to internal/learn/server/util/
- [ ] Leave server.go and listener-related code in internal/learn/server/ package
- [ ] Update all imports across the codebase
- [ ] Run all tests recursively under internal\learn to ensure no breakage

### ⏳ Phase 0.4: Refactor internal/template/server/ Files

- [ ] Move repository code from internal/template/server/ to internal/template/server/repository/
- [ ] Move listener-related code from internal/template/server/ to internal/template/server/listener/
- [ ] Move API handler code from internal/template/server/ to internal/template/server/apis/
- [ ] Move authentication/authorization code from internal/template/server/ to internal/template/server/realms/
- [ ] Move business logic code from internal/template/server/ to internal/template/server/businesslogic/
- [ ] Move utility functions from internal/template/server/ to internal/template/server/util/
- [ ] Leave core files (service_template.go if needed) in internal/template/server/ package
- [ ] Update all imports across the codebase
- [ ] Run all tests recursively under internal\template to ensure no breakage

### ✅ Phase 1: File Size Analysis

- [ ] Run file size scan for files >400 lines
- [ ] Document files approaching 500-line hard limit
- [ ] Create refactoring plan for oversized files

### ⏳ Phase 2: Hardcoded Password Fixes (PARTIAL)

- [x] Replaced 18 GeneratePasswordSimple() instances
- [ ] Replaced 12 realm_validation_test.go passwords with GeneratePasswordSimple(), with add/remove characters for password policy tests
- [ ] Verify no new hardcoded passwords added
- [ ] Run realm validation tests

### ⏳ Phase 3: Windows Firewall Exception Fix

- [ ] Scan for `0.0.0.0` bindings in test files
- [ ] Scan for hardcoded ports (`:8080`, `:9090`)
- [ ] Replace with `cryptoutilMagic.IPv4Loopback` and port `:0`
- [ ] Verify no Windows Firewall prompts during test runs (i.e. server.test.exe built by Go)
- [ ] Add detection to lint-go-test

### ⏳ Phase 4: context.TODO() Replacement

- [ ] Replace context.TODO() in server_lifecycle_test.go:40
- [ ] Replace context.TODO() in register_test.go:355
- [ ] Verify zero context.TODO() in internal/learn
- [ ] Run tests to confirm behavior unchanged

### ⏳ Phase 5: Switch Statement Conversion

- [ ] Convert if/else chains to switch in handlers.go
- [ ] Verify golangci-lint passes
- [ ] Run handler tests

### ⏳ Phase 6: Quality Gates Execution

- [ ] 6A: Build validation → learn_build_evidence.txt
- [ ] 6B: Linting validation → learn_lint_evidence.txt
- [ ] 6C: Test validation → learn_test_evidence.txt
- [ ] 6D: Coverage validation → learn_coverage_summary.txt + .html
  - [ ] Production code ≥95%
  - [ ] Infrastructure code ≥98%
- [ ] 6E: Mutation testing → learn_mutation_evidence.txt (≥80%)
- [ ] 6F: Race detection → learn_race_evidence.txt (zero races)

### ⏳ Phase 7a: Remove Obsolete Database Tables

- [ ] Remove users_jwks table
- [ ] Remove users_messages_jwks table
- [ ] Remove messages_jwks table
- [ ] Update migration files
- [ ] Verify no code references remain
- [ ] Run tests with new schema

### ⏳ Phase 7c: Implement Barrier Encryption for JWKs

**NOTE**: Must complete BEFORE Phase 7b

- [ ] Integrate KMS barrier encryption pattern
- [ ] Update JWK storage to use barrier encryption
- [ ] Update JWK retrieval to use barrier decryption
- [ ] Add tests for encryption/decryption
- [ ] Run E2E tests

### ⏳ Phase 7b: Use EncryptBytesWithContext Pattern

**NOTE**: Depends on Phase 7c completion

- [ ] Update jwe_message_util.go to use EncryptBytesWithContext
- [ ] Replace old encryption calls with context-aware version
- [ ] Replace old decryption calls with context-aware version
- [ ] Run encryption tests

### ⏳ Phase 7d: Manual Key Rotation Admin API

- [ ] Create admin_handlers.go with rotation endpoints
- [ ] Add POST /admin/v1/keys/rotate endpoint
- [ ] Add GET /admin/v1/keys/status endpoint
- [ ] Update OpenAPI spec
- [ ] Add E2E tests for rotation

### ⏳ Phase 1.6: CGO Check Consolidation

- [ ] Consolidate CGO detection logic
- [ ] Document CGO requirements
- [ ] Add tests for CGO detection

---

## 🎯 CURRENT BLOCKERS

### Test Failures (CRITICAL - BLOCKING ALL PROGRESS)

Based on test run from test-output/all_tests_output.txt:

#### 1. internal/cmd/learn (1 failure)

- **TestPrintIMVersion**: Expected exit code 0, got 1
- **Root Cause**: Version command returning non-zero status
- **Fix Needed**: Investigate version command exit code logic

#### 2. internal/cmd/learn/im (36 failures)

- **Database Migration Issues**:
  - "index idx_users_username already exists"
  - "Dirty database version 1"
  - **Fix Needed**: Add IF NOT EXISTS to CREATE INDEX, add migration cleanup

- **Docker Compatibility**:
  - TestInitDatabase_PostgreSQL panics: "rootless Docker is not supported on Windows"
  - **Fix Needed**: Skip PostgreSQL tests on Windows OR use non-rootless mode

- **Output Assertion Failures**:
  - Tests expect empty output but get Unicode error messages
  - **Fix Needed**: Normalize output or fix encoding in assertions

#### 3. internal/learn/server (TIMEOUT)

- **Test hung for 902.850s** (15 minutes)
- **Goroutine Analysis**: 1500+ goroutines blocked in IO wait
- **Servers listening** on 127.0.0.1:8080 and 127.0.0.1:9090
- **Fix Needed**: Add proper shutdown/cleanup to server tests

#### 4. internal/shared/config (1 failure)

- **TestAnalyzeSettings_RealSettings**: Duplicate shorthands [R C R C]
- **Fix Needed**: Deduplicate command-line flag shorthands

---

## 📋 EXECUTION ORDER (DEPENDENCY-BASED)

1. **FIX TEST FAILURES FIRST** (BLOCKING)
   - Cannot proceed with phases until tests pass
   - Database migrations must be idempotent
   - Server tests need cleanup mechanisms
   - Config shorthands need deduplication

2. **Phase 0**: Test Baseline Establishment (IMMEDIATE)
   - Establish test baseline and fix broken tests

3. **Phase 0.1**: Test Performance Optimization
   - Identify and fix slow tests

4. **Phase 0.2**: TestMain Pattern
   - Complete TestMain implementation

5. **Phase 0.3**: Refactor internal/learn/server/ Files
   - Code organization for learn service

6. **Phase 0.4**: Refactor internal/template/server/ Files
   - Code organization for template service

7. **Complete Phase 1**: File Size Analysis
   - Independent of test failures
   - Identifies refactoring needs

8. **Complete Phase 2**: Pragma comments for passwords
   - Quick win, independent

9. **Phase 3**: Windows Firewall fixes
   - CRITICAL recurring regression prevention

10. **Phase 4**: context.TODO() replacement
    - Quick win, code quality

11. **Phase 5**: Switch statement conversion
    - Code quality improvement

12. **Phase 6**: Quality Gates
    - **MANDATORY BEFORE** claiming any phase complete
    - Provides evidence for completion

13. **Phase 7 (Sequence: 7a → 7c → 7b → 7d)**:
    - 7a: Remove obsolete tables
    - 7c: Barrier encryption (prerequisite for 7b)
    - 7b: EncryptBytesWithContext (depends on 7c)
    - 7d: Manual rotation API

14. **Phase 1.6**: CGO check consolidation
    - Final cleanup task

---

## ⚠️ CRITICAL REQUIREMENTS

### Evidence Required for Completion

**NO PHASE IS COMPLETE WITHOUT**:

1. ✅ All checkboxes marked
2. ✅ Evidence files created in ./test-output/
3. ✅ Post-mortem analysis written
4. ✅ Git commit with conventional message
5. ✅ Quality gates passed (Phase 6)

### Quality Gates (Phase 6)

**MANDATORY** for claiming completion:

- Build: Zero errors
- Linting: Zero violations (no exceptions)
- Tests: All pass, zero skips
- Coverage: ≥95% production, ≥98% infrastructure
- Mutation: ≥80% efficacy
- Race: Zero race conditions

### File Size Limits

**Per copilot instruction 03-01.coding.instructions.md**:

- Soft: 300 lines (ideal)
- Medium: 400 lines (acceptable with justification)
- **Hard: 500 lines (NEVER EXCEED)**
  - IMMEDIATE refactoring required if file reaches 500 lines

### Git Hygiene

**NEVER commit without**:

- Conventional commit message format
- Reference to evidence files
- Tests passing
- Linting clean

---

## 📚 POST-MORTEM TRACKING

### Completed Phases Needing Post-Mortems

The following phases are marked complete in backup but **LACK POST-MORTEMS**:

1. Phase 8.11: Magic constants migration
2. Phase 8.12: Magic constants consolidation
3. Phase 8.13: Password generation (18 instances)
4. Phase 9.2: TOTP test data fields
5. Phase 18: Service instantiation extraction
6. Phase 1 (server): TestMain - server package
7. Phase 1 (crypto): TestMain - crypto package

**ACTION REQUIRED**: Create post-mortems for all completed phases before claiming new phases complete.

**Template Location**: See docs/SERVICE-TEMPLATE.backup.md for post-mortem template.

---

## 🔄 DECISIONS MADE (From QUIZME)

1. **Hardcoded Passwords**: Hybrid approach - Replace GeneratePasswordSimple() ✅ + Add pragma to validation tests ⏳
2. **Windows Firewall Phase**: Execute immediately after TestMain (CRITICAL recurring regression)
3. **Phase 7 Sequence**: 7a → 7c → 7b → 7d (dependency-based order)
4. **TestMain Priority**: HIGH (significant test speedup for heavyweight setup)
5. **CGO Check Priority**: MEDIUM (code cleanup, not blocking)

---

## 📝 NOTES

### Investigations Completed

1. ✅ **internal/learn/crypto should NOT be removed** - Package needed for password hashing, must use shared/hash
2. ✅ **adaptive-sim should move** - Should be in internal/identity/tools/ not cicd/ (deferred to Phase 8)
3. ✅ **identity_requirements_check is actively used** - Keep in cicd/, referenced in workflows and docs

### Tools to Update (User Request)

Per original user request, these still need review/updates:

- [ ] Pre-commit hooks configuration
- [ ] GitHub workflows (13 workflow files)
- [ ] Python scripts (need to identify which ones)
- [x] Code generation (oapi-codegen fixed ✅)

---

## 🎯 NEXT IMMEDIATE ACTIONS

1. **Phase 0: Test Baseline Establishment**
   - Re-run all unit, integration, and e2e tests with code coverage; save the baseline
   - Identify broken tests
   - Fix all broken tests

2. **Phase 0.1: Test Performance Optimization**
   - Re-run all unit, integration, and e2e tests with code coverage; save the baseline
   - Identify top 10 slowest tests
   - Analyze root causes, and fix problems or make them more efficient

3. **Fix Test Failures** (BLOCKING)
   - Read ./test-output/all_tests_output.txt for details
   - Fix database migration idempotence
   - Fix Windows Docker compatibility
   - Fix server test hangs
   - Fix config shorthand duplicates

4. **Phase 0.2: Complete TestMain Pattern**
   - Create e2e/testmain_e2e_test.go
   - Create integration/testmain_integration_test.go
   - Measure speedup

5. **Phase 0.3: Refactor internal/learn/server/ Files**
   - Move repository code from internal/learn/server/ to internal/learn/server/repository/
   - Move authentication/authorization code from internal/learn/server/ to internal/learn/server/realms/
   - Move auth-related middleware from internal/learn/server/ to internal/learn/server/realms/
   - Move realm validation code from internal/learn/server/ to internal/learn/server/realms/
   - Move API handler code from internal/learn/server/ to internal/learn/server/apis/
   - Move business logic code from internal/learn/server/ to internal/learn/server/businesslogic/
   - Move utility functions from internal/learn/server/ to internal/learn/server/util/
   - Leave server.go and listener-related code in internal/learn/server/ package
   - Update all imports across the codebase
   - Run all tests recursively under internal\learn to ensure no breakage

6. **Phase 0.4: Refactor internal/template/server/ Files**
   - Move repository code from internal/template/server/ to internal/template/server/repository/
   - Move listener-related code from internal/template/server/ to internal/template/server/listener/
   - Move API handler code from internal/template/server/ to internal/template/server/apis/
   - Move authentication/authorization code from internal/template/server/ to internal/template/server/realms/
   - Move business logic code from internal/template/server/ to internal/template/server/businesslogic/
   - Move utility functions from internal/template/server/ to internal/template/server/util/
   - Leave core files (service_template.go if needed) in internal/template/server/ package
   - Update all imports across the codebase
   - Run all tests recursively under internal\template to ensure no breakage

7. **Phase 1: File Size Analysis**
   - Run scan: `Get-ChildItem -Recurse -Filter "*.go" -Path internal/learn | Where-Object { (Get-Content $_.FullName).Count -gt 400 }`
   - Document findings
   - Plan refactoring

8. **Continue Sequential Execution**
   - Mark checkboxes as completed
   - Create evidence files
   - Write post-mortems
   - Commit with evidence

---

**Remember**: This document is the SINGLE SOURCE OF TRUTH for incomplete work. Update checkboxes in real-time as tasks complete.
