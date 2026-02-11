-- Migration: 0006_logout_channels.down.sql
-- Description: Remove front-channel and back-channel logout configuration columns from clients table.

-- SQLite doesn't support ALTER TABLE DROP COLUMN before version 3.35.0
-- Use a table rebuild approach for compatibility
CREATE TABLE clients_backup AS SELECT
    id, client_id, client_secret, client_type, j_w_ks, name, description, logo_uri, home_page_uri,
    policy_uri, tos_uri, redirect_uris, post_logout_redirect_uris, allowed_grant_types,
    allowed_response_types, allowed_scopes, token_endpoint_auth_method, require_pkce,
    pkce_challenge_method, access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
    client_profile_id, certificate_subject, certificate_fingerprint, enabled, created_at,
    updated_at, deleted_at
FROM clients;

DROP TABLE clients;

ALTER TABLE clients_backup RENAME TO clients;

CREATE UNIQUE INDEX idx_clients_client_id ON clients(client_id);
CREATE INDEX idx_clients_deleted_at ON clients(deleted_at);
CREATE INDEX idx_clients_certificate_subject ON clients(certificate_subject);
CREATE INDEX idx_clients_certificate_fingerprint ON clients(certificate_fingerprint);
CREATE INDEX idx_clients_client_profile_id ON clients(client_profile_id);
