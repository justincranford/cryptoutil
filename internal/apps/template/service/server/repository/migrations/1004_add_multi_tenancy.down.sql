--
-- Rollback Multi-Tenancy tables for template service
--

-- Drop indexes first
DROP INDEX IF EXISTS idx_service_sessions_created_at;
DROP INDEX IF EXISTS idx_service_sessions_client_id;
DROP INDEX IF EXISTS idx_service_sessions_last_activity;
DROP INDEX IF EXISTS idx_service_sessions_expiration;
DROP INDEX IF EXISTS idx_service_sessions_token_hash;
DROP INDEX IF EXISTS idx_service_sessions_realm_id;
DROP INDEX IF EXISTS idx_service_sessions_tenant_id;

DROP INDEX IF EXISTS idx_browser_sessions_created_at;
DROP INDEX IF EXISTS idx_browser_sessions_user_id;
DROP INDEX IF EXISTS idx_browser_sessions_last_activity;
DROP INDEX IF EXISTS idx_browser_sessions_expiration;
DROP INDEX IF EXISTS idx_browser_sessions_token_hash;
DROP INDEX IF EXISTS idx_browser_sessions_realm_id;
DROP INDEX IF EXISTS idx_browser_sessions_tenant_id;

DROP INDEX IF EXISTS idx_tenant_realms_tenant_realm;
DROP INDEX IF EXISTS idx_tenant_realms_created_at;
DROP INDEX IF EXISTS idx_tenant_realms_source;
DROP INDEX IF EXISTS idx_tenant_realms_active;
DROP INDEX IF EXISTS idx_tenant_realms_type;
DROP INDEX IF EXISTS idx_tenant_realms_realm_id;
DROP INDEX IF EXISTS idx_tenant_realms_tenant_id;

DROP INDEX IF EXISTS idx_client_roles_created_at;
DROP INDEX IF EXISTS idx_client_roles_tenant_id;
DROP INDEX IF EXISTS idx_client_roles_role_id;
DROP INDEX IF EXISTS idx_client_roles_client_id;

DROP INDEX IF EXISTS idx_user_roles_created_at;
DROP INDEX IF EXISTS idx_user_roles_tenant_id;
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_user_id;

DROP INDEX IF EXISTS idx_roles_tenant_name;
DROP INDEX IF EXISTS idx_roles_created_at;
DROP INDEX IF EXISTS idx_roles_name;
DROP INDEX IF EXISTS idx_roles_tenant_id;

DROP INDEX IF EXISTS idx_unverified_clients_created_at;
DROP INDEX IF EXISTS idx_unverified_clients_expires_at;
DROP INDEX IF EXISTS idx_unverified_clients_client_id;
DROP INDEX IF EXISTS idx_unverified_clients_tenant_id;

DROP INDEX IF EXISTS idx_unverified_users_created_at;
DROP INDEX IF EXISTS idx_unverified_users_expires_at;
DROP INDEX IF EXISTS idx_unverified_users_email;
DROP INDEX IF EXISTS idx_unverified_users_username;
DROP INDEX IF EXISTS idx_unverified_users_tenant_id;

DROP INDEX IF EXISTS idx_clients_created_at;
DROP INDEX IF EXISTS idx_clients_active;
DROP INDEX IF EXISTS idx_clients_client_id;
DROP INDEX IF EXISTS idx_clients_tenant_id;

DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_tenant_id;

DROP INDEX IF EXISTS idx_tenants_created_at;
DROP INDEX IF EXISTS idx_tenants_name;
DROP INDEX IF EXISTS idx_tenants_active;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS service_sessions;
DROP TABLE IF EXISTS browser_sessions;
DROP TABLE IF EXISTS tenant_realms;
DROP TABLE IF EXISTS client_roles;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS unverified_clients;
DROP TABLE IF EXISTS unverified_users;
DROP TABLE IF EXISTS clients;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;

-- Recreate old browser_sessions table (without tenant_id)
CREATE TABLE IF NOT EXISTS browser_sessions (
    id TEXT PRIMARY KEY NOT NULL,
    token_hash TEXT,
    expiration TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP,
    realm TEXT,
    user_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_browser_sessions_token_hash ON browser_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_expiration ON browser_sessions(expiration);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_last_activity ON browser_sessions(last_activity);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_user_id ON browser_sessions(user_id);

-- Recreate old service_sessions table (without tenant_id)
CREATE TABLE IF NOT EXISTS service_sessions (
    id TEXT PRIMARY KEY NOT NULL,
    token_hash TEXT,
    expiration TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP,
    realm TEXT,
    client_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_service_sessions_token_hash ON service_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_service_sessions_expiration ON service_sessions(expiration);
CREATE INDEX IF NOT EXISTS idx_service_sessions_last_activity ON service_sessions(last_activity);
CREATE INDEX IF NOT EXISTS idx_service_sessions_client_id ON service_sessions(client_id);
