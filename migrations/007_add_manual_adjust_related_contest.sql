-- 为 rating_history 添加手动调整关联比赛字段
-- 用于在重算时按“比赛 -> 手动调整”的顺序回放

ALTER TABLE `rating_history`
ADD COLUMN `related_contest_id` BIGINT UNSIGNED NULL COMMENT '手动调整关联的比赛ID' AFTER `contest_id`;

ALTER TABLE `rating_history`
ADD INDEX `idx_related_contest_id` (`related_contest_id`);
