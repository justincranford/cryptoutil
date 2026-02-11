-- Create device_authorizations table for RFC 8628 Device Authorization Grant.
CREATE TABLE IF NOT EXISTS device_authorizations (
    id TEXT PRIMARY KEY,
    client_id TEXT NOT NULL,
    device_code TEXT NOT NULL UNIQUE,
    user_code TEXT NOT NULL UNIQUE,
    scope TEXT,
    user_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    last_polled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP
);

-- Create indexes for efficient queries.
CREATE INDEX IF NOT EXISTS idx_device_authorizations_client_id ON device_authorizations(client_id);
CREATE INDEX IF NOT EXISTS idx_device_authorizations_status ON device_authorizations(status);
CREATE INDEX IF NOT EXISTS idx_device_authorizations_expires_at ON device_authorizations(expires_at);
CREATE INDEX IF NOT EXISTS idx_device_authorizations_user_id ON device_authorizations(user_id);
CREATE INDEX IF NOT EXISTS idx_device_authorizations_last_polled_at ON device_authorizations(last_polled_at);
CREATE INDEX IF NOT EXISTS idx_device_authorizations_used_at ON device_authorizations(used_at);
