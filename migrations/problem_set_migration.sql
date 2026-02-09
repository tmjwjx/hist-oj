-- XCPC 题目集 PDF 生成器相关表

-- 题目集表
CREATE TABLE IF NOT EXISTS `xcpc_problem_set` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '创建者ID',
    `title` VARCHAR(255) NOT NULL COMMENT '比赛/题集名称',
    `author` VARCHAR(100) DEFAULT NULL COMMENT '作者',
    `contest_date` DATE DEFAULT NULL COMMENT '比赛日期',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户题目集';

-- 题目表
CREATE TABLE IF NOT EXISTS `xcpc_problem_set_problem` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `set_id` BIGINT UNSIGNED NOT NULL COMMENT '题目集ID',
    `problem_letter` CHAR(1) DEFAULT NULL COMMENT '题目字母（A, B, C...）',
    `title` VARCHAR(255) NOT NULL COMMENT '题目标题',
    `time_limit` INT DEFAULT 1000 COMMENT '时间限制(ms)',
    `description` TEXT DEFAULT NULL COMMENT '题目描述',
    `input_format` TEXT DEFAULT NULL COMMENT '输入格式',
    `output_format` TEXT DEFAULT NULL COMMENT '输出格式',
    `note` TEXT DEFAULT NULL COMMENT '注意事项/提示',
    `sort_order` INT DEFAULT 0 COMMENT '排序顺序',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_set_id` (`set_id`),
    FOREIGN KEY (`set_id`) REFERENCES `xcpc_problem_set`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目集中的题目';

-- 样例表
CREATE TABLE IF NOT EXISTS `xcpc_problem_example` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `problem_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
    `example_no` INT NOT NULL COMMENT '样例编号',
    `description` VARCHAR(500) DEFAULT NULL COMMENT '样例描述（可选）',
    `input` TEXT NOT NULL COMMENT '样例输入',
    `output` TEXT NOT NULL COMMENT '样例输出',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_problem_id` (`problem_id`),
    UNIQUE KEY `uk_problem_example` (`problem_id`, `example_no`),
    FOREIGN KEY (`problem_id`) REFERENCES `xcpc_problem_set_problem`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目样例';
