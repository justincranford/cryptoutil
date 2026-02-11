--
-- Migration: Remove post_logout_redirect_uris from clients table
-- Purpose: Rollback OIDC RP-Initiated Logout support
--

-- SQLite does not support DROP COLUMN directly.
-- To rollback, we would need to recreate the table.
-- For simplicity, this down migration is a no-op comment.
-- In production, use a proper table recreation if rollback is needed.

-- No-op: SQLite ALTER TABLE DROP COLUMN not supported before 3.35.0.
-- Schema will retain the column but it will be ignored by application.
