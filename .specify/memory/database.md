# Database and ORM Patterns - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/03-04.database.instructions.md`

## Core Database Requirements

### ORM Framework

**MANDATORY: Use GORM ORM, never raw database/sql**

**Rationale**:
- Consistent API across PostgreSQL and SQLite
- Automatic migrations support
- Type-safe query building
- Error handling consistency
- Transaction management

**Key Features**:
- Transactions with context support
- Error mapping (GORM errors → application HTTP errors)
- Embedded SQL migrations (golang-migrate)
- Connection pooling (configurable by backend)
- Pagination (offset/limit patterns)
- Filters and sorting
- Debug logging for troubleshooting

---

## Cross-Database Compatibility - CRITICAL

### UUID Type Handling

**SQLite does not support native UUID type - ALWAYS use TEXT for UUIDs**

**Pattern for Cross-DB Compatibility**:

```go
import googleUuid "github.com/google/uuid"

// ✅ CORRECT: Works on PostgreSQL AND SQLite
type User struct {
    ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`
}

// ❌ WRONG: Breaks SQLite (no native UUID type)
type User struct {
    ID googleUuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
}
```

**Why**:
- SQLite lacks native UUID type (only INTEGER, TEXT, BLOB, REAL, NUMERIC)
- PostgreSQL supports both `uuid` and `text` types for UUIDs
- Using `text` ensures compatibility with both databases
- Performance difference negligible (UUIDs indexed as text in SQLite)

**SQL Migration Consistency**:

```sql
-- PostgreSQL migration
CREATE TABLE users (
    id TEXT PRIMARY KEY  -- TEXT, not UUID type
);

-- SQLite migration (identical)
CREATE TABLE users (
    id TEXT PRIMARY KEY
);
```

### Nullable UUID Foreign Keys

**Use NullableUUID type for optional UUID foreign keys**

**Problem with Pointer UUIDs**:

```go
// ❌ WRONG: Pointer UUIDs cause "row value misused" errors in SQLite
type Client struct {
    ClientProfileID *googleUuid.UUID `gorm:"type:text;index"`
}
```

**Correct Pattern**:

```go
// Domain model (internal/identity/domain/nullable_uuid.go)
type NullableUUID struct {
    UUID  googleUuid.UUID
    Valid bool
}

// Implements sql.Scanner and driver.Valuer for proper NULL handling
func (nu *NullableUUID) Scan(value interface{}) error {
    if value == nil {
        nu.UUID, nu.Valid = googleUuid.Nil, false
        return nil
    }
    
    // Handle string (SQLite) or byte slice (PostgreSQL)
    var uuidStr string
    switch v := value.(type) {
    case string:
        uuidStr = v
    case []byte:
        uuidStr = string(v)
    default:
        return fmt.Errorf("unsupported type for NullableUUID: %T", value)
    }
    
    parsed, err := googleUuid.Parse(uuidStr)
    if err != nil {
        return err
    }
    
    nu.UUID, nu.Valid = parsed, true
    return nil
}

func (nu NullableUUID) Value() (driver.Value, error) {
    if !nu.Valid {
        return nil, nil
    }
    return nu.UUID.String(), nil
}

// ✅ CORRECT Usage in domain models
type Client struct {
    ID              googleUuid.UUID `gorm:"type:text;primaryKey"`
    ClientProfileID NullableUUID    `gorm:"type:text;index"`
}
```

**Why NullableUUID Pattern Required**:
- Pointer UUIDs (`*googleUuid.UUID`) don't serialize correctly to TEXT columns in SQLite
- GORM tries to use `row.Scan(&uuid)` which expects binary UUID type
- SQLite stores as TEXT, causing "SQL logic error: row value misused"
- NullableUUID implements proper sql.Scanner for TEXT serialization

### JSON Array/Object Fields

**ALWAYS use `serializer:json` instead of `type:json` for cross-DB compatibility**

**Pattern**:

```go
// ✅ CORRECT: Works on PostgreSQL AND SQLite
type Client struct {
    AllowedScopes []string `gorm:"serializer:json" json:"allowed_scopes"`
    MFAChain      []string `gorm:"serializer:json" json:"mfa_chain"`
    Metadata      map[string]string `gorm:"serializer:json" json:"metadata"`
}

// ❌ WRONG: Breaks SQLite (no native JSON type)
type Client struct {
    AllowedScopes []string `gorm:"type:json" json:"allowed_scopes"`
}
```

**Why**:
- SQLite lacks native JSON type (stores as TEXT)
- PostgreSQL has native JSON/JSONB types
- `serializer:json` tells GORM to handle JSON encoding/decoding for TEXT columns
- `type:json` tells GORM to use native JSON type (fails on SQLite with "row value misused")

**SQL Migration Consistency**:

```sql
-- PostgreSQL migration
CREATE TABLE clients (
    id TEXT PRIMARY KEY,
    allowed_scopes TEXT DEFAULT '[]'  -- JSON array as TEXT
);

-- SQLite migration (identical)
CREATE TABLE clients (
    id TEXT PRIMARY KEY,
    allowed_scopes TEXT DEFAULT '[]'  -- JSON array as TEXT
);
```

**GORM Domain Model MUST Match Migration**:

```go
type Client struct {
    ID            googleUuid.UUID `gorm:"type:text;primaryKey"`
    AllowedScopes []string        `gorm:"serializer:json"`  // Matches TEXT column
}
```

---

## SQLite Concurrent Write Operations - CRITICAL

### Required PRAGMA Settings

**ALWAYS configure WAL mode and busy timeout for concurrent operations**

```go
// Enable WAL mode for better concurrency (multiple readers + 1 writer)
if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL;"); err != nil {
    return fmt.Errorf("failed to enable WAL mode: %w", err)
}

// Set busy timeout for handling concurrent write operations (30 seconds)
if _, err := sqlDB.Exec("PRAGMA busy_timeout = 30000;"); err != nil {
    return fmt.Errorf("failed to set busy timeout: %w", err)
}
```

**Why**:
- **WAL mode** (Write-Ahead Logging): Allows multiple concurrent readers + 1 writer
- **busy_timeout**: Makes SQLite retry when database is locked instead of immediately failing
- **Without these settings**: Parallel Go tests using `t.Parallel()` will fail with "database is locked"

**See**: `.specify/memory/sqlite-gorm.md` for complete SQLite configuration patterns

### Connection Pool Configuration

**SQLite vs GORM Transaction Requirements**:

```go
// For SQLite with GORM transactions, use MaxOpenConns=5
sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)  // 5
sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)  // 5

// Rationale:
// - GORM transaction wrapper needs separate connection from base operations
// - MaxOpenConns=1 causes deadlock (transaction holds connection, operations can't proceed)
// - 5 connections allows: 1 transaction wrapper + 4 concurrent operations
```

**See**: `03-05.sqlite-gorm.instructions.md` for detailed explanation

### Magic Constants

```go
// From internal/shared/magic/magic_database.go
const (
    DBSQLiteBusyTimeout      = 30000  // 30 seconds
    SQLiteMaxOpenConnections = 5      // GORM transaction support
)
```

---

## SQLite Read-Only Transactions - CRITICAL

**SQLite does NOT support read-only transactions - NEVER use them**

```go
// ❌ WRONG: Fails on SQLite with "cannot start a transaction" or "SQLite doesn't support read-only transactions"
tx := db.Begin(&sql.TxOptions{ReadOnly: true})

// ✅ CORRECT Option 1: Standard read-write transaction
tx := db.Begin()

// ✅ CORRECT Option 2: Direct query without transaction (for simple reads)
result := db.Find(&models)
```

**Why**:
- SQLite does not implement `SET TRANSACTION READ ONLY` isolation level
- PostgreSQL supports `SET TRANSACTION READ ONLY` but SQLite ignores/errors on this
- Tests using read-only transactions will fail on SQLite

**Cross-DB Pattern for Read-Heavy Operations**:

```go
func (r *Repository) GetMany(ctx context.Context) ([]Model, error) {
    var results []Model
    // Don't use Begin() for simple reads - direct query is sufficient
    if err := r.db.WithContext(ctx).Find(&results).Error; err != nil {
        return nil, fmt.Errorf("failed to query: %w", err)
    }
    return results, nil
}
```

---

## Database DSN Patterns

### PostgreSQL DSN

```go
// Use localhost for PostgreSQL driver (handles IPv4/IPv6 resolution)
dsn := "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
```

**Connection String Components**:
- Protocol: `postgres://` or `postgresql://`
- Credentials: `user:pass@` (from Docker secrets)
- Host: `localhost` (driver resolves to 127.0.0.1 or ::1)
- Port: `5432` (default PostgreSQL port)
- Database: `dbname`
- SSL Mode: `sslmode=disable` (dev/test), `sslmode=require` (production)

### SQLite DSN

```go
// File-based SQLite
dsn := "file:/var/lib/cryptoutil/data.db?cache=shared&mode=rwc"

// In-memory SQLite with shared cache
dsn := "file::memory:?cache=shared"
```

**Connection String Components**:
- Protocol: `file:` (SQLite file URI)
- Path: `/var/lib/cryptoutil/data.db` or `:memory:` (in-memory)
- Cache: `cache=shared` (required for multiple connections to share same DB)
- Mode: `mode=rwc` (read-write-create)

---

## Error Mapping

### GORM Errors to Application HTTP Errors

**Pattern: toAppErr method in repository layer**

```go
func toAppErr(err error) error {
    if err == nil {
        return nil
    }
    
    switch {
    case errors.Is(err, gorm.ErrRecordNotFound):
        return ErrNotFound  // HTTP 404
    case errors.Is(err, gorm.ErrDuplicatedKey):
        return ErrConflict  // HTTP 409
    case errors.Is(err, gorm.ErrInvalidTransaction):
        return ErrInternalServer  // HTTP 500
    default:
        return fmt.Errorf("database error: %w", err)
    }
}

// Repository usage
func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
    if err != nil {
        return nil, toAppErr(err)  // Maps to ErrNotFound if not found
    }
    return &user, nil
}
```

---

## Migrations

### Embedded SQL Migrations

**Pattern: Use golang-migrate with embedded files**

```go
import (
    "embed"
    
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    "github.com/golang-migrate/migrate/v4/database/sqlite3"
    "github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB, dialect string) error {
    source, err := iofs.New(migrationsFS, "migrations")
    if err != nil {
        return fmt.Errorf("failed to create migration source: %w", err)
    }
    
    var driver migrate.Driver
    switch dialect {
    case "postgres":
        driver, err = postgres.WithInstance(db, &postgres.Config{})
    case "sqlite":
        driver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
    default:
        return fmt.Errorf("unsupported database dialect: %s", dialect)
    }
    if err != nil {
        return fmt.Errorf("failed to create migration driver: %w", err)
    }
    
    m, err := migrate.NewWithInstance("iofs", source, dialect, driver)
    if err != nil {
        return fmt.Errorf("failed to create migration instance: %w", err)
    }
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("failed to apply migrations: %w", err)
    }
    
    return nil
}
```

**Migration File Naming**:
- `0001_init.up.sql` - Initial schema
- `0001_init.down.sql` - Rollback initial schema
- `0002_add_users.up.sql` - Add users table
- `0002_add_users.down.sql` - Remove users table

**Always Apply Migrations on Startup**:

```go
func (app *Application) Start(ctx context.Context) error {
    // 1. Run migrations
    if err := RunMigrations(app.db, app.config.DatabaseDialect); err != nil {
        return fmt.Errorf("failed to run migrations: %w", err)
    }
    
    // 2. Start servers
    // ...
}
```

---

## Connection Pooling

### PostgreSQL Configuration

```go
sqlDB.SetMaxOpenConns(25)   // Maximum concurrent connections
sqlDB.SetMaxIdleConns(10)   // Idle connections pool
sqlDB.SetConnMaxLifetime(1 * time.Hour)  // Connection lifetime
```

### SQLite Configuration

```go
// GORM with transactions: Use 5 connections
sqlDB.SetMaxOpenConns(5)   // Allow transaction wrapper + operations
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(0)  // In-memory: never close connections
```

**See**: `.specify/memory/sqlite-gorm.md` for rationale

---

## Pagination and Filtering

### Pagination Pattern

```go
func (r *Repository) List(ctx context.Context, page, pageSize int) ([]Model, int64, error) {
    var models []Model
    var total int64
    
    // Count total records
    if err := r.db.WithContext(ctx).Model(&Model{}).Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to count records: %w", err)
    }
    
    // Fetch page
    offset := (page - 1) * pageSize
    if err := r.db.WithContext(ctx).
        Offset(offset).
        Limit(pageSize).
        Find(&models).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to fetch records: %w", err)
    }
    
    return models, total, nil
}
```

### Filtering Pattern

```go
func (r *Repository) FindByFilter(ctx context.Context, filter Filter) ([]Model, error) {
    var models []Model
    
    query := r.db.WithContext(ctx).Model(&Model{})
    
    if filter.Name != "" {
        query = query.Where("name LIKE ?", "%"+filter.Name+"%")
    }
    
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    
    if err := query.Find(&models).Error; err != nil {
        return nil, fmt.Errorf("failed to filter records: %w", err)
    }
    
    return models, nil
}
```

---

## Debug Logging

### Enable Debug Mode for Troubleshooting

```go
import "gorm.io/gorm/logger"

// Enable debug logging in development
if app.config.DebugMode {
    db.Logger = logger.Default.LogMode(logger.Info)
}

// Log database schema information
if app.config.DebugMode {
    var tables []string
    db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables)
    log.Printf("Database tables: %v", tables)
}
```

---

## Key Takeaways

1. **GORM Always**: Never use raw database/sql (use GORM for consistency)
2. **UUID as TEXT**: ALWAYS `type:text` for UUIDs (cross-DB compatibility)
3. **NullableUUID**: Use custom type for optional UUID foreign keys (not pointer UUIDs)
4. **serializer:json**: Use for JSON arrays/objects (not `type:json`)
5. **WAL Mode + Busy Timeout**: Required for SQLite concurrent operations
6. **No Read-Only Transactions**: SQLite limitation (use standard transactions or direct queries)
7. **Migrations on Startup**: ALWAYS apply migrations before starting servers
8. **Error Mapping**: toAppErr method maps GORM errors to application HTTP errors
