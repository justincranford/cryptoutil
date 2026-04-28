-- Provider database schema
CREATE TABLE IF NOT EXISTS identity_idp_sessions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    subject TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_identity_idp_sessions_tenant_id
ON identity_idp_sessions (tenant_id);
