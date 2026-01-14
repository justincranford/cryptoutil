-- Migration: Add cipher_im_realms table for authentication realm configuration
-- Supports 6 non-federated authn methods with config & DB storage
-- Config > DB priority pattern (config file overrides database)

CREATE TABLE IF NOT EXISTS cipher_im_realms (
    id TEXT PRIMARY KEY,
    realm_id TEXT NOT NULL,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    config TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    source TEXT NOT NULL DEFAULT 'db',
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient lookups.
CREATE UNIQUE INDEX IF NOT EXISTS idx_realms_realm_id ON cipher_im_realms(realm_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_realms_name ON cipher_im_realms(name);
CREATE INDEX IF NOT EXISTS idx_realms_type ON cipher_im_realms(type);
CREATE INDEX IF NOT EXISTS idx_realms_active ON cipher_im_realms(active);
CREATE INDEX IF NOT EXISTS idx_realms_priority ON cipher_im_realms(priority);

-- Insert default realms for 6 non-federated authn methods.
-- JWE Session Cookie realm (browser).
INSERT INTO cipher_im_realms (id, realm_id, type, name, config, active, source, priority)
SELECT
    '01960000-0000-0000-0000-000000000001',
    '01960000-0000-0000-0000-000000000011',
    'jwe-session-cookie',
    'JWE Session Cookie',
    '{"session_timeout":3600,"session_absolute_max":86400,"session_refresh_enabled":true,"mfa_required":false,"mfa_methods":[],"login_rate_limit":5,"message_rate_limit":10}',
    true,
    'db',
    100
WHERE NOT EXISTS (SELECT 1 FROM cipher_im_realms WHERE name = 'JWE Session Cookie');

-- JWS Session Cookie realm (browser).
INSERT INTO cipher_im_realms (id, realm_id, type, name, config, active, source, priority)
SELECT
    '01960000-0000-0000-0000-000000000002',
    '01960000-0000-0000-0000-000000000012',
    'jws-session-cookie',
    'JWS Session Cookie',
    '{"session_timeout":3600,"session_absolute_max":86400,"session_refresh_enabled":true,"mfa_required":false,"mfa_methods":[],"login_rate_limit":5,"message_rate_limit":10}',
    true,
    'db',
    90
WHERE NOT EXISTS (SELECT 1 FROM cipher_im_realms WHERE name = 'JWS Session Cookie');

-- OPAQUE Session Cookie realm (browser).
INSERT INTO cipher_im_realms (id, realm_id, type, name, config, active, source, priority)
SELECT
    '01960000-0000-0000-0000-000000000003',
    '01960000-0000-0000-0000-000000000013',
    'opaque-session-cookie',
    'OPAQUE Session Cookie',
    '{"session_timeout":3600,"session_absolute_max":86400,"session_refresh_enabled":true,"mfa_required":false,"mfa_methods":[],"login_rate_limit":5,"message_rate_limit":10}',
    true,
    'db',
    80
WHERE NOT EXISTS (SELECT 1 FROM cipher_im_realms WHERE name = 'OPAQUE Session Cookie');

-- Basic Auth realm (username/password for browser, client ID/secret for headless).
INSERT INTO cipher_im_realms (id, realm_id, type, name, config, active, source, priority)
SELECT
    '01960000-0000-0000-0000-000000000004',
    '01960000-0000-0000-0000-000000000014',
    'basic-username-password',
    'Basic Username Password',
    '{"password_min_length":12,"password_require_uppercase":true,"password_require_lowercase":true,"password_require_digits":true,"password_require_special":true,"password_min_unique_chars":8,"password_max_repeated_chars":3,"mfa_required":false,"mfa_methods":[],"login_rate_limit":5,"message_rate_limit":10}',
    true,
    'db',
    70
WHERE NOT EXISTS (SELECT 1 FROM cipher_im_realms WHERE name = 'Basic Username Password');

-- Bearer Token realm (API tokens for both browser and headless).
INSERT INTO cipher_im_realms (id, realm_id, type, name, config, active, source, priority)
SELECT
    '01960000-0000-0000-0000-000000000005',
    '01960000-0000-0000-0000-000000000015',
    'bearer-api-token',
    'Bearer API Token',
    '{"token_expiry":3600,"mfa_required":false,"mfa_methods":[],"login_rate_limit":5,"message_rate_limit":10}',
    true,
    'db',
    60
WHERE NOT EXISTS (SELECT 1 FROM cipher_im_realms WHERE name = 'Bearer API Token');

-- HTTPS Client Cert realm (mTLS for both browser and headless).
INSERT INTO cipher_im_realms (id, realm_id, type, name, config, active, source, priority)
SELECT
    '01960000-0000-0000-0000-000000000006',
    '01960000-0000-0000-0000-000000000016',
    'https-client-cert',
    'HTTPS Client Certificate',
    '{"require_client_cert":true,"trusted_cas":[],"mfa_required":false,"mfa_methods":[],"login_rate_limit":5,"message_rate_limit":10}',
    true,
    'db',
    50
WHERE NOT EXISTS (SELECT 1 FROM cipher_im_realms WHERE name = 'HTTPS Client Certificate');
