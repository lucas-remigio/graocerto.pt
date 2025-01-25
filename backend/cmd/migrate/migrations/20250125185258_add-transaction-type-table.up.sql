CREATE TABLE IF NOT EXISTS transaction_types (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY NOT NULL,
    `type_name` VARCHAR(50) NOT NULL,
    `type_slug` VARCHAR(50) NOT NULL,

    UNIQUE KEY `type_name` (`type_name`)
);

