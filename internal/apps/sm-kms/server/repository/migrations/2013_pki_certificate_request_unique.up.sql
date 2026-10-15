-- Key Management database schema (SQLite + PostgreSQL compatible)
-- Exclude the all-zero UUID sentinel so legacy/non-enrollment certificates may
-- continue to omit request IDs while real request IDs remain tenant-unique.
CREATE UNIQUE INDEX idx_pki_certificates_tenant_request_unique
ON pki_certificates (tenant_id, request_id)
WHERE request_id <> '00000000-0000-0000-0000-000000000000';
DROP INDEX IF EXISTS idx_pki_certificates_tenant_request;
