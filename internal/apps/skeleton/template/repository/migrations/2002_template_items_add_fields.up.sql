-- Skeleton Template database schema
-- Migration range 2001+: domain-specific (non-template).
ALTER TABLE template_items ADD COLUMN name TEXT NOT NULL DEFAULT '';
ALTER TABLE template_items ADD COLUMN description TEXT DEFAULT '';
ALTER TABLE template_items ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
