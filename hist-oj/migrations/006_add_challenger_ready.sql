-- 添加挑战者准备状态字段
-- Date: 2025-01-10
-- Description: 为 battle_room 表添加 challenger_ready 字段,用于实现挑战者准备机制

-- 添加 challenger_ready 字段
ALTER TABLE `battle_room`
ADD COLUMN `challenger_ready` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '挑战者是否已准备 (0-未准备, 1-已准备)' AFTER `challenger_username`;

-- 添加索引以提高查询性能
ALTER TABLE `battle_room`
ADD INDEX `idx_challenger_ready` (`challenger_ready`);
