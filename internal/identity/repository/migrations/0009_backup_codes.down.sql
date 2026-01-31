-- Migration: 0009_backup_codes.down.sql
-- Description: Drop backup codes table.

DROP INDEX IF EXISTS idx_backup_codes_created_at;
DROP INDEX IF EXISTS idx_backup_codes_user_id_used;
DROP INDEX IF EXISTS idx_backup_codes_used;
DROP INDEX IF EXISTS idx_backup_codes_user_id;
DROP TABLE IF EXISTS backup_codes;
