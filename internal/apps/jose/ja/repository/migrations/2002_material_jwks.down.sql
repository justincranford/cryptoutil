--
-- JOSE-JA database schema rollback
-- 2002: Material JWKs - drop material_jwks table
--

DROP INDEX IF EXISTS idx_material_jwks_barrier_version;
DROP INDEX IF EXISTS idx_material_jwks_material_kid;
DROP INDEX IF EXISTS idx_material_jwks_active;
DROP INDEX IF EXISTS idx_material_jwks_elastic;
DROP TABLE IF EXISTS material_jwks;
