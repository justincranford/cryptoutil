-- Key Management database schema
CREATE TABLE IF NOT EXISTS pki_certificates (
    id TEXT NOT NULL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    request_id TEXT NOT NULL,
    serial_number TEXT NOT NULL,
    subject_dn TEXT NOT NULL,
    issuer_dn TEXT NOT NULL,
    not_before TIMESTAMP NOT NULL,
    not_after TIMESTAMP NOT NULL,
    status TEXT NOT NULL,
    profile_id TEXT NOT NULL,
    requester_id TEXT NOT NULL,
    certificate_pem TEXT NOT NULL,
    certificate_der BYTEA,
    public_key_hash TEXT NOT NULL,
    subject_key_id TEXT NOT NULL,
    authority_key_id TEXT NOT NULL,
    revocation_time TIMESTAMP,
    revocation_reason INTEGER,
    metadata TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pki_certificates_tenant_serial
ON pki_certificates (tenant_id, serial_number);
CREATE INDEX IF NOT EXISTS idx_pki_certificates_tenant_status
ON pki_certificates (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_pki_certificates_tenant_request
ON pki_certificates (tenant_id, request_id);
