-- Down migration: Drop WebAuthn sessions table and indexes

DROP INDEX IF EXISTS idx_webauthn_sessions_expires_at;
DROP INDEX IF EXISTS idx_webauthn_sessions_user_id;
DROP TABLE IF EXISTS webauthn_sessions;
