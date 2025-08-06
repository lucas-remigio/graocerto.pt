CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    account_token VARCHAR(255) NOT NULL,
    category_id INTEGER DEFAULT NULL,
    amount NUMERIC(15, 2) NOT NULL,
    description TEXT DEFAULT NULL,
    date TIMESTAMPTZ NOT NULL,
    balance NUMERIC(15, 2) DEFAULT 0.00,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (account_token) REFERENCES accounts(token) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
);