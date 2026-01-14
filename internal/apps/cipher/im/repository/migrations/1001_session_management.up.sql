--
-- Session Management Tables (SQLite + PostgreSQL compatible)
-- Supports JWE, JWS, and OPAQUE (hashed UUIDv7) session token types
--

-- Browser Session JWKs table
-- Stores encrypted JWKs for JWE/JWS session token signing/encryption
-- Empty if using OPAQUE algorithm (UUIDv7 tokens)
CREATE TABLE IF NOT EXISTS browser_session_jwks (
    id TEXT PRIMARY KEY NOT NULL,                    -- UUIDv7
    encrypted_jwk TEXT NOT NULL,                     -- JWK encrypted with barrier layer
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    algorithm TEXT NOT NULL,                         -- JWE or JWS algorithm identifier
    active BOOLEAN NOT NULL DEFAULT TRUE             -- Active key for signing, historical keys for verification
);

CREATE INDEX IF NOT EXISTS idx_browser_session_jwks_created_at ON browser_session_jwks(created_at);
CREATE INDEX IF NOT EXISTS idx_browser_session_jwks_active ON browser_session_jwks(active);

-- Service Session JWKs table
-- Stores encrypted JWKs for JWE/JWS session token signing/encryption
-- Empty if using OPAQUE algorithm (UUIDv7 tokens)
CREATE TABLE IF NOT EXISTS service_session_jwks (
    id TEXT PRIMARY KEY NOT NULL,                    -- UUIDv7
    encrypted_jwk TEXT NOT NULL,                     -- JWK encrypted with barrier layer
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    algorithm TEXT NOT NULL,                         -- JWE or JWS algorithm identifier
    active BOOLEAN NOT NULL DEFAULT TRUE             -- Active key for signing, historical keys for verification
);

CREATE INDEX IF NOT EXISTS idx_service_session_jwks_created_at ON service_session_jwks(created_at);
CREATE INDEX IF NOT EXISTS idx_service_session_jwks_active ON service_session_jwks(active);

-- Browser Sessions table
-- Stores session identifiers and metadata for browser users
-- For OPAQUE algorithm: stores hashed UUIDv7 token
-- For JWE/JWS: stores jti claim from JWT
CREATE TABLE IF NOT EXISTS browser_sessions (
    id TEXT PRIMARY KEY NOT NULL,                    -- UUIDv7 session identifier
    token_hash TEXT,                                 -- Hashed token (OPAQUE only), NULL for JWE/JWS
    user_id TEXT,                                    -- User identifier (optional, depends on service implementation)
    expiration TIMESTAMP NOT NULL,                   -- Session expiration timestamp
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP,                         -- Last activity timestamp for idle timeout
    realm TEXT                                       -- Realm identifier for multi-tenancy
);

CREATE INDEX IF NOT EXISTS idx_browser_sessions_expiration ON browser_sessions(expiration);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_token_hash ON browser_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_user_id ON browser_sessions(user_id);

-- Service Sessions table
-- Stores session identifiers and metadata for service-to-service clients
-- For OPAQUE algorithm: stores hashed UUIDv7 token
-- For JWE/JWS: stores jti claim from JWT
CREATE TABLE IF NOT EXISTS service_sessions (
    id TEXT PRIMARY KEY NOT NULL,                    -- UUIDv7 session identifier
    token_hash TEXT,                                 -- Hashed token (OPAQUE only), NULL for JWE/JWS
    client_id TEXT,                                  -- Client identifier (optional, depends on service implementation)
    expiration TIMESTAMP NOT NULL,                   -- Session expiration timestamp
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP,                         -- Last activity timestamp for idle timeout
    realm TEXT                                       -- Realm identifier for multi-tenancy
);

CREATE INDEX IF NOT EXISTS idx_service_sessions_expiration ON service_sessions(expiration);
CREATE INDEX IF NOT EXISTS idx_service_sessions_token_hash ON service_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_service_sessions_client_id ON service_sessions(client_id);
