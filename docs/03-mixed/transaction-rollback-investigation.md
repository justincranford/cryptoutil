# Transaction Rollback Investigation

## Problem Statement

`TestTransactionRollback` creates entities within a GORM transaction, returns an error to trigger rollback, then verifies that the entities do NOT exist. **The test fails because the rolled-back entities are still visible after rollback**.

## Test Behavior

```text
Created user in transaction: ID=019aa93b-6e4f-7d0f-a9fc-2b510488a301, Sub=rollback-test-user-019aa93b-6e4d-781e-a417-f37e5d6c1a3f
Created client in transaction: ID=019aa93b-6e50-7ec2-98fe-8439701ca8c1, ClientID=rollback-test-client-019aa93b-6e4d-781e-a417-f37e5d6c1a3f
Expected error finding rolled-back user, but found user with ID: 019aa93b-6e4f-7d0f-a9fc-2b510488a301, Sub: rollback-test-user-019aa93b-6e4d-781e-a417-f37e5d6c1a3f
```

**Expected**: GetBySub returns "record not found" error
**Actual**: GetBySub successfully finds the user that was supposedly rolled back

## Current Architecture

### Database Configuration (`internal/identity/repository/database.go`)

```go
// SQLite connection pool settings for GORM transaction pattern.
sqliteMaxOpenConns = 5 // Balance between concurrency and resource usage.
sqliteMaxIdleConns = 5

// DSN handling
const (
    dsnMemory = ":memory:"
    dsnMemoryShared = "file::memory:?cache=shared"
)

// GORM configuration
db, err := gorm.Open(dialector, &gorm.Config{
    SkipDefaultTransaction: true,  // Prevent nested transaction deadlocks
})

// SQLite PRAGMA settings (applied to sql.DB before passing to GORM)
PRAGMA journal_mode=WAL;
PRAGMA busy_timeout = 30000;
```

### Transaction Implementation (`internal/identity/repository/factory.go`)

```go
func (f *RepositoryFactory) Transaction(ctx context.Context, fn func(context.Context) error) error {
    return f.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Store transaction DB in context so repositories can use it.
        txCtx := context.WithValue(ctx, txKey, tx)
        return fn(txCtx)
    })
}
```

### Repository Pattern (`internal/identity/repository/orm/transaction.go`)

```go
func getDB(ctx context.Context, baseDB *gorm.DB) *gorm.DB {
    if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
        return tx  // Use transaction DB from context
    }
    return baseDB  // Fallback to base DB for non-transactional operations
}
```

## Hypothesis: Read Isolation Issue

### Theory

SQLite's transaction isolation with shared cache + WAL mode may allow uncommitted reads from other connections:

1. **Connection 1** (transaction): Writes user/client, returns error → ROLLBACK
2. **Connection 2** (verification read): Sees uncommitted data from Connection 1
3. **MaxOpenConns=5**: Multiple connections can read from shared cache simultaneously

### SQLite Isolation Levels

From [SQLite docs on isolation](https://www.sqlite.org/isolation.html):

> "SQLite implements serializable transactions by actually serializing them."
> "WAL mode allows readers to proceed concurrently with a writer."

**Key insight**: WAL mode optimizes for read concurrency, which may allow stale reads during rollback.

### Shared Cache Mode

From [SQLite docs on shared cache](https://www.sqlite.org/sharedcache.html):

> "Shared cache mode is designed to reduce memory usage when multiple connections to the same database are required."
> "However, shared cache mode changes the semantics of ACID transactions in subtle ways."

**Potential issue**: `:memory:` database converted to `file::memory:?cache=shared` (line 59 of database.go) enables shared cache, which alters transaction semantics.

## Investigation Checklist

- [ ] **Test with unique database files**: Does `file:uuid.db?mode=memory` fix the issue?
- [ ] **Test without shared cache**: Does plain `:memory:` (no shared cache) fix the issue?
- [ ] **Test with MaxOpenConns=1**: Does single connection eliminate read-after-rollback?
- [ ] **Add explicit BEGIN/ROLLBACK logging**: Does GORM actually call ROLLBACK?
- [ ] **Test with explicit SAVEPOINT**: Does nested savepoint fix isolation?
- [ ] **Check GORM transaction state**: Does tx.Statement.DB show rolled back state?

## Potential Solutions

### Option 1: Unique Database Per Test

```go
func setupTestRepositoryFactory(t *testing.T, ctx context.Context) *RepositoryFactory {
    uuidSuffix := googleUuid.Must(googleUuid.NewV7()).String()
    dbConfig := &DatabaseConfig{
        Type: "sqlite",
        DSN:  "file:" + uuidSuffix + ".db?mode=memory",  // Unique in-memory file
    }
}
```

**Pros**: Complete isolation, no shared state
**Cons**: Higher memory usage, slower test execution

### Option 2: Explicit Cleanup

```go
func TestTransactionRollback(t *testing.T) {
    // ... create entities in transaction ...

    // After rollback, explicitly verify via raw SQL
    var count int64
    repoFactory.DB().Raw("SELECT COUNT(*) FROM users WHERE sub = ?", sub).Scan(&count)
    require.Equal(t, int64(0), count)
}
```

**Pros**: Diagnoses read isolation issue
**Cons**: Doesn't fix root cause

### Option 3: Disable Shared Cache

```go
// In database.go, don't convert :memory: to shared cache
dsn := cfg.DSN
// REMOVE THIS:
// if dsn == dsnMemory {
//     dsn = dsnMemoryShared
// }
```

**Pros**: Simplest change, may fix isolation
**Cons**: Each connection gets separate database (defeats purpose of tests)

### Option 4: Force Serializable Isolation

```go
// After opening SQLite database
sqlDB.Exec("PRAGMA read_uncommitted = 0;")
sqlDB.Exec("PRAGMA locking_mode = EXCLUSIVE;")
```

**Pros**: Strongest isolation guarantees
**Cons**: Defeats WAL concurrency benefits

## Next Steps

1. **Quick Test**: Remove shared cache conversion, rerun TestTransactionRollback
2. **Diagnostic Test**: Add ROLLBACK logging to confirm GORM behavior
3. **Isolation Test**: Try MaxOpenConns=1 to eliminate concurrent reads
4. **Alternative**: Use unique file-based databases per test as long-term solution

## References

- [SQLite Transaction Isolation](https://www.sqlite.org/isolation.html)
- [SQLite WAL Mode](https://www.sqlite.org/wal.html)
- [SQLite Shared Cache](https://www.sqlite.org/sharedcache.html)
- [GORM Transactions](https://gorm.io/docs/transactions.html)
- [04-01.sqlite-gorm.instructions.md](../.github/instructions/04-01.sqlite-gorm.instructions.md)
