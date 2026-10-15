-- Key Management database schema
-- Preserve the configured authority identity independently of issuer subject/key details.
ALTER TABLE pki_certificates
ADD COLUMN issuer_authority_id TEXT NOT NULL DEFAULT '';
