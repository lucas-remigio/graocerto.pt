DROP INDEX IF EXISTS idx_categories_parent;

ALTER TABLE categories DROP CONSTRAINT IF EXISTS chk_not_self_parent;

ALTER TABLE categories DROP COLUMN IF EXISTS parent_category_id;