-- Drop device_authorizations table and indexes.
DROP INDEX IF EXISTS idx_device_authorizations_used_at;
DROP INDEX IF EXISTS idx_device_authorizations_last_polled_at;
DROP INDEX IF EXISTS idx_device_authorizations_user_id;
DROP INDEX IF EXISTS idx_device_authorizations_expires_at;
DROP INDEX IF EXISTS idx_device_authorizations_status;
DROP INDEX IF EXISTS idx_device_authorizations_client_id;
DROP TABLE IF EXISTS device_authorizations;
