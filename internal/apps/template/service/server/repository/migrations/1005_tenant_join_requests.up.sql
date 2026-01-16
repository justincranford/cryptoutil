-- Copyright (c) 2025 Justin Cranford
-- Migration: Add tenant join requests table for multi-tenant user/client registration.

CREATE TABLE IF NOT EXISTS tenant_join_requests (
    id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT,
    client_id TEXT,
    tenant_id TEXT NOT NULL,
    status TEXT NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    processed_by TEXT,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (processed_by) REFERENCES users(id),
    CHECK ((user_id IS NOT NULL AND client_id IS NULL) OR (user_id IS NULL AND client_id IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS idx_tenant_join_requests_tenant ON tenant_join_requests(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_join_requests_status ON tenant_join_requests(status);
