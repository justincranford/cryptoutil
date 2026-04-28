-- Relying Party database schema
CREATE TABLE IF NOT EXISTS identity_rp_sessions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    subject TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_identity_rp_sessions_tenant_id
ON identity_rp_sessions (tenant_id);
