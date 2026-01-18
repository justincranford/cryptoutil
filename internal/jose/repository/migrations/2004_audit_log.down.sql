-- Copyright (c) 2025 Justin Cranford
--
--

DROP INDEX IF EXISTS idx_audit_log_resource;
DROP INDEX IF EXISTS idx_audit_log_created_at;
DROP INDEX IF EXISTS idx_audit_log_operation;
DROP INDEX IF EXISTS idx_audit_log_tenant_realm;
DROP TABLE IF EXISTS tenant_audit_log;
