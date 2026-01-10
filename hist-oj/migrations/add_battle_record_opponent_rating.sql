-- 为 battle_record 表添加 opponent_rating 字段
-- 用于存储对手的 rating 值，便于前端显示不同颜色的对手名称

ALTER TABLE `battle_record`
ADD COLUMN `opponent_rating` INT NULL DEFAULT NULL COMMENT '对手的rating值' AFTER `opponent_username`;

-- 添加索引以优化查询性能
ALTER TABLE `battle_record`
ADD INDEX `idx_opponent_rating` (`opponent_rating`);
