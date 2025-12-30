# learn-im Service Template Migration - EXECUTION PLAN

**⚠️ CRITICAL: CONTINUOUS WORK DIRECTIVE** - Read 01-02.continuous-work.instructions.md BEFORE STARTING

**User's #1 Complaint**: "You violated continuous work directives AGAIN!!!"

**Instructions**:

- ✅ Execute ALL phases WITHOUT stopping between phases
- ✅ Do NOT ask questions (use copilot instructions)
- ✅ Do NOT stop for "user approval" (autonomous execution)
- ✅ Commit work at logical checkpoints, but CONTINUE immediately
- ✅ Only stop when ALL work in SERVICE-TEMPLATE.md is 100% complete

**Violation Pattern to Avoid**:

- ❌ Complete Phase 5 → Stop and wait for user
- ✅ Complete Phase 5 → IMMEDIATELY start Phase 6 → Continue until ALL done

---

**CRITICAL**: Evidence-based completion per 06-01.evidence-based.instructions.md

## Your Decisions (QUIZME Answers)

- **Q1**: D (Hybrid) - Fix register/password test files, keep realm_validation with pragma
- **Q2**: A (Replace context.TODO with context.Background)
- **Q3**: E (Do TestMain first, then phases 3,5,4,6 in that order)
- **Q4**: A (Check sizes NOW, split if >500)
- **Q5**: A (Implement TestMain - HIGH PRIORITY to speed up tests)

## EXECUTION SEQUENCE

### ✅ COMPLETED (Phases 8.11-8.13, 9.2, 18)

**Phase 8.11**: COMPLETE - internal/learn/magic deleted, moved to shared
**Phase 8.12**: COMPLETE - All magic constants in magic_learn.go
**Phase 8.13**: COMPLETE - 18 passwords replaced with random generation
**Phase 9.2**: COMPLETE - TOTP fields added to test data
**Phase 18**: COMPLETE - Service instantiation pattern extracted

---

### Phase 0: File Size Analysis (IMMEDIATE)

**Priority**: CRITICAL (Q4: Check NOW, split if >500)

```powershell
# Check all Go files in learn-im
Get-ChildItem -Recurse -Filter *.go internal/learn | ForEach-Object {
    $lines = (Get-Content $_.FullName | Measure-Object -Line).Lines
    if ($lines -gt 400) {
        Write-Output "$lines $($_.FullName)"
    }
} | Sort-Object -Property @{Expression={[int]($_ -split ' ')[0]}} -Descending
```

**Success Criteria**:

- [ ] All files <500 lines (HARD LIMIT) documented
- [ ] Files 400-500 lines justified with comment
- [ ] Files >500 split into logical modules

**Evidence Required**: File size report showing all files under hard limit

---

### Phase 1: TestMain Pattern Implementation (HIGH PRIORITY)

**Priority**: HIGH (Q5: TestMain FIRST to speed up test execution)
**Rationale**: User wants faster local test runs before working on phases 3-6
**Target Packages**: server, crypto, repository tests

**Files to Modify**:

1. `internal/learn/server/server_test.go` - Create TestMain
2. `internal/learn/crypto/password_test.go` - Create TestMain
3. `internal/learn/repository/repository_test.go` - Create TestMain (if exists)

**Pattern**:

```go
var (
    testDB *gorm.DB
    testServer *Server
)

func TestMain(m *testing.M) {
    // Setup: Start heavyweight dependencies ONCE
    ctx := context.Background()

    // In-memory SQLite (lightweight, but still benefits from shared setup)
    testDB, _ = setupTestDatabase()
    testServer, _ = setupTestServer(testDB)

    // Run all tests
    exitCode := m.Run()

    // Cleanup
    cleanupTestResources()
    os.Exit(exitCode)
}

func TestSomething(t *testing.T) {
    // Use testDB and testServer - already initialized
}
```

**Success Criteria**:

- [ ] TestMain implemented in 3+ test packages
- [ ] Tests still pass: `go test ./internal/learn/... -v`
- [ ] Measurable speedup documented (before/after timing)

**Evidence Required**:

```powershell
# Before TestMain:
Measure-Command { go test ./internal/learn/... }

# After TestMain:
Measure-Command { go test ./internal/learn/... }

# Document speedup percentage
```

---

### Phase 1.5: Windows Firewall Exception Fix (CRITICAL - RECURRING REGRESSION)

**Priority**: CRITICAL (Regression occurs repeatedly)
**Rationale**: Tests binding to non-localhost addresses trigger Windows Firewall exceptions, blocking CI/CD automation
**Root Cause Analysis Required**: Understand WHY this pattern keeps reappearing despite copilot instructions

**Investigation Steps**:

1. **Find Offending Tests**:

```powershell
# Search for network binding patterns in server tests
grep -r "Listen\|Serve\|Bind" internal/learn/server/*_test.go
grep -r '"0.0.0.0' internal/learn/server/*_test.go
grep -r 'net\.Listen' internal/learn/server/*_test.go
```

1. **Root Cause Analysis**:

- WHY do tests keep using non-localhost addresses?
- Is it copy-paste from examples?
- Is copilot instruction unclear?
- Are tests starting actual HTTP servers unnecessarily?

1. **Fix According to Copilot Instructions** (03-06.security.instructions.md):

```go
// ❌ WRONG: Triggers Windows Firewall exception
listener, err := net.Listen("tcp", "0.0.0.0:8080")
listener, err := net.Listen("tcp", ":8080")  // Defaults to 0.0.0.0

// ✅ CORRECT: Bind to localhost only
import cryptoutilMagic "cryptoutil/internal/shared/magic"

addr := fmt.Sprintf("%s:%d", cryptoutilMagic.IPv4Loopback, port)  // "127.0.0.1:port"
listener, err := net.Listen("tcp", addr)

// ✅ CORRECT: Use port 0 for dynamic allocation
listener, err := net.Listen("tcp", "127.0.0.1:0")
actualPort := listener.Addr().(*net.TCPAddr).Port
```

1. **Add Detection in cryptoutilCmdCicdLintGotest.Lint**:

**File to Modify**: `internal/cmd/cicd/lint_gotest.go`

**Add Check**:

```go
// Check for Windows Firewall triggering patterns
if strings.Contains(content, `"0.0.0.0`) || strings.Contains(content, `":8080"`) {
    // Verify it's in a test file and likely a Listen/Serve call
    if strings.Contains(content, "net.Listen") || strings.Contains(content, "http.Serve") {
        violations = append(violations, fmt.Sprintf(
            "%s: FIREWALL RISK - Test binds to non-localhost address. "+
            "Use 127.0.0.1 or cryptoutilMagic.IPv4Loopback to prevent Windows Firewall exceptions. "+
            "See 03-06.security.instructions.md",
            filePath,
        ))
    }
}
```

**Success Criteria**:

- [ ] All tests bind to `127.0.0.1` or `cryptoutilMagic.IPv4Loopback` (NEVER `0.0.0.0`)
- [ ] Tests use port `0` for dynamic allocation (prevents conflicts)
- [ ] cryptoutilCmdCicdLintGotest.Lint detects and rejects `0.0.0.0` pattern
- [ ] Root cause documented in DETAILED.md (why this keeps happening)
- [ ] go test runs without triggering Windows Firewall exception

**Evidence Required**:

```powershell
# Verify no firewall-triggering patterns
grep -r '"0.0.0.0' internal/learn/server/*_test.go  # Should return ZERO
grep -r '":[0-9]\{4\}"' internal/learn/server/*_test.go | grep -v ':0"'  # Should return ZERO hardcoded ports

# Verify detection works
go run ./cmd/cicd lint-go-test  # Should catch any violations

# Run tests without firewall prompt
go test ./internal/learn/server -v  # No Windows Firewall popup
```

**Root Cause Prevention**:

- Document in 06-02.anti-patterns.instructions.md
- Add to copilot instructions with examples
- Automated detection in lint-go-test prevents recurrence

---

### Phase 1.6: CGO Check Consolidation (CODE ORGANIZATION)

**Priority**: MEDIUM (Code cleanup, not blocking)
**Rationale**: cryptoutilCmdCicdCheckNoCgoSqlite.Check fits better with linting prod+test code than as separate subcommand

**Consolidation Steps**:

1. **Read Current Implementation**:

```powershell
# Understand what check-no-cgo-sqlite does
cat internal/cmd/cicd/check_no_cgo_sqlite.go
```\r\n\r\n1. **Move Logic into lint_go.go**:

**File to Modify**: `internal/cmd/cicd/lint_go.go`

**Add to CryptoutilCmdCicdLintGo.Lint()**:

```go
func Lint(logger *Logger) error {
    // Existing circular dependency check...

    // Add: CGO-free SQLite check (moved from check_no_cgo_sqlite.go)
    if err := checkNoCgoSqlite(logger); err != nil {
        return fmt.Errorf("CGO SQLite check failed: %w", err)
    }

    return nil
}

func checkNoCgoSqlite(logger *Logger) error {
    // Move implementation from check_no_cgo_sqlite.go here
    // Check for github.com/mattn/go-sqlite3 import
    // Ensure only modernc.org/sqlite is used
    return nil
}
```

1. **Update cicd.go**:

**File to Modify**: `internal/cmd/cicd/cicd.go`

**Remove**:

```go
import (
    cryptoutilCmdCicdCheckNoCgoSqlite "cryptoutil/internal/cmd/cicd/check_no_cgo_sqlite"  // DELETE
    // ... other imports ...
)

const (
    cmdCheckNoCgoSqlite = "check-no-cgo-sqlite" // DELETE
)

switch command {
    case cmdCheckNoCgoSqlite:  // DELETE THIS CASE
        cmdErr = cryptoutilCmdCicdCheckNoCgoSqlite.Check(logger)
}
```

1. **Update CI/CD Workflows**:

**Files to Modify**: `.github/workflows/*.yml`

**Replace**:

```yaml
# Before:
- run: go run ./cmd/cicd check-no-cgo-sqlite

# After: (Already runs as part of lint-go)
- run: go run ./cmd/cicd lint-go
```\r\n\r\n1. **Delete Obsolete File**:

```powershell
Remove-Item internal/cmd/cicd/check_no_cgo_sqlite.go
```\r\n\r\n1. **Update magic/magic_cicd.go**:

**File to Modify**: `internal/shared/magic/magic_cicd.go`

**Remove**:

```go
ValidCommands = map[string]bool{
    "check-no-cgo-sqlite": true,  // DELETE
    // ... other commands ...
}

UsageCICD = `...check-no-cgo-sqlite...`  // DELETE from help text
```

**Success Criteria**:

- [ ] CGO check logic moved into lint_go.go
- [ ] cicd.go no longer has check-no-cgo-sqlite case
- [ ] check_no_cgo_sqlite.go deleted
- [ ] CI/CD workflows updated (use lint-go only)
- [ ] magic_cicd.go ValidCommands updated
- [ ] go build ./cmd/cicd passes
- [ ] go run ./cmd/cicd lint-go includes CGO check

**Evidence Required**:

```powershell
# Verify check-no-cgo-sqlite removed
grep -r "check-no-cgo-sqlite" .  # Should only find in git history, not current files

# Verify lint-go includes CGO check
go run ./cmd/cicd lint-go 2>&1 | grep -i "cgo\|sqlite"  # Should show CGO check running

# Build validation
go build ./cmd/cicd
```

**Rationale**:

- No-CGO import checking is a LINTING activity (checking code quality/compliance)
- Fits naturally with other lint-go checks (circular dependencies)
- Reduces number of subcommands (simpler CLI)
- Consolidates related checks in one place

---

### Phase 2: Hardcoded Password Fixes (CRITICAL)

**Priority**: CRITICAL (Q1: Hybrid approach)
**Decision**: ✅ COMPLETE - 18 passwords already replaced with GeneratePasswordSimple()

**Remaining Work**: Add pragma comments to realm_validation_test.go

#### 2A. Add Pragma Allowlist to realm_validation_test.go (12 passwords)

**File to Modify**: `internal/learn/server/realm_validation_test.go`

**Pattern**:

```go
password: "Abc123!@#xyz", // pragma: allowlist secret - Test vector for validation logic
```

**Rationale**: These are INTENTIONAL test data with specific character patterns to validate realm rules. Not actual secrets.

**Success Criteria**:

- [ ] All 12 hardcoded passwords have pragma comment
- [ ] Comment explains WHY hardcoded (test vectors, not secrets)
- [ ] Tests still pass: `go test ./internal/learn/server -v -run TestRealm`

**Evidence Required**:

```powershell
# Verify no hardcoded passwords WITHOUT pragma:
grep -n '"password"' internal/learn/server/realm_validation_test.go | grep -v "pragma: allowlist secret"
# Should return ZERO results
```

---

### Phase 3: context.TODO() Replacement (QUICK WIN)

**Priority**: HIGH (Q2: Replace with context.Background)

**Files to Modify**:

1. `internal/learn/server/server_lifecycle_test.go:40`
2. `internal/learn/server/register_test.go:355`

**Pattern**:

```go
// Before:
_, err = server.NewPublicServer(context.TODO(), ...)

// After:
_, err = server.NewPublicServer(context.Background(), ...)
```

**Success Criteria**:

- [ ] Zero context.TODO() in learn-im: `grep -r "context.TODO()" internal/learn/` = 0 results
- [ ] Tests still pass

**Evidence Required**: grep output showing zero matches

---

### Phase 4: Switch Statement Conversion (HIGH)

**Priority**: HIGH (copilot instruction 03-01.coding.instructions.md)

**Files to Modify**:

1. `internal/learn/server/handlers.go` - Convert if/else if/else chains

**Pattern**:

```go
// Before:
if username == "" {
    return c.Status(400).JSON(...)
} else if password == "" {
    return c.Status(400).JSON(...)
} else if len(username) < 3 {
    return c.Status(400).JSON(...)
}

// After:
switch {
case username == "":
    return c.Status(400).JSON(...)
case password == "":
    return c.Status(400).JSON(...)
case len(username) < 3:
    return c.Status(400).JSON(...)
}
```

**Success Criteria**:

- [ ] All validation chains converted to switch
- [ ] golangci-lint passes
- [ ] Tests still pass

---

### Phase 5: Quality Gates Execution (CRITICAL - EVIDENCE REQUIRED)

**Priority**: CRITICAL (06-01.evidence-based.instructions.md)

**MANDATORY**: Capture ALL output to files for evidence

#### 5A. Build Validation

```powershell
go build ./internal/learn/... 2>&1 | Tee-Object -FilePath ./test-output/learn_build_evidence.txt
```

**Success**: Zero errors

#### 5B. Linting Validation

```powershell
golangci-lint run ./internal/learn/... 2>&1 | Tee-Object -FilePath ./test-output/learn_lint_evidence.txt
```

**Success**: Zero violations (no exceptions allowed)

#### 5C. Test Validation

```powershell
go test ./internal/learn/... -v -timeout=10m 2>&1 | Tee-Object -FilePath ./test-output/learn_test_evidence.txt
```

**Success**: All tests pass, zero skips

#### 5D. Coverage Validation

```powershell
go test ./internal/learn/... -coverprofile=./test-output/learn_coverage.out -timeout=10m
go tool cover -func=./test-output/learn_coverage.out | Tee-Object -FilePath ./test-output/learn_coverage_summary.txt
go tool cover -html=./test-output/learn_coverage.out -o ./test-output/learn_coverage.html
```

**Success**:

- Production code (server/, crypto/): ≥95% coverage
- Infrastructure (repository/): ≥98% coverage

#### 5E. Mutation Testing

```powershell
gremlins unleash --tags="~integration,~e2e" ./internal/learn/... 2>&1 | Tee-Object -FilePath ./test-output/learn_mutation_evidence.txt
```

**Success**: ≥80% mutation efficacy (Phase 8 target)

#### 5F. Race Detection

```powershell
go test ./internal/learn/... -race -count=2 -timeout=15m 2>&1 | Tee-Object -FilePath ./test-output/learn_race_evidence.txt
```

**Success**: Zero race conditions detected

**Evidence Files Created**:

- `./test-output/learn_build_evidence.txt`
- `./test-output/learn_lint_evidence.txt`
- `./test-output/learn_test_evidence.txt`
- `./test-output/learn_coverage_summary.txt`
- `./test-output/learn_coverage.html`
- `./test-output/learn_mutation_evidence.txt`
- `./test-output/learn_race_evidence.txt`

---

### Phase 6: Phases 3-6 Refactoring (User Sequence 3→5→4→6)

**Priority**: MEDIUM (Q3: Do TestMain first, then 3,5,4,6)
**User Decision**: Complete in order 3→5→4→6 (NOT defer to separate spec)

#### Phase 3: Remove Obsolete Database Tables

**Tables to Remove**:

- users_jwks
- users_messages_jwks
- messages_jwks

**Files to Modify**:

- Migration files in `internal/learn/repository/migrations/`
- Schema files
- Any queries/models referencing obsolete tables

**Success Criteria**:

- [ ] Tables removed from migrations
- [ ] No code references obsolete tables: `grep -r "users_jwks\|users_messages_jwks\|messages_jwks" internal/learn/`
- [ ] Tests pass with new schema

#### Phase 5: Use EncryptBytesWithContext Pattern

**Files to Modify**:

- `internal/learn/crypto/jwe_message_util.go` - Update encryption calls
- Any files calling old encryption functions

**Pattern**:

```go
// Before:
ciphertext, err := encryptFunction(plaintext, key)

// After:
ciphertext, err := jweMessageUtil.EncryptBytesWithContext(ctx, plaintext, key)
```

**Success Criteria**:

- [ ] All encryption calls use EncryptBytesWithContext
- [ ] All decryption calls use DecryptBytesWithContext
- [ ] Tests pass

#### Phase 4: Implement Barrier Encryption for JWKs

**Files to Modify**:

- JWK storage/retrieval code
- Database models for JWK columns

**Pattern**: Integrate barrier encryption from KMS barrier package

**Success Criteria**:

- [ ] JWKs encrypted at rest with barrier pattern
- [ ] Tests verify encryption/decryption
- [ ] E2E tests pass

#### Phase 6: Manual Key Rotation Admin API

**Files to Create/Modify**:

- `internal/learn/server/admin_handlers.go` - New admin endpoints
- Admin API routes registration

**Endpoints**:

- `POST /admin/v1/keys/rotate` - Trigger manual rotation
- `GET /admin/v1/keys/status` - View current key status

**Success Criteria**:

- [ ] Admin endpoints implemented
- [ ] OpenAPI spec updated
- [ ] E2E tests verify rotation works

---

## COMPLETION EVIDENCE CHECKLIST

**CRITICAL**: Do NOT claim completion without ALL evidence below

### Code Quality Evidence

- [ ] `learn_build_evidence.txt` - Zero build errors
- [ ] `learn_lint_evidence.txt` - Zero linting violations
- [ ] `learn_test_evidence.txt` - All tests pass, zero skips

### Coverage Evidence

- [ ] `learn_coverage_summary.txt` - Shows ≥95% production, ≥98% infrastructure
- [ ] `learn_coverage.html` - Visual proof of coverage

### Mutation Evidence

- [ ] `learn_mutation_evidence.txt` - Shows ≥80% mutation efficacy

### Race Evidence

- [ ] `learn_race_evidence.txt` - Zero race conditions

### Compliance Evidence

- [ ] Zero hardcoded passwords (except realm_validation with pragma)
- [ ] Zero context.TODO() usage
- [ ] All if/else chains converted to switch
- [ ] All files <500 lines (HARD LIMIT)
- [ ] TestMain pattern implemented

### Git Evidence

- [ ] Clean working tree: `git status` shows no uncommitted changes
- [ ] Conventional commit message with evidence summary
- [ ] All evidence files committed to repository

---

## COMMIT MESSAGE TEMPLATE

```
feat(learn-im): complete service template migration with evidence

PHASES COMPLETED:
- Phase 0: File size analysis (all <500 lines)
- Phase 1: TestMain pattern (3+ packages, X% speedup)
- Phase 2: Pragma allowlist for realm_validation (12 passwords)
- Phase 3: context.TODO() replacement (2 instances)
- Phase 4: Switch statement conversion (handlers.go)
- Phase 5: Quality gates with evidence capture
- Phase 3-6: Database cleanup, barrier encryption, key rotation

EVIDENCE PROVIDED:
- Build: PASS (learn_build_evidence.txt)
- Linting: PASS (learn_lint_evidence.txt)
- Tests: PASS (learn_test_evidence.txt)
- Coverage: X% production / X% infrastructure (learn_coverage_summary.txt)
- Mutation: X% efficacy (learn_mutation_evidence.txt)
- Race: PASS (learn_race_evidence.txt)

COMPLIANCE:
- 02-07.cryptography: Secure random enforced (18 passwords replaced, 12 pragma added)
- 03-01.coding: Switch statements enforced (handlers.go converted)
- 03-02.testing: TestMain pattern implemented (3+ packages)
- 03-07.linting: Zero violations (full run passed)
- 06-01.evidence-based: ALL evidence files committed

Refs: Phase 8.13 complete, Q1-Q5 QUIZME decisions implemented
```

---

## SUCCESS CRITERIA

This migration is COMPLETE when:

1. **All Phases Done**: 0,1,2,3,4,5,3-6 executed in sequence
2. **All Evidence Files Created**: 7 evidence files in ./test-output/
3. **All Quality Gates Pass**: Build, lint, test, coverage, mutation, race
4. **Zero Violations**: No hardcoded passwords (except pragma), no context.TODO(), no if/else chains, no files >500 lines
5. **Git Clean**: All changes committed with evidence-based proof
6. **User Validation**: User confirms evidence files prove completion

**DO NOT mark complete until ALL evidence exists and user reviews.**
