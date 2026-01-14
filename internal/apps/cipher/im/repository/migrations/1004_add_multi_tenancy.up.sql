--
-- Multi-Tenancy tables for template service
-- Provides tenant isolation with users, clients, roles, and realm configuration
--

-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    active INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tenants_active ON tenants(active);
CREATE INDEX IF NOT EXISTS idx_tenants_name ON tenants(name);
CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at);

-- Users table (verified users)
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    email TEXT,
    active INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(active);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Clients table (verified non-browser clients)
CREATE TABLE IF NOT EXISTS clients (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    client_id TEXT NOT NULL UNIQUE,
    client_secret TEXT NOT NULL,
    name TEXT,
    active INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_clients_tenant_id ON clients(tenant_id);
CREATE INDEX IF NOT EXISTS idx_clients_client_id ON clients(client_id);
CREATE INDEX IF NOT EXISTS idx_clients_active ON clients(active);
CREATE INDEX IF NOT EXISTS idx_clients_created_at ON clients(created_at);

-- Unverified Users table (pending admin verification)
CREATE TABLE IF NOT EXISTS unverified_users (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    email TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_unverified_users_tenant_id ON unverified_users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_unverified_users_username ON unverified_users(username);
CREATE INDEX IF NOT EXISTS idx_unverified_users_email ON unverified_users(email);
CREATE INDEX IF NOT EXISTS idx_unverified_users_expires_at ON unverified_users(expires_at);
CREATE INDEX IF NOT EXISTS idx_unverified_users_created_at ON unverified_users(created_at);

-- Unverified Clients table (pending admin verification)
CREATE TABLE IF NOT EXISTS unverified_clients (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    client_id TEXT NOT NULL UNIQUE,
    client_secret TEXT NOT NULL,
    name TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_unverified_clients_tenant_id ON unverified_clients(tenant_id);
CREATE INDEX IF NOT EXISTS idx_unverified_clients_client_id ON unverified_clients(client_id);
CREATE INDEX IF NOT EXISTS idx_unverified_clients_expires_at ON unverified_clients(expires_at);
CREATE INDEX IF NOT EXISTS idx_unverified_clients_created_at ON unverified_clients(created_at);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
CREATE INDEX IF NOT EXISTS idx_roles_created_at ON roles(created_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_tenant_name ON roles(tenant_id, name);

-- User Roles junction table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_tenant_id ON user_roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_created_at ON user_roles(created_at);

-- Client Roles junction table
CREATE TABLE IF NOT EXISTS client_roles (
    client_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (client_id, role_id),
    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_client_roles_client_id ON client_roles(client_id);
CREATE INDEX IF NOT EXISTS idx_client_roles_role_id ON client_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_client_roles_tenant_id ON client_roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_client_roles_created_at ON client_roles(created_at);

-- Tenant Realms table (per-tenant realm configuration)
CREATE TABLE IF NOT EXISTS tenant_realms (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    realm_id TEXT NOT NULL,
    type TEXT NOT NULL,                     -- username_password, ldap, oauth2
    config TEXT,                            -- JSON configuration
    active INTEGER NOT NULL DEFAULT 1,
    source TEXT NOT NULL DEFAULT 'db',     -- db or file
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tenant_realms_tenant_id ON tenant_realms(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_realms_realm_id ON tenant_realms(realm_id);
CREATE INDEX IF NOT EXISTS idx_tenant_realms_type ON tenant_realms(type);
CREATE INDEX IF NOT EXISTS idx_tenant_realms_active ON tenant_realms(active);
CREATE INDEX IF NOT EXISTS idx_tenant_realms_source ON tenant_realms(source);
CREATE INDEX IF NOT EXISTS idx_tenant_realms_created_at ON tenant_realms(created_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_realms_tenant_realm ON tenant_realms(tenant_id, realm_id);

-- Update browser_sessions to include tenant_id
-- Drop and recreate with tenant_id foreign key
DROP INDEX IF EXISTS idx_browser_sessions_token_hash;
DROP INDEX IF EXISTS idx_browser_sessions_expiration;
DROP INDEX IF EXISTS idx_browser_sessions_last_activity;
DROP INDEX IF EXISTS idx_browser_sessions_user_id;
DROP TABLE IF EXISTS browser_sessions;

CREATE TABLE IF NOT EXISTS browser_sessions (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    realm_id TEXT NOT NULL,
    token_hash TEXT,
    expiration TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP,
    user_id TEXT,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_browser_sessions_tenant_id ON browser_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_realm_id ON browser_sessions(realm_id);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_token_hash ON browser_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_expiration ON browser_sessions(expiration);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_last_activity ON browser_sessions(last_activity);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_user_id ON browser_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_created_at ON browser_sessions(created_at);

-- Update service_sessions to include tenant_id
-- Drop and recreate with tenant_id foreign key
DROP INDEX IF EXISTS idx_service_sessions_token_hash;
DROP INDEX IF EXISTS idx_service_sessions_expiration;
DROP INDEX IF EXISTS idx_service_sessions_last_activity;
DROP INDEX IF EXISTS idx_service_sessions_client_id;
DROP TABLE IF EXISTS service_sessions;

CREATE TABLE IF NOT EXISTS service_sessions (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    realm_id TEXT NOT NULL,
    token_hash TEXT,
    expiration TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP,
    client_id TEXT,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_service_sessions_tenant_id ON service_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_service_sessions_realm_id ON service_sessions(realm_id);
CREATE INDEX IF NOT EXISTS idx_service_sessions_token_hash ON service_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_service_sessions_expiration ON service_sessions(expiration);
CREATE INDEX IF NOT EXISTS idx_service_sessions_last_activity ON service_sessions(last_activity);
CREATE INDEX IF NOT EXISTS idx_service_sessions_client_id ON service_sessions(client_id);
CREATE INDEX IF NOT EXISTS idx_service_sessions_created_at ON service_sessions(created_at);
