--
-- Cipher IM database schema (SQLite + PostgreSQL compatible)
-- 3-table design: users, messages, messages_recipient_jwks
--

-- Users table for user accounts
-- Password stored as PBKDF2-HMAC-SHA256 hash
-- No ECDH keys (ephemeral per-message encryption)
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY NOT NULL,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(username)
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Messages table with multi-recipient JWE encryption
-- JWE JSON format (NOT Compact Serialization) with N recipient keys
-- Algorithm: enc=A256GCM (content encryption), alg=A256GCMKW (key wrapping per recipient)
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY NOT NULL,
    sender_id TEXT NOT NULL,
    jwe TEXT NOT NULL,  -- JWE JSON format (multi-recipient)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_read_at ON messages(read_at);

-- Messages Recipient JWKs table for per-recipient decryption keys
-- Each recipient gets their own encrypted JWK for decrypting the message
-- JWK encrypted with alg=dir (direct encryption), enc=A256GCM
CREATE TABLE IF NOT EXISTS messages_recipient_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    recipient_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    encrypted_jwk TEXT NOT NULL,  -- Encrypted JWK in JSON format (enc=A256GCM, alg=dir)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (recipient_id) REFERENCES users(id),
    FOREIGN KEY (message_id) REFERENCES messages(id),
    UNIQUE(recipient_id, message_id)
);

CREATE INDEX IF NOT EXISTS idx_messages_recipient_jwks_recipient_id ON messages_recipient_jwks(recipient_id);
CREATE INDEX IF NOT EXISTS idx_messages_recipient_jwks_message_id ON messages_recipient_jwks(message_id);
