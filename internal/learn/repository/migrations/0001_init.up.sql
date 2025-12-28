--
-- Learn IM database schema (SQLite + PostgreSQL compatible)
--

-- Users table for user accounts
-- NOTE: PublicKey/PrivateKey stored as BLOB for backward compatibility
-- Phase 4 will migrate to JWK-based storage in users_jwks table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY NOT NULL,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    public_key BLOB NOT NULL,   -- ECDH public key (temporary, migrate to users_jwks in Phase 4)
    private_key BLOB NOT NULL,  -- ECDH private key (temporary, migrate to users_jwks in Phase 4)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(username)
);

CREATE INDEX idx_users_username ON users(username);

-- Users JWKs table for per-user encryption keys
-- Algorithm: ECDH-ES (key agreement) + A256GCM (content encryption)
-- Phase 4 will migrate User.PublicKey/PrivateKey to this table
CREATE TABLE IF NOT EXISTS users_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL,
    jwk_json TEXT NOT NULL,  -- JWK in JSON format
    algorithm TEXT NOT NULL DEFAULT 'ECDH-ES',
    encryption TEXT NOT NULL DEFAULT 'A256GCM',
    key_id TEXT NOT NULL,  -- kid claim from JWK
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, key_id)
);

CREATE INDEX idx_users_jwks_user_id ON users_jwks(user_id);
CREATE INDEX idx_users_jwks_key_id ON users_jwks(key_id);
CREATE INDEX idx_users_jwks_is_active ON users_jwks(is_active);

-- Users Messages JWKs table for per-user/message encryption keys
-- Algorithm: dir (direct encryption) + A256GCM (content encryption)
CREATE TABLE IF NOT EXISTS users_messages_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    jwk_json TEXT NOT NULL,  -- JWK in JSON format
    algorithm TEXT NOT NULL DEFAULT 'dir',
    encryption TEXT NOT NULL DEFAULT 'A256GCM',
    key_id TEXT NOT NULL,  -- kid claim from JWK
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, message_id, key_id)
);

CREATE INDEX idx_users_messages_jwks_user_id ON users_messages_jwks(user_id);
CREATE INDEX idx_users_messages_jwks_message_id ON users_messages_jwks(message_id);
CREATE INDEX idx_users_messages_jwks_key_id ON users_messages_jwks(key_id);

-- Messages JWKs table for per-message encryption keys
-- Algorithm: dir (direct encryption) + A256GCM (content encryption)
CREATE TABLE IF NOT EXISTS messages_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    message_id TEXT NOT NULL,
    jwk_json TEXT NOT NULL,  -- JWK in JSON format
    algorithm TEXT NOT NULL DEFAULT 'dir',
    encryption TEXT NOT NULL DEFAULT 'A256GCM',
    key_id TEXT NOT NULL,  -- kid claim from JWK
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(message_id, key_id)
);

CREATE INDEX idx_messages_jwks_message_id ON messages_jwks(message_id);
CREATE INDEX idx_messages_jwks_key_id ON messages_jwks(key_id);
CREATE INDEX idx_messages_jwks_is_active ON messages_jwks(is_active);

-- Messages table with encrypted content
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY NOT NULL,
    sender_id TEXT NOT NULL,
    recipient_id TEXT NOT NULL,
    jwe_compact TEXT NOT NULL,  -- JWE Compact Serialization format (eyJ...)
    key_id TEXT NOT NULL,  -- References messages_jwks.key_id
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users_jwks(user_id),
    FOREIGN KEY (recipient_id) REFERENCES users_jwks(user_id),
    FOREIGN KEY (key_id) REFERENCES messages_jwks(key_id)
);

CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_recipient_id ON messages(recipient_id);
CREATE INDEX idx_messages_key_id ON messages(key_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_read_at ON messages(read_at);
