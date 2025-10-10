--
-- SQLite database schema
--

-- Barrier content keys table
CREATE TABLE barrier_content_keys (
    uuid TEXT PRIMARY KEY NOT NULL,
    encrypted TEXT NOT NULL,
    kek_uuid TEXT NOT NULL
);

-- Barrier intermediate keys table
CREATE TABLE barrier_intermediate_keys (
    uuid TEXT PRIMARY KEY NOT NULL,
    encrypted TEXT NOT NULL,
    kek_uuid TEXT NOT NULL
);

-- Barrier root keys table
CREATE TABLE barrier_root_keys (
    uuid TEXT PRIMARY KEY NOT NULL,
    encrypted TEXT NOT NULL,
    kek_uuid TEXT NOT NULL
);

-- Elastic keys table
CREATE TABLE elastic_keys (
    elastic_key_id TEXT PRIMARY KEY NOT NULL,
    elastic_key_name TEXT NOT NULL,
    elastic_key_description TEXT NOT NULL,
    elastic_key_provider TEXT NOT NULL,
    elastic_key_algorithm TEXT NOT NULL,
    elastic_key_versioning_allowed INTEGER NOT NULL,
    elastic_key_import_allowed INTEGER NOT NULL,
    elastic_key_status TEXT NOT NULL,
    -- Add SQLite compatible constraints with CHECK
    CHECK (length(elastic_key_description) >= 1),
    CHECK (elastic_key_import_allowed IN (0, 1)),
    CHECK (length(elastic_key_name) >= 1),
    CHECK (elastic_key_provider = 'Internal'),
    CHECK (elastic_key_status IN ('creating', 'import_failed', 'pending_import', 'pending_generate',
                                'generate_failed', 'active', 'disabled', 'pending_delete_was_import_failed',
                                'pending_delete_was_pending_import', 'pending_delete_was_active',
                                'pending_delete_was_disabled', 'pending_delete_was_generate_failed',
                                'started_delete', 'finished_delete')),
    CHECK (elastic_key_versioning_allowed IN (0, 1))
);

-- Create unique index for elastic key names
CREATE UNIQUE INDEX idx_elastic_keys_name ON elastic_keys(elastic_key_name);

-- Material keys table
CREATE TABLE material_keys (
    elastic_key_id TEXT NOT NULL,
    material_key_id TEXT NOT NULL,
    material_key_clear_public BLOB,
    material_key_encrypted_non_public BLOB NOT NULL,
    material_key_generate_date TIMESTAMP,
    material_key_import_date TIMESTAMP,
    material_key_expiration_date TIMESTAMP,
    material_key_revocation_date TIMESTAMP,
    PRIMARY KEY (elastic_key_id, material_key_id)
);
