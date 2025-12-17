Lessons Learned - Preventing Regressions
=========================================

Date: 2025-12-16
Context: User frustration: "ALL OF THESE ARE REGRESSIONS DUE TO REPEAT MISTAKES!!!"

## CRITICAL FORMAT_GO SELF-MODIFICATION REGRESSION

### Timeline of Failures

1. **Nov 17, 2025 (b934879b)**: First fix - added backticks to comments
2. **Nov 20, 2025 (71b0e90d)**: Added self-exclusion patterns
3. **Dec 16, 2025 (b0e4b6ef)**: Fixed infinite loop (wrong counter)
4. **Dec 16, 2025 (8c855a6e)**: Fixed test data corruption
5. **Dec 16, 2025 (a6a5a750)**: COMPLETE FIX - restored clean baseline + defensive check

### Root Cause Analysis

**CRITICAL: LLM agents (Copilot, Grok) lose exclusion context during narrow-focus refactoring**

When asked to "refactor" or "improve" code:

- Agent focuses on narrow scope (single function/file)
- Agent does NOT read broader context (filter.go, magic_cicd.go)
- Agent sees "verbose comments" and "simplifies" them
- Agent sees "`interface{}`" and "modernizes" to "any"
- Agent sees test data with "interface{}" and "fixes" to "any"

Result: Self-modification protection bypassed by "helpful" refactoring

### Why This Regression Occurred MULTIPLE Times

1. **No runtime verification**: Tests didn't validate source file integrity
2. **HEAD was corrupted**: Previous fix (20a06ceb) only fixed enforce_any.go, not test files
3. **Incomplete validation**: self_modification_test.go existed but wasn't catching mutations
4. **Trust in HEAD**: Assumed HEAD was correct, applied fixes incrementally
5. **LLM context loss**: Each refactoring session forgot exclusion requirements

### Prevention Strategies (MANDATORY)

#### 1. ALWAYS Restore from Clean Baseline FIRST

```bash
# DON'T: Apply fixes on potentially corrupted HEAD
git add file.go
git commit --amend

# DO: Restore clean baseline, THEN apply targeted fixes
git checkout <clean-commit> -- path/to/package/
# Apply ONLY the new fix (e.g., defensive check)
# Verify tests pass
# Commit as NEW commit (not amend)
```

#### 2. ALWAYS Read Entire Context Before Refactoring

```bash
# DON'T: Just read the target file
read_file enforce_any.go

# DO: Read ALL related context files
read_file enforce_any.go
read_file filter.go                    # Self-exclusion patterns
read_file magic_cicd.go                # Exclusion constants
read_file format_go_test.go            # Test data patterns
read_file self_modification_test.go    # Validation patterns
```

#### 3. ALWAYS Validate Source Integrity After Tests

```bash
# DON'T: Just check test pass/fail
go test ./internal/cmd/cicd/format_go/

# DO: Verify no self-modifications occurred
go test ./internal/cmd/cicd/format_go/
git status --short  # Should show ZERO uncommitted changes
grep "interface{}" enforce_any.go  # Should find CORRECT strings
grep "interface{}" format_go_test.go  # Should find test data
```

#### 4. ALWAYS Use Defensive Checks with filepath.Abs()

```go
// DON'T: String matching on relative paths (breaks with tmpDir)
if strings.Contains(filePath, "format_go") {
    return 0, nil  // Blocks tmpDir test files too!
}

// DO: Use filepath.Abs() to distinguish source from test files
absPath, pathErr := filepath.Abs(filePath)
if pathErr == nil {
    if strings.Contains(absPath, filepath.Join("internal", "cmd", "cicd", "format_go")) &&
        !strings.Contains(absPath, filepath.Join("R:", "temp")) &&  // Not tmpDir
        !strings.Contains(absPath, filepath.Join("C:", "temp")) {   // Not tmpDir
        return 0, nil  // Skip only actual source files
    }
}
```

#### 5. ALWAYS Commit Incrementally (NOT Amend)

```bash
# DON'T: Amend repeatedly (loses history, masks mistakes)
git commit -m "fix"
git add more_fixes
git commit --amend

# DO: Commit incrementally (preserves history, enables bisect)
git commit -m "fix(format_go): restore clean baseline from 07192eac"
# Test, verify
git commit -m "fix(format_go): add defensive check with filepath.Abs()"
# Test, verify
git commit -m "fix(format_go): verify self_modification_test catches regressions"
```

## FLAKY TEST REGRESSION

### Timeline of Failures

- **Dec 16, 2025 (ec9a071c)**: TestHealthChecks timeout (context deadline exceeded) - FIXED
- **Dec 16, 2025 (ec9a071c)**: TestMutualTLS timeout (i/o timeout) - FIXED
- **Dec 16, 2025 (full test run)**: TestMutualTLS failed with EOF (read timeout on TLS connection)
- **Dec 16, 2025 (manual re-test)**: TestMutualTLS passed 6 consecutive times (1 isolated + 5 count runs)

### Root Cause Analysis

**CRITICAL: Hard-coded short timeouts insufficient for shared CI/CD resources**

Tests assumed:

- Fast local execution (<1s)
- Dedicated CPU resources
- No network latency

Reality (GitHub Actions, local under load):

- Shared CPU (steal time, context switches)
- Network latency (loopback still has overhead)
- Parallel test execution (resource contention)

**TestMutualTLS Specific Issue**:

- **Full test suite run**: Server-side read timeout occurs before client completes (EOF error)
- **Isolated runs**: No timeout, test passes consistently
- **Hypothesis**: Parallel test execution causes TLS handshake delays under load
- **Status**: Intermittent, race condition likely, but passes when run in isolation or with -count

### Prevention Strategies (MANDATORY)

#### 1. ALWAYS Use Generous Timeouts for Network Operations

```go
// DON'T: Use implicit timeouts or short durations
client := &http.Client{}  // No timeout = infinite wait (bad)
client := &http.Client{Timeout: 1 * time.Second}  // Too short for CI

// DO: Use 5s+ timeouts for network operations
client := &http.Client{
    Transport: &http.Transport{TLSClientConfig: tlsConfig},
    Timeout: 5 * time.Second,  // Generous for shared resources
}
```

#### 2. ALWAYS Use Context Timeouts for HTTP Requests

```go
// DON'T: Use bare requests without timeout context
req, _ := http.NewRequest("GET", url, nil)
resp, err := client.Do(req)

// DO: Always use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
```

#### 3. ALWAYS Check Error Types in Retry Loops

```go
// DON'T: Retry blindly on all errors
for retry := 0; retry < maxRetries; retry++ {
    resp, err := client.Do(req)
    if err == nil { break }
    time.Sleep(delay)  // What if error is permanent?
}

// DO: Check if error is retryable (timeout, connection refused)
for retry := 0; retry < maxRetries; retry++ {
    resp, err := client.Do(req)
    if err == nil && resp.StatusCode == http.StatusOK { break }

    var ne net.Error
    if errors.As(err, &ne) && !ne.Timeout() {
        // Permanent error, don't retry
        return err
    }

    time.Sleep(time.Duration(retry+1) * baseDelay)  // Exponential backoff
}
```

#### 4. ALWAYS Run Tests Multiple Times to Detect Flakiness

```bash
# DON'T: Run once and assume stable
go test ./internal/kms/server/application/

# DO: Run 3-5 times to detect intermittent failures
go test -count=3 ./internal/kms/server/application/
go test -count=5 ./internal/shared/crypto/certificate/
```

## COVERAGE REGRESSION

### Current State (Dec 16, 2025)

15+ production packages below 95% target:

- internal/cmd/cicd/format_go: 69.4% (should be 95%+)
- internal/identity/authz: 72.9%
- internal/identity/idp: 66.0%
- internal/jose/server: 62.1%
- internal/kms/server/application: 64.6%

### Root Cause Analysis

**CRITICAL: Accepting incremental coverage improvements without enforcing target**

Pattern:

1. Package starts at 60% coverage
2. Add tests, reach 70% coverage
3. Celebrate improvement (+10 percentage points)
4. Stop working, leave at 70% (still 25 points below target)
5. Move to next package, repeat pattern

Result: Technical debt accumulates, target never reached

### Prevention Strategies (MANDATORY)

#### 1. ALWAYS Generate Baseline Coverage HTML BEFORE Writing Tests

```bash
# DON'T: Write tests blindly hoping for improvement
go test ./internal/jose/crypto/ -coverprofile=coverage.out
# Add 60 tests
go test ./internal/jose/crypto/ -coverprofile=coverage.out  # Still 82.7%?

# DO: Analyze gaps FIRST, then write targeted tests
go test ./internal/jose/crypto/ -coverprofile=baseline.out
go tool cover -html=baseline.out -o baseline.html
# Open HTML, find RED lines (uncovered)
# Write tests for RED lines ONLY
go test ./internal/jose/crypto/ -coverprofile=new.out
go tool cover -func new.out | grep -v "100.0%"  # Verify improvement
```

#### 2. ALWAYS Use Coverage Tools to Find Exact Gaps

```bash
# DON'T: Guess which tests to add
# "I think we need tests for error handling"

# DO: Use go tool cover -func to find specific gaps
go tool cover -func ./test-output/coverage_jose.out | \
    Where-Object { $_ -match 'jose.*\.go:' -and \
    $_ -match '(\d+\.\d+)%' -and \
    [double]$Matches[1] -lt 90.0 }

# Output shows EXACT functions below threshold:
# internal/jose/crypto/jwe_util.go:45: buildJWE  75.0%
# internal/jose/crypto/jws_util.go:78: buildJWS  68.2%
```

#### 3. ALWAYS Enforce "No Exceptions" Rule

```bash
# User's frustration: "I am tired of you making exceptions"

# DON'T: Rationalize low coverage
# "This package is mostly error handling" (add error tests!)
# "This is just a thin wrapper" (still needs 95% coverage!)
# "We improved from 60% to 70%" (target is 95%, not improvement)

# DO: Treat <95% as BLOCKING issue
if coverage < 95%; then
    echo "BLOCKING: Coverage below 95% - add tests"
    exit 1
fi
```

#### 4. ALWAYS Use Table-Driven Tests to Maximize Coverage/LOC Ratio

```go
// DON'T: Individual test functions (verbose, redundant)
func TestHS256(t *testing.T) { /* ... */ }
func TestHS384(t *testing.T) { /* ... */ }
func TestHS512(t *testing.T) { /* ... */ }
// Result: 150 lines, 75% coverage

// DO: Table-driven with all variants (compact, comprehensive)
func TestHMAC_AllAlgorithms(t *testing.T) {
    tests := []struct {
        name string
        alg  Algorithm
        bits int
    }{
        {"HS256", AlgHS256, 256},
        {"HS384", AlgHS384, 384},
        {"HS512", AlgHS512, 512},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { /* test logic */ })
    }
}
// Result: 20 lines, 95% coverage
```

## EXECUTION TIME REGRESSION

### Current State (Dec 16, 2025)

Target: <20s per package
Reality: 7 packages violate (3 critically slow):

- internal/jose/crypto: 89.849s (4.5x over budget)
- internal/kms/server/application: 88.282s (4.4x over budget)
- internal/jose/server: 52.018s (2.6x over budget)

### Root Cause Analysis

**CRITICAL: Running ALL algorithm variants at 100% probability**

Pattern:

1. Tests for RSA 2048/3072/4096 all run every time (TestProbAlways)
2. Tests for AES 128/192/256 all run every time (TestProbAlways)
3. Tests for ECDSA P-256/P-384/P-521 all run every time (TestProbAlways)

Result: 100+ crypto tests × 1-5s each = 90s execution time

Coverage Value: Testing RSA 3072 adds <1% additional coverage vs RSA 2048
Time Cost: Each RSA 3072 test adds 2-5s execution time

### Prevention Strategies (MANDATORY)

#### 1. ALWAYS Use Probabilistic Execution for Algorithm Variants

```go
// DON'T: Run all key sizes every time (100% probability)
tests := []struct {
    name string
    bits int
}{
    {"RSA_2048", 2048},  // Runs every time
    {"RSA_3072", 3072},  // Runs every time (adds 5s, 0.1% coverage)
    {"RSA_4096", 4096},  // Runs every time (adds 10s, 0.1% coverage)
}

// DO: Run base algorithm always, variants probabilistically
tests := []struct {
    name string
    bits int
    prob float64  // Execution probability
}{
    {"RSA_2048", 2048, cryptoutilMagic.TestProbAlways},   // 100% (base)
    {"RSA_3072", 3072, cryptoutilMagic.TestProbTenth},    // 10% (variant)
    {"RSA_4096", 4096, cryptoutilMagic.TestProbTenth},    // 10% (variant)
}

for _, tt := range tests {
    if rand.Float64() > tt.prob {
        t.Skip("Skipped due to probabilistic execution")
    }
    // ... test logic
}

// Result: 90s → 15s execution time (6x faster)
// Coverage: 95% maintained (statistical sampling works)
```

#### 2. ALWAYS Apply Probability to Redundant Coverage Only

```go
// ALWAYS RUN (fundamentally different algorithms):
- RSA vs ECDSA vs EdDSA (100% probability)
- AES vs ChaCha20 (100% probability)
- HMAC-SHA256 vs HMAC-SHA512 (100% probability)

// PROBABILISTIC RUN (same algorithm, different key size):
- RSA 2048 vs 3072 vs 4096 (base 100%, variants 10%)
- AES 128 vs 192 vs 256 (base 100%, variants 10%)
- ECDSA P-256 vs P-384 vs P-521 (base 100%, variants 10%)
```

#### 3. ALWAYS Measure Before/After Execution Times

```bash
# DON'T: Apply probabilistic execution blindly
# "I think this will be faster"

# DO: Measure baseline, apply optimization, measure again
# Baseline
time go test ./internal/jose/crypto/
# Output: 89.849s

# Apply probabilistic execution
# (modify test code)

# Verify improvement
time go test ./internal/jose/crypto/
# Output: 15.234s (target: <20s) ✅
```

## META-LESSONS (MOST IMPORTANT)

### 1. Trust, But Verify

- ✅ Trust pre-commit hooks (they catch 95% of issues)
- ❌ Don't trust HEAD is always correct (verify baseline)
- ✅ Trust test results (green = good coverage)
- ❌ Don't trust coverage % without HTML analysis (may miss critical branches)

### 2. Evidence-Based Progress

- ❌ Subjective: "I added lots of tests" (quantity ≠ quality)
- ✅ Objective: "Coverage increased from 82.7% to 95.2%" (measurable)
- ❌ Subjective: "Tests run fast for me locally" (local ≠ CI)
- ✅ Objective: "go test reports 15.2s (target <20s)" (repeatable)

### 3. No Exceptions Rule

- ❌ "This package is special" (every package needs 95% coverage)
- ❌ "We improved by 10 points" (target is absolute, not relative)
- ❌ "This is just a demo" (demos need tests too if committed to main)
- ✅ "Coverage: 95.2%" (meets standard, no justification needed)

### 4. Incremental Commits, Not Amendments

- ❌ Amend repeatedly (loses history, masks mistakes, hard to bisect)
- ✅ Commit incrementally (preserves timeline, enables rollback, shows thought process)
- ❌ "Fixed everything in one commit" (impossible to review, hard to revert)
- ✅ "Fix baseline" → "Add defensive check" → "Verify tests" (clear progression)

### 5. Context Completeness

- ❌ Read single file in isolation (miss exclusion patterns, design intent)
- ✅ Read entire package context (understand WHY code exists, HOW it's protected)
- ❌ Refactor "obviously verbose" code (verbosity may be intentional protection)
- ✅ Understand design rationale BEFORE suggesting changes

## IMPLEMENTATION CHECKLIST (MANDATORY FOR ALL FUTURE WORK)

Before Starting Work:

- [ ] Read ALL related files (not just target file)
- [ ] Generate baseline coverage HTML
- [ ] Verify HEAD is clean (git status)
- [ ] Run tests 3x to detect flakiness

During Work:

- [ ] Write targeted tests for RED lines (from HTML)
- [ ] Use table-driven patterns (maximize coverage/LOC)
- [ ] Apply probabilistic execution to variants
- [ ] Check coverage after each test addition

Before Committing:

- [ ] Verify coverage ≥95% (no exceptions)
- [ ] Verify execution time <20s per package
- [ ] Verify no self-modifications (git status clean)
- [ ] Run golangci-lint --fix (catch trivial issues)
- [ ] Run tests 3x (verify stability)

After Committing:

- [ ] Update DETAILED.md timeline
- [ ] Document before/after metrics
- [ ] Note any remaining gaps
- [ ] Create follow-up tasks if needed

## CRITICAL TAKEAWAY

**User frustration stems from repeated regressions due to:**

1. Incomplete baseline verification (trusting HEAD was correct)
2. Narrow context reading (missing exclusion patterns)
3. Accepting "improvement" instead of "target met"
4. Missing probabilistic execution (slow tests)
5. Insufficient timeout margins (flaky tests)

**Solution: NEVER SKIP VERIFICATION STEPS**

- Baseline coverage HTML analysis (before writing tests)
- Complete context reading (before refactoring)
- Evidence-based completion (objective metrics)
- Incremental commits (not amendments)
- Multiple test runs (3-5x to catch flakiness)

**User is correct: These are regressions, not new bugs**

- format_go self-modification: FOUR previous attempts
- Flaky tests: Predictable (timeouts under load)
- Coverage gaps: Systematic (15+ packages)
- Slow execution: Obvious (89s vs 20s target)

**We must do better. This document exists to prevent repeat mistakes.**
