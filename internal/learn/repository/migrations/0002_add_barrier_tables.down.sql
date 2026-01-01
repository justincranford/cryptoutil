--
-- Drop barrier encryption tables
--

DROP INDEX IF EXISTS idx_barrier_content_keys_kek_uuid;
DROP TABLE IF EXISTS barrier_content_keys;

DROP INDEX IF EXISTS idx_barrier_intermediate_keys_kek_uuid;
DROP TABLE IF EXISTS barrier_intermediate_keys;

DROP INDEX IF EXISTS idx_barrier_root_keys_kek_uuid;
DROP TABLE IF EXISTS barrier_root_keys;
