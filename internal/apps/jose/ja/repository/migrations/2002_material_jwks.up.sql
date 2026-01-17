--
-- JOSE-JA database schema (SQLite + PostgreSQL compatible)
-- 2002: Material JWKs - actual key material for elastic JWKs
-- Material keys are encrypted at rest using barrier encryption
--

-- Material JWKs table: encrypted key material for elastic JWKs
-- Each material has private and public JWK encrypted with barrier
-- Only one material can be active per elastic JWK (used for sign/encrypt)
-- Retired materials remain available for verify/decrypt operations
CREATE TABLE IF NOT EXISTS material_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    elastic_jwk_id TEXT NOT NULL,
    material_kid TEXT NOT NULL,                     -- Material-specific KID (unique per elastic JWK)
    private_jwk_jwe TEXT NOT NULL,                  -- JWE-encrypted private JWK (barrier encryption)
    public_jwk_jwe TEXT NOT NULL,                   -- JWE-encrypted public JWK (barrier encryption)
    active BOOLEAN NOT NULL DEFAULT FALSE,          -- TRUE if this is the current active material
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    retired_at TIMESTAMP,                           -- When material was retired (rotated out)
    barrier_version INTEGER NOT NULL,               -- Barrier key version used for encryption
    FOREIGN KEY (elastic_jwk_id) REFERENCES elastic_jwks(id),
    UNIQUE(elastic_jwk_id, material_kid)
);

CREATE INDEX IF NOT EXISTS idx_material_jwks_elastic ON material_jwks(elastic_jwk_id);
CREATE INDEX IF NOT EXISTS idx_material_jwks_active ON material_jwks(elastic_jwk_id, active);
CREATE INDEX IF NOT EXISTS idx_material_jwks_material_kid ON material_jwks(material_kid);
CREATE INDEX IF NOT EXISTS idx_material_jwks_barrier_version ON material_jwks(barrier_version);
