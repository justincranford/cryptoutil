--
-- Tenant Join Requests table for multi-tenant registration workflow
--
-- Users/clients request to join existing tenants via this table.
-- Admin users can approve/reject join requests via admin API.
--

CREATE TABLE IF NOT EXISTS tenant_join_requests (
    id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT,                           -- Nullable - mutually exclusive with client_id
    client_id TEXT,                         -- Nullable - mutually exclusive with user_id
    tenant_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,                 -- Timestamp when request was approved/rejected
    processed_by TEXT,                      -- User ID of admin who processed the request
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES unverified_users(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES unverified_clients(id) ON DELETE CASCADE,
    FOREIGN KEY (processed_by) REFERENCES users(id) ON DELETE SET NULL,
    CHECK (
        (user_id IS NOT NULL AND client_id IS NULL) OR
        (user_id IS NULL AND client_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_tenant_join_requests_user_id ON tenant_join_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_join_requests_client_id ON tenant_join_requests(client_id);
CREATE INDEX IF NOT EXISTS idx_tenant_join_requests_tenant_id ON tenant_join_requests(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_join_requests_status ON tenant_join_requests(status);
CREATE INDEX IF NOT EXISTS idx_tenant_join_requests_requested_at ON tenant_join_requests(requested_at);
CREATE INDEX IF NOT EXISTS idx_tenant_join_requests_processed_by ON tenant_join_requests(processed_by);

