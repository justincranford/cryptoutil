--
-- JOSE-JA database schema (SQLite + PostgreSQL compatible)
-- 2004: Audit log - cryptographic operation audit entries
-- CRITICAL: tenant_id for data scoping, session_id for traceability - realms are authn-only
--

-- Audit log table for tracking cryptographic operations
-- Sampling is controlled by tenant_audit_config
-- Retention is configurable per tenant (default: 90 days)
CREATE TABLE IF NOT EXISTS audit_log (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    session_id TEXT,                                -- Session context for operation
    elastic_jwk_id TEXT,                            -- NULL for operations not involving specific key
    material_kid TEXT,                              -- NULL for operations not involving specific material
    operation TEXT NOT NULL,                        -- Operation: generate, sign, verify, encrypt, decrypt, rotate
    success BOOLEAN NOT NULL,                       -- Whether operation succeeded
    error_message TEXT,                             -- Error message if operation failed
    user_id TEXT,                                   -- User who performed operation (NULL for service calls)
    client_id TEXT,                                 -- Client/service that performed operation (NULL for user calls)
    request_id TEXT NOT NULL,                       -- Correlation ID for tracing
    ip_address TEXT,                                -- Client IP address
    user_agent TEXT,                                -- Client user agent
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (elastic_jwk_id) REFERENCES elastic_jwks(id)
);

CREATE INDEX IF NOT EXISTS idx_audit_log_tenant ON audit_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_session ON audit_log(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_request_id ON audit_log(request_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_success ON audit_log(success);
