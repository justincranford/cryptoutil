# Database and ORM Patterns - Complete Specifications

**Referenced by**: `.github/instructions/03-04.database.instructions.md`

## Core Database Requirements

**ORM**: GORM (MANDATORY, never raw database/sql)
**Rationale**: Consistent API across PostgreSQL/SQLite, migrations, type-safe queries, error handling, transactions

## Cross-Database Compatibility - CRITICAL

### UUID Handling

**ALWAYS use TEXT for UUIDs** (SQLite has no native UUID type)

**Pattern**: `gorm:"type:text;primaryKey"` works on both PostgreSQL and SQLite

### Nullable UUID Foreign Keys

**Use NullableUUID type** (pointer UUIDs cause "row value misused" errors in SQLite)

**Implementation**: Custom type with sql.Scanner/driver.Valuer, handles NULL properly for TEXT columns

### JSON Array/Object Fields

**ALWAYS use `serializer:json`** (NOT `type:json`, SQLite has no native JSON type)

**Pattern**: `AllowedScopes []string` with `gorm:"serializer:json"` works on both PostgreSQL and SQLite

---

## SQLite Concurrent Operations - CRITICAL

**Required PRAGMA**: WAL mode + busy timeout (30s) for concurrent operations
**Connection Pool**: MaxOpenConns=5 for GORM transactions (see sqlite-gorm.md)
**Read-Only Transactions**: NOT supported - use standard transactions or direct queries

**Magic Constants**: `cryptoutilMagic.DBSQLiteBusyTimeout` (30s), `cryptoutilMagic.SQLiteMaxOpenConnections` (5)

## Database DSN Patterns

**PostgreSQL**: `postgres://user:pass@localhost:5432/dbname?sslmode=disable` (dev) or `sslmode=require` (prod)
**SQLite File**: `file:/var/lib/cryptoutil/data.db?cache=shared&mode=rwc`
**SQLite Memory**: `file::memory:?cache=shared`

## Error Mapping

**Pattern**: toAppErr method maps GORM errors to HTTP errors (ErrRecordNotFound → 404, ErrDuplicatedKey → 409)

## Migrations

**Use golang-migrate with embedded files** (`//go:embed migrations/*.sql`), run on startup before starting servers

**Naming**: `0001_init.up.sql`, `0001_init.down.sql`, `0002_add_users.up.sql`, `0002_add_users.down.sql`

## Connection Pooling

**PostgreSQL**: MaxOpenConns=25, MaxIdleConns=10, ConnMaxLifetime=1h
**SQLite**: MaxOpenConns=5, MaxIdleConns=5, ConnMaxLifetime=0 (in-memory)

## Pagination and Filtering

**Pagination**: Offset/limit pattern with total count
**Filtering**: Dynamic WHERE clauses based on filter struct

## Debug Logging

**Enable**: `db.Logger = logger.Default.LogMode(logger.Info)` (development only)

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
