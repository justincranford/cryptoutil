-- Drop all tables in reverse dependency order

DROP INDEX IF EXISTS idx_auth_flows_deleted_at;
DROP INDEX IF EXISTS idx_auth_flows_client_profile_id;
DROP INDEX IF EXISTS idx_auth_flows_name;
DROP TABLE IF EXISTS auth_flows;

DROP INDEX IF EXISTS idx_mfa_factors_deleted_at;
DROP INDEX IF EXISTS idx_mfa_factors_auth_profile_id;
DROP INDEX IF EXISTS idx_mfa_factors_name;
DROP TABLE IF EXISTS mfa_factors;

DROP INDEX IF EXISTS idx_auth_profiles_deleted_at;
DROP INDEX IF EXISTS idx_auth_profiles_name;
DROP TABLE IF EXISTS auth_profiles;

DROP INDEX IF EXISTS idx_sessions_deleted_at;
DROP INDEX IF EXISTS idx_sessions_active;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_issued_at;
DROP INDEX IF EXISTS idx_sessions_client_id;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_session_id;
DROP TABLE IF EXISTS sessions;

DROP INDEX IF EXISTS idx_tokens_deleted_at;
DROP INDEX IF EXISTS idx_tokens_revoked;
DROP INDEX IF EXISTS idx_tokens_expires_at;
DROP INDEX IF EXISTS idx_tokens_issued_at;
DROP INDEX IF EXISTS idx_tokens_refresh_token_id;
DROP INDEX IF EXISTS idx_tokens_user_id;
DROP INDEX IF EXISTS idx_tokens_client_id;
DROP INDEX IF EXISTS idx_tokens_token_type;
DROP INDEX IF EXISTS idx_tokens_token_value;
DROP TABLE IF EXISTS tokens;

DROP INDEX IF EXISTS idx_clients_deleted_at;
DROP INDEX IF EXISTS idx_clients_client_profile_id;
DROP INDEX IF EXISTS idx_clients_client_id;
DROP TABLE IF EXISTS clients;

DROP INDEX IF EXISTS idx_client_profiles_deleted_at;
DROP INDEX IF EXISTS idx_client_profiles_name;
DROP TABLE IF EXISTS client_profiles;

DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_preferred_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_sub;
DROP TABLE IF EXISTS users;
