--
-- KMS Tenant ID Migration - Rollback
-- Domain migration 2002: Remove tenant_id from elastic_keys
--

-- Recreate elastic_keys without tenant_id
CREATE TABLE IF NOT EXISTS elastic_keys_old (
    elastic_key_id TEXT PRIMARY KEY NOT NULL,
    elastic_key_name TEXT NOT NULL,
    elastic_key_description TEXT NOT NULL,
    elastic_key_provider TEXT NOT NULL,
    elastic_key_algorithm TEXT NOT NULL,
    elastic_key_versioning_allowed INTEGER NOT NULL,
    elastic_key_import_allowed INTEGER NOT NULL,
    elastic_key_status TEXT NOT NULL,
    CHECK (length(elastic_key_name) >= 1),
    -- Name pattern validation (alphanumeric, hyphens, underscores) enforced at application layer.
    -- GLOB is SQLite-only; removed for cross-database (PostgreSQL + SQLite) compatibility.
    CHECK (length(elastic_key_description) >= 1),
    CHECK (elastic_key_provider = 'Internal'),
    CHECK (elastic_key_versioning_allowed IN (0, 1)),
    CHECK (elastic_key_import_allowed IN (0, 1)),
    CHECK (elastic_key_status IN (
        'creating', 'import_failed', 'pending_import', 'pending_generate',
        'generate_failed', 'active', 'disabled', 'pending_delete_was_import_failed',
        'pending_delete_was_pending_import', 'pending_delete_was_active',
        'pending_delete_was_disabled', 'pending_delete_was_generate_failed',
        'started_delete', 'finished_delete'
    ))
);

-- Copy data without tenant_id
INSERT INTO elastic_keys_old (
    elastic_key_id, elastic_key_name, elastic_key_description,
    elastic_key_provider, elastic_key_algorithm, elastic_key_versioning_allowed,
    elastic_key_import_allowed, elastic_key_status
)
SELECT
    elastic_key_id, elastic_key_name, elastic_key_description,
    elastic_key_provider, elastic_key_algorithm, elastic_key_versioning_allowed,
    elastic_key_import_allowed, elastic_key_status
FROM elastic_keys;

-- Drop new table and rename old
DROP TABLE elastic_keys;
ALTER TABLE elastic_keys_old RENAME TO elastic_keys;

-- Recreate original indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_elastic_keys_name ON elastic_keys(elastic_key_name);
CREATE INDEX IF NOT EXISTS idx_elastic_keys_status ON elastic_keys(elastic_key_status);
