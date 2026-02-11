-- Copyright (c) 2025 Justin Cranford
--
--

-- Create client_secret_versions table for tracking secret lifecycle.
CREATE TABLE IF NOT EXISTS client_secret_versions (
    id TEXT PRIMARY KEY,
    client_id TEXT NOT NULL,
    version INTEGER NOT NULL,
    secret_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    expires_at TIMESTAMP,
    revoked_at TIMESTAMP,
    revoked_by TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes for client_secret_versions.
CREATE INDEX IF NOT EXISTS idx_client_secret_versions_client_id ON client_secret_versions(client_id);
CREATE INDEX IF NOT EXISTS idx_client_secret_versions_expires_at ON client_secret_versions(expires_at);
CREATE INDEX IF NOT EXISTS idx_client_secret_versions_deleted_at ON client_secret_versions(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_client_secret_versions_client_version ON client_secret_versions(client_id, version) WHERE deleted_at IS NULL;

-- Create key_rotation_events table for audit trail.
CREATE TABLE IF NOT EXISTS key_rotation_events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    key_type TEXT NOT NULL,
    key_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    initiator TEXT NOT NULL,
    old_key_version INTEGER,
    new_key_version INTEGER,
    grace_period TEXT,
    reason TEXT,
    metadata TEXT,
    success BOOLEAN NOT NULL DEFAULT TRUE,
    error_message TEXT,
    deleted_at TIMESTAMP
);

-- Indexes for key_rotation_events.
CREATE INDEX IF NOT EXISTS idx_key_rotation_events_event_type ON key_rotation_events(event_type);
CREATE INDEX IF NOT EXISTS idx_key_rotation_events_key_type ON key_rotation_events(key_type);
CREATE INDEX IF NOT EXISTS idx_key_rotation_events_key_id ON key_rotation_events(key_id);
CREATE INDEX IF NOT EXISTS idx_key_rotation_events_timestamp ON key_rotation_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_key_rotation_events_initiator ON key_rotation_events(initiator);
CREATE INDEX IF NOT EXISTS idx_key_rotation_events_deleted_at ON key_rotation_events(deleted_at);
