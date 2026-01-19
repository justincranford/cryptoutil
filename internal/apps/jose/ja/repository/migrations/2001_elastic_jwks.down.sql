--
-- JOSE-JA database schema rollback
-- 2001: Elastic JWKs - drop elastic_jwks table
--

DROP INDEX IF EXISTS idx_elastic_jwks_use;
DROP INDEX IF EXISTS idx_elastic_jwks_alg;
DROP INDEX IF EXISTS idx_elastic_jwks_kid;
DROP INDEX IF EXISTS idx_elastic_jwks_tenant;
