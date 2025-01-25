CREATE TABLE IF NOT EXISTS transactions (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY NOT NULL,
    `account_token` VARCHAR(255) NOT NULL,
    `transaction_type_id` INT UNSIGNED NOT NULL,
    `category_id` INT UNSIGNED DEFAULT NULL,
    `amount` DECIMAL(15, 2) NOT NULL,
    `description` TEXT DEFAULT NULL,
    `date` DATETIME NOT NULL,
    `balance` DECIMAL(15, 2) DEFAULT 0.00,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (`account_token`) REFERENCES accounts(`token`) ON DELETE CASCADE,
    FOREIGN KEY (`transaction_type_id`) REFERENCES transaction_types(`id`) ON DELETE RESTRICT,
    FOREIGN KEY (`category_id`) REFERENCES categories(`id`) ON DELETE SET NULL
);