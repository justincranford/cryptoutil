# SERVICE-TEMPLATE v4 - Complete Remediation Plan

**Created**: 2026-01-03
**Status**: ACTIVE - ALL TASKS HIGHEST PRIORITY AND BLOCKING
**Previous**: SERVICE-TEMPLATE-v3.md (95% complete but WITH CRITICAL VIOLATIONS)

---

## CRITICAL VIOLATIONS SUMMARY

- [ ] **bcrypt FIPS violation** (16 instances) - Replace with LowEntropyRandom (PBKDF2)
- [ ] **No Hash Service integration** - Use `internal/shared/crypto/hash/`
- [ ] **Windows Firewall triggers** (11 instances of `0.0.0.0`) - Change to `127.0.0.1`
- [ ] **Template linting violations** (50+ issues)
- [ ] **No CICD non-FIPS linter** - Add checkNonFips to pre-commit hooks
- [ ] **Pepper not implemented** - Move to final phase (strategically last)

**Migration Order** (per `.github/instructions/02-02.service-template.instructions.md`):
cipher-im → jose-ja → pki-ca → identity services (authz, idp, rs, rp, spa) → sm-kms

## PHASE 1: FIPS COMPLIANCE - REPLACE bcrypt WITH LowEntropyRandom

**Estimated**: 6-8 hours | **CRITICAL BLOCKER**

### Task 1.1: Integrate Hash Service from internal/shared/crypto/hash/

- [ ] **1.1.1** Verify Hash Service exists in `internal/shared/crypto/hash/` (NOT sm-kms, NOT cipher/crypto)
  - If missing, extract from existing implementation
  - Ensure registries: LowEntropyDeterministic, LowEntropyRandom, HighEntropyDeterministic, HighEntropyRandom
  - Ensure version framework (v1/v2/v3 support)
  - Ensure hash format: `{version}:{algorithm}:{iterations}:base64(salt):base64(hash)}`

- [ ] **1.1.2** Inject Hash Service into template realms UserServiceImpl
  ```go
  import cryptoutilHash "cryptoutil/internal/shared/crypto/hash"
  type UserServiceImpl struct {
      hashService *cryptoutilHash.Service
  }
  ```

- [ ] **1.1.3** Configure for LowEntropyRandom registry (passwords)
  - Set current version: v3 (2025 OWASP)
  - Use PBKDF2-HMAC-SHA256 with OWASP-safe pre-configured parameters
  - Future-proof via standardized encoding with versioning

### Task 1.2: Replace bcrypt Usage

- [ ] **1.2.1** Replace RegisterUser hashing: `bcrypt.GenerateFromPassword()` → `hashService.HashPassword(ctx, password)` using LowEntropyRandom
- [ ] **1.2.2** Replace AuthenticateUser verification: `bcrypt.CompareHashAndPassword()` → `hashService.VerifyPassword(ctx, hash, password)`
- [ ] **1.2.3** Remove bcrypt import and `const bcryptCostFactor`
- [ ] **1.2.4** Update comments: bcrypt → LowEntropyRandom/PBKDF2/versioned hash

### Task 1.3: Update Cipher-IM Integration

- [ ] **1.3.1** Inject Hash Service into cipher-im server initialization
- [ ] **1.3.2** Verify cipher-im tests pass
- [ ] **1.3.3** Verify hash format in database: `{3}:PBKDF2-HMAC-SHA256:rounds=600000:...`

### Task 1.4: Verify FIPS Compliance

- [ ] **1.4.1** `grep -r "bcrypt" internal/template/ internal/cipher/` = 0 matches
- [ ] **1.4.2** Run tests: `go test ./internal/template/... ./internal/cipher/...` = all pass
- [ ] **1.4.3** Commit: `feat(template): replace bcrypt with LowEntropyRandom for FIPS compliance`

## PHASE 2: WINDOWS FIREWALL PREVENTION

**Estimated**: 3-5 hours | **HIGH PRIORITY**

### Task 2.1: Add Linter for Test Bind Addresses (STRATEGIC - DO FIRST)

- [ ] **2.1.1** Augment `internal/cmd/cicd/lint_gotest/` with check for `0.0.0.0` in test bind addresses
  - Register as linter in existing `cicd lint-gotest` command
  - Pattern: Reject `"0.0.0.0"` in NewXXXServer(), ServerSettings creation, net.Listen() calls
  - Message: "Use 127.0.0.1 in tests to prevent Windows Firewall prompts"

- [ ] **2.1.2** Test linter on existing violations
  ```bash
  go run ./cmd/cicd lint-gotest ./internal/shared/config/config_coverage_test.go
  # Expected: 2 violations reported (lines 46, 70)
  ```

- [ ] **2.1.3** Commit: `feat(cicd): add lint-gotest check for 0.0.0.0 in test bind addresses`

### Task 2.2: Fix Active Violations

- [ ] **2.2.1** Fix `internal/shared/config/config_coverage_test.go` line 46
  - Change: `NewForJOSEServer("0.0.0.0", 8443, false)` → `NewForJOSEServer("127.0.0.1", 8443, false)`

- [ ] **2.2.2** Fix `internal/shared/config/config_coverage_test.go` line 70
  - Change: `NewForCAServer("0.0.0.0", 9380, false)` → `NewForCAServer("127.0.0.1", 9380, false)`

- [ ] **2.2.3** Verify tests pass and NO firewall prompts
  ```bash
  go test -v ./internal/shared/config/... -run TestNewForJOSEServer
  go test -v ./internal/shared/config/... -run TestNewForCAServer
  ```

- [ ] **2.2.4** Commit: `fix(test): use 127.0.0.1 instead of 0.0.0.0 to prevent Windows Firewall prompts`

### Task 2.3: Verify url_test.go Safety

- [ ] **2.3.1** Confirm `internal/shared/config/url_test.go` only generates URL strings (no server binding)
- [ ] **2.3.2** Verify `grep -r "net.Listen" internal/shared/config/url_test.go` = 0 matches
- [ ] **2.3.3** If safe, document as test data (no changes needed)

### Task 2.4: Root Cause Analysis and Prevention

- [ ] **2.4.1** Scan ALL test executables for bind addresses after full test run
  ```bash
  strings bin/*.test 2>/dev/null | grep "0.0.0.0" || echo "No violations found"
  ```

- [ ] **2.4.2** Add runtime validation in NewTestConfig()
  ```go
  // internal/shared/config/config_test_helper.go
  if bindAddr == "" || bindAddr == "0.0.0.0" {
      panic("CRITICAL: use 127.0.0.1 in tests to prevent Windows Firewall prompts")
  }
  ```

- [ ] **2.4.3** Update anti-patterns documentation
  - File: `.github/instructions/06-02.anti-patterns.instructions.md`
  - Add: Windows Firewall root cause (blank bind address → defaults to 0.0.0.0)
  - Pattern: ALWAYS use NewTestConfig("127.0.0.1", 0, true) in tests

- [ ] **2.4.4** Commit: `docs(anti-patterns): document Windows Firewall root cause and prevention`

## PHASE 3: TEMPLATE LINTING

**Estimated**: 2-3 hours | **MEDIUM PRIORITY**

### Task 3.1: Fix Linting Violations

- [ ] **3.1.1** Run auto-fix: `golangci-lint run --fix ./internal/template/...`
- [ ] **3.1.2** Fix manual violations: errcheck, mnd, nilnil, noctx, unused, wrapcheck
- [ ] **3.1.3** Verify clean: `golangci-lint run ./internal/template/...` = 0 violations
- [ ] **3.1.4** Commit: `fix(lint): resolve 50+ linting violations in template realms`

## PHASE 4: SERVICE TEMPLATE REUSABILITY

**Estimated**: 2-3 hours | **MEDIUM PRIORITY**

### Task 4.1: Document Service Template Patterns

**Migration Order** (per `.github/instructions/02-02.service-template.instructions.md`):
cipher-im ✅ → jose-ja → pki-ca → identity (authz, idp, rs, rp, spa) → sm-kms

- [ ] **4.1.1** Create succinct `docs/SERVICE-TEMPLATE-REUSABILITY.md` (NOT sprawling doc)
  - Realms service pattern (schema lifecycle, tenant isolation, generic interfaces)
  - Barrier service pattern (already in `internal/template/server/barrier/`)
  - Hash Service pattern (extracted to `internal/shared/crypto/hash/`)
  - Telemetry pattern (OTLP integration)
  - Repository patterns (GORM, PostgreSQL/SQLite, test-containers)
  - Test patterns (TestMain, NewTestConfig, t.Cleanup)

- [ ] **4.1.2** Document migration readiness for ALL 9 services
  - FIPS compliance complete ✅
  - Windows Firewall prevention ✅
  - Template linting clean ✅
  - Reference: cipher-im as blueprint for jose-ja, pki-ca, identity services, sm-kms

- [ ] **4.1.3** Commit: `docs(template): document service template reusability for 9-service migration`

---

## PHASE 5: CICD NON-FIPS ALGORITHM LINTER

**Estimated**: 2-3 hours | **HIGH PRIORITY**

### Task 5.1: Augment internal/cmd/cicd/lint_go/ with checkNonFips

- [ ] **5.1.1** Add `checkNonFips` to registeredLinters in `internal/cmd/cicd/lint_go/`
  - Detect: bcrypt, scrypt, Argon2, MD5, SHA-1, DES, 3DES, RSA <2048
  - Pattern: Search for imports and function calls
  - Message: "Non-FIPS algorithm detected - use FIPS-approved algorithms only (see .github/instructions/02-07.cryptography.instructions.md)"

- [ ] **5.1.2** Test on template realms (should catch bcrypt before fix)
  ```bash
  go run ./cmd/cicd lint-go ./internal/template/server/realms/
  # Expected: Violations reported for bcrypt usage
  ```

- [ ] **5.1.3** Integrate into git pre-commit hooks via `.pre-commit-config.yaml`
  ```yaml
  - id: cicd-lint-go
    name: Check Go code for non-FIPS algorithms
    entry: go run ./cmd/cicd lint-go
    language: system
    types: [go]
  ```

- [ ] **5.1.4** Verify rejection at pre-commit time
  - Attempt commit with bcrypt usage → rejected
  - Message shows FIPS-approved alternatives

- [ ] **5.1.5** Commit: `feat(cicd): add checkNonFips linter to detect banned algorithms at pre-commit`

---

## PHASE 6: WINDOWS FIREWALL ROOT CAUSE

**Priority**: ❌ **BLOCKING - MUST COMPLETE**
**Estimated Effort**: 2-4 hours
**Severity**: HIGH - User frustration

### Problem Statement

**User complaint**: "im fucking sick of this shit continuing to happen... many chat sessions"

**Need to find**: WHY does Windows Firewall still prompt despite fixes?

### Task 5.1: Deep Diagnostic Analysis

**BLOCKING**: Find ALL sources of firewall prompts

**Sub-Tasks**:
- [ ] **5.1.1**: Scan ALL test executables for bind addresses
  ```bash
  # After running tests:
  strings bin/*.test | grep "0.0.0.0"
  # Expected: 0 matches
  ```

- [ ] **5.1.2**: Check for dynamic port allocation with wrong bind address
  ```bash
  grep -r "Listen.*:0" internal/**/*_test.go
  grep -r "net.Listen" internal/**/*_test.go
  # Verify ALL use "127.0.0.1:0" pattern
  ```

- [ ] **5.1.3**: Check NewTestConfig() implementation
  ```go
  // internal/shared/config/config_test_helper.go
  func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *ServerSettings {
      // Verify this ALWAYS enforces 127.0.0.1 for tests
      // Verify this REJECTS 0.0.0.0 or blank addresses
  }
  ```

- [ ] **5.1.4**: Add runtime validation
  ```go
  // internal/template/server/listener/listener.go
  func validateTestBindAddress(addr string) error {
      if strings.Contains(addr, "0.0.0.0") {
          return fmt.Errorf("CRITICAL: 0.0.0.0 binding in test environment triggers Windows Firewall prompts - use 127.0.0.1")
      }
      if addr == "" || addr == ":" {
          return fmt.Errorf("CRITICAL: blank bind address defaults to 0.0.0.0 - use 127.0.0.1")
      }
      return nil
  }
  ```

### Task 5.2: Create Comprehensive Prevention Strategy

**BLOCKING**: Prevent future regressions

**Sub-Tasks**:
- [ ] **5.2.1**: Update anti-patterns documentation
  ```markdown
  # docs/cipher-im-migration/WINDOWS-FIREWALL-ROOT-CAUSE.md

  ## Root Cause Analysis

  **Problem**: Windows Firewall prompts when test executables bind to 0.0.0.0

  **Evidence**:
  1. Blank bind addresses (BindPublicAddress="") default to 0.0.0.0
  2. Explicit 0.0.0.0 in test configs (NEVER acceptable)
  3. Dynamic port allocation (:0) without explicit bind address defaults to 0.0.0.0

  **Solution**:
  1. ALWAYS use NewTestConfig() in tests
  2. ALWAYS pass "127.0.0.1" as bind address
  3. NEVER use blank bind addresses
  4. NEVER use 0.0.0.0 in test code
  5. Add runtime validation to reject invalid bind addresses
  ```

- [ ] **5.2.2**: Update copilot instructions with findings
  ```markdown
  # .github/instructions/06-02.anti-patterns.instructions.md

  ### Windows Firewall Root Cause - P0 INCIDENT

  **Root Cause Identified**: Creating ServerSettings with blank BindPublicAddress=""
  causes fmt.Sprintf("%s:%d", "", 0) → ":0" → net.Listen() binds to 0.0.0.0 →
  Windows Firewall exception prompt.

  **NEVER DO**:
  - ❌ Bind to 0.0.0.0 or use blank BindPublicAddress/BindPrivateAddress in tests
  - ❌ Use &ServerSettings{...} with partial field initialization in tests

  **ALWAYS DO**:
  - ✅ Use NewTestConfig(bindAddr, bindPort, devMode)
  - ✅ Use 127.0.0.1 or cryptoutilMagic.IPv4Loopback for test bind addresses
  ```

### Evidence Required for Phase 5 Completion

- ✅ **Code Evidence**:
  - All test bind addresses validated
  - Runtime validation added
  - NewTestConfig() enforces 127.0.0.1

- ✅ **Documentation Evidence**:
  - Root cause analysis document created
  - Anti-patterns instructions updated

- ✅ **Test Evidence**:
  - User confirms: NO firewall prompts after fixes
  - Multiple test runs complete without prompts

- ✅ **Git Evidence**:
  - Commit: "docs(anti-patterns): add Windows Firewall root cause analysis"
  - Commit: "feat(config): add runtime validation for test bind addresses"

---

## PHASE 6: OPTIMIZE TASK ORDER FOR MAXIMUM COMPLETION

**Priority**: ❌ **BLOCKING - MUST COMPLETE**
**Estimated Effort**: 1-2 hours
**Severity**: LOW - Planning optimization

### Task 6.1: Analyze Task Dependencies

**BLOCKING**: Identify optimal execution order

**Sub-Tasks**:
- [ ] **6.1.1**: Map task dependencies
  - Phase 1 (FIPS) → Blocks Phase 4 (JOSE-JA readiness)
  - Phase 2 (Firewall) → Independent
  - Phase 3 (Linting) → Independent
  - Phase 5 (Root Cause) → Builds on Phase 2

- [ ] **6.1.2**: Identify parallelizable work
  - Phase 2 (Firewall) and Phase 3 (Linting) can run in parallel
  - Phase 1 (FIPS) must complete before Phase 4

- [ ] **6.1.3**: Create optimized execution plan
  ```
  1. START: Phase 1 (FIPS) - CRITICAL BLOCKER (8-12 hours)
  2. PARALLEL: Phase 2 (Firewall) + Phase 3 (Linting) (2-4 hours each)
  3. SEQUENTIAL: Phase 5 (Root Cause) builds on Phase 2 (2-4 hours)
  4. FINAL: Phase 4 (Verify Readiness) requires Phase 1 complete (4-8 hours)

  Total Time: 16-30 hours (with parallelization)
  Without Parallelization: 22-38 hours
  Time Saved: 6-8 hours
  ```

### Evidence Required for Phase 6 Completion

- ✅ **Planning Evidence**:
  - Task dependency map created
  - Optimized execution plan documented

- ✅ **Git Evidence**:
  - Commit: "docs(plan): optimize SERVICE-TEMPLATE-v4 task execution order"

---

## PHASE 7: UPDATE TOP-LEVEL COPILOT INSTRUCTIONS

**Priority**: ❌ **BLOCKING - MUST COMPLETE**
**Estimated Effort**: 1-2 hours
**Severity**: HIGH - Prevent future agent mistakes

### Task 7.1: Add CRITICAL Rules to Top-Level Instructions

**BLOCKING**: Prevent bcrypt/firewall regressions

**Sub-Tasks**:
- [ ] **7.1.1**: Add FIPS compliance rule
  ```markdown
  # .github/copilot-instructions.md

  ## CRITICAL RULES (NEVER VIOLATE)

  1. **FIPS Compliance** - ALWAYS check `.github/instructions/02-07.cryptography.instructions.md`
     before implementing password hashing. NEVER use bcrypt (use PBKDF2). ALWAYS verify BANNED
     algorithms list.

  2. **Windows Firewall Prevention** - ALWAYS bind to 127.0.0.1 in tests (NEVER 0.0.0.0).
     ALWAYS use NewTestConfig(). See `.github/instructions/06-02.anti-patterns.instructions.md`.

  3. **Zero Task Skipping** - ALL tasks are HIGHEST PRIORITY. NEVER mark tasks as "optional"
     or "low priority". NEVER skip work. DO NOT STOP until user clicks STOP button.
  ```

- [ ] **7.1.2**: Add pre-implementation checklist
  ```markdown
  # .github/copilot-instructions.md

  ## Before Implementing Any Cryptographic Feature

  **MANDATORY CHECKLIST**:
  - [ ] Read `.github/instructions/02-07.cryptography.instructions.md` (FIPS requirements)
  - [ ] Read `.github/instructions/02-08.hashes.instructions.md` (Hash Service architecture)
  - [ ] Verify algorithm is in APPROVED list (NOT BANNED list)
  - [ ] Check if Hash Service already implements pattern
  - [ ] Verify pepper requirement for passwords
  - [ ] Confirm output format matches versioned pattern
  ```

- [ ] **7.1.3**: Add post-implementation verification
  ```markdown
  # .github/copilot-instructions.md

  ## After Implementing Any Feature

  **MANDATORY VERIFICATION**:
  - [ ] Run `grep -r "bcrypt\|scrypt\|argon2" .` = 0 matches
  - [ ] Run `grep -r "0\.0\.0\.0.*test" .` = 0 matches in bind addresses
  - [ ] Run `golangci-lint run ./...` = 0 violations
  - [ ] Run all tests = 0 failures
  - [ ] Commit immediately with conventional commit message
  ```

### Evidence Required for Phase 7 Completion

- ✅ **Documentation Evidence**:
  - CRITICAL RULES added to top-level instructions
  - Pre-implementation checklist added
  - Post-implementation verification added

- ✅ **Git Evidence**:
  - Commit: "docs(instructions): add CRITICAL rules to prevent bcrypt/firewall regressions"

---

## EXECUTION STRATEGY

### Immediate Actions (DO NOW)

1. ✅ **Read this document completely** - Understand all tasks
2. 🔄 **Start Phase 1 (FIPS)** - CRITICAL BLOCKER, highest priority
3. 🔄 **Parallel: Start Phase 2 (Firewall) + Phase 3 (Linting)** - Independent work
4. 🔄 **Sequential: Phase 5 (Root Cause)** - After Phase 2 complete
5. 🔄 **Final: Phase 4 (Verify Readiness)** - After Phase 1 complete
6. 🔄 **Commit frequently** - After each task/subtask completion
7. 🔄 **Update this document** - Mark tasks complete with ✅ as work progresses

### Continuous Work Philosophy

**PER USER DIRECTIVE**:
- ❌ DO NOT STOP until all work complete
- ❌ DO NOT ASK if user wants to continue
- ❌ DO NOT SKIP any tasks
- ❌ DO NOT DEPRIORITIZE any tasks
- ❌ DO NOT OMIT any tasks
- ✅ ALL WORK IS HIGHEST PRIORITY AND BLOCKING
- ✅ WORK AUTONOMOUSLY UNTIL COMPLETION
- ✅ COMMIT FREQUENTLY TO SHOW PROGRESS

### Evidence-Based Completion

**EVERY task MUST have**:
- ✅ Code evidence (files changed, tests passing, linting clean)
- ✅ Test evidence (coverage maintained, no failures)
- ✅ Git evidence (conventional commits pushed)

**NO task marked complete without evidence**

---

## SUCCESS CRITERIA

### Phase-Level Success

- ✅ **Phase 1**: Zero bcrypt usage, PBKDF2 implemented, Hash Service integrated, pepper configured
- ✅ **Phase 2**: Zero `0.0.0.0` in test bind addresses, user confirms no firewall prompts
- ✅ **Phase 3**: Zero linting violations in template
- ✅ **Phase 4**: JOSE-JA migration plan complete, sm-kms reusability verified
- ✅ **Phase 5**: Root cause documented, prevention strategy implemented
- ✅ **Phase 6**: Optimized task order documented
- ✅ **Phase 7**: CRITICAL rules added to top-level instructions

### Overall Success

- ✅ **ALL phases complete**
- ✅ **ALL tasks complete**
- ✅ **ALL tests passing**
- ✅ **ALL linting clean**
- ✅ **ALL commits pushed**
- ✅ **User confirms: No bcrypt, no firewall prompts, all work done**

---

## LESSONS LEARNED (To Be Updated)

**Mistakes That Led to v4**:
1. Agent created template realms without checking FIPS requirements
2. Agent used bcrypt (convenient) instead of PBKDF2 (compliant)
3. Agent marked work as "optional" when user wanted ALL work
4. Agent didn't find Windows Firewall root cause despite multiple sessions
5. Agent stopped after partial completion instead of ALL work

**Prevention for Future**:
1. ALWAYS check copilot instructions BEFORE implementing crypto
2. ALWAYS verify FIPS compliance BEFORE marking work complete
3. NEVER mark tasks as "optional" or "low priority"
4. ALWAYS find root causes (not just symptoms)
5. ALWAYS work until ALL tasks complete (not partial)

---

## FINAL COMMIT MESSAGE TEMPLATE

```
feat(template): complete SERVICE-TEMPLATE-v4 remediation

CRITICAL FIXES:
- Replace bcrypt with PBKDF2 for FIPS-140-2/3 compliance
- Integrate Hash Service with version-based policy framework
- Implement MANDATORY pepper requirement (OWASP)
- Fix Windows Firewall triggers (0.0.0.0 → 127.0.0.1)
- Resolve 50+ template linting violations
- Complete v3 incomplete work
- Document root cause analyses
- Update top-level copilot instructions

EVIDENCE:
- All tests passing (cipher-im, template, shared)
- Zero linting violations
- Zero FIPS violations
- User confirms: No firewall prompts
- All phases complete with evidence

BREAKING CHANGES:
- Password hash format changed from bcrypt to PBKDF2 versioned format
- Requires pepper secret configuration
- Requires Hash Service initialization

Closes: (GitHub issue numbers if applicable)
```

---

**DOCUMENT STATUS**: ✅ **READY FOR EXECUTION - ALL TASKS HIGHEST PRIORITY**

**NEXT IMMEDIATE ACTION**: Start Phase 1 (FIPS Compliance) - bcrypt → PBKDF2 remediation
