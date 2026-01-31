-- Migration: 0009_backup_codes.up.sql
-- Description: Create backup codes table for TOTP account recovery.
-- Reference: OWASP Authentication Cheat Sheet - Backup Codes for MFA

CREATE TABLE backup_codes (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES totp_secrets(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_backup_codes_user_id ON backup_codes(user_id);
CREATE INDEX idx_backup_codes_used ON backup_codes(used);
CREATE INDEX idx_backup_codes_user_id_used ON backup_codes(user_id, used);
CREATE INDEX idx_backup_codes_created_at ON backup_codes(created_at);
