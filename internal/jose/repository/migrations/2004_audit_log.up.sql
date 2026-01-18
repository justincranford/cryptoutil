-- Copyright (c) 2025 Justin Cranford
--
--

-- Audit log entries: Sampled cryptographic operations for compliance.
CREATE TABLE IF NOT EXISTS tenant_audit_log (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    realm_id TEXT NOT NULL,
    user_id TEXT,
    operation TEXT NOT NULL,  -- encrypt, decrypt, sign, verify, keygen, rotate.
    resource_type TEXT NOT NULL,  -- elastic_jwk, material_jwk.
    resource_id TEXT NOT NULL,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    metadata TEXT,  -- JSON blob with operation-specific details.
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id, realm_id) REFERENCES tenant_realms(tenant_id, realm_id)
);

CREATE INDEX IF NOT EXISTS idx_audit_log_tenant_realm ON tenant_audit_log(tenant_id, realm_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_operation ON tenant_audit_log(operation);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON tenant_audit_log(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_log_resource ON tenant_audit_log(resource_type, resource_id);
