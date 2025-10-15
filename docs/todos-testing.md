# Cryptoutil Testing Infrastructure TODOs

**Last Updated**: October 14, 2025
**Status**: Testing infrastructure improvements planned for Q4 2025

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

### Task T5: Integration Test Automation
- **Description**: Implement automated integration testing in CI pipeline
- **Current State**: Unit tests only, no integration testing
- **Action Items**:
  - Create integration test suite with database and external dependencies
  - Set up test database instances for CI
  - Implement API contract testing
  - Add integration tests to CI pipeline
- **Files**: Integration test files, CI workflow updates
- **Expected Outcome**: End-to-end testing validation
- **Priority**: Medium - Production readiness

### Task T6: Code Robustness Review for Brittleness Patterns
- **Description**: Review all production and main branch code for brittleness patterns that could cause failures under concurrent or parallel execution
- **Current State**: Recent concurrency test failures revealed brittleness in test code (shared database, non-unique names)
- **Action Items**:
  - Conduct systematic review of all production code for shared state assumptions
  - Identify code that assumes single-threaded execution or unique resources
  - Review database operations for proper isolation and transaction handling
  - Check for hardcoded resource names that could conflict in concurrent scenarios
  - Document findings and implement fixes for identified brittleness patterns
- **Files**: All production code files, database operations, resource management code
- **Expected Outcome**: More robust codebase resistant to concurrency-related failures
- **Priority**: High - Prevents production incidents from concurrency issues
- **Dependencies**: None - Can be done incrementally

---

## Quick Reference

### Testing Commands
```powershell
# Test ZAP fix with quick scan
.\scripts\run-act-dast.ps1 -ScanProfile quick -Timeout 600

# Verify report generation
ls .\dast-reports\*.html, .\dast-reports\*.json, .\dast-reports\*.md
```
