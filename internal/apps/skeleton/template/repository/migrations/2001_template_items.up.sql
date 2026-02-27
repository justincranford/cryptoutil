-- skeleton-template domain migration: create template_items table.
-- Migration range 2001+: domain-specific (non-template).
CREATE TABLE IF NOT EXISTS template_items (
    id         TEXT NOT NULL PRIMARY KEY,
    tenant_id  TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_template_items_tenant ON template_items (tenant_id);
