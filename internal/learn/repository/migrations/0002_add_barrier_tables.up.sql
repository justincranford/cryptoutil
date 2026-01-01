--
-- Barrier encryption tables for multi-layer key hierarchy
-- Pattern: unseal → root → intermediate → content keys
-- Extracted from KMS service into reusable service-template pattern
--

-- Barrier Root Keys table
-- Root keys are unsealed by HSM, KMS, Shamir Key Shares, etc.
-- Encrypted column contains JWE (JOSE Encrypted JSON document)
CREATE TABLE IF NOT EXISTS barrier_root_keys (
    uuid TEXT PRIMARY KEY NOT NULL,
    encrypted TEXT NOT NULL,  -- JWE encrypted root key
    kek_uuid TEXT NOT NULL,   -- Key Encryption Key UUID (unseal key)
    created_at INTEGER NOT NULL,  -- Unix epoch milliseconds
    updated_at INTEGER NOT NULL   -- Unix epoch milliseconds
);

CREATE INDEX IF NOT EXISTS idx_barrier_root_keys_kek_uuid ON barrier_root_keys(kek_uuid);

-- Barrier Intermediate Keys table
-- Intermediate keys are wrapped by root keys
-- Rotation is encouraged and can be frequent
CREATE TABLE IF NOT EXISTS barrier_intermediate_keys (
    uuid TEXT PRIMARY KEY NOT NULL,
    encrypted TEXT NOT NULL,    -- JWE encrypted intermediate key
    kek_uuid TEXT NOT NULL,     -- Root key UUID that encrypted this key
    created_at INTEGER NOT NULL,  -- Unix epoch milliseconds
    updated_at INTEGER NOT NULL,  -- Unix epoch milliseconds
    FOREIGN KEY (kek_uuid) REFERENCES barrier_root_keys(uuid)
);

CREATE INDEX IF NOT EXISTS idx_barrier_intermediate_keys_kek_uuid ON barrier_intermediate_keys(kek_uuid);

-- Barrier Content Keys table
-- Content keys are wrapped by intermediate keys
-- Rotation is encouraged and can be very frequent
CREATE TABLE IF NOT EXISTS barrier_content_keys (
    uuid TEXT PRIMARY KEY NOT NULL,
    encrypted TEXT NOT NULL,    -- JWE encrypted content key
    kek_uuid TEXT NOT NULL,     -- Intermediate key UUID that encrypted this key
    created_at INTEGER NOT NULL,  -- Unix epoch milliseconds
    updated_at INTEGER NOT NULL,  -- Unix epoch milliseconds
    FOREIGN KEY (kek_uuid) REFERENCES barrier_intermediate_keys(uuid)
);

CREATE INDEX IF NOT EXISTS idx_barrier_content_keys_kek_uuid ON barrier_content_keys(kek_uuid);
