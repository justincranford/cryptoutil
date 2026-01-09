--
-- Add rotated_at timestamp column to barrier tables for key rotation tracking
--

-- Add rotated_at to barrier_root_keys
ALTER TABLE barrier_root_keys
ADD COLUMN rotated_at INTEGER;  -- Unix epoch milliseconds (NULL = still active)

-- Add rotated_at to barrier_intermediate_keys
ALTER TABLE barrier_intermediate_keys
ADD COLUMN rotated_at INTEGER;  -- Unix epoch milliseconds (NULL = still active)

-- Add rotated_at to barrier_content_keys
ALTER TABLE barrier_content_keys
ADD COLUMN rotated_at INTEGER;  -- Unix epoch milliseconds (NULL = still active)
