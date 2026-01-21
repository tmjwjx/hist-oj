-- 为 training_participant 表添加进度缓存字段
-- 用于存储用户的训练进度，避免每次都查询 judge 表

ALTER TABLE training_participant
ADD COLUMN solved_count BIGINT DEFAULT 0 COMMENT '已解决题目数（缓存）' AFTER gmt_modified,
ADD COLUMN total_count BIGINT DEFAULT 0 COMMENT '训练总题目数（缓存）' AFTER solved_count;

-- 添加索引以加速查询
ALTER TABLE training_participant ADD INDEX idx_training_uid (training_id, uid);
