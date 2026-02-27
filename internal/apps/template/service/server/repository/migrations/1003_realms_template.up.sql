--
-- Realms table for authentication realm configuration
-- Supports non-federated authn methods with config & DB storage
-- Config > DB priority pattern (config file overrides database)
--
-- CRITICAL: This is a template table structure. Service-specific
-- realm tables should use service_name_realms naming convention
-- (e.g., sm_im_realms, identity_realms, etc.)
--

-- Template Realms table (services should create their own with service_name_realms)
-- This migration documents the standard schema - services copy and customize
CREATE TABLE IF NOT EXISTS template_realms (
    id TEXT PRIMARY KEY,
    realm_id TEXT NOT NULL,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    config TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    source TEXT NOT NULL DEFAULT 'db',
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient lookups.
CREATE UNIQUE INDEX IF NOT EXISTS idx_template_realms_realm_id ON template_realms(realm_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_template_realms_name ON template_realms(name);
CREATE INDEX IF NOT EXISTS idx_template_realms_type ON template_realms(type);
CREATE INDEX IF NOT EXISTS idx_template_realms_active ON template_realms(active);
CREATE INDEX IF NOT EXISTS idx_template_realms_priority ON template_realms(priority);
