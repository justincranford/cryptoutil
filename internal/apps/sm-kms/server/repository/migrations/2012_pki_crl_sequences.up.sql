-- Key Management database schema (SQLite + PostgreSQL compatible)
-- The signed 64-bit sequence is far above practical PKI CRL issuance volume and is
-- incremented atomically by the database for each issuer key identifier.
CREATE TABLE IF NOT EXISTS pki_crl_sequences (
    issuer_key_id TEXT NOT NULL PRIMARY KEY,
    number BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
