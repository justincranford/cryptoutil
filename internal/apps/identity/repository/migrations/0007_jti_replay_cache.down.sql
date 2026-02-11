-- Migration: 0007_jti_replay_cache.down.sql
-- Description: Drop JTI replay cache table.

DROP INDEX IF EXISTS idx_jti_replay_cache_expires_at;
DROP TABLE jti_replay_cache;
