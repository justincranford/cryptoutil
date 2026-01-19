--
-- Rollback Tenant Join Requests table
--

DROP INDEX IF EXISTS idx_tenant_join_requests_processed_by;
DROP INDEX IF EXISTS idx_tenant_join_requests_requested_at;
DROP INDEX IF EXISTS idx_tenant_join_requests_status;
DROP INDEX IF EXISTS idx_tenant_join_requests_tenant_id;
DROP INDEX IF EXISTS idx_tenant_join_requests_client_id;
DROP INDEX IF EXISTS idx_tenant_join_requests_user_id;

DROP TABLE IF EXISTS tenant_join_requests;

