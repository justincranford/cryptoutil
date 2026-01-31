-- WebAuthn credentials table for storing user credentials (passkeys, security keys)
-- Supports FIDO2/WebAuthn registration and authentication ceremonies.

CREATE TABLE webauthn_credentials (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    credential_id BLOB NOT NULL,
    public_key BLOB NOT NULL,
    attestation_type TEXT NOT NULL,
    transports TEXT,
    sign_count INTEGER NOT NULL DEFAULT 0,
    aaguid BLOB,
    clone_warning INTEGER NOT NULL DEFAULT 0,
    display_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes for efficient queries
CREATE INDEX idx_webauthn_credentials_user_id ON webauthn_credentials(user_id);
CREATE INDEX idx_webauthn_credentials_last_used_at ON webauthn_credentials(last_used_at);
CREATE INDEX idx_webauthn_credentials_deleted_at ON webauthn_credentials(deleted_at);

-- Unique constraint on credential_id per user
CREATE UNIQUE INDEX idx_webauthn_credentials_user_credential ON webauthn_credentials(user_id, credential_id);
