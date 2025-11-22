# Database Schema Issues - TODO

## Critical Fixes Needed

### 1. AuthProfile.MTLSDomains Column Name Mismatch
- **Problem**: GORM converts `MTLSDomains` field to `m_tls_domains` column name (snake_case with acronym split)
- **Migration has**: `mtls_domains` column
- **GORM expects**: `m_tls_domains` column
- **Fix**: Add column tag `gorm:"column:mtls_domains;serializer:json"` to AuthProfile.MTLSDomains field
- **File**: `internal/identity/domain/auth_profile.go` line 40

### 2. Session.RevokedAt Column Missing
- **Problem**: Migration missing `revoked_at` column but domain model has field
- **Migration location**: `internal/identity/repository/migrations/0001_init.up.sql` line 165
- **Domain model**: `internal/identity/domain/session.go` line 37 has `RevokedAt *time.Time`
- **Fix**: Migration already has `revoked_at TIMESTAMP,` - check if GORM tag needed

### 3. Dirty Database Migration State
- **Problem**: Tests fail with "Dirty database version 1. Fix and force version."
- **Root cause**: Previous test runs left migration in dirty state
- **Fix**: Clear migration state between test runs or use fresh databases
- **Recommendation**: Each test should use unique database file or in-memory DB with shared cache

### 4. Test Data Uniqueness Violations
- **Problem**: UNIQUE constraint failures on `users.preferred_username` and `clients.client_id`
- **Root cause**: Parallel tests creating same test data simultaneously
- **Affected tests**:
  - TestUserRepositoryCRUD
  - TestClientRepositoryCRUD
  - TestTransactionIsolation
  - TestConcurrentTransactions
- **Fix**: Use UUIDv7 suffixes for all test data to ensure uniqueness
- **Example**: `preferred_username: "test-user-" + uuid.NewV7().String()`

## Test Execution Strategy

### Short-Term Fix
1. Fix column tags for MTLSDomains
2. Add UUIDv7 suffixes to all test fixture data
3. Clear test database state between runs

### Long-Term Solution
1. Each test creates isolated database (`:memory:` with unique DSN)
2. Migration applies cleanly to each test DB
3. No shared state between tests
4. Full parallel execution support

## Test Coverage Target
- **Current**: ~5-15% coverage for identity packages
- **Target**: ≥85% for infrastructure code (repositories, migrations)
- **Blocker**: Schema mismatches preventing test execution

## Related Files
- `internal/identity/domain/*.go` - Domain models with GORM tags
- `internal/identity/repository/migrations/0001_init.up.sql` - Database schema
- `internal/identity/storage/tests/*_test.go` - Integration tests
- `internal/identity/repository/database.go` - Database connection setup
