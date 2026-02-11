--
-- Authorization requests and consent decisions (OAuth 2.1 / OIDC)
--

-- Authorization requests table
CREATE TABLE IF NOT EXISTS authorization_requests (
    id TEXT PRIMARY KEY NOT NULL,

    -- Client information
    client_id TEXT NOT NULL,
    redirect_uri TEXT NOT NULL,

    -- Request parameters
    response_type TEXT NOT NULL,
    scope TEXT,
    state TEXT,
    nonce TEXT,

    -- PKCE parameters (OAuth 2.1 required)
    code_challenge TEXT NOT NULL,
    code_challenge_method TEXT NOT NULL DEFAULT 'S256',

    -- User information (populated after authentication)
    user_id TEXT,

    -- Authorization code (generated after consent)
    code TEXT,

    -- Request metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,

    -- Consent status
    consent_granted INTEGER NOT NULL DEFAULT 0,

    -- Single-use enforcement
    used INTEGER NOT NULL DEFAULT 0,
    used_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_authorization_requests_client_id ON authorization_requests(client_id);
CREATE INDEX IF NOT EXISTS idx_authorization_requests_user_id ON authorization_requests(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_authorization_requests_code ON authorization_requests(code) WHERE code IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_authorization_requests_expires_at ON authorization_requests(expires_at);
CREATE INDEX IF NOT EXISTS idx_authorization_requests_used ON authorization_requests(used);

-- Consent decisions table
CREATE TABLE IF NOT EXISTS consent_decisions (
    id TEXT PRIMARY KEY NOT NULL,

    -- User and client information
    user_id TEXT NOT NULL,
    client_id TEXT NOT NULL,

    -- Granted scopes
    scope TEXT NOT NULL,

    -- Consent metadata
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,

    -- Revocation tracking
    revoked_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_consent_decisions_user_id ON consent_decisions(user_id);
CREATE INDEX IF NOT EXISTS idx_consent_decisions_client_id ON consent_decisions(client_id);
CREATE INDEX IF NOT EXISTS idx_consent_decisions_expires_at ON consent_decisions(expires_at);
CREATE INDEX IF NOT EXISTS idx_consent_decisions_revoked_at ON consent_decisions(revoked_at);
CREATE INDEX IF NOT EXISTS idx_consent_decisions_user_client_scope ON consent_decisions(user_id, client_id, scope);
