# Cryptoutil Testing Infrastructure TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: November 21, 2025
**Status**: GORM AutoMigrate blocker RESOLVED - Fixed UUID type handling, nullable foreign keys, and JSON serialization for SQLite cross-DB compatibility. TestHealthCheckEndpoints passes. Remaining integration test failures are application logic issues, not database issues.

---

## ✅ RESOLVED - GORM AutoMigrate SQLite Compatibility Issues

### Task TB1: Fix GORM AutoMigrate Failure in Identity Integration Tests - **RESOLVED**

- **Resolution Date**: November 21, 2025
- **Root Causes Identified**:
  1. **UUID Type Mismatch**: GORM `type:uuid` annotation incompatible with SQLite (no native UUID type)
  2. **Nullable Foreign Keys**: Pointer UUIDs (`*googleUuid.UUID`) caused "row value misused" errors
  3. **JSON Serialization**: GORM `type:json` annotation incompatible with SQLite (no native JSON type)
  4. **Shared Database**: Tests shared `:memory:` database causing unique constraint violations
- **Fixes Applied**:
  - ✅ Changed all UUID fields from `type:uuid` to `type:text` (13 fields across 7 models)
  - ✅ Created `NullableUUID` type with `sql.Scanner`/`driver.Valuer` for proper NULL handling
  - ✅ Replaced pointer UUID foreign keys with `NullableUUID` (5 fields across 4 models)
  - ✅ Changed JSON arrays from `type:json` to `serializer:json` (13 fields across 6 models)
  - ✅ Isolated test databases: `file::memory:?mode=memory&cache=private` per test
  - ✅ Updated `database.instructions.md` with cross-DB compatibility section
- **Test Results**:
  - ✅ TestHealthCheckEndpoints - **PASSES**
  - ⚠️ Remaining test failures are application logic issues (redirect handling, scope enforcement)
- **Documentation**: Complete cross-DB compatibility guide in `.github/instructions/01-04.database.instructions.md`
- **Commits**:
  - `5c8b4b38` - NullableUUID implementation and UUID type fixes
  - `4993e0c4` - JSON serializer fixes and database instructions update
  - `85549ad1` - Isolated database instances for tests
- **Impact**: Unblocks Task 10.5 completion and entire identityV2 task chain (10.5 → 10.6 → 10.7 → 11-20)

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
