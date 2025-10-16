# Cryptoutil Testing Infrastructure TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: October 16, 2025
**Status**: Testing infrastructure improvements planned for Q4 2025 - Go test flags analysis added for optimization review

---

## 🟡 MEDIUM - Testing Infrastructure Improvements

### Task T1: Add Fuzz Testing for Security-Critical Code
- **Description**: Implement fuzz testing for cryptographic operations and key generation functions
- **Current State**: No fuzz tests implemented (no `func Fuzz*` functions found)
- **Action Items**:
  - Add fuzz tests for key generation functions in `pkg/keygen`
  - Add fuzz tests for cryptographic operations in `internal/crypto`
  - Configure appropriate fuzz timeouts (5-30 minutes)
  - Add fuzz testing to nightly/release CI pipeline
- **Files**: `*_test.go` files in `pkg/keygen/`, `internal/crypto/`
- **Expected Outcome**: Property-based testing for security-critical code paths
- **Priority**: Medium - Security testing enhancement
- **Dependencies**: Separate invocations required

### Task T2: Add Benchmark Testing Infrastructure
- **Description**: Implement performance benchmarking for key cryptographic operations
- **Current State**: No benchmark tests implemented (no `func Benchmark*` functions found)
- **Action Items**:
  - Add benchmarks for key generation performance
  - Add benchmarks for cryptographic operations
  - Configure benchmark testing in CI pipeline
  - Add performance regression detection
- **Files**: `*_test.go` files in `pkg/keygen/`, `internal/crypto/`
- **Expected Outcome**: Performance monitoring and regression detection
- **Priority**: Medium - Performance validation
- **Dependencies**: Separate invocations required

### Task T3: Optimize Test Parallelization Strategy
- **Description**: Fine-tune test parallelization for optimal CI performance
- **Current State**: Basic `-p=2` parallelization, but may not be optimal for all packages
- **Action Items**:
  - Analyze test execution times by package
  - Optimize `-p` flag usage (parallel packages vs parallel tests within packages)
  - Balance test load across CI runners
  - Monitor and adjust based on CI performance metrics
- **Files**: `.github/workflows/ci.yml`, test execution analysis
- **Expected Outcome**: Faster CI execution with better resource utilization
- **Priority**: Medium - CI performance optimization
- **Dependencies**: Task T1/T2 completion

### Task T4: Performance/Load Testing Framework
- **Description**: Implement comprehensive performance and load testing capabilities
- **Current State**: Basic unit tests only, no performance testing
- **Action Items**:
  - Set up performance testing framework (k6, Artillery, or custom)
  - Create load test scenarios for key API endpoints
  - Implement performance regression detection
  - Add performance testing to CI pipeline
- **Files**: Performance test scripts, CI workflow updates
- **Expected Outcome**: Automated performance validation and regression detection
- **Priority**: Medium - Production readiness

### Task T6: Go Test Flags Optimization & Best Practices Analysis
- **Description**: Analyze and optimize Go test flag usage for comprehensive testing coverage and CI efficiency
- **Current State**: Basic flag usage implemented, but comprehensive analysis needed for optimal combinations
- **Action Items**:
  - Document effects of key Go test flags (`-race`, `-covermode`, `-count`, `-p`)
  - Analyze flag interactions and problematic combinations
  - Optimize CI workflow flag usage for performance vs coverage balance
  - Establish testing flag best practices for the project
- **Files**: `.github/workflows/test.yml`, `.github/workflows/ci.yml`, test execution analysis
- **Expected Outcome**: Optimized test execution with comprehensive coverage and minimal redundancy
- **Priority**: Medium - Testing efficiency and reliability
- **Dependencies**: None

#### Go Test Flags Reference & Analysis

##### `-race` Flag Effects
- **Purpose**: Enables Go's race detector for concurrent memory access validation
- **Performance Impact**: 2-10x slower execution, 5-10x higher memory usage
- **Detection**: Identifies race conditions in concurrent code
- **Best Usage**: CI/CD pipelines, development testing, pre-release validation
- **Avoid In**: Production builds, performance benchmarks, memory-constrained environments

##### `-covermode=atomic` Effects
- **Purpose**: Thread-safe coverage counting for concurrent programs
- **Behavior**: Uses atomic operations for accurate coverage in multi-goroutine tests
- **Performance**: Minimal overhead, automatically enabled with `-race`
- **Alternatives**: `set` (fast but not thread-safe), `count` (slower but thread-safe)
- **Best Usage**: Always with concurrent tests, required for accurate coverage reporting

##### Problematic Flag Combinations (AVOID)
- **`-race` + `-bench`**: Race detector distorts benchmark timing results
- **`-race` + high `-cpu` values**: Overhead scales poorly (limit to `-cpu=4` max)
- **`-covermode=set` + concurrent tests**: Not thread-safe, causes inaccurate coverage
- **`-short` + comprehensive testing**: May miss race conditions that need load
- **`-race` + memory-intensive tests**: May cause OOM in constrained environments

##### Recommended Flag Combinations (SAFE)
- **`-race -covermode=atomic -count=2 -p=2`**: Comprehensive concurrent testing (CURRENT USAGE)
- **`-race -timeout=15m -v`**: Verbose race detection with timeout protection
- **`-cover -coverprofile=coverage.out -covermode=atomic`**: Standard coverage collection
- **`-fuzz=. -fuzztime=30s`**: Fuzz testing (separate from race detection)

##### Optimization Opportunities
- **Current Workflow Analysis**: Combined concurrency+race detection job is optimal
- **Performance Monitoring**: Track test execution times, identify bottlenecks
- **Flag Tuning**: Adjust `-count`, `-p`, and timeout values based on performance metrics
- **Separate Concerns**: Consider separating fast unit tests from slow comprehensive tests

---

## Quick Reference

### Testing Commands
```powershell
# Test ZAP fix with quick scan
.\scripts\run-act-dast.ps1 -ScanProfile quick -Timeout 600

# Verify report generation
ls .\dast-reports\*.html, .\dast-reports\*.json, .\dast-reports\*.md
```

---

## Go Testing Patterns & Conventions Reference

### Common Test File Naming Patterns

| Test Type | File Suffix | Package Suffix | Execution | Purpose |
|-----------|-------------|----------------|-----------|---------|
| Unit (external) | `_test.go` | `_test` | `go test` | Blackbox testing of public API |
| Unit (internal) | `_internal_test.go` | (same) | `go test` | Whitebox testing of internal logic |
| Integration | `_integration_test.go` | `_test` | `go test -tags=integration` | End-to-end with real dependencies |
| Benchmarks | `_bench_test.go` | `_test` or same | `go test -bench=.` | Performance testing |
| Fuzz | `_fuzz_test.go` | `_test` or same | `go test -fuzz=.` | Property-based testing |
| E2E | `e2e/*_test.go` | `e2e` | `go test ./e2e` | Full system testing |

### Testing Approach Examples

#### External Unit Tests (Blackbox)
```go
// calculator_test.go
package calculator_test

func TestAdd(t *testing.T) {
    calc := calculator.New()
    result := calc.Add(2, 3)
    assert.Equal(t, 5, result)
}
```

#### Internal Unit Tests (Whitebox)
```go
// calculator_internal_test.go
package calculator

func TestValidateInput(t *testing.T) {
    result := validateInput("test")  // Can test unexported function
    assert.True(t, result)
}
```

#### Integration Tests with Build Tags
```go
// +build integration
// calculator_integration_test.go
package calculator_test

func TestDatabaseOperations(t *testing.T) {
    // Tests that hit real database
    db := setupRealDatabase()
    defer db.Close()
    // ... integration test logic
}
```

#### Benchmark Tests
```go
func BenchmarkAdd(b *testing.B) {
    calc := calculator.New()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        calc.Add(1, 2)
    }
}
```

#### Fuzz Tests (Go 1.18+)
```go
func FuzzAdd(f *testing.F) {
    f.Add(1, 2)  // Seed corpus
    f.Fuzz(func(t *testing.T, a, b int) {
        result := calculator.Add(a, b)
        // Test mathematical properties
        assert.Equal(t, calculator.Add(b, a), result) // Commutative
    })
}
```

### Implementation Priority Recommendations

1. **High Priority**: External unit tests (`*_test.go`) - Establish API contracts
2. **Medium Priority**: Internal unit tests (`*_internal_test.go`) - Cover complex internals
3. **Medium Priority**: Integration tests (`*_integration_test.go`) - Validate real dependencies
4. **Low Priority**: Benchmarks (`*_bench_test.go`) - Performance optimization
5. **Low Priority**: Fuzz tests (`*_fuzz_test.go`) - Advanced property testing
6. **Optional**: E2E tests (`e2e/`) - Full system validation

### Current Project Assessment

- **Existing**: Mix of internal/external test patterns
- **testpackage linter**: Currently configured to allow internal testing
- **Recommendation**: Gradually migrate toward external testing for better API design
- **Next Steps**: Audit current test files and plan migration strategy
