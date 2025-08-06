CREATE TABLE IF NOT EXISTS transaction_types (
    id SERIAL PRIMARY KEY,
    type_name VARCHAR(50) NOT NULL,
    type_slug VARCHAR(50) NOT NULL,
    UNIQUE (type_name)
);

