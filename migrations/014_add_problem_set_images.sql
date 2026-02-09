-- 题目集图片表（用于 PDF 生成中的图片支持）
CREATE TABLE IF NOT EXISTS `xcpc_problem_set_image` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `set_id` BIGINT UNSIGNED NOT NULL COMMENT '题目集ID',
    `filename` VARCHAR(255) NOT NULL COMMENT '原始文件名',
    `file_path` VARCHAR(500) NOT NULL COMMENT '存储路径',
    `file_size` BIGINT NOT NULL COMMENT '文件大小（字节）',
    `mime_type` VARCHAR(100) DEFAULT NULL COMMENT 'MIME类型',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_set_id` (`set_id`),
    UNIQUE KEY `uk_set_filename` (`set_id`, `filename`),
    FOREIGN KEY (`set_id`) REFERENCES `xcpc_problem_set`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目集图片';
