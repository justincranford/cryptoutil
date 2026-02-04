--
-- KMS Business Tables Migration
-- Domain migration 2001: elastic_keys and material_keys tables
-- Template migration 1002 provides barrier tables
--
-- Cross-database compatible: PostgreSQL and SQLite
--

-- Elastic Keys table - key envelopes that can contain multiple material key versions
CREATE TABLE IF NOT EXISTS elastic_keys (
    elastic_key_id TEXT PRIMARY KEY NOT NULL,
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
    ))
);

-- Unique index for elastic key names
CREATE UNIQUE INDEX IF NOT EXISTS idx_elastic_keys_name ON elastic_keys(elastic_key_name);
CREATE INDEX IF NOT EXISTS idx_elastic_keys_status ON elastic_keys(elastic_key_status);

-- Material Keys table - specific key versions within an elastic key
CREATE TABLE IF NOT EXISTS material_keys (
    elastic_key_id TEXT NOT NULL,
    material_key_id TEXT NOT NULL,
    material_key_clear_public BLOB,
    material_key_encrypted_non_public BLOB NOT NULL,
    material_key_generate_date BIGINT,
    material_key_import_date BIGINT,
    material_key_expiration_date BIGINT,
    material_key_revocation_date BIGINT,
    -- Composite primary key
    PRIMARY KEY (elastic_key_id, material_key_id),
    -- Foreign key to elastic_keys
    FOREIGN KEY (elastic_key_id) REFERENCES elastic_keys(elastic_key_id) ON DELETE CASCADE
);

-- Indexes for material keys
CREATE INDEX IF NOT EXISTS idx_material_keys_elastic_key ON material_keys(elastic_key_id);
CREATE INDEX IF NOT EXISTS idx_material_keys_generate_date ON material_keys(material_key_generate_date);
CREATE INDEX IF NOT EXISTS idx_material_keys_expiration ON material_keys(material_key_expiration_date);
