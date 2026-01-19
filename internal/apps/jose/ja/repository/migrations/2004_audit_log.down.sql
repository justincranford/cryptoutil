--
-- JOSE-JA database schema rollback
-- 2004: Audit log - drop audit_log table
--

DROP INDEX IF EXISTS idx_audit_log_success;
DROP INDEX IF EXISTS idx_audit_log_request_id;
DROP INDEX IF EXISTS idx_audit_log_created_at;
DROP INDEX IF EXISTS idx_audit_log_operation;
DROP INDEX IF EXISTS idx_audit_log_elastic_jwk;
DROP INDEX IF EXISTS idx_audit_log_session;
