--
-- Rollback SessionManager tables
--

DROP INDEX IF EXISTS idx_session_tokens_jwk_id;
DROP INDEX IF EXISTS idx_session_tokens_revoked_at;
DROP INDEX IF EXISTS idx_session_tokens_expires_at;
DROP INDEX IF EXISTS idx_session_tokens_algorithm_type;
DROP INDEX IF EXISTS idx_session_tokens_user_id;
DROP INDEX IF EXISTS idx_session_tokens_token_id;
DROP TABLE IF EXISTS session_tokens;

DROP INDEX IF EXISTS idx_session_jwks_created_at;
DROP INDEX IF EXISTS idx_session_jwks_status;
DROP INDEX IF EXISTS idx_session_jwks_algorithm_type;
DROP TABLE IF EXISTS session_jwks;