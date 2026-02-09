-- ========================================
-- Rating Skip功能 - 数据库迁移脚本
-- ========================================

-- 1. 修改 rating_history 表，添加skip相关字段
ALTER TABLE `rating_history`
ADD COLUMN `is_skip` TINYINT(1) DEFAULT 0 COMMENT '是否被skip',
ADD COLUMN `skip_reason` VARCHAR(500) DEFAULT '' COMMENT 'skip原因';

-- 2. 修改 contest_rating_status 表，添加重算相关字段
ALTER TABLE `contest_rating_status`
ADD COLUMN `skip_count` INT DEFAULT 0 COMMENT 'Skip用户数量',
ADD COLUMN `has_pending_skip` TINYINT(1) DEFAULT 0 COMMENT '是否有待处理的skip',
ADD COLUMN `recalculate_status` VARCHAR(20) DEFAULT 'none' COMMENT '重算状态: none/pending/calculating/completed/failed',
ADD COLUMN `last_recalculate_at` DATETIME COMMENT '最后重算时间',
ADD COLUMN `recalculate_lock` TINYINT(1) DEFAULT 0 COMMENT '重算锁（防止并发）';

-- 3. 创建 contest_skip_users 表（比赛Skip用户记录）
CREATE TABLE IF NOT EXISTS `contest_skip_users` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `contest_id` BIGINT UNSIGNED NOT NULL COMMENT '比赛ID',
    `uid` VARCHAR(32) NOT NULL COMMENT '被skip的用户UID',
    `username` VARCHAR(100) NOT NULL COMMENT '用户名（冗余）',
    `reason` VARCHAR(500) NOT NULL COMMENT 'skip原因（如：代码抄袭）',
    `operator_uid` VARCHAR(32) NOT NULL COMMENT '操作人UID',
    `operator_username` VARCHAR(100) COMMENT '操作人用户名',
    `is_applied` TINYINT(1) DEFAULT 0 COMMENT '是否已应用到rating计算',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    UNIQUE KEY `uk_contest_uid` (`contest_id`, `uid`) COMMENT '比赛+用户唯一索引',
    INDEX `idx_contest_id` (`contest_id`) COMMENT '比赛ID索引',
    INDEX `idx_uid` (`uid`) COMMENT '用户UID索引',
    INDEX `idx_contest_applied` (`contest_id`, `is_applied`) COMMENT '比赛+应用状态复合索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='比赛Rating Skip用户记录';

-- 4. 创建 rating_recalculate_queue 表（Rating重算任务队列）
CREATE TABLE IF NOT EXISTS `rating_recalculate_queue` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `contest_id` BIGINT UNSIGNED NOT NULL COMMENT '起始比赛ID',
    `status` VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/running/completed/failed',
    `total_contests` INT DEFAULT 0 COMMENT '需要重算的比赛总数',
    `processed_contests` INT DEFAULT 0 COMMENT '已处理的比赛数',
    `error_message` TEXT COMMENT '错误信息',
    `created_by` VARCHAR(32) COMMENT '创建人UID',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `started_at` DATETIME COMMENT '开始时间',
    `completed_at` DATETIME COMMENT '完成时间',
    INDEX `idx_status` (`status`) COMMENT '状态索引',
    INDEX `idx_contest` (`contest_id`) COMMENT '比赛ID索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Rating重算任务队列';

-- 5. 创建 rating_operation_logs 表（Rating操作日志）
CREATE TABLE IF NOT EXISTS `rating_operation_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `operator_uid` VARCHAR(32) NOT NULL COMMENT '操作人UID',
    `operator_username` VARCHAR(100) COMMENT '操作人用户名',
    `operation_type` VARCHAR(50) NOT NULL COMMENT '操作类型: personal_adjust/skip_user/cancel_skip/recalculate',
    `target_type` VARCHAR(50) COMMENT '目标类型: user/contest',
    `target_id` VARCHAR(100) COMMENT '目标ID: 用户名或比赛ID',
    `operation_detail` JSON COMMENT '操作详情（JSON格式）',
    `ip` VARCHAR(50) COMMENT '操作IP',
    `created_at` DATETIME NOT NULL COMMENT '操作时间',
    INDEX `idx_operator` (`operator_uid`) COMMENT '操作人索引',
    INDEX `idx_type` (`operation_type`) COMMENT '操作类型索引',
    INDEX `idx_created` (`created_at`) COMMENT '创建时间索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Rating操作日志表';
