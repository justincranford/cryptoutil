--
-- SessionManager tables for template-based session management
-- Supports OPAQUE, JWS, and JWE session types with mixed algorithms
--
-- Table names match GORM model TableName() methods:
--   - browser_session_jwks, service_session_jwks
--   - browser_sessions, service_sessions
--

-- Browser Session JWKs table (Elastic Key Ring pattern)
-- Stores active + historical JWKs for browser session type
CREATE TABLE IF NOT EXISTS browser_session_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    encrypted_jwk TEXT NOT NULL,      -- Encrypted JWK using barrier service (JWE format)
    algorithm TEXT NOT NULL,          -- OPAQUE, JWS_RSA2048, JWS_RSA3072, JWS_RSA4096, JWE_AES256GCM, JWE_AES384HS, JWE_AES512HS
    active INTEGER NOT NULL DEFAULT 1,  -- 1=active (current), 0=inactive (for verification only)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_browser_session_jwks_algorithm ON browser_session_jwks(algorithm);
CREATE INDEX IF NOT EXISTS idx_browser_session_jwks_active ON browser_session_jwks(active);
CREATE INDEX IF NOT EXISTS idx_browser_session_jwks_created_at ON browser_session_jwks(created_at);

-- Service Session JWKs table (Elastic Key Ring pattern)
-- Stores active + historical JWKs for service session type
CREATE TABLE IF NOT EXISTS service_session_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    encrypted_jwk TEXT NOT NULL,      -- Encrypted JWK using barrier service (JWE format)
    algorithm TEXT NOT NULL,          -- OPAQUE, JWS_RSA2048, JWS_RSA3072, JWS_RSA4096, JWE_AES256GCM, JWE_AES384HS, JWE_AES512HS
    active INTEGER NOT NULL DEFAULT 1,  -- 1=active (current), 0=inactive (for verification only)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_service_session_jwks_algorithm ON service_session_jwks(algorithm);
CREATE INDEX IF NOT EXISTS idx_service_session_jwks_active ON service_session_jwks(active);
CREATE INDEX IF NOT EXISTS idx_service_session_jwks_created_at ON service_session_jwks(created_at);

-- Browser Sessions table
-- Active browser sessions with metadata
CREATE TABLE IF NOT EXISTS browser_sessions (
    id TEXT PRIMARY KEY NOT NULL,
    token_hash TEXT,                  -- Hashed token (OPAQUE only), NULL for JWE/JWS
    expiration TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP,          -- Last activity timestamp for idle timeout
    realm TEXT,                       -- Realm identifier for multi-tenancy
    user_id TEXT                      -- User identifier
);

CREATE INDEX IF NOT EXISTS idx_browser_sessions_token_hash ON browser_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_expiration ON browser_sessions(expiration);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_user_id ON browser_sessions(user_id);

-- Service Sessions table
-- Active service sessions with metadata
CREATE TABLE IF NOT EXISTS service_sessions (
    id TEXT PRIMARY KEY NOT NULL,
    token_hash TEXT,                  -- Hashed token (OPAQUE only), NULL for JWE/JWS
    expiration TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP,          -- Last activity timestamp for idle timeout
    realm TEXT,                       -- Realm identifier for multi-tenancy
    client_id TEXT                    -- Client identifier
);

CREATE INDEX IF NOT EXISTS idx_service_sessions_token_hash ON service_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_service_sessions_expiration ON service_sessions(expiration);
CREATE INDEX IF NOT EXISTS idx_service_sessions_client_id ON service_sessions(client_id);
