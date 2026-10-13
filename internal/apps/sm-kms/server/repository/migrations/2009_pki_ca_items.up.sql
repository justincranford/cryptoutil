-- Key Management database schema: PKI certificate-authority items migrated into sm-kms.
CREATE TABLE IF NOT EXISTS ca_items (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ca_items_tenant ON ca_items (tenant_id);
