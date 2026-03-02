WITH ranked AS (
    SELECT
        id,
        ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at, id) AS rn
    FROM categories
)
UPDATE categories
SET order_index = ranked.rn
FROM ranked
WHERE categories.id = ranked.id;
