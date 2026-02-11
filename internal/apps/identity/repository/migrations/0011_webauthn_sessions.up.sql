-- WebAuthn sessions table for storing temporary ceremony data
-- Sessions expire after the ceremony timeout period (typically 60 seconds)

CREATE TABLE webauthn_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    session_data BLOB NOT NULL,
    ceremony_type TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

-- Indexes for efficient queries
CREATE INDEX idx_webauthn_sessions_user_id ON webauthn_sessions(user_id);
CREATE INDEX idx_webauthn_sessions_expires_at ON webauthn_sessions(expires_at);
