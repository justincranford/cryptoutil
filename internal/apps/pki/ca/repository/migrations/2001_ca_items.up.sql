-- pki-ca domain migration: create ca_items table.
-- Migration range 2001+: domain-specific (non-template).
CREATE TABLE IF NOT EXISTS ca_items (
    id         TEXT NOT NULL PRIMARY KEY,
    tenant_id  TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ca_items_tenant ON ca_items (tenant_id);
