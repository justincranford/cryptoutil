-- Key Management database schema rollback.
-- Remove the persisted encrypted private-key envelope format version.
ALTER TABLE pki_authorities
DROP COLUMN private_key_envelope_version;
