--
-- Barrier encryption tables rollback
-- Drops barrier_content_keys, barrier_intermediate_keys, barrier_root_keys
--

-- Drop indexes first (some databases require this)
DROP INDEX IF EXISTS idx_barrier_content_keys_rotated_at;
DROP INDEX IF EXISTS idx_barrier_content_keys_kek_uuid;
DROP INDEX IF EXISTS idx_barrier_intermediate_keys_rotated_at;
DROP INDEX IF EXISTS idx_barrier_intermediate_keys_kek_uuid;
DROP INDEX IF EXISTS idx_barrier_root_keys_rotated_at;
DROP INDEX IF EXISTS idx_barrier_root_keys_kek_uuid;

-- Drop tables in reverse order (foreign key dependencies)
DROP TABLE IF EXISTS barrier_content_keys;
DROP TABLE IF EXISTS barrier_intermediate_keys;
DROP TABLE IF EXISTS barrier_root_keys;
