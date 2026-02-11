--
-- Migration: Add post_logout_redirect_uris to clients table
-- Purpose: Support OIDC RP-Initiated Logout (RFC)
--

-- Add post_logout_redirect_uris column to clients table (JSON array for string slice)
ALTER TABLE clients ADD COLUMN post_logout_redirect_uris TEXT DEFAULT '[]';
