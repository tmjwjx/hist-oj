-- 为 rating_history 表添加手动调整相关字段
-- 执行时间: 2024-01-05

-- 1. 添加 reason 字段（操作原因）
ALTER TABLE `rating_history`
ADD COLUMN `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作原因（手动调整时的备注）'
AFTER `participants`;

-- 2. 添加 is_manual 字段（是否为手动调整）
ALTER TABLE `rating_history`
ADD COLUMN `is_manual` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为手动调整（0=比赛计算，1=手动调整）'
AFTER `reason`;

-- 3. 添加 operator_uid 字段（操作人UID）
ALTER TABLE `rating_history`
ADD COLUMN `operator_uid` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '操作人UID（手动调整时记录）'
AFTER `is_manual`;

-- 4. 添加索引以提高查询性能
ALTER TABLE `rating_history`
ADD INDEX `idx_is_manual` (`is_manual`);

-- 5. 修改 contest_id 字段，允许为 NULL（手动调整时为 0 或 NULL）
ALTER TABLE `rating_history`
MODIFY COLUMN `contest_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '比赛ID（0表示手动调整）';

-- 6. 为已有的记录设置默认值（确保历史数据不为 NULL）
UPDATE `rating_history`
SET `contest_id` = `contest_id` WHERE `contest_id` IS NULL;
