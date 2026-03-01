# migration-create

Create numbered golang-migrate SQL migration files for cryptoutil services.

## Purpose

Use when adding database schema changes. Ensures correct version numbering,
paired up/down files, and proper SQL idioms.

## Version Ranges

| Type | Range | Examples |
|------|-------|---------|
| Template | 1001–1999 | sessions, barrier, realms, tenants (NEVER modify) |
| Domain | 2001+ | Application-specific tables |

## Key Rules

- ALWAYS create both `.up.sql` and `.down.sql` files
- Filenames: `NNNN_description.up.sql` / `NNNN_description.down.sql`  
- Domain migrations START at 2001 (never overlap with template 1001-1999)
- `.down.sql` must fully reverse `.up.sql` (idempotent rollback)
- Use `IF NOT EXISTS` / `IF EXISTS` for safety
- UUID columns: `TEXT` type (cross-DB: PostgreSQL + SQLite)
- Timestamps: `TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`

## File Structure

```
internal/apps/PRODUCT/SERVICE/repository/migrations/
├── 2001_create_keys.up.sql
├── 2001_create_keys.down.sql
├── 2002_add_key_metadata.up.sql
└── 2002_add_key_metadata.down.sql
```

## Template: up.sql

```sql
-- 2001_create_keys.up.sql
CREATE TABLE IF NOT EXISTS keys (
    id          TEXT        NOT NULL,
    tenant_id   TEXT        NOT NULL,
    algorithm   TEXT        NOT NULL,
    key_data    TEXT        NOT NULL,
    created_at  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
CREATE INDEX IF NOT EXISTS idx_keys_tenant_id ON keys(tenant_id);
```

## Template: down.sql

```sql
-- 2001_create_keys.down.sql
DROP TABLE IF EXISTS keys;
```

## Registration in Go

```go
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// In builder:
builder.WithDomainMigrations(repository.MigrationsFS, "migrations")
```

## References

See [ARCHITECTURE.md Section 7 Data Architecture](../../docs/ARCHITECTURE.md#7-data-architecture) for migration versioning and naming.
See [ARCHITECTURE.md Section 5.2 Service Builder Pattern](../../docs/ARCHITECTURE.md#52-service-builder-pattern) for migration registration.
