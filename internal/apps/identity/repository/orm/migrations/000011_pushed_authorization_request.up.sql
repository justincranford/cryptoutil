-- Pushed Authorization Requests table (RFC 9126)
CREATE TABLE IF NOT EXISTS pushed_authorization_requests (
    id TEXT PRIMARY KEY,
    request_uri TEXT NOT NULL UNIQUE,
    client_id TEXT NOT NULL,

    -- Authorization parameters.
    response_type TEXT NOT NULL,
    redirect_uri TEXT NOT NULL,
    scope TEXT,
    state TEXT,
    code_challenge TEXT NOT NULL,
    code_challenge_method TEXT NOT NULL,
    nonce TEXT,

    -- Additional parameters (JSON).
    additional_params TEXT,

    -- Lifecycle tracking.
    used BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used_at TIMESTAMP
);

-- Indexes for performance.
CREATE INDEX IF NOT EXISTS idx_par_client_id ON pushed_authorization_requests(client_id);
CREATE INDEX IF NOT EXISTS idx_par_expires_at ON pushed_authorization_requests(expires_at);
CREATE INDEX IF NOT EXISTS idx_par_used ON pushed_authorization_requests(used);
