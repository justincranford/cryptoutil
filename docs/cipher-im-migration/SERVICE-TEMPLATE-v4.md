# SERVICE-TEMPLATE v4 - Complete Remediation Plan

**Created**: 2026-01-02
**Status**: ACTIVE - ALL TASKS HIGHEST PRIORITY AND BLOCKING
**Previous**: SERVICE-TEMPLATE-v3.md (95% complete but WITH CRITICAL VIOLATIONS)

---

## EXECUTIVE SUMMARY

### CRITICAL VIOLATIONS DISCOVERED IN PHASE 7

**Phase 7 Status**: ❌ **COMPLETE BUT WITH FIPS VIOLATIONS**
- ✅ Template realms service created (5 files, 694 lines)
- ✅ Cipher-IM migrated to template realms
- ❌ **bcrypt used instead of PBKDF2** (16 instances) - **FIPS-140-2/3 VIOLATION**
- ❌ **No Hash Service integration** - **CRITICAL**
- ❌ **No pepper implementation** - **MANDATORY OWASP REQUIREMENT**
- ❌ **Wrong hash output format** - **CRITICAL**
- ❌ **Windows Firewall triggers remain** (11 instances of `0.0.0.0`)
- ❌ **Template linting violations** (50+ issues)

### USER FRUSTRATIONS ADDRESSED

1. **bcrypt Violation** - "FUCKING bcrypt IS BACK AGAIN FOR THE FUCKING 10th TIME"
   - Root Cause: Agent missed FIPS compliance check during template creation
   - Impact: Template realms unusable for production, JOSE-JA migration blocked
   - Fix: Complete rewrite of password hashing using PBKDF2 + Hash Service

2. **Windows Firewall** - "im fucking sick of this shit continuing to happen"
   - Root Cause: Some tests still use `0.0.0.0` instead of `127.0.0.1`
   - Impact: User clicks "Cancel" on every firewall prompt, test quality degraded
   - Fix: Change all test bindings to `127.0.0.1`, verify no firewall prompts

3. **Work Skipping** - "STOP FUCKING SKIPPING OR DEPRIORITIZING TASKS"
   - Root Cause: v3 marked tasks as "optional" or "low priority"
   - Impact: Critical work incomplete, violations introduced
   - Fix: ALL tasks in v4 marked HIGHEST PRIORITY AND BLOCKING

### COMPLETION PHILOSOPHY FOR v4

**ALL WORK IS HIGHEST PRIORITY AND BLOCKING FOR NEXT CHAT SESSION**

- ❌ NO tasks marked "optional"
- ❌ NO tasks marked "low priority"
- ❌ NO tasks skipped
- ❌ NO tasks deprioritized
- ❌ NO tasks omitted
- ✅ **ALL tasks MUST be completed**
- ✅ **DO NOT STOP until all work done**
- ✅ **DO NOT ASK for permission to continue**

---

## PHASE 1: CRITICAL FIPS COMPLIANCE - BCRYPT → PBKDF2

**Priority**: ❌ **BLOCKING - MUST COMPLETE BEFORE ALL OTHER WORK**
**Estimated Effort**: 8-12 hours
**Severity**: CRITICAL - Production compliance violation

### Problem Statement

**Files Affected**:
- `internal/template/server/realms/service.go` (12 instances)
- `internal/template/server/realms/interfaces.go` (4 instances)
- `internal/cipher/domain/user.go` (inherits violation)
- `internal/cipher/repository/user_repository_adapter.go` (inherits violation)

**Violations**:
1. ❌ Import: `"golang.org/x/crypto/bcrypt"` (line 12 of service.go)
2. ❌ Usage: `bcrypt.GenerateFromPassword()` (line 101)
3. ❌ Usage: `bcrypt.CompareHashAndPassword()` (line 161)
4. ❌ Cost: `const bcryptCostFactor = 10` (line 99)
5. ❌ FALSE CLAIM: Comment says "FIPS-compliant via PBKDF2 fallback" - **CODE USES BCRYPT**

**FIPS Requirements** (from `.github/instructions/02-07.cryptography.instructions.md`):
```
### BANNED Algorithms

❌ bcrypt, scrypt, Argon2 (use PBKDF2) | MD5, SHA-1 (use SHA-256+) | RSA <2048 | DES, 3DES
```

### Task 1.1: Integrate Hash Service into Template Realms

**BLOCKING**: Create or import Hash Service implementation

**Sub-Tasks**:
- [ ] **1.1.1**: Find sm-kms Hash Service implementation
  - Search for `internal/kms/**/hash*.go` files
  - Read existing PBKDF2 implementation patterns
  - Identify reusable components (registries, version framework, pepper handling)

- [ ] **1.1.2**: Extract Hash Service to `internal/shared/crypto/hash/`
  - Move registries: LowEntropyDeterministic, LowEntropyRandom, HighEntropyDeterministic, HighEntropyRandom
  - Move version framework (v1/v2/v3 support)
  - Move pepper handling (Docker secrets, config file)
  - Move hash format parsing/generation: `{version}:{algorithm}:{iterations}:base64(salt):base64(hash)}`
  - Add comprehensive tests

- [ ] **1.1.3**: Create template realms Hash Service integration
  - Inject Hash Service into UserServiceImpl constructor
  - Configure for password hashing (LowEntropyRandomHashRegistry)
  - Set current version (v3 = 2025 OWASP)
  - Configure pepper from Docker secret
  - Add integration tests

**Expected Outcome**:
```go
// internal/template/server/realms/service.go
import (
    cryptoutilHash "cryptoutil/internal/shared/crypto/hash"
)

type UserServiceImpl struct {
    userRepo    UserRepository
    userFactory func() UserModel
    hashService *cryptoutilHash.Service  // NEW - Hash Service integration
}

func NewUserService(
    userRepo UserRepository,
    userFactory func() UserModel,
    hashService *cryptoutilHash.Service,  // NEW - Inject Hash Service
) *UserServiceImpl {
    return &UserServiceImpl{
        userRepo:    userRepo,
        userFactory: userFactory,
        hashService: hashService,
    }
}
```

### Task 1.2: Replace bcrypt with PBKDF2

**BLOCKING**: Remove all bcrypt usage

**Sub-Tasks**:
- [ ] **1.2.1**: Replace RegisterUser password hashing
  ```go
  // OLD (line 99-101):
  const bcryptCostFactor = 10
  passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCostFactor)
  
  // NEW:
  passwordHash, err := s.hashService.HashPassword(ctx, password)
  // Returns: {3}:PBKDF2-HMAC-SHA256:rounds=600000:base64(randomSalt):base64(hash)
  ```

- [ ] **1.2.2**: Replace AuthenticateUser password verification
  ```go
  // OLD (line 161):
  err = bcrypt.CompareHashAndPassword([]byte(user.GetPasswordHash()), []byte(password))
  
  // NEW:
  valid, err := s.hashService.VerifyPassword(ctx, user.GetPasswordHash(), password)
  if err != nil || !valid {
      return nil, fmt.Errorf("invalid credentials")
  }
  ```

- [ ] **1.2.3**: Remove bcrypt import
  ```go
  // DELETE line 12:
  import "golang.org/x/crypto/bcrypt"
  ```

- [ ] **1.2.4**: Update all comments referencing bcrypt
  - Line 15: "implements user registration and authentication using **PBKDF2**"
  - Line 39: "Hash password using **PBKDF2** (OWASP 2025 recommendations)"
  - Line 49: "Password hashed with **PBKDF2** (FIPS-compliant, NIST-approved)"
  - Line 99: Delete `const bcryptCostFactor` entirely
  - Line 193: "returns the user's **PBKDF2** password hash"

- [ ] **1.2.5**: Update interfaces.go comments
  - Line 15: "**PBKDF2** password hashing (FIPS-140-2/3 compliant)"
  - Line 60: Remove BcryptCost field example
  - Line 112: "GetPasswordHash returns the **versioned PBKDF2** hash"
  - Line 124: "SetPasswordHash sets the **versioned PBKDF2** hash"

### Task 1.3: Add Pepper Support

**BLOCKING**: Implement MANDATORY OWASP pepper requirement

**Sub-Tasks**:
- [ ] **1.3.1**: Add pepper configuration to template
  ```yaml
  # configs/test/cryptoutil-common.yml
  hash_service:
    current_version: 3  # 2025 OWASP
    pepper_secret: file:///run/secrets/hash_pepper_v3
  ```

- [ ] **1.3.2**: Create Docker secret for pepper
  ```yaml
  # deployments/compose/compose.yml
  secrets:
    hash_pepper_v3:
      file: ./secrets/hash_pepper_v3.secret
  ```

- [ ] **1.3.3**: Generate secure pepper (32 bytes random)
  ```bash
  openssl rand -base64 32 > deployments/compose/secrets/hash_pepper_v3.secret
  chmod 440 deployments/compose/secrets/hash_pepper_v3.secret
  ```

- [ ] **1.3.4**: Load pepper in Hash Service initialization
  ```go
  pepperPath := viper.GetString("hash_service.pepper_secret")
  pepperBytes, err := loadSecret(pepperPath)  // Handles file:// prefix
  if err != nil {
      return nil, fmt.Errorf("failed to load pepper: %w", err)
  }
  ```

### Task 1.4: Verify FIPS Compliance

**BLOCKING**: Confirm no FIPS violations remain

**Sub-Tasks**:
- [ ] **1.4.1**: Grep for banned algorithms
  ```bash
  grep -r "bcrypt\|scrypt\|argon2\|MD5\|SHA-1" internal/template/ internal/cipher/
  # Expected: 0 matches (except in comments explaining why NOT to use them)
  ```

- [ ] **1.4.2**: Verify Hash Service uses PBKDF2
  ```bash
  grep -r "PBKDF2" internal/shared/crypto/hash/
  # Expected: Multiple matches in implementation
  ```

- [ ] **1.4.3**: Verify pepper is loaded from secrets
  ```bash
  grep -r "loadSecret.*pepper" internal/template/ internal/cipher/
  # Expected: Matches in Hash Service initialization
  ```

- [ ] **1.4.4**: Run FIPS compliance tests
  ```bash
  go test -v ./internal/shared/crypto/hash/... -run TestFIPSCompliance
  go test -v ./internal/template/server/realms/... -run TestPasswordHashing
  ```

### Task 1.5: Update Cipher-IM Integration

**BLOCKING**: Ensure cipher-im uses PBKDF2 (not bcrypt)

**Sub-Tasks**:
- [ ] **1.5.1**: Update cipher-im server to inject Hash Service
  ```go
  // internal/cipher/server/server.go
  hashService, err := cryptoutilHash.NewService(hashConfig)
  if err != nil {
      return nil, fmt.Errorf("failed to create hash service: %w", err)
  }
  
  userService := cryptoutilRealms.NewUserService(
      userRepo,
      func() cryptoutilRealms.UserModel { return &domain.User{} },
      hashService,  // NEW - Inject Hash Service
  )
  ```

- [ ] **1.5.2**: Verify cipher-im tests pass with PBKDF2
  ```bash
  go test -v ./internal/cipher/... -run TestRegistration
  go test -v ./internal/cipher/... -run TestAuthentication
  go test -v ./internal/cipher/e2e/... -run TestUserFlow
  ```

- [ ] **1.5.3**: Verify hash format in database
  ```bash
  # After creating user in cipher-im:
  # SELECT password_hash FROM cipher_users LIMIT 1;
  # Expected: {3}:PBKDF2-HMAC-SHA256:rounds=600000:abc123...:def456...
  ```

### Evidence Required for Phase 1 Completion

- ✅ **Code Evidence**:
  - `grep -r bcrypt internal/template/ internal/cipher/` = 0 matches
  - `grep -r PBKDF2 internal/template/ internal/cipher/` > 0 matches
  - Hash Service extracted to `internal/shared/crypto/hash/`
  - Pepper loaded from Docker secret

- ✅ **Test Evidence**:
  - All template realms tests pass
  - All cipher-im tests pass
  - E2E tests verify user registration/authentication works

- ✅ **Git Evidence**:
  - Commit: "feat(template): replace bcrypt with PBKDF2 for FIPS compliance"
  - Commit: "feat(shared): extract Hash Service from kms to shared"
  - Commit: "fix(cipher-im): integrate PBKDF2 Hash Service"

---

## PHASE 2: WINDOWS FIREWALL PREVENTION

**Priority**: ❌ **BLOCKING - MUST COMPLETE**
**Estimated Effort**: 2-4 hours
**Severity**: HIGH - User extremely frustrated

### Problem Statement

**11 instances of `0.0.0.0` found in test files**:
- `internal/shared/config/config_coverage_test.go` (lines 46, 70) - **ACTIVE VIOLATIONS**
- `internal/shared/config/url_test.go` (lines 78-84) - **TEST CASE DATA (likely safe)**
- `internal/shared/crypto/certificate/certificates_test.go` (lines 91, 147) - **COMMENTS ONLY (safe)**
- `internal/shared/config/config_test.go` (line 119) - **CIDR TEST DATA (safe)**

**Root Cause**: `NewForJOSEServer("0.0.0.0", 8443, false)` triggers Windows Firewall prompts

### Task 2.1: Fix config_coverage_test.go Violations

**BLOCKING**: Active `0.0.0.0` usage

**Sub-Tasks**:
- [ ] **2.1.1**: Fix line 46 (JOSE server)
  ```go
  // OLD:
  settings := NewForJOSEServer("0.0.0.0", 8443, false)
  
  // NEW:
  settings := NewForJOSEServer("127.0.0.1", 8443, false)
  ```

- [ ] **2.1.2**: Fix line 70 (CA server)
  ```go
  // OLD:
  settings := NewForCAServer("0.0.0.0", 9380, false)
  
  // NEW:
  settings := NewForCAServer("127.0.0.1", 9380, false)
  ```

- [ ] **2.1.3**: Run test to verify no firewall prompts
  ```bash
  go test -v ./internal/shared/config/... -run TestNewForJOSEServer
  go test -v ./internal/shared/config/... -run TestNewForCAServer
  # User verifies: NO Windows Firewall prompts appear
  ```

### Task 2.2: Analyze url_test.go for Server Creation

**BLOCKING**: Verify test case data doesn't create servers

**Sub-Tasks**:
- [ ] **2.2.1**: Read full url_test.go (121 lines)
  - Confirm tests only build URL strings
  - Verify no `net.Listen()` calls
  - Check for server creation in table tests

- [ ] **2.2.2**: If NO server creation, mark as SAFE
  - ✅ URL string generation with `0.0.0.0` = OK (no firewall trigger)
  - ❌ Server binding to `0.0.0.0` = NOT OK (triggers firewall)

- [ ] **2.2.3**: If server creation found, fix immediately
  ```go
  // Replace 0.0.0.0 with 127.0.0.1 in test cases that create servers
  ```

### Task 2.3: Add Validation to Prevent Future Regressions

**BLOCKING**: Enforce `127.0.0.1` pattern

**Sub-Tasks**:
- [ ] **2.3.1**: Add pre-commit hook check
  ```bash
  # .pre-commit-config.yaml
  - id: check-test-bind-addresses
    name: Check test bind addresses use 127.0.0.1
    entry: bash -c 'grep -r "0\.0\.0\.0" **/*_test.go && exit 1 || exit 0'
    language: system
  ```

- [ ] **2.3.2**: Add golangci-lint custom rule
  ```yaml
  # .golangci.yml
  linters-settings:
    custom:
      test-bind-address:
        path: ./cmd/cicd/lint-test-bind-address
        description: Enforce 127.0.0.1 in test bind addresses
  ```

- [ ] **2.3.3**: Document pattern in copilot instructions
  ```markdown
  # .github/instructions/06-02.anti-patterns.instructions.md
  
  ## Windows Firewall Prevention - CRITICAL
  
  **ALWAYS bind to 127.0.0.1 in tests (NEVER 0.0.0.0)**
  
  - ❌ `NewForJOSEServer("0.0.0.0", 8443, false)` - TRIGGERS FIREWALL
  - ✅ `NewForJOSEServer("127.0.0.1", 8443, false)` - NO FIREWALL
  ```

### Evidence Required for Phase 2 Completion

- ✅ **Code Evidence**:
  - `grep -r "0\.0\.0\.0.*Server" **/*_test.go` = 0 matches in server creation
  - All test bind addresses use `127.0.0.1`

- ✅ **Test Evidence**:
  - All config tests pass
  - User confirms: NO Windows Firewall prompts during test runs

- ✅ **Git Evidence**:
  - Commit: "fix(test): use 127.0.0.1 instead of 0.0.0.0 to prevent Windows Firewall prompts"

---

## PHASE 3: TEMPLATE LINTING FIXES

**Priority**: ❌ **BLOCKING - MUST COMPLETE**
**Estimated Effort**: 2-4 hours
**Severity**: MEDIUM - Code quality

### Problem Statement

**50+ linting violations in template realms** (from Phase 7.4):
- errcheck: Unchecked errors
- mnd: Magic number detector
- nilnil: Return nil, nil pattern
- noctx: Missing context
- unused: Unused variables/functions
- wrapcheck: Unwrapped errors
- wsl_v5: Whitespace linter

### Task 3.1: Fix All Linting Violations

**BLOCKING**: Zero linting errors required

**Sub-Tasks**:
- [ ] **3.1.1**: Run golangci-lint on template
  ```bash
  golangci-lint run --fix ./internal/template/...
  # Let auto-fix handle: gofmt, gofumpt, goimports, godot, wsl_v5
  ```

- [ ] **3.1.2**: Fix manual violations
  ```bash
  golangci-lint run ./internal/template/...
  # Address: errcheck, mnd, nilnil, noctx, unused, wrapcheck
  ```

- [ ] **3.1.3**: Verify clean output
  ```bash
  golangci-lint run ./internal/template/...
  # Expected: No violations
  ```

### Evidence Required for Phase 3 Completion

- ✅ **Code Evidence**:
  - `golangci-lint run ./internal/template/...` = 0 violations

- ✅ **Git Evidence**:
  - Commit: "fix(lint): resolve 50+ linting violations in template realms"

---

## PHASE 4: COMPLETE v3 INCOMPLETE WORK

**Priority**: ❌ **BLOCKING - MUST COMPLETE**
**Estimated Effort**: 4-8 hours
**Severity**: MEDIUM - Unfinished work

### Task 4.1: Extract Realms Service Patterns

**From v3 REALMS SERVICE EXTRACTION section**

**Sub-Tasks**:
- [ ] **4.1.1**: Document realms service reusability
  - How cipher-im user realm pattern can be generalized
  - How JOSE-JA OAuth realm will use same pattern
  - How Identity authentication realm will use same pattern
  - Schema lifecycle management (CREATE SCHEMA, DROP SCHEMA CASCADE)
  - Tenant isolation middleware (search_path setting)

- [ ] **4.1.2**: Create realms service design document
  - Generic RealmService interface
  - Generic RealmRepository interface
  - TenantIsolationMiddleware pattern
  - Product-specific integration patterns

### Task 4.2: Verify JOSE-JA Migration Readiness

**BLOCKING**: Confirm template ready for JOSE-JA

**Sub-Tasks**:
- [ ] **4.2.1**: Check FIPS compliance complete
  - ✅ bcrypt removed
  - ✅ PBKDF2 implemented
  - ✅ Hash Service integrated
  - ✅ Pepper configured

- [ ] **4.2.2**: Check Windows Firewall fixes complete
  - ✅ All test bindings use `127.0.0.1`
  - ✅ Validation prevents future regressions

- [ ] **4.2.3**: Check template linting clean
  - ✅ Zero linting violations
  - ✅ All tests passing

- [ ] **4.2.4**: Document JOSE-JA migration plan
  - Use cipher-im as blueprint
  - OAuth realm schema requirements
  - JWK/JWS/JWE integration with barrier service
  - Test migration patterns (TestMain, NewTestConfig, t.Cleanup)

### Task 4.3: Verify sm-kms Reusability Patterns

**BLOCKING**: Confirm template offers all reusable parts from sm-kms

**Sub-Tasks**:
- [ ] **4.3.1**: Verify barrier service reuse
  - ✅ Already extracted to `internal/template/server/barrier/`
  - ✅ Cipher-IM using it successfully
  - ✅ JOSE-JA can use same pattern

- [ ] **4.3.2**: Verify telemetry service reuse
  - ✅ Already extracted to observability patterns
  - ✅ OTLP integration standardized
  - ✅ All services use same pattern

- [ ] **4.3.3**: Verify repository patterns reuse
  - ✅ GORM patterns standardized
  - ✅ PostgreSQL/SQLite dual support
  - ✅ Test-containers pattern established

- [ ] **4.3.4**: Verify Hash Service extraction (NEW - from Phase 1)
  - ✅ Extracted to `internal/shared/crypto/hash/`
  - ✅ Version-based policy framework
  - ✅ Four registries (LowEntropyDeterministic, LowEntropyRandom, HighEntropyDeterministic, HighEntropyRandom)
  - ✅ Pepper support

### Evidence Required for Phase 4 Completion

- ✅ **Documentation Evidence**:
  - Realms service design document created
  - JOSE-JA migration plan documented
  - sm-kms reusability patterns verified

- ✅ **Git Evidence**:
  - Commit: "docs(template): document realms service reusability patterns"
  - Commit: "docs(jose): create JOSE-JA migration plan"

---

## PHASE 5: WINDOWS FIREWALL ROOT CAUSE ANALYSIS

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
