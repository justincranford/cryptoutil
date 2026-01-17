--
-- JOSE-JA database schema (SQLite + PostgreSQL compatible)
-- 2001: Elastic JWKs - key containers with versioned material keys
-- Elastic JWKs hold multiple material keys for key rotation support
--

-- Elastic JWKs table: logical key containers supporting rotation
-- Each elastic JWK can have up to max_materials material keys (default: 1000)
-- Active material is used for signing/encrypting, retired materials for verify/decrypt
CREATE TABLE IF NOT EXISTS elastic_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    realm_id TEXT NOT NULL,
    kid TEXT NOT NULL,                              -- External key identifier (unique per tenant+realm)
    kty TEXT NOT NULL,                              -- Key type: RSA, EC, OKP, oct
    alg TEXT NOT NULL,                              -- Algorithm: RS256, ES256, EdDSA, A256GCM, etc.
    use TEXT NOT NULL,                              -- Key use: sig (signing) or enc (encryption)
    max_materials INTEGER NOT NULL DEFAULT 1000,    -- Maximum material versions allowed
    current_material_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id, realm_id) REFERENCES tenant_realms(tenant_id, realm_id),
    UNIQUE(tenant_id, realm_id, kid)
);

CREATE INDEX IF NOT EXISTS idx_elastic_jwks_tenant_realm ON elastic_jwks(tenant_id, realm_id);
CREATE INDEX IF NOT EXISTS idx_elastic_jwks_kid ON elastic_jwks(kid);
CREATE INDEX IF NOT EXISTS idx_elastic_jwks_alg ON elastic_jwks(alg);
CREATE INDEX IF NOT EXISTS idx_elastic_jwks_use ON elastic_jwks(use);
