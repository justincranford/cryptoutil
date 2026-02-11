--
-- Identity domain database schema (SQLite)
--

-- Users table with OIDC standard claims
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY NOT NULL,

    -- OIDC standard claims
    sub TEXT NOT NULL,
    name TEXT,
    given_name TEXT,
    family_name TEXT,
    middle_name TEXT,
    nickname TEXT,
    preferred_username TEXT,
    profile TEXT,
    picture TEXT,
    website TEXT,
    email TEXT,
    email_verified INTEGER DEFAULT 0,
    gender TEXT,
    birthdate TEXT,
    zoneinfo TEXT,
    locale TEXT,
    phone_number TEXT,
    phone_number_verified INTEGER DEFAULT 0,

    -- Address embedded fields (with address_ prefix)
    address_formatted TEXT,
    address_street_address TEXT,
    address_locality TEXT,
    address_region TEXT,
    address_postal_code TEXT,
    address_country TEXT,

    -- MFA device tokens
    push_device_token TEXT,

    -- Authentication credentials
    password_hash TEXT NOT NULL,

    -- Account status
    enabled BOOLEAN DEFAULT true,
    locked BOOLEAN DEFAULT false,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create unique indexes for OIDC identifier fields
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_sub ON users(sub);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_preferred_username ON users(preferred_username) WHERE preferred_username IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Client profiles table (must exist before clients table for foreign keys)
CREATE TABLE IF NOT EXISTS client_profiles (
    id TEXT PRIMARY KEY NOT NULL,

    -- Profile metadata
    name TEXT NOT NULL,
    description TEXT,

    -- Scope configuration (JSON arrays)
    required_scopes TEXT DEFAULT '[]',
    optional_scopes TEXT DEFAULT '[]',

    -- Consent configuration
    consent_screen_count INTEGER DEFAULT 1,
    consent_screen_1_text TEXT,
    consent_screen_2_text TEXT,

    -- MFA configuration for client authentication (JSON arrays)
    require_client_mfa BOOLEAN DEFAULT false,
    client_mfa_chain TEXT DEFAULT '[]',

    -- Account status
    enabled BOOLEAN DEFAULT true,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_client_profiles_name ON client_profiles(name);
CREATE INDEX IF NOT EXISTS idx_client_profiles_deleted_at ON client_profiles(deleted_at);

-- Clients table for OAuth 2.1 client configuration
CREATE TABLE IF NOT EXISTS clients (
    id TEXT PRIMARY KEY NOT NULL,

    -- Client identification
    client_id TEXT NOT NULL,
    client_secret TEXT,
    client_type TEXT NOT NULL,

    -- Client JWK Set (for private_key_jwt authentication)
    j_w_ks TEXT,

    -- Client metadata
    name TEXT NOT NULL,
    description TEXT,
    logo_uri TEXT,
    home_page_uri TEXT,
    policy_uri TEXT,
    tos_uri TEXT,

    -- OAuth 2.1 configuration (JSON arrays for string slices)
    redirect_uris TEXT DEFAULT '[]',
    allowed_grant_types TEXT DEFAULT '[]',
    allowed_response_types TEXT DEFAULT '[]',
    allowed_scopes TEXT DEFAULT '[]',
    token_endpoint_auth_method TEXT NOT NULL,

    -- PKCE configuration
    require_pkce BOOLEAN DEFAULT true,
    pkce_challenge_method TEXT DEFAULT 'S256',

    -- Token configuration (lifetimes in seconds)
    access_token_lifetime INTEGER DEFAULT 3600,
    refresh_token_lifetime INTEGER DEFAULT 86400,
    id_token_lifetime INTEGER DEFAULT 3600,

    -- Client profile reference (optional foreign key)
    client_profile_id TEXT,

    -- Certificate-based authentication fields (added R04)
    certificate_subject TEXT,
    certificate_fingerprint TEXT,

    -- Account status
    enabled BOOLEAN DEFAULT true,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Foreign key constraint
    FOREIGN KEY (client_profile_id) REFERENCES client_profiles(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_client_id ON clients(client_id);
CREATE INDEX IF NOT EXISTS idx_clients_certificate_subject ON clients(certificate_subject) WHERE certificate_subject IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_clients_certificate_fingerprint ON clients(certificate_fingerprint) WHERE certificate_fingerprint IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_clients_client_profile_id ON clients(client_profile_id);
CREATE INDEX IF NOT EXISTS idx_clients_deleted_at ON clients(deleted_at);

-- Tokens table for OAuth 2.1 / OIDC tokens
CREATE TABLE IF NOT EXISTS tokens (
    id TEXT PRIMARY KEY NOT NULL,

    -- Token identification
    token_value TEXT NOT NULL,
    token_type TEXT NOT NULL,
    token_format TEXT NOT NULL,

    -- Token associations (foreign keys)
    client_id TEXT NOT NULL,
    user_id TEXT,

    -- Token metadata (JSON array for scopes)
    scopes TEXT DEFAULT '[]',
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    not_before TIMESTAMP,

    -- Token status
    revoked INTEGER DEFAULT 0,
    revoked_at TIMESTAMP,

    -- Refresh token association (optional foreign key)
    refresh_token_id TEXT,

    -- PKCE code challenge (for authorization codes)
    code_challenge TEXT,
    code_challenge_method TEXT,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Foreign key constraints
    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (refresh_token_id) REFERENCES tokens(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tokens_token_value ON tokens(token_value);
CREATE INDEX IF NOT EXISTS idx_tokens_token_type ON tokens(token_type);
CREATE INDEX IF NOT EXISTS idx_tokens_client_id ON tokens(client_id);
CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_refresh_token_id ON tokens(refresh_token_id);
CREATE INDEX IF NOT EXISTS idx_tokens_issued_at ON tokens(issued_at);
CREATE INDEX IF NOT EXISTS idx_tokens_expires_at ON tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_tokens_revoked ON tokens(revoked);
CREATE INDEX IF NOT EXISTS idx_tokens_deleted_at ON tokens(deleted_at);

-- Sessions table for user authentication sessions
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY NOT NULL,

    -- Session identification
    session_id TEXT NOT NULL,

    -- Session associations (foreign keys)
    user_id TEXT NOT NULL,
    client_id TEXT,

    -- Session metadata
    ip_address TEXT,
    user_agent TEXT,

    -- Session lifetime
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    last_seen_at TIMESTAMP,

    -- Session status
    active BOOLEAN DEFAULT true,
    revoked_at TIMESTAMP,

    -- Authentication context (JSON arrays)
    authentication_methods TEXT DEFAULT '[]',
    authentication_time TIMESTAMP,

    -- OIDC context
    nonce TEXT,
    code_challenge TEXT,
    granted_scopes TEXT DEFAULT '[]',

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Foreign key constraints
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_client_id ON sessions(client_id);
CREATE INDEX IF NOT EXISTS idx_sessions_issued_at ON sessions(issued_at);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(active);
CREATE INDEX IF NOT EXISTS idx_sessions_deleted_at ON sessions(deleted_at);

-- Auth profiles table (must exist before mfa_factors table for foreign keys)
CREATE TABLE IF NOT EXISTS auth_profiles (
    id TEXT PRIMARY KEY NOT NULL,

    -- Profile metadata
    name TEXT NOT NULL,
    description TEXT,
    profile_type TEXT NOT NULL,

    -- MFA configuration (JSON arrays)
    require_mfa BOOLEAN DEFAULT false,
    mfa_chain TEXT DEFAULT '[]',

    -- mTLS configuration (JSON array)
    mtls_domains TEXT DEFAULT '[]',

    -- Account status
    enabled BOOLEAN DEFAULT true,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_profiles_name ON auth_profiles(name);
CREATE INDEX IF NOT EXISTS idx_auth_profiles_deleted_at ON auth_profiles(deleted_at);

-- MFA factors table for multi-factor authentication factor configuration
CREATE TABLE IF NOT EXISTS mfa_factors (
    id TEXT PRIMARY KEY NOT NULL,

    -- Factor metadata
    name TEXT NOT NULL,
    description TEXT,
    factor_type TEXT NOT NULL,

    -- Factor ordering
    "order" INTEGER NOT NULL,

    -- Factor configuration
    required INTEGER DEFAULT 0,

    -- TOTP/HOTP configuration
    totp_algorithm TEXT,
    totp_digits INTEGER,
    totp_period INTEGER,

    -- Authentication profile reference (foreign key)
    auth_profile_id TEXT NOT NULL,

    -- Replay prevention (time-bound nonces)
    nonce TEXT,
    nonce_expires_at TIMESTAMP,
    nonce_used_at TIMESTAMP,

    -- Account status
    enabled BOOLEAN DEFAULT true,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Foreign key constraint
    FOREIGN KEY (auth_profile_id) REFERENCES auth_profiles(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_mfa_factors_name ON mfa_factors(name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_mfa_factors_nonce ON mfa_factors(nonce);
CREATE INDEX IF NOT EXISTS idx_mfa_factors_auth_profile_id ON mfa_factors(auth_profile_id);
CREATE INDEX IF NOT EXISTS idx_mfa_factors_nonce_expires_at ON mfa_factors(nonce_expires_at);
CREATE INDEX IF NOT EXISTS idx_mfa_factors_nonce_used_at ON mfa_factors(nonce_used_at);
CREATE INDEX IF NOT EXISTS idx_mfa_factors_deleted_at ON mfa_factors(deleted_at);

-- Auth flows table for authorization code flow configuration
CREATE TABLE IF NOT EXISTS auth_flows (
    id TEXT PRIMARY KEY NOT NULL,

    -- Flow metadata
    name TEXT NOT NULL,
    description TEXT,
    flow_type TEXT NOT NULL,

    -- PKCE configuration
    require_pkce BOOLEAN DEFAULT true,
    pkce_challenge_method TEXT DEFAULT 'S256',

    -- Scope configuration (JSON array)
    allowed_scopes TEXT DEFAULT '[]',

    -- Consent configuration
    require_consent BOOLEAN DEFAULT true,
    consent_screen_count INTEGER DEFAULT 1,
    remember_consent BOOLEAN DEFAULT false,

    -- State parameter configuration
    require_state BOOLEAN DEFAULT true,

    -- Client profile reference (optional foreign key)
    client_profile_id TEXT,

    -- Account status
    enabled BOOLEAN DEFAULT true,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Foreign key constraint
    FOREIGN KEY (client_profile_id) REFERENCES client_profiles(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_flows_name ON auth_flows(name);
CREATE INDEX IF NOT EXISTS idx_auth_flows_client_profile_id ON auth_flows(client_profile_id);
CREATE INDEX IF NOT EXISTS idx_auth_flows_deleted_at ON auth_flows(deleted_at);
