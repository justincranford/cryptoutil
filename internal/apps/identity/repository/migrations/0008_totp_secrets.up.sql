-- Migration: 0008_totp_secrets.up.sql
-- Description: Create TOTP secrets table for Time-based One-Time Password MFA.
-- Reference: RFC 6238 (TOTP: Time-Based One-Time Password Algorithm)

CREATE TABLE totp_secrets (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL,
    algorithm TEXT NOT NULL DEFAULT 'SHA1',
    digits INTEGER NOT NULL DEFAULT 6,
    period INTEGER NOT NULL DEFAULT 30,
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_totp_secrets_user_id ON totp_secrets(user_id);
CREATE INDEX idx_totp_secrets_locked_until ON totp_secrets(locked_until);
CREATE INDEX idx_totp_secrets_last_used_at ON totp_secrets(last_used_at);
