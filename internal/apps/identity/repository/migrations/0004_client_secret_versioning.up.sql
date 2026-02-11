--
-- Client secret versioning and key rotation event tracking (P5.08)
--

CREATE TABLE IF NOT EXISTS client_secret_versions (
    id TEXT PRIMARY KEY NOT NULL,
    client_id TEXT NOT NULL,
    version INTEGER NOT NULL,
    secret_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    revoked_at TIMESTAMP,
    rotated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    created_by TEXT,
    revoked_by TEXT,

    -- Foreign key constraint
    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_client_secret_versions_client_id ON client_secret_versions(client_id);
CREATE INDEX IF NOT EXISTS idx_client_secret_versions_status ON client_secret_versions(status);
CREATE INDEX IF NOT EXISTS idx_client_secret_versions_expires_at ON client_secret_versions(expires_at);
CREATE INDEX IF NOT EXISTS idx_client_secret_versions_revoked_at ON client_secret_versions(revoked_at);
CREATE INDEX IF NOT EXISTS idx_client_secret_versions_deleted_at ON client_secret_versions(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_client_secret_versions_client_version ON client_secret_versions(client_id, version) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS key_rotation_events (
    id TEXT PRIMARY KEY NOT NULL,
    event_type TEXT NOT NULL,
    key_type TEXT NOT NULL,
    key_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    initiator TEXT NOT NULL,
    old_key_version INTEGER,
    new_key_version INTEGER,
    grace_period TEXT,
    reason TEXT,
    metadata TEXT,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_key_rotation_events_event_type ON key_rotation_events(event_type);
CREATE INDEX IF NOT EXISTS idx_key_rotation_events_key_type ON key_rotation_events(key_type);
CREATE INDEX IF NOT EXISTS idx_key_rotation_events_key_id ON key_rotation_events(key_id);
CREATE INDEX IF NOT EXISTS idx_key_rotation_events_timestamp ON key_rotation_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_key_rotation_events_initiator ON key_rotation_events(initiator);
CREATE INDEX IF NOT EXISTS idx_key_rotation_events_deleted_at ON key_rotation_events(deleted_at);
