---
name: migration-create
description: "Create numbered golang-migrate SQL migration files for cryptoutil services. Use when adding database schema changes to ensure correct version ranges (template 1001-1999, domain 2001+), paired up/down files, and cross-DB SQL idioms."
argument-hint: "[NNN description of change]"
---

Create numbered golang-migrate SQL migration files for a PS-ID service.

**Full Copilot original**: [.github/skills/migration-create/SKILL.md](.github/skills/migration-create/SKILL.md)

Provide: PS-ID (e.g., `sm-kms`), migration description (e.g., `add_audit_log`).

## Key Rules

- ALWAYS create both `.up.sql` and `.down.sql` files
- Filenames: `NNNN_description.up.sql` / `NNNN_description.down.sql`
- Domain migrations START at 2001 (never overlap with template 1001-1999)
- `.down.sql` must fully reverse `.up.sql` (idempotent rollback)
- Use `IF NOT EXISTS` / `IF EXISTS` for safety
- UUID columns: `TEXT` type (cross-DB: PostgreSQL + SQLite)
- Timestamps: `TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`

## Migration Number Ranges (from registry.yaml)

| PS-ID | Range |
|-------|-------|
| Framework template (shared) | 1001–1999 |
| sm-kms | 2001–2999 |
| sm-im | 3001–3999 |
| jose-ja | 4001–4999 |
| pki-ca | 5001–5999 |
| identity-authz | 6001–6999 |
| identity-idp | 7001–7999 |
| identity-rs | 8001–8999 |
| identity-rp | 9001–9999 |
| identity-spa | 10001–10999 |
| skeleton-template | 11001–11999 |

**NEVER** modify 1001–1999 (framework template migrations).

## File Naming

```
internal/apps/{ps-id}/server/repository/migrations/{NNNN}_{description}.up.sql
internal/apps/{ps-id}/server/repository/migrations/{NNNN}_{description}.down.sql
```

Where `NNNN` is the next sequential number in the PS-ID's range.

## SQL Conventions

- UUID columns: `TEXT` type (cross-DB compatible)
- Timestamps: `CREATED_AT TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP`
- Primary keys: `id TEXT NOT NULL PRIMARY KEY`
- Foreign keys: reference with `ON DELETE CASCADE` or `ON DELETE SET NULL`
- Use `IF NOT EXISTS` / `IF EXISTS` for idempotency

## Template

```sql
-- {NNNN}_{description}.up.sql
CREATE TABLE IF NOT EXISTS {table_name} (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_{table_name}_tenant_id ON {table_name}(tenant_id);
```

```sql
-- {NNNN}_{description}.down.sql
DROP TABLE IF EXISTS {table_name};
```

## Registration

In `migrations.go`:
```go
//go:embed migrations/*.sql
var MigrationsFS embed.FS
```

In server builder:
```go
builder.WithDomainMigrations(MigrationsFS)
```
