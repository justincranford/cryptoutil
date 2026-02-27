--
-- Barrier encryption tables for multi-layer key hierarchy
-- Pattern: unseal → root → intermediate → content keys
-- Extracted from sm-im into reusable service-template pattern
--
-- CRITICAL: These tables are required by barrier service (internal/apps/template/service/server/barrier/)
--

-- Barrier Root Keys table
-- Root keys are unsealed by HSM, KMS, Shamir Key Shares, etc.
-- Encrypted column contains JWE (JOSE Encrypted JSON document)
-- NOTE: Uses BIGINT for timestamps because PostgreSQL INTEGER is 32-bit (int4),
--       which cannot store Unix epoch milliseconds (requires 41+ bits).
--       SQLite treats BIGINT the same as INTEGER (64-bit).
CREATE TABLE IF NOT EXISTS barrier_root_keys (
    uuid TEXT PRIMARY KEY NOT NULL,
    encrypted TEXT NOT NULL,  -- JWE encrypted root key
    kek_uuid TEXT NOT NULL,   -- Key Encryption Key UUID (unseal key)
    created_at BIGINT NOT NULL,  -- Unix epoch milliseconds (BIGINT for PostgreSQL compat)
    updated_at BIGINT NOT NULL,  -- Unix epoch milliseconds (BIGINT for PostgreSQL compat)
    rotated_at BIGINT            -- Unix epoch milliseconds (NULL = still active)
);

CREATE INDEX IF NOT EXISTS idx_barrier_root_keys_kek_uuid ON barrier_root_keys(kek_uuid);
CREATE INDEX IF NOT EXISTS idx_barrier_root_keys_rotated_at ON barrier_root_keys(rotated_at);

-- Barrier Intermediate Keys table
-- Intermediate keys are wrapped by root keys
-- Rotation is encouraged and can be frequent
-- NOTE: Uses BIGINT for timestamps (see barrier_root_keys comment)
CREATE TABLE IF NOT EXISTS barrier_intermediate_keys (
    uuid TEXT PRIMARY KEY NOT NULL,
    encrypted TEXT NOT NULL,    -- JWE encrypted intermediate key
    kek_uuid TEXT NOT NULL,     -- Root key UUID that encrypted this key
    created_at BIGINT NOT NULL,   -- Unix epoch milliseconds (BIGINT for PostgreSQL compat)
    updated_at BIGINT NOT NULL,   -- Unix epoch milliseconds (BIGINT for PostgreSQL compat)
    rotated_at BIGINT,            -- Unix epoch milliseconds (NULL = still active)
    FOREIGN KEY (kek_uuid) REFERENCES barrier_root_keys(uuid)
);

CREATE INDEX IF NOT EXISTS idx_barrier_intermediate_keys_kek_uuid ON barrier_intermediate_keys(kek_uuid);
CREATE INDEX IF NOT EXISTS idx_barrier_intermediate_keys_rotated_at ON barrier_intermediate_keys(rotated_at);

-- Barrier Content Keys table
-- Content keys are wrapped by intermediate keys
-- Rotation is encouraged and can be very frequent
-- NOTE: Uses BIGINT for timestamps (see barrier_root_keys comment)
CREATE TABLE IF NOT EXISTS barrier_content_keys (
    uuid TEXT PRIMARY KEY NOT NULL,
    encrypted TEXT NOT NULL,    -- JWE encrypted content key
    kek_uuid TEXT NOT NULL,     -- Intermediate key UUID that encrypted this key
    created_at BIGINT NOT NULL,   -- Unix epoch milliseconds (BIGINT for PostgreSQL compat)
    updated_at BIGINT NOT NULL,   -- Unix epoch milliseconds (BIGINT for PostgreSQL compat)
    rotated_at BIGINT,            -- Unix epoch milliseconds (NULL = still active)
    FOREIGN KEY (kek_uuid) REFERENCES barrier_intermediate_keys(uuid)
);

CREATE INDEX IF NOT EXISTS idx_barrier_content_keys_kek_uuid ON barrier_content_keys(kek_uuid);
CREATE INDEX IF NOT EXISTS idx_barrier_content_keys_rotated_at ON barrier_content_keys(rotated_at);
