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
  - Ensure registries: LowEntropyDeterministic, LowEntropyRandom, HighEntropyDeterministic, HighEntropyRandom
  - Ensure version framework (v1/v2/v3 support)
  - Ensure hash format: `{version}:{algorithm}:{iterations}:base64(salt):base64(hash)}`

- [ ] **1.1.2** Inject Hash Service into service-template realms UserServiceImpl for reuse by ALL 9 services
  ```go
  import cryptoutilHash "cryptoutil/internal/shared/crypto/hash"
  type UserServiceImpl struct {
      hashService *cryptoutilHash.Service
  }
  ```
  **Note**: LowEntropyRandom registry already configured with v3 (2025 OWASP), PBKDF2-HMAC-SHA256, versioned format

### Task 1.2: Replace bcrypt Usage in Service-Template

- [ ] **1.2.1** Replace RegisterUser hashing: `bcrypt.GenerateFromPassword()` → `hashService.HashPassword(ctx, password)` using LowEntropyRandom
- [ ] **1.2.2** Replace AuthenticateUser verification: `bcrypt.CompareHashAndPassword()` → `hashService.VerifyPassword(ctx, hash, password)`
- [ ] **1.2.3** Remove bcrypt import and `const bcryptCostFactor`
- [ ] **1.2.4** Update comments: bcrypt → LowEntropyRandom/PBKDF2/versioned hash

### Task 1.3: Verify Cipher-IM Integration (Indirect via Realms)

**CRITICAL**: Cipher-im uses Hash Service INDIRECTLY via service-template username/password realms (NOT direct injection). All 9 services MUST use realms for registration, authentication, authorization.

- [ ] **1.3.1** Verify cipher-im uses service-template username/password realm (file or DB type) for user operations
- [ ] **1.3.2** Verify cipher-im tests pass (uses realm methods, NOT direct Hash Service calls)
- [ ] **1.3.3** Verify hash format in database: `{3}:PBKDF2-HMAC-SHA256:rounds=600000:...`

### Task 1.4: Verify FIPS Compliance

- [ ] **1.4.1** `grep -r "bcrypt" internal/template/ internal/cipher/` = 0 matches
- [ ] **1.4.2** Run tests: `go test ./internal/template/... ./internal/cipher/...` = all pass
- [ ] **1.4.3** Commit: `feat(template): replace bcrypt with LowEntropyRandom for FIPS compliance`

## PHASE 2: WINDOWS FIREWALL PREVENTION

**Estimated**: 3-5 hours | **HIGH PRIORITY**

### Task 2.1: Add Linter for Test Bind Addresses (STRATEGIC - DO FIRST)

- [ ] **2.1.1** Augment `internal/cmd/cicd/lint_gotest/` with check for `0.0.0.0` AND empty bind addresses
  - Register as linter in existing `cicd lint-gotest` command
  - Pattern: Reject `"0.0.0.0"` AND `""` (blank) in:
    * NewXXXServer() calls with bind address arguments
    * ServerSettings struct initialization (partial or full)
    * net.Listen() calls with address parameters
    * BindPublicAddress/BindPrivateAddress field assignments
  - Message: "CRITICAL: don't use 0.0.0.0 or blank bind address in tests, use 127.0.0.1 to prevent Windows Firewall exception prompts"
  - Create comprehensive tests in `internal/cmd/cicd/lint_gotest/` subdirectory for all patterns:
    * Direct `"0.0.0.0"` usage
    * Blank `""` usage (defaults to 0.0.0.0)
    * Struct literal with blank fields
    * Variable assignments
    * Function call arguments

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
  **Detection**: Watch for Windows Security Alert dialog "Windows Defender Firewall has blocked some features". If prompt appears, tests are still binding to 0.0.0.0 or blank address.

- [ ] **2.2.4** Commit: `fix(test): use 127.0.0.1 instead of 0.0.0.0 to prevent Windows Firewall prompts`

### Task 2.3: Verify url_test.go Safety AND Add Bind Address Validation Coverage

- [ ] **2.3.1** Confirm `internal/shared/config/url_test.go` only generates URL strings (no server binding)
- [ ] **2.3.2** Verify `grep -r "net.Listen" internal/shared/config/url_test.go` = 0 matches
- [ ] **2.3.3** Add test coverage to `internal/shared/config/url_test.go` for detecting/rejecting blank or 0.0.0.0 bind addresses
  - Test validateBindAddress() helper for public listener
  - Test validateBindAddress() helper for private listener
  - Verify rejection of blank bind addresses
  - Verify rejection of 0.0.0.0 bind addresses
  - Verify acceptance of 127.0.0.1 bind addresses

### Task 2.4: Root Cause Analysis and Prevention

- [ ] **2.4.1** Scan ALL test executables for bind addresses after full test run (0.0.0.0 AND empty strings)
  ```bash
  strings bin/*.test 2>/dev/null | grep "0.0.0.0" || echo "No 0.0.0.0 violations found"
  # Also check for empty bind address patterns (harder to detect via strings)
  grep -r 'BindPublicAddress.*""' **/*_test.go
  grep -r 'BindPrivateAddress.*""' **/*_test.go
  ```

- [ ] **2.4.2** Add runtime validation in NewTestConfig()
  ```go
  // internal/shared/config/config_test_helper.go
  if bindAddr == "" || bindAddr == "0.0.0.0" {
      panic("CRITICAL: don't use 0.0.0.0 or blank bind address in tests, use 127.0.0.1 to prevent Windows Firewall exception prompts")
  }
  ```

- [ ] **2.4.3** Update anti-patterns documentation
  - File: `.github/instructions/06-02.anti-patterns.instructions.md`
  - Add: Windows Firewall root cause (blank bind address → defaults to 0.0.0.0)
  - Pattern: ALWAYS use NewTestConfig("127.0.0.1", 0, true) in tests

- [ ] **2.4.4** Commit: `docs(anti-patterns): document Windows Firewall root cause and prevention`

- [ ] **2.4.5** Add requires checks in NewTestConfig to reject bind values NOT equal to 127.0.0.1
  ```go
  // internal/shared/config/config_test_helper.go
  func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *ServerSettings {
      if bindAddr != "127.0.0.1" {
          panic(fmt.Sprintf("CRITICAL: bind address must be 127.0.0.1 in tests (got %q), prevents Windows Firewall prompts", bindAddr))
      }
      // ... rest of implementation
  }
  ```
  **Rationale**: Fail fast to reveal ANY callers that were missed during 127.0.0.1 migration

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
  - Hash Service pattern (extracted to `internal/shared/crypto/hash/`, used INDIRECTLY via realms)
    * ALL 9 services MUST use Hash Service via username/password realms (file/DB types)
    * Also via username/email realms, magic link realms, random OTP realms, etc.
    * NEVER direct Hash Service injection into services (violates reusability)
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
  - Detect: bcrypt, scrypt, Argon2, MD5, SHA-1, SHA-224 (weak), DES, 3DES, RC4, RC2, EC P-224 (weak), RSA <2048, DSA (any size)
  - Pattern: Search for imports and function calls
  - Message: "Non-FIPS algorithm detected - use FIPS-approved algorithms only (see .github/instructions/02-07.cryptography.instructions.md)"
  - Approved alternatives: SHA-256/384/512, RSA ≥2048, ECDSA P-256/384/521, EdDSA, AES ≥128, PBKDF2, HKDF

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

## PHASE 6: WINDOWS FIREWALL ROOT CAUSE PREVENTION

**Estimated**: 2-4 hours | **HIGH PRIORITY**

### Task 6.1: Research Additional Firewall Trigger Use Cases

- [ ] **6.1.1** Search Go documentation for network binding patterns that trigger OS firewall prompts
  - net.Listen() variants
  - http.Server binding patterns
  - UDP socket binding
  - Raw socket creation

- [ ] **6.1.2** Internet search for "Windows Firewall prompt Go testing" and related queries
  - Stack Overflow discussions
  - GitHub issues in Go repos
  - Windows developer documentation

- [ ] **6.1.3** AI analysis of Go network code patterns that may trigger firewall prompts
  - Multicast group joining
  - Network interface enumeration with binding
  - IPv6 wildcard binding (::)

### Task 6.2: Deep Diagnostic Analysis

- [ ] **6.2.1** Scan ALL test executables for bind addresses
  ```bash
  # After running tests:
  strings bin/*.test | grep "0.0.0.0"
  # Expected: 0 matches
  ```

- [ ] **6.2.2** Check for dynamic port allocation with wrong bind address
  ```bash
  grep -r "Listen.*:0" internal/**/*_test.go
  grep -r "net.Listen" internal/**/*_test.go
  # Verify ALL use "127.0.0.1:0" pattern
  ```

- [ ] **6.2.3** Check NewTestConfig() implementation
  ```go
  // internal/shared/config/config_test_helper.go
  func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *ServerSettings {
      // Verify this ALWAYS enforces 127.0.0.1 for tests
      // Verify this REJECTS 0.0.0.0 or blank addresses
  }
  ```

- [ ] **6.2.4** Add runtime validation
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

### Task 6.3: Create Comprehensive Prevention Strategy

- [ ] **6.3.1** Update anti-patterns documentation
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

- [ ] **6.3.2** Add runtime validation in NewTestConfig
  ```go
  func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *ServerSettings {
      if bindAddr == "0.0.0.0" || bindAddr == "" {
          panic("NEVER bind to 0.0.0.0 in tests - use 127.0.0.1")
      }
      // ... rest of implementation
  }
  ```

- [ ] **6.3.3** Update `.github/instructions/06-02.anti-patterns.instructions.md`
  - Add "Windows Firewall Exception Prevention" section
  - Document: `&ServerSettings{}` partial initialization pattern is UNSAFE
  - Document: ALWAYS use `NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)`

- [ ] **6.3.4** Commit: `docs(anti-patterns): document Windows Firewall root cause and prevention`

---

## PHASE 7: PEPPER IMPLEMENTATION (STRATEGIC - END)

**Estimated**: 2-3 hours | **CRITICAL** (but strategically last)

**Rationale**: Other fixes needed first, but pepper is MANDATORY OWASP requirement

### Task 7.1: Add Pepper Configuration

- [ ] **7.1.1** Add pepper config to `configs/test/cryptoutil-common.yml`
  ```yaml
  hash_service:
    current_version: 3  # 2025 OWASP
    pepper_secret: file:///run/secrets/hash_pepper_v3
  ```

- [ ] **7.1.2** Create Docker secret for pepper
  ```yaml
  # deployments/compose/compose.yml
  secrets:
    hash_pepper_v3:
      file: ./secrets/hash_pepper_v3.secret
  ```

- [ ] **7.1.3** Generate secure pepper (32 bytes)
  ```bash
  openssl rand -base64 32 > deployments/compose/secrets/hash_pepper_v3.secret
  chmod 440 deployments/compose/secrets/hash_pepper_v3.secret
  ```

### Task 7.2: Load Pepper in Hash Service

- [ ] **7.2.1** Update Hash Service initialization to load pepper
  ```go
  pepperPath := viper.GetString("hash_service.pepper_secret")
  pepperBytes, err := loadSecret(pepperPath)  // Handles file:// prefix
  ```

- [ ] **7.2.2** Verify pepper loaded from Docker secrets (NOT env vars, NOT plaintext config)

- [ ] **7.2.3** Test hashing produces different outputs with different peppers

- [ ] **7.2.4** Commit: `feat(hash): implement MANDATORY pepper requirement from Docker secrets`
