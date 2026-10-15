-- Key Management database schema rollback
DROP INDEX IF EXISTS idx_pki_certificates_tenant_request_unique;
CREATE INDEX IF NOT EXISTS idx_pki_certificates_tenant_request
ON pki_certificates (tenant_id, request_id);
