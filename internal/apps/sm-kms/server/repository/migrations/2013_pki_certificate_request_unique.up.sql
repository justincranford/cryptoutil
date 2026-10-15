-- Key Management database schema (SQLite + PostgreSQL compatible)
-- Exclude historical all-zero UUID rows; application write boundaries reject
-- the sentinel so new certificate records always have a unique request ID.
CREATE UNIQUE INDEX idx_pki_certificates_tenant_request_unique
ON pki_certificates (tenant_id, request_id)
WHERE request_id <> '00000000-0000-0000-0000-000000000000';
DROP INDEX IF EXISTS idx_pki_certificates_tenant_request;
