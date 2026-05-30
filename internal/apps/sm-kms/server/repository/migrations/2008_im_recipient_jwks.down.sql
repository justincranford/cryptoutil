--
-- Key Management database schema rollback
--

DROP INDEX IF EXISTS idx_messages_recipient_jwks_message_id;
DROP INDEX IF EXISTS idx_messages_recipient_jwks_recipient_id;
DROP TABLE IF EXISTS messages_recipient_jwks;
