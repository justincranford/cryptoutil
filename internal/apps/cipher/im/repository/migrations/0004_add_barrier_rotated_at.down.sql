--
-- Rollback: Remove rotated_at timestamp column from barrier tables
--

-- Remove rotated_at from barrier_root_keys
ALTER TABLE barrier_root_keys
DROP COLUMN rotated_at;

-- Remove rotated_at from barrier_intermediate_keys
ALTER TABLE barrier_intermediate_keys
DROP COLUMN rotated_at;

-- Remove rotated_at from barrier_content_keys
ALTER TABLE barrier_content_keys
DROP COLUMN rotated_at;
