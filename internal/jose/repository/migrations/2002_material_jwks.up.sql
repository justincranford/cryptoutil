-- Copyright (c) 2025 Justin Cranford
--
--

-- Material JWKs: Actual cryptographic key material for Elastic JWKs.
-- Each Material JWK is a versioned key used for encryption/signing.
-- Active key used for new operations, retired keys used for decryption/verification.
CREATE TABLE IF NOT EXISTS material_jwks (
    id TEXT PRIMARY KEY NOT NULL,
    elastic_jwk_id TEXT NOT NULL,
    material_kid TEXT NOT NULL,
    private_jwk_jwe TEXT NOT NULL,  -- Private key encrypted with barrier.
    public_jwk_jwe TEXT NOT NULL,   -- Public key encrypted with barrier.
    active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    retired_at TIMESTAMP,
    barrier_version INTEGER NOT NULL,
    FOREIGN KEY (elastic_jwk_id) REFERENCES elastic_jwks(id),
    UNIQUE(elastic_jwk_id, material_kid)
);

CREATE INDEX IF NOT EXISTS idx_material_jwks_elastic ON material_jwks(elastic_jwk_id);
CREATE INDEX IF NOT EXISTS idx_material_jwks_active ON material_jwks(elastic_jwk_id, active);
