WITH ranked AS (
    SELECT
        id,
        ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at, id) AS rn
    FROM accounts
)
UPDATE accounts
SET order_index = ranked.rn
FROM ranked
WHERE accounts.id = ranked.id;