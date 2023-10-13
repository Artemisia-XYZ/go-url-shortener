CREATE TABLE short_links (
    id CHAR(36) PRIMARY KEY,
    slash_code VARCHAR(12) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL UNIQUE,
    destination VARCHAR(512) NOT NULL,
    visitors INT UNSIGNED NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_unicode_ci;