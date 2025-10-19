# Cryptoutil Testing Infrastructure TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: October 17, 2025
**Status**: Testing infrastructure improvements completed - fuzz and benchmark testing implemented for cryptographic operations. Test file organization audit and migration completed. Performance testing framework fully implemented with k6 integration, automated CI/CD, and GitHub Pages dashboards.

---

## 🟡 MEDIUM - Testing Infrastructure Improvements

### Task T3: Optimize Test Parallelization Strategy
- **Description**: Fine-tune test parallelization for optimal CI performance
- **Current State**: Basic `-p=2` parallelization, but may not be optimal for all packages
- **Action Items**:
  - Balance test load across CI runners
  - Monitor and adjust based on CI performance metrics
- **Files**: `.github/workflows/ci.yml`, test execution analysis
- **Expected Outcome**: Faster CI execution with better resource utilization
- **Priority**: Medium - CI performance optimization
- **Dependencies**: Task T1/T2 completion

### Task T4: Local Docker Compose Testing and Workflow Re-enablement
- **Description**: Perform local testing and fixes of Docker Compose functionality before re-enabling disabled workflow steps
- **Current State**: Key DAST workflow steps disabled with `if: false` conditions due to Docker Compose issues
- **Action Items**:
  - Test Docker Compose locally using `.\scripts\run-act-dast.ps1` with different scan profiles
  - Verify all services start correctly (PostgreSQL, cryptoutil instances, OpenTelemetry, Grafana)
  - Fix any connectivity or configuration issues discovered during local testing
  - Ensure health checks pass for all services
  - Re-enable workflow steps by removing `if: false` conditions once local testing confirms functionality
  - Test workflow re-enablement with act before pushing changes
- **Files**: `.github/workflows/dockercompose.yml`, `deployments/compose/compose.yml`, `scripts/run-act-dast.ps1`
- **Expected Outcome**: Reliable Docker Compose deployment with all services functional
- **Priority**: High - CI/CD pipeline restoration
- **Testing Commands**:
  ```powershell
  # Test quick scan locally
  .\scripts\run-act-dast.ps1 -ScanProfile quick -Timeout 600

  # Test full scan locally  
  .\scripts\run-act-dast.ps1 -ScanProfile full -Timeout 900

  # Verify reports are generated
  ls .\dast-reports\*.html, .\dast-reports\*.json, .\dast-reports\*.md
  ```

---

## Quick Reference

### Testing Commands
```powershell
# Test ZAP fix with quick scan
.\scripts\run-act-dast.ps1 -ScanProfile quick -Timeout 600

# Verify report generation
ls .\dast-reports\*.html, .\dast-reports\*.json, .\dast-reports\*.md
```

### Performance Testing Commands
```bash
# Run quick performance test
go run ./scripts/run-performance-tests -profile quick

# Run full performance test with custom base URL
go run ./scripts/run-performance-tests -profile full -base-url https://api.example.com

# Analyze results and generate dashboard
go run ./scripts/analyze-performance-results

# View dashboard
start .\performance-reports\performance-dashboard.html
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
- **Completed**: ✅ Audited current test files and implemented migration strategy - separated mixed test types into dedicated files (_test.go for unit tests, _fuzz_test.go for fuzz tests, _bench_test.go for benchmarks)
