--
-- Key Management database schema rollback
-- messages table only (recipient JWK mapping rolled back in 2008)
-- Users table is provided by template service (1004_add_multi_tenancy)
--

DROP INDEX IF EXISTS idx_messages_read_at;
DROP INDEX IF EXISTS idx_messages_created_at;
DROP INDEX IF EXISTS idx_messages_sender_id;
DROP TABLE IF EXISTS messages;
