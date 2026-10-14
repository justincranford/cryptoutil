-- Key Management database schema
CREATE TABLE IF NOT EXISTS pki_authorities (
    id TEXT NOT NULL PRIMARY KEY,
    certificate_der BYTEA NOT NULL,
    private_key_encrypted BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
