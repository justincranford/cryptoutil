-- Down migration: Drop WebAuthn credentials table and indexes

DROP INDEX IF EXISTS idx_webauthn_credentials_user_credential;
DROP INDEX IF EXISTS idx_webauthn_credentials_deleted_at;
DROP INDEX IF EXISTS idx_webauthn_credentials_last_used_at;
DROP INDEX IF EXISTS idx_webauthn_credentials_user_id;
DROP TABLE IF EXISTS webauthn_credentials;
