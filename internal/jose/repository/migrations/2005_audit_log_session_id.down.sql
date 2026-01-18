-- Copyright (c) 2025 Justin Cranford
--
--

-- Remove session_id column from audit_log.
-- NOTE: SQLite doesn't support DROP COLUMN, use ALTER TABLE for other databases.
-- For SQLite, this migration cannot be rolled back without recreating the table.
-- In production, consider using a full table recreation if rollback is needed.

-- For PostgreSQL (uncomment if using PostgreSQL):
-- ALTER TABLE tenant_audit_log DROP COLUMN session_id;

-- For SQLite: Cannot easily drop columns. Table would need to be recreated.
-- This is acceptable as rollbacks of audit schema changes are rare in production.
