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

MUST: Complete all tasks without stopping
MUST: Do regular commit and push
MUST: Must track progress in docs\SERVICE-TEMPLATE.md with checkmarks for complete tasks and phases
MUST: Create evidence of completion per phase and task in docs/learn-im-migration/evidence/; prefixed with phase number and task number
MUST: Write post-mortem analysis for each completed phase in docs/learn-im-migration/post-mortems/
MUST: Use port-mortem analysis to improve future phases and tasks; add/append/insert extra tasks or phases as needed

---

## 📊 QUICK PROGRESS TRACKER

### ⏳ P0.0: Test Baseline Establishment

- [x] P0.0.1: Re-run all unit, integration, and e2e tests with code coverage; save the baseline
- [x] P0.0.2: Identify broken tests
- [x] P0.0.3: Fix all broken tests
- [ ] P0.0.4: Fix all lint and format errors (BLOCKED: golangci-lint config issue)

### ✅ P0.1: Test Performance Optimization

- [x] P0.1.1: Re-run all unit, integration, and e2e tests with code coverage; save the baseline
- [x] P0.1.2: Identify top 10 slowest tests
- [x] P0.1.3: Analyze root causes, and fix problems or make them more efficient
- [ ] P0.1.4: Fix all lint and format errors (BLOCKED: golangci-lint config issue)

**Post-Mortem**: `post-mortems/P0.1-test-performance-optimization.md`

### ⏳ P0.2: TestMain Pattern (PARTIAL)

- [ ] P0.2.1: Re-run all unit, integration, and e2e tests with code coverage; save the baseline
- [x] P0.2.2: server/testmain_test.go created
- [x] P0.2.3: crypto/testmain_test.go created
- [x] P0.2.4: template/server/test_main_test.go created
- [ ] P0.2.5: e2e/testmain_e2e_test.go needs creation
- [ ] P0.2.6: integration/testmain_integration_test.go needs creation
- [ ] P0.2.7: Measure test speedup (before/after)
- [ ] P0.2.8: Re-run all unit, integration, and e2e tests with code coverage; save the baseline

### ⏳ P0.3: Refactor internal/learn/server/ Files

- [ ] P0.3.1: Move repository code from internal/learn/server/ to internal/learn/server/repository/
- [ ] P0.3.2: Move authentication/authorization code from internal/learn/server/ to internal/learn/server/realms/
- [ ] P0.3.3: Move auth-related middleware from internal/learn/server/ to internal/learn/server/realms/
- [ ] P0.3.4: Move realm validation code from internal/learn/server/ to internal/learn/server/realms/
- [ ] P0.3.5: Move API handler code from internal/learn/server/ to internal/learn/server/apis/
- [ ] P0.3.6: Move business logic code from internal/learn/server/ to internal/learn/server/businesslogic/
- [ ] P0.3.7: Move utility functions from internal/learn/server/ to internal/learn/server/util/
- [ ] P0.3.8: Leave server.go and listener-related code in internal/learn/server/ package
- [ ] P0.3.9: Update all imports across the codebase
- [ ] P0.3.10: Run all tests recursively under internal\learn to ensure no breakage

### ⏳ P0.4: Refactor internal/template/server/ Files

- [ ] P0.4.1: Move repository code from internal/template/server/ to internal/template/server/repository/
- [ ] P0.4.2: Move listener-related code from internal/template/server/ to internal/template/server/listener/
- [ ] P0.4.3: Move API handler code from internal/template/server/ to internal/template/server/apis/
- [ ] P0.4.4: Move authentication/authorization code from internal/template/server/ to internal/template/server/realms/
- [ ] P0.4.5: Move business logic code from internal/template/server/ to internal/template/server/businesslogic/
- [ ] P0.4.6: Move utility functions from internal/template/server/ to internal/template/server/util/
- [ ] P0.4.7: Leave core files (service_template.go if needed) in internal/template/server/ package
- [ ] P0.4.8: Update all imports across the codebase
- [ ] P0.4.9: Run all tests recursively under internal\template to ensure no breakage

### ⏳ P1.0: File Size Analysis

- [ ] P1.0.1: Run file size scan for files >400 lines
- [ ] P1.0.2: Document files approaching 500-line hard limit
- [ ] P1.0.3: Create refactoring plan for oversized files

### ⏳ P2.0: Hardcoded Password Fixes (PARTIAL)

- [x] P2.0.1: Replaced 18 GeneratePasswordSimple() instances
- [ ] P2.0.2: Replaced 12 realm_validation_test.go passwords with GeneratePasswordSimple(), with add/remove characters for password policy tests
- [ ] P2.0.3: Verify no new hardcoded passwords added
- [ ] P2.0.4: Run realm validation tests

### ⏳ P3.0: Windows Firewall Exception Fix

- [ ] P3.0.1: Scan for `0.0.0.0` bindings in test files
- [ ] P3.0.2: Scan for hardcoded ports (`:8080`, `:9090`)
- [ ] P3.0.3: Replace with `cryptoutilMagic.IPv4Loopback` and port `:0`
- [ ] P3.0.4: Verify no Windows Firewall prompts during test runs (i.e. server.test.exe built by Go)
- [ ] P3.0.5: Add detection to lint-go-test

### ⏳ P4.0: context.TODO() Replacement

- [ ] P4.0.1: Replace context.TODO() in server_lifecycle_test.go:40
- [ ] P4.0.2: Replace context.TODO() in register_test.go:355
- [ ] P4.0.3: Verify zero context.TODO() in internal/learn
- [ ] P4.0.4: Run tests to confirm behavior unchanged

### ⏳ P5.0: Switch Statement Conversion

- [ ] P5.0.1: Convert if/else chains to switch in handlers.go
- [ ] P5.0.2: Verify golangci-lint passes
- [ ] P5.0.3: Run handler tests

### ⏳ P6.0: Quality Gates Execution

- [ ] P6.0.1: Build validation → learn_build_evidence.txt
- [ ] P6.0.2: Linting validation → learn_lint_evidence.txt
- [ ] P6.0.3: Test validation → learn_test_evidence.txt
- [ ] P6.0.4: Coverage validation → learn_coverage_summary.txt + .html
  - [ ] P6.0.4a: Production code ≥95%
  - [ ] P6.0.4b: Infrastructure code ≥98%
- [ ] P6.0.5: Mutation testing → learn_mutation_evidence.txt (≥80%)
- [ ] P6.0.6: Race detection → learn_race_evidence.txt (zero races)

### ⏳ P7.1: Remove Obsolete Database Tables

- [ ] P7.1.1: Remove users_jwks table
- [ ] P7.1.2: Remove users_messages_jwks table
- [ ] P7.1.3: Remove messages_jwks table
- [ ] P7.1.4: Update migration files
- [ ] P7.1.5: Verify no code references remain
- [ ] P7.1.6: Run tests with new schema

### ⏳ P7.3: Implement Barrier Encryption for JWKs

**NOTE**: Must complete BEFORE P7.2

- [ ] P7.3.1: Integrate KMS barrier encryption pattern
- [ ] P7.3.2: Update JWK storage to use barrier encryption
- [ ] P7.3.3: Update JWK retrieval to use barrier decryption
- [ ] P7.3.4: Add tests for encryption/decryption
- [ ] P7.3.5: Run E2E tests

### ⏳ P7.2: Use EncryptBytesWithContext Pattern

**NOTE**: Depends on P7.3 completion

- [ ] P7.2.1: Update jwe_message_util.go to use EncryptBytesWithContext
- [ ] P7.2.2: Replace old encryption calls with context-aware version
- [ ] P7.2.3: Replace old decryption calls with context-aware version
- [ ] P7.2.4: Run encryption tests

### ⏳ P7.4: Manual Key Rotation Admin API

- [ ] P7.4.1: Create admin_handlers.go with rotation endpoints
- [ ] P7.4.2: Add POST /admin/v1/keys/rotate endpoint
- [ ] P7.4.3: Add GET /admin/v1/keys/status endpoint
- [ ] P7.4.4: Update OpenAPI spec
- [ ] P7.4.5: Add E2E tests for rotation

### ⏳ P8.0: CGO Check Consolidation

- [ ] P8.0.1: Consolidate CGO detection logic
- [ ] P8.0.2: Document CGO requirements
- [ ] P8.0.3: Add tests for CGO detection

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

2. **P0.0**: Test Baseline Establishment (IMMEDIATE)
   - Establish test baseline and fix broken tests

3. **P0.1**: Test Performance Optimization
   - Identify and fix slow tests

4. **P0.2**: TestMain Pattern
   - Complete TestMain implementation

5. **P0.3**: Refactor internal/learn/server/ Files
   - Code organization for learn service

6. **P0.4**: Refactor internal/template/server/ Files
   - Code organization for template service

7. **Complete P1.0**: File Size Analysis
   - Independent of test failures
   - Identifies refactoring needs

8. **Complete P2.0**: Pragma comments for passwords
   - Quick win, independent

9. **P3.0**: Windows Firewall fixes
   - CRITICAL recurring regression prevention

10. **P4.0**: context.TODO() replacement
    - Quick win, code quality

11. **P5.0**: Switch statement conversion
    - Code quality improvement

12. **P6.0**: Quality Gates
    - **MANDATORY BEFORE** claiming any phase complete
    - Provides evidence for completion

13. **P7.x (Sequence: P7.1 → P7.3 → P7.2 → P7.4)**:
    - P7.1: Remove obsolete tables
    - P7.3: Barrier encryption (prerequisite for P7.2)
    - P7.2: EncryptBytesWithContext (depends on P7.3)
    - P7.4: Manual rotation API

14. **P8.0**: CGO check consolidation
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

### Quality Gates (P6.0)

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

1. P8.11: Magic constants migration
2. P8.12: Magic constants consolidation
3. P8.13: Password generation (18 instances)
4. P9.2: TOTP test data fields
5. P18: Service instantiation extraction
6. P1 (server): TestMain - server package
7. P1 (crypto): TestMain - crypto package

**ACTION REQUIRED**: Create post-mortems for all completed phases before claiming new phases complete.

**Template Location**: See docs/SERVICE-TEMPLATE.backup.md for post-mortem template.

---

## 🔄 DECISIONS MADE (From QUIZME)

1. **Hardcoded Passwords**: Hybrid approach - Replace GeneratePasswordSimple() ✅ + Add pragma to validation tests ⏳
2. **Windows Firewall Phase**: Execute immediately after TestMain (CRITICAL recurring regression)
3. **P7.x Sequence**: P7.1 → P7.3 → P7.2 → P7.4 (dependency-based order)
4. **TestMain Priority**: HIGH (significant test speedup for heavyweight setup)
5. **CGO Check Priority**: MEDIUM (code cleanup, not blocking)

---

## 📝 NOTES

### Investigations Completed

1. ✅ **internal/learn/crypto should NOT be removed** - Package needed for password hashing, must use shared/hash
2. ✅ **adaptive-sim should move** - Should be in internal/identity/tools/ not cicd/ (deferred to P8.x)
3. ✅ **identity_requirements_check is actively used** - Keep in cicd/, referenced in workflows and docs

### Tools to Update (User Request)

Per original user request, these still need review/updates:

- [ ] Pre-commit hooks configuration
- [ ] GitHub workflows (13 workflow files)
- [ ] Python scripts (need to identify which ones)
- [x] Code generation (oapi-codegen fixed ✅)

---

## 🎯 NEXT IMMEDIATE ACTIONS

1. **P0.0: Test Baseline Establishment**
   - Re-run all unit, integration, and e2e tests with code coverage; save the baseline
   - Identify broken tests
   - Fix all broken tests

2. **P0.1: Test Performance Optimization**
   - Re-run all unit, integration, and e2e tests with code coverage; save the baseline
   - Identify top 10 slowest tests
   - Analyze root causes, and fix problems or make them more efficient

3. **Fix Test Failures** (BLOCKING)
   - Read ./test-output/all_tests_output.txt for details
   - Fix database migration idempotence
   - Fix Windows Docker compatibility
   - Fix server test hangs
   - Fix config shorthand duplicates

4. **P0.2: Complete TestMain Pattern**
   - Create e2e/testmain_e2e_test.go
   - Create integration/testmain_integration_test.go
   - Measure speedup

5. **P0.3: Refactor internal/learn/server/ Files**
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

6. **P0.4: Refactor internal/template/server/ Files**
   - Move repository code from internal/template/server/ to internal/template/server/repository/
   - Move listener-related code from internal/template/server/ to internal/template/server/listener/
   - Move API handler code from internal/template/server/ to internal/template/server/apis/
   - Move authentication/authorization code from internal/template/server/ to internal/template/server/realms/
   - Move business logic code from internal/template/server/ to internal/template/server/businesslogic/
   - Move utility functions from internal/template/server/ to internal/template/server/util/
   - Leave core files (service_template.go if needed) in internal/template/server/ package
   - Update all imports across the codebase
   - Run all tests recursively under internal\template to ensure no breakage

7. **P1.0: File Size Analysis**
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
