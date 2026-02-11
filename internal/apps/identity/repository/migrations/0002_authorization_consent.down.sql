--
-- Drop authorization requests and consent decisions tables
--

DROP INDEX IF EXISTS idx_consent_decisions_user_client_scope;
DROP INDEX IF EXISTS idx_consent_decisions_revoked_at;
DROP INDEX IF EXISTS idx_consent_decisions_expires_at;
DROP INDEX IF EXISTS idx_consent_decisions_client_id;
DROP INDEX IF EXISTS idx_consent_decisions_user_id;
DROP TABLE IF EXISTS consent_decisions;

DROP INDEX IF EXISTS idx_authorization_requests_used;
DROP INDEX IF EXISTS idx_authorization_requests_expires_at;
DROP INDEX IF EXISTS idx_authorization_requests_code;
DROP INDEX IF EXISTS idx_authorization_requests_user_id;
DROP INDEX IF EXISTS idx_authorization_requests_client_id;
DROP TABLE IF EXISTS authorization_requests;
