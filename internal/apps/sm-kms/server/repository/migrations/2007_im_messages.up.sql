--
-- Key Management database schema (SQLite + PostgreSQL compatible)
-- messages table only (recipient JWK mapping moved to 2008)
-- Users table is provided by template service (1004_add_multi_tenancy)
--

-- Messages table with multi-recipient JWE encryption
-- JWE JSON format (NOT Compact Serialization) with N recipient keys
-- Algorithm: enc=A256GCM (content encryption), alg=A256GCMKW (key wrapping per recipient)
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY NOT NULL,
    sender_id TEXT NOT NULL,
    jwe TEXT NOT NULL,  -- JWE JSON format (multi-recipient)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users (id)
);

CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages (sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at);
CREATE INDEX IF NOT EXISTS idx_messages_read_at ON messages (read_at);
