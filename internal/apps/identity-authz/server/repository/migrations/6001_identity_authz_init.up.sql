-- Authorization Server database schema
CREATE TABLE IF NOT EXISTS identity_authz_policies (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    policy_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_identity_authz_policies_tenant_id
ON identity_authz_policies (tenant_id);
