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

### Task T7: PostgreSQL Schema-Based Test Isolation
- **Description**: Implement and validate PostgreSQL schema-based isolation for concurrent testing
- **Current State**: ✅ **COMPLETED** - Schema isolation fully implemented with automatic schema creation
  - Basic schema approach implemented in `config_test_util.go`
  - **Additional fix implemented**: Added schema creation logic in `sql_provider.go` to create PostgreSQL schemas before applying migrations
  - Verified PostgreSQL does NOT auto-create schemas when referenced in `search_path`
  - Migrations now run in correct isolated schemas
  - Schema logging updated to work with non-public schemas
  - GORM configuration validated for schema isolation
  - Tested with GitHub Actions concurrency tests - UNIQUE constraint violations resolved
- **Action Items**:
  - ✅ Test current schema implementation with concurrent tests
  - ✅ Verify if PostgreSQL auto-creates schemas when referenced in `search_path` (confirmed: NO)
  - ✅ Check if migrations run in correct schema when `search_path` is set
  - ✅ Update schema logging code to work with non-public schemas
  - ✅ Modify migrations to use dynamic schema names if needed (not needed - search_path works)
  - ✅ Validate GORM configuration works with schema isolation
  - ✅ Test with actual GitHub Actions concurrency tests
- **Files**: `internal/common/config/config_test_util.go`, `internal/server/repository/sqlrepository/sql_provider.go`, PostgreSQL migration files
- **Expected Outcome**: Reliable concurrent testing without database conflicts
- **Priority**: High - Fixes current CI test failures
- **Dependencies**: None - Can be tested immediately

---

## Quick Reference

### Testing Commands
```powershell
# Test ZAP fix with quick scan
.\scripts\run-act-dast.ps1 -ScanProfile quick -Timeout 600

# Verify report generation
ls .\dast-reports\*.html, .\dast-reports\*.json, .\dast-reports\*.md
```
