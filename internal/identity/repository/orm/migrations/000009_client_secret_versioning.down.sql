-- Copyright (c) 2025 Justin Cranford
--
--

-- Drop key_rotation_events table and indexes.
DROP INDEX IF EXISTS idx_key_rotation_events_deleted_at;
DROP INDEX IF EXISTS idx_key_rotation_events_initiator;
DROP INDEX IF EXISTS idx_key_rotation_events_timestamp;
DROP INDEX IF EXISTS idx_key_rotation_events_key_id;
DROP INDEX IF EXISTS idx_key_rotation_events_key_type;
DROP INDEX IF EXISTS idx_key_rotation_events_event_type;
DROP TABLE IF EXISTS key_rotation_events;

-- Drop client_secret_versions table and indexes.
DROP INDEX IF EXISTS idx_client_secret_versions_client_version;
DROP INDEX IF EXISTS idx_client_secret_versions_deleted_at;
DROP INDEX IF EXISTS idx_client_secret_versions_expires_at;
DROP INDEX IF EXISTS idx_client_secret_versions_client_id;
DROP TABLE IF EXISTS client_secret_versions;
