-- Key Management database schema rollback
ALTER TABLE pki_certificates
DROP COLUMN issuer_authority_id;
