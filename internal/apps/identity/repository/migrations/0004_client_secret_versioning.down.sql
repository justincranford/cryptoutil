--
-- Rollback client secret versioning and key rotation event tracking (P5.08)
--

DROP INDEX IF EXISTS idx_key_rotation_events_deleted_at;
DROP INDEX IF EXISTS idx_key_rotation_events_initiator;
DROP INDEX IF EXISTS idx_key_rotation_events_timestamp;
DROP INDEX IF EXISTS idx_key_rotation_events_key_id;
DROP INDEX IF EXISTS idx_key_rotation_events_key_type;
DROP INDEX IF EXISTS idx_key_rotation_events_event_type;
DROP TABLE IF EXISTS key_rotation_events;

DROP INDEX IF EXISTS idx_client_secret_versions_client_version;
DROP INDEX IF EXISTS idx_client_secret_versions_deleted_at;
DROP INDEX IF EXISTS idx_client_secret_versions_revoked_at;
DROP INDEX IF EXISTS idx_client_secret_versions_expires_at;
DROP INDEX IF EXISTS idx_client_secret_versions_status;
DROP INDEX IF EXISTS idx_client_secret_versions_client_id;
DROP TABLE IF EXISTS client_secret_versions;
