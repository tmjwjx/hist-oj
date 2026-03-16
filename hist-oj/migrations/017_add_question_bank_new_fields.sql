-- 为 question_bank 表添加新字段：题目解析、题目标签、题目所属课程
-- Date: 2025-03-06

-- 添加题目解析字段
ALTER TABLE `question_bank`
ADD COLUMN `analysis` TEXT NULL COMMENT '题目解析' AFTER `answer`;

-- 添加题目标签字段（JSON格式，存储标签数组）
ALTER TABLE `question_bank`
ADD COLUMN `tags` JSON NULL COMMENT '题目标签（JSON数组格式）' AFTER `analysis`;

-- 添加题目所属课程字段
ALTER TABLE `question_bank`
ADD COLUMN `course` VARCHAR(100) NULL COMMENT '题目所属课程' AFTER `tags`;

-- 为课程字段添加索引，方便按课程筛选
ALTER TABLE `question_bank`
ADD INDEX `idx_course` (`course`);

-- 为标签字段添加虚拟列索引（可选，用于JSON查询优化）
-- ALTER TABLE `question_bank`
-- ADD COLUMN `tags_virtual` VARCHAR(255) AS (JSON_UNQUOTE(JSON_EXTRACT(tags, '$'))) VIRTUAL,
-- ADD INDEX `idx_tags_virtual` (`tags_virtual`);
