--
-- KMS Tenant ID Migration
-- Domain migration 2002: Add tenant_id to elastic_keys and material_keys
-- Enables multi-tenant isolation for KMS data
--

-- Step 1: Add tenant_id column to elastic_keys
-- Uses TEXT type for cross-database compatibility (PostgreSQL and SQLite)
ALTER TABLE elastic_keys ADD COLUMN tenant_id TEXT;

-- Step 2: Set default tenant_id for existing records (backward compatibility)
-- Uses well-known default tenant UUID from template migration 1004
UPDATE elastic_keys SET tenant_id = '00000000-0000-0000-0000-000000000001' WHERE tenant_id IS NULL;

-- Step 3: Make tenant_id NOT NULL after backfill
-- SQLite doesn't support ALTER COLUMN directly, so we recreate the table
-- This is safe for fresh databases and preserves existing data

-- Create new table with proper schema
CREATE TABLE IF NOT EXISTS elastic_keys_new (
    elastic_key_id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    elastic_key_name TEXT NOT NULL,
    elastic_key_description TEXT NOT NULL,
    elastic_key_provider TEXT NOT NULL,
    elastic_key_algorithm TEXT NOT NULL,
    elastic_key_versioning_allowed INTEGER NOT NULL,
    elastic_key_import_allowed INTEGER NOT NULL,
    elastic_key_status TEXT NOT NULL,
    -- Constraints
    CHECK (length(elastic_key_name) >= 1),
    CHECK (elastic_key_name GLOB '[A-Za-z0-9_-]*'),
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
    )),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Copy data from old table
INSERT INTO elastic_keys_new (
    elastic_key_id, tenant_id, elastic_key_name, elastic_key_description,
    elastic_key_provider, elastic_key_algorithm, elastic_key_versioning_allowed,
    elastic_key_import_allowed, elastic_key_status
)
SELECT
    elastic_key_id, tenant_id, elastic_key_name, elastic_key_description,
    elastic_key_provider, elastic_key_algorithm, elastic_key_versioning_allowed,
    elastic_key_import_allowed, elastic_key_status
FROM elastic_keys;

-- Drop old table and rename new
DROP TABLE elastic_keys;
ALTER TABLE elastic_keys_new RENAME TO elastic_keys;

-- Recreate indexes with tenant_id awareness
-- Unique name per tenant (not globally unique)
CREATE UNIQUE INDEX IF NOT EXISTS idx_elastic_keys_tenant_name ON elastic_keys(tenant_id, elastic_key_name);
CREATE INDEX IF NOT EXISTS idx_elastic_keys_tenant_id ON elastic_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_elastic_keys_status ON elastic_keys(elastic_key_status);
