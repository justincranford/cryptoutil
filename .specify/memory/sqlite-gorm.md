# SQLite Configuration with GORM - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/03-05.sqlite-gorm.instructions.md`

## Critical Architecture Constraints

### GORM Transaction Model

**Root Cause of Configuration Differences**:
- GORM explicit transactions (`db.Begin()`) require separate database connection from base operations
- MaxOpenConns=1 causes deadlock: base DB tries to create transaction, but connection already in use
- Solution: MaxOpenConns=5 allows concurrent transaction wrapper + internal operations

**Pattern Comparison**:

| Aspect | KMS Server (database/sql) | Identity Server (GORM) |
|--------|---------------------------|------------------------|
| ORM Layer | No (raw database/sql) | Yes (GORM ORM) |
| MaxOpenConns | 1 | 5 |
| Rationale | Single operation per connection | Transaction wrapper needs separate connection |
| Transaction Pattern | Manual tx := db.Begin() | GORM db.Begin() + context injection |
| Connection Lifecycle | Explicit open/close | GORM-managed pooling |

### CGO-Free SQLite Driver

**MANDATORY: Use modernc.org/sqlite (CGO-free) not mattn/go-sqlite3 (requires CGO)**

**Correct Initialization Pattern**:

```go
import (
    "database/sql"
    "modernc.org/sqlite"
    "gorm.io/driver/sqlite"
)

// CORRECT: Open with modernc driver (CGO-free)
sqlDB, err := sql.Open("sqlite", dsn)
if err != nil {
    return fmt.Errorf("failed to open SQLite: %w", err)
}

// WRONG: Direct GORM SQLite open (may use CGO driver)
// db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
```

**Why This Matters**:
- Project has CGO_ENABLED=0 globally (see `03-03.golang.instructions.md`)
- modernc.org/sqlite is pure Go (no C dependencies)
- mattn/go-sqlite3 requires CGO, breaks static builds

---

## Required SQLite Configuration

### 1. Database Connection Setup

```go
package sqlrepository

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    cryptoutilMagic "cryptoutil/internal/shared/magic"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func OpenSQLite(dsn string, debugMode bool) (*gorm.DB, error) {
    // Step 1: Convert :memory: to shared cache for multi-connection access
    if dsn == ":memory:" {
        dsn = "file::memory:?cache=shared"
    }
    
    // Step 2: Open with modernc driver (CGO-free)
    sqlDB, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open SQLite: %w", err)
    }
    
    // Step 3: Configure SQLite for concurrent operations
    if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL;"); err != nil {
        return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
    }
    
    if _, err := sqlDB.Exec("PRAGMA busy_timeout = 30000;"); err != nil {
        return nil, fmt.Errorf("failed to set busy timeout: %w", err)
    }
    
    // Step 4: Pass to GORM with auto-transactions disabled
    dialector := sqlite.Dialector{Conn: sqlDB}
    gormConfig := &gorm.Config{
        SkipDefaultTransaction: true,  // Explicit transaction control
    }
    
    if debugMode {
        gormConfig.Logger = logger.Default.LogMode(logger.Info)
    }
    
    db, err := gorm.Open(dialector, gormConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize GORM: %w", err)
    }
    
    // Step 5: Configure connection pool for GORM transactions
    sqlDB, err = db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get database instance: %w", err)
    }
    
    // CRITICAL: Allow multiple connections for transaction wrapper pattern
    sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)  // 5
    sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)  // 5
    sqlDB.SetConnMaxLifetime(0)  // In-memory: never close connections
    
    return db, nil
}
```

### 2. PRAGMA Settings Explained

#### WAL Mode (Write-Ahead Logging)

```go
sqlDB.Exec("PRAGMA journal_mode=WAL;")
```

**Purpose**: Enable concurrent readers + 1 writer
**Default**: DELETE mode (exclusive locking, no concurrent access)
**Benefits**:
- Multiple readers can access database while writer is active
- Faster writes (append-only log, not in-place updates)
- Reduces "database is locked" errors in parallel tests

**Trade-offs**:
- Requires 3 files: main DB, WAL file, SHM file
- Not suitable for network filesystems (use DELETE mode)
- Slight overhead for checkpoint operations

#### Busy Timeout

```go
sqlDB.Exec("PRAGMA busy_timeout = 30000;")  // 30 seconds
```

**Purpose**: Retry when database is locked instead of immediately failing
**Default**: 0 (immediate failure on lock contention)
**Benefits**:
- Prevents "database is locked" errors in concurrent operations
- Automatically retries with exponential backoff
- Required for parallel Go tests with `t.Parallel()`

**Magic Constant**: Use `cryptoutilMagic.DBSQLiteBusyTimeout` (30000 milliseconds = 30 seconds)

### 3. Connection Pool Configuration

```go
sqlDB.SetMaxOpenConns(5)   // CRITICAL: Required for GORM transactions
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(0)  // In-memory: never expire connections
```

**Why MaxOpenConns=5**:
- GORM transaction pattern: `db.Begin()` creates new connection
- Base operations need separate connection from transaction wrapper
- MaxOpenConns=1 causes deadlock: transaction holds connection, operations can't proceed
- 5 connections allows: 1 transaction wrapper + 4 concurrent operations

**Connection Lifecycle**:
- **File-based SQLite**: SetConnMaxLifetime(time.Hour) to prevent stale connections
- **In-memory SQLite**: SetConnMaxLifetime(0) to keep connections alive (closing loses data)

---

## Transaction Context Pattern

### Context Injection for Repository Transparency

**Problem**: Repositories need to use transaction if active, otherwise base DB

**Solution**: Store transaction in context, repositories check context first

```go
package sqlrepository

import (
    "context"
    "gorm.io/gorm"
)

type txKey struct{}

// WithTransaction stores transaction in context
func WithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
    return context.WithValue(ctx, txKey{}, tx)
}

// getDB returns transaction from context if exists, otherwise base DB
func getDB(ctx context.Context, baseDB *gorm.DB) *gorm.DB {
    if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
        return tx
    }
    return baseDB
}

// Repository pattern
type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    // Automatically uses transaction if in context, otherwise base DB
    return getDB(ctx, r.db).WithContext(ctx).Create(user).Error
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    var user User
    err := getDB(ctx, r.db).WithContext(ctx).First(&user, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

**Usage in Service Layer**:

```go
func (s *UserService) CreateUserWithProfile(ctx context.Context, user *User, profile *Profile) error {
    // Start transaction
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    // Inject transaction into context
    txCtx := WithTransaction(ctx, tx)
    
    // Repositories automatically use transaction
    if err := s.userRepo.Create(txCtx, user); err != nil {
        tx.Rollback()
        return err
    }
    
    if err := s.profileRepo.Create(txCtx, profile); err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit().Error
}
```

---

## In-Memory Shared Cache

### Problem: Each Connection Gets Separate In-Memory Database

**Default Behavior**:

```go
dsn := ":memory:"
// Connection 1: Creates in-memory DB #1
// Connection 2: Creates in-memory DB #2 (ISOLATED from #1!)
```

**Solution: Use Shared Cache Mode**:

```go
// Convert :memory: to shared cache
if dsn == ":memory:" {
    dsn = "file::memory:?cache=shared"
}
// All connections share SAME in-memory database
```

**Why This Matters**:
- GORM with MaxOpenConns=5 creates up to 5 connections
- Without shared cache, each connection sees different database state
- Tests create data with connection 1, query returns nothing with connection 2
- Shared cache ensures all connections see same in-memory data

**URI Parameters**:
- `file::memory:` - In-memory database with file URI syntax
- `?cache=shared` - Enable shared cache across all connections to this URI

---

## Troubleshooting Guide

### "go-sqlite3 requires cgo"

**Symptom**: Build error: `package github.com/mattn/go-sqlite3 requires cgo`

**Root Cause**: Using mattn/go-sqlite3 driver with CGO_ENABLED=0

**Solution**: Use modernc.org/sqlite driver

```go
// WRONG
import _ "github.com/mattn/go-sqlite3"
db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})

// CORRECT
import "database/sql"
import _ "modernc.org/sqlite"
import "gorm.io/driver/sqlite"

sqlDB, _ := sql.Open("sqlite", dsn)
db, _ := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
```

### Tests Hang with No Error

**Symptom**: Tests start but never complete, no error output

**Root Cause**: MaxOpenConns=1 + GORM transactions = deadlock

**Solution**: Set MaxOpenConns=5

```go
// WRONG: Causes deadlock
sqlDB.SetMaxOpenConns(1)

// CORRECT: Allows transaction wrapper + operations
sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)  // 5
```

**Debugging**:
- Add logging: `gormConfig.Logger = logger.Default.LogMode(logger.Info)`
- Check stack traces: `GOTRACEBACK=all go test -v ./...`
- Look for: "waiting for connection" or similar blocking messages

### "database is locked"

**Symptom**: Error: `database is locked` during concurrent operations

**Root Causes**:
1. Repositories not using `getDB(ctx, r.db)` pattern (bypass transaction context)
2. Missing WAL mode (still using DELETE journal mode)
3. Missing busy_timeout (immediate failure on lock contention)

**Solutions**:

```go
// 1. Fix repository pattern
// WRONG: Always uses base DB
return r.db.WithContext(ctx).Create(user).Error

// CORRECT: Uses transaction from context if exists
return getDB(ctx, r.db).WithContext(ctx).Create(user).Error

// 2. Verify WAL mode enabled
sqlDB.Exec("PRAGMA journal_mode=WAL;")

// 3. Set busy timeout
sqlDB.Exec("PRAGMA busy_timeout = 30000;")
```

### "SQLite doesn't support read-only transactions"

**Symptom**: Error when using `sql.TxOptions{ReadOnly: true}`

**Root Cause**: SQLite does NOT implement read-only transaction isolation level

**Solution**: Use standard transactions or direct queries for read operations

```go
// WRONG: Read-only transaction fails on SQLite
tx := db.Begin(&sql.TxOptions{ReadOnly: true})

// CORRECT Option 1: Standard transaction
tx := db.Begin()

// CORRECT Option 2: Direct query without transaction
result := db.Find(&models)
```

**Cross-Database Compatibility**: See `03-04.database.instructions.md` for PostgreSQL vs SQLite patterns

---

## Key Takeaways

1. **MaxOpenConns=5 for GORM**: Transaction wrapper needs separate connection
2. **modernc.org/sqlite**: CGO-free driver for static builds
3. **WAL mode + busy_timeout**: Enable concurrent operations, prevent "database is locked"
4. **Shared cache for :memory:**: `file::memory:?cache=shared` ensures all connections see same data
5. **Transaction context pattern**: `getDB(ctx, r.db)` for repository transparency
6. **No read-only transactions**: SQLite limitation, use standard transactions or direct queries
