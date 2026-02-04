--
-- KMS Business Tables Migration - Rollback
-- Domain migration 2001: elastic_keys and material_keys tables
--

-- Drop indexes first
DROP INDEX IF EXISTS idx_material_keys_expiration;
DROP INDEX IF EXISTS idx_material_keys_generate_date;
DROP INDEX IF EXISTS idx_material_keys_elastic_key;
DROP INDEX IF EXISTS idx_elastic_keys_status;
DROP INDEX IF EXISTS idx_elastic_keys_name;

-- Drop tables in reverse order (material_keys depends on elastic_keys)
DROP TABLE IF EXISTS material_keys;
DROP TABLE IF EXISTS elastic_keys;
