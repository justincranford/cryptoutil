--
-- Realms template table rollback
--

-- Drop indexes first
DROP INDEX IF EXISTS idx_template_realms_priority;
DROP INDEX IF EXISTS idx_template_realms_active;
DROP INDEX IF EXISTS idx_template_realms_type;
DROP INDEX IF EXISTS idx_template_realms_name;
DROP INDEX IF EXISTS idx_template_realms_realm_id;

-- Drop table
DROP TABLE IF EXISTS template_realms;
