--
-- Key Management database schema (SQLite + PostgreSQL compatible)
--

CREATE TABLE IF NOT EXISTS messages_recipient_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    recipient_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    encrypted_jwk TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (recipient_id) REFERENCES users (id),
    FOREIGN KEY (message_id) REFERENCES messages (id),
    UNIQUE (recipient_id, message_id)
);

CREATE INDEX IF NOT EXISTS idx_messages_recipient_jwks_recipient_id ON messages_recipient_jwks (recipient_id);
CREATE INDEX IF NOT EXISTS idx_messages_recipient_jwks_message_id ON messages_recipient_jwks (message_id);
