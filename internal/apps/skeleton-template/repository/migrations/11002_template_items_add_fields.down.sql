-- Skeleton Template database schema rollback
ALTER TABLE template_items DROP COLUMN updated_at;
ALTER TABLE template_items DROP COLUMN description;
ALTER TABLE template_items DROP COLUMN name;
