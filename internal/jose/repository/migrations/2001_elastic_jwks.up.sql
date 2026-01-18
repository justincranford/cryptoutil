-- Copyright (c) 2025 Justin Cranford
--
--

-- Elastic JWKs: Key rings with multiple material JWKs (key rotation support).
-- Each Elastic JWK represents a logical key that can have many Material JWKs.
CREATE TABLE IF NOT EXISTS elastic_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    realm_id TEXT NOT NULL,
    kid TEXT NOT NULL,
    kty TEXT NOT NULL,  -- RSA, EC, oct.
    alg TEXT NOT NULL,  -- RS256, ES256, A256GCM, etc.
    use TEXT NOT NULL,  -- sig, enc.
    max_materials INTEGER NOT NULL DEFAULT 1000,
    current_material_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id, realm_id) REFERENCES tenant_realms(tenant_id, realm_id),
    UNIQUE(tenant_id, realm_id, kid)
);

CREATE INDEX IF NOT EXISTS idx_elastic_jwks_tenant_realm ON elastic_jwks(tenant_id, realm_id);
CREATE INDEX IF NOT EXISTS idx_elastic_jwks_kid ON elastic_jwks(kid);
