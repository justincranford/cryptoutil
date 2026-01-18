-- Copyright (c) 2025 Justin Cranford
--
--

-- Audit configuration: Per-tenant, per-operation audit settings.
CREATE TABLE IF NOT EXISTS tenant_audit_config (
    tenant_id TEXT NOT NULL,
    operation TEXT NOT NULL,  -- encrypt, decrypt, sign, verify, keygen, rotate.
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sampling_rate REAL NOT NULL DEFAULT 0.01,  -- 1% sampling by default.
    PRIMARY KEY (tenant_id, operation),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_audit_config_tenant ON tenant_audit_config(tenant_id);
