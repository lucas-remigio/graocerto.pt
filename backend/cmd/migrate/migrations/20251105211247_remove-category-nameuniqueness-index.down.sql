-- Recreate the unique constraint on category names
CREATE UNIQUE INDEX idx_57701_user_category_id ON public.categories 
USING btree (user_id, transaction_type_id, category_name);