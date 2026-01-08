--
-- Rollback SessionManager tables
-- Drops tables in reverse order of creation to respect any foreign key constraints
--

DROP INDEX IF EXISTS idx_service_sessions_client_id;
DROP INDEX IF EXISTS idx_service_sessions_expiration;
DROP INDEX IF EXISTS idx_service_sessions_token_hash;
DROP TABLE IF EXISTS service_sessions;

DROP INDEX IF EXISTS idx_browser_sessions_user_id;
DROP INDEX IF EXISTS idx_browser_sessions_expiration;
DROP INDEX IF EXISTS idx_browser_sessions_token_hash;
DROP TABLE IF EXISTS browser_sessions;

DROP INDEX IF EXISTS idx_service_session_jwks_created_at;
DROP INDEX IF EXISTS idx_service_session_jwks_active;
DROP INDEX IF EXISTS idx_service_session_jwks_algorithm;
DROP TABLE IF EXISTS service_session_jwks;

DROP INDEX IF EXISTS idx_browser_session_jwks_created_at;
DROP INDEX IF EXISTS idx_browser_session_jwks_active;
DROP INDEX IF EXISTS idx_browser_session_jwks_algorithm;
DROP TABLE IF EXISTS browser_session_jwks;
