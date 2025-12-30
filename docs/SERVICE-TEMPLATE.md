# Learn-IM Service Template Migration

**⚠️ CONTINUOUS WORK DIRECTIVE**: This document tracks implementation progress. DO NOT stop work until ALL phases complete and user confirms satisfaction. Progress updates required in DETAILED.md Section 2 timeline.

---

## 📊 EXECUTION STATUS DASHBOARD

**Last Updated**: 2025-12-30
**Overall Progress**: 7/15 phases complete (47%)

### ✅ Completed Phases

| Phase | Title | Completion Date | Evidence | Post-Mortem |
|-------|-------|----------------|----------|-------------|
| 8.11 | Magic constants migration | Previous session | git commit | ❌ Missing |
| 8.12 | Magic constants consolidation | Previous session | git commit | ❌ Missing |
| 8.13 | Password generation (18 instances) | Previous session | git commit | ❌ Missing |
| 9.2 | TOTP test data fields | Previous session | git commit | ❌ Missing |
| 18 | Service instantiation extraction | Previous session | git commit | ❌ Missing |
| 1 | TestMain - server package | Current session | testmain_test.go | ❌ Missing |
| 1 | TestMain - crypto package | Current session | testmain_test.go | ❌ Missing |

### 🔄 In Progress

| Phase | Title | Current Step | Blocker | Next Action |
|-------|-------|--------------|---------|-------------|
| 2 | Hardcoded password pragma | Need to add comments | None | Add pragma comments to realm_validation_test.go |
| 3 | Windows Firewall fixes | Need to verify all test files | None | Grep for 0.0.0.0 patterns |
| 4 | TestMain - e2e package | Not started | None | Create e2e/testmain_test.go |
| 5 | TestMain - integration package | Not started | None | Create integration/testmain_test.go |

### ⏳ Not Started (Ordered by Dependency)

| # | Phase | Title | Depends On | Priority | Estimated Effort |
|---|-------|-------|------------|----------|------------------|
| 6 | 0 | File Size Analysis | None | IMMEDIATE | 1h |
| 7 | 1.6 | CGO Check Consolidation | None | MEDIUM | 2h |
| 8 | 2B | Hardcoded password fixes (complete) | Phase 2 | CRITICAL | 1h |
| 9 | 4 | context.TODO() Replacement | None | HIGH | 30min |
| 10 | 5 | Switch Statement Conversion | None | HIGH | 1h |
| 11 | 6 | Quality Gates Execution | Phases 2-5 | CRITICAL | 2h |
| 12 | 7a | Remove Obsolete DB Tables | Phase 6 | MEDIUM | 1h |
| 13 | 7c | Implement Barrier Encryption | Phase 7a | MEDIUM | 3h |
| 14 | 7b | Use EncryptBytesWithContext | Phase 7c | MEDIUM | 2h |
| 15 | 7d | Manual Key Rotation API | Phases 7a-c | MEDIUM | 3h |

### ❌ Deferred/Questioned

| Item | Title | Reason | Resolution Path |
|------|-------|--------|-----------------|
| - | internal/learn/crypto removal | ❌ **INCORRECT** - crypto/ package is needed | Keep package, ensure it uses Low Entropy Random Hasher from shared/hash |
| - | Move adaptive-sim out of cicd/ | ✅ **CORRECT** - belongs in identity/ | Create task to move to internal/identity/tools/adaptive-sim |
| - | Remove identity_requirements_check | ❌ **INCORRECT** - actively used | Keep - used in feature-template docs and workflows |

### 🎯 Quick Status Summary

- **Build Status**: ❌ Tests failing (exit code 1)
- **Last Test Run**: go test ./internal/learn/... failed
- **Coverage Target**: ≥95% production, ≥98% infrastructure
- **Current Coverage**: Unknown (need to run Phase 6 quality gates)
- **Mutation Target**: ≥80% for Phase 8
- **Blockers**: Test failures preventing progress

---

## 🔍 INVESTIGATION FINDINGS

### Finding 1: internal/learn/crypto Should NOT Be Removed

**User Concern**: "internal\learn\crypto still exists, but there was a phase/task to remove it"

**Investigation Results**:

- ✅ Directory exists: `internal/learn/crypto/`
- ✅ Contains password.go (password hashing utilities)
- ✅ Has TestMain implementation (testmain_test.go)
- ✅ Used by learn-im service for password hashing

**Resolution**:

- ❌ **DO NOT DELETE** - Package is needed for learn-im functionality
- ✅ **ACTION REQUIRED**: Ensure crypto/password.go uses `internal/shared/hash` Low Entropy Random Hasher
- ✅ **VERIFY**: Check if custom password hashing logic should be replaced with shared hasher
- ✅ **DOCUMENT**: Add comment explaining why this package exists despite "crypto" name

### Finding 2: adaptive-sim Location

**User Concern**: "Is internal\cmd\cicd\adaptive-sim used or maybe it is in wrong directory?"

**Investigation Results**:

- ✅ Tool exists: `internal/cmd/cicd/adaptive-sim/`
- ✅ Purpose: Simulates adaptive authentication policy changes for identity service
- ✅ Imports: `cryptoutil/internal/identity/idp/userauth`
- ❌ **INCORRECT LOCATION**: Not a CICD linting/formatting tool

**Resolution**:

- ✅ **CREATE TASK**: Move to `internal/identity/tools/adaptive-sim/`
- ✅ **RATIONALE**: Tool is identity-specific, not general CICD
- ✅ **TIMELINE**: Phase 8 (after learn-im template complete)

### Finding 3: identity_requirements_check Usage

**User Concern**: "Is internal\cmd\cicd\identity_requirements_check used or can it be removed?"

**Investigation Results**:

- ✅ Tool exists: `internal/cmd/cicd/identity_requirements_check/`
- ✅ Purpose: Validates requirements traceability from acceptance criteria to test implementations
- ✅ **ACTIVELY USED**:
  - Referenced in `docs/feature-template/*.md` (multiple locations)
  - Used in `.github/workflows/ci-quality.yml` (commented out but intended)
  - Command: `go run ./cmd/cicd go-identity-requirements-check --strict`

**Resolution**:

- ❌ **DO NOT REMOVE** - Tool is actively used
- ✅ **KEEP**: Essential for requirements validation in identity service
- ✅ **LOCATION CORRECT**: Belongs in cicd/ as it validates test coverage against requirements

---

## 📋 QUIZME ANSWERS (Decisions Made)

**Q1**: 18 hardcoded passwords in learn-im tests - keep or replace?
**A1**: **D** - Hybrid: Replace GeneratePasswordSimple() (done ✅), add pragma allowlist to validation tests (12 passwords) (pending ⏳)

**Q2**: Phase 1.5 (Windows Firewall) - when to execute?
**A2**: **A** - Immediately after Phase 1 TestMain (CRITICAL recurring regression)

**Q3**: Phases 3-6 execution sequence?
**A3**: **E** (Write-in) - Do TestMain first (Phase 1), then execute 3→5→4→6 in dependency order:

- Phase 3: Remove obsolete DB tables
- Phase 5: Implement barrier encryption (dependency for Phase 4)
- Phase 4: Use EncryptBytesWithContext (depends on Phase 5)
- Phase 6: Manual key rotation API

**Q4**: TestMain pattern priority?
**A4**: **A** - HIGH priority (significant test speedup for heavyweight setup)

**Q5**: Phase 1.6 (CGO check consolidation) priority?
**A5**: **A** - MEDIUM priority (code cleanup, not blocking)

---

## 📝 EXECUTION SEQUENCE

### Phase 0: File Size Analysis (IMMEDIATE) ⏳

**Priority**: IMMEDIATE (copilot instruction 03-01.coding.instructions.md line size limits)

**Files to Check**:

```powershell
# Find all Go files >400 lines (approaching 500 hard limit)
Get-ChildItem -Recurse -Filter "*.go" -Path internal/learn |
    Where-Object { (Get-Content $_.FullName).Count -gt 400 } |
    Select-Object FullName, @{Name="Lines";Expression={(Get-Content $_.FullName).Count}} |
    Sort-Object Lines -Descending
```

**Thresholds**:

- Soft limit: 300 lines (ideal)
- Medium limit: 400 lines (acceptable with justification)
- Hard limit: 500 lines (NEVER EXCEED - refactor required)

**Success Criteria**:

- [ ] All files documented with line counts
- [ ] Any files >400 lines flagged for refactoring
- [ ] Plan created for files approaching/exceeding 500 lines

---

### Phase 1: TestMain Pattern Implementation (HIGH PRIORITY) ✅ PARTIAL

**Completed** ✅:

1. ✅ `internal/learn/server/testmain_test.go` - Exists
2. ✅ `internal/learn/crypto/testmain_test.go` - Exists
3. ✅ `internal/template/server/test_main_test.go` - Exists

**Incomplete** ⏳:

1. ⏳ `internal/learn/e2e/` - Missing TestMain (user confirmed)
2. ⏳ `internal/learn/integration/` - Missing TestMain (has concurrent_test.go)

**Files to Create**:

**1. internal/learn/e2e/testmain_e2e_test.go**:

```go
// Copyright (c) 2025 Justin Cranford

package e2e

import (
    "context"
    "os"
    "testing"
    "time"

    googleUuid "github.com/google/uuid"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
    testDBContainer *postgres.PostgresContainer
    testDBDSN       string
)

func TestMain(m *testing.M) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    // Start PostgreSQL test container ONCE for all E2E tests
    var err error
    testDBContainer, err = postgres.RunContainer(ctx,
        postgres.WithDatabase("test_e2e_"+googleUuid.NewV7().String()),
        postgres.WithUsername("test_user_"+googleUuid.NewV7().String()),
        postgres.WithPassword("test_pass_"+googleUuid.NewV7().String()),
    )
    if err != nil {
        panic("Failed to start PostgreSQL container: " + err.Error())
    }

    testDBDSN, err = testDBContainer.ConnectionString(ctx)
    if err != nil {
        _ = testDBContainer.Terminate(ctx)
        panic("Failed to get connection string: " + err.Error())
    }

    // Run all E2E tests
    exitCode := m.Run()

    // Cleanup
    _ = testDBContainer.Terminate(ctx)

    os.Exit(exitCode)
}
```

**2. internal/learn/integration/testmain_integration_test.go**:

```go
// Copyright (c) 2025 Justin Cranford

package integration

import (
    "context"
    "os"
    "testing"
    "time"

    googleUuid "github.com/google/uuid"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
    testDBContainer *postgres.PostgresContainer
    testDBDSN       string
)

func TestMain(m *testing.M) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    // Start PostgreSQL test container ONCE for all integration tests
    var err error
    testDBContainer, err = postgres.RunContainer(ctx,
        postgres.WithDatabase("test_integration_"+googleUuid.NewV7().String()),
        postgres.WithUsername("test_user_"+googleUuid.NewV7().String()),
        postgres.WithPassword("test_pass_"+googleUuid.NewV7().String()),
    )
    if err != nil {
        panic("Failed to start PostgreSQL container: " + err.Error())
    }

    testDBDSN, err = testDBContainer.ConnectionString(ctx)
    if err != nil {
        _ = testDBContainer.Terminate(ctx)
        panic("Failed to get connection string: " + err.Error())
    }

    // Run all integration tests
    exitCode := m.Run()

    // Cleanup
    _ = testDBContainer.Terminate(ctx)

    os.Exit(exitCode)
}
```

**Success Criteria**:

- [x] server/testmain_test.go exists ✅
- [x] crypto/testmain_test.go exists ✅
- [x] template/server/test_main_test.go exists ✅
- [ ] e2e/testmain_e2e_test.go created
- [ ] integration/testmain_integration_test.go created
- [ ] All test packages pass after TestMain implementation
- [ ] Measure test speedup (before/after timing)

**Evidence Required**:

```powershell
# Before TestMain (baseline timing)
Measure-Command { go test ./internal/learn/e2e -v }
Measure-Command { go test ./internal/learn/integration -v }

# After TestMain (improved timing)
Measure-Command { go test ./internal/learn/e2e -v }
Measure-Command { go test ./internal/learn/integration -v }

# Calculate speedup percentage
```

**POST-MORTEM (Server Package TestMain)** ✅:

**Completion Date**: Previous session (exact date unknown)
**Commits**: Unknown (need to check git log)

**What Went Well**:

- TestMain pattern successfully implemented
- Tests still pass with shared setup

**Issues Encountered**:

- Unknown (no documentation from previous session)

**Lessons Learned**:

- ⚠️ **CRITICAL LESSON**: TestMain was only partially implemented
- ⚠️ **E2E and integration packages were missed**
- ⚠️ **No evidence or post-mortem from previous session**

**Impact on Remaining Phases**:

- Must complete e2e and integration TestMain before claiming Phase 1 complete
- Missing post-mortems make it impossible to learn from previous work

---

### Phase 2: Hardcoded Password Fixes (CRITICAL) 🔄

**Decision**: Hybrid approach - 18 passwords replaced ✅, 12 pragma comments needed ⏳

#### 2A. Verify GeneratePasswordSimple() Replacement ✅

**Status**: ✅ CLAIMED COMPLETE in Phase 8.13

**Verification Needed**:

```powershell
# Check that GeneratePasswordSimple() is used instead of hardcoded passwords
grep -r "GeneratePasswordSimple" internal/learn/server/*_test.go
# Should show 18+ instances

# Verify no NEW hardcoded passwords added
grep -n 'password.*:.*"[A-Z]' internal/learn/server/*_test.go | grep -v "pragma: allowlist"
```

#### 2B. Add Pragma Allowlist to realm_validation_test.go ⏳

**File to Modify**: `internal/learn/server/realm_validation_test.go`

**Pattern** (apply to all 12 instances):

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
# Verify ALL hardcoded passwords have pragma comment
grep -n '"password".*:' internal/learn/server/realm_validation_test.go | grep -v "pragma: allowlist secret"
# Should return ZERO results

# Verify tests pass
go test ./internal/learn/server -v -run TestRealm
```

---

### Phase 3: Windows Firewall Exception Fix (CRITICAL) ⏳

**Priority**: CRITICAL (06-02.anti-patterns.instructions.md - recurring regression)

**Investigation Questions**:

- Are there `0.0.0.0` bindings in test files?
- Are there hardcoded ports (`:8080`) instead of dynamic (`:0`)?
- Is copilot instruction unclear?
- Are tests starting actual HTTP servers unnecessarily?

**Scan for Violations**:

```powershell
# Check for Windows Firewall triggering patterns
grep -r '"0.0.0.0' internal/learn/**/*_test.go
grep -r '":8080"' internal/learn/**/*_test.go
grep -r '":9090"' internal/learn/**/*_test.go

# Check if tests use proper loopback binding
grep -r 'cryptoutilMagic.IPv4Loopback' internal/learn/**/*_test.go
grep -r '"127.0.0.1:0"' internal/learn/**/*_test.go
```

**Fix Pattern** (if violations found):

```go
// ❌ WRONG: Triggers Windows Firewall exception
listener, err := net.Listen("tcp", "0.0.0.0:8080")

// ✅ CORRECT: Bind to localhost only with dynamic port
import cryptoutilMagic "cryptoutil/internal/shared/magic"

addr := fmt.Sprintf("%s:%d", cryptoutilMagic.IPv4Loopback, 0)  // "127.0.0.1:0"
listener, err := net.Listen("tcp", addr)
actualPort := listener.Addr().(*net.TCPAddr).Port
```

**Success Criteria**:

- [ ] Zero `0.0.0.0` bindings in test files
- [ ] All tests use `127.0.0.1` or `cryptoutilMagic.IPv4Loopback`
- [ ] Tests use port `0` for dynamic allocation
- [ ] cryptoutilCmdCicdLintGotest.Lint detects and rejects violations
- [ ] Tests run without triggering Windows Firewall prompt

**Evidence Required**:

```powershell
# Verify no firewall-triggering patterns
grep -r '"0.0.0.0' internal/learn/**/*_test.go  # Should return ZERO
grep -r '":[0-9]\{4\}"' internal/learn/**/*_test.go | grep -v ':0"'  # Should return ZERO

# Verify detection works
go run ./cmd/cicd lint-go-test  # Should catch any violations

# Run tests without firewall prompt
go test ./internal/learn/... -v  # No Windows Firewall popup
```

---

### Phase 4: context.TODO() Replacement (QUICK WIN) ⏳

**Priority**: HIGH (easy fix, improves code quality)

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
- [ ] Tests still pass after replacement

**Evidence Required**:

```powershell
# Verify zero context.TODO()
grep -r "context.TODO()" internal/learn/
# Should return 0 matches

# Verify tests pass
go test ./internal/learn/server -v
```

---

### Phase 5: Switch Statement Conversion (HIGH) ⏳

**Priority**: HIGH (copilot instruction 03-01.coding.instructions.md)

**Files to Modify**:

1. `internal/learn/server/handlers.go` - Convert if/else if/else chains to switch statements

**Pattern**:

```go
// Before:
if username == "" {
    return c.Status(400).JSON(fiber.Map{"error": "username required"})
} else if password == "" {
    return c.Status(400).JSON(fiber.Map{"error": "password required"})
} else if len(username) < 3 {
    return c.Status(400).JSON(fiber.Map{"error": "username too short"})
}

// After:
switch {
case username == "":
    return c.Status(400).JSON(fiber.Map{"error": "username required"})
case password == "":
    return c.Status(400).JSON(fiber.Map{"error": "password required"})
case len(username) < 3:
    return c.Status(400).JSON(fiber.Map{"error": "username too short"})
}
```

**Success Criteria**:

- [ ] All validation chains converted to switch
- [ ] golangci-lint passes
- [ ] Tests still pass

**Evidence Required**:

```powershell
# Verify changes
git diff internal/learn/server/handlers.go

# Verify linting passes
golangci-lint run ./internal/learn/server

# Verify tests pass
go test ./internal/learn/server -v
```

---

### Phase 6: Quality Gates Execution (CRITICAL - EVIDENCE REQUIRED) ⏳

**Priority**: CRITICAL (06-01.evidence-based.instructions.md)

**MANDATORY**: Capture ALL output to files for evidence

#### 6A. Build Validation

```powershell
go build ./internal/learn/... 2>&1 | Tee-Object -FilePath ./test-output/learn_build_evidence.txt
```

**Success**: Zero errors

#### 6B. Linting Validation

```powershell
golangci-lint run ./internal/learn/... 2>&1 | Tee-Object -FilePath ./test-output/learn_lint_evidence.txt
```

**Success**: Zero violations (no exceptions allowed)

#### 6C. Test Validation

```powershell
go test ./internal/learn/... -v -timeout=10m 2>&1 | Tee-Object -FilePath ./test-output/learn_test_evidence.txt
```

**Success**: All tests pass, zero skips

#### 6D. Coverage Validation

```powershell
go test ./internal/learn/... -coverprofile=./test-output/learn_coverage.out -timeout=10m
go tool cover -func=./test-output/learn_coverage.out | Tee-Object -FilePath ./test-output/learn_coverage_summary.txt
go tool cover -html=./test-output/learn_coverage.out -o ./test-output/learn_coverage.html
```

**Success**:

- Production code (server/, crypto/): ≥95% coverage
- Infrastructure (repository/): ≥98% coverage

#### 6E. Mutation Testing

```powershell
gremlins unleash --tags="~integration,~e2e" ./internal/learn/... 2>&1 | Tee-Object -FilePath ./test-output/learn_mutation_evidence.txt
```

**Success**: ≥80% mutation efficacy (Phase 8 target)

#### 6F. Race Detection

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

### Phase 7: Database and Encryption Refactoring ⏳

**User Decision**: Complete in order 7a→7c→7b→7d (dependency-based sequence)

#### Phase 7a: Remove Obsolete Database Tables

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
- [ ] No code references: `grep -r "users_jwks\|users_messages_jwks\|messages_jwks" internal/learn/` = 0
- [ ] Tests pass with new schema

#### Phase 7c: Implement Barrier Encryption for JWKs (BEFORE 7b)

**Dependency**: Must complete BEFORE Phase 7b

**Files to Modify**:

- JWK storage/retrieval code
- Database models for JWK columns

**Pattern**: Integrate barrier encryption from KMS barrier package

**Success Criteria**:

- [ ] JWKs encrypted at rest with barrier pattern
- [ ] Tests verify encryption/decryption
- [ ] E2E tests pass

#### Phase 7b: Use EncryptBytesWithContext Pattern (AFTER 7c)

**Dependency**: Requires Phase 7c barrier encryption implementation

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

#### Phase 7d: Manual Key Rotation Admin API

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

## ✅ COMPLETION EVIDENCE CHECKLIST

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
- [ ] TestMain pattern implemented in ALL test packages

### Git Evidence

- [ ] Clean working tree: `git status` shows no uncommitted changes
- [ ] Conventional commit message with evidence summary
- [ ] All evidence files committed to repository

---

## 🎯 SUCCESS CRITERIA

This migration is COMPLETE when:

1. ✅ **All Phases Done**: 0-7 executed in dependency order
2. ✅ **All Evidence Files Created**: 7 evidence files in ./test-output/
3. ✅ **All Quality Gates Pass**: Build, lint, test, coverage, mutation, race
4. ✅ **Zero Violations**: No hardcoded passwords (except pragma), no context.TODO(), no files >500 lines
5. ✅ **Git Clean**: All changes committed with evidence
6. ✅ **Post-Mortems Added**: Each completed phase has retrospective analysis
7. ✅ **User Validation**: User confirms evidence files prove completion

**DO NOT mark complete until ALL evidence exists and user reviews.**

---

## 📚 POST-MORTEM TEMPLATE

```markdown
### Phase X: [Title] - ✅ COMPLETE

**Completion Date**: YYYY-MM-DD
**Commits**: [hash1], [hash2]
**Evidence**: [test output], [coverage report]

#### Post-Mortem Analysis

**What Went Well**:
- [Success point 1]
- [Success point 2]

**Issues Encountered**:
- [Issue 1 with resolution]
- [Issue 2 with resolution]

**Lessons Learned**:
- [Lesson 1]
- [Lesson 2]

**Impact on Remaining Phases**:
- [How this affects phase Y]
- [New risk identified for phase Z]

**Reordering Needed**:
- [If dependencies changed, reorder phases]
```

---

## 📝 NEXT ACTIONS

1. **IMMEDIATE**: Fix test failures (exit code 1)
   - Read `test-output/learn_test_evidence_final.txt`
   - Identify which tests failed
   - Fix root cause
   - Re-run tests until passing

2. **Phase 0**: Run file size analysis
   - Find files >400 lines
   - Plan refactoring for files approaching 500 lines

3. **Complete Phase 1**: Add TestMain to e2e and integration packages
   - Create testmain_e2e_test.go
   - Create testmain_integration_test.go
   - Measure speedup

4. **Complete Phases 2-7**: Execute in sequence
   - Each phase must have evidence
   - Each phase must have post-mortem
   - No skipping phases

5. **Final Commit**: Evidence-based completion proof
   - All quality gates passing
   - All evidence files committed
   - Comprehensive commit message
