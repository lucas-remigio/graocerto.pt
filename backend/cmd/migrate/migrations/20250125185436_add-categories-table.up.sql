CREATE TABLE IF NOT EXISTS categories (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY NOT NULL,
    `user_id` INT UNSIGNED NOT NULL,
    `transaction_type_id` INT UNSIGNED NOT NULL,
    `category_name` VARCHAR(255) NOT NULL,
    `color` VARCHAR(7) NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (`user_id`) REFERENCES users(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`transaction_type_id`) REFERENCES transaction_types(`id`) ON DELETE CASCADE,
    
    UNIQUE KEY `user_category` (`user_id`, `transaction_type_id`, `category_name`)
);