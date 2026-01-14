--
-- Cipher IM database schema rollback
-- 2-table design: messages, messages_recipient_jwks
-- Users table is provided by template service (1004_add_multi_tenancy)
--

DROP INDEX IF EXISTS idx_messages_recipient_jwks_message_id;
DROP INDEX IF EXISTS idx_messages_recipient_jwks_recipient_id;
DROP TABLE IF EXISTS messages_recipient_jwks;

DROP INDEX IF EXISTS idx_messages_read_at;
DROP INDEX IF EXISTS idx_messages_created_at;
DROP INDEX IF EXISTS idx_messages_sender_id;
DROP TABLE IF EXISTS messages;
