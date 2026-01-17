--
-- JOSE-JA database schema rollback
-- 2003: Audit configuration - drop tenant_audit_config table
--

DROP INDEX IF EXISTS idx_tenant_audit_config_operation;
DROP INDEX IF EXISTS idx_tenant_audit_config_tenant;
DROP TABLE IF EXISTS tenant_audit_config;
