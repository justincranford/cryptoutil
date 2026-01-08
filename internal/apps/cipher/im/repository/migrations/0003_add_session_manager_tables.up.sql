--
-- SessionManager tables for template-based session management
-- Supports OPAQUE, JWS, and JWE session types with mixed algorithms
--

-- Session JWKs table (Elastic Key Ring pattern)
-- Stores active + historical JWKs for each session type and algorithm
-- Active key encrypts/signs new sessions; historical keys decrypt/verify existing sessions
CREATE TABLE IF NOT EXISTS session_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    algorithm TEXT NOT NULL,          -- OPAQUE, JWS_RSA2048, JWS_RSA3072, JWS_RSA4096, JWE_AES256GCM, JWE_AES384HS, JWE_AES512HS
    session_type TEXT NOT NULL,       -- BROWSER, SERVICE
    jwk_data TEXT NOT NULL,           -- Encrypted JWK using barrier service (JWE format)
    status TEXT NOT NULL DEFAULT 'ACTIVE',  -- ACTIVE (current), HISTORICAL (for verification only), REVOKED
    created_at INTEGER NOT NULL,      -- Unix epoch milliseconds
    rotated_at INTEGER              -- Unix epoch milliseconds (when rotated from ACTIVE to HISTORICAL)
);

CREATE INDEX IF NOT EXISTS idx_session_jwks_algorithm_type ON session_jwks(algorithm, session_type);
CREATE INDEX IF NOT EXISTS idx_session_jwks_status ON session_jwks(status);
CREATE INDEX IF NOT EXISTS idx_session_jwks_created_at ON session_jwks(created_at);

-- Session tokens table
-- Active sessions with metadata and revocation support
-- Token value NOT stored (privacy); only metadata for management/cleanup
CREATE TABLE IF NOT EXISTS session_tokens (
    id TEXT PRIMARY KEY NOT NULL,
    token_id TEXT NOT NULL,           -- UUID extracted from token payload for identification
    user_id TEXT NOT NULL,            -- Associated user identifier
    algorithm TEXT NOT NULL,          -- OPAQUE, JWS_RSA2048, etc.
    session_type TEXT NOT NULL,       -- BROWSER, SERVICE
    jwk_id TEXT NOT NULL,             -- JWK used to create this session
    issued_at INTEGER NOT NULL,       -- Unix epoch milliseconds
    expires_at INTEGER NOT NULL,      -- Unix epoch milliseconds
    revoked_at INTEGER,              -- Unix epoch milliseconds (NULL if not revoked)
    FOREIGN KEY (jwk_id) REFERENCES session_jwks(id)
);

CREATE INDEX IF NOT EXISTS idx_session_tokens_token_id ON session_tokens(token_id);
CREATE INDEX IF NOT EXISTS idx_session_tokens_user_id ON session_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_session_tokens_algorithm_type ON session_tokens(algorithm, session_type);
CREATE INDEX IF NOT EXISTS idx_session_tokens_expires_at ON session_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_session_tokens_revoked_at ON session_tokens(revoked_at);
CREATE INDEX IF NOT EXISTS idx_session_tokens_jwk_id ON session_tokens(jwk_id);