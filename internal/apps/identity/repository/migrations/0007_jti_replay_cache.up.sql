-- Migration: 0007_jti_replay_cache.up.sql
-- Description: Create JTI replay cache table for client authentication JWT replay attack prevention.
-- Reference: RFC 7523 Section 3 (JWT assertions must have unique jti values)

CREATE TABLE jti_replay_cache (
    jti TEXT PRIMARY KEY,
    client_id TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_jti_replay_cache_expires_at ON jti_replay_cache(expires_at);
