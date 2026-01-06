--
-- Cipher IM database schema rollback
-- 3-table design: users, messages, messages_recipient_jwks
--

DROP INDEX IF EXISTS idx_messages_recipient_jwks_message_id;
DROP INDEX IF EXISTS idx_messages_recipient_jwks_recipient_id;
DROP TABLE IF EXISTS messages_recipient_jwks;

DROP INDEX IF EXISTS idx_messages_read_at;
DROP INDEX IF EXISTS idx_messages_created_at;
DROP INDEX IF EXISTS idx_messages_sender_id;
DROP TABLE IF EXISTS messages;

DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users;
