-- skeleton-template domain migration: remove name, description, updated_at from template_items.
ALTER TABLE template_items DROP COLUMN updated_at;
ALTER TABLE template_items DROP COLUMN description;
ALTER TABLE template_items DROP COLUMN name;
