
UPDATE accounts a
JOIN (
    SELECT
        id,
        ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at, id) AS rn
    FROM accounts
) ranked ON a.id = ranked.id
SET a.order_index = ranked.rn;