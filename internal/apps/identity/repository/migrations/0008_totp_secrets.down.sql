-- Migration: 0008_totp_secrets.down.sql
-- Description: Drop TOTP secrets table.

DROP INDEX IF EXISTS idx_totp_secrets_last_used_at;
DROP INDEX IF EXISTS idx_totp_secrets_locked_until;
DROP INDEX IF EXISTS idx_totp_secrets_user_id;
DROP TABLE IF EXISTS totp_secrets;
