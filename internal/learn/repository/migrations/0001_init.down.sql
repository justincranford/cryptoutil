--
-- Learn IM database schema rollback
--

DROP INDEX IF EXISTS idx_messages_read_at;
DROP INDEX IF EXISTS idx_messages_created_at;
DROP INDEX IF EXISTS idx_messages_key_id;
DROP INDEX IF EXISTS idx_messages_recipient_id;
DROP INDEX IF EXISTS idx_messages_sender_id;
DROP TABLE IF EXISTS messages;

DROP INDEX IF EXISTS idx_messages_jwks_is_active;
DROP INDEX IF EXISTS idx_messages_jwks_key_id;
DROP INDEX IF EXISTS idx_messages_jwks_message_id;
DROP TABLE IF EXISTS messages_jwks;

DROP INDEX IF EXISTS idx_users_messages_jwks_key_id;
DROP INDEX IF EXISTS idx_users_messages_jwks_message_id;
DROP INDEX IF EXISTS idx_users_messages_jwks_user_id;
DROP TABLE IF EXISTS users_messages_jwks;

DROP INDEX IF EXISTS idx_users_jwks_is_active;
DROP INDEX IF EXISTS idx_users_jwks_key_id;
DROP INDEX IF EXISTS idx_users_jwks_user_id;
DROP TABLE IF EXISTS users_jwks;

DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users;

DROP INDEX IF EXISTS idx_users_jwks_is_active;
DROP INDEX IF EXISTS idx_users_jwks_key_id;
DROP INDEX IF EXISTS idx_users_jwks_user_id;
DROP TABLE IF EXISTS users_jwks;
