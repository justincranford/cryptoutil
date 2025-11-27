--
-- Client secret rotation history tracking
--

CREATE TABLE IF NOT EXISTS client_secret_history (
    id TEXT PRIMARY KEY NOT NULL,
    client_id TEXT NOT NULL,
    secret_hash TEXT NOT NULL,
    rotated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rotated_by TEXT,
    reason TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key constraint
    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_client_secret_history_client_id ON client_secret_history(client_id);
CREATE INDEX IF NOT EXISTS idx_client_secret_history_rotated_at ON client_secret_history(rotated_at);
