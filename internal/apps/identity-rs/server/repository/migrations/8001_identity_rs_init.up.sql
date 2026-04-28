-- Resource Server database schema
CREATE TABLE IF NOT EXISTS identity_rs_resources (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_identity_rs_resources_tenant_id
ON identity_rs_resources (tenant_id);
