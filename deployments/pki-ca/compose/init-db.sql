-- CA Database Initialization Script
-- Creates tables for certificate storage and management

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Certificate status enum
CREATE TYPE cert_status AS ENUM ('valid', 'revoked', 'expired', 'suspended');

-- Certificate table
CREATE TABLE IF NOT EXISTS certificates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    serial_number TEXT NOT NULL UNIQUE,
    subject_dn TEXT NOT NULL,
    issuer_dn TEXT NOT NULL,
    not_before TIMESTAMP WITH TIME ZONE NOT NULL,
    not_after TIMESTAMP WITH TIME ZONE NOT NULL,
    public_key_algorithm TEXT NOT NULL,
    signature_algorithm TEXT NOT NULL,
    certificate_pem TEXT NOT NULL,
    status cert_status NOT NULL DEFAULT 'valid',
    profile TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE,
    revocation_reason INTEGER
);

-- Indexes for certificate lookups
CREATE INDEX IF NOT EXISTS idx_certificates_serial ON certificates(serial_number);
CREATE INDEX IF NOT EXISTS idx_certificates_subject ON certificates(subject_dn);
CREATE INDEX IF NOT EXISTS idx_certificates_issuer ON certificates(issuer_dn);
CREATE INDEX IF NOT EXISTS idx_certificates_status ON certificates(status);
CREATE INDEX IF NOT EXISTS idx_certificates_not_after ON certificates(not_after);

-- CRL table
CREATE TABLE IF NOT EXISTS crls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    issuer_dn TEXT NOT NULL,
    crl_number BIGINT NOT NULL,
    this_update TIMESTAMP WITH TIME ZONE NOT NULL,
    next_update TIMESTAMP WITH TIME ZONE NOT NULL,
    crl_pem TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for CRL lookups
CREATE INDEX IF NOT EXISTS idx_crls_issuer ON crls(issuer_dn);
CREATE INDEX IF NOT EXISTS idx_crls_number ON crls(crl_number);

-- Revocation entries table
CREATE TABLE IF NOT EXISTS revocation_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    certificate_id UUID NOT NULL REFERENCES certificates(id),
    serial_number TEXT NOT NULL,
    revocation_date TIMESTAMP WITH TIME ZONE NOT NULL,
    reason INTEGER NOT NULL DEFAULT 0,
    invalidity_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for revocation lookups
CREATE INDEX IF NOT EXISTS idx_revocation_serial ON revocation_entries(serial_number);

-- Audit log table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    event_type TEXT NOT NULL,
    actor TEXT NOT NULL,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    outcome TEXT NOT NULL,
    details JSONB,
    client_ip INET,
    session_id TEXT,
    correlation_id TEXT
);

-- Index for audit log queries
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_event_type ON audit_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_logs(actor);

-- Key storage table (encrypted keys)
CREATE TABLE IF NOT EXISTS key_store (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_id TEXT NOT NULL UNIQUE,
    key_type TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    encrypted_key BYTEA NOT NULL,
    key_wrap_algorithm TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB
);

-- Index for key lookups
CREATE INDEX IF NOT EXISTS idx_key_store_key_id ON key_store(key_id);

-- Certificate requests table
CREATE TABLE IF NOT EXISTS certificate_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id TEXT NOT NULL UNIQUE,
    csr_pem TEXT NOT NULL,
    subject_dn TEXT NOT NULL,
    requested_profile TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    requester TEXT NOT NULL,
    approver TEXT,
    approved_at TIMESTAMP WITH TIME ZONE,
    rejected_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    certificate_id UUID REFERENCES certificates(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for request lookups
CREATE INDEX IF NOT EXISTS idx_cert_requests_status ON certificate_requests(status);
CREATE INDEX IF NOT EXISTS idx_cert_requests_requester ON certificate_requests(requester);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER certificates_updated_at
    BEFORE UPDATE ON certificates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER certificate_requests_updated_at
    BEFORE UPDATE ON certificate_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ca;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ca;
