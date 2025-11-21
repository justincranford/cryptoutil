# Cryptoutil Testing Infrastructure TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: November 10, 2025
**Status**: CRITICAL BLOCKER - GORM AutoMigrate failing in identity integration tests. Testing infrastructure improvements completed - fuzz and benchmark testing implemented for cryptographic operations. Test file organization audit and migration completed.

---

## 🔴 CRITICAL - Active Blockers

### Task TB1: Fix GORM AutoMigrate Failure in Identity Integration Tests

- **Description**: GORM AutoMigrate fails with "no such table: main.users" error when trying to create User table in SQLite
- **Current State**: **BLOCKED** - Integration tests cannot seed test database
- **Root Cause Investigation**:
  - Error: `SQL logic error: no such table: main.users (1)` occurs during User table creation
  - Not a foreign key issue (PRAGMA foreign_keys OFF doesn't help)
  - Not a migration ordering issue (User is first model migrated)
  - Suggests circular dependency or schema generation issue in User domain model
  - User model has complex structure: embedded Address, multiple unique indexes, OIDC claims
- **Recent Fixes Applied**:
  - ✅ Removed `gorm.Model` duplication from all 8 domain models (commit bcd2171)
  - ✅ Changed User ID type from `gorm:"type:uuid"` to `gorm:"type:text"` for SQLite compatibility
  - ✅ Added missing DBUpdatedAt field to User model (avoid conflict with OIDC UpdatedAt claim)
  - ✅ Added PRAGMA foreign_keys OFF during migrations
  - ✅ Changed AutoMigrate to iterate models individually for better error reporting
  - ❌ Still failing with same error
- **Investigation Steps Remaining**:
  - Try migrating User model alone in minimal test case
  - Check if embedded Address struct causes issues
  - Try removing unique indexes temporarily to isolate problem
  - Check GORM SQLite driver compatibility with UUID fields
  - Consider using GORM Debug mode to see generated SQL
  - Review GORM issues for similar "no such table" errors during CREATE TABLE
- **Files**:
  - `internal/identity/domain/user.go` - User model with complex structure
  - `internal/identity/repository/factory.go` - AutoMigrate implementation
  - `internal/identity/integration/integration_test.go` - Integration test setup
- **Impact**: Blocks Task 10.5 completion, prevents all identity integration testing
- **Priority**: **CRITICAL** - Blocks entire identityV2 task chain (10.5 → 10.6 → 10.7 → 11-20)
- **Timeline**: Must resolve before proceeding to Task 10.6
- **Commits**:
  - `628f290` - Task 10.5 partial progress (health endpoints, authorize redirect, PKCE)
  - `bcd2171` - Domain model gorm.Model removal and fixes (still blocked)

---

## 🟡 MEDIUM - Testing Infrastructure Improvements

### Task T4: Implement Coverage Trend Analysis

- **Description**: Add coverage trend analysis to CI workflow to track coverage changes over time
- **Current State**: Basic coverage collection implemented, trend analysis not yet added
- **Proposed Implementation**:
  - Calculate current coverage percentage from `go tool cover -func` output
  - Download previous run's coverage data from artifacts
  - Compare current vs previous coverage and calculate difference
  - Display trend indicators (📈 increased, 📉 decreased, ➡️ unchanged, 📊 baseline)
  - Store current coverage for next run comparison
  - Show trend in GitHub Actions summary with visual indicators
- **Files**: `.github/workflows/ci-coverage.yml`
- **Expected Outcome**: Track coverage improvements/declines over time, provide data for coverage decisions
- **Priority**: Medium - Testing metrics enhancement
- **Dependencies**: ci-coverage.yml workflow completion

---

### Implementation Priority Recommendations

```text

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
