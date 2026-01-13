-- Migration: Drop cipher_im_realms table

DROP INDEX IF EXISTS idx_realms_priority;
DROP INDEX IF EXISTS idx_realms_active;
DROP INDEX IF EXISTS idx_realms_type;
DROP INDEX IF EXISTS idx_realms_name;
DROP INDEX IF EXISTS idx_realms_realm_id;

DROP TABLE IF EXISTS cipher_im_realms;
