--
-- JOSE-JA database schema (SQLite + PostgreSQL compatible)
-- 2003: Audit configuration - per-tenant audit sampling settings
--

-- Tenant audit configuration table
-- Controls which operations are audited and at what sampling rate
-- Sampling allows audit logging without excessive storage costs
CREATE TABLE IF NOT EXISTS tenant_audit_config (
    tenant_id TEXT NOT NULL,
    operation TEXT NOT NULL,                        -- Operation type: generate, sign, verify, encrypt, decrypt, rotate
    enabled BOOLEAN NOT NULL DEFAULT TRUE,          -- Whether auditing is enabled for this operation
    sampling_rate REAL NOT NULL DEFAULT 0.01,       -- Sampling rate (0.0 to 1.0, default 1%)
    PRIMARY KEY (tenant_id, operation),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_audit_config_tenant ON tenant_audit_config(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_audit_config_operation ON tenant_audit_config(operation);
