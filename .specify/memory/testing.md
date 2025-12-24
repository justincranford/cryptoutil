# Testing Standards and Best Practices - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/03-02.testing.instructions.md`

## Test Concurrency Requirements

### Concurrent Execution - MANDATORY

**ALWAYS use concurrent test execution, NEVER disable parallelization**

**Rationale**:

1. **Fastest test execution**: Parallel tests = faster feedback loop
2. **Reveals production bugs**: Race conditions, deadlocks, data conflicts exposed
3. **Production validation**: If tests can't run concurrently, production code can't either
4. **Quality assurance**: Concurrent tests = higher confidence in code correctness and robustness

**Correct Execution Commands**:

```bash
# ✅ CORRECT - Concurrent with shuffle (recommended)
go test ./... -cover -shuffle=on

# ✅ CORRECT - Default concurrent execution
go test ./...

# ❌ WRONG - Sequential package execution (hides concurrency bugs!)
go test ./... -p=1  # NEVER DO THIS

# ❌ WRONG - Sequential test execution within packages
go test ./... -parallel=1  # NEVER DO THIS
```

**Race Detection**:

```bash
# Requires CGO_ENABLED=1 (only exception to CGO ban)
go test -race -count=2 ./...
```

---

## Coverage Targets - NO EXCEPTIONS

### Package Classification for Coverage Targets

| Category | Packages | Minimum Coverage | Rationale |
|----------|----------|------------------|-----------|
| Production | internal/{jose,identity,kms,ca} | 95% | Business logic, customer-facing |
| Infrastructure | internal/cmd/cicd/* | 98% | Critical tooling, self-modifying |
| Utility | internal/shared/*, pkg/* | 98% | Foundational code, widely reused |
| Main Functions | cmd/*/main.go | 0%* | *if internalMain() ≥95% |

### Why "No Exceptions" Rule Matters

**UNACCEPTABLE Rationalizations**:

- ❌ "70% is better than 60%" → Still 25 points below target (25% technical debt)
- ❌ "This package is mostly error handling" → Add error path tests
- ❌ "This is just a thin wrapper" → Still needs 95% coverage
- ❌ "We'll improve it later" → Incremental improvements accumulate debt

**Enforcement Pattern**:

```bash
# DON'T: Celebrate improvement without meeting target
coverage_before=60.0
coverage_after=70.0
echo "✅ Improved by 10 percentage points!"  # ❌ Still 25 points below target

# DO: Enforce target, reject anything below 95%
if [ "$coverage" -lt 95 ]; then
    echo "❌ BLOCKING: Coverage $coverage% < 95% target"
    echo "Required: Write tests for RED lines in coverage HTML"
    exit 1
fi
```

---

## main() Function Testing Pattern - MANDATORY

### Testable main() Architecture

**ALL main() functions MUST delegate to co-located testable internalMain()**

```go
// ✅ CORRECT - Thin main() delegates to testable internalMain()
func main() {
    os.Exit(internalMain(os.Args, os.Stdin, os.Stdout, os.Stderr))
}

// internalMain is testable - accepts injected dependencies
func internalMain(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
    // All logic here - fully testable with mocks
    if len(args) < 2 {
        fmt.Fprintln(stderr, "usage: cmd <arg>")
        return 1
    }
    // ... business logic
    return 0
}

// ❌ WRONG - Logic directly in main() blocks testing
func main() {
    if len(os.Args) < 2 {  // ❌ Untestable - hardcoded os.Args
        fmt.Fprintln(os.Stderr, "usage: cmd <arg>")  // ❌ Untestable - hardcoded os.Stderr
        os.Exit(1)  // ❌ Untestable - terminates process
    }
}
```

**Why This Pattern is MANDATORY**:

- **95%+ coverage achievable**: main() 0% acceptable when internalMain() ≥95%
- **Dependency injection**: Tests inject mocks for args, stdin, stdout, stderr
- **Exit code testing**: Tests verify return codes without terminating test process
- **Happy/sad path testing**: Test all branches (missing args, invalid input, success cases)

**Commands Requiring Pattern**:

- cmd/cicd/cicd.go → internalMain() testable function
- cmd/cryptoutil/main.go → internalMain() testable function
- cmd/workflow/main.go → internalMain() testable function
- internal/cmd/cicd/*/main.go → internalMain() testable function

**Testing Pattern**:

```go
func TestInternalMain_HappyPath(t *testing.T) {
    args := []string{"cmd", "arg1"}
    stdin := strings.NewReader("")
    stdout := &bytes.Buffer{}
    stderr := &bytes.Buffer{}

    exitCode := internalMain(args, stdin, stdout, stderr)

    require.Equal(t, 0, exitCode)
    require.Contains(t, stdout.String(), "success")
}

func TestInternalMain_MissingArgs(t *testing.T) {
    args := []string{"cmd"}  // Missing required arg
    stderr := &bytes.Buffer{}

    exitCode := internalMain(args, nil, nil, stderr)

    require.Equal(t, 1, exitCode)
    require.Contains(t, stderr.String(), "usage")
}
```

---

## Coverage Analysis Workflow - MANDATORY

### Pre-Writing Test Analysis

**ALWAYS analyze baseline coverage BEFORE writing tests**

**Workflow Steps**:

1. **Generate baseline**: `go test ./pkg -coverprofile=./test-output/coverage_pkg.out`
2. **Analyze uncovered lines**: `go tool cover -html=./test-output/coverage_pkg.out -o ./test-output/coverage_pkg.html`
3. **Identify specific gaps**: Open HTML, find RED (uncovered) lines
4. **Target specific branches**: Write tests ONLY for identified uncovered code
5. **Verify improvement**: Re-run coverage, confirm gaps filled

**Why This Matters**:

- Prevents wasted effort writing tests for already-covered code
- Targets exact missing branches, not guesswork
- Measurable progress per test addition
- Avoids trial-and-error test deletion cycles

**Example Gap Analysis**:

```bash
# Step 1: Generate baseline
go test ./internal/jose -coverprofile=./test-output/coverage_jose_baseline.out

# Step 2: Check per-function coverage (PowerShell)
go tool cover -func ./test-output/coverage_jose_baseline.out | Where-Object { $_ -match 'jose.*\.go:' -and $_ -match '(\d+\.\d+)%' -and [double]$Matches[1] -lt 90.0 }

# Step 3: Visual analysis
go tool cover -html=./test-output/coverage_jose_baseline.out -o ./test-output/coverage_jose_baseline.html
# Open HTML, find RED lines in low-coverage functions

# Step 4: Write targeted tests for RED lines only

# Step 5: Re-run and verify
go test ./internal/jose -coverprofile=./test-output/coverage_jose_new.out
```

### Test Output File Locations

**MANDATORY Directory Structure**:

- **Test outputs**: `./test-output/` (project root)
- **Test fixtures**: `./testdata/` or `pkg/testdata/`

**File Naming Convention**:

- ✅ `./test-output/coverage_<package>_<variant>.out`
- ✅ `./test-output/coverage_jose_baseline.out`
- ✅ `./test-output/coverage_identity_auth.out`
- ❌ `internal/jose/test-coverage.out` (pollutes source tree)
- ❌ `coverage.out` (non-descriptive)

---

## Timeout Configuration - MANDATORY

### HTTP Client and Context Timeouts

**Network Operations Require Generous Timeouts**

```go
// ❌ DON'T: No timeout or short timeout
client := &http.Client{}  // Infinite wait
client := &http.Client{Timeout: 1 * time.Second}  // Too short for CI/CD

// ✅ DO: Use 5s+ timeouts for network operations
client := &http.Client{
    Transport: &http.Transport{TLSClientConfig: tlsConfig},
    Timeout: 5 * time.Second,  // Generous for shared resources
}

// ❌ DON'T: Bare requests without timeout context
req, _ := http.NewRequest("GET", url, nil)
resp, err := client.Do(req)

// ✅ DO: Always use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
```

### Retry Loops with Exponential Backoff

```go
// ❌ DON'T: Retry blindly on all errors
for retry := 0; retry < maxRetries; retry++ {
    resp, err := client.Do(req)
    if err == nil { break }
    time.Sleep(delay)  // What if error is permanent?
}

// ✅ DO: Check error type and use exponential backoff
for retry := 0; retry < maxRetries; retry++ {
    resp, err := client.Do(req)
    if err == nil && resp.StatusCode == http.StatusOK { break }

    var ne net.Error
    if errors.As(err, &ne) && !ne.Timeout() {
        return err  // Permanent error, don't retry
    }

    time.Sleep(time.Duration(retry+1) * baseDelay)  // Exponential backoff
}
```

### Why Generous Timeouts Matter

| Factor | Impact | Recommendation |
|--------|--------|----------------|
| Shared CI/CD resources | Variable CPU steal time | Use 5s+ for network ops |
| Network latency | Even loopback has overhead | Use 10s+ for TLS handshakes |
| Parallel execution | Resource contention | Add 50-100% safety margin |
| TLS handshakes | CPU-intensive crypto | Use 10-15s in GitHub Actions |

**Flaky Test Detection**:

```bash
# ❌ DON'T: Run once and assume stable
go test ./internal/server/application/

# ✅ DO: Run 3-5 times to detect flakiness
go test -count=3 ./internal/server/application/
go test -count=5 ./internal/shared/crypto/certificate/
```

**Pattern Recognition**: Timeout errors (context deadline exceeded, i/o timeout, EOF on TLS) under parallel execution = insufficient timeout, NOT test bug.

---

**GitHub Actions Performance**: Expect 2.5-5× slower than local (shared CPU, network latency). Apply generous timeouts: 5-10s for network ops, 10-15s for TLS. Infrastructure overhead dominates - optimize locally, accept GitHub slowdown.

**Correct t.Parallel() Behavior**: Individual tests show 0.00s (fast logic), total package time includes infrastructure (setup, pause/resume coordination, teardown). Example: 0.00s individual + 303s total = infrastructure overhead, NOT test bug.

**See**: `github.md` for complete GitHub Actions configuration patterns

---

## Table-Driven Test Patterns - MANDATORY

### Happy Paths Pattern

| Local Development | <2s | 1x (baseline) | Dedicated resources |
| GitHub Actions (typical) | 5-10s | 2.5-5x | Shared CPU, network latency |
| GitHub Actions (extreme) | 303s | 150x | Cold starts, container overhead |

**Evidence** (from archived WORKFLOW-sqlrepository-TEST-TIMES.md):

- Local execution: <2s (same tests, same code)
- GitHub execution: 303s → 601s before t.Parallel() optimization
- Optimization with t.Parallel(): 601s → 303s (50% reduction)
- Individual test timing: 0.00s each (infrastructure overhead, NOT test code)

### Timing Strategy

**Local Development**:

- Fast iteration with minimal timeouts
- Unit tests: 1-5s typical per package
- Network operations: 2-5s typical

**GitHub Actions**:

- Apply 2.5-3× multiplier minimum to local timings
- Add 50-100% safety margin for reliability
- Unit tests: 5-15s per package target
- Network operations: 5-10s (general), 10-15s (TLS handshakes)
- Health checks: 300s (5 minutes) for full Docker Compose stack

### Parallel Test Execution Pattern

**Correct t.Parallel() Behavior in Test Output**:

```
=== RUN TestName
=== PAUSE TestName          ← Parallel tests PAUSE (expected)
=== CONT TestName           ← Resume after all tests PAUSEd (expected)
--- PASS: TestName (0.00s)  ← Individual test executes instantly (expected)
```

**Why Tests Show 0.00s**:

- Individual test logic executes quickly (milliseconds)
- Total package time includes: setup, pause/resume coordination, teardown, infrastructure
- 0.00s individual + 303s total = infrastructure overhead, NOT test performance bugs

### Optimization Guidelines

1. **Apply t.Parallel()**: Typical 50% speedup (observed: 601s → 303s)
2. **Don't over-optimize test code**: Infrastructure overhead dominates execution time
3. **Focus on local speed**: Fast local tests = better developer experience
4. **Accept GitHub slowdown**: 2-3× multiplier is normal and expected
5. **Use timeouts wisely**: Too short = flaky tests, too long = slow failure detection
6. **Probabilistic execution**: Use TestProbTenth/TestProbQuarter for algorithm variants

---

## Table-Driven Test Patterns - MANDATORY

### Happy Paths Pattern

```go
func TestHappyPaths(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name string
        alg  Algorithm
        // ... test parameters
    }{
        {name: "HMAC_HS256", alg: AlgHS256, ...},
        {name: "HMAC_HS384", alg: AlgHS384, ...},
        {name: "RSA_2048", alg: AlgRS256, ...},
        // ... all variants
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // test logic
        })
    }
}
```

### Sad Paths Pattern

```go
func TestSadPaths(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name        string
        input       Input
        expectedErr string
    }{
        {name: "NilKid", input: Input{Kid: nil}, expectedErr: "kid required"},
        {name: "NilAlg", input: Input{Alg: nil}, expectedErr: "alg required"},
        {name: "NilKey", input: Input{Key: nil}, expectedErr: "key required"},
        // ... all error cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // test logic
        })
    }
}
```

---

## Probability-Based Test Execution - MANDATORY

### When to Use Statistical Sampling

**Use Probability-Based Testing**:

- ✅ Multiple key sizes of same algorithm (RSA 2048/3072/4096, AES 128/192/256)
- ✅ Multiple variants of same operation (HMAC-SHA256/384/512, ECDSA P-256/384/521)
- ✅ Large test suites (>50 test cases with redundant coverage)

**NEVER Use Probability-Based Testing**:

- ❌ Fundamentally different algorithms (RSA vs ECDSA vs EdDSA - always test all)
- ❌ Business logic branches (error paths, edge cases - always test all)
- ❌ Small test suites (<20 cases - overhead not worth it)

### Magic Constants

**Defined in `internal/shared/magic/magic_testing.go`**:

- `TestProbAlways = 100` (100%) - Base algorithms, critical paths
- `TestProbQuarter = 25` (25%) - Important variants
- `TestProbTenth = 10` (10%) - Redundant key size variants

**Rationale**:

- **Faster test execution**: Running all key size variants adds minimal coverage value but significant time
- **Comprehensive vs. sampling**: Base algorithms use TestProbAlways (100%), variants use TestProbTenth (10%) or TestProbQuarter (25%)
- **Maintains quality**: Statistical sampling ensures bugs eventually caught without running all variants every time
- **Scales better**: Projects with 100+ algorithm variants become testable in reasonable time

---

## Test Execution Time Targets

### Unit Test Timing Requirements

**MANDATORY Timing Targets**:

- **Per-package**: <15 seconds (unit tests only)
- **Full suite**: <180 seconds (3 minutes, unit tests only)
- **Integration/E2E**: Excluded from strict timing (Docker startup overhead acceptable)
- **Probabilistic execution**: MANDATORY for packages approaching 15s limit

**Enforcement**:

```bash
# Measure per-package timing
go test -json ./internal/jose | jq -r 'select(.Action == "pass" and .Package != null) | "\(.Package): \(.Elapsed)s"'

# Fail if any package >15s
if [ "$elapsed" -gt 15 ]; then
    echo "❌ BLOCKING: Package $package took ${elapsed}s > 15s target"
    echo "Required: Apply probabilistic execution to algorithm variants"
    exit 1
fi
```

---

## Test Data Isolation Requirements

### Unique Values and Dynamic Ports

**MANDATORY Patterns**:

- **Unique values**: UUIDv7 for all test data (thread-safe, process-safe)
- **Dynamic ports**: port 0 pattern for all test servers
- **TestMain for dependencies**: Start once per package (PostgreSQL containers, service dependencies)

**Test Values - Two Options**:

- **Option A**: Generate once, reuse: `id := googleUuid.NewV7()` then use `id` in test cases
- **Option B**: Magic values from `internal/identity/magic` package

**NEVER**:

- ❌ Inline hardcoded UUIDs, strings
- ❌ Call `NewV7()` twice expecting same result
- ❌ Hardcode port numbers (use port 0)

### Real Dependencies vs Mocks

**Preferred Approach**:

- ✅ **Real dependencies**: Test containers (PostgreSQL, Otel Collector Contrib), real crypto, real HTTP servers
- ✅ **Rationale**: Real dependencies reveal production bugs; mocks hide integration issues

**Mocks ONLY for**:

- Hard-to-reach corner cases
- External services that can't run locally (email, SMS, cloud-only APIs)

---

## Core Test Rules

1. **Coverage targets**: 95%+ production, 98%+ infrastructure/utility
2. **Table-driven tests**: ALWAYS use with `t.Parallel()` for orthogonal data
3. **Assertions**: ALWAYS use testify `require` for fast fail
4. **Exit codes**: NEVER os.Exit() in test code (only main() calls os.Exit())
5. **Response bodies**: ALWAYS check and close: `defer require.NoError(t, resp.Body.Close())`
6. **Error handling**: ALWAYS check errors from helper functions
7. **Mutation testing**: ≥85% Phase 4, ≥98% Phase 5+ gremlins score per package
8. **Temporary files**: NEVER create temporary test files requiring deletion
9. **Test values**: NEVER hardcode - use UUIDv7 or magic constants
10. **Concurrency**: ALWAYS use concurrent execution, NEVER `-p=1` or `-parallel=1`

---

## Test File Organization

| Type | Suffix | Example | Purpose |
|------|--------|---------|---------|
| Unit | `_test.go` | `calc_test.go` | Standard unit tests |
| Bench | `_bench_test.go` | `calc_bench_test.go` | Performance benchmarks |
| Fuzz | `_fuzz_test.go` | `calc_fuzz_test.go` | Fuzz testing |
| Property | `_property_test.go` | `calc_property_test.go` | Property-based testing |
| Integration | `_integration_test.go` | `api_integration_test.go` | Integration tests |

---

## Race Condition Prevention

**MANDATORY Rules**:

- NEVER write to parent scope variables in parallel sub-tests
- NEVER use t.Parallel() with global state manipulation (os.Stdout, env vars)
- ALWAYS use inline assertions: `require.NoError(t, resp.Body.Close())`
- ALWAYS create fresh test data per test case (new sessions, UUIDs)
- ALWAYS protect shared maps/slices with sync.Mutex or sync.Map
- NEVER compare database timestamps against `time.Now()` in concurrent tests

**Detection**:

```bash
# Requires CGO_ENABLED=1 (only exception to CGO ban)
go test -race -count=2 ./...
```

---

## Mutation Testing - MANDATORY

### Configuration and Phased Targets

**Phased Targets**:

- **Phase 4**: ≥85% efficacy (early implementation quality baseline)
- **Phase 5+**: ≥98% efficacy (production-ready quality)

**Recommended `.gremlins.yaml`**:

```yaml
threshold:
  efficacy: 85  # Phase 4 target, raise to 98 in Phase 5+
  mutant-coverage: 90  # Target: ≥90% mutant coverage

workers: 4  # Parallel mutant execution
test-cpu: 2  # CPU per test run
timeout-coefficient: 2  # Timeout multiplier
```

**Execution**:

```bash
# Exclude integration tests
gremlins unleash --tags=!integration

# Focus on specific packages
gremlins unleash ./internal/jose
```

**Scope**:

- Focus on: Business logic, parsers, validators, crypto operations
- Exclude: Generated code, vendor directories, tests

**GitHub Actions Optimization**:

- Package-level parallelization (matrix strategy)
- Per-package timeout (15 minutes max)
- Total execution target: <20 minutes

**Windows Compatibility**:

- **Known Issue**: gremlins v0.6.0 panics on Windows in some scenarios
- **Workaround**: Use CI/CD (Linux) for mutation testing until Windows compatibility verified

---

## Benchmarking - Mandatory for Crypto

**ALWAYS create benchmarks for cryptographic operations**

**File Suffix**: `_bench_test.go`

**Execution**:

```bash
go test -bench=. -benchmem ./pkg/crypto
```

**Pattern**:

```go
func BenchmarkAESEncrypt(b *testing.B) {
    key := make([]byte, 32)
    plaintext := make([]byte, 1024)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = encrypt(key, plaintext)
    }
}
```

**Coverage**: Key generation, encryption/decryption, signing/verification, hashing (happy and sad paths)

---

## Fuzz Testing - MANDATORY

**CRITICAL: Fuzz tests MUST ONLY contain fuzz functions (Fuzz*)**

**Requirements**:

- Use `//go:build !fuzz` tag to exclude property tests from fuzz test runs
- Fuzz test names MUST be unique, NOT substrings of others (e.g., `FuzzHKDFAllVariants` not `FuzzHKDF`)
- ALWAYS run from project root: `go test -fuzz=FuzzXXX -fuzztime=15s ./path`
- Minimum fuzz time: 15 seconds per test
- Use unquoted names, PowerShell `;` for chaining

**File Organization**:

- `*_fuzz_test.go`: ONLY fuzz functions (FuzzX), no unit tests
- `*_test.go`: Unit, integration, table-driven tests

**Coverage**: Parsers, validators, input handlers

---

## Property-Based Testing - MANDATORY

**Library**: [gopter](https://github.com/leanovate/gopter)

**File Suffix**: `_property_test.go`

**Pattern**:

```go
func TestEncryptionRoundTrip(t *testing.T) {
    properties := gopter.NewProperties(nil)
    properties.Property("encrypt then decrypt returns original", prop.ForAll(
        func(plaintext []byte) bool {
            ciphertext, _ := Encrypt(key, plaintext)
            result, _ := Decrypt(key, ciphertext)
            return bytes.Equal(plaintext, result)
        },
        gen.SliceOf(gen.UInt8()),
    ))
    properties.TestingRun(t)
}
```

**Coverage**: Invariants, mathematical properties (encryption(decryption(x)) == x, sign(verify(x)) == x)

---

## Test File Size Limits - MANDATORY

### Strict Enforcement

| Limit | Lines | Action Required |
|-------|-------|-----------------|
| Soft | 300 | Ideal target |
| Medium | 400 | Acceptable with justification |
| Hard | 500 | NEVER EXCEED - refactor required |

**Why Size Limits Matter**:

- Faster LLM processing and token usage
- Easier human review and maintenance
- Better test organization and discoverability
- Forces logical test grouping

**Refactoring Strategies** (when file exceeds 400 lines):

1. Split by functionality: `jwk_util_test.go` → `jwk_util_create_test.go` + `jwk_util_validate_test.go`
2. Split by algorithm type: `crypto_test.go` → `crypto_rsa_test.go` + `crypto_ecdsa_test.go`
3. Extract test helpers to `*_test_util.go` files
4. Move integration tests to `*_integration_test.go`

**Example**:

```
❌ jwk_util_test.go (1371 lines) - VIOLATES HARD LIMIT

✅ jwk_util_create_test.go (280 lines) - CreateJWK* functions
✅ jwk_util_validate_test.go (250 lines) - validateOrGenerate* functions
✅ jwk_util_extract_test.go (180 lines) - Extract* functions
✅ jwk_util_is_test.go (150 lines) - Is* functions
✅ jwk_util_test_util.go (120 lines) - Shared test helpers
```

---

## Dynamic Port Allocation - MANDATORY

**ALWAYS use port 0 and extract actual assigned port**

**Why**: Hard-coded ports cause port conflicts when tests run in parallel. Dynamic allocation ensures each test gets a unique port.

**Pattern**:

```go
// Start server with port 0
listener, err := net.Listen("tcp", "127.0.0.1:0")
require.NoError(t, err)

// Extract actual port
actualPort := listener.Addr().(*net.TCPAddr).Port

// Use actualPort in test HTTP requests
resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api", actualPort))
```

---

## Common Testing Anti-Patterns - NEVER DO THESE

### Anti-Pattern 1: Writing Tests Without Baseline Coverage Analysis

**Mistake**: Adding 60+ tests without analyzing baseline coverage HTML first
**Result**: 0% coverage improvement - tests duplicated already-covered paths
**Correct Pattern**: ALWAYS generate baseline, analyze HTML for RED lines, target specific gaps

### Anti-Pattern 2: Individual Test Functions Instead of Table-Driven

**Mistake**: Creating TestFunc_Variant1, TestFunc_Variant2, TestFunc_Variant3 as separate functions
**Result**: 1371-line test file (2.7x hard limit), maintenance nightmare, slower LLM processing
**Correct Pattern**: Use table-driven tests with variants as rows

### Anti-Pattern 3: Test Outputs in Source Directories

**Mistake**: Placing coverage files in internal/jose/test-coverage.out
**Result**: Source tree pollution, confuses version control, clutters package directories
**Correct Pattern**: ALWAYS use ./test-output/ directory for all test artifacts

### Anti-Pattern 4: Exceeding File Size Limits

**Mistake**: Allowing test files to grow to 1371 lines without refactoring
**Result**: Slower LLM processing, harder maintenance, poor discoverability
**Correct Pattern**: Split at 500 lines using functional grouping

### Anti-Pattern 5: Trial-and-Error Test Writing

**Mistake**: Writing tests, checking coverage, writing more tests, repeat cycle
**Result**: Wasted effort, unclear progress, no measurable improvement strategy
**Correct Pattern**: Baseline → HTML analysis → targeted test → verify improvement cycle

**Key Insight**: Coverage ≠ test count. Many tests can add 0% coverage if exercising already-covered code paths. HTML baseline analysis eliminates guesswork and waste.

---

## cicd Utility Testing Requirements

**Commands organized as `internal/cmd/cicd/<snake_case>/` subdirectories**

**Requirements**:

- EVERY command MUST exclude its own subdirectory (self-exclusion pattern)
- Define exclusion in `internal/common/magic/magic_cicd.go`
- Add self-exclusion test to verify pattern works
- Target 95%+ coverage per command subdirectory
