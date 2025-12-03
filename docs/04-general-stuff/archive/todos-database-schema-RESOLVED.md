# Database Schema Issues - RESOLVED (see Transaction Isolation Issues below)

## ✅ Completed GORM Column Mismatches

All GORM column tags fixed in domain models and SQL migrations:

1. ✅ **MTLSDomains** - Fixed column name in auth_profile.go (commit 6f198651)
2. ✅ **PhoneVerified** - Fixed column name in user.go (commit e2ed567e)
3. ✅ **ConsentScreen1Text, ConsentScreen2Text** - Fixed column names in client_profile.go (commit e2ed567e)
4. ✅ **RevokedAt** - Fixed column name in session.go (commit f1cd0913)
5. ✅ **CodeChallenge, Nonce** - Fixed column names and added missing column to migration (commit f1cd0913)
6. ✅ **Test Uniqueness** - Added UUIDv7 suffixes to all test data (commit f1cd0913)
7. ✅ **Session Foreign Keys** - Added user creation before session in tests (commit f1cd0913)

## 🔄 Transaction Isolation Issues (NEW - In Progress)

### Problem: TestTransactionRollback Failure

- **Symptom**: Transaction rollback appears successful but rolled-back data is still visible to subsequent reads
- **Root Cause**: SQLite shared in-memory database (:memory: with cache=shared) + GORM transaction isolation
- **Investigation Commits**: f1cd0913, 5a86bbdf
- **Current Status**: Diagnosing transaction isolation behavior with logging

### Problem: Shared Database Pollution

- **Symptom**: CRUD tests seeing data from other parallel tests (TestUserRepositoryCRUD expects 1 item, finds 2-4)
- **Root Cause**: All t.Parallel() tests share same :memory: database instance
- **Affected Tests**: TestUserRepositoryCRUD, TestClientRepositoryCRUD, TestMFAFactorRepositoryCRUD
- **Investigation Needed**: Unique database instances per test OR test cleanup strategy

## Investigation Leads

1. **GORM Transaction Implementation**: How does GORM handle SQLite transactions with WAL mode?
2. **SQLite Isolation Levels**: What isolation guarantees exist with shared cache mode?
3. **Connection Pool Settings**: MaxOpenConns=5 for GORM pattern - does this affect isolation?
4. **Alternative Approaches**:
   - Use unique file-based databases per test
   - Implement test cleanup with TRUNCATE/DELETE
   - Use database snapshots/rollback points

## ✅ Historical Issues (Archived - Fixed in commits f1cd0913, 5a86bbdf, 6a463f9f)

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
