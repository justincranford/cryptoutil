-- Key Management database schema (SQLite + PostgreSQL compatible).
-- Persist the encrypted private-key envelope format for migration and recovery.
ALTER TABLE pki_authorities
ADD COLUMN private_key_envelope_version INTEGER NOT NULL DEFAULT 1;
