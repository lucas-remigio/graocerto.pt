ALTER TABLE categories ADD COLUMN parent_category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE;

CREATE INDEX idx_categories_parent ON categories(parent_category_id);

-- Add a check constraint to prevent a category from being its own parent
ALTER TABLE categories ADD CONSTRAINT chk_not_self_parent CHECK (id != parent_category_id);