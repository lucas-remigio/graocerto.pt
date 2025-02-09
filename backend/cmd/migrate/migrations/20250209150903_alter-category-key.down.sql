ALTER TABLE categories
  DROP FOREIGN KEY categories_ibfk_1,
  DROP INDEX user_category,
  ADD UNIQUE KEY user_category (user_id, category_name),
  ADD CONSTRAINT categories_ibfk_1 FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;